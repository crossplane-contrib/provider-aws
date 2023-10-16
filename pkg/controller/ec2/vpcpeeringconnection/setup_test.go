/*
Copyright 2023 The Crossplane Authors.

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

package vpcpeeringconnection

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	cpresource "github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/ec2/v1alpha1"
)

type args struct {
	kube client.Client
	cr   *v1alpha1.VPCPeeringConnection
}

type vPCPeeringConnectionModifier func(*v1alpha1.VPCPeeringConnection)

func withName(name string) vPCPeeringConnectionModifier {
	return func(r *v1alpha1.VPCPeeringConnection) { r.SetName(name) }
}

func withSpec(p v1alpha1.VPCPeeringConnectionParameters) vPCPeeringConnectionModifier {
	return func(o *v1alpha1.VPCPeeringConnection) { o.Spec.ForProvider = p }
}

func vPCPeeringConnection(m ...vPCPeeringConnectionModifier) *v1alpha1.VPCPeeringConnection {
	cr := &v1alpha1.VPCPeeringConnection{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func TestTagger(t *testing.T) {
	type want struct {
		cr  *v1alpha1.VPCPeeringConnection
		err error
	}

	tag := func(k, v string) *v1alpha1.Tag {
		return &v1alpha1.Tag{Key: ptr.To(k), Value: ptr.To(v)}
	}

	cases := map[string]struct {
		args
		want
	}{
		"ShouldAddTagsIfSpecIsNil": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				cr: vPCPeeringConnection(
					withName("test"),
					withSpec(v1alpha1.VPCPeeringConnectionParameters{}),
				),
			},
			want: want{
				cr: vPCPeeringConnection(
					withName("test"),
					withSpec(v1alpha1.VPCPeeringConnectionParameters{
						TagSpecifications: []*v1alpha1.TagSpecification{
							{
								ResourceType: aws.String("vpc-peering-connection"),
								Tags: []*v1alpha1.Tag{
									tag("Name", "test"),
									tag(cpresource.ExternalResourceTagKeyKind, ""),
									tag(cpresource.ExternalResourceTagKeyName, "test"),
								},
							},
						},
					}),
				),
			},
		},
		"ShouldOverwriteTags": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				cr: vPCPeeringConnection(
					withName("test"),
					withSpec(v1alpha1.VPCPeeringConnectionParameters{
						TagSpecifications: []*v1alpha1.TagSpecification{
							{
								ResourceType: aws.String("vpc-peering-connection"),
								Tags: []*v1alpha1.Tag{
									tag(cpresource.ExternalResourceTagKeyName, "preset"),
								},
							},
						},
					}),
				),
			},
			want: want{
				cr: vPCPeeringConnection(
					withName("test"),
					withSpec(v1alpha1.VPCPeeringConnectionParameters{
						TagSpecifications: []*v1alpha1.TagSpecification{
							{
								ResourceType: aws.String("vpc-peering-connection"),
								Tags: []*v1alpha1.Tag{
									tag("Name", "test"),
									tag(cpresource.ExternalResourceTagKeyKind, ""),
									tag(cpresource.ExternalResourceTagKeyName, "test"),
								},
							},
						},
					}),
				),
			},
		},
		"ShouldMergeTags": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				cr: vPCPeeringConnection(
					withName("test"),
					withSpec(v1alpha1.VPCPeeringConnectionParameters{
						TagSpecifications: []*v1alpha1.TagSpecification{
							{
								ResourceType: aws.String("vpc-peering-connection"),
								Tags: []*v1alpha1.Tag{
									tag("Name", "test"),
									tag(cpresource.ExternalResourceTagKeyKind, ""),
									tag(cpresource.ExternalResourceTagKeyName, "test"),
								},
							},
						},
					}),
				),
			},
			want: want{
				cr: vPCPeeringConnection(
					withName("test"),
					withSpec(v1alpha1.VPCPeeringConnectionParameters{
						TagSpecifications: []*v1alpha1.TagSpecification{
							{
								ResourceType: aws.String("vpc-peering-connection"),
								Tags: []*v1alpha1.Tag{
									tag("Name", "test"),
									tag(cpresource.ExternalResourceTagKeyKind, ""),
									tag(cpresource.ExternalResourceTagKeyName, "test"),
								},
							},
						},
					}),
				),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ta := tagger{kube: tc.args.kube}
			err := ta.Initialize(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
