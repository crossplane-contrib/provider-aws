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

package internetgateway

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-aws/apis/ec2/v1beta1"
	"github.com/crossplane/provider-aws/pkg/clients/ec2"
	"github.com/crossplane/provider-aws/pkg/clients/ec2/fake"
)

var (
	vpcID        = "some vpc"
	anotherVpcID = "another vpc"
	igID         = "some ID"

	errBoom = errors.New("boom")
)

type args struct {
	ig   ec2.InternetGatewayClient
	kube client.Client
	cr   *v1beta1.InternetGateway
}

type igModifier func(*v1beta1.InternetGateway)

func withExternalName(name string) igModifier {
	return func(r *v1beta1.InternetGateway) { meta.SetExternalName(r, name) }
}

func withConditions(c ...xpv1.Condition) igModifier {
	return func(r *v1beta1.InternetGateway) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpec(p v1beta1.InternetGatewayParameters) igModifier {
	return func(r *v1beta1.InternetGateway) { r.Spec.ForProvider = p }
}

func withStatus(s v1beta1.InternetGatewayObservation) igModifier {
	return func(r *v1beta1.InternetGateway) { r.Status.AtProvider = s }
}

func ig(m ...igModifier) *v1beta1.InternetGateway {
	cr := &v1beta1.InternetGateway{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func igAttachments() []awsec2.InternetGatewayAttachment {
	return []awsec2.InternetGatewayAttachment{
		{
			VpcId: aws.String(vpcID),
			State: v1beta1.AttachmentStatusAvailable,
		},
	}
}

func specAttachments() []v1beta1.InternetGatewayAttachment {
	return []v1beta1.InternetGatewayAttachment{
		{
			AttachmentStatus: v1beta1.AttachmentStatusAvailable,
			VPCID:            vpcID,
		},
	}
}

var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connector{}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1beta1.InternetGateway
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulAvailable": {
			args: args{
				ig: &fake.MockInternetGatewayClient{
					MockDescribe: func(input *awsec2.DescribeInternetGatewaysInput) awsec2.DescribeInternetGatewaysRequest {
						return awsec2.DescribeInternetGatewaysRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeInternetGatewaysOutput{
								InternetGateways: []awsec2.InternetGateway{
									{
										Attachments: igAttachments(),
									},
								},
							}},
						}
					},
				},
				cr: ig(withSpec(v1beta1.InternetGatewayParameters{
					VPCID: aws.String(vpcID),
				}),
					withStatus(v1beta1.InternetGatewayObservation{
						InternetGatewayID: igID,
						Attachments:       specAttachments(),
					}),
					withExternalName(igID)),
			},
			want: want{
				cr: ig(withSpec(v1beta1.InternetGatewayParameters{
					VPCID: aws.String(vpcID),
				}),
					withStatus(v1beta1.InternetGatewayObservation{
						Attachments: specAttachments(),
					}),
					withExternalName(igID),
					withConditions(xpv1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"MultipleIGs": {
			args: args{
				ig: &fake.MockInternetGatewayClient{
					MockDescribe: func(input *awsec2.DescribeInternetGatewaysInput) awsec2.DescribeInternetGatewaysRequest {
						return awsec2.DescribeInternetGatewaysRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeInternetGatewaysOutput{
								InternetGateways: []awsec2.InternetGateway{
									{},
									{},
								},
							},
							},
						}
					},
				},
				cr: ig(withStatus(v1beta1.InternetGatewayObservation{
					InternetGatewayID: igID,
					Attachments:       specAttachments(),
				}),
					withExternalName(igID)),
			},
			want: want{
				cr: ig(withStatus(v1beta1.InternetGatewayObservation{
					InternetGatewayID: igID,
					Attachments:       specAttachments(),
				}),
					withExternalName(igID)),
				err: errors.Errorf(errNotSingleItem),
			},
		},
		"FailedRequest": {
			args: args{
				ig: &fake.MockInternetGatewayClient{
					MockDescribe: func(input *awsec2.DescribeInternetGatewaysInput) awsec2.DescribeInternetGatewaysRequest {
						return awsec2.DescribeInternetGatewaysRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: ig(withStatus(v1beta1.InternetGatewayObservation{
					InternetGatewayID: igID,
					Attachments:       specAttachments(),
				}),
					withExternalName(igID)),
			},
			want: want{
				cr: ig(withStatus(v1beta1.InternetGatewayObservation{
					InternetGatewayID: igID,
					Attachments:       specAttachments(),
				}),
					withExternalName(igID)),
				err: errors.Wrap(errBoom, errDescribe),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.ig}
			o, err := e.Observe(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	type want struct {
		cr     *v1beta1.InternetGateway
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				kube: &test.MockClient{
					MockUpdate:       test.NewMockClient().Update,
					MockStatusUpdate: test.NewMockClient().MockStatusUpdate,
				},
				ig: &fake.MockInternetGatewayClient{
					MockCreate: func(input *awsec2.CreateInternetGatewayInput) awsec2.CreateInternetGatewayRequest {
						return awsec2.CreateInternetGatewayRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.CreateInternetGatewayOutput{
								InternetGateway: &awsec2.InternetGateway{
									Attachments:       igAttachments(),
									InternetGatewayId: aws.String(igID),
								},
							}},
						}
					},
				},
				cr: ig(withSpec(v1beta1.InternetGatewayParameters{
					VPCID: aws.String(vpcID),
				})),
			},
			want: want{
				cr: ig(withSpec(v1beta1.InternetGatewayParameters{
					VPCID: aws.String(vpcID),
				}),
					withExternalName(igID),
					withConditions(xpv1.Creating())),
				result: managed.ExternalCreation{ExternalNameAssigned: true},
			},
		},
		"FailedRequest": {
			args: args{
				kube: &test.MockClient{
					MockStatusUpdate: test.NewMockClient().MockStatusUpdate,
				},
				ig: &fake.MockInternetGatewayClient{
					MockCreate: func(input *awsec2.CreateInternetGatewayInput) awsec2.CreateInternetGatewayRequest {
						return awsec2.CreateInternetGatewayRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: ig(),
			},
			want: want{
				cr:  ig(withConditions(xpv1.Creating())),
				err: errors.Wrap(errBoom, errCreate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.ig}
			o, err := e.Create(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	type want struct {
		cr     *v1beta1.InternetGateway
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				ig: &fake.MockInternetGatewayClient{
					MockDescribe: func(input *awsec2.DescribeInternetGatewaysInput) awsec2.DescribeInternetGatewaysRequest {
						return awsec2.DescribeInternetGatewaysRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeInternetGatewaysOutput{
								InternetGateways: []awsec2.InternetGateway{{
									Attachments: igAttachments(),
								}},
							}},
						}
					},
					MockAttach: func(input *awsec2.AttachInternetGatewayInput) awsec2.AttachInternetGatewayRequest {
						return awsec2.AttachInternetGatewayRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.AttachInternetGatewayOutput{}},
						}
					},
					MockDetach: func(input *awsec2.DetachInternetGatewayInput) awsec2.DetachInternetGatewayRequest {
						return awsec2.DetachInternetGatewayRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DetachInternetGatewayOutput{}},
						}
					},
					MockCreateTags: func(input *awsec2.CreateTagsInput) awsec2.CreateTagsRequest {
						return awsec2.CreateTagsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.CreateTagsOutput{}},
						}
					},
				},
				cr: ig(withSpec(v1beta1.InternetGatewayParameters{
					VPCID: aws.String(anotherVpcID),
				}), withStatus(v1beta1.InternetGatewayObservation{
					InternetGatewayID: igID,
				})),
			},
			want: want{
				cr: ig(withSpec(v1beta1.InternetGatewayParameters{
					VPCID: aws.String(anotherVpcID),
				}), withStatus(v1beta1.InternetGatewayObservation{
					InternetGatewayID: igID,
				})),
			},
		},
		"NoUpdateNeeded": {
			args: args{
				ig: &fake.MockInternetGatewayClient{
					MockDescribe: func(input *awsec2.DescribeInternetGatewaysInput) awsec2.DescribeInternetGatewaysRequest {
						return awsec2.DescribeInternetGatewaysRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeInternetGatewaysOutput{
								InternetGateways: []awsec2.InternetGateway{{
									Attachments: igAttachments(),
								}},
							}},
						}
					},
					MockAttach: func(input *awsec2.AttachInternetGatewayInput) awsec2.AttachInternetGatewayRequest {
						return awsec2.AttachInternetGatewayRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.AttachInternetGatewayOutput{}},
						}
					},
					MockDetach: func(input *awsec2.DetachInternetGatewayInput) awsec2.DetachInternetGatewayRequest {
						return awsec2.DetachInternetGatewayRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DetachInternetGatewayOutput{}},
						}
					},
					MockCreateTags: func(input *awsec2.CreateTagsInput) awsec2.CreateTagsRequest {
						return awsec2.CreateTagsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.CreateTagsOutput{}},
						}
					},
				},
				cr: ig(withStatus(v1beta1.InternetGatewayObservation{
					InternetGatewayID: igID,
				})),
			},
			want: want{
				cr: ig(withStatus(v1beta1.InternetGatewayObservation{
					InternetGatewayID: igID,
				})),
			},
		},
		"DetachFail": {
			args: args{
				ig: &fake.MockInternetGatewayClient{
					MockDescribe: func(input *awsec2.DescribeInternetGatewaysInput) awsec2.DescribeInternetGatewaysRequest {
						return awsec2.DescribeInternetGatewaysRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DescribeInternetGatewaysOutput{
								InternetGateways: []awsec2.InternetGateway{{
									Attachments: igAttachments(),
								}},
							}},
						}
					},
					MockDetach: func(input *awsec2.DetachInternetGatewayInput) awsec2.DetachInternetGatewayRequest {
						return awsec2.DetachInternetGatewayRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
					MockCreateTags: func(input *awsec2.CreateTagsInput) awsec2.CreateTagsRequest {
						return awsec2.CreateTagsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.CreateTagsOutput{}},
						}
					},
				},
				cr: ig(withSpec(v1beta1.InternetGatewayParameters{
					VPCID: aws.String(anotherVpcID),
				}), withExternalName(igID)),
			},
			want: want{
				cr: ig(withSpec(v1beta1.InternetGatewayParameters{
					VPCID: aws.String(anotherVpcID),
				}), withExternalName(igID)),
				err: errors.Wrap(errBoom, errDetach),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.ig}
			o, err := e.Update(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
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
		cr  *v1beta1.InternetGateway
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				ig: &fake.MockInternetGatewayClient{
					MockDelete: func(input *awsec2.DeleteInternetGatewayInput) awsec2.DeleteInternetGatewayRequest {
						return awsec2.DeleteInternetGatewayRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DeleteInternetGatewayOutput{}},
						}
					},
					MockDetach: func(input *awsec2.DetachInternetGatewayInput) awsec2.DetachInternetGatewayRequest {
						return awsec2.DetachInternetGatewayRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DetachInternetGatewayOutput{}},
						}
					},
				},
				cr: ig(withStatus(v1beta1.InternetGatewayObservation{
					InternetGatewayID: igID,
					Attachments:       specAttachments(),
				}), withExternalName(igID)),
			},
			want: want{
				cr: ig(withStatus(v1beta1.InternetGatewayObservation{
					InternetGatewayID: igID,
					Attachments:       specAttachments(),
				}), withExternalName(igID),
					withConditions(xpv1.Deleting())),
			},
		},
		"NotAvailable": {
			args: args{
				ig: &fake.MockInternetGatewayClient{
					MockDelete: func(input *awsec2.DeleteInternetGatewayInput) awsec2.DeleteInternetGatewayRequest {
						return awsec2.DeleteInternetGatewayRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DeleteInternetGatewayOutput{}},
						}
					},
					MockDetach: func(input *awsec2.DetachInternetGatewayInput) awsec2.DetachInternetGatewayRequest {
						return awsec2.DetachInternetGatewayRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DetachInternetGatewayOutput{}},
						}
					},
				},
				cr: ig(),
			},
			want: want{
				cr: ig(withConditions(xpv1.Deleting())),
			},
		},
		"DetachFail": {
			args: args{
				ig: &fake.MockInternetGatewayClient{
					MockDelete: func(input *awsec2.DeleteInternetGatewayInput) awsec2.DeleteInternetGatewayRequest {
						return awsec2.DeleteInternetGatewayRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DeleteInternetGatewayOutput{}},
						}
					},
					MockDetach: func(input *awsec2.DetachInternetGatewayInput) awsec2.DetachInternetGatewayRequest {
						return awsec2.DetachInternetGatewayRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: ig(withStatus(v1beta1.InternetGatewayObservation{
					InternetGatewayID: igID,
					Attachments:       specAttachments(),
				}), withExternalName(igID)),
			},
			want: want{
				cr: ig(withStatus(v1beta1.InternetGatewayObservation{
					InternetGatewayID: igID,
					Attachments:       specAttachments(),
				}), withExternalName(igID),
					withConditions(xpv1.Deleting())),
				err: errors.Wrap(errBoom, errDetach),
			},
		},
		"DeleteFail": {
			args: args{
				ig: &fake.MockInternetGatewayClient{
					MockDelete: func(input *awsec2.DeleteInternetGatewayInput) awsec2.DeleteInternetGatewayRequest {
						return awsec2.DeleteInternetGatewayRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
					MockDetach: func(input *awsec2.DetachInternetGatewayInput) awsec2.DetachInternetGatewayRequest {
						return awsec2.DetachInternetGatewayRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsec2.DetachInternetGatewayOutput{}},
						}
					},
				},
				cr: ig(withStatus(v1beta1.InternetGatewayObservation{
					InternetGatewayID: igID,
					Attachments:       specAttachments(),
				}), withExternalName(igID)),
			},
			want: want{
				cr: ig(withStatus(v1beta1.InternetGatewayObservation{
					InternetGatewayID: igID,
					Attachments:       specAttachments(),
				}), withExternalName(igID),
					withConditions(xpv1.Deleting())),
				err: errors.Wrap(errBoom, errDelete),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.ig}
			err := e.Delete(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
