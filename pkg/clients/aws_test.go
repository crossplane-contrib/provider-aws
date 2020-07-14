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
