package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type wipeTable struct {
	name string
}

func main() {
	if !confirmWipe() {
		log.Println("SERVER WIPE ABORTED: missing confirmation")
		os.Exit(2)
	}

	dbURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("failed to open database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("failed to ping database:", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	log.Println("WARNING: SERVER WIPE INITIATED")
	log.Println("This will permanently delete all gameplay data.")

	if err := runWipe(ctx, db); err != nil {
		log.Fatal("wipe failed:", err)
	}

	log.Println("SERVER WIPE COMPLETE")
	log.Println("All gameplay data removed. Admin access must be re-established manually.")
	os.Exit(0)
}

func confirmWipe() bool {
	allow := strings.TrimSpace(os.Getenv("ALLOW_SERVER_WIPE"))
	confirm := strings.TrimSpace(os.Getenv("CONFIRM_SERVER_WIPE"))
	if allow != "true" {
		return false
	}
	if confirm != "YES_I_UNDERSTAND" {
		return false
	}
	if len(os.Args) < 2 {
		return false
	}
	return os.Args[1] == "--i-understand"
}

func runWipe(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	tables := []wipeTable{
		{name: "player_telemetry"},
		{name: "star_purchase_log"},
		{name: "player_star_variants"},
		{name: "player_boosts"},
		{name: "player_faucet_claims"},
		{name: "notification_reads"},
		{name: "notifications"},
		{name: "password_resets"},
		{name: "refresh_tokens"},
		{name: "sessions"},
		{name: "ip_whitelist_requests"},
		{name: "ip_whitelist"},
		{name: "player_ip_associations"},
		{name: "season_final_rankings"},
		{name: "season_end_snapshots"},
		{name: "season_economy"},
		{name: "players"},
		{name: "accounts"},
		{name: "global_settings"},
	}

	log.Println("Wiping tables:")
	for _, table := range tables {
		count, err := countRows(ctx, tx, table.name)
		if err != nil {
			return fmt.Errorf("count %s: %w", table.name, err)
		}
		if _, err := tx.ExecContext(ctx, "DELETE FROM "+table.name); err != nil {
			return fmt.Errorf("delete %s: %w", table.name, err)
		}
		log.Printf("- %s: %d rows deleted", table.name, count)
	}

	if err := resetSequences(ctx, tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Println("All sequences reset.")
	return nil
}

func countRows(ctx context.Context, tx *sql.Tx, table string) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
	var count int64
	if err := tx.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func resetSequences(ctx context.Context, tx *sql.Tx) error {
	sequences := []string{
		"notifications_id_seq",
		"star_purchase_log_id_seq",
		"player_telemetry_id_seq",
		"refresh_tokens_id_seq",
	}
	for _, seq := range sequences {
		if _, err := tx.ExecContext(ctx, "ALTER SEQUENCE "+seq+" RESTART WITH 1"); err != nil {
			return fmt.Errorf("reset sequence %s: %w", seq, err)
		}
	}
	return nil
}
