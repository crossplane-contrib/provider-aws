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

package httpnamespace

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
	e.delete = h.delete
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
	cr, ok := mg.(*svcapitypes.HTTPNamespace)
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

	if meta.GetExternalName(cr) != awsclient.StringValue(namespaceID) {
		// We need to make sure external name makes it to api-server no matter what.
		err := retry.OnError(retry.DefaultRetry, cpresource.IsAPIError, func() error {
			nn := types.NamespacedName{Name: cr.GetName()}
			if err := h.kube.Get(ctx, nn, cr); err != nil {
				return err
			}
			meta.SetExternalName(cr, awsclient.StringValue(namespaceID))
			return h.kube.Update(ctx, cr)
		})
		if err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, "cannot update with external name")
		}
	}
	cr.Status.SetConditions(xpv1.Available())
	lateInited := false
	if cr.Spec.ForProvider.Description == nil {
		cr.Spec.ForProvider.Description = nsReqResp.Namespace.Description
		lateInited = true
	}
	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceLateInitialized: lateInited,
		ResourceUpToDate:        true, // Namespaces cannot be updated.
	}, nil
}

func preCreate(_ context.Context, cr *svcapitypes.HTTPNamespace, obj *svcsdk.CreateHttpNamespaceInput) error {
	obj.CreatorRequestId = awsclient.String(string(cr.UID))
	return nil
}

func (h *hooks) delete(ctx context.Context, mg cpresource.Managed) error {
	cr, ok := mg.(*svcapitypes.HTTPNamespace)
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	input := &svcsdk.DeleteNamespaceInput{
		Id: awsclient.String(meta.GetExternalName(cr)),
	}
	_, err := h.client.DeleteNamespaceWithContext(ctx, input)
	return awsclient.Wrap(cpresource.IgnoreAny(err, ActualIsNotFound, IsDuplicateRequest), errDelete)
}
