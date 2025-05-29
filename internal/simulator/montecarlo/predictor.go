// Package montecarlo Monte Carlo simülasyonu ile şampiyonluk tahminleri yapar
package montecarlo

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/tarikbacak/insider-league-simulator/internal/db"
	"github.com/tarikbacak/insider-league-simulator/internal/models"
	"github.com/tarikbacak/insider-league-simulator/internal/simulator/poisson"
	"gorm.io/gorm"
)

// MonteCarloPredictor Monte Carlo simülasyonu ile şampiyonluk tahmini yapar
type MonteCarloPredictor struct {
	db         *gorm.DB
	simulator  *poisson.PoissonSimulator
	iterations int
	teamStats  map[uint]*TeamStats // Cache for team stats
}

// TeamStats team statistics cache
type TeamStats struct {
	AttackStrength  float64
	DefenseStrength float64
}

// NewMonteCarloPredictor yeni bir Monte Carlo tahmin edici oluşturur
func NewMonteCarloPredictor(iterations int) *MonteCarloPredictor {
	// Hızlı tahminler için iterasyon sayısını azalt
	if iterations > 5000 {
		iterations = 5000 // Maximum 5000 iteration for speed
	}
	if iterations < 1000 {
		iterations = 1000 // Minimum 1000 for accuracy
	}

	return &MonteCarloPredictor{
		db:         db.GetDB(),
		simulator:  poisson.NewPoissonSimulator(),
		iterations: iterations,
		teamStats:  make(map[uint]*TeamStats),
	}
}

// PredictChampionshipProbabilities belirtilen hafta için şampiyonluk olasılıklarını hesaplar
func (mcp *MonteCarloPredictor) PredictChampionshipProbabilities(week uint) (map[uint]float64, error) { // Cache team stats once for all iterations
	if err := mcp.loadTeamStats(); err != nil {
		return nil, fmt.Errorf("takım istatistikleri yüklenemedi: %v", err)
	}

	currentStandings, err := mcp.getCurrentStandings()
	if err != nil {
		return nil, fmt.Errorf("mevcut puan durumu alınamadı: %v", err)
	}

	remainingMatches, err := mcp.getRemainingMatches()
	if err != nil {
		return nil, fmt.Errorf("kalan maçlar alınamadı: %v", err)
	}

	if len(remainingMatches) == 0 {
		return nil, fmt.Errorf("tüm maçlar tamamlanmış, tahmin yapılamaz")
	}

	log.Printf("Monte Carlo simülasyonu başlatılıyor: %d iterasyon, %d kalan maç", mcp.iterations, len(remainingMatches))

	championCounts := make(map[uint]int)

	// Batch process iterations for speed
	batchSize := 100
	for batch := 0; batch < mcp.iterations; batch += batchSize {
		currentBatchSize := batchSize
		if batch+batchSize > mcp.iterations {
			currentBatchSize = mcp.iterations - batch
		}

		for i := 0; i < currentBatchSize; i++ {
			standings := mcp.copyStandings(currentStandings)
			// Kalan maçları hızlı simüle et
			for _, match := range remainingMatches {
				homeGoals, awayGoals := mcp.fastSimulateMatch(match.HomeTeamID, match.AwayTeamID)
				mcp.updateStandings(standings, match.HomeTeamID, match.AwayTeamID, homeGoals, awayGoals)
			}

			championID := mcp.findChampion(standings)
			championCounts[championID]++
		}
	}

	// Olasılıkları hesapla
	probabilities := mcp.calculateProbabilities(championCounts)

	// Tahminleri kaydet
	if err := mcp.savePredictions(week, probabilities); err != nil {
		log.Printf("Tahminler kaydedilirken hata: %v", err)
	}

	return probabilities, nil
}

// getCurrentStandings mevcut puan durumunu döndürür
func (mcp *MonteCarloPredictor) getCurrentStandings() (map[uint]int, error) {
	var teams []struct {
		ID     uint
		Points int
	}

	err := mcp.db.Model(&models.Team{}).
		Select("teams.id, COALESCE(SUM(CASE " +
			"WHEN matches.home_team_id = teams.id AND matches.home_goals > matches.away_goals THEN 3 " +
			"WHEN matches.away_team_id = teams.id AND matches.away_goals > matches.home_goals THEN 3 " +
			"WHEN matches.home_goals = matches.away_goals THEN 1 " +
			"ELSE 0 END), 0) as points").
		Joins("LEFT JOIN matches ON (teams.id = matches.home_team_id OR teams.id = matches.away_team_id) " +
			"AND matches.home_goals IS NOT NULL AND matches.away_goals IS NOT NULL").
		Group("teams.id").
		Scan(&teams).Error

	if err != nil {
		return nil, err
	}

	standings := make(map[uint]int)
	for _, t := range teams {
		standings[t.ID] = t.Points
	}

	return standings, nil
}

// getRemainingMatches oynanmamış maçları döndürür
func (mcp *MonteCarloPredictor) getRemainingMatches() ([]models.Match, error) {
	var matches []models.Match
	err := mcp.db.Where("home_goals IS NULL AND away_goals IS NULL").
		Order("week, id").
		Find(&matches).Error
	return matches, err
}

