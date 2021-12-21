package transitgatewayvpcattachment

import (
	"context"
	"sort"
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
	errKubeUpdateFailed = "cannot update TransitGatewayAttachment"
)

// SetupTransitGatewayVPCAttachment adds a controller that reconciles TransitGatewayVPCAttachment.
func SetupTransitGatewayVPCAttachment(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(svcapitypes.TransitGatewayVPCAttachmentGroupKind)
	opts := []option{
		func(e *external) {
			c := &custom{client: e.client, kube: e.kube}
			e.postObserve = postObserve
			e.postCreate = postCreate
			e.preCreate = c.preCreate
			e.filterList = filterList
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewController(rl),
		}).
		For(&svcapitypes.TransitGatewayVPCAttachment{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.TransitGatewayVPCAttachmentGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithInitializers(&tagger{kube: mgr.GetClient()}),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func filterList(cr *svcapitypes.TransitGatewayVPCAttachment, obj *svcsdk.DescribeTransitGatewayVpcAttachmentsOutput) *svcsdk.DescribeTransitGatewayVpcAttachmentsOutput {
	resp := &svcsdk.DescribeTransitGatewayVpcAttachmentsOutput{}
	for _, transitGatewayVpcAttachment := range obj.TransitGatewayVpcAttachments {
		if aws.StringValue(transitGatewayVpcAttachment.TransitGatewayAttachmentId) == meta.GetExternalName(cr) {
			resp.TransitGatewayVpcAttachments = append(resp.TransitGatewayVpcAttachments, transitGatewayVpcAttachment)
			break
		}
	}
	return resp
}

func postObserve(_ context.Context, cr *svcapitypes.TransitGatewayVPCAttachment, obj *svcsdk.DescribeTransitGatewayVpcAttachmentsOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	switch awsclients.StringValue(obj.TransitGatewayVpcAttachments[0].State) {
	case string(svcapitypes.TransitGatewayAttachmentState_available):
		cr.SetConditions(xpv1.Available())
	case string(svcapitypes.TransitGatewayAttachmentState_pending):
		cr.SetConditions(xpv1.Creating())
	case string(svcapitypes.TransitGatewayAttachmentState_modifying):
		cr.SetConditions(xpv1.Unavailable())
	case string(svcapitypes.TransitGatewayAttachmentState_deleting):
		cr.SetConditions(xpv1.Deleting())
	case string(svcapitypes.TransitGatewayAttachmentState_deleted):
		// TransitGatewayAttachment is in status deleted - and really removed after 6 hours in aws
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	return obs, nil
}

type custom struct {
	kube   client.Client
	client svcsdkapi.EC2API
}

func (e *custom) preCreate(ctx context.Context, cr *svcapitypes.TransitGatewayVPCAttachment, obj *svcsdk.CreateTransitGatewayVpcAttachmentInput) error {
	// need extra call for error:
	// cannot create TransitGatewayVPCAttachment in AWS: IncorrectState: tgw is in invalid state
	input := &svcsdk.DescribeTransitGatewaysInput{}
	input.TransitGatewayIds = append(input.TransitGatewayIds, cr.Spec.ForProvider.TransitGatewayID)
	tgwState, err := e.client.DescribeTransitGatewaysWithContext(ctx, input)
	if err != nil {
		return err
	}

	if awsclients.StringValue(tgwState.TransitGateways[0].State) != string(svcapitypes.TransitGatewayState_available) {
		return errors.New("referenced transitgateway is not available for vpcattachment " + awsclients.StringValue(tgwState.TransitGateways[0].State))
	}

	obj.VpcId = cr.Spec.ForProvider.VPCID
	obj.TransitGatewayId = cr.Spec.ForProvider.TransitGatewayID
	obj.SubnetIds = cr.Spec.ForProvider.SubnetIDs

	return nil
}

func postCreate(ctx context.Context, cr *svcapitypes.TransitGatewayVPCAttachment, obj *svcsdk.CreateTransitGatewayVpcAttachmentOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	// set transitgatewayvpcattachment id as external name annotation on k8s object after creation
	meta.SetExternalName(cr, aws.StringValue(obj.TransitGatewayVpcAttachment.TransitGatewayAttachmentId))
	return cre, nil
}

type tagger struct {
	kube client.Client
}

func (t *tagger) Initialize(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*svcapitypes.TransitGatewayVPCAttachment)
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	var transitGatewayAttachmentTags svcapitypes.TagSpecification
	for _, tagSpecification := range cr.Spec.ForProvider.TagSpecifications {
		if aws.StringValue(tagSpecification.ResourceType) == "transit-gateway-attachment" {
			transitGatewayAttachmentTags = *tagSpecification
		}
	}

	tagMap := map[string]string{}
	tagMap["Name"] = cr.Name
	for _, t := range transitGatewayAttachmentTags.Tags {
		tagMap[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
	}
	for k, v := range resource.GetExternalTags(mgd) {
		tagMap[k] = v
	}
	transitGatewayAttachmentTags.Tags = make([]*svcapitypes.Tag, len(tagMap))
	transitGatewayAttachmentTags.ResourceType = aws.String("transit-gateway-attachment")
	i := 0
	for k, v := range tagMap {
		transitGatewayAttachmentTags.Tags[i] = &svcapitypes.Tag{Key: aws.String(k), Value: aws.String(v)}
		i++
	}
	sort.Slice(transitGatewayAttachmentTags.Tags, func(i, j int) bool {
		return aws.StringValue(transitGatewayAttachmentTags.Tags[i].Key) < aws.StringValue(transitGatewayAttachmentTags.Tags[j].Key)
	})

	cr.Spec.ForProvider.TagSpecifications = []*svcapitypes.TagSpecification{&transitGatewayAttachmentTags}
	return errors.Wrap(t.kube.Update(ctx, cr), errKubeUpdateFailed)
}
