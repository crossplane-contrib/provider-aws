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
	"github.com/aws/aws-sdk-go-v2/service/sns"

	"github.com/crossplane/provider-aws/apis/notification/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// SubscriptionAttributes refers to AWS SNS Subscription Attributes List
// ref: https://docs.aws.amazon.com/cli/latest/reference/sns/get-subscription-attributes.html#output
type SubscriptionAttributes string

const (
	// SubscriptionDeliveryPolicy is DeliveryPolicy of SNS Subscription
	SubscriptionDeliveryPolicy SubscriptionAttributes = "DeliveryPolicy"
	// SubscriptionFilterPolicy is FilterPolicy of SNS Subscription
	SubscriptionFilterPolicy SubscriptionAttributes = "FilterPolicy"
	// SubscriptionRawMessageDelivery is RawMessageDelivery of SNS Subscription
	SubscriptionRawMessageDelivery SubscriptionAttributes = "RawMessageDelivery"
	// SubscriptionRedrivePolicy is RedrivePolicy of SNS Subscription
	SubscriptionRedrivePolicy SubscriptionAttributes = "RedrivePolicy"
	// SubscriptionOwner is Owner of SNS Subscription
	SubscriptionOwner SubscriptionAttributes = "Owner"
	// SubscriptionPendingConfirmation is Confirmation Status of SNS Subscription
	SubscriptionPendingConfirmation SubscriptionAttributes = "PendingConfirmation"
	// SubscriptionConfirmationWasAuthenticated is Confirmation Authenication Status od SNS Subscription
	SubscriptionConfirmationWasAuthenticated SubscriptionAttributes = "ConfirmationWasAuthenticated"
)

// SubscriptionClient is the external client used for AWS SNSSubscription
type SubscriptionClient interface {
	SubscribeRequest(*sns.SubscribeInput) sns.SubscribeRequest
	UnsubscribeRequest(*sns.UnsubscribeInput) sns.UnsubscribeRequest
	GetSubscriptionAttributesRequest(*sns.GetSubscriptionAttributesInput) sns.GetSubscriptionAttributesRequest
	SetSubscriptionAttributesRequest(*sns.SetSubscriptionAttributesInput) sns.SetSubscriptionAttributesRequest
}

// NewSubscriptionClient returns a new client using AWS credentials as JSON encoded
// data
func NewSubscriptionClient(conf *aws.Config) (SubscriptionClient, error) {
	return sns.New(*conf), nil
}

// GenerateSubscribeInput prepares input for SubscribeRequest
func GenerateSubscribeInput(p *v1alpha1.SNSSubscriptionParameters) *sns.SubscribeInput {
	input := &sns.SubscribeInput{
		Endpoint:              aws.String(p.Endpoint),
		Protocol:              aws.String(p.Protocol),
		TopicArn:              aws.String(p.TopicARN),
		ReturnSubscriptionArn: aws.Bool(true),
	}

	return input
}

// GenerateSubscriptionObservation is used to produce SNSSubscriptionObservation
// from resource at cloud & its attributes
func GenerateSubscriptionObservation(attr map[string]string) v1alpha1.SNSSubscriptionObservation {

	o := v1alpha1.SNSSubscriptionObservation{}
	o.Owner = aws.String(attr[string(SubscriptionOwner)])
	var status v1alpha1.ConfirmationStatus
	if s, err := strconv.ParseBool(attr[string(SubscriptionPendingConfirmation)]); err == nil {
		if s {
			status = v1alpha1.ConfirmationPending
		} else {
			status = v1alpha1.ConfirmationSuccessful
		}
	}
	o.Status = &status

	if s, err := strconv.ParseBool(attr[string(SubscriptionConfirmationWasAuthenticated)]); err == nil {
		o.ConfirmationWasAuthenticated = aws.Bool(s)
	}

	return o
}

// LateInitializeSubscription fills the empty fields in
// *v1alpha1.SNSSubscriptionParameters with the values seen in
// sns.Subscription
func LateInitializeSubscription(in *v1alpha1.SNSSubscriptionParameters, subAttributes map[string]string) {
	in.DeliveryPolicy = awsclients.LateInitializeStringPtr(in.DeliveryPolicy, aws.String(subAttributes[string(SubscriptionDeliveryPolicy)]))
	in.FilterPolicy = awsclients.LateInitializeStringPtr(in.FilterPolicy, aws.String(subAttributes[string(SubscriptionFilterPolicy)]))
	in.RawMessageDelivery = awsclients.LateInitializeStringPtr(in.RawMessageDelivery, aws.String(subAttributes[string(SubscriptionRawMessageDelivery)]))
	in.RedrivePolicy = awsclients.LateInitializeStringPtr(in.RedrivePolicy, aws.String(subAttributes[string(SubscriptionRedrivePolicy)]))
}

// getSubAttributes returns map of SNS Sunscription Attributes
func getSubAttributes(p v1alpha1.SNSSubscriptionParameters) map[string]string {

	attr := make(map[string]string)

	attr[string(SubscriptionDeliveryPolicy)] = aws.StringValue(p.DeliveryPolicy)
	attr[string(SubscriptionFilterPolicy)] = aws.StringValue(p.FilterPolicy)
	attr[string(SubscriptionRawMessageDelivery)] = aws.StringValue(p.RawMessageDelivery)
	attr[string(SubscriptionRedrivePolicy)] = aws.StringValue(p.RedrivePolicy)

	return attr
}

// GetChangedSubAttributes will return the changed attributes  for a subscription
// in provider side
func GetChangedSubAttributes(p v1alpha1.SNSSubscriptionParameters, attrs map[string]string) map[string]string {
	subAttrs := getSubAttributes(p)
	changedAttrs := make(map[string]string)
	for k, v := range subAttrs {
		if v != attrs[k] {
			changedAttrs[k] = v
		}
	}

	return changedAttrs
}

// IsSNSSubscriptionAttributesUpToDate checks if attributes are up to date
func IsSNSSubscriptionAttributesUpToDate(p v1alpha1.SNSSubscriptionParameters, subAttributes map[string]string) bool {
	return aws.StringValue(p.DeliveryPolicy) == subAttributes[string(SubscriptionDeliveryPolicy)] &&
		aws.StringValue(p.FilterPolicy) == subAttributes[string(SubscriptionFilterPolicy)] &&
		aws.StringValue(p.RawMessageDelivery) == subAttributes[string(SubscriptionRawMessageDelivery)] &&
		aws.StringValue(p.RedrivePolicy) == subAttributes[string(SubscriptionRedrivePolicy)]
}
