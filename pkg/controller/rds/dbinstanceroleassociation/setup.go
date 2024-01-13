package dbinstanceroleassociation

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/rds"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	cpresource "github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/rds/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errDescribeAssoc = "failed to describe DBInstance for DBInstanceRoleAssociation"
)

// SetupDBInstanceRoleAssociation adds a controller that reconciles DBInstanceRoleAssociation.
func SetupDBInstanceRoleAssociation(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.DBInstanceRoleAssociationGroupKind)
	opts := []option{
		func(e *external) {
			e.preCreate = preCreate
			e.observe = e.observer
			e.preDelete = preDelete
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
		cpresource.ManagedKind(svcapitypes.DBInstanceRoleAssociationGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&svcapitypes.DBInstanceRoleAssociation{}).
		WithEventFilter(cpresource.DesiredStateChanged()).
		Complete(r)
}

// GenerateDescribeDBInstancesInput returns the input for the read operation
func GenerateDescribeDBInstancesInput(cr *svcapitypes.DBInstanceRoleAssociation) *svcsdk.DescribeDBInstancesInput {
	return &svcsdk.DescribeDBInstancesInput{
		DBInstanceIdentifier: cr.Spec.ForProvider.DBInstanceIdentifier,
	}
}

func preCreate(_ context.Context, cr *svcapitypes.DBInstanceRoleAssociation, input *svcsdk.AddRoleToDBInstanceInput) error {
	input.DBInstanceIdentifier = cr.Spec.ForProvider.DBInstanceIdentifier
	input.RoleArn = cr.Spec.ForProvider.RoleARN
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.DBInstanceRoleAssociation, input *svcsdk.RemoveRoleFromDBInstanceInput) (bool, error) {
	input.DBInstanceIdentifier = cr.Spec.ForProvider.DBInstanceIdentifier
	input.RoleArn = cr.Spec.ForProvider.RoleARN
	return input.RoleArn == nil || input.DBInstanceIdentifier == nil, nil
}

func (e *external) observer(ctx context.Context, mg cpresource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*svcapitypes.DBInstanceRoleAssociation)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	input := GenerateDescribeDBInstancesInput(cr)
	resp, err := e.client.DescribeDBInstancesWithContext(ctx, input)
	if err != nil {
		return managed.ExternalObservation{ResourceExists: false}, errorutils.Wrap(cpresource.Ignore(IsNotFound, err), errDescribeAssoc)
	}
	if len(resp.DBInstances) == 0 {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	var status *string
	for _, role := range resp.DBInstances[0].AssociatedRoles {
		if pointer.StringValue(role.FeatureName) == pointer.StringValue(cr.Spec.ForProvider.FeatureName) && pointer.StringValue(role.RoleArn) == pointer.StringValue(cr.Spec.ForProvider.RoleARN) {
			status = role.Status
			break
		}
	}

	exists := status != nil

	if aws.StringValue(status) == "ACTIVE" {
		cr.SetConditions(xpv1.Available())
	} else {
		// At the moment we can't add custom atProvider fields to the
		// status, so we have to settle for the condition message
		cr.SetConditions(xpv1.Unavailable().WithMessage(aws.StringValue(status)))
	}

	// if roleArn is different, should we return exists but not up to
	// date? can we have two roles for one feature, probably not..

	return managed.ExternalObservation{
		ResourceExists:          exists,
		ResourceUpToDate:        true,
		ResourceLateInitialized: false,
	}, nil
}
