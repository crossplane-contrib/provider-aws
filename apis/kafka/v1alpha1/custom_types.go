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

package v1alpha1

// CustomClusterParameters contains the additional fields for ClusterParameters.
type CustomClusterParameters struct {

	// The connection string to use to connect to the Apache ZooKeeper cluster.
	// +optional
	ZookeeperConnectString *string `json:"zookeeperConnectString,omitempty"`

	// The connection string to use to connect to zookeeper cluster on Tls port.
	// +optional
	ZookeeperConnectStringTLS *string `json:"zookeeperConnectStringTLS,omitempty"`
}
