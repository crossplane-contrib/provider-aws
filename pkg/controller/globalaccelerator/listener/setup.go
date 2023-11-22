package listener

import (
	"context"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/globalaccelerator"
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
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupListener adds a controller that reconciles Listener.
func SetupListener(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.ListenerGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.preCreate = preCreate
			e.preUpdate = preUpdate
			e.postObserve = postObserve
			e.postCreate = postCreate
			e.isUpToDate = isUpToDate
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
		For(&svcapitypes.Listener{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.ListenerGroupVersionKind),
			managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithInitializers(),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

func preObserve(ctx context.Context, cr *svcapitypes.Listener, obj *svcsdk.DescribeListenerInput) error {
	obj.ListenerArn = aws.String(meta.GetExternalName(cr))
	return nil
}

func preCreate(_ context.Context, cr *svcapitypes.Listener, obj *svcsdk.CreateListenerInput) error {
	obj.AcceleratorArn = aws.String(ptr.Deref(cr.Spec.ForProvider.CustomListenerParameters.AcceleratorArn, ""))
	obj.IdempotencyToken = aws.String(string(cr.UID))
	return nil
}

func preUpdate(_ context.Context, cr *svcapitypes.Listener, obj *svcsdk.UpdateListenerInput) error {
	obj.ListenerArn = aws.String(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.Listener, resp *svcsdk.DescribeListenerOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(xpv1.Available())
	return obs, nil
}

func postCreate(_ context.Context, cr *svcapitypes.Listener, resp *svcsdk.CreateListenerOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	meta.SetExternalName(cr, aws.StringValue(resp.Listener.ListenerArn))
	return cre, err
}

func isUpToDate(_ context.Context, cr *svcapitypes.Listener, resp *svcsdk.DescribeListenerOutput) (bool, string, error) {
	// unequal amount of port-ranges, resource not up to date
	if len(cr.Spec.ForProvider.PortRanges) != len(resp.Listener.PortRanges) {
		return false, "", nil
	}
	portRangesCpy := svcapitypes.ListenerParameters{}

	// In order to compare the ports, we first sort both the cr and api response
	// We copy the cr parameters to not alter them
	cr.Spec.ForProvider.DeepCopyInto(&portRangesCpy)

	sort.Slice(portRangesCpy.PortRanges, func(i, j int) bool {
		if *portRangesCpy.PortRanges[i].FromPort != *portRangesCpy.PortRanges[j].FromPort {
			return *portRangesCpy.PortRanges[i].FromPort < *portRangesCpy.PortRanges[j].FromPort
		}
		return *portRangesCpy.PortRanges[i].ToPort < *portRangesCpy.PortRanges[j].ToPort
	})

	sort.Slice(resp.Listener.PortRanges, func(i, j int) bool {
		if *resp.Listener.PortRanges[i].FromPort != *resp.Listener.PortRanges[j].FromPort {
			return *resp.Listener.PortRanges[i].FromPort < *resp.Listener.PortRanges[j].FromPort
		}
		return *resp.Listener.PortRanges[i].ToPort < *resp.Listener.PortRanges[j].ToPort
	})

	for i := range portRangesCpy.PortRanges {
		if *resp.Listener.PortRanges[i].FromPort != *portRangesCpy.PortRanges[i].FromPort {
			return false, "", nil
		}
		if *resp.Listener.PortRanges[i].ToPort != *portRangesCpy.PortRanges[i].ToPort {
			return false, "", nil
		}
	}

	if ptr.Deref(cr.Spec.ForProvider.ClientAffinity, "") != *resp.Listener.ClientAffinity {
		return false, "", nil
	}

	if ptr.Deref(cr.Spec.ForProvider.Protocol, "") != *resp.Listener.Protocol {
		return false, "", nil
	}

	return true, "", nil
}
