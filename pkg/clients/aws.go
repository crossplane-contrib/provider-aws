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
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	ec2type "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	awsv1 "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	credentialsv1 "github.com/aws/aws-sdk-go/aws/credentials"
	endpointsv1 "github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/smithy-go"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/go-ini/ini"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/v1alpha3"
	"github.com/crossplane/provider-aws/apis/v1beta1"
)

// DefaultSection for INI files.
const DefaultSection = ini.DefaultSection

// GlobalRegion is the region name used for AWS services that do not have a notion
// of region.
const GlobalRegion = "aws-global"

// Endpoint URL configuration types.
const (
	URLConfigTypeStatic  = "Static"
	URLConfigTypeDynamic = "Dynamic"
)

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
func GetConfig(ctx context.Context, c client.Client, mg resource.Managed, region string) (*aws.Config, error) {
	switch {
	case mg.GetProviderConfigReference() != nil:
		return UseProviderConfig(ctx, c, mg, region)
	case mg.GetProviderReference() != nil:
		return UseProvider(ctx, c, mg, region)
	default:
		return nil, errors.New("neither providerConfigRef nor providerRef is given")
	}
}

// UseProviderConfig to produce a config that can be used to authenticate to AWS.
func UseProviderConfig(ctx context.Context, c client.Client, mg resource.Managed, region string) (*aws.Config, error) {
	pc := &v1beta1.ProviderConfig{}
	if err := c.Get(ctx, types.NamespacedName{Name: mg.GetProviderConfigReference().Name}, pc); err != nil {
		return nil, errors.Wrap(err, "cannot get referenced Provider")
	}

	t := resource.NewProviderConfigUsageTracker(c, &v1beta1.ProviderConfigUsage{})
	if err := t.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, "cannot track ProviderConfig usage")
	}

	switch s := pc.Spec.Credentials.Source; s { //nolint:exhaustive
	case xpv1.CredentialsSourceInjectedIdentity:
		cfg, err := UsePodServiceAccount(ctx, []byte{}, DefaultSection, region)
		if err != nil {
			return nil, err
		}
		return SetResolver(pc, cfg), nil
	default:
		data, err := resource.CommonCredentialExtractor(ctx, s, c, pc.Spec.Credentials.CommonCredentialSelectors)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get credentials")
		}
		cfg, err := UseProviderSecret(ctx, data, DefaultSection, region)
		if err != nil {
			return nil, err
		}
		return SetResolver(pc, cfg), nil
	}
}

