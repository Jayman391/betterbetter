package src

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)


type ArbitrageResult struct {
	ExpectedValue float64
	BookProb float64
	ModelProb float64
	Differential float64
	Bet map[string]any
}


func Arbitrage(statspath string, oddspath string) []ArbitrageResult {

	results := make([]ArbitrageResult, 0)

	oddsMap := make(map[string][]map[string]any)
	// Read odds from JSON files
	oddsData := ReadOdds(oddspath)
	// extract odds and store in oddsMap
	for _, data := range oddsData {
		// convert data to map[string]any
		data := data.(map[string]interface{})
		gameData := data["data"]
		// convert gameData to map[string]any
		gameDataMap := gameData.(map[string]interface{})
		time := gameDataMap["commence_time"]
		homeTeam := gameDataMap["home_team"]
		awayTeam := gameDataMap["away_team"]

		bookmakers := gameDataMap["bookmakers"]
		// convert bookmakers to map[string]any
		bookmakersMap := bookmakers.([]interface{})
		for _, bookmaker := range bookmakersMap {
			markets := bookmaker.(map[string]interface{})["markets"]
			// convert markets to map[string]any
			marketsMap := markets.([]interface{})
			for _, market := range marketsMap {
				marketMap := market.(map[string]interface{})
				key := marketMap["key"].(string)
				outcomes := marketMap["outcomes"].([]interface{})
				
				for _, outcome := range outcomes {
					outcomeMap := outcome.(map[string]interface{})
					outcomeMap["time"] = time	
					outcomeMap["home_team"] = homeTeam
					outcomeMap["away_team"] = awayTeam
					oddsMap[key] = append(oddsMap[key], outcomeMap)
				}
			}
		}
	}

	//h2h := oddsMap["h2h"]
	//spreads := oddsMap["spreads"]
	//totals := oddsMap["totals"]
	playerPoints := oddsMap["player_points"]
	playerRebounds := oddsMap["player_rebounds"]
	playerAssists := oddsMap["player_assists"]
	playerBlocks := oddsMap["player_blocks"]
	playerSteals := oddsMap["player_steals"]
	playerTurnUnders := oddsMap["player_turnUnders"]
	//playerDoubleDouble := oddsMap["player_double_double"]
	//playerTripleDouble := oddsMap["player_triple_double"]
	playerPointsRebounds := oddsMap["player_points_rebounds"]
	playerPointsAssists := oddsMap["player_points_assists"]
	playerReboundsAssists := oddsMap["player_rebounds_assists"]
	playerPointsReboundsAssists := oddsMap["player_points_rebounds_assists"]

	// Read stats from CSV files
	stats := ReadPreds(statspath)

	for player, playerStats := range stats {
		// get player stats
		points := playerStats["points"]
		rebounds := playerStats["totReb"]
		assists := playerStats["assists"]
		blocks := playerStats["blocks"]
		steals := playerStats["steals"]
		turnUnders := playerStats["turnUnders"]

		// for all player odds, get subset for which descripton=player

		pointsOdds := SearchPlayerOdds(playerPoints, player)
		reboundsOdds := SearchPlayerOdds(playerRebounds, player)
		assistsOdds := SearchPlayerOdds(playerAssists, player)
		blocksOdds := SearchPlayerOdds(playerBlocks, player)
		stealsOdds := SearchPlayerOdds(playerSteals, player)
		turnUndersOdds := SearchPlayerOdds(playerTurnUnders, player)
	
		pointsReboundsOdds := SearchPlayerOdds(playerPointsRebounds, player)
		pointsAssistsOdds := SearchPlayerOdds(playerPointsAssists, player)
		reboundsAssistsOdds := SearchPlayerOdds(playerReboundsAssists, player)
		pointsReboundsAssistsOdds := SearchPlayerOdds(playerPointsReboundsAssists, player)

		//points odds
			//cdf
		for _, pointBet := range pointsOdds {
			value := int64(math.Ceil(pointBet["point"].(float64))) // Perform type assertion to convert to int64
			cdf := CDF(points, value)

			if pointBet["name"] == "Under" {
				cdf = 1 - cdf
			}

			odds := 1.0/pointBet["price"].(float64)

			differential := cdf - odds

			pointBet["type"] = "points"

			results = append(results, ArbitrageResult{
				ExpectedValue: pointBet["price"].(float64),
				BookProb: odds,
				ModelProb: cdf,
				Differential: differential,
				Bet: pointBet,
			})
		}


		//rebound odds
			//cdf

		for _, reboundBet := range reboundsOdds {
			value := int64(math.Ceil(reboundBet["point"].(float64))) // Perform type assertion to convert to int64
			cdf := CDF(rebounds, value)

			if reboundBet["name"] == "Under" {
				cdf = 1 - cdf
			}

			odds := 1.0/reboundBet["price"].(float64)

			differential := cdf - odds

			reboundBet["type"] = "rebounds"

			results = append(results, ArbitrageResult{
				ExpectedValue: reboundBet["price"].(float64),
				BookProb: odds,
				ModelProb: cdf,
				Differential: differential,
				Bet: reboundBet,
			})

		}

		//assist odds
			//cdf

		for _, assistBet := range assistsOdds {
			value := int64(math.Ceil(assistBet["point"].(float64))) // Perform type assertion to convert to int64
			cdf := CDF(assists, value)

			if assistBet["name"] == "Under" {
				cdf = 1 - cdf
			}

			odds := 1.0/assistBet["price"].(float64)

			differential := cdf - odds

			assistBet["type"] = "assists"

			results = append(results, ArbitrageResult{
				ExpectedValue: assistBet["price"].(float64),
				BookProb: odds,
				ModelProb: cdf,
				Differential: differential,
				Bet: assistBet,
			})

		}

		//blocks odds
			//cdf	

		for _, blockBet := range blocksOdds {
			value := int64(math.Ceil(blockBet["point"].(float64))) // Perform type assertion to convert to int64
			cdf := CDF(blocks, value)

			if blockBet["name"] == "Under" {
				cdf = 1 - cdf
			}

			odds := 1.0/blockBet["price"].(float64)

			differential := cdf - odds

			blockBet["type"] = "blocks"

			results = append(results, ArbitrageResult{
				ExpectedValue: blockBet["price"].(float64),
				BookProb: odds,
				ModelProb: cdf,
				Differential: differential,
				Bet: blockBet,
			})

		}

		//steals odds
			//cdf

		for _, stealBet := range stealsOdds {
			value := int64(math.Ceil(stealBet["point"].(float64))) // Perform type assertion to convert to int64
			cdf := CDF(steals, value)

			if stealBet["name"] == "Under" {
				cdf = 1 - cdf
			}

			odds := 1.0/stealBet["price"].(float64)

			differential := cdf - odds

			stealBet["type"] = "steals"

			results = append(results, ArbitrageResult{
				ExpectedValue: stealBet["price"].(float64),
				BookProb: odds,
				ModelProb: cdf,
				Differential: differential,
				Bet: stealBet,
			})

		}

		//turnUnders odds
			//cdf

		for _, turnUnderBet := range turnUndersOdds {
			value := int64(math.Ceil(turnUnderBet["point"].(float64))) // Perform type assertion to convert to int64
			cdf := CDF(turnUnders, value)

			if turnUnderBet["name"] == "Under" {
				cdf = 1 - cdf
			}

			odds := 1.0/turnUnderBet["price"].(float64)

			differential := cdf - odds

			turnUnderBet["type"] = "turnUnders"

			results = append(results, ArbitrageResult{
				ExpectedValue: turnUnderBet["price"].(float64),
				BookProb: odds,
				ModelProb: cdf,
				Differential: differential,
				Bet: turnUnderBet,
			})

		}

		//triple double odds
			//cartesian3
				//cdf3
			
			

		//points rebounds odds
			//CombinationSum
				//cdf

		for _, pointReboundBet := range pointsReboundsOdds {
			value := int64(math.Ceil(pointReboundBet["point"].(float64))) // Perform type assertion to convert to int64
			combo := CombinationSum(points, rebounds)
			cdf := CDF(combo, value)

			if pointReboundBet["name"] == "Under" {
				cdf = 1 - cdf
			}

			odds := 1.0/pointReboundBet["price"].(float64)

			differential := cdf - odds

			pointReboundBet["type"] = "points_rebounds"

			results = append(results, ArbitrageResult{
				ExpectedValue: pointReboundBet["price"].(float64),
				BookProb: odds,
				ModelProb: cdf,
				Differential: differential,
				Bet: pointReboundBet,
			})
		}

		//points assists odds
			//CombinationSum
				//cdf
		
		for _, pointAssistBet := range pointsAssistsOdds {
			value := int64(math.Ceil(pointAssistBet["point"].(float64))) // Perform type assertion to convert to int64

			combo := CombinationSum(points, assists)

			cdf := CDF(combo, value)

			if pointAssistBet["name"] == "Under" {
				cdf = 1 - cdf
			}

			odds := 1.0/pointAssistBet["price"].(float64)

			differential := cdf - odds

			pointAssistBet["type"] = "points_assists"

			results = append(results, ArbitrageResult{
				ExpectedValue: pointAssistBet["price"].(float64),
				BookProb: odds,
				ModelProb: cdf,
				Differential: differential,
				Bet: pointAssistBet,
			})
		}

		//rebounds assists odds
			//CombinationSum
				//cdf

		for _, reboundAssistBet := range reboundsAssistsOdds {
			value := int64(math.Ceil(reboundAssistBet["point"].(float64))) // Perform type assertion to convert to int64

			combo := CombinationSum(rebounds, assists)

			cdf := CDF(combo, value)

			if reboundAssistBet["name"] == "Under" {
				cdf = 1 - cdf
			}

			odds := 1.0/reboundAssistBet["price"].(float64)

			differential := cdf - odds

			reboundAssistBet["type"] = "rebounds_assists"

			results = append(results, ArbitrageResult{
				ExpectedValue: reboundAssistBet["price"].(float64),
				BookProb: odds,
				ModelProb: cdf,
				Differential: differential,
				Bet: reboundAssistBet,
			})
		}

		//points rebounds assists odds
			//CombinationSum
				//CombinationSum
					//cdf

		for _, pointReboundAssistBet := range pointsReboundsAssistsOdds {
			value := int64(math.Ceil(pointReboundAssistBet["point"].(float64))) // Perform type assertion to convert to int64
			combo := CombinationSum(points, rebounds)
			combo = CombinationSum(combo, assists)

			cdf := CDF(combo, value)

			if pointReboundAssistBet["name"] == "Under" {
				cdf = 1 - cdf
			}

			odds := 1.0/pointReboundAssistBet["price"].(float64)

			differential := cdf - odds

			pointReboundAssistBet["type"] = "points_rebounds_assists"

			results = append(results, ArbitrageResult{
				ExpectedValue: pointReboundAssistBet["price"].(float64),
				BookProb: odds,
				ModelProb: cdf,
				Differential: differential,
				Bet: pointReboundAssistBet,
			})

		}

	}


	
	

	// for player odds
		// get odds value and probability
		// get CDF of value for player
		// calculate differential 
		// return expected value and differential

	// for team odds
		// get odds value and probability
		// for totals
			// do combinationsum Under all player stats for that total to get team distribution
			// get CDF of value for team
			// calculate differential
			// return expected value and differential
		// for spreads
			// do combinationsum Under all player stats for that spread to get points distribution of both teams
			// calculate differentials between total points of both teams
			// get cdf of value for team differentials
			// calculate differential
			// return expected value and differential
		// for h2h
				// do combinationsum Under all player stats for that spread to get points distribution of both teams
				// calculate differentials between total points of both teams
				// get cdf of value for team differentials
				// calculate differential
				// return expected value and differential

		SaveToFile(results, oddspath, "arbitrage.json")

		return results
}

