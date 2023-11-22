package domain

import (
	"context"
	"strings"

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
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/opensearchservice/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
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

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(svcapitypes.DomainGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.Domain{}).
		Complete(r)
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
		"arn": []byte(pointer.StringValue(resp.DomainStatus.ARN)),
	}

	// public endpoints
	if resp.DomainStatus.Endpoint != nil {
		obs.ConnectionDetails["endpoint"] = []byte(pointer.StringValue(resp.DomainStatus.Endpoint))
	}

	// vpc endpoints
	if resp.DomainStatus.Endpoints != nil {
		var keys []string
		for key := range resp.DomainStatus.Endpoints {
			keys = append(keys, key)
		}
		for _, key := range keys {
			obs.ConnectionDetails[key+"Endpoint"] = []byte(pointer.StringValue(resp.DomainStatus.Endpoints[key]))
		}
	}

	return obs, nil
}

// GenerateObservation is used to produce DomainObservation
func GenerateObservation(obj *svcsdk.DomainStatus) svcapitypes.DomainObservation { //nolint:gocyclo
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

func isUpToDate(_ context.Context, obj *svcapitypes.Domain, out *svcsdk.DescribeDomainOutput) (bool, string, error) {

	switch {
	case aws.StringValue(obj.Spec.ForProvider.AccessPolicies) != aws.StringValue(out.DomainStatus.AccessPolicies),
		!isMapUpToDate(obj.Spec.ForProvider.AdvancedOptions, out.DomainStatus.AdvancedOptions, []string{"override_main_response_version"}),
		!isAdvancedSecurityOptionsUpToDate(obj.Spec.ForProvider.AdvancedSecurityOptions, out.DomainStatus.AdvancedSecurityOptions),
		!isAutoTuneOptionsUpToDate(obj.Spec.ForProvider.AutoTuneOptions, out.DomainStatus.AutoTuneOptions),
		!isClusterConfigUpToDate(obj.Spec.ForProvider.ClusterConfig, out.DomainStatus.ClusterConfig),
		!isCognitoOptionsUpToDate(obj.Spec.ForProvider.CognitoOptions, out.DomainStatus.CognitoOptions),
		!isDomainEndpointOptionsUpToDate(obj.Spec.ForProvider.DomainEndpointOptions, out.DomainStatus.DomainEndpointOptions),
		meta.GetExternalName(obj) != aws.StringValue(out.DomainStatus.DomainName),
		!isEbsOptionsUpToDate(obj.Spec.ForProvider.EBSOptions, out.DomainStatus.EBSOptions),
		!isLogPublishingOptionsUpToDate(obj.Spec.ForProvider.LogPublishingOptions, out.DomainStatus.LogPublishingOptions),
		!isNodeToNodeEncryptionOptionsUpToDate(obj.Spec.ForProvider.NodeToNodeEncryptionOptions, out.DomainStatus.NodeToNodeEncryptionOptions),
		!isSnapshotOptionsUpToDate(obj.Spec.ForProvider.SnapshotOptions, out.DomainStatus.SnapshotOptions),
		!isVpcOptionsUpToDate(obj.Spec.ForProvider.VPCOptions, out.DomainStatus.VPCOptions),
		!isEncryptionAtRestUpToDate(obj.Spec.ForProvider.EncryptionAtRestOptions, out.DomainStatus.EncryptionAtRestOptions):
		return false, "", nil
	case !*out.DomainStatus.Processing && !*out.DomainStatus.UpgradeProcessing && aws.StringValue(obj.Spec.ForProvider.EngineVersion) != aws.StringValue(out.DomainStatus.EngineVersion):
		// We cant update when processing or already updating so we consider this as up to date for now
		return false, "", nil
	default:
		return true, "", nil
	}
}

func lateInitialize(cr *svcapitypes.DomainParameters, resp *svcsdk.DescribeDomainOutput) error { //nolint:gocyclo

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
	// if resp.DomainStatus.DomainName != nil && cr.Name == nil {
	// 	cr.Name = resp.DomainStatus.DomainName
	// }
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

	if resp.DomainStatus.EncryptionAtRestOptions != nil && cr.EncryptionAtRestOptions == nil {
		f20 := &svcapitypes.CustomEncryptionAtRestOptions{}
		cr.EncryptionAtRestOptions = f20
	}

	return nil
}

type updateDomain struct {
	client svcsdkapi.OpenSearchServiceAPI
}

func (e *updateDomain) update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) { //nolint:gocyclo
	cr, ok := mg.(*svcapitypes.Domain)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}
	domainName := meta.GetExternalName(cr)
	stateInput := GenerateDescribeDomainInput(cr)
	stateInput.DomainName = &domainName
	obj, err := e.client.DescribeDomainWithContext(ctx, stateInput)

	if err != nil {
		return managed.ExternalUpdate{}, errors.New(errDescribe)
	}
	needDomainConfigUpdate := false
	input := svcsdk.UpdateDomainConfigInput{
		DomainName: &domainName,
	}
	if aws.StringValue(cr.Spec.ForProvider.AccessPolicies) != aws.StringValue(obj.DomainStatus.AccessPolicies) {
		needDomainConfigUpdate = true
		input.AccessPolicies = cr.Spec.ForProvider.AccessPolicies
	}
	if !isMapUpToDate(cr.Spec.ForProvider.AdvancedOptions, obj.DomainStatus.AdvancedOptions, []string{"override_main_response_version"}) {
		needDomainConfigUpdate = true
		input.AdvancedOptions = cr.Spec.ForProvider.AdvancedOptions
		delete(input.AdvancedOptions, "override_main_response_version") // We need to ignore this
	}

	if !isAdvancedSecurityOptionsUpToDate(cr.Spec.ForProvider.AdvancedSecurityOptions, obj.DomainStatus.AdvancedSecurityOptions) {
		needDomainConfigUpdate = true
		input.AdvancedSecurityOptions = generateAdvancedSecurityOptionsInput(cr.Spec.ForProvider.AdvancedSecurityOptions)
	}

	if !isClusterConfigUpToDate(cr.Spec.ForProvider.ClusterConfig, obj.DomainStatus.ClusterConfig) {
		needDomainConfigUpdate = true
		input.ClusterConfig = generateClusterConfig(cr.Spec.ForProvider.ClusterConfig)
	}

	if !isAutoTuneOptionsUpToDate(cr.Spec.ForProvider.AutoTuneOptions, obj.DomainStatus.AutoTuneOptions) {
		needDomainConfigUpdate = true
		input.AutoTuneOptions = generateAutotuneOptionsInput(cr.Spec.ForProvider.AutoTuneOptions)
	}

	if !isCognitoOptionsUpToDate(cr.Spec.ForProvider.CognitoOptions, obj.DomainStatus.CognitoOptions) {
		needDomainConfigUpdate = true
		input.CognitoOptions = generateCognitoOptions(cr.Spec.ForProvider.CognitoOptions)
	}

	if !isDomainEndpointOptionsUpToDate(cr.Spec.ForProvider.DomainEndpointOptions, obj.DomainStatus.DomainEndpointOptions) {
		needDomainConfigUpdate = true
		input.DomainEndpointOptions = generateDomainEndpointsOptions(cr.Spec.ForProvider.DomainEndpointOptions)
	}

	if !isEbsOptionsUpToDate(cr.Spec.ForProvider.EBSOptions, obj.DomainStatus.EBSOptions) {
		needDomainConfigUpdate = true
		input.EBSOptions = generateEbsOptions(cr.Spec.ForProvider.EBSOptions)
	}

	if !isLogPublishingOptionsUpToDate(cr.Spec.ForProvider.LogPublishingOptions, obj.DomainStatus.LogPublishingOptions) {
		needDomainConfigUpdate = true
		input.LogPublishingOptions = generateLogPublishingOptions(cr.Spec.ForProvider.LogPublishingOptions)
	}

	if !isNodeToNodeEncryptionOptionsUpToDate(cr.Spec.ForProvider.NodeToNodeEncryptionOptions, obj.DomainStatus.NodeToNodeEncryptionOptions) {
		needDomainConfigUpdate = true
		input.NodeToNodeEncryptionOptions = generateNodeToNodeEncryptionOptions(cr.Spec.ForProvider.NodeToNodeEncryptionOptions)
	}

	if !isSnapshotOptionsUpToDate(cr.Spec.ForProvider.SnapshotOptions, obj.DomainStatus.SnapshotOptions) {
		needDomainConfigUpdate = true
		input.SnapshotOptions = generateSnapshotOptions(cr.Spec.ForProvider.SnapshotOptions)
	}

	if !isVpcOptionsUpToDate(cr.Spec.ForProvider.VPCOptions, obj.DomainStatus.VPCOptions) {
		needDomainConfigUpdate = true
		input.VPCOptions = generateVpcOptions(cr.Spec.ForProvider.VPCOptions)
	}

	if !isEncryptionAtRestUpToDate(cr.Spec.ForProvider.EncryptionAtRestOptions, obj.DomainStatus.EncryptionAtRestOptions) {
		needDomainConfigUpdate = true
		input.EncryptionAtRestOptions = generateEncryptionAtRestOptions(cr.Spec.ForProvider.EncryptionAtRestOptions)
	}

	if needDomainConfigUpdate {
		_, err := e.client.UpdateDomainConfigWithContext(ctx, &input)
		if err != nil {
			return managed.ExternalUpdate{}, err
		}

	}

	if !*obj.DomainStatus.Processing && !*obj.DomainStatus.UpgradeProcessing && aws.StringValue(cr.Spec.ForProvider.EngineVersion) != aws.StringValue(obj.DomainStatus.EngineVersion) {
		// We cant update when processing or already updating so we dont update now
		upgradeInput := &svcsdk.UpgradeDomainInput{
			DomainName:    &domainName,
			TargetVersion: cr.Spec.ForProvider.EngineVersion,
		}
		if overrideMainResponseVersion, ok := cr.Spec.ForProvider.AdvancedOptions["override_main_response_version"]; ok {
			upgradeInput.AdvancedOptions = map[string]*string{
				"override_main_response_version": overrideMainResponseVersion,
			}
		}

		_, err := e.client.UpgradeDomainWithContext(ctx, upgradeInput)
		if err != nil {
			return managed.ExternalUpdate{}, err
		}

	}

	return managed.ExternalUpdate{}, nil
}

