package listener

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/elbv2"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/elbv2/v1alpha1"
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
			e.postObserve = postObserve
			e.preCreate = preCreate
			e.postCreate = postCreate
			e.preDelete = preDelete
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
		managed.WithInitializers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(svcapitypes.ListenerGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.Listener{}).
		Complete(r)
}

func preObserve(_ context.Context, cr *svcapitypes.Listener, obj *svcsdk.DescribeListenersInput) error {
	obj.ListenerArns = append(obj.ListenerArns, aws.String(meta.GetExternalName(cr)))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.Listener, _ *svcsdk.DescribeListenersOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(xpv1.Available())
	return obs, nil
}

func generateDefaultActions(cr *svcapitypes.Listener) []*svcsdk.Action { //nolint:gocyclo // This func is long by necessity of needing to recursively copy all values from the API type into the SDK type
	actions := []*svcsdk.Action{}
	if cr.Spec.ForProvider.DefaultActions == nil {
		return actions
	}

	for _, actionsiter := range cr.Spec.ForProvider.DefaultActions {
		actionselem := &svcsdk.Action{}
		if actionsiter.AuthenticateCognitoConfig != nil {
			actionselemf0 := &svcsdk.AuthenticateCognitoActionConfig{}
			if actionsiter.AuthenticateCognitoConfig.AuthenticationRequestExtraParams != nil {
				actionselemf0f0 := map[string]*string{}
				for actionselemf0f0key, actionselemf0f0valiter := range actionsiter.AuthenticateCognitoConfig.AuthenticationRequestExtraParams {
					actionselemf0f0val := *actionselemf0f0valiter
					actionselemf0f0[actionselemf0f0key] = &actionselemf0f0val
				}
				actionselemf0.SetAuthenticationRequestExtraParams(actionselemf0f0)
			}
			if actionsiter.AuthenticateCognitoConfig.OnUnauthenticatedRequest != nil {
				actionselemf0.SetOnUnauthenticatedRequest(*actionsiter.AuthenticateCognitoConfig.OnUnauthenticatedRequest)
			}
			if actionsiter.AuthenticateCognitoConfig.Scope != nil {
				actionselemf0.SetScope(*actionsiter.AuthenticateCognitoConfig.Scope)
			}
			if actionsiter.AuthenticateCognitoConfig.SessionCookieName != nil {
				actionselemf0.SetSessionCookieName(*actionsiter.AuthenticateCognitoConfig.SessionCookieName)
			}
			if actionsiter.AuthenticateCognitoConfig.SessionTimeout != nil {
				actionselemf0.SetSessionTimeout(*actionsiter.AuthenticateCognitoConfig.SessionTimeout)
			}
			if actionsiter.AuthenticateCognitoConfig.UserPoolARN != nil {
				actionselemf0.SetUserPoolArn(*actionsiter.AuthenticateCognitoConfig.UserPoolARN)
			}
			if actionsiter.AuthenticateCognitoConfig.UserPoolClientID != nil {
				actionselemf0.SetUserPoolClientId(*actionsiter.AuthenticateCognitoConfig.UserPoolClientID)
			}
			if actionsiter.AuthenticateCognitoConfig.UserPoolDomain != nil {
				actionselemf0.SetUserPoolDomain(*actionsiter.AuthenticateCognitoConfig.UserPoolDomain)
			}
			actionselem.SetAuthenticateCognitoConfig(actionselemf0)
		}
		if actionsiter.AuthenticateOidcConfig != nil {
			actionselemf1 := &svcsdk.AuthenticateOidcActionConfig{}
			if actionsiter.AuthenticateOidcConfig.AuthenticationRequestExtraParams != nil {
				actionselemf1f0 := map[string]*string{}
				for actionselemf1f0key, actionselemf1f0valiter := range actionsiter.AuthenticateOidcConfig.AuthenticationRequestExtraParams {
					actionselemf1f0val := *actionselemf1f0valiter
					actionselemf1f0[actionselemf1f0key] = &actionselemf1f0val
				}
				actionselemf1.SetAuthenticationRequestExtraParams(actionselemf1f0)
			}
			if actionsiter.AuthenticateOidcConfig.AuthorizationEndpoint != nil {
				actionselemf1.SetAuthorizationEndpoint(*actionsiter.AuthenticateOidcConfig.AuthorizationEndpoint)
			}
			if actionsiter.AuthenticateOidcConfig.ClientID != nil {
				actionselemf1.SetClientId(*actionsiter.AuthenticateOidcConfig.ClientID)
			}
			if actionsiter.AuthenticateOidcConfig.ClientSecret != nil {
				actionselemf1.SetClientSecret(*actionsiter.AuthenticateOidcConfig.ClientSecret)
			}
			if actionsiter.AuthenticateOidcConfig.Issuer != nil {
				actionselemf1.SetIssuer(*actionsiter.AuthenticateOidcConfig.Issuer)
			}
			if actionsiter.AuthenticateOidcConfig.OnUnauthenticatedRequest != nil {
				actionselemf1.SetOnUnauthenticatedRequest(*actionsiter.AuthenticateOidcConfig.OnUnauthenticatedRequest)
			}
			if actionsiter.AuthenticateOidcConfig.Scope != nil {
				actionselemf1.SetScope(*actionsiter.AuthenticateOidcConfig.Scope)
			}
			if actionsiter.AuthenticateOidcConfig.SessionCookieName != nil {
				actionselemf1.SetSessionCookieName(*actionsiter.AuthenticateOidcConfig.SessionCookieName)
			}
			if actionsiter.AuthenticateOidcConfig.SessionTimeout != nil {
				actionselemf1.SetSessionTimeout(*actionsiter.AuthenticateOidcConfig.SessionTimeout)
			}
			if actionsiter.AuthenticateOidcConfig.TokenEndpoint != nil {
				actionselemf1.SetTokenEndpoint(*actionsiter.AuthenticateOidcConfig.TokenEndpoint)
			}
			if actionsiter.AuthenticateOidcConfig.UseExistingClientSecret != nil {
				actionselemf1.SetUseExistingClientSecret(*actionsiter.AuthenticateOidcConfig.UseExistingClientSecret)
			}
			if actionsiter.AuthenticateOidcConfig.UserInfoEndpoint != nil {
				actionselemf1.SetUserInfoEndpoint(*actionsiter.AuthenticateOidcConfig.UserInfoEndpoint)
			}
			actionselem.SetAuthenticateOidcConfig(actionselemf1)
		}
		if actionsiter.FixedResponseConfig != nil {
			actionselemactions := &svcsdk.FixedResponseActionConfig{}
			if actionsiter.FixedResponseConfig.ContentType != nil {
				actionselemactions.SetContentType(*actionsiter.FixedResponseConfig.ContentType)
			}
			if actionsiter.FixedResponseConfig.MessageBody != nil {
				actionselemactions.SetMessageBody(*actionsiter.FixedResponseConfig.MessageBody)
			}
			if actionsiter.FixedResponseConfig.StatusCode != nil {
				actionselemactions.SetStatusCode(*actionsiter.FixedResponseConfig.StatusCode)
			}
			actionselem.SetFixedResponseConfig(actionselemactions)
		}
		if actionsiter.ForwardConfig != nil {
			actionselemf3 := &svcsdk.ForwardActionConfig{}
			if actionsiter.ForwardConfig.TargetGroupStickinessConfig != nil {
				actionselemf3f0 := &svcsdk.TargetGroupStickinessConfig{}
				if actionsiter.ForwardConfig.TargetGroupStickinessConfig.DurationSeconds != nil {
					actionselemf3f0.SetDurationSeconds(*actionsiter.ForwardConfig.TargetGroupStickinessConfig.DurationSeconds)
				}
				if actionsiter.ForwardConfig.TargetGroupStickinessConfig.Enabled != nil {
					actionselemf3f0.SetEnabled(*actionsiter.ForwardConfig.TargetGroupStickinessConfig.Enabled)
				}
				actionselemf3.SetTargetGroupStickinessConfig(actionselemf3f0)
			}
			if actionsiter.ForwardConfig.TargetGroups != nil {
				actionselemf3f1 := []*svcsdk.TargetGroupTuple{}
				for _, actionselemf3f1iter := range actionsiter.ForwardConfig.TargetGroups {
					actionselemf3f1elem := &svcsdk.TargetGroupTuple{}
					if actionselemf3f1iter.TargetGroupARN != nil {
						actionselemf3f1elem.SetTargetGroupArn(*actionselemf3f1iter.TargetGroupARN)
					}
					if actionselemf3f1iter.Weight != nil {
						actionselemf3f1elem.SetWeight(*actionselemf3f1iter.Weight)
					}
					actionselemf3f1 = append(actionselemf3f1, actionselemf3f1elem)
				}
				actionselemf3.SetTargetGroups(actionselemf3f1)
			}
			actionselem.SetForwardConfig(actionselemf3)
		}
		if actionsiter.Order != nil {
			actionselem.SetOrder(*actionsiter.Order)
		}
		if actionsiter.RedirectConfig != nil {
			actionselemf5 := &svcsdk.RedirectActionConfig{}
			if actionsiter.RedirectConfig.Host != nil {
				actionselemf5.SetHost(*actionsiter.RedirectConfig.Host)
			}
			if actionsiter.RedirectConfig.Path != nil {
				actionselemf5.SetPath(*actionsiter.RedirectConfig.Path)
			}
			if actionsiter.RedirectConfig.Port != nil {
				actionselemf5.SetPort(*actionsiter.RedirectConfig.Port)
			}
			if actionsiter.RedirectConfig.Protocol != nil {
				actionselemf5.SetProtocol(*actionsiter.RedirectConfig.Protocol)
			}
			if actionsiter.RedirectConfig.Query != nil {
				actionselemf5.SetQuery(*actionsiter.RedirectConfig.Query)
			}
			if actionsiter.RedirectConfig.StatusCode != nil {
				actionselemf5.SetStatusCode(*actionsiter.RedirectConfig.StatusCode)
			}
			actionselem.SetRedirectConfig(actionselemf5)
		}
		if actionsiter.TargetGroupARN != nil {
			actionselem.SetTargetGroupArn(*actionsiter.TargetGroupARN)
		}
		if actionsiter.Type != nil {
			actionselem.SetType(*actionsiter.Type)
		}
		actions = append(actions, actionselem)
	}
	return actions
}

func preCreate(_ context.Context, cr *svcapitypes.Listener, obs *svcsdk.CreateListenerInput) error {
	obs.DefaultActions = generateDefaultActions(cr)
	obs.LoadBalancerArn = cr.Spec.ForProvider.LoadBalancerARN
	for i := range cr.Spec.ForProvider.Certificates {
		if cr.Spec.ForProvider.Certificates[i].CertificateARN != nil {
			obs.Certificates = append(obs.Certificates, &svcsdk.Certificate{
				CertificateArn: cr.Spec.ForProvider.Certificates[i].CertificateARN,
			})
		}
	}
	return nil
}

func postCreate(_ context.Context, cr *svcapitypes.Listener, resp *svcsdk.CreateListenerOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, aws.StringValue(resp.Listeners[0].ListenerArn))
	return cre, nil
}

func preDelete(_ context.Context, cr *svcapitypes.Listener, obj *svcsdk.DeleteListenerInput) (bool, error) {
	obj.ListenerArn = aws.String(meta.GetExternalName(cr))
	return false, nil
}
