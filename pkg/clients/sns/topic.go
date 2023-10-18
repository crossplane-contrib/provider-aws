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

package sns

import (
	"context"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	snstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/pkg/errors"
	"k8s.io/utils/ptr"

	"github.com/crossplane-contrib/provider-aws/apis/sns/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	policyutils "github.com/crossplane-contrib/provider-aws/pkg/utils/policy"
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
	// TopicFifoTopic is whether or not Topic is fifo
	TopicFifoTopic TopicAttributes = "FifoTopic"
)

// TopicClient is the external client used for AWS Topic
type TopicClient interface {
	CreateTopic(ctx context.Context, input *sns.CreateTopicInput, opts ...func(*sns.Options)) (*sns.CreateTopicOutput, error)
	DeleteTopic(ctx context.Context, input *sns.DeleteTopicInput, opts ...func(*sns.Options)) (*sns.DeleteTopicOutput, error)
	GetTopicAttributes(ctx context.Context, input *sns.GetTopicAttributesInput, opts ...func(*sns.Options)) (*sns.GetTopicAttributesOutput, error)
	SetTopicAttributes(ctx context.Context, input *sns.SetTopicAttributesInput, opts ...func(*sns.Options)) (*sns.SetTopicAttributesOutput, error)
}

// NewTopicClient returns a new client using AWS credentials as JSON encoded data.
func NewTopicClient(cfg aws.Config) TopicClient {
	return sns.NewFromConfig(cfg)
}

// GenerateCreateTopicInput prepares input for CreateTopicRequest
func GenerateCreateTopicInput(p *v1beta1.TopicParameters) *sns.CreateTopicInput {
	attr := make(map[string]string)

	if p.FifoTopic != nil {
		attr["FifoTopic"] = strconv.FormatBool(*p.FifoTopic)
	}

	input := &sns.CreateTopicInput{
		Attributes: attr,
		Name:       &p.Name,
	}

	if len(p.Tags) != 0 {
		input.Tags = make([]snstypes.Tag, len(p.Tags))
		for i, val := range p.Tags {
			input.Tags[i] = snstypes.Tag{
				Key:   aws.String(val.Key),
				Value: val.Value,
			}
		}
	}

	return input
}

// LateInitializeTopicAttr fills the empty fields in *v1beta1.TopicParameters with the
// values seen in sns.Topic.
func LateInitializeTopicAttr(in *v1beta1.TopicParameters, attrs map[string]string) {
	in.DisplayName = pointer.LateInitialize(in.DisplayName, aws.String(attrs[string(TopicDisplayName)]))
	in.DeliveryPolicy = pointer.LateInitialize(in.DeliveryPolicy, aws.String(attrs[string(TopicDeliveryPolicy)]))
	in.KMSMasterKeyID = pointer.LateInitialize(in.KMSMasterKeyID, aws.String(attrs[string(TopicKmsMasterKeyID)]))
	in.Policy = pointer.LateInitialize(in.Policy, aws.String(attrs[string(TopicPolicy)]))

	in.FifoTopic = nil
	fifoTopic, err := strconv.ParseBool(attrs[string(TopicFifoTopic)])
	if err == nil && fifoTopic {
		in.FifoTopic = pointer.LateInitialize(in.FifoTopic, aws.Bool(fifoTopic))
	}
}

// GetChangedAttributes will return the changed attributes for a topic in AWS side.
//
// This is needed as currently AWS SDK allows to set Attribute Topics one at a time.
// Please see https://docs.aws.amazon.com/sns/latest/api/API_SetTopicAttributes.html
// So we need to compare each topic attribute and call SetTopicAttribute for ones which has
// changed.
func GetChangedAttributes(p v1beta1.TopicParameters, attrs map[string]string) (map[string]string, error) {
	topicAttrs := getTopicAttributes(p)
	changedAttrs := make(map[string]string)
	for k, v := range topicAttrs {
		if k == string(TopicPolicy) {
			isPolicyUpToDate, err := isSNSPolicyUpToDate(v, attrs[string(TopicPolicy)])
			if err != nil {
				return nil, errors.Wrap(err, "cannot compare policies")
			}
			if !isPolicyUpToDate {
				changedAttrs[k] = v
			}
		} else if v != attrs[k] {
			changedAttrs[k] = v
		}
	}

	return changedAttrs, nil
}

// GenerateTopicObservation is used to produce TopicObservation from attributes
func GenerateTopicObservation(attr map[string]string) v1beta1.TopicObservation {
	o := v1beta1.TopicObservation{}

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
func IsSNSTopicUpToDate(p v1beta1.TopicParameters, attr map[string]string) (bool, error) {
	fifoTopic, _ := strconv.ParseBool(attr[string(TopicFifoTopic)])

	isPolicyUpToDate, err := isSNSPolicyUpToDate(ptr.Deref(p.Policy, ""), attr[string(TopicPolicy)])
	if err != nil {
		return false, err
	}

	return aws.ToString(p.DeliveryPolicy) == attr[string(TopicDeliveryPolicy)] &&
		aws.ToString(p.DisplayName) == attr[string(TopicDisplayName)] &&
		aws.ToString(p.KMSMasterKeyID) == attr[string(TopicKmsMasterKeyID)] &&
		aws.ToBool(p.FifoTopic) == fifoTopic &&
		isPolicyUpToDate, nil
}

// IsSNSPolicyChanged determines whether a SNS topic policy needs to be updated
func isSNSPolicyUpToDate(specPolicyStr, currPolicyStr string) (bool, error) {
	if specPolicyStr == "" {
		return currPolicyStr == "", nil
	} else if currPolicyStr == "" {
		return false, nil
	}

	currPolicy, err := policyutils.ParsePolicyString(currPolicyStr)
	if err != nil {
		return false, errors.Wrap(err, "current policy")
	}
	specPolicy, err := policyutils.ParsePolicyString(specPolicyStr)
	if err != nil {
		return false, errors.Wrap(err, "spec policy")
	}
	equalPolicies, _ := policyutils.ArePoliciesEqal(&currPolicy, &specPolicy)
	return equalPolicies, nil
}

func getTopicAttributes(p v1beta1.TopicParameters) map[string]string {
	topicAttr := make(map[string]string)

	topicAttr[string(TopicDeliveryPolicy)] = aws.ToString(p.DeliveryPolicy)
	topicAttr[string(TopicDisplayName)] = aws.ToString(p.DisplayName)
	topicAttr[string(TopicKmsMasterKeyID)] = aws.ToString(p.KMSMasterKeyID)
	topicAttr[string(TopicPolicy)] = aws.ToString(p.Policy)

	fifoTopic := aws.ToBool(p.FifoTopic)
	if fifoTopic {
		topicAttr[string(TopicFifoTopic)] = strconv.FormatBool(fifoTopic)
	}

	return topicAttr
}

// IsTopicNotFound returns true if the error code indicates that the item was not found
func IsTopicNotFound(err error) bool {
	var nfe *snstypes.NotFoundException
	var rnfe *snstypes.ResourceNotFoundException
	return errors.As(err, &nfe) || errors.As(err, &rnfe)
}
