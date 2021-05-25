package function

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/lambda"
	svcsdkapi "github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/lambda/v1alpha1"
	svcapitypes "github.com/crossplane/provider-aws/apis/lambda/v1alpha1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
)

// SetupFunction adds a controller that reconciles Function.
func SetupFunction(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(v1alpha1.FunctionGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preDelete = preDelete
			e.preCreate = preCreate
			e.isUpToDate = isUpToDate
			e.lateInitialize = LateInitialize
			u := &updater{client: e.client}
			e.update = u.update
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&v1alpha1.Function{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.FunctionGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

// LateInitialize fills the empty fields in *svcapitypes.FunctionParameters with
// the values seen in svcsdk.GetFunctionOutput.
func LateInitialize(cr *svcapitypes.FunctionParameters, resp *svcsdk.GetFunctionOutput) error {
	cr.MemorySize = aws.LateInitializeInt64Ptr(cr.MemorySize, resp.Configuration.MemorySize)
	cr.Timeout = aws.LateInitializeInt64Ptr(cr.Timeout, resp.Configuration.Timeout)
	if cr.TracingConfig == nil {
		cr.TracingConfig = &svcapitypes.TracingConfig{Mode: resp.Configuration.TracingConfig.Mode}
	}
	return nil
}

func preCreate(_ context.Context, cr *svcapitypes.Function, obj *svcsdk.CreateFunctionInput) error {
	obj.FunctionName = aws.String(meta.GetExternalName(cr))
	return nil
}

func preObserve(_ context.Context, cr *svcapitypes.Function, obj *svcsdk.GetFunctionInput) error {
	obj.FunctionName = aws.String(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.Function, resp *svcsdk.GetFunctionOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	switch aws.StringValue(resp.Configuration.State) {
	case string(svcapitypes.State_Active):
		cr.SetConditions(xpv1.Available())
	case string(svcapitypes.State_Pending):
		cr.SetConditions(xpv1.Creating())
	case string(svcapitypes.State_Failed), string(svcapitypes.State_Inactive):
		cr.SetConditions(xpv1.Unavailable())
	}
	return obs, nil
}

func preDelete(_ context.Context, cr *svcapitypes.Function, obj *svcsdk.DeleteFunctionInput) (bool, error) {
	obj.FunctionName = aws.String(meta.GetExternalName(cr))
	return false, nil
}

// nolint:gocyclo
func isUpToDate(cr *svcapitypes.Function, obj *svcsdk.GetFunctionOutput) (bool, error) {

	// Compare CODE
	// GetFunctionOutput returns
	// Code *FunctionCodeLocation `type:"structure"`
	// which does not map to
	// Code *FunctionCode `type:"structure" required:"true"`
	// which is used when creating the function.
	// We can't currently properly implement a comparison

	// Compare CONFIGURATION
	if aws.StringValue(cr.Spec.ForProvider.Description) != aws.StringValue(obj.Configuration.Description) {
		return false, nil
	}

	if !isUpToDateEnvironment(cr, obj) {
		return false, nil
	}

	// Connection settings for an Amazon EFS file system.
	if !isUpToDateFileSystemConfigs(cr, obj) {
		return false, nil
	}

	if aws.StringValue(cr.Spec.ForProvider.Handler) != aws.StringValue(obj.Configuration.Handler) {
		return false, nil
	}

	if aws.StringValue(cr.Spec.ForProvider.KMSKeyARN) != aws.StringValue(obj.Configuration.KMSKeyArn) {
		return false, nil
	}

	// The function's layers (https://docs.aws.amazon.com/lambda/latest/dg/configuration-layers.html).
	// The generator is creating layers with type of
	// 	Layers []*string `json:"layers,omitempty"`
	// Instead of []*svcsdk.Layer
	// We can't properly implement a comparison until that is fixed

	// set default
	if aws.Int64Value(cr.Spec.ForProvider.MemorySize) != aws.Int64Value(obj.Configuration.MemorySize) {
		return false, nil
	}

	if aws.StringValue(cr.Spec.ForProvider.Role) != aws.StringValue(obj.Configuration.Role) {
		return false, nil
	}

	if aws.StringValue(cr.Spec.ForProvider.Runtime) != aws.StringValue(obj.Configuration.Runtime) {
		return false, nil
	}

	if aws.Int64Value(cr.Spec.ForProvider.Timeout) != aws.Int64Value(obj.Configuration.Timeout) {
		return false, nil
	}

	// This should never be nil.  We set this in LateInit as aws will initialize a default value
	if aws.StringValue(cr.Spec.ForProvider.TracingConfig.Mode) != aws.StringValue(obj.Configuration.TracingConfig.Mode) {
		return false, nil
	}

	if !isUpToDateSecurityGroupIDs(cr, obj) {
		return false, nil
	}

	addTags, removeTags := diffTags(cr.Spec.ForProvider.Tags, obj.Tags)
	return len(addTags) == 0 && len(removeTags) == 0, nil

}

// isUpToDateEnvironment checks if FunctionConfiguration EnvironmentResponse Variables are up to date
func isUpToDateEnvironment(cr *svcapitypes.Function, obj *svcsdk.GetFunctionOutput) bool {
	// Handle nil pointer refs
	envVars := map[string]*string{}
	awsVars := map[string]*string{}
	if cr.Spec.ForProvider.Environment != nil &&
		cr.Spec.ForProvider.Environment.Variables != nil {
		envVars = cr.Spec.ForProvider.Environment.Variables
	}
	if obj.Configuration.Environment != nil &&
		obj.Configuration.Environment.Variables != nil {
		awsVars = obj.Configuration.Environment.Variables
	}

	// Compare whether the maps are equal, ignore ordering
	sortCmp := cmpopts.SortMaps(func(i, j *string) bool {
		return aws.StringValue(i) < aws.StringValue(j)
	})

	return cmp.Equal(envVars, awsVars, sortCmp, cmpopts.EquateEmpty())
}

func isUpToDateFileSystemConfigs(cr *svcapitypes.Function, obj *svcsdk.GetFunctionOutput) bool {
	// Handle nil pointer refs
	fileSystemConfigs := make([]*svcsdk.FileSystemConfig, 0)
	awsFileSystemConfigs := make([]*svcsdk.FileSystemConfig, 0)
	if cr.Spec.ForProvider.FileSystemConfigs != nil {
		for _, v := range cr.Spec.ForProvider.FileSystemConfigs {
			fileSystemConfigs = append(fileSystemConfigs, &svcsdk.FileSystemConfig{
				Arn:            v.ARN,
				LocalMountPath: v.LocalMountPath,
			})
		}
	}
	if obj.Configuration.FileSystemConfigs != nil {
		awsFileSystemConfigs = obj.Configuration.FileSystemConfigs
	}

	if len(fileSystemConfigs) != len(awsFileSystemConfigs) {
		return false
	}
	// Compare whether the slices are equal, ignore ordering
	sortCmp := cmpopts.SortSlices(func(i, j *svcsdk.FileSystemConfig) bool {
		return aws.StringValue(i.Arn) < aws.StringValue(j.Arn)
	})

	return cmp.Equal(fileSystemConfigs, awsFileSystemConfigs, sortCmp)
}

func isUpToDateTracingConfig(cr *svcapitypes.Function, obj *svcsdk.GetFunctionOutput) bool {
	// Handle nil pointer refs
	mode := svcsdk.TracingModePassThrough
	if cr.Spec.ForProvider.TracingConfig != nil &&
		cr.Spec.ForProvider.TracingConfig.Mode != nil {
		mode = *cr.Spec.ForProvider.TracingConfig.Mode
	}
	if mode != aws.StringValue(obj.Configuration.TracingConfig.Mode) {
		return false
	}
	return true
}

func isUpToDateSecurityGroupIDs(cr *svcapitypes.Function, obj *svcsdk.GetFunctionOutput) bool {
	// Handle nil pointer refs
	var securityGroupIDs []*string
	var awsSecurityGroupIDs []*string
	if cr.Spec.ForProvider.VPCConfig != nil &&
		cr.Spec.ForProvider.VPCConfig.SecurityGroupIDs != nil {
		securityGroupIDs = cr.Spec.ForProvider.VPCConfig.SecurityGroupIDs
	}
	if obj.Configuration.VpcConfig != nil &&
		obj.Configuration.VpcConfig.SecurityGroupIds != nil {
		awsSecurityGroupIDs = obj.Configuration.VpcConfig.SecurityGroupIds
	}

	// Compare whether the slices are equal, ignore ordering
	sortCmp := cmpopts.SortSlices(func(i, j *string) bool {
		return aws.StringValue(i) < aws.StringValue(j)
	})

	return cmp.Equal(securityGroupIDs, awsSecurityGroupIDs, sortCmp, cmpopts.EquateEmpty())
}

// returns which AWS Tags exist in the resource tags and which are outdated and should be removed
func diffTags(spec map[string]*string, current map[string]*string) (map[string]*string, []*string) {
	addMap := make(map[string]*string, len(spec))
	removeTags := make([]*string, 0)
	for k, v := range current {
		if aws.StringValue(spec[k]) == aws.StringValue(v) {
			continue
		}
		removeTags = append(removeTags, aws.String(k))
	}
	for k, v := range spec {
		if aws.StringValue(current[k]) == aws.StringValue(v) {
			continue
		}
		addMap[k] = v
	}
	return addMap, removeTags
}

type updater struct {
	client svcsdkapi.LambdaAPI
}

func (u *updater) update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*svcapitypes.Function)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	// https://docs.aws.amazon.com/sdk-for-go/api/service/lambda/#Lambda.UpdateFunctionCode
	updateFunctionCodeInput := GenerateUpdateFunctionCodeInput(cr)
	if _, err := u.client.UpdateFunctionCodeWithContext(ctx, updateFunctionCodeInput); err != nil {
		return managed.ExternalUpdate{}, aws.Wrap(err, errUpdate)
	}

	updateFunctionConfigurationInput := GenerateUpdateFunctionConfigurationInput(cr)
	if _, err := u.client.UpdateFunctionConfigurationWithContext(ctx, updateFunctionConfigurationInput); err != nil {
		return managed.ExternalUpdate{}, aws.Wrap(err, errUpdate)
	}

	// Should store the ARN somewhere else?
	functionConfiguration, err := u.client.GetFunctionConfigurationWithContext(ctx, &svcsdk.GetFunctionConfigurationInput{
		FunctionName: aws.String(meta.GetExternalName(cr)),
	})
	if err != nil {
		return managed.ExternalUpdate{}, aws.Wrap(err, errUpdate)
	}

	// Tags
	tags, err := u.client.ListTagsWithContext(ctx, &svcsdk.ListTagsInput{
		Resource: functionConfiguration.FunctionArn,
	})
	if err != nil {
		return managed.ExternalUpdate{}, aws.Wrap(err, errUpdate)
	}

	addTags, removeTags := diffTags(cr.Spec.ForProvider.Tags, tags.Tags)
	// Remove old tags before adding new tags in case values change for keys
	if len(removeTags) > 0 {
		if _, err := u.client.UntagResourceWithContext(ctx, &svcsdk.UntagResourceInput{
			Resource: functionConfiguration.FunctionArn,
			TagKeys:  removeTags,
		}); err != nil {
			return managed.ExternalUpdate{}, aws.Wrap(err, errUpdate)
		}
	}
	if len(addTags) > 0 {
		if _, err := u.client.TagResourceWithContext(ctx, &svcsdk.TagResourceInput{
			Resource: functionConfiguration.FunctionArn,
			Tags:     addTags,
		}); err != nil {
			return managed.ExternalUpdate{}, aws.Wrap(err, errUpdate)
		}
	}

	//	if _, err := u.client.UpdateFunctionEventInvokeConfigWithContext(ctx, &svcsdk.UpdateFunctionEventInvokeConfigInput{
	//		FunctionName:       aws.String(meta.GetExternalName(cr)),
	//		DestinationConfig : cr.Spec.ForProvider.DestinationConfig .,
	//	}); err != nil {
	//		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdate)
	//	}

	return managed.ExternalUpdate{}, nil
}

// GenerateUpdateFunctionCodeInput is similar to GenerateCreateFunctionConfigurationInput
// Copied almost verbatim from the zz_conversions generated code
func GenerateUpdateFunctionCodeInput(cr *svcapitypes.Function) *svcsdk.UpdateFunctionCodeInput {
	f0 := &svcsdk.UpdateFunctionCodeInput{}
	f0.SetFunctionName(cr.Name)
	if cr.Spec.ForProvider.Code != nil {
		if cr.Spec.ForProvider.Code.ImageURI != nil {
			f0.SetImageUri(*cr.Spec.ForProvider.Code.ImageURI)
		}
		if cr.Spec.ForProvider.Code.S3Bucket != nil {
			f0.SetS3Bucket(*cr.Spec.ForProvider.Code.S3Bucket)
		}
		if cr.Spec.ForProvider.Code.S3Key != nil {
			f0.SetS3Key(*cr.Spec.ForProvider.Code.S3Key)
		}
		if cr.Spec.ForProvider.Code.S3ObjectVersion != nil {
			f0.SetS3ObjectVersion(*cr.Spec.ForProvider.Code.S3ObjectVersion)
		}
		if cr.Spec.ForProvider.Code.ZipFile != nil {
			f0.SetZipFile(cr.Spec.ForProvider.Code.ZipFile)
		}
	}
	return f0
}

// GenerateUpdateFunctionConfigurationInput is similar to GenerateCreateFunctionConfigurationInput
// Copied almost verbatim from the zz_conversions generated code
// nolint:gocyclo
func GenerateUpdateFunctionConfigurationInput(cr *svcapitypes.Function) *svcsdk.UpdateFunctionConfigurationInput {
	res := &svcsdk.UpdateFunctionConfigurationInput{}
	res.SetFunctionName(cr.Name)

	if cr.Spec.ForProvider.DeadLetterConfig != nil {
		f2 := &svcsdk.DeadLetterConfig{}
		if cr.Spec.ForProvider.DeadLetterConfig.TargetARN != nil {
			f2.SetTargetArn(*cr.Spec.ForProvider.DeadLetterConfig.TargetARN)
		}
		res.SetDeadLetterConfig(f2)
	}
	if cr.Spec.ForProvider.Description != nil {
		res.SetDescription(*cr.Spec.ForProvider.Description)
	}
	if cr.Spec.ForProvider.Environment != nil {
		f4 := &svcsdk.Environment{}
		if cr.Spec.ForProvider.Environment.Variables != nil {
			f4f0 := map[string]*string{}
			for f4f0key, f4f0valiter := range cr.Spec.ForProvider.Environment.Variables {
				var f4f0val = *f4f0valiter
				f4f0[f4f0key] = &f4f0val
			}
			f4.SetVariables(f4f0)
		}
		res.SetEnvironment(f4)
	}
	if cr.Spec.ForProvider.FileSystemConfigs != nil {
		f5 := []*svcsdk.FileSystemConfig{}
		for _, f5iter := range cr.Spec.ForProvider.FileSystemConfigs {
			f5elem := &svcsdk.FileSystemConfig{}
			if f5iter.ARN != nil {
				f5elem.SetArn(*f5iter.ARN)
			}
			if f5iter.LocalMountPath != nil {
				f5elem.SetLocalMountPath(*f5iter.LocalMountPath)
			}
			f5 = append(f5, f5elem)
		}
		res.SetFileSystemConfigs(f5)
	}
	if cr.Spec.ForProvider.Handler != nil {
		res.SetHandler(*cr.Spec.ForProvider.Handler)
	}
	if cr.Spec.ForProvider.ImageConfig != nil {
		f8 := &svcsdk.ImageConfig{}
		if cr.Spec.ForProvider.ImageConfig.Command != nil {
			f8f0 := []*string{}
			for _, f8f0iter := range cr.Spec.ForProvider.ImageConfig.Command {
				var f8f0elem = *f8f0iter
				f8f0 = append(f8f0, &f8f0elem)
			}
			f8.SetCommand(f8f0)
		}
		if cr.Spec.ForProvider.ImageConfig.EntryPoint != nil {
			f8f1 := []*string{}
			for _, f8f1iter := range cr.Spec.ForProvider.ImageConfig.EntryPoint {
				var f8f1elem = *f8f1iter
				f8f1 = append(f8f1, &f8f1elem)
			}
			f8.SetEntryPoint(f8f1)
		}
		if cr.Spec.ForProvider.ImageConfig.WorkingDirectory != nil {
			f8.SetWorkingDirectory(*cr.Spec.ForProvider.ImageConfig.WorkingDirectory)
		}
		res.SetImageConfig(f8)
	}
	if cr.Spec.ForProvider.KMSKeyARN != nil {
		res.SetKMSKeyArn(*cr.Spec.ForProvider.KMSKeyARN)
	}
	if cr.Spec.ForProvider.Layers != nil {
		f10 := []*string{}
		for _, f10iter := range cr.Spec.ForProvider.Layers {
			var f10elem = *f10iter
			f10 = append(f10, &f10elem)
		}
		res.SetLayers(f10)
	}
	if cr.Spec.ForProvider.MemorySize != nil {
		res.SetMemorySize(*cr.Spec.ForProvider.MemorySize)
	}
	if cr.Spec.ForProvider.Role != nil {
		res.SetRole(*cr.Spec.ForProvider.Role)
	}
	if cr.Spec.ForProvider.Runtime != nil {
		res.SetRuntime(*cr.Spec.ForProvider.Runtime)
	}
	if cr.Spec.ForProvider.Timeout != nil {
		res.SetTimeout(*cr.Spec.ForProvider.Timeout)
	}
	if cr.Spec.ForProvider.TracingConfig != nil {
		f18 := &svcsdk.TracingConfig{}
		if cr.Spec.ForProvider.TracingConfig.Mode != nil {
			f18.SetMode(*cr.Spec.ForProvider.TracingConfig.Mode)
		}
		res.SetTracingConfig(f18)
	}
	if cr.Spec.ForProvider.VPCConfig != nil {
		f19 := &svcsdk.VpcConfig{}
		if cr.Spec.ForProvider.VPCConfig.SecurityGroupIDs != nil {
			f19f0 := []*string{}
			for _, f19f0iter := range cr.Spec.ForProvider.VPCConfig.SecurityGroupIDs {
				var f19f0elem = *f19f0iter
				f19f0 = append(f19f0, &f19f0elem)
			}
			f19.SetSecurityGroupIds(f19f0)
		}
		if cr.Spec.ForProvider.VPCConfig.SubnetIDs != nil {
			f19f1 := []*string{}
			for _, f19f1iter := range cr.Spec.ForProvider.VPCConfig.SubnetIDs {
				var f19f1elem = *f19f1iter
				f19f1 = append(f19f1, &f19f1elem)
			}
			f19.SetSubnetIds(f19f1)
		}
		res.SetVpcConfig(f19)
	}
	return res
}
