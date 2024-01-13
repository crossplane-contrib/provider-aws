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
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go/document"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"k8s.io/utils/ptr"

	"github.com/crossplane-contrib/provider-aws/apis/s3/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/s3"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/s3/bucket/convert"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

const (
	notificationGetFailed = "cannot get Bucket notification"
	notificationPutFailed = "cannot put Bucket notification (may be caused by insuffifienct permissions on the target)"
)

// NotificationConfigurationClient is the client for API methods and reconciling the LifecycleConfiguration
type NotificationConfigurationClient struct {
	client s3.BucketClient
}

// NewNotificationConfigurationClient creates the client for Accelerate Configuration
func NewNotificationConfigurationClient(client s3.BucketClient) *NotificationConfigurationClient {
	return &NotificationConfigurationClient{client: client}
}

// Observe checks if the resource exists and if it matches the local configuration
func (in *NotificationConfigurationClient) Observe(ctx context.Context, bucket *v1beta1.Bucket) (ResourceStatus, error) {
	external, err := in.client.GetBucketNotificationConfiguration(ctx, &awss3.GetBucketNotificationConfigurationInput{Bucket: pointer.ToOrNilIfZeroValue(meta.GetExternalName(bucket))})
	if err != nil {
		return NeedsUpdate, errorutils.Wrap(err, notificationGetFailed)
	}

	return IsNotificationConfigurationUpToDate(bucket.Spec.ForProvider.NotificationConfiguration, external)
}

// IsNotificationConfigurationUpToDate determines whether a notification configuration needs to be updated
func IsNotificationConfigurationUpToDate(cr *v1beta1.NotificationConfiguration, external *awss3.GetBucketNotificationConfigurationOutput) (ResourceStatus, error) { //nolint:gocyclo
	// Note - aws API treats nil configuration different than empty configuration
	// As such, we can't prealloc this due to the API
	// If no configuration is defined but there is one in aws, we must delete it
	if cr == nil && (len(external.QueueConfigurations) != 0 || len(external.LambdaFunctionConfigurations) != 0 || len(external.TopicConfigurations) != 0) {
		return NeedsDeletion, nil
	}
	// We can't prealloc this for the API but we can to make comparison easier
	if cr == nil {
		cr = &v1beta1.NotificationConfiguration{
			LambdaFunctionConfigurations: []v1beta1.LambdaFunctionConfiguration{},
			QueueConfigurations:          []v1beta1.QueueConfiguration{},
			TopicConfigurations:          []v1beta1.TopicConfiguration{},
		}
	}
	// If any of the lengths in aws are different then there is something to delete
	if len(cr.LambdaFunctionConfigurations) < len(external.LambdaFunctionConfigurations) || len(cr.QueueConfigurations) < len(external.QueueConfigurations) || len(cr.TopicConfigurations) < len(external.TopicConfigurations) {
		return NeedsDeletion, nil
	}

	// Convert to a comparable object
	generated := GenerateConfiguration(cr)

	// Sort everything
	sortLambda(generated.LambdaFunctionConfigurations)
	sortLambda(external.LambdaFunctionConfigurations)

	sortQueue(generated.QueueConfigurations)
	sortQueue(external.QueueConfigurations)

	sortTopic(generated.TopicConfigurations)
	sortTopic(external.TopicConfigurations)

	// The AWS API returns QueueConfiguration.Filter.Key.FilterRules.Name as "Prefix"/"Suffix" but expects
	// "prefix"/"suffix" this leads to inconsistency and a constant diff. Fixes
	// https://github.com/crossplane-contrib/provider-aws/issues/1165
	external.QueueConfigurations = sanitizedQueueConfigurations(external.QueueConfigurations)

	if cmp.Equal(external.LambdaFunctionConfigurations, generated.LambdaFunctionConfigurations, cmpopts.IgnoreTypes(document.NoSerde{}, types.LambdaFunctionConfiguration{}.Id), cmpopts.EquateEmpty()) &&
		cmp.Equal(external.QueueConfigurations, generated.QueueConfigurations, cmpopts.IgnoreTypes(document.NoSerde{}, types.QueueConfiguration{}.Id), cmpopts.EquateEmpty()) &&
		cmp.Equal(external.TopicConfigurations, generated.TopicConfigurations, cmpopts.IgnoreTypes(document.NoSerde{}, types.TopicConfiguration{}.Id), cmpopts.EquateEmpty()) {
		return Updated, nil
	}

	return NeedsUpdate, nil
}

