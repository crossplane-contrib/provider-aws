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
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	awsv1 "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	endpointsv1 "github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/go-ini/ini"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/provider-aws/apis/v1alpha3"
	"github.com/crossplane/provider-aws/apis/v1beta1"
)

// DefaultSection for INI files.
const DefaultSection = ini.DefaultSection

// GetGlobalRegion returns the global region for AWS services that do not have a notion
// of region.  Ex. iam
// Defaults to 'aws' partition
func GetGlobalRegion(partition string) string {
	// Global regions are convention based by partition
	// Ex. "aws-global"
	if partition == "" {
		return "aws-global"
	}
	return partition + "-global"
}

// GetGlobalRegionForProviderConfig returns the global region for a provider config
func GetGlobalRegionForProviderConfig(ctx context.Context, c client.Client, mg resource.Managed) (string, error) {
	pc, err := GetProviderConfig(ctx, c, mg)
	if err != nil {
		return "", err
	}
	partition := pc.Spec.Credentials.Partition
	region := GetGlobalRegion(partition)
	return region, nil
}

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

// GetProviderConfig retrieves the providerConfig from kubernetes associated with a managed resource
func GetProviderConfig(ctx context.Context, c client.Client, mg resource.Managed) (*v1beta1.ProviderConfig, error) {
	pc := &v1beta1.ProviderConfig{}
	if err := c.Get(ctx, types.NamespacedName{Name: mg.GetProviderConfigReference().Name}, pc); err != nil {
		return nil, errors.Wrap(err, "cannot get referenced Provider")
	}
	return pc, nil
}

// UseProviderConfig to produce a config that can be used to authenticate to AWS.
func UseProviderConfig(ctx context.Context, c client.Client, mg resource.Managed, region string) (*aws.Config, error) {
	pc, err := GetProviderConfig(ctx, c, mg)
	if err != nil {
		return nil, err
	}

	t := resource.NewProviderConfigUsageTracker(c, &v1beta1.ProviderConfigUsage{})
	if err := t.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, "cannot track ProviderConfig usage")
	}

	switch s := pc.Spec.Credentials.Source; s { //nolint:exhaustive
	case xpv1.CredentialsSourceInjectedIdentity:
		cfg, err := UsePodServiceAccount(ctx, []byte{}, DefaultSection, region)
		return SetResolver(ctx, mg, cfg), err
	default:
		data, err := resource.CommonCredentialExtractor(ctx, s, c, pc.Spec.Credentials.CommonCredentialSelectors)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get credentials")
		}
		cfg, err := UseProviderSecret(ctx, data, DefaultSection, region)
		return SetResolver(ctx, mg, cfg), err
	}
}

// SetResolver parses annotations from the managed resource
// and returns a configuration accordingly.
func SetResolver(ctx context.Context, mg resource.Managed, cfg *aws.Config) *aws.Config {
	if ServiceID, ok := mg.GetAnnotations()["aws.alpha.crossplane.io/endpointServiceID"]; ok {
		if URL, ok := mg.GetAnnotations()["aws.alpha.crossplane.io/endpointURL"]; ok {
			endpoint := aws.Endpoint{
				URL: URL,
			}
			if Region, ok := mg.GetAnnotations()["aws.alpha.crossplane.io/endpointSigningRegion"]; ok {
				endpoint.SigningRegion = Region
			}

			defaultResolver := endpoints.NewDefaultResolver()
			endpointResolver := func(service, region string) (aws.Endpoint, error) {
				if strings.Contains(ServiceID, service) {
					return endpoint, nil
				}

				return defaultResolver.ResolveEndpoint(service, region)
			}
			cfg.EndpointResolver = aws.EndpointResolverFunc(endpointResolver)
		}
	}
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

	if aws.BoolValue(p.Spec.UseServiceAccount) {
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
		cfg, err := UsePodServiceAccountV1(ctx, []byte{}, mg, DefaultSection, region)
		if err != nil {
			return nil, errors.Wrap(err, "cannot use pod service account")
		}
		return session.NewSession(cfg)
	default:
		data, err := resource.CommonCredentialExtractor(ctx, s, c, pc.Spec.Credentials.CommonCredentialSelectors)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get credentials")
		}
		cfg, err := UseProviderSecretV1(ctx, data, mg, DefaultSection, region)
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
func UseProviderSecretV1(ctx context.Context, data []byte, mg resource.Managed, profile, region string) (*awsv1.Config, error) {
	config, err := ini.InsensitiveLoad(data)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse credentials secret")
	}

	iniProfile, err := config.GetSection(profile)
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

	creds := credentials.NewStaticCredentials(accessKeyID.Value(), secretAccessKey.Value(), sessionToken.Value())
	return SetResolverV1(ctx, mg, awsv1.NewConfig().WithCredentials(creds).WithRegion(region)), nil
}

