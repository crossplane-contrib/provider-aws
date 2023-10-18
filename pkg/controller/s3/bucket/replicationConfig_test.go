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
	"github.com/aws/smithy-go"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane-contrib/provider-aws/apis/s3/v1beta1"
	clientss3 "github.com/crossplane-contrib/provider-aws/pkg/clients/s3"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/s3/fake"
	s3testing "github.com/crossplane-contrib/provider-aws/pkg/controller/s3/testing"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

var (
	role                              = "replication-role"
	owner                             = "Destination"
	accountID                         = "test-account-id"
	kmsID                             = "encKmsID"
	replicationTime                   = 15
	priority        int32             = 1
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
				EncryptionConfiguration:  &v1beta1.EncryptionConfiguration{ReplicaKmsKeyID: &kmsID},
				Metrics: &v1beta1.Metrics{
					EventThreshold: &v1beta1.ReplicationTimeValue{Minutes: int32(replicationTime)},
					Status:         enabled,
				},
				ReplicationTime: &v1beta1.ReplicationTime{
					Status: enabled,
					Time:   v1beta1.ReplicationTimeValue{Minutes: int32(replicationTime)},
				},
				StorageClass: &storage,
			},
			ExistingObjectReplication: &v1beta1.ExistingObjectReplication{Status: enabled},
			Filter: &v1beta1.ReplicationRuleFilter{
				And: &v1beta1.ReplicationRuleAndOperator{
					Prefix: &prefix,
					Tags:   tags,
				},
			},
			ID:                      &id,
			Priority:                priority,
			SourceSelectionCriteria: &v1beta1.SourceSelectionCriteria{SseKmsEncryptedObjects: v1beta1.SseKmsEncryptedObjects{Status: enabled}},
			Status:                  enabled,
		}},
	}
}

