package virtualcluster

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/emrcontainers"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	cpresource "github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/emrcontainers/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/tags"
)

const (
	firstObserveID    = "invalid"
	terminatedMessage = "VirtualCluster is already terminated"
	errListTag        = "cannot list tags"
	errUntag          = "cannot remove tags"
	errTag            = "cannot add tags"
)

// SetupVirtualCluster adds a controller that reconciles VirtualCluster.
func SetupVirtualCluster(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.VirtualClusterGroupKind)
	opts := []option{
		func(e *external) {
			e.isUpToDate = isUpToDate
			e.preCreate = preCreate
			e.preObserve = preObserve
			e.postCreate = postCreate
			e.postDelete = postDelete
			e.postObserve = postObserve
			e.update = e.updater
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
		resource.ManagedKind(svcapitypes.VirtualClusterGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.VirtualCluster{}).
		Complete(r)
}

func preObserve(ctx context.Context, cr *svcapitypes.VirtualCluster, input *svcsdk.DescribeVirtualClusterInput) error {
	externalName := meta.GetExternalName(cr)
	if externalName == cr.Name {
		input.Id = aws.String(firstObserveID)
	} else {
		input.Id = aws.String(externalName)
	}
	return nil
}

func postObserve(ctx context.Context, cr *svcapitypes.VirtualCluster, resp *svcsdk.DescribeVirtualClusterOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return obs, err
	}

	switch state := *resp.VirtualCluster.State; state {
	case svcsdk.VirtualClusterStateRunning:
		cr.SetConditions(xpv1.Available())
	case svcsdk.VirtualClusterStateTerminating:
		cr.SetConditions(xpv1.Deleting())
	// deleted virtual clusters are in terminated state but not removed from API responses
	case svcsdk.VirtualClusterStateTerminated:
		obs.ResourceExists = false
	case svcsdk.VirtualClusterStateArrested:
		cr.SetConditions(xpv1.Unavailable())
	}

	return obs, err
}

func preCreate(ctx context.Context, cr *svcapitypes.VirtualCluster, input *svcsdk.CreateVirtualClusterInput) error {
	input.Name = &cr.Name
	return nil
}

func postCreate(ctx context.Context, cr *svcapitypes.VirtualCluster, resp *svcsdk.CreateVirtualClusterOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return cre, err
	}
	meta.SetExternalName(cr, aws.StringValue(resp.Id))
	return cre, nil
}

func postDelete(ctx context.Context, cr *svcapitypes.VirtualCluster, resp *svcsdk.DeleteVirtualClusterOutput, err error) error {
	if err == nil {
		return nil
	}
	// error context is stripped. cannot type assert.
	cause := errors.Cause(err)
	if cause.Error() == fmt.Sprintf("%s: %s", svcsdk.ErrCodeValidationException, terminatedMessage) {
		return nil
	}
	return err
}

func isUpToDate(_ context.Context, cr *svcapitypes.VirtualCluster, output *svcsdk.DescribeVirtualClusterOutput) (bool, string, error) {
	add, remove := tags.DiffTagsMapPtr(cr.Spec.ForProvider.Tags, output.VirtualCluster.Tags)
	return len(add) == 0 && len(remove) == 0, "", nil
}

func (e *external) updater(ctx context.Context, mg cpresource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*svcapitypes.VirtualCluster)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}
	return managed.ExternalUpdate{}, e.updateTags(ctx, cr)
}

func (e *external) updateTags(ctx context.Context, cr *svcapitypes.VirtualCluster) error {
	resp, err := e.client.ListTagsForResourceWithContext(ctx, &svcsdk.ListTagsForResourceInput{
		ResourceArn: cr.Status.AtProvider.ARN,
	})
	if err != nil {
		return errors.Wrap(err, errListTag)
	}

	add, remove := tags.DiffTagsMapPtr(cr.Spec.ForProvider.Tags, resp.Tags)
	if len(remove) > 0 {
		_, err = e.client.UntagResourceWithContext(ctx, &svcsdk.UntagResourceInput{
			ResourceArn: cr.Status.AtProvider.ARN,
			TagKeys:     remove,
		})
		if err != nil {
			return errors.Wrap(err, errUntag)
		}
	}
	if len(add) > 0 {
		_, err = e.client.TagResourceWithContext(ctx, &svcsdk.TagResourceInput{
			ResourceArn: cr.Status.AtProvider.ARN,
			Tags:        add,
		})
		if err != nil {
			return errors.Wrap(err, errTag)
		}
	}
	return nil
}
