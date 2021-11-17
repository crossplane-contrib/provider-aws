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

	svcapitypes "github.com/crossplane/provider-aws/apis/cloudfront/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
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
		"LateInitAllFields": {
			args: args{
				dp: &svcapitypes.DistributionParameters{},
				gdo: &svcsdk.GetDistributionOutput{
					Distribution: &svcsdk.Distribution{
						DistributionConfig: &svcsdk.DistributionConfig{
							Aliases: &svcsdk.Aliases{
								Items:    []*string{awsclients.String("example.org")},
								Quantity: awsclients.Int64(1),
							},
							CacheBehaviors: &svcsdk.CacheBehaviors{
								Items: []*svcsdk.CacheBehavior{{
									AllowedMethods: &svcsdk.AllowedMethods{
										Items:    []*string{awsclients.String("GET")},
										Quantity: awsclients.Int64(1),
										CachedMethods: &svcsdk.CachedMethods{
											Items:    []*string{awsclients.String("GET")},
											Quantity: awsclients.Int64(1),
										},
									},
									CachePolicyId:          awsclients.String("example"),
									Compress:               awsclients.Bool(true),
									DefaultTTL:             awsclients.Int64(42),
									FieldLevelEncryptionId: awsclients.String("example"),
									ForwardedValues: &svcsdk.ForwardedValues{
										Cookies: &svcsdk.CookiePreference{
											Forward: awsclients.String("example"),
											WhitelistedNames: &svcsdk.CookieNames{
												Items:    []*string{awsclients.String("example")},
												Quantity: awsclients.Int64(1),
											},
										},
										Headers: &svcsdk.Headers{
											Items:    []*string{awsclients.String("X-Hello")},
											Quantity: awsclients.Int64(1),
										},
										QueryString: awsclients.Bool(true),
										QueryStringCacheKeys: &svcsdk.QueryStringCacheKeys{
											Items:    []*string{awsclients.String("search")},
											Quantity: awsclients.Int64(1),
										},
									},
									LambdaFunctionAssociations: &svcsdk.LambdaFunctionAssociations{
										Items: []*svcsdk.LambdaFunctionAssociation{{
											EventType:         awsclients.String("good"),
											IncludeBody:       awsclients.Bool(true),
											LambdaFunctionARN: awsclients.String("arn"),
										}},
										Quantity: awsclients.Int64(1),
									},
									MaxTTL:                awsclients.Int64(42),
									MinTTL:                awsclients.Int64(42),
									OriginRequestPolicyId: awsclients.String("example"),
									PathPattern:           awsclients.String("example"),
									RealtimeLogConfigArn:  awsclients.String("example"),
									SmoothStreaming:       awsclients.Bool(true),
									TargetOriginId:        awsclients.String("example"),
									TrustedKeyGroups: &svcsdk.TrustedKeyGroups{
										Enabled:  awsclients.Bool(true),
										Items:    []*string{awsclients.String("the-good-key")},
										Quantity: awsclients.Int64(1),
									},
									TrustedSigners: &svcsdk.TrustedSigners{
										Enabled:  awsclients.Bool(true),
										Items:    []*string{awsclients.String("the-good-signer")},
										Quantity: awsclients.Int64(1),
									},
								}},
								Quantity: awsclients.Int64(1),
							},
							CustomErrorResponses: &svcsdk.CustomErrorResponses{
								Items: []*svcsdk.CustomErrorResponse{{
									ErrorCachingMinTTL: awsclients.Int64(42),
									ErrorCode:          awsclients.Int64(418),
									ResponseCode:       awsclients.String("I'm a teapot"),
									ResponsePagePath:   awsclients.String("/teapot"),
								}},
								Quantity: awsclients.Int64(1),
							},
							DefaultCacheBehavior: &svcsdk.DefaultCacheBehavior{
								AllowedMethods: &svcsdk.AllowedMethods{
									Items:    []*string{awsclients.String("GET")},
									Quantity: awsclients.Int64(1),
									CachedMethods: &svcsdk.CachedMethods{
										Items:    []*string{awsclients.String("GET")},
										Quantity: awsclients.Int64(1),
									},
								},
								CachePolicyId:          awsclients.String("example"),
								Compress:               awsclients.Bool(true),
								DefaultTTL:             awsclients.Int64(42),
								FieldLevelEncryptionId: awsclients.String("example"),
								ForwardedValues: &svcsdk.ForwardedValues{
									Cookies: &svcsdk.CookiePreference{
										Forward: awsclients.String("example"),
										WhitelistedNames: &svcsdk.CookieNames{
											Items:    []*string{awsclients.String("example")},
											Quantity: awsclients.Int64(1),
										},
									},
									Headers: &svcsdk.Headers{
										Items:    []*string{awsclients.String("X-Hello")},
										Quantity: awsclients.Int64(1),
									},
									QueryString: awsclients.Bool(true),
									QueryStringCacheKeys: &svcsdk.QueryStringCacheKeys{
										Items:    []*string{awsclients.String("search")},
										Quantity: awsclients.Int64(1),
									},
								},
								LambdaFunctionAssociations: &svcsdk.LambdaFunctionAssociations{
									Items: []*svcsdk.LambdaFunctionAssociation{{
										EventType:         awsclients.String("good"),
										IncludeBody:       awsclients.Bool(true),
										LambdaFunctionARN: awsclients.String("arn"),
									}},
									Quantity: awsclients.Int64(1),
								},
								MaxTTL:                awsclients.Int64(42),
								MinTTL:                awsclients.Int64(42),
								OriginRequestPolicyId: awsclients.String("example"),
								RealtimeLogConfigArn:  awsclients.String("example"),
								SmoothStreaming:       awsclients.Bool(true),
								TargetOriginId:        awsclients.String("example"),
								TrustedKeyGroups: &svcsdk.TrustedKeyGroups{
									Enabled:  awsclients.Bool(true),
									Items:    []*string{awsclients.String("the-good-key")},
									Quantity: awsclients.Int64(1),
								},
								TrustedSigners: &svcsdk.TrustedSigners{
									Enabled:  awsclients.Bool(true),
									Items:    []*string{awsclients.String("the-good-signer")},
									Quantity: awsclients.Int64(1),
								},
							},
							DefaultRootObject: awsclients.String("the-good-one"),
							Enabled:           awsclients.Bool(true),
							HttpVersion:       awsclients.String("1.1"),
							IsIPV6Enabled:     awsclients.Bool(true),
							Logging: &svcsdk.LoggingConfig{
								Bucket:         awsclients.String("big-logs"),
								Enabled:        awsclients.Bool(true),
								IncludeCookies: awsclients.Bool(true),
								Prefix:         awsclients.String("one-large-log-"),
							},
							OriginGroups: &svcsdk.OriginGroups{
								Items: []*svcsdk.OriginGroup{{
									FailoverCriteria: &svcsdk.OriginGroupFailoverCriteria{
										StatusCodes: &svcsdk.StatusCodes{
											Items:    []*int64{awsclients.Int64(418)},
											Quantity: awsclients.Int64(1),
										},
									},
									Members: &svcsdk.OriginGroupMembers{
										Items: []*svcsdk.OriginGroupMember{{
											OriginId: awsclients.String("example"),
										}},
										Quantity: awsclients.Int64(1),
									},
								}},
								Quantity: awsclients.Int64(1),
							},
							Origins: &svcsdk.Origins{
								Items: []*svcsdk.Origin{{
									ConnectionAttempts: awsclients.Int64(42),
									ConnectionTimeout:  awsclients.Int64(42),
									CustomHeaders: &svcsdk.CustomHeaders{
										Items: []*svcsdk.OriginCustomHeader{{
											HeaderName:  awsclients.String("X-Cool"),
											HeaderValue: awsclients.String("very"),
										}},
										Quantity: awsclients.Int64(1),
									},
									CustomOriginConfig: &svcsdk.CustomOriginConfig{
										HTTPPort:               awsclients.Int64(8080),
										HTTPSPort:              awsclients.Int64(443),
										OriginKeepaliveTimeout: awsclients.Int64(42),
										OriginProtocolPolicy:   awsclients.String("all-of-them"),
										OriginReadTimeout:      awsclients.Int64(42),
										OriginSslProtocols: &svcsdk.OriginSslProtocols{
											Items:    []*string{awsclients.String("TLS_1.2")},
											Quantity: awsclients.Int64(1),
										},
									},
									DomainName: awsclients.String("example.org"),
									Id:         awsclients.String("custom"),
									OriginPath: awsclients.String("/"),
									OriginShield: &svcsdk.OriginShield{
										Enabled:            awsclients.Bool(true),
										OriginShieldRegion: awsclients.String("us-east-1"),
									},
									S3OriginConfig: &svcsdk.S3OriginConfig{
										OriginAccessIdentity: awsclients.String("cool-guy"),
									},
								}},
								Quantity: awsclients.Int64(1),
							},
							PriceClass: awsclients.String("really-cheap"),
							Restrictions: &svcsdk.Restrictions{
								GeoRestriction: &svcsdk.GeoRestriction{
									RestrictionType: awsclients.String("no-australians"),
									Items:           []*string{awsclients.String("negz"), awsclients.String("kylie")},
									Quantity:        awsclients.Int64(1),
								},
							},
							ViewerCertificate: &svcsdk.ViewerCertificate{
								ACMCertificateArn:            awsclients.String("example"),
								Certificate:                  awsclients.String("example"),
								CertificateSource:            awsclients.String("trusty-source"),
								CloudFrontDefaultCertificate: awsclients.Bool(false),
								IAMCertificateId:             awsclients.String("example"),
								MinimumProtocolVersion:       awsclients.String("TLS_1.2"),
								SSLSupportMethod:             awsclients.String("fax"),
							},
							WebACLId: awsclients.String("example"),
						},
					},
				},
			},
			want: &svcapitypes.DistributionParameters{
				DistributionConfig: &svcapitypes.DistributionConfig{
					Aliases: &svcapitypes.Aliases{
						Items:    []*string{awsclients.String("example.org")},
						Quantity: awsclients.Int64(1),
					},
					CacheBehaviors: &svcapitypes.CacheBehaviors{
						Items: []*svcapitypes.CacheBehavior{{
							AllowedMethods: &svcapitypes.AllowedMethods{
								Items:    []*string{awsclients.String("GET")},
								Quantity: awsclients.Int64(1),
								CachedMethods: &svcapitypes.CachedMethods{
									Items:    []*string{awsclients.String("GET")},
									Quantity: awsclients.Int64(1),
								},
							},
							CachePolicyID:          awsclients.String("example"),
							Compress:               awsclients.Bool(true),
							DefaultTTL:             awsclients.Int64(42),
							FieldLevelEncryptionID: awsclients.String("example"),
							ForwardedValues: &svcapitypes.ForwardedValues{
								Cookies: &svcapitypes.CookiePreference{
									Forward: awsclients.String("example"),
									WhitelistedNames: &svcapitypes.CookieNames{
										Items:    []*string{awsclients.String("example")},
										Quantity: awsclients.Int64(1),
									},
								},
								Headers: &svcapitypes.Headers{
									Items:    []*string{awsclients.String("X-Hello")},
									Quantity: awsclients.Int64(1),
								},
								QueryString: awsclients.Bool(true),
								QueryStringCacheKeys: &svcapitypes.QueryStringCacheKeys{
									Items:    []*string{awsclients.String("search")},
									Quantity: awsclients.Int64(1),
								},
							},
							LambdaFunctionAssociations: &svcapitypes.LambdaFunctionAssociations{
								Items: []*svcapitypes.LambdaFunctionAssociation{{
									EventType:         awsclients.String("good"),
									IncludeBody:       awsclients.Bool(true),
									LambdaFunctionARN: awsclients.String("arn"),
								}},
								Quantity: awsclients.Int64(1),
							},
							MaxTTL:                awsclients.Int64(42),
							MinTTL:                awsclients.Int64(42),
							OriginRequestPolicyID: awsclients.String("example"),
							PathPattern:           awsclients.String("example"),
							RealtimeLogConfigARN:  awsclients.String("example"),
							SmoothStreaming:       awsclients.Bool(true),
							TargetOriginID:        awsclients.String("example"),
							TrustedKeyGroups: &svcapitypes.TrustedKeyGroups{
								Enabled:  awsclients.Bool(true),
								Items:    []*string{awsclients.String("the-good-key")},
								Quantity: awsclients.Int64(1),
							},
							TrustedSigners: &svcapitypes.TrustedSigners{
								Enabled:  awsclients.Bool(true),
								Items:    []*string{awsclients.String("the-good-signer")},
								Quantity: awsclients.Int64(1),
							},
						}},
						Quantity: awsclients.Int64(1),
					},
					CustomErrorResponses: &svcapitypes.CustomErrorResponses{
						Items: []*svcapitypes.CustomErrorResponse{{
							ErrorCachingMinTTL: awsclients.Int64(42),
							ErrorCode:          awsclients.Int64(418),
							ResponseCode:       awsclients.String("I'm a teapot"),
							ResponsePagePath:   awsclients.String("/teapot"),
						}},
						Quantity: awsclients.Int64(1),
					},
					DefaultCacheBehavior: &svcapitypes.DefaultCacheBehavior{
						AllowedMethods: &svcapitypes.AllowedMethods{
							Items:    []*string{awsclients.String("GET")},
							Quantity: awsclients.Int64(1),
							CachedMethods: &svcapitypes.CachedMethods{
								Items:    []*string{awsclients.String("GET")},
								Quantity: awsclients.Int64(1),
							},
						},
						CachePolicyID:          awsclients.String("example"),
						Compress:               awsclients.Bool(true),
						DefaultTTL:             awsclients.Int64(42),
						FieldLevelEncryptionID: awsclients.String("example"),
						ForwardedValues: &svcapitypes.ForwardedValues{
							Cookies: &svcapitypes.CookiePreference{
								Forward: awsclients.String("example"),
								WhitelistedNames: &svcapitypes.CookieNames{
									Items:    []*string{awsclients.String("example")},
									Quantity: awsclients.Int64(1),
								},
							},
							Headers: &svcapitypes.Headers{
								Items:    []*string{awsclients.String("X-Hello")},
								Quantity: awsclients.Int64(1),
							},
							QueryString: awsclients.Bool(true),
							QueryStringCacheKeys: &svcapitypes.QueryStringCacheKeys{
								Items:    []*string{awsclients.String("search")},
								Quantity: awsclients.Int64(1),
							},
						},
						LambdaFunctionAssociations: &svcapitypes.LambdaFunctionAssociations{
							Items: []*svcapitypes.LambdaFunctionAssociation{{
								EventType:         awsclients.String("good"),
								IncludeBody:       awsclients.Bool(true),
								LambdaFunctionARN: awsclients.String("arn"),
							}},
							Quantity: awsclients.Int64(1),
						},
						MaxTTL:                awsclients.Int64(42),
						MinTTL:                awsclients.Int64(42),
						OriginRequestPolicyID: awsclients.String("example"),
						RealtimeLogConfigARN:  awsclients.String("example"),
						SmoothStreaming:       awsclients.Bool(true),
						TargetOriginID:        awsclients.String("example"),
						TrustedKeyGroups: &svcapitypes.TrustedKeyGroups{
							Enabled:  awsclients.Bool(true),
							Items:    []*string{awsclients.String("the-good-key")},
							Quantity: awsclients.Int64(1),
						},
						TrustedSigners: &svcapitypes.TrustedSigners{
							Enabled:  awsclients.Bool(true),
							Items:    []*string{awsclients.String("the-good-signer")},
							Quantity: awsclients.Int64(1),
						},
					},
					DefaultRootObject: awsclients.String("the-good-one"),
					Enabled:           awsclients.Bool(true),
					HTTPVersion:       awsclients.String("1.1"),
					IsIPV6Enabled:     awsclients.Bool(true),
					Logging: &svcapitypes.LoggingConfig{
						Bucket:         awsclients.String("big-logs"),
						Enabled:        awsclients.Bool(true),
						IncludeCookies: awsclients.Bool(true),
						Prefix:         awsclients.String("one-large-log-"),
					},
					OriginGroups: &svcapitypes.OriginGroups{
						Items: []*svcapitypes.OriginGroup{{
							FailoverCriteria: &svcapitypes.OriginGroupFailoverCriteria{
								StatusCodes: &svcapitypes.StatusCodes{
									Items:    []*int64{awsclients.Int64(418)},
									Quantity: awsclients.Int64(1),
								},
							},
							Members: &svcapitypes.OriginGroupMembers{
								Items: []*svcapitypes.OriginGroupMember{{
									OriginID: awsclients.String("example"),
								}},
								Quantity: awsclients.Int64(1),
							},
						}},
						Quantity: awsclients.Int64(1),
					},
					Origins: &svcapitypes.Origins{
						Items: []*svcapitypes.Origin{{
							ConnectionAttempts: awsclients.Int64(42),
							ConnectionTimeout:  awsclients.Int64(42),
							CustomHeaders: &svcapitypes.CustomHeaders{
								Items: []*svcapitypes.OriginCustomHeader{{
									HeaderName:  awsclients.String("X-Cool"),
									HeaderValue: awsclients.String("very"),
								}},
								Quantity: awsclients.Int64(1),
							},
							CustomOriginConfig: &svcapitypes.CustomOriginConfig{
								HTTPPort:               awsclients.Int64(8080),
								HTTPSPort:              awsclients.Int64(443),
								OriginKeepaliveTimeout: awsclients.Int64(42),
								OriginProtocolPolicy:   awsclients.String("all-of-them"),
								OriginReadTimeout:      awsclients.Int64(42),
								OriginSSLProtocols: &svcapitypes.OriginSSLProtocols{
									Items:    []*string{awsclients.String("TLS_1.2")},
									Quantity: awsclients.Int64(1),
								},
							},
							DomainName: awsclients.String("example.org"),
							ID:         awsclients.String("custom"),
							OriginPath: awsclients.String("/"),
							OriginShield: &svcapitypes.OriginShield{
								Enabled:            awsclients.Bool(true),
								OriginShieldRegion: awsclients.String("us-east-1"),
							},
							S3OriginConfig: &svcapitypes.S3OriginConfig{
								OriginAccessIDentity: awsclients.String("cool-guy"),
							},
						}},
					},
					PriceClass: awsclients.String("really-cheap"),
					Restrictions: &svcapitypes.Restrictions{
						GeoRestriction: &svcapitypes.GeoRestriction{
							RestrictionType: awsclients.String("no-australians"),
							Items:           []*string{awsclients.String("negz"), awsclients.String("kylie")},
							Quantity:        awsclients.Int64(1),
						},
					},
					ViewerCertificate: &svcapitypes.ViewerCertificate{
						ACMCertificateARN:            awsclients.String("example"),
						Certificate:                  awsclients.String("example"),
						CertificateSource:            awsclients.String("trusty-source"),
						CloudFrontDefaultCertificate: awsclients.Bool(false),
						IAMCertificateID:             awsclients.String("example"),
						MinimumProtocolVersion:       awsclients.String("TLS_1.2"),
						SSLSupportMethod:             awsclients.String("fax"),
					},
					WebACLID: awsclients.String("example"),
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
									ID: awsclients.String("desired-only"),
								},
								{
									// We want to late-init domain-name here.
									ID: awsclients.String("custom"),
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
										Id: awsclients.String("actual-only"),
									},
									{
										DomainName: awsclients.String("example.org"),
										Id:         awsclients.String("custom"),
									},
								},
								Quantity: awsclients.Int64(3),
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
								ID: awsclients.String("desired-only"),
							},
							{
								DomainName: awsclients.String("example.org"),
								ID:         awsclients.String("custom"),
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
