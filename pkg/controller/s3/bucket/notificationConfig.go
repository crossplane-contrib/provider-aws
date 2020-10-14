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

package bucket

import (
	"context"

	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/s3"
)

const (
	notificationGetFailed = "cannot get Bucket notification"
	notificationPutFailed = "cannot put Bucket notification"
)

// NotificationConfigurationClient is the client for API methods and reconciling the LifecycleConfiguration
type NotificationConfigurationClient struct {
	client s3.BucketClient
}

// LateInitialize is responsible for initializing the resource based on the external value
func (in *NotificationConfigurationClient) LateInitialize(ctx context.Context, bucket *v1beta1.Bucket) error {
	external, err := in.client.GetBucketNotificationConfigurationRequest(&awss3.GetBucketNotificationConfigurationInput{Bucket: aws.String(meta.GetExternalName(bucket))}).Send(ctx)
	if err != nil {
		return errors.Wrap(err, notificationGetFailed)
	}
	if emptyConfiguration(external) {
		// There is nothing to initialize from AWS
		return nil
	}
	config := bucket.Spec.ForProvider.NotificationConfiguration
	if config == nil {
		// We need the configuration to exist so we can initialize
		bucket.Spec.ForProvider.NotificationConfiguration = &v1beta1.NotificationConfiguration{}
		config = bucket.Spec.ForProvider.NotificationConfiguration
	}

	// A list is provided by AWS
	if external.LambdaFunctionConfigurations != nil {
		if config.LambdaFunctionConfigurations == nil {
			config.LambdaFunctionConfigurations = make([]v1beta1.LambdaFunctionConfiguration, len(external.LambdaFunctionConfigurations))
		}
		LateInitializeLambda(external.LambdaFunctionConfigurations, config.LambdaFunctionConfigurations)
	}

	// A list is provided by AWS
	if external.QueueConfigurations != nil {
		if config.QueueConfigurations == nil {
			config.QueueConfigurations = make([]v1beta1.QueueConfiguration, len(external.QueueConfigurations))
		}
		LateInitializeQueue(external.QueueConfigurations, config.QueueConfigurations)
	}

	// A list is provided by AWS
	if external.TopicConfigurations != nil {
		if config.TopicConfigurations == nil {
			config.TopicConfigurations = make([]v1beta1.TopicConfiguration, len(external.TopicConfigurations))
		}
		LateInitializeTopic(external.TopicConfigurations, config.TopicConfigurations)
	}
	return nil
}

// LateInitializeFilter initializes the external awss3.NotificationConfigurationFilter to a local v1beta.NotificationConfigurationFilter
func LateInitializeFilter(local *v1beta1.NotificationConfigurationFilter, external *awss3.NotificationConfigurationFilter) *v1beta1.NotificationConfigurationFilter {
	if local != nil {
		return local
	}
	if external == nil {
		return nil
	}
	local = &v1beta1.NotificationConfigurationFilter{}
	if external.Key == nil {
		return local
	}
	local.Key = &v1beta1.S3KeyFilter{}
	if external.Key.FilterRules != nil {
		local.Key.FilterRules = make([]v1beta1.FilterRule, len(external.Key.FilterRules))
		for i, v := range external.Key.FilterRules {
			local.Key.FilterRules[i] = v1beta1.FilterRule{
				Name:  string(v.Name),
				Value: v.Value,
			}
		}
	}
	return local
}

// LateInitializeEvents initializes the external []awss3.Event to a local []string
func LateInitializeEvents(local []string, external []awss3.Event) []string {
	if local != nil {
		return local
	}
	newLocal := make([]string, len(external))
	for i, v := range external {
		newLocal[i] = string(v)
	}
	return newLocal
}

// LateInitializeLambda initializes the external awss3.LambdaFunctionConfiguration to a local v1beta.LambdaFunctionConfiguration
func LateInitializeLambda(external []awss3.LambdaFunctionConfiguration, local []v1beta1.LambdaFunctionConfiguration) {
	for i, v := range external {
		if i >= len(local) {
			break
		}
		local[i] = v1beta1.LambdaFunctionConfiguration{
			Events:            LateInitializeEvents(local[i].Events, v.Events),
			Filter:            LateInitializeFilter(local[i].Filter, v.Filter),
			ID:                aws.LateInitializeStringPtr(local[i].ID, v.Id),
			LambdaFunctionArn: aws.LateInitializeString(local[i].LambdaFunctionArn, v.LambdaFunctionArn),
		}
	}
}

