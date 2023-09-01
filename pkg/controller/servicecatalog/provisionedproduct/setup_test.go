/*
Copyright 2023 The Crossplane Authors.

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

package provisionedproduct

import (
	"context"
	"testing"
	"time"

	cfsdkv2types "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	svcsdk "github.com/aws/aws-sdk-go/service/servicecatalog"
	svcsdkapi "github.com/aws/aws-sdk-go/service/servicecatalog/servicecatalogiface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/servicecatalog/v1alpha1"
	aws "github.com/crossplane-contrib/provider-aws/pkg/clients"
	clientset "github.com/crossplane-contrib/provider-aws/pkg/clients/servicecatalog"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/servicecatalog/fake"
)

const (
	provisioningArtifactID          = "pa-1234567890"
	acceptLanguage                  = "jp"
	latelyInitializedAcceptLanguage = "en"
)

type args struct {
	kube                             client.Client
	client                           svcsdkapi.ServiceCatalogAPI
	customClient                     *fake.MockCustomServiceCatalogClient
	provisionedProduct               *v1alpha1.ProvisionedProduct
	describeProvisionedProductOutput *svcsdk.DescribeProvisionedProductOutput
}

type provisionedProductModifier func(provisionedProduct *v1alpha1.ProvisionedProduct)

func withSpec(p v1alpha1.ProvisionedProductParameters) provisionedProductModifier {
	return func(cr *v1alpha1.ProvisionedProduct) { cr.Spec.ForProvider = p }
}

func withStatus(p v1alpha1.ProvisionedProductStatus) provisionedProductModifier {
	return func(cr *v1alpha1.ProvisionedProduct) { cr.Status = p }
}

func provisionedProduct(m ...provisionedProductModifier) *v1alpha1.ProvisionedProduct {
	cr := &v1alpha1.ProvisionedProduct{}
	cr.Name = "test-provisioned-product-name"
	for _, f := range m {
		f(cr)
	}
	return cr
}

type describeProvisionedProductOutputModifier func(describeProvisionedProductOutput *svcsdk.DescribeProvisionedProductOutput)

func withDetails(d svcsdk.ProvisionedProductDetail) describeProvisionedProductOutputModifier {
	return func(output *svcsdk.DescribeProvisionedProductOutput) { output.ProvisionedProductDetail = &d }
}

func describeProvisionedProduct(m ...describeProvisionedProductOutputModifier) *svcsdk.DescribeProvisionedProductOutput {
	output := &svcsdk.DescribeProvisionedProductOutput{}
	for _, f := range m {
		f(output)
	}
	return output
}

func prepareFakeExternal(fakeClient clientset.Client) func(*external) {
	return func(e *external) {
		c := &custom{client: fakeClient}
		e.isUpToDate = c.isUpToDate
		e.lateInitialize = c.lateInitialize
		e.postObserve = c.postObserve
		e.preCreate = c.preCreate
		e.preDelete = c.preDelete
		e.preUpdate = c.preUpdate
	}
}

func TestIsUpToDate(t *testing.T) {
	provisioningArtifactID := provisioningArtifactID
	type want struct {
		result bool
		err    error
	}
	cases := map[string]struct {
		args
		want
	}{
		"ArtifactIdHasChanged": {
			args: args{
				provisionedProduct: provisionedProduct([]provisionedProductModifier{
					withSpec(v1alpha1.ProvisionedProductParameters{
						ProvisioningArtifactID: aws.String("pa-new1234567890"),
						ProductID:              aws.String("prod-1234567890"),
						ProvisioningParameters: []*v1alpha1.ProvisioningParameter{
							{Key: aws.String("Parameter"), Value: aws.String("foo")}},
					}),
				}...),
				describeProvisionedProductOutput: describeProvisionedProduct([]describeProvisionedProductOutputModifier{
					withDetails(svcsdk.ProvisionedProductDetail{
						Id:                     aws.String("pp-fake"),
						ProductId:              aws.String("prod-1234567890"),
						ProvisioningArtifactId: aws.String(provisioningArtifactID),
					}),
				}...),
				customClient: &fake.MockCustomServiceCatalogClient{
					MockGetCloudformationStackParameters: func(provisionedProductOutputs []*svcsdk.RecordOutput) ([]cfsdkv2types.Parameter, error) {
						return []cfsdkv2types.Parameter{{ParameterKey: aws.String("Parameter"), ParameterValue: aws.String("foo")}}, nil
					},
					MockGetProvisionedProductOutputs: func(getPPInput *svcsdk.GetProvisionedProductOutputsInput) (*svcsdk.GetProvisionedProductOutputsOutput, error) {
						return &svcsdk.GetProvisionedProductOutputsOutput{}, nil
					},
					MockDescribeProduct: func(dpInput *svcsdk.DescribeProductInput) (*svcsdk.DescribeProductOutput, error) {
						return &svcsdk.DescribeProductOutput{
							ProductViewSummary: &svcsdk.ProductViewSummary{
								ProductId: dpInput.Id,
								Name:      aws.String("fake product"),
							},
							ProvisioningArtifacts: []*svcsdk.ProvisioningArtifact{
								{
									Id: aws.String("prod-new1234567890"),
								},
							},
						}, nil
					},
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ParametersValueHasChanged": {
			args: args{
				provisionedProduct: provisionedProduct([]provisionedProductModifier{
					withSpec(v1alpha1.ProvisionedProductParameters{
						ProvisioningArtifactID: aws.String(provisioningArtifactID),
						ProvisioningParameters: []*v1alpha1.ProvisioningParameter{
							{Key: aws.String("ParameterIsNotChanged"), Value: aws.String("true")}},
					}),
				}...),
				describeProvisionedProductOutput: describeProvisionedProduct([]describeProvisionedProductOutputModifier{
					withDetails(svcsdk.ProvisionedProductDetail{
						Id:                     aws.String("pp-fake"),
						ProvisioningArtifactId: aws.String(provisioningArtifactID),
					}),
				}...),
				customClient: &fake.MockCustomServiceCatalogClient{
					MockGetCloudformationStackParameters: func(provisionedProductOutputs []*svcsdk.RecordOutput) ([]cfsdkv2types.Parameter, error) {
						return []cfsdkv2types.Parameter{{ParameterKey: aws.String("ParameterIsNotChanged"), ParameterValue: aws.String("false")}}, nil
					},
					MockGetProvisionedProductOutputs: func(getPPInput *svcsdk.GetProvisionedProductOutputsInput) (*svcsdk.GetProvisionedProductOutputsOutput, error) {
						return &svcsdk.GetProvisionedProductOutputsOutput{}, nil
					},
					MockDescribeProduct: func(dpInput *svcsdk.DescribeProductInput) (*svcsdk.DescribeProductOutput, error) {
						return &svcsdk.DescribeProductOutput{
							ProductViewSummary: &svcsdk.ProductViewSummary{
								ProductId: dpInput.Id,
								Name:      aws.String("fake product"),
							},
							ProvisioningArtifacts: []*svcsdk.ProvisioningArtifact{
								{
									Id: aws.String(provisioningArtifactID),
								},
							},
						}, nil
					},
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"NewParameterHasBeenAdded": {
			args: args{
				provisionedProduct: provisionedProduct([]provisionedProductModifier{
					withSpec(v1alpha1.ProvisionedProductParameters{
						ProvisioningArtifactID: aws.String(provisioningArtifactID),
						ProvisioningParameters: []*v1alpha1.ProvisioningParameter{
							{Key: aws.String("OldParameter"), Value: aws.String("foo")},
							{Key: aws.String("NewParameter"), Value: aws.String("bar")},
						},
					}),
				}...),
				describeProvisionedProductOutput: describeProvisionedProduct([]describeProvisionedProductOutputModifier{
					withDetails(svcsdk.ProvisionedProductDetail{
						Id:                     aws.String("pp-fake"),
						ProvisioningArtifactId: aws.String(provisioningArtifactID),
					}),
				}...),
				customClient: &fake.MockCustomServiceCatalogClient{
					MockGetCloudformationStackParameters: func(provisionedProductOutputs []*svcsdk.RecordOutput) ([]cfsdkv2types.Parameter, error) {
						return []cfsdkv2types.Parameter{
							{ParameterKey: aws.String("OldParameter"), ParameterValue: aws.String("foo")}}, nil
					},
					MockGetProvisionedProductOutputs: func(getPPInput *svcsdk.GetProvisionedProductOutputsInput) (*svcsdk.GetProvisionedProductOutputsOutput, error) {
						return &svcsdk.GetProvisionedProductOutputsOutput{}, nil
					},
					MockDescribeProduct: func(dpInput *svcsdk.DescribeProductInput) (*svcsdk.DescribeProductOutput, error) {
						return &svcsdk.DescribeProductOutput{
							ProductViewSummary: &svcsdk.ProductViewSummary{
								ProductId: aws.String("prod-fake"),
								Name:      aws.String("fake product"),
							},
							ProvisioningArtifacts: []*svcsdk.ProvisioningArtifact{
								{
									Id: aws.String(provisioningArtifactID),
								},
							},
						}, nil
					},
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ExistingParameterHasBeenRemoved": {
			args: args{
				provisionedProduct: provisionedProduct([]provisionedProductModifier{
					withSpec(v1alpha1.ProvisionedProductParameters{
						ProvisioningArtifactID: aws.String(provisioningArtifactID),
						ProvisioningParameters: []*v1alpha1.ProvisioningParameter{
							{Key: aws.String("Parameter1"), Value: aws.String("foo")},
						},
					}),
				}...),
				describeProvisionedProductOutput: describeProvisionedProduct([]describeProvisionedProductOutputModifier{
					withDetails(svcsdk.ProvisionedProductDetail{
						Id:                     aws.String("pp-fake"),
						ProvisioningArtifactId: aws.String(provisioningArtifactID),
					}),
				}...),
				customClient: &fake.MockCustomServiceCatalogClient{
					MockGetCloudformationStackParameters: func(provisionedProductOutputs []*svcsdk.RecordOutput) ([]cfsdkv2types.Parameter, error) {
						return []cfsdkv2types.Parameter{
								{ParameterKey: aws.String("Parameter1"), ParameterValue: aws.String("foo")},
								{ParameterKey: aws.String("Parameter2"), ParameterValue: aws.String("bar")},
							},
							nil
					},
					MockGetProvisionedProductOutputs: func(getPPInput *svcsdk.GetProvisionedProductOutputsInput) (*svcsdk.GetProvisionedProductOutputsOutput, error) {
						return &svcsdk.GetProvisionedProductOutputsOutput{}, nil
					},
					MockDescribeProduct: func(dpInput *svcsdk.DescribeProductInput) (*svcsdk.DescribeProductOutput, error) {
						return &svcsdk.DescribeProductOutput{
							ProductViewSummary: &svcsdk.ProductViewSummary{
								ProductId: dpInput.Id,
								Name:      aws.String("fake product"),
							},
							ProvisioningArtifacts: []*svcsdk.ProvisioningArtifact{
								{
									Id: aws.String(provisioningArtifactID),
								},
							},
						}, nil
					},
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ParametersAreNotChanged": {
			args: args{
				provisionedProduct: provisionedProduct([]provisionedProductModifier{
					withSpec(v1alpha1.ProvisionedProductParameters{
						ProvisioningArtifactID: aws.String(provisioningArtifactID),
						ProvisioningParameters: []*v1alpha1.ProvisioningParameter{
							{Key: aws.String("Parameter1"), Value: aws.String("foo")},
							{Key: aws.String("Parameter2"), Value: aws.String("bar")},
						},
					}),
				}...),
				describeProvisionedProductOutput: describeProvisionedProduct([]describeProvisionedProductOutputModifier{
					withDetails(svcsdk.ProvisionedProductDetail{
						Id:                     aws.String("pp-fake"),
						ProvisioningArtifactId: aws.String(provisioningArtifactID),
					}),
				}...),
				customClient: &fake.MockCustomServiceCatalogClient{
					MockGetCloudformationStackParameters: func(provisionedProductOutputs []*svcsdk.RecordOutput) ([]cfsdkv2types.Parameter, error) {
						return []cfsdkv2types.Parameter{
								{ParameterKey: aws.String("Parameter1"), ParameterValue: aws.String("foo")},
								{ParameterKey: aws.String("Parameter2"), ParameterValue: aws.String("bar")},
							},
							nil
					},
					MockGetProvisionedProductOutputs: func(getPPInput *svcsdk.GetProvisionedProductOutputsInput) (*svcsdk.GetProvisionedProductOutputsOutput, error) {
						return &svcsdk.GetProvisionedProductOutputsOutput{}, nil
					},
					MockDescribeProduct: func(dpInput *svcsdk.DescribeProductInput) (*svcsdk.DescribeProductOutput, error) {
						return &svcsdk.DescribeProductOutput{
							ProductViewSummary: &svcsdk.ProductViewSummary{
								ProductId: dpInput.Id,
								Name:      aws.String("fake product"),
							},
							ProvisioningArtifacts: []*svcsdk.ProvisioningArtifact{
								{
									Id: aws.String(provisioningArtifactID),
								},
							},
						}, nil
					},
				},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			opts := []option{prepareFakeExternal(tc.args.customClient)}
			e := newExternal(tc.args.kube, tc.args.client, opts)
			result, _, err := e.isUpToDate(nil, tc.args.provisionedProduct, tc.args.describeProvisionedProductOutput)
			if diff := cmp.Diff(err, tc.want.err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(result, tc.want.result); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitialize(t *testing.T) {
	type want struct {
		acceptLanguage string
	}
	cases := map[string]struct {
		args
		want
	}{
		"ValuesAreNotSpecified": {
			args: args{
				provisionedProduct: provisionedProduct([]provisionedProductModifier{
					withSpec(v1alpha1.ProvisionedProductParameters{
						AcceptLanguage: aws.String(""),
					}),
				}...),
			},
			want: want{
				acceptLanguage: latelyInitializedAcceptLanguage,
			},
		},
		"ValuesAreSpecified": {
			args: args{
				provisionedProduct: provisionedProduct([]provisionedProductModifier{
					withSpec(v1alpha1.ProvisionedProductParameters{
						AcceptLanguage: aws.String(acceptLanguage),
					}),
				}...),
			},
			want: want{
				acceptLanguage: acceptLanguage,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			opts := []option{prepareFakeExternal(tc.args.customClient)}
			e := newExternal(tc.args.kube, tc.args.client, opts)
			_ = e.lateInitialize(&tc.args.provisionedProduct.Spec.ForProvider, tc.args.describeProvisionedProductOutput)
			if diff := cmp.Diff(*tc.args.provisionedProduct.Spec.ForProvider.AcceptLanguage, tc.want.acceptLanguage); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestPostObserve(t *testing.T) {
	type want struct {
		status xpv1.Condition
	}
	testStaringTime := time.Now()
	provisionedProductStatus := withStatus(v1alpha1.ProvisionedProductStatus{
		ResourceStatus: xpv1.ResourceStatus{
			ConditionedStatus: xpv1.ConditionedStatus{
				Conditions: []xpv1.Condition{}}},
	})
	cases := map[string]struct {
		args
		want
	}{
		"StatusAvailable": {
			args: args{
				provisionedProduct: provisionedProduct([]provisionedProductModifier{provisionedProductStatus}...),
				describeProvisionedProductOutput: describeProvisionedProduct([]describeProvisionedProductOutputModifier{
					withDetails(svcsdk.ProvisionedProductDetail{
						Status:                             aws.String(string(v1alpha1.ProvisionedProductStatus_SDK_AVAILABLE)),
						Arn:                                aws.String("arn:aws:servicecatalog:fake"),
						CreatedTime:                        &testStaringTime,
						LastSuccessfulProvisioningRecordId: aws.String("rec-fake"),
						LaunchRoleArn:                      aws.String("arn:aws:iam::fake"),
						StatusMessage:                      aws.String("fake"),
						Type:                               aws.String("CFN_STACK"),
					}),
				}...),
				customClient: &fake.MockCustomServiceCatalogClient{
					MockDescribeRecord: func(describeRecordInput *svcsdk.DescribeRecordInput) (*svcsdk.DescribeRecordOutput, error) {
						return &svcsdk.DescribeRecordOutput{RecordDetail: &svcsdk.RecordDetail{RecordType: aws.String("PROVISION_PRODUCT")}}, nil
					},
				},
			},
			want: want{
				status: xpv1.Available(),
			},
		},
		"StatusUnavailableWithAmendment": {
			args: args{
				provisionedProduct: provisionedProduct([]provisionedProductModifier{provisionedProductStatus}...),
				describeProvisionedProductOutput: describeProvisionedProduct([]describeProvisionedProductOutputModifier{
					withDetails(svcsdk.ProvisionedProductDetail{
						Status:                             aws.String(string(v1alpha1.ProvisionedProductStatus_SDK_UNDER_CHANGE)),
						Arn:                                aws.String("arn:aws:servicecatalog:fake"),
						CreatedTime:                        &testStaringTime,
						LastSuccessfulProvisioningRecordId: aws.String("rec-fake"),
						LaunchRoleArn:                      aws.String("arn:aws:iam::fake"),
						StatusMessage:                      aws.String("fake"),
						Type:                               aws.String("CFN_STACK"),
					}),
				}...),
				customClient: &fake.MockCustomServiceCatalogClient{
					MockDescribeRecord: func(describeRecordInput *svcsdk.DescribeRecordInput) (*svcsdk.DescribeRecordOutput, error) {
						return &svcsdk.DescribeRecordOutput{RecordDetail: &svcsdk.RecordDetail{RecordType: aws.String("UPDATE_PROVISIONED_PRODUCT")}}, nil
					},
				},
			},
			want: want{
				status: xpv1.Unavailable().WithMessage(msgProvisionedProductStatusSdkUnderChange),
			},
		},
		"StatusReconcileErrorProductError": {
			args: args{
				provisionedProduct: provisionedProduct([]provisionedProductModifier{provisionedProductStatus}...),
				describeProvisionedProductOutput: describeProvisionedProduct([]describeProvisionedProductOutputModifier{
					withDetails(svcsdk.ProvisionedProductDetail{
						Status:                             aws.String(string(v1alpha1.ProvisionedProductStatus_SDK_ERROR)),
						Arn:                                aws.String("arn:aws:servicecatalog:fake"),
						CreatedTime:                        &testStaringTime,
						LastSuccessfulProvisioningRecordId: aws.String("rec-fake"),
						LaunchRoleArn:                      aws.String("arn:aws:iam::fake"),
						StatusMessage:                      aws.String("fake"),
						Type:                               aws.String("CFN_STACK"),
					}),
				}...),
				customClient: &fake.MockCustomServiceCatalogClient{
					MockDescribeRecord: func(describeRecordInput *svcsdk.DescribeRecordInput) (*svcsdk.DescribeRecordOutput, error) {
						return &svcsdk.DescribeRecordOutput{RecordDetail: &svcsdk.RecordDetail{RecordType: aws.String("PROVISION_PRODUCT")}}, nil
					},
				},
			},
			want: want{
				status: xpv1.Unavailable().WithMessage(msgProvisionedProductStatusSdkError),
			},
		},
		"StatusReconcileErrorProductTainted": {
			args: args{
				provisionedProduct: provisionedProduct([]provisionedProductModifier{provisionedProductStatus}...),
				describeProvisionedProductOutput: describeProvisionedProduct([]describeProvisionedProductOutputModifier{
					withDetails(svcsdk.ProvisionedProductDetail{
						Status:                             aws.String(string(v1alpha1.ProvisionedProductStatus_SDK_TAINTED)),
						Arn:                                aws.String("arn:aws:servicecatalog:fake"),
						CreatedTime:                        &testStaringTime,
						LastSuccessfulProvisioningRecordId: aws.String("rec-fake"),
						LaunchRoleArn:                      aws.String("arn:aws:iam::fake"),
						StatusMessage:                      aws.String("fake"),
						Type:                               aws.String("CFN_STACK"),
					}),
				}...),
				customClient: &fake.MockCustomServiceCatalogClient{
					MockDescribeRecord: func(describeRecordInput *svcsdk.DescribeRecordInput) (*svcsdk.DescribeRecordOutput, error) {
						return &svcsdk.DescribeRecordOutput{RecordDetail: &svcsdk.RecordDetail{RecordType: aws.String("PROVISION_PRODUCT")}}, nil
					},
				},
			},
			want: want{
				status: xpv1.Unavailable().WithMessage(msgProvisionedProductStatusSdkTainted),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			opts := []option{prepareFakeExternal(tc.args.customClient)}
			e := newExternal(tc.args.kube, tc.args.client, opts)
			_, _ = e.postObserve(context.TODO(), tc.args.provisionedProduct, tc.args.describeProvisionedProductOutput, managed.ExternalObservation{}, nil)
			conditions := tc.args.provisionedProduct.Status.Conditions
			latestCondition := conditions[len(conditions)-1]
			if diff := cmp.Diff(latestCondition, tc.want.status); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			test.EquateConditions()
		})
	}
}

func TestPreDelete(t *testing.T) {
	type want struct {
		ignoreDeletion bool
	}
	cases := map[string]struct {
		args
		want
	}{
		"ignoreDeletion": {
			args: args{
				provisionedProduct: provisionedProduct([]provisionedProductModifier{
					withStatus(v1alpha1.ProvisionedProductStatus{
						AtProvider: v1alpha1.ProvisionedProductObservation{
							Status: aws.String(string(v1alpha1.ProvisionedProductStatus_SDK_UNDER_CHANGE))},
					}),
				}...),
			},
			want: want{
				ignoreDeletion: true,
			},
		},
		"passDeletion": {
			args: args{
				provisionedProduct: provisionedProduct([]provisionedProductModifier{
					withStatus(v1alpha1.ProvisionedProductStatus{
						AtProvider: v1alpha1.ProvisionedProductObservation{
							Status: aws.String("NOT_UNDER_CHANGE")},
					}),
				}...),
			},
			want: want{
				ignoreDeletion: false,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			opts := []option{prepareFakeExternal(tc.args.customClient)}
			e := newExternal(tc.args.kube, tc.args.client, opts)
			ignore, _ := e.preDelete(context.TODO(), tc.args.provisionedProduct, &svcsdk.TerminateProvisionedProductInput{})
			if diff := cmp.Diff(ignore, tc.want.ignoreDeletion); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)

			}
		})
	}
}
