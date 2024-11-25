package src

import (
	//"fmt"
	"math/rand"

	"gonum.org/v1/gonum/mat"
)



type MarkovChain struct {
	Distributions []Distribution
	Grid 				mat.Dense
	Likelihood Likelihood
	SampleSize 	int
	Sampler string
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

func  (m *MarkovChain) GetNeighbors(point int64) []int64 {
	// Take index of row 
		// starting with the last column, neighbors are index + 1 and index - 1
			// for each next row, neighbors are index + 1 + Samplesize * x and index - 1 - Samplesize * x
		
	neighbors := []int64{point+1, point-1}

	for i := 1; i < m.Grid.RawMatrix().Cols; i++ {
		neighbors = append(neighbors, point+1+int64(m.SampleSize)*int64(i))
		neighbors = append(neighbors, point-1-int64(m.SampleSize)*int64(i))
	}

	return neighbors

}

func (m *MarkovChain) UnitRandomWalk(index int64, numsteps int) ([]int64, []float64) {
	indices := make([]int64, numsteps + 1)
	likelihoods := make([]float64, numsteps + 1)
	// find closest point in grid to initial conditions

	startingIndex := index


	point := m.Grid.RawRowView(int(index))
	m.Likelihood.Params = point  // Fix: Assign the value to a slice of length 1
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
		m.Likelihood.Params = point // Fix: Assign the value to a slice of length 1
		likelihoods[i+1] = m.Likelihood.CalcLikelihood()
	}

	indices, likelihoods = RemoveDuplicates(indices, likelihoods)

	return indices, likelihoods

}

func (m *MarkovChain) LatticeRandomWalk(initialConditions []float64, numsteps int) {
	// find closest point in grid to initial conditions
	// find neighbors of closest point
	// calculate likelihood of each neighbor
	// normalize likelihoods from just the neighbors
	// randomly select neighbor based on normalized likelihood 
}

func (m *MarkovChain) GaussianRandomWalk(initialConditions []float64, numsteps int, mu []float64, stdev []float64) {
	// make normal distributions with mu and stdev for each dimension
	// generate random numbers between 0 and 1 for each dimension
	// find closest point in grid to initial conditions
	// take Inverse CDF of distribution for each dimension, round to nearest integer, move that many steps in that direction for each dimension
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
