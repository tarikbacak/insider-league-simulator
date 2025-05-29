// Package api - HTTP API handler functions
package api

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tarikbacak/insider-league-simulator/internal/db"
	"github.com/tarikbacak/insider-league-simulator/internal/models"
	"github.com/tarikbacak/insider-league-simulator/internal/simulator"
)

// GetStandings returns the current league standings
// @Summary Get league standings
// @Description Returns current league table with teams' points, goals, and other statistics
// @Tags standings
// @Produce json
// @Success 200 {object} StandingsResponse
// @Failure 500 {object} ErrorResponse
// @Router /standings [get]
func GetStandings(c *gin.Context) {
	database := db.GetDB()

	var teams []models.Team
	// Fetch teams and their related matches
	err := database.
		Preload("HomeGames", "home_goals IS NOT NULL AND away_goals IS NOT NULL").
		Preload("AwayGames", "home_goals IS NOT NULL AND away_goals IS NOT NULL").
		Find(&teams).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:  "Could not calculate standings",
			Detail: err.Error(),
		})
		return
	}

	var standings []models.Standing
	for _, team := range teams {
		standing := models.Standing{
			TeamID:   team.ID,
			TeamName: team.Name,
		}

		// Calculate match statistics
		for _, match := range team.HomeGames {
			if match.HomeGoals != nil && match.AwayGoals != nil {
				standing.Played++
				standing.GoalsFor += uint(*match.HomeGoals) // Assuming HomeGoals is *int or similar, casting to uint
				standing.GoalsAgainst += uint(*match.AwayGoals)
				if *match.HomeGoals > *match.AwayGoals {
					standing.Won++
					standing.Points += 3
				} else if *match.HomeGoals < *match.AwayGoals {
					standing.Lost++
				} else {
					standing.Drawn++
					standing.Points++
				}
			}
		}

		for _, match := range team.AwayGames {
			if match.HomeGoals != nil && match.AwayGoals != nil {
				standing.Played++
				standing.GoalsFor += uint(*match.AwayGoals) // Team was away, so AwayGoals are their scored goals
				standing.GoalsAgainst += uint(*match.HomeGoals)
				if *match.AwayGoals > *match.HomeGoals {
					standing.Won++
					standing.Points += 3
				} else if *match.AwayGoals < *match.HomeGoals {
					standing.Lost++
				} else {
					standing.Drawn++
					standing.Points++
				}
			}
		}

		standing.GoalDifference = int(standing.GoalsFor) - int(standing.GoalsAgainst)
		standings = append(standings, standing)
	}

	// Sort standings (by points and goal difference)
	sort.SliceStable(standings, func(i, j int) bool {
		if standings[i].Points != standings[j].Points {
			return standings[i].Points > standings[j].Points
		}
		return standings[i].GoalDifference > standings[j].GoalDifference
	})

	c.JSON(http.StatusOK, StandingsResponse{
		Standings:  standings,
		TotalTeams: len(standings),
	})
}

// GetMatches returns the list of matches
// @Summary Get match list
// @Description Returns list of matches for a specific week or all weeks
// @Tags matches
// @Produce json
// @Param week query integer false "Week number" mininum(1)
// @Success 200 {object} MatchesResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /matches [get]
func GetMatches(c *gin.Context) {
	database := db.GetDB()
	weekParam := c.Query("week")

	var totalTeams int64
	database.Model(&models.Team{}).Count(&totalTeams)
	maxWeeks := 0
	if totalTeams > 1 {
		maxWeeks = (int(totalTeams) - 1) * 2
	}

	var matches []models.Match
	query := database.
		Preload("HomeTeam").
		Preload("AwayTeam")

	if weekParam != "" {
		week, err := strconv.Atoi(weekParam)
		if err != nil || week < 1 || (maxWeeks > 0 && week > maxWeeks) {
			detail := "Week must be a positive number."
			if maxWeeks > 0 {
				detail = "Week must be a number between 1 and " + strconv.Itoa(maxWeeks) + "."
			}
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:  "Invalid week parameter",
				Detail: detail,
			})
			return
		}
		query = query.Where("week = ?", week)
	}

	err := query.Order("week, id").Find(&matches).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:  "Could not retrieve matches",
			Detail: err.Error(),
		})
		return
	}

	var matchesResponse []MatchDetailResponse
	for _, match := range matches {
		played := match.HomeGoals != nil && match.AwayGoals != nil
		var homeGoals, awayGoals *int
		var playedAtStr *string

		if match.HomeGoals != nil { // Assuming models.Match.HomeGoals is *int or convertible
			hg := int(*match.HomeGoals) // Convert to int if necessary
			homeGoals = &hg
		}
		if match.AwayGoals != nil { // Assuming models.Match.AwayGoals is *int or convertible
			ag := int(*match.AwayGoals)
			awayGoals = &ag
		}
		// Assuming PlayedAt is sql.NullTime or similar structure that has Valid and Time fields
		if !match.PlayedAt.IsZero() { // Check if PlayedAt is not the zero value for time.Time
			pat := match.PlayedAt.Format(time.RFC3339)
			playedAtStr = &pat
		}

		matchesResponse = append(matchesResponse, MatchDetailResponse{
			ID:         match.ID,
			Week:       match.Week,
			HomeTeamID: match.HomeTeamID,
			HomeTeam:   match.HomeTeam.Name,
			AwayTeamID: match.AwayTeamID,
			AwayTeam:   match.AwayTeam.Name,
			HomeGoals:  homeGoals,
			AwayGoals:  awayGoals,
			PlayedAt:   playedAtStr,
			Played:     played,
		})
	}

	c.JSON(http.StatusOK, MatchesResponse{
		Matches:      matchesResponse,
		TotalMatches: len(matchesResponse),
		Week:         weekParam, // Keep original string for response if provided
	})
}

