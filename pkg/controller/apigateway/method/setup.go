package method

import (
	"context"
	"fmt"
	"strings"

	svcsdk "github.com/aws/aws-sdk-go/service/apigateway"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/apigateway/v1alpha1"
	aws "github.com/crossplane-contrib/provider-aws/pkg/clients"
	apigwclient "github.com/crossplane-contrib/provider-aws/pkg/clients/apigateway"
)

// ControllerName of this controller.
var ControllerName = managed.ControllerName(svcapitypes.MethodGroupKind)

// SetupMethod adds a controller that reconciles Method.
func SetupMethod(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.MethodGroupKind)
	opts := []option{
		func(e *external) {
			c := &custom{
				Client: &apigwclient.GatewayClient{Client: e.client},
			}
			e.preCreate = c.preCreate
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.postCreate = postCreate
			e.preDelete = preDelete
			e.lateInitialize = c.lateInitialize
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&svcapitypes.Method{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.MethodGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithInitializers(),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type custom struct {
	Client *apigwclient.GatewayClient
}

func getResourceIDFromExternalName(cr *svcapitypes.Method) *string {
	ext := meta.GetExternalName(cr)
	spl := strings.Split(ext, "-")
	val := aws.StringValue(&spl[1])

	return &val
}

func preObserve(_ context.Context, cr *svcapitypes.Method, obj *svcsdk.GetMethodInput) error {
	obj.RestApiId = cr.Spec.ForProvider.RestAPIID
	obj.ResourceId = getResourceIDFromExternalName(cr)

	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.Method, _ *svcsdk.Method, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	cr.SetConditions(xpv1.Available())
	return obs, nil
}

func postCreate(_ context.Context, cr *svcapitypes.Method, resp *svcsdk.Method, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	meta.SetExternalName(cr, fmt.Sprintf(
		"%s-%s-%s",
		aws.StringValue(cr.Spec.ForProvider.RestAPIID),
		aws.StringValue(cr.Spec.ForProvider.ResourceID),
		aws.StringValue(cr.Spec.ForProvider.HTTPMethod),
	))
	return cre, nil
}

func (c *custom) preCreate(ctx context.Context, cr *svcapitypes.Method, obj *svcsdk.PutMethodInput) error {
	obj.RestApiId = cr.Spec.ForProvider.RestAPIID
	if cr.Spec.ForProvider.ResourceID == nil {
		resourceID, err := c.Client.GetRestAPIRootResource(ctx, obj.RestApiId)
		if err != nil {
			return errors.Wrap(err, "couldnt get root resource for restapi")
		}
		obj.ResourceId = resourceID
		cr.Spec.ForProvider.ResourceID = resourceID
	} else {
		obj.ResourceId = cr.Spec.ForProvider.ResourceID
	}
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.Method, obj *svcsdk.DeleteMethodInput) (bool, error) {
	if cr.Spec.ForProvider.ResourceID == nil {
		cr.Spec.ForProvider.ResourceID = getResourceIDFromExternalName(cr)
	} else {
		obj.ResourceId = cr.Spec.ForProvider.ResourceID
	}

	obj.RestApiId = cr.Spec.ForProvider.RestAPIID

	return false, nil
}

func (c *custom) lateInitialize(cr *svcapitypes.MethodParameters, cur *svcsdk.Method) error {
	if cr.ResourceID == nil {
		resourceID, err := c.Client.GetRestAPIRootResource(context.TODO(), cr.RestAPIID)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("lateinit method %s for restapi %s", *cr.HTTPMethod, *cr.RestAPIID))
		}
		cr.ResourceID = resourceID
	}
	cr.APIKeyRequired = aws.LateInitializeBoolPtr(cr.APIKeyRequired, cur.ApiKeyRequired)
	cr.AuthorizationScopes = aws.LateInitializeStringPtrSlice(cr.AuthorizationScopes, cur.AuthorizationScopes)
	cr.AuthorizationType = aws.LateInitializeStringPtr(cr.AuthorizationType, cur.AuthorizationType)
	cr.OperationName = aws.LateInitializeStringPtr(cr.OperationName, cur.OperationName)

	if cr.RequestModels == nil && cur.RequestModels != nil {
		cr.RequestModels = cur.RequestModels
	}
	if cr.RequestParameters == nil && cur.RequestParameters != nil {
		cr.RequestParameters = cur.RequestParameters
	}

	return nil
}
