package link_service

import (
	"context"
	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/weeb-vip/scraper-api/internal/db/repositories/thetvdblink"
)

type LinkProducerStruct struct {
	Id            string `json:"id"`
	AnimeID       string `json:"anime_id"`
	TheTVDBLinkID string `json:"thetvdb_link_id"`
	Season        int    `json:"season"`
}

type PayloadStruct struct {
	Data LinkProducerStruct `json:"data"`
}

type LinkService interface {
	FindAll(ctx context.Context) ([]*thetvdblink.TheTVDBLink, error)
	FindById(ctx context.Context, id string) (*thetvdblink.TheTVDBLink, error)
	Save(ctx context.Context, animeId string, TVDBID string, season int, name string) (*thetvdblink.TheTVDBLink, error)
	Sync(ctx context.Context, id string) error
}

type Link struct {
	repo     thetvdblink.TheTVDBLinkRepositoryImpl
	Producer func(ctx context.Context, message *kafka.Message) error
}

func NewLinkService(repo thetvdblink.TheTVDBLinkRepositoryImpl, producer func(ctx context.Context, message *kafka.Message) error) LinkService {
	return &Link{repo: repo, Producer: producer}
}

func (l *Link) FindAll(ctx context.Context) ([]*thetvdblink.TheTVDBLink, error) {
	links, err := l.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	return links, nil
}

func (l *Link) FindById(ctx context.Context, id string) (*thetvdblink.TheTVDBLink, error) {
	link, err := l.repo.FindById(ctx, id)
	if err != nil {
		return nil, err
	}
	return link, nil
}

func (l *Link) Save(ctx context.Context, animeId string, TVDBID string, season int, name string) (*thetvdblink.TheTVDBLink, error) {
	link, err := l.repo.Save(ctx, animeId, TVDBID, season, name)
	if err != nil {
		return nil, err
	}
	return link, nil
}

func (l *Link) Sync(ctx context.Context, id string) error {
	link, err := l.repo.FindById(ctx, id)
	if err != nil {
		return err
	}

	jsonLink, err := json.Marshal(PayloadStruct{
		Data: LinkProducerStruct{
			Id:            link.ID,
			AnimeID:       link.AnimeID,
			TheTVDBLinkID: link.TheTVDBLinkID,
			Season:        link.Season,
		},
	})
	// convert to bytes

	return l.Producer(ctx, &kafka.Message{
		Value: jsonLink,
	})
}
