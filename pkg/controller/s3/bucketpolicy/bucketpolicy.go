/*
Copyright 2020 The Crossplane Authors.

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

package bucketpolicy

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/s3/v1alpha3"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/s3"
)

const (
	errUnexpectedObject = "The managed resource is not a BucketPolicy resource"
	errAttach           = "failed to attach the policy to bucket"
	errDelete           = "failed to delete the policy for bucket"
	errGet              = "failed to get BucketPolicy for bucket with name"
	errUpdate           = "failed to update the policy for bucket"
	errNotSpecified     = "failed to format bucketPolicy, no policyBody or jsonBody specified"
)

// SetupBucketPolicy adds a controller that reconciles
// BucketPolicies.
func SetupBucketPolicy(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha3.BucketPolicyGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha3.BucketPolicy{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha3.BucketPolicyGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(),
				newClientFn: s3.NewBucketPolicyClient}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) s3.BucketPolicyClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha3.BucketPolicy)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	cfg, err := awsclient.GetConfig(ctx, c.kube, mg, cr.Spec.PolicyBody.Region)
	if err != nil {
		return nil, err
	}
	return &external{client: c.newClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	client s3.BucketPolicyClient
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1alpha3.BucketPolicy)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	resp, err := e.client.GetBucketPolicyRequest(&awss3.GetBucketPolicyInput{
		Bucket: cr.Spec.Parameters.BucketName,
	}).Send(ctx)
	if err != nil {
		if s3.IsErrorBucketNotFound(err) {
			return managed.ExternalObservation{}, nil
		}
		return managed.ExternalObservation{}, awsclient.Wrap(resource.Ignore(s3.IsErrorPolicyNotFound, err), errGet)
	}

	policyData, err := e.formatBucketPolicy(cr)

	if err != nil {
		return managed.ExternalObservation{}, awsclient.Wrap(resource.Ignore(s3.IsErrorPolicyNotFound, err), errGet)
	}

	cr.SetConditions(xpv1.Available())

	// If our version and the external version are the same, we return ResourceUpToDate: true
	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: cmp.Equal(*policyData, *resp.Policy),
	}, nil
}

// formatBucketPolicy parses and formats the bucket.Spec.BucketPolicy struct
func (e *external) formatBucketPolicy(original *v1alpha2.BucketPolicy) (*string, error) {
	switch {
	case original.Spec.Parameters.JSONBody != nil:
		return original.Spec.Parameters.JSONBody, nil
	case original.Spec.Parameters.PolicyBody != nil:
		c := original.DeepCopy()
		body, err := s3.Serialize(c.Spec.Parameters.PolicyBody)
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
	return nil, errors.New(errNotSpecified)
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1alpha3.BucketPolicy)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	cr.SetConditions(xpv1.Creating())

	policyData, err := e.formatBucketPolicy(cr)
	if err != nil {
		return managed.ExternalCreation{}, awsclient.Wrap(err, errAttach)
	}

	policyString := *policyData
	_, err = e.client.PutBucketPolicyRequest(&awss3.PutBucketPolicyInput{Bucket: cr.Spec.Parameters.BucketName, Policy: aws.String(policyString)}).Send(ctx)
	return managed.ExternalCreation{}, errors.Wrap(err, errAttach)
}

// Update patches the existing policy for the bucket with the policy in the request body
func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*v1alpha3.BucketPolicy)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	policyData, err := e.formatBucketPolicy(cr)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdate)
	}

	_, err = e.client.PutBucketPolicyRequest(&awss3.PutBucketPolicyInput{Bucket: cr.Spec.Parameters.BucketName, Policy: aws.String(*policyData)}).Send(ctx)
	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdate)
}

// Delete removes the existing policy for a bucket
func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1alpha3.BucketPolicy)
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	cr.SetConditions(xpv1.Deleting())
	_, err := e.client.DeleteBucketPolicyRequest(&awss3.DeleteBucketPolicyInput{Bucket: cr.Spec.Parameters.BucketName}).Send(ctx)
	if s3.IsErrorBucketNotFound(err) {
		return nil
	}

	return awsclient.Wrap(resource.Ignore(s3.IsErrorPolicyNotFound, err), errDelete)
}
