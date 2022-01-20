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
	"fmt"
	"regexp"
	"time"

	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/crossplane-runtime/pkg/logging"

	acm "github.com/crossplane/provider-aws/pkg/controller/acm"
	acmpca_certificateauthority "github.com/crossplane/provider-aws/pkg/controller/acmpca/certificateauthority"
	acmpca_certificateauthoritypermission "github.com/crossplane/provider-aws/pkg/controller/acmpca/certificateauthoritypermission"
	apigatewayv2_api "github.com/crossplane/provider-aws/pkg/controller/apigatewayv2/api"
	apigatewayv2_apimapping "github.com/crossplane/provider-aws/pkg/controller/apigatewayv2/apimapping"
	apigatewayv2_authorizer "github.com/crossplane/provider-aws/pkg/controller/apigatewayv2/authorizer"
	apigatewayv2_deployment "github.com/crossplane/provider-aws/pkg/controller/apigatewayv2/deployment"
	apigatewayv2_domainname "github.com/crossplane/provider-aws/pkg/controller/apigatewayv2/domainname"
	apigatewayv2_integration "github.com/crossplane/provider-aws/pkg/controller/apigatewayv2/integration"
	apigatewayv2_integrationresponse "github.com/crossplane/provider-aws/pkg/controller/apigatewayv2/integrationresponse"
	apigatewayv2_model "github.com/crossplane/provider-aws/pkg/controller/apigatewayv2/model"
	apigatewayv2_route "github.com/crossplane/provider-aws/pkg/controller/apigatewayv2/route"
	apigatewayv2_routeresponse "github.com/crossplane/provider-aws/pkg/controller/apigatewayv2/routeresponse"
	apigatewayv2_stage "github.com/crossplane/provider-aws/pkg/controller/apigatewayv2/stage"
	apigatewayv2_vpclink "github.com/crossplane/provider-aws/pkg/controller/apigatewayv2/vpclink"
	athena_workgroup "github.com/crossplane/provider-aws/pkg/controller/athena/workgroup"
	cache "github.com/crossplane/provider-aws/pkg/controller/cache"
	cache_cachesubnetgroup "github.com/crossplane/provider-aws/pkg/controller/cache/cachesubnetgroup"
	cache_cluster "github.com/crossplane/provider-aws/pkg/controller/cache/cluster"
	cloudfront_cachepolicy "github.com/crossplane/provider-aws/pkg/controller/cloudfront/cachepolicy"
	cloudfront_cloudfrontoriginaccessidentity "github.com/crossplane/provider-aws/pkg/controller/cloudfront/cloudfrontoriginaccessidentity"
	cloudfront_distribution "github.com/crossplane/provider-aws/pkg/controller/cloudfront/distribution"
	cloudwatchlogs_loggroup "github.com/crossplane/provider-aws/pkg/controller/cloudwatchlogs/loggroup"
	"github.com/crossplane/provider-aws/pkg/controller/config"
	database "github.com/crossplane/provider-aws/pkg/controller/database"
	database_dbsubnetgroup "github.com/crossplane/provider-aws/pkg/controller/database/dbsubnetgroup"
	docdb_dbcluster "github.com/crossplane/provider-aws/pkg/controller/docdb/dbcluster"
	docdb_dbclusterparametergroup "github.com/crossplane/provider-aws/pkg/controller/docdb/dbclusterparametergroup"
	docdb_dbinstance "github.com/crossplane/provider-aws/pkg/controller/docdb/dbinstance"
	docdb_dbsubnetgroup "github.com/crossplane/provider-aws/pkg/controller/docdb/dbsubnetgroup"
	dynamodb_backup "github.com/crossplane/provider-aws/pkg/controller/dynamodb/backup"
	dynamodb_globaltable "github.com/crossplane/provider-aws/pkg/controller/dynamodb/globaltable"
	dynamodb_table "github.com/crossplane/provider-aws/pkg/controller/dynamodb/table"
	ec2_address "github.com/crossplane/provider-aws/pkg/controller/ec2/address"
	ec2_instance "github.com/crossplane/provider-aws/pkg/controller/ec2/instance"
	ec2_internetgateway "github.com/crossplane/provider-aws/pkg/controller/ec2/internetgateway"
	ec2_launchtemplate "github.com/crossplane/provider-aws/pkg/controller/ec2/launchtemplate"
	ec2_launchtemplateversion "github.com/crossplane/provider-aws/pkg/controller/ec2/launchtemplateversion"
	ec2_natgateway "github.com/crossplane/provider-aws/pkg/controller/ec2/natgateway"
	ec2_route "github.com/crossplane/provider-aws/pkg/controller/ec2/route"
	ec2_routetable "github.com/crossplane/provider-aws/pkg/controller/ec2/routetable"
	ec2_securitygroup "github.com/crossplane/provider-aws/pkg/controller/ec2/securitygroup"
	ec2_subnet "github.com/crossplane/provider-aws/pkg/controller/ec2/subnet"
	ec2_transitgateway "github.com/crossplane/provider-aws/pkg/controller/ec2/transitgateway"
	ec2_transitgatewayroute "github.com/crossplane/provider-aws/pkg/controller/ec2/transitgatewayroute"
	ec2_transitgatewayroutetable "github.com/crossplane/provider-aws/pkg/controller/ec2/transitgatewayroutetable"
	ec2_transitgatewayvpcattachment "github.com/crossplane/provider-aws/pkg/controller/ec2/transitgatewayvpcattachment"
	ec2_volume "github.com/crossplane/provider-aws/pkg/controller/ec2/volume"
	ec2_vpc "github.com/crossplane/provider-aws/pkg/controller/ec2/vpc"
	ec2_vpccidrblock "github.com/crossplane/provider-aws/pkg/controller/ec2/vpccidrblock"
	ec2_vpcendpoint "github.com/crossplane/provider-aws/pkg/controller/ec2/vpcendpoint"
	ec2_vpcendpointserviceconfiguration "github.com/crossplane/provider-aws/pkg/controller/ec2/vpcendpointserviceconfiguration"
	ec2_vpcpeeringconnection "github.com/crossplane/provider-aws/pkg/controller/ec2/vpcpeeringconnection"
	ecr_repository "github.com/crossplane/provider-aws/pkg/controller/ecr/repository"
	ecr_repositorypolicy "github.com/crossplane/provider-aws/pkg/controller/ecr/repositorypolicy"
	efs_filesystem "github.com/crossplane/provider-aws/pkg/controller/efs/filesystem"
	efs_mounttarget "github.com/crossplane/provider-aws/pkg/controller/efs/mounttarget"
	eks "github.com/crossplane/provider-aws/pkg/controller/eks"
	eks_addon "github.com/crossplane/provider-aws/pkg/controller/eks/addon"
	eks_fargateprofile "github.com/crossplane/provider-aws/pkg/controller/eks/fargateprofile"
	eks_identityproviderconfig "github.com/crossplane/provider-aws/pkg/controller/eks/identityproviderconfig"
	eks_nodegroup "github.com/crossplane/provider-aws/pkg/controller/eks/nodegroup"
	elasticloadbalancing_elb "github.com/crossplane/provider-aws/pkg/controller/elasticloadbalancing/elb"
	elasticloadbalancing_elbattachment "github.com/crossplane/provider-aws/pkg/controller/elasticloadbalancing/elbattachment"
	elbv2_listener "github.com/crossplane/provider-aws/pkg/controller/elbv2/listener"
	elbv2_loadbalancer "github.com/crossplane/provider-aws/pkg/controller/elbv2/loadbalancer"
	elbv2_targetgroup "github.com/crossplane/provider-aws/pkg/controller/elbv2/targetgroup"
	glue_classifier "github.com/crossplane/provider-aws/pkg/controller/glue/classifier"
	glue_connection "github.com/crossplane/provider-aws/pkg/controller/glue/connection"
	glue_crawler "github.com/crossplane/provider-aws/pkg/controller/glue/crawler"
	glue_database "github.com/crossplane/provider-aws/pkg/controller/glue/database"
	glue_job "github.com/crossplane/provider-aws/pkg/controller/glue/job"
	glue_securityconfiguration "github.com/crossplane/provider-aws/pkg/controller/glue/securityconfiguration"
	iam_accesskey "github.com/crossplane/provider-aws/pkg/controller/iam/accesskey"
	iam_group "github.com/crossplane/provider-aws/pkg/controller/iam/group"
	iam_grouppolicyattachment "github.com/crossplane/provider-aws/pkg/controller/iam/grouppolicyattachment"
	iam_groupusermembership "github.com/crossplane/provider-aws/pkg/controller/iam/groupusermembership"
	iam_openidconnectprovider "github.com/crossplane/provider-aws/pkg/controller/iam/openidconnectprovider"
	iam_policy "github.com/crossplane/provider-aws/pkg/controller/iam/policy"
	iam_role "github.com/crossplane/provider-aws/pkg/controller/iam/role"
	iam_rolepolicyattachment "github.com/crossplane/provider-aws/pkg/controller/iam/rolepolicyattachment"
	iam_user "github.com/crossplane/provider-aws/pkg/controller/iam/user"
	iam_userpolicyattachment "github.com/crossplane/provider-aws/pkg/controller/iam/userpolicyattachment"
	iot_policy "github.com/crossplane/provider-aws/pkg/controller/iot/policy"
	iot_thing "github.com/crossplane/provider-aws/pkg/controller/iot/thing"
	kafka_cluster "github.com/crossplane/provider-aws/pkg/controller/kafka/cluster"
	kafka_configuration "github.com/crossplane/provider-aws/pkg/controller/kafka/configuration"
	kinesis_stream "github.com/crossplane/provider-aws/pkg/controller/kinesis/stream"
	kms_alias "github.com/crossplane/provider-aws/pkg/controller/kms/alias"
	kms_key "github.com/crossplane/provider-aws/pkg/controller/kms/key"
	lambda_function "github.com/crossplane/provider-aws/pkg/controller/lambda/function"
	mq_broker "github.com/crossplane/provider-aws/pkg/controller/mq/broker"
	mq_user "github.com/crossplane/provider-aws/pkg/controller/mq/user"
	notification_snssubscription "github.com/crossplane/provider-aws/pkg/controller/notification/snssubscription"
	notification_snstopic "github.com/crossplane/provider-aws/pkg/controller/notification/snstopic"
	ram_resourceshare "github.com/crossplane/provider-aws/pkg/controller/ram/resourceshare"
	rds_dbcluster "github.com/crossplane/provider-aws/pkg/controller/rds/dbcluster"
	rds_dbclusterparametergroup "github.com/crossplane/provider-aws/pkg/controller/rds/dbclusterparametergroup"
	rds_dbinstance "github.com/crossplane/provider-aws/pkg/controller/rds/dbinstance"
	rds_dbparametergroup "github.com/crossplane/provider-aws/pkg/controller/rds/dbparametergroup"
	rds_globalcluster "github.com/crossplane/provider-aws/pkg/controller/rds/globalcluster"
	redshift "github.com/crossplane/provider-aws/pkg/controller/redshift"
	route53_hostedzone "github.com/crossplane/provider-aws/pkg/controller/route53/hostedzone"
	route53_resourcerecordset "github.com/crossplane/provider-aws/pkg/controller/route53/resourcerecordset"
	route53resolver_resolverendpoint "github.com/crossplane/provider-aws/pkg/controller/route53resolver/resolverendpoint"
	route53resolver_resolverrule "github.com/crossplane/provider-aws/pkg/controller/route53resolver/resolverrule"
	s3 "github.com/crossplane/provider-aws/pkg/controller/s3"
	s3_bucketpolicy "github.com/crossplane/provider-aws/pkg/controller/s3/bucketpolicy"
	secretsmanager_secret "github.com/crossplane/provider-aws/pkg/controller/secretsmanager/secret"
	servicediscovery_httpnamespace "github.com/crossplane/provider-aws/pkg/controller/servicediscovery/httpnamespace"
	servicediscovery_privatednsnamespace "github.com/crossplane/provider-aws/pkg/controller/servicediscovery/privatednsnamespace"
	servicediscovery_publicdnsnamespace "github.com/crossplane/provider-aws/pkg/controller/servicediscovery/publicdnsnamespace"
	sfn_activity "github.com/crossplane/provider-aws/pkg/controller/sfn/activity"
	sfn_statemachine "github.com/crossplane/provider-aws/pkg/controller/sfn/statemachine"
	sqs_queue "github.com/crossplane/provider-aws/pkg/controller/sqs/queue"
	transfer_server "github.com/crossplane/provider-aws/pkg/controller/transfer/server"
	transfer_user "github.com/crossplane/provider-aws/pkg/controller/transfer/user"
)

