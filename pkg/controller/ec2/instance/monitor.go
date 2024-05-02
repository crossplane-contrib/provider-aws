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

package instance

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/source"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/ec2/manualv1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/monitor"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/cache"
)

func eventsFromMonitor(mgr ctrl.Manager, mnt monitor.Monitor) source.Source {
	mnt.AddPrepareHook(func(ctx context.Context) error {
		if err := cache.IndexByExternalName(ctx, mgr.GetCache(), &svcapitypes.Instance{}); err != nil {
			return fmt.Errorf("setting up indexer: %w", err)
		}
		return nil
	})

	eventsChannel := make(chan event.GenericEvent)
	mnt.AddSubscriber(func(ctx context.Context, evt monitor.Event) {
		if evt.GVK != svcapitypes.InstanceGroupVersionKind {
			return
		}
		var list svcapitypes.InstanceList
		if err := cache.ListByExternalName(ctx, mgr.GetCache(), &list, evt.ExternalName); err != nil {
			mgr.GetLogger().Error(err, "failed to list objects", "externalName", evt.ExternalName)
			return
		}
		for _, object := range list.Items {
			object := object
			eventsChannel <- event.GenericEvent{Object: &object}
		}
	})
	return &source.Channel{Source: eventsChannel}
}
