package sync

import (
	"context"
	"github.com/weeb-vip/scraper-api/config"
	"github.com/weeb-vip/scraper-api/internal/db"
	"github.com/weeb-vip/scraper-api/internal/db/repositories/thetvdblink"
	"github.com/weeb-vip/scraper-api/internal/producer"
	"github.com/weeb-vip/scraper-api/internal/services/link_service"
)

func Sync() error {
	ctx := context.Background()
	conf := config.LoadConfigOrPanic()

	database := db.NewDatabase(conf.DBConfig)
	theTVDBLinkRepository := thetvdblink.NewTheTVDBLinkRepository(database)

	producerService := producer.NewProducer[link_service.LinkProducerStruct](ctx, conf.PulsarConfig)

	linkService := link_service.NewLinkService(theTVDBLinkRepository, producerService)

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
