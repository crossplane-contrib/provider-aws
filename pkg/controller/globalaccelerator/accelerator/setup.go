package accelerator

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/globalaccelerator"
	svcsdkapi "github.com/aws/aws-sdk-go/service/globalaccelerator/globalacceleratoriface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/globalaccelerator/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupAccelerator adds a controller that reconciles an Accelerator.
func SetupAccelerator(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.AcceleratorGroupKind)
	opts := []option{
		func(e *external) {
			e.isUpToDate = isUpToDate
			c := &gaClient{client: e.client}
			e.preDelete = c.preDelete
			e.preCreate = preCreate
			e.postCreate = postCreate
			e.postObserve = postObserve
			e.preObserve = preObserve
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
		For(&svcapitypes.Accelerator{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.AcceleratorGroupVersionKind),
			managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithInitializers(),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

type gaClient struct {
	client svcsdkapi.GlobalAcceleratorAPI
}

func preObserve(ctx context.Context, cr *svcapitypes.Accelerator, obj *svcsdk.DescribeAcceleratorInput) error {
	obj.AcceleratorArn = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

func preCreate(_ context.Context, cr *svcapitypes.Accelerator, obj *svcsdk.CreateAcceleratorInput) error {
	obj.Name = cr.Spec.ForProvider.Name
	obj.IdempotencyToken = pointer.ToOrNilIfZeroValue(string(cr.UID))
	return nil
}

func (d gaClient) preDelete(ctx context.Context, cr *svcapitypes.Accelerator, obj *svcsdk.DeleteAcceleratorInput) (bool, error) {
	accArn := meta.GetExternalName(cr)
	obj.AcceleratorArn = pointer.ToOrNilIfZeroValue(accArn)

	// we need to check first if the accelerator is already disabled on remote
	// because sending an update request will bring it into pending state and
	// sending delete requests against an accelerator in pending state will result
	// in a AcceleratorNotDisabledException
	descReq := &svcsdk.DescribeAcceleratorInput{
		AcceleratorArn: pointer.ToOrNilIfZeroValue(accArn),
	}

	descResp, err := d.client.DescribeAccelerator(descReq)
	if err != nil {
		return false, err
	}

	if ptr.Deref(descResp.Accelerator.Enabled, true) && ptr.Deref(descResp.Accelerator.Status, "") != svcsdk.AcceleratorStatusInProgress {
		enabled := false
		updReq := &svcsdk.UpdateAcceleratorInput{
			Enabled:        &enabled,
			AcceleratorArn: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
			Name:           cr.Spec.ForProvider.Name,
			IpAddressType:  cr.Spec.ForProvider.IPAddressType,
		}

		_, err := d.client.UpdateAcceleratorWithContext(ctx, updReq)
		if err != nil {
			return false, err
		}
	}

	return false, nil
}

func postObserve(_ context.Context, cr *svcapitypes.Accelerator, resp *svcsdk.DescribeAcceleratorOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	switch aws.StringValue(resp.Accelerator.Status) {
	case string(svcapitypes.AcceleratorStatus_SDK_DEPLOYED):
		cr.SetConditions(xpv1.Available())
	case string(svcapitypes.AcceleratorStatus_SDK_IN_PROGRESS):
		cr.SetConditions(xpv1.Creating())
	default:
		cr.SetConditions(xpv1.Unavailable())
	}

	return obs, nil
}

func postCreate(_ context.Context, cr *svcapitypes.Accelerator, resp *svcsdk.CreateAcceleratorOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	meta.SetExternalName(cr, pointer.StringValue(resp.Accelerator.AcceleratorArn))
	return cre, nil
}

func isUpToDate(_ context.Context, cr *svcapitypes.Accelerator, resp *svcsdk.DescribeAcceleratorOutput) (bool, string, error) {
	if ptr.Deref(cr.Spec.ForProvider.Enabled, false) != ptr.Deref(resp.Accelerator.Enabled, false) {
		return false, "", nil
	}

	if ptr.Deref(cr.Spec.ForProvider.Name, "") != ptr.Deref(resp.Accelerator.Name, "") {
		return false, "", nil
	}

	return true, "", nil
}
