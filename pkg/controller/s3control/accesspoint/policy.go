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

	"github.com/aws/aws-sdk-go/aws/awserr"
	svcsdk "github.com/aws/aws-sdk-go/service/s3control"
	svcsdkapi "github.com/aws/aws-sdk-go/service/s3control/s3controliface"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/pkg/errors"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/s3control/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/s3"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

// policyStatus represents the current status if the policy is updated.
type policyStatus int

const (
	// updated is returned if the policy is updated.
	updated policyStatus = iota
	// needsUpdate is returned if the policy required updating.
	needsUpdate
	// needsDeletion is returned if the policy needs to be deleted.
	needsDeletion
)

const (
	errPutPolicy      = "failed to put access point policy"
	errDescribePolicy = "failed to describe access point policy"
	errDeletePolicy   = "failed to delete access point policy"
	errFormatPolicy   = "failed to format access point policy"
)

// policyClient is the client for API methods and reconciling the Policy.
type policyClient struct {
	client svcsdkapi.S3ControlAPI
}

// observe observes the current status of the Access Point policy.
func (p *policyClient) observe(cr *svcapitypes.AccessPoint) (policyStatus, error) {
	output, err := p.client.GetAccessPointPolicy(generateGetAccessPointPolicyInput(cr))
	if err != nil {
		if isErrorPolicyNotFound(err) {
			if cr.Spec.ForProvider.Policy == nil {
				return updated, nil
			}
			return needsUpdate, nil
		}
		return needsUpdate, errorutils.Wrap(err, errDescribePolicy)
	}

	if cr.Spec.ForProvider.Policy == nil {
		if output.Policy != nil {
			return needsDeletion, nil
		}
		return updated, nil
	}

	diff, err := s3.DiffParsedPolicies(cr.Spec.ForProvider.Policy, output.Policy)
	if diff != "" || err != nil {
		return needsUpdate, err
	}
	return updated, nil
}

// createOrUpdate creates or updates the Access Point policy.
func (p *policyClient) createOrUpdate(ctx context.Context, cr *svcapitypes.AccessPoint) error {
	if cr.Spec.ForProvider.Policy == nil {
		return nil
	}
	input := &svcsdk.PutAccessPointPolicyInput{}
	input.Name = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))

	if cr.Spec.ForProvider.AccountID != nil {
		input.SetAccountId(*cr.Spec.ForProvider.AccountID)
	}

	policy, err := s3.FormatPolicy(cr.Spec.ForProvider.Policy)
	if err != nil {
		return errors.Wrap(err, errFormatPolicy)
	}
	input.SetPolicy(*policy)
	_, err = p.client.PutAccessPointPolicyWithContext(ctx, input)
	return errorutils.Wrap(err, errPutPolicy)
}

// delete deletes the Access Point policy.
func (p *policyClient) delete(ctx context.Context, cr *svcapitypes.AccessPoint) error {
	_, err := p.client.DeleteAccessPointPolicyWithContext(ctx, generateDeleteAccessPointPolicyInput(cr))
	return errorutils.Wrap(err, errDeletePolicy)
}

// isErrorPolicyNotFound returns whether the given error is of type NotFound or not.
func isErrorPolicyNotFound(err error) bool {
	awsErr, ok := err.(awserr.Error) //nolint:errorlint
	return ok && awsErr.Code() == "NoSuchAccessPointPolicy"
}

// generateGetAccessPointPolicyInput returns a get input.
func generateGetAccessPointPolicyInput(cr *svcapitypes.AccessPoint) *svcsdk.GetAccessPointPolicyInput {
	input := &svcsdk.GetAccessPointPolicyInput{}
	input.Name = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))

	if cr.Spec.ForProvider.AccountID != nil {
		input.SetAccountId(*cr.Spec.ForProvider.AccountID)
	}

	return input
}

// generateDeleteAccessPointPolicyInput returns a deletion input.
func generateDeleteAccessPointPolicyInput(cr *svcapitypes.AccessPoint) *svcsdk.DeleteAccessPointPolicyInput {
	input := &svcsdk.DeleteAccessPointPolicyInput{}
	input.Name = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	if cr.Spec.ForProvider.AccountID != nil {
		input.SetAccountId(*cr.Spec.ForProvider.AccountID)
	}

	return input
}