// copyStandings puan durumunun bir kopyasını oluşturur
func (mcp *MonteCarloPredictor) copyStandings(original map[uint]int) map[uint]int {
	copy := make(map[uint]int)
	for k, v := range original {
		copy[k] = v
	}
	return copy
}

// updateStandings puan durumunu günceller
func (mcp *MonteCarloPredictor) updateStandings(standings map[uint]int, homeID, awayID uint, homeGoals, awayGoals int) {
	if homeGoals > awayGoals {
		standings[homeID] += 3
	} else if homeGoals < awayGoals {
		standings[awayID] += 3
	} else {
		standings[homeID]++
		standings[awayID]++
	}
}

// findChampion en yüksek puana sahip takımı bulur
func (mcp *MonteCarloPredictor) findChampion(standings map[uint]int) uint {
	var championID uint
	maxPoints := -1
	for teamID, points := range standings {
		if points > maxPoints {
			maxPoints = points
			championID = teamID
		}
	}
	return championID
}

// calculateProbabilities şampiyonluk olasılıklarını hesaplar
func (mcp *MonteCarloPredictor) calculateProbabilities(championCounts map[uint]int) map[uint]float64 {
	probabilities := make(map[uint]float64)
	totalIterations := float64(mcp.iterations)

	// İlk önce tüm takımları 0% ile başlat
	var teams []models.Team
	if err := mcp.db.Find(&teams).Error; err == nil {
		for _, team := range teams {
			probabilities[team.ID] = 0.0
		}
	}

	// Gerçek olasılıkları hesapla
	for teamID, count := range championCounts {
		probabilities[teamID] = (float64(count) / totalIterations) * 100.0
	}

	// Normalizasyon: Toplamın %100 olduğundan emin ol
	total := 0.0
	for _, prob := range probabilities {
		total += prob
	}

	// Eğer toplam %100 değilse, normalize et
	if total > 0 && total != 100.0 {
		normalizationFactor := 100.0 / total
		for teamID := range probabilities {
			probabilities[teamID] *= normalizationFactor
		}
	}

	return probabilities
}

// savePredictions tahminleri veritabanına kaydeder
func (mcp *MonteCarloPredictor) savePredictions(week uint, probabilities map[uint]float64) error {
	return mcp.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&models.Prediction{}, "week = ?", week).Error; err != nil {
			return err
		}
		for teamID, probability := range probabilities {
			prediction := models.Prediction{
				Week:        week,
				TeamID:      teamID,
				Probability: probability,
			}
			if err := tx.Create(&prediction).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// loadTeamStats loads and caches team statistics for fast access
func (mcp *MonteCarloPredictor) loadTeamStats() error {
	var teams []models.Team
	if err := mcp.db.Find(&teams).Error; err != nil {
		return err
	}

	for _, team := range teams {
		var stats models.TeamStats
		if err := mcp.db.Where("team_id = ?", team.ID).First(&stats).Error; err != nil {
			continue
		}

		mcp.teamStats[team.ID] = &TeamStats{
			AttackStrength:  stats.AttackStrength,
			DefenseStrength: stats.DefenseStrength,
		}
	}

	return nil
}

// fastSimulateMatch performs fast match simulation without database access
func (mcp *MonteCarloPredictor) fastSimulateMatch(homeTeamID, awayTeamID uint) (homeGoals, awayGoals int) {
	homeStats, homeExists := mcp.teamStats[homeTeamID]
	awayStats, awayExists := mcp.teamStats[awayTeamID]

	if !homeExists || !awayExists {
		// Fallback to simple random
		return mcp.simpleRandomScore(), mcp.simpleRandomScore()
	}

	// Simplified lambda calculation (no database access)
	leagueAvg := 1.5                                                                     // Average goals per team per match
	homeLambda := homeStats.AttackStrength * awayStats.DefenseStrength * leagueAvg * 1.1 // Home advantage
	awayLambda := awayStats.AttackStrength * homeStats.DefenseStrength * leagueAvg

	// Fast Poisson approximation
	homeGoals = mcp.fastPoisson(homeLambda)
	awayGoals = mcp.fastPoisson(awayLambda)

	return homeGoals, awayGoals
}

// fastPoisson fast approximation of Poisson distribution
func (mcp *MonteCarloPredictor) fastPoisson(lambda float64) int {
	if lambda < 0.1 {
		return 0
	}

	// Simple approximation for speed
	// For small lambda, use probability tables
	switch {
	case lambda < 1.0:
		r := mcp.simulator.GenerateGoals(lambda)
		if r > 3 {
			r = 3
		} // Cap for realism
		return r
	case lambda < 2.0:
		r := mcp.simulator.GenerateGoals(lambda)
		if r > 5 {
			r = 5
		}
		return r
	default:
		r := mcp.simulator.GenerateGoals(lambda)
		if r > 7 {
			r = 7
		}
		return r
	}
}

// simpleRandomScore generates simple random score for fallback
func (mcp *MonteCarloPredictor) simpleRandomScore() int {
	// Weighted random for realistic football scores
	r := float64(rand.Intn(1000)) / 1000.0
	switch {
	case r < 0.3:
		return 0 // 30% chance
	case r < 0.6:
		return 1 // 30% chance
	case r < 0.8:
		return 2 // 20% chance
	case r < 0.95:
		return 3 // 15% chance
	default:
		return 4 // 5% chance
	}
}
