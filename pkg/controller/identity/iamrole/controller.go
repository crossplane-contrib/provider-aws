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
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/identity/v1beta1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/iam"
)

const (
	errUnexpectedObject = "The managed resource is not an IAMRole resource"
	errGet              = "failed to get IAMRole with name"
	errCreate           = "failed to create the IAMRole resource"
	errDelete           = "failed to delete the IAMRole resource"
	errUpdate           = "failed to update the IAMRole resource"
	errSDK              = "empty IAMRole received from IAM API"
	errCreatePatch      = "failed to create patch object for comparison"

	errKubeUpdateFailed = "cannot late initialize IAMRole"
	errUpToDateFailed   = "cannot check whether object is up-to-date"
)

// SetupIAMRole adds a controller that reconciles IAMRoles.
func SetupIAMRole(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(v1beta1.IAMRoleGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&v1beta1.IAMRole{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1beta1.IAMRoleGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: iam.NewRoleClient}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithConnectionPublishers(),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) iam.RoleClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cfg, err := awsclient.GetConfig(ctx, c.kube, mg, awsclient.GlobalRegion)
	if err != nil {
		return nil, err
	}
	return &external{client: c.newClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	client iam.RoleClient
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1beta1.IAMRole)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	observed, err := e.client.GetRoleRequest(&awsiam.GetRoleInput{
		RoleName: aws.String(meta.GetExternalName(cr)),
	}).Send(ctx)

	if err != nil {
		return managed.ExternalObservation{}, awsclient.Wrap(resource.Ignore(iam.IsErrorNotFound, err), errGet)
	}

	if observed.Role == nil {
		return managed.ExternalObservation{}, errors.New(errSDK)
	}

	role := *observed.Role
	current := cr.Spec.ForProvider.DeepCopy()
	iam.LateInitializeRole(&cr.Spec.ForProvider, &role)
	if !cmp.Equal(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errKubeUpdateFailed)
		}
	}

	cr.SetConditions(xpv1.Available())

	cr.Status.AtProvider = iam.GenerateRoleObservation(*observed.Role)

	upToDate, observationMessage, err := iam.IsRoleUpToDate(cr.Spec.ForProvider, role)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errUpToDateFailed)
	}

	return managed.ExternalObservation{
		ResourceExists:     true,
		ResourceUpToDate:   upToDate,
		ObservationMessage: observationMessage,
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1beta1.IAMRole)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Creating())

	_, err := e.client.CreateRoleRequest(iam.GenerateCreateRoleInput(meta.GetExternalName(cr), &cr.Spec.ForProvider)).Send(ctx)
	return managed.ExternalCreation{}, awsclient.Wrap(err, errCreate)
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) { // nolint:gocyclo
	cr, ok := mgd.(*v1beta1.IAMRole)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	observed, err := e.client.GetRoleRequest(&awsiam.GetRoleInput{
		RoleName: aws.String(meta.GetExternalName(cr)),
	}).Send(ctx)
	if err != nil {
		return managed.ExternalUpdate{}, awsclient.Wrap(resource.Ignore(iam.IsErrorNotFound, err), errGet)
	}
	if observed.Role == nil {
		return managed.ExternalUpdate{}, errors.New(errSDK)
	}

	add, remove := iam.DiffIAMTags(cr.Spec.ForProvider.Tags, observed.Role.Tags)
	if len(remove) != 0 {
		if _, err := e.client.UntagRoleRequest(&awsiam.UntagRoleInput{
			RoleName: aws.String(meta.GetExternalName(cr)),
			TagKeys:  remove,
		}).Send(ctx); err != nil {
			return managed.ExternalUpdate{}, awsclient.Wrap(err, "cannot untag")
		}
	}
	if len(add) != 0 {
		if _, err := e.client.TagRoleRequest(&awsiam.TagRoleInput{
			RoleName: aws.String(meta.GetExternalName(cr)),
			Tags:     add,
		}).Send(ctx); err != nil {
			return managed.ExternalUpdate{}, awsclient.Wrap(err, "cannot tag")
		}
	}

	patch, err := iam.CreatePatch(observed.Role, &cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errCreatePatch)
	}

	if patch.Description != nil || patch.MaxSessionDuration != nil {
		_, err = e.client.UpdateRoleRequest(&awsiam.UpdateRoleInput{
			RoleName:           aws.String(meta.GetExternalName(cr)),
			Description:        cr.Spec.ForProvider.Description,
			MaxSessionDuration: cr.Spec.ForProvider.MaxSessionDuration,
		}).Send(ctx)

		if err != nil {
			return managed.ExternalUpdate{}, awsclient.Wrap(err, errUpdate)
		}
	}

	if patch.AssumeRolePolicyDocument != "" {
		_, err = e.client.UpdateAssumeRolePolicyRequest(&awsiam.UpdateAssumeRolePolicyInput{
			PolicyDocument: &cr.Spec.ForProvider.AssumeRolePolicyDocument,
			RoleName:       aws.String(meta.GetExternalName(cr)),
		}).Send(ctx)
		if err != nil {
			return managed.ExternalUpdate{}, awsclient.Wrap(err, errUpdate)
		}
	}
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1beta1.IAMRole)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	_, err := e.client.DeleteRoleRequest(&awsiam.DeleteRoleInput{
		RoleName: aws.String(meta.GetExternalName(cr)),
	}).Send(ctx)

	return awsclient.Wrap(resource.Ignore(iam.IsErrorNotFound, err), errDelete)
}
