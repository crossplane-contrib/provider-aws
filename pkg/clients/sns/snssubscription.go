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
	"context"
	"errors"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	snstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"

	"github.com/crossplane/provider-aws/apis/notification/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// SubscriptionAttributes refers to AWS SNS Subscription Attributes List
// ref: https://docs.aws.amazon.com/cli/latest/reference/sns/get-subscription-attributes.html#output
type SubscriptionAttributes string

const (
	// SubscriptionDeliveryPolicy is DeliveryPolicy of SNS Subscription
	SubscriptionDeliveryPolicy = "DeliveryPolicy"
	// SubscriptionFilterPolicy is FilterPolicy of SNS Subscription
	SubscriptionFilterPolicy = "FilterPolicy"
	// SubscriptionRawMessageDelivery is RawMessageDelivery of SNS Subscription
	SubscriptionRawMessageDelivery = "RawMessageDelivery"
	// SubscriptionRedrivePolicy is RedrivePolicy of SNS Subscription
	SubscriptionRedrivePolicy = "RedrivePolicy"
	// SubscriptionOwner is Owner of SNS Subscription
	SubscriptionOwner = "Owner"
	// SubscriptionPendingConfirmation is Confirmation Status of SNS Subscription
	SubscriptionPendingConfirmation = "PendingConfirmation"
	// SubscriptionConfirmationWasAuthenticated is Confirmation Authenication Status od SNS Subscription
	SubscriptionConfirmationWasAuthenticated = "ConfirmationWasAuthenticated"
)

// SubscriptionClient is the external client used for AWS SNSSubscription
type SubscriptionClient interface {
	Subscribe(ctx context.Context, input *sns.SubscribeInput, opts ...func(*sns.Options)) (*sns.SubscribeOutput, error)
	Unsubscribe(ctx context.Context, input *sns.UnsubscribeInput, opts ...func(*sns.Options)) (*sns.UnsubscribeOutput, error)
	GetSubscriptionAttributes(ctx context.Context, input *sns.GetSubscriptionAttributesInput, opts ...func(*sns.Options)) (*sns.GetSubscriptionAttributesOutput, error)
	SetSubscriptionAttributes(ctx context.Context, input *sns.SetSubscriptionAttributesInput, opts ...func(*sns.Options)) (*sns.SetSubscriptionAttributesOutput, error)
}

// NewSubscriptionClient returns a new client using AWS credentials as JSON encoded
// data
func NewSubscriptionClient(cfg aws.Config) SubscriptionClient {
	return sns.NewFromConfig(cfg)
}

// GenerateSubscribeInput prepares input for SubscribeRequest
func GenerateSubscribeInput(p *v1alpha1.SNSSubscriptionParameters) *sns.SubscribeInput {
	input := &sns.SubscribeInput{
		Endpoint:              aws.String(p.Endpoint),
		Protocol:              aws.String(p.Protocol),
		TopicArn:              aws.String(p.TopicARN),
		ReturnSubscriptionArn: true,
	}

	return input
}

// GenerateSubscriptionObservation is used to produce SNSSubscriptionObservation
// from resource at cloud & its attributes
func GenerateSubscriptionObservation(attr map[string]string) v1alpha1.SNSSubscriptionObservation {

	o := v1alpha1.SNSSubscriptionObservation{}
	o.Owner = aws.String(attr[SubscriptionOwner])
	var status v1alpha1.ConfirmationStatus
	if s, err := strconv.ParseBool(attr[SubscriptionPendingConfirmation]); err == nil {
		if s {
			status = v1alpha1.ConfirmationPending
		} else {
			status = v1alpha1.ConfirmationSuccessful
		}
	}
	o.Status = &status

	if s, err := strconv.ParseBool(attr[SubscriptionConfirmationWasAuthenticated]); err == nil {
		o.ConfirmationWasAuthenticated = aws.Bool(s)
	}

	return o
}

// LateInitializeSubscription fills the empty fields in
// *v1alpha1.SNSSubscriptionParameters with the values seen in
// sns.Subscription
func LateInitializeSubscription(in *v1alpha1.SNSSubscriptionParameters, subAttributes map[string]string) {
	in.DeliveryPolicy = awsclients.LateInitializeStringPtr(in.DeliveryPolicy, awsclients.String(subAttributes[SubscriptionDeliveryPolicy]))
	in.FilterPolicy = awsclients.LateInitializeStringPtr(in.FilterPolicy, awsclients.String(subAttributes[SubscriptionFilterPolicy]))
	in.RawMessageDelivery = awsclients.LateInitializeStringPtr(in.RawMessageDelivery, awsclients.String(subAttributes[SubscriptionRawMessageDelivery]))
	in.RedrivePolicy = awsclients.LateInitializeStringPtr(in.RedrivePolicy, awsclients.String(subAttributes[SubscriptionRedrivePolicy]))
}

// getSubAttributes returns map of SNS Sunscription Attributes
func getSubAttributes(p v1alpha1.SNSSubscriptionParameters) map[string]string {
	return map[string]string{
		SubscriptionDeliveryPolicy:     aws.ToString(p.DeliveryPolicy),
		SubscriptionFilterPolicy:       aws.ToString(p.FilterPolicy),
		SubscriptionRawMessageDelivery: aws.ToString(p.RawMessageDelivery),
		SubscriptionRedrivePolicy:      aws.ToString(p.RedrivePolicy),
	}
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
	return aws.ToString(p.DeliveryPolicy) == subAttributes[SubscriptionDeliveryPolicy] &&
		aws.ToString(p.FilterPolicy) == subAttributes[SubscriptionFilterPolicy] &&
		aws.ToString(p.RawMessageDelivery) == subAttributes[SubscriptionRawMessageDelivery] &&
		aws.ToString(p.RedrivePolicy) == subAttributes[SubscriptionRedrivePolicy]
}

// IsSubscriptionNotFound returns true if the error code indicates that the item was not found
func IsSubscriptionNotFound(err error) bool {
	var nfe *snstypes.NotFoundException
	var rnfe *snstypes.ResourceNotFoundException
	if errors.As(err, &nfe) || errors.As(err, &rnfe) {
		return true
	}
	return false
}
