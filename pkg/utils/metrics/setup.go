package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	k8smetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	metricAWSAPICalls = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "aws_api_calls_total",
		Help: "Number of API calls to the AWS API",
	}, []string{"service", "operation", "api_version"})
)

// SetupMetrics will register the known Prometheus metrics with controller-runtime's metrics registry
func SetupMetrics() error {
	return k8smetrics.Registry.Register(metricAWSAPICalls)
}

// IncAWSAPICall will increment the aws_api_calls_total metric for the specified service, operation, and apiVersion tuple
func IncAWSAPICall(service, operation, apiVersion string) {
	metricAWSAPICalls.WithLabelValues(service, operation, apiVersion).Inc()
}
