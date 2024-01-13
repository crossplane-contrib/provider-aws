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

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/cloudfront/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

// TODO(negz): Every field of the generated DistributionParameters is a pointer.
// Are they really all optional and thus late-init-able? I'm pretty sure despite
// what the AWS SDK says that some of these fields are required.

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

	// "CallerReference" is not exposed to users from the console, therefore we need to derive it from the AWS output
	// during import scenarios
	in.CallerReference = pointer.LateInitialize(in.CallerReference, from.CallerReference)

	if from.Aliases != nil {
		if in.Aliases == nil {
			in.Aliases = &svcapitypes.Aliases{}
		}

		in.Aliases.Items = pointer.LateInitializeSlice(in.Aliases.Items, from.Aliases.Items)
	}

	if from.CacheBehaviors != nil {
		if in.CacheBehaviors == nil {
			in.CacheBehaviors = &svcapitypes.CacheBehaviors{}
		}

		lateInitCacheBehaviors(in.CacheBehaviors, from.CacheBehaviors)
	}

	in.Comment = pointer.LateInitialize(in.Comment, from.Comment)

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

	in.DefaultRootObject = pointer.LateInitialize(in.DefaultRootObject, from.DefaultRootObject)
	in.Enabled = pointer.LateInitialize(in.Enabled, from.Enabled)
	in.HTTPVersion = pointer.LateInitialize(in.HTTPVersion, from.HttpVersion)
	in.IsIPV6Enabled = pointer.LateInitialize(in.IsIPV6Enabled, from.IsIPV6Enabled)

	if from.Logging != nil {
		if in.Logging == nil {
			in.Logging = &svcapitypes.LoggingConfig{}
		}
		in.Logging.Bucket = pointer.LateInitialize(in.Logging.Bucket, from.Logging.Bucket)
		in.Logging.Enabled = pointer.LateInitialize(in.Logging.Enabled, from.Logging.Enabled)
		in.Logging.IncludeCookies = pointer.LateInitialize(in.Logging.IncludeCookies, from.Logging.IncludeCookies)
		in.Logging.Prefix = pointer.LateInitialize(in.Logging.Prefix, from.Logging.Prefix)
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

	in.PriceClass = pointer.LateInitialize(in.PriceClass, from.PriceClass)

	if from.Restrictions != nil {
		if in.Restrictions == nil {
			in.Restrictions = &svcapitypes.Restrictions{}
		}

		if from.Restrictions.GeoRestriction != nil {
			if in.Restrictions.GeoRestriction == nil {
				in.Restrictions.GeoRestriction = &svcapitypes.GeoRestriction{}
			}

			in.Restrictions.GeoRestriction.Items = pointer.LateInitializeSlice(in.Restrictions.GeoRestriction.Items, from.Restrictions.GeoRestriction.Items)
			in.Restrictions.GeoRestriction.RestrictionType = pointer.LateInitialize(in.Restrictions.GeoRestriction.RestrictionType, from.Restrictions.GeoRestriction.RestrictionType)
		}
	}
	if from.ViewerCertificate != nil {
		if in.ViewerCertificate == nil {
			in.ViewerCertificate = &svcapitypes.ViewerCertificate{}
		}

		in.ViewerCertificate.ACMCertificateARN = pointer.LateInitialize(in.ViewerCertificate.ACMCertificateARN, from.ViewerCertificate.ACMCertificateArn)
		in.ViewerCertificate.Certificate = pointer.LateInitialize(in.ViewerCertificate.Certificate, from.ViewerCertificate.Certificate)
		in.ViewerCertificate.CertificateSource = pointer.LateInitialize(in.ViewerCertificate.CertificateSource, from.ViewerCertificate.CertificateSource)
		in.ViewerCertificate.CloudFrontDefaultCertificate = pointer.LateInitialize(in.ViewerCertificate.CloudFrontDefaultCertificate, from.ViewerCertificate.CloudFrontDefaultCertificate)
		in.ViewerCertificate.IAMCertificateID = pointer.LateInitialize(in.ViewerCertificate.IAMCertificateID, from.ViewerCertificate.IAMCertificateId)
		in.ViewerCertificate.MinimumProtocolVersion = pointer.LateInitialize(in.ViewerCertificate.MinimumProtocolVersion, from.ViewerCertificate.MinimumProtocolVersion)
		in.ViewerCertificate.SSLSupportMethod = pointer.LateInitialize(in.ViewerCertificate.SSLSupportMethod, from.ViewerCertificate.SSLSupportMethod)
	}

	in.WebACLID = pointer.LateInitialize(in.WebACLID, from.WebACLId)

	return nil
}

