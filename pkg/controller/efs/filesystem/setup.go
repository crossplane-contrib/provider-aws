package filesystem

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/efs"
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

	svcapitypes "github.com/crossplane/provider-aws/apis/efs/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// SetupFileSystem adds a controller that reconciles FileSystem.
func SetupFileSystem(mgr ctrl.Manager, l logging.Logger, limiter workqueue.RateLimiter) error {
	name := managed.ControllerName(svcapitypes.FileSystemGroupKind)
	opts := []option{
		func(e *external) {
			e.isUpToDate = isUpToDate
			e.preCreate = preCreate
			e.postCreate = postCreate
			e.preObserve = preObserve
			e.preUpdate = preUpdate
			e.preDelete = preDelete
			e.postObserve = postObserve
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(limiter),
		}).
		For(&svcapitypes.FileSystem{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.FileSystemGroupVersionKind),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func isUpToDate(cr *svcapitypes.FileSystem, obj *svcsdk.DescribeFileSystemsOutput) (bool, error) {
	for _, res := range obj.FileSystems {
		if awsclients.Int64Value(cr.Spec.ForProvider.ProvisionedThroughputInMibps) != int64(aws.Float64Value(res.ProvisionedThroughputInMibps)) {
			return false, nil
		}
	}
	return true, nil
}

func preObserve(_ context.Context, cr *svcapitypes.FileSystem, obj *svcsdk.DescribeFileSystemsInput) error {
	// Describe query doesn't allow both CreationToken and FileSystemId to be given.
	obj.CreationToken = nil
	obj.FileSystemId = awsclients.String(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.FileSystem, obj *svcsdk.DescribeFileSystemsOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	if awsclients.StringValue(obj.FileSystems[0].LifeCycleState) == string(svcapitypes.LifeCycleState_available) {
		cr.SetConditions(xpv1.Available())
	}
	obs.ConnectionDetails = managed.ConnectionDetails{
		svcapitypes.ResourceCredentialsSecretIDKey: []byte(meta.GetExternalName(cr)),
	}
	return obs, nil
}

func preUpdate(_ context.Context, cr *svcapitypes.FileSystem, obj *svcsdk.UpdateFileSystemInput) error {
	obj.FileSystemId = awsclients.String(meta.GetExternalName(cr))
	// Type of this field is *float64 but in practice, only integer values are allowed.
	if cr.Spec.ForProvider.ProvisionedThroughputInMibps != nil {
		obj.ProvisionedThroughputInMibps = aws.Float64(float64(awsclients.Int64Value(cr.Spec.ForProvider.ProvisionedThroughputInMibps)))
	}
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.FileSystem, obj *svcsdk.DeleteFileSystemInput) (bool, error) {
	obj.FileSystemId = awsclients.String(meta.GetExternalName(cr))
	return false, nil
}

func preCreate(_ context.Context, cr *svcapitypes.FileSystem, obj *svcsdk.CreateFileSystemInput) error {
	obj.CreationToken = awsclients.String(string(cr.UID))
	// Type of this field is *float64 but in practice, only integer values are allowed.
	if cr.Spec.ForProvider.ProvisionedThroughputInMibps != nil {
		obj.ProvisionedThroughputInMibps = aws.Float64(float64(awsclients.Int64Value(cr.Spec.ForProvider.ProvisionedThroughputInMibps)))
	}
	return nil
}

func postCreate(_ context.Context, cr *svcapitypes.FileSystem, obj *svcsdk.FileSystemDescription, _ managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, awsclients.StringValue(obj.FileSystemId))
	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}