func isMapUpToDate(wanted, current map[string]*string, ignore []string) bool {
	if len(wanted) != len(current) {
		return false
	}
out:
	for key, value := range wanted {
		for _, ign := range ignore {
			if ign == key {
				continue out
			}
		}
		if aws.StringValue(current[key]) != aws.StringValue(value) {
			return false
		}
	}
	return true
}

func isAdvancedSecurityOptionsUpToDate(wanted *svcapitypes.AdvancedSecurityOptionsInput, current *svcsdk.AdvancedSecurityOptions) bool { //nolint:gocyclo
	if wanted != nil {
		if current == nil {
			return false
		}
		if aws.BoolValue(wanted.AnonymousAuthEnabled) != aws.BoolValue(current.AnonymousAuthEnabled) {
			return false
		}
		if aws.BoolValue(wanted.Enabled) != aws.BoolValue(current.Enabled) {
			return false
		}
		if aws.BoolValue(wanted.InternalUserDatabaseEnabled) != aws.BoolValue(current.InternalUserDatabaseEnabled) {
			return false
		}
		if wanted.SAMLOptions != nil {
			if current.SAMLOptions == nil {
				return false
			}
			if aws.BoolValue(wanted.SAMLOptions.Enabled) != aws.BoolValue(current.SAMLOptions.Enabled) {
				return false
			}
			if aws.StringValue(current.SAMLOptions.RolesKey) != aws.StringValue(wanted.SAMLOptions.RolesKey) {
				return false
			}
			if aws.Int64Value(current.SAMLOptions.SessionTimeoutMinutes) != aws.Int64Value(wanted.SAMLOptions.SessionTimeoutMinutes) {
				return false
			}
			if aws.StringValue(current.SAMLOptions.SubjectKey) != aws.StringValue(wanted.SAMLOptions.SubjectKey) {
				return false
			}
			if wanted.SAMLOptions.IDp != nil {
				if current.SAMLOptions.Idp == nil {
					return false
				}
				if aws.StringValue(wanted.SAMLOptions.IDp.EntityID) != aws.StringValue(current.SAMLOptions.Idp.EntityId) {
					return false
				}
				if aws.StringValue(wanted.SAMLOptions.IDp.MetadataContent) != aws.StringValue(current.SAMLOptions.Idp.MetadataContent) {
					return false
				}
			} else if current.SAMLOptions != nil {
				return false
			}

		} else if current.SAMLOptions != nil {
			return false
		}
	} else if current != nil {
		return false
	}
	return true
}

