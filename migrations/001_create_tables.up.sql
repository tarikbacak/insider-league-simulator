-- Teams table
-- Stores a unique ID and name for each team
CREATE TABLE teams (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

-- Matches table
-- Stores match details including week, participating teams, and results
-- If home_goals and away_goals are NULL, the match has not been played yet
CREATE TABLE matches (
    id BIGSERIAL PRIMARY KEY,
    week BIGINT NOT NULL CHECK (week >= 1 AND week <= 6), -- 6-week league
    home_team_id BIGINT NOT NULL REFERENCES teams(id),
    away_team_id BIGINT NOT NULL REFERENCES teams(id),
    home_goals INT,      -- NULL if not yet played
    away_goals INT,      -- NULL if not yet played
    played_at TIMESTAMP, -- Date and time the match was played
    CONSTRAINT different_teams CHECK (home_team_id != away_team_id)
);

-- Team statistics table
-- Stores parameters required for the Poisson model
CREATE TABLE team_stats (
    team_id BIGINT PRIMARY KEY REFERENCES teams(id),
    avg_scored NUMERIC(4,2) NOT NULL,    -- Average goals scored per match
    avg_conceded NUMERIC(4,2) NOT NULL,  -- Average goals conceded per match
    attack_strength NUMERIC(4,2),        -- Attack strength (λ_scored / λ_league)
    defense_strength NUMERIC(4,2)        -- Defense strength (λ_conceded / λ_league)
);

-- Championship predictions table
-- Stores the results of Monte Carlo simulation
CREATE TABLE predictions (
    id BIGSERIAL PRIMARY KEY,
    week BIGINT NOT NULL, -- Predictions only for weeks 4 and 5
    team_id BIGINT NOT NULL REFERENCES teams(id),
    probability NUMERIC(5,2) NOT NULL,         -- Percentage value (0.00–100.00)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(week, team_id) -- Only one prediction per team per week
);

-- Indexes for performance
CREATE INDEX idx_matches_week ON matches(week);
CREATE INDEX idx_matches_teams ON matches(home_team_id, away_team_id);
CREATE INDEX idx_predictions_week ON predictions(week);
