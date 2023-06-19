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
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/crossplane-runtime/pkg/controller"

	"github.com/crossplane-contrib/provider-aws/pkg/controller/acm"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/acmpca/certificateauthority"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/acmpca/certificateauthoritypermission"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/apigateway/method"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/apigateway/resource"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/apigateway/restapi"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/apigatewayv2/api"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/apigatewayv2/apimapping"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/apigatewayv2/authorizer"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/apigatewayv2/deployment"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/apigatewayv2/domainname"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/apigatewayv2/integration"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/apigatewayv2/integrationresponse"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/apigatewayv2/model"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/apigatewayv2/route"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/apigatewayv2/routeresponse"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/apigatewayv2/stage"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/apigatewayv2/vpclink"
	athenaworkgroup "github.com/crossplane-contrib/provider-aws/pkg/controller/athena/workgroup"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/autoscaling/autoscalinggroup"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/batch/computeenvironment"
	batchjob "github.com/crossplane-contrib/provider-aws/pkg/controller/batch/job"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/batch/jobdefinition"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/batch/jobqueue"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/cache"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/cache/cachesubnetgroup"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/cache/cluster"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/cloudfront/cachepolicy"
	cloudfrontorginaccessidentity "github.com/crossplane-contrib/provider-aws/pkg/controller/cloudfront/cloudfrontoriginaccessidentity"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/cloudfront/distribution"
	cloudfrontresponseheaderspolicy "github.com/crossplane-contrib/provider-aws/pkg/controller/cloudfront/responseheaderspolicy"
	domain "github.com/crossplane-contrib/provider-aws/pkg/controller/cloudsearch/domain"
	cwloggroup "github.com/crossplane-contrib/provider-aws/pkg/controller/cloudwatchlogs/loggroup"
	cognitoidentitypool "github.com/crossplane-contrib/provider-aws/pkg/controller/cognitoidentity/identitypool"
	cognitogroup "github.com/crossplane-contrib/provider-aws/pkg/controller/cognitoidentityprovider/group"
	cognitogroupusermembership "github.com/crossplane-contrib/provider-aws/pkg/controller/cognitoidentityprovider/groupusermembership"
	cognitoidentityprovider "github.com/crossplane-contrib/provider-aws/pkg/controller/cognitoidentityprovider/identityprovider"
	cognitoresourceserver "github.com/crossplane-contrib/provider-aws/pkg/controller/cognitoidentityprovider/resourceserver"
	cognitouserpool "github.com/crossplane-contrib/provider-aws/pkg/controller/cognitoidentityprovider/userpool"
	cognitouserpoolclient "github.com/crossplane-contrib/provider-aws/pkg/controller/cognitoidentityprovider/userpoolclient"
	cognitouserpooldomain "github.com/crossplane-contrib/provider-aws/pkg/controller/cognitoidentityprovider/userpooldomain"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/config"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/database"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/database/dbsubnetgroup"
	daxcluster "github.com/crossplane-contrib/provider-aws/pkg/controller/dax/cluster"
	daxparametergroup "github.com/crossplane-contrib/provider-aws/pkg/controller/dax/parametergroup"
	daxsubnetgroup "github.com/crossplane-contrib/provider-aws/pkg/controller/dax/subnetgroup"
	docdbcluster "github.com/crossplane-contrib/provider-aws/pkg/controller/docdb/dbcluster"
	docdbclusterparametergroup "github.com/crossplane-contrib/provider-aws/pkg/controller/docdb/dbclusterparametergroup"
	docdbinstance "github.com/crossplane-contrib/provider-aws/pkg/controller/docdb/dbinstance"
	docdbsubnetgroup "github.com/crossplane-contrib/provider-aws/pkg/controller/docdb/dbsubnetgroup"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/dynamodb/backup"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/dynamodb/globaltable"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/dynamodb/table"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/address"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/flowlog"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/instance"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/internetgateway"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/launchtemplate"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/launchtemplateversion"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/natgateway"
	ec2route "github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/route"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/routetable"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/securitygroup"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/securitygrouprule"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/subnet"
	transitgateway "github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/transitgateway"
	transitgatewayroute "github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/transitgatewayroute"
	transitgatewayroutetable "github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/transitgatewayroutetable"
	transitgatewayvpcattachment "github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/transitgatewayvpcattachment"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/volume"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/vpc"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/vpccidrblock"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/vpcendpoint"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/vpcendpointserviceconfiguration"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/vpcpeeringconnection"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ecr/lifecyclepolicy"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ecr/repository"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ecr/repositorypolicy"
	ecscluster "github.com/crossplane-contrib/provider-aws/pkg/controller/ecs/cluster"
	ecsservice "github.com/crossplane-contrib/provider-aws/pkg/controller/ecs/service"
	ecstask "github.com/crossplane-contrib/provider-aws/pkg/controller/ecs/taskdefinition"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/efs/accesspoint"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/efs/filesystem"
	efsmounttarget "github.com/crossplane-contrib/provider-aws/pkg/controller/efs/mounttarget"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/eks"
	eksaddon "github.com/crossplane-contrib/provider-aws/pkg/controller/eks/addon"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/eks/fargateprofile"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/eks/identityproviderconfig"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/eks/nodegroup"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/elasticache/cacheparametergroup"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/elasticloadbalancing/elb"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/elasticloadbalancing/elbattachment"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/elbv2/listener"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/elbv2/loadbalancer"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/elbv2/target"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/elbv2/targetgroup"
	emrcontainersjobrun "github.com/crossplane-contrib/provider-aws/pkg/controller/emrcontainers/jobrun"
	emrcontainersvirtualcluster "github.com/crossplane-contrib/provider-aws/pkg/controller/emrcontainers/virtualcluster"
	glueclassifier "github.com/crossplane-contrib/provider-aws/pkg/controller/glue/classifier"
	glueconnection "github.com/crossplane-contrib/provider-aws/pkg/controller/glue/connection"
	gluecrawler "github.com/crossplane-contrib/provider-aws/pkg/controller/glue/crawler"
	glueDatabase "github.com/crossplane-contrib/provider-aws/pkg/controller/glue/database"
	gluejob "github.com/crossplane-contrib/provider-aws/pkg/controller/glue/job"
	gluesecurityconfiguration "github.com/crossplane-contrib/provider-aws/pkg/controller/glue/securityconfiguration"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/iam/accesskey"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/iam/group"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/iam/grouppolicyattachment"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/iam/groupusermembership"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/iam/instanceprofile"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/iam/openidconnectprovider"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/iam/policy"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/iam/role"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/iam/rolepolicyattachment"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/iam/user"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/iam/userpolicyattachment"
	iotpolicy "github.com/crossplane-contrib/provider-aws/pkg/controller/iot/policy"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/iot/thing"
	kafkacluster "github.com/crossplane-contrib/provider-aws/pkg/controller/kafka/cluster"
	kafkaconfiguration "github.com/crossplane-contrib/provider-aws/pkg/controller/kafka/configuration"
	kinesisstream "github.com/crossplane-contrib/provider-aws/pkg/controller/kinesis/stream"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/kms/alias"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/kms/key"
	lambdafunction "github.com/crossplane-contrib/provider-aws/pkg/controller/lambda/function"
	lambdaurlconfig "github.com/crossplane-contrib/provider-aws/pkg/controller/lambda/functionurlconfig"
	lambdapermission "github.com/crossplane-contrib/provider-aws/pkg/controller/lambda/permission"
	mqbroker "github.com/crossplane-contrib/provider-aws/pkg/controller/mq/broker"
	mquser "github.com/crossplane-contrib/provider-aws/pkg/controller/mq/user"
	mwaaenvironment "github.com/crossplane-contrib/provider-aws/pkg/controller/mwaa/environment"
	neptunecluster "github.com/crossplane-contrib/provider-aws/pkg/controller/neptune/dbcluster"
	opensearchdomain "github.com/crossplane-contrib/provider-aws/pkg/controller/opensearchservice/domain"
	prometheusservicealertmanagerdefinition "github.com/crossplane-contrib/provider-aws/pkg/controller/prometheusservice/alertmanagerdefinition"
	prometheusservicerulegroupnamespace "github.com/crossplane-contrib/provider-aws/pkg/controller/prometheusservice/rulegroupsnamespace"
	prometheusserviceworkspace "github.com/crossplane-contrib/provider-aws/pkg/controller/prometheusservice/workspace"
	resourceshare "github.com/crossplane-contrib/provider-aws/pkg/controller/ram/resourceshare"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/rds/dbcluster"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/rds/dbclusterparametergroup"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/rds/dbinstance"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/rds/dbinstanceroleassociation"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/rds/dbparametergroup"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/rds/globalcluster"
	optiongroup "github.com/crossplane-contrib/provider-aws/pkg/controller/rds/optiongroup"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/redshift"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/route53/hostedzone"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/route53/resourcerecordset"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/route53resolver/resolverendpoint"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/route53resolver/resolverrule"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/route53resolver/resolverruleassociation"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/s3"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/s3/bucketpolicy"
	s3control "github.com/crossplane-contrib/provider-aws/pkg/controller/s3control/accesspoint"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/secretsmanager/secret"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/servicediscovery/httpnamespace"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/servicediscovery/privatednsnamespace"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/servicediscovery/publicdnsnamespace"
	servicediscoveryservice "github.com/crossplane-contrib/provider-aws/pkg/controller/servicediscovery/service"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/sfn/activity"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/sfn/statemachine"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/sns/subscription"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/sns/topic"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/sqs/queue"
	transferserver "github.com/crossplane-contrib/provider-aws/pkg/controller/transfer/server"
	transferuser "github.com/crossplane-contrib/provider-aws/pkg/controller/transfer/user"
)

