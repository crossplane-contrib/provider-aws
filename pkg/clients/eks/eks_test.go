/*
Copyright 2020 The Crossplane Authors.

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

package eks

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/eks"
	ekstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/aws/smithy-go/document"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane-contrib/provider-aws/apis/eks/v1beta1"
)

var (
	clusterName = "my-cool-cluster"
	keyArn      = "mykey:arn"
	roleArn     = "myrole:arn"
	falseVal    = false
	trueVal     = true
	version     = "1.16"
)

func TestIsErrorNotFound(t *testing.T) {
	cases := map[string]struct {
		err  error
		want bool
	}{
		"IsErrorNotFound": {
			err:  &ekstypes.ResourceNotFoundException{},
			want: true,
		},
		"NotErrorNotFound": {
			err:  &ekstypes.InvalidRequestException{},
			want: false,
		},
		"Nil": {
			err:  nil,
			want: false,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsErrorNotFound(tc.err)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsErrorInUse(t *testing.T) {
	cases := map[string]struct {
		err  error
		want bool
	}{
		"IsErrorInUse": {
			err:  &ekstypes.ResourceInUseException{},
			want: true,
		},
		"NotErrorInUse": {
			err:  &ekstypes.ResourceNotFoundException{},
			want: false,
		},
		"Nil": {
			err:  nil,
			want: false,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsErrorInUse(tc.err)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsErrorInvalidRequest(t *testing.T) {
	cases := map[string]struct {
		err  error
		want bool
	}{
		"IsErrorInvalidRequest": {
			err:  &ekstypes.InvalidRequestException{},
			want: true,
		},
		"NotErrorInvalidRequest": {
			err:  &ekstypes.ResourceNotFoundException{},
			want: false,
		},
		"Nil": {
			err:  nil,
			want: false,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsErrorInvalidRequest(tc.err)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateCreateClusterInput(t *testing.T) {
	type args struct {
		name string
		p    *v1beta1.ClusterParameters
	}

	cases := map[string]struct {
		args args
		want *eks.CreateClusterInput
	}{
		"AllFields": {
			args: args{
				name: clusterName,
				p: &v1beta1.ClusterParameters{
					EncryptionConfig: []v1beta1.EncryptionConfig{
						{
							Provider: v1beta1.Provider{
								KeyArn: keyArn,
							},
							Resources: []string{"secrets"},
						},
					},
					Logging: &v1beta1.Logging{
						ClusterLogging: []v1beta1.LogSetup{
							{
								Enabled: &falseVal,
								Types: []v1beta1.LogType{
									v1beta1.LogTypeAPI,
								},
							},
						},
					},
					ResourcesVpcConfig: v1beta1.VpcConfigRequest{
						EndpointPrivateAccess: &trueVal,
						EndpointPublicAccess:  &trueVal,
						PublicAccessCidrs:     []string{"0.0.0.0/0"},
						SecurityGroupIDs:      []string{"cool-sg-1"},
						SubnetIDs:             []string{"cool-subnet"},
					},
					RoleArn: roleArn,
					Tags:    map[string]string{"key": "val"},
					Version: &version,
				},
			},
			want: &eks.CreateClusterInput{
				EncryptionConfig: []ekstypes.EncryptionConfig{
					{
						Provider: &ekstypes.Provider{
							KeyArn: &keyArn,
						},
						Resources: []string{"secrets"},
					},
				},
				Logging: &ekstypes.Logging{
					ClusterLogging: []ekstypes.LogSetup{
						{
							Enabled: &falseVal,
							Types: []ekstypes.LogType{
								ekstypes.LogTypeApi,
							},
						},
					},
				},
				Name: &clusterName,
				ResourcesVpcConfig: &ekstypes.VpcConfigRequest{
					EndpointPrivateAccess: &trueVal,
					EndpointPublicAccess:  &trueVal,
					PublicAccessCidrs:     []string{"0.0.0.0/0"},
					SecurityGroupIds:      []string{"cool-sg-1"},
					SubnetIds:             []string{"cool-subnet"},
				},
				RoleArn: &roleArn,
				Tags:    map[string]string{"key": "val"},
				Version: &version,
			},
		},
		"SomeFields": {
			args: args{
				name: clusterName,
				p: &v1beta1.ClusterParameters{
					ResourcesVpcConfig: v1beta1.VpcConfigRequest{
						EndpointPrivateAccess: &trueVal,
						EndpointPublicAccess:  &trueVal,
						PublicAccessCidrs:     []string{"0.0.0.0/0"},
						SecurityGroupIDs:      []string{"cool-sg-1"},
						SubnetIDs:             []string{"cool-subnet"},
					},
					RoleArn: roleArn,
					Version: &version,
				},
			},
			want: &eks.CreateClusterInput{
				Name: &clusterName,
				ResourcesVpcConfig: &ekstypes.VpcConfigRequest{
					EndpointPrivateAccess: &trueVal,
					EndpointPublicAccess:  &trueVal,
					PublicAccessCidrs:     []string{"0.0.0.0/0"},
					SecurityGroupIds:      []string{"cool-sg-1"},
					SubnetIds:             []string{"cool-subnet"},
				},
				RoleArn: &roleArn,
				Version: &version,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateCreateClusterInput(tc.args.name, tc.args.p)
			if diff := cmp.Diff(tc.want, got, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateUpdateClusterConfigInputForLogging(t *testing.T) {
	type args struct {
		name string
		p    *v1beta1.ClusterParameters
	}

	cases := map[string]struct {
		args args
		want *eks.UpdateClusterConfigInput
	}{
		"AllFields": {
			args: args{
				name: clusterName,
				p: &v1beta1.ClusterParameters{
					EncryptionConfig: []v1beta1.EncryptionConfig{
						{
							Provider: v1beta1.Provider{
								KeyArn: keyArn,
							},
							Resources: []string{"secrets"},
						},
					},
					Logging: &v1beta1.Logging{
						ClusterLogging: []v1beta1.LogSetup{
							{
								Enabled: &trueVal,
								Types: []v1beta1.LogType{
									v1beta1.LogTypeAPI,
									v1beta1.LogTypeAudit,
									v1beta1.LogTypeAuthenticator,
									v1beta1.LogTypeControllerManager,
									v1beta1.LogTypeScheduler,
								},
							},
						},
					},
					ResourcesVpcConfig: v1beta1.VpcConfigRequest{
						EndpointPrivateAccess: &trueVal,
						EndpointPublicAccess:  &trueVal,
						PublicAccessCidrs:     []string{"0.0.0.0/0"},
					},
					RoleArn: roleArn,
					Tags:    map[string]string{"key": "val"},
					Version: &version,
				},
			},
			want: &eks.UpdateClusterConfigInput{
				Logging: &ekstypes.Logging{
					ClusterLogging: []ekstypes.LogSetup{
						{
							Enabled: &trueVal,
							Types: []ekstypes.LogType{
								ekstypes.LogTypeApi,
								ekstypes.LogTypeAudit,
								ekstypes.LogTypeAuthenticator,
								ekstypes.LogTypeControllerManager,
								ekstypes.LogTypeScheduler,
							},
						},
					},
				},
				Name: &clusterName,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateUpdateClusterConfigInputForLogging(tc.args.name, tc.args.p)
			if diff := cmp.Diff(tc.want, got, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateUpdateClusterConfigInputForVPC(t *testing.T) {
	type args struct {
		name string
		p    *v1beta1.ClusterParameters
	}

	cases := map[string]struct {
		args args
		want *eks.UpdateClusterConfigInput
	}{
		"AllFields": {
			args: args{
				name: clusterName,
				p: &v1beta1.ClusterParameters{
					EncryptionConfig: []v1beta1.EncryptionConfig{
						{
							Provider: v1beta1.Provider{
								KeyArn: keyArn,
							},
							Resources: []string{"secrets"},
						},
					},
					Logging: &v1beta1.Logging{
						ClusterLogging: []v1beta1.LogSetup{
							{
								Enabled: &falseVal,
								Types: []v1beta1.LogType{
									v1beta1.LogTypeAPI,
								},
							},
						},
					},
					ResourcesVpcConfig: v1beta1.VpcConfigRequest{
						EndpointPrivateAccess: &trueVal,
						EndpointPublicAccess:  &trueVal,
						PublicAccessCidrs:     []string{"0.0.0.0/0"},
					},
					RoleArn: roleArn,
					Tags:    map[string]string{"key": "val"},
					Version: &version,
				},
			},
			want: &eks.UpdateClusterConfigInput{
				Name: &clusterName,
				ResourcesVpcConfig: &ekstypes.VpcConfigRequest{
					EndpointPrivateAccess: &trueVal,
					EndpointPublicAccess:  &trueVal,
					PublicAccessCidrs:     []string{"0.0.0.0/0"},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateUpdateClusterConfigInputForVPC(tc.args.name, tc.args.p)
			if diff := cmp.Diff(tc.want, got, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateObservation(t *testing.T) {
	createTime := time.Now()
	clusterArn := "my:arn"
	endpoint := "https://my-endpoint.com"
	caData := "Y2VydGlmaWNhdGUtYXV0aG9yaXR5LWRhdGE="
	oidcIssuer := "secret-issuer"
	version := "1.21"
	platformVersion := "eks1.0"
	securityGrp := "sg-1234"
	vpc := "vpc-1234"

	cases := map[string]struct {
		cluster *ekstypes.Cluster
		want    v1beta1.ClusterObservation
	}{
		"AllFields": {
			cluster: &ekstypes.Cluster{
				Arn:       &clusterArn,
				CreatedAt: &createTime,
				Endpoint:  &endpoint,
				CertificateAuthority: &ekstypes.Certificate{
					Data: &caData,
				},
				Identity: &ekstypes.Identity{
					Oidc: &ekstypes.OIDC{
						Issuer: &oidcIssuer,
					},
				},
				Version:         &version,
				PlatformVersion: &platformVersion,
				ResourcesVpcConfig: &ekstypes.VpcConfigResponse{
					ClusterSecurityGroupId: &securityGrp,
					VpcId:                  &vpc,
				},
				Status: ekstypes.ClusterStatusActive,
			},
			want: v1beta1.ClusterObservation{
				Arn:                      clusterArn,
				CreatedAt:                &metav1.Time{Time: createTime},
				Endpoint:                 endpoint,
				CertificateAuthorityData: caData,
				Identity: v1beta1.Identity{
					OIDC: v1beta1.OIDC{
						Issuer: oidcIssuer,
					},
				},
				Version:         version,
				PlatformVersion: platformVersion,
				ResourcesVpcConfig: v1beta1.VpcConfigResponse{
					ClusterSecurityGroupID: securityGrp,
					VpcID:                  vpc,
				},
				Status: v1beta1.ClusterStatusActive,
			},
		},
		"SomeFields": {
			cluster: &ekstypes.Cluster{
				Arn:             &clusterArn,
				CreatedAt:       &createTime,
				Version:         &version,
				PlatformVersion: &platformVersion,
				ResourcesVpcConfig: &ekstypes.VpcConfigResponse{
					ClusterSecurityGroupId: &securityGrp,
					VpcId:                  &vpc,
				},
				Status: ekstypes.ClusterStatusActive,
			},
			want: v1beta1.ClusterObservation{
				Arn:             clusterArn,
				CreatedAt:       &metav1.Time{Time: createTime},
				Version:         version,
				PlatformVersion: platformVersion,
				ResourcesVpcConfig: v1beta1.VpcConfigResponse{
					ClusterSecurityGroupID: securityGrp,
					VpcID:                  vpc,
				},
				Status: v1beta1.ClusterStatusActive,
			},
		},
		"NilCluster": {
			cluster: nil,
			want:    v1beta1.ClusterObservation{},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateObservation(tc.cluster)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitialize(t *testing.T) {
	ServiceIpv4Cidr := "172.20.0.0/16"
	Ipv4Family := "ipv4"
	cases := map[string]struct {
		parameters *v1beta1.ClusterParameters
		cluster    *ekstypes.Cluster
		want       *v1beta1.ClusterParameters
	}{
		"AllOptionalFields": {
			parameters: &v1beta1.ClusterParameters{
				ResourcesVpcConfig: v1beta1.VpcConfigRequest{
					SecurityGroupIDs: []string{"cool-sg-1"},
					SubnetIDs:        []string{"cool-subnet"},
				},
				RoleArn: roleArn,
				Tags:    map[string]string{"key": "val"},
				Version: &version,
			},
			cluster: &ekstypes.Cluster{
				EncryptionConfig: []ekstypes.EncryptionConfig{
					{
						Provider: &ekstypes.Provider{
							KeyArn: &keyArn,
						},
						Resources: []string{"secrets"},
					},
				},
				Logging: &ekstypes.Logging{
					ClusterLogging: []ekstypes.LogSetup{
						{
							Enabled: &falseVal,
							Types: []ekstypes.LogType{
								ekstypes.LogTypeApi,
							},
						},
					},
				},
				Name: &clusterName,
				ResourcesVpcConfig: &ekstypes.VpcConfigResponse{
					EndpointPrivateAccess: trueVal,
					EndpointPublicAccess:  trueVal,
					PublicAccessCidrs:     []string{"0.0.0.0/0"},
					SecurityGroupIds:      []string{"cool-sg-1"},
					SubnetIds:             []string{"cool-subnet"},
				},
				KubernetesNetworkConfig: &ekstypes.KubernetesNetworkConfigResponse{
					IpFamily:        ekstypes.IpFamily(Ipv4Family),
					ServiceIpv4Cidr: &ServiceIpv4Cidr,
				},
				RoleArn: &roleArn,
				Tags:    map[string]string{"key": "val"},
				Version: &version,
			},
			want: &v1beta1.ClusterParameters{
				EncryptionConfig: []v1beta1.EncryptionConfig{
					{
						Provider: v1beta1.Provider{
							KeyArn: keyArn,
						},
						Resources: []string{"secrets"},
					},
				},
				Logging: &v1beta1.Logging{
					ClusterLogging: []v1beta1.LogSetup{
						{
							Enabled: &falseVal,
							Types: []v1beta1.LogType{
								v1beta1.LogTypeAPI,
							},
						},
					},
				},
				ResourcesVpcConfig: v1beta1.VpcConfigRequest{
					EndpointPrivateAccess: &trueVal,
					EndpointPublicAccess:  &trueVal,
					PublicAccessCidrs:     []string{"0.0.0.0/0"},
					SecurityGroupIDs:      []string{"cool-sg-1"},
					SubnetIDs:             []string{"cool-subnet"},
				},
				KubernetesNetworkConfig: &v1beta1.KubernetesNetworkConfigRequest{
					ServiceIpv4Cidr: ServiceIpv4Cidr,
					IPFamily:        v1beta1.IPFamily(Ipv4Family),
				},
				RoleArn: roleArn,
				Tags:    map[string]string{"key": "val"},
				Version: &version,
			},
		},
		"SomeFieldsDontOverwrite": {
			parameters: &v1beta1.ClusterParameters{
				ResourcesVpcConfig: v1beta1.VpcConfigRequest{
					SecurityGroupIDs: []string{"cool-sg-1"},
					SubnetIDs:        []string{"cool-subnet"},
				},
				RoleArn: roleArn,
				Tags:    map[string]string{"key": "val"},
				Version: &version,
			},
			cluster: &ekstypes.Cluster{
				EncryptionConfig: []ekstypes.EncryptionConfig{
					{
						Provider: &ekstypes.Provider{
							KeyArn: &keyArn,
						},
						Resources: []string{"secrets"},
					},
				},
				RoleArn: &roleArn,
				Tags:    map[string]string{"dont": "overwrite"},
				Version: &version,
			},
			want: &v1beta1.ClusterParameters{
				EncryptionConfig: []v1beta1.EncryptionConfig{
					{
						Provider: v1beta1.Provider{
							KeyArn: keyArn,
						},
						Resources: []string{"secrets"},
					},
				},
				ResourcesVpcConfig: v1beta1.VpcConfigRequest{
					SecurityGroupIDs: []string{"cool-sg-1"},
					SubnetIDs:        []string{"cool-subnet"},
				},
				RoleArn: roleArn,
				Tags:    map[string]string{"key": "val"},
				Version: &version,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitialize(tc.parameters, tc.cluster)
			if diff := cmp.Diff(tc.want, tc.parameters); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsUpToDate(t *testing.T) {
	otherVersion := "1.15"

	type args struct {
		cluster *ekstypes.Cluster
		p       *v1beta1.ClusterParameters
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFields": {
			args: args{
				p: &v1beta1.ClusterParameters{
					EncryptionConfig: []v1beta1.EncryptionConfig{
						{
							Provider: v1beta1.Provider{
								KeyArn: keyArn,
							},
							Resources: []string{"secrets"},
						},
					},
					Logging: &v1beta1.Logging{
						ClusterLogging: []v1beta1.LogSetup{
							{
								Enabled: &falseVal,
								Types: []v1beta1.LogType{
									v1beta1.LogTypeAPI,
								},
							},
						},
					},
					ResourcesVpcConfig: v1beta1.VpcConfigRequest{
						EndpointPrivateAccess: &trueVal,
						EndpointPublicAccess:  &trueVal,
						PublicAccessCidrs:     []string{"0.0.0.0/0"},
						SecurityGroupIDs:      []string{"cool-sg-1"},
						SubnetIDs:             []string{"cool-subnet"},
					},
					RoleArn: roleArn,
					Tags:    map[string]string{"key": "val"},
					Version: &version,
				},
				cluster: &ekstypes.Cluster{
					EncryptionConfig: []ekstypes.EncryptionConfig{
						{
							Provider: &ekstypes.Provider{
								KeyArn: &keyArn,
							},
							Resources: []string{"secrets"},
						},
					},
					Logging: &ekstypes.Logging{
						ClusterLogging: []ekstypes.LogSetup{
							{
								Enabled: &falseVal,
								Types: []ekstypes.LogType{
									ekstypes.LogTypeApi,
								},
							},
						},
					},
					Name: &clusterName,
					ResourcesVpcConfig: &ekstypes.VpcConfigResponse{
						EndpointPrivateAccess: trueVal,
						EndpointPublicAccess:  trueVal,
						PublicAccessCidrs:     []string{"0.0.0.0/0"},
						SecurityGroupIds:      []string{"cool-sg-1"},
						SubnetIds:             []string{"cool-subnet"},
					},
					RoleArn: &roleArn,
					Tags:    map[string]string{"key": "val"},
					Version: &version,
				},
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				p: &v1beta1.ClusterParameters{
					ResourcesVpcConfig: v1beta1.VpcConfigRequest{
						EndpointPrivateAccess: &trueVal,
						EndpointPublicAccess:  &trueVal,
						PublicAccessCidrs:     []string{"0.0.0.0/0"},
						SecurityGroupIDs:      []string{"cool-sg-1"},
						SubnetIDs:             []string{"cool-subnet"},
					},
					RoleArn: roleArn,
					Tags:    map[string]string{"key": "val"},
					Version: &version,
				},
				cluster: &ekstypes.Cluster{
					Name: &clusterName,
					ResourcesVpcConfig: &ekstypes.VpcConfigResponse{
						EndpointPrivateAccess: trueVal,
						EndpointPublicAccess:  trueVal,
						PublicAccessCidrs:     []string{"0.0.0.0/0"},
						SecurityGroupIds:      []string{"cool-sg-1"},
						SubnetIds:             []string{"cool-subnet"},
					},
					RoleArn: &roleArn,
					Tags:    map[string]string{"key": "val"},
					Version: &otherVersion,
				},
			},
			want: false,
		},
		"IgnoreRefs": {
			args: args{
				p: &v1beta1.ClusterParameters{
					ResourcesVpcConfig: v1beta1.VpcConfigRequest{
						EndpointPrivateAccess: &trueVal,
						EndpointPublicAccess:  &trueVal,
						PublicAccessCidrs:     []string{"0.0.0.0/0"},
						SecurityGroupIDs:      []string{"cool-sg-1"},
						SecurityGroupIDRefs: []xpv1.Reference{
							{
								Name: "cool-ref",
							},
						},
						SecurityGroupIDSelector: &xpv1.Selector{
							MatchLabels: map[string]string{"key": "val"},
						},
						SubnetIDs: []string{"cool-subnet"},
						SubnetIDRefs: []xpv1.Reference{
							{
								Name: "cool-ref",
							},
						},
						SubnetIDSelector: &xpv1.Selector{
							MatchLabels: map[string]string{"key": "val"},
						},
					},
					RoleArn: roleArn,
					RoleArnRef: &xpv1.Reference{
						Name: "fun-ref",
					},
					RoleArnSelector: &xpv1.Selector{
						MatchLabels: map[string]string{"key": "val"},
					},
					Tags:    map[string]string{"key": "val"},
					Version: &version,
				},
				cluster: &ekstypes.Cluster{
					Name: &clusterName,
					ResourcesVpcConfig: &ekstypes.VpcConfigResponse{
						EndpointPrivateAccess: trueVal,
						EndpointPublicAccess:  trueVal,
						PublicAccessCidrs:     []string{"0.0.0.0/0"},
						SecurityGroupIds:      []string{"cool-sg-1"},
						SubnetIds:             []string{"cool-subnet"},
					},
					RoleArn: &roleArn,
					Tags:    map[string]string{"key": "val"},
					Version: &version,
				},
			},
			want: true,
		},
		"EquivalentCIDRs": {
			args: args{
				p: &v1beta1.ClusterParameters{
					ResourcesVpcConfig: v1beta1.VpcConfigRequest{
						PublicAccessCidrs: []string{"0.0.0.10/24"},
					},
				},
				cluster: &ekstypes.Cluster{
					ResourcesVpcConfig: &ekstypes.VpcConfigResponse{
						PublicAccessCidrs: []string{"0.0.0.0/24"},
					},
				},
			},
			want: true,
		},
		"MultipleEquivalentCIDRs": {
			args: args{
				p: &v1beta1.ClusterParameters{
					ResourcesVpcConfig: v1beta1.VpcConfigRequest{
						PublicAccessCidrs: []string{"0.0.0.10/24", "0.0.0.255/24", "1.1.1.1/8"},
					},
				},
				cluster: &ekstypes.Cluster{
					ResourcesVpcConfig: &ekstypes.VpcConfigResponse{
						PublicAccessCidrs: []string{"0.0.0.0/24", "1.0.0.0/8"},
					},
				},
			},
			want: true,
		},
		"NotEquivalentCIDRs": {
			args: args{
				p: &v1beta1.ClusterParameters{
					ResourcesVpcConfig: v1beta1.VpcConfigRequest{
						PublicAccessCidrs: []string{"0.0.0.10/24", "0.0.0.255/24", "1.1.1.1/16"},
					},
				},
				cluster: &ekstypes.Cluster{
					ResourcesVpcConfig: &ekstypes.VpcConfigResponse{
						PublicAccessCidrs: []string{"0.0.0.0/24", "1.0.0.0/8"},
					},
				},
			},
			want: false,
		},
		"DifferentTags": {
			args: args{
				p: &v1beta1.ClusterParameters{
					ResourcesVpcConfig: v1beta1.VpcConfigRequest{
						EndpointPrivateAccess: &trueVal,
						EndpointPublicAccess:  &trueVal,
						PublicAccessCidrs:     []string{"0.0.0.0/0"},
						SecurityGroupIDs:      []string{"cool-sg-1"},
						SecurityGroupIDRefs: []xpv1.Reference{
							{
								Name: "cool-ref",
							},
						},
						SecurityGroupIDSelector: &xpv1.Selector{
							MatchLabels: map[string]string{"key": "val"},
						},
						SubnetIDs: []string{"cool-subnet"},
						SubnetIDRefs: []xpv1.Reference{
							{
								Name: "cool-ref",
							},
						},
						SubnetIDSelector: &xpv1.Selector{
							MatchLabels: map[string]string{"key": "val"},
						},
					},
					RoleArn: roleArn,
					RoleArnRef: &xpv1.Reference{
						Name: "fun-ref",
					},
					RoleArnSelector: &xpv1.Selector{
						MatchLabels: map[string]string{"key": "val"},
					},
					Tags:    map[string]string{"key": "val", "another": "tag"},
					Version: &version,
				},
				cluster: &ekstypes.Cluster{
					Name: &clusterName,
					ResourcesVpcConfig: &ekstypes.VpcConfigResponse{
						EndpointPrivateAccess: trueVal,
						EndpointPublicAccess:  trueVal,
						PublicAccessCidrs:     []string{"0.0.0.0/0"},
						SecurityGroupIds:      []string{"cool-sg-1"},
						SubnetIds:             []string{"cool-subnet"},
					},
					RoleArn: &roleArn,
					Tags:    map[string]string{"key": "val"},
					Version: &version,
				},
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, _ := IsUpToDate(tc.args.p, tc.args.cluster)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
