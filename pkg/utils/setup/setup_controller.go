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

package setup

import (
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	ctrl "sigs.k8s.io/controller-runtime"
)

// SetupControllerFnXp is a delegate to initialize a controller with crossplane Options.
type SetupControllerFnXp func(ctrl.Manager, controller.Options) error //nolint:golint

// SetupControllers is a shortcut to call a list of SetupControllerFns with mgr
// and o.
func SetupControllers(mgr ctrl.Manager, o controller.Options, setups ...SetupControllerFnXp) error { //nolint:golint
	for _, setup := range setups {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
