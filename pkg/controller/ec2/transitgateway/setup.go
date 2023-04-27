package transitgateway

import (
	"context"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/ec2"
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

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/ec2/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	awsclients "github.com/crossplane-contrib/provider-aws/pkg/clients"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
)

const (
	errKubeUpdateFailed = "cannot update TransitGateway"
)

// SetupTransitGateway adds a controller that reconciles TransitGateway.
func SetupTransitGateway(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.TransitGatewayGroupKind)
	opts := []option{
		func(e *external) {
			e.postObserve = postObserve
			e.postCreate = postCreate
			e.isUpToDate = isUpToDate
			e.lateInitialize = LateInitialize
			e.filterList = filterList
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithInitializers(&tagger{kube: mgr.GetClient()}),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(svcapitypes.TransitGatewayGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.TransitGateway{}).
		Complete(r)
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

func postObserve(_ context.Context, cr *svcapitypes.TransitGateway, obj *svcsdk.DescribeTransitGatewaysOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
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

func postCreate(ctx context.Context, cr *svcapitypes.TransitGateway, obj *svcsdk.CreateTransitGatewayOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
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
	var transitGatewayTags svcapitypes.TagSpecification
	for _, tagSpecification := range cr.Spec.ForProvider.TagSpecifications {
		if aws.StringValue(tagSpecification.ResourceType) == "transit-gateway" {
			transitGatewayTags = *tagSpecification
		}
	}

	tagMap := map[string]string{}
	tagMap["Name"] = cr.Name
	for _, t := range cr.Spec.ForProvider.Tags {
		tagMap[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
	}
	for k, v := range resource.GetExternalTags(mgd) {
		tagMap[k] = v
	}
	transitGatewayTags.Tags = make([]*svcapitypes.Tag, len(tagMap))
	transitGatewayTags.ResourceType = aws.String("transit-gateway")
	i := 0
	for k, v := range tagMap {
		transitGatewayTags.Tags[i] = &svcapitypes.Tag{Key: aws.String(k), Value: aws.String(v)}
		i++
	}
	sort.Slice(transitGatewayTags.Tags, func(i, j int) bool {
		return aws.StringValue(transitGatewayTags.Tags[i].Key) < aws.StringValue(transitGatewayTags.Tags[j].Key)
	})

	cr.Spec.ForProvider.TagSpecifications = []*svcapitypes.TagSpecification{&transitGatewayTags}
	return errors.Wrap(t.kube.Update(ctx, cr), errKubeUpdateFailed)
}
