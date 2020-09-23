/*
Copyright 2019 The Crossplane Authors.

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

package s3bucketpolicy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/storage/v1alpha1"
	awscommon "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/iam"
	"github.com/crossplane/provider-aws/pkg/clients/s3"
)

const (
	errUnexpectedObject = "The managed resource is not an IAMRolePolicyAttachment resource"
	errAttach           = "failed to attach the policy to role"
	errDelete           = "failed to delete the policy for bucket"
	errGet              = "failed to get S3BucketPolicy for bucket with name"
	errUpdate           = "failed to update the policy for bucket"
)

// SetupS3BucketPolicy adds a controller that reconciles
// S3BucketPolicies.
func SetupS3BucketPolicy(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.S3BucketPolicyGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.S3BucketPolicy{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.S3BucketPolicyGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(),
				newClientFn:    s3.NewBucketPolicyClient,
				newIAMClientFn: iam.NewClient}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube           client.Client
	newClientFn    func(config aws.Config) s3.BucketPolicyClient
	newIAMClientFn func(config aws.Config) iam.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cfg, err := awscommon.GetConfig(ctx, c.kube, mg, "")
	if err != nil {
		return nil, err
	}
	return &external{client: c.newClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	client    s3.BucketPolicyClient
	iamclient iam.Client
	kube      client.Client
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1alpha1.S3BucketPolicy)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	resp, err := e.client.GetBucketPolicyRequest(&awss3.GetBucketPolicyInput{
		Bucket: cr.Spec.PolicyBody.BucketName,
	}).Send(ctx)
	if err != nil {
		if s3.IsErrorBucketNotFound(err) {
			return managed.ExternalObservation{}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(s3.IsErrorPolicyNotFound, err), errGet)
	}

	policyData, err := e.formatBucketPolicy(cr)

	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(s3.IsErrorPolicyNotFound, err), errGet)
	}

	cr.SetConditions(runtimev1alpha1.Available())

	// If our version and the external version are the same, we return ResourceUpToDate: true
	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: cmp.Equal(*policyData, *resp.Policy),
	}, nil
}

//formatBucketPolicy parses and formats the bucket.Spec.BucketPolicy struct
func (e *external) formatBucketPolicy(original *v1alpha1.S3BucketPolicy) (*string, error) {
	c := original.DeepCopy()
	iamUsername := aws.StringValue(c.Spec.PolicyBody.UserName)
	accountID, err := e.iamclient.GetAccountID()
	if err != nil {
		return nil, err
	}
	statements := c.Spec.PolicyBody.PolicyStatement
	newStatements := make([]v1alpha1.S3BucketPolicyStatement, 0)
	for _, statement := range statements {
		if statement.ApplyToIAMUser {
			if statement.Principal == nil {
				statement.Principal = &v1alpha1.S3BucketPrincipal{}
			}
			if statement.Principal.AWSPrincipal == nil {
				statement.Principal.AWSPrincipal = make([]string, 0)
			}
			statement.Principal.AWSPrincipal = append(statement.Principal.AWSPrincipal, fmt.Sprintf("arn:aws:iam::%s:user/%s", accountID, iamUsername))
		}
		updatedPaths := make([]string, 0)
		for _, v := range statement.ResourcePath {
			formatted := fmt.Sprintf("arn:aws:s3:::%s", v)
			updatedPaths = append(updatedPaths, formatted)
		}
		statement.ResourcePath = updatedPaths
		newStatements = append(newStatements, statement)
	}
	c.Spec.PolicyBody.PolicyStatement = newStatements
	body, err := c.Spec.PolicyBody.Serialize()
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

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1alpha1.S3BucketPolicy)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	cr.SetConditions(runtimev1alpha1.Creating())

	policyData, err := e.formatBucketPolicy(cr)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errAttach)
	}

	policyString := *policyData
	_, err = e.client.PutBucketPolicyRequest(&awss3.PutBucketPolicyInput{Bucket: cr.Spec.PolicyBody.BucketName, Policy: aws.String(policyString)}).Send(context.TODO())
	return managed.ExternalCreation{}, errors.Wrap(err, errAttach)
}

// Update patches the existing policy for the bucket with the policy in the request body
func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*v1alpha1.S3BucketPolicy)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	policyData, err := e.formatBucketPolicy(cr)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdate)
	}

	_, err = e.client.PutBucketPolicyRequest(&awss3.PutBucketPolicyInput{Bucket: cr.Spec.PolicyBody.BucketName, Policy: aws.String(*policyData)}).Send(context.TODO())
	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdate)
}

// Delete removes the existing policy for a bucket
func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1alpha1.S3BucketPolicy)
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	cr.SetConditions(runtimev1alpha1.Deleting())
	_, err := e.client.DeleteBucketPolicyRequest(&awss3.DeleteBucketPolicyInput{Bucket: cr.Spec.PolicyBody.BucketName}).Send(context.TODO())
	if s3.IsErrorBucketNotFound(err) {
		return nil
	}

	return errors.Wrap(resource.Ignore(s3.IsErrorPolicyNotFound, err), errDelete)
}
