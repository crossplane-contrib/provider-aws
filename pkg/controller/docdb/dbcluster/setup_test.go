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

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/docdb"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	svcapitypes "github.com/crossplane/provider-aws/apis/docdb/v1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/docdb/fake"
	svcutils "github.com/crossplane/provider-aws/pkg/controller/docdb"
)

const (
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
	testOtherDBSubnetGroupName           = "some-other-db-subnet-group"
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
	testErrBoom                     = "boom"
	testErrGetSecret                = "testErrGetSecret"
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
		ptrArr[i] = awsclient.String(s)
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

func withDBClusterIdentifier(value string) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Status.AtProvider.DBClusterIdentifier = awsclient.String(value)
	}
}

func withConditions(value ...xpv1.Condition) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Status.SetConditions(value...)
	}
}

func withStatus(value string) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Status.AtProvider.Status = awsclient.String(value)
	}
}

func withAvailabilityZones(values ...string) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Spec.ForProvider.AvailabilityZones = toStringPtrArray(values...)
	}
}

func withBackupRetentionPeriod(value int) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Spec.ForProvider.BackupRetentionPeriod = awsclient.Int64(value)
	}
}

func withDBClusterParameterGroupName(value string) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Spec.ForProvider.DBClusterParameterGroupName = awsclient.String(value)
	}
}

func withStatusDBClusterParameterGroupName(value string) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Status.AtProvider.DBClusterParameterGroup = awsclient.String(value)
	}
}

func withStatusDBClusterArn(value string) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Status.AtProvider.DBClusterARN = awsclient.String(value)
	}
}

func withDBSubnetGroup(value string) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Spec.ForProvider.DBSubnetGroupName = awsclient.String(value)
	}
}

func withStatusDBSubnetGroup(value string) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Status.AtProvider.DBSubnetGroup = awsclient.String(value)
	}
}

func withDeletionProtection(value bool) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Spec.ForProvider.DeletionProtection = awsclient.Bool(value, awsclient.FieldRequired)
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
		o.Spec.ForProvider.Engine = awsclient.String(value)
	}
}

func withEngineVersion(value string) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Spec.ForProvider.EngineVersion = awsclient.String(value)
	}
}

func withMasterUserName(value string) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Spec.ForProvider.MasterUsername = awsclient.String(value)
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
		o.Spec.ForProvider.KMSKeyID = awsclient.String(value)
	}
}

func withPort(value int) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Spec.ForProvider.Port = awsclient.Int64(value)
	}
}

func withEndpoint(value string) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Status.AtProvider.Endpoint = awsclient.String(value)
	}
}

func withReaderEndpoint(value string) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Status.AtProvider.ReaderEndpoint = awsclient.String(value)
	}
}

func withPreSignedURL(value string) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Spec.ForProvider.PreSignedURL = awsclient.String(value)
	}
}

func withPreferredBackupWindow(value string) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Spec.ForProvider.PreferredBackupWindow = awsclient.String(value)
	}
}

func withPreferredMaintenanceWindow(value string) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Spec.ForProvider.PreferredMaintenanceWindow = awsclient.String(value)
	}
}

func withStorageEncrypted(value bool) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Spec.ForProvider.StorageEncrypted = awsclient.Bool(value)
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
		o.Spec.ForProvider.SkipFinalSnapshot = awsclient.Bool(value)
	}
}

func withFinalDBSnapshotIdentifier(value string) docDBModifier {
	return func(o *svcapitypes.DBCluster) {
		o.Spec.ForProvider.FinalDBSnapshotIdentifier = awsclient.String(value)
	}
}

func mergeTags(lists ...[]*svcapitypes.Tag) []*svcapitypes.Tag {
	res := []*svcapitypes.Tag{}
	for _, list := range lists {
		res = append(res, list...)
	}
	return res
}

