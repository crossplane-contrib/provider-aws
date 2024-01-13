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

package connection

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
	errBuildARN       = "cannot built the ARN for Connection" // to be able to update Tags, add the correct ARN to the annotation
	errGetARN         = "cannot get a correct ARN for Connection"
	errMissingARNAnno = "cannot find the annotation for the Connection ARN"
	annotationARN     = "crossplane.io/external-aws-glue-connection-arn"
)

// SetupConnection adds a controller that reconciles Connection.
func SetupConnection(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.ConnectionGroupKind)
	opts := []option{
		func(e *external) {
			h := &hooks{kube: e.kube, client: e.client}
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preDelete = preDelete
			e.postCreate = h.postCreate
			e.preCreate = preCreate
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
		resource.ManagedKind(svcapitypes.ConnectionGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.Connection{}).
		Complete(r)
}

type hooks struct {
	kube   client.Client
	client glueiface.GlueAPI
}

func preDelete(_ context.Context, cr *svcapitypes.Connection, obj *svcsdk.DeleteConnectionInput) (bool, error) {
	obj.ConnectionName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return false, nil
}

func preObserve(_ context.Context, cr *svcapitypes.Connection, obj *svcsdk.GetConnectionInput) error {
	obj.Name = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

func postObserve(ctx context.Context, cr *svcapitypes.Connection, obj *svcsdk.GetConnectionOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	// not needed if we get the fields properly set in GenerateConnection() (metav1 import issue)
	cr.Status.AtProvider.CreationTime = fromTimePtr(obj.Connection.CreationTime)
	cr.Status.AtProvider.LastUpdatedTime = fromTimePtr(obj.Connection.LastUpdatedTime)

	cr.SetConditions(xpv1.Available())
	return obs, nil
}

func (h *hooks) isUpToDate(_ context.Context, cr *svcapitypes.Connection, resp *svcsdk.GetConnectionOutput) (bool, string, error) {
	currentParams := customGenerateConnection(resp).Spec.ForProvider

	if diff := cmp.Diff(cr.Spec.ForProvider, currentParams, cmpopts.EquateEmpty(),
		cmpopts.IgnoreTypes(&xpv1.Reference{}, &xpv1.Selector{}, []xpv1.Reference{}),
		cmpopts.IgnoreFields(svcapitypes.ConnectionParameters{}, "Region", "Tags", "CatalogID")); diff != "" {

		return false, diff, nil
	}
	// CatalogID is updatable (and is given to UpdateConnectionInput),
	// however the field seems not to be readable through the API for an isUpToDate-check

	// retrieve ARN and check if Tags need update
	arn, err := h.getARN(cr)
	if err != nil {
		return true, "", err
	}
	areTagsUpToDate, err := svcutils.AreTagsUpToDate(h.client, cr.Spec.ForProvider.Tags, arn)
	return areTagsUpToDate, "", err
}

func preUpdate(_ context.Context, cr *svcapitypes.Connection, obj *svcsdk.UpdateConnectionInput) error {
	obj.Name = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))

	if cr.Spec.ForProvider.CustomConnectionInput != nil {
		obj.ConnectionInput = &svcsdk.ConnectionInput{
			Name:                 pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
			ConnectionProperties: cr.Spec.ForProvider.CustomConnectionInput.ConnectionProperties,
			ConnectionType:       pointer.ToOrNilIfZeroValue(cr.Spec.ForProvider.CustomConnectionInput.ConnectionType),
			Description:          cr.Spec.ForProvider.CustomConnectionInput.Description,
			MatchCriteria:        cr.Spec.ForProvider.CustomConnectionInput.MatchCriteria,
		}

		if cr.Spec.ForProvider.CustomConnectionInput.CustomPhysicalConnectionRequirements != nil {
			obj.ConnectionInput.PhysicalConnectionRequirements = &svcsdk.PhysicalConnectionRequirements{
				AvailabilityZone: cr.Spec.ForProvider.CustomConnectionInput.CustomPhysicalConnectionRequirements.AvailabilityZone,
				SubnetId:         cr.Spec.ForProvider.CustomConnectionInput.CustomPhysicalConnectionRequirements.SubnetID,
			}
			for i := range cr.Spec.ForProvider.CustomConnectionInput.CustomPhysicalConnectionRequirements.SecurityGroupIDList {
				obj.ConnectionInput.PhysicalConnectionRequirements.SecurityGroupIdList = append(obj.ConnectionInput.PhysicalConnectionRequirements.SecurityGroupIdList, &cr.Spec.ForProvider.CustomConnectionInput.CustomPhysicalConnectionRequirements.SecurityGroupIDList[i])
			}
		}
	}

	return nil
}

func (h *hooks) postUpdate(ctx context.Context, cr *svcapitypes.Connection, obj *svcsdk.UpdateConnectionOutput, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	// update Tags if ARN available
	connectionARN, err := h.getARN(cr)
	if err != nil {
		return upd, err
	}

	return upd, svcutils.UpdateTagsForResource(ctx, h.client, cr.Spec.ForProvider.Tags, connectionARN)
}

