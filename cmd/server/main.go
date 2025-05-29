// Package main is the entry point for the Insider League Simulator API
// @title Insider League Simulator API
// @version 1.0
// @description Insider Football League Simulator using Poisson Distribution & Monte Carlo methods
// @host localhost:8080
// @BasePath /api/v1
package main

import (
	"log"
	"net/http"

	_ "github.com/lib/pq"
	"github.com/tarikbacak/insider-league-simulator/config"
	_ "github.com/tarikbacak/insider-league-simulator/docs" // Swagger docs
	"github.com/tarikbacak/insider-league-simulator/internal/api"
	"github.com/tarikbacak/insider-league-simulator/internal/db"
)

func main() {
	// Initialize configuration (load .env file)
	config.Init()
	// Initialize database connection
	db.InitDB()

	// Initialize database data (create fixtures)
	db.InitializeData()

	// Set up the router
	router := api.SetupRouter()

	// Get server port from config
	cfg := config.GetConfig()
	port := ":" + cfg.Server.Port

	// Start the server
	log.Printf("Starting server on port %s...", port)
	if err := http.ListenAndServe(port, router); err != nil {
		log.Fatal(err)
	}
}
