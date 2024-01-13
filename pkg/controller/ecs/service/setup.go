package service

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/ecs"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/ecs/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupService adds a controller that reconciles Service.
func SetupService(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.ServiceGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preCreate = preCreate
			e.preDelete = preDelete
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
		resource.ManagedKind(svcapitypes.ServiceGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.Service{}).
		Complete(r)
}

func preObserve(_ context.Context, cr *svcapitypes.Service, obj *svcsdk.DescribeServicesInput) error {
	obj.Cluster = cr.Spec.ForProvider.Cluster
	obj.Services = []*string{aws.String(meta.GetExternalName(cr))}

	if err := obj.Validate(); err != nil {
		return err
	}
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.Service, resp *svcsdk.DescribeServicesOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return obs, err
	}
	if len(resp.Services) == 0 {
		obs.ResourceExists = false
		return obs, err
	}

	switch aws.StringValue(resp.Services[0].Status) {
	case "ACTIVE":
		cr.SetConditions(xpv1.Available())
	case "DRAINING":
		cr.SetConditions(xpv1.Deleting())
	case "INACTIVE":
		// Deleted services can still be described in the API and show up with
		// an INACTIVE status, which means we need to re-create the service.
		obs.ResourceExists = false
		cr.SetConditions(xpv1.Unavailable())
	}
	return obs, nil
}

func preCreate(_ context.Context, cr *svcapitypes.Service, obj *svcsdk.CreateServiceInput) error {
	obj.ClientToken = aws.String(string(cr.UID))
	obj.Cluster = cr.Spec.ForProvider.Cluster
	obj.LoadBalancers = generateLoadBalancers(cr)
	obj.NetworkConfiguration = generateNetworkConfiguration(cr)
	obj.SetServiceName(meta.GetExternalName(cr))
	obj.TaskDefinition = cr.Spec.ForProvider.TaskDefinition

	if err := obj.Validate(); err != nil {
		return err
	}
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.Service, obj *svcsdk.DeleteServiceInput) (bool, error) {

	obj.SetForce(cr.Spec.ForProvider.ForceDeletion)
	obj.Cluster = cr.Spec.ForProvider.Cluster
	obj.SetService(meta.GetExternalName(cr))

	if err := obj.Validate(); err != nil {
		return false, err
	}
	return false, nil
}

// Helper func to generate SDK LoadBalancer types from the crossplane
// types included in the Service api.
func generateLoadBalancers(cr *svcapitypes.Service) []*svcsdk.LoadBalancer {
	loadBalancers := []*svcsdk.LoadBalancer{}

	if cr.Spec.ForProvider.LoadBalancers == nil {
		return loadBalancers
	}

	for _, loadBalancer := range cr.Spec.ForProvider.LoadBalancers {
		convertedLB := &svcsdk.LoadBalancer{}
		convertedLB.ContainerName = loadBalancer.ContainerName
		convertedLB.ContainerPort = loadBalancer.ContainerPort
		convertedLB.LoadBalancerName = loadBalancer.LoadBalancerName
		convertedLB.TargetGroupArn = loadBalancer.TargetGroupARN

		loadBalancers = append(loadBalancers, convertedLB)
	}
	return loadBalancers
}

// Helper func to generate SDK NetworkConfiguration types from the crossplane
// types included in the Service api.
func generateNetworkConfiguration(cr *svcapitypes.Service) *svcsdk.NetworkConfiguration {
	networkConfiguration := &svcsdk.NetworkConfiguration{}

	if cr.Spec.ForProvider.NetworkConfiguration == nil {
		return networkConfiguration
	}

	if cr.Spec.ForProvider.NetworkConfiguration.AWSvpcConfiguration != nil {
		convertedVPCconf := &svcsdk.AwsVpcConfiguration{}
		convertedVPCconf.AssignPublicIp = cr.Spec.ForProvider.NetworkConfiguration.AWSvpcConfiguration.AssignPublicIP
		convertedVPCconf.SecurityGroups = cr.Spec.ForProvider.NetworkConfiguration.AWSvpcConfiguration.SecurityGroups
		convertedVPCconf.Subnets = cr.Spec.ForProvider.NetworkConfiguration.AWSvpcConfiguration.Subnets
		networkConfiguration.AwsvpcConfiguration = convertedVPCconf
	}

	return networkConfiguration
}
