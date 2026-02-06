package main

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"
)

// =======================
// TOKEN / KEY HELPERS
// =======================

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func generateAdminKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// =======================
// ROLE NORMALIZATION
// =======================

func baseRoleFromRaw(role string) string {
	role = strings.ToLower(strings.TrimSpace(role))
	role = strings.TrimPrefix(role, "frozen:")
	if role == "frozen" {
		return "user"
	}
	return normalizeRole(role)
}

// =======================
// ADMIN MUTATIONS
// =======================

func setAccountRoleByUsername(db *sql.DB, username string, role string) error {
	role = normalizeRole(role)
	_, err := db.Exec(`
		UPDATE accounts
		SET role = $2
		WHERE username = $1
	`, strings.ToLower(username), role)
	return err
}

func setAdminKeyByUsername(db *sql.DB, username string, key string) error {
	hash, err := hashPassword(key)
	if err != nil {
		return err
	}
	_, err = db.Exec(`
		UPDATE accounts
		SET admin_key_hash = $2
		WHERE username = $1
	`, strings.ToLower(username), hash)
	return err
}

// =======================
// PASSWORD VERIFY (SHARED)
// =======================

// =======================
// PASSWORD RESET HELPERS
// =======================

func lookupAccountForReset(db *sql.DB, identifier string) (*Account, error) {
	identifier = strings.TrimSpace(strings.ToLower(identifier))
	if identifier == "" {
		return nil, errors.New("INVALID_REQUEST")
	}

	var account Account
	var email sql.NullString
<<<<<<< HEAD
	var playerID sql.NullString
=======
>>>>>>> a7f569c (Refactor authentication flow and database schema for Phase 0)

	err := db.QueryRow(`
		SELECT account_id, username, display_name, player_id, email
		FROM accounts
		WHERE username = $1 OR email = $1
		LIMIT 1
	`, identifier).Scan(
		&account.AccountID,
		&account.Username,
		&account.DisplayName,
<<<<<<< HEAD
		&playerID,
=======
		&account.PlayerID,
>>>>>>> a7f569c (Refactor authentication flow and database schema for Phase 0)
		&email,
	)
	if err != nil {
		return nil, err
	}

	if email.Valid {
		account.Email = email.String
	}
<<<<<<< HEAD
	if playerID.Valid {
		account.PlayerID = playerID.String
	}
=======
>>>>>>> a7f569c (Refactor authentication flow and database schema for Phase 0)

	return &account, nil
}

func createPasswordResetToken(db *sql.DB, accountID string) (string, error) {
	resetID, err := randomToken(16)
	if err != nil {
		return "", err
	}
	token, err := randomToken(32)
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().UTC().Add(1 * time.Hour)

	_, err = db.Exec(`
		INSERT INTO password_resets (
			reset_id,
			account_id,
			token_hash,
			expires_at,
			created_at
		)
		VALUES ($1, $2, $3, $4, NOW())
	`, resetID, accountID, hashToken(token), expiresAt)
	if err != nil {
		return "", err
	}

	return token, nil
}

func resetPasswordWithToken(db *sql.DB, token string, newPassword string) error {
	if len(newPassword) < 8 || len(newPassword) > 128 {
		return errors.New("INVALID_PASSWORD")
	}
	if token == "" {
		return errors.New("INVALID_TOKEN")
	}

	hash := hashToken(token)

	var accountID string
	var expiresAt time.Time
	var usedAt sql.NullTime
	var role string
	var mustChangePassword bool

	err := db.QueryRow(`
		SELECT pr.account_id, pr.expires_at, pr.used_at,
		       a.role, a.must_change_password
		FROM password_resets pr
		JOIN accounts a ON a.account_id = pr.account_id
		WHERE pr.token_hash = $1
		ORDER BY pr.created_at DESC
		LIMIT 1
	`, hash).Scan(
		&accountID,
		&expiresAt,
		&usedAt,
		&role,
		&mustChangePassword,
	)
	if err == sql.ErrNoRows {
		return errors.New("INVALID_TOKEN")
	}
	if err != nil {
		return err
	}

	if normalizeRole(role) == "admin" && mustChangePassword {
		return errors.New("ADMIN_BOOTSTRAP_REQUIRED")
	}

	if usedAt.Valid {
		return errors.New("TOKEN_USED")
	}
	if time.Now().UTC().After(expiresAt) {
		return errors.New("TOKEN_EXPIRED")
	}

	passwordHash, err := hashPassword(newPassword)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		UPDATE accounts
		SET password_hash = $2
		WHERE account_id = $1
	`, accountID, passwordHash)
	if err != nil {
		return err
	}

	_, _ = db.Exec(`
		UPDATE password_resets
		SET used_at = NOW()
		WHERE token_hash = $1
	`, hash)

	return nil
}

// =======================
// SAFETY GUARD
// =======================

var errAdminOnly = errors.New("ADMIN_ONLY")
