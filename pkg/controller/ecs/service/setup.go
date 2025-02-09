package service

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/ecs"
	svcsdkapi "github.com/aws/aws-sdk-go/service/ecs/ecsiface"
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
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/ecs/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

type custom struct {
	kube   client.Client
	client svcsdkapi.ECSAPI
}

// SetupService adds a controller that reconciles Service.
func SetupService(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.ServiceGroupKind)
	opts := []option{
		func(e *external) {
			c := &custom{client: e.client, kube: e.kube}
			e.preObserve = preObserve
			e.postObserve = c.postObserve
			e.preCreate = preCreate
			e.preUpdate = preUpdate
			e.preDelete = preDelete
			e.isUpToDate = isUpToDate
			e.lateInitialize = lateInitialize
		},
	}
	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithTypedExternalConnector(&connector{kube: mgr.GetClient(), opts: opts}),
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

func isUpToDate(context context.Context, service *svcapitypes.Service, output *svcsdk.DescribeServicesOutput) (bool, string, error) {
	if len(output.Services) != 1 {
		return false, "", nil
	}

	t := service.Spec.ForProvider.DeepCopy()
	c := GenerateServiceCustom(output).Spec.ForProvider.DeepCopy()

	tags := func(a, b *svcapitypes.Tag) bool { return aws.StringValue(a.Key) < aws.StringValue(b.Key) }
	stringpointer := func(a, b *string) bool { return aws.StringValue(a) < aws.StringValue(b) }
	keyValuePair := func(a, b *svcsdk.KeyValuePair) bool { return aws.StringValue(a.Name) < aws.StringValue(b.Name) }

	diff := cmp.Diff(c, t,
		cmpopts.EquateEmpty(),
		cmpopts.SortSlices(tags),
		cmpopts.SortSlices(stringpointer),
		cmpopts.SortSlices(keyValuePair),
		// Not present in DescribeServicesOutput
		cmpopts.IgnoreFields(svcapitypes.ServiceParameters{}, "Region"),
		cmpopts.IgnoreFields(svcapitypes.CustomServiceParameters{}, "Cluster"),
		cmpopts.IgnoreFields(svcapitypes.CustomServiceParameters{}, "ForceDeletion"),
		cmpopts.IgnoreTypes(&xpv1.Reference{}, &xpv1.Selector{}, []xpv1.Reference{}))

	return diff == "", diff, nil
}

func lateInitialize(in *svcapitypes.ServiceParameters, out *svcsdk.DescribeServicesOutput) error { //nolint:gocyclo
	if len(out.Services) != 1 {
		return nil
	}
	o := out.Services[0]

	if in.PlatformVersion == nil {
		in.PlatformVersion = o.PlatformVersion
	}

	if in.EnableECSManagedTags == nil {
		in.EnableECSManagedTags = o.EnableECSManagedTags
	}

	if in.SchedulingStrategy == nil {
		in.SchedulingStrategy = o.SchedulingStrategy
	}

	if in.HealthCheckGracePeriodSeconds == nil {
		in.HealthCheckGracePeriodSeconds = o.HealthCheckGracePeriodSeconds
	}

	if in.DeploymentController == nil && o.DeploymentController != nil {
		in.DeploymentController = &svcapitypes.DeploymentController{
			Type: o.DeploymentController.Type,
		}
	}

	if o.DeploymentConfiguration != nil {
		if in.DeploymentConfiguration == nil {
			in.DeploymentConfiguration = &svcapitypes.DeploymentConfiguration{}
		}

		if in.DeploymentConfiguration.MaximumPercent == nil {
			in.DeploymentConfiguration.MaximumPercent = o.DeploymentConfiguration.MaximumPercent
		}
		if in.DeploymentConfiguration.MinimumHealthyPercent == nil {
			in.DeploymentConfiguration.MinimumHealthyPercent = o.DeploymentConfiguration.MinimumHealthyPercent
		}
	}

	if in.EnableECSManagedTags == nil {
		in.EnableECSManagedTags = o.EnableECSManagedTags
	}

	if in.PropagateTags == nil {
		in.PropagateTags = o.PropagateTags
	}

	if o.NetworkConfiguration != nil {
		if in.CustomServiceParameters.NetworkConfiguration == nil {
			in.CustomServiceParameters.NetworkConfiguration = &svcapitypes.CustomNetworkConfiguration{}
		}

		if o.NetworkConfiguration.AwsvpcConfiguration != nil {
			if in.CustomServiceParameters.NetworkConfiguration.AWSvpcConfiguration == nil {
				in.CustomServiceParameters.NetworkConfiguration.AWSvpcConfiguration = &svcapitypes.CustomAWSVPCConfiguration{}
			}

			if in.CustomServiceParameters.NetworkConfiguration.AWSvpcConfiguration.AssignPublicIP == nil {
				in.CustomServiceParameters.NetworkConfiguration.AWSvpcConfiguration.AssignPublicIP = o.NetworkConfiguration.AwsvpcConfiguration.AssignPublicIp
			}
		}
	}

	return nil
}

