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
)

// CustomStreamParameters contains the additional fields for StreamParameters.
type CustomStreamParameters struct {
	// The retention period of the stream, in hours.
	// Default: 24 hours
	RetentionPeriodHours *int64 `json:"retentionPeriodHours,omitempty"`

	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/kms/v1alpha1.Key
	// +crossplane:generate:reference:extractor=github.com/crossplane-contrib/provider-aws/apis/kms/v1alpha1.KMSKeyARN()
	KMSKeyARN *string `json:"kmsKeyARN,omitempty"`

	KMSKeyARNRef *xpv1.Reference `json:"kmsKeyARNRef,omitempty"`

	KMSKeyARNSelector *xpv1.Selector `json:"kmsKeyARNSelector,omitempty"`

	// List of shard-level metrics.
	//
	// The following are the valid shard-level metrics. The value "ALL" enhances
	// every metric.
	//
	//    * IncomingBytes
	//
	//    * IncomingRecords
	//
	//    * OutgoingBytes
	//
	//    * OutgoingRecords
	//
	//    * WriteProvisionedThroughputExceeded
	//
	//    * ReadProvisionedThroughputExceeded
	//
	//    * IteratorAgeMilliseconds
	//
	//    * ALL
	//
	// For more information, see Monitoring the Amazon Kinesis Data Streams Service
	// with Amazon CloudWatch (https://docs.aws.amazon.com/kinesis/latest/dev/monitoring-with-cloudwatch.html)
	// in the Amazon Kinesis Data Streams Developer Guide.
	EnhancedMetrics []*EnhancedMetrics `json:"enhancedMetrics,omitempty"`

	Tags []CustomTag `json:"tags,omitempty"`

	// If this parameter is unset (null) or if you set it to false, and the stream
	// has registered consumers, the call to DeleteStream fails with a ResourceInUseException.
	EnforceConsumerDeletion *bool `json:"enforceConsumerDeletion,omitempty"`
}

// CustomStreamObservation includes the custom status fields of Stream.
type CustomStreamObservation struct{}

// CustomTag contains the additional fields for Tag.
type CustomTag struct {
	// A unique identifier for the tag.
	Key string `json:"key"`

	// An optional string, typically used to describe or define the tag.
	Value string `json:"value,omitempty"`
}
