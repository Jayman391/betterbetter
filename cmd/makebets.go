package cmd

import (
	"betterbetter/src"
	"github.com/spf13/cobra"
)

func init() {

  var RiskReward float64
	var MaxBets int

  betCMD.Flags().Float64VarP(&RiskReward, "rr", "r", 1, "Risk Reward Ratio")
	betCMD.Flags().IntVarP(&MaxBets, "maxbets", "m", 3, "Max number of bets to make")
  rootCmd.AddCommand(betCMD)
}

var betCMD = &cobra.Command{
	Use:   "makebets",
	Short: "Print the version number of betterbetter",
	Long:  `All software has versions. This is betterbetter's`,
	Run: func(cmd *cobra.Command, args []string) {
		rr, err := cmd.Flags().GetFloat64("rr")
		if err != nil {
			panic(err)
		}
		maxbets, err := cmd.Flags().GetInt("maxbets")
		if err != nil {
			panic(err)
		}
		src.MakeBets(rr, maxbets)
	},
}