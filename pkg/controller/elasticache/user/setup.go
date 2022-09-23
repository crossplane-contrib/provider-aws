package user

import (
	"context"
	"strings"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcsdk "github.com/aws/aws-sdk-go/service/elasticache"
	svcsdkapi "github.com/aws/aws-sdk-go/service/elasticache/elasticacheiface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/elasticache/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	awsclients "github.com/crossplane-contrib/provider-aws/pkg/clients"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/elasticache"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
)

const active = "active"

// SetupUser adds a controller that reconciles User.
func SetupUser(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.UserGroupKind)
	opts := []option{setupExternal}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&svcapitypes.User{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.UserGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

func setupExternal(e *external) {
	c := &hooks{client: e.client, kube: e.kube, external: e}
	e.isUpToDate = c.isUpToDate
	e.preCreate = c.preCreate
	e.postCreate = postCreate
	e.preObserve = preObserve
	e.preDelete = preDelete
	e.postObserve = postObserve
	e.postDelete = postDelete
	e.preUpdate = c.preUpdate
	e.postUpdate = c.postUpdate
	e.filterList = filterList
}

type hooks struct {
	kube     client.Client
	client   svcsdkapi.ElastiCacheAPI
	external *external
}

func filterList(cr *svcapitypes.User, obj *svcsdk.DescribeUsersOutput) *svcsdk.DescribeUsersOutput {
	resp := &svcsdk.DescribeUsersOutput{}
	for _, user := range obj.Users {
		if awsclients.StringValue(user.UserId) == meta.GetExternalName(cr) {
			resp.Users = append(resp.Users, user)
			break
		}
	}
	return resp
}

func (e *hooks) preCreate(ctx context.Context, cr *svcapitypes.User, obj *svcsdk.CreateUserInput) error {
	obj.UserId = awsclients.String(meta.GetExternalName(cr))

	if !awsclients.BoolValue(cr.Spec.ForProvider.NoPasswordRequired) {
		pwds, err := elasticache.GetPasswords(ctx, e.kube, cr.Spec.ForProvider.PasswordSecretRef)
		if resource.IgnoreNotFound(err) != nil {
			return errors.Wrap(err, "cannot get password from the given secret")
		}
		obj.Passwords = pwds

	}

	return nil
}

func postCreate(_ context.Context, cr *svcapitypes.User, obj *svcsdk.CreateUserOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, awsclients.StringValue(obj.UserId))
	return cre, nil
}

func preObserve(_ context.Context, cr *svcapitypes.User, obj *svcsdk.DescribeUsersInput) error {
	obj.UserId = awsclients.String(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.User, obj *svcsdk.DescribeUsersOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	switch awsclients.StringValue(obj.Users[0].Status) {
	case "active":
		cr.SetConditions(xpv1.Available())
	case "modifying":
		cr.SetConditions(xpv1.Unavailable())
	case "deleting":
		cr.SetConditions(xpv1.Deleting())
	}

	return obs, nil
}

func preDelete(_ context.Context, cr *svcapitypes.User, obj *svcsdk.DeleteUserInput) (bool, error) {
	obj.UserId = awsclients.String(meta.GetExternalName(cr))
	return false, nil
}

func postDelete(_ context.Context, cr *svcapitypes.User, obj *svcsdk.DeleteUserOutput, err error) error {
	if err != nil {
		if strings.Contains(err.Error(), svcsdk.ErrCodeInvalidUserStateFault) {
			// skip: failed to delete User: InvalidUserState: User has status deleting.
			return nil
		}
		return err
	}
	return err
}

func (e *hooks) isUpToDate(cr *svcapitypes.User, obj *svcsdk.DescribeUsersOutput) (bool, error) {
	ctx := context.Background()

	if awsclients.StringValue(obj.Users[0].Status) != active {
		return true, nil
	}

	// to fields ex. $AccessString vs. $AccessString -@all
	if awsclients.StringValue(cr.Spec.ForProvider.AccessString) != strings.Fields(awsclients.StringValue(obj.Users[0].AccessString))[0] {
		return false, nil
	}

	isUpToDate, err := elasticache.PasswordsUpToDate(ctx, e.kube, cr.Spec.ForProvider.PasswordSecretRef, cr.Spec.WriteConnectionSecretToReference)
	if err != nil {
		return false, err
	}
	return isUpToDate, nil
}

func (e *hooks) preUpdate(ctx context.Context, cr *svcapitypes.User, obj *svcsdk.ModifyUserInput) error {
	obj.UserId = awsclients.String(meta.GetExternalName(cr))

	pwds, err := elasticache.GetPasswords(ctx, e.kube, cr.Spec.ForProvider.PasswordSecretRef)
	if err != nil {
		return err
	}

	obj.Passwords = pwds

	return nil
}

func (e *hooks) postUpdate(ctx context.Context, cr *svcapitypes.User, obj *svcsdk.ModifyUserOutput, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	if err != nil {
		return upd, err
	}

	pwds, err := elasticache.GetPasswords(ctx, e.kube, cr.Spec.ForProvider.PasswordSecretRef)
	if err != nil {
		return upd, err
	}

	connectionDetails := managed.ConnectionDetails{}

	if len(pwds) > 0 {
		connectionDetails[elasticache.PasswordKey1] = []byte(awsclients.StringValue(pwds[0]))
		if len(pwds) > 1 {
			connectionDetails[elasticache.PasswordKey2] = []byte(awsclients.StringValue(pwds[1]))
		} else {
			connectionDetails[elasticache.PasswordKey2] = nil
		}
	}

	return managed.ExternalUpdate{ConnectionDetails: connectionDetails}, nil

}
