package dbcluster

import (
	"context"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

	svcsdk "github.com/aws/aws-sdk-go/service/rds"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/rds/v1alpha1"
)

// SetupDbCluster adds a controller that reconciles DbCluster.
func SetupDbCluster(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(svcapitypes.DBClusterGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preCreate = preCreate
			e.preUpdate = preUpdate
			e.preDelete = preDelete
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&svcapitypes.DBCluster{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.DBClusterGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func preObserve(_ context.Context, cr *svcapitypes.DBCluster, obj *svcsdk.DescribeDBClustersInput) error {
	obj.DBClusterIdentifier = aws.String(meta.GetExternalName(cr))
	return nil
}

// this probably requires custom Conditions to be defined for handling all statuses
// described here https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/Aurora.Status.html
// Need to get help from community on how to deal with this.
func postObserve(_ context.Context, cr *svcapitypes.DBCluster, resp *svcsdk.DescribeDBClustersOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	switch aws.StringValue(resp.DBClusters[0].Status) {
	case "available":
		cr.SetConditions(xpv1.Available())
	case "deleting", "stopped", "stopping":
		cr.SetConditions(xpv1.Unavailable())
	case "creating":
		cr.SetConditions(xpv1.Creating())
	}
	return obs, nil
}

func preCreate(_ context.Context, cr *svcapitypes.DBCluster, obj *svcsdk.CreateDBClusterInput) error {
	obj.DBClusterIdentifier = aws.String(meta.GetExternalName(cr))
	return nil
}

func preUpdate(_ context.Context, cr *svcapitypes.DBCluster, obj *svcsdk.ModifyDBClusterInput) error {
	obj.DBClusterIdentifier = aws.String(meta.GetExternalName(cr))
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.DBCluster, obj *svcsdk.DeleteDBClusterInput) error {
	obj.DBClusterIdentifier = aws.String(meta.GetExternalName(cr))
	return nil
}
