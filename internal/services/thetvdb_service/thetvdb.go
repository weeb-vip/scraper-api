package thetvdb_service

import (
	"context"
	"github.com/weeb-vip/scraper-api/internal/services/thetvdb_api"
	"strconv"
	"sync"
)

type AnimeEpisodeWithTranslation struct {
	thetvdb_api.EpisodeBaseRecord
	Translations map[string]*thetvdb_api.Translation
}

type AnimeWithEpisodes struct {
	thetvdb_api.SearchResult
	Episodes []AnimeEpisodeWithTranslation
}

type TheTVDBService interface {
	FindAnime(ctx context.Context, title string, year string) (*AnimeWithEpisodes, error)
	SearchAnimes(ctx context.Context, title string) (*[]thetvdb_api.SearchResult, error)
	GetEpisodesBySeriesID(ctx context.Context, seriesID string) (*[]AnimeEpisodeWithTranslation, error)
}

type TheTVDBServiceImpl struct {
	api thetvdb_api.TheTVDBApi
}

func NewTheTVDBService(api thetvdb_api.TheTVDBApi) TheTVDBService {
	return &TheTVDBServiceImpl{
		api: api,
	}
}

func (s *TheTVDBServiceImpl) SearchAnimes(ctx context.Context, title string) (*[]thetvdb_api.SearchResult, error) {
	results, err := s.api.FindAnimeByTitle(ctx, title)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (s *TheTVDBServiceImpl) FindAnime(ctx context.Context, title string, year string) (*AnimeWithEpisodes, error) {
	results, err := s.api.FindAnimeByTitle(ctx, title)
	if err != nil {
		return nil, err
	}
	var animeResult *thetvdb_api.SearchResult
	for _, result := range *results {
		if result.Year != nil && *result.Year == year {
			animeResult = &result
		}
	}
	if animeResult != nil {
		animeEpisodes, err := s.getAnimeEpisodes(ctx, *animeResult.TVDBID)
		if err != nil {
			return nil, err
		}
		return &AnimeWithEpisodes{
			SearchResult: *animeResult,
			Episodes:     *animeEpisodes,
		}, nil

	}

	return nil, nil
}

func (s *TheTVDBServiceImpl) getAnimeEpisodes(ctx context.Context, seriesID string) (*[]AnimeEpisodeWithTranslation, error) {
	episodes, err := s.api.GetEpisodesBySeriesID(ctx, seriesID)
	if err != nil {
		return nil, err
	}

	episodesWithTranslations := make([]AnimeEpisodeWithTranslation, len(episodes.Episodes))
	var wg sync.WaitGroup
	errChan := make(chan error, len(episodes.Episodes))

	for i, episode := range episodes.Episodes {
		wg.Add(1)
		go func(index int, ep thetvdb_api.EpisodeBaseRecord) {
			defer wg.Done()

			translations := make(map[string]*thetvdb_api.Translation)
			episodeID := strconv.FormatInt(*ep.ID, 10)

			var translationWg sync.WaitGroup
			translationChan := make(chan struct {
				lang        string
				translation *thetvdb_api.Translation
				err         error
			}, len(ep.NameTranslations))

			for _, translation := range ep.NameTranslations {
				translationWg.Add(1)
				go func(lang string) {
					defer translationWg.Done()
					translationRes, err := s.getEpisodeTranslation(ctx, episodeID, lang)
					translationChan <- struct {
						lang        string
						translation *thetvdb_api.Translation
						err         error
					}{lang, translationRes, err}
				}(translation)
			}

			go func() {
				translationWg.Wait()
				close(translationChan)
			}()

			for result := range translationChan {
				if result.err != nil {
					errChan <- result.err
					return
				}
				translations[result.lang] = result.translation
			}

			episodesWithTranslations[index] = AnimeEpisodeWithTranslation{
				EpisodeBaseRecord: ep,
				Translations:      translations,
			}
		}(i, episode)
	}

	wg.Wait()
	close(errChan)

	if err := <-errChan; err != nil {
		return nil, err
	}

	return &episodesWithTranslations, nil
}

func (s *TheTVDBServiceImpl) getEpisodeTranslation(ctx context.Context, episodeID string, lang string) (*thetvdb_api.Translation, error) {
	translation, err := s.api.GetEpisodeTranslation(ctx, episodeID, lang)
	if err != nil {
		return nil, err
	}

	return translation, nil
}

func (s *TheTVDBServiceImpl) GetEpisodesBySeriesID(ctx context.Context, seriesID string) (*[]AnimeEpisodeWithTranslation, error) {
	return s.getAnimeEpisodes(ctx, seriesID)
}
