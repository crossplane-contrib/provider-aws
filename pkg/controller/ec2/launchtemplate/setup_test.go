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

package launchtemplate

import (
	"context"
	"testing"
	"time"

	svcsdk "github.com/aws/aws-sdk-go/service/ec2"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"k8s.io/utils/ptr"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/ec2/v1alpha1"
	ec2mock "github.com/crossplane-contrib/provider-aws/pkg/clients/mock/ec2iface"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

var (
	timeNow     = time.Now()
	timeNowMeta = pointer.TimeToMetaTime(&timeNow)
)

type launchTemplateModifier func(*svcapitypes.LaunchTemplate)

func withExternalName(name string) launchTemplateModifier {
	return func(r *svcapitypes.LaunchTemplate) { meta.SetExternalName(r, name) }
}

func withConditions(c ...xpv1.Condition) launchTemplateModifier {
	return func(r *svcapitypes.LaunchTemplate) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpec(p svcapitypes.LaunchTemplateParameters) launchTemplateModifier {
	return func(r *svcapitypes.LaunchTemplate) { r.Spec.ForProvider = p }
}

func withStatus(s svcapitypes.LaunchTemplateObservation) launchTemplateModifier {
	return func(r *svcapitypes.LaunchTemplate) { r.Status.AtProvider = s }
}

func launchTemplate(m ...launchTemplateModifier) *svcapitypes.LaunchTemplate {
	cr := &svcapitypes.LaunchTemplate{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

type ec2MockModifier func(m *ec2mock.MockEC2API)

func TestObserve(t *testing.T) {
	type args struct {
		cr  *svcapitypes.LaunchTemplate
		ec2 ec2MockModifier
	}

	type want struct {
		cr     *svcapitypes.LaunchTemplate
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulAvailable": {
			args: args{
				cr: launchTemplate(
					withExternalName("test-name"),
					withSpec(svcapitypes.LaunchTemplateParameters{
						Region: "us-east-1",
					}),
				),
				ec2: func(m *ec2mock.MockEC2API) {
					m.EXPECT().DescribeLaunchTemplatesWithContext(context.Background(), &svcsdk.DescribeLaunchTemplatesInput{
						LaunchTemplateNames: []*string{ptr.To("test-name")},
					}).Return(&svcsdk.DescribeLaunchTemplatesOutput{
						LaunchTemplates: []*svcsdk.LaunchTemplate{
							{
								CreateTime:           &timeNow,
								CreatedBy:            ptr.To("test"),
								DefaultVersionNumber: ptr.To(int64(1)),
								LatestVersionNumber:  ptr.To(int64(2)),
								LaunchTemplateId:     ptr.To("test-id"),
								LaunchTemplateName:   ptr.To("test-name"),
								Tags: []*svcsdk.Tag{
									{
										Key:   ptr.To("foo"),
										Value: ptr.To("bar"),
									},
								},
							},
						},
					}, nil)
				},
			},
			want: want{
				cr: launchTemplate(
					withExternalName("test-name"),
					withSpec(svcapitypes.LaunchTemplateParameters{
						Region: "us-east-1",
					}),
					withConditions(xpv1.Available()),
					withStatus(svcapitypes.LaunchTemplateObservation{
						LaunchTemplate: &svcapitypes.LaunchTemplate_SDK{
							CreateTime:           timeNowMeta,
							CreatedBy:            ptr.To("test"),
							DefaultVersionNumber: ptr.To(int64(1)),
							LatestVersionNumber:  ptr.To(int64(2)),
							LaunchTemplateID:     ptr.To("test-id"),
							LaunchTemplateName:   ptr.To("test-name"),
							Tags: []*svcapitypes.Tag{
								{
									Key:   ptr.To("foo"),
									Value: ptr.To("bar"),
								},
							},
						},
					}),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ec2APIMock := ec2mock.NewMockEC2API(gomock.NewController(t))
			if tc.args.ec2 != nil {
				tc.args.ec2(ec2APIMock)
			}
			e := newExternal(nil, ec2APIMock, []option{setupExternal()})
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
