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

package dynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	awsdynamo "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/database/v1alpha1"
	awsv1alpha3 "github.com/crossplane/provider-aws/apis/v1alpha3"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/dynamodb"
)

const (
	errNotDynamoTable   = "managed resource is not an DynamoTable custom resource"
	errKubeUpdateFailed = "cannot update DynamoDB table custom resource"

	errCreateDynamoClient = "cannot create DynamoDB client"
	errGetProvider        = "cannot get provider"
	errGetProviderSecret  = "cannot get provider secret"

	errCreateFailed   = "cannot create DynamoDB table"
	errDeleteFailed   = "cannot delete DynamoDB table"
	errDescribeFailed = "cannot describe DynamoDB table"
	errUpdateFailed   = "cannot update DynamoDB table"
	errUpToDateFailed = "cannot check whether object is up-to-date"
)

// SetupDynamoTable adds a controller that reconciles DynamoTable.
func SetupDynamoTable(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.DynamoTableGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.DynamoTable{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.DynamoTableGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: dynamodb.NewClient}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (dynamodb.Client, error)
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.DynamoTable)
	if !ok {
		return nil, errors.New(errNotDynamoTable)
	}

	p := &awsv1alpha3.Provider{}
	if err := c.kube.Get(ctx, meta.NamespacedNameOf(cr.Spec.ProviderReference), p); err != nil {
		return nil, errors.Wrap(err, errGetProvider)
	}

	if aws.BoolValue(p.Spec.UseServiceAccount) {
		dynamoClient, err := c.newClientFn(ctx, []byte{}, p.Spec.Region, awsclients.UsePodServiceAccount)
		return &external{client: dynamoClient, kube: c.kube}, errors.Wrap(err, errCreateDynamoClient)
	}

	if p.GetCredentialsSecretReference() == nil {
		return nil, errors.New(errGetProviderSecret)
	}

	s := &corev1.Secret{}
	n := types.NamespacedName{Namespace: p.Spec.CredentialsSecretRef.Namespace, Name: p.Spec.CredentialsSecretRef.Name}
	if err := c.kube.Get(ctx, n, s); err != nil {
		return nil, errors.Wrap(err, errGetProviderSecret)
	}

	dynamoClient, err := c.newClientFn(ctx, s.Data[p.Spec.CredentialsSecretRef.Key], p.Spec.Region, awsclients.UseProviderSecret)
	return &external{client: dynamoClient, kube: c.kube}, errors.Wrap(err, errCreateDynamoClient)
}

type external struct {
	client dynamodb.Client
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.DynamoTable)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotDynamoTable)
	}

	rsp, err := e.client.DescribeTableRequest(&awsdynamo.DescribeTableInput{
		TableName: aws.String(meta.GetExternalName(cr)),
	}).Send(ctx)

	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(dynamodb.IsErrorNotFound, err), errDescribeFailed)
	}

	table := rsp.DescribeTableOutput.Table

	current := cr.Spec.ForProvider.DeepCopy()
	dynamodb.LateInitialize(&cr.Spec.ForProvider, table)
	if !cmp.Equal(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errKubeUpdateFailed)
		}
	}
	cr.Status.AtProvider = dynamodb.GenerateObservation(*table)

	switch cr.Status.AtProvider.TableStatus {
	case v1alpha1.DynamoTableStateAvailable:
		cr.Status.SetConditions(runtimev1alpha1.Available())
		resource.SetBindable(cr)
	case v1alpha1.DynamoTableStateCreating:
		cr.Status.SetConditions(runtimev1alpha1.Creating())
	case v1alpha1.DynamoTableStateDeleting:
		cr.Status.SetConditions(runtimev1alpha1.Deleting())
	default:
		cr.Status.SetConditions(runtimev1alpha1.Unavailable())
	}

	upToDate, err := dynamodb.IsUpToDate(cr.Spec.ForProvider, *table)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errUpToDateFailed)
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.DynamoTable)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotDynamoTable)
	}

	cr.SetConditions(runtimev1alpha1.Creating())

	if cr.Status.AtProvider.TableStatus == v1alpha1.DynamoTableStateCreating {
		return managed.ExternalCreation{}, nil
	}

	_, err := e.client.CreateTableRequest(dynamodb.GenerateCreateTableInput(meta.GetExternalName(cr), &cr.Spec.ForProvider)).Send(ctx)
	return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.DynamoTable)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotDynamoTable)
	}

	if cr.Status.AtProvider.TableStatus == v1alpha1.DynamoTableStateModifying {
		return managed.ExternalUpdate{}, nil
	}

	_, err := e.client.UpdateTableRequest(dynamodb.GenerateUpdateTableInput(cr.Status.AtProvider.TableName, &cr.Spec.ForProvider)).Send(ctx)

	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.DynamoTable)
	if !ok {
		return errors.New(errNotDynamoTable)
	}
	cr.SetConditions(runtimev1alpha1.Deleting())
	if cr.Status.AtProvider.TableStatus == v1alpha1.DynamoTableStateDeleting {
		return nil
	}

	_, err := e.client.DeleteTableRequest(&awsdynamo.DeleteTableInput{
		TableName: aws.String(meta.GetExternalName(cr)),
	}).Send(ctx)
	return errors.Wrap(resource.Ignore(dynamodb.IsErrorNotFound, err), errDeleteFailed)
}
