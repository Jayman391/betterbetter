package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"betterbetter/src"

	"os"

	"strings"
)

type mapFunc[E any] func(E) E

func Map[S ~[]E, E any](s S, f mapFunc[E]) S {
	result := make(S, len(s))
	for i := range s {
			result[i] = f(s[i])
	}
	return result
}

func TrimBracket(s string) string {
	return strings.Trim(s, "[]")
}

func init() {
	var Year string
	var Sport string
	var Teams []string
	var Players []string

	FetchCmd.Flags().StringVarP(&Sport, "sport", "s", "", "Sport to fetch data for")
	FetchCmd.Flags().StringSliceVarP(&Players, "players", "p", []string{}, "players to fetch data for")
	FetchCmd.Flags().StringSliceVarP(&Teams, "teams", "t", []string{}, "teams to fetch data for")
	FetchCmd.Flags().StringVarP(&Year, "year", "y", "", "Year to fetch data for")

	rootCmd.AddCommand(FetchCmd)
}

var FetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch sports data",
	Long:  `Fetch sports data from the internet`,
	Run: func(cmd *cobra.Command, args []string) {

		sport := cmd.Flag("sport").Value.String()
		year := cmd.Flag("year").Value.String()

		teams := strings.Split(cmd.Flag("teams").Value.String(), ",")
		teams = Map(teams, TrimBracket)

		fmt.Println(teams)


		teampaths := []string{}
		if teams[0] != "" {
			for _, team := range teams {
				teampaths = append(teampaths, fmt.Sprintf("data/%s/%s/%s", sport, year, team))
			}
	  }

		players := strings.Split(cmd.Flag("players").Value.String(), ",")
		players = Map(players, TrimBracket)

		playerpaths := []string{}
		if players[0] != "" {
			for _, player := range players {
				playerpaths = append(playerpaths, fmt.Sprintf("data/%s/%s/%s", sport, year, player))
			}
		}

		os.Mkdir(fmt.Sprintf("data/%s", sport), os.ModePerm)
		os.Mkdir(fmt.Sprintf("data/%s/%s", sport, year), os.ModePerm)

		if len(teampaths) > 0 {
			for _, path := range teampaths {
				os.Mkdir(path, os.ModePerm)
			}
		}

		if len(playerpaths) > 0 {
			for _, path := range playerpaths {
				os.Mkdir(path, os.ModePerm)
			}
		}

		if teams[0] != "" {
			for _, team := range teams {

				team = strings.Trim(team, "[]")

				params := map[string]string{}
				
				params["search"] = string(team)

				Data := src.FetchData(sport, "team", params)

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

				direrr = os.MkdirAll(fmt.Sprintf("data/%s/%s", year, string(team)), os.ModePerm)
				if direrr != nil {
					fmt.Printf("Error creating player directory: %v\n", direrr)
					return
				}
				
				err := src.SaveToFile(parsed_data, fmt.Sprintf("data/%s/%s", year, string(team)), "team_data.json")
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
				Stats := src.FetchData(sport, "team-stats", statsParams)
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
				err = src.SaveToFile(statsParsed, fmt.Sprintf("data/%s/%s", year, string(team)), "team_stats.json")
				if err != nil {
					fmt.Printf("Error saving team stats: %v\n", err)
					return
				}
				fmt.Println("Team stats saved successfully.")
			}
		}

		if players[0] != "" {

			for _, player := range players {

				player = strings.Trim(player, "[]")

				params := map[string]string{}

				params["search"] = string(player)
			

				Data := src.FetchData(sport, "player", params)

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

				direrr = os.MkdirAll(fmt.Sprintf("data/%s/%s", year, string(player)), os.ModePerm)
				if direrr != nil {
					fmt.Printf("Error creating player directory: %v\n", direrr)
					return
				}

				err := src.SaveToFile(parsed_data, fmt.Sprintf("data/%s/%s", year, string(player)), "player_data.json")
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
				Stats := src.FetchData(sport, "player-stats", statsParams)
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
				err = src.SaveToFile(statsParsed, fmt.Sprintf("data/%s/%s", year, string(player)), "player_stats.json")
				if err != nil {
					fmt.Printf("Error saving player stats: %v\n", err)
					return
				}
				fmt.Println("Player stats saved successfully.")
			
		}
	}
},
}