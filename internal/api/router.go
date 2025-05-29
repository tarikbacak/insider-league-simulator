// Package api - HTTP API router and middleware configuration
package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/tarikbacak/insider-league-simulator/internal/db"
)

// SetupRouter configures API routes and middlewares
// Defines RESTful API endpoints using Gin framework
func SetupRouter() *gin.Engine { // Gin router'ını oluştur	// Use gin.SetMode(gin.ReleaseMode) in production environment
	router := gin.Default()

	// Add CORS middleware (for frontend integration)
	router.Use(corsMiddleware())

	// For static files (HTML, CSS, JS) - before applying middlewares
	router.Static("/web", "./web")
	// Create API v1 route group
	v1 := router.Group("/api/v1")
	v1.Use(jsonMiddleware()) // JSON middleware only for API endpoints
	{                        // Standings endpoint
		// GET /api/v1/standings - Returns current league standings
		v1.GET("/standings", GetStandings)

		// Match listing endpoint
		// GET /api/v1/matches?week=n - Returns matches for the specified week
		// Returns all matches without query parameter
		v1.GET("/matches", GetMatches)

		// Next week simulation endpoint
		// POST /api/v1/matches/next - Simulates the next week
		v1.POST("/matches/next", PlayNextWeek)

		// Simulate all remaining weeks endpoint
		// POST /api/v1/matches/all - Simulates all remaining weeks
		v1.POST("/matches/all", PlayAllWeeks) // Championship predictions endpoint

		// GET /api/v1/predictions?week=4|5 - Monte Carlo simulation for championship probabilities
		v1.GET("/predictions", GetPredictions)

		// Database initialization endpoint
		// POST /api/v1/init - Resets and initializes database (for development)
		v1.POST("/init", InitializeDatabase)
	} // Health check endpoint
	router.GET("/health", jsonMiddleware(), HealthCheck)

	// Swagger documentation
	// @title Insider League Simulator API
	// @version 1.0
	// @description This is a football league simulator API.
	// @termsOfService http://swagger.io/terms/
	// @contact.name API Support
	// @contact.url http://www.swagger.io/support
	// @contact.email support@swagger.io
	// @license.name Apache 2.0
	// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
	// @host localhost:8080
	// @BasePath /api/v1
	// @schemes http https
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Homepage endpoint - Returns API information
	router.GET("/", jsonMiddleware(), GetAPIInfo)

	return router
}

// corsMiddleware configures Cross-Origin Resource Sharing settings
// Required for frontend application to access the API
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		// For preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// jsonMiddleware sets necessary headers for JSON responses
func jsonMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.Next()
	}
}

// HealthCheck simple health check endpoint
// @Summary Health check
// @Description Checks the health of the service
// @Tags health
// @Produce json
// @Success 200 {object} HealthCheckResponse
// @Router /health [get]
func HealthCheck(c *gin.Context) {
	c.JSON(200, HealthCheckResponse{
		Status:  "healthy",
		Service: "insider-league-simulator",
		Timestamp: HealthCheckTimestamp{
			Unix: HealthCheckUnixTimestamp{
				Seconds: HealthCheckSecondsValue{
					Value: time.Now().Format("2006-01-02 15:04:05"),
				},
			},
		},
	})
}

// InitializeDatabase resets the database and initializes it with new data
// @Summary Initialize database
// @Description Resets and initializes the database with new random fixtures
// @Tags reset
// @Produce json
// @Success 200 {object} InitResponse
// @Failure 500 {object} ErrorResponse
// @Router /init [post]
func InitializeDatabase(c *gin.Context) {
	// Should only be used in development environment
	if err := db.InitializeData(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:     "Database initialization error",
			Detail:    err.Error(),
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		})
		return
	}

	c.JSON(http.StatusOK, InitResponse{
		Message:   "Database initialized successfully",
		Note:      "New random fixture generated",
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	})
}

// Homepage endpoint - Returns API information
// @Summary API Information
// @Description Returns basic information about the API and its endpoints.
// @Tags info
// @Produce json
// @Success 200 {object} APIInfoResponse
// @Router / [get]
func GetAPIInfo(c *gin.Context) {
	c.JSON(http.StatusOK, APIInfoResponse{
		Message: "Insider League Simulator API",
		Version: "1.0.0",
		Endpoints: map[string]string{
			"standings":   "GET /api/v1/standings",
			"matches":     "GET /api/v1/matches?week=n",
			"next_week":   "POST /api/v1/matches/next",
			"play_all":    "POST /api/v1/matches/all",
			"predictions": "GET /api/v1/predictions?week=n",
			"init_db":     "POST /api/v1/init",
			"health":      "GET /health",
			"swagger":     "GET /swagger/index.html",
			"web_ui":      "GET /web/league.html",
		},
	})
}
