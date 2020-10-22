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

package iamaccesskey

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/identity/v1alpha1"
	awscommon "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/iam"
)

const (
	errUnexpectedObject = "The managed resource is not an IAMAccessKey resource"
	errList             = "failed to list IAMAccessKeys"
	errCreate           = "failed to create the IAMAccessKey resource"
	errDelete           = "failed to delete the IAMAccessKey resource"
	errUpdate           = "failed to update the IAMAccessKey resource"

	errKubeUpdateFailed = "cannot late initialize IAMAccessKey"
)

// SetupIAMAccessKey adds a controller that reconciles IAMAccessKeys.
func SetupIAMAccessKey(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.IAMAccessKeyGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.IAMAccessKey{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.IAMAccessKeyGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: iam.NewAccessClient}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithConnectionPublishers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) iam.AccessClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cfg, err := awscommon.GetConfig(ctx, c.kube, mg, awscommon.GlobalRegion)
	if err != nil {
		return nil, err
	}
	return &external{client: c.newClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	client iam.AccessClient
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1alpha1.IAMAccessKey)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	keys, err := e.client.ListAccessKeysRequest(&awsiam.ListAccessKeysInput{UserName: aws.String(cr.Spec.ForProvider.IAMUsername)}).Send(ctx)
	if resource.Ignore(iam.IsErrorNotFound, err) != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errList)
	}

	if keys != nil {
		for _, key := range keys.AccessKeyMetadata {
			if aws.StringValue(key.AccessKeyId) == cr.Status.AtProvider.AccessKeyID {
				cr.SetConditions(runtimev1alpha1.Available())

				return managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: string(key.Status) == cr.Status.AtProvider.Status,
				}, nil
			}
		}
	}

	return managed.ExternalObservation{}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1alpha1.IAMAccessKey)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Creating())

	response, err := e.client.CreateAccessKeyRequest(&awsiam.CreateAccessKeyInput{UserName: aws.String(cr.Spec.ForProvider.IAMUsername)}).Send(ctx)

	var conn managed.ConnectionDetails = nil
	if response != nil && response.AccessKey != nil {
		conn = managed.ConnectionDetails{}
		conn[runtimev1alpha1.ResourceCredentialsSecretUserKey] = []byte(aws.StringValue(response.AccessKey.AccessKeyId))
		conn[runtimev1alpha1.ResourceCredentialsSecretPasswordKey] = []byte(aws.StringValue(response.AccessKey.SecretAccessKey))

		cr.Status.AtProvider.Status = string(response.AccessKey.Status)
		cr.Status.AtProvider.AccessKeyID = aws.StringValue(response.AccessKey.AccessKeyId)

		err := e.kube.Update(ctx, cr)
		if err != nil {
			return managed.ExternalCreation{}, errors.Wrap(err, errKubeUpdateFailed)
		}
	}

	return managed.ExternalCreation{ConnectionDetails: conn}, errors.Wrap(err, errCreate)
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*v1alpha1.IAMAccessKey)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	_, err := e.client.UpdateAccessKeyRequest(&awsiam.UpdateAccessKeyInput{
		AccessKeyId: aws.String(cr.Status.AtProvider.AccessKeyID),
		Status:      awsiam.StatusType(cr.Status.AtProvider.Status),
		UserName:    aws.String(cr.Spec.ForProvider.IAMUsername),
	}).Send(ctx)

	return managed.ExternalUpdate{}, errors.Wrap(resource.Ignore(iam.IsErrorNotFound, err), errUpdate)
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1alpha1.IAMAccessKey)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Deleting())

	_, err := e.client.DeleteAccessKeyRequest(&awsiam.DeleteAccessKeyInput{
		UserName:    aws.String(cr.Spec.ForProvider.IAMUsername),
		AccessKeyId: aws.String(cr.Status.AtProvider.AccessKeyID),
	}).Send(ctx)

	return errors.Wrap(resource.Ignore(iam.IsErrorNotFound, err), errDelete)
}
