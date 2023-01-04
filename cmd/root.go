package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/zapier/prom-aggregation-gateway/config"
)

var cfg = config.Server{}

var rootCmd = &cobra.Command{
	Use:   "prom-aggregation-gateway",
	Short: "prometheus aggregation gateway",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return config.Initialize(cmd)
	},
	// have the start func as the default entry point to keep the API the same
	RunE: startFunc,
}

func Execute() {
	rootCmd.SilenceUsage = true

	rootCmd.PersistentFlags().StringSliceVar(&cfg.AuthUsers, "AuthUsers", []string{}, "List of allowed auth users and their passwords comma separated\n Example: \"user1=pass1,user2=pass2\"")
	rootCmd.PersistentFlags().StringVar(&cfg.ApiListen, "apiListen", ":80", "Listen for API requests on this host/port.")
	rootCmd.PersistentFlags().StringVar(&cfg.LifecycleListen, "lifecycleListen", ":8888", "Listen for lifecycle requests (health, metrics) on this host/port")
	rootCmd.PersistentFlags().StringVar(&cfg.CorsDomain, "cors", "*", "The 'Access-Control-Allow-Origin' value to be returned.")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
