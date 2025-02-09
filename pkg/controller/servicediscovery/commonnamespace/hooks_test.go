/*
Copyright 2021 The Crossplane Authors.

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

package commonnamespace

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/aws/aws-sdk-go/service/servicediscovery/servicediscoveryiface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/servicediscovery/v1alpha1"
	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/servicediscovery/v1alpha1"
	svcclient "github.com/crossplane-contrib/provider-aws/pkg/clients/servicediscovery"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/servicediscovery/fake"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

const (
	validOpID        string = "123"
	validNSID        string = "ns-id"
	validDescription string = "valid description"
	validArn         string = "arn:string"
	validTagKey1     string = "key1"
	validTagValue1   string = "value1"
	validTagKey2     string = "key2"
	validTagValue2   string = "value2"
	validTagKey3     string = "key3"
	validTagValue3   string = "value3"
)

type args struct {
	client svcclient.Client
	kube   client.Client
	cr     *svcapitypes.HTTPNamespace
}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *svcapitypes.HTTPNamespace
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"NewNoOpID": {
			args: args{
				client: &fake.MockServicediscoveryClient{},
				kube:   nil,
				cr: &svcapitypes.HTTPNamespace{
					Spec: svcapitypes.HTTPNamespaceSpec{
						ForProvider: svcapitypes.HTTPNamespaceParameters{
							Region: "eu-central-1",
							Name:   aws.String("test"),
						},
					},
				},
			},
			want: want{
				cr: &svcapitypes.HTTPNamespace{
					Spec: svcapitypes.HTTPNamespaceSpec{
						ForProvider: svcapitypes.HTTPNamespaceParameters{
							Region: "eu-central-1",
							Name:   aws.String("test"),
						},
					},
					Status: svcapitypes.HTTPNamespaceStatus{},
				},
				result: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"NewOperationSubmitted": {
			args: args{
				client: &fake.MockServicediscoveryClient{
					MockGetOperation: func(input *svcsdk.GetOperationInput) (*svcsdk.GetOperationOutput, error) {
						if pointer.StringValue(input.OperationId) != validOpID {
							return &svcsdk.GetOperationOutput{}, nil
						}
						return &svcsdk.GetOperationOutput{
							Operation: &svcsdk.Operation{
								Status:  aws.String("SUBMITTED"),
								Targets: map[string]*string{"NAMESPACE": aws.String(validNSID)},
							},
						}, nil
					},
				},
				kube: nil,
				cr: &svcapitypes.HTTPNamespace{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							"servicediscovery.aws.crossplane.io/operation-id": validOpID,
						},
					},
					Spec: svcapitypes.HTTPNamespaceSpec{
						ForProvider: svcapitypes.HTTPNamespaceParameters{
							Region: "eu-central-1",
							Name:   aws.String("test"),
						},
					},
				},
			},
			want: want{
				cr: &svcapitypes.HTTPNamespace{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							"servicediscovery.aws.crossplane.io/operation-id": validOpID,
						},
					},
					Spec: svcapitypes.HTTPNamespaceSpec{
						ForProvider: svcapitypes.HTTPNamespaceParameters{
							Region: "eu-central-1",
							Name:   aws.String("test"),
						},
					},
				},
				result: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"NewOperationPending": {
			args: args{
				client: &fake.MockServicediscoveryClient{
					MockGetOperation: func(input *svcsdk.GetOperationInput) (*svcsdk.GetOperationOutput, error) {
						if pointer.StringValue(input.OperationId) != validOpID {
							return &svcsdk.GetOperationOutput{}, nil
						}
						return &svcsdk.GetOperationOutput{
							Operation: &svcsdk.Operation{
								Status:  aws.String("PENDING"),
								Targets: map[string]*string{"NAMESPACE": aws.String(validNSID)},
							},
						}, nil
					},
				},
				kube: nil,
				cr: &svcapitypes.HTTPNamespace{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							"servicediscovery.aws.crossplane.io/operation-id": validOpID,
						},
					},
					Spec: svcapitypes.HTTPNamespaceSpec{
						ForProvider: svcapitypes.HTTPNamespaceParameters{
							Region: "eu-central-1",
							Name:   aws.String("test"),
						},
					},
				},
			},
			want: want{
				cr: &svcapitypes.HTTPNamespace{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							"servicediscovery.aws.crossplane.io/operation-id": validOpID,
						},
					},
					Spec: svcapitypes.HTTPNamespaceSpec{
						ForProvider: svcapitypes.HTTPNamespaceParameters{
							Region: "eu-central-1",
							Name:   aws.String("test"),
						},
					},
				},
				result: managed.ExternalObservation{},
			},
		},
		"NewOperationFailed": {
			args: args{
				client: &fake.MockServicediscoveryClient{
					MockGetOperation: func(input *svcsdk.GetOperationInput) (*svcsdk.GetOperationOutput, error) {
						if pointer.StringValue(input.OperationId) != validOpID {
							return &svcsdk.GetOperationOutput{}, nil
						}
						return &svcsdk.GetOperationOutput{
							Operation: &svcsdk.Operation{
								Status:  aws.String("FAIL"),
								Targets: map[string]*string{"NAMESPACE": aws.String(validNSID)},
							},
						}, nil
					},
				},
				kube: nil,
				cr: &svcapitypes.HTTPNamespace{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							"servicediscovery.aws.crossplane.io/operation-id": validOpID,
						},
					},
					Spec: svcapitypes.HTTPNamespaceSpec{
						ForProvider: svcapitypes.HTTPNamespaceParameters{
							Region: "eu-central-1",
							Name:   aws.String("test"),
						},
					},
				},
			},
			want: want{
				cr: &svcapitypes.HTTPNamespace{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							"servicediscovery.aws.crossplane.io/operation-id": validOpID,
						},
					},
					Spec: svcapitypes.HTTPNamespaceSpec{
						ForProvider: svcapitypes.HTTPNamespaceParameters{
							Region: "eu-central-1",
							Name:   aws.String("test"),
						},
					},
					Status: svcapitypes.HTTPNamespaceStatus{
						ResourceStatus: xpv1.ResourceStatus{
							ConditionedStatus: xpv1.ConditionedStatus{
								Conditions: []xpv1.Condition{xpv1.Unavailable()},
							},
						},
					},
				},
				result: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"DeletingOperationFail": {
			args: args{
				client: &fake.MockServicediscoveryClient{
					MockGetOperation: func(input *svcsdk.GetOperationInput) (*svcsdk.GetOperationOutput, error) {
						if pointer.StringValue(input.OperationId) != validOpID {
							return &svcsdk.GetOperationOutput{}, nil
						}
						return &svcsdk.GetOperationOutput{
							Operation: &svcsdk.Operation{
								Status:  aws.String("FAIL"),
								Targets: map[string]*string{"NAMESPACE": aws.String(validNSID)},
							},
						}, nil
					},
				},
				kube: nil,
				cr: &svcapitypes.HTTPNamespace{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							"servicediscovery.aws.crossplane.io/operation-id": validOpID,
						},
					},
					Spec: svcapitypes.HTTPNamespaceSpec{
						ForProvider: svcapitypes.HTTPNamespaceParameters{
							Region: "eu-central-1",
							Name:   aws.String("test"),
						},
					},
					Status: svcapitypes.HTTPNamespaceStatus{
						ResourceStatus: xpv1.ResourceStatus{
							ConditionedStatus: xpv1.ConditionedStatus{
								Conditions: []xpv1.Condition{xpv1.Deleting()},
							},
						},
					},
				},
			},
			want: want{
				cr: &svcapitypes.HTTPNamespace{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							"servicediscovery.aws.crossplane.io/operation-id": validOpID,
						},
					},
					Spec: svcapitypes.HTTPNamespaceSpec{
						ForProvider: svcapitypes.HTTPNamespaceParameters{
							Region: "eu-central-1",
							Name:   aws.String("test"),
						},
					},
					Status: svcapitypes.HTTPNamespaceStatus{
						ResourceStatus: xpv1.ResourceStatus{
							ConditionedStatus: xpv1.ConditionedStatus{
								Conditions: []xpv1.Condition{xpv1.Unavailable()},
							},
						},
					},
				},
				result: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"NewExternalNameNotSet": {
			args: args{
				client: &fake.MockServicediscoveryClient{
					MockListTagsForResource: func(*svcsdk.ListTagsForResourceInput) (*svcsdk.ListTagsForResourceOutput, error) {
						return &svcsdk.ListTagsForResourceOutput{
							Tags: nil,
						}, nil
					},
					MockGetOperation: func(input *svcsdk.GetOperationInput) (*svcsdk.GetOperationOutput, error) {
						if pointer.StringValue(input.OperationId) != validOpID {
							return &svcsdk.GetOperationOutput{}, nil
						}
						return &svcsdk.GetOperationOutput{
							Operation: &svcsdk.Operation{
								Status:  aws.String("SUCCESS"),
								Targets: map[string]*string{"NAMESPACE": aws.String(validNSID)},
							},
						}, nil
					},
					MockGetNamespace: func(input *svcsdk.GetNamespaceInput) (*svcsdk.GetNamespaceOutput, error) {
						if pointer.StringValue(input.Id) != validNSID {
							return &svcsdk.GetNamespaceOutput{}, nil
						}
						return &svcsdk.GetNamespaceOutput{
							Namespace: &svcsdk.Namespace{
								Arn:         aws.String(validArn),
								Name:        aws.String(validNSID),
								Description: aws.String(validDescription),
							},
						}, nil
					},
				},
				kube: &test.MockClient{
					MockGet:    test.NewMockGetFn(nil),
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				cr: &svcapitypes.HTTPNamespace{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							"servicediscovery.aws.crossplane.io/operation-id": validOpID,
						},
					},
					Spec: svcapitypes.HTTPNamespaceSpec{
						ForProvider: svcapitypes.HTTPNamespaceParameters{
							Region:      "eu-central-1",
							Name:        aws.String("test"),
							Description: aws.String(validDescription),
						},
					},
				},
			},
			want: want{
				cr: &svcapitypes.HTTPNamespace{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							"crossplane.io/external-name":                     validNSID,
							"servicediscovery.aws.crossplane.io/operation-id": validOpID,
						},
					},
					Spec: svcapitypes.HTTPNamespaceSpec{
						ForProvider: svcapitypes.HTTPNamespaceParameters{
							Region:      "eu-central-1",
							Name:        aws.String("test"),
							Description: aws.String(validDescription),
						},
					},
					Status: svcapitypes.HTTPNamespaceStatus{
						ResourceStatus: xpv1.ResourceStatus{
							ConditionedStatus: xpv1.ConditionedStatus{
								Conditions: []xpv1.Condition{xpv1.Available()},
							},
						},
					},
				},
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceLateInitialized: false,
					ResourceUpToDate:        true,
				},
			},
		},
		"NewWithExternalName": {
			args: args{
				client: &fake.MockServicediscoveryClient{
					MockGetOperation: func(input *svcsdk.GetOperationInput) (*svcsdk.GetOperationOutput, error) {
						if pointer.StringValue(input.OperationId) != validOpID {
							return &svcsdk.GetOperationOutput{}, nil
						}
						return &svcsdk.GetOperationOutput{
							Operation: &svcsdk.Operation{
								Status:  aws.String("SUCCESS"),
								Targets: map[string]*string{"NAMESPACE": aws.String(validNSID)},
							},
						}, nil
					},
					MockGetNamespace: func(input *svcsdk.GetNamespaceInput) (*svcsdk.GetNamespaceOutput, error) {
						if pointer.StringValue(input.Id) != validNSID {
							return &svcsdk.GetNamespaceOutput{}, nil
						}
						return &svcsdk.GetNamespaceOutput{
							Namespace: &svcsdk.Namespace{
								Arn:         aws.String(validArn),
								Name:        aws.String(validNSID),
								Description: aws.String(validDescription),
							},
						}, nil
					},
				},
				kube: nil,
				cr: &svcapitypes.HTTPNamespace{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							"crossplane.io/external-name":                     validNSID,
							"servicediscovery.aws.crossplane.io/operation-id": validOpID,
						},
					},
					Spec: svcapitypes.HTTPNamespaceSpec{
						ForProvider: svcapitypes.HTTPNamespaceParameters{
							Region:      "eu-central-1",
							Name:        aws.String("test"),
							Description: aws.String(validDescription),
						},
					},
				},
			},
			want: want{
				cr: &svcapitypes.HTTPNamespace{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							"crossplane.io/external-name":                     validNSID,
							"servicediscovery.aws.crossplane.io/operation-id": validOpID,
						},
					},
					Spec: svcapitypes.HTTPNamespaceSpec{
						ForProvider: svcapitypes.HTTPNamespaceParameters{
							Region:      "eu-central-1",
							Name:        aws.String("test"),
							Description: aws.String(validDescription),
						},
					},
					Status: svcapitypes.HTTPNamespaceStatus{
						ResourceStatus: xpv1.ResourceStatus{
							ConditionedStatus: xpv1.ConditionedStatus{
								Conditions: []xpv1.Condition{xpv1.Available()},
							},
						},
					},
				},
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceLateInitialized: false,
					ResourceUpToDate:        true,
				},
			},
		},
		"DescriptionChange": {
			args: args{
				client: &fake.MockServicediscoveryClient{
					MockGetOperation: func(input *svcsdk.GetOperationInput) (*svcsdk.GetOperationOutput, error) {
						if pointer.StringValue(input.OperationId) != validOpID {
							return &svcsdk.GetOperationOutput{}, nil
						}
						return &svcsdk.GetOperationOutput{
							Operation: &svcsdk.Operation{
								Status:  aws.String("SUCCESS"),
								Targets: map[string]*string{"NAMESPACE": aws.String(validNSID)},
							},
						}, nil
					},
					MockGetNamespace: func(input *svcsdk.GetNamespaceInput) (*svcsdk.GetNamespaceOutput, error) {
						if pointer.StringValue(input.Id) != validNSID {
							return &svcsdk.GetNamespaceOutput{}, nil
						}
						return &svcsdk.GetNamespaceOutput{
							Namespace: &svcsdk.Namespace{
								Arn:         aws.String(validArn),
								Name:        aws.String(validNSID),
								Description: aws.String(validDescription),
							},
						}, nil
					},
				},
				kube: nil,
				cr: &svcapitypes.HTTPNamespace{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							"crossplane.io/external-name":                     validNSID,
							"servicediscovery.aws.crossplane.io/operation-id": validOpID,
						},
					},
					Spec: svcapitypes.HTTPNamespaceSpec{
						ForProvider: svcapitypes.HTTPNamespaceParameters{
							Region:      "eu-central-1",
							Name:        aws.String("test"),
							Description: aws.String("change Description"),
						},
					},
				},
			},
			want: want{
				cr: &svcapitypes.HTTPNamespace{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							"crossplane.io/external-name":                     validNSID,
							"servicediscovery.aws.crossplane.io/operation-id": validOpID,
						},
					},
					Spec: svcapitypes.HTTPNamespaceSpec{
						ForProvider: svcapitypes.HTTPNamespaceParameters{
							Region:      "eu-central-1",
							Name:        aws.String("test"),
							Description: aws.String("change Description"),
						},
					},
					Status: svcapitypes.HTTPNamespaceStatus{
						ResourceStatus: xpv1.ResourceStatus{
							ConditionedStatus: xpv1.ConditionedStatus{
								Conditions: []xpv1.Condition{xpv1.Available()},
							},
						},
					},
				},
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceLateInitialized: false,
					ResourceUpToDate:        false,
				},
			},
		},
		"NoChangeTag": {
			args: args{
				client: &fake.MockServicediscoveryClient{
					MockGetOperation: func(input *svcsdk.GetOperationInput) (*svcsdk.GetOperationOutput, error) {
						if pointer.StringValue(input.OperationId) != validOpID {
							return &svcsdk.GetOperationOutput{}, nil
						}
						return &svcsdk.GetOperationOutput{
							Operation: &svcsdk.Operation{
								Status:  aws.String("SUCCESS"),
								Targets: map[string]*string{"NAMESPACE": aws.String(validNSID)},
							},
						}, nil
					},
					MockGetNamespace: func(input *svcsdk.GetNamespaceInput) (*svcsdk.GetNamespaceOutput, error) {
						if pointer.StringValue(input.Id) != validNSID {
							return &svcsdk.GetNamespaceOutput{}, nil
						}
						return &svcsdk.GetNamespaceOutput{
							Namespace: &svcsdk.Namespace{
								Arn:         aws.String(validArn),
								Name:        aws.String(validNSID),
								Description: aws.String(validDescription),
							},
						}, nil
					},
					MockListTagsForResource: func(input *svcsdk.ListTagsForResourceInput) (*svcsdk.ListTagsForResourceOutput, error) {
						return &svcsdk.ListTagsForResourceOutput{
							Tags: []*svcsdk.Tag{
								{
									Key:   aws.String(validTagKey1),
									Value: aws.String(validTagValue1),
								},
								{
									Key:   aws.String(validTagKey2),
									Value: aws.String(validTagValue2),
								},
								{
									Key:   aws.String(validTagKey3),
									Value: aws.String(validTagValue3),
								},
							},
						}, nil
					},
				},
				kube: nil,
				cr: &svcapitypes.HTTPNamespace{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							"crossplane.io/external-name":                     validNSID,
							"servicediscovery.aws.crossplane.io/operation-id": validOpID,
						},
					},
					Spec: svcapitypes.HTTPNamespaceSpec{
						ForProvider: svcapitypes.HTTPNamespaceParameters{
							Region:      "eu-central-1",
							Name:        aws.String("test"),
							Description: aws.String(validDescription),
							Tags: []*svcapitypes.Tag{
								{
									Key:   aws.String(validTagKey1),
									Value: aws.String(validTagValue1),
								},
								{
									Key:   aws.String(validTagKey2),
									Value: aws.String(validTagValue2),
								},
								{
									Key:   aws.String(validTagKey3),
									Value: aws.String(validTagValue3),
								},
							},
						},
					},
				},
			},
			want: want{
				cr: &svcapitypes.HTTPNamespace{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							"crossplane.io/external-name":                     validNSID,
							"servicediscovery.aws.crossplane.io/operation-id": validOpID,
						},
					},
					Spec: svcapitypes.HTTPNamespaceSpec{
						ForProvider: svcapitypes.HTTPNamespaceParameters{
							Region:      "eu-central-1",
							Name:        aws.String("test"),
							Description: aws.String(validDescription),
							Tags: []*svcapitypes.Tag{
								{
									Key:   aws.String(validTagKey1),
									Value: aws.String(validTagValue1),
								},
								{
									Key:   aws.String(validTagKey2),
									Value: aws.String(validTagValue2),
								},
								{
									Key:   aws.String(validTagKey3),
									Value: aws.String(validTagValue3),
								},
							},
						},
					},
					Status: svcapitypes.HTTPNamespaceStatus{
						ResourceStatus: xpv1.ResourceStatus{
							ConditionedStatus: xpv1.ConditionedStatus{
								Conditions: []xpv1.Condition{xpv1.Available()},
							},
						},
					},
				},
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceLateInitialized: false,
					ResourceUpToDate:        true,
				},
			},
		},
		"NewTag": {
			args: args{
				client: &fake.MockServicediscoveryClient{
					MockGetOperation: func(input *svcsdk.GetOperationInput) (*svcsdk.GetOperationOutput, error) {
						if pointer.StringValue(input.OperationId) != validOpID {
							return &svcsdk.GetOperationOutput{}, nil
						}
						return &svcsdk.GetOperationOutput{
							Operation: &svcsdk.Operation{
								Status:  aws.String("SUCCESS"),
								Targets: map[string]*string{"NAMESPACE": aws.String(validNSID)},
							},
						}, nil
					},
					MockGetNamespace: func(input *svcsdk.GetNamespaceInput) (*svcsdk.GetNamespaceOutput, error) {
						if pointer.StringValue(input.Id) != validNSID {
							return &svcsdk.GetNamespaceOutput{}, nil
						}
						return &svcsdk.GetNamespaceOutput{
							Namespace: &svcsdk.Namespace{
								Arn:         aws.String(validArn),
								Name:        aws.String(validNSID),
								Description: aws.String(validDescription),
							},
						}, nil
					},
					MockListTagsForResource: func(input *svcsdk.ListTagsForResourceInput) (*svcsdk.ListTagsForResourceOutput, error) {
						return &svcsdk.ListTagsForResourceOutput{
							Tags: []*svcsdk.Tag{
								{
									Key:   aws.String(validTagKey1),
									Value: aws.String(validTagValue1),
								},
								{
									Key:   aws.String(validTagKey2),
									Value: aws.String(validTagValue2),
								},
								{
									Key:   aws.String(validTagKey3),
									Value: aws.String(validTagValue3),
								},
							},
						}, nil
					},
				},
				kube: nil,
				cr: &svcapitypes.HTTPNamespace{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							"crossplane.io/external-name":                     validNSID,
							"servicediscovery.aws.crossplane.io/operation-id": validOpID,
						},
					},
					Spec: svcapitypes.HTTPNamespaceSpec{
						ForProvider: svcapitypes.HTTPNamespaceParameters{
							Region:      "eu-central-1",
							Name:        aws.String("test"),
							Description: aws.String(validDescription),
							Tags: []*svcapitypes.Tag{
								{
									Key:   aws.String(validTagKey1),
									Value: aws.String(validTagValue1),
								},
							},
						},
					},
				},
			},
			want: want{
				cr: &svcapitypes.HTTPNamespace{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							"crossplane.io/external-name":                     validNSID,
							"servicediscovery.aws.crossplane.io/operation-id": validOpID,
						},
					},
					Spec: svcapitypes.HTTPNamespaceSpec{
						ForProvider: svcapitypes.HTTPNamespaceParameters{
							Region:      "eu-central-1",
							Name:        aws.String("test"),
							Description: aws.String(validDescription),
							Tags: []*svcapitypes.Tag{
								{
									Key:   aws.String(validTagKey1),
									Value: aws.String(validTagValue1),
								},
							},
						},
					},
					Status: svcapitypes.HTTPNamespaceStatus{
						ResourceStatus: xpv1.ResourceStatus{
							ConditionedStatus: xpv1.ConditionedStatus{
								Conditions: []xpv1.Condition{xpv1.Available()},
							},
						},
					},
				},
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceLateInitialized: false,
					ResourceUpToDate:        false,
				},
			},
		},
		"DeleteTag": {
			args: args{
				client: &fake.MockServicediscoveryClient{
					MockGetOperation: func(input *svcsdk.GetOperationInput) (*svcsdk.GetOperationOutput, error) {
						if pointer.StringValue(input.OperationId) != validOpID {
							return &svcsdk.GetOperationOutput{}, nil
						}
						return &svcsdk.GetOperationOutput{
							Operation: &svcsdk.Operation{
								Status:  aws.String("SUCCESS"),
								Targets: map[string]*string{"NAMESPACE": aws.String(validNSID)},
							},
						}, nil
					},
					MockGetNamespace: func(input *svcsdk.GetNamespaceInput) (*svcsdk.GetNamespaceOutput, error) {
						if pointer.StringValue(input.Id) != validNSID {
							return &svcsdk.GetNamespaceOutput{}, nil
						}
						return &svcsdk.GetNamespaceOutput{
							Namespace: &svcsdk.Namespace{
								Arn:         aws.String(validArn),
								Name:        aws.String(validNSID),
								Description: aws.String(validDescription),
							},
						}, nil
					},
					MockListTagsForResource: func(input *svcsdk.ListTagsForResourceInput) (*svcsdk.ListTagsForResourceOutput, error) {
						return &svcsdk.ListTagsForResourceOutput{
							Tags: []*svcsdk.Tag{
								{
									Key:   aws.String(validTagKey1),
									Value: aws.String(validTagValue1),
								},
							},
						}, nil
					},
				},
				kube: nil,
				cr: &svcapitypes.HTTPNamespace{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							"crossplane.io/external-name":                     validNSID,
							"servicediscovery.aws.crossplane.io/operation-id": validOpID,
						},
					},
					Spec: svcapitypes.HTTPNamespaceSpec{
						ForProvider: svcapitypes.HTTPNamespaceParameters{
							Region:      "eu-central-1",
							Name:        aws.String("test"),
							Description: aws.String(validDescription),
							Tags: []*svcapitypes.Tag{
								{
									Key:   aws.String(validTagKey1),
									Value: aws.String(validTagValue1),
								},
								{
									Key:   aws.String(validTagKey2),
									Value: aws.String(validTagValue2),
								},
								{
									Key:   aws.String(validTagKey3),
									Value: aws.String(validTagValue3),
								},
							},
						},
					},
				},
			},
			want: want{
				cr: &svcapitypes.HTTPNamespace{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							"crossplane.io/external-name":                     validNSID,
							"servicediscovery.aws.crossplane.io/operation-id": validOpID,
						},
					},
					Spec: svcapitypes.HTTPNamespaceSpec{
						ForProvider: svcapitypes.HTTPNamespaceParameters{
							Region:      "eu-central-1",
							Name:        aws.String("test"),
							Description: aws.String(validDescription),
							Tags: []*svcapitypes.Tag{
								{
									Key:   aws.String(validTagKey1),
									Value: aws.String(validTagValue1),
								},
								{
									Key:   aws.String(validTagKey2),
									Value: aws.String(validTagValue2),
								},
								{
									Key:   aws.String(validTagKey3),
									Value: aws.String(validTagValue3),
								},
							},
						},
					},
					Status: svcapitypes.HTTPNamespaceStatus{
						ResourceStatus: xpv1.ResourceStatus{
							ConditionedStatus: xpv1.ConditionedStatus{
								Conditions: []xpv1.Condition{xpv1.Available()},
							},
						},
					},
				},
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceLateInitialized: false,
					ResourceUpToDate:        false,
				},
			},
		},
		"ChangeTag": {
			args: args{
				client: &fake.MockServicediscoveryClient{
					MockGetOperation: func(input *svcsdk.GetOperationInput) (*svcsdk.GetOperationOutput, error) {
						if pointer.StringValue(input.OperationId) != validOpID {
							return &svcsdk.GetOperationOutput{}, nil
						}
						return &svcsdk.GetOperationOutput{
							Operation: &svcsdk.Operation{
								Status:  aws.String("SUCCESS"),
								Targets: map[string]*string{"NAMESPACE": aws.String(validNSID)},
							},
						}, nil
					},
					MockGetNamespace: func(input *svcsdk.GetNamespaceInput) (*svcsdk.GetNamespaceOutput, error) {
						if pointer.StringValue(input.Id) != validNSID {
							return &svcsdk.GetNamespaceOutput{}, nil
						}
						return &svcsdk.GetNamespaceOutput{
							Namespace: &svcsdk.Namespace{
								Arn:         aws.String(validArn),
								Name:        aws.String(validNSID),
								Description: aws.String(validDescription),
							},
						}, nil
					},
					MockListTagsForResource: func(input *svcsdk.ListTagsForResourceInput) (*svcsdk.ListTagsForResourceOutput, error) {
						return &svcsdk.ListTagsForResourceOutput{
							Tags: []*svcsdk.Tag{
								{
									Key:   aws.String(validTagKey1),
									Value: aws.String(validTagValue1),
								},
								{
									Key:   aws.String(validTagKey2),
									Value: aws.String(validTagValue2),
								},
								{
									Key:   aws.String(validTagKey3),
									Value: aws.String("changeValue"),
								},
							},
						}, nil
					},
				},
				kube: nil,
				cr: &svcapitypes.HTTPNamespace{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							"crossplane.io/external-name":                     validNSID,
							"servicediscovery.aws.crossplane.io/operation-id": validOpID,
						},
					},
					Spec: svcapitypes.HTTPNamespaceSpec{
						ForProvider: svcapitypes.HTTPNamespaceParameters{
							Region:      "eu-central-1",
							Name:        aws.String("test"),
							Description: aws.String(validDescription),
							Tags: []*svcapitypes.Tag{
								{
									Key:   aws.String(validTagKey1),
									Value: aws.String(validTagValue1),
								},
								{
									Key:   aws.String(validTagKey2),
									Value: aws.String(validTagValue2),
								},
								{
									Key:   aws.String(validTagKey3),
									Value: aws.String(validTagValue3),
								},
							},
						},
					},
				},
			},
			want: want{
				cr: &svcapitypes.HTTPNamespace{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							"crossplane.io/external-name":                     validNSID,
							"servicediscovery.aws.crossplane.io/operation-id": validOpID,
						},
					},
					Spec: svcapitypes.HTTPNamespaceSpec{
						ForProvider: svcapitypes.HTTPNamespaceParameters{
							Region:      "eu-central-1",
							Name:        aws.String("test"),
							Description: aws.String(validDescription),
							Tags: []*svcapitypes.Tag{
								{
									Key:   aws.String(validTagKey1),
									Value: aws.String(validTagValue1),
								},
								{
									Key:   aws.String(validTagKey2),
									Value: aws.String(validTagValue2),
								},
								{
									Key:   aws.String(validTagKey3),
									Value: aws.String(validTagValue3),
								},
							},
						},
					},
					Status: svcapitypes.HTTPNamespaceStatus{
						ResourceStatus: xpv1.ResourceStatus{
							ConditionedStatus: xpv1.ConditionedStatus{
								Conditions: []xpv1.Condition{xpv1.Available()},
							},
						},
					},
				},
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceLateInitialized: false,
					ResourceUpToDate:        false,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := NewHooks[namespace](tc.kube, tc.client)

			o, err := e.Observe(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions(),
				cmpopts.IgnoreFields(v1.Condition{}, "LastTransitionTime"),
			); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type want struct {
		cr  *svcapitypes.HTTPNamespace
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"NewHTTPNamespace": {
			args: args{
				client: &fake.MockServicediscoveryClient{
					MockDeleteNamespace: func(input *svcsdk.DeleteNamespaceInput) (*svcsdk.DeleteNamespaceOutput, error) {
						return &svcsdk.DeleteNamespaceOutput{
							OperationId: aws.String(validOpID),
						}, nil
					},
				},
				kube: nil,
				cr: &svcapitypes.HTTPNamespace{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							"servicediscovery.aws.crossplane.io/operation-id": validOpID,
						},
					},
					Spec: svcapitypes.HTTPNamespaceSpec{
						ForProvider: svcapitypes.HTTPNamespaceParameters{
							Region: "eu-central-1",
							Name:   aws.String("test"),
						},
					},
				},
			},
			want: want{
				cr: &svcapitypes.HTTPNamespace{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							"servicediscovery.aws.crossplane.io/operation-id": validOpID,
						},
					},
					Spec: svcapitypes.HTTPNamespaceSpec{
						ForProvider: svcapitypes.HTTPNamespaceParameters{
							Region: "eu-central-1",
							Name:   aws.String("test"),
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := NewHooks[namespace](tc.kube, tc.client)

			_, err := e.Delete(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions(),
				cmpopts.IgnoreFields(v1.Condition{}, "LastTransitionTime"),
			); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestUpdateTagsForResource(t *testing.T) {

	var actualTag []*svcsdk.Tag

	type args struct {
		client servicediscoveryiface.ServiceDiscoveryAPI
		tag    []*v1alpha1.Tag
		cr     v1.Object
	}
	tests := map[string]struct {
		args     args
		wantErr  bool
		wantTags []*svcapitypes.Tag
	}{
		"NoTagUpdate": {
			args: args{
				client: &fake.MockServicediscoveryClient{
					MockGetNamespace: func(input *svcsdk.GetNamespaceInput) (*svcsdk.GetNamespaceOutput, error) {
						return &svcsdk.GetNamespaceOutput{
							Namespace: &svcsdk.Namespace{},
						}, nil
					},
					MockListTagsForResource: func(input *svcsdk.ListTagsForResourceInput) (*svcsdk.ListTagsForResourceOutput, error) {
						actualTag = []*svcsdk.Tag{
							{
								Key:   aws.String(validTagKey1),
								Value: aws.String(validTagValue1),
							},
							{
								Key:   aws.String(validTagKey2),
								Value: aws.String(validTagValue2),
							},
							{
								Key:   aws.String(validTagKey3),
								Value: aws.String(validTagValue3),
							},
						}
						return &svcsdk.ListTagsForResourceOutput{
							Tags: []*svcsdk.Tag{
								{
									Key:   aws.String(validTagKey1),
									Value: aws.String(validTagValue1),
								},
								{
									Key:   aws.String(validTagKey2),
									Value: aws.String(validTagValue2),
								},
								{
									Key:   aws.String(validTagKey3),
									Value: aws.String(validTagValue3),
								},
							},
						}, nil
					},
					MockUntagResource: func(input *svcsdk.UntagResourceInput) (*svcsdk.UntagResourceOutput, error) {
						t.Error("no untag are necessary")
						return &svcsdk.UntagResourceOutput{}, nil
					},
					MockTagResource: func(input *svcsdk.TagResourceInput) (*svcsdk.TagResourceOutput, error) {
						t.Error("no untag are necessary")
						return &svcsdk.TagResourceOutput{}, nil
					},
				},
				tag: []*svcapitypes.Tag{
					{
						Key:   aws.String(validTagKey1),
						Value: aws.String(validTagValue1),
					},
					{
						Key:   aws.String(validTagKey2),
						Value: aws.String(validTagValue2),
					},
					{
						Key:   aws.String(validTagKey3),
						Value: aws.String(validTagValue3),
					},
				},
				cr: &svcapitypes.HTTPNamespace{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							"crossplane.io/external-name": validNSID,
						},
					},
					Spec: svcapitypes.HTTPNamespaceSpec{
						ForProvider: svcapitypes.HTTPNamespaceParameters{

							Tags: []*svcapitypes.Tag{
								{
									Key:   aws.String(validTagKey1),
									Value: aws.String(validTagValue1),
								},
								{
									Key:   aws.String(validTagKey2),
									Value: aws.String(validTagValue2),
								},
								{
									Key:   aws.String(validTagKey3),
									Value: aws.String(validTagValue3),
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		"AddDeleteTagUpdate": {
			args: args{
				client: &fake.MockServicediscoveryClient{
					MockGetNamespace: func(input *svcsdk.GetNamespaceInput) (*svcsdk.GetNamespaceOutput, error) {
						return &svcsdk.GetNamespaceOutput{
							Namespace: &svcsdk.Namespace{},
						}, nil
					},
					MockListTagsForResource: func(input *svcsdk.ListTagsForResourceInput) (*svcsdk.ListTagsForResourceOutput, error) {
						actualTag = []*svcsdk.Tag{
							{
								Key:   aws.String(validTagKey2),
								Value: aws.String(validTagValue2),
							},
							{
								Key:   aws.String(validTagKey3),
								Value: aws.String(validTagValue3),
							},
						}
						return &svcsdk.ListTagsForResourceOutput{
							Tags: []*svcsdk.Tag{
								{
									Key:   aws.String(validTagKey2),
									Value: aws.String(validTagValue2),
								},
								{
									Key:   aws.String(validTagKey3),
									Value: aws.String(validTagValue3),
								},
							},
						}, nil
					},
					MockUntagResource: func(input *svcsdk.UntagResourceInput) (*svcsdk.UntagResourceOutput, error) {
						var tmpTag []*svcsdk.Tag

						for _, el := range actualTag {
							var rm = false
							for _, rmK := range input.TagKeys {
								if aws.StringValue(el.Key) == aws.StringValue(rmK) {
									rm = true
									break
								}
							}
							if !rm {
								tmpTag = append(tmpTag, el)
							}
						}
						actualTag = tmpTag
						return &svcsdk.UntagResourceOutput{}, nil
					},
					MockTagResource: func(input *svcsdk.TagResourceInput) (*svcsdk.TagResourceOutput, error) {
						actualTag = append(actualTag, input.Tags...)
						return &svcsdk.TagResourceOutput{}, nil
					},
				},
				tag: []*svcapitypes.Tag{
					{
						Key:   aws.String(validTagKey1),
						Value: aws.String(validTagValue1),
					},
					{
						Key:   aws.String(validTagKey2),
						Value: aws.String(validTagValue2),
					},
				},
				cr: &svcapitypes.HTTPNamespace{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							"crossplane.io/external-name": validNSID,
						},
					},
					Spec: svcapitypes.HTTPNamespaceSpec{
						ForProvider: svcapitypes.HTTPNamespaceParameters{
							Tags: []*svcapitypes.Tag{
								{
									Key:   aws.String(validTagKey1),
									Value: aws.String(validTagValue1),
								},
								{
									Key:   aws.String(validTagKey2),
									Value: aws.String(validTagValue2),
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if err := UpdateTagsForResource(tt.args.client, tt.args.tag, tt.args.cr); (err != nil) != tt.wantErr {
				t.Errorf("UpdateTagsForResource() error = %v, wantErr %v", err, tt.wantErr)
			}

			add, remove := DiffTags(tt.args.tag, actualTag)
			if len(add) > 0 || len(remove) > 0 {
				t.Errorf("UpdateTagsForResource() diffAdd = %v, diffRemove %v", add, remove)
			}
		})
	}
}
