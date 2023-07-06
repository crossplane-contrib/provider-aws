package filesystem

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/efs"
	"github.com/aws/aws-sdk-go/service/efs/efsiface"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/efs/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	awsclients "github.com/crossplane-contrib/provider-aws/pkg/clients"
	svcutils "github.com/crossplane-contrib/provider-aws/pkg/controller/efs"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
)

const (
	errNotFileSystem    = "managed resource is not a FileSystem custom resource"
	errKubeUpdateFailed = "cannot update EFS FileSystem custom resource"
)

// SetupFileSystem adds a controller that reconciles FileSystem.
func SetupFileSystem(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.FileSystemGroupKind)
	opts := []option{
		func(e *external) {
			h := &hooks{client: e.client}
			e.isUpToDate = h.isUpToDate
			e.preCreate = preCreate
			e.postCreate = postCreate
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preUpdate = preUpdate
			e.postUpdate = h.postUpdate
			e.preDelete = preDelete
			e.lateInitialize = lateInitialize
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.FileSystem{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.FileSystemGroupVersionKind),
			managed.WithInitializers(&tagger{kube: mgr.GetClient()}),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

type hooks struct {
	client efsiface.EFSAPI
}

func (e *hooks) isUpToDate(cr *svcapitypes.FileSystem, obj *svcsdk.DescribeFileSystemsOutput) (bool, error) {
	in := cr.Spec.ForProvider
	out := obj.FileSystems[0]

	// to avoid AWS 409-error ("IncorrectFileSystemLifeCycleState" "File system ... is already being updated. Try again later.")
	if awsclients.StringValue(obj.FileSystems[0].LifeCycleState) == string(svcapitypes.LifeCycleState_updating) {
		return true, nil
	}
	switch {
	case awsclients.StringValue(in.ThroughputMode) != awsclients.StringValue(out.ThroughputMode),
		awsclients.Int64Value(in.ProvisionedThroughputInMibps) != int64(aws.Float64Value(out.ProvisionedThroughputInMibps)):
		return false, nil
	}

	return svcutils.AreTagsUpToDate(e.client, cr.Spec.ForProvider.Tags, cr.Status.AtProvider.FileSystemID)
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

	switch awsclients.StringValue(obj.FileSystems[0].LifeCycleState) {
	case string(svcapitypes.LifeCycleState_available):
		cr.Status.SetConditions(xpv1.Available())
	case string(svcapitypes.LifeCycleState_creating):
		cr.Status.SetConditions(xpv1.Creating())
	case string(svcapitypes.LifeCycleState_deleting):
		cr.Status.SetConditions(xpv1.Deleting())
	default:
		cr.Status.SetConditions(xpv1.Unavailable())
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

func (e *hooks) postUpdate(_ context.Context, cr *svcapitypes.FileSystem, resp *svcsdk.UpdateFileSystemOutput, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	// Ignore a specific error AWS throws, when there were no actual changes in the UpdateFileSystem-Request
	// to allow case where we only want to update the tags
	// entire message of the error to ignore: "BadRequest: The file system won't be updated. The requested throughput mode or provisioned throughput value are the same as the current mode and value."
	if err != nil && !strings.Contains(err.Error(), string("the same as the current")) {
		return upd, err
	}

	return upd, svcutils.UpdateTagsForResource(e.client, cr.Spec.ForProvider.Tags, cr.Status.AtProvider.FileSystemID)
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
	return managed.ExternalCreation{}, nil
}

func lateInitialize(spec *svcapitypes.FileSystemParameters, current *svcsdk.DescribeFileSystemsOutput) error {
	obj := current.FileSystems[0]
	spec.ThroughputMode = awsclients.LateInitializeStringPtr(spec.ThroughputMode, obj.ThroughputMode)
	return nil
}

type tagger struct {
	kube client.Client
}

func (t *tagger) Initialize(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*svcapitypes.FileSystem)
	if !ok {
		return errors.New(errNotFileSystem)
	}

	cr.Spec.ForProvider.Tags = svcutils.AddExternalTags(mg, cr.Spec.ForProvider.Tags)
	return errors.Wrap(t.kube.Update(ctx, cr), errKubeUpdateFailed)
}