func lateInitDefaultCacheBehavior(in *svcapitypes.DefaultCacheBehavior, from *svcsdk.DefaultCacheBehavior) {
	if from.AllowedMethods != nil {
		if in.AllowedMethods == nil {
			in.AllowedMethods = &svcapitypes.AllowedMethods{}
		}

		in.AllowedMethods.Items = pointer.LateInitializeSlice(in.AllowedMethods.Items, from.AllowedMethods.Items)

		if from.AllowedMethods.CachedMethods != nil {
			if in.AllowedMethods.CachedMethods == nil {
				in.AllowedMethods.CachedMethods = &svcapitypes.CachedMethods{}
			}

			in.AllowedMethods.CachedMethods.Items = pointer.LateInitializeSlice(in.AllowedMethods.CachedMethods.Items, from.AllowedMethods.CachedMethods.Items)
		}
	}

	in.CachePolicyID = pointer.LateInitialize(in.CachePolicyID, from.CachePolicyId)
	in.Compress = pointer.LateInitialize(in.Compress, from.Compress)
	in.DefaultTTL = pointer.LateInitialize(in.DefaultTTL, from.DefaultTTL)
	in.FieldLevelEncryptionID = pointer.LateInitialize(in.FieldLevelEncryptionID, from.FieldLevelEncryptionId)

	if from.ForwardedValues != nil {
		if in.ForwardedValues == nil {
			in.ForwardedValues = &svcapitypes.ForwardedValues{}
		}

		if from.ForwardedValues.Cookies != nil {
			if in.ForwardedValues.Cookies == nil {
				in.ForwardedValues.Cookies = &svcapitypes.CookiePreference{}
			}

			in.ForwardedValues.Cookies.Forward = pointer.LateInitialize(in.ForwardedValues.Cookies.Forward, from.ForwardedValues.Cookies.Forward)

			if from.ForwardedValues.Cookies.WhitelistedNames != nil {
				if in.ForwardedValues.Cookies.WhitelistedNames == nil {
					in.ForwardedValues.Cookies.WhitelistedNames = &svcapitypes.CookieNames{}
				}

				in.ForwardedValues.Cookies.WhitelistedNames.Items = pointer.LateInitializeSlice(in.ForwardedValues.Cookies.WhitelistedNames.Items, from.ForwardedValues.Cookies.WhitelistedNames.Items)
			}
		}

		if from.ForwardedValues.Headers != nil {
			if in.ForwardedValues.Headers == nil {
				in.ForwardedValues.Headers = &svcapitypes.Headers{}
			}

			in.ForwardedValues.Headers.Items = pointer.LateInitializeSlice(in.ForwardedValues.Headers.Items, from.ForwardedValues.Headers.Items)
		}

		in.ForwardedValues.QueryString = pointer.LateInitialize(in.ForwardedValues.QueryString, from.ForwardedValues.QueryString)

		if from.ForwardedValues.QueryStringCacheKeys != nil {
			if in.ForwardedValues.QueryStringCacheKeys == nil {
				in.ForwardedValues.QueryStringCacheKeys = &svcapitypes.QueryStringCacheKeys{}
			}

			in.ForwardedValues.QueryStringCacheKeys.Items = pointer.LateInitializeSlice(in.ForwardedValues.QueryStringCacheKeys.Items, from.ForwardedValues.QueryStringCacheKeys.Items)
		}
	}

	if from.FunctionAssociations != nil {
		if in.FunctionAssociations == nil {
			in.FunctionAssociations = &svcapitypes.FunctionAssociations{}
		}
		lateInitFunctionAssociations(in.FunctionAssociations, from.FunctionAssociations)
	}

	if from.LambdaFunctionAssociations != nil {
		if in.LambdaFunctionAssociations == nil {
			in.LambdaFunctionAssociations = &svcapitypes.LambdaFunctionAssociations{}
		}
		lateInitLambdaFunctionAssociations(in.LambdaFunctionAssociations, from.LambdaFunctionAssociations)
	}

	in.MaxTTL = pointer.LateInitialize(in.MaxTTL, from.MaxTTL)
	in.MinTTL = pointer.LateInitialize(in.MinTTL, from.MinTTL)
	in.OriginRequestPolicyID = pointer.LateInitialize(in.OriginRequestPolicyID, from.OriginRequestPolicyId)
	in.RealtimeLogConfigARN = pointer.LateInitialize(in.RealtimeLogConfigARN, from.RealtimeLogConfigArn)
	in.SmoothStreaming = pointer.LateInitialize(in.SmoothStreaming, from.SmoothStreaming)
	in.ResponseHeadersPolicyID = pointer.LateInitialize(in.ResponseHeadersPolicyID, from.ResponseHeadersPolicyId)
	in.TargetOriginID = pointer.LateInitialize(in.TargetOriginID, from.TargetOriginId)

	if from.TrustedKeyGroups != nil {
		if in.TrustedKeyGroups == nil {
			in.TrustedKeyGroups = &svcapitypes.TrustedKeyGroups{}
		}

		in.TrustedKeyGroups.Enabled = pointer.LateInitialize(in.TrustedKeyGroups.Enabled, from.TrustedKeyGroups.Enabled)
		in.TrustedKeyGroups.Items = pointer.LateInitializeSlice(in.TrustedKeyGroups.Items, from.TrustedKeyGroups.Items)
	}

	if from.TrustedSigners != nil {
		if in.TrustedSigners == nil {
			in.TrustedSigners = &svcapitypes.TrustedSigners{}
		}

		in.TrustedSigners.Enabled = pointer.LateInitialize(in.TrustedSigners.Enabled, from.TrustedSigners.Enabled)
		in.TrustedSigners.Items = pointer.LateInitializeSlice(in.TrustedSigners.Items, from.TrustedSigners.Items)
	}

	in.ViewerProtocolPolicy = pointer.LateInitialize(in.ViewerProtocolPolicy, from.ViewerProtocolPolicy)
}

