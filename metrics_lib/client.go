package metrics_lib

type Client interface {
	Histogram(metric string, value float64, labels map[string]string, rate float64) error
	Count(metric string, labels map[string]string, rate float64) error
	Gauge(metric string, value float64, labels map[string]string, rate float64) error
	Summary(metric string, value float64, labels map[string]string, rate float64) error
}
