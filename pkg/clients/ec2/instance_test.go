/*
Copyright 2021 The Crossplane Authors.
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

package ec2

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-aws/apis/ec2/manualv1alpha1"
)

const (
	managedName           = "sample-instance"
	managedKind           = "instance.ec2.aws.crossplane.io"
	managedProviderConfig = "example"
)

func TestGenerateInstanceConditions(t *testing.T) {
	type args struct {
		obeserved manualv1alpha1.InstanceObservation
	}
	cases := map[string]struct {
		args args
		want Condition
	}{
		"AllInstancesAreRunning": {
			args: args{
				obeserved: manualv1alpha1.InstanceObservation{
					State: manualv1alpha1.InstancesState{
						Running: 6,
						Total:   6,
					},
				},
			},
			want: Available,
		},
		"SomeInstancesRunningAndSomePending": {
			args: args{
				obeserved: manualv1alpha1.InstanceObservation{
					State: manualv1alpha1.InstancesState{
						Running: 4,
						Pending: 2,
						Total:   6,
					},
				},
			},
			want: Creating,
		},
		"SomeInstancesRunningAndSomeStopping": {
			args: args{
				obeserved: manualv1alpha1.InstanceObservation{
					State: manualv1alpha1.InstancesState{
						Running:  4,
						Stopping: 2,
						Total:    6,
					},
				},
			},
			want: Deleting,
		},
		"SomeInstancesPendingAndSomeShuttingDown": {
			args: args{
				obeserved: manualv1alpha1.InstanceObservation{
					State: manualv1alpha1.InstancesState{
						Running:  4,
						Stopping: 2,
						Total:    6,
					},
				},
			},
			want: Deleting,
		},
		"AllInstancesAreTerminated": {
			args: args{
				obeserved: manualv1alpha1.InstanceObservation{
					State: manualv1alpha1.InstancesState{
						Terminated: 6,
						Total:      6,
					},
				},
			},
			want: Deleted,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			condition := GenerateInstanceCondition(tc.args.obeserved)

			if diff := cmp.Diff(tc.want, condition, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateDescribeInstancesByExternalTags(t *testing.T) {
	type args struct {
		extTags map[string]string
	}
	cases := map[string]struct {
		args args
		want *ec2.DescribeInstancesInput
	}{
		"TagsAreAddedToFilter": {
			args: args{
				extTags: map[string]string{
					"crossplane-name":           managedName,
					"crossplane-kind":           managedKind,
					"crossplane-providerconfig": managedProviderConfig,
				},
			},
			want: &ec2.DescribeInstancesInput{
				Filters: []ec2.Filter{
					{
						Name:   aws.String("tag:crossplane-kind"),
						Values: []string{managedKind},
					},
					{
						Name:   aws.String("tag:crossplane-name"),
						Values: []string{managedName},
					},
					{
						Name:   aws.String("tag:crossplane-providerconfig"),
						Values: []string{managedProviderConfig},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			input := GenerateDescribeInstancesByExternalTags(tc.args.extTags)

			if diff := cmp.Diff(tc.want, input, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
