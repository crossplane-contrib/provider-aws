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
	"context"

	svcapi "github.com/aws/aws-sdk-go/service/cloudfront"
	svcsdk "github.com/aws/aws-sdk-go/service/cloudfront"
	svcsdkapi "github.com/aws/aws-sdk-go/service/cloudfront/cloudfrontiface"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	cpresource "github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/cloudfront/v1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
)

const (
	errUnexpectedObject = "managed resource is not an CachePolicy resource"

	errCreateSession = "cannot create a new session"
	errCreate        = "cannot create CachePolicy in AWS"
	errUpdate        = "cannot update CachePolicy in AWS"
	errDescribe      = "failed to describe CachePolicy"
	errDelete        = "failed to delete CachePolicy"
)

type connector struct {
	kube client.Client
	opts []option
}

func (c *connector) Connect(ctx context.Context, mg cpresource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*svcapitypes.CachePolicy)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	sess, err := awsclient.GetConfigV1(ctx, c.kube, mg, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, errors.Wrap(err, errCreateSession)
	}
	return newExternal(c.kube, svcapi.New(sess), c.opts), nil
}

func (e *external) Observe(ctx context.Context, mg cpresource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*svcapitypes.CachePolicy)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}
	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}
	input := GenerateGetCachePolicyInput(cr)
	if err := e.preObserve(ctx, cr, input); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "pre-observe failed")
	}
	resp, err := e.client.GetCachePolicyWithContext(ctx, input)
	if err != nil {
		return managed.ExternalObservation{ResourceExists: false}, awsclient.Wrap(cpresource.Ignore(IsNotFound, err), errDescribe)
	}
	currentSpec := cr.Spec.ForProvider.DeepCopy()
	if err := e.lateInitialize(&cr.Spec.ForProvider, resp); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "late-init failed")
	}
	GenerateCachePolicy(resp).Status.AtProvider.DeepCopyInto(&cr.Status.AtProvider)

	upToDate, err := e.isUpToDate(cr, resp)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "isUpToDate check failed")
	}
	return e.postObserve(ctx, cr, resp, managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        upToDate,
		ResourceLateInitialized: !cmp.Equal(&cr.Spec.ForProvider, currentSpec),
	}, nil)
}

func (e *external) Create(ctx context.Context, mg cpresource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*svcapitypes.CachePolicy)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}
	cr.Status.SetConditions(xpv1.Creating())
	input := GenerateCreateCachePolicyInput(cr)
	if err := e.preCreate(ctx, cr, input); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "pre-create failed")
	}
	resp, err := e.client.CreateCachePolicyWithContext(ctx, input)
	if err != nil {
		return managed.ExternalCreation{}, awsclient.Wrap(err, errCreate)
	}

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
						if resp.CachePolicy.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.HeadersConfig.Headers.Quantity != nil {
							f0f0f5f3f1.Quantity = resp.CachePolicy.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin.HeadersConfig.Headers.Quantity
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
	if resp.Location != nil {
		cr.Status.AtProvider.Location = resp.Location
	} else {
		cr.Status.AtProvider.Location = nil
	}

	return e.postCreate(ctx, cr, resp, managed.ExternalCreation{}, err)
}

func (e *external) Update(ctx context.Context, mg cpresource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*svcapitypes.CachePolicy)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}
	input := GenerateUpdateCachePolicyInput(cr)
	if err := e.preUpdate(ctx, cr, input); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "pre-update failed")
	}
	resp, err := e.client.UpdateCachePolicyWithContext(ctx, input)
	return e.postUpdate(ctx, cr, resp, managed.ExternalUpdate{}, awsclient.Wrap(err, errUpdate))
}

func (e *external) Delete(ctx context.Context, mg cpresource.Managed) error {
	cr, ok := mg.(*svcapitypes.CachePolicy)
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	cr.Status.SetConditions(xpv1.Deleting())
	input := GenerateDeleteCachePolicyInput(cr)
	ignore, err := e.preDelete(ctx, cr, input)
	if err != nil {
		return errors.Wrap(err, "pre-delete failed")
	}
	if ignore {
		return nil
	}
	resp, err := e.client.DeleteCachePolicyWithContext(ctx, input)
	return e.postDelete(ctx, cr, resp, awsclient.Wrap(cpresource.Ignore(IsNotFound, err), errDelete))
}

