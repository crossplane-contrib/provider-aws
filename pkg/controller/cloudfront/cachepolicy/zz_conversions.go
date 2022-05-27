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

// Code generated by ack-generate. DO NOT EDIT.

package cachepolicy

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	svcsdk "github.com/aws/aws-sdk-go/service/cloudfront"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/cloudfront/v1alpha1"
)

// NOTE(muvaf): We return pointers in case the function needs to start with an
// empty object, hence need to return a new pointer.

// GenerateGetCachePolicyInput returns input for read
// operation.
func GenerateGetCachePolicyInput(cr *svcapitypes.CachePolicy) *svcsdk.GetCachePolicyInput {
	res := &svcsdk.GetCachePolicyInput{}

	return res
}

// GenerateCachePolicy returns the current state in the form of *svcapitypes.CachePolicy.
func GenerateCachePolicy(resp *svcsdk.GetCachePolicyOutput) *svcapitypes.CachePolicy {
	cr := &svcapitypes.CachePolicy{}

	if resp.CachePolicy != nil {
		f0 := &svcapitypes.CachePolicy_SDK{}
		if resp.CachePolicy.CachePolicyConfig != nil {
			f0f0 := &svcapitypes.CachePolicyConfig{}
			if resp.CachePolicy.CachePolicyConfig.Comment != nil {
				f0f0.Comment = resp.CachePolicy.CachePolicyConfig.Comment
			}
			if resp.CachePolicy.CachePolicyConfig.DefaultTTL != nil {
				f0f0.DefaultTTL = resp.CachePolicy.CachePolicyConfig.DefaultTTL
			}
			if resp.CachePolicy.CachePolicyConfig.MaxTTL != nil {
				f0f0.MaxTTL = resp.CachePolicy.CachePolicyConfig.MaxTTL
			}
			if resp.CachePolicy.CachePolicyConfig.MinTTL != nil {
				f0f0.MinTTL = resp.CachePolicy.CachePolicyConfig.MinTTL
			}
			if resp.CachePolicy.CachePolicyConfig.Name != nil {
				f0f0.Name = resp.CachePolicy.CachePolicyConfig.Name
			}
			if resp.CachePolicy.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin != nil {
				f0f0f5 := &svcapitypes.ParametersInCacheKeyAndForwardedToOrigin{}
				if resp.CachePolicy.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.CookiesConfig != nil {
					f0f0f5f0 := &svcapitypes.CachePolicyCookiesConfig{}
					if resp.CachePolicy.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.CookiesConfig.CookieBehavior != nil {
						f0f0f5f0.CookieBehavior = resp.CachePolicy.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.CookiesConfig.CookieBehavior
					}
					if resp.CachePolicy.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.CookiesConfig.Cookies != nil {
						f0f0f5f0f1 := &svcapitypes.CookieNames{}
						if resp.CachePolicy.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.CookiesConfig.Cookies.Items != nil {
							f0f0f5f0f1f0 := []*string{}
							for _, f0f0f5f0f1f0iter := range resp.CachePolicy.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.CookiesConfig.Cookies.Items {
								var f0f0f5f0f1f0elem string
								f0f0f5f0f1f0elem = *f0f0f5f0f1f0iter
								f0f0f5f0f1f0 = append(f0f0f5f0f1f0, &f0f0f5f0f1f0elem)
							}
							f0f0f5f0f1.Items = f0f0f5f0f1f0
						}
						if resp.CachePolicy.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.CookiesConfig.Cookies.Quantity != nil {
							f0f0f5f0f1.Quantity = resp.CachePolicy.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.CookiesConfig.Cookies.Quantity
						}
						f0f0f5f0.Cookies = f0f0f5f0f1
					}
					f0f0f5.CookiesConfig = f0f0f5f0
				}
				if resp.CachePolicy.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.EnableAcceptEncodingBrotli != nil {
					f0f0f5.EnableAcceptEncodingBrotli = resp.CachePolicy.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.EnableAcceptEncodingBrotli
				}
				if resp.CachePolicy.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.EnableAcceptEncodingGzip != nil {
					f0f0f5.EnableAcceptEncodingGzip = resp.CachePolicy.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.EnableAcceptEncodingGzip
				}
				if resp.CachePolicy.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.HeadersConfig != nil {
					f0f0f5f3 := &svcapitypes.CachePolicyHeadersConfig{}
					if resp.CachePolicy.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.HeadersConfig.HeaderBehavior != nil {
						f0f0f5f3.HeaderBehavior = resp.CachePolicy.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.HeadersConfig.HeaderBehavior
					}
					if resp.CachePolicy.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.HeadersConfig.Headers != nil {
						f0f0f5f3f1 := &svcapitypes.Headers{}
						if resp.CachePolicy.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.HeadersConfig.Headers.Items != nil {
							f0f0f5f3f1f0 := []*string{}
							for _, f0f0f5f3f1f0iter := range resp.CachePolicy.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.HeadersConfig.Headers.Items {
								var f0f0f5f3f1f0elem string
								f0f0f5f3f1f0elem = *f0f0f5f3f1f0iter
								f0f0f5f3f1f0 = append(f0f0f5f3f1f0, &f0f0f5f3f1f0elem)
							}
							f0f0f5f3f1.Items = f0f0f5f3f1f0
						}
						f0f0f5f3.Headers = f0f0f5f3f1
					}
					f0f0f5.HeadersConfig = f0f0f5f3
				}
				if resp.CachePolicy.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.QueryStringsConfig != nil {
					f0f0f5f4 := &svcapitypes.CachePolicyQueryStringsConfig{}
					if resp.CachePolicy.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.QueryStringsConfig.QueryStringBehavior != nil {
						f0f0f5f4.QueryStringBehavior = resp.CachePolicy.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.QueryStringsConfig.QueryStringBehavior
					}
					if resp.CachePolicy.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.QueryStringsConfig.QueryStrings != nil {
						f0f0f5f4f1 := &svcapitypes.QueryStringNames{}
						if resp.CachePolicy.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.QueryStringsConfig.QueryStrings.Items != nil {
							f0f0f5f4f1f0 := []*string{}
							for _, f0f0f5f4f1f0iter := range resp.CachePolicy.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.QueryStringsConfig.QueryStrings.Items {
								var f0f0f5f4f1f0elem string
								f0f0f5f4f1f0elem = *f0f0f5f4f1f0iter
								f0f0f5f4f1f0 = append(f0f0f5f4f1f0, &f0f0f5f4f1f0elem)
							}
							f0f0f5f4f1.Items = f0f0f5f4f1f0
						}
						if resp.CachePolicy.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.QueryStringsConfig.QueryStrings.Quantity != nil {
							f0f0f5f4f1.Quantity = resp.CachePolicy.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.QueryStringsConfig.QueryStrings.Quantity
						}
						f0f0f5f4.QueryStrings = f0f0f5f4f1
					}
					f0f0f5.QueryStringsConfig = f0f0f5f4
				}
				f0f0.ParametersInCacheKeyAndForwardedToOrigin = f0f0f5
			}
			f0.CachePolicyConfig = f0f0
		}
		if resp.CachePolicy.Id != nil {
			f0.ID = resp.CachePolicy.Id
		}
		if resp.CachePolicy.LastModifiedTime != nil {
			f0.LastModifiedTime = &metav1.Time{*resp.CachePolicy.LastModifiedTime}
		}
		cr.Status.AtProvider.CachePolicy = f0
	} else {
		cr.Status.AtProvider.CachePolicy = nil
	}
	if resp.ETag != nil {
		cr.Status.AtProvider.ETag = resp.ETag
	} else {
		cr.Status.AtProvider.ETag = nil
	}

	return cr
}

