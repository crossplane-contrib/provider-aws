package broker

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcsdk "github.com/aws/aws-sdk-go/service/mq"
	svcsdkapi "github.com/aws/aws-sdk-go/service/mq/mqiface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"

	svcapitypes "github.com/crossplane/provider-aws/apis/mq/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/mq"
)

// SetupBroker adds a controller that reconciles Broker.
func SetupBroker(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.BrokerGroupKind)
	opts := []option{
		func(e *external) {
			c := &custom{client: e.client, kube: e.kube, external: e}
			e.preCreate = c.preCreate
			e.postCreate = postCreate
			e.preObserve = preObserve
			e.preDelete = preDelete
			e.postObserve = c.postObserve
			e.lateInitialize = LateInitialize
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&svcapitypes.Broker{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.BrokerGroupVersionKind),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type custom struct {
	kube     client.Client
	client   svcsdkapi.MQAPI
	external *external
}

func preObserve(_ context.Context, cr *svcapitypes.Broker, obj *svcsdk.DescribeBrokerInput) error {
	obj.BrokerId = awsclients.String(meta.GetExternalName(cr))
	return nil
}

func (e *custom) postObserve(ctx context.Context, cr *svcapitypes.Broker, obj *svcsdk.DescribeBrokerResponse, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	switch awsclients.StringValue(obj.BrokerState) {
	case string(svcapitypes.BrokerState_RUNNING):
		cr.SetConditions(xpv1.Available())
	case string(svcapitypes.BrokerState_CREATION_IN_PROGRESS):
		cr.SetConditions(xpv1.Creating())
	case string(svcapitypes.BrokerState_REBOOT_IN_PROGRESS):
		cr.SetConditions(xpv1.Unavailable())
	case string(svcapitypes.BrokerState_DELETION_IN_PROGRESS):
		cr.SetConditions(xpv1.Deleting())
	case string(svcapitypes.BrokerState_CREATION_FAILED):
		cr.SetConditions(xpv1.Unavailable())
	}

	pw, _, err := mq.GetPassword(ctx, e.kube, &cr.Spec.ForProvider.CustomUsers[0].PasswordSecretRef, cr.Spec.WriteConnectionSecretToReference)
	if resource.IgnoreNotFound(err) != nil || pw == "" {
		return obs, errors.Wrap(err, "cannot get password from the given secret")
	}

	obs.ConnectionDetails = managed.ConnectionDetails{
		"BrokerID": []byte(awsclients.StringValue(cr.Status.AtProvider.BrokerID)),
		"Region":   []byte(awsclients.StringValue(&cr.Spec.ForProvider.Region)),
		"Username": []byte(awsclients.StringValue(cr.Spec.ForProvider.CustomUsers[0].Username)),
		"Password": []byte(pw),
	}

	return obs, nil

}

func preDelete(_ context.Context, cr *svcapitypes.Broker, obj *svcsdk.DeleteBrokerInput) (bool, error) {
	obj.BrokerId = awsclients.String(meta.GetExternalName(cr))
	return false, nil
}

func (e *custom) preCreate(ctx context.Context, cr *svcapitypes.Broker, obj *svcsdk.CreateBrokerRequest) error {

	obj.BrokerName = awsclients.String(cr.Name)

	pw, _, err := mq.GetPassword(ctx, e.kube, &cr.Spec.ForProvider.CustomUsers[0].PasswordSecretRef, cr.Spec.WriteConnectionSecretToReference)
	if resource.IgnoreNotFound(err) != nil || pw == "" {
		return errors.Wrap(err, "cannot get password from the given secret")
	}

	obj.Users = []*svcsdk.User{
		{
			Username:      cr.Spec.ForProvider.CustomUsers[0].Username,
			Password:      awsclients.String(pw),
			ConsoleAccess: cr.Spec.ForProvider.CustomUsers[0].ConsoleAccess,
			Groups:        cr.Spec.ForProvider.CustomUsers[0].Groups,
		},
	}

	return nil
}

func postCreate(_ context.Context, cr *svcapitypes.Broker, obj *svcsdk.CreateBrokerResponse, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	meta.SetExternalName(cr, awsclients.StringValue(obj.BrokerId))
	return cre, nil
}

// LateInitialize fills the empty fields in *svcapitypes.BrokerParameters with
// the values seen in svcsdk.DescribeBrokerResponse.
// nolint:gocyclo
func LateInitialize(cr *svcapitypes.BrokerParameters, obj *svcsdk.DescribeBrokerResponse) error {
	if cr.AutoMinorVersionUpgrade == nil && obj.AutoMinorVersionUpgrade != nil {
		cr.AutoMinorVersionUpgrade = awsclients.LateInitializeBoolPtr(cr.AutoMinorVersionUpgrade, obj.AutoMinorVersionUpgrade)
	}

	if cr.PubliclyAccessible == nil && obj.PubliclyAccessible != nil {
		cr.PubliclyAccessible = awsclients.LateInitializeBoolPtr(cr.PubliclyAccessible, obj.PubliclyAccessible)
	}

	return nil
}