func preObserve(_ context.Context, cr *svcapitypes.Service, obj *svcsdk.DescribeServicesInput) error {
	obj.Cluster = cr.Spec.ForProvider.Cluster
	obj.Services = []*string{aws.String(meta.GetExternalName(cr))}
	obj.Include = []*string{ptr.To("TAGS")}

	if err := obj.Validate(); err != nil {
		return err
	}
	return nil
}

func (e *custom) postObserve(_ context.Context, cr *svcapitypes.Service, resp *svcsdk.DescribeServicesOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return obs, err
	}
	if len(resp.Services) == 0 {
		obs.ResourceExists = false
		return obs, err
	}

	listTasksOutput, err := e.client.ListTasks(&svcsdk.ListTasksInput{
		Cluster:     cr.Spec.ForProvider.Cluster,
		ServiceName: aws.String(meta.GetExternalName(cr)),
	})
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "ListTasks failed")
	}
	cr.Status.AtProvider.TaskARNs = listTasksOutput.TaskArns

	switch aws.StringValue(resp.Services[0].Status) {
	case "ACTIVE":
		if resp.Services[0].DesiredCount == nil || resp.Services[0].RunningCount == nil || *resp.Services[0].DesiredCount != *resp.Services[0].RunningCount {
			cr.SetConditions(xpv1.Creating())
		} else {
			cr.SetConditions(xpv1.Available())
		}
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

func preUpdate(context context.Context, cr *svcapitypes.Service, obj *svcsdk.UpdateServiceInput) error {
	obj.Cluster = cr.Spec.ForProvider.Cluster
	obj.Service = aws.String(meta.GetExternalName(cr))

	mapUpdateServiceInput(cr, obj)

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

// Helper func to map the values of the APIs Service to the
// SDKs UpdateServiceInput.
func mapUpdateServiceInput(cr *svcapitypes.Service, obj *svcsdk.UpdateServiceInput) { //nolint:gocyclo
	// From: ecs/service/zz_conversions.go
	// Removed DeploymentConfiguration, LaunchType, Role, SchedulingStrategy and Tags as they cannot be updated
	if cr.Spec.ForProvider.CapacityProviderStrategy != nil {
		f0 := []*svcsdk.CapacityProviderStrategyItem{}
		for _, f0iter := range cr.Spec.ForProvider.CapacityProviderStrategy {
			f0elem := &svcsdk.CapacityProviderStrategyItem{}
			if f0iter.Base != nil {
				f0elem.SetBase(*f0iter.Base)
			}
			if f0iter.CapacityProvider != nil {
				f0elem.SetCapacityProvider(*f0iter.CapacityProvider)
			}
			if f0iter.Weight != nil {
				f0elem.SetWeight(*f0iter.Weight)
			}
			f0 = append(f0, f0elem)
		}
		obj.SetCapacityProviderStrategy(f0)
	}
	if cr.Spec.ForProvider.DeploymentConfiguration != nil {
		f1 := &svcsdk.DeploymentConfiguration{}
		if cr.Spec.ForProvider.DeploymentConfiguration.Alarms != nil {
			f1f0 := &svcsdk.DeploymentAlarms{}
			if cr.Spec.ForProvider.DeploymentConfiguration.Alarms.AlarmNames != nil {
				f1f0f0 := []*string{}
				for _, f1f0f0iter := range cr.Spec.ForProvider.DeploymentConfiguration.Alarms.AlarmNames {
					f1f0f0elem := *f1f0f0iter
					f1f0f0 = append(f1f0f0, &f1f0f0elem)
				}
				f1f0.SetAlarmNames(f1f0f0)
			}
			if cr.Spec.ForProvider.DeploymentConfiguration.Alarms.Enable != nil {
				f1f0.SetEnable(*cr.Spec.ForProvider.DeploymentConfiguration.Alarms.Enable)
			}
			if cr.Spec.ForProvider.DeploymentConfiguration.Alarms.Rollback != nil {
				f1f0.SetRollback(*cr.Spec.ForProvider.DeploymentConfiguration.Alarms.Rollback)
			}
			f1.SetAlarms(f1f0)
		}
		if cr.Spec.ForProvider.DeploymentConfiguration.DeploymentCircuitBreaker != nil {
			f1f1 := &svcsdk.DeploymentCircuitBreaker{}
			if cr.Spec.ForProvider.DeploymentConfiguration.DeploymentCircuitBreaker.Enable != nil {
				f1f1.SetEnable(*cr.Spec.ForProvider.DeploymentConfiguration.DeploymentCircuitBreaker.Enable)
			}
			if cr.Spec.ForProvider.DeploymentConfiguration.DeploymentCircuitBreaker.Rollback != nil {
				f1f1.SetRollback(*cr.Spec.ForProvider.DeploymentConfiguration.DeploymentCircuitBreaker.Rollback)
			}
			f1.SetDeploymentCircuitBreaker(f1f1)
		}
		if cr.Spec.ForProvider.DeploymentConfiguration.MaximumPercent != nil {
			f1.SetMaximumPercent(*cr.Spec.ForProvider.DeploymentConfiguration.MaximumPercent)
		}
		if cr.Spec.ForProvider.DeploymentConfiguration.MinimumHealthyPercent != nil {
			f1.SetMinimumHealthyPercent(*cr.Spec.ForProvider.DeploymentConfiguration.MinimumHealthyPercent)
		}
		obj.SetDeploymentConfiguration(f1)
	}
	if cr.Spec.ForProvider.DesiredCount != nil {
		obj.SetDesiredCount(*cr.Spec.ForProvider.DesiredCount)
	}
	if cr.Spec.ForProvider.EnableECSManagedTags != nil {
		obj.SetEnableECSManagedTags(*cr.Spec.ForProvider.EnableECSManagedTags)
	}
	if cr.Spec.ForProvider.EnableExecuteCommand != nil {
		obj.SetEnableExecuteCommand(*cr.Spec.ForProvider.EnableExecuteCommand)
	}
	if cr.Spec.ForProvider.HealthCheckGracePeriodSeconds != nil {
		obj.SetHealthCheckGracePeriodSeconds(*cr.Spec.ForProvider.HealthCheckGracePeriodSeconds)
	}
	if cr.Spec.ForProvider.PlacementConstraints != nil {
		f8 := []*svcsdk.PlacementConstraint{}
		for _, f8iter := range cr.Spec.ForProvider.PlacementConstraints {
			f8elem := &svcsdk.PlacementConstraint{}
			if f8iter.Expression != nil {
				f8elem.SetExpression(*f8iter.Expression)
			}
			if f8iter.Type != nil {
				f8elem.SetType(*f8iter.Type)
			}
			f8 = append(f8, f8elem)
		}
		obj.SetPlacementConstraints(f8)
	}
	if cr.Spec.ForProvider.PlacementStrategy != nil {
		f9 := []*svcsdk.PlacementStrategy{}
		for _, f9iter := range cr.Spec.ForProvider.PlacementStrategy {
			f9elem := &svcsdk.PlacementStrategy{}
			if f9iter.Field != nil {
				f9elem.SetField(*f9iter.Field)
			}
			if f9iter.Type != nil {
				f9elem.SetType(*f9iter.Type)
			}
			f9 = append(f9, f9elem)
		}
		obj.SetPlacementStrategy(f9)
	}
	if cr.Spec.ForProvider.PlatformVersion != nil {
		obj.SetPlatformVersion(*cr.Spec.ForProvider.PlatformVersion)
	}
	if cr.Spec.ForProvider.PropagateTags != nil {
		obj.SetPropagateTags(*cr.Spec.ForProvider.PropagateTags)
	}
	if cr.Spec.ForProvider.ServiceConnectConfiguration != nil {
		f14 := &svcsdk.ServiceConnectConfiguration{}
		if cr.Spec.ForProvider.ServiceConnectConfiguration.Enabled != nil {
			f14.SetEnabled(*cr.Spec.ForProvider.ServiceConnectConfiguration.Enabled)
		}
		if cr.Spec.ForProvider.ServiceConnectConfiguration.LogConfiguration != nil {
			f14f1 := &svcsdk.LogConfiguration{}
			if cr.Spec.ForProvider.ServiceConnectConfiguration.LogConfiguration.LogDriver != nil {
				f14f1.SetLogDriver(*cr.Spec.ForProvider.ServiceConnectConfiguration.LogConfiguration.LogDriver)
			}
			if cr.Spec.ForProvider.ServiceConnectConfiguration.LogConfiguration.Options != nil {
				f14f1f1 := map[string]*string{}
				for f14f1f1key, f14f1f1valiter := range cr.Spec.ForProvider.ServiceConnectConfiguration.LogConfiguration.Options {
					f14f1f1val := *f14f1f1valiter
					f14f1f1[f14f1f1key] = &f14f1f1val
				}
				f14f1.SetOptions(f14f1f1)
			}
			if cr.Spec.ForProvider.ServiceConnectConfiguration.LogConfiguration.SecretOptions != nil {
				f14f1f2 := []*svcsdk.Secret{}
				for _, f14f1f2iter := range cr.Spec.ForProvider.ServiceConnectConfiguration.LogConfiguration.SecretOptions {
					f14f1f2elem := &svcsdk.Secret{}
					if f14f1f2iter.Name != nil {
						f14f1f2elem.SetName(*f14f1f2iter.Name)
					}
					if f14f1f2iter.ValueFrom != nil {
						f14f1f2elem.SetValueFrom(*f14f1f2iter.ValueFrom)
					}
					f14f1f2 = append(f14f1f2, f14f1f2elem)
				}
				f14f1.SetSecretOptions(f14f1f2)
			}
			f14.SetLogConfiguration(f14f1)
		}
		if cr.Spec.ForProvider.ServiceConnectConfiguration.Namespace != nil {
			f14.SetNamespace(*cr.Spec.ForProvider.ServiceConnectConfiguration.Namespace)
		}
		if cr.Spec.ForProvider.ServiceConnectConfiguration.Services != nil {
			f14f3 := []*svcsdk.ServiceConnectService{}
			for _, f14f3iter := range cr.Spec.ForProvider.ServiceConnectConfiguration.Services {
				f14f3elem := &svcsdk.ServiceConnectService{}
				if f14f3iter.ClientAliases != nil {
					f14f3elemf0 := []*svcsdk.ServiceConnectClientAlias{}
					for _, f14f3elemf0iter := range f14f3iter.ClientAliases {
						f14f3elemf0elem := &svcsdk.ServiceConnectClientAlias{}
						if f14f3elemf0iter.DNSName != nil {
							f14f3elemf0elem.SetDnsName(*f14f3elemf0iter.DNSName)
						}
						if f14f3elemf0iter.Port != nil {
							f14f3elemf0elem.SetPort(*f14f3elemf0iter.Port)
						}
						f14f3elemf0 = append(f14f3elemf0, f14f3elemf0elem)
					}
					f14f3elem.SetClientAliases(f14f3elemf0)
				}
				if f14f3iter.DiscoveryName != nil {
					f14f3elem.SetDiscoveryName(*f14f3iter.DiscoveryName)
				}
				if f14f3iter.IngressPortOverride != nil {
					f14f3elem.SetIngressPortOverride(*f14f3iter.IngressPortOverride)
				}
				if f14f3iter.PortName != nil {
					f14f3elem.SetPortName(*f14f3iter.PortName)
				}
				f14f3 = append(f14f3, f14f3elem)
			}
			f14.SetServices(f14f3)
		}
		obj.SetServiceConnectConfiguration(f14)
	}
	if cr.Spec.ForProvider.ServiceRegistries != nil {
		f15 := []*svcsdk.ServiceRegistry{}
		for _, f15iter := range cr.Spec.ForProvider.ServiceRegistries {
			f15elem := &svcsdk.ServiceRegistry{}
			if f15iter.ContainerName != nil {
				f15elem.SetContainerName(*f15iter.ContainerName)
			}
			if f15iter.ContainerPort != nil {
				f15elem.SetContainerPort(*f15iter.ContainerPort)
			}
			if f15iter.Port != nil {
				f15elem.SetPort(*f15iter.Port)
			}
			if f15iter.RegistryARN != nil {
				f15elem.SetRegistryArn(*f15iter.RegistryARN)
			}
			f15 = append(f15, f15elem)
		}
		obj.SetServiceRegistries(f15)
	}

	// CustomServiceParameters
	if cr.Spec.ForProvider.CustomServiceParameters.NetworkConfiguration != nil {
		obj.NetworkConfiguration = &svcsdk.NetworkConfiguration{
			AwsvpcConfiguration: &svcsdk.AwsVpcConfiguration{
				AssignPublicIp: cr.Spec.ForProvider.CustomServiceParameters.NetworkConfiguration.AWSvpcConfiguration.AssignPublicIP,
				SecurityGroups: cr.Spec.ForProvider.CustomServiceParameters.NetworkConfiguration.AWSvpcConfiguration.SecurityGroups,
				Subnets:        cr.Spec.ForProvider.CustomServiceParameters.NetworkConfiguration.AWSvpcConfiguration.Subnets,
			},
		}
	}

	if cr.Spec.ForProvider.CustomServiceParameters.LoadBalancers != nil {
		obj.LoadBalancers = []*svcsdk.LoadBalancer{}
		for _, lb := range cr.Spec.ForProvider.CustomServiceParameters.LoadBalancers {
			obj.LoadBalancers = append(obj.LoadBalancers, &svcsdk.LoadBalancer{
				ContainerName:    lb.ContainerName,
				ContainerPort:    lb.ContainerPort,
				LoadBalancerName: lb.LoadBalancerName,
				TargetGroupArn:   lb.TargetGroupARN,
			})
		}
	}

	if cr.Spec.ForProvider.TaskDefinition != nil {
		obj.TaskDefinition = cr.Spec.ForProvider.TaskDefinition
	}
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
