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
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	redshifttypes "github.com/aws/aws-sdk-go-v2/service/redshift/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane-contrib/provider-aws/apis/redshift/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/jsonpatch"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

// Client defines Redshift client operations
type Client interface {
	DescribeClusters(ctx context.Context, input *redshift.DescribeClustersInput, opts ...func(*redshift.Options)) (*redshift.DescribeClustersOutput, error)
	CreateCluster(ctx context.Context, input *redshift.CreateClusterInput, opts ...func(*redshift.Options)) (*redshift.CreateClusterOutput, error)
	ModifyCluster(ctx context.Context, input *redshift.ModifyClusterInput, opts ...func(*redshift.Options)) (*redshift.ModifyClusterOutput, error)
	DeleteCluster(ctx context.Context, input *redshift.DeleteClusterInput, opts ...func(*redshift.Options)) (*redshift.DeleteClusterOutput, error)
}

// NewClient creates new Redshift Client with provided AWS Configurations/Credentials
func NewClient(cfg aws.Config) Client {
	return redshift.NewFromConfig(cfg)
}

// LateInitialize fills the empty fields in *v1alpha1.ClusterParameters with
// the values seen in redshift.Cluster.
func LateInitialize(in *v1alpha1.ClusterParameters, cl *redshifttypes.Cluster) { //nolint:gocyclo
	if cl == nil {
		return
	}
	in.AllowVersionUpgrade = pointer.LateInitialize(in.AllowVersionUpgrade, &cl.AllowVersionUpgrade)
	in.AutomatedSnapshotRetentionPeriod = pointer.LateInitialize(in.AutomatedSnapshotRetentionPeriod, &cl.AutomatedSnapshotRetentionPeriod)
	in.AvailabilityZone = pointer.LateInitialize(in.AvailabilityZone, cl.AvailabilityZone)
	in.ClusterVersion = pointer.LateInitialize(in.ClusterVersion, cl.ClusterVersion)
	in.ClusterSubnetGroupName = pointer.LateInitialize(in.ClusterSubnetGroupName, cl.ClusterSubnetGroupName)
	in.DBName = pointer.LateInitialize(in.DBName, cl.DBName)
	in.Encrypted = pointer.LateInitialize(in.Encrypted, &cl.Encrypted)
	in.EnhancedVPCRouting = pointer.LateInitialize(in.EnhancedVPCRouting, &cl.EnhancedVpcRouting)
	in.KMSKeyID = pointer.LateInitialize(in.KMSKeyID, cl.KmsKeyId)
	in.MaintenanceTrackName = pointer.LateInitialize(in.MaintenanceTrackName, cl.MaintenanceTrackName)
	in.ManualSnapshotRetentionPeriod = pointer.LateInitialize(in.ManualSnapshotRetentionPeriod, &cl.ManualSnapshotRetentionPeriod)
	in.MasterUsername = pointer.LateInitializeValueFromPtr(in.MasterUsername, cl.MasterUsername)
	in.NodeType = pointer.LateInitializeValueFromPtr(in.NodeType, cl.NodeType)
	in.NumberOfNodes = pointer.LateInitialize(in.NumberOfNodes, &cl.NumberOfNodes)
	in.PreferredMaintenanceWindow = pointer.LateInitialize(in.PreferredMaintenanceWindow, cl.PreferredMaintenanceWindow)
	in.PubliclyAccessible = pointer.LateInitialize(in.PubliclyAccessible, &cl.PubliclyAccessible)
	in.SnapshotScheduleIdentifier = pointer.LateInitialize(in.SnapshotScheduleIdentifier, cl.SnapshotScheduleIdentifier)

	// If ClusterType is not provided by the user then set it to it's default value.
	// As redshift.Cluster type doesn't hold this info.
	if in.ClusterType == nil {
		if cl.NumberOfNodes > 1 {
			in.ClusterType = aws.String("multi-node")
		}
		if cl.NumberOfNodes == 1 {
			in.ClusterType = aws.String("single-node")
		}
	}
	if cl.Endpoint != nil {
		in.Port = pointer.LateInitialize(in.Port, &cl.Endpoint.Port)
	}
	if cl.HsmStatus != nil {
		in.HSMClientCertificateIdentifier = pointer.LateInitialize(in.HSMClientCertificateIdentifier, cl.HsmStatus.HsmClientCertificateIdentifier)
		in.HSMConfigurationIdentifier = pointer.LateInitialize(in.HSMConfigurationIdentifier, cl.HsmStatus.HsmConfigurationIdentifier)
	}
	if cl.ElasticIpStatus != nil {
		in.ElasticIP = pointer.LateInitialize(in.ElasticIP, cl.ElasticIpStatus.ElasticIp)
	}

	if len(cl.ClusterSecurityGroups) != 0 {
		s := make([]string, len(cl.ClusterSecurityGroups))
		for i, v := range cl.ClusterSecurityGroups {
			s[i] = aws.ToString(v.ClusterSecurityGroupName)
		}
		in.ClusterSecurityGroups = s
	}
	if len(cl.IamRoles) != 0 {
		s := make([]string, len(cl.IamRoles))
		for i, v := range cl.IamRoles {
			s[i] = aws.ToString(v.IamRoleArn)
		}
		in.IAMRoles = s
	}
	if len(cl.Tags) != 0 {
		s := make([]v1alpha1.Tag, len(cl.Tags))
		for i, v := range cl.Tags {
			s[i] = v1alpha1.Tag{Key: aws.ToString(v.Key), Value: aws.ToString(v.Value)}
		}
		in.Tags = s
	}
	if len(cl.VpcSecurityGroups) != 0 {
		s := make([]string, len(cl.VpcSecurityGroups))
		for i, v := range cl.VpcSecurityGroups {
			s[i] = aws.ToString(v.VpcSecurityGroupId)
		}
		if in.VPCSecurityGroupIDs == nil {
			in.VPCSecurityGroupIDs = make([]string, len(cl.VpcSecurityGroups))
		}
		in.VPCSecurityGroupIDs = s
	}
}

