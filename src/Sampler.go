package src

import (
	//"fmt"
	"math"
	"math/rand"
	"slices"

	"gonum.org/v1/gonum/mat"
)

type MarkovChain struct {
	Distributions []Distribution
	Grid 				mat.Dense
	Likelihood Likelihood
	SampleSize 	int
	Sampler string
}

func (m *MarkovChain) UnitRandomWalk(index int64, numsteps int) ([]int64, []float64) {
	indices := make([]int64, numsteps + 1)
	likelihoods := make([]float64, numsteps + 1)
	// find closest point in grid to initial conditions

	startingIndex := index


	point := m.Grid.RawRowView(int(index))
	m.Likelihood.Params = point  // Fix: Assign the value to a slice of length 1

	indices[0] = index
	likelihoods[0] = m.Likelihood.CalcLikelihood()

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
		likelihoods[i+1] = m.Likelihood.CalcLikelihood()
	}

	indices, likelihoods = RemoveDuplicates(indices, likelihoods)

	return indices, likelihoods

}

func (m *MarkovChain) LatticeRandomWalk(index int64, numsteps int) ([]int64, []float64) {
	// find closest point in grid to initial conditions
	// find neighbors of closest point
	// calculate likelihood of each neighbor
	// normalize likelihoods from just the neighbors
	// randomly select neighbor based on normalized likelihood
	indices := make([]int64, numsteps + 1)
	likelihoods := make([]float64, numsteps + 1)
	// find closest point in grid to initial conditions

	startingIndex := index


	point := m.Grid.RawRowView(int(index))
	m.Likelihood.Params = point  // Fix: Assign the value to a slice of length 1

	indices[0] = index
	likelihoods[0] = m.Likelihood.CalcLikelihood()

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
			neighborLikelihoods[j] = m.Likelihood.CalcLikelihood()
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

	}

	indices, likelihoods = RemoveDuplicates(indices, likelihoods)

	return indices, likelihoods
 
}

