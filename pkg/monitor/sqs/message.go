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

package sqs

import (
	"encoding/json"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/sqs/types"

	ec2manualv1alpha1 "github.com/crossplane-contrib/provider-aws/apis/ec2/manualv1alpha1"
	route53v1alpha1 "github.com/crossplane-contrib/provider-aws/apis/route53/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/monitor"
)

func parseEventsFromMessage(message types.Message) ([]monitor.Event, error) {
	if message.Body == nil {
		return nil, nil
	}
	msg := map[string]any{}
	if err := json.Unmarshal([]byte(*message.Body), &msg); err != nil {
		return nil, err
	}
	return eventsFromMessage(msg), nil
}

func eventsFromMessage(msg map[string]any) []monitor.Event { //nolint:gocyclo
	events := []monitor.Event{}

	if resources, ok := dig(msg, "resources").([]any); ok && len(resources) > 0 {
		for _, resource := range resources {
			if arn, ok := resource.(string); ok {
				if event := eventFromARN(arn); (event != monitor.Event{}) {
					events = append(events, event)
				}
			}
		}
		// If resources are present, we don't need to parse the rest of the message
		return events
	}

	// EC2 API call that supports different resource types, like tagging
	if resourcesSet, ok := dig(msg, "detail", "requestParameters", "resourcesSet", "items").([]any); ok {
		for _, item := range resourcesSet {
			if resourceId, ok := dig(item, "resourceId").(string); ok {
				if strings.HasPrefix(resourceId, "i-") {
					events = append(events, monitor.Event{
						GVK:          ec2manualv1alpha1.InstanceGroupVersionKind,
						ExternalName: resourceId,
					})
				}
			}
		}
	}

	// EC2 API call for multiple instances
	if instancesSet, ok := dig(msg, "detail", "requestParameters", "instancesSet", "items").([]any); ok {
		for _, item := range instancesSet {
			if instanceId, ok := dig(item, "instanceId").(string); ok {
				events = append(events, monitor.Event{
					GVK:          ec2manualv1alpha1.InstanceGroupVersionKind,
					ExternalName: instanceId,
				})
			}
		}
	}

	// EC2 API call for a single instance
	if instanceId, ok := dig(msg, "detail", "requestParameters", "instanceId").(string); ok {
		events = append(events, monitor.Event{
			GVK:          ec2manualv1alpha1.InstanceGroupVersionKind,
			ExternalName: instanceId,
		})
	}

	// EC2 Instance State-change Notification
	if instanceId, ok := dig(msg, "detail", "instance-id").(string); ok {
		events = append(events, monitor.Event{
			GVK:          ec2manualv1alpha1.InstanceGroupVersionKind,
			ExternalName: instanceId,
		})
	}

	// Route53 API call
	if changes, ok := dig(msg, "detail", "requestParameters", "changeBatch", "changes").([]any); ok {
		for _, item := range changes {
			if name, ok := dig(item, "resourceRecordSet", "name").(string); ok {
				events = append(events, monitor.Event{
					GVK:          route53v1alpha1.ResourceRecordSetGroupVersionKind,
					ExternalName: name,
				})
			}
		}
	}

	return events
}

// dig is a helper function to return a nested value from a map following the provided path.
func dig(obj any, parts ...string) any {
	for _, part := range parts {
		if strMap, ok := obj.(map[string]any); !ok {
			return nil
		} else {
			obj = strMap[part]
		}
	}
	return obj
}

// Parses ARN and returns Event for the resource.
// ARN example: "arn:aws:ec2:us-east-1:123456789012:instance/i-abcd1111".
func eventFromARN(arn string) monitor.Event {
	parts := strings.Split(arn, ":")
	if len(parts) < 6 {
		return monitor.Event{}
	}
	resourceParts := strings.Split(parts[5], "/")
	if len(resourceParts) < 2 {
		return monitor.Event{}
	}
	service := parts[2]
	resourceType := resourceParts[0]
	name := resourceParts[1]

	if service == "ec2" && resourceType == "instance" {
		return monitor.Event{
			GVK:          ec2manualv1alpha1.InstanceGroupVersionKind,
			ExternalName: name,
		}
	} else if service == "route53" && resourceType == "resourceRecordSet" {
		return monitor.Event{
			GVK:          route53v1alpha1.ResourceRecordSetGroupVersionKind,
			ExternalName: name,
		}
	}
	return monitor.Event{}
}
