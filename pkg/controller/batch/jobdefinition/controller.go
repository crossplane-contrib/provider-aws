/*
Copyright 2022 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permission and
limitations under the License.
*/

package jobdefinition

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/batch"
	"github.com/aws/aws-sdk-go/service/batch/batchiface"
	"github.com/aws/smithy-go"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/batch/manualv1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	svcutils "github.com/crossplane-contrib/provider-aws/pkg/controller/batch/utils"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errCreateSession           = "cannot create a new session"
	errSetExternalNameFailed   = "cannot set external name of Batch JobDefinition"
	errNotBatchJobDefinition   = "managed resource is not an Batch JobDefinition custom resource"
	errRegisterJobDefinition   = "cannot register Batch JobDefinition"
	errDeregisterJobDefinition = "cannot deregister Batch JobDefinition"
	errDescribeJobDefinition   = "cannot describe Batch JobDefinition"
)

// SetupJobDefinition adds a controller that reconciles JobDefinitions.
func SetupJobDefinition(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.JobDefinitionKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
		managed.WithInitializers(&externalNameGenerator{kube: mgr.GetClient()}),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(svcapitypes.JobDefinitionGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.JobDefinition{}).
		Complete(r)
}

type connector struct {
	kube client.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*svcapitypes.JobDefinition)
	if !ok {
		return nil, errors.New(errNotBatchJobDefinition)
	}
	sess, err := connectaws.GetConfigV1(ctx, c.kube, mg, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, errors.Wrap(err, errCreateSession)
	}
	return &external{client: svcsdk.New(sess), kube: c.kube}, nil
}

type external struct {
	client batchiface.BatchAPI
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*svcapitypes.JobDefinition)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotBatchJobDefinition)
	}

	resp, err := e.client.DescribeJobDefinitionsWithContext(ctx, &svcsdk.DescribeJobDefinitionsInput{
		JobDefinitionName: &cr.Name,
		Status:            pointer.ToOrNilIfZeroValue("ACTIVE"), // to not get an older, inactive version/revision before we finish create!
	})
	if err != nil {
		return managed.ExternalObservation{}, errorutils.Wrap(resource.Ignore(isErrorNotFound, err), errDescribeJobDefinition)
	}
	if len(resp.JobDefinitions) == 0 {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	currentJobDefinition := generateJobDefinition(resp)
	current := cr.Spec.ForProvider.DeepCopy()

	e.lateInitialize(&cr.Spec.ForProvider, &currentJobDefinition.Spec.ForProvider)

	cr.Status.AtProvider = currentJobDefinition.Status.AtProvider

	switch pointer.StringValue(cr.Status.AtProvider.Status) {
	case "ACTIVE":
		cr.SetConditions(xpv1.Available().WithMessage(pointer.StringValue(resp.JobDefinitions[0].Status)))
	case "INACTIVE":
		cr.SetConditions(xpv1.Unavailable().WithMessage(pointer.StringValue(resp.JobDefinitions[0].Status) + " INACTIVE is considered deleted"))
		return managed.ExternalObservation{ResourceExists: false}, nil
		// INACTIVE JobDefinitions are only permanently deleted by AWS after 180 days.
		// These JDs could only be used to make a revision (copy/clone). Or to edit tags.
		// They cannot be made ACTIVE again. INACTIVE seems to be the closest to what we would consider DELETED.
	}

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        e.isUpToDate(cr, currentJobDefinition),
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

func (e *external) isUpToDate(spec, current *svcapitypes.JobDefinition) bool {

	// only check the tags as that is the only thing that can be updated
	return cmp.Equal(spec.Spec.ForProvider.Tags, current.Spec.ForProvider.Tags, cmpopts.EquateEmpty())
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*svcapitypes.JobDefinition)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotBatchJobDefinition)
	}
	cr.Status.SetConditions(xpv1.Creating())

	_, err := e.client.RegisterJobDefinitionWithContext(ctx, generateRegisterJobDefinitionInput(cr))
	return managed.ExternalCreation{}, errorutils.Wrap(err, errRegisterJobDefinition)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*svcapitypes.JobDefinition)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotBatchJobDefinition)
	}

	// for JobDefinition only tags are updatable
	// the AWS "revision" concept (number in ARN after the name) - which comes closest to updating entire ressource - is basically just cloning or copying
	// which means deleting the resource and making a new one -> new ARN (if same name, AWS ++ the revision number)
	return managed.ExternalUpdate{}, svcutils.UpdateTagsForResource(ctx, e.client, cr.Spec.ForProvider.Tags, cr.Status.AtProvider.JobDefinitionArn)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*svcapitypes.JobDefinition)
	if !ok {
		return errors.New(errNotBatchJobDefinition)
	}

	if cr.Status.AtProvider.Status == pointer.ToOrNilIfZeroValue("INACTIVE") {
		return nil
	}
	cr.Status.SetConditions(xpv1.Deleting())

	_, err := e.client.DeregisterJobDefinitionWithContext(ctx, generateDeregisterJobDefinitionInput(cr))
	return errorutils.Wrap(resource.Ignore(isErrorNotFound, err), errDeregisterJobDefinition)
}

func (e *external) lateInitialize(spec, current *svcapitypes.JobDefinitionParameters) {
	// spec.PlatformCapabilities = pointer.LateInitializeSlice(spec.PlatformCapabilities, current.PlatformCapabilities)
	// spec.PropagateTags = pointer.LateInitialize(spec.PropagateTags, current.PropagateTags)
	// ^ doc hints default value, however these fields (also in AWS Console) stay empty...

	if current.ContainerProperties != nil {

		if cmp.Equal(spec.PlatformCapabilities, []*string{pointer.ToOrNilIfZeroValue(svcsdk.PlatformCapabilityFargate)}) && spec.ContainerProperties.FargatePlatformConfiguration == nil {
			spec.ContainerProperties.FargatePlatformConfiguration = &svcapitypes.FargatePlatformConfiguration{PlatformVersion: current.ContainerProperties.FargatePlatformConfiguration.PlatformVersion}
		}
	}
}

// IsErrorNotFound helper function to test for ResourceNotFoundException error.
func isErrorNotFound(err error) bool {
	var awsErr smithy.APIError
	return errors.As(err, &awsErr) && awsErr.ErrorCode() == "JobDefinitionNotFoundException"
}

type externalNameGenerator struct {
	kube client.Client
}

// Initialize the given managed resource.
func (e *externalNameGenerator) Initialize(ctx context.Context, mg resource.Managed) error {
	if meta.GetExternalName(mg) != "" {
		return nil
	}
	cr, ok := mg.(*svcapitypes.JobDefinition)
	if !ok {
		return errors.New(errNotBatchJobDefinition)
	}

	meta.SetExternalName(mg, cr.Name)
	return errors.Wrap(e.kube.Update(ctx, mg), errSetExternalNameFailed)
}
