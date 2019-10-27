package rds

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"

	"github.com/crossplaneio/stack-aws/apis/database/v1alpha2"
	"github.com/crossplaneio/stack-aws/apis/database/v1beta1"
	aws "github.com/crossplaneio/stack-aws/pkg/clients"
)

var (
	dbName           = "example-name"
	characterSetName = "utf8"
)

func TestCreatePatch(t *testing.T) {
	dbSubnetGroupName := "example-subnet"

	type args struct {
		db *rds.DBInstance
		p  *v1beta1.RDSInstanceParameters
	}

	type want struct {
		patch *v1beta1.RDSInstanceParameters
	}

	cases := map[string]struct {
		args
		want
	}{
		"SomeFields": {
			args: args{
				db: &rds.DBInstance{
					AllocatedStorage: aws.Int64(20),
					CharacterSetName: &characterSetName,
					DBName:           &dbName,
				},
				p: &v1beta1.RDSInstanceParameters{
					AllocatedStorage: aws.IntAddress(aws.Int64(30)),
					CharacterSetName: &characterSetName,
					DBName:           &dbName,
				},
			},
			want: want{
				patch: &v1beta1.RDSInstanceParameters{
					AllocatedStorage: aws.IntAddress(aws.Int64(30)),
				},
			},
		},
		"IgnoresRefs": {
			args: args{
				db: &rds.DBInstance{
					DBName:        &dbName,
					DBSubnetGroup: &rds.DBSubnetGroup{DBSubnetGroupName: &dbSubnetGroupName},
				},
				p: &v1beta1.RDSInstanceParameters{
					DBName:            &dbName,
					DBSubnetGroupName: &dbSubnetGroupName,
					DBSubnetGroupNameRef: &v1beta1.DBSubnetGroupNameReferencerForRDSInstance{
						DBSubnetGroupNameReferencer: v1alpha2.DBSubnetGroupNameReferencer{
							LocalObjectReference: corev1.LocalObjectReference{Name: "coolgroup"},
						},
					},
				},
			},
			want: want{
				patch: &v1beta1.RDSInstanceParameters{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			result, _ := CreatePatch(tc.args.db, tc.args.p)
			if diff := cmp.Diff(tc.want.patch, result); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
