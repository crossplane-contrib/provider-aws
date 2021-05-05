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

package bucket

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	clients3 "github.com/crossplane/provider-aws/pkg/clients/s3"
	"github.com/crossplane/provider-aws/pkg/clients/s3/fake"
	s3Testing "github.com/crossplane/provider-aws/pkg/controller/s3/testing"
)

var (
	errorObjectKey                   = "errorKey"
	indexSuffix                      = "suffix"
	hostname                         = "web-hostname"
	webProtocol                      = "https"
	errorCode                        = "httpErrorCode"
	keyPrefix                        = "condition-key-prefix"
	httpRedirect                     = "http-redirect-code"
	replacePrefix                    = "replace-prefix-key"
	replaceKey                       = "replace-key"
	_              SubresourceClient = &WebsiteConfigurationClient{}
)

func generateWebsiteConfig() *v1beta1.WebsiteConfiguration {
	return &v1beta1.WebsiteConfiguration{
		ErrorDocument: &v1beta1.ErrorDocument{Key: errorObjectKey},
		IndexDocument: &v1beta1.IndexDocument{Suffix: indexSuffix},
		RedirectAllRequestsTo: &v1beta1.RedirectAllRequestsTo{
			HostName: hostname,
			Protocol: webProtocol,
		},
		RoutingRules: []v1beta1.RoutingRule{
			{
				Condition: &v1beta1.Condition{
					HTTPErrorCodeReturnedEquals: &errorCode,
					KeyPrefixEquals:             &keyPrefix,
				},
				Redirect: v1beta1.Redirect{
					HostName:             &hostname,
					HTTPRedirectCode:     &httpRedirect,
					Protocol:             webProtocol,
					ReplaceKeyPrefixWith: &replacePrefix,
					ReplaceKeyWith:       &replaceKey,
				},
			},
		},
	}
}

func generateAWSWebsite() *s3.WebsiteConfiguration {
	return &s3.WebsiteConfiguration{
		ErrorDocument: &s3.ErrorDocument{Key: &errorObjectKey},
		IndexDocument: &s3.IndexDocument{Suffix: &indexSuffix},
		RedirectAllRequestsTo: &s3.RedirectAllRequestsTo{
			HostName: &hostname,
			Protocol: s3.ProtocolHttps,
		},
		RoutingRules: []s3.RoutingRule{
			{
				Condition: &s3.Condition{
					HttpErrorCodeReturnedEquals: &errorCode,
					KeyPrefixEquals:             &keyPrefix,
				},
				Redirect: &s3.Redirect{
					HostName:             &hostname,
					HttpRedirectCode:     &httpRedirect,
					Protocol:             s3.ProtocolHttps,
					ReplaceKeyPrefixWith: &replacePrefix,
					ReplaceKeyWith:       &replaceKey,
				},
			},
		},
	}
}