// LateInitializeQueue initializes the external awss3.QueueConfiguration to a local v1beta.QueueConfiguration
func LateInitializeQueue(external []awss3.QueueConfiguration, local []v1beta1.QueueConfiguration) {
	for i, v := range external {
		if i >= len(local) {
			break
		}
		local[i] = v1beta1.QueueConfiguration{
			Events:   LateInitializeEvents(local[i].Events, v.Events),
			Filter:   LateInitializeFilter(local[i].Filter, v.Filter),
			ID:       aws.LateInitializeStringPtr(local[i].ID, v.Id),
			QueueArn: aws.LateInitializeString(local[i].QueueArn, v.QueueArn),
		}
	}
}

// LateInitializeTopic initializes the external awss3.TopicConfiguration to a local v1beta.TopicConfiguration
func LateInitializeTopic(external []awss3.TopicConfiguration, local []v1beta1.TopicConfiguration) {
	for i, v := range external {
		if i >= len(local) {
			break
		}
		local[i] = v1beta1.TopicConfiguration{
			Events:   LateInitializeEvents(local[i].Events, v.Events),
			Filter:   LateInitializeFilter(local[i].Filter, v.Filter),
			ID:       aws.LateInitializeStringPtr(local[i].ID, v.Id),
			TopicArn: aws.LateInitializeStringPtr(local[i].TopicArn, v.TopicArn),
		}
	}
}

// NewNotificationConfigurationClient creates the client for Accelerate Configuration
func NewNotificationConfigurationClient(client s3.BucketClient) *NotificationConfigurationClient {
	return &NotificationConfigurationClient{client: client}
}

func emptyConfiguration(external *awss3.GetBucketNotificationConfigurationResponse) bool {
	return external == nil || len(external.TopicConfigurations) == 0 || len(external.QueueConfigurations) == 0 || len(external.LambdaFunctionConfigurations) == 0
}

func bucketStatus(config *v1beta1.NotificationConfiguration, external *awss3.GetBucketNotificationConfigurationResponse) ResourceStatus { // nolint:gocyclo
	if config == nil && len(external.QueueConfigurations) == 0 && len(external.LambdaFunctionConfigurations) == 0 && len(external.TopicConfigurations) == 0 {
		return Updated
	} else if config == nil && (len(external.QueueConfigurations) != 0 || len(external.LambdaFunctionConfigurations) != 0 || len(external.TopicConfigurations) != 0) {
		return NeedsDeletion
	}
	return NeedsUpdate
}

// Observe checks if the resource exists and if it matches the local configuration
func (in *NotificationConfigurationClient) Observe(ctx context.Context, bucket *v1beta1.Bucket) (ResourceStatus, error) {
	external, err := in.client.GetBucketNotificationConfigurationRequest(&awss3.GetBucketNotificationConfigurationInput{Bucket: aws.String(meta.GetExternalName(bucket))}).Send(ctx)
	if err != nil {
		return NeedsUpdate, errors.Wrap(err, notificationGetFailed)
	}

	config := bucket.Spec.ForProvider.NotificationConfiguration
	status := bucketStatus(config, external)
	switch status { // nolint:exhaustive
	case Updated, NeedsDeletion:
		return status, nil
	}

	generated := GenerateConfiguration(config)

	if cmp.Equal(external.LambdaFunctionConfigurations, generated.LambdaFunctionConfigurations) &&
		cmp.Equal(external.QueueConfigurations, generated.QueueConfigurations) &&
		cmp.Equal(external.TopicConfigurations, generated.TopicConfigurations) {
		return Updated, nil
	}

	return NeedsUpdate, nil
}

func copyEvents(src []string) []awss3.Event {
	if len(src) == 0 {
		return nil
	}
	out := make([]awss3.Event, len(src))
	for i, v := range src {
		cast := awss3.Event(v)
		out[i] = cast
	}
	return out
}

func generateFilter(src *v1beta1.NotificationConfigurationFilter) *awss3.NotificationConfigurationFilter {
	if src == nil || src.Key == nil {
		return nil
	}
	out := &awss3.NotificationConfigurationFilter{Key: &awss3.S3KeyFilter{}}
	if src.Key.FilterRules == nil {
		return out
	}
	out.Key.FilterRules = make([]awss3.FilterRule, len(src.Key.FilterRules))
	for i, v := range src.Key.FilterRules {
		out.Key.FilterRules[i] = awss3.FilterRule{
			Name:  awss3.FilterRuleName(v.Name),
			Value: v.Value,
		}
	}
	return out
}