// SetResolver parses annotations from the managed resource
// and returns a configuration accordingly.
func SetResolver(pc *v1beta1.ProviderConfig, cfg *aws.Config) *aws.Config { // nolint:gocyclo
	if pc.Spec.Endpoint == nil {
		return cfg
	}
	cfg.EndpointResolver = aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		fullURL := ""
		switch pc.Spec.Endpoint.URL.Type {
		case URLConfigTypeStatic:
			if pc.Spec.Endpoint.URL.Static == nil {
				return aws.Endpoint{}, errors.New("static type is chosen but static field does not have a value")
			}
			fullURL = StringValue(pc.Spec.Endpoint.URL.Static)
		case URLConfigTypeDynamic:
			if pc.Spec.Endpoint.URL.Dynamic == nil {
				return aws.Endpoint{}, errors.New("dynamic type is chosen but dynamic configuration is not given")
			}
			// NOTE(muvaf): IAM does not have any region.
			if service == "IAM" {
				fullURL = fmt.Sprintf("%s://%s.%s", pc.Spec.Endpoint.URL.Dynamic.Protocol, strings.ToLower(service), pc.Spec.Endpoint.URL.Dynamic.Host)
			} else {
				fullURL = fmt.Sprintf("%s://%s.%s.%s", pc.Spec.Endpoint.URL.Dynamic.Protocol, strings.ToLower(service), region, pc.Spec.Endpoint.URL.Dynamic.Host)
			}
		default:
			return aws.Endpoint{}, errors.New("unsupported url config type is chosen")
		}
		e := aws.Endpoint{
			URL:               fullURL,
			HostnameImmutable: BoolValue(pc.Spec.Endpoint.HostnameImmutable),
			PartitionID:       StringValue(pc.Spec.Endpoint.PartitionID),
			SigningName:       StringValue(pc.Spec.Endpoint.SigningName),
			SigningRegion:     StringValue(LateInitializeStringPtr(pc.Spec.Endpoint.SigningRegion, &region)),
			SigningMethod:     StringValue(pc.Spec.Endpoint.SigningMethod),
		}
		// Only IAM does not have a region parameter and "aws-global" is used in
		// SDK setup. However, signing region has to be us-east-1 and it needs
		// to be set.
		if region == "aws-global" {
			switch StringValue(pc.Spec.Endpoint.PartitionID) {
			case "aws-us-gov", "aws-cn":
				e.SigningRegion = StringValue(LateInitializeStringPtr(pc.Spec.Endpoint.SigningRegion, &region))
			default:
				e.SigningRegion = "us-east-1"
			}
		}
		if pc.Spec.Endpoint.Source != nil {
			switch *pc.Spec.Endpoint.Source {
			case "ServiceMetadata":
				e.Source = aws.EndpointSourceServiceMetadata
			case "Custom":
				e.Source = aws.EndpointSourceCustom
			}
		}
		return e, nil
	})
	return cfg
}

// UseProvider to produce a config that can be used to authenticate to AWS.
// Deprecated: Use UseProviderConfig.
func UseProvider(ctx context.Context, c client.Client, mg resource.Managed, region string) (*aws.Config, error) {
	p := &v1alpha3.Provider{}
	if err := c.Get(ctx, types.NamespacedName{Name: mg.GetProviderReference().Name}, p); err != nil {
		return nil, errors.Wrap(err, "cannot get referenced Provider")
	}

	if region == "" {
		region = p.Spec.Region
	}

	if BoolValue(p.Spec.UseServiceAccount) {
		return UsePodServiceAccount(ctx, []byte{}, DefaultSection, region)
	}

	if p.Spec.CredentialsSecretRef == nil {
		return nil, errors.New("provider does not have a secret reference")
	}

	csr := p.Spec.CredentialsSecretRef
	secret := &corev1.Secret{}
	if err := c.Get(ctx, types.NamespacedName{Namespace: csr.Namespace, Name: csr.Name}, secret); err != nil {
		return nil, errors.Wrap(err, "cannot get credentials secret")
	}

	return UseProviderSecret(ctx, secret.Data[csr.Key], DefaultSection, region)
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
func UseProviderSecret(ctx context.Context, data []byte, profile, region string) (*aws.Config, error) {
	creds, err := CredentialsIDSecret(data, profile)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse credentials secret")
	}

	config, err := config.LoadDefaultConfig(ctx, config.WithRegion(region), config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
		Value: creds,
	}))
	return &config, err
}

// UsePodServiceAccount assumes an IAM role configured via a ServiceAccount.
// https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html
func UsePodServiceAccount(ctx context.Context, _ []byte, _, region string) (*aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(region),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load default AWS config")
	}
	return &cfg, err
}

// NOTE(muvaf): ACK-generated controllers use aws/aws-sdk-go instead of
// aws/aws-sdk-go-v2. These functions are implemented to be used by those controllers.

