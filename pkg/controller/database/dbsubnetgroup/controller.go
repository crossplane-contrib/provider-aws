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
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsrds "github.com/aws/aws-sdk-go-v2/service/rds"
	awsrdstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/database/v1beta1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	dbsg "github.com/crossplane-contrib/provider-aws/pkg/clients/dbsubnetgroup"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errLateInit         = "cannot late initialize DBSubnetGroup"
	errUnexpectedObject = "the managed resource is not an DBSubnetGroup"
	errDescribe         = "cannot describe DBSubnetGroup"
	errCreate           = "cannot create the DBSubnetGroup"
	errDelete           = "cannot delete the DBSubnetGroup"
	errUpdate           = "cannot update the DBSubnetGroup"
	errAddTagsFailed    = "cannot add tags to DBSubnetGroup"
	errListTagsFailed   = "cannot list tags for DBSubnetGroup"
	errNotOne           = "expected exactly one DBSubnetGroup"
)

// SetupDBSubnetGroup adds a controller that reconciles DBSubnetGroups.
func SetupDBSubnetGroup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.DBSubnetGroupGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: dbsg.NewClient}),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithConnectionPublishers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.DBSubnetGroupGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1beta1.DBSubnetGroup{}).
		Complete(r)
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) dbsg.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1beta1.DBSubnetGroup)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	cfg, err := connectaws.GetConfig(ctx, c.kube, mg, aws.ToString(cr.Spec.ForProvider.Region))
	if err != nil {
		return nil, err
	}
	return &external{c.newClientFn(*cfg), c.kube}, nil
}

type external struct {
	client dbsg.Client
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1beta1.DBSubnetGroup)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	res, err := e.client.DescribeDBSubnetGroups(ctx, &awsrds.DescribeDBSubnetGroupsInput{
		DBSubnetGroupName: aws.String(meta.GetExternalName(cr)),
	})
	if err != nil {
		return managed.ExternalObservation{}, errorutils.Wrap(resource.Ignore(dbsg.IsDBSubnetGroupNotFoundErr, err), errDescribe)
	}

	// in a successful response, there should be one and only one object
	if len(res.DBSubnetGroups) != 1 {
		return managed.ExternalObservation{}, errors.New(errNotOne)
	}

	observed := res.DBSubnetGroups[0]
	current := cr.Spec.ForProvider.DeepCopy()
	dbsg.LateInitialize(&cr.Spec.ForProvider, &observed)
	if !reflect.DeepEqual(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errLateInit)
		}
	}
	cr.Status.AtProvider = dbsg.GenerateObservation(observed)

	if strings.EqualFold(cr.Status.AtProvider.State, v1beta1.DBSubnetGroupStateAvailable) {
		cr.Status.SetConditions(xpv1.Available())
	} else {
		cr.Status.SetConditions(xpv1.Unavailable())
	}

	tags, err := e.client.ListTagsForResource(ctx, &awsrds.ListTagsForResourceInput{
		ResourceName: aws.String(cr.Status.AtProvider.ARN),
	})
	if err != nil {
		return managed.ExternalObservation{}, errorutils.Wrap(err, errListTagsFailed)
	}

	return managed.ExternalObservation{
		ResourceUpToDate: dbsg.IsDBSubnetGroupUpToDate(cr.Spec.ForProvider, observed, tags.TagList),
		ResourceExists:   true,
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1beta1.DBSubnetGroup)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	cr.SetConditions(xpv1.Creating())
	input := &awsrds.CreateDBSubnetGroupInput{
		DBSubnetGroupDescription: aws.String(cr.Spec.ForProvider.Description),
		DBSubnetGroupName:        aws.String(meta.GetExternalName(cr)),
		SubnetIds:                cr.Spec.ForProvider.SubnetIDs,
	}

	if len(cr.Spec.ForProvider.Tags) != 0 {
		input.Tags = make([]awsrdstypes.Tag, len(cr.Spec.ForProvider.Tags))
		for i, val := range cr.Spec.ForProvider.Tags {
			input.Tags[i] = awsrdstypes.Tag{Key: aws.String(val.Key), Value: aws.String(val.Value)}
		}
	}

	_, err := e.client.CreateDBSubnetGroup(ctx, input)
	return managed.ExternalCreation{}, errorutils.Wrap(err, errCreate)
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*v1beta1.DBSubnetGroup)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	_, err := e.client.ModifyDBSubnetGroup(ctx, &awsrds.ModifyDBSubnetGroupInput{
		DBSubnetGroupName:        aws.String(meta.GetExternalName(cr)),
		DBSubnetGroupDescription: aws.String(cr.Spec.ForProvider.Description),
		SubnetIds:                cr.Spec.ForProvider.SubnetIDs,
	})

	if err != nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
	}

	if len(cr.Spec.ForProvider.Tags) > 0 {
		tags := make([]awsrdstypes.Tag, len(cr.Spec.ForProvider.Tags))
		for i, t := range cr.Spec.ForProvider.Tags {
			tags[i] = awsrdstypes.Tag{Key: aws.String(t.Key), Value: aws.String(t.Value)}
		}
		_, err = e.client.AddTagsToResource(ctx, &awsrds.AddTagsToResourceInput{
			ResourceName: aws.String(cr.Status.AtProvider.ARN),
			Tags:         tags,
		})
		if err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errAddTagsFailed)
		}
	}
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1beta1.DBSubnetGroup)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.SetConditions(xpv1.Deleting())
	_, err := e.client.DeleteDBSubnetGroup(ctx, &awsrds.DeleteDBSubnetGroupInput{
		DBSubnetGroupName: aws.String(meta.GetExternalName(cr)),
	})
	return errorutils.Wrap(resource.Ignore(dbsg.IsDBSubnetGroupNotFoundErr, err), errDelete)
}
