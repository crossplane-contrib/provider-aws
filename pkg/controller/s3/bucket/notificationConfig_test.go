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
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
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
)

func generateNotificationEvents() []string {
	return []string{"s3:ReducedRedundancyLostObject"}
}

func generateNotificationAWSEvents() []s3.Event {
	return []s3.Event{s3.EventS3ReducedRedundancyLostObject}
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

func generateAWSNotificationFilter() *s3.NotificationConfigurationFilter {
	return &s3.NotificationConfigurationFilter{
		Key: &s3.S3KeyFilter{
			FilterRules: []s3.FilterRule{{
				Name:  s3.FilterRuleNamePrefix,
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

func generateAWSNotification() *s3.NotificationConfiguration {
	return &s3.NotificationConfiguration{
		LambdaFunctionConfigurations: []s3.LambdaFunctionConfiguration{{
			Events:            generateNotificationAWSEvents(),
			Filter:            generateAWSNotificationFilter(),
			Id:                &id,
			LambdaFunctionArn: &lambdaArn,
		}},
		QueueConfigurations: []s3.QueueConfiguration{{
			Events:   generateNotificationAWSEvents(),
			Filter:   generateAWSNotificationFilter(),
			Id:       &id,
			QueueArn: &queueArn,
		}},
		TopicConfigurations: []s3.TopicConfiguration{{
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
					MockGetBucketNotificationConfigurationRequest: func(input *s3.GetBucketNotificationConfigurationInput) s3.GetBucketNotificationConfigurationRequest {
						return s3.GetBucketNotificationConfigurationRequest{
							Request: s3Testing.CreateRequest(errBoom, &s3.GetBucketNotificationConfigurationOutput{}),
						}
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
				err:    errors.Wrap(errBoom, notificationGetFailed),
			},
		},
		"UpdateNeededFull": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithNotificationConfig(generateNotificationConfig())),
				cl: NewNotificationConfigurationClient(fake.MockBucketClient{
					MockGetBucketNotificationConfigurationRequest: func(input *s3.GetBucketNotificationConfigurationInput) s3.GetBucketNotificationConfigurationRequest {
						return s3.GetBucketNotificationConfigurationRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketNotificationConfigurationOutput{}),
						}
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
					MockGetBucketNotificationConfigurationRequest: func(input *s3.GetBucketNotificationConfigurationInput) s3.GetBucketNotificationConfigurationRequest {
						return s3.GetBucketNotificationConfigurationRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketNotificationConfigurationOutput{
								LambdaFunctionConfigurations: generateAWSNotification().LambdaFunctionConfigurations,
								TopicConfigurations:          generateAWSNotification().TopicConfigurations,
							}),
						}
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
					MockGetBucketNotificationConfigurationRequest: func(input *s3.GetBucketNotificationConfigurationInput) s3.GetBucketNotificationConfigurationRequest {
						return s3.GetBucketNotificationConfigurationRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketNotificationConfigurationOutput{}),
						}
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
					MockGetBucketNotificationConfigurationRequest: func(input *s3.GetBucketNotificationConfigurationInput) s3.GetBucketNotificationConfigurationRequest {
						return s3.GetBucketNotificationConfigurationRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketNotificationConfigurationOutput{
								LambdaFunctionConfigurations: generateAWSNotification().LambdaFunctionConfigurations,
								QueueConfigurations:          generateAWSNotification().QueueConfigurations,
								TopicConfigurations:          generateAWSNotification().TopicConfigurations,
							}),
						}
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
					MockPutBucketNotificationConfigurationRequest: func(input *s3.PutBucketNotificationConfigurationInput) s3.PutBucketNotificationConfigurationRequest {
						return s3.PutBucketNotificationConfigurationRequest{
							Request: s3Testing.CreateRequest(errBoom, &s3.PutBucketNotificationConfigurationOutput{}),
						}
					},
				}),
			},
			want: want{
				err: errors.Wrap(errBoom, notificationPutFailed),
			},
		},
		"InvalidConfig": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithNotificationConfig(generateNotificationConfig())),
				cl: NewNotificationConfigurationClient(fake.MockBucketClient{
					MockPutBucketNotificationConfigurationRequest: func(input *s3.PutBucketNotificationConfigurationInput) s3.PutBucketNotificationConfigurationRequest {
						return s3.PutBucketNotificationConfigurationRequest{
							Request: s3Testing.CreateRequest(nil, &s3.PutBucketNotificationConfigurationOutput{}),
						}
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
					MockPutBucketNotificationConfigurationRequest: func(input *s3.PutBucketNotificationConfigurationInput) s3.PutBucketNotificationConfigurationRequest {
						return s3.PutBucketNotificationConfigurationRequest{
							Request: s3Testing.CreateRequest(nil, &s3.PutBucketNotificationConfigurationOutput{}),
						}
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
