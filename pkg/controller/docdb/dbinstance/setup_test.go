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

package dbinstance

import (
	"context"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/docdb"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	cpresource "github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/docdb/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/docdb/fake"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

var (
	testDBIdentifier                    = "some-db-name"
	testDBInstanceArn                   = "aws:test:docdb:test-region:test-account:test-db"
	testAddress                         = "127.0.0.1"
	testHostedZone                      = "some-hosted-zone"
	testPort                            = 27017
	testAvailabilityZone                = "test-zone-a"
	testOtherAvailabilityZone           = "test-zone-b"
	testCACertificateIdentifier         = "some-certificate"
	testOtherCACertificateIdentifier    = "some-other-certificate"
	testDBInstanceClass                 = "some-db-instance-class"
	testOtherDBInstanceClass            = "some-other-db-instance-class"
	testPreferredMaintenanceWindow      = "1 day"
	testOtherPreferredMaintenanceWindow = "10 days"
	testPromotionTier                   = 42
	testOtherPromotionTier              = 9000
	testTagKey                          = "some-tag-key"
	testTagValue                        = "some-tag-value"
	testOtherTagKey                     = "some-other-tag-key"
	testOtherTagValue                   = "some-other-tag-value"

	testErrDescribeDBInstancesFailed = "DescribeDBInstances failed"
	testErrCreateDBInstanceFailed    = "CreateDBInstance failed"
	testErrDeleteDBInstanceFailed    = "DeleteDBInstance failed"
	testErrModifyDBInstanceFailed    = "ModifyDBInstance failed"
)

type args struct {
	docdb *fake.MockDocDBClient
	kube  client.Client
	cr    *svcapitypes.DBInstance
}

type docDBModifier func(*svcapitypes.DBInstance)

func instance(m ...docDBModifier) *svcapitypes.DBInstance {
	cr := &svcapitypes.DBInstance{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func withExternalName(value string) docDBModifier {
	return func(o *svcapitypes.DBInstance) {
		meta.SetExternalName(o, value)
	}
}

func withDBIdentifier(value string) docDBModifier {
	return func(o *svcapitypes.DBInstance) {
		o.Status.AtProvider.DBInstanceIdentifier = pointer.ToOrNilIfZeroValue(value)
	}
}

func withDBInstanceArn(value string) docDBModifier {
	return func(o *svcapitypes.DBInstance) {
		o.Status.AtProvider.DBInstanceARN = pointer.ToOrNilIfZeroValue(value)
	}
}

func withEndpoint(value *svcapitypes.Endpoint) docDBModifier {
	return func(o *svcapitypes.DBInstance) {
		o.Status.AtProvider.Endpoint = value
	}
}

func withStatus(value string) docDBModifier {
	return func(o *svcapitypes.DBInstance) {
		o.Status.AtProvider.DBInstanceStatus = pointer.ToOrNilIfZeroValue(value)
	}
}

func withConditions(value ...xpv1.Condition) docDBModifier {
	return func(o *svcapitypes.DBInstance) {
		o.Status.SetConditions(value...)
	}
}

func withAvailabilityZone(value string) docDBModifier {
	return func(o *svcapitypes.DBInstance) {
		o.Spec.ForProvider.AvailabilityZone = pointer.ToOrNilIfZeroValue(value)
	}
}

func withAutoMinorVersionUpgrade(value bool) docDBModifier {
	return func(o *svcapitypes.DBInstance) {
		o.Spec.ForProvider.AutoMinorVersionUpgrade = ptr.To(value)
	}
}

func withCACertificateIdentifier(value string) docDBModifier {
	return func(o *svcapitypes.DBInstance) {
		o.Spec.ForProvider.CACertificateIdentifier = pointer.ToOrNilIfZeroValue(value)
	}
}

func withStatusCACertificateIdentifier(value string) docDBModifier {
	return func(o *svcapitypes.DBInstance) {
		o.Status.AtProvider.CACertificateIdentifier = pointer.ToOrNilIfZeroValue(value)
	}
}

func withDBInstanceClass(value string) docDBModifier {
	return func(o *svcapitypes.DBInstance) {
		o.Spec.ForProvider.DBInstanceClass = pointer.ToOrNilIfZeroValue(value)
	}
}

func withPreferredMaintenanceWindow(value string) docDBModifier {
	return func(o *svcapitypes.DBInstance) {
		o.Spec.ForProvider.PreferredMaintenanceWindow = pointer.ToOrNilIfZeroValue(value)
	}
}

func withPromotionTier(value int) docDBModifier {
	return func(o *svcapitypes.DBInstance) {
		o.Spec.ForProvider.PromotionTier = pointer.ToIntAsInt64(value)
	}
}

func withTags(values ...*svcapitypes.Tag) docDBModifier {
	return func(o *svcapitypes.DBInstance) {
		o.Spec.ForProvider.Tags = values
	}
}

func generateConnectionDetails(address string, port int) managed.ConnectionDetails {
	return managed.ConnectionDetails{
		xpv1.ResourceCredentialsSecretEndpointKey: []byte(address),
		xpv1.ResourceCredentialsSecretPortKey:     []byte(strconv.Itoa(port)),
	}
}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *svcapitypes.DBInstance
		result managed.ExternalObservation
		err    error
		docdb  fake.MockDocDBClientCall
	}

	cases := map[string]struct {
		args
		want
	}{
		"AvailableState_and_UpToDate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBInstancesWithContext: func(c context.Context, ddi *docdb.DescribeDBInstancesInput, o []request.Option) (*docdb.DescribeDBInstancesOutput, error) {
						return &docdb.DescribeDBInstancesOutput{DBInstances: []*docdb.DBInstance{
							{
								DBInstanceStatus:     pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateAvailable),
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
								DBInstanceArn:        pointer.ToOrNilIfZeroValue(testDBInstanceArn),
								Endpoint:             &docdb.Endpoint{Address: pointer.ToOrNilIfZeroValue(testAddress), Port: pointer.ToIntAsInt64(testPort)},
							},
						}}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{TagList: []*docdb.Tag{}}, nil
					},
				},
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withDBInstanceArn(testDBInstanceArn),
					withStatus(svcapitypes.DocDBInstanceStateAvailable),
					withConditions(xpv1.Available()),
					withEndpoint(&svcapitypes.Endpoint{
						Address: pointer.ToOrNilIfZeroValue(testAddress),
						Port:    pointer.ToIntAsInt64(testPort),
					}),
				),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: generateConnectionDetails(testAddress, testPort),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBInstancesWithContext: []*fake.CallDescribeDBInstancesWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeDBInstancesInput(instance(
								withExternalName(testDBIdentifier),
								withDBIdentifier(testDBIdentifier),
							)),
							Opts: nil,
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{I: &docdb.ListTagsForResourceInput{ResourceName: pointer.ToOrNilIfZeroValue(testDBInstanceArn)}},
					},
				},
			},
		},
		"AvailableState_and_outdated_AvailabilityZone": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBInstancesWithContext: func(c context.Context, ddi *docdb.DescribeDBInstancesInput, o []request.Option) (*docdb.DescribeDBInstancesOutput, error) {
						return &docdb.DescribeDBInstancesOutput{DBInstances: []*docdb.DBInstance{
							{
								DBInstanceStatus:     pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateAvailable),
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
								DBInstanceArn:        pointer.ToOrNilIfZeroValue(testDBInstanceArn),
								Endpoint:             &docdb.Endpoint{Address: pointer.ToOrNilIfZeroValue(testAddress), Port: pointer.ToIntAsInt64(testPort)},
								AvailabilityZone:     pointer.ToOrNilIfZeroValue(testAvailabilityZone),
							},
						}}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{TagList: []*docdb.Tag{}}, nil
					},
				},
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withAvailabilityZone(testOtherAvailabilityZone),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withDBInstanceArn(testDBInstanceArn),
					withConditions(xpv1.Available()),
					withStatus(svcapitypes.DocDBInstanceStateAvailable),
					withEndpoint(&svcapitypes.Endpoint{
						Address: pointer.ToOrNilIfZeroValue(testAddress),
						Port:    pointer.ToIntAsInt64(testPort),
					}),
					withAvailabilityZone(testOtherAvailabilityZone),
				),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: generateConnectionDetails(testAddress, testPort),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBInstancesWithContext: []*fake.CallDescribeDBInstancesWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeDBInstancesInput(instance(
								withExternalName(testDBIdentifier),
								withDBIdentifier(testDBIdentifier),
							)),
							Opts: nil,
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{
								ResourceName: pointer.ToOrNilIfZeroValue(testDBInstanceArn),
							},
						},
					},
				},
			},
		},
		"AvailableState_and_outdated_AutoMinorVersionUpgrade": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBInstancesWithContext: func(c context.Context, ddi *docdb.DescribeDBInstancesInput, o []request.Option) (*docdb.DescribeDBInstancesOutput, error) {
						return &docdb.DescribeDBInstancesOutput{DBInstances: []*docdb.DBInstance{
							{
								DBInstanceStatus:        pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateAvailable),
								DBInstanceIdentifier:    pointer.ToOrNilIfZeroValue(testDBIdentifier),
								DBInstanceArn:           pointer.ToOrNilIfZeroValue(testDBInstanceArn),
								Endpoint:                &docdb.Endpoint{Address: pointer.ToOrNilIfZeroValue(testAddress), Port: pointer.ToIntAsInt64(testPort)},
								AutoMinorVersionUpgrade: pointer.ToOrNilIfZeroValue(true),
							},
						}}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{TagList: []*docdb.Tag{}}, nil
					},
				},
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withAutoMinorVersionUpgrade(false),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withDBInstanceArn(testDBInstanceArn),
					withStatus(svcapitypes.DocDBInstanceStateAvailable),
					withConditions(xpv1.Available()),
					withEndpoint(&svcapitypes.Endpoint{
						Address: pointer.ToOrNilIfZeroValue(testAddress),
						Port:    pointer.ToIntAsInt64(testPort),
					}),
					withAutoMinorVersionUpgrade(false),
				),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  false,
					ConnectionDetails: generateConnectionDetails(testAddress, testPort),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBInstancesWithContext: []*fake.CallDescribeDBInstancesWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeDBInstancesInput(instance(
								withExternalName(testDBIdentifier),
								withDBIdentifier(testDBIdentifier),
							)),
							Opts: nil,
						},
					},
				},
			},
		},
		"AvailableState_and_outdated_CACertificateIdentifier": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBInstancesWithContext: func(c context.Context, ddi *docdb.DescribeDBInstancesInput, o []request.Option) (*docdb.DescribeDBInstancesOutput, error) {
						return &docdb.DescribeDBInstancesOutput{DBInstances: []*docdb.DBInstance{
							{
								DBInstanceStatus:        pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateAvailable),
								DBInstanceIdentifier:    pointer.ToOrNilIfZeroValue(testDBIdentifier),
								DBInstanceArn:           pointer.ToOrNilIfZeroValue(testDBInstanceArn),
								Endpoint:                &docdb.Endpoint{Address: pointer.ToOrNilIfZeroValue(testAddress), Port: pointer.ToIntAsInt64(testPort)},
								CACertificateIdentifier: pointer.ToOrNilIfZeroValue(testCACertificateIdentifier),
							},
						}}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{TagList: []*docdb.Tag{}}, nil
					},
				},
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withCACertificateIdentifier(testOtherCACertificateIdentifier),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withDBInstanceArn(testDBInstanceArn),
					withStatus(svcapitypes.DocDBInstanceStateAvailable),
					withConditions(xpv1.Available()),
					withEndpoint(&svcapitypes.Endpoint{
						Address: pointer.ToOrNilIfZeroValue(testAddress),
						Port:    pointer.ToIntAsInt64(testPort),
					}),
					withCACertificateIdentifier(testOtherCACertificateIdentifier),
					withStatusCACertificateIdentifier(testCACertificateIdentifier),
				),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  false,
					ConnectionDetails: generateConnectionDetails(testAddress, testPort),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBInstancesWithContext: []*fake.CallDescribeDBInstancesWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeDBInstancesInput(instance(
								withExternalName(testDBIdentifier),
								withDBIdentifier(testDBIdentifier),
							)),
							Opts: nil,
						},
					},
				},
			},
		},
		"AvailableState_and_outdated_DBInstanceClass": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBInstancesWithContext: func(c context.Context, ddi *docdb.DescribeDBInstancesInput, o []request.Option) (*docdb.DescribeDBInstancesOutput, error) {
						return &docdb.DescribeDBInstancesOutput{DBInstances: []*docdb.DBInstance{
							{
								DBInstanceStatus:     pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateAvailable),
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
								DBInstanceArn:        pointer.ToOrNilIfZeroValue(testDBInstanceArn),
								Endpoint:             &docdb.Endpoint{Address: pointer.ToOrNilIfZeroValue(testAddress), Port: pointer.ToIntAsInt64(testPort)},
								DBInstanceClass:      pointer.ToOrNilIfZeroValue(testDBInstanceClass),
							},
						}}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{TagList: []*docdb.Tag{}}, nil
					},
				},
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withDBInstanceClass(testOtherDBInstanceClass),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withDBInstanceArn(testDBInstanceArn),
					withStatus(svcapitypes.DocDBInstanceStateAvailable),
					withConditions(xpv1.Available()),
					withEndpoint(&svcapitypes.Endpoint{
						Address: pointer.ToOrNilIfZeroValue(testAddress),
						Port:    pointer.ToIntAsInt64(testPort),
					}),
					withDBInstanceClass(testOtherDBInstanceClass),
				),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  false,
					ConnectionDetails: generateConnectionDetails(testAddress, testPort),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBInstancesWithContext: []*fake.CallDescribeDBInstancesWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeDBInstancesInput(instance(
								withExternalName(testDBIdentifier),
								withDBIdentifier(testDBIdentifier),
							)),
							Opts: nil,
						},
					},
				},
			},
		},
		"AvailableState_and_outdated_PreferredMaintenanceWindow": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBInstancesWithContext: func(c context.Context, ddi *docdb.DescribeDBInstancesInput, o []request.Option) (*docdb.DescribeDBInstancesOutput, error) {
						return &docdb.DescribeDBInstancesOutput{DBInstances: []*docdb.DBInstance{
							{
								DBInstanceStatus:           pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateAvailable),
								DBInstanceIdentifier:       pointer.ToOrNilIfZeroValue(testDBIdentifier),
								DBInstanceArn:              pointer.ToOrNilIfZeroValue(testDBInstanceArn),
								Endpoint:                   &docdb.Endpoint{Address: pointer.ToOrNilIfZeroValue(testAddress), Port: pointer.ToIntAsInt64(testPort)},
								PreferredMaintenanceWindow: pointer.ToOrNilIfZeroValue(testPreferredMaintenanceWindow),
							},
						}}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{TagList: []*docdb.Tag{}}, nil
					},
				},
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withPreferredMaintenanceWindow(testOtherPreferredMaintenanceWindow),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withDBInstanceArn(testDBInstanceArn),
					withStatus(svcapitypes.DocDBInstanceStateAvailable),
					withConditions(xpv1.Available()),
					withEndpoint(&svcapitypes.Endpoint{
						Address: pointer.ToOrNilIfZeroValue(testAddress),
						Port:    pointer.ToIntAsInt64(testPort),
					}),
					withPreferredMaintenanceWindow(testOtherPreferredMaintenanceWindow),
				),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  false,
					ConnectionDetails: generateConnectionDetails(testAddress, testPort),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBInstancesWithContext: []*fake.CallDescribeDBInstancesWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeDBInstancesInput(instance(
								withExternalName(testDBIdentifier),
								withDBIdentifier(testDBIdentifier),
							)),
							Opts: nil,
						},
					},
				},
			},
		},
		"AvailableState_and_outdated_PromotionTier": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBInstancesWithContext: func(c context.Context, ddi *docdb.DescribeDBInstancesInput, o []request.Option) (*docdb.DescribeDBInstancesOutput, error) {
						return &docdb.DescribeDBInstancesOutput{DBInstances: []*docdb.DBInstance{
							{
								DBInstanceStatus:     pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateAvailable),
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
								Endpoint:             &docdb.Endpoint{Address: pointer.ToOrNilIfZeroValue(testAddress), Port: pointer.ToIntAsInt64(testPort)},
								PromotionTier:        pointer.ToIntAsInt64(testPromotionTier),
							},
						}}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{TagList: []*docdb.Tag{}}, nil
					},
				},
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withPromotionTier(testOtherPromotionTier),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withStatus(svcapitypes.DocDBInstanceStateAvailable),
					withConditions(xpv1.Available()),
					withEndpoint(&svcapitypes.Endpoint{
						Address: pointer.ToOrNilIfZeroValue(testAddress),
						Port:    pointer.ToIntAsInt64(testPort),
					}),
					withPromotionTier(testOtherPromotionTier),
				),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  false,
					ConnectionDetails: generateConnectionDetails(testAddress, testPort),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBInstancesWithContext: []*fake.CallDescribeDBInstancesWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeDBInstancesInput(instance(
								withExternalName(testDBIdentifier),
								withDBIdentifier(testDBIdentifier),
							)),
							Opts: nil,
						},
					},
				},
			},
		},
		"AvailableState_and_outdated_tags": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBInstancesWithContext: func(c context.Context, ddi *docdb.DescribeDBInstancesInput, o []request.Option) (*docdb.DescribeDBInstancesOutput, error) {
						return &docdb.DescribeDBInstancesOutput{DBInstances: []*docdb.DBInstance{
							{
								DBInstanceStatus:     pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateAvailable),
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
								DBInstanceArn:        pointer.ToOrNilIfZeroValue(testDBInstanceArn),
								Endpoint:             &docdb.Endpoint{Address: pointer.ToOrNilIfZeroValue(testAddress), Port: pointer.ToIntAsInt64(testPort)},
							},
						}}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{TagList: []*docdb.Tag{
							{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testTagValue)},
						}}, nil
					},
				},
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withTags(&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testOtherTagKey), Value: pointer.ToOrNilIfZeroValue(testOtherTagValue)}),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withDBInstanceArn(testDBInstanceArn),
					withStatus(svcapitypes.DocDBInstanceStateAvailable),
					withConditions(xpv1.Available()),
					withEndpoint(&svcapitypes.Endpoint{
						Address: pointer.ToOrNilIfZeroValue(testAddress),
						Port:    pointer.ToIntAsInt64(testPort),
					}),
					withTags(&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testOtherTagKey), Value: pointer.ToOrNilIfZeroValue(testOtherTagValue)}),
				),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  false,
					ConnectionDetails: generateConnectionDetails(testAddress, testPort),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBInstancesWithContext: []*fake.CallDescribeDBInstancesWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeDBInstancesInput(instance(
								withExternalName(testDBIdentifier),
								withDBIdentifier(testDBIdentifier),
							)),
							Opts: nil,
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{ResourceName: pointer.ToOrNilIfZeroValue(testDBInstanceArn)},
						},
					},
				},
			},
		},
		"AvailableState_and_same_tags": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBInstancesWithContext: func(c context.Context, ddi *docdb.DescribeDBInstancesInput, o []request.Option) (*docdb.DescribeDBInstancesOutput, error) {
						return &docdb.DescribeDBInstancesOutput{DBInstances: []*docdb.DBInstance{
							{
								DBInstanceStatus:     pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateAvailable),
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
								DBInstanceArn:        pointer.ToOrNilIfZeroValue(testDBInstanceArn),
								Endpoint:             &docdb.Endpoint{Address: pointer.ToOrNilIfZeroValue(testAddress), Port: pointer.ToIntAsInt64(testPort)},
							},
						}}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{TagList: []*docdb.Tag{
							{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testTagValue)},
						}}, nil
					},
				},
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withTags(&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testTagValue)}),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withDBInstanceArn(testDBInstanceArn),
					withStatus(svcapitypes.DocDBInstanceStateAvailable),
					withConditions(xpv1.Available()),
					withEndpoint(&svcapitypes.Endpoint{
						Address: pointer.ToOrNilIfZeroValue(testAddress),
						Port:    pointer.ToIntAsInt64(testPort),
					}),
					withTags(&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testTagValue)}),
				),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: generateConnectionDetails(testAddress, testPort),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBInstancesWithContext: []*fake.CallDescribeDBInstancesWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeDBInstancesInput(instance(
								withExternalName(testDBIdentifier),
								withDBIdentifier(testDBIdentifier),
							)),
							Opts: nil,
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{
								ResourceName: pointer.ToOrNilIfZeroValue(testDBInstanceArn),
							},
						},
					},
				},
			},
		},
		"FailedState_and_UpToDate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBInstancesWithContext: func(c context.Context, ddi *docdb.DescribeDBInstancesInput, o []request.Option) (*docdb.DescribeDBInstancesOutput, error) {
						return &docdb.DescribeDBInstancesOutput{DBInstances: []*docdb.DBInstance{
							{
								DBInstanceStatus:     pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateFailed),
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
								DBInstanceArn:        pointer.ToOrNilIfZeroValue(testDBInstanceArn),
							},
						}}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{TagList: []*docdb.Tag{}}, nil
					},
				},
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withDBInstanceArn(testDBInstanceArn),
					withStatus(svcapitypes.DocDBInstanceStateFailed),
					withConditions(xpv1.Unavailable()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBInstancesWithContext: []*fake.CallDescribeDBInstancesWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeDBInstancesInput(instance(
								withExternalName(testDBIdentifier),
								withDBIdentifier(testDBIdentifier),
							)),
							Opts: nil,
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{
								ResourceName: pointer.ToOrNilIfZeroValue(testDBInstanceArn),
							},
						},
					},
				},
			},
		},
		"FailedState_and_outdated_AvailabilityZone": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBInstancesWithContext: func(c context.Context, ddi *docdb.DescribeDBInstancesInput, o []request.Option) (*docdb.DescribeDBInstancesOutput, error) {
						return &docdb.DescribeDBInstancesOutput{DBInstances: []*docdb.DBInstance{
							{
								DBInstanceStatus:     pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateFailed),
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
								Endpoint:             &docdb.Endpoint{Address: pointer.ToOrNilIfZeroValue(testAddress), Port: pointer.ToIntAsInt64(testPort)},
								AvailabilityZone:     pointer.ToOrNilIfZeroValue(testAvailabilityZone),
							},
						}}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{TagList: []*docdb.Tag{}}, nil
					},
				},
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withAvailabilityZone(testOtherAvailabilityZone),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withConditions(xpv1.Unavailable()),
					withStatus(svcapitypes.DocDBInstanceStateFailed),
					withEndpoint(&svcapitypes.Endpoint{
						Address: pointer.ToOrNilIfZeroValue(testAddress),
						Port:    pointer.ToIntAsInt64(testPort),
					}),
					withAvailabilityZone(testOtherAvailabilityZone),
				),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: generateConnectionDetails(testAddress, testPort),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBInstancesWithContext: []*fake.CallDescribeDBInstancesWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeDBInstancesInput(instance(
								withExternalName(testDBIdentifier),
								withDBIdentifier(testDBIdentifier),
							)),
							Opts: nil,
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
		"FailedState_and_outdated_AutoMinorVersionUpgrade": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBInstancesWithContext: func(c context.Context, ddi *docdb.DescribeDBInstancesInput, o []request.Option) (*docdb.DescribeDBInstancesOutput, error) {
						return &docdb.DescribeDBInstancesOutput{DBInstances: []*docdb.DBInstance{
							{
								DBInstanceStatus:        pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateFailed),
								DBInstanceIdentifier:    pointer.ToOrNilIfZeroValue(testDBIdentifier),
								Endpoint:                &docdb.Endpoint{Address: pointer.ToOrNilIfZeroValue(testAddress), Port: pointer.ToIntAsInt64(testPort)},
								AutoMinorVersionUpgrade: pointer.ToOrNilIfZeroValue(true),
							},
						}}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{TagList: []*docdb.Tag{}}, nil
					},
				},
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withAutoMinorVersionUpgrade(false),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withStatus(svcapitypes.DocDBInstanceStateFailed),
					withConditions(xpv1.Unavailable()),
					withEndpoint(&svcapitypes.Endpoint{
						Address: pointer.ToOrNilIfZeroValue(testAddress),
						Port:    pointer.ToIntAsInt64(testPort),
					}),
					withAutoMinorVersionUpgrade(false),
				),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  false,
					ConnectionDetails: generateConnectionDetails(testAddress, testPort),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBInstancesWithContext: []*fake.CallDescribeDBInstancesWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeDBInstancesInput(instance(
								withExternalName(testDBIdentifier),
								withDBIdentifier(testDBIdentifier),
							)),
							Opts: nil,
						},
					},
				},
			},
		},
		"FailedState_and_outdated_CACertificateIdentifier": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBInstancesWithContext: func(c context.Context, ddi *docdb.DescribeDBInstancesInput, o []request.Option) (*docdb.DescribeDBInstancesOutput, error) {
						return &docdb.DescribeDBInstancesOutput{DBInstances: []*docdb.DBInstance{
							{
								DBInstanceStatus:        pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateFailed),
								DBInstanceIdentifier:    pointer.ToOrNilIfZeroValue(testDBIdentifier),
								Endpoint:                &docdb.Endpoint{Address: pointer.ToOrNilIfZeroValue(testAddress), Port: pointer.ToIntAsInt64(testPort)},
								CACertificateIdentifier: pointer.ToOrNilIfZeroValue(testCACertificateIdentifier),
							},
						}}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{TagList: []*docdb.Tag{}}, nil
					},
				},
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withCACertificateIdentifier(testOtherCACertificateIdentifier),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withStatus(svcapitypes.DocDBInstanceStateFailed),
					withConditions(xpv1.Unavailable()),
					withEndpoint(&svcapitypes.Endpoint{
						Address: pointer.ToOrNilIfZeroValue(testAddress),
						Port:    pointer.ToIntAsInt64(testPort),
					}),
					withCACertificateIdentifier(testOtherCACertificateIdentifier),
					withStatusCACertificateIdentifier(testCACertificateIdentifier),
				),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  false,
					ConnectionDetails: generateConnectionDetails(testAddress, testPort),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBInstancesWithContext: []*fake.CallDescribeDBInstancesWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeDBInstancesInput(instance(
								withExternalName(testDBIdentifier),
								withDBIdentifier(testDBIdentifier),
							)),
							Opts: nil,
						},
					},
				},
			},
		},
		"FailedState_and_outdated_DBInstanceClass": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBInstancesWithContext: func(c context.Context, ddi *docdb.DescribeDBInstancesInput, o []request.Option) (*docdb.DescribeDBInstancesOutput, error) {
						return &docdb.DescribeDBInstancesOutput{DBInstances: []*docdb.DBInstance{
							{
								DBInstanceStatus:     pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateFailed),
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
								Endpoint:             &docdb.Endpoint{Address: pointer.ToOrNilIfZeroValue(testAddress), Port: pointer.ToIntAsInt64(testPort)},
								DBInstanceClass:      pointer.ToOrNilIfZeroValue(testDBInstanceClass),
							},
						}}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{TagList: []*docdb.Tag{}}, nil
					},
				},
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withDBInstanceClass(testOtherDBInstanceClass),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withStatus(svcapitypes.DocDBInstanceStateFailed),
					withConditions(xpv1.Unavailable()),
					withEndpoint(&svcapitypes.Endpoint{
						Address: pointer.ToOrNilIfZeroValue(testAddress),
						Port:    pointer.ToIntAsInt64(testPort),
					}),
					withDBInstanceClass(testOtherDBInstanceClass),
				),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  false,
					ConnectionDetails: generateConnectionDetails(testAddress, testPort),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBInstancesWithContext: []*fake.CallDescribeDBInstancesWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeDBInstancesInput(instance(
								withExternalName(testDBIdentifier),
								withDBIdentifier(testDBIdentifier),
							)),
							Opts: nil,
						},
					},
				},
			},
		},
		"FailedState_and_outdated_PreferredMaintenanceWindow": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBInstancesWithContext: func(c context.Context, ddi *docdb.DescribeDBInstancesInput, o []request.Option) (*docdb.DescribeDBInstancesOutput, error) {
						return &docdb.DescribeDBInstancesOutput{DBInstances: []*docdb.DBInstance{
							{
								DBInstanceStatus:           pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateFailed),
								DBInstanceIdentifier:       pointer.ToOrNilIfZeroValue(testDBIdentifier),
								Endpoint:                   &docdb.Endpoint{Address: pointer.ToOrNilIfZeroValue(testAddress), Port: pointer.ToIntAsInt64(testPort)},
								PreferredMaintenanceWindow: pointer.ToOrNilIfZeroValue(testPreferredMaintenanceWindow),
							},
						}}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{TagList: []*docdb.Tag{}}, nil
					},
				},
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withPreferredMaintenanceWindow(testOtherPreferredMaintenanceWindow),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withStatus(svcapitypes.DocDBInstanceStateFailed),
					withConditions(xpv1.Unavailable()),
					withEndpoint(&svcapitypes.Endpoint{
						Address: pointer.ToOrNilIfZeroValue(testAddress),
						Port:    pointer.ToIntAsInt64(testPort),
					}),
					withPreferredMaintenanceWindow(testOtherPreferredMaintenanceWindow),
				),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  false,
					ConnectionDetails: generateConnectionDetails(testAddress, testPort),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBInstancesWithContext: []*fake.CallDescribeDBInstancesWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeDBInstancesInput(instance(
								withExternalName(testDBIdentifier),
								withDBIdentifier(testDBIdentifier),
							)),
							Opts: nil,
						},
					},
				},
			},
		},
		"FailedState_and_outdated_PromotionTier": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBInstancesWithContext: func(c context.Context, ddi *docdb.DescribeDBInstancesInput, o []request.Option) (*docdb.DescribeDBInstancesOutput, error) {
						return &docdb.DescribeDBInstancesOutput{DBInstances: []*docdb.DBInstance{
							{
								DBInstanceStatus:     pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateFailed),
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
								Endpoint:             &docdb.Endpoint{Address: pointer.ToOrNilIfZeroValue(testAddress), Port: pointer.ToIntAsInt64(testPort)},
								PromotionTier:        pointer.ToIntAsInt64(testPromotionTier),
							},
						}}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{TagList: []*docdb.Tag{}}, nil
					},
				},
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withPromotionTier(testOtherPromotionTier),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withStatus(svcapitypes.DocDBInstanceStateFailed),
					withConditions(xpv1.Unavailable()),
					withEndpoint(&svcapitypes.Endpoint{
						Address: pointer.ToOrNilIfZeroValue(testAddress),
						Port:    pointer.ToIntAsInt64(testPort),
					}),
					withPromotionTier(testOtherPromotionTier),
				),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  false,
					ConnectionDetails: generateConnectionDetails(testAddress, testPort),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBInstancesWithContext: []*fake.CallDescribeDBInstancesWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeDBInstancesInput(instance(
								withExternalName(testDBIdentifier),
								withDBIdentifier(testDBIdentifier),
							)),
							Opts: nil,
						},
					},
				},
			},
		},
		"FailedState_and_outdated_tags": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBInstancesWithContext: func(c context.Context, ddi *docdb.DescribeDBInstancesInput, o []request.Option) (*docdb.DescribeDBInstancesOutput, error) {
						return &docdb.DescribeDBInstancesOutput{DBInstances: []*docdb.DBInstance{
							{
								DBInstanceStatus:     pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateFailed),
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
								DBInstanceArn:        pointer.ToOrNilIfZeroValue(testDBInstanceArn),
								Endpoint:             &docdb.Endpoint{Address: pointer.ToOrNilIfZeroValue(testAddress), Port: pointer.ToIntAsInt64(testPort)},
							},
						}}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{TagList: []*docdb.Tag{
							{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testTagValue)},
						}}, nil
					},
				},
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withTags(&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testOtherTagKey), Value: pointer.ToOrNilIfZeroValue(testOtherTagValue)}),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withDBInstanceArn(testDBInstanceArn),
					withStatus(svcapitypes.DocDBInstanceStateFailed),
					withConditions(xpv1.Unavailable()),
					withEndpoint(&svcapitypes.Endpoint{
						Address: pointer.ToOrNilIfZeroValue(testAddress),
						Port:    pointer.ToIntAsInt64(testPort),
					}),
					withTags(&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testOtherTagKey), Value: pointer.ToOrNilIfZeroValue(testOtherTagValue)}),
				),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  false,
					ConnectionDetails: generateConnectionDetails(testAddress, testPort),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBInstancesWithContext: []*fake.CallDescribeDBInstancesWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeDBInstancesInput(instance(
								withExternalName(testDBIdentifier),
								withDBIdentifier(testDBIdentifier),
							)),
							Opts: nil,
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{
								ResourceName: pointer.ToOrNilIfZeroValue(testDBInstanceArn),
							},
						},
					},
				},
			},
		},
		"FailedState_and_same_tags": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBInstancesWithContext: func(c context.Context, ddi *docdb.DescribeDBInstancesInput, o []request.Option) (*docdb.DescribeDBInstancesOutput, error) {
						return &docdb.DescribeDBInstancesOutput{DBInstances: []*docdb.DBInstance{
							{
								DBInstanceStatus:     pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateFailed),
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
								DBInstanceArn:        pointer.ToOrNilIfZeroValue(testDBInstanceArn),
								Endpoint:             &docdb.Endpoint{Address: pointer.ToOrNilIfZeroValue(testAddress), Port: pointer.ToIntAsInt64(testPort)},
							},
						}}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{TagList: []*docdb.Tag{
							{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testTagValue)},
						}}, nil
					},
				},
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withTags(&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testTagValue)}),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withDBInstanceArn(testDBInstanceArn),
					withStatus(svcapitypes.DocDBInstanceStateFailed),
					withConditions(xpv1.Unavailable()),
					withEndpoint(&svcapitypes.Endpoint{
						Address: pointer.ToOrNilIfZeroValue(testAddress),
						Port:    pointer.ToIntAsInt64(testPort),
					}),
					withTags(&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testTagValue)}),
				),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: generateConnectionDetails(testAddress, testPort),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBInstancesWithContext: []*fake.CallDescribeDBInstancesWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeDBInstancesInput(instance(
								withExternalName(testDBIdentifier),
								withDBIdentifier(testDBIdentifier),
							)),
							Opts: nil,
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{
								ResourceName: pointer.ToOrNilIfZeroValue(testDBInstanceArn),
							},
						},
					},
				},
			},
		},
		"DeletingState_and_UpToDate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBInstancesWithContext: func(c context.Context, ddi *docdb.DescribeDBInstancesInput, o []request.Option) (*docdb.DescribeDBInstancesOutput, error) {
						return &docdb.DescribeDBInstancesOutput{DBInstances: []*docdb.DBInstance{
							{
								DBInstanceStatus:     pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateDeleting),
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
								DBInstanceArn:        pointer.ToOrNilIfZeroValue(testDBInstanceArn),
							},
						}}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{TagList: []*docdb.Tag{}}, nil
					},
				},
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withDBInstanceArn(testDBInstanceArn),
					withStatus(svcapitypes.DocDBInstanceStateDeleting),
					withConditions(xpv1.Deleting()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBInstancesWithContext: []*fake.CallDescribeDBInstancesWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeDBInstancesInput(instance(
								withExternalName(testDBIdentifier),
								withDBIdentifier(testDBIdentifier),
							)),
							Opts: nil,
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{
								ResourceName: pointer.ToOrNilIfZeroValue(testDBInstanceArn),
							},
						},
					},
				},
			},
		},
		"DeletingState_and_outdated_AvailabilityZone": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBInstancesWithContext: func(c context.Context, ddi *docdb.DescribeDBInstancesInput, o []request.Option) (*docdb.DescribeDBInstancesOutput, error) {
						return &docdb.DescribeDBInstancesOutput{DBInstances: []*docdb.DBInstance{
							{
								DBInstanceStatus:     pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateDeleting),
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
								Endpoint:             &docdb.Endpoint{Address: pointer.ToOrNilIfZeroValue(testAddress), Port: pointer.ToIntAsInt64(testPort)},
								AvailabilityZone:     pointer.ToOrNilIfZeroValue(testAvailabilityZone),
							},
						}}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{TagList: []*docdb.Tag{}}, nil
					},
				},
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withAvailabilityZone(testOtherAvailabilityZone),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withConditions(xpv1.Deleting()),
					withStatus(svcapitypes.DocDBInstanceStateDeleting),
					withEndpoint(&svcapitypes.Endpoint{
						Address: pointer.ToOrNilIfZeroValue(testAddress),
						Port:    pointer.ToIntAsInt64(testPort),
					}),
					withAvailabilityZone(testOtherAvailabilityZone),
				),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: generateConnectionDetails(testAddress, testPort),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBInstancesWithContext: []*fake.CallDescribeDBInstancesWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeDBInstancesInput(instance(
								withExternalName(testDBIdentifier),
								withDBIdentifier(testDBIdentifier),
							)),
							Opts: nil,
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
		"DeletingState_and_outdated_AutoMinorVersionUpgrade": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBInstancesWithContext: func(c context.Context, ddi *docdb.DescribeDBInstancesInput, o []request.Option) (*docdb.DescribeDBInstancesOutput, error) {
						return &docdb.DescribeDBInstancesOutput{DBInstances: []*docdb.DBInstance{
							{
								DBInstanceStatus:        pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateDeleting),
								DBInstanceIdentifier:    pointer.ToOrNilIfZeroValue(testDBIdentifier),
								Endpoint:                &docdb.Endpoint{Address: pointer.ToOrNilIfZeroValue(testAddress), Port: pointer.ToIntAsInt64(testPort)},
								AutoMinorVersionUpgrade: pointer.ToOrNilIfZeroValue(true),
							},
						}}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{TagList: []*docdb.Tag{}}, nil
					},
				},
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withAutoMinorVersionUpgrade(false),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withStatus(svcapitypes.DocDBInstanceStateDeleting),
					withConditions(xpv1.Deleting()),
					withEndpoint(&svcapitypes.Endpoint{
						Address: pointer.ToOrNilIfZeroValue(testAddress),
						Port:    pointer.ToIntAsInt64(testPort),
					}),
					withAutoMinorVersionUpgrade(false),
				),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  false,
					ConnectionDetails: generateConnectionDetails(testAddress, testPort),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBInstancesWithContext: []*fake.CallDescribeDBInstancesWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeDBInstancesInput(instance(
								withExternalName(testDBIdentifier),
								withDBIdentifier(testDBIdentifier),
							)),
							Opts: nil,
						},
					},
				},
			},
		},
		"DeletingState_and_outdated_CACertificateIdentifier": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBInstancesWithContext: func(c context.Context, ddi *docdb.DescribeDBInstancesInput, o []request.Option) (*docdb.DescribeDBInstancesOutput, error) {
						return &docdb.DescribeDBInstancesOutput{DBInstances: []*docdb.DBInstance{
							{
								DBInstanceStatus:        pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateDeleting),
								DBInstanceIdentifier:    pointer.ToOrNilIfZeroValue(testDBIdentifier),
								Endpoint:                &docdb.Endpoint{Address: pointer.ToOrNilIfZeroValue(testAddress), Port: pointer.ToIntAsInt64(testPort)},
								CACertificateIdentifier: pointer.ToOrNilIfZeroValue(testCACertificateIdentifier),
							},
						}}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{TagList: []*docdb.Tag{}}, nil
					},
				},
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withCACertificateIdentifier(testOtherCACertificateIdentifier),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withStatus(svcapitypes.DocDBInstanceStateDeleting),
					withConditions(xpv1.Deleting()),
					withEndpoint(&svcapitypes.Endpoint{
						Address: pointer.ToOrNilIfZeroValue(testAddress),
						Port:    pointer.ToIntAsInt64(testPort),
					}),
					withCACertificateIdentifier(testOtherCACertificateIdentifier),
					withStatusCACertificateIdentifier(testCACertificateIdentifier),
				),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  false,
					ConnectionDetails: generateConnectionDetails(testAddress, testPort),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBInstancesWithContext: []*fake.CallDescribeDBInstancesWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeDBInstancesInput(instance(
								withExternalName(testDBIdentifier),
								withDBIdentifier(testDBIdentifier),
							)),
							Opts: nil,
						},
					},
				},
			},
		},
		"DeletingState_and_outdated_DBInstanceClass": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBInstancesWithContext: func(c context.Context, ddi *docdb.DescribeDBInstancesInput, o []request.Option) (*docdb.DescribeDBInstancesOutput, error) {
						return &docdb.DescribeDBInstancesOutput{DBInstances: []*docdb.DBInstance{
							{
								DBInstanceStatus:     pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateDeleting),
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
								Endpoint:             &docdb.Endpoint{Address: pointer.ToOrNilIfZeroValue(testAddress), Port: pointer.ToIntAsInt64(testPort)},
								DBInstanceClass:      pointer.ToOrNilIfZeroValue(testDBInstanceClass),
							},
						}}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{TagList: []*docdb.Tag{}}, nil
					},
				},
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withDBInstanceClass(testOtherDBInstanceClass),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withStatus(svcapitypes.DocDBInstanceStateDeleting),
					withConditions(xpv1.Deleting()),
					withEndpoint(&svcapitypes.Endpoint{
						Address: pointer.ToOrNilIfZeroValue(testAddress),
						Port:    pointer.ToIntAsInt64(testPort),
					}),
					withDBInstanceClass(testOtherDBInstanceClass),
				),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  false,
					ConnectionDetails: generateConnectionDetails(testAddress, testPort),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBInstancesWithContext: []*fake.CallDescribeDBInstancesWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeDBInstancesInput(instance(
								withExternalName(testDBIdentifier),
								withDBIdentifier(testDBIdentifier),
							)),
							Opts: nil,
						},
					},
				},
			},
		},
		"DeletingState_and_outdated_PreferredMaintenanceWindow": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBInstancesWithContext: func(c context.Context, ddi *docdb.DescribeDBInstancesInput, o []request.Option) (*docdb.DescribeDBInstancesOutput, error) {
						return &docdb.DescribeDBInstancesOutput{DBInstances: []*docdb.DBInstance{
							{
								DBInstanceStatus:           pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateDeleting),
								DBInstanceIdentifier:       pointer.ToOrNilIfZeroValue(testDBIdentifier),
								Endpoint:                   &docdb.Endpoint{Address: pointer.ToOrNilIfZeroValue(testAddress), Port: pointer.ToIntAsInt64(testPort)},
								PreferredMaintenanceWindow: pointer.ToOrNilIfZeroValue(testPreferredMaintenanceWindow),
							},
						}}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{TagList: []*docdb.Tag{}}, nil
					},
				},
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withPreferredMaintenanceWindow(testOtherPreferredMaintenanceWindow),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withStatus(svcapitypes.DocDBInstanceStateDeleting),
					withConditions(xpv1.Deleting()),
					withEndpoint(&svcapitypes.Endpoint{
						Address: pointer.ToOrNilIfZeroValue(testAddress),
						Port:    pointer.ToIntAsInt64(testPort),
					}),
					withPreferredMaintenanceWindow(testOtherPreferredMaintenanceWindow),
				),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  false,
					ConnectionDetails: generateConnectionDetails(testAddress, testPort),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBInstancesWithContext: []*fake.CallDescribeDBInstancesWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeDBInstancesInput(instance(
								withExternalName(testDBIdentifier),
								withDBIdentifier(testDBIdentifier),
							)),
							Opts: nil,
						},
					},
				},
			},
		},
		"DeletingState_and_outdated_PromotionTier": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBInstancesWithContext: func(c context.Context, ddi *docdb.DescribeDBInstancesInput, o []request.Option) (*docdb.DescribeDBInstancesOutput, error) {
						return &docdb.DescribeDBInstancesOutput{DBInstances: []*docdb.DBInstance{
							{
								DBInstanceStatus:     pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateDeleting),
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
								Endpoint:             &docdb.Endpoint{Address: pointer.ToOrNilIfZeroValue(testAddress), Port: pointer.ToIntAsInt64(testPort)},
								PromotionTier:        pointer.ToIntAsInt64(testPromotionTier),
							},
						}}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{TagList: []*docdb.Tag{}}, nil
					},
				},
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withPromotionTier(testOtherPromotionTier),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withStatus(svcapitypes.DocDBInstanceStateDeleting),
					withConditions(xpv1.Deleting()),
					withEndpoint(&svcapitypes.Endpoint{
						Address: pointer.ToOrNilIfZeroValue(testAddress),
						Port:    pointer.ToIntAsInt64(testPort),
					}),
					withPromotionTier(testOtherPromotionTier),
				),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  false,
					ConnectionDetails: generateConnectionDetails(testAddress, testPort),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBInstancesWithContext: []*fake.CallDescribeDBInstancesWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeDBInstancesInput(instance(
								withExternalName(testDBIdentifier),
								withDBIdentifier(testDBIdentifier),
							)),
							Opts: nil,
						},
					},
				},
			},
		},
		"DeletingState_and_outdated_tags": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBInstancesWithContext: func(c context.Context, ddi *docdb.DescribeDBInstancesInput, o []request.Option) (*docdb.DescribeDBInstancesOutput, error) {
						return &docdb.DescribeDBInstancesOutput{DBInstances: []*docdb.DBInstance{
							{
								DBInstanceStatus:     pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateDeleting),
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
								DBInstanceArn:        pointer.ToOrNilIfZeroValue(testDBInstanceArn),
								Endpoint:             &docdb.Endpoint{Address: pointer.ToOrNilIfZeroValue(testAddress), Port: pointer.ToIntAsInt64(testPort)},
							},
						}}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{TagList: []*docdb.Tag{
							{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testTagValue)},
						}}, nil
					},
				},
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withTags(&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testOtherTagKey), Value: pointer.ToOrNilIfZeroValue(testOtherTagValue)}),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withDBInstanceArn(testDBInstanceArn),
					withStatus(svcapitypes.DocDBInstanceStateDeleting),
					withConditions(xpv1.Deleting()),
					withEndpoint(&svcapitypes.Endpoint{
						Address: pointer.ToOrNilIfZeroValue(testAddress),
						Port:    pointer.ToIntAsInt64(testPort),
					}),
					withTags(&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testOtherTagKey), Value: pointer.ToOrNilIfZeroValue(testOtherTagValue)}),
				),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  false,
					ConnectionDetails: generateConnectionDetails(testAddress, testPort),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBInstancesWithContext: []*fake.CallDescribeDBInstancesWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeDBInstancesInput(instance(
								withExternalName(testDBIdentifier),
								withDBIdentifier(testDBIdentifier),
							)),
							Opts: nil,
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{
								ResourceName: pointer.ToOrNilIfZeroValue(testDBInstanceArn),
							},
						},
					},
				},
			},
		},
		"DeletingState_and_same_tags": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBInstancesWithContext: func(c context.Context, ddi *docdb.DescribeDBInstancesInput, o []request.Option) (*docdb.DescribeDBInstancesOutput, error) {
						return &docdb.DescribeDBInstancesOutput{DBInstances: []*docdb.DBInstance{
							{
								DBInstanceStatus:     pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateDeleting),
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
								DBInstanceArn:        pointer.ToOrNilIfZeroValue(testDBInstanceArn),
								Endpoint:             &docdb.Endpoint{Address: pointer.ToOrNilIfZeroValue(testAddress), Port: pointer.ToIntAsInt64(testPort)},
							},
						}}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{TagList: []*docdb.Tag{
							{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testTagValue)},
						}}, nil
					},
				},
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withTags(&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testTagValue)}),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withDBInstanceArn(testDBInstanceArn),
					withStatus(svcapitypes.DocDBInstanceStateDeleting),
					withConditions(xpv1.Deleting()),
					withEndpoint(&svcapitypes.Endpoint{
						Address: pointer.ToOrNilIfZeroValue(testAddress),
						Port:    pointer.ToIntAsInt64(testPort),
					}),
					withTags(&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testTagValue)}),
				),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: generateConnectionDetails(testAddress, testPort),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBInstancesWithContext: []*fake.CallDescribeDBInstancesWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeDBInstancesInput(instance(
								withExternalName(testDBIdentifier),
								withDBIdentifier(testDBIdentifier),
							)),
							Opts: nil,
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{
								ResourceName: pointer.ToOrNilIfZeroValue(testDBInstanceArn),
							},
						},
					},
				},
			},
		},
		"CreatingState": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBInstancesWithContext: func(c context.Context, ddi *docdb.DescribeDBInstancesInput, o []request.Option) (*docdb.DescribeDBInstancesOutput, error) {
						return &docdb.DescribeDBInstancesOutput{DBInstances: []*docdb.DBInstance{
							{
								DBInstanceStatus:     pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateCreating),
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
								DBInstanceArn:        pointer.ToOrNilIfZeroValue(testDBInstanceArn),
							},
						}}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{TagList: []*docdb.Tag{}}, nil
					},
				},
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withDBInstanceArn(testDBInstanceArn),
					withStatus(svcapitypes.DocDBInstanceStateCreating),
					withConditions(xpv1.Creating()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBInstancesWithContext: []*fake.CallDescribeDBInstancesWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeDBInstancesInput(instance(
								withExternalName(testDBIdentifier),
								withDBIdentifier(testDBIdentifier),
							)),
							Opts: nil,
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{
								ResourceName: pointer.ToOrNilIfZeroValue(testDBInstanceArn),
							},
						},
					},
				},
			},
		},
		"CreatingState_and_outdated_AvailabilityZone": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBInstancesWithContext: func(c context.Context, ddi *docdb.DescribeDBInstancesInput, o []request.Option) (*docdb.DescribeDBInstancesOutput, error) {
						return &docdb.DescribeDBInstancesOutput{DBInstances: []*docdb.DBInstance{
							{
								DBInstanceStatus:     pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateCreating),
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
								Endpoint:             &docdb.Endpoint{Address: pointer.ToOrNilIfZeroValue(testAddress), Port: pointer.ToIntAsInt64(testPort)},
								AvailabilityZone:     pointer.ToOrNilIfZeroValue(testAvailabilityZone),
							},
						}}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{TagList: []*docdb.Tag{}}, nil
					},
				},
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withAvailabilityZone(testOtherAvailabilityZone),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withConditions(xpv1.Creating()),
					withStatus(svcapitypes.DocDBInstanceStateCreating),
					withEndpoint(&svcapitypes.Endpoint{
						Address: pointer.ToOrNilIfZeroValue(testAddress),
						Port:    pointer.ToIntAsInt64(testPort),
					}),
					withAvailabilityZone(testOtherAvailabilityZone),
				),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: generateConnectionDetails(testAddress, testPort),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBInstancesWithContext: []*fake.CallDescribeDBInstancesWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeDBInstancesInput(instance(
								withExternalName(testDBIdentifier),
								withDBIdentifier(testDBIdentifier),
							)),
							Opts: nil,
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
		"CreatingState_and_outdated_AutoMinorVersionUpgrade": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBInstancesWithContext: func(c context.Context, ddi *docdb.DescribeDBInstancesInput, o []request.Option) (*docdb.DescribeDBInstancesOutput, error) {
						return &docdb.DescribeDBInstancesOutput{DBInstances: []*docdb.DBInstance{
							{
								DBInstanceStatus:        pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateCreating),
								DBInstanceIdentifier:    pointer.ToOrNilIfZeroValue(testDBIdentifier),
								Endpoint:                &docdb.Endpoint{Address: pointer.ToOrNilIfZeroValue(testAddress), Port: pointer.ToIntAsInt64(testPort)},
								AutoMinorVersionUpgrade: pointer.ToOrNilIfZeroValue(true),
							},
						}}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{TagList: []*docdb.Tag{}}, nil
					},
				},
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withAutoMinorVersionUpgrade(false),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withStatus(svcapitypes.DocDBInstanceStateCreating),
					withConditions(xpv1.Creating()),
					withEndpoint(&svcapitypes.Endpoint{
						Address: pointer.ToOrNilIfZeroValue(testAddress),
						Port:    pointer.ToIntAsInt64(testPort),
					}),
					withAutoMinorVersionUpgrade(false),
				),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  false,
					ConnectionDetails: generateConnectionDetails(testAddress, testPort),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBInstancesWithContext: []*fake.CallDescribeDBInstancesWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeDBInstancesInput(instance(
								withExternalName(testDBIdentifier),
								withDBIdentifier(testDBIdentifier),
							)),
							Opts: nil,
						},
					},
				},
			},
		},
		"CreatingState_and_outdated_CACertificateIdentifier": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBInstancesWithContext: func(c context.Context, ddi *docdb.DescribeDBInstancesInput, o []request.Option) (*docdb.DescribeDBInstancesOutput, error) {
						return &docdb.DescribeDBInstancesOutput{DBInstances: []*docdb.DBInstance{
							{
								DBInstanceStatus:        pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateCreating),
								DBInstanceIdentifier:    pointer.ToOrNilIfZeroValue(testDBIdentifier),
								Endpoint:                &docdb.Endpoint{Address: pointer.ToOrNilIfZeroValue(testAddress), Port: pointer.ToIntAsInt64(testPort)},
								CACertificateIdentifier: pointer.ToOrNilIfZeroValue(testCACertificateIdentifier),
							},
						}}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{TagList: []*docdb.Tag{}}, nil
					},
				},
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withCACertificateIdentifier(testOtherCACertificateIdentifier),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withStatus(svcapitypes.DocDBInstanceStateCreating),
					withConditions(xpv1.Creating()),
					withEndpoint(&svcapitypes.Endpoint{
						Address: pointer.ToOrNilIfZeroValue(testAddress),
						Port:    pointer.ToIntAsInt64(testPort),
					}),
					withCACertificateIdentifier(testOtherCACertificateIdentifier),
					withStatusCACertificateIdentifier(testCACertificateIdentifier),
				),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  false,
					ConnectionDetails: generateConnectionDetails(testAddress, testPort),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBInstancesWithContext: []*fake.CallDescribeDBInstancesWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeDBInstancesInput(instance(
								withExternalName(testDBIdentifier),
								withDBIdentifier(testDBIdentifier),
							)),
							Opts: nil,
						},
					},
				},
			},
		},
		"CreatingState_and_outdated_DBInstanceClass": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBInstancesWithContext: func(c context.Context, ddi *docdb.DescribeDBInstancesInput, o []request.Option) (*docdb.DescribeDBInstancesOutput, error) {
						return &docdb.DescribeDBInstancesOutput{DBInstances: []*docdb.DBInstance{
							{
								DBInstanceStatus:     pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateCreating),
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
								Endpoint:             &docdb.Endpoint{Address: pointer.ToOrNilIfZeroValue(testAddress), Port: pointer.ToIntAsInt64(testPort)},
								DBInstanceClass:      pointer.ToOrNilIfZeroValue(testDBInstanceClass),
							},
						}}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{TagList: []*docdb.Tag{}}, nil
					},
				},
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withDBInstanceClass(testOtherDBInstanceClass),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withStatus(svcapitypes.DocDBInstanceStateCreating),
					withConditions(xpv1.Creating()),
					withEndpoint(&svcapitypes.Endpoint{
						Address: pointer.ToOrNilIfZeroValue(testAddress),
						Port:    pointer.ToIntAsInt64(testPort),
					}),
					withDBInstanceClass(testOtherDBInstanceClass),
				),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  false,
					ConnectionDetails: generateConnectionDetails(testAddress, testPort),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBInstancesWithContext: []*fake.CallDescribeDBInstancesWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeDBInstancesInput(instance(
								withExternalName(testDBIdentifier),
								withDBIdentifier(testDBIdentifier),
							)),
							Opts: nil,
						},
					},
				},
			},
		},
		"CreatingState_and_outdated_PreferredMaintenanceWindow": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBInstancesWithContext: func(c context.Context, ddi *docdb.DescribeDBInstancesInput, o []request.Option) (*docdb.DescribeDBInstancesOutput, error) {
						return &docdb.DescribeDBInstancesOutput{DBInstances: []*docdb.DBInstance{
							{
								DBInstanceStatus:           pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateCreating),
								DBInstanceIdentifier:       pointer.ToOrNilIfZeroValue(testDBIdentifier),
								Endpoint:                   &docdb.Endpoint{Address: pointer.ToOrNilIfZeroValue(testAddress), Port: pointer.ToIntAsInt64(testPort)},
								PreferredMaintenanceWindow: pointer.ToOrNilIfZeroValue(testPreferredMaintenanceWindow),
							},
						}}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{TagList: []*docdb.Tag{}}, nil
					},
				},
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withPreferredMaintenanceWindow(testOtherPreferredMaintenanceWindow),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withStatus(svcapitypes.DocDBInstanceStateCreating),
					withConditions(xpv1.Creating()),
					withEndpoint(&svcapitypes.Endpoint{
						Address: pointer.ToOrNilIfZeroValue(testAddress),
						Port:    pointer.ToIntAsInt64(testPort),
					}),
					withPreferredMaintenanceWindow(testOtherPreferredMaintenanceWindow),
				),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  false,
					ConnectionDetails: generateConnectionDetails(testAddress, testPort),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBInstancesWithContext: []*fake.CallDescribeDBInstancesWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeDBInstancesInput(instance(
								withExternalName(testDBIdentifier),
								withDBIdentifier(testDBIdentifier),
							)),
							Opts: nil,
						},
					},
				},
			},
		},
		"CreatingState_and_outdated_PromotionTier": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBInstancesWithContext: func(c context.Context, ddi *docdb.DescribeDBInstancesInput, o []request.Option) (*docdb.DescribeDBInstancesOutput, error) {
						return &docdb.DescribeDBInstancesOutput{DBInstances: []*docdb.DBInstance{
							{
								DBInstanceStatus:     pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateCreating),
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
								Endpoint:             &docdb.Endpoint{Address: pointer.ToOrNilIfZeroValue(testAddress), Port: pointer.ToIntAsInt64(testPort)},
								PromotionTier:        pointer.ToIntAsInt64(testPromotionTier),
							},
						}}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{TagList: []*docdb.Tag{}}, nil
					},
				},
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withPromotionTier(testOtherPromotionTier),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withStatus(svcapitypes.DocDBInstanceStateCreating),
					withConditions(xpv1.Creating()),
					withEndpoint(&svcapitypes.Endpoint{
						Address: pointer.ToOrNilIfZeroValue(testAddress),
						Port:    pointer.ToIntAsInt64(testPort),
					}),
					withPromotionTier(testOtherPromotionTier),
				),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  false,
					ConnectionDetails: generateConnectionDetails(testAddress, testPort),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBInstancesWithContext: []*fake.CallDescribeDBInstancesWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeDBInstancesInput(instance(
								withExternalName(testDBIdentifier),
								withDBIdentifier(testDBIdentifier),
							)),
							Opts: nil,
						},
					},
				},
			},
		},
		"CreatingState_and_outdated_tags": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBInstancesWithContext: func(c context.Context, ddi *docdb.DescribeDBInstancesInput, o []request.Option) (*docdb.DescribeDBInstancesOutput, error) {
						return &docdb.DescribeDBInstancesOutput{DBInstances: []*docdb.DBInstance{
							{
								DBInstanceStatus:     pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateCreating),
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
								DBInstanceArn:        pointer.ToOrNilIfZeroValue(testDBInstanceArn),
								Endpoint:             &docdb.Endpoint{Address: pointer.ToOrNilIfZeroValue(testAddress), Port: pointer.ToIntAsInt64(testPort)},
							},
						}}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{TagList: []*docdb.Tag{
							{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testTagValue)},
						}}, nil
					},
				},
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withTags(&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testOtherTagKey), Value: pointer.ToOrNilIfZeroValue(testOtherTagValue)}),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withDBInstanceArn(testDBInstanceArn),
					withStatus(svcapitypes.DocDBInstanceStateCreating),
					withConditions(xpv1.Creating()),
					withEndpoint(&svcapitypes.Endpoint{
						Address: pointer.ToOrNilIfZeroValue(testAddress),
						Port:    pointer.ToIntAsInt64(testPort),
					}),
					withTags(&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testOtherTagKey), Value: pointer.ToOrNilIfZeroValue(testOtherTagValue)}),
				),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  false,
					ConnectionDetails: generateConnectionDetails(testAddress, testPort),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBInstancesWithContext: []*fake.CallDescribeDBInstancesWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeDBInstancesInput(instance(
								withExternalName(testDBIdentifier),
								withDBIdentifier(testDBIdentifier),
							)),
							Opts: nil,
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{ResourceName: pointer.ToOrNilIfZeroValue(testDBInstanceArn)},
						},
					},
				},
			},
		},
		"CreatingState_and_same_tags": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBInstancesWithContext: func(c context.Context, ddi *docdb.DescribeDBInstancesInput, o []request.Option) (*docdb.DescribeDBInstancesOutput, error) {
						return &docdb.DescribeDBInstancesOutput{DBInstances: []*docdb.DBInstance{
							{
								DBInstanceStatus:     pointer.ToOrNilIfZeroValue(svcapitypes.DocDBInstanceStateCreating),
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
								DBInstanceArn:        pointer.ToOrNilIfZeroValue(testDBInstanceArn),
								Endpoint:             &docdb.Endpoint{Address: pointer.ToOrNilIfZeroValue(testAddress), Port: pointer.ToIntAsInt64(testPort)},
							},
						}}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{TagList: []*docdb.Tag{
							{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testTagValue)},
						}}, nil
					},
				},
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withTags(&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testTagValue)}),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBIdentifier),
					withDBIdentifier(testDBIdentifier),
					withDBInstanceArn(testDBInstanceArn),
					withStatus(svcapitypes.DocDBInstanceStateCreating),
					withConditions(xpv1.Creating()),
					withEndpoint(&svcapitypes.Endpoint{
						Address: pointer.ToOrNilIfZeroValue(testAddress),
						Port:    pointer.ToIntAsInt64(testPort),
					}),
					withTags(&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testTagValue)}),
				),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: generateConnectionDetails(testAddress, testPort),
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBInstancesWithContext: []*fake.CallDescribeDBInstancesWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeDBInstancesInput(instance(
								withExternalName(testDBIdentifier),
								withDBIdentifier(testDBIdentifier),
							)),
							Opts: nil,
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{
								ResourceName: pointer.ToOrNilIfZeroValue(testDBInstanceArn),
							},
						},
					},
				},
			},
		},
		"EmptyDescribeInstancesOutput": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBInstancesWithContext: func(c context.Context, ddi *docdb.DescribeDBInstancesInput, o []request.Option) (*docdb.DescribeDBInstancesOutput, error) {
						return &docdb.DescribeDBInstancesOutput{DBInstances: []*docdb.DBInstance{}}, nil
					},
				},
				cr: instance(
					withDBIdentifier(testDBIdentifier),
					withExternalName(testDBIdentifier),
				),
			},
			want: want{
				cr: instance(
					withDBIdentifier(testDBIdentifier),
					withExternalName(testDBIdentifier),
				),
				result: managed.ExternalObservation{},
				docdb: fake.MockDocDBClientCall{
					DescribeDBInstancesWithContext: []*fake.CallDescribeDBInstancesWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeDBInstancesInput(instance(
								withExternalName(testDBIdentifier),
								withDBIdentifier(testDBIdentifier),
							)),
							Opts: nil,
						},
					},
				},
			},
		},
		"ErrorDescribeDBInstances": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBInstancesWithContext: func(c context.Context, ddi *docdb.DescribeDBInstancesInput, o []request.Option) (*docdb.DescribeDBInstancesOutput, error) {
						return nil, errors.New(testErrDescribeDBInstancesFailed)
					},
				},
				cr: instance(
					withDBIdentifier(testDBIdentifier),
					withExternalName(testDBIdentifier),
				),
			},
			want: want{
				cr: instance(
					withDBIdentifier(testDBIdentifier),
					withExternalName(testDBIdentifier),
				),
				result: managed.ExternalObservation{},
				err:    errorutils.Wrap(cpresource.Ignore(IsNotFound, errors.New(testErrDescribeDBInstancesFailed)), errDescribe),
				docdb: fake.MockDocDBClientCall{
					DescribeDBInstancesWithContext: []*fake.CallDescribeDBInstancesWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeDBInstancesInput(instance(
								withExternalName(testDBIdentifier),
								withDBIdentifier(testDBIdentifier),
							)),
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
		cr     *svcapitypes.DBInstance
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
				docdb: &fake.MockDocDBClient{
					MockCreateDBInstanceWithContext: func(c context.Context, cdi *docdb.CreateDBInstanceInput, opts []request.Option) (*docdb.CreateDBInstanceOutput, error) {
						return &docdb.CreateDBInstanceOutput{
							DBInstance: &docdb.DBInstance{
								AutoMinorVersionUpgrade: pointer.ToOrNilIfZeroValue(true),
								AvailabilityZone:        &testAvailabilityZone,
								DBInstanceClass:         &testDBInstanceClass,
								DBInstanceIdentifier:    pointer.ToOrNilIfZeroValue(testDBIdentifier),
								Endpoint: &docdb.Endpoint{
									Address:      pointer.ToOrNilIfZeroValue(testAddress),
									HostedZoneId: pointer.ToOrNilIfZeroValue(testHostedZone),
									Port:         pointer.ToIntAsInt64(testPort),
								},
								PreferredMaintenanceWindow: &testPreferredMaintenanceWindow,
								PromotionTier:              pointer.ToIntAsInt64(testPromotionTier),
							},
						}, nil
					},
				},
				cr: instance(
					withExternalName(testDBIdentifier),
					withAutoMinorVersionUpgrade(true),
					withAvailabilityZone(testAvailabilityZone),
					withCACertificateIdentifier(testCACertificateIdentifier),
					withDBInstanceClass(testDBInstanceClass),
					withPreferredMaintenanceWindow(testPreferredMaintenanceWindow),
					withPromotionTier(testPromotionTier),
				),
			},
			want: want{
				cr: instance(
					withDBIdentifier(testDBIdentifier),
					withExternalName(testDBIdentifier),
					withAutoMinorVersionUpgrade(true),
					withAvailabilityZone(testAvailabilityZone),
					withCACertificateIdentifier(testCACertificateIdentifier),
					withDBInstanceClass(testDBInstanceClass),
					withPreferredMaintenanceWindow(testPreferredMaintenanceWindow),
					withPromotionTier(testPromotionTier),
					withConditions(xpv1.Creating()),
					withEndpoint(&svcapitypes.Endpoint{
						Address:      pointer.ToOrNilIfZeroValue(testAddress),
						HostedZoneID: pointer.ToOrNilIfZeroValue(testHostedZone),
						Port:         pointer.ToIntAsInt64(testPort),
					}),
				),
				result: managed.ExternalCreation{
					ConnectionDetails: generateConnectionDetails(testAddress, testPort),
				},
				docdb: fake.MockDocDBClientCall{
					CreateDBInstanceWithContext: []*fake.CallCreateDBInstanceWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.CreateDBInstanceInput{
								DBInstanceIdentifier:       pointer.ToOrNilIfZeroValue(testDBIdentifier),
								AutoMinorVersionUpgrade:    pointer.ToOrNilIfZeroValue(true),
								AvailabilityZone:           pointer.ToOrNilIfZeroValue(testAvailabilityZone),
								DBInstanceClass:            pointer.ToOrNilIfZeroValue(testDBInstanceClass),
								PreferredMaintenanceWindow: pointer.ToOrNilIfZeroValue(testPreferredMaintenanceWindow),
								PromotionTier:              pointer.ToIntAsInt64(testPromotionTier),
							},
						},
					},
				},
			},
		},
		"ErrorCreate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockCreateDBInstanceWithContext: func(c context.Context, cdi *docdb.CreateDBInstanceInput, opts []request.Option) (*docdb.CreateDBInstanceOutput, error) {
						return nil, errors.New(testErrCreateDBInstanceFailed)
					},
				},
				cr: instance(
					withDBIdentifier(testDBIdentifier),
					withExternalName(testDBIdentifier),
				),
			},
			want: want{
				cr: instance(
					withDBIdentifier(testDBIdentifier),
					withExternalName(testDBIdentifier),
					withConditions(xpv1.Creating()),
				),
				result: managed.ExternalCreation{},
				err:    errors.Wrap(errors.New(testErrCreateDBInstanceFailed), errCreate),
				docdb: fake.MockDocDBClientCall{
					CreateDBInstanceWithContext: []*fake.CallCreateDBInstanceWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.CreateDBInstanceInput{
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
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
		cr    *svcapitypes.DBInstance
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
					MockDeleteDBInstanceWithContext: func(c context.Context, ddi *docdb.DeleteDBInstanceInput, o []request.Option) (*docdb.DeleteDBInstanceOutput, error) {
						return &docdb.DeleteDBInstanceOutput{
							DBInstance: &docdb.DBInstance{
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
							},
						}, nil
					},
				},
				cr: instance(
					withDBIdentifier(testDBIdentifier),
					withExternalName(testDBIdentifier),
				),
			},
			want: want{
				cr: instance(
					withDBIdentifier(testDBIdentifier),
					withExternalName(testDBIdentifier),
					withConditions(xpv1.Deleting()),
				),
				docdb: fake.MockDocDBClientCall{
					DeleteDBInstanceWithContext: []*fake.CallDeleteDBInstanceWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DeleteDBInstanceInput{
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
							},
						},
					},
				},
			},
		},
		"ErrorDelete": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDeleteDBInstanceWithContext: func(c context.Context, ddi *docdb.DeleteDBInstanceInput, o []request.Option) (*docdb.DeleteDBInstanceOutput, error) {
						return nil, errors.New(testErrDeleteDBInstanceFailed)
					},
				},
				cr: instance(
					withDBIdentifier(testDBIdentifier),
					withExternalName(testDBIdentifier),
				),
			},
			want: want{
				cr: instance(
					withDBIdentifier(testDBIdentifier),
					withExternalName(testDBIdentifier),
					withConditions(xpv1.Deleting()),
				),
				err: errors.Wrap(errors.New(testErrDeleteDBInstanceFailed), errDelete),
				docdb: fake.MockDocDBClientCall{
					DeleteDBInstanceWithContext: []*fake.CallDeleteDBInstanceWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DeleteDBInstanceInput{
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
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
		cr     *svcapitypes.DBInstance
		result managed.ExternalUpdate
		err    error
		docdb  fake.MockDocDBClientCall
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulUpdate_no_tags": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockModifyDBInstanceWithContext: func(c context.Context, mdi *docdb.ModifyDBInstanceInput, o []request.Option) (*docdb.ModifyDBInstanceOutput, error) {
						return &docdb.ModifyDBInstanceOutput{
							DBInstance: &docdb.DBInstance{
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
								DBInstanceArn:        pointer.ToOrNilIfZeroValue(testDBInstanceArn),
							},
						}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{}, nil
					},
				},
				cr: instance(
					withDBIdentifier(testDBIdentifier),
					withDBInstanceArn(testDBInstanceArn),
					withExternalName(testDBIdentifier),
				),
			},
			want: want{
				cr: instance(
					withDBIdentifier(testDBIdentifier),
					withDBInstanceArn(testDBInstanceArn),
					withExternalName(testDBIdentifier),
				),
				result: managed.ExternalUpdate{},
				docdb: fake.MockDocDBClientCall{
					ModifyDBInstanceWithContext: []*fake.CallModifyDBInstanceWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.ModifyDBInstanceInput{
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
							},
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{
								ResourceName: pointer.ToOrNilIfZeroValue(testDBInstanceArn),
							},
						},
					},
				},
			},
		},
		"SuccessfulUpdate_same_tags": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockModifyDBInstanceWithContext: func(c context.Context, mdi *docdb.ModifyDBInstanceInput, o []request.Option) (*docdb.ModifyDBInstanceOutput, error) {
						return &docdb.ModifyDBInstanceOutput{
							DBInstance: &docdb.DBInstance{
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
								DBInstanceArn:        pointer.ToOrNilIfZeroValue(testDBInstanceArn),
							},
						}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{
							TagList: []*docdb.Tag{
								{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testTagValue)},
							},
						}, nil
					},
				},
				cr: instance(
					withDBIdentifier(testDBIdentifier),
					withExternalName(testDBIdentifier),
					withDBInstanceArn(testDBInstanceArn),
					withTags(&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testTagValue)}),
				),
			},
			want: want{
				cr: instance(
					withDBIdentifier(testDBIdentifier),
					withExternalName(testDBIdentifier),
					withDBInstanceArn(testDBInstanceArn),
					withTags(&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testTagValue)}),
				),
				result: managed.ExternalUpdate{},
				docdb: fake.MockDocDBClientCall{
					ModifyDBInstanceWithContext: []*fake.CallModifyDBInstanceWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.ModifyDBInstanceInput{
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
							},
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{
								ResourceName: pointer.ToOrNilIfZeroValue(testDBInstanceArn),
							},
						},
					},
				},
			},
		},
		"SuccessfulUpdate_add_tag": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockModifyDBInstanceWithContext: func(c context.Context, mdi *docdb.ModifyDBInstanceInput, o []request.Option) (*docdb.ModifyDBInstanceOutput, error) {
						return &docdb.ModifyDBInstanceOutput{
							DBInstance: &docdb.DBInstance{
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
								DBInstanceArn:        pointer.ToOrNilIfZeroValue(testDBInstanceArn),
							},
						}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{
							TagList: []*docdb.Tag{
								{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testTagValue)},
							},
						}, nil
					},
					MockAddTagsToResource: func(attri *docdb.AddTagsToResourceInput) (*docdb.AddTagsToResourceOutput, error) {
						return &docdb.AddTagsToResourceOutput{}, nil
					},
				},
				cr: instance(
					withDBIdentifier(testDBIdentifier),
					withExternalName(testDBIdentifier),
					withDBInstanceArn(testDBInstanceArn),
					withTags(
						&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testTagValue)},
						&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testOtherTagKey), Value: pointer.ToOrNilIfZeroValue(testOtherTagValue)},
					),
				),
			},
			want: want{
				cr: instance(
					withDBIdentifier(testDBIdentifier),
					withExternalName(testDBIdentifier),
					withDBInstanceArn(testDBInstanceArn),
					withTags(
						&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testTagValue)},
						&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testOtherTagKey), Value: pointer.ToOrNilIfZeroValue(testOtherTagValue)},
					),
				),
				result: managed.ExternalUpdate{},
				docdb: fake.MockDocDBClientCall{
					ModifyDBInstanceWithContext: []*fake.CallModifyDBInstanceWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.ModifyDBInstanceInput{
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
							},
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{
								ResourceName: pointer.ToOrNilIfZeroValue(testDBInstanceArn),
							},
						},
					},
					AddTagsToResource: []*fake.CallAddTagsToResource{
						{
							I: &docdb.AddTagsToResourceInput{
								ResourceName: pointer.ToOrNilIfZeroValue(testDBInstanceArn),
								Tags: []*docdb.Tag{
									{Key: pointer.ToOrNilIfZeroValue(testOtherTagKey), Value: pointer.ToOrNilIfZeroValue(testOtherTagValue)},
								},
							},
						},
					},
				},
			},
		},
		"SuccessfulUpdate_remove_tag": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockModifyDBInstanceWithContext: func(c context.Context, mdi *docdb.ModifyDBInstanceInput, o []request.Option) (*docdb.ModifyDBInstanceOutput, error) {
						return &docdb.ModifyDBInstanceOutput{
							DBInstance: &docdb.DBInstance{
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
								DBInstanceArn:        pointer.ToOrNilIfZeroValue(testDBInstanceArn),
							},
						}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{
							TagList: []*docdb.Tag{
								{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testTagValue)},
							},
						}, nil
					},
					MockRemoveTagsFromResource: func(rtfri *docdb.RemoveTagsFromResourceInput) (*docdb.RemoveTagsFromResourceOutput, error) {
						return &docdb.RemoveTagsFromResourceOutput{}, nil
					},
				},
				cr: instance(
					withDBIdentifier(testDBIdentifier),
					withExternalName(testDBIdentifier),
					withDBInstanceArn(testDBInstanceArn),
				),
			},
			want: want{
				cr: instance(
					withDBIdentifier(testDBIdentifier),
					withExternalName(testDBIdentifier),
					withDBInstanceArn(testDBInstanceArn),
				),
				result: managed.ExternalUpdate{},
				docdb: fake.MockDocDBClientCall{
					ModifyDBInstanceWithContext: []*fake.CallModifyDBInstanceWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.ModifyDBInstanceInput{
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
							},
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{
								ResourceName: pointer.ToOrNilIfZeroValue(testDBInstanceArn),
							},
						},
					},
					RemoveTagsFromResource: []*fake.CallRemoveTagsFromResource{
						{
							I: &docdb.RemoveTagsFromResourceInput{
								ResourceName: pointer.ToOrNilIfZeroValue(testDBInstanceArn),
								TagKeys:      []*string{pointer.ToOrNilIfZeroValue(testTagKey)},
							},
						},
					},
				},
			},
		},
		"SuccessfulUpdate_overwrite_tag": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockModifyDBInstanceWithContext: func(c context.Context, mdi *docdb.ModifyDBInstanceInput, o []request.Option) (*docdb.ModifyDBInstanceOutput, error) {
						return &docdb.ModifyDBInstanceOutput{
							DBInstance: &docdb.DBInstance{
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
								DBInstanceArn:        pointer.ToOrNilIfZeroValue(testDBInstanceArn),
							},
						}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{
							TagList: []*docdb.Tag{
								{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testTagValue)},
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
					withDBIdentifier(testDBIdentifier),
					withExternalName(testDBIdentifier),
					withDBInstanceArn(testDBInstanceArn),
					withTags(
						&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testOtherTagValue)},
					),
				),
			},
			want: want{
				cr: instance(
					withDBIdentifier(testDBIdentifier),
					withExternalName(testDBIdentifier),
					withDBInstanceArn(testDBInstanceArn),
					withTags(
						&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testOtherTagValue)},
					),
				),
				result: managed.ExternalUpdate{},
				docdb: fake.MockDocDBClientCall{
					ModifyDBInstanceWithContext: []*fake.CallModifyDBInstanceWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.ModifyDBInstanceInput{
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
							},
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{
								ResourceName: pointer.ToOrNilIfZeroValue(testDBInstanceArn),
							},
						},
					},
					RemoveTagsFromResource: []*fake.CallRemoveTagsFromResource{
						{
							I: &docdb.RemoveTagsFromResourceInput{
								ResourceName: pointer.ToOrNilIfZeroValue(testDBInstanceArn),
								TagKeys:      []*string{pointer.ToOrNilIfZeroValue(testTagKey)},
							},
						},
					},
					AddTagsToResource: []*fake.CallAddTagsToResource{
						{
							I: &docdb.AddTagsToResourceInput{
								ResourceName: pointer.ToOrNilIfZeroValue(testDBInstanceArn),
								Tags: []*docdb.Tag{
									{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testOtherTagValue)},
								},
							},
						},
					},
				},
			},
		},
		"SuccessfulUpdate_remove_and_add_tag": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockModifyDBInstanceWithContext: func(c context.Context, mdi *docdb.ModifyDBInstanceInput, o []request.Option) (*docdb.ModifyDBInstanceOutput, error) {
						return &docdb.ModifyDBInstanceOutput{
							DBInstance: &docdb.DBInstance{
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
								DBInstanceArn:        pointer.ToOrNilIfZeroValue(testDBInstanceArn),
							},
						}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{
							TagList: []*docdb.Tag{
								{Key: pointer.ToOrNilIfZeroValue(testTagKey), Value: pointer.ToOrNilIfZeroValue(testTagValue)},
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
					withDBIdentifier(testDBIdentifier),
					withExternalName(testDBIdentifier),
					withDBInstanceArn(testDBInstanceArn),
					withTags(
						&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testOtherTagKey), Value: pointer.ToOrNilIfZeroValue(testOtherTagValue)},
					),
				),
			},
			want: want{
				cr: instance(
					withDBIdentifier(testDBIdentifier),
					withExternalName(testDBIdentifier),
					withDBInstanceArn(testDBInstanceArn),
					withTags(
						&svcapitypes.Tag{Key: pointer.ToOrNilIfZeroValue(testOtherTagKey), Value: pointer.ToOrNilIfZeroValue(testOtherTagValue)},
					),
				),
				result: managed.ExternalUpdate{},
				docdb: fake.MockDocDBClientCall{
					ModifyDBInstanceWithContext: []*fake.CallModifyDBInstanceWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.ModifyDBInstanceInput{
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
							},
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{
								ResourceName: pointer.ToOrNilIfZeroValue(testDBInstanceArn),
							},
						},
					},
					RemoveTagsFromResource: []*fake.CallRemoveTagsFromResource{
						{
							I: &docdb.RemoveTagsFromResourceInput{
								ResourceName: pointer.ToOrNilIfZeroValue(testDBInstanceArn),
								TagKeys:      []*string{pointer.ToOrNilIfZeroValue(testTagKey)},
							},
						},
					},
					AddTagsToResource: []*fake.CallAddTagsToResource{
						{
							I: &docdb.AddTagsToResourceInput{
								ResourceName: pointer.ToOrNilIfZeroValue(testDBInstanceArn),
								Tags: []*docdb.Tag{
									{Key: pointer.ToOrNilIfZeroValue(testOtherTagKey), Value: pointer.ToOrNilIfZeroValue(testOtherTagValue)},
								},
							},
						},
					},
				},
			},
		},
		"ErrorModifyDBInstanceWithContext": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockModifyDBInstanceWithContext: func(c context.Context, mdi *docdb.ModifyDBInstanceInput, o []request.Option) (*docdb.ModifyDBInstanceOutput, error) {
						return nil, errors.New(testErrModifyDBInstanceFailed)
					},
				},
				cr: instance(
					withDBIdentifier(testDBIdentifier),
					withExternalName(testDBIdentifier),
				),
			},
			want: want{
				cr: instance(
					withDBIdentifier(testDBIdentifier),
					withExternalName(testDBIdentifier),
				),
				err: errors.Wrap(errors.New(testErrModifyDBInstanceFailed), errUpdate),
				docdb: fake.MockDocDBClientCall{
					ModifyDBInstanceWithContext: []*fake.CallModifyDBInstanceWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.ModifyDBInstanceInput{
								DBInstanceIdentifier: pointer.ToOrNilIfZeroValue(testDBIdentifier),
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
