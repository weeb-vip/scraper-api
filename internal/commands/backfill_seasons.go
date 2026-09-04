package commands

import (
	"context"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/weeb-vip/scraper-api/config"
	"github.com/weeb-vip/scraper-api/internal/db"
	anime2 "github.com/weeb-vip/scraper-api/internal/db/repositories/anime"
	"github.com/weeb-vip/scraper-api/internal/db/repositories/thetvdblink"
	"github.com/weeb-vip/scraper-api/internal/services/link_service"
	"github.com/weeb-vip/scraper-api/internal/services/season_backfill"
	"github.com/weeb-vip/scraper-api/internal/services/thetvdb_api"
)

// backfillSeasonsCmd automates what a person does in the admin panel: deciding
// which season of a TheTVDB series an anime is, and saving that as a link.
//
// It belongs in this service because this is where that decision is made and
// where thetvdb_link lives. thetvdb-enrichment consumes the decision -- it is
// handed a season in its event payload and never looks one up -- so it is the
// wrong side of the pipeline to derive one on.
var backfillSeasonsCmd = &cobra.Command{
	Use:   "backfill-seasons",
	Short: "Derive each anime's series season from TheTVDB air dates",
	Long: `Walks anime carrying a thetvdbid, fetches each series' episodes once, and
matches the days our episodes aired against the days each season aired.

TheTVDB cannot be asked directly: its series carry remote ids for TheMovieDB and
IMDB only, and its seasons carry none, so nothing there knows a MyAnimeList
entry exists. Air dates are the only key the two systems share.

A season is accepted only when it explains most of an anime's episodes and
explains them better than any other season. Ties, weak matches and anime with no
usable dates are left alone, so season_number stays null: an unknown season
renders as nothing, a wrong one renders as a lie.

Anime whose episodes we do not hold fall back to exact equality between the
anime's start date and a season's first air day -- exact, not nearest.

Writes links, which the database trigger mirrors onto anime.season_number.
Publishing the enrichment event is deliberately left to the admin panel's own
Sync: a run over the whole catalogue would otherwise queue thousands of episode
and artwork fetches in one burst.

Flags:
  --dry-run          match and log, write nothing
  --limit N          stop after N anime (0 = all)
  --delay-ms N       pause after each series fetch, to stay under rate limits
  --after ID         resume from an anime id (exclusive)
  --required-ratio F share of an anime's dated episodes that must land on a
                     season's air days for it to be accepted (default 0.8)
  --only-unlinked    skip anime that already have a link. What the nightly sync
                     uses; leave it off for a full re-derive`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		limit, _ := cmd.Flags().GetInt("limit")
		delayMs, _ := cmd.Flags().GetInt("delay-ms")
		after, _ := cmd.Flags().GetString("after")
		ratio, _ := cmd.Flags().GetFloat64("required-ratio")
		onlyUnlinked, _ := cmd.Flags().GetBool("only-unlinked")

		conf := config.LoadConfigOrPanic()
		database := db.NewDatabase(conf.DBConfig)
		animeRepository := anime2.NewAnimeRepository(database)

		// A producer that is never called: LinkService uses it only in Sync,
		// and this command deliberately does not publish. Passing nil would
		// leave a live nil call one refactor away.
		noopProducer := func(ctx context.Context, value []byte) error { return nil }

		runner := &season_backfill.Runner{
			DB:    database.DB,
			Anime: animeRepository,
			Links: link_service.NewLinkService(
				thetvdblink.NewTheTVDBLinkRepository(database),
				animeRepository,
				noopProducer,
			),
			API: thetvdb_api.NewTheTVDBApi(conf.TheTVDBConfig, &http.Client{}),
		}

		result, err := runner.Run(cmd.Context(), season_backfill.Options{
			DryRun:        dryRun,
			Limit:         limit,
			DelayMs:       delayMs,
			After:         after,
			RequiredRatio: ratio,
			OnlyUnlinked:  onlyUnlinked,
		}, func(format string, args ...interface{}) {
			fmt.Printf(format+"\n", args...)
		})
		if err != nil {
			return err
		}

		fmt.Printf("processed=%d linked=%d skipped=%d failed=%d seriesFetched=%d dryRun=%v\n",
			result.Processed, result.Linked, result.Skipped, result.Failed, result.SeriesFetched, dryRun)

		return nil
	},
}

func init() {
	backfillSeasonsCmd.Flags().Bool("dry-run", false, "match and log, write nothing")
	backfillSeasonsCmd.Flags().Int("limit", 0, "stop after N anime (0 = all)")
	backfillSeasonsCmd.Flags().Int("delay-ms", 200, "pause after each series fetch")
	backfillSeasonsCmd.Flags().String("after", "", "resume from an anime id (exclusive)")
	backfillSeasonsCmd.Flags().Bool("only-unlinked", false, "skip anime that already have a link")
	backfillSeasonsCmd.Flags().Float64("required-ratio", season_backfill.DefaultRequiredRatio,
		"share of an anime's episodes that must land on a season's air days")
	rootCmd.AddCommand(backfillSeasonsCmd)
}
