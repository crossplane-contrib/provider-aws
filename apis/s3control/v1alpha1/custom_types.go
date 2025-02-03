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

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

	"github.com/crossplane-contrib/provider-aws/apis/s3/common"
)

// CustomAccessPointParameters includes the custom fields of Stage.
type CustomAccessPointParameters struct {
	// BucketName is the name of the Bucket for AccessPoint
	// +immutable
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/s3/v1beta1.Bucket
	BucketName *string `json:"bucketName,omitempty"`

	// BucketNameRef is a reference to a Bucket used to set the BucketName
	// +optional
	BucketNameRef *xpv1.Reference `json:"bucketNameRef,omitempty"`

	// BucketNameSelector selects a references to used to set the BucketName
	// +optional
	BucketNameSelector *xpv1.Selector `json:"bucketNameSelector,omitempty"`

	// The policy that you want to apply to the specified access point. For more
	// information about access point policies, see Managing data access with Amazon
	// S3 access points (https://docs.aws.amazon.com/AmazonS3/latest/userguide/access-points.html)
	// in the Amazon S3 User Guide.
	// +optional
	Policy *common.BucketPolicyBody `json:"policy"`
}

// CustomAccessPointObservation includes the custom status fields of AccessPoint.
type CustomAccessPointObservation struct{}
