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

package elb

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/crossplane-contrib/provider-aws/apis/elasticloadbalancing/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/jsonpatch"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

// A Client handles CRUD operations for Elastic Load Balancing resources.
type Client interface {
	DescribeLoadBalancers(ctx context.Context, input *elb.DescribeLoadBalancersInput, opts ...func(*elb.Options)) (*elb.DescribeLoadBalancersOutput, error)
	CreateLoadBalancer(ctx context.Context, input *elb.CreateLoadBalancerInput, opts ...func(*elb.Options)) (*elb.CreateLoadBalancerOutput, error)
	DeleteLoadBalancer(ctx context.Context, input *elb.DeleteLoadBalancerInput, opts ...func(*elb.Options)) (*elb.DeleteLoadBalancerOutput, error)
	EnableAvailabilityZonesForLoadBalancer(ctx context.Context, input *elb.EnableAvailabilityZonesForLoadBalancerInput, opts ...func(*elb.Options)) (*elb.EnableAvailabilityZonesForLoadBalancerOutput, error)
	DisableAvailabilityZonesForLoadBalancer(ctx context.Context, input *elb.DisableAvailabilityZonesForLoadBalancerInput, opts ...func(*elb.Options)) (*elb.DisableAvailabilityZonesForLoadBalancerOutput, error)
	DetachLoadBalancerFromSubnets(ctx context.Context, input *elb.DetachLoadBalancerFromSubnetsInput, opts ...func(*elb.Options)) (*elb.DetachLoadBalancerFromSubnetsOutput, error)
	AttachLoadBalancerToSubnets(ctx context.Context, input *elb.AttachLoadBalancerToSubnetsInput, opts ...func(*elb.Options)) (*elb.AttachLoadBalancerToSubnetsOutput, error)
	ApplySecurityGroupsToLoadBalancer(ctx context.Context, input *elb.ApplySecurityGroupsToLoadBalancerInput, opts ...func(*elb.Options)) (*elb.ApplySecurityGroupsToLoadBalancerOutput, error)
	CreateLoadBalancerListeners(ctx context.Context, input *elb.CreateLoadBalancerListenersInput, opts ...func(*elb.Options)) (*elb.CreateLoadBalancerListenersOutput, error)
	DeleteLoadBalancerListeners(ctx context.Context, input *elb.DeleteLoadBalancerListenersInput, opts ...func(*elb.Options)) (*elb.DeleteLoadBalancerListenersOutput, error)
	RegisterInstancesWithLoadBalancer(ctx context.Context, input *elb.RegisterInstancesWithLoadBalancerInput, opts ...func(*elb.Options)) (*elb.RegisterInstancesWithLoadBalancerOutput, error)
	DeregisterInstancesFromLoadBalancer(ctx context.Context, input *elb.DeregisterInstancesFromLoadBalancerInput, opts ...func(*elb.Options)) (*elb.DeregisterInstancesFromLoadBalancerOutput, error)
	DescribeTags(ctx context.Context, input *elb.DescribeTagsInput, opts ...func(*elb.Options)) (*elb.DescribeTagsOutput, error)
	AddTags(ctx context.Context, input *elb.AddTagsInput, opts ...func(*elb.Options)) (*elb.AddTagsOutput, error)
	RemoveTags(ctx context.Context, input *elb.RemoveTagsInput, opts ...func(*elb.Options)) (*elb.RemoveTagsOutput, error)
	ConfigureHealthCheck(ctx context.Context, params *elb.ConfigureHealthCheckInput, opts ...func(*elb.Options)) (*elb.ConfigureHealthCheckOutput, error)
}

// NewClient returns a new Elastic Load Balancer client. Credentials must be passed as
// JSON encoded data.
func NewClient(cfg aws.Config) Client {
	return elb.NewFromConfig(cfg)
}