type option func(*external)

func newExternal(kube client.Client, client svcsdkapi.CloudFrontAPI, opts []option) *external {
	e := &external{
		kube:           kube,
		client:         client,
		preObserve:     nopPreObserve,
		postObserve:    nopPostObserve,
		lateInitialize: nopLateInitialize,
		isUpToDate:     alwaysUpToDate,
		preCreate:      nopPreCreate,
		postCreate:     nopPostCreate,
		preDelete:      nopPreDelete,
		postDelete:     nopPostDelete,
		preUpdate:      nopPreUpdate,
		postUpdate:     nopPostUpdate,
	}
	for _, f := range opts {
		f(e)
	}
	return e
}

type external struct {
	kube           client.Client
	client         svcsdkapi.CloudFrontAPI
	preObserve     func(context.Context, *svcapitypes.CachePolicy, *svcsdk.GetCachePolicyInput) error
	postObserve    func(context.Context, *svcapitypes.CachePolicy, *svcsdk.GetCachePolicyOutput, managed.ExternalObservation, error) (managed.ExternalObservation, error)
	lateInitialize func(*svcapitypes.CachePolicyParameters, *svcsdk.GetCachePolicyOutput) error
	isUpToDate     func(*svcapitypes.CachePolicy, *svcsdk.GetCachePolicyOutput) (bool, error)
	preCreate      func(context.Context, *svcapitypes.CachePolicy, *svcsdk.CreateCachePolicyInput) error
	postCreate     func(context.Context, *svcapitypes.CachePolicy, *svcsdk.CreateCachePolicyOutput, managed.ExternalCreation, error) (managed.ExternalCreation, error)
	preDelete      func(context.Context, *svcapitypes.CachePolicy, *svcsdk.DeleteCachePolicyInput) (bool, error)
	postDelete     func(context.Context, *svcapitypes.CachePolicy, *svcsdk.DeleteCachePolicyOutput, error) error
	preUpdate      func(context.Context, *svcapitypes.CachePolicy, *svcsdk.UpdateCachePolicyInput) error
	postUpdate     func(context.Context, *svcapitypes.CachePolicy, *svcsdk.UpdateCachePolicyOutput, managed.ExternalUpdate, error) (managed.ExternalUpdate, error)
}

func nopPreObserve(context.Context, *svcapitypes.CachePolicy, *svcsdk.GetCachePolicyInput) error {
	return nil
}

func nopPostObserve(_ context.Context, _ *svcapitypes.CachePolicy, _ *svcsdk.GetCachePolicyOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	return obs, err
}
func nopLateInitialize(*svcapitypes.CachePolicyParameters, *svcsdk.GetCachePolicyOutput) error {
	return nil
}
func alwaysUpToDate(*svcapitypes.CachePolicy, *svcsdk.GetCachePolicyOutput) (bool, error) {
	return true, nil
}

func nopPreCreate(context.Context, *svcapitypes.CachePolicy, *svcsdk.CreateCachePolicyInput) error {
	return nil
}
func nopPostCreate(_ context.Context, _ *svcapitypes.CachePolicy, _ *svcsdk.CreateCachePolicyOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	return cre, err
}
func nopPreDelete(context.Context, *svcapitypes.CachePolicy, *svcsdk.DeleteCachePolicyInput) (bool, error) {
	return false, nil
}
func nopPostDelete(_ context.Context, _ *svcapitypes.CachePolicy, _ *svcsdk.DeleteCachePolicyOutput, err error) error {
	return err
}
func nopPreUpdate(context.Context, *svcapitypes.CachePolicy, *svcsdk.UpdateCachePolicyInput) error {
	return nil
}
func nopPostUpdate(_ context.Context, _ *svcapitypes.CachePolicy, _ *svcsdk.UpdateCachePolicyOutput, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	return upd, err
}
