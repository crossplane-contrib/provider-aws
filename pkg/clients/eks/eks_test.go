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
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"

	"github.com/crossplane/provider-aws/apis/eks/v1beta1"
)

func TestIsErrorNotFound(t *testing.T) {
	cases := map[string]struct {
		err  error
		want bool
	}{
		"IsErrorNotFound": {
			err:  errors.New(eks.ErrCodeResourceNotFoundException),
			want: true,
		},
		"NotErrorNotFound": {
			err:  errors.New(eks.ErrCodeInvalidRequestException),
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
			err:  errors.New(eks.ErrCodeResourceInUseException),
			want: true,
		},
		"NotErrorInUse": {
			err:  errors.New(eks.ErrCodeNotFoundException),
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

func TestGenerateCreateClusterInput(t *testing.T) {
	clusterName := "my-cool-cluster"
	keyArn := "mykey:arn"
	roleArn := "myrole:arn"
	falseVal := false
	trueVal := true
	version := "1.16"

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
					RoleArn: &roleArn,
					Tags:    map[string]string{"key": "val"},
					Version: &version,
				},
			},
			want: &eks.CreateClusterInput{
				EncryptionConfig: []eks.EncryptionConfig{
					{
						Provider: &eks.Provider{
							KeyArn: &keyArn,
						},
						Resources: []string{"secrets"},
					},
				},
				Logging: &eks.Logging{
					ClusterLogging: []eks.LogSetup{
						{
							Enabled: &falseVal,
							Types: []eks.LogType{
								eks.LogTypeApi,
							},
						},
					},
				},
				Name: &clusterName,
				ResourcesVpcConfig: &eks.VpcConfigRequest{
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
					RoleArn: &roleArn,
					Version: &version,
				},
			},
			want: &eks.CreateClusterInput{
				Name: &clusterName,
				ResourcesVpcConfig: &eks.VpcConfigRequest{
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
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateUpdateClusterInput(t *testing.T) {
	clusterName := "my-cool-cluster"
	keyArn := "mykey:arn"
	roleArn := "myrole:arn"
	falseVal := false
	trueVal := true
	version := "1.16"

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
						SecurityGroupIDs:      []string{"cool-sg-1"},
						SubnetIDs:             []string{"cool-subnet"},
					},
					RoleArn: &roleArn,
					Tags:    map[string]string{"key": "val"},
					Version: &version,
				},
			},
			want: &eks.UpdateClusterConfigInput{
				Logging: &eks.Logging{
					ClusterLogging: []eks.LogSetup{
						{
							Enabled: &falseVal,
							Types: []eks.LogType{
								eks.LogTypeApi,
							},
						},
					},
				},
				Name: &clusterName,
				ResourcesVpcConfig: &eks.VpcConfigRequest{
					EndpointPrivateAccess: &trueVal,
					EndpointPublicAccess:  &trueVal,
					PublicAccessCidrs:     []string{"0.0.0.0/0"},
					SecurityGroupIds:      []string{"cool-sg-1"},
					SubnetIds:             []string{"cool-subnet"},
				},
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
					RoleArn: &roleArn,
					Version: &version,
				},
			},
			want: &eks.UpdateClusterConfigInput{
				Name: &clusterName,
				ResourcesVpcConfig: &eks.VpcConfigRequest{
					EndpointPrivateAccess: &trueVal,
					EndpointPublicAccess:  &trueVal,
					PublicAccessCidrs:     []string{"0.0.0.0/0"},
					SecurityGroupIds:      []string{"cool-sg-1"},
					SubnetIds:             []string{"cool-subnet"},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateUpdateClusterConfigInput(tc.args.name, tc.args.p)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateObservation(t *testing.T) {
	createTime := time.Now()
	clusterArn := "my:arn"
	endpoint := "https://my-endpoint.com"
	oidcIssuer := "secret-issuer"
	platformVersion := "eks1.0"
	securityGrp := "sg-1234"
	vpc := "vpc-1234"

	cases := map[string]struct {
		cluster *eks.Cluster
		want    v1beta1.ClusterObservation
	}{
		"AllFields": {
			cluster: &eks.Cluster{
				Arn:       &clusterArn,
				CreatedAt: &createTime,
				Endpoint:  &endpoint,
				Identity: &eks.Identity{
					Oidc: &eks.OIDC{
						Issuer: &oidcIssuer,
					},
				},
				PlatformVersion: &platformVersion,
				ResourcesVpcConfig: &eks.VpcConfigResponse{
					ClusterSecurityGroupId: &securityGrp,
					VpcId:                  &vpc,
				},
				Status: eks.ClusterStatusActive,
			},
			want: v1beta1.ClusterObservation{
				Arn:       clusterArn,
				CreatedAt: &metav1.Time{Time: createTime},
				Endpoint:  endpoint,
				Identity: v1beta1.Identity{
					OIDC: v1beta1.OIDC{
						Issuer: oidcIssuer,
					},
				},
				PlatformVersion: platformVersion,
				ResourcesVpcConfig: v1beta1.VpcConfigResponse{
					ClusterSecurityGroupID: securityGrp,
					VpcID:                  vpc,
				},
				Status: v1beta1.ClusterStatusActive,
			},
		},
		"SomeFields": {
			cluster: &eks.Cluster{
				Arn:             &clusterArn,
				CreatedAt:       &createTime,
				PlatformVersion: &platformVersion,
				ResourcesVpcConfig: &eks.VpcConfigResponse{
					ClusterSecurityGroupId: &securityGrp,
					VpcId:                  &vpc,
				},
				Status: eks.ClusterStatusActive,
			},
			want: v1beta1.ClusterObservation{
				Arn:             clusterArn,
				CreatedAt:       &metav1.Time{Time: createTime},
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
	clusterName := "my-cool-cluster"
	keyArn := "mykey:arn"
	roleArn := "myrole:arn"
	falseVal := false
	trueVal := true
	version := "1.16"

	cases := map[string]struct {
		parameters *v1beta1.ClusterParameters
		cluster    *eks.Cluster
		want       *v1beta1.ClusterParameters
	}{
		"AllOptionalFields": {
			parameters: &v1beta1.ClusterParameters{
				ResourcesVpcConfig: v1beta1.VpcConfigRequest{
					SecurityGroupIDs: []string{"cool-sg-1"},
					SubnetIDs:        []string{"cool-subnet"},
				},
				RoleArn: &roleArn,
				Tags:    map[string]string{"key": "val"},
				Version: &version,
			},
			cluster: &eks.Cluster{
				EncryptionConfig: []eks.EncryptionConfig{
					{
						Provider: &eks.Provider{
							KeyArn: &keyArn,
						},
						Resources: []string{"secrets"},
					},
				},
				Logging: &eks.Logging{
					ClusterLogging: []eks.LogSetup{
						{
							Enabled: &falseVal,
							Types: []eks.LogType{
								eks.LogTypeApi,
							},
						},
					},
				},
				Name: &clusterName,
				ResourcesVpcConfig: &eks.VpcConfigResponse{
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
				RoleArn: &roleArn,
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
				RoleArn: &roleArn,
				Tags:    map[string]string{"key": "val"},
				Version: &version,
			},
			cluster: &eks.Cluster{
				EncryptionConfig: []eks.EncryptionConfig{
					{
						Provider: &eks.Provider{
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
				RoleArn: &roleArn,
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
	clusterName := "my-cool-cluster"
	keyArn := "mykey:arn"
	roleArn := "myrole:arn"
	falseVal := false
	trueVal := true
	version := "1.16"
	otherVersion := "1.15"

	type args struct {
		cluster *eks.Cluster
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
					RoleArn: &roleArn,
					Tags:    map[string]string{"key": "val"},
					Version: &version,
				},
				cluster: &eks.Cluster{
					EncryptionConfig: []eks.EncryptionConfig{
						{
							Provider: &eks.Provider{
								KeyArn: &keyArn,
							},
							Resources: []string{"secrets"},
						},
					},
					Logging: &eks.Logging{
						ClusterLogging: []eks.LogSetup{
							{
								Enabled: &falseVal,
								Types: []eks.LogType{
									eks.LogTypeApi,
								},
							},
						},
					},
					Name: &clusterName,
					ResourcesVpcConfig: &eks.VpcConfigResponse{
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
					RoleArn: &roleArn,
					Tags:    map[string]string{"key": "val"},
					Version: &version,
				},
				cluster: &eks.Cluster{
					Name: &clusterName,
					ResourcesVpcConfig: &eks.VpcConfigResponse{
						EndpointPrivateAccess: &trueVal,
						EndpointPublicAccess:  &trueVal,
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
						SecurityGroupIDRefs: []v1alpha1.Reference{
							{
								Name: "cool-ref",
							},
						},
						SecurityGroupIDSelector: &v1alpha1.Selector{
							MatchLabels: map[string]string{"key": "val"},
						},
						SubnetIDs: []string{"cool-subnet"},
						SubnetIDRefs: []v1alpha1.Reference{
							{
								Name: "cool-ref",
							},
						},
						SubnetIDSelector: &v1alpha1.Selector{
							MatchLabels: map[string]string{"key": "val"},
						},
					},
					RoleArn: &roleArn,
					RoleArnRef: &v1alpha1.Reference{
						Name: "fun-ref",
					},
					RoleArnSelector: &v1alpha1.Selector{
						MatchLabels: map[string]string{"key": "val"},
					},
					Tags:    map[string]string{"key": "val"},
					Version: &version,
				},
				cluster: &eks.Cluster{
					Name: &clusterName,
					ResourcesVpcConfig: &eks.VpcConfigResponse{
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
			want: true,
		},
		"EquivalentCIDRs": {
			args: args{
				p: &v1beta1.ClusterParameters{
					ResourcesVpcConfig: v1beta1.VpcConfigRequest{
						PublicAccessCidrs: []string{"0.0.0.10/24"},
					},
				},
				cluster: &eks.Cluster{
					ResourcesVpcConfig: &eks.VpcConfigResponse{
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
				cluster: &eks.Cluster{
					ResourcesVpcConfig: &eks.VpcConfigResponse{
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
				cluster: &eks.Cluster{
					ResourcesVpcConfig: &eks.VpcConfigResponse{
						PublicAccessCidrs: []string{"0.0.0.0/24", "1.0.0.0/8"},
					},
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
