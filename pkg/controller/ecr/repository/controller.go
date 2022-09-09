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

package repository

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsecr "github.com/aws/aws-sdk-go-v2/service/ecr"
	awsecrtypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-aws/apis/ecr/v1beta1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	awsclient "github.com/crossplane-contrib/provider-aws/pkg/clients"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/ecr"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
)

const (
	errUnexpectedObject = "managed resource is not an repository resource"
	errKubeUpdateFailed = "cannot update repository custom resource"

	errDescribe            = "failed to describe repository with id"
	errMultipleItems       = "retrieved multiple repository for the given ECR name"
	errCreate              = "failed to create the repository resource"
	errCreateTags          = "failed to create tags for the repository resource"
	errRemoveTags          = "failed to remove tags for the repository resource"
	errListTags            = "failed to list tags for the repository resource"
	errDelete              = "failed to delete the repository resource"
	errSpecUpdate          = "cannot update spec of repository custom resource"
	errStatusUpdate        = "cannot update status of repository custom resource"
	errUpdateScan          = "failed to update scan config for repository resource"
	errUpdateMutability    = "failed to update mutability for repository resource"
	errPatchCreationFailed = "cannot create a patch object"
)

// SetupRepository adds a controller that reconciles ECR.
func SetupRepository(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.RepositoryGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1beta1.Repository{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1beta1.RepositoryGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithConnectionPublishers(),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient()), managed.NewNameAsExternalName(mgr.GetClient()), &tagger{kube: mgr.GetClient()}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

type connector struct {
	kube client.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1beta1.Repository)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	cfg, err := awsclient.GetConfig(ctx, c.kube, mg, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, err
	}
	return &external{client: awsecr.NewFromConfig(*cfg), kube: c.kube}, nil
}

