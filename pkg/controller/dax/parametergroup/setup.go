package parametergroup

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

	svcapitypes "github.com/crossplane/provider-aws/apis/dax/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// SetupParameterGroup adds a controller that reconciles ParameterGroup.
func SetupParameterGroup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.ParameterGroupGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preCreate = preCreate
			e.postCreate = postCreate
			e.preUpdate = preUpdate
			e.preDelete = preDelete
			e.isUpToDate = isUpToDate
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&svcapitypes.ParameterGroup{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.ParameterGroupGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithInitializers(),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func preObserve(_ context.Context, cr *svcapitypes.ParameterGroup, obj *svcsdk.DescribeParameterGroupsInput) error {
	obj.ParameterGroupNames = append(obj.ParameterGroupNames, awsclients.String(meta.GetExternalName(cr)))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.ParameterGroup, _ *svcsdk.DescribeParameterGroupsOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(xpv1.Available())
	return obs, nil
}

func preCreate(_ context.Context, cr *svcapitypes.ParameterGroup, obj *svcsdk.CreateParameterGroupInput) error {
	obj.ParameterGroupName = awsclients.String(cr.Name)
	return nil
}

func postCreate(_ context.Context, cr *svcapitypes.ParameterGroup, obj *svcsdk.CreateParameterGroupOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, awsclients.StringValue(obj.ParameterGroup.ParameterGroupName))
	return cre, nil
}

func preUpdate(_ context.Context, cr *svcapitypes.ParameterGroup, obj *svcsdk.UpdateParameterGroupInput) error {
	obj.ParameterGroupName = awsclients.String(meta.GetExternalName(cr))
	obj.ParameterNameValues = make([]*svcsdk.ParameterNameValue, len(cr.Spec.ForProvider.ParameterNameValues))

	for i, v := range cr.Spec.ForProvider.ParameterNameValues {
		obj.ParameterNameValues[i] = &svcsdk.ParameterNameValue{
			ParameterName:  v.ParameterName,
			ParameterValue: v.ParameterValue,
		}
	}
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.ParameterGroup, obj *svcsdk.DeleteParameterGroupInput) (bool, error) {
	obj.ParameterGroupName = awsclients.String(meta.GetExternalName(cr))
	return false, nil
}

func isUpToDate(cr *svcapitypes.ParameterGroup, output *svcsdk.DescribeParameterGroupsOutput) (bool, error) {
	in := cr.Spec.ForProvider
	out := output.ParameterGroups[0]

	if !cmp.Equal(in.Description, out.Description) {
		return false, nil
	}

	return true, nil
}