// GetConfigV1 constructs an *awsv1.Config that can be used to authenticate to AWS
// API by the AWSv1 clients.
func GetConfigV1(ctx context.Context, c client.Client, mg resource.Managed, region string) (*session.Session, error) { // nolint:gocyclo
	if mg.GetProviderConfigReference() == nil {
		return nil, errors.New("providerConfigRef cannot be empty")
	}
	pc := &v1beta1.ProviderConfig{}
	if err := c.Get(ctx, types.NamespacedName{Name: mg.GetProviderConfigReference().Name}, pc); err != nil {
		return nil, errors.Wrap(err, "cannot get referenced ProviderConfig")
	}

	t := resource.NewProviderConfigUsageTracker(c, &v1beta1.ProviderConfigUsage{})
	if err := t.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, "cannot track ProviderConfig usage")
	}
	switch s := pc.Spec.Credentials.Source; s { //nolint:exhaustive
	case xpv1.CredentialsSourceInjectedIdentity:
		cfg, err := UsePodServiceAccountV1(ctx, []byte{}, pc, DefaultSection, region)
		if err != nil {
			return nil, errors.Wrap(err, "cannot use pod service account")
		}
		return session.NewSession(cfg)
	default:
		data, err := resource.CommonCredentialExtractor(ctx, s, c, pc.Spec.Credentials.CommonCredentialSelectors)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get credentials")
		}
		cfg, err := UseProviderSecretV1(ctx, data, pc, DefaultSection, region)
		if err != nil {
			return nil, errors.Wrap(err, "cannot use secret")
		}
		return session.NewSession(cfg)
	}
}

// UseProviderSecretV1 retrieves AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY from
// the data which contains aws credentials under given profile and produces a *awsv1.Config
// Example:
// [default]
// aws_access_key_id = <YOUR_ACCESS_KEY_ID>
// aws_secret_access_key = <YOUR_SECRET_ACCESS_KEY>
func UseProviderSecretV1(_ context.Context, data []byte, pc *v1beta1.ProviderConfig, profile, region string) (*awsv1.Config, error) {
	cfg, err := ini.InsensitiveLoad(data)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse credentials secret")
	}

	iniProfile, err := cfg.GetSection(profile)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("cannot get %s profile in credentials secret", profile))
	}

	accessKeyID := iniProfile.Key("aws_access_key_id")
	secretAccessKey := iniProfile.Key("aws_secret_access_key")
	sessionToken := iniProfile.Key("aws_session_token")

	// NOTE(muvaf): Key function implementation never returns nil but still its
	// type is pointer so we check to make sure its next versions doesn't break
	// that implicit contract.
	if accessKeyID == nil || secretAccessKey == nil || sessionToken == nil {
		return nil, errors.New("returned key can be empty but cannot be nil")
	}

	creds := credentialsv1.NewStaticCredentials(accessKeyID.Value(), secretAccessKey.Value(), sessionToken.Value())
	return SetResolverV1(pc, awsv1.NewConfig().WithCredentials(creds).WithRegion(region)), nil
}

// UsePodServiceAccountV1 assumes an IAM role configured via a ServiceAccount.
// https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html
func UsePodServiceAccountV1(ctx context.Context, _ []byte, pc *v1beta1.ProviderConfig, _, region string) (*awsv1.Config, error) {
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(region),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load default AWS config")
	}
	v2creds, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve credentials")
	}
	v1creds := credentialsv1.NewStaticCredentials(
		v2creds.AccessKeyID,
		v2creds.SecretAccessKey,
		v2creds.SessionToken)
	return SetResolverV1(pc, awsv1.NewConfig().WithCredentials(v1creds).WithRegion(region)), nil
}

