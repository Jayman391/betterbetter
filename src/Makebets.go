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

func MakeBets(rr float64, maxbets int) map[string]interface{} {
	arbs := LoadData()

	// Initialize a slice to store Bet structs
	var betsList []Bet

	// For each date in arbs
	for _, arb := range arbs {
		arblist := arb.([]map[string]interface{})
		var betCombos [][]int
		// Get number of bets
		numBets := len(arblist)

		// Generate combinations of bets (1, 2, or 3 bets at a time)
		for i := 0; i < numBets; i++ {
			betCombos = append(betCombos, []int{i})
			for j := i + 1; j < numBets; j++ {
				betCombos = append(betCombos, []int{i, j})
				for k := j + 1; k < numBets; k++ {
					betCombos = append(betCombos, []int{i, j, k})
				}
			}
		}

		// Iterate through each combination and calculate profits
	outerLoop:
		for _, combo := range betCombos {
			EV := 1.0
			bookProbs := 1.0
			modelProbs := 1.0

			actualBets := make([]map[string]interface{}, 0)
			betKeys := make(map[string]bool) // To store unique bets

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

					if expectedValue, ok := arblist[i]["ExpectedValue"].(float64); ok {
						EV *= expectedValue
					}
					if bookprob, ok := arblist[i]["BookProb"].(float64); ok {
						bookProbs *= bookprob
					}
					if modelprob, ok := arblist[i]["ModelProb"].(float64); ok {
						modelProbs *= modelprob
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

	// Select the top maxbets bets by modelProfit
	selectedBets := make(map[string]interface{})

	// Loop through filteredBets and add the actual bets to selectedBets
	for i := 0; i < len(filteredBets) && i < maxbets; i++ {
		bet := filteredBets[i]

		// Create a key for the combo
		comboKey := fmt.Sprintf("%v", bet.Combo)

		// Add the actual bets and profit information to selectedBets
		selectedBets[comboKey] = map[string]interface{}{
			"bookProfit":   bet.BookProfit,
			"bookProbs":    bet.BookProbs,
			"modelProfit":  bet.ModelProfit,
			"modelProbs":   bet.ModelProbs,
			"differential": bet.Differential,
			"EV":           bet.EV,
			"bets":         bet.Bets, // Add the actual bets for this combination
		}
	}

	// Print total profits for debugging
	var totalBookProfit, totalModelProfit float64
	for _, bet := range selectedBets {
		betInfo := bet.(map[string]interface{})
		totalBookProfit += betInfo["bookProfit"].(float64)
		totalModelProfit += betInfo["modelProfit"].(float64)
	}

	fmt.Printf("Total Book Profit for %d bets: %.2f\n", len(selectedBets), totalBookProfit)
	fmt.Printf("Total Model Profit for %d bets: %.2f\n", len(selectedBets), totalModelProfit)

	fmt.Println("Selected Bets:")
	for combo, bet := range selectedBets {
		fmt.Printf("Combo: %s\n", combo)
		fmt.Printf("Book Profit: %.2f\n", bet.(map[string]interface{})["bookProfit"].(float64))
		fmt.Printf("Model Profit: %.2f\n", bet.(map[string]interface{})["modelProfit"].(float64))
		fmt.Printf("Probability Differential: %.2f\n", bet.(map[string]interface{})["differential"].(float64))
		fmt.Printf("EV: %.2f\n", bet.(map[string]interface{})["EV"].(float64))
		fmt.Printf("Bets: %v\n", bet.(map[string]interface{})["bets"])
		fmt.Println()
	}

	// Return the selected bets
	return selectedBets
}

// Function to generate a unique key for each bet
func generateBetKey(betData map[string]interface{}) string {
	// Extract fields that uniquely define a bet
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
