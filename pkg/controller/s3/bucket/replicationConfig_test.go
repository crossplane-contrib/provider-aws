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

	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
	clients3 "github.com/crossplane/provider-aws/pkg/clients/s3"
	"github.com/crossplane/provider-aws/pkg/clients/s3/fake"
	s3Testing "github.com/crossplane/provider-aws/pkg/controller/s3/testing"
)

var (
	role                              = "replication-role"
	owner                             = "Destination"
	accountID                         = "test-account-id"
	kmsID                             = "encKmsID"
	replicationTime                   = 15
	priority                          = 1
	_               SubresourceClient = &ReplicationConfigurationClient{}
)

func generateReplicationConfig() *v1beta1.ReplicationConfiguration {
	return &v1beta1.ReplicationConfiguration{
		Role: &role,
		Rules: []v1beta1.ReplicationRule{{
			DeleteMarkerReplication: &v1beta1.DeleteMarkerReplication{Status: enabled},
			Destination: v1beta1.Destination{
				AccessControlTranslation: &v1beta1.AccessControlTranslation{Owner: owner},
				Account:                  &accountID,
				Bucket:                   &bucketName,
				EncryptionConfiguration:  &v1beta1.EncryptionConfiguration{ReplicaKmsKeyID: kmsID},
				Metrics: &v1beta1.Metrics{
					EventThreshold: v1beta1.ReplicationTimeValue{Minutes: int64(replicationTime)},
					Status:         enabled,
				},
				ReplicationTime: &v1beta1.ReplicationTime{
					Status: enabled,
					Time:   v1beta1.ReplicationTimeValue{Minutes: int64(replicationTime)},
				},
				StorageClass: &storage,
			},
			ExistingObjectReplication: &v1beta1.ExistingObjectReplication{Status: enabled},
			Filter: &v1beta1.ReplicationRuleFilter{
				And: &v1beta1.ReplicationRuleAndOperator{
					Prefix: &prefix,
					Tags:   tags,
				},
				Prefix: &prefix,
				Tag:    &tag,
			},
			ID:                      &id,
			Priority:                aws.Int64(priority),
			SourceSelectionCriteria: &v1beta1.SourceSelectionCriteria{SseKmsEncryptedObjects: v1beta1.SseKmsEncryptedObjects{Status: enabled}},
			Status:                  enabled,
		}},
	}
}

func generateAWSReplication() *s3.ReplicationConfiguration {
	return &s3.ReplicationConfiguration{
		Role: &role,
		Rules: []s3.ReplicationRule{{
			DeleteMarkerReplication: &s3.DeleteMarkerReplication{Status: s3.DeleteMarkerReplicationStatusEnabled},
			Destination: &s3.Destination{
				AccessControlTranslation: &s3.AccessControlTranslation{Owner: s3.OwnerOverrideDestination},
				Account:                  &accountID,
				Bucket:                   &bucketName,
				EncryptionConfiguration:  &s3.EncryptionConfiguration{ReplicaKmsKeyID: &kmsID},
				Metrics: &s3.Metrics{
					EventThreshold: &s3.ReplicationTimeValue{Minutes: aws.Int64(replicationTime)},
					Status:         s3.MetricsStatusEnabled,
				},
				ReplicationTime: &s3.ReplicationTime{
					Time:   &s3.ReplicationTimeValue{Minutes: aws.Int64(replicationTime)},
					Status: s3.ReplicationTimeStatusEnabled,
				},
				StorageClass: s3.StorageClassOnezoneIa,
			},
			ExistingObjectReplication: &s3.ExistingObjectReplication{Status: s3.ExistingObjectReplicationStatusEnabled},
			Filter: &s3.ReplicationRuleFilter{
				And: &s3.ReplicationRuleAndOperator{
					Prefix: &prefix,
					Tags:   awsTags,
				},
				Prefix: &prefix,
				Tag:    &awsTag,
			},
			ID:                      &id,
			Priority:                aws.Int64(priority),
			SourceSelectionCriteria: &s3.SourceSelectionCriteria{SseKmsEncryptedObjects: &s3.SseKmsEncryptedObjects{Status: s3.SseKmsEncryptedObjectsStatusEnabled}},
			Status:                  s3.ReplicationRuleStatusEnabled,
		}},
	}
}

