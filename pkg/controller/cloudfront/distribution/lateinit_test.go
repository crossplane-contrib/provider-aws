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
								Items: []*string{pointer.String("example.org")},
							},
							CacheBehaviors: &svcsdk.CacheBehaviors{
								Items: []*svcsdk.CacheBehavior{{
									AllowedMethods: &svcsdk.AllowedMethods{
										Items: []*string{pointer.String("GET")},
										CachedMethods: &svcsdk.CachedMethods{
											Items: []*string{pointer.String("GET")},
										},
									},
									CachePolicyId:          pointer.String("example"),
									Compress:               pointer.Bool(true),
									DefaultTTL:             pointer.Int64(42),
									FieldLevelEncryptionId: pointer.String("example"),
									ForwardedValues: &svcsdk.ForwardedValues{
										Cookies: &svcsdk.CookiePreference{
											Forward: pointer.String("example"),
											WhitelistedNames: &svcsdk.CookieNames{
												Items: []*string{pointer.String("example")},
											},
										},
										Headers: &svcsdk.Headers{
											Items: []*string{pointer.String("X-Hello")},
										},
										QueryString: pointer.Bool(true),
										QueryStringCacheKeys: &svcsdk.QueryStringCacheKeys{
											Items: []*string{pointer.String("search")},
										},
									},
									LambdaFunctionAssociations: &svcsdk.LambdaFunctionAssociations{
										Items: []*svcsdk.LambdaFunctionAssociation{{
											EventType:         pointer.String("good"),
											IncludeBody:       pointer.Bool(true),
											LambdaFunctionARN: pointer.String("arn"),
										}},
									},
									MaxTTL:                pointer.Int64(42),
									MinTTL:                pointer.Int64(42),
									OriginRequestPolicyId: pointer.String("example"),
									PathPattern:           pointer.String("example"),
									RealtimeLogConfigArn:  pointer.String("example"),
									SmoothStreaming:       pointer.Bool(true),
									TargetOriginId:        pointer.String("example"),
									TrustedKeyGroups: &svcsdk.TrustedKeyGroups{
										Enabled: pointer.Bool(true),
										Items:   []*string{pointer.String("the-good-key")},
									},
									TrustedSigners: &svcsdk.TrustedSigners{
										Enabled: pointer.Bool(true),
										Items:   []*string{pointer.String("the-good-signer")},
									},
								}},
							},
							CustomErrorResponses: &svcsdk.CustomErrorResponses{
								Items: []*svcsdk.CustomErrorResponse{{
									ErrorCachingMinTTL: pointer.Int64(42),
									ErrorCode:          pointer.Int64(418),
									ResponseCode:       pointer.String("I'm a teapot"),
									ResponsePagePath:   pointer.String("/teapot"),
								}},
							},
							DefaultCacheBehavior: &svcsdk.DefaultCacheBehavior{
								AllowedMethods: &svcsdk.AllowedMethods{
									Items: []*string{pointer.String("GET")},
									CachedMethods: &svcsdk.CachedMethods{
										Items: []*string{pointer.String("GET")},
									},
								},
								CachePolicyId:          pointer.String("example"),
								Compress:               pointer.Bool(true),
								DefaultTTL:             pointer.Int64(42),
								FieldLevelEncryptionId: pointer.String("example"),
								ForwardedValues: &svcsdk.ForwardedValues{
									Cookies: &svcsdk.CookiePreference{
										Forward: pointer.String("example"),
										WhitelistedNames: &svcsdk.CookieNames{
											Items: []*string{pointer.String("example")},
										},
									},
									Headers: &svcsdk.Headers{
										Items: []*string{pointer.String("X-Hello")},
									},
									QueryString: pointer.Bool(true),
									QueryStringCacheKeys: &svcsdk.QueryStringCacheKeys{
										Items: []*string{pointer.String("search")},
									},
								},
								LambdaFunctionAssociations: &svcsdk.LambdaFunctionAssociations{
									Items: []*svcsdk.LambdaFunctionAssociation{{
										EventType:         pointer.String("good"),
										IncludeBody:       pointer.Bool(true),
										LambdaFunctionARN: pointer.String("arn"),
									}},
								},
								MaxTTL:                pointer.Int64(42),
								MinTTL:                pointer.Int64(42),
								OriginRequestPolicyId: pointer.String("example"),
								RealtimeLogConfigArn:  pointer.String("example"),
								SmoothStreaming:       pointer.Bool(true),
								TargetOriginId:        pointer.String("example"),
								TrustedKeyGroups: &svcsdk.TrustedKeyGroups{
									Enabled: pointer.Bool(true),
									Items:   []*string{pointer.String("the-good-key")},
								},
								TrustedSigners: &svcsdk.TrustedSigners{
									Enabled: pointer.Bool(true),
									Items:   []*string{pointer.String("the-good-signer")},
								},
							},
							DefaultRootObject: pointer.String("the-good-one"),
							Enabled:           pointer.Bool(true),
							HttpVersion:       pointer.String("1.1"),
							IsIPV6Enabled:     pointer.Bool(true),
							Logging: &svcsdk.LoggingConfig{
								Bucket:         pointer.String("big-logs"),
								Enabled:        pointer.Bool(true),
								IncludeCookies: pointer.Bool(true),
								Prefix:         pointer.String("one-large-log-"),
							},
							OriginGroups: &svcsdk.OriginGroups{
								Items: []*svcsdk.OriginGroup{{
									FailoverCriteria: &svcsdk.OriginGroupFailoverCriteria{
										StatusCodes: &svcsdk.StatusCodes{
											Items: []*int64{pointer.Int64(418)},
										},
									},
									Members: &svcsdk.OriginGroupMembers{
										Items: []*svcsdk.OriginGroupMember{{
											OriginId: pointer.String("example"),
										}},
									},
								}},
							},
							Origins: &svcsdk.Origins{
								Items: []*svcsdk.Origin{{
									ConnectionAttempts: pointer.Int64(42),
									ConnectionTimeout:  pointer.Int64(42),
									CustomHeaders: &svcsdk.CustomHeaders{
										Items: []*svcsdk.OriginCustomHeader{{
											HeaderName:  pointer.String("X-Cool"),
											HeaderValue: pointer.String("very"),
										}},
									},
									CustomOriginConfig: &svcsdk.CustomOriginConfig{
										HTTPPort:               pointer.Int64(8080),
										HTTPSPort:              pointer.Int64(443),
										OriginKeepaliveTimeout: pointer.Int64(42),
										OriginProtocolPolicy:   pointer.String("all-of-them"),
										OriginReadTimeout:      pointer.Int64(42),
										OriginSslProtocols: &svcsdk.OriginSslProtocols{
											Items: []*string{pointer.String("TLS_1.2")},
										},
									},
									DomainName: pointer.String("example.org"),
									Id:         pointer.String("custom"),
									OriginPath: pointer.String("/"),
									OriginShield: &svcsdk.OriginShield{
										Enabled:            pointer.Bool(true),
										OriginShieldRegion: pointer.String("us-east-1"),
									},
									S3OriginConfig: &svcsdk.S3OriginConfig{
										OriginAccessIdentity: pointer.String("cool-guy"),
									},
								}},
							},
							PriceClass: pointer.String("really-cheap"),
							Restrictions: &svcsdk.Restrictions{
								GeoRestriction: &svcsdk.GeoRestriction{
									RestrictionType: pointer.String("no-australians"),
									Items:           []*string{pointer.String("negz"), pointer.String("kylie")},
								},
							},
							ViewerCertificate: &svcsdk.ViewerCertificate{
								ACMCertificateArn:            pointer.String("example"),
								Certificate:                  pointer.String("example"),
								CertificateSource:            pointer.String("trusty-source"),
								CloudFrontDefaultCertificate: pointer.Bool(false),
								IAMCertificateId:             pointer.String("example"),
								MinimumProtocolVersion:       pointer.String("TLS_1.2"),
								SSLSupportMethod:             pointer.String("fax"),
							},
							WebACLId: pointer.String("example"),
						},
					},
				},
			},
			want: &svcapitypes.DistributionParameters{
				DistributionConfig: &svcapitypes.DistributionConfig{
					Aliases: &svcapitypes.Aliases{
						Items: []*string{pointer.String("example.org")},
					},
					CacheBehaviors: &svcapitypes.CacheBehaviors{
						Items: []*svcapitypes.CacheBehavior{{
							AllowedMethods: &svcapitypes.AllowedMethods{
								Items: []*string{pointer.String("GET")},
								CachedMethods: &svcapitypes.CachedMethods{
									Items: []*string{pointer.String("GET")},
								},
							},
							CachePolicyID:          pointer.String("example"),
							Compress:               pointer.Bool(true),
							DefaultTTL:             pointer.Int64(42),
							FieldLevelEncryptionID: pointer.String("example"),
							ForwardedValues: &svcapitypes.ForwardedValues{
								Cookies: &svcapitypes.CookiePreference{
									Forward: pointer.String("example"),
									WhitelistedNames: &svcapitypes.CookieNames{
										Items: []*string{pointer.String("example")},
									},
								},
								Headers: &svcapitypes.Headers{
									Items: []*string{pointer.String("X-Hello")},
								},
								QueryString: pointer.Bool(true),
								QueryStringCacheKeys: &svcapitypes.QueryStringCacheKeys{
									Items: []*string{pointer.String("search")},
								},
							},
							LambdaFunctionAssociations: &svcapitypes.LambdaFunctionAssociations{
								Items: []*svcapitypes.LambdaFunctionAssociation{{
									EventType:         pointer.String("good"),
									IncludeBody:       pointer.Bool(true),
									LambdaFunctionARN: pointer.String("arn"),
								}},
							},
							MaxTTL:                pointer.Int64(42),
							MinTTL:                pointer.Int64(42),
							OriginRequestPolicyID: pointer.String("example"),
							PathPattern:           pointer.String("example"),
							RealtimeLogConfigARN:  pointer.String("example"),
							SmoothStreaming:       pointer.Bool(true),
							TargetOriginID:        pointer.String("example"),
							TrustedKeyGroups: &svcapitypes.TrustedKeyGroups{
								Enabled: pointer.Bool(true),
								Items:   []*string{pointer.String("the-good-key")},
							},
							TrustedSigners: &svcapitypes.TrustedSigners{
								Enabled: pointer.Bool(true),
								Items:   []*string{pointer.String("the-good-signer")},
							},
						}},
					},
					CustomErrorResponses: &svcapitypes.CustomErrorResponses{
						Items: []*svcapitypes.CustomErrorResponse{{
							ErrorCachingMinTTL: pointer.Int64(42),
							ErrorCode:          pointer.Int64(418),
							ResponseCode:       pointer.String("I'm a teapot"),
							ResponsePagePath:   pointer.String("/teapot"),
						}},
					},
					DefaultCacheBehavior: &svcapitypes.DefaultCacheBehavior{
						AllowedMethods: &svcapitypes.AllowedMethods{
							Items: []*string{pointer.String("GET")},
							CachedMethods: &svcapitypes.CachedMethods{
								Items: []*string{pointer.String("GET")},
							},
						},
						CachePolicyID:          pointer.String("example"),
						Compress:               pointer.Bool(true),
						DefaultTTL:             pointer.Int64(42),
						FieldLevelEncryptionID: pointer.String("example"),
						ForwardedValues: &svcapitypes.ForwardedValues{
							Cookies: &svcapitypes.CookiePreference{
								Forward: pointer.String("example"),
								WhitelistedNames: &svcapitypes.CookieNames{
									Items: []*string{pointer.String("example")},
								},
							},
							Headers: &svcapitypes.Headers{
								Items: []*string{pointer.String("X-Hello")},
							},
							QueryString: pointer.Bool(true),
							QueryStringCacheKeys: &svcapitypes.QueryStringCacheKeys{
								Items: []*string{pointer.String("search")},
							},
						},
						LambdaFunctionAssociations: &svcapitypes.LambdaFunctionAssociations{
							Items: []*svcapitypes.LambdaFunctionAssociation{{
								EventType:         pointer.String("good"),
								IncludeBody:       pointer.Bool(true),
								LambdaFunctionARN: pointer.String("arn"),
							}},
						},
						MaxTTL:                pointer.Int64(42),
						MinTTL:                pointer.Int64(42),
						OriginRequestPolicyID: pointer.String("example"),
						RealtimeLogConfigARN:  pointer.String("example"),
						SmoothStreaming:       pointer.Bool(true),
						TargetOriginID:        pointer.String("example"),
						TrustedKeyGroups: &svcapitypes.TrustedKeyGroups{
							Enabled: pointer.Bool(true),
							Items:   []*string{pointer.String("the-good-key")},
						},
						TrustedSigners: &svcapitypes.TrustedSigners{
							Enabled: pointer.Bool(true),
							Items:   []*string{pointer.String("the-good-signer")},
						},
					},
					DefaultRootObject: pointer.String("the-good-one"),
					Enabled:           pointer.Bool(true),
					HTTPVersion:       pointer.String("1.1"),
					IsIPV6Enabled:     pointer.Bool(true),
					Logging: &svcapitypes.LoggingConfig{
						Bucket:         pointer.String("big-logs"),
						Enabled:        pointer.Bool(true),
						IncludeCookies: pointer.Bool(true),
						Prefix:         pointer.String("one-large-log-"),
					},
					OriginGroups: &svcapitypes.OriginGroups{
						Items: []*svcapitypes.OriginGroup{{
							FailoverCriteria: &svcapitypes.OriginGroupFailoverCriteria{
								StatusCodes: &svcapitypes.StatusCodes{
									Items: []*int64{pointer.Int64(418)},
								},
							},
							Members: &svcapitypes.OriginGroupMembers{
								Items: []*svcapitypes.OriginGroupMember{{
									OriginID: pointer.String("example"),
								}},
							},
						}},
					},
					Origins: &svcapitypes.Origins{
						Items: []*svcapitypes.Origin{{
							ConnectionAttempts: pointer.Int64(42),
							ConnectionTimeout:  pointer.Int64(42),
							CustomHeaders: &svcapitypes.CustomHeaders{
								Items: []*svcapitypes.OriginCustomHeader{{
									HeaderName:  pointer.String("X-Cool"),
									HeaderValue: pointer.String("very"),
								}},
							},
							CustomOriginConfig: &svcapitypes.CustomOriginConfig{
								HTTPPort:               pointer.Int64(8080),
								HTTPSPort:              pointer.Int64(443),
								OriginKeepaliveTimeout: pointer.Int64(42),
								OriginProtocolPolicy:   pointer.String("all-of-them"),
								OriginReadTimeout:      pointer.Int64(42),
								OriginSSLProtocols: &svcapitypes.OriginSSLProtocols{
									Items: []*string{pointer.String("TLS_1.2")},
								},
							},
							DomainName: pointer.String("example.org"),
							ID:         pointer.String("custom"),
							OriginPath: pointer.String("/"),
							OriginShield: &svcapitypes.OriginShield{
								Enabled:            pointer.Bool(true),
								OriginShieldRegion: pointer.String("us-east-1"),
							},
							S3OriginConfig: &svcapitypes.S3OriginConfig{
								OriginAccessIdentity: pointer.String("cool-guy"),
							},
						}},
					},
					PriceClass: pointer.String("really-cheap"),
					Restrictions: &svcapitypes.Restrictions{
						GeoRestriction: &svcapitypes.GeoRestriction{
							RestrictionType: pointer.String("no-australians"),
							Items:           []*string{pointer.String("negz"), pointer.String("kylie")},
						},
					},
					ViewerCertificate: &svcapitypes.ViewerCertificate{
						ACMCertificateARN:            pointer.String("example"),
						Certificate:                  pointer.String("example"),
						CertificateSource:            pointer.String("trusty-source"),
						CloudFrontDefaultCertificate: pointer.Bool(false),
						IAMCertificateID:             pointer.String("example"),
						MinimumProtocolVersion:       pointer.String("TLS_1.2"),
						SSLSupportMethod:             pointer.String("fax"),
					},
					WebACLID: pointer.String("example"),
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
									ID: pointer.String("desired-only"),
								},
								{
									// We want to late-init domain-name here.
									ID: pointer.String("custom"),
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
										Id: pointer.String("actual-only"),
									},
									{
										DomainName: pointer.String("example.org"),
										Id:         pointer.String("custom"),
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
								ID: pointer.String("desired-only"),
							},
							{
								DomainName: pointer.String("example.org"),
								ID:         pointer.String("custom"),
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
