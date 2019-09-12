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

package dbsubnetgroup

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	awsrds "github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/onsi/gomega"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"

	v1alpha2 "github.com/crossplaneio/stack-aws/aws/apis/storage/v1alpha2"
	"github.com/crossplaneio/stack-aws/pkg/clients/aws/rds"
	"github.com/crossplaneio/stack-aws/pkg/clients/aws/rds/fake"
)

var (
	mockExternalClient external
	mockClient         fake.MockDBSubnetGroupClient

	// an arbitrary managed resource
	unexpecedItem resource.Managed
)

func TestMain(m *testing.M) {

	mockClient = fake.MockDBSubnetGroupClient{}
	mockExternalClient = external{&mockClient}

	os.Exit(m.Run())
}

func Test_Connect(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	mockManaged := &v1alpha2.DBSubnetGroup{}
	var clientErr error
	var configErr error

	conn := connector{
		client: nil,
		newClientFn: func(conf *aws.Config) (rds.DBSubnetGroupClient, error) {
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

	mockManaged := v1alpha2.DBSubnetGroup{
		Spec: v1alpha2.DBSubnetGroupSpec{
			DBSubnetGroupParameters: v1alpha2.DBSubnetGroupParameters{
				DBSubnetGroupDescription: "arbitrary description",
				DBSubnetGroupName:        "arbitrary group name",
				SubnetIDs:                []string{"subnetid1", "subnetid2"},
			},
		},
	}

	mockExternal := &awsrds.DBSubnetGroup{
		VpcId:             aws.String("arbitrary vpcId"),
		DBSubnetGroupArn:  aws.String("arbitrary group arn"),
		SubnetGroupStatus: aws.String("arbitrary group status"),
	}
	var mockClientErr error
	var itemsList []awsrds.DBSubnetGroup
	mockClient.MockDescribeDBSubnetGroupsRequest = func(input *awsrds.DescribeDBSubnetGroupsInput) awsrds.DescribeDBSubnetGroupsRequest {
		return awsrds.DescribeDBSubnetGroupsRequest{
			Request: &aws.Request{
				HTTPRequest: &http.Request{},
				Data: &awsrds.DescribeDBSubnetGroupsOutput{
					DBSubnetGroups: itemsList,
				},
				Error: mockClientErr,
			},
		}
	}

	for _, tc := range []struct {
		description           string
		managedObj            resource.Managed
		itemsReturned         []awsrds.DBSubnetGroup
		clientErr             error
		expectedErrNil        bool
		expectedResourceExist bool
	}{
		{
			"valid input should return expected",
			mockManaged.DeepCopy(),
			[]awsrds.DBSubnetGroup{*mockExternal},
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
			"if external resource doesn't exist, it should return expected",
			mockManaged.DeepCopy(),
			nil,
			awserr.New(awsrds.ErrCodeDBSubnetGroupNotFoundFault, "", nil),
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
			[]awsrds.DBSubnetGroup{},
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
			mgd := tc.managedObj.(*v1alpha2.DBSubnetGroup)
			g.Expect(mgd.Status.Conditions[0].Type).To(gomega.Equal(corev1alpha1.TypeReady), tc.description)
			g.Expect(mgd.Status.Conditions[0].Status).To(gomega.Equal(corev1.ConditionTrue), tc.description)
			g.Expect(mgd.Status.Conditions[0].Reason).To(gomega.Equal(corev1alpha1.ReasonAvailable), tc.description)
			g.Expect(mgd.Status.DBSubnetGroupExternalStatus.SubnetGroupStatus).To(gomega.Equal(aws.StringValue(mockExternal.SubnetGroupStatus)), tc.description)
		}
	}
}

func Test_Create(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	mockManaged := v1alpha2.DBSubnetGroup{
		Spec: v1alpha2.DBSubnetGroupSpec{
			DBSubnetGroupParameters: v1alpha2.DBSubnetGroupParameters{
				DBSubnetGroupDescription: "arbitrary description",
				DBSubnetGroupName:        "arbitrary group name",
				SubnetIDs:                []string{"subnetid1", "subnetid2"},
				Tags: []v1alpha2.Tag{
					{"tagKey1", "tagValue1"}, {"tagKey2", "tagValue2"},
				},
			},
		},
	}
	mockExternal := &awsrds.DBSubnetGroup{
		VpcId:             aws.String("arbitrary vpcId"),
		DBSubnetGroupArn:  aws.String("arbitrary group arn"),
		SubnetGroupStatus: aws.String("arbitrary group status"),
	}
	var mockClientErr error
	mockClient.MockCreateDBSubnetGroupRequest = func(input *awsrds.CreateDBSubnetGroupInput) awsrds.CreateDBSubnetGroupRequest {
		g.Expect(aws.StringValue(input.DBSubnetGroupDescription)).To(gomega.Equal(mockManaged.Spec.DBSubnetGroupDescription), "the passed parameters are not valid")
		g.Expect(aws.StringValue(input.DBSubnetGroupName)).To(gomega.Equal(mockManaged.Spec.DBSubnetGroupName), "the passed parameters are not valid")
		g.Expect(input.SubnetIds).To(gomega.Equal(mockManaged.Spec.SubnetIDs), "the passed parameters are not valid")
		g.Expect(len(input.Tags)).To(gomega.Equal(len(mockManaged.Spec.Tags)), "the passed parameters are not valid")
		return awsrds.CreateDBSubnetGroupRequest{
			Request: &aws.Request{
				HTTPRequest: &http.Request{},
				Data: &awsrds.CreateDBSubnetGroupOutput{
					DBSubnetGroup: mockExternal,
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
			mgd := tc.managedObj.(*v1alpha2.DBSubnetGroup)
			g.Expect(mgd.Status.Conditions[0].Type).To(gomega.Equal(corev1alpha1.TypeReady), tc.description)
			g.Expect(mgd.Status.Conditions[0].Status).To(gomega.Equal(corev1.ConditionFalse), tc.description)
			g.Expect(mgd.Status.Conditions[0].Reason).To(gomega.Equal(corev1alpha1.ReasonCreating), tc.description)
			g.Expect(mgd.Status.SubnetGroupStatus).To(gomega.Equal(aws.StringValue(mockExternal.SubnetGroupStatus)))
		}
	}
}

func Test_Update(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	mockManaged := v1alpha2.DBSubnetGroup{}

	_, err := mockExternalClient.Update(context.Background(), &mockManaged)

	g.Expect(err).To(gomega.BeNil())
}

func Test_Delete(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	mockManaged := v1alpha2.DBSubnetGroup{
		Spec: v1alpha2.DBSubnetGroupSpec{
			DBSubnetGroupParameters: v1alpha2.DBSubnetGroupParameters{
				DBSubnetGroupDescription: "arbitrary description",
				DBSubnetGroupName:        "arbitrary group name",
				SubnetIDs:                []string{"subnetid1", "subnetid2"},
			},
		},
	}
	var mockClientErr error
	mockClient.MockDeleteDBSubnetGroupRequest = func(input *awsrds.DeleteDBSubnetGroupInput) awsrds.DeleteDBSubnetGroupRequest {
		g.Expect(aws.StringValue(input.DBSubnetGroupName)).To(gomega.Equal(mockManaged.Spec.DBSubnetGroupName), "the passed parameters are not valid")
		return awsrds.DeleteDBSubnetGroupRequest{
			Request: &aws.Request{
				HTTPRequest: &http.Request{},
				Data:        &awsrds.DeleteDBSubnetGroupOutput{},
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
			awserr.New(awsrds.ErrCodeDBSubnetGroupNotFoundFault, "", nil),
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
			mgd := tc.managedObj.(*v1alpha2.DBSubnetGroup)
			g.Expect(mgd.Status.Conditions[0].Type).To(gomega.Equal(corev1alpha1.TypeReady), tc.description)
			g.Expect(mgd.Status.Conditions[0].Status).To(gomega.Equal(corev1.ConditionFalse), tc.description)
			g.Expect(mgd.Status.Conditions[0].Reason).To(gomega.Equal(corev1alpha1.ReasonDeleting), tc.description)
		}
	}
}