func lateInitCacheBehaviors(in *svcapitypes.CacheBehaviors, from *svcsdk.CacheBehaviors) {
	if len(from.Items) == 0 {
		return
	}

	// If we have no cache behaviors, late init the entire slice
	if in.Items == nil {
		in.Items = make([]*svcapitypes.CacheBehavior, len(from.Items))
		for i := range from.Items {
			in.Items[i] = &svcapitypes.CacheBehavior{}
			lateInitCacheBehavior(in.Items[i], from.Items[i])
		}

		return
	}

	// If we have some cache behaviors, we need to late init each one of them
	existing := make(map[string]*svcsdk.CacheBehavior)
	for i := range from.Items {
		o := from.Items[i]
		if o.PathPattern == nil {
			continue
		}
		// PathPattern must be unique for each CacheBehavior
		existing[pointer.StringValue(o.PathPattern)] = o
	}

	for i := range in.Items {
		ic := in.Items[i]
		if ic.PathPattern == nil {
			continue
		}

		fc := existing[pointer.StringValue(ic.PathPattern)]
		if fc == nil {
			continue
		}

		lateInitCacheBehavior(ic, fc)
	}
}

// This is _almost_ identical to lateInitDefaultCacheBehaviour, but it has an
// additional 'PathPattern' field.
func lateInitCacheBehavior(in *svcapitypes.CacheBehavior, from *svcsdk.CacheBehavior) {
	if from.AllowedMethods != nil {
		if in.AllowedMethods == nil {
			in.AllowedMethods = &svcapitypes.AllowedMethods{}
		}

		in.AllowedMethods.Items = pointer.LateInitializeSlice(in.AllowedMethods.Items, from.AllowedMethods.Items)

		if from.AllowedMethods.CachedMethods != nil {
			if in.AllowedMethods.CachedMethods == nil {
				in.AllowedMethods.CachedMethods = &svcapitypes.CachedMethods{}
			}

			in.AllowedMethods.CachedMethods.Items = pointer.LateInitializeSlice(in.AllowedMethods.CachedMethods.Items, from.AllowedMethods.CachedMethods.Items)
		}
	}

	in.CachePolicyID = pointer.LateInitialize(in.CachePolicyID, from.CachePolicyId)
	in.Compress = pointer.LateInitialize(in.Compress, from.Compress)
	in.DefaultTTL = pointer.LateInitialize(in.DefaultTTL, from.DefaultTTL)
	in.FieldLevelEncryptionID = pointer.LateInitialize(in.FieldLevelEncryptionID, from.FieldLevelEncryptionId)

	if from.ForwardedValues != nil {
		if in.ForwardedValues == nil {
			in.ForwardedValues = &svcapitypes.ForwardedValues{}
		}

		if from.ForwardedValues.Cookies != nil {
			if in.ForwardedValues.Cookies == nil {
				in.ForwardedValues.Cookies = &svcapitypes.CookiePreference{}
			}

			in.ForwardedValues.Cookies.Forward = pointer.LateInitialize(in.ForwardedValues.Cookies.Forward, from.ForwardedValues.Cookies.Forward)

			if from.ForwardedValues.Cookies.WhitelistedNames != nil {
				if in.ForwardedValues.Cookies.WhitelistedNames == nil {
					in.ForwardedValues.Cookies.WhitelistedNames = &svcapitypes.CookieNames{}
				}

				in.ForwardedValues.Cookies.WhitelistedNames.Items = pointer.LateInitializeSlice(in.ForwardedValues.Cookies.WhitelistedNames.Items, from.ForwardedValues.Cookies.WhitelistedNames.Items)
			}
		}

		if from.ForwardedValues.Headers != nil {
			if in.ForwardedValues.Headers == nil {
				in.ForwardedValues.Headers = &svcapitypes.Headers{}
			}

			in.ForwardedValues.Headers.Items = pointer.LateInitializeSlice(in.ForwardedValues.Headers.Items, from.ForwardedValues.Headers.Items)
		}

		in.ForwardedValues.QueryString = pointer.LateInitialize(in.ForwardedValues.QueryString, from.ForwardedValues.QueryString)

		if from.ForwardedValues.QueryStringCacheKeys != nil {
			if in.ForwardedValues.QueryStringCacheKeys == nil {
				in.ForwardedValues.QueryStringCacheKeys = &svcapitypes.QueryStringCacheKeys{}
			}

			in.ForwardedValues.QueryStringCacheKeys.Items = pointer.LateInitializeSlice(in.ForwardedValues.QueryStringCacheKeys.Items, from.ForwardedValues.QueryStringCacheKeys.Items)
		}
	}

	if from.LambdaFunctionAssociations != nil {
		if in.LambdaFunctionAssociations == nil {
			in.LambdaFunctionAssociations = &svcapitypes.LambdaFunctionAssociations{}
		}

		lateInitLambdaFunctionAssociations(in.LambdaFunctionAssociations, from.LambdaFunctionAssociations)
	}

	in.MaxTTL = pointer.LateInitialize(in.MaxTTL, from.MaxTTL)
	in.MinTTL = pointer.LateInitialize(in.MinTTL, from.MinTTL)
	in.OriginRequestPolicyID = pointer.LateInitialize(in.OriginRequestPolicyID, from.OriginRequestPolicyId)
	in.PathPattern = pointer.LateInitialize(in.PathPattern, from.PathPattern)
	in.RealtimeLogConfigARN = pointer.LateInitialize(in.RealtimeLogConfigARN, from.RealtimeLogConfigArn)
	in.SmoothStreaming = pointer.LateInitialize(in.SmoothStreaming, from.SmoothStreaming)
	in.TargetOriginID = pointer.LateInitialize(in.TargetOriginID, from.TargetOriginId)
	in.ResponseHeadersPolicyID = pointer.LateInitialize(in.ResponseHeadersPolicyID, from.ResponseHeadersPolicyId)

	if from.TrustedKeyGroups != nil && *from.TrustedKeyGroups.Enabled && len(from.TrustedKeyGroups.Items) != 0 {
		if in.TrustedKeyGroups == nil {
			in.TrustedKeyGroups = &svcapitypes.TrustedKeyGroups{}
		}

		in.TrustedKeyGroups.Enabled = pointer.LateInitialize(in.TrustedKeyGroups.Enabled, from.TrustedKeyGroups.Enabled)
		in.TrustedKeyGroups.Items = pointer.LateInitializeSlice(in.TrustedKeyGroups.Items, from.TrustedKeyGroups.Items)
	}

	if from.TrustedSigners != nil && *from.TrustedSigners.Enabled && len(from.TrustedSigners.Items) != 0 {
		if in.TrustedSigners == nil {
			in.TrustedSigners = &svcapitypes.TrustedSigners{}
		}

		in.TrustedSigners.Enabled = pointer.LateInitialize(in.TrustedSigners.Enabled, from.TrustedSigners.Enabled)
		in.TrustedSigners.Items = pointer.LateInitializeSlice(in.TrustedSigners.Items, from.TrustedSigners.Items)
	}

	in.ViewerProtocolPolicy = pointer.LateInitialize(in.ViewerProtocolPolicy, from.ViewerProtocolPolicy)
}