// GenerateCreateELBInput generate instance of elasticLoadBlancing.CreateLoadBalancerInput
func GenerateCreateELBInput(name string, p v1alpha1.ELBParameters) *elb.CreateLoadBalancerInput {
	input := elb.CreateLoadBalancerInput{
		AvailabilityZones: p.AvailabilityZones,
		LoadBalancerName:  aws.String(name),
		Scheme:            p.Scheme,
		Subnets:           p.SubnetIDs,
		SecurityGroups:    p.SecurityGroupIDs,
	}
	input.Listeners = BuildELBListeners(p.Listeners)

	return &input
}

// LateInitializeELB fills the empty fields in *v1alpha1.ELBParameters with
// the values seen in elasticLoadBalancing.ELB.
func LateInitializeELB(in *v1alpha1.ELBParameters, v *elbtypes.LoadBalancerDescription, elbTags []elbtypes.Tag) { //nolint:gocyclo
	if v == nil {
		return
	}

	in.Scheme = pointer.LateInitialize(in.Scheme, v.Scheme)

	if len(in.AvailabilityZones) == 0 && len(v.AvailabilityZones) != 0 {
		in.AvailabilityZones = v.AvailabilityZones
	}

	if len(in.SecurityGroupIDs) == 0 && len(v.SecurityGroups) != 0 {
		in.SecurityGroupIDs = v.SecurityGroups
	}

	if len(in.SubnetIDs) == 0 && len(v.Subnets) != 0 {
		in.SubnetIDs = v.Subnets
	}

	if len(in.Listeners) == 0 && len(v.ListenerDescriptions) != 0 {
		in.Listeners = make([]v1alpha1.Listener, len(v.ListenerDescriptions))
		for k, l := range v.ListenerDescriptions {
			in.Listeners[k] = v1alpha1.Listener{
				InstancePort:     l.Listener.InstancePort,
				InstanceProtocol: l.Listener.InstanceProtocol,
				LoadBalancerPort: l.Listener.LoadBalancerPort,
				Protocol:         aws.ToString(l.Listener.Protocol),
				SSLCertificateID: l.Listener.SSLCertificateId,
			}
		}
	}

	if len(in.Tags) == 0 && len(elbTags) != 0 {
		in.Tags = make([]v1alpha1.Tag, len(elbTags))
		for k, t := range elbTags {
			in.Tags[k] = v1alpha1.Tag{
				Key:   aws.ToString(t.Key),
				Value: t.Value,
			}
		}
	}
}

// IsELBNotFound returns true if the error is because the item doesn't exist.
func IsELBNotFound(err error) bool {
	var apnf *elbtypes.AccessPointNotFoundException
	return errors.As(err, &apnf)
}

// GenerateELBObservation is used to produce v1alpha1.ELBObservation from
// elasticLoadBalancing.LoadBalancerDescription.
func GenerateELBObservation(e elbtypes.LoadBalancerDescription) v1alpha1.ELBObservation {
	o := v1alpha1.ELBObservation{
		CanonicalHostedZoneName:   aws.ToString(e.CanonicalHostedZoneName),
		CanonicalHostedZoneNameID: aws.ToString(e.CanonicalHostedZoneNameID),
		DNSName:                   aws.ToString(e.DNSName),
		VPCID:                     aws.ToString(e.VPCId),
	}

	if len(e.BackendServerDescriptions) > 0 {
		descriptions := []v1alpha1.BackendServerDescription{}
		for _, v := range e.BackendServerDescriptions {
			descriptions = append(descriptions, v1alpha1.BackendServerDescription{
				InstancePort: v.InstancePort,
				PolicyNames:  v.PolicyNames,
			})
		}
		o.BackendServerDescriptions = descriptions
	}

	return o
}

