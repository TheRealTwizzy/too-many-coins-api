package main

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
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

func verifyPassword(stored, password string) bool {
	parts := strings.Split(stored, ":")
	if len(parts) != 2 {
		return false
	}
	salt := parts[0]
	expected := parts[1]

	sum := sha256.Sum256([]byte(salt + password))
	computed := base64.RawURLEncoding.EncodeToString(sum[:])

	return subtle.ConstantTimeCompare([]byte(computed), []byte(expected)) == 1
}

// =======================
// SAFETY GUARD
// =======================

var errAdminOnly = errors.New("ADMIN_ONLY")
