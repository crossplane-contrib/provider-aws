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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-ini/ini"
	"k8s.io/client-go/kubernetes"

	"github.com/crossplaneio/stack-aws/apis/v1alpha2"

	"github.com/crossplaneio/crossplane-runtime/pkg/util"
)

// DefaultSection for INI files.
const DefaultSection = ini.DefaultSection

// A FieldOption determines how common Go types are translated to the types
// required by the AWS Go SDK.
type FieldOption int

// Field options.
const (
	// FieldRequired causes zero values to be converted to a pointer to the zero
	// value, rather than a nil pointer. AWS Go SDK types use pointer fields,
	// with a nil pointer indicating an unset field. Our ToPtr functions return
	// a nil pointer for a zero values, unless FieldRequired is set.
	FieldRequired FieldOption = iota
)

// CredentialsIDSecret retrieves AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY from the data which contains
// aws credentials under given profile
// Example:
// [default]
// aws_access_key_id = <YOUR_ACCESS_KEY_ID>
// aws_secret_access_key = <YOUR_SECRET_ACCESS_KEY>
func CredentialsIDSecret(data []byte, profile string) (string, string, error) {
	config, err := ini.InsensitiveLoad(data)
	if err != nil {
		return "", "", err
	}

	iniProfile, err := config.GetSection(profile)
	if err != nil {
		return "", "", err
	}

	id, err := iniProfile.GetKey(external.AWSAccessKeyIDEnvVar)
	if err != nil {
		return "", "", err
	}

	secret, err := iniProfile.GetKey(external.AWSSecreteAccessKeyEnvVar)
	if err != nil {
		return "", "", err
	}

	return id.Value(), secret.Value(), err
}

// LoadConfig - AWS configuration which can be used to issue requests against AWS API
func LoadConfig(data []byte, profile, region string) (*aws.Config, error) {
	id, secret, err := CredentialsIDSecret(data, profile)
	if err != nil {
		return nil, err
	}

	creds := aws.Credentials{
		AccessKeyID:     id,
		SecretAccessKey: secret,
	}

	shared := external.SharedConfig{
		Credentials: creds,
		Region:      region,
	}

	config, err := external.LoadDefaultAWSConfig(shared)
	return &config, err
}

// ValidateConfig - validates AWS configuration by issuing list s3 buckets request
// TODO: find a better way to validate credentials
func ValidateConfig(config *aws.Config) error {
	svc := s3.New(*config)
	_, err := svc.ListBucketsRequest(nil).Send()
	return err
}

// Config - crate AWS Config based on credentials data using [default] profile
func Config(client kubernetes.Interface, p *v1alpha2.Provider) (*aws.Config, error) {
	data, err := util.SecretData(client, p.Namespace, p.Spec.Secret)
	if err != nil {
		return nil, err
	}

	return LoadConfig(data, DefaultSection, p.Spec.Region)
}

// StringAddress returns address of the given string.
func StringAddress(v string) *string {
	return &v
}

// Int64Address returns address of the given int64.
func Int64Address(i *int) *int64 {
	if i == nil {
		return nil
	}
	n := int64(*i)
	return &n
}

// String converts the supplied string for use with the AWS Go SDK.
func String(v string, o ...FieldOption) *string {
	for _, fo := range o {
		if fo == FieldRequired && v == "" {
			return aws.String(v)
		}
	}

	if v == "" {
		return nil
	}

	return aws.String(v)
}

// Int64 converts the supplied int for use with the AWS Go SDK.
func Int64(v int, o ...FieldOption) *int64 {
	for _, fo := range o {
		if fo == FieldRequired && v == 0 {
			return aws.Int64(int64(v))
		}
	}

	if v == 0 {
		return nil
	}

	return aws.Int64(int64(v))
}

// Bool converts the supplied bool for use with the AWS Go SDK.
func Bool(v bool, o ...FieldOption) *bool {
	for _, fo := range o {
		if fo == FieldRequired && !v {
			return aws.Bool(v)
		}
	}

	if !v {
		return nil
	}
	return aws.Bool(v)
}

// StringValue converts the supplied string pointer to a string, returning the
// empty string if the pointer is nil.
func StringValue(v *string) string {
	return aws.StringValue(v)
}

// Int64Value converts the supplied int64 pointer to an int, returning zero if
// the pointer is nil.
func Int64Value(v *int64) int {
	return int(aws.Int64Value(v))
}

// Int64Ref converts the supplied int64 pointer to an int pointer, returning nil if
// the pointer is nil.
func Int64Ref(i *int64) *int {
	if i == nil {
		return nil
	}
	r := int(*i)
	return &r
}

// Float64Value converts the supplied float64 pointer to an int, returning zero if
// the pointer is nil.
func Float64Value(v *float64) int {
	return int(aws.Float64Value(v))
}

// BoolValue converts the supplied bool pointer to a bool, returning false if
// the pointer is nil.
func BoolValue(v *bool) bool {
	return aws.BoolValue(v)
}

// IntValue converts the supplied int pointer to a int, returning 0 if
// the pointer is nil.
func IntValue(v *int) int {
	return aws.IntValue(v)
}

func LateInitializeStringPtr(in *string, from *string) *string {
	if in != nil {
		return in
	}
	return from
}

func LateInitializeBoolPtr(in *bool, from *bool) *bool {
	if in != nil {
		return in
	}
	return from
}

func LateInitializeIntPtr(in *int, from *int64) *int {
	if in != nil {
		return in
	}
	if from != nil {
		i := int(*from)
		return &i
	}
	return nil
}
