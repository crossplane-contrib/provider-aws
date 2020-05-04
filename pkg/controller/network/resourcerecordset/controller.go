package resourcerecordset

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/network/v1alpha3"
	"github.com/crossplane/provider-aws/pkg/clients/resourcerecordset"
	"github.com/crossplane/provider-aws/pkg/controller/utils"
)

const (
	errUnexpectedObject = "The managed resource is not an ResourceRecordSet resource"
	errChange           = "failed to change the ResourceRecordSet resource"
	errList             = "failed to list the ResourceRecordSet resource"
)

// SetupResourceRecordSet adds a controller that reconciles ResourceRecordSets.
func SetupResourceRecordSet(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha3.ResourceRecordSetGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha3.ResourceRecordSet{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha3.ResourceRecordSetGroupVersionKind),
			managed.WithExternalConnecter(&connector{client: mgr.GetClient(), newClientFn: resourcerecordset.NewClient, awsConfigFn: utils.RetrieveAwsConfigFromProvider}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithConnectionPublishers(),
			managed.WithInitializers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	client      client.Client
	newClientFn func(*aws.Config) resourcerecordset.Client
	awsConfigFn func(context.Context, client.Reader, *corev1.ObjectReference) (*aws.Config, error)
}

func (conn *connector) Connect(ctx context.Context, mgd resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mgd.(*v1alpha3.ResourceRecordSet)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}

	awsconfig, err := conn.awsConfigFn(ctx, conn.client, cr.Spec.ProviderReference)
	if err != nil {
		return nil, err
	}

	c := conn.newClientFn(awsconfig)

	return &external{kube: conn.client, client: c}, nil
}

type external struct {
	kube   client.Client
	client resourcerecordset.Client
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1alpha3.ResourceRecordSet)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	rrset, err := resourcerecordset.GetResourceRecordSetOrErr(ctx, e.client, cr.Spec.ForProvider)
	if resourcerecordset.IsRRSetNotFoundErr(err) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	if err != nil {
		return managed.ExternalObservation{
			ResourceExists:    false,
			ConnectionDetails: managed.ConnectionDetails{},
		}, errors.Wrap(err, errList)
	}

	cr.Status.AtProvider = resourcerecordset.GenerateObservation(rrset)

	upToDate, err := resourcerecordset.IsUpToDate(cr.Spec.ForProvider, rrset)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "asdf")
	}

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  upToDate,
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {

	cr, ok := mgd.(*v1alpha3.ResourceRecordSet)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Creating())

	input := resourcerecordset.GenerateChangeResourceRecordSetsInput(&cr.Spec.ForProvider, route53.ChangeActionCreate)
	_, err := e.client.ChangeResourceRecordSetsRequest(input).Send(ctx)

	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errChange)
	}

	return managed.ExternalCreation{}, errors.Wrap(nil, errChange)
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {

	cr, ok := mgd.(*v1alpha3.ResourceRecordSet)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}
	input := resourcerecordset.GenerateChangeResourceRecordSetsInput(&cr.Spec.ForProvider, route53.ChangeActionUpsert)
	_, err := e.client.ChangeResourceRecordSetsRequest(input).Send(ctx)

	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errChange)
	}

	return managed.ExternalUpdate{}, errors.Wrap(nil, errChange)
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1alpha3.ResourceRecordSet)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Deleting())

	input := resourcerecordset.GenerateChangeResourceRecordSetsInput(&cr.Spec.ForProvider, route53.ChangeActionDelete)
	_, err := e.client.ChangeResourceRecordSetsRequest(input).Send(ctx)

	if err != nil {
		return errors.Wrap(nil, errChange)
	}

	return errors.Wrap(nil, errChange)
}
