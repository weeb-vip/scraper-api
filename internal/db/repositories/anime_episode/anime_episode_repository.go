package anime

import (
	"context"
	"github.com/weeb-vip/scraper-api/internal/db"
)

type RECORD_TYPE string

type AnimeEpisodeRepositoryImpl interface {
	Upsert(ctx context.Context, anime *AnimeEpisode) error
	Delete(ctx context.Context, anime *AnimeEpisode) error
	FindByAnimeID(ctx context.Context, animeID string) ([]*AnimeEpisode, error)
}

type AnimeEpisodeRepository struct {
	db *db.DB
}

func NewAnimeEpisodeRepository(db *db.DB) AnimeEpisodeRepositoryImpl {
	return &AnimeEpisodeRepository{db: db}
}

func (a *AnimeEpisodeRepository) Upsert(ctx context.Context, episode *AnimeEpisode) error {
	err := a.db.DB.Save(episode).Error
	if err != nil {
		return err
	}
	return nil
}

func (a *AnimeEpisodeRepository) Delete(ctx context.Context, episode *AnimeEpisode) error {
	err := a.db.DB.Delete(episode).Error
	if err != nil {
		return err
	}
	return nil
}

func (a *AnimeEpisodeRepository) FindByAnimeID(ctx context.Context, animeID string) ([]*AnimeEpisode, error) {
	var episodes []*AnimeEpisode
	err := a.db.DB.Where("anime_id = ?", animeID).Find(&episodes).Error
	if err != nil {
		return nil, err
	}
	return episodes, nil
}
