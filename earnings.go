package main

import (
	"database/sql"
	"errors"
	"math"
	"time"
)

var errDailyCapReached = errors.New("daily cap reached")

func seasonProgress(now time.Time) float64 {
	seasonSeconds := seasonLength.Seconds()
	if seasonSeconds <= 0 {
		return 0
	}
	progress := 1 - (seasonEnd().Sub(now).Seconds() / seasonSeconds)
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	return progress
}

func DailyEarnCap(now time.Time) int {
	params := economy.Calibration()
	progress := seasonProgress(now)
	return DailyEarnCapForParams(params, progress)
}

func DailyEarnCapForParams(params CalibrationParams, progress float64) int {
	decay := math.Pow(progress, 1.1)
	cap := float64(params.DailyCapEarly) - (float64(params.DailyCapEarly-params.DailyCapLate) * decay)
	if cap < float64(params.DailyCapLate) {
		cap = float64(params.DailyCapLate)
	}
	return int(cap + 0.5)
}

func resetDailyEarnIfNeeded(db *sql.DB, playerID string, now time.Time) error {
	var lastReset time.Time
	if err := db.QueryRow(`
		SELECT last_earn_reset_at
		FROM players
		WHERE player_id = $1
	`, playerID).Scan(&lastReset); err != nil {
		return err
	}
	if seasonDayIndex(lastReset) == seasonDayIndex(now) {
		return nil
	}
	_, err := db.Exec(`
		UPDATE players
		SET daily_earn_total = 0,
			last_earn_reset_at = $2
		WHERE player_id = $1
	`, playerID, now)
	return err
}

func RemainingDailyCap(db *sql.DB, playerID string, now time.Time) (int, error) {
	if err := resetDailyEarnIfNeeded(db, playerID, now); err != nil {
		return 0, err
	}
	var currentTotal int64
	if err := db.QueryRow(`
		SELECT daily_earn_total
		FROM players
		WHERE player_id = $1
	`, playerID).Scan(&currentTotal); err != nil {
		return 0, err
	}
	cap := DailyEarnCap(now)
	remaining := cap - int(currentTotal)
	if remaining < 0 {
		remaining = 0
	}
	return remaining, nil
}

func seasonDayIndex(t time.Time) int {
	start := seasonStart()
	if t.Before(start) {
		return 0
	}
	return int(t.Sub(start).Hours() / 24)
}

func GrantCoinsWithCap(db *sql.DB, playerID string, amount int, now time.Time) (int, int, error) {
	if amount <= 0 {
		return 0, 0, nil
	}
	if err := resetDailyEarnIfNeeded(db, playerID, now); err != nil {
		return 0, 0, err
	}

	var currentTotal int64
	if err := db.QueryRow(`
		SELECT daily_earn_total
		FROM players
		WHERE player_id = $1
	`, playerID).Scan(&currentTotal); err != nil {
		return 0, 0, err
	}

	cap := DailyEarnCap(now)
	remaining := cap - int(currentTotal)
	if remaining <= 0 {
		return 0, 0, errDailyCapReached
	}

	grant := amount
	if grant > remaining {
		grant = remaining
	}

	_, err := db.Exec(`
		UPDATE players
		SET coins = coins + $2,
			daily_earn_total = daily_earn_total + $2,
			last_coin_grant_at = $3
		WHERE player_id = $1
	`, playerID, grant, now)
	if err != nil {
		return 0, remaining, err
	}

	return grant, remaining - grant, nil
}
