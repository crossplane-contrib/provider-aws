/*
Copyright 2024 The Crossplane Authors.

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

package monitor

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// Monitor monitors resources and notifies subscribers when resources are changed.
type Monitor interface {
	manager.Runnable
	// AddSubscriber adds a subscriber to the monitor.
	AddSubscriber(Subscriber)
	// Prepare prepares the monitor for running, and runs registered prepare hooks.
	Prepare(context.Context) error
	// AddPrepareHook adds a prepare hook to the monitor.
	// Hooks can be used to setup field indexers with correct context,
	// as they cannot be configured with existing controller-runtime APIs:
	// controller cannot register a code to be run before event sources are initialized.
	AddPrepareHook(PrepareHook)
}

type Subscriber func(context.Context, Event)
type PrepareHook func(context.Context) error

// Event represents an external resource event.
type Event struct {
	GVK          schema.GroupVersionKind
	ExternalName string
}
