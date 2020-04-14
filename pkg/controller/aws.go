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

	"github.com/crossplane/provider-aws/pkg/controller/cache"
	"github.com/crossplane/provider-aws/pkg/controller/cache/cachesubnetgroup"
	"github.com/crossplane/provider-aws/pkg/controller/compute"
	"github.com/crossplane/provider-aws/pkg/controller/database"
	"github.com/crossplane/provider-aws/pkg/controller/database/dbsubnetgroup"
	"github.com/crossplane/provider-aws/pkg/controller/database/dynamodb"
	"github.com/crossplane/provider-aws/pkg/controller/identity/iamrole"
	"github.com/crossplane/provider-aws/pkg/controller/identity/iamrolepolicyattachment"
	"github.com/crossplane/provider-aws/pkg/controller/identity/iamuser"
	"github.com/crossplane/provider-aws/pkg/controller/identity/iamuserpolicyattachment"
	"github.com/crossplane/provider-aws/pkg/controller/network/internetgateway"
	"github.com/crossplane/provider-aws/pkg/controller/network/routetable"
	"github.com/crossplane/provider-aws/pkg/controller/network/securitygroup"
	"github.com/crossplane/provider-aws/pkg/controller/network/subnet"
	"github.com/crossplane/provider-aws/pkg/controller/network/vpc"
	"github.com/crossplane/provider-aws/pkg/controller/s3"
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
		s3.SetupBucketClaimScheduling,
		s3.SetupBucketClaimDefaulting,
		s3.SetupBucketClaimBinding,
		s3.SetupS3Bucket,
		iamuser.SetupIAMUser,
		iamrole.SetupIAMRole,
		iamuserpolicyattachment.SetupIAMUserPolicyAttachment,
		iamrolepolicyattachment.SetupIAMRolePolicyAttachment,
		vpc.SetupVPC,
		subnet.SetupSubnet,
		securitygroup.SetupSecurityGroup,
		internetgateway.SetupInternetGateway,
		routetable.SetupRouteTable,
		dbsubnetgroup.SetupDBSubnetGroup,
		dynamodb.SetupDynamoTable,
	} {
		if err := setup(mgr, l); err != nil {
			return err
		}
	}

	return nil
}