// SetResolverV1 parses annotations from the managed resource
// and returns a V1 configuration accordingly.
func SetResolverV1(pc *v1beta1.ProviderConfig, cfg *awsv1.Config) *awsv1.Config { // nolint:gocyclo
	if pc.Spec.Endpoint == nil {
		return cfg
	}
	cfg.EndpointResolver = endpointsv1.ResolverFunc(func(service, region string, optFns ...func(*endpointsv1.Options)) (endpointsv1.ResolvedEndpoint, error) {
		fullURL := ""
		switch pc.Spec.Endpoint.URL.Type {
		case URLConfigTypeStatic:
			if pc.Spec.Endpoint.URL.Static == nil {
				return endpointsv1.ResolvedEndpoint{}, errors.New("static type is chosen but static field does not have a value")
			}
			fullURL = StringValue(pc.Spec.Endpoint.URL.Static)
		case URLConfigTypeDynamic:
			if pc.Spec.Endpoint.URL.Dynamic == nil {
				return endpointsv1.ResolvedEndpoint{}, errors.New("dynamic type is chosen but dynamic configuration is not given")
			}
			// NOTE(muvaf): IAM does not have any region.
			if service == "IAM" {
				fullURL = fmt.Sprintf("%s://%s.%s", pc.Spec.Endpoint.URL.Dynamic.Protocol, strings.ToLower(service), pc.Spec.Endpoint.URL.Dynamic.Host)
			} else {
				fullURL = fmt.Sprintf("%s://%s.%s.%s", pc.Spec.Endpoint.URL.Dynamic.Protocol, strings.ToLower(service), region, pc.Spec.Endpoint.URL.Dynamic.Host)
			}
		default:
			return endpointsv1.ResolvedEndpoint{}, errors.New("unsupported url config type is chosen")
		}
		e := endpointsv1.ResolvedEndpoint{
			URL:           fullURL,
			PartitionID:   StringValue(pc.Spec.Endpoint.PartitionID),
			SigningName:   StringValue(pc.Spec.Endpoint.SigningName),
			SigningRegion: StringValue(LateInitializeStringPtr(pc.Spec.Endpoint.SigningRegion, &region)),
			SigningMethod: StringValue(pc.Spec.Endpoint.SigningMethod),
		}
		// Only IAM does not have a region parameter and "aws-global" is used in
		// SDK setup. However, signing region has to be us-east-1 and it needs
		// to be set.
		if region == "aws-global" {
			switch StringValue(pc.Spec.Endpoint.PartitionID) {
			case "aws-us-gov", "aws-cn":
				e.SigningRegion = StringValue(LateInitializeStringPtr(pc.Spec.Endpoint.SigningRegion, &region))
			default:
				e.SigningRegion = "us-east-1"
			}
		}
		return e, nil
	})
	return cfg
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
	return aws.ToString(v)
}

// BoolValue calls underlying aws ToBool
func BoolValue(v *bool) bool {
	return aws.ToBool(v)
}

