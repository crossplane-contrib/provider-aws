package function

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/lambda"
	svcsdkapi "github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/lambda/v1beta1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
	tagutils "github.com/crossplane-contrib/provider-aws/pkg/utils/tags"
)

const (
	isLastUpdateStatusSuccessfulCheckInterval = 30 * time.Second

	// used in creation
	packageTypeImage = string(svcapitypes.PackageType_Image)
	packageTypeZip   = string(svcapitypes.PackageType_Zip)

	// used in observation
	repositoryTypeECR = "ECR"
	repositoryTypeS3  = "S3"
)

// SetupFunction adds a controller that reconciles Function.
func SetupFunction(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.FunctionGroupKind)
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

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(svcapitypes.FunctionGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.Function{}).
		Complete(r)
}

// LateInitialize fills the empty fields in *svcapitypes.FunctionParameters with
// the values seen in svcsdk.GetFunctionOutput.
func LateInitialize(cr *svcapitypes.FunctionParameters, resp *svcsdk.GetFunctionOutput) error {
	cr.MemorySize = pointer.LateInitialize(cr.MemorySize, resp.Configuration.MemorySize)
	cr.Timeout = pointer.LateInitialize(cr.Timeout, resp.Configuration.Timeout)
	if cr.TracingConfig == nil {
		cr.TracingConfig = &svcapitypes.TracingConfig{Mode: resp.Configuration.TracingConfig.Mode}
	}
	return nil
}

