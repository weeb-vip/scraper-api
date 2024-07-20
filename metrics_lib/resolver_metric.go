package metrics_lib

type Result = string

type ResolverMetricLabels struct {
	Resolver string
	Service  string
	Protocol string
	Result   Result
}

func ResolverMetric(client Client, value float64, labels ResolverMetricLabels) error {
	err := client.Histogram("resolver_request_duration_histogram_milliseconds", value, map[string]string{
		"resolver": labels.Resolver,
		"result":   labels.Result,
	}, 1)

	return err
}
