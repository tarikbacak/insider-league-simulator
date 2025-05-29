package models

import "gorm.io/gorm"

// Team represents a football team.
type Team struct {
	gorm.Model           // ID field is already uint from gorm.Model
	Name       string    `json:"name" gorm:"unique;not null"`                       // Name of the team
	Attack     int       `json:"attack"`                                            // Offensive strength of the team
	Defense    int       `json:"defense"`                                           // Defensive strength of the team
	HomeGames  []Match   `json:"home_games,omitempty" gorm:"foreignKey:HomeTeamID"` // Matches where the team is the home side
	AwayGames  []Match   `json:"away_games,omitempty" gorm:"foreignKey:AwayTeamID"` // Matches where the team is the away side
	Stats      TeamStats `json:"stats" gorm:"foreignKey:TeamID"`                    // Team statistics
}
