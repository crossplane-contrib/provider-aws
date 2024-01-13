package environment

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/mwaa"
	svcsdkapi "github.com/aws/aws-sdk-go/service/mwaa/mwaaiface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/mwaa/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errNotEnvironment   = "managed resource is not a environment custom resource"
	errKubeUpdateFailed = "failed to update Secret custom resource"
	errCreateCLIToken   = "cannot create CLI token"
	errCreateWebToken   = "cannot create web token"
	errGetEnvironemt    = "cannot get environment"
	errTagResource      = "cannot tag resource"
	errUntagResource    = "cannot untag resource"
)

// SetupEnvironment adds a controller that reconciles Environment.
func SetupEnvironment(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.EnvironmentGroupKind)
	opts := []option{
		func(e *external) {
			c := &custom{client: e.client, kube: e.kube, external: e}
			e.isUpToDate = isUpToDate
			e.preCreate = c.preCreate
			e.postCreate = c.postCreate
			e.preObserve = preObserve
			e.preDelete = preDelete
			e.preUpdate = c.preUpdate
			e.postUpdate = c.postUpdate
			e.postObserve = c.postObserve
			e.lateInitialize = lateInitialize
		},
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient())),
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(svcapitypes.EnvironmentGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.Environment{}).
		Complete(r)
}

type custom struct {
	kube     client.Client
	client   svcsdkapi.MWAAAPI
	external *external
}

func preObserve(_ context.Context, cr *svcapitypes.Environment, obj *svcsdk.GetEnvironmentInput) error {
	obj.Name = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

func (e *custom) postObserve(ctx context.Context, cr *svcapitypes.Environment, obj *svcsdk.GetEnvironmentOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil || obj.Environment == nil {
		return managed.ExternalObservation{}, err
	}

	// obj.Pending.PendingChange is nil if Environment is available
	if status := pointer.StringValue(obj.Environment.Status); status != string(svcapitypes.EnvironmentStatus_SDK_AVAILABLE) {
		switch status {
		case string(svcapitypes.EnvironmentStatus_SDK_CREATING), string(svcapitypes.EnvironmentStatus_SDK_UPDATING):
			cr.SetConditions(xpv1.Creating().WithMessage(status))
		case string(svcapitypes.EnvironmentStatus_SDK_DELETING):
			cr.SetConditions(xpv1.Deleting())
		default:
			cr.SetConditions(xpv1.Unavailable().WithMessage(status))
		}
		// Update call not possible during creation or udpate.
		obs.ResourceUpToDate = true
		return obs, nil
	}

	cr.SetConditions(xpv1.Available())
	obs.ConnectionDetails = managed.ConnectionDetails{
		svcapitypes.ConnectionDetailsWebServerURL: []byte(pointer.StringValue(obj.Environment.WebserverUrl)),
	}
	return obs, nil
}

func preDelete(_ context.Context, cr *svcapitypes.Environment, obj *svcsdk.DeleteEnvironmentInput) (bool, error) {
	// Only environments that are not PENDING can be deleted
	switch pointer.StringValue(cr.Status.AtProvider.Status) {
	case svcsdk.EnvironmentStatusCreateFailed,
		svcsdk.EnvironmentStatusAvailable:
		obj.Name = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
		return false, nil
	}

	return true, nil // Skip
}

func (e *custom) preCreate(ctx context.Context, cr *svcapitypes.Environment, obj *svcsdk.CreateEnvironmentInput) error {
	obj.Name = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	obj.SourceBucketArn = cr.Spec.ForProvider.SourceBucketARN
	obj.ExecutionRoleArn = cr.Spec.ForProvider.ExecutionRoleARN
	obj.KmsKey = cr.Spec.ForProvider.KMSKey
	obj.NetworkConfiguration = &svcsdk.NetworkConfiguration{
		SecurityGroupIds: pointer.SliceValueToPtr(cr.Spec.ForProvider.NetworkConfiguration.SecurityGroupIDs),
		SubnetIds:        pointer.SliceValueToPtr(cr.Spec.ForProvider.NetworkConfiguration.SubnetIDs),
	}
	return nil
}

func (e *custom) postCreate(ctx context.Context, cr *svcapitypes.Environment, obj *svcsdk.CreateEnvironmentOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return cre, err
	}

	cliTokenRes, err := e.client.CreateCliTokenWithContext(ctx, &svcsdk.CreateCliTokenInput{Name: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))})
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateCLIToken)
	}

	webTokenRes, err := e.client.CreateWebLoginTokenWithContext(ctx, &svcsdk.CreateWebLoginTokenInput{Name: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))})
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateWebToken)
	}

	return managed.ExternalCreation{
		ConnectionDetails: managed.ConnectionDetails{
			svcapitypes.ConnectionDetailsCLITokenKey: []byte(pointer.StringValue(cliTokenRes.CliToken)),
			svcapitypes.ConnectionDetailsWebTokenKey: []byte(pointer.StringValue(webTokenRes.WebToken)),
		},
	}, err
}