// GenerateCreateCachePolicyInput returns a create input.
func GenerateCreateCachePolicyInput(cr *svcapitypes.CachePolicy) *svcsdk.CreateCachePolicyInput {
	res := &svcsdk.CreateCachePolicyInput{}

	if cr.Spec.ForProvider.CachePolicyConfig != nil {
		f0 := &svcsdk.CachePolicyConfig{}
		if cr.Spec.ForProvider.CachePolicyConfig.Comment != nil {
			f0.SetComment(*cr.Spec.ForProvider.CachePolicyConfig.Comment)
		}
		if cr.Spec.ForProvider.CachePolicyConfig.DefaultTTL != nil {
			f0.SetDefaultTTL(*cr.Spec.ForProvider.CachePolicyConfig.DefaultTTL)
		}
		if cr.Spec.ForProvider.CachePolicyConfig.MaxTTL != nil {
			f0.SetMaxTTL(*cr.Spec.ForProvider.CachePolicyConfig.MaxTTL)
		}
		if cr.Spec.ForProvider.CachePolicyConfig.MinTTL != nil {
			f0.SetMinTTL(*cr.Spec.ForProvider.CachePolicyConfig.MinTTL)
		}
		if cr.Spec.ForProvider.CachePolicyConfig.Name != nil {
			f0.SetName(*cr.Spec.ForProvider.CachePolicyConfig.Name)
		}
		if cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin != nil {
			f0f5 := &svcsdk.ParametersInCacheKeyAndForwardedToOrigin{}
			if cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.CookiesConfig != nil {
				f0f5f0 := &svcsdk.CachePolicyCookiesConfig{}
				if cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.CookiesConfig.CookieBehavior != nil {
					f0f5f0.SetCookieBehavior(*cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.CookiesConfig.CookieBehavior)
				}
				if cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.CookiesConfig.Cookies != nil {
					f0f5f0f1 := &svcsdk.CookieNames{}
					if cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.CookiesConfig.Cookies.Items != nil {
						f0f5f0f1f0 := []*string{}
						for _, f0f5f0f1f0iter := range cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.CookiesConfig.Cookies.Items {
							var f0f5f0f1f0elem string
							f0f5f0f1f0elem = *f0f5f0f1f0iter
							f0f5f0f1f0 = append(f0f5f0f1f0, &f0f5f0f1f0elem)
						}
						f0f5f0f1.SetItems(f0f5f0f1f0)
					}
					if cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.CookiesConfig.Cookies.Quantity != nil {
						f0f5f0f1.SetQuantity(*cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.CookiesConfig.Cookies.Quantity)
					}
					f0f5f0.SetCookies(f0f5f0f1)
				}
				f0f5.SetCookiesConfig(f0f5f0)
			}
			if cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.EnableAcceptEncodingBrotli != nil {
				f0f5.SetEnableAcceptEncodingBrotli(*cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.EnableAcceptEncodingBrotli)
			}
			if cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.EnableAcceptEncodingGzip != nil {
				f0f5.SetEnableAcceptEncodingGzip(*cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.EnableAcceptEncodingGzip)
			}
			if cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.HeadersConfig != nil {
				f0f5f3 := &svcsdk.CachePolicyHeadersConfig{}
				if cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.HeadersConfig.HeaderBehavior != nil {
					f0f5f3.SetHeaderBehavior(*cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.HeadersConfig.HeaderBehavior)
				}
				if cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.HeadersConfig.Headers != nil {
					f0f5f3f1 := &svcsdk.Headers{}
					if cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.HeadersConfig.Headers.Items != nil {
						f0f5f3f1f0 := []*string{}
						for _, f0f5f3f1f0iter := range cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.HeadersConfig.Headers.Items {
							var f0f5f3f1f0elem string
							f0f5f3f1f0elem = *f0f5f3f1f0iter
							f0f5f3f1f0 = append(f0f5f3f1f0, &f0f5f3f1f0elem)
						}
						f0f5f3f1.SetItems(f0f5f3f1f0)
					}
					f0f5f3.SetHeaders(f0f5f3f1)
				}
				f0f5.SetHeadersConfig(f0f5f3)
			}
			if cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.QueryStringsConfig != nil {
				f0f5f4 := &svcsdk.CachePolicyQueryStringsConfig{}
				if cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.QueryStringsConfig.QueryStringBehavior != nil {
					f0f5f4.SetQueryStringBehavior(*cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.QueryStringsConfig.QueryStringBehavior)
				}
				if cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.QueryStringsConfig.QueryStrings != nil {
					f0f5f4f1 := &svcsdk.QueryStringNames{}
					if cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.QueryStringsConfig.QueryStrings.Items != nil {
						f0f5f4f1f0 := []*string{}
						for _, f0f5f4f1f0iter := range cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.QueryStringsConfig.QueryStrings.Items {
							var f0f5f4f1f0elem string
							f0f5f4f1f0elem = *f0f5f4f1f0iter
							f0f5f4f1f0 = append(f0f5f4f1f0, &f0f5f4f1f0elem)
						}
						f0f5f4f1.SetItems(f0f5f4f1f0)
					}
					if cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.QueryStringsConfig.QueryStrings.Quantity != nil {
						f0f5f4f1.SetQuantity(*cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.QueryStringsConfig.QueryStrings.Quantity)
					}
					f0f5f4.SetQueryStrings(f0f5f4f1)
				}
				f0f5.SetQueryStringsConfig(f0f5f4)
			}
			f0.SetParametersInCacheKeyAndForwardedToOrigin(f0f5)
		}
		res.SetCachePolicyConfig(f0)
	}

	return res
}

