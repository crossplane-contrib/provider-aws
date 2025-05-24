package service

import (
	svcsdk "github.com/aws/aws-sdk-go/service/ecs"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/ecs/v1alpha1"
)

// GenerateServiceCustom returns the current state in the form of *svcapitypes.Service with custom prarameters set
func GenerateServiceCustom(resp *svcsdk.DescribeServicesOutput) *svcapitypes.Service { //nolint:gocyclo
	if len(resp.Services) != 1 {
		return nil
	}

	out := resp.Services[0]
	service := GenerateService(resp)
	params := &service.Spec.ForProvider

	if out.NetworkConfiguration != nil && out.NetworkConfiguration.AwsvpcConfiguration != nil {
		params.CustomServiceParameters.NetworkConfiguration = &svcapitypes.CustomNetworkConfiguration{
			AWSvpcConfiguration: &svcapitypes.CustomAWSVPCConfiguration{
				AssignPublicIP: out.NetworkConfiguration.AwsvpcConfiguration.AssignPublicIp,
				SecurityGroups: out.NetworkConfiguration.AwsvpcConfiguration.SecurityGroups,
				Subnets:        out.NetworkConfiguration.AwsvpcConfiguration.Subnets,
			},
		}
	}

	if out.LoadBalancers != nil {
		params.CustomServiceParameters.LoadBalancers = []*svcapitypes.CustomLoadBalancer{}
		for _, lb := range out.LoadBalancers {
			params.CustomServiceParameters.LoadBalancers = append(params.CustomServiceParameters.LoadBalancers, &svcapitypes.CustomLoadBalancer{
				ContainerName:    lb.ContainerName,
				ContainerPort:    lb.ContainerPort,
				LoadBalancerName: lb.LoadBalancerName,
				TargetGroupARN:   lb.TargetGroupArn,
			})
		}
	}

	if out.TaskDefinition != nil {
		params.CustomServiceParameters.TaskDefinition = out.TaskDefinition
	}

	// Get ServiceConnectConfiguration from last deployment
	if len(out.Deployments) > 0 {
		current := out.Deployments[0]
		if current != nil && current.ServiceConnectConfiguration != nil {
			params.ServiceConnectConfiguration = &svcapitypes.ServiceConnectConfiguration{
				Enabled:   current.ServiceConnectConfiguration.Enabled,
				Namespace: current.ServiceConnectConfiguration.Namespace,
			}

			if current.ServiceConnectConfiguration.LogConfiguration != nil {
				params.ServiceConnectConfiguration.LogConfiguration = &svcapitypes.LogConfiguration{
					LogDriver: current.ServiceConnectConfiguration.LogConfiguration.LogDriver,
					Options:   current.ServiceConnectConfiguration.LogConfiguration.Options,
				}

				for _, so := range params.ServiceConnectConfiguration.LogConfiguration.SecretOptions {
					params.ServiceConnectConfiguration.LogConfiguration.SecretOptions = append(params.ServiceConnectConfiguration.LogConfiguration.SecretOptions, &svcapitypes.Secret{
						Name:      so.Name,
						ValueFrom: so.ValueFrom,
					})
				}
			}

			for _, s := range current.ServiceConnectConfiguration.Services {
				service := &svcapitypes.ServiceConnectService{
					DiscoveryName:       s.DiscoveryName,
					IngressPortOverride: s.IngressPortOverride,
					PortName:            s.PortName,
				}

				for _, ca := range s.ClientAliases {
					service.ClientAliases = append(service.ClientAliases, &svcapitypes.ServiceConnectClientAlias{
						DNSName: ca.DnsName,
						Port:    ca.Port,
					})
				}

				params.ServiceConnectConfiguration.Services = append(params.ServiceConnectConfiguration.Services, service)
			}
		}
	}

	return service
}
