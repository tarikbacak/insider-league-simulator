// Package simulator provides main simulator interface implementations and exports subpackages.
package simulator

import (
	"github.com/tarikbacak/insider-league-simulator/internal/simulator/base"
	"github.com/tarikbacak/insider-league-simulator/internal/simulator/montecarlo"
	"github.com/tarikbacak/insider-league-simulator/internal/simulator/poisson"
)

// GetPoissonSimulator returns a new Poisson-based match simulator
func GetPoissonSimulator() base.Simulator {
	return poisson.NewPoissonSimulator()
}

// GetMonteCarloPredictor returns a new Monte Carlo championship predictor
func GetMonteCarloPredictor(iterations int) base.Predictor {
	return montecarlo.NewMonteCarloPredictor(iterations)
}
