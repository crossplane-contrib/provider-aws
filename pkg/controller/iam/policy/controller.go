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

package policy

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	awsiamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/iam"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errUnexpectedObject = "The managed resource is not a Policy resource"

	errGet              = "failed to get IAM Policy"
	errCreate           = "failed to create the IAM Policy"
	errDelete           = "failed to delete the IAM Policy"
	errUpdate           = "failed to update the IAM Policy"
	errExternalName     = "failed to update the IAM Policy external-name"
	errEmptyPolicy      = "empty IAM Policy received from IAM API"
	errPolicyVersion    = "No version for policy received from IAM API"
	errUpToDate         = "cannot check if policy is up to date"
	errKubeUpdateFailed = "cannot late initialize IAM Policy"
	errTag              = "cannot tag policy"
	errUntag            = "cannot untag policy"
)

// SetupPolicy adds a controller that reconciles IAM Policy.
func SetupPolicy(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.PolicyGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: iam.NewPolicyClient, newSTSClientFn: iam.NewSTSClient}),
		managed.WithConnectionPublishers(),
		managed.WithInitializers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.PolicyGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1beta1.Policy{}).
		Complete(r)
}

type connector struct {
	kube           client.Client
	newClientFn    func(config aws.Config) iam.PolicyClient
	newSTSClientFn func(config aws.Config) iam.STSClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cfg, err := connectaws.GetConfig(ctx, c.kube, mg, connectaws.GlobalRegion)
	if err != nil {
		return nil, err
	}
	return &external{client: c.newClientFn(*cfg), sts: c.newSTSClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	client iam.PolicyClient
	sts    iam.STSClient
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) { //nolint:gocyclo
	cr, ok := mgd.(*v1beta1.Policy)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if meta.GetExternalName(cr) == "" {
		// If external name not set there is still a change it may already exist
		// Try to get the policy by name
		policyArn, policyErr := e.getPolicyArnByNameAndPath(ctx, cr.Spec.ForProvider.Name, cr.Spec.ForProvider.Path)
		if policyArn == nil || policyErr != nil {
			return managed.ExternalObservation{}, errorutils.Wrap(policyErr, errExternalName)
		}
		meta.SetExternalName(cr, aws.ToString(policyArn))
		_ = e.kube.Update(ctx, cr)
	}

	policyResp, err := e.client.GetPolicy(ctx, &awsiam.GetPolicyInput{
		PolicyArn: aws.String(meta.GetExternalName(cr)),
	})

	if err != nil {
		return managed.ExternalObservation{}, errorutils.Wrap(resource.Ignore(iam.IsErrorNotFound, err), errGet)
	}

	if policyResp.Policy == nil {
		return managed.ExternalObservation{}, errors.New(errEmptyPolicy)
	}
	policy := policyResp.Policy

	cr.SetConditions(xpv1.Available())

	cr.Status.AtProvider = v1beta1.PolicyObservation{
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
		return managed.ExternalObservation{}, errorutils.Wrap(err, errPolicyVersion)
	}

	update, diff, err := iam.IsPolicyUpToDate(cr.Spec.ForProvider, *versionRsp.PolicyVersion)

	if err != nil {
		return managed.ExternalObservation{}, errorutils.Wrap(err, errUpToDate)
	}

	crTagMap := make(map[string]string, len(cr.Spec.ForProvider.Tags))
	for _, v := range cr.Spec.ForProvider.Tags {
		crTagMap[v.Key] = v.Value
	}
	_, _, areRolesUpdated := iam.DiffIAMTags(crTagMap, policyResp.Policy.Tags)

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: update && areRolesUpdated,
		Diff:             diff,
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1beta1.Policy)
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
		return managed.ExternalCreation{}, errorutils.Wrap(err, errCreate)
	}

	meta.SetExternalName(cr, aws.ToString(createOutput.Policy.Arn))

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*v1beta1.Policy)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	// An update to AWS Policy is a new version of that policy.
	// A maximum of 5 versions are allowed. Below, the oldest version is deleted
	// for an update request when 5 versions already exist.
	// The new version is set as default.

	if err := e.deleteOldestVersion(ctx, meta.GetExternalName(cr)); err != nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
	}

	_, err := e.client.CreatePolicyVersion(ctx, &awsiam.CreatePolicyVersionInput{
		PolicyArn:      aws.String(meta.GetExternalName(cr)),
		PolicyDocument: aws.String(cr.Spec.ForProvider.Document),
		SetAsDefault:   true,
	})

	if err != nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
	}

	observed, err := e.client.GetPolicy(ctx, &awsiam.GetPolicyInput{
		PolicyArn: aws.String(meta.GetExternalName(cr)),
	})

	if err != nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(resource.Ignore(iam.IsErrorNotFound, err), errGet)
	}

	add, remove, _ := iam.DiffIAMTagsWithUpdates(cr.Spec.ForProvider.Tags, observed.Policy.Tags)
	if len(add) != 0 {
		if _, err := e.client.TagPolicy(ctx, &awsiam.TagPolicyInput{
			PolicyArn: aws.String(meta.GetExternalName(cr)),
			Tags:      add,
		}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errTag)
		}
	}

	if len(remove) != 0 {
		if _, err := e.client.UntagPolicy(ctx, &awsiam.UntagPolicyInput{
			PolicyArn: aws.String(meta.GetExternalName(cr)),
			TagKeys:   remove,
		}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errUntag)
		}
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1beta1.Policy)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	if err := e.deleteNonDefaultVersions(ctx, meta.GetExternalName(cr)); err != nil {
		return errorutils.Wrap(err, errDelete)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	_, err := e.client.DeletePolicy(ctx, &awsiam.DeletePolicyInput{
		PolicyArn: aws.String(meta.GetExternalName(cr)),
	})

	return errorutils.Wrap(resource.Ignore(iam.IsErrorNotFound, err), errDelete)
}

func (e *external) getCallerIdentityArn(ctx context.Context) (arn.ARN, error) {
	resp, err := e.sts.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return arn.ARN{}, err
	}
	return arn.Parse(aws.ToString(resp.Arn))
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

// getPolicyArnByNameAndPath will attempt to determine the arn for a policy using the current caller identity
func (e *external) getPolicyArnByNameAndPath(ctx context.Context, policyName string, policyPath *string) (*string, error) {

	// Get the ARN of the current identity
	identityArn, err := e.getCallerIdentityArn(ctx)
	if err != nil {
		return nil, err
	}

	// Per the aws docs
	// This parameter is optional. If it is not included, it defaults to a slash (/).
	// This parameter allows (through its regex pattern ) a string of characters consisting
	// of either a forward slash (/) by itself or a string that must begin and end with forward
	// slashes. In addition, it can contain any ASCII character from the ! (\u0021 ) through the
	// DEL character (\u007F ), including most punctuation characters, digits, and upper and lowercased letters.
	if policyPath == nil {
		policyPath = pointer.ToOrNilIfZeroValue("/")
	}

	// Use it to construct an arn for the policy
	policyArn := arn.ARN{Partition: identityArn.Partition,
		Service:   "iam",
		Region:    identityArn.Region,
		AccountID: identityArn.AccountID,
		Resource:  "policy" + pointer.StringValue(policyPath) + policyName}

	return aws.String(policyArn.String()), nil
}
