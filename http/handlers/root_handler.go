package handlers

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/weeb-vip/scraper-api/config"
	"github.com/weeb-vip/scraper-api/graph"
	"github.com/weeb-vip/scraper-api/graph/generated"
	"github.com/weeb-vip/scraper-api/http/handlers/requestinfo"
	"github.com/weeb-vip/scraper-api/internal/db"
	anime2 "github.com/weeb-vip/scraper-api/internal/db/repositories/anime"
	anime3 "github.com/weeb-vip/scraper-api/internal/db/repositories/anime_episode"
	"github.com/weeb-vip/scraper-api/internal/db/repositories/thetvdblink"
	"github.com/weeb-vip/scraper-api/internal/directives"
	"github.com/weeb-vip/scraper-api/internal/producer"
	"github.com/weeb-vip/scraper-api/internal/services/anime"
	"github.com/weeb-vip/scraper-api/internal/services/episodes"
	"github.com/weeb-vip/scraper-api/internal/services/link_service"
	"github.com/weeb-vip/scraper-api/internal/services/thetvdb_api"
	"github.com/weeb-vip/scraper-api/internal/services/thetvdb_service"
	"net/http"
)

func BuildRootHandler(conf config.Config) http.Handler {
	database := db.NewDatabase(conf.DBConfig)
	animeRepository := anime2.NewAnimeRepository(database)
	episodeRepository := anime3.NewAnimeEpisodeRepository(database)
	animeService := anime.NewAnimeService(animeRepository)
	animeEpisodeService := episodes.NewAnimeEpisodeService(episodeRepository)
	theTVDBAPI := thetvdb_api.NewTheTVDBApi(conf.TheTVDBConfig, &http.Client{})
	theTVDBService := thetvdb_service.NewTheTVDBService(theTVDBAPI)
	theTVDBLinkRepository := thetvdblink.NewTheTVDBLinkRepository(database)
	producerService := producer.NewProducer[link_service.LinkProducerStruct](context.Background(), conf.PulsarConfig)
	linkService := link_service.NewLinkService(theTVDBLinkRepository, producerService)
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

	return srv
}
