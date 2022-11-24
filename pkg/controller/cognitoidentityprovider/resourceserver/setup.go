package resourceserver

import (
	"context"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	svcsdk "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/cognitoidentityprovider/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	awsclients "github.com/crossplane-contrib/provider-aws/pkg/clients"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
)

// SetupResourceServer adds a controller that reconciles Stage.
func SetupResourceServer(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.ResourceServerGroupKind)

	opts := []option{
		func(e *external) {
			e.postObserve = postObserve
			e.preObserve = preObserve
			e.preDelete = preDelete
			e.preCreate = preCreate
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&svcapitypes.ResourceServer{},
			builder.WithPredicates(predicate.Or(
				predicate.GenerationChangedPredicate{},
				predicate.LabelChangedPredicate{},
				predicate.AnnotationChangedPredicate{},
			))).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.ResourceServerGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))

}

func preObserve(_ context.Context, cr *svcapitypes.ResourceServer, obj *svcsdk.DescribeResourceServerInput) error {
	obj.Identifier = awsclients.String(meta.GetExternalName(cr))
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.ResourceServer, obj *svcsdk.DeleteResourceServerInput) (bool, error) {
	obj.Identifier = awsclients.String(meta.GetExternalName(cr))
	obj.UserPoolId = cr.Spec.ForProvider.UserPoolID
	return false, nil
}

func postObserve(_ context.Context, cr *svcapitypes.ResourceServer, obj *svcsdk.DescribeResourceServerOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	if obj.ResourceServer.UserPoolId == nil {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	cr.SetConditions(xpv1.Available())

	return obs, nil
}

func preCreate(_ context.Context, cr *svcapitypes.ResourceServer, obj *svcsdk.CreateResourceServerInput) error {
	obj.Identifier = awsclients.String(meta.GetExternalName(cr))
	obj.UserPoolId = cr.Spec.ForProvider.UserPoolID
	return nil
}
