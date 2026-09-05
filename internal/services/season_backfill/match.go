// Package season_backfill works out which season of its TheTVDB series each
// anime is, and records it the same way a human would from the admin panel.
//
// MyAnimeList files every broadcast run as its own anime; TheTVDB keeps them as
// seasons of one series. So thetvdbid is identical across runs -- 1,756 ids are
// shared by 6,425 anime -- and the id that says which series something is
// cannot say which run.
//
// TheTVDB cannot simply be asked. Its series carry remote ids for TheMovieDB
// and IMDB only, and its season records carry none at all, so nothing on that
// side knows a MyAnimeList entry exists. Air dates are the only key the two
// systems share, so a join is unavoidable -- what is avoidable is joining
// loosely, and every rule here exists to refuse rather than guess.
package season_backfill

import (
	"strings"
	"time"

	"github.com/weeb-vip/scraper-api/internal/services/thetvdb_api"
)

const (
	// DateLayout is how TheTVDB writes an air date.
	DateLayout = "2006-01-02"

	// DefaultRequiredRatio is the share of an anime's dated episodes that must
	// land on a season's air days for that season to be accepted.
	//
	// Not 1.0: TheTVDB and MyAnimeList disagree about a delayed broadcast or a
	// recap often enough that demanding every episode would reject correct
	// matches. High enough that a season can only win by being the season.
	DefaultRequiredRatio = 0.8
)

// startDateLayouts covers the shapes anime.start_date arrives in. The column is
// text and has been written by more than one scraper version, so it holds
// RFC3339, postgres' own timestamptz rendering, and bare dates.
var startDateLayouts = []string{
	time.RFC3339,
	"2006-01-02 15:04:05-07",
	"2006-01-02 15:04:05",
	DateLayout,
}

// SeasonWindow is one season's air range as TheTVDB reports it.
type SeasonWindow struct {
	Number int
	First  time.Time
	Last   time.Time
	Count  int
}

// DayKey reduces a timestamp to the calendar day, which is the granularity both
// sides agree on: TheTVDB publishes dates, we store timestamps.
func DayKey(t time.Time) string {
	return t.UTC().Format(DateLayout)
}

// ParseStartDate copes with every shape start_date is stored in.
func ParseStartDate(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}

	for _, layout := range startDateLayouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, true
		}
	}

	if len(value) >= len(DateLayout) {
		if parsed, err := time.Parse(DateLayout, value[:len(DateLayout)]); err == nil {
			return parsed, true
		}
	}

	return time.Time{}, false
}

// SeasonAirDays is the set of calendar days each season aired on, which is what
// the episode match joins against.
func SeasonAirDays(episodes []thetvdb_api.EpisodeBaseRecord) map[int]map[string]bool {
	days := map[int]map[string]bool{}

	for _, episode := range episodes {
		if episode.SeasonNumber == nil || episode.Aired == nil {
			continue
		}
		aired, err := time.Parse(DateLayout, *episode.Aired)
		if err != nil {
			continue
		}
		if days[*episode.SeasonNumber] == nil {
			days[*episode.SeasonNumber] = map[string]bool{}
		}
		days[*episode.SeasonNumber][DayKey(aired)] = true
	}

	return days
}

// SeasonWindows groups a series' episodes into per-season air ranges.
func SeasonWindows(episodes []thetvdb_api.EpisodeBaseRecord) []SeasonWindow {
	type span struct {
		first, last time.Time
		count       int
	}
	byNumber := map[int]*span{}

	for _, episode := range episodes {
		if episode.SeasonNumber == nil {
			continue
		}
		entry, ok := byNumber[*episode.SeasonNumber]
		if !ok {
			entry = &span{}
			byNumber[*episode.SeasonNumber] = entry
		}
		entry.count++

		if episode.Aired == nil {
			continue
		}
		aired, err := time.Parse(DateLayout, *episode.Aired)
		if err != nil {
			continue
		}
		if entry.first.IsZero() || aired.Before(entry.first) {
			entry.first = aired
		}
		if entry.last.IsZero() || aired.After(entry.last) {
			entry.last = aired
		}
	}

	windows := make([]SeasonWindow, 0, len(byNumber))
	for number, entry := range byNumber {
		windows = append(windows, SeasonWindow{
			Number: number, First: entry.first, Last: entry.last, Count: entry.count,
		})
	}

	return windows
}

