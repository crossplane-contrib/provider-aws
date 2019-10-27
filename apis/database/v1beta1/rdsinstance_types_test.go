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

func TestSecurityGroupIDReferencerForRDSInstance_AssignInvalidType_ReturnsErr(t *testing.T) {

	r := &VPCSecurityGroupIDReferencerForRDSInstance{}
	expectedErr := errors.New(errResourceIsNotRDSInstance)

	err := r.Assign(&struct{ resource.CanReference }{}, "mockValue")
	if diff := cmp.Diff(expectedErr, err, test.EquateErrors()); diff != "" {
		t.Errorf("Assign(...): -want error, +got error:\n%s", diff)
	}
}

func TestVPCSecurityGroupIDReferencerForRDSInstance_AssignValidType_ReturnsExpected(t *testing.T) {

	r := &VPCSecurityGroupIDReferencerForRDSInstance{}
	res := &RDSInstance{}
	var expectedErr error

	err := r.Assign(res, "mockValue")
	if diff := cmp.Diff(expectedErr, err, test.EquateErrors()); diff != "" {
		t.Errorf("Assign(...): -want error, +got error:\n%s", diff)
	}

	if diff := cmp.Diff(res.Spec.ForProvider.VPCSecurityGroupIDs, []string{"mockValue"}); diff != "" {
		t.Errorf("Assign(...): -want value, +got value:\n%s", diff)
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
