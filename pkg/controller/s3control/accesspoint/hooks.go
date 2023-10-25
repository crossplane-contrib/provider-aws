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

package accesspoint

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/s3control"
	svcsdkapi "github.com/aws/aws-sdk-go/service/s3control/s3controliface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/s3control/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

type hooks struct {
	policyClient *policyClient
}

func newHooks(api svcsdkapi.S3ControlAPI) *hooks {
	return &hooks{policyClient: &policyClient{client: api}}
}

func (h *hooks) isUpToDate(_ context.Context, point *svcapitypes.AccessPoint, _ *svcsdk.GetAccessPointOutput) (bool, string, error) {
	obs, err := h.policyClient.observe(point)
	if err != nil {
		return false, "", err
	}
	return obs == updated, "", nil
}

func (h *hooks) update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*svcapitypes.AccessPoint)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}
	obs, err := h.policyClient.observe(cr)
	if err != nil {
		cr.Status.SetConditions(xpv1.ReconcileError(err))
		return managed.ExternalUpdate{}, err
	}
	switch obs {
	case needsDeletion:
		err = h.policyClient.delete(ctx, cr)
		if err != nil {
			return managed.ExternalUpdate{}, err
		}
	case needsUpdate:
		if err := h.policyClient.createOrUpdate(ctx, cr); err != nil {
			return managed.ExternalUpdate{}, err
		}
	case updated:
		return managed.ExternalUpdate{}, nil
	}
	return managed.ExternalUpdate{}, nil
}

func preDelete(_ context.Context, point *svcapitypes.AccessPoint, input *svcsdk.DeleteAccessPointInput) (bool, error) {
	input.Name = pointer.ToOrNilIfZeroValue(meta.GetExternalName(point))
	return false, nil
}

func preCreate(_ context.Context, point *svcapitypes.AccessPoint, input *svcsdk.CreateAccessPointInput) error {
	input.Name = pointer.ToOrNilIfZeroValue(meta.GetExternalName(point))
	input.Bucket = point.Spec.ForProvider.BucketName
	return nil
}

func preObserve(_ context.Context, point *svcapitypes.AccessPoint, input *svcsdk.GetAccessPointInput) error {
	input.Name = pointer.ToOrNilIfZeroValue(meta.GetExternalName(point))
	return nil
}

func postObserve(_ context.Context, point *svcapitypes.AccessPoint, _ *svcsdk.GetAccessPointOutput, observation managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	point.SetConditions(xpv1.Available())
	return observation, nil
}
