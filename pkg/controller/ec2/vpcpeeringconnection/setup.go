package vpcpeeringconnection

import (
	"context"
	"log"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcsdk "github.com/aws/aws-sdk-go/service/ec2"
	svcapitypes "github.com/crossplane/provider-aws/apis/ec2/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"

	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

// SetupVPCPeeringConnection adds a controller that reconciles VPCPeeringConnection.
func SetupVPCPeeringConnection(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(svcapitypes.VPCPeeringConnectionGroupKind)
	opts := []option{
		func(e *external) {
			e.postCreate = postCreate
			e.preCreate = preCreate
			e.isUpToDate = isUpToDate
			//e.preObserve = preObserve
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

func isUpToDate(cr *svcapitypes.VPCPeeringConnection, obj *svcsdk.DescribeVpcPeeringConnectionsOutput) (bool, error) {
	return false, nil
}

func preCreate(_ context.Context, cr *svcapitypes.VPCPeeringConnection, obj *svcsdk.CreateVpcPeeringConnectionInput) error {
	resType := "vpc-peering-connection"
	key := "Name"
	value := meta.GetExternalName(cr)

	// set external name as tag
	spec := svcsdk.TagSpecification{
		ResourceType: &resType,
		Tags: []*svcsdk.Tag{
			&svcsdk.Tag{
				Key:   &key,
				Value: &value,
			},
		},
	}
	obj.TagSpecifications = append(obj.TagSpecifications, &spec)

	return nil
}

func postCreate(_ context.Context, cr *svcapitypes.VPCPeeringConnection, obj *svcsdk.CreateVpcPeeringConnectionOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	// after creation, we have to approve the VPCPeeringConnection
	accepted := svcsdk.AcceptVpcPeeringConnectionInput{
		VpcPeeringConnectionId: awsclients.String(*obj.VpcPeeringConnection.VpcPeeringConnectionId),
	}

	log.Println("Accepted:")
	log.Println(accepted)
	return cre, err
}

// unused atm
// func acceptRequest(cr *svcapitypes.VPCPeeringConnection, obj *svcsdk.DescribeVpcPeeringConnectionsOutput) {

// 	connectionCounter := -1
// 	for _, v := range obj.VpcPeeringConnections {
// 		connectionCounter++

// 		if v.AccepterVpcInfo.VpcId == cr.Spec.ForProvider.PeerVPCID {
// 			if v.RequesterVpcInfo.VpcId == cr.Spec.ForProvider.VPCID {
// 				log.Println("Match!")
// 				break
// 			}
// 		}
// 	}
// 	accepted := &svcsdk.AcceptVpcPeeringConnectionInput{
// 		VpcPeeringConnectionId: awsclients.String(*obj.VpcPeeringConnections[connectionCounter].VpcPeeringConnectionId),
// 	}

// 	log.Println(accepted)
// }
