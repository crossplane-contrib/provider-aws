package globalcluster

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/rds"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/rds/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupGlobalCluster adds a controller that reconciles GlobalCluster.
func SetupGlobalCluster(mgr ctrl.Manager, o controller.Options) error {
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

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(svcapitypes.GlobalClusterGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.GlobalCluster{}).
		Complete(r)
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
