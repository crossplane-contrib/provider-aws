package vpcpeeringconnection

import (
	"context"

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

	svcapitypes "github.com/crossplane/provider-aws/apis/ec2/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"

	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

// SetupVPCPeeringConnection adds a controller that reconciles VPCPeeringConnection.
func SetupVPCPeeringConnection(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(svcapitypes.VPCPeeringConnectionGroupKind)
	opts := []option{
		func(e *external) {
			c := &custom{client: e.client, kube: e.kube}
			e.postObserve = c.postObserve
			e.preObserve = preObserve
			e.postCreate = c.postCreate
			e.preCreate = preCreate
			e.isUpToDate = c.isUpToDate
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&svcapitypes.VPCPeeringConnection{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.VPCPeeringConnectionGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type custom struct {
	kube   client.Client
	client svcsdkapi.EC2API
}

func preObserve(ctx context.Context, cr *svcapitypes.VPCPeeringConnection, obj *svcsdk.DescribeVpcPeeringConnectionsInput) error {
	filterName := "tag:crossplane-external-name"
	externalName := meta.GetExternalName(cr)
	filterValue := []*string{&externalName}
	filter := svcsdk.Filter{
		Name:   &filterName,
		Values: filterValue,
	}
	filters := []*svcsdk.Filter{&filter}
	obj.SetFilters(filters)
	return nil
}

func (e *custom) postObserve(_ context.Context, cr *svcapitypes.VPCPeeringConnection, obj *svcsdk.DescribeVpcPeeringConnectionsOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	connectionCounter := -1
	for _, v := range obj.VpcPeeringConnections {
		connectionCounter++
		for _, tag := range v.Tags {
			if awsclients.StringValue(tag.Key) == "crossplane-external-name" {
				if *v.Status.Code == "pending-acceptance" && cr.Spec.ForProvider.AcceptRequest {
					// if acceptRequest is true, we automatically accept the request on AWS
					req := svcsdk.AcceptVpcPeeringConnectionInput{
						VpcPeeringConnectionId: awsclients.String(*v.VpcPeeringConnectionId),
					}
					request, _ := e.client.AcceptVpcPeeringConnectionRequest(&req)
					err := request.Send()
					if err != nil {
						return obs, err
					}
				}
				if awsclients.StringValue(tag.Value) == cr.ObjectMeta.Name {
					available := setCondition(v.Status, cr)
					if !available {
						return managed.ExternalObservation{ResourceExists: false}, nil
					}
				}

			}
		}
	}
	cr.SetConditions(xpv1.Available())
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
	}
	return false
}

func (e *custom) isUpToDate(cr *svcapitypes.VPCPeeringConnection, obj *svcsdk.DescribeVpcPeeringConnectionsOutput) (bool, error) {
	connectionCounter := -1

	for _, v := range obj.VpcPeeringConnections {
		connectionCounter++

		for _, tag := range v.Tags {
			if awsclients.StringValue(tag.Key) == "crossplane-external-name" {
				if awsclients.StringValue(tag.Value) == cr.ObjectMeta.Name {
					switch *v.Status.Code {
					case "active":
						return true, nil
					case "deleted":
						return false, nil
					}
				}
			}
		}
	}
	return false, nil
}

func preCreate(ctx context.Context, cr *svcapitypes.VPCPeeringConnection, obj *svcsdk.CreateVpcPeeringConnectionInput) error {
	// set external name as tag on the vpc peering connection
	resType := "vpc-peering-connection"
	key := "crossplane-external-name"
	value := meta.GetExternalName(cr)

	spec := svcsdk.TagSpecification{
		ResourceType: &resType,
		Tags: []*svcsdk.Tag{
			{
				Key:   &key,
				Value: &value,
			},
		},
	}
	obj.TagSpecifications = append(obj.TagSpecifications, &spec)

	return nil
}

func (e *custom) postCreate(ctx context.Context, cr *svcapitypes.VPCPeeringConnection, obj *svcsdk.CreateVpcPeeringConnectionOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	// set peering connection id as label on the k8s object after creation
	value := obj.VpcPeeringConnection.VpcPeeringConnectionId

	labels := make(map[string]string)
	labels["peering-connection-id"] = *value
	cr.SetLabels(labels)
	updateError := e.kube.Update(ctx, cr)
	if updateError != nil {
		return cre, updateError
	}

	return cre, err
}
