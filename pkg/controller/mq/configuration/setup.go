package configuration

import (
	"context"
	"encoding/base64"

	svcsdk "github.com/aws/aws-sdk-go/service/mq"
	svcsdkapi "github.com/aws/aws-sdk-go/service/mq/mqiface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/mq/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	mqconfutils "github.com/crossplane-contrib/provider-aws/pkg/controller/mq/configuration/utils"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/tags"
)

func SetupConfiguration(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.ConfigurationGroupKind)
	opts := []option{
		func(e *external) {
			c := &custom{client: e.client, kube: e.kube, external: e}
			e.isUpToDate = c.isUpToDate
			e.postCreate = c.postCreate
			e.preObserve = preObserve
			e.postObserve = c.postObserve
			e.update = c.update
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}
	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithInitializers(),
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithTypedExternalConnector(&connector{kube: mgr.GetClient(), opts: opts}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(svcapitypes.ConfigurationGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.Configuration{}).
		Complete(r)
}

type custom struct {
	kube     client.Client
	client   svcsdkapi.MQAPI
	external *external
}

func preObserve(_ context.Context, cr *svcapitypes.Configuration, obj *svcsdk.DescribeConfigurationInput) error {
	obj.ConfigurationId = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

func (e *custom) postObserve(ctx context.Context, cr *svcapitypes.Configuration, obj *svcsdk.DescribeConfigurationOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	switch {
	case meta.WasDeleted(cr):
		obs.ResourceExists = false
	case mqconfutils.HasBeenSanitized(cr):
		cr.SetConditions(xpv1.Unavailable())
	default:
		cr.SetConditions(xpv1.Available())
	}

	return obs, nil
}

func (e *custom) postCreate(ctx context.Context, cr *svcapitypes.Configuration, obj *svcsdk.CreateConfigurationResponse, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, pointer.StringValue(obj.Id))
	return cre, nil
}

func (e *custom) isUpToDate(ctx context.Context, cr *svcapitypes.Configuration, describeConfigOutput *svcsdk.DescribeConfigurationOutput) (bool, string, error) {
	atProviderConfig, err := setData(ctx, e, cr, describeConfigOutput.Id, describeConfigOutput.LatestRevision.Revision)
	if err != nil {
		return false, "", err
	}
	setTags(cr, describeConfigOutput)
	lateInitialize(describeConfigOutput, &cr.Spec.ForProvider, atProviderConfig)
	// stops the update loop for the data field if the MQ Configuration sanitizes a revision
	// any additional pending updates are postponed until the sanitization warnings
	// in the error message are applied
	if mqconfutils.HasBeenSanitized(cr) {
		hasBeenUpdated := mqconfutils.HasBeenUpdatedPostSanitization(cr)
		return !hasBeenUpdated, "", nil
	}
	isRevisionUpToDate, err := isRevisionUpToDate(cr)
	if err != nil {
		return false, "", err
	}

	add, remove := tags.DiffTagsMapPtr(cr.Spec.ForProvider.Tags, describeConfigOutput.Tags)
	areTagsUpToDate := len(add) == 0 && len(remove) == 0

	return isRevisionUpToDate && areTagsUpToDate, "", nil
}

func (e *custom) update(ctx context.Context, cr *svcapitypes.Configuration) (managed.ExternalUpdate, error) {
	if cr.Status.AtProvider.ARN == nil {
		return managed.ExternalUpdate{}, nil
	}
	added, removed := tags.DiffTagsMapPtr(cr.Spec.ForProvider.Tags, cr.Status.AtProvider.Tags)
	if len(added) > 0 {
		_, err := e.client.CreateTagsWithContext(ctx, &svcsdk.CreateTagsInput{
			ResourceArn: cr.Status.AtProvider.ARN,
			Tags:        added,
		})
		if err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, ErrCannotAddTags)
		}
	}
	if len(removed) > 0 {
		_, err := e.client.DeleteTagsWithContext(ctx, &svcsdk.DeleteTagsInput{
			ResourceArn: cr.Status.AtProvider.ARN,
			TagKeys:     removed,
		})
		if err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, ErrCannotRemoveTags)
		}

	}
	// update configuration only if its latest revision is not up to date
	// a revision is up to date if both description and data are the same
	// sanitization without warnings of the data field may potentially result in a new revision
	// which can be halted in isUpToDate for the following observe
	isRevisionUpToDate, err := isRevisionUpToDate(cr)
	if err != nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(err, "cannot check if revision is up to date")
	}
	if isRevisionUpToDate {
		return managed.ExternalUpdate{}, nil
	}
	input := generateUpdateConfigurationRequest(cr)
	resp, err := e.client.UpdateConfigurationWithContext(ctx, input)
	return e.postUpdate(ctx, cr, resp, managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate))
}

