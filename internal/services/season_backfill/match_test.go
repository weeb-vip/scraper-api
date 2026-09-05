package season_backfill

import (
	"testing"
	"time"

	"github.com/weeb-vip/scraper-api/internal/services/thetvdb_api"
)

func ep(season int, aired string) thetvdb_api.EpisodeBaseRecord {
	s, a := season, aired

	return thetvdb_api.EpisodeBaseRecord{SeasonNumber: &s, Aired: &a}
}

func days(t *testing.T, values ...string) []time.Time {
	t.Helper()
	out := make([]time.Time, 0, len(values))
	for _, v := range values {
		parsed, err := time.Parse(DateLayout, v)
		if err != nil {
			t.Fatalf("bad test date %q: %v", v, err)
		}
		out = append(out, parsed)
	}

	return out
}

// TheTVDB series 79414 as the API actually returns it, which is the case this
// exists for: MyAnimeList splits it into two anime (849 and 4382), TheTVDB
// keeps one series with two seasons, and only the air dates connect them.
func haruhi() []thetvdb_api.EpisodeBaseRecord {
	return []thetvdb_api.EpisodeBaseRecord{
		ep(0, "2010-02-06"),
		ep(1, "2006-04-03"), ep(1, "2006-04-10"), ep(1, "2006-04-17"), ep(1, "2006-07-03"),
		ep(2, "2009-05-22"), ep(2, "2009-06-19"), ep(2, "2009-06-26"), ep(2, "2009-07-03"),
		ep(2, "2009-07-10"), ep(2, "2009-07-17"), ep(2, "2009-07-24"), ep(2, "2009-07-31"),
		ep(2, "2009-08-07"), ep(2, "2009-09-11"),
	}
}

func TestSeasonAirDaysBucketsBySeason(t *testing.T) {
	got := SeasonAirDays(haruhi())

	if len(got) != 3 {
		t.Fatalf("want 3 seasons, got %d", len(got))
	}
	if !got[1]["2006-04-03"] || !got[2]["2009-05-22"] || !got[0]["2010-02-06"] {
		t.Error("an episode landed in the wrong season")
	}
	if got[1]["2009-05-22"] {
		t.Error("a 2009 day appeared under season 1")
	}
}

func TestMatchByEpisodesJoinsOnAirDays(t *testing.T) {
	airDays := SeasonAirDays(haruhi())

	for _, tc := range []struct {
		name string
		ours []time.Time
		want int
	}{
		{
			name: "the 2006 anime is season 1",
			ours: days(t, "2006-04-03", "2006-04-10", "2006-04-17", "2006-07-03"),
			want: 1,
		},
		{
			name: "the 2009 anime is season 2",
			ours: days(t, "2009-05-22", "2009-06-19", "2009-06-26", "2009-07-03",
				"2009-07-10", "2009-07-17", "2009-07-24", "2009-07-31", "2009-08-07", "2009-09-11"),
			want: 2,
		},
		{
			// TheTVDB and MyAnimeList disagree about a delayed broadcast often
			// enough that demanding every episode would reject correct matches.
			name: "nine of ten still matches",
			ours: days(t, "2009-05-22", "2009-06-19", "2009-06-26", "2009-07-03",
				"2009-07-10", "2009-07-17", "2009-07-24", "2009-07-31", "2009-08-07", "2011-01-01"),
			want: 2,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := MatchByEpisodes(tc.ours, airDays, DefaultRequiredRatio)
			if !ok {
				t.Fatal("no match")
			}
			if got != tc.want {
				t.Errorf("season = %d, want %d", got, tc.want)
			}
		})
	}
}

// The refusals. Each would previously have been answered with a nearest-season
// guess, and each would have written a wrong season.
func TestMatchByEpisodesRefusesRatherThanGuess(t *testing.T) {
	airDays := SeasonAirDays(haruhi())

	for _, tc := range []struct {
		name string
		ours []time.Time
	}{
		{
			name: "a spin-off sharing the series id but airing on its own days",
			ours: days(t, "2015-10-07", "2015-10-14", "2015-10-21"),
		},
		{
			name: "too few of our episodes land on the season's days",
			ours: days(t, "2009-05-22", "2009-06-19", "2011-01-08", "2011-01-15", "2011-01-22"),
		},
		{
			name: "an anime we hold no dated episodes for",
			ours: nil,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if season, ok := MatchByEpisodes(tc.ours, airDays, DefaultRequiredRatio); ok {
				t.Errorf("matched season %d; want a refusal", season)
			}
		})
	}
}

