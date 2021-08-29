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

// Code generated by ack-generate. DO NOT EDIT.

package cluster

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	svcsdk "github.com/aws/aws-sdk-go/service/kafka"

	svcapitypes "github.com/crossplane/provider-aws/apis/kafka/v1alpha1"
)

// NOTE(muvaf): We return pointers in case the function needs to start with an
// empty object, hence need to return a new pointer.

// GenerateDescribeClusterInput returns input for read
// operation.
func GenerateDescribeClusterInput(cr *svcapitypes.Cluster) *svcsdk.DescribeClusterInput {
	res := &svcsdk.DescribeClusterInput{}

	if cr.Status.AtProvider.ClusterARN != nil {
		res.SetClusterArn(*cr.Status.AtProvider.ClusterARN)
	}

	return res
}

// GenerateCluster returns the current state in the form of *svcapitypes.Cluster.
func GenerateCluster(resp *svcsdk.DescribeClusterOutput) *svcapitypes.Cluster {
	cr := &svcapitypes.Cluster{}

	if resp.ClusterInfo.BrokerNodeGroupInfo != nil {
		f1 := &svcapitypes.BrokerNodeGroupInfo{}
		if resp.ClusterInfo.BrokerNodeGroupInfo.BrokerAZDistribution != nil {
			f1.BrokerAZDistribution = resp.ClusterInfo.BrokerNodeGroupInfo.BrokerAZDistribution
		}
		if resp.ClusterInfo.BrokerNodeGroupInfo.ClientSubnets != nil {
			f1f1 := []*string{}
			for _, f1f1iter := range resp.ClusterInfo.BrokerNodeGroupInfo.ClientSubnets {
				var f1f1elem string
				f1f1elem = *f1f1iter
				f1f1 = append(f1f1, &f1f1elem)
			}
			f1.ClientSubnets = f1f1
		}
		if resp.ClusterInfo.BrokerNodeGroupInfo.InstanceType != nil {
			f1.InstanceType = resp.ClusterInfo.BrokerNodeGroupInfo.InstanceType
		}
		if resp.ClusterInfo.BrokerNodeGroupInfo.SecurityGroups != nil {
			f1f3 := []*string{}
			for _, f1f3iter := range resp.ClusterInfo.BrokerNodeGroupInfo.SecurityGroups {
				var f1f3elem string
				f1f3elem = *f1f3iter
				f1f3 = append(f1f3, &f1f3elem)
			}
			f1.SecurityGroups = f1f3
		}
		if resp.ClusterInfo.BrokerNodeGroupInfo.StorageInfo != nil {
			f1f4 := &svcapitypes.StorageInfo{}
			if resp.ClusterInfo.BrokerNodeGroupInfo.StorageInfo.EbsStorageInfo != nil {
				f1f4f0 := &svcapitypes.EBSStorageInfo{}
				if resp.ClusterInfo.BrokerNodeGroupInfo.StorageInfo.EbsStorageInfo.VolumeSize != nil {
					f1f4f0.VolumeSize = resp.ClusterInfo.BrokerNodeGroupInfo.StorageInfo.EbsStorageInfo.VolumeSize
				}
				f1f4.EBSStorageInfo = f1f4f0
			}
			f1.StorageInfo = f1f4
		}
		cr.Spec.ForProvider.BrokerNodeGroupInfo = f1
	} else {
		cr.Spec.ForProvider.BrokerNodeGroupInfo = nil
	}
	if resp.ClusterInfo.ClientAuthentication != nil {
		f2 := &svcapitypes.ClientAuthentication{}
		if resp.ClusterInfo.ClientAuthentication.Sasl != nil {
			f2f0 := &svcapitypes.Sasl{}
			if resp.ClusterInfo.ClientAuthentication.Sasl.Scram != nil {
				f2f0f0 := &svcapitypes.Scram{}
				if resp.ClusterInfo.ClientAuthentication.Sasl.Scram.Enabled != nil {
					f2f0f0.Enabled = resp.ClusterInfo.ClientAuthentication.Sasl.Scram.Enabled
				}
				f2f0.Scram = f2f0f0
			}
			f2.Sasl = f2f0
		}
		if resp.ClusterInfo.ClientAuthentication.Tls != nil {
			f2f1 := &svcapitypes.TLS{}
			if resp.ClusterInfo.ClientAuthentication.Tls.CertificateAuthorityArnList != nil {
				f2f1f0 := []*string{}
				for _, f2f1f0iter := range resp.ClusterInfo.ClientAuthentication.Tls.CertificateAuthorityArnList {
					var f2f1f0elem string
					f2f1f0elem = *f2f1f0iter
					f2f1f0 = append(f2f1f0, &f2f1f0elem)
				}
				f2f1.CertificateAuthorityARNList = f2f1f0
			}
			f2.TLS = f2f1
		}
		cr.Spec.ForProvider.ClientAuthentication = f2
	} else {
		cr.Spec.ForProvider.ClientAuthentication = nil
	}
	if resp.ClusterInfo.ClusterArn != nil {
		cr.Status.AtProvider.ClusterARN = resp.ClusterInfo.ClusterArn
	} else {
		cr.Status.AtProvider.ClusterARN = nil
	}
	if resp.ClusterInfo.ClusterName != nil {
		cr.Spec.ForProvider.ClusterName = resp.ClusterInfo.ClusterName
	} else {
		cr.Spec.ForProvider.ClusterName = nil
	}
	if resp.ClusterInfo.EncryptionInfo != nil {
		f8 := &svcapitypes.EncryptionInfo{}
		if resp.ClusterInfo.EncryptionInfo.EncryptionAtRest != nil {
			f8f0 := &svcapitypes.EncryptionAtRest{}
			if resp.ClusterInfo.EncryptionInfo.EncryptionAtRest.DataVolumeKMSKeyId != nil {
				f8f0.DataVolumeKMSKeyID = resp.ClusterInfo.EncryptionInfo.EncryptionAtRest.DataVolumeKMSKeyId
			}
			f8.EncryptionAtRest = f8f0
		}
		if resp.ClusterInfo.EncryptionInfo.EncryptionInTransit != nil {
			f8f1 := &svcapitypes.EncryptionInTransit{}
			if resp.ClusterInfo.EncryptionInfo.EncryptionInTransit.ClientBroker != nil {
				f8f1.ClientBroker = resp.ClusterInfo.EncryptionInfo.EncryptionInTransit.ClientBroker
			}
			if resp.ClusterInfo.EncryptionInfo.EncryptionInTransit.InCluster != nil {
				f8f1.InCluster = resp.ClusterInfo.EncryptionInfo.EncryptionInTransit.InCluster
			}
			f8.EncryptionInTransit = f8f1
		}
		cr.Spec.ForProvider.EncryptionInfo = f8
	} else {
		cr.Spec.ForProvider.EncryptionInfo = nil
	}
	if resp.ClusterInfo.EnhancedMonitoring != nil {
		cr.Spec.ForProvider.EnhancedMonitoring = resp.ClusterInfo.EnhancedMonitoring
	} else {
		cr.Spec.ForProvider.EnhancedMonitoring = nil
	}
	if resp.ClusterInfo.LoggingInfo != nil {
		f10 := &svcapitypes.LoggingInfo{}
		if resp.ClusterInfo.LoggingInfo.BrokerLogs != nil {
			f10f0 := &svcapitypes.BrokerLogs{}
			if resp.ClusterInfo.LoggingInfo.BrokerLogs.CloudWatchLogs != nil {
				f10f0f0 := &svcapitypes.CloudWatchLogs{}
				if resp.ClusterInfo.LoggingInfo.BrokerLogs.CloudWatchLogs.Enabled != nil {
					f10f0f0.Enabled = resp.ClusterInfo.LoggingInfo.BrokerLogs.CloudWatchLogs.Enabled
				}
				if resp.ClusterInfo.LoggingInfo.BrokerLogs.CloudWatchLogs.LogGroup != nil {
					f10f0f0.LogGroup = resp.ClusterInfo.LoggingInfo.BrokerLogs.CloudWatchLogs.LogGroup
				}
				f10f0.CloudWatchLogs = f10f0f0
			}
			if resp.ClusterInfo.LoggingInfo.BrokerLogs.Firehose != nil {
				f10f0f1 := &svcapitypes.Firehose{}
				if resp.ClusterInfo.LoggingInfo.BrokerLogs.Firehose.DeliveryStream != nil {
					f10f0f1.DeliveryStream = resp.ClusterInfo.LoggingInfo.BrokerLogs.Firehose.DeliveryStream
				}
				if resp.ClusterInfo.LoggingInfo.BrokerLogs.Firehose.Enabled != nil {
					f10f0f1.Enabled = resp.ClusterInfo.LoggingInfo.BrokerLogs.Firehose.Enabled
				}
				f10f0.Firehose = f10f0f1
			}
			if resp.ClusterInfo.LoggingInfo.BrokerLogs.S3 != nil {
				f10f0f2 := &svcapitypes.S3{}
				if resp.ClusterInfo.LoggingInfo.BrokerLogs.S3.Bucket != nil {
					f10f0f2.Bucket = resp.ClusterInfo.LoggingInfo.BrokerLogs.S3.Bucket
				}
				if resp.ClusterInfo.LoggingInfo.BrokerLogs.S3.Enabled != nil {
					f10f0f2.Enabled = resp.ClusterInfo.LoggingInfo.BrokerLogs.S3.Enabled
				}
				if resp.ClusterInfo.LoggingInfo.BrokerLogs.S3.Prefix != nil {
					f10f0f2.Prefix = resp.ClusterInfo.LoggingInfo.BrokerLogs.S3.Prefix
				}
				f10f0.S3 = f10f0f2
			}
			f10.BrokerLogs = f10f0
		}
		cr.Spec.ForProvider.LoggingInfo = f10
	} else {
		cr.Spec.ForProvider.LoggingInfo = nil
	}
	if resp.ClusterInfo.NumberOfBrokerNodes != nil {
		cr.Spec.ForProvider.NumberOfBrokerNodes = resp.ClusterInfo.NumberOfBrokerNodes
	} else {
		cr.Spec.ForProvider.NumberOfBrokerNodes = nil
	}
	if resp.ClusterInfo.OpenMonitoring != nil {
		f12 := &svcapitypes.OpenMonitoringInfo{}
		if resp.ClusterInfo.OpenMonitoring.Prometheus != nil {
			f12f0 := &svcapitypes.PrometheusInfo{}
			if resp.ClusterInfo.OpenMonitoring.Prometheus.JmxExporter != nil {
				f12f0f0 := &svcapitypes.JmxExporterInfo{}
				if resp.ClusterInfo.OpenMonitoring.Prometheus.JmxExporter.EnabledInBroker != nil {
					f12f0f0.EnabledInBroker = resp.ClusterInfo.OpenMonitoring.Prometheus.JmxExporter.EnabledInBroker
				}
				f12f0.JmxExporter = f12f0f0
			}
			if resp.ClusterInfo.OpenMonitoring.Prometheus.NodeExporter != nil {
				f12f0f1 := &svcapitypes.NodeExporterInfo{}
				if resp.ClusterInfo.OpenMonitoring.Prometheus.NodeExporter.EnabledInBroker != nil {
					f12f0f1.EnabledInBroker = resp.ClusterInfo.OpenMonitoring.Prometheus.NodeExporter.EnabledInBroker
				}
				f12f0.NodeExporter = f12f0f1
			}
			f12.Prometheus = f12f0
		}
		cr.Spec.ForProvider.OpenMonitoring = f12
	} else {
		cr.Spec.ForProvider.OpenMonitoring = nil
	}
	if resp.ClusterInfo.State != nil {
		cr.Status.AtProvider.State = resp.ClusterInfo.State
	} else {
		cr.Status.AtProvider.State = nil
	}
	if resp.ClusterInfo.Tags != nil {
		f14 := map[string]*string{}
		for f14key, f14valiter := range resp.ClusterInfo.Tags {
			var f14val string
			f14val = *f14valiter
			f14[f14key] = &f14val
		}
		cr.Spec.ForProvider.Tags = f14
	} else {
		cr.Spec.ForProvider.Tags = nil
	}

	return cr
}

