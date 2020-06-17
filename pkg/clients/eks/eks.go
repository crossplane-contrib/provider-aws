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
	"net"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/eksiface"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/stsiface"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"

	"github.com/crossplane/provider-aws/apis/eks/v1beta1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

const (
	clusterIDHeader = "x-k8s-aws-id"
	v1Prefix        = "k8s-aws-v1."
)

// Client defines EKS Client operations
type Client eksiface.ClientAPI

// STSClient defines STS Client operations
type STSClient stsiface.ClientAPI

// NewClient creates new EKS Client with provided AWS Configurations/Credentials.
func NewClient(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (Client, STSClient, error) {
	cfg, err := auth(ctx, credentials, awsclients.DefaultSection, region)
	if cfg == nil {
		return nil, nil, err
	}
	return eks.New(*cfg), sts.New(*cfg), err
}

// IsErrorNotFound helper function to test for ErrCodeResourceNotFoundException error.
func IsErrorNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), eks.ErrCodeResourceNotFoundException)
}

// IsErrorInUse helper function to test for ErrCodeResourceInUseException error.
func IsErrorInUse(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), eks.ErrCodeResourceInUseException)
}

// GenerateCreateClusterInput from ClusterParameters.
func GenerateCreateClusterInput(name string, p *v1beta1.ClusterParameters) *eks.CreateClusterInput {
	c := &eks.CreateClusterInput{
		Name:    awsclients.String(name),
		RoleArn: p.RoleArn,
		Version: p.Version,
	}

	if len(p.EncryptionConfig) > 0 {
		c.EncryptionConfig = make([]eks.EncryptionConfig, len(p.EncryptionConfig))
		for i, conf := range p.EncryptionConfig {
			c.EncryptionConfig[i] = eks.EncryptionConfig{
				Provider: &eks.Provider{
					KeyArn: awsclients.String(conf.Provider.KeyArn),
				},
				Resources: conf.Resources,
			}
		}
	}

	c.ResourcesVpcConfig = &eks.VpcConfigRequest{
		EndpointPrivateAccess: p.ResourcesVpcConfig.EndpointPrivateAccess,
		EndpointPublicAccess:  p.ResourcesVpcConfig.EndpointPublicAccess,
		PublicAccessCidrs:     p.ResourcesVpcConfig.PublicAccessCidrs,
		SecurityGroupIds:      p.ResourcesVpcConfig.SecurityGroupIDs,
		SubnetIds:             p.ResourcesVpcConfig.SubnetIDs,
	}

	if p.Logging != nil {
		c.Logging = &eks.Logging{
			ClusterLogging: make([]eks.LogSetup, len(p.Logging.ClusterLogging)),
		}
		for i, cl := range p.Logging.ClusterLogging {
			types := make([]eks.LogType, len(cl.Types))
			for j, t := range cl.Types {
				types[j] = eks.LogType(t)
			}
			c.Logging.ClusterLogging[i] = eks.LogSetup{
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

// CreatePatch creates a *v1beta1.ClusterParameters that has only the changed
// values between the target *v1beta1.ClusterParameters and the current
// *eks.Cluster.
func CreatePatch(in *eks.Cluster, target *v1beta1.ClusterParameters) (*v1beta1.ClusterParameters, error) {
	currentParams := &v1beta1.ClusterParameters{}
	LateInitialize(currentParams, in)

	jsonPatch, err := awsclients.CreateJSONPatch(currentParams, target)
	if err != nil {
		return nil, err
	}
	patch := &v1beta1.ClusterParameters{}
	if err := json.Unmarshal(jsonPatch, patch); err != nil {
		return nil, err
	}
	return patch, nil
}

// GenerateUpdateClusterConfigInput from ClusterParameters.
func GenerateUpdateClusterConfigInput(name string, p *v1beta1.ClusterParameters) *eks.UpdateClusterConfigInput {
	u := &eks.UpdateClusterConfigInput{
		Name: awsclients.String(name),
	}

	if p.Logging != nil {
		u.Logging = &eks.Logging{
			ClusterLogging: make([]eks.LogSetup, len(p.Logging.ClusterLogging)),
		}
		for i, cl := range p.Logging.ClusterLogging {
			types := make([]eks.LogType, len(cl.Types))
			for j, t := range cl.Types {
				types[j] = eks.LogType(t)
			}
			u.Logging.ClusterLogging[i] = eks.LogSetup{
				Enabled: cl.Enabled,
				Types:   types,
			}
		}
	}

	u.ResourcesVpcConfig = &eks.VpcConfigRequest{
		EndpointPrivateAccess: p.ResourcesVpcConfig.EndpointPrivateAccess,
		EndpointPublicAccess:  p.ResourcesVpcConfig.EndpointPublicAccess,
		PublicAccessCidrs:     p.ResourcesVpcConfig.PublicAccessCidrs,
		SecurityGroupIds:      p.ResourcesVpcConfig.SecurityGroupIDs,
		SubnetIds:             p.ResourcesVpcConfig.SubnetIDs,
	}
	return u
}

// GenerateObservation is used to produce v1beta1.ClusterObservation from
// eks.Cluster.
func GenerateObservation(cluster *eks.Cluster) v1beta1.ClusterObservation { // nolint:gocyclo
	if cluster == nil {
		return v1beta1.ClusterObservation{}
	}
	o := v1beta1.ClusterObservation{
		Arn:             awsclients.StringValue(cluster.Arn),
		Endpoint:        awsclients.StringValue(cluster.Endpoint),
		PlatformVersion: awsclients.StringValue(cluster.PlatformVersion),
		Status:          v1beta1.ClusterStatusType(cluster.Status),
	}

	if cluster.CreatedAt != nil {
		o.CreatedAt = &metav1.Time{Time: *cluster.CreatedAt}
	}

	if cluster.Identity != nil && cluster.Identity.Oidc != nil {
		o.Identity = v1beta1.Identity{
			OIDC: v1beta1.OIDC{
				Issuer: awsclients.StringValue(cluster.Identity.Oidc.Issuer),
			},
		}
	}

	if cluster.ResourcesVpcConfig != nil {
		o.ResourcesVpcConfig = v1beta1.VpcConfigResponse{
			ClusterSecurityGroupID: awsclients.StringValue(cluster.ResourcesVpcConfig.ClusterSecurityGroupId),
			VpcID:                  awsclients.StringValue(cluster.ResourcesVpcConfig.VpcId),
		}
	}
	return o
}

// LateInitialize fills the empty fields in *v1beta1.ClusterParameters with the
// values seen in eks.Cluster.
func LateInitialize(in *v1beta1.ClusterParameters, cluster *eks.Cluster) { // nolint:gocyclo
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
	if cluster.ResourcesVpcConfig != nil {
		in.ResourcesVpcConfig.EndpointPrivateAccess = awsclients.LateInitializeBoolPtr(in.ResourcesVpcConfig.EndpointPrivateAccess, cluster.ResourcesVpcConfig.EndpointPrivateAccess)
		in.ResourcesVpcConfig.EndpointPublicAccess = awsclients.LateInitializeBoolPtr(in.ResourcesVpcConfig.EndpointPublicAccess, cluster.ResourcesVpcConfig.EndpointPublicAccess)
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
	in.RoleArn = awsclients.LateInitializeStringPtr(in.RoleArn, cluster.RoleArn)
	in.Version = awsclients.LateInitializeStringPtr(in.Version, cluster.Version)
	// NOTE(hasheddan): we always will set the default Crossplane tags in
	// practice during initialization in the controller, but we check if no tags
	// exist for consistency with expected late initialization behavior.
	if in.Tags == nil {
		in.Tags = cluster.Tags
	}
}

// IsUpToDate checks whether there is a change in any of the modifiable fields.
func IsUpToDate(p *v1beta1.ClusterParameters, cluster *eks.Cluster) (bool, error) {
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

	return cmp.Equal(&v1beta1.ClusterParameters{}, patch, cmpopts.EquateEmpty(),
		cmpopts.IgnoreTypes(&v1alpha1.Reference{}, &v1alpha1.Selector{}),
		cmpopts.IgnoreFields(v1beta1.VpcConfigRequest{}, "SecurityGroupIDRefs", "SubnetIDRefs", "PublicAccessCidrs")), nil
}

// GetConnectionDetails extracts managed.ConnectionDetails out of eks.Cluster.
func GetConnectionDetails(cluster *eks.Cluster, stsClient STSClient) managed.ConnectionDetails {
	if cluster == nil || cluster.Name == nil || cluster.Endpoint == nil || cluster.CertificateAuthority == nil || cluster.CertificateAuthority.Data == nil {
		return managed.ConnectionDetails{}
	}
	request := stsClient.GetCallerIdentityRequest(&sts.GetCallerIdentityInput{})
	request.HTTPRequest.Header.Add(clusterIDHeader, *cluster.Name)

	// NOTE(hasheddan): This is carried over from the v1alpha3 version of the
	// EKS cluster resource. Signing the URL means that anyone in possession of
	// this Kubeconfig will now be able to access the EKS cluster until this URL
	// expires. This is necessary for other systems, such as core Crossplane, to
	// be able to schedule workloads to the cluster for now, but is not the most
	// secure way of accessing the cluster.
	// More information: https://docs.aws.amazon.com/eks/latest/userguide/create-kubeconfig.html
	presignedURLString, err := request.Presign(60 * time.Second)
	if err != nil {
		return managed.ConnectionDetails{}
	}
	token := v1Prefix + base64.RawURLEncoding.EncodeToString([]byte(presignedURLString))

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
		v1alpha1.ResourceCredentialsSecretEndpointKey:   []byte(*cluster.Endpoint),
		v1alpha1.ResourceCredentialsSecretKubeconfigKey: rawConfig,
		v1alpha1.ResourceCredentialsSecretCAKey:         caData,
	}
}
