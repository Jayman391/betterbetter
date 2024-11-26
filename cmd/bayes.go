package cmd

import (
	"fmt"

	"betterbetter/src"

	"github.com/spf13/cobra"
	"gonum.org/v1/gonum/mat"
)

func init() {
	rootCmd.AddCommand(bayesCmd)
}

var bayesCmd = &cobra.Command{
	Use:   "bayes",
	Short: "Bayesian Stats",
	Long:  `Bayesian Stats`,
	Run: func(cmd *cobra.Command, args []string) {
		
		// Step 6: Initialize Priors for Mu and Sigma
		// Assuming priors for Mu and Sigma are both Normal distributions
		priorMuParams := src.DistributionParams{
			Dist: "Normal",
			Params: map[string]float64{
				"Mu": 1.0,
				"Sigma": 0.5,
			},
		}
		priorSigmaParams := src.DistributionParams{
			Dist: "Exponential",
			Params: map[string]float64{
				"Rate": 0.5,
			},
		}

		// Step 7: Create Prior Distributions
		priorMuDist := priorMuParams.CreateDist()
		
		priorSigmaDist := priorSigmaParams.CreateDist()
	

		// Step 8: generate Fake Data
		dataparams := src.DistributionParams{
			Dist: "Normal",
			Params: map[string]float64{
				"Mu":    1.0, // Likelihood mean
				"Sigma": 1.0, // Likelihood standard deviation
			},
		}

		datadist := dataparams.CreateDist()

		data := src.SampleDist(datadist, 1000)

		datamatrix := mat.NewDense(len(data), 1, data)

		// Step 9: Define Priors
		priorObjMu := src.Prior{Distribution: priorMuDist}
		priorObjSigma := src.Prior{Distribution: priorSigmaDist}
		priors := []src.Prior{priorObjMu, priorObjSigma}

		

		// Step 10: Create Likelihood

		LikelihoodParams := src.DistributionParams{
			Dist: "Normal",
			Params: map[string]float64{
				"Mu":    0.0, // Likelihood mean
				"Sigma": 1.0, // Likelihood standard deviation
			},
		}

		Likelihood := src.Likelihood{
			Params : []float64{0.0, 1.0},
			DistributionParams: LikelihoodParams,
			Data: datamatrix,
		}

		// Step 11: Create Markov Chain
		mc := src.MarkovChain{
			Distributions: []src.Distribution{priorMuDist, priorSigmaDist},
			Grid: 				mat.Dense{},
			Likelihood: Likelihood,
			SampleSize: 20,
			Sampler: "Metropolis",
		}

		// Step 12: Create Posterior

		post := src.Posterior{
			Priors: priors,
			Data:   datamatrix,
			LikelihoodParams: LikelihoodParams,
			MarkovChain: mc,
		}

		posteriorResults := post.CalcPosterior()

		totalProb := 0.0


		for _, res := range posteriorResults {
			fmt.Printf("Params: Mu = %f, Sigma = %f, Probability: %f\n", res.Params[0], res.Params[1], res.Probability)
			totalProb += res.Probability
		}
		fmt.Println("Total Probability:", totalProb)

		// Sample posterior predictive

		postPred := post.CalcPosteriorPredictive(posteriorResults, 100)

	  fmt.Println("Posterior Predictive Samples: ", postPred)		
	},
}