func lateInitCustomErrorResponses(in *svcapitypes.CustomErrorResponses, from *svcsdk.CustomErrorResponses) {
	if len(from.Items) == 0 {
		return
	}

	// If we have no custom error responses, late init the entire slice.
	if in.Items == nil {
		in.Items = make([]*svcapitypes.CustomErrorResponse, len(from.Items))
		for i := range from.Items {
			in.Items[i] = &svcapitypes.CustomErrorResponse{}
			lateInitCustomErrorResponse(in.Items[i], from.Items[i])
		}

		return
	}

	// If we have some custom error responses, we need to late init each one them
	existing := make(map[int64]*svcsdk.CustomErrorResponse)
	for i := range from.Items {
		o := from.Items[i]
		if o.ErrorCode == nil {
			continue
		}
		// ErrorCode must be unique for each CustomErrorResponse
		existing[pointer.Int64Value(o.ErrorCode)] = o
	}

	for i := range in.Items {
		ie := in.Items[i]
		if ie.ErrorCode == nil {
			continue
		}

		fe := existing[pointer.Int64Value(ie.ErrorCode)]
		if fe == nil {
			continue
		}

		lateInitCustomErrorResponse(ie, fe)
	}
}

func lateInitCustomErrorResponse(in *svcapitypes.CustomErrorResponse, from *svcsdk.CustomErrorResponse) {
	in.ErrorCachingMinTTL = pointer.LateInitialize(in.ErrorCachingMinTTL, from.ErrorCachingMinTTL)
	in.ErrorCode = pointer.LateInitialize(in.ErrorCode, from.ErrorCode)
	in.ResponseCode = pointer.LateInitialize(in.ResponseCode, from.ResponseCode)
	in.ResponsePagePath = pointer.LateInitialize(in.ResponsePagePath, from.ResponsePagePath)
}

