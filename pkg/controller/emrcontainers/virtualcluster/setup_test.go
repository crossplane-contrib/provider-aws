package virtualcluster

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/emrcontainers"
	"github.com/crossplane-contrib/provider-aws/apis/emrcontainers/v1alpha1"
	v1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	ignoreTransitionOpt = cmpopts.IgnoreFields(v1.Condition{}, "LastTransitionTime")
)

func TestPostObserve(t *testing.T) {
	type args struct {
		cr            *v1alpha1.VirtualCluster
		clusterOutput *svcsdk.DescribeVirtualClusterOutput
		obs           managed.ExternalObservation
	}

	type want struct {
		status v1.ConditionedStatus
		err    error
		result managed.ExternalObservation
	}

	cases := map[string]struct {
		args
		want
	}{
		"running": {
			args: args{
				cr: &v1alpha1.VirtualCluster{},
				clusterOutput: &svcsdk.DescribeVirtualClusterOutput{
					VirtualCluster: &svcsdk.VirtualCluster{
						State: aws.String(svcsdk.VirtualClusterStateRunning),
					},
				},
			},
			want: want{
				err: nil,
				status: v1.ConditionedStatus{Conditions: []v1.Condition{
					{
						Type:   "Ready",
						Status: "True",
						Reason: "Available",
					},
				}},
			},
		},
		"invalidState": {
			args: args{
				cr: &v1alpha1.VirtualCluster{
					Status: v1alpha1.VirtualClusterStatus{
						ResourceStatus: v1.ResourceStatus{
							v1.ConditionedStatus{Conditions: []v1.Condition{
								{
									Type:   "Ready",
									Status: "False",
									Reason: "Unavailable",
								},
							}},
						},
					},
				},
				clusterOutput: &svcsdk.DescribeVirtualClusterOutput{
					VirtualCluster: &svcsdk.VirtualCluster{
						State: aws.String("invalid"),
					},
				},
			},
			want: want{
				err: nil,
				status: v1.ConditionedStatus{Conditions: []v1.Condition{
					{
						Type:   "Ready",
						Status: "False",
						Reason: "Unavailable",
					},
				}},
			},
		},
		"terminated": {
			args: args{
				cr: &v1alpha1.VirtualCluster{
					Status: v1alpha1.VirtualClusterStatus{},
				},
				clusterOutput: &svcsdk.DescribeVirtualClusterOutput{
					VirtualCluster: &svcsdk.VirtualCluster{
						State: aws.String(svcsdk.VirtualClusterStateTerminated),
					},
				},
				obs: managed.ExternalObservation{
					ResourceExists: true,
				},
			},
			want: want{
				err: nil,
				result: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			obs, _ := postObserve(context.Background(), tc.args.cr, tc.args.clusterOutput, tc.args.obs, nil)

			diff := cmp.Diff(tc.want.status, tc.cr.Status.ConditionedStatus, test.EquateConditions(), ignoreTransitionOpt)
			if diff != "" {
				t.Errorf("diff -want, +got:\n%s", diff)
			}
			obsDiff := cmp.Diff(tc.want.result, obs, test.EquateConditions())
			if obsDiff != "" {
				t.Errorf("obsDiff -want, +got:\n%s", diff)
			}
		})
	}
}

func TestPreObserve(t *testing.T) {
	type args struct {
		cr           *v1alpha1.VirtualCluster
		clusterInput *svcsdk.DescribeVirtualClusterInput
	}

	type want struct {
		err    error
		result *svcsdk.DescribeVirtualClusterInput
	}

	cases := map[string]struct {
		args
		want
	}{
		"noResourceExists": {
			args: args{
				cr: &v1alpha1.VirtualCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name: "abc",
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: "abc",
						},
					},
				},
				clusterInput: &svcsdk.DescribeVirtualClusterInput{},
			},
			want: want{
				err: nil,
				result: &svcsdk.DescribeVirtualClusterInput{
					Id: aws.String(firstObserveId),
				},
			},
		},
		"resourceExists": {
			args: args{
				cr: &v1alpha1.VirtualCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name: "abc",
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: "external-name",
						},
					},
				},
				clusterInput: &svcsdk.DescribeVirtualClusterInput{},
			},
			want: want{
				err: nil,
				result: &svcsdk.DescribeVirtualClusterInput{
					Id: aws.String("external-name"),
				},
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := preObserve(context.Background(), tc.args.cr, tc.args.clusterInput)
			errDiff := cmp.Diff(tc.want.err, err, test.EquateErrors())
			if errDiff != "" {
				t.Errorf("diff: %s", errDiff)
			}
			diff := cmp.Diff(tc.want.result, tc.args.clusterInput, test.EquateConditions())
			if diff != "" {
				t.Errorf("-want, +got:\n%s", diff)
			}
		})
	}
}
