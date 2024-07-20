package resolvers

import (
	"context"
	"github.com/weeb-vip/scraper-api/graph/model"
	"github.com/weeb-vip/scraper-api/internal/services/link_service"
	"github.com/weeb-vip/scraper-api/internal/services/thetvdb_api"
	"github.com/weeb-vip/scraper-api/internal/services/thetvdb_service"
	"strconv"
)

func SearchAnime(ctx context.Context, theTVDBService thetvdb_service.TheTVDBService, title string) ([]*model.TheTVDBAnime, error) {
	results, err := theTVDBService.SearchAnimes(ctx, title)
	if err != nil {
		return nil, err
	}

	var animes []*model.TheTVDBAnime
	for _, result := range *results {
		anime := &model.TheTVDBAnime{
			ID:           *result.TVDBID,
			Title:        *result.Name,
			Link:         "",
			Image:        result.ImageURL,
			Year:         result.Year,
			Studios:      result.Studios,
			Genres:       result.Genres,
			Translations: mapToTranslationTuple(result.Translations),
		}
		animes = append(animes, anime)
	}

	return animes, nil

}

func GetAnimeEpisodes(ctx context.Context, theTVDBService thetvdb_service.TheTVDBService, seriesID string) ([]*model.TheTVDBEpisode, error) {
	episodes, err := theTVDBService.GetEpisodesBySeriesID(ctx, seriesID)
	if err != nil {
		return nil, err
	}

	var theTVDBEpisodes []*model.TheTVDBEpisode
	for _, episode := range *episodes {
		// convert int64 to string
		idStr := strconv.FormatInt(*episode.ID, 10)
		var translation *thetvdb_api.Translation
		translation, ok := episode.Translations["eng"]
		if !ok {
			translation, ok = episode.Translations["jpn"]
			if !ok {
				nameStr := "undefined"
				isAlias := false
				isPrimary := false
				language := "eng"
				overview := "undefined"
				tagline := "undefined"
				translation = &thetvdb_api.Translation{
					Name:      &nameStr,
					Aliases:   make([]string, 0),
					IsAlias:   &isAlias,
					IsPrimary: &isPrimary,
					Overview:  &overview,
					Language:  &language,
					Tagline:   &tagline,
				}
			}
		}
		engTitle := translation.Name
		theTVDBEpisode := &model.TheTVDBEpisode{
			ID:            idStr,
			Title:         *engTitle,
			SeasonNumber:  *episode.SeasonNumber,
			EpisodeNumber: *episode.Number,
			Image:         episode.Image,
			Description:   episode.Overview,
		}
		theTVDBEpisodes = append(theTVDBEpisodes, theTVDBEpisode)
	}

	return theTVDBEpisodes, nil
}

func mapToTranslationTuple(m map[string]interface{}) []*model.TranslationTuple {
	var tuples []*model.TranslationTuple
	for k, v := range m {
		key := k
		vStr := v.(string)
		tuple := &model.TranslationTuple{
			Key:   &key,
			Value: &vStr,
		}
		tuples = append(tuples, tuple)
	}

	return tuples
	//var tuple model.TranslationTuple
	//for k, v := range m {
	//
	//	tuple.Key = &k
	//	// read as a string
	//	vStr := v.(string)
	//	tuple.Value = &vStr
	//}
	//return tuple
}

func GetSavedLinks(ctx context.Context, theTVDBService link_service.LinkService) ([]*model.Link, error) {
	links, err := theTVDBService.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	var theTVDBLinks []*model.Link
	for _, link := range links {
		theTVDBLink := &model.Link{
			ID:        link.ID,
			Name:      link.Name,
			AnimeID:   link.AnimeID,
			ThetvdbID: link.TheTVDBLinkID,
			Season:    link.Season,
		}
		theTVDBLinks = append(theTVDBLinks, theTVDBLink)
	}

	return theTVDBLinks, nil
}
