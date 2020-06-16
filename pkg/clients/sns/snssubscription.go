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
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/notification/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

const (
	//SNSSubscriptionNotFound is the error code that is returned if SNS Subscription is not present
	SNSSubscriptionNotFound = "InvalidSNSSubscription.NotFound"

	// SNSSubscriptionInPendingConfirmation will be raise when SNS Subscription is found
	// but in "pending confirmation" state.
	SNSSubscriptionInPendingConfirmation = "InvalidSNSSubscription.PendingConfirmation"
)

// IsSubscriptionNotFound returns true if the error code indicates that the item was not found
func IsSubscriptionNotFound(err error) bool {
	if _, ok := err.(*SubscriptionNotFound); ok {
		return true
	}
	return false
}

// SubscriptionNotFound will be raised when there is no SNSTopic
type SubscriptionNotFound struct{}

func (err *SubscriptionNotFound) Error() string {
	return fmt.Sprint(SNSSubscriptionNotFound)
}

// IsSubscriptionConfirmationPending returns true if the error code indicates that the item was not found
func IsSubscriptionConfirmationPending(err error) bool {
	if _, ok := err.(*SubscriptionInPendingConfirmation); ok {
		return true
	}
	return false
}

// SubscriptionInPendingConfirmation will be raised when subscription is in
// "pending confirmation" state.
type SubscriptionInPendingConfirmation struct{}

func (err *SubscriptionInPendingConfirmation) Error() string {
	return fmt.Sprint(SNSSubscriptionInPendingConfirmation)
}

// SubscriptionClient is the external
type SubscriptionClient interface {
	ListSubscriptionsByTopicRequest(*sns.ListSubscriptionsByTopicInput) sns.ListSubscriptionsByTopicRequest
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
		TopicArn:              p.TopicARN,
		ReturnSubscriptionArn: aws.Bool(true),
	}

	return input
}

// GenerateSubscriptionObservation is used to produce SNSSubscriptionObservation
// from resource at cloud & its attributes
func GenerateSubscriptionObservation(attr map[string]string) v1alpha1.SNSSubscriptionObservation {

	o := v1alpha1.SNSSubscriptionObservation{}
	o.Owner = aws.String(attr["Owner"])
	var status v1alpha1.ConfirmationStatus
	if s, err := strconv.ParseBool(attr["PendingConfirmation"]); err == nil {
		if s {
			status = v1alpha1.ConfirmationPending
		} else {
			status = v1alpha1.ConfirmationSuccessful
		}
	}
	o.Status = &status

	if s, err := strconv.ParseBool(attr["ConfirmationWasAuthenticated"]); err == nil {
		o.ConfirmationWasAuthenticated = aws.Bool(s)
	}

	return o
}

// LateInitializeSubscription fills the empty fields in
// *v1alpha1.SNSSubscriptionParameters with the values seen in
// sns.Subscription
func LateInitializeSubscription(in *v1alpha1.SNSSubscriptionParameters, subAttributes map[string]string) {
	in.DeliveryPolicy = awsclients.LateInitializeStringPtr(in.DeliveryPolicy, aws.String(subAttributes["DeliveryPolicy"]))
	in.FilterPolicy = awsclients.LateInitializeStringPtr(in.FilterPolicy, aws.String(subAttributes["FilterPolicy"]))
	in.RawMessageDelivery = awsclients.LateInitializeStringPtr(in.RawMessageDelivery, aws.String(subAttributes["RawMessageDelivery"]))
	in.RedrivePolicy = awsclients.LateInitializeStringPtr(in.RedrivePolicy, aws.String(subAttributes["RedrivePolicy"]))
}

// GetSNSSubscription returns SNSSubscription from List of SNSSubscription
func GetSNSSubscription(res *sns.ListSubscriptionsByTopicResponse, cr *v1alpha1.SNSSubscription) (sns.Subscription, error) {

	p := cr.Spec.ForProvider
	for _, sub := range res.Subscriptions {
		if cmp.Equal(*sub.TopicArn, cr.Spec.ForProvider.TopicARN) && cmp.Equal(sub.Endpoint, p.Endpoint) && cmp.Equal(sub.Protocol, p.Protocol) {
			return sub, nil
		}
	}

	return sns.Subscription{}, &SubscriptionNotFound{}

}

// getSubAttributes returns map of SNS Sunscription Attributes
func getSubAttributes(p v1alpha1.SNSSubscriptionParameters) map[string]string {

	attr := make(map[string]string)

	attr["DeliveryPolicy"] = aws.StringValue(p.DeliveryPolicy)
	attr["FilterPolicy"] = aws.StringValue(p.FilterPolicy)
	attr["RawMessageDelivery"] = aws.StringValue(p.RawMessageDelivery)
	attr["RedrivePolicy"] = aws.StringValue(p.RedrivePolicy)

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

// IsSNSSubscriptionUpToDate checks if object is up to date
func IsSNSSubscriptionUpToDate(p v1alpha1.SNSSubscriptionParameters, sub *sns.Subscription, subAttributes map[string]string) bool {
	return p.Endpoint == aws.StringValue(sub.Endpoint) && p.Protocol == aws.StringValue(sub.Protocol) && isSNSSubscriptionAttributesUpToDate(p, subAttributes)
}

// isSNSSubscriptionAttributesUpToDate checks if attributes are up to date
func isSNSSubscriptionAttributesUpToDate(p v1alpha1.SNSSubscriptionParameters, subAttributes map[string]string) bool {
	return *p.DeliveryPolicy == subAttributes["DeliveryPolicy"] &&
		*p.FilterPolicy == subAttributes["FilterPolicy"] &&
		*p.RawMessageDelivery == subAttributes["RawMessageDelivery"] &&
		*p.RedrivePolicy == subAttributes["RedrivePolicy"] &&
		*p.TopicARN == subAttributes["TopicArn"]
}
