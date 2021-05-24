package function

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/lambda"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/lambda/v1alpha1"
)

type args struct {
	cr  *v1alpha1.Function
	obj *svcsdk.GetFunctionOutput
}

type functionModifier func(*v1alpha1.Function)

func withSpec(p v1alpha1.FunctionParameters) functionModifier {
	return func(r *v1alpha1.Function) { r.Spec.ForProvider = p }
}

func function(m ...functionModifier) *v1alpha1.Function {
	cr := &v1alpha1.Function{}
	cr.Name = "test-function-name"
	for _, f := range m {
		f(cr)
	}
	return cr
}

var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connector{}

func TestIsUpToDateEnvironment(t *testing.T) {
	type want struct {
		result bool
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"NilSourceNoUpdate": {
			args: args{
				cr:  function(withSpec(v1alpha1.FunctionParameters{})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{Environment: &svcsdk.EnvironmentResponse{Variables: map[string]*string{}}}},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
		"NilSourceNilAwsNoUpdate": {
			args: args{
				cr:  function(withSpec(v1alpha1.FunctionParameters{})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{}},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
		"EmptySourceNoUpdate": {
			args: args{
				cr: function(withSpec(v1alpha1.FunctionParameters{
					Environment: &v1alpha1.Environment{
						Variables: map[string]*string{},
					}})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{Environment: &svcsdk.EnvironmentResponse{Variables: map[string]*string{}}}},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
		"NilSourceWithUpdate": {
			args: args{
				cr: function(withSpec(v1alpha1.FunctionParameters{})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{Environment: &svcsdk.EnvironmentResponse{
					Variables: map[string]*string{"tagKey2": aws.String("tagValue2")}}}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"NilAwsWithUpdate": {
			args: args{
				cr: function(withSpec(v1alpha1.FunctionParameters{
					Environment: &v1alpha1.Environment{
						Variables: map[string]*string{"tagKey1": aws.String("tagValue1")},
					}})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"NeedsUpdate": {
			args: args{
				cr: function(withSpec(v1alpha1.FunctionParameters{
					Environment: &v1alpha1.Environment{
						Variables: map[string]*string{"tagKey1": aws.String("tagValue1")},
					}})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{Environment: &svcsdk.EnvironmentResponse{
					Variables: map[string]*string{"tagKey2": aws.String("tagValue2")}}}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"NoUpdateNeeded": {
			args: args{
				cr: function(withSpec(v1alpha1.FunctionParameters{
					Environment: &v1alpha1.Environment{
						Variables: map[string]*string{"tagKey1": aws.String("tagValue1")},
					}})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{Environment: &svcsdk.EnvironmentResponse{
					Variables: map[string]*string{"tagKey1": aws.String("tagValue1")}}}},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
		"NoUpdateNeededOutOfOrder": {
			args: args{
				cr: function(withSpec(v1alpha1.FunctionParameters{
					Environment: &v1alpha1.Environment{
						Variables: map[string]*string{"tagKey1": aws.String("tagValue1"), "tagKey2": aws.String("tagValue2"), "tagKey3": aws.String("tagValue3")},
					}})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{Environment: &svcsdk.EnvironmentResponse{
					Variables: map[string]*string{"tagKey3": aws.String("tagValue3"), "tagKey2": aws.String("tagValue2"), "tagKey1": aws.String("tagValue1")}}}},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Act
			result := isUpToDateEnvironment(tc.args.cr, tc.args.obj)

			// Assert
			if diff := cmp.Diff(tc.want.result, result, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsUpToDateFileSystemConfigs(t *testing.T) {
	type want struct {
		result bool
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"NilSourceNoUpdate": {
			args: args{
				cr:  function(withSpec(v1alpha1.FunctionParameters{})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{FileSystemConfigs: []*svcsdk.FileSystemConfig{}}},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
		"NilSourceNilAwsNoUpdate": {
			args: args{
				cr:  function(withSpec(v1alpha1.FunctionParameters{})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{}},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
		"EmptySourceNoUpdate": {
			args: args{
				cr: function(withSpec(v1alpha1.FunctionParameters{
					FileSystemConfigs: []*v1alpha1.FileSystemConfig{}})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{
					FileSystemConfigs: []*svcsdk.FileSystemConfig{}}},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
		"NilSourceWithUpdate": {
			args: args{
				cr: function(withSpec(v1alpha1.FunctionParameters{})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{
					FileSystemConfigs: []*svcsdk.FileSystemConfig{{Arn: aws.String("arn1"), LocalMountPath: aws.String(" localMountPath1")}}}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"NilAwsWithUpdate": {
			args: args{
				cr: function(withSpec(v1alpha1.FunctionParameters{
					FileSystemConfigs: []*v1alpha1.FileSystemConfig{{ARN: aws.String("arn1"), LocalMountPath: aws.String(" localMountPath1")}}})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"NeedsUpdateArnAndMounthPath": {
			args: args{
				cr: function(withSpec(v1alpha1.FunctionParameters{
					FileSystemConfigs: []*v1alpha1.FileSystemConfig{{ARN: aws.String("arn1"), LocalMountPath: aws.String(" localMountPath1")}}})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{
					FileSystemConfigs: []*svcsdk.FileSystemConfig{{Arn: aws.String("arn2"), LocalMountPath: aws.String(" localMountPath2")}}}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"NeedsUpdateArn": {
			args: args{
				cr: function(withSpec(v1alpha1.FunctionParameters{
					FileSystemConfigs: []*v1alpha1.FileSystemConfig{{ARN: aws.String("arn1"), LocalMountPath: aws.String(" localMountPath1")}}})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{
					FileSystemConfigs: []*svcsdk.FileSystemConfig{{Arn: aws.String("arn2"), LocalMountPath: aws.String(" localMountPath1")}}}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"NeedsUpdateMountPath": {
			args: args{
				cr: function(withSpec(v1alpha1.FunctionParameters{
					FileSystemConfigs: []*v1alpha1.FileSystemConfig{{ARN: aws.String("arn1"), LocalMountPath: aws.String(" localMountPath1")}}})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{
					FileSystemConfigs: []*svcsdk.FileSystemConfig{{Arn: aws.String("arn1"), LocalMountPath: aws.String(" localMountPath2")}}}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"NoUpdateNeeded": {
			args: args{
				cr: function(withSpec(v1alpha1.FunctionParameters{
					FileSystemConfigs: []*v1alpha1.FileSystemConfig{{ARN: aws.String("arn1"), LocalMountPath: aws.String(" localMountPath1")}}})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{
					FileSystemConfigs: []*svcsdk.FileSystemConfig{{Arn: aws.String("arn1"), LocalMountPath: aws.String(" localMountPath1")}}}},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
		"NoUpdateNeededSortOrder": {
			args: args{
				cr: function(withSpec(v1alpha1.FunctionParameters{
					FileSystemConfigs: []*v1alpha1.FileSystemConfig{{ARN: aws.String("arn1"), LocalMountPath: aws.String(" localMountPath1")},
						{ARN: aws.String("arn2"), LocalMountPath: aws.String(" localMountPath2")}}})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{
					FileSystemConfigs: []*svcsdk.FileSystemConfig{{Arn: aws.String("arn2"), LocalMountPath: aws.String(" localMountPath2")},
						{Arn: aws.String("arn1"), LocalMountPath: aws.String(" localMountPath1")},
					}}},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Act
			result := isUpToDateFileSystemConfigs(tc.args.cr, tc.args.obj)

			// Assert
			if diff := cmp.Diff(tc.want.result, result, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsUpToDateTracingConfig(t *testing.T) {
	type want struct {
		result bool
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"NilSourceNoUpdate": {
			args: args{
				cr: function(withSpec(v1alpha1.FunctionParameters{})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{
					TracingConfig: &svcsdk.TracingConfigResponse{
						Mode: aws.String(svcsdk.TracingModePassThrough)}}},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
		"EmptySourceNoUpdate": {
			args: args{
				cr: function(withSpec(v1alpha1.FunctionParameters{
					TracingConfig: &v1alpha1.TracingConfig{}})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{
					TracingConfig: &svcsdk.TracingConfigResponse{
						Mode: aws.String(svcsdk.TracingModePassThrough)}}},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
		"NilSourceWithUpdate": {
			args: args{
				cr: function(withSpec(v1alpha1.FunctionParameters{})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{
					TracingConfig: &svcsdk.TracingConfigResponse{
						Mode: aws.String(svcsdk.TracingModeActive)}}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"NeedsUpdate": {
			args: args{
				cr: function(withSpec(v1alpha1.FunctionParameters{
					TracingConfig: &v1alpha1.TracingConfig{Mode: aws.String(svcsdk.TracingModeActive)}})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{
					TracingConfig: &svcsdk.TracingConfigResponse{
						Mode: aws.String(svcsdk.TracingModePassThrough)}}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"NoUpdateNeeded": {
			args: args{
				cr: function(withSpec(v1alpha1.FunctionParameters{
					TracingConfig: &v1alpha1.TracingConfig{Mode: aws.String(svcsdk.TracingModeActive)}})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{
					TracingConfig: &svcsdk.TracingConfigResponse{
						Mode: aws.String(svcsdk.TracingModeActive)}}},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Act
			result := isUpToDateTracingConfig(tc.args.cr, tc.args.obj)

			// Assert
			if diff := cmp.Diff(tc.want.result, result, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsUpToDateSecurityGroupIDs(t *testing.T) {
	type want struct {
		result bool
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"NilSourceNoUpdate": {
			args: args{
				cr: function(withSpec(v1alpha1.FunctionParameters{})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{
					VpcConfig: &svcsdk.VpcConfigResponse{SecurityGroupIds: []*string{}}}},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
		"NilSourceNilAwsNoUpdate": {
			args: args{
				cr:  function(withSpec(v1alpha1.FunctionParameters{})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{}},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
		"EmptySourceNoUpdate": {
			args: args{
				cr: function(withSpec(v1alpha1.FunctionParameters{
					VPCConfig: &v1alpha1.VPCConfig{SecurityGroupIDs: []*string{}}})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{
					VpcConfig: &svcsdk.VpcConfigResponse{SecurityGroupIds: []*string{}}}},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
		"NilSourceWithUpdate": {
			args: args{
				cr: function(withSpec(v1alpha1.FunctionParameters{})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{
					VpcConfig: &svcsdk.VpcConfigResponse{SecurityGroupIds: []*string{aws.String("id1")}}}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"NilAwsWithUpdate": {
			args: args{
				cr: function(withSpec(v1alpha1.FunctionParameters{
					VPCConfig: &v1alpha1.VPCConfig{SecurityGroupIDs: []*string{aws.String("id1")}}})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"NeedsUpdate": {
			args: args{
				cr: function(withSpec(v1alpha1.FunctionParameters{
					VPCConfig: &v1alpha1.VPCConfig{SecurityGroupIDs: []*string{aws.String("id1")}}})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{
					VpcConfig: &svcsdk.VpcConfigResponse{SecurityGroupIds: []*string{aws.String("id2")}}}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"NoUpdateNeededSortOrderIsDifferent": {
			args: args{
				cr: function(withSpec(v1alpha1.FunctionParameters{
					VPCConfig: &v1alpha1.VPCConfig{SecurityGroupIDs: []*string{aws.String("id1"), aws.String("id2")}}})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{
					VpcConfig: &svcsdk.VpcConfigResponse{SecurityGroupIds: []*string{aws.String("id2"), aws.String("id1")}}}},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
		"NoUpdateNeeded": {
			args: args{
				cr: function(withSpec(v1alpha1.FunctionParameters{
					VPCConfig: &v1alpha1.VPCConfig{SecurityGroupIDs: []*string{aws.String("id1")}}})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{
					VpcConfig: &svcsdk.VpcConfigResponse{SecurityGroupIds: []*string{aws.String("id1")}}}},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Act
			result := isUpToDateSecurityGroupIDs(tc.args.cr, tc.args.obj)

			// Assert
			if diff := cmp.Diff(tc.want.result, result, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDiffTags(t *testing.T) {
	type args struct {
		cr  map[string]*string
		obj map[string]*string
	}
	type want struct {
		addTags    map[string]*string
		removeTags []*string
	}

	cases := map[string]struct {
		args
		want
	}{
		"AddNewTag": {
			args: args{
				cr: map[string]*string{
					"k1": aws.String("exists_in_both"),
					"k2": aws.String("only_in_cr"),
				},
				obj: map[string]*string{
					"k1": aws.String("exists_in_both"),
				}},
			want: want{
				addTags: map[string]*string{
					"k2": aws.String("only_in_cr"),
				},
				removeTags: []*string{},
			},
		},
		"RemoveExistingTag": {
			args: args{
				cr: map[string]*string{
					"k1": aws.String("exists_in_both"),
				},
				obj: map[string]*string{
					"k1": aws.String("exists_in_both"),
					"k2": aws.String("only_in_aws"),
				}},
			want: want{
				addTags: map[string]*string{},
				removeTags: []*string{
					aws.String("k2"),
				}},
		},
		"AddAndRemoveWhenKeyChanges": {
			args: args{
				cr: map[string]*string{
					"k1": aws.String("exists_in_both"),
					"k2": aws.String("same_key_different_value_1"),
				},
				obj: map[string]*string{
					"k1": aws.String("exists_in_both"),
					"k2": aws.String("same_key_different_value_2"),
				}},
			want: want{
				addTags: map[string]*string{
					"k2": aws.String("same_key_different_value_1"),
				},
				removeTags: []*string{
					aws.String("k2"),
				}},
		},
		"NoChange": {
			args: args{
				cr: map[string]*string{
					"k1": aws.String("exists_in_both"),
				},
				obj: map[string]*string{
					"k1": aws.String("exists_in_both"),
				}},
			want: want{
				addTags:    map[string]*string{},
				removeTags: []*string{},
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Act
			addTags, removeTags := diffTags(tc.args.cr, tc.args.obj)

			// Assert
			if diff := cmp.Diff(tc.want.addTags, addTags, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.removeTags, removeTags, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateUpdateFunctionCodeInput(t *testing.T) {
	type args struct {
		cr *v1alpha1.Function
	}
	type want struct {
		obj *svcsdk.UpdateFunctionCodeInput
	}

	cases := map[string]struct {
		args
		want
	}{
		"ConvertS3": {
			args: args{
				cr: function(withSpec(v1alpha1.FunctionParameters{Code: &v1alpha1.FunctionCode{ImageURI: aws.String("test_image"),
					S3Bucket: aws.String("test_bucket"),
					S3Key:    aws.String("test_key")}}))},
			want: want{
				obj: &svcsdk.UpdateFunctionCodeInput{
					FunctionName: aws.String("test-function-name"),
					ImageUri:     aws.String("test_image"),
					S3Bucket:     aws.String("test_bucket"),
					S3Key:        aws.String("test_key")}},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Act
			actual := GenerateUpdateFunctionCodeInput(tc.args.cr)

			// Assert
			if diff := cmp.Diff(tc.want.obj, actual, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateUpdateFunctionConfigurationInput(t *testing.T) {
	type args struct {
		cr *v1alpha1.Function
	}
	type want struct {
		obj *svcsdk.UpdateFunctionConfigurationInput
	}

	cases := map[string]struct {
		args
		want
	}{
		"ConvertToUpdateFunctionConfigurationInput": {
			args: args{
				cr: function(withSpec(v1alpha1.FunctionParameters{
					DeadLetterConfig: &v1alpha1.DeadLetterConfig{TargetARN: aws.String("test_dlqArn")},
					Description:      aws.String("test_description"),
					Environment: &v1alpha1.Environment{
						Variables: map[string]*string{"tagKey1": aws.String("tagValue1")},
					},
					FileSystemConfigs: []*v1alpha1.FileSystemConfig{{ARN: aws.String("arn1")}},
					Handler:           aws.String("test_handler"),
					KMSKeyARN:         aws.String("test_kms"),
					MemorySize:        aws.Int64(128),
					Role:              aws.String("test_role"),
					Runtime:           aws.String("test_runtime"),
					Timeout:           aws.Int64(128),
					TracingConfig:     &v1alpha1.TracingConfig{Mode: aws.String(svcsdk.TracingModeActive)},
					VPCConfig:         &v1alpha1.VPCConfig{SecurityGroupIDs: []*string{aws.String("id1")}}},
				))},
			want: want{
				obj: &svcsdk.UpdateFunctionConfigurationInput{
					DeadLetterConfig: &svcsdk.DeadLetterConfig{TargetArn: aws.String("test_dlqArn")},
					Description:      aws.String("test_description"),
					Environment: &svcsdk.Environment{
						Variables: map[string]*string{"tagKey1": aws.String("tagValue1")},
					},
					FileSystemConfigs: []*svcsdk.FileSystemConfig{{Arn: aws.String("arn1")}},
					FunctionName:      aws.String("test-function-name"),
					Handler:           aws.String("test_handler"),
					KMSKeyArn:         aws.String("test_kms"),
					MemorySize:        aws.Int64(128),
					Role:              aws.String("test_role"),
					Runtime:           aws.String("test_runtime"),
					Timeout:           aws.Int64(128),
					TracingConfig:     &svcsdk.TracingConfig{Mode: aws.String(svcsdk.TracingModeActive)},
					VpcConfig:         &svcsdk.VpcConfig{SecurityGroupIds: []*string{aws.String("id1")}},
				},
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Act
			actual := GenerateUpdateFunctionConfigurationInput(tc.args.cr)

			// Assert
			if diff := cmp.Diff(tc.want.obj, actual, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
