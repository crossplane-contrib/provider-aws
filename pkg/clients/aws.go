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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/crossplane/provider-aws/apis/v1beta1"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/go-ini/ini"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/v1alpha3"
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

// GetConfig constructs an *aws.Config that can be used to authenticate to AWS
// API by the AWS clients.
func GetConfig(ctx context.Context, kube client.Client, cr resource.Managed, region string) (*aws.Config, error) { // nolint:gocyclo
	pc := &v1beta1.ProviderConfig{}
	switch {
	case cr.GetProviderConfigReference() != nil && cr.GetProviderConfigReference().Name != "":
		nn := types.NamespacedName{Name: cr.GetProviderConfigReference().Name}
		if err := kube.Get(ctx, nn, pc); resource.IgnoreNotFound(err) != nil {
			return nil, errors.Wrap(err, "cannot get referenced ProviderConfig")
		}
		if region == "" {
			region = pc.Spec.Region
		}
	case cr.GetProviderReference() != nil && cr.GetProviderReference().Name != "":
		nn := types.NamespacedName{Name: cr.GetProviderReference().Name}
		p := &v1alpha3.Provider{}
		if err := kube.Get(ctx, nn, p); err != nil {
			return nil, errors.Wrap(err, "cannot get referenced Provider")
		}
		p.DeepCopyIntoPC(pc)
		if region == "" {
			region = p.Spec.Region
		}
	default:
		return nil, errors.New("neither providerConfigRef nor providerRef is given")
	}

	if aws.BoolValue(pc.Spec.UseServiceAccount) {
		return UsePodServiceAccount(ctx, []byte{}, DefaultSection, region)
	}

	if pc.Spec.CredentialsSecretRef == nil {
		return nil, errors.New("provider does not have a secret reference")
	}

	secret := &corev1.Secret{}
	nn := types.NamespacedName{Namespace: pc.Spec.CredentialsSecretRef.Namespace, Name: pc.Spec.CredentialsSecretRef.Name}
	err := kube.Get(ctx, nn, secret)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get credentials secret %s", nn)
	}

	return UseProviderSecret(ctx, secret.Data[pc.Spec.CredentialsSecretRef.Key], DefaultSection, region)
}

// CredentialsIDSecret retrieves AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY from the data which contains
// aws credentials under given profile
// Example:
// [default]
// aws_access_key_id = <YOUR_ACCESS_KEY_ID>
// aws_secret_access_key = <YOUR_SECRET_ACCESS_KEY>
func CredentialsIDSecret(data []byte, profile string) (aws.Credentials, error) {
	config, err := ini.InsensitiveLoad(data)
	if err != nil {
		return aws.Credentials{}, errors.Wrap(err, "cannot parse credentials secret")
	}

	iniProfile, err := config.GetSection(profile)
	if err != nil {
		return aws.Credentials{}, errors.Wrap(err, fmt.Sprintf("cannot get %s profile in credentials secret", profile))
	}

	accessKeyID := iniProfile.Key("aws_access_key_id")
	secretAccessKey := iniProfile.Key("aws_secret_access_key")
	sessionToken := iniProfile.Key("aws_session_token")

	// NOTE(muvaf): Key function implementation never returns nil but still its
	// type is pointer so we check to make sure its next versions doesn't break
	// that implicit contract.
	if accessKeyID == nil || secretAccessKey == nil || sessionToken == nil {
		return aws.Credentials{}, errors.New("returned key can be empty but cannot be nil")
	}

	return aws.Credentials{
		AccessKeyID:     accessKeyID.Value(),
		SecretAccessKey: secretAccessKey.Value(),
		SessionToken:    sessionToken.Value(),
	}, nil
}

// AuthMethod is a method of authenticating to the AWS API
type AuthMethod func(context.Context, []byte, string, string) (*aws.Config, error)

// UseProviderSecret - AWS configuration which can be used to issue requests against AWS API
func UseProviderSecret(_ context.Context, data []byte, profile, region string) (*aws.Config, error) {
	creds, err := CredentialsIDSecret(data, profile)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse credentials secret")
	}

	shared := external.SharedConfig{
		Credentials: creds,
		Region:      region,
	}

	config, err := external.LoadDefaultAWSConfig(shared)
	return &config, err
}

