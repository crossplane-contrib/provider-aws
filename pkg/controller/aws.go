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

	"github.com/crossplaneio/stack-aws/pkg/controller/cache"
	"github.com/crossplaneio/stack-aws/pkg/controller/compute"
	"github.com/crossplaneio/stack-aws/pkg/controller/database"
	"github.com/crossplaneio/stack-aws/pkg/controller/database/dbsubnetgroup"
	"github.com/crossplaneio/stack-aws/pkg/controller/identity/iamrole"
	"github.com/crossplaneio/stack-aws/pkg/controller/identity/iamrolepolicyattachment"
	"github.com/crossplaneio/stack-aws/pkg/controller/network/internetgateway"
	"github.com/crossplaneio/stack-aws/pkg/controller/network/routetable"
	"github.com/crossplaneio/stack-aws/pkg/controller/network/securitygroup"
	"github.com/crossplaneio/stack-aws/pkg/controller/network/subnet"
	"github.com/crossplaneio/stack-aws/pkg/controller/network/vpc"
	"github.com/crossplaneio/stack-aws/pkg/controller/s3"
)

// Controllers passes down config and adds individual controllers to the manager.
type Controllers struct{}

// SetupWithManager adds all AWS controllers to the manager.
func (c *Controllers) SetupWithManager(mgr ctrl.Manager) error {

	controllers := []interface {
		SetupWithManager(ctrl.Manager) error
	}{
		&cache.ReplicationGroupClaimSchedulingController{},
		&cache.ReplicationGroupClaimDefaultingController{},
		&cache.ReplicationGroupClaimController{},
		&cache.ReplicationGroupController{},
		&compute.EKSClusterClaimSchedulingController{},
		&compute.EKSClusterClaimDefaultingController{},
		&compute.EKSClusterClaimController{},
		&compute.EKSClusterSecretController{},
		&compute.EKSClusterController{},
		&database.PostgreSQLInstanceClaimSchedulingController{},
		&database.PostgreSQLInstanceClaimDefaultingController{},
		&database.PostgreSQLInstanceClaimController{},
		&database.MySQLInstanceClaimSchedulingController{},
		&database.MySQLInstanceClaimDefaultingController{},
		&database.MySQLInstanceClaimController{},
		&database.RDSInstanceController{},
		&s3.BucketClaimSchedulingController{},
		&s3.BucketClaimDefaultingController{},
		&s3.BucketClaimController{},
		&s3.BucketController{},
		&iamrole.Controller{},
		&iamrolepolicyattachment.Controller{},
		&vpc.Controller{},
		&subnet.Controller{},
		&securitygroup.Controller{},
		&internetgateway.Controller{},
		&routetable.Controller{},
		&dbsubnetgroup.Controller{},
	}

	for _, c := range controllers {
		if err := c.SetupWithManager(mgr); err != nil {
			return err
		}
	}
	return nil
}
