package jobrun

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/emrcontainers"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane-contrib/provider-aws/apis/emrcontainers/v1alpha1"
)

func newMockExternal(output output) external {
	return external{
		client: &MockClient{
			output: output,
		},
	}
}

func TestPreObserve(t *testing.T) {
	type args struct {
		cr    *v1alpha1.JobRun
		input *svcsdk.DescribeJobRunInput
	}

	type want struct {
		err    error
		result *svcsdk.DescribeJobRunInput
	}

	cases := map[string]struct {
		args
		want
	}{
		"noResourceExists": {
			args: args{
				cr: &v1alpha1.JobRun{
					ObjectMeta: metav1.ObjectMeta{
						Name: "abc",
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: "abc",
						},
					},
				},
				input: &svcsdk.DescribeJobRunInput{},
			},
			want: want{
				result: &svcsdk.DescribeJobRunInput{
					Id: aws.String(firstObserveJobRunID),
				},
			},
		},
		"resourceExists": {
			args: args{
				cr: &v1alpha1.JobRun{
					ObjectMeta: metav1.ObjectMeta{
						Name: "abc",
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: "external-name",
						},
					},
				},
				input: &svcsdk.DescribeJobRunInput{},
			},
			want: want{
				err: nil,
				result: &svcsdk.DescribeJobRunInput{
					Id: aws.String("external-name"),
				},
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := preObserve(context.Background(), tc.args.cr, tc.args.input)
			errDiff := cmp.Diff(tc.want.err, err, test.EquateConditions())
			if errDiff != "" {
				t.Errorf("diff: %s", errDiff)
			}
			diff := cmp.Diff(tc.want.result, tc.args.input, test.EquateConditions())
			if diff != "" {
				t.Errorf("-want, +got:\n%s", diff)
			}
		})
	}
}

func TestDeleter(t *testing.T) {
	validateErrMsg := fmt.Sprintf("%s: Job run 00000000000 is not in a cancellable state", svcsdk.ErrCodeValidationException)
	validationErr := errors.Wrap(errors.New(validateErrMsg), "wrapped")

	type args struct {
		err                  error
		cancelJobRunOutput   svcsdk.CancelJobRunOutput
		describeJobRunOutput svcsdk.DescribeJobRunOutput
	}

	type want struct {
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"jobCompleted": {
			args: args{
				err:                  validationErr,
				describeJobRunOutput: svcsdk.DescribeJobRunOutput{JobRun: &svcsdk.JobRun{State: aws.String(svcsdk.JobRunStateCompleted)}},
			},
			want: want{
				err: nil,
			},
		},
		"jobRunning": {
			args: args{
				err:                  validationErr,
				describeJobRunOutput: svcsdk.DescribeJobRunOutput{JobRun: &svcsdk.JobRun{State: aws.String(svcsdk.JobRunStateRunning)}},
			},
			want: want{
				err: validationErr,
			},
		},
		"noError": {
			args: args{
				err: nil,
			},
			want: want{
				err: nil,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ext := newMockExternal(output{
				nil, nil, nil, nil, &tc.args.describeJobRunOutput,
			})
			_, result := ext.postDeleter(context.Background(), &v1alpha1.JobRun{}, &tc.args.cancelJobRunOutput, tc.args.err)
			diff := cmp.Diff(tc.want.err, result, test.EquateErrors())
			if diff != "" {
				t.Errorf("-want, +got:\n%s", diff)
			}
		})
	}
}
