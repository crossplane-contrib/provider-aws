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

package vpc

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/onsi/gomega"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	v1alpha3 "github.com/crossplane/provider-aws/apis/network/v1alpha3"
	"github.com/crossplane/provider-aws/pkg/clients/ec2"
	"github.com/crossplane/provider-aws/pkg/clients/ec2/fake"
)

var (
	mockExternalClient external
	mockClient         fake.MockVPCClient

	// an arbitrary managed resource
	unexpectedItem resource.Managed
)

func TestMain(m *testing.M) {

	mockClient = fake.MockVPCClient{}
	mockExternalClient = external{
		client: &mockClient,
		kube: &test.MockClient{
			MockUpdate: test.NewMockUpdateFn(nil),
		},
	}

	os.Exit(m.Run())
}

func Test_Connect(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	mockManaged := &v1alpha3.VPC{}
	var clientErr error
	var configErr error

	conn := connector{
		client: nil,
		newClientFn: func(conf *aws.Config) (ec2.VPCClient, error) {
			return &mockClient, clientErr
		},
		awsConfigFn: func(context.Context, client.Reader, *corev1.ObjectReference) (*aws.Config, error) {
			return &aws.Config{}, configErr
		},
	}

	for _, tc := range []struct {
		description       string
		managedObj        resource.Managed
		configErr         error
		clientErr         error
		expectedClientNil bool
		expectedErrNil    bool
	}{
		{
			"valid input should return expected",
			mockManaged,
			nil,
			nil,
			false,
			true,
		},
		{
			"unexpected managed resource should return error",
			unexpectedItem,
			nil,
			nil,
			true,
			false,
		},
		{
			"if aws config provider fails, should return error",
			mockManaged, // an arbitrary managed resource which is not expected
			errors.New("some error"),
			nil,
			true,
			false,
		},
		{
			"if aws client provider fails, should return error",
			mockManaged, // an arbitrary managed resource which is not expected
			nil,
			errors.New("some error"),
			true,
			false,
		},
	} {
		clientErr = tc.clientErr
		configErr = tc.configErr

		res, err := conn.Connect(context.Background(), tc.managedObj)
		g.Expect(res == nil).To(gomega.Equal(tc.expectedClientNil), tc.description)
		g.Expect(err == nil).To(gomega.Equal(tc.expectedErrNil), tc.description)
	}
}

func Test_Observe(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	mockManaged := v1alpha3.VPC{}
	meta.SetExternalName(&mockManaged, "some arbitrary id")

	mockExternal := &awsec2.Vpc{
		VpcId: aws.String("some arbitrary Id"),
		State: awsec2.VpcStateAvailable,
	}
	var mockClientErr error
	var itemsList []awsec2.Vpc
	mockClient.MockDescribeVpcsRequest = func(input *awsec2.DescribeVpcsInput) awsec2.DescribeVpcsRequest {
		return awsec2.DescribeVpcsRequest{
			Request: &aws.Request{
				HTTPRequest: &http.Request{},
				Data: &awsec2.DescribeVpcsOutput{
					Vpcs: itemsList,
				},
				Error: mockClientErr,
			},
		}
	}

	for _, tc := range []struct {
		description           string
		managedObj            resource.Managed
		itemsReturned         []awsec2.Vpc
		clientErr             error
		expectedErrNil        bool
		expectedResourceExist bool
	}{
		{
			"valid input should return expected",
			mockManaged.DeepCopy(),
			[]awsec2.Vpc{*mockExternal},
			nil,
			true,
			true,
		},
		{
			"unexpected managed resource should return error",
			unexpectedItem,
			nil,
			nil,
			false,
			false,
		},
		{
			"if item's identifier is not yet set, returns expected",
			&v1alpha3.VPC{},
			nil,
			nil,
			true,
			false,
		},
		{
			"if external resource doesn't exist, it should return expected",
			mockManaged.DeepCopy(),
			nil,
			awserr.New(ec2.VPCIDNotFound, "", nil),
			true,
			false,
		},
		{
			"if external resource fails, it should return error",
			mockManaged.DeepCopy(),
			nil,
			errors.New("some error"),
			false,
			false,
		},
		{
			"if external resource returns a list with other than one item, it should return error",
			mockManaged.DeepCopy(),
			[]awsec2.Vpc{},
			nil,
			false,
			false,
		},
	} {
		mockClientErr = tc.clientErr
		itemsList = tc.itemsReturned

		result, err := mockExternalClient.Observe(context.Background(), tc.managedObj)

		g.Expect(err == nil).To(gomega.Equal(tc.expectedErrNil), tc.description)
		g.Expect(result.ResourceExists).To(gomega.Equal(tc.expectedResourceExist), tc.description)
		if tc.expectedResourceExist {
			mgd := tc.managedObj.(*v1alpha3.VPC)
			g.Expect(mgd.Status.Conditions[0].Type).To(gomega.Equal(corev1alpha1.TypeReady), tc.description)
			g.Expect(mgd.Status.Conditions[0].Status).To(gomega.Equal(corev1.ConditionTrue), tc.description)
			g.Expect(mgd.Status.Conditions[0].Reason).To(gomega.Equal(corev1alpha1.ReasonAvailable), tc.description)
			g.Expect(mgd.Status.VPCExternalStatus.VPCState).To(gomega.Equal(string(mockExternal.State)), tc.description)
		}
	}
}

