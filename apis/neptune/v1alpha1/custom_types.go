/*
Copyright 2022 The Crossplane Authors.

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

// CustomDBClusterParameters contains the additional fields for DB Cluster
type CustomDBClusterParameters struct {
	// The ApplyImmediately parameter only affects the NewDBClusterIdentifier and
	// MasterUserPassword values. If you set the ApplyImmediately parameter value
	// to false, then changes to the NewDBClusterIdentifier and MasterUserPassword
	// values are applied during the next maintenance window. All other changes
	// are applied immediately, regardless of the value of the ApplyImmediately
	// parameter.
	//
	// Default: false
	ApplyImmediately *bool `json:"applyImmediately,omitempty"`

	// Determines whether a final DB cluster snapshot is created before the DB cluster
	// is deleted. If true is specified, no DB cluster snapshot is created. If false
	// is specified, a DB cluster snapshot is created before the DB cluster is deleted.
	//
	// You must specify a FinalDBSnapshotIdentifier parameter if SkipFinalSnapshot
	// is false.
	//
	// Default: false
	SkipFinalSnapshot *bool `json:"skipFinalSnapshot,omitempty"`
}

// CustomDBClusterObservation includes the custom status fields of DB Cluster.
type CustomDBClusterObservation struct{}
