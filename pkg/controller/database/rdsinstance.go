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
	awsrds "github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/meta"
	"github.com/crossplaneio/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	"github.com/crossplaneio/crossplane-runtime/pkg/util"

	"github.com/crossplaneio/stack-aws/apis/database/v1beta1"
	awsv1alpha3 "github.com/crossplaneio/stack-aws/apis/v1alpha3"
	"github.com/crossplaneio/stack-aws/pkg/clients/rds"
)

const (
	errNotRDSInstance   = "managed resource is not an RDS instance custom resource"
	errKubeUpdateFailed = "cannot update RDS instance custom resource"

	errCreateRDSClient   = "cannot create RDS client"
	errGetProvider       = "cannot get provider"
	errGetProviderSecret = "cannot get provider secret"

	errCreateFailed        = "cannot create RDS instance"
	errModifyFailed        = "cannot modify RDS instance"
	errDeleteFailed        = "cannot delete RDS instance"
	errDescribeFailed      = "cannot describe RDS instance"
	errPatchCreationFailed = "cannot create a patch object"
	errUpToDateFailed      = "cannot check whether object is up-to-date"
)

// RDSInstanceController is responsible for adding the RDSInstance
// controller and its corresponding reconciler to the manager with any runtime configuration.
type RDSInstanceController struct{}

// SetupWithManager creates a new Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func (c *RDSInstanceController) SetupWithManager(mgr ctrl.Manager) error {
	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.RDSInstanceGroupVersionKind),
		managed.WithExternalConnecter(&connector{
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

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1beta1.RDSInstance)
	if !ok {
		return nil, errors.New(errNotRDSInstance)
	}

	p := &awsv1alpha3.Provider{}
	if err := c.kube.Get(ctx, meta.NamespacedNameOf(cr.Spec.ProviderReference), p); err != nil {
		return nil, errors.Wrap(err, errGetProvider)
	}

	s := &corev1.Secret{}
	n := types.NamespacedName{Namespace: p.Spec.CredentialsSecretRef.Namespace, Name: p.Spec.CredentialsSecretRef.Name}
	if err := c.kube.Get(ctx, n, s); err != nil {
		return nil, errors.Wrap(err, errGetProviderSecret)
	}

	rdsClient, err := c.newClientFn(s.Data[p.Spec.CredentialsSecretRef.Key], p.Spec.Region)
	return &external{client: rdsClient, kube: c.kube}, errors.Wrap(err, errCreateRDSClient)
}

