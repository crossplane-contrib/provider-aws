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
	"github.com/crossplane-contrib/provider-aws/pkg/controller/redshift"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/route53/hostedzone"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/route53/resourcerecordset"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/route53resolver/resolverendpoint"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/route53resolver/resolverrule"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/route53resolver/resolverruleassociation"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/s3"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/s3/bucketpolicy"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/secretsmanager/secret"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/servicediscovery/httpnamespace"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/servicediscovery/privatednsnamespace"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/servicediscovery/publicdnsnamespace"
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
func Setup(mgr ctrl.Manager, o controller.Options, controllerPollInterval map[string]time.Duration) error {
	for name, setup := range map[string]func(ctrl.Manager, controller.Options) error{
		cache.ControllerName:                                   cache.SetupReplicationGroup,
		cachesubnetgroup.ControllerName:                        cachesubnetgroup.SetupCacheSubnetGroup,
		cacheparametergroup.ControllerName:                     cacheparametergroup.SetupCacheParameterGroup,
		cluster.ControllerName:                                 cluster.SetupCacheCluster,
		database.ControllerName:                                database.SetupRDSInstance,
		daxcluster.ControllerName:                              daxcluster.SetupCluster,
		daxparametergroup.ControllerName:                       daxparametergroup.SetupParameterGroup,
		daxsubnetgroup.ControllerName:                          daxsubnetgroup.SetupSubnetGroup,
		domain.ControllerName:                                  domain.SetupDomain,
		docdbinstance.ControllerName:                           docdbinstance.SetupDBInstance,
		docdbcluster.ControllerName:                            docdbcluster.SetupDBCluster,
		docdbclusterparametergroup.ControllerName:              docdbclusterparametergroup.SetupDBClusterParameterGroup,
		docdbsubnetgroup.ControllerName:                        docdbsubnetgroup.SetupDBSubnetGroup,
		ecscluster.ControllerName:                              ecscluster.SetupCluster,
		ecsservice.ControllerName:                              ecsservice.SetupService,
		ecstask.ControllerName:                                 ecstask.SetupTaskDefinition,
		eks.ControllerName:                                     eks.SetupCluster,
		eksaddon.ControllerName:                                eksaddon.SetupAddon,
		identityproviderconfig.ControllerName:                  identityproviderconfig.SetupIdentityProviderConfig,
		instanceprofile.ControllerName:                         instanceprofile.SetupInstanceProfile,
		elb.ControllerName:                                     elb.SetupELB,
		elbattachment.ControllerName:                           elbattachment.SetupELBAttachment,
		nodegroup.ControllerName:                               nodegroup.SetupNodeGroup,
		s3.ControllerName:                                      s3.SetupBucket,
		bucketpolicy.ControllerName:                            bucketpolicy.SetupBucketPolicy,
		accesskey.ControllerName:                               accesskey.SetupAccessKey,
		user.ControllerName:                                    user.SetupUser,
		group.ControllerName:                                   group.SetupGroup,
		policy.ControllerName:                                  policy.SetupPolicy,
		role.ControllerName:                                    role.SetupRole,
		groupusermembership.ControllerName:                     groupusermembership.SetupGroupUserMembership,
		userpolicyattachment.ControllerName:                    userpolicyattachment.SetupUserPolicyAttachment,
		grouppolicyattachment.ControllerName:                   grouppolicyattachment.SetupGroupPolicyAttachment,
		rolepolicyattachment.ControllerName:                    rolepolicyattachment.SetupRolePolicyAttachment,
		vpc.ControllerName:                                     vpc.SetupVPC,
		subnet.ControllerName:                                  subnet.SetupSubnet,
		securitygroup.ControllerName:                           securitygroup.SetupSecurityGroup,
		securitygrouprule.ControllerName:                       securitygrouprule.SetupSecurityGroupRule,
		internetgateway.ControllerName:                         internetgateway.SetupInternetGateway,
		launchtemplate.ControllerName:                          launchtemplate.SetupLaunchTemplate,
		launchtemplateversion.ControllerName:                   launchtemplateversion.SetupLaunchTemplateVersion,
		natgateway.ControllerName:                              natgateway.SetupNatGateway,
		routetable.ControllerName:                              routetable.SetupRouteTable,
		dbsubnetgroup.ControllerName:                           dbsubnetgroup.SetupDBSubnetGroup,
		certificateauthority.ControllerName:                    certificateauthority.SetupCertificateAuthority,
		certificateauthoritypermission.ControllerName:          certificateauthoritypermission.SetupCertificateAuthorityPermission,
		acm.ControllerName:                                     acm.SetupCertificate,
		resourcerecordset.ControllerName:                       resourcerecordset.SetupResourceRecordSet,
		hostedzone.ControllerName:                              hostedzone.SetupHostedZone,
		secret.ControllerName:                                  secret.SetupSecret,
		topic.ControllerName:                                   topic.SetupSNSTopic,
		subscription.ControllerName:                            subscription.SetupSubscription,
		queue.ControllerName:                                   queue.SetupQueue,
		redshift.ControllerName:                                redshift.SetupCluster,
		address.ControllerName:                                 address.SetupAddress,
		repository.ControllerName:                              repository.SetupRepository,
		repositorypolicy.ControllerName:                        repositorypolicy.SetupRepositoryPolicy,
		lifecyclepolicy.ControllerName:                         lifecyclepolicy.SetupLifecyclePolicy,
		api.ControllerName:                                     api.SetupAPI,
		stage.ControllerName:                                   stage.SetupStage,
		route.ControllerName:                                   route.SetupRoute,
		authorizer.ControllerName:                              authorizer.SetupAuthorizer,
		integration.ControllerName:                             integration.SetupIntegration,
		deployment.ControllerName:                              deployment.SetupDeployment,
		domainname.ControllerName:                              domainname.SetupDomainName,
		integrationresponse.ControllerName:                     integrationresponse.SetupIntegrationResponse,
		model.ControllerName:                                   model.SetupModel,
		apimapping.ControllerName:                              apimapping.SetupAPIMapping,
		routeresponse.ControllerName:                           routeresponse.SetupRouteResponse,
		vpclink.ControllerName:                                 vpclink.SetupVPCLink,
		fargateprofile.ControllerName:                          fargateprofile.SetupFargateProfile,
		activity.ControllerName:                                activity.SetupActivity,
		statemachine.ControllerName:                            statemachine.SetupStateMachine,
		table.ControllerName:                                   table.SetupTable,
		backup.ControllerName:                                  backup.SetupBackup,
		globaltable.ControllerName:                             globaltable.SetupGlobalTable,
		key.ControllerName:                                     key.SetupKey,
		alias.ControllerName:                                   alias.SetupAlias,
		accesspoint.ControllerName:                             accesspoint.SetupAccessPoint,
		filesystem.ControllerName:                              filesystem.SetupFileSystem,
		dbcluster.ControllerName:                               dbcluster.SetupDBCluster,
		dbclusterparametergroup.ControllerName:                 dbclusterparametergroup.SetupDBClusterParameterGroup,
		dbinstance.ControllerName:                              dbinstance.SetupDBInstance,
		dbinstanceroleassociation.ControllerName:               dbinstanceroleassociation.SetupDBInstanceRoleAssociation,
		dbparametergroup.ControllerName:                        dbparametergroup.SetupDBParameterGroup,
		globalcluster.ControllerName:                           globalcluster.SetupGlobalCluster,
		vpccidrblock.ControllerName:                            vpccidrblock.SetupVPCCIDRBlock,
		privatednsnamespace.ControllerName:                     privatednsnamespace.SetupPrivateDNSNamespace,
		publicdnsnamespace.ControllerName:                      publicdnsnamespace.SetupPublicDNSNamespace,
		httpnamespace.ControllerName:                           httpnamespace.SetupHTTPNamespace,
		lambdafunction.ControllerName:                          lambdafunction.SetupFunction,
		lambdapermission.ControllerName:                        lambdapermission.SetupPermission,
		lambdaurlconfig.ControllerName:                         lambdaurlconfig.SetupFunctionURL,
		openidconnectprovider.ControllerName:                   openidconnectprovider.SetupOpenIDConnectProvider,
		distribution.ControllerName:                            distribution.SetupDistribution,
		cachepolicy.ControllerName:                             cachepolicy.SetupCachePolicy,
		cloudfrontorginaccessidentity.ControllerName:           cloudfrontorginaccessidentity.SetupCloudFrontOriginAccessIdentity,
		cloudfrontresponseheaderspolicy.ControllerName:         cloudfrontresponseheaderspolicy.SetupResponseHeadersPolicy,
		resolverendpoint.ControllerName:                        resolverendpoint.SetupResolverEndpoint,
		resolverrule.ControllerName:                            resolverrule.SetupResolverRule,
		vpcpeeringconnection.ControllerName:                    vpcpeeringconnection.SetupVPCPeeringConnection,
		vpcendpoint.ControllerName:                             vpcendpoint.SetupVPCEndpoint,
		kafkacluster.ControllerName:                            kafkacluster.SetupCluster,
		efsmounttarget.ControllerName:                          efsmounttarget.SetupMountTarget,
		transferserver.ControllerName:                          transferserver.SetupServer,
		transferuser.ControllerName:                            transferuser.SetupUser,
		instance.ControllerName:                                instance.SetupInstance,
		gluejob.ControllerName:                                 gluejob.SetupJob,
		gluesecurityconfiguration.ControllerName:               gluesecurityconfiguration.SetupSecurityConfiguration,
		glueconnection.ControllerName:                          glueconnection.SetupConnection,
		glueDatabase.ControllerName:                            glueDatabase.SetupDatabase,
		gluecrawler.ControllerName:                             gluecrawler.SetupCrawler,
		glueclassifier.ControllerName:                          glueclassifier.SetupClassifier,
		mqbroker.ControllerName:                                mqbroker.SetupBroker,
		mquser.ControllerName:                                  mquser.SetupUser,
		mwaaenvironment.ControllerName:                         mwaaenvironment.SetupEnvironment,
		cwloggroup.ControllerName:                              cwloggroup.SetupLogGroup,
		volume.ControllerName:                                  volume.SetupVolume,
		transitgateway.ControllerName:                          transitgateway.SetupTransitGateway,
		transitgatewayvpcattachment.ControllerName:             transitgatewayvpcattachment.SetupTransitGatewayVPCAttachment,
		thing.ControllerName:                                   thing.SetupThing,
		iotpolicy.ControllerName:                               iotpolicy.SetupPolicy,
		ec2route.ControllerName:                                ec2route.SetupRoute,
		athenaworkgroup.ControllerName:                         athenaworkgroup.SetupWorkGroup,
		resourceshare.ControllerName:                           resourceshare.SetupResourceShare,
		kafkaconfiguration.ControllerName:                      kafkaconfiguration.SetupConfiguration,
		listener.ControllerName:                                listener.SetupListener,
		loadbalancer.ControllerName:                            loadbalancer.SetupLoadBalancer,
		targetgroup.ControllerName:                             targetgroup.SetupTargetGroup,
		target.ControllerName:                                  target.SetupTarget,
		transitgatewayroute.ControllerName:                     transitgatewayroute.SetupTransitGatewayRoute,
		transitgatewayroutetable.ControllerName:                transitgatewayroutetable.SetupTransitGatewayRouteTable,
		vpcendpointserviceconfiguration.ControllerName:         vpcendpointserviceconfiguration.SetupVPCEndpointServiceConfiguration,
		kinesisstream.ControllerName:                           kinesisstream.SetupStream,
		resolverruleassociation.ControllerName:                 resolverruleassociation.SetupResolverRuleAssociation,
		cognitouserpool.ControllerName:                         cognitouserpool.SetupUserPool,
		cognitouserpooldomain.ControllerName:                   cognitouserpooldomain.SetupUserPoolDomain,
		cognitogroup.ControllerName:                            cognitogroup.SetupGroup,
		cognitouserpoolclient.ControllerName:                   cognitouserpoolclient.SetupUserPoolClient,
		cognitoidentityprovider.ControllerName:                 cognitoidentityprovider.SetupIdentityProvider,
		cognitoresourceserver.ControllerName:                   cognitoresourceserver.SetupResourceServer,
		cognitogroupusermembership.ControllerName:              cognitogroupusermembership.SetupGroupUserMembership,
		neptunecluster.ControllerName:                          neptunecluster.SetupDBCluster,
		prometheusserviceworkspace.ControllerName:              prometheusserviceworkspace.SetupWorkspace,
		prometheusservicerulegroupnamespace.ControllerName:     prometheusservicerulegroupnamespace.SetupRuleGroupsNamespace,
		prometheusservicealertmanagerdefinition.ControllerName: prometheusservicealertmanagerdefinition.SetupAlertManagerDefinition,
		resource.ControllerName:                                resource.SetupResource,
		restapi.ControllerName:                                 restapi.SetupRestAPI,
		method.ControllerName:                                  method.SetupMethod,
		cognitoidentitypool.ControllerName:                     cognitoidentitypool.SetupIdentityPool,
		flowlog.ControllerName:                                 flowlog.SetupFlowLog,
		opensearchdomain.ControllerName:                        opensearchdomain.SetupDomain,
		computeenvironment.ControllerName:                      computeenvironment.SetupComputeEnvironment,
		jobqueue.ControllerName:                                jobqueue.SetupJobQueue,
		jobdefinition.ControllerName:                           jobdefinition.SetupJobDefinition,
		batchjob.ControllerName:                                batchjob.SetupJob,
		emrcontainersjobrun.ControllerName:                     emrcontainersjobrun.SetupJobRun,
		emrcontainersvirtualcluster.ControllerName:             emrcontainersvirtualcluster.SetupVirtualCluster,
	} {
		if poll, ok := controllerPollInterval[name]; ok {
			o.Logger.Info("Setting controller poll interval", "controller", name, "interval", poll)
			o.PollInterval = poll
		}

		if err := setup(mgr, o); err != nil {
			return err
		}
	}

	return config.Setup(mgr, o)
}
