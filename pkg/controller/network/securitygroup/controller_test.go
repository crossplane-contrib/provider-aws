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

package securitygroup

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/onsi/gomega"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-aws/apis/network/v1alpha3"
	"github.com/crossplane/provider-aws/pkg/clients/ec2"
	"github.com/crossplane/provider-aws/pkg/clients/ec2/fake"
)

var (
	// an arbitrary managed resource
	unexpecedItem resource.Managed
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func Test_Connect(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	mockClient := fake.MockSecurityGroupClient{}
	mockManaged := &v1alpha3.SecurityGroup{}
	var clientErr error
	var configErr error

	conn := connector{
		kube: nil,
		newClientFn: func(conf *aws.Config) (ec2.SecurityGroupClient, error) {
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
			unexpecedItem,
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
			"if aws sg provider fails, should return error",
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
	mockClient := fake.MockSecurityGroupClient{}
	mockExternalClient := external{sg: &mockClient}
	mockManaged := &v1alpha3.SecurityGroup{}
	meta.SetExternalName(mockManaged, "some arbitrary id")

	mockExternal := &awsec2.SecurityGroup{
		GroupId: aws.String("some arbitrary Id"),
		Tags: []awsec2.Tag{
			{Key: aws.String("key1"), Value: aws.String("value1")},
			{Key: aws.String("key2"), Value: aws.String("value2")},
		},
	}
	var mockClientErr error
	var itemsList []awsec2.SecurityGroup
	mockClient.MockDescribeSecurityGroupsRequest = func(input *awsec2.DescribeSecurityGroupsInput) awsec2.DescribeSecurityGroupsRequest {
		return awsec2.DescribeSecurityGroupsRequest{
			Request: &aws.Request{
				HTTPRequest: &http.Request{},
				Data: &awsec2.DescribeSecurityGroupsOutput{
					SecurityGroups: itemsList,
				},
				Error: mockClientErr,
			},
		}
	}

	for _, tc := range []struct {
		description           string
		managedObj            resource.Managed
		itemsReturned         []awsec2.SecurityGroup
		clientErr             error
		expectedErrNil        bool
		expectedResourceExist bool
	}{
		{
			"valid input should return expected",
			mockManaged.DeepCopy(),
			[]awsec2.SecurityGroup{*mockExternal},
			nil,
			true,
			true,
		},
		{
			"unexpected managed resource should return error",
			unexpecedItem,
			nil,
			nil,
			false,
			false,
		},
		{
			"if item's identifier is not yet set, returns expected",
			&v1alpha3.SecurityGroup{},
			nil,
			nil,
			true,
			false,
		},
		{
			"if external resource doesn't exist, it should return expected",
			mockManaged.DeepCopy(),
			nil,
			awserr.New(ec2.InvalidGroupNotFound, "", nil),
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
			[]awsec2.SecurityGroup{},
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
			mgd := tc.managedObj.(*v1alpha3.SecurityGroup)
			g.Expect(mgd.Status.Conditions[0].Type).To(gomega.Equal(corev1alpha1.TypeReady), tc.description)
			g.Expect(mgd.Status.Conditions[0].Status).To(gomega.Equal(corev1.ConditionTrue), tc.description)
			g.Expect(mgd.Status.Conditions[0].Reason).To(gomega.Equal(corev1alpha1.ReasonAvailable), tc.description)
			g.Expect(len(mgd.Status.SecurityGroupExternalStatus.Tags)).To(gomega.Equal(len(mockExternal.Tags)), tc.description)
		}
	}
}

func Test_Create(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	mockClient := fake.MockSecurityGroupClient{}
	mockExternalClient := external{
		sg:   &mockClient,
		kube: test.NewMockClient()}
	mockManaged := v1alpha3.SecurityGroup{
		Spec: v1alpha3.SecurityGroupSpec{
			SecurityGroupParameters: v1alpha3.SecurityGroupParameters{
				VPCID:       aws.String("arbitrary vpcId"),
				Description: "arbitrary description",
				GroupName:   "arbitrary group name",
				Ingress: []v1alpha3.IPPermission{
					{
						FromPort:   aws.Int64(7766),
						ToPort:     aws.Int64(9988),
						IPProtocol: "an arbitrary protocol",
						IPRanges: []v1alpha3.IPRange{
							{
								CIDRIP:      "0.0.0.0/0",
								Description: aws.String("an arbitrary cidr block"),
							},
						},
					}, {}, {},
				},
				Egress: []v1alpha3.IPPermission{
					{
						FromPort:   aws.Int64(1122),
						ToPort:     aws.Int64(3344),
						IPProtocol: "an arbitrary protocol",
						IPRanges: []v1alpha3.IPRange{
							{
								CIDRIP:      "0.0.0.0/0",
								Description: aws.String("an arbitrary cidr block"),
							},
						},
					}, {},
				},
			},
		},
		Status: v1alpha3.SecurityGroupStatus{
			SecurityGroupExternalStatus: v1alpha3.SecurityGroupExternalStatus{
				SecurityGroupID: "some arbitrary id",
			},
		},
	}
	mockExternal := &awsec2.SecurityGroup{
		GroupId: aws.String("some arbitrary id"),
	}
	var mockClientErr error
	mockClient.MockCreateSecurityGroupRequest = func(input *awsec2.CreateSecurityGroupInput) awsec2.CreateSecurityGroupRequest {
		g.Expect(input.VpcId).To(gomega.Equal(mockManaged.Spec.VPCID), "the passed parameters are not valid")
		return awsec2.CreateSecurityGroupRequest{
			Request: &aws.Request{
				HTTPRequest: &http.Request{},
				Data: &awsec2.CreateSecurityGroupOutput{
					GroupId: mockExternal.GroupId,
				},
				Error: mockClientErr,
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
			unexpecedItem,
			nil,
			false,
		},
		{
			"if creating resource fails, it should return error",
			mockManaged.DeepCopy(),
			errors.New("some error"),
			false,
		},
	} {
		mockClientErr = tc.clientErr

		_, err := mockExternalClient.Create(context.Background(), tc.managedObj)

		g.Expect(err == nil).To(gomega.Equal(tc.expectedErrNil), tc.description)
		if tc.expectedErrNil {
			mgd := tc.managedObj.(*v1alpha3.SecurityGroup)
			g.Expect(mgd.Status.Conditions[0].Type).To(gomega.Equal(corev1alpha1.TypeReady), tc.description)
			g.Expect(mgd.Status.Conditions[0].Status).To(gomega.Equal(corev1.ConditionFalse), tc.description)
			g.Expect(mgd.Status.Conditions[0].Reason).To(gomega.Equal(corev1alpha1.ReasonCreating), tc.description)
			g.Expect(mgd.Status.SecurityGroupID).To(gomega.Equal(aws.StringValue(mockExternal.GroupId)), tc.description)
		}
	}
}

func Test_Update(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	mockClient := fake.MockSecurityGroupClient{}
	mockExternalClient := external{sg: &mockClient}
	mockManaged := &v1alpha3.SecurityGroup{
		Spec: v1alpha3.SecurityGroupSpec{
			SecurityGroupParameters: v1alpha3.SecurityGroupParameters{
				VPCID:       aws.String("arbitrary vpcId"),
				Description: "arbitrary description",
				GroupName:   "arbitrary group name",
				Ingress: []v1alpha3.IPPermission{
					{
						FromPort:   aws.Int64(7766),
						ToPort:     aws.Int64(9988),
						IPProtocol: "an arbitrary protocol",
						IPRanges: []v1alpha3.IPRange{
							{
								CIDRIP:      "0.0.0.0/0",
								Description: aws.String("an arbitrary cidr block"),
							},
						},
					}, {}, {},
				},
				Egress: []v1alpha3.IPPermission{
					{
						FromPort:   aws.Int64(1122),
						ToPort:     aws.Int64(3344),
						IPProtocol: "an arbitrary protocol",
						IPRanges: []v1alpha3.IPRange{
							{
								CIDRIP:      "0.0.0.0/0",
								Description: aws.String("an arbitrary cidr block"),
							},
						},
					}, {},
				},
			},
		},
		Status: v1alpha3.SecurityGroupStatus{
			SecurityGroupExternalStatus: v1alpha3.SecurityGroupExternalStatus{
				SecurityGroupID: "some arbitrary id",
			},
		},
	}
	meta.SetExternalName(mockManaged, mockManaged.Status.SecurityGroupID)

	var mockClientIngressErr error
	var ingressCalled bool
	mockClient.MockAuthorizeSecurityGroupIngressRequest = func(input *awsec2.AuthorizeSecurityGroupIngressInput) awsec2.AuthorizeSecurityGroupIngressRequest {
		ingressCalled = true
		g.Expect(aws.StringValue(input.GroupId)).To(gomega.Equal(meta.GetExternalName(mockManaged)), "the passed parameters are not valid")
		g.Expect(len(input.IpPermissions)).To(gomega.Equal(len(mockManaged.Spec.Ingress)), "the passed parameters are not valid")
		return awsec2.AuthorizeSecurityGroupIngressRequest{
			Request: &aws.Request{
				HTTPRequest: &http.Request{},
				Data:        &awsec2.AuthorizeSecurityGroupIngressOutput{},
				Error:       mockClientIngressErr,
			},
		}
	}

	var mockClientEgressErr error
	var egressCalled bool
	mockClient.MockAuthorizeSecurityGroupEgressRequest = func(input *awsec2.AuthorizeSecurityGroupEgressInput) awsec2.AuthorizeSecurityGroupEgressRequest {
		egressCalled = true
		g.Expect(aws.StringValue(input.GroupId)).To(gomega.Equal(mockManaged.Status.SecurityGroupID), "the passed parameters are not valid")
		g.Expect(len(input.IpPermissions)).To(gomega.Equal(len(mockManaged.Spec.Egress)), "the passed parameters are not valid")
		return awsec2.AuthorizeSecurityGroupEgressRequest{
			Request: &aws.Request{
				HTTPRequest: &http.Request{},
				Data:        &awsec2.AuthorizeSecurityGroupEgressOutput{},
				Error:       mockClientEgressErr,
			},
		}
	}

	for _, tc := range []struct {
		description         string
		managedObj          resource.Managed
		clientIngressErr    error
		clientEgressErr     error
		expectedIngressCall bool
		expectedEgressCall  bool
		expectedErrNil      bool
	}{
		{
			"if creating ingress rules fails, it should return error",
			mockManaged.DeepCopy(),
			errors.New("some error"),
			nil,
			true,
			false,
			false,
		},
		{
			"if there are no ingress rules fails, it should return expected",
			(&v1alpha3.SecurityGroup{
				ObjectMeta: v1.ObjectMeta{
					Annotations: map[string]string{
						meta.AnnotationKeyExternalName: mockManaged.Status.SecurityGroupID,
					},
				},
				Spec: v1alpha3.SecurityGroupSpec{
					SecurityGroupParameters: v1alpha3.SecurityGroupParameters{
						VPCID:       aws.String("arbitrary vpcId"),
						Description: "arbitrary description",
						GroupName:   "arbitrary group name",
						Egress:      mockManaged.Spec.Egress,
					},
				},
				Status: mockManaged.Status,
			}).DeepCopy(),
			nil,
			nil,
			false,
			true,
			true,
		},
		{
			"if creating egress rules fails, it should return error",
			mockManaged.DeepCopy(),
			nil,
			errors.New("some error"),
			true,
			true,
			false,
		},
		{
			"if there are no egress rules fails, it should return expected",
			(&v1alpha3.SecurityGroup{
				ObjectMeta: v1.ObjectMeta{
					Annotations: map[string]string{
						meta.AnnotationKeyExternalName: mockManaged.Status.SecurityGroupID,
					},
				},
				Spec: v1alpha3.SecurityGroupSpec{
					SecurityGroupParameters: v1alpha3.SecurityGroupParameters{
						VPCID:       aws.String("arbitrary vpcId"),
						Description: "arbitrary description",
						GroupName:   "arbitrary group name",
						Ingress:     mockManaged.Spec.Ingress,
					},
				},
				Status: mockManaged.Status,
			}).DeepCopy(),
			nil,
			nil,
			true,
			false,
			true,
		},
	} {
		ingressCalled = false
		egressCalled = false
		mockClientIngressErr = tc.clientIngressErr
		mockClientEgressErr = tc.clientEgressErr

		_, err := mockExternalClient.Update(context.Background(), tc.managedObj)

		g.Expect(err == nil).To(gomega.Equal(tc.expectedErrNil), tc.description)

		g.Expect(ingressCalled).To(gomega.Equal(tc.expectedIngressCall), tc.description)
		g.Expect(egressCalled).To(gomega.Equal(tc.expectedEgressCall), tc.description)
	}
}

func Test_Delete(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	mockClient := fake.MockSecurityGroupClient{}
	mockExternalClient := external{sg: &mockClient}
	mockManaged := &v1alpha3.SecurityGroup{}
	meta.SetExternalName(mockManaged, "some arbitrary id")

	var mockClientErr error
	mockClient.MockDeleteSecurityGroupRequest = func(input *awsec2.DeleteSecurityGroupInput) awsec2.DeleteSecurityGroupRequest {
		g.Expect(aws.StringValue(input.GroupId)).To(gomega.Equal(meta.GetExternalName(mockManaged)), "the passed parameters are not valid")
		return awsec2.DeleteSecurityGroupRequest{
			Request: &aws.Request{
				HTTPRequest: &http.Request{},
				Data:        &awsec2.DeleteSecurityGroupOutput{},
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
			unexpecedItem,
			nil,
			false,
		},
		{
			"if the resource doesn't exist deleting resource should not return an error",
			mockManaged.DeepCopy(),
			awserr.New(ec2.InvalidGroupNotFound, "", nil),
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
			mgd := tc.managedObj.(*v1alpha3.SecurityGroup)
			g.Expect(mgd.Status.Conditions[0].Type).To(gomega.Equal(corev1alpha1.TypeReady), tc.description)
			g.Expect(mgd.Status.Conditions[0].Status).To(gomega.Equal(corev1.ConditionFalse), tc.description)
			g.Expect(mgd.Status.Conditions[0].Reason).To(gomega.Equal(corev1alpha1.ReasonDeleting), tc.description)
		}
	}
}
