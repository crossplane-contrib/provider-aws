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
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	stscredstypesv2 "github.com/aws/aws-sdk-go-v2/service/sts/types"
	awsv1 "github.com/aws/aws-sdk-go/aws"
	credentialsv1 "github.com/aws/aws-sdk-go/aws/credentials"
	stscredsv1 "github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	endpointsv1 "github.com/aws/aws-sdk-go/aws/endpoints"
	requestv1 "github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/smithy-go/middleware"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/go-ini/ini"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/metrics"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	"github.com/crossplane-contrib/provider-aws/pkg/version"
)

// DefaultSection for INI files.
const DefaultSection = "DEFAULT"

// GlobalRegion is the region name used for AWS services that do not have a notion
// of region.
const GlobalRegion = "aws-global"

// Endpoint URL configuration types.
const (
	URLConfigTypeStatic  = "Static"
	URLConfigTypeDynamic = "Dynamic"
)

// middlewareV2 constructs the AWS SDK v2 middleware
var middlewareV2 = config.WithAPIOptions([]func(*middleware.Stack) error{
	awsmiddleware.AddUserAgentKeyValue("crossplane-provider-aws", version.Version),
	func(s *middleware.Stack) error {
		return s.Finalize.Add(recordRequestMetrics, middleware.After)
	},
})

// recordRequestMetrics records Prometheus metrics for requests to the AWS APIs
var recordRequestMetrics = middleware.FinalizeMiddlewareFunc("recordRequestMetrics", func(ctx context.Context, in middleware.FinalizeInput, next middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
	metrics.IncAWSAPICall(awsmiddleware.GetServiceID(ctx), awsmiddleware.GetOperationName(ctx), "2")
	return next.HandleFinalize(ctx, in)
})

// userAgentV1 constructs the Crossplane user agent for AWS v1 clients
var userAgentV1 = requestv1.NamedHandler{
	Name: "crossplane.UserAgentHandler",
	Fn:   requestv1.MakeAddToUserAgentHandler("crossplane-provider-aws", version.Version),
}

// userAgentV2 constructs the Crossplane user agent for AWS v2 clients
var userAgentV2 = config.WithAPIOptions([]func(*middleware.Stack) error{
	awsmiddleware.AddUserAgentKeyValue("crossplane-provider-aws", version.Version),
})

var (
	muV1            sync.Mutex
	muV2            sync.Mutex
	defaultConfigV2 *aws.Config
	defaultConfigV1 *awsv1.Config
)

// GetConfig constructs an *aws.Config that can be used to authenticate to AWS
// API by the AWS clients.
func GetConfig(ctx context.Context, c client.Client, mg resource.Managed, region string) (*aws.Config, error) {
	return UseProviderConfig(ctx, c, mg, region)
}

// UseProviderConfig to produce a config that can be used to authenticate to AWS.
func UseProviderConfig(ctx context.Context, c client.Client, mg resource.Managed, region string) (*aws.Config, error) { //nolint:gocyclo
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
		if pc.Spec.AssumeRole != nil || pc.Spec.AssumeRoleARN != nil {
			cfg, err := UsePodServiceAccountAssumeRole(ctx, []byte{}, DefaultSection, region, pc)
			if err != nil {
				return nil, err
			}
			return SetResolver(pc, cfg), nil
		}
		if pc.Spec.AssumeRoleWithWebIdentity != nil && pc.Spec.AssumeRoleWithWebIdentity.RoleARN != nil {
			cfg, err := UsePodServiceAccountAssumeRoleWithWebIdentity(ctx, []byte{}, DefaultSection, region, pc)
			if err != nil {
				return nil, err
			}
			return SetResolver(pc, cfg), nil
		}
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
		if pc.Spec.AssumeRole != nil || pc.Spec.AssumeRoleARN != nil {
			cfg, err := UseProviderSecretAssumeRole(ctx, data, DefaultSection, region, pc)
			if err != nil {
				return nil, err
			}
			return SetResolver(pc, cfg), nil
		}
		cfg, err := UseProviderSecret(ctx, data, DefaultSection, region)
		if err != nil {
			return nil, err
		}
		return SetResolver(pc, cfg), nil
	}
}

type awsEndpointResolverAdaptorWithOptions func(service, region string, options interface{}) (aws.Endpoint, error)