type external struct {
	kube   client.Client
	client ecr.RepositoryClient
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1beta1.Repository)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	response, err := e.client.DescribeRepositories(ctx, &awsecr.DescribeRepositoriesInput{
		RepositoryNames: []string{meta.GetExternalName(cr)},
	})
	if err != nil {
		return managed.ExternalObservation{}, awsclient.Wrap(resource.Ignore(ecr.IsRepoNotFoundErr, err), errDescribe)
	}

	// in a successful response, there should be one and only one object
	if len(response.Repositories) != 1 {
		return managed.ExternalObservation{}, errors.New(errMultipleItems)
	}

	observed := response.Repositories[0]
	tagsResp, err := e.client.ListTagsForResource(ctx, &awsecr.ListTagsForResourceInput{
		ResourceArn: observed.RepositoryArn,
	})
	if err != nil {
		return managed.ExternalObservation{}, awsclient.Wrap(resource.Ignore(ecr.IsRepoNotFoundErr, err), errListTags)
	}
	// update the CRD spec for any new values from provider
	current := cr.Spec.ForProvider.DeepCopy()
	ecr.LateInitializeRepository(&cr.Spec.ForProvider, &observed)
	if !cmp.Equal(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, awsclient.Wrap(err, errSpecUpdate)
		}
	}

	cr.SetConditions(xpv1.Available())

	cr.Status.AtProvider = ecr.GenerateRepositoryObservation(observed)

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: ecr.IsRepositoryUpToDate(&cr.Spec.ForProvider, tagsResp.Tags, &observed),
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1beta1.Repository)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Creating())
	if err := e.kube.Status().Update(ctx, cr); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errStatusUpdate)
	}

	_, err := e.client.CreateRepository(ctx, ecr.GenerateCreateRepositoryInput(meta.GetExternalName(cr), &cr.Spec.ForProvider))
	if err != nil {
		return managed.ExternalCreation{}, awsclient.Wrap(err, errCreate)
	}
	return managed.ExternalCreation{}, errors.Wrap(e.kube.Update(ctx, cr), errSpecUpdate)
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*v1beta1.Repository)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	err := e.updateTags(ctx, cr)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	response, err := e.client.DescribeRepositories(ctx, &awsecr.DescribeRepositoriesInput{
		RepositoryNames: []string{meta.GetExternalName(cr)},
	})
	if err != nil {
		return managed.ExternalUpdate{}, awsclient.Wrap(resource.Ignore(ecr.IsRepoNotFoundErr, err), errDescribe)
	}

	// in a successful response, there should be one and only one object
	if len(response.Repositories) != 1 {
		return managed.ExternalUpdate{}, errors.New(errMultipleItems)
	}

	observed := response.Repositories[0]

	patch, err := ecr.CreatePatch(&observed, &cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errPatchCreationFailed)
	}

	if patch.ImageTagMutability != nil {
		_, err := e.client.PutImageTagMutability(ctx, &awsecr.PutImageTagMutabilityInput{
			RepositoryName:     awsclient.String(meta.GetExternalName(cr)),
			ImageTagMutability: awsecrtypes.ImageTagMutability(aws.ToString(patch.ImageTagMutability)),
		})
		if err != nil {
			return managed.ExternalUpdate{}, awsclient.Wrap(resource.Ignore(ecr.IsRepoNotFoundErr, err), errUpdateMutability)
		}
	}

	if patch.ImageScanningConfiguration != nil {
		_, err := e.client.PutImageScanningConfiguration(ctx, &awsecr.PutImageScanningConfigurationInput{
			RepositoryName: awsclient.String(meta.GetExternalName(cr)),
			ImageScanningConfiguration: &awsecrtypes.ImageScanningConfiguration{
				ScanOnPush: patch.ImageScanningConfiguration.ScanOnPush,
			},
		})
		if err != nil {
			return managed.ExternalUpdate{}, awsclient.Wrap(resource.Ignore(ecr.IsRepoNotFoundErr, err), errUpdateScan)
		}
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1beta1.Repository)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Deleting())
	_, err := e.client.DeleteRepository(ctx, &awsecr.DeleteRepositoryInput{
		RepositoryName: aws.String(meta.GetExternalName(cr)),
		Force:          aws.ToBool(cr.Spec.ForProvider.ForceDelete),
	})
	return awsclient.Wrap(resource.Ignore(ecr.IsRepoNotFoundErr, err), errDelete)
}

type tagger struct {
	kube client.Client
}

func (t *tagger) Initialize(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1beta1.Repository)
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	added := false
	tagMap := map[string]string{}
	for _, t := range cr.Spec.ForProvider.Tags {
		tagMap[t.Key] = t.Value
	}
	for k, v := range resource.GetExternalTags(mgd) {
		if tagMap[k] != v {
			cr.Spec.ForProvider.Tags = append(cr.Spec.ForProvider.Tags, v1beta1.Tag{Key: k, Value: v})
			added = true
		}
	}
	if !added {
		return nil
	}
	return errors.Wrap(t.kube.Update(ctx, cr), errKubeUpdateFailed)
}

func (e *external) updateTags(ctx context.Context, repo *v1beta1.Repository) error {
	resp, err := e.client.ListTagsForResource(ctx, &awsecr.ListTagsForResourceInput{ResourceArn: &repo.Status.AtProvider.RepositoryArn})
	if err != nil {
		return awsclient.Wrap(err, errListTags)
	}
	add, remove := ecr.DiffTags(repo.Spec.ForProvider.Tags, resp.Tags)
	if len(remove) != 0 {
		if _, err := e.client.UntagResource(ctx, &awsecr.UntagResourceInput{ResourceArn: &repo.Status.AtProvider.RepositoryArn, TagKeys: remove}); err != nil {
			return awsclient.Wrap(err, errRemoveTags)
		}
	}
	if len(add) != 0 {
		if _, err := e.client.TagResource(ctx, &awsecr.TagResourceInput{ResourceArn: &repo.Status.AtProvider.RepositoryArn, Tags: add}); err != nil {
			return awsclient.Wrap(err, errCreateTags)
		}
	}
	return nil
}
