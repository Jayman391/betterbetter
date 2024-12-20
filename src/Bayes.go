package src
import (
	"math"
	"math/rand"
	"slices"

	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat/distuv"
	"fmt"
)

// Distribution interface
type Distribution interface {
	Rand() float64
	CDF(float64) float64
}

type ContinuousDistribution interface {
	Distribution
	Quantile(float64) float64
}

type DiscreteDistribution interface {
	Distribution
	Prob(float64) float64
}

type DistributionParams struct {
	Dist   string
	DistType string
	Params map[string]float64
}

type Prior struct {
	Distribution Distribution
}

type Likelihood struct {
	Params             []float64
	DistributionParams DistributionParams
	InputData               mat.Dense
	OutputData              mat.VecDense
	Link               func([]float64 , []float64) []float64
}

type Posterior struct {
	Priors           []DistributionParams
	Data             mat.Dense
	LikelihoodParams DistributionParams
	MarkovChain      MarkovChain
}

type PosteriorResult struct {
	Params      []float64
	LogLikelihood float64
}

type MarkovChain struct {
	Distributions []DistributionParams
	Grid          mat.Dense
	Likelihood    Likelihood
	SampleSize    int
	Sampler       string
}

func (l *Likelihood) CalcDataLikelihood() float64 {
	NegLogLikelihood := 0.0
	distType := l.DistributionParams.Dist

	for i := range l.InputData.RawMatrix().Rows {

		inputdata := make([]float64, l.InputData.RawMatrix().Cols)
		for j := range inputdata {
			inputdata[j] = l.InputData.At(i, j)
		}
		//fmt.Println(data)
		//fmt.Println(l.Params)
		selectedParams := l.Link(l.Params, inputdata)
		paramKeys := getParamKeys(distType)
		for i, key := range paramKeys {
			l.DistributionParams.Params[key] = selectedParams[i]
		}

		outputdata := make([]float64, l.OutputData.Len())

		for j := 0; j < l.OutputData.Len(); j++ {
			outputdata[j] = l.OutputData.At(j,0)
		}

		//fmt.Println(l.DistributionParams.Params)

		switch distType {
			case "Normal":
				mu := l.DistributionParams.Params["Mu"]
				sigma := l.DistributionParams.Params["Sigma"]
				NegLogLikelihood += -UVNormalLogLikelihood(mu, sigma, outputdata)
			case "Poisson":
				lambda := l.DistributionParams.Params["Lambda"]
				NegLogLikelihood += -UVPoissonLogLikelihood(lambda, outputdata)
			case "Exponential":
				rate := l.DistributionParams.Params["Rate"]
				NegLogLikelihood += -UVExponentialLogLikelihood(rate, outputdata)
			case "Uniform":
				min := l.DistributionParams.Params["Min"]
				max := l.DistributionParams.Params["Max"]
				n := float64(len(outputdata))
				NegLogLikelihood += -UVUniformLogLikelihood(min, max, n)
		}
}

	return NegLogLikelihood
}

func UVNormalLogLikelihood(mu float64, sigma float64, data []float64) float64 {
	sum := -float64(len(data)) / 2.0 * math.Log(2*math.Pi*math.Pow(sigma, 2))
	for _, d := range data {
		sum -= (1.0 / 2.0 * math.Pow(sigma, 2)) * math.Pow((d-mu), 2)
	}
	return sum
}

func UVPoissonLogLikelihood(lambda float64, data []float64) float64 {
	sum := -float64(len(data)) * lambda
	for _, d := range data {
		sum += d*math.Log(lambda) - math.Log(math.Gamma(d+1))
	}
	return sum
}

func UVExponentialLogLikelihood(rate float64, data []float64) float64 {
	sum := (float64(len(data))) * math.Log(rate)
	for _, d := range data {
		sum -= rate * d
	}
	return sum
}

func UVUniformLogLikelihood(min float64, max float64, n float64) float64 {
	return -n * math.Log(max-min)
}

