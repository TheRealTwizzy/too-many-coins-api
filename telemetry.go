package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

type TelemetryEventRequest struct {
	EventType string          `json:"eventType"`
	Payload   json.RawMessage `json:"payload"`
}

func telemetryHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if !featureFlags.Telemetry {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		var req TelemetryEventRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if req.EventType == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		account, _, _ := getSessionAccount(db, r)
		playerID := ""
		if account != nil {
			playerID = account.PlayerID
		}

		_, _ = db.Exec(`
			INSERT INTO player_telemetry (account_id, player_id, event_type, payload, created_at)
			VALUES ($1, $2, $3, $4, NOW())
		`, nullableAccountID(account), playerID, req.EventType, req.Payload)

		w.WriteHeader(http.StatusNoContent)
	}
}

func nullableAccountID(account *Account) interface{} {
	if account == nil {
		return nil
	}
	return account.AccountID
}
