package eventbus

import (
	"context"
	"fmt"
	"strings"

	"github.com/ThatCatDev/ep/v2/drivers"
	epKafka "github.com/ThatCatDev/ep/v2/drivers/kafka"
	epNats "github.com/ThatCatDev/ep/v2/drivers/nats"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/weeb-vip/scraper-api/config"
)

// Publish sends one encoded event to the configured destination.
//
// It takes the value rather than a driver message because the two transports
// name that field differently -- Value on kafka.Message, Data on the NATS one --
// and every call site here only ever set the value.
type Publish func(ctx context.Context, value []byte) error

// New builds the producer named by PRODUCER_TYPE and returns it with a close.
//
// A runtime switch rather than a separate command: this producer is built both
// inside the GraphQL handler and inside the sync job, so there is no single
// process to name the way the standalone consumers have. Staging moves by
// setting PRODUCER_TYPE=nats while production stays on the default.
//
// Unknown values are rejected rather than falling back to Kafka. A producer
// writing to a broker nobody reads fails silently -- no errors, the events just
// never arrive -- so a typo must not be allowed to look like a healthy service.
//
// The caller owns the returned close. Note that a handler builder must NOT
// defer it: BuildRootHandler returns an http.Handler and runs once at startup,
// so deferring there closes the driver before the first request.
func New(conf config.Config) (Publish, func(), error) {
	switch strings.ToLower(strings.TrimSpace(conf.ProducerType)) {
	case "", "kafka":
		driver := epKafka.NewKafkaDriver(&epKafka.KafkaConfig{
			ConsumerGroupName:       conf.KafkaConfig.ConsumerGroupName,
			BootstrapServers:        conf.KafkaConfig.BootstrapServers,
			ConsumerAutoOffsetReset: &conf.KafkaConfig.Offset,
		})
		topic := conf.KafkaConfig.ProducerTopic

		return kafkaPublish(driver, topic), func() { _ = driver.Close() }, nil

	case "nats":
		driver := epNats.NewNatsDriver(&epNats.Config{
			URL:        conf.NatsConfig.URL,
			StreamName: conf.NatsConfig.StreamName,
		})
		subject := conf.NatsConfig.ProducerSubject

		return natsPublish(driver, subject), func() { _ = driver.Close() }, nil

	default:
		return nil, nil, fmt.Errorf("unknown PRODUCER_TYPE %q: expected kafka or nats", conf.ProducerType)
	}
}

func kafkaPublish(driver drivers.Driver[*kafka.Message], topic string) Publish {
	return func(ctx context.Context, value []byte) error {
		return driver.Produce(ctx, topic, &kafka.Message{Value: value})
	}
}

func natsPublish(driver drivers.Driver[*epNats.Message], subject string) Publish {
	return func(ctx context.Context, value []byte) error {
		return driver.Produce(ctx, subject, &epNats.Message{Data: value})
	}
}