// IsUpToDate checks whether there is a change in any of the modifiable fields.
func IsUpToDate(p v1alpha1.ClusterParameters, cl redshifttypes.Cluster) (bool, error) {
	// We need to check it explicitly as redshift.Cluster can have multiple ClusterParameterGroups
	found := isClusterParameterGroupNameUpdated(p.ClusterParameterGroupName, cl.ClusterParameterGroups)

	// Check if it is a cluster rename request
	if p.NewClusterIdentifier != nil && (aws.ToString(p.NewClusterIdentifier) != aws.ToString(cl.ClusterIdentifier)) {
		return false, nil
	}

	// Since redshift.Cluster doesn't have a ClusterType field therefore determine its value based upon number of nodes.
	if cl.NumberOfNodes > 1 && aws.ToString(p.ClusterType) != "multi-node" {
		return false, nil
	}
	if cl.NumberOfNodes == 1 && aws.ToString(p.ClusterType) != "single-node" {
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
func CreatePatch(target *v1alpha1.ClusterParameters, in *redshifttypes.Cluster) (*v1alpha1.ClusterParameters, error) {
	currentParams := &v1alpha1.ClusterParameters{}
	initializeModifyandDeleteParameters(target, currentParams)
	LateInitialize(currentParams, in)

	jsonPatch, err := jsonpatch.CreateJSONPatch(currentParams, target)
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
	var cnff *redshifttypes.ClusterNotFoundFault
	return errors.As(err, &cnff)
}

// GenerateCreateClusterInput from RedshiftSpec
func GenerateCreateClusterInput(p *v1alpha1.ClusterParameters, cid, pw *string) *redshift.CreateClusterInput {
	var tags []redshifttypes.Tag

	if len(p.Tags) != 0 {
		tags = make([]redshifttypes.Tag, len(p.Tags))
		for i, val := range p.Tags {
			tags[i] = redshifttypes.Tag{Key: aws.String(val.Key), Value: aws.String(val.Value)}
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
func GenerateModifyClusterInput(p *v1alpha1.ClusterParameters, cl redshifttypes.Cluster) *redshift.ModifyClusterInput { //nolint:gocyclo
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
	if aws.ToString(p.NewClusterIdentifier) != aws.ToString(cl.ClusterIdentifier) {
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
		SkipFinalClusterSnapshot:            aws.ToBool(p.SkipFinalClusterSnapshot),
	}
}

// GenerateObservation is used to produce v1alpha1.ClusterObservation from
// redshift.Cluster.
func GenerateObservation(in redshifttypes.Cluster) v1alpha1.ClusterObservation { //nolint:gocyclo
	o := v1alpha1.ClusterObservation{
		ClusterPublicKey:                       aws.ToString(in.ClusterPublicKey),
		ClusterRevisionNumber:                  aws.ToString(in.ClusterRevisionNumber),
		ClusterStatus:                          aws.ToString(in.ClusterStatus),
		ClusterAvailabilityStatus:              aws.ToString(in.ClusterAvailabilityStatus),
		ElasticResizeNumberOfNodeOptions:       aws.ToString(in.ElasticResizeNumberOfNodeOptions),
		ExpectedNextSnapshotScheduleTimeStatus: aws.ToString(in.ExpectedNextSnapshotScheduleTimeStatus),
		ModifyStatus:                           aws.ToString(in.ModifyStatus),
		SnapshotScheduleState:                  string(in.SnapshotScheduleState),
		VPCID:                                  aws.ToString(in.VpcId),
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
				NodeRole:         aws.ToString(v.NodeRole),
				PrivateIPAddress: aws.ToString(v.PrivateIPAddress),
				PublicIPAddress:  aws.ToString(v.PublicIPAddress),
			}
		}
		o.ClusterNodes = s
	}
	if len(in.ClusterParameterGroups) != 0 {
		s := make([]v1alpha1.ClusterParameterGroupStatus, len(in.ClusterParameterGroups))
		for i, v := range in.ClusterParameterGroups {
			s[i] = v1alpha1.ClusterParameterGroupStatus{
				ParameterApplyStatus: aws.ToString(v.ParameterApplyStatus),
				ParameterGroupName:   aws.ToString(v.ParameterGroupName),
			}
			if len(v.ClusterParameterStatusList) != 0 {
				cps := make([]v1alpha1.ClusterParameterStatus, len(v.ClusterParameterStatusList))
				for j, k := range v.ClusterParameterStatusList {
					cps[j] = v1alpha1.ClusterParameterStatus{
						ParameterApplyErrorDescription: aws.ToString(k.ParameterApplyErrorDescription),
						ParameterApplyStatus:           aws.ToString(k.ParameterApplyStatus),
						ParameterName:                  aws.ToString(k.ParameterName),
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
				DeferMaintenanceIdentifier: aws.ToString(v.DeferMaintenanceIdentifier),
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
			DestinationRegion:             aws.ToString(in.ClusterSnapshotCopyStatus.DestinationRegion),
			ManualSnapshotRetentionPeriod: in.ClusterSnapshotCopyStatus.ManualSnapshotRetentionPeriod,
			RetentionPeriod:               in.ClusterSnapshotCopyStatus.RetentionPeriod,
			SnapshotCopyGrantName:         aws.ToString(in.ClusterSnapshotCopyStatus.SnapshotCopyGrantName),
		}
	}
	if in.DataTransferProgress != nil {
		o.DataTransferProgress = v1alpha1.DataTransferProgress{
			CurrentRateInMegaBytesPerSecond:    int(aws.ToFloat64(in.DataTransferProgress.CurrentRateInMegaBytesPerSecond)),
			DataTransferredInMegaBytes:         in.DataTransferProgress.DataTransferredInMegaBytes,
			ElapsedTimeInSeconds:               aws.ToInt64(in.DataTransferProgress.ElapsedTimeInSeconds),
			EstimatedTimeToCompletionInSeconds: aws.ToInt64(in.DataTransferProgress.EstimatedTimeToCompletionInSeconds),
			Status:                             aws.ToString(in.DataTransferProgress.Status),
			TotalDataInMegaBytes:               in.DataTransferProgress.TotalDataInMegaBytes,
		}
	}
	if in.ElasticIpStatus != nil {
		o.ElasticIPStatus = v1alpha1.ElasticIPStatus{
			ElasticIP: aws.ToString(in.ElasticIpStatus.ElasticIp),
			Status:    aws.ToString(in.ElasticIpStatus.Status),
		}
	}
	if in.Endpoint != nil {
		o.Endpoint = v1alpha1.Endpoint{
			Address: aws.ToString(in.Endpoint.Address),
			Port:    in.Endpoint.Port,
		}
	}
	if in.HsmStatus != nil {
		o.HSMStatus = v1alpha1.HSMStatus{
			HSMClientCertificateIdentifier: aws.ToString(in.HsmStatus.HsmClientCertificateIdentifier),
			HSMConfigurationIdentifier:     aws.ToString(in.HsmStatus.HsmConfigurationIdentifier),
			Status:                         aws.ToString(in.HsmStatus.Status),
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
func isClusterParameterGroupNameUpdated(name *string, status []redshifttypes.ClusterParameterGroupStatus) bool {
	var updated = true
	if name != nil {
		for _, v := range status {
			if aws.ToString(name) != aws.ToString(v.ParameterGroupName) {
				updated = false
			}
		}
	}
	return updated
}
