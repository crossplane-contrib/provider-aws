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

package redshift

import (
	"context"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsredshift "github.com/aws/aws-sdk-go-v2/service/redshift"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/password"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/redshift/v1alpha1"
	awscommon "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/redshift"
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
func SetupCluster(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.ClusterGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.Cluster{}).
		Complete(managed.NewReconciler(
			mgr, resource.ManagedKind(v1alpha1.ClusterGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: redshift.NewClient, awsConfigFn: awscommon.GetConfig}),
			managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient())),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) redshift.Client
	awsConfigFn func(context.Context, client.Client, resource.Managed, string) (*aws.Config, error)
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cfg, err := c.awsConfigFn(ctx, c.kube, mg, "")
	if err != nil {
		return nil, err
	}
	return &external{client: c.newClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	kube   client.Client
	client redshift.Client
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) { //nolint:gocyclo
	cr, ok := mg.(*v1alpha1.Cluster)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	rsp, err := e.client.DescribeClustersRequest(&awsredshift.DescribeClustersInput{
		ClusterIdentifier: aws.String(meta.GetExternalName(cr)),
	}).Send(ctx)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(redshift.IsNotFound, err), errDescribeFailed)
	}

	// Describe requests can be used with filters, which then returns a list.
	// But we use an explicit identifier, so, if there is no error, there should
	// be only 1 element in the list.
	if len(rsp.Clusters) != 1 {
		return managed.ExternalObservation{}, errors.Wrap(errors.New(errMultipleCluster), errMultipleCluster)
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
	case v1alpha1.StateAvailable:
		cr.Status.SetConditions(runtimev1alpha1.Available())
	case v1alpha1.StateCreating:
		cr.Status.SetConditions(runtimev1alpha1.Creating())
	case v1alpha1.StateDeleting:
		cr.Status.SetConditions(runtimev1alpha1.Deleting())
	default:
		cr.Status.SetConditions(runtimev1alpha1.Unavailable())
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
	cr, ok := mg.(*v1alpha1.Cluster)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}
	cr.SetConditions(runtimev1alpha1.Creating())
	if cr.Status.AtProvider.ClusterStatus == v1alpha1.StateCreating {
		return managed.ExternalCreation{}, nil
	}
	pw, err := password.Generate()
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	input := redshift.GenerateCreateClusterInput(&cr.Spec.ForProvider, aws.String(meta.GetExternalName(cr)), aws.String(pw))
	_, err = e.client.CreateClusterRequest(input).Send(ctx)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	conn := managed.ConnectionDetails{
		runtimev1alpha1.ResourceCredentialsSecretPasswordKey: []byte(aws.StringValue(input.MasterUserPassword)),
		runtimev1alpha1.ResourceCredentialsSecretUserKey:     []byte(aws.StringValue(input.MasterUsername)),
	}

	return managed.ExternalCreation{ConnectionDetails: conn}, errors.Wrap(err, errCreateFailed)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Cluster)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}
	switch cr.Status.AtProvider.ClusterStatus {
	case v1alpha1.StateModifying, v1alpha1.StateCreating:
		return managed.ExternalUpdate{}, nil
	}

	rsp, err := e.client.DescribeClustersRequest(&awsredshift.DescribeClustersInput{
		ClusterIdentifier: aws.String(meta.GetExternalName(cr)),
	}).Send(ctx)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(resource.Ignore(redshift.IsNotFound, err), errDescribeFailed)
	}

	_, err = e.client.ModifyClusterRequest(redshift.GenerateModifyClusterInput(&cr.Spec.ForProvider, rsp.Clusters[0])).Send(ctx)

	if err == nil && aws.StringValue(cr.Spec.ForProvider.NewClusterIdentifier) != meta.GetExternalName(cr) {
		meta.SetExternalName(cr, aws.StringValue(cr.Spec.ForProvider.NewClusterIdentifier))

		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errKubeUpdateFailed)
		}
	}

	return managed.ExternalUpdate{}, errors.Wrap(err, errModifyFailed)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Cluster)
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	cr.SetConditions(runtimev1alpha1.Deleting())
	if cr.Status.AtProvider.ClusterStatus == v1alpha1.StateDeleting {
		return nil
	}

	_, err := e.client.DeleteClusterRequest(redshift.GenerateDeleteClusterInput(&cr.Spec.ForProvider, aws.String(meta.GetExternalName(cr)))).Send(ctx)

	return errors.Wrap(resource.Ignore(redshift.IsNotFound, err), errDeleteFailed)
}