func (p *Posterior) CalcPosterior(chains int) []PosteriorResult {
	// Create the grid
	p.MarkovChain.CreateGrid()

	// generate initial state for the Markov Chain (random row in grid)
	numCombos := p.MarkovChain.Grid.RawMatrix().Rows

	index := int64(math.Ceil(rand.Float64() * float64(numCombos)))

	numsteps := 5000

	likelihoods := make([]float64, numsteps+1)
	samples := make([][]float64, numsteps+1)

	switch p.MarkovChain.Sampler {
	case "Unit":
		samples, likelihoods = p.MarkovChain.UnitRandomWalk(int64(index), numsteps)
	case "Lattice":
		samples, likelihoods = p.MarkovChain.LatticeRandomWalk(int64(index), numsteps)
	case "Gaussian":
		samples, likelihoods = p.MarkovChain.GaussianRandomWalk(int64(index), numsteps)
	case "Metropolis":
		samples, likelihoods = p.MarkovChain.MetropolisHastings(int64(index), int(float64(numsteps) * 1.5), int(float64(numsteps)/2.0) )
	case "Hamiltonian":
		samples, likelihoods = p.MarkovChain.HamiltonianMonteCarlo(int64(index), numsteps)
	}

	// take index, get prior params. take CDF of prior params, multiply by likelihood
	// for prior in priors
	// take cdf of value at index given prior distribution
	// multiply by likelihood
	for i, sample := range samples {
		NegLogLikelihood := 0.0
		data := sample
		index := 0
		for _, prior := range p.Priors {
			distType := prior.Dist
			switch distType {
			case "Normal":
				mu := prior.Params["Mu"]
				sigma := prior.Params["Sigma"]
				NegLogLikelihood -= UVNormalLogLikelihood(mu, sigma, []float64{data[index]})
				index += 1
			case "Poisson":
				lambda := prior.Params["Lambda"]
				NegLogLikelihood -= UVPoissonLogLikelihood(lambda, []float64{data[index]})
				index += 1
			case "Exponential":
				rate := prior.Params["Rate"]
				NegLogLikelihood -= UVExponentialLogLikelihood(rate, []float64{data[index]})
				index += 1
			case "Uniform":
				min := prior.Params["Min"]
				max := prior.Params["Max"]
				NegLogLikelihood -= UVUniformLogLikelihood(min, max, float64(len(samples)))
				index += 1
			}
		}
		likelihoods[i] += NegLogLikelihood
	}

	// remove the first element of the likelihoods array and indices array
	samples = samples[1:]
	likelihoods = likelihoods[1:]

	// Collect results
	results := make([]PosteriorResult, len(samples))
	for i, sample := range samples {
		results[i] = PosteriorResult{
			Params:      sample,
			LogLikelihood: likelihoods[i],
		}
	}
	//fmt.Println(results)
	return results
}

func (p *Posterior) CalcPosteriorPredictive(results []PosteriorResult, data [][]float64, numsamples int, linkfunc func([]float64, []float64) []float64) []float64 {
	//weights correspond to likelihoods
	// for num samples, weighted randomly select a result based on likelihood
	// for each result, sample a data point from the likelihood distribution using the result's parameters as priors
	// return the samples
	weights := make([]float64, len(results))
	posteriorIndices := make([]int64, len(results))

	for i, result := range results {
		posteriorIndices[i] = int64(i)
		weights[i] = result.LogLikelihood   
	}
	

	//get minumum weight
	maxLL := weights[0]
	maxLLIndex := 0

	for index, weight := range weights {
		if weight > maxLL {
			maxLL = weight
			maxLLIndex = index
		}
	}

	//these params are from Markov Chain Grid, must be transformed to rate variable
	maxLLParams := results[maxLLIndex].Params
	tempparams := make([]float64, len(maxLLParams))

	//fmt.Println(data)

	for _, d := range data {


		linkedParams := linkfunc(maxLLParams, d)

		for i, j := range linkedParams {
			tempparams[i] += (j / float64(len(data)))
		}
	}
	maxLLParams = tempparams
	

	maxLLDistParams := p.LikelihoodParams.Params
	
	LLKeys := getParamKeys(p.LikelihoodParams.Dist)

	for i, key := range LLKeys {
		maxLLDistParams[key] = maxLLParams[i]
	}

	MaxLLDist := p.LikelihoodParams.CreateDist() 

	fmt.Println(maxLLDistParams)

	posteriorSamples := SampleDist(MaxLLDist, numsamples)

	return posteriorSamples
}

