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

package iamrole

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/event"
	"github.com/crossplaneio/crossplane-runtime/pkg/logging"
	"github.com/crossplaneio/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"

	v1alpha3 "github.com/crossplaneio/stack-aws/apis/identity/v1alpha3"
	"github.com/crossplaneio/stack-aws/pkg/clients/iam"
	"github.com/crossplaneio/stack-aws/pkg/controller/utils"
)

const (
	errUnexpectedObject = "The managed resource is not an IAMRole resource"
	errClient           = "cannot create a new IAMRole client"
	errGet              = "failed to get IAMRole with name: %v"
	errCreate           = "failed to create the IAMRole resource"
	errDelete           = "failed to delete the IAMRole resource"
)

// SetupIAMRole adds a controller that reconciles IAMRoles.
func SetupIAMRole(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha3.IAMRoleGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha3.IAMRole{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha3.IAMRoleGroupVersionKind),
			managed.WithExternalConnecter(&connector{client: mgr.GetClient(), newClientFn: iam.NewRoleClient, awsConfigFn: utils.RetrieveAwsConfigFromProvider}),
			managed.WithConnectionPublishers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	client      client.Client
	newClientFn func(*aws.Config) (iam.RoleClient, error)
	awsConfigFn func(context.Context, client.Reader, *corev1.ObjectReference) (*aws.Config, error)
}

func (conn *connector) Connect(ctx context.Context, mgd resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mgd.(*v1alpha3.IAMRole)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}

	awsconfig, err := conn.awsConfigFn(ctx, conn.client, cr.Spec.ProviderReference)
	if err != nil {
		return nil, err
	}

	c, err := conn.newClientFn(awsconfig)
	if err != nil {
		return nil, errors.Wrap(err, errClient)
	}
	return &external{c}, nil
}

type external struct {
	client iam.RoleClient
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1alpha3.IAMRole)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	req := e.client.GetRoleRequest(&awsiam.GetRoleInput{
		RoleName: aws.String(cr.Spec.RoleName),
	})

	observed, err := req.Send(ctx)

	if iam.IsErrorNotFound(err) {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	if err != nil {
		return managed.ExternalObservation{}, errors.Wrapf(err, errGet, cr.Spec.RoleName)
	}

	cr.SetConditions(runtimev1alpha1.Available())

	cr.UpdateExternalStatus(*observed.Role)

	return managed.ExternalObservation{
		ResourceExists: true,
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1alpha3.IAMRole)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Creating())

	req := e.client.CreateRoleRequest(&awsiam.CreateRoleInput{
		RoleName:                 aws.String(cr.Spec.RoleName),
		AssumeRolePolicyDocument: aws.String(cr.Spec.AssumeRolePolicyDocument),
		Description:              aws.String(cr.Spec.Description),
	})

	result, err := req.Send(ctx)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreate)
	}

	cr.UpdateExternalStatus(*result.Role)

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	// TODO(soorena776): add more sophisticated Update logic, once we
	// categorize immutable vs mutable fields (see #727)

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1alpha3.IAMRole)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Deleting())

	req := e.client.DeleteRoleRequest(&awsiam.DeleteRoleInput{
		RoleName: aws.String(cr.Spec.RoleName),
	})

	_, err := req.Send(ctx)

	if iam.IsErrorNotFound(err) {
		return nil
	}

	return errors.Wrap(err, errDelete)
}
