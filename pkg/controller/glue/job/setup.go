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

package job

import (
	"context"
	"errors"
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
	errBuildARN       = "cannot built the ARN for Job" // to be able to update Tags, add the correct ARN to the annotation
	errGetARN         = "cannot get a correct ARN for Job"
	errMissingARNAnno = "cannot find the annotation for the Job ARN"
	annotationARN     = "crossplane.io/external-aws-glue-job-arn"
)

// SetupJob adds a controller that reconciles Job.
func SetupJob(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.JobGroupKind)
	opts := []option{
		func(e *external) {
			h := &hooks{kube: e.kube, client: e.client}
			e.postCreate = h.postCreate
			e.preDelete = preDelete
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preCreate = preCreate
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
		resource.ManagedKind(svcapitypes.JobGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.Job{}).
		Complete(r)
}

type hooks struct {
	kube   client.Client
	client glueiface.GlueAPI
}

func preDelete(_ context.Context, cr *svcapitypes.Job, obj *svcsdk.DeleteJobInput) (bool, error) {
	obj.JobName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return false, nil
}

func preObserve(_ context.Context, cr *svcapitypes.Job, obj *svcsdk.GetJobInput) error {
	obj.JobName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.Job, obj *svcsdk.GetJobOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	// not needed if we get the fields properly set in GenerateCrawler() (metav1 import issue)
	cr.Status.AtProvider.CreatedOn = fromTimePtr(obj.Job.CreatedOn)
	cr.Status.AtProvider.LastModifiedOn = fromTimePtr(obj.Job.LastModifiedOn)

	cr.SetConditions(xpv1.Available())
	return obs, nil
}

func lateInitialize(spec *svcapitypes.JobParameters, resp *svcsdk.GetJobOutput) error {

	// Command is required, so spec should never be nil
	if resp.Job.Command != nil {
		spec.Command.PythonVersion = pointer.LateInitialize(spec.Command.PythonVersion, resp.Job.Command.PythonVersion)
	}

	if spec.ExecutionProperty == nil {
		spec.ExecutionProperty = &svcapitypes.ExecutionProperty{}
	}
	spec.ExecutionProperty.MaxConcurrentRuns = pointer.LateInitialize(spec.ExecutionProperty.MaxConcurrentRuns, resp.Job.ExecutionProperty.MaxConcurrentRuns)

	spec.GlueVersion = pointer.LateInitialize(spec.GlueVersion, resp.Job.GlueVersion)
	spec.MaxRetries = pointer.LateInitialize(spec.MaxRetries, resp.Job.MaxRetries)
	spec.Timeout = pointer.LateInitialize(spec.Timeout, resp.Job.Timeout)

	// if WorkerType & NumberOfWorkers are used, AWS is in charge of setting MaxCapacity (not the user, not we)
	if spec.MaxCapacity == nil || (spec.WorkerType != nil && spec.NumberOfWorkers != nil) {

		spec.MaxCapacity = resp.Job.MaxCapacity
	}

	return nil
}

func (h *hooks) isUpToDate(_ context.Context, cr *svcapitypes.Job, resp *svcsdk.GetJobOutput) (bool, string, error) {
	currentParams := customGenerateJob(resp).Spec.ForProvider

	if diff := cmp.Diff(cr.Spec.ForProvider, currentParams, cmpopts.EquateEmpty(),
		cmpopts.IgnoreTypes(&xpv1.Reference{}, &xpv1.Selector{}, []xpv1.Reference{}),
		cmpopts.IgnoreFields(svcapitypes.JobParameters{}, "Region", "AllocatedCapacity", "Tags")); diff != "" {
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

func preUpdate(_ context.Context, cr *svcapitypes.Job, obj *svcsdk.UpdateJobInput) error {
	obj.JobName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))

	obj.JobUpdate = &svcsdk.JobUpdate{
		DefaultArguments:        cr.Spec.ForProvider.DefaultArguments,
		Description:             cr.Spec.ForProvider.Description,
		GlueVersion:             cr.Spec.ForProvider.GlueVersion,
		LogUri:                  cr.Spec.ForProvider.LogURI,
		MaxRetries:              cr.Spec.ForProvider.MaxRetries,
		NonOverridableArguments: cr.Spec.ForProvider.NonOverridableArguments,
		NumberOfWorkers:         cr.Spec.ForProvider.NumberOfWorkers,
		Role:                    &cr.Spec.ForProvider.Role,
		SecurityConfiguration:   cr.Spec.ForProvider.SecurityConfiguration,
		Timeout:                 cr.Spec.ForProvider.Timeout,
		WorkerType:              cr.Spec.ForProvider.WorkerType,
	}

	if cr.Spec.ForProvider.Command != nil {
		obj.JobUpdate.Command = &svcsdk.JobCommand{
			Name:           cr.Spec.ForProvider.Command.Name,
			PythonVersion:  cr.Spec.ForProvider.Command.PythonVersion,
			ScriptLocation: cr.Spec.ForProvider.Command.ScriptLocation,
		}
	}

	if cr.Spec.ForProvider.Connections != nil {
		obj.JobUpdate.Connections = &svcsdk.ConnectionsList{
			Connections: cr.Spec.ForProvider.Connections}
	}

	if cr.Spec.ForProvider.ExecutionProperty != nil {
		obj.JobUpdate.ExecutionProperty = &svcsdk.ExecutionProperty{
			MaxConcurrentRuns: cr.Spec.ForProvider.ExecutionProperty.MaxConcurrentRuns,
		}
	}

	if cr.Spec.ForProvider.NotificationProperty != nil {
		obj.JobUpdate.NotificationProperty = &svcsdk.NotificationProperty{
			NotifyDelayAfter: cr.Spec.ForProvider.NotificationProperty.NotifyDelayAfter,
		}
	}

	// cannot set MaxCapacity if using WorkerType or NumberOfWorkers
	// but spec MaxCapacity is still set in lateInit bc AWS uses/sets MaxCapacity somehow in relation to NumberOfWorkers
	if obj.JobUpdate.WorkerType != nil || obj.JobUpdate.NumberOfWorkers != nil {
		return nil
	}

	obj.JobUpdate.MaxCapacity = cr.Spec.ForProvider.MaxCapacity
	return nil
}

func (h *hooks) postUpdate(ctx context.Context, cr *svcapitypes.Job, obj *svcsdk.UpdateJobOutput, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
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

func preCreate(_ context.Context, cr *svcapitypes.Job, obj *svcsdk.CreateJobInput) error {
	obj.Name = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))

	obj.Role = pointer.ToOrNilIfZeroValue(cr.Spec.ForProvider.Role)
	obj.SecurityConfiguration = cr.Spec.ForProvider.SecurityConfiguration

	if cr.Spec.ForProvider.Connections != nil {
		obj.Connections = &svcsdk.ConnectionsList{
			Connections: cr.Spec.ForProvider.Connections}
	}

	return nil
}

