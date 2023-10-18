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

	svcsdk "github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/aws/aws-sdk-go/service/servicediscovery/servicediscoveryiface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	cpresource "github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/servicediscovery/v1alpha1"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

const (
	errUnexpectedObject           = "managed resource is not a namespace resource"
	errGetNamespace               = "get-namespace failed"
	errDeleteNamespace            = "delete-namespace failed"
	errOperationResponseMalformed = "get-operation result malformed"

	errListTagsForResource = "cannot list tags"
	errRemoveTags          = "cannot remove tags"
	errCreateTags          = "cannot create tags"
)

type namespace interface {
	cpresource.Managed
	GetOperationID() *string
	SetOperationID(*string)
	GetDescription() *string
	GetTTL() *int64
	GetTags() []*v1alpha1.Tag
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
func (h *Hooks) Observe(ctx context.Context, mg cpresource.Managed) (managed.ExternalObservation, error) { //nolint:gocyclo
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
		if pointer.StringValue(cr.GetOperationID()) == "" {
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
		switch pointer.StringValue(opResp.Operation.Status) {
		case "PENDING", "SUBMITTED":
			return managed.ExternalObservation{}, nil
		case "FAIL":
			cr.SetConditions(xpv1.Unavailable().WithMessage(pointer.StringValue(opResp.Operation.ErrorMessage)))
			return managed.ExternalObservation{}, nil
		}
		namespaceID, ok := opResp.Operation.Targets["NAMESPACE"]
		if !ok {
			return managed.ExternalObservation{}, errors.New(errOperationResponseMalformed)
		}

		if meta.GetExternalName(mg) != pointer.StringValue(namespaceID) {
			// We need to make sure external name makes it to api-server no matter what.
			err := retry.OnError(retry.DefaultRetry, cpresource.IsAPIError, func() error {
				nn := types.NamespacedName{Name: cr.GetName()}
				if err := h.kube.Get(ctx, nn, mg); err != nil {
					return err
				}
				meta.SetExternalName(mg, pointer.StringValue(namespaceID))
				return h.kube.Update(ctx, mg)
			})
			if err != nil {
				return managed.ExternalObservation{}, errors.Wrap(err, "cannot update with external name")
			}
		}
	}

	nsInput := &svcsdk.GetNamespaceInput{
		Id: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
	}
	nsReqResp, err := h.client.GetNamespaceWithContext(ctx, nsInput)
	if err != nil {
		// Deleting is done
		if cr.GetCondition(xpv1.TypeReady).Reason == xpv1.ReasonDeleting {
			return managed.ExternalObservation{}, nil
		}
		cr.SetConditions(xpv1.Unavailable())
		return managed.ExternalObservation{},
			errorutils.Wrap(cpresource.Ignore(ActualIsNotFound, err), errGetNamespace)
	}

	// Deleting is still on-going.
	if cr.GetCondition(xpv1.TypeReady).Reason == xpv1.ReasonDeleting {
		return managed.ExternalObservation{
			ResourceExists: true,
		}, nil
	}

	cr.SetConditions(xpv1.Available())

	upToDate := true
	tagUpToDate, err := AreTagsUpToDate(h.client, cr.GetTags(), nsReqResp.Namespace.Arn)
	if err != nil {
		cr.SetConditions(xpv1.Unavailable())
		return managed.ExternalObservation{
				ResourceExists: true,
			},
			errorutils.Wrap(cpresource.Ignore(ActualIsNotFound, err), errListTagsForResource)
	}
	if !tagUpToDate {
		// Update Tags
		upToDate = false
	}

	if cr.GetDescription() != nil && // Ignore aws value if no description are set
		pointer.StringValue(cr.GetDescription()) != pointer.StringValue(nsReqResp.Namespace.Description) {
		// Update Description
		upToDate = false
	}

	if cr.GetTTL() != nil && // Ignore aws value if no ttl are set
		(nsReqResp.Namespace == nil || nsReqResp.Namespace.Properties == nil || nsReqResp.Namespace.Properties.DnsProperties == nil ||
			pointer.Int64Value(cr.GetTTL()) != pointer.Int64Value(nsReqResp.Namespace.Properties.DnsProperties.SOA.TTL)) {
		// Update TTL
		upToDate = false
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
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
		Id: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
	}
	op, err := h.client.DeleteNamespaceWithContext(ctx, input)
	if cpresource.IgnoreAny(err, ActualIsNotFound, IsDuplicateRequest) != nil {
		return errorutils.Wrap(err, errDeleteNamespace)
	}
	cr.SetOperationID(op.OperationId)
	return nil
}

// ActualIsNotFound reimplements IsNotFound which doesn't do it's job
// IsNotFound test for error code UNKNOWN
func ActualIsNotFound(err error) bool {
	var namespaceNotFound *svcsdk.NamespaceNotFound
	var operationNotFound *svcsdk.OperationNotFound
	return errors.As(err, &namespaceNotFound) || errors.As(err, &operationNotFound)
}

// IsDuplicateRequest checks if an error is DuplicateRequest
func IsDuplicateRequest(err error) bool {
	var duplicateRequest *svcsdk.DuplicateRequest
	return errors.As(err, &duplicateRequest)
}

// AreTagsUpToDate for spec and resourceName
func AreTagsUpToDate(client servicediscoveryiface.ServiceDiscoveryAPI, spec []*v1alpha1.Tag, resourceName *string) (bool, error) {
	current, err := ListTagsForResource(client, resourceName)
	if err != nil {
		return false, err
	}

	add, remove := DiffTags(spec, current)

	return len(add) == 0 && len(remove) == 0, nil
}

// UpdateTagsForResource with resourceName
func UpdateTagsForResource(client servicediscoveryiface.ServiceDiscoveryAPI, spec []*v1alpha1.Tag, cr v1.Object) error {

	nsInput := &svcsdk.GetNamespaceInput{
		Id: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
	}

	nsReqResp, err := client.GetNamespace(nsInput)
	if err != nil {
		return err
	}

	current, err := ListTagsForResource(client, nsReqResp.Namespace.Arn)
	if err != nil {
		return err
	}

	add, remove := DiffTags(spec, current)
	if len(remove) != 0 {
		if _, err := client.UntagResource(&svcsdk.UntagResourceInput{
			ResourceARN: nsReqResp.Namespace.Arn,
			TagKeys:     remove,
		}); err != nil {
			return errors.Wrap(err, errRemoveTags)
		}
	}
	if len(add) != 0 {
		if _, err := client.TagResource(&svcsdk.TagResourceInput{
			ResourceARN: nsReqResp.Namespace.Arn,
			Tags:        add,
		}); err != nil {
			return errors.Wrap(err, errCreateTags)
		}
	}

	return nil
}

// ListTagsForResource for the given resource
func ListTagsForResource(client servicediscoveryiface.ServiceDiscoveryAPI, resourceARN *string) ([]*svcsdk.Tag, error) {
	req := &svcsdk.ListTagsForResourceInput{
		ResourceARN: resourceARN,
	}

	resp, err := client.ListTagsForResource(req)
	if err != nil {
		return nil, errors.Wrap(err, errListTagsForResource)
	}

	return resp.Tags, nil
}

// DiffTags between spec and current
func DiffTags(spec []*v1alpha1.Tag, current []*svcsdk.Tag) (addTags []*svcsdk.Tag, removeTags []*string) {
	currentMap := make(map[string]string, len(current))
	for _, t := range current {
		currentMap[pointer.StringValue(t.Key)] = pointer.StringValue(t.Value)
	}

	specMap := make(map[string]string, len(spec))
	for _, t := range spec {
		key := pointer.StringValue(t.Key)
		val := pointer.StringValue(t.Value)
		specMap[key] = pointer.StringValue(t.Value)

		if currentVal, exists := currentMap[key]; exists {
			if currentVal != val {
				removeTags = append(removeTags, t.Key)
				addTags = append(addTags, &svcsdk.Tag{
					Key:   pointer.ToOrNilIfZeroValue(key),
					Value: pointer.ToOrNilIfZeroValue(val),
				})
			}
		} else {
			addTags = append(addTags, &svcsdk.Tag{
				Key:   pointer.ToOrNilIfZeroValue(key),
				Value: pointer.ToOrNilIfZeroValue(val),
			})
		}
	}

	for _, t := range current {
		key := pointer.StringValue(t.Key)
		if _, exists := specMap[key]; !exists {
			removeTags = append(removeTags, pointer.ToOrNilIfZeroValue(key))
		}
	}

	return addTags, removeTags
}
