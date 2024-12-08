package cmd

import (
	"betterbetter/src"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"gonum.org/v1/gonum/mat"
)

func init() {
	var chains int
	var trainSamples int
	var testSamples int

	bayesCmd.Flags().IntVarP(&lags, "lags", "l", 4, "Number of lags")
	bayesCmd.Flags().IntVarP(&chains, "chains", "c", 4, "Number of chains")
	bayesCmd.Flags().IntVarP(&trainSamples, "train", "e", 5, "Number of posterior predictive examples")
	bayesCmd.Flags().IntVarP(&testSamples, "test", "s", 500, "Number of samples for posterior predictive")

	rootCmd.AddCommand(bayesCmd)
}

var lags int

var bayesCmd = &cobra.Command{
	Use:   "bayes",
	Short: "Bayesian Stats",
	Long:  `Bayesian Stats`,
	Run: func(cmd *cobra.Command, args []string) {
		dir, err := ioutil.ReadDir("data")
		if err != nil {
			log.Fatal(err)
		}

		chains, err := strconv.Atoi(cmd.Flag("chains").Value.String())
		if err != nil {
			log.Fatal(err)
		}

		testSamples, err := strconv.Atoi(cmd.Flag("test").Value.String())
		if err != nil {
			log.Fatal(err)
		}

		for _, folder := range dir {
			years, err := ioutil.ReadDir("data/" + folder.Name())
			if err != nil {
				log.Fatal(err)
			}
			for _, year := range years {
				teams, err := ioutil.ReadDir("data/" + folder.Name() + "/" + year.Name())
				if err != nil {
					log.Fatal(err)
				}
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

							rawData, err := ioutil.ReadAll(stats)
							if err != nil {
								log.Fatal(err)
							}
							statsData := src.ParseData(string(rawData))["response"]

							timeseries := CreateTimeseries(statsData.([]interface{}))

							for player, data := range timeseries {

								playerPreds := make(map[string][][]float64)

								for name, metric := range data {
									lagMatrix := CreateLags(metric, lags)

									// Create priors dynamically based on number of lags
									var priors []src.DistributionParams
									for i := 0; i < lags; i++ {
										priorRateParams := src.DistributionParams{
											Dist: "Uniform",
											Params: map[string]float64{
												"Min": 0,
												"Max": 1,
											},
										}
										priors = append(priors, priorRateParams)
									}

									interceptMean := src.Sum(metric) / float64(len(metric))
									interceptParams := src.DistributionParams{
										Dist: "Normal",
										Params: map[string]float64{
											"Mu":    interceptMean,
											"Sigma": math.Sqrt(1 + interceptMean),
										},
									}
									priors = append(priors, interceptParams)

									// Create Likelihood
									likelihoodParams := src.DistributionParams{
										Dist: "Normal",
										Params: map[string]float64{
											"Mu":    0,
											"Sigma": 1,
										},
									}

									linkFunc := func(point []float64, data []float64) []float64 {
										lambda := 0.0
										for i, val := range data {
											lambda += val * point[i]
										}
										lambda += point[len(point)-1] // intercept
										return []float64{math.Max(lambda, 0), math.Abs(lambda)}
									}

									trainSamples, _ := strconv.Atoi(cmd.Flag("train").Value.String())

									// Training data: exclude last `testSamples` observations
									trainSize := trainSamples
								
									if trainSize > 0 && len(lagMatrix) >= trainSize && len(lagMatrix[:trainSize]) == trainSize {
										fmt.Println("Training on", player, "with", trainSize, "samples")
										lagmatTrain := mat.NewDense(trainSize, len(lagMatrix[0]), nil)
										for i, lag := range lagMatrix[:trainSize] {
											for j, val := range lag {
												lagmatTrain.Set(i, j, val)
											}
										}


										// Test data: last `testSamples` rows
										lagmatTest := mat.NewDense(testSamples, len(lagMatrix[0]), nil)
										for i, lag := range lagMatrix[trainSize:] {
											for j, val := range lag {
												lagmatTest.Set(i, j, val)
											}
										}

										// Output data: assume metric is aligned with lagMatrix, training excludes last testSamples
										metricTrain := metric[:len(metric)-trainSamples]

										initialParams := make([]float64, lags+1)
										for i := range initialParams {
											initialParams[i] = 1.0
										}

										likelihood := src.Likelihood{
											Params:             initialParams,
											DistributionParams: likelihoodParams,
											InputData:          *lagmatTrain,
											OutputData:         *mat.NewVecDense(len(metricTrain), metricTrain),
											Link:               linkFunc,
										}

										mc := src.MarkovChain{
											Distributions: priors,
											Grid:          mat.Dense{},
											Likelihood:    likelihood,
											SampleSize:    25,
											Sampler:       "Metropolis",
										}

										posterior := src.Posterior{
											Priors:           priors,
											Data:             *lagmatTrain,
											LikelihoodParams: likelihoodParams,
											MarkovChain:      mc,
										}

										fmt.Println("Calculating Posterior for", player, "with", len(metricTrain), "training samples")

										posteriorResults := posterior.CalcPosterior(chains)
										
										fmt.Println(lagmatTrain)

										// Updated call to CalcPosteriorPredictive with multiple test examples
										// We'll pass the entire lagmatTest data and let the function handle multiple rows.
										// Convert lagmatTest (a mat.Dense) to [][]float64
										rows, cols := lagmatTest.Dims()
										testData := make([][]float64, rows)
										for i := 0; i < rows; i++ {
												testData[i] = make([]float64, cols)
												for j := 0; j < cols; j++ {
														testData[i][j] = lagmatTest.At(i, j)
												}
										}

										fmt.Println("Calculating Posterior Predictive for", player, "with metric " , name)

										// Now call CalcPosteriorPredictive with the testData as [][]float64
										postPred := posterior.CalcPosteriorPredictive(
												posteriorResults,
												testData,
												testSamples,
												linkFunc,
										)

										postPredFiltered := make([]float64, len(postPred))
										// take min value and add that to every element
										for _, val := range postPred {
												if val > 0 {
													postPredFiltered = append(postPredFiltered, val)
												}
										}
										

										playerPreds[player] = append(playerPreds[player], postPredFiltered)
										src.SaveToFile(playerPreds, "data/"+folder.Name()+"/"+year.Name()+"/"+team.Name()+"/preds/", player+"_preds.json")
										}
								}
							}
						}
					}
				}
			}
		}
	},
}