func preCreate(_ context.Context, cr *svcapitypes.Function, obj *svcsdk.CreateFunctionInput) error {
	obj.FunctionName = aws.String(meta.GetExternalName(cr))
	obj.Role = cr.Spec.ForProvider.Role
	obj.Code = &svcsdk.FunctionCode{
		ImageUri:        cr.Spec.ForProvider.CustomFunctionCodeParameters.ImageURI,
		S3Bucket:        cr.Spec.ForProvider.CustomFunctionCodeParameters.S3Bucket,
		S3Key:           cr.Spec.ForProvider.CustomFunctionCodeParameters.S3Key,
		S3ObjectVersion: cr.Spec.ForProvider.CustomFunctionCodeParameters.S3ObjectVersion,
	}
	if cr.Spec.ForProvider.CustomFunctionVPCConfigParameters != nil {
		obj.VpcConfig = &svcsdk.VpcConfig{
			SecurityGroupIds: cr.Spec.ForProvider.CustomFunctionVPCConfigParameters.SecurityGroupIDs,
			SubnetIds:        cr.Spec.ForProvider.CustomFunctionVPCConfigParameters.SubnetIDs,
		}
	}
	obj.Layers = cr.Spec.ForProvider.Layers
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
	cr.Status.AtProvider = generateFuntionObservation(resp)
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

//nolint:gocyclo
func isUpToDate(_ context.Context, cr *svcapitypes.Function, obj *svcsdk.GetFunctionOutput) (bool, string, error) {

	// Compare CODE
	// GetFunctionOutput returns
	// Code *FunctionCodeLocation `type:"structure"`
	// which does not map to
	// Code *FunctionCode `type:"structure" required:"true"`
	// which is used when creating the function.
	// As of 2022-11-04 we can't currently properly implement a full comparison.
	// It is partially possible for code supplied via FunctionCode.ImageUri

	if !isUpToDateCodeImage(cr, obj) {
		return false, "", nil
	}

	// Compare CONFIGURATION
	if aws.StringValue(cr.Spec.ForProvider.Description) != aws.StringValue(obj.Configuration.Description) {
		return false, "", nil
	}

	if !isUpToDateEnvironment(cr, obj) {
		return false, "", nil
	}

	// Connection settings for an Amazon EFS file system.
	if !isUpToDateFileSystemConfigs(cr, obj) {
		return false, "", nil
	}

	if aws.StringValue(cr.Spec.ForProvider.Handler) != aws.StringValue(obj.Configuration.Handler) {
		return false, "", nil
	}

	if aws.StringValue(cr.Spec.ForProvider.KMSKeyARN) != aws.StringValue(obj.Configuration.KMSKeyArn) {
		return false, "", nil
	}

	// The function's layers (https://docs.aws.amazon.com/lambda/latest/dg/configuration-layers.html).
	// The generator is creating layers with type of
	// 	Layers []*string `json:"layers,omitempty"`
	// Instead of []*svcsdk.Layer
	// We can't properly implement a comparison until that is fixed

	// set default
	if aws.Int64Value(cr.Spec.ForProvider.MemorySize) != aws.Int64Value(obj.Configuration.MemorySize) {
		return false, "", nil
	}

	if aws.StringValue(cr.Spec.ForProvider.Role) != aws.StringValue(obj.Configuration.Role) {
		return false, "", nil
	}

	if aws.StringValue(cr.Spec.ForProvider.Runtime) != aws.StringValue(obj.Configuration.Runtime) {
		return false, "", nil
	}

	if aws.Int64Value(cr.Spec.ForProvider.Timeout) != aws.Int64Value(obj.Configuration.Timeout) {
		return false, "", nil
	}

	// This should never be nil.  We set this in LateInit as aws will initialize a default value
	if aws.StringValue(cr.Spec.ForProvider.TracingConfig.Mode) != aws.StringValue(obj.Configuration.TracingConfig.Mode) {
		return false, "", nil
	}

	if !isUpToDateSecurityGroupIDs(cr, obj) {
		return false, "", nil
	}

	addTags, removeTags := tagutils.DiffTagsMapPtr(cr.Spec.ForProvider.Tags, obj.Tags)
	return len(addTags) == 0 && len(removeTags) == 0, "", nil

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

func equalPackageType(cr *svcapitypes.Function, obj *svcsdk.GetFunctionOutput) bool {
	if obj.Configuration == nil {
		return false
	}
	return aws.StringValue(cr.Spec.ForProvider.PackageType) == aws.StringValue(obj.Configuration.PackageType)
}

func isRepositoryType(obj *svcsdk.GetFunctionOutput, repositoryType string) bool {
	if obj.Code == nil {
		return false
	}
	return aws.StringValue(obj.Code.RepositoryType) == repositoryType
}

func equalImageURI(cr *svcapitypes.Function, obj *svcsdk.GetFunctionOutput) bool {
	if obj.Code == nil {
		return false
	}
	return aws.StringValue(cr.Spec.ForProvider.CustomFunctionCodeParameters.ImageURI) == aws.StringValue(obj.Code.ImageUri)
}

// isUpToDateCodeImage checks if FunctionConfiguration FunctionCodeLocation (Image) is up-to-date
// Returns true when function code is supplied via Zip file
func isUpToDateCodeImage(cr *svcapitypes.Function, obj *svcsdk.GetFunctionOutput) bool {
	desired := cr
	actual := obj

	if aws.StringValue(desired.Spec.ForProvider.PackageType) == packageTypeZip {
		return true
	}
	if !equalPackageType(desired, actual) {
		return false
	}
	if !isRepositoryType(actual, repositoryTypeECR) {
		return false
	}
	return equalImageURI(desired, actual)
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
	if cr.Spec.ForProvider.CustomFunctionVPCConfigParameters != nil &&
		cr.Spec.ForProvider.CustomFunctionVPCConfigParameters.SecurityGroupIDs != nil {
		securityGroupIDs = cr.Spec.ForProvider.CustomFunctionVPCConfigParameters.SecurityGroupIDs
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

type updater struct {
	client svcsdkapi.LambdaAPI
}

func (u *updater) isLastUpdateStatusSuccessful(ctx context.Context, cr *svcapitypes.Function) error {
	// LastUpdateStatus must be Successful before running UpdateFunction*
	// https://docs.aws.amazon.com/lambda/latest/dg/functions-states.html
	// https://aws.amazon.com/blogs/compute/coming-soon-expansion-of-aws-lambda-states-to-all-functions/

	for {
		out, err := u.client.GetFunctionWithContext(ctx, &svcsdk.GetFunctionInput{
			FunctionName: aws.String(meta.GetExternalName(cr)),
		})
		if err != nil {
			return err
		}
		if aws.StringValue(out.Configuration.LastUpdateStatus) == svcsdk.LastUpdateStatusSuccessful {
			return nil
		}
		time.Sleep(isLastUpdateStatusSuccessfulCheckInterval)
	}
}

//nolint:gocyclo
func (u *updater) update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*svcapitypes.Function)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	// LastUpdateStatus must be Successful before running UpdateFunctionCode
	if err := u.isLastUpdateStatusSuccessful(ctx, cr); err != nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
	}

	// https://docs.aws.amazon.com/sdk-for-go/api/service/lambda/#Lambda.UpdateFunctionCode
	updateFunctionCodeInput := GenerateUpdateFunctionCodeInput(cr)
	if _, err := u.client.UpdateFunctionCodeWithContext(ctx, updateFunctionCodeInput); err != nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
	}

	// LastUpdateStatus must be Successful before running UpdateFunctionConfiguration
	if err := u.isLastUpdateStatusSuccessful(ctx, cr); err != nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
	}

	updateFunctionConfigurationInput := GenerateUpdateFunctionConfigurationInput(cr)
	if _, err := u.client.UpdateFunctionConfigurationWithContext(ctx, updateFunctionConfigurationInput); err != nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
	}

	// Should store the ARN somewhere else?
	functionConfiguration, err := u.client.GetFunctionConfigurationWithContext(ctx, &svcsdk.GetFunctionConfigurationInput{
		FunctionName: aws.String(meta.GetExternalName(cr)),
	})
	if err != nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
	}

	// Tags
	tags, err := u.client.ListTagsWithContext(ctx, &svcsdk.ListTagsInput{
		Resource: functionConfiguration.FunctionArn,
	})
	if err != nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
	}

	addTags, removeTags := tagutils.DiffTagsMapPtr(cr.Spec.ForProvider.Tags, tags.Tags)
	// Remove old tags before adding new tags in case values change for keys
	if len(removeTags) > 0 {
		if _, err := u.client.UntagResourceWithContext(ctx, &svcsdk.UntagResourceInput{
			Resource: functionConfiguration.FunctionArn,
			TagKeys:  removeTags,
		}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
		}
	}

	if len(addTags) > 0 {
		if _, err := u.client.TagResourceWithContext(ctx, &svcsdk.TagResourceInput{
			Resource: functionConfiguration.FunctionArn,
			Tags:     addTags,
		}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
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
	if cr.Spec.ForProvider.CustomFunctionCodeParameters.ImageURI != nil {
		f0.SetImageUri(*cr.Spec.ForProvider.CustomFunctionCodeParameters.ImageURI)
	}
	if cr.Spec.ForProvider.CustomFunctionCodeParameters.S3Bucket != nil {
		f0.SetS3Bucket(*cr.Spec.ForProvider.CustomFunctionCodeParameters.S3Bucket)
	}
	if cr.Spec.ForProvider.CustomFunctionCodeParameters.S3Key != nil {
		f0.SetS3Key(*cr.Spec.ForProvider.CustomFunctionCodeParameters.S3Key)
	}
	if cr.Spec.ForProvider.CustomFunctionCodeParameters.S3ObjectVersion != nil {
		f0.SetS3ObjectVersion(*cr.Spec.ForProvider.CustomFunctionCodeParameters.S3ObjectVersion)
	}
	return f0
}

// GenerateUpdateFunctionConfigurationInput is similar to GenerateCreateFunctionConfigurationInput
// Copied almost verbatim from the zz_conversions generated code
//
//nolint:gocyclo
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
		f4f0 := map[string]*string{}
		for f4f0key, f4f0valiter := range cr.Spec.ForProvider.Environment.Variables {
			var f4f0val = *f4f0valiter
			f4f0[f4f0key] = &f4f0val
		}
		f4.SetVariables(f4f0)
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
	if cr.Spec.ForProvider.CustomFunctionVPCConfigParameters != nil {
		f19 := &svcsdk.VpcConfig{}

		f19f0 := []*string{}
		for _, f19f0iter := range cr.Spec.ForProvider.CustomFunctionVPCConfigParameters.SecurityGroupIDs {
			var f19f0elem = *f19f0iter
			f19f0 = append(f19f0, &f19f0elem)
		}
		f19.SetSecurityGroupIds(f19f0)

		f19f1 := []*string{}
		for _, f19f1iter := range cr.Spec.ForProvider.CustomFunctionVPCConfigParameters.SubnetIDs {
			var f19f1elem = *f19f1iter
			f19f1 = append(f19f1, &f19f1elem)
		}
		f19.SetSubnetIds(f19f1)

		res.SetVpcConfig(f19)
	}
	return res
}

