package main

import (
	"database/sql"
	"strings"
	"time"
)

func createNotification(db *sql.DB, targetRole string, accountID string, message string, level string, expiresAt *time.Time) error {
	targetRole = strings.ToLower(strings.TrimSpace(targetRole))
	if targetRole == "" {
		targetRole = "all"
	}
	level = strings.ToLower(strings.TrimSpace(level))
	if level == "" {
		level = "info"
	}
	var expires sql.NullTime
	if expiresAt != nil {
		expires = sql.NullTime{Time: *expiresAt, Valid: true}
	}
	_, err := db.Exec(`
		INSERT INTO notifications (target_role, account_id, message, level, created_at, expires_at)
		VALUES ($1, $2, $3, $4, NOW(), $5)
	`, targetRole, strings.TrimSpace(accountID), message, level, expires)
	return err
}
