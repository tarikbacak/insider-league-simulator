package models

// Prediction represents a team's championship prediction.
type Prediction struct {
	ID          uint    `json:"id" gorm:"primaryKey"`       // Unique ID of the prediction
	Week        uint    `json:"week"`                       // Week in which the prediction was made
	TeamID      uint    `json:"team_id"`                    // ID of the team
	Team        Team    `json:"-" gorm:"foreignKey:TeamID"` // Associated team (not exposed in JSON)
	Probability float64 `json:"probability"`                // Probability of winning the championship (percentage)
}
