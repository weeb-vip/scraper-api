package metrics_lib

import (
	"net/http"
)

const (
	Success Result = "success"
	Error   Result = "error"
)

type StandardMetrics interface {
	ResolverMetric(value float64, labels ResolverMetricLabels) error
	HttpMiddlewareMetric(config HttpMiddlewareMetricConfig) func(http.Handler) http.Handler
	ApiMetric(value float64, labels ApiMetricLabels) error
	DatabaseMetric(value float64, labels DatabaseMetricLabels) error
	CallMetric(value float64, labels CallMetricLabels) error
}