func lateInitOriginGroups(in *svcapitypes.OriginGroups, from *svcsdk.OriginGroups) {
	if len(from.Items) == 0 {
		return
	}

	// If we have no origin groups, late init the entire slice.
	if in.Items == nil {
		in.Items = make([]*svcapitypes.OriginGroup, len(from.Items))
		for i := range from.Items {
			in.Items[i] = &svcapitypes.OriginGroup{}
			lateInitOriginGroup(in.Items[i], from.Items[i])
		}

		return
	}

	// If we have some origin groups, we need to late init each one them
	existing := make(map[string]*svcsdk.OriginGroup)
	for i := range from.Items {
		o := from.Items[i]
		if o.Id == nil {
			continue
		}
		existing[pointer.StringValue(o.Id)] = o
	}

	for i := range in.Items {
		io := in.Items[i]
		if io.ID == nil {
			continue
		}

		fo := existing[pointer.StringValue(io.ID)]
		if fo == nil {
			continue
		}

		lateInitOriginGroup(io, fo)
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

			in.FailoverCriteria.StatusCodes.Items = pointer.LateInitializeSlice(in.FailoverCriteria.StatusCodes.Items, from.FailoverCriteria.StatusCodes.Items)
		}
	}

	in.ID = pointer.LateInitialize(in.ID, from.Id)

	if from.Members != nil {
		if in.Members == nil {
			in.Members = &svcapitypes.OriginGroupMembers{}
		}

		lateInitOriginGroupMembers(in.Members, from.Members)
	}
}

