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

package job

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
	errCreateSession = "cannot create a new session"
	errNotBatchJob   = "managed resource is not an Batch Job custom resource"
	errSubmitJob     = "cannot submit Batch Job"
	errTerminateJob  = "cannot terminate/delete Batch Job"
	errDescribeJob   = "cannot describe Batch Job"
)

// SetupJob adds a controller that reconciles Jobs.
func SetupJob(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.JobKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
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
		resource.ManagedKind(svcapitypes.JobGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.Job{}).
		Complete(r)
}

type connector struct {
	kube client.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*svcapitypes.Job)
	if !ok {
		return nil, errors.New(errNotBatchJob)
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
	cr, ok := mg.(*svcapitypes.Job)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotBatchJob)
	}
	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	resp, err := e.client.DescribeJobsWithContext(ctx, &svcsdk.DescribeJobsInput{
		Jobs: []*string{pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))},
	})
	if err != nil {
		return managed.ExternalObservation{}, errorutils.Wrap(resource.Ignore(isErrorNotFound, err), errDescribeJob)
	}
	if len(resp.Jobs) == 0 {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	currentJob := generateJob(resp)
	current := cr.Spec.ForProvider.DeepCopy()

	e.lateInitialize(&cr.Spec.ForProvider, &currentJob.Spec.ForProvider)

	cr.Status.AtProvider = currentJob.Status.AtProvider

	// Only consider a finished Job as deleted, when the user requested the deletion
	if meta.WasDeleted(cr) {
		// (unfinished Jobs are moved to Failed-status by AWS after deletion/termination-request completed)
		switch pointer.StringValue(cr.Status.AtProvider.Status) {
		case svcsdk.JobStatusFailed,
			svcsdk.JobStatusSucceeded:
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
	}

	cr.SetConditions(xpv1.Available().WithMessage(pointer.StringValue(cr.Status.AtProvider.Status)))

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        e.isUpToDate(cr, currentJob),
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider, cmpopts.EquateEmpty()),
	}, nil
}

func (e *external) isUpToDate(spec, current *svcapitypes.Job) bool {

	// only check the tags as that is the only thing that can be updated
	return cmp.Equal(spec.Spec.ForProvider.Tags, current.Spec.ForProvider.Tags, cmpopts.EquateEmpty())
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*svcapitypes.Job)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotBatchJob)
	}
	cr.Status.SetConditions(xpv1.Creating())

	resp, err := e.client.SubmitJobWithContext(ctx, generateSubmitJobInput(cr))
	if err != nil {
		return managed.ExternalCreation{}, errorutils.Wrap(err, errSubmitJob)
	}
	meta.SetExternalName(cr, pointer.StringValue(resp.JobId))
	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*svcapitypes.Job)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotBatchJob)
	}

	// for Job only tags are updatable
	return managed.ExternalUpdate{}, svcutils.UpdateTagsForResource(ctx, e.client, cr.Spec.ForProvider.Tags, cr.Status.AtProvider.JobArn)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*svcapitypes.Job)
	if !ok {
		return errors.New(errNotBatchJob)
	}

	// no check possible for a "deleting"-like status at Provider -> many delete-requests

	cr.Status.SetConditions(xpv1.Deleting())
	// No terminate-request needed, when Job is already finished
	if cr.Status.AtProvider.Status == pointer.ToOrNilIfZeroValue(svcsdk.JobStatusFailed) ||
		cr.Status.AtProvider.Status == pointer.ToOrNilIfZeroValue(svcsdk.JobStatusSucceeded) {
		return nil
	}

	// terminate includes cancel (seems no need to differentiate)

	// Termination/Cancelation can take a while
	// e.g. when Job is PENDING bc it's dependend on another Job to finish,
	// termination seems to not get through until dependend Job is fisnished ... (-> tested on AWS Console)

	_, err := e.client.TerminateJobWithContext(ctx, generateTerminateJobInput(cr, pointer.ToOrNilIfZeroValue("Terminated for crossplane deletion")))
	return errorutils.Wrap(resource.Ignore(isErrorNotFound, err), errTerminateJob)
}

