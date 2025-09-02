package sync

import (
	"context"
	"github.com/ThatCatDev/ep/v2/drivers"
	epKafka "github.com/ThatCatDev/ep/v2/drivers/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/weeb-vip/scraper-api/config"
	"github.com/weeb-vip/scraper-api/internal/db"
	"github.com/weeb-vip/scraper-api/internal/db/repositories/anime"
	"github.com/weeb-vip/scraper-api/internal/db/repositories/thetvdblink"
	"github.com/weeb-vip/scraper-api/internal/logger"
	"github.com/weeb-vip/scraper-api/internal/services/link_service"
	"go.uber.org/zap"
)

func Sync() error {
	cfg := config.LoadConfigOrPanic()

	ctx := context.Background()
	log := logger.Get()
	ctx = logger.WithCtx(ctx, log)

	database := db.NewDatabase(cfg.DBConfig)
	theTVDBLinkRepository := thetvdblink.NewTheTVDBLinkRepository(database)
	animeRepository := anime.NewAnimeRepository(database)

	kafkaConfig := &epKafka.KafkaConfig{
		ConsumerGroupName:        cfg.KafkaConfig.ConsumerGroupName,
		BootstrapServers:         cfg.KafkaConfig.BootstrapServers,
		SaslMechanism:            nil,
		SecurityProtocol:         nil,
		Username:                 nil,
		Password:                 nil,
		ConsumerSessionTimeoutMs: nil,
		ConsumerAutoOffsetReset:  &cfg.KafkaConfig.Offset,
		ClientID:                 nil,
		Debug:                    nil,
	}

	driver := epKafka.NewKafkaDriver(kafkaConfig)
	defer func(driver drivers.Driver[*kafka.Message]) {
		err := driver.Close()
		if err != nil {
			log.Error("Error closing Kafka driver", zap.String("error", err.Error()))
		} else {
			log.Info("Kafka driver closed successfully")
		}
	}(driver)

	linkService := link_service.NewLinkService(theTVDBLinkRepository, animeRepository, kafkaProducer(ctx, driver, cfg.KafkaConfig.ProducerTopic))

	// get all links
	theTVDBLinks, err := theTVDBLinkRepository.FindAll(ctx)

	if err != nil {
		return err
	}

	for _, link := range theTVDBLinks {
		err = linkService.Sync(ctx, link.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func kafkaProducer(ctx context.Context, driver drivers.Driver[*kafka.Message], topic string) func(ctx context.Context, message *kafka.Message) error {
	return func(ctx context.Context, message *kafka.Message) error {
		log := logger.FromCtx(ctx)
		log.Info("Producing message to Kafka", zap.String("topic", topic), zap.String("key", string(message.Key)), zap.String("value", string(message.Value)))
		if err := driver.Produce(ctx, topic, message); err != nil {
			log.Error("Failed to produce message", zap.String("topic", topic), zap.Error(err))
			return err
		}
		return nil
	}
}
