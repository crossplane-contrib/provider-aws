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

package kube

import (
	"os"
)

const (
	fileServiceAccountNamespace = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	environPodNamespace         = "POD_NAMESPACE"
	defaultNamespace            = "crossplane-system"
)

// GetProviderNamespace that is considered the default to store and access
// namespaced resources like cache secrets.
// This function is necessary since client-go does not set a namespace by
// default and the provider is not guaranteed to always run in the
// crossplane-system namespace.
func GetProviderNamespace() string {
	// Try inferring the namespace from the environment
	if podNs := os.Getenv(environPodNamespace); podNs != "" {
		return podNs
	}

	// Try to get the service account namespace from file
	saNsRaw, err := os.ReadFile(fileServiceAccountNamespace)
	saNs := string(saNsRaw)
	if err == nil && saNs != "" {
		return saNs
	}

	// TODO: Do we want to read from the local kubeconfig if the provider is
	//       running on a local dev machine?
	//       While this might certainly work, it could produce trigger
	//       unintended behaviour if the user context points to something else
	//       then crossplane-system (which he might expect).

	return defaultNamespace
}
