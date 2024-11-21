package src

import (
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat/distuv"
	"github.com/schwarmco/go-cartesian-product"

)

// Define the Distribution interface
type Distribution interface {
	Rand() float64
}

type DistributionParams struct {
	Dist   string
	Params map[string]float64
}
func (d *DistributionParams) CreateDist() Distribution {
	switch d.Dist {
		case "Bernoulli":
				return distuv.Bernoulli{
						P: d.Params["P"],
				}
		case "Beta":
				return distuv.Beta{
						Alpha: d.Params["Alpha"],
						Beta:  d.Params["Beta"],
				}
		case "Binomial":
				return distuv.Binomial{
						N: d.Params["N"],
						P: d.Params["P"],
				}	
		case "ChiSquared":
				return distuv.ChiSquared{
						K: d.Params["K"],
				}
		case "Exponential":
				return distuv.Exponential{
						Rate: d.Params["Rate"],
				}
		case "Gamma":
				return distuv.Gamma{
						Alpha: d.Params["Alpha"],
						Beta:  d.Params["Beta"],
				}
		case "LogNormal":
				return distuv.LogNormal{
						Mu:    d.Params["Mu"],
						Sigma: d.Params["Sigma"],
				}
		case "Normal":
				return distuv.Normal{
						Mu:    d.Params["Mu"],
						Sigma: d.Params["Sigma"],
				}
		case "Pareto":
				return distuv.Pareto{
						Xm:    d.Params["Xm"],
						Alpha: d.Params["Alpha"],
				}
		case "Poisson":
				return distuv.Poisson{
						Lambda: d.Params["Lambda"],
				}
		case "StudentsT":
				return distuv.StudentsT{
						Mu:    d.Params["Mu"],
						Sigma: d.Params["Sigma"],
						Nu:    d.Params["Nu"],
				}
		case "Uniform":
				return distuv.Uniform{
						Min: d.Params["Min"],
						Max: d.Params["Max"],
				}
		case "Weibull":
				return distuv.Weibull{
						K:      d.Params["K"],
						Lambda: d.Params["Lambda"],
				}
		default:
				return nil
	}
}

func SampleDist(dist Distribution, num_samples int64) []float64 {
	if dist == nil {
		return nil
	}

	samples := make([]float64, num_samples)

	for i := int64(0); i < num_samples; i++ {
		samples[i] = dist.Rand()
	}

	return samples
}

type Likelihood struct {
	Params []float64
	Data mat.Matrix
}

func (l *Likelihood) CalcLikelihood() []float64 {
	numrows, _ := l.Data.Dims()

	var likelihoods []float64

	paramvec := mat.NewVecDense(len(l.Params), l.Params)

	for i := 0; i < numrows; i++ {
		var row []float64
		row = mat.Row(row, i, l.Data)
		rowvec := mat.NewVecDense(len(row), row)

		likelihoods = append(likelihoods, mat.Dot(rowvec, paramvec))
	}
	return likelihoods
}

type Prior struct {
	distribution Distribution 
}


type Posterior struct {
	priors []Prior
	likelihood Likelihood
}


func (p *Posterior) calcPosterior() mat.Dense {
	var priorSamples [][]float64
	for _, prior := range p.priors {
		// Sample 1000 values for each prior
		samples := SampleDist(prior.distribution, 1000)
		if samples == nil {
			// Handle nil distribution if necessary
			continue
		}
		priorSamples = append(priorSamples, samples)
	}

	// Prepare the parameter grid using cartesian product
	var paramGrid [][]interface{}
	ParamSamples := make([][]interface{}, len(priorSamples))
	for i, sample := range priorSamples {
		// Convert the sample to interface{} for use with Iter
		convertedSample := make([]interface{}, len(sample))
		for j, s := range sample {
			convertedSample[j] = s
		}
		ParamSamples[i] = convertedSample
	}

	// Use the Iter function to generate the Cartesian product of all prior samples
	ch := cartesian.Iter(ParamSamples...)
	for Params := range ch {
		// Params will be a slice of interfaces representing one combination of parameters
		// Convert the interface slice back to float64
		convertedParams := make([]float64, len(Params))
		for i, p := range Params {
			convertedParams[i] = p.(float64)
		}

		// Set up the likelihood with the current combination of parameters
		likelihood := Likelihood{
			Params: convertedParams,
			Data:   p.likelihood.Data,
		}

		// Calculate the likelihood values
		likelihoodValues := likelihood.CalcLikelihood()

		// You could store the likelihood values or further process them here
		// For example, append to a slice or compute posterior directly
		// This part can be adjusted based on how you want to structure the posterior results.
		// Here, I am just appending the log sum of likelihoods for demonstration purposes
		var logSum float64
		for _, val := range likelihoodValues {
			logSum += val // Replace with log(val) if using log-likelihoods
		}
		convertedParamsInterface := make([]interface{}, len(convertedParams))
		for i, p := range convertedParams {
			convertedParamsInterface[i] = p
		}
		paramGrid = append(paramGrid, convertedParamsInterface) // Or use paramGrid as needed
	}

	// Convert the parameter grid and likelihoods to a matrix (e.g., posterior)
	// Here, we just return the parameter grid as an example.
	posteriorMatrix := mat.NewDense(len(paramGrid), len(paramGrid[0]), nil)
	for i, row := range paramGrid {
		for j, v := range row {
			posteriorMatrix.Set(i, j, v.(float64)) // Assuming posterior matrix values are float64
		}
	}

	return *posteriorMatrix
}


type Sampler struct {
	Params map[string]float64
	initialState mat.Matrix
	priors []Prior
	Likelihood Likelihood
}

type BayesianModel struct {
	priors []Prior
	likelihood Likelihood
	posterior Posterior
	sampler Sampler
}