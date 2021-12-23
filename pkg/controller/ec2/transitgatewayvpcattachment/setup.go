package transitgatewayvpcattachment

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/ec2"
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
			e.postObserve = postObserve
			e.postCreate = postCreate
			e.preCreate = preCreate
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
			managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient()), &tagger{kube: mgr.GetClient()}),
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

func preCreate(ctx context.Context, cr *svcapitypes.TransitGatewayVPCAttachment, obj *svcsdk.CreateTransitGatewayVpcAttachmentInput) error {
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