// CreatePatch creates a v1alpha1.ELBParameters that has only the changed
// values between the target v1alpha1.ELBParameters and the current
// elbtypes.LoadBalancerDescription.
func CreatePatch(in elbtypes.LoadBalancerDescription, target v1alpha1.ELBParameters, elbTags []elbtypes.Tag) (*v1alpha1.ELBParameters, error) {
	// v1alpha1.ELBParameters contains multiple list types. Sorting these list types is required before
	// creating a patch as jsonpatch.CreateMergePatch considers the order of items in a list.

	currentParams := &v1alpha1.ELBParameters{}
	LateInitializeELB(currentParams, &in, elbTags)
	sortParametersArrays(currentParams)

	targetCopy := target.DeepCopy()
	sortParametersArrays(targetCopy)

	// For listener.Protocol and listener.InstanceProtocol, values in lower and upper case
	// are allowed. But the AWS API always returns the upper case strings.
	for i, v := range targetCopy.Listeners {
		targetCopy.Listeners[i].Protocol = strings.ToUpper(v.Protocol)
		targetCopy.Listeners[i].InstanceProtocol = aws.String(strings.ToUpper(aws.ToString(v.InstanceProtocol)))
	}

	// Make sure the listeners are sorted by port number for both currentParams and target.
	sort.Slice(currentParams.Listeners, func(i, j int) bool {
		return currentParams.Listeners[i].LoadBalancerPort < currentParams.Listeners[j].LoadBalancerPort
	})

	sort.Slice(targetCopy.Listeners, func(i, j int) bool {
		return targetCopy.Listeners[i].LoadBalancerPort < targetCopy.Listeners[j].LoadBalancerPort
	})

	jsonPatch, err := jsonpatch.CreateJSONPatch(currentParams, targetCopy)
	if err != nil {
		return nil, err
	}
	patch := &v1alpha1.ELBParameters{}
	if err := json.Unmarshal(jsonPatch, patch); err != nil {
		return nil, err
	}
	return patch, nil
}

// IsUpToDate checks whether there is a change in any of the modifiable fields.
func IsUpToDate(p v1alpha1.ELBParameters, elb elbtypes.LoadBalancerDescription, elbTags []elbtypes.Tag) (bool, error) {
	patch, err := CreatePatch(elb, p, elbTags)
	if err != nil {
		return false, err
	}
	return cmp.Equal(&v1alpha1.ELBParameters{}, patch,
		cmpopts.IgnoreTypes([]xpv1.Reference{}, []xpv1.Selector{}),
		cmpopts.IgnoreFields(v1alpha1.ELBParameters{}, "Region")), nil
}

// BuildELBListeners builds a list of elbtypes.Listener from given list of v1alpha1.Listener.
func BuildELBListeners(l []v1alpha1.Listener) []elbtypes.Listener {
	out := make([]elbtypes.Listener, len(l))
	for i := range l {
		out[i] = elbtypes.Listener{
			InstancePort:     aws.ToInt32(&l[i].InstancePort),
			InstanceProtocol: l[i].InstanceProtocol,
			LoadBalancerPort: aws.ToInt32(&l[i].LoadBalancerPort),
			Protocol:         &l[i].Protocol,
			SSLCertificateId: l[i].SSLCertificateID,
		}
	}
	return out
}

// BuildELBTags generates a list of elbtypes.Tag from given list of v1alpha1.Tag
func BuildELBTags(tags []v1alpha1.Tag) []elbtypes.Tag {
	if len(tags) == 0 {
		return nil
	}

	elbTags := make([]elbtypes.Tag, len(tags))
	for k, t := range tags {
		elbTags[k] = elbtypes.Tag{
			Key:   aws.String(t.Key),
			Value: t.Value,
		}
	}
	return elbTags
}

func sortParametersArrays(p *v1alpha1.ELBParameters) {
	sort.Strings(p.AvailabilityZones)
	sort.Strings(p.SecurityGroupIDs)
	sort.Strings(p.SubnetIDs)

	sort.Slice(p.Tags, func(i, j int) bool {
		return p.Tags[i].Key < p.Tags[j].Key
	})
}
