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

	"github.com/crossplane/crossplane-runtime/pkg/logging"

	"github.com/crossplane/provider-aws/pkg/controller/acm"
	"github.com/crossplane/provider-aws/pkg/controller/acmpca/certificateauthority"
	"github.com/crossplane/provider-aws/pkg/controller/acmpca/certificateauthoritypermission"
	"github.com/crossplane/provider-aws/pkg/controller/applicationintegration/sqs"
	"github.com/crossplane/provider-aws/pkg/controller/cache"
	"github.com/crossplane/provider-aws/pkg/controller/cache/cachesubnetgroup"
	"github.com/crossplane/provider-aws/pkg/controller/compute"
	"github.com/crossplane/provider-aws/pkg/controller/database"
	"github.com/crossplane/provider-aws/pkg/controller/database/dbsubnetgroup"
	"github.com/crossplane/provider-aws/pkg/controller/database/dynamodb"
	"github.com/crossplane/provider-aws/pkg/controller/ec2/internetgateway"
	"github.com/crossplane/provider-aws/pkg/controller/ec2/routetable"
	"github.com/crossplane/provider-aws/pkg/controller/ec2/securitygroup"
	"github.com/crossplane/provider-aws/pkg/controller/ec2/subnet"
	"github.com/crossplane/provider-aws/pkg/controller/ec2/vpc"
	"github.com/crossplane/provider-aws/pkg/controller/eks"
	"github.com/crossplane/provider-aws/pkg/controller/eks/nodegroup"
	"github.com/crossplane/provider-aws/pkg/controller/elasticloadbalancing/elb"
	"github.com/crossplane/provider-aws/pkg/controller/elasticloadbalancing/elbattachment"
	"github.com/crossplane/provider-aws/pkg/controller/identity/iamgroup"
	"github.com/crossplane/provider-aws/pkg/controller/identity/iamgrouppolicyattachment"
	"github.com/crossplane/provider-aws/pkg/controller/identity/iamgroupusermembership"
	"github.com/crossplane/provider-aws/pkg/controller/identity/iampolicy"
	"github.com/crossplane/provider-aws/pkg/controller/identity/iamrole"
	"github.com/crossplane/provider-aws/pkg/controller/identity/iamrolepolicyattachment"
	"github.com/crossplane/provider-aws/pkg/controller/identity/iamuser"
	"github.com/crossplane/provider-aws/pkg/controller/identity/iamuserpolicyattachment"
	"github.com/crossplane/provider-aws/pkg/controller/notification/snssubscription"
	"github.com/crossplane/provider-aws/pkg/controller/notification/snstopic"
	"github.com/crossplane/provider-aws/pkg/controller/redshift"
	"github.com/crossplane/provider-aws/pkg/controller/route53/hostedzone"
	"github.com/crossplane/provider-aws/pkg/controller/route53/resourcerecordset"
	"github.com/crossplane/provider-aws/pkg/controller/s3"
	"github.com/crossplane/provider-aws/pkg/controller/s3/s3bucketpolicy"
)

// Setup creates all AWS controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, l logging.Logger) error {
	for _, setup := range []func(ctrl.Manager, logging.Logger) error{
		cache.SetupReplicationGroupClaimScheduling,
		cache.SetupReplicationGroupClaimDefaulting,
		cache.SetupReplicationGroupClaimBinding,
		cache.SetupReplicationGroup,
		cachesubnetgroup.SetupCacheSubnetGroup,
		compute.SetupEKSClusterClaimScheduling,
		compute.SetupEKSClusterClaimDefaulting,
		compute.SetupEKSClusterClaimBinding,
		compute.SetupEKSClusterSecret,
		compute.SetupEKSClusterTarget,
		compute.SetupEKSCluster,
		database.SetupPostgreSQLInstanceClaimScheduling,
		database.SetupPostgreSQLInstanceClaimDefaulting,
		database.SetupPostgreSQLInstanceClaimBinding,
		database.SetupMySQLInstanceClaimScheduling,
		database.SetupMySQLInstanceClaimDefaulting,
		database.SetupMySQLInstanceClaimBinding,
		database.SetupRDSInstance,
		eks.SetupCluster,
		eks.SetupClusterSecret,
		eks.SetupClusterTarget,
		elb.SetupELB,
		elbattachment.SetupELBAttachment,
		nodegroup.SetupNodeGroup,
		s3.SetupBucketClaimScheduling,
		s3.SetupBucketClaimDefaulting,
		s3.SetupBucketClaimBinding,
		s3.SetupS3Bucket,
		s3bucketpolicy.SetupS3BucketPolicy,
		iamuser.SetupIAMUser,
		iamgroup.SetupIAMGroup,
		iampolicy.SetupIAMPolicy,
		iamrole.SetupIAMRole,
		iamgroupusermembership.SetupIAMGroupUserMembership,
		iamuserpolicyattachment.SetupIAMUserPolicyAttachment,
		iamgrouppolicyattachment.SetupIAMGroupPolicyAttachment,
		iamrolepolicyattachment.SetupIAMRolePolicyAttachment,
		vpc.SetupVPC,
		subnet.SetupSubnet,
		securitygroup.SetupSecurityGroup,
		internetgateway.SetupInternetGateway,
		routetable.SetupRouteTable,
		dbsubnetgroup.SetupDBSubnetGroup,
		certificateauthority.SetupCertificateAuthority,
		certificateauthoritypermission.SetupCertificateAuthorityPermission,
		acm.SetupCertificate,
		dynamodb.SetupDynamoTable,
		resourcerecordset.SetupResourceRecordSet,
		hostedzone.SetupHostedZone,
		snstopic.SetupSNSTopic,
		snssubscription.SetupSubscription,
		sqs.SetupQueue,
		redshift.SetupCluster,
	} {
		if err := setup(mgr, l); err != nil {
			return err
		}
	}

	return nil
}
