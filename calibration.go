package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"
)

type CalibrationParams struct {
	SeasonID                     string
	Seed                         int64
	P0                           int
	CBase                        int
	Alpha                        float64
	SScale                       float64
	GScale                       float64
	Beta                         float64
	Gamma                        float64
	DailyLoginReward             int
	DailyLoginCooldownHours      int
	ActivityReward               int
	ActivityCooldownSeconds      int
	DailyCapEarly                int
	DailyCapLate                 int
	PassiveActiveIntervalSeconds int
	PassiveIdleIntervalSeconds   int
	PassiveActiveAmount          int
	PassiveIdleAmount            int
	HopeThreshold                float64
}

type TelemetrySnapshot struct {
	ActivePlayers24h int
	ActivePlayers7d  int
	Telemetry7d      int
}

func LoadOrCalibrateSeason(db *sql.DB, seasonID string) (CalibrationParams, error) {
	if db != nil {
		if existing, ok := loadCalibration(db, seasonID); ok {
			economy.SetCalibration(existing)
			return existing, nil
		}
	}

	telemetry := deriveTelemetrySnapshot(db)
	params := CalibrateSeason(seasonID, seasonStart(), telemetry)
	if db != nil {
		if err := saveCalibration(db, params); err != nil {
			return params, err
		}
	}
	economy.SetCalibration(params)
	return params, nil
}

func CalibrateSeason(seasonID string, start time.Time, telemetry TelemetrySnapshot) CalibrationParams {
	seed := calibrationSeed(seasonID, start, telemetry)
	rng := rand.New(rand.NewSource(seed))

	expected := deriveExpectedParticipants(telemetry)
	participantBias := 0.95 + rng.Float64()*0.1
	adjustedParticipants := int(math.Max(10, math.Round(float64(expected)*participantBias)))

	dailyCapEarly := clampInt(int(30+6*math.Sqrt(float64(adjustedParticipants))), 30, 180)
	dailyCapLate := clampInt(int(float64(dailyCapEarly)*0.35), 10, 70)

	cBase := clampInt(int(float64(dailyCapEarly)*float64(adjustedParticipants)*0.6), 300, 240000)
	p0 := clampInt(int(float64(dailyCapEarly)*0.45), 8, 70)

	alpha := clampFloat(2.4+0.4*math.Log10(float64(adjustedParticipants)+1), 2.4, 5.6)
	beta := clampFloat(2.2+0.25*math.Log10(float64(adjustedParticipants)+1), 2.2, 3.2)

	totalCoins := float64(cBase) * 28 * 0.55
	expectedTotalStars := totalCoins / float64(p0) / 3.0
	sScale := clampFloat(expectedTotalStars/8, 20, 420)
	gScale := clampFloat(float64(cBase)*2.5, 800, 60000)
	gamma := clampFloat(0.06+0.01*math.Log10(float64(adjustedParticipants)+1), 0.06, 0.16)

	dailyLoginReward := clampInt(int(float64(dailyCapEarly)*0.25), 10, 45)
	activityReward := clampInt(int(float64(dailyCapEarly)*0.04), 1, 6)
	activityCooldownSeconds := clampInt(6*60, 300, 720)

	passiveActiveInterval := 90
	passiveIdleInterval := 240
	passiveActiveAmount := clampInt(activityReward-1, 1, 4)
	passiveIdleAmount := 1

	params := CalibrationParams{
		SeasonID:                     seasonID,
		Seed:                         seed,
		P0:                           p0,
		CBase:                        cBase,
		Alpha:                        alpha,
		SScale:                       sScale,
		GScale:                       gScale,
		Beta:                         beta,
		Gamma:                        gamma,
		DailyLoginReward:             dailyLoginReward,
		DailyLoginCooldownHours:      20,
		ActivityReward:               activityReward,
		ActivityCooldownSeconds:      activityCooldownSeconds,
		DailyCapEarly:                dailyCapEarly,
		DailyCapLate:                 dailyCapLate,
		PassiveActiveIntervalSeconds: passiveActiveInterval,
		PassiveIdleIntervalSeconds:   passiveIdleInterval,
		PassiveActiveAmount:          passiveActiveAmount,
		PassiveIdleAmount:            passiveIdleAmount,
		HopeThreshold:                0.22,
	}

	return params
}