func Test_Create(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	mockManaged := v1alpha3.VPC{
		Spec: v1alpha3.VPCSpec{
			VPCParameters: v1alpha3.VPCParameters{
				CIDRBlock: "arbitrary cidr block string",
			},
		},
	}
	mockExternal := &awsec2.Vpc{
		VpcId: aws.String("some arbitrary id"),
	}
	var mockClientErr error
	var numCreateCalled int
	mockClient.MockCreateVpcRequest = func(input *awsec2.CreateVpcInput) awsec2.CreateVpcRequest {
		numCreateCalled++
		g.Expect(aws.StringValue(input.CidrBlock)).To(gomega.Equal(mockManaged.Spec.CIDRBlock), "the passed parameters are not valid")
		return awsec2.CreateVpcRequest{
			Request: &aws.Request{
				HTTPRequest: &http.Request{},
				Data: &awsec2.CreateVpcOutput{
					Vpc: mockExternal,
				},
				Error: mockClientErr,
			},
		}
	}

	var mockModifyVpcErr error
	var numModifyCalled int
	mockClient.MockModifyVpcAttributeRequest = func(input *awsec2.ModifyVpcAttributeInput) awsec2.ModifyVpcAttributeRequest {
		numModifyCalled++
		g.Expect(input.VpcId).To(gomega.Equal(mockExternal.VpcId), "the passed parameters are not valid")
		return awsec2.ModifyVpcAttributeRequest{
			Request: &aws.Request{
				HTTPRequest: &http.Request{},
				Data:        &awsec2.ModifyVpcAttributeOutput{},
				Error:       mockModifyVpcErr,
			},
		}
	}

	for _, tc := range []struct {
		description         string
		managedObj          resource.Managed
		clientErr           error
		modifyVpcErr        error
		expectedErrNil      bool
		expectedCreateCalls int
		expectedModifyCalls int
	}{
		{
			"valid input should return expected",
			mockManaged.DeepCopy(),
			nil,
			nil,
			true,
			1,
			2,
		},
		{
			"unexpected managed resource should return error",
			unexpectedItem,
			nil,
			nil,
			false,
			0,
			0,
		},
		{
			"if creating resource fails, it should return error",
			mockManaged.DeepCopy(),
			errors.New("some error"),
			nil,
			false,
			1,
			0,
		},
		{
			"if external name is set, it should skip creating vpc",
			func() *v1alpha3.VPC {
				v := &v1alpha3.VPC{
					Spec: v1alpha3.VPCSpec{
						VPCParameters: v1alpha3.VPCParameters{
							CIDRBlock: "arbitrary cidr block string",
						},
					},
				}
				v.SetConditions(runtimev1alpha1.Creating())
				meta.SetExternalName(v, "some arbitrary id")
				return v
			}(),
			nil,
			nil,
			true,
			0,
			2,
		},
		{
			"if modifying attributes fails, it should return error",
			mockManaged.DeepCopy(),
			nil,
			errors.New("some error"),
			false,
			1,
			1,
		},
	} {
		numCreateCalled = 0
		numModifyCalled = 0
		mockClientErr = tc.clientErr
		mockModifyVpcErr = tc.modifyVpcErr

		_, err := mockExternalClient.Create(context.Background(), tc.managedObj)

		g.Expect(err == nil).To(gomega.Equal(tc.expectedErrNil), tc.description)
		if tc.expectedErrNil {
			mgd := tc.managedObj.(*v1alpha3.VPC)
			g.Expect(mgd.Status.Conditions[0].Type).To(gomega.Equal(corev1alpha1.TypeReady), tc.description)
			g.Expect(mgd.Status.Conditions[0].Status).To(gomega.Equal(corev1.ConditionFalse), tc.description)
			g.Expect(mgd.Status.Conditions[0].Reason).To(gomega.Equal(corev1alpha1.ReasonCreating), tc.description)
			g.Expect(mgd.Status.VPCExternalStatus.VPCState).To(gomega.Equal(string(mockExternal.State)), tc.description)
		}

		g.Expect(numCreateCalled).To(gomega.Equal(tc.expectedCreateCalls), tc.description)
		g.Expect(numModifyCalled).To(gomega.Equal(tc.expectedModifyCalls), tc.description)
	}
}