func SearchPlayerOdds(m []map[string]any, val string) []map[string]any {
	odds := make([]map[string]any, 0) // Specify the length and capacity of the slice
	for _, v := range m {
		if v["description"] == val {
			odds = append(odds, v)
		}
	}
	return odds
}

func IntSum(x []int64) int64 {
	var sum int64
	for _, v := range x {
		sum += v
	}
	return sum
}

func CombinationSum(x []int64, y []int64) []int64 {

	// sort x and y, split into groups of 10, take average
	slices.Sort(x)
	slices.Sort(y)

	var xaug []int64
	var yaug []int64

	for i := 0; i < len(x); i += 5000 {
    end := i + 5000
    if end > len(x) {
        end = len(x) // Ensure you don't exceed the slice length
    }
    xaug = append(xaug, IntSum(x[i:end])/int64(end-i)) // Use actual length of the slice for averaging
	}

	for i := 0; i < len(y); i += 5000 {
			end := i + 5000
			if end > len(y) {
					end = len(y) // Ensure you don't exceed the slice length
			}
			yaug = append(yaug, IntSum(y[i:end])/int64(end-i)) // Use actual length of the slice for averaging
	}

	var z []int64
	for i := 0; i < len(xaug); i++ {
		for j := 0; j < len(yaug); j++ {
			z = append(z, xaug[i]+yaug[j])
		}
	}
	return z
}