func generateConnectionDetails(username, password, readerEndpoint, endpoint string, port int) managed.ConnectionDetails {
	return managed.ConnectionDetails{
		xpv1.ResourceCredentialsSecretEndpointKey: []byte(endpoint),
		xpv1.ResourceCredentialsSecretPortKey:     []byte(strconv.Itoa(port)),
		xpv1.ResourceCredentialsSecretUserKey:     []byte(username),
		xpv1.ResourceCredentialsSecretPasswordKey: []byte(password),
		"readerEndpoint":                          []byte(readerEndpoint),
	}
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
									DBClusterIdentifier:   awsclient.String(testDBClusterIdentifier),
									Status:                awsclient.String(svcapitypes.DocDBInstanceStateAvailable),
									BackupRetentionPeriod: awsclient.Int64(testBackupRetentionPeriod),
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
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
									DBClusterIdentifier:   awsclient.String(testDBClusterIdentifier),
									Status:                awsclient.String(svcapitypes.DocDBInstanceStateAvailable),
									BackupRetentionPeriod: awsclient.Int64(testBackupRetentionPeriod),
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
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
									DBClusterIdentifier:   awsclient.String(testDBClusterIdentifier),
									Status:                awsclient.String(svcapitypes.DocDBInstanceStateAvailable),
									BackupRetentionPeriod: awsclient.Int64(testBackupRetentionPeriod),
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
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
									DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
									Status:              awsclient.String(svcapitypes.DocDBInstanceStateAvailable),
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
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
									DBClusterIdentifier:     awsclient.String(testDBClusterIdentifier),
									Status:                  awsclient.String(svcapitypes.DocDBInstanceStateAvailable),
									DBClusterParameterGroup: awsclient.String(testDBClusterParameterGroupName),
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
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
									DBClusterIdentifier:     awsclient.String(testDBClusterIdentifier),
									Status:                  awsclient.String(svcapitypes.DocDBInstanceStateAvailable),
									DBClusterParameterGroup: awsclient.String(testDBClusterParameterGroupName),
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
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
									DBClusterIdentifier:     awsclient.String(testDBClusterIdentifier),
									Status:                  awsclient.String(svcapitypes.DocDBInstanceStateAvailable),
									DBClusterParameterGroup: awsclient.String(testDBClusterParameterGroupName),
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
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
		"AvailableState_and_changed_DBSubnetGroupName_should_not_be_UpToDate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBClustersWithContext: func(c context.Context, ddi *docdb.DescribeDBClustersInput, o []request.Option) (*docdb.DescribeDBClustersOutput, error) {
						return &docdb.DescribeDBClustersOutput{
							DBClusters: []*docdb.DBCluster{
								{
									DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
									Status:              awsclient.String(svcapitypes.DocDBInstanceStateAvailable),
									DBSubnetGroup:       awsclient.String(testDBSubnetGroupName),
								},
							},
						}, nil
					},
				},
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withDBSubnetGroup(testOtherDBSubnetGroupName),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withConditions(xpv1.Available()),
					withStatus(svcapitypes.DocDBInstanceStateAvailable),
					withDBSubnetGroup(testOtherDBSubnetGroupName),
					withStatusDBSubnetGroup(testDBSubnetGroupName),
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
							},
						},
					},
				},
			},
		},
		"AvailableState_and_same_DBSubnetGroupName_should_be_UpToDate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBClustersWithContext: func(c context.Context, ddi *docdb.DescribeDBClustersInput, o []request.Option) (*docdb.DescribeDBClustersOutput, error) {
						return &docdb.DescribeDBClustersOutput{
							DBClusters: []*docdb.DBCluster{
								{
									DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
									Status:              awsclient.String(svcapitypes.DocDBInstanceStateAvailable),
									DBSubnetGroup:       awsclient.String(testDBSubnetGroupName),
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
					withDBSubnetGroup(testDBSubnetGroupName),
				),
			},
			want: want{
				cr: instance(
					withDBClusterIdentifier(testDBClusterIdentifier),
					withExternalName(testDBClusterIdentifier),
					withConditions(xpv1.Available()),
					withStatus(svcapitypes.DocDBInstanceStateAvailable),
					withDBSubnetGroup(testDBSubnetGroupName),
					withStatusDBSubnetGroup(testDBSubnetGroupName),
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
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
		"AvailableState_and_no_spec_DBSubnetGroupName_should_be_UpToDate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBClustersWithContext: func(c context.Context, ddi *docdb.DescribeDBClustersInput, o []request.Option) (*docdb.DescribeDBClustersOutput, error) {
						return &docdb.DescribeDBClustersOutput{
							DBClusters: []*docdb.DBCluster{
								{
									DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
									Status:              awsclient.String(svcapitypes.DocDBInstanceStateAvailable),
									DBSubnetGroup:       awsclient.String(testDBSubnetGroupName),
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
					withDBSubnetGroup(testDBSubnetGroupName),
					withExternalName(testDBClusterIdentifier),
					withConditions(xpv1.Available()),
					withStatus(svcapitypes.DocDBInstanceStateAvailable),
					withStatusDBSubnetGroup(testDBSubnetGroupName),
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
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
									DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
									Status:              awsclient.String(svcapitypes.DocDBInstanceStateAvailable),
									DeletionProtection:  awsclient.Bool(true),
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
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
									DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
									Status:              awsclient.String(svcapitypes.DocDBInstanceStateAvailable),
									DeletionProtection:  awsclient.Bool(true),
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
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
									DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
									Status:              awsclient.String(svcapitypes.DocDBInstanceStateAvailable),
									DeletionProtection:  awsclient.Bool(true),
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
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
									DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
									Status:              awsclient.String(svcapitypes.DocDBInstanceStateAvailable),
									EnabledCloudwatchLogsExports: []*string{
										awsclient.String(testCloudWatchLog),
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
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
									DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
									Status:              awsclient.String(svcapitypes.DocDBInstanceStateAvailable),
									EnabledCloudwatchLogsExports: []*string{
										awsclient.String(testCloudWatchLog),
										awsclient.String(testOtherCloudWatchLog),
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
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
									DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
									Status:              awsclient.String(svcapitypes.DocDBInstanceStateAvailable),
									EnabledCloudwatchLogsExports: []*string{
										awsclient.String(testCloudWatchLog),
										awsclient.String(testOtherCloudWatchLog),
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
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
									DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
									Status:              awsclient.String(svcapitypes.DocDBInstanceStateAvailable),
									EnabledCloudwatchLogsExports: []*string{
										awsclient.String(testCloudWatchLog),
										awsclient.String(testOtherCloudWatchLog),
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
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
									DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
									Status:              awsclient.String(svcapitypes.DocDBInstanceStateAvailable),
									EnabledCloudwatchLogsExports: []*string{
										awsclient.String(testCloudWatchLog),
										awsclient.String(testOtherCloudWatchLog),
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
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
									DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
									Status:              awsclient.String(svcapitypes.DocDBInstanceStateAvailable),
									Port:                awsclient.Int64(testPort),
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
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
									DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
									Status:              awsclient.String(svcapitypes.DocDBInstanceStateAvailable),
									Port:                awsclient.Int64(testPort),
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
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
									DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
									Status:              awsclient.String(svcapitypes.DocDBInstanceStateAvailable),
									Port:                awsclient.Int64(testPort),
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
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
									DBClusterIdentifier:   awsclient.String(testDBClusterIdentifier),
									Status:                awsclient.String(svcapitypes.DocDBInstanceStateAvailable),
									PreferredBackupWindow: awsclient.String(testPreferredBackupWindow),
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
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
									DBClusterIdentifier:   awsclient.String(testDBClusterIdentifier),
									Status:                awsclient.String(svcapitypes.DocDBInstanceStateAvailable),
									PreferredBackupWindow: awsclient.String(testPreferredBackupWindow),
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
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
									DBClusterIdentifier:   awsclient.String(testDBClusterIdentifier),
									Status:                awsclient.String(svcapitypes.DocDBInstanceStateAvailable),
									PreferredBackupWindow: awsclient.String(testPreferredBackupWindow),
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
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
									DBClusterIdentifier:        awsclient.String(testDBClusterIdentifier),
									Status:                     awsclient.String(svcapitypes.DocDBInstanceStateAvailable),
									PreferredMaintenanceWindow: awsclient.String(testPreferredMaintenanceWindow),
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
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
									DBClusterIdentifier:        awsclient.String(testDBClusterIdentifier),
									Status:                     awsclient.String(svcapitypes.DocDBInstanceStateAvailable),
									PreferredMaintenanceWindow: awsclient.String(testPreferredMaintenanceWindow),
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
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
									DBClusterIdentifier:        awsclient.String(testDBClusterIdentifier),
									Status:                     awsclient.String(svcapitypes.DocDBInstanceStateAvailable),
									PreferredMaintenanceWindow: awsclient.String(testPreferredMaintenanceWindow),
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
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
			if diff := cmp.Diff(tc.want.docdb, tc.args.docdb.Called); diff != "" {
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
								DBClusterArn:        awsclient.String(testDBClusterArn),
								Endpoint:            awsclient.String(testEndpoint),
								ReaderEndpoint:      awsclient.String(testReaderEndpoint),
								Port:                awsclient.Int64(testPort),
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
						&svcapitypes.Tag{Key: awsclient.String(testTagKey), Value: awsclient.String(testTagValue)},
						&svcapitypes.Tag{Key: awsclient.String(testOtherTagKey), Value: awsclient.String(testOtherTagValue)},
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
						&svcapitypes.Tag{Key: awsclient.String(testTagKey), Value: awsclient.String(testTagValue)},
						&svcapitypes.Tag{Key: awsclient.String(testOtherTagKey), Value: awsclient.String(testOtherTagValue)},
					),
					withVpcSecurityGroupIds(
						testVpcSecurityGroup,
						testOtherVpcSecurityGroup,
					),
					withEndpoint(testEndpoint),
					withReaderEndpoint(testReaderEndpoint),
					withMasterPasswordSecretRef(testMasterPasswordSecretNamespace, testMasterPasswordSecretName, testMasterPasswordSecretKey),
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
								AvailabilityZones: toStringPtrArray(
									testAvailabilityZone,
									testOtherAvailabilityZone,
								),
								BackupRetentionPeriod:       awsclient.Int64(testBackupRetentionPeriod),
								DBClusterParameterGroupName: awsclient.String(testDBClusterParameterGroupName),
								DBSubnetGroupName:           awsclient.String(testDBSubnetGroupName),
								DeletionProtection:          awsclient.Bool(true),
								EnableCloudwatchLogsExports: toStringPtrArray(
									testCloudWatchLog,
									testOtherCloudWatchLog,
								),
								Engine:                     awsclient.String(testEngine),
								EngineVersion:              awsclient.String(testEngineVersion),
								KmsKeyId:                   awsclient.String(testKMSKeyID),
								MasterUsername:             awsclient.String(testMasterUserName),
								MasterUserPassword:         awsclient.String(testMasterUserPassword),
								Port:                       awsclient.Int64(testPort),
								PreSignedUrl:               awsclient.String(testPresignedURL),
								PreferredBackupWindow:      awsclient.String(testPreferredBackupWindow),
								PreferredMaintenanceWindow: awsclient.String(testPreferredMaintenanceWindow),
								StorageEncrypted:           awsclient.Bool(true),
								VpcSecurityGroupIds: toStringPtrArray(
									testVpcSecurityGroup,
									testOtherVpcSecurityGroup,
								),
								Tags: []*docdb.Tag{
									{Key: awsclient.String(testTagKey), Value: awsclient.String(testTagValue)},
									{Key: awsclient.String(testOtherTagKey), Value: awsclient.String(testOtherTagValue)},
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
								DBClusterArn:        awsclient.String(testDBClusterArn),
								Endpoint:            awsclient.String(testEndpoint),
								ReaderEndpoint:      awsclient.String(testReaderEndpoint),
								Port:                awsclient.Int64(testPort),
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
			if diff := cmp.Diff(tc.want.docdb, tc.args.docdb.Called); diff != "" {
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
								DBClusterIdentifier:       awsclient.String(testDBClusterIdentifier),
								FinalDBSnapshotIdentifier: awsclient.String(testFinalDBSnapshotIdentifier),
								SkipFinalSnapshot:         awsclient.Bool(true),
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
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
			if diff := cmp.Diff(tc.want.docdb, tc.args.docdb.Called); diff != "" {
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
								DBClusterArn:        awsclient.String(testDBClusterArn),
							},
						}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{
							TagList: []*docdb.Tag{
								{
									Key:   awsclient.String(testOtherOtherTagKey),
									Value: awsclient.String(testOtherOtherTagValue),
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
						&svcapitypes.Tag{Key: awsclient.String(testTagKey), Value: awsclient.String(testTagValue)},
						&svcapitypes.Tag{Key: awsclient.String(testOtherTagKey), Value: awsclient.String(testOtherTagValue)},
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
						&svcapitypes.Tag{Key: awsclient.String(testTagKey), Value: awsclient.String(testTagValue)},
						&svcapitypes.Tag{Key: awsclient.String(testOtherTagKey), Value: awsclient.String(testOtherTagValue)},
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
								DBClusterIdentifier:         awsclient.String(testDBClusterIdentifier),
								BackupRetentionPeriod:       awsclient.Int64(testBackupRetentionPeriod),
								DBClusterParameterGroupName: awsclient.String(testDBClusterParameterGroupName),
								DeletionProtection:          awsclient.Bool(true),
								EngineVersion:               awsclient.String(testEngineVersion),
								Port:                        awsclient.Int64(testPort),
								PreferredBackupWindow:       awsclient.String(testPreferredBackupWindow),
								PreferredMaintenanceWindow:  awsclient.String(testPreferredMaintenanceWindow),
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
								ResourceName: awsclient.String(testDBClusterArn),
							},
						},
					},
					AddTagsToResource: []*fake.CallAddTagsToResource{
						{
							I: &docdb.AddTagsToResourceInput{
								ResourceName: awsclient.String(testDBClusterArn),
								Tags: []*docdb.Tag{
									{Key: awsclient.String(testTagKey), Value: awsclient.String(testTagValue)},
									{Key: awsclient.String(testOtherTagKey), Value: awsclient.String(testOtherTagValue)},
								},
							},
						},
					},
					RemoveTagsFromResource: []*fake.CallRemoveTagsFromResource{
						{
							I: &docdb.RemoveTagsFromResourceInput{
								ResourceName: awsclient.String(testDBClusterArn),
								TagKeys: []*string{
									awsclient.String(testOtherOtherTagKey),
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
								DBClusterArn:        awsclient.String(testDBClusterArn),
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
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
								ResourceName: awsclient.String(testDBClusterArn),
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
								DBClusterArn:        awsclient.String(testDBClusterArn),
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
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
								ResourceName: awsclient.String(testDBClusterArn),
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
								DBClusterArn:        awsclient.String(testDBClusterArn),
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
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
								ResourceName: awsclient.String(testDBClusterArn),
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
								DBClusterIdentifier: awsclient.String(testDBClusterIdentifier),
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
			if diff := cmp.Diff(tc.want.docdb, tc.args.docdb.Called); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestInitialize(t *testing.T) {
	type want struct {
		cr  *svcapitypes.DBCluster
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				cr: instance(withTags(
					&svcapitypes.Tag{Key: awsclient.String(testTagKey), Value: awsclient.String(testTagValue)},
					&svcapitypes.Tag{Key: awsclient.String(testOtherTagKey), Value: awsclient.String(testOtherTagValue)},
				)),
				kube: &test.MockClient{MockUpdate: test.NewMockUpdateFn(nil)},
			},
			want: want{
				cr: instance(withTags(
					mergeTags(
						[]*svcapitypes.Tag{
							{Key: awsclient.String(testTagKey), Value: awsclient.String(testTagValue)},
							{Key: awsclient.String(testOtherTagKey), Value: awsclient.String(testOtherTagValue)},
						},
						svcutils.GetExternalTags(instance()),
					)...,
				)),
			},
		},
		"UpdateFailed": {
			args: args{
				cr: instance(withTags(
					&svcapitypes.Tag{Key: awsclient.String(testTagKey), Value: awsclient.String(testTagValue)},
					&svcapitypes.Tag{Key: awsclient.String(testOtherTagKey), Value: awsclient.String(testOtherTagValue)},
				)),
				kube: &test.MockClient{MockUpdate: test.NewMockUpdateFn(errors.New(testErrBoom))},
			},
			want: want{
				cr: instance(withTags(
					mergeTags(
						[]*svcapitypes.Tag{
							{Key: awsclient.String(testTagKey), Value: awsclient.String(testTagValue)},
							{Key: awsclient.String(testOtherTagKey), Value: awsclient.String(testOtherTagValue)},
						},
						svcutils.GetExternalTags(instance()),
					)...,
				)),
				err: awsclient.Wrap(errors.New(testErrBoom), errKubeUpdateFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &tagger{kube: tc.kube}
			err := e.Initialize(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
