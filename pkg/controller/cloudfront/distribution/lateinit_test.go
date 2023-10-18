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

package distribution

import (
	"testing"

	svcsdk "github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/google/go-cmp/cmp"
	"k8s.io/utils/ptr"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/cloudfront/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

func TestLateInitialize(t *testing.T) {
	type args struct {
		dp  *svcapitypes.DistributionParameters
		gdo *svcsdk.GetDistributionOutput
	}
	cases := map[string]struct {
		args args
		want *svcapitypes.DistributionParameters
	}{
		"NilDistribution": {
			args: args{
				gdo: &svcsdk.GetDistributionOutput{},
			},
		},
		"NilDistributionConfig": {
			args: args{
				gdo: &svcsdk.GetDistributionOutput{Distribution: &svcsdk.Distribution{}},
			},
		},
		// TODO(negz): Exhaustive tests for nil handling would be ideal.
		"NilDistributionConfigFields": {
			args: args{
				gdo: &svcsdk.GetDistributionOutput{Distribution: &svcsdk.Distribution{
					DistributionConfig: &svcsdk.DistributionConfig{},
				}},
				dp: &svcapitypes.DistributionParameters{
					DistributionConfig: &svcapitypes.DistributionConfig{},
				},
			},
			want: &svcapitypes.DistributionParameters{
				DistributionConfig: &svcapitypes.DistributionConfig{},
			},
		},
		"NilDistributionConfigChildrenFields": {
			args: args{
				gdo: &svcsdk.GetDistributionOutput{Distribution: &svcsdk.Distribution{
					DistributionConfig: &svcsdk.DistributionConfig{
						Aliases:              &svcsdk.Aliases{},
						CacheBehaviors:       &svcsdk.CacheBehaviors{},
						CustomErrorResponses: &svcsdk.CustomErrorResponses{},
						DefaultCacheBehavior: &svcsdk.DefaultCacheBehavior{},
						Logging:              &svcsdk.LoggingConfig{},
						OriginGroups:         &svcsdk.OriginGroups{},
						Origins:              &svcsdk.Origins{},
						Restrictions:         &svcsdk.Restrictions{},
						ViewerCertificate:    &svcsdk.ViewerCertificate{},
					},
				}},
				dp: &svcapitypes.DistributionParameters{
					DistributionConfig: &svcapitypes.DistributionConfig{},
				},
			},
			want: &svcapitypes.DistributionParameters{
				DistributionConfig: &svcapitypes.DistributionConfig{
					Aliases:              &svcapitypes.Aliases{},
					CacheBehaviors:       &svcapitypes.CacheBehaviors{},
					CustomErrorResponses: &svcapitypes.CustomErrorResponses{},
					DefaultCacheBehavior: &svcapitypes.DefaultCacheBehavior{},
					Logging:              &svcapitypes.LoggingConfig{},
					OriginGroups:         &svcapitypes.OriginGroups{},
					Origins:              &svcapitypes.Origins{},
					Restrictions:         &svcapitypes.Restrictions{},
					ViewerCertificate:    &svcapitypes.ViewerCertificate{},
				},
			},
		},
		"NilDistributionConfigGrandchildrenFields": {
			args: args{
				gdo: &svcsdk.GetDistributionOutput{Distribution: &svcsdk.Distribution{
					DistributionConfig: &svcsdk.DistributionConfig{
						CacheBehaviors: &svcsdk.CacheBehaviors{
							Items: []*svcsdk.CacheBehavior{{}},
						},
						CustomErrorResponses: &svcsdk.CustomErrorResponses{
							Items: []*svcsdk.CustomErrorResponse{{}},
						},
						OriginGroups: &svcsdk.OriginGroups{
							Items: []*svcsdk.OriginGroup{{}},
						},
						Origins: &svcsdk.Origins{
							Items: []*svcsdk.Origin{{}},
						},
						Restrictions: &svcsdk.Restrictions{
							GeoRestriction: &svcsdk.GeoRestriction{},
						},
					},
				}},
				dp: &svcapitypes.DistributionParameters{},
			},
			want: &svcapitypes.DistributionParameters{
				DistributionConfig: &svcapitypes.DistributionConfig{
					CacheBehaviors: &svcapitypes.CacheBehaviors{
						Items: []*svcapitypes.CacheBehavior{{}},
					},
					CustomErrorResponses: &svcapitypes.CustomErrorResponses{
						Items: []*svcapitypes.CustomErrorResponse{{}},
					},
					OriginGroups: &svcapitypes.OriginGroups{
						Items: []*svcapitypes.OriginGroup{{}},
					},
					Origins: &svcapitypes.Origins{
						Items: []*svcapitypes.Origin{{}},
					},
					Restrictions: &svcapitypes.Restrictions{
						GeoRestriction: &svcapitypes.GeoRestriction{},
					},
				},
			},
		},
		"LateInitAllFields": {
			args: args{
				dp: &svcapitypes.DistributionParameters{},
				gdo: &svcsdk.GetDistributionOutput{
					Distribution: &svcsdk.Distribution{
						DistributionConfig: &svcsdk.DistributionConfig{
							Aliases: &svcsdk.Aliases{
								Items: []*string{pointer.ToOrNilIfZeroValue("example.org")},
							},
							CacheBehaviors: &svcsdk.CacheBehaviors{
								Items: []*svcsdk.CacheBehavior{{
									AllowedMethods: &svcsdk.AllowedMethods{
										Items: []*string{pointer.ToOrNilIfZeroValue("GET")},
										CachedMethods: &svcsdk.CachedMethods{
											Items: []*string{pointer.ToOrNilIfZeroValue("GET")},
										},
									},
									CachePolicyId:          pointer.ToOrNilIfZeroValue("example"),
									Compress:               pointer.ToOrNilIfZeroValue(true),
									DefaultTTL:             ptr.To[int64](42),
									FieldLevelEncryptionId: pointer.ToOrNilIfZeroValue("example"),
									ForwardedValues: &svcsdk.ForwardedValues{
										Cookies: &svcsdk.CookiePreference{
											Forward: pointer.ToOrNilIfZeroValue("example"),
											WhitelistedNames: &svcsdk.CookieNames{
												Items: []*string{pointer.ToOrNilIfZeroValue("example")},
											},
										},
										Headers: &svcsdk.Headers{
											Items: []*string{pointer.ToOrNilIfZeroValue("X-Hello")},
										},
										QueryString: pointer.ToOrNilIfZeroValue(true),
										QueryStringCacheKeys: &svcsdk.QueryStringCacheKeys{
											Items: []*string{pointer.ToOrNilIfZeroValue("search")},
										},
									},
									LambdaFunctionAssociations: &svcsdk.LambdaFunctionAssociations{
										Items: []*svcsdk.LambdaFunctionAssociation{{
											EventType:         pointer.ToOrNilIfZeroValue("good"),
											IncludeBody:       pointer.ToOrNilIfZeroValue(true),
											LambdaFunctionARN: pointer.ToOrNilIfZeroValue("arn"),
										}},
									},
									MaxTTL:                ptr.To[int64](42),
									MinTTL:                ptr.To[int64](42),
									OriginRequestPolicyId: pointer.ToOrNilIfZeroValue("example"),
									PathPattern:           pointer.ToOrNilIfZeroValue("example"),
									RealtimeLogConfigArn:  pointer.ToOrNilIfZeroValue("example"),
									SmoothStreaming:       pointer.ToOrNilIfZeroValue(true),
									TargetOriginId:        pointer.ToOrNilIfZeroValue("example"),
									TrustedKeyGroups: &svcsdk.TrustedKeyGroups{
										Enabled: pointer.ToOrNilIfZeroValue(true),
										Items:   []*string{pointer.ToOrNilIfZeroValue("the-good-key")},
									},
									TrustedSigners: &svcsdk.TrustedSigners{
										Enabled: pointer.ToOrNilIfZeroValue(true),
										Items:   []*string{pointer.ToOrNilIfZeroValue("the-good-signer")},
									},
								}},
							},
							CustomErrorResponses: &svcsdk.CustomErrorResponses{
								Items: []*svcsdk.CustomErrorResponse{{
									ErrorCachingMinTTL: ptr.To[int64](42),
									ErrorCode:          ptr.To[int64](418),
									ResponseCode:       pointer.ToOrNilIfZeroValue("I'm a teapot"),
									ResponsePagePath:   pointer.ToOrNilIfZeroValue("/teapot"),
								}},
							},
							DefaultCacheBehavior: &svcsdk.DefaultCacheBehavior{
								AllowedMethods: &svcsdk.AllowedMethods{
									Items: []*string{pointer.ToOrNilIfZeroValue("GET")},
									CachedMethods: &svcsdk.CachedMethods{
										Items: []*string{pointer.ToOrNilIfZeroValue("GET")},
									},
								},
								CachePolicyId:          pointer.ToOrNilIfZeroValue("example"),
								Compress:               pointer.ToOrNilIfZeroValue(true),
								DefaultTTL:             ptr.To[int64](42),
								FieldLevelEncryptionId: pointer.ToOrNilIfZeroValue("example"),
								ForwardedValues: &svcsdk.ForwardedValues{
									Cookies: &svcsdk.CookiePreference{
										Forward: pointer.ToOrNilIfZeroValue("example"),
										WhitelistedNames: &svcsdk.CookieNames{
											Items: []*string{pointer.ToOrNilIfZeroValue("example")},
										},
									},
									Headers: &svcsdk.Headers{
										Items: []*string{pointer.ToOrNilIfZeroValue("X-Hello")},
									},
									QueryString: pointer.ToOrNilIfZeroValue(true),
									QueryStringCacheKeys: &svcsdk.QueryStringCacheKeys{
										Items: []*string{pointer.ToOrNilIfZeroValue("search")},
									},
								},
								LambdaFunctionAssociations: &svcsdk.LambdaFunctionAssociations{
									Items: []*svcsdk.LambdaFunctionAssociation{{
										EventType:         pointer.ToOrNilIfZeroValue("good"),
										IncludeBody:       pointer.ToOrNilIfZeroValue(true),
										LambdaFunctionARN: pointer.ToOrNilIfZeroValue("arn"),
									}},
								},
								MaxTTL:                ptr.To[int64](42),
								MinTTL:                ptr.To[int64](42),
								OriginRequestPolicyId: pointer.ToOrNilIfZeroValue("example"),
								RealtimeLogConfigArn:  pointer.ToOrNilIfZeroValue("example"),
								SmoothStreaming:       pointer.ToOrNilIfZeroValue(true),
								TargetOriginId:        pointer.ToOrNilIfZeroValue("example"),
								TrustedKeyGroups: &svcsdk.TrustedKeyGroups{
									Enabled: pointer.ToOrNilIfZeroValue(true),
									Items:   []*string{pointer.ToOrNilIfZeroValue("the-good-key")},
								},
								TrustedSigners: &svcsdk.TrustedSigners{
									Enabled: pointer.ToOrNilIfZeroValue(true),
									Items:   []*string{pointer.ToOrNilIfZeroValue("the-good-signer")},
								},
							},
							DefaultRootObject: pointer.ToOrNilIfZeroValue("the-good-one"),
							Enabled:           pointer.ToOrNilIfZeroValue(true),
							HttpVersion:       pointer.ToOrNilIfZeroValue("1.1"),
							IsIPV6Enabled:     pointer.ToOrNilIfZeroValue(true),
							Logging: &svcsdk.LoggingConfig{
								Bucket:         pointer.ToOrNilIfZeroValue("big-logs"),
								Enabled:        pointer.ToOrNilIfZeroValue(true),
								IncludeCookies: pointer.ToOrNilIfZeroValue(true),
								Prefix:         pointer.ToOrNilIfZeroValue("one-large-log-"),
							},
							OriginGroups: &svcsdk.OriginGroups{
								Items: []*svcsdk.OriginGroup{{
									FailoverCriteria: &svcsdk.OriginGroupFailoverCriteria{
										StatusCodes: &svcsdk.StatusCodes{
											Items: []*int64{ptr.To[int64](418)},
										},
									},
									Members: &svcsdk.OriginGroupMembers{
										Items: []*svcsdk.OriginGroupMember{{
											OriginId: pointer.ToOrNilIfZeroValue("example"),
										}},
									},
								}},
							},
							Origins: &svcsdk.Origins{
								Items: []*svcsdk.Origin{{
									ConnectionAttempts: ptr.To[int64](42),
									ConnectionTimeout:  ptr.To[int64](42),
									CustomHeaders: &svcsdk.CustomHeaders{
										Items: []*svcsdk.OriginCustomHeader{{
											HeaderName:  pointer.ToOrNilIfZeroValue("X-Cool"),
											HeaderValue: pointer.ToOrNilIfZeroValue("very"),
										}},
									},
									CustomOriginConfig: &svcsdk.CustomOriginConfig{
										HTTPPort:               ptr.To[int64](8080),
										HTTPSPort:              ptr.To[int64](443),
										OriginKeepaliveTimeout: ptr.To[int64](42),
										OriginProtocolPolicy:   pointer.ToOrNilIfZeroValue("all-of-them"),
										OriginReadTimeout:      ptr.To[int64](42),
										OriginSslProtocols: &svcsdk.OriginSslProtocols{
											Items: []*string{pointer.ToOrNilIfZeroValue("TLS_1.2")},
										},
									},
									DomainName: pointer.ToOrNilIfZeroValue("example.org"),
									Id:         pointer.ToOrNilIfZeroValue("custom"),
									OriginPath: pointer.ToOrNilIfZeroValue("/"),
									OriginShield: &svcsdk.OriginShield{
										Enabled:            pointer.ToOrNilIfZeroValue(true),
										OriginShieldRegion: pointer.ToOrNilIfZeroValue("us-east-1"),
									},
									S3OriginConfig: &svcsdk.S3OriginConfig{
										OriginAccessIdentity: pointer.ToOrNilIfZeroValue("cool-guy"),
									},
								}},
							},
							PriceClass: pointer.ToOrNilIfZeroValue("really-cheap"),
							Restrictions: &svcsdk.Restrictions{
								GeoRestriction: &svcsdk.GeoRestriction{
									RestrictionType: pointer.ToOrNilIfZeroValue("no-australians"),
									Items:           []*string{pointer.ToOrNilIfZeroValue("negz"), pointer.ToOrNilIfZeroValue("kylie")},
								},
							},
							ViewerCertificate: &svcsdk.ViewerCertificate{
								ACMCertificateArn:            pointer.ToOrNilIfZeroValue("example"),
								Certificate:                  pointer.ToOrNilIfZeroValue("example"),
								CertificateSource:            pointer.ToOrNilIfZeroValue("trusty-source"),
								CloudFrontDefaultCertificate: pointer.ToOrNilIfZeroValue(false),
								IAMCertificateId:             pointer.ToOrNilIfZeroValue("example"),
								MinimumProtocolVersion:       pointer.ToOrNilIfZeroValue("TLS_1.2"),
								SSLSupportMethod:             pointer.ToOrNilIfZeroValue("fax"),
							},
							WebACLId: pointer.ToOrNilIfZeroValue("example"),
						},
					},
				},
			},
			want: &svcapitypes.DistributionParameters{
				DistributionConfig: &svcapitypes.DistributionConfig{
					Aliases: &svcapitypes.Aliases{
						Items: []*string{pointer.ToOrNilIfZeroValue("example.org")},
					},
					CacheBehaviors: &svcapitypes.CacheBehaviors{
						Items: []*svcapitypes.CacheBehavior{{
							AllowedMethods: &svcapitypes.AllowedMethods{
								Items: []*string{pointer.ToOrNilIfZeroValue("GET")},
								CachedMethods: &svcapitypes.CachedMethods{
									Items: []*string{pointer.ToOrNilIfZeroValue("GET")},
								},
							},
							CachePolicyID:          pointer.ToOrNilIfZeroValue("example"),
							Compress:               pointer.ToOrNilIfZeroValue(true),
							DefaultTTL:             ptr.To[int64](42),
							FieldLevelEncryptionID: pointer.ToOrNilIfZeroValue("example"),
							ForwardedValues: &svcapitypes.ForwardedValues{
								Cookies: &svcapitypes.CookiePreference{
									Forward: pointer.ToOrNilIfZeroValue("example"),
									WhitelistedNames: &svcapitypes.CookieNames{
										Items: []*string{pointer.ToOrNilIfZeroValue("example")},
									},
								},
								Headers: &svcapitypes.Headers{
									Items: []*string{pointer.ToOrNilIfZeroValue("X-Hello")},
								},
								QueryString: pointer.ToOrNilIfZeroValue(true),
								QueryStringCacheKeys: &svcapitypes.QueryStringCacheKeys{
									Items: []*string{pointer.ToOrNilIfZeroValue("search")},
								},
							},
							LambdaFunctionAssociations: &svcapitypes.LambdaFunctionAssociations{
								Items: []*svcapitypes.LambdaFunctionAssociation{{
									EventType:         pointer.ToOrNilIfZeroValue("good"),
									IncludeBody:       pointer.ToOrNilIfZeroValue(true),
									LambdaFunctionARN: pointer.ToOrNilIfZeroValue("arn"),
								}},
							},
							MaxTTL:                ptr.To[int64](42),
							MinTTL:                ptr.To[int64](42),
							OriginRequestPolicyID: pointer.ToOrNilIfZeroValue("example"),
							PathPattern:           pointer.ToOrNilIfZeroValue("example"),
							RealtimeLogConfigARN:  pointer.ToOrNilIfZeroValue("example"),
							SmoothStreaming:       pointer.ToOrNilIfZeroValue(true),
							TargetOriginID:        pointer.ToOrNilIfZeroValue("example"),
							TrustedKeyGroups: &svcapitypes.TrustedKeyGroups{
								Enabled: pointer.ToOrNilIfZeroValue(true),
								Items:   []*string{pointer.ToOrNilIfZeroValue("the-good-key")},
							},
							TrustedSigners: &svcapitypes.TrustedSigners{
								Enabled: pointer.ToOrNilIfZeroValue(true),
								Items:   []*string{pointer.ToOrNilIfZeroValue("the-good-signer")},
							},
						}},
					},
					CustomErrorResponses: &svcapitypes.CustomErrorResponses{
						Items: []*svcapitypes.CustomErrorResponse{{
							ErrorCachingMinTTL: ptr.To[int64](42),
							ErrorCode:          ptr.To[int64](418),
							ResponseCode:       pointer.ToOrNilIfZeroValue("I'm a teapot"),
							ResponsePagePath:   pointer.ToOrNilIfZeroValue("/teapot"),
						}},
					},
					DefaultCacheBehavior: &svcapitypes.DefaultCacheBehavior{
						AllowedMethods: &svcapitypes.AllowedMethods{
							Items: []*string{pointer.ToOrNilIfZeroValue("GET")},
							CachedMethods: &svcapitypes.CachedMethods{
								Items: []*string{pointer.ToOrNilIfZeroValue("GET")},
							},
						},
						CachePolicyID:          pointer.ToOrNilIfZeroValue("example"),
						Compress:               pointer.ToOrNilIfZeroValue(true),
						DefaultTTL:             ptr.To[int64](42),
						FieldLevelEncryptionID: pointer.ToOrNilIfZeroValue("example"),
						ForwardedValues: &svcapitypes.ForwardedValues{
							Cookies: &svcapitypes.CookiePreference{
								Forward: pointer.ToOrNilIfZeroValue("example"),
								WhitelistedNames: &svcapitypes.CookieNames{
									Items: []*string{pointer.ToOrNilIfZeroValue("example")},
								},
							},
							Headers: &svcapitypes.Headers{
								Items: []*string{pointer.ToOrNilIfZeroValue("X-Hello")},
							},
							QueryString: pointer.ToOrNilIfZeroValue(true),
							QueryStringCacheKeys: &svcapitypes.QueryStringCacheKeys{
								Items: []*string{pointer.ToOrNilIfZeroValue("search")},
							},
						},
						LambdaFunctionAssociations: &svcapitypes.LambdaFunctionAssociations{
							Items: []*svcapitypes.LambdaFunctionAssociation{{
								EventType:         pointer.ToOrNilIfZeroValue("good"),
								IncludeBody:       pointer.ToOrNilIfZeroValue(true),
								LambdaFunctionARN: pointer.ToOrNilIfZeroValue("arn"),
							}},
						},
						MaxTTL:                ptr.To[int64](42),
						MinTTL:                ptr.To[int64](42),
						OriginRequestPolicyID: pointer.ToOrNilIfZeroValue("example"),
						RealtimeLogConfigARN:  pointer.ToOrNilIfZeroValue("example"),
						SmoothStreaming:       pointer.ToOrNilIfZeroValue(true),
						TargetOriginID:        pointer.ToOrNilIfZeroValue("example"),
						TrustedKeyGroups: &svcapitypes.TrustedKeyGroups{
							Enabled: pointer.ToOrNilIfZeroValue(true),
							Items:   []*string{pointer.ToOrNilIfZeroValue("the-good-key")},
						},
						TrustedSigners: &svcapitypes.TrustedSigners{
							Enabled: pointer.ToOrNilIfZeroValue(true),
							Items:   []*string{pointer.ToOrNilIfZeroValue("the-good-signer")},
						},
					},
					DefaultRootObject: pointer.ToOrNilIfZeroValue("the-good-one"),
					Enabled:           pointer.ToOrNilIfZeroValue(true),
					HTTPVersion:       pointer.ToOrNilIfZeroValue("1.1"),
					IsIPV6Enabled:     pointer.ToOrNilIfZeroValue(true),
					Logging: &svcapitypes.LoggingConfig{
						Bucket:         pointer.ToOrNilIfZeroValue("big-logs"),
						Enabled:        pointer.ToOrNilIfZeroValue(true),
						IncludeCookies: pointer.ToOrNilIfZeroValue(true),
						Prefix:         pointer.ToOrNilIfZeroValue("one-large-log-"),
					},
					OriginGroups: &svcapitypes.OriginGroups{
						Items: []*svcapitypes.OriginGroup{{
							FailoverCriteria: &svcapitypes.OriginGroupFailoverCriteria{
								StatusCodes: &svcapitypes.StatusCodes{
									Items: []*int64{ptr.To[int64](418)},
								},
							},
							Members: &svcapitypes.OriginGroupMembers{
								Items: []*svcapitypes.OriginGroupMember{{
									OriginID: pointer.ToOrNilIfZeroValue("example"),
								}},
							},
						}},
					},
					Origins: &svcapitypes.Origins{
						Items: []*svcapitypes.Origin{{
							ConnectionAttempts: ptr.To[int64](42),
							ConnectionTimeout:  ptr.To[int64](42),
							CustomHeaders: &svcapitypes.CustomHeaders{
								Items: []*svcapitypes.OriginCustomHeader{{
									HeaderName:  pointer.ToOrNilIfZeroValue("X-Cool"),
									HeaderValue: pointer.ToOrNilIfZeroValue("very"),
								}},
							},
							CustomOriginConfig: &svcapitypes.CustomOriginConfig{
								HTTPPort:               ptr.To[int64](8080),
								HTTPSPort:              ptr.To[int64](443),
								OriginKeepaliveTimeout: ptr.To[int64](42),
								OriginProtocolPolicy:   pointer.ToOrNilIfZeroValue("all-of-them"),
								OriginReadTimeout:      ptr.To[int64](42),
								OriginSSLProtocols: &svcapitypes.OriginSSLProtocols{
									Items: []*string{pointer.ToOrNilIfZeroValue("TLS_1.2")},
								},
							},
							DomainName: pointer.ToOrNilIfZeroValue("example.org"),
							ID:         pointer.ToOrNilIfZeroValue("custom"),
							OriginPath: pointer.ToOrNilIfZeroValue("/"),
							OriginShield: &svcapitypes.OriginShield{
								Enabled:            pointer.ToOrNilIfZeroValue(true),
								OriginShieldRegion: pointer.ToOrNilIfZeroValue("us-east-1"),
							},
							S3OriginConfig: &svcapitypes.S3OriginConfig{
								OriginAccessIdentity: pointer.ToOrNilIfZeroValue("cool-guy"),
							},
						}},
					},
					PriceClass: pointer.ToOrNilIfZeroValue("really-cheap"),
					Restrictions: &svcapitypes.Restrictions{
						GeoRestriction: &svcapitypes.GeoRestriction{
							RestrictionType: pointer.ToOrNilIfZeroValue("no-australians"),
							Items:           []*string{pointer.ToOrNilIfZeroValue("negz"), pointer.ToOrNilIfZeroValue("kylie")},
						},
					},
					ViewerCertificate: &svcapitypes.ViewerCertificate{
						ACMCertificateARN:            pointer.ToOrNilIfZeroValue("example"),
						Certificate:                  pointer.ToOrNilIfZeroValue("example"),
						CertificateSource:            pointer.ToOrNilIfZeroValue("trusty-source"),
						CloudFrontDefaultCertificate: pointer.ToOrNilIfZeroValue(false),
						IAMCertificateID:             pointer.ToOrNilIfZeroValue("example"),
						MinimumProtocolVersion:       pointer.ToOrNilIfZeroValue("TLS_1.2"),
						SSLSupportMethod:             pointer.ToOrNilIfZeroValue("fax"),
					},
					WebACLID: pointer.ToOrNilIfZeroValue("example"),
				},
			},
		},
		"LateInitOriginsByID": {
			args: args{
				dp: &svcapitypes.DistributionParameters{
					DistributionConfig: &svcapitypes.DistributionConfig{
						Origins: &svcapitypes.Origins{
							Items: []*svcapitypes.Origin{
								{}, // This one has a nil ID.
								{
									// This one only exists in desired state.
									ID: pointer.ToOrNilIfZeroValue("desired-only"),
								},
								{
									// We want to late-init domain-name here.
									ID: pointer.ToOrNilIfZeroValue("custom"),
								},
							},
						},
					},
				},
				gdo: &svcsdk.GetDistributionOutput{
					Distribution: &svcsdk.Distribution{
						DistributionConfig: &svcsdk.DistributionConfig{
							Origins: &svcsdk.Origins{
								Items: []*svcsdk.Origin{
									{}, // This one has a nil Id.
									{
										// This one only exists in actual state.
										Id: pointer.ToOrNilIfZeroValue("actual-only"),
									},
									{
										DomainName: pointer.ToOrNilIfZeroValue("example.org"),
										Id:         pointer.ToOrNilIfZeroValue("custom"),
									},
								},
							},
						},
					},
				},
			},
			want: &svcapitypes.DistributionParameters{
				DistributionConfig: &svcapitypes.DistributionConfig{
					Origins: &svcapitypes.Origins{
						Items: []*svcapitypes.Origin{
							{}, // This one has a nil ID.
							{
								// This one only exists in desired state.
								ID: pointer.ToOrNilIfZeroValue("desired-only"),
							},
							{
								DomainName: pointer.ToOrNilIfZeroValue("example.org"),
								ID:         pointer.ToOrNilIfZeroValue("custom"),
							},
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_ = lateInitialize(tc.args.dp, tc.args.gdo)
			if diff := cmp.Diff(tc.want, tc.args.dp); diff != "" {
				t.Errorf("\nlateInitialize(...): -want, +got:\n%s", diff)
			}

		})
	}

}
