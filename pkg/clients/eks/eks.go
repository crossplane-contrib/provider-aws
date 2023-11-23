/*
Copyright 2020 The Crossplane Authors.

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

package eks

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	ekstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"github.com/crossplane-contrib/provider-aws/apis/eks/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/jsonpatch"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

const (
	clusterIDHeader  = "x-k8s-aws-id"
	expireHeader     = "X-Amz-Expires"
	expireHeaderTime = "60"
	v1Prefix         = "k8s-aws-v1."
)

// Client defines EKS Client operations
type Client interface {
	CreateCluster(ctx context.Context, input *eks.CreateClusterInput, opts ...func(*eks.Options)) (*eks.CreateClusterOutput, error)
	DescribeCluster(ctx context.Context, input *eks.DescribeClusterInput, opts ...func(*eks.Options)) (*eks.DescribeClusterOutput, error)
	UpdateClusterConfig(ctx context.Context, input *eks.UpdateClusterConfigInput, opts ...func(*eks.Options)) (*eks.UpdateClusterConfigOutput, error)
	DeleteCluster(ctx context.Context, input *eks.DeleteClusterInput, opts ...func(*eks.Options)) (*eks.DeleteClusterOutput, error)
	TagResource(ctx context.Context, input *eks.TagResourceInput, opts ...func(*eks.Options)) (*eks.TagResourceOutput, error)
	UntagResource(ctx context.Context, input *eks.UntagResourceInput, opts ...func(*eks.Options)) (*eks.UntagResourceOutput, error)
	UpdateClusterVersion(ctx context.Context, input *eks.UpdateClusterVersionInput, opts ...func(*eks.Options)) (*eks.UpdateClusterVersionOutput, error)
	AssociateEncryptionConfig(ctx context.Context, params *eks.AssociateEncryptionConfigInput, optFns ...func(*eks.Options)) (*eks.AssociateEncryptionConfigOutput, error)

	DescribeNodegroup(ctx context.Context, input *eks.DescribeNodegroupInput, opts ...func(*eks.Options)) (*eks.DescribeNodegroupOutput, error)
	CreateNodegroup(ctx context.Context, input *eks.CreateNodegroupInput, opts ...func(*eks.Options)) (*eks.CreateNodegroupOutput, error)
	UpdateNodegroupVersion(ctx context.Context, input *eks.UpdateNodegroupVersionInput, opts ...func(*eks.Options)) (*eks.UpdateNodegroupVersionOutput, error)
	UpdateNodegroupConfig(ctx context.Context, input *eks.UpdateNodegroupConfigInput, opts ...func(*eks.Options)) (*eks.UpdateNodegroupConfigOutput, error)
	DeleteNodegroup(ctx context.Context, input *eks.DeleteNodegroupInput, opts ...func(*eks.Options)) (*eks.DeleteNodegroupOutput, error)

	DescribeFargateProfile(ctx context.Context, input *eks.DescribeFargateProfileInput, opts ...func(*eks.Options)) (*eks.DescribeFargateProfileOutput, error)
	CreateFargateProfile(ctx context.Context, input *eks.CreateFargateProfileInput, opts ...func(*eks.Options)) (*eks.CreateFargateProfileOutput, error)
	DeleteFargateProfile(ctx context.Context, input *eks.DeleteFargateProfileInput, opts ...func(*eks.Options)) (*eks.DeleteFargateProfileOutput, error)

	DescribeIdentityProviderConfig(ctx context.Context, input *eks.DescribeIdentityProviderConfigInput, opts ...func(*eks.Options)) (*eks.DescribeIdentityProviderConfigOutput, error)
	AssociateIdentityProviderConfig(ctx context.Context, input *eks.AssociateIdentityProviderConfigInput, opts ...func(*eks.Options)) (*eks.AssociateIdentityProviderConfigOutput, error)
	DisassociateIdentityProviderConfig(ctx context.Context, input *eks.DisassociateIdentityProviderConfigInput, opts ...func(*eks.Options)) (*eks.DisassociateIdentityProviderConfigOutput, error)
}

// STSClient STS presigner
type STSClient interface {
	PresignGetCallerIdentity(ctx context.Context, input *sts.GetCallerIdentityInput, opts ...func(*sts.PresignOptions)) (*v4.PresignedHTTPRequest, error)
}

// NewEKSClient creates new EKS Client with provided AWS Configurations/Credentials.
func NewEKSClient(cfg aws.Config) Client {
	return eks.NewFromConfig(cfg)
}

// NewSTSClient creates a new STS Client.
func NewSTSClient(cfg aws.Config) STSClient {
	return sts.NewPresignClient(sts.NewFromConfig(cfg))
}

// IsErrorNotFound helper function to test for ResourceNotFoundException error.
func IsErrorNotFound(err error) bool {
	var nfe *ekstypes.ResourceNotFoundException
	return errors.As(err, &nfe)
}

// IsErrorInUse helper function to test for eResourceInUseException error.
func IsErrorInUse(err error) bool {
	var iue *ekstypes.ResourceInUseException
	return errors.As(err, &iue)
}

// IsErrorInvalidRequest helper function to test for InvalidRequestException error.
func IsErrorInvalidRequest(err error) bool {
	var ire *ekstypes.InvalidRequestException
	return errors.As(err, &ire)
}

// GenerateCreateClusterInput from ClusterParameters.
func GenerateCreateClusterInput(name string, p *v1beta1.ClusterParameters) *eks.CreateClusterInput {
	c := &eks.CreateClusterInput{
		Name:    pointer.ToOrNilIfZeroValue(name),
		RoleArn: &p.RoleArn,
		Version: p.Version,
	}

	if len(p.EncryptionConfig) > 0 {
		c.EncryptionConfig = GenerateEncryptionConfig(p)
	}

	if p.KubernetesNetworkConfig != nil {
		c.KubernetesNetworkConfig = &ekstypes.KubernetesNetworkConfigRequest{
			IpFamily: ekstypes.IpFamily(p.KubernetesNetworkConfig.IPFamily),
		}
		if p.KubernetesNetworkConfig.ServiceIpv4Cidr != "" {
			c.KubernetesNetworkConfig.ServiceIpv4Cidr = pointer.ToOrNilIfZeroValue(p.KubernetesNetworkConfig.ServiceIpv4Cidr)
		}
	}

	c.ResourcesVpcConfig = &ekstypes.VpcConfigRequest{
		EndpointPrivateAccess: p.ResourcesVpcConfig.EndpointPrivateAccess,
		EndpointPublicAccess:  p.ResourcesVpcConfig.EndpointPublicAccess,
		PublicAccessCidrs:     p.ResourcesVpcConfig.PublicAccessCidrs,
		SecurityGroupIds:      p.ResourcesVpcConfig.SecurityGroupIDs,
		SubnetIds:             p.ResourcesVpcConfig.SubnetIDs,
	}

	if p.Logging != nil {
		c.Logging = &ekstypes.Logging{
			ClusterLogging: make([]ekstypes.LogSetup, len(p.Logging.ClusterLogging)),
		}
		for i, cl := range p.Logging.ClusterLogging {
			types := make([]ekstypes.LogType, len(cl.Types))
			for j, t := range cl.Types {
				types[j] = ekstypes.LogType(t)
			}
			c.Logging.ClusterLogging[i] = ekstypes.LogSetup{
				Enabled: cl.Enabled,
				Types:   types,
			}
		}
	}
	if len(p.Tags) != 0 {
		c.Tags = p.Tags
	}
	return c
}

// GenerateEncryptionConfig creates the config needed to enable encryption
func GenerateEncryptionConfig(parameters *v1beta1.ClusterParameters) []ekstypes.EncryptionConfig {
	encryptionConfig := make([]ekstypes.EncryptionConfig, len(parameters.EncryptionConfig))
	if len(parameters.EncryptionConfig) > 0 {
		for i, conf := range parameters.EncryptionConfig {
			encryptionConfig[i] = ekstypes.EncryptionConfig{
				Provider: &ekstypes.Provider{
					KeyArn: pointer.ToOrNilIfZeroValue(conf.Provider.KeyArn),
				},
				Resources: conf.Resources,
			}
		}
	}
	return encryptionConfig
}

// CreatePatch creates a *v1beta1.ClusterParameters that has only the changed
// values between the target *v1beta1.ClusterParameters and the current
// *ekstypes.Cluster.
func CreatePatch(in *ekstypes.Cluster, target *v1beta1.ClusterParameters) (*v1beta1.ClusterParameters, error) {
	currentParams := &v1beta1.ClusterParameters{}
	LateInitialize(currentParams, in)

	jsonPatch, err := jsonpatch.CreateJSONPatch(currentParams, target)
	if err != nil {
		return nil, err
	}
	patch := &v1beta1.ClusterParameters{}
	if err := json.Unmarshal(jsonPatch, patch); err != nil {
		return nil, err
	}
	return patch, nil
}

// GenerateUpdateClusterConfigInputForLogging from ClusterParameters.
func GenerateUpdateClusterConfigInputForLogging(name string, p *v1beta1.ClusterParameters) *eks.UpdateClusterConfigInput {
	u := &eks.UpdateClusterConfigInput{
		Name: pointer.ToOrNilIfZeroValue(name),
	}

	u.Logging = &ekstypes.Logging{
		ClusterLogging: make([]ekstypes.LogSetup, len(p.Logging.ClusterLogging)),
	}
	for i, cl := range p.Logging.ClusterLogging {
		types := make([]ekstypes.LogType, len(cl.Types))
		for j, t := range cl.Types {
			types[j] = ekstypes.LogType(t)
		}
		u.Logging.ClusterLogging[i] = ekstypes.LogSetup{
			Enabled: cl.Enabled,
			Types:   types,
		}
	}
	return u
}

// GenerateUpdateClusterConfigInputForVPC from ClusterParameters.
func GenerateUpdateClusterConfigInputForVPC(name string, p *v1beta1.ClusterParameters) *eks.UpdateClusterConfigInput {
	u := &eks.UpdateClusterConfigInput{
		Name: pointer.ToOrNilIfZeroValue(name),
	}

	// NOTE(muvaf): SecurityGroupIds and SubnetIds cannot be updated. They are
	// included in VpcConfigRequest probably because it is used in Create call
	// as well.
	u.ResourcesVpcConfig = &ekstypes.VpcConfigRequest{
		EndpointPrivateAccess: p.ResourcesVpcConfig.EndpointPrivateAccess,
		EndpointPublicAccess:  p.ResourcesVpcConfig.EndpointPublicAccess,
		PublicAccessCidrs:     p.ResourcesVpcConfig.PublicAccessCidrs,
	}
	return u
}

// GenerateObservation is used to produce v1beta1.ClusterObservation from
// ekstypes.Cluster.
func GenerateObservation(cluster *ekstypes.Cluster) v1beta1.ClusterObservation {
	if cluster == nil {
		return v1beta1.ClusterObservation{}
	}
	o := v1beta1.ClusterObservation{
		Arn:             pointer.StringValue(cluster.Arn),
		Endpoint:        pointer.StringValue(cluster.Endpoint),
		PlatformVersion: pointer.StringValue(cluster.PlatformVersion),
		Version:         pointer.StringValue(cluster.Version),
		Status:          v1beta1.ClusterStatusType(cluster.Status),
	}

	if cluster.CreatedAt != nil {
		o.CreatedAt = &metav1.Time{Time: *cluster.CreatedAt}
	}

	if cluster.Identity != nil && cluster.Identity.Oidc != nil {
		o.Identity = v1beta1.Identity{
			OIDC: v1beta1.OIDC{
				Issuer: pointer.StringValue(cluster.Identity.Oidc.Issuer),
			},
		}
	}

	if cluster.OutpostConfig != nil {
		o.OutpostConfig = v1beta1.OutpostConfigResponse{
			ControlPlaneInstanceType: pointer.StringValue(cluster.OutpostConfig.ControlPlaneInstanceType),
			OutpostArns:              cluster.OutpostConfig.OutpostArns,
		}
	}

	if cluster.KubernetesNetworkConfig != nil {
		o.KubernetesNetworkConfig = v1beta1.KubernetesNetworkConfigResponse{
			IPFamily:        v1beta1.IPFamily(cluster.KubernetesNetworkConfig.IpFamily),
			ServiceIpv4Cidr: pointer.StringValue(cluster.KubernetesNetworkConfig.ServiceIpv4Cidr),
			ServiceIpv6Cidr: pointer.StringValue(cluster.KubernetesNetworkConfig.ServiceIpv6Cidr),
		}
	}

	if cluster.ResourcesVpcConfig != nil {
		o.ResourcesVpcConfig = v1beta1.VpcConfigResponse{
			ClusterSecurityGroupID: pointer.StringValue(cluster.ResourcesVpcConfig.ClusterSecurityGroupId),
			VpcID:                  pointer.StringValue(cluster.ResourcesVpcConfig.VpcId),
		}
	}

	if cluster.CertificateAuthority != nil {
		o.CertificateAuthorityData = pointer.StringValue(cluster.CertificateAuthority.Data)
	}
	return o
}

// LateInitialize fills the empty fields in *v1beta1.ClusterParameters with the
// values seen in ekstypes.Cluster.
func LateInitialize(in *v1beta1.ClusterParameters, cluster *ekstypes.Cluster) { //nolint:gocyclo
	if cluster == nil {
		return
	}
	if len(in.EncryptionConfig) == 0 && len(cluster.EncryptionConfig) > 0 {
		in.EncryptionConfig = make([]v1beta1.EncryptionConfig, len(cluster.EncryptionConfig))
		for i, e := range cluster.EncryptionConfig {
			in.EncryptionConfig[i] = v1beta1.EncryptionConfig{
				Resources: e.Resources,
			}
			if e.Provider != nil {
				in.EncryptionConfig[i].Provider = v1beta1.Provider{
					KeyArn: *e.Provider.KeyArn,
				}
			}
		}
	}
	if in.Logging == nil && cluster.Logging != nil && len(cluster.Logging.ClusterLogging) > 0 {
		in.Logging = &v1beta1.Logging{
			ClusterLogging: make([]v1beta1.LogSetup, len(cluster.Logging.ClusterLogging)),
		}
		for i, cl := range cluster.Logging.ClusterLogging {
			types := make([]v1beta1.LogType, len(cl.Types))
			for j, t := range cl.Types {
				types[j] = v1beta1.LogType(t)
			}
			in.Logging.ClusterLogging[i] = v1beta1.LogSetup{
				Enabled: cl.Enabled,
				Types:   types,
			}
		}
	}
	if cluster.OutpostConfig != nil {
		in.OutpostConfig.ControlPlaneInstanceType = pointer.StringValue(cluster.OutpostConfig.ControlPlaneInstanceType)
		if len(in.OutpostConfig.OutpostArns) == 0 && len(cluster.OutpostConfig.OutpostArns) > 0 {
			in.OutpostConfig.OutpostArns = cluster.OutpostConfig.OutpostArns
		}
	}
	if cluster.ResourcesVpcConfig != nil {
		in.ResourcesVpcConfig.EndpointPrivateAccess = pointer.LateInitialize(in.ResourcesVpcConfig.EndpointPrivateAccess, &cluster.ResourcesVpcConfig.EndpointPrivateAccess)
		in.ResourcesVpcConfig.EndpointPublicAccess = pointer.LateInitialize(in.ResourcesVpcConfig.EndpointPublicAccess, &cluster.ResourcesVpcConfig.EndpointPublicAccess)
		if len(in.ResourcesVpcConfig.PublicAccessCidrs) == 0 && len(cluster.ResourcesVpcConfig.PublicAccessCidrs) > 0 {
			in.ResourcesVpcConfig.PublicAccessCidrs = cluster.ResourcesVpcConfig.PublicAccessCidrs
		}

		if len(in.ResourcesVpcConfig.SecurityGroupIDs) == 0 && len(cluster.ResourcesVpcConfig.SecurityGroupIds) > 0 {
			in.ResourcesVpcConfig.SecurityGroupIDs = cluster.ResourcesVpcConfig.SecurityGroupIds
		}
		if len(in.ResourcesVpcConfig.SubnetIDs) == 0 && len(cluster.ResourcesVpcConfig.SubnetIds) > 0 {
			in.ResourcesVpcConfig.SubnetIDs = cluster.ResourcesVpcConfig.SubnetIds
		}
	}
	if in.KubernetesNetworkConfig == nil && cluster.KubernetesNetworkConfig != nil {
		in.KubernetesNetworkConfig = &v1beta1.KubernetesNetworkConfigRequest{
			ServiceIpv4Cidr: pointer.StringValue(cluster.KubernetesNetworkConfig.ServiceIpv4Cidr),
			IPFamily:        v1beta1.IPFamily(cluster.KubernetesNetworkConfig.IpFamily),
		}
	}

	in.RoleArn = pointer.LateInitializeValueFromPtr(in.RoleArn, cluster.RoleArn)
	in.Version = pointer.LateInitialize(in.Version, cluster.Version)
	// NOTE(hasheddan): we always will set the default Crossplane tags in
	// practice during initialization in the controller, but we check if no tags
	// exist for consistency with expected late initialization behavior.
	if in.Tags == nil {
		in.Tags = cluster.Tags
	}
}

// IsUpToDate checks whether there is a change in any of the modifiable fields.
func IsUpToDate(p *v1beta1.ClusterParameters, cluster *ekstypes.Cluster) (bool, error) {
	patch, err := CreatePatch(cluster, p)
	if err != nil {
		return false, err
	}

	// NOTE(hasheddan): AWS removes insignificant bits from CIDRs, so we must
	// compare by converting user-supplied CIDRs to network blocks. We only skip
	// comparison if both external and local have no CIDR blocks defined.
	if (cluster.ResourcesVpcConfig != nil && len(cluster.ResourcesVpcConfig.PublicAccessCidrs) > 0) || len(p.ResourcesVpcConfig.PublicAccessCidrs) > 0 {
		// Convert user-supplied slice of CIDRs to map of networks.
		netMap := map[string]bool{}
		for _, c := range p.ResourcesVpcConfig.PublicAccessCidrs {
			_, ipNet, err := net.ParseCIDR(c)
			if err != nil {
				return false, err
			}
			netMap[ipNet.String()] = true
		}
		// If length of networks does not match the length of CIDR blocks
		// returned by AWS then we need update.
		if len(netMap) != len(cluster.ResourcesVpcConfig.PublicAccessCidrs) {
			return false, nil
		}
		// If AWS returns a CIDR block that is not in the map, then we need
		// update.
		for _, pc := range cluster.ResourcesVpcConfig.PublicAccessCidrs {
			if !netMap[pc] {
				return false, nil
			}
		}
	}
	res := cmp.Equal(&v1beta1.ClusterParameters{}, patch, cmpopts.EquateEmpty(),
		cmpopts.IgnoreTypes(&xpv1.Reference{}, &xpv1.Selector{}, []xpv1.Reference{}),
		cmpopts.IgnoreFields(v1beta1.ClusterParameters{}, "Region"),
		cmpopts.IgnoreFields(v1beta1.VpcConfigRequest{}, "PublicAccessCidrs", "SubnetIDs", "SecurityGroupIDs"))
	return res, nil
}

// GetConnectionDetails extracts managed.ConnectionDetails out of ekstypes.Cluster.
func GetConnectionDetails(ctx context.Context, cluster *ekstypes.Cluster, stsClient STSClient) managed.ConnectionDetails {
	if cluster == nil || cluster.Name == nil || cluster.Endpoint == nil || cluster.CertificateAuthority == nil || cluster.CertificateAuthority.Data == nil {
		return managed.ConnectionDetails{}
	}

	getCallerIdentity, _ := stsClient.PresignGetCallerIdentity(ctx, &sts.GetCallerIdentityInput{},
		func(po *sts.PresignOptions) {
			po.ClientOptions = []func(*sts.Options){
				sts.WithAPIOptions(
					smithyhttp.AddHeaderValue(clusterIDHeader, *cluster.Name),
					smithyhttp.AddHeaderValue(expireHeader, expireHeaderTime), // otherwise we get in authenticator log invalid X-Amz-Expires parameter in pre-signed URL: 0
				),
			}
		},
	)

	// NOTE(hasheddan): This is carried over from the v1alpha3 version of the
	// EKS cluster resource. Signing the URL means that anyone in possession of
	// this Kubeconfig will now be able to access the EKS cluster until this URL
	// expires. This is necessary for other systems, such as core Crossplane, to
	// be able to schedule workloads to the cluster for now, but is not the most
	// secure way of accessing the cluster.
	// More information: https://docs.aws.amazon.com/eks/latest/userguide/create-kubeconfig.html
	token := v1Prefix + base64.RawURLEncoding.EncodeToString([]byte(getCallerIdentity.URL))

	// NOTE(hasheddan): We must decode the CA data before constructing our
	// Kubeconfig, as the raw Kubeconfig will be base64 encoded again when
	// written as a Secret.
	caData, err := base64.StdEncoding.DecodeString(*cluster.CertificateAuthority.Data)
	if err != nil {
		return managed.ConnectionDetails{}
	}
	kc := clientcmdapi.Config{
		Clusters: map[string]*clientcmdapi.Cluster{
			*cluster.Name: {
				Server:                   *cluster.Endpoint,
				CertificateAuthorityData: caData,
			},
		},
		Contexts: map[string]*clientcmdapi.Context{
			*cluster.Name: {
				Cluster:  *cluster.Name,
				AuthInfo: *cluster.Name,
			},
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			*cluster.Name: {
				Token: token,
			},
		},
		CurrentContext: *cluster.Name,
	}

	rawConfig, err := clientcmd.Write(kc)
	if err != nil {
		return managed.ConnectionDetails{}
	}
	return managed.ConnectionDetails{
		xpv1.ResourceCredentialsSecretEndpointKey:   []byte(*cluster.Endpoint),
		xpv1.ResourceCredentialsSecretKubeconfigKey: rawConfig,
		xpv1.ResourceCredentialsSecretCAKey:         caData,
	}
}
