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

// NOTE(negz): We disable the gocyclo and staticcheck linters below because late
// init functions are inherently branchy, and because staticheck reports a lot
// of deprecation warnings in the CloudFront SDK that we can't address without a
// breaking change. We also need to turn off golint so that it doesn't complain
// about the nolint directive not being a valid package comment string. :| I
// believe this is only turning off these linters for this file.

//nolint:gocyclo,staticcheck,golint
package distribution

import (
	svcsdk "github.com/aws/aws-sdk-go/service/cloudfront"

	svcapitypes "github.com/crossplane/provider-aws/apis/cloudfront/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// TODO(negz): Every field of the generated DistributionParameters is a pointer.
// Are they really all optional and thus late-init-able? I'm pretty sure despite
// what the AWS SDK says that some of these fields are required.

// TODO(negz): We should ask the code generator to ignore all the 'quantity'
// fields and just infer them from the length of their associated items slice.
// Currently we only do this for Origins, for reasons that are unclear.
// https://github.com/crossplane/provider-aws/issues/789

func lateInitialize(dp *svcapitypes.DistributionParameters, gdo *svcsdk.GetDistributionOutput) error {
	if gdo.Distribution == nil || gdo.Distribution.DistributionConfig == nil {
		return nil
	}

	if dp.DistributionConfig == nil {
		dp.DistributionConfig = &svcapitypes.DistributionConfig{}
	}

	// For brevity, since this is really the only thing we're working with.
	in := dp.DistributionConfig
	from := gdo.Distribution.DistributionConfig

	if from.Aliases != nil {
		if in.Aliases == nil {
			in.Aliases = &svcapitypes.Aliases{}
		}

		in.Aliases.Items = awsclients.LateInitializeStringPtrSlice(in.Aliases.Items, from.Aliases.Items)
		in.Aliases.Quantity = awsclients.LateInitializeInt64Ptr(in.Aliases.Quantity, from.Aliases.Quantity)
	}

	if from.CacheBehaviors != nil {
		if in.CacheBehaviors == nil {
			in.CacheBehaviors = &svcapitypes.CacheBehaviors{}
		}

		lateInitCacheBehaviors(in.CacheBehaviors, from.CacheBehaviors)
	}

	in.Comment = awsclients.LateInitializeStringPtr(in.Comment, from.Comment)

	if from.CustomErrorResponses != nil {
		if in.CustomErrorResponses == nil {
			in.CustomErrorResponses = &svcapitypes.CustomErrorResponses{}
		}

		lateInitCustomErrorResponses(in.CustomErrorResponses, from.CustomErrorResponses)
	}

	if from.DefaultCacheBehavior != nil {
		if in.DefaultCacheBehavior == nil {
			in.DefaultCacheBehavior = &svcapitypes.DefaultCacheBehavior{}
		}

		lateInitDefaultCacheBehavior(in.DefaultCacheBehavior, from.DefaultCacheBehavior)
	}

	in.DefaultRootObject = awsclients.LateInitializeStringPtr(in.DefaultRootObject, from.DefaultRootObject)
	in.Enabled = awsclients.LateInitializeBoolPtr(in.Enabled, from.Enabled)
	in.HTTPVersion = awsclients.LateInitializeStringPtr(in.HTTPVersion, from.HttpVersion)
	in.IsIPV6Enabled = awsclients.LateInitializeBoolPtr(in.IsIPV6Enabled, from.IsIPV6Enabled)

	if from.Logging != nil {
		if in.Logging == nil {
			in.Logging = &svcapitypes.LoggingConfig{}
		}
		in.Logging.Bucket = awsclients.LateInitializeStringPtr(in.Logging.Bucket, from.Logging.Bucket)
		in.Logging.Enabled = awsclients.LateInitializeBoolPtr(in.Logging.Enabled, from.Logging.Enabled)
		in.Logging.IncludeCookies = awsclients.LateInitializeBoolPtr(in.Logging.IncludeCookies, from.Logging.IncludeCookies)
		in.Logging.Prefix = awsclients.LateInitializeStringPtr(in.Logging.Prefix, from.Logging.Prefix)
	}

	if from.OriginGroups != nil {
		if in.OriginGroups == nil {
			in.OriginGroups = &svcapitypes.OriginGroups{}
		}

		lateInitOriginGroups(in.OriginGroups, from.OriginGroups)
	}

	if from.Origins != nil {
		if in.Origins == nil {
			in.Origins = &svcapitypes.Origins{}
		}

		lateInitOrigins(in.Origins, from.Origins)
	}

	in.PriceClass = awsclients.LateInitializeStringPtr(in.PriceClass, from.PriceClass)

	if from.Restrictions != nil {
		if in.Restrictions == nil {
			in.Restrictions = &svcapitypes.Restrictions{}
		}

		if from.Restrictions.GeoRestriction != nil {
			if in.Restrictions.GeoRestriction == nil {
				in.Restrictions.GeoRestriction = &svcapitypes.GeoRestriction{}
			}

			in.Restrictions.GeoRestriction.Quantity = awsclients.LateInitializeInt64Ptr(in.Restrictions.GeoRestriction.Quantity, from.Restrictions.GeoRestriction.Quantity)
			in.Restrictions.GeoRestriction.Items = awsclients.LateInitializeStringPtrSlice(in.Restrictions.GeoRestriction.Items, from.Restrictions.GeoRestriction.Items)
			in.Restrictions.GeoRestriction.RestrictionType = awsclients.LateInitializeStringPtr(in.Restrictions.GeoRestriction.RestrictionType, from.Restrictions.GeoRestriction.RestrictionType)
		}
	}
	if from.ViewerCertificate != nil {
		if in.ViewerCertificate == nil {
			in.ViewerCertificate = &svcapitypes.ViewerCertificate{}
		}

		in.ViewerCertificate.ACMCertificateARN = awsclients.LateInitializeStringPtr(in.ViewerCertificate.ACMCertificateARN, from.ViewerCertificate.ACMCertificateArn)
		in.ViewerCertificate.Certificate = awsclients.LateInitializeStringPtr(in.ViewerCertificate.Certificate, from.ViewerCertificate.Certificate)
		in.ViewerCertificate.CertificateSource = awsclients.LateInitializeStringPtr(in.ViewerCertificate.CertificateSource, from.ViewerCertificate.CertificateSource)
		in.ViewerCertificate.CloudFrontDefaultCertificate = awsclients.LateInitializeBoolPtr(in.ViewerCertificate.CloudFrontDefaultCertificate, from.ViewerCertificate.CloudFrontDefaultCertificate)
		in.ViewerCertificate.IAMCertificateID = awsclients.LateInitializeStringPtr(in.ViewerCertificate.IAMCertificateID, from.ViewerCertificate.IAMCertificateId)
		in.ViewerCertificate.MinimumProtocolVersion = awsclients.LateInitializeStringPtr(in.ViewerCertificate.MinimumProtocolVersion, from.ViewerCertificate.MinimumProtocolVersion)
		in.ViewerCertificate.SSLSupportMethod = awsclients.LateInitializeStringPtr(in.ViewerCertificate.SSLSupportMethod, from.ViewerCertificate.SSLSupportMethod)
	}

	in.WebACLID = awsclients.LateInitializeStringPtr(in.WebACLID, from.WebACLId)

	return nil
}

func lateInitDefaultCacheBehavior(in *svcapitypes.DefaultCacheBehavior, from *svcsdk.DefaultCacheBehavior) {
	if from.AllowedMethods != nil {
		if in.AllowedMethods == nil {
			in.AllowedMethods = &svcapitypes.AllowedMethods{}
		}

		in.AllowedMethods.Items = awsclients.LateInitializeStringPtrSlice(in.AllowedMethods.Items, from.AllowedMethods.Items)
		in.AllowedMethods.Quantity = awsclients.LateInitializeInt64Ptr(in.AllowedMethods.Quantity, from.AllowedMethods.Quantity)

		if from.AllowedMethods.CachedMethods != nil {
			if in.AllowedMethods.CachedMethods == nil {
				in.AllowedMethods.CachedMethods = &svcapitypes.CachedMethods{}
			}

			in.AllowedMethods.CachedMethods.Items = awsclients.LateInitializeStringPtrSlice(in.AllowedMethods.CachedMethods.Items, from.AllowedMethods.CachedMethods.Items)
			in.AllowedMethods.CachedMethods.Quantity = awsclients.LateInitializeInt64Ptr(in.AllowedMethods.CachedMethods.Quantity, from.AllowedMethods.CachedMethods.Quantity)
		}
	}

	in.CachePolicyID = awsclients.LateInitializeStringPtr(in.CachePolicyID, from.CachePolicyId)
	in.Compress = awsclients.LateInitializeBoolPtr(in.Compress, from.Compress)
	in.DefaultTTL = awsclients.LateInitializeInt64Ptr(in.DefaultTTL, from.DefaultTTL)
	in.FieldLevelEncryptionID = awsclients.LateInitializeStringPtr(in.FieldLevelEncryptionID, from.FieldLevelEncryptionId)

	if from.ForwardedValues != nil {
		if in.ForwardedValues == nil {
			in.ForwardedValues = &svcapitypes.ForwardedValues{}
		}

		if from.ForwardedValues.Cookies != nil {
			if in.ForwardedValues.Cookies == nil {
				in.ForwardedValues.Cookies = &svcapitypes.CookiePreference{}
			}

			in.ForwardedValues.Cookies.Forward = awsclients.LateInitializeStringPtr(in.ForwardedValues.Cookies.Forward, from.ForwardedValues.Cookies.Forward)

			if from.ForwardedValues.Cookies.WhitelistedNames != nil {
				if in.ForwardedValues.Cookies.WhitelistedNames == nil {
					in.ForwardedValues.Cookies.WhitelistedNames = &svcapitypes.CookieNames{}
				}

				in.ForwardedValues.Cookies.WhitelistedNames.Items = awsclients.LateInitializeStringPtrSlice(in.ForwardedValues.Cookies.WhitelistedNames.Items, from.ForwardedValues.Cookies.WhitelistedNames.Items)
				in.ForwardedValues.Cookies.WhitelistedNames.Quantity = awsclients.LateInitializeInt64Ptr(in.ForwardedValues.Cookies.WhitelistedNames.Quantity, from.ForwardedValues.Cookies.WhitelistedNames.Quantity)
			}
		}

		if from.ForwardedValues.Headers != nil {
			if in.ForwardedValues.Headers == nil {
				in.ForwardedValues.Headers = &svcapitypes.Headers{}
			}

			in.ForwardedValues.Headers.Items = awsclients.LateInitializeStringPtrSlice(in.ForwardedValues.Headers.Items, from.ForwardedValues.Headers.Items)
			in.ForwardedValues.Headers.Quantity = awsclients.LateInitializeInt64Ptr(in.ForwardedValues.Headers.Quantity, from.ForwardedValues.Headers.Quantity)

		}

		in.ForwardedValues.QueryString = awsclients.LateInitializeBoolPtr(in.ForwardedValues.QueryString, from.ForwardedValues.QueryString)

		if from.ForwardedValues.QueryStringCacheKeys != nil {
			if in.ForwardedValues.QueryStringCacheKeys == nil {
				in.ForwardedValues.QueryStringCacheKeys = &svcapitypes.QueryStringCacheKeys{}
			}

			in.ForwardedValues.QueryStringCacheKeys.Items = awsclients.LateInitializeStringPtrSlice(in.ForwardedValues.QueryStringCacheKeys.Items, from.ForwardedValues.QueryStringCacheKeys.Items)
			in.ForwardedValues.QueryStringCacheKeys.Quantity = awsclients.LateInitializeInt64Ptr(in.ForwardedValues.QueryStringCacheKeys.Quantity, from.ForwardedValues.QueryStringCacheKeys.Quantity)
		}
	}

	if from.LambdaFunctionAssociations != nil {
		if in.LambdaFunctionAssociations == nil {
			in.LambdaFunctionAssociations = &svcapitypes.LambdaFunctionAssociations{}
		}
		lateInitLambdaFunctionAssociations(in.LambdaFunctionAssociations, from.LambdaFunctionAssociations)
	}

	in.MaxTTL = awsclients.LateInitializeInt64Ptr(in.MaxTTL, from.MaxTTL)
	in.MinTTL = awsclients.LateInitializeInt64Ptr(in.MinTTL, from.MinTTL)
	in.OriginRequestPolicyID = awsclients.LateInitializeStringPtr(in.OriginRequestPolicyID, from.OriginRequestPolicyId)
	in.RealtimeLogConfigARN = awsclients.LateInitializeStringPtr(in.RealtimeLogConfigARN, from.RealtimeLogConfigArn)
	in.SmoothStreaming = awsclients.LateInitializeBoolPtr(in.SmoothStreaming, from.SmoothStreaming)
	in.TargetOriginID = awsclients.LateInitializeStringPtr(in.TargetOriginID, from.TargetOriginId)

	if from.TrustedKeyGroups != nil {
		if in.TrustedKeyGroups == nil {
			in.TrustedKeyGroups = &svcapitypes.TrustedKeyGroups{}
		}

		in.TrustedKeyGroups.Enabled = awsclients.LateInitializeBoolPtr(in.TrustedKeyGroups.Enabled, from.TrustedKeyGroups.Enabled)
		in.TrustedKeyGroups.Items = awsclients.LateInitializeStringPtrSlice(in.TrustedKeyGroups.Items, from.TrustedKeyGroups.Items)
		in.TrustedKeyGroups.Quantity = awsclients.LateInitializeInt64Ptr(in.TrustedKeyGroups.Quantity, from.TrustedKeyGroups.Quantity)
	}

	if from.TrustedSigners != nil {
		if in.TrustedSigners == nil {
			in.TrustedSigners = &svcapitypes.TrustedSigners{}
		}

		in.TrustedSigners.Enabled = awsclients.LateInitializeBoolPtr(in.TrustedSigners.Enabled, from.TrustedSigners.Enabled)
		in.TrustedSigners.Items = awsclients.LateInitializeStringPtrSlice(in.TrustedSigners.Items, from.TrustedSigners.Items)
		in.TrustedSigners.Quantity = awsclients.LateInitializeInt64Ptr(in.TrustedSigners.Quantity, from.TrustedSigners.Quantity)
	}

	in.ViewerProtocolPolicy = awsclients.LateInitializeStringPtr(in.ViewerProtocolPolicy, from.ViewerProtocolPolicy)
}

func lateInitCacheBehaviors(in *svcapitypes.CacheBehaviors, from *svcsdk.CacheBehaviors) {
	in.Quantity = awsclients.LateInitializeInt64Ptr(in.Quantity, from.Quantity)

	if len(from.Items) == 0 || in.Items != nil {
		return
	}

	in.Items = make([]*svcapitypes.CacheBehavior, len(from.Items))
	for i := range from.Items {
		in.Items[i] = &svcapitypes.CacheBehavior{}
		lateInitCacheBehavior(in.Items[i], from.Items[i])
	}
}

// This is _almost_ identical to lateInitDefaultCacheBehaviour, but it has an
// additional 'PathPattern' field.
func lateInitCacheBehavior(in *svcapitypes.CacheBehavior, from *svcsdk.CacheBehavior) {
	if from.AllowedMethods != nil {
		if in.AllowedMethods == nil {
			in.AllowedMethods = &svcapitypes.AllowedMethods{}
		}

		in.AllowedMethods.Items = awsclients.LateInitializeStringPtrSlice(in.AllowedMethods.Items, from.AllowedMethods.Items)
		in.AllowedMethods.Quantity = awsclients.LateInitializeInt64Ptr(in.AllowedMethods.Quantity, from.AllowedMethods.Quantity)

		if from.AllowedMethods.CachedMethods != nil {
			if in.AllowedMethods.CachedMethods == nil {
				in.AllowedMethods.CachedMethods = &svcapitypes.CachedMethods{}
			}

			in.AllowedMethods.CachedMethods.Items = awsclients.LateInitializeStringPtrSlice(in.AllowedMethods.CachedMethods.Items, from.AllowedMethods.CachedMethods.Items)
			in.AllowedMethods.CachedMethods.Quantity = awsclients.LateInitializeInt64Ptr(in.AllowedMethods.CachedMethods.Quantity, from.AllowedMethods.CachedMethods.Quantity)
		}
	}

	in.CachePolicyID = awsclients.LateInitializeStringPtr(in.CachePolicyID, from.CachePolicyId)
	in.Compress = awsclients.LateInitializeBoolPtr(in.Compress, from.Compress)
	in.DefaultTTL = awsclients.LateInitializeInt64Ptr(in.DefaultTTL, from.DefaultTTL)
	in.FieldLevelEncryptionID = awsclients.LateInitializeStringPtr(in.FieldLevelEncryptionID, from.FieldLevelEncryptionId)

	if from.ForwardedValues != nil {
		if in.ForwardedValues == nil {
			in.ForwardedValues = &svcapitypes.ForwardedValues{}
		}

		if from.ForwardedValues.Cookies != nil {
			if in.ForwardedValues.Cookies == nil {
				in.ForwardedValues.Cookies = &svcapitypes.CookiePreference{}
			}

			in.ForwardedValues.Cookies.Forward = awsclients.LateInitializeStringPtr(in.ForwardedValues.Cookies.Forward, from.ForwardedValues.Cookies.Forward)

			if from.ForwardedValues.Cookies.WhitelistedNames != nil {
				if in.ForwardedValues.Cookies.WhitelistedNames == nil {
					in.ForwardedValues.Cookies.WhitelistedNames = &svcapitypes.CookieNames{}
				}

				in.ForwardedValues.Cookies.WhitelistedNames.Items = awsclients.LateInitializeStringPtrSlice(in.ForwardedValues.Cookies.WhitelistedNames.Items, from.ForwardedValues.Cookies.WhitelistedNames.Items)
				in.ForwardedValues.Cookies.WhitelistedNames.Quantity = awsclients.LateInitializeInt64Ptr(in.ForwardedValues.Cookies.WhitelistedNames.Quantity, from.ForwardedValues.Cookies.WhitelistedNames.Quantity)
			}
		}

		if from.ForwardedValues.Headers != nil {
			if in.ForwardedValues.Headers == nil {
				in.ForwardedValues.Headers = &svcapitypes.Headers{}
			}

			in.ForwardedValues.Headers.Items = awsclients.LateInitializeStringPtrSlice(in.ForwardedValues.Headers.Items, from.ForwardedValues.Headers.Items)
			in.ForwardedValues.Headers.Quantity = awsclients.LateInitializeInt64Ptr(in.ForwardedValues.Headers.Quantity, from.ForwardedValues.Headers.Quantity)

		}

		in.ForwardedValues.QueryString = awsclients.LateInitializeBoolPtr(in.ForwardedValues.QueryString, from.ForwardedValues.QueryString)

		if from.ForwardedValues.QueryStringCacheKeys != nil {
			if in.ForwardedValues.QueryStringCacheKeys == nil {
				in.ForwardedValues.QueryStringCacheKeys = &svcapitypes.QueryStringCacheKeys{}
			}

			in.ForwardedValues.QueryStringCacheKeys.Items = awsclients.LateInitializeStringPtrSlice(in.ForwardedValues.QueryStringCacheKeys.Items, from.ForwardedValues.QueryStringCacheKeys.Items)
			in.ForwardedValues.QueryStringCacheKeys.Quantity = awsclients.LateInitializeInt64Ptr(in.ForwardedValues.QueryStringCacheKeys.Quantity, from.ForwardedValues.QueryStringCacheKeys.Quantity)
		}
	}

	if from.LambdaFunctionAssociations != nil {
		if in.LambdaFunctionAssociations == nil {
			in.LambdaFunctionAssociations = &svcapitypes.LambdaFunctionAssociations{}
		}

		lateInitLambdaFunctionAssociations(in.LambdaFunctionAssociations, from.LambdaFunctionAssociations)
	}

	in.MaxTTL = awsclients.LateInitializeInt64Ptr(in.MaxTTL, from.MaxTTL)
	in.MinTTL = awsclients.LateInitializeInt64Ptr(in.MinTTL, from.MinTTL)
	in.OriginRequestPolicyID = awsclients.LateInitializeStringPtr(in.OriginRequestPolicyID, from.OriginRequestPolicyId)
	in.PathPattern = awsclients.LateInitializeStringPtr(in.PathPattern, from.PathPattern)
	in.RealtimeLogConfigARN = awsclients.LateInitializeStringPtr(in.RealtimeLogConfigARN, from.RealtimeLogConfigArn)
	in.SmoothStreaming = awsclients.LateInitializeBoolPtr(in.SmoothStreaming, from.SmoothStreaming)
	in.TargetOriginID = awsclients.LateInitializeStringPtr(in.TargetOriginID, from.TargetOriginId)

	if from.TrustedKeyGroups != nil {
		if in.TrustedKeyGroups == nil {
			in.TrustedKeyGroups = &svcapitypes.TrustedKeyGroups{}
		}

		in.TrustedKeyGroups.Enabled = awsclients.LateInitializeBoolPtr(in.TrustedKeyGroups.Enabled, from.TrustedKeyGroups.Enabled)
		in.TrustedKeyGroups.Items = awsclients.LateInitializeStringPtrSlice(in.TrustedKeyGroups.Items, from.TrustedKeyGroups.Items)
		in.TrustedKeyGroups.Quantity = awsclients.LateInitializeInt64Ptr(in.TrustedKeyGroups.Quantity, from.TrustedKeyGroups.Quantity)
	}

	if from.TrustedSigners != nil {
		if in.TrustedSigners == nil {
			in.TrustedSigners = &svcapitypes.TrustedSigners{}
		}

		in.TrustedSigners.Enabled = awsclients.LateInitializeBoolPtr(in.TrustedSigners.Enabled, from.TrustedSigners.Enabled)
		in.TrustedSigners.Items = awsclients.LateInitializeStringPtrSlice(in.TrustedSigners.Items, from.TrustedSigners.Items)
		in.TrustedSigners.Quantity = awsclients.LateInitializeInt64Ptr(in.TrustedSigners.Quantity, from.TrustedSigners.Quantity)
	}

	in.ViewerProtocolPolicy = awsclients.LateInitializeStringPtr(in.ViewerProtocolPolicy, from.ViewerProtocolPolicy)
}

func lateInitCustomErrorResponses(in *svcapitypes.CustomErrorResponses, from *svcsdk.CustomErrorResponses) {
	in.Quantity = awsclients.LateInitializeInt64Ptr(in.Quantity, from.Quantity)

	if len(from.Items) == 0 || in.Items != nil {
		return
	}

	in.Items = make([]*svcapitypes.CustomErrorResponse, len(from.Items))
	for i := range from.Items {
		in.Items[i] = &svcapitypes.CustomErrorResponse{}
		lateInitCustomErrorResponse(in.Items[i], from.Items[i])
	}
}

func lateInitCustomErrorResponse(in *svcapitypes.CustomErrorResponse, from *svcsdk.CustomErrorResponse) {
	in.ErrorCachingMinTTL = awsclients.LateInitializeInt64Ptr(in.ErrorCachingMinTTL, from.ErrorCachingMinTTL)
	in.ErrorCode = awsclients.LateInitializeInt64Ptr(in.ErrorCode, from.ErrorCode)
	in.ResponseCode = awsclients.LateInitializeStringPtr(in.ResponseCode, from.ResponseCode)
	in.ResponsePagePath = awsclients.LateInitializeStringPtr(in.ResponsePagePath, from.ResponsePagePath)
}

func lateInitOriginGroups(in *svcapitypes.OriginGroups, from *svcsdk.OriginGroups) {
	in.Quantity = awsclients.LateInitializeInt64Ptr(in.Quantity, from.Quantity)

	if len(from.Items) == 0 || in.Items != nil {
		return
	}

	in.Items = make([]*svcapitypes.OriginGroup, len(from.Items))
	for i := range from.Items {
		in.Items[i] = &svcapitypes.OriginGroup{}
		lateInitOriginGroup(in.Items[i], from.Items[i])
	}
}

func lateInitOriginGroup(in *svcapitypes.OriginGroup, from *svcsdk.OriginGroup) {
	if from.FailoverCriteria != nil {
		if in.FailoverCriteria == nil {
			in.FailoverCriteria = &svcapitypes.OriginGroupFailoverCriteria{}
		}

		if from.FailoverCriteria.StatusCodes != nil {
			if in.FailoverCriteria.StatusCodes == nil {
				in.FailoverCriteria.StatusCodes = &svcapitypes.StatusCodes{}
			}

			in.FailoverCriteria.StatusCodes.Items = awsclients.LateInitializeInt64PtrSlice(in.FailoverCriteria.StatusCodes.Items, from.FailoverCriteria.StatusCodes.Items)
			in.FailoverCriteria.StatusCodes.Quantity = awsclients.LateInitializeInt64Ptr(in.FailoverCriteria.StatusCodes.Quantity, from.FailoverCriteria.StatusCodes.Quantity)
		}
	}

	in.ID = awsclients.LateInitializeStringPtr(in.ID, from.Id)

	if from.Members != nil {
		if in.Members == nil {
			in.Members = &svcapitypes.OriginGroupMembers{}
		}

		lateInitOriginGroupMembers(in.Members, from.Members)
	}
}

func lateInitOriginGroupMembers(in *svcapitypes.OriginGroupMembers, from *svcsdk.OriginGroupMembers) {
	in.Quantity = awsclients.LateInitializeInt64Ptr(in.Quantity, from.Quantity)

	// TODO(negz): I believe OriginGroupMembers have an ID field, so we may
	// be able to match them by ID when late-initializing like we do for
	// Origins.
	if len(from.Items) == 0 || in.Items != nil {
		return
	}

	in.Items = make([]*svcapitypes.OriginGroupMember, len(from.Items))
	for i := range from.Items {
		in.Items[i] = &svcapitypes.OriginGroupMember{}
		lateInitOriginGroupMember(in.Items[i], from.Items[i])
	}
}

func lateInitOriginGroupMember(in *svcapitypes.OriginGroupMember, from *svcsdk.OriginGroupMember) {
	in.OriginID = awsclients.LateInitializeStringPtr(in.OriginID, from.OriginId)
}

// NOTE(negz): The CloudFront API relies heavily on late-initialization. There
// are more required fields when _updating_ a Distribution than there are when
// creating a Distribution. Callers are expected to create a Distribution, read
// back the defaulted fields, then supply those fields to do an update. This
// means we need to thoroughly late-init those defaulted fields.
//
// This is a problem because the API has a lot of fields that are optional at
// create time. It's an even bigger problem because many of the fields we need
// to late initialize are fields of structs that are slice members, and because
// the AWS API doesn't seem to return those slices in the same order they're
// sent so we can't simply use ordering to match them.
//
// We know we need to late-init origins for updates to work, and we can do that
// by using the ID key to match them.
//
// https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/distribution-overview-required-fields.html
func lateInitOrigins(in *svcapitypes.Origins, from *svcsdk.Origins) {
	// NOTE(negz): This is the only type where we the code generator ignores
	// the quantity field and we infer it from the slice's length.

	if len(from.Items) == 0 {
		return
	}

	// If we have no origins, just late init the entire slice.
	if in.Items == nil {
		in.Items = make([]*svcapitypes.Origin, len(from.Items))
		for i := range from.Items {
			in.Items[i] = &svcapitypes.Origin{}
			lateInitOrigin(in.Items[i], from.Items[i])
		}

		return
	}

	// If we have some origins we need to late init each one from its
	// corresponding origin in the API (if any).
	existing := make(map[string]*svcsdk.Origin)
	for i := range from.Items {
		o := from.Items[i]
		if o.Id == nil {
			continue
		}
		existing[awsclients.StringValue(o.Id)] = o
	}

	for i := range in.Items {
		io := in.Items[i]
		if io.ID == nil {
			continue
		}

		fo := existing[awsclients.StringValue(io.ID)]
		if fo == nil {
			continue
		}

		lateInitOrigin(io, fo)
	}
}

func lateInitOrigin(in *svcapitypes.Origin, from *svcsdk.Origin) {
	in.ConnectionAttempts = awsclients.LateInitializeInt64Ptr(in.ConnectionAttempts, from.ConnectionAttempts)
	in.ConnectionTimeout = awsclients.LateInitializeInt64Ptr(in.ConnectionTimeout, from.ConnectionTimeout)

	if from.CustomHeaders != nil {
		if in.CustomHeaders == nil {
			in.CustomHeaders = &svcapitypes.CustomHeaders{}
		}

		lateInitOriginCustomHeaders(in.CustomHeaders, from.CustomHeaders)
	}

	if from.CustomOriginConfig != nil {
		if in.CustomOriginConfig == nil {
			in.CustomOriginConfig = &svcapitypes.CustomOriginConfig{}
		}

		in.CustomOriginConfig.HTTPPort = awsclients.LateInitializeInt64Ptr(in.CustomOriginConfig.HTTPPort, from.CustomOriginConfig.HTTPPort)
		in.CustomOriginConfig.HTTPSPort = awsclients.LateInitializeInt64Ptr(in.CustomOriginConfig.HTTPSPort, from.CustomOriginConfig.HTTPSPort)
		in.CustomOriginConfig.OriginKeepaliveTimeout = awsclients.LateInitializeInt64Ptr(in.CustomOriginConfig.OriginKeepaliveTimeout, from.CustomOriginConfig.OriginKeepaliveTimeout)
		in.CustomOriginConfig.OriginProtocolPolicy = awsclients.LateInitializeStringPtr(in.CustomOriginConfig.OriginProtocolPolicy, from.CustomOriginConfig.OriginProtocolPolicy)
		in.CustomOriginConfig.OriginReadTimeout = awsclients.LateInitializeInt64Ptr(in.CustomOriginConfig.OriginReadTimeout, from.CustomOriginConfig.OriginReadTimeout)

		if from.CustomOriginConfig.OriginSslProtocols != nil {
			if in.CustomOriginConfig.OriginSSLProtocols == nil {
				in.CustomOriginConfig.OriginSSLProtocols = &svcapitypes.OriginSSLProtocols{}
			}

			in.CustomOriginConfig.OriginSSLProtocols.Items = awsclients.LateInitializeStringPtrSlice(in.CustomOriginConfig.OriginSSLProtocols.Items, from.CustomOriginConfig.OriginSslProtocols.Items)
			in.CustomOriginConfig.OriginSSLProtocols.Quantity = awsclients.LateInitializeInt64Ptr(in.CustomOriginConfig.OriginSSLProtocols.Quantity, from.CustomOriginConfig.OriginSslProtocols.Quantity)
		}
	}

	in.DomainName = awsclients.LateInitializeStringPtr(in.DomainName, from.DomainName)
	in.ID = awsclients.LateInitializeStringPtr(in.ID, from.Id)
	in.OriginPath = awsclients.LateInitializeStringPtr(in.OriginPath, from.OriginPath)

	if from.OriginShield != nil {
		if in.OriginShield == nil {
			in.OriginShield = &svcapitypes.OriginShield{}
		}

		in.OriginShield.Enabled = awsclients.LateInitializeBoolPtr(in.OriginShield.Enabled, from.OriginShield.Enabled)
		in.OriginShield.OriginShieldRegion = awsclients.LateInitializeStringPtr(in.OriginShield.OriginShieldRegion, from.OriginShield.OriginShieldRegion)
	}

	if from.S3OriginConfig != nil {
		if in.S3OriginConfig == nil {
			in.S3OriginConfig = &svcapitypes.S3OriginConfig{}
		}

		in.S3OriginConfig.OriginAccessIdentity = awsclients.LateInitializeStringPtr(in.S3OriginConfig.OriginAccessIdentity, from.S3OriginConfig.OriginAccessIdentity)
	}
}

func lateInitOriginCustomHeaders(in *svcapitypes.CustomHeaders, from *svcsdk.CustomHeaders) {
	in.Quantity = awsclients.LateInitializeInt64Ptr(in.Quantity, from.Quantity)

	if len(from.Items) == 0 || in.Items != nil {
		return
	}

	in.Items = make([]*svcapitypes.OriginCustomHeader, len(from.Items))
	for i := range from.Items {
		in.Items[i] = &svcapitypes.OriginCustomHeader{}
		lateInitOriginCustomHeader(in.Items[i], from.Items[i])
	}
}

func lateInitOriginCustomHeader(in *svcapitypes.OriginCustomHeader, from *svcsdk.OriginCustomHeader) {
	in.HeaderName = awsclients.LateInitializeStringPtr(in.HeaderName, from.HeaderName)
	in.HeaderValue = awsclients.LateInitializeStringPtr(in.HeaderValue, from.HeaderValue)
}

func lateInitLambdaFunctionAssociations(in *svcapitypes.LambdaFunctionAssociations, from *svcsdk.LambdaFunctionAssociations) {
	in.Quantity = awsclients.LateInitializeInt64Ptr(in.Quantity, from.Quantity)

	if len(from.Items) == 0 || in.Items != nil {
		return
	}

	in.Items = make([]*svcapitypes.LambdaFunctionAssociation, len(from.Items))
	for i := range from.Items {
		in.Items[i] = &svcapitypes.LambdaFunctionAssociation{}
		lateInitLambdaFunctionAssociation(in.Items[i], from.Items[i])
	}
}

func lateInitLambdaFunctionAssociation(in *svcapitypes.LambdaFunctionAssociation, from *svcsdk.LambdaFunctionAssociation) {
	in.EventType = awsclients.LateInitializeStringPtr(in.EventType, from.EventType)
	in.IncludeBody = awsclients.LateInitializeBoolPtr(in.IncludeBody, from.IncludeBody)
	in.LambdaFunctionARN = awsclients.LateInitializeStringPtr(in.LambdaFunctionARN, from.LambdaFunctionARN)
}