func (m *MarkovChain) GaussianRandomWalk(index int64, numsteps int) ([]int64, []float64) {
	// make normal distributions with mu and stdev for each dimension

	indices := make([]int64, numsteps + 1)
	likelihoods := make([]float64, numsteps + 1)

	point := m.Grid.RawRowView(int(index))
	m.Likelihood.Params = point  // Fix: Assign the value to a slice of length 1

	indices[0] = index
	likelihoods[0] = m.Likelihood.CalcLikelihood()


	distparams := DistributionParams{
		Dist : "Normal",
		Params : map[string]float64{
			"Mu" : 0.0,
			"Sigma" : 1.0,
		},
	}

	dist := distparams.CreateDist()

	for i := 0; i < numsteps; i++ {
		var samples []int64
		sample := SampleDist(dist, 1)
		cdf := dist.CDF(sample[0])
		steps := math.Round(cdf * float64(m.SampleSize)) 
		
	  samples = append(samples, (int64(steps) + index) % int64(m.Grid.RawMatrix().Rows))
		samples = append(samples, (index - int64(steps)) % int64(m.Grid.RawMatrix().Rows))

		for j := 1; j < len(m.Distributions); j++ {
			sample := SampleDist(dist, 1)
			cdf := dist.CDF(sample[0])
			steps := math.Round(cdf * float64(m.SampleSize) * float64(j))

			samples = append(samples, (int64(steps) + index) % int64(m.Grid.RawMatrix().Rows))
		}

		// calculate likelihood of each neighbor sample
		neighborLikelihoods := make([]float64, len(samples))
		for j, neighbor := range samples {
			if neighbor < 0 {
				neighbor = int64(m.Grid.RawMatrix().Rows) + neighbor
			}
			point := m.Grid.RawRowView(int(neighbor))
			m.Likelihood.Params = point // Fix: Assign the value to a slice of length 1
			neighborLikelihoods[j] = m.Likelihood.CalcLikelihood()			
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
		likelihoods[i+1] = neighborLL
	}

	indices, likelihoods = RemoveDuplicates(indices, likelihoods)

	return indices, likelihoods

}

func (m *MarkovChain) MetropolisHastings(index int64, numsteps int) ([]int64, []float64) {
	indices := make([]int64, numsteps+1)
	likelihoods := make([]float64, numsteps+1)

	// Initialize
	currentIndex := index
	point := m.Grid.RawRowView(int(currentIndex))
	m.Likelihood.Params = point
	currentLikelihood := m.Likelihood.CalcLikelihood()

	indices[0] = currentIndex
	likelihoods[0] = currentLikelihood

	// Metropolis-Hastings Sampling
	for i := 1; i <= numsteps; i++ {
			// Propose a new index (neighbor)
			neighbors := m.GetNeighbors(currentIndex)
			numNeighbors := len(neighbors)

			// Compute likelihoods of neighbors
			neighborLikelihoods := make([]float64, numNeighbors)
			totalLikelihood := 0.0
			for j, neighborIndex := range neighbors {
					if neighborIndex < 0 {
							neighborIndex += int64(m.Grid.RawMatrix().Rows)
					} else if neighborIndex >= int64(m.Grid.RawMatrix().Rows) {
							neighborIndex -= int64(m.Grid.RawMatrix().Rows)
					}
					neighborPoint := m.Grid.RawRowView(int(neighborIndex))
					m.Likelihood.Params = neighborPoint
					likelihood := math.Exp(-m.Likelihood.CalcLikelihood())
					neighborLikelihoods[j] = likelihood
					totalLikelihood += likelihood
			}

        // Normalize neighbor likelihoods to create a probability distribution
        for j := range neighborLikelihoods {
            neighborLikelihoods[j] /= totalLikelihood
        }

        // Sample a neighbor index weighted by likelihoods
        proposedIndex := weightedSample(neighbors, neighborLikelihoods)

			// Handle index wrapping
			if proposedIndex < 0 {
					proposedIndex += int64(m.Grid.RawMatrix().Rows)
			} else if proposedIndex >= int64(m.Grid.RawMatrix().Rows) {
					proposedIndex -= int64(m.Grid.RawMatrix().Rows)
			}

			// Compute likelihood of proposed state
			proposedPoint := m.Grid.RawRowView(int(proposedIndex))
			m.Likelihood.Params = proposedPoint
			proposedLikelihood := m.Likelihood.CalcLikelihood()

			// Acceptance probability
			acceptanceRatio := proposedLikelihood / currentLikelihood

			// Accept or reject the proposal
			if rand.Float64() < acceptanceRatio {
					// Accept the proposed state
					currentIndex = proposedIndex
					currentLikelihood = proposedLikelihood
			}
			// Record the current state
			indices[i] = currentIndex
			likelihoods[i] = currentLikelihood
	}

	indices, likelihoods = RemoveDuplicates(indices, likelihoods)

	return indices, likelihoods
}


func (m *MarkovChain) HMC(initialConditions []float64, numsteps int, timestep float64) {
	// https://faculty.washington.edu/yenchic/19A_stat535/Lec9_HMC.pdf
	// find closest point in grid to initial conditions
	// draw a random momentum vector from a multivariate normal distribution
	// Apply Hamiltonian dynamics to the point and momentum vector to calculate derivatives
		// Potential Energy H = -log(likelihood) - log(momentum vector)
		// Kinetic Energy P 
	// Use leapfrog integration to update the point and momentum vector for numsteps with timestep as the step size
	// Calculate the acceptance ratio of the new point compared to the old point like in Metropolis-Hastings

}


func (m *MarkovChain) CreateGrid()  {
	first := SampleDist(m.Distributions[0], m.SampleSize)
	var cart [][]float64
	for _, f := range first {
		cart = append(cart, []float64{f})
	}

	// Iteratively calculate Cartesian products with remaining priors
	for prior := 1; prior < len(m.Distributions); prior++ {
		sampled := SampleDist(m.Distributions[prior], m.SampleSize)
		cart = Combination(cart, sampled)
	}

	matrix := mat.NewDense(len(cart), len(cart[0]), nil)
	for i, row := range cart {
		for j, val := range row {
			matrix.Set(i, j, val)
		}
	}

	m.Grid = *matrix

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

func weightedSample(indices []int64, weights []float64) int64 {
	cumulativeWeights := make([]float64, len(weights))
	cumulativeWeights[0] = weights[0]
	for i := 1; i < len(weights); i++ {
			cumulativeWeights[i] = cumulativeWeights[i-1] + weights[i]
	}
	r := rand.Float64()
	for i, cw := range cumulativeWeights {
			if r < cw {
					return indices[i]
			}
	}
	return indices[len(indices)-1] // Return the last index if not found earlier
}

