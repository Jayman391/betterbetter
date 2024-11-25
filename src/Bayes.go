package src

import (
	"math"
	"math/rand"
	"slices"

	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat/distuv"
	//"fmt"
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
	Priors           []Prior
	Data             mat.Matrix
	LikelihoodParams DistributionParams
	MarkovChain      MarkovChain
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

	// Map the slice values to the Params map
	for i, key := range paramKeys {
		l.DistributionParams.Params[key] = l.Params[i]
	}
	// Create the distribution
	dist := l.DistributionParams.CreateDist()

	var logSum float64 = 0.0
	// Calculate log-likelihood for each row
	for i := 0; i < numRows; i++ {

		for j := 0; j < numCols; j++ {
			dataPoint := l.Data.At(i, j)
			cdf := dist.CDF(dataPoint) // Consider using PDF instead
			if cdf > 0 {
				logLiklihood := math.Log10(cdf)
				logSum += logLiklihood
			}
		}

	}

	return -logSum
}

func (p *Posterior) CalcPosterior() []PosteriorResult {
	// Create the grid
	p.MarkovChain.CreateGrid()

	// generate initial state for the Markov Chain (random row in grid)
	numCombos := p.MarkovChain.Grid.RawMatrix().Rows

	index := int64(math.Ceil(rand.Float64() * float64(numCombos)))

	numsteps := 200

	likelihoods := make([]float64, numsteps+1)
	indices := make([]int64, numsteps+1)

	switch p.MarkovChain.Sampler {
	case "UnitRandomWalk":
		indices, likelihoods = p.MarkovChain.UnitRandomWalk(int64(index), numsteps)
	case "LatticeRandomWalk":
		indices, likelihoods = p.MarkovChain.LatticeRandomWalk(int64(index), numsteps)
	case "GaussianRandomWalk":
		indices, likelihoods = p.MarkovChain.GaussianRandomWalk(int64(index), numsteps)
	}

	// take index, get prior params. take CDF of prior params, multiply by likelihood
	// for prior in priors
	// take cdf of value at index given prior distribution
	// multiply by likelihood
	for i, indice := range indices {
		if indice < 0 {
			indices[i] = index
		}

		row := p.MarkovChain.Grid.RawRowView(int(indices[i]))
		priorprob := 1.0
		for i, prior := range p.Priors {
			dist := prior.Distribution
			cdf := dist.CDF(row[i])
			priorprob *= cdf
		}
		likelihoods[i] *= priorprob
	}

	// remove the first element of the likelihoods array and indices array
	likelihoods = likelihoods[1:]
	indices = indices[1:]

	indices, likelihoods = RemoveDuplicates(indices, likelihoods)

	likelihoods = Normalize(likelihoods)

	// Collect results
	results := make([]PosteriorResult, len(indices))
	for i := range indices {
		results[i] = PosteriorResult{
			Params:      p.MarkovChain.Grid.RawRowView(int(indices[i])),
			Probability: likelihoods[i],
		}
	}
	return results
}

type BayesianModel struct {
	priors     []Prior
	likelihood Likelihood
	posterior  Posterior
}

// / Helper functions
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
