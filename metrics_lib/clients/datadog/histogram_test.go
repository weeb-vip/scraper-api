package datadog_test

import (
	"github.com/TempMee/go-metrics-lib/clients/datadog"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHistogram_GenerateMetric(t *testing.T) {
	metric := datadog.NewHistogram("test", []float64{10, 20, 30}, map[string]string{"test": "test"}, 1.0)
	t.Run("Should generate metric and set bucket to 10 for value 5", func(t *testing.T) {
		a := assert.New(t)

		val, err := metric.GenerateMetric(1, map[string]string{"test": "test"}, 1.0)

		a.NoError(err)
		a.Equal(val.Labels["le"], "10")

	})
	t.Run("Should generate metric and set bucket to 10 for value 10", func(t *testing.T) {
		a := assert.New(t)

		val, err := metric.GenerateMetric(10, map[string]string{"test": "test"}, 1.0)

		a.NoError(err)
		a.Equal(val.Labels["le"], "10")

	})

	t.Run("Should generate metric and set bucket to 30 for value 25", func(t *testing.T) {
		a := assert.New(t)

		val, err := metric.GenerateMetric(25, map[string]string{"test": "test"}, 1.0)

		a.NoError(err)
		a.Equal(val.Labels["le"], "30")

	})

	t.Run("Verify to trim float", func(t *testing.T) {
		a := assert.New(t)

		metric := datadog.NewHistogram("test", []float64{10.50, 20, 30}, map[string]string{"test": "test"}, 1.0)
		val, err := metric.GenerateMetric(1, map[string]string{"test": "test"}, 1.0)

		a.NoError(err)
		a.Equal(val.Labels["le"], "10.5")

	})

	t.Run("Should generate metric with no buckets", func(t *testing.T) {
		a := assert.New(t)

		metric := datadog.NewHistogram("test", []float64{}, map[string]string{"test": "test"}, 1.0)
		val, err := metric.GenerateMetric(1, map[string]string{"test": "test"}, 1.0)

		a.NoError(err)
		a.Equal(val.Labels["le"], "+Inf")
	})

}
