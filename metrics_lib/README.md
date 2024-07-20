# Go Metrics Library

The purpose of this library is to make it easier to write metrics and provide a standard for usage of metrics. This
library currently supports these standard metrics:

## Standard Metrics

| Metric                                           | Labels                                                                                                                                               | Description                                                                                                                                 | |
|:-------------------------------------------------|:-----------------------------------------------------------------------------------------------------------------------------------------------------|:--------------------------------------------------------------------------------------------------------------------------------------------|:-|
| resolver_request_duration_histogram_milliseconds | resolver= function name of the resolver<br/>result= success \|fail<br/>service= name of the service<br/>protocol= http \|grpc \|graphql              | This metric gives an overview of success/failures of resolvers, the duration of resolvers, and the distribution of the duration of requests | |
| http_request_duration_histogram_milliseconds     | result= success \|fail<br/>service= name of the service<br/>method= POST\|GET\|PATCH…                                                                | all http requests to our service (datadog gives to us for free).                                                                            | |
| api_request_duration_histogram_milliseconds      | service= current service<br/>vendor= internal or external vendor<br/>call= name of the query being called (function name)<br/>result=success \| fail | Calculating communication between services or vendors, where they came from, where they are meant to go, duration of request.               | |
| database_query_duration_histogram_milliseconds   | service= service name<br/>result= success \|fail<br/>table= table name<br/>method= insert \|delete \|find<br/>database= mongodb \|postgres           | Getting duration of queries in respect to the service they are in.                                                                          | |
| call_duration_histogram_milliseconds             | service= service name<br/>result= success \|fail<br/>function= function name                                                                         | Looking at the duration of a call for a function (not for every function, used for things we want to watch)                                 | |


### Usage of metrics specific to clients

Some clients will provide their own metrics, such as the prometheus client. Additionally, metrics like Histograms 
support more details such as buckets. In the case of histograms, Buckets need to be set initially to keep consistency in
metrics.

How to setup metric:
```go
datadogClient.CreateHistogram("graphql.resolver.millisecond", []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}, map[string]string{
    "resolver": "resolver",
    "service":  "graphql",
    "result":   "success",
}, 1)
```

How to use metric:
```go
metrics := MetricsLib.NewMetrics(
    datadogClient,
    1,
)
err := metrics.HistogramMetric("graphql.resolver.millisecond", 100, // if metric not created, will have empty buckets (le:+Inf) 
    map[string]string{
        "resolver": "resolver",
        "service":  "graphql",
        "result":   "success",
    },
)
```

Reference to metrics is done through the metric name, thus metrics cannot be overwritten. You may add on more labels 
and overwrite label values, however labels initially set cannot be removed during runtime.

## Examples

See examples in examples folder.

Example Usage:

```go
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
	err := datadogClient.CreateHistogram("graphql.resolver.millisecond", []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}, map[string]string{
		"resolver": "resolver",
		"service":  "graphql",
		"result":   "success",
	}, 1)
	if err != nil {
		log.Println("Failed to create histogram")
	}

	metrics := MetricsLib.NewMetrics(
		datadogClient,
		1,
	)

	err = metrics.HistogramMetric("graphql.resolver.millisecond", 100,
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


```

# References for Metrics
https://prometheus.io/docs/concepts/metric_types/

## Best Practices (applies to datadog as well)

https://prometheus.io/docs/practices/naming/
https://prometheus.io/docs/practices/consoles/
https://prometheus.io/docs/practices/instrumentation/
https://prometheus.io/docs/practices/histograms/
https://prometheus.io/docs/practices/alerting/
https://prometheus.io/docs/practices/rules/