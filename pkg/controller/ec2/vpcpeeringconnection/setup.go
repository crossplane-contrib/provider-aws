package vpcpeeringconnection

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
	errKubeUpdateFailed = "cannot update VPCPeeringConnection"
)

// SetupVPCPeeringConnection adds a controller that reconciles VPCPeeringConnection.
func SetupVPCPeeringConnection(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(svcapitypes.VPCPeeringConnectionGroupKind)
	opts := []option{
		func(e *external) {
			c := &custom{client: e.client, kube: e.kube}
			e.postObserve = c.postObserve
			e.postCreate = c.postCreate
			e.preCreate = preCreate
			e.isUpToDate = c.isUpToDate
			e.filterList = filterList
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewController(rl),
		}).
		For(&svcapitypes.VPCPeeringConnection{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.VPCPeeringConnectionGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithCreationGracePeriod(3*time.Minute),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithInitializers(&tagger{kube: mgr.GetClient()}),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type custom struct {
	kube   client.Client
	client svcsdkapi.EC2API
}

func filterList(cr *svcapitypes.VPCPeeringConnection, obj *svcsdk.DescribeVpcPeeringConnectionsOutput) *svcsdk.DescribeVpcPeeringConnectionsOutput {
	connectionIdentifier := aws.String(meta.GetExternalName(cr))
	resp := &svcsdk.DescribeVpcPeeringConnectionsOutput{}
	for _, vpcPeeringConnection := range obj.VpcPeeringConnections {
		if aws.StringValue(vpcPeeringConnection.VpcPeeringConnectionId) == aws.StringValue(connectionIdentifier) {
			resp.VpcPeeringConnections = append(resp.VpcPeeringConnections, vpcPeeringConnection)
			break
		}
	}
	return resp
}

func (e *custom) postObserve(_ context.Context, cr *svcapitypes.VPCPeeringConnection, obj *svcsdk.DescribeVpcPeeringConnectionsOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	if awsclients.StringValue(obj.VpcPeeringConnections[0].Status.Code) == "pending-acceptance" && cr.Spec.ForProvider.AcceptRequest {
		req := svcsdk.AcceptVpcPeeringConnectionInput{
			VpcPeeringConnectionId: awsclients.String(*obj.VpcPeeringConnections[0].VpcPeeringConnectionId),
		}
		request, _ := e.client.AcceptVpcPeeringConnectionRequest(&req)
		err := request.Send()
		if err != nil {
			return obs, err
		}
	}

	available := setCondition(obj.VpcPeeringConnections[0].Status, cr)
	if !available {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	return obs, nil
}

func setCondition(code *svcsdk.VpcPeeringConnectionStateReason, cr *svcapitypes.VPCPeeringConnection) bool {
	switch aws.StringValue(code.Code) {
	case string(svcapitypes.VPCPeeringConnectionStateReasonCode_pending_acceptance):
		cr.SetConditions(xpv1.Creating())
		return true
	case string(svcapitypes.VPCPeeringConnectionStateReasonCode_deleted):
		cr.SetConditions(xpv1.Unavailable())
		return false
	case string(svcapitypes.VPCPeeringConnectionStateReasonCode_active):
		cr.SetConditions(xpv1.Available())
		return true
	case string(svcapitypes.VPCPeeringConnectionStateReasonCode_failed):
		cr.SetConditions(xpv1.Unavailable())
		return false
	}
	return false
}

func (e *custom) isUpToDate(cr *svcapitypes.VPCPeeringConnection, obj *svcsdk.DescribeVpcPeeringConnectionsOutput) (bool, error) {
	return true, nil
}

func preCreate(ctx context.Context, cr *svcapitypes.VPCPeeringConnection, obj *svcsdk.CreateVpcPeeringConnectionInput) error {
	obj.PeerVpcId = cr.Spec.ForProvider.PeerVPCID
	obj.VpcId = cr.Spec.ForProvider.VPCID

	return nil
}

func (e *custom) postCreate(ctx context.Context, cr *svcapitypes.VPCPeeringConnection, obj *svcsdk.CreateVpcPeeringConnectionOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	// set peering connection id as external name annotation on k8s object after creation
	meta.SetExternalName(cr, aws.StringValue(obj.VpcPeeringConnection.VpcPeeringConnectionId))
	return cre, nil
}

type tagger struct {
	kube client.Client
}

func (t *tagger) Initialize(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*svcapitypes.VPCPeeringConnection)
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	var vpcPeeringConnectionTags svcapitypes.TagSpecification
	for _, tagSpecification := range cr.Spec.ForProvider.TagSpecifications {
		if aws.StringValue(tagSpecification.ResourceType) == "vpc-peering-connection" {
			vpcPeeringConnectionTags = *tagSpecification
		}
	}

	tagMap := map[string]string{}
	tagMap["Name"] = cr.Name
	for _, t := range vpcPeeringConnectionTags.Tags {
		tagMap[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
	}
	for k, v := range resource.GetExternalTags(mgd) {
		tagMap[k] = v
	}
	vpcPeeringConnectionTags.Tags = make([]*svcapitypes.Tag, len(tagMap))
	vpcPeeringConnectionTags.ResourceType = aws.String("vpc-peering-connection")
	i := 0
	for k, v := range tagMap {
		vpcPeeringConnectionTags.Tags[i] = &svcapitypes.Tag{Key: aws.String(k), Value: aws.String(v)}
		i++
	}
	sort.Slice(vpcPeeringConnectionTags.Tags, func(i, j int) bool {
		return aws.StringValue(vpcPeeringConnectionTags.Tags[i].Key) < aws.StringValue(vpcPeeringConnectionTags.Tags[j].Key)
	})

	cr.Spec.ForProvider.TagSpecifications = []*svcapitypes.TagSpecification{&vpcPeeringConnectionTags}
	return errors.Wrap(t.kube.Update(ctx, cr), errKubeUpdateFailed)
}
