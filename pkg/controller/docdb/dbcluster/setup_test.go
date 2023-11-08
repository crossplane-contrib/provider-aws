/*
Copyright 2021 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS_IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package dbcluster

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/docdb"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/docdb/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/docdb/fake"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

var (
	testAvailabilityZone                 = "test-zone-a"
	testOtherAvailabilityZone            = "test-zone-b"
	testBackupRetentionPeriod            = 10
	testOtherBackupRetentionPeriod       = 1000
	testDBClusterIdentifier              = "some-db-cluster"
	testOtherDBClusterIdentifier         = "some-other-db-cluster"
	testDBClusterArn                     = "some-db-cluster-arn"
	testDBClusterParameterGroupName      = "some-db-cluster-parameter-group"
	testOtherDBClusterParameterGroupName = "some-other-db-cluster-parameter-group"
	testDBSubnetGroupName                = "some-db-subnet-group"
	testCloudWatchLog                    = "some-log"
	testOtherCloudWatchLog               = "some-other-log"
	testEngine                           = "some-engine"
	testEngineVersion                    = "some-engine-version"
	testKMSKeyID                         = "some-key"
	testMasterUserName                   = "some-user"
	testMasterUserPassword               = "some-pw"
	testEndpoint                         = "some-endpoint"
	testReaderEndpoint                   = "some-reader-endpoint"
	testPort                             = 3210
	testOtherPort                        = 5432
	testPresignedURL                     = "some-url"
	testPreferredBackupWindow            = "some-window"
	testOtherPreferredBackupWindow       = "some-other-window"
	testPreferredMaintenanceWindow       = "some-window"
	testOtherPreferredMaintenanceWindow  = "some-other-window"
	testTagKey                           = "some-tag-key"
	testTagValue                         = "some-tag-value"
	testOtherTagKey                      = "some-other-tag-key"
	testOtherTagValue                    = "some-other-tag-value"
	testOtherOtherTagKey                 = "some-other-other-tag-key"
	testOtherOtherTagValue               = "some-other-other-tag-value"
	testVpcSecurityGroup                 = "some-group"
	testOtherVpcSecurityGroup            = "some-other-group"
	testFinalDBSnapshotIdentifier        = "some-snapshot"
	testMasterPasswordSecretNamespace    = "some-namespace"
	testMasterPasswordSecretName         = "some-name"
	testMasterPasswordSecretKey          = "some-key"

	testErrDescribeDBClustersFailed = "DescribeDBClusters failed"
	testErrCreateDBClusterFailed    = "CreateDBCluster failed"
	testErrDeleteDBClusterFailed    = "DeleteDBCluster failed"
	testErrModifyDBClusterFailed    = "ModifyDBCluster failed"
	testErrGetSecret                = "testErrGetSecret"

	timeNow = time.Now()
)

type args struct {
	docdb *fake.MockDocDBClient
	kube  client.Client
	cr    *svcapitypes.DBCluster
}

type docDBModifier func(*svcapitypes.DBCluster)

func toStringPtrArray(values ...string) []*string {
	ptrArr := make([]*string, len(values))
	for i, s := range values {
		ptrArr[i] = pointer.ToOrNilIfZeroValue(s)
	}
	return ptrArr
}

func instance(m ...docDBModifier) *svcapitypes.DBCluster {
	cr := &svcapitypes.DBCluster{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func withExternalName(value string) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		meta.SetExternalName(o, value)

	}
}

func withDeletionTimestamp(v *metav1.Time) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.SetDeletionTimestamp(v)
	}
}

func withDBClusterIdentifier(value string) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Status.AtProvider.DBClusterIdentifier = pointer.ToOrNilIfZeroValue(value)
	}
}

func withConditions(value ...xpv1.Condition) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Status.SetConditions(value...)
	}
}

func withStatus(value string) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Status.AtProvider.Status = pointer.ToOrNilIfZeroValue(value)
	}
}

func withAvailabilityZones(values ...string) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Spec.ForProvider.AvailabilityZones = toStringPtrArray(values...)
	}
}

func withBackupRetentionPeriod(value int) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Spec.ForProvider.BackupRetentionPeriod = pointer.ToIntAsInt64(value)
	}
}

func withDBClusterParameterGroupName(value string) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Spec.ForProvider.DBClusterParameterGroupName = pointer.ToOrNilIfZeroValue(value)
	}
}

func withStatusDBClusterParameterGroupName(value string) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Status.AtProvider.DBClusterParameterGroup = pointer.ToOrNilIfZeroValue(value)
	}
}

func withStatusDBClusterArn(value string) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Status.AtProvider.DBClusterARN = pointer.ToOrNilIfZeroValue(value)
	}
}

func withDBSubnetGroup(value string) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Spec.ForProvider.DBSubnetGroupName = pointer.ToOrNilIfZeroValue(value)
	}
}

func withDeletionProtection(value bool) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Spec.ForProvider.DeletionProtection = ptr.To(value)
	}
}

func withEnableCloudWatchLogsExports(values ...string) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Spec.ForProvider.EnableCloudwatchLogsExports = toStringPtrArray(values...)
	}
}

func withStatusEnableCloudWatchLogsExports(values ...string) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Status.AtProvider.EnabledCloudwatchLogsExports = toStringPtrArray(values...)
	}
}

func withEngine(value string) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Spec.ForProvider.Engine = pointer.ToOrNilIfZeroValue(value)
	}
}

func withEngineVersion(value string) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Spec.ForProvider.EngineVersion = pointer.ToOrNilIfZeroValue(value)
	}
}

func withMasterUserName(value string) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Spec.ForProvider.MasterUsername = pointer.ToOrNilIfZeroValue(value)
	}
}

func withMasterPasswordSecretRef(namesapce, name, key string) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Spec.ForProvider.MasterUserPasswordSecretRef = &xpv1.SecretKeySelector{
			SecretReference: xpv1.SecretReference{
				Name:      name,
				Namespace: o.Namespace,
			},
			Key: key,
		}
	}
}

func withKmsKeyID(value string) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Spec.ForProvider.KMSKeyID = pointer.ToOrNilIfZeroValue(value)
	}
}

func withPort(value int) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Spec.ForProvider.Port = pointer.ToIntAsInt64(value)
	}
}

func withEndpoint(value string) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Status.AtProvider.Endpoint = pointer.ToOrNilIfZeroValue(value)
	}
}

func withReaderEndpoint(value string) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Status.AtProvider.ReaderEndpoint = pointer.ToOrNilIfZeroValue(value)
	}
}

func withPreSignedURL(value string) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Spec.ForProvider.PreSignedURL = pointer.ToOrNilIfZeroValue(value)
	}
}

func withPreferredBackupWindow(value string) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Spec.ForProvider.PreferredBackupWindow = pointer.ToOrNilIfZeroValue(value)
	}
}

func withPreferredMaintenanceWindow(value string) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Spec.ForProvider.PreferredMaintenanceWindow = pointer.ToOrNilIfZeroValue(value)
	}
}

func withStorageEncrypted(value bool) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Spec.ForProvider.StorageEncrypted = pointer.ToOrNilIfZeroValue(value)
	}
}

func withVpcSecurityGroupIds(values ...string) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Spec.ForProvider.VPCSecurityGroupIDs = toStringPtrArray(values...)
	}
}

func withTags(values ...*svcapitypes.Tag) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		if values != nil {
			o.Spec.ForProvider.Tags = values
		} else {
			o.Spec.ForProvider.Tags = []*svcapitypes.Tag{}
		}
	}
}

func withSkipFinalSnapshot(value bool) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Spec.ForProvider.SkipFinalSnapshot = pointer.ToOrNilIfZeroValue(value)
	}
}

func withFinalDBSnapshotIdentifier(value string) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Spec.ForProvider.FinalDBSnapshotIdentifier = pointer.ToOrNilIfZeroValue(value)
	}
}

func withRestoreFromSnapshot(v *svcapitypes.RestoreSnapshotConfiguration) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Spec.ForProvider.RestoreFrom = &svcapitypes.RestoreDBClusterBackupConfiguration{
			Snapshot: v,
			Source:   svcapitypes.RestoreSourceSnapshot,
		}
	}
}

func withRestoreToPointInTime(v *svcapitypes.RestorePointInTimeConfiguration) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Spec.ForProvider.RestoreFrom = &svcapitypes.RestoreDBClusterBackupConfiguration{
			PointInTime: v,
			Source:      svcapitypes.RestoreSourcePointInTime,
		}
	}
}

func generateConnectionDetails(username, password, readerEndpoint, endpoint string, port int) managed.ConnectionDetails {

	mcd := managed.ConnectionDetails{
		xpv1.ResourceCredentialsSecretEndpointKey: []byte(endpoint),
		xpv1.ResourceCredentialsSecretPortKey:     []byte(strconv.Itoa(port)),
		xpv1.ResourceCredentialsSecretUserKey:     []byte(username),
		"readerEndpoint":                          []byte(readerEndpoint),
	}
	if password != "" {
		mcd[xpv1.ResourceCredentialsSecretPasswordKey] = []byte(password)
	}
	return mcd
}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *svcapitypes.DBCluster
		result managed.ExternalObservation
		err    error
		docdb  fake.MockDocDBClientCall
	}

	cases := map[string]struct {
		args
		want
	}{
		"AvailableState_and_changed_BackupRetentionPeriod_should_not_be_UpToDate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBClustersWithContext: func(c context.Context, ddi *docdb.DescribeDBClustersInput, o []request.Option) (*docdb.DescribeDBClustersOutput, error) {
						return &docdb.DescribeDBClustersOutput{
							DBClusters: []*docdb.DBCluster{
								{
									DBClusterIdentifier:   pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
									Status:                pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateAvailable),
									BackupRetentionPeriod: pointer.ToIntAsInt64(testBackupRetentionPeriod),
								},
							},
						}, nil
					},
				},
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withBackupRetentionPeriod(testOtherBackupRetentionPeriod),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withConditions(xpv1.Available()),
					withStatus(svcapitypes.DocDBInstanceStateAvailable),
					withBackupRetentionPeriod(testOtherBackupRetentionPeriod),
					withVpcSecurityGroupIds(),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        false,
					ResourceLateInitialized: true,
					ConnectionDetails:       generateConnectionDetails("", "", "", "", 0),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBClustersWithContext: []*fake.CallDescribeDBClustersWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClustersInput{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
							},
						},
					},
				},
			},
		},
		"AvailableState_and_same_BackupRetentionPeriod_should_be_UpToDate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBClustersWithContext: func(c context.Context, ddi *docdb.DescribeDBClustersInput, o []request.Option) (*docdb.DescribeDBClustersOutput, error) {
						return &docdb.DescribeDBClustersOutput{
							DBClusters: []*docdb.DBCluster{
								{
									DBClusterIdentifier:   pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
									Status:                pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateAvailable),
									BackupRetentionPeriod: pointer.ToIntAsInt64(testBackupRetentionPeriod),
								},
							},
						}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{
							TagList: []*docdb.Tag{},
						}, nil
					},
				},
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withBackupRetentionPeriod(testBackupRetentionPeriod),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withConditions(xpv1.Available()),
					withStatus(svcapitypes.DocDBInstanceStateAvailable),
					withBackupRetentionPeriod(testBackupRetentionPeriod),
					withVpcSecurityGroupIds(),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
					ConnectionDetails:       generateConnectionDetails("", "", "", "", 0),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBClustersWithContext: []*fake.CallDescribeDBClustersWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClustersInput{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
							},
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{},
						},
					},
				},
			},
		},
		"AvailableState_and_no_configured_BackupRetentionPeriod_should_be_UpToDate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBClustersWithContext: func(c context.Context, ddi *docdb.DescribeDBClustersInput, o []request.Option) (*docdb.DescribeDBClustersOutput, error) {
						return &docdb.DescribeDBClustersOutput{
							DBClusters: []*docdb.DBCluster{
								{
									DBClusterIdentifier:   pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
									Status:                pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateAvailable),
									BackupRetentionPeriod: pointer.ToIntAsInt64(testBackupRetentionPeriod),
								},
							},
						}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{
							TagList: []*docdb.Tag{},
						}, nil
					},
				},
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withBackupRetentionPeriod(testBackupRetentionPeriod),
					withExternalName(testDBClusterIdentifier),
					withConditions(xpv1.Available()),
					withStatus(svcapitypes.DocDBInstanceStateAvailable),
					withVpcSecurityGroupIds(),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
					ConnectionDetails:       generateConnectionDetails("", "", "", "", 0),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBClustersWithContext: []*fake.CallDescribeDBClustersWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClustersInput{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
							},
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{},
						},
					},
				},
			},
		},
		"AvailableState_and_same_DBClusterIdentifier_should_be_UpToDate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBClustersWithContext: func(c context.Context, ddi *docdb.DescribeDBClustersInput, o []request.Option) (*docdb.DescribeDBClustersOutput, error) {
						return &docdb.DescribeDBClustersOutput{
							DBClusters: []*docdb.DBCluster{
								{
									DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
									Status:              pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateAvailable),
								},
							},
						}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{
							TagList: []*docdb.Tag{},
						}, nil
					},
				},
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withConditions(xpv1.Available()),
					withStatus(svcapitypes.DocDBInstanceStateAvailable),
					withVpcSecurityGroupIds(),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
					ConnectionDetails:       generateConnectionDetails("", "", "", "", 0),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBClustersWithContext: []*fake.CallDescribeDBClustersWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClustersInput{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
							},
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{},
						},
					},
				},
			},
		},
		"AvailableState_and_changed_DBClusterParameterGroupName_should_not_be_UpToDate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBClustersWithContext: func(c context.Context, ddi *docdb.DescribeDBClustersInput, o []request.Option) (*docdb.DescribeDBClustersOutput, error) {
						return &docdb.DescribeDBClustersOutput{
							DBClusters: []*docdb.DBCluster{
								{
									DBClusterIdentifier:     pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
									Status:                  pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateAvailable),
									DBClusterParameterGroup: pointer.ToOrNilIfZeroValue(testDBClusterParameterGroupName),
								},
							},
						}, nil
					},
				},
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withDBClusterParameterGroupName(testOtherDBClusterParameterGroupName),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withConditions(xpv1.Available()),
					withStatus(svcapitypes.DocDBInstanceStateAvailable),
					withDBClusterParameterGroupName(testOtherDBClusterParameterGroupName),
					withStatusDBClusterParameterGroupName(testDBClusterParameterGroupName),
					withVpcSecurityGroupIds(),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        false,
					ResourceLateInitialized: true,
					ConnectionDetails:       generateConnectionDetails("", "", "", "", 0),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBClustersWithContext: []*fake.CallDescribeDBClustersWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClustersInput{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
							},
						},
					},
				},
			},
		},
		"AvailableState_and_same_DBClusterParameterGroupName_should_be_UpToDate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBClustersWithContext: func(c context.Context, ddi *docdb.DescribeDBClustersInput, o []request.Option) (*docdb.DescribeDBClustersOutput, error) {
						return &docdb.DescribeDBClustersOutput{
							DBClusters: []*docdb.DBCluster{
								{
									DBClusterIdentifier:     pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
									Status:                  pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateAvailable),
									DBClusterParameterGroup: pointer.ToOrNilIfZeroValue(testDBClusterParameterGroupName),
								},
							},
						}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{
							TagList: []*docdb.Tag{},
						}, nil
					},
				},
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withDBClusterParameterGroupName(testDBClusterParameterGroupName),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withConditions(xpv1.Available()),
					withStatus(svcapitypes.DocDBInstanceStateAvailable),
					withDBClusterParameterGroupName(testDBClusterParameterGroupName),
					withStatusDBClusterParameterGroupName(testDBClusterParameterGroupName),
					withVpcSecurityGroupIds(),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
					ConnectionDetails:       generateConnectionDetails("", "", "", "", 0),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBClustersWithContext: []*fake.CallDescribeDBClustersWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClustersInput{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
							},
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{},
						},
					},
				},
			},
		},
		"AvailableState_and_no_spec_DBClusterParameterGroupName_should_be_UpToDate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBClustersWithContext: func(c context.Context, ddi *docdb.DescribeDBClustersInput, o []request.Option) (*docdb.DescribeDBClustersOutput, error) {
						return &docdb.DescribeDBClustersOutput{
							DBClusters: []*docdb.DBCluster{
								{
									DBClusterIdentifier:     pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
									Status:                  pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateAvailable),
									DBClusterParameterGroup: pointer.ToOrNilIfZeroValue(testDBClusterParameterGroupName),
								},
							},
						}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{
							TagList: []*docdb.Tag{},
						}, nil
					},
				},
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withDBClusterParameterGroupName(testDBClusterParameterGroupName),
					withExternalName(testDBClusterIdentifier),
					withConditions(xpv1.Available()),
					withStatus(svcapitypes.DocDBInstanceStateAvailable),
					withStatusDBClusterParameterGroupName(testDBClusterParameterGroupName),
					withVpcSecurityGroupIds(),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
					ConnectionDetails:       generateConnectionDetails("", "", "", "", 0),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBClustersWithContext: []*fake.CallDescribeDBClustersWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClustersInput{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
							},
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{},
						},
					},
				},
			},
		},
		"AvailableState_and_changed_DeletionProtection_should_not_be_UpToDate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBClustersWithContext: func(c context.Context, ddi *docdb.DescribeDBClustersInput, o []request.Option) (*docdb.DescribeDBClustersOutput, error) {
						return &docdb.DescribeDBClustersOutput{
							DBClusters: []*docdb.DBCluster{
								{
									DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
									Status:              pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateAvailable),
									DeletionProtection:  pointer.ToOrNilIfZeroValue(true),
								},
							},
						}, nil
					},
				},
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withDeletionProtection(false),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withConditions(xpv1.Available()),
					withStatus(svcapitypes.DocDBInstanceStateAvailable),
					withDeletionProtection(false),
					withVpcSecurityGroupIds(),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        false,
					ResourceLateInitialized: true,
					ConnectionDetails:       generateConnectionDetails("", "", "", "", 0),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBClustersWithContext: []*fake.CallDescribeDBClustersWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClustersInput{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
							},
						},
					},
				},
			},
		},
		"AvailableState_and_same_DeletionProtection_should_be_UpToDate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBClustersWithContext: func(c context.Context, ddi *docdb.DescribeDBClustersInput, o []request.Option) (*docdb.DescribeDBClustersOutput, error) {
						return &docdb.DescribeDBClustersOutput{
							DBClusters: []*docdb.DBCluster{
								{
									DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
									Status:              pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateAvailable),
									DeletionProtection:  pointer.ToOrNilIfZeroValue(true),
								},
							},
						}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{
							TagList: []*docdb.Tag{},
						}, nil
					},
				},
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withDeletionProtection(true),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withConditions(xpv1.Available()),
					withStatus(svcapitypes.DocDBInstanceStateAvailable),
					withDeletionProtection(true),
					withVpcSecurityGroupIds(),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
					ConnectionDetails:       generateConnectionDetails("", "", "", "", 0),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBClustersWithContext: []*fake.CallDescribeDBClustersWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClustersInput{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
							},
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{},
						},
					},
				},
			},
		},
		"AvailableState_and_no_spec_DeletionProtection_should_be_UpToDate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBClustersWithContext: func(c context.Context, ddi *docdb.DescribeDBClustersInput, o []request.Option) (*docdb.DescribeDBClustersOutput, error) {
						return &docdb.DescribeDBClustersOutput{
							DBClusters: []*docdb.DBCluster{
								{
									DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
									Status:              pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateAvailable),
									DeletionProtection:  pointer.ToOrNilIfZeroValue(true),
								},
							},
						}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{
							TagList: []*docdb.Tag{},
						}, nil
					},
				},
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withConditions(xpv1.Available()),
					withDeletionProtection(true),
					withStatus(svcapitypes.DocDBInstanceStateAvailable),
					withVpcSecurityGroupIds(),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
					ConnectionDetails:       generateConnectionDetails("", "", "", "", 0),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBClustersWithContext: []*fake.CallDescribeDBClustersWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClustersInput{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
							},
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{},
						},
					},
				},
			},
		},
		"AvailableState_and_added_EnableCloudwatchLogsExports_should_not_be_UpToDate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBClustersWithContext: func(c context.Context, ddi *docdb.DescribeDBClustersInput, o []request.Option) (*docdb.DescribeDBClustersOutput, error) {
						return &docdb.DescribeDBClustersOutput{
							DBClusters: []*docdb.DBCluster{
								{
									DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
									Status:              pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateAvailable),
									EnabledCloudwatchLogsExports: []*string{
										pointer.ToOrNilIfZeroValue(testCloudWatchLog),
									},
								},
							},
						}, nil
					},
				},
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withConditions(xpv1.Available()),
					withEnableCloudWatchLogsExports(
						testCloudWatchLog,
						testOtherCloudWatchLog,
					),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withConditions(xpv1.Available()),
					withStatus(svcapitypes.DocDBInstanceStateAvailable),
					withEnableCloudWatchLogsExports(
						testCloudWatchLog,
						testOtherCloudWatchLog,
					),
					withStatusEnableCloudWatchLogsExports(
						testCloudWatchLog,
					),
					withVpcSecurityGroupIds(),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        false,
					ResourceLateInitialized: true,
					ConnectionDetails:       generateConnectionDetails("", "", "", "", 0),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBClustersWithContext: []*fake.CallDescribeDBClustersWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClustersInput{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
							},
						},
					},
				},
			},
		},
		"AvailableState_and_removed_EnableCloudwatchLogsExports_should_not_be_UpToDate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBClustersWithContext: func(c context.Context, ddi *docdb.DescribeDBClustersInput, o []request.Option) (*docdb.DescribeDBClustersOutput, error) {
						return &docdb.DescribeDBClustersOutput{
							DBClusters: []*docdb.DBCluster{
								{
									DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
									Status:              pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateAvailable),
									EnabledCloudwatchLogsExports: []*string{
										pointer.ToOrNilIfZeroValue(testCloudWatchLog),
										pointer.ToOrNilIfZeroValue(testOtherCloudWatchLog),
									},
								},
							},
						}, nil
					},
				},
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withEnableCloudWatchLogsExports(
						testCloudWatchLog,
					),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withConditions(xpv1.Available()),
					withStatus(svcapitypes.DocDBInstanceStateAvailable),
					withEnableCloudWatchLogsExports(
						testCloudWatchLog,
					),
					withStatusEnableCloudWatchLogsExports(
						testCloudWatchLog,
						testOtherCloudWatchLog,
					),
					withVpcSecurityGroupIds(),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        false,
					ResourceLateInitialized: true,
					ConnectionDetails:       generateConnectionDetails("", "", "", "", 0),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBClustersWithContext: []*fake.CallDescribeDBClustersWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClustersInput{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
							},
						},
					},
				},
			},
		},
		"AvailableState_and_same_order_EnableCloudwatchLogsExports_should_be_UpToDate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBClustersWithContext: func(c context.Context, ddi *docdb.DescribeDBClustersInput, o []request.Option) (*docdb.DescribeDBClustersOutput, error) {
						return &docdb.DescribeDBClustersOutput{
							DBClusters: []*docdb.DBCluster{
								{
									DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
									Status:              pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateAvailable),
									EnabledCloudwatchLogsExports: []*string{
										pointer.ToOrNilIfZeroValue(testCloudWatchLog),
										pointer.ToOrNilIfZeroValue(testOtherCloudWatchLog),
									},
								},
							},
						}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{
							TagList: []*docdb.Tag{},
						}, nil
					},
				},
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withEnableCloudWatchLogsExports(
						testCloudWatchLog,
						testOtherCloudWatchLog,
					),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withConditions(xpv1.Available()),
					withStatus(svcapitypes.DocDBInstanceStateAvailable),
					withEnableCloudWatchLogsExports(
						testCloudWatchLog,
						testOtherCloudWatchLog,
					),
					withStatusEnableCloudWatchLogsExports(
						testCloudWatchLog,
						testOtherCloudWatchLog,
					),
					withVpcSecurityGroupIds(),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
					ConnectionDetails:       generateConnectionDetails("", "", "", "", 0),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBClustersWithContext: []*fake.CallDescribeDBClustersWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClustersInput{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
							},
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{},
						},
					},
				},
			},
		},
		"AvailableState_and_different_order_EnableCloudwatchLogsExports_should_be_UpToDate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBClustersWithContext: func(c context.Context, ddi *docdb.DescribeDBClustersInput, o []request.Option) (*docdb.DescribeDBClustersOutput, error) {
						return &docdb.DescribeDBClustersOutput{
							DBClusters: []*docdb.DBCluster{
								{
									DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
									Status:              pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateAvailable),
									EnabledCloudwatchLogsExports: []*string{
										pointer.ToOrNilIfZeroValue(testCloudWatchLog),
										pointer.ToOrNilIfZeroValue(testOtherCloudWatchLog),
									},
								},
							},
						}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{
							TagList: []*docdb.Tag{},
						}, nil
					},
				},
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withEnableCloudWatchLogsExports(
						testOtherCloudWatchLog,
						testCloudWatchLog,
					),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withConditions(xpv1.Available()),
					withStatus(svcapitypes.DocDBInstanceStateAvailable),
					withEnableCloudWatchLogsExports(
						testOtherCloudWatchLog,
						testCloudWatchLog,
					),
					withStatusEnableCloudWatchLogsExports(
						testCloudWatchLog,
						testOtherCloudWatchLog,
					),
					withVpcSecurityGroupIds(),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
					ConnectionDetails:       generateConnectionDetails("", "", "", "", 0),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBClustersWithContext: []*fake.CallDescribeDBClustersWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClustersInput{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
							},
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{},
						},
					},
				},
			},
		},
		"AvailableState_and_no_spec_EnableCloudwatchLogsExports_should_be_UpToDate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBClustersWithContext: func(c context.Context, ddi *docdb.DescribeDBClustersInput, o []request.Option) (*docdb.DescribeDBClustersOutput, error) {
						return &docdb.DescribeDBClustersOutput{
							DBClusters: []*docdb.DBCluster{
								{
									DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
									Status:              pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateAvailable),
									EnabledCloudwatchLogsExports: []*string{
										pointer.ToOrNilIfZeroValue(testCloudWatchLog),
										pointer.ToOrNilIfZeroValue(testOtherCloudWatchLog),
									},
								},
							},
						}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{
							TagList: []*docdb.Tag{},
						}, nil
					},
				},
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withConditions(xpv1.Available()),
					withStatus(svcapitypes.DocDBInstanceStateAvailable),
					withEnableCloudWatchLogsExports(
						testCloudWatchLog,
						testOtherCloudWatchLog,
					),
					withStatusEnableCloudWatchLogsExports(
						testCloudWatchLog,
						testOtherCloudWatchLog,
					),
					withVpcSecurityGroupIds(),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
					ConnectionDetails:       generateConnectionDetails("", "", "", "", 0),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBClustersWithContext: []*fake.CallDescribeDBClustersWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClustersInput{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
							},
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{},
						},
					},
				},
			},
		},
		"AvailableState_and_changed_Port_should_not_be_UpToDate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBClustersWithContext: func(c context.Context, ddi *docdb.DescribeDBClustersInput, o []request.Option) (*docdb.DescribeDBClustersOutput, error) {
						return &docdb.DescribeDBClustersOutput{
							DBClusters: []*docdb.DBCluster{
								{
									DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
									Status:              pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateAvailable),
									Port:                pointer.ToIntAsInt64(testPort),
								},
							},
						}, nil
					},
				},
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withPort(testOtherPort),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withConditions(xpv1.Available()),
					withStatus(svcapitypes.DocDBInstanceStateAvailable),
					withPort(testOtherPort),
					withVpcSecurityGroupIds(),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        false,
					ResourceLateInitialized: true,
					ConnectionDetails:       generateConnectionDetails("", "", "", "", testOtherPort),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBClustersWithContext: []*fake.CallDescribeDBClustersWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClustersInput{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
							},
						},
					},
				},
			},
		},
		"AvailableState_and_same_Port_should_be_UpToDate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBClustersWithContext: func(c context.Context, ddi *docdb.DescribeDBClustersInput, o []request.Option) (*docdb.DescribeDBClustersOutput, error) {
						return &docdb.DescribeDBClustersOutput{
							DBClusters: []*docdb.DBCluster{
								{
									DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
									Status:              pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateAvailable),
									Port:                pointer.ToIntAsInt64(testPort),
								},
							},
						}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{
							TagList: []*docdb.Tag{},
						}, nil
					},
				},
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withPort(testPort),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withConditions(xpv1.Available()),
					withStatus(svcapitypes.DocDBInstanceStateAvailable),
					withPort(testPort),
					withVpcSecurityGroupIds(),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
					ConnectionDetails:       generateConnectionDetails("", "", "", "", testPort),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBClustersWithContext: []*fake.CallDescribeDBClustersWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClustersInput{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
							},
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{},
						},
					},
				},
			},
		},
		"AvailableState_and_no_spec_Port_should_be_UpToDate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBClustersWithContext: func(c context.Context, ddi *docdb.DescribeDBClustersInput, o []request.Option) (*docdb.DescribeDBClustersOutput, error) {
						return &docdb.DescribeDBClustersOutput{
							DBClusters: []*docdb.DBCluster{
								{
									DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
									Status:              pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateAvailable),
									Port:                pointer.ToIntAsInt64(testPort),
								},
							},
						}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{
							TagList: []*docdb.Tag{},
						}, nil
					},
				},
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withPort(testPort),
					withExternalName(testDBClusterIdentifier),
					withConditions(xpv1.Available()),
					withStatus(svcapitypes.DocDBInstanceStateAvailable),
					withVpcSecurityGroupIds(),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
					ConnectionDetails:       generateConnectionDetails("", "", "", "", testPort),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBClustersWithContext: []*fake.CallDescribeDBClustersWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClustersInput{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
							},
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{},
						},
					},
				},
			},
		},
		"AvailableState_and_changed_PreferredBackupWindow_should_not_be_UpToDate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBClustersWithContext: func(c context.Context, ddi *docdb.DescribeDBClustersInput, o []request.Option) (*docdb.DescribeDBClustersOutput, error) {
						return &docdb.DescribeDBClustersOutput{
							DBClusters: []*docdb.DBCluster{
								{
									DBClusterIdentifier:   pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
									Status:                pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateAvailable),
									PreferredBackupWindow: pointer.ToOrNilIfZeroValue(testPreferredBackupWindow),
								},
							},
						}, nil
					},
				},
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withPreferredBackupWindow(testOtherPreferredBackupWindow),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withConditions(xpv1.Available()),
					withStatus(svcapitypes.DocDBInstanceStateAvailable),
					withPreferredBackupWindow(testOtherPreferredBackupWindow),
					withVpcSecurityGroupIds(),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        false,
					ResourceLateInitialized: true,
					ConnectionDetails:       generateConnectionDetails("", "", "", "", 0),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBClustersWithContext: []*fake.CallDescribeDBClustersWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClustersInput{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
							},
						},
					},
				},
			},
		},
		"AvailableState_and_same_PreferredBackupWindow_should_be_UpToDate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBClustersWithContext: func(c context.Context, ddi *docdb.DescribeDBClustersInput, o []request.Option) (*docdb.DescribeDBClustersOutput, error) {
						return &docdb.DescribeDBClustersOutput{
							DBClusters: []*docdb.DBCluster{
								{
									DBClusterIdentifier:   pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
									Status:                pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateAvailable),
									PreferredBackupWindow: pointer.ToOrNilIfZeroValue(testPreferredBackupWindow),
								},
							},
						}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{
							TagList: []*docdb.Tag{},
						}, nil
					},
				},
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withPreferredBackupWindow(testPreferredBackupWindow),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withConditions(xpv1.Available()),
					withStatus(svcapitypes.DocDBInstanceStateAvailable),
					withPreferredBackupWindow(testPreferredBackupWindow),
					withVpcSecurityGroupIds(),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
					ConnectionDetails:       generateConnectionDetails("", "", "", "", 0),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBClustersWithContext: []*fake.CallDescribeDBClustersWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClustersInput{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
							},
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{},
						},
					},
				},
			},
		},
		"AvailableState_and_no_spec_PreferredBackupWindow_should_be_UpToDate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBClustersWithContext: func(c context.Context, ddi *docdb.DescribeDBClustersInput, o []request.Option) (*docdb.DescribeDBClustersOutput, error) {
						return &docdb.DescribeDBClustersOutput{
							DBClusters: []*docdb.DBCluster{
								{
									DBClusterIdentifier:   pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
									Status:                pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateAvailable),
									PreferredBackupWindow: pointer.ToOrNilIfZeroValue(testPreferredBackupWindow),
								},
							},
						}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{
							TagList: []*docdb.Tag{},
						}, nil
					},
				},
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withPreferredBackupWindow(testPreferredBackupWindow),
					withExternalName(testDBClusterIdentifier),
					withConditions(xpv1.Available()),
					withStatus(svcapitypes.DocDBInstanceStateAvailable),
					withVpcSecurityGroupIds(),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
					ConnectionDetails:       generateConnectionDetails("", "", "", "", 0),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBClustersWithContext: []*fake.CallDescribeDBClustersWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClustersInput{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
							},
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{},
						},
					},
				},
			},
		},
		"AvailableState_and_changed_PreferredMaintenanceWindow_should_not_be_UpToDate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBClustersWithContext: func(c context.Context, ddi *docdb.DescribeDBClustersInput, o []request.Option) (*docdb.DescribeDBClustersOutput, error) {
						return &docdb.DescribeDBClustersOutput{
							DBClusters: []*docdb.DBCluster{
								{
									DBClusterIdentifier:        pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
									Status:                     pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateAvailable),
									PreferredMaintenanceWindow: pointer.ToOrNilIfZeroValue(testPreferredMaintenanceWindow),
								},
							},
						}, nil
					},
				},
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withPreferredMaintenanceWindow(testOtherPreferredMaintenanceWindow),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withConditions(xpv1.Available()),
					withStatus(svcapitypes.DocDBInstanceStateAvailable),
					withPreferredMaintenanceWindow(testOtherPreferredMaintenanceWindow),
					withVpcSecurityGroupIds(),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        false,
					ResourceLateInitialized: true,
					ConnectionDetails:       generateConnectionDetails("", "", "", "", 0),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBClustersWithContext: []*fake.CallDescribeDBClustersWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClustersInput{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
							},
						},
					},
				},
			},
		},
		"AvailableState_and_same_PreferredMaintenanceWindow_should_be_UpToDate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBClustersWithContext: func(c context.Context, ddi *docdb.DescribeDBClustersInput, o []request.Option) (*docdb.DescribeDBClustersOutput, error) {
						return &docdb.DescribeDBClustersOutput{
							DBClusters: []*docdb.DBCluster{
								{
									DBClusterIdentifier:        pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
									Status:                     pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateAvailable),
									PreferredMaintenanceWindow: pointer.ToOrNilIfZeroValue(testPreferredMaintenanceWindow),
								},
							},
						}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{
							TagList: []*docdb.Tag{},
						}, nil
					},
				},
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withPreferredMaintenanceWindow(testPreferredBackupWindow),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withConditions(xpv1.Available()),
					withStatus(svcapitypes.DocDBInstanceStateAvailable),
					withPreferredMaintenanceWindow(testPreferredBackupWindow),
					withVpcSecurityGroupIds(),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
					ConnectionDetails:       generateConnectionDetails("", "", "", "", 0),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBClustersWithContext: []*fake.CallDescribeDBClustersWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClustersInput{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
							},
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{},
						},
					},
				},
			},
		},
		"AvailableState_and_no_spec_PreferredMaintenanceWindow_should_be_UpToDate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBClustersWithContext: func(c context.Context, ddi *docdb.DescribeDBClustersInput, o []request.Option) (*docdb.DescribeDBClustersOutput, error) {
						return &docdb.DescribeDBClustersOutput{
							DBClusters: []*docdb.DBCluster{
								{
									DBClusterIdentifier:        pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
									Status:                     pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateAvailable),
									PreferredMaintenanceWindow: pointer.ToOrNilIfZeroValue(testPreferredMaintenanceWindow),
								},
							},
						}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{
							TagList: []*docdb.Tag{},
						}, nil
					},
				},
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withPreferredMaintenanceWindow(testPreferredMaintenanceWindow),
					withExternalName(testDBClusterIdentifier),
					withConditions(xpv1.Available()),
					withStatus(svcapitypes.DocDBInstanceStateAvailable),
					withVpcSecurityGroupIds(),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
					ConnectionDetails:       generateConnectionDetails("", "", "", "", 0),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBClustersWithContext: []*fake.CallDescribeDBClustersWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClustersInput{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
							},
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{},
						},
					},
				},
			},
		},
		"AvailableState_and_Deleted_should_be_UpToDate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBClustersWithContext: func(c context.Context, ddi *docdb.DescribeDBClustersInput, o []request.Option) (*docdb.DescribeDBClustersOutput, error) {
						return &docdb.DescribeDBClustersOutput{
							DBClusters: []*docdb.DBCluster{
								{
									DBClusterIdentifier:        pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
									Status:                     pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateAvailable),
									PreferredMaintenanceWindow: pointer.ToOrNilIfZeroValue(testPreferredMaintenanceWindow),
								},
							},
						}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{
							TagList: []*docdb.Tag{},
						}, nil
					},
				},
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withDeletionTimestamp(&metav1.Time{Time: timeNow}),
					withExternalName(testDBClusterIdentifier),
					withPreferredMaintenanceWindow(testOtherPreferredMaintenanceWindow),
					withVpcSecurityGroupIds(),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withDeletionTimestamp(&metav1.Time{Time: timeNow}),
					withPreferredMaintenanceWindow(testOtherPreferredMaintenanceWindow),
					withExternalName(testDBClusterIdentifier),
					withConditions(xpv1.Available()),
					withStatus(svcapitypes.DocDBInstanceStateAvailable),
					withVpcSecurityGroupIds(),
				),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: generateConnectionDetails("", "", "", "", 0),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBClustersWithContext: []*fake.CallDescribeDBClustersWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClustersInput{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
							},
						},
					},
				},
			},
		},
		"Empty_DescribeClustersOutput": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBClustersWithContext: func(c context.Context, ddi *docdb.DescribeDBClustersInput, o []request.Option) (*docdb.DescribeDBClustersOutput, error) {
						return &docdb.DescribeDBClustersOutput{
							DBClusters: []*docdb.DBCluster{},
						}, nil
					},
				},
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
				),
				result: managed.ExternalObservation{
					ResourceExists:   false,
					ResourceUpToDate: false,
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBClustersWithContext: []*fake.CallDescribeDBClustersWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClustersInput{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
							},
						},
					},
				},
			},
		},
		"ErrDescribeDBClustersWithContext": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBClustersWithContext: func(c context.Context, ddpgi *docdb.DescribeDBClustersInput, o []request.Option) (*docdb.DescribeDBClustersOutput, error) {
						return &docdb.DescribeDBClustersOutput{}, errors.New(testErrDescribeDBClustersFailed)
					},
				},
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
				),
				result: managed.ExternalObservation{
					ResourceExists:   false,
					ResourceUpToDate: false,
				},
				err: errors.Wrap(errors.New(testErrDescribeDBClustersFailed), errDescribe),
				docdb: fake.MockDocDBClientCall{
					DescribeDBClustersWithContext: []*fake.CallDescribeDBClustersWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClustersInput{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
							},
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			opts := []option{setupExternal}
			e := newExternal(tc.args.kube, tc.args.docdb, opts)
			o, err := e.Observe(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.docdb, tc.args.docdb.Called, cmpopts.IgnoreInterfaces(struct{ context.Context }{})); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	type want struct {
		cr     *svcapitypes.DBCluster
		result managed.ExternalCreation
		err    error
		docdb  fake.MockDocDBClientCall
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulCreate": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						s := obj.(*v1.Secret)
						s.Data = map[string][]byte{
							testMasterPasswordSecretKey: []byte(testMasterUserPassword),
						}
						return nil
					},
				},
				docdb: &fake.MockDocDBClient{
					MockCreateDBClusterWithContext: func(c context.Context, cdpgi *docdb.CreateDBClusterInput, o []request.Option) (*docdb.CreateDBClusterOutput, error) {
						return &docdb.CreateDBClusterOutput{
							DBCluster: &docdb.DBCluster{
								AvailabilityZones: []*string{
									&testAvailabilityZone,
									&testOtherAvailabilityZone,
								},
								BackupRetentionPeriod:      pointer.ToIntAsInt64(testBackupRetentionPeriod),
								DBClusterParameterGroup:    &testDBClusterParameterGroupName,
								DBClusterIdentifier:        pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
								DBClusterArn:               pointer.ToOrNilIfZeroValue(testDBClusterArn),
								DeletionProtection:         pointer.ToOrNilIfZeroValue(true),
								Endpoint:                   pointer.ToOrNilIfZeroValue(testEndpoint),
								Engine:                     &testEngine,
								EngineVersion:              &testEngineVersion,
								KmsKeyId:                   &testKMSKeyID,
								MasterUsername:             &testMasterUserName,
								ReaderEndpoint:             pointer.ToOrNilIfZeroValue(testReaderEndpoint),
								Port:                       pointer.ToIntAsInt64(testPort),
								PreferredBackupWindow:      &testPreferredBackupWindow,
								PreferredMaintenanceWindow: &testPreferredMaintenanceWindow,
								StorageEncrypted:           pointer.ToOrNilIfZeroValue(true),
							},
						}, nil
					},
				},
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withAvailabilityZones(
						testAvailabilityZone,
						testOtherAvailabilityZone,
					),
					withBackupRetentionPeriod(testBackupRetentionPeriod),
					withDBClusterParameterGroupName(testDBClusterParameterGroupName),
					withDBSubnetGroup(testDBSubnetGroupName),
					withDeletionProtection(true),
					withEnableCloudWatchLogsExports(
						testCloudWatchLog,
						testOtherCloudWatchLog,
					),
					withEngine(testEngine),
					withEngineVersion(testEngineVersion),
					withKmsKeyID(testKMSKeyID),
					withMasterUserName(testMasterUserName),
					withPort(testPort),
					withPreSignedURL(testPresignedURL),
					withPreferredBackupWindow(testPreferredBackupWindow),
					withPreferredMaintenanceWindow(testPreferredMaintenanceWindow),
					withStorageEncrypted(true),
					withTags(
						&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testTagValue)},
						&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testOtherTagKey), Value: pointer.ToOrNilIfZeroValue(testOtherTagValue)},
					),
					withVpcSecurityGroupIds(
						testVpcSecurityGroup,
						testOtherVpcSecurityGroup,
					),
					withMasterPasswordSecretRef(testMasterPasswordSecretNamespace, testMasterPasswordSecretName, testMasterPasswordSecretKey),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withConditions(xpv1.Creating()),
					withAvailabilityZones(
						testAvailabilityZone,
						testOtherAvailabilityZone,
					),
					withBackupRetentionPeriod(testBackupRetentionPeriod),
					withDBClusterParameterGroupName(testDBClusterParameterGroupName),
					withDBSubnetGroup(testDBSubnetGroupName),
					withStatusDBClusterArn(testDBClusterArn),
					withDeletionProtection(true),
					withEnableCloudWatchLogsExports(
						testCloudWatchLog,
						testOtherCloudWatchLog,
					),
					withEngine(testEngine),
					withEngineVersion(testEngineVersion),
					withKmsKeyID(testKMSKeyID),
					withMasterUserName(testMasterUserName),
					withPort(testPort),
					withPreSignedURL(testPresignedURL),
					withPreferredBackupWindow(testPreferredBackupWindow),
					withPreferredMaintenanceWindow(testPreferredMaintenanceWindow),
					withStorageEncrypted(true),
					withTags(
						&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testTagValue)},
						&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testOtherTagKey), Value: pointer.ToOrNilIfZeroValue(testOtherTagValue)},
					),
					withVpcSecurityGroupIds(
						testVpcSecurityGroup,
						testOtherVpcSecurityGroup,
					),
					withEndpoint(testEndpoint),
					withReaderEndpoint(testReaderEndpoint),
					withMasterPasswordSecretRef(testMasterPasswordSecretNamespace, testMasterPasswordSecretName, testMasterPasswordSecretKey),
					withStatusDBClusterParameterGroupName(testDBClusterParameterGroupName),
				),
				result: managed.ExternalCreation{
					ConnectionDetails: generateConnectionDetails(
						testMasterUserName,
						testMasterUserPassword,
						testReaderEndpoint,
						testEndpoint,
						testPort,
					),
				},
				docdb: fake.MockDocDBClientCall{
					CreateDBClusterWithContext: []*fake.CallCreateDBClusterWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.CreateDBClusterInput{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
								AvailabilityZones: toStringPtrArray(
									testAvailabilityZone,
									testOtherAvailabilityZone,
								),
								BackupRetentionPeriod:       pointer.ToIntAsInt64(testBackupRetentionPeriod),
								DBClusterParameterGroupName: pointer.ToOrNilIfZeroValue(testDBClusterParameterGroupName),
								DBSubnetGroupName:           pointer.ToOrNilIfZeroValue(testDBSubnetGroupName),
								DeletionProtection:          pointer.ToOrNilIfZeroValue(true),
								EnableCloudwatchLogsExports: toStringPtrArray(
									testCloudWatchLog,
									testOtherCloudWatchLog,
								),
								Engine:                     pointer.ToOrNilIfZeroValue(testEngine),
								EngineVersion:              pointer.ToOrNilIfZeroValue(testEngineVersion),
								KmsKeyId:                   pointer.ToOrNilIfZeroValue(testKMSKeyID),
								MasterUsername:             pointer.ToOrNilIfZeroValue(testMasterUserName),
								MasterUserPassword:         pointer.ToOrNilIfZeroValue(testMasterUserPassword),
								Port:                       pointer.ToIntAsInt64(testPort),
								PreSignedUrl:               pointer.ToOrNilIfZeroValue(testPresignedURL),
								PreferredBackupWindow:      pointer.ToOrNilIfZeroValue(testPreferredBackupWindow),
								PreferredMaintenanceWindow: pointer.ToOrNilIfZeroValue(testPreferredMaintenanceWindow),
								StorageEncrypted:           pointer.ToOrNilIfZeroValue(true),
								VpcSecurityGroupIds: toStringPtrArray(
									testVpcSecurityGroup,
									testOtherVpcSecurityGroup,
								),
								Tags: []*docdb.Tag{
									{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testTagValue)},
									{Key: pointer.ToOrNilIfZeroValue(testOtherTagKey), Value: pointer.ToOrNilIfZeroValue(testOtherTagValue)},
								},
							},
						},
					},
				},
			},
		},
		"SuccessfulRestoreFromSnapshot": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						s := obj.(*v1.Secret)
						s.Data = map[string][]byte{
							testMasterPasswordSecretKey: []byte(testMasterUserPassword),
						}
						return nil
					},
				},
				docdb: &fake.MockDocDBClient{
					MockRestoreDBClusterFromSnapshotWithContext: func(c context.Context, cdpgi *docdb.RestoreDBClusterFromSnapshotInput, o []request.Option) (*docdb.RestoreDBClusterFromSnapshotOutput, error) {
						return &docdb.RestoreDBClusterFromSnapshotOutput{
							DBCluster: &docdb.DBCluster{
								AvailabilityZones: []*string{
									&testAvailabilityZone,
									&testOtherAvailabilityZone,
								},
								BackupRetentionPeriod:      pointer.ToIntAsInt64(testBackupRetentionPeriod),
								DBClusterParameterGroup:    &testDBClusterParameterGroupName,
								DBClusterIdentifier:        pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
								DBClusterArn:               pointer.ToOrNilIfZeroValue(testDBClusterArn),
								DeletionProtection:         pointer.ToOrNilIfZeroValue(true),
								Endpoint:                   pointer.ToOrNilIfZeroValue(testEndpoint),
								Engine:                     &testEngine,
								EngineVersion:              &testEngineVersion,
								KmsKeyId:                   &testKMSKeyID,
								MasterUsername:             &testMasterUserName,
								ReaderEndpoint:             pointer.ToOrNilIfZeroValue(testReaderEndpoint),
								Port:                       pointer.ToIntAsInt64(testPort),
								PreferredBackupWindow:      &testPreferredBackupWindow,
								PreferredMaintenanceWindow: &testPreferredMaintenanceWindow,
								StorageEncrypted:           pointer.ToOrNilIfZeroValue(true),
							},
						}, nil
					},
					MockCreateDBClusterWithContext: func(c context.Context, cdpgi *docdb.CreateDBClusterInput, o []request.Option) (*docdb.CreateDBClusterOutput, error) {
						return &docdb.CreateDBClusterOutput{
							DBCluster: &docdb.DBCluster{
								AvailabilityZones: []*string{
									&testAvailabilityZone,
									&testOtherAvailabilityZone,
								},
								BackupRetentionPeriod:      pointer.ToIntAsInt64(testBackupRetentionPeriod),
								DBClusterParameterGroup:    &testDBClusterParameterGroupName,
								DBClusterIdentifier:        pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
								DBClusterArn:               pointer.ToOrNilIfZeroValue(testDBClusterArn),
								DeletionProtection:         pointer.ToOrNilIfZeroValue(true),
								Endpoint:                   pointer.ToOrNilIfZeroValue(testEndpoint),
								Engine:                     &testEngine,
								EngineVersion:              &testEngineVersion,
								KmsKeyId:                   &testKMSKeyID,
								MasterUsername:             &testMasterUserName,
								ReaderEndpoint:             pointer.ToOrNilIfZeroValue(testReaderEndpoint),
								Port:                       pointer.ToIntAsInt64(testPort),
								PreferredBackupWindow:      &testPreferredBackupWindow,
								PreferredMaintenanceWindow: &testPreferredMaintenanceWindow,
								StorageEncrypted:           pointer.ToOrNilIfZeroValue(true),
							},
						}, nil
					},
				},
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withAvailabilityZones(
						testAvailabilityZone,
						testOtherAvailabilityZone,
					),
					withBackupRetentionPeriod(testBackupRetentionPeriod),
					withDBClusterParameterGroupName(testDBClusterParameterGroupName),
					withDBSubnetGroup(testDBSubnetGroupName),
					withDeletionProtection(true),
					withEnableCloudWatchLogsExports(
						testCloudWatchLog,
						testOtherCloudWatchLog,
					),
					withEngine(testEngine),
					withEngineVersion(testEngineVersion),
					withKmsKeyID(testKMSKeyID),
					withMasterUserName(testMasterUserName),
					withPort(testPort),
					withPreSignedURL(testPresignedURL),
					withPreferredBackupWindow(testPreferredBackupWindow),
					withPreferredMaintenanceWindow(testPreferredMaintenanceWindow),
					withStorageEncrypted(true),
					withTags(
						&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testTagValue)},
						&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testOtherTagKey), Value: pointer.ToOrNilIfZeroValue(testOtherTagValue)},
					),
					withVpcSecurityGroupIds(
						testVpcSecurityGroup,
						testOtherVpcSecurityGroup,
					),
					withMasterPasswordSecretRef(testMasterPasswordSecretNamespace, testMasterPasswordSecretName, testMasterPasswordSecretKey),
					withRestoreFromSnapshot(&svcapitypes.RestoreSnapshotConfiguration{
						SnapshotIdentifier: "abcd",
					}),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withConditions(xpv1.Creating()),
					withAvailabilityZones(
						testAvailabilityZone,
						testOtherAvailabilityZone,
					),
					withBackupRetentionPeriod(testBackupRetentionPeriod),
					withDBClusterParameterGroupName(testDBClusterParameterGroupName),
					withDBSubnetGroup(testDBSubnetGroupName),
					withStatusDBClusterArn(testDBClusterArn),
					withDeletionProtection(true),
					withEnableCloudWatchLogsExports(
						testCloudWatchLog,
						testOtherCloudWatchLog,
					),
					withEngine(testEngine),
					withEngineVersion(testEngineVersion),
					withKmsKeyID(testKMSKeyID),
					withMasterUserName(testMasterUserName),
					withPort(testPort),
					withPreSignedURL(testPresignedURL),
					withPreferredBackupWindow(testPreferredBackupWindow),
					withPreferredMaintenanceWindow(testPreferredMaintenanceWindow),
					withStorageEncrypted(true),
					withTags(
						&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testTagValue)},
						&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testOtherTagKey), Value: pointer.ToOrNilIfZeroValue(testOtherTagValue)},
					),
					withVpcSecurityGroupIds(
						testVpcSecurityGroup,
						testOtherVpcSecurityGroup,
					),
					withEndpoint(testEndpoint),
					withReaderEndpoint(testReaderEndpoint),
					withMasterPasswordSecretRef(testMasterPasswordSecretNamespace, testMasterPasswordSecretName, testMasterPasswordSecretKey),
					withStatusDBClusterParameterGroupName(testDBClusterParameterGroupName),
					withRestoreFromSnapshot(&svcapitypes.RestoreSnapshotConfiguration{
						SnapshotIdentifier: "abcd",
					}),
				),
				result: managed.ExternalCreation{
					ConnectionDetails: generateConnectionDetails(
						testMasterUserName,
						testMasterUserPassword,
						testReaderEndpoint,
						testEndpoint,
						testPort,
					),
				},
				docdb: fake.MockDocDBClientCall{
					RestoreDBClusterFromSnapshotWithContext: []*fake.CallRestoreDBClusterFromSnapshotWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.RestoreDBClusterFromSnapshotInput{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
								AvailabilityZones: toStringPtrArray(
									testAvailabilityZone,
									testOtherAvailabilityZone,
								),
								DBSubnetGroupName:  pointer.ToOrNilIfZeroValue(testDBSubnetGroupName),
								DeletionProtection: pointer.ToOrNilIfZeroValue(true),
								EnableCloudwatchLogsExports: toStringPtrArray(
									testCloudWatchLog,
									testOtherCloudWatchLog,
								),
								Engine:        pointer.ToOrNilIfZeroValue(testEngine),
								EngineVersion: pointer.ToOrNilIfZeroValue(testEngineVersion),
								KmsKeyId:      pointer.ToOrNilIfZeroValue(testKMSKeyID),
								Port:          pointer.ToIntAsInt64(testPort),
								VpcSecurityGroupIds: toStringPtrArray(
									testVpcSecurityGroup,
									testOtherVpcSecurityGroup,
								),
								Tags: []*docdb.Tag{
									{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testTagValue)},
									{Key: pointer.ToOrNilIfZeroValue(testOtherTagKey), Value: pointer.ToOrNilIfZeroValue(testOtherTagValue)},
								},
								SnapshotIdentifier: ptr.To("abcd"),
							},
						},
					},
					CreateDBClusterWithContext: []*fake.CallCreateDBClusterWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.CreateDBClusterInput{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
								AvailabilityZones: toStringPtrArray(
									testAvailabilityZone,
									testOtherAvailabilityZone,
								),
								BackupRetentionPeriod:       pointer.ToIntAsInt64(testBackupRetentionPeriod),
								DBClusterParameterGroupName: pointer.ToOrNilIfZeroValue(testDBClusterParameterGroupName),
								DBSubnetGroupName:           pointer.ToOrNilIfZeroValue(testDBSubnetGroupName),
								DeletionProtection:          pointer.ToOrNilIfZeroValue(true),
								EnableCloudwatchLogsExports: toStringPtrArray(
									testCloudWatchLog,
									testOtherCloudWatchLog,
								),
								Engine:                     pointer.ToOrNilIfZeroValue(testEngine),
								EngineVersion:              pointer.ToOrNilIfZeroValue(testEngineVersion),
								KmsKeyId:                   pointer.ToOrNilIfZeroValue(testKMSKeyID),
								MasterUsername:             pointer.ToOrNilIfZeroValue(testMasterUserName),
								MasterUserPassword:         pointer.ToOrNilIfZeroValue(testMasterUserPassword),
								Port:                       pointer.ToIntAsInt64(testPort),
								PreSignedUrl:               pointer.ToOrNilIfZeroValue(testPresignedURL),
								PreferredBackupWindow:      pointer.ToOrNilIfZeroValue(testPreferredBackupWindow),
								PreferredMaintenanceWindow: pointer.ToOrNilIfZeroValue(testPreferredMaintenanceWindow),
								StorageEncrypted:           pointer.ToOrNilIfZeroValue(true),
								VpcSecurityGroupIds: toStringPtrArray(
									testVpcSecurityGroup,
									testOtherVpcSecurityGroup,
								),
								Tags: []*docdb.Tag{
									{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testTagValue)},
									{Key: pointer.ToOrNilIfZeroValue(testOtherTagKey), Value: pointer.ToOrNilIfZeroValue(testOtherTagValue)},
								},
							},
						},
					},
				},
			},
		},
		"SuccessfulRestoreFromPointInTime": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						s := obj.(*v1.Secret)
						s.Data = map[string][]byte{
							testMasterPasswordSecretKey: []byte(testMasterUserPassword),
						}
						return nil
					},
				},
				docdb: &fake.MockDocDBClient{
					MockRestoreDBClusterToPointInTimeWithContext: func(c context.Context, cdpgi *docdb.RestoreDBClusterToPointInTimeInput, o []request.Option) (*docdb.RestoreDBClusterToPointInTimeOutput, error) {
						return &docdb.RestoreDBClusterToPointInTimeOutput{
							DBCluster: &docdb.DBCluster{
								AvailabilityZones: []*string{
									&testAvailabilityZone,
									&testOtherAvailabilityZone,
								},
								BackupRetentionPeriod:      pointer.ToIntAsInt64(testBackupRetentionPeriod),
								DBClusterParameterGroup:    &testDBClusterParameterGroupName,
								DBClusterIdentifier:        pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
								DBClusterArn:               pointer.ToOrNilIfZeroValue(testDBClusterArn),
								DeletionProtection:         pointer.ToOrNilIfZeroValue(true),
								Endpoint:                   pointer.ToOrNilIfZeroValue(testEndpoint),
								Engine:                     &testEngine,
								EngineVersion:              &testEngineVersion,
								KmsKeyId:                   &testKMSKeyID,
								MasterUsername:             &testMasterUserName,
								ReaderEndpoint:             pointer.ToOrNilIfZeroValue(testReaderEndpoint),
								Port:                       pointer.ToIntAsInt64(testPort),
								PreferredBackupWindow:      &testPreferredBackupWindow,
								PreferredMaintenanceWindow: &testPreferredMaintenanceWindow,
								StorageEncrypted:           pointer.ToOrNilIfZeroValue(true),
							},
						}, nil
					},
					MockCreateDBClusterWithContext: func(c context.Context, cdpgi *docdb.CreateDBClusterInput, o []request.Option) (*docdb.CreateDBClusterOutput, error) {
						return &docdb.CreateDBClusterOutput{
							DBCluster: &docdb.DBCluster{
								AvailabilityZones: []*string{
									&testAvailabilityZone,
									&testOtherAvailabilityZone,
								},
								BackupRetentionPeriod:      pointer.ToIntAsInt64(testBackupRetentionPeriod),
								DBClusterParameterGroup:    &testDBClusterParameterGroupName,
								DBClusterIdentifier:        pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
								DBClusterArn:               pointer.ToOrNilIfZeroValue(testDBClusterArn),
								DeletionProtection:         pointer.ToOrNilIfZeroValue(true),
								Endpoint:                   pointer.ToOrNilIfZeroValue(testEndpoint),
								Engine:                     &testEngine,
								EngineVersion:              &testEngineVersion,
								KmsKeyId:                   &testKMSKeyID,
								MasterUsername:             &testMasterUserName,
								ReaderEndpoint:             pointer.ToOrNilIfZeroValue(testReaderEndpoint),
								Port:                       pointer.ToIntAsInt64(testPort),
								PreferredBackupWindow:      &testPreferredBackupWindow,
								PreferredMaintenanceWindow: &testPreferredMaintenanceWindow,
								StorageEncrypted:           pointer.ToOrNilIfZeroValue(true),
							},
						}, nil
					},
				},
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withAvailabilityZones(
						testAvailabilityZone,
						testOtherAvailabilityZone,
					),
					withBackupRetentionPeriod(testBackupRetentionPeriod),
					withDBClusterParameterGroupName(testDBClusterParameterGroupName),
					withDBSubnetGroup(testDBSubnetGroupName),
					withDeletionProtection(true),
					withEnableCloudWatchLogsExports(
						testCloudWatchLog,
						testOtherCloudWatchLog,
					),
					withEngine(testEngine),
					withEngineVersion(testEngineVersion),
					withKmsKeyID(testKMSKeyID),
					withMasterUserName(testMasterUserName),
					withPort(testPort),
					withPreSignedURL(testPresignedURL),
					withPreferredBackupWindow(testPreferredBackupWindow),
					withPreferredMaintenanceWindow(testPreferredMaintenanceWindow),
					withStorageEncrypted(true),
					withTags(
						&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testTagValue)},
						&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testOtherTagKey), Value: pointer.ToOrNilIfZeroValue(testOtherTagValue)},
					),
					withVpcSecurityGroupIds(
						testVpcSecurityGroup,
						testOtherVpcSecurityGroup,
					),
					withMasterPasswordSecretRef(testMasterPasswordSecretNamespace, testMasterPasswordSecretName, testMasterPasswordSecretKey),
					withRestoreToPointInTime(&svcapitypes.RestorePointInTimeConfiguration{
						RestoreTime:               &metav1.Time{Time: timeNow},
						UseLatestRestorableTime:   ptr.To(true),
						SourceDBClusterIdentifier: "abcd",
						RestoreType:               ptr.To("full-copy"),
					}),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withConditions(xpv1.Creating()),
					withAvailabilityZones(
						testAvailabilityZone,
						testOtherAvailabilityZone,
					),
					withBackupRetentionPeriod(testBackupRetentionPeriod),
					withDBClusterParameterGroupName(testDBClusterParameterGroupName),
					withDBSubnetGroup(testDBSubnetGroupName),
					withStatusDBClusterArn(testDBClusterArn),
					withDeletionProtection(true),
					withEnableCloudWatchLogsExports(
						testCloudWatchLog,
						testOtherCloudWatchLog,
					),
					withEngine(testEngine),
					withEngineVersion(testEngineVersion),
					withKmsKeyID(testKMSKeyID),
					withMasterUserName(testMasterUserName),
					withPort(testPort),
					withPreSignedURL(testPresignedURL),
					withPreferredBackupWindow(testPreferredBackupWindow),
					withPreferredMaintenanceWindow(testPreferredMaintenanceWindow),
					withStorageEncrypted(true),
					withTags(
						&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testTagValue)},
						&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testOtherTagKey), Value: pointer.ToOrNilIfZeroValue(testOtherTagValue)},
					),
					withVpcSecurityGroupIds(
						testVpcSecurityGroup,
						testOtherVpcSecurityGroup,
					),
					withEndpoint(testEndpoint),
					withReaderEndpoint(testReaderEndpoint),
					withMasterPasswordSecretRef(testMasterPasswordSecretNamespace, testMasterPasswordSecretName, testMasterPasswordSecretKey),
					withStatusDBClusterParameterGroupName(testDBClusterParameterGroupName),
					withRestoreToPointInTime(&svcapitypes.RestorePointInTimeConfiguration{
						RestoreTime:               &metav1.Time{Time: timeNow},
						UseLatestRestorableTime:   ptr.To(true),
						SourceDBClusterIdentifier: "abcd",
						RestoreType:               ptr.To("full-copy"),
					}),
				),
				result: managed.ExternalCreation{
					ConnectionDetails: generateConnectionDetails(
						testMasterUserName,
						testMasterUserPassword,
						testReaderEndpoint,
						testEndpoint,
						testPort,
					),
				},
				docdb: fake.MockDocDBClientCall{
					RestoreDBClusterToPointInTimeWithContext: []*fake.CallRestoreDBClusterToPointInTimeWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.RestoreDBClusterToPointInTimeInput{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
								DBSubnetGroupName:   pointer.ToOrNilIfZeroValue(testDBSubnetGroupName),
								DeletionProtection:  pointer.ToOrNilIfZeroValue(true),
								EnableCloudwatchLogsExports: toStringPtrArray(
									testCloudWatchLog,
									testOtherCloudWatchLog,
								),
								KmsKeyId: pointer.ToOrNilIfZeroValue(testKMSKeyID),
								Port:     pointer.ToIntAsInt64(testPort),
								VpcSecurityGroupIds: toStringPtrArray(
									testVpcSecurityGroup,
									testOtherVpcSecurityGroup,
								),
								Tags: []*docdb.Tag{
									{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testTagValue)},
									{Key: pointer.ToOrNilIfZeroValue(testOtherTagKey), Value: pointer.ToOrNilIfZeroValue(testOtherTagValue)},
								},
								RestoreToTime:             &timeNow,
								UseLatestRestorableTime:   ptr.To(true),
								SourceDBClusterIdentifier: ptr.To("abcd"),
								RestoreType:               ptr.To("full-copy"),
							},
						},
					},
					CreateDBClusterWithContext: []*fake.CallCreateDBClusterWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.CreateDBClusterInput{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
								AvailabilityZones: toStringPtrArray(
									testAvailabilityZone,
									testOtherAvailabilityZone,
								),
								BackupRetentionPeriod:       pointer.ToIntAsInt64(testBackupRetentionPeriod),
								DBClusterParameterGroupName: pointer.ToOrNilIfZeroValue(testDBClusterParameterGroupName),
								DBSubnetGroupName:           pointer.ToOrNilIfZeroValue(testDBSubnetGroupName),
								DeletionProtection:          pointer.ToOrNilIfZeroValue(true),
								EnableCloudwatchLogsExports: toStringPtrArray(
									testCloudWatchLog,
									testOtherCloudWatchLog,
								),
								Engine:                     pointer.ToOrNilIfZeroValue(testEngine),
								EngineVersion:              pointer.ToOrNilIfZeroValue(testEngineVersion),
								KmsKeyId:                   pointer.ToOrNilIfZeroValue(testKMSKeyID),
								MasterUsername:             pointer.ToOrNilIfZeroValue(testMasterUserName),
								MasterUserPassword:         pointer.ToOrNilIfZeroValue(testMasterUserPassword),
								Port:                       pointer.ToIntAsInt64(testPort),
								PreSignedUrl:               pointer.ToOrNilIfZeroValue(testPresignedURL),
								PreferredBackupWindow:      pointer.ToOrNilIfZeroValue(testPreferredBackupWindow),
								PreferredMaintenanceWindow: pointer.ToOrNilIfZeroValue(testPreferredMaintenanceWindow),
								StorageEncrypted:           pointer.ToOrNilIfZeroValue(true),
								VpcSecurityGroupIds: toStringPtrArray(
									testVpcSecurityGroup,
									testOtherVpcSecurityGroup,
								),
								Tags: []*docdb.Tag{
									{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testTagValue)},
									{Key: pointer.ToOrNilIfZeroValue(testOtherTagKey), Value: pointer.ToOrNilIfZeroValue(testOtherTagValue)},
								},
							},
						},
					},
				},
			},
		},
		"ErrorCreate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockCreateDBClusterWithContext: func(c context.Context, cdpgi *docdb.CreateDBClusterInput, o []request.Option) (*docdb.CreateDBClusterOutput, error) {
						return &docdb.CreateDBClusterOutput{}, errors.New(testErrCreateDBClusterFailed)
					},
				},
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withConditions(xpv1.Creating()),
				),
				result: managed.ExternalCreation{},
				err:    errors.Wrap(errors.New(testErrCreateDBClusterFailed), errCreate),
				docdb: fake.MockDocDBClientCall{
					CreateDBClusterWithContext: []*fake.CallCreateDBClusterWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.CreateDBClusterInput{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
							},
						},
					},
				},
			},
		},
		"ErrorCreate_masterPassword_secret_not_found": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						return errors.New(testErrGetSecret)
					},
				},
				docdb: &fake.MockDocDBClient{
					MockCreateDBClusterWithContext: func(c context.Context, cdpgi *docdb.CreateDBClusterInput, o []request.Option) (*docdb.CreateDBClusterOutput, error) {
						return &docdb.CreateDBClusterOutput{
							DBCluster: &docdb.DBCluster{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
								DBClusterArn:        pointer.ToOrNilIfZeroValue(testDBClusterArn),
								Endpoint:            pointer.ToOrNilIfZeroValue(testEndpoint),
								ReaderEndpoint:      pointer.ToOrNilIfZeroValue(testReaderEndpoint),
								Port:                pointer.ToIntAsInt64(testPort),
							},
						}, nil
					},
				},
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withMasterPasswordSecretRef(testMasterPasswordSecretNamespace, testMasterPasswordSecretName, testMasterPasswordSecretKey),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withConditions(xpv1.Creating()),
					withMasterPasswordSecretRef(testMasterPasswordSecretNamespace, testMasterPasswordSecretName, testMasterPasswordSecretKey),
				),
				result: managed.ExternalCreation{},
				err:    errors.Wrap(errors.Wrap(errors.New(testErrGetSecret), errGetPasswordSecretFailed), "pre-create failed"),
				docdb:  fake.MockDocDBClientCall{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			opts := []option{setupExternal}
			e := newExternal(tc.args.kube, tc.args.docdb, opts)
			o, err := e.Create(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.docdb, tc.args.docdb.Called, cmpopts.IgnoreInterfaces(struct{ context.Context }{})); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type want struct {
		cr    *svcapitypes.DBCluster
		err   error
		docdb fake.MockDocDBClientCall
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulDelete": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDeleteDBClusterWithContext: func(c context.Context, ddpgi *docdb.DeleteDBClusterInput, o []request.Option) (*docdb.DeleteDBClusterOutput, error) {
						return &docdb.DeleteDBClusterOutput{}, nil
					},
				},
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withFinalDBSnapshotIdentifier(testFinalDBSnapshotIdentifier),
					withSkipFinalSnapshot(true),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withConditions(xpv1.Deleting()),
					withFinalDBSnapshotIdentifier(testFinalDBSnapshotIdentifier),
					withSkipFinalSnapshot(true),
				),
				docdb: fake.MockDocDBClientCall{
					DeleteDBClusterWithContext: []*fake.CallDeleteDBClusterWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DeleteDBClusterInput{
								DBClusterIdentifier:       pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
								FinalDBSnapshotIdentifier: pointer.ToOrNilIfZeroValue(testFinalDBSnapshotIdentifier),
								SkipFinalSnapshot:         pointer.ToOrNilIfZeroValue(true),
							},
						},
					},
				},
			},
		},
		"ErrorDelete": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDeleteDBClusterWithContext: func(c context.Context, ddpgi *docdb.DeleteDBClusterInput, o []request.Option) (*docdb.DeleteDBClusterOutput, error) {
						return &docdb.DeleteDBClusterOutput{}, errors.New(testErrDeleteDBClusterFailed)
					},
				},
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withConditions(xpv1.Deleting()),
				),
				err: errors.Wrap(errors.New(testErrDeleteDBClusterFailed), errDelete),
				docdb: fake.MockDocDBClientCall{
					DeleteDBClusterWithContext: []*fake.CallDeleteDBClusterWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DeleteDBClusterInput{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
							},
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			opts := []option{setupExternal}
			e := newExternal(tc.args.kube, tc.args.docdb, opts)
			err := e.Delete(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.docdb, tc.args.docdb.Called, cmpopts.IgnoreInterfaces(struct{ context.Context }{})); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	type want struct {
		cr     *svcapitypes.DBCluster
		err    error
		result managed.ExternalUpdate
		docdb  fake.MockDocDBClientCall
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulUpdate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockModifyDBClusterWithContext: func(c context.Context, mdpgi *docdb.ModifyDBClusterInput, o []request.Option) (*docdb.ModifyDBClusterOutput, error) {
						return &docdb.ModifyDBClusterOutput{
							DBCluster: &docdb.DBCluster{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
								DBClusterArn:        pointer.ToOrNilIfZeroValue(testDBClusterArn),
							},
						}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{
							TagList: []*docdb.Tag{
								{
									Key:   pointer.ToOrNilIfZeroValue(testOtherOtherTagKey),
									Value: pointer.ToOrNilIfZeroValue(testOtherOtherTagValue),
								},
							},
						}, nil
					},
					MockAddTagsToResource: func(attri *docdb.AddTagsToResourceInput) (*docdb.AddTagsToResourceOutput, error) {
						return &docdb.AddTagsToResourceOutput{}, nil
					},
					MockRemoveTagsFromResource: func(rtfri *docdb.RemoveTagsFromResourceInput) (*docdb.RemoveTagsFromResourceOutput, error) {
						return &docdb.RemoveTagsFromResourceOutput{}, nil
					},
				},
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withAvailabilityZones(
						testAvailabilityZone,
						testOtherAvailabilityZone,
					),
					withBackupRetentionPeriod(testBackupRetentionPeriod),
					withDBClusterParameterGroupName(testDBClusterParameterGroupName),
					withDBSubnetGroup(testDBSubnetGroupName),
					withDeletionProtection(true),
					withEngine(testEngine),
					withEngineVersion(testEngineVersion),
					withKmsKeyID(testKMSKeyID),
					withMasterUserName(testMasterUserName),
					withPort(testPort),
					withPreSignedURL(testPresignedURL),
					withPreferredBackupWindow(testPreferredBackupWindow),
					withPreferredMaintenanceWindow(testPreferredMaintenanceWindow),
					withStorageEncrypted(true),
					withTags(
						&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testTagValue)},
						&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testOtherTagKey), Value: pointer.ToOrNilIfZeroValue(testOtherTagValue)},
					),
					withVpcSecurityGroupIds(
						testVpcSecurityGroup,
						testOtherVpcSecurityGroup,
					),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withAvailabilityZones(
						testAvailabilityZone,
						testOtherAvailabilityZone,
					),
					withBackupRetentionPeriod(testBackupRetentionPeriod),
					withDBClusterParameterGroupName(testDBClusterParameterGroupName),
					withDBSubnetGroup(testDBSubnetGroupName),
					withDeletionProtection(true),
					withEngine(testEngine),
					withEngineVersion(testEngineVersion),
					withKmsKeyID(testKMSKeyID),
					withMasterUserName(testMasterUserName),
					withPort(testPort),
					withPreSignedURL(testPresignedURL),
					withPreferredBackupWindow(testPreferredBackupWindow),
					withPreferredMaintenanceWindow(testPreferredMaintenanceWindow),
					withStorageEncrypted(true),
					withTags(
						&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testTagValue)},
						&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testOtherTagKey), Value: pointer.ToOrNilIfZeroValue(testOtherTagValue)},
					),
					withVpcSecurityGroupIds(
						testVpcSecurityGroup,
						testOtherVpcSecurityGroup,
					),
				),
				docdb: fake.MockDocDBClientCall{
					ModifyDBClusterWithContext: []*fake.CallModifyDBClusterWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.ModifyDBClusterInput{
								DBClusterIdentifier:         pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
								BackupRetentionPeriod:       pointer.ToIntAsInt64(testBackupRetentionPeriod),
								DBClusterParameterGroupName: pointer.ToOrNilIfZeroValue(testDBClusterParameterGroupName),
								DeletionProtection:          pointer.ToOrNilIfZeroValue(true),
								EngineVersion:               pointer.ToOrNilIfZeroValue(testEngineVersion),
								Port:                        pointer.ToIntAsInt64(testPort),
								PreferredBackupWindow:       pointer.ToOrNilIfZeroValue(testPreferredBackupWindow),
								PreferredMaintenanceWindow:  pointer.ToOrNilIfZeroValue(testPreferredMaintenanceWindow),
								VpcSecurityGroupIds: toStringPtrArray(
									testVpcSecurityGroup,
									testOtherVpcSecurityGroup,
								),
								CloudwatchLogsExportConfiguration: &docdb.CloudwatchLogsExportConfiguration{
									DisableLogTypes: []*string{},
									EnableLogTypes:  []*string{},
								},
							},
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{
								ResourceName: pointer.ToOrNilIfZeroValue(testDBClusterArn),
							},
						},
					},
					AddTagsToResource: []*fake.CallAddTagsToResource{
						{
							I: &docdb.AddTagsToResourceInput{
								ResourceName: pointer.ToOrNilIfZeroValue(testDBClusterArn),
								Tags: []*docdb.Tag{
									{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testTagValue)},
									{Key: pointer.ToOrNilIfZeroValue(testOtherTagKey), Value: pointer.ToOrNilIfZeroValue(testOtherTagValue)},
								},
							},
						},
					},
					RemoveTagsFromResource: []*fake.CallRemoveTagsFromResource{
						{
							I: &docdb.RemoveTagsFromResourceInput{
								ResourceName: pointer.ToOrNilIfZeroValue(testDBClusterArn),
								TagKeys: []*string{
									pointer.ToOrNilIfZeroValue(testOtherOtherTagKey),
								},
							},
						},
					},
				},
			},
		},
		"SuccessfulUpdate_enable_CloudwatchLogsExportConfiguration": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockModifyDBClusterWithContext: func(c context.Context, mdpgi *docdb.ModifyDBClusterInput, o []request.Option) (*docdb.ModifyDBClusterOutput, error) {
						return &docdb.ModifyDBClusterOutput{
							DBCluster: &docdb.DBCluster{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
								DBClusterArn:        pointer.ToOrNilIfZeroValue(testDBClusterArn),
							},
						}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{
							TagList: []*docdb.Tag{},
						}, nil
					},
				},
				cr: instance(
					withDBClusterIdentifier(testOtherDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withEnableCloudWatchLogsExports(
						testCloudWatchLog,
						testOtherCloudWatchLog,
					),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testOtherDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withEnableCloudWatchLogsExports(
						testCloudWatchLog,
						testOtherCloudWatchLog,
					),
				),
				docdb: fake.MockDocDBClientCall{
					ModifyDBClusterWithContext: []*fake.CallModifyDBClusterWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.ModifyDBClusterInput{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
								CloudwatchLogsExportConfiguration: &docdb.CloudwatchLogsExportConfiguration{
									DisableLogTypes: []*string{},
									EnableLogTypes: toStringPtrArray(
										testCloudWatchLog,
										testOtherCloudWatchLog,
									),
								},
							},
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{
								ResourceName: pointer.ToOrNilIfZeroValue(testDBClusterArn),
							},
						},
					},
				},
			},
		},
		"SuccessfulUpdate_disable_CloudwatchLogsExportConfiguration": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockModifyDBClusterWithContext: func(c context.Context, mdpgi *docdb.ModifyDBClusterInput, o []request.Option) (*docdb.ModifyDBClusterOutput, error) {
						return &docdb.ModifyDBClusterOutput{
							DBCluster: &docdb.DBCluster{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
								DBClusterArn:        pointer.ToOrNilIfZeroValue(testDBClusterArn),
							},
						}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{
							TagList: []*docdb.Tag{},
						}, nil
					},
				},
				cr: instance(
					withDBClusterIdentifier(testOtherDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withEnableCloudWatchLogsExports(),
					withStatusEnableCloudWatchLogsExports(
						testCloudWatchLog,
					),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testOtherDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withEnableCloudWatchLogsExports(),
					withStatusEnableCloudWatchLogsExports(
						testCloudWatchLog,
					),
				),
				docdb: fake.MockDocDBClientCall{
					ModifyDBClusterWithContext: []*fake.CallModifyDBClusterWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.ModifyDBClusterInput{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
								CloudwatchLogsExportConfiguration: &docdb.CloudwatchLogsExportConfiguration{
									DisableLogTypes: toStringPtrArray(
										testCloudWatchLog,
									),
									EnableLogTypes: toStringPtrArray(),
								},
							},
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{
								ResourceName: pointer.ToOrNilIfZeroValue(testDBClusterArn),
							},
						},
					},
				},
			},
		},
		"SuccessfulUpdate_enable_and_disable_CloudwatchLogsExportConfiguration": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockModifyDBClusterWithContext: func(c context.Context, mdpgi *docdb.ModifyDBClusterInput, o []request.Option) (*docdb.ModifyDBClusterOutput, error) {
						return &docdb.ModifyDBClusterOutput{
							DBCluster: &docdb.DBCluster{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
								DBClusterArn:        pointer.ToOrNilIfZeroValue(testDBClusterArn),
							},
						}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{
							TagList: []*docdb.Tag{},
						}, nil
					},
				},
				cr: instance(
					withDBClusterIdentifier(testOtherDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withEnableCloudWatchLogsExports(
						testOtherCloudWatchLog,
					),
					withStatusEnableCloudWatchLogsExports(
						testCloudWatchLog,
					),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testOtherDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withEnableCloudWatchLogsExports(
						testOtherCloudWatchLog,
					),
					withStatusEnableCloudWatchLogsExports(
						testCloudWatchLog,
					),
				),
				docdb: fake.MockDocDBClientCall{
					ModifyDBClusterWithContext: []*fake.CallModifyDBClusterWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.ModifyDBClusterInput{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
								CloudwatchLogsExportConfiguration: &docdb.CloudwatchLogsExportConfiguration{
									DisableLogTypes: toStringPtrArray(
										testCloudWatchLog,
									),
									EnableLogTypes: toStringPtrArray(
										testOtherCloudWatchLog,
									),
								},
							},
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{
								ResourceName: pointer.ToOrNilIfZeroValue(testDBClusterArn),
							},
						},
					},
				},
			},
		},
		"ErrorUpdate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockModifyDBClusterWithContext: func(c context.Context, mdpgi *docdb.ModifyDBClusterInput, o []request.Option) (*docdb.ModifyDBClusterOutput, error) {
						return &docdb.ModifyDBClusterOutput{}, errors.New(testErrModifyDBClusterFailed)
					},
				},
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
				),
				err: errors.Wrap(errors.New(testErrModifyDBClusterFailed), errUpdate),
				docdb: fake.MockDocDBClientCall{
					ModifyDBClusterWithContext: []*fake.CallModifyDBClusterWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.ModifyDBClusterInput{
								DBClusterIdentifier: pointer.ToOrNilIfZeroValue(testDBClusterIdentifier),
								CloudwatchLogsExportConfiguration: &docdb.CloudwatchLogsExportConfiguration{
									DisableLogTypes: []*string{},
									EnableLogTypes:  []*string{},
								},
							},
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			opts := []option{setupExternal}
			e := newExternal(tc.args.kube, tc.args.docdb, opts)
			o, err := e.Update(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.docdb, tc.args.docdb.Called, cmpopts.IgnoreInterfaces(struct{ context.Context }{})); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