// Setup creates all AWS controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration, pattern string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid pattern %s to enable controllers: %v", pattern, err)
	}
	for name, setup := range map[string]func(ctrl.Manager, logging.Logger, workqueue.RateLimiter, time.Duration) error{
		"acm/Certificate": acm.SetupCertificate,
		"acmpca/certificateauthority/CertificateAuthority":                     acmpca_certificateauthority.SetupCertificateAuthority,
		"acmpca/certificateauthoritypermission/CertificateAuthorityPermission": acmpca_certificateauthoritypermission.SetupCertificateAuthorityPermission,
		"apigatewayv2/api/API":                                                     apigatewayv2_api.SetupAPI,
		"apigatewayv2/apimapping/APIMapping":                                       apigatewayv2_apimapping.SetupAPIMapping,
		"apigatewayv2/authorizer/Authorizer":                                       apigatewayv2_authorizer.SetupAuthorizer,
		"apigatewayv2/deployment/Deployment":                                       apigatewayv2_deployment.SetupDeployment,
		"apigatewayv2/domainname/DomainName":                                       apigatewayv2_domainname.SetupDomainName,
		"apigatewayv2/integration/Integration":                                     apigatewayv2_integration.SetupIntegration,
		"apigatewayv2/integrationresponse/IntegrationResponse":                     apigatewayv2_integrationresponse.SetupIntegrationResponse,
		"apigatewayv2/model/Model":                                                 apigatewayv2_model.SetupModel,
		"apigatewayv2/route/Route":                                                 apigatewayv2_route.SetupRoute,
		"apigatewayv2/routeresponse/RouteResponse":                                 apigatewayv2_routeresponse.SetupRouteResponse,
		"apigatewayv2/stage/Stage":                                                 apigatewayv2_stage.SetupStage,
		"apigatewayv2/vpclink/VPCLink":                                             apigatewayv2_vpclink.SetupVPCLink,
		"athena/workgroup/WorkGroup":                                               athena_workgroup.SetupWorkGroup,
		"cache/cachesubnetgroup/CacheSubnetGroup":                                  cache_cachesubnetgroup.SetupCacheSubnetGroup,
		"cache/cluster/CacheCluster":                                               cache_cluster.SetupCacheCluster,
		"cache/ReplicationGroup":                                                   cache.SetupReplicationGroup,
		"cloudfront/cachepolicy/CachePolicy":                                       cloudfront_cachepolicy.SetupCachePolicy,
		"cloudfront/cloudfrontoriginaccessidentity/CloudFrontOriginAccessIdentity": cloudfront_cloudfrontoriginaccessidentity.SetupCloudFrontOriginAccessIdentity,
		"cloudfront/distribution/Distribution":                                     cloudfront_distribution.SetupDistribution,
		"cloudwatchlogs/loggroup/LogGroup":                                         cloudwatchlogs_loggroup.SetupLogGroup,
		"database/dbsubnetgroup/DBSubnetGroup":                                     database_dbsubnetgroup.SetupDBSubnetGroup,
		"database/RDSInstance":                                                     database.SetupRDSInstance,
		"docdb/dbcluster/DBCluster":                                                docdb_dbcluster.SetupDBCluster,
		"docdb/dbclusterparametergroup/DBClusterParameterGroup":                    docdb_dbclusterparametergroup.SetupDBClusterParameterGroup,
		"docdb/dbinstance/DBInstance":                                              docdb_dbinstance.SetupDBInstance,
		"docdb/dbsubnetgroup/DBSubnetGroup":                                        docdb_dbsubnetgroup.SetupDBSubnetGroup,
		"dynamodb/backup/Backup":                                                   dynamodb_backup.SetupBackup,
		"dynamodb/globaltable/GlobalTable":                                         dynamodb_globaltable.SetupGlobalTable,
		"dynamodb/table/Table":                                                     dynamodb_table.SetupTable,
		"ec2/address/Address":                                                      ec2_address.SetupAddress,
		"ec2/instance/Instance":                                                    ec2_instance.SetupInstance,
		"ec2/internetgateway/InternetGateway":                                      ec2_internetgateway.SetupInternetGateway,
		"ec2/launchtemplate/LaunchTemplate":                                        ec2_launchtemplate.SetupLaunchTemplate,
		"ec2/launchtemplateversion/LaunchTemplateVersion":                          ec2_launchtemplateversion.SetupLaunchTemplateVersion,
		"ec2/natgateway/NatGateway":                                                ec2_natgateway.SetupNatGateway,
		"ec2/route/Route":                                                          ec2_route.SetupRoute,
		"ec2/routetable/RouteTable":                                                ec2_routetable.SetupRouteTable,
		"ec2/securitygroup/SecurityGroup":                                          ec2_securitygroup.SetupSecurityGroup,
		"ec2/subnet/Subnet":                                                        ec2_subnet.SetupSubnet,
		"ec2/transitgateway/TransitGateway":                                        ec2_transitgateway.SetupTransitGateway,
		"ec2/transitgatewayroute/TransitGatewayRoute":                              ec2_transitgatewayroute.SetupTransitGatewayRoute,
		"ec2/transitgatewayroutetable/TransitGatewayRouteTable":                    ec2_transitgatewayroutetable.SetupTransitGatewayRouteTable,
		"ec2/transitgatewayvpcattachment/TransitGatewayVPCAttachment":              ec2_transitgatewayvpcattachment.SetupTransitGatewayVPCAttachment,
		"ec2/volume/Volume":                                                        ec2_volume.SetupVolume,
		"ec2/vpc/VPC":                                                              ec2_vpc.SetupVPC,
		"ec2/vpccidrblock/VPCCIDRBlock":                                            ec2_vpccidrblock.SetupVPCCIDRBlock,
		"ec2/vpcendpoint/VPCEndpoint":                                              ec2_vpcendpoint.SetupVPCEndpoint,
		"ec2/vpcendpointserviceconfiguration/VPCEndpointServiceConfiguration":      ec2_vpcendpointserviceconfiguration.SetupVPCEndpointServiceConfiguration,
		"ec2/vpcpeeringconnection/VPCPeeringConnection":                            ec2_vpcpeeringconnection.SetupVPCPeeringConnection,
		"ecr/repository/Repository":                                                ecr_repository.SetupRepository,
		"ecr/repositorypolicy/RepositoryPolicy":                                    ecr_repositorypolicy.SetupRepositoryPolicy,
		"efs/filesystem/FileSystem":                                                efs_filesystem.SetupFileSystem,
		"efs/mounttarget/MountTarget":                                              efs_mounttarget.SetupMountTarget,
		"eks/addon/Addon":                                                          eks_addon.SetupAddon,
		"eks/Cluster":                                                              eks.SetupCluster,
		"eks/fargateprofile/FargateProfile":                                        eks_fargateprofile.SetupFargateProfile,
		"eks/identityproviderconfig/IdentityProviderConfig":                        eks_identityproviderconfig.SetupIdentityProviderConfig,
		"eks/nodegroup/NodeGroup":                                                  eks_nodegroup.SetupNodeGroup,
		"elasticloadbalancing/elb/ELB":                                             elasticloadbalancing_elb.SetupELB,
		"elasticloadbalancing/elbattachment/ELBAttachment":                         elasticloadbalancing_elbattachment.SetupELBAttachment,
		"elbv2/listener/Listener":                                                  elbv2_listener.SetupListener,
		"elbv2/loadbalancer/LoadBalancer":                                          elbv2_loadbalancer.SetupLoadBalancer,
		"elbv2/targetgroup/TargetGroup":                                            elbv2_targetgroup.SetupTargetGroup,
		"glue/classifier/Classifier":                                               glue_classifier.SetupClassifier,
		"glue/connection/Connection":                                               glue_connection.SetupConnection,
		"glue/crawler/Crawler":                                                     glue_crawler.SetupCrawler,
		"glue/database/Database":                                                   glue_database.SetupDatabase,
		"glue/job/Job":                                                             glue_job.SetupJob,
		"glue/securityconfiguration/SecurityConfiguration":                         glue_securityconfiguration.SetupSecurityConfiguration,
		"iam/accesskey/AccessKey":                                                  iam_accesskey.SetupAccessKey,
		"iam/group/Group":                                                          iam_group.SetupGroup,
		"iam/grouppolicyattachment/GroupPolicyAttachment":                          iam_grouppolicyattachment.SetupGroupPolicyAttachment,
		"iam/groupusermembership/GroupUserMembership":                              iam_groupusermembership.SetupGroupUserMembership,
		"iam/openidconnectprovider/OpenIDConnectProvider":                          iam_openidconnectprovider.SetupOpenIDConnectProvider,
		"iam/policy/Policy":                                                        iam_policy.SetupPolicy,
		"iam/role/Role":                                                            iam_role.SetupRole,
		"iam/rolepolicyattachment/RolePolicyAttachment":                            iam_rolepolicyattachment.SetupRolePolicyAttachment,
		"iam/user/User":                                                            iam_user.SetupUser,
		"iam/userpolicyattachment/UserPolicyAttachment":                            iam_userpolicyattachment.SetupUserPolicyAttachment,
		"iot/policy/Policy":                                                        iot_policy.SetupPolicy,
		"iot/thing/Thing":                                                          iot_thing.SetupThing,
		"kafka/cluster/Cluster":                                                    kafka_cluster.SetupCluster,
		"kafka/configuration/Configuration":                                        kafka_configuration.SetupConfiguration,
		"kinesis/stream/Stream":                                                    kinesis_stream.SetupStream,
		"kms/alias/Alias":                                                          kms_alias.SetupAlias,
		"kms/key/Key":                                                              kms_key.SetupKey,
		"lambda/function/Function":                                                 lambda_function.SetupFunction,
		"mq/broker/Broker":                                                         mq_broker.SetupBroker,
		"mq/user/User":                                                             mq_user.SetupUser,
		"notification/snssubscription/Subscription":                                notification_snssubscription.SetupSubscription,
		"notification/snstopic/SNSTopic":                                           notification_snstopic.SetupSNSTopic,
		"ram/resourceshare/ResourceShare":                                          ram_resourceshare.SetupResourceShare,
		"rds/dbcluster/DBCluster":                                                  rds_dbcluster.SetupDBCluster,
		"rds/dbclusterparametergroup/DBClusterParameterGroup":                      rds_dbclusterparametergroup.SetupDBClusterParameterGroup,
		"rds/dbinstance/DBInstance":                                                rds_dbinstance.SetupDBInstance,
		"rds/dbparametergroup/DBParameterGroup":                                    rds_dbparametergroup.SetupDBParameterGroup,
		"rds/globalcluster/GlobalCluster":                                          rds_globalcluster.SetupGlobalCluster,
		"redshift/Cluster":                                                         redshift.SetupCluster,
		"route53/hostedzone/HostedZone":                                            route53_hostedzone.SetupHostedZone,
		"route53/resourcerecordset/ResourceRecordSet":                              route53_resourcerecordset.SetupResourceRecordSet,
		"route53resolver/resolverendpoint/ResolverEndpoint":                        route53resolver_resolverendpoint.SetupResolverEndpoint,
		"route53resolver/resolverrule/ResolverRule":                                route53resolver_resolverrule.SetupResolverRule,
		"s3/Bucket":                                                s3.SetupBucket,
		"s3/bucketpolicy/BucketPolicy":                             s3_bucketpolicy.SetupBucketPolicy,
		"secretsmanager/secret/Secret":                             secretsmanager_secret.SetupSecret,
		"servicediscovery/httpnamespace/HTTPNamespace":             servicediscovery_httpnamespace.SetupHTTPNamespace,
		"servicediscovery/privatednsnamespace/PrivateDNSNamespace": servicediscovery_privatednsnamespace.SetupPrivateDNSNamespace,
		"servicediscovery/publicdnsnamespace/PublicDNSNamespace":   servicediscovery_publicdnsnamespace.SetupPublicDNSNamespace,
		"sfn/activity/Activity":                                    sfn_activity.SetupActivity,
		"sfn/statemachine/StateMachine":                            sfn_statemachine.SetupStateMachine,
		"sqs/queue/Queue":                                          sqs_queue.SetupQueue,
		"transfer/server/Server":                                   transfer_server.SetupServer,
		"transfer/user/User":                                       transfer_user.SetupUser,
	} {
		if re.MatchString(name) {
			if err := setup(mgr, l, rl, poll); err != nil {
				return err
			}
		}
	}

	return config.Setup(mgr, l, rl)
}
