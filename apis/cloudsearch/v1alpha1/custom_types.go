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

// CustomDomainParameters includes the custom fields of a CloudSearch Domain.
type CustomDomainParameters struct {

	// The desired number of search instances that are available to process search requests.
	// +optional
	DesiredReplicationCount *int64 `json:"desiredReplicationCount,omitempty"`

	// The desired instance type that is being used to process search requests.
	// +optional
	DesiredInstanceType *string `json:"desiredInstanceType,omitempty"`

	// The desired number of partitions across which the search index is spread.
	// +optional
	DesiredPartitionCount *int64 `json:"desiredPartitionCount,omitempty"`

	// The access rules you want to configure.
	// +optional
	AccessPolicies *string `json:"accessPolicies,omitempty"`
}

// CustomDomainParameters includes the custom status fields of a CloudSearch Domain.
type CustomDomainObservation struct{}
