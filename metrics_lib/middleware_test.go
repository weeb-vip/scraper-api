package metrics_lib_test

import (
	metrics_lib "github.com/TempMee/go-metrics-lib"
	"github.com/TempMee/go-metrics-lib/mocks"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// create handler that waits for 1 second
func waitHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
	})
}

func TestMetrics_HttpMiddleware(t *testing.T) {
	t.Run("TestMetrics_HttpMiddleware", func(t *testing.T) {
		//a := assert.New(t)

		ctrl := gomock.NewController(t)

		client := mocks.NewMockClient(ctrl)

		client.EXPECT().Histogram(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
			func(metric string, value float64, labels map[string]string, rate float64) error {
				if metric != "http_request_duration_histogram_milliseconds" {
					t.Errorf("metric name is not http_request_duration_seconds")
				}
				if value == float64(0) {
					t.Errorf("metric value is not 0.0")
				}
				return nil
			},
		)

		metrics := metrics_lib.NewMetrics(client, 1.0)

		req, err := http.NewRequest("GET", "/health-check", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()

		handler := metrics.HttpMiddlewareMetric(
			metrics_lib.HttpMiddlewareMetricConfig{
				Service: "test",
			})(waitHandler())

		handler.ServeHTTP(rr, req)
	})
}