// A tie is a refusal: two seasons explaining our episodes equally well means we
// cannot say which one this anime is.
func TestMatchByEpisodesRefusesATie(t *testing.T) {
	airDays := SeasonAirDays([]thetvdb_api.EpisodeBaseRecord{
		ep(1, "2020-01-05"), ep(2, "2020-01-05"),
	})

	if season, ok := MatchByEpisodes(days(t, "2020-01-05"), airDays, 1.0); ok {
		t.Errorf("matched season %d on a tie; want a refusal", season)
	}
}

func TestMatchByExactStartRequiresExactness(t *testing.T) {
	windows := SeasonWindows(haruhi())

	if got, ok := MatchByExactStart(days(t, "2009-05-22")[0], windows); !ok || got != 2 {
		t.Errorf("exact start: got season %d ok=%v, want season 2", got, ok)
	}

	// One day off is not a match. An earlier version allowed 45 days of slack,
	// which is how a spin-off premiering near a season boundary was mislabelled.
	if season, ok := MatchByExactStart(days(t, "2009-05-23")[0], windows); ok {
		t.Errorf("matched season %d one day off; want a refusal", season)
	}
}

// TheTVDB series 294002 in the shape that matters: the Ple Ple Pleiades shorts
// air alongside the episodes they accompany, so season 0 carries every one of
// season 2's air days.
func overlord() []thetvdb_api.EpisodeBaseRecord {
	return []thetvdb_api.EpisodeBaseRecord{
		ep(2, "2018-01-09"), ep(2, "2018-01-16"), ep(2, "2018-01-23"),
		ep(0, "2018-01-09"), ep(0, "2018-01-16"), ep(0, "2018-01-23"),
		// Season 0 also holds extras from years the show was off air, which is
		// why its first day is nowhere near season 2's.
		ep(0, "2015-07-07"),
	}
}

func TestMatchSeasonPrefersEpisodes(t *testing.T) {
	season, how, ok := MatchSeason(days(t, "2009-05-22", "2009-06-19", "2009-06-26"),
		SeasonAirDays(haruhi()), SeasonWindows(haruhi()), days(t, "2009-05-22")[0], DefaultRequiredRatio)

	if !ok || season != 2 || how != "episodes" {
		t.Errorf("got season %d via %q ok=%v, want season 2 via episodes", season, how, ok)
	}
}

func TestMatchSeasonFallsBackOnlyWithoutEpisodes(t *testing.T) {
	airDays, windows := SeasonAirDays(haruhi()), SeasonWindows(haruhi())

	// No episodes of our own: the start date is all there is, and it is allowed
	// to decide.
	season, how, ok := MatchSeason(nil, airDays, windows, days(t, "2009-05-22")[0], DefaultRequiredRatio)
	if !ok || season != 2 || how != "exact-start" {
		t.Errorf("got season %d via %q ok=%v, want season 2 via exact-start", season, how, ok)
	}
}

// The regression this whole change exists for. An anime whose episodes tie
// across two seasons must be refused, not handed to the start-date rule --
// which would answer confidently, because only one season begins on that day.
func TestMatchSeasonRefusesWhenEpisodesAreAmbiguous(t *testing.T) {
	airDays, windows := SeasonAirDays(overlord()), SeasonWindows(overlord())
	ours := days(t, "2018-01-09", "2018-01-16", "2018-01-23")
	start := ours[0]

	// The old order: the tie is refused, and then exact-start supplies the very
	// answer the tie said we could not have.
	if _, ok := MatchByEpisodes(ours, airDays, DefaultRequiredRatio); ok {
		t.Fatal("episodes matched a tie; the rest of this test assumes a refusal")
	}
	if season, ok := MatchByExactStart(start, windows); !ok || season != 2 {
		t.Fatalf("exact-start got season %d ok=%v; the regression needs it to answer 2", season, ok)
	}

	// The gate: holding episodes at all means the start date does not get to
	// speak.
	if season, how, ok := MatchSeason(ours, airDays, windows, start, DefaultRequiredRatio); ok {
		t.Errorf("matched season %d via %q; want a refusal", season, how)
	}
}

func TestParseStartDateHandlesEveryShapeTheColumnHolds(t *testing.T) {
	for _, in := range []string{"2006-04-03", "2006-04-03T04:00:00Z", "2006-04-03 04:00:00+00"} {
		got, ok := ParseStartDate(in)
		if !ok {
			t.Errorf("%q did not parse", in)
			continue
		}
		if got.Format(DateLayout) != "2006-04-03" {
			t.Errorf("%q parsed to %s", in, got.Format(DateLayout))
		}
	}

	if _, ok := ParseStartDate(""); ok {
		t.Error("empty string parsed")
	}
}
