package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"math"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"
)

type AbuseEnforcement struct {
	Score                float64
	Severity             int
	PriceMultiplier      float64
	MaxBulkQty           int
	EarnMultiplier       float64
	CooldownJitterFactor float64
}

type AbuseSignal struct {
	PlayerID  string
	EventType string
	Delta     float64
	Severity  int
	Details   map[string]interface{}
}

var (
	abuseRandMu sync.Mutex
	abuseRand   = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func abuseIncludeBots() bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv("ABUSE_INCLUDE_BOTS")))
	if value == "" {
		return false
	}
	return value == "true" || value == "1" || value == "yes" || value == "on"
}

func abuseSeverityForScore(score float64) int {
	switch {
	case score >= 45:
		return 3
	case score >= 25:
		return 2
	case score >= 10:
		return 1
	default:
		return 0
	}
}

func abuseDecayRateForSeverity(severity int) float64 {
	switch severity {
	case 3:
		return 0.15
	case 2:
		return 0.3
	case 1:
		return 0.6
	default:
		return 1.0
	}
}

func abusePersistentDuration(severity int) time.Duration {
	switch severity {
	case 3:
		return 7 * 24 * time.Hour
	case 2:
		return 72 * time.Hour
	default:
		return 0
	}
}

func abuseEnforcementForScore(score float64, severity int, baseMaxBulk int) AbuseEnforcement {
	enforcement := AbuseEnforcement{
		Score:                score,
		Severity:             severity,
		PriceMultiplier:      1.0,
		MaxBulkQty:           baseMaxBulk,
		EarnMultiplier:       1.0,
		CooldownJitterFactor: 0,
	}

	switch severity {
	case 1:
		enforcement.PriceMultiplier = 1.05
		enforcement.MaxBulkQty = minInt(baseMaxBulk, 4)
		enforcement.EarnMultiplier = 0.9
		enforcement.CooldownJitterFactor = 0.1
	case 2:
		enforcement.PriceMultiplier = 1.15
		enforcement.MaxBulkQty = minInt(baseMaxBulk, 3)
		enforcement.EarnMultiplier = 0.75
		enforcement.CooldownJitterFactor = 0.25
	case 3:
		enforcement.PriceMultiplier = 1.3
		enforcement.MaxBulkQty = minInt(baseMaxBulk, 2)
		enforcement.EarnMultiplier = 0.6
		enforcement.CooldownJitterFactor = 0.5
	}

	if enforcement.MaxBulkQty < 1 {
		enforcement.MaxBulkQty = 1
	}

	return enforcement
}

func abuseCooldownJitter(base time.Duration, factor float64) time.Duration {
	if base <= 0 || factor <= 0 {
		return 0
	}
	maxJitter := time.Duration(float64(base) * factor)
	if maxJitter <= 0 {
		return 0
	}
	if maxJitter > 5*time.Minute {
		maxJitter = 5 * time.Minute
	}
	abuseRandMu.Lock()
	n := abuseRand.Int63n(int64(maxJitter) + 1)
	abuseRandMu.Unlock()
	return time.Duration(n)
}

