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

package sns

import (
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/sns"

	"github.com/crossplane/provider-aws/apis/notification/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// TopicAttributes refers to AWS SNS Topic Attributes List
// ref: https://docs.aws.amazon.com/cli/latest/reference/sns/get-topic-attributes.html#output
type TopicAttributes string

const (
	// TopicDisplayName is Display Name of SNS Topic
	TopicDisplayName TopicAttributes = "DisplayName"
	// TopicDeliveryPolicy is Delivery Policy of SNS Topic
	TopicDeliveryPolicy TopicAttributes = "DeliveryPolicy"
	// TopicKmsMasterKeyID is KmsMasterKeyId of SNS Topic
	TopicKmsMasterKeyID TopicAttributes = "KmsMasterKeyId"
	// TopicPolicy is Policy of SNS Topic
	TopicPolicy TopicAttributes = "Policy"
	// TopicOwner is Owner of SNS Topic
	TopicOwner TopicAttributes = "Owner"
	// TopicSubscriptionsConfirmed is status of SNS Topic Subscription Confirmation
	TopicSubscriptionsConfirmed TopicAttributes = "SubscriptionsConfirmed"
	// TopicSubscriptionsPending is status of SNS Topic Subscription Confirmation
	TopicSubscriptionsPending TopicAttributes = "SubscriptionsPending"
	// TopicSubscriptionsDeleted is status of SNS Topic Subscription Confirmation
	TopicSubscriptionsDeleted TopicAttributes = "SubscriptionsDeleted"
	// TopicARN is the ARN for the SNS Topic
	TopicARN TopicAttributes = "TopicArn"
)

// TopicClient is the external client used for AWS SNSTopic
type TopicClient interface {
	CreateTopicRequest(*sns.CreateTopicInput) sns.CreateTopicRequest
	DeleteTopicRequest(*sns.DeleteTopicInput) sns.DeleteTopicRequest
	GetTopicAttributesRequest(*sns.GetTopicAttributesInput) sns.GetTopicAttributesRequest
	SetTopicAttributesRequest(*sns.SetTopicAttributesInput) sns.SetTopicAttributesRequest
}

// NewTopicClient returns a new client using AWS credentials as JSON encoded data.
func NewTopicClient(cfg aws.Config) TopicClient {
	return sns.New(cfg)
}

// GenerateCreateTopicInput prepares input for CreateTopicRequest
func GenerateCreateTopicInput(p *v1alpha1.SNSTopicParameters) *sns.CreateTopicInput {
	input := &sns.CreateTopicInput{
		Name: &p.Name,
	}

	if len(p.Tags) != 0 {
		input.Tags = make([]sns.Tag, len(p.Tags))
		for i, val := range p.Tags {
			input.Tags[i] = sns.Tag{
				Key:   aws.String(val.Key),
				Value: val.Value,
			}
		}
	}

	return input
}

// LateInitializeTopicAttr fills the empty fields in *v1alpha1.SNSTopicParameters with the
// values seen in sns.Topic.
func LateInitializeTopicAttr(in *v1alpha1.SNSTopicParameters, attrs map[string]string) {
	in.DisplayName = awsclients.LateInitializeStringPtr(in.DisplayName, aws.String(attrs[string(TopicDisplayName)]))
	in.DeliveryPolicy = awsclients.LateInitializeStringPtr(in.DeliveryPolicy, aws.String(attrs[string(TopicDeliveryPolicy)]))
	in.KMSMasterKeyID = awsclients.LateInitializeStringPtr(in.KMSMasterKeyID, aws.String(attrs[string(TopicKmsMasterKeyID)]))
	in.Policy = awsclients.LateInitializeStringPtr(in.Policy, aws.String(attrs[string(TopicPolicy)]))

}

// GetChangedAttributes will return the changed attributes for a topic in AWS side.
//
// This is needed as currently AWS SDK allows to set Attribute Topics one at a time.
// Please see https://docs.aws.amazon.com/sns/latest/api/API_SetTopicAttributes.html
// So we need to compare each topic attribute and call SetTopicAttribute for ones which has
// changed.
func GetChangedAttributes(p v1alpha1.SNSTopicParameters, attrs map[string]string) map[string]string {
	topicAttrs := getTopicAttributes(p)
	changedAttrs := make(map[string]string)
	for k, v := range topicAttrs {
		if v != attrs[k] {
			changedAttrs[k] = v
		}
	}

	return changedAttrs
}

// GenerateTopicObservation is used to produce SNSTopicObservation from attributes
func GenerateTopicObservation(attr map[string]string) v1alpha1.SNSTopicObservation {
	o := v1alpha1.SNSTopicObservation{}

	o.Owner = aws.String(attr[string(TopicOwner)])

	if s, err := strconv.ParseInt(attr[string(TopicSubscriptionsConfirmed)], 10, 64); err == nil {
		o.ConfirmedSubscriptions = aws.Int64(s)
	}

	if s, err := strconv.ParseInt(attr[string(TopicSubscriptionsPending)], 10, 64); err == nil {
		o.PendingSubscriptions = aws.Int64(s)
	}

	if s, err := strconv.ParseInt(attr[string(TopicSubscriptionsDeleted)], 10, 64); err == nil {
		o.DeletedSubscriptions = aws.Int64(s)
	}

	o.ARN = attr[string(TopicARN)]

	return o
}

// IsSNSTopicUpToDate checks if object is up to date
func IsSNSTopicUpToDate(p v1alpha1.SNSTopicParameters, attr map[string]string) bool {
	return aws.StringValue(p.DeliveryPolicy) == attr[string(TopicDeliveryPolicy)] &&
		aws.StringValue(p.DisplayName) == attr[string(TopicDisplayName)] &&
		aws.StringValue(p.KMSMasterKeyID) == attr[string(TopicKmsMasterKeyID)] &&
		aws.StringValue(p.Policy) == attr[string(TopicPolicy)]
}

func getTopicAttributes(p v1alpha1.SNSTopicParameters) map[string]string {

	topicAttr := make(map[string]string)

	topicAttr[string(TopicDeliveryPolicy)] = aws.StringValue(p.DeliveryPolicy)
	topicAttr[string(TopicDisplayName)] = aws.StringValue(p.DisplayName)
	topicAttr[string(TopicKmsMasterKeyID)] = aws.StringValue(p.KMSMasterKeyID)
	topicAttr[string(TopicPolicy)] = aws.StringValue(p.Policy)

	return topicAttr
}

// IsTopicNotFound returns true if the error code indicates that the item was not found
func IsTopicNotFound(err error) bool {
	if topicErr, ok := err.(awserr.Error); ok && topicErr.Code() == sns.ErrCodeNotFoundException {
		return true
	}
	return false
}
