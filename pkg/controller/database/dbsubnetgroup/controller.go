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

package dbsubnetgroup

import (
	"context"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsrds "github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	v1beta1 "github.com/crossplane/provider-aws/apis/database/v1beta1"
	awsv1alpha3 "github.com/crossplane/provider-aws/apis/v1alpha3"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
	dbsg "github.com/crossplane/provider-aws/pkg/clients/dbsubnetgroup"
)

const (
	errKubeUpdateFailed          = "cannot update DBSubnetGroup custom resource"
	errGetProvider               = "cannot get provider"
	errCreateDBSubnetGroupClient = "cannot create DBSubnetGroup client"
	errGetProviderSecret         = "cannot get provider secret"

	errUnexpectedObject = "The managed resource is not an DBSubnetGroup resource"
	errDescribe         = "failed to describe DBSubnetGroup with groupName: %v"
	errMultipleItems    = "retrieved multiple DBSubnetGroups for the given groupName: %v"
	errCreate           = "failed to create the DBSubnetGroup resource with name: %v"
	errDelete           = "failed to delete the DBSubnetGroup resource"
	errUpdate           = "failed to update the DBSubnetGroup resource"
	errUpToDateFailed   = "cannot check whether object is up-to-date"
)

// SetupDBSubnetGroup adds a controller that reconciles DBSubnetGroups.
func SetupDBSubnetGroup(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1beta1.DBSubnetGroupGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1beta1.DBSubnetGroup{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1beta1.DBSubnetGroupGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: dbsg.NewClient}),
			managed.WithConnectionPublishers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (dbsg.Client, error)
}

func (conn *connector) Connect(ctx context.Context, mgd resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mgd.(*v1beta1.DBSubnetGroup)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}

	p := &awsv1alpha3.Provider{}
	if err := conn.kube.Get(ctx, meta.NamespacedNameOf(cr.Spec.ProviderReference), p); err != nil {
		return nil, errors.Wrap(err, errGetProvider)
	}

	if aws.BoolValue(p.Spec.UseServiceAccount) {
		dbSubnetGroupclient, err := conn.newClientFn(ctx, []byte{}, p.Spec.Region, awsclients.UsePodServiceAccount)
		return &external{client: dbSubnetGroupclient, kube: conn.kube}, errors.Wrap(err, errCreateDBSubnetGroupClient)
	}

	if p.GetCredentialsSecretReference() == nil {
		return nil, errors.New(errGetProviderSecret)
	}

	s := &corev1.Secret{}
	n := types.NamespacedName{Namespace: p.Spec.CredentialsSecretRef.Namespace, Name: p.Spec.CredentialsSecretRef.Name}
	if err := conn.kube.Get(ctx, n, s); err != nil {
		return nil, errors.Wrap(err, errGetProviderSecret)
	}

	dbSubnetGroupclient, err := conn.newClientFn(ctx, s.Data[p.Spec.CredentialsSecretRef.Key], p.Spec.Region, awsclients.UseProviderSecret)
	return &external{client: dbSubnetGroupclient, kube: conn.kube}, errors.Wrap(err, errCreateDBSubnetGroupClient)
}

type external struct {
	client dbsg.Client
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	cr, ok := mgd.(*v1beta1.DBSubnetGroup)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	req := e.client.DescribeDBSubnetGroupsRequest(&awsrds.DescribeDBSubnetGroupsInput{
		DBSubnetGroupName: aws.String(meta.GetExternalName(cr)),
	})
	res, err := req.Send(ctx)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(dbsg.IsErrorNotFound, err), errDescribe)
	}

	// in a successful response, there should be one and only one object
	if len(res.DBSubnetGroups) != 1 {
		return managed.ExternalObservation{}, errors.Errorf(errMultipleItems, cr.Spec.ForProvider.DBSubnetGroupName)
	}

	observed := res.DBSubnetGroups[0]
	current := cr.Spec.ForProvider.DeepCopy()
	dbsg.LateInitialize(&cr.Spec.ForProvider, &observed)
	if !reflect.DeepEqual(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errKubeUpdateFailed)
		}
	}
	cr.Status.AtProvider = dbsg.GenerateObservation(observed)

	switch cr.Status.AtProvider.SubnetGroupStatus {
	case v1beta1.DBSubnetGroupStateAvailable:
		cr.Status.SetConditions(runtimev1alpha1.Available())
		resource.SetBindable(cr)
	case v1beta1.DBSubnetGroupStateCreating:
		cr.Status.SetConditions(runtimev1alpha1.Creating())
	case v1beta1.DBSubnetGroupStateDeleting:
		cr.Status.SetConditions(runtimev1alpha1.Deleting())
	default:
		cr.Status.SetConditions(runtimev1alpha1.Unavailable())
	}

	upToDate, err := dbsg.IsDBSubnetGroupUpToDate(cr.Spec.ForProvider, observed)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errUpToDateFailed)
	}

	return managed.ExternalObservation{
		ResourceUpToDate: upToDate,
		ResourceExists:   true,
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1beta1.DBSubnetGroup)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	cr.SetConditions(runtimev1alpha1.Creating())

	if cr.Status.AtProvider.SubnetGroupStatus == v1beta1.DBSubnetGroupStateCreating {
		return managed.ExternalCreation{}, nil
	}

	input := &awsrds.CreateDBSubnetGroupInput{
		DBSubnetGroupDescription: aws.String(cr.Spec.ForProvider.DBSubnetGroupDescription),
		DBSubnetGroupName:        aws.String(meta.GetExternalName(cr)),
		SubnetIds:                cr.Spec.ForProvider.SubnetIDs,
		Tags:                     []awsrds.Tag{},
	}

	_, err := e.client.CreateDBSubnetGroupRequest(input).Send(ctx)
	return managed.ExternalCreation{}, errors.Wrap(err, errCreate)
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*v1beta1.DBSubnetGroup)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	if cr.Status.AtProvider.SubnetGroupStatus == v1beta1.DBSubnetGroupStateModifying {
		return managed.ExternalUpdate{}, nil
	}

	_, err := e.client.ModifyDBSubnetGroupRequest(&awsrds.ModifyDBSubnetGroupInput{
		DBSubnetGroupName:        aws.String(meta.GetExternalName(cr)),
		DBSubnetGroupDescription: aws.String(cr.Spec.ForProvider.DBSubnetGroupDescription),
		SubnetIds:                cr.Spec.ForProvider.SubnetIDs,
	}).Send(ctx)

	return managed.ExternalUpdate{}, errors.Wrap(resource.Ignore(dbsg.IsErrorAlreadyExists, err), errUpdate)
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1beta1.DBSubnetGroup)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.SetConditions(runtimev1alpha1.Deleting())
	if cr.Status.AtProvider.SubnetGroupStatus == v1beta1.DBSubnetGroupStateDeleting {
		return nil
	}

	_, err := e.client.DeleteDBSubnetGroupRequest(&awsrds.DeleteDBSubnetGroupInput{
		DBSubnetGroupName: aws.String(meta.GetExternalName(cr)),
	}).Send(ctx)
	return errors.Wrap(resource.Ignore(dbsg.IsDBSubnetGroupNotFoundErr, err), errDelete)
}
