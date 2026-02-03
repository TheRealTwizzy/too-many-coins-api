package main

import (
	"database/sql"
	"log"
	"math"
	"time"
)

func updateMarketPressure(db *sql.DB, now time.Time) {
	seasonID := currentSeasonID()
	var last24h int
	var last7d int
	if err := db.QueryRow(`
		SELECT COUNT(*)
		FROM star_purchase_log
		WHERE season_id = $1 AND created_at >= $2
	`, seasonID, now.Add(-24*time.Hour)).Scan(&last24h); err != nil {
		log.Println("market pressure: last24h query failed:", err)
		return
	}
	if err := db.QueryRow(`
		SELECT COUNT(*)
		FROM star_purchase_log
		WHERE season_id = $1 AND created_at >= $2
	`, seasonID, now.Add(-7*24*time.Hour)).Scan(&last7d); err != nil {
		log.Println("market pressure: last7d query failed:", err)
		return
	}

	longTermDaily := float64(last7d) / 7.0
	if longTermDaily < 1 {
		longTermDaily = 1
	}
	ratio := float64(last24h) / longTermDaily

	desired := 1.0
	if ratio >= 1 {
		desired = 1 + math.Min(0.8, 0.25*(ratio-1))
	} else {
		desired = 1 - math.Min(0.3, 0.15*(1-ratio))
	}

	maxDeltaPerHour := 0.02
	maxDelta := maxDeltaPerHour / 60
	economy.UpdateMarketPressure(desired, maxDelta)
}
