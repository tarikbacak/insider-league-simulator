// Package base contains the core simulator interfaces and structures
package base

// Simulator is the core interface for match simulation
type Simulator interface {
	SimulateMatch(homeTeamID, awayTeamID uint) (homeGoals, awayGoals int, err error)
	PlayNextWeek() error
	PlayAllRemainingWeeks() error
}

// Predictor is the core interface for championship prediction
type Predictor interface {
	PredictChampionshipProbabilities(week uint) (map[uint]float64, error)
}