func TestReplicationObserve(t *testing.T) {
	type args struct {
		cl *ReplicationConfigurationClient
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
				b: s3Testing.Bucket(s3Testing.WithReplConfig(generateReplicationConfig())),
				cl: NewReplicationConfigurationClient(fake.MockBucketClient{
					MockGetBucketReplicationRequest: func(input *s3.GetBucketReplicationInput) s3.GetBucketReplicationRequest {
						return s3.GetBucketReplicationRequest{
							Request: s3Testing.CreateRequest(errBoom, &s3.GetBucketReplicationOutput{}),
						}
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
				err:    errors.Wrap(errBoom, replicationGetFailed),
			},
		},
		"UpdateNeeded": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithReplConfig(generateReplicationConfig())),
				cl: NewReplicationConfigurationClient(fake.MockBucketClient{
					MockGetBucketReplicationRequest: func(input *s3.GetBucketReplicationInput) s3.GetBucketReplicationRequest {
						return s3.GetBucketReplicationRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketReplicationOutput{ReplicationConfiguration: nil}),
						}
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
				err:    nil,
			},
		},
		"NeedsDelete": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithReplConfig(nil)),
				cl: NewReplicationConfigurationClient(fake.MockBucketClient{
					MockGetBucketReplicationRequest: func(input *s3.GetBucketReplicationInput) s3.GetBucketReplicationRequest {
						return s3.GetBucketReplicationRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketReplicationOutput{ReplicationConfiguration: generateAWSReplication()}),
						}
					},
				}),
			},
			want: want{
				status: NeedsDeletion,
				err:    nil,
			},
		},
		"NoUpdateNotExists": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithReplConfig(nil)),
				cl: NewReplicationConfigurationClient(fake.MockBucketClient{
					MockGetBucketReplicationRequest: func(input *s3.GetBucketReplicationInput) s3.GetBucketReplicationRequest {
						return s3.GetBucketReplicationRequest{
							Request: s3Testing.CreateRequest(awserr.New(clients3.ReplicationErrCode, "", nil), &s3.GetBucketReplicationOutput{}),
						}
					},
				}),
			},
			want: want{
				status: Updated,
				err:    nil,
			},
		},
		"NoUpdateNotExistsNil": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithReplConfig(nil)),
				cl: NewReplicationConfigurationClient(fake.MockBucketClient{
					MockGetBucketReplicationRequest: func(input *s3.GetBucketReplicationInput) s3.GetBucketReplicationRequest {
						return s3.GetBucketReplicationRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketReplicationOutput{ReplicationConfiguration: nil}),
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
				b: s3Testing.Bucket(s3Testing.WithReplConfig(generateReplicationConfig())),
				cl: NewReplicationConfigurationClient(fake.MockBucketClient{
					MockGetBucketReplicationRequest: func(input *s3.GetBucketReplicationInput) s3.GetBucketReplicationRequest {
						return s3.GetBucketReplicationRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketReplicationOutput{ReplicationConfiguration: generateAWSReplication()}),
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

func TestReplicationCreateOrUpdate(t *testing.T) {
	type args struct {
		cl *ReplicationConfigurationClient
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
				b: s3Testing.Bucket(s3Testing.WithReplConfig(generateReplicationConfig())),
				cl: NewReplicationConfigurationClient(fake.MockBucketClient{
					MockPutBucketReplicationRequest: func(input *s3.PutBucketReplicationInput) s3.PutBucketReplicationRequest {
						return s3.PutBucketReplicationRequest{
							Request: s3Testing.CreateRequest(errBoom, &s3.PutBucketReplicationOutput{}),
						}
					},
				}),
			},
			want: want{
				err: errors.Wrap(errBoom, replicationPutFailed),
			},
		},
		"InvalidConfig": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithReplConfig(generateReplicationConfig())),
				cl: NewReplicationConfigurationClient(fake.MockBucketClient{
					MockPutBucketReplicationRequest: func(input *s3.PutBucketReplicationInput) s3.PutBucketReplicationRequest {
						return s3.PutBucketReplicationRequest{
							Request: s3Testing.CreateRequest(nil, &s3.PutBucketReplicationOutput{}),
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
				b: s3Testing.Bucket(s3Testing.WithReplConfig(generateReplicationConfig())),
				cl: NewReplicationConfigurationClient(fake.MockBucketClient{
					MockPutBucketReplicationRequest: func(input *s3.PutBucketReplicationInput) s3.PutBucketReplicationRequest {
						return s3.PutBucketReplicationRequest{
							Request: s3Testing.CreateRequest(nil, &s3.PutBucketReplicationOutput{}),
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

func TestReplicationDelete(t *testing.T) {
	type args struct {
		cl *ReplicationConfigurationClient
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
				b: s3Testing.Bucket(s3Testing.WithReplConfig(generateReplicationConfig())),
				cl: NewReplicationConfigurationClient(fake.MockBucketClient{
					MockDeleteBucketReplicationRequest: func(input *s3.DeleteBucketReplicationInput) s3.DeleteBucketReplicationRequest {
						return s3.DeleteBucketReplicationRequest{
							Request: s3Testing.CreateRequest(errBoom, &s3.DeleteBucketReplicationOutput{}),
						}
					},
				}),
			},
			want: want{
				err: errors.Wrap(errBoom, replicationDeleteFailed),
			},
		},
		"SuccessfulDelete": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithReplConfig(generateReplicationConfig())),
				cl: NewReplicationConfigurationClient(fake.MockBucketClient{
					MockDeleteBucketReplicationRequest: func(input *s3.DeleteBucketReplicationInput) s3.DeleteBucketReplicationRequest {
						return s3.DeleteBucketReplicationRequest{
							Request: s3Testing.CreateRequest(nil, &s3.DeleteBucketReplicationOutput{}),
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
			err := tc.args.cl.Delete(context.Background(), tc.args.b)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
