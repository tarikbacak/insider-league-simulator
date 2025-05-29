// Package poisson Poisson dağılımı bazlı maç simülasyonu sağlar
package poisson

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/tarikbacak/insider-league-simulator/internal/db"
	"github.com/tarikbacak/insider-league-simulator/internal/models"
	simModels "github.com/tarikbacak/insider-league-simulator/internal/simulator/models"
	"gorm.io/gorm"
)

// PoissonSimulator Poisson dağılımı ile maç simülasyonu yapar
type PoissonSimulator struct {
	db  *gorm.DB
	rng *rand.Rand // Global random number generator
}

// NewPoissonSimulator yeni bir Poisson simülatörü oluşturur
func NewPoissonSimulator() *PoissonSimulator {
	// Daha iyi random seed için çoklu kaynak kullan
	seed := time.Now().UnixNano() + int64(rand.Intn(1000000))
	return &PoissonSimulator{
		db:  db.GetDB(),
		rng: rand.New(rand.NewSource(seed)),
	}
}

// GetTeamStats veritabanından takım istatistiklerini alır
func (ps *PoissonSimulator) GetTeamStats(teamID uint) (*simModels.TeamStats, error) {
	var dbStats models.TeamStats
	err := ps.db.Select("team_id, avg_scored, avg_conceded, attack_strength, defense_strength").
		Where("team_id = ?", teamID).First(&dbStats).Error
	if err != nil {
		return nil, fmt.Errorf("takım istatistikleri alınamadı (ID: %d): %v", teamID, err)
	}

	stats := &simModels.TeamStats{
		TeamID:          dbStats.TeamID,
		AvgScored:       dbStats.AvgScored,
		AvgConceded:     dbStats.AvgConceded,
		AttackStrength:  dbStats.AttackStrength,
		DefenseStrength: dbStats.DefenseStrength,
	}

	return stats, nil
}

// CalculateMatchLambdas bir maç için ev sahibi ve deplasman lambda değerlerini hesaplar
func (ps *PoissonSimulator) CalculateMatchLambdas(homeTeamID, awayTeamID uint) (homeLambda, awayLambda float64, err error) {
	homeStats, err := ps.GetTeamStats(homeTeamID)
	if err != nil {
		return 0, 0, err
	}

	awayStats, err := ps.GetTeamStats(awayTeamID)
	if err != nil {
		return 0, 0, err
	}

	// Temel lambda değerlerini hesapla
	baseLambdaHome := homeStats.AttackStrength * awayStats.DefenseStrength * simModels.LeagueAverage
	baseLambdaAway := awayStats.AttackStrength * homeStats.DefenseStrength * simModels.LeagueAverage

	// Ev sahibi avantajı (gerçek futbolda %5-15 avantaj)
	homeAdvantage := 1.0 + (ps.rng.Float64() * 0.15) // %0-15 arası random avantaj
	baseLambdaHome *= homeAdvantage

	// Rastgele form faktörü (takımların o günkü performansı)
	homeFormFactor := 0.8 + (ps.rng.Float64() * 0.4) // 0.8 - 1.2 arası
	awayFormFactor := 0.8 + (ps.rng.Float64() * 0.4) // 0.8 - 1.2 arası

	homeLambda = baseLambdaHome * homeFormFactor
	awayLambda = baseLambdaAway * awayFormFactor

	// Minimum değer kontrolü
	if homeLambda < simModels.MinLambda {
		homeLambda = simModels.MinLambda
	}
	if awayLambda < simModels.MinLambda {
		awayLambda = simModels.MinLambda
	}

	// Maximum değer kontrolü (çok yüksek lambda'ları engelle)
	maxLambda := 4.0
	if homeLambda > maxLambda {
		homeLambda = maxLambda
	}
	if awayLambda > maxLambda {
		awayLambda = maxLambda
	}

	return homeLambda, awayLambda, nil
}

