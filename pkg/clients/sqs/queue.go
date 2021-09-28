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

package sqs

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/smithy-go"
	"github.com/google/go-cmp/cmp"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"

	"github.com/crossplane/provider-aws/apis/sqs/v1beta1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

const (
	// QueueNotFound is the code that is returned by AWS when the given QueueURL is not valid
	QueueNotFound = "AWS.SimpleQueueService.NonExistentQueue"
)

// Client defines Queue client operations
type Client interface {
	CreateQueue(ctx context.Context, input *sqs.CreateQueueInput, opts ...func(*sqs.Options)) (*sqs.CreateQueueOutput, error)
	DeleteQueue(ctx context.Context, input *sqs.DeleteQueueInput, opts ...func(*sqs.Options)) (*sqs.DeleteQueueOutput, error)
	TagQueue(ctx context.Context, input *sqs.TagQueueInput, opts ...func(*sqs.Options)) (*sqs.TagQueueOutput, error)
	UntagQueue(ctx context.Context, input *sqs.UntagQueueInput, opts ...func(*sqs.Options)) (*sqs.UntagQueueOutput, error)
	ListQueueTags(ctx context.Context, input *sqs.ListQueueTagsInput, opts ...func(*sqs.Options)) (*sqs.ListQueueTagsOutput, error)
	GetQueueAttributes(ctx context.Context, input *sqs.GetQueueAttributesInput, opts ...func(*sqs.Options)) (*sqs.GetQueueAttributesOutput, error)
	SetQueueAttributes(ctx context.Context, input *sqs.SetQueueAttributesInput, opts ...func(*sqs.Options)) (*sqs.SetQueueAttributesOutput, error)
	GetQueueUrl(ctx context.Context, input *sqs.GetQueueUrlInput, opts ...func(*sqs.Options)) (*sqs.GetQueueUrlOutput, error)
}

// NewClient returns a new SQS Client.
func NewClient(cfg aws.Config) Client {
	return sqs.NewFromConfig(cfg)
}

// GenerateCreateAttributes returns a map of queue attributes for Create operation
func GenerateCreateAttributes(p *v1beta1.QueueParameters) map[string]string {
	m := GenerateQueueAttributes(p)
	if aws.ToBool(p.FIFOQueue) {
		// SQS expects this attribute only if its value is true.
		// https://github.com/aws/aws-sdk-php/issues/1331
		if m == nil {
			m = map[string]string{}
		}
		m[v1beta1.AttributeFifoQueue] = "true"
	}
	return m
}

// GenerateQueueAttributes returns a map of queue attributes
func GenerateQueueAttributes(p *v1beta1.QueueParameters) map[string]string { // nolint:gocyclo
	m := map[string]string{}
	if p.DelaySeconds != nil {
		m[v1beta1.AttributeDelaySeconds] = strconv.FormatInt(aws.ToInt64(p.DelaySeconds), 10)
	}
	if p.MaximumMessageSize != nil {
		m[v1beta1.AttributeMaximumMessageSize] = strconv.FormatInt(aws.ToInt64(p.MaximumMessageSize), 10)
	}
	if p.MessageRetentionPeriod != nil {
		m[v1beta1.AttributeMessageRetentionPeriod] = strconv.FormatInt(aws.ToInt64(p.MessageRetentionPeriod), 10)
	}
	if p.Policy != nil {
		m[v1beta1.AttributePolicy] = aws.ToString(p.Policy)
	}
	if p.ReceiveMessageWaitTimeSeconds != nil {
		m[v1beta1.AttributeReceiveMessageWaitTimeSeconds] = strconv.FormatInt(aws.ToInt64(p.ReceiveMessageWaitTimeSeconds), 10)
	}
	if p.ReceiveMessageWaitTimeSeconds != nil {
		m[v1beta1.AttributeReceiveMessageWaitTimeSeconds] = strconv.FormatInt(aws.ToInt64(p.ReceiveMessageWaitTimeSeconds), 10)
	}
	if p.RedrivePolicy != nil && aws.ToString(p.RedrivePolicy.DeadLetterTargetARN) != "" {
		r := map[string]interface{}{
			"deadLetterTargetArn": p.RedrivePolicy.DeadLetterTargetARN,
			"maxReceiveCount":     p.RedrivePolicy.MaxReceiveCount,
		}
		val, err := json.Marshal(r)
		if err == nil {
			m[v1beta1.AttributeRedrivePolicy] = string(val)
		}
	}
	if p.VisibilityTimeout != nil {
		m[v1beta1.AttributeVisibilityTimeout] = strconv.FormatInt(aws.ToInt64(p.VisibilityTimeout), 10)
	}
	if p.KMSMasterKeyID != nil {
		m[v1beta1.AttributeKmsMasterKeyID] = aws.ToString(p.KMSMasterKeyID)
	}
	if p.KMSDataKeyReusePeriodSeconds != nil {
		m[v1beta1.AttributeKmsDataKeyReusePeriodSeconds] = strconv.FormatInt(aws.ToInt64(p.KMSDataKeyReusePeriodSeconds), 10)
	}
	if p.ContentBasedDeduplication != nil {
		m[v1beta1.AttributeContentBasedDeduplication] = strconv.FormatBool(aws.ToBool(p.ContentBasedDeduplication))
	}
	if len(m) == 0 {
		return nil
	}
	return m
}

