/*
Copyright 2021 The Crossplane Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package crawler

import (
	"context"
	"errors"
	"strings"
	"time"

	awsarn "github.com/aws/aws-sdk-go/aws/arn"
	svcsdk "github.com/aws/aws-sdk-go/service/glue"
	"github.com/aws/aws-sdk-go/service/glue/glueiface"
	svcsdksts "github.com/aws/aws-sdk-go/service/sts"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/glue/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	svcutils "github.com/crossplane-contrib/provider-aws/pkg/controller/glue/utils"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errBuildARN       = "cannot built the ARN for Crawler" // to be able to update Tags, add the correct ARN to the annotation
	errGetARN         = "cannot get a correct ARN for Crawler"
	errMissingARNAnno = "cannot find the annotation for the Crawler ARN"
	annotationARN     = "crossplane.io/external-aws-glue-crawler-arn"
)

// SetupCrawler adds a controller that reconciles Crawler.
func SetupCrawler(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.CrawlerGroupKind)
	opts := []option{
		func(e *external) {
			h := &hooks{kube: e.kube, client: e.client}
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preDelete = h.preDelete
			e.preCreate = preCreate
			e.postCreate = h.postCreate
			e.lateInitialize = lateInitialize
			e.isUpToDate = h.isUpToDate
			e.preUpdate = preUpdate
			e.postUpdate = h.postUpdate
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(svcapitypes.CrawlerGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.Crawler{}).
		Complete(r)
}

type hooks struct {
	kube   client.Client
	client glueiface.GlueAPI
}

func (h *hooks) preDelete(ctx context.Context, cr *svcapitypes.Crawler, obj *svcsdk.DeleteCrawlerInput) (bool, error) {
	obj.Name = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))

	// delete-requests to AWS will throw an error while the crawler is still working

	// a currently running Crawler cannot be deleted. Need to stop first
	if pointer.StringValue(cr.Status.AtProvider.State) == svcsdk.CrawlerStateRunning {
		_, err := h.client.StopCrawlerWithContext(ctx, &svcsdk.StopCrawlerInput{Name: obj.Name})

		return true, errorutils.Wrap(err, errDelete)
	}
	// wait with delete-request while Crawler is stopping
	if pointer.StringValue(cr.Status.AtProvider.State) == svcsdk.CrawlerStateStopping {

		return true, nil
	}
	return false, nil
}

func preObserve(_ context.Context, cr *svcapitypes.Crawler, obj *svcsdk.GetCrawlerInput) error {
	obj.Name = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.Crawler, obj *svcsdk.GetCrawlerOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	if obj.Crawler.Schedule != nil {
		cr.Status.AtProvider.ScheduleState = obj.Crawler.Schedule.State
	}

	// not needed if we get the fields properly set in GenerateCrawler() (metav1 import issue)
	cr.Status.AtProvider.CreationTime = fromTimePtr(obj.Crawler.CreationTime)
	cr.Status.AtProvider.LastUpdated = fromTimePtr(obj.Crawler.LastUpdated)

	switch pointer.StringValue(obj.Crawler.State) {
	case svcsdk.CrawlerStateRunning,
		svcsdk.CrawlerStateStopping:
		cr.SetConditions(xpv1.Unavailable().WithMessage(pointer.StringValue(obj.Crawler.State)))
		// Prevent Update() call during Running/Stopping state - which will fail.
		obs.ResourceUpToDate = true
	default:
		cr.SetConditions(xpv1.Available())
	}

	return obs, nil
}

//nolint:gocyclo
func lateInitialize(spec *svcapitypes.CrawlerParameters, resp *svcsdk.GetCrawlerOutput) error {

	spec.Configuration = pointer.LateInitialize(spec.Configuration, resp.Crawler.Configuration)

	if spec.LakeFormationConfiguration == nil {
		spec.LakeFormationConfiguration = &svcapitypes.LakeFormationConfiguration{}
	}
	spec.LakeFormationConfiguration.AccountID = pointer.LateInitialize(spec.LakeFormationConfiguration.AccountID, resp.Crawler.LakeFormationConfiguration.AccountId)
	spec.LakeFormationConfiguration.UseLakeFormationCredentials = pointer.LateInitialize(spec.LakeFormationConfiguration.UseLakeFormationCredentials, resp.Crawler.LakeFormationConfiguration.UseLakeFormationCredentials)

	if spec.LineageConfiguration == nil {
		spec.LineageConfiguration = &svcapitypes.LineageConfiguration{}
	}
	spec.LineageConfiguration.CrawlerLineageSettings = pointer.LateInitialize(spec.LineageConfiguration.CrawlerLineageSettings, resp.Crawler.LineageConfiguration.CrawlerLineageSettings)

	if spec.RecrawlPolicy == nil {
		spec.RecrawlPolicy = &svcapitypes.RecrawlPolicy{}
	}
	spec.RecrawlPolicy.RecrawlBehavior = pointer.LateInitialize(spec.RecrawlPolicy.RecrawlBehavior, resp.Crawler.RecrawlPolicy.RecrawlBehavior)

	if spec.SchemaChangePolicy == nil {
		spec.SchemaChangePolicy = &svcapitypes.SchemaChangePolicy{}
	}
	spec.SchemaChangePolicy.DeleteBehavior = pointer.LateInitialize(spec.SchemaChangePolicy.DeleteBehavior, resp.Crawler.SchemaChangePolicy.DeleteBehavior)
	spec.SchemaChangePolicy.UpdateBehavior = pointer.LateInitialize(spec.SchemaChangePolicy.UpdateBehavior, resp.Crawler.SchemaChangePolicy.UpdateBehavior)

	if resp.Crawler.Targets.JdbcTargets != nil && spec.Targets.JDBCTargets != nil {

		for i, jdbcTarsIter := range resp.Crawler.Targets.JdbcTargets {

			if spec.Targets.JDBCTargets[i] != nil {
				spec.Targets.JDBCTargets[i].Path = pointer.LateInitialize(spec.Targets.JDBCTargets[i].Path, jdbcTarsIter.Path)
			}
		}
	}
	if resp.Crawler.Targets.MongoDBTargets != nil && spec.Targets.MongoDBTargets != nil {

		for i, monTarsIter := range resp.Crawler.Targets.MongoDBTargets {
			if spec.Targets.MongoDBTargets[i] != nil {
				spec.Targets.MongoDBTargets[i].ScanAll = pointer.LateInitialize(spec.Targets.MongoDBTargets[i].ScanAll, monTarsIter.ScanAll)
			}
		}
	}

	return nil
}

func (h *hooks) isUpToDate(_ context.Context, cr *svcapitypes.Crawler, resp *svcsdk.GetCrawlerOutput) (bool, string, error) {
	currentParams := customGenerateCrawler(resp).Spec.ForProvider

	// separate check bc: 1.lowercase handling 2.field Schedule has different input/output shapes (see generator-config.yaml)
	if !strings.EqualFold(pointer.StringValue(cr.Spec.ForProvider.Schedule), pointer.StringValue(currentParams.Schedule)) {

		return false, "", nil
	}

	// user can provide either ARN or name for role; AWS API gives role name back
	if awsarn.IsARN(cr.Spec.ForProvider.Role) {

		roleARN, _ := awsarn.Parse(cr.Spec.ForProvider.Role)
		roleName := strings.TrimPrefix(roleARN.Resource, "role/")
		if !strings.EqualFold(roleName, currentParams.Role) {

			return false, "", nil
		}
	} else if diff := cmp.Diff(cr.Spec.ForProvider.Role, currentParams.Role); diff != "" {

		return false, diff, nil
	}

	if diff := cmp.Diff(cr.Spec.ForProvider, currentParams, cmpopts.EquateEmpty(),
		cmpopts.IgnoreTypes(&xpv1.Reference{}, &xpv1.Selector{}, []xpv1.Reference{}),
		cmpopts.IgnoreFields(svcapitypes.CrawlerParameters{}, "Region", "Schedule", "Role", "Tags")); diff != "" {

		return false, diff, nil
	}

	// retrieve ARN and check if Tags need update
	arn, err := h.getARN(cr)
	if err != nil {
		return true, "", err
	}
	areTagsUpToDate, err := svcutils.AreTagsUpToDate(h.client, cr.Spec.ForProvider.Tags, arn)
	return areTagsUpToDate, "", err
}

//nolint:gocyclo
func preUpdate(_ context.Context, cr *svcapitypes.Crawler, obj *svcsdk.UpdateCrawlerInput) error {
	obj.Name = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))

	obj.SetClassifiers(cr.Spec.ForProvider.Classifiers)
	obj.CrawlerSecurityConfiguration = cr.Spec.ForProvider.CrawlerSecurityConfiguration
	obj.DatabaseName = cr.Spec.ForProvider.DatabaseName
	obj.Role = pointer.ToOrNilIfZeroValue(cr.Spec.ForProvider.Role)

	obj.Targets = &svcsdk.CrawlerTargets{}

	if cr.Spec.ForProvider.Targets.CatalogTargets != nil {
		catTars := []*svcsdk.CatalogTarget{}
		for _, catTarsIter := range cr.Spec.ForProvider.Targets.CatalogTargets {
			catTarsElem := &svcsdk.CatalogTarget{
				DatabaseName: ptr.To(catTarsIter.DatabaseName),
				Tables:       pointer.SliceValueToPtr(catTarsIter.Tables),
			}
			catTars = append(catTars, catTarsElem)
		}
		obj.Targets.CatalogTargets = catTars
	}
	if cr.Spec.ForProvider.Targets.DynamoDBTargets != nil {
		dynTars := []*svcsdk.DynamoDBTarget{}
		for _, dynTarsIter := range cr.Spec.ForProvider.Targets.DynamoDBTargets {
			dynTarsElem := &svcsdk.DynamoDBTarget{
				Path:     dynTarsIter.Path,
				ScanAll:  dynTarsIter.ScanAll,
				ScanRate: dynTarsIter.ScanRate,
			}
			dynTars = append(dynTars, dynTarsElem)
		}
		obj.Targets.DynamoDBTargets = dynTars
	}
	if cr.Spec.ForProvider.Targets.JDBCTargets != nil {
		jdbcTars := []*svcsdk.JdbcTarget{}
		for _, jdbcTarsIter := range cr.Spec.ForProvider.Targets.JDBCTargets {
			jdbcTarsElem := &svcsdk.JdbcTarget{
				ConnectionName: jdbcTarsIter.ConnectionName,
				Exclusions:     jdbcTarsIter.Exclusions,
				Path:           jdbcTarsIter.Path,
			}
			jdbcTars = append(jdbcTars, jdbcTarsElem)
		}
		obj.Targets.JdbcTargets = jdbcTars
	}
	if cr.Spec.ForProvider.Targets.MongoDBTargets != nil {
		monTars := []*svcsdk.MongoDBTarget{}
		for _, monTarsIter := range cr.Spec.ForProvider.Targets.MongoDBTargets {
			monTarsElem := &svcsdk.MongoDBTarget{
				ConnectionName: monTarsIter.ConnectionName,
				Path:           monTarsIter.Path,
				ScanAll:        monTarsIter.ScanAll,
			}
			monTars = append(monTars, monTarsElem)
		}
		obj.Targets.MongoDBTargets = monTars
	}
	if cr.Spec.ForProvider.Targets.S3Targets != nil {
		s3Tars := []*svcsdk.S3Target{}
		for _, s3TarsIter := range cr.Spec.ForProvider.Targets.S3Targets {
			s3TarsElem := &svcsdk.S3Target{
				ConnectionName:   s3TarsIter.ConnectionName,
				DlqEventQueueArn: s3TarsIter.DlqEventQueueARN,
				EventQueueArn:    s3TarsIter.EventQueueARN,
				Exclusions:       s3TarsIter.Exclusions,
				Path:             s3TarsIter.Path,
				SampleSize:       s3TarsIter.SampleSize,
			}
			s3Tars = append(s3Tars, s3TarsElem)
		}
		obj.Targets.S3Targets = s3Tars
	}

	return nil
}

func (h *hooks) postUpdate(ctx context.Context, cr *svcapitypes.Crawler, obj *svcsdk.UpdateCrawlerOutput, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	if err != nil {
		return managed.ExternalUpdate{}, err
	}
	// update Tags if ARN available
	crawlerARN, err := h.getARN(cr)
	if err != nil {
		return upd, err
	}

	return upd, svcutils.UpdateTagsForResource(ctx, h.client, cr.Spec.ForProvider.Tags, crawlerARN)
}

//nolint:gocyclo
func preCreate(_ context.Context, cr *svcapitypes.Crawler, obj *svcsdk.CreateCrawlerInput) error {
	obj.Name = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))

	obj.Role = pointer.ToOrNilIfZeroValue(cr.Spec.ForProvider.Role)
	obj.SetClassifiers(cr.Spec.ForProvider.Classifiers)
	obj.CrawlerSecurityConfiguration = cr.Spec.ForProvider.CrawlerSecurityConfiguration
	obj.DatabaseName = cr.Spec.ForProvider.DatabaseName

	obj.Targets = &svcsdk.CrawlerTargets{}

	if cr.Spec.ForProvider.Targets.CatalogTargets != nil {
		catTars := []*svcsdk.CatalogTarget{}
		for _, catTarsIter := range cr.Spec.ForProvider.Targets.CatalogTargets {
			catTarsElem := &svcsdk.CatalogTarget{
				DatabaseName: ptr.To(catTarsIter.DatabaseName),
				Tables:       pointer.SliceValueToPtr(catTarsIter.Tables),
			}
			catTars = append(catTars, catTarsElem)
		}
		obj.Targets.CatalogTargets = catTars
	}
	if cr.Spec.ForProvider.Targets.DynamoDBTargets != nil {
		dynTars := []*svcsdk.DynamoDBTarget{}
		for _, dynTarsIter := range cr.Spec.ForProvider.Targets.DynamoDBTargets {
			dynTarsElem := &svcsdk.DynamoDBTarget{
				Path:     dynTarsIter.Path,
				ScanAll:  dynTarsIter.ScanAll,
				ScanRate: dynTarsIter.ScanRate,
			}
			dynTars = append(dynTars, dynTarsElem)
		}
		obj.Targets.DynamoDBTargets = dynTars
	}
	if cr.Spec.ForProvider.Targets.JDBCTargets != nil {
		jdbcTars := []*svcsdk.JdbcTarget{}
		for _, jdbcTarsIter := range cr.Spec.ForProvider.Targets.JDBCTargets {
			jdbcTarsElem := &svcsdk.JdbcTarget{
				ConnectionName: jdbcTarsIter.ConnectionName,
				Exclusions:     jdbcTarsIter.Exclusions,
				Path:           jdbcTarsIter.Path,
			}
			jdbcTars = append(jdbcTars, jdbcTarsElem)
		}
		obj.Targets.JdbcTargets = jdbcTars
	}
	if cr.Spec.ForProvider.Targets.MongoDBTargets != nil {
		monTars := []*svcsdk.MongoDBTarget{}
		for _, monTarsIter := range cr.Spec.ForProvider.Targets.MongoDBTargets {
			monTarsElem := &svcsdk.MongoDBTarget{
				ConnectionName: monTarsIter.ConnectionName,
				Path:           monTarsIter.Path,
				ScanAll:        monTarsIter.ScanAll,
			}
			monTars = append(monTars, monTarsElem)
		}
		obj.Targets.MongoDBTargets = monTars
	}
	if cr.Spec.ForProvider.Targets.S3Targets != nil {
		s3Tars := []*svcsdk.S3Target{}
		for _, s3TarsIter := range cr.Spec.ForProvider.Targets.S3Targets {
			s3TarsElem := &svcsdk.S3Target{
				ConnectionName:   s3TarsIter.ConnectionName,
				DlqEventQueueArn: s3TarsIter.DlqEventQueueARN,
				EventQueueArn:    s3TarsIter.EventQueueARN,
				Exclusions:       s3TarsIter.Exclusions,
				Path:             s3TarsIter.Path,
				SampleSize:       s3TarsIter.SampleSize,
			}
			s3Tars = append(s3Tars, s3TarsElem)
		}
		obj.Targets.S3Targets = s3Tars
	}

	return nil
}

func (h *hooks) postCreate(ctx context.Context, cr *svcapitypes.Crawler, obj *svcsdk.CreateCrawlerOutput, _ managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	// the AWS API does not provide ARNs for Glue Resources
	// however an ARN is needed to read and update Tags
	// therefore we built the ARN of the Glue Resource from
	// the Caller's AccountID directly after creation and
	// save it as an annotation for later retrieval
	// this also allows the user to manually add or
	// change the ARN if we get it wrong

	// build ARN and save/add it to the annotations
	crawlerARN, err := h.buildARN(ctx, cr)
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	annotation := map[string]string{
		annotationARN: pointer.StringValue(crawlerARN),
	}
	meta.AddAnnotations(cr, annotation)

	return managed.ExternalCreation{}, nil
}

// Custom GenerateCrawler for isuptodate to fill the missing fields not forwarded by GenerateCrawler in zz_conversion.go
//
//nolint:gocyclo
func customGenerateCrawler(resp *svcsdk.GetCrawlerOutput) *svcapitypes.Crawler {

	cr := GenerateCrawler(resp)

	cr.Spec.ForProvider.Classifiers = resp.Crawler.Classifiers
	cr.Spec.ForProvider.CrawlerSecurityConfiguration = resp.Crawler.CrawlerSecurityConfiguration
	cr.Spec.ForProvider.DatabaseName = resp.Crawler.DatabaseName
	cr.Spec.ForProvider.Role = pointer.StringValue(resp.Crawler.Role)
	cr.Spec.ForProvider.Targets = svcapitypes.CustomCrawlerTargets{}

	if resp.Crawler.Targets.CatalogTargets != nil {
		catTars := []*svcapitypes.CustomCatalogTarget{}
		for _, catTarsIter := range resp.Crawler.Targets.CatalogTargets {
			catTarsElem := &svcapitypes.CustomCatalogTarget{
				DatabaseName: pointer.StringValue(catTarsIter.DatabaseName),
				Tables:       pointer.SlicePtrToValue(catTarsIter.Tables),
			}
			catTars = append(catTars, catTarsElem)
		}
		cr.Spec.ForProvider.Targets.CatalogTargets = catTars
	}
	if resp.Crawler.Targets.DynamoDBTargets != nil {
		dynTars := []*svcapitypes.DynamoDBTarget{}
		for _, dynTarsIter := range resp.Crawler.Targets.DynamoDBTargets {
			dynTarsElem := &svcapitypes.DynamoDBTarget{
				Path:     dynTarsIter.Path,
				ScanAll:  dynTarsIter.ScanAll,
				ScanRate: dynTarsIter.ScanRate,
			}
			dynTars = append(dynTars, dynTarsElem)
		}
		cr.Spec.ForProvider.Targets.DynamoDBTargets = dynTars
	}
	if resp.Crawler.Targets.JdbcTargets != nil {
		jdbcTars := []*svcapitypes.CustomJDBCTarget{}
		for _, jdbcTarsIter := range resp.Crawler.Targets.JdbcTargets {
			jdbcTarsElem := &svcapitypes.CustomJDBCTarget{
				ConnectionName: jdbcTarsIter.ConnectionName,
				Exclusions:     jdbcTarsIter.Exclusions,
				Path:           jdbcTarsIter.Path,
			}
			jdbcTars = append(jdbcTars, jdbcTarsElem)
		}
		cr.Spec.ForProvider.Targets.JDBCTargets = jdbcTars
	}
	if resp.Crawler.Targets.MongoDBTargets != nil {
		monTars := []*svcapitypes.CustomMongoDBTarget{}
		for _, monTarsIter := range resp.Crawler.Targets.MongoDBTargets {
			monTarsElem := &svcapitypes.CustomMongoDBTarget{
				ConnectionName: monTarsIter.ConnectionName,
				Path:           monTarsIter.Path,
				ScanAll:        monTarsIter.ScanAll,
			}
			monTars = append(monTars, monTarsElem)
		}
		cr.Spec.ForProvider.Targets.MongoDBTargets = monTars
	}
	if resp.Crawler.Targets.S3Targets != nil {
		s3Tars := []*svcapitypes.CustomS3Target{}
		for _, s3TarsIter := range resp.Crawler.Targets.S3Targets {
			s3TarsElem := &svcapitypes.CustomS3Target{
				ConnectionName:   s3TarsIter.ConnectionName,
				DlqEventQueueARN: s3TarsIter.DlqEventQueueArn,
				EventQueueARN:    s3TarsIter.EventQueueArn,
				Exclusions:       s3TarsIter.Exclusions,
				Path:             s3TarsIter.Path,
				SampleSize:       s3TarsIter.SampleSize,
			}
			s3Tars = append(s3Tars, s3TarsElem)
		}
		cr.Spec.ForProvider.Targets.S3Targets = s3Tars
	}

	return cr
}

// fromTimePtr probably not needed if metav1 import issue in zz_conversions.go is fixed
// see https://github.com/aws-controllers-k8s/community/issues/1372

// fromTimePtr is a helper for converting a *time.Time to a *metav1.Time
func fromTimePtr(t *time.Time) *metav1.Time {
	if t != nil {
		m := metav1.NewTime(*t)
		return &m
	}
	return nil
}

// getARN is a helper to retrieve the saved ARN from the annotation
func (h *hooks) getARN(cr *svcapitypes.Crawler) (*string, error) {

	var arn string
	// retrieve
	for anno, content := range cr.GetAnnotations() {

		if anno == annotationARN {
			arn = content
		}
	}

	if !awsarn.IsARN(arn) {

		if arn == "" {
			return nil, errors.New(errMissingARNAnno)
		}

		return nil, errors.New(errGetARN)
	}

	return &arn, nil
}

// buildARN builds the Resource ARN from the Caller AccountID
func (h *hooks) buildARN(ctx context.Context, cr *svcapitypes.Crawler) (*string, error) {

	var accountID string

	sess, err := connectaws.GetConfigV1(ctx, h.kube, cr, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, errorutils.Wrap(err, errBuildARN)
	}

	stsclient := svcsdksts.New(sess)

	callerID, err := stsclient.GetCallerIdentityWithContext(ctx, &svcsdksts.GetCallerIdentityInput{})
	if err != nil {
		return nil, err
	}
	accountID = pointer.StringValue(callerID.Account)

	crawlerARN := ("arn:aws:glue:" +
		cr.Spec.ForProvider.Region + ":" +
		accountID + ":crawler/" +
		meta.GetExternalName(cr))

	if !awsarn.IsARN(crawlerARN) {

		return nil, errors.New(errBuildARN)
	}
	return &crawlerARN, nil
}
