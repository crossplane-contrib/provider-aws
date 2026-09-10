/*
Copyright 2023 The Crossplane Authors.

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

package utils

import (
	svcsdk "github.com/aws/aws-sdk-go/service/rds"

	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

// GenerateCloudWatchExportConfiguration computes the CloudwatchLogsExportConfiguration
// (log types to enable/disable) needed to reconcile the currently enabled CloudWatch
// log exports on AWS towards the desired spec. It returns nil when no change is
// required, so callers can leave the field unset on the modify request in that case.
func GenerateCloudWatchExportConfiguration(spec, current []*string) *svcsdk.CloudwatchLogsExportConfiguration {
	toEnable := []*string{}
	toDisable := []*string{}

	currentMap := make(map[string]struct{}, len(current))
	for _, currentID := range current {
		currentMap[pointer.StringValue(currentID)] = struct{}{}
	}

	specMap := make(map[string]struct{}, len(spec))
	for _, specID := range spec {
		key := pointer.StringValue(specID)
		specMap[key] = struct{}{}

		if _, exists := currentMap[key]; !exists {
			toEnable = append(toEnable, specID)
		}
	}

	for _, currentID := range current {
		if _, exists := specMap[pointer.StringValue(currentID)]; !exists {
			toDisable = append(toDisable, currentID)
		}
	}

	if len(toEnable) == 0 && len(toDisable) == 0 {
		return nil
	}

	return &svcsdk.CloudwatchLogsExportConfiguration{
		EnableLogTypes:  toEnable,
		DisableLogTypes: toDisable,
	}
}

// AreSameElements reports whether the two slices contain the same set of elements,
// ignoring order and duplicates.
func AreSameElements(a1, a2 []*string) bool {
	if len(a1) != len(a2) {
		return false
	}

	m2 := make(map[string]struct{}, len(a2))
	for _, s2 := range a2 {
		m2[pointer.StringValue(s2)] = struct{}{}
	}

	for _, s1 := range a1 {
		v1 := pointer.StringValue(s1)
		if _, exists := m2[v1]; !exists {
			return false
		}
	}

	return true
}
