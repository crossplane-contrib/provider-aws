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

package cluster

import (
	"context"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
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

	"github.com/crossplane-contrib/provider-aws/apis/eks/v1beta1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/eks"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/tags"
)

const (
	errNotEKSCluster    = "managed resource is not an EKS cluster custom resource"
	errKubeUpdateFailed = "cannot update EKS cluster custom resource"

	errCreateFailed        = "cannot create EKS cluster"
	errUpdateConfigFailed  = "cannot update EKS cluster configuration"
	errUpdateVersionFailed = "cannot update EKS cluster version"
	errAddTagsFailed       = "cannot add tags to EKS cluster"
	errDeleteFailed        = "cannot delete EKS cluster"
	errDescribeFailed      = "cannot describe EKS cluster"
	errPatchCreationFailed = "cannot create a patch object"
	errUpToDateFailed      = "cannot check whether object is up-to-date"
)

// SetupCluster adds a controller that reconciles Clusters.
func SetupCluster(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.ClusterGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: eks.NewEKSClient, newSTSClientFn: eks.NewSTSClient}),
		managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient())),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.ClusterGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1beta1.Cluster{}).
		Complete(r)
}

type connector struct {
	kube           client.Client
	newClientFn    func(config aws.Config) eks.Client
	newSTSClientFn func(config aws.Config) eks.STSClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1beta1.Cluster)
	if !ok {
		return nil, errors.New(errNotEKSCluster)
	}
	cfg, err := connectaws.GetConfig(ctx, c.kube, mg, aws.ToString(cr.Spec.ForProvider.Region))
	if err != nil {
		return nil, err
	}
	return &external{client: c.newClientFn(*cfg), sts: c.newSTSClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	client eks.Client
	sts    eks.STSClient
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1beta1.Cluster)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotEKSCluster)
	}

	rsp, err := e.client.DescribeCluster(ctx, &awseks.DescribeClusterInput{Name: aws.String(meta.GetExternalName(cr))})
	if err != nil {
		return managed.ExternalObservation{}, errorutils.Wrap(resource.Ignore(eks.IsErrorNotFound, err), errDescribeFailed)
	}

	current := cr.Spec.ForProvider.DeepCopy()
	eks.LateInitialize(&cr.Spec.ForProvider, rsp.Cluster)
	if !reflect.DeepEqual(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errKubeUpdateFailed)
		}
	}

	cr.Status.AtProvider = eks.GenerateObservation(rsp.Cluster)
	switch cr.Status.AtProvider.Status { //nolint:exhaustive
	case v1beta1.ClusterStatusActive:
		cr.Status.SetConditions(xpv1.Available())
	case v1beta1.ClusterStatusCreating:
		cr.Status.SetConditions(xpv1.Creating())
	case v1beta1.ClusterStatusDeleting:
		cr.Status.SetConditions(xpv1.Deleting())
	default:
		cr.Status.SetConditions(xpv1.Unavailable())
	}
	upToDate, err := eks.IsUpToDate(&cr.Spec.ForProvider, rsp.Cluster)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errUpToDateFailed)
	}

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  upToDate,
		ConnectionDetails: eks.GetConnectionDetails(ctx, rsp.Cluster, e.sts),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.Cluster)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotEKSCluster)
	}
	cr.SetConditions(xpv1.Creating())
	if cr.Status.AtProvider.Status == v1beta1.ClusterStatusCreating {
		return managed.ExternalCreation{}, nil
	}
	_, err := e.client.CreateCluster(ctx, eks.GenerateCreateClusterInput(meta.GetExternalName(cr), &cr.Spec.ForProvider))
	return managed.ExternalCreation{}, errorutils.Wrap(err, errCreateFailed)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) { //nolint:gocyclo
	cr, ok := mg.(*v1beta1.Cluster)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotEKSCluster)
	}
	switch cr.Status.AtProvider.Status { //nolint:exhaustive
	case v1beta1.ClusterStatusUpdating, v1beta1.ClusterStatusCreating:
		return managed.ExternalUpdate{}, nil
	}

	// NOTE(hasheddan): we have to describe the cluster again because different
	// fields require different update methods.
	rsp, err := e.client.DescribeCluster(ctx, &awseks.DescribeClusterInput{Name: aws.String(meta.GetExternalName(cr))})
	if err != nil || rsp.Cluster == nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(err, errDescribeFailed)
	}
	add, remove := tags.DiffTags(cr.Spec.ForProvider.Tags, rsp.Cluster.Tags)
	if len(remove) != 0 {
		if _, err := e.client.UntagResource(ctx, &awseks.UntagResourceInput{ResourceArn: rsp.Cluster.Arn, TagKeys: remove}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(resource.Ignore(eks.IsErrorInUse, err), errAddTagsFailed)
		}
	}
	if len(add) != 0 {
		if _, err := e.client.TagResource(ctx, &awseks.TagResourceInput{ResourceArn: rsp.Cluster.Arn, Tags: add}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(resource.Ignore(eks.IsErrorInUse, err), errAddTagsFailed)
		}
	}
	patch, err := eks.CreatePatch(rsp.Cluster, &cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(err, errPatchCreationFailed)
	}
	if patch.EncryptionConfig != nil {
		_, err := e.client.AssociateEncryptionConfig(ctx, &awseks.AssociateEncryptionConfigInput{
			ClusterName:      pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
			EncryptionConfig: eks.GenerateEncryptionConfig(&cr.Spec.ForProvider),
		})
		return managed.ExternalUpdate{}, errorutils.Wrap(resource.Ignore(eks.IsErrorInUse, err), errUpdateVersionFailed)
	}
	if patch.Version != nil {
		_, err := e.client.UpdateClusterVersion(ctx, &awseks.UpdateClusterVersionInput{Name: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)), Version: patch.Version})
		return managed.ExternalUpdate{}, errorutils.Wrap(resource.Ignore(eks.IsErrorInUse, err), errUpdateVersionFailed)
	}
	if patch.Logging != nil {
		_, err = e.client.UpdateClusterConfig(ctx, eks.GenerateUpdateClusterConfigInputForLogging(meta.GetExternalName(cr), patch))
		return managed.ExternalUpdate{}, errorutils.Wrap(resource.Ignore(eks.IsErrorInUse, err), errUpdateVersionFailed)
	}
	_, err = e.client.UpdateClusterConfig(ctx, eks.GenerateUpdateClusterConfigInputForVPC(meta.GetExternalName(cr), patch))
	return managed.ExternalUpdate{}, errorutils.Wrap(resource.Ignore(eks.IsErrorInUse, err), errUpdateConfigFailed)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1beta1.Cluster)
	if !ok {
		return errors.New(errNotEKSCluster)
	}
	cr.SetConditions(xpv1.Deleting())
	if cr.Status.AtProvider.Status == v1beta1.ClusterStatusDeleting {
		return nil
	}
	_, err := e.client.DeleteCluster(ctx, &awseks.DeleteClusterInput{Name: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))})
	return errorutils.Wrap(resource.Ignore(eks.IsErrorNotFound, err), errDeleteFailed)
}
