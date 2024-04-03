package controller

import (
	"testing"
	"time"

	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/google/go-cmp/cmp"
)

func TestOptionsOverrides(t *testing.T) {
	options := NewOptions(controller.Options{
		PollInterval:            2 * time.Minute,
		MaxConcurrentReconciles: 3,
	})
	options.AddOverrides(map[string]string{
		"pollInterval":                    "1m",
		"ec2.instance.pollInterval":       "30s",
		"route53.maxConcurrentReconciles": "5",
	})

	// defaults with overrides
	if diff := cmp.Diff(1*time.Minute, options.Default().PollInterval); diff != "" {
		t.Errorf("default.PollInterval: -want, +got:\n%s", diff)
	}
	if diff := cmp.Diff(3, options.Default().MaxConcurrentReconciles); diff != "" {
		t.Errorf("default.MaxConcurrentReconciles: -want, +got:\n%s", diff)
	}

	// overrides without dot in the scope name
	if diff := cmp.Diff(30*time.Second, options.Get("ec2.instance").PollInterval); diff != "" {
		t.Errorf("ec2.instance.PollInterval: -want, +got:\n%s", diff)
	}
	if diff := cmp.Diff(3, options.Get("ec2.instance").MaxConcurrentReconciles); diff != "" {
		t.Errorf("ec2.instance.MaxConcurrentReconciles: -want, +got:\n%s", diff)
	}

	// overrides without dot in the scope name
	if diff := cmp.Diff(1*time.Minute, options.Get("route53").PollInterval); diff != "" {
		t.Errorf("route53.PollInterval: -want, +got:\n%s", diff)
	}
	if diff := cmp.Diff(5, options.Get("route53").MaxConcurrentReconciles); diff != "" {
		t.Errorf("route53.MaxConcurrentReconciles: -want, +got:\n%s", diff)
	}

	// No overrides
	if diff := cmp.Diff(1*time.Minute, options.Get("sqs").PollInterval); diff != "" {
		t.Errorf("sqs.PollInterval: -want, +got:\n%s", diff)
	}
	if diff := cmp.Diff(3, options.Get("sqs").MaxConcurrentReconciles); diff != "" {
		t.Errorf("sqs.MaxConcurrentReconciles: -want, +got:\n%s", diff)
	}
}
