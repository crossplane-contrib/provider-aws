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
	"regexp"
	"strings"

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

type ReconileDefinition struct {
	Type  func() string
	Setup func(ctrl.Manager, controller.Options) error
}

type FilterArgument struct {
	IncludeCrds []string `json:"includeCrds"`
	ExcludeCrds []string `json:"excludeCrds"`
}

// Setup creates all AWS controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, filterCRDs FilterArgument, o controller.Options) error {

	// List all controllers with ther managesKind and setup function
	allResources := []ReconileDefinition{
		{
			Type:  cache.ManagesKind,
			Setup: cache.SetupReplicationGroup,
		},
		{
			Type:  cachesubnetgroup.ManagesKind,
			Setup: cachesubnetgroup.SetupCacheSubnetGroup,
		},
		{
			Type:  cacheparametergroup.ManagesKind,
			Setup: cacheparametergroup.SetupCacheParameterGroup,
		},
		{
			Type:  cluster.ManagesKind,
			Setup: cluster.SetupCacheCluster,
		},
		{
			Type:  database.ManagesKind,
			Setup: database.SetupRDSInstance,
		},
		{
			Type:  daxcluster.ManagesKind,
			Setup: daxcluster.SetupCluster,
		},
		{
			Type:  daxparametergroup.ManagesKind,
			Setup: daxparametergroup.SetupParameterGroup,
		},
		{
			Type:  daxsubnetgroup.ManagesKind,
			Setup: daxsubnetgroup.SetupSubnetGroup,
		},
		{
			Type:  domain.ManagesKind,
			Setup: domain.SetupDomain,
		},
		{
			Type:  docdbinstance.ManagesKind,
			Setup: docdbinstance.SetupDBInstance,
		},
		{
			Type:  docdbcluster.ManagesKind,
			Setup: docdbcluster.SetupDBCluster,
		},
		{
			Type:  docdbclusterparametergroup.ManagesKind,
			Setup: docdbclusterparametergroup.SetupDBClusterParameterGroup,
		},
		{
			Type:  docdbsubnetgroup.ManagesKind,
			Setup: docdbsubnetgroup.SetupDBSubnetGroup,
		},
		{
			Type:  ecscluster.ManagesKind,
			Setup: ecscluster.SetupCluster,
		},
		{
			Type:  ecsservice.ManagesKind,
			Setup: ecsservice.SetupService,
		},
		{
			Type:  ecstask.ManagesKind,
			Setup: ecstask.SetupTaskDefinition,
		},
		{
			Type:  eks.ManagesKind,
			Setup: eks.SetupCluster,
		},
		{
			Type:  eksaddon.ManagesKind,
			Setup: eksaddon.SetupAddon,
		},
		{
			Type:  identityproviderconfig.ManagesKind,
			Setup: identityproviderconfig.SetupIdentityProviderConfig,
		},
		{
			Type:  instanceprofile.ManagesKind,
			Setup: instanceprofile.SetupInstanceProfile,
		},
		{
			Type:  elb.ManagesKind,
			Setup: elb.SetupELB,
		},
		{
			Type:  elbattachment.ManagesKind,
			Setup: elbattachment.SetupELBAttachment,
		},
		{
			Type:  nodegroup.ManagesKind,
			Setup: nodegroup.SetupNodeGroup,
		},
		{
			Type:  s3.ManagesKind,
			Setup: s3.SetupBucket,
		},
		{
			Type:  bucketpolicy.ManagesKind,
			Setup: bucketpolicy.SetupBucketPolicy,
		},
		{
			Type:  accesskey.ManagesKind,
			Setup: accesskey.SetupAccessKey,
		},
		{
			Type:  user.ManagesKind,
			Setup: user.SetupUser,
		},
		{
			Type:  group.ManagesKind,
			Setup: group.SetupGroup,
		},
		{
			Type:  policy.ManagesKind,
			Setup: policy.SetupPolicy,
		},
		{
			Type:  role.ManagesKind,
			Setup: role.SetupRole,
		},
		{
			Type:  groupusermembership.ManagesKind,
			Setup: groupusermembership.SetupGroupUserMembership,
		},
		{
			Type:  userpolicyattachment.ManagesKind,
			Setup: userpolicyattachment.SetupUserPolicyAttachment,
		},
		{
			Type:  grouppolicyattachment.ManagesKind,
			Setup: grouppolicyattachment.SetupGroupPolicyAttachment,
		},
		{
			Type:  rolepolicyattachment.ManagesKind,
			Setup: rolepolicyattachment.SetupRolePolicyAttachment,
		},
		{
			Type:  vpc.ManagesKind,
			Setup: vpc.SetupVPC,
		},
		{
			Type:  subnet.ManagesKind,
			Setup: subnet.SetupSubnet,
		},
		{
			Type:  securitygroup.ManagesKind,
			Setup: securitygroup.SetupSecurityGroup,
		},
		{
			Type:  securitygrouprule.ManagesKind,
			Setup: securitygrouprule.SetupSecurityGroupRule,
		},
		{
			Type:  internetgateway.ManagesKind,
			Setup: internetgateway.SetupInternetGateway,
		},
		{
			Type:  launchtemplate.ManagesKind,
			Setup: launchtemplate.SetupLaunchTemplate,
		},
		{
			Type:  launchtemplateversion.ManagesKind,
			Setup: launchtemplateversion.SetupLaunchTemplateVersion,
		},
		{
			Type:  natgateway.ManagesKind,
			Setup: natgateway.SetupNatGateway,
		},
		{
			Type:  routetable.ManagesKind,
			Setup: routetable.SetupRouteTable,
		},
		{
			Type:  dbsubnetgroup.ManagesKind,
			Setup: dbsubnetgroup.SetupDBSubnetGroup,
		},
		{
			Type:  certificateauthority.ManagesKind,
			Setup: certificateauthority.SetupCertificateAuthority,
		},
		{
			Type:  certificateauthoritypermission.ManagesKind,
			Setup: certificateauthoritypermission.SetupCertificateAuthorityPermission,
		},
		{
			Type:  acm.ManagesKind,
			Setup: acm.SetupCertificate,
		},
		{
			Type:  resourcerecordset.ManagesKind,
			Setup: resourcerecordset.SetupResourceRecordSet,
		},
		{
			Type:  hostedzone.ManagesKind,
			Setup: hostedzone.SetupHostedZone,
		},
		{
			Type:  secret.ManagesKind,
			Setup: secret.SetupSecret,
		},
		{
			Type:  topic.ManagesKind,
			Setup: topic.SetupSNSTopic,
		},
		{
			Type:  subscription.ManagesKind,
			Setup: subscription.SetupSubscription,
		},
		{
			Type:  queue.ManagesKind,
			Setup: queue.SetupQueue,
		},
		{
			Type:  redshift.ManagesKind,
			Setup: redshift.SetupCluster,
		},
		{
			Type:  address.ManagesKind,
			Setup: address.SetupAddress,
		},
		{
			Type:  repository.ManagesKind,
			Setup: repository.SetupRepository,
		},
		{
			Type:  repositorypolicy.ManagesKind,
			Setup: repositorypolicy.SetupRepositoryPolicy,
		},
		{
			Type:  lifecyclepolicy.ManagesKind,
			Setup: lifecyclepolicy.SetupLifecyclePolicy,
		},
		{
			Type:  api.ManagesKind,
			Setup: api.SetupAPI,
		},
		{
			Type:  stage.ManagesKind,
			Setup: stage.SetupStage,
		},
		{
			Type:  route.ManagesKind,
			Setup: route.SetupRoute,
		},
		{
			Type:  authorizer.ManagesKind,
			Setup: authorizer.SetupAuthorizer,
		},
		{
			Type:  integration.ManagesKind,
			Setup: integration.SetupIntegration,
		},
		{
			Type:  deployment.ManagesKind,
			Setup: deployment.SetupDeployment,
		},
		{
			Type:  domainname.ManagesKind,
			Setup: domainname.SetupDomainName,
		},
		{
			Type:  integrationresponse.ManagesKind,
			Setup: integrationresponse.SetupIntegrationResponse,
		},
		{
			Type:  model.ManagesKind,
			Setup: model.SetupModel,
		},
		{
			Type:  apimapping.ManagesKind,
			Setup: apimapping.SetupAPIMapping,
		},
		{
			Type:  routeresponse.ManagesKind,
			Setup: routeresponse.SetupRouteResponse,
		},
		{
			Type:  vpclink.ManagesKind,
			Setup: vpclink.SetupVPCLink,
		},
		{
			Type:  fargateprofile.ManagesKind,
			Setup: fargateprofile.SetupFargateProfile,
		},
		{
			Type:  activity.ManagesKind,
			Setup: activity.SetupActivity,
		},
		{
			Type:  statemachine.ManagesKind,
			Setup: statemachine.SetupStateMachine,
		},
		{
			Type:  table.ManagesKind,
			Setup: table.SetupTable,
		},
		{
			Type:  backup.ManagesKind,
			Setup: backup.SetupBackup,
		},
		{
			Type:  globaltable.ManagesKind,
			Setup: globaltable.SetupGlobalTable,
		},
		{
			Type:  key.ManagesKind,
			Setup: key.SetupKey,
		},
		{
			Type:  alias.ManagesKind,
			Setup: alias.SetupAlias,
		},
		{
			Type:  accesspoint.ManagesKind,
			Setup: accesspoint.SetupAccessPoint,
		},
		{
			Type:  filesystem.ManagesKind,
			Setup: filesystem.SetupFileSystem,
		},
		{
			Type:  dbcluster.ManagesKind,
			Setup: dbcluster.SetupDBCluster,
		},
		{
			Type:  dbclusterparametergroup.ManagesKind,
			Setup: dbclusterparametergroup.SetupDBClusterParameterGroup,
		},
		{
			Type:  dbinstance.ManagesKind,
			Setup: dbinstance.SetupDBInstance,
		},
		{
			Type:  dbinstanceroleassociation.ManagesKind,
			Setup: dbinstanceroleassociation.SetupDBInstanceRoleAssociation,
		},
		{
			Type:  dbparametergroup.ManagesKind,
			Setup: dbparametergroup.SetupDBParameterGroup,
		},
		{
			Type:  globalcluster.ManagesKind,
			Setup: globalcluster.SetupGlobalCluster,
		},
		{
			Type:  vpccidrblock.ManagesKind,
			Setup: vpccidrblock.SetupVPCCIDRBlock,
		},
		{
			Type:  privatednsnamespace.ManagesKind,
			Setup: privatednsnamespace.SetupPrivateDNSNamespace,
		},
		{
			Type:  publicdnsnamespace.ManagesKind,
			Setup: publicdnsnamespace.SetupPublicDNSNamespace,
		},
		{
			Type:  httpnamespace.ManagesKind,
			Setup: httpnamespace.SetupHTTPNamespace,
		},
		{
			Type:  lambdafunction.ManagesKind,
			Setup: lambdafunction.SetupFunction,
		},
		{
			Type:  lambdapermission.ManagesKind,
			Setup: lambdapermission.SetupPermission,
		},
		{
			Type:  lambdaurlconfig.ManagesKind,
			Setup: lambdaurlconfig.SetupFunctionURL,
		},
		{
			Type:  openidconnectprovider.ManagesKind,
			Setup: openidconnectprovider.SetupOpenIDConnectProvider,
		},
		{
			Type:  distribution.ManagesKind,
			Setup: distribution.SetupDistribution,
		},
		{
			Type:  cachepolicy.ManagesKind,
			Setup: cachepolicy.SetupCachePolicy,
		},
		{
			Type:  cloudfrontorginaccessidentity.ManagesKind,
			Setup: cloudfrontorginaccessidentity.SetupCloudFrontOriginAccessIdentity,
		},
		{
			Type:  cloudfrontresponseheaderspolicy.ManagesKind,
			Setup: cloudfrontresponseheaderspolicy.SetupResponseHeadersPolicy,
		},
		{
			Type:  resolverendpoint.ManagesKind,
			Setup: resolverendpoint.SetupResolverEndpoint,
		},
		{
			Type:  resolverrule.ManagesKind,
			Setup: resolverrule.SetupResolverRule,
		},
		{
			Type:  vpcpeeringconnection.ManagesKind,
			Setup: vpcpeeringconnection.SetupVPCPeeringConnection,
		},
		{
			Type:  vpcendpoint.ManagesKind,
			Setup: vpcendpoint.SetupVPCEndpoint,
		},
		{
			Type:  kafkacluster.ManagesKind,
			Setup: kafkacluster.SetupCluster,
		},
		{
			Type:  efsmounttarget.ManagesKind,
			Setup: efsmounttarget.SetupMountTarget,
		},
		{
			Type:  transferserver.ManagesKind,
			Setup: transferserver.SetupServer,
		},
		{
			Type:  transferuser.ManagesKind,
			Setup: transferuser.SetupUser,
		},
		{
			Type:  instance.ManagesKind,
			Setup: instance.SetupInstance,
		},
		{
			Type:  gluejob.ManagesKind,
			Setup: gluejob.SetupJob,
		},
		{
			Type:  gluesecurityconfiguration.ManagesKind,
			Setup: gluesecurityconfiguration.SetupSecurityConfiguration,
		},
		{
			Type:  glueconnection.ManagesKind,
			Setup: glueconnection.SetupConnection,
		},
		{
			Type:  glueDatabase.ManagesKind,
			Setup: glueDatabase.SetupDatabase,
		},
		{
			Type:  gluecrawler.ManagesKind,
			Setup: gluecrawler.SetupCrawler,
		},
		{
			Type:  glueclassifier.ManagesKind,
			Setup: glueclassifier.SetupClassifier,
		},
		{
			Type:  mqbroker.ManagesKind,
			Setup: mqbroker.SetupBroker,
		},
		{
			Type:  mquser.ManagesKind,
			Setup: mquser.SetupUser,
		},
		{
			Type:  mwaaenvironment.ManagesKind,
			Setup: mwaaenvironment.SetupEnvironment,
		},
		{
			Type:  cwloggroup.ManagesKind,
			Setup: cwloggroup.SetupLogGroup,
		},
		{
			Type:  volume.ManagesKind,
			Setup: volume.SetupVolume,
		},
		{
			Type:  transitgateway.ManagesKind,
			Setup: transitgateway.SetupTransitGateway,
		},
		{
			Type:  transitgatewayvpcattachment.ManagesKind,
			Setup: transitgatewayvpcattachment.SetupTransitGatewayVPCAttachment,
		},
		{
			Type:  thing.ManagesKind,
			Setup: thing.SetupThing,
		},
		{
			Type:  iotpolicy.ManagesKind,
			Setup: iotpolicy.SetupPolicy,
		},
		{
			Type:  ec2route.ManagesKind,
			Setup: ec2route.SetupRoute,
		},
		{
			Type:  athenaworkgroup.ManagesKind,
			Setup: athenaworkgroup.SetupWorkGroup,
		},
		{
			Type:  resourceshare.ManagesKind,
			Setup: resourceshare.SetupResourceShare,
		},
		{
			Type:  kafkaconfiguration.ManagesKind,
			Setup: kafkaconfiguration.SetupConfiguration,
		},
		{
			Type:  listener.ManagesKind,
			Setup: listener.SetupListener,
		},
		{
			Type:  loadbalancer.ManagesKind,
			Setup: loadbalancer.SetupLoadBalancer,
		},
		{
			Type:  targetgroup.ManagesKind,
			Setup: targetgroup.SetupTargetGroup,
		},
		{
			Type:  target.ManagesKind,
			Setup: target.SetupTarget,
		},
		{
			Type:  transitgatewayroute.ManagesKind,
			Setup: transitgatewayroute.SetupTransitGatewayRoute,
		},
		{
			Type:  transitgatewayroutetable.ManagesKind,
			Setup: transitgatewayroutetable.SetupTransitGatewayRouteTable,
		},
		{
			Type:  vpcendpointserviceconfiguration.ManagesKind,
			Setup: vpcendpointserviceconfiguration.SetupVPCEndpointServiceConfiguration,
		},
		{
			Type:  kinesisstream.ManagesKind,
			Setup: kinesisstream.SetupStream,
		},
		{
			Type:  resolverruleassociation.ManagesKind,
			Setup: resolverruleassociation.SetupResolverRuleAssociation,
		},
		{
			Type:  cognitouserpool.ManagesKind,
			Setup: cognitouserpool.SetupUserPool,
		},
		{
			Type:  cognitouserpooldomain.ManagesKind,
			Setup: cognitouserpooldomain.SetupUserPoolDomain,
		},
		{
			Type:  cognitogroup.ManagesKind,
			Setup: cognitogroup.SetupGroup,
		},
		{
			Type:  cognitouserpoolclient.ManagesKind,
			Setup: cognitouserpoolclient.SetupUserPoolClient,
		},
		{
			Type:  cognitoidentityprovider.ManagesKind,
			Setup: cognitoidentityprovider.SetupIdentityProvider,
		},
		{
			Type:  cognitoresourceserver.ManagesKind,
			Setup: cognitoresourceserver.SetupResourceServer,
		},
		{
			Type:  cognitogroupusermembership.ManagesKind,
			Setup: cognitogroupusermembership.SetupGroupUserMembership,
		},
		{
			Type:  neptunecluster.ManagesKind,
			Setup: neptunecluster.SetupDBCluster,
		},
		{
			Type:  topic.ManagesKind,
			Setup: topic.SetupSNSTopic,
		},
		{
			Type:  subscription.ManagesKind,
			Setup: subscription.SetupSubscription,
		},
		{
			Type:  prometheusserviceworkspace.ManagesKind,
			Setup: prometheusserviceworkspace.SetupWorkspace,
		},
		{
			Type:  prometheusservicerulegroupnamespace.ManagesKind,
			Setup: prometheusservicerulegroupnamespace.SetupRuleGroupsNamespace,
		},
		{
			Type:  prometheusservicealertmanagerdefinition.ManagesKind,
			Setup: prometheusservicealertmanagerdefinition.SetupAlertManagerDefinition,
		},
		{
			Type:  resource.ManagesKind,
			Setup: resource.SetupResource,
		},
		{
			Type:  restapi.ManagesKind,
			Setup: restapi.SetupRestAPI,
		},
		{
			Type:  method.ManagesKind,
			Setup: method.SetupMethod,
		},
		{
			Type:  cognitoidentitypool.ManagesKind,
			Setup: cognitoidentitypool.SetupIdentityPool,
		},
		{
			Type:  flowlog.ManagesKind,
			Setup: flowlog.SetupFlowLog,
		},
		{
			Type:  opensearchdomain.ManagesKind,
			Setup: opensearchdomain.SetupDomain,
		},
		{
			Type:  computeenvironment.ManagesKind,
			Setup: computeenvironment.SetupComputeEnvironment,
		},
		{
			Type:  jobqueue.ManagesKind,
			Setup: jobqueue.SetupJobQueue,
		},
		{
			Type:  jobdefinition.ManagesKind,
			Setup: jobdefinition.SetupJobDefinition,
		},
		{
			Type:  batchjob.ManagesKind,
			Setup: batchjob.SetupJob,
		},
		{
			Type:  emrcontainersjobrun.ManagesKind,
			Setup: emrcontainersjobrun.SetupJobRun,
		},
		{
			Type:  emrcontainersvirtualcluster.ManagesKind,
			Setup: emrcontainersvirtualcluster.SetupVirtualCluster,
		},
		{
			Type:  optiongroup.ManagesKind,
			Setup: optiongroup.SetupOptionGroup,
		},
		{
			Type:  autoscalinggroup.ManagesKind,
			Setup: autoscalinggroup.SetupAutoScalingGroup,
		},
	}

	exclude := filterCRDs.ExcludeCrds
	include := filterCRDs.IncludeCrds
	var appliedReconciler []func(ctrl.Manager, controller.Options) error

objectLoop:
	for _, obj := range allResources {

		// get the group and kind the controller manages
		managedKind := strings.ToLower(obj.Type())

		// include if matches entries in include list
		for _, includeReg := range include {
			matchedInclude, _ := regexp.MatchString(strings.ToLower(includeReg), managedKind)

			if matchedInclude {
				appliedReconciler = append(appliedReconciler, obj.Setup)
				continue objectLoop
			}
		}

		// exclude if matches entry in exclude list
		for _, excludeReg := range exclude {
			matchedExclude, _ := regexp.MatchString(strings.ToLower(excludeReg), managedKind)

			if matchedExclude {
				continue objectLoop
			}
		}

		// include if not expecially excluded
		appliedReconciler = append(appliedReconciler, obj.Setup)
	}

	// setup the filtered list of controllers
	for _, setup := range appliedReconciler {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}

	return config.Setup(mgr, o)
}