func (a awsEndpointResolverAdaptorWithOptions) ResolveEndpoint(service, region string, options ...interface{}) (aws.Endpoint, error) {
	return a(service, region, options)
}

// SetResolver parses annotations from the managed resource
// and returns a configuration accordingly.
func SetResolver(pc *v1beta1.ProviderConfig, cfg *aws.Config) *aws.Config { //nolint:gocyclo
	if pc.Spec.Endpoint == nil {
		return cfg
	}
	cfg.EndpointResolverWithOptions = awsEndpointResolverAdaptorWithOptions(func(service, region string, options interface{}) (aws.Endpoint, error) {
		fullURL := ""
		switch pc.Spec.Endpoint.URL.Type {
		case URLConfigTypeStatic:
			if pc.Spec.Endpoint.URL.Static == nil {
				return aws.Endpoint{}, errors.New("static type is chosen but static field does not have a value")
			}
			fullURL = pointer.StringValue(pc.Spec.Endpoint.URL.Static)
		case URLConfigTypeDynamic:
			if pc.Spec.Endpoint.URL.Dynamic == nil {
				return aws.Endpoint{}, errors.New("dynamic type is chosen but dynamic configuration is not given")
			}
			// NOTE(muvaf): IAM and Route 53 do not have a region.
			if service == "IAM" || service == "Route 53" {
				fullURL = fmt.Sprintf("%s://%s.%s", pc.Spec.Endpoint.URL.Dynamic.Protocol, strings.ReplaceAll(strings.ToLower(service), " ", ""), pc.Spec.Endpoint.URL.Dynamic.Host)
			} else {
				fullURL = fmt.Sprintf("%s://%s.%s.%s", pc.Spec.Endpoint.URL.Dynamic.Protocol, strings.ToLower(service), region, pc.Spec.Endpoint.URL.Dynamic.Host)
			}
		default:
			return aws.Endpoint{}, errors.New("unsupported url config type is chosen")
		}
		e := aws.Endpoint{
			URL:               fullURL,
			HostnameImmutable: pointer.BoolValue(pc.Spec.Endpoint.HostnameImmutable),
			PartitionID:       pointer.StringValue(pc.Spec.Endpoint.PartitionID),
			SigningName:       pointer.StringValue(pc.Spec.Endpoint.SigningName),
			SigningRegion:     pointer.StringValue(pointer.LateInitialize(pc.Spec.Endpoint.SigningRegion, &region)),
			SigningMethod:     pointer.StringValue(pc.Spec.Endpoint.SigningMethod),
		}
		// Only IAM does not have a region parameter and "aws-global" is used in
		// SDK setup. However, signing region has to be us-east-1 and it needs
		// to be set.
		if region == "aws-global" {
			switch pointer.StringValue(pc.Spec.Endpoint.PartitionID) {
			case "aws-us-gov", "aws-cn", "aws-iso", "aws-iso-b":
				e.SigningRegion = pointer.StringValue(pointer.LateInitialize(pc.Spec.Endpoint.SigningRegion, &region))
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

	config, err := config.LoadDefaultConfig(
		ctx,
		middlewareV2,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: creds,
		}),
	)
	return &config, err
}

// UseProviderSecretAssumeRole - AWS configuration which can be used to issue requests against AWS API
// assume Cross account IAM roles
func UseProviderSecretAssumeRole(ctx context.Context, data []byte, profile, region string, pc *v1beta1.ProviderConfig) (*aws.Config, error) {
	creds, err := CredentialsIDSecret(data, profile)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse credentials secret")
	}

	config, err := config.LoadDefaultConfig(
		ctx,
		middlewareV2,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: creds,
		}),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load default AWS config")
	}

	roleArn, err := GetAssumeRoleARN(pc.Spec.DeepCopy())
	if err != nil {
		return nil, err
	}

	stsSvc := sts.NewFromConfig(config)

	stsAssumeRoleOptions := SetAssumeRoleOptions(pc)
	stsAssume := stscreds.NewAssumeRoleProvider(
		stsSvc,
		pointer.StringValue(roleArn),
		stsAssumeRoleOptions,
	)
	config.Credentials = aws.NewCredentialsCache(stsAssume)

	return &config, err
}

