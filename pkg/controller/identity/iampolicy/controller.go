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

package iampolicy

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	awsiamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/pkg/errors"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/identity/v1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/iam"
)

const (
	errUnexpectedObject = "The managed resource is not a IAMPolicy resource"

	errGet           = "failed to get IAM Policy"
	errCreate        = "failed to create the IAM Policy"
	errDelete        = "failed to delete the IAM Policy"
	errUpdate        = "failed to update the IAM Policy"
	errEmptyPolicy   = "empty IAM Policy received from IAM API"
	errPolicyVersion = "No version for policy received from IAM API"
	errUpToDate      = "cannt check if policy is up to date"
)

// SetupIAMPolicy adds a controller that reconciles IAM Policy.
func SetupIAMPolicy(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(v1alpha1.IAMPolicyGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewController(rl),
		}).
		For(&v1alpha1.IAMPolicy{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.IAMPolicyGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: iam.NewPolicyClient}),
			managed.WithInitializers(),
			managed.WithConnectionPublishers(),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) iam.PolicyClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cfg, err := awsclient.GetConfig(ctx, c.kube, mg, awsclient.GlobalRegion)
	if err != nil {
		return nil, err
	}
	return &external{client: c.newClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	client iam.PolicyClient
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1alpha1.IAMPolicy)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{}, nil
	}

	policyResp, err := e.client.GetPolicy(ctx, &awsiam.GetPolicyInput{
		PolicyArn: aws.String(meta.GetExternalName(cr)),
	})

	if err != nil {
		return managed.ExternalObservation{}, awsclient.Wrap(resource.Ignore(iam.IsErrorNotFound, err), errGet)
	}

	if policyResp.Policy == nil {
		return managed.ExternalObservation{}, errors.New(errEmptyPolicy)
	}
	policy := policyResp.Policy

	cr.SetConditions(xpv1.Available())

	cr.Status.AtProvider = v1alpha1.IAMPolicyObservation{
		ARN:                           aws.ToString(policy.Arn),
		AttachmentCount:               aws.ToInt32(policy.AttachmentCount),
		DefaultVersionID:              aws.ToString(policy.DefaultVersionId),
		IsAttachable:                  policy.IsAttachable,
		PermissionsBoundaryUsageCount: aws.ToInt32(policy.PermissionsBoundaryUsageCount),
		PolicyID:                      aws.ToString(policy.PolicyId),
	}

	versionRsp, err := e.client.GetPolicyVersion(ctx, &awsiam.GetPolicyVersionInput{
		PolicyArn: aws.String(meta.GetExternalName(cr)),
		VersionId: aws.String(cr.Status.AtProvider.DefaultVersionID),
	})

	if err != nil || versionRsp.PolicyVersion == nil {
		return managed.ExternalObservation{}, awsclient.Wrap(err, errPolicyVersion)
	}

	update, err := iam.IsPolicyUpToDate(cr.Spec.ForProvider, *versionRsp.PolicyVersion)

	if err != nil {
		return managed.ExternalObservation{}, awsclient.Wrap(err, errUpToDate)
	}

	crTagMap := make(map[string]string, len(cr.Spec.ForProvider.Tags))
	for _, v := range cr.Spec.ForProvider.Tags {
		crTagMap[v.Key] = v.Value
	}
	_, _, areRolesUpdated := iam.DiffIAMTags(crTagMap, policyResp.Policy.Tags)

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: update && areRolesUpdated,
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1alpha1.IAMPolicy)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	tags := cr.Spec.ForProvider.Tags
	inputPolicyTags := make([]awsiamtypes.Tag, len(tags))
	for i := range tags {
		inputPolicyTags[i] = awsiamtypes.Tag{
			Key:   &tags[i].Key,
			Value: &tags[i].Value,
		}
	}

	createOutput, err := e.client.CreatePolicy(ctx, &awsiam.CreatePolicyInput{
		Description:    cr.Spec.ForProvider.Description,
		Path:           cr.Spec.ForProvider.Path,
		PolicyDocument: aws.String(cr.Spec.ForProvider.Document),
		PolicyName:     aws.String(cr.Spec.ForProvider.Name),
		Tags:           inputPolicyTags,
	})

	if err != nil {
		return managed.ExternalCreation{}, awsclient.Wrap(err, errCreate)
	}

	meta.SetExternalName(cr, aws.ToString(createOutput.Policy.Arn))

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*v1alpha1.IAMPolicy)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	// An update to AWS Policy is a new version of that policy.
	// A maximum of 5 versions are allowed. Below, the oldest version is deleted
	// for an update request when 5 versions already exist.
	// The new version is set as default.

	if err := e.deleteOldestVersion(ctx, meta.GetExternalName(cr)); err != nil {
		return managed.ExternalUpdate{}, awsclient.Wrap(err, errUpdate)
	}

	_, err := e.client.CreatePolicyVersion(ctx, &awsiam.CreatePolicyVersionInput{
		PolicyArn:      aws.String(meta.GetExternalName(cr)),
		PolicyDocument: aws.String(cr.Spec.ForProvider.Document),
		SetAsDefault:   true,
	})

	if err != nil {
		return managed.ExternalUpdate{}, awsclient.Wrap(err, errUpdate)
	}

	observed, err := e.client.GetPolicy(ctx, &awsiam.GetPolicyInput{
		PolicyArn: aws.String(meta.GetExternalName(cr)),
	})

	if err != nil {
		return managed.ExternalUpdate{}, awsclient.Wrap(resource.Ignore(iam.IsErrorNotFound, err), errGet)
	}

	crTagMap := make(map[string]string, len(cr.Spec.ForProvider.Tags))
	for _, v := range cr.Spec.ForProvider.Tags {
		crTagMap[v.Key] = v.Value
	}

	add, remove, _ := iam.DiffIAMTags(crTagMap, observed.Policy.Tags)
	if len(add) != 0 {
		if _, err := e.client.TagPolicy(ctx, &awsiam.TagPolicyInput{
			PolicyArn: aws.String(meta.GetExternalName(cr)),
			Tags:      add,
		}); err != nil {
			return managed.ExternalUpdate{}, awsclient.Wrap(err, "cannot tag policy")
		}
	}

	if len(remove) != 0 {
		if _, err := e.client.UntagPolicy(ctx, &awsiam.UntagPolicyInput{
			PolicyArn: aws.String(meta.GetExternalName(cr)),
			TagKeys:   remove,
		}); err != nil {
			return managed.ExternalUpdate{}, awsclient.Wrap(err, "cannot untag policy")
		}
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1alpha1.IAMPolicy)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	if err := e.deleteNonDefaultVersions(ctx, meta.GetExternalName(cr)); err != nil {
		return awsclient.Wrap(err, errDelete)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	_, err := e.client.DeletePolicy(ctx, &awsiam.DeletePolicyInput{
		PolicyArn: aws.String(meta.GetExternalName(cr)),
	})

	return awsclient.Wrap(resource.Ignore(iam.IsErrorNotFound, err), errDelete)
}

