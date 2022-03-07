/*
Copyright 2022 The Crossplane Authors.

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

package v1beta1

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/crossplane/crossplane-runtime/pkg/errors"
)

// NOTE(muvaf): In the future, we will generate this function using controller-tools.

// +kubebuilder:webhook:verbs=update,path=/validate-ec2-aws-crossplane-io-v1beta1-vpc,mutating=false,failurePolicy=fail,groups=ec2.aws.crossplane.io,resources=vpcs,versions=v1beta1,name=vpcs.ec2.aws.crossplane.io,sideEffects=None,admissionReviewVersions=v1

// VPCImmutabilityCheck makes sure immutable fields that were set initially are
// not changed in given update call.
func VPCImmutabilityCheck(_ context.Context, oldObj, newObj runtime.Object) error {
	oldVPC, ok := oldObj.(*VPC)
	if !ok {
		return errors.Errorf("unexpected type")
	}
	newVPC, ok := newObj.(*VPC)
	if !ok {
		return errors.Errorf("unexpected type")
	}
	oldRegion := ""
	if oldVPC.Spec.ForProvider.Region != nil {
		oldRegion = *oldVPC.Spec.ForProvider.Region
	}
	newRegion := ""
	if newVPC.Spec.ForProvider.Region != nil {
		newRegion = *newVPC.Spec.ForProvider.Region
	}
	// To handle the case where user starts to use this field instead of ProviderConfig.
	if oldRegion == "" && newRegion != "" {
		return nil
	}
	if oldRegion != newRegion {
		return errors.Errorf("spec.forProvider.region is immutable")
	}
	return nil
}

func init() {
	VPCValidator.UpdateChain = append(VPCValidator.UpdateChain, VPCImmutabilityCheck)
}