func abuseEffectiveEnforcement(db *sql.DB, playerID string, baseMaxBulk int) AbuseEnforcement {
	if playerID == "" {
		return abuseEnforcementForScore(0, 0, baseMaxBulk)
	}

	isBot, _, err := getPlayerBotInfo(db, playerID)
	if err == nil && isBot && !abuseIncludeBots() {
		return abuseEnforcementForScore(0, 0, baseMaxBulk)
	}

	seasonScore, seasonSeverity, err := getPlayerAbuseScore(db, playerID, currentSeasonID())
	if err != nil {
		log.Println("abuse: load player state failed:", err)
		return abuseEnforcementForScore(0, 0, baseMaxBulk)
	}

	accountID, err := accountIDForPlayer(db, playerID)
	if err != nil {
		log.Println("abuse: load account id failed:", err)
		return abuseEnforcementForScore(seasonScore, seasonSeverity, baseMaxBulk)
	}

	accountScore, accountSeverity, err := getAccountAbuseScore(db, accountID)
	if err != nil {
		log.Println("abuse: load account state failed:", err)
		return abuseEnforcementForScore(seasonScore, seasonSeverity, baseMaxBulk)
	}

	combinedScore := math.Max(seasonScore, accountScore*0.75)
	combinedSeverity := seasonSeverity
	if accountSeverity > combinedSeverity {
		combinedSeverity = accountSeverity
	}
	derivedSeverity := abuseSeverityForScore(combinedScore)
	if derivedSeverity > combinedSeverity {
		combinedSeverity = derivedSeverity
	}

	return abuseEnforcementForScore(combinedScore, combinedSeverity, baseMaxBulk)
}

func abuseAdjustedReward(base int, multiplier float64) int {
	if base <= 0 {
		return base
	}
	adjusted := int(math.Floor(float64(base) * multiplier))
	if adjusted < 1 {
		adjusted = 1
	}
	return adjusted
}

func abuseAdjustedPrice(base int, multiplier float64) int {
	if base <= 0 {
		return base
	}
	final := int(float64(base)*multiplier + 0.9999)
	if final < 1 {
		final = 1
	}
	return final
}

func accountIDForPlayer(db *sql.DB, playerID string) (string, error) {
	var accountID string
	if err := db.QueryRow(`
		SELECT account_id
		FROM accounts
		WHERE player_id = $1
	`, playerID).Scan(&accountID); err != nil {
		return "", err
	}
	return accountID, nil
}

func getPlayerAbuseScore(db *sql.DB, playerID string, seasonID string) (float64, int, error) {
	var score float64
	var severity int
	if err := db.QueryRow(`
		SELECT score, severity
		FROM player_abuse_state
		WHERE player_id = $1 AND season_id = $2
	`, playerID, seasonID).Scan(&score, &severity); err != nil {
		if err == sql.ErrNoRows {
			return 0, 0, nil
		}
		return 0, 0, err
	}
	return score, severity, nil
}

func getAccountAbuseScore(db *sql.DB, accountID string) (float64, int, error) {
	var score float64
	var severity int
	if err := db.QueryRow(`
		SELECT score, severity
		FROM account_abuse_reputation
		WHERE account_id = $1
	`, accountID).Scan(&score, &severity); err != nil {
		if err == sql.ErrNoRows {
			return 0, 0, nil
		}
		return 0, 0, err
	}
	return score, severity, nil
}

func applyPlayerAbuseDelta(db *sql.DB, playerID string, seasonID string, delta float64, now time.Time) (float64, int, error) {
	var score float64
	var severity int
	var lastSignal sql.NullTime
	var persistentUntil sql.NullTime
	var lastDecay time.Time

	err := db.QueryRow(`
		SELECT score, severity, last_signal_at, persistent_until, last_decay_at
		FROM player_abuse_state
		WHERE player_id = $1 AND season_id = $2
	`, playerID, seasonID).Scan(&score, &severity, &lastSignal, &persistentUntil, &lastDecay)
	if err != nil && err != sql.ErrNoRows {
		return 0, 0, err
	}

	if err == sql.ErrNoRows {
		score = 0
		severity = 0
		lastDecay = now
	}

	newScore := score + delta
	if newScore < 0 {
		newScore = 0
	}
	newSeverity := abuseSeverityForScore(newScore)

	if newSeverity >= 2 {
		if lastSignal.Valid && now.Sub(lastSignal.Time) <= 6*time.Hour {
			persistence := abusePersistentDuration(newSeverity)
			if persistence > 0 {
				candidate := now.Add(persistence)
				if !persistentUntil.Valid || candidate.After(persistentUntil.Time) {
					persistentUntil = sql.NullTime{Time: candidate, Valid: true}
				}
			}
		}
	}

	_, err = db.Exec(`
		INSERT INTO player_abuse_state (
			player_id, season_id, score, severity, last_signal_at, last_decay_at, persistent_until, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (player_id, season_id) DO UPDATE
		SET score = EXCLUDED.score,
			severity = EXCLUDED.severity,
			last_signal_at = EXCLUDED.last_signal_at,
			last_decay_at = EXCLUDED.last_decay_at,
			persistent_until = EXCLUDED.persistent_until,
			updated_at = NOW()
	`, playerID, seasonID, newScore, newSeverity, now, lastDecay, persistentUntil)
	if err != nil {
		return 0, 0, err
	}

	return newScore, newSeverity, nil
}