// PlayNextWeek simulates matches for the next week
// @Summary Simulate next week's matches
// @Description Simulates all matches for the next unplayed week using Poisson distribution
// @Tags simulation
// @Produce json
// @Success 200 {object} SimulationResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /matches/next [post]
func PlayNextWeek(c *gin.Context) {
	sim := simulator.GetPoissonSimulator()

	// Assuming PlayNextWeek now returns (playedWeek uint, err error)
	// If it only returns error, we need to adjust how to get the playedWeek for the message.
	// For now, let's assume it only returns error and we simplify the message.
	err := sim.PlayNextWeek() // If PlayNextWeek returns playedWeek, use: playedWeek, err := sim.PlayNextWeek()
	if err != nil {
		if err.Error() == "all matches have been played" || err.Error() == "no more weeks to play" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:  "Simulation error",
				Detail: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:  "Failed to simulate next week",
				Detail: err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, SimulationResponse{
		Message: "Next week successfully simulated", // Simplified message
		Success: true,
	})
}

// PlayAllWeeks simulates all remaining weeks
// @Summary Simulate all remaining weeks
// @Description Simulates all remaining unplayed matches until the end of the season
// @Tags simulation
// @Produce json
// @Success 200 {object} SimulationResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /matches/all [post]
func PlayAllWeeks(c *gin.Context) {
	sim := simulator.GetPoissonSimulator()

	err := sim.PlayAllRemainingWeeks()
	if err != nil {
		if err.Error() == "all matches have been played" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:  "Simulation error",
				Detail: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:  "Failed to simulate all weeks",
				Detail: err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, SimulationResponse{
		Message: "All remaining weeks successfully simulated",
		Success: true,
	})
}

// GetPredictions returns championship predictions
// @Summary Get championship predictions
// @Description Returns championship predictions based on Monte Carlo simulation for a specific week.
// @Tags predictions
// @Produce json
// @Param week query integer true "Week number for prediction"
// @Success 200 {object} PredictionsListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /predictions [get]
func GetPredictions(c *gin.Context) {
	weekParam := c.Query("week")
	if weekParam == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Week parameter is required"})
		return
	}
	weekInt, err := strconv.ParseUint(weekParam, 10, 32)

	database := db.GetDB()
	var totalTeams int64
	database.Model(&models.Team{}).Count(&totalTeams)
	maxWeeks := 0
	if totalTeams > 1 {
		maxWeeks = (int(totalTeams) - 1) * 2
	}
	minPredictionWeek := uint64(1) // Allow predictions from week 1
	if maxWeeks > 0 {
		minPredictionWeek = uint64(maxWeeks / 2) // Or some other logic e.g. start predictions from mid-season
		if minPredictionWeek == 0 {
			minPredictionWeek = 1
		}
	}

	if err != nil || weekInt < minPredictionWeek || (maxWeeks > 0 && weekInt > uint64(maxWeeks)) {
		detail := "Invalid week parameter for predictions. "
		if maxWeeks > 0 {
			detail += "Week must be a number between " + strconv.FormatUint(minPredictionWeek, 10) + " and " + strconv.Itoa(maxWeeks) + "."
		} else {
			detail += "Please ensure teams are initialized for dynamic week calculation."
		}
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:  "Invalid week parameter for predictions",
			Detail: detail,
		})
		return
	}

	week := uint(weekInt)

	// Check existing predictions
	var count int64
	database.Model(&models.Prediction{}).Where("week = ?", week).Count(&count)
	// If no predictions exist, generate new ones
	if count == 0 {
		predictor := simulator.GetMonteCarloPredictor(2000)        // Reduced from 10000 for faster predictions
		_, err := predictor.PredictChampionshipProbabilities(week) // This will save predictions
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:  "Failed to generate predictions",
				Detail: err.Error(),
			})
			return
		}
	}

	var predictions []PredictionResult

	err = database.Table("predictions").
		Select("predictions.team_id, teams.name as team_name, predictions.probability, TO_CHAR(predictions.created_at, 'YYYY-MM-DD HH24:MI:SS') as created_at").
		Joins("JOIN teams ON predictions.team_id = teams.id").
		Where("predictions.week = ?", week).
		Order("predictions.probability DESC").
		Scan(&predictions).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:  "Could not retrieve predictions",
			Detail: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, PredictionsListResponse{
		Week:        week,
		Predictions: predictions,
		TotalTeams:  len(predictions),
		Method:      "Monte Carlo Simulation (2,000 iterations)",
	})
}