// GenerateGoals Poisson dağılımına göre gol sayısı üretir
// Daha gerçekçi sonuçlar için ek randomness ve futbol faktörleri ekler
func (ps *PoissonSimulator) GenerateGoals(lambda float64) int {
	// Temel Poisson dağılımı
	L := math.Exp(-lambda)
	k := 0
	p := 1.0

	for p > L {
		k++
		p *= ps.rng.Float64()
	}

	baseGoals := k - 1

	// Futbol gerçekçiliği için ek faktörler

	// 1. Momentum faktörü (takımların "şanslı/şanssız" günleri)
	momentumFactor := ps.rng.Float64()
	if momentumFactor < 0.1 { // %10 şans ile +1 gol bonus
		baseGoals++
	} else if momentumFactor > 0.9 { // %10 şans ile -1 gol penalty (minimum 0)
		if baseGoals > 0 {
			baseGoals--
		}
	}

	// 2. Çok nadir yüksek skorlar (0.5% şans)
	if ps.rng.Float64() < 0.005 {
		extraGoals := ps.rng.Intn(3) + 1 // 1-3 extra gol
		baseGoals += extraGoals
	}

	// 3. Çok düşük skorlar için bias (futbolda 0-0, 1-0 daha yaygın)
	if lambda < 1.0 && ps.rng.Float64() < 0.15 { // %15 şans ile düşük skor
		if baseGoals > 0 && ps.rng.Float64() < 0.5 {
			baseGoals = 0
		}
	}

	// Maximum skor limiti (çok absürd skorları engelle)
	if baseGoals > 8 {
		baseGoals = 8
	}

	return baseGoals
}

// SimulateMatch bir maçı simüle eder ve sonucu döndürür
func (ps *PoissonSimulator) SimulateMatch(homeTeamID, awayTeamID uint) (homeGoals, awayGoals int, err error) {
	homeLambda, awayLambda, err := ps.CalculateMatchLambdas(homeTeamID, awayTeamID)
	if err != nil {
		return 0, 0, err
	}

	homeGoals = ps.GenerateGoals(homeLambda)
	awayGoals = ps.GenerateGoals(awayLambda)

	log.Printf("Maç simülasyonu: Takım %d (%d gol) vs Takım %d (%d gol) - Lambda: %.2f, %.2f",
		homeTeamID, homeGoals, awayTeamID, awayGoals, homeLambda, awayLambda)

	return homeGoals, awayGoals, nil
}

// PlayNextWeek bir sonraki haftanın maçlarını oynatır
func (ps *PoissonSimulator) PlayNextWeek() error {
	var nextWeek uint
	result := ps.db.Model(&models.Match{}).
		Where("home_goals IS NULL AND away_goals IS NULL").
		Select("MIN(week)").
		Scan(&nextWeek)

	if result.Error != nil {
		return fmt.Errorf("sonraki hafta sorgulanamadı: %v", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("oynanmamış maç bulunamadı")
	}

	var matches []models.Match
	if err := ps.db.Where("week = ? AND home_goals IS NULL AND away_goals IS NULL", nextWeek).Find(&matches).Error; err != nil {
		return fmt.Errorf("hafta maçları sorgulanamadı: %v", err)
	}
	for _, match := range matches {
		// Her maç için random seed'i yenile (daha fazla çeşitlilik)
		ps.rng.Seed(time.Now().UnixNano() + int64(match.ID) + int64(ps.rng.Intn(10000)))

		homeGoals, awayGoals, err := ps.SimulateMatch(match.HomeTeamID, match.AwayTeamID)
		if err != nil {
			log.Printf("Maç simüle edilemedi (ID: %d): %v", match.ID, err)
			continue
		}

		homeGoalsUint := uint(homeGoals)
		awayGoalsUint := uint(awayGoals)
		match.HomeGoals = &homeGoalsUint
		match.AwayGoals = &awayGoalsUint
		match.PlayedAt = time.Now()

		if err := ps.db.Save(&match).Error; err != nil {
			log.Printf("Maç sonucu kaydedilemedi (ID: %d): %v", match.ID, err)
		}
	}

	log.Printf("%d. hafta maçları başarıyla simüle edildi", nextWeek)
	return nil
}

// PlayAllRemainingWeeks kalan tüm haftaları oynatır
func (ps *PoissonSimulator) PlayAllRemainingWeeks() error {
	for {
		var count int64
		if err := ps.db.Model(&models.Match{}).
			Where("home_goals IS NULL AND away_goals IS NULL").
			Count(&count).Error; err != nil {
			return fmt.Errorf("oynanmamış maç sayısı sorgulanamadı: %v", err)
		}

		if count == 0 {
			log.Println("Tüm maçlar tamamlandı")
			break
		}

		if err := ps.PlayNextWeek(); err != nil {
			return fmt.Errorf("hafta oynatılırken hata: %v", err)
		}
	}

	return nil
}
