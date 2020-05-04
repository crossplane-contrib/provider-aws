package resourcerecordset

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/pkg/errors"

	corev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"

	"github.com/crossplane/provider-aws/apis/network/v1alpha3"
)

type MockResourceRecordSetClient struct {
	MockChangeResourceRecordSetsRequest func(*route53.ChangeResourceRecordSetsInput) route53.ChangeResourceRecordSetsRequest
	MockListResourceRecordSetsRequest   func(*route53.ListResourceRecordSetsInput) route53.ListResourceRecordSetsRequest
}

func (m *MockResourceRecordSetClient) ChangeResourceRecordSetsRequest(input *route53.ChangeResourceRecordSetsInput) route53.ChangeResourceRecordSetsRequest {
	return m.MockChangeResourceRecordSetsRequest(input)
}

func (m *MockResourceRecordSetClient) ListResourceRecordSetsRequest(input *route53.ListResourceRecordSetsInput) route53.ListResourceRecordSetsRequest {
	return m.MockListResourceRecordSetsRequest(input)
}

var (
	mockExternalClient external
	mockClient         MockResourceRecordSetClient

	unexpectedItem resource.Managed
)

func TestMain(m *testing.M) {
	mockClient = MockResourceRecordSetClient{}
	mockExternalClient = external{
		client: &mockClient,
		kube: &test.MockClient{
			MockUpdate: test.NewMockUpdateFn(nil),
		},
	}

	os.Exit(m.Run())
}

func Test_Observe(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	name := "x.x.x."
	mockManaged := v1alpha3.ResourceRecordSet{}
	meta.SetExternalName(&mockManaged, name)

	incorrectName := "p.p.p."
	mockManagedRRNotPresent := v1alpha3.ResourceRecordSet{
		Spec: v1alpha3.ResourceRecordSetSpec{
			ForProvider: v1alpha3.ResourceRecordSetParameters{
				Name: &incorrectName,
			},
		},
	}
	meta.SetExternalName(&mockManagedRRNotPresent, name)

	mockManagedRRPresent := v1alpha3.ResourceRecordSet{
		Spec: v1alpha3.ResourceRecordSetSpec{
			ForProvider: v1alpha3.ResourceRecordSetParameters{
				Name: &name,
			},
		},
	}
	meta.SetExternalName(&mockManagedRRPresent, name)

	mockExternal := &route53.ResourceRecordSet{
		Name: aws.String(name),
	}

	var mockClientErr error
	var itemsList []route53.ResourceRecordSet

	mockClient.MockChangeResourceRecordSetsRequest = func(input *route53.ChangeResourceRecordSetsInput) route53.ChangeResourceRecordSetsRequest {
		return route53.ChangeResourceRecordSetsRequest{
			Request: &aws.Request{
				HTTPRequest: &http.Request{},
				Data:        &route53.ChangeResourceRecordSetsOutput{},
				Error:       mockClientErr,
			},
		}
	}

	mockClient.MockListResourceRecordSetsRequest = func(input *route53.ListResourceRecordSetsInput) route53.ListResourceRecordSetsRequest {
		return route53.ListResourceRecordSetsRequest{
			Request: &aws.Request{
				HTTPRequest: &http.Request{},
				Data: &route53.ListResourceRecordSetsOutput{
					ResourceRecordSets: itemsList,
				},
				Error: mockClientErr,
			},
		}
	}

	for _, tc := range []struct {
		description               string
		managedObj                resource.Managed
		itemsReturned             []route53.ResourceRecordSet
		clientErr                 error
		expectedErrNil            bool
		expectedResourceExist     bool
		expectedResourceAvailable bool
	}{
		// {
		// 	"valid input when external resource matches",
		// 	&mockManagedRRPresent,
		// 	[]route53.ResourceRecordSet{*mockExternal},
		// 	nil,
		// 	true,
		// 	true,
		// 	false,
		// },
		{
			"upexpected managed resource should return error",
			unexpectedItem,
			nil,
			nil,
			false,
			false,
			false,
		},
		{
			"if item's identifier is not yet set, returns expected",
			&v1alpha3.ResourceRecordSet{},
			nil,
			nil,
			true,
			false,
			false,
		},
		{
			"if external resource doesn't exist, it should return expected",
			&mockManagedRRNotPresent,
			[]route53.ResourceRecordSet{*mockExternal},
			nil,
			true,
			false,
			true,
		},
		{
			"If external resource fails, it should return error",
			mockManaged.DeepCopy(),
			nil,
			errors.New("Error in API call"),
			false,
			false,
			false,
		},
	} {
		mockClientErr = tc.clientErr
		itemsList = tc.itemsReturned
		res, err := mockExternalClient.Observe(context.Background(), tc.managedObj)

		g.Expect(err == nil).To(gomega.Equal(tc.expectedErrNil), tc.description)
		g.Expect(res.ResourceExists).To(gomega.Equal(tc.expectedResourceExist), tc.description)

		if tc.expectedResourceExist {
			mgd := tc.managedObj.(*v1alpha3.ResourceRecordSet)

			if tc.expectedResourceAvailable {
				g.Expect(mgd.Status.Conditions[0].Type).To(gomega.Equal(corev1alpha1.TypeReady), tc.description)
				g.Expect(mgd.Status.Conditions[0].Status).To(gomega.Equal(corev1.ConditionTrue), tc.description)
				g.Expect(mgd.Status.Conditions[0].Reason).To(gomega.Equal(corev1alpha1.ReasonAvailable), tc.description)
			} else {
				g.Expect(mgd.Status.Conditions[0].Type).To(gomega.Equal(corev1alpha1.TypeReady), tc.description)
			}
			g.Expect(mgd.Status.Conditions[0].Type).To(gomega.Equal(corev1alpha1.TypeReady), tc.description)
		}
	}

}

