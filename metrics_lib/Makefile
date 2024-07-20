generate: mocks

mocks:
	go get go.uber.org/mock/mockgen/model
	go install go.uber.org/mock/mockgen@latest
	mockgen -destination=./mocks/metrics_impl.go -package=mocks github.com/TempMee/go-metrics-lib MetricsImpl
	mockgen -destination=./mocks/metrics_client.go -package=mocks github.com/TempMee/go-metrics-lib Client
	mockgen -destination=./mocks/stasd_client.go -package=mocks github.com/DataDog/datadog-go/v5/statsd ClientInterface