func isClusterConfigUpToDate(wanted *svcapitypes.ClusterConfig, current *svcsdk.ClusterConfig) bool { //nolint:gocyclo
	if wanted != nil {
		if current == nil {
			return false
		}
		if wanted.ColdStorageOptions != nil {
			if current.ColdStorageOptions == nil {
				return false
			}
			if aws.BoolValue(wanted.ColdStorageOptions.Enabled) != aws.BoolValue(current.ColdStorageOptions.Enabled) {
				return false
			}
		} else if current.ColdStorageOptions != nil {
			return false
		}
		if aws.Int64Value(wanted.DedicatedMasterCount) != aws.Int64Value(current.DedicatedMasterCount) {
			return false
		}
		if aws.BoolValue(wanted.DedicatedMasterEnabled) != aws.BoolValue(current.DedicatedMasterEnabled) {
			return false
		}
		if aws.StringValue(wanted.DedicatedMasterType) != aws.StringValue(current.DedicatedMasterType) {
			return false
		}
		if aws.Int64Value(wanted.InstanceCount) != aws.Int64Value(current.InstanceCount) {
			return false
		}
		if aws.StringValue(wanted.InstanceType) != aws.StringValue(current.InstanceType) {
			return false
		}
		if aws.Int64Value(wanted.WarmCount) != aws.Int64Value(current.WarmCount) {
			return false
		}
		if aws.BoolValue(wanted.WarmEnabled) != aws.BoolValue(current.WarmEnabled) {
			return false
		}
		if aws.StringValue(wanted.WarmType) != aws.StringValue(current.WarmType) {
			return false
		}
		if aws.BoolValue(wanted.ZoneAwarenessEnabled) != aws.BoolValue(current.ZoneAwarenessEnabled) {
			return false
		}
		if wanted.ZoneAwarenessConfig != nil {
			if current.ZoneAwarenessConfig == nil {
				return false
			}
			if aws.Int64Value(wanted.ZoneAwarenessConfig.AvailabilityZoneCount) != aws.Int64Value(current.ZoneAwarenessConfig.AvailabilityZoneCount) {
				return false
			}
		} else if current.ZoneAwarenessConfig != nil {
			return false
		}
	} else if current != nil {
		return false
	}
	return true
}

