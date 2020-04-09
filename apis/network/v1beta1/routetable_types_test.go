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
	corev1 "k8s.io/api/core/v1"

	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
)

var _ resource.AttributeReferencer = (*VPCIDReferencerForRouteTable)(nil)
var _ resource.AttributeReferencer = (*SubnetIDReferencerForRouteTable)(nil)
var _ resource.AttributeReferencer = (*InternetGatewayIDReferencerForRouteTable)(nil)

var rtMockValue = "mockValue"

func TestVPCIDReferencerForRouteTable_AssignInvalidType_ReturnsErr(t *testing.T) {

	r := &VPCIDReferencerForRouteTable{}
	expectedErr := errors.New(errResourceIsNotRouteTable)

	err := r.Assign(nil, rtMockValue)
	if diff := cmp.Diff(expectedErr, err, test.EquateErrors()); diff != "" {
		t.Errorf("Assign(...): -want error, +got error:\n%s", diff)
	}
}

func TestVPCIDReferencerForRouteTable_AssignValidType_ReturnsExpected(t *testing.T) {

	r := &VPCIDReferencerForRouteTable{}
	res := &RouteTable{}
	var expectedErr error

	err := r.Assign(res, rtMockValue)
	if diff := cmp.Diff(expectedErr, err, test.EquateErrors()); diff != "" {
		t.Errorf("Assign(...): -want error, +got error:\n%s", diff)
	}

	if diff := cmp.Diff(res.Spec.ForProvider.VPCID, &rtMockValue); diff != "" {
		t.Errorf("Assign(...): -want value, +got value:\n%s", diff)
	}
}

func TestSubnetIDReferencerForRouteTable_AssignInvalidType_ReturnsErr(t *testing.T) {

	r := &SubnetIDReferencerForRouteTable{}
	expectedErr := errors.New(errResourceIsNotRouteTable)

	err := r.Assign(nil, rtMockValue)
	if diff := cmp.Diff(expectedErr, err, test.EquateErrors()); diff != "" {
		t.Errorf("Assign(...): -want error, +got error:\n%s", diff)
	}
}

func TestSubnetIDReferencerForRouteTable_AssociationWithSameNameNotExist_ReturnsErr(t *testing.T) {

	r := &SubnetIDReferencerForRouteTable{
		SubnetIDReferencer: SubnetIDReferencer{
			LocalObjectReference: corev1.LocalObjectReference{Name: "mockObjectName1"},
		},
	}

	expectedErr := errors.New(errAssociationNotFound)

	err := r.Assign(&RouteTable{}, rtMockValue)
	if diff := cmp.Diff(expectedErr, err, test.EquateErrors()); diff != "" {
		t.Errorf("Assign(...): -want error, +got error:\n%s", diff)
	}
}

func TestSubnetIDReferencerForRouteTable_AssignValidType_ReturnsExpected(t *testing.T) {

	r1 := &SubnetIDReferencerForRouteTable{
		SubnetIDReferencer: SubnetIDReferencer{
			LocalObjectReference: corev1.LocalObjectReference{Name: "mockObjectName1"},
		},
	}

	r2 := &SubnetIDReferencerForRouteTable{
		SubnetIDReferencer: SubnetIDReferencer{
			LocalObjectReference: corev1.LocalObjectReference{Name: "mockObjectName2"},
		},
	}

	res := &RouteTable{
		Spec: RouteTableSpec{
			ForProvider: RouteTableParameters{
				Associations: []Association{{SubnetIDRef: r2}, {SubnetIDRef: r1}},
			},
		},
	}

	var expectedErr error

	err := r1.Assign(res, "mockSubnetID")
	if diff := cmp.Diff(expectedErr, err, test.EquateErrors()); diff != "" {
		t.Errorf("Assign(...): -want error, +got error:\n%s", diff)
	}

	if diff := cmp.Diff(res.Spec.ForProvider.Associations[1].SubnetID, "mockSubnetID"); diff != "" {
		t.Errorf("Assign(...): -want value, +got value:\n%s", diff)
	}
}

func TestInternetGatewayIDReferencerForRouteTable_AssignInvalidType_ReturnsErr(t *testing.T) {

	r := &InternetGatewayIDReferencerForRouteTable{}
	expectedErr := errors.New(errResourceIsNotRouteTable)

	err := r.Assign(nil, rtMockValue)
	if diff := cmp.Diff(expectedErr, err, test.EquateErrors()); diff != "" {
		t.Errorf("Assign(...): -want error, +got error:\n%s", diff)
	}
}

func TestInternetGatewayIDReferencerForRouteTable_RouteWithSameNameNotExist_ReturnsErr(t *testing.T) {

	r := &InternetGatewayIDReferencerForRouteTable{
		InternetGatewayIDReferencer: InternetGatewayIDReferencer{
			LocalObjectReference: corev1.LocalObjectReference{Name: "mockObjectName1"},
		},
	}

	expectedErr := errors.New(errRouteNotFound)

	err := r.Assign(&RouteTable{}, rtMockValue)
	if diff := cmp.Diff(expectedErr, err, test.EquateErrors()); diff != "" {
		t.Errorf("Assign(...): -want error, +got error:\n%s", diff)
	}
}

func TestInternetGatewayIDReferencerForRouteTable_AssignValidType_ReturnsExpected(t *testing.T) {

	r1 := &InternetGatewayIDReferencerForRouteTable{
		InternetGatewayIDReferencer: InternetGatewayIDReferencer{
			LocalObjectReference: corev1.LocalObjectReference{Name: "mockObjectName1"},
		},
	}

	r2 := &InternetGatewayIDReferencerForRouteTable{
		InternetGatewayIDReferencer: InternetGatewayIDReferencer{
			LocalObjectReference: corev1.LocalObjectReference{Name: "mockObjectName2"},
		},
	}

	res := &RouteTable{
		Spec: RouteTableSpec{
			ForProvider: RouteTableParameters{
				Routes: []Route{{GatewayIDRef: r2}, {GatewayIDRef: r1}},
			},
		},
	}

	var expectedErr error

	err := r1.Assign(res, "mockGatewayID")
	if diff := cmp.Diff(expectedErr, err, test.EquateErrors()); diff != "" {
		t.Errorf("Assign(...): -want error, +got error:\n%s", diff)
	}

	if diff := cmp.Diff(res.Spec.ForProvider.Routes[1].GatewayID, "mockGatewayID"); diff != "" {
		t.Errorf("Assign(...): -want value, +got value:\n%s", diff)
	}
}
