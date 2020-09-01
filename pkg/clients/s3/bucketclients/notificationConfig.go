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
}

// CreateNotificationConfigurationClient creates the client for Accelerate Configuration
func CreateNotificationConfigurationClient(parameters v1beta1.BucketParameters) BucketResource {
	return &NotificationConfigurationClient{config: parameters.NotificationConfiguration}
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

// ExistsAndUpdated checks if the resource exists and if it matches the local configuration
func (in *NotificationConfigurationClient) ExistsAndUpdated(ctx context.Context, client s3.BucketClient, bucketName *string) (ResourceStatus, error) {
	conf, err := client.GetBucketNotificationConfigurationRequest(&awss3.GetBucketNotificationConfigurationInput{Bucket: bucketName}).Send(ctx)
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

// GenerateLifecycleConfigurationInput creates the input for the LifecycleConfiguration request for the S3 Client
func (in *NotificationConfigurationClient) GenerateLifecycleConfigurationInput(name string) (*awss3.PutBucketNotificationConfigurationInput, error) {
	conf := in.generateConfiguration()
	return &awss3.PutBucketNotificationConfigurationInput{
		Bucket:                    aws.String(name),
		NotificationConfiguration: conf,
	}, nil
}

// CreateResource sends a request to have resource created on AWS
func (in *NotificationConfigurationClient) CreateResource(ctx context.Context, client s3.BucketClient, cr *v1beta1.Bucket) (managed.ExternalUpdate, error) {
	if in.config != nil {
		input, err := in.GenerateLifecycleConfigurationInput(meta.GetExternalName(cr))
		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, "unable to create input for bucket notification request")
		}
		if _, err := client.PutBucketNotificationConfigurationRequest(input).Send(ctx); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, "cannot put bucket notification")
		}
	}
	return managed.ExternalUpdate{}, nil
}

// DeleteResource creates the request to delete the resource on AWS or set it to the default value.
func (in *NotificationConfigurationClient) DeleteResource(ctx context.Context, client s3.BucketClient, cr *v1beta1.Bucket) error {
	input := &awss3.PutBucketNotificationConfigurationInput{
		Bucket:                    aws.String(meta.GetExternalName(cr)),
		NotificationConfiguration: &awss3.NotificationConfiguration{},
	}
	if _, err := client.PutBucketNotificationConfigurationRequest(input).Send(ctx); err != nil {
		return errors.Wrap(err, "cannot delete bucket notification")
	}
	return nil
}