func sortLambda(configs []types.LambdaFunctionConfiguration) {
	sort.Slice(configs, func(i, j int) bool {
		if a, b := configs[i].LambdaFunctionArn, configs[j].LambdaFunctionArn; a != b {
			return aws.ToString(a) < aws.ToString(b)
		}
		return true
	})
}

func sortQueue(configs []types.QueueConfiguration) {
	sort.Slice(configs, func(i, j int) bool {
		if a, b := configs[i].QueueArn, configs[j].QueueArn; a != b {
			return aws.ToString(a) < aws.ToString(b)
		}
		return true
	})
}

func sortTopic(configs []types.TopicConfiguration) {
	sort.Slice(configs, func(i, j int) bool {
		if a, b := configs[i].TopicArn, configs[j].TopicArn; a != b {
			return aws.ToString(a) < aws.ToString(b)
		}
		return true
	})
}

func sanitizedQueueConfigurations(configs []types.QueueConfiguration) []types.QueueConfiguration {
	if configs == nil {
		return []types.QueueConfiguration{}
	}

	sConfig := (&convert.ConverterImpl{}).DeepCopyAWSQueueConfiguration(configs)
	for c := range sConfig {
		if sConfig[c].Events == nil {
			sConfig[c].Events = []types.Event{}
		}
		if sConfig[c].Filter == nil {
			continue
		}
		if sConfig[c].Filter.Key == nil {
			continue
		}
		if sConfig[c].Filter.Key.FilterRules == nil {
			sConfig[c].Filter.Key.FilterRules = []types.FilterRule{}
		}
		for r := range sConfig[c].Filter.Key.FilterRules {
			name := string(sConfig[c].Filter.Key.FilterRules[r].Name)
			sConfig[c].Filter.Key.FilterRules[r].Name = types.FilterRuleName(strings.ToLower(name))
		}
	}

	return sConfig
}

// GenerateLambdaConfiguration creates []awss3.LambdaFunctionConfiguration from the local NotificationConfiguration
func GenerateLambdaConfiguration(config *v1beta1.NotificationConfiguration) []types.LambdaFunctionConfiguration {
	// NOTE(muvaf): We skip prealloc because the behavior of AWS SDK differs when
	// the array is 0 element vs nil.
	var configurations []types.LambdaFunctionConfiguration //nolint:prealloc
	for _, v := range config.LambdaFunctionConfigurations {
		conf := types.LambdaFunctionConfiguration{
			Filter:            nil,
			Id:                v.ID,
			LambdaFunctionArn: pointer.ToOrNilIfZeroValue(v.LambdaFunctionArn),
		}
		if v.Events != nil {
			conf.Events = copyEvents(v.Events)
		}
		if v.Filter != nil {
			conf.Filter = generateFilter(v.Filter)
		}
		configurations = append(configurations, conf)
	}
	return configurations
}

// GenerateTopicConfigurations creates []awss3.TopicConfiguration from the local NotificationConfiguration
func GenerateTopicConfigurations(config *v1beta1.NotificationConfiguration) []types.TopicConfiguration {
	// NOTE(muvaf): We skip prealloc because the behavior of AWS SDK differs when
	// the array is 0 element vs nil.
	var configurations []types.TopicConfiguration //nolint:prealloc
	for _, v := range config.TopicConfigurations {
		conf := types.TopicConfiguration{
			Id:       v.ID,
			TopicArn: v.TopicArn,
		}
		if v.Events != nil {
			conf.Events = copyEvents(v.Events)
		}
		if v.Filter != nil {
			conf.Filter = generateFilter(v.Filter)
		}
		configurations = append(configurations, conf)
	}
	return configurations
}

