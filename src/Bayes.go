package src

import (
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat/distuv"
	"math"
	"slices"
	"fmt"
)

// Distribution interface
type Distribution interface {
	Rand() float64
	CDF(float64) float64
}

type DistributionParams struct {
	Dist   string
	Params map[string]float64
}

type Prior struct {
	Distribution Distribution
}

type Likelihood struct {
	Params             []float64
	DistributionParams DistributionParams
	Data               mat.Matrix
}

type Posterior struct {
	Priors []Prior
	Data   mat.Matrix
	LikelihoodParams DistributionParams
}

type PosteriorResult struct {
	Params      []float64
	Probability float64
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

func (l *Likelihood) CalcLikelihood() float64 {
	numRows, numCols := l.Data.Dims()

	// Get the ordered list of parameter keys for the distribution
	paramKeys := getParamKeys(l.DistributionParams.Dist)

	fmt.Println(l.Params)

	// Map the slice values to the Params map
	for i, key := range paramKeys {
		fmt.Println("Key:", key)
		l.DistributionParams.Params[key] = l.Params[i]
	}
	fmt.Println(l.DistributionParams)
	// Create the distribution
	dist := l.DistributionParams.CreateDist()

	fmt.Println("Distribution:", dist)
	var logSum float64 = 0.0
	// Calculate log-likelihood for each row
	for i := 0; i < numRows; i++ {
		
		for j := 0; j < numCols; j++ {
			dataPoint := l.Data.At(i, j)
			fmt.Println("Data point:", dataPoint)
			cdf := dist.CDF(dataPoint) // Consider using PDF instead
			fmt.Println("CDF:", cdf)
			if cdf <= 0 {
				// Handle cases where CDF is zero or negative
				logLiklihood := math.Inf(-1)
				fmt.Println("Log likelihood:", logLiklihood)
			} else {
				logLiklihood := math.Log10(cdf)
				fmt.Println("Log likelihood:", logLiklihood)
				logSum += logLiklihood
			}	
		}
	
	}

	return logSum
}

func (p *Posterior) CalcPosterior() []PosteriorResult {
	numpriors := len(p.Priors)

	// Sample the first prior and initialize the Cartesian product
	first := SampleDist(p.Priors[0].Distribution, 50)
	var cart [][]float64
	for _, f := range first {
		cart = append(cart, []float64{f})
	}

	// Iteratively calculate Cartesian products with remaining priors
	for prior := 1; prior < numpriors; prior++ {
		sampled := SampleDist(p.Priors[prior].Distribution, 50)
		cart = Combination(cart, sampled)
	}

	fmt.Println("Cartesian product:", cart)

	// Calculate likelihoods for each combination
	likelihoods := make([]float64, len(cart))
	for i, row := range cart {
		likelihood := Likelihood{Params: row, Data: p.Data, DistributionParams: p.LikelihoodParams}
		individuals := likelihood.CalcLikelihood()
  	likelihoods[i] = individuals
	}

	fmt.Println("Likelihoods:", likelihoods)

	maxLL  := -999999999999.0
	for _, ll := range likelihoods {
		if ll > maxLL {
			maxLL = ll
		}
	}
	fmt.Println("Max Log Likelihood:", maxLL)
	for i, ll := range likelihoods {
		likelihoods[i] = math.Exp(ll - maxLL)
	}
	fmt.Println("Pre-Normalized Standardized likelihoods:", likelihoods)
	likelihoods = Normalize(likelihoods)

	fmt.Println("Normalized likelihoods:", likelihoods)
	// Collect results
	results := make([]PosteriorResult, len(cart))
	for i := range cart {
		results[i] = PosteriorResult{
			Params:      cart[i],
			Probability: likelihoods[i],
		}
	}
	return results
}



type BayesianModel struct {
	priors     []Prior
	likelihood Likelihood
	posterior  Posterior
	sampler    Sampler
}

/// Helper functions
// Map function for generic slices
type mapFunc[E any] func(E) E

func Map[S ~[]E, E any](s S, f mapFunc[E]) S {
	result := make(S, len(s))
	for i := range s {
		result[i] = f(s[i])
	}
	return result
}

func SampleDist(dist Distribution, num_samples int) []float64 {
	if dist == nil {
		return nil
	}
	samples := make([]float64, num_samples)
	for i := 0; i < num_samples; i++ {
		samples[i] = dist.Rand()
	}

	slices.Sort(samples)

	return samples
}

func getParamKeys(distType string) []string {
	switch distType {
	case "Normal":
		return []string{"Mu", "Sigma"}
	case "Bernoulli":
		return []string{"P"}
	case "Beta":
		return []string{"Alpha", "Beta"}
	case "Binomial":
		return []string{"N", "P"}
	case "ChiSquared":
		return []string{"K"}
	case "Exponential":
		return []string{"Rate"}
	case "Gamma":
		return []string{"Alpha", "Beta"}
	case "LogNormal":
		return []string{"Mu", "Sigma"}
	case "Pareto":
		return []string{"Xm", "Alpha"}
	case "Poisson":
		return []string{"Lambda"}
	case "StudentsT":
		return []string{"Mu", "Sigma", "Nu"}
	case "Uniform":
		return []string{"Min", "Max"}
	case "Weibull":
		return []string{"K", "Lambda"}
	default:
		return nil
	}
}

func Combination(current [][]float64, next []float64) [][]float64 {
	var result [][]float64
	for _, c := range current {
		for _, n := range next {
			combined := append(append([]float64{}, c...), n)
			result = append(result, combined)
		}
	}
	return result
}

func Sum(data []float64) float64 {
	var total float64
	for _, val := range data {
		total += val
	}
	return total
}

func Normalize(data []float64) []float64 {
	sum := Sum(data)
	for i, val := range data {
		data[i] = val / sum
	}
	return data
}