package metrics_lib

type DatabaseMetricMethod = string

const (
	DatabaseMetricMethodInsert DatabaseMetricMethod = "insert"
	DatabaseMetricMethodUpdate DatabaseMetricMethod = "update"
	DatabaseMetricMethodDelete DatabaseMetricMethod = "delete"
	DatabaseMetricMethodSelect DatabaseMetricMethod = "select"
)

type DatabaseMetricLabels struct {
	Service string
	Table   string
	Method  DatabaseMetricMethod
	Result  Result
}

func DatabaseMetric(client Client, value float64, labels DatabaseMetricLabels) error {
	err := client.Histogram("database_query_duration_histogram_milliseconds", value, map[string]string{
		"service": labels.Service,
		"table":   labels.Table,
		"method":  string(labels.Method),
		"result":  labels.Result,
	}, 1)

	return err
}
