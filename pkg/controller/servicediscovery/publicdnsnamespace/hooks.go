package publicdnsnamespace

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/awserr"
	svcsdk "github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/aws/aws-sdk-go/service/servicediscovery/servicediscoveryiface"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	cpresource "github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/servicediscovery/v1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
)

const (
	errGetNamespace = "get-namespace failed"
)

func useHooks(e *external) {
	h := &hooks{client: e.client, kube: e.kube}
	e.observe = h.observe
	e.preCreate = preCreate
	e.postCreate = nopPostCreate
	e.delete = h.delete
	e.update = nopUpdate
}

type hooks struct {
	client servicediscoveryiface.ServiceDiscoveryAPI
	kube   client.Client
}

// ActualIsNotFound reimplements IsNotFound which doesn't do it's job
// IsNotFound test for error code UNKNOWN
func ActualIsNotFound(err error) bool {
	awsErr, ok := err.(awserr.Error)
	return ok && (awsErr.Code() == svcsdk.ErrCodeNamespaceNotFound ||
		awsErr.Code() == svcsdk.ErrCodeOperationNotFound)
}

// IsDuplicateRequest checks if an error is DuplicateRequest
func IsDuplicateRequest(err error) bool {
	awsErr, ok := err.(awserr.Error)
	return ok && (awsErr.Code() == svcsdk.ErrCodeDuplicateRequest)
}

func (h *hooks) observe(ctx context.Context, mg cpresource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	cr, ok := mg.(*svcapitypes.PublicDNSNamespace)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if awsclient.StringValue(cr.Status.AtProvider.OperationID) == "" {
		return managed.ExternalObservation{}, nil
	}
	opInput := &svcsdk.GetOperationInput{
		OperationId: cr.Status.AtProvider.OperationID,
	}

	opResp, err := h.client.GetOperationWithContext(ctx, opInput)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "get-operation failed")
	}

	opStatus := awsclient.StringValue(opResp.Operation.Status)
	if opStatus == "PENDING" || opStatus == "SUBMITTED" {
		return managed.ExternalObservation{
			ResourceExists: true,
		}, nil
	}

	if opStatus == "FAIL" {
		errMsg := ""
		if opResp.Operation.ErrorMessage != nil {
			errMsg = *opResp.Operation.ErrorMessage
		}
		isDeleting := cr.GetCondition(xpv1.TypeReady).Reason == xpv1.ReasonDeleting
		cr.Status.SetConditions(xpv1.Unavailable().WithMessage(errMsg))
		if !isDeleting {
			return managed.ExternalObservation{
				ResourceExists: true,
			}, nil
		}
		return managed.ExternalObservation{}, nil
	}

	namespaceID, ok := opResp.Operation.Targets["NAMESPACE"]
	if !ok {
		return managed.ExternalObservation{
			ResourceExists: true,
		}, errors.New(errDescribe)
	}

	nsInput := &svcsdk.GetNamespaceInput{
		Id: namespaceID,
	}
	nsReqResp, err := h.client.GetNamespaceWithContext(ctx, nsInput)
	if err != nil {
		if ActualIsNotFound(err) || IsDuplicateRequest(err) {
			cr.Status.SetConditions(xpv1.Unavailable())
		}
		return managed.ExternalObservation{},
			awsclient.Wrap(cpresource.IgnoreAny(err, ActualIsNotFound, IsDuplicateRequest), errGetNamespace)
	}

	resourceLateInitialized := false
	cr.Status.SetConditions(xpv1.Available())
	namespaceIDStr := awsclient.StringValue(namespaceID)
	externalName := meta.GetExternalName(cr)
	if externalName != namespaceIDStr {
		meta.SetExternalName(cr, namespaceIDStr)
		resourceLateInitialized = true
	}

	currentSpec := cr.Spec.ForProvider.DeepCopy()
	lateInitialize(&cr.Spec.ForProvider, nsReqResp)
	resourceLateInitialized = resourceLateInitialized || !cmp.Equal(&cr.Spec.ForProvider, currentSpec)

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceLateInitialized: resourceLateInitialized,
		ResourceUpToDate:        isUpToDate(cr, nsReqResp),
	}, nil
}

func isUpToDate(cr *svcapitypes.PublicDNSNamespace, resp *svcsdk.GetNamespaceOutput) bool {
	if meta.GetExternalName(cr) != awsclient.StringValue(resp.Namespace.Id) {
		return false
	}

	if cr.Spec.ForProvider.CreatorRequestID != resp.Namespace.CreatorRequestId {
		return false
	}
	if cr.Spec.ForProvider.Description != resp.Namespace.Description {
		return false
	}

	// if cr.Spec.ForProvider.VPC
	// VPC information is not available through servicediscovery API. Instead
	// one could use the HostedZone information to verify the VPC configuration
	// through Route53 HostedZone records

	// tags := map[string]string
	// if cmp.Equal(cr.Spec.ForProvider.Tags, tags) {
	// Where does the servicediscovery API provide Tag information?

	// if cr.Spec.ForProvider.Region
	// Region information is not available through servicediscovery API
	return true
}

func preCreate(ctx context.Context, cr *svcapitypes.PublicDNSNamespace, input *svcsdk.CreatePublicDnsNamespaceInput) error {
	input.Name = awsclient.String(meta.GetExternalName(cr))
	return nil
}

func (h *hooks) delete(ctx context.Context, mg cpresource.Managed) error {
	cr, ok := mg.(*svcapitypes.PublicDNSNamespace)
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	input := &svcsdk.DeleteNamespaceInput{
		Id: awsclient.String(meta.GetExternalName(cr)),
	}
	_, err := h.client.DeleteNamespaceWithContext(ctx, input)
	return awsclient.Wrap(cpresource.IgnoreAny(err, ActualIsNotFound, IsDuplicateRequest), errDelete)
}

func lateInitialize(forProvider *svcapitypes.PublicDNSNamespaceParameters, nsReqResp *svcsdk.GetNamespaceOutput) {
	if nsReqResp == nil {
		return
	}

	if forProvider.Description == nil {
		forProvider.Description = nsReqResp.Namespace.Description
	}

	if forProvider.CreatorRequestID == nil {
		forProvider.CreatorRequestID = nsReqResp.Namespace.CreatorRequestId
	}
}