// UsePodServiceAccountAssumeRole assumes an IAM role configured via a ServiceAccount
// assume Cross account IAM roles
// https://aws.amazon.com/blogs/containers/cross-account-iam-roles-for-kubernetes-service-accounts/
func UsePodServiceAccountAssumeRole(ctx context.Context, _ []byte, _, region string, pc *v1beta1.ProviderConfig) (*aws.Config, error) {
	cfg, err := UsePodServiceAccount(ctx, []byte{}, DefaultSection, region)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load default AWS config")
	}
	roleArn, err := GetAssumeRoleARN(pc.Spec.DeepCopy())
	if err != nil {
		return nil, err
	}
	stsclient := sts.NewFromConfig(*cfg)
	stsAssumeRoleOptions := SetAssumeRoleOptions(pc)
	cnf, err := config.LoadDefaultConfig(
		ctx,
		middlewareV2,
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(aws.NewCredentialsCache(
			stscreds.NewAssumeRoleProvider(
				stsclient,
				pointer.StringValue(roleArn),
				stsAssumeRoleOptions,
			)),
		),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load assumed role AWS config")
	}
	return &cnf, err
}

// UsePodServiceAccountAssumeRoleWithWebIdentity assumes an IAM role
// configured via a ServiceAccount assume Cross account IAM roles
// https://aws.amazon.com/blogs/containers/cross-account-iam-roles-for-kubernetes-service-accounts/
func UsePodServiceAccountAssumeRoleWithWebIdentity(ctx context.Context, _ []byte, _, region string, pc *v1beta1.ProviderConfig) (*aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx, middlewareV2)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load default AWS config")
	}

	roleArn, err := GetAssumeRoleWithWebIdentityARN(pc.Spec.DeepCopy())
	if err != nil {
		return nil, err
	}

	stsclient := sts.NewFromConfig(cfg)
	webIdentityRoleOptions := SetWebIdentityRoleOptions(pc)

	cnf, err := config.LoadDefaultConfig(
		ctx,
		middlewareV2,
		config.WithRegion(region),
		config.WithCredentialsProvider(aws.NewCredentialsCache(
			stscreds.NewWebIdentityRoleProvider(
				stsclient,
				pointer.StringValue(roleArn),
				stscreds.IdentityTokenFile(getWebidentityTokenFilePath()),
				webIdentityRoleOptions,
			)),
		),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load assumed role AWS config")
	}
	return &cnf, err
}

const webIdentityTokenFileDefaultPath = "/var/run/secrets/eks.amazonaws.com/serviceaccount/token"

func getWebidentityTokenFilePath() string {
	if path := os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE"); path != "" {
		return path
	}
	return webIdentityTokenFileDefaultPath
}

// GetDefaultConfigV2 returns a shallow copy of a default SDK
// config. We use this to get a shared credentials cache.
func GetDefaultConfigV2(ctx context.Context) (aws.Config, error) {
	// TODO: Possible performance improvement by using an RWMutex and RLock
	//       to allow parallel copying.
	//       However, this would likely increase the complexity of the code.
	muV2.Lock()
	defer muV2.Unlock()

	if defaultConfigV2 == nil {
		cfg, err := config.LoadDefaultConfig(ctx, userAgentV2)
		if err != nil {
			return aws.Config{}, errors.Wrap(err, "failed to load default AWS config")
		}
		defaultConfigV2 = &cfg
	}

	return defaultConfigV2.Copy(), nil
}

// UsePodServiceAccount assumes an IAM role configured via a ServiceAccount.
// https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html
func UsePodServiceAccount(ctx context.Context, _ []byte, _, region string) (*aws.Config, error) {
	cfg, err := GetDefaultConfigV2(ctx)
	if err != nil {
		return nil, err
	}
	if region != GlobalRegion {
		cfg.Region = region
	}
	return &cfg, nil
}

// NOTE(muvaf): ACK-generated controllers use aws/aws-sdk-go instead of
// aws/aws-sdk-go-v2. These functions are implemented to be used by those controllers.

