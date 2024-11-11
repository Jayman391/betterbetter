package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"betterbetter/src"
)

func init() {	
	var Year string
	var Sport string

	TeamCMD.Flags().StringVarP(&Sport,"sport", "s", "", "Sport to fetch data for")
	TeamCMD.Flags().StringVarP(&Year,"year", "y", "", "Year to fetch data for")
	
	PlayerCMD.Flags().StringVarP(&Sport,"sport", "s", "", "Sport to fetch data for")
	PlayerCMD.Flags().StringVarP(&Year,"year", "y", "", "Year to fetch data for")


	FetchCmd.AddCommand(TeamCMD)
	FetchCmd.AddCommand(PlayerCMD)
	rootCmd.AddCommand(FetchCmd)
}

var FetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch sports data",
	Long:  `Fetch sports data from the internet`,
	Run: func(cmd *cobra.Command, args []string) {
	},
}


var TeamCMD = &cobra.Command{
	Use:   "team",
	Short: "Fetch data for a specific team",
	Long:  `Fetch data for a specific team`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Fetching data for a specific team")
		fmt.Println(args)
		fmt.Println(cmd.Flag("year").Value)
		
		var Data string 

		if len(args) != 0 {
			params := map[string]string{
				"search": args[0],
			}
	
			Data = src.FetchData(cmd.Flag("sport").Value.String(), "team", params)
		} else {
			params := map[string]string{}

			Data = src.FetchData(cmd.Flag("sport").Value.String(), "team", params)
		}

		if Data != "" {
			parsed_data := src.ParseData(Data)
			fmt.Println(parsed_data["response"])
		} else {
			fmt.Println("No data found")
		}

		var statsParams = map[string]string{}

		if cmd.Flag("year").Value.String() != "" {
			statsParams["season"] = cmd.Flag("year").Value.String()
		}

	},
	
}

var PlayerCMD = &cobra.Command{
	Use:   "player",
	Short: "Fetch data for a specific player",
	Long:  `Fetch data for a specific player`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Fetching data for a specific player")
		fmt.Println(args)
		fmt.Println(cmd.Flag("year").Value)

		var Data string 

		if len(args) != 0 {
			params := map[string]string{
				"search": args[0],
			}
		
			Data = src.FetchData(cmd.Flag("sport").Value.String(), "player", params)
		} else {
			params := map[string]string{}

			Data = src.FetchData(cmd.Flag("sport").Value.String(), "player", params)
		}


		var parsed_data map[string]interface{}

		if Data != "" {
			parsed_data = src.ParseData(Data)
			fmt.Println(parsed_data["response"])
		} else {
			fmt.Println("No data found")
		}

		var statsParams = map[string]string{}

		if cmd.Flag("year").Value.String() != "" {
			statsParams["season"] = cmd.Flag("year").Value.String()
		}
	},
}

