package datadog

import (
	"fmt"
	"strings"
)

type Histogram struct {
	MetricName string
	Buckets    []float64
	Labels     map[string]string
	Rate       float64
}

func NewHistogram(metricName string, buckets []float64, labels map[string]string, rate float64) *Histogram {
	return &Histogram{
		MetricName: metricName,
		Buckets:    buckets,
		Labels:     labels,
		Rate:       rate,
	}
}

// trimFloat remove excess 0s from float
func trimFloat(s string) string {
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}

// GenerateMetric generates a metric based on the value and labels
// buckets are filled based on the value such that 5 fills 10, 25 fills 30, 50 fills +Inf
func (h *Histogram) GenerateMetric(value float64, labels map[string]string, rate float64) (Histogram, error) {
	// set le label related to value in buckets
	le := ""
	for _, bucket := range h.Buckets {
		if value <= bucket {
			le = trimFloat(fmt.Sprintf("%f", bucket))
			break
		}
	}

	if le == "" {
		le = "+Inf"
	}

	// set labels
	h.Labels = labels
	h.Labels["le"] = le
	h.Rate = rate

	return *h, nil

}
