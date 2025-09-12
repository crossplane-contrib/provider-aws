/*
Copyright 2025 The Crossplane Authors.

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

package webacl

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/wafv2/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/wafv2/fake"
)

type args struct {
	desired  *svcapitypes.WebACL
	observed *svcsdk.GetWebACLOutput
	client   *fake.MockWAFV2Client
	//	cache    *cache
}

func TestIsUpToDate(t *testing.T) {
	webAclName := "webacl"
	webAclId := "88c4049f-eba9-4666-b9a9-f6aec5b5b41b"

	visibilityConfigMetricName := "metricName"
	visibilityConfigSampledRequestsEnabled := false
	visibilityConfigCloudWatchMetricsEnabled := false

	scope := "REGIONAL"

	tag0Key := "used-by-elb"
	tag0Value := "true"
	tag0NewValue := "false"

	tag1Key := "managed-by-crossplane"
	tag1Value := "true"

	ruleName0 := "ruleName0"
	ruleName1 := "ruleName1"
	rulePriority0 := int64(0)
	rulePriority1 := int64(1)
	ruleAndStatement0 := ` {
	              "Statements": [
{
	                  "ByteMatchStatement":
	                    {
	                      "FieldToMatch": {
	                        "SingleHeader": {
	                          "Name": "User-Agent"
	                        }
	                      },
	                      "PositionalConstraint": "CONTAINS",
	                      "SearchString": "badBot",
	                      "TextTransformations": [
	                        {
	                          "Priority": 0,
	                          "Type": "NONE"
	                        }
	                      ]
	                    }
	                },
	                {
	                  "ByteMatchStatement":
	                    {
	                      "FieldToMatch": {
	                        "SingleHeader": {
	                          "Name": "Host"
	                        }
	                      },
	                      "PositionalConstraint": "CONTAINS",
	                      "SearchString": "badBot",
	                      "TextTransformations": [
	                        {
	                          "Priority": 1,
	                          "Type": "NONE"
	                        }
	                      ]
	                    }
	                }
	              ]
	            }`
	ruleAndStatement1 := ` {
	              "Statements": [
	                {
	                  "ByteMatchStatement":
	                    {
	                      "FieldToMatch": {
	                        "SingleHeader": {
	                          "Name": "User-Agent"
	                        }
	                      },
	                      "PositionalConstraint": "CONTAINS",
	                      "SearchString": "badBot",
	                      "TextTransformations": [
	                        {
	                          "Priority": 0,
	                          "Type": "NONE"
	                        }
	                      ]
	                    }
	                },
	                {
	                  "ByteMatchStatement":
	                    {
	                      "FieldToMatch": {
	                        "SingleHeader": {
	                          "Name": "Host"
	                        }
	                      },
	                      "PositionalConstraint": "CONTAINS",
	                      "SearchString": "crossplane.io",
	                      "TextTransformations": [
	                        {
	                          "Priority": 1,
	                          "Type": "NONE"
	                        }
	                      ]
	                    }
	                }
	              ]
	            }`

	ruleAndStatement0FieldToMatchSingleHeaderName := "user-agent"
	ruleAndStatement0PositionalConstraint := "CONTAINS"
	ruleAndStatement0SearchString := []byte("badBot")
	ruleAndStatement0TextTransformations0Priority := int64(0)
	ruleAndStatement0TextTransformations1Priority := int64(1)
	ruleAndStatement0TextTransformations0Type := svcsdk.TextTransformationTypeNone
	ruleAndStatement0TextTransformations1Type := ruleAndStatement0TextTransformations0Type

	ruleAndStatement1FieldToMatchSingleHeaderName := "host"
	ruleAndStatement1PositionalConstraint := ruleAndStatement0PositionalConstraint
	ruleAndStatement1SearchString := []byte("crossplane.io")
	ruleAndStatement1TextTransformations0Priority := int64(0)
	ruleAndStatement1TextTransformations1Priority := int64(1)
	ruleAndStatement1TextTransformations0Type := ruleAndStatement0TextTransformations0Type
	ruleAndStatement1TextTransformations1Type := ruleAndStatement1TextTransformations0Type

	rulesJSON := `[
                    {
                      "Action": {
                        "Allow": {}
                      },
                      "Name": "ruleName0",
                      "Priority": 0,
                      "Statement": {
                        "AndStatement": {
                          "Statements": [
                            {
                              "ByteMatchStatement": {
                                "FieldToMatch": {
                                  "SingleHeader": {
                                    "Name": "User-Agent"
                                  }
                                },
                                "PositionalConstraint": "CONTAINS",
                                "SearchString": "badBot",
                                "TextTransformations": [
                                  {
                                    "Priority": 0,
                                    "Type": "NONE"
                                  }
                                ]
                              }
                            },
                            {
                              "ByteMatchStatement": {
                                "FieldToMatch": {
                                  "SingleHeader": {
                                    "Name": "Host"
                                  }
                                },
                                "PositionalConstraint": "CONTAINS",
                                "SearchString": "badBot",
                                "TextTransformations": [
                                  {
                                    "Priority": 1,
                                    "Type": "NONE"
                                  }
                                ]
                              }
                            }
                          ]
                        }
                      },
                      "VisibilityConfig": {
                        "CloudWatchMetricsEnabled": false,
                        "MetricName": "metricName",
                        "SampledRequestsEnabled": false
                      }
                    },
                    {
                      "Action": {
                        "Allow": {}
                      },
                      "Name": "ruleName1",
                      "Priority": 1,
                      "Statement": {
                        "AndStatement": {
                          "Statements": [
                            {
                              "ByteMatchStatement": {
                                "FieldToMatch": {
                                  "SingleHeader": {
                                    "Name": "User-Agent"
                                  }
                                },
                                "PositionalConstraint": "CONTAINS",
                                "SearchString": "badBot",
                                "TextTransformations": [
                                  {
                                    "Priority": 0,
                                    "Type": "NONE"
                                  }
                                ]
                              }
                            },
                            {
                              "ByteMatchStatement": {
                                "FieldToMatch": {
                                  "SingleHeader": {
                                    "Name": "Host"
                                  }
                                },
                                "PositionalConstraint": "CONTAINS",
                                "SearchString": "crossplane.io",
                                "TextTransformations": [
                                  {
                                    "Priority": 1,
                                    "Type": "NONE"
                                  }
                                ]
                              }
                            }
                          ]
                        }
                      },
                      "VisibilityConfig": {
                        "CloudWatchMetricsEnabled": false,
                        "MetricName": "metricName",
                        "SampledRequestsEnabled": false
                      }
                    }
                  ]`
	rulesJSON2 := `[
                     {
                       "Action": {
                         "Allow": {}
                       },
                       "Name": "ruleName0",
                       "Priority": 0,
                       "Statement": {
                         "ByteMatchStatement": {
                           "FieldToMatch": {
                             "SingleHeader": {
                               "Name": "user-agent"
                             },
                             "UriPath": null
                           },
                           "PositionalConstraint": "CONTAINS",
                           "SearchString": "badBot",
                           "TextTransformations": [
                             {
                               "Priority": 1,
                               "Type": "NONE"
                             }
                           ]
                         }
                       },
                       "VisibilityConfig": {
                         "CloudWatchMetricsEnabled": false,
                         "MetricName": "metricName",
                         "SampledRequestsEnabled": false
                       }
                     }
                   ]`

	type want struct {
		result bool
		err    error
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"Same": {
			args: args{
				client: &fake.MockWAFV2Client{
					MockListTagsForResource: func(input *svcsdk.ListTagsForResourceInput) (*svcsdk.ListTagsForResourceOutput, error) {
						return &svcsdk.ListTagsForResourceOutput{
							TagInfoForResource: &svcsdk.TagInfoForResource{TagList: []*svcsdk.Tag{
								{Key: &tag0Key, Value: &tag0Value},
							}},
						}, nil
					},
					MockListResourcesForWebACL: func(input *svcsdk.ListResourcesForWebACLInput) (*svcsdk.ListResourcesForWebACLOutput, error) {
						return &svcsdk.ListResourcesForWebACLOutput{
							ResourceArns: nil,
						}, nil
					},
				},
				desired: &svcapitypes.WebACL{
					ObjectMeta: metav1.ObjectMeta{
						Name: webAclName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: webAclName,
						},
					},
					Spec: svcapitypes.WebACLSpec{
						ForProvider: svcapitypes.WebACLParameters{
							Region: "eu-central-1",
							VisibilityConfig: &svcapitypes.VisibilityConfig{
								MetricName:               &visibilityConfigMetricName,
								SampledRequestsEnabled:   &visibilityConfigSampledRequestsEnabled,
								CloudWatchMetricsEnabled: &visibilityConfigCloudWatchMetricsEnabled,
							},
							DefaultAction: &svcapitypes.DefaultAction{
								Allow: &svcapitypes.AllowAction{},
							},
							Scope: &scope,
							Rules: []*svcapitypes.Rule{
								{
									Name: &ruleName0,
									VisibilityConfig: &svcapitypes.VisibilityConfig{
										MetricName:               &visibilityConfigMetricName,
										SampledRequestsEnabled:   &visibilityConfigSampledRequestsEnabled,
										CloudWatchMetricsEnabled: &visibilityConfigCloudWatchMetricsEnabled,
									},
									Priority: &rulePriority0,
									Action: &svcapitypes.RuleAction{
										Allow: &svcapitypes.AllowAction{},
									},
									Statement: &svcapitypes.Statement{
										AndStatement: &ruleAndStatement0,
									},
								},
								{
									Name: &ruleName1,
									VisibilityConfig: &svcapitypes.VisibilityConfig{
										MetricName:               &visibilityConfigMetricName,
										SampledRequestsEnabled:   &visibilityConfigSampledRequestsEnabled,
										CloudWatchMetricsEnabled: &visibilityConfigCloudWatchMetricsEnabled,
									},
									Priority: &rulePriority1,
									Action: &svcapitypes.RuleAction{
										Allow: &svcapitypes.AllowAction{},
									},
									Statement: &svcapitypes.Statement{
										AndStatement: &ruleAndStatement1,
									},
								},
							},
							Tags: []*svcapitypes.Tag{
								{Key: &tag0Key, Value: &tag0Value},
							},
						},
					},
				},
				observed: &svcsdk.GetWebACLOutput{
					WebACL: &svcsdk.WebACL{
						Name:        &webAclName,
						Id:          &webAclId,
						Description: aws.String(""),
						VisibilityConfig: &svcsdk.VisibilityConfig{
							MetricName:               &visibilityConfigMetricName,
							SampledRequestsEnabled:   &visibilityConfigSampledRequestsEnabled,
							CloudWatchMetricsEnabled: &visibilityConfigCloudWatchMetricsEnabled,
						},
						DefaultAction: &svcsdk.DefaultAction{
							Allow: &svcsdk.AllowAction{},
						},
						Rules: []*svcsdk.Rule{
							{
								Name: &ruleName1,
								VisibilityConfig: &svcsdk.VisibilityConfig{
									MetricName:               &visibilityConfigMetricName,
									SampledRequestsEnabled:   &visibilityConfigSampledRequestsEnabled,
									CloudWatchMetricsEnabled: &visibilityConfigCloudWatchMetricsEnabled,
								},
								Priority: &rulePriority1,
								Action: &svcsdk.RuleAction{
									Allow: &svcsdk.AllowAction{},
								},
								Statement: &svcsdk.Statement{
									AndStatement: &svcsdk.AndStatement{
										Statements: []*svcsdk.Statement{
											{ByteMatchStatement: &svcsdk.ByteMatchStatement{
												FieldToMatch: &svcsdk.FieldToMatch{
													SingleHeader: &svcsdk.SingleHeader{
														Name: &ruleAndStatement0FieldToMatchSingleHeaderName,
													},
												},
												PositionalConstraint: &ruleAndStatement0PositionalConstraint,
												SearchString:         ruleAndStatement0SearchString,
												TextTransformations: []*svcsdk.TextTransformation{
													{Priority: &ruleAndStatement0TextTransformations0Priority, Type: &ruleAndStatement0TextTransformations0Type},
												},
											},
											},
											{ByteMatchStatement: &svcsdk.ByteMatchStatement{
												FieldToMatch: &svcsdk.FieldToMatch{
													SingleHeader: &svcsdk.SingleHeader{
														Name: &ruleAndStatement1FieldToMatchSingleHeaderName,
													},
												},
												PositionalConstraint: &ruleAndStatement1PositionalConstraint,
												SearchString:         ruleAndStatement1SearchString,
												TextTransformations: []*svcsdk.TextTransformation{
													{Priority: &ruleAndStatement1TextTransformations1Priority, Type: &ruleAndStatement1TextTransformations1Type},
												},
											},
											},
										},
									},
								},
							},
							{
								Name: &ruleName0,
								VisibilityConfig: &svcsdk.VisibilityConfig{
									MetricName:               &visibilityConfigMetricName,
									SampledRequestsEnabled:   &visibilityConfigSampledRequestsEnabled,
									CloudWatchMetricsEnabled: &visibilityConfigCloudWatchMetricsEnabled,
								},
								Priority: &rulePriority0,
								Action: &svcsdk.RuleAction{
									Allow: &svcsdk.AllowAction{},
								},
								Statement: &svcsdk.Statement{
									AndStatement: &svcsdk.AndStatement{
										Statements: []*svcsdk.Statement{
											{ByteMatchStatement: &svcsdk.ByteMatchStatement{
												FieldToMatch: &svcsdk.FieldToMatch{
													SingleHeader: &svcsdk.SingleHeader{
														Name: &ruleAndStatement0FieldToMatchSingleHeaderName,
													},
												},
												PositionalConstraint: &ruleAndStatement0PositionalConstraint,
												SearchString:         ruleAndStatement0SearchString,
												TextTransformations: []*svcsdk.TextTransformation{
													{Priority: &ruleAndStatement0TextTransformations0Priority, Type: &ruleAndStatement0TextTransformations0Type},
												},
											},
											},
											{ByteMatchStatement: &svcsdk.ByteMatchStatement{
												FieldToMatch: &svcsdk.FieldToMatch{
													SingleHeader: &svcsdk.SingleHeader{
														Name: &ruleAndStatement1FieldToMatchSingleHeaderName,
													},
												},
												PositionalConstraint: &ruleAndStatement1PositionalConstraint,
												SearchString:         ruleAndStatement0SearchString,
												TextTransformations: []*svcsdk.TextTransformation{
													{Priority: &ruleAndStatement0TextTransformations1Priority, Type: &ruleAndStatement0TextTransformations1Type},
												},
											},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
		"SameRulesJSON": {
			args: args{
				client: &fake.MockWAFV2Client{
					MockListTagsForResource: func(input *svcsdk.ListTagsForResourceInput) (*svcsdk.ListTagsForResourceOutput, error) {
						return &svcsdk.ListTagsForResourceOutput{
							TagInfoForResource: &svcsdk.TagInfoForResource{TagList: []*svcsdk.Tag{
								{Key: &tag0Key, Value: &tag0Value},
							}},
						}, nil
					},
					MockListResourcesForWebACL: func(input *svcsdk.ListResourcesForWebACLInput) (*svcsdk.ListResourcesForWebACLOutput, error) {
						return &svcsdk.ListResourcesForWebACLOutput{
							ResourceArns: nil,
						}, nil
					},
				},
				desired: &svcapitypes.WebACL{
					ObjectMeta: metav1.ObjectMeta{
						Name: webAclName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: webAclName,
						},
					},
					Spec: svcapitypes.WebACLSpec{
						ForProvider: svcapitypes.WebACLParameters{
							Region: "eu-central-1",
							VisibilityConfig: &svcapitypes.VisibilityConfig{
								MetricName:               &visibilityConfigMetricName,
								SampledRequestsEnabled:   &visibilityConfigSampledRequestsEnabled,
								CloudWatchMetricsEnabled: &visibilityConfigCloudWatchMetricsEnabled,
							},
							DefaultAction: &svcapitypes.DefaultAction{
								Allow: &svcapitypes.AllowAction{},
							},
							Scope: &scope,
							CustomWebACLParameters: svcapitypes.CustomWebACLParameters{
								RulesJSON: aws.String(rulesJSON),
							},
							Tags: []*svcapitypes.Tag{
								{Key: &tag0Key, Value: &tag0Value},
							},
						},
					},
				},
				observed: &svcsdk.GetWebACLOutput{
					WebACL: &svcsdk.WebACL{
						Name:        &webAclName,
						Id:          &webAclId,
						Description: aws.String(""),
						VisibilityConfig: &svcsdk.VisibilityConfig{
							MetricName:               &visibilityConfigMetricName,
							SampledRequestsEnabled:   &visibilityConfigSampledRequestsEnabled,
							CloudWatchMetricsEnabled: &visibilityConfigCloudWatchMetricsEnabled,
						},
						DefaultAction: &svcsdk.DefaultAction{
							Allow: &svcsdk.AllowAction{},
						},
						Rules: []*svcsdk.Rule{
							{
								Name: &ruleName1,
								VisibilityConfig: &svcsdk.VisibilityConfig{
									MetricName:               &visibilityConfigMetricName,
									SampledRequestsEnabled:   &visibilityConfigSampledRequestsEnabled,
									CloudWatchMetricsEnabled: &visibilityConfigCloudWatchMetricsEnabled,
								},
								Priority: &rulePriority1,
								Action: &svcsdk.RuleAction{
									Allow: &svcsdk.AllowAction{},
								},
								Statement: &svcsdk.Statement{
									AndStatement: &svcsdk.AndStatement{
										Statements: []*svcsdk.Statement{
											{ByteMatchStatement: &svcsdk.ByteMatchStatement{
												FieldToMatch: &svcsdk.FieldToMatch{
													SingleHeader: &svcsdk.SingleHeader{
														Name: &ruleAndStatement0FieldToMatchSingleHeaderName,
													},
												},
												PositionalConstraint: &ruleAndStatement0PositionalConstraint,
												SearchString:         ruleAndStatement0SearchString,
												TextTransformations: []*svcsdk.TextTransformation{
													{Priority: &ruleAndStatement0TextTransformations0Priority, Type: &ruleAndStatement0TextTransformations0Type},
												},
											},
											},
											{ByteMatchStatement: &svcsdk.ByteMatchStatement{
												FieldToMatch: &svcsdk.FieldToMatch{
													SingleHeader: &svcsdk.SingleHeader{
														Name: &ruleAndStatement1FieldToMatchSingleHeaderName,
													},
												},
												PositionalConstraint: &ruleAndStatement1PositionalConstraint,
												SearchString:         ruleAndStatement1SearchString,
												TextTransformations: []*svcsdk.TextTransformation{
													{Priority: &ruleAndStatement1TextTransformations1Priority, Type: &ruleAndStatement1TextTransformations1Type},
												},
											},
											},
										},
									},
								},
							},
							{
								Name: &ruleName0,
								VisibilityConfig: &svcsdk.VisibilityConfig{
									MetricName:               &visibilityConfigMetricName,
									SampledRequestsEnabled:   &visibilityConfigSampledRequestsEnabled,
									CloudWatchMetricsEnabled: &visibilityConfigCloudWatchMetricsEnabled,
								},
								Priority: &rulePriority0,
								Action: &svcsdk.RuleAction{
									Allow: &svcsdk.AllowAction{},
								},
								Statement: &svcsdk.Statement{
									AndStatement: &svcsdk.AndStatement{
										Statements: []*svcsdk.Statement{
											{ByteMatchStatement: &svcsdk.ByteMatchStatement{
												FieldToMatch: &svcsdk.FieldToMatch{
													SingleHeader: &svcsdk.SingleHeader{
														Name: &ruleAndStatement0FieldToMatchSingleHeaderName,
													},
												},
												PositionalConstraint: &ruleAndStatement0PositionalConstraint,
												SearchString:         ruleAndStatement0SearchString,
												TextTransformations: []*svcsdk.TextTransformation{
													{Priority: &ruleAndStatement0TextTransformations0Priority, Type: &ruleAndStatement0TextTransformations0Type},
												},
											},
											},
											{ByteMatchStatement: &svcsdk.ByteMatchStatement{
												FieldToMatch: &svcsdk.FieldToMatch{
													SingleHeader: &svcsdk.SingleHeader{
														Name: &ruleAndStatement1FieldToMatchSingleHeaderName,
													},
												},
												PositionalConstraint: &ruleAndStatement1PositionalConstraint,
												SearchString:         ruleAndStatement0SearchString,
												TextTransformations: []*svcsdk.TextTransformation{
													{Priority: &ruleAndStatement0TextTransformations1Priority, Type: &ruleAndStatement0TextTransformations1Type},
												},
											},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
		"SameRulesJSON2": {
			args: args{
				client: &fake.MockWAFV2Client{
					MockListTagsForResource: func(input *svcsdk.ListTagsForResourceInput) (*svcsdk.ListTagsForResourceOutput, error) {
						return &svcsdk.ListTagsForResourceOutput{
							TagInfoForResource: &svcsdk.TagInfoForResource{TagList: []*svcsdk.Tag{}},
						}, nil
					},
					MockListResourcesForWebACL: func(input *svcsdk.ListResourcesForWebACLInput) (*svcsdk.ListResourcesForWebACLOutput, error) {
						return &svcsdk.ListResourcesForWebACLOutput{
							ResourceArns: nil,
						}, nil
					},
				},
				desired: &svcapitypes.WebACL{
					ObjectMeta: metav1.ObjectMeta{
						Name: webAclName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: webAclName,
						},
					},
					Spec: svcapitypes.WebACLSpec{
						ForProvider: svcapitypes.WebACLParameters{
							Region: "eu-central-1",
							VisibilityConfig: &svcapitypes.VisibilityConfig{
								MetricName:               &visibilityConfigMetricName,
								SampledRequestsEnabled:   &visibilityConfigSampledRequestsEnabled,
								CloudWatchMetricsEnabled: &visibilityConfigCloudWatchMetricsEnabled,
							},
							DefaultAction: &svcapitypes.DefaultAction{
								Allow: &svcapitypes.AllowAction{},
							},
							Scope: &scope,
							CustomWebACLParameters: svcapitypes.CustomWebACLParameters{
								RulesJSON: aws.String(rulesJSON2),
							},
						},
					},
				},
				observed: &svcsdk.GetWebACLOutput{
					WebACL: &svcsdk.WebACL{
						Name:        &webAclName,
						Description: aws.String(""),
						Id:          &webAclId,
						VisibilityConfig: &svcsdk.VisibilityConfig{
							MetricName:               &visibilityConfigMetricName,
							SampledRequestsEnabled:   &visibilityConfigSampledRequestsEnabled,
							CloudWatchMetricsEnabled: &visibilityConfigCloudWatchMetricsEnabled,
						},
						DefaultAction: &svcsdk.DefaultAction{
							Allow: &svcsdk.AllowAction{},
						},
						Rules: []*svcsdk.Rule{
							{
								Name: &ruleName0,
								VisibilityConfig: &svcsdk.VisibilityConfig{
									MetricName:               &visibilityConfigMetricName,
									SampledRequestsEnabled:   &visibilityConfigSampledRequestsEnabled,
									CloudWatchMetricsEnabled: &visibilityConfigCloudWatchMetricsEnabled,
								},
								Priority: &rulePriority0,
								Action: &svcsdk.RuleAction{
									Allow: &svcsdk.AllowAction{},
								},
								Statement: &svcsdk.Statement{
									ByteMatchStatement: &svcsdk.ByteMatchStatement{
										FieldToMatch: &svcsdk.FieldToMatch{
											SingleHeader: &svcsdk.SingleHeader{
												Name: aws.String("user-agent"),
											},
										},
										PositionalConstraint: aws.String("CONTAINS"),
										SearchString:         []byte("badBot"),
										TextTransformations: []*svcsdk.TextTransformation{
											{Priority: aws.Int64(1), Type: aws.String("NONE")},
										},
									},
								},
							},
						},
					},
				},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
		"ChangedRulesJSON2": {
			args: args{
				client: &fake.MockWAFV2Client{
					MockListTagsForResource: func(input *svcsdk.ListTagsForResourceInput) (*svcsdk.ListTagsForResourceOutput, error) {
						return &svcsdk.ListTagsForResourceOutput{
							TagInfoForResource: &svcsdk.TagInfoForResource{TagList: []*svcsdk.Tag{}},
						}, nil
					},
					MockListResourcesForWebACL: func(input *svcsdk.ListResourcesForWebACLInput) (*svcsdk.ListResourcesForWebACLOutput, error) {
						return &svcsdk.ListResourcesForWebACLOutput{
							ResourceArns: nil,
						}, nil
					},
				},
				desired: &svcapitypes.WebACL{
					ObjectMeta: metav1.ObjectMeta{
						Name: webAclName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: webAclName,
						},
					},
					Spec: svcapitypes.WebACLSpec{
						ForProvider: svcapitypes.WebACLParameters{
							Region: "eu-central-1",
							VisibilityConfig: &svcapitypes.VisibilityConfig{
								MetricName:               &visibilityConfigMetricName,
								SampledRequestsEnabled:   &visibilityConfigSampledRequestsEnabled,
								CloudWatchMetricsEnabled: &visibilityConfigCloudWatchMetricsEnabled,
							},
							DefaultAction: &svcapitypes.DefaultAction{
								Allow: &svcapitypes.AllowAction{},
							},
							Scope: &scope,
							CustomWebACLParameters: svcapitypes.CustomWebACLParameters{
								RulesJSON: aws.String(rulesJSON2),
							},
						},
					},
				},
				observed: &svcsdk.GetWebACLOutput{
					WebACL: &svcsdk.WebACL{
						Name:        &webAclName,
						Description: aws.String(""),
						Id:          &webAclId,
						VisibilityConfig: &svcsdk.VisibilityConfig{
							MetricName:               &visibilityConfigMetricName,
							SampledRequestsEnabled:   &visibilityConfigSampledRequestsEnabled,
							CloudWatchMetricsEnabled: &visibilityConfigCloudWatchMetricsEnabled,
						},
						DefaultAction: &svcsdk.DefaultAction{
							Allow: &svcsdk.AllowAction{},
						},
						Rules: []*svcsdk.Rule{
							{
								Name: &ruleName0,
								VisibilityConfig: &svcsdk.VisibilityConfig{
									MetricName:               &visibilityConfigMetricName,
									SampledRequestsEnabled:   &visibilityConfigSampledRequestsEnabled,
									CloudWatchMetricsEnabled: &visibilityConfigCloudWatchMetricsEnabled,
								},
								Priority: &rulePriority1,
								Action: &svcsdk.RuleAction{
									Allow: &svcsdk.AllowAction{},
								},
								Statement: &svcsdk.Statement{
									ByteMatchStatement: &svcsdk.ByteMatchStatement{
										FieldToMatch: &svcsdk.FieldToMatch{
											SingleHeader: &svcsdk.SingleHeader{
												Name: aws.String("user-agent"),
											},
										},
										PositionalConstraint: aws.String("CONTAINS"),
										SearchString:         []byte("crossplane.io"),
										TextTransformations: []*svcsdk.TextTransformation{
											{Priority: aws.Int64(1), Type: aws.String("NONE")},
										},
									},
								},
							},
						},
					},
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"JsonifiedRulesStatementChanged": {
			args: args{
				client: &fake.MockWAFV2Client{
					MockListTagsForResource: func(input *svcsdk.ListTagsForResourceInput) (*svcsdk.ListTagsForResourceOutput, error) {
						return &svcsdk.ListTagsForResourceOutput{
							TagInfoForResource: &svcsdk.TagInfoForResource{TagList: []*svcsdk.Tag{
								{Key: &tag0Key, Value: &tag0Value},
							}},
						}, nil
					},
					MockListResourcesForWebACL: func(input *svcsdk.ListResourcesForWebACLInput) (*svcsdk.ListResourcesForWebACLOutput, error) {
						return &svcsdk.ListResourcesForWebACLOutput{
							ResourceArns: nil,
						}, nil
					},
				},
				desired: &svcapitypes.WebACL{
					ObjectMeta: metav1.ObjectMeta{
						Name: webAclName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: webAclName,
						},
					},
					Spec: svcapitypes.WebACLSpec{
						ForProvider: svcapitypes.WebACLParameters{
							Region: "eu-central-1",
							VisibilityConfig: &svcapitypes.VisibilityConfig{
								MetricName:               &visibilityConfigMetricName,
								SampledRequestsEnabled:   &visibilityConfigSampledRequestsEnabled,
								CloudWatchMetricsEnabled: &visibilityConfigCloudWatchMetricsEnabled,
							},
							DefaultAction: &svcapitypes.DefaultAction{
								Allow: &svcapitypes.AllowAction{},
							},
							Scope: &scope,
							Rules: []*svcapitypes.Rule{
								{
									Name: &ruleName0,
									VisibilityConfig: &svcapitypes.VisibilityConfig{
										MetricName:               &visibilityConfigMetricName,
										SampledRequestsEnabled:   &visibilityConfigSampledRequestsEnabled,
										CloudWatchMetricsEnabled: &visibilityConfigCloudWatchMetricsEnabled,
									},
									Priority: &rulePriority0,
									Action: &svcapitypes.RuleAction{
										Allow: &svcapitypes.AllowAction{},
									},
									Statement: &svcapitypes.Statement{
										AndStatement: &ruleAndStatement1,
									},
								},
							},
							Tags: []*svcapitypes.Tag{
								{Key: &tag0Key, Value: &tag0Value},
							},
						},
					},
				},
				observed: &svcsdk.GetWebACLOutput{
					WebACL: &svcsdk.WebACL{
						Name:        &webAclName,
						Description: aws.String(""),
						Id:          &webAclId,
						VisibilityConfig: &svcsdk.VisibilityConfig{
							MetricName:               &visibilityConfigMetricName,
							SampledRequestsEnabled:   &visibilityConfigSampledRequestsEnabled,
							CloudWatchMetricsEnabled: &visibilityConfigCloudWatchMetricsEnabled,
						},
						DefaultAction: &svcsdk.DefaultAction{
							Allow: &svcsdk.AllowAction{},
						},
						Rules: []*svcsdk.Rule{
							{
								Name: &ruleName0,
								VisibilityConfig: &svcsdk.VisibilityConfig{
									MetricName:               &visibilityConfigMetricName,
									SampledRequestsEnabled:   &visibilityConfigSampledRequestsEnabled,
									CloudWatchMetricsEnabled: &visibilityConfigCloudWatchMetricsEnabled,
								},
								Priority: &rulePriority0,
								Action: &svcsdk.RuleAction{
									Allow: &svcsdk.AllowAction{},
								},
								Statement: &svcsdk.Statement{
									AndStatement: &svcsdk.AndStatement{
										Statements: []*svcsdk.Statement{
											{ByteMatchStatement: &svcsdk.ByteMatchStatement{
												FieldToMatch: &svcsdk.FieldToMatch{
													SingleHeader: &svcsdk.SingleHeader{
														Name: &ruleAndStatement0FieldToMatchSingleHeaderName,
													},
												},
												PositionalConstraint: &ruleAndStatement0PositionalConstraint,
												SearchString:         ruleAndStatement0SearchString,
												TextTransformations: []*svcsdk.TextTransformation{
													{Priority: &ruleAndStatement0TextTransformations0Priority, Type: &ruleAndStatement0TextTransformations0Type},
												},
											},
											},
											{ByteMatchStatement: &svcsdk.ByteMatchStatement{
												FieldToMatch: &svcsdk.FieldToMatch{
													SingleHeader: &svcsdk.SingleHeader{
														Name: &ruleAndStatement1FieldToMatchSingleHeaderName,
													},
												},
												PositionalConstraint: &ruleAndStatement1PositionalConstraint,
												SearchString:         ruleAndStatement1SearchString,
												TextTransformations: []*svcsdk.TextTransformation{
													{Priority: &ruleAndStatement1TextTransformations0Priority, Type: &ruleAndStatement1TextTransformations0Type},
												},
											},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"NormalStatementIsNotChanged": {
			args: args{
				client: &fake.MockWAFV2Client{
					MockListTagsForResource: func(input *svcsdk.ListTagsForResourceInput) (*svcsdk.ListTagsForResourceOutput, error) {
						return &svcsdk.ListTagsForResourceOutput{
							TagInfoForResource: &svcsdk.TagInfoForResource{TagList: []*svcsdk.Tag{}},
						}, nil
					},
					MockListResourcesForWebACL: func(input *svcsdk.ListResourcesForWebACLInput) (*svcsdk.ListResourcesForWebACLOutput, error) {
						return &svcsdk.ListResourcesForWebACLOutput{
							ResourceArns: nil,
						}, nil
					},
				},
				desired: &svcapitypes.WebACL{
					ObjectMeta: metav1.ObjectMeta{
						Name: webAclName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: webAclName,
						},
					},
					Spec: svcapitypes.WebACLSpec{
						ForProvider: svcapitypes.WebACLParameters{
							Region: "eu-central-1",
							VisibilityConfig: &svcapitypes.VisibilityConfig{
								MetricName:               &visibilityConfigMetricName,
								SampledRequestsEnabled:   &visibilityConfigSampledRequestsEnabled,
								CloudWatchMetricsEnabled: &visibilityConfigCloudWatchMetricsEnabled,
							},
							DefaultAction: &svcapitypes.DefaultAction{
								Allow: &svcapitypes.AllowAction{},
							},
							Scope: &scope,
							Rules: []*svcapitypes.Rule{
								{
									Name: &ruleName0,
									VisibilityConfig: &svcapitypes.VisibilityConfig{
										MetricName:               &visibilityConfigMetricName,
										SampledRequestsEnabled:   &visibilityConfigSampledRequestsEnabled,
										CloudWatchMetricsEnabled: &visibilityConfigCloudWatchMetricsEnabled,
									},
									Priority: &rulePriority0,
									Action: &svcapitypes.RuleAction{
										Allow: &svcapitypes.AllowAction{},
									},
									Statement: &svcapitypes.Statement{
										ByteMatchStatement: &svcapitypes.ByteMatchStatement{
											FieldToMatch: &svcapitypes.FieldToMatch{
												SingleHeader: &svcapitypes.SingleHeader{
													Name: aws.String("User-Agent"),
												},
											},
											PositionalConstraint: aws.String("CONTAINS"),
											SearchString:         aws.String("badBot"),
											TextTransformations: []*svcapitypes.TextTransformation{
												{Priority: aws.Int64(1), Type: aws.String("NONE")},
											},
										},
									},
								},
							},
						},
					},
				},
				observed: &svcsdk.GetWebACLOutput{
					WebACL: &svcsdk.WebACL{
						Name:        &webAclName,
						Description: aws.String(""),
						Id:          &webAclId,
						VisibilityConfig: &svcsdk.VisibilityConfig{
							MetricName:               &visibilityConfigMetricName,
							SampledRequestsEnabled:   &visibilityConfigSampledRequestsEnabled,
							CloudWatchMetricsEnabled: &visibilityConfigCloudWatchMetricsEnabled,
						},
						DefaultAction: &svcsdk.DefaultAction{
							Allow: &svcsdk.AllowAction{},
						},
						Rules: []*svcsdk.Rule{
							{
								Name: &ruleName0,
								VisibilityConfig: &svcsdk.VisibilityConfig{
									MetricName:               &visibilityConfigMetricName,
									SampledRequestsEnabled:   &visibilityConfigSampledRequestsEnabled,
									CloudWatchMetricsEnabled: &visibilityConfigCloudWatchMetricsEnabled,
								},
								Priority: &rulePriority0,
								Action: &svcsdk.RuleAction{
									Allow: &svcsdk.AllowAction{},
								},
								Statement: &svcsdk.Statement{
									ByteMatchStatement: &svcsdk.ByteMatchStatement{
										FieldToMatch: &svcsdk.FieldToMatch{
											SingleHeader: &svcsdk.SingleHeader{
												Name: aws.String("user-agent"),
											},
										},
										PositionalConstraint: aws.String("CONTAINS"),
										SearchString:         []byte("badBot"),
										TextTransformations: []*svcsdk.TextTransformation{
											{Priority: aws.Int64(1), Type: aws.String("NONE")},
										},
									},
								},
							},
						},
					},
				},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
		"NormalStatementChanged": {
			args: args{
				client: &fake.MockWAFV2Client{
					MockListTagsForResource: func(input *svcsdk.ListTagsForResourceInput) (*svcsdk.ListTagsForResourceOutput, error) {
						return &svcsdk.ListTagsForResourceOutput{
							TagInfoForResource: &svcsdk.TagInfoForResource{TagList: []*svcsdk.Tag{}},
						}, nil
					},
					MockListResourcesForWebACL: func(input *svcsdk.ListResourcesForWebACLInput) (*svcsdk.ListResourcesForWebACLOutput, error) {
						return &svcsdk.ListResourcesForWebACLOutput{
							ResourceArns: nil,
						}, nil
					},
				},
				desired: &svcapitypes.WebACL{
					ObjectMeta: metav1.ObjectMeta{
						Name: webAclName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: webAclName,
						},
					},
					Spec: svcapitypes.WebACLSpec{
						ForProvider: svcapitypes.WebACLParameters{
							Region: "eu-central-1",
							VisibilityConfig: &svcapitypes.VisibilityConfig{
								MetricName:               &visibilityConfigMetricName,
								SampledRequestsEnabled:   &visibilityConfigSampledRequestsEnabled,
								CloudWatchMetricsEnabled: &visibilityConfigCloudWatchMetricsEnabled,
							},
							DefaultAction: &svcapitypes.DefaultAction{
								Allow: &svcapitypes.AllowAction{},
							},
							Scope: &scope,
							Rules: []*svcapitypes.Rule{
								{
									Name: &ruleName0,
									VisibilityConfig: &svcapitypes.VisibilityConfig{
										MetricName:               &visibilityConfigMetricName,
										SampledRequestsEnabled:   &visibilityConfigSampledRequestsEnabled,
										CloudWatchMetricsEnabled: &visibilityConfigCloudWatchMetricsEnabled,
									},
									Priority: &rulePriority0,
									Action: &svcapitypes.RuleAction{
										Allow: &svcapitypes.AllowAction{},
									},
									Statement: &svcapitypes.Statement{
										ManagedRuleGroupStatement: &svcapitypes.ManagedRuleGroupStatement{
											Name:       aws.String("AWSManagedRulesCommonRuleSet"),
											VendorName: aws.String("AWS"),
											RuleActionOverrides: []*svcapitypes.RuleActionOverride{
												{
													Name: aws.String("SizeRestrictions_BODY"),
													ActionToUse: &svcapitypes.RuleAction{
														Count: &svcapitypes.CountAction{},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				observed: &svcsdk.GetWebACLOutput{
					WebACL: &svcsdk.WebACL{
						Name:        &webAclName,
						Description: aws.String(""),
						Id:          &webAclId,
						VisibilityConfig: &svcsdk.VisibilityConfig{
							MetricName:               &visibilityConfigMetricName,
							SampledRequestsEnabled:   &visibilityConfigSampledRequestsEnabled,
							CloudWatchMetricsEnabled: &visibilityConfigCloudWatchMetricsEnabled,
						},
						DefaultAction: &svcsdk.DefaultAction{
							Allow: &svcsdk.AllowAction{},
						},
						Rules: []*svcsdk.Rule{
							{
								Name: &ruleName0,
								VisibilityConfig: &svcsdk.VisibilityConfig{
									MetricName:               &visibilityConfigMetricName,
									SampledRequestsEnabled:   &visibilityConfigSampledRequestsEnabled,
									CloudWatchMetricsEnabled: &visibilityConfigCloudWatchMetricsEnabled,
								},
								Priority: &rulePriority0,
								Action: &svcsdk.RuleAction{
									Allow: &svcsdk.AllowAction{},
								},
								Statement: &svcsdk.Statement{
									ManagedRuleGroupStatement: &svcsdk.ManagedRuleGroupStatement{
										Name:       aws.String("AWSManagedRulesCommonRuleSet"),
										VendorName: aws.String("AWS"),
										RuleActionOverrides: []*svcsdk.RuleActionOverride{
											{
												Name:        aws.String("SizeRestrictions_BODY"),
												ActionToUse: &svcsdk.RuleAction{Count: &svcsdk.CountAction{}},
											},
											{
												Name:        aws.String("SizeRestrictions_URIPATH"),
												ActionToUse: &svcsdk.RuleAction{Count: &svcsdk.CountAction{}},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"TagsAreChanged": {
			args: args{
				client: &fake.MockWAFV2Client{
					MockListTagsForResource: func(input *svcsdk.ListTagsForResourceInput) (*svcsdk.ListTagsForResourceOutput, error) {
						return &svcsdk.ListTagsForResourceOutput{
							TagInfoForResource: &svcsdk.TagInfoForResource{TagList: []*svcsdk.Tag{
								{Key: &tag0Key, Value: &tag0Value},
							}},
						}, nil
					},
					MockListResourcesForWebACL: func(input *svcsdk.ListResourcesForWebACLInput) (*svcsdk.ListResourcesForWebACLOutput, error) {
						return &svcsdk.ListResourcesForWebACLOutput{
							ResourceArns: nil,
						}, nil
					},
				},
				desired: &svcapitypes.WebACL{
					ObjectMeta: metav1.ObjectMeta{
						Name: webAclName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: webAclName,
						},
					},
					Spec: svcapitypes.WebACLSpec{
						ForProvider: svcapitypes.WebACLParameters{
							Region: "eu-central-1",
							VisibilityConfig: &svcapitypes.VisibilityConfig{
								MetricName:               &visibilityConfigMetricName,
								SampledRequestsEnabled:   &visibilityConfigSampledRequestsEnabled,
								CloudWatchMetricsEnabled: &visibilityConfigCloudWatchMetricsEnabled,
							},
							DefaultAction: &svcapitypes.DefaultAction{
								Allow: &svcapitypes.AllowAction{},
							},
							Scope: &scope,
							Tags: []*svcapitypes.Tag{
								{Key: &tag0Key, Value: &tag0NewValue},
							},
						},
					},
				},
				observed: &svcsdk.GetWebACLOutput{
					WebACL: &svcsdk.WebACL{
						Name: &webAclName,
						Id:   &webAclId,
						VisibilityConfig: &svcsdk.VisibilityConfig{
							MetricName:               &visibilityConfigMetricName,
							SampledRequestsEnabled:   &visibilityConfigSampledRequestsEnabled,
							CloudWatchMetricsEnabled: &visibilityConfigCloudWatchMetricsEnabled,
						},
						DefaultAction: &svcsdk.DefaultAction{
							Allow: &svcsdk.AllowAction{},
						},
					},
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"TagsAreNotChanged": {
			args: args{
				client: &fake.MockWAFV2Client{
					MockListTagsForResource: func(input *svcsdk.ListTagsForResourceInput) (*svcsdk.ListTagsForResourceOutput, error) {
						return &svcsdk.ListTagsForResourceOutput{
							TagInfoForResource: &svcsdk.TagInfoForResource{TagList: []*svcsdk.Tag{
								{Key: &tag0Key, Value: &tag0Value},
								{Key: &tag1Key, Value: &tag1Value},
							}},
						}, nil
					},
					MockListResourcesForWebACL: func(input *svcsdk.ListResourcesForWebACLInput) (*svcsdk.ListResourcesForWebACLOutput, error) {
						return &svcsdk.ListResourcesForWebACLOutput{
							ResourceArns: nil,
						}, nil
					},
				},
				desired: &svcapitypes.WebACL{
					ObjectMeta: metav1.ObjectMeta{
						Name: webAclName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: webAclName,
						},
					},
					Spec: svcapitypes.WebACLSpec{
						ForProvider: svcapitypes.WebACLParameters{
							Region: "eu-central-1",
							VisibilityConfig: &svcapitypes.VisibilityConfig{
								MetricName:               &visibilityConfigMetricName,
								SampledRequestsEnabled:   &visibilityConfigSampledRequestsEnabled,
								CloudWatchMetricsEnabled: &visibilityConfigCloudWatchMetricsEnabled,
							},
							DefaultAction: &svcapitypes.DefaultAction{
								Allow: &svcapitypes.AllowAction{},
							},
							Scope: &scope,
							Tags: []*svcapitypes.Tag{
								{Key: &tag1Key, Value: &tag1Value},
								{Key: &tag0Key, Value: &tag0Value},
							},
						},
					},
				},
				observed: &svcsdk.GetWebACLOutput{
					WebACL: &svcsdk.WebACL{
						Name:        &webAclName,
						Description: aws.String(""),
						Id:          &webAclId,
						VisibilityConfig: &svcsdk.VisibilityConfig{
							MetricName:               &visibilityConfigMetricName,
							SampledRequestsEnabled:   &visibilityConfigSampledRequestsEnabled,
							CloudWatchMetricsEnabled: &visibilityConfigCloudWatchMetricsEnabled,
						},
						DefaultAction: &svcsdk.DefaultAction{
							Allow: &svcsdk.AllowAction{},
						},
					},
				},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
		"AssociatedResourceWasChanged": {
			args: args{
				client: &fake.MockWAFV2Client{
					MockListTagsForResource: func(input *svcsdk.ListTagsForResourceInput) (*svcsdk.ListTagsForResourceOutput, error) {
						return &svcsdk.ListTagsForResourceOutput{
							TagInfoForResource: &svcsdk.TagInfoForResource{TagList: []*svcsdk.Tag{
								{Key: &tag0Key, Value: &tag0Value},
							}},
						}, nil
					},
					MockListResourcesForWebACL: func(input *svcsdk.ListResourcesForWebACLInput) (*svcsdk.ListResourcesForWebACLOutput, error) {
						return &svcsdk.ListResourcesForWebACLOutput{
							ResourceArns: []*string{aws.String("arn:aws:elasticloadbalancing:eu-central-1:123456789012:loadbalancer/app/my-load-balancer/50dc6c495c0c9188")},
						}, nil
					},
				},
				desired: &svcapitypes.WebACL{
					ObjectMeta: metav1.ObjectMeta{
						Name: webAclName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: webAclName,
						},
					},
					Spec: svcapitypes.WebACLSpec{
						ForProvider: svcapitypes.WebACLParameters{
							Region: "eu-central-1",
							VisibilityConfig: &svcapitypes.VisibilityConfig{
								MetricName:               &visibilityConfigMetricName,
								SampledRequestsEnabled:   &visibilityConfigSampledRequestsEnabled,
								CloudWatchMetricsEnabled: &visibilityConfigCloudWatchMetricsEnabled,
							},
							DefaultAction: &svcapitypes.DefaultAction{
								Allow: &svcapitypes.AllowAction{},
							},
							Scope: &scope,
							CustomWebACLParameters: svcapitypes.CustomWebACLParameters{
								AssociatedAWSResources: []*svcapitypes.AssociatedResource{
									{ResourceARN: aws.String("arn:aws:elasticloadbalancing:eu-central-1:123456789012:loadbalancer/app/new-load-balancer/60dc6c495c0c9188")},
								},
							},
							Tags: []*svcapitypes.Tag{
								{Key: &tag0Key, Value: &tag0Value},
							},
						},
					},
				},
				observed: &svcsdk.GetWebACLOutput{
					WebACL: &svcsdk.WebACL{
						Name:        &webAclName,
						Description: aws.String(""),
						Id:          &webAclId,
						VisibilityConfig: &svcsdk.VisibilityConfig{
							MetricName:               &visibilityConfigMetricName,
							SampledRequestsEnabled:   &visibilityConfigSampledRequestsEnabled,
							CloudWatchMetricsEnabled: &visibilityConfigCloudWatchMetricsEnabled,
						},
						DefaultAction: &svcsdk.DefaultAction{
							Allow: &svcsdk.AllowAction{},
						},
					},
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			s := shared{cache: &cache{}, client: tc.args.client}
			result, funcDiff, err := s.isUpToDate(context.Background(), tc.args.desired, tc.args.observed)
			if err != nil {
				t.Logf("error: %s", err.Error())
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, result); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
				if funcDiff != "" {
					t.Errorf("isUpTODate diff: %s", funcDiff)
				}
			}
		})
	}
}
