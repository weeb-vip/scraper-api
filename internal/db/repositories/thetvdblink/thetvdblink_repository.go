package thetvdblink

import (
	"context"
	metrics_lib "github.com/TempMee/go-metrics-lib"
	"github.com/google/uuid"
	"github.com/weeb-vip/scraper-api/internal/db"
	"github.com/weeb-vip/scraper-api/metrics"
	"time"
)

type TheTVDBLinkRepositoryImpl interface {
	FindAll(ctx context.Context) ([]*TheTVDBLink, error)
	FindById(ctx context.Context, id string) (*TheTVDBLink, error)
	Save(ctx context.Context, animeId string, TVDBID string, season int, name string) (*TheTVDBLink, error)
}

type TheTVDBLinkRepository struct {
	db *db.DB
}

func NewTheTVDBLinkRepository(db *db.DB) TheTVDBLinkRepositoryImpl {
	return &TheTVDBLinkRepository{db: db}
}

func (a *TheTVDBLinkRepository) FindAll(ctx context.Context) ([]*TheTVDBLink, error) {
	startTime := time.Now()

	var links []*TheTVDBLink
	err := a.db.DB.Find(&links).Error
	if err != nil {
		_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
			Service: "scraper-api",
			Table:   "thetvdb_link",
			Method:  metrics_lib.DatabaseMetricMethodSelect,
			Result:  metrics_lib.Error,
		})
		return nil, err
	}

	_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
		Service: "scraper-api",
		Table:   "thetvdb_link",
		Method:  metrics_lib.DatabaseMetricMethodSelect,
		Result:  metrics_lib.Success,
	})
	return links, nil
}

func (a *TheTVDBLinkRepository) FindById(ctx context.Context, id string) (*TheTVDBLink, error) {
	startTime := time.Now()

	var link TheTVDBLink
	err := a.db.DB.Where("id = ?", id).First(&link).Error
	if err != nil {
		_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
			Service: "scraper-api",
			Table:   "thetvdb_link",
			Method:  metrics_lib.DatabaseMetricMethodSelect,
			Result:  metrics_lib.Error,
		})
		return nil, err
	}

	_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
		Service: "link-api",
		Table:   "thetvdb_link",
		Method:  metrics_lib.DatabaseMetricMethodSelect,
		Result:  metrics_lib.Success,
	})
	return &link, nil
}

func (a *TheTVDBLinkRepository) Save(ctx context.Context, animeId string, TVDBID string, season int, name string) (*TheTVDBLink, error) {
	startTime := time.Now()

	// create uuid
	uuidString := uuid.New().String()

	// find if link already exists
	var existing TheTVDBLink
	err := a.db.DB.Where("anime_id = ? AND thetvdb_id = ?", animeId, TVDBID, season).First(&existing).Error
	if err == nil {
		uuidString = existing.ID
	}
	link := &TheTVDBLink{
		ID:            uuidString,
		AnimeID:       animeId,
		TheTVDBLinkID: TVDBID,
		Season:        season,
		Name:          name,
	}

	err = a.db.DB.Save(link).Error
	if err != nil {
		_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
			Service: "scraper-api",
			Table:   "thetvdb_link",
			Method:  metrics_lib.DatabaseMetricMethodInsert,
			Result:  metrics_lib.Error,
		})
		return nil, err

	}

	_ = metrics.NewMetricsInstance().DatabaseMetric(float64(time.Since(startTime).Milliseconds()), metrics_lib.DatabaseMetricLabels{
		Service: "scraper-api",
		Table:   "thetvdb_link",
		Method:  metrics_lib.DatabaseMetricMethodInsert,
		Result:  metrics_lib.Success,
	})

	return link, nil

}
