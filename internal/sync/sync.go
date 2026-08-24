package sync

import (
	"context"
	"github.com/weeb-vip/scraper-api/config"
	"github.com/weeb-vip/scraper-api/internal/db"
	"github.com/weeb-vip/scraper-api/internal/db/repositories/anime"
	"github.com/weeb-vip/scraper-api/internal/db/repositories/thetvdblink"
	"github.com/weeb-vip/scraper-api/internal/eventbus"
	"github.com/weeb-vip/scraper-api/internal/logger"
	"github.com/weeb-vip/scraper-api/internal/services/link_service"
)

func Sync() error {
	cfg := config.LoadConfigOrPanic()

	ctx := context.Background()
	log := logger.Get()
	ctx = logger.WithCtx(ctx, log)

	database := db.NewDatabase(cfg.DBConfig)
	theTVDBLinkRepository := thetvdblink.NewTheTVDBLinkRepository(database)
	animeRepository := anime.NewAnimeRepository(database)

	// Safe to defer here, unlike the handler builder: this is a job that runs to
	// completion, so close fires at the end of the run rather than immediately.
	publish, closePublisher, err := eventbus.New(cfg)
	if err != nil {
		return err
	}
	defer closePublisher()

	linkService := link_service.NewLinkService(theTVDBLinkRepository, animeRepository, publish)

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
