package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"betterbetter/src"

	"os"
)

func init() {
	var Year string
	var Sport string

	TeamCMD.Flags().StringVarP(&Sport, "sport", "s", "", "Sport to fetch data for")
	TeamCMD.Flags().StringVarP(&Year, "year", "y", "", "Year to fetch data for")

	PlayerCMD.Flags().StringVarP(&Sport, "sport", "s", "", "Sport to fetch data for")
	PlayerCMD.Flags().StringVarP(&Year, "year", "y", "", "Year to fetch data for")

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

		var Data string

		// Fetch team data
		params := map[string]string{}
		if len(args) != 0 {
			params["search"] = args[0]
		}

		Data = src.FetchData(cmd.Flag("sport").Value.String(), "team", params)

		if Data == "" {
			fmt.Println("No data found")
			return
		}

		// Parse the fetched data
		parsed_data := src.ParseData(Data)
		if parsed_data == nil {
			fmt.Println("Error parsing data")
			return
		}

		// Save the team data to a JSON file
		year := cmd.Flag("year").Value.String()

		direrr := os.MkdirAll("data", os.ModePerm)
		if direrr != nil {
			fmt.Printf("Error creating data directory: %v\n", direrr)
			return
		}

		direrr = os.MkdirAll(fmt.Sprintf("data/%s", year), os.ModePerm)
		if direrr != nil {
			fmt.Printf("Error creating year directory: %v\n", direrr)
			return
		}

		direrr = os.MkdirAll(fmt.Sprintf("data/%s/%s", year, args[0]), os.ModePerm)
		if direrr != nil {
			fmt.Printf("Error creating player directory: %v\n", direrr)
			return
		}
		
		err := src.SaveToFile(parsed_data, fmt.Sprintf("data/%s/%s", year, args[0]), "team_data.json")
		if err != nil {
			fmt.Printf("Error saving team data: %v\n", err)
			return
		}
		fmt.Println("Team data saved successfully.")

		// Prepare and fetch team stats
		statsParams := map[string]string{}
		if year != "" {
			statsParams["season"] = year
		}

		// Extract team ID
		response, ok := parsed_data["response"].([]interface{})
		if !ok || len(response) == 0 {
			fmt.Println("Invalid or missing response data")
			return
		}
		teamData := response[0].(map[string]interface{})
		idValue, exists := teamData["id"]
		if !exists {
			fmt.Println("Team ID key not found in teamData")
			return
		}
		id, ok := idValue.(float64)
		if !ok {
			fmt.Printf("Team ID is not a float64, actual type: %T, value: %v\n", idValue, idValue)
			return
		}

		statsParams["team"] = fmt.Sprintf("%.0f", id)
		Stats := src.FetchData(cmd.Flag("sport").Value.String(), "team-stats", statsParams)
		if Stats == "" {
			fmt.Println("No stats data found")
			return
		}

		// Save the stats data to a JSON file
		statsParsed := src.ParseData(Stats)
		if statsParsed == nil {
			fmt.Println("Error parsing stats data")
			return
		}
		err = src.SaveToFile(statsParsed, fmt.Sprintf("data/%s/%s", year, args[0]), "team_stats.json")
		if err != nil {
			fmt.Printf("Error saving team stats: %v\n", err)
			return
		}
		fmt.Println("Team stats saved successfully.")
	},
}


var PlayerCMD = &cobra.Command{
	Use:   "player",
	Short: "Fetch data for a specific player",
	Long:  `Fetch data for a specific player`,
	Run: func(cmd *cobra.Command, args []string) {
		var Data string

		// Fetch player data
		params := map[string]string{}
		if len(args) != 0 {
			params["search"] = args[0]
		}

		Data = src.FetchData(cmd.Flag("sport").Value.String(), "player", params)

		if Data == "" {
			fmt.Println("No data found")
			return
		}

		// Parse the fetched data
		parsed_data := src.ParseData(Data)
		if parsed_data == nil {
			fmt.Println("Error parsing data")
			return
		}

		// Save the player data to a JSON file
		year := cmd.Flag("year").Value.String()

		direrr := os.MkdirAll("data", os.ModePerm)
		if direrr != nil {
			fmt.Printf("Error creating data directory: %v\n", direrr)
			return
		}

		direrr = os.MkdirAll(fmt.Sprintf("data/%s", year), os.ModePerm)
		if direrr != nil {
			fmt.Printf("Error creating year directory: %v\n", direrr)
			return
		}

		direrr = os.MkdirAll(fmt.Sprintf("data/%s/%s", year, args[0]), os.ModePerm)
		if direrr != nil {
			fmt.Printf("Error creating player directory: %v\n", direrr)
			return
		}

		err := src.SaveToFile(parsed_data, fmt.Sprintf("data/%s/%s", year, args[0]), "player_data.json")
		if err != nil {
			fmt.Printf("Error saving player data: %v\n", err)
			return
		}
		fmt.Println("Player data saved successfully.")

		// Prepare and fetch player stats
		statsParams := map[string]string{}
		if year != "" {
			statsParams["season"] = year
		}

		// Extract player ID
		response, ok := parsed_data["response"].([]interface{})
		if !ok || len(response) == 0 {
			fmt.Println("Invalid or missing response data")
			return
		}
		playerData := response[0].(map[string]interface{})
		idValue, exists := playerData["id"]
		if !exists {
			fmt.Println("Player ID key not found in playerData")
			return
		}
		id, ok := idValue.(float64)
		if !ok {
			fmt.Printf("Player ID is not a float64, actual type: %T, value: %v\n", idValue, idValue)
			return
		}

		statsParams["id"] = fmt.Sprintf("%.0f", id)
		Stats := src.FetchData(cmd.Flag("sport").Value.String(), "player-stats", statsParams)
		if Stats == "" {
			fmt.Println("No stats data found")
			return
		}

		// Save the stats data to a JSON file
		statsParsed := src.ParseData(Stats)
		if statsParsed == nil {
			fmt.Println("Error parsing stats data")
			return
		}
		err = src.SaveToFile(statsParsed, fmt.Sprintf("data/%s/%s", year, args[0]), "player_stats.json")
		if err != nil {
			fmt.Printf("Error saving player stats: %v\n", err)
			return
		}
		fmt.Println("Player stats saved successfully.")
	},
}




