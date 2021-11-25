package user

import (
	"context"
	"strings"
	"time"

	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	svcsdk "github.com/aws/aws-sdk-go/service/elasticache"
	svcsdkapi "github.com/aws/aws-sdk-go/service/elasticache/elasticacheiface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"

	svcapitypes "github.com/crossplane/provider-aws/apis/elasticache/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/elasticache"
)

const active = "active"

// SetupUser adds a controller that reconciles User.
func SetupUser(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(svcapitypes.UserGroupKind)
	opts := []option{
		func(e *external) {
			c := &custom{client: e.client, kube: e.kube, external: e}
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
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewController(rl),
		}).
		For(&svcapitypes.User{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.UserGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type custom struct {
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

func (e *custom) preCreate(ctx context.Context, cr *svcapitypes.User, obj *svcsdk.CreateUserInput) error {
	obj.UserId = awsclients.String(meta.GetExternalName(cr))

	if !awsclients.BoolValue(cr.Spec.ForProvider.NoPasswordRequired) {
		pw, _, err := elasticache.GetPassword(ctx, e.kube, &cr.Spec.ForProvider.PasswordSecretRef, cr.Spec.WriteConnectionSecretToReference)
		if resource.IgnoreNotFound(err) != nil {
			return errors.Wrap(err, "cannot get password from the given secret")
		}
		obj.Passwords = append(obj.Passwords, awsclients.String(pw))
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

func (e *custom) isUpToDate(cr *svcapitypes.User, obj *svcsdk.DescribeUsersOutput) (bool, error) {
	ctx := context.Background()

	if awsclients.StringValue(obj.Users[0].Status) != active {
		return true, nil
	}

	// to fields ex. $AccessString vs. $AccessString -@all
	if awsclients.StringValue(cr.Spec.ForProvider.AccessString) != strings.Fields(awsclients.StringValue(obj.Users[0].AccessString))[0] {
		return false, nil
	}

	_, pwChanged, err := elasticache.GetPassword(ctx, e.kube, &cr.Spec.ForProvider.PasswordSecretRef, cr.Spec.WriteConnectionSecretToReference)
	if err != nil {
		return false, err
	}
	return !pwChanged, nil
}

func (e *custom) preUpdate(ctx context.Context, cr *svcapitypes.User, obj *svcsdk.ModifyUserInput) error {
	obj.UserId = awsclients.String(meta.GetExternalName(cr))

	pw, pwchanged, err := elasticache.GetPassword(ctx, e.kube, &cr.Spec.ForProvider.PasswordSecretRef, cr.Spec.WriteConnectionSecretToReference)
	if err != nil {
		return err
	}

	if pwchanged {
		obj.Passwords = append(obj.Passwords, awsclients.String(pw))
	}
	return nil
}

func (e *custom) postUpdate(ctx context.Context, cr *svcapitypes.User, obj *svcsdk.ModifyUserOutput, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	if err != nil {
		return upd, err
	}

	pw, _, err := elasticache.GetPassword(ctx, e.kube, &cr.Spec.ForProvider.PasswordSecretRef, cr.Spec.WriteConnectionSecretToReference)
	if err != nil {
		return upd, err
	}

	return managed.ExternalUpdate{ConnectionDetails: managed.ConnectionDetails{
		xpv1.ResourceCredentialsSecretPasswordKey: []byte(pw),
	}}, nil

}