func (m *MarkovChain) UnitRandomWalk(index int64, numsteps int) ([][]float64, []float64) {
	indices := make([]int64, numsteps+1)
	likelihoods := make([]float64, numsteps+1)
	samples := make([][]float64, numsteps+1)
	// find closest point in grid to initial conditions

	startingIndex := index

	point := m.Grid.RawRowView(int(index))
	m.Likelihood.Params = point // Fix: Assign the value to a slice of length 1

	indices[0] = index
	samples[0] = point
	likelihoods[0] = m.Likelihood.CalcDataLikelihood()

	for i := 0; i < numsteps; i++ {
		// find neighbors of closest point
		neighbors := m.GetNeighbors(index)
		// randomly select neighbor
		index = neighbors[rand.Intn(len(neighbors))]

		indices[i+1] = index

		if index < 0 || index >= int64(m.Grid.RawMatrix().Rows) {
			index = startingIndex
		}

		point := m.Grid.RawRowView(int(index))
		m.Likelihood.Params = point
		likelihoods[i+1] = m.Likelihood.CalcDataLikelihood()
		samples[i+1] = point
	}

	return samples, likelihoods
}

func (m *MarkovChain) LatticeRandomWalk(index int64, numsteps int) ([][]float64, []float64) {
	// find closest point in grid to initial conditions
	// find neighbors of closest point
	// calculate likelihood of each neighbor
	// normalize likelihoods from just the neighbors
	// randomly select neighbor based on normalized likelihood
	indices := make([]int64, numsteps+1)
	likelihoods := make([]float64, numsteps+1)
	samples := make([][]float64, numsteps+1)
	// find closest point in grid to initial conditions

	startingIndex := index

	point := m.Grid.RawRowView(int(index))
	m.Likelihood.Params = point // Fix: Assign the value to a slice of length 1

	indices[0] = index
	samples[0] = point
	likelihoods[0] = m.Likelihood.CalcDataLikelihood()

	for i := 0; i < numsteps; i++ {
		// find neighbors of closest point
		neighbors := m.GetNeighbors(index)
		// calculate likelihood of each neighbor
		neighborLikelihoods := make([]float64, len(neighbors))
		for j, neighbor := range neighbors {
			if neighbor < 0 {
				neighbor = int64(m.Grid.RawMatrix().Rows) + neighbor
			}
			point := m.Grid.RawRowView(int(neighbor))
			m.Likelihood.Params = point // Fix: Assign the value to a slice of length 1
			neighborLikelihoods[j] = m.Likelihood.CalcDataLikelihood()
		}
		// normalize likelihoods from just the neighbors
		neighborLikelihoods = Normalize(neighborLikelihoods)
		// randomly select neighbor based on normalized likelihood
		slices.Sort(neighborLikelihoods)
		rand := rand.Float64()
		neighborIndex := 0
		neighborLL := 0.0
		minDiff := 1.0
		for j, likelihood := range neighborLikelihoods {
			diff := math.Abs(likelihood - rand)
			if diff < minDiff {
				minDiff = diff
				neighborIndex = j
				neighborLL = likelihood
			}
		}

		index = neighbors[neighborIndex]

		indices[i+1] = index

		if index < 0 || index >= int64(m.Grid.RawMatrix().Rows) {
			index = startingIndex
		}

		point := m.Grid.RawRowView(int(index))
		m.Likelihood.Params = point
		likelihoods[i+1] = neighborLL
		samples[i+1] = point

	}

	return samples, likelihoods
}

