package src

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
)

// Define the Bet struct
type Bet struct {
	Combo        []int
	BookProfit   float64
	BookProbs    float64
	ModelProfit  float64
	ModelProbs   float64
	Differential float64
	EV           float64
	Bets         []map[string]interface{}
}

func MakeBets(rr float64, maxbets int) []map[string]interface{} {
	rr += 1.0
	arbs := LoadData()

	// Initialize a slice to store Bet structs
	var betsList []Bet

	fmt.Println(arbs)

	// For each date in arbs
	for _, arb := range arbs {
		arblist := arb.([]map[string]interface{})
		var betCombos [][]int
		// Get number of bets
		numBets := len(arblist)
		fmt.Println(numBets)
		// Generate combinations of bets (1, 2, or 3 bets at a time)
		for i := 0; i < numBets; i++ {
			betCombos = append(betCombos, []int{i})
			for j := i + 1; j < numBets; j++ {
				betCombos = append(betCombos, []int{i, j})
			}
		}
		fmt.Println(betCombos)

		// Iterate through each combination and calculate profits
	outerLoop:
		for _, combo := range betCombos {
			EV := 1.0
			bookProbs := 1.0
			modelProbs := 1.0

			actualBets := make([]map[string]interface{}, 0)
			betKeys := make(map[string]bool)        // To store unique bets
			conflictKeys := make(map[string]string) // To detect over/under conflicts

			// Calculate EV, bookProbs, and modelProbs for each combination
			for _, i := range combo {
				if arblist[i] != nil {
					// Extract 'Bet' field
					betData, ok := arblist[i]["Bet"].(map[string]interface{})
					if !ok {
						continue outerLoop
					}

					// Generate a unique key for the bet
					betKey := generateBetKey(betData)

					// Check if bet is already in the set
					if betKeys[betKey] {
						// Duplicate bet found, skip this combination
						continue outerLoop
					}
					// Add betKey to the set
					betKeys[betKey] = true

					// Get 'name' field (Over or Under)
					side, ok := betData["name"].(string)
					if !ok {
						continue outerLoop
					}

					// Generate conflictKey (excluding 'name')
					conflictKey := generateConflictKey(betData)

					// Check for over/under conflicts
					if existingSide, exists := conflictKeys[conflictKey]; exists {
						if existingSide != side {
							// Over and Under on same bet, skip combination
							continue outerLoop
						}
					} else {
						// Add to conflictKeys
						conflictKeys[conflictKey] = side
					}

					// **New Check**: Exclude bets where BookProb > ModelProb
					bookProb, ok1 := arblist[i]["BookProb"].(float64)
					modelProb, ok2 := arblist[i]["ModelProb"].(float64)
					if ok1 && ok2 && bookProb > modelProb {
						// Skip this combination if any bet has BookProb > ModelProb
						continue outerLoop
					}

					// Now proceed to process the bet
					if expectedValue, ok := arblist[i]["ExpectedValue"].(float64); ok {
						EV *= expectedValue
					}
					if ok1 {
						bookProbs *= bookProb
					}
					if ok2 {
						modelProbs *= modelProb
					}
					actualBets = append(actualBets, arblist[i])
				}
			}

			// Calculate profits
			bookProfit := (EV - 1) * bookProbs
			modelProfit := (EV - 1) * modelProbs

			// Create Bet struct
			bet := Bet{
				Combo:        combo,
				BookProfit:   bookProfit,
				BookProbs:    bookProbs,
				ModelProfit:  modelProfit,
				ModelProbs:   modelProbs,
				Differential: modelProbs - bookProbs,
				EV:           EV,
				Bets:         actualBets,
			}

			// Add the bet keys to the Bet struct for later use
			// (We can store comboBetKeys here if needed)

			betsList = append(betsList, bet)
		}
	}

	// Sort the bets by differential in descending order
	sort.Slice(betsList, func(i, j int) bool {
		return betsList[i].Differential > betsList[j].Differential
	})

	// Filter out bets with EV < rr
	var filteredBets []Bet
	for _, bet := range betsList {
		if bet.EV <= rr {
			filteredBets = append(filteredBets, bet)
		}
	}

	// Select the top maxbets bets
	var selectedBets []Bet
	combinationKeys := make(map[string]bool)

	for _, bet := range filteredBets {
		if len(selectedBets) >= maxbets {
			break
		}

		// Generate a unique key for the combination based on bet keys
		comboBetKeys := []string{}
		for _, betMap := range bet.Bets {
			betData, ok := betMap["Bet"].(map[string]interface{})
			if !ok {
				continue
			}
			betKey := generateBetKey(betData)
			comboBetKeys = append(comboBetKeys, betKey)
		}

		// Sort the bet keys to ensure consistent combination keys
		sort.Strings(comboBetKeys)
		comboKey := generateComboKeyFromBetKeys(comboBetKeys)

		// Check if this combination is already included
		if combinationKeys[comboKey] {
			// Combination already included, skip
			continue
		}

		// Add to selectedBets and mark the combination
		selectedBets = append(selectedBets, bet)
		combinationKeys[comboKey] = true
	}

	// Now, sort the selectedBets by Probability Differential in descending order
	sort.Slice(selectedBets, func(i, j int) bool {
		return selectedBets[i].Differential > selectedBets[j].Differential
	})

	// Prepare the output as a slice of maps
	var output []map[string]interface{}

	// Loop through selectedBets and prepare the output
	for _, bet := range selectedBets {
		// Create a key for the combo
		comboKey := fmt.Sprintf("%v", bet.Combo)

		// Add the actual bets and profit information to the output
		output = append(output, map[string]interface{}{
			"combo":        comboKey,
			"bookProfit":   bet.BookProfit,
			"bookProbs":    bet.BookProbs,
			"modelProfit":  bet.ModelProfit,
			"modelProbs":   bet.ModelProbs,
			"differential": bet.Differential,
			"EV":           bet.EV,
			"bets":         bet.Bets, // Add the actual bets for this combination
		})
	}

	// Print total profits for debugging
	var totalBookProfit, totalModelProfit float64
	for _, bet := range selectedBets {
		totalBookProfit += bet.BookProfit
		totalModelProfit += bet.ModelProfit
	}

	fmt.Printf("Total Book Profit for %d bets: %.2f\n", len(selectedBets), totalBookProfit)
	fmt.Printf("Total Model Profit for %d bets: %.2f\n", len(selectedBets), totalModelProfit)

	fmt.Println("Selected Bets:")
	for _, bet := range selectedBets {
		fmt.Printf("Combo: %v\n", bet.Combo)
		fmt.Printf("Book Profit: %.2f\n", bet.BookProfit)
		fmt.Printf("Model Profit: %.2f\n", bet.ModelProfit)
		fmt.Printf("Probability Differential: %.2f\n", bet.Differential)
		fmt.Printf("EV: %.2f\n", bet.EV)
		fmt.Printf("Bets: %v\n", bet.Bets)
		fmt.Println()
	}

	// Return the selected bets
	return output
}

