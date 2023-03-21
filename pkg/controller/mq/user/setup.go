package user

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/mq"
	svcsdkapi "github.com/aws/aws-sdk-go/service/mq/mqiface"
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

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/mq/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	awsclients "github.com/crossplane-contrib/provider-aws/pkg/clients"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/mq"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
)

// SetupUser adds a controller that reconciles User.
func SetupUser(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.UserGroupKind)
	opts := []option{
		func(e *external) {
			c := &custom{client: e.client, kube: e.kube, external: e}
			e.isUpToDate = c.isUpToDate
			e.preCreate = c.preCreate
			e.postCreate = postCreate
			e.preObserve = preObserve
			e.preDelete = preDelete
			e.postObserve = c.postObserve
			e.preUpdate = c.preUpdate
			e.postUpdate = c.postUpdate
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
		For(&svcapitypes.User{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.UserGroupVersionKind),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

type custom struct {
	kube     client.Client
	client   svcsdkapi.MQAPI
	external *external
}

func preObserve(_ context.Context, cr *svcapitypes.User, obj *svcsdk.DescribeUserInput) error {
	obj.BrokerId = cr.Spec.ForProvider.BrokerID
	obj.Username = awsclients.String(meta.GetExternalName(cr))
	return nil
}

func (e *custom) postObserve(ctx context.Context, cr *svcapitypes.User, obj *svcsdk.DescribeUserResponse, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	// obj.Pending.PendingChange is nil if User is available
	if obj.Pending != nil {
		switch awsclients.StringValue(obj.Pending.PendingChange) {
		case string(svcapitypes.ChangeType_CREATE):
			cr.SetConditions(xpv1.Creating().WithMessage("wait for the next maintenance window or reboot the broker."))
		case string(svcapitypes.ChangeType_DELETE):
			cr.SetConditions(xpv1.Deleting().WithMessage("wait for the next maintenance window or reboot the broker."))
		case string(svcapitypes.ChangeType_UPDATE):
			cr.SetConditions(xpv1.Available().WithMessage("wait for the next maintenance window or reboot the broker."))
		}
		return obs, nil
	}

	cr.SetConditions(xpv1.Available())

	pw, _, err := mq.GetPassword(ctx, e.kube, &cr.Spec.ForProvider.PasswordSecretRef, cr.Spec.WriteConnectionSecretToReference)
	if resource.IgnoreNotFound(err) != nil || pw == "" {
		return obs, errors.Wrap(err, "cannot get password from the given secret")
	}
	obs.ConnectionDetails = managed.ConnectionDetails{
		"Password": []byte(pw),
	}
	return obs, nil
}

func preDelete(_ context.Context, cr *svcapitypes.User, obj *svcsdk.DeleteUserInput) (bool, error) {
	obj.BrokerId = cr.Spec.ForProvider.BrokerID
	obj.Username = awsclients.String(meta.GetExternalName(cr))

	return false, nil
}

func (e *custom) preCreate(ctx context.Context, cr *svcapitypes.User, obj *svcsdk.CreateUserRequest) error {
	brokerState, err := e.client.DescribeBroker(
		&svcsdk.DescribeBrokerInput{
			BrokerId: cr.Spec.ForProvider.BrokerID,
		},
	)
	if err != nil {
		return err
	}

	if awsclients.StringValue(brokerState.BrokerState) != svcsdk.BrokerStateRunning ||
		awsclients.StringValue(brokerState.BrokerState) == svcsdk.BrokerStateDeletionInProgress {
		return errors.New("broker is not ready for user creation " + awsclients.StringValue(brokerState.BrokerState))
	}

	pw, _, err := mq.GetPassword(ctx, e.kube, &cr.Spec.ForProvider.PasswordSecretRef, cr.Spec.WriteConnectionSecretToReference)
	if resource.IgnoreNotFound(err) != nil {
		return errors.Wrap(err, "cannot get password from the given secret")
	}
	obj.Password = awsclients.String(pw)
	obj.Username = awsclients.String(cr.Name)
	obj.BrokerId = cr.Spec.ForProvider.BrokerID
	return nil
}

func postCreate(_ context.Context, cr *svcapitypes.User, obj *svcsdk.CreateUserOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, cr.Name)
	return cre, nil
}

func (e *custom) preUpdate(ctx context.Context, cr *svcapitypes.User, obj *svcsdk.UpdateUserRequest) error {
	obj.BrokerId = cr.Spec.ForProvider.BrokerID
	obj.Username = awsclients.String(cr.Name)

	pw, pwchanged, err := mq.GetPassword(ctx, e.kube, &cr.Spec.ForProvider.PasswordSecretRef, cr.Spec.WriteConnectionSecretToReference)
	if err != nil {
		return err
	}
	if pwchanged {
		obj.Password = aws.String(pw)
	}
	return nil
}

func (e *custom) postUpdate(ctx context.Context, cr *svcapitypes.User, obj *svcsdk.UpdateUserOutput, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	if err != nil {
		return upd, err
	}

	pw, _, err := mq.GetPassword(ctx, e.kube, &cr.Spec.ForProvider.PasswordSecretRef, cr.Spec.WriteConnectionSecretToReference)
	if err != nil {
		return upd, err
	}

	var conn = managed.ConnectionDetails{
		xpv1.ResourceCredentialsSecretPasswordKey: []byte(pw),
	}
	return managed.ExternalUpdate{ConnectionDetails: conn}, nil
}

func (e *custom) isUpToDate(cr *svcapitypes.User, obj *svcsdk.DescribeUserResponse) (bool, error) {
	ctx := context.Background()

	if obj.Pending != nil {
		return true, nil
	}
	_, pwChanged, err := mq.GetPassword(ctx, e.kube, &cr.Spec.ForProvider.PasswordSecretRef, cr.Spec.WriteConnectionSecretToReference)
	if err != nil {
		return false, err
	}
	return !pwChanged, nil
}
