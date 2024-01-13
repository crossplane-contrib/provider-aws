/*
Copyright 2023 The Crossplane Authors.

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

package connectaws

import (
	"context"
	"fmt"
	"os"
	"testing"

	stscreds "github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	stscredstypesv2 "github.com/aws/aws-sdk-go-v2/service/sts/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/resource/fake"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

const (
	awsCredentialsFileFormat = "[%s]\naws_access_key_id = %s\naws_secret_access_key = %s"
)

func TestCredentialsIdSecret(t *testing.T) {
	g := NewGomegaWithT(t)

	profile := "default"
	id := "testID"
	secret := "testSecret"
	token := "testtoken"
	credentials := []byte(fmt.Sprintf(awsCredentialsFileFormat, profile, id, secret))

	// valid profile
	creds, err := CredentialsIDSecret(credentials, profile)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(creds.AccessKeyID).To(Equal(id))
	g.Expect(creds.SecretAccessKey).To(Equal(secret))

	// valid profile with session token
	credentialsWithToken := []byte(fmt.Sprintf(awsCredentialsFileFormat+"\naws_session_token = %s", profile, id, secret, token))
	creds, err = CredentialsIDSecret(credentialsWithToken, profile)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(creds.AccessKeyID).To(Equal(id))
	g.Expect(creds.SecretAccessKey).To(Equal(secret))
	g.Expect(creds.SessionToken).To(Equal(token))

	// invalid profile - foo does not exist
	creds, err = CredentialsIDSecret(credentials, "foo")
	g.Expect(err).To(HaveOccurred())
	g.Expect(creds.AccessKeyID).To(Equal(""))
	g.Expect(creds.SecretAccessKey).To(Equal(""))
}

func TestUseProviderSecret(t *testing.T) {
	g := NewGomegaWithT(t)

	testProfile := "default"
	testID := "testID"
	testSecret := "testSecret"
	testRegion := "us-west-2"
	credentials := []byte(fmt.Sprintf(awsCredentialsFileFormat, testProfile, testID, testSecret))

	config, err := UseProviderSecret(context.TODO(), credentials, testProfile, testRegion)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(config).NotTo(BeNil())
}

func TestGetAssumeRoleARN(t *testing.T) {
	roleARN := "test-arn"
	roleARNDep := "test-arn-deprecated"

	type args struct {
		pcs v1beta1.ProviderConfigSpec
	}
	type want struct {
		arn string
		err error
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"NoArnSetError": {
			args: args{
				pcs: v1beta1.ProviderConfigSpec{},
			},
			want: want{
				err: errors.New("a RoleARN must be set to assume an IAM Role"),
			},
		},
		"EmptyAssumeRoleOptions": {
			args: args{
				pcs: v1beta1.ProviderConfigSpec{
					AssumeRole: &v1beta1.AssumeRoleOptions{},
				},
			},
			want: want{
				err: errors.New("a RoleARN must be set to assume an IAM Role"),
			},
		},
		"AssumeRoleOptions": {
			args: args{
				pcs: v1beta1.ProviderConfigSpec{
					AssumeRole: &v1beta1.AssumeRoleOptions{
						RoleARN: &roleARN,
					},
				},
			},
			want: want{
				arn: "test-arn",
			},
		},
		"IgnoreDeprecatedSetting": {
			args: args{
				pcs: v1beta1.ProviderConfigSpec{
					AssumeRoleARN: &roleARNDep,
					AssumeRole: &v1beta1.AssumeRoleOptions{
						RoleARN: &roleARN,
					},
				},
			},
			want: want{
				arn: "test-arn",
			},
		},
		"EmptyAssumeRoleOptionsOldSetting": {
			args: args{
				pcs: v1beta1.ProviderConfigSpec{
					AssumeRoleARN: &roleARNDep,
					AssumeRole:    &v1beta1.AssumeRoleOptions{},
				},
			},
			want: want{
				arn: "test-arn-deprecated",
			},
		},
		"DeprecatedArn": {
			args: args{
				pcs: v1beta1.ProviderConfigSpec{
					AssumeRoleARN: &roleARNDep,
				},
			},
			want: want{
				arn: "test-arn-deprecated",
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {

			roleArn, err := GetAssumeRoleARN(&tc.args.pcs)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Wrap: -want, +got:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.arn, pointer.StringValue(roleArn)); diff != "" {
				t.Errorf("Wrap: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGetAssumeRoleWithWebIdentityARN(t *testing.T) {
	roleARN := "test-arn"

	type args struct {
		pcs v1beta1.ProviderConfigSpec
	}
	type want struct {
		arn string
		err error
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"NoArnSetError": {
			args: args{
				pcs: v1beta1.ProviderConfigSpec{},
			},
			want: want{
				err: errors.New("a RoleARN must be set to assume with web identity"),
			},
		},
		"EmptyAssumeRoleWithWebIdentityOptions": {
			args: args{
				pcs: v1beta1.ProviderConfigSpec{
					AssumeRoleWithWebIdentity: &v1beta1.AssumeRoleWithWebIdentityOptions{},
				},
			},
			want: want{
				err: errors.New("a RoleARN must be set to assume with web identity"),
			},
		},
		"AssumeRoleWithWebIdentityOptions": {
			args: args{
				pcs: v1beta1.ProviderConfigSpec{
					AssumeRoleWithWebIdentity: &v1beta1.AssumeRoleWithWebIdentityOptions{
						RoleARN: &roleARN,
					},
				},
			},
			want: want{
				arn: "test-arn",
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {

			roleArn, err := GetAssumeRoleWithWebIdentityARN(&tc.args.pcs)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Wrap: -want, +got:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.arn, pointer.StringValue(roleArn)); diff != "" {
				t.Errorf("Wrap: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestSetAssumeroleOptions(t *testing.T) {
	externalID := "test-id"
	externalIDDep := "test-id-deprecated"

	key1 := "key1"
	value1 := "value1"

	type args struct {
		pc v1beta1.ProviderConfig
	}
	type want struct {
		aro stscreds.AssumeRoleOptions
	}
	cases := map[string]struct {
		args args
		want want
	}{

		"NoOptionsSet": {
			args: args{
				pc: v1beta1.ProviderConfig{
					Spec: v1beta1.ProviderConfigSpec{},
				},
			},
			want: want{
				aro: stscreds.AssumeRoleOptions{},
			},
		},
		"BasicAssumerole": {
			args: args{
				pc: v1beta1.ProviderConfig{
					Spec: v1beta1.ProviderConfigSpec{
						AssumeRole: &v1beta1.AssumeRoleOptions{
							ExternalID: &externalID,
						},
					},
				},
			},
			want: want{
				aro: stscreds.AssumeRoleOptions{
					ExternalID: &externalID,
				},
			},
		},
		"SpecExternalIDDeprecated": {
			args: args{
				pc: v1beta1.ProviderConfig{
					Spec: v1beta1.ProviderConfigSpec{
						ExternalID: &externalIDDep,
					},
				},
			},
			want: want{
				aro: stscreds.AssumeRoleOptions{
					ExternalID: &externalIDDep,
				},
			},
		},
		"SetTagsAndTransitiveTagKeys": {
			args: args{
				pc: v1beta1.ProviderConfig{
					Spec: v1beta1.ProviderConfigSpec{
						ExternalID: &externalIDDep, // should be ignored if AssumeRoleOptions set
						AssumeRole: &v1beta1.AssumeRoleOptions{
							ExternalID:        &externalID,
							Tags:              []v1beta1.Tag{{Key: &key1, Value: &value1}},
							TransitiveTagKeys: []string{"a", "b", "c"},
						},
					},
				},
			},
			want: want{
				aro: stscreds.AssumeRoleOptions{
					ExternalID:        &externalID,
					Tags:              []stscredstypesv2.Tag{{Key: &key1, Value: &value1}},
					TransitiveTagKeys: []string{"a", "b", "c"},
				},
			},
		},
		"ZeroLengthTags": {
			args: args{
				pc: v1beta1.ProviderConfig{
					Spec: v1beta1.ProviderConfigSpec{
						AssumeRole: &v1beta1.AssumeRoleOptions{
							Tags:              []v1beta1.Tag{},
							TransitiveTagKeys: []string{},
						},
					},
				},
			},
			want: want{
				aro: stscreds.AssumeRoleOptions{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {

			aro := stscreds.AssumeRoleOptions{}
			f := SetAssumeRoleOptions(&tc.args.pc)
			f(&aro)

			if diff := cmp.Diff(tc.want.aro, aro, cmpopts.IgnoreUnexported(stscredstypesv2.Tag{})); diff != "" {
				t.Errorf("Wrap: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestSetWebIdentityRoleOptions(t *testing.T) {
	sessionName := "test-id"

	type args struct {
		pc v1beta1.ProviderConfig
	}
	type want struct {
		aro stscreds.WebIdentityRoleOptions
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"NoOptionsSet": {
			args: args{
				pc: v1beta1.ProviderConfig{
					Spec: v1beta1.ProviderConfigSpec{},
				},
			},
			want: want{
				aro: stscreds.WebIdentityRoleOptions{},
			},
		},
		"BasicAssumeRoleWithWebIdentity": {
			args: args{
				pc: v1beta1.ProviderConfig{
					Spec: v1beta1.ProviderConfigSpec{
						AssumeRoleWithWebIdentity: &v1beta1.AssumeRoleWithWebIdentityOptions{
							RoleSessionName: sessionName,
						},
					},
				},
			},
			want: want{
				aro: stscreds.WebIdentityRoleOptions{
					RoleSessionName: sessionName,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {

			aro := stscreds.WebIdentityRoleOptions{}
			f := SetWebIdentityRoleOptions(&tc.args.pc)
			f(&aro)

			if diff := cmp.Diff(tc.want.aro, aro, cmpopts.IgnoreUnexported(stscredstypesv2.Tag{})); diff != "" {
				t.Errorf("Wrap: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestUseProviderConfigResolveEndpoint(t *testing.T) {
	providerConfigReferenceName := "ProviderConfigReference"

	type args struct {
		endpointConfig *v1beta1.EndpointConfig
		region         string
		service        string
	}

	type want struct {
		// the URL returned by the endpoint resolver
		url string
		// any exception which should be generated by the resolver
		error error
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"EndpointConfigNotSet": {
			args: args{
				region: "us-east-1",
			},
			want: want{},
		},
		"DynamicURLConfig": {
			args: args{
				region:  "aws-global",
				service: "iam",
				endpointConfig: &v1beta1.EndpointConfig{URL: v1beta1.URLConfig{
					Type: "Dynamic",
					Dynamic: &v1beta1.DynamicURLConfig{
						Protocol: "https",
						Host:     "amazonaws.com",
					},
				},
				},
			},
			want: want{
				url: "https://iam.aws-global.amazonaws.com",
			},
		},
		"StaticURLConfig": {
			args: args{
				region:  "us-east-1",
				service: "iam",
				endpointConfig: &v1beta1.EndpointConfig{
					HostnameImmutable: pointer.ToOrNilIfZeroValue(true),
					URL: v1beta1.URLConfig{
						Type:   "Static",
						Static: pointer.ToOrNilIfZeroValue("http://localstack:4566"),
					},
				},
			},
			want: want{
				url: "http://localstack:4566",
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mg := fake.Managed{
				ProviderConfigReferencer: fake.ProviderConfigReferencer{
					Ref: &xpv1.Reference{Name: providerConfigReferenceName},
				},
			}
			providerCredentials := v1beta1.ProviderCredentials{Source: xpv1.CredentialsSourceNone}

			kubeClient := &test.MockClient{
				MockGet: test.NewMockGetFn(nil, func(obj client.Object) error {
					switch fake.GVK(obj).Kind {
					case "ProviderConfig":
						*obj.(*v1beta1.ProviderConfig) = v1beta1.ProviderConfig{
							ObjectMeta: v1.ObjectMeta{Name: providerConfigReferenceName},
							Spec:       v1beta1.ProviderConfigSpec{Credentials: providerCredentials, Endpoint: tc.args.endpointConfig},
							Status:     v1beta1.ProviderConfigStatus{},
						}
					case "ProviderConfigUsage":
						*obj.(*v1beta1.ProviderConfigUsage) = v1beta1.ProviderConfigUsage{
							ProviderConfigUsage: xpv1.ProviderConfigUsage{ProviderConfigReference: xpv1.Reference{Name: providerConfigReferenceName}},
						}
					}
					return nil
				}),
			}

			config, err := UseProviderConfig(context.TODO(), kubeClient, &mg, tc.args.region)
			if err != nil {
				t.Errorf("UseProviderConfig threw exception:\n%s", err)
			}

			// If no endpointConfig was provided the returned endpointResolver should be nil
			if tc.args.endpointConfig != nil {
				actual, endpointError := config.EndpointResolverWithOptions.ResolveEndpoint(tc.args.service, tc.args.region, nil)
				// Assert exceptions match
				if diff := cmp.Diff(tc.want.error, endpointError, test.EquateConditions()); diff != "" {
					t.Errorf("r: -want error, +got error:\n%s", diff)
				}
				// Assert endpoints match
				if diff := cmp.Diff(tc.want.url, actual.URL); diff != "" {
					t.Errorf("add: -want, +got:\n%s", diff)
				}
			} else if config.EndpointResolverWithOptions != nil {
				t.Errorf("Expected config.EndpointResolverWithOptions to be nil")
			}
		})
	}
}

func TestUsePodServiceAccount(t *testing.T) {
	awsRegion := "eu-somewhere-1"
	err := os.Setenv("AWS_REGION", awsRegion)
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]string{
		"us-west-2":     "us-west-2",
		"us-gov-east-1": "us-gov-east-1",
		GlobalRegion:    awsRegion,
	}
	for inpuRegion, expectedRegion := range cases {
		cfg, tErr := UsePodServiceAccount(context.Background(), nil, DefaultSection, inpuRegion)
		if tErr != nil {
			t.Error(tErr)
			continue
		}
		if cfg.Region != expectedRegion {
			t.Errorf("expected region was not returend. expected: %s, actually: %s", expectedRegion, cfg.Region)
		}
	}
	err = os.Unsetenv("AWS_REGION")
	if err != nil {
		t.Error(err)
	}
}

func TestUsePodServiceAccountAssumeRole(t *testing.T) {
	awsRegion := "eu-somewhere-1"
	err := os.Setenv("AWS_REGION", awsRegion)
	if err != nil {
		t.Fatal(err)
	}
	providerConfig := v1beta1.ProviderConfig{
		Spec: v1beta1.ProviderConfigSpec{
			AssumeRole: &v1beta1.AssumeRoleOptions{
				RoleARN: pointer.ToOrNilIfZeroValue("arn:aws:iam::123456789:role/crossplane-role"),
			},
			Credentials: v1beta1.ProviderCredentials{Source: xpv1.CredentialsSourceInjectedIdentity},
		},
	}
	cases := map[string]string{
		"us-west-2":     "us-west-2",
		"us-gov-east-1": "us-gov-east-1",
		GlobalRegion:    awsRegion,
	}
	for inpuRegion, expectedRegion := range cases {
		cfg, tErr := UsePodServiceAccountAssumeRole(context.Background(), nil, DefaultSection, inpuRegion, &providerConfig)
		if tErr != nil {
			t.Error(tErr)
			continue
		}
		if cfg.Region != expectedRegion {
			t.Errorf("expected region was not returend. expected: %s, received: %s", expectedRegion, cfg.Region)
		}
	}
	err = os.Unsetenv("AWS_REGION")
	if err != nil {
		t.Error(err)
	}
}