// Int64Value converts the supplied int64 pointer to a int64, returning
// 0 if the pointer is nil.
func Int64Value(v *int64) int64 {
	if v != nil {
		return *v
	}
	return 0
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

// LateInitializeTimePtr returns in if it's non-nil, otherwise returns from
// which is the backup for the cases in is nil.
func LateInitializeTimePtr(in *metav1.Time, from *time.Time) *metav1.Time {
	if in != nil {
		return in
	}
	if from != nil {
		t := metav1.NewTime(*from)
		return &t
	}
	return nil
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

// Int32 converts the supplied int for use with the AWS Go SDK.
func Int32(v int, o ...FieldOption) *int32 {
	for _, fo := range o {
		if fo == FieldRequired && v == 0 {
			return aws.Int32(int32(v))
		}
	}

	if v == 0 {
		return nil
	}

	return aws.Int32(int32(v))
}

// Int64Address returns the given *int in the form of *int64.
func Int64Address(i *int) *int64 {
	if i == nil {
		return nil
	}
	return aws.Int64(int64(*i))
}

// Int32Address returns the given *int in the form of *int32.
func Int32Address(i *int) *int32 {
	if i == nil {
		return nil
	}
	return aws.Int32(int32(*i))
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

// IntFrom32Address converts the supplied int32 pointer to an int pointer, returning nil if
// the pointer is nil.
func IntFrom32Address(i *int32) *int {
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

// LateInitializeIntFrom32Ptr returns in if it's non-nil, otherwise returns from
// which is the backup for the cases in is nil.
func LateInitializeIntFrom32Ptr(in *int, from *int32) *int {
	if in != nil {
		return in
	}
	if from != nil && *from != 0 {
		i := int(*from)
		return &i
	}
	return nil
}

// LateInitializeInt32Ptr returns in if it's non-nil, otherwise returns from
// which is the backup for the cases in is nil.
func LateInitializeInt32Ptr(in *int32, from *int32) *int32 {
	if in != nil {
		return in
	}
	return from
}

// LateInitializeInt64Ptr returns in if it's non-nil, otherwise returns from
// which is the backup for the cases in is nil.
func LateInitializeInt64Ptr(in *int64, from *int64) *int64 {
	if in != nil {
		return in
	}
	return from
}

// LateInitializeInt32 returns in if it's non-zero, otherwise returns from
// which is the backup for the cases in is zero.
func LateInitializeInt32(in int32, from int32) int32 {
	if in != 0 {
		return in
	}
	return from
}

// LateInitializeInt64 returns in if it's non-zero, otherwise returns from
// which is the backup for the cases in is zero.
func LateInitializeInt64(in int64, from int64) int64 {
	if in != 0 {
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

// DiffEC2Tags returns []ec2type.Tag that should be added or removed.
func DiffEC2Tags(local []ec2type.Tag, remote []ec2type.Tag) (add []ec2type.Tag, remove []ec2type.Tag) {
	var tagsToAdd = make(map[string]string, len(local))
	add = []ec2type.Tag{}
	remove = []ec2type.Tag{}
	for _, j := range local {
		tagsToAdd[aws.ToString(j.Key)] = aws.ToString(j.Value)
	}
	for _, j := range remote {
		switch val, ok := tagsToAdd[aws.ToString(j.Key)]; {
		case ok && val == aws.ToString(j.Value):
			delete(tagsToAdd, aws.ToString(j.Key))
		case !ok:
			remove = append(remove, ec2type.Tag{
				Key:   j.Key,
				Value: nil,
			})
		}
	}
	for i, j := range tagsToAdd {
		add = append(add, ec2type.Tag{
			Key:   aws.String(i),
			Value: aws.String(j),
		})
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

// IsPolicyUpToDate Marshall policies to json for a compare to get around string ordering
func IsPolicyUpToDate(local, remote *string) bool {
	var localUnmarshalled interface{}
	var remoteUnmarshalled interface{}

	var err error
	err = json.Unmarshal([]byte(*local), &localUnmarshalled)
	if err != nil {
		return false
	}
	err = json.Unmarshal([]byte(*remote), &remoteUnmarshalled)
	if err != nil {
		return false
	}

	sortSlicesOpt := cmpopts.SortSlices(func(x, y interface{}) bool {
		if a, ok := x.(string); ok {
			if b, ok := y.(string); ok {
				return a < b
			}
		}
		// Note: Unknown types in slices will not cause a panic, but
		// may not be sorted correctly. Depending on how AWS handles
		// these, it may cause constant updates - but better this than
		// panicing.
		return false
	})
	return cmp.Equal(localUnmarshalled, remoteUnmarshalled, cmpopts.EquateEmpty(), sortSlicesOpt)
}

// Wrap will remove the request-specific information from the error and only then
// wrap it.
func Wrap(err error, msg string) error {
	// NOTE(muvaf): nil check is done for performance, otherwise errors.As makes
	// a few reflection calls before returning false, letting awsErr be nil.
	if err == nil {
		return nil
	}
	var awsErr smithy.APIError
	if errors.As(err, &awsErr) {
		return errors.Wrap(awsErr, msg)
	}
	// AWS SDK v1 uses different interfaces than v2 and it doesn't unwrap to
	// the underlying error. So, we need to strip off the unique request ID
	// manually.
	if v1RequestError, ok := err.(awserr.RequestFailure); ok {
		return errors.Wrap(errors.New(strings.ReplaceAll(err.Error(), v1RequestError.RequestID(), "")), msg)
	}
	return errors.Wrap(err, msg)
}
