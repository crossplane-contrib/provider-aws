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
func Setup(mgr ctrl.Manager, o controller.OptionsSet) error {
	batch := setup.NewBatch(mgr, o, "ec2")
	batch.AddXp("address", address.SetupAddress)
	batch.AddXp("flowlog", flowlog.SetupFlowLog)
	batch.Add("instance", instance.SetupInstance)
	batch.AddXp("internetgateway", internetgateway.SetupInternetGateway)
	batch.AddXp("launchtemplate", launchtemplate.SetupLaunchTemplate)
	batch.AddXp("launchtemplateversion", launchtemplateversion.SetupLaunchTemplateVersion)
	batch.AddXp("natgateway", natgateway.SetupNatGateway)
	batch.AddXp("route", route.SetupRoute)
	batch.AddXp("routetable", routetable.SetupRouteTable)
	batch.AddXp("securitygroup", securitygroup.SetupSecurityGroup)
	batch.AddXp("securitygrouprule", securitygrouprule.SetupSecurityGroupRule)
	batch.AddXp("subnet", subnet.SetupSubnet)
	batch.AddXp("transitgateway", transitgateway.SetupTransitGateway)
	batch.AddXp("transitgatewayroute", transitgatewayroute.SetupTransitGatewayRoute)
	batch.AddXp("transitgatewayroutetable", transitgatewayroutetable.SetupTransitGatewayRouteTable)
	batch.AddXp("transitgatewayvpcattachment", transitgatewayvpcattachment.SetupTransitGatewayVPCAttachment)
	batch.AddXp("volume", volume.SetupVolume)
	batch.AddXp("vpc", vpc.SetupVPC)
	batch.AddXp("vpccidrblock", vpccidrblock.SetupVPCCIDRBlock)
	batch.AddXp("vpcendpoint", vpcendpoint.SetupVPCEndpoint)
	batch.AddXp("vpcendpointserviceconfiguration", vpcendpointserviceconfiguration.SetupVPCEndpointServiceConfiguration)
	batch.AddXp("vpcpeeringconnection", vpcpeeringconnection.SetupVPCPeeringConnection)
	return batch.Run()
}
