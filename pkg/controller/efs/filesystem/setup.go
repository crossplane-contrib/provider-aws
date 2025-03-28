package filesystem

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	svcsdk "github.com/aws/aws-sdk-go/service/efs"
	svcsdkapi "github.com/aws/aws-sdk-go/service/efs/efsiface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/efs/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupFileSystem adds a controller that reconciles FileSystem.
func SetupFileSystem(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.FileSystemGroupKind)
	opts := []option{
		func(e *external) {
			c := &custom{client: e.client, external: e}
			e.isUpToDate = isUpToDate
			e.preCreate = preCreate
			e.postCreate = postCreate
			e.preObserve = preObserve
			e.preUpdate = preUpdate
			e.preDelete = preDelete
			e.postObserve = c.postObserve
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithInitializers(),
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithTypedExternalConnector(&connector{kube: mgr.GetClient(), opts: opts}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(svcapitypes.FileSystemGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.FileSystem{}).
		Complete(r)
}

type custom struct {
	client   svcsdkapi.EFSAPI
	external *external
}

func isUpToDate(_ context.Context, cr *svcapitypes.FileSystem, obj *svcsdk.DescribeFileSystemsOutput) (bool, string, error) {
	for _, res := range obj.FileSystems {
		if pointer.Int64Value(cr.Spec.ForProvider.ProvisionedThroughputInMibps) != int64(aws.Float64Value(res.ProvisionedThroughputInMibps)) {
			return false, "", nil
		}
		if !ptr.Equal(cr.Spec.ForProvider.ThroughputMode, res.ThroughputMode) {
			return false, "", nil
		}

	}
	return true, "", nil
}

func preObserve(_ context.Context, cr *svcapitypes.FileSystem, obj *svcsdk.DescribeFileSystemsInput) error {
	// Describe query doesn't allow both CreationToken and FileSystemId to be given.
	obj.CreationToken = nil
	obj.FileSystemId = ptr.To(meta.GetExternalName(cr))
	return nil
}

func (e *custom) postObserve(ctx context.Context, cr *svcapitypes.FileSystem, obj *svcsdk.DescribeFileSystemsOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	if pointer.StringValue(obj.FileSystems[0].LifeCycleState) == string(svcapitypes.LifeCycleState_available) {
		cr.SetConditions(xpv1.Available())
	}
	obs.ConnectionDetails = managed.ConnectionDetails{
		svcapitypes.ResourceCredentialsSecretIDKey: []byte(meta.GetExternalName(cr)),
	}

	// BackupPolicy is managed separately from the FileSystem, so we update it here to avoid empty requests in preUpdate
	if obs.ResourceExists {
		if err := e.updateBackupPolicy(ctx, cr, obj); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, "cannot update backup policy")
		}
	}
	return obs, nil
}

func preUpdate(_ context.Context, cr *svcapitypes.FileSystem, obj *svcsdk.UpdateFileSystemInput) error {
	obj.FileSystemId = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	// Type of this field is *float64 but in practice, only integer values are allowed.
	if cr.Spec.ForProvider.ProvisionedThroughputInMibps != nil {
		obj.ProvisionedThroughputInMibps = aws.Float64(float64(pointer.Int64Value(cr.Spec.ForProvider.ProvisionedThroughputInMibps)))
	}
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.FileSystem, obj *svcsdk.DeleteFileSystemInput) (bool, error) {
	obj.FileSystemId = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return false, nil
}

func preCreate(_ context.Context, cr *svcapitypes.FileSystem, obj *svcsdk.CreateFileSystemInput) error {
	obj.CreationToken = pointer.ToOrNilIfZeroValue(string(cr.UID))
	// Type of this field is *float64 but in practice, only integer values are allowed.
	if cr.Spec.ForProvider.ProvisionedThroughputInMibps != nil {
		obj.ProvisionedThroughputInMibps = aws.Float64(float64(pointer.Int64Value(cr.Spec.ForProvider.ProvisionedThroughputInMibps)))
	}
	return nil
}

func postCreate(_ context.Context, cr *svcapitypes.FileSystem, obj *svcsdk.FileSystemDescription, _ managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, pointer.StringValue(obj.FileSystemId))
	return managed.ExternalCreation{}, nil
}

func (e *custom) getCurrentBackupPolicy(filesystemId *string) (bool, error) {
	// Get BackupPolicy
	// default is no backup
	var currentBackupPolicyEnabled bool
	backupPolicyOutput, err := e.client.DescribeBackupPolicy(&svcsdk.DescribeBackupPolicyInput{FileSystemId: filesystemId})
	if err != nil {
		currentBackupPolicyEnabled = !backupPolicyIsNotFound(err) // if backup policy is not found, there is no backup and thus false
		return currentBackupPolicyEnabled, resource.Ignore(backupPolicyIsNotFound, err)
	}
	policyStatus := aws.StringValue(backupPolicyOutput.BackupPolicy.Status)
	switch policyStatus {
	case "ENABLED", "ENABLING":
		currentBackupPolicyEnabled = true
	case "DISABLED", "DISABLING":
		currentBackupPolicyEnabled = false
	}
	return currentBackupPolicyEnabled, nil
}

func (e *custom) updateBackupPolicy(ctx context.Context, cr *svcapitypes.FileSystem, obj *svcsdk.DescribeFileSystemsOutput) error {
	for _, res := range obj.FileSystems {
		currentBackupPolicyEnabled, err := e.getCurrentBackupPolicy(res.FileSystemId)
		if err != nil {
			return errors.Wrap(err, "failed to get backup policy")
		}

		if currentBackupPolicyEnabled != ptr.Deref(cr.Spec.ForProvider.Backup, false) {
			var policy string
			if ptr.Deref(cr.Spec.ForProvider.Backup, false) {
				policy = "ENABLED"
			} else {
				policy = "DISABLED"
			}
			_, err := e.client.PutBackupPolicyWithContext(ctx,
				&svcsdk.PutBackupPolicyInput{
					FileSystemId: res.FileSystemId,
					BackupPolicy: &svcsdk.BackupPolicy{Status: &policy},
				})

			if err != nil {
				return errors.Wrap(err, "failed to put backup policy")
			}
		}
	}
	return nil
}

func backupPolicyIsNotFound(err error) bool {
	// default BackupPolicy is false, so if backupPolicy is not found, it's false
	var awsErr awserr.Error
	ok := errors.As(err, &awsErr)
	return ok && awsErr.Code() == "PolicyNotFound"
}
