package v1alpha1

import xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

// CustomClusterParameters includes the custom fields of Cluster.
type CustomClusterParameters struct {
	// IAMRoleARN contains the ARN of an IAMRole
	// +immutable
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1.RoleARN()
	IAMRoleARN *string `json:"iamRoleARN,omitempty"`

	// IAMRoleARNRef is a reference to an IAMRole used to set
	// the IAMRoleARN.
	// +optional
	IAMRoleARNRef *xpv1.Reference `json:"iamRoleARNRef,omitempty"`

	// IAMRoleARNRefSelector selects references to IAMRole used
	// to set the IAMRoleARN.
	// +optional
	IAMRoleARNSelector *xpv1.Selector `json:"iamRoleARNSelector,omitempty"`

	// ParameterGroupName contains the name of the ParameterGroup
	// +immutable
	// +crossplane:generate:reference:type=ParameterGroup
	ParameterGroupName *string `json:"parameterGroupName,omitempty"`

	// ParameterGroupNameRef is a reference to an ParameterGroup used to set
	// the ParameterGroupName.
	// +optional
	ParameterGroupNameRef *xpv1.Reference `json:"parameterGroupNameRef,omitempty"`

	// ParameterGroupNameSelector selects references to ParameterGroup used
	// to set the ParameterGroupName.
	// +optional
	ParameterGroupNameSelector *xpv1.Selector `json:"parameterGroupNameSelector,omitempty"`

	// SubnetGroupName contains the name of the SubnetGroup
	// +immutable
	// +crossplane:generate:reference:type=SubnetGroup
	SubnetGroupName *string `json:"subnetGroupName,omitempty"`

	// SubnetGroupNameRef is a reference to an SubnetGroup used to set
	// the SubnetGroupName.
	// +optional
	SubnetGroupNameRef *xpv1.Reference `json:"subnetGroupNameRef,omitempty"`

	// SubnetGroupNameSelector selects references to SubnetGroup used
	// to set the SubnetGroupName.
	// +optional
	SubnetGroupNameSelector *xpv1.Selector `json:"subnetGroupNameSelector,omitempty"`

	// SecurityGroupIDs is the list of IDs for the SecurityGroups
	// +immutable
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1.SecurityGroup
	// +crossplane:generate:reference:refFieldName=SecurityGroupIDRefs
	// +crossplane:generate:reference:selectorFieldName=SecurityGroupIDSelector
	SecurityGroupIDs []*string `json:"securityGroupIds,omitempty"`

	// SecurityGroupIDRefs is a list of references to SecurityGroups used to set
	// the SecurityGroupIDs.
	// +optional
	SecurityGroupIDRefs []xpv1.Reference `json:"securityGroupIdRefs,omitempty"`

	// SecurityGroupIDsSelector selects references to SecurityGroupID used
	// to set the SecurityGroupIDs.
	// +optional
	SecurityGroupIDSelector *xpv1.Selector `json:"securityGroupIdSelector,omitempty"`

	// The Amazon Resource Name (ARN) of the Amazon SNS topic to which notifications
	// will be sent.
	//
	// The Amazon SNS topic owner must be same as the DAX cluster owner.
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/sns/v1beta1.Topic
	// +crossplane:generate:reference:extractor=github.com/crossplane-contrib/provider-aws/apis/sns/v1beta1.SNSTopicARN()
	NotificationTopicARN *string `json:"notificationTopicARN,omitempty"`

	// NotificationTopicARNRef references an SNS Topic to retrieve its NotificationTopicARN
	// +optional
	NotificationTopicARNRef *xpv1.Reference `json:"notificationTopicArnRef,omitempty"`

	// NotificationTopicARNSelector selects a reference to an SNS Topic to retrieve its NotificationTopicARN
	// +optional
	NotificationTopicARNSelector *xpv1.Selector `json:"notificationTopicArnSelector,omitempty"`
}

// CustomClusterParameters includes the custom status fields of Cluster.
type CustomClusterObservation struct{}

// CustomParameterGroupParameters includes the custom fields of ParameterGroup
type CustomParameterGroupParameters struct {
	// An array of name-value pairs for the parameters in the group. Each element
	// in the array represents a single parameter.
	//
	// record-ttl-millis and query-ttl-millis are the only supported parameter names.
	// For more details, see Configuring TTL Settings (https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/DAX.cluster-management.html#DAX.cluster-management.custom-settings.ttl).
	ParameterNameValues []*ParameterNameValue `json:"parameterNameValues,omitempty"`
}

// CustomClusterParameters includes the custom status fields of ParameterGroup.
type CustomParameterGroupObservation struct{}

// CustomSubnetGroupParameters includes the custom fields of SubnetGroup
type CustomSubnetGroupParameters struct {
	// SubnetIds is the list of Ids for the Subnets.
	// +crossplane:generate:reference:refFieldName=SubnetIDRefs
	// +crossplane:generate:reference:selectorFieldName=SubnetIDSelector
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1.Subnet
	SubnetIds []*string `json:"subnetIds,omitempty"`

	// SubnetIDRefs is a list of references to Subnets used to set
	// the SubnetIDs.
	// +optional
	SubnetIDRefs []xpv1.Reference `json:"subnetIDRefs,omitempty"`

	// SubnetIDSelector selects references to Subnets used
	// to set the SubnetIds.
	// +optional
	SubnetIDSelector *xpv1.Selector `json:"subnetIDSelector,omitempty"`
}

// CustomSubnetGroupObservation includes the custom status fields of SubnetGroup.
type CustomSubnetGroupObservation struct{}
