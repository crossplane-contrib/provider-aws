package service

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/ecs"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/google/go-cmp/cmp"
	"k8s.io/utils/ptr"

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

func TestMapUpdateServiceInput(t *testing.T) {
	cases := map[string]struct {
		reason string
		cr     *svcapitypes.Service
		want   *svcsdk.UpdateServiceInput
	}{
		"TestAllParameters": {
			"When passed all parameters, generate a complete UpdateServiceInput",
			&svcapitypes.Service{
				Spec: svcapitypes.ServiceSpec{
					ForProvider: svcapitypes.ServiceParameters{
						Region: "us-east-1",
						CapacityProviderStrategy: []*svcapitypes.CapacityProviderStrategyItem{
							{
								Base:             ptr.To(int64(1)),
								Weight:           ptr.To(int64(2)),
								CapacityProvider: ptr.To("Test"),
							},
						},
						DeploymentConfiguration: &svcapitypes.DeploymentConfiguration{
							Alarms: &svcapitypes.DeploymentAlarms{
								AlarmNames: []*string{ptr.To("Test 1"), ptr.To("Test 2")},
								Enable:     ptr.To(true),
								Rollback:   ptr.To(false),
							},
							DeploymentCircuitBreaker: &svcapitypes.DeploymentCircuitBreaker{
								Enable:   ptr.To(true),
								Rollback: ptr.To(false),
							},
							MaximumPercent:        ptr.To(int64(3)),
							MinimumHealthyPercent: ptr.To(int64(4)),
						},
						DeploymentController: &svcapitypes.DeploymentController{
							Type: ptr.To("Test 3"),
						},
						DesiredCount:                  ptr.To(int64(5)),
						EnableECSManagedTags:          ptr.To(true),
						EnableExecuteCommand:          ptr.To(false),
						HealthCheckGracePeriodSeconds: ptr.To(int64(6)),
						LaunchType:                    ptr.To("Test 4"),
						PlacementConstraints: []*svcapitypes.PlacementConstraint{
							{
								Expression: ptr.To("Test 5"),
								Type:       ptr.To("Test 6"),
							},
						},
						PlacementStrategy: []*svcapitypes.PlacementStrategy{
							{
								Field: ptr.To("Test 7"),
								Type:  ptr.To("Test 8"),
							},
						},
						PlatformVersion:    ptr.To("Test 9"),
						PropagateTags:      ptr.To("Test 10"),
						Role:               ptr.To("Test 11"),
						SchedulingStrategy: ptr.To("Test 12"),
						ServiceConnectConfiguration: &svcapitypes.ServiceConnectConfiguration{
							Enabled: ptr.To(true),
							LogConfiguration: &svcapitypes.LogConfiguration{
								LogDriver: ptr.To("Test 13"),
								Options:   map[string]*string{"test": ptr.To("Test 14")},
								SecretOptions: []*svcapitypes.Secret{
									{
										Name:      ptr.To("Test 15"),
										ValueFrom: ptr.To("Test 16"),
									},
								},
							},
							Namespace: ptr.To("Test 17"),
							Services: []*svcapitypes.ServiceConnectService{
								{
									ClientAliases: []*svcapitypes.ServiceConnectClientAlias{
										{
											DNSName: ptr.To("Test 17"),
											Port:    ptr.To(int64(8080)),
										},
									},
									DiscoveryName:       ptr.To("Test 18"),
									IngressPortOverride: ptr.To(int64(8081)),
									PortName:            ptr.To("Test 19"),
								},
							},
						},
						ServiceRegistries: []*svcapitypes.ServiceRegistry{
							{
								ContainerName: ptr.To("Test 20"),
								ContainerPort: ptr.To(int64(8082)),
								Port:          ptr.To(int64(8083)),
								RegistryARN:   ptr.To("Test 21"),
							},
						},
						Tags: []*svcapitypes.Tag{
							{
								Key:   ptr.To("Test 22"),
								Value: ptr.To("Test 23"),
							},
						},
						CustomServiceParameters: svcapitypes.CustomServiceParameters{
							ForceDeletion: true,
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
							TaskDefinition: ptr.To("arn:aws:ecs:xx:xx:task-definition/xx:1"),
						},
					},
				},
			},
			&svcsdk.UpdateServiceInput{
				CapacityProviderStrategy: []*svcsdk.CapacityProviderStrategyItem{
					{
						Base:             ptr.To(int64(1)),
						Weight:           ptr.To(int64(2)),
						CapacityProvider: ptr.To("Test"),
					},
				},
				DeploymentConfiguration: &svcsdk.DeploymentConfiguration{
					Alarms: &svcsdk.DeploymentAlarms{
						AlarmNames: []*string{ptr.To("Test 1"), ptr.To("Test 2")},
						Enable:     ptr.To(true),
						Rollback:   ptr.To(false),
					},
					DeploymentCircuitBreaker: &svcsdk.DeploymentCircuitBreaker{
						Enable:   ptr.To(true),
						Rollback: ptr.To(false),
					},
					MaximumPercent:        ptr.To(int64(3)),
					MinimumHealthyPercent: ptr.To(int64(4)),
				},
				DesiredCount:                  ptr.To(int64(5)),
				EnableECSManagedTags:          ptr.To(true),
				EnableExecuteCommand:          ptr.To(false),
				HealthCheckGracePeriodSeconds: ptr.To(int64(6)),
				PlacementConstraints: []*svcsdk.PlacementConstraint{
					{
						Expression: ptr.To("Test 5"),
						Type:       ptr.To("Test 6"),
					},
				},
				PlacementStrategy: []*svcsdk.PlacementStrategy{
					{
						Field: ptr.To("Test 7"),
						Type:  ptr.To("Test 8"),
					},
				},
				PlatformVersion: ptr.To("Test 9"),
				PropagateTags:   ptr.To("Test 10"),
				ServiceConnectConfiguration: &svcsdk.ServiceConnectConfiguration{
					Enabled: ptr.To(true),
					LogConfiguration: &svcsdk.LogConfiguration{
						LogDriver: ptr.To("Test 13"),
						Options:   map[string]*string{"test": ptr.To("Test 14")},
						SecretOptions: []*svcsdk.Secret{
							{
								Name:      ptr.To("Test 15"),
								ValueFrom: ptr.To("Test 16"),
							},
						},
					},
					Namespace: ptr.To("Test 17"),
					Services: []*svcsdk.ServiceConnectService{
						{
							ClientAliases: []*svcsdk.ServiceConnectClientAlias{
								{
									DnsName: ptr.To("Test 17"),
									Port:    ptr.To(int64(8080)),
								},
							},
							DiscoveryName:       ptr.To("Test 18"),
							IngressPortOverride: ptr.To(int64(8081)),
							PortName:            ptr.To("Test 19"),
						},
					},
				},
				ServiceRegistries: []*svcsdk.ServiceRegistry{
					{
						ContainerName: ptr.To("Test 20"),
						ContainerPort: ptr.To(int64(8082)),
						Port:          ptr.To(int64(8083)),
						RegistryArn:   ptr.To("Test 21"),
					},
				},
				LoadBalancers: []*svcsdk.LoadBalancer{
					{
						ContainerName:    aws.String("test-container"),
						ContainerPort:    aws.Int64(443),
						LoadBalancerName: aws.String("test-loadbalancer"),
						TargetGroupArn:   aws.String("arn:::test-listener"),
					},
				},
				NetworkConfiguration: &svcsdk.NetworkConfiguration{
					AwsvpcConfiguration: &svcsdk.AwsVpcConfiguration{
						Subnets: []*string{
							aws.String("subnet-12345"),
							aws.String("subnet-45678"),
						},
					},
				},
				TaskDefinition: ptr.To("arn:aws:ecs:xx:xx:task-definition/xx:1"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := &svcsdk.UpdateServiceInput{}
			mapUpdateServiceInput(tc.cr, got)

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("%s\nExample(...): -want, +got:\n%s", tc.reason, diff)
			}
		})
	}
}