// MatchByEpisodes picks the season whose episodes aired on the same days as
// this anime's.
//
// This is the rule that makes the result data rather than a guess. It needs
// most of our episodes accounted for, and it needs one season to explain them
// better than any other: a series that was re-cut, or an anime whose episode
// list we only half hold, produces a tie and is left alone.
func MatchByEpisodes(ourAired []time.Time, seasonDays map[int]map[string]bool, requiredRatio float64) (int, bool) {
	if len(ourAired) == 0 {
		return 0, false
	}

	ours := map[string]bool{}
	for _, aired := range ourAired {
		ours[DayKey(aired)] = true
	}

	bestSeason, bestHits, runnerUp := 0, 0, 0
	for season, days := range seasonDays {
		hits := 0
		for day := range ours {
			if days[day] {
				hits++
			}
		}
		if hits > bestHits {
			bestSeason, bestHits, runnerUp = season, hits, bestHits
		} else if hits > runnerUp {
			runnerUp = hits
		}
	}

	if bestHits == 0 {
		return 0, false
	}
	// Ambiguity is a refusal: two seasons explaining our episodes equally well
	// means we cannot tell which one this anime is.
	if bestHits == runnerUp {
		return 0, false
	}
	if float64(bestHits)/float64(len(ours)) < requiredRatio {
		return 0, false
	}

	return bestSeason, true
}

// MatchSeason applies the two rules in order and reports which one decided.
//
// The order matters and so does the gate between them. Episode agreement is
// evidence; a start date landing on a season boundary is a coincidence that
// usually holds and sometimes does not. So the start-date rule runs only for
// anime whose episodes we do not hold -- which is what it was always
// documented to be for, and what this makes true.
//
// The gate is the part worth keeping. When we do hold episodes and
// MatchByEpisodes still declined, it declined for a reason, and the most
// common reason is two seasons explaining our episodes equally well. Falling
// through then lets a rule that compares only the first day of each season
// overturn one that compared every day: it finds exactly one season beginning
// on our start day and answers confidently where the episodes said "cannot
// tell".
//
// Overlord is the case that found this. Its season 0 carries all 13 of season
// 2's air days, because the Ple Ple Pleiades shorts air alongside the episodes
// they accompany. The tie was refused correctly -- and then the start-date rule
// handed the shorts season 2, the season of the show they are extras for. The
// same shape put a Re:ZERO shorts entry in season 3.
func MatchSeason(
	aired []time.Time,
	seasonDays map[int]map[string]bool,
	windows []SeasonWindow,
	start time.Time,
	requiredRatio float64,
) (season int, how string, ok bool) {
	if season, ok := MatchByEpisodes(aired, seasonDays, requiredRatio); ok {
		return season, "episodes", true
	}

	if len(aired) > 0 {
		return 0, "", false
	}

	if season, ok := MatchByExactStart(start, windows); ok {
		return season, "exact-start", true
	}

	return 0, "", false
}

// MatchByExactStart is the fallback for anime whose episodes we do not hold.
// MatchSeason is what enforces that -- calling this for an anime whose episodes
// we do hold lets one day overrule every day.
//
// Exact equality, not proximity: the anime's first day must be the day a season
// began, and exactly one season must begin on it. A "nearest season within N
// days" rule would happily label a spin-off that premiered near a season
// boundary, which is the failure this whole command exists to avoid.
func MatchByExactStart(start time.Time, windows []SeasonWindow) (int, bool) {
	season, found := 0, 0

	for _, window := range windows {
		if !window.First.IsZero() && DayKey(window.First) == DayKey(start) {
			season = window.Number
			found++
		}
	}

	if found != 1 {
		return 0, false
	}

	return season, true
}
