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

package ec2

import (
	"errors"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/smithy-go"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/onsi/gomega"

	"github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
)

func TestMain(m *testing.M) {

	os.Exit(m.Run())
}

func Test_IsInternetGatewayNotFoundErr(t *testing.T) {

	testCases := []struct {
		name string
		got  error
		want bool
	}{
		{
			"nil error is not",
			nil,
			false,
		},
		{
			"other error is not",
			errors.New("some error"),
			false,
		},
		{
			"InternetGatewayNotFoundErr is",
			&smithy.GenericAPIError{Code: InternetGatewayIDNotFound},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			if diff := cmp.Diff(tc.want, IsInternetGatewayNotFoundErr(tc.got), test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func Test_IsRouteTableNotFoundErr(t *testing.T) {

	testCases := []struct {
		name string
		got  error
		want bool
	}{
		{
			"nil error is not",
			nil,
			false,
		},
		{
			"other error is not",
			errors.New("some error"),
			false,
		},
		{
			"RouteTableNotFoundErr is",
			&smithy.GenericAPIError{Code: RouteTableIDNotFound},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			if diff := cmp.Diff(tc.want, IsRouteTableNotFoundErr(tc.got), test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func Test_IsRouteNotFoundErr(t *testing.T) {

	testCases := []struct {
		name string
		got  error
		want bool
	}{
		{
			"nil error is not",
			nil,
			false,
		},
		{
			"other error is not",
			errors.New("some error"),
			false,
		},
		{
			"RouteNotFoundErr is",
			&smithy.GenericAPIError{Code: RouteNotFound},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			if diff := cmp.Diff(tc.want, IsRouteNotFoundErr(tc.got), test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func Test_IsAssociationIDNotFoundErr(t *testing.T) {

	testCases := []struct {
		name string
		got  error
		want bool
	}{
		{
			"nil error is not",
			nil,
			false,
		},
		{
			"other error is not",
			errors.New("some error"),
			false,
		},
		{
			"AssociationIDNotFoundErr is",
			&smithy.GenericAPIError{Code: AssociationIDNotFound},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			if diff := cmp.Diff(tc.want, IsAssociationIDNotFoundErr(tc.got), test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func Test_IsSecurityGroupNotFoundErr(t *testing.T) {

	testCases := []struct {
		name string
		got  error
		want bool
	}{
		{
			"nil error is not",
			nil,
			false,
		},
		{
			"other error is not",
			errors.New("some error"),
			false,
		},
		{
			"SecurityGroupNotFoundErr is",
			&smithy.GenericAPIError{Code: InvalidGroupNotFound},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			if diff := cmp.Diff(tc.want, IsSecurityGroupNotFoundErr(tc.got), test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func Test_IsSubnetNotFoundErr(t *testing.T) {

	testCases := []struct {
		name string
		got  error
		want bool
	}{
		{
			"nil error is not",
			nil,
			false,
		},
		{
			"other error is not",
			errors.New("some error"),
			false,
		},
		{
			"SubnetNotFoundErr is",
			&smithy.GenericAPIError{Code: SubnetIDNotFound},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			if diff := cmp.Diff(tc.want, IsSubnetNotFoundErr(tc.got), test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func Test_IsVPCNotFoundErr(t *testing.T) {

	testCases := []struct {
		name string
		got  error
		want bool
	}{
		{
			"nil error is not",
			nil,
			false,
		},
		{
			"other error is not",
			errors.New("some error"),
			false,
		},
		{
			"VPCNotFoundErr is",
			&smithy.GenericAPIError{Code: VPCIDNotFound},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			if diff := cmp.Diff(tc.want, IsVPCNotFoundErr(tc.got), test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func Test_SecurityGroup_BuildEC2Permissions(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	res := GenerateEC2Permissions([]v1beta1.IPPermission{
		{
			FromPort: aws.Int32(8080),
			ToPort:   aws.Int32(8080),
			IPRanges: []v1beta1.IPRange{
				{CIDRIP: "arbitranry cidrip"},
			},
		},
	})

	g.Expect(res).ToNot(gomega.BeNil())
}
