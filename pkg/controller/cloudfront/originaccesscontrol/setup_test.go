/*
Copyright 2024 The Crossplane Authors.

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

package originaccesscontrol

import (
	"context"
	"testing"

	svcsdk "github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/cloudfront/v1alpha1"
)

type upToDateArgs struct {
	originAccessControl *svcapitypes.OriginAccessControl
	getOACOutput        *svcsdk.GetOriginAccessControlOutput
}

type preCreateArgs struct {
	originAccessControl *svcapitypes.OriginAccessControl
	createOACInput      *svcsdk.CreateOriginAccessControlInput
}

func TestIsUpToDate(t *testing.T) {
	region := "eu-central-2"

	resourceName := "oac-test"
	externalResource := "E0AA0AA0AA00AA"
	id := "EIDA0AA0AA00AA"

	claimName := "oac-claim-name"

	description := "Origin Access Control"
	newDescription := "New Description"

	originType := "s3"
	signingBehavior := "always"
	signingProtocol := "sigv4"

	type want struct {
		result bool
		err    error
	}
	cases := map[string]struct {
		upToDateArgs
		want
	}{
		"NothingChanged": {
			upToDateArgs: upToDateArgs{
				originAccessControl: &svcapitypes.OriginAccessControl{
					ObjectMeta: metav1.ObjectMeta{
						Name: resourceName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: externalResource,
						},
					},
					Spec: svcapitypes.OriginAccessControlSpec{
						ForProvider: svcapitypes.OriginAccessControlParameters{
							Region: region,
							OriginAccessControlConfig: &svcapitypes.OriginAccessControlConfig{
								Description:                   &description,
								Name:                          &claimName,
								OriginAccessControlOriginType: &originType,
								SigningBehavior:               &signingBehavior,
								SigningProtocol:               &signingProtocol,
							},
						},
					},
				},
				getOACOutput: &svcsdk.GetOriginAccessControlOutput{
					OriginAccessControl: &svcsdk.OriginAccessControl{
						Id: &id,
						OriginAccessControlConfig: &svcsdk.OriginAccessControlConfig{
							Description:                   &newDescription,
							Name:                          &claimName,
							OriginAccessControlOriginType: &originType,
							SigningBehavior:               &signingBehavior,
							SigningProtocol:               &signingProtocol,
						},
					},
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"SomethingChanged": {
			upToDateArgs: upToDateArgs{
				originAccessControl: &svcapitypes.OriginAccessControl{
					ObjectMeta: metav1.ObjectMeta{
						Name: resourceName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: externalResource,
						},
					},
					Spec: svcapitypes.OriginAccessControlSpec{
						ForProvider: svcapitypes.OriginAccessControlParameters{
							Region: region,
							OriginAccessControlConfig: &svcapitypes.OriginAccessControlConfig{
								Description:                   &description,
								Name:                          &claimName,
								OriginAccessControlOriginType: &originType,
								SigningBehavior:               &signingBehavior,
								SigningProtocol:               &signingProtocol,
							},
						},
					},
				},
				getOACOutput: &svcsdk.GetOriginAccessControlOutput{
					OriginAccessControl: &svcsdk.OriginAccessControl{
						Id: &id,
						OriginAccessControlConfig: &svcsdk.OriginAccessControlConfig{
							Description:                   &description,
							Name:                          &claimName,
							OriginAccessControlOriginType: &originType,
							SigningBehavior:               &signingBehavior,
							SigningProtocol:               &signingProtocol,
						},
					},
				},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			result, _, _ := isUpToDate(context.TODO(), tc.upToDateArgs.originAccessControl, tc.upToDateArgs.getOACOutput)
			if diff := cmp.Diff(tc.want.result, result); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestPreCreate(t *testing.T) {
	region := "eu-central-2"

	resourceName := "oac-test"
	externalResource := "E0AA0AA0AA00AA"

	claimName := "oac-claim-name"
	longerClaimName := "originaccesscontrol-claim-name-which-exceeds-the-64-char-number-by-some-number"

	description := "Origin Access Control"

	originType := "s3"
	wrongOriginType := "wrong-type"
	signingBehavior := "always"
	wrongSigningBehaviour := "wrong-param"
	signingProtocol := "sigv4"
	wrongSigningProtocol := "wrong-protocol"

	type want struct {
		result bool
		err    error
	}
	cases := map[string]struct {
		preCreateArgs
		want
	}{
		"GoodParams": {
			preCreateArgs: preCreateArgs{
				originAccessControl: &svcapitypes.OriginAccessControl{
					ObjectMeta: metav1.ObjectMeta{
						Name: resourceName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: externalResource,
						},
					},
					Spec: svcapitypes.OriginAccessControlSpec{
						ForProvider: svcapitypes.OriginAccessControlParameters{
							Region: region,
							OriginAccessControlConfig: &svcapitypes.OriginAccessControlConfig{
								Description:                   &description,
								Name:                          &claimName,
								OriginAccessControlOriginType: &originType,
								SigningBehavior:               &signingBehavior,
								SigningProtocol:               &signingProtocol,
							},
						},
					},
				},
				createOACInput: &svcsdk.CreateOriginAccessControlInput{
					OriginAccessControlConfig: &svcsdk.OriginAccessControlConfig{
						Description:                   &description,
						Name:                          &claimName,
						OriginAccessControlOriginType: &originType,
						SigningBehavior:               &signingBehavior,
						SigningProtocol:               &signingProtocol,
					},
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"WrongOriginType": {
			preCreateArgs: preCreateArgs{
				originAccessControl: &svcapitypes.OriginAccessControl{
					ObjectMeta: metav1.ObjectMeta{
						Name: resourceName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: externalResource,
						},
					},
					Spec: svcapitypes.OriginAccessControlSpec{
						ForProvider: svcapitypes.OriginAccessControlParameters{
							Region: region,
							OriginAccessControlConfig: &svcapitypes.OriginAccessControlConfig{
								Description:                   &claimName,
								Name:                          &claimName,
								OriginAccessControlOriginType: &wrongOriginType,
								SigningBehavior:               &signingBehavior,
								SigningProtocol:               &signingProtocol,
							},
						},
					},
				},
				createOACInput: &svcsdk.CreateOriginAccessControlInput{
					OriginAccessControlConfig: &svcsdk.OriginAccessControlConfig{
						Description:                   &description,
						Name:                          &claimName,
						OriginAccessControlOriginType: &originType,
						SigningBehavior:               &signingBehavior,
						SigningProtocol:               &signingProtocol,
					},
				},
			},
			want: want{
				result: false,
				err:    errors.New("originAccessControlOriginType invalid"),
			},
		},
		"LongerName": {
			preCreateArgs: preCreateArgs{
				originAccessControl: &svcapitypes.OriginAccessControl{
					ObjectMeta: metav1.ObjectMeta{
						Name: resourceName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: externalResource,
						},
					},
					Spec: svcapitypes.OriginAccessControlSpec{
						ForProvider: svcapitypes.OriginAccessControlParameters{
							Region: region,
							OriginAccessControlConfig: &svcapitypes.OriginAccessControlConfig{
								Description:                   &description,
								Name:                          &longerClaimName,
								OriginAccessControlOriginType: &originType,
								SigningBehavior:               &signingBehavior,
								SigningProtocol:               &signingProtocol,
							},
						},
					},
				},
				createOACInput: &svcsdk.CreateOriginAccessControlInput{
					OriginAccessControlConfig: &svcsdk.OriginAccessControlConfig{
						Description:                   &description,
						Name:                          &claimName,
						OriginAccessControlOriginType: &originType,
						SigningBehavior:               &signingBehavior,
						SigningProtocol:               &signingProtocol,
					},
				},
			},
			want: want{
				result: false,
				err:    errors.New("name is more than 64 characters"),
			},
		},
		"WrongSigningBehaviour": {
			preCreateArgs: preCreateArgs{
				originAccessControl: &svcapitypes.OriginAccessControl{
					ObjectMeta: metav1.ObjectMeta{
						Name: resourceName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: externalResource,
						},
					},
					Spec: svcapitypes.OriginAccessControlSpec{
						ForProvider: svcapitypes.OriginAccessControlParameters{
							Region: region,
							OriginAccessControlConfig: &svcapitypes.OriginAccessControlConfig{
								Description:                   &claimName,
								Name:                          &claimName,
								OriginAccessControlOriginType: &originType,
								SigningBehavior:               &wrongSigningBehaviour,
								SigningProtocol:               &signingProtocol,
							},
						},
					},
				},
				createOACInput: &svcsdk.CreateOriginAccessControlInput{
					OriginAccessControlConfig: &svcsdk.OriginAccessControlConfig{
						Description:                   &description,
						Name:                          &claimName,
						OriginAccessControlOriginType: &originType,
						SigningBehavior:               &signingBehavior,
						SigningProtocol:               &signingProtocol,
					},
				},
			},
			want: want{
				result: false,
				err:    errors.New("signingBehavior invalid"),
			},
		},
		"WrongSigningProtocol": {
			preCreateArgs: preCreateArgs{
				originAccessControl: &svcapitypes.OriginAccessControl{
					ObjectMeta: metav1.ObjectMeta{
						Name: resourceName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: externalResource,
						},
					},
					Spec: svcapitypes.OriginAccessControlSpec{
						ForProvider: svcapitypes.OriginAccessControlParameters{
							Region: region,
							OriginAccessControlConfig: &svcapitypes.OriginAccessControlConfig{
								Description:                   &claimName,
								Name:                          &claimName,
								OriginAccessControlOriginType: &originType,
								SigningBehavior:               &signingBehavior,
								SigningProtocol:               &wrongSigningProtocol,
							},
						},
					},
				},
				createOACInput: &svcsdk.CreateOriginAccessControlInput{
					OriginAccessControlConfig: &svcsdk.OriginAccessControlConfig{
						Description:                   &description,
						Name:                          &claimName,
						OriginAccessControlOriginType: &originType,
						SigningBehavior:               &signingBehavior,
						SigningProtocol:               &signingProtocol,
					},
				},
			},
			want: want{
				result: false,
				err:    errors.New("signingProtocol invalid"),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := preCreate(context.TODO(), tc.preCreateArgs.originAccessControl, tc.preCreateArgs.createOACInput)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
