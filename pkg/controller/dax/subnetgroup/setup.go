package subnetgroup

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/dax"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/dax/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupSubnetGroup adds a controller that reconciles SubnetGroup.
func SetupSubnetGroup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.SubnetGroupGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preCreate = preCreate
			e.preUpdate = preUpdate
			e.preDelete = preDelete
			e.isUpToDate = isUpToDate
		},
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
		managed.WithInitializers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(svcapitypes.SubnetGroupGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.SubnetGroup{}).
		Complete(r)
}

func preObserve(_ context.Context, cr *svcapitypes.SubnetGroup, obj *svcsdk.DescribeSubnetGroupsInput) error {
	obj.SubnetGroupNames = append(obj.SubnetGroupNames, pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.SubnetGroup, _ *svcsdk.DescribeSubnetGroupsOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(xpv1.Available())
	return obs, nil
}

func preCreate(_ context.Context, cr *svcapitypes.SubnetGroup, obj *svcsdk.CreateSubnetGroupInput) error {
	meta.SetExternalName(cr, cr.Name)
	obj.SubnetGroupName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	for _, s := range cr.Spec.ForProvider.SubnetIds {
		obj.SubnetIds = append(obj.SubnetIds, pointer.ToOrNilIfZeroValue(*s))
	}
	return nil
}

func preUpdate(_ context.Context, cr *svcapitypes.SubnetGroup, obj *svcsdk.UpdateSubnetGroupInput) error {
	obj.SubnetGroupName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	for _, s := range cr.Spec.ForProvider.SubnetIds {
		obj.SubnetIds = append(obj.SubnetIds, pointer.ToOrNilIfZeroValue(*s))
	}
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.SubnetGroup, obj *svcsdk.DeleteSubnetGroupInput) (bool, error) {
	obj.SubnetGroupName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return false, nil
}

func isUpToDate(_ context.Context, cr *svcapitypes.SubnetGroup, output *svcsdk.DescribeSubnetGroupsOutput) (bool, string, error) {
	in := cr.Spec.ForProvider
	out := output.SubnetGroups[0]

	if !cmp.Equal(in.Description, out.Description) {
		return false, "", nil
	}

	subnetsOut := make([]*string, len(out.Subnets))
	for i, subnet := range out.Subnets {
		subnetsOut[i] = subnet.SubnetIdentifier
	}

	if !cmp.Equal(in.SubnetIds, subnetsOut) {
		return false, "", nil
	}

	return true, "", nil
}