func Test_Delete(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	mockManaged := v1alpha3.VPC{
		Spec: v1alpha3.VPCSpec{
			VPCParameters: v1alpha3.VPCParameters{
				CIDRBlock: "arbitrary cidr block",
			},
		},
	}
	meta.SetExternalName(&mockManaged, "some arbitrary id")
	var mockClientErr error
	mockClient.MockDeleteVpcRequest = func(input *awsec2.DeleteVpcInput) awsec2.DeleteVpcRequest {
		g.Expect(aws.StringValue(input.VpcId)).To(gomega.Equal(meta.GetExternalName(&mockManaged)), "the passed parameters are not valid")
		return awsec2.DeleteVpcRequest{
			Request: &aws.Request{
				HTTPRequest: &http.Request{},
				Data:        &awsec2.DeleteVpcOutput{},
				Error:       mockClientErr,
			},
		}
	}

	for _, tc := range []struct {
		description    string
		managedObj     resource.Managed
		clientErr      error
		expectedErrNil bool
	}{
		{
			"valid input should return expected",
			mockManaged.DeepCopy(),
			nil,
			true,
		},
		{
			"unexpected managed resource should return error",
			unexpectedItem,
			nil,
			false,
		},
		{
			"if the resource doesn't exist deleting resource should not return an error",
			mockManaged.DeepCopy(),
			awserr.New(ec2.VPCIDNotFound, "", nil),
			true,
		},
		{
			"if deleting resource fails, it should return error",
			mockManaged.DeepCopy(),
			errors.New("some error"),
			false,
		},
	} {
		mockClientErr = tc.clientErr

		err := mockExternalClient.Delete(context.Background(), tc.managedObj)

		g.Expect(err == nil).To(gomega.Equal(tc.expectedErrNil), tc.description)
		if tc.expectedErrNil {
			mgd := tc.managedObj.(*v1alpha3.VPC)
			g.Expect(mgd.Status.Conditions[0].Type).To(gomega.Equal(corev1alpha1.TypeReady), tc.description)
			g.Expect(mgd.Status.Conditions[0].Status).To(gomega.Equal(corev1.ConditionFalse), tc.description)
			g.Expect(mgd.Status.Conditions[0].Reason).To(gomega.Equal(corev1alpha1.ReasonDeleting), tc.description)
		}
	}
}

const (
	providerName = "aws-creds"
)

var (
	errBoom = errors.New("boom")
)

type vpcModifier func(vpc *v1alpha3.VPC)

func withTags(tagMaps ...map[string]string) vpcModifier {
	var tagList []v1alpha3.Tag
	for _, tagMap := range tagMaps {
		for k, v := range tagMap {
			tagList = append(tagList, v1alpha3.Tag{Key: k, Value: v})
		}
	}
	return func(r *v1alpha3.VPC) { r.Spec.Tags = tagList }
}

func instance(m ...vpcModifier) *v1alpha3.VPC {
	cr := &v1alpha3.VPC{
		Spec: v1alpha3.VPCSpec{
			ResourceSpec: corev1alpha1.ResourceSpec{
				ProviderReference: &corev1.ObjectReference{Name: providerName},
			},
		},
	}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func TestInitialize(t *testing.T) {
	type args struct {
		cr   *v1alpha3.VPC
		kube client.Client
	}
	type want struct {
		cr  *v1alpha3.VPC
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				cr:   instance(withTags(map[string]string{"foo": "bar"})),
				kube: &test.MockClient{MockUpdate: test.NewMockUpdateFn(nil)},
			},
			want: want{
				cr: instance(withTags(resource.GetExternalTags(instance()), map[string]string{"foo": "bar"})),
			},
		},
		"UpdateFailed": {
			args: args{
				cr:   instance(),
				kube: &test.MockClient{MockUpdate: test.NewMockUpdateFn(errBoom)},
			},
			want: want{
				err: errors.Wrap(errBoom, errKubeUpdateFailed),
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
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, cmpopts.SortSlices(func(a, b v1alpha3.Tag) bool { return a.Key > b.Key })); err == nil && diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
