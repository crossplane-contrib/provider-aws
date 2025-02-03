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

package v1alpha1

import xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

// CustomAcceleratorParameters contains the additional fields for AcceleratorParameters
type CustomAcceleratorParameters struct{}

// CustomAcceleratorObservation includes the custom status fields of Accelerator.
type CustomAcceleratorObservation struct{}

// CustomEndpointGroupParameters contains the additional fields for EndpointGroupParameters
type CustomEndpointGroupParameters struct {
	// ListenerArn is the ARN for the Listener.
	// +immutable
	// +crossplane:generate:reference:type=Listener
	// +crossplane:generate:reference:extractor=ListenerARN()
	// +crossplane:generate:reference:refFieldName=ListenerArnRef
	// +crossplane:generate:reference:selectorFieldName=ListenerArnSelector
	ListenerARN *string `json:"listenerArn,omitempty"`

	// ListenerArnRef is a reference to an ARN used to set
	// the ListenerArn.
	// +optional
	ListenerArnRef *xpv1.Reference `json:"listenerArnRef,omitempty"`

	// ListenerArnSelector selects references to Listener used
	// to set the Arn.
	// +optional
	ListenerArnSelector *xpv1.Selector `json:"listenerArnSelector,omitempty"`
}

// CustomEndpointGroupObservation includes the custom status fields of EndpointGroup.
type CustomEndpointGroupObservation struct{}

// CustomListenerParameters contains the additional fields for ListenerParameters
type CustomListenerParameters struct {
	// AcceleratorArn is the ARN for the Accelerator.
	// +immutable
	// +crossplane:generate:reference:type=Accelerator
	// +crossplane:generate:reference:extractor=AcceleratorARN()
	// +crossplane:generate:reference:refFieldName=AcceleratorArnRef
	// +crossplane:generate:reference:selectorFieldName=AcceleratorArnSelector
	AcceleratorArn *string `json:"acceleratorArn,omitempty"`

	// AcceleratorArnRef is a reference to an ARN used to set
	// the AcceleratorArn.
	// +optional
	AcceleratorArnRef *xpv1.Reference `json:"acceleratorArnRef,omitempty"`

	// AcceleratorArnSelector selects references to Accelerator used
	// to set the Arn.
	// +optional
	AcceleratorArnSelector *xpv1.Selector `json:"acceleratorArnSelector,omitempty"`
}

// CustomListenerObservation includes the custom status fields of Listener.
type CustomListenerObservation struct{}
