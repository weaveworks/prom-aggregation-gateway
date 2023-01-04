package cmd

import (
	"github.com/spf13/cobra"
	"github.com/zapier/prom-aggregation-gateway/routers"
)

func init() {
	rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "starts up the server",
	Long:  `Starts up the aggregation server`,
	RunE:  startFunc,
}

func startFunc(cmd *cobra.Command, args []string) error {

	apiCfg := routers.ApiRouterConfig{
		CorsDomain: cfg.CorsDomain,
		Accounts:   cfg.AuthUsers,
	}

	routers.RunServers(apiCfg, cfg.ApiListen, cfg.LifecycleListen)

	return nil
}
