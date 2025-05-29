// Package models contains model structures required for the simulation
package models

// TeamStats holds a team's statistical data
type TeamStats struct {
	TeamID          uint    `json:"team_id"`
	AvgScored       float64 `json:"avg_scored"`       // Average goals scored per match
	AvgConceded     float64 `json:"avg_conceded"`     // Average goals conceded per match
	AttackStrength  float64 `json:"attack_strength"`  // Attack strength (λ_scored / λ_league)
	DefenseStrength float64 `json:"defense_strength"` // Defense strength (λ_conceded / λ_league)
}

// Constants for simulation
const (
	LeagueAverage = 1.5 // League-wide average number of goals (λ_league)
	MinLambda     = 0.1 // Minimum value for lambda
)
