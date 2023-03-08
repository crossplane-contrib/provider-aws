/*
Copyright 2023 The Crossplane Authors.

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

package bucket

import (
	"context"
	"encoding/json"

	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"k8s.io/utils/strings/slices"

	"github.com/crossplane/crossplane-runtime/pkg/meta"

	"github.com/crossplane-contrib/provider-aws/apis/s3/v1beta1"
	awsclient "github.com/crossplane-contrib/provider-aws/pkg/clients"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/s3"
)

const (
	policyGetFailed    = "cannot get bucket policy"
	policyFormatFailed = "cannot format bucket policy"
	policyPutFailed    = "cannot put bucket policy"
	policyDeleteFailed = "cannot delete bucket policy"
)

// PolicyClient is the client for API methods and reconciling the PublicAccessBlock
type PolicyClient struct {
	client s3.BucketPolicyClient
}

// NewPolicyClient creates the client for Accelerate Configuration
func NewPolicyClient(client s3.BucketPolicyClient) *PolicyClient {
	return &PolicyClient{client: client}
}

// Observe checks if the resource exists and if it matches the local configuration
func (e *PolicyClient) Observe(ctx context.Context, cr *v1beta1.Bucket) (ResourceStatus, error) {
	resp, err := e.client.GetBucketPolicy(ctx, &awss3.GetBucketPolicyInput{
		Bucket: awsclient.String(meta.GetExternalName(cr)),
	})
	if err != nil {
		if s3.IsErrorPolicyNotFound(err) {
			if cr.Spec.ForProvider.Policy == nil {
				return Updated, nil
			}
			return NeedsUpdate, nil
		}
		return NeedsUpdate, errors.Wrap(err, policyGetFailed)
	}
	policy, err := e.formatBucketPolicy(cr)
	if err != nil {
		return NeedsUpdate, errors.Wrap(err, policyFormatFailed)
	}

	// To ensure backwards compatbility with the previous behaviour
	// (Bucket + BucketPolicy).
	// Only delete the policy on AWS if the user has specified to do so.
	if policy == nil && resp.Policy != nil && getBucketPolicyDeletionPolicy(cr) == v1beta1.BucketPolicyDeletionPolicyIfNull {
		return NeedsDeletion, nil
	}
	if cmp.Equal(policy, resp.Policy) {
		return Updated, nil
	}
	return NeedsUpdate, nil
}

// formatBucketPolicy parses and formats the bucket.Spec.BucketPolicy struct
func (e *PolicyClient) formatBucketPolicy(cr *v1beta1.Bucket) (*string, error) {
	if cr.Spec.ForProvider.Policy == nil {
		return nil, nil
	}
	c := cr.DeepCopy()
	body, err := s3.Serialize(c.Spec.ForProvider.Policy)
	if err != nil {
		return nil, err
	}
	byteData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	str := string(byteData)
	return &str, nil
}

// CreateOrUpdate sends a request to have resource created on AWS
func (e *PolicyClient) CreateOrUpdate(ctx context.Context, cr *v1beta1.Bucket) error {
	if cr.Spec.ForProvider.Policy == nil {
		return nil
	}
	policy, err := e.formatBucketPolicy(cr)
	if err != nil {
		return errors.Wrap(err, policyFormatFailed)
	}
	_, err = e.client.PutBucketPolicy(ctx, &awss3.PutBucketPolicyInput{
		Bucket: awsclient.String(meta.GetExternalName(cr)),
		Policy: policy,
	})
	return errors.Wrap(err, policyPutFailed)
}

// Delete removes the public access block configuration.
func (e *PolicyClient) Delete(ctx context.Context, cr *v1beta1.Bucket) error {
	_, err := e.client.DeleteBucketPolicy(ctx, &awss3.DeleteBucketPolicyInput{
		Bucket: awsclient.String(meta.GetExternalName(cr)),
	})
	return errors.Wrap(err, policyDeleteFailed)
}

// LateInitialize is responsible for initializing the resource based on the external value
func (e *PolicyClient) LateInitialize(ctx context.Context, cr *v1beta1.Bucket) error {
	// TODO: For now LateInitialize is not implemented because of the
	//       inconsistencies between remote and local structures.
	//       A manual converter needs to be written, pretty much the inverse of
	//       s3.SerializeBucketPolicyStatement.

	// The only thing that can be done easily is setting the resource field to
	// the bucket ARN automaitcally since resource policies only apply to the
	// object they are attached to.
	bucketARN := cr.Status.AtProvider.ARN
	if cr.Spec.ForProvider.Policy != nil && bucketARN != "" {
		for i, statement := range cr.Spec.ForProvider.Policy.Statements {
			if !slices.Contains(statement.Resource, bucketARN) {
				statement.Resource = []string{bucketARN}
				cr.Spec.ForProvider.Policy.Statements[i] = statement
			}
		}
	}
	return nil
}

// SubresourceExists checks if the subresource this controller manages currently exists
func (e *PolicyClient) SubresourceExists(cr *v1beta1.Bucket) bool {
	return cr.Spec.ForProvider.Policy != nil
}

// getBucketPolicyDeletionPolicy returns the bucket policy deletion policy if
// set. Otherwise the default (Never).
func getBucketPolicyDeletionPolicy(cr *v1beta1.Bucket) v1beta1.BucketPolicyDeletionPolicy {
	if cr.Spec.ForProvider.PolicyUpdatePolicy != nil {
		return cr.Spec.ForProvider.PolicyUpdatePolicy.DeletionPolicy
	}
	return v1beta1.BucketPolicyDeletionPolicyNever
}
