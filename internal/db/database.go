package db

import (
	"fmt"
	"log"

	"github.com/tarikbacak/insider-league-simulator/config"
	"github.com/tarikbacak/insider-league-simulator/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB is the global database connection object
var DB *gorm.DB

// InitDB initializes the database connection and performs auto-migration using GORM models
// It uses the config package to get connection info from the .env file
func InitDB() {
	cfg := config.GetConfig()

	var err error
	// Connect to the database using GORM
	DB, err = gorm.Open(postgres.Open(cfg.GetDatabaseURL()), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	// Auto-Migration: Automatically create/update tables
	err = DB.AutoMigrate(&models.Team{}, &models.Match{}, &models.TeamStats{}, &models.Prediction{})
	if err != nil {
		log.Fatalf("Auto-migration error: %v", err)
	}

	fmt.Printf("Database connection and migration completed successfully - %s:%s\n", cfg.Database.Host, cfg.Database.Port)
}

// GetDB returns the global database connection object
func GetDB() *gorm.DB {
	if DB == nil {
		log.Fatal("Database connection has not been initialized")
	}
	return DB
}

// InitializeData creates the initial data required for the league
func InitializeData() error {
	// First, clear existing data
	if err := clearExistingData(); err != nil {
		return fmt.Errorf("error clearing existing data: %v", err)
	}

	// Define the teams
	teams := []models.Team{
		{Name: "Galatasaray", Attack: 80, Defense: 75},
		{Name: "Fenerbahçe", Attack: 70, Defense: 70},
		{Name: "Beşiktaş", Attack: 60, Defense: 60},
		{Name: "Trabzonspor", Attack: 50, Defense: 50},
	}

	// Create TeamStats for each team
	for _, team := range teams {
		if err := DB.Create(&team).Error; err != nil {
			return fmt.Errorf("error creating team: %v", err)
		}

		stats := models.TeamStats{
			TeamID:          team.ID,
			Played:          0,
			Won:             0,
			Drawn:           0,
			Lost:            0,
			GoalsFor:        0,
			GoalsAway:       0,
			Points:          0,
			AvgScored:       float64(team.Attack) / 100.0,
			AvgConceded:     float64(100-team.Defense) / 100.0,
			AttackStrength:  float64(team.Attack) / 75.0,
			DefenseStrength: float64(team.Defense) / 75.0,
		}
		if err := DB.Create(&stats).Error; err != nil {
			return fmt.Errorf("error creating team stats: %v", err)
		}
	}

	// Fetch all teams from the DB
	var allTeams []models.Team
	if err := DB.Find(&allTeams).Error; err != nil {
		return fmt.Errorf("error fetching teams: %v", err)
	}

	// Generate fixtures
	matches := generateFixtures(allTeams)

	// Insert matches into the DB
	for _, match := range matches {
		if err := DB.Create(&match).Error; err != nil {
			return fmt.Errorf("error creating match: %v", err)
		}
	}

	return nil
}

// generateFixtures creates a round-robin fixture for all teams
// For 4 teams: a 6-week round-robin league, 2 matches per week
func generateFixtures(teams []models.Team) []models.Match {
	var matches []models.Match

	if len(teams) != 4 {
		log.Printf("Warning: Expected 4 teams, got %d", len(teams))
		return matches
	}

	// First half of the season
	firstHalfFixtures := [][]int{
		{0, 1, 2, 3}, // Week 1: 0 vs 1, 2 vs 3
		{0, 2, 1, 3}, // Week 2: 0 vs 2, 1 vs 3
		{0, 3, 1, 2}, // Week 3: 0 vs 3, 1 vs 2
	}

	for week, fixtures := range firstHalfFixtures {
		weekNumber := uint(week + 1)
		for i := 0; i < len(fixtures); i += 2 {
			match := models.Match{
				HomeTeamID: teams[fixtures[i]].ID,
				AwayTeamID: teams[fixtures[i+1]].ID,
				Week:       weekNumber,
			}
			matches = append(matches, match)
		}
	}

	// Second half (reverse fixtures)
	secondHalfFixtures := [][]int{
		{1, 0, 3, 2}, // Week 4
		{2, 0, 3, 1}, // Week 5
		{3, 0, 2, 1}, // Week 6
	}

	for week, fixtures := range secondHalfFixtures {
		weekNumber := uint(week + 4)
		for i := 0; i < len(fixtures); i += 2 {
			match := models.Match{
				HomeTeamID: teams[fixtures[i]].ID,
				AwayTeamID: teams[fixtures[i+1]].ID,
				Week:       weekNumber,
			}
			matches = append(matches, match)
		}
	}

	log.Printf("Generated %d matches for 6 weeks", len(matches))
	for week := uint(1); week <= 6; week++ {
		count := 0
		for _, match := range matches {
			if match.Week == week {
				count++
			}
		}
		log.Printf("Week %d: %d matches", week, count)
	}

	return matches
}

// clearExistingData deletes all existing records in proper order
func clearExistingData() error {
	if err := DB.Exec("DELETE FROM predictions").Error; err != nil {
		return fmt.Errorf("error deleting predictions: %v", err)
	}
	if err := DB.Exec("DELETE FROM team_stats").Error; err != nil {
		return fmt.Errorf("error deleting team_stats: %v", err)
	}
	if err := DB.Exec("DELETE FROM matches").Error; err != nil {
		return fmt.Errorf("error deleting matches: %v", err)
	}
	if err := DB.Exec("DELETE FROM teams").Error; err != nil {
		return fmt.Errorf("error deleting teams: %v", err)
	}
	return nil
}
