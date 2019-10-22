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

package v1alpha2

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	"github.com/crossplaneio/crossplane-runtime/pkg/test"
)

var _ resource.AttributeReferencer = (*VPCIDReferencerForEKSCluster)(nil)
var _ resource.AttributeReferencer = (*IAMRoleARNReferencerForEKSCluster)(nil)
var _ resource.AttributeReferencer = (*SubnetIDReferencerForEKSCluster)(nil)
var _ resource.AttributeReferencer = (*SecurityGroupIDReferencerForEKSCluster)(nil)
var _ resource.AttributeReferencer = (*SecurityGroupIDReferencerForEKSWorkerNodes)(nil)

func TestVPCIDReferencerForEKSCluster_AssignInvalidType_ReturnsErr(t *testing.T) {

	r := &VPCIDReferencerForEKSCluster{}
	expectedErr := errors.New(errResourceIsNotEKSCluster)

	err := r.Assign(&struct{ resource.CanReference }{}, "mockValue")
	if diff := cmp.Diff(expectedErr, err, test.EquateErrors()); diff != "" {
		t.Errorf("Assign(...): -want error, +got error:\n%s", diff)
	}
}

func TestVPCIDReferencerForEKSCluster_AssignValidType_ReturnsExpected(t *testing.T) {

	r := &VPCIDReferencerForEKSCluster{}
	res := &EKSCluster{}
	var expectedErr error

	err := r.Assign(res, "mockValue")
	if diff := cmp.Diff(expectedErr, err, test.EquateErrors()); diff != "" {
		t.Errorf("Assign(...): -want error, +got error:\n%s", diff)
	}

	if diff := cmp.Diff(res.Spec.VPCID, "mockValue"); diff != "" {
		t.Errorf("Assign(...): -want value, +got value:\n%s", diff)
	}
}

func TestIAMRoleARNReferencerForEKSCluster_AssignInvalidType_ReturnsErr(t *testing.T) {

	r := &IAMRoleARNReferencerForEKSCluster{}
	expectedErr := errors.New(errResourceIsNotEKSCluster)

	err := r.Assign(&struct{ resource.CanReference }{}, "mockValue")
	if diff := cmp.Diff(expectedErr, err, test.EquateErrors()); diff != "" {
		t.Errorf("Assign(...): -want error, +got error:\n%s", diff)
	}
}

func TestIAMRoleARNReferencerForEKSCluster_AssignValidType_ReturnsExpected(t *testing.T) {

	r := &IAMRoleARNReferencerForEKSCluster{}
	res := &EKSCluster{}
	var expectedErr error

	err := r.Assign(res, "mockValue")
	if diff := cmp.Diff(expectedErr, err, test.EquateErrors()); diff != "" {
		t.Errorf("Assign(...): -want error, +got error:\n%s", diff)
	}

	if diff := cmp.Diff(res.Spec.RoleARN, "mockValue"); diff != "" {
		t.Errorf("Assign(...): -want value, +got value:\n%s", diff)
	}
}

func TestSubnetIDReferencerForEKSCluster_AssignInvalidType_ReturnsErr(t *testing.T) {

	r := &SubnetIDReferencerForEKSCluster{}
	expectedErr := errors.New(errResourceIsNotEKSCluster)

	err := r.Assign(&struct{ resource.CanReference }{}, "mockValue")
	if diff := cmp.Diff(expectedErr, err, test.EquateErrors()); diff != "" {
		t.Errorf("Assign(...): -want error, +got error:\n%s", diff)
	}
}

func TestSubnetIDReferencerForEKSCluster_AssignValidType_ReturnsExpected(t *testing.T) {

	r := &SubnetIDReferencerForEKSCluster{}
	res := &EKSCluster{}
	var expectedErr error

	err := r.Assign(res, "mockValue")
	if diff := cmp.Diff(expectedErr, err, test.EquateErrors()); diff != "" {
		t.Errorf("Assign(...): -want error, +got error:\n%s", diff)
	}

	if diff := cmp.Diff(res.Spec.SubnetIDs, []string{"mockValue"}); diff != "" {
		t.Errorf("Assign(...): -want value, +got value:\n%s", diff)
	}
}

func TestSecurityGroupIDReferencerForEKSCluster_AssignInvalidType_ReturnsErr(t *testing.T) {

	r := &SecurityGroupIDReferencerForEKSCluster{}
	expectedErr := errors.New(errResourceIsNotEKSCluster)

	err := r.Assign(&struct{ resource.CanReference }{}, "mockValue")
	if diff := cmp.Diff(expectedErr, err, test.EquateErrors()); diff != "" {
		t.Errorf("Assign(...): -want error, +got error:\n%s", diff)
	}
}

func TestSecurityGroupIDReferencerForEKSCluster_AssignValidType_ReturnsExpected(t *testing.T) {

	r := &SecurityGroupIDReferencerForEKSCluster{}
	res := &EKSCluster{}
	var expectedErr error

	err := r.Assign(res, "mockValue")
	if diff := cmp.Diff(expectedErr, err, test.EquateErrors()); diff != "" {
		t.Errorf("Assign(...): -want error, +got error:\n%s", diff)
	}

	if diff := cmp.Diff(res.Spec.SecurityGroupIDs, []string{"mockValue"}); diff != "" {
		t.Errorf("Assign(...): -want value, +got value:\n%s", diff)
	}
}

func TestSecurityGroupIDReferencerForEKSWorkerNodes_AssignInvalidType_ReturnsErr(t *testing.T) {

	r := &SecurityGroupIDReferencerForEKSWorkerNodes{}
	expectedErr := errors.New(errResourceIsNotEKSCluster)

	err := r.Assign(&struct{ resource.CanReference }{}, "mockValue")
	if diff := cmp.Diff(expectedErr, err, test.EquateErrors()); diff != "" {
		t.Errorf("Assign(...): -want error, +got error:\n%s", diff)
	}
}

func TestSecurityGroupIDReferencerForEKSWorkerNodes_AssignValidType_ReturnsExpected(t *testing.T) {

	r := &SecurityGroupIDReferencerForEKSWorkerNodes{}
	res := &EKSCluster{}
	var expectedErr error

	err := r.Assign(res, "mockValue")
	if diff := cmp.Diff(expectedErr, err, test.EquateErrors()); diff != "" {
		t.Errorf("Assign(...): -want error, +got error:\n%s", diff)
	}

	if diff := cmp.Diff(res.Spec.WorkerNodes.ClusterControlPlaneSecurityGroup, "mockValue"); diff != "" {
		t.Errorf("Assign(...): -want value, +got value:\n%s", diff)
	}
}
