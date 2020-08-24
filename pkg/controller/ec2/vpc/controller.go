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

package vpc

import (
	"context"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/ec2/v1beta1"
	awscommon "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/ec2"
)

const (
	errUnexpectedObject = "The managed resource is not an VPC resource"
	errKubeUpdateFailed = "cannot update VPC custom resource"

	errDescribe            = "failed to describe VPC with id"
	errMultipleItems       = "retrieved multiple VPCs for the given vpcId"
	errCreate              = "failed to create the VPC resource"
	errUpdate              = "failed to update VPC resource"
	errModifyVPCAttributes = "failed to modify the VPC resource attributes"
	errCreateTags          = "failed to create tags for the VPC resource"
	errDelete              = "failed to delete the VPC resource"
	errSpecUpdate          = "cannot update spec of VPC custom resource"
	errStatusUpdate        = "cannot update status of VPC custom resource"
)

// SetupVPC adds a controller that reconciles VPCs.
func SetupVPC(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1beta1.VPCGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1beta1.VPC{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1beta1.VPCGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: ec2.NewVPCClient, awsConfigFn: awscommon.GetConfig}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithConnectionPublishers(),
			managed.WithInitializers(&tagger{kube: mgr.GetClient()}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) ec2.VPCClient
	awsConfigFn func(context.Context, client.Client, resource.Managed, string) (*aws.Config, error)
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cfg, err := c.awsConfigFn(ctx, c.kube, mg, "")
	if err != nil {
		return nil, err
	}
	return &external{client: c.newClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	kube   client.Client
	client ec2.VPCClient
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	cr, ok := mgd.(*v1beta1.VPC)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	response, err := e.client.DescribeVpcsRequest(&awsec2.DescribeVpcsInput{
		VpcIds: []string{meta.GetExternalName(cr)},
	}).Send(ctx)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrapf(resource.Ignore(ec2.IsVPCNotFoundErr, err), errDescribe)
	}

	// in a successful response, there should be one and only one object
	if len(response.Vpcs) != 1 {
		return managed.ExternalObservation{}, errors.New(errMultipleItems)
	}

	observed := response.Vpcs[0]

	// update the CRD spec for any new values from provider
	current := cr.Spec.ForProvider.DeepCopy()
	ec2.LateInitializeVPC(&cr.Spec.ForProvider, &observed)
	if !cmp.Equal(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errSpecUpdate)
		}
	}

	switch observed.State {
	case awsec2.VpcStateAvailable:
		cr.SetConditions(runtimev1alpha1.Available())
	case awsec2.VpcStatePending:
		cr.SetConditions(runtimev1alpha1.Creating())
	}

	cr.Status.AtProvider = ec2.GenerateVpcObservation(observed)

	o := awsec2.DescribeVpcAttributeOutput{}

	for _, input := range []awsec2.VpcAttributeName{
		awsec2.VpcAttributeNameEnableDnsSupport,
		awsec2.VpcAttributeNameEnableDnsHostnames,
	} {
		r, err := e.client.DescribeVpcAttributeRequest(&awsec2.DescribeVpcAttributeInput{
			VpcId:     aws.String(meta.GetExternalName(cr)),
			Attribute: input,
		}).Send(context.Background())

		if err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errDescribe)
		}

		if r.EnableDnsHostnames != nil {
			o.EnableDnsHostnames = r.EnableDnsHostnames
		}

		if r.EnableDnsSupport != nil {
			o.EnableDnsSupport = r.EnableDnsSupport
		}
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: ec2.IsVpcUpToDate(cr.Spec.ForProvider, observed, o),
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1beta1.VPC)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Creating())
	if err := e.kube.Status().Update(ctx, cr); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errStatusUpdate)
	}

	result, err := e.client.CreateVpcRequest(&awsec2.CreateVpcInput{
		CidrBlock:       aws.String(cr.Spec.ForProvider.CIDRBlock),
		InstanceTenancy: awsec2.Tenancy(aws.StringValue(cr.Spec.ForProvider.InstanceTenancy)),
	}).Send(ctx)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreate)
	}

	meta.SetExternalName(cr, aws.StringValue(result.Vpc.VpcId))

	return managed.ExternalCreation{}, errors.Wrap(e.kube.Update(ctx, cr), errSpecUpdate)
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*v1beta1.VPC)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	for _, input := range []*awsec2.ModifyVpcAttributeInput{
		{
			VpcId:            aws.String(meta.GetExternalName(cr)),
			EnableDnsSupport: &awsec2.AttributeBooleanValue{Value: cr.Spec.ForProvider.EnableDNSSupport},
		},
		{
			VpcId:              aws.String(meta.GetExternalName(cr)),
			EnableDnsHostnames: &awsec2.AttributeBooleanValue{Value: cr.Spec.ForProvider.EnableDNSHostNames},
		},
	} {
		if _, err := e.client.ModifyVpcAttributeRequest(input).Send(ctx); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errModifyVPCAttributes)
		}
	}

	// NOTE(muvaf): VPCs can only be tagged after the creation and this request
	// is idempotent.
	if _, err := e.client.CreateTagsRequest(&awsec2.CreateTagsInput{
		Resources: []string{meta.GetExternalName(cr)},
		Tags:      v1beta1.GenerateEC2Tags(cr.Spec.ForProvider.Tags),
	}).Send(ctx); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errCreateTags)
	}

	_, err := e.client.ModifyVpcTenancyRequest(&awsec2.ModifyVpcTenancyInput{
		InstanceTenancy: awsec2.VpcTenancy(aws.StringValue(cr.Spec.ForProvider.InstanceTenancy)),
		VpcId:           aws.String(meta.GetExternalName(cr)),
	}).Send(ctx)

	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdate)
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1beta1.VPC)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Deleting())

	_, err := e.client.DeleteVpcRequest(&awsec2.DeleteVpcInput{
		VpcId: aws.String(meta.GetExternalName(cr)),
	}).Send(ctx)

	return errors.Wrap(resource.Ignore(ec2.IsVPCNotFoundErr, err), errDelete)
}

type tagger struct {
	kube client.Client
}

func (t *tagger) Initialize(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1beta1.VPC)
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	tagMap := map[string]string{}
	for _, t := range cr.Spec.ForProvider.Tags {
		tagMap[t.Key] = t.Value
	}
	for k, v := range resource.GetExternalTags(mgd) {
		tagMap[k] = v
	}
	cr.Spec.ForProvider.Tags = make([]v1beta1.Tag, len(tagMap))
	i := 0
	for k, v := range tagMap {
		cr.Spec.ForProvider.Tags[i] = v1beta1.Tag{Key: k, Value: v}
		i++
	}
	sort.Slice(cr.Spec.ForProvider.Tags, func(i, j int) bool {
		return cr.Spec.ForProvider.Tags[i].Key < cr.Spec.ForProvider.Tags[j].Key
	})
	return errors.Wrap(t.kube.Update(ctx, cr), errKubeUpdateFailed)
}