// GetConfigV1 constructs an *awsv1.Config that can be used to authenticate to AWS
// API by the AWSv1 clients.
func GetConfigV1(ctx context.Context, c client.Client, mg resource.Managed, region string) (*session.Session, error) { //nolint:gocyclo
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
		if pc.Spec.AssumeRoleARN != nil || pc.Spec.AssumeRole != nil {
			cfg, err := UsePodServiceAccountV1AssumeRole(ctx, []byte{}, pc, DefaultSection, region)
			if err != nil {
				return nil, errors.Wrap(err, "cannot use pod service account to assume role")
			}
			return GetSessionV1(cfg)
		}
		if pc.Spec.AssumeRoleWithWebIdentity != nil && pc.Spec.AssumeRoleWithWebIdentity.RoleARN != nil {
			cfg, err := UsePodServiceAccountV1AssumeRoleWithWebIdentity(ctx, []byte{}, pc, DefaultSection, region)
			if err != nil {
				return nil, err
			}
			return GetSessionV1(cfg)
		}
		cfg, err := UsePodServiceAccountV1(ctx, []byte{}, pc, DefaultSection, region)
		if err != nil {
			return nil, errors.Wrap(err, "cannot use pod service account")
		}
		return GetSessionV1(cfg)
	default:
		data, err := resource.CommonCredentialExtractor(ctx, s, c, pc.Spec.Credentials.CommonCredentialSelectors)
		if err != nil {
			return nil, errors.Wrap(err, "cannot get credentials")
		}

		if pc.Spec.AssumeRole != nil || pc.Spec.AssumeRoleARN != nil {
			cfg, err := UseProviderSecretV1AssumeRole(ctx, data, pc, DefaultSection, region)
			if err != nil {
				return nil, errors.Wrap(err, "cannot use secret")
			}
			return GetSessionV1(cfg)
		}
		cfg, err := UseProviderSecretV1(ctx, data, pc, DefaultSection, region)
		if err != nil {
			return nil, errors.Wrap(err, "cannot use secret")
		}
		return GetSessionV1(cfg)
	}
}

// GetSessionV1 constructs an AWS V1 client session, with common configuration like the user agent handler
func GetSessionV1(cfg *awsv1.Config) (*session.Session, error) {
	session, err := session.NewSession(cfg)
	if err != nil {
		return nil, err
	}
	session.Handlers.Build.PushBackNamed(userAgentV1)
	session.Handlers.Send.PushFront(func(r *requestv1.Request) {
		metrics.IncAWSAPICall(r.ClientInfo.ServiceName, r.Operation.Name, "1")
	})
	return session, nil
}

// UseProviderSecretV1AssumeRole - AWS v1 configuration which can be used to issue requests against AWS API
// assume Cross account IAM roles
func UseProviderSecretV1AssumeRole(ctx context.Context, data []byte, pc *v1beta1.ProviderConfig, profile, region string) (*awsv1.Config, error) {
	creds, err := CredentialsIDSecret(data, profile)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse credentials secret")
	}

	config, err := config.LoadDefaultConfig(
		ctx,
		middlewareV2,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: creds,
		}),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load credentials")
	}

	roleArn, err := GetAssumeRoleARN(pc.Spec.DeepCopy())
	if err != nil {
		return nil, errors.Wrap(err, "failed to assume IAM Role")
	}

	stsSvc := sts.NewFromConfig(config)
	stsAssumeRoleOptions := SetAssumeRoleOptions(pc)
	stsAssume := stscreds.NewAssumeRoleProvider(
		stsSvc,
		pointer.StringValue(roleArn),
		stsAssumeRoleOptions,
	)
	config.Credentials = aws.NewCredentialsCache(stsAssume)

	v2creds, err := config.Credentials.Retrieve(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve credentials")
	}

	v1creds := credentialsv1.NewStaticCredentials(
		v2creds.AccessKeyID,
		v2creds.SecretAccessKey,
		v2creds.SessionToken)

	return SetResolverV1(pc, awsv1.NewConfig().WithCredentials(v1creds).WithRegion(region)), nil
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

// UsePodServiceAccountV1AssumeRole assumes an IAM role configured via a ServiceAccount and
// assume Cross account IAM role
// https://aws.amazon.com/blogs/containers/cross-account-iam-roles-for-kubernetes-service-accounts/
func UsePodServiceAccountV1AssumeRole(ctx context.Context, _ []byte, pc *v1beta1.ProviderConfig, _, region string) (*awsv1.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx, middlewareV2)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load default AWS config")
	}

	roleArn, err := GetAssumeRoleARN(pc.Spec.DeepCopy())
	if err != nil {
		return nil, errors.Wrap(err, "failed to assume IAM Role")
	}
	stsclient := sts.NewFromConfig(cfg)
	stsAssumeRoleOptions := SetAssumeRoleOptions(pc)
	if region == GlobalRegion {
		region = cfg.Region
	}
	cnf, err := config.LoadDefaultConfig(
		ctx,
		middlewareV2,
		config.WithRegion(region),
		config.WithCredentialsProvider(aws.NewCredentialsCache(
			stscreds.NewAssumeRoleProvider(
				stsclient,
				pointer.StringValue(roleArn),
				stsAssumeRoleOptions,
			)),
		),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load assumed role AWS config")
	}
	v2creds, err := cnf.Credentials.Retrieve(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve credentials")
	}
	v1creds := credentialsv1.NewStaticCredentials(
		v2creds.AccessKeyID,
		v2creds.SecretAccessKey,
		v2creds.SessionToken)
	return SetResolverV1(pc, awsv1.NewConfig().WithCredentials(v1creds).WithRegion(region)), nil
}

