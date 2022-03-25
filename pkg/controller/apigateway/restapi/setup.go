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

package restapi

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	svcsdk "github.com/aws/aws-sdk-go/service/apigateway"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane/provider-aws/apis/apigateway/v1alpha1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
	apigwclient "github.com/crossplane/provider-aws/pkg/clients/apigateway"
)

// SetupRestAPI adds a controller that reconciles RestAPI.
func SetupRestAPI(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.RestAPIGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preDelete = preDelete
			e.lateInitialize = lateInitialize
			e.isUpToDate = isUpToDate
			e.postCreate = postCreate
			c := &custom{
				Client: &apigwclient.GatewayClient{Client: e.client},
			}
			e.preUpdate = c.preUpdate
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&svcapitypes.RestAPI{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.RestAPIGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithInitializers(),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type custom struct {
	Client apigwclient.Client
}

func preObserve(_ context.Context, cr *svcapitypes.RestAPI, obj *svcsdk.GetRestApiInput) error {
	obj.RestApiId = aws.String(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.RestAPI, obj *svcsdk.RestApi, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	cr.SetConditions(xpv1.Available())
	return obs, nil
}

func postCreate(_ context.Context, cr *svcapitypes.RestAPI, resp *svcsdk.RestApi, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	meta.SetExternalName(cr, aws.StringValue(resp.Id))
	return cre, nil
}

func (c *custom) preUpdate(ctx context.Context, cr *svcapitypes.RestAPI, obj *svcsdk.UpdateRestApiInput) error {
	cur := &svcapitypes.RestAPIParameters{
		Region: cr.Spec.ForProvider.Region,
	}

	rapi, err := c.Client.GetRestAPIByID(ctx, aws.String(meta.GetExternalName(cr)))
	if err != nil {
		return errors.Wrap(err, "cant get rest api")
	}

	cur.Name = rapi.Name
	rapi.Policy = unescapePolicy(rapi.Policy)

	if err := lateInitialize(cur, rapi); err != nil {
		return errors.Wrap(err, "cant late init current restApi")
	}

	if cr.Spec.ForProvider.Policy != nil {
		pol, err := normalizePolicy(cr.Spec.ForProvider.Policy)
		if err != nil {
			return errors.Wrap(err, "cant normalize managed resource policy")
		}

		cr.Spec.ForProvider.Policy = pol
	}

	pOps, err := apigwclient.GetPatchOperations(&cur, &cr.Spec.ForProvider)
	if err != nil {
		return errors.Wrap(err, "cant compute patch preUpdate")
	}

	obj.PatchOperations = pOps
	obj.RestApiId = aws.String(meta.GetExternalName(cr))

	return nil
}

func policiesAreKindOfTheSame(a *string, b *string) (bool, error) {
	if a != nil && b != nil {
		aPol, bPol, err := parsePolicies(a, b)
		if err != nil {
			return false, errors.Wrap(err, "cant parse policies")
		}
		polPatch, err := apigwclient.GetJSONPatch(aPol, bPol)
		if err != nil {
			return false, errors.Wrap(err, "cant compute jsonpatch")
		}

		for _, p := range polPatch {
			re := regexp.MustCompile(`/Statement/(\d+)/Resource`)
			if p.Operation != "replace" || !re.MatchString(p.Path) {
				return false, nil
			}

			index, _ := strconv.Atoi(re.FindString(p.Path))
			fmt.Println(aPol["Statement"].([]interface{})[index].(map[string]interface{})["Resource"])
			if aPol["Statement"].([]interface{})[index].(map[string]interface{})["Resource"] != "execute-api:/*/*/*" {
				return false, nil
			}
		}

		return true, nil
	}

	return cmp.Equal(a, b), nil
}

func preDelete(_ context.Context, cr *svcapitypes.RestAPI, obj *svcsdk.DeleteRestApiInput) (bool, error) {
	obj.RestApiId = aws.String(meta.GetExternalName(cr))
	return false, nil
}

func isUpToDate(cr *svcapitypes.RestAPI, cur *svcsdk.RestApi) (bool, error) {
	s := &svcapitypes.RestAPIParameters{
		Name:   cur.Name,
		Region: cr.Spec.ForProvider.Region,
	}

	if cur.Policy != nil {
		cur.Policy = unescapePolicy(cur.Policy)
	}

	if err := lateInitialize(s, cur); err != nil {
		return false, errors.Wrap(err, "cant lateinit")
	}

	patchJSON, err := aws.CreateJSONPatch(cr.Spec.ForProvider, &s)
	if err != nil {
		return false, errors.Wrap(err, "error checking up to date")
	}

	patch := &svcapitypes.RestAPIParameters{}
	if err := json.Unmarshal(patchJSON, patch); err != nil {
		return false, errors.Wrap(err, "error checking up to date")
	}

	return cmp.Equal(&svcapitypes.RestAPIParameters{}, patch,
		cmpopts.IgnoreTypes([]xpv1.Reference{}, []xpv1.Selector{}),
		cmpopts.IgnoreFields(svcapitypes.RestAPIParameters{}, "Region"),
	), nil
}

func normalizePolicy(p *string) (*string, error) {
	if p == nil {
		return p, nil
	}
	var mappedPol map[string]interface{}
	if err := json.Unmarshal([]byte(*p), &mappedPol); err != nil {
		return nil, errors.Wrap(err, "cant unmarshal policy")
	}

	parsed, err := json.Marshal(mappedPol)
	if err != nil {
		return nil, err
	}

	return aws.String(string(parsed)), nil
}

func unescapePolicy(p *string) *string {
	if p == nil {
		return p
	}

	s := strings.ReplaceAll(*p, "\\\"", "\"")
	return &s
}

func parsePolicies(a *string, b *string) (map[string]interface{}, map[string]interface{}, error) {
	var aPol, bPol map[string]interface{}
	if err := json.Unmarshal([]byte(*a), &aPol); err != nil {
		return nil, nil, errors.Wrap(err, "cant unmarshal policy")
	}
	if err := json.Unmarshal([]byte(*b), &bPol); err != nil {
		return nil, nil, errors.Wrap(err, "cant unmarshal policy")
	}

	return aPol, bPol, nil
}

func lateInitialize(in *svcapitypes.RestAPIParameters, cur *svcsdk.RestApi) error {
	in.APIKeySource = aws.LateInitializeStringPtr(in.APIKeySource, cur.ApiKeySource)
	in.BinaryMediaTypes = aws.LateInitializeStringPtrSlice(in.BinaryMediaTypes, cur.BinaryMediaTypes)
	in.Description = aws.LateInitializeStringPtr(in.Description, cur.Description)
	in.DisableExecuteAPIEndpoint = aws.LateInitializeBoolPtr(in.DisableExecuteAPIEndpoint, cur.DisableExecuteApiEndpoint)
	in.MinimumCompressionSize = aws.LateInitializeInt64Ptr(in.MinimumCompressionSize, cur.MinimumCompressionSize)

	// this is a hack since AWS does minor adaptions to the policy after creation. We want to treat just that case
	// and avoid copying from the AWS status of a resources as the source of truth
	res, err := policiesAreKindOfTheSame(in.Policy, cur.Policy)
	if err != nil {
		return errors.Wrap(err, "policies couldnt be compared")
	} else if res {
		in.Policy = cur.Policy
	}

	pol, err := normalizePolicy(aws.LateInitializeStringPtr(in.Policy, cur.Policy))
	if err != nil {
		return errors.Wrap(err, "cant normalize policy")
	}

	in.Policy = pol

	if cur.EndpointConfiguration != nil {
		if in.EndpointConfiguration == nil {
			in.EndpointConfiguration = &svcapitypes.EndpointConfiguration{}
		}
		in.EndpointConfiguration.Types = aws.LateInitializeStringPtrSlice(in.EndpointConfiguration.Types, cur.EndpointConfiguration.Types)
		in.EndpointConfiguration.VPCEndpointIDs = aws.LateInitializeStringPtrSlice(in.EndpointConfiguration.VPCEndpointIDs, cur.EndpointConfiguration.VpcEndpointIds)
	}

	return nil
}