// GenerateQueueObservation returns a QueueObservation with information retrieved
// from AWS.
func GenerateQueueObservation(url string, attr map[string]string) v1beta1.QueueObservation {
	o := v1beta1.QueueObservation{
		URL: url,
		ARN: attr[v1beta1.AttributeQueueArn],
	}
	o.ApproximateNumberOfMessages, _ = strconv.ParseInt(attr[v1beta1.AttributeApproximateNumberOfMessages], 10, 0)
	o.ApproximateNumberOfMessagesDelayed, _ = strconv.ParseInt(attr[v1beta1.AttributeApproximateNumberOfMessagesDelayed], 10, 0)
	o.ApproximateNumberOfMessagesNotVisible, _ = strconv.ParseInt(attr[v1beta1.AttributeApproximateNumberOfMessagesNotVisible], 10, 0)
	if i, err := strconv.ParseInt(attr[v1beta1.AttributeCreatedTimestamp], 10, 64); err == nil {
		t := metav1.NewTime(time.Unix(i, 0))
		o.CreatedTimestamp = &t
	}
	if i, err := strconv.ParseInt(attr[v1beta1.AttributeLastModifiedTimestamp], 10, 64); err == nil {
		t := metav1.NewTime(time.Unix(i, 0))
		o.LastModifiedTimestamp = &t
	}
	return o
}

// IsNotFound checks if the error returned by AWS API says that the queue being probed doesn't exist
func IsNotFound(err error) bool {
	if awsErr, ok := err.(smithy.APIError); ok {
		if awsErr.ErrorCode() == QueueNotFound {
			return true
		}
	}
	return false
}

// LateInitialize fills the empty fields in *v1beta1.QueueParameters with
// the values seen in queue.Attributes
func LateInitialize(in *v1beta1.QueueParameters, attributes map[string]string, tags map[string]string) {
	if in.Tags == nil && len(tags) > 0 {
		in.Tags = map[string]string{}
		for k, v := range tags {
			in.Tags[k] = v
		}
	}

	in.DelaySeconds = awsclients.LateInitializeInt64Ptr(in.DelaySeconds, int64Ptr(attributes[v1beta1.AttributeDelaySeconds]))
	in.KMSDataKeyReusePeriodSeconds = awsclients.LateInitializeInt64Ptr(in.KMSDataKeyReusePeriodSeconds, int64Ptr(attributes[v1beta1.AttributeKmsDataKeyReusePeriodSeconds]))
	in.MaximumMessageSize = awsclients.LateInitializeInt64Ptr(in.MaximumMessageSize, int64Ptr(attributes[v1beta1.AttributeMaximumMessageSize]))
	in.MessageRetentionPeriod = awsclients.LateInitializeInt64Ptr(in.MessageRetentionPeriod, int64Ptr(attributes[v1beta1.AttributeMessageRetentionPeriod]))
	in.ReceiveMessageWaitTimeSeconds = awsclients.LateInitializeInt64Ptr(in.ReceiveMessageWaitTimeSeconds, int64Ptr(attributes[v1beta1.AttributeReceiveMessageWaitTimeSeconds]))
	in.VisibilityTimeout = awsclients.LateInitializeInt64Ptr(in.VisibilityTimeout, int64Ptr(attributes[v1beta1.AttributeVisibilityTimeout]))
	if in.KMSMasterKeyID == nil && attributes[v1beta1.AttributeKmsMasterKeyID] != "" {
		in.KMSMasterKeyID = aws.String(attributes[v1beta1.AttributeKmsMasterKeyID])
	}
}