func generateFuntionObservation(resp *svcsdk.GetFunctionOutput) svcapitypes.FunctionObservation {
	if resp == nil || resp.Configuration == nil {
		return svcapitypes.FunctionObservation{}
	}

	o := svcapitypes.FunctionObservation{
		CodeSHA256:                 resp.Configuration.CodeSha256,
		CodeSize:                   resp.Configuration.CodeSize,
		FunctionARN:                resp.Configuration.FunctionArn,
		FunctionName:               resp.Configuration.FunctionName,
		LastModified:               resp.Configuration.LastModified,
		LastUpdateStatus:           resp.Configuration.LastUpdateStatus,
		LastUpdateStatusReason:     resp.Configuration.LastUpdateStatusReason,
		LastUpdateStatusReasonCode: resp.Configuration.LastUpdateStatusReasonCode,
		MasterARN:                  resp.Configuration.MasterArn,
		RevisionID:                 resp.Configuration.RevisionId,
		Role:                       resp.Configuration.Role,
		SigningJobARN:              resp.Configuration.SigningJobArn,
		SigningProfileVersionARN:   resp.Configuration.SigningProfileVersionArn,
		State:                      resp.Configuration.State,
		StateReason:                resp.Configuration.StateReason,
		StateReasonCode:            resp.Configuration.StateReasonCode,
		Version:                    resp.Configuration.Version,
	}
	if resp.Configuration.VpcConfig != nil {
		o.VPCConfig = &svcapitypes.VPCConfigResponse{
			SecurityGroupIDs: resp.Configuration.VpcConfig.SecurityGroupIds,
			SubnetIDs:        resp.Configuration.VpcConfig.SubnetIds,
			VPCID:            resp.Configuration.VpcConfig.VpcId,
		}
	}
	if resp.Configuration.ImageConfigResponse != nil {
		o.ImageConfigResponse = &svcapitypes.ImageConfigResponse{}
		if resp.Configuration.ImageConfigResponse.Error != nil {
			o.ImageConfigResponse.Error = &svcapitypes.ImageConfigError{
				ErrorCode: resp.Configuration.Environment.Error.ErrorCode,
				Message:   resp.Configuration.Environment.Error.Message,
			}
		}
		if resp.Configuration.ImageConfigResponse.ImageConfig != nil {
			o.ImageConfigResponse.ImageConfig = &svcapitypes.ImageConfig{
				Command:          resp.Configuration.ImageConfigResponse.ImageConfig.Command,
				EntryPoint:       resp.Configuration.ImageConfigResponse.ImageConfig.EntryPoint,
				WorkingDirectory: resp.Configuration.ImageConfigResponse.ImageConfig.WorkingDirectory,
			}
		}
	}
	return o
}
