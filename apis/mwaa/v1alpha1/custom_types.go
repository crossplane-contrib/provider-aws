package v1alpha1

import xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

const (
	// ConnectionDetailsCLITokenKey for cli token
	ConnectionDetailsCLITokenKey string = "cliToken"
	// ConnectionDetailsWebTokenKey for web token
	ConnectionDetailsWebTokenKey string = "webToken"
	// ConnectionDetailsWebServerURL for web server URL
	ConnectionDetailsWebServerURL string = "webServerURL"
)

// CustomEnvironmentParameters for an Environment.
type CustomEnvironmentParameters struct {
	// The AWS Key Management Service (KMS) key to encrypt the data in your environment.
	// You can use an AWS owned CMK, or a Customer managed CMK (advanced). To learn
	// more, see Get started with Amazon Managed Workflows for Apache Airflow (https://docs.aws.amazon.com/mwaa/latest/userguide/get-started.html).
	//
	// This field or KMSKeyRef or KMSKeySelector is required.
	// +optional
	// +immutable
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/kms/v1alpha1.Key
	KMSKey *string `json:"kmsKey,omitempty"`

	// KMSKeyRef is a reference to the KMSKey used to set.
	// the SubnetIDs.
	// +optional
	KMSKeyRef *xpv1.Reference `json:"kmsKeyRef,omitempty"`

	// KMSKeySelector selects the reference to the KMSKey.
	// +optional
	KMSKeySelector *xpv1.Selector `json:"kmsKeySelector,omitempty"`

	// The Amazon Resource Name (ARN) of the Amazon S3 bucket where your DAG code
	// and supporting files are stored. For example, arn:aws:s3:::my-airflow-bucket-unique-name.
	// To learn more, see Create an Amazon S3 bucket for Amazon MWAA (https://docs.aws.amazon.com/mwaa/latest/userguide/mwaa-s3-bucket.html).
	//
	// This field or SourceBucketARNRef or SourceBucketARNSelector is required.
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/s3/v1beta1.Bucket
	// +crossplane:generate:reference:extractor=github.com/crossplane-contrib/provider-aws/apis/s3/v1beta1.BucketARN()
	SourceBucketARN *string `json:"sourceBucketARN"`

	// SourceBucketARNRef is a reference to the SourceBucketARN used to set.
	// the SubnetIDs.
	// +optional
	SourceBucketARNRef *xpv1.Reference `json:"sourceBucketARNRef,omitempty"`

	// SourceBucketARNSelector selects the reference to the SourceBucketARN.
	// +optional
	SourceBucketARNSelector *xpv1.Selector `json:"sourceBucketARNSelector,omitempty"`

	// The Amazon Resource Name (ARN) of the execution role for your environment.
	// An execution role is an AWS Identity and Access Management (IAM) role that
	// grants MWAA permission to access AWS services and resources used by your
	// environment. For example, arn:aws:iam::123456789:role/my-execution-role.
	// To learn more, see Amazon MWAA Execution role (https://docs.aws.amazon.com/mwaa/latest/userguide/mwaa-create-role.html).
	//
	// This field or SourceBucketARNRef or SourceBucketARNSelector is required.
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1.RoleARN()
	ExecutionRoleARN *string `json:"executionRoleARN"`

	// ExecutionRoleARNRef is a reference to the ExecutionRoleARN used to set.
	// the SubnetIDs.
	// +optional
	ExecutionRoleARNRef *xpv1.Reference `json:"executionRoleARNRef,omitempty"`

	// ExecutionRoleARNSelector selects the reference to the ExecutionRoleARN.
	// +optional
	ExecutionRoleARNSelector *xpv1.Selector `json:"executionRoleARNSelector,omitempty"`

	// The VPC networking components used to secure and enable network traffic between
	// the AWS resources for your environment. To learn more, see About networking
	// on Amazon MWAA (https://docs.aws.amazon.com/mwaa/latest/userguide/networking-about.html).
	// +kubebuilder:validation:Required
	NetworkConfiguration CustomNetworkConfiguration `json:"networkConfiguration"`
}

// CustomEnvironmentObservation includes the custom status fields of Environment.
type CustomEnvironmentObservation struct{}

// CustomNetworkConfiguration for an Environment.
type CustomNetworkConfiguration struct {
	// SecurityGroupIDs is the list of IDs for the SecurityGroups.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1.SecurityGroup
	// +crossplane:generate:reference:refFieldName=SecurityGroupIDRefs
	// +crossplane:generate:reference:selectorFieldName=SecurityGroupIDSelector
	SecurityGroupIDs []string `json:"securityGroupIds,omitempty"`

	// SecurityGroupIDRefs is a list of references to SecurityGroups used to set
	// the SecurityGroupIDs.
	// +optional
	SecurityGroupIDRefs []xpv1.Reference `json:"securityGroupIdRefs,omitempty"`

	// SecurityGroupIDsSelector selects references to SecurityGroupID used
	// to set the SecurityGroupIDs.
	// +optional
	SecurityGroupIDSelector *xpv1.Selector `json:"securityGroupIdSelector,omitempty"`

	// SecurityGroupIDs is the list of IDs for the SecurityGroups.
	// +immutable
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1.Subnet
	// +crossplane:generate:reference:refFieldName=SubnetIDRefs
	// +crossplane:generate:reference:selectorFieldName=SubnetIDSelector
	SubnetIDs []string `json:"subnetIds,omitempty"`

	// SubnetIDRefs is a list of references to Subnets used to set
	// the SubnetIDs.
	// +optional
	SubnetIDRefs []xpv1.Reference `json:"subnetIdRefs,omitempty"`

	// SubnetIDsSelector selects references to Subnets used
	// to set the SubnetIDs.
	// +optional
	SubnetIDSelector *xpv1.Selector `json:"subnetIdSelector,omitempty"`
}
