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

// CustomWorkGroupParameters contains the additional fields for WorkGroupParameters.
type CustomWorkGroupParameters struct{}

// CustomCapacityReservationParameters includes the custom fields of CapacityReservation
// ToDo: implement all custom parameters
type CustomCapacityReservationParameters struct {

	// Metadata tagging key value pairs
	// +optional
	Tags []Tag `json:"tags,omitempty"`
}
