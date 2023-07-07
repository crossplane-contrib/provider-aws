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

package cluster

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/pkg/errors"

	svcsdk "github.com/aws/aws-sdk-go/service/kafka"
	"github.com/aws/aws-sdk-go/service/kafka/kafkaiface"
	ctrl "sigs.k8s.io/controller-runtime"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	cpresource "github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/kafka/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	awsclients "github.com/crossplane-contrib/provider-aws/pkg/clients"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
)

const (
	errUpdateBrokerType           = "cannot update BrokerType of Cluster in AWS"
	errUpdateBrokerStorage        = "cannot update BrokerStorage of Cluster in AWS"
	errUpdateBrokerCount          = "cannot update BrokerCount of Cluster in AWS"
	errUpdateMonitoring           = "cannot update Monitoring of Cluster in AWS"
	errUpdateClusterConfiguration = "cannot update ClusterConfiguration of Cluster in AWS"
	errUpdateClusterKafkaVersion  = "cannot update ClusterKafkaVersion of Cluster in AWS"
	errUpdateSecurity             = "cannot update Security of Cluster in AWS"
	errUpdateTags                 = "cannot update Tags of Cluster in AWS"
	errStateForUpdate             = "cannot update cluster if not in status ACTIVE"
	stateActive                   = "ACTIVE"
)

