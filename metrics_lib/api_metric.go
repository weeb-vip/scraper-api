package metrics_lib

type ApiMetricLabels struct {
	Service string
	Vendor  string
	Call    string
	Result  Result
}

func ApiMetric(client Client, value float64, labels ApiMetricLabels) error {
	err := client.Histogram("api_request_duration_histogram_milliseconds", value, map[string]string{
		"service": labels.Service,
		"vendor":  labels.Vendor,
		"call":    labels.Call,
		"result":  labels.Result,
	}, 1)

	return err
}
