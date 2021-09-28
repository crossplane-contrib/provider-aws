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

package eks

import (
	"github.com/aws/aws-sdk-go-v2/service/eks"
	ekstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane/provider-aws/apis/eks/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// GenerateCreateFargateProfileInput from FargateProfileInputParameters.
func GenerateCreateFargateProfileInput(name string, p v1alpha1.FargateProfileParameters) *eks.CreateFargateProfileInput {
	c := &eks.CreateFargateProfileInput{
		FargateProfileName:  &name,
		ClusterName:         &p.ClusterName,
		PodExecutionRoleArn: &p.PodExecutionRoleArn,
		Subnets:             p.Subnets,
		Tags:                p.Tags,
	}
	if len(p.Selectors) > 0 {
		c.Selectors = make([]ekstypes.FargateProfileSelector, len(p.Selectors))
		for i, sel := range p.Selectors {
			c.Selectors[i] = ekstypes.FargateProfileSelector{
				Labels:    sel.Labels,
				Namespace: sel.Namespace,
			}
		}
	}
	return c
}

// GenerateFargateProfileObservation is used to produce v1alpha1.FargateProfileObservation
// from eks.FargateProfile.
func GenerateFargateProfileObservation(fp *ekstypes.FargateProfile) v1alpha1.FargateProfileObservation { // nolint:gocyclo
	if fp == nil {
		return v1alpha1.FargateProfileObservation{}
	}
	o := v1alpha1.FargateProfileObservation{
		FargateProfileArn: awsclients.StringValue(fp.FargateProfileArn),
		Status:            v1alpha1.FargateProfileStatusType(fp.Status),
	}
	if fp.CreatedAt != nil {
		o.CreatedAt = &metav1.Time{Time: *fp.CreatedAt}
	}
	return o
}

// LateInitializeFargateProfile fills the empty fields in *v1alpha1.FargateProfileParameters with the
// values seen in eks.FargateProfile.
func LateInitializeFargateProfile(in *v1alpha1.FargateProfileParameters, fp *ekstypes.FargateProfile) { // nolint:gocyclo
	if fp == nil {
		return
	}

	if len(in.Subnets) == 0 && len(fp.Subnets) > 0 {
		in.Subnets = fp.Subnets
	}
	// NOTE(hasheddan): we always will set the default Crossplane tags in
	// practice during initialization in the controller, but we check if no tags
	// exist for consistency with expected late initialization behavior.
	if len(in.Tags) == 0 {
		in.Tags = fp.Tags
	}
}

// IsFargateProfileUpToDate checks whether there is a change in the tags.
// Any other field is immutable and can't be updated.
func IsFargateProfileUpToDate(p v1alpha1.FargateProfileParameters, fp *ekstypes.FargateProfile) bool { // nolint:gocyclo
	return cmp.Equal(p.Tags, fp.Tags, cmpopts.EquateEmpty())
}
