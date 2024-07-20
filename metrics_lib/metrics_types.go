package metrics_lib

import (
	"net/http"
)

type MetricsImpl interface {
	HistogramMetric(name string, value float64, labels map[string]string) error
	SummaryMetric(name string, value float64, labels map[string]string) error
	CountMetric(name string, labels map[string]string) error
	GaugeMetric(name string, value float64, labels map[string]string) error
	StandardMetrics
}

type Metrics struct {
	client Client
	rate   float64
}

func NewMetrics(client Client, rate float64) MetricsImpl {
	return &Metrics{
		client: client,
		rate:   rate,
	}
}

func (m *Metrics) HistogramMetric(name string, value float64, labels map[string]string) error {
	return m.client.Histogram(name, value, labels, m.rate)
}

func (m *Metrics) CountMetric(name string, labels map[string]string) error {
	return m.client.Count(name, labels, m.rate)
}

func (m *Metrics) GaugeMetric(name string, value float64, labels map[string]string) error {
	return m.client.Gauge(name, value, labels, m.rate)
}

func (m *Metrics) SummaryMetric(name string, value float64, labels map[string]string) error {
	return m.client.Summary(name, value, labels, m.rate)
}

func (m *Metrics) ResolverMetric(value float64, labels ResolverMetricLabels) error {
	return ResolverMetric(m.client, value, labels)
}

func (m *Metrics) HttpMiddlewareMetric(config HttpMiddlewareMetricConfig) func(http.Handler) http.Handler {
	return HttpMiddlewareMetric(m.client, config, m.rate)
}

func (m *Metrics) ApiMetric(value float64, labels ApiMetricLabels) error {
	return ApiMetric(m.client, value, labels)
}

func (m *Metrics) DatabaseMetric(value float64, labels DatabaseMetricLabels) error {
	return DatabaseMetric(m.client, value, labels)
}

func (m *Metrics) CallMetric(value float64, labels CallMetricLabels) error {
	return CallMetric(m.client, value, labels)
}