// SetupCluster adds a controller that reconciles Cluster.
func SetupCluster(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.ClusterGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preDelete = preDelete
			e.postDelete = postDelete
			e.preCreate = preCreate
			e.postCreate = postCreate
			e.lateInitialize = LateInitialize
			u := &updater{client: e.client}
			e.update = u.update
			e.isUpToDate = isUpToDate
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
		For(&svcapitypes.Cluster{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.ClusterGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithInitializers(),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

func preDelete(_ context.Context, cr *svcapitypes.Cluster, obj *svcsdk.DeleteClusterInput) (bool, error) {
	obj.ClusterArn = awsclients.String(meta.GetExternalName(cr))
	return false, nil
}

func postDelete(_ context.Context, cr *svcapitypes.Cluster, obj *svcsdk.DeleteClusterOutput, err error) error {
	if err != nil {
		if strings.Contains(err.Error(), svcsdk.ErrCodeBadRequestException) {
			// skip: failed to delete Cluster: BadRequestException: You can't delete cluster in DELETING state.
			return nil
		}
		return err
	}
	return err
}

func preObserve(_ context.Context, cr *svcapitypes.Cluster, obj *svcsdk.DescribeClusterInput) error {
	obj.ClusterArn = awsclients.String(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.Cluster, obj *svcsdk.DescribeClusterOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	switch awsclients.StringValue(obj.ClusterInfo.State) {
	case string(svcapitypes.ClusterState_ACTIVE):
		cr.SetConditions(xpv1.Available())
	case string(svcapitypes.ClusterState_CREATING):
		cr.SetConditions(xpv1.Creating())
	case string(svcapitypes.ClusterState_FAILED), string(svcapitypes.ClusterState_MAINTENANCE), string(svcapitypes.ClusterState_UPDATING):
		cr.SetConditions(xpv1.Unavailable())
	case string(svcapitypes.ClusterState_DELETING):
		cr.SetConditions(xpv1.Deleting())
	}

	obs.ConnectionDetails = managed.ConnectionDetails{
		// see: https://docs.aws.amazon.com/msk/latest/developerguide/client-access.html
		// no endpoint informations available in DescribeClusterOutput only endpoints for zookeeperPlain/Tls
		"zookeeperEndpointPlain": []byte(awsclients.StringValue(obj.ClusterInfo.ZookeeperConnectString)),
		"zookeeperEndpointTls":   []byte(awsclients.StringValue(obj.ClusterInfo.ZookeeperConnectStringTls)),
		"clusterEndpointPlain":   []byte(strings.ReplaceAll(awsclients.StringValue(obj.ClusterInfo.ZookeeperConnectString), "2181", "9092")),
		"clusterEndpointTls":     []byte(strings.ReplaceAll(awsclients.StringValue(obj.ClusterInfo.ZookeeperConnectString), "2181", "9094")),
		"clusterEndpointIAM":     []byte(strings.ReplaceAll(awsclients.StringValue(obj.ClusterInfo.ZookeeperConnectString), "2181", "9098")),
	}

	return obs, nil
}

func preCreate(_ context.Context, cr *svcapitypes.Cluster, obj *svcsdk.CreateClusterInput) error {
	obj.BrokerNodeGroupInfo = &svcsdk.BrokerNodeGroupInfo{
		ClientSubnets:  cr.Spec.ForProvider.CustomBrokerNodeGroupInfo.ClientSubnets,
		InstanceType:   cr.Spec.ForProvider.CustomBrokerNodeGroupInfo.InstanceType,
		SecurityGroups: cr.Spec.ForProvider.CustomBrokerNodeGroupInfo.SecurityGroups,
		StorageInfo: &svcsdk.StorageInfo{
			EbsStorageInfo: &svcsdk.EBSStorageInfo{
				VolumeSize: cr.Spec.ForProvider.CustomBrokerNodeGroupInfo.StorageInfo.EBSStorageInfo.VolumeSize,
			},
		},
	}
	if cr.Spec.ForProvider.CustomConfigurationInfo != nil {
		obj.ConfigurationInfo = &svcsdk.ConfigurationInfo{
			Arn:      cr.Spec.ForProvider.CustomConfigurationInfo.ARN,
			Revision: cr.Spec.ForProvider.CustomConfigurationInfo.Revision,
		}
	}
	return nil
}

func postCreate(_ context.Context, cr *svcapitypes.Cluster, obj *svcsdk.CreateClusterOutput, _ managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, awsclients.StringValue(obj.ClusterArn))
	return managed.ExternalCreation{}, nil
}

// LateInitialize fills the empty fields in *svcapitypes.ClusterParameters with
// the values seen in svcsdk.DescribeClusterOutput.
// nolint:gocyclo
func LateInitialize(cr *svcapitypes.ClusterParameters, obj *svcsdk.DescribeClusterOutput) error {

	if cr.EnhancedMonitoring == nil && obj.ClusterInfo.EnhancedMonitoring != nil {
		cr.EnhancedMonitoring = awsclients.LateInitializeStringPtr(cr.EnhancedMonitoring, obj.ClusterInfo.EnhancedMonitoring)
	}

	if cr.CustomBrokerNodeGroupInfo.SecurityGroups == nil && obj.ClusterInfo.BrokerNodeGroupInfo.SecurityGroups != nil {
		cr.CustomBrokerNodeGroupInfo.SecurityGroups = obj.ClusterInfo.BrokerNodeGroupInfo.SecurityGroups
	}

	if cr.EncryptionInfo == nil && obj.ClusterInfo.EncryptionInfo != nil {
		cr.EncryptionInfo = &svcapitypes.EncryptionInfo{
			EncryptionAtRest: &svcapitypes.EncryptionAtRest{
				DataVolumeKMSKeyID: obj.ClusterInfo.EncryptionInfo.EncryptionAtRest.DataVolumeKMSKeyId,
			},
			EncryptionInTransit: &svcapitypes.EncryptionInTransit{
				ClientBroker: obj.ClusterInfo.EncryptionInfo.EncryptionInTransit.ClientBroker,
				InCluster:    obj.ClusterInfo.EncryptionInfo.EncryptionInTransit.InCluster,
			},
		}
	}

	if cr.OpenMonitoring == nil && obj.ClusterInfo.OpenMonitoring != nil {
		cr.OpenMonitoring = &svcapitypes.OpenMonitoringInfo{
			Prometheus: &svcapitypes.PrometheusInfo{
				JmxExporter: &svcapitypes.JmxExporterInfo{
					EnabledInBroker: obj.ClusterInfo.OpenMonitoring.Prometheus.JmxExporter.EnabledInBroker,
				},
				NodeExporter: &svcapitypes.NodeExporterInfo{
					EnabledInBroker: obj.ClusterInfo.OpenMonitoring.Prometheus.NodeExporter.EnabledInBroker,
				},
			},
		}
	}

	return nil
}

func isUpToDate(wanted *svcapitypes.Cluster, current *svcsdk.DescribeClusterOutput) (bool, error) { // nolint:gocyclo
	forProvider := wanted.Spec.ForProvider
	clusterInfo := current.ClusterInfo

	switch {
	// A cluster can not be updated while not in active status, therefore we consider the cluster as up to date
	case aws.StringValue(clusterInfo.State) != stateActive:
		return true, nil
	case !isInstanceTypeUpToDate(forProvider.CustomBrokerNodeGroupInfo, clusterInfo.BrokerNodeGroupInfo),
		!isStorageInfoUpToDate(forProvider.CustomBrokerNodeGroupInfo, clusterInfo.BrokerNodeGroupInfo),
		!isNumberOfBrokerNodesUpToDate(forProvider, clusterInfo),
		!isEnhancedMonitoringUpToDate(forProvider, clusterInfo),
		!isLoggingInfoUpToDate(forProvider.LoggingInfo, clusterInfo.LoggingInfo),
		!isOpenMonitoringInfoUpToDate(forProvider.OpenMonitoring, clusterInfo.OpenMonitoring),
		!isCustomConfigurationInfoUpToDate(forProvider.CustomConfigurationInfo, clusterInfo.CurrentBrokerSoftwareInfo),
		!isKafkaVersionUpToDate(forProvider.KafkaVersion, clusterInfo.CurrentBrokerSoftwareInfo),
		!isEncryptionInfoUpToDate(forProvider.EncryptionInfo, clusterInfo.EncryptionInfo),
		!isClientAuthenticationUpToDate(forProvider.ClientAuthentication, clusterInfo.ClientAuthentication),
		!isTagsUpToDate(forProvider.Tags, clusterInfo.Tags):
		return false, nil
	}
	return true, nil
}

func isInstanceTypeUpToDate(wanted *svcapitypes.CustomBrokerNodeGroupInfo, current *svcsdk.BrokerNodeGroupInfo) bool {

	if wanted != nil {
		if current == nil {
			return false
		}
		if aws.StringValue(wanted.InstanceType) != aws.StringValue(current.InstanceType) {
			return false
		}
	} else if current != nil {
		return false
	}
	return true
}

func isStorageInfoUpToDate(wanted *svcapitypes.CustomBrokerNodeGroupInfo, current *svcsdk.BrokerNodeGroupInfo) bool { // nolint:gocyclo
	if wanted != nil {
		if current == nil {
			return false
		}

		if wanted.StorageInfo != nil {
			if current.StorageInfo == nil {
				return false
			}
			if wanted.StorageInfo.EBSStorageInfo != nil {
				if current.StorageInfo.EbsStorageInfo == nil {
					return false
				}

				if aws.Int64Value(wanted.StorageInfo.EBSStorageInfo.VolumeSize) != aws.Int64Value(current.StorageInfo.EbsStorageInfo.VolumeSize) {
					return false
				}
			} else if current.StorageInfo.EbsStorageInfo != nil {
				return false
			}
		} else if current.StorageInfo != nil {
			return false
		}
	} else if current != nil {
		return false
	}
	return true
}

func isNumberOfBrokerNodesUpToDate(wanted svcapitypes.ClusterParameters, current *svcsdk.ClusterInfo) bool {
	if current == nil {
		return false
	}
	if aws.Int64Value(wanted.NumberOfBrokerNodes) != aws.Int64Value(current.NumberOfBrokerNodes) {
		return false
	}
	return true
}

func isEnhancedMonitoringUpToDate(wanted svcapitypes.ClusterParameters, current *svcsdk.ClusterInfo) bool {
	if current == nil {
		return false
	}
	if aws.StringValue(wanted.EnhancedMonitoring) != aws.StringValue(current.EnhancedMonitoring) {
		return false
	}
	return true
}

func isLoggingInfoUpToDate(wanted *svcapitypes.LoggingInfo, current *svcsdk.LoggingInfo) bool { // nolint:gocyclo

	if wanted != nil {
		if current == nil {
			return false
		}

		if wanted.BrokerLogs != nil {
			if current.BrokerLogs == nil {
				return false
			}

			if wanted.BrokerLogs.CloudWatchLogs != nil {
				if current.BrokerLogs.CloudWatchLogs == nil {
					return false
				}
				if aws.BoolValue(wanted.BrokerLogs.CloudWatchLogs.Enabled) != aws.BoolValue(current.BrokerLogs.CloudWatchLogs.Enabled) {
					return false
				}
				if aws.StringValue(wanted.BrokerLogs.CloudWatchLogs.LogGroup) != aws.StringValue(current.BrokerLogs.CloudWatchLogs.LogGroup) {
					return false
				}
			} else if current.BrokerLogs.CloudWatchLogs != nil {
				return false
			}

			if wanted.BrokerLogs.Firehose != nil {
				if current.BrokerLogs.Firehose == nil {
					return false
				}
				if aws.StringValue(wanted.BrokerLogs.Firehose.DeliveryStream) != aws.StringValue(current.BrokerLogs.Firehose.DeliveryStream) {
					return false
				}
				if aws.BoolValue(wanted.BrokerLogs.Firehose.Enabled) != aws.BoolValue(current.BrokerLogs.Firehose.Enabled) {
					return false
				}
			} else if current.BrokerLogs.Firehose != nil {
				return false
			}

			if wanted.BrokerLogs.S3 != nil {
				if current.BrokerLogs.S3 == nil {
					return false
				}

				if aws.StringValue(wanted.BrokerLogs.S3.Bucket) != aws.StringValue(current.BrokerLogs.S3.Bucket) {
					return false
				}

				if aws.BoolValue(wanted.BrokerLogs.S3.Enabled) != aws.BoolValue(current.BrokerLogs.S3.Enabled) {
					return false
				}

				if aws.StringValue(wanted.BrokerLogs.S3.Prefix) != aws.StringValue(current.BrokerLogs.S3.Prefix) {
					return false
				}
			} else if current.BrokerLogs.S3 != nil {
				return false
			}
		} else if current.BrokerLogs != nil {
			return false
		}
	} else if current != nil {
		return false
	}

	return true
}

func isOpenMonitoringInfoUpToDate(wanted *svcapitypes.OpenMonitoringInfo, current *svcsdk.OpenMonitoring) bool { // nolint:gocyclo
	if wanted != nil {
		if current == nil {
			return false
		}

		if wanted.Prometheus != nil {
			if current.Prometheus == nil {
				return false
			}

			if wanted.Prometheus.JmxExporter != nil {
				if current.Prometheus.JmxExporter == nil {
					return false
				}

				if aws.BoolValue(wanted.Prometheus.JmxExporter.EnabledInBroker) != aws.BoolValue(current.Prometheus.JmxExporter.EnabledInBroker) {
					return false
				}

			} else if current.Prometheus.JmxExporter != nil {
				return false
			}

			if wanted.Prometheus.NodeExporter != nil {
				if current.Prometheus.NodeExporter == nil {
					return false
				}
				if aws.BoolValue(wanted.Prometheus.NodeExporter.EnabledInBroker) != aws.BoolValue(current.Prometheus.NodeExporter.EnabledInBroker) {
					return false
				}
			} else if current.Prometheus.NodeExporter != nil {
				return false
			}
		} else if current.Prometheus != nil {
			return false
		}
	} else if current != nil {
		return false
	}
	return true
}

func isCustomConfigurationInfoUpToDate(wanted *svcapitypes.CustomConfigurationInfo, current *svcsdk.BrokerSoftwareInfo) bool {
	if wanted != nil {
		if current == nil {
			return false
		}

		if aws.StringValue(wanted.ARN) != aws.StringValue(current.ConfigurationArn) {
			return false
		}
		if aws.Int64Value(wanted.Revision) != aws.Int64Value((current.ConfigurationRevision)) {
			return false
		}
	} else if current != nil {
		return false
	}

	return true
}

func isKafkaVersionUpToDate(wanted *string, current *svcsdk.BrokerSoftwareInfo) bool {
	if wanted != nil {
		if current == nil {
			return false
		}
		if aws.StringValue(wanted) != aws.StringValue(current.KafkaVersion) {
			return false
		}
	} else if current != nil {
		return false
	}
	return true
}

func isTagsUpToDate(wanted map[string]*string, current map[string]*string) bool {
	for key, value := range wanted {
		if aws.StringValue(current[key]) != aws.StringValue(value) {
			return false
		}
	}
	return true
}

func isEncryptionInfoUpToDate(wanted *svcapitypes.EncryptionInfo, current *svcsdk.EncryptionInfo) bool { // nolint:gocyclo
	if wanted != nil {
		if current == nil {
			return false
		}
		if wanted.EncryptionAtRest != nil {
			if current.EncryptionAtRest == nil {
				return false
			}
			if aws.StringValue(wanted.EncryptionAtRest.DataVolumeKMSKeyID) != aws.StringValue(current.EncryptionAtRest.DataVolumeKMSKeyId) {
				return false
			}
		} else if current.EncryptionAtRest != nil {
			return false
		}
		if wanted.EncryptionInTransit != nil {
			if current.EncryptionInTransit == nil {
				return false
			}
			if aws.StringValue(wanted.EncryptionInTransit.ClientBroker) != aws.StringValue(current.EncryptionInTransit.ClientBroker) {
				return false
			}
			if aws.BoolValue(wanted.EncryptionInTransit.InCluster) != aws.BoolValue(current.EncryptionInTransit.InCluster) {
				return false
			}
		} else if current.EncryptionInTransit != nil {
			return false
		}
	} else if current != nil {
		return false
	}
	return true
}

func isClientAuthenticationUpToDate(wanted *svcapitypes.ClientAuthentication, current *svcsdk.ClientAuthentication) bool { // nolint:gocyclo
	if wanted != nil {
		if current == nil {
			return false
		}

		if wanted.SASL != nil {
			if current.Sasl == nil {
				return false
			}
			if wanted.SASL.IAM != nil {
				if current.Sasl.Iam == nil {
					return false
				}
				if aws.BoolValue(wanted.SASL.IAM.Enabled) != aws.BoolValue(current.Sasl.Iam.Enabled) {
					return false
				}
			} else if current.Sasl.Iam != nil {
				return false
			}
			if wanted.SASL.SCRAM != nil {
				if current.Sasl.Scram == nil {
					return false
				}
				if aws.BoolValue(wanted.SASL.SCRAM.Enabled) != aws.BoolValue(current.Sasl.Scram.Enabled) {
					return false
				}
			} else if current.Sasl.Scram != nil {
				return false
			}
		} else if current.Sasl != nil {
			return false
		}

		if wanted.TLS != nil {
			if current.Tls == nil {
				return false
			}
			for _, wantedValue := range wanted.TLS.CertificateAuthorityARNList {
				found := false
				for _, currentValue := range current.Tls.CertificateAuthorityArnList {
					if aws.StringValue(wantedValue) == aws.StringValue(currentValue) {
						found = true
						break
					}
				}
				if !found {
					return false
				}
			}
			if aws.BoolValue(wanted.TLS.Enabled) != aws.BoolValue(current.Tls.Enabled) {
				return false
			}
		} else if current.Tls != nil {
			return false
		}

		if wanted.Unauthenticated != nil {
			if current.Unauthenticated == nil {
				return false
			}
			if aws.BoolValue(wanted.Unauthenticated.Enabled) != aws.BoolValue(current.Unauthenticated.Enabled) {
				return false
			}
		} else if current.Unauthenticated != nil {
			return false
		}
	} else if current != nil {
		return false
	}

	return true
}

type updater struct {
	client kafkaiface.KafkaAPI
}

func (u *updater) update(ctx context.Context, mg cpresource.Managed) (managed.ExternalUpdate, error) { // nolint:gocyclo
	cr, ok := mg.(*svcapitypes.Cluster)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}
	if aws.StringValue(cr.Status.AtProvider.State) != stateActive {
		return managed.ExternalUpdate{}, errors.New(errStateForUpdate)
	}
	input := GenerateDescribeClusterInput(cr)

	if meta.GetExternalName(cr) != "" {
		obj, err := u.client.DescribeClusterWithContext(ctx, input)

		if err != nil {
			return managed.ExternalUpdate{}, awsclients.Wrap(err, errDescribe)
		}
		currentARN := meta.GetExternalName(cr)
		currentVersion := obj.ClusterInfo.CurrentVersion
		wanted := cr.Spec.ForProvider
		if !isInstanceTypeUpToDate(wanted.CustomBrokerNodeGroupInfo, obj.ClusterInfo.BrokerNodeGroupInfo) {
			if aws.StringValue(obj.ClusterInfo.State) != stateActive {
				return managed.ExternalUpdate{}, errors.New(errStateForUpdate)
			}
			// UpdateBrokerTypeWithContext
			input := &svcsdk.UpdateBrokerTypeInput{
				ClusterArn:         &currentARN,
				CurrentVersion:     currentVersion,
				TargetInstanceType: wanted.CustomBrokerNodeGroupInfo.InstanceType,
			}

			_, err := u.client.UpdateBrokerTypeWithContext(ctx, input)
			if err != nil {
				return managed.ExternalUpdate{}, awsclients.Wrap(err, errUpdateBrokerType)
			}
		}

		if !isStorageInfoUpToDate(wanted.CustomBrokerNodeGroupInfo, obj.ClusterInfo.BrokerNodeGroupInfo) {

			obj, err = u.client.DescribeClusterWithContext(ctx, input)
			currentVersion = obj.ClusterInfo.CurrentVersion
			if err != nil {
				return managed.ExternalUpdate{}, awsclients.Wrap(err, errDescribe)
			}

			if aws.StringValue(obj.ClusterInfo.State) != stateActive {
				return managed.ExternalUpdate{}, errors.New(errStateForUpdate)
			}
			input := &svcsdk.UpdateBrokerStorageInput{
				ClusterArn:     &currentARN,
				CurrentVersion: currentVersion,
				TargetBrokerEBSVolumeInfo: []*svcsdk.BrokerEBSVolumeInfo{{
					KafkaBrokerNodeId: aws.String("ALL"),
				},
				},
			}
			if wanted.CustomBrokerNodeGroupInfo != nil && wanted.CustomBrokerNodeGroupInfo.StorageInfo != nil && wanted.CustomBrokerNodeGroupInfo.StorageInfo.EBSStorageInfo != nil {
				if wanted.CustomBrokerNodeGroupInfo.StorageInfo.EBSStorageInfo.VolumeSize != nil {
					input.TargetBrokerEBSVolumeInfo[0].VolumeSizeGB = wanted.CustomBrokerNodeGroupInfo.StorageInfo.EBSStorageInfo.VolumeSize
				}

				if wanted.CustomBrokerNodeGroupInfo.StorageInfo.EBSStorageInfo.ProvisionedThroughput != nil {
					input.TargetBrokerEBSVolumeInfo[0].ProvisionedThroughput = &svcsdk.ProvisionedThroughput{
						Enabled:          wanted.CustomBrokerNodeGroupInfo.StorageInfo.EBSStorageInfo.ProvisionedThroughput.Enabled,
						VolumeThroughput: wanted.CustomBrokerNodeGroupInfo.StorageInfo.EBSStorageInfo.ProvisionedThroughput.VolumeThroughput,
					}
				}
			}

			_, err := u.client.UpdateBrokerStorageWithContext(ctx, input)
			if err != nil {
				return managed.ExternalUpdate{}, awsclients.Wrap(err, errUpdateBrokerStorage)
			}
		}

		if !isNumberOfBrokerNodesUpToDate(wanted, obj.ClusterInfo) {
			obj, err = u.client.DescribeClusterWithContext(ctx, input)
			currentVersion = obj.ClusterInfo.CurrentVersion
			if err != nil {
				return managed.ExternalUpdate{}, awsclients.Wrap(err, errDescribe)
			}
			if aws.StringValue(obj.ClusterInfo.State) != stateActive {
				return managed.ExternalUpdate{}, errors.New(errStateForUpdate)
			}
			input := &svcsdk.UpdateBrokerCountInput{
				ClusterArn:                &currentARN,
				CurrentVersion:            currentVersion,
				TargetNumberOfBrokerNodes: wanted.NumberOfBrokerNodes,
			}

			_, err := u.client.UpdateBrokerCountWithContext(ctx, input)
			if err != nil {
				return managed.ExternalUpdate{}, awsclients.Wrap(err, errUpdateBrokerCount)
			}
		}

		if !isEnhancedMonitoringUpToDate(wanted, obj.ClusterInfo) || !isLoggingInfoUpToDate(wanted.LoggingInfo, obj.ClusterInfo.LoggingInfo) || !isOpenMonitoringInfoUpToDate(wanted.OpenMonitoring, obj.ClusterInfo.OpenMonitoring) {
			obj, err = u.client.DescribeClusterWithContext(ctx, input)
			currentVersion = obj.ClusterInfo.CurrentVersion
			if err != nil {
				return managed.ExternalUpdate{}, awsclients.Wrap(err, errDescribe)
			}
			if aws.StringValue(obj.ClusterInfo.State) != stateActive {
				return managed.ExternalUpdate{}, errors.New(errStateForUpdate)
			}
			input := &svcsdk.UpdateMonitoringInput{
				ClusterArn:         &currentARN,
				CurrentVersion:     currentVersion,
				EnhancedMonitoring: wanted.EnhancedMonitoring,
				LoggingInfo:        generateLoggingInfoInput(wanted.LoggingInfo),
				OpenMonitoring:     generateOpenMonitorinInput(wanted.OpenMonitoring),
			}

			_, err := u.client.UpdateMonitoringWithContext(ctx, input)
			if err != nil {
				return managed.ExternalUpdate{}, awsclients.Wrap(err, errUpdateMonitoring)
			}

		}

		if !isCustomConfigurationInfoUpToDate(wanted.CustomConfigurationInfo, obj.ClusterInfo.CurrentBrokerSoftwareInfo) && isKafkaVersionUpToDate(wanted.KafkaVersion, obj.ClusterInfo.CurrentBrokerSoftwareInfo) {
			obj, err = u.client.DescribeClusterWithContext(ctx, input)
			currentVersion = obj.ClusterInfo.CurrentVersion
			if err != nil {
				return managed.ExternalUpdate{}, awsclients.Wrap(err, errDescribe)
			}
			if aws.StringValue(obj.ClusterInfo.State) != stateActive {
				return managed.ExternalUpdate{}, errors.New(errStateForUpdate)
			}
			input := &svcsdk.UpdateClusterConfigurationInput{
				ClusterArn:        &currentARN,
				CurrentVersion:    currentVersion,
				ConfigurationInfo: generateCustomConfigurationInfo(wanted.CustomConfigurationInfo),
			}

			_, err := u.client.UpdateClusterConfigurationWithContext(ctx, input)
			if err != nil {
				return managed.ExternalUpdate{}, awsclients.Wrap(err, errUpdateClusterConfiguration)
			}
		}

		if !isKafkaVersionUpToDate(wanted.KafkaVersion, obj.ClusterInfo.CurrentBrokerSoftwareInfo) {
			obj, err = u.client.DescribeClusterWithContext(ctx, input)
			currentVersion = obj.ClusterInfo.CurrentVersion
			if err != nil {
				return managed.ExternalUpdate{}, awsclients.Wrap(err, errDescribe)
			}
			if aws.StringValue(obj.ClusterInfo.State) != stateActive {
				return managed.ExternalUpdate{}, errors.New(errStateForUpdate)
			}
			input := &svcsdk.UpdateClusterKafkaVersionInput{
				ClusterArn:         &currentARN,
				CurrentVersion:     currentVersion,
				TargetKafkaVersion: wanted.KafkaVersion,
			}

			if !isCustomConfigurationInfoUpToDate(wanted.CustomConfigurationInfo, obj.ClusterInfo.CurrentBrokerSoftwareInfo) {
				input.ConfigurationInfo = generateCustomConfigurationInfo(wanted.CustomConfigurationInfo)
			}

			_, err := u.client.UpdateClusterKafkaVersionWithContext(ctx, input)
			if err != nil {
				return managed.ExternalUpdate{}, awsclients.Wrap(err, errUpdateClusterKafkaVersion)
			}
		}

		encryptionUpToDate := isEncryptionInfoUpToDate(wanted.EncryptionInfo, obj.ClusterInfo.EncryptionInfo)
		clientAuthenticationUpToDate := isClientAuthenticationUpToDate(wanted.ClientAuthentication, obj.ClusterInfo.ClientAuthentication)
		if !encryptionUpToDate || !clientAuthenticationUpToDate {
			obj, err = u.client.DescribeClusterWithContext(ctx, input)
			currentVersion = obj.ClusterInfo.CurrentVersion
			if err != nil {
				return managed.ExternalUpdate{}, awsclients.Wrap(err, errDescribe)
			}
			if aws.StringValue(obj.ClusterInfo.State) != stateActive {
				return managed.ExternalUpdate{}, errors.New(errStateForUpdate)
			}
			input := &svcsdk.UpdateSecurityInput{
				ClusterArn:     &currentARN,
				CurrentVersion: currentVersion,
			}

			if !clientAuthenticationUpToDate {
				input.ClientAuthentication = generateClientAuthentication(wanted.ClientAuthentication)
			}

			if !encryptionUpToDate {
				input.EncryptionInfo = generateEncryptionInfo(wanted.EncryptionInfo)

				input.EncryptionInfo.EncryptionAtRest = nil // "Updating encryption-at-rest settings on your cluster is not currently supported."
				if input.EncryptionInfo.EncryptionInTransit != nil {
					input.EncryptionInfo.EncryptionInTransit.InCluster = nil // "Updating the inter-broker encryption setting on your cluster is not currently supported."
				}
			}

			_, err := u.client.UpdateSecurityWithContext(ctx, input)
			if err != nil {
				return managed.ExternalUpdate{}, awsclients.Wrap(err, errUpdateSecurity)
			}
		}

		if !isTagsUpToDate(wanted.Tags, obj.ClusterInfo.Tags) {

			if aws.StringValue(obj.ClusterInfo.State) != stateActive {
				return managed.ExternalUpdate{}, errors.New(errStateForUpdate)
			}
			input := &svcsdk.TagResourceInput{
				ResourceArn: &currentARN,
				Tags:        generateTags(wanted.Tags),
			}

			_, err := u.client.TagResourceWithContext(ctx, input)
			if err != nil {
				return managed.ExternalUpdate{}, awsclients.Wrap(err, errUpdateTags)
			}
		}

	}

	return managed.ExternalUpdate{}, nil
}

func generateEncryptionInfo(wanted *svcapitypes.EncryptionInfo) *svcsdk.EncryptionInfo {
	if wanted != nil {
		output := &svcsdk.EncryptionInfo{}
		if wanted.EncryptionAtRest != nil {
			encryptrionAtRest := &svcsdk.EncryptionAtRest{}
			if wanted.EncryptionAtRest.DataVolumeKMSKeyID != nil {
				encryptrionAtRest.DataVolumeKMSKeyId = wanted.EncryptionAtRest.DataVolumeKMSKeyID
			}
			output.EncryptionAtRest = encryptrionAtRest
		}
		if wanted.EncryptionInTransit != nil {
			encryptionInTransit := &svcsdk.EncryptionInTransit{}
			if wanted.EncryptionInTransit.ClientBroker != nil {
				encryptionInTransit.ClientBroker = wanted.EncryptionInTransit.ClientBroker
			}
			if wanted.EncryptionInTransit.InCluster != nil {
				encryptionInTransit.InCluster = wanted.EncryptionInTransit.InCluster
			}
			output.EncryptionInTransit = encryptionInTransit
		}
		return output
	}
	return nil

}

func generateTags(wanted map[string]*string) map[string]*string {
	if wanted != nil {
		output := map[string]*string{}
		for key, value := range wanted {
			output[key] = value
		}
		return output
	}
	return nil
}

func generateClientAuthentication(wanted *svcapitypes.ClientAuthentication) *svcsdk.ClientAuthentication { // nolint:gocyclo

	if wanted != nil {
		output := &svcsdk.ClientAuthentication{}
		if wanted.SASL != nil {
			sasl := &svcsdk.Sasl{}
			if wanted.SASL.IAM != nil {
				iam := &svcsdk.Iam{}
				if wanted.SASL.IAM.Enabled != nil {
					iam.Enabled = wanted.TLS.Enabled
				}
				sasl.Iam = iam
			}
			if wanted.SASL.SCRAM != nil {
				scram := &svcsdk.Scram{}
				if wanted.SASL.SCRAM.Enabled != nil {
					scram.Enabled = wanted.SASL.SCRAM.Enabled
				}
				sasl.Scram = scram
			}

			output.Sasl = sasl
		}

		if wanted.TLS != nil {
			tls := &svcsdk.Tls{}
			if wanted.TLS.CertificateAuthorityARNList != nil {
				certArnList := []*string{}
				certArnList = append(certArnList, wanted.TLS.CertificateAuthorityARNList...)
				tls.CertificateAuthorityArnList = certArnList
			}
			if wanted.TLS.Enabled != nil {
				tls.Enabled = wanted.TLS.Enabled
			}
			output.Tls = tls
		}

		if wanted.Unauthenticated != nil {
			unauthenticated := &svcsdk.Unauthenticated{}
			if wanted.Unauthenticated.Enabled != nil {
				unauthenticated.Enabled = wanted.Unauthenticated.Enabled
			}
			output.Unauthenticated = unauthenticated
		}

		return output
	}
	return nil

}

func generateCustomConfigurationInfo(wanted *svcapitypes.CustomConfigurationInfo) *svcsdk.ConfigurationInfo {
	if wanted != nil && wanted.ARN != nil && wanted.Revision != nil {
		return &svcsdk.ConfigurationInfo{
			Arn:      wanted.ARN,
			Revision: wanted.Revision,
		}
	}
	return nil

}

func generateLoggingInfoInput(wanted *svcapitypes.LoggingInfo) *svcsdk.LoggingInfo { // nolint:gocyclo

	output := &svcsdk.LoggingInfo{}
	if wanted.BrokerLogs != nil {
		brokerLogs := &svcsdk.BrokerLogs{}
		if wanted.BrokerLogs.CloudWatchLogs != nil {
			cloudWatchLogs := &svcsdk.CloudWatchLogs{}
			if wanted.BrokerLogs.CloudWatchLogs.Enabled != nil {
				cloudWatchLogs.SetEnabled(*wanted.BrokerLogs.CloudWatchLogs.Enabled)
			}
			if wanted.BrokerLogs.CloudWatchLogs.LogGroup != nil {
				cloudWatchLogs.SetLogGroup(*wanted.BrokerLogs.CloudWatchLogs.LogGroup)
			}
			brokerLogs.SetCloudWatchLogs(cloudWatchLogs)
		}
		if wanted.BrokerLogs.Firehose != nil {
			firehose := &svcsdk.Firehose{}
			if wanted.BrokerLogs.Firehose.DeliveryStream != nil {
				firehose.SetDeliveryStream(*wanted.BrokerLogs.Firehose.DeliveryStream)
			}
			if wanted.BrokerLogs.Firehose.Enabled != nil {
				firehose.SetEnabled(*wanted.BrokerLogs.Firehose.Enabled)
			}
			brokerLogs.SetFirehose(firehose)
		}
		if wanted.BrokerLogs.S3 != nil {
			s3 := &svcsdk.S3{}
			if wanted.BrokerLogs.S3.Bucket != nil {
				s3.SetBucket(*wanted.BrokerLogs.S3.Bucket)
			}
			if wanted.BrokerLogs.S3.Enabled != nil {
				s3.SetEnabled(*wanted.BrokerLogs.S3.Enabled)
			}
			if wanted.BrokerLogs.S3.Prefix != nil {
				s3.SetPrefix(*wanted.BrokerLogs.S3.Prefix)
			}
			brokerLogs.SetS3(s3)
		}
		output.SetBrokerLogs(brokerLogs)
	}
	return output
}

func generateOpenMonitorinInput(wanted *svcapitypes.OpenMonitoringInfo) *svcsdk.OpenMonitoringInfo {
	output := &svcsdk.OpenMonitoringInfo{}
	if wanted.Prometheus != nil {
		prometheus := &svcsdk.PrometheusInfo{}
		if wanted.Prometheus.JmxExporter != nil {
			jmxExporter := &svcsdk.JmxExporterInfo{}
			if wanted.Prometheus.JmxExporter.EnabledInBroker != nil {
				jmxExporter.SetEnabledInBroker(*wanted.Prometheus.JmxExporter.EnabledInBroker)
			}
			prometheus.SetJmxExporter(jmxExporter)
		}
		if wanted.Prometheus.NodeExporter != nil {
			NodeExporter := &svcsdk.NodeExporterInfo{}
			if wanted.Prometheus.NodeExporter.EnabledInBroker != nil {
				NodeExporter.SetEnabledInBroker(*wanted.Prometheus.NodeExporter.EnabledInBroker)
			}
			prometheus.SetNodeExporter(NodeExporter)
		}
		output.SetPrometheus(prometheus)
	}
	return output

}
