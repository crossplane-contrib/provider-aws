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

package elb

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	"github.com/aws/smithy-go/document"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"k8s.io/utils/ptr"

	"github.com/crossplane-contrib/provider-aws/apis/elasticloadbalancing/v1alpha1"
)

var (
	elbName           = "someELB"
	availabilityZones = []string{"us-east-1a", "us-east-1b"}
	listener          = v1alpha1.Listener{
		InstancePort:     80,
		InstanceProtocol: aws.String("HTTP"),
		LoadBalancerPort: 80,
		Protocol:         "HTTP",
	}
	elbListener = elbtypes.Listener{
		InstancePort:     ptr.To[int32](80),
		InstanceProtocol: aws.String("HTTP"),
		LoadBalancerPort: int32(80),
		Protocol:         aws.String("HTTP"),
	}
	scheme  = "internal"
	elbTags = []elbtypes.Tag{
		{
			Key:   aws.String("k1"),
			Value: aws.String("v1"),
		},
		{
			Key:   aws.String("k2"),
			Value: aws.String("v2"),
		},
	}
	tags = []v1alpha1.Tag{
		{
			Key:   "k1",
			Value: aws.String("v1"),
		},
		{
			Key:   "k2",
			Value: aws.String("v2"),
		},
	}
)

func elbParams(m ...func(*v1alpha1.ELBParameters)) *v1alpha1.ELBParameters {
	o := &v1alpha1.ELBParameters{
		AvailabilityZones: availabilityZones,
		Listeners:         []v1alpha1.Listener{listener},
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func loadBalancer(m ...func(*elbtypes.LoadBalancerDescription)) *elbtypes.LoadBalancerDescription {
	o := &elbtypes.LoadBalancerDescription{
		AvailabilityZones:    availabilityZones,
		ListenerDescriptions: []elbtypes.ListenerDescription{{Listener: &elbListener}},
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func TestLateInitializeELB(t *testing.T) {
	type args struct {
		spec *v1alpha1.ELBParameters
		in   elbtypes.LoadBalancerDescription
		tags []elbtypes.Tag
	}
	cases := map[string]struct {
		args args
		want *v1alpha1.ELBParameters
	}{
		"AllFilledNoDiff": {
			args: args{
				spec: elbParams(),
				in:   *loadBalancer(),
			},
			want: elbParams(),
		},
		"AllFilledExternalDiff": {
			args: args{
				spec: elbParams(),
				in: *loadBalancer(func(lb *elbtypes.LoadBalancerDescription) {
					lb.Scheme = aws.String(scheme)
				}),
			},
			want: elbParams(func(p *v1alpha1.ELBParameters) {
				p.Scheme = aws.String(scheme)
			}),
		},
		"PartialFilled": {
			args: args{
				spec: elbParams(func(p *v1alpha1.ELBParameters) {
					p.AvailabilityZones = nil
				}),
				in: *loadBalancer(),
			},
			want: elbParams(),
		},
		"Tags": {
			args: args{
				spec: elbParams(),
				in:   *loadBalancer(),
				tags: elbTags,
			},
			want: elbParams(func(p *v1alpha1.ELBParameters) {
				p.Tags = tags
			}),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitializeELB(tc.args.spec, &tc.args.in, tc.args.tags)
			if diff := cmp.Diff(tc.args.spec, tc.want); diff != "" {
				t.Errorf("LateInitializeSpec(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateCreateRoleInput(t *testing.T) {
	cases := map[string]struct {
		in  v1alpha1.ELBParameters
		out elb.CreateLoadBalancerInput
	}{
		"FilledInput": {
			in: *elbParams(),
			out: elb.CreateLoadBalancerInput{
				LoadBalancerName:  &elbName,
				AvailabilityZones: availabilityZones,
				Listeners:         []elbtypes.Listener{elbListener},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateCreateELBInput(elbName, tc.in)
			if diff := cmp.Diff(r, &tc.out, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("GenerateNetworkObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestBuildELBListeners(t *testing.T) {
	cases := map[string]struct {
		in  []v1alpha1.Listener
		out []elbtypes.Listener
	}{
		"FilledInput": {
			in:  []v1alpha1.Listener{listener},
			out: []elbtypes.Listener{elbListener},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := BuildELBListeners(tc.in)
			if diff := cmp.Diff(r, tc.out, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("GenerateNetworkObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestBuildELBTags(t *testing.T) {
	cases := map[string]struct {
		in  []v1alpha1.Tag
		out []elbtypes.Tag
	}{
		"FilledInput": {
			in:  tags,
			out: elbTags,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := BuildELBTags(tc.in)
			if diff := cmp.Diff(r, tc.out, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("GenerateNetworkObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestCreatePatch(t *testing.T) {
	type args struct {
		lb   elbtypes.LoadBalancerDescription
		p    v1alpha1.ELBParameters
		tags []elbtypes.Tag
	}

	type want struct {
		patch *v1alpha1.ELBParameters
	}

	cases := map[string]struct {
		args
		want
	}{
		"SameFields": {
			args: args{
				lb: elbtypes.LoadBalancerDescription{
					AvailabilityZones: availabilityZones,
					ListenerDescriptions: []elbtypes.ListenerDescription{{
						Listener: &elbListener,
					}},
				},
				p: v1alpha1.ELBParameters{
					AvailabilityZones: availabilityZones,
					Listeners:         []v1alpha1.Listener{listener},
				},
			},
			want: want{
				patch: &v1alpha1.ELBParameters{},
			},
		},
		"DifferentOrder": {
			args: args{
				lb: elbtypes.LoadBalancerDescription{
					AvailabilityZones: availabilityZones,
					Subnets:           []string{"sub1", "sub2"},
					SecurityGroups:    []string{"sg1", "sg2"},
				},
				p: v1alpha1.ELBParameters{
					AvailabilityZones: []string{"us-east-1b", "us-east-1a"},
					SubnetIDs:         []string{"sub2", "sub1"},
					SecurityGroupIDs:  []string{"sg2", "sg1"},
					Tags:              tags,
				},
				tags: []elbtypes.Tag{{
					Key:   aws.String("k2"),
					Value: aws.String("v2"),
				},
					{
						Key:   aws.String("k1"),
						Value: aws.String("v1"),
					}},
			},
			want: want{
				patch: &v1alpha1.ELBParameters{},
			},
		},
		"DifferentFields": {
			args: args{
				lb: elbtypes.LoadBalancerDescription{
					AvailabilityZones: availabilityZones,
					ListenerDescriptions: []elbtypes.ListenerDescription{{
						Listener: &elbListener,
					}},
					Subnets: []string{"subnet1", "subnet2"},
				},
				p: v1alpha1.ELBParameters{
					AvailabilityZones: availabilityZones,
					Listeners:         []v1alpha1.Listener{listener},
					SubnetIDs:         []string{"subnet1", "subnet3"},
				},
			},
			want: want{
				patch: &v1alpha1.ELBParameters{
					SubnetIDs: []string{"subnet1", "subnet3"},
				},
			},
		},
		"DifferentTags": {
			args: args{
				tags: elbTags,
				p: v1alpha1.ELBParameters{
					Tags: []v1alpha1.Tag{
						{Key: "k1", Value: aws.String("v1")},
					},
				},
			},
			want: want{
				patch: &v1alpha1.ELBParameters{
					Tags: []v1alpha1.Tag{
						{Key: "k1", Value: aws.String("v1")},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			result, _ := CreatePatch(tc.args.lb, tc.args.p, tc.args.tags)
			if diff := cmp.Diff(tc.want.patch, result); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsUpToDate(t *testing.T) {
	type args struct {
		lb   elbtypes.LoadBalancerDescription
		p    v1alpha1.ELBParameters
		tags []elbtypes.Tag
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFields": {
			args: args{
				lb: elbtypes.LoadBalancerDescription{
					AvailabilityZones: availabilityZones,
					ListenerDescriptions: []elbtypes.ListenerDescription{{
						Listener: &elbListener,
					}},
				},
				p: v1alpha1.ELBParameters{
					AvailabilityZones: availabilityZones,
					Listeners:         []v1alpha1.Listener{listener},
				},
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				lb: elbtypes.LoadBalancerDescription{
					AvailabilityZones: availabilityZones,
					ListenerDescriptions: []elbtypes.ListenerDescription{{
						Listener: &elbListener,
					}},
					SecurityGroups: []string{"sg1", "sg2"},
				},
				p: v1alpha1.ELBParameters{
					AvailabilityZones: availabilityZones,
					Listeners:         []v1alpha1.Listener{listener},
					SecurityGroupIDs:  []string{"sg1", "sg3"},
				},
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, _ := IsUpToDate(tc.args.p, tc.args.lb, tc.args.tags)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
