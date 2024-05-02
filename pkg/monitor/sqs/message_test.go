package sqs

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ec2manualv1alpha1 "github.com/crossplane-contrib/provider-aws/apis/ec2/manualv1alpha1"
	route53v1alpha1 "github.com/crossplane-contrib/provider-aws/apis/route53/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/monitor"
)

type j = map[string]any

func Test_EventMessage_events(t *testing.T) {
	jsonStr := `{
		"version": "0",
		"detail": {
			"resultCode":1,
			"requestParameters": {
				"instanceIds": ["i-1234"]
			}
		}
	}`
	msg := j{}
	require.NoError(t, json.Unmarshal([]byte(jsonStr), &msg))
	assert.Equal(t, "0", dig(msg, "version").(string))
	assert.InEpsilon(t, 1.0, dig(msg, "detail", "resultCode").(float64), 0.001)
	assert.Equal(t, []any{"i-1234"}, dig(msg, "detail", "requestParameters", "instanceIds").([]any))
}

func eventsFromJson(t *testing.T, str string) []monitor.Event {
	msg := j{}
	require.NoError(t, json.Unmarshal([]byte(str), &msg))
	return eventsFromMessage(msg)
}

func Test_eventsFromMessage_instanceStateChange(t *testing.T) {
	events := eventsFromJson(t, `{
		"detail-type": "EC2 Instance State-change Notification",
		"detail": {
			"instance-id": "i-1234"
		}
	}`)
	assert.Equal(t, []monitor.Event{
		{
			GVK:          ec2manualv1alpha1.InstanceGroupVersionKind,
			ExternalName: "i-1234",
		},
	}, events)

	// Don't produce duplicates when both resources and detail.instance-id are present
	events = eventsFromJson(t, `{
		"detail-type": "EC2 Instance State-change Notification",
		"resources": [
			"arn:aws:ec2:us-east-1:0123456789ab:instance/i-1234"
		],
		"detail": {
			"instance-id": "i-1234"
		}
	}`)
	assert.Equal(t, []monitor.Event{
		{
			GVK:          ec2manualv1alpha1.InstanceGroupVersionKind,
			ExternalName: "i-1234",
		},
	}, events)
}

func Test_eventsFromMessage_instanceApiCall(t *testing.T) {
	events := eventsFromJson(t, `{
		"detail-type": "AWS API Call via CloudTrail",
		"detail": {
			"requestParameters": {
				"instanceId": "i-1234"
			}
		}
	}`)
	assert.Equal(t, []monitor.Event{
		{
			GVK:          ec2manualv1alpha1.InstanceGroupVersionKind,
			ExternalName: "i-1234",
		},
	}, events)
}

func Test_eventsFromMessage_instancesApiCall(t *testing.T) {
	events := eventsFromJson(t, `{
		"detail-type": "AWS API Call via CloudTrail",
		"detail": {
			"requestParameters": {
				"instancesSet": {
					"items": [
						{"instanceId": "i-1234"},
						{"instanceId": "i-5678"}
					]
				}
			}
		}
	}`)
	assert.Equal(t, []monitor.Event{
		{
			GVK:          ec2manualv1alpha1.InstanceGroupVersionKind,
			ExternalName: "i-1234",
		},
		{
			GVK:          ec2manualv1alpha1.InstanceGroupVersionKind,
			ExternalName: "i-5678",
		},
	}, events)
}

func Test_eventsFromMessage_ec2ApiCall(t *testing.T) {
	events := eventsFromJson(t, `{
		"detail-type": "AWS API Call via CloudTrail",
		"detail": {
			"requestParameters": {
				"resourcesSet": {
					"items": [
						{"resourceId": "i-1234"},
						{"resourceId": "sg-5678"}
					]
				}
			}
		}
	}`)
	assert.Equal(t, []monitor.Event{
		{
			GVK:          ec2manualv1alpha1.InstanceGroupVersionKind,
			ExternalName: "i-1234",
		},
	}, events)
}

func Test_eventsFromMessage_ebsNotification(t *testing.T) {
	events := eventsFromJson(t, `{
		"detail-type": "EBS Volume Notification",
		"resources": [
			"arn:aws:ec2:us-east-1:0123456789ab:volume/vol-01234567",
			"arn:aws:kms:us-east-1:0123456789ab:key/01234567-0123-0123-0123-0123456789ab",
			"arn:aws:ec2:us-east-1:0123456789ab:instance/i-1234"
		]
	}`)
	assert.Equal(t, []monitor.Event{
		{
			GVK:          ec2manualv1alpha1.InstanceGroupVersionKind,
			ExternalName: "i-1234",
		},
	}, events)
}

func Test_eventsFromMessage_route53ApiCall(t *testing.T) {
	events := eventsFromJson(t, `{
		"detail-type": "AWS API Call via CloudTrail",
		"detail": {
			"requestParameters": {
				"changeBatch": {
					"changes": [
						{"resourceRecordSet": {"name": "test1.com"}},
						{"resourceRecordSet": {"name": "test2.com"}}
					]
				}
			}
		}
	}`)
	assert.Equal(t, []monitor.Event{
		{
			GVK:          route53v1alpha1.ResourceRecordSetGroupVersionKind,
			ExternalName: "test1.com",
		},
		{
			GVK:          route53v1alpha1.ResourceRecordSetGroupVersionKind,
			ExternalName: "test2.com",
		},
	}, events)
}
