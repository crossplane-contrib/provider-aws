package service

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/ecs"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/ecs/v1alpha1"
)

func TestGenerateLoadBalancers(t *testing.T) {
	cases := map[string]struct {
		reason string
		cr     *svcapitypes.Service
		want   []*svcsdk.LoadBalancer
	}{
		"TestEmptyLoadBalancerSpec": {
			"When passed an empty loadbalancer param, generate an empty slice",
			&svcapitypes.Service{
				Spec: svcapitypes.ServiceSpec{
					ForProvider: svcapitypes.ServiceParameters{
						CustomServiceParameters: svcapitypes.CustomServiceParameters{
							LoadBalancers: []*svcapitypes.CustomLoadBalancer{},
						},
					},
				},
			},
			[]*svcsdk.LoadBalancer{},
		},
		"TestLoadBalancerSpec": {
			"When passed a loadbalancerparam, generate a slice of loadbalancer types",
			&svcapitypes.Service{
				Spec: svcapitypes.ServiceSpec{
					ForProvider: svcapitypes.ServiceParameters{
						CustomServiceParameters: svcapitypes.CustomServiceParameters{
							LoadBalancers: []*svcapitypes.CustomLoadBalancer{
								{
									ContainerName:    aws.String("test-container"),
									ContainerPort:    aws.Int64(443),
									LoadBalancerName: aws.String("test-loadbalancer"),
									TargetGroupARN:   aws.String("arn:::test-listener"),
									TargetGroupARNRef: &xpv1.Reference{
										Name: "test-listener",
									},
									TargetGroupARNSelector: &xpv1.Selector{
										MatchLabels: map[string]string{
											"crossplane.io/name": "test-loadbalancer",
										},
									},
								},
							},
						},
					},
				},
			},
			[]*svcsdk.LoadBalancer{
				{
					ContainerName:    aws.String("test-container"),
					ContainerPort:    aws.Int64(443),
					LoadBalancerName: aws.String("test-loadbalancer"),
					TargetGroupArn:   aws.String("arn:::test-listener"),
				},
			},
		},
		"TestMultipleLoadBalancerSpec": {
			"When passed multiple loadbalancerparams, generate a slice of loadbalancer types",
			&svcapitypes.Service{
				Spec: svcapitypes.ServiceSpec{
					ForProvider: svcapitypes.ServiceParameters{
						CustomServiceParameters: svcapitypes.CustomServiceParameters{
							LoadBalancers: []*svcapitypes.CustomLoadBalancer{
								{
									ContainerName:    aws.String("test-container"),
									ContainerPort:    aws.Int64(443),
									LoadBalancerName: aws.String("test-loadbalancer"),
									TargetGroupARN:   aws.String("arn:::test-listener"),
									TargetGroupARNRef: &xpv1.Reference{
										Name: "test-listener",
									},
									TargetGroupARNSelector: &xpv1.Selector{
										MatchLabels: map[string]string{
											"crossplane.io/name": "test-loadbalancer",
										},
									},
								},
								{
									ContainerName:  aws.String("test-container2"),
									ContainerPort:  aws.Int64(443),
									TargetGroupARN: aws.String("arn:::test-listener"),
									TargetGroupARNRef: &xpv1.Reference{
										Name: "test-listener",
									},
									TargetGroupARNSelector: &xpv1.Selector{
										MatchLabels: map[string]string{
											"crossplane.io/name": "test-loadbalancer",
										},
									},
								},
								{
									ContainerName:    aws.String("test-container3"),
									ContainerPort:    aws.Int64(443),
									LoadBalancerName: aws.String("test-loadbalancer"),
								},
							},
						},
					},
				},
			},
			[]*svcsdk.LoadBalancer{
				{
					ContainerName:    aws.String("test-container"),
					ContainerPort:    aws.Int64(443),
					LoadBalancerName: aws.String("test-loadbalancer"),
					TargetGroupArn:   aws.String("arn:::test-listener"),
				},
				{
					ContainerName:  aws.String("test-container2"),
					ContainerPort:  aws.Int64(443),
					TargetGroupArn: aws.String("arn:::test-listener"),
				},
				{
					ContainerName:    aws.String("test-container3"),
					ContainerPort:    aws.Int64(443),
					LoadBalancerName: aws.String("test-loadbalancer"),
				},
			},
		},
	}

	for name, tc := range cases {

		t.Run(name, func(t *testing.T) {
			got := generateLoadBalancers(tc.cr)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("%s\nExample(...): -want, +got:\n%s", tc.reason, diff)
			}
		})
	}
}

