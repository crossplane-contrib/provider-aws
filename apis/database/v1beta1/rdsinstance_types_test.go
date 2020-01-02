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

package v1beta1

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	"github.com/crossplaneio/crossplane-runtime/pkg/test"

	aws "github.com/crossplaneio/stack-aws/pkg/clients"
)

var _ resource.AttributeReferencer = (*VPCSecurityGroupIDReferencerForRDSInstance)(nil)
var _ resource.AttributeReferencer = (*DBSubnetGroupNameReferencerForRDSInstance)(nil)

func TestVPCSecurityGroupIDReferencerForRDSInstance(t *testing.T) {
	value := "cool"

	type args struct {
		res   resource.CanReference
		value string
	}
	type want struct {
		res resource.CanReference
		err error
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"AssignWrongType": {
			reason: "Assign should return an error when the supplied CanReference does not contain an *RDSInstance.",
			args: args{
				res: nil,
			},
			want: want{
				err: errors.New(errResourceIsNotRDSInstance),
			},
		},
		"AssignSuccessful": {
			reason: "Assign should append to Spec.VPCSecurityGroupIDs.",
			args: args{
				res:   &RDSInstance{},
				value: value,
			},
			want: want{
				res: &RDSInstance{
					Spec: RDSInstanceSpec{
						ForProvider: RDSInstanceParameters{VPCSecurityGroupIDs: []string{value}},
					},
				},
			},
		},
		"AssignNoOp": {
			reason: "Assign should not append existing values to Spec.VPCSecurityGroupIDs.",
			args: args{
				res: &RDSInstance{
					Spec: RDSInstanceSpec{
						ForProvider: RDSInstanceParameters{VPCSecurityGroupIDs: []string{value}},
					},
				},
				value: value,
			},
			want: want{
				res: &RDSInstance{
					Spec: RDSInstanceSpec{
						ForProvider: RDSInstanceParameters{VPCSecurityGroupIDs: []string{value}},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := &VPCSecurityGroupIDReferencerForRDSInstance{}
			err := r.Assign(tc.args.res, tc.args.value)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\nReason: %s\nAssign(...): -want error, +got error:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.res, tc.args.res, test.EquateErrors()); diff != "" {
				t.Errorf("\nReason: %s\nAssign(...): -want, +got:\n%s", tc.reason, diff)
			}
		})
	}
}

func TestDBSubnetGroupNameReferencerForRDSInstance_AssignInvalidType_ReturnsErr(t *testing.T) {

	r := &DBSubnetGroupNameReferencerForRDSInstance{}
	expectedErr := errors.New(errResourceIsNotRDSInstance)

	err := r.Assign(nil, "mockValue")
	if diff := cmp.Diff(expectedErr, err, test.EquateErrors()); diff != "" {
		t.Errorf("Assign(...): -want error, +got error:\n%s", diff)
	}
}

func TestDBSubnetGroupNameReferencerForRDSInstance_AssignValidType_ReturnsExpected(t *testing.T) {

	r := &DBSubnetGroupNameReferencerForRDSInstance{}
	res := &RDSInstance{}
	var expectedErr error

	err := r.Assign(res, "mockValue")
	if diff := cmp.Diff(expectedErr, err, test.EquateErrors()); diff != "" {
		t.Errorf("Assign(...): -want error, +got error:\n%s", diff)
	}

	if diff := cmp.Diff(res.Spec.ForProvider.DBSubnetGroupName, aws.String("mockValue")); diff != "" {
		t.Errorf("Assign(...): -want value, +got value:\n%s", diff)
	}
}

func TestIAMRoleARNReferencerForRDSInstanceMonitoringRole_AssignInvalidType_ReturnsErr(t *testing.T) {

	r := &IAMRoleARNReferencerForRDSInstanceMonitoringRole{}
	expectedErr := errors.New(errResourceIsNotRDSInstance)

	err := r.Assign(nil, "mockValue")
	if diff := cmp.Diff(expectedErr, err, test.EquateErrors()); diff != "" {
		t.Errorf("Assign(...): -want error, +got error:\n%s", diff)
	}
}

func TestIAMRoleARNReferencerForRDSInstanceMonitoringRole_AssignValidType_ReturnsExpected(t *testing.T) {

	r := &IAMRoleARNReferencerForRDSInstanceMonitoringRole{}
	res := &RDSInstance{}
	var expectedErr error

	err := r.Assign(res, "mockValue")
	if diff := cmp.Diff(expectedErr, err, test.EquateErrors()); diff != "" {
		t.Errorf("Assign(...): -want error, +got error:\n%s", diff)
	}

	if diff := cmp.Diff(res.Spec.ForProvider.MonitoringRoleARN, aws.String("mockValue")); diff != "" {
		t.Errorf("Assign(...): -want value, +got value:\n%s", diff)
	}
}

func TestIAMRoleNameReferencerForRDSInstanceDomainRole_AssignInvalidType_ReturnsErr(t *testing.T) {

	r := &IAMRoleNameReferencerForRDSInstanceDomainRole{}
	expectedErr := errors.New(errResourceIsNotRDSInstance)

	err := r.Assign(nil, "mockValue")
	if diff := cmp.Diff(expectedErr, err, test.EquateErrors()); diff != "" {
		t.Errorf("Assign(...): -want error, +got error:\n%s", diff)
	}
}

func TestIAMRoleNameReferencerForRDSInstanceDomainRole_AssignValidType_ReturnsExpected(t *testing.T) {

	r := &IAMRoleNameReferencerForRDSInstanceDomainRole{}
	res := &RDSInstance{}
	var expectedErr error

	err := r.Assign(res, "mockValue")
	if diff := cmp.Diff(expectedErr, err, test.EquateErrors()); diff != "" {
		t.Errorf("Assign(...): -want error, +got error:\n%s", diff)
	}

	if diff := cmp.Diff(res.Spec.ForProvider.DomainIAMRoleName, aws.String("mockValue")); diff != "" {
		t.Errorf("Assign(...): -want value, +got value:\n%s", diff)
	}
}