func (e *external) listPolicyVersions(ctx context.Context, policyArn string) ([]awsiamtypes.PolicyVersion, error) {
	resp, err := e.client.ListPolicyVersions(ctx, &awsiam.ListPolicyVersionsInput{
		PolicyArn: aws.String(policyArn),
	})

	if err != nil || resp.Versions == nil {
		return nil, err
	}

	return resp.Versions, nil
}

func (e *external) deleteOldestVersion(ctx context.Context, arn string) error {
	allVersions, err := e.listPolicyVersions(ctx, arn)
	if err != nil {
		return err
	}

	if len(allVersions) < 5 {
		return nil
	}

	var oldestVersion awsiamtypes.PolicyVersion

	// loop through all version to find the oldest version.
	for _, version := range allVersions {
		if version.IsDefaultVersion {
			continue
		}
		if oldestVersion.CreateDate == nil ||
			version.CreateDate.Before(*oldestVersion.CreateDate) {
			oldestVersion = version
		}
	}

	_, err = e.client.DeletePolicyVersion(ctx, &awsiam.DeletePolicyVersionInput{
		PolicyArn: aws.String(arn),
		VersionId: oldestVersion.VersionId,
	})

	return err
}

func (e *external) deleteNonDefaultVersions(ctx context.Context, policyArn string) error {
	allVersions, err := e.listPolicyVersions(ctx, policyArn)
	if err != nil {
		return err
	}

	// loop through all the version and delete all non-default versions.
	for _, version := range allVersions {
		if version.IsDefaultVersion {
			continue
		}
		if _, err := e.client.DeletePolicyVersion(ctx, &awsiam.DeletePolicyVersionInput{
			PolicyArn: aws.String(policyArn),
			VersionId: version.VersionId,
		}); err != nil {
			return err
		}
	}

	return nil
}
