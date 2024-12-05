package cmd

import (
	//"fmt"

	"betterbetter/src"
	"math"

	"github.com/spf13/cobra"
	"gonum.org/v1/gonum/mat"

	"io/ioutil"
	"log"
	"os"
	"strings"
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
									
									averages := make([]float64, 4)
									for i := 0; i < 4; i++ {
										average := 0.0
										for _, lag := range lags {
											average += lag[i]
										}
										average /= float64(len(lags))
										averages[i] = average
									}

									averageAverage := src.Sum(averages) / float64(len(averages))
							
									
									priorRate1Params := src.DistributionParams{
										Dist: "Uniform",
										Params: map[string]float64{
											"Min": 0,
											"Max": 1,
										},
									}

									priorRate2Params := src.DistributionParams{
										Dist: "Uniform",
										Params: map[string]float64{
											"Min": 0,
											"Max": 1,
										},
									}

									priorRate3Params := src.DistributionParams{
										Dist: "Uniform",
										Params: map[string]float64{
											"Min": 0,
											"Max": 1,
										},
									}

									priorRate4Params := src.DistributionParams{
										Dist: "Uniform",
										Params: map[string]float64{
											"Min": 0,
											"Max": 1,
										},
									}

									interceptParams := src.DistributionParams{
										Dist: "Normal",
										Params: map[string]float64{
											"Mu": averageAverage,
											"Sigma": math.Sqrt(math.Abs(averageAverage) + 1),
										}, 
									}
									// Create Likelihood
									likelihoodParams := src.DistributionParams{
										Dist: "Normal",
										Params: map[string]float64{
											"Mu": 0,
											"Sigma": 1,
										},
									}

									
									// Create link func
									linkFunc := func(point []float64, data []float64) []float64 {
										lambda := 0.0
										for i, val := range data {
											lambda += val * point[i]
										}
										lambda += point[len(point)-1]
										return []float64{lambda , math.Sqrt(lambda)}
									}
									
									// turn lags to dense matrix

									if len(lags) > 1 {

										

										//take last value of lagmat and metric
										//metricTest := metric[:len(metric)-1]										
										metricTrain := metric[len(metric)-1:]

										lagmatTrain := mat.NewDense(len(lags)-1, len(lags[0]), nil)
										for i, lag := range lags[:len(lags)-1] {
											for j, val := range lag {
												lagmatTrain.Set(i, j, val)
											}
										}

										lagmatTest := mat.NewDense(1, len(lags[0]), nil)
										for i, lag := range lags[len(lags)-1:] {
											for j, val := range lag {
												lagmatTest.Set(i, j, val)
											}
										}
										

									

										likelihood := src.Likelihood{
											Params: []float64{1.0, 1.0, 1.0, 1.0, 1.0},
											DistributionParams: likelihoodParams,
											InputData: *lagmatTrain,
											OutputData: *mat.NewVecDense(len(metricTrain), metricTrain),
											Link: linkFunc,
										}
										

										// Create Markov Chain
										mc := src.MarkovChain{
											Distributions: []src.DistributionParams{priorRate1Params, priorRate2Params, 
												priorRate3Params, priorRate4Params, interceptParams},
											Grid: mat.Dense{},
											Likelihood: likelihood,
											SampleSize: 25,
											Sampler: "Metropolis",
										}

										priors := []src.DistributionParams{priorRate1Params, priorRate2Params, priorRate3Params, priorRate4Params, interceptParams}

										// Create Posterior
										posterior := src.Posterior{
											Priors: priors,
											Data: *lagmatTrain,
											LikelihoodParams: likelihoodParams,
											MarkovChain: mc,
										}

										posteriorResults := posterior.CalcPosterior()

										postPred := posterior.CalcPosteriorPredictive(posteriorResults, lagmatTest.RawMatrix().Data, 2500)
										// remove all negative values
										postPredReduced := make([]float64, 0)
										
										for _, val := range postPred {
											if val > 0 {
												postPredReduced = append(postPredReduced, val)
											}
										}
										// turn list of lists into list
										postPredList := append([]float64{}, postPredReduced...)

										playerPreds[player] = append(playerPreds[player], postPredList)

									}
								}

								// save playerPreds to file
								src.SaveToFile(playerPreds, "data/" + folder.Name() + "/" + year.Name() + "/" + team.Name() + "/preds/", player + "_preds.json")

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