// UsePodServiceAccountV1 assumes an IAM role configured via a ServiceAccount.
// https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html
func UsePodServiceAccountV1(ctx context.Context, _ []byte, mg resource.Managed, _, region string) (*awsv1.Config, error) {
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
	creds := credentials.NewStaticCredentials(
		aws.StringValue(resp.Credentials.AccessKeyId),
		aws.StringValue(resp.Credentials.SecretAccessKey),
		aws.StringValue(resp.Credentials.SessionToken))

	return SetResolverV1(ctx, mg, awsv1.NewConfig().WithCredentials(creds).WithRegion(region)), nil
}

// SetResolverV1 parses annotations from the managed resource
// and returns a V1 configuration accordingly.
func SetResolverV1(ctx context.Context, mg resource.Managed, cfg *awsv1.Config) *awsv1.Config {
	if ServiceID, ok := mg.GetAnnotations()["aws.alpha.crossplane.io/endpointServiceID"]; ok {
		if URL, ok := mg.GetAnnotations()["aws.alpha.crossplane.io/endpointURL"]; ok {
			endpoint := endpointsv1.ResolvedEndpoint{
				URL: URL,
			}
			if Region, ok := mg.GetAnnotations()["aws.alpha.crossplane.io/endpointSigningRegion"]; ok {
				endpoint.SigningRegion = Region
			}

			endpointResolver := func(service, region string, optFns ...func(*endpointsv1.Options)) (endpointsv1.ResolvedEndpoint, error) {
				if strings.Contains(ServiceID, service) {
					return endpoint, nil
				}

				return endpointsv1.DefaultResolver().EndpointFor(service, region, optFns...)
			}
			cfg.EndpointResolver = endpointsv1.ResolverFunc(endpointResolver)
		}
	}
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
	return aws.StringValue(v)
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

// BoolValue returns the value of the bool pointer passed in or
// false if the pointer is nil.
func BoolValue(v *bool) bool {
	return aws.BoolValue(v)
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

// DiffEC2Tags returns []ec2.Tag that should be added or removed.
func DiffEC2Tags(local []ec2.Tag, remote []ec2.Tag) (add []ec2.Tag, remove []ec2.Tag) {
	var tagsToAdd = make(map[string]string, len(local))
	add = []ec2.Tag{}
	remove = []ec2.Tag{}
	for _, j := range local {
		tagsToAdd[aws.StringValue(j.Key)] = aws.StringValue(j.Value)
	}
	for _, j := range remote {
		switch val, ok := tagsToAdd[aws.StringValue(j.Key)]; {
		case ok && val == aws.StringValue(j.Value):
			delete(tagsToAdd, aws.StringValue(j.Key))
		case !ok:
			remove = append(remove, ec2.Tag{
				Key:   j.Key,
				Value: nil,
			})
		}
	}
	for i, j := range tagsToAdd {
		add = append(add, ec2.Tag{
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

// CleanError Will remove the requestID from a awserr.Error and return a new awserr.
// If not awserr it will return the original error
func CleanError(err error) error {
	if err == nil {
		return err
	}
	if awsErr, ok := err.(awserr.Error); ok {
		return awserr.New(awsErr.Code(), awsErr.Message(), nil)
	}
	return err
}

// Wrap Attempts to remove requestID from awserr before calling Wrap
func Wrap(err error, msg string) error {
	return errors.Wrap(CleanError(err), msg)
}
