package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

const startupAdvisoryLockID int64 = 824173921

var startupLockConn *sql.Conn

func acquireStartupLock(ctx context.Context, db *sql.DB) (*sql.Conn, bool, error) {
	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, false, err
	}
	var acquired bool
	if err := conn.QueryRowContext(ctx, `SELECT pg_try_advisory_lock($1)`, startupAdvisoryLockID).Scan(&acquired); err != nil {
		_ = conn.Close()
		return nil, false, err
	}
	if !acquired {
		_ = conn.Close()
		return nil, false, nil
	}
	return conn, true, nil
}

func ensureAlphaAdmin(ctx context.Context, db *sql.DB) error {
	if CurrentPhase() != PhaseAlpha {
		return nil
	}

	const username = "alpha-admin"
	const displayName = "Alpha Admin"
	const email = ""

	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bootstrapComplete := false
	var bootstrapValue string
	if err := tx.QueryRowContext(ctx, `
		SELECT value
		FROM global_settings
		WHERE key = 'admin_bootstrap_complete'
		FOR UPDATE
	`).Scan(&bootstrapValue); err == nil {
		bootstrapComplete = strings.ToLower(strings.TrimSpace(bootstrapValue)) == "true"
	} else if err != sql.ErrNoRows {
		return err
	}

	adminCount := 0
	var adminAccountID string
	var adminUsername string
	var adminMustChange bool
	var adminPlayerID sql.NullString
	rows, err := tx.QueryContext(ctx, `
		SELECT account_id, username, must_change_password, player_id
		FROM accounts
		WHERE role IN ('admin', 'frozen:admin')
		ORDER BY created_at ASC
		LIMIT 2
		FOR UPDATE
	`)
	if err != nil {
		return err
	}
	for rows.Next() {
		adminCount++
		if adminCount == 1 {
			if err := rows.Scan(&adminAccountID, &adminUsername, &adminMustChange, &adminPlayerID); err != nil {
				_ = rows.Close()
				return err
			}
		}
	}
	_ = rows.Close()
	if adminCount > 1 {
		return errors.New("multiple admin accounts exist; refuse to start")
	}
	if adminCount == 1 {
		if adminPlayerID.Valid && strings.TrimSpace(adminPlayerID.String) != "" {
			if err := scrubAdminPlayerArtifactsTx(ctx, tx, adminAccountID, adminPlayerID.String); err != nil {
				return err
			}
		}
		if bootstrapComplete && adminMustChange {
			return errors.New("bootstrap sealed but admin still locked; refuse to start")
		}
		if !bootstrapComplete && !adminMustChange {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO global_settings (key, value, updated_at)
				VALUES ('admin_bootstrap_complete', 'true', NOW())
				ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()
			`); err != nil {
				return err
			}
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		if adminMustChange {
			log.Printf("Alpha admin bootstrap: admin %s locked, awaiting claim", adminUsername)
		} else {
			log.Printf("Alpha admin bootstrap: admin %s already active", adminUsername)
		}
		return nil
	}
	if bootstrapComplete {
		return errors.New("bootstrap sealed but no admin exists; refuse to start")
	}

	var existingAccountID string
	if err := tx.QueryRowContext(ctx, `
		SELECT account_id
		FROM accounts
		WHERE username = $1
		LIMIT 1
		FOR UPDATE
	`, username).Scan(&existingAccountID); err == nil {
		return errors.New("bootstrap admin username already exists without admin role")
	} else if err != sql.ErrNoRows {
		return err
	}

	accountID, err := randomToken(16)
	if err != nil {
		return err
	}
	bootstrapPassword, err := randomToken(32)
	if err != nil {
		return err
	}
	passwordHash, err := hashPassword(bootstrapPassword)
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO accounts (
			account_id,
			username,
			password_hash,
			display_name,
			email,
			role,
			must_change_password,
			created_at,
			last_login_at
		)
		VALUES ($1, $2, $3, $4, $5, 'admin', TRUE, NOW(), NOW())
	`, accountID, username, passwordHash, displayName, email); err != nil {
		return err
	}

	bootstrapDetails := map[string]interface{}{
		"username":    username,
		"displayName": displayName,
	}
	payload, err := json.Marshal(bootstrapDetails)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO admin_audit_log (admin_account_id, action_type, scope_type, scope_id, reason, details, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`, accountID, "auto_admin_bootstrap", "account", accountID, "bootstrap", string(payload)); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO global_settings (key, value, updated_at)
		VALUES ('admin_bootstrap_complete', 'false', NOW())
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()
	`); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Println("Alpha admin bootstrap: created alpha-admin (locked)")
	return nil
}

func scrubAdminPlayerArtifactsTx(ctx context.Context, tx *sql.Tx, accountID string, playerID string) error {
	if strings.TrimSpace(playerID) == "" {
		return nil
	}
	deleteIfExists := func(table string, column string) error {
		var reg sql.NullString
		if err := tx.QueryRowContext(ctx, `SELECT to_regclass($1)`, table).Scan(&reg); err != nil {
			return err
		}
		if !reg.Valid {
			return nil
		}
		_, err := tx.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE %s = $1", table, column), playerID)
		return err
	}

	if err := deleteIfExists("coin_earning_log", "player_id"); err != nil {
		return err
	}
	if err := deleteIfExists("star_purchase_log", "player_id"); err != nil {
		return err
	}
	if err := deleteIfExists("player_faucet_claims", "player_id"); err != nil {
		return err
	}
	if err := deleteIfExists("player_star_variants", "player_id"); err != nil {
		return err
	}
	if err := deleteIfExists("player_boosts", "player_id"); err != nil {
		return err
	}
	if err := deleteIfExists("player_ip_associations", "player_id"); err != nil {
		return err
	}
	if err := deleteIfExists("player_telemetry", "player_id"); err != nil {
		return err
	}
	if err := deleteIfExists("bug_reports", "player_id"); err != nil {
		return err
	}
	if err := deleteIfExists("tsa_cinder_sigils", "owner_player_id"); err != nil {
		return err
	}
	if err := deleteIfExists("tsa_mint_log", "buyer_player_id"); err != nil {
		return err
	}
	if err := deleteIfExists("tsa_trade_log", "seller_player_id"); err != nil {
		return err
	}
	if err := deleteIfExists("tsa_trade_log", "buyer_player_id"); err != nil {
		return err
	}
	if err := deleteIfExists("tsa_activation_log", "player_id"); err != nil {
		return err
	}
	if err := deleteIfExists("players", "player_id"); err != nil {
		return err
	}

	_, err := tx.ExecContext(ctx, `UPDATE accounts SET player_id = NULL WHERE account_id = $1`, accountID)
	return err
}

func ensureActiveSeason(ctx context.Context, db *sql.DB) error {
	seasonID := currentSeasonID()
	_, err := db.ExecContext(ctx, `
		INSERT INTO season_economy (
			season_id,
			global_coin_pool,
			global_stars_purchased,
			coins_distributed,
			emission_remainder,
			market_pressure,
			price_floor,
			last_updated
		)
		VALUES ($1, 0, 0, 0, 0, 1.0, 0, NOW())
		ON CONFLICT (season_id) DO NOTHING
	`, seasonID)
	return err
}

func updateTickHeartbeat(db *sql.DB, now time.Time) {
	_, err := db.Exec(`
		INSERT INTO global_settings (key, value, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()
	`, "tick_last_utc", now.UTC().Format(time.RFC3339))
	if err != nil {
		log.Println("tick heartbeat update failed:", err)
	}
}

func claimTick(db *sql.DB, now time.Time) bool {
	value := now.UTC().Format(time.RFC3339)
	result, err := db.Exec(`
		INSERT INTO global_settings (key, value, updated_at)
		VALUES ('tick_last_utc', $1, NOW())
		ON CONFLICT (key) DO UPDATE
		SET value = EXCLUDED.value, updated_at = NOW()
		WHERE global_settings.value < EXCLUDED.value
	`, value)
	if err != nil {
		log.Println("tick claim failed:", err)
		return false
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false
	}
	return rows > 0
}

func readTickHeartbeat(ctx context.Context, db *sql.DB) (time.Time, error) {
	var value string
	if err := db.QueryRowContext(ctx, `
		SELECT value
		FROM global_settings
		WHERE key = 'tick_last_utc'
	`).Scan(&value); err != nil {
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339, value)
}
