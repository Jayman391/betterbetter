package cmd

import (
	"betterbetter/src"

	"fmt"
	"os"
	"strings"
	"strconv"
	"time"

	"github.com/spf13/cobra"
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

	FetchDataCmd.Flags().StringVarP(&Sport, "sport", "s", "", "Sport to fetch data for")
	FetchDataCmd.Flags().StringSliceVarP(&Players, "players", "p", []string{}, "players to fetch data for")
	FetchDataCmd.Flags().StringSliceVarP(&Teams, "teams", "t", []string{}, "teams to fetch data for")
	FetchDataCmd.Flags().StringVarP(&Year, "year", "y", "", "Year to fetch data for")

	FetchOddsCmd.Flags().StringVarP(&Sport, "sport", "s", "", "Sport to fetch odds for")
	FetchOddsCmd.Flags().StringVarP(&Year, "date", "d", "", "YYYY-MM-DD date to fetch odds for")


	rootCmd.AddCommand(FetchDataCmd)
	rootCmd.AddCommand(FetchOddsCmd)
}

var FetchDataCmd = &cobra.Command{
	Use:   "fetchdata",
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

			
				
				err := src.SaveToFile(parsed_data, fmt.Sprintf("data/%s/%s/%s",sport, year, string(team)), "team_data.json")
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
				err = src.SaveToFile(statsParsed, fmt.Sprintf("data/%s/%s/%s",sport, year, string(team)), "team_stats.json")
				if err != nil {
					fmt.Printf("Error saving team stats: %v\n", err)
					return
				}
				fmt.Println("Team stats saved successfully.")

				Games := src.FetchData(sport, "game", statsParams)
				if Games == "" {
					fmt.Println("No games data found")
					return
				}

				// Save the stats data to a JSON file
				gamesParsed := src.ParseData(Games)
				if gamesParsed == nil {
					fmt.Println("Error parsing games data")
					return
				}
				err = src.SaveToFile(gamesParsed, fmt.Sprintf("data/%s/%s/%s",sport, year, string(team)), "games.json")
				if err != nil {
					fmt.Printf("Error saving games: %v\n", err)
					return
				}
				fmt.Println("Games saved successfully.")
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

							

				err := src.SaveToFile(parsed_data, fmt.Sprintf("data/%s/%s/%s",sport, year, string(player)), "player_data.json")
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
				err = src.SaveToFile(statsParsed, fmt.Sprintf("data/%s/%s/%s",sport, year, string(player)), "player_stats.json")
				if err != nil {
					fmt.Printf("Error saving player stats: %v\n", err)
					return
				}
				fmt.Println("Player stats saved successfully.")
			
		}
	}
},
}

var FetchOddsCmd = &cobra.Command{
	Use:   "fetchodds",
	Short: "Fetch odds data",
	Long:  `Fetch odds data from the internet`,
	Run: func(cmd *cobra.Command, args []string) {

			// Retrieve the date flag value
			date := cmd.Flag("date").Value.String()

			// Split the date by "-" into year, month, day
			dateArr := strings.Split(date, "-")
			if len(dateArr) != 3 {
					fmt.Println("Error: Date must be in the format YYYY-MM-DD")
					return
			}

			// Initialize a slice with length 3 to hold year, month, day as integers
			intDateArr := make([]int, 3)

			// Convert each part of the date to an integer
			for i, v := range dateArr {
					intValue, err := strconv.Atoi(v)
					if err != nil {
							fmt.Printf("Error converting %s to int: %v\n", v, err)
							return
					}
					intDateArr[i] = intValue
			}

			// Create a time.Time object from the integer date components
			dateObj := time.Date(
					intDateArr[0],
					time.Month(intDateArr[1]),
					intDateArr[2],
					0, 0, 0, 0,
					time.UTC,
			)

			// Format the date in RFC3339 format
			formattedDate := dateObj.UTC().Format(time.RFC3339)

			sport := cmd.Flag("sport").Value.String()

			// Fetch and parse odds data
			odds := src.FetchGames(formattedDate, sport)
			parsedOdds := src.ParseData(odds)

			// Retrieve the "data" key from parsed_odds
			gamesInterface, ok := parsedOdds["data"].([]interface{})
			if !ok {
					fmt.Println("Invalid or missing game data")
					return
			}	

			for _, game := range gamesInterface {
				gameMap, ok := game.(map[string]interface{})
				if !ok {
					fmt.Println("Invalid game data")
					return
				}

				odds := src.FetchOdds(formattedDate, sport, gameMap["id"].(string))
				parsedOdds := src.ParseData(odds)
				err := src.SaveToFile(parsedOdds, fmt.Sprintf("data/%s/%s/%s/%s", sport, dateArr[0], date, gameMap["away_team"].(string)+"_"+gameMap["home_team"].(string)), "odds.json")
				if err != nil {
					fmt.Printf("Error saving odds data: %v\n", err)
					return
				}
			}
	},
}
