package cmd

import (
	//"fmt"

	"betterbetter/src"

	"github.com/spf13/cobra"
	"gonum.org/v1/gonum/mat"

	"io/ioutil"
	"strings"
	"os"
	"log"
)

func init() {
	rootCmd.AddCommand(bayesCmd)
}

var bayesCmd = &cobra.Command{
	Use:   "bayes",
	Short: "Bayesian Stats",
	Long:  `Bayesian Stats`,
	Run: func(cmd *cobra.Command, args []string) {
		// data dir
		dir, err := ioutil.ReadDir("data")
		if err != nil {
			log.Fatal(err)
		}
		// sport dir
		for _, folder := range dir {
			years, err := ioutil.ReadDir("data/"+folder.Name())
			if err != nil {
				log.Fatal(err)
			}
			// year dir
			for _, year := range years {
				teams, err := ioutil.ReadDir("data/" + folder.Name() + "/" + year.Name())
				if err != nil {
					log.Fatal(err)
				}
				// team dir
				for _, team := range teams {
					data, err := ioutil.ReadDir("data/" + folder.Name() + "/" + year.Name() + "/" + team.Name())
					if err != nil {
						log.Fatal(err)
					}
					for _, file := range data {
						if strings.Contains(file.Name(), "stats.json") {
							stats, err := os.Open("data/" + folder.Name() + "/" + year.Name() + "/" + team.Name() + "/" + file.Name())
							if err != nil {
								log.Fatal(err)
							}

							data, err := ioutil.ReadAll(stats)
							if err != nil {
								log.Fatal(err)
							}
							statsData := src.ParseData(string(data))["response"]

							timeseries := CreateTimeseries(statsData.([]interface{}))

							for player, data := range timeseries {

								playerPreds := make(map[string][][]float64)

								for _, metric := range data {
									lags := CreateLags(metric, 4)
									// regression model 

									
									priorRate1Params := src.DistributionParams{
										Dist: "Exponential",
										Params: map[string]float64{
											"Rate": 0.075,
										},
									}
									priorRate1 := src.Prior{priorRate1Params.CreateDist()}

									priorRate2Params := src.DistributionParams{
										Dist: "Exponential",
										Params: map[string]float64{
											"Rate": 0.125,
										},
									}
									priorRate2 := src.Prior{priorRate2Params.CreateDist()}

									priorRate3Params := src.DistributionParams{
										Dist: "Exponential",
										Params: map[string]float64{
											"Rate": 0.25,
										},
									}
									priorRate3 := src.Prior{priorRate3Params.CreateDist()}

									priorRate4Params := src.DistributionParams{
										Dist: "Exponential",
										Params: map[string]float64{
											"Rate": 0.5,
										},
									}
									priorRate4 := src.Prior{priorRate4Params.CreateDist()}

									avg_metric := src.Sum(metric) / float64(len(metric))

									priorInterceptParams := src.DistributionParams{
										Dist: "Poisson",
										Params: map[string]float64{
											"Rate": avg_metric,
										},
									}
									priorIntercept := src.Prior{priorInterceptParams.CreateDist()}

									priors := []src.Prior{priorRate1, priorRate2, priorRate3, priorRate4, priorIntercept}

									// Create Likelihood
									likelihoodParams := src.DistributionParams{
										Dist: "Poisson",
										Params: map[string]float64{
											"Rate": avg_metric,
										},
									}

									// Create link func
									linkFunc := func(point []float64, data []float64) []float64 {
										return []float64{point[0]*data[0] + point[1]*data[1] + point[2]*data[2] + point[3]*data[3] + point[4]}
									}
									
									// turn lags to dense matrix

									if len(lags) > 1 {

										lagmat := mat.NewDense(len(lags), len(lags[0]), nil)
										for i, lag := range lags {
											for j, val := range lag {
												lagmat.Set(i, j, val)
											}
										}

										likelihood := src.Likelihood{
											Params: []float64{1.0, 1.0, 1.0, 1.0, 1.0},
											DistributionParams: likelihoodParams,
											Data: *lagmat,
											Link: linkFunc,
										}
										

										// Create Markov Chain
										mc := src.MarkovChain{
											Distributions: []src.Distribution{priorRate1Params.CreateDist(), priorRate2Params.CreateDist(), 
												priorRate3Params.CreateDist(), priorRate4Params.CreateDist(), priorInterceptParams.CreateDist()},
											Grid: mat.Dense{},
											Likelihood: likelihood,
											SampleSize: 25,
											Sampler: "Metropolis",
										}

										// Create Posterior
										posterior := src.Posterior{
											Priors: priors,
											Data: *lagmat,
											LikelihoodParams: likelihoodParams,
											MarkovChain: mc,
										}

										posteriorResults := posterior.CalcPosterior()

										postPred := posterior.CalcPosteriorPredictive(posteriorResults, 500)

										// turn list of lists into list
										postPredList := make([]float64, len(postPred))
										for i, val := range postPred {
											postPredList[i] = val[0]
										}

										playerPreds[player] = append(playerPreds[player], postPredList)

									}
								}

								// save playerPreds to file
								src.SaveToFile(playerPreds, "data/" + folder.Name() + "/" + year.Name() + "/" + team.Name(), player + "_preds.json")

							}

							
						}
					}	
				}
			}
		}
	},
}

func CreateTimeseries(data []interface{}) map[string] [][]float64{
	
	playerData := make(map[string][][]float64)

	for _, gameData := range data {
		gameData := gameData.(map[string]interface{})
		player := gameData["player"].(map[string]interface{})
		firstName := player["firstname"].(string)
		lastName := player["lastname"].(string)

		name := firstName + "_" + lastName
		points := gameData["points"].(float64)
		assists := gameData["assists"].(float64)
		rebounds := gameData["totReb"].(float64)
		blocks := gameData["blocks"].(float64)
		steals := gameData["steals"].(float64)
		turnovers := gameData["turnovers"].(float64)

		playerData[name] = append(playerData[name], []float64{points, assists, rebounds, blocks, steals, turnovers})
	}

	for player, data := range playerData {
		points := make([]float64, len(data))
		assists := make([]float64, len(data))
		rebounds := make([]float64, len(data))
		blocks := make([]float64, len(data))
		steals := make([]float64, len(data))
		turnovers := make([]float64, len(data))

		for i, gameData := range data {
			points[i] = gameData[0]
			assists[i] = gameData[1]
			rebounds[i] = gameData[2]
			blocks[i] = gameData[3]
			steals[i] = gameData[4]
			turnovers[i] = gameData[5]
		}
		playerData[player] = [][]float64{points, assists, rebounds, blocks, steals, turnovers}
	}

	return playerData
	
}

func CreateLags(data []float64, lags int) [][]float64 {
	if len(data) > lags {
		lagData := make([][]float64, len(data)-lags)
		for i := 0; i < len(data)-lags; i++ {
			lagData[i] = data[i:i+lags]
		}
		return lagData
	} else {
		return nil
	}
	
}