func (h *hooks) postCreate(ctx context.Context, cr *svcapitypes.Job, obj *svcsdk.CreateJobOutput, _ managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
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
	jobARN, err := h.buildARN(ctx, cr)
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	annotation := map[string]string{
		annotationARN: pointer.StringValue(jobARN),
	}
	meta.AddAnnotations(cr, annotation)

	return managed.ExternalCreation{}, nil
}

// Custom GenerateJob for isuptodate to fill the missing fields not forwarded by GenerateJob in zz_conversion.go
func customGenerateJob(resp *svcsdk.GetJobOutput) *svcapitypes.Job {

	cr := GenerateJob(resp)

	cr.Spec.ForProvider.Role = pointer.StringValue(resp.Job.Role)
	cr.Spec.ForProvider.SecurityConfiguration = resp.Job.SecurityConfiguration

	if resp.Job.Connections != nil {
		cr.Spec.ForProvider.Connections = resp.Job.Connections.Connections
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
func (h *hooks) getARN(cr *svcapitypes.Job) (*string, error) {

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
func (h *hooks) buildARN(ctx context.Context, cr *svcapitypes.Job) (*string, error) {

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

	jobARN := ("arn:aws:glue:" +
		cr.Spec.ForProvider.Region + ":" +
		accountID + ":job/" +
		meta.GetExternalName(cr))

	if !awsarn.IsARN(jobARN) {

		return nil, errors.New(errBuildARN)
	}
	return &jobARN, nil
}
