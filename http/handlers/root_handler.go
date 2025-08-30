package handlers

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/ThatCatDev/ep/v2/drivers"
	epKafka "github.com/ThatCatDev/ep/v2/drivers/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/weeb-vip/scraper-api/config"
	"github.com/weeb-vip/scraper-api/graph"
	"github.com/weeb-vip/scraper-api/graph/generated"
	"github.com/weeb-vip/scraper-api/http/handlers/logger"
	"github.com/weeb-vip/scraper-api/http/handlers/requestinfo"
	"github.com/weeb-vip/scraper-api/internal/db"
	anime2 "github.com/weeb-vip/scraper-api/internal/db/repositories/anime"
	anime3 "github.com/weeb-vip/scraper-api/internal/db/repositories/anime_episode"
	"github.com/weeb-vip/scraper-api/internal/db/repositories/thetvdblink"
	"github.com/weeb-vip/scraper-api/internal/directives"
	logger2 "github.com/weeb-vip/scraper-api/internal/logger"
	"github.com/weeb-vip/scraper-api/internal/services/anime"
	"github.com/weeb-vip/scraper-api/internal/services/episodes"
	"github.com/weeb-vip/scraper-api/internal/services/link_service"
	"github.com/weeb-vip/scraper-api/internal/services/thetvdb_api"
	"github.com/weeb-vip/scraper-api/internal/services/thetvdb_service"
	"go.uber.org/zap"
	"net/http"
)

func BuildRootHandler(conf config.Config) http.Handler {
	log := logger2.Get()
	kafkaConfig := &epKafka.KafkaConfig{
		ConsumerGroupName:        conf.KafkaConfig.ConsumerGroupName,
		BootstrapServers:         conf.KafkaConfig.BootstrapServers,
		SaslMechanism:            nil,
		SecurityProtocol:         nil,
		Username:                 nil,
		Password:                 nil,
		ConsumerSessionTimeoutMs: nil,
		ConsumerAutoOffsetReset:  &conf.KafkaConfig.Offset,
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

	database := db.NewDatabase(conf.DBConfig)
	animeRepository := anime2.NewAnimeRepository(database)
	episodeRepository := anime3.NewAnimeEpisodeRepository(database)
	animeService := anime.NewAnimeService(animeRepository)
	animeEpisodeService := episodes.NewAnimeEpisodeService(episodeRepository)
	theTVDBAPI := thetvdb_api.NewTheTVDBApi(conf.TheTVDBConfig, &http.Client{})
	theTVDBService := thetvdb_service.NewTheTVDBService(theTVDBAPI)
	theTVDBLinkRepository := thetvdblink.NewTheTVDBLinkRepository(database)

	linkService := link_service.NewLinkService(theTVDBLinkRepository, kafkaProducer(context.Background(), driver, conf.KafkaConfig.ProducerTopic))
	resolvers := &graph.Resolver{
		Config:              conf,
		AnimeService:        animeService,
		AnimeEpisodeService: animeEpisodeService,
		TheTVDBService:      theTVDBService,
		LinkService:         linkService,
	}

	cfg := generated.Config{Resolvers: resolvers, Directives: directives.GetDirectives()}
	cfg.Directives.Authenticated = func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
		req := requestinfo.FromContext(ctx)

		if req.UserID == nil {
			// unauthorized
			return nil, fmt.Errorf("Access denied")
		}

		return next(ctx)
	}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(cfg))

	return requestinfo.Handler()(logger.Handler()(srv))
}

func kafkaProducer(ctx context.Context, driver drivers.Driver[*kafka.Message], topic string) func(ctx context.Context, message *kafka.Message) error {
	return func(ctx context.Context, message *kafka.Message) error {
		log := logger2.FromCtx(ctx)
		log.Info("Producing message to Kafka", zap.String("topic", topic), zap.String("key", string(message.Key)), zap.String("value", string(message.Value)))
		if err := driver.Produce(ctx, topic, message); err != nil {
			log.Error("Failed to produce message", zap.String("topic", topic), zap.Error(err))
			return err
		}
		return nil
	}
}
