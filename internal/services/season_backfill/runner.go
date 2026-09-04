package season_backfill

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	animerepo "github.com/weeb-vip/scraper-api/internal/db/repositories/anime"
	"github.com/weeb-vip/scraper-api/internal/services/link_service"
	"github.com/weeb-vip/scraper-api/internal/services/thetvdb_api"
)

// Options controls one backfill run.
type Options struct {
	DryRun        bool
	Limit         int
	DelayMs       int
	After         string
	RequiredRatio float64
}

// Result is what a run did, for the caller to log.
type Result struct {
	Processed     int
	Linked        int
	Skipped       int
	Failed        int
	SeriesFetched int
}

// Runner derives seasons and records them through the link service.
//
// It writes links rather than touching anime.season_number directly, because
// that is what the admin panel does and because saving a link is what publishes
// the event thetvdb-enrichment consumes -- so every season derived here also
// gets its episodes and artwork enriched, exactly as a hand-made link would.
type Runner struct {
	DB    *gorm.DB
	Anime animerepo.AnimeRepositoryImpl
	Links link_service.LinkService
	API   thetvdb_api.TheTVDBApi
}

// animeRow is the slice of anime this needs, read directly rather than through
// the anime repository: its interface has no keyset-paged "everything with a
// thetvdbid" read, and adding one for a single command would widen an already
// very wide interface.
// Columns are tagged explicitly. Relying on gorm to derive them meant
// TheTVDBID silently scanned as empty -- the SQL filter still worked, so every
// row looked up the same blank series id, one fetch was cached for all of them
// and every anime came back unmatched.
type animeRow struct {
	ID        string  `gorm:"column:id"`
	TheTVDBID string  `gorm:"column:thetvdbid"`
	TitleEn   *string `gorm:"column:title_en"`
	StartDate *string `gorm:"column:start_date"`
}

// episodeDays reads an anime's episode air days.
//
// Queries anime_episodes by name rather than through the AnimeEpisode entity,
// whose TableName() returns "episodes" -- a table this service's database does
// not have. That mismatch predates this command and is left alone here rather
// than fixed underneath the code that already depends on it.
func (r *Runner) episodeDays(ctx context.Context, animeID string) ([]time.Time, error) {
	var aired []time.Time
	err := r.DB.WithContext(ctx).
		Table("anime_episodes").
		Where("anime_id = ? AND aired IS NOT NULL", animeID).
		Pluck("aired", &aired).Error

	return aired, err
}

func (r *Runner) page(ctx context.Context, after string, limit int) ([]animeRow, error) {
	var rows []animeRow
	err := r.DB.WithContext(ctx).
		Table("anime").
		Select("id, thetvdbid, title_en, start_date").
		Where("thetvdbid IS NOT NULL AND thetvdbid <> '' AND id::text > ?", after).
		Order("id::text").
		Limit(limit).
		Scan(&rows).Error

	return rows, err
}

func (r *Runner) Run(ctx context.Context, opts Options, log func(string, ...interface{})) (Result, error) {
	if opts.RequiredRatio <= 0 {
		opts.RequiredRatio = DefaultRequiredRatio
	}

	type series struct {
		windows []SeasonWindow
		days    map[int]map[string]bool
	}
	// One fetch per series, not per anime: without this, Pokemon alone would be
	// 76 identical requests.
	cache := map[string]series{}

	result := Result{}
	after := opts.After

	for {
		rows, err := r.page(ctx, after, 100)
		if err != nil {
			return result, err
		}
		if len(rows) == 0 {
			break
		}

		for _, row := range rows {
			after = row.ID
			if opts.Limit > 0 && result.Processed >= opts.Limit {
				break
			}
			result.Processed++

			if row.StartDate == nil {
				result.Skipped++
				continue
			}
			start, ok := ParseStartDate(*row.StartDate)
			if !ok {
				result.Skipped++
				continue
			}

			found, cached := cache[row.TheTVDBID]
			if !cached {
				data, err := r.API.GetEpisodesBySeriesID(ctx, row.TheTVDBID)
				if err != nil {
					log("fetch failed series=%s err=%v", row.TheTVDBID, err)
					result.Failed++
					continue
				}
				found = series{
					windows: SeasonWindows(data.Episodes),
					days:    SeasonAirDays(data.Episodes),
				}
				cache[row.TheTVDBID] = found
				result.SeriesFetched++

				if opts.DelayMs > 0 {
					time.Sleep(time.Duration(opts.DelayMs) * time.Millisecond)
				}
			}

			// Episodes first: agreement on air days is evidence, while a start
			// date landing on a season boundary is a coincidence that usually
			// holds and sometimes does not.
			aired, err := r.episodeDays(ctx, row.ID)
			if err != nil {
				log("episode read failed anime=%s err=%v", row.ID, err)
				result.Failed++
				continue
			}

			season, matched := MatchByEpisodes(aired, found.days, opts.RequiredRatio)
			how := "episodes"
			if !matched {
				season, matched = MatchByExactStart(start, found.windows)
				how = "exact-start"
			}
			if !matched {
				result.Skipped++
				continue
			}

			title := ""
			if row.TitleEn != nil {
				title = *row.TitleEn
			}
			name := fmt.Sprintf("%s (season %d)", title, season)
			if title == "" {
				name = fmt.Sprintf("season %d", season)
			}

			log("matched anime=%s series=%s season=%d via=%s episodes=%d",
				row.ID, row.TheTVDBID, season, how, len(aired))

			if opts.DryRun {
				result.Linked++
				continue
			}

			if _, err := r.Links.Save(ctx, row.ID, row.TheTVDBID, season, name); err != nil {
				log("link save failed anime=%s err=%v", row.ID, err)
				result.Failed++
				continue
			}
			result.Linked++
		}

		if opts.Limit > 0 && result.Processed >= opts.Limit {
			break
		}
	}

	return result, nil
}
