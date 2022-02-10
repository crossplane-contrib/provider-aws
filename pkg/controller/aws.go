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
	"time"

	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/crossplane-runtime/pkg/logging"

	"github.com/crossplane/provider-aws/pkg/controller/acm"
	"github.com/crossplane/provider-aws/pkg/controller/acmpca/certificateauthority"
	"github.com/crossplane/provider-aws/pkg/controller/acmpca/certificateauthoritypermission"
	"github.com/crossplane/provider-aws/pkg/controller/apigatewayv2/api"
	"github.com/crossplane/provider-aws/pkg/controller/apigatewayv2/apimapping"
	"github.com/crossplane/provider-aws/pkg/controller/apigatewayv2/authorizer"
	"github.com/crossplane/provider-aws/pkg/controller/apigatewayv2/deployment"
	"github.com/crossplane/provider-aws/pkg/controller/apigatewayv2/domainname"
	"github.com/crossplane/provider-aws/pkg/controller/apigatewayv2/integration"
	"github.com/crossplane/provider-aws/pkg/controller/apigatewayv2/integrationresponse"
	"github.com/crossplane/provider-aws/pkg/controller/apigatewayv2/model"
	"github.com/crossplane/provider-aws/pkg/controller/apigatewayv2/route"
	"github.com/crossplane/provider-aws/pkg/controller/apigatewayv2/routeresponse"
	"github.com/crossplane/provider-aws/pkg/controller/apigatewayv2/stage"
	"github.com/crossplane/provider-aws/pkg/controller/apigatewayv2/vpclink"
	athenaworkgroup "github.com/crossplane/provider-aws/pkg/controller/athena/workgroup"
	"github.com/crossplane/provider-aws/pkg/controller/cache"
	"github.com/crossplane/provider-aws/pkg/controller/cache/cachesubnetgroup"
	"github.com/crossplane/provider-aws/pkg/controller/cache/cluster"
	"github.com/crossplane/provider-aws/pkg/controller/cloudfront/cachepolicy"
	cloudfrontorginaccessidentity "github.com/crossplane/provider-aws/pkg/controller/cloudfront/cloudfrontoriginaccessidentity"
	"github.com/crossplane/provider-aws/pkg/controller/cloudfront/distribution"
	cwloggroup "github.com/crossplane/provider-aws/pkg/controller/cloudwatchlogs/loggroup"
	"github.com/crossplane/provider-aws/pkg/controller/config"
	"github.com/crossplane/provider-aws/pkg/controller/database"
	"github.com/crossplane/provider-aws/pkg/controller/database/dbsubnetgroup"
	docdbcluster "github.com/crossplane/provider-aws/pkg/controller/docdb/dbcluster"
	docdbclusterparametergroup "github.com/crossplane/provider-aws/pkg/controller/docdb/dbclusterparametergroup"
	docdbinstance "github.com/crossplane/provider-aws/pkg/controller/docdb/dbinstance"
	docdbsubnetgroup "github.com/crossplane/provider-aws/pkg/controller/docdb/dbsubnetgroup"
	"github.com/crossplane/provider-aws/pkg/controller/dynamodb/backup"
	"github.com/crossplane/provider-aws/pkg/controller/dynamodb/globaltable"
	"github.com/crossplane/provider-aws/pkg/controller/dynamodb/table"
	"github.com/crossplane/provider-aws/pkg/controller/ec2/address"
	"github.com/crossplane/provider-aws/pkg/controller/ec2/instance"
	"github.com/crossplane/provider-aws/pkg/controller/ec2/internetgateway"
	"github.com/crossplane/provider-aws/pkg/controller/ec2/launchtemplate"
	"github.com/crossplane/provider-aws/pkg/controller/ec2/launchtemplateversion"
	"github.com/crossplane/provider-aws/pkg/controller/ec2/natgateway"
	ec2route "github.com/crossplane/provider-aws/pkg/controller/ec2/route"
	"github.com/crossplane/provider-aws/pkg/controller/ec2/routetable"
	"github.com/crossplane/provider-aws/pkg/controller/ec2/securitygroup"
	"github.com/crossplane/provider-aws/pkg/controller/ec2/subnet"
	transitgateway "github.com/crossplane/provider-aws/pkg/controller/ec2/transitgateway"
	transitgatewayroute "github.com/crossplane/provider-aws/pkg/controller/ec2/transitgatewayroute"
	transitgatewayroutetable "github.com/crossplane/provider-aws/pkg/controller/ec2/transitgatewayroutetable"
	transitgatewayvpcattachment "github.com/crossplane/provider-aws/pkg/controller/ec2/transitgatewayvpcattachment"
	"github.com/crossplane/provider-aws/pkg/controller/ec2/volume"
	"github.com/crossplane/provider-aws/pkg/controller/ec2/vpc"
	"github.com/crossplane/provider-aws/pkg/controller/ec2/vpccidrblock"
	"github.com/crossplane/provider-aws/pkg/controller/ec2/vpcendpoint"
	"github.com/crossplane/provider-aws/pkg/controller/ec2/vpcendpointserviceconfiguration"
	"github.com/crossplane/provider-aws/pkg/controller/ec2/vpcpeeringconnection"
	"github.com/crossplane/provider-aws/pkg/controller/ecr/repository"
	"github.com/crossplane/provider-aws/pkg/controller/ecr/repositorypolicy"
	"github.com/crossplane/provider-aws/pkg/controller/efs/filesystem"
	efsmounttarget "github.com/crossplane/provider-aws/pkg/controller/efs/mounttarget"
	"github.com/crossplane/provider-aws/pkg/controller/eks"
	eksaddon "github.com/crossplane/provider-aws/pkg/controller/eks/addon"
	"github.com/crossplane/provider-aws/pkg/controller/eks/fargateprofile"
	"github.com/crossplane/provider-aws/pkg/controller/eks/identityproviderconfig"
	"github.com/crossplane/provider-aws/pkg/controller/eks/nodegroup"
	"github.com/crossplane/provider-aws/pkg/controller/elasticache/cacheparametergroup"
	"github.com/crossplane/provider-aws/pkg/controller/elasticloadbalancing/elb"
	"github.com/crossplane/provider-aws/pkg/controller/elasticloadbalancing/elbattachment"
	"github.com/crossplane/provider-aws/pkg/controller/elbv2/listener"
	"github.com/crossplane/provider-aws/pkg/controller/elbv2/loadbalancer"
	"github.com/crossplane/provider-aws/pkg/controller/elbv2/targetgroup"
	glueclassifier "github.com/crossplane/provider-aws/pkg/controller/glue/classifier"
	glueconnection "github.com/crossplane/provider-aws/pkg/controller/glue/connection"
	gluecrawler "github.com/crossplane/provider-aws/pkg/controller/glue/crawler"
	glueDatabase "github.com/crossplane/provider-aws/pkg/controller/glue/database"
	gluejob "github.com/crossplane/provider-aws/pkg/controller/glue/job"
	gluesecurityconfiguration "github.com/crossplane/provider-aws/pkg/controller/glue/securityconfiguration"
	"github.com/crossplane/provider-aws/pkg/controller/iam/accesskey"
	"github.com/crossplane/provider-aws/pkg/controller/iam/group"
	"github.com/crossplane/provider-aws/pkg/controller/iam/grouppolicyattachment"
	"github.com/crossplane/provider-aws/pkg/controller/iam/groupusermembership"
	"github.com/crossplane/provider-aws/pkg/controller/iam/openidconnectprovider"
	"github.com/crossplane/provider-aws/pkg/controller/iam/policy"
	"github.com/crossplane/provider-aws/pkg/controller/iam/role"
	"github.com/crossplane/provider-aws/pkg/controller/iam/rolepolicyattachment"
	"github.com/crossplane/provider-aws/pkg/controller/iam/user"
	"github.com/crossplane/provider-aws/pkg/controller/iam/userpolicyattachment"
	iotpolicy "github.com/crossplane/provider-aws/pkg/controller/iot/policy"
	"github.com/crossplane/provider-aws/pkg/controller/iot/thing"
	kafkacluster "github.com/crossplane/provider-aws/pkg/controller/kafka/cluster"
	kafkaconfiguration "github.com/crossplane/provider-aws/pkg/controller/kafka/configuration"
	kinesisstream "github.com/crossplane/provider-aws/pkg/controller/kinesis/stream"
	"github.com/crossplane/provider-aws/pkg/controller/kms/alias"
	"github.com/crossplane/provider-aws/pkg/controller/kms/key"
	"github.com/crossplane/provider-aws/pkg/controller/lambda/function"
	mqbroker "github.com/crossplane/provider-aws/pkg/controller/mq/broker"
	mquser "github.com/crossplane/provider-aws/pkg/controller/mq/user"
	neptunecluster "github.com/crossplane/provider-aws/pkg/controller/neptune/dbcluster"
	notsubscription "github.com/crossplane/provider-aws/pkg/controller/notification/snssubscription"
	nottopic "github.com/crossplane/provider-aws/pkg/controller/notification/snstopic"
	resourceshare "github.com/crossplane/provider-aws/pkg/controller/ram/resourceshare"
	"github.com/crossplane/provider-aws/pkg/controller/rds/dbcluster"
	"github.com/crossplane/provider-aws/pkg/controller/rds/dbclusterparametergroup"
	"github.com/crossplane/provider-aws/pkg/controller/rds/dbinstance"
	"github.com/crossplane/provider-aws/pkg/controller/rds/dbparametergroup"
	"github.com/crossplane/provider-aws/pkg/controller/rds/globalcluster"
	"github.com/crossplane/provider-aws/pkg/controller/redshift"
	"github.com/crossplane/provider-aws/pkg/controller/route53/hostedzone"
	"github.com/crossplane/provider-aws/pkg/controller/route53/resourcerecordset"
	"github.com/crossplane/provider-aws/pkg/controller/route53resolver/resolverendpoint"
	"github.com/crossplane/provider-aws/pkg/controller/route53resolver/resolverrule"
	"github.com/crossplane/provider-aws/pkg/controller/route53resolver/resolverruleassociation"
	"github.com/crossplane/provider-aws/pkg/controller/s3"
	"github.com/crossplane/provider-aws/pkg/controller/s3/bucketpolicy"
	"github.com/crossplane/provider-aws/pkg/controller/secretsmanager/secret"
	"github.com/crossplane/provider-aws/pkg/controller/servicediscovery/httpnamespace"
	"github.com/crossplane/provider-aws/pkg/controller/servicediscovery/privatednsnamespace"
	"github.com/crossplane/provider-aws/pkg/controller/servicediscovery/publicdnsnamespace"
	"github.com/crossplane/provider-aws/pkg/controller/sfn/activity"
	"github.com/crossplane/provider-aws/pkg/controller/sfn/statemachine"
	"github.com/crossplane/provider-aws/pkg/controller/sns/subscription"
	"github.com/crossplane/provider-aws/pkg/controller/sns/topic"
	"github.com/crossplane/provider-aws/pkg/controller/sqs/queue"
	transferserver "github.com/crossplane/provider-aws/pkg/controller/transfer/server"
	transferuser "github.com/crossplane/provider-aws/pkg/controller/transfer/user"
)