func lateInitOriginGroupMembers(in *svcapitypes.OriginGroupMembers, from *svcsdk.OriginGroupMembers) {
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
	in.OriginID = pointer.LateInitialize(in.OriginID, from.OriginId)
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
		existing[pointer.StringValue(o.Id)] = o
	}

	for i := range in.Items {
		io := in.Items[i]
		if io.ID == nil {
			continue
		}

		fo := existing[pointer.StringValue(io.ID)]
		if fo == nil {
			continue
		}

		lateInitOrigin(io, fo)
	}
}

func lateInitOrigin(in *svcapitypes.Origin, from *svcsdk.Origin) {
	in.ConnectionAttempts = pointer.LateInitialize(in.ConnectionAttempts, from.ConnectionAttempts)
	in.ConnectionTimeout = pointer.LateInitialize(in.ConnectionTimeout, from.ConnectionTimeout)

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

		in.CustomOriginConfig.HTTPPort = pointer.LateInitialize(in.CustomOriginConfig.HTTPPort, from.CustomOriginConfig.HTTPPort)
		in.CustomOriginConfig.HTTPSPort = pointer.LateInitialize(in.CustomOriginConfig.HTTPSPort, from.CustomOriginConfig.HTTPSPort)
		in.CustomOriginConfig.OriginKeepaliveTimeout = pointer.LateInitialize(in.CustomOriginConfig.OriginKeepaliveTimeout, from.CustomOriginConfig.OriginKeepaliveTimeout)
		in.CustomOriginConfig.OriginProtocolPolicy = pointer.LateInitialize(in.CustomOriginConfig.OriginProtocolPolicy, from.CustomOriginConfig.OriginProtocolPolicy)
		in.CustomOriginConfig.OriginReadTimeout = pointer.LateInitialize(in.CustomOriginConfig.OriginReadTimeout, from.CustomOriginConfig.OriginReadTimeout)

		if from.CustomOriginConfig.OriginSslProtocols != nil {
			if in.CustomOriginConfig.OriginSSLProtocols == nil {
				in.CustomOriginConfig.OriginSSLProtocols = &svcapitypes.OriginSSLProtocols{}
			}

			in.CustomOriginConfig.OriginSSLProtocols.Items = pointer.LateInitializeSlice(in.CustomOriginConfig.OriginSSLProtocols.Items, from.CustomOriginConfig.OriginSslProtocols.Items)
		}
	}

	in.DomainName = pointer.LateInitialize(in.DomainName, from.DomainName)
	in.ID = pointer.LateInitialize(in.ID, from.Id)
	in.OriginPath = pointer.LateInitialize(in.OriginPath, from.OriginPath)

	if from.OriginShield != nil {
		if in.OriginShield == nil {
			in.OriginShield = &svcapitypes.OriginShield{}
		}

		in.OriginShield.Enabled = pointer.LateInitialize(in.OriginShield.Enabled, from.OriginShield.Enabled)
		in.OriginShield.OriginShieldRegion = pointer.LateInitialize(in.OriginShield.OriginShieldRegion, from.OriginShield.OriginShieldRegion)
	}

	if from.S3OriginConfig != nil {
		if in.S3OriginConfig == nil {
			in.S3OriginConfig = &svcapitypes.S3OriginConfig{}
		}

		in.S3OriginConfig.OriginAccessIdentity = pointer.LateInitialize(in.S3OriginConfig.OriginAccessIdentity, from.S3OriginConfig.OriginAccessIdentity)
	}
}

func lateInitOriginCustomHeaders(in *svcapitypes.CustomHeaders, from *svcsdk.CustomHeaders) {
	if len(from.Items) == 0 {
		return
	}

	in.Items = make([]*svcapitypes.OriginCustomHeader, len(from.Items))
	for i := range from.Items {
		in.Items[i] = &svcapitypes.OriginCustomHeader{}
		lateInitOriginCustomHeader(in.Items[i], from.Items[i])
	}

	// If we have some origin custom headers, we need to late init each one of them
	existing := make(map[string]*svcsdk.OriginCustomHeader)
	for i := range from.Items {
		o := from.Items[i]
		if o.HeaderName == nil {
			continue
		}
		// HeaderName must be unique for each OriginCustomHeader
		existing[pointer.StringValue(o.HeaderName)] = o
	}

	for i := range in.Items {
		ih := in.Items[i]
		if ih.HeaderName == nil {
			continue
		}

		fh := existing[pointer.StringValue(ih.HeaderName)]
		if fh == nil {
			continue
		}

		lateInitOriginCustomHeader(ih, fh)
	}
}

