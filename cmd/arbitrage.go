package cmd

import (
	"betterbetter/src"
  "fmt"
	"github.com/spf13/cobra"
)

func init() {
  rootCmd.AddCommand(arbCMD)
}

var arbCMD = &cobra.Command{
  Use:   "arbitrage",
  Short: "Print the version number of betterbetter",
  Long:  `All software has versions. This is betterbetter's`,
  Run: func(cmd *cobra.Command, args []string) {
    results := src.Arbitrage("/Users/user/Desktop/betterbetter/data/nba/2024/celtics", "/Users/user/Desktop/betterbetter/data/nba/2024/2024-11-01")
    fmt.Println(results)
  },
}