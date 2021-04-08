package dbparametergroup

import (
	"context"

	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	svcsdk "github.com/aws/aws-sdk-go/service/rds"
	svcsdkapi "github.com/aws/aws-sdk-go/service/rds/rdsiface"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

	svcapitypes "github.com/crossplane/provider-aws/apis/rds/v1alpha1"

	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// SetupDBParameterGroup adds a controller that reconciles DBParametergroup.
func SetupDBParameterGroup(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(svcapitypes.DBParameterGroupGroupKind)
	opts := []option{
		func(e *external) {
			e.preCreate = preCreate
			e.preObserve = preObserve
			e.preUpdate = preUpdate
			e.preDelete = preDelete
			e.postObserve = postObserve
			c := &custom{client: e.client, kube: e.kube}
			e.isUpToDate = c.isUpToDate
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&svcapitypes.DBParameterGroup{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.DBParameterGroupGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type custom struct {
	kube   client.Client
	client svcsdkapi.RDSAPI
}

func preObserve(_ context.Context, cr *svcapitypes.DBParameterGroup, obj *svcsdk.DescribeDBParameterGroupsInput) error {
	obj.DBParameterGroupName = awsclients.String(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.DBParameterGroup, obj *svcsdk.DescribeDBParameterGroupsOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(xpv1.Available())
	return obs, err
}

func preCreate(_ context.Context, cr *svcapitypes.DBParameterGroup, obj *svcsdk.CreateDBParameterGroupInput) error {
	obj.DBParameterGroupName = awsclients.String(meta.GetExternalName(cr))
	return nil
}

func preUpdate(_ context.Context, cr *svcapitypes.DBParameterGroup, obj *svcsdk.ModifyDBParameterGroupInput) error {
	obj.DBParameterGroupName = awsclients.String(meta.GetExternalName(cr))
	obj.Parameters = make([]*svcsdk.Parameter, len(cr.Spec.ForProvider.Parameters))

	for i, v := range cr.Spec.ForProvider.Parameters {
		// check if mandatory parameters are set (ApplyMethod, ParameterName, ParameterValue)
		if (v.ApplyMethod == nil) || (v.ParameterName == nil) || (v.ParameterValue == nil) {
			return errors.New("ApplyMethod, ParameterName and ParameterValue are mandatory fields and can not be nil")
		}
		obj.Parameters[i] = &svcsdk.Parameter{
			AllowedValues:        v.AllowedValues,
			ApplyMethod:          v.ApplyMethod,
			ApplyType:            v.ApplyType,
			DataType:             v.DataType,
			Description:          v.Description,
			IsModifiable:         v.IsModifiable,
			MinimumEngineVersion: v.MinimumEngineVersion,
			ParameterName:        v.ParameterName,
			ParameterValue:       v.ParameterValue,
			Source:               v.Source,
			SupportedEngineModes: v.SupportedEngineModes,
		}
	}
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.DBParameterGroup, obj *svcsdk.DeleteDBParameterGroupInput) (bool, error) {
	obj.DBParameterGroupName = awsclients.String(meta.GetExternalName(cr))
	return false, nil
}

func (e *custom) isUpToDate(cr *svcapitypes.DBParameterGroup, obj *svcsdk.DescribeDBParameterGroupsOutput) (bool, error) {
	// TODO(Dkaykay): We need isUpToDate to have context.
	ctx := context.TODO()
	results, err := e.getCurrentDBParameters(ctx, cr)
	if err != nil {
		return false, err
	}
	observed := make(map[string]svcsdk.Parameter, len(results))
	for _, p := range results {
		observed[awsclients.StringValue(p.ParameterName)] = *p
	}
	// compare CR with currently set Parameters
	for _, v := range cr.Spec.ForProvider.Parameters {
		existing, ok := observed[awsclients.StringValue(v.ParameterName)]
		if !ok {
			return false, nil
		}
		switch {
		case awsclients.StringValue(existing.ParameterValue) != awsclients.StringValue(v.ParameterValue):
			return false, nil
		case awsclients.StringValue(existing.ApplyMethod) != awsclients.StringValue(v.ApplyMethod):
			return false, nil
		}
	}
	return true, err
}

func (e *custom) getCurrentDBParameters(ctx context.Context, cr *svcapitypes.DBParameterGroup) ([]*svcsdk.Parameter, error) {
	input := &svcsdk.DescribeDBParametersInput{
		DBParameterGroupName: awsclients.String(meta.GetExternalName(cr)),
		MaxRecords:           awsclients.Int64(20),
	}
	var results []*svcsdk.Parameter
	err := e.client.DescribeDBParametersPagesWithContext(ctx, input, func(page *svcsdk.DescribeDBParametersOutput, lastPage bool) bool {
		results = append(results, page.Parameters...)
		return !lastPage
	})
	if err != nil {
		return results, err
	}
	return results, nil
}
