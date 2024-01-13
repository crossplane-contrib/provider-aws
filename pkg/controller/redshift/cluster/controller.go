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
	awsredshift "github.com/aws/aws-sdk-go-v2/service/redshift"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/password"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	redshiftv1alpha1 "github.com/crossplane-contrib/provider-aws/apis/redshift/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/redshift"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errUnexpectedObject = "managed resource is not a Redshift custom resource"
	errKubeUpdateFailed = "cannot update Redshift cluster custom resource"
	errMultipleCluster  = "multiple clusters with the same name found"
	errCreateFailed     = "cannot create Redshift cluster"
	errModifyFailed     = "cannot modify Redshift cluster"
	errDeleteFailed     = "cannot delete Redshift cluster"
	errDescribeFailed   = "cannot describe Redshift cluster"
	errUpToDateFailed   = "cannot check whether object is up-to-date"
)

// SetupCluster adds a controller that reconciles Redshift clusters.
func SetupCluster(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(redshiftv1alpha1.ClusterGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: redshift.NewClient}),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(
		mgr, resource.ManagedKind(redshiftv1alpha1.ClusterGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&redshiftv1alpha1.Cluster{}).
		Complete(r)
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) redshift.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*redshiftv1alpha1.Cluster)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	cfg, err := connectaws.GetConfig(ctx, c.kube, mg, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, err
	}
	return &external{client: c.newClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	kube   client.Client
	client redshift.Client
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*redshiftv1alpha1.Cluster)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	rsp, err := e.client.DescribeClusters(ctx, &awsredshift.DescribeClustersInput{
		ClusterIdentifier: aws.String(meta.GetExternalName(cr)),
	})
	if err != nil {
		return managed.ExternalObservation{}, errorutils.Wrap(resource.Ignore(redshift.IsNotFound, err), errDescribeFailed)
	}

	// Describe requests can be used with filters, which then returns a list.
	// But we use an explicit identifier, so, if there is no error, there should
	// be only 1 element in the list.
	if len(rsp.Clusters) != 1 {
		return managed.ExternalObservation{}, errors.New(errMultipleCluster)
	}
	instance := rsp.Clusters[0]
	current := cr.Spec.ForProvider.DeepCopy()
	redshift.LateInitialize(&cr.Spec.ForProvider, &instance)
	if !reflect.DeepEqual(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errKubeUpdateFailed)
		}
	}

	cr.Status.AtProvider = redshift.GenerateObservation(rsp.Clusters[0])
	switch cr.Status.AtProvider.ClusterStatus {
	case redshiftv1alpha1.StateAvailable:
		cr.Status.SetConditions(xpv1.Available())
	case redshiftv1alpha1.StateCreating:
		cr.Status.SetConditions(xpv1.Creating())
	case redshiftv1alpha1.StateDeleting:
		cr.Status.SetConditions(xpv1.Deleting())
	default:
		cr.Status.SetConditions(xpv1.Unavailable())
	}

	updated, err := redshift.IsUpToDate(cr.Spec.ForProvider, instance)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errUpToDateFailed)
	}

	return managed.ExternalObservation{
		ResourceUpToDate:  updated,
		ResourceExists:    true,
		ConnectionDetails: redshift.GetConnectionDetails(*cr),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*redshiftv1alpha1.Cluster)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}
	cr.SetConditions(xpv1.Creating())
	if cr.Status.AtProvider.ClusterStatus == redshiftv1alpha1.StateCreating {
		return managed.ExternalCreation{}, nil
	}
	pw, err := password.Generate()
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	input := redshift.GenerateCreateClusterInput(&cr.Spec.ForProvider, aws.String(meta.GetExternalName(cr)), aws.String(pw))
	_, err = e.client.CreateCluster(ctx, input)
	if err != nil {
		return managed.ExternalCreation{}, errorutils.Wrap(err, errCreateFailed)
	}

	conn := managed.ConnectionDetails{
		xpv1.ResourceCredentialsSecretPasswordKey: []byte(aws.ToString(input.MasterUserPassword)),
		xpv1.ResourceCredentialsSecretUserKey:     []byte(aws.ToString(input.MasterUsername)),
	}

	return managed.ExternalCreation{ConnectionDetails: conn}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*redshiftv1alpha1.Cluster)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}
	switch cr.Status.AtProvider.ClusterStatus {
	case redshiftv1alpha1.StateModifying, redshiftv1alpha1.StateCreating:
		return managed.ExternalUpdate{}, nil
	}

	rsp, err := e.client.DescribeClusters(ctx, &awsredshift.DescribeClustersInput{
		ClusterIdentifier: aws.String(meta.GetExternalName(cr)),
	})
	if err != nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(resource.Ignore(redshift.IsNotFound, err), errDescribeFailed)
	}

	_, err = e.client.ModifyCluster(ctx, redshift.GenerateModifyClusterInput(&cr.Spec.ForProvider, rsp.Clusters[0]))

	if err == nil && aws.ToString(cr.Spec.ForProvider.NewClusterIdentifier) != meta.GetExternalName(cr) {
		meta.SetExternalName(cr, aws.ToString(cr.Spec.ForProvider.NewClusterIdentifier))

		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errKubeUpdateFailed)
		}
	}

	return managed.ExternalUpdate{}, errorutils.Wrap(err, errModifyFailed)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*redshiftv1alpha1.Cluster)
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	cr.SetConditions(xpv1.Deleting())
	if cr.Status.AtProvider.ClusterStatus == redshiftv1alpha1.StateDeleting {
		return nil
	}

	_, err := e.client.DeleteCluster(ctx, redshift.GenerateDeleteClusterInput(&cr.Spec.ForProvider, aws.String(meta.GetExternalName(cr))))

	return errorutils.Wrap(resource.Ignore(redshift.IsNotFound, err), errDeleteFailed)
}
