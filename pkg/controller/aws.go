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

package controller

import (
	"github.com/crossplane/provider-aws/pkg/controller/vpcpeering"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/crossplane-runtime/pkg/logging"

	"github.com/crossplane/provider-aws/pkg/controller/config"
	"github.com/crossplane/provider-aws/pkg/controller/ec2/vpc"
	"github.com/crossplane/provider-aws/pkg/controller/elasticloadbalancing/elb"
	"github.com/crossplane/provider-aws/pkg/controller/elasticloadbalancing/elbattachment"
	"github.com/crossplane/provider-aws/pkg/controller/identity/iampolicy"
	"github.com/crossplane/provider-aws/pkg/controller/identity/iamrole"
	"github.com/crossplane/provider-aws/pkg/controller/identity/iamrolepolicyattachment"
	"github.com/crossplane/provider-aws/pkg/controller/kms/key"
	"github.com/crossplane/provider-aws/pkg/controller/route53/hostedzone"
	"github.com/crossplane/provider-aws/pkg/controller/route53/resourcerecordset"
	"github.com/crossplane/provider-aws/pkg/controller/s3"
	"github.com/crossplane/provider-aws/pkg/controller/s3/bucketpolicy"
	"github.com/crossplane/provider-aws/pkg/controller/sqs/queue"
)

// Setup creates all AWS controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	for _, setup := range []func(ctrl.Manager, logging.Logger, workqueue.RateLimiter) error{
		config.Setup,
		//cache.SetupReplicationGroup,
		//cachesubnetgroup.SetupCacheSubnetGroup,
		//cluster.SetupCacheCluster,
		//database.SetupRDSInstance,
		//eks.SetupCluster,
		elb.SetupELB,
		elbattachment.SetupELBAttachment,
		//nodegroup.SetupNodeGroup,
		s3.SetupBucket,
		bucketpolicy.SetupBucketPolicy,
		//iamaccesskey.SetupIAMAccessKey,
		//iamuser.SetupIAMUser,
		//iamgroup.SetupIAMGroup,
		iampolicy.SetupIAMPolicy,
		iamrole.SetupIAMRole,
		//iamgroupusermembership.SetupIAMGroupUserMembership,
		//iamuserpolicyattachment.SetupIAMUserPolicyAttachment,
		//iamgrouppolicyattachment.SetupIAMGroupPolicyAttachment,
		iamrolepolicyattachment.SetupIAMRolePolicyAttachment,
		vpc.SetupVPC,
		//subnet.SetupSubnet,
		//securitygroup.SetupSecurityGroup,
		//internetgateway.SetupInternetGateway,
		//natgateway.SetupNatGateway,
		//routetable.SetupRouteTable,
		//dbsubnetgroup.SetupDBSubnetGroup,
		//certificateauthority.SetupCertificateAuthority,
		//certificateauthoritypermission.SetupCertificateAuthorityPermission,
		//acm.SetupCertificate,
		resourcerecordset.SetupResourceRecordSet,
		hostedzone.SetupHostedZone,
		//secret.SetupSecret,
		//snstopic.SetupSNSTopic,
		//snssubscription.SetupSubscription,
		queue.SetupQueue,
		//redshift.SetupCluster,
		//address.SetupAddress,
		//repository.SetupRepository,
		//repositorypolicy.SetupRepositoryPolicy,
		//api.SetupAPI,
		//stage.SetupStage,
		//route.SetupRoute,
		//authorizer.SetupAuthorizer,
		//integration.SetupIntegration,
		//deployment.SetupDeployment,
		//domainname.SetupDomainName,
		//integrationresponse.SetupIntegrationResponse,
		//model.SetupModel,
		//apimapping.SetupAPIMapping,
		//routeresponse.SetupRouteResponse,
		//vpclink.SetupVPCLink,
		//fargateprofile.SetupFargateProfile,
		//activity.SetupActivity,
		//statemachine.SetupStateMachine,
		//table.SetupTable,
		//backup.SetupBackup,
		//globaltable.SetupGlobalTable,
		key.SetupKey,
		//filesystem.SetupFileSystem,
		//dbcluster.SetupDBCluster,
		//dbparametergroup.SetupDBParameterGroup,
		//globalcluster.SetupGlobalCluster,
		//vpccidrblock.SetupVPCCIDRBlock,
		//privatednsnamespace.SetupPrivateDNSNamespace,
		//publicdnsnamespace.SetupPublicDNSNamespace,
		//httpnamespace.SetupHTTPNamespace,
		//function.SetupFunction,
		//openidconnectprovider.SetupOpenIDConnectProvider,
		vpcpeering.SetupVPCPeeringConnection,
	} {
		if err := setup(mgr, l, rl); err != nil {
			return err
		}
	}

	return nil
}
