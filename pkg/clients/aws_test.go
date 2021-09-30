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
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/gomega"
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
		local  []ec2.Tag
		remote []ec2.Tag
	}
	type want struct {
		add    []ec2.Tag
		remove []ec2.Tag
	}
	cases := map[string]struct {
		args
		want
	}{
		"EmptyLocalAndRemote": {
			args: args{
				local:  []ec2.Tag{},
				remote: []ec2.Tag{},
			},
			want: want{
				add:    []ec2.Tag{},
				remove: []ec2.Tag{},
			},
		},
		"TagsWithSameKeyValuesAndLength": {
			args: args{
				local: []ec2.Tag{
					{
						Key:   aws.String("name"),
						Value: aws.String("somename"),
					},
				},
				remote: []ec2.Tag{
					{
						Key:   aws.String("name"),
						Value: aws.String("somename"),
					},
				},
			},
			want: want{
				add:    []ec2.Tag{},
				remove: []ec2.Tag{},
			},
		},
		"TagsWithSameKeyDifferentValuesAndSameLength": {
			args: args{
				local: []ec2.Tag{
					{
						Key:   aws.String("name"),
						Value: aws.String("somename"),
					},
				},
				remote: []ec2.Tag{
					{
						Key:   aws.String("name"),
						Value: aws.String("somenames"),
					},
				},
			},
			want: want{
				add: []ec2.Tag{
					{
						Key:   aws.String("name"),
						Value: aws.String("somename"),
					},
				},
				remove: []ec2.Tag{},
			},
		},
		"EmptyRemoteAndMultipleInputs": {
			args: args{
				local: []ec2.Tag{
					{
						Key:   aws.String("name"),
						Value: aws.String("somename"),
					},
					{
						Key:   aws.String("tags"),
						Value: aws.String("True"),
					},
				},
				remote: []ec2.Tag{},
			},
			want: want{
				add: []ec2.Tag{
					{
						Key:   aws.String("name"),
						Value: aws.String("somename"),
					},
					{
						Key:   aws.String("tags"),
						Value: aws.String("True"),
					},
				},
				remove: []ec2.Tag{},
			},
		},
		"EmptyLocalAndMultipleRemote": {
			args: args{
				local: []ec2.Tag{},
				remote: []ec2.Tag{
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
				add: []ec2.Tag{},
				remove: []ec2.Tag{
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
				local: []ec2.Tag{
					{
						Key:   aws.String("name"),
						Value: aws.String("somename"),
					},
					{
						Key:   aws.String("tags"),
						Value: aws.String("True"),
					},
				},
				remote: []ec2.Tag{
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
				add: []ec2.Tag{
					{
						Key:   aws.String("tags"),
						Value: aws.String("True"),
					},
				},
				remove: []ec2.Tag{
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
				local: []ec2.Tag{
					{
						Key:   aws.String("name"),
						Value: aws.String("somename"),
					},
					{
						Key:   aws.String("val"),
						Value: aws.String("key"),
					},
				},
				remote: []ec2.Tag{
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
				add: []ec2.Tag{},
				remove: []ec2.Tag{
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
			tagCmp := cmpopts.SortSlices(func(i, j ec2.Tag) bool {
				return StringValue(i.Key) < StringValue(j.Key)
			})
			add, remove := DiffEC2Tags(tc.args.local, tc.args.remote)
			if diff := cmp.Diff(tc.want.add, add, tagCmp); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.remove, remove, tagCmp); diff != "" {
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