// GenerateQueueConfigurations creates []awss3.QueueConfiguration from the local NotificationConfiguration
func GenerateQueueConfigurations(config *v1beta1.NotificationConfiguration) []types.QueueConfiguration {
	// NOTE(muvaf): We skip prealloc because the behavior of AWS SDK differs when
	// the array is 0 element vs nil.
	var configurations []types.QueueConfiguration //nolint:prealloc
	for _, v := range config.QueueConfigurations {
		conf := types.QueueConfiguration{
			Id:       v.ID,
			QueueArn: v.QueueArn,
		}
		if v.Events != nil {
			conf.Events = copyEvents(v.Events)
		}
		if v.Filter != nil {
			conf.Filter = generateFilter(v.Filter)
		}
		configurations = append(configurations, conf)
	}
	return configurations
}

func copyEvents(src []string) []types.Event {
	if len(src) == 0 {
		return nil
	}
	out := make([]types.Event, len(src))
	for i, v := range src {
		cast := types.Event(v)
		out[i] = cast
	}
	return out
}

func generateFilter(src *v1beta1.NotificationConfigurationFilter) *types.NotificationConfigurationFilter {
	if src == nil || src.Key == nil {
		return nil
	}
	out := &types.NotificationConfigurationFilter{Key: &types.S3KeyFilter{}}
	if src.Key.FilterRules == nil {
		return out
	}
	out.Key.FilterRules = make([]types.FilterRule, len(src.Key.FilterRules))
	for i, v := range src.Key.FilterRules {
		out.Key.FilterRules[i] = types.FilterRule{
			Name:  types.FilterRuleName(v.Name),
			Value: v.Value,
		}
	}
	return out
}

// GenerateConfiguration creates the external aws NotificationConfiguration from the local representation
func GenerateConfiguration(config *v1beta1.NotificationConfiguration) *types.NotificationConfiguration {
	return &types.NotificationConfiguration{
		LambdaFunctionConfigurations: GenerateLambdaConfiguration(config),
		QueueConfigurations:          GenerateQueueConfigurations(config),
		TopicConfigurations:          GenerateTopicConfigurations(config),
	}
}

// GenerateNotificationConfigurationInput creates the input for the LifecycleConfiguration request for the S3 Client
func GenerateNotificationConfigurationInput(name string, config *v1beta1.NotificationConfiguration) *awss3.PutBucketNotificationConfigurationInput {
	return &awss3.PutBucketNotificationConfigurationInput{
		Bucket:                    pointer.ToOrNilIfZeroValue(name),
		NotificationConfiguration: GenerateConfiguration(config),
	}
}

// CreateOrUpdate sends a request to have resource created on AWS
func (in *NotificationConfigurationClient) CreateOrUpdate(ctx context.Context, bucket *v1beta1.Bucket) error {
	if bucket.Spec.ForProvider.NotificationConfiguration == nil {
		return nil
	}
	input := GenerateNotificationConfigurationInput(meta.GetExternalName(bucket), bucket.Spec.ForProvider.NotificationConfiguration)
	_, err := in.client.PutBucketNotificationConfiguration(ctx, input)
	return errorutils.Wrap(err, notificationPutFailed)
}

// Delete resets the buckets notification configuration to empty.
func (in *NotificationConfigurationClient) Delete(ctx context.Context, bucket *v1beta1.Bucket) error {
	_, err := in.client.PutBucketNotificationConfiguration(ctx, &awss3.PutBucketNotificationConfigurationInput{
		Bucket: ptr.To(meta.GetExternalName(bucket)),
		NotificationConfiguration: &types.NotificationConfiguration{
			EventBridgeConfiguration:     &types.EventBridgeConfiguration{},
			LambdaFunctionConfigurations: []types.LambdaFunctionConfiguration{},
			QueueConfigurations:          []types.QueueConfiguration{},
			TopicConfigurations:          []types.TopicConfiguration{},
		},
	})
	return errorutils.Wrap(err, notificationPutFailed)
}

