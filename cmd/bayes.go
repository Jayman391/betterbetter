package cmd

import (
	"betterbetter/src"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"slices"
	"gonum.org/v1/gonum/mat"
	"github.com/spatialcurrent/go-math/pkg/math"
)

func init() {
	rootCmd.AddCommand(bayesCmd)
}

var bayesCmd = &cobra.Command{
	Use:   "bayes",
	Short: "Bayesian Network",
	Long:  `Bayesian Network`,
	Run: func(cmd *cobra.Command, args []string) {
			// Initialize DistributionParams with exported fields
			testdist := src.DistributionParams{
					Dist: "Normal",
					Params: map[string]float64{
							"Mu":    0.0,
							"Sigma": 1.0,
					},
			}

			// Invoke the CreateDist method
			dist := testdist.CreateDist()
			if dist == nil {
					log.Fatal("Unsupported distribution type or missing parameters.")
			}

			samples := src.SampleDist(dist, 1000)


			// sort the samples
			

			slices.Sort(samples)


		
			// turn sampledata tp a mat.Matrix
			
			sampledataMat := mat.NewDense(len(samples), 1, samples)

			// get first sample and make it a list
		  sampledatatest := []float64{samples[500]}

			likelihood := src.Likelihood{sampledatatest, sampledataMat}

			calculated := likelihood.CalcLikelihood()

			sum, err := math.Sum(calculated)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(sum)

			anothersampledatatest := []float64{samples[0]}

			likelihood2 := src.Likelihood{anothersampledatatest, sampledataMat}

			calculated2 := likelihood2.CalcLikelihood()

			sum2, err := math.Sum(calculated2)

			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(sum2)

			aanothersampledatatest := []float64{samples[999]}

			alikelihood2 := src.Likelihood{aanothersampledatatest, sampledataMat}

			acalculated2 := alikelihood2.CalcLikelihood()

			asum2, err := math.Sum(acalculated2)

			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(asum2)

	},
}