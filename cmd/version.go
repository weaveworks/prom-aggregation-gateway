package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/zapier/prom-aggregation-gateway/config"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Prints out the version and commit short sha information`,
	Run:   versionFunc,
}

func versionFunc(cmd *cobra.Command, args []string) {
	log.Printf("\n%s\nVersion: %s\nCommitSHA: %s", config.Name, config.Version, config.CommitSHA)
}
