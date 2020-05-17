/*
Copyright 2019 The Crossplane Authors.

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

package rds

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"

	"github.com/crossplane/provider-aws/apis/database/v1beta1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
)

var (
	dbName           = "example-name"
	characterSetName = "utf8"
)

func TestCreatePatch(t *testing.T) {
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
		"SameFields": {
			args: args{
				db: &rds.DBInstance{
					AllocatedStorage: aws.Int64(20),
					CharacterSetName: &characterSetName,
					DBName:           &dbName,
				},
				p: &v1beta1.RDSInstanceParameters{
					AllocatedStorage: aws.IntAddress(aws.Int64(20)),
					CharacterSetName: &characterSetName,
					DBName:           &dbName,
				},
			},
			want: want{
				patch: &v1beta1.RDSInstanceParameters{},
			},
		},
		"DifferentFields": {
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

func TestIsUpToDate(t *testing.T) {
	dbSubnetGroupName := "example-subnet"

	type args struct {
		db rds.DBInstance
		p  v1beta1.RDSInstanceParameters
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFields": {
			args: args{
				db: rds.DBInstance{
					AllocatedStorage: aws.Int64(20),
					CharacterSetName: &characterSetName,
					DBName:           &dbName,
				},
				p: v1beta1.RDSInstanceParameters{
					AllocatedStorage: aws.IntAddress(aws.Int64(20)),
					CharacterSetName: &characterSetName,
					DBName:           &dbName,
				},
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				db: rds.DBInstance{
					AllocatedStorage: aws.Int64(20),
					CharacterSetName: &characterSetName,
					DBName:           &dbName,
				},
				p: v1beta1.RDSInstanceParameters{
					AllocatedStorage: aws.IntAddress(aws.Int64(30)),
					CharacterSetName: &characterSetName,
					DBName:           &dbName,
				},
			},
			want: false,
		},
		"IgnoresRefs": {
			args: args{
				db: rds.DBInstance{
					DBName:        &dbName,
					DBSubnetGroup: &rds.DBSubnetGroup{DBSubnetGroupName: &dbSubnetGroupName},
				},
				p: v1beta1.RDSInstanceParameters{
					DBName:               &dbName,
					DBSubnetGroupName:    &dbSubnetGroupName,
					DBSubnetGroupNameRef: &v1alpha1.Reference{Name: "coolgroup"},
				},
			},
			want: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, _ := IsUpToDate(tc.args.p, tc.args.db)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
