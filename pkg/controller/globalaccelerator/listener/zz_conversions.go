/*
Copyright 2021 The Crossplane Authors.

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

// Code generated by ack-generate. DO NOT EDIT.

package listener

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	svcsdk "github.com/aws/aws-sdk-go/service/globalaccelerator"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/globalaccelerator/v1alpha1"
)

// NOTE(muvaf): We return pointers in case the function needs to start with an
// empty object, hence need to return a new pointer.

// GenerateDescribeListenerInput returns input for read
// operation.
func GenerateDescribeListenerInput(cr *svcapitypes.Listener) *svcsdk.DescribeListenerInput {
	res := &svcsdk.DescribeListenerInput{}

	if cr.Status.AtProvider.ListenerARN != nil {
		res.SetListenerArn(*cr.Status.AtProvider.ListenerARN)
	}

	return res
}

// GenerateListener returns the current state in the form of *svcapitypes.Listener.
func GenerateListener(resp *svcsdk.DescribeListenerOutput) *svcapitypes.Listener {
	cr := &svcapitypes.Listener{}

	if resp.Listener.ClientAffinity != nil {
		cr.Spec.ForProvider.ClientAffinity = resp.Listener.ClientAffinity
	} else {
		cr.Spec.ForProvider.ClientAffinity = nil
	}
	if resp.Listener.ListenerArn != nil {
		cr.Status.AtProvider.ListenerARN = resp.Listener.ListenerArn
	} else {
		cr.Status.AtProvider.ListenerARN = nil
	}
	if resp.Listener.PortRanges != nil {
		f2 := []*svcapitypes.PortRange{}
		for _, f2iter := range resp.Listener.PortRanges {
			f2elem := &svcapitypes.PortRange{}
			if f2iter.FromPort != nil {
				f2elem.FromPort = f2iter.FromPort
			}
			if f2iter.ToPort != nil {
				f2elem.ToPort = f2iter.ToPort
			}
			f2 = append(f2, f2elem)
		}
		cr.Spec.ForProvider.PortRanges = f2
	} else {
		cr.Spec.ForProvider.PortRanges = nil
	}
	if resp.Listener.Protocol != nil {
		cr.Spec.ForProvider.Protocol = resp.Listener.Protocol
	} else {
		cr.Spec.ForProvider.Protocol = nil
	}

	return cr
}

// GenerateCreateListenerInput returns a create input.
func GenerateCreateListenerInput(cr *svcapitypes.Listener) *svcsdk.CreateListenerInput {
	res := &svcsdk.CreateListenerInput{}

	if cr.Spec.ForProvider.ClientAffinity != nil {
		res.SetClientAffinity(*cr.Spec.ForProvider.ClientAffinity)
	}
	if cr.Spec.ForProvider.PortRanges != nil {
		f1 := []*svcsdk.PortRange{}
		for _, f1iter := range cr.Spec.ForProvider.PortRanges {
			f1elem := &svcsdk.PortRange{}
			if f1iter.FromPort != nil {
				f1elem.SetFromPort(*f1iter.FromPort)
			}
			if f1iter.ToPort != nil {
				f1elem.SetToPort(*f1iter.ToPort)
			}
			f1 = append(f1, f1elem)
		}
		res.SetPortRanges(f1)
	}
	if cr.Spec.ForProvider.Protocol != nil {
		res.SetProtocol(*cr.Spec.ForProvider.Protocol)
	}

	return res
}

// GenerateUpdateListenerInput returns an update input.
func GenerateUpdateListenerInput(cr *svcapitypes.Listener) *svcsdk.UpdateListenerInput {
	res := &svcsdk.UpdateListenerInput{}

	if cr.Spec.ForProvider.ClientAffinity != nil {
		res.SetClientAffinity(*cr.Spec.ForProvider.ClientAffinity)
	}
	if cr.Status.AtProvider.ListenerARN != nil {
		res.SetListenerArn(*cr.Status.AtProvider.ListenerARN)
	}
	if cr.Spec.ForProvider.PortRanges != nil {
		f2 := []*svcsdk.PortRange{}
		for _, f2iter := range cr.Spec.ForProvider.PortRanges {
			f2elem := &svcsdk.PortRange{}
			if f2iter.FromPort != nil {
				f2elem.SetFromPort(*f2iter.FromPort)
			}
			if f2iter.ToPort != nil {
				f2elem.SetToPort(*f2iter.ToPort)
			}
			f2 = append(f2, f2elem)
		}
		res.SetPortRanges(f2)
	}
	if cr.Spec.ForProvider.Protocol != nil {
		res.SetProtocol(*cr.Spec.ForProvider.Protocol)
	}

	return res
}

// GenerateDeleteListenerInput returns a deletion input.
func GenerateDeleteListenerInput(cr *svcapitypes.Listener) *svcsdk.DeleteListenerInput {
	res := &svcsdk.DeleteListenerInput{}

	if cr.Status.AtProvider.ListenerARN != nil {
		res.SetListenerArn(*cr.Status.AtProvider.ListenerARN)
	}

	return res
}

// IsNotFound returns whether the given error is of type NotFound or not.
func IsNotFound(err error) bool {
	awsErr, ok := err.(awserr.Error)
	return ok && awsErr.Code() == "ListenerNotFoundException"
}
