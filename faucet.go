package main

import (
	"database/sql"
	"time"
)

const (
	FaucetPassive  = "passive"
	FaucetDaily    = "daily"
	FaucetActivity = "activity"
	FaucetLogin    = "login"
	FaucetUBI      = "ubi"
)

func CanAccessFaucetByPriority(faucetType string, available int) bool {
	return available > 0
}

func ThrottleFaucetReward(faucetType string, amount int, available int) int {
	if amount <= 0 || available <= 0 {
		return 0
	}
	if amount > available {
		return available
	}
	return amount
}

func TryDistributeCoinsWithPriority(faucetType string, amount int) (int, bool) {
	available := economy.AvailableCoins()
	if !CanAccessFaucetByPriority(faucetType, available) {
		return 0, false
	}
	adjusted := ThrottleFaucetReward(faucetType, amount, available)
	if adjusted <= 0 {
		return 0, false
	}
	if !economy.TryDistributeCoins(adjusted) {
		return 0, false
	}
	return adjusted, true
}

func CanClaimFaucet(
	db *sql.DB,
	playerID string,
	faucetKey string,
	cooldown time.Duration,
) (bool, time.Duration, error) {
	var lastClaim time.Time

	err := db.QueryRow(`
		SELECT last_claim_at
		FROM player_faucet_claims
		WHERE player_id = $1 AND faucet_key = $2
	`, playerID, faucetKey).Scan(&lastClaim)

	if err == sql.ErrNoRows {
		return true, 0, nil
	}
	if err != nil {
		return false, 0, err
	}

	now := time.Now().UTC()
	next := lastClaim.Add(cooldown)
	if !now.Before(next) {
		return true, 0, nil
	}

	return false, next.Sub(now), nil
}

func RecordFaucetClaim(db *sql.DB, playerID string, faucetKey string) error {
	_, err := db.Exec(`
		INSERT INTO player_faucet_claims (
			player_id,
			faucet_key,
			last_claim_at,
			claim_count
		)
		VALUES ($1, $2, NOW(), 1)
		ON CONFLICT (player_id, faucet_key)
		DO UPDATE SET
			last_claim_at = NOW(),
			claim_count = player_faucet_claims.claim_count + 1
	`, playerID, faucetKey)

	return err
}

type FaucetScaling struct {
	RewardMultiplier   float64
	CooldownMultiplier float64
}

func currentFaucetScaling(now time.Time) FaucetScaling {
	progress := seasonProgress(now)
	reward := 1.6 - (1.0 * progress)
	if reward < 0.6 {
		reward = 0.6
	} else if reward > 1.6 {
		reward = 1.6
	}

	cooldown := 0.55 + (1.15 * progress)
	if cooldown < 0.5 {
		cooldown = 0.5
	} else if cooldown > 1.7 {
		cooldown = 1.7
	}

	return FaucetScaling{
		RewardMultiplier:   reward,
		CooldownMultiplier: cooldown,
	}
}

func applyFaucetRewardScaling(reward int, multiplier float64) int {
	if reward <= 0 {
		return reward
	}
	adjusted := int(float64(reward)*multiplier + 0.9999)
	if adjusted < 1 {
		return 1
	}
	return adjusted
}

func applyFaucetCooldownScaling(cooldown time.Duration, multiplier float64) time.Duration {
	if cooldown <= 0 {
		return cooldown
	}
	adjusted := time.Duration(float64(cooldown) * multiplier)
	if adjusted < time.Second {
		return time.Second
	}
	return adjusted
}

// DistributeUniversalBasicIncome grants dynamic income based on player activity.
// Base UBI is 1 microcoin, but active players earn up to 10x more through activity warmup.
// UBI is foundation income: always-on, emission-backed, non-negotiable.
// Failures per-player do not block UBI for other players.
func DistributeUniversalBasicIncome(db *sql.DB, now time.Time) (ubiCount int, ubiTotal int, poolExhausted bool) {
	const baseUBIPerTick = 1 // 1 microcoin = 0.001 coins per tick (minimum)

	if db == nil {
		return 0, 0, false
	}

	seasonID := currentSeasonID()
	if seasonID == "" {
		return 0, 0, false
	}

	// Query all players with their activity state
	rows, err := db.Query(`
		SELECT player_id, last_active_at, activity_warmup_level, 
		       activity_warmup_updated_at, recent_activity_seconds
		FROM players
		ORDER BY player_id
	`)
	if err != nil {
		return 0, 0, false
	}
	defer rows.Close()

	granted := 0
	total := 0
	activityWindow := ActiveActivityWindow()

	for rows.Next() {
		var playerID string
		var lastActive time.Time
		var warmupLevel float64
		var warmupUpdated time.Time
		var recentActivitySeconds int64

		if err := rows.Scan(&playerID, &lastActive, &warmupLevel, &warmupUpdated, &recentActivitySeconds); err != nil {
			continue
		}

		// Update warmup level based on activity
		newWarmup, newRecentActivity := UpdateActivityWarmup(
			lastActive, warmupLevel, warmupUpdated, recentActivitySeconds, now, activityWindow)

		// Calculate dynamic UBI based on warmup
		ubiAmount := CalculateDynamicUBI(baseUBIPerTick, newWarmup)

		// Check pool availability before attempting grant
		available := economy.AvailableCoins()
		if available < ubiAmount {
			poolExhausted = true
			break
		}

		// Update warmup state in database
		_, err := db.Exec(`
			UPDATE players 
			SET activity_warmup_level = $1, 
			    activity_warmup_updated_at = $2,
			    recent_activity_seconds = $3
			WHERE player_id = $4
		`, newWarmup, now, newRecentActivity, playerID)
		if err != nil {
			log.Println("Failed to update warmup for", playerID, ":", err)
		}

		// Grant dynamic UBI with no daily cap (foundation income)
		grantedAmount, err := GrantCoinsNoCap(db, playerID, ubiAmount, now, FaucetUBI, nil)
		if err != nil {
			continue
		}
		if grantedAmount > 0 {
			granted++
			total += grantedAmount
		}
	}

	if featureFlags.Telemetry && (granted > 0 || poolExhausted) {
		snapshot := economy.InvariantSnapshot()
		emitServerTelemetry(db, nil, "", "ubi_tick", map[string]interface{}{
			"seasonId":                 seasonID,
			"baseUBIPerTick":           baseUBIPerTick,
			"playersGranted":           granted,
			"totalGranted":             total,
			"poolExhausted":            poolExhausted,
			"availableCoins":           snapshot.AvailableCoins,
			"activeCoinsInCirculation": snapshot.ActiveCoinsInCirculation,
			"activePlayers":            snapshot.ActivePlayers,
			"totalCoinsInCirculation":  snapshot.TotalCoinsInCirculation,
			"dynamicUBIEnabled":        true,
			"maxWarmupMultiplier":      maxWarmupMultiplier,
		})
	}

	return granted, total, poolExhausted
}