// GenerateUpdateCachePolicyInput returns an update input.
func GenerateUpdateCachePolicyInput(cr *svcapitypes.CachePolicy) *svcsdk.UpdateCachePolicyInput {
	res := &svcsdk.UpdateCachePolicyInput{}

	if cr.Spec.ForProvider.CachePolicyConfig != nil {
		f0 := &svcsdk.CachePolicyConfig{}
		if cr.Spec.ForProvider.CachePolicyConfig.Comment != nil {
			f0.SetComment(*cr.Spec.ForProvider.CachePolicyConfig.Comment)
		}
		if cr.Spec.ForProvider.CachePolicyConfig.DefaultTTL != nil {
			f0.SetDefaultTTL(*cr.Spec.ForProvider.CachePolicyConfig.DefaultTTL)
		}
		if cr.Spec.ForProvider.CachePolicyConfig.MaxTTL != nil {
			f0.SetMaxTTL(*cr.Spec.ForProvider.CachePolicyConfig.MaxTTL)
		}
		if cr.Spec.ForProvider.CachePolicyConfig.MinTTL != nil {
			f0.SetMinTTL(*cr.Spec.ForProvider.CachePolicyConfig.MinTTL)
		}
		if cr.Spec.ForProvider.CachePolicyConfig.Name != nil {
			f0.SetName(*cr.Spec.ForProvider.CachePolicyConfig.Name)
		}
		if cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin != nil {
			f0f5 := &svcsdk.ParametersInCacheKeyAndForwardedToOrigin{}
			if cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.CookiesConfig != nil {
				f0f5f0 := &svcsdk.CachePolicyCookiesConfig{}
				if cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.CookiesConfig.CookieBehavior != nil {
					f0f5f0.SetCookieBehavior(*cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.CookiesConfig.CookieBehavior)
				}
				if cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.CookiesConfig.Cookies != nil {
					f0f5f0f1 := &svcsdk.CookieNames{}
					if cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.CookiesConfig.Cookies.Items != nil {
						f0f5f0f1f0 := []*string{}
						for _, f0f5f0f1f0iter := range cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.CookiesConfig.Cookies.Items {
							var f0f5f0f1f0elem string
							f0f5f0f1f0elem = *f0f5f0f1f0iter
							f0f5f0f1f0 = append(f0f5f0f1f0, &f0f5f0f1f0elem)
						}
						f0f5f0f1.SetItems(f0f5f0f1f0)
					}
					if cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.CookiesConfig.Cookies.Quantity != nil {
						f0f5f0f1.SetQuantity(*cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.CookiesConfig.Cookies.Quantity)
					}
					f0f5f0.SetCookies(f0f5f0f1)
				}
				f0f5.SetCookiesConfig(f0f5f0)
			}
			if cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.EnableAcceptEncodingBrotli != nil {
				f0f5.SetEnableAcceptEncodingBrotli(*cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.EnableAcceptEncodingBrotli)
			}
			if cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.EnableAcceptEncodingGzip != nil {
				f0f5.SetEnableAcceptEncodingGzip(*cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.EnableAcceptEncodingGzip)
			}
			if cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.HeadersConfig != nil {
				f0f5f3 := &svcsdk.CachePolicyHeadersConfig{}
				if cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.HeadersConfig.HeaderBehavior != nil {
					f0f5f3.SetHeaderBehavior(*cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.HeadersConfig.HeaderBehavior)
				}
				if cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.HeadersConfig.Headers != nil {
					f0f5f3f1 := &svcsdk.Headers{}
					if cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.HeadersConfig.Headers.Items != nil {
						f0f5f3f1f0 := []*string{}
						for _, f0f5f3f1f0iter := range cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.HeadersConfig.Headers.Items {
							var f0f5f3f1f0elem string
							f0f5f3f1f0elem = *f0f5f3f1f0iter
							f0f5f3f1f0 = append(f0f5f3f1f0, &f0f5f3f1f0elem)
						}
						f0f5f3f1.SetItems(f0f5f3f1f0)
					}
					f0f5f3.SetHeaders(f0f5f3f1)
				}
				f0f5.SetHeadersConfig(f0f5f3)
			}
			if cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.QueryStringsConfig != nil {
				f0f5f4 := &svcsdk.CachePolicyQueryStringsConfig{}
				if cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.QueryStringsConfig.QueryStringBehavior != nil {
					f0f5f4.SetQueryStringBehavior(*cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.QueryStringsConfig.QueryStringBehavior)
				}
				if cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.QueryStringsConfig.QueryStrings != nil {
					f0f5f4f1 := &svcsdk.QueryStringNames{}
					if cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.QueryStringsConfig.QueryStrings.Items != nil {
						f0f5f4f1f0 := []*string{}
						for _, f0f5f4f1f0iter := range cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.QueryStringsConfig.QueryStrings.Items {
							var f0f5f4f1f0elem string
							f0f5f4f1f0elem = *f0f5f4f1f0iter
							f0f5f4f1f0 = append(f0f5f4f1f0, &f0f5f4f1f0elem)
						}
						f0f5f4f1.SetItems(f0f5f4f1f0)
					}
					if cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.QueryStringsConfig.QueryStrings.Quantity != nil {
						f0f5f4f1.SetQuantity(*cr.Spec.ForProvider.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.QueryStringsConfig.QueryStrings.Quantity)
					}
					f0f5f4.SetQueryStrings(f0f5f4f1)
				}
				f0f5.SetQueryStringsConfig(f0f5f4)
			}
			f0.SetParametersInCacheKeyAndForwardedToOrigin(f0f5)
		}
		res.SetCachePolicyConfig(f0)
	}

	return res
}

// GenerateDeleteCachePolicyInput returns a deletion input.
func GenerateDeleteCachePolicyInput(cr *svcapitypes.CachePolicy) *svcsdk.DeleteCachePolicyInput {
	res := &svcsdk.DeleteCachePolicyInput{}

	return res
}

// IsNotFound returns whether the given error is of type NotFound or not.
func IsNotFound(err error) bool {
	awsErr, ok := err.(awserr.Error)
	return ok && awsErr.Code() == "NoSuchCachePolicy"
}
