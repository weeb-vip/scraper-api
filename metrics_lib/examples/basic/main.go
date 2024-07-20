package main

import (
	MetricsLib "github.com/TempMee/go-metrics-lib"
	"github.com/TempMee/go-metrics-lib/clients/datadog"
	"log"
)

type Result string

const (
	ResultSuccess Result = "success"
	ResultError   Result = "error"
)

type Labels struct {
	Name    string
	Service string
	Result  Result
}

func main() {
	datadogClient := datadog.NewDatadogClient(datadog.DataDogConfig{
		DD_AGENT_HOST: "localhost",
		DD_AGENT_PORT: 8125,
	})
	datadogClient.CreateHistogram("graphql.resolver.millisecond", []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}, map[string]string{
		"resolver": "resolver",
		"service":  "graphql",
		"result":   "success",
	}, 1)

	metrics := MetricsLib.NewMetrics(
		datadogClient,
		1,
	)

	err := metrics.HistogramMetric("graphql.resolver.millisecond", 100,
		map[string]string{
			"resolver": "resolver",
			"service":  "graphql",
			"result":   "success",
		},
	)

	if err != nil {
		log.Println("BORKED!")
		panic(err)
	}

	err = metrics.SummaryMetric("graphql.resolver.millisecond", 100, map[string]string{
		"resolver": "resolver",
		"service":  "graphql",
		"result":   "success",
	})

	if err != nil {
		log.Println("BORKED!")

	}

	err = metrics.ResolverMetric(100, MetricsLib.ResolverMetricLabels{
		Resolver: "resolver",
		Result:   MetricsLib.Success,
	})

	if err != nil {
		log.Println("BORKED!")
		panic(err)
	}

}