func applyAccountAbuseDelta(db *sql.DB, accountID string, delta float64, now time.Time) (float64, int, error) {
	var score float64
	var severity int
	var lastSignal sql.NullTime
	var persistentUntil sql.NullTime
	var lastDecay time.Time

	err := db.QueryRow(`
		SELECT score, severity, last_signal_at, persistent_until, last_decay_at
		FROM account_abuse_reputation
		WHERE account_id = $1
	`, accountID).Scan(&score, &severity, &lastSignal, &persistentUntil, &lastDecay)
	if err != nil && err != sql.ErrNoRows {
		return 0, 0, err
	}

	if err == sql.ErrNoRows {
		score = 0
		severity = 0
		lastDecay = now
	}

	newScore := score + delta
	if newScore < 0 {
		newScore = 0
	}
	newSeverity := abuseSeverityForScore(newScore)

	if newSeverity >= 2 {
		if lastSignal.Valid && now.Sub(lastSignal.Time) <= 6*time.Hour {
			persistence := abusePersistentDuration(newSeverity)
			if persistence > 0 {
				candidate := now.Add(persistence)
				if !persistentUntil.Valid || candidate.After(persistentUntil.Time) {
					persistentUntil = sql.NullTime{Time: candidate, Valid: true}
				}
			}
		}
	}

	_, err = db.Exec(`
		INSERT INTO account_abuse_reputation (
			account_id, score, severity, last_signal_at, last_decay_at, persistent_until, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (account_id) DO UPDATE
		SET score = EXCLUDED.score,
			severity = EXCLUDED.severity,
			last_signal_at = EXCLUDED.last_signal_at,
			last_decay_at = EXCLUDED.last_decay_at,
			persistent_until = EXCLUDED.persistent_until,
			updated_at = NOW()
	`, accountID, newScore, newSeverity, now, lastDecay, persistentUntil)
	if err != nil {
		return 0, 0, err
	}

	return newScore, newSeverity, nil
}

func decayPlayerAbuseStates(db *sql.DB, now time.Time) error {
	rows, err := db.Query(`
		SELECT player_id, season_id, score, severity, last_decay_at, persistent_until
		FROM player_abuse_state
		WHERE season_id = $1
	`, currentSeasonID())
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var playerID string
		var seasonID string
		var score float64
		var severity int
		var lastDecay time.Time
		var persistentUntil sql.NullTime
		if err := rows.Scan(&playerID, &seasonID, &score, &severity, &lastDecay, &persistentUntil); err != nil {
			continue
		}

		if persistentUntil.Valid && now.Before(persistentUntil.Time) {
			continue
		}

		elapsedHours := now.Sub(lastDecay).Hours()
		if elapsedHours <= 0 {
			continue
		}

		newScore := score - (abuseDecayRateForSeverity(severity) * elapsedHours)
		if newScore < 0 {
			newScore = 0
		}
		newSeverity := abuseSeverityForScore(newScore)

		_, _ = db.Exec(`
			UPDATE player_abuse_state
			SET score = $1,
				severity = $2,
				last_decay_at = $3,
				updated_at = NOW()
			WHERE player_id = $4 AND season_id = $5
		`, newScore, newSeverity, now, playerID, seasonID)
	}

	return rows.Err()
}

