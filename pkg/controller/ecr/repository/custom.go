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
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	awsecr "github.com/aws/aws-sdk-go/service/ecr"
	svcsdk "github.com/aws/aws-sdk-go/service/ecr"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	cpresource "github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/ecr/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

const (
	errTag      = "cannot tag repository"
	errUntag    = "cannot untag repository"
	errListTags = "cannot list tags for repository"
)

// SetupRepository adds a controller that reconciles Repository.
func SetupRepository(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.RepositoryGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.Repository{}).
		Complete(managed.NewReconciler(mgr,
			cpresource.ManagedKind(v1alpha1.RepositoryGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient()), managed.InitializerFn(ecrDefaultTags)),
			managed.WithConnectionPublishers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func lateInitialize(spec *v1alpha1.RepositoryParameters, resp *svcsdk.DescribeRepositoriesOutput) {
	repo := resp.Repositories[0]
	if spec.ImageScanningConfiguration == nil && repo.ImageScanningConfiguration != nil {
		spec.ImageScanningConfiguration = &v1alpha1.ImageScanningConfiguration{
			ScanOnPush: repo.ImageScanningConfiguration.ScanOnPush,
		}
	}
	spec.ImageTagMutability = awsclients.LateInitializeStringPtr(spec.ImageTagMutability, repo.ImageTagMutability)
}

func ecrDefaultTags(_ context.Context, mg cpresource.Managed) error {
	cr, ok := mg.(*v1alpha1.Repository)
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	tagMap := map[string]string{}
	for _, t := range cr.Spec.ForProvider.Tags {
		tagMap[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
	}
	for k, v := range cpresource.GetExternalTags(mg) {
		tagMap[k] = v
	}
	cr.Spec.ForProvider.Tags = make([]*v1alpha1.Tag, len(tagMap))
	i := 0
	for k, v := range tagMap {
		cr.Spec.ForProvider.Tags[i] = &v1alpha1.Tag{Key: aws.String(k), Value: aws.String(v)}
		i++
	}
	sort.Slice(cr.Spec.ForProvider.Tags, func(i, j int) bool {
		return aws.StringValue(cr.Spec.ForProvider.Tags[i].Key) < aws.StringValue(cr.Spec.ForProvider.Tags[j].Key)
	})
	return nil
}

// diffTags returns tags that should be added or removed.
func diffTags(spec []*v1alpha1.Tag, current []*ecr.Tag) (addTags []*ecr.Tag, remove []*string) {
	add := map[string]string{}
	for _, t := range spec {
		if t == nil {
			continue
		}
		add[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
	}
	for _, t := range current {
		if t == nil {
			continue
		}
		if add[aws.StringValue(t.Key)] == "" {
			remove = append(remove, t.Key)
		} else {
			delete(add, aws.StringValue(t.Key))
		}
	}
	for k, v := range add {
		addTags = append(addTags, &ecr.Tag{Key: aws.String(k), Value: aws.String(v)})
	}
	return
}

func (e *external) preObserve(_ context.Context, _ *v1alpha1.Repository) error {
	return nil
}

func (e *external) postObserve(_ context.Context, cr *v1alpha1.Repository, _ *svcsdk.DescribeRepositoriesOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	cr.SetConditions(runtimev1alpha1.Available())
	return obs, err
}

func (e *external) preCreate(_ context.Context, _ *v1alpha1.Repository) error {
	return nil
}

func (e *external) postCreate(_ context.Context, _ *v1alpha1.Repository, _ *svcsdk.CreateRepositoryOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	return cre, err
}

func (e *external) preUpdate(_ context.Context, _ *v1alpha1.Repository) error {
	return nil
}

// TODO(muvaf): Is nolint necessary here?
func (e *external) postUpdate(ctx context.Context, cr *v1alpha1.Repository, _ managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) { // nolint:gocyclo
	if err != nil {
		return managed.ExternalUpdate{}, err
	}
	if cr.Spec.ForProvider.ImageTagMutability != nil {
		if _, err := e.client.PutImageTagMutabilityWithContext(ctx, &svcsdk.PutImageTagMutabilityInput{
			RepositoryName:     aws.String(meta.GetExternalName(cr)),
			ImageTagMutability: cr.Spec.ForProvider.ImageTagMutability,
		}); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, "cannot put image tag mutability")
		}
	}
	if cr.Spec.ForProvider.ImageScanningConfiguration != nil {
		if _, err := e.client.PutImageScanningConfigurationWithContext(ctx, &svcsdk.PutImageScanningConfigurationInput{
			RepositoryName: aws.String(meta.GetExternalName(cr)),
			ImageScanningConfiguration: &svcsdk.ImageScanningConfiguration{
				ScanOnPush: cr.Spec.ForProvider.ImageScanningConfiguration.ScanOnPush,
			},
		}); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, "cannot put image scanning configuration")
		}
	}
	resp, err := e.client.ListTagsForResourceWithContext(ctx, &svcsdk.ListTagsForResourceInput{ResourceArn: cr.Status.AtProvider.RepositoryARN})
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errListTags)
	}
	add, remove := diffTags(cr.Spec.ForProvider.Tags, resp.Tags)
	if len(remove) != 0 {
		if _, err := e.client.UntagResourceWithContext(ctx, &awsecr.UntagResourceInput{ResourceArn: cr.Status.AtProvider.RepositoryARN, TagKeys: remove}); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errUntag)
		}
	}
	if len(add) != 0 {
		if _, err := e.client.TagResourceWithContext(ctx, &awsecr.TagResourceInput{ResourceArn: cr.Status.AtProvider.RepositoryARN, Tags: add}); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errTag)
		}
	}
	return managed.ExternalUpdate{}, nil
}

func preGenerateDescribeRepositoriesInput(_ *v1alpha1.Repository, res *svcsdk.DescribeRepositoriesInput) *svcsdk.DescribeRepositoriesInput {
	return res
}

func postGenerateDescribeRepositoriesInput(cr *v1alpha1.Repository, _ *svcsdk.DescribeRepositoriesInput) *svcsdk.DescribeRepositoriesInput {
	return &svcsdk.DescribeRepositoriesInput{
		RepositoryNames: []*string{aws.String(meta.GetExternalName(cr))},
	}
}

func postGenerateCreateRepositoryInput(cr *v1alpha1.Repository, in *svcsdk.CreateRepositoryInput) *svcsdk.CreateRepositoryInput {
	in.RepositoryName = aws.String(meta.GetExternalName(cr))
	return in
}

func preGenerateCreateRepositoryInput(_ *v1alpha1.Repository, in *svcsdk.CreateRepositoryInput) *svcsdk.CreateRepositoryInput {
	return in
}

func postGenerateDeleteRepositoryInput(cr *v1alpha1.Repository, in *svcsdk.DeleteRepositoryInput) *svcsdk.DeleteRepositoryInput {
	in.RepositoryName = aws.String(meta.GetExternalName(cr))
	return in
}

func preGenerateDeleteRepositoryInput(_ *v1alpha1.Repository, in *svcsdk.DeleteRepositoryInput) *svcsdk.DeleteRepositoryInput {
	return in
}