func isAutoTuneOptionsUpToDate(wanted *svcapitypes.AutoTuneOptionsInput, current *svcsdk.AutoTuneOptionsOutput_) bool {
	if wanted != nil {
		if current == nil {
			return false
		}
		if wanted.DesiredState != nil {
			if aws.StringValue(wanted.DesiredState) != aws.StringValue(current.State) {
				return false
			}
		}
	} else if current != nil {
		return false
	}

	return true
}

func isCognitoOptionsUpToDate(wanted *svcapitypes.CognitoOptions, current *svcsdk.CognitoOptions) bool {
	if wanted != nil {
		if current == nil {
			return false
		}
		if aws.BoolValue(wanted.Enabled) != aws.BoolValue(current.Enabled) {
			return false
		}
		if aws.StringValue(wanted.IdentityPoolID) != aws.StringValue(current.IdentityPoolId) {
			return false
		}
		if aws.StringValue(wanted.RoleARN) != aws.StringValue(current.RoleArn) {
			return false
		}
		if aws.StringValue(wanted.UserPoolID) != aws.StringValue(current.UserPoolId) {
			return false
		}
	} else if current != nil {
		return false
	}
	return true
}

func isDomainEndpointOptionsUpToDate(wanted *svcapitypes.DomainEndpointOptions, current *svcsdk.DomainEndpointOptions) bool {
	if wanted != nil {
		if current == nil {
			return false
		}
		if aws.StringValue(wanted.CustomEndpoint) != aws.StringValue(current.CustomEndpoint) {
			return false
		}
		if aws.StringValue(wanted.CustomEndpointCertificateARN) != aws.StringValue(current.CustomEndpointCertificateArn) {
			return false
		}
		if aws.BoolValue(wanted.CustomEndpointEnabled) != aws.BoolValue(current.CustomEndpointEnabled) {
			return false
		}
		if aws.BoolValue(wanted.EnforceHTTPS) != aws.BoolValue(current.EnforceHTTPS) {
			return false
		}
		if aws.StringValue(wanted.TLSSecurityPolicy) != aws.StringValue(current.TLSSecurityPolicy) {
			return false
		}
	} else if current != nil {
		return false
	}
	return true
}