func decayAccountAbuseStates(db *sql.DB, now time.Time) error {
	rows, err := db.Query(`
		SELECT account_id, score, severity, last_decay_at, persistent_until
		FROM account_abuse_reputation
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var accountID string
		var score float64
		var severity int
		var lastDecay time.Time
		var persistentUntil sql.NullTime
		if err := rows.Scan(&accountID, &score, &severity, &lastDecay, &persistentUntil); err != nil {
			continue
		}

		if persistentUntil.Valid && now.Before(persistentUntil.Time) {
			continue
		}

		elapsedHours := now.Sub(lastDecay).Hours()
		if elapsedHours <= 0 {
			continue
		}

		newScore := score - (abuseDecayRateForSeverity(severity) * elapsedHours)
		if newScore < 0 {
			newScore = 0
		}
		newSeverity := abuseSeverityForScore(newScore)

		_, _ = db.Exec(`
			UPDATE account_abuse_reputation
			SET score = $1,
				severity = $2,
				last_decay_at = $3,
				updated_at = NOW()
			WHERE account_id = $4
		`, newScore, newSeverity, now, accountID)
	}

	return rows.Err()
}

func logAbuseEvent(db *sql.DB, signal AbuseSignal, accountID string, seasonID string, now time.Time) {
	payload, _ := json.Marshal(signal.Details)
	_, _ = db.Exec(`
		INSERT INTO abuse_events (
			account_id, player_id, season_id, event_type, severity, score_delta, details, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, nullableString(accountID), signal.PlayerID, seasonID, signal.EventType, signal.Severity, signal.Delta, payload, now)
}

func nullableString(value string) interface{} {
	if value == "" {
		return nil
	}
	return value
}

func UpdateAbuseMonitoring(db *sql.DB, now time.Time) {
	if err := decayPlayerAbuseStates(db, now); err != nil {
		log.Println("abuse: decay player states failed:", err)
	}
	if err := decayAccountAbuseStates(db, now); err != nil {
		log.Println("abuse: decay account states failed:", err)
	}

	signals, err := collectAbuseSignals(db, now)
	if err != nil {
		log.Println("abuse: collect signals failed:", err)
		return
	}

	for _, signal := range signals {
		accountID, _ := accountIDForPlayer(db, signal.PlayerID)
		seasonID := currentSeasonID()

		_, _, err := applyPlayerAbuseDelta(db, signal.PlayerID, seasonID, signal.Delta, now)
		if err != nil {
			log.Println("abuse: apply player delta failed:", err)
			continue
		}

		if signal.Severity >= 2 && accountID != "" {
			_, _, _ = applyAccountAbuseDelta(db, accountID, signal.Delta*0.6, now)
		}

		logAbuseEvent(db, signal, accountID, seasonID, now)
	}
}

func collectAbuseSignals(db *sql.DB, now time.Time) ([]AbuseSignal, error) {
	includeBots := abuseIncludeBots()
	results := make([]AbuseSignal, 0)

	burstSignals, err := signalStarPurchaseBurst(db, now, includeBots)
	if err != nil {
		return results, err
	}
	results = append(results, burstSignals...)

	regularSignals, err := signalRegularPurchaseCadence(db, now, includeBots)
	if err != nil {
		return results, err
	}
	results = append(results, regularSignals...)

	activitySignals, err := signalRegularActivityCadence(db, now, includeBots)
	if err != nil {
		return results, err
	}
	results = append(results, activitySignals...)

	reactionSignals, err := signalTickReaction(db, now, includeBots)
	if err != nil {
		return results, err
	}
	results = append(results, reactionSignals...)

	ipSignals, err := signalIPCluster(db, now, includeBots)
	if err != nil {
		return results, err
	}
	results = append(results, ipSignals...)

	return results, nil
}

func signalStarPurchaseBurst(db *sql.DB, now time.Time, includeBots bool) ([]AbuseSignal, error) {
	rows, err := db.Query(`
		SELECT s.player_id, COUNT(*)
		FROM star_purchase_log s
		JOIN players p ON p.player_id = s.player_id
		WHERE s.created_at >= $1
			AND ($2 = TRUE OR p.is_bot = FALSE)
		GROUP BY s.player_id
		HAVING COUNT(*) >= 6
	`, now.Add(-10*time.Minute), includeBots)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	signals := []AbuseSignal{}
	for rows.Next() {
		var playerID string
		var count int
		if err := rows.Scan(&playerID, &count); err != nil {
			continue
		}
		delta := float64(count-5) * 1.2
		signals = append(signals, AbuseSignal{
			PlayerID:  playerID,
			EventType: "purchase_burst",
			Delta:     delta,
			Severity:  1,
			Details: map[string]interface{}{
				"count":         count,
				"windowMinutes": 10,
			},
		})
	}

	return signals, rows.Err()
}

func signalRegularPurchaseCadence(db *sql.DB, now time.Time, includeBots bool) ([]AbuseSignal, error) {
	rows, err := db.Query(`
		SELECT s.player_id, s.created_at
		FROM star_purchase_log s
		JOIN players p ON p.player_id = s.player_id
		WHERE s.created_at >= $1
			AND ($2 = TRUE OR p.is_bot = FALSE)
		ORDER BY s.player_id, s.created_at ASC
	`, now.Add(-60*time.Minute), includeBots)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byPlayer := map[string][]time.Time{}
	for rows.Next() {
		var playerID string
		var createdAt time.Time
		if err := rows.Scan(&playerID, &createdAt); err != nil {
			continue
		}
		byPlayer[playerID] = append(byPlayer[playerID], createdAt)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	signals := []AbuseSignal{}
	for playerID, times := range byPlayer {
		if len(times) < 6 {
			continue
		}
		intervals := make([]float64, 0, len(times)-1)
		for i := 1; i < len(times); i++ {
			intervals = append(intervals, times[i].Sub(times[i-1]).Seconds())
		}
		mean, stddev := stats(intervals)
		if mean > 0 && mean <= 180 && stddev <= 2.0 {
			signals = append(signals, AbuseSignal{
				PlayerID:  playerID,
				EventType: "purchase_regular_interval",
				Delta:     2.5,
				Severity:  2,
				Details: map[string]interface{}{
					"intervalMeanSeconds": mean,
					"intervalStdSeconds":  stddev,
					"count":               len(times),
				},
			})
		}
	}

	return signals, nil
}

func signalRegularActivityCadence(db *sql.DB, now time.Time, includeBots bool) ([]AbuseSignal, error) {
	rows, err := db.Query(`
		SELECT c.player_id, c.created_at
		FROM coin_earning_log c
		JOIN players p ON p.player_id = c.player_id
		WHERE c.created_at >= $1
			AND c.source_type = 'activity'
			AND ($2 = TRUE OR p.is_bot = FALSE)
		ORDER BY c.player_id, c.created_at ASC
	`, now.Add(-60*time.Minute), includeBots)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byPlayer := map[string][]time.Time{}
	for rows.Next() {
		var playerID string
		var createdAt time.Time
		if err := rows.Scan(&playerID, &createdAt); err != nil {
			continue
		}
		byPlayer[playerID] = append(byPlayer[playerID], createdAt)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	signals := []AbuseSignal{}
	for playerID, times := range byPlayer {
		if len(times) < 6 {
			continue
		}
		intervals := make([]float64, 0, len(times)-1)
		for i := 1; i < len(times); i++ {
			intervals = append(intervals, times[i].Sub(times[i-1]).Seconds())
		}
		mean, stddev := stats(intervals)
		if mean > 0 && mean <= 240 && stddev <= 3.0 {
			signals = append(signals, AbuseSignal{
				PlayerID:  playerID,
				EventType: "activity_regular_interval",
				Delta:     2.0,
				Severity:  1,
				Details: map[string]interface{}{
					"intervalMeanSeconds": mean,
					"intervalStdSeconds":  stddev,
					"count":               len(times),
				},
			})
		}
	}

	return signals, nil
}

func signalTickReaction(db *sql.DB, now time.Time, includeBots bool) ([]AbuseSignal, error) {
	rows, err := db.Query(`
		SELECT s.player_id, COUNT(*)
		FROM star_purchase_log s
		JOIN players p ON p.player_id = s.player_id
		WHERE s.created_at >= $1
			AND ($2 = TRUE OR p.is_bot = FALSE)
			AND (EXTRACT(SECOND FROM s.created_at) <= 2 OR EXTRACT(SECOND FROM s.created_at) >= 58)
		GROUP BY s.player_id
		HAVING COUNT(*) >= 3
	`, now.Add(-30*time.Minute), includeBots)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	signals := []AbuseSignal{}
	for rows.Next() {
		var playerID string
		var count int
		if err := rows.Scan(&playerID, &count); err != nil {
			continue
		}
		signals = append(signals, AbuseSignal{
			PlayerID:  playerID,
			EventType: "tick_reaction_burst",
			Delta:     float64(count) * 0.8,
			Severity:  1,
			Details: map[string]interface{}{
				"count":         count,
				"windowMinutes": 30,
			},
		})
	}

	return signals, rows.Err()
}

func signalIPCluster(db *sql.DB, now time.Time, includeBots bool) ([]AbuseSignal, error) {
	rows, err := db.Query(`
		SELECT p.ip, COUNT(DISTINCT s.player_id)
		FROM star_purchase_log s
		JOIN player_ip_associations p ON p.player_id = s.player_id
		JOIN players pl ON pl.player_id = s.player_id
		WHERE s.created_at >= $1
			AND p.last_seen >= $2
			AND ($3 = TRUE OR pl.is_bot = FALSE)
		GROUP BY p.ip
		HAVING COUNT(DISTINCT s.player_id) >= 3
	`, now.Add(-10*time.Minute), now.Add(-24*time.Hour), includeBots)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ips := []string{}
	counts := map[string]int{}
	for rows.Next() {
		var ip string
		var count int
		if err := rows.Scan(&ip, &count); err != nil {
			continue
		}
		ips = append(ips, ip)
		counts[ip] = count
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(ips) == 0 {
		return []AbuseSignal{}, nil
	}

	signals := []AbuseSignal{}
	for _, ip := range ips {
		playerRows, err := db.Query(`
			SELECT player_id
			FROM player_ip_associations
			WHERE ip = $1
		`, ip)
		if err != nil {
			continue
		}
		for playerRows.Next() {
			var playerID string
			if err := playerRows.Scan(&playerID); err != nil {
				continue
			}
			if !includeBots {
				isBot, _, err := getPlayerBotInfo(db, playerID)
				if err == nil && isBot {
					continue
				}
			}
			signals = append(signals, AbuseSignal{
				PlayerID:  playerID,
				EventType: "ip_cluster_activity",
				Delta:     float64(counts[ip]) * 0.7,
				Severity:  2,
				Details: map[string]interface{}{
					"ip":            ip,
					"activePlayers": counts[ip],
					"windowMinutes": 10,
				},
			})
		}
		playerRows.Close()
	}

	return signals, nil
}

func stats(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0, 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))
	if len(values) < 2 {
		return mean, 0
	}
	variance := 0.0
	for _, v := range values {
		d := v - mean
		variance += d * d
	}
	variance = variance / float64(len(values))
	return mean, math.Sqrt(variance)
}

func abuseMaxBulkQty(db *sql.DB, playerID string, baseMax int) int {
	enforcement := abuseEffectiveEnforcement(db, playerID, baseMax)
	return enforcement.MaxBulkQty
}