func CreateTimeseries(data []interface{}) map[string][][]float64 {
	playerData := make(map[string][][]float64)

	for _, gd := range data {
		gameData := gd.(map[string]interface{})
		player := gameData["player"].(map[string]interface{})
		firstName := player["firstname"].(string)
		lastName := player["lastname"].(string)

		name := firstName + "_" + lastName
		points := gameData["points"].(float64)
		rebounds := gameData["totReb"].(float64)
		assists := gameData["assists"].(float64)
		blocks := gameData["blocks"].(float64)
		steals := gameData["steals"].(float64)
		turnovers := gameData["turnovers"].(float64)

		playerData[name] = append(playerData[name], []float64{points, rebounds, assists, blocks, steals, turnovers})
	}

	for player, pdata := range playerData {
		points := make([]float64, len(pdata))
		rebounds := make([]float64, len(pdata))
		assists := make([]float64, len(pdata))
		blocks := make([]float64, len(pdata))
		steals := make([]float64, len(pdata))
		turnovers := make([]float64, len(pdata))

		for i, gameData := range pdata {
			points[i] = gameData[0]
			rebounds[i] = gameData[1]
			assists[i] = gameData[2]
			blocks[i] = gameData[3]
			steals[i] = gameData[4]
			turnovers[i] = gameData[5]
		}

		playerData[player] = [][]float64{points, rebounds, assists, blocks, steals, turnovers}
	}

	return playerData
}

func CreateLags(data []float64, lags int) [][]float64 {
	if len(data) <= lags {
		return nil
	}
	lagData := make([][]float64, len(data)-lags)
	for i := 0; i < len(data)-lags; i++ {
		lagData[i] = data[i : i+lags]
	}
	return lagData
}
