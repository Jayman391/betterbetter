package cmd

import (
	"betterbetter/src"
	"github.com/spf13/cobra"
)

func init() {

  var StatsPath string
  var OddsPath string

  arbCMD.Flags().StringVarP(&StatsPath, "stats", "s", "", "Path to stats data")
  arbCMD.Flags().StringVarP(&OddsPath, "odds", "o", "", "Path to odds data")

  rootCmd.AddCommand(arbCMD)
}

var arbCMD = &cobra.Command{
  Use:   "arbitrage",
  Short: "Print the version number of betterbetter",
  Long:  `All software has versions. This is betterbetter's`,
  Run: func(cmd *cobra.Command, args []string) {
    src.Arbitrage(cmd.Flag("stats").Value.String(), cmd.Flag("odds").Value.String())
  },
}