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
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/s3/fake"
	s3Testing "github.com/crossplane/provider-aws/pkg/controller/s3/testing"
)

var (
	_               SubresourceClient = &NotificationConfigurationClient{}
	filterRuleName                    = "prefix"
	filterRuleValue                   = "value"
	lambdaArn                         = "lambda::123"
	queueArn                          = "queue::123"
	topicArn                          = "topic::123"
	lostEvent                         = s3types.Event("s3:ReducedRedundancyLostObject")
)

func generateNotificationEvents() []string {
	return []string{"s3:ReducedRedundancyLostObject"}
}

func generateNotificationAWSEvents() []s3types.Event {
	return []s3types.Event{lostEvent}
}

func generateNotificationFilter() *v1beta1.NotificationConfigurationFilter {
	return &v1beta1.NotificationConfigurationFilter{
		Key: &v1beta1.S3KeyFilter{
			FilterRules: []v1beta1.FilterRule{{
				Name:  filterRuleName,
				Value: &filterRuleValue,
			}},
		},
	}
}

func generateAWSNotificationFilter() *s3types.NotificationConfigurationFilter {
	return &s3types.NotificationConfigurationFilter{
		Key: &s3types.S3KeyFilter{
			FilterRules: []s3types.FilterRule{{
				Name:  s3types.FilterRuleNamePrefix,
				Value: &filterRuleValue,
			}},
		},
	}
}

func generateNotificationConfig() *v1beta1.NotificationConfiguration {
	return &v1beta1.NotificationConfiguration{
		LambdaFunctionConfigurations: []v1beta1.LambdaFunctionConfiguration{{
			Events:            generateNotificationEvents(),
			Filter:            generateNotificationFilter(),
			ID:                &id,
			LambdaFunctionArn: lambdaArn,
		}},
		QueueConfigurations: []v1beta1.QueueConfiguration{{
			Events:   generateNotificationEvents(),
			Filter:   generateNotificationFilter(),
			ID:       &id,
			QueueArn: queueArn,
		}},
		TopicConfigurations: []v1beta1.TopicConfiguration{{
			Events:   generateNotificationEvents(),
			Filter:   generateNotificationFilter(),
			ID:       &id,
			TopicArn: &topicArn,
		}},
	}
}

func generateAWSNotification() *s3types.NotificationConfiguration {
	return &s3types.NotificationConfiguration{
		LambdaFunctionConfigurations: []s3types.LambdaFunctionConfiguration{{
			Events:            generateNotificationAWSEvents(),
			Filter:            generateAWSNotificationFilter(),
			Id:                &id,
			LambdaFunctionArn: &lambdaArn,
		}},
		QueueConfigurations: []s3types.QueueConfiguration{{
			Events:   generateNotificationAWSEvents(),
			Filter:   generateAWSNotificationFilter(),
			Id:       &id,
			QueueArn: &queueArn,
		}},
		TopicConfigurations: []s3types.TopicConfiguration{{
			Events:   generateNotificationAWSEvents(),
			Filter:   generateAWSNotificationFilter(),
			Id:       &id,
			TopicArn: &topicArn,
		}},
	}
}

