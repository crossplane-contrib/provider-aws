package dbcluster

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/rds"
	svcsdkapi "github.com/aws/aws-sdk-go/service/rds/rdsiface"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/rds/v1alpha1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/rds"
)

// SetupDBCluster adds a controller that reconciles DbCluster.
func SetupDBCluster(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.DBClusterGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.postObserve = postObserve
			c := &custom{client: e.client, kube: e.kube}
			e.isUpToDate = isUpToDate
			e.preUpdate = preUpdate
			e.preCreate = c.preCreate
			e.postCreate = c.postCreate
			e.preDelete = preDelete
			e.filterList = filterList
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&svcapitypes.DBCluster{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.DBClusterGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func preObserve(_ context.Context, cr *svcapitypes.DBCluster, obj *svcsdk.DescribeDBClustersInput) error {
	obj.DBClusterIdentifier = aws.String(meta.GetExternalName(cr))
	return nil
}

// This probably requires custom Conditions to be defined for handling all statuses
// described here https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/Aurora.Status.html
// Need to get help from community on how to deal with this. Ideally the status should reflect
// the true status value as described by the provider.
func postObserve(_ context.Context, cr *svcapitypes.DBCluster, resp *svcsdk.DescribeDBClustersOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	switch aws.StringValue(resp.DBClusters[0].Status) {
	case "available", "modifying":
		cr.SetConditions(xpv1.Available())
	case "deleting", "stopped", "stopping":
		cr.SetConditions(xpv1.Unavailable())
	case "creating":
		cr.SetConditions(xpv1.Creating())
	}
	return obs, nil
}

type custom struct {
	kube   client.Client
	client svcsdkapi.RDSAPI
}

func (e *custom) preCreate(ctx context.Context, cr *svcapitypes.DBCluster, obj *svcsdk.CreateDBClusterInput) error {
	pw, _, err := rds.GetPassword(ctx, e.kube, &cr.Spec.ForProvider.MasterUserPasswordSecretRef, cr.Spec.WriteConnectionSecretToReference)
	if err != nil {
		return errors.Wrap(err, "cannot get password from the given secret")
	}
	obj.MasterUserPassword = aws.String(pw)
	obj.DBClusterIdentifier = aws.String(meta.GetExternalName(cr))
	obj.VpcSecurityGroupIds = make([]*string, len(cr.Spec.ForProvider.VPCSecurityGroupIDs))
	for i, v := range cr.Spec.ForProvider.VPCSecurityGroupIDs {
		obj.VpcSecurityGroupIds[i] = aws.String(v)
	}
	return nil
}

func (e *custom) postCreate(ctx context.Context, cr *svcapitypes.DBCluster, _ *svcsdk.CreateDBClusterOutput, _ managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	conn := managed.ConnectionDetails{
		xpv1.ResourceCredentialsSecretEndpointKey: []byte(aws.StringValue(cr.Status.AtProvider.Endpoint)),
		xpv1.ResourceCredentialsSecretUserKey:     []byte(aws.StringValue(cr.Spec.ForProvider.MasterUsername)),
	}
	pw, _, err := rds.GetPassword(ctx, e.kube, &cr.Spec.ForProvider.MasterUserPasswordSecretRef, cr.Spec.WriteConnectionSecretToReference)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "cannot get password from the given secret")
	}
	if pw != "" {
		conn[xpv1.ResourceCredentialsSecretPasswordKey] = []byte(pw)
	}
	return managed.ExternalCreation{
		ConnectionDetails: conn,
	}, nil
}

func isUpToDate(cr *svcapitypes.DBCluster, out *svcsdk.DescribeDBClustersOutput) (bool, error) {
	status := aws.StringValue(out.DBClusters[0].Status)
	if status == "modifying" || status == "upgrading" || status == "configuring-iam-database-auth" {
		return true, nil
	}

	if aws.BoolValue(cr.Spec.ForProvider.EnableIAMDatabaseAuthentication) != aws.BoolValue(out.DBClusters[0].IAMDatabaseAuthenticationEnabled) {
		return false, nil
	}

	return true, nil
}

func preUpdate(_ context.Context, cr *svcapitypes.DBCluster, obj *svcsdk.ModifyDBClusterInput) error {
	obj.DBClusterIdentifier = aws.String(meta.GetExternalName(cr))
	obj.ApplyImmediately = cr.Spec.ForProvider.ApplyImmediately

	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.DBCluster, obj *svcsdk.DeleteDBClusterInput) (bool, error) {
	obj.DBClusterIdentifier = aws.String(meta.GetExternalName(cr))
	obj.FinalDBSnapshotIdentifier = aws.String(cr.Spec.ForProvider.FinalDBSnapshotIdentifier)
	obj.SkipFinalSnapshot = aws.Bool(cr.Spec.ForProvider.SkipFinalSnapshot)
	return false, nil
}

func filterList(cr *svcapitypes.DBCluster, obj *svcsdk.DescribeDBClustersOutput) *svcsdk.DescribeDBClustersOutput {
	clusterIdentifier := aws.String(meta.GetExternalName(cr))
	resp := &svcsdk.DescribeDBClustersOutput{}
	for _, dbCluster := range obj.DBClusters {
		if aws.StringValue(dbCluster.DBClusterIdentifier) == aws.StringValue(clusterIdentifier) {
			resp.DBClusters = append(resp.DBClusters, dbCluster)
			break
		}
	}
	return resp
}
