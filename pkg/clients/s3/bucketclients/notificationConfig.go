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

package bucketclients

import (
	"context"

	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/s3"
)

// NotificationConfigurationClient is the client for API methods and reconciling the LifecycleConfiguration
type NotificationConfigurationClient struct {
	config *v1beta1.NotificationConfiguration
	client s3.BucketClient
}

// NewNotificationConfigurationClient creates the client for Accelerate Configuration
func NewNotificationConfigurationClient(bucket *v1beta1.Bucket, client s3.BucketClient) *NotificationConfigurationClient {
	return &NotificationConfigurationClient{config: bucket.Spec.Parameters.NotificationConfiguration, client: client}
}

func notExistsUpdated(config *v1beta1.NotificationConfiguration, external *awss3.GetBucketNotificationConfigurationResponse) bool {
	return (config == nil || len(config.TopicConfigurations) == 0 || len(config.QueueConfigurations) == 0 || len(config.LambdaFunctionConfigurations) == 0) &&
		len(external.QueueConfigurations) == 0 && len(external.LambdaFunctionConfigurations) == 0 && len(external.TopicConfigurations) == 0
}

func bucketStatus(config *v1beta1.NotificationConfiguration, external *awss3.GetBucketNotificationConfigurationResponse) ResourceStatus {
	if notExistsUpdated(config, external) {
		return Updated
	} else if config == nil && (len(external.QueueConfigurations) != 0 || len(external.LambdaFunctionConfigurations) != 0 || len(external.TopicConfigurations) != 0) {
		return NeedsDeletion
	}
	return NeedsUpdate
}

// Observe checks if the resource exists and if it matches the local configuration
func (in *NotificationConfigurationClient) Observe(ctx context.Context, bucket *v1beta1.Bucket) (ResourceStatus, error) {
	conf, err := in.client.GetBucketNotificationConfigurationRequest(&awss3.GetBucketNotificationConfigurationInput{Bucket: aws.String(meta.GetExternalName(bucket))}).Send(ctx)
	if err != nil {
		return NeedsUpdate, errors.Wrap(err, "cannot get bucket notification")
	}

	status := bucketStatus(in.config, conf)
	switch status {
	case Updated, NeedsDeletion:
		return status, nil
	}

	generated := in.generateConfiguration()
	if err != nil {
		return NeedsUpdate, errors.Wrap(err, "unable to create rules for bucket notification reconcile")
	}

	if cmp.Equal(conf.LambdaFunctionConfigurations, generated.LambdaFunctionConfigurations) &&
		cmp.Equal(conf.QueueConfigurations, generated.QueueConfigurations) &&
		cmp.Equal(conf.TopicConfigurations, generated.TopicConfigurations) {
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

func (in *NotificationConfigurationClient) generateLambdaConfiguration() []awss3.LambdaFunctionConfiguration {
	if in.config.LambdaFunctionConfigurations == nil {
		return make([]awss3.LambdaFunctionConfiguration, 0)
	}
	configurations := make([]awss3.LambdaFunctionConfiguration, len(in.config.LambdaFunctionConfigurations))
	for i, v := range in.config.LambdaFunctionConfigurations {
		conf := awss3.LambdaFunctionConfiguration{
			Filter:            nil,
			Id:                v.ID,
			LambdaFunctionArn: v.LambdaFunctionArn,
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

func (in *NotificationConfigurationClient) generateTopicConfigurations() []awss3.TopicConfiguration {
	if in.config.TopicConfigurations == nil {
		return make([]awss3.TopicConfiguration, 0)
	}
	configurations := make([]awss3.TopicConfiguration, len(in.config.TopicConfigurations))
	for i, v := range in.config.TopicConfigurations {
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

func (in *NotificationConfigurationClient) generateQueueConfigurations() []awss3.QueueConfiguration {
	if in.config.QueueConfigurations == nil {
		return make([]awss3.QueueConfiguration, 0)
	}
	configurations := make([]awss3.QueueConfiguration, len(in.config.QueueConfigurations))
	for i, v := range in.config.QueueConfigurations {
		conf := awss3.QueueConfiguration{
			Filter:   nil,
			Id:       v.ID,
			QueueArn: v.QueueArn,
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

func (in *NotificationConfigurationClient) generateConfiguration() *awss3.NotificationConfiguration {
	conf := &awss3.NotificationConfiguration{
		LambdaFunctionConfigurations: in.generateLambdaConfiguration(),
		QueueConfigurations:          in.generateQueueConfigurations(),
		TopicConfigurations:          in.generateTopicConfigurations(),
	}
	return conf
}

// GenerateNotificationConfigurationInput creates the input for the LifecycleConfiguration request for the S3 Client
func (in *NotificationConfigurationClient) GenerateNotificationConfigurationInput(name string) (*awss3.PutBucketNotificationConfigurationInput, error) {
	conf := in.generateConfiguration()
	return &awss3.PutBucketNotificationConfigurationInput{
		Bucket:                    aws.String(name),
		NotificationConfiguration: conf,
	}, nil
}

// Create sends a request to have resource created on AWS
func (in *NotificationConfigurationClient) Create(ctx context.Context, bucket *v1beta1.Bucket) (managed.ExternalUpdate, error) {
	if in.config == nil {
		return managed.ExternalUpdate{}, nil
	}
	input, err := in.GenerateNotificationConfigurationInput(meta.GetExternalName(bucket))
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "unable to create input for bucket notification request")
	}
	_, err = in.client.PutBucketNotificationConfigurationRequest(input).Send(ctx)
	return managed.ExternalUpdate{}, errors.Wrap(err, "cannot put bucket notification")
}

// Delete creates the request to delete the resource on AWS or set it to the default value.
func (in *NotificationConfigurationClient) Delete(ctx context.Context, bucket *v1beta1.Bucket) error {
	input := &awss3.PutBucketNotificationConfigurationInput{
		Bucket:                    aws.String(meta.GetExternalName(bucket)),
		NotificationConfiguration: &awss3.NotificationConfiguration{},
	}
	_, err := in.client.PutBucketNotificationConfigurationRequest(input).Send(ctx)
	return errors.Wrap(err, "cannot delete bucket notification")
}
