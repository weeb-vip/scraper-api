package main

import (
	MetricsLib "github.com/TempMee/go-metrics-lib"
	"github.com/TempMee/go-metrics-lib/clients/prometheus"
	"log"
	"math/rand"
	"reflect"
	"strings"
	"sync"
	"time"
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

func labelsToMapString(labels any) map[string]string {
	values := reflect.ValueOf(labels)
	types := values.Type()
	tags := make(map[string]string)
	for i := 0; i < values.NumField(); i++ {
		tags[strings.ToLower(types.Field(i).Name)] = values.Field(i).String()
	}

	return tags
}

func main() {
	prometheusClient := prometheus.NewPrometheusClient()
	// create wait group
	var wg sync.WaitGroup

	go func() {
		wg.Add(1)
		defer wg.Done()
		prometheusClient.ServerHandler()
		log.Println("Server Finished")
	}()
	metrics := MetricsLib.NewMetrics(
		prometheusClient,
		1,
	)

	prometheusClient.CreateHistogramVec("graphql_resolver_millisecond4", "graphql resolver millisecond", []string{"resolver", "service", "result"}, []float64{
		// create buckets 10000 split into 10 buckets
		1000,
		2000,
		3000,
		4000,
		5000,
		6000,
		7000,
		8000,
		9000,
		10000,
	})

	go func() {
		// run loop for every every 1-3 seconds
		for {
			successFailure := rand.Intn(2)
			if successFailure == 0 {
				err := metrics.HistogramMetric("graphql_resolver_millisecond4", float64(rand.Intn(10000)),
					map[string]string{
						"resolver": "resolver",
						"service":  "graphql",
						"result":   "error",
					},
				)

				if err != nil {
					log.Println("BORKED!")
					panic(err)
				}

				_ = metrics.CallMetric(float64(rand.Intn(10000)), MetricsLib.CallMetricLabels{
					Service:  "graphql",
					Function: "resolver",
					Result:   MetricsLib.Error,
				})
				time.Sleep(time.Duration(float64(rand.Intn(5))) * time.Second)
				continue
			}
			err := metrics.HistogramMetric("graphql_resolver_millisecond4", float64(rand.Intn(10000)),
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

			_ = metrics.CallMetric(float64(rand.Intn(10000)), MetricsLib.CallMetricLabels{
				Service:  "graphql",
				Function: "resolver",
				Result:   MetricsLib.Success,
			})

			time.Sleep(time.Duration(float64(rand.Intn(5))) * time.Second)
		}
	}()

	wg.Wait()
}