func Cartesian2(x []int64, y []int64) [][]int64 {
	var z [][]int64
	for i := 0; i < len(x); i++ {
		for j := 0; j < len(y); j++ {
			z = append(z, []int64{x[i], y[j]})
		}
	}
	return z
}

func Cartesian3(x []int64, y []int64, z []int64) [][]int64 {
	var w [][]int64
	for i := 0; i < len(x); i++ {
		for j := 0; j < len(y); j++ {
			for k := 0; k < len(z); k++ {
				w = append(w, []int64{x[i], y[j], z[k]})
			}
		}
	}
	return w
}

func CDF(x []int64, val int64) float64 {
	slices.Sort(x)

	count := 0

	for _, v := range x {
		if v <= val {
			count++
		}
	}

	return float64(count) / float64(len(x))
}

func CDF2(x []int64, y []int64, val []int64) float64 {
	if len(val) != 2 {
			panic("val must have 2 elements for CDF2")
	}
	cartesian := Cartesian2(x, y)
	count := 0

	for _, pair := range cartesian {
			if pair[0] <= val[0] && pair[1] <= val[1] {
					count++
			}
	}

	return float64(count) / float64(len(cartesian))
}

func CDF3(x []int64, y []int64, z []int64, val []int64) float64 {
	if len(val) != 3 {
			panic("val must have 3 elements for CDF3")
	}
	cartesian := Cartesian3(x, y, z)
	count := 0

	for _, triple := range cartesian {
			if triple[0] <= val[0] && triple[1] <= val[1] && triple[2] <= val[2] {
					count++
			}
	}

	return float64(count) / float64(len(cartesian))
}

