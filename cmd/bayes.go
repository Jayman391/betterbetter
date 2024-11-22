// cmd/bayes.go

package cmd

import (
	"fmt"

	"betterbetter/src"

	"github.com/spf13/cobra"
	"gonum.org/v1/gonum/mat"
	"golang.org/x/exp/slices" // Ensure you have Go 1.21+ for the slices package
)

func init() {
	rootCmd.AddCommand(bayesCmd)
}

var bayesCmd = &cobra.Command{
	Use:   "bayes",
	Short: "Bayesian Network",
	Long:  `Bayesian Network`,
	Run: func(cmd *cobra.Command, args []string) {
		
		// Step 6: Initialize Priors for Mu and Sigma
		// Assuming priors for Mu and Sigma are both Normal distributions
		priorMuParams := src.DistributionParams{
			Dist: "Normal",
			Params: map[string]float64{
				"Mu":    0.0, // Prior mean for Mu
				"Sigma": 1.0, // Prior standard deviation for Mu
			},
		}
		priorSigmaParams := src.DistributionParams{
			Dist: "Normal",
			Params: map[string]float64{
				"Mu":    0.0, // Prior mean for Sigma
				"Sigma": 1.0, // Prior standard deviation for Sigma
			},
		}

		// Step 7: Create Prior Distributions
		priorMuDist := priorMuParams.CreateDist()
		
		priorSigmaDist := priorSigmaParams.CreateDist()
	

		// Step 8: Sample from Priors
		priorMuSamples := src.SampleDist(priorMuDist, 50)
		priorSigmaSamples := src.SampleDist(priorSigmaDist, 50)
		slices.Sort(priorMuSamples)
		slices.Sort(priorSigmaSamples)

		// Step 9: Combine Prior Samples into a Matrix
		numSamples := len(priorMuSamples)
		priorDataMat := mat.NewDense(numSamples, 2, nil) // 2 columns: Mu and Sigma
		for i := 0; i < numSamples; i++ {
			priorDataMat.Set(i, 0, priorMuSamples[i])    // Mu samples
			priorDataMat.Set(i, 1, priorSigmaSamples[i]) // Sigma samples
		}

		// Step 10: Define Priors
		priorObjMu := src.Prior{Distribution: priorMuDist}
		priorObjSigma := src.Prior{Distribution: priorSigmaDist}
		priors := []src.Prior{priorObjMu, priorObjSigma}

		// Step 11: Calculate the Posterior
		post := src.Posterior{
			Priors: priors,
			Data:   priorDataMat,
			LikelihoodParams: src.DistributionParams{
				Dist: "Normal",
				Params: map[string]float64{
					"Mu":    0.0, // Likelihood mean
					"Sigma": 1.0, // Likelihood standard deviation
				},
			},
		}

		posteriorResults := post.CalcPosterior()

		totalProb := 0.0


		for _, res := range posteriorResults {
			fmt.Printf("Params: Mu = %f, Sigma = %f, Probability: %f\n", res.Params[0], res.Params[1], res.Probability)
			totalProb += res.Probability
		}
		fmt.Println("Total Probability:", totalProb)
	},
}
