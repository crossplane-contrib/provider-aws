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

//go:generate go run github.com/crossplane/crossplane-tools/cmd/angryjet generate-methodsets ./...

package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// CustomAddonParameters contains the additional fields for AddonParameters.
type CustomAddonParameters struct {
	// The name of the cluster to create the add-on for.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/eks/v1beta1.Cluster
	// +crossplane:generate:reference:refFieldName=ClusterNameRef
	// +crossplane:generate:reference:selectorFieldName=ClusterNameSelector
	ClusterName *string `json:"clusterName,omitempty"`

	// ClusterNameRef is a reference to a Cluster used to set
	// the ClusterName.
	// +immutable
	// +optional
	ClusterNameRef *xpv1.Reference `json:"clusterNameRef,omitempty"`

	// ClusterNameSelector selects references to a Cluster used
	// to set the ClusterName.
	// +immutable
	// +optional
	ClusterNameSelector *xpv1.Selector `json:"clusterNameSelector,omitempty"`
}

// CustomAddonObservation includes the custom status fields of Addon.
type CustomAddonObservation struct{}