func (m *MarkovChain) GaussianRandomWalk(index int64, numsteps int) ([][]float64, []float64) {
	// make normal distributions with mu and stdev for each dimension

	indices := make([]int64, numsteps+1)
	likelihoods := make([]float64, numsteps+1)
	fullsamples := make([][]float64, numsteps+1)

	point := m.Grid.RawRowView(int(index))
	m.Likelihood.Params = point // Fix: Assign the value to a slice of length 1

	indices[0] = index
	fullsamples[0] = point
	likelihoods[0] = m.Likelihood.CalcDataLikelihood()

	distparams := DistributionParams{
		Dist: "Normal",
		Params: map[string]float64{
			"Mu":    0.0,
			"Sigma": 1.0,
		},
	}

	dist := distparams.CreateDist()

	for i := 0; i < numsteps; i++ {
		var samples []int64
		sample := SampleDist(dist, 1)
		cdf := dist.CDF(sample[0])
		steps := math.Round(cdf * float64(m.SampleSize))

		samples = append(samples, (int64(steps)+index)%int64(m.Grid.RawMatrix().Rows))
		samples = append(samples, (index-int64(steps))%int64(m.Grid.RawMatrix().Rows))

		for j := 1; j < len(m.Distributions); j++ {
			sample := SampleDist(dist, 1)
			cdf := dist.CDF(sample[0])
			steps := math.Round(cdf * float64(m.SampleSize) * float64(j))

			samples = append(samples, (int64(steps)+index)%int64(m.Grid.RawMatrix().Rows))
		}

		// calculate likelihood of each neighbor sample
		neighborLikelihoods := make([]float64, len(samples))
		for j, neighbor := range samples {
			if neighbor < 0 {
				neighbor = int64(m.Grid.RawMatrix().Rows) + neighbor
			}
			point := m.Grid.RawRowView(int(neighbor))
			m.Likelihood.Params = point // Fix: Assign the value to a slice of length 1
			neighborLikelihoods[j] = m.Likelihood.CalcDataLikelihood()
		}

		// normalize likelihoods from just the neighbors
		neighborLikelihoods = Normalize(neighborLikelihoods)

		// randomly select neighbor based on normalized likelihood
		slices.Sort(neighborLikelihoods)
		rand := rand.Float64()
		neighborIndex := 0
		neighborLL := 0.0
		minDiff := 1.0
		for j, likelihood := range neighborLikelihoods {
			diff := math.Abs(likelihood - rand)
			if diff < minDiff {
				minDiff = diff
				neighborIndex = j
				neighborLL = likelihood
			}
		}

		index = samples[neighborIndex]

		indices[i+1] = index
		fullsamples[i+1] = m.Grid.RawRowView(int(math.Abs(float64(index))+1) % m.Grid.RawMatrix().Rows)
		likelihoods[i+1] = neighborLL
	}

	return fullsamples, likelihoods
}

func (m *MarkovChain) MetropolisHastings(index int64, numsteps int, burnin int) ([][]float64, []float64) {
	// Validate burn-in
	if burnin >= numsteps {
		panic("Burn-in period must be less than the total number of steps")
	}

	indices := make([]int64, numsteps+1)
	likelihoods := make([]float64, numsteps+1)

	// Initialize
	currentIndex := index
	point := m.Grid.RawRowView(int(currentIndex))
	m.Likelihood.Params = point
	currentLikelihood := m.Likelihood.CalcDataLikelihood()

	indices[0] = currentIndex
	likelihoods[0] = currentLikelihood

	samples := make([][]float64, numsteps+1)
	samples[0] = make([]float64, len(point))
	copy(samples[0], point)

	// Metropolis-Hastings Sampling
	for i := 1; i <= numsteps; i++ {
		// Propose a new index (neighbor)
		neighbors := m.GetNeighbors(currentIndex)
		numNeighbors := len(neighbors)

		// Randomly select a neighbor
		proposedIndex := neighbors[rand.Intn(numNeighbors)]

		// Handle index wrapping
		if proposedIndex < 0 {
			proposedIndex += int64(m.Grid.RawMatrix().Rows)
		} else if proposedIndex >= int64(m.Grid.RawMatrix().Rows) {
			proposedIndex -= int64(m.Grid.RawMatrix().Rows)
		}

		// Compute likelihood of proposed state
		proposedPoint := m.Grid.RawRowView(int(proposedIndex))
		m.Likelihood.Params = proposedPoint
		proposedLikelihood := m.Likelihood.CalcDataLikelihood()

		// Acceptance probability
		acceptanceProb := math.Exp(currentLikelihood - proposedLikelihood)
		if acceptanceProb > 1 {
			acceptanceProb = 1
		}

		accepted := rand.Float64() < acceptanceProb

		if accepted {
			// Accept the proposed state
			samples[i] = make([]float64, len(proposedPoint))
			copy(samples[i], proposedPoint)
			likelihoods[i] = proposedLikelihood
			currentIndex = proposedIndex
			currentLikelihood = proposedLikelihood
		} else {
			// Reject the proposed state
			samples[i] = make([]float64, len(samples[i-1]))
			copy(samples[i], samples[i-1])
			likelihoods[i] = currentLikelihood
		}
	}

	// Discard burn-in samples
	return samples[burnin:], likelihoods[burnin:]
}


