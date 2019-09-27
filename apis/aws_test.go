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

// Package apis contains Kubernetes API groups for AWS cloud provider.
package apis

import (
	"testing"

	cachev1alpha2 "github.com/crossplaneio/stack-aws/apis/cache/v1alpha2"
	computev1alpha2 "github.com/crossplaneio/stack-aws/apis/compute/v1alpha2"
	databasev1alpha2 "github.com/crossplaneio/stack-aws/apis/database/v1alpha2"
	storagev1alpha2 "github.com/crossplaneio/stack-aws/apis/storage/v1alpha2"
	awsv1alpha2 "github.com/crossplaneio/stack-aws/apis/v1alpha2"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestAddToScheme(t *testing.T) {
	s := runtime.NewScheme()
	if err := AddToScheme(s); err != nil {
		t.Errorf("AddToScheme() error = %v", err)
	}
	gvs := []schema.GroupVersion{
		awsv1alpha2.SchemeGroupVersion,
		cachev1alpha2.SchemeGroupVersion,
		computev1alpha2.SchemeGroupVersion,
		databasev1alpha2.SchemeGroupVersion,
		storagev1alpha2.SchemeGroupVersion,
	}
	for _, gv := range gvs {
		if !s.IsVersionRegistered(gv) {
			t.Errorf("AddToScheme() %v should be registered", gv)
		}
	}
}
