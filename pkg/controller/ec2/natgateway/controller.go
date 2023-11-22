package natgateway

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	awsec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
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

	"github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/ec2"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errUnexpectedObject = "The managed resource is not an NATGateway resource"
	errDescribe         = "failed to describe NATGateway"
	errNotSingleItem    = "either no or multiple NATGateways retrieved for the given natGatewayId"
	errCreate           = "failed to create the NATGateway resource"
	errDelete           = "failed to delete the NATGateway resource"
	errUpdateTags       = "failed to update tags for the NATGateway resource"
	errDeleteTags       = "failed to delete tags for NATGateway resource"
)

// SetupNatGateway adds a controller that reconciles NatGateways.
func SetupNatGateway(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.NATGatewayGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: ec2.NewNatGatewayClient}),
		managed.WithCreationGracePeriod(3 * time.Minute),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithInitializers(),
		managed.WithConnectionPublishers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.NATGatewayGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1beta1.NATGateway{}).
		Complete(r)
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) ec2.NatGatewayClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1beta1.NATGateway)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	cfg, err := connectaws.GetConfig(ctx, c.kube, mg, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, err
	}
	return &external{client: c.newClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	kube   client.Client
	client ec2.NatGatewayClient
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1beta1.NATGateway)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	response, err := e.client.DescribeNatGateways(ctx, &awsec2.DescribeNatGatewaysInput{
		NatGatewayIds: []string{meta.GetExternalName(cr)},
	})
	if err != nil {
		return managed.ExternalObservation{}, errorutils.Wrap(resource.Ignore(ec2.IsNatGatewayNotFoundErr, err), errDescribe)
	}

	// in a successful response, there should be one and only one object
	if len(response.NatGateways) != 1 {
		return managed.ExternalObservation{}, errors.New(errNotSingleItem)
	}

	observed := response.NatGateways[0]

	cr.Status.AtProvider = ec2.GenerateNATGatewayObservation(observed)

	switch cr.Status.AtProvider.State {
	case v1beta1.NatGatewayStatusPending:
		cr.SetConditions(xpv1.Unavailable())
	case v1beta1.NatGatewayStatusFailed:
		cr.SetConditions(xpv1.Unavailable().WithMessage(aws.ToString(observed.FailureMessage)))
	case v1beta1.NatGatewayStatusAvailable:
		cr.SetConditions(xpv1.Available())
	case v1beta1.NatGatewayStatusDeleting:
		cr.SetConditions(xpv1.Deleting())
	case v1beta1.NatGatewayStatusDeleted:
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: ec2.CompareTagsV1Beta1(cr.Spec.ForProvider.Tags, observed.Tags),
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1beta1.NATGateway)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	// Create an input without tags.
	input := &awsec2.CreateNatGatewayInput{
		ConnectivityType: awsec2types.ConnectivityType(cr.Spec.ForProvider.ConnectivityType),
		AllocationId:     cr.Spec.ForProvider.AllocationID,
		SubnetId:         cr.Spec.ForProvider.SubnetID,
	}

	// If we specified tags, update the above input.
	if cr.Spec.ForProvider.Tags != nil {
		input.TagSpecifications = []awsec2types.TagSpecification{{
			ResourceType: "natgateway",
			Tags:         ec2.GenerateEC2TagsV1Beta1(cr.Spec.ForProvider.Tags),
		}}
	}

	nat, err := e.client.CreateNatGateway(ctx, input)
	if err != nil {
		return managed.ExternalCreation{}, errorutils.Wrap(err, errCreate)
	}
	meta.SetExternalName(cr, aws.ToString(nat.NatGateway.NatGatewayId))
	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*v1beta1.NATGateway)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	response, err := e.client.DescribeNatGateways(ctx, &awsec2.DescribeNatGatewaysInput{
		NatGatewayIds: []string{meta.GetExternalName(cr)},
	})
	if err != nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(resource.Ignore(ec2.IsNatGatewayNotFoundErr, err), errDescribe)
	}

	// in a successful response, there should be one and only one object
	if len(response.NatGateways) != 1 {
		return managed.ExternalUpdate{}, errors.New(errNotSingleItem)
	}

	observed := response.NatGateways[0]

	addTags, RemoveTags := ec2.DiffEC2Tags(ec2.GenerateEC2TagsV1Beta1(cr.Spec.ForProvider.Tags), observed.Tags)
	if len(RemoveTags) > 0 {
		if _, err := e.client.DeleteTags(ctx, &awsec2.DeleteTagsInput{
			Resources: []string{meta.GetExternalName(cr)},
			Tags:      RemoveTags,
		}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errDeleteTags)
		}
	}
	if len(addTags) > 0 {
		if _, err := e.client.CreateTags(ctx, &awsec2.CreateTagsInput{
			Resources: []string{meta.GetExternalName(cr)},
			Tags:      addTags,
		}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdateTags)
		}
	}
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1beta1.NATGateway)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Deleting())
	if cr.Status.AtProvider.State == v1beta1.NatGatewayStatusDeleted ||
		cr.Status.AtProvider.State == v1beta1.NatGatewayStatusDeleting {
		return nil
	}

	_, err := e.client.DeleteNatGateway(ctx, &awsec2.DeleteNatGatewayInput{
		NatGatewayId: aws.String(meta.GetExternalName(cr)),
	})

	return errorutils.Wrap(resource.Ignore(ec2.IsNatGatewayNotFoundErr, err), errDelete)
}