func (e *external) lateInitialize(spec, current *svcapitypes.JobParameters) { //nolint:gocyclo

	if spec.RetryStrategy == nil && current.RetryStrategy != nil {
		spec.RetryStrategy = current.RetryStrategy
	}

	// The ContainerOverides (or NodeOverrides) fields not defined by the Job specs,
	// are auto-filled through the underlying JobDefinition by AWS

	if spec.ContainerOverrides != nil || current.ContainerOverrides != nil {
		if spec.ContainerOverrides == nil {
			spec.ContainerOverrides = &svcapitypes.ContainerOverrides{}
		}
		lateInitContainerOverrides(spec.ContainerOverrides, current.ContainerOverrides)
	}

	if spec.NodeOverrides != nil || current.NodeOverrides != nil {
		if spec.NodeOverrides == nil {
			spec.NodeOverrides = &svcapitypes.NodeOverrides{}
			spec.NodeOverrides.NodePropertyOverrides = make([]*svcapitypes.NodePropertyOverride, len(current.NodeOverrides.NodePropertyOverrides))
		}

		if current.NodeOverrides != nil {
			spec.NodeOverrides.NumNodes = pointer.LateInitialize(spec.NodeOverrides.NumNodes, current.NodeOverrides.NumNodes)

			if current.NodeOverrides.NodePropertyOverrides != nil {

				for i, noProOver := range current.NodeOverrides.NodePropertyOverrides {
					specNoProOver := &svcapitypes.NodePropertyOverride{}

					if spec.NodeOverrides.NodePropertyOverrides[i] != nil {
						specNoProOver = spec.NodeOverrides.NodePropertyOverrides[i]
					}

					if specNoProOver.ContainerOverrides == nil {
						specNoProOver.ContainerOverrides = &svcapitypes.ContainerOverrides{}
					}
					lateInitContainerOverrides(specNoProOver.ContainerOverrides, noProOver.ContainerOverrides)
					specNoProOver.TargetNodes = pointer.LateInitializeValueFromPtr(specNoProOver.TargetNodes, pointer.ToOrNilIfZeroValue(noProOver.TargetNodes))
					spec.NodeOverrides.NodePropertyOverrides[i] = specNoProOver
				}
			}
		}
	}

	if spec.Timeout == nil && current.Timeout != nil {
		spec.Timeout = current.Timeout
	}
}

// Helper for lateInitialize() with ContainerOverrides
func lateInitContainerOverrides(spec, current *svcapitypes.ContainerOverrides) {

	spec.Command = pointer.LateInitializeSlice(spec.Command, current.Command)
	spec.InstanceType = pointer.LateInitialize(spec.InstanceType, current.InstanceType)
	if spec.Environment == nil && current.Environment != nil {
		env := []*svcapitypes.KeyValuePair{}
		for _, pair := range current.Environment {
			env = append(env, &svcapitypes.KeyValuePair{
				Name:  pair.Name,
				Value: pair.Value,
			})
		}
		spec.Environment = env
	}
	if spec.ResourceRequirements == nil && current.ResourceRequirements != nil {
		resReqs := []*svcapitypes.ResourceRequirement{}
		for _, resReq := range current.ResourceRequirements {
			resReqs = append(resReqs, &svcapitypes.ResourceRequirement{
				ResourceType: resReq.ResourceType,
				Value:        resReq.Value,
			})
		}
		spec.ResourceRequirements = resReqs
	}
}

// IsErrorNotFound helper function to test for ResourceNotFoundException error.
func isErrorNotFound(err error) bool {
	var awsErr smithy.APIError
	return errors.As(err, &awsErr) && awsErr.ErrorCode() == "JobNotFoundException"
}