func (m *MarkovChain) HamiltonianMonteCarlo(index int64, numSamples int) ([][]float64, []float64) {
	indices := make([]int64, numSamples+1)
	likelihoods := make([]float64, numSamples+1)

	// Initialize
	currentIndex := index
	point := m.Grid.RawRowView(int(currentIndex))
	dimension := len(point)

	// Negative log probability function
	negativeLogProb := func(q []float64) float64 {
		m.Likelihood.Params = q
		return m.Likelihood.CalcDataLikelihood()
	}

	// Gradient of the negative log probability
	dVdq := func(q []float64) []float64 {
		return gradient(negativeLogProb, q)
	}

	normalDist := distuv.Normal{
		Mu:    0,
		Sigma: 1,
	}

	// Start sampling
	samples := make([][]float64, numSamples+1)
	samples[0] = make([]float64, len(point))
	copy(samples[0], point)

	indices[0] = currentIndex
	likelihoods[0] = negativeLogProb(point)

	// Initialize step size

	stepSize := 0.0001

	p0 := make([]float64, dimension)
	for i := range p0 {
		p0[i] = normalDist.Rand()
	}

	leapfrogSteps := 25
	for i := 1; i <= numSamples; i++ {
		qCurrent := samples[i-1]
		p0 := make([]float64, dimension)
		for j := range p0 {
			p0[j] = normalDist.Rand()
		}

		qNew, pNew := leapfrog(qCurrent, p0, dVdq, stepSize, leapfrogSteps)

		// Compute Hamiltonian for current and new positions
		startHamiltonian := hamiltonian(qCurrent, p0, negativeLogProb)
		newHamiltonian := hamiltonian(qNew, pNew, negativeLogProb)

		// Acceptance probability (correct)
		acceptanceProb := math.Exp(startHamiltonian - newHamiltonian)
		if acceptanceProb > 1 {
			acceptanceProb = 1
		}
		accepted := rand.Float64() < acceptanceProb

		if accepted {
			// Accept the new sample
			samples[i] = qNew
			likelihoods[i] = negativeLogProb(qNew)
		} else {
			// Reject the new sample; use the current sample again
			samples[i] = qCurrent
			likelihoods[i] = negativeLogProb(qCurrent)
		}

	}

	return samples, likelihoods
}

