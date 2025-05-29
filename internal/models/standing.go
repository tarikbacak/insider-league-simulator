package models

// Standing represents a team's position in the league table.
type Standing struct {
	TeamID         uint   `json:"team_id"`         // Team ID
	TeamName       string `json:"team_name"`       // Team name
	Played         uint   `json:"played"`          // Number of matches played
	Won            uint   `json:"won"`             // Number of matches won
	Drawn          uint   `json:"drawn"`           // Number of matches drawn
	Lost           uint   `json:"lost"`            // Number of matches lost
	GoalsFor       uint   `json:"goals_for"`       // Number of goals scored
	GoalsAgainst   uint   `json:"goals_against"`   // Number of goals conceded
	GoalDifference int    `json:"goal_difference"` // Goal difference (can be negative)
	Points         uint   `json:"points"`          // Total points
}
