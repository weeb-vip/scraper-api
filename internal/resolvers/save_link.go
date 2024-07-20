package resolvers

import (
	"context"
	"github.com/weeb-vip/scraper-api/graph/model"
	"github.com/weeb-vip/scraper-api/internal/services/link_service"
)

func SaveLink(ctx context.Context, linkService link_service.LinkService, animeId string, TVDBID string, season int, name string) (*model.Link, error) {
	link, err := linkService.Save(ctx, animeId, TVDBID, season, name)
	if err != nil {

		return nil, err
	}
	return &model.Link{
		ID:        link.ID,
		AnimeID:   link.AnimeID,
		ThetvdbID: link.TheTVDBLinkID,
		Season:    link.Season,
	}, nil
}