func TestNotificationObserve(t *testing.T) {
	type args struct {
		cl *NotificationConfigurationClient
		b  *v1beta1.Bucket
	}

	type want struct {
		status ResourceStatus
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Error": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithNotificationConfig(generateNotificationConfig())),
				cl: NewNotificationConfigurationClient(fake.MockBucketClient{
					MockGetBucketNotificationConfiguration: func(ctx context.Context, input *s3.GetBucketNotificationConfigurationInput, opts []func(*s3.Options)) (*s3.GetBucketNotificationConfigurationOutput, error) {
						return nil, errBoom
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
				err:    awsclient.Wrap(errBoom, notificationGetFailed),
			},
		},
		"UpdateNeededFull": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithNotificationConfig(generateNotificationConfig())),
				cl: NewNotificationConfigurationClient(fake.MockBucketClient{
					MockGetBucketNotificationConfiguration: func(ctx context.Context, input *s3.GetBucketNotificationConfigurationInput, opts []func(*s3.Options)) (*s3.GetBucketNotificationConfigurationOutput, error) {
						return &s3.GetBucketNotificationConfigurationOutput{}, nil
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
				err:    nil,
			},
		},
		"UpdateNeededPartial": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithNotificationConfig(generateNotificationConfig())),
				cl: NewNotificationConfigurationClient(fake.MockBucketClient{
					MockGetBucketNotificationConfiguration: func(ctx context.Context, input *s3.GetBucketNotificationConfigurationInput, opts []func(*s3.Options)) (*s3.GetBucketNotificationConfigurationOutput, error) {
						return &s3.GetBucketNotificationConfigurationOutput{
							LambdaFunctionConfigurations: generateAWSNotification().LambdaFunctionConfigurations,
							TopicConfigurations:          generateAWSNotification().TopicConfigurations,
						}, nil
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
				err:    nil,
			},
		},
		"NoUpdateNotExists": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithNotificationConfig(nil)),
				cl: NewNotificationConfigurationClient(fake.MockBucketClient{
					MockGetBucketNotificationConfiguration: func(ctx context.Context, input *s3.GetBucketNotificationConfigurationInput, opts []func(*s3.Options)) (*s3.GetBucketNotificationConfigurationOutput, error) {
						return &s3.GetBucketNotificationConfigurationOutput{}, nil
					},
				}),
			},
			want: want{
				status: Updated,
				err:    nil,
			},
		},
		"NoUpdateExists": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithNotificationConfig(generateNotificationConfig())),
				cl: NewNotificationConfigurationClient(fake.MockBucketClient{
					MockGetBucketNotificationConfiguration: func(ctx context.Context, input *s3.GetBucketNotificationConfigurationInput, opts []func(*s3.Options)) (*s3.GetBucketNotificationConfigurationOutput, error) {
						return &s3.GetBucketNotificationConfigurationOutput{
							LambdaFunctionConfigurations: generateAWSNotification().LambdaFunctionConfigurations,
							QueueConfigurations:          generateAWSNotification().QueueConfigurations,
							TopicConfigurations:          generateAWSNotification().TopicConfigurations,
						}, nil
					},
				}),
			},
			want: want{
				status: Updated,
				err:    nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			status, err := tc.args.cl.Observe(context.Background(), tc.args.b)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.status, status); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestNotificationCreateOrUpdate(t *testing.T) {
	type args struct {
		cl *NotificationConfigurationClient
		b  *v1beta1.Bucket
	}

	type want struct {
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Error": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithNotificationConfig(generateNotificationConfig())),
				cl: NewNotificationConfigurationClient(fake.MockBucketClient{
					MockPutBucketNotificationConfiguration: func(ctx context.Context, input *s3.PutBucketNotificationConfigurationInput, opts []func(*s3.Options)) (*s3.PutBucketNotificationConfigurationOutput, error) {
						return nil, errBoom
					},
				}),
			},
			want: want{
				err: awsclient.Wrap(errBoom, notificationPutFailed),
			},
		},
		"InvalidConfig": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithNotificationConfig(generateNotificationConfig())),
				cl: NewNotificationConfigurationClient(fake.MockBucketClient{
					MockPutBucketNotificationConfiguration: func(ctx context.Context, input *s3.PutBucketNotificationConfigurationInput, opts []func(*s3.Options)) (*s3.PutBucketNotificationConfigurationOutput, error) {
						return &s3.PutBucketNotificationConfigurationOutput{}, nil
					},
				}),
			},
			want: want{
				err: nil,
			},
		},
		"SuccessfulCreate": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithNotificationConfig(generateNotificationConfig())),
				cl: NewNotificationConfigurationClient(fake.MockBucketClient{
					MockPutBucketNotificationConfiguration: func(ctx context.Context, input *s3.PutBucketNotificationConfigurationInput, opts []func(*s3.Options)) (*s3.PutBucketNotificationConfigurationOutput, error) {
						return &s3.PutBucketNotificationConfigurationOutput{}, nil
					},
				}),
			},
			want: want{
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := tc.args.cl.CreateOrUpdate(context.Background(), tc.args.b)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestNotifLateInit(t *testing.T) {
	type args struct {
		cl SubresourceClient
		b  *v1beta1.Bucket
	}

	type want struct {
		err error
		cr  *v1beta1.Bucket
	}

	cases := map[string]struct {
		args
		want
	}{
		"Error": {
			args: args{
				b: s3Testing.Bucket(),
				cl: NewNotificationConfigurationClient(fake.MockBucketClient{
					MockGetBucketNotificationConfiguration: func(ctx context.Context, input *s3.GetBucketNotificationConfigurationInput, opts []func(*s3.Options)) (*s3.GetBucketNotificationConfigurationOutput, error) {
						return &s3.GetBucketNotificationConfigurationOutput{}, errBoom
					},
				}),
			},
			want: want{
				err: awsclient.Wrap(errBoom, notificationGetFailed),
				cr:  s3Testing.Bucket(),
			},
		},
		"NoLateInitEmpty": {
			args: args{
				b: s3Testing.Bucket(),
				cl: NewNotificationConfigurationClient(fake.MockBucketClient{
					MockGetBucketNotificationConfiguration: func(ctx context.Context, input *s3.GetBucketNotificationConfigurationInput, opts []func(*s3.Options)) (*s3.GetBucketNotificationConfigurationOutput, error) {
						return &s3.GetBucketNotificationConfigurationOutput{}, nil
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3Testing.Bucket(),
			},
		},
		"SuccessfulLateInit": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithNotificationConfig(nil)),
				cl: NewNotificationConfigurationClient(fake.MockBucketClient{
					MockGetBucketNotificationConfiguration: func(ctx context.Context, input *s3.GetBucketNotificationConfigurationInput, opts []func(*s3.Options)) (*s3.GetBucketNotificationConfigurationOutput, error) {
						return &s3.GetBucketNotificationConfigurationOutput{
							LambdaFunctionConfigurations: generateAWSNotification().LambdaFunctionConfigurations,
							TopicConfigurations:          generateAWSNotification().TopicConfigurations,
							QueueConfigurations:          generateAWSNotification().QueueConfigurations,
						}, nil
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3Testing.Bucket(s3Testing.WithNotificationConfig(generateNotificationConfig())),
			},
		},
		"NoOpLateInit": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithNotificationConfig(generateNotificationConfig())),
				cl: NewNotificationConfigurationClient(fake.MockBucketClient{
					MockGetBucketNotificationConfiguration: func(ctx context.Context, input *s3.GetBucketNotificationConfigurationInput, opts []func(*s3.Options)) (*s3.GetBucketNotificationConfigurationOutput, error) {
						return &s3.GetBucketNotificationConfigurationOutput{}, nil
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3Testing.Bucket(s3Testing.WithNotificationConfig(generateNotificationConfig())),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := tc.args.cl.LateInitialize(context.Background(), tc.args.b)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.b, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
