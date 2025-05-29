package models

import (
	"time"

	"gorm.io/gorm"
)

// Match represents a football match.
type Match struct {
	gorm.Model
	Week       uint      `json:"week" gorm:"not null;check:week >= 1 AND week <= 6"` // Week in which the match is played
	HomeTeamID uint      `json:"home_team_id" gorm:"not null"`                       // ID of the home team
	AwayTeamID uint      `json:"away_team_id" gorm:"not null"`                       // ID of the away team
	HomeTeam   Team      `json:"home_team" gorm:"foreignKey:HomeTeamID"`             // Home team
	AwayTeam   Team      `json:"away_team" gorm:"foreignKey:AwayTeamID"`             // Away team
	HomeGoals  *uint     `json:"home_goals"`                                         // Number of goals scored by the home team (nil if not played)
	AwayGoals  *uint     `json:"away_goals"`                                         // Number of goals scored by the away team (nil if not played)
	PlayedAt   time.Time `json:"played_at" gorm:"index"`                             // Date and time the match was played
}
