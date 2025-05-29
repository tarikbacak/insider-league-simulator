package api

import "github.com/tarikbacak/insider-league-simulator/internal/models"

// ErrorResponse represents a generic error response.
type ErrorResponse struct {
	Error     string `json:"error" example:"Error message"`
	Detail    string `json:"detail,omitempty" example:"Detailed error information"`
	Timestamp string `json:"timestamp,omitempty" example:"2023-10-27 10:00:00"`
}

// APIInfoResponse defines the structure for the homepage API information.
type APIInfoResponse struct {
	Message   string            `json:"message" example:"Insider League Simulator API"`
	Version   string            `json:"version" example:"1.0.0"`
	Endpoints map[string]string `json:"endpoints"`
}

// HealthCheckResponse defines the structure for the health check endpoint.
type HealthCheckResponse struct {
	Status    string               `json:"status" example:"healthy"`
	Service   string               `json:"service" example:"insider-league-simulator"`
	Timestamp HealthCheckTimestamp `json:"timestamp"`
}

// HealthCheckTimestamp holds nested timestamp info for health check.
type HealthCheckTimestamp struct {
	Unix HealthCheckUnixTimestamp `json:"unix"`
}

// HealthCheckUnixTimestamp holds nested unix timestamp info.
type HealthCheckUnixTimestamp struct {
	Seconds HealthCheckSecondsValue `json:"seconds"`
}

// HealthCheckSecondsValue holds the actual timestamp string.
type HealthCheckSecondsValue struct {
	Value string `json:"value" example:"2023-10-27 10:00:00"`
}

// InitResponse defines the structure for the database initialization endpoint.
type InitResponse struct {
	Message   string `json:"message" example:"Database initialized successfully"`
	Note      string `json:"note,omitempty" example:"New random fixture generated"`
	Timestamp string `json:"timestamp" example:"2023-10-27 10:00:00"`
}

// StandingsResponse wraps the list of standings and total count.
type StandingsResponse struct {
	Standings  []models.Standing `json:"standings"`
	TotalTeams int               `json:"total_teams" example:"4"`
}

// MatchDetailResponse defines the structure for individual match details in a list.
type MatchDetailResponse struct {
	ID         uint    `json:"id" example:"1"`
	Week       uint    `json:"week" example:"1"`
	HomeTeamID uint    `json:"home_team_id" example:"1"`
	HomeTeam   string  `json:"home_team_name" example:"Team A"`
	AwayTeamID uint    `json:"away_team_id" example:"2"`
	AwayTeam   string  `json:"away_team_name" example:"Team B"`
	HomeGoals  *int    `json:"home_goals,omitempty" example:"2"`
	AwayGoals  *int    `json:"away_goals,omitempty" example:"1"`
	PlayedAt   *string `json:"played_at,omitempty" example:"2023-10-27T15:00:00Z"`
	Played     bool    `json:"played" example:"true"`
}

// MatchesResponse wraps the list of matches, total count, and week.
type MatchesResponse struct {
	Matches      []MatchDetailResponse `json:"matches"`
	TotalMatches int                   `json:"total_matches" example:"10"`
	Week         string                `json:"week,omitempty" example:"3"`
}

// SimulationResponse is a generic response for simulation actions.
type SimulationResponse struct {
	Message string `json:"message" example:"Operation successful"`
	Success bool   `json:"success" example:"true"`
}

// PredictionResult holds information for a single team's prediction.
// This struct was previously defined inline in GetPredictions handler.
type PredictionResult struct {
	TeamID      uint    `json:"team_id" example:"1"`
	TeamName    string  `json:"team_name" example:"Team A"`
	Probability float64 `json:"probability" example:"0.25"`
	CreatedAt   string  `json:"created_at" example:"2023-10-27 10:00:00"`
}

// PredictionsListResponse wraps the list of predictions.
type PredictionsListResponse struct {
	Week        uint               `json:"week" example:"4"`
	Predictions []PredictionResult `json:"predictions"`
	TotalTeams  int                `json:"total_teams" example:"4"`
	Method      string             `json:"method" example:"Monte Carlo Simulation (2,000 iterations)"`
}

// TeamPrediction holds information for a single team's champion prediction.
// This struct was previously defined inline in PredictChampion handler.
type TeamPrediction struct {
	TeamID      uint    `json:"team_id" example:"1"`
	TeamName    string  `json:"team_name" example:"Team X"`
	Probability float64 `json:"probability" example:"0.3"`
}

// ChampionPredictionResponse wraps the list of champion predictions.
type ChampionPredictionResponse struct {
	Status           string           `json:"status" example:"success"`
	Predictions      []TeamPrediction `json:"predictions"`
	RemainingMatches int64            `json:"remaining_matches" example:"5"`
	Method           string           `json:"method" example:"Monte Carlo Simulation (3,000 iterations)"`
	Message          string           `json:"message" example:"Championship predictions calculated based on current standings"`
}
