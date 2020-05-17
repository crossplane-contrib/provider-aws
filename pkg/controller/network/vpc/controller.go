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
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/network/v1alpha3"
	"github.com/crossplane/provider-aws/pkg/clients/ec2"
	"github.com/crossplane/provider-aws/pkg/controller/utils"
)

const (
	errUnexpectedObject    = "The managed resource is not an VPC resource"
	errKubeUpdateFailed    = "cannot update VPC custom resource"
	errClient              = "cannot create a new VPCClient"
	errDescribe            = "failed to describe VPC"
	errMultipleItems       = "retrieved multiple VPCs for the given vpcId"
	errCreate              = "failed to create the VPC resource"
	errPersistExternalName = "failed to persist InternetGateway ID"
	errModifyVPCAttributes = "failed to modify the VPC resource attributes"
	errCreateTags          = "failed to create tags for the VPC resource"
	errDelete              = "failed to delete the VPC resource"
)

// SetupVPC adds a controller that reconciles VPCs.
func SetupVPC(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha3.VPCGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha3.VPC{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha3.VPCGroupVersionKind),
			managed.WithExternalConnecter(&connector{client: mgr.GetClient(), newClientFn: ec2.NewVPCClient, awsConfigFn: utils.RetrieveAwsConfigFromProvider}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithConnectionPublishers(),
			managed.WithInitializers(&tagger{kube: mgr.GetClient()}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	client      client.Client
	newClientFn func(*aws.Config) (ec2.VPCClient, error)
	awsConfigFn func(context.Context, client.Reader, *corev1.ObjectReference) (*aws.Config, error)
}

func (conn *connector) Connect(ctx context.Context, mgd resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mgd.(*v1alpha3.VPC)
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
	return &external{kube: conn.client, client: c}, nil
}

type external struct {
	kube   client.Client
	client ec2.VPCClient
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*v1alpha3.VPC)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	// AWS network resources are uniquely identified by an ID that is returned
	// on create time; we can't tell whether they exist unless we have recorded
	// their ID.
	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	req := e.client.DescribeVpcsRequest(&awsec2.DescribeVpcsInput{
		VpcIds: []string{meta.GetExternalName(cr)},
	})
	rsp, err := req.Send(ctx)
	if ec2.IsVPCNotFoundErr(err) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errDescribe)
	}

	// in a successful response, there should be one and only one object
	if len(rsp.Vpcs) != 1 {
		return managed.ExternalObservation{}, errors.New(errMultipleItems)
	}

	observed := rsp.Vpcs[0]

	if observed.State == awsec2.VpcStateAvailable {
		cr.SetConditions(runtimev1alpha1.Available())
	}

	cr.UpdateExternalStatus(observed)

	return managed.ExternalObservation{
		ResourceExists:    true,
		ConnectionDetails: managed.ConnectionDetails{},
		ResourceUpToDate:  ec2.IsUpToDate(cr.Spec.VPCParameters, observed),
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1alpha3.VPC)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	// if VPCID already exists, skip creating the vpc
	// this happens when an error has occurred when modifying vpc attributes
	if meta.GetExternalName(cr) == "" {

		req := e.client.CreateVpcRequest(&awsec2.CreateVpcInput{
			CidrBlock: aws.String(cr.Spec.CIDRBlock),
		})

		rsp, err := req.Send(ctx)
		if err != nil {
			return managed.ExternalCreation{}, errors.Wrap(err, errCreate)
		}

		meta.SetExternalName(cr, aws.StringValue(rsp.Vpc.VpcId))
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalCreation{}, errors.Wrap(err, errPersistExternalName)
		}

		cr.SetConditions(runtimev1alpha1.Creating())
		cr.UpdateExternalStatus(*rsp.Vpc)
	}

	// modify vpc attributes
	for _, input := range []*awsec2.ModifyVpcAttributeInput{
		{
			VpcId:            aws.String(meta.GetExternalName(cr)),
			EnableDnsSupport: &awsec2.AttributeBooleanValue{Value: aws.Bool(cr.Spec.EnableDNSSupport)},
		},
		{
			VpcId:              aws.String(meta.GetExternalName(cr)),
			EnableDnsHostnames: &awsec2.AttributeBooleanValue{Value: aws.Bool(cr.Spec.EnableDNSHostNames)},
		},
	} {
		attrReq := e.client.ModifyVpcAttributeRequest(input)

		if _, err := attrReq.Send(ctx); err != nil {
			return managed.ExternalCreation{}, errors.Wrap(err, errModifyVPCAttributes)
		}
	}

	return managed.ExternalCreation{ConnectionDetails: managed.ConnectionDetails{}}, nil
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*v1alpha3.VPC)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	// TODO(soorena776): add more sophisticated Update logic, once we
	// categorize immutable vs mutable fields (see #727)

	// NOTE(muvaf): VPCs can only be tagged after the creation and this request
	// is idempotent.
	if _, err := e.client.CreateTagsRequest(&awsec2.CreateTagsInput{
		Resources: []string{meta.GetExternalName(cr)},
		Tags:      v1alpha3.GenerateEC2Tags(cr.Spec.Tags),
	}).Send(ctx); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errCreateTags)
	}
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1alpha3.VPC)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Deleting())

	req := e.client.DeleteVpcRequest(&awsec2.DeleteVpcInput{
		VpcId: aws.String(meta.GetExternalName(cr)),
	})

	_, err := req.Send(ctx)
	return errors.Wrap(resource.Ignore(ec2.IsVPCNotFoundErr, err), errDelete)
}

type tagger struct {
	kube client.Client
}

func (t *tagger) Initialize(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1alpha3.VPC)
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	tagMap := map[string]string{}
	for _, t := range cr.Spec.Tags {
		tagMap[t.Key] = t.Value
	}
	for k, v := range resource.GetExternalTags(mgd) {
		tagMap[k] = v
	}
	cr.Spec.Tags = make([]v1alpha3.Tag, len(tagMap))
	i := 0
	for k, v := range tagMap {
		cr.Spec.Tags[i] = v1alpha3.Tag{Key: k, Value: v}
		i++
	}
	sort.Slice(cr.Spec.Tags, func(i, j int) bool {
		return cr.Spec.Tags[i].Key < cr.Spec.Tags[j].Key
	})
	return errors.Wrap(t.kube.Update(ctx, cr), errKubeUpdateFailed)
}