// Setup creates all AWS controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		cache.SetupReplicationGroup,
		cachesubnetgroup.SetupCacheSubnetGroup,
		cacheparametergroup.SetupCacheParameterGroup,
		cluster.SetupCacheCluster,
		database.SetupRDSInstance,
		daxcluster.SetupCluster,
		daxparametergroup.SetupParameterGroup,
		daxsubnetgroup.SetupSubnetGroup,
		domain.SetupDomain,
		docdbinstance.SetupDBInstance,
		docdbcluster.SetupDBCluster,
		docdbclusterparametergroup.SetupDBClusterParameterGroup,
		docdbsubnetgroup.SetupDBSubnetGroup,
		ecscluster.SetupCluster,
		ecsservice.SetupService,
		ecstask.SetupTaskDefinition,
		eks.SetupCluster,
		eksaddon.SetupAddon,
		identityproviderconfig.SetupIdentityProviderConfig,
		instanceprofile.SetupInstanceProfile,
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
		securitygrouprule.SetupSecurityGroupRule,
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
		lifecyclepolicy.SetupLifecyclePolicy,
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
		accesspoint.SetupAccessPoint,
		filesystem.SetupFileSystem,
		dbcluster.SetupDBCluster,
		dbclusterparametergroup.SetupDBClusterParameterGroup,
		dbinstance.SetupDBInstance,
		dbinstanceroleassociation.SetupDBInstanceRoleAssociation,
		dbparametergroup.SetupDBParameterGroup,
		globalcluster.SetupGlobalCluster,
		vpccidrblock.SetupVPCCIDRBlock,
		privatednsnamespace.SetupPrivateDNSNamespace,
		publicdnsnamespace.SetupPublicDNSNamespace,
		httpnamespace.SetupHTTPNamespace,
		lambdafunction.SetupFunction,
		lambdapermission.SetupPermission,
		lambdaurlconfig.SetupFunctionURL,
		openidconnectprovider.SetupOpenIDConnectProvider,
		distribution.SetupDistribution,
		cachepolicy.SetupCachePolicy,
		cloudfrontorginaccessidentity.SetupCloudFrontOriginAccessIdentity,
		cloudfrontresponseheaderspolicy.SetupResponseHeadersPolicy,
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
		mwaaenvironment.SetupEnvironment,
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
		target.SetupTarget,
		transitgatewayroute.SetupTransitGatewayRoute,
		transitgatewayroutetable.SetupTransitGatewayRouteTable,
		vpcendpointserviceconfiguration.SetupVPCEndpointServiceConfiguration,
		kinesisstream.SetupStream,
		resolverruleassociation.SetupResolverRuleAssociation,
		cognitouserpool.SetupUserPool,
		cognitouserpooldomain.SetupUserPoolDomain,
		cognitogroup.SetupGroup,
		cognitouserpoolclient.SetupUserPoolClient,
		cognitoidentityprovider.SetupIdentityProvider,
		cognitoresourceserver.SetupResourceServer,
		cognitogroupusermembership.SetupGroupUserMembership,
		neptunecluster.SetupDBCluster,
		topic.SetupSNSTopic,
		subscription.SetupSubscription,
		prometheusserviceworkspace.SetupWorkspace,
		prometheusservicerulegroupnamespace.SetupRuleGroupsNamespace,
		prometheusservicealertmanagerdefinition.SetupAlertManagerDefinition,
		resource.SetupResource,
		restapi.SetupRestAPI,
		method.SetupMethod,
		cognitoidentitypool.SetupIdentityPool,
		flowlog.SetupFlowLog,
		opensearchdomain.SetupDomain,
		computeenvironment.SetupComputeEnvironment,
		jobqueue.SetupJobQueue,
		jobdefinition.SetupJobDefinition,
		batchjob.SetupJob,
		emrcontainersjobrun.SetupJobRun,
		emrcontainersvirtualcluster.SetupVirtualCluster,
		optiongroup.SetupOptionGroup,
		autoscalinggroup.SetupAutoScalingGroup,
		s3control.SetupAccessPoint,
		servicediscoveryservice.SetupService,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}

	return config.Setup(mgr, o)
}