// GenerateCreateClusterInput returns a create input.
func GenerateCreateClusterInput(cr *svcapitypes.Cluster) *svcsdk.CreateClusterInput {
	res := &svcsdk.CreateClusterInput{}

	if cr.Spec.ForProvider.BrokerNodeGroupInfo != nil {
		f0 := &svcsdk.BrokerNodeGroupInfo{}
		if cr.Spec.ForProvider.BrokerNodeGroupInfo.BrokerAZDistribution != nil {
			f0.SetBrokerAZDistribution(*cr.Spec.ForProvider.BrokerNodeGroupInfo.BrokerAZDistribution)
		}
		if cr.Spec.ForProvider.BrokerNodeGroupInfo.ClientSubnets != nil {
			f0f1 := []*string{}
			for _, f0f1iter := range cr.Spec.ForProvider.BrokerNodeGroupInfo.ClientSubnets {
				var f0f1elem string
				f0f1elem = *f0f1iter
				f0f1 = append(f0f1, &f0f1elem)
			}
			f0.SetClientSubnets(f0f1)
		}
		if cr.Spec.ForProvider.BrokerNodeGroupInfo.InstanceType != nil {
			f0.SetInstanceType(*cr.Spec.ForProvider.BrokerNodeGroupInfo.InstanceType)
		}
		if cr.Spec.ForProvider.BrokerNodeGroupInfo.SecurityGroups != nil {
			f0f3 := []*string{}
			for _, f0f3iter := range cr.Spec.ForProvider.BrokerNodeGroupInfo.SecurityGroups {
				var f0f3elem string
				f0f3elem = *f0f3iter
				f0f3 = append(f0f3, &f0f3elem)
			}
			f0.SetSecurityGroups(f0f3)
		}
		if cr.Spec.ForProvider.BrokerNodeGroupInfo.StorageInfo != nil {
			f0f4 := &svcsdk.StorageInfo{}
			if cr.Spec.ForProvider.BrokerNodeGroupInfo.StorageInfo.EBSStorageInfo != nil {
				f0f4f0 := &svcsdk.EBSStorageInfo{}
				if cr.Spec.ForProvider.BrokerNodeGroupInfo.StorageInfo.EBSStorageInfo.VolumeSize != nil {
					f0f4f0.SetVolumeSize(*cr.Spec.ForProvider.BrokerNodeGroupInfo.StorageInfo.EBSStorageInfo.VolumeSize)
				}
				f0f4.SetEbsStorageInfo(f0f4f0)
			}
			f0.SetStorageInfo(f0f4)
		}
		res.SetBrokerNodeGroupInfo(f0)
	}
	if cr.Spec.ForProvider.ClientAuthentication != nil {
		f1 := &svcsdk.ClientAuthentication{}
		if cr.Spec.ForProvider.ClientAuthentication.Sasl != nil {
			f1f0 := &svcsdk.Sasl{}
			if cr.Spec.ForProvider.ClientAuthentication.Sasl.Scram != nil {
				f1f0f0 := &svcsdk.Scram{}
				if cr.Spec.ForProvider.ClientAuthentication.Sasl.Scram.Enabled != nil {
					f1f0f0.SetEnabled(*cr.Spec.ForProvider.ClientAuthentication.Sasl.Scram.Enabled)
				}
				f1f0.SetScram(f1f0f0)
			}
			f1.SetSasl(f1f0)
		}
		if cr.Spec.ForProvider.ClientAuthentication.TLS != nil {
			f1f1 := &svcsdk.Tls{}
			if cr.Spec.ForProvider.ClientAuthentication.TLS.CertificateAuthorityARNList != nil {
				f1f1f0 := []*string{}
				for _, f1f1f0iter := range cr.Spec.ForProvider.ClientAuthentication.TLS.CertificateAuthorityARNList {
					var f1f1f0elem string
					f1f1f0elem = *f1f1f0iter
					f1f1f0 = append(f1f1f0, &f1f1f0elem)
				}
				f1f1.SetCertificateAuthorityArnList(f1f1f0)
			}
			f1.SetTls(f1f1)
		}
		res.SetClientAuthentication(f1)
	}
	if cr.Spec.ForProvider.ClusterName != nil {
		res.SetClusterName(*cr.Spec.ForProvider.ClusterName)
	}
	if cr.Spec.ForProvider.ConfigurationInfo != nil {
		f3 := &svcsdk.ConfigurationInfo{}
		if cr.Spec.ForProvider.ConfigurationInfo.ARN != nil {
			f3.SetArn(*cr.Spec.ForProvider.ConfigurationInfo.ARN)
		}
		if cr.Spec.ForProvider.ConfigurationInfo.Revision != nil {
			f3.SetRevision(*cr.Spec.ForProvider.ConfigurationInfo.Revision)
		}
		res.SetConfigurationInfo(f3)
	}
	if cr.Spec.ForProvider.EncryptionInfo != nil {
		f4 := &svcsdk.EncryptionInfo{}
		if cr.Spec.ForProvider.EncryptionInfo.EncryptionAtRest != nil {
			f4f0 := &svcsdk.EncryptionAtRest{}
			if cr.Spec.ForProvider.EncryptionInfo.EncryptionAtRest.DataVolumeKMSKeyID != nil {
				f4f0.SetDataVolumeKMSKeyId(*cr.Spec.ForProvider.EncryptionInfo.EncryptionAtRest.DataVolumeKMSKeyID)
			}
			f4.SetEncryptionAtRest(f4f0)
		}
		if cr.Spec.ForProvider.EncryptionInfo.EncryptionInTransit != nil {
			f4f1 := &svcsdk.EncryptionInTransit{}
			if cr.Spec.ForProvider.EncryptionInfo.EncryptionInTransit.ClientBroker != nil {
				f4f1.SetClientBroker(*cr.Spec.ForProvider.EncryptionInfo.EncryptionInTransit.ClientBroker)
			}
			if cr.Spec.ForProvider.EncryptionInfo.EncryptionInTransit.InCluster != nil {
				f4f1.SetInCluster(*cr.Spec.ForProvider.EncryptionInfo.EncryptionInTransit.InCluster)
			}
			f4.SetEncryptionInTransit(f4f1)
		}
		res.SetEncryptionInfo(f4)
	}
	if cr.Spec.ForProvider.EnhancedMonitoring != nil {
		res.SetEnhancedMonitoring(*cr.Spec.ForProvider.EnhancedMonitoring)
	}
	if cr.Spec.ForProvider.KafkaVersion != nil {
		res.SetKafkaVersion(*cr.Spec.ForProvider.KafkaVersion)
	}
	if cr.Spec.ForProvider.LoggingInfo != nil {
		f7 := &svcsdk.LoggingInfo{}
		if cr.Spec.ForProvider.LoggingInfo.BrokerLogs != nil {
			f7f0 := &svcsdk.BrokerLogs{}
			if cr.Spec.ForProvider.LoggingInfo.BrokerLogs.CloudWatchLogs != nil {
				f7f0f0 := &svcsdk.CloudWatchLogs{}
				if cr.Spec.ForProvider.LoggingInfo.BrokerLogs.CloudWatchLogs.Enabled != nil {
					f7f0f0.SetEnabled(*cr.Spec.ForProvider.LoggingInfo.BrokerLogs.CloudWatchLogs.Enabled)
				}
				if cr.Spec.ForProvider.LoggingInfo.BrokerLogs.CloudWatchLogs.LogGroup != nil {
					f7f0f0.SetLogGroup(*cr.Spec.ForProvider.LoggingInfo.BrokerLogs.CloudWatchLogs.LogGroup)
				}
				f7f0.SetCloudWatchLogs(f7f0f0)
			}
			if cr.Spec.ForProvider.LoggingInfo.BrokerLogs.Firehose != nil {
				f7f0f1 := &svcsdk.Firehose{}
				if cr.Spec.ForProvider.LoggingInfo.BrokerLogs.Firehose.DeliveryStream != nil {
					f7f0f1.SetDeliveryStream(*cr.Spec.ForProvider.LoggingInfo.BrokerLogs.Firehose.DeliveryStream)
				}
				if cr.Spec.ForProvider.LoggingInfo.BrokerLogs.Firehose.Enabled != nil {
					f7f0f1.SetEnabled(*cr.Spec.ForProvider.LoggingInfo.BrokerLogs.Firehose.Enabled)
				}
				f7f0.SetFirehose(f7f0f1)
			}
			if cr.Spec.ForProvider.LoggingInfo.BrokerLogs.S3 != nil {
				f7f0f2 := &svcsdk.S3{}
				if cr.Spec.ForProvider.LoggingInfo.BrokerLogs.S3.Bucket != nil {
					f7f0f2.SetBucket(*cr.Spec.ForProvider.LoggingInfo.BrokerLogs.S3.Bucket)
				}
				if cr.Spec.ForProvider.LoggingInfo.BrokerLogs.S3.Enabled != nil {
					f7f0f2.SetEnabled(*cr.Spec.ForProvider.LoggingInfo.BrokerLogs.S3.Enabled)
				}
				if cr.Spec.ForProvider.LoggingInfo.BrokerLogs.S3.Prefix != nil {
					f7f0f2.SetPrefix(*cr.Spec.ForProvider.LoggingInfo.BrokerLogs.S3.Prefix)
				}
				f7f0.SetS3(f7f0f2)
			}
			f7.SetBrokerLogs(f7f0)
		}
		res.SetLoggingInfo(f7)
	}
	if cr.Spec.ForProvider.NumberOfBrokerNodes != nil {
		res.SetNumberOfBrokerNodes(*cr.Spec.ForProvider.NumberOfBrokerNodes)
	}
	if cr.Spec.ForProvider.OpenMonitoring != nil {
		f9 := &svcsdk.OpenMonitoringInfo{}
		if cr.Spec.ForProvider.OpenMonitoring.Prometheus != nil {
			f9f0 := &svcsdk.PrometheusInfo{}
			if cr.Spec.ForProvider.OpenMonitoring.Prometheus.JmxExporter != nil {
				f9f0f0 := &svcsdk.JmxExporterInfo{}
				if cr.Spec.ForProvider.OpenMonitoring.Prometheus.JmxExporter.EnabledInBroker != nil {
					f9f0f0.SetEnabledInBroker(*cr.Spec.ForProvider.OpenMonitoring.Prometheus.JmxExporter.EnabledInBroker)
				}
				f9f0.SetJmxExporter(f9f0f0)
			}
			if cr.Spec.ForProvider.OpenMonitoring.Prometheus.NodeExporter != nil {
				f9f0f1 := &svcsdk.NodeExporterInfo{}
				if cr.Spec.ForProvider.OpenMonitoring.Prometheus.NodeExporter.EnabledInBroker != nil {
					f9f0f1.SetEnabledInBroker(*cr.Spec.ForProvider.OpenMonitoring.Prometheus.NodeExporter.EnabledInBroker)
				}
				f9f0.SetNodeExporter(f9f0f1)
			}
			f9.SetPrometheus(f9f0)
		}
		res.SetOpenMonitoring(f9)
	}
	if cr.Spec.ForProvider.Tags != nil {
		f10 := map[string]*string{}
		for f10key, f10valiter := range cr.Spec.ForProvider.Tags {
			var f10val string
			f10val = *f10valiter
			f10[f10key] = &f10val
		}
		res.SetTags(f10)
	}

	return res
}

// GenerateDeleteClusterInput returns a deletion input.
func GenerateDeleteClusterInput(cr *svcapitypes.Cluster) *svcsdk.DeleteClusterInput {
	res := &svcsdk.DeleteClusterInput{}

	if cr.Status.AtProvider.ClusterARN != nil {
		res.SetClusterArn(*cr.Status.AtProvider.ClusterARN)
	}

	return res
}

// IsNotFound returns whether the given error is of type NotFound or not.
func IsNotFound(err error) bool {
	awsErr, ok := err.(awserr.Error)
	return ok && awsErr.Code() == "NotFoundException"
}
