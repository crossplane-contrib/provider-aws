package vpcpeeringconnection

import (
	"context"
	"reflect"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/ec2"
	svcsdkapi "github.com/aws/aws-sdk-go/service/ec2/ec2iface"
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
	errKubeUpdateFailed = "cannot update VPCPeeringConnection"
)

// SetupVPCPeeringConnection adds a controller that reconciles VPCPeeringConnection.
func SetupVPCPeeringConnection(mgr ctrl.Manager, o controller.Options) error {
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

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
		managed.WithCreationGracePeriod(3 * time.Minute),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithInitializers(&tagger{kube: mgr.GetClient()}),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(svcapitypes.VPCPeeringConnectionGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.VPCPeeringConnection{}).
		Complete(r)
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

func (e *custom) postObserve(ctx context.Context, cr *svcapitypes.VPCPeeringConnection, obj *svcsdk.DescribeVpcPeeringConnectionsOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) { // nolint:gocyclo
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	// The accept and modify operations for the Peer VPC have to be executed in the PeerRegion
	var pc svcsdkapi.EC2API
	if *cr.Spec.ForProvider.PeerRegion != cr.Spec.ForProvider.Region {
		sess, err := awsclients.GetConfigV1(ctx, e.kube, cr, *cr.Spec.ForProvider.PeerRegion)
		if err != nil {
			return obs, errors.Wrap(err, errCreateSession)
		}
		pc = svcsdk.New(sess)
	} else {
		pc = e.client
	}

	if awsclients.StringValue(obj.VpcPeeringConnections[0].Status.Code) == "pending-acceptance" && cr.Spec.ForProvider.AcceptRequest && !meta.WasDeleted(cr) {
		req := svcsdk.AcceptVpcPeeringConnectionInput{
			VpcPeeringConnectionId: awsclients.String(*obj.VpcPeeringConnections[0].VpcPeeringConnectionId),
		}
		request, _ := pc.AcceptVpcPeeringConnectionRequest(&req)
		err = request.Send()
		if err != nil {
			return obs, err
		}
	}

	if meta.WasDeleted(cr) && awsclients.StringValue(obj.VpcPeeringConnections[0].Status.Code) == "deleted" {
		return managed.ExternalObservation{
			ResourceExists:   false,
			ResourceUpToDate: false,
		}, nil
	}

	if awsclients.StringValue(obj.VpcPeeringConnections[0].Status.Code) == "active" {
		if !reflect.DeepEqual(obj.VpcPeeringConnections[0].AccepterVpcInfo.PeeringOptions, cr.Spec.ForProvider.AccepterPeeringOptions) ||
			!reflect.DeepEqual(obj.VpcPeeringConnections[0].RequesterVpcInfo.PeeringOptions, cr.Spec.ForProvider.RequesterPeeringOptions) {
			req := svcsdk.ModifyVpcPeeringConnectionOptionsInput{
				VpcPeeringConnectionId: awsclients.String(*obj.VpcPeeringConnections[0].VpcPeeringConnectionId),
			}
			if *cr.Spec.ForProvider.PeerRegion == cr.Spec.ForProvider.Region {
				setAccepterRequester(&req, cr)
			} else {
				acc := svcsdk.ModifyVpcPeeringConnectionOptionsInput{
					VpcPeeringConnectionId: awsclients.String(*obj.VpcPeeringConnections[0].VpcPeeringConnectionId),
				}
				setAccepter(&acc, cr)
				request, _ := pc.ModifyVpcPeeringConnectionOptionsRequest(&acc)
				err := request.Send()
				if err != nil {
					return obs, err
				}
				setRequester(&req, cr)
			}

			request, _ := e.client.ModifyVpcPeeringConnectionOptionsRequest(&req)
			err := request.Send()
			if err != nil {
				return obs, err
			}
		}
	}

	available := setCondition(obj.VpcPeeringConnections[0].Status, cr)
	if !available {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	return obs, nil
}

func setAccepterRequester(req *svcsdk.ModifyVpcPeeringConnectionOptionsInput, cr *svcapitypes.VPCPeeringConnection) {
	setAccepter(req, cr)
	setRequester(req, cr)
}

func setAccepter(req *svcsdk.ModifyVpcPeeringConnectionOptionsInput, cr *svcapitypes.VPCPeeringConnection) {
	if cr.Spec.ForProvider.AccepterPeeringOptions != nil {
		if *cr.Spec.ForProvider.PeerRegion == cr.Spec.ForProvider.Region {
			req.AccepterPeeringConnectionOptions = &svcsdk.PeeringConnectionOptionsRequest{
				AllowDnsResolutionFromRemoteVpc:            cr.Spec.ForProvider.AccepterPeeringOptions.AllowDNSResolutionFromRemoteVPC,
				AllowEgressFromLocalClassicLinkToRemoteVpc: cr.Spec.ForProvider.AccepterPeeringOptions.AllowEgressFromLocalClassicLinkToRemoteVPC,
				AllowEgressFromLocalVpcToRemoteClassicLink: cr.Spec.ForProvider.AccepterPeeringOptions.AllowEgressFromLocalVPCToRemoteClassicLink,
			}
		} else {
			req.AccepterPeeringConnectionOptions = &svcsdk.PeeringConnectionOptionsRequest{
				AllowDnsResolutionFromRemoteVpc: cr.Spec.ForProvider.AccepterPeeringOptions.AllowDNSResolutionFromRemoteVPC,
			}
		}
	}
}
func setRequester(req *svcsdk.ModifyVpcPeeringConnectionOptionsInput, cr *svcapitypes.VPCPeeringConnection) {
	if cr.Spec.ForProvider.RequesterPeeringOptions != nil {
		if *cr.Spec.ForProvider.PeerRegion == cr.Spec.ForProvider.Region {
			req.RequesterPeeringConnectionOptions = &svcsdk.PeeringConnectionOptionsRequest{
				AllowDnsResolutionFromRemoteVpc:            cr.Spec.ForProvider.RequesterPeeringOptions.AllowDNSResolutionFromRemoteVPC,
				AllowEgressFromLocalClassicLinkToRemoteVpc: cr.Spec.ForProvider.RequesterPeeringOptions.AllowEgressFromLocalClassicLinkToRemoteVPC,
				AllowEgressFromLocalVpcToRemoteClassicLink: cr.Spec.ForProvider.RequesterPeeringOptions.AllowEgressFromLocalVPCToRemoteClassicLink,
			}
		} else {
			req.RequesterPeeringConnectionOptions = &svcsdk.PeeringConnectionOptionsRequest{
				AllowDnsResolutionFromRemoteVpc: cr.Spec.ForProvider.RequesterPeeringOptions.AllowDNSResolutionFromRemoteVPC,
			}
		}
	}
}

func setCondition(code *svcsdk.VpcPeeringConnectionStateReason, cr *svcapitypes.VPCPeeringConnection) bool {
	switch aws.StringValue(code.Code) {
	case string(svcapitypes.VPCPeeringConnectionStateReasonCode_pending_acceptance), string(svcapitypes.VPCPeeringConnectionStateReasonCode_provisioning):
		cr.SetConditions(xpv1.Creating())
		return true
	case string(svcapitypes.VPCPeeringConnectionStateReasonCode_deleted), string(svcapitypes.VPCPeeringConnectionStateReasonCode_deleting), string(svcapitypes.VPCPeeringConnectionStateReasonCode_failed), string(svcapitypes.VPCPeeringConnectionStateReasonCode_rejected), string(svcapitypes.VPCPeeringConnectionStateReasonCode_expired):
		cr.SetConditions(xpv1.Unavailable())
		return false
	case string(svcapitypes.VPCPeeringConnectionStateReasonCode_active):
		cr.SetConditions(xpv1.Available())
		return true
	}
	return false
}

func (e *custom) isUpToDate(_ *svcapitypes.VPCPeeringConnection, _ *svcsdk.DescribeVpcPeeringConnectionsOutput) (bool, error) {
	return true, nil
}

func preCreate(_ context.Context, cr *svcapitypes.VPCPeeringConnection, obj *svcsdk.CreateVpcPeeringConnectionInput) error {
	obj.PeerVpcId = cr.Spec.ForProvider.PeerVPCID
	obj.VpcId = cr.Spec.ForProvider.VPCID

	return nil
}

func (e *custom) postCreate(_ context.Context, cr *svcapitypes.VPCPeeringConnection, obj *svcsdk.CreateVpcPeeringConnectionOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
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
	for _, t := range cr.Spec.ForProvider.Tags {
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
