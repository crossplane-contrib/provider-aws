package broker

import (
	"context"
	"strings"
	"time"

	svcsdk "github.com/aws/aws-sdk-go/service/mq"
	svcsdkapi "github.com/aws/aws-sdk-go/service/mq/mqiface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/mq/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/mq"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
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

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithInitializers(),
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
		resource.ManagedKind(svcapitypes.BrokerGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.Broker{}).
		Complete(r)
}

type custom struct {
	kube     client.Client
	client   svcsdkapi.MQAPI
	external *external
}

func preObserve(_ context.Context, cr *svcapitypes.Broker, obj *svcsdk.DescribeBrokerInput) error {
	obj.BrokerId = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

//nolint:gocyclo
func (e *custom) postObserve(ctx context.Context, cr *svcapitypes.Broker, obj *svcsdk.DescribeBrokerResponse, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	switch pointer.StringValue(obj.BrokerState) {
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

	// not needed if we get the field properly set in GenerateBroker() (metav1 import issue)
	cr.Status.AtProvider.Created = fromTimePtr(obj.Created)

	// somehow not in zz_conversions.go/GenerateBroker()
	if obj.Logs != nil {
		cr.Status.AtProvider.LogsSummary = &svcapitypes.LogsSummary{
			Audit:           obj.Logs.Audit,
			AuditLogGroup:   obj.Logs.AuditLogGroup,
			General:         obj.Logs.General,
			GeneralLogGroup: obj.Logs.GeneralLogGroup,
		}
		if obj.Logs.Pending != nil {
			cr.Status.AtProvider.LogsSummary.Pending = &svcapitypes.PendingLogs{
				Audit:   obj.Logs.Pending.Audit,
				General: obj.Logs.Pending.General,
			}
		}
	}

	obs.ConnectionDetails = managed.ConnectionDetails{
		"BrokerID": []byte(pointer.StringValue(cr.Status.AtProvider.BrokerID)),
		"Region":   []byte(pointer.StringValue(&cr.Spec.ForProvider.Region)),
		"Username": []byte(pointer.StringValue(cr.Spec.ForProvider.CustomUsers[0].Username)),
		"Password": []byte(pw),
	}

	return obs, nil

}

func preDelete(_ context.Context, cr *svcapitypes.Broker, obj *svcsdk.DeleteBrokerInput) (bool, error) {
	obj.BrokerId = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return false, nil
}

func (e *custom) preCreate(ctx context.Context, cr *svcapitypes.Broker, obj *svcsdk.CreateBrokerRequest) error {

	obj.BrokerName = pointer.ToOrNilIfZeroValue(cr.Name)

	pw, _, err := mq.GetPassword(ctx, e.kube, &cr.Spec.ForProvider.CustomUsers[0].PasswordSecretRef, cr.Spec.WriteConnectionSecretToReference)
	if resource.IgnoreNotFound(err) != nil || pw == "" {
		return errors.Wrap(err, "cannot get password from the given secret")
	}

	obj.SecurityGroups = cr.Spec.ForProvider.SecurityGroups
	obj.SubnetIds = cr.Spec.ForProvider.SubnetIDs

	obj.Users = []*svcsdk.User{
		{
			Username:      cr.Spec.ForProvider.CustomUsers[0].Username,
			Password:      pointer.ToOrNilIfZeroValue(pw),
			ConsoleAccess: cr.Spec.ForProvider.CustomUsers[0].ConsoleAccess,
			Groups:        cr.Spec.ForProvider.CustomUsers[0].Groups,
		},
	}

	// EncryptionOptions are not supported for RabbitMQ.
	// See https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-amazonmq-broker-encryptionoptions.html
	if strings.ToLower(ptr.Deref(obj.EngineType, "")) == "rabbitmq" {
		obj.EncryptionOptions = nil
	}

	return nil
}

func postCreate(_ context.Context, cr *svcapitypes.Broker, obj *svcsdk.CreateBrokerResponse, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	meta.SetExternalName(cr, pointer.StringValue(obj.BrokerId))
	return cre, nil
}

// LateInitialize fills the empty fields in *svcapitypes.BrokerParameters with
// the values seen in svcsdk.DescribeBrokerResponse.
//
//nolint:gocyclo
func LateInitialize(cr *svcapitypes.BrokerParameters, obj *svcsdk.DescribeBrokerResponse) error {
	if cr.AuthenticationStrategy == nil && obj.AuthenticationStrategy != nil {
		cr.AuthenticationStrategy = pointer.LateInitialize(cr.AuthenticationStrategy, obj.AuthenticationStrategy)
	}

	if cr.AutoMinorVersionUpgrade == nil && obj.AutoMinorVersionUpgrade != nil {
		cr.AutoMinorVersionUpgrade = pointer.LateInitialize(cr.AutoMinorVersionUpgrade, obj.AutoMinorVersionUpgrade)
	}

	if cr.EncryptionOptions == nil && obj.EncryptionOptions != nil {
		cr.EncryptionOptions = &svcapitypes.EncryptionOptions{
			KMSKeyID:       obj.EncryptionOptions.KmsKeyId,
			UseAWSOwnedKey: obj.EncryptionOptions.UseAwsOwnedKey,
		}
	}

	if cr.Logs == nil && obj.Logs != nil {
		cr.Logs = &svcapitypes.Logs{
			Audit:   obj.Logs.Audit,
			General: obj.Logs.General,
		}
	}

	if cr.MaintenanceWindowStartTime == nil && obj.MaintenanceWindowStartTime != nil {
		cr.MaintenanceWindowStartTime = &svcapitypes.WeeklyStartTime{
			DayOfWeek: obj.MaintenanceWindowStartTime.DayOfWeek,
			TimeOfDay: obj.MaintenanceWindowStartTime.TimeOfDay,
			TimeZone:  obj.MaintenanceWindowStartTime.TimeZone,
		}
	}

	if cr.PubliclyAccessible == nil && obj.PubliclyAccessible != nil {
		cr.PubliclyAccessible = pointer.LateInitialize(cr.PubliclyAccessible, obj.PubliclyAccessible)
	}

	return nil
}

// fromTimePtr probably not needed if metav1 import issue in zz_conversions.go is fixed
// see https://github.com/aws-controllers-k8s/community/issues/1372

// fromTimePtr is a helper for converting a *time.Time to a *metav1.Time
func fromTimePtr(t *time.Time) *metav1.Time {
	if t != nil {
		m := metav1.NewTime(*t)
		return &m
	}
	return nil
}
