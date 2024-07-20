package metrics_lib

type CallMetricLabels struct {
	Service  string
	Function string
	Result   Result
}

func CallMetric(client Client, value float64, labels CallMetricLabels) error {
	err := client.Histogram("call_duration_histogram_milliseconds", value, map[string]string{
		"service":  labels.Service,
		"function": labels.Function,
		"result":   labels.Result,
	}, 1)

	return err
}