func TestWebsiteObserve(t *testing.T) {
	type args struct {
		cl *WebsiteConfigurationClient
		b  *v1beta1.Bucket
	}

	type want struct {
		status ResourceStatus
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Error": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithWebConfig(generateWebsiteConfig())),
				cl: NewWebsiteConfigurationClient(fake.MockBucketClient{
					MockGetBucketWebsiteRequest: func(input *s3.GetBucketWebsiteInput) s3.GetBucketWebsiteRequest {
						return s3.GetBucketWebsiteRequest{
							Request: s3Testing.CreateRequest(errBoom, &s3.GetBucketWebsiteOutput{}),
						}
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
				err:    awsclient.Wrap(errBoom, websiteGetFailed),
			},
		},
		"UpdateNeededFull": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithWebConfig(generateWebsiteConfig())),
				cl: NewWebsiteConfigurationClient(fake.MockBucketClient{
					MockGetBucketWebsiteRequest: func(input *s3.GetBucketWebsiteInput) s3.GetBucketWebsiteRequest {
						return s3.GetBucketWebsiteRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketWebsiteOutput{}),
						}
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
				err:    nil,
			},
		},
		"UpdateNeededPartial": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithWebConfig(generateWebsiteConfig())),
				cl: NewWebsiteConfigurationClient(fake.MockBucketClient{
					MockGetBucketWebsiteRequest: func(input *s3.GetBucketWebsiteInput) s3.GetBucketWebsiteRequest {
						return s3.GetBucketWebsiteRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketWebsiteOutput{
								IndexDocument: generateAWSWebsite().IndexDocument,
							}),
						}
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
				err:    nil,
			},
		},
		"NeedsDelete": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithWebConfig(nil)),
				cl: NewWebsiteConfigurationClient(fake.MockBucketClient{
					MockGetBucketWebsiteRequest: func(input *s3.GetBucketWebsiteInput) s3.GetBucketWebsiteRequest {
						return s3.GetBucketWebsiteRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketWebsiteOutput{
								ErrorDocument:         generateAWSWebsite().ErrorDocument,
								IndexDocument:         generateAWSWebsite().IndexDocument,
								RedirectAllRequestsTo: generateAWSWebsite().RedirectAllRequestsTo,
								RoutingRules:          generateAWSWebsite().RoutingRules,
							}),
						}
					},
				}),
			},
			want: want{
				status: NeedsDeletion,
				err:    nil,
			},
		},
		"NoUpdateNotExists": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithWebConfig(nil)),
				cl: NewWebsiteConfigurationClient(fake.MockBucketClient{
					MockGetBucketWebsiteRequest: func(input *s3.GetBucketWebsiteInput) s3.GetBucketWebsiteRequest {
						return s3.GetBucketWebsiteRequest{
							Request: s3Testing.CreateRequest(awserr.New(clients3.WebsiteNotFoundErrCode, "", nil), &s3.GetBucketWebsiteOutput{}),
						}
					},
				}),
			},
			want: want{
				status: Updated,
				err:    nil,
			},
		},
		"NoUpdateNotExistsNil": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithWebConfig(nil)),
				cl: NewWebsiteConfigurationClient(fake.MockBucketClient{
					MockGetBucketWebsiteRequest: func(input *s3.GetBucketWebsiteInput) s3.GetBucketWebsiteRequest {
						return s3.GetBucketWebsiteRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketWebsiteOutput{}),
						}
					},
				}),
			},
			want: want{
				status: Updated,
				err:    nil,
			},
		},
		"NoUpdateExists": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithWebConfig(generateWebsiteConfig())),
				cl: NewWebsiteConfigurationClient(fake.MockBucketClient{
					MockGetBucketWebsiteRequest: func(input *s3.GetBucketWebsiteInput) s3.GetBucketWebsiteRequest {
						return s3.GetBucketWebsiteRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketWebsiteOutput{
								ErrorDocument:         generateAWSWebsite().ErrorDocument,
								IndexDocument:         generateAWSWebsite().IndexDocument,
								RedirectAllRequestsTo: generateAWSWebsite().RedirectAllRequestsTo,
								RoutingRules:          generateAWSWebsite().RoutingRules,
							}),
						}
					},
				}),
			},
			want: want{
				status: Updated,
				err:    nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			status, err := tc.args.cl.Observe(context.Background(), tc.args.b)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.status, status); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestWebsiteCreateOrUpdate(t *testing.T) {
	type args struct {
		cl *WebsiteConfigurationClient
		b  *v1beta1.Bucket
	}

	type want struct {
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Error": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithWebConfig(generateWebsiteConfig())),
				cl: NewWebsiteConfigurationClient(fake.MockBucketClient{
					MockPutBucketWebsiteRequest: func(input *s3.PutBucketWebsiteInput) s3.PutBucketWebsiteRequest {
						return s3.PutBucketWebsiteRequest{
							Request: s3Testing.CreateRequest(errBoom, &s3.PutBucketWebsiteOutput{}),
						}
					},
				}),
			},
			want: want{
				err: awsclient.Wrap(errBoom, websitePutFailed),
			},
		},
		"InvalidConfig": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithWebConfig(generateWebsiteConfig())),
				cl: NewWebsiteConfigurationClient(fake.MockBucketClient{
					MockPutBucketWebsiteRequest: func(input *s3.PutBucketWebsiteInput) s3.PutBucketWebsiteRequest {
						return s3.PutBucketWebsiteRequest{
							Request: s3Testing.CreateRequest(nil, &s3.PutBucketWebsiteOutput{}),
						}
					},
				}),
			},
			want: want{
				err: nil,
			},
		},
		"SuccessfulCreate": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithWebConfig(generateWebsiteConfig())),
				cl: NewWebsiteConfigurationClient(fake.MockBucketClient{
					MockPutBucketWebsiteRequest: func(input *s3.PutBucketWebsiteInput) s3.PutBucketWebsiteRequest {
						return s3.PutBucketWebsiteRequest{
							Request: s3Testing.CreateRequest(nil, &s3.PutBucketWebsiteOutput{}),
						}
					},
				}),
			},
			want: want{
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := tc.args.cl.CreateOrUpdate(context.Background(), tc.args.b)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestWebsiteDelete(t *testing.T) {
	type args struct {
		cl *WebsiteConfigurationClient
		b  *v1beta1.Bucket
	}

	type want struct {
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Error": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithWebConfig(generateWebsiteConfig())),
				cl: NewWebsiteConfigurationClient(fake.MockBucketClient{
					MockDeleteBucketWebsiteRequest: func(input *s3.DeleteBucketWebsiteInput) s3.DeleteBucketWebsiteRequest {
						return s3.DeleteBucketWebsiteRequest{
							Request: s3Testing.CreateRequest(errBoom, &s3.DeleteBucketWebsiteOutput{}),
						}
					},
				}),
			},
			want: want{
				err: awsclient.Wrap(errBoom, websiteDeleteFailed),
			},
		},
		"SuccessfulDelete": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithWebConfig(generateWebsiteConfig())),
				cl: NewWebsiteConfigurationClient(fake.MockBucketClient{
					MockDeleteBucketWebsiteRequest: func(input *s3.DeleteBucketWebsiteInput) s3.DeleteBucketWebsiteRequest {
						return s3.DeleteBucketWebsiteRequest{
							Request: s3Testing.CreateRequest(nil, &s3.DeleteBucketWebsiteOutput{}),
						}
					},
				}),
			},
			want: want{
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := tc.args.cl.Delete(context.Background(), tc.args.b)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestWebsiteLateInit(t *testing.T) {
	type args struct {
		cl SubresourceClient
		b  *v1beta1.Bucket
	}

	type want struct {
		err error
		cr  *v1beta1.Bucket
	}

	cases := map[string]struct {
		args
		want
	}{
		"Error": {
			args: args{
				b: s3Testing.Bucket(),
				cl: NewWebsiteConfigurationClient(fake.MockBucketClient{
					MockGetBucketWebsiteRequest: func(input *s3.GetBucketWebsiteInput) s3.GetBucketWebsiteRequest {
						return s3.GetBucketWebsiteRequest{
							Request: s3Testing.CreateRequest(errBoom, &s3.GetBucketWebsiteOutput{}),
						}
					},
				}),
			},
			want: want{
				err: awsclient.Wrap(errBoom, websiteGetFailed),
				cr:  s3Testing.Bucket(),
			},
		},
		"ErrorWebsiteConfigurationNotFound": {
			args: args{
				b: s3Testing.Bucket(),
				cl: NewWebsiteConfigurationClient(fake.MockBucketClient{
					MockGetBucketWebsiteRequest: func(input *s3.GetBucketWebsiteInput) s3.GetBucketWebsiteRequest {
						return s3.GetBucketWebsiteRequest{
							Request: s3Testing.CreateRequest(awserr.New(clients3.WebsiteNotFoundErrCode, "error", nil), &s3.GetBucketWebsiteOutput{}),
						}
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3Testing.Bucket(),
			},
		},
		"NoLateInitNil": {
			args: args{
				b: s3Testing.Bucket(),
				cl: NewWebsiteConfigurationClient(fake.MockBucketClient{
					MockGetBucketWebsiteRequest: func(input *s3.GetBucketWebsiteInput) s3.GetBucketWebsiteRequest {
						return s3.GetBucketWebsiteRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketWebsiteOutput{}),
						}
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3Testing.Bucket(),
			},
		},
		"SuccessfulLateInit": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithWebConfig(nil)),
				cl: NewWebsiteConfigurationClient(fake.MockBucketClient{
					MockGetBucketWebsiteRequest: func(input *s3.GetBucketWebsiteInput) s3.GetBucketWebsiteRequest {
						return s3.GetBucketWebsiteRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketWebsiteOutput{
								ErrorDocument:         generateAWSWebsite().ErrorDocument,
								IndexDocument:         generateAWSWebsite().IndexDocument,
								RedirectAllRequestsTo: generateAWSWebsite().RedirectAllRequestsTo,
								RoutingRules:          generateAWSWebsite().RoutingRules,
							}),
						}
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3Testing.Bucket(s3Testing.WithWebConfig(generateWebsiteConfig())),
			},
		},
		"NoOpLateInit": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithWebConfig(generateWebsiteConfig())),
				cl: NewWebsiteConfigurationClient(fake.MockBucketClient{
					MockGetBucketWebsiteRequest: func(input *s3.GetBucketWebsiteInput) s3.GetBucketWebsiteRequest {
						return s3.GetBucketWebsiteRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketWebsiteOutput{
								RedirectAllRequestsTo: generateAWSWebsite().RedirectAllRequestsTo,
								RoutingRules:          generateAWSWebsite().RoutingRules,
							}),
						}
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3Testing.Bucket(s3Testing.WithWebConfig(generateWebsiteConfig())),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := tc.args.cl.LateInitialize(context.Background(), tc.args.b)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.b, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