type external struct {
	client rds.Client
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1beta1.RDSInstance)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotRDSInstance)
	}
	// TODO(muvaf): There are some parameters that require a specific call
	// for retrieval. For example, DescribeDBInstancesOutput does not expose
	// the tags map of the RDS instance, you have to make ListTagsForResourceRequest
	req := e.client.DescribeDBInstancesRequest(&awsrds.DescribeDBInstancesInput{DBInstanceIdentifier: aws.String(meta.GetExternalName(cr))})
	req.SetContext(ctx)
	rsp, err := req.Send()
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(rds.IsErrorNotFound, err), errDescribeFailed)
	}

	// Describe requests can be used with filters, which then returns a list.
	// But we use an explicit identifier, so, if there is no error, there should
	// be only 1 element in the list.
	instance := rsp.DBInstances[0]
	current := cr.Spec.ForProvider.DeepCopy()
	rds.LateInitialize(&cr.Spec.ForProvider, &instance)
	if !reflect.DeepEqual(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errKubeUpdateFailed)
		}
	}
	cr.Status.AtProvider = rds.GenerateObservation(instance)

	switch cr.Status.AtProvider.DBInstanceStatus {
	case v1beta1.RDSInstanceStateAvailable:
		cr.Status.SetConditions(runtimev1alpha1.Available())
		resource.SetBindable(cr)
	case v1beta1.RDSInstanceStateCreating:
		cr.Status.SetConditions(runtimev1alpha1.Creating())
	case v1beta1.RDSInstanceStateDeleting:
		cr.Status.SetConditions(runtimev1alpha1.Deleting())
	default:
		cr.Status.SetConditions(runtimev1alpha1.Unavailable())
	}
	upToDate, err := rds.IsUpToDate(cr.Spec.ForProvider, instance)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errUpToDateFailed)
	}

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  upToDate,
		ConnectionDetails: rds.GetConnectionDetails(*cr),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.RDSInstance)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotRDSInstance)
	}
	cr.SetConditions(runtimev1alpha1.Creating())
	if cr.Status.AtProvider.DBInstanceStatus == v1beta1.RDSInstanceStateCreating {
		return managed.ExternalCreation{}, nil
	}
	password, err := util.GeneratePassword(20)
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	req := e.client.CreateDBInstanceRequest(rds.GenerateCreateDBInstanceInput(meta.GetExternalName(cr), password, &cr.Spec.ForProvider))
	req.SetContext(ctx)
	_, err = req.Send()
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}
	conn := managed.ConnectionDetails{
		runtimev1alpha1.ResourceCredentialsSecretPasswordKey: []byte(password),
	}
	if cr.Spec.ForProvider.MasterUsername != nil {
		conn[runtimev1alpha1.ResourceCredentialsSecretUserKey] = []byte(aws.StringValue(cr.Spec.ForProvider.MasterUsername))
	}
	return managed.ExternalCreation{ConnectionDetails: conn}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta1.RDSInstance)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotRDSInstance)
	}
	if cr.Status.AtProvider.DBInstanceStatus == v1beta1.RDSInstanceStateModifying {
		return managed.ExternalUpdate{}, nil
	}
	// AWS rejects modification requests if you send fields whose value is same
	// as the current one. So, we have to create a patch out of the desired state
	// and the current state. Since the DBInstance is not fully mirrored in status,
	// we lose the current state after a change is made to spec, which forces us
	// to make a DescribeDBInstancesRequest to get the current state.
	describe := e.client.DescribeDBInstancesRequest(&awsrds.DescribeDBInstancesInput{DBInstanceIdentifier: aws.String(meta.GetExternalName(cr))})
	describe.SetContext(ctx)
	rsp, err := describe.Send()
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errDescribeFailed)
	}
	patch, err := rds.CreatePatch(&rsp.DBInstances[0], &cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errPatchCreationFailed)
	}
	modify := e.client.ModifyDBInstanceRequest(rds.GenerateModifyDBInstanceInput(meta.GetExternalName(cr), patch))
	modify.SetContext(ctx)
	_, err = modify.Send()
	return managed.ExternalUpdate{}, errors.Wrap(err, errModifyFailed)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1beta1.RDSInstance)
	if !ok {
		return errors.New(errNotRDSInstance)
	}
	cr.SetConditions(runtimev1alpha1.Deleting())
	if cr.Status.AtProvider.DBInstanceStatus == v1beta1.RDSInstanceStateDeleting {
		return nil
	}
	// TODO(muvaf): There are cases where deletion results in an error that can
	// be solved only by a config change. But to do that, reconciler has to call
	// Update before Delete, which is not the case currently. In RDS, deletion
	// protection is an example for that and it's pretty common to use it. So,
	// until managed reconciler does Update before Delete, we do it here manually.
	if _, err := e.Update(ctx, cr); err != nil {
		return resource.Ignore(rds.IsErrorNotFound, err)
	}
	input := awsrds.DeleteDBInstanceInput{
		DBInstanceIdentifier: aws.String(meta.GetExternalName(cr)),
		SkipFinalSnapshot:    cr.Spec.ForProvider.SkipFinalSnapshotBeforeDeletion,
	}
	req := e.client.DeleteDBInstanceRequest(&input)
	req.SetContext(ctx)
	_, err := req.Send()
	return errors.Wrap(resource.Ignore(rds.IsErrorNotFound, err), errDeleteFailed)
}
