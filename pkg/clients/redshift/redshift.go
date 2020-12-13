/*
Copyright 2020 The Crossplane Authors.

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

package redshift

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"

	"github.com/crossplane/provider-aws/apis/redshift/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// Client defines Redshift client operations
type Client interface {
	DescribeClustersRequest(input *redshift.DescribeClustersInput) redshift.DescribeClustersRequest
	CreateClusterRequest(input *redshift.CreateClusterInput) redshift.CreateClusterRequest
	ModifyClusterRequest(input *redshift.ModifyClusterInput) redshift.ModifyClusterRequest
	DeleteClusterRequest(input *redshift.DeleteClusterInput) redshift.DeleteClusterRequest
}

// NewClient creates new Redshift Client with provided AWS Configurations/Credentials
func NewClient(cfg aws.Config) Client {
	return redshift.New(cfg)
}

// LateInitialize fills the empty fields in *v1alpha1.ClusterParameters with
// the values seen in redshift.Cluster.
func LateInitialize(in *v1alpha1.ClusterParameters, cl *redshift.Cluster) { // nolint:gocyclo
	if cl == nil {
		return
	}
	in.AllowVersionUpgrade = awsclients.LateInitializeBoolPtr(in.AllowVersionUpgrade, cl.AllowVersionUpgrade)
	in.AutomatedSnapshotRetentionPeriod = awsclients.LateInitializeInt64Ptr(in.AutomatedSnapshotRetentionPeriod, cl.AutomatedSnapshotRetentionPeriod)
	in.AvailabilityZone = awsclients.LateInitializeStringPtr(in.AvailabilityZone, cl.AvailabilityZone)
	in.ClusterVersion = awsclients.LateInitializeStringPtr(in.ClusterVersion, cl.ClusterVersion)
	in.ClusterSubnetGroupName = awsclients.LateInitializeStringPtr(in.ClusterSubnetGroupName, cl.ClusterSubnetGroupName)
	in.DBName = awsclients.LateInitializeStringPtr(in.DBName, cl.DBName)
	in.Encrypted = awsclients.LateInitializeBoolPtr(in.Encrypted, cl.Encrypted)
	in.EnhancedVPCRouting = awsclients.LateInitializeBoolPtr(in.EnhancedVPCRouting, cl.EnhancedVpcRouting)
	in.KMSKeyID = awsclients.LateInitializeStringPtr(in.KMSKeyID, cl.KmsKeyId)
	in.MaintenanceTrackName = awsclients.LateInitializeStringPtr(in.MaintenanceTrackName, cl.MaintenanceTrackName)
	in.ManualSnapshotRetentionPeriod = awsclients.LateInitializeInt64Ptr(in.ManualSnapshotRetentionPeriod, cl.ManualSnapshotRetentionPeriod)
	in.MasterUsername = awsclients.LateInitializeString(in.MasterUsername, cl.MasterUsername)
	in.NodeType = awsclients.LateInitializeString(in.NodeType, cl.NodeType)
	in.NumberOfNodes = awsclients.LateInitializeInt64Ptr(in.NumberOfNodes, cl.NumberOfNodes)
	in.PreferredMaintenanceWindow = awsclients.LateInitializeStringPtr(in.PreferredMaintenanceWindow, cl.PreferredMaintenanceWindow)
	in.PubliclyAccessible = awsclients.LateInitializeBoolPtr(in.PubliclyAccessible, cl.PubliclyAccessible)
	in.SnapshotScheduleIdentifier = awsclients.LateInitializeStringPtr(in.SnapshotScheduleIdentifier, cl.SnapshotScheduleIdentifier)

	// If ClusterType is not provided by the user then set it to it's default value.
	// As redshift.Cluster type doesn't hold this info.
	if in.ClusterType == nil {
		if aws.Int64Value(cl.NumberOfNodes) > 1 {
			in.ClusterType = aws.String("multi-node")
		}
		if aws.Int64Value(cl.NumberOfNodes) == 1 {
			in.ClusterType = aws.String("single-node")
		}
	}
	if cl.Endpoint != nil {
		in.Port = awsclients.LateInitializeInt64Ptr(in.Port, cl.Endpoint.Port)
	}
	if cl.HsmStatus != nil {
		in.HSMClientCertificateIdentifier = awsclients.LateInitializeStringPtr(in.HSMClientCertificateIdentifier, cl.HsmStatus.HsmClientCertificateIdentifier)
		in.HSMConfigurationIdentifier = awsclients.LateInitializeStringPtr(in.HSMConfigurationIdentifier, cl.HsmStatus.HsmConfigurationIdentifier)
	}
	if cl.ElasticIpStatus != nil {
		in.ElasticIP = awsclients.LateInitializeStringPtr(in.ElasticIP, cl.ElasticIpStatus.ElasticIp)
	}

	if len(cl.ClusterSecurityGroups) != 0 {
		s := make([]string, len(cl.ClusterSecurityGroups))
		for i, v := range cl.ClusterSecurityGroups {
			s[i] = aws.StringValue(v.ClusterSecurityGroupName)
		}
		in.ClusterSecurityGroups = s
	}
	if len(cl.IamRoles) != 0 {
		s := make([]string, len(cl.IamRoles))
		for i, v := range cl.IamRoles {
			s[i] = aws.StringValue(v.IamRoleArn)
		}
		in.IAMRoles = s
	}
	if len(cl.Tags) != 0 {
		s := make([]v1alpha1.Tag, len(cl.Tags))
		for i, v := range cl.Tags {
			s[i] = v1alpha1.Tag{Key: aws.StringValue(v.Key), Value: aws.StringValue(v.Value)}
		}
		in.Tags = s
	}
	if len(cl.VpcSecurityGroups) != 0 {
		s := make([]string, len(cl.VpcSecurityGroups))
		for i, v := range cl.VpcSecurityGroups {
			s[i] = aws.StringValue(v.VpcSecurityGroupId)
		}
		in.VPCSecurityGroupIDs = s
	}
}

// IsUpToDate checks whether there is a change in any of the modifiable fields.
func IsUpToDate(p v1alpha1.ClusterParameters, cl redshift.Cluster) (bool, error) { // nolint:gocyclo
	// We need to check it explicitly as redshift.Cluster can have multiple ClusterParameterGroups
	found := isClusterParameterGroupNameUpdated(p.ClusterParameterGroupName, cl.ClusterParameterGroups)

	// Check if it is a cluster rename request
	if p.NewClusterIdentifier != nil && (aws.StringValue(p.NewClusterIdentifier) != aws.StringValue(cl.ClusterIdentifier)) {
		return false, nil
	}

	// Since redshift.Cluster doesn't have a ClusterType field therefore determine its value based upon number of nodes.
	if aws.Int64Value(cl.NumberOfNodes) > 1 && aws.StringValue(p.ClusterType) != "multi-node" {
		return false, nil
	}
	if aws.Int64Value(cl.NumberOfNodes) == 1 && aws.StringValue(p.ClusterType) != "single-node" {
		return false, nil
	}

	patch, err := CreatePatch(&p, &cl)
	if err != nil {
		return false, err
	}
	updated := cmp.Equal(&v1alpha1.ClusterParameters{}, patch,
		cmpopts.IgnoreTypes(&xpv1.Reference{}, &xpv1.Selector{}),
		cmpopts.IgnoreFields(v1alpha1.ClusterParameters{}, "Region"))
	return updated && found, nil
}

// initializeModifyandDeleteParameters fills the four v1alpha1.ClusterParameters
// fields that aren't available in redshift.Cluster and are for Modify or Delete input.
func initializeModifyandDeleteParameters(orig *v1alpha1.ClusterParameters, new *v1alpha1.ClusterParameters) *v1alpha1.ClusterParameters {
	new.FinalClusterSnapshotIdentifier = orig.FinalClusterSnapshotIdentifier
	new.FinalClusterSnapshotRetentionPeriod = orig.FinalClusterSnapshotRetentionPeriod
	new.NewClusterIdentifier = orig.NewClusterIdentifier
	new.SkipFinalClusterSnapshot = orig.SkipFinalClusterSnapshot
	return new
}

// CreatePatch creates a *v1alpha1.ClusterParameters that has only the changed
// values between the target *v1alpha1.ClusterParameters and the current
// *redshift.Cluster
func CreatePatch(target *v1alpha1.ClusterParameters, in *redshift.Cluster) (*v1alpha1.ClusterParameters, error) {
	currentParams := &v1alpha1.ClusterParameters{}
	initializeModifyandDeleteParameters(target, currentParams)
	LateInitialize(currentParams, in)

	jsonPatch, err := awsclients.CreateJSONPatch(currentParams, target)
	if err != nil {
		return nil, err
	}
	patch := &v1alpha1.ClusterParameters{}
	if err := json.Unmarshal(jsonPatch, patch); err != nil {
		return nil, err
	}
	return patch, nil
}

// IsNotFound helper function to test for ErrCodeClusterNotFoundFault error
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), redshift.ErrCodeClusterNotFoundFault)
}

// GenerateCreateClusterInput from RedshiftSpec
func GenerateCreateClusterInput(p *v1alpha1.ClusterParameters, cid, pw *string) *redshift.CreateClusterInput {
	var tags []redshift.Tag

	if len(p.Tags) != 0 {
		tags = make([]redshift.Tag, len(p.Tags))
		for i, val := range p.Tags {
			tags[i] = redshift.Tag{Key: aws.String(val.Key), Value: aws.String(val.Value)}
		}
	}

	return &redshift.CreateClusterInput{
		AllowVersionUpgrade:              p.AllowVersionUpgrade,
		AutomatedSnapshotRetentionPeriod: p.AutomatedSnapshotRetentionPeriod,
		AvailabilityZone:                 p.AvailabilityZone,
		ClusterIdentifier:                cid,
		ClusterParameterGroupName:        p.ClusterParameterGroupName,
		ClusterSecurityGroups:            p.ClusterSecurityGroups,
		ClusterSubnetGroupName:           p.ClusterSubnetGroupName,
		ClusterType:                      p.ClusterType,
		ClusterVersion:                   p.ClusterVersion,
		DBName:                           p.DBName,
		ElasticIp:                        p.ElasticIP,
		Encrypted:                        p.Encrypted,
		EnhancedVpcRouting:               p.EnhancedVPCRouting,
		HsmClientCertificateIdentifier:   p.HSMClientCertificateIdentifier,
		HsmConfigurationIdentifier:       p.HSMConfigurationIdentifier,
		IamRoles:                         p.IAMRoles,
		KmsKeyId:                         p.KMSKeyID,
		MaintenanceTrackName:             p.MaintenanceTrackName,
		ManualSnapshotRetentionPeriod:    p.ManualSnapshotRetentionPeriod,
		MasterUserPassword:               pw,
		MasterUsername:                   &p.MasterUsername,
		NodeType:                         &p.NodeType,
		NumberOfNodes:                    p.NumberOfNodes,
		Port:                             p.Port,
		PreferredMaintenanceWindow:       p.PreferredMaintenanceWindow,
		PubliclyAccessible:               p.PubliclyAccessible,
		SnapshotScheduleIdentifier:       p.SnapshotScheduleIdentifier,
		Tags:                             tags,
		VpcSecurityGroupIds:              p.VPCSecurityGroupIDs,
	}
}

// GenerateModifyClusterInput from RedshiftSpec
func GenerateModifyClusterInput(p *v1alpha1.ClusterParameters, cl redshift.Cluster) *redshift.ModifyClusterInput { //nolint:gocyclo
	patch, err := CreatePatch(p, &cl)
	if err != nil {
		return &redshift.ModifyClusterInput{}
	}

	o := &redshift.ModifyClusterInput{ClusterIdentifier: cl.ClusterIdentifier}
	if patch.AllowVersionUpgrade != nil {
		o.AllowVersionUpgrade = p.AllowVersionUpgrade
	}
	if patch.AutomatedSnapshotRetentionPeriod != nil {
		o.AutomatedSnapshotRetentionPeriod = p.AutomatedSnapshotRetentionPeriod
	}
	// If the cluster type, node type, or number of nodes changed, then the AWS API expects all three
	// items to be sent over
	// When a resize operation is requested, no other modifications are allowed in the same request
	if patch.ClusterType != nil || patch.NodeType != "" || patch.NumberOfNodes != nil {
		o.ClusterType = p.ClusterType
		o.NodeType = &p.NodeType
		o.NumberOfNodes = p.NumberOfNodes
		return o
	}
	if patch.ClusterVersion != nil {
		o.ClusterVersion = p.ClusterVersion
	}
	// The request to modify publicly accessible options for the cluster cannot be made in the same request as other modifications
	if patch.ElasticIP != nil || patch.PubliclyAccessible != nil {
		o.ElasticIp = p.ElasticIP
		o.PubliclyAccessible = p.PubliclyAccessible
		return o
	}
	if patch.Encrypted != nil {
		o.Encrypted = p.Encrypted
	}
	// Enhanced VPC routing changes must be made in a different request than other changes to the cluster
	if patch.EnhancedVPCRouting != nil {
		o.EnhancedVpcRouting = p.EnhancedVPCRouting
		return o
	}
	if patch.HSMClientCertificateIdentifier != nil {
		o.HsmClientCertificateIdentifier = p.HSMClientCertificateIdentifier
	}
	if patch.HSMConfigurationIdentifier != nil {
		o.HsmConfigurationIdentifier = p.HSMConfigurationIdentifier
	}
	if patch.KMSKeyID != nil {
		o.KmsKeyId = p.KMSKeyID
	}
	if patch.MaintenanceTrackName != nil {
		o.MaintenanceTrackName = p.MaintenanceTrackName
	}
	if patch.ManualSnapshotRetentionPeriod != nil {
		o.ManualSnapshotRetentionPeriod = p.ManualSnapshotRetentionPeriod
	}
	if patch.NewMasterUserPassword != nil {
		o.MasterUserPassword = p.NewMasterUserPassword
	}
	// When a rename operation is requested, no other modifications are allowed in the same request
	if aws.StringValue(p.NewClusterIdentifier) != aws.StringValue(cl.ClusterIdentifier) {
		o.NewClusterIdentifier = p.NewClusterIdentifier
		return o
	}
	if patch.PreferredMaintenanceWindow != nil {
		o.PreferredMaintenanceWindow = p.PreferredMaintenanceWindow
	}
	if len(patch.ClusterSecurityGroups) != 0 {
		o.ClusterSecurityGroups = p.ClusterSecurityGroups
	}
	if len(patch.VPCSecurityGroupIDs) != 0 {
		o.VpcSecurityGroupIds = p.VPCSecurityGroupIDs
	}

	updated := isClusterParameterGroupNameUpdated(p.ClusterParameterGroupName, cl.ClusterParameterGroups)
	if !updated {
		o.ClusterParameterGroupName = p.ClusterParameterGroupName
	}

	return o
}

// GenerateDeleteClusterInput from RedshiftSpec
func GenerateDeleteClusterInput(p *v1alpha1.ClusterParameters, cid *string) *redshift.DeleteClusterInput {
	return &redshift.DeleteClusterInput{
		ClusterIdentifier:                   cid,
		FinalClusterSnapshotIdentifier:      p.FinalClusterSnapshotIdentifier,
		FinalClusterSnapshotRetentionPeriod: p.FinalClusterSnapshotRetentionPeriod,
		SkipFinalClusterSnapshot:            p.SkipFinalClusterSnapshot,
	}
}

// GenerateObservation is used to produce v1alpha1.ClusterObservation from
// redshift.Cluster.
func GenerateObservation(in redshift.Cluster) v1alpha1.ClusterObservation { // nolint:gocyclo
	o := v1alpha1.ClusterObservation{
		ClusterPublicKey:                       aws.StringValue(in.ClusterPublicKey),
		ClusterRevisionNumber:                  aws.StringValue(in.ClusterRevisionNumber),
		ClusterStatus:                          aws.StringValue(in.ClusterStatus),
		ClusterAvailabilityStatus:              aws.StringValue(in.ClusterAvailabilityStatus),
		ElasticResizeNumberOfNodeOptions:       aws.StringValue(in.ElasticResizeNumberOfNodeOptions),
		ExpectedNextSnapshotScheduleTimeStatus: aws.StringValue(in.ExpectedNextSnapshotScheduleTimeStatus),
		ModifyStatus:                           aws.StringValue(in.ModifyStatus),
		SnapshotScheduleState:                  string(in.SnapshotScheduleState),
		VPCID:                                  aws.StringValue(in.VpcId),
	}

	if in.ClusterCreateTime != nil {
		t := metav1.NewTime(*in.ClusterCreateTime)
		o.ClusterCreateTime = &t
	}
	if in.ExpectedNextSnapshotScheduleTime != nil {
		t := metav1.NewTime(*in.ExpectedNextSnapshotScheduleTime)
		o.ExpectedNextSnapshotScheduleTime = &t
	}
	if in.NextMaintenanceWindowStartTime != nil {
		t := metav1.NewTime(*in.NextMaintenanceWindowStartTime)
		o.NextMaintenanceWindowStartTime = &t
	}

	if len(in.ClusterNodes) != 0 {
		s := make([]v1alpha1.ClusterNode, len(in.ClusterNodes))
		for i, v := range in.ClusterNodes {
			s[i] = v1alpha1.ClusterNode{
				NodeRole:         aws.StringValue(v.NodeRole),
				PrivateIPAddress: aws.StringValue(v.PrivateIPAddress),
				PublicIPAddress:  aws.StringValue(v.PublicIPAddress),
			}
		}
		o.ClusterNodes = s
	}
	if len(in.ClusterParameterGroups) != 0 {
		s := make([]v1alpha1.ClusterParameterGroupStatus, len(in.ClusterParameterGroups))
		for i, v := range in.ClusterParameterGroups {
			s[i] = v1alpha1.ClusterParameterGroupStatus{
				ParameterApplyStatus: aws.StringValue(v.ParameterApplyStatus),
				ParameterGroupName:   aws.StringValue(v.ParameterGroupName),
			}
			if len(v.ClusterParameterStatusList) != 0 {
				cps := make([]v1alpha1.ClusterParameterStatus, len(v.ClusterParameterStatusList))
				for j, k := range v.ClusterParameterStatusList {
					cps[j] = v1alpha1.ClusterParameterStatus{
						ParameterApplyErrorDescription: aws.StringValue(k.ParameterApplyErrorDescription),
						ParameterApplyStatus:           aws.StringValue(k.ParameterApplyStatus),
						ParameterName:                  aws.StringValue(k.ParameterName),
					}
				}
				s[i].ClusterParameterStatusList = cps
			}
		}
		o.ClusterParameterGroups = s
	}
	if len(in.DeferredMaintenanceWindows) != 0 {
		s := make([]v1alpha1.DeferredMaintenanceWindow, len(in.DeferredMaintenanceWindows))
		for i, v := range in.DeferredMaintenanceWindows {
			s[i] = v1alpha1.DeferredMaintenanceWindow{
				DeferMaintenanceIdentifier: aws.StringValue(v.DeferMaintenanceIdentifier),
			}
			if v.DeferMaintenanceStartTime != nil {
				t := metav1.NewTime(*v.DeferMaintenanceStartTime)
				s[i].DeferMaintenanceStartTime = &t
			}
			if v.DeferMaintenanceEndTime != nil {
				t := metav1.NewTime(*v.DeferMaintenanceEndTime)
				s[i].DeferMaintenanceEndTime = &t
			}
		}
		o.DeferredMaintenanceWindows = s
	}
	if len(in.PendingActions) != 0 {
		copy(o.PendingActions, in.PendingActions)
	}

	if in.ClusterSnapshotCopyStatus != nil {
		o.ClusterSnapshotCopyStatus = v1alpha1.ClusterSnapshotCopyStatus{
			DestinationRegion:             aws.StringValue(in.ClusterSnapshotCopyStatus.DestinationRegion),
			ManualSnapshotRetentionPeriod: aws.Int64Value(in.ClusterSnapshotCopyStatus.ManualSnapshotRetentionPeriod),
			RetentionPeriod:               aws.Int64Value(in.ClusterSnapshotCopyStatus.RetentionPeriod),
			SnapshotCopyGrantName:         aws.StringValue(in.ClusterSnapshotCopyStatus.SnapshotCopyGrantName),
		}
	}
	if in.DataTransferProgress != nil {
		o.DataTransferProgress = v1alpha1.DataTransferProgress{
			CurrentRateInMegaBytesPerSecond:    int(aws.Float64Value(in.DataTransferProgress.CurrentRateInMegaBytesPerSecond)),
			DataTransferredInMegaBytes:         aws.Int64Value(in.DataTransferProgress.DataTransferredInMegaBytes),
			ElapsedTimeInSeconds:               aws.Int64Value(in.DataTransferProgress.ElapsedTimeInSeconds),
			EstimatedTimeToCompletionInSeconds: aws.Int64Value(in.DataTransferProgress.EstimatedTimeToCompletionInSeconds),
			Status:                             aws.StringValue(in.DataTransferProgress.Status),
			TotalDataInMegaBytes:               aws.Int64Value(in.DataTransferProgress.TotalDataInMegaBytes),
		}
	}
	if in.ElasticIpStatus != nil {
		o.ElasticIPStatus = v1alpha1.ElasticIPStatus{
			ElasticIP: aws.StringValue(in.ElasticIpStatus.ElasticIp),
			Status:    aws.StringValue(in.ElasticIpStatus.Status),
		}
	}
	if in.Endpoint != nil {
		o.Endpoint = v1alpha1.Endpoint{
			Address: aws.StringValue(in.Endpoint.Address),
			Port:    aws.Int64Value(in.Endpoint.Port),
		}
	}
	if in.HsmStatus != nil {
		o.HSMStatus = v1alpha1.HSMStatus{
			HSMClientCertificateIdentifier: aws.StringValue(in.HsmStatus.HsmClientCertificateIdentifier),
			HSMConfigurationIdentifier:     aws.StringValue(in.HsmStatus.HsmConfigurationIdentifier),
			Status:                         aws.StringValue(in.HsmStatus.Status),
		}
	}

	return o
}

// GetConnectionDetails extracts managed.ConnectionDetails out of v1alpha1.Cluster.
func GetConnectionDetails(in v1alpha1.Cluster) managed.ConnectionDetails {
	if in.Status.AtProvider.Endpoint.Address == "" {
		return nil
	}
	return managed.ConnectionDetails{
		xpv1.ResourceCredentialsSecretEndpointKey: []byte(in.Status.AtProvider.Endpoint.Address),
		xpv1.ResourceCredentialsSecretPortKey:     []byte(strconv.Itoa(int(in.Status.AtProvider.Endpoint.Port))),
	}
}

// isClusterParameterGroupNameUpdated check if ClusterParameterGroupName is updated or not.
func isClusterParameterGroupNameUpdated(name *string, status []redshift.ClusterParameterGroupStatus) bool {
	var updated = true
	if name != nil {
		for _, v := range status {
			if aws.StringValue(name) != aws.StringValue(v.ParameterGroupName) {
				updated = false
			}
		}
	}
	return updated
}
