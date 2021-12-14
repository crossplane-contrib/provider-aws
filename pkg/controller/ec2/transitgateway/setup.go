package transitgateway

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/ec2"
	svcsdkapi "github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"

	svcapitypes "github.com/crossplane/provider-aws/apis/ec2/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"

	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

const (
	errKubeUpdateFailed = "cannot update TransitGateway"
)

// SetupTransitGateway adds a controller that reconciles TransitGateway.
func SetupTransitGateway(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(svcapitypes.TransitGatewayGroupKind)
	opts := []option{
		func(e *external) {
			c := &custom{client: e.client, kube: e.kube}
			e.postObserve = c.postObserve
			e.postCreate = c.postCreate
			e.isUpToDate = isUpToDate
			e.lateInitialize = LateInitialize
			e.filterList = filterList
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewController(rl),
		}).
		For(&svcapitypes.TransitGateway{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.TransitGatewayGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient()), &tagger{kube: mgr.GetClient()}),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type custom struct {
	kube   client.Client
	client svcsdkapi.EC2API
}

func filterList(cr *svcapitypes.TransitGateway, obj *svcsdk.DescribeTransitGatewaysOutput) *svcsdk.DescribeTransitGatewaysOutput {
	resp := &svcsdk.DescribeTransitGatewaysOutput{}
	for _, TransitGateway := range obj.TransitGateways {
		if aws.StringValue(TransitGateway.TransitGatewayId) == meta.GetExternalName(cr) {
			resp.TransitGateways = append(resp.TransitGateways, TransitGateway)
			break
		}
	}
	return resp
}

func (e *custom) postObserve(_ context.Context, cr *svcapitypes.TransitGateway, obj *svcsdk.DescribeTransitGatewaysOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	switch awsclients.StringValue(obj.TransitGateways[0].State) {
	case string(svcapitypes.TransitGatewayState_available):
		cr.SetConditions(xpv1.Available())
	case string(svcapitypes.TransitGatewayState_pending):
		cr.SetConditions(xpv1.Creating())
	case string(svcapitypes.TransitGatewayState_modifying):
		cr.SetConditions(xpv1.Unavailable())
	case string(svcapitypes.TransitGatewayState_deleting):
		cr.SetConditions(xpv1.Deleting())
	case string(svcapitypes.TransitGatewayState_deleted):
		// TransitGateway is in status deleted - and really removed after 6 hours in aws
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	return obs, nil
}

func isUpToDate(cr *svcapitypes.TransitGateway, obj *svcsdk.DescribeTransitGatewaysOutput) (bool, error) {
	return true, nil
}

func (e *custom) postCreate(ctx context.Context, cr *svcapitypes.TransitGateway, obj *svcsdk.CreateTransitGatewayOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	// set transit gateway id as external name annotation on k8s object after creation
	meta.SetExternalName(cr, aws.StringValue(obj.TransitGateway.TransitGatewayId))
	return cre, nil
}

// LateInitialize fills the empty fields in *svcapitypes.TransitGateway with
// the values seen in svcsdk.DescribeTransitGatewaysOutput.
// nolint:gocyclo
func LateInitialize(cr *svcapitypes.TransitGatewayParameters, obj *svcsdk.DescribeTransitGatewaysOutput) error { // nolint:gocyclo
	if len(obj.TransitGateways) > 0 {
		cr.Options = &svcapitypes.TransitGatewayRequestOptions{
			AmazonSideASN:                obj.TransitGateways[0].Options.AmazonSideAsn,
			DNSSupport:                   obj.TransitGateways[0].Options.DnsSupport,
			AutoAcceptSharedAttachments:  obj.TransitGateways[0].Options.AutoAcceptSharedAttachments,
			DefaultRouteTableAssociation: obj.TransitGateways[0].Options.DefaultRouteTableAssociation,
			DefaultRouteTablePropagation: obj.TransitGateways[0].Options.DefaultRouteTablePropagation,
			MulticastSupport:             obj.TransitGateways[0].Options.MulticastSupport,
			VPNECMPSupport:               obj.TransitGateways[0].Options.VpnEcmpSupport,
			TransitGatewayCIDRBlocks:     obj.TransitGateways[0].Options.TransitGatewayCidrBlocks,
		}
	}

	return nil
}

type tagger struct {
	kube client.Client
}

func (t *tagger) Initialize(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*svcapitypes.TransitGateway)
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	added := false
	tagMap := map[string]string{}
	for _, t := range cr.Spec.ForProvider.Tags {
		tagMap[awsclients.StringValue(t.Key)] = awsclients.StringValue(t.Value)
	}
	for k, v := range resource.GetExternalTags(mgd) {
		if tagMap[k] != v {
			cr.Spec.ForProvider.Tags = append(cr.Spec.ForProvider.Tags, svcapitypes.Tag{Key: awsclients.String(k), Value: awsclients.String(v)})
			added = true
		}
	}
	if !added {
		return nil
	}
	return errors.Wrap(t.kube.Update(ctx, cr), errKubeUpdateFailed)
}