func lateInitOriginCustomHeader(in *svcapitypes.OriginCustomHeader, from *svcsdk.OriginCustomHeader) {
	in.HeaderName = pointer.LateInitialize(in.HeaderName, from.HeaderName)
	in.HeaderValue = pointer.LateInitialize(in.HeaderValue, from.HeaderValue)
}

func lateInitLambdaFunctionAssociations(in *svcapitypes.LambdaFunctionAssociations, from *svcsdk.LambdaFunctionAssociations) {
	if len(from.Items) == 0 {
		return
	}

	// If we have no lambda function associations, late init the entire slice
	if in.Items == nil {
		in.Items = make([]*svcapitypes.LambdaFunctionAssociation, len(from.Items))
		for i := range from.Items {
			in.Items[i] = &svcapitypes.LambdaFunctionAssociation{}
			lateInitLambdaFunctionAssociation(in.Items[i], from.Items[i])
		}

		return
	}
	// If we have some lambda function associations, we need to late init each one of them
	existing := make(map[string]*svcsdk.LambdaFunctionAssociation)
	for i := range from.Items {
		o := from.Items[i]
		if o.LambdaFunctionARN == nil {
			continue
		}
		// TODO(ezgidemirel): Instead of using FunctionARNs, we should use EventTypes as keys
		// LambdaFunctionARN must be unique for each LambdaFunctionAssociation
		existing[pointer.StringValue(o.LambdaFunctionARN)] = o
	}

	for i := range in.Items {
		il := in.Items[i]
		if il.LambdaFunctionARN == nil {
			continue
		}

		fl := existing[pointer.StringValue(il.LambdaFunctionARN)]
		if fl == nil {
			continue
		}

		lateInitLambdaFunctionAssociation(il, fl)
	}
}

func lateInitLambdaFunctionAssociation(in *svcapitypes.LambdaFunctionAssociation, from *svcsdk.LambdaFunctionAssociation) {
	in.EventType = pointer.LateInitialize(in.EventType, from.EventType)
	in.IncludeBody = pointer.LateInitialize(in.IncludeBody, from.IncludeBody)
	in.LambdaFunctionARN = pointer.LateInitialize(in.LambdaFunctionARN, from.LambdaFunctionARN)
}

func lateInitFunctionAssociations(in *svcapitypes.FunctionAssociations, from *svcsdk.FunctionAssociations) {
	if len(from.Items) == 0 {
		return
	}

	// If we have no function associations, late init the entire slice
	if in.Items == nil {
		in.Items = make([]*svcapitypes.FunctionAssociation, len(from.Items))
		for i := range from.Items {
			in.Items[i] = &svcapitypes.FunctionAssociation{}
			lateInitFunctionAssociation(in.Items[i], from.Items[i])
		}

		return
	}

	// If we have some function associations, we need to late init each one of them
	existing := make(map[string]*svcsdk.FunctionAssociation)
	for _, o := range from.Items {
		if o.EventType == nil {
			continue
		}
		// AWS Console allows us to set a single FunctionARN for each predefined EventType
		existing[pointer.StringValue(o.EventType)] = o
	}

	for _, il := range in.Items {
		// If EventType is not nil, we want to use the value coming from the input
		if il.EventType != nil {
			continue
		}

		fl := existing[pointer.StringValue(il.EventType)]
		if fl == nil {
			continue
		}

		lateInitFunctionAssociation(il, fl)
	}
}

func lateInitFunctionAssociation(in *svcapitypes.FunctionAssociation, from *svcsdk.FunctionAssociation) {
	in.EventType = pointer.LateInitialize(in.EventType, from.EventType)
	in.FunctionARN = pointer.LateInitialize(in.FunctionARN, from.FunctionARN)
}
