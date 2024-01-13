/*
Copyright 2022 The Crossplane Authors.

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

package target

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awselasticloadbalancingv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/elbv2/manualv1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errNotElasticloadbalancingv2Target = "managed resource is not an elbv2 Target resource"
	errRegisterTargetFailed            = "failed to register target"
	errDeregisterTargetFailed          = "failed to deregister target"
	errDescribeTargetHealthFailed      = "failed to describe target health"
)

// SetupTarget adds a controller that reconciles Targets.
func SetupTarget(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(manualv1alpha1.TargetKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: awselasticloadbalancingv2.NewFromConfig}),
		managed.WithInitializers(),
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
		resource.ManagedKind(manualv1alpha1.TargetGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&manualv1alpha1.Target{}).
		Complete(r)
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config, optFns ...func(*awselasticloadbalancingv2.Options)) *awselasticloadbalancingv2.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*manualv1alpha1.Target)
	if !ok {
		return nil, errors.New(errNotElasticloadbalancingv2Target)
	}
	cfg, err := connectaws.GetConfig(ctx, c.kube, mg, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, err
	}
	return &external{client: c.newClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	client *awselasticloadbalancingv2.Client
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*manualv1alpha1.Target)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotElasticloadbalancingv2Target)
	}

	// Set the external-name to the LambdARN if not set.
	if meta.GetExternalName(cr) == "" {
		meta.SetExternalName(cr, pointer.StringValue(cr.Spec.ForProvider.LambdaARN))
	}
	res, err := e.client.DescribeTargetHealth(ctx, &awselasticloadbalancingv2.DescribeTargetHealthInput{
		TargetGroupArn: cr.Spec.ForProvider.TargetGroupARN,
		Targets: []types.TargetDescription{
			{
				Id:               pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
				AvailabilityZone: cr.Spec.ForProvider.AvailabilityZone,
				Port:             cr.Spec.ForProvider.Port,
			},
		},
	})
	if err != nil || len(res.TargetHealthDescriptions) == 0 {
		return managed.ExternalObservation{}, errorutils.Wrap(err, errDescribeTargetHealthFailed)
	}

	cr.Status.AtProvider = generateTargetObservation(res)
	switch cr.Status.AtProvider.GetState() {
	case manualv1alpha1.TargetStatusHealthy:
		cr.Status.SetConditions(xpv1.Available())
	case manualv1alpha1.TargetStatusInitial:
		cr.Status.SetConditions(xpv1.Creating())
	case manualv1alpha1.TargetStatusDraining:
		cr.Status.SetConditions(xpv1.Deleting())
	default:
		cr.Status.SetConditions(xpv1.Unavailable())
	}

	return managed.ExternalObservation{
		ResourceExists:   cr.Status.AtProvider.GetReason() != string(types.TargetHealthReasonEnumNotRegistered),
		ResourceUpToDate: true, // Targets cannot be updated
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*manualv1alpha1.Target)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotElasticloadbalancingv2Target)
	}
	cr.SetConditions(xpv1.Creating())
	_, err := e.client.RegisterTargets(ctx, &awselasticloadbalancingv2.RegisterTargetsInput{
		TargetGroupArn: cr.Spec.ForProvider.TargetGroupARN,
		Targets: []types.TargetDescription{
			{
				Id:               pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
				AvailabilityZone: cr.Spec.ForProvider.AvailabilityZone,
				Port:             cr.Spec.ForProvider.Port,
			},
		},
	})
	return managed.ExternalCreation{}, errorutils.Wrap(err, errRegisterTargetFailed)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	// Never invoked.
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*manualv1alpha1.Target)
	if !ok {
		return errors.New(errNotElasticloadbalancingv2Target)
	}
	cr.SetConditions(xpv1.Deleting())
	// Check if the target is already unregistered.
	if cr.Status.AtProvider.TargetHealth != nil && pointer.StringValue(cr.Status.AtProvider.TargetHealth.State) == manualv1alpha1.TargetStatusDraining {
		return nil
	}
	_, err := e.client.DeregisterTargets(ctx, &awselasticloadbalancingv2.DeregisterTargetsInput{
		TargetGroupArn: cr.Spec.ForProvider.TargetGroupARN,
		Targets: []types.TargetDescription{
			{
				Id:               pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
				AvailabilityZone: cr.Spec.ForProvider.AvailabilityZone,
				Port:             cr.Spec.ForProvider.Port,
			},
		},
	})
	return errorutils.Wrap(err, errDeregisterTargetFailed)
}

func generateTargetObservation(i *awselasticloadbalancingv2.DescribeTargetHealthOutput) manualv1alpha1.TargetObservation {
	o := manualv1alpha1.TargetObservation{}
	if i == nil || len(i.TargetHealthDescriptions) == 0 {
		return o
	}
	desc := i.TargetHealthDescriptions[0]
	o.HealthCheckPort = desc.HealthCheckPort
	if desc.TargetHealth != nil {
		o.TargetHealth = &manualv1alpha1.TargetHealth{
			Description: desc.TargetHealth.Description,
			Reason:      (*string)(&desc.TargetHealth.Reason),
			State:       (*string)(&desc.TargetHealth.State),
		}
	}
	return o
}
