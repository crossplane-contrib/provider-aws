package launchtemplate

import (
	"context"
	"sort"

	svcsdk "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/ec2/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	aws "github.com/crossplane-contrib/provider-aws/pkg/clients"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
)

// ControllerName of this controller.
var ControllerName = managed.ControllerName(svcapitypes.LaunchTemplateGroupKind)

// SetupLaunchTemplate adds a controller that reconciles LaunchTemplate.
func SetupLaunchTemplate(mgr ctrl.Manager, o controller.Options) error {
	name := ControllerName
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.preUpdate = preUpdate
			e.preDelete = preDelete
			e.postCreate = postCreate
			e.postObserve = postObserve
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&svcapitypes.LaunchTemplate{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.LaunchTemplateGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient()), &tagger{kube: mgr.GetClient()}),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

func preObserve(_ context.Context, cr *svcapitypes.LaunchTemplate, obj *svcsdk.DescribeLaunchTemplatesInput) error {
	obj.LaunchTemplateNames = append(obj.LaunchTemplateNames, aws.String(meta.GetExternalName(cr)))
	return nil
}

func preUpdate(_ context.Context, cr *svcapitypes.LaunchTemplate, obj *svcsdk.ModifyLaunchTemplateInput) error {
	obj.LaunchTemplateName = aws.String(meta.GetExternalName(cr))
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.LaunchTemplate, obj *svcsdk.DeleteLaunchTemplateInput) (bool, error) {
	obj.LaunchTemplateName = aws.String(meta.GetExternalName(cr))
	return false, nil
}

func postCreate(_ context.Context, cr *svcapitypes.LaunchTemplate, resp *svcsdk.CreateLaunchTemplateOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, aws.StringValue(resp.LaunchTemplate.LaunchTemplateName))
	return cre, nil
}

func postObserve(_ context.Context, cr *svcapitypes.LaunchTemplate, _ *svcsdk.DescribeLaunchTemplatesOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(xpv1.Available())
	return obs, nil
}

const (
	errKubeUpdateFailed = "cannot update LaunchTemplate custom resource"
)

type tagger struct {
	kube client.Client
}

func (t *tagger) Initialize(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*svcapitypes.LaunchTemplate)
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	var launchTemplateTags svcapitypes.TagSpecification
	for _, tagSpecification := range cr.Spec.ForProvider.TagSpecifications {
		if aws.StringValue(tagSpecification.ResourceType) == "launch-template" {
			launchTemplateTags = *tagSpecification
		}
	}

	tagMap := map[string]string{}
	tagMap["Name"] = cr.Name
	for _, t := range launchTemplateTags.Tags {
		tagMap[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
	}
	for k, v := range resource.GetExternalTags(mgd) {
		tagMap[k] = v
	}
	launchTemplateTags.Tags = make([]*svcapitypes.Tag, len(tagMap))
	launchTemplateTags.ResourceType = aws.String("launch-template")
	i := 0
	for k, v := range tagMap {
		launchTemplateTags.Tags[i] = &svcapitypes.Tag{Key: aws.String(k), Value: aws.String(v)}
		i++
	}
	sort.Slice(launchTemplateTags.Tags, func(i, j int) bool {
		return aws.StringValue(launchTemplateTags.Tags[i].Key) < aws.StringValue(launchTemplateTags.Tags[j].Key)
	})

	cr.Spec.ForProvider.TagSpecifications = []*svcapitypes.TagSpecification{&launchTemplateTags}
	return errors.Wrap(t.kube.Update(ctx, cr), errKubeUpdateFailed)
}
