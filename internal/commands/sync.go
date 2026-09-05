/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/weeb-vip/scraper-api/config"
	"github.com/weeb-vip/scraper-api/internal/sync"
	"net/http"
)

// syncCmd is the nightly job: derive seasons for anime TheTVDB can place, then
// republish every link so thetvdb-enrichment refreshes episodes and artwork.
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Derive seasons and republish every TheTVDB link",
	Long: `Runs the nightly TheTVDB sync.

Derives a season for each anime that has a thetvdbid and no link yet, then
republishes every link so thetvdb-enrichment refreshes its episodes and
artwork.

Flags:
  --derive-all    also re-derive anime that already have a link. Off by
                  default: the linked ones cost thousands of TheTVDB calls to
                  confirm seasons that have not changed, and a derivation that
                  disagrees with a hand-made link would overwrite it. Use it
                  after TheTVDB revises a series, or after a change to the
                  matcher that should reach anime linked under the old rules.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		deriveAll, _ := cmd.Flags().GetBool("derive-all")

		// start simple http healthcheck
		go func() {
			conf := config.LoadConfigOrPanic()
			http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("ok"))
			})
			// use conf.AppConfig.Port
			http.ListenAndServe(fmt.Sprintf(":%d", conf.AppConfig.Port), nil)
		}()

		return sync.Sync(sync.Options{DeriveAll: deriveAll})
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)

	syncCmd.Flags().Bool("derive-all", false, "re-derive anime that already have a link, not just the unlinked ones")
}