func calibrationSeed(seasonID string, start time.Time, telemetry TelemetrySnapshot) int64 {
	key := fmt.Sprintf("%s|%s|%d|%d|%d", seasonID, start.UTC().Format(time.RFC3339), telemetry.ActivePlayers7d, telemetry.ActivePlayers24h, telemetry.Telemetry7d)
	hash := sha256.Sum256([]byte(key))
	return int64(binary.BigEndian.Uint64(hash[:8]))
}

func deriveExpectedParticipants(telemetry TelemetrySnapshot) int {
	base := telemetry.ActivePlayers7d
	if telemetry.Telemetry7d > base {
		base = telemetry.Telemetry7d
	}
	weighted := float64(base)*0.85 + float64(telemetry.ActivePlayers24h)*0.35
	if weighted < 10 {
		weighted = 10
	}
	return int(math.Round(weighted))
}

func deriveTelemetrySnapshot(db *sql.DB) TelemetrySnapshot {
	if db == nil {
		return TelemetrySnapshot{}
	}

	snapshot := TelemetrySnapshot{}
	if err := db.QueryRow(`
		SELECT COUNT(*)
		FROM players
		WHERE last_active_at >= NOW() - INTERVAL '24 hours'
	`).Scan(&snapshot.ActivePlayers24h); err != nil {
		log.Println("calibration: active24h query failed:", err)
	}
	if err := db.QueryRow(`
		SELECT COUNT(*)
		FROM players
		WHERE last_active_at >= NOW() - INTERVAL '7 days'
	`).Scan(&snapshot.ActivePlayers7d); err != nil {
		log.Println("calibration: active7d query failed:", err)
	}
	if err := db.QueryRow(`
		SELECT COUNT(DISTINCT player_id)
		FROM player_telemetry
		WHERE created_at >= NOW() - INTERVAL '7 days'
	`).Scan(&snapshot.Telemetry7d); err != nil {
		log.Println("calibration: telemetry7d query failed:", err)
	}

	return snapshot
}

func loadCalibration(db *sql.DB, seasonID string) (CalibrationParams, bool) {
	row := db.QueryRow(`
		SELECT season_id, seed, p0, c_base, alpha, s_scale, g_scale, beta, gamma,
			daily_login_reward, daily_login_cooldown_hours, activity_reward, activity_cooldown_seconds,
			daily_cap_early, daily_cap_late, passive_active_interval_seconds, passive_idle_interval_seconds,
			passive_active_amount, passive_idle_amount, hope_threshold
		FROM season_calibration
		WHERE season_id = $1
	`, seasonID)

	var params CalibrationParams
	if err := row.Scan(
		&params.SeasonID,
		&params.Seed,
		&params.P0,
		&params.CBase,
		&params.Alpha,
		&params.SScale,
		&params.GScale,
		&params.Beta,
		&params.Gamma,
		&params.DailyLoginReward,
		&params.DailyLoginCooldownHours,
		&params.ActivityReward,
		&params.ActivityCooldownSeconds,
		&params.DailyCapEarly,
		&params.DailyCapLate,
		&params.PassiveActiveIntervalSeconds,
		&params.PassiveIdleIntervalSeconds,
		&params.PassiveActiveAmount,
		&params.PassiveIdleAmount,
		&params.HopeThreshold,
	); err != nil {
		return CalibrationParams{}, false
	}

	return params, true
}

func saveCalibration(db *sql.DB, params CalibrationParams) error {
	_, err := db.Exec(`
		INSERT INTO season_calibration (
			season_id,
			seed,
			p0,
			c_base,
			alpha,
			s_scale,
			g_scale,
			beta,
			gamma,
			daily_login_reward,
			daily_login_cooldown_hours,
			activity_reward,
			activity_cooldown_seconds,
			daily_cap_early,
			daily_cap_late,
			passive_active_interval_seconds,
			passive_idle_interval_seconds,
			passive_active_amount,
			passive_idle_amount,
			hope_threshold,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, NOW())
		ON CONFLICT (season_id) DO NOTHING
	`,
		params.SeasonID,
		params.Seed,
		params.P0,
		params.CBase,
		params.Alpha,
		params.SScale,
		params.GScale,
		params.Beta,
		params.Gamma,
		params.DailyLoginReward,
		params.DailyLoginCooldownHours,
		params.ActivityReward,
		params.ActivityCooldownSeconds,
		params.DailyCapEarly,
		params.DailyCapLate,
		params.PassiveActiveIntervalSeconds,
		params.PassiveIdleIntervalSeconds,
		params.PassiveActiveAmount,
		params.PassiveIdleAmount,
		params.HopeThreshold,
	)
	return err
}

func clampInt(value int, min int, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func clampFloat(value float64, min float64, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
