/*
Copyright 2023 The Crossplane Authors.

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

package ec2

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/address"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/flowlog"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/instance"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/internetgateway"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/launchtemplate"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/launchtemplateversion"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/natgateway"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/route"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/routetable"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/securitygroup"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/securitygrouprule"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/subnet"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/transitgateway"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/transitgatewayroute"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/transitgatewayroutetable"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/transitgatewayvpcattachment"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/volume"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/vpc"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/vpccidrblock"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/vpcendpoint"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/vpcendpointserviceconfiguration"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ec2/vpcpeeringconnection"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/controller"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/setup"
)

// Setup ec2 controllers.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	batch := setup.NewBatch(mgr, o, "ec2")
	batch.Add("address", address.SetupAddress)
	batch.Add("flowlog", flowlog.SetupFlowLog)
	batch.Add("instance", instance.SetupInstance)
	batch.Add("internetgateway", internetgateway.SetupInternetGateway)
	batch.Add("launchtemplate", launchtemplate.SetupLaunchTemplate)
	batch.Add("launchtemplateversion", launchtemplateversion.SetupLaunchTemplateVersion)
	batch.Add("natgateway", natgateway.SetupNatGateway)
	batch.Add("route", route.SetupRoute)
	batch.Add("routetable", routetable.SetupRouteTable)
	batch.Add("securitygroup", securitygroup.SetupSecurityGroup)
	batch.Add("securitygrouprule", securitygrouprule.SetupSecurityGroupRule)
	batch.Add("subnet", subnet.SetupSubnet)
	batch.Add("transitgateway", transitgateway.SetupTransitGateway)
	batch.Add("transitgatewayroute", transitgatewayroute.SetupTransitGatewayRoute)
	batch.Add("transitgatewayroutetable", transitgatewayroutetable.SetupTransitGatewayRouteTable)
	batch.Add("transitgatewayvpcattachment", transitgatewayvpcattachment.SetupTransitGatewayVPCAttachment)
	batch.Add("volume", volume.SetupVolume)
	batch.Add("vpc", vpc.SetupVPC)
	batch.Add("vpccidrblock", vpccidrblock.SetupVPCCIDRBlock)
	batch.Add("vpcendpoint", vpcendpoint.SetupVPCEndpoint)
	batch.Add("vpcendpointserviceconfiguration", vpcendpointserviceconfiguration.SetupVPCEndpointServiceConfiguration)
	batch.Add("vpcpeeringconnection", vpcpeeringconnection.SetupVPCPeeringConnection)
	return batch.Run()
}