// UsePodServiceAccountV1AssumeRoleWithWebIdentity assumes an IAM role configured via a ServiceAccount and
// assume Cross account IAM role
// https://aws.amazon.com/blogs/containers/cross-account-iam-roles-for-kubernetes-service-accounts/
func UsePodServiceAccountV1AssumeRoleWithWebIdentity(ctx context.Context, _ []byte, pc *v1beta1.ProviderConfig, _, region string) (*awsv1.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx, middlewareV2)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load default AWS config")
	}

	roleArn, err := GetAssumeRoleWithWebIdentityARN(pc.Spec.DeepCopy())
	if err != nil {
		return nil, errors.Wrap(err, "failed to get role arn for assume role with web identity")
	}

	stsclient := sts.NewFromConfig(cfg)
	webIdentityRoleOptions := SetWebIdentityRoleOptions(pc)

	cnf, err := config.LoadDefaultConfig(
		ctx,
		middlewareV2,
		config.WithRegion(region),
		config.WithCredentialsProvider(aws.NewCredentialsCache(
			stscreds.NewWebIdentityRoleProvider(
				stsclient,
				pointer.StringValue(roleArn),
				stscreds.IdentityTokenFile("/var/run/secrets/eks.amazonaws.com/serviceaccount/token"),
				webIdentityRoleOptions,
			)),
		),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load assumed role AWS config")
	}
	v2creds, err := cnf.Credentials.Retrieve(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve credentials")
	}
	v1creds := credentialsv1.NewStaticCredentials(
		v2creds.AccessKeyID,
		v2creds.SecretAccessKey,
		v2creds.SessionToken)
	return SetResolverV1(pc, awsv1.NewConfig().WithCredentials(v1creds).WithRegion(region)), nil
}

// GetDefaultConfigV1 returns a shallow copy of a default SDK
// config. We use this to get a shared credentials cache.
func GetDefaultConfigV1() (*awsv1.Config, error) {
	// TODO: Possible performance improvement by using an RWMutex and RLock
	//       to allow parallel copying.
	//       However, this would likely increase the complexity of the code.
	muV1.Lock()
	defer muV1.Unlock()
	if defaultConfigV1 == nil {
		cfg := awsv1.NewConfig()
		sess, err := GetSessionV1(cfg)
		if err != nil {
			return nil, errors.Wrap(err, "failed to load default AWS config")
		}
		envCfg, err := config.NewEnvConfig()
		if err != nil {
			return nil, errors.Wrap(err, "failed to load default AWS env config")
		}
		creds := stscredsv1.NewWebIdentityCredentials(sess, envCfg.RoleARN, envCfg.RoleSessionName, envCfg.WebIdentityTokenFilePath) //nolint:staticcheck
		defaultConfigV1 = cfg.WithCredentials(creds)
	}
	return defaultConfigV1.Copy(), nil
}

// UsePodServiceAccountV1 assumes an IAM role configured via a ServiceAccount.
// https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html
func UsePodServiceAccountV1(ctx context.Context, _ []byte, pc *v1beta1.ProviderConfig, _, region string) (*awsv1.Config, error) {
	cfg, err := GetDefaultConfigV1()
	if err != nil {
		return nil, err
	}
	if region != GlobalRegion {
		cfg = cfg.WithRegion(region)
	}
	return SetResolverV1(pc, cfg), nil
}

