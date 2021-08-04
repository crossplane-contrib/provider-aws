/*
Copyright 2020 The Crossplane Authors.
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

package resolverruleassociation

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsr53r "github.com/aws/aws-sdk-go-v2/service/route53resolver"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/provider-aws/apis/route53resolver/v1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/route53resolver"
	"github.com/crossplane/provider-aws/pkg/clients/route53resolver/fake"
)

var (
	id      = "some id"
	vpcID   = "some vpc id"
	ruleID  = "some resolver rule id"
	errBoom = errors.New("boom")
)

type args struct {
	resolverruleassociation route53resolver.ResolverRuleAssociationClient
	kube                    client.Client
	cr                      *v1alpha1.ResolverRuleAssociation
}

type resolverRuleAssociationModifier func(*v1alpha1.ResolverRuleAssociation)

func withExternalName(name string) resolverRuleAssociationModifier {
	return func(r *v1alpha1.ResolverRuleAssociation) { meta.SetExternalName(r, name) }
}

func withConditions(c ...xpv1.Condition) resolverRuleAssociationModifier {
	return func(r *v1alpha1.ResolverRuleAssociation) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpec(p v1alpha1.ResolverRuleAssociationParameters) resolverRuleAssociationModifier {
	return func(r *v1alpha1.ResolverRuleAssociation) { r.Spec.ForProvider = p }
}

func withStatus(s v1alpha1.ResolverRuleAssociationObservation) resolverRuleAssociationModifier {
	return func(r *v1alpha1.ResolverRuleAssociation) { r.Status.AtProvider = s }
}

func resolverruleassociation(m ...resolverRuleAssociationModifier) *v1alpha1.ResolverRuleAssociation {
	cr := &v1alpha1.ResolverRuleAssociation{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connector{}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1alpha1.ResolverRuleAssociation
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
				},
				resolverruleassociation: &fake.MockResolverRuleAssociationClient{
					MockGet: func(input *awsr53r.GetResolverRuleAssociationInput) awsr53r.GetResolverRuleAssociationRequest {
						return awsr53r.GetResolverRuleAssociationRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsr53r.GetResolverRuleAssociationOutput{
								ResolverRuleAssociation: &awsr53r.ResolverRuleAssociation{
									Id: &id,
								},
							},
							},
						}
					},
				},
				cr: resolverruleassociation(withSpec(v1alpha1.ResolverRuleAssociationParameters{
					VPCID: &vpcID,
				}), withExternalName(id)),
			},

			want: want{
				cr: resolverruleassociation(withSpec(v1alpha1.ResolverRuleAssociationParameters{
					VPCID: &vpcID,
				}), withStatus(v1alpha1.ResolverRuleAssociationObservation{
					ID: id,
				}), withExternalName(id),
					withConditions(xpv1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"Fail": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
				},
				resolverruleassociation: &fake.MockResolverRuleAssociationClient{
					MockGet: func(input *awsr53r.GetResolverRuleAssociationInput) awsr53r.GetResolverRuleAssociationRequest {
						return awsr53r.GetResolverRuleAssociationRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: resolverruleassociation(withSpec(v1alpha1.ResolverRuleAssociationParameters{
					VPCID: &vpcID,
				}), withExternalName(id)),
			},

			want: want{
				cr: resolverruleassociation(withSpec(v1alpha1.ResolverRuleAssociationParameters{
					VPCID: &vpcID,
				}), withExternalName(id)),
				err: awsclient.Wrap(errBoom, errDescribe),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.resolverruleassociation}
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
		cr     *v1alpha1.ResolverRuleAssociation
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
				resolverruleassociation: &fake.MockResolverRuleAssociationClient{
					MockAssociate: func(input *awsr53r.AssociateResolverRuleInput) awsr53r.AssociateResolverRuleRequest {
						return awsr53r.AssociateResolverRuleRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsr53r.AssociateResolverRuleOutput{
								ResolverRuleAssociation: &awsr53r.ResolverRuleAssociation{
									Id: &id,
								},
							},
							},
						}
					},
				},
				cr: resolverruleassociation(withSpec(v1alpha1.ResolverRuleAssociationParameters{
					VPCID:          &vpcID,
					ResolverRuleID: &ruleID,
				})),
			},

			want: want{
				cr: resolverruleassociation(withSpec(v1alpha1.ResolverRuleAssociationParameters{
					VPCID:          &vpcID,
					ResolverRuleID: &ruleID,
				}), withExternalName(id),
					withConditions(xpv1.Creating())),
				result: managed.ExternalCreation{ExternalNameAssigned: true},
			},
		},
		"Fail": {
			args: args{
				kube: &test.MockClient{
					MockUpdate:       test.NewMockClient().Update,
					MockStatusUpdate: test.NewMockClient().MockStatusUpdate,
				},
				resolverruleassociation: &fake.MockResolverRuleAssociationClient{
					MockAssociate: func(input *awsr53r.AssociateResolverRuleInput) awsr53r.AssociateResolverRuleRequest {
						return awsr53r.AssociateResolverRuleRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: resolverruleassociation(withSpec(v1alpha1.ResolverRuleAssociationParameters{
					VPCID:          &vpcID,
					ResolverRuleID: &ruleID,
				})),
			},

			want: want{
				cr: resolverruleassociation(withSpec(v1alpha1.ResolverRuleAssociationParameters{
					VPCID:          &vpcID,
					ResolverRuleID: &ruleID,
				}), withConditions(xpv1.Creating())),
				err: awsclient.Wrap(errBoom, errCreate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.resolverruleassociation}
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

func TestDelete(t *testing.T) {
	type want struct {
		cr  *v1alpha1.ResolverRuleAssociation
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				resolverruleassociation: &fake.MockResolverRuleAssociationClient{
					MockDisassociate: func(input *awsr53r.DisassociateResolverRuleInput) awsr53r.DisassociateResolverRuleRequest {
						return awsr53r.DisassociateResolverRuleRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsr53r.DisassociateResolverRuleOutput{}},
						}
					},
				},
				cr: resolverruleassociation(),
			},
			want: want{
				cr: resolverruleassociation(
					withConditions(xpv1.Deleting())),
			},
		},
		"Fail": {
			args: args{
				resolverruleassociation: &fake.MockResolverRuleAssociationClient{
					MockDisassociate: func(input *awsr53r.DisassociateResolverRuleInput) awsr53r.DisassociateResolverRuleRequest {
						return awsr53r.DisassociateResolverRuleRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: resolverruleassociation(),
			},
			want: want{
				cr: resolverruleassociation(
					withConditions(xpv1.Deleting())),
				err: awsclient.Wrap(errBoom, errDelete),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.resolverruleassociation}
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
