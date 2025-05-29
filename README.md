# Insider League Simulator

## Description

Insider League Simulator is a backend application that simulates a football league. It uses Poisson Distribution for match simulations and Monte Carlo methods for championship predictions. The API allows users to view league standings, match results, simulate upcoming weeks, and get championship predictions.

## Features

-   **League Simulation:** Simulates football matches week by week.
-   **Match Results:** Provides results for past and simulated matches.
-   **League Standings:** Displays the current league table.
-   **Championship Predictions:** Uses Monte Carlo simulation to predict championship probabilities for later stages of the league.
-   **RESTful API:** Exposes endpoints for interacting with the simulator.
-   **Database Integration:** Stores team data, match results, and predictions in a PostgreSQL database.
-   **Web UI:** A simple web interface to view league progress.

## Technologies Used

-   **Backend:** Go (Golang)
-   **API Framework:** Gin-Gonic
-   **Database:** PostgreSQL
-   **Database Migration:** SQL
-   **Environment Configuration:** godotenv

## API Endpoints

The API is versioned and accessible under the `/api/v1` path.

-   `GET /api/v1/standings`: Returns the current league standings.
-   `GET /api/v1/matches?week=n`: Returns matches for the specified week. If no week is specified, returns all matches.
-   `POST /api/v1/matches/next`: Simulates the next week of the league.
-   `POST /api/v1/matches/all`: Simulates all remaining weeks of the league.
-   `GET /api/v1/predictions?week=n`: Returns championship predictions based on Monte Carlo simulation for the specified week (e.g., week 4, 5, or 6 for a 4-team league). This is the primary endpoint used by the web UI.
-   `POST /api/v1/init`: Resets and initializes the database with new random fixtures (for development purposes).
-   `GET /health`: Health check endpoint for the API.
-   `GET /swagger/*any`: Swagger API documentation.
-   `GET /web/league.html`: Access the simple web UI for the league.
-   `GET /`: Returns basic API information.

## Database Schema

The database schema consists of the following tables:

-   **`teams`**: Stores team information (id, name).
-   **`matches`**: Stores match details (id, week, home\_team\_id, away\_team\_id, home\_goals, away\_goals, played\_at).
-   **`team_stats`**: Stores team statistics for the Poisson model (team\_id, avg\_scored, avg\_conceded, attack\_strength, defense\_strength).
-   **`predictions`**: Stores championship prediction probabilities from Monte Carlo simulations (id, week, team\_id, probability, created\_at).

For more details, refer to the migration file: `migrations/001_create_tables.up.sql`.

## Setup and Installation

1.  **Prerequisites:**
    *   Go (version 1.x or higher)
    *   PostgreSQL
    *   Git

2.  **Clone the repository:**
    ```bash
    git clone github.com/tarikbacak/insider-league-simulator
    cd insider-league-simulator
    ```

3.  **Database Setup:**
    *   Create a PostgreSQL database (e.g., `insider_league`).
    *   Update the database connection details in a `.env` file in the root directory. Create one if it doesn't exist, based on `.env.example` (if provided) or the default values in `config/config.go`.
        Example `.env` file:
        ```env
        DB_HOST=localhost
        DB_PORT=5432
        DB_USER=your_postgres_user
        DB_PASSWORD=your_postgres_password
        DB_NAME=insider_league
        SERVER_PORT=8080
        ```

4.  **Install Dependencies:**
    ```bash
    go mod tidy
    ```

5.  **Run Migrations:**
    ```bash
    PGPASSWORD=password psql -h db_host -U db_user -d db_name -f migrations/001_create_tables.up.sql
    ```

6.  **Build the application:**
    ```bash
    go build -o bin/league-simulator ./cmd/server/main.go
    ```

## Usage

1.  **Run the server:**
    ```bash
    ./bin/league-simulator
    ```
    Or, for development:
    ```bash
    go run ./cmd/server/main.go
    ```
    The server will start, typically on `localhost:8080` (or the port specified in `SERVER_PORT`).

2.  **Access the API:**
    *   Use a tool like Postman or `curl` to interact with the API endpoints listed above.
    *   Access the Swagger documentation at `http://localhost:8080/swagger/index.html`.
    *   Access the simple web UI at `http://localhost:8080/web/league.html`.

## 🚀 Deployment
The application is deployed on Render.com and can be accessed at:
[https://insider-league-simulator.onrender.com/](https://insider-league-simulator.onrender.com/)

### Go Version
The project uses Go version `1.20.0`. This version was chosen for compatibility with the Render.com deployment environment.