// RenameDirsInDir renames all directories in a directory tree by replacing spaces with underscores
func RenameDirsInDir(dir string) error {
	// Read all entries in the directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// Iterate Under all entries in the directory
	for _, entry := range entries {
		oldName := entry.Name()
		oldPath := filepath.Join(dir, oldName)

		// Process directories (skip files)
		if entry.IsDir() {
			// Rename the directory if it contains spaces
			if strings.Contains(oldName, " ") {
				newName := strings.ReplaceAll(oldName, " ", "_")
				newPath := filepath.Join(dir, newName)
				err := os.Rename(oldPath, newPath)
				if err != nil {
					return fmt.Errorf("failed to rename directory %s to %s: %w", oldName, newName, err)
				}
				fmt.Printf("Renamed directory: %s -> %s\n", oldName, newName)
				oldPath = newPath // Update path to renamed directory
			}

			// Recursively rename subdirectories
			err := RenameDirsInDir(oldPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
// ReadOdds reads and combines JSON files from all directories and subdirectories into a single map
func ReadOdds(dir string) map[string]interface{} {
	// First, rename all directories
	err := RenameDirsInDir(dir)
	if err != nil {
		panic(fmt.Errorf("failed to rename directories: %w", err))
	}

	// Initialize the map to hold combined parsed data
	data := make(map[string]interface{})

	// Walk through the directory and all its subdirectories
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Process only JSON files
		if filepath.Ext(info.Name()) == ".json" {
			fmt.Println("Reading JSON file:", path)

			// Read the JSON file
			fileData, err := os.ReadFile(path) // Reads the entire file into memory
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", path, err)
			}

			// Parse the JSON file content
			var fileContent interface{}
			err = json.Unmarshal(fileData, &fileContent)
			if err != nil {
				return fmt.Errorf("failed to unmarshal JSON from file %s: %w", path, err)
			}

			// Accumulate parsed data in the map under the filename as the key
			data[path] = fileContent
		}
		return nil
	})

	if err != nil {
		panic(fmt.Errorf("failed to walk the directory: %w", err))
	}

	return data
}
// ReadPreds reads and parses CSV files from a directory and returns a map of player data
func ReadPreds(dir string) map[string]map[string][]int64 {
	// Initialize the map to hold parsed data with player names as keys
	data := make(map[string]map[string][]int64)

	// Read all files in the directory
	files, err := os.ReadDir(dir)
	if err != nil {
		panic(fmt.Errorf("failed to read directory: %w", err))
	}

	for _, file := range files {
		name := file.Name()

		// Match files with the expected naming pattern
		if !strings.HasPrefix(name, "team_stats_") || !strings.HasSuffix(name, "_posterior_predictive.csv") {
			continue
		}

		// Extract the player name
		player := strings.TrimSuffix(strings.TrimPrefix(name, "team_stats_"), "_posterior_predictive.csv")

		// Build the file path
		path := filepath.Join(dir, name)

		// Open the file
		f, err := os.Open(path)
		if err != nil {
			panic(fmt.Errorf("failed to open file %s: %w", path, err))
		}

		// Use defer in a separate function to ensure proper file closure
		func(file *os.File) {
			defer file.Close()

			// Create a CSV reader and read all records
			reader := csv.NewReader(file)
			records, err := reader.ReadAll()
			if err != nil {
				panic(fmt.Errorf("failed to read CSV data from file %s: %w", path, err))
			}

			// Read the header (first row)
			headers := records[0]

			// Initialize the player's data map with headers as keys
			playerData := make(map[string][]int64)

			// Parse CSV data into playerData map
			for i, record := range records[1:] {
				for j, value := range record {
					val, err := strconv.ParseInt(value, 10, 64)
					if err != nil {
						panic(fmt.Errorf("invalid integer value in file %s at row %d, column %d: %w", path, i+2, j+1, err)) // i+2 because of header row
					}
					// Use header as key and store the values
					playerData[headers[j]] = append(playerData[headers[j]], val)
				}
			}

			// change _ to space in player name
			player = strings.ReplaceAll(player, "_", " ")
			// Store the parsed player data in the map
			data[player] = playerData
		}(f)
	}

	return data
}
