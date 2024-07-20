package metrics_lib

import (
	"log"
	"net/http"
	"time"
)

type HttpMiddlewareMetricConfig struct {
	Service string
}

func HttpMiddlewareMetric(client Client, config HttpMiddlewareMetricConfig, rate float64) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return measurementHandler(client, config, rate, h)
	}
}

func measurementHandler(client Client, config HttpMiddlewareMetricConfig, rate float64, next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		startTime := time.Now()

		defer func() {
			elasped := time.Since(startTime).Milliseconds()
			err := client.Histogram("http_request_duration_histogram_milliseconds", float64(elasped),
				map[string]string{
					"service": config.Service,
					"method":  r.Method,
					"result":  "success",
				}, rate)
			if err != nil {
				log.Fatal(err)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
