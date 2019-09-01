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

// Generate deepcopy for apis
//go:generate go run ../../vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go object:headerFile=../../hack/boilerplate.go.txt paths=./...

// Package apis contains Kubernetes API groups for AWS cloud provider.
package apis

import (
	"k8s.io/apimachinery/pkg/runtime"

	cachev1alpha1 "github.com/crossplaneio/stack-aws/aws/apis/cache/v1alpha1"
	computev1alpha1 "github.com/crossplaneio/stack-aws/aws/apis/compute/v1alpha1"
	databasev1alpha1 "github.com/crossplaneio/stack-aws/aws/apis/database/v1alpha1"
	storagev1alpha1 "github.com/crossplaneio/stack-aws/aws/apis/storage/v1alpha1"
	awsv1alpha1 "github.com/crossplaneio/stack-aws/aws/apis/v1alpha1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes,
		cachev1alpha1.SchemeBuilder.AddToScheme,
		computev1alpha1.SchemeBuilder.AddToScheme,
		databasev1alpha1.SchemeBuilder.AddToScheme,
		awsv1alpha1.SchemeBuilder.AddToScheme,
		storagev1alpha1.SchemeBuilder.AddToScheme,
	)
}

// AddToSchemes may be used to add all resources defined in the project to a Scheme
var AddToSchemes runtime.SchemeBuilder

// AddToScheme adds all Resources to the Scheme
func AddToScheme(s *runtime.Scheme) error {
	return AddToSchemes.AddToScheme(s)
}
