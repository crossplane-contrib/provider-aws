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

package iamgroup

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
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

	"github.com/crossplane/provider-aws/apis/identity/v1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/iam"
)

const (
	errUnexpectedObject = "The managed resource is not an IAM Group resource"
	errGet              = "failed to get IAM Group with name"
	errCreate           = "failed to create the IAM Group resource"
	errDelete           = "failed to delete the IAM Group resource"
	errUpdate           = "failed to update the IAM Group resource"
	errSDK              = "empty IAM Group received from IAM API"

	errKubeUpdateFailed = "cannot late initialize IAM Group"
)

// SetupIAMGroup adds a controller that reconciles Groups.
func SetupIAMGroup(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(v1alpha1.IAMGroupGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewController(rl),
		}).
		For(&v1alpha1.IAMGroup{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.IAMGroupGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: iam.NewGroupClient}),
			managed.WithConnectionPublishers(),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) iam.GroupClient
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
	client iam.GroupClient
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1alpha1.IAMGroup)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	observed, err := e.client.GetGroup(ctx, &awsiam.GetGroupInput{
		GroupName: aws.String(meta.GetExternalName(cr)),
	})

	if err != nil {
		return managed.ExternalObservation{}, awsclient.Wrap(resource.Ignore(iam.IsErrorNotFound, err), errGet)
	}

	if observed.Group == nil {
		return managed.ExternalObservation{}, errors.New(errSDK)
	}

	group := *observed.Group

	current := cr.Spec.ForProvider.DeepCopy()
	cr.Spec.ForProvider.Path = awsclient.LateInitializeStringPtr(cr.Spec.ForProvider.Path, group.Path)

	if aws.ToString(current.Path) != aws.ToString(cr.Spec.ForProvider.Path) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errKubeUpdateFailed)
		}
	}

	cr.SetConditions(xpv1.Available())

	cr.Status.AtProvider = v1alpha1.IAMGroupObservation{
		ARN:     aws.ToString(group.Arn),
		GroupID: aws.ToString(group.GroupId),
	}

	return managed.ExternalObservation{
		ResourceExists: true,
		ResourceUpToDate: aws.ToString(cr.Spec.ForProvider.Path) == aws.ToString(group.Path) &&
			meta.GetExternalName(cr) == aws.ToString(group.GroupName),
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1alpha1.IAMGroup)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}
	cr.Status.SetConditions(xpv1.Creating())

	_, err := e.client.CreateGroup(ctx, &awsiam.CreateGroupInput{
		GroupName: aws.String(meta.GetExternalName(cr)),
		Path:      cr.Spec.ForProvider.Path,
	})
	return managed.ExternalCreation{}, awsclient.Wrap(err, errCreate)
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*v1alpha1.IAMGroup)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	_, err := e.client.UpdateGroup(ctx, &awsiam.UpdateGroupInput{
		NewPath:      cr.Spec.ForProvider.Path,
		NewGroupName: aws.String(meta.GetExternalName(cr)),
	})

	return managed.ExternalUpdate{}, awsclient.Wrap(err, errUpdate)
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1alpha1.IAMGroup)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	_, err := e.client.DeleteGroup(ctx, &awsiam.DeleteGroupInput{
		GroupName: aws.String(meta.GetExternalName(cr)),
	})

	return awsclient.Wrap(resource.Ignore(iam.IsErrorNotFound, err), errDelete)
}
