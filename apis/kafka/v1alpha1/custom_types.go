/*
Copyright 2021 The Crossplane Authors.

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

package v1alpha1

import xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

// CustomConfigurationParameters contains the additional fields for ConfigurationParameters.
type CustomConfigurationParameters struct {
	// kafka server configurations
	Properties []string `json:"properties"`
}

// CustomClusterParameters contains the additional fields for ClusterParameters.
type CustomClusterParameters struct {
	// Information about the brokers
	CustomBrokerNodeGroupInfo *CustomBrokerNodeGroupInfo `json:"brokerNodeGroupInfo,omitempty"`

	// Represents the configuration that you want MSK to use for the cluster.
	CustomConfigurationInfo *CustomConfigurationInfo `json:"configurationInfo,omitempty"`
}

// CustomConfigurationInfo contains the additional fields for ConfigurationInfo.
type CustomConfigurationInfo struct {
	// ARN of the configuration to use.
	// +optional
	// +crossplane:generate:reference:type=Configuration
	ARN *string `json:"arn,omitempty"`

	// ARNRef is a reference to a Kafka Configuration used to set ARN.
	// +optional
	ARNRef *xpv1.Reference `json:"arnRef,omitempty"`

	// ARNSelector selects a reference to a Kafka Configuration used to set ARN.
	// +optional
	ARNSelector *xpv1.Selector `json:"arnSelector,omitempty"`

	Revision *int64 `json:"revision,omitempty"`
}

// CustomBrokerNodeGroupInfo contains the additional fields for BrokerNodeGroupInfo.
type CustomBrokerNodeGroupInfo struct {
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane/provider-aws/apis/ec2/v1beta1.Subnet
	// +crossplane:generate:reference:refFieldName=ClientSubnetRefs
	// +crossplane:generate:reference:selectorFieldName=ClientSubnetSelector
	ClientSubnets []*string `json:"clientSubnets,omitempty"`

	// ClientSubnetRefs is a list of references to Subnets used to set
	// the ClientSubnets.
	// +optional
	ClientSubnetRefs []xpv1.Reference `json:"clientSubnetRefs,omitempty"`

	// ClientSubnetSelector selects references to Subnets used
	// to set the ClientSubnets.
	// +optional
	ClientSubnetSelector *xpv1.Selector `json:"clientSubnetSelector,omitempty"`

	InstanceType *string `json:"instanceType,omitempty"`

	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane/provider-aws/apis/ec2/v1beta1.SecurityGroup
	// +crossplane:generate:reference:refFieldName=SecurityGroupRefs
	// +crossplane:generate:reference:selectorFieldName=SecurityGroupSelector
	SecurityGroups []*string `json:"securityGroups,omitempty"`

	// SecurityGroupRefs is a list of references to SecurityGroups used to set
	// the SecurityGroups.
	// +optional
	SecurityGroupRefs []xpv1.Reference `json:"securityGroupRefs,omitempty"`

	// SecurityGroupSelector selects references to SecurityGroup used
	// to set the SecurityGroups.
	// +optional
	SecurityGroupSelector *xpv1.Selector `json:"securityGroupSelector,omitempty"`

	// Contains information about storage volumes attached to MSK broker nodes.
	StorageInfo *StorageInfo `json:"storageInfo,omitempty"`
}
