package optiongroup

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/rds/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	svcutils "github.com/crossplane-contrib/provider-aws/pkg/controller/rds/utils"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupOptionGroup adds a controller that reconciles OptionGroup.
func SetupOptionGroup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.OptionGroupGroupKind)
	opts := []option{
		func(e *external) {
			h := &hooks{client: e.client, kube: e.kube}
			e.preObserve = preObserve
			e.preCreate = preCreate
			e.preDelete = preDelete
			e.filterList = filterList
			e.postObserve = postObserve
			e.isUpToDate = h.isUpToDate
			e.preUpdate = h.preUpdate
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
		resource.ManagedKind(svcapitypes.OptionGroupGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.OptionGroup{}).
		Complete(r)
}

type hooks struct {
	client rdsiface.RDSAPI
	kube   client.Client
}

func preCreate(_ context.Context, cr *svcapitypes.OptionGroup, obj *svcsdk.CreateOptionGroupInput) error {
	obj.OptionGroupName = aws.String(meta.GetExternalName(cr))
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.OptionGroup, obj *svcsdk.DeleteOptionGroupInput) (bool, error) {
	obj.OptionGroupName = aws.String(meta.GetExternalName(cr))
	return false, nil
}

func preObserve(_ context.Context, cr *svcapitypes.OptionGroup, obj *svcsdk.DescribeOptionGroupsInput) error {
	obj.OptionGroupName = aws.String(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.OptionGroup, resp *svcsdk.DescribeOptionGroupsOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(xpv1.Available())
	return obs, nil
}

func filterList(cr *svcapitypes.OptionGroup, obj *svcsdk.DescribeOptionGroupsOutput) *svcsdk.DescribeOptionGroupsOutput {
	resp := &svcsdk.DescribeOptionGroupsOutput{}
	for _, optionGroup := range obj.OptionGroupsList {
		if aws.StringValue(optionGroup.OptionGroupName) == meta.GetExternalName(cr) {
			resp.OptionGroupsList = append(resp.OptionGroupsList, optionGroup)
			break
		}
	}
	return resp
}

func (e *hooks) isUpToDate(ctx context.Context, cr *svcapitypes.OptionGroup, obj *svcsdk.DescribeOptionGroupsOutput) (bool, string, error) {

	if aws.StringValue(cr.Spec.ForProvider.OptionGroupDescription) != aws.StringValue(obj.OptionGroupsList[0].OptionGroupDescription) {
		return false, "", nil
	}

	if aws.StringValue(cr.Spec.ForProvider.MajorEngineVersion) != aws.StringValue(obj.OptionGroupsList[0].MajorEngineVersion) {
		return false, "", nil
	}

	createOption, deleteOption := diffOptions(cr.Spec.ForProvider.Option, obj.OptionGroupsList[0].Options)
	if len(createOption) != 0 || len(deleteOption) != 0 {
		return false, "", nil
	}

	// for tagging: at least one option must be added, modified, or removed.
	tagsUpToDate, _, _, _ := svcutils.AreTagsUpToDate(ctx, e.client, cr.Spec.ForProvider.Tags, cr.Status.AtProvider.OptionGroupARN)
	if !tagsUpToDate {
		err := svcutils.UpdateTagsForResource(ctx, e.client, cr.Spec.ForProvider.Tags, cr.Status.AtProvider.OptionGroupARN)
		if err != nil {
			return true, "", errorutils.Wrap(err, errDescribe)
		}
	}

	return true, "", nil
}

func (e *hooks) preUpdate(ctx context.Context, cr *svcapitypes.OptionGroup, obj *svcsdk.ModifyOptionGroupInput) error {

	describe, err := e.client.DescribeOptionGroupsWithContext(ctx, &svcsdk.DescribeOptionGroupsInput{
		OptionGroupName: aws.String(meta.GetExternalName(cr)),
	})
	if err != nil {
		return errorutils.Wrap(err, errDescribe)
	}

	optionsToAdd, optionsToRemove := diffOptions(cr.Spec.ForProvider.Option, describe.OptionGroupsList[0].Options)
	if len(optionsToRemove) > 0 {
		obj.OptionsToRemove = optionsToRemove
	}

	if len(optionsToAdd) > 0 {
		optionsToInclude := []*svcsdk.OptionConfiguration{}
		for _, option := range cr.Spec.ForProvider.Option {
			include := &svcsdk.OptionConfiguration{
				OptionName:    option.OptionName,
				OptionVersion: option.OptionVersion,
				Port:          option.Port,
			}

			for _, optionSettings := range option.OptionSettings {
				optionSetting := &svcsdk.OptionSetting{
					Name:  optionSettings.Name,
					Value: optionSettings.Value,
				}
				include.OptionSettings = append(include.OptionSettings, optionSetting)
			}

			include.DBSecurityGroupMemberships = option.DBSecurityGroupMemberships
			include.VpcSecurityGroupMemberships = option.VPCSecurityGroupMemberships
			optionsToInclude = append(optionsToInclude, include)
		}
		obj.OptionsToInclude = optionsToInclude
	}

	obj.ApplyImmediately = cr.Spec.ForProvider.ApplyImmediately

	return nil
}

// diffOptions returns the lists of Options that need to be removed and added according
// to current and desired states.
func diffOptions(local []*svcapitypes.CustomOptionConfiguration, remote []*svcsdk.Option) ([]*string, []*string) {
	createOption := []*string{}
	deleteOption := []*string{}
	m := map[string]int{}

	for _, value := range local {
		m[*value.OptionName] = 1
	}

	for _, value := range remote {
		m[*value.OptionName] += 2
	}

	for mKey, mVal := range m {
		// need for scopelint
		mKey2 := mKey
		if mVal == 1 {
			createOption = append(createOption, &mKey2)
		}

		if mVal == 2 {
			deleteOption = append(deleteOption, &mKey2)
		}
	}
	return createOption, deleteOption
}
