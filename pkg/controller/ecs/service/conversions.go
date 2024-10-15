package service

import (
	svcsdk "github.com/aws/aws-sdk-go/service/ecs"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/ecs/v1alpha1"
)

// GenerateServiceCustom returns the current state in the form of *svcapitypes.Service with custom prarameters set
func GenerateServiceCustom(resp *svcsdk.DescribeServicesOutput) *svcapitypes.Service {
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

	return service
}
