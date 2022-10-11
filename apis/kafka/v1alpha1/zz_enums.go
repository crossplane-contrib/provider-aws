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

package v1alpha1

type BrokerAZDistribution string

const (
	BrokerAZDistribution_DEFAULT BrokerAZDistribution = "DEFAULT"
)

type ClientBroker string

const (
	ClientBroker_TLS           ClientBroker = "TLS"
	ClientBroker_TLS_PLAINTEXT ClientBroker = "TLS_PLAINTEXT"
	ClientBroker_PLAINTEXT     ClientBroker = "PLAINTEXT"
)

type ClusterState string

const (
	ClusterState_ACTIVE           ClusterState = "ACTIVE"
	ClusterState_CREATING         ClusterState = "CREATING"
	ClusterState_DELETING         ClusterState = "DELETING"
	ClusterState_FAILED           ClusterState = "FAILED"
	ClusterState_HEALING          ClusterState = "HEALING"
	ClusterState_MAINTENANCE      ClusterState = "MAINTENANCE"
	ClusterState_REBOOTING_BROKER ClusterState = "REBOOTING_BROKER"
	ClusterState_UPDATING         ClusterState = "UPDATING"
)

type ClusterType string

const (
	ClusterType_PROVISIONED ClusterType = "PROVISIONED"
	ClusterType_SERVERLESS  ClusterType = "SERVERLESS"
)

type ConfigurationState string

const (
	ConfigurationState_ACTIVE        ConfigurationState = "ACTIVE"
	ConfigurationState_DELETING      ConfigurationState = "DELETING"
	ConfigurationState_DELETE_FAILED ConfigurationState = "DELETE_FAILED"
)

type EnhancedMonitoring string

const (
	EnhancedMonitoring_DEFAULT                 EnhancedMonitoring = "DEFAULT"
	EnhancedMonitoring_PER_BROKER              EnhancedMonitoring = "PER_BROKER"
	EnhancedMonitoring_PER_TOPIC_PER_BROKER    EnhancedMonitoring = "PER_TOPIC_PER_BROKER"
	EnhancedMonitoring_PER_TOPIC_PER_PARTITION EnhancedMonitoring = "PER_TOPIC_PER_PARTITION"
)

type KafkaVersionStatus string

const (
	KafkaVersionStatus_ACTIVE     KafkaVersionStatus = "ACTIVE"
	KafkaVersionStatus_DEPRECATED KafkaVersionStatus = "DEPRECATED"
)

type NodeType string

const (
	NodeType_BROKER NodeType = "BROKER"
)