func TestGenerateNetworkConfiguraiton(t *testing.T) {
	cases := map[string]struct {
		reason string
		cr     *svcapitypes.Service
		want   *svcsdk.NetworkConfiguration
	}{
		"TestEmptyNetworkSpec": {
			"When passed an empty networkconfiguration param, generate an empty struct",
			&svcapitypes.Service{
				Spec: svcapitypes.ServiceSpec{
					ForProvider: svcapitypes.ServiceParameters{
						CustomServiceParameters: svcapitypes.CustomServiceParameters{
							NetworkConfiguration: &svcapitypes.CustomNetworkConfiguration{},
						},
					},
				},
			},
			&svcsdk.NetworkConfiguration{},
		},
		"TestEmptySubnets": {
			"When passed a networkconfiguration with empty subnets, generate an networkconfiguration struct with empty subnets",
			&svcapitypes.Service{
				Spec: svcapitypes.ServiceSpec{
					ForProvider: svcapitypes.ServiceParameters{
						CustomServiceParameters: svcapitypes.CustomServiceParameters{
							NetworkConfiguration: &svcapitypes.CustomNetworkConfiguration{
								AWSvpcConfiguration: &svcapitypes.CustomAWSVPCConfiguration{
									AssignPublicIP: aws.String("ENABLED"),
									SecurityGroups: []*string{
										aws.String("sg-12345"),
										aws.String("sg-45678"),
									},
									SecurityGroupRefs: []xpv1.Reference{
										{
											Name: "security-group1234",
										},
										{
											Name: "security-group5678",
										},
									},
									SecurityGroupSelector: &xpv1.Selector{
										MatchLabels: map[string]string{
											"createdBy": "test",
										},
									},
								},
							},
						},
					},
				},
			},
			&svcsdk.NetworkConfiguration{
				AwsvpcConfiguration: &svcsdk.AwsVpcConfiguration{
					AssignPublicIp: aws.String("ENABLED"),
					SecurityGroups: []*string{
						aws.String("sg-12345"),
						aws.String("sg-45678"),
					},
				},
			},
		},
		"TestEmptySecurityGroups": {
			"When passed a networkconfiguration with empty securitygroups, generate an networkconfiguration struct with empty securitygroups",
			&svcapitypes.Service{
				Spec: svcapitypes.ServiceSpec{
					ForProvider: svcapitypes.ServiceParameters{
						CustomServiceParameters: svcapitypes.CustomServiceParameters{
							NetworkConfiguration: &svcapitypes.CustomNetworkConfiguration{
								AWSvpcConfiguration: &svcapitypes.CustomAWSVPCConfiguration{
									Subnets: []*string{
										aws.String("subnet-12345"),
										aws.String("subnet-45678"),
									},
									SubnetRefs: []xpv1.Reference{
										{
											Name: "subnet1234",
										},
										{
											Name: "subnet5678",
										},
									},
									SubnetSelector: &xpv1.Selector{
										MatchLabels: map[string]string{
											"createdBy": "test",
										},
									},
								},
							},
						},
					},
				},
			},
			&svcsdk.NetworkConfiguration{
				AwsvpcConfiguration: &svcsdk.AwsVpcConfiguration{
					Subnets: []*string{
						aws.String("subnet-12345"),
						aws.String("subnet-45678"),
					},
				},
			},
		},
	}

	for name, tc := range cases {

		t.Run(name, func(t *testing.T) {
			got := generateNetworkConfiguration(tc.cr)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("%s\nExample(...): -want, +got:\n%s", tc.reason, diff)
			}
		})
	}
}
