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

package database

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	rds2 "github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/meta"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	"github.com/crossplaneio/crossplane-runtime/pkg/util"

	"github.com/crossplaneio/stack-aws/apis/database/v1beta1"
	awsv1alpha2 "github.com/crossplaneio/stack-aws/apis/v1alpha2"
	"github.com/crossplaneio/stack-aws/pkg/clients/rds"
)

const (
	errNotRDSInstance   = "managed resource is not an RDS instance custom resource"
	errKubeUpdateFailed = "cannot update RDS instance custom resource"

	errCreateRDSClient   = "cannot create RDS client"
	errGetProvider       = "cannot get provider"
	errGetProviderSecret = "cannot get provider secret"

	errCreateFailed   = "cannot create RDS instance"
	errModifyFailed   = "cannot modify RDS instance"
	errDeleteFailed   = "cannot delete RDS instance"
	errDescribeFailed = "cannot describe RDS instance"
)

// RDSInstanceController is responsible for adding the RDSInstance
// controller and its corresponding reconciler to the manager with any runtime configuration.
type RDSInstanceController struct{}

// SetupWithManager creates a new Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func (c *RDSInstanceController) SetupWithManager(mgr ctrl.Manager) error {
	r := resource.NewManagedReconciler(mgr,
		resource.ManagedKind(v1beta1.RDSInstanceGroupVersionKind),
		resource.WithExternalConnecter(&connector{
			kube:        mgr.GetClient(),
			newClientFn: rds.NewClient,
		}))

	name := strings.ToLower(fmt.Sprintf("%s.%s", v1beta1.RDSInstanceKind, v1beta1.Group))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1beta1.RDSInstance{}).
		Complete(r)
}

type connector struct {
	kube        client.Client
	newClientFn func(credentials []byte, region string) (rds.Client, error)
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (resource.ExternalClient, error) {
	cr, ok := mg.(*v1beta1.RDSInstance)
	if !ok {
		return nil, errors.New(errNotRDSInstance)
	}

	p := &awsv1alpha2.Provider{}
	if err := c.kube.Get(ctx, meta.NamespacedNameOf(cr.Spec.ProviderReference), p); err != nil {
		return nil, errors.Wrap(err, errGetProvider)
	}

	s := &corev1.Secret{}
	n := types.NamespacedName{Namespace: p.Spec.Secret.Namespace, Name: p.Spec.Secret.Name}
	if err := c.kube.Get(ctx, n, s); err != nil {
		return nil, errors.Wrap(err, errGetProviderSecret)
	}

	rdsClient, err := c.newClientFn(s.Data[p.Spec.Secret.Key], p.Spec.Region)
	return &external{client: rdsClient, kube: c.kube}, errors.Wrap(err, errCreateRDSClient)
}

type external struct {
	client rds.Client
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (resource.ExternalObservation, error) {
	cr, ok := mg.(*v1beta1.RDSInstance)
	if !ok {
		return resource.ExternalObservation{}, errors.New(errNotRDSInstance)
	}
	req := e.client.DescribeDBInstancesRequest(&rds2.DescribeDBInstancesInput{DBInstanceIdentifier: aws.String(meta.GetExternalName(cr))})
	req.SetContext(ctx)
	rsp, err := req.Send()
	if err != nil {
		return resource.ExternalObservation{}, errors.Wrap(resource.Ignore(rds.IsErrorNotFound, err), errDescribeFailed)
	}

	// Describe requests can be used with filters, which then returns a list.
	// But we use an explicit identifier, so, if there is no error, there should
	// be only 1 element in the list.
	instance := rsp.DBInstances[0]
	current := cr.Spec.ForProvider.DeepCopy()
	rds.LateInitialize(&cr.Spec.ForProvider, instance)
	if !reflect.DeepEqual(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return resource.ExternalObservation{}, errors.Wrap(err, errKubeUpdateFailed)
		}
	}
	cr.Status.AtProvider = rds.GenerateObservation(instance)

	switch cr.Status.AtProvider.DBInstanceStatus {
	case string(v1beta1.RDSInstanceStateAvailable):
		cr.Status.SetConditions(runtimev1alpha1.Available())
		resource.SetBindable(cr)
	case string(v1beta1.RDSInstanceStateCreating):
		cr.Status.SetConditions(runtimev1alpha1.Creating())
	case string(v1beta1.RDSInstanceStateDeleting):
		cr.Status.SetConditions(runtimev1alpha1.Deleting())
	default:
		cr.Status.SetConditions(runtimev1alpha1.Unavailable())
	}

	return resource.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  rds.IsUpToDate(cr.Spec.ForProvider, instance),
		ConnectionDetails: rds.GetConnectionDetails(*cr),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (resource.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.RDSInstance)
	if !ok {
		return resource.ExternalCreation{}, errors.New(errNotRDSInstance)
	}
	// generate new password
	password, err := util.GeneratePassword(20)
	if err != nil {
		return resource.ExternalCreation{}, err
	}
	req := e.client.CreateDBInstanceRequest(rds.GenerateCreateDBInstanceInput(meta.GetExternalName(cr), password, &cr.Spec.ForProvider))
	req.SetContext(ctx)
	_, err = req.Send()
	if err != nil {
		return resource.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}
	conn := resource.ConnectionDetails{
		runtimev1alpha1.ResourceCredentialsSecretPasswordKey: []byte(password),
	}
	if cr.Spec.ForProvider.MasterUsername != nil {
		conn[runtimev1alpha1.ResourceCredentialsSecretUserKey] = []byte(aws.StringValue(cr.Spec.ForProvider.MasterUsername))
	}
	return resource.ExternalCreation{ConnectionDetails: conn}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (resource.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta1.RDSInstance)
	if !ok {
		return resource.ExternalUpdate{}, errors.New(errNotRDSInstance)
	}
	req := e.client.ModifyDBInstanceRequest(rds.GenerateModifyDBInstanceInput(meta.GetExternalName(cr), &cr.Spec.ForProvider))
	req.SetContext(ctx)
	_, err := req.Send()
	return resource.ExternalUpdate{}, errors.Wrap(err, errModifyFailed)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1beta1.RDSInstance)
	if !ok {
		return errors.New(errNotRDSInstance)
	}
	input := rds2.DeleteDBInstanceInput{
		DBInstanceIdentifier: aws.String(meta.GetExternalName(cr)),
		SkipFinalSnapshot:    cr.Spec.ForProvider.SkipFinalSnapshotBeforeDeletion,
	}
	req := e.client.DeleteDBInstanceRequest(&input)
	req.SetContext(ctx)
	_, err := req.Send()
	return errors.Wrap(resource.Ignore(rds.IsErrorNotFound, err), errDeleteFailed)
}
