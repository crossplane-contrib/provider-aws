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

package user

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

	"github.com/crossplane/provider-aws/apis/iam/v1beta1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/iam"
)

const (
	errUnexpectedObject = "The managed resource is not an IAM User resource"

	errGet    = "cannot get IAM User"
	errCreate = "cannot create the IAM User resource"
	errDelete = "cannot delete the IAM User resource"
	errUpdate = "cannot update the IAM User resource"
	errSDK    = "empty IAM User received from IAM API"

	errKubeUpdateFailed = "cannot late initialize IAM User"
)

// SetupUser adds a controller that reconciles Users.
func SetupUser(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(v1beta1.UserGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewController(rl),
		}).
		For(&v1beta1.User{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1beta1.UserGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: iam.NewUserClient}),
			managed.WithConnectionPublishers(),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) iam.UserClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cfg, err := awsclient.GetConfig(ctx, c.kube, mg, awsclient.GlobalRegion)
	if err != nil {
		return nil, err
	}
	return &external{client: c.newClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	kube   client.Client
	client iam.UserClient
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1beta1.User)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	observed, err := e.client.GetUser(ctx, &awsiam.GetUserInput{
		UserName: aws.String(meta.GetExternalName(cr)),
	})

	if err != nil {
		return managed.ExternalObservation{}, awsclient.Wrap(resource.Ignore(iam.IsErrorNotFound, err), errGet)
	}

	if observed.User == nil {
		return managed.ExternalObservation{}, errors.New(errSDK)
	}

	user := *observed.User
	current := cr.Spec.ForProvider.DeepCopy()
	iam.LateInitializeUser(&cr.Spec.ForProvider, &user)
	if !cmp.Equal(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errKubeUpdateFailed)
		}
	}

	cr.SetConditions(xpv1.Available())

	cr.Status.AtProvider = v1beta1.UserObservation{
		ARN:    aws.ToString(user.Arn),
		UserID: aws.ToString(user.UserId),
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: aws.ToString(cr.Spec.ForProvider.Path) == aws.ToString(user.Path),
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1beta1.User)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Creating())

	_, err := e.client.CreateUser(ctx, &awsiam.CreateUserInput{
		Path:                cr.Spec.ForProvider.Path,
		PermissionsBoundary: cr.Spec.ForProvider.PermissionsBoundary,
		Tags:                iam.BuildIAMTags(cr.Spec.ForProvider.Tags),
		UserName:            aws.String(meta.GetExternalName(cr)),
	})
	return managed.ExternalCreation{}, awsclient.Wrap(err, errCreate)
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*v1beta1.User)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	_, err := e.client.UpdateUser(ctx, &awsiam.UpdateUserInput{
		NewPath:  cr.Spec.ForProvider.Path,
		UserName: aws.String(meta.GetExternalName(cr)),
	})

	return managed.ExternalUpdate{}, awsclient.Wrap(err, errUpdate)
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1beta1.User)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	_, err := e.client.DeleteUser(ctx, &awsiam.DeleteUserInput{
		UserName: aws.String(meta.GetExternalName(cr)),
	})

	return awsclient.Wrap(resource.Ignore(iam.IsErrorNotFound, err), errDelete)
}
