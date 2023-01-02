package domain

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/opensearchservice"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
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
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
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
