package resolvers

import (
	"context"
	"github.com/weeb-vip/scraper-api/internal/services/anime"
	"github.com/weeb-vip/scraper-api/internal/services/link_service"
)

func SyncIDs(ctx context.Context, linkService link_service.LinkService, animeService anime.AnimeServiceImpl) (bool, error) {
	// Get all saved links
	links, err := linkService.FindAll(ctx)
	if err != nil {
		return false, err
	}

	// Update each anime with its corresponding thetvdbid
	for _, link := range links {
		if link.AnimeID != "" && link.TheTVDBLinkID != "" {
			// Update the anime's thetvdbid field
			err := animeService.UpdateTheTVDBID(ctx, link.AnimeID, link.TheTVDBLinkID)
			if err != nil {
				// Log error but continue with other links
				// You might want to add proper logging here
				continue
			}
		}
	}

	return true, nil
}