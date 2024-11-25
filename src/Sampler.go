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

func (m *MarkovChain) MetropolisHastings(initialConditions []float64, numsteps int) {
	// find closest point in grid to initial conditions
	// calculate likelihood of initial point
	// find neighbors of closest point
	// calculate likelihoods of each neighbor
	// normalize likelihoods from just the neighbors
	// randomly select neighbor based on normalized likelihood
	// calculate acceptance ratio of new point compared to old point
	// take new likelihood divided by old likelihood
	// generate random number between 0 and 1
	// if random number is less than acceptance ratio, accept new point
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


func  (m *MarkovChain) GetNeighbors(point int64) []int64 {
	// Take index of row 
		// starting with the last column, neighbors are index + 1 and index - 1
			// for each next row, neighbors are index + 1 + Samplesize * x and index - 1 - Samplesize * x
		
	neighbors := []int64{point+1, point-1}

	for i := 1; i < m.Grid.RawMatrix().Cols; i++ {
		neighbors = append(neighbors, (point+1+int64(m.SampleSize)*int64(i))%int64(m.Grid.RawMatrix().Rows))
		neighbors = append(neighbors, (point-1-int64(m.SampleSize)*int64(i))%int64(m.Grid.RawMatrix().Rows))
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

	// if there are duplicate indices, remove the repeated indices and likelihoods
	for i := 0; i < len(indices); i++ {
		for j := i + 1; j < len(indices); j++ {
			if indices[i] == indices[j] {
				// remove jth element from indices and likelihoods
				indices = append(indices[:j], indices[j+1:]...)
				likelihoods = append(likelihoods[:j], likelihoods[j+1:]...)
			}
		}
	}

	return indices, likelihoods

}