// Function to generate a unique key for each bet (including 'name')
func generateBetKey(betData map[string]interface{}) string {
	// Extract fields that uniquely define a bet, including 'name'
	awayTeam := fmt.Sprintf("%v", betData["away_team"])
	homeTeam := fmt.Sprintf("%v", betData["home_team"])
	name := fmt.Sprintf("%v", betData["name"])
	point := fmt.Sprintf("%v", betData["point"])
	time := fmt.Sprintf("%v", betData["time"])
	betType := fmt.Sprintf("%v", betData["type"])
	description := fmt.Sprintf("%v", betData["description"])

	// Combine them into a unique key
	key := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s", awayTeam, homeTeam, name, point, time, betType, description)
	return key
}

// Function to generate a conflict key (excluding 'name')
func generateConflictKey(betData map[string]interface{}) string {
	// Extract fields that uniquely define a bet, excluding 'name'
	awayTeam := fmt.Sprintf("%v", betData["away_team"])
	homeTeam := fmt.Sprintf("%v", betData["home_team"])
	point := fmt.Sprintf("%v", betData["point"])
	time := fmt.Sprintf("%v", betData["time"])
	betType := fmt.Sprintf("%v", betData["type"])
	description := fmt.Sprintf("%v", betData["description"])

	// Combine them into a conflict key
	key := fmt.Sprintf("%s|%s|%s|%s|%s|%s", awayTeam, homeTeam, point, time, betType, description)
	return key
}

// Function to generate a unique key for a combination based on bet keys
func generateComboKeyFromBetKeys(betKeys []string) string {
	// Since betKeys are sorted, we can concatenate them to form a unique key
	return fmt.Sprintf("%v", betKeys)
}

func LoadData() map[string]interface{} {
	arbs := make(map[string]interface{})
	// read all folders in data folder (data => sport)
	dir, err := ioutil.ReadDir("data")
	if err != nil {
		log.Fatal(err)
	}
	// sport dir
	for _, folder := range dir {
		years, err := ioutil.ReadDir("data/" + folder.Name())
		if err != nil {
			log.Fatal(err)
		}
		// year dir
		for _, year := range years {
			teams, err := ioutil.ReadDir("data/" + folder.Name() + "/" + year.Name())
			if err != nil {
				log.Fatal(err)
			}
			// date dir
			for _, date := range teams {
				data, err := ioutil.ReadDir("data/" + folder.Name() + "/" + year.Name() + "/" + date.Name())
				if err != nil {
					log.Fatal(err)
				}
				for _, file := range data {
					if file.Name() == "arbitrage.json" {
						// read file
						jsonFile, err := os.Open("data/" + folder.Name() + "/" + year.Name() + "/" + date.Name() + "/" + file.Name())
						if err != nil {
							log.Fatal(err)
						}
						defer jsonFile.Close()

						data, err := ioutil.ReadAll(jsonFile)
						if err != nil {
							log.Fatal(err)
						}

						var arbMap []map[string]interface{}
						err = json.Unmarshal(data, &arbMap)
						if err != nil {
							log.Fatal(err)
						}

						arbs[folder.Name()+year.Name()+date.Name()] = arbMap
					}
				}
			}
		}
	}
	return arbs
}