// GenerateLambdaConfiguration creates []awss3.LambdaFunctionConfiguration from the local NotificationConfiguration
func GenerateLambdaConfiguration(config *v1beta1.NotificationConfiguration) []awss3.LambdaFunctionConfiguration {
	if config.LambdaFunctionConfigurations == nil {
		return nil
	}
	configurations := make([]awss3.LambdaFunctionConfiguration, len(config.LambdaFunctionConfigurations))
	for i, v := range config.LambdaFunctionConfigurations {
		conf := awss3.LambdaFunctionConfiguration{
			Filter:            nil,
			Id:                v.ID,
			LambdaFunctionArn: aws.String(v.LambdaFunctionArn),
		}
		if v.Events != nil {
			conf.Events = copyEvents(v.Events)
		}
		if v.Filter != nil {
			conf.Filter = generateFilter(v.Filter)
		}
		configurations[i] = conf
	}
	return configurations
}

// GenerateTopicConfigurations creates []awss3.TopicConfiguration from the local NotificationConfiguration
func GenerateTopicConfigurations(config *v1beta1.NotificationConfiguration) []awss3.TopicConfiguration {
	if config.TopicConfigurations == nil {
		return nil
	}
	configurations := make([]awss3.TopicConfiguration, len(config.TopicConfigurations))
	for i, v := range config.TopicConfigurations {
		conf := awss3.TopicConfiguration{
			Id:       v.ID,
			TopicArn: v.TopicArn,
		}
		if v.Events != nil {
			conf.Events = copyEvents(v.Events)
		}
		if v.Filter != nil {
			conf.Filter = generateFilter(v.Filter)
		}
		configurations[i] = conf
	}
	return configurations
}

// GenerateQueueConfigurations creates []awss3.QueueConfiguration from the local NotificationConfiguration
func GenerateQueueConfigurations(config *v1beta1.NotificationConfiguration) []awss3.QueueConfiguration {
	if config.QueueConfigurations == nil {
		return make([]awss3.QueueConfiguration, 0)
	}
	configurations := make([]awss3.QueueConfiguration, len(config.QueueConfigurations))
	for i, v := range config.QueueConfigurations {
		conf := awss3.QueueConfiguration{
			Filter:   nil,
			Id:       v.ID,
			QueueArn: aws.String(v.QueueArn),
		}
		if v.Events != nil {
			conf.Events = copyEvents(v.Events)
		}
		if v.Filter != nil {
			conf.Filter = generateFilter(v.Filter)
		}
		configurations[i] = conf
	}
	return configurations
}

// GenerateConfiguration creates the external aws NotificationConfiguration from the local representation
func GenerateConfiguration(config *v1beta1.NotificationConfiguration) *awss3.NotificationConfiguration {
	awsConfig := &awss3.NotificationConfiguration{}
	lambda := GenerateLambdaConfiguration(config)
	if len(lambda) != 0 {
		awsConfig.LambdaFunctionConfigurations = lambda
	}
	queue := GenerateQueueConfigurations(config)
	if len(queue) != 0 {
		awsConfig.QueueConfigurations = queue
	}
	topic := GenerateTopicConfigurations(config)
	if len(topic) != 0 {
		awsConfig.TopicConfigurations = topic
	}
	return awsConfig
}

// GenerateNotificationConfigurationInput creates the input for the LifecycleConfiguration request for the S3 Client
func GenerateNotificationConfigurationInput(name string, config *v1beta1.NotificationConfiguration) *awss3.PutBucketNotificationConfigurationInput {
	awsConfig := GenerateConfiguration(config)
	return &awss3.PutBucketNotificationConfigurationInput{
		Bucket:                    aws.String(name),
		NotificationConfiguration: awsConfig,
	}
}

// CreateOrUpdate sends a request to have resource created on AWS
func (in *NotificationConfigurationClient) CreateOrUpdate(ctx context.Context, bucket *v1beta1.Bucket) error {
	if bucket.Spec.ForProvider.NotificationConfiguration == nil {
		return nil
	}
	input := GenerateNotificationConfigurationInput(meta.GetExternalName(bucket), bucket.Spec.ForProvider.NotificationConfiguration)
	_, err := in.client.PutBucketNotificationConfigurationRequest(input).Send(ctx)
	return errors.Wrap(err, notificationPutFailed)
}

// Delete does nothing because there is no corresponding deletion call in AWS.
func (*NotificationConfigurationClient) Delete(_ context.Context, _ *v1beta1.Bucket) error {
	return nil
}
