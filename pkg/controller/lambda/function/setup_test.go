package function

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/lambda"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane-contrib/provider-aws/apis/lambda/v1beta1"
	svcapitypesv1beta1 "github.com/crossplane-contrib/provider-aws/apis/lambda/v1beta1"
)

type args struct {
	cr  *v1beta1.Function
	obj *svcsdk.GetFunctionOutput
}

type functionModifier func(*v1beta1.Function)

func withSpec(p v1beta1.FunctionParameters) functionModifier {
	return func(r *v1beta1.Function) { r.Spec.ForProvider = p }
}

func function(m ...functionModifier) *v1beta1.Function {
	cr := &v1beta1.Function{}
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
	}

	cases := map[string]struct {
		args
		want
	}{
		"NilSourceNoUpdate": {
			args: args{
				cr:  function(withSpec(v1beta1.FunctionParameters{})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{Environment: &svcsdk.EnvironmentResponse{Variables: map[string]*string{}}}},
			},
			want: want{
				result: true,
			},
		},
		"NilSourceNilAwsNoUpdate": {
			args: args{
				cr:  function(withSpec(v1beta1.FunctionParameters{})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{}},
			},
			want: want{
				result: true,
			},
		},
		"EmptySourceEmptyAWSNoUpdate": {
			args: args{
				cr: function(withSpec(v1beta1.FunctionParameters{
					Environment: &v1beta1.Environment{
						Variables: map[string]*string{},
					}})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{Environment: &svcsdk.EnvironmentResponse{Variables: map[string]*string{}}}},
			},
			want: want{
				result: true,
			},
		},
		"NilSourceEnvAWSWithUpdate": {
			args: args{
				cr: function(withSpec(v1beta1.FunctionParameters{})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{Environment: &svcsdk.EnvironmentResponse{
					Variables: map[string]*string{"tagKey2": aws.String("tagValue2")}}}},
			},
			want: want{
				result: false,
			},
		},
		"EnvSourceNilAwsWithUpdate": {
			args: args{
				cr: function(withSpec(v1beta1.FunctionParameters{
					Environment: &v1beta1.Environment{
						Variables: map[string]*string{"tagKey1": aws.String("tagValue1")},
					}})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{}},
			},
			want: want{
				result: false,
			},
		},
		"NeedsUpdateDiffKeys": {
			args: args{
				cr: function(withSpec(v1beta1.FunctionParameters{
					Environment: &v1beta1.Environment{
						Variables: map[string]*string{"tagKey1": aws.String("tagValue1")},
					}})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{Environment: &svcsdk.EnvironmentResponse{
					Variables: map[string]*string{"tagKey2": aws.String("tagValue2")}}}},
			},
			want: want{
				result: false,
			},
		},
		"NeedsUpdateDiffValues": {
			args: args{
				cr: function(withSpec(v1beta1.FunctionParameters{
					Environment: &v1beta1.Environment{
						Variables: map[string]*string{"tagKey1": aws.String("tagValue1")},
					}})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{Environment: &svcsdk.EnvironmentResponse{
					Variables: map[string]*string{"tagKey1": aws.String("tagValue2")}}}},
			},
			want: want{
				result: false,
			},
		},
		"NoUpdateNeeded": {
			args: args{
				cr: function(withSpec(v1beta1.FunctionParameters{
					Environment: &v1beta1.Environment{
						Variables: map[string]*string{"tagKey1": aws.String("tagValue1")},
					}})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{Environment: &svcsdk.EnvironmentResponse{
					Variables: map[string]*string{"tagKey1": aws.String("tagValue1")}}}},
			},
			want: want{
				result: true,
			},
		},
		"NoUpdateNeededOutOfOrder": {
			args: args{
				cr: function(withSpec(v1beta1.FunctionParameters{
					Environment: &v1beta1.Environment{
						Variables: map[string]*string{"tagKey1": aws.String("tagValue1"), "tagKey2": aws.String("tagValue2"), "tagKey3": aws.String("tagValue3")},
					}})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{Environment: &svcsdk.EnvironmentResponse{
					Variables: map[string]*string{"tagKey3": aws.String("tagValue3"), "tagKey2": aws.String("tagValue2"), "tagKey1": aws.String("tagValue1")}}}},
			},
			want: want{
				result: true,
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
				cr:  function(withSpec(v1beta1.FunctionParameters{})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{FileSystemConfigs: []*svcsdk.FileSystemConfig{}}},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
		"NilSourceNilAwsNoUpdate": {
			args: args{
				cr:  function(withSpec(v1beta1.FunctionParameters{})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{}},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
		"EmptySourceNoUpdate": {
			args: args{
				cr: function(withSpec(v1beta1.FunctionParameters{
					FileSystemConfigs: []*v1beta1.FileSystemConfig{}})),
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
				cr: function(withSpec(v1beta1.FunctionParameters{})),
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
				cr: function(withSpec(v1beta1.FunctionParameters{
					FileSystemConfigs: []*v1beta1.FileSystemConfig{{ARN: aws.String("arn1"), LocalMountPath: aws.String(" localMountPath1")}}})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"NeedsUpdateArnAndMounthPath": {
			args: args{
				cr: function(withSpec(v1beta1.FunctionParameters{
					FileSystemConfigs: []*v1beta1.FileSystemConfig{{ARN: aws.String("arn1"), LocalMountPath: aws.String(" localMountPath1")}}})),
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
				cr: function(withSpec(v1beta1.FunctionParameters{
					FileSystemConfigs: []*v1beta1.FileSystemConfig{{ARN: aws.String("arn1"), LocalMountPath: aws.String(" localMountPath1")}}})),
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
				cr: function(withSpec(v1beta1.FunctionParameters{
					FileSystemConfigs: []*v1beta1.FileSystemConfig{{ARN: aws.String("arn1"), LocalMountPath: aws.String(" localMountPath1")}}})),
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
				cr: function(withSpec(v1beta1.FunctionParameters{
					FileSystemConfigs: []*v1beta1.FileSystemConfig{{ARN: aws.String("arn1"), LocalMountPath: aws.String(" localMountPath1")}}})),
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
				cr: function(withSpec(v1beta1.FunctionParameters{
					FileSystemConfigs: []*v1beta1.FileSystemConfig{{ARN: aws.String("arn1"), LocalMountPath: aws.String(" localMountPath1")},
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
				cr: function(withSpec(v1beta1.FunctionParameters{})),
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
				cr: function(withSpec(v1beta1.FunctionParameters{
					TracingConfig: &v1beta1.TracingConfig{}})),
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
				cr: function(withSpec(v1beta1.FunctionParameters{})),
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
				cr: function(withSpec(v1beta1.FunctionParameters{
					TracingConfig: &v1beta1.TracingConfig{Mode: aws.String(svcsdk.TracingModeActive)}})),
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
				cr: function(withSpec(v1beta1.FunctionParameters{
					TracingConfig: &v1beta1.TracingConfig{Mode: aws.String(svcsdk.TracingModeActive)}})),
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
				cr: function(withSpec(v1beta1.FunctionParameters{})),
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
				cr:  function(withSpec(v1beta1.FunctionParameters{})),
				obj: &svcsdk.GetFunctionOutput{Configuration: &svcsdk.FunctionConfiguration{}},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
		"EmptySourceNoUpdate": {
			args: args{
				cr: function(withSpec(v1beta1.FunctionParameters{
					CustomFunctionParameters: v1beta1.CustomFunctionParameters{
						CustomFunctionVPCConfigParameters: &v1beta1.CustomFunctionVPCConfigParameters{
							SecurityGroupIDs: []*string{},
						},
					}})),
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
				cr: function(withSpec(v1beta1.FunctionParameters{})),
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
				cr: function(withSpec(v1beta1.FunctionParameters{
					CustomFunctionParameters: v1beta1.CustomFunctionParameters{
						CustomFunctionVPCConfigParameters: &v1beta1.CustomFunctionVPCConfigParameters{
							SecurityGroupIDs: []*string{aws.String("id1")},
						},
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
				cr: function(withSpec(v1beta1.FunctionParameters{
					CustomFunctionParameters: v1beta1.CustomFunctionParameters{
						CustomFunctionVPCConfigParameters: &v1beta1.CustomFunctionVPCConfigParameters{
							SecurityGroupIDs: []*string{aws.String("id1")},
						},
					}})),
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
				cr: function(withSpec(v1beta1.FunctionParameters{
					CustomFunctionParameters: v1beta1.CustomFunctionParameters{
						CustomFunctionVPCConfigParameters: &v1beta1.CustomFunctionVPCConfigParameters{
							SecurityGroupIDs: []*string{aws.String("id1"), aws.String("id2")},
						},
					}})),
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
				cr: function(withSpec(v1beta1.FunctionParameters{
					CustomFunctionParameters: v1beta1.CustomFunctionParameters{
						CustomFunctionVPCConfigParameters: &v1beta1.CustomFunctionVPCConfigParameters{
							SecurityGroupIDs: []*string{aws.String("id1")},
						},
					}})),
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

func TestGenerateUpdateFunctionCodeInput(t *testing.T) {
	type args struct {
		cr *v1beta1.Function
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
				cr: function(withSpec(v1beta1.FunctionParameters{
					CustomFunctionParameters: v1beta1.CustomFunctionParameters{
						CustomFunctionCodeParameters: v1beta1.CustomFunctionCodeParameters{
							ImageURI: aws.String("test_image"),
							S3Bucket: aws.String("test_bucket"),
							S3Key:    aws.String("test_key"),
						},
					},
				}))},
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
		cr *v1beta1.Function
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
				cr: function(withSpec(v1beta1.FunctionParameters{
					DeadLetterConfig: &v1beta1.DeadLetterConfig{TargetARN: aws.String("test_dlqArn")},
					Description:      aws.String("test_description"),
					Environment: &v1beta1.Environment{
						Variables: map[string]*string{"tagKey1": aws.String("tagValue1")},
					},
					FileSystemConfigs: []*v1beta1.FileSystemConfig{{ARN: aws.String("arn1")}},
					Handler:           aws.String("test_handler"),
					KMSKeyARN:         aws.String("test_kms"),
					MemorySize:        aws.Int64(128),
					Runtime:           aws.String("test_runtime"),
					Timeout:           aws.Int64(128),
					TracingConfig:     &v1beta1.TracingConfig{Mode: aws.String(svcsdk.TracingModeActive)},
					CustomFunctionParameters: v1beta1.CustomFunctionParameters{
						Role: aws.String("test_role"),
						CustomFunctionVPCConfigParameters: &v1beta1.CustomFunctionVPCConfigParameters{
							SecurityGroupIDs: []*string{aws.String("id1")},
						},
					},
				},
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
					VpcConfig: &svcsdk.VpcConfig{
						SecurityGroupIds: []*string{aws.String("id1")},
						SubnetIds:        []*string{},
					},
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

func TestIsUpToDateCodeImage(t *testing.T) {
	type args struct {
		cr  *svcapitypesv1beta1.Function
		obj *svcsdk.GetFunctionOutput
	}
	type want struct {
		codeUpToDate bool
	}

	cases := map[string]struct {
		args
		want
	}{
		"UpdToDateForZipCodeSupply": {
			args: args{
				cr: &svcapitypesv1beta1.Function{
					Spec: svcapitypesv1beta1.FunctionSpec{
						ForProvider: svcapitypesv1beta1.FunctionParameters{
							PackageType: aws.String(packageTypeZip),
						},
					},
				},
				obj: &svcsdk.GetFunctionOutput{},
			},
			want: want{
				codeUpToDate: true,
			},
		},
		"NotUpToDateIfNothingIsSet": {
			args: args{
				cr: &svcapitypesv1beta1.Function{
					Spec: svcapitypesv1beta1.FunctionSpec{
						ForProvider: svcapitypesv1beta1.FunctionParameters{
							PackageType: aws.String(packageTypeImage),
						},
					},
				},
				obj: &svcsdk.GetFunctionOutput{},
			},
			want: want{
				codeUpToDate: false,
			},
		},
		"NotUpToDateIfNotConfiguration": {
			args: args{
				cr: &svcapitypesv1beta1.Function{
					Spec: svcapitypesv1beta1.FunctionSpec{
						ForProvider: svcapitypesv1beta1.FunctionParameters{
							PackageType: aws.String(packageTypeImage),
						},
					},
				},
				obj: &svcsdk.GetFunctionOutput{
					Configuration: nil,
				},
			},
			want: want{
				codeUpToDate: false,
			},
		},
		"NotUpToDateIfNotConfigurationPackageType": {
			args: args{
				cr: &svcapitypesv1beta1.Function{
					Spec: svcapitypesv1beta1.FunctionSpec{
						ForProvider: svcapitypesv1beta1.FunctionParameters{
							PackageType: aws.String(packageTypeImage),
						},
					},
				},
				obj: &svcsdk.GetFunctionOutput{
					Configuration: &svcsdk.FunctionConfiguration{
						PackageType: nil,
					},
				},
			},
			want: want{
				codeUpToDate: false,
			},
		},
		"NotUpToDateIfConfigurationPackageTypeZip": {
			args: args{
				cr: &svcapitypesv1beta1.Function{
					Spec: svcapitypesv1beta1.FunctionSpec{
						ForProvider: svcapitypesv1beta1.FunctionParameters{
							PackageType: aws.String(packageTypeImage),
						},
					},
				},
				obj: &svcsdk.GetFunctionOutput{
					Configuration: &svcsdk.FunctionConfiguration{
						PackageType: aws.String(packageTypeZip),
					},
				},
			},
			want: want{
				codeUpToDate: false,
			},
		},
		"NotUpToDateIfNoCode": {
			args: args{
				cr: &svcapitypesv1beta1.Function{
					Spec: svcapitypesv1beta1.FunctionSpec{
						ForProvider: svcapitypesv1beta1.FunctionParameters{
							PackageType: aws.String(packageTypeImage),
						},
					},
				},
				obj: &svcsdk.GetFunctionOutput{
					Configuration: &svcsdk.FunctionConfiguration{
						PackageType: aws.String(packageTypeImage),
					},
					Code: nil,
				},
			},
			want: want{
				codeUpToDate: false,
			},
		},
		"NotUpToDateIfEmptyCode": {
			args: args{
				cr: &svcapitypesv1beta1.Function{
					Spec: svcapitypesv1beta1.FunctionSpec{
						ForProvider: svcapitypesv1beta1.FunctionParameters{
							PackageType: aws.String(packageTypeImage),
						},
					},
				},
				obj: &svcsdk.GetFunctionOutput{
					Configuration: &svcsdk.FunctionConfiguration{
						PackageType: aws.String(packageTypeImage),
					},
					Code: &svcsdk.FunctionCodeLocation{},
				},
			},
			want: want{
				codeUpToDate: false,
			},
		},
		"NotUpToDateIfCodePackageTypeNotECR": {
			args: args{
				cr: &svcapitypesv1beta1.Function{
					Spec: svcapitypesv1beta1.FunctionSpec{
						ForProvider: svcapitypesv1beta1.FunctionParameters{
							PackageType: aws.String(packageTypeImage),
						},
					},
				},
				obj: &svcsdk.GetFunctionOutput{
					Configuration: &svcsdk.FunctionConfiguration{
						PackageType: aws.String(packageTypeImage),
					},
					Code: &svcsdk.FunctionCodeLocation{
						RepositoryType: aws.String(repositoryTypeS3),
					},
				},
			},
			want: want{
				codeUpToDate: false,
			},
		},
		"NotUpToDateIfNoCodeImageURI": {
			args: args{
				cr: &svcapitypesv1beta1.Function{
					Spec: svcapitypesv1beta1.FunctionSpec{
						ForProvider: svcapitypesv1beta1.FunctionParameters{
							PackageType: aws.String(packageTypeImage),
							CustomFunctionParameters: v1beta1.CustomFunctionParameters{
								CustomFunctionCodeParameters: v1beta1.CustomFunctionCodeParameters{
									ImageURI: aws.String("ecr.aws/repository/foo-image"),
								},
							},
						},
					},
				},
				obj: &svcsdk.GetFunctionOutput{
					Configuration: &svcsdk.FunctionConfiguration{
						PackageType: aws.String(packageTypeImage),
					},
					Code: &svcsdk.FunctionCodeLocation{
						ImageUri:       nil,
						RepositoryType: aws.String(repositoryTypeECR),
					},
				},
			},
			want: want{
				codeUpToDate: false,
			},
		},
		"NotUpToDateIfCodeImageNotMatch": {
			args: args{
				cr: &svcapitypesv1beta1.Function{
					Spec: svcapitypesv1beta1.FunctionSpec{
						ForProvider: svcapitypesv1beta1.FunctionParameters{
							PackageType: aws.String(packageTypeImage),
							CustomFunctionParameters: v1beta1.CustomFunctionParameters{
								CustomFunctionCodeParameters: v1beta1.CustomFunctionCodeParameters{
									ImageURI: aws.String("ecr.aws/repository/foo-image"),
								},
							},
						},
					},
				},
				obj: &svcsdk.GetFunctionOutput{
					Configuration: &svcsdk.FunctionConfiguration{
						PackageType: aws.String(packageTypeImage),
					},
					Code: &svcsdk.FunctionCodeLocation{
						ImageUri:       aws.String("ecr.aws/repository/bar-image"),
						RepositoryType: aws.String(repositoryTypeECR),
					},
				},
			},
			want: want{
				codeUpToDate: false,
			},
		},
		"UpToDate": {
			args: args{
				cr: &svcapitypesv1beta1.Function{
					Spec: svcapitypesv1beta1.FunctionSpec{
						ForProvider: svcapitypesv1beta1.FunctionParameters{
							PackageType: aws.String(packageTypeImage),
							CustomFunctionParameters: v1beta1.CustomFunctionParameters{
								CustomFunctionCodeParameters: v1beta1.CustomFunctionCodeParameters{
									ImageURI: aws.String("ecr.aws/repository/foo-image"),
								},
							},
						},
					},
				},
				obj: &svcsdk.GetFunctionOutput{
					Configuration: &svcsdk.FunctionConfiguration{
						PackageType: aws.String(packageTypeImage),
					},
					Code: &svcsdk.FunctionCodeLocation{
						ImageUri:       aws.String("ecr.aws/repository/foo-image"),
						RepositoryType: aws.String(repositoryTypeECR),
					},
				},
			},
			want: want{
				codeUpToDate: true,
			},
		},
		"UpToDateIfBothImageURIsNil": {
			args: args{
				cr: &svcapitypesv1beta1.Function{
					Spec: svcapitypesv1beta1.FunctionSpec{
						ForProvider: svcapitypesv1beta1.FunctionParameters{
							PackageType: aws.String(packageTypeImage),
							CustomFunctionParameters: v1beta1.CustomFunctionParameters{
								CustomFunctionCodeParameters: v1beta1.CustomFunctionCodeParameters{
									ImageURI: nil,
								},
							},
						},
					},
				},
				obj: &svcsdk.GetFunctionOutput{
					Configuration: &svcsdk.FunctionConfiguration{
						PackageType: aws.String(packageTypeImage),
					},
					Code: &svcsdk.FunctionCodeLocation{
						ImageUri:       nil,
						RepositoryType: aws.String(repositoryTypeECR),
					},
				},
			},
			want: want{
				codeUpToDate: true,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			actualUpToDate := isUpToDateCodeImage(tc.args.cr, tc.args.obj)

			// Assert
			if diff := cmp.Diff(tc.want.codeUpToDate, actualUpToDate); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
