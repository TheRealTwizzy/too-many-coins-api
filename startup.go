package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"os"
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
	if strings.ToLower(strings.TrimSpace(os.Getenv("ALPHA_AUTO_ADMIN"))) != "true" {
		return nil
	}

	username := strings.ToLower(strings.TrimSpace(os.Getenv("ALPHA_ADMIN_USERNAME")))
	password := strings.TrimSpace(os.Getenv("ALPHA_ADMIN_PASSWORD"))
	if username == "" || password == "" {
		return errors.New("ALPHA_AUTO_ADMIN requires ALPHA_ADMIN_USERNAME and ALPHA_ADMIN_PASSWORD")
	}
	displayName := strings.TrimSpace(os.Getenv("ALPHA_ADMIN_DISPLAY_NAME"))
	if displayName == "" {
		displayName = username
	}
	email := strings.TrimSpace(os.Getenv("ALPHA_ADMIN_EMAIL"))
	adminKeyEnv := strings.TrimSpace(os.Getenv("ALPHA_ADMIN_KEY"))

	if AdminExists(ctx, db) {
		log.Println("ALPHA-ONLY AUTO ADMIN: skipped (admin already exists)")
		return nil
	}

	var accountID string
	var role string
	var adminKey sql.NullString
	rowErr := db.QueryRowContext(ctx, `
		SELECT account_id, role, admin_key_hash
		FROM accounts
		WHERE username = $1
	`, username).Scan(&accountID, &role, &adminKey)
	if rowErr != nil && rowErr != sql.ErrNoRows {
		return rowErr
	}

	if rowErr == sql.ErrNoRows {
		account, err := createAccount(db, username, password, displayName, email)
		if err != nil {
			return err
		}
		accountID = account.AccountID
		adminKey = sql.NullString{}
	}

	if err := setAccountRole(db, accountID, "admin"); err != nil {
		return err
	}

	if adminKeyEnv == "" && (!adminKey.Valid || adminKey.String == "") {
		generated, err := generateAdminKey()
		if err != nil {
			return err
		}
		adminKeyEnv = generated
		log.Println("ALPHA-ONLY AUTO ADMIN: generated admin key (not printed)")
	}
	if adminKeyEnv != "" {
		if err := setAdminKey(db, accountID, adminKeyEnv); err != nil {
			return err
		}
	}

	log.Println("ALPHA-ONLY AUTO ADMIN: ensured admin account for @" + username)
	return nil
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
