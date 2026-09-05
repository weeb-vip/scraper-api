package sync

import (
	"context"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/weeb-vip/scraper-api/config"
	"github.com/weeb-vip/scraper-api/internal/db"
	"github.com/weeb-vip/scraper-api/internal/db/repositories/anime"
	"github.com/weeb-vip/scraper-api/internal/db/repositories/thetvdblink"
	"github.com/weeb-vip/scraper-api/internal/eventbus"
	"github.com/weeb-vip/scraper-api/internal/logger"
	"github.com/weeb-vip/scraper-api/internal/services/link_service"
	"github.com/weeb-vip/scraper-api/internal/services/season_backfill"
	"github.com/weeb-vip/scraper-api/internal/services/thetvdb_api"
)

// syncDeriveDelayMs paces the TheTVDB calls the derivation makes. The nightly
// job has all night; the API does not have all night for us.
const syncDeriveDelayMs = 200

// Options changes what a run does. The zero value is the nightly run.
type Options struct {
	// DeriveAll re-derives every anime that has a thetvdbid, instead of only
	// those without a link yet.
	//
	// Off nightly, because the linked ones cost thousands of series fetches to
	// confirm seasons that have not changed, and because a derivation that
	// disagrees with a hand-made link would overwrite it. On when TheTVDB has
	// revised its seasons, or when a change to the matcher should be applied to
	// anime that were linked under the old rules.
	DeriveAll bool
}

func Sync(opts Options) error {
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

	// Derive seasons for anime that have a thetvdbid and no link yet, before
	// publishing. A season derived on this pass gets its link published in the
	// same run, so thetvdb-enrichment pulls that season's episodes and artwork
	// tonight rather than tomorrow night.
	//
	// Only the unlinked ones unless asked otherwise. After the first full pass
	// the remaining work is new shows, and re-deriving every linked anime
	// nightly would spend thousands of TheTVDB calls confirming what it already
	// knows. `sync --derive-all` does the full pass, as does `backfill-seasons`
	// without --only-unlinked.
	//
	// A failure here does not abort the run. Deriving is an improvement to the
	// catalogue; republishing existing links is the job this cron was created
	// for, and it should still happen if TheTVDB is unreachable.
	runner := &season_backfill.Runner{
		DB:    database.DB,
		Anime: animeRepository,
		Links: linkService,
		API:   thetvdb_api.NewTheTVDBApi(cfg.TheTVDBConfig, &http.Client{}),
	}
	derived, err := runner.Run(ctx, season_backfill.Options{
		OnlyUnlinked: !opts.DeriveAll,
		DelayMs:      syncDeriveDelayMs,
	}, func(format string, args ...interface{}) {
		log.Info(fmt.Sprintf(format, args...))
	})
	if err != nil {
		log.Error("season derivation failed, continuing with the sync", zap.Error(err))
	} else {
		log.Info("season derivation finished",
			zap.Bool("deriveAll", opts.DeriveAll),
			zap.Int("processed", derived.Processed),
			zap.Int("linked", derived.Linked),
			zap.Int("skipped", derived.Skipped),
			zap.Int("failed", derived.Failed),
			zap.Int("seriesFetched", derived.SeriesFetched))
	}

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
