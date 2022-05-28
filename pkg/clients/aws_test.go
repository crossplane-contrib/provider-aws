/*
Copyright 2019 The Crossplane Authors.

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

package aws

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/transport/http"
	stscreds "github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	stscredstypesv2 "github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/aws/smithy-go"
	"github.com/aws/smithy-go/document"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/resource/fake"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane-contrib/provider-aws/apis/v1beta1"
)

const (
	awsCredentialsFileFormat = "[%s]\naws_access_key_id = %s\naws_secret_access_key = %s"

	errBoom = "boom"
	errMsg  = "example err msg"
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

			if diff := cmp.Diff(tc.want.arn, StringValue(roleArn)); diff != "" {
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

			if diff := cmp.Diff(tc.want.arn, StringValue(roleArn)); diff != "" {
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

func TestDiffTags(t *testing.T) {
	type args struct {
		local  map[string]string
		remote map[string]string
	}

	type want struct {
		add    map[string]string
		remove []string
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"Add": {
			args: args{
				local:  map[string]string{"key": "val", "another": "tag"},
				remote: map[string]string{},
			},
			want: want{
				add: map[string]string{
					"key":     "val",
					"another": "tag",
				},
				remove: []string{},
			},
		},
		"Remove": {
			args: args{
				local: map[string]string{},

				remote: map[string]string{"key": "val", "test": "one"},
			},
			want: want{
				add:    map[string]string{},
				remove: []string{"key", "test"},
			},
		},
		"AddAndRemove": {
			args: args{
				local:  map[string]string{"key": "val", "another": "tag"},
				remote: map[string]string{"key": "val", "test": "one"},
			},
			want: want{
				add: map[string]string{
					"another": "tag",
				},
				remove: []string{"test"},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			add, remove := DiffTags(tc.args.local, tc.args.remote)
			if diff := cmp.Diff(tc.want.add, add); diff != "" {
				t.Errorf("add: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.remove, remove, cmpopts.SortSlices(func(a, b string) bool { return a > b })); diff != "" {
				t.Errorf("remove: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDiffEC2Tags(t *testing.T) {
	type args struct {
		local  []ec2types.Tag
		remote []ec2types.Tag
	}
	type want struct {
		add    []ec2types.Tag
		remove []ec2types.Tag
	}
	cases := map[string]struct {
		args
		want
	}{
		"EmptyLocalAndRemote": {
			args: args{
				local:  []ec2types.Tag{},
				remote: []ec2types.Tag{},
			},
			want: want{
				add:    []ec2types.Tag{},
				remove: []ec2types.Tag{},
			},
		},
		"TagsWithSameKeyValuesAndLength": {
			args: args{
				local: []ec2types.Tag{
					{
						Key:   aws.String("name"),
						Value: aws.String("somename"),
					},
				},
				remote: []ec2types.Tag{
					{
						Key:   aws.String("name"),
						Value: aws.String("somename"),
					},
				},
			},
			want: want{
				add:    []ec2types.Tag{},
				remove: []ec2types.Tag{},
			},
		},
		"TagsWithSameKeyDifferentValuesAndSameLength": {
			args: args{
				local: []ec2types.Tag{
					{
						Key:   aws.String("name"),
						Value: aws.String("somename"),
					},
				},
				remote: []ec2types.Tag{
					{
						Key:   aws.String("name"),
						Value: aws.String("somenames"),
					},
				},
			},
			want: want{
				add: []ec2types.Tag{
					{
						Key:   aws.String("name"),
						Value: aws.String("somename"),
					},
				},
				remove: []ec2types.Tag{},
			},
		},
		"EmptyRemoteAndMultipleInputs": {
			args: args{
				local: []ec2types.Tag{
					{
						Key:   aws.String("name"),
						Value: aws.String("somename"),
					},
					{
						Key:   aws.String("tags"),
						Value: aws.String("True"),
					},
				},
				remote: []ec2types.Tag{},
			},
			want: want{
				add: []ec2types.Tag{
					{
						Key:   aws.String("name"),
						Value: aws.String("somename"),
					},
					{
						Key:   aws.String("tags"),
						Value: aws.String("True"),
					},
				},
				remove: []ec2types.Tag{},
			},
		},
		"EmptyLocalAndMultipleRemote": {
			args: args{
				local: []ec2types.Tag{},
				remote: []ec2types.Tag{
					{
						Key:   aws.String("name"),
						Value: aws.String("somename"),
					},
					{
						Key:   aws.String("tags"),
						Value: aws.String("True"),
					},
				},
			},
			want: want{
				add: []ec2types.Tag{},
				remove: []ec2types.Tag{
					{
						Key:   aws.String("name"),
						Value: nil,
					},
					{
						Key:   aws.String("tags"),
						Value: nil,
					},
				},
			},
		},
		"LocalHaveMoreTags": {
			args: args{
				local: []ec2types.Tag{
					{
						Key:   aws.String("name"),
						Value: aws.String("somename"),
					},
					{
						Key:   aws.String("tags"),
						Value: aws.String("True"),
					},
				},
				remote: []ec2types.Tag{
					{
						Key:   aws.String("name"),
						Value: aws.String("somename"),
					},
					{
						Key:   aws.String("val"),
						Value: aws.String("key"),
					},
					{
						Key:   aws.String("val1"),
						Value: aws.String("key2"),
					},
				},
			},
			want: want{
				add: []ec2types.Tag{
					{
						Key:   aws.String("tags"),
						Value: aws.String("True"),
					},
				},
				remove: []ec2types.Tag{
					{
						Key:   aws.String("val"),
						Value: nil,
					},
					{
						Key:   aws.String("val1"),
						Value: nil,
					},
				},
			},
		},
		"RemoteHaveMoreTags": {
			args: args{
				local: []ec2types.Tag{
					{
						Key:   aws.String("name"),
						Value: aws.String("somename"),
					},
					{
						Key:   aws.String("val"),
						Value: aws.String("key"),
					},
				},
				remote: []ec2types.Tag{
					{
						Key:   aws.String("name"),
						Value: aws.String("somename"),
					},
					{
						Key:   aws.String("tags"),
						Value: aws.String("True"),
					},
					{
						Key:   aws.String("val"),
						Value: aws.String("key"),
					},
				},
			},
			want: want{
				add: []ec2types.Tag{},
				remove: []ec2types.Tag{
					{
						Key:   aws.String("tags"),
						Value: nil,
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			tagCmp := cmpopts.SortSlices(func(i, j ec2types.Tag) bool {
				return StringValue(i.Key) < StringValue(j.Key)
			})
			add, remove := DiffEC2Tags(tc.args.local, tc.args.remote)
			if diff := cmp.Diff(tc.want.add, add, tagCmp, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.remove, remove, tagCmp, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDiffLabels(t *testing.T) {
	type args struct {
		local  map[string]string
		remote map[string]string
	}

	type want struct {
		addOrModify map[string]string
		remove      []string
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"Add": {
			args: args{
				local:  map[string]string{"key": "val", "another": "label"},
				remote: map[string]string{},
			},
			want: want{
				addOrModify: map[string]string{
					"key":     "val",
					"another": "label",
				},
				remove: []string{},
			},
		},
		"Remove": {
			args: args{
				local: map[string]string{},

				remote: map[string]string{"key": "val", "test": "one"},
			},
			want: want{
				addOrModify: map[string]string{},
				remove:      []string{"key", "test"},
			},
		},
		"AddAndRemove": {
			args: args{
				local:  map[string]string{"key": "val", "another": "label"},
				remote: map[string]string{"key": "val", "test": "one"},
			},
			want: want{
				addOrModify: map[string]string{
					"another": "label",
				},
				remove: []string{"test"},
			},
		},
		"ModifyOnly": {
			args: args{
				local:  map[string]string{"key": "val"},
				remote: map[string]string{"key": "badval"},
			},
			want: want{
				addOrModify: map[string]string{
					"key": "val",
				},
				remove: []string{},
			},
		},
		"AddModifyRemove": {
			args: args{
				local:  map[string]string{"key": "val", "keytwo": "valtwo", "another": "tag"},
				remote: map[string]string{"key": "val", "keytwo": "badval", "test": "one"},
			},
			want: want{
				addOrModify: map[string]string{
					"keytwo":  "valtwo",
					"another": "tag",
				},
				remove: []string{"test"},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			addOrModify, remove := DiffLabels(tc.args.local, tc.args.remote)
			if diff := cmp.Diff(tc.want.addOrModify, addOrModify); diff != "" {
				t.Errorf("addOrModify: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.remove, remove, cmpopts.SortSlices(func(a, b string) bool { return a > b })); diff != "" {
				t.Errorf("remove: -want, +got:\n%s", diff)
			}
		})
	}
}

const (
	policy = `{"Statement":[{"Action":"ecr:ListImages","Effect":"Allow","Principal":"*"}],"Version":"2012-10-17"}`

	cpxPolicy = `{"Statement":[{"Action":"ecr:ListImages","Effect":"Allow","Principal":{"AWS":["arn:aws:iam::111122223333:userARN","111122223334","arn:aws:iam::111122223333:roleARN"]}}],"Version":"2012-10-17"}`
	// Note: different sort order of principals than input above
	cpxRemPolicy = `{"Statement":[{"Action":"ecr:ListImages","Effect":"Allow","Principal":{"AWS":["111122223334","arn:aws:iam::111122223333:userARN","arn:aws:iam::111122223333:roleARN"]}}],"Version":"2012-10-17"}`
)

func TestIsPolicyUpToDate(t *testing.T) {
	type args struct {
		local  string
		remote string
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFields": {
			args: args{
				local:  "{\"testone\": \"one\", \"testtwo\": \"two\"}",
				remote: "{\"testtwo\": \"two\", \"testone\": \"one\"}",
			},
			want: true,
		},
		"SameFieldsRealPolicy": {
			args: args{
				local:  policy,
				remote: `{"Statement":[{"Effect":"Allow","Action":"ecr:ListImages","Principal":"*"}],"Version":"2012-10-17"}`,
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				local:  "{\"testone\": \"one\", \"testtwo\": \"two\"}",
				remote: "{\"testthree\": \"three\", \"testone\": \"one\"}",
			},
			want: false,
		},
		"SameFieldsPrincipalPolicy": {
			args: args{
				local:  cpxPolicy,
				remote: cpxRemPolicy,
			},
			want: true,
		},
		"SameFieldsNumericPrincipals": {
			args: args{
				// This is to test that our slice sorting does not
				// panic with unexpected value types.
				local:  `{"Statement":[{"Effect":"Allow","Action":"ecr:ListImages","Principal":[2,1,"foo","bar"]}],"Version":"2012-10-17"}`,
				remote: `{"Statement":[{"Effect":"Allow","Action":"ecr:ListImages","Principal":[2,1,"bar","foo"]}],"Version":"2012-10-17"}`,
			},
			want: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsPolicyUpToDate(&tc.args.local, &tc.args.remote)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestWrap(t *testing.T) {
	rootErr := &smithy.GenericAPIError{
		Code:    "InvalidVpcID.NotFound",
		Message: "The vpc ID 'vpc-06f35a4eaed9b4609' does not exist",
		Fault:   smithy.FaultUnknown,
	}
	cases := map[string]struct {
		reason string
		arg    error
		want   error
	}{
		"Nil": {
			arg:  nil,
			want: nil,
		},
		"NonAWSError": {
			reason: "It should not change anything if the error is not coming from AWS",
			arg:    errors.New(errBoom),
			want:   errors.Wrap(errors.New(errBoom), errMsg),
		},
		"AWSError": {
			reason: "Request ID should be removed from the final error if it's an AWS error",
			arg: &smithy.OperationError{
				ServiceID:     "EC2",
				OperationName: "DescribeVpcs",
				Err: &http.ResponseError{
					ResponseError: &smithyhttp.ResponseError{
						Err: rootErr,
					},
					RequestID: "c3dc34d4-b9d6-42a1-9909-7e8f62c6b9cc",
				},
			},
			want: errors.Wrap(rootErr, errMsg),
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := Wrap(tc.arg, errMsg)
			if diff := cmp.Diff(tc.want, err, test.EquateErrors()); diff != "" {
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
					HostnameImmutable: aws.Bool(true),
					URL: v1beta1.URLConfig{
						Type:   "Static",
						Static: aws.String("http://localstack:4566"),
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

func TestDiffTagsMapPtr(t *testing.T) {
	type args struct {
		cr  map[string]*string
		obj map[string]*string
	}
	type want struct {
		addTags    map[string]*string
		removeTags []*string
	}

	cases := map[string]struct {
		args
		want
	}{
		"AddNewTag": {
			args: args{
				cr: map[string]*string{
					"k1": String("exists_in_both"),
					"k2": String("only_in_cr"),
				},
				obj: map[string]*string{
					"k1": String("exists_in_both"),
				}},
			want: want{
				addTags: map[string]*string{
					"k2": String("only_in_cr"),
				},
				removeTags: []*string{},
			},
		},
		"RemoveExistingTag": {
			args: args{
				cr: map[string]*string{
					"k1": String("exists_in_both"),
				},
				obj: map[string]*string{
					"k1": String("exists_in_both"),
					"k2": String("only_in_aws"),
				}},
			want: want{
				addTags: map[string]*string{},
				removeTags: []*string{
					String("k2"),
				}},
		},
		"AddAndRemoveWhenKeyChanges": {
			args: args{
				cr: map[string]*string{
					"k1": String("exists_in_both"),
					"k2": String("same_key_different_value_1"),
				},
				obj: map[string]*string{
					"k1": String("exists_in_both"),
					"k2": String("same_key_different_value_2"),
				}},
			want: want{
				addTags: map[string]*string{
					"k2": String("same_key_different_value_1"),
				},
				removeTags: []*string{
					String("k2"),
				}},
		},
		"NoChange": {
			args: args{
				cr: map[string]*string{
					"k1": String("exists_in_both"),
				},
				obj: map[string]*string{
					"k1": String("exists_in_both"),
				}},
			want: want{
				addTags:    map[string]*string{},
				removeTags: []*string{},
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Act
			addTags, removeTags := DiffTagsMapPtr(tc.args.cr, tc.args.obj)

			// Assert
			if diff := cmp.Diff(tc.want.addTags, addTags, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.removeTags, removeTags, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitStringPtrSlice(t *testing.T) {
	type args struct {
		in   []*string
		from []*string
	}

	cases := map[string]struct {
		args args
		want []*string
	}{
		"BothNil": {
			args: args{},
			want: nil,
		},
		"BothEmpty": {
			args: args{
				in:   []*string{},
				from: []*string{},
			},
			want: []*string{},
		},
		"FromNil": {
			args: args{
				in:   aws.StringSlice([]string{"hi!"}),
				from: nil,
			},
			want: aws.StringSlice([]string{"hi!"}),
		},
		"InNil": {
			args: args{
				in:   nil,
				from: aws.StringSlice([]string{"hi!"}),
			},
			want: aws.StringSlice([]string{"hi!"}),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := LateInitializeStringPtrSlice(tc.args.in, tc.args.from)
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Errorf("\nLateInitializeStringPtrSlice(...): -got, +want:\n%s", diff)
			}
		})
	}
}

func TestLateInitInt64PtrSlice(t *testing.T) {
	type args struct {
		in   []*int64
		from []*int64
	}

	cases := map[string]struct {
		args args
		want []*int64
	}{
		"BothNil": {
			args: args{},
			want: nil,
		},
		"BothEmpty": {
			args: args{
				in:   []*int64{},
				from: []*int64{},
			},
			want: []*int64{},
		},
		"FromNil": {
			args: args{
				in:   aws.Int64Slice([]int64{1}),
				from: nil,
			},
			want: aws.Int64Slice([]int64{1}),
		},
		"InNil": {
			args: args{
				in:   nil,
				from: aws.Int64Slice([]int64{1}),
			},
			want: aws.Int64Slice([]int64{1}),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := LateInitializeInt64PtrSlice(tc.args.in, tc.args.from)
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Errorf("\nLateInitializeInt64PtrSlice(...): -got, +want:\n%s", diff)
			}
		})
	}
}
