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

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	svcsdk "github.com/aws/aws-sdk-go/service/kafka"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/kafka/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// SetupCluster adds a controller that reconciles Cluster.
func SetupCluster(mgr ctrl.Manager, l logging.Logger, limiter workqueue.RateLimiter) error {
	name := managed.ControllerName(svcapitypes.ClusterGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preDelete = preDelete
			e.postCreate = postCreate
			e.lateInitialize = LateInitialize
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(limiter),
		}).
		For(&svcapitypes.Cluster{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.ClusterGroupVersionKind),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func preDelete(_ context.Context, cr *svcapitypes.Cluster, obj *svcsdk.DeleteClusterInput) (bool, error) {
	obj.ClusterArn = awsclients.String(meta.GetExternalName(cr))
	return false, nil
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
	return obs, nil
}

func postCreate(_ context.Context, cr *svcapitypes.Cluster, obj *svcsdk.CreateClusterOutput, _ managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, awsclients.StringValue(obj.ClusterArn))
	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

// LateInitialize fills the empty fields in *svcapitypes.ClusterParameters with
// the values seen in svcsdk.DescribeClusterOutput.
// nolint:gocyclo
func LateInitialize(cr *svcapitypes.ClusterParameters, obj *svcsdk.DescribeClusterOutput) error {

	if cr.EnhancedMonitoring == nil && obj.ClusterInfo.EnhancedMonitoring != nil {
		cr.EnhancedMonitoring = awsclients.LateInitializeStringPtr(cr.EnhancedMonitoring, obj.ClusterInfo.EnhancedMonitoring)
	}

	if cr.BrokerNodeGroupInfo.BrokerAZDistribution == nil && obj.ClusterInfo.BrokerNodeGroupInfo.BrokerAZDistribution != nil {
		cr.BrokerNodeGroupInfo.BrokerAZDistribution = awsclients.LateInitializeStringPtr(cr.BrokerNodeGroupInfo.BrokerAZDistribution, obj.ClusterInfo.BrokerNodeGroupInfo.BrokerAZDistribution)
	}

	if cr.CustomClusterParameters.ZookeeperConnectString == nil && obj.ClusterInfo.ZookeeperConnectString != nil {
		cr.CustomClusterParameters.ZookeeperConnectString = awsclients.LateInitializeStringPtr(cr.CustomClusterParameters.ZookeeperConnectString, obj.ClusterInfo.ZookeeperConnectString)
	}

	if cr.CustomClusterParameters.ZookeeperConnectStringTLS == nil && obj.ClusterInfo.ZookeeperConnectStringTls != nil {
		cr.CustomClusterParameters.ZookeeperConnectStringTLS = awsclients.LateInitializeStringPtr(cr.CustomClusterParameters.ZookeeperConnectStringTLS, obj.ClusterInfo.ZookeeperConnectStringTls)
	}

	if cr.BrokerNodeGroupInfo.SecurityGroups == nil && obj.ClusterInfo.BrokerNodeGroupInfo.SecurityGroups != nil {
		cr.BrokerNodeGroupInfo.SecurityGroups = obj.ClusterInfo.BrokerNodeGroupInfo.SecurityGroups
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
