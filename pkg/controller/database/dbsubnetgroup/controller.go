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

	"github.com/aws/aws-sdk-go-v2/aws"
	awsrds "github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	v1alpha3 "github.com/crossplane/provider-aws/apis/database/v1alpha3"
	"github.com/crossplane/provider-aws/pkg/clients/rds"
	"github.com/crossplane/provider-aws/pkg/controller/utils"
)

const (
	errUnexpectedObject = "The managed resource is not an DBSubnetGroup resource"
	errClient           = "cannot create a new DBSubnetGroup client"
	errDescribe         = "failed to describe DBSubnetGroup with groupName: %v"
	errMultipleItems    = "retrieved multiple DBSubnetGroups for the given groupName: %v"
	errCreate           = "failed to create the DBSubnetGroup resource with name: %v"
	errDelete           = "failed to delete the DBSubnetGroup resource"
)

// SetupDBSubnetGroup adds a controller that reconciles DBSubnetGroups.
func SetupDBSubnetGroup(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha3.DBSubnetGroupGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha3.DBSubnetGroup{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha3.DBSubnetGroupGroupVersionKind),
			managed.WithExternalConnecter(&connector{client: mgr.GetClient(), newClientFn: rds.NewDBSubnetGroupClient, awsConfigFn: utils.RetrieveAwsConfigFromProvider}),
			managed.WithConnectionPublishers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	client      client.Client
	newClientFn func(*aws.Config) (rds.DBSubnetGroupClient, error)
	awsConfigFn func(context.Context, client.Reader, *corev1.ObjectReference) (*aws.Config, error)
}

func (conn *connector) Connect(ctx context.Context, mgd resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mgd.(*v1alpha3.DBSubnetGroup)
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
	client rds.DBSubnetGroupClient
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1alpha3.DBSubnetGroup)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	req := e.client.DescribeDBSubnetGroupsRequest(&awsrds.DescribeDBSubnetGroupsInput{
		DBSubnetGroupName: aws.String(cr.Spec.DBSubnetGroupName),
	})

	response, err := req.Send(ctx)

	if err != nil {
		if rds.IsDBSubnetGroupNotFoundErr(err) {
			return managed.ExternalObservation{
				ResourceExists: false,
			}, nil
		}

		return managed.ExternalObservation{}, errors.Wrapf(err, errDescribe, cr.Spec.DBSubnetGroupName)
	}

	// in a successful response, there should be one and only one object
	if len(response.DBSubnetGroups) != 1 {
		return managed.ExternalObservation{}, errors.Errorf(errMultipleItems, cr.Spec.DBSubnetGroupName)
	}

	observed := response.DBSubnetGroups[0]

	cr.SetConditions(runtimev1alpha1.Available())

	cr.UpdateExternalStatus(observed)

	return managed.ExternalObservation{
		ResourceExists:    true,
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1alpha3.DBSubnetGroup)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Creating())

	input := &awsrds.CreateDBSubnetGroupInput{
		DBSubnetGroupDescription: aws.String(cr.Spec.DBSubnetGroupDescription),
		DBSubnetGroupName:        aws.String(cr.Spec.DBSubnetGroupName),
		SubnetIds:                cr.Spec.SubnetIDs,
		Tags:                     []awsrds.Tag{},
	}

	for _, t := range cr.Spec.Tags {
		input.Tags = append(input.Tags, awsrds.Tag{
			Key:   aws.String(t.Key),
			Value: aws.String(t.Value),
		})
	}

	req := e.client.CreateDBSubnetGroupRequest(input)

	response, err := req.Send(ctx)

	if err != nil {
		return managed.ExternalCreation{}, errors.Wrapf(err, errCreate, cr.Spec.DBSubnetGroupName)
	}

	cr.UpdateExternalStatus(*response.DBSubnetGroup)

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	// TODO(soorena776): add more sophisticated Update logic, once we
	// categorize immutable vs mutable fields (see #727)

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1alpha3.DBSubnetGroup)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Deleting())

	req := e.client.DeleteDBSubnetGroupRequest(&awsrds.DeleteDBSubnetGroupInput{
		DBSubnetGroupName: aws.String(cr.Spec.DBSubnetGroupName),
	})

	_, err := req.Send(ctx)

	if rds.IsDBSubnetGroupNotFoundErr(err) {
		return nil
	}

	return errors.Wrap(err, errDelete)
}