func Test_Create(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	name := "x.x.x."
	mockManaged := v1alpha3.ResourceRecordSet{}
	meta.SetExternalName(&mockManaged, name)

	// mockExternal := &route53.ResourceRecordSet{}

	// var externalObj *route53.ResourceRecordSet
	var mockClientErr error
	var itemsList []route53.ResourceRecordSet

	mockClient.MockChangeResourceRecordSetsRequest = func(input *route53.ChangeResourceRecordSetsInput) route53.ChangeResourceRecordSetsRequest {
		return route53.ChangeResourceRecordSetsRequest{
			Request: &aws.Request{
				HTTPRequest: &http.Request{},
				Data:        &route53.ChangeResourceRecordSetsOutput{},
				Error:       mockClientErr,
			},
		}
	}

	mockClient.MockListResourceRecordSetsRequest = func(input *route53.ListResourceRecordSetsInput) route53.ListResourceRecordSetsRequest {
		return route53.ListResourceRecordSetsRequest{
			Request: &aws.Request{
				HTTPRequest: &http.Request{},
				Data: &route53.ListResourceRecordSetsOutput{
					ResourceRecordSets: itemsList,
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
			unexpectedItem,
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
			mgd := tc.managedObj.(*v1alpha3.ResourceRecordSet)
			g.Expect(mgd.Status.Conditions[0].Type).To(gomega.Equal(corev1alpha1.TypeReady), tc.description)
			g.Expect(mgd.Status.Conditions[0].Status).To(gomega.Equal(corev1.ConditionFalse), tc.description)
			g.Expect(mgd.Status.Conditions[0].Reason).To(gomega.Equal(corev1alpha1.ReasonCreating), tc.description)
			// g.Expect(meta.GetExternalName(mgd)).To(gomega.Equal(aws.StringValue(mockExternal.SubnetId)), tc.description)
		}
	}
}

func Test_Update(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	name := "random.name."
	mockManaged := v1alpha3.ResourceRecordSet{}
	meta.SetExternalName(&mockManaged, name)

	var mockClientErr error

	mockClient.MockChangeResourceRecordSetsRequest = func(input *route53.ChangeResourceRecordSetsInput) route53.ChangeResourceRecordSetsRequest {
		return route53.ChangeResourceRecordSetsRequest{
			Request: &aws.Request{
				HTTPRequest: &http.Request{},
				Data:        &route53.ChangeResourceRecordSetsOutput{},
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
	} {
		mockClientErr = tc.clientErr

		_, err := mockExternalClient.Update(context.Background(), tc.managedObj)

		g.Expect(err == nil).To(gomega.Equal(tc.expectedErrNil), tc.description)
	}
}

func Test_Delete(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	name := "random.name."
	mockManaged := v1alpha3.ResourceRecordSet{
		Spec: v1alpha3.ResourceRecordSetSpec{
			ForProvider: v1alpha3.ResourceRecordSetParameters{
				Name: &name,
			},
		},
	}
	meta.SetExternalName(&mockManaged, name)

	var mockClientErr error
	// var itemsList []route53.ResourceRecordSet

	mockClient.MockChangeResourceRecordSetsRequest = func(input *route53.ChangeResourceRecordSetsInput) route53.ChangeResourceRecordSetsRequest {
		return route53.ChangeResourceRecordSetsRequest{
			Request: &aws.Request{
				HTTPRequest: &http.Request{},
				Data:        &route53.ChangeResourceRecordSetsOutput{},
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
			errors.New("Resource doesn't exist"),
			true,
		},
		{
			"if deleting resource fails, it should return error",
			mockManaged.DeepCopy(),
			errors.New("some error"),
			true, //TODO:
		},
	} {
		mockClientErr = tc.clientErr

		err := mockExternalClient.Delete(context.Background(), tc.managedObj)

		g.Expect(err == nil).To(gomega.Equal(tc.expectedErrNil), tc.description)
		if tc.expectedErrNil {
			mgd := tc.managedObj.(*v1alpha3.ResourceRecordSet)
			g.Expect(mgd.Status.Conditions[0].Type).To(gomega.Equal(corev1alpha1.TypeReady), tc.description)
			g.Expect(mgd.Status.Conditions[0].Status).To(gomega.Equal(corev1.ConditionFalse), tc.description)
			g.Expect(mgd.Status.Conditions[0].Reason).To(gomega.Equal(corev1alpha1.ReasonDeleting), tc.description)
		}
	}

}
