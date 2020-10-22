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

package v1beta1

// AccelerateConfiguration configures the transfer acceleration state for an
// Amazon S3 bucket. For more information, see Amazon S3 Transfer Acceleration
// (https://docs.aws.amazon.com/AmazonS3/latest/dev/transfer-acceleration.html)
// in the Amazon Simple Storage Service Developer Guide.
type AccelerateConfiguration struct {
	// Status specifies the transfer acceleration status of the bucket.
	// +kubebuilder:validation:Enum=Enabled;Suspended
	Status string `json:"status"`
}
