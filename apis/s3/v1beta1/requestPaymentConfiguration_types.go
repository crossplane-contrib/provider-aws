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

// PaymentConfiguration specifies who pays for the download and request fees.
type PaymentConfiguration struct {
	// Payer is a required field, detailing who pays
	// Valid values are "Requester" and "BucketOwner"
	// +kubebuilder:validation:Enum=Requester;BucketOwner
	Payer string `json:"payer"`
}
