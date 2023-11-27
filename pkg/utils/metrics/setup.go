package metrics

import (
	"github.com/crossplane/crossplane-runtime/pkg/feature"
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"

	"github.com/crossplane-contrib/provider-aws/pkg/features"
)

var (
	metricAWSAPICalls = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "aws_api_calls_total",
		Help: "Number of API calls to the AWS API",
	}, []string{"service", "operation", "api_version"})

	MetricAWSAPICallsRec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "aws_api_calls_reconcile_managed_resource_total",
		Help: "Amount of calls to the AWS API produced by controller per reconciliation for every managed resource and controller operation type",
	}, []string{"api_version", "kind", "resource_name", "controller_operation_type"})

	MetricManagedResRec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "managed_resource_reconciles_total",
		Help: "Total amount of reconciliation loops per managed resource",
	}, []string{"api_version", "kind", "resource_name"})
)

type metric interface {
	Describe(chan<- *prometheus.Desc)
	Collect(chan<- prometheus.Metric)
}

// SetupMetrics will register the known Prometheus metrics with controller-runtime's metrics registry
func SetupMetrics(flags *feature.Flags) error {
	metricsList := []metric{
		metricAWSAPICalls,
	}
	managedResourceMetricsList := []metric{
		MetricAWSAPICallsRec,
		MetricManagedResRec,
	}

	if flags.Enabled(features.EnableManagedResourceMetrics) {
		metricsList = append(metricsList, managedResourceMetricsList...)
	}
	for _, m := range metricsList {
		err := metrics.Registry.Register(m)
		if err != nil {
			return err
		}
	}
	return nil
}

// IncAWSAPICall will increment the aws_api_calls_total metric for the specified service, operation, and apiVersion tuple
func IncAWSAPICall(service, operation, apiVersion string) {
	metricAWSAPICalls.WithLabelValues(service, operation, apiVersion).Inc()
}
