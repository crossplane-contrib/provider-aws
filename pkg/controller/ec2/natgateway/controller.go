package natgateway

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/ec2/v1alpha1"
	"github.com/crossplane/provider-aws/apis/ec2/v1beta1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/ec2"
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
func SetupNatGateway(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.NATGatewayGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.NATGateway{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.NATGatewayGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: ec2.NewNatGatewayClient}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithConnectionPublishers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) ec2.NatGatewayClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.NATGateway)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	cfg, err := awsclient.GetConfig(ctx, c.kube, mg, cr.Spec.ForProvider.Region)
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
	cr, ok := mgd.(*v1alpha1.NATGateway)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	response, err := e.client.DescribeNatGatewaysRequest(&awsec2.DescribeNatGatewaysInput{
		NatGatewayIds: []string{meta.GetExternalName(cr)},
	}).Send(ctx)
	if err != nil {
		return managed.ExternalObservation{}, awsclient.Wrap(resource.Ignore(ec2.IsNatGatewayNotFoundErr, err), errDescribe)
	}

	// in a successful response, there should be one and only one object
	if len(response.NatGateways) != 1 {
		return managed.ExternalObservation{}, errors.New(errNotSingleItem)
	}

	observed := response.NatGateways[0]

	cr.Status.AtProvider = ec2.GenerateNATGatewayObservation(observed)

	switch cr.Status.AtProvider.State {
	case v1alpha1.NatGatewayStatusPending:
		cr.SetConditions(xpv1.Unavailable())
	case v1alpha1.NatGatewayStatusFailed:
		cr.SetConditions(xpv1.Unavailable().WithMessage(aws.StringValue(observed.FailureMessage)))
	case v1alpha1.NatGatewayStatusAvailable:
		cr.SetConditions(xpv1.Available())
	case v1alpha1.NatGatewayStatusDeleting:
		cr.SetConditions(xpv1.Deleting())
	case v1alpha1.NatGatewayStatusDeleted:
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: v1beta1.CompareTags(cr.Spec.ForProvider.Tags, observed.Tags),
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1alpha1.NATGateway)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	nat, err := e.client.CreateNatGatewayRequest(&awsec2.CreateNatGatewayInput{
		AllocationId: cr.Spec.ForProvider.AllocationID,
		SubnetId:     cr.Spec.ForProvider.SubnetID,
		TagSpecifications: []awsec2.TagSpecification{
			{
				ResourceType: "natgateway",
				Tags:         v1beta1.GenerateEC2Tags(cr.Spec.ForProvider.Tags),
			},
		},
	}).Send(ctx)
	if err != nil {
		return managed.ExternalCreation{}, awsclient.Wrap(err, errCreate)
	}
	meta.SetExternalName(cr, aws.StringValue(nat.NatGateway.NatGatewayId))
	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*v1alpha1.NATGateway)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	response, err := e.client.DescribeNatGatewaysRequest(&awsec2.DescribeNatGatewaysInput{
		NatGatewayIds: []string{meta.GetExternalName(cr)},
	}).Send(ctx)
	if err != nil {
		return managed.ExternalUpdate{}, awsclient.Wrap(resource.Ignore(ec2.IsNatGatewayNotFoundErr, err), errDescribe)
	}

	// in a successful response, there should be one and only one object
	if len(response.NatGateways) != 1 {
		return managed.ExternalUpdate{}, errors.New(errNotSingleItem)
	}

	observed := response.NatGateways[0]

	addTags, RemoveTags := awsclient.DiffEC2Tags(v1beta1.GenerateEC2Tags(cr.Spec.ForProvider.Tags), observed.Tags)
	if len(RemoveTags) > 0 {
		if _, err := e.client.DeleteTagsRequest(&awsec2.DeleteTagsInput{
			Resources: []string{meta.GetExternalName(cr)},
			Tags:      RemoveTags,
		}).Send(ctx); err != nil {
			return managed.ExternalUpdate{}, awsclient.Wrap(err, errDeleteTags)
		}
	}
	if len(addTags) > 0 {
		if _, err := e.client.CreateTagsRequest(&awsec2.CreateTagsInput{
			Resources: []string{meta.GetExternalName(cr)},
			Tags:      addTags,
		}).Send(ctx); err != nil {
			return managed.ExternalUpdate{}, awsclient.Wrap(err, errUpdateTags)
		}
	}
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1alpha1.NATGateway)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Deleting())
	if cr.Status.AtProvider.State == v1alpha1.NatGatewayStatusDeleted ||
		cr.Status.AtProvider.State == v1alpha1.NatGatewayStatusDeleting {
		return nil
	}

	_, err := e.client.DeleteNatGatewayRequest(&awsec2.DeleteNatGatewayInput{
		NatGatewayId: aws.String(meta.GetExternalName(cr)),
	}).Send(ctx)

	return awsclient.Wrap(resource.Ignore(ec2.IsNatGatewayNotFoundErr, err), errDelete)
}