// Setup creates all AWS controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	for _, setup := range []func(ctrl.Manager, logging.Logger, workqueue.RateLimiter, time.Duration) error{
		cache.SetupReplicationGroup,
		cachesubnetgroup.SetupCacheSubnetGroup,
		cacheparametergroup.SetupCacheParameterGroup,
		cluster.SetupCacheCluster,
		database.SetupRDSInstance,
		docdbinstance.SetupDBInstance,
		docdbcluster.SetupDBCluster,
		docdbclusterparametergroup.SetupDBClusterParameterGroup,
		docdbsubnetgroup.SetupDBSubnetGroup,
		eks.SetupCluster,
		eksaddon.SetupAddon,
		identityproviderconfig.SetupIdentityProviderConfig,
		elb.SetupELB,
		elbattachment.SetupELBAttachment,
		nodegroup.SetupNodeGroup,
		s3.SetupBucket,
		bucketpolicy.SetupBucketPolicy,
		accesskey.SetupAccessKey,
		user.SetupUser,
		group.SetupGroup,
		policy.SetupPolicy,
		role.SetupRole,
		groupusermembership.SetupGroupUserMembership,
		userpolicyattachment.SetupUserPolicyAttachment,
		grouppolicyattachment.SetupGroupPolicyAttachment,
		rolepolicyattachment.SetupRolePolicyAttachment,
		vpc.SetupVPC,
		subnet.SetupSubnet,
		securitygroup.SetupSecurityGroup,
		internetgateway.SetupInternetGateway,
		launchtemplate.SetupLaunchTemplate,
		launchtemplateversion.SetupLaunchTemplateVersion,
		natgateway.SetupNatGateway,
		routetable.SetupRouteTable,
		dbsubnetgroup.SetupDBSubnetGroup,
		certificateauthority.SetupCertificateAuthority,
		certificateauthoritypermission.SetupCertificateAuthorityPermission,
		acm.SetupCertificate,
		resourcerecordset.SetupResourceRecordSet,
		hostedzone.SetupHostedZone,
		secret.SetupSecret,
		topic.SetupSNSTopic,
		subscription.SetupSubscription,
		queue.SetupQueue,
		redshift.SetupCluster,
		address.SetupAddress,
		repository.SetupRepository,
		repositorypolicy.SetupRepositoryPolicy,
		api.SetupAPI,
		stage.SetupStage,
		route.SetupRoute,
		authorizer.SetupAuthorizer,
		integration.SetupIntegration,
		deployment.SetupDeployment,
		domainname.SetupDomainName,
		integrationresponse.SetupIntegrationResponse,
		model.SetupModel,
		apimapping.SetupAPIMapping,
		routeresponse.SetupRouteResponse,
		vpclink.SetupVPCLink,
		fargateprofile.SetupFargateProfile,
		activity.SetupActivity,
		statemachine.SetupStateMachine,
		table.SetupTable,
		backup.SetupBackup,
		globaltable.SetupGlobalTable,
		key.SetupKey,
		alias.SetupAlias,
		filesystem.SetupFileSystem,
		dbcluster.SetupDBCluster,
		dbclusterparametergroup.SetupDBClusterParameterGroup,
		dbinstance.SetupDBInstance,
		dbparametergroup.SetupDBParameterGroup,
		globalcluster.SetupGlobalCluster,
		vpccidrblock.SetupVPCCIDRBlock,
		privatednsnamespace.SetupPrivateDNSNamespace,
		publicdnsnamespace.SetupPublicDNSNamespace,
		httpnamespace.SetupHTTPNamespace,
		function.SetupFunction,
		openidconnectprovider.SetupOpenIDConnectProvider,
		distribution.SetupDistribution,
		cachepolicy.SetupCachePolicy,
		cloudfrontorginaccessidentity.SetupCloudFrontOriginAccessIdentity,
		resolverendpoint.SetupResolverEndpoint,
		resolverrule.SetupResolverRule,
		vpcpeeringconnection.SetupVPCPeeringConnection,
		vpcendpoint.SetupVPCEndpoint,
		kafkacluster.SetupCluster,
		efsmounttarget.SetupMountTarget,
		transferserver.SetupServer,
		transferuser.SetupUser,
		instance.SetupInstance,
		gluejob.SetupJob,
		gluesecurityconfiguration.SetupSecurityConfiguration,
		glueconnection.SetupConnection,
		glueDatabase.SetupDatabase,
		gluecrawler.SetupCrawler,
		glueclassifier.SetupClassifier,
		mqbroker.SetupBroker,
		mquser.SetupUser,
		cwloggroup.SetupLogGroup,
		volume.SetupVolume,
		transitgateway.SetupTransitGateway,
		transitgatewayvpcattachment.SetupTransitGatewayVPCAttachment,
		thing.SetupThing,
		iotpolicy.SetupPolicy,
		ec2route.SetupRoute,
		athenaworkgroup.SetupWorkGroup,
		resourceshare.SetupResourceShare,
		kafkaconfiguration.SetupConfiguration,
		listener.SetupListener,
		loadbalancer.SetupLoadBalancer,
		targetgroup.SetupTargetGroup,
		transitgatewayroute.SetupTransitGatewayRoute,
		transitgatewayroutetable.SetupTransitGatewayRouteTable,
		vpcendpointserviceconfiguration.SetupVPCEndpointServiceConfiguration,
		kinesisstream.SetupStream,
		resolverruleassociation.SetupResolverRuleAssociation,
		neptunecluster.SetupDBCluster,
		topic.SetupSNSTopic,
		subscription.SetupSubscription,
		nottopic.SetupSNSTopic,
		notsubscription.SetupSubscription,
	} {
		if err := setup(mgr, l, rl, poll); err != nil {
			return err
		}
	}

	return config.Setup(mgr, l, rl)
}
