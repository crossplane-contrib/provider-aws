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

package v1alpha1

import xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

// CustomBackupParameters are custom parameters for Backup.
type CustomBackupParameters struct {
	// TableName is the name of the Table whose backup will be taken.
	TableName string `json:"tableName,omitempty"`

	// TableNameRef points to the Table resource whose Name will be used to fill
	// TableName field.
	TableNameRef *xpv1.Reference `json:"tableNameRef,omitempty"`

	// TableNameSelector selects a Table resource.
	TableNameSelector *xpv1.Selector `json:"tableNameSelector,omitempty"`
}

// CustomBackupObservation are custom obervations for Backup.
type CustomBackupObservation struct{}

// CustomTableParameters are custom parameters for Table.
type CustomTableParameters struct{}

// CustomBackupObservation are custom obervations for Table.
type CustomTableObservation struct{}

// CustomGlobalTableParameters are custom parameters for GlobalTable.
type CustomGlobalTableParameters struct{}

// CustomGlobalTableObservation are custom obervations for GlobalTable.
type CustomGlobalTableObservation struct{}