// LateInitialize is responsible for initializing the resource based on the external value
func (in *NotificationConfigurationClient) LateInitialize(ctx context.Context, bucket *v1beta1.Bucket) error {
	external, err := in.client.GetBucketNotificationConfiguration(ctx, &awss3.GetBucketNotificationConfigurationInput{Bucket: pointer.ToOrNilIfZeroValue(meta.GetExternalName(bucket))})
	if err != nil {
		return errorutils.Wrap(err, notificationGetFailed)
	}
	if emptyConfiguration(external) {
		// There is nothing to initialize from AWS
		return nil
	}

	if bucket.Spec.ForProvider.NotificationConfiguration == nil {
		// We need the configuration to exist so we can initialize
		bucket.Spec.ForProvider.NotificationConfiguration = &v1beta1.NotificationConfiguration{}
	}
	config := bucket.Spec.ForProvider.NotificationConfiguration

	// A list is provided by AWS
	if len(external.LambdaFunctionConfigurations) != 0 {
		config.LambdaFunctionConfigurations = LateInitializeLambda(external.LambdaFunctionConfigurations, config.LambdaFunctionConfigurations)
	}

	// A list is provided by AWS
	if len(external.QueueConfigurations) != 0 {
		config.QueueConfigurations = LateInitializeQueue(external.QueueConfigurations, config.QueueConfigurations)
	}

	// A list is provided by AWS
	if len(external.TopicConfigurations) != 0 {
		config.TopicConfigurations = LateInitializeTopic(external.TopicConfigurations, config.TopicConfigurations)
	}
	return nil
}

// SubresourceExists checks if the subresource this controller manages currently exists
func (in *NotificationConfigurationClient) SubresourceExists(bucket *v1beta1.Bucket) bool {
	return bucket.Spec.ForProvider.NotificationConfiguration != nil
}

// LateInitializeFilter initializes the external awss3.NotificationConfigurationFilter to a local v1beta.NotificationConfigurationFilter
func LateInitializeFilter(local *v1beta1.NotificationConfigurationFilter, external *types.NotificationConfigurationFilter) *v1beta1.NotificationConfigurationFilter {
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
func LateInitializeEvents(local []string, external []types.Event) []string {
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
func LateInitializeLambda(external []types.LambdaFunctionConfiguration, local []v1beta1.LambdaFunctionConfiguration) []v1beta1.LambdaFunctionConfiguration {
	if len(local) != 0 {
		return local
	}
	local = make([]v1beta1.LambdaFunctionConfiguration, len(external))
	for i, v := range external {
		local[i] = v1beta1.LambdaFunctionConfiguration{
			Events:            LateInitializeEvents(local[i].Events, v.Events),
			Filter:            LateInitializeFilter(local[i].Filter, v.Filter),
			ID:                pointer.LateInitialize(local[i].ID, v.Id),
			LambdaFunctionArn: pointer.LateInitializeValueFromPtr(local[i].LambdaFunctionArn, v.LambdaFunctionArn),
		}
	}
	return local
}

// LateInitializeQueue initializes the external awss3.QueueConfiguration to a local v1beta.QueueConfiguration
func LateInitializeQueue(external []types.QueueConfiguration, local []v1beta1.QueueConfiguration) []v1beta1.QueueConfiguration {
	if len(local) != 0 {
		return local
	}
	local = make([]v1beta1.QueueConfiguration, len(external))
	for i, v := range external {
		local[i] = v1beta1.QueueConfiguration{
			Events:   LateInitializeEvents(local[i].Events, v.Events),
			Filter:   LateInitializeFilter(local[i].Filter, v.Filter),
			ID:       pointer.LateInitialize(local[i].ID, v.Id),
			QueueArn: pointer.LateInitialize(local[i].QueueArn, v.QueueArn),
		}
	}
	return local
}

// LateInitializeTopic initializes the external awss3.TopicConfiguration to a local v1beta.TopicConfiguration
func LateInitializeTopic(external []types.TopicConfiguration, local []v1beta1.TopicConfiguration) []v1beta1.TopicConfiguration {
	if len(local) != 0 {
		return local
	}
	local = make([]v1beta1.TopicConfiguration, len(external))
	for i, v := range external {
		local[i] = v1beta1.TopicConfiguration{
			Events:   LateInitializeEvents(local[i].Events, v.Events),
			Filter:   LateInitializeFilter(local[i].Filter, v.Filter),
			ID:       pointer.LateInitialize(local[i].ID, v.Id),
			TopicArn: pointer.LateInitialize(local[i].TopicArn, v.TopicArn),
		}
	}
	return local
}

func emptyConfiguration(external *awss3.GetBucketNotificationConfigurationOutput) bool {
	return (external == nil) || (len(external.TopicConfigurations) == 0 && len(external.QueueConfigurations) == 0 && len(external.LambdaFunctionConfigurations) == 0)
}
