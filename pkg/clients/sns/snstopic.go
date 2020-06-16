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
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"

	"github.com/crossplane/crossplane-runtime/pkg/meta"

	"github.com/crossplane/provider-aws/apis/notification/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

const (
	//SNSTopicNotFound is the error code that is returned if SNS Topic is not present
	SNSTopicNotFound = "InvalidSNSTopic.NotFound"
)

// IsErrorTopicNotFound returns true if the error code indicates that the
// item was not found
func IsErrorTopicNotFound(err error) bool {
	if _, ok := err.(*TopicNotFound); ok {
		return true
	}
	return false
}

// TopicNotFound will be raised when there is no SNSTopic
type TopicNotFound struct{}

func (err *TopicNotFound) Error() string {
	return fmt.Sprint(SNSTopicNotFound)
}

// TopicClient is the external client used for SNSTopic custom resource
type TopicClient interface {
	CreateTopicRequest(*sns.CreateTopicInput) sns.CreateTopicRequest
	ListTopicsRequest(*sns.ListTopicsInput) sns.ListTopicsRequest
	DeleteTopicRequest(*sns.DeleteTopicInput) sns.DeleteTopicRequest
	GetTopicAttributesRequest(*sns.GetTopicAttributesInput) sns.GetTopicAttributesRequest
	SetTopicAttributesRequest(*sns.SetTopicAttributesInput) sns.SetTopicAttributesRequest
}

// NewTopicClient returns a new client using AWS credentials as JSON encoded data.
func NewTopicClient(conf *aws.Config) (TopicClient, error) {
	return sns.New(*conf), nil
}

// GetSNSTopic returns SNSTopic if present in list of topics
func GetSNSTopic(res *sns.ListTopicsResponse, cr *v1alpha1.SNSTopic) (sns.Topic, error) {

	for _, topic := range res.Topics {
		if aws.StringValue(topic.TopicArn) == meta.GetExternalName(cr) {
			return topic, nil
		}
	}
	return sns.Topic{}, &TopicNotFound{}
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
				Value: aws.String(val.Value),
			}
		}
	}

	return input
}

// LateInitializeTopic fills the empty fields in *v1alpha1.SNSTopicParameters with the
// values seen in sns.Topic.
func LateInitializeTopic(in *v1alpha1.SNSTopicParameters, topic sns.Topic, attrs map[string]string) {
	in.Name = *awsclients.LateInitializeStringPtr(&in.Name, topic.TopicArn)
	in.DisplayName = awsclients.LateInitializeStringPtr(in.DisplayName, aws.String(attrs["DisplayName"]))
	in.DeliveryPolicy = awsclients.LateInitializeStringPtr(in.DeliveryPolicy, aws.String(attrs["DeliveryPolicy"]))
	in.KmsMasterKeyID = awsclients.LateInitializeStringPtr(in.KmsMasterKeyID, aws.String(attrs["KmsMasterKeyId"]))
	in.Policy = awsclients.LateInitializeStringPtr(in.Policy, aws.String(attrs["Policy"]))

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

	o.Owner = aws.String(attr["Owner"])

	if s, err := strconv.ParseInt(attr["SubscriptionsConfirmed"], 10, 64); err == nil {
		o.ConfirmedSubscriptions = aws.Int64(s)
	}

	if s, err := strconv.ParseInt(attr["SubscriptionsPending"], 10, 64); err == nil {
		o.PendingSubscriptions = aws.Int64(s)
	}

	if s, err := strconv.ParseInt(attr["SubscriptionsDeleted"], 10, 64); err == nil {
		o.DeletedSubscriptions = aws.Int64(s)
	}

	return o
}

// IsSNSTopicUpToDate checks if object is up to date
func IsSNSTopicUpToDate(p v1alpha1.SNSTopicParameters, attr map[string]string) bool {
	return aws.StringValue(p.DeliveryPolicy) == attr["DeliveryPolicy"] &&
		aws.StringValue(p.DisplayName) == attr["DisplayName"] &&
		aws.StringValue(p.KmsMasterKeyID) == attr["KmsMasterKeyId"] &&
		aws.StringValue(p.Policy) == attr["Policy"]
}

func getTopicAttributes(p v1alpha1.SNSTopicParameters) map[string]string {

	topicAttr := make(map[string]string)

	topicAttr["DeliveryPolicy"] = aws.StringValue(p.DeliveryPolicy)
	topicAttr["DisplayName"] = aws.StringValue(p.DisplayName)
	topicAttr["KmsMasterKeyId"] = aws.StringValue(p.KmsMasterKeyID)
	topicAttr["Policy"] = aws.StringValue(p.Policy)

	return topicAttr
}