// Activity Warmup Constants
const (
	// Time to reach maximum warmup multiplier through sustained activity
	warmupDurationSeconds = 30 * 60 // 30 minutes of sustained activity

	// Maximum multiplier for fully warmed-up players
	maxWarmupMultiplier = 10.0 // 10x base UBI at max warmup

	// Decay rate: warmup decreases when idle
	// Decay is slower if player was recently very active
	warmupDecayBaseRate = 0.002 // base decay per tick (60 seconds)
)

// UpdateActivityWarmup calculates new warmup level based on player activity.
// Returns (newWarmupLevel, newRecentActivitySeconds)
func UpdateActivityWarmup(
	lastActive time.Time,
	currentWarmup float64,
	warmupUpdated time.Time,
	recentActivitySeconds int64,
	now time.Time,
	activityWindow time.Duration,
) (float64, int64) {
	ticksElapsed := now.Sub(warmupUpdated).Seconds() / 60.0 // ticks are 60 seconds
	if ticksElapsed < 0.1 {
		// Already updated this tick
		return currentWarmup, recentActivitySeconds
	}

	inactiveFor := now.Sub(lastActive)
	isActive := inactiveFor <= activityWindow

	newWarmup := currentWarmup
	newRecentActivity := recentActivitySeconds

	if isActive {
		// Player is active: increase warmup
		// Warmup increases by (1.0 / warmupDurationTicks) per tick
		warmupDurationTicks := float64(warmupDurationSeconds) / 60.0
		warmupIncreasePerTick := 1.0 / warmupDurationTicks

		newWarmup += warmupIncreasePerTick * ticksElapsed
		if newWarmup > 1.0 {
			newWarmup = 1.0
		}

		// Track recent activity time (for decay calculation)
		newRecentActivity += int64(ticksElapsed * 60)
		// Cap at 2x warmup duration to prevent overflow
		if newRecentActivity > warmupDurationSeconds*2 {
			newRecentActivity = warmupDurationSeconds * 2
		}
	} else {
		// Player is idle: decrease warmup
		// Decay rate is inversely proportional to recent activity
		// More recent activity = slower decay
		activityRatio := float64(newRecentActivity) / float64(warmupDurationSeconds)
		if activityRatio > 1.0 {
			activityRatio = 1.0
		}

		// Decay slower if player was recently very active
		decayMultiplier := 1.0 + (activityRatio * 2.0) // 1x to 3x slower decay
		decayRate := warmupDecayBaseRate / decayMultiplier

		newWarmup -= decayRate * ticksElapsed
		if newWarmup < 0 {
			newWarmup = 0
		}

		// Decay recent activity time as well
		recentActivityDecay := int64(ticksElapsed * 60 * 0.5) // decay at half rate
		newRecentActivity -= recentActivityDecay
		if newRecentActivity < 0 {
			newRecentActivity = 0
		}
	}

	return newWarmup, newRecentActivity
}

// CalculateDynamicUBI returns the UBI amount based on warmup level.
// Base UBI is 1 microcoin, max is baseUBI * maxWarmupMultiplier.
func CalculateDynamicUBI(baseUBI int, warmupLevel float64) int {
	if warmupLevel < 0 {
		warmupLevel = 0
	}
	if warmupLevel > 1.0 {
		warmupLevel = 1.0
	}

	// Linear scaling from base to max
	multiplier := 1.0 + (warmupLevel * (maxWarmupMultiplier - 1.0))
	amount := int(float64(baseUBI) * multiplier)

	if amount < baseUBI {
		amount = baseUBI
	}

	return amount
}