func generateAdvancedSecurityOptionsInput(wanted *svcapitypes.AdvancedSecurityOptionsInput) *svcsdk.AdvancedSecurityOptionsInput_ {
	if wanted != nil {
		output := &svcsdk.AdvancedSecurityOptionsInput_{
			AnonymousAuthEnabled:        wanted.AnonymousAuthEnabled,
			Enabled:                     wanted.Enabled,
			InternalUserDatabaseEnabled: wanted.InternalUserDatabaseEnabled,
		}

		if wanted.MasterUserOptions != nil {
			output.MasterUserOptions = &svcsdk.MasterUserOptions{
				MasterUserARN:      wanted.MasterUserOptions.MasterUserARN,
				MasterUserName:     wanted.MasterUserOptions.MasterUserName,
				MasterUserPassword: wanted.MasterUserOptions.MasterUserPassword,
			}
		}

		if wanted.SAMLOptions != nil {
			output.SAMLOptions = &svcsdk.SAMLOptionsInput_{
				Enabled:               wanted.SAMLOptions.Enabled,
				MasterBackendRole:     wanted.SAMLOptions.MasterBackendRole,
				MasterUserName:        wanted.SAMLOptions.MasterUserName,
				RolesKey:              wanted.SAMLOptions.RolesKey,
				SessionTimeoutMinutes: wanted.SAMLOptions.SessionTimeoutMinutes,
				SubjectKey:            wanted.SAMLOptions.SubjectKey,
			}

			if wanted.SAMLOptions.IDp != nil {
				output.SAMLOptions.Idp = &svcsdk.SAMLIdp{
					EntityId:        wanted.SAMLOptions.IDp.EntityID,
					MetadataContent: wanted.SAMLOptions.IDp.MetadataContent,
				}
			}
		}
		return output
	}
	return nil
}

func isEbsOptionsUpToDate(wanted *svcapitypes.EBSOptions, current *svcsdk.EBSOptions) bool {
	if wanted != nil {
		if current == nil {
			return false
		}
		if aws.BoolValue(wanted.EBSEnabled) != aws.BoolValue(current.EBSEnabled) {
			return false
		}
		if aws.Int64Value(wanted.IOPS) != aws.Int64Value(current.Iops) {
			return false
		}
		if aws.Int64Value(wanted.VolumeSize) != aws.Int64Value(current.VolumeSize) {
			return false
		}
		if aws.StringValue(wanted.VolumeType) != aws.StringValue(current.VolumeType) {
			return false
		}
	} else if current != nil {
		return false
	}
	return true
}

func isLogPublishingOptionsUpToDate(wanted map[string]*svcapitypes.LogPublishingOption, current map[string]*svcsdk.LogPublishingOption) bool {
	if wanted != nil {
		if current == nil {
			return false
		}
		if len(wanted) != len(current) {
			return false
		}
		for key, value := range wanted {
			if aws.StringValue(value.CloudWatchLogsLogGroupARN) != aws.StringValue(current[key].CloudWatchLogsLogGroupArn) {
				return false
			}
			if aws.BoolValue(value.Enabled) != aws.BoolValue(current[key].Enabled) {
				return false
			}
		}
	} else if current != nil {
		return false
	}
	return true
}

