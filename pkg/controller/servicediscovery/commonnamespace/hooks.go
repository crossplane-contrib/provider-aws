/*
Copyright 2021 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package commonnamespace

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/awserr"
	svcsdk "github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/aws/aws-sdk-go/service/servicediscovery/servicediscoveryiface"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	cpresource "github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/servicediscovery/v1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
)

const (
	errUnexpectedObject           = "managed resource is not a namespace resource"
	errGetNamespace               = "get-namespace failed"
	errDeleteNamespace            = "delete-namespace failed"
	errOperationResponseMalformed = "get-operation result malformed"
)

type namespace interface {
	cpresource.Managed
	GetOperationID() *string
	SetOperationID(*string)
	GetDescription() *string
	SetDescription(*string)
}

// NewHooks returns a new Hooks object.
func NewHooks(kube client.Client, client servicediscoveryiface.ServiceDiscoveryAPI) *Hooks {
	return &Hooks{
		client: client,
		kube:   kube,
	}
}

// Hooks implements common hooks so that all ServiceDiscovery Namespace resources can use.
type Hooks struct {
	client servicediscoveryiface.ServiceDiscoveryAPI
	kube   client.Client
}

// Observe observes any of HTTPNamespace, PrivateDNSNamespace or PublicDNSNamespace types.
func (h *Hooks) Observe(ctx context.Context, mg cpresource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	var cr namespace
	switch i := mg.(type) {
	case *v1alpha1.HTTPNamespace:
		cr = i
	case *v1alpha1.PrivateDNSNamespace:
		cr = i
	case *v1alpha1.PublicDNSNamespace:
		cr = i
	default:
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}
	// Creation is still on-going.
	if meta.GetExternalName(cr) == "" {
		if awsclient.StringValue(cr.GetOperationID()) == "" {
			return managed.ExternalObservation{}, nil
		}
		opInput := &svcsdk.GetOperationInput{
			OperationId: cr.GetOperationID(),
		}
		opResp, err := h.client.GetOperationWithContext(ctx, opInput)
		if err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, "get-operation failed")
		}
		if opResp.Operation == nil || len(opResp.Operation.Targets) == 0 {
			return managed.ExternalObservation{}, errors.New(errOperationResponseMalformed)
		}
		switch awsclient.StringValue(opResp.Operation.Status) {
		case "PENDING", "SUBMITTED":
			return managed.ExternalObservation{ResourceExists: true}, nil
		case "FAIL":
			cr.SetConditions(xpv1.Unavailable().WithMessage(awsclient.StringValue(opResp.Operation.ErrorMessage)))
			return managed.ExternalObservation{}, nil
		}
		namespaceID, ok := opResp.Operation.Targets["NAMESPACE"]
		if !ok {
			return managed.ExternalObservation{}, errors.New(errOperationResponseMalformed)
		}
		if meta.GetExternalName(mg) != awsclient.StringValue(namespaceID) {
			// We need to make sure external name makes it to api-server no matter what.
			err := retry.OnError(retry.DefaultRetry, cpresource.IsAPIError, func() error {
				nn := types.NamespacedName{Name: cr.GetName()}
				if err := h.kube.Get(ctx, nn, mg); err != nil {
					return err
				}
				meta.SetExternalName(mg, awsclient.StringValue(namespaceID))
				return h.kube.Update(ctx, mg)
			})
			if err != nil {
				return managed.ExternalObservation{}, errors.Wrap(err, "cannot update with external name")
			}
		}
	}

	nsInput := &svcsdk.GetNamespaceInput{
		Id: awsclient.String(meta.GetExternalName(cr)),
	}
	nsReqResp, err := h.client.GetNamespaceWithContext(ctx, nsInput)
	if err != nil {
		cr.SetConditions(xpv1.Unavailable())
		return managed.ExternalObservation{},
			awsclient.Wrap(cpresource.Ignore(ActualIsNotFound, err), errGetNamespace)
	}
	cr.SetConditions(xpv1.Available())
	lateInited := false
	if awsclient.StringValue(cr.GetDescription()) == "" {
		cr.SetDescription(nsReqResp.Namespace.Description)
		lateInited = true
	}
	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceLateInitialized: lateInited,
		ResourceUpToDate:        true, // Namespaces cannot be updated.
	}, nil
}

// Delete deletes any of HTTPNamespace, PrivateDNSNamespace or PublicDNSNamespace types.
func (h *Hooks) Delete(ctx context.Context, mg cpresource.Managed) error {
	var cr namespace
	switch i := mg.(type) {
	case *v1alpha1.HTTPNamespace:
		cr = i
	case *v1alpha1.PrivateDNSNamespace:
		cr = i
	case *v1alpha1.PublicDNSNamespace:
		cr = i
	default:
		return errors.New(errUnexpectedObject)
	}
	input := &svcsdk.DeleteNamespaceInput{
		Id: awsclient.String(meta.GetExternalName(cr)),
	}
	op, err := h.client.DeleteNamespaceWithContext(ctx, input)
	if cpresource.IgnoreAny(err, ActualIsNotFound, IsDuplicateRequest) != nil {
		return awsclient.Wrap(err, errDeleteNamespace)
	}
	cr.SetOperationID(op.OperationId)
	return nil
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
