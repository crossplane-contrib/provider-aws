package domain

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/opensearchservice"
	svcsdkapi "github.com/aws/aws-sdk-go/service/opensearchservice/opensearchserviceiface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/opensearchservice/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	awsclients "github.com/crossplane-contrib/provider-aws/pkg/clients"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
)

// SetupDomain adds a controller that reconciles Domain.
func SetupDomain(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.DomainGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.preCreate = preCreate
			e.preDelete = preDelete
			e.postObserve = postObserve
			e.isUpToDate = isUpToDate
			e.lateInitialize = lateInitialize
			du := &updateDomain{client: e.client}
			e.update = du.update
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.Domain{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.DomainGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

func preObserve(_ context.Context, cr *svcapitypes.Domain, obj *svcsdk.DescribeDomainInput) error {
	obj.DomainName = aws.String(meta.GetExternalName(cr))
	return nil
}

func preCreate(_ context.Context, cr *svcapitypes.Domain, obj *svcsdk.CreateDomainInput) error {
	obj.DomainName = aws.String(meta.GetExternalName(cr))

	if cr.Spec.ForProvider.VPCOptions != nil {
		obj.VPCOptions = &svcsdk.VPCOptions{}
		if len(cr.Spec.ForProvider.VPCOptions.SubnetIDs) > 0 {
			obj.VPCOptions.SubnetIds = make([]*string, len(cr.Spec.ForProvider.VPCOptions.SubnetIDs))
			copy(obj.VPCOptions.SubnetIds, cr.Spec.ForProvider.VPCOptions.SubnetIDs)
		}

		if len(cr.Spec.ForProvider.VPCOptions.SecurityGroupIDs) > 0 {
			obj.VPCOptions.SecurityGroupIds = make([]*string, len(cr.Spec.ForProvider.VPCOptions.SecurityGroupIDs))
			copy(obj.VPCOptions.SecurityGroupIds, cr.Spec.ForProvider.VPCOptions.SecurityGroupIDs)
		}
	}

	if cr.Spec.ForProvider.EncryptionAtRestOptions != nil {
		obj.EncryptionAtRestOptions = &svcsdk.EncryptionAtRestOptions{
			Enabled:  cr.Spec.ForProvider.EncryptionAtRestOptions.Enabled,
			KmsKeyId: cr.Spec.ForProvider.EncryptionAtRestOptions.KMSKeyID,
		}
	}

	if cr.Spec.ForProvider.SnapshotOptions != nil {
		obj.SnapshotOptions = &svcsdk.SnapshotOptions{
			AutomatedSnapshotStartHour: obj.SnapshotOptions.AutomatedSnapshotStartHour,
		}
	}

	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.Domain, obj *svcsdk.DeleteDomainInput) (bool, error) {
	obj.DomainName = aws.String(meta.GetExternalName(cr))
	return false, nil
}

func postObserve(_ context.Context, cr *svcapitypes.Domain, resp *svcsdk.DescribeDomainOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	if resp.DomainStatus != nil {
		if *resp.DomainStatus.Deleted {
			cr.SetConditions(xpv1.Deleting())
		}
		if !*resp.DomainStatus.Created {
			cr.SetConditions(xpv1.Creating())
		} else {
			cr.SetConditions(xpv1.Available())
		}
	}

	cr.Status.AtProvider = GenerateObservation(resp.DomainStatus)
	obs.ConnectionDetails = managed.ConnectionDetails{
		"arn": []byte(awsclients.StringValue(resp.DomainStatus.ARN)),
	}

	// public endpoints
	if resp.DomainStatus.Endpoint != nil {
		obs.ConnectionDetails["endpoint"] = []byte(awsclients.StringValue(resp.DomainStatus.Endpoint))
	}

	// vpc endpoints
	if resp.DomainStatus.Endpoints != nil {
		var keys []string
		for key := range resp.DomainStatus.Endpoints {
			keys = append(keys, key)
		}
		for _, key := range keys {
			obs.ConnectionDetails[key+"Endpoint"] = []byte(awsclients.StringValue(resp.DomainStatus.Endpoints[key]))
		}
	}

	return obs, nil
}

// GenerateObservation is used to produce DomainObservation
func GenerateObservation(obj *svcsdk.DomainStatus) svcapitypes.DomainObservation { // nolint:gocyclo
	if obj == nil {
		return svcapitypes.DomainObservation{}
	}

	o := svcapitypes.DomainObservation{
		ARN:               obj.ARN,
		AccessPolicies:    obj.AccessPolicies,
		Created:           obj.Created,
		Deleted:           obj.Deleted,
		DomainID:          obj.DomainId,
		Endpoint:          obj.Endpoint,
		EngineVersion:     obj.EngineVersion,
		Processing:        obj.Processing,
		UpgradeProcessing: obj.UpgradeProcessing,
	}

	if obj.AdvancedOptions != nil {
		f1 := map[string]*string{}
		for f1key, f1valiter := range obj.AdvancedOptions {
			var f1val = *f1valiter
			f1[f1key] = &f1val
		}
		o.AdvancedOptions = f1
	}

	if obj.ChangeProgressDetails != nil {
		f2 := &svcapitypes.ChangeProgressDetails{}
		if obj.ChangeProgressDetails.ChangeId != nil {
			f2.ChangeID = obj.ChangeProgressDetails.ChangeId
		}
		if obj.ChangeProgressDetails.Message != nil {
			f2.Message = obj.ChangeProgressDetails.Message
		}
		o.ChangeProgressDetails = f2
	}

	if obj.AutoTuneOptions != nil {
		f3 := &svcapitypes.AutoTuneOptionsOutput{}
		if obj.AutoTuneOptions.State != nil {
			f3.State = obj.AutoTuneOptions.State
		}
		if obj.AutoTuneOptions.ErrorMessage != nil {
			f3.ErrorMessage = obj.AutoTuneOptions.ErrorMessage
		}
		o.AutoTuneOptions = f3
	}

	if obj.CognitoOptions != nil {
		f4 := &svcapitypes.CognitoOptions{}
		if obj.CognitoOptions.Enabled != nil {
			f4.Enabled = obj.CognitoOptions.Enabled
		}
		if obj.CognitoOptions.IdentityPoolId != nil {
			f4.IdentityPoolID = obj.CognitoOptions.IdentityPoolId
		}
		if obj.CognitoOptions.RoleArn != nil {
			f4.RoleARN = obj.CognitoOptions.RoleArn
		}
		if obj.CognitoOptions.UserPoolId != nil {
			f4.UserPoolID = obj.CognitoOptions.UserPoolId
		}
		o.CognitoOptions = f4
	}

	if obj.DomainEndpointOptions != nil {
		f5 := &svcapitypes.DomainEndpointOptions{}

		if obj.DomainEndpointOptions.CustomEndpoint != nil {
			f5.CustomEndpoint = obj.DomainEndpointOptions.CustomEndpoint
		}
		if obj.DomainEndpointOptions.CustomEndpointCertificateArn != nil {
			f5.CustomEndpointCertificateARN = obj.DomainEndpointOptions.CustomEndpointCertificateArn
		}
		if obj.DomainEndpointOptions.CustomEndpointEnabled != nil {
			f5.CustomEndpointEnabled = obj.DomainEndpointOptions.CustomEndpointEnabled
		}
		if obj.DomainEndpointOptions.EnforceHTTPS != nil {
			f5.EnforceHTTPS = obj.DomainEndpointOptions.EnforceHTTPS
		}
		if obj.DomainEndpointOptions.TLSSecurityPolicy != nil {
			f5.TLSSecurityPolicy = obj.DomainEndpointOptions.TLSSecurityPolicy
		}
		o.DomainEndpointOptions = f5
	}

	if obj.ClusterConfig != nil {
		f6 := &svcapitypes.ClusterConfig{}
		if obj.ClusterConfig.ColdStorageOptions != nil {
			f6f0 := &svcapitypes.ColdStorageOptions{}
			if obj.ClusterConfig.ColdStorageOptions.Enabled != nil {
				f6f0.Enabled = obj.ClusterConfig.ColdStorageOptions.Enabled
			}
			f6.ColdStorageOptions = f6f0
		}
		if obj.ClusterConfig.DedicatedMasterCount != nil {
			f6.DedicatedMasterCount = obj.ClusterConfig.DedicatedMasterCount
		}
		if obj.ClusterConfig.DedicatedMasterEnabled != nil {
			f6.DedicatedMasterEnabled = obj.ClusterConfig.DedicatedMasterEnabled
		}
		if obj.ClusterConfig.DedicatedMasterType != nil {
			f6.DedicatedMasterType = obj.ClusterConfig.DedicatedMasterType
		}
		if obj.ClusterConfig.InstanceCount != nil {
			f6.InstanceCount = obj.ClusterConfig.InstanceCount
		}
		if obj.ClusterConfig.InstanceType != nil {
			f6.InstanceType = obj.ClusterConfig.InstanceType
		}
		if obj.ClusterConfig.WarmCount != nil {
			f6.WarmCount = obj.ClusterConfig.WarmCount
		}
		if obj.ClusterConfig.WarmEnabled != nil {
			f6.WarmEnabled = obj.ClusterConfig.WarmEnabled
		}
		if obj.ClusterConfig.WarmType != nil {
			f6.WarmType = obj.ClusterConfig.WarmType
		}
		if obj.ClusterConfig.ZoneAwarenessConfig != nil {
			f6f9 := &svcapitypes.ZoneAwarenessConfig{}
			if obj.ClusterConfig.ZoneAwarenessConfig.AvailabilityZoneCount != nil {
				f6f9.AvailabilityZoneCount = obj.ClusterConfig.ZoneAwarenessConfig.AvailabilityZoneCount
			}
			f6.ZoneAwarenessConfig = f6f9
		}
		if obj.ClusterConfig.ZoneAwarenessEnabled != nil {
			f6.ZoneAwarenessEnabled = obj.ClusterConfig.ZoneAwarenessEnabled
		}
		o.ClusterConfig = f6
	}

	if obj.EncryptionAtRestOptions != nil {
		f7 := &svcapitypes.EncryptionAtRestOptions{}
		if obj.EncryptionAtRestOptions.Enabled != nil {
			f7.Enabled = obj.EncryptionAtRestOptions.Enabled
		}
		if obj.EncryptionAtRestOptions.KmsKeyId != nil {
			f7.KMSKeyID = obj.EncryptionAtRestOptions.KmsKeyId
		}
		o.EncryptionAtRestOptions = f7
	}

	if obj.NodeToNodeEncryptionOptions != nil {
		f8 := &svcapitypes.NodeToNodeEncryptionOptions{}
		if obj.NodeToNodeEncryptionOptions.Enabled != nil {
			f8.Enabled = obj.NodeToNodeEncryptionOptions.Enabled
		}
		o.NodeToNodeEncryptionOptions = f8
	}

	if obj.ServiceSoftwareOptions != nil {
		f9 := &svcapitypes.ServiceSoftwareOptions{}
		if obj.ServiceSoftwareOptions.AutomatedUpdateDate != nil {
			f9.AutomatedUpdateDate = &metav1.Time{Time: *obj.ServiceSoftwareOptions.AutomatedUpdateDate}
		}
		if obj.ServiceSoftwareOptions.Cancellable != nil {
			f9.Cancellable = obj.ServiceSoftwareOptions.Cancellable
		}
		if obj.ServiceSoftwareOptions.CurrentVersion != nil {
			f9.CurrentVersion = obj.ServiceSoftwareOptions.CurrentVersion
		}
		if obj.ServiceSoftwareOptions.Description != nil {
			f9.Description = obj.ServiceSoftwareOptions.Description
		}
		if obj.ServiceSoftwareOptions.NewVersion != nil {
			f9.NewVersion = obj.ServiceSoftwareOptions.NewVersion
		}
		if obj.ServiceSoftwareOptions.OptionalDeployment != nil {
			f9.OptionalDeployment = obj.ServiceSoftwareOptions.OptionalDeployment
		}
		if obj.ServiceSoftwareOptions.UpdateAvailable != nil {
			f9.UpdateAvailable = obj.ServiceSoftwareOptions.UpdateAvailable
		}
		if obj.ServiceSoftwareOptions.UpdateStatus != nil {
			f9.UpdateStatus = obj.ServiceSoftwareOptions.UpdateStatus
		}
		o.ServiceSoftwareOptions = f9
	}

	if obj.SnapshotOptions != nil {
		f10 := &svcapitypes.SnapshotOptions{}
		if obj.SnapshotOptions.AutomatedSnapshotStartHour != nil {
			f10.AutomatedSnapshotStartHour = obj.SnapshotOptions.AutomatedSnapshotStartHour
		}
		o.SnapshotOptions = f10
	}

	if obj.VPCOptions != nil {
		f11 := &svcapitypes.VPCDerivedInfo{}
		if obj.VPCOptions.AvailabilityZones != nil {
			f11.AvailabilityZones = obj.VPCOptions.AvailabilityZones
		}
		if obj.VPCOptions.SecurityGroupIds != nil {
			f11.SecurityGroupIDs = obj.VPCOptions.SecurityGroupIds
		}
		if obj.VPCOptions.SubnetIds != nil {
			f11.SubnetIDs = obj.VPCOptions.SubnetIds
		}
		if obj.VPCOptions.VPCId != nil {
			f11.VPCID = obj.VPCOptions.VPCId
		}
		o.VPCOptions = f11
	}

	return o
}

func isUpToDate(obj *svcapitypes.Domain, out *svcsdk.DescribeDomainOutput) (bool, error) { // nolint:gocyclo

	if aws.StringValue(obj.Spec.ForProvider.AccessPolicies) != aws.StringValue(out.DomainStatus.AccessPolicies) {
		return false, nil
	}
	if len(obj.Spec.ForProvider.AdvancedOptions) != len(out.DomainStatus.AdvancedOptions) {
		return false, nil
	}
	for key, value := range obj.Spec.ForProvider.AdvancedOptions {
		if aws.StringValue(out.DomainStatus.AdvancedOptions[key]) != aws.StringValue(value) {
			return false, nil
		}
	}

	if aws.BoolValue(obj.Spec.ForProvider.AdvancedSecurityOptions.AnonymousAuthEnabled) != aws.BoolValue(out.DomainStatus.AdvancedSecurityOptions.AnonymousAuthEnabled) {
		return false, nil
	}
	if aws.BoolValue(obj.Spec.ForProvider.AdvancedSecurityOptions.Enabled) != aws.BoolValue(out.DomainStatus.AdvancedSecurityOptions.Enabled) {
		return false, nil
	}
	if aws.BoolValue(obj.Spec.ForProvider.AdvancedSecurityOptions.InternalUserDatabaseEnabled) != aws.BoolValue(out.DomainStatus.AdvancedSecurityOptions.InternalUserDatabaseEnabled) {
		return false, nil
	}

	if out.DomainStatus.AdvancedSecurityOptions.SAMLOptions != nil {
		if obj.Spec.ForProvider.AdvancedSecurityOptions.SAMLOptions != nil {
			if aws.BoolValue(obj.Spec.ForProvider.AdvancedSecurityOptions.SAMLOptions.Enabled) != aws.BoolValue(out.DomainStatus.AdvancedSecurityOptions.SAMLOptions.Enabled) {
				return false, nil
			}
			if out.DomainStatus.AdvancedSecurityOptions.SAMLOptions.Idp != nil {
				if obj.Spec.ForProvider.AdvancedSecurityOptions.SAMLOptions.IDp != nil {
					if aws.StringValue(obj.Spec.ForProvider.AdvancedSecurityOptions.SAMLOptions.IDp.EntityID) != aws.StringValue(out.DomainStatus.AdvancedSecurityOptions.SAMLOptions.Idp.EntityId) {
						return false, nil
					}
					if aws.StringValue(obj.Spec.ForProvider.AdvancedSecurityOptions.SAMLOptions.IDp.MetadataContent) != aws.StringValue(out.DomainStatus.AdvancedSecurityOptions.SAMLOptions.Idp.MetadataContent) {
						return false, nil
					}
				} else {
					return false, nil
				}
			} else if obj.Spec.ForProvider.AdvancedSecurityOptions.SAMLOptions.IDp != nil {
				return false, nil
			}
			if aws.StringValue(obj.Spec.ForProvider.AdvancedSecurityOptions.SAMLOptions.RolesKey) != aws.StringValue(out.DomainStatus.AdvancedSecurityOptions.SAMLOptions.RolesKey) {
				return false, nil
			}
			if aws.Int64Value(obj.Spec.ForProvider.AdvancedSecurityOptions.SAMLOptions.SessionTimeoutMinutes) != aws.Int64Value(out.DomainStatus.AdvancedSecurityOptions.SAMLOptions.SessionTimeoutMinutes) {
				return false, nil
			}
			if aws.StringValue(obj.Spec.ForProvider.AdvancedSecurityOptions.SAMLOptions.SubjectKey) != aws.StringValue(out.DomainStatus.AdvancedSecurityOptions.SAMLOptions.SubjectKey) {
				return false, nil
			}
		} else {
			return false, nil
		}
	} else if obj.Spec.ForProvider.AdvancedSecurityOptions.SAMLOptions != nil {
		return false, nil
	}

	if out.DomainStatus.ClusterConfig != nil {
		if obj.Spec.ForProvider.ClusterConfig != nil {
			if out.DomainStatus.ClusterConfig.ColdStorageOptions != nil {
				if obj.Spec.ForProvider.ClusterConfig.ColdStorageOptions != nil {
					if aws.BoolValue(obj.Spec.ForProvider.ClusterConfig.ColdStorageOptions.Enabled) != aws.BoolValue(out.DomainStatus.ClusterConfig.ColdStorageOptions.Enabled) {
						return false, nil
					}
				}
			} else if obj.Spec.ForProvider.ClusterConfig.ColdStorageOptions != nil {
				return false, nil
			}
			if aws.Int64Value(obj.Spec.ForProvider.ClusterConfig.DedicatedMasterCount) != aws.Int64Value(out.DomainStatus.ClusterConfig.DedicatedMasterCount) {
				return false, nil
			}
			if aws.BoolValue(obj.Spec.ForProvider.ClusterConfig.DedicatedMasterEnabled) != aws.BoolValue(out.DomainStatus.ClusterConfig.DedicatedMasterEnabled) {
				return false, nil
			}
			if aws.StringValue(obj.Spec.ForProvider.ClusterConfig.DedicatedMasterType) != aws.StringValue(out.DomainStatus.ClusterConfig.DedicatedMasterType) {
				return false, nil
			}
			if aws.Int64Value(obj.Spec.ForProvider.ClusterConfig.InstanceCount) != aws.Int64Value(out.DomainStatus.ClusterConfig.InstanceCount) {
				return false, nil
			}
			if aws.StringValue(obj.Spec.ForProvider.ClusterConfig.InstanceType) != aws.StringValue(out.DomainStatus.ClusterConfig.InstanceType) {
				return false, nil
			}
			if aws.Int64Value(obj.Spec.ForProvider.ClusterConfig.WarmCount) != aws.Int64Value(out.DomainStatus.ClusterConfig.WarmCount) {
				return false, nil
			}
			if aws.BoolValue(obj.Spec.ForProvider.ClusterConfig.WarmEnabled) != aws.BoolValue(out.DomainStatus.ClusterConfig.WarmEnabled) {
				return false, nil
			}
			if aws.StringValue(obj.Spec.ForProvider.ClusterConfig.WarmType) != aws.StringValue(out.DomainStatus.ClusterConfig.WarmType) {
				return false, nil
			}

			if out.DomainStatus.ClusterConfig.ZoneAwarenessConfig != nil {
				if obj.Spec.ForProvider.ClusterConfig.ZoneAwarenessConfig != nil {
					if aws.Int64Value(obj.Spec.ForProvider.ClusterConfig.ZoneAwarenessConfig.AvailabilityZoneCount) != aws.Int64Value(out.DomainStatus.ClusterConfig.ZoneAwarenessConfig.AvailabilityZoneCount) {
						return false, nil
					}
				}
			} else if obj.Spec.ForProvider.ClusterConfig.ZoneAwarenessConfig != nil {
				return false, nil
			}
			if aws.BoolValue(obj.Spec.ForProvider.ClusterConfig.ZoneAwarenessEnabled) != aws.BoolValue(out.DomainStatus.ClusterConfig.ZoneAwarenessEnabled) {
				return false, nil
			}
		} else {
			return false, nil
		}
	} else if obj.Spec.ForProvider.ClusterConfig != nil {
		return false, nil
	}
	if out.DomainStatus.CognitoOptions != nil {
		if obj.Spec.ForProvider.CognitoOptions != nil {
			if aws.BoolValue(obj.Spec.ForProvider.CognitoOptions.Enabled) != aws.BoolValue(out.DomainStatus.CognitoOptions.Enabled) {
				return false, nil
			}
			if aws.StringValue(obj.Spec.ForProvider.CognitoOptions.IdentityPoolID) != aws.StringValue(out.DomainStatus.CognitoOptions.IdentityPoolId) {
				return false, nil
			}
			if aws.StringValue(obj.Spec.ForProvider.CognitoOptions.RoleARN) != aws.StringValue(out.DomainStatus.CognitoOptions.RoleArn) {
				return false, nil
			}
			if aws.StringValue(obj.Spec.ForProvider.CognitoOptions.UserPoolID) != aws.StringValue(out.DomainStatus.CognitoOptions.UserPoolId) {
				return false, nil
			}
		} else {
			return false, nil
		}
	} else if obj.Spec.ForProvider.DomainEndpointOptions != nil {
		return false, nil
	}
	if out.DomainStatus.DomainEndpointOptions != nil {
		if obj.Spec.ForProvider.DomainEndpointOptions != nil {
			if aws.StringValue(obj.Spec.ForProvider.DomainEndpointOptions.CustomEndpoint) != aws.StringValue(out.DomainStatus.DomainEndpointOptions.CustomEndpoint) {
				return false, nil
			}
			if aws.StringValue(obj.Spec.ForProvider.DomainEndpointOptions.CustomEndpointCertificateARN) != aws.StringValue(out.DomainStatus.DomainEndpointOptions.CustomEndpointCertificateArn) {
				return false, nil
			}
			if aws.BoolValue(obj.Spec.ForProvider.DomainEndpointOptions.CustomEndpointEnabled) != aws.BoolValue(out.DomainStatus.DomainEndpointOptions.CustomEndpointEnabled) {
				return false, nil
			}
			if aws.BoolValue(obj.Spec.ForProvider.DomainEndpointOptions.EnforceHTTPS) != aws.BoolValue(out.DomainStatus.DomainEndpointOptions.EnforceHTTPS) {
				return false, nil
			}
			if aws.StringValue(obj.Spec.ForProvider.DomainEndpointOptions.TLSSecurityPolicy) != aws.StringValue(out.DomainStatus.DomainEndpointOptions.TLSSecurityPolicy) {
				return false, nil
			}
		} else {
			return false, nil
		}
	} else if obj.Spec.ForProvider.DomainEndpointOptions != nil {
		return false, nil
	}
	if aws.StringValue(obj.Spec.ForProvider.Name) != aws.StringValue(out.DomainStatus.DomainName) {
		return false, nil
	}
	if out.DomainStatus.EBSOptions != nil {
		if obj.Spec.ForProvider.EBSOptions != nil {
			if aws.BoolValue(obj.Spec.ForProvider.EBSOptions.EBSEnabled) != aws.BoolValue(out.DomainStatus.EBSOptions.EBSEnabled) {
				return false, nil
			}
			if aws.Int64Value(obj.Spec.ForProvider.EBSOptions.IOPS) != aws.Int64Value(out.DomainStatus.EBSOptions.Iops) {
				return false, nil
			}
			if aws.Int64Value(obj.Spec.ForProvider.EBSOptions.VolumeSize) != aws.Int64Value(out.DomainStatus.EBSOptions.VolumeSize) {
				return false, nil
			}
			if aws.StringValue(obj.Spec.ForProvider.EBSOptions.VolumeType) != aws.StringValue(out.DomainStatus.EBSOptions.VolumeType) {
				return false, nil
			}
		} else {
			return false, nil
		}
	} else if obj.Spec.ForProvider.EBSOptions != nil {
		return false, nil
	}

	if aws.StringValue(obj.Spec.ForProvider.EngineVersion) != aws.StringValue(out.DomainStatus.EngineVersion) {
		return false, nil
	}

	if out.DomainStatus.LogPublishingOptions != nil {
		if obj.Spec.ForProvider.LogPublishingOptions != nil {
			if len(obj.Spec.ForProvider.LogPublishingOptions) != len(out.DomainStatus.LogPublishingOptions) {
				return false, nil
			}
			for key, value := range obj.Spec.ForProvider.LogPublishingOptions {
				if aws.StringValue(value.CloudWatchLogsLogGroupARN) != aws.StringValue(out.DomainStatus.LogPublishingOptions[key].CloudWatchLogsLogGroupArn) {
					return false, nil
				}
				if aws.BoolValue(value.Enabled) != aws.BoolValue(out.DomainStatus.LogPublishingOptions[key].Enabled) {
					return false, nil
				}
			}

		} else {
			return false, nil
		}
	} else if obj.Spec.ForProvider.LogPublishingOptions != nil {
		return false, nil
	}

	if out.DomainStatus.NodeToNodeEncryptionOptions != nil {
		if obj.Spec.ForProvider.NodeToNodeEncryptionOptions != nil {
			if aws.BoolValue(obj.Spec.ForProvider.NodeToNodeEncryptionOptions.Enabled) != aws.BoolValue(out.DomainStatus.NodeToNodeEncryptionOptions.Enabled) {
				return false, nil
			}
		} else {
			return false, nil
		}
	} else if obj.Spec.ForProvider.NodeToNodeEncryptionOptions != nil {
		return false, nil
	}

	if out.DomainStatus.SnapshotOptions != nil {
		if obj.Spec.ForProvider.SnapshotOptions != nil {
			if aws.Int64Value(out.DomainStatus.SnapshotOptions.AutomatedSnapshotStartHour) != aws.Int64Value(obj.Spec.ForProvider.SnapshotOptions.AutomatedSnapshotStartHour) {
				return false, nil
			}
		}
	} else if obj.Spec.ForProvider.SnapshotOptions != nil {
		return false, nil
	}

	if out.DomainStatus.VPCOptions != nil {
		if obj.Spec.ForProvider.VPCOptions != nil {
			if len(out.DomainStatus.VPCOptions.SecurityGroupIds) != len(obj.Spec.ForProvider.VPCOptions.SecurityGroupIDRefs) {
				return false, nil
			}
			for _, objValue := range obj.Spec.ForProvider.VPCOptions.SecurityGroupIDs {
				found := false
				for _, valueOut := range out.DomainStatus.VPCOptions.SecurityGroupIds {
					if aws.StringValue(objValue) == aws.StringValue(valueOut) {
						found = true
						break
					}
				}
				if !found {
					return false, nil
				}

			}
			if len(out.DomainStatus.VPCOptions.SubnetIds) != len(obj.Spec.ForProvider.VPCOptions.SubnetIDs) {
				return false, nil
			}
			for _, objValue := range obj.Spec.ForProvider.VPCOptions.SubnetIDs {
				found := false
				for _, valueOut := range out.DomainStatus.VPCOptions.SubnetIds {
					if aws.StringValue(objValue) == aws.StringValue(valueOut) {
						found = true
						break
					}
				}
				if !found {
					return false, nil
				}
			}
		}
	} else if obj.Spec.ForProvider.VPCOptions != nil {
		return false, nil
	}

	return true, nil
}

func lateInitialize(cr *svcapitypes.DomainParameters, resp *svcsdk.DescribeDomainOutput) error { // nolint:gocyclo

	if resp.DomainStatus.AccessPolicies != nil && resp.DomainStatus.AccessPolicies == nil {
		cr.AccessPolicies = resp.DomainStatus.AccessPolicies
	}
	if resp.DomainStatus.AdvancedOptions != nil {
		if cr.AdvancedOptions != nil {
			for key, value := range resp.DomainStatus.AdvancedOptions {
				if _, ok := cr.AdvancedOptions[key]; !ok {
					cr.AdvancedOptions[key] = value
				}
			}
		} else {
			f2 := map[string]*string{}
			for f2key, f2valiter := range resp.DomainStatus.AdvancedOptions {
				f2val := *f2valiter
				f2[f2key] = &f2val
			}
			cr.AdvancedOptions = f2
		}
	}
	if resp.DomainStatus.AdvancedSecurityOptions != nil {
		if cr.AdvancedSecurityOptions != nil {
			if resp.DomainStatus.AdvancedSecurityOptions.AnonymousAuthEnabled != nil && cr.AdvancedSecurityOptions.AnonymousAuthEnabled == nil {
				cr.AdvancedSecurityOptions.AnonymousAuthEnabled = resp.DomainStatus.AdvancedSecurityOptions.AnonymousAuthEnabled
			}
			if resp.DomainStatus.AdvancedSecurityOptions.Enabled != nil && cr.AdvancedSecurityOptions.Enabled == nil {
				cr.AdvancedSecurityOptions.Enabled = resp.DomainStatus.AdvancedSecurityOptions.Enabled
			}
			if resp.DomainStatus.AdvancedSecurityOptions.InternalUserDatabaseEnabled != nil && cr.AdvancedSecurityOptions.InternalUserDatabaseEnabled == nil {
				cr.AdvancedSecurityOptions.InternalUserDatabaseEnabled = resp.DomainStatus.AdvancedSecurityOptions.InternalUserDatabaseEnabled
			}
			if resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions != nil {
				if cr.AdvancedSecurityOptions.SAMLOptions != nil {
					if resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.Enabled != nil && cr.AdvancedSecurityOptions.SAMLOptions.Enabled == nil {
						cr.AdvancedSecurityOptions.SAMLOptions.Enabled = resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.Enabled
					}
					if resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.Idp != nil {
						if cr.AdvancedSecurityOptions.SAMLOptions.IDp != nil {
							if resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.Idp.EntityId != nil && cr.AdvancedSecurityOptions.SAMLOptions.IDp.EntityID == nil {
								cr.AdvancedSecurityOptions.SAMLOptions.IDp.EntityID = resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.Idp.EntityId
							}
							if resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.Idp.MetadataContent != nil && cr.AdvancedSecurityOptions.SAMLOptions.IDp.MetadataContent == nil {
								cr.AdvancedSecurityOptions.SAMLOptions.IDp.MetadataContent = resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.Idp.MetadataContent
							}
						} else {
							f3f4f1 := &svcapitypes.SAMLIDp{}
							if resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.Idp.EntityId != nil {
								f3f4f1.EntityID = resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.Idp.EntityId
							}
							if resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.Idp.MetadataContent != nil {
								f3f4f1.MetadataContent = resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.Idp.MetadataContent
							}
							cr.AdvancedSecurityOptions.SAMLOptions.IDp = f3f4f1
						}
					}
					if resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.RolesKey != nil && cr.AdvancedSecurityOptions.SAMLOptions.RolesKey == nil {
						cr.AdvancedSecurityOptions.SAMLOptions.RolesKey = resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.RolesKey
					}
					if resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.SessionTimeoutMinutes != nil && cr.AdvancedSecurityOptions.SAMLOptions.SessionTimeoutMinutes == nil {
						cr.AdvancedSecurityOptions.SAMLOptions.SessionTimeoutMinutes = resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.SessionTimeoutMinutes
					}
					if resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.SubjectKey != nil && cr.AdvancedSecurityOptions.SAMLOptions.SubjectKey == nil {
						cr.AdvancedSecurityOptions.SAMLOptions.SubjectKey = resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.SubjectKey
					}
				} else {
					f3f4 := &svcapitypes.SAMLOptionsInput{}
					if resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.Enabled != nil {
						f3f4.Enabled = resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.Enabled
					}
					if resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.Idp != nil {
						f3f4f1 := &svcapitypes.SAMLIDp{}
						if resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.Idp.EntityId != nil {
							f3f4f1.EntityID = resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.Idp.EntityId
						}
						if resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.Idp.MetadataContent != nil {
							f3f4f1.MetadataContent = resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.Idp.MetadataContent
						}
						f3f4.IDp = f3f4f1
					}
					if resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.RolesKey != nil {
						f3f4.RolesKey = resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.RolesKey
					}
					if resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.SessionTimeoutMinutes != nil {
						f3f4.SessionTimeoutMinutes = resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.SessionTimeoutMinutes
					}
					if resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.SubjectKey != nil {
						f3f4.SubjectKey = resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.SubjectKey
					}
					cr.AdvancedSecurityOptions.SAMLOptions = f3f4
				}
			}
		} else {
			f3 := &svcapitypes.AdvancedSecurityOptionsInput{}
			if resp.DomainStatus.AdvancedSecurityOptions.AnonymousAuthEnabled != nil {
				f3.AnonymousAuthEnabled = resp.DomainStatus.AdvancedSecurityOptions.AnonymousAuthEnabled
			}
			if resp.DomainStatus.AdvancedSecurityOptions.Enabled != nil {
				f3.Enabled = resp.DomainStatus.AdvancedSecurityOptions.Enabled
			}
			if resp.DomainStatus.AdvancedSecurityOptions.InternalUserDatabaseEnabled != nil {
				f3.InternalUserDatabaseEnabled = resp.DomainStatus.AdvancedSecurityOptions.InternalUserDatabaseEnabled
			}
			if resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions != nil {
				f3f4 := &svcapitypes.SAMLOptionsInput{}
				if resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.Enabled != nil {
					f3f4.Enabled = resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.Enabled
				}
				if resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.Idp != nil {
					f3f4f1 := &svcapitypes.SAMLIDp{}
					if resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.Idp.EntityId != nil {
						f3f4f1.EntityID = resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.Idp.EntityId
					}
					if resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.Idp.MetadataContent != nil {
						f3f4f1.MetadataContent = resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.Idp.MetadataContent
					}
					f3f4.IDp = f3f4f1
				}
				if resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.RolesKey != nil {
					f3f4.RolesKey = resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.RolesKey
				}
				if resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.SessionTimeoutMinutes != nil {
					f3f4.SessionTimeoutMinutes = resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.SessionTimeoutMinutes
				}
				if resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.SubjectKey != nil {
					f3f4.SubjectKey = resp.DomainStatus.AdvancedSecurityOptions.SAMLOptions.SubjectKey
				}
				f3.SAMLOptions = f3f4
			}
			cr.AdvancedSecurityOptions = f3
		}
	}
	if resp.DomainStatus.AutoTuneOptions != nil && cr.AutoTuneOptions == nil {
		f4 := &svcapitypes.AutoTuneOptionsInput{}
		cr.AutoTuneOptions = f4
	}
	if resp.DomainStatus.ClusterConfig != nil {
		if cr.ClusterConfig != nil {

			if resp.DomainStatus.ClusterConfig.ColdStorageOptions != nil {
				if cr.ClusterConfig.ColdStorageOptions != nil {
					if resp.DomainStatus.ClusterConfig.ColdStorageOptions.Enabled != nil && cr.ClusterConfig.ColdStorageOptions.Enabled == nil {
						cr.ClusterConfig.ColdStorageOptions.Enabled = resp.DomainStatus.ClusterConfig.ColdStorageOptions.Enabled
					}
				} else {
					f6f0 := &svcapitypes.ColdStorageOptions{}
					if resp.DomainStatus.ClusterConfig.ColdStorageOptions.Enabled != nil {
						f6f0.Enabled = resp.DomainStatus.ClusterConfig.ColdStorageOptions.Enabled
					}
					cr.ClusterConfig.ColdStorageOptions = f6f0
				}
				if resp.DomainStatus.ClusterConfig.DedicatedMasterCount != nil && cr.ClusterConfig.DedicatedMasterCount == nil {
					cr.ClusterConfig.DedicatedMasterCount = resp.DomainStatus.ClusterConfig.DedicatedMasterCount
				}
				if resp.DomainStatus.ClusterConfig.DedicatedMasterEnabled != nil && cr.ClusterConfig.DedicatedMasterEnabled == nil {
					cr.ClusterConfig.DedicatedMasterEnabled = resp.DomainStatus.ClusterConfig.DedicatedMasterEnabled
				}
				if resp.DomainStatus.ClusterConfig.DedicatedMasterType != nil && cr.ClusterConfig.DedicatedMasterType == nil {
					cr.ClusterConfig.DedicatedMasterType = resp.DomainStatus.ClusterConfig.DedicatedMasterType
				}
				if resp.DomainStatus.ClusterConfig.InstanceCount != nil && cr.ClusterConfig.InstanceCount == nil {
					cr.ClusterConfig.InstanceCount = resp.DomainStatus.ClusterConfig.InstanceCount
				}
				if resp.DomainStatus.ClusterConfig.InstanceType != nil && cr.ClusterConfig.InstanceType == nil {
					cr.ClusterConfig.InstanceType = resp.DomainStatus.ClusterConfig.InstanceType
				}
				if resp.DomainStatus.ClusterConfig.WarmCount != nil && cr.ClusterConfig.WarmCount == nil {
					cr.ClusterConfig.WarmCount = resp.DomainStatus.ClusterConfig.WarmCount
				}
				if resp.DomainStatus.ClusterConfig.WarmEnabled != nil && cr.ClusterConfig.WarmEnabled == nil {
					cr.ClusterConfig.WarmEnabled = resp.DomainStatus.ClusterConfig.WarmEnabled
				}
				if resp.DomainStatus.ClusterConfig.WarmType != nil && cr.ClusterConfig.WarmType == nil {
					cr.ClusterConfig.WarmType = resp.DomainStatus.ClusterConfig.WarmType
				}
				if resp.DomainStatus.ClusterConfig.ZoneAwarenessConfig != nil {
					if cr.ClusterConfig.ZoneAwarenessConfig != nil {
						if resp.DomainStatus.ClusterConfig.ZoneAwarenessConfig.AvailabilityZoneCount != nil && cr.ClusterConfig.ZoneAwarenessConfig.AvailabilityZoneCount == nil {
							cr.ClusterConfig.ZoneAwarenessConfig.AvailabilityZoneCount = resp.DomainStatus.ClusterConfig.ZoneAwarenessConfig.AvailabilityZoneCount
						}
					} else {
						f6f9 := &svcapitypes.ZoneAwarenessConfig{}
						if resp.DomainStatus.ClusterConfig.ZoneAwarenessConfig.AvailabilityZoneCount != nil {
							f6f9.AvailabilityZoneCount = resp.DomainStatus.ClusterConfig.ZoneAwarenessConfig.AvailabilityZoneCount
						}
						cr.ClusterConfig.ZoneAwarenessConfig = f6f9
					}
				}
				if resp.DomainStatus.ClusterConfig.ZoneAwarenessEnabled != nil && cr.ClusterConfig.ZoneAwarenessEnabled == nil {
					cr.ClusterConfig.ZoneAwarenessEnabled = resp.DomainStatus.ClusterConfig.ZoneAwarenessEnabled
				}
			}
		} else {
			f6 := &svcapitypes.ClusterConfig{}
			if resp.DomainStatus.ClusterConfig.ColdStorageOptions != nil {
				f6f0 := &svcapitypes.ColdStorageOptions{}
				if resp.DomainStatus.ClusterConfig.ColdStorageOptions.Enabled != nil {
					f6f0.Enabled = resp.DomainStatus.ClusterConfig.ColdStorageOptions.Enabled
				}
				f6.ColdStorageOptions = f6f0
			}
			if resp.DomainStatus.ClusterConfig.DedicatedMasterCount != nil {
				f6.DedicatedMasterCount = resp.DomainStatus.ClusterConfig.DedicatedMasterCount
			}
			if resp.DomainStatus.ClusterConfig.DedicatedMasterEnabled != nil {
				f6.DedicatedMasterEnabled = resp.DomainStatus.ClusterConfig.DedicatedMasterEnabled
			}
			if resp.DomainStatus.ClusterConfig.DedicatedMasterType != nil {
				f6.DedicatedMasterType = resp.DomainStatus.ClusterConfig.DedicatedMasterType
			}
			if resp.DomainStatus.ClusterConfig.InstanceCount != nil {
				f6.InstanceCount = resp.DomainStatus.ClusterConfig.InstanceCount
			}
			if resp.DomainStatus.ClusterConfig.InstanceType != nil {
				f6.InstanceType = resp.DomainStatus.ClusterConfig.InstanceType
			}
			if resp.DomainStatus.ClusterConfig.WarmCount != nil {
				f6.WarmCount = resp.DomainStatus.ClusterConfig.WarmCount
			}
			if resp.DomainStatus.ClusterConfig.WarmEnabled != nil {
				f6.WarmEnabled = resp.DomainStatus.ClusterConfig.WarmEnabled
			}
			if resp.DomainStatus.ClusterConfig.WarmType != nil {
				f6.WarmType = resp.DomainStatus.ClusterConfig.WarmType
			}
			if resp.DomainStatus.ClusterConfig.ZoneAwarenessConfig != nil {
				f6f9 := &svcapitypes.ZoneAwarenessConfig{}
				if resp.DomainStatus.ClusterConfig.ZoneAwarenessConfig.AvailabilityZoneCount != nil {
					f6f9.AvailabilityZoneCount = resp.DomainStatus.ClusterConfig.ZoneAwarenessConfig.AvailabilityZoneCount
				}
				f6.ZoneAwarenessConfig = f6f9
			}
			if resp.DomainStatus.ClusterConfig.ZoneAwarenessEnabled != nil {
				f6.ZoneAwarenessEnabled = resp.DomainStatus.ClusterConfig.ZoneAwarenessEnabled
			}
			cr.ClusterConfig = f6
		}
	}
	if resp.DomainStatus.CognitoOptions != nil {
		if cr.CognitoOptions != nil {
			if resp.DomainStatus.CognitoOptions.Enabled != nil && cr.CognitoOptions.Enabled == nil {
				cr.CognitoOptions.Enabled = resp.DomainStatus.CognitoOptions.Enabled
			}
			if resp.DomainStatus.CognitoOptions.IdentityPoolId != nil && cr.CognitoOptions.IdentityPoolID == nil {
				cr.CognitoOptions.IdentityPoolID = resp.DomainStatus.CognitoOptions.IdentityPoolId
			}
			if resp.DomainStatus.CognitoOptions.RoleArn != nil && cr.CognitoOptions.RoleARN == nil {
				cr.CognitoOptions.RoleARN = resp.DomainStatus.CognitoOptions.RoleArn
			}
			if resp.DomainStatus.CognitoOptions.UserPoolId != nil && cr.CognitoOptions.UserPoolID == nil {
				cr.CognitoOptions.UserPoolID = resp.DomainStatus.CognitoOptions.UserPoolId
			}
		} else {
			f7 := &svcapitypes.CognitoOptions{}
			if resp.DomainStatus.CognitoOptions.Enabled != nil {
				f7.Enabled = resp.DomainStatus.CognitoOptions.Enabled
			}
			if resp.DomainStatus.CognitoOptions.IdentityPoolId != nil {
				f7.IdentityPoolID = resp.DomainStatus.CognitoOptions.IdentityPoolId
			}
			if resp.DomainStatus.CognitoOptions.RoleArn != nil {
				f7.RoleARN = resp.DomainStatus.CognitoOptions.RoleArn
			}
			if resp.DomainStatus.CognitoOptions.UserPoolId != nil {
				f7.UserPoolID = resp.DomainStatus.CognitoOptions.UserPoolId
			}
			cr.CognitoOptions = f7
		}

	}
	if resp.DomainStatus.DomainEndpointOptions != nil {
		if cr.DomainEndpointOptions != nil {
			if resp.DomainStatus.DomainEndpointOptions.CustomEndpoint != nil && cr.DomainEndpointOptions.CustomEndpoint == nil {
				cr.DomainEndpointOptions.CustomEndpoint = resp.DomainStatus.DomainEndpointOptions.CustomEndpoint
			}
			if resp.DomainStatus.DomainEndpointOptions.CustomEndpointCertificateArn != nil && cr.DomainEndpointOptions.CustomEndpointCertificateARN == nil {
				cr.DomainEndpointOptions.CustomEndpointCertificateARN = resp.DomainStatus.DomainEndpointOptions.CustomEndpointCertificateArn
			}
			if resp.DomainStatus.DomainEndpointOptions.CustomEndpointEnabled != nil && cr.DomainEndpointOptions.CustomEndpointEnabled == nil {
				cr.DomainEndpointOptions.CustomEndpointEnabled = resp.DomainStatus.DomainEndpointOptions.CustomEndpointEnabled
			}
			if resp.DomainStatus.DomainEndpointOptions.EnforceHTTPS != nil && cr.DomainEndpointOptions.EnforceHTTPS == nil {
				cr.DomainEndpointOptions.EnforceHTTPS = resp.DomainStatus.DomainEndpointOptions.EnforceHTTPS
			}
			if resp.DomainStatus.DomainEndpointOptions.TLSSecurityPolicy != nil && cr.DomainEndpointOptions.TLSSecurityPolicy == nil {
				cr.DomainEndpointOptions.TLSSecurityPolicy = resp.DomainStatus.DomainEndpointOptions.TLSSecurityPolicy
			}
		} else {
			f10 := &svcapitypes.DomainEndpointOptions{}
			if resp.DomainStatus.DomainEndpointOptions.CustomEndpoint != nil {
				f10.CustomEndpoint = resp.DomainStatus.DomainEndpointOptions.CustomEndpoint
			}
			if resp.DomainStatus.DomainEndpointOptions.CustomEndpointCertificateArn != nil {
				f10.CustomEndpointCertificateARN = resp.DomainStatus.DomainEndpointOptions.CustomEndpointCertificateArn
			}
			if resp.DomainStatus.DomainEndpointOptions.CustomEndpointEnabled != nil {
				f10.CustomEndpointEnabled = resp.DomainStatus.DomainEndpointOptions.CustomEndpointEnabled
			}
			if resp.DomainStatus.DomainEndpointOptions.EnforceHTTPS != nil {
				f10.EnforceHTTPS = resp.DomainStatus.DomainEndpointOptions.EnforceHTTPS
			}
			if resp.DomainStatus.DomainEndpointOptions.TLSSecurityPolicy != nil {
				f10.TLSSecurityPolicy = resp.DomainStatus.DomainEndpointOptions.TLSSecurityPolicy
			}
			cr.DomainEndpointOptions = f10
		}
	}
	if resp.DomainStatus.DomainName != nil && cr.Name == nil {
		cr.Name = resp.DomainStatus.DomainName
	}
	if resp.DomainStatus.EBSOptions != nil {
		if cr.EBSOptions != nil {
			if resp.DomainStatus.EBSOptions.EBSEnabled != nil && cr.EBSOptions.EBSEnabled == nil {
				cr.EBSOptions.EBSEnabled = resp.DomainStatus.EBSOptions.EBSEnabled
			}
			if resp.DomainStatus.EBSOptions.Iops != nil && cr.EBSOptions.IOPS == nil {
				cr.EBSOptions.IOPS = resp.DomainStatus.EBSOptions.Iops
			}
			if resp.DomainStatus.EBSOptions.VolumeSize != nil && cr.EBSOptions.VolumeSize == nil {
				cr.EBSOptions.VolumeSize = resp.DomainStatus.EBSOptions.VolumeSize
			}
			if resp.DomainStatus.EBSOptions.VolumeType != nil && cr.EBSOptions.VolumeType == nil {
				cr.EBSOptions.VolumeType = resp.DomainStatus.EBSOptions.VolumeType
			}
		} else {
			f13 := &svcapitypes.EBSOptions{}
			if resp.DomainStatus.EBSOptions.EBSEnabled != nil {
				f13.EBSEnabled = resp.DomainStatus.EBSOptions.EBSEnabled
			}
			if resp.DomainStatus.EBSOptions.Iops != nil {
				f13.IOPS = resp.DomainStatus.EBSOptions.Iops
			}
			if resp.DomainStatus.EBSOptions.VolumeSize != nil {
				f13.VolumeSize = resp.DomainStatus.EBSOptions.VolumeSize
			}
			if resp.DomainStatus.EBSOptions.VolumeType != nil {
				f13.VolumeType = resp.DomainStatus.EBSOptions.VolumeType
			}
			cr.EBSOptions = f13
		}
	}
	if resp.DomainStatus.EngineVersion != nil && cr.EngineVersion == nil {
		cr.EngineVersion = resp.DomainStatus.EngineVersion
	}
	if resp.DomainStatus.LogPublishingOptions != nil {
		if cr.LogPublishingOptions != nil {
			for key, value := range resp.DomainStatus.LogPublishingOptions {
				if _, ok := cr.LogPublishingOptions[key]; !ok {
					cr.LogPublishingOptions[key] = &svcapitypes.LogPublishingOption{
						CloudWatchLogsLogGroupARN: value.CloudWatchLogsLogGroupArn,
					}
				}
			}
		} else {
			f18 := map[string]*svcapitypes.LogPublishingOption{}
			for f18key, f18valiter := range resp.DomainStatus.LogPublishingOptions {
				f18val := &svcapitypes.LogPublishingOption{}
				if f18valiter.CloudWatchLogsLogGroupArn != nil {
					f18val.CloudWatchLogsLogGroupARN = f18valiter.CloudWatchLogsLogGroupArn
				}
				if f18valiter.Enabled != nil {
					f18val.Enabled = f18valiter.Enabled
				}
				f18[f18key] = f18val
			}
			cr.LogPublishingOptions = f18
		}
	}
	if resp.DomainStatus.NodeToNodeEncryptionOptions != nil {
		if cr.NodeToNodeEncryptionOptions != nil {
			if resp.DomainStatus.NodeToNodeEncryptionOptions.Enabled != nil && cr.NodeToNodeEncryptionOptions.Enabled == nil {
				cr.NodeToNodeEncryptionOptions.Enabled = resp.DomainStatus.NodeToNodeEncryptionOptions.Enabled
			}
		} else {
			f19 := &svcapitypes.NodeToNodeEncryptionOptions{}
			if resp.DomainStatus.NodeToNodeEncryptionOptions.Enabled != nil {
				f19.Enabled = resp.DomainStatus.NodeToNodeEncryptionOptions.Enabled
			}
			cr.NodeToNodeEncryptionOptions = f19
		}
	}

	if resp.DomainStatus.SnapshotOptions != nil {
		if cr.SnapshotOptions != nil {
			if resp.DomainStatus.SnapshotOptions.AutomatedSnapshotStartHour != nil && cr.SnapshotOptions.AutomatedSnapshotStartHour == nil {
				cr.SnapshotOptions.AutomatedSnapshotStartHour = resp.DomainStatus.SnapshotOptions.AutomatedSnapshotStartHour
			}
		} else {
			cr.SnapshotOptions = &svcapitypes.SnapshotOptions{
				AutomatedSnapshotStartHour: resp.DomainStatus.SnapshotOptions.AutomatedSnapshotStartHour,
			}
		}
	}

	if resp.DomainStatus.VPCOptions != nil {
		if cr.VPCOptions != nil {
			if len(resp.DomainStatus.VPCOptions.SecurityGroupIds) > 0 && len(cr.VPCOptions.SecurityGroupIDs) == 0 {
				cr.VPCOptions.SecurityGroupIDs = resp.DomainStatus.VPCOptions.SecurityGroupIds
			}
			if len(resp.DomainStatus.VPCOptions.SubnetIds) > 0 && len(cr.VPCOptions.SubnetIDs) == 0 {
				cr.VPCOptions.SubnetIDs = resp.DomainStatus.VPCOptions.SubnetIds
			}
		} else {
			cr.VPCOptions = &svcapitypes.CustomVPCDerivedInfo{
				SecurityGroupIDs: resp.DomainStatus.VPCOptions.SecurityGroupIds,
				SubnetIDs:        resp.DomainStatus.VPCOptions.SubnetIds,
			}

		}
	}

	return nil
}

type updateDomain struct {
	client svcsdkapi.OpenSearchServiceAPI
}

func (e *updateDomain) update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) { // nolint:gocyclo
	cr, ok := mg.(*svcapitypes.Domain)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}
	input := svcsdk.UpdateDomainConfigInput{
		AccessPolicies:  cr.Spec.ForProvider.AccessPolicies,
		AdvancedOptions: cr.Spec.ForProvider.AdvancedOptions,
		DomainName:      cr.Spec.ForProvider.Name,
	}

	delete(input.AdvancedOptions, "override_main_response_version") // We need to ignore this

	if cr.Spec.ForProvider.AdvancedSecurityOptions != nil {
		input.AdvancedSecurityOptions = &svcsdk.AdvancedSecurityOptionsInput_{
			AnonymousAuthEnabled:        cr.Spec.ForProvider.AdvancedSecurityOptions.AnonymousAuthEnabled,
			Enabled:                     cr.Spec.ForProvider.AdvancedSecurityOptions.Enabled,
			InternalUserDatabaseEnabled: cr.Spec.ForProvider.AdvancedSecurityOptions.InternalUserDatabaseEnabled,
		}

		if cr.Spec.ForProvider.AdvancedSecurityOptions.MasterUserOptions != nil {
			input.AdvancedSecurityOptions.MasterUserOptions = &svcsdk.MasterUserOptions{
				MasterUserARN:      cr.Spec.ForProvider.AdvancedSecurityOptions.MasterUserOptions.MasterUserARN,
				MasterUserName:     cr.Spec.ForProvider.AdvancedSecurityOptions.MasterUserOptions.MasterUserName,
				MasterUserPassword: cr.Spec.ForProvider.AdvancedSecurityOptions.MasterUserOptions.MasterUserPassword,
			}
		}

		if cr.Spec.ForProvider.AdvancedSecurityOptions.SAMLOptions != nil {
			input.AdvancedSecurityOptions.SAMLOptions = &svcsdk.SAMLOptionsInput_{
				Enabled:               cr.Spec.ForProvider.AdvancedSecurityOptions.SAMLOptions.Enabled,
				MasterBackendRole:     cr.Spec.ForProvider.AdvancedSecurityOptions.SAMLOptions.MasterBackendRole,
				MasterUserName:        cr.Spec.ForProvider.AdvancedSecurityOptions.SAMLOptions.MasterUserName,
				RolesKey:              cr.Spec.ForProvider.AdvancedSecurityOptions.SAMLOptions.RolesKey,
				SessionTimeoutMinutes: cr.Spec.ForProvider.AdvancedSecurityOptions.SAMLOptions.SessionTimeoutMinutes,
				SubjectKey:            cr.Spec.ForProvider.AdvancedSecurityOptions.SAMLOptions.SubjectKey,
			}

			if cr.Spec.ForProvider.AdvancedSecurityOptions.SAMLOptions.IDp != nil {
				input.AdvancedSecurityOptions.SAMLOptions.Idp = &svcsdk.SAMLIdp{
					EntityId:        cr.Spec.ForProvider.AdvancedSecurityOptions.SAMLOptions.IDp.EntityID,
					MetadataContent: cr.Spec.ForProvider.AdvancedSecurityOptions.SAMLOptions.IDp.MetadataContent,
				}
			}
		}
	}

	if cr.Spec.ForProvider.AutoTuneOptions != nil {
		input.AutoTuneOptions = &svcsdk.AutoTuneOptions{
			DesiredState: cr.Spec.ForProvider.AutoTuneOptions.DesiredState,
		}
		if cr.Spec.ForProvider.AutoTuneOptions.MaintenanceSchedules != nil {
			msList := []*svcsdk.AutoTuneMaintenanceSchedule{}
			for _, ms := range cr.Spec.ForProvider.AutoTuneOptions.MaintenanceSchedules {
				msNew := &svcsdk.AutoTuneMaintenanceSchedule{
					CronExpressionForRecurrence: ms.CronExpressionForRecurrence,
				}
				if ms.Duration != nil {
					msNew.Duration = &svcsdk.Duration{
						Unit:  ms.Duration.Unit,
						Value: ms.Duration.Value,
					}
				}
				if ms.StartAt != nil {
					msNew.StartAt = &ms.StartAt.Time
				}
				msList = append(msList, msNew)
			}
			input.AutoTuneOptions.MaintenanceSchedules = msList
		}
	}

	if cr.Spec.ForProvider.ClusterConfig != nil {

		input.ClusterConfig = &svcsdk.ClusterConfig{
			DedicatedMasterCount:   cr.Spec.ForProvider.ClusterConfig.DedicatedMasterCount,
			DedicatedMasterEnabled: cr.Spec.ForProvider.ClusterConfig.DedicatedMasterEnabled,
			DedicatedMasterType:    cr.Spec.ForProvider.ClusterConfig.DedicatedMasterType,
			InstanceCount:          cr.Spec.ForProvider.ClusterConfig.InstanceCount,
			InstanceType:           cr.Spec.ForProvider.ClusterConfig.InstanceType,
			WarmCount:              cr.Spec.ForProvider.ClusterConfig.WarmCount,
			WarmEnabled:            cr.Spec.ForProvider.ClusterConfig.WarmEnabled,
			WarmType:               cr.Spec.ForProvider.ClusterConfig.WarmType,
			ZoneAwarenessEnabled:   cr.Spec.ForProvider.ClusterConfig.ZoneAwarenessEnabled,
		}

		if cr.Spec.ForProvider.ClusterConfig.ColdStorageOptions != nil {
			input.ClusterConfig.ColdStorageOptions = &svcsdk.ColdStorageOptions{
				Enabled: cr.Spec.ForProvider.ClusterConfig.ColdStorageOptions.Enabled,
			}
		}
		if cr.Spec.ForProvider.ClusterConfig.ZoneAwarenessConfig != nil {
			input.ClusterConfig.ZoneAwarenessConfig = &svcsdk.ZoneAwarenessConfig{
				AvailabilityZoneCount: cr.Spec.ForProvider.ClusterConfig.ZoneAwarenessConfig.AvailabilityZoneCount,
			}
		}

	}

	if cr.Spec.ForProvider.CognitoOptions != nil {
		input.CognitoOptions = &svcsdk.CognitoOptions{
			Enabled:        cr.Spec.ForProvider.CognitoOptions.Enabled,
			IdentityPoolId: cr.Spec.ForProvider.CognitoOptions.IdentityPoolID,
			RoleArn:        cr.Spec.ForProvider.CognitoOptions.RoleARN,
			UserPoolId:     cr.Spec.ForProvider.CognitoOptions.UserPoolID,
		}
	}

	if cr.Spec.ForProvider.DomainEndpointOptions != nil {
		input.DomainEndpointOptions = &svcsdk.DomainEndpointOptions{
			CustomEndpoint:               cr.Spec.ForProvider.DomainEndpointOptions.CustomEndpoint,
			CustomEndpointCertificateArn: cr.Spec.ForProvider.DomainEndpointOptions.CustomEndpointCertificateARN,
			CustomEndpointEnabled:        cr.Spec.ForProvider.DomainEndpointOptions.CustomEndpointEnabled,
			EnforceHTTPS:                 cr.Spec.ForProvider.DomainEndpointOptions.EnforceHTTPS,
			TLSSecurityPolicy:            cr.Spec.ForProvider.DomainEndpointOptions.TLSSecurityPolicy,
		}
	}

	if cr.Spec.ForProvider.EBSOptions != nil {
		input.EBSOptions = &svcsdk.EBSOptions{
			EBSEnabled: cr.Spec.ForProvider.EBSOptions.EBSEnabled,
			Iops:       cr.Spec.ForProvider.EBSOptions.IOPS,
			VolumeSize: cr.Spec.ForProvider.EBSOptions.VolumeSize,
			VolumeType: cr.Spec.ForProvider.EBSOptions.VolumeType,
		}
	}

	if cr.Spec.ForProvider.EncryptionAtRestOptions != nil {
		input.EncryptionAtRestOptions = &svcsdk.EncryptionAtRestOptions{
			Enabled:  cr.Spec.ForProvider.EncryptionAtRestOptions.Enabled,
			KmsKeyId: cr.Spec.ForProvider.EncryptionAtRestOptions.KMSKeyID,
		}
	}

	if cr.Spec.ForProvider.LogPublishingOptions != nil {
		input.LogPublishingOptions = map[string]*svcsdk.LogPublishingOption{}

		for key, value := range cr.Spec.ForProvider.LogPublishingOptions {
			input.LogPublishingOptions[key] = &svcsdk.LogPublishingOption{
				CloudWatchLogsLogGroupArn: value.CloudWatchLogsLogGroupARN,
				Enabled:                   value.Enabled,
			}
		}
	}

	if cr.Spec.ForProvider.NodeToNodeEncryptionOptions != nil {
		input.NodeToNodeEncryptionOptions = &svcsdk.NodeToNodeEncryptionOptions{
			Enabled: cr.Spec.ForProvider.NodeToNodeEncryptionOptions.Enabled,
		}
	}

	if cr.Spec.ForProvider.SnapshotOptions != nil {
		input.SnapshotOptions = &svcsdk.SnapshotOptions{
			AutomatedSnapshotStartHour: cr.Spec.ForProvider.SnapshotOptions.AutomatedSnapshotStartHour,
		}
	}

	if cr.Spec.ForProvider.VPCOptions != nil {
		input.VPCOptions = &svcsdk.VPCOptions{
			SecurityGroupIds: cr.Spec.ForProvider.VPCOptions.SecurityGroupIDs,
			SubnetIds:        cr.Spec.ForProvider.VPCOptions.SubnetIDs,
		}
	}

	resp, err := e.client.UpdateDomainConfigWithContext(ctx, &input)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	if resp.DomainConfig.AccessPolicies != nil && resp.DomainConfig.AccessPolicies.Options != nil {
		cr.Spec.ForProvider.AccessPolicies = resp.DomainConfig.AccessPolicies.Options
	} else {
		cr.Spec.ForProvider.AccessPolicies = nil
	}
	if resp.DomainConfig.AdvancedOptions != nil && resp.DomainConfig.AdvancedOptions.Options != nil {
		f2 := map[string]*string{}
		for f2key, f2valiter := range resp.DomainConfig.AdvancedOptions.Options {
			f2val := *f2valiter
			f2[f2key] = &f2val
		}
		cr.Spec.ForProvider.AdvancedOptions = f2
	} else {
		cr.Spec.ForProvider.AdvancedOptions = nil
	}
	if resp.DomainConfig.AdvancedSecurityOptions != nil && resp.DomainConfig.AdvancedSecurityOptions.Options != nil {
		f3 := &svcapitypes.AdvancedSecurityOptionsInput{}
		if resp.DomainConfig.AdvancedSecurityOptions.Options.AnonymousAuthEnabled != nil {
			f3.AnonymousAuthEnabled = resp.DomainConfig.AdvancedSecurityOptions.Options.AnonymousAuthEnabled
		}
		if resp.DomainConfig.AdvancedSecurityOptions.Options.Enabled != nil {
			f3.Enabled = resp.DomainConfig.AdvancedSecurityOptions.Options.Enabled
		}
		if resp.DomainConfig.AdvancedSecurityOptions.Options.InternalUserDatabaseEnabled != nil {
			f3.InternalUserDatabaseEnabled = resp.DomainConfig.AdvancedSecurityOptions.Options.InternalUserDatabaseEnabled
		}
		if resp.DomainConfig.AdvancedSecurityOptions.Options.SAMLOptions != nil {
			f3f4 := &svcapitypes.SAMLOptionsInput{}
			if resp.DomainConfig.AdvancedSecurityOptions.Options.SAMLOptions.Enabled != nil {
				f3f4.Enabled = resp.DomainConfig.AdvancedSecurityOptions.Options.SAMLOptions.Enabled
			}
			if resp.DomainConfig.AdvancedSecurityOptions.Options.SAMLOptions.Idp != nil {
				f3f4f1 := &svcapitypes.SAMLIDp{}
				if resp.DomainConfig.AdvancedSecurityOptions.Options.SAMLOptions.Idp.EntityId != nil {
					f3f4f1.EntityID = resp.DomainConfig.AdvancedSecurityOptions.Options.SAMLOptions.Idp.EntityId
				}
				if resp.DomainConfig.AdvancedSecurityOptions.Options.SAMLOptions.Idp.MetadataContent != nil {
					f3f4f1.MetadataContent = resp.DomainConfig.AdvancedSecurityOptions.Options.SAMLOptions.Idp.MetadataContent
				}
				f3f4.IDp = f3f4f1
			}
			if resp.DomainConfig.AdvancedSecurityOptions.Options.SAMLOptions.RolesKey != nil {
				f3f4.RolesKey = resp.DomainConfig.AdvancedSecurityOptions.Options.SAMLOptions.RolesKey
			}
			if resp.DomainConfig.AdvancedSecurityOptions.Options.SAMLOptions.SessionTimeoutMinutes != nil {
				f3f4.SessionTimeoutMinutes = resp.DomainConfig.AdvancedSecurityOptions.Options.SAMLOptions.SessionTimeoutMinutes
			}
			if resp.DomainConfig.AdvancedSecurityOptions.Options.SAMLOptions.SubjectKey != nil {
				f3f4.SubjectKey = resp.DomainConfig.AdvancedSecurityOptions.Options.SAMLOptions.SubjectKey
			}
			f3.SAMLOptions = f3f4
		}
		cr.Spec.ForProvider.AdvancedSecurityOptions = f3
	} else {
		cr.Spec.ForProvider.AdvancedSecurityOptions = nil
	}
	if resp.DomainConfig.AutoTuneOptions != nil {
		f4 := &svcapitypes.AutoTuneOptionsInput{}
		cr.Spec.ForProvider.AutoTuneOptions = f4
	} else {
		cr.Spec.ForProvider.AutoTuneOptions = nil
	}
	if resp.DomainConfig.ChangeProgressDetails != nil {
		f5 := &svcapitypes.ChangeProgressDetails{}
		if resp.DomainConfig.ChangeProgressDetails.ChangeId != nil {
			f5.ChangeID = resp.DomainConfig.ChangeProgressDetails.ChangeId
		}
		if resp.DomainConfig.ChangeProgressDetails.Message != nil {
			f5.Message = resp.DomainConfig.ChangeProgressDetails.Message
		}
		cr.Status.AtProvider.ChangeProgressDetails = f5
	} else {
		cr.Status.AtProvider.ChangeProgressDetails = nil
	}
	if resp.DomainConfig.ClusterConfig != nil && resp.DomainConfig.ClusterConfig.Options != nil {
		f6 := &svcapitypes.ClusterConfig{}
		if resp.DomainConfig.ClusterConfig.Options.ColdStorageOptions != nil {
			f6f0 := &svcapitypes.ColdStorageOptions{}
			if resp.DomainConfig.ClusterConfig.Options.ColdStorageOptions.Enabled != nil {
				f6f0.Enabled = resp.DomainConfig.ClusterConfig.Options.ColdStorageOptions.Enabled
			}
			f6.ColdStorageOptions = f6f0
		}
		if resp.DomainConfig.ClusterConfig.Options.DedicatedMasterCount != nil {
			f6.DedicatedMasterCount = resp.DomainConfig.ClusterConfig.Options.DedicatedMasterCount
		}
		if resp.DomainConfig.ClusterConfig.Options.DedicatedMasterEnabled != nil {
			f6.DedicatedMasterEnabled = resp.DomainConfig.ClusterConfig.Options.DedicatedMasterEnabled
		}
		if resp.DomainConfig.ClusterConfig.Options.DedicatedMasterType != nil {
			f6.DedicatedMasterType = resp.DomainConfig.ClusterConfig.Options.DedicatedMasterType
		}
		if resp.DomainConfig.ClusterConfig.Options.InstanceCount != nil {
			f6.InstanceCount = resp.DomainConfig.ClusterConfig.Options.InstanceCount
		}
		if resp.DomainConfig.ClusterConfig.Options.InstanceType != nil {
			f6.InstanceType = resp.DomainConfig.ClusterConfig.Options.InstanceType
		}
		if resp.DomainConfig.ClusterConfig.Options.WarmCount != nil {
			f6.WarmCount = resp.DomainConfig.ClusterConfig.Options.WarmCount
		}
		if resp.DomainConfig.ClusterConfig.Options.WarmEnabled != nil {
			f6.WarmEnabled = resp.DomainConfig.ClusterConfig.Options.WarmEnabled
		}
		if resp.DomainConfig.ClusterConfig.Options.WarmType != nil {
			f6.WarmType = resp.DomainConfig.ClusterConfig.Options.WarmType
		}
		if resp.DomainConfig.ClusterConfig.Options.ZoneAwarenessConfig != nil {
			f6f9 := &svcapitypes.ZoneAwarenessConfig{}
			if resp.DomainConfig.ClusterConfig.Options.ZoneAwarenessConfig.AvailabilityZoneCount != nil {
				f6f9.AvailabilityZoneCount = resp.DomainConfig.ClusterConfig.Options.ZoneAwarenessConfig.AvailabilityZoneCount
			}
			f6.ZoneAwarenessConfig = f6f9
		}
		if resp.DomainConfig.ClusterConfig.Options.ZoneAwarenessEnabled != nil {
			f6.ZoneAwarenessEnabled = resp.DomainConfig.ClusterConfig.Options.ZoneAwarenessEnabled
		}
		cr.Spec.ForProvider.ClusterConfig = f6
	} else {
		cr.Spec.ForProvider.ClusterConfig = nil
	}
	if resp.DomainConfig.CognitoOptions != nil && resp.DomainConfig.CognitoOptions.Options != nil {
		f7 := &svcapitypes.CognitoOptions{}
		if resp.DomainConfig.CognitoOptions.Options.Enabled != nil {
			f7.Enabled = resp.DomainConfig.CognitoOptions.Options.Enabled
		}
		if resp.DomainConfig.CognitoOptions.Options.IdentityPoolId != nil {
			f7.IdentityPoolID = resp.DomainConfig.CognitoOptions.Options.IdentityPoolId
		}
		if resp.DomainConfig.CognitoOptions.Options.RoleArn != nil {
			f7.RoleARN = resp.DomainConfig.CognitoOptions.Options.RoleArn
		}
		if resp.DomainConfig.CognitoOptions.Options.UserPoolId != nil {
			f7.UserPoolID = resp.DomainConfig.CognitoOptions.Options.UserPoolId
		}
		cr.Spec.ForProvider.CognitoOptions = f7
	} else {
		cr.Spec.ForProvider.CognitoOptions = nil
	}

	if resp.DomainConfig.DomainEndpointOptions != nil && resp.DomainConfig.DomainEndpointOptions.Options != nil {
		f10 := &svcapitypes.DomainEndpointOptions{}
		if resp.DomainConfig.DomainEndpointOptions.Options.CustomEndpoint != nil {
			f10.CustomEndpoint = resp.DomainConfig.DomainEndpointOptions.Options.CustomEndpoint
		}
		if resp.DomainConfig.DomainEndpointOptions.Options.CustomEndpointCertificateArn != nil {
			f10.CustomEndpointCertificateARN = resp.DomainConfig.DomainEndpointOptions.Options.CustomEndpointCertificateArn
		}
		if resp.DomainConfig.DomainEndpointOptions.Options.CustomEndpointEnabled != nil {
			f10.CustomEndpointEnabled = resp.DomainConfig.DomainEndpointOptions.Options.CustomEndpointEnabled
		}
		if resp.DomainConfig.DomainEndpointOptions.Options.EnforceHTTPS != nil {
			f10.EnforceHTTPS = resp.DomainConfig.DomainEndpointOptions.Options.EnforceHTTPS
		}
		if resp.DomainConfig.DomainEndpointOptions.Options.TLSSecurityPolicy != nil {
			f10.TLSSecurityPolicy = resp.DomainConfig.DomainEndpointOptions.Options.TLSSecurityPolicy
		}
		cr.Spec.ForProvider.DomainEndpointOptions = f10
	} else {
		cr.Spec.ForProvider.DomainEndpointOptions = nil
	}
	if resp.DomainConfig.EBSOptions != nil && resp.DomainConfig.EBSOptions.Options != nil {
		f13 := &svcapitypes.EBSOptions{}
		if resp.DomainConfig.EBSOptions.Options.EBSEnabled != nil {
			f13.EBSEnabled = resp.DomainConfig.EBSOptions.Options.EBSEnabled
		}
		if resp.DomainConfig.EBSOptions.Options.Iops != nil {
			f13.IOPS = resp.DomainConfig.EBSOptions.Options.Iops
		}
		if resp.DomainConfig.EBSOptions.Options.VolumeSize != nil {
			f13.VolumeSize = resp.DomainConfig.EBSOptions.Options.VolumeSize
		}
		if resp.DomainConfig.EBSOptions.Options.VolumeType != nil {
			f13.VolumeType = resp.DomainConfig.EBSOptions.Options.VolumeType
		}
		cr.Spec.ForProvider.EBSOptions = f13
	} else {
		cr.Spec.ForProvider.EBSOptions = nil
	}
	if resp.DomainConfig.EncryptionAtRestOptions != nil && resp.DomainConfig.EncryptionAtRestOptions.Options != nil {
		f14 := &svcapitypes.EncryptionAtRestOptions{}
		if resp.DomainConfig.EncryptionAtRestOptions.Options.Enabled != nil {
			f14.Enabled = resp.DomainConfig.EncryptionAtRestOptions.Options.Enabled
		}
		if resp.DomainConfig.EncryptionAtRestOptions.Options.KmsKeyId != nil {
			f14.KMSKeyID = resp.DomainConfig.EncryptionAtRestOptions.Options.KmsKeyId
		}
		cr.Status.AtProvider.EncryptionAtRestOptions = f14
	} else {
		cr.Status.AtProvider.EncryptionAtRestOptions = nil
	}
	if resp.DomainConfig.EngineVersion != nil && resp.DomainConfig.EngineVersion.Options != nil {
		cr.Spec.ForProvider.EngineVersion = resp.DomainConfig.EngineVersion.Options
	} else {
		cr.Spec.ForProvider.EngineVersion = nil
	}
	if resp.DomainConfig.LogPublishingOptions != nil && resp.DomainConfig.LogPublishingOptions.Options != nil {
		f18 := map[string]*svcapitypes.LogPublishingOption{}
		for f18key, f18valiter := range resp.DomainConfig.LogPublishingOptions.Options {
			f18val := &svcapitypes.LogPublishingOption{}
			if f18valiter.CloudWatchLogsLogGroupArn != nil {
				f18val.CloudWatchLogsLogGroupARN = f18valiter.CloudWatchLogsLogGroupArn
			}
			if f18valiter.Enabled != nil {
				f18val.Enabled = f18valiter.Enabled
			}
			f18[f18key] = f18val
		}
		cr.Spec.ForProvider.LogPublishingOptions = f18
	} else {
		cr.Spec.ForProvider.LogPublishingOptions = nil
	}
	if resp.DomainConfig.NodeToNodeEncryptionOptions != nil && resp.DomainConfig.NodeToNodeEncryptionOptions.Options != nil {
		f19 := &svcapitypes.NodeToNodeEncryptionOptions{}
		if resp.DomainConfig.NodeToNodeEncryptionOptions.Options.Enabled != nil {
			f19.Enabled = resp.DomainConfig.NodeToNodeEncryptionOptions.Options.Enabled
		}
		cr.Spec.ForProvider.NodeToNodeEncryptionOptions = f19
	} else {
		cr.Spec.ForProvider.NodeToNodeEncryptionOptions = nil
	}

	if resp.DomainConfig.SnapshotOptions != nil && resp.DomainConfig.SnapshotOptions.Options != nil {
		f22 := &svcapitypes.SnapshotOptions{}
		if resp.DomainConfig.SnapshotOptions.Options.AutomatedSnapshotStartHour != nil {
			f22.AutomatedSnapshotStartHour = resp.DomainConfig.SnapshotOptions.Options.AutomatedSnapshotStartHour
		}
		cr.Status.AtProvider.SnapshotOptions = f22
	} else {
		cr.Status.AtProvider.SnapshotOptions = nil
	}

	if resp.DomainConfig.VPCOptions != nil && resp.DomainConfig.VPCOptions.Options != nil {
		f24 := &svcapitypes.VPCDerivedInfo{}
		if resp.DomainConfig.VPCOptions.Options.AvailabilityZones != nil {
			f24f0 := []*string{}
			for _, f24f0iter := range resp.DomainConfig.VPCOptions.Options.AvailabilityZones {
				f24f0elem := *f24f0iter
				f24f0 = append(f24f0, &f24f0elem)
			}
			f24.AvailabilityZones = f24f0
		}
		if resp.DomainConfig.VPCOptions.Options.SecurityGroupIds != nil {
			f24f1 := []*string{}
			for _, f24f1iter := range resp.DomainConfig.VPCOptions.Options.SecurityGroupIds {
				f24f1elem := *f24f1iter
				f24f1 = append(f24f1, &f24f1elem)
			}
			f24.SecurityGroupIDs = f24f1
		}
		if resp.DomainConfig.VPCOptions.Options.SubnetIds != nil {
			f24f2 := []*string{}
			for _, f24f2iter := range resp.DomainConfig.VPCOptions.Options.SubnetIds {
				f24f2elem := *f24f2iter
				f24f2 = append(f24f2, &f24f2elem)
			}
			f24.SubnetIDs = f24f2
		}
		if resp.DomainConfig.VPCOptions.Options.VPCId != nil {
			f24.VPCID = resp.DomainConfig.VPCOptions.Options.VPCId
		}
		cr.Status.AtProvider.VPCOptions = f24
	} else {
		cr.Status.AtProvider.VPCOptions = nil
	}
	return managed.ExternalUpdate{}, nil
}