func preCreate(_ context.Context, cr *svcapitypes.Connection, obj *svcsdk.CreateConnectionInput) error {

	if cr.Spec.ForProvider.CustomConnectionInput != nil {
		obj.ConnectionInput = &svcsdk.ConnectionInput{
			Name:                 pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
			ConnectionProperties: cr.Spec.ForProvider.CustomConnectionInput.ConnectionProperties,
			ConnectionType:       pointer.ToOrNilIfZeroValue(cr.Spec.ForProvider.CustomConnectionInput.ConnectionType),
			Description:          cr.Spec.ForProvider.CustomConnectionInput.Description,
			MatchCriteria:        cr.Spec.ForProvider.CustomConnectionInput.MatchCriteria,
		}

		if cr.Spec.ForProvider.CustomConnectionInput.CustomPhysicalConnectionRequirements != nil {
			obj.ConnectionInput.PhysicalConnectionRequirements = &svcsdk.PhysicalConnectionRequirements{
				AvailabilityZone: cr.Spec.ForProvider.CustomConnectionInput.CustomPhysicalConnectionRequirements.AvailabilityZone,
				SubnetId:         cr.Spec.ForProvider.CustomConnectionInput.CustomPhysicalConnectionRequirements.SubnetID,
			}
			for i := range cr.Spec.ForProvider.CustomConnectionInput.CustomPhysicalConnectionRequirements.SecurityGroupIDList {
				obj.ConnectionInput.PhysicalConnectionRequirements.SecurityGroupIdList = append(obj.ConnectionInput.PhysicalConnectionRequirements.SecurityGroupIdList, &cr.Spec.ForProvider.CustomConnectionInput.CustomPhysicalConnectionRequirements.SecurityGroupIDList[i])
			}
		}
	}

	return nil
}

func (h *hooks) postCreate(ctx context.Context, cr *svcapitypes.Connection, obj *svcsdk.CreateConnectionOutput, _ managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
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
	connectionARN, err := h.buildARN(ctx, cr)
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	annotation := map[string]string{
		annotationARN: pointer.StringValue(connectionARN),
	}
	meta.AddAnnotations(cr, annotation)

	return managed.ExternalCreation{}, nil
}

// Custom GenerateConnection for isuptodate (the generated one in zz_conversion.go is missing too much)
func customGenerateConnection(resp *svcsdk.GetConnectionOutput) *svcapitypes.Connection {

	cr := &svcapitypes.Connection{}
	cr.Spec.ForProvider.CustomConnectionInput = &svcapitypes.CustomConnectionInput{}

	if resp.Connection.ConnectionProperties != nil {
		cr.Spec.ForProvider.CustomConnectionInput.ConnectionProperties = resp.Connection.ConnectionProperties
	} else {
		cr.Spec.ForProvider.CustomConnectionInput.ConnectionProperties = nil
	}

	if resp.Connection.ConnectionType != nil {
		cr.Spec.ForProvider.CustomConnectionInput.ConnectionType = pointer.StringValue(resp.Connection.ConnectionType)
	}

	if resp.Connection.Description != nil {
		cr.Spec.ForProvider.CustomConnectionInput.Description = resp.Connection.Description
	} else {
		cr.Spec.ForProvider.CustomConnectionInput.Description = nil
	}

	if resp.Connection.MatchCriteria != nil {
		cr.Spec.ForProvider.CustomConnectionInput.MatchCriteria = resp.Connection.MatchCriteria
	} else {
		cr.Spec.ForProvider.CustomConnectionInput.MatchCriteria = nil
	}

	if resp.Connection.PhysicalConnectionRequirements != nil {
		cr.Spec.ForProvider.CustomConnectionInput.CustomPhysicalConnectionRequirements = &svcapitypes.CustomPhysicalConnectionRequirements{
			AvailabilityZone: resp.Connection.PhysicalConnectionRequirements.AvailabilityZone,
			SubnetID:         resp.Connection.PhysicalConnectionRequirements.SubnetId,
		}
		for i := range resp.Connection.PhysicalConnectionRequirements.SecurityGroupIdList {
			cr.Spec.ForProvider.CustomConnectionInput.CustomPhysicalConnectionRequirements.SecurityGroupIDList = append(cr.Spec.ForProvider.CustomConnectionInput.CustomPhysicalConnectionRequirements.SecurityGroupIDList, *resp.Connection.PhysicalConnectionRequirements.SecurityGroupIdList[i])
		}
	} else {
		cr.Spec.ForProvider.CustomConnectionInput.CustomPhysicalConnectionRequirements = nil
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
func (h *hooks) getARN(cr *svcapitypes.Connection) (*string, error) {

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
func (h *hooks) buildARN(ctx context.Context, cr *svcapitypes.Connection) (*string, error) {

	var accountID string
	// when CatalogID is provided, fetching the CallerID is unneeded
	if cr.Spec.ForProvider.CatalogID != nil {
		accountID = pointer.StringValue(cr.Spec.ForProvider.CatalogID)
	} else {
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
	}
	connectionARN := ("arn:aws:glue:" +
		cr.Spec.ForProvider.Region + ":" +
		accountID + ":connection/" +
		meta.GetExternalName(cr))

	if !awsarn.IsARN(connectionARN) {

		return nil, errors.New(errBuildARN)
	}
	return &connectionARN, nil
}