func (e *custom) preUpdate(ctx context.Context, cr *svcapitypes.Environment, obj *svcsdk.UpdateEnvironmentInput) error {
	obj.Name = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	obj.SourceBucketArn = cr.Spec.ForProvider.SourceBucketARN
	obj.ExecutionRoleArn = cr.Spec.ForProvider.ExecutionRoleARN
	obj.NetworkConfiguration = &svcsdk.UpdateNetworkConfigurationInput{
		SecurityGroupIds: pointer.SliceValueToPtr(cr.Spec.ForProvider.NetworkConfiguration.SecurityGroupIDs),
	}
	return nil
}

func (e *custom) postUpdate(ctx context.Context, cr *svcapitypes.Environment, obj *svcsdk.UpdateEnvironmentOutput, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	if err != nil {
		return upd, err
	}

	res, err := e.client.GetEnvironmentWithContext(ctx, &svcsdk.GetEnvironmentInput{
		Name: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
	})
	if err != nil || res.Environment == nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errGetEnvironemt)
	}

	add, remove := diffTags(cr.Spec.ForProvider.Tags, res.Environment.Tags)
	if len(add) > 0 {
		_, err := e.client.TagResourceWithContext(ctx, &svcsdk.TagResourceInput{
			Tags: add,
		})
		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errTagResource)
		}
	}
	if len(remove) > 0 {
		_, err := e.client.UntagResourceWithContext(ctx, &svcsdk.UntagResourceInput{
			TagKeys: remove,
		})
		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errUntagResource)
		}
	}
	return managed.ExternalUpdate{}, nil
}

func isUpToDate(_ context.Context, cr *svcapitypes.Environment, obj *svcsdk.GetEnvironmentOutput) (bool, string, error) {
	if obj.Environment == nil {
		return false, "", nil
	}

	env := generateEnvironment(obj)
	diff := cmp.Diff(
		cr.Spec.ForProvider,
		env.Spec.ForProvider,
		cmpopts.IgnoreTypes(&xpv1.Reference{}, &xpv1.Selector{}, []xpv1.Reference{}),
		cmpopts.IgnoreFields(svcapitypes.EnvironmentParameters{}, "Region"),
		cmpopts.IgnoreFields(svcapitypes.CustomEnvironmentParameters{}, "KMSKey"),
	)
	return diff == "", diff, nil
}

func lateInitialize(spec *svcapitypes.EnvironmentParameters, obj *svcsdk.GetEnvironmentOutput) error {
	current := generateEnvironment(obj).Spec.ForProvider

	if spec.AirflowConfigurationOptions == nil {
		spec.AirflowConfigurationOptions = current.AirflowConfigurationOptions
	}
	spec.AirflowVersion = pointer.LateInitialize(spec.AirflowVersion, current.AirflowVersion)
	spec.EnvironmentClass = pointer.LateInitialize(spec.EnvironmentClass, current.EnvironmentClass)
	if spec.LoggingConfiguration == nil {
		spec.LoggingConfiguration = current.LoggingConfiguration
	}
	spec.MaxWorkers = pointer.LateInitialize(spec.MaxWorkers, current.MaxWorkers)
	spec.MinWorkers = pointer.LateInitialize(spec.MinWorkers, current.MinWorkers)
	spec.PluginsS3ObjectVersion = pointer.LateInitialize(spec.PluginsS3ObjectVersion, current.PluginsS3ObjectVersion)
	spec.PluginsS3Path = pointer.LateInitialize(spec.PluginsS3Path, current.PluginsS3Path)
	spec.RequirementsS3ObjectVersion = pointer.LateInitialize(spec.RequirementsS3ObjectVersion, current.RequirementsS3ObjectVersion)
	spec.RequirementsS3Path = pointer.LateInitialize(spec.RequirementsS3Path, current.RequirementsS3Path)
	spec.Schedulers = pointer.LateInitialize(spec.Schedulers, current.Schedulers)
	spec.WebserverAccessMode = pointer.LateInitialize(spec.WebserverAccessMode, current.WebserverAccessMode)
	spec.WeeklyMaintenanceWindowStart = pointer.LateInitialize(spec.WeeklyMaintenanceWindowStart, current.WeeklyMaintenanceWindowStart)
	return nil
}

// DiffTags returns tags that should be added or removed.
func diffTags(spec, current map[string]*string) (map[string]*string, []*string) {
	addTags := map[string]*string{}
	remove := []*string{}

	for k, specVal := range spec {
		curVal, exists := current[k]
		if !exists || pointer.StringValue(curVal) != pointer.StringValue(specVal) {
			addTags[k] = specVal
		}
	}

	for k := range current {
		if _, exists := spec[k]; !exists {
			remove = append(remove, pointer.ToOrNilIfZeroValue(k))
		}
	}

	return addTags, remove
}
