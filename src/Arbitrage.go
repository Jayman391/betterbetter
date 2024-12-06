package src

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Removed the ArbitrageResult struct since we no longer need it.

// Arbitrage now returns a slice of maps
func Arbitrage(statspath string, oddspath string) []map[string]any {
	results := make([]map[string]any, 0)

	oddsMap := make(map[string][]map[string]any)
	// Read odds from JSON files
	oddsData := ReadOdds(oddspath)
	// extract odds and store in oddsMap
	for _, d := range oddsData {
		data := d.(map[string]interface{})
		gameData := data["data"].(map[string]interface{})
		time := gameData["commence_time"]
		homeTeam := gameData["home_team"]
		awayTeam := gameData["away_team"]

		bookmakers := gameData["bookmakers"].([]interface{})
		for _, bookmaker := range bookmakers {
			markets := bookmaker.(map[string]interface{})["markets"].([]interface{})
			for _, m := range markets {
				marketMap := m.(map[string]interface{})
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

	playerPoints := oddsMap["player_points"]
	playerRebounds := oddsMap["player_rebounds"]
	playerAssists := oddsMap["player_assists"]
	playerBlocks := oddsMap["player_blocks"]
	playerSteals := oddsMap["player_steals"]
	playerTurnUnders := oddsMap["player_turnUnders"]
	playerPointsRebounds := oddsMap["player_points_rebounds"]
	playerPointsAssists := oddsMap["player_points_assists"]
	playerReboundsAssists := oddsMap["player_rebounds_assists"]
	playerPointsReboundsAssists := oddsMap["player_points_rebounds_assists"]

	// Read stats from directory
	stats := ReadPreds(statspath)
	for player, playerStats := range stats {
		points := playerStats["points"]
		rebounds := playerStats["totReb"]
		assists := playerStats["assists"]
		blocks := playerStats["blocks"]
		steals := playerStats["steals"]
		turnUnders := playerStats["turnovers"]

		playerName := strings.ReplaceAll(player, "_", " ")

		pointsOdds := SearchPlayerOdds(playerPoints, playerName)
		reboundsOdds := SearchPlayerOdds(playerRebounds, playerName)
		assistsOdds := SearchPlayerOdds(playerAssists, playerName)
		blocksOdds := SearchPlayerOdds(playerBlocks, playerName)
		stealsOdds := SearchPlayerOdds(playerSteals, playerName)
		turnUndersOdds := SearchPlayerOdds(playerTurnUnders, playerName)
		pointsReboundsOdds := SearchPlayerOdds(playerPointsRebounds, playerName)
		pointsAssistsOdds := SearchPlayerOdds(playerPointsAssists, playerName)
		reboundsAssistsOdds := SearchPlayerOdds(playerReboundsAssists, playerName)
		pointsReboundsAssistsOdds := SearchPlayerOdds(playerPointsReboundsAssists, playerName)

		// Points odds
		for _, pointBet := range pointsOdds {
			value := float64(math.Ceil(pointBet["point"].(float64)))
			cdf := CDF(points, value)
			if pointBet["name"] == "Under" {
				cdf = 1 - cdf
			}

			odds := 1.0 / pointBet["price"].(float64)
			differential := cdf - odds
			pointBet["type"] = "points"

			results = append(results, map[string]any{
				"ExpectedValue": pointBet["price"].(float64),
				"BookProb":      odds,
				"ModelProb":     cdf,
				"Differential":  differential,
				"Bet":           pointBet,
			})
		}

		// Rebounds odds
		for _, reboundBet := range reboundsOdds {
			value := float64(math.Ceil(reboundBet["point"].(float64)))
			cdf := CDF(rebounds, value)

			if reboundBet["name"] == "Under" {
				cdf = 1 - cdf
			}

			odds := 1.0 / reboundBet["price"].(float64)
			differential := cdf - odds
			reboundBet["type"] = "rebounds"

			results = append(results, map[string]any{
				"ExpectedValue": reboundBet["price"].(float64),
				"BookProb":      odds,
				"ModelProb":     cdf,
				"Differential":  differential,
				"Bet":           reboundBet,
			})
		}

		// Assist odds
		for _, assistBet := range assistsOdds {
			value := float64(math.Ceil(assistBet["point"].(float64)))
			cdf := CDF(assists, value)

			if assistBet["name"] == "Under" {
				cdf = 1 - cdf
			}

			odds := 1.0 / assistBet["price"].(float64)
			differential := cdf - odds
			assistBet["type"] = "assists"

			results = append(results, map[string]any{
				"ExpectedValue": assistBet["price"].(float64),
				"BookProb":      odds,
				"ModelProb":     cdf,
				"Differential":  differential,
				"Bet":           assistBet,
			})
		}

		// Blocks odds
		for _, blockBet := range blocksOdds {
			value := float64(math.Ceil(blockBet["point"].(float64)))
			cdf := CDF(blocks, value)

			if blockBet["name"] == "Under" {
				cdf = 1 - cdf
			}

			odds := 1.0 / blockBet["price"].(float64)
			differential := cdf - odds
			blockBet["type"] = "blocks"

			results = append(results, map[string]any{
				"ExpectedValue": blockBet["price"].(float64),
				"BookProb":      odds,
				"ModelProb":     cdf,
				"Differential":  differential,
				"Bet":           blockBet,
			})
		}

		// Steals odds
		for _, stealBet := range stealsOdds {
			value := float64(math.Ceil(stealBet["point"].(float64)))
			cdf := CDF(steals, value)

			if stealBet["name"] == "Under" {
				cdf = 1 - cdf
			}

			odds := 1.0 / stealBet["price"].(float64)
			differential := cdf - odds
			stealBet["type"] = "steals"

			results = append(results, map[string]any{
				"ExpectedValue": stealBet["price"].(float64),
				"BookProb":      odds,
				"ModelProb":     cdf,
				"Differential":  differential,
				"Bet":           stealBet,
			})
		}

		// Turnovers odds
		for _, turnUnderBet := range turnUndersOdds {
			value := float64(math.Ceil(turnUnderBet["point"].(float64)))
			cdf := CDF(turnUnders, value)

			if turnUnderBet["name"] == "Under" {
				cdf = 1 - cdf
			}

			odds := 1.0 / turnUnderBet["price"].(float64)
			differential := cdf - odds
			turnUnderBet["type"] = "turnovers"

			results = append(results, map[string]any{
				"ExpectedValue": turnUnderBet["price"].(float64),
				"BookProb":      odds,
				"ModelProb":     cdf,
				"Differential":  differential,
				"Bet":           turnUnderBet,
			})
		}

		// Points + Rebounds odds
		for _, pointReboundBet := range pointsReboundsOdds {
			value := float64(math.Ceil(pointReboundBet["point"].(float64)))
			combo := CombinationSum(points, rebounds)
			cdf := CDF(combo, value)

			if pointReboundBet["name"] == "Under" {
				cdf = 1 - cdf
			}

			odds := 1.0 / pointReboundBet["price"].(float64)
			differential := cdf - odds
			pointReboundBet["type"] = "points_rebounds"

			results = append(results, map[string]any{
				"ExpectedValue": pointReboundBet["price"].(float64),
				"BookProb":      odds,
				"ModelProb":     cdf,
				"Differential":  differential,
				"Bet":           pointReboundBet,
			})
		}

		// Points + Assists odds
		for _, pointAssistBet := range pointsAssistsOdds {
			value := float64(math.Ceil(pointAssistBet["point"].(float64)))
			combo := CombinationSum(points, assists)
			cdf := CDF(combo, value)

			if pointAssistBet["name"] == "Under" {
				cdf = 1 - cdf
			}

			odds := 1.0 / pointAssistBet["price"].(float64)
			differential := cdf - odds
			pointAssistBet["type"] = "points_assists"

			results = append(results, map[string]any{
				"ExpectedValue": pointAssistBet["price"].(float64),
				"BookProb":      odds,
				"ModelProb":     cdf,
				"Differential":  differential,
				"Bet":           pointAssistBet,
			})
		}

		// Rebounds + Assists odds
		for _, reboundAssistBet := range reboundsAssistsOdds {
			value := float64(math.Ceil(reboundAssistBet["point"].(float64)))
			combo := CombinationSum(rebounds, assists)
			cdf := CDF(combo, value)

			if reboundAssistBet["name"] == "Under" {
				cdf = 1 - cdf
			}

			odds := 1.0 / reboundAssistBet["price"].(float64)
			differential := cdf - odds
			reboundAssistBet["type"] = "rebounds_assists"

			results = append(results, map[string]any{
				"ExpectedValue": reboundAssistBet["price"].(float64),
				"BookProb":      odds,
				"ModelProb":     cdf,
				"Differential":  differential,
				"Bet":           reboundAssistBet,
			})
		}

		// Points + Rebounds + Assists odds
		for _, pointReboundAssistBet := range pointsReboundsAssistsOdds {
			value := float64(math.Ceil(pointReboundAssistBet["point"].(float64)))
			combo := CombinationSum(points, rebounds)
			combo = CombinationSum(combo, assists)

			cdf := CDF(combo, value)

			if pointReboundAssistBet["name"] == "Under" {
				cdf = 1 - cdf
			}

			odds := 1.0 / pointReboundAssistBet["price"].(float64)
			differential := cdf - odds
			pointReboundAssistBet["type"] = "points_rebounds_assists"

			results = append(results, map[string]any{
				"ExpectedValue": pointReboundAssistBet["price"].(float64),
				"BookProb":      odds,
				"ModelProb":     cdf,
				"Differential":  differential,
				"Bet":           pointReboundAssistBet,
			})
		}
	}

	fmt.Println(results)

	err := SaveResultsToFile(results, oddspath, "arbitrage.json")
	if err != nil {
			fmt.Printf("Error saving file: %v\n", err)
	} else {
			fmt.Println("File saved successfully.")
	}
	
	return results
}

func SearchPlayerOdds(m []map[string]any, val string) []map[string]any {
	odds := make([]map[string]any, 0)
	for _, v := range m {
		if v["description"] == val {
			odds = append(odds, v)
		}
	}
	return odds
}

// Sums a slice of float64
func FloatSum(x []float64) float64 {
	var sum float64
	for _, v := range x {
		sum += v
	}
	return sum
}

// CombinationSum approximates distribution combinations by sampling averages in chunks
func CombinationSum(x []float64, y []float64) []float64 {
	slices.Sort(x)
	slices.Sort(y)

	var xaug []float64
	var yaug []float64

	for i := 0; i < len(x); i += 5000 {
		end := i + 5000
		if end > len(x) {
			end = len(x)
		}
		xaug = append(xaug, FloatSum(x[i:end])/float64(end-i))
	}

	for i := 0; i < len(y); i += 5000 {
		end := i + 5000
		if end > len(y) {
			end = len(y)
		}
		yaug = append(yaug, FloatSum(y[i:end])/float64(end-i))
	}

	var z []float64
	for i := 0; i < len(xaug); i++ {
		for j := 0; j < len(yaug); j++ {
			z = append(z, xaug[i]+yaug[j])
		}
	}
	return z
}

func Cartesian2(x []float64, y []float64) [][]float64 {
	var z [][]float64
	for i := 0; i < len(x); i++ {
		for j := 0; j < len(y); j++ {
			z = append(z, []float64{x[i], y[j]})
		}
	}
	return z
}

func Cartesian3(x []float64, y []float64, w []float64) [][]float64 {
	var res [][]float64
	for i := 0; i < len(x); i++ {
		for j := 0; j < len(y); j++ {
			for k := 0; k < len(w); k++ {
				res = append(res, []float64{x[i], y[j], w[k]})
			}
		}
	}
	return res
}

func CDF(x []float64, val float64) float64 {
	slices.Sort(x)
	count := 0
	for _, v := range x {
		if v <= val {
			count++
		}
	}
	return float64(count) / float64(len(x))
}

func CDF2(x []float64, y []float64, val []float64) float64 {
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

func CDF3(x []float64, y []float64, z []float64, val []float64) float64 {
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

// RenameDirsInDir renames directories
func RenameDirsInDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		oldName := entry.Name()
		oldPath := filepath.Join(dir, oldName)

		if entry.IsDir() {
			if strings.Contains(oldName, " ") {
				newName := strings.ReplaceAll(oldName, " ", "_")
				newPath := filepath.Join(dir, newName)
				err := os.Rename(oldPath, newPath)
				if err != nil {
					return fmt.Errorf("failed to rename directory %s to %s: %w", oldName, newName, err)
				}
				fmt.Printf("Renamed directory: %s -> %s\n", oldName, newName)
				oldPath = newPath
			}

			err := RenameDirsInDir(oldPath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func ReadOdds(dir string) map[string]interface{} {
	err := RenameDirsInDir(dir)
	if err != nil {
		panic(fmt.Errorf("failed to rename directories: %w", err))
	}

	data := make(map[string]interface{})

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(info.Name()) == ".json" {
			fmt.Println("Reading JSON file:", path)
			fileData, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", path, err)
			}

			var fileContent interface{}
			err = json.Unmarshal(fileData, &fileContent)
			if err != nil {
				return fmt.Errorf("failed to unmarshal JSON from file %s: %w", path, err)
			}
			data[path] = fileContent
		}
		return nil
	})

	if err != nil {
		panic(fmt.Errorf("failed to walk the directory: %w", err))
	}

	return data
}

func ReadPreds(dir string) map[string]map[string][]float64 {
	data := make(map[string]map[string][]float64)
	files, err := os.ReadDir(dir)
	if err != nil {
		panic(fmt.Errorf("failed to read directory: %w", err))
	}

	for _, file := range files {
		name := file.Name()
		player := strings.TrimSuffix(name, "_preds.json")
		path := filepath.Join(dir, name)

		fmt.Println("Reading JSON file:", path)

		f, err := os.Open(path)
		if err != nil {
			panic(fmt.Errorf("failed to open file %s: %w", path, err))
		}
		defer f.Close()

		jsonData := make(map[string]interface{})
		err = json.NewDecoder(f).Decode(&jsonData)
		if err != nil {
			panic(fmt.Errorf("failed to decode JSON file %s: %w", path, err))
		}

		var key string
		for k := range jsonData {
			key = k
			break
		}
		playerPredsIface, ok := jsonData[key].([]interface{})
		if !ok {
			panic(fmt.Errorf("unexpected structure for player predictions"))
		}

		if len(playerPredsIface) < 6 {
			panic(fmt.Errorf("not enough prediction arrays for player %s", player))
		}

		points := toFloat64Slice(playerPredsIface[0].([]interface{}))
		assists := toFloat64Slice(playerPredsIface[1].([]interface{}))
		totReb := toFloat64Slice(playerPredsIface[2].([]interface{}))
		blocks := toFloat64Slice(playerPredsIface[3].([]interface{}))
		steals := toFloat64Slice(playerPredsIface[4].([]interface{}))
		turnovers := toFloat64Slice(playerPredsIface[5].([]interface{}))

		data[player] = map[string][]float64{
			"points":    points,
			"assists":   assists,
			"totReb":    totReb,
			"blocks":    blocks,
			"steals":    steals,
			"turnovers": turnovers,
		}
		fmt.Println(player)
	}
	return data
}

func toFloat64Slice(arr []interface{}) []float64 {
	out := make([]float64, len(arr))
	for i, v := range arr {
		f, ok := v.(float64)
		if !ok {
			panic(fmt.Errorf("expected a float64, got %T: %v", v, v))
		}
		out[i] = f
	}
	return out
}



// sanitizeResults replaces any NaN or Inf values with 0.0
func sanitizeResults(results []map[string]any) []map[string]any {
	for _, r := range results {
			for k, v := range r {
					if f, ok := v.(float64); ok {
							if math.IsNaN(f) || math.IsInf(f, 0) {
									r[k] = 0.0
							}
					}
					// If the value is nested map/array that might contain floats,
					// you'd need to recursively sanitize those as well.
			}
	}
	return results
}

func SaveResultsToFile(results []map[string]any, dir string, filename string) error {
	// Ensure directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directories for %s: %w", dir, err)
	}

	// Sanitize results to remove NaNs
	results = sanitizeResults(results)

	outputPath := filepath.Join(dir, filename)
	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
			return fmt.Errorf("failed to marshal results: %w", err)
	}

	if err := os.WriteFile(outputPath, jsonData, 0644); err != nil {
			return fmt.Errorf("failed to write results to file %s: %w", outputPath, err)
	}

	fmt.Printf("Results successfully written to %s\n", outputPath)
	return nil
}