// SetResolverV1 parses annotations from the managed resource
// and returns a V1 configuration accordingly.
func SetResolverV1(pc *v1beta1.ProviderConfig, cfg *awsv1.Config) *awsv1.Config {
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
			fullURL = pointer.StringValue(pc.Spec.Endpoint.URL.Static)
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
			PartitionID:   pointer.StringValue(pc.Spec.Endpoint.PartitionID),
			SigningName:   pointer.StringValue(pc.Spec.Endpoint.SigningName),
			SigningRegion: pointer.StringValue(pointer.LateInitialize(pc.Spec.Endpoint.SigningRegion, &region)),
			SigningMethod: pointer.StringValue(pc.Spec.Endpoint.SigningMethod),
		}
		// Only IAM does not have a region parameter and "aws-global" is used in
		// SDK setup. However, signing region has to be us-east-1 and it needs
		// to be set.
		if region == "aws-global" {
			switch pointer.StringValue(pc.Spec.Endpoint.PartitionID) {
			case "aws-us-gov", "aws-cn", "aws-iso", "aws-iso-b":
				e.SigningRegion = pointer.StringValue(pointer.LateInitialize(pc.Spec.Endpoint.SigningRegion, &region))
			default:
				e.SigningRegion = "us-east-1"
			}
		}
		return e, nil
	})
	return cfg
}

// GetAssumeRoleARN gets the AssumeRoleArn from a ProviderConfigSpec
func GetAssumeRoleARN(pcs *v1beta1.ProviderConfigSpec) (*string, error) {
	if pcs.AssumeRole != nil && pointer.StringValue(pcs.AssumeRole.RoleARN) != "" {
		return pcs.AssumeRole.RoleARN, nil
	}

	// Deprecated. Use AssumeRole.RoleARN
	if pcs.AssumeRoleARN != nil {
		return pcs.AssumeRoleARN, nil
	}

	return nil, errors.New("a RoleARN must be set to assume an IAM Role")
}

// GetAssumeRoleWithWebIdentityARN gets the RoleArn from a ProviderConfigSpec
func GetAssumeRoleWithWebIdentityARN(pcs *v1beta1.ProviderConfigSpec) (*string, error) {
	if pcs.AssumeRoleWithWebIdentity != nil {
		if pcs.AssumeRoleWithWebIdentity.RoleARN != nil && pointer.StringValue(pcs.AssumeRoleWithWebIdentity.RoleARN) != "" {
			return pcs.AssumeRoleWithWebIdentity.RoleARN, nil
		}
	}

	return nil, errors.New("a RoleARN must be set to assume with web identity")
}

// SetAssumeRoleOptions sets options when Assuming an IAM Role
func SetAssumeRoleOptions(pc *v1beta1.ProviderConfig) func(*stscreds.AssumeRoleOptions) {
	if pc.Spec.AssumeRole != nil {
		return func(opt *stscreds.AssumeRoleOptions) {
			if pc.Spec.AssumeRole.ExternalID != nil {
				opt.ExternalID = pc.Spec.AssumeRole.ExternalID
			}

			if pc.Spec.AssumeRole.Tags != nil && len(pc.Spec.AssumeRole.Tags) > 0 {
				for _, t := range pc.Spec.AssumeRole.Tags {
					opt.Tags = append(
						opt.Tags,
						stscredstypesv2.Tag{Key: t.Key, Value: t.Value})
				}
			}

			if pc.Spec.AssumeRole.TransitiveTagKeys != nil && len(pc.Spec.AssumeRole.TransitiveTagKeys) > 0 {
				opt.TransitiveTagKeys = pc.Spec.AssumeRole.TransitiveTagKeys
			}
		}
	}

	// Deprecated. Use AssumeRole.ExternalID
	if pc.Spec.ExternalID != nil {
		return func(opt *stscreds.AssumeRoleOptions) { opt.ExternalID = pc.Spec.ExternalID }
	}

	return func(opt *stscreds.AssumeRoleOptions) {}
}

// SetWebIdentityRoleOptions sets options when exchanging a WebIdentity Token for a Role
func SetWebIdentityRoleOptions(pc *v1beta1.ProviderConfig) func(*stscreds.WebIdentityRoleOptions) {
	if pc.Spec.AssumeRoleWithWebIdentity != nil {
		return func(opt *stscreds.WebIdentityRoleOptions) {
			if pc.Spec.AssumeRoleWithWebIdentity.RoleSessionName != "" {
				opt.RoleSessionName = pc.Spec.AssumeRoleWithWebIdentity.RoleSessionName
			}
		}
	}

	return func(opt *stscreds.WebIdentityRoleOptions) {}
}