func (e *custom) postUpdate(ctx context.Context, cr *svcapitypes.Configuration, obj *svcsdk.UpdateConfigurationResponse, cre managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	if err != nil {
		return cre, err
	}

	if len(obj.Warnings) == 0 && !mqconfutils.HasBeenSanitized(cr) {
		return cre, nil
	}

	currentObj := cr.DeepCopy()
	if len(obj.Warnings) == 0 {
		meta.RemoveAnnotations(cr, svcapitypes.LatestUnsanitizedConfiguration)
	} else {
		err = handleSanitizationWarnings(obj.Warnings)
		mqconfutils.SetLatestUnsanitizedConfiguration(cr)
	}
	patch := client.MergeFrom(currentObj)
	if err := e.kube.Patch(ctx, cr, patch); err != nil {
		return cre, err
	}
	return cre, err
}

func setData(ctx context.Context, e *custom, cr *svcapitypes.Configuration, id *string, revision *int64) (string, error) {
	describeConfigRevisionOutput, err := getLatestRevisionData(ctx, e, id, revision)
	if err != nil {
		return "", err
	}
	atProviderDataLatestRevision, err := base64.StdEncoding.DecodeString(pointer.StringValue(describeConfigRevisionOutput.Data))
	if err != nil {
		return "", err
	}
	atProviderConfig := string(atProviderDataLatestRevision)
	cr.Status.AtProvider.LatestRevisionData = &atProviderConfig
	return atProviderConfig, nil
}

func setTags(cr *svcapitypes.Configuration, describeConfigOutput *svcsdk.DescribeConfigurationOutput) {
	cr.Status.AtProvider.Tags = describeConfigOutput.Tags
}

func getLatestRevisionData(ctx context.Context, e *custom, id *string, revision *int64) (*svcsdk.DescribeConfigurationRevisionResponse, error) {
	describeRevisionInput := generateDescribeConfigurationRevisionInput(id, revision)
	describeConfigRevisionOutput, err := e.client.DescribeConfigurationRevisionWithContext(ctx, describeRevisionInput)
	if err != nil {
		return nil, err
	}
	return describeConfigRevisionOutput, nil
}

func lateInitialize(describeConfigOutput *svcsdk.DescribeConfigurationOutput, in *svcapitypes.ConfigurationParameters, out string) {
	if pointer.StringValue(in.Data) == "" {
		in.Data = pointer.ToOrNilIfZeroValue(out)
	}
	in.Description = pointer.LateInitialize(in.Description, describeConfigOutput.LatestRevision.Description)
}

func isRevisionUpToDate(cr *svcapitypes.Configuration) (bool, error) {
	forProviderConfig := pointer.StringValue(cr.Spec.ForProvider.Data)
	atProviderConfig := pointer.StringValue(cr.Status.AtProvider.LatestRevisionData)
	diffDataAtAndForProviderConfig, err := mqconfutils.DiffXMLConfigs(atProviderConfig, forProviderConfig)
	if err != nil {
		return false, err
	}

	isDataUpToDate := diffDataAtAndForProviderConfig == ""
	isDescriptionUpToDate := pointer.StringValue(cr.Spec.ForProvider.Description) == pointer.StringValue(cr.Status.AtProvider.LatestRevision.Description)
	return isDataUpToDate && isDescriptionUpToDate, nil
}