func isNodeToNodeEncryptionOptionsUpToDate(wanted *svcapitypes.NodeToNodeEncryptionOptions, current *svcsdk.NodeToNodeEncryptionOptions) bool {
	if wanted != nil {
		if current == nil {
			return false
		}
		if aws.BoolValue(wanted.Enabled) != aws.BoolValue(current.Enabled) {
			return false
		}
	} else if current != nil {
		return false
	}
	return true
}

func isSnapshotOptionsUpToDate(wanted *svcapitypes.SnapshotOptions, current *svcsdk.SnapshotOptions) bool {
	if wanted != nil {
		if current == nil {
			return false
		}
		if aws.Int64Value(wanted.AutomatedSnapshotStartHour) != aws.Int64Value(current.AutomatedSnapshotStartHour) {
			return false
		}
	} else if current != nil {
		return false
	}
	return true
}

func isVpcOptionsUpToDate(wanted *svcapitypes.CustomVPCDerivedInfo, current *svcsdk.VPCDerivedInfo) bool { //nolint:gocyclo
	if wanted != nil {
		if current == nil {
			return false
		}
		if len(wanted.SecurityGroupIDs) != len(current.SecurityGroupIds) {
			return false
		}
		for _, objValue := range wanted.SecurityGroupIDs {
			found := false
			for _, valueOut := range current.SecurityGroupIds {
				if aws.StringValue(objValue) == aws.StringValue(valueOut) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		if len(wanted.SubnetIDs) != len(current.SubnetIds) {
			return false
		}
		for _, objValue := range wanted.SubnetIDs {
			found := false
			for _, valueOut := range current.SubnetIds {
				if aws.StringValue(objValue) == aws.StringValue(valueOut) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	} else if current != nil {
		return false
	}
	return true
}

func isEncryptionAtRestUpToDate(wanted *svcapitypes.CustomEncryptionAtRestOptions, current *svcsdk.EncryptionAtRestOptions) bool {
	if wanted != nil {
		if current == nil {
			return false
		}
		if aws.BoolValue(wanted.Enabled) != aws.BoolValue(current.Enabled) {
			return false
		}
		spriltedKey := strings.Split(aws.StringValue(current.KmsKeyId), "/")
		currentShortKeyID := ""
		if len(spriltedKey) > 0 {
			currentShortKeyID = spriltedKey[len(spriltedKey)-1]
		}

		if aws.StringValue(wanted.KMSKeyID) != currentShortKeyID {
			return false
		}

	} else if current != nil {
		return false
	}

	return true
}

func generateClusterConfig(wanted *svcapitypes.ClusterConfig) *svcsdk.ClusterConfig {
	if wanted != nil {
		output := &svcsdk.ClusterConfig{
			DedicatedMasterCount:   wanted.DedicatedMasterCount,
			DedicatedMasterEnabled: wanted.DedicatedMasterEnabled,
			DedicatedMasterType:    wanted.DedicatedMasterType,
			InstanceCount:          wanted.InstanceCount,
			InstanceType:           wanted.InstanceType,
			WarmCount:              wanted.WarmCount,
			WarmEnabled:            wanted.WarmEnabled,
			WarmType:               wanted.WarmType,
			ZoneAwarenessEnabled:   wanted.ZoneAwarenessEnabled,
		}

		if wanted.ColdStorageOptions != nil {
			output.ColdStorageOptions = &svcsdk.ColdStorageOptions{
				Enabled: wanted.ColdStorageOptions.Enabled,
			}
		}
		if wanted.ZoneAwarenessConfig != nil {
			output.ZoneAwarenessConfig = &svcsdk.ZoneAwarenessConfig{
				AvailabilityZoneCount: wanted.ZoneAwarenessConfig.AvailabilityZoneCount,
			}
		}
		return output

	}
	return nil
}

func generateAutotuneOptionsInput(wanted *svcapitypes.AutoTuneOptionsInput) *svcsdk.AutoTuneOptions {
	if wanted != nil {
		output := &svcsdk.AutoTuneOptions{
			DesiredState: wanted.DesiredState,
		}
		if wanted.MaintenanceSchedules != nil {
			msList := []*svcsdk.AutoTuneMaintenanceSchedule{}
			for _, ms := range wanted.MaintenanceSchedules {
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
					msNew.StartAt = ptr.To(ms.StartAt.Time)
				}
				msList = append(msList, msNew)
			}
			output.MaintenanceSchedules = msList
		}
		return output
	}
	return nil
}

func generateCognitoOptions(wanted *svcapitypes.CognitoOptions) *svcsdk.CognitoOptions {
	if wanted != nil {
		output := &svcsdk.CognitoOptions{
			Enabled:        wanted.Enabled,
			IdentityPoolId: wanted.IdentityPoolID,
			RoleArn:        wanted.RoleARN,
			UserPoolId:     wanted.UserPoolID,
		}
		return output
	}
	return nil
}

func generateDomainEndpointsOptions(wanted *svcapitypes.DomainEndpointOptions) *svcsdk.DomainEndpointOptions {
	if wanted != nil {
		output := &svcsdk.DomainEndpointOptions{
			CustomEndpoint:               wanted.CustomEndpoint,
			CustomEndpointCertificateArn: wanted.CustomEndpointCertificateARN,
			CustomEndpointEnabled:        wanted.CustomEndpointEnabled,
			EnforceHTTPS:                 wanted.EnforceHTTPS,
			TLSSecurityPolicy:            wanted.TLSSecurityPolicy,
		}
		return output
	}
	return nil
}

func generateEbsOptions(wanted *svcapitypes.EBSOptions) *svcsdk.EBSOptions {
	if wanted != nil {

		output := &svcsdk.EBSOptions{
			EBSEnabled: wanted.EBSEnabled,
			Iops:       wanted.IOPS,
			VolumeSize: wanted.VolumeSize,
			VolumeType: wanted.VolumeType,
		}
		return output
	}
	return nil
}

func generateLogPublishingOptions(wanted map[string]*svcapitypes.LogPublishingOption) map[string]*svcsdk.LogPublishingOption {
	if wanted != nil {
		output := map[string]*svcsdk.LogPublishingOption{}

		for key, value := range wanted {
			output[key] = &svcsdk.LogPublishingOption{
				CloudWatchLogsLogGroupArn: value.CloudWatchLogsLogGroupARN,
				Enabled:                   value.Enabled,
			}
		}
		return output
	}
	return nil
}

func generateNodeToNodeEncryptionOptions(wanted *svcapitypes.NodeToNodeEncryptionOptions) *svcsdk.NodeToNodeEncryptionOptions {
	if wanted != nil {
		output := &svcsdk.NodeToNodeEncryptionOptions{
			Enabled: wanted.Enabled,
		}
		return output
	}
	return nil
}

func generateSnapshotOptions(wanted *svcapitypes.SnapshotOptions) *svcsdk.SnapshotOptions {
	if wanted != nil {
		output := &svcsdk.SnapshotOptions{
			AutomatedSnapshotStartHour: wanted.AutomatedSnapshotStartHour,
		}
		return output
	}
	return nil
}

func generateVpcOptions(wanted *svcapitypes.CustomVPCDerivedInfo) *svcsdk.VPCOptions {
	if wanted != nil {
		output := &svcsdk.VPCOptions{
			SecurityGroupIds: wanted.SecurityGroupIDs,
			SubnetIds:        wanted.SubnetIDs,
		}
		return output
	}
	return nil
}

func generateEncryptionAtRestOptions(wanted *svcapitypes.CustomEncryptionAtRestOptions) *svcsdk.EncryptionAtRestOptions {
	if wanted != nil {
		output := &svcsdk.EncryptionAtRestOptions{
			Enabled:  wanted.Enabled,
			KmsKeyId: wanted.KMSKeyID,
		}
		return output
	}
	return nil
}