func (m *MarkovChain) CreateGrid() {
	// Create grid by sampling directly from each distribution
	DistSamples := make([][]float64, len(m.Distributions))

	for i, distParams := range m.Distributions {
		dist := distParams.CreateDist()
		DistSamples[i] = make([]float64, m.SampleSize)

		// Generate samples uniformly from the distribution
		for j := 0; j < m.SampleSize; j++ {
			DistSamples[i][j] = dist.Rand()
		}
		slices.Sort(DistSamples[i])
	}

	// Create a Cartesian product of the samples to form the grid
	numPoints := math.Pow(float64(m.SampleSize), float64(len(m.Distributions)))
	grid := mat.NewDense(int(numPoints), len(m.Distributions), nil)

	// Combine samples from each dimension
	DistComb := [][]float64{make([]float64, 0)}
	for _, dist := range DistSamples {
		DistComb = Combination(DistComb, dist)
	}


	// Fill the grid with the combined points
	for i, row := range DistComb {
		for j, val := range row {
			grid.Set(i, j, val)
		}
	}


	m.Grid = *grid
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
func (m *MarkovChain) GetNeighbors(index int64) []int64 {
	neighbors := []int64{}

	// For each dimension, consider moving +1 or -1 step
	for dim := 0; dim < m.Grid.RawMatrix().Cols; dim++ {
		step := int64(math.Pow(float64(m.SampleSize), float64(dim)))

		// Move forward in this dimension
		forward := (index + step) % int64(m.Grid.RawMatrix().Rows)
		neighbors = append(neighbors, forward)

		// Move backward in this dimension
		backward := (index - step + int64(m.Grid.RawMatrix().Rows)) % int64(m.Grid.RawMatrix().Rows)
		neighbors = append(neighbors, backward)
	}

	return neighbors
}
func ClosestPoint(grid mat.Dense, point []float64) int64 {
	// turn point to Vector
	pVec := mat.NewVecDense(len(point), point)

	dots := make([]float64, grid.RawMatrix().Rows)

	for i := 0; i < grid.RawMatrix().Rows; i++ {
		rVec := grid.RowView(i)
		dots[i] = mat.Dot(pVec, rVec)
	}

	max := dots[0]
	index := 0
	for i := 1; i < len(dots); i++ {
		if dots[i] > max {
			max = dots[i]
			index = i
		}
	}

	return int64(index)
}
func RemoveDuplicates(indices []int64, likelihoods []float64) ([]int64, []float64) {
	seen := make(map[int64]bool)
	var newIndices []int64
	var newLikelihoods []float64

	for i, index := range indices {
		if !seen[index] {
			seen[index] = true
			newIndices = append(newIndices, index)
			newLikelihoods = append(newLikelihoods, likelihoods[i])
		}
	}

	return newIndices, newLikelihoods
}

func gradient(f func([]float64) float64, x []float64) []float64 {
	grad := make([]float64, len(x))
	epsilon := 1e-5
	fx := f(x)

	for i := range x {
		xForward := make([]float64, len(x))
		copy(xForward, x)
		xForward[i] += epsilon
		fxForward := f(xForward)
		grad[i] = (fxForward - fx) / epsilon
	}
	return grad
}
func leapfrog(q []float64, p []float64, dVdq func([]float64) []float64, stepSize float64, leapfrogSteps int) ([]float64, []float64) {
	qNew := make([]float64, len(q))
	pNew := make([]float64, len(p))
	copy(qNew, q)
	copy(pNew, p)

	// Half step for momentum
	gradV := dVdq(qNew)

	for i := 0; i < leapfrogSteps; i++ {
		for i := range pNew {
			pNew[i] -= stepSize * gradV[i] / 2.0
		}
		// Full step for position
		for j := range qNew {
			qNew[j] += stepSize * pNew[j]
		}
		// Full step for momentum, except at the end of trajectory
		gradV = dVdq(qNew)
		for j := range pNew {
			pNew[j] -= stepSize * gradV[j] / 2.0
		}
	}

	return qNew, pNew
}
func hamiltonian(q []float64, p []float64, negativeLogProb func([]float64) float64) float64 {
	kineticEnergy := 0.0
	// create vector p and dot it with transpose
	pvec := mat.NewVecDense(len(p), p)
	kineticEnergy = mat.Dot(pvec, pvec) / 2.0
	potentialEnergy := negativeLogProb(q)
	return potentialEnergy + kineticEnergy
}
func (l *Likelihood) NegativeLogProb(q []float64) float64 {
	// Update l.Params with q
	copy(l.Params, q)
	return l.CalcDataLikelihood() // Assuming CalcDataLikelihood returns negative log-likelihood
}

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