func generateAWSReplication() *s3types.ReplicationConfiguration {
	return &s3types.ReplicationConfiguration{
		Role: &role,
		Rules: []s3types.ReplicationRule{{
			DeleteMarkerReplication: &s3types.DeleteMarkerReplication{Status: s3types.DeleteMarkerReplicationStatusEnabled},
			Destination: &s3types.Destination{
				AccessControlTranslation: &s3types.AccessControlTranslation{Owner: s3types.OwnerOverrideDestination},
				Account:                  &accountID,
				Bucket:                   &bucketName,
				EncryptionConfiguration:  &s3types.EncryptionConfiguration{ReplicaKmsKeyID: &kmsID},
				Metrics: &s3types.Metrics{
					EventThreshold: &s3types.ReplicationTimeValue{Minutes: int32(replicationTime)},
					Status:         s3types.MetricsStatusEnabled,
				},
				ReplicationTime: &s3types.ReplicationTime{
					Time:   &s3types.ReplicationTimeValue{Minutes: int32(replicationTime)},
					Status: s3types.ReplicationTimeStatusEnabled,
				},
				StorageClass: s3types.StorageClassOnezoneIa,
			},
			ExistingObjectReplication: &s3types.ExistingObjectReplication{Status: s3types.ExistingObjectReplicationStatusEnabled},
			Filter: &s3types.ReplicationRuleFilterMemberAnd{
				Value: s3types.ReplicationRuleAndOperator{
					Prefix: &prefix,
					Tags:   awsTags,
				},
			},
			ID:                      &id,
			Priority:                priority,
			SourceSelectionCriteria: &s3types.SourceSelectionCriteria{SseKmsEncryptedObjects: &s3types.SseKmsEncryptedObjects{Status: s3types.SseKmsEncryptedObjectsStatusEnabled}},
			Status:                  s3types.ReplicationRuleStatusEnabled,
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
				b: s3testing.Bucket(s3testing.WithReplConfig(generateReplicationConfig())),
				cl: NewReplicationConfigurationClient(fake.MockBucketClient{
					MockGetBucketReplication: func(ctx context.Context, input *s3.GetBucketReplicationInput, opts []func(*s3.Options)) (*s3.GetBucketReplicationOutput, error) {
						return nil, errBoom
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
				err:    errorutils.Wrap(errBoom, replicationGetFailed),
			},
		},
		"UpdateNeeded": {
			args: args{
				b: s3testing.Bucket(s3testing.WithReplConfig(generateReplicationConfig())),
				cl: NewReplicationConfigurationClient(fake.MockBucketClient{
					MockGetBucketReplication: func(ctx context.Context, input *s3.GetBucketReplicationInput, opts []func(*s3.Options)) (*s3.GetBucketReplicationOutput, error) {
						return &s3.GetBucketReplicationOutput{ReplicationConfiguration: nil}, nil
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
				b: s3testing.Bucket(s3testing.WithReplConfig(nil)),
				cl: NewReplicationConfigurationClient(fake.MockBucketClient{
					MockGetBucketReplication: func(ctx context.Context, input *s3.GetBucketReplicationInput, opts []func(*s3.Options)) (*s3.GetBucketReplicationOutput, error) {
						return &s3.GetBucketReplicationOutput{ReplicationConfiguration: generateAWSReplication()}, nil
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
				b: s3testing.Bucket(s3testing.WithReplConfig(nil)),
				cl: NewReplicationConfigurationClient(fake.MockBucketClient{
					MockGetBucketReplication: func(ctx context.Context, input *s3.GetBucketReplicationInput, opts []func(*s3.Options)) (*s3.GetBucketReplicationOutput, error) {
						return nil, &smithy.GenericAPIError{Code: clientss3.ReplicationNotFoundErrCode}
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
				b: s3testing.Bucket(s3testing.WithReplConfig(nil)),
				cl: NewReplicationConfigurationClient(fake.MockBucketClient{
					MockGetBucketReplication: func(ctx context.Context, input *s3.GetBucketReplicationInput, opts []func(*s3.Options)) (*s3.GetBucketReplicationOutput, error) {
						return &s3.GetBucketReplicationOutput{ReplicationConfiguration: nil}, nil
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
				b: s3testing.Bucket(s3testing.WithReplConfig(generateReplicationConfig())),
				cl: NewReplicationConfigurationClient(fake.MockBucketClient{
					MockGetBucketReplication: func(ctx context.Context, input *s3.GetBucketReplicationInput, opts []func(*s3.Options)) (*s3.GetBucketReplicationOutput, error) {
						return &s3.GetBucketReplicationOutput{ReplicationConfiguration: generateAWSReplication()}, nil
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
				b: s3testing.Bucket(s3testing.WithReplConfig(generateReplicationConfig())),
				cl: NewReplicationConfigurationClient(fake.MockBucketClient{
					MockPutBucketReplication: func(ctx context.Context, input *s3.PutBucketReplicationInput, opts []func(*s3.Options)) (*s3.PutBucketReplicationOutput, error) {
						return nil, errBoom
					},
				}),
			},
			want: want{
				err: errorutils.Wrap(errBoom, replicationPutFailed),
			},
		},
		"InvalidConfig": {
			args: args{
				b: s3testing.Bucket(s3testing.WithReplConfig(generateReplicationConfig())),
				cl: NewReplicationConfigurationClient(fake.MockBucketClient{
					MockPutBucketReplication: func(ctx context.Context, input *s3.PutBucketReplicationInput, opts []func(*s3.Options)) (*s3.PutBucketReplicationOutput, error) {
						return &s3.PutBucketReplicationOutput{}, nil
					},
				}),
			},
			want: want{
				err: nil,
			},
		},
		"SuccessfulCreate": {
			args: args{
				b: s3testing.Bucket(s3testing.WithReplConfig(generateReplicationConfig())),
				cl: NewReplicationConfigurationClient(fake.MockBucketClient{
					MockPutBucketReplication: func(ctx context.Context, input *s3.PutBucketReplicationInput, opts []func(*s3.Options)) (*s3.PutBucketReplicationOutput, error) {
						return &s3.PutBucketReplicationOutput{}, nil
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
				b: s3testing.Bucket(s3testing.WithReplConfig(generateReplicationConfig())),
				cl: NewReplicationConfigurationClient(fake.MockBucketClient{
					MockDeleteBucketReplication: func(ctx context.Context, input *s3.DeleteBucketReplicationInput, opts []func(*s3.Options)) (*s3.DeleteBucketReplicationOutput, error) {
						return nil, errBoom
					},
				}),
			},
			want: want{
				err: errorutils.Wrap(errBoom, replicationDeleteFailed),
			},
		},
		"SuccessfulDelete": {
			args: args{
				b: s3testing.Bucket(s3testing.WithReplConfig(generateReplicationConfig())),
				cl: NewReplicationConfigurationClient(fake.MockBucketClient{
					MockDeleteBucketReplication: func(ctx context.Context, input *s3.DeleteBucketReplicationInput, opts []func(*s3.Options)) (*s3.DeleteBucketReplicationOutput, error) {
						return &s3.DeleteBucketReplicationOutput{}, nil
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

func TestReplicationLateInit(t *testing.T) {
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
				b: s3testing.Bucket(),
				cl: NewReplicationConfigurationClient(fake.MockBucketClient{
					MockGetBucketReplication: func(ctx context.Context, input *s3.GetBucketReplicationInput, opts []func(*s3.Options)) (*s3.GetBucketReplicationOutput, error) {
						return &s3.GetBucketReplicationOutput{}, errBoom
					},
				}),
			},
			want: want{
				err: errorutils.Wrap(errBoom, replicationGetFailed),
				cr:  s3testing.Bucket(),
			},
		},
		"ErrorReplicationConfigurationNotFound": {
			args: args{
				b: s3testing.Bucket(),
				cl: NewReplicationConfigurationClient(fake.MockBucketClient{
					MockGetBucketReplication: func(ctx context.Context, input *s3.GetBucketReplicationInput, opts []func(*s3.Options)) (*s3.GetBucketReplicationOutput, error) {
						return &s3.GetBucketReplicationOutput{}, &smithy.GenericAPIError{Code: clientss3.ReplicationNotFoundErrCode}
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3testing.Bucket(),
			},
		},
		"NoLateInitNil": {
			args: args{
				b: s3testing.Bucket(),
				cl: NewReplicationConfigurationClient(fake.MockBucketClient{
					MockGetBucketReplication: func(ctx context.Context, input *s3.GetBucketReplicationInput, opts []func(*s3.Options)) (*s3.GetBucketReplicationOutput, error) {
						return &s3.GetBucketReplicationOutput{}, nil
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3testing.Bucket(),
			},
		},
		"NoLateInitEmpty": {
			args: args{
				b: s3testing.Bucket(),
				cl: NewReplicationConfigurationClient(fake.MockBucketClient{
					MockGetBucketReplication: func(ctx context.Context, input *s3.GetBucketReplicationInput, opts []func(*s3.Options)) (*s3.GetBucketReplicationOutput, error) {
						return &s3.GetBucketReplicationOutput{ReplicationConfiguration: nil}, nil
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3testing.Bucket(),
			},
		},
		"SuccessfulLateInit": {
			args: args{
				b: s3testing.Bucket(s3testing.WithReplConfig(nil)),
				cl: NewReplicationConfigurationClient(fake.MockBucketClient{
					MockGetBucketReplication: func(ctx context.Context, input *s3.GetBucketReplicationInput, opts []func(*s3.Options)) (*s3.GetBucketReplicationOutput, error) {
						return &s3.GetBucketReplicationOutput{ReplicationConfiguration: generateAWSReplication()}, nil
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3testing.Bucket(s3testing.WithReplConfig(generateReplicationConfig())),
			},
		},
		"NoOpLateInit": {
			args: args{
				b: s3testing.Bucket(s3testing.WithReplConfig(generateReplicationConfig())),
				cl: NewReplicationConfigurationClient(fake.MockBucketClient{
					MockGetBucketReplication: func(ctx context.Context, input *s3.GetBucketReplicationInput, opts []func(*s3.Options)) (*s3.GetBucketReplicationOutput, error) {
						return &s3.GetBucketReplicationOutput{
							ReplicationConfiguration: &s3types.ReplicationConfiguration{},
						}, nil
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3testing.Bucket(s3testing.WithReplConfig(generateReplicationConfig())),
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

func TestIsUpToDate(t *testing.T) {
	type args struct {
		cr *v1beta1.ReplicationConfiguration
		b  *s3types.ReplicationConfiguration
	}

	type want struct {
		isUpToDate ResourceStatus
		err        error
	}
	cases := map[string]struct {
		args
		want
	}{
		"IsUpToDate": {
			args: args{
				cr: &v1beta1.ReplicationConfiguration{
					Role: &role,
					Rules: []v1beta1.ReplicationRule{{
						DeleteMarkerReplication: &v1beta1.DeleteMarkerReplication{Status: enabled},
						Destination: v1beta1.Destination{
							AccessControlTranslation: &v1beta1.AccessControlTranslation{Owner: owner},
							Account:                  &accountID,
							Bucket:                   &bucketName,
							EncryptionConfiguration:  &v1beta1.EncryptionConfiguration{ReplicaKmsKeyID: &kmsID},
							Metrics: &v1beta1.Metrics{
								EventThreshold: &v1beta1.ReplicationTimeValue{Minutes: int32(replicationTime)},
								Status:         enabled,
							},
							ReplicationTime: &v1beta1.ReplicationTime{
								Status: enabled,
								Time:   v1beta1.ReplicationTimeValue{Minutes: int32(replicationTime)},
							},
							StorageClass: &storage,
						},
						ExistingObjectReplication: &v1beta1.ExistingObjectReplication{Status: enabled},
						Filter: &v1beta1.ReplicationRuleFilter{
							And: &v1beta1.ReplicationRuleAndOperator{
								Prefix: &prefix,
								Tags:   tags,
							},
						},
						ID:                      &id,
						Priority:                priority,
						SourceSelectionCriteria: &v1beta1.SourceSelectionCriteria{SseKmsEncryptedObjects: v1beta1.SseKmsEncryptedObjects{Status: enabled}},
						Status:                  enabled,
					}},
				},
				b: &s3types.ReplicationConfiguration{
					Role: &role,
					Rules: []s3types.ReplicationRule{{
						DeleteMarkerReplication: &s3types.DeleteMarkerReplication{Status: s3types.DeleteMarkerReplicationStatusEnabled},
						Destination: &s3types.Destination{
							AccessControlTranslation: &s3types.AccessControlTranslation{Owner: s3types.OwnerOverrideDestination},
							Account:                  &accountID,
							Bucket:                   &bucketName,
							EncryptionConfiguration:  &s3types.EncryptionConfiguration{ReplicaKmsKeyID: &kmsID},
							Metrics: &s3types.Metrics{
								EventThreshold: &s3types.ReplicationTimeValue{Minutes: int32(replicationTime)},
								Status:         s3types.MetricsStatusEnabled,
							},
							ReplicationTime: &s3types.ReplicationTime{
								Time:   &s3types.ReplicationTimeValue{Minutes: int32(replicationTime)},
								Status: s3types.ReplicationTimeStatusEnabled,
							},
							StorageClass: s3types.StorageClassOnezoneIa,
						},
						ExistingObjectReplication: &s3types.ExistingObjectReplication{Status: s3types.ExistingObjectReplicationStatusEnabled},
						Filter: &s3types.ReplicationRuleFilterMemberAnd{
							Value: s3types.ReplicationRuleAndOperator{
								Prefix: &prefix,
								Tags:   awsTags,
							},
						},
						ID:                      &id,
						Priority:                priority,
						SourceSelectionCriteria: &s3types.SourceSelectionCriteria{SseKmsEncryptedObjects: &s3types.SseKmsEncryptedObjects{Status: s3types.SseKmsEncryptedObjectsStatusEnabled}},
						Status:                  s3types.ReplicationRuleStatusEnabled,
					}},
				},
			},
			want: want{
				isUpToDate: 0,
			},
		},
		"IsUpToDateRulesOutOfOrder": {
			args: args{
				cr: &v1beta1.ReplicationConfiguration{
					Role: &role,
					Rules: []v1beta1.ReplicationRule{{
						DeleteMarkerReplication: &v1beta1.DeleteMarkerReplication{Status: enabled},
						Destination: v1beta1.Destination{
							AccessControlTranslation: &v1beta1.AccessControlTranslation{Owner: owner},
							Account:                  &accountID,
							Bucket:                   pointer.ToOrNilIfZeroValue("bucket-1"),
							EncryptionConfiguration:  &v1beta1.EncryptionConfiguration{ReplicaKmsKeyID: &kmsID},
							Metrics: &v1beta1.Metrics{
								EventThreshold: &v1beta1.ReplicationTimeValue{Minutes: int32(replicationTime)},
								Status:         enabled,
							},
							ReplicationTime: &v1beta1.ReplicationTime{
								Status: enabled,
								Time:   v1beta1.ReplicationTimeValue{Minutes: int32(replicationTime)},
							},
							StorageClass: &storage,
						},
						ExistingObjectReplication: &v1beta1.ExistingObjectReplication{Status: enabled},
						Filter: &v1beta1.ReplicationRuleFilter{
							And: &v1beta1.ReplicationRuleAndOperator{
								Prefix: &prefix,
								Tags:   tags,
							},
						},
						ID:                      pointer.ToOrNilIfZeroValue("rule-1"),
						Priority:                priority,
						SourceSelectionCriteria: &v1beta1.SourceSelectionCriteria{SseKmsEncryptedObjects: v1beta1.SseKmsEncryptedObjects{Status: enabled}},
						Status:                  enabled,
					},
						{
							DeleteMarkerReplication: &v1beta1.DeleteMarkerReplication{Status: enabled},
							Destination: v1beta1.Destination{
								AccessControlTranslation: &v1beta1.AccessControlTranslation{Owner: owner},
								Account:                  &accountID,
								Bucket:                   pointer.ToOrNilIfZeroValue("bucket-2"),
								EncryptionConfiguration:  &v1beta1.EncryptionConfiguration{ReplicaKmsKeyID: &kmsID},
								Metrics: &v1beta1.Metrics{
									EventThreshold: &v1beta1.ReplicationTimeValue{Minutes: int32(replicationTime)},
									Status:         enabled,
								},
								ReplicationTime: &v1beta1.ReplicationTime{
									Status: enabled,
									Time:   v1beta1.ReplicationTimeValue{Minutes: int32(replicationTime)},
								},
								StorageClass: &storage,
							},
							ExistingObjectReplication: &v1beta1.ExistingObjectReplication{Status: enabled},
							Filter: &v1beta1.ReplicationRuleFilter{
								And: &v1beta1.ReplicationRuleAndOperator{
									Prefix: &prefix,
									Tags:   tags,
								},
							},
							ID:                      pointer.ToOrNilIfZeroValue("rule-2"),
							Priority:                priority,
							SourceSelectionCriteria: &v1beta1.SourceSelectionCriteria{SseKmsEncryptedObjects: v1beta1.SseKmsEncryptedObjects{Status: enabled}},
							Status:                  enabled,
						}},
				},
				b: &s3types.ReplicationConfiguration{
					Role: &role,
					Rules: []s3types.ReplicationRule{{
						DeleteMarkerReplication: &s3types.DeleteMarkerReplication{Status: s3types.DeleteMarkerReplicationStatusEnabled},
						Destination: &s3types.Destination{
							AccessControlTranslation: &s3types.AccessControlTranslation{Owner: s3types.OwnerOverrideDestination},
							Account:                  &accountID,
							Bucket:                   pointer.ToOrNilIfZeroValue("bucket-2"),
							EncryptionConfiguration:  &s3types.EncryptionConfiguration{ReplicaKmsKeyID: &kmsID},
							Metrics: &s3types.Metrics{
								EventThreshold: &s3types.ReplicationTimeValue{Minutes: int32(replicationTime)},
								Status:         s3types.MetricsStatusEnabled,
							},
							ReplicationTime: &s3types.ReplicationTime{
								Time:   &s3types.ReplicationTimeValue{Minutes: int32(replicationTime)},
								Status: s3types.ReplicationTimeStatusEnabled,
							},
							StorageClass: s3types.StorageClassOnezoneIa,
						},
						ExistingObjectReplication: &s3types.ExistingObjectReplication{Status: s3types.ExistingObjectReplicationStatusEnabled},
						Filter: &s3types.ReplicationRuleFilterMemberAnd{
							Value: s3types.ReplicationRuleAndOperator{
								Prefix: &prefix,
								Tags:   awsTags,
							},
						},
						ID:                      pointer.ToOrNilIfZeroValue("rule-2"),
						Priority:                priority,
						SourceSelectionCriteria: &s3types.SourceSelectionCriteria{SseKmsEncryptedObjects: &s3types.SseKmsEncryptedObjects{Status: s3types.SseKmsEncryptedObjectsStatusEnabled}},
						Status:                  s3types.ReplicationRuleStatusEnabled,
					},
						{
							DeleteMarkerReplication: &s3types.DeleteMarkerReplication{Status: s3types.DeleteMarkerReplicationStatusEnabled},
							Destination: &s3types.Destination{
								AccessControlTranslation: &s3types.AccessControlTranslation{Owner: s3types.OwnerOverrideDestination},
								Account:                  &accountID,
								Bucket:                   pointer.ToOrNilIfZeroValue("bucket-1"),
								EncryptionConfiguration:  &s3types.EncryptionConfiguration{ReplicaKmsKeyID: &kmsID},
								Metrics: &s3types.Metrics{
									EventThreshold: &s3types.ReplicationTimeValue{Minutes: int32(replicationTime)},
									Status:         s3types.MetricsStatusEnabled,
								},
								ReplicationTime: &s3types.ReplicationTime{
									Time:   &s3types.ReplicationTimeValue{Minutes: int32(replicationTime)},
									Status: s3types.ReplicationTimeStatusEnabled,
								},
								StorageClass: s3types.StorageClassOnezoneIa,
							},
							ExistingObjectReplication: &s3types.ExistingObjectReplication{Status: s3types.ExistingObjectReplicationStatusEnabled},
							Filter: &s3types.ReplicationRuleFilterMemberAnd{
								Value: s3types.ReplicationRuleAndOperator{
									Prefix: &prefix,
									Tags:   awsTags,
								},
							},
							ID:                      pointer.ToOrNilIfZeroValue("rule-1"),
							Priority:                priority,
							SourceSelectionCriteria: &s3types.SourceSelectionCriteria{SseKmsEncryptedObjects: &s3types.SseKmsEncryptedObjects{Status: s3types.SseKmsEncryptedObjectsStatusEnabled}},
							Status:                  s3types.ReplicationRuleStatusEnabled,
						}},
				},
			},
			want: want{
				isUpToDate: 0,
			},
		},
		"IsUpToDateTagsOutOfOrder": {
			args: args{
				cr: &v1beta1.ReplicationConfiguration{
					Role: &role,
					Rules: []v1beta1.ReplicationRule{{
						DeleteMarkerReplication: &v1beta1.DeleteMarkerReplication{Status: enabled},
						Destination: v1beta1.Destination{
							AccessControlTranslation: &v1beta1.AccessControlTranslation{Owner: owner},
							Account:                  &accountID,
							Bucket:                   &bucketName,
							EncryptionConfiguration:  &v1beta1.EncryptionConfiguration{ReplicaKmsKeyID: &kmsID},
							Metrics: &v1beta1.Metrics{
								EventThreshold: &v1beta1.ReplicationTimeValue{Minutes: int32(replicationTime)},
								Status:         enabled,
							},
							ReplicationTime: &v1beta1.ReplicationTime{
								Status: enabled,
								Time:   v1beta1.ReplicationTimeValue{Minutes: int32(replicationTime)},
							},
							StorageClass: &storage,
						},
						ExistingObjectReplication: &v1beta1.ExistingObjectReplication{Status: enabled},
						Filter: &v1beta1.ReplicationRuleFilter{
							And: &v1beta1.ReplicationRuleAndOperator{
								Prefix: &prefix,
								Tags: []v1beta1.Tag{{
									Key:   "test",
									Value: "value",
								},
									{
										Key:   "xyz",
										Value: "abc",
									}},
							},
						},
						ID:                      &id,
						Priority:                priority,
						SourceSelectionCriteria: &v1beta1.SourceSelectionCriteria{SseKmsEncryptedObjects: v1beta1.SseKmsEncryptedObjects{Status: enabled}},
						Status:                  enabled,
					}},
				},
				b: &s3types.ReplicationConfiguration{
					Role: &role,
					Rules: []s3types.ReplicationRule{{
						DeleteMarkerReplication: &s3types.DeleteMarkerReplication{Status: s3types.DeleteMarkerReplicationStatusEnabled},
						Destination: &s3types.Destination{
							AccessControlTranslation: &s3types.AccessControlTranslation{Owner: s3types.OwnerOverrideDestination},
							Account:                  &accountID,
							Bucket:                   &bucketName,
							EncryptionConfiguration:  &s3types.EncryptionConfiguration{ReplicaKmsKeyID: &kmsID},
							Metrics: &s3types.Metrics{
								EventThreshold: &s3types.ReplicationTimeValue{Minutes: int32(replicationTime)},
								Status:         s3types.MetricsStatusEnabled,
							},
							ReplicationTime: &s3types.ReplicationTime{
								Time:   &s3types.ReplicationTimeValue{Minutes: int32(replicationTime)},
								Status: s3types.ReplicationTimeStatusEnabled,
							},
							StorageClass: s3types.StorageClassOnezoneIa,
						},
						ExistingObjectReplication: &s3types.ExistingObjectReplication{Status: s3types.ExistingObjectReplicationStatusEnabled},
						Filter: &s3types.ReplicationRuleFilterMemberAnd{
							Value: s3types.ReplicationRuleAndOperator{
								Prefix: &prefix,
								Tags: []s3types.Tag{
									{
										Key:   pointer.ToOrNilIfZeroValue("xyz"),
										Value: pointer.ToOrNilIfZeroValue("abc"),
									},
									{
										Key:   pointer.ToOrNilIfZeroValue("test"),
										Value: pointer.ToOrNilIfZeroValue("value"),
									},
								},
							},
						},
						ID:                      &id,
						Priority:                priority,
						SourceSelectionCriteria: &s3types.SourceSelectionCriteria{SseKmsEncryptedObjects: &s3types.SseKmsEncryptedObjects{Status: s3types.SseKmsEncryptedObjectsStatusEnabled}},
						Status:                  s3types.ReplicationRuleStatusEnabled,
					}},
				},
			},
			want: want{
				isUpToDate: 0,
			},
		},
		"IsUpToDateEmptyFilter": {
			args: args{
				cr: &v1beta1.ReplicationConfiguration{
					Role: &role,
					Rules: []v1beta1.ReplicationRule{{
						DeleteMarkerReplication: &v1beta1.DeleteMarkerReplication{Status: enabled},
						Destination: v1beta1.Destination{
							AccessControlTranslation: &v1beta1.AccessControlTranslation{Owner: owner},
							Account:                  &accountID,
							Bucket:                   &bucketName,
							EncryptionConfiguration:  &v1beta1.EncryptionConfiguration{ReplicaKmsKeyID: &kmsID},
							Metrics: &v1beta1.Metrics{
								EventThreshold: &v1beta1.ReplicationTimeValue{Minutes: int32(replicationTime)},
								Status:         enabled,
							},
							ReplicationTime: &v1beta1.ReplicationTime{
								Status: enabled,
								Time:   v1beta1.ReplicationTimeValue{Minutes: int32(replicationTime)},
							},
							StorageClass: &storage,
						},
						ExistingObjectReplication: &v1beta1.ExistingObjectReplication{Status: enabled},
						Filter:                    nil,
						ID:                        &id,
						Priority:                  priority,
						SourceSelectionCriteria:   &v1beta1.SourceSelectionCriteria{SseKmsEncryptedObjects: v1beta1.SseKmsEncryptedObjects{Status: enabled}},
						Status:                    enabled,
					}},
				},
				b: &s3types.ReplicationConfiguration{
					Role: &role,
					Rules: []s3types.ReplicationRule{{
						DeleteMarkerReplication: &s3types.DeleteMarkerReplication{Status: s3types.DeleteMarkerReplicationStatusEnabled},
						Destination: &s3types.Destination{
							AccessControlTranslation: &s3types.AccessControlTranslation{Owner: s3types.OwnerOverrideDestination},
							Account:                  &accountID,
							Bucket:                   &bucketName,
							EncryptionConfiguration:  &s3types.EncryptionConfiguration{ReplicaKmsKeyID: &kmsID},
							Metrics: &s3types.Metrics{
								EventThreshold: &s3types.ReplicationTimeValue{Minutes: int32(replicationTime)},
								Status:         s3types.MetricsStatusEnabled,
							},
							ReplicationTime: &s3types.ReplicationTime{
								Time:   &s3types.ReplicationTimeValue{Minutes: int32(replicationTime)},
								Status: s3types.ReplicationTimeStatusEnabled,
							},
							StorageClass: s3types.StorageClassOnezoneIa,
						},
						ExistingObjectReplication: &s3types.ExistingObjectReplication{Status: s3types.ExistingObjectReplicationStatusEnabled},
						Filter: &s3types.ReplicationRuleFilterMemberPrefix{
							Value: "",
						},
						ID:                      &id,
						Priority:                priority,
						SourceSelectionCriteria: &s3types.SourceSelectionCriteria{SseKmsEncryptedObjects: &s3types.SseKmsEncryptedObjects{Status: s3types.SseKmsEncryptedObjectsStatusEnabled}},
						Status:                  s3types.ReplicationRuleStatusEnabled,
					}},
				},
			},
			want: want{
				isUpToDate: 0,
			},
		},
		"IsUpToDateFalseRuleStatusDisabled": {
			args: args{
				cr: &v1beta1.ReplicationConfiguration{
					Role: &role,
					Rules: []v1beta1.ReplicationRule{{
						DeleteMarkerReplication: &v1beta1.DeleteMarkerReplication{Status: enabled},
						Destination: v1beta1.Destination{
							AccessControlTranslation: &v1beta1.AccessControlTranslation{Owner: owner},
							Account:                  &accountID,
							Bucket:                   &bucketName,
							EncryptionConfiguration:  &v1beta1.EncryptionConfiguration{ReplicaKmsKeyID: &kmsID},
							Metrics: &v1beta1.Metrics{
								EventThreshold: &v1beta1.ReplicationTimeValue{Minutes: int32(replicationTime)},
								Status:         enabled,
							},
							ReplicationTime: &v1beta1.ReplicationTime{
								Status: enabled,
								Time:   v1beta1.ReplicationTimeValue{Minutes: int32(replicationTime)},
							},
							StorageClass: &storage,
						},
						ExistingObjectReplication: &v1beta1.ExistingObjectReplication{Status: enabled},
						Filter: &v1beta1.ReplicationRuleFilter{
							And: &v1beta1.ReplicationRuleAndOperator{
								Prefix: &prefix,
								Tags:   tags,
							},
						},
						ID:                      &id,
						Priority:                priority,
						SourceSelectionCriteria: &v1beta1.SourceSelectionCriteria{SseKmsEncryptedObjects: v1beta1.SseKmsEncryptedObjects{Status: enabled}},
						Status:                  "Disabled",
					}},
				},
				b: &s3types.ReplicationConfiguration{
					Role: &role,
					Rules: []s3types.ReplicationRule{{
						DeleteMarkerReplication: &s3types.DeleteMarkerReplication{Status: s3types.DeleteMarkerReplicationStatusEnabled},
						Destination: &s3types.Destination{
							AccessControlTranslation: &s3types.AccessControlTranslation{Owner: s3types.OwnerOverrideDestination},
							Account:                  &accountID,
							Bucket:                   &bucketName,
							EncryptionConfiguration:  &s3types.EncryptionConfiguration{ReplicaKmsKeyID: &kmsID},
							Metrics: &s3types.Metrics{
								EventThreshold: &s3types.ReplicationTimeValue{Minutes: int32(replicationTime)},
								Status:         s3types.MetricsStatusEnabled,
							},
							ReplicationTime: &s3types.ReplicationTime{
								Time:   &s3types.ReplicationTimeValue{Minutes: int32(replicationTime)},
								Status: s3types.ReplicationTimeStatusEnabled,
							},
							StorageClass: s3types.StorageClassOnezoneIa,
						},
						ExistingObjectReplication: &s3types.ExistingObjectReplication{Status: s3types.ExistingObjectReplicationStatusEnabled},
						Filter: &s3types.ReplicationRuleFilterMemberAnd{
							Value: s3types.ReplicationRuleAndOperator{
								Prefix: &prefix,
								Tags:   awsTags,
							},
						},
						ID:                      &id,
						Priority:                priority,
						SourceSelectionCriteria: &s3types.SourceSelectionCriteria{SseKmsEncryptedObjects: &s3types.SseKmsEncryptedObjects{Status: s3types.SseKmsEncryptedObjectsStatusEnabled}},
						Status:                  s3types.ReplicationRuleStatusEnabled,
					}},
				},
			},
			want: want{
				isUpToDate: 1,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mg := GenerateReplicationConfiguration(tc.args.cr)
			actual, err := IsUpToDate(tc.args.b, mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want error, +got error:\n%s", diff)
			}

			if diff := cmp.Diff(actual, tc.want.isUpToDate, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