// IsUpToDate checks whether there is a change in any of the modifiable fields.
func IsUpToDate(p v1beta1.QueueParameters, attributes map[string]string, tags map[string]string) bool { // nolint:gocyclo
	if len(p.Tags) != len(tags) {
		return false
	}

	for k, v := range p.Tags {
		pVal, ok := tags[k]
		if !ok || !strings.EqualFold(pVal, v) {
			return false
		}
	}

	if aws.ToInt64(p.DelaySeconds) != toInt64(attributes[v1beta1.AttributeDelaySeconds]) {
		return false
	}
	if aws.ToInt64(p.KMSDataKeyReusePeriodSeconds) != toInt64(attributes[v1beta1.AttributeKmsDataKeyReusePeriodSeconds]) {
		return false
	}
	if aws.ToInt64(p.MaximumMessageSize) != toInt64(attributes[v1beta1.AttributeMaximumMessageSize]) {
		return false
	}
	if aws.ToInt64(p.MessageRetentionPeriod) != toInt64(attributes[v1beta1.AttributeMessageRetentionPeriod]) {
		return false
	}
	if aws.ToInt64(p.ReceiveMessageWaitTimeSeconds) != toInt64(attributes[v1beta1.AttributeReceiveMessageWaitTimeSeconds]) {
		return false
	}
	if aws.ToInt64(p.VisibilityTimeout) != toInt64(attributes[v1beta1.AttributeVisibilityTimeout]) {
		return false
	}
	if !cmp.Equal(aws.ToString(p.KMSMasterKeyID), attributes[v1beta1.AttributeKmsMasterKeyID]) {
		return false
	}
	if !cmp.Equal(aws.ToString(p.Policy), attributes[v1beta1.AttributePolicy]) {
		return false
	}
	if attributes[v1beta1.AttributeContentBasedDeduplication] != "" && strconv.FormatBool(aws.ToBool(p.ContentBasedDeduplication)) != attributes[v1beta1.AttributeContentBasedDeduplication] {
		return false
	}
	if p.RedrivePolicy != nil {
		r := map[string]interface{}{
			"deadLetterTargetArn": p.RedrivePolicy.DeadLetterTargetARN,
			"maxReceiveCount":     p.RedrivePolicy.MaxReceiveCount,
		}
		val, err := json.Marshal(r)
		if err == nil {
			if string(val) != attributes[v1beta1.AttributeRedrivePolicy] {
				return false
			}
		}
	}
	return true
}

// TagsDiff returns the tags added and removed from spec when compared to the AWS SQS tags.
func TagsDiff(sqsTags map[string]string, newTags map[string]string) (removed, added map[string]string) {
	removed = map[string]string{}
	for k, v := range sqsTags {
		if _, ok := newTags[k]; !ok {
			removed[k] = v
		}
	}

	added = map[string]string{}
	for k, newV := range newTags {
		if oldV, ok := sqsTags[k]; !ok || oldV != newV {
			added[k] = newV
		}
	}
	return
}

func toInt64(s string) int64 {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return v
}

func int64Ptr(s string) *int64 {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil
	}
	return &v
}

// GetConnectionDetails extracts managed.ConnectionDetails out of v1beta1.Queue.
func GetConnectionDetails(in v1beta1.Queue) managed.ConnectionDetails {
	if in.Status.AtProvider.URL == "" {
		return nil
	}
	return managed.ConnectionDetails{
		xpv1.ResourceCredentialsSecretEndpointKey: []byte(in.Status.AtProvider.URL),
	}
}
