package dbparametergroup

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/rds"
	svcsdkapi "github.com/aws/aws-sdk-go/service/rds/rdsiface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/rds/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	svcutils "github.com/crossplane-contrib/provider-aws/pkg/controller/rds/utils"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	maxParametersPerUpdate = 20
)

const (
	errRequireDBParameterGroupFamilyOrFromEngine = "either spec.forProvider.dbParameterGroupFamily or spec.forProvider.dbParameterGroupFamilyFromEngine is required"
	errDetermineDBParameterGroupFamily           = "cannot determine DB parametergroup family"
	errGetDBEngineVersion                        = "cannot decsribe DB engine versions"
	errNoDBEngineVersions                        = "no DB engine versions returned by AWS"
	errCompareTags                               = "cannot compare tags"
	errAddTags                                   = "cannot add tags"
	errRemoveTags                                = "cannot remove tags"
)

// SetupDBParameterGroup adds a controller that reconciles DBParametergroup.
func SetupDBParameterGroup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.DBParameterGroupGroupKind)
	opts := []option{
		func(e *external) {
			c := &custom{client: e.client, kube: e.kube}
			e.preCreate = c.preCreate
			e.preObserve = preObserve
			e.preUpdate = c.preUpdate
			e.postUpdate = c.postUpdate
			e.preDelete = preDelete
			e.postObserve = postObserve
			e.lateInitialize = lateInitialize
			e.isUpToDate = c.isUpToDate
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
		resource.ManagedKind(svcapitypes.DBParameterGroupGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.DBParameterGroup{}).
		Complete(r)
}

type custom struct {
	kube   client.Client
	client svcsdkapi.RDSAPI

	cache struct {
		addTags    []*svcsdk.Tag
		removeTags []*string
	}
}

func preObserve(_ context.Context, cr *svcapitypes.DBParameterGroup, obj *svcsdk.DescribeDBParameterGroupsInput) error {
	obj.DBParameterGroupName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.DBParameterGroup, obj *svcsdk.DescribeDBParameterGroupsOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(xpv1.Available())
	return obs, err
}

func (c *custom) preCreate(ctx context.Context, cr *svcapitypes.DBParameterGroup, obj *svcsdk.CreateDBParameterGroupInput) error {
	if err := c.ensureParameterGroupFamily(ctx, cr); err != nil {
		return errors.Wrap(err, errDetermineDBParameterGroupFamily)
	}
	obj.DBParameterGroupName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	obj.DBParameterGroupFamily = cr.Spec.ForProvider.DBParameterGroupFamily
	return nil
}

func (c *custom) preUpdate(ctx context.Context, cr *svcapitypes.DBParameterGroup, obj *svcsdk.ModifyDBParameterGroupInput) error {
	obj.DBParameterGroupName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	currentParameters, err := c.getCurrentDBParameters(ctx, cr)

	if err != nil {
		return err
	}

	// The update call will not handle any removed parameters, this ensures
	// any removed parameters will be reset to default values
	parametersToReset := c.parametersToReset(cr, currentParameters)
	if len(parametersToReset) > 0 {
		if _, err := c.client.ResetDBParameterGroupWithContext(ctx, &svcsdk.ResetDBParameterGroupInput{
			DBParameterGroupName: obj.DBParameterGroupName,
			ResetAllParameters:   pointer.ToOrNilIfZeroValue(false),
			Parameters:           parametersToReset,
		}); err != nil {
			return err
		}
	}

	// Only 20 parameters are allowed per update request
	// this ensures we will only include parameters that require an update.
	// Any additional parameters will be handled during the next reconciliation.
	parametersToUpdate := c.parametersToUpdate(cr, currentParameters)
	if len(parametersToUpdate) > maxParametersPerUpdate {
		obj.Parameters = make([]*svcsdk.Parameter, maxParametersPerUpdate)
	} else {
		obj.Parameters = make([]*svcsdk.Parameter, len(parametersToUpdate))
	}

	for i, v := range parametersToUpdate {
		// We have reached the maximum number of
		// parameters per update
		if i > (maxParametersPerUpdate - 1) {
			break
		}

		obj.Parameters[i] = &svcsdk.Parameter{
			ApplyMethod:    v.ApplyMethod,
			ParameterName:  v.ParameterName,
			ParameterValue: v.ParameterValue,
		}
	}
	return nil
}

func (c *custom) postUpdate(ctx context.Context, cr *svcapitypes.DBParameterGroup, _ *svcsdk.DBParameterGroupNameMessage, upd managed.ExternalUpdate, _ error) (managed.ExternalUpdate, error) {
	if len(c.cache.addTags) > 0 {
		_, err := c.client.AddTagsToResourceWithContext(ctx, &svcsdk.AddTagsToResourceInput{
			ResourceName: cr.Status.AtProvider.DBParameterGroupARN,
			Tags:         c.cache.addTags,
		})
		if err != nil {
			return upd, errors.Wrap(err, errAddTags)
		}
	}
	if len(c.cache.removeTags) > 0 {
		_, err := c.client.RemoveTagsFromResourceWithContext(ctx, &svcsdk.RemoveTagsFromResourceInput{
			ResourceName: cr.Status.AtProvider.DBParameterGroupARN,
			TagKeys:      c.cache.removeTags,
		})
		if err != nil {
			return upd, errors.Wrap(err, errRemoveTags)
		}
	}
	return upd, nil
}

func preDelete(_ context.Context, cr *svcapitypes.DBParameterGroup, obj *svcsdk.DeleteDBParameterGroupInput) (bool, error) {
	obj.DBParameterGroupName = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return false, nil
}

func lateInitialize(spec *svcapitypes.DBParameterGroupParameters, current *svcsdk.DescribeDBParameterGroupsOutput) error {
	// Len > 0 is ensured by the generated controller.
	obj := current.DBParameterGroups[0]
	spec.DBParameterGroupFamily = obj.DBParameterGroupFamily
	return nil
}

func (c *custom) isUpToDate(ctx context.Context, cr *svcapitypes.DBParameterGroup, obj *svcsdk.DescribeDBParameterGroupsOutput) (bool, string, error) {
	results, err := c.getCurrentDBParameters(ctx, cr)
	if err != nil {
		return false, "", err
	}

	if len(c.parametersToUpdate(cr, results)) != 0 || len(c.parametersToReset(cr, results)) != 0 {
		return false, "", nil
	}

	areTagsUpToDate, addTags, removeTags, err := svcutils.AreTagsUpToDate(ctx, c.client, cr.Spec.ForProvider.Tags, obj.DBParameterGroups[0].DBParameterGroupArn)
	c.cache.addTags = addTags
	c.cache.removeTags = removeTags
	if err != nil || !areTagsUpToDate {
		return false, "spec.forProvider.tags", errors.Wrap(err, errCompareTags)
	}

	return true, "", err
}

func (c *custom) getCurrentDBParameters(ctx context.Context, cr *svcapitypes.DBParameterGroup) ([]*svcsdk.Parameter, error) {
	input := &svcsdk.DescribeDBParametersInput{
		DBParameterGroupName: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
		MaxRecords:           pointer.ToIntAsInt64(100),
	}
	var results []*svcsdk.Parameter
	err := c.client.DescribeDBParametersPagesWithContext(ctx, input, func(page *svcsdk.DescribeDBParametersOutput, lastPage bool) bool {
		results = append(results, page.Parameters...)
		return !lastPage
	})
	if err != nil {
		return results, err
	}
	return results, nil
}

func (c *custom) ensureParameterGroupFamily(ctx context.Context, cr *svcapitypes.DBParameterGroup) error {
	if cr.Spec.ForProvider.DBParameterGroupFamily == nil {
		engineVersion, err := c.getDBEngineVersion(ctx, cr.Spec.ForProvider.DBParameterGroupFamilySelector)
		if err != nil {
			return errors.Wrap(err, errGetDBEngineVersion)
		}
		cr.Spec.ForProvider.DBParameterGroupFamily = engineVersion.DBParameterGroupFamily
	}
	return nil
}

func (c *custom) getDBEngineVersion(ctx context.Context, selector *svcapitypes.DBParameterGroupFamilyNameSelector) (*svcsdk.DBEngineVersion, error) {
	if selector == nil {
		return nil, errors.New(errRequireDBParameterGroupFamilyOrFromEngine)
	}

	resp, err := c.client.DescribeDBEngineVersionsWithContext(ctx, &svcsdk.DescribeDBEngineVersionsInput{
		Engine:        &selector.Engine,
		EngineVersion: selector.EngineVersion,
		DefaultOnly:   pointer.ToOrNilIfZeroValue(selector.EngineVersion == nil),
	})
	if err != nil {
		return nil, err
	}
	if resp.DBEngineVersions == nil || len(resp.DBEngineVersions) == 0 || resp.DBEngineVersions[0] == nil {
		return nil, errors.New(errNoDBEngineVersions)
	}
	return resp.DBEngineVersions[0], nil
}

func (c *custom) parametersToUpdate(cr *svcapitypes.DBParameterGroup, current []*svcsdk.Parameter) []svcapitypes.CustomParameter {
	var parameters []svcapitypes.CustomParameter
	observed := make(map[string]svcsdk.Parameter, len(current))

	for _, p := range current {
		observed[pointer.StringValue(p.ParameterName)] = *p
	}

	// compare CR with currently set Parameters
	for _, v := range cr.Spec.ForProvider.Parameters {
		existing, ok := observed[pointer.StringValue(v.ParameterName)]

		if !ok {
			parameters = append(parameters, v)
			continue
		}

		if pointer.StringValue(existing.ParameterValue) != pointer.StringValue(v.ParameterValue) {
			parameters = append(parameters, v)
		}
	}

	return parameters
}

func (c *custom) parametersToReset(cr *svcapitypes.DBParameterGroup, current []*svcsdk.Parameter) []*svcsdk.Parameter {
	var parameters []*svcsdk.Parameter
	set := make(map[string]svcapitypes.CustomParameter, len(cr.Spec.ForProvider.Parameters))

	for _, p := range cr.Spec.ForProvider.Parameters {
		set[pointer.StringValue(p.ParameterName)] = p
	}

	for _, v := range current {
		if pointer.StringValue(v.Source) != "user" {
			// The describe operation lists all possible parameters
			// and their values, we only want to reset the parameter if
			// it's been changed from the default
			continue
		}

		if _, exists := set[pointer.StringValue(v.ParameterName)]; !exists {
			parameter := *v
			parameters = append(parameters, &parameter)
		}
	}

	return parameters
}
