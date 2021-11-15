package globalcluster

import (
	"context"
	"time"

	svcsdk "github.com/aws/aws-sdk-go/service/rds"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/rds/v1alpha1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
)

// SetupGlobalCluster adds a controller that reconciles GlobalCluster.
func SetupGlobalCluster(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(svcapitypes.GlobalClusterGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.preCreate = preCreate
			e.preDelete = preDelete
			e.filterList = filterList
			e.postObserve = postObserve
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewController(rl),
		}).
		For(&svcapitypes.GlobalCluster{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.GlobalClusterGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func preObserve(_ context.Context, cr *svcapitypes.GlobalCluster, obj *svcsdk.DescribeGlobalClustersInput) error {
	obj.GlobalClusterIdentifier = aws.String(meta.GetExternalName(cr))
	return nil
}

func preCreate(_ context.Context, cr *svcapitypes.GlobalCluster, obj *svcsdk.CreateGlobalClusterInput) error {
	obj.GlobalClusterIdentifier = aws.String(meta.GetExternalName(cr))
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.GlobalCluster, obj *svcsdk.DeleteGlobalClusterInput) (bool, error) {
	obj.GlobalClusterIdentifier = aws.String(meta.GetExternalName(cr))
	return false, nil
}

func postObserve(_ context.Context, cr *svcapitypes.GlobalCluster, resp *svcsdk.DescribeGlobalClustersOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	switch aws.StringValue(resp.GlobalClusters[0].Status) {
	case "available":
		cr.SetConditions(xpv1.Available())
	case "deleting", "stopped", "stopping":
		cr.SetConditions(xpv1.Unavailable())
	case "creating":
		cr.SetConditions(xpv1.Creating())
	}
	return obs, nil
}

func filterList(cr *svcapitypes.GlobalCluster, obj *svcsdk.DescribeGlobalClustersOutput) *svcsdk.DescribeGlobalClustersOutput {
	resp := &svcsdk.DescribeGlobalClustersOutput{}
	for _, dbCluster := range obj.GlobalClusters {
		if aws.StringValue(dbCluster.GlobalClusterIdentifier) == meta.GetExternalName(cr) {
			resp.GlobalClusters = append(resp.GlobalClusters, dbCluster)
			break
		}
	}
	return resp
}