// UsePodServiceAccount assumes an IAM role configured via a ServiceAccount.
// https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html
//
// TODO(hasheddan): This should be replaced by the implementation of the Web
// Identity Token Provider in the following PR after merge and subsequent
// release of AWS SDK: https://github.com/aws/aws-sdk-go-v2/pull/488
func UsePodServiceAccount(ctx context.Context, _ []byte, _, region string) (*aws.Config, error) {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to load default AWS config")
	}
	cfg.Region = region
	svc := sts.New(cfg)

	b, err := ioutil.ReadFile(os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE"))
	if err != nil {
		return nil, errors.Wrap(err, "unable to read web identity token file in pod")
	}
	token := string(b)
	sess := strconv.FormatInt(time.Now().UnixNano(), 10)
	role := os.Getenv("AWS_ROLE_ARN")
	resp, err := svc.AssumeRoleWithWebIdentityRequest(
		&sts.AssumeRoleWithWebIdentityInput{
			RoleSessionName:  &sess,
			WebIdentityToken: &token,
			RoleArn:          &role,
		}).Send(ctx)
	if err != nil {
		return nil, err
	}
	creds := aws.Credentials{
		AccessKeyID:     aws.StringValue(resp.Credentials.AccessKeyId),
		SecretAccessKey: aws.StringValue(resp.Credentials.SecretAccessKey),
		SessionToken:    aws.StringValue(resp.Credentials.SessionToken),
	}
	shared := external.SharedConfig{
		Credentials: creds,
		Region:      region,
	}
	config, err := external.LoadDefaultAWSConfig(shared)
	return &config, err
}

// TODO(muvaf): All the types that use CreateJSONPatch are known during
// development time. In order to avoid unnecessary panic checks, we can generate
// the code that creates a patch between two objects that share the same type.

// CreateJSONPatch creates a diff JSON object that can be applied to any other
// JSON object.
func CreateJSONPatch(source, destination interface{}) ([]byte, error) {
	sourceJSON, err := json.Marshal(source)
	if err != nil {
		return nil, err
	}
	destinationJSON, err := json.Marshal(destination)
	if err != nil {
		return nil, err
	}
	patchJSON, err := jsonpatch.CreateMergePatch(sourceJSON, destinationJSON)
	if err != nil {
		return nil, err
	}
	return patchJSON, nil
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

// StringValue converts the supplied string pointer to a string, returning the
// empty string if the pointer is nil.
// TODO(muvaf): is this really meaningful? why not implement it?
func StringValue(v *string) string {
	return aws.StringValue(v)
}

// LateInitializeStringPtr returns in if it's non-nil, otherwise returns from
// which is the backup for the cases in is nil.
func LateInitializeStringPtr(in *string, from *string) *string {
	if in != nil {
		return in
	}
	return from
}

// LateInitializeString returns `from` if `in` is empty and `from` is non-nil,
// in other cases it returns `in`.
func LateInitializeString(in string, from *string) string {
	if in == "" && from != nil {
		return *from
	}
	return in
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

// Int64Address returns the given *int in the form of *int64.
func Int64Address(i *int) *int64 {
	if i == nil {
		return nil
	}
	return aws.Int64(int64(*i))
}

// IntAddress converts the supplied int64 pointer to an int pointer, returning nil if
// the pointer is nil.
func IntAddress(i *int64) *int {
	if i == nil {
		return nil
	}
	r := int(*i)
	return &r
}

// LateInitializeIntPtr returns in if it's non-nil, otherwise returns from
// which is the backup for the cases in is nil.
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

// LateInitializeInt64Ptr returns in if it's non-nil, otherwise returns from
// which is the backup for the cases in is nil.
func LateInitializeInt64Ptr(in *int64, from *int64) *int64 {
	if in != nil {
		return in
	}
	return from
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

// LateInitializeBoolPtr returns in if it's non-nil, otherwise returns from
// which is the backup for the cases in is nil.
func LateInitializeBoolPtr(in *bool, from *bool) *bool {
	if in != nil {
		return in
	}
	return from
}

// CompactAndEscapeJSON removes space characters and URL-encodes the JSON string.
func CompactAndEscapeJSON(s string) (string, error) {
	buffer := new(bytes.Buffer)
	if err := json.Compact(buffer, []byte(s)); err != nil {
		return "", err
	}
	return url.QueryEscape(buffer.String()), nil
}

// DiffTags returns tags that should be added or removed.
func DiffTags(local, remote map[string]string) (add map[string]string, remove []string) {
	add = make(map[string]string, len(local))
	remove = []string{}
	for k, v := range local {
		add[k] = v
	}
	for k, v := range remote {
		switch val, ok := local[k]; {
		case ok && val != v:
			remove = append(remove, k)
		case !ok:
			remove = append(remove, k)
			delete(add, k)
		default:
			delete(add, k)
		}
	}
	return
}

// DiffLabels returns labels that should be added, modified, or removed.
func DiffLabels(local, remote map[string]string) (addOrModify map[string]string, remove []string) {
	addOrModify = make(map[string]string, len(local))
	remove = []string{}
	for k, v := range local {
		addOrModify[k] = v
	}
	for k, v := range remote {
		switch val, ok := local[k]; {
		case ok && val != v:
			// if value does not match key it will be updated by the correct
			// key-value pair being present in the returned addOrModify map
			continue
		case !ok:
			remove = append(remove, k)
			delete(addOrModify, k)
		default:
			delete(addOrModify, k)
		}
	}
	return
}
