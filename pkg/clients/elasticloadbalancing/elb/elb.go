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

	"github.com/aws/aws-sdk-go-v2/aws"
	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/elasticloadbalancingiface"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/elasticloadbalancing/v1alpha1"
	clients "github.com/crossplane/provider-aws/pkg/clients"
)

const (
	// ELBNotFound is the code returned by AWS when there is not ELB with a specified name.
	ELBNotFound = "LoadBalancerNotFound"
)

// A Client handles CRUD operations for Elastic Load Balancing resources.
type Client elasticloadbalancingiface.ClientAPI

// NewClient returns a new Elastic Load Balancer client. Credentials must be passed as
// JSON encoded data.
func NewClient(ctx context.Context, credentials []byte, region string, auth clients.AuthMethod) (Client, error) {
	cfg, err := auth(ctx, credentials, clients.DefaultSection, region)
	if cfg == nil {
		return nil, err
	}
	return elb.New(*cfg), err
}

// GenerateCreateELBInput generate instance of elasticLoadBlancing.CreateLoadBalancerInput
func GenerateCreateELBInput(name string, p v1alpha1.ELBParameters) *elb.CreateLoadBalancerInput {
	input := elb.CreateLoadBalancerInput{
		AvailabilityZones: p.AvailabilityZones,
		LoadBalancerName:  aws.String(name),
		Scheme:            p.Scheme,
		SecurityGroups:    p.SecurityGroups,
		Subnets:           p.Subnets,
	}
	input.Listeners = BuildELBListeners(p.Listeners)

	return &input
}

// LateInitializeELB fills the empty fields in *v1alpha1.ELBParameters with
// the values seen in elasticLoadBalancing.ELB.
func LateInitializeELB(in *v1alpha1.ELBParameters, v *elb.LoadBalancerDescription) { // nolint:gocyclo
	if v == nil {
		return
	}

	in.Scheme = clients.LateInitializeStringPtr(in.Scheme, v.Scheme)

	if len(in.AvailabilityZones) == 0 && len(v.AvailabilityZones) != 0 {
		in.AvailabilityZones = v.AvailabilityZones
	}

	if len(in.SecurityGroups) == 0 && len(v.SecurityGroups) != 0 {
		in.SecurityGroups = v.SecurityGroups
	}

	if len(in.Listeners) == 0 && len(v.ListenerDescriptions) != 0 {
		in.Listeners = make([]v1alpha1.Listener, len(v.ListenerDescriptions))
		for k, l := range v.ListenerDescriptions {
			in.Listeners[k] = v1alpha1.Listener{
				InstancePort:     aws.Int64Value(l.Listener.InstancePort),
				InstanceProtocol: l.Listener.InstanceProtocol,
				LoadBalancerPort: aws.Int64Value(l.Listener.LoadBalancerPort),
				Protocol:         aws.StringValue(l.Listener.Protocol),
				SSLCertificateID: l.Listener.SSLCertificateId,
			}
		}
	}
}

// IsELBNotFound returns true if the error is because the item doesn't exist.
func IsELBNotFound(err error) bool {
	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() == ELBNotFound {
			return true
		}
	}

	return false
}

// GenerateELBObservation is used to produce v1alpha1.ELBObservation from
// elasticLoadBalancing.LoadBalancerDescription.
func GenerateELBObservation(e elb.LoadBalancerDescription) v1alpha1.ELBObservation {
	o := v1alpha1.ELBObservation{
		CanonicalHostedZoneName:   aws.StringValue(e.CanonicalHostedZoneName),
		CanonicalHostedZoneNameID: aws.StringValue(e.CanonicalHostedZoneNameID),
		DNSName:                   aws.StringValue(e.DNSName),
		VPCID:                     aws.StringValue(e.VPCId),
	}

	if len(e.BackendServerDescriptions) > 0 {
		descriptions := []v1alpha1.BackendServerDescription{}
		for _, v := range e.BackendServerDescriptions {
			descriptions = append(descriptions, v1alpha1.BackendServerDescription{
				InstancePort: aws.Int64Value(v.InstancePort),
				PolicyNames:  v.PolicyNames,
			})
		}
		o.BackendServerDescriptions = descriptions
	}

	return o
}

// CreatePatch creates a v1alpha1.ELBParameters that has only the changed
// values between the target v1alpha1.ELBParameters and the current
// elb.LoadBalancerDescription.
func CreatePatch(in elb.LoadBalancerDescription, target v1alpha1.ELBParameters) (*v1alpha1.ELBParameters, error) {
	currentParams := &v1alpha1.ELBParameters{}
	LateInitializeELB(currentParams, &in)

	jsonPatch, err := clients.CreateJSONPatch(currentParams, target)
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
func IsUpToDate(p v1alpha1.ELBParameters, elb elb.LoadBalancerDescription) (bool, error) {
	patch, err := CreatePatch(elb, p)
	if err != nil {
		return false, err
	}
	return cmp.Equal(&v1alpha1.ELBParameters{}, patch), nil
}

// BuildELBListeners builds a list of elb.Listener from given list of v1alpha1.Listener.
func BuildELBListeners(listeners []v1alpha1.Listener) []elb.Listener {
	if len(listeners) > 0 {
		elbListeners := []elb.Listener{}
		for _, v := range listeners {
			elbListeners = append(elbListeners, elb.Listener{
				InstancePort:     &v.InstancePort,
				InstanceProtocol: v.InstanceProtocol,
				LoadBalancerPort: &v.LoadBalancerPort,
				Protocol:         &v.Protocol,
				SSLCertificateId: v.SSLCertificateID,
			})
		}
		return elbListeners
	}
	return nil
}
