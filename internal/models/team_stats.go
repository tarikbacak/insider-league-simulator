package models

import "gorm.io/gorm"

// TeamStats holds team statistics
type TeamStats struct {
	gorm.Model
	TeamID          uint    `json:"team_id" gorm:"uniqueIndex;not null"`
	Played          uint    `json:"played"`           // Number of matches played
	Won             uint    `json:"won"`              // Number of matches won
	Drawn           uint    `json:"drawn"`            // Number of matches drawn
	Lost            uint    `json:"lost"`             // Number of matches lost
	GoalsFor        uint    `json:"goals_for"`        // Number of goals scored
	GoalsAway       uint    `json:"goals_away"`       // Number of goals conceded
	Points          uint    `json:"points"`           // Total points
	AvgScored       float64 `json:"avg_scored"`       // Average goals scored per match
	AvgConceded     float64 `json:"avg_conceded"`     // Average goals conceded per match
	AttackStrength  float64 `json:"attack_strength"`  // Attack strength (λ_scored / λ_league)
	DefenseStrength float64 `json:"defense_strength"` // Defense strength (λ_conceded / λ_league)
}

// UpdateStats updates the team's stats after a match
func (ts *TeamStats) UpdateStats(goalsScored, goalsConceded uint) {
	ts.Played++
	ts.GoalsFor += goalsScored
	ts.GoalsAway += goalsConceded

	if goalsScored > goalsConceded {
		ts.Won++
		ts.Points += 3
	} else if goalsScored == goalsConceded {
		ts.Drawn++
		ts.Points++
	} else {
		ts.Lost++
	}

	// Update averages
	ts.AvgScored = float64(ts.GoalsFor) / float64(ts.Played)
	ts.AvgConceded = float64(ts.GoalsAway) / float64(ts.Played)
}

// CalculateStrengths calculates attack and defense strengths based on league average
func (ts *TeamStats) CalculateStrengths(leagueAverage float64) {
	if leagueAverage <= 0 {
		leagueAverage = 1.5 // Default value if no league average available
	}

	if ts.Played > 0 {
		ts.AttackStrength = ts.AvgScored / leagueAverage
		ts.DefenseStrength = ts.AvgConceded / leagueAverage
	} else {
		// If no games played, set neutral values
		ts.AttackStrength = 1.0
		ts.DefenseStrength = 1.0
	}
}

// CalculateWinPercentage returns the team's win percentage
func (ts *TeamStats) CalculateWinPercentage() float64 {
	if ts.Played == 0 {
		return 0
	}
	return float64(ts.Won) / float64(ts.Played) * 100.0
}

// CalculateDrawPercentage returns the team's draw percentage
func (ts *TeamStats) CalculateDrawPercentage() float64 {
	if ts.Played == 0 {
		return 0
	}
	return float64(ts.Drawn) / float64(ts.Played) * 100.0
}

// CalculateLossPercentage returns the team's loss percentage
func (ts *TeamStats) CalculateLossPercentage() float64 {
	if ts.Played == 0 {
		return 0
	}
	return float64(ts.Lost) / float64(ts.Played) * 100.0
}

// ResetStats resets all team statistics
func (ts *TeamStats) ResetStats() {
	ts.Played = 0
	ts.Won = 0
	ts.Drawn = 0
	ts.Lost = 0
	ts.GoalsFor = 0
	ts.GoalsAway = 0
	ts.Points = 0
	ts.AvgScored = 0
	ts.AvgConceded = 0
	ts.AttackStrength = 0
	ts.DefenseStrength = 0
}
