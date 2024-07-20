package graph

import (
	"github.com/weeb-vip/scraper-api/config"
	"github.com/weeb-vip/scraper-api/internal/services/anime"
	"github.com/weeb-vip/scraper-api/internal/services/episodes"
	"github.com/weeb-vip/scraper-api/internal/services/link_service"
	"github.com/weeb-vip/scraper-api/internal/services/thetvdb_service"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	Config              config.Config
	AnimeService        anime.AnimeServiceImpl
	AnimeEpisodeService episodes.AnimeEpisodeServiceImpl
	TheTVDBService      thetvdb_service.TheTVDBService
	LinkService         link_service.LinkService
}
