package episodes

import (
	"context"
	animeEpisode "github.com/weeb-vip/scraper-api/internal/db/repositories/anime_episode"
)

type AnimeEpisodeServiceImpl interface {
	GetEpisodesByAnimeID(ctx context.Context, animeID string) ([]*animeEpisode.AnimeEpisode, error)
}

type AnimeEpisodeService struct {
	Repository animeEpisode.AnimeEpisodeRepositoryImpl
}

func NewAnimeEpisodeService(repository animeEpisode.AnimeEpisodeRepositoryImpl) AnimeEpisodeServiceImpl {
	return &AnimeEpisodeService{
		Repository: repository,
	}
}

func (a *AnimeEpisodeService) GetEpisodesByAnimeID(ctx context.Context, animeID string) ([]*animeEpisode.AnimeEpisode, error) {
	return a.Repository.FindByAnimeID(ctx, animeID)
}
