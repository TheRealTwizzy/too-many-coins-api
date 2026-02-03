package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type AdminTelemetrySeriesPoint struct {
	Bucket    time.Time `json:"bucket"`
	EventType string    `json:"eventType"`
	Count     int       `json:"count"`
}

type AdminTelemetryResponse struct {
	OK       bool                        `json:"ok"`
	Error    string                      `json:"error,omitempty"`
	Series   []AdminTelemetrySeriesPoint `json:"series,omitempty"`
	Feedback []AdminFeedbackItem         `json:"feedback,omitempty"`
}

type AdminFeedbackItem struct {
	Rating    *int      `json:"rating,omitempty"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"createdAt"`
}

type AdminEconomyResponse struct {
	OK                  bool   `json:"ok"`
	Error               string `json:"error,omitempty"`
	DailyEmissionTarget int    `json:"dailyEmissionTarget,omitempty"`
	FaucetsEnabled      bool   `json:"faucetsEnabled,omitempty"`
	SinksEnabled        bool   `json:"sinksEnabled,omitempty"`
	TelemetryEnabled    bool   `json:"telemetryEnabled,omitempty"`
}

type AdminEconomyUpdateRequest struct {
	DailyEmissionTarget *int  `json:"dailyEmissionTarget,omitempty"`
	FaucetsEnabled      *bool `json:"faucetsEnabled,omitempty"`
	SinksEnabled        *bool `json:"sinksEnabled,omitempty"`
	TelemetryEnabled    *bool `json:"telemetryEnabled,omitempty"`
}

type AdminStarPurchaseLogItem struct {
	ID           int64     `json:"id"`
	AccountID    string    `json:"accountId"`
	PlayerID     string    `json:"playerId"`
	SeasonID     string    `json:"seasonId"`
	PurchaseType string    `json:"purchaseType"`
	Variant      string    `json:"variant,omitempty"`
	PricePaid    int64     `json:"pricePaid"`
	CoinsBefore  int64     `json:"coinsBefore"`
	CoinsAfter   int64     `json:"coinsAfter"`
	StarsBefore  int64     `json:"starsBefore"`
	StarsAfter   int64     `json:"starsAfter"`
	CreatedAt    time.Time `json:"createdAt"`
}

type AdminStarPurchaseLogResponse struct {
	OK    bool                       `json:"ok"`
	Error string                     `json:"error,omitempty"`
	Items []AdminStarPurchaseLogItem `json:"items,omitempty"`
}

func requireAdmin(db *sql.DB, w http.ResponseWriter, r *http.Request) (*Account, bool) {
	account, _, err := getSessionAccount(db, r)
	if err != nil || account == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return nil, false
	}
	if account.Role != "admin" || account.AdminKeyHash == "" {
		w.WriteHeader(http.StatusForbidden)
		return nil, false
	}
	provided := r.Header.Get("X-Admin-Key")
	if provided == "" || !verifyAdminKey(account.AdminKeyHash, provided) {
		w.WriteHeader(http.StatusUnauthorized)
		return nil, false
	}
	return account, true
}

func requireModerator(db *sql.DB, w http.ResponseWriter, r *http.Request) (*Account, bool) {
	account, _, err := getSessionAccount(db, r)
	if err != nil || account == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return nil, false
	}
	if (account.Role != "admin" && account.Role != "moderator") || account.AdminKeyHash == "" {
		w.WriteHeader(http.StatusForbidden)
		return nil, false
	}
	provided := r.Header.Get("X-Admin-Key")
	if provided == "" || !verifyAdminKey(account.AdminKeyHash, provided) {
		w.WriteHeader(http.StatusUnauthorized)
		return nil, false
	}
	return account, true
}

func adminTelemetryHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(db, w, r); !ok {
			return
		}

		rows, err := db.Query(`
			SELECT date_trunc('hour', created_at) AS bucket, event_type, COUNT(*)
			FROM player_telemetry
			WHERE created_at >= NOW() - INTERVAL '48 hours'
			GROUP BY bucket, event_type
			ORDER BY bucket ASC
		`)
		if err != nil {
			json.NewEncoder(w).Encode(AdminTelemetryResponse{OK: false, Error: "INTERNAL_ERROR"})
			return
		}
		defer rows.Close()

		series := []AdminTelemetrySeriesPoint{}
		for rows.Next() {
			var point AdminTelemetrySeriesPoint
			if err := rows.Scan(&point.Bucket, &point.EventType, &point.Count); err != nil {
				continue
			}
			series = append(series, point)
		}

		feedbackRows, err := db.Query(`
			SELECT rating, message, created_at
			FROM player_feedback
			ORDER BY created_at DESC
			LIMIT 25
		`)
		if err != nil {
			json.NewEncoder(w).Encode(AdminTelemetryResponse{OK: false, Error: "INTERNAL_ERROR"})
			return
		}
		defer feedbackRows.Close()

		feedback := []AdminFeedbackItem{}
		for feedbackRows.Next() {
			var rating sql.NullInt64
			var message string
			var created time.Time
			if err := feedbackRows.Scan(&rating, &message, &created); err != nil {
				continue
			}
			item := AdminFeedbackItem{
				Message:   message,
				CreatedAt: created,
			}
			if rating.Valid {
				value := int(rating.Int64)
				item.Rating = &value
			}
			feedback = append(feedback, item)
		}

		json.NewEncoder(w).Encode(AdminTelemetryResponse{
			OK:       true,
			Series:   series,
			Feedback: feedback,
		})
	}
}

func adminEconomyHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(db, w, r); !ok {
			return
		}

		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode(AdminEconomyResponse{
				OK:                  true,
				DailyEmissionTarget: economy.DailyEmissionTarget(),
				FaucetsEnabled:      featureFlags.FaucetsEnabled,
				SinksEnabled:        featureFlags.SinksEnabled,
				TelemetryEnabled:    featureFlags.Telemetry,
			})
			return
		case http.MethodPost:
			var req AdminEconomyUpdateRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				json.NewEncoder(w).Encode(AdminEconomyResponse{OK: false, Error: "INVALID_REQUEST"})
				return
			}
			if req.DailyEmissionTarget != nil {
				value := *req.DailyEmissionTarget
				if value < 0 || value > 10000 {
					json.NewEncoder(w).Encode(AdminEconomyResponse{OK: false, Error: "INVALID_EMISSION_TARGET"})
					return
				}
				economy.SetDailyEmissionTarget(value)
			}
			if req.FaucetsEnabled != nil {
				featureFlags.FaucetsEnabled = *req.FaucetsEnabled
			}
			if req.SinksEnabled != nil {
				featureFlags.SinksEnabled = *req.SinksEnabled
			}
			if req.TelemetryEnabled != nil {
				featureFlags.Telemetry = *req.TelemetryEnabled
			}

			json.NewEncoder(w).Encode(AdminEconomyResponse{
				OK:                  true,
				DailyEmissionTarget: economy.DailyEmissionTarget(),
				FaucetsEnabled:      featureFlags.FaucetsEnabled,
				SinksEnabled:        featureFlags.SinksEnabled,
				TelemetryEnabled:    featureFlags.Telemetry,
			})
			return
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
}

type AdminKeySetRequest struct {
	AdminKey string `json:"adminKey"`
}

type AdminKeySetResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

func adminKeySetHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		setupKey := os.Getenv("ADMIN_SETUP_KEY")
		if setupKey == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		provided := r.Header.Get("X-Admin-Setup")
		if provided != setupKey {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		account, _, err := getSessionAccount(db, r)
		if err != nil || account == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		var req AdminKeySetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.AdminKey == "" {
			json.NewEncoder(w).Encode(AdminKeySetResponse{OK: false, Error: "INVALID_REQUEST"})
			return
		}
		if len(req.AdminKey) < 8 {
			json.NewEncoder(w).Encode(AdminKeySetResponse{OK: false, Error: "WEAK_KEY"})
			return
		}
		if err := setAdminKey(db, account.AccountID, req.AdminKey); err != nil {
			json.NewEncoder(w).Encode(AdminKeySetResponse{OK: false, Error: "INTERNAL_ERROR"})
			return
		}
		_ = setAccountRole(db, account.AccountID, "admin")
		json.NewEncoder(w).Encode(AdminKeySetResponse{OK: true})
	}
}

type AdminRoleRequest struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

type AdminRoleResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

func adminRoleHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if _, ok := requireAdmin(db, w, r); !ok {
			return
		}
		var req AdminRoleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Username == "" {
			json.NewEncoder(w).Encode(AdminRoleResponse{OK: false, Error: "INVALID_REQUEST"})
			return
		}
		if err := setAccountRoleByUsername(db, req.Username, req.Role); err != nil {
			json.NewEncoder(w).Encode(AdminRoleResponse{OK: false, Error: "INTERNAL_ERROR"})
			return
		}
		json.NewEncoder(w).Encode(AdminRoleResponse{OK: true})
	}
}

type AdminKeyForUserRequest struct {
	Username string `json:"username"`
	AdminKey string `json:"adminKey"`
}

type AdminKeyForUserResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

func adminKeyForUserHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if _, ok := requireAdmin(db, w, r); !ok {
			return
		}
		var req AdminKeyForUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Username == "" || req.AdminKey == "" {
			json.NewEncoder(w).Encode(AdminKeyForUserResponse{OK: false, Error: "INVALID_REQUEST"})
			return
		}
		if len(req.AdminKey) < 8 {
			json.NewEncoder(w).Encode(AdminKeyForUserResponse{OK: false, Error: "WEAK_KEY"})
			return
		}
		if err := setAdminKeyByUsername(db, req.Username, req.AdminKey); err != nil {
			json.NewEncoder(w).Encode(AdminKeyForUserResponse{OK: false, Error: "INTERNAL_ERROR"})
			return
		}
		json.NewEncoder(w).Encode(AdminKeyForUserResponse{OK: true})
	}
}

type ModeratorProfileRequest struct {
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
}

type ModeratorProfileResponse struct {
	OK          bool   `json:"ok"`
	Error       string `json:"error,omitempty"`
	Username    string `json:"username,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
}

func moderatorProfileHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireModerator(db, w, r); !ok {
			return
		}
		switch r.Method {
		case http.MethodGet:
			username := r.URL.Query().Get("username")
			if username == "" {
				json.NewEncoder(w).Encode(ModeratorProfileResponse{OK: false, Error: "INVALID_REQUEST"})
				return
			}
			var displayName string
			err := db.QueryRow(`
				SELECT display_name
				FROM accounts
				WHERE username = $1
			`, strings.ToLower(username)).Scan(&displayName)
			if err == sql.ErrNoRows {
				json.NewEncoder(w).Encode(ModeratorProfileResponse{OK: false, Error: "NOT_FOUND"})
				return
			}
			if err != nil {
				json.NewEncoder(w).Encode(ModeratorProfileResponse{OK: false, Error: "INTERNAL_ERROR"})
				return
			}
			json.NewEncoder(w).Encode(ModeratorProfileResponse{OK: true, Username: strings.ToLower(username), DisplayName: displayName})
			return
		case http.MethodPost:
			var req ModeratorProfileRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Username == "" {
				json.NewEncoder(w).Encode(ModeratorProfileResponse{OK: false, Error: "INVALID_REQUEST"})
				return
			}
			displayName := strings.TrimSpace(req.DisplayName)
			if displayName == "" || len(displayName) > 32 {
				json.NewEncoder(w).Encode(ModeratorProfileResponse{OK: false, Error: "INVALID_DISPLAY_NAME"})
				return
			}
			_, err := db.Exec(`
				UPDATE accounts
				SET display_name = $2
				WHERE username = $1
			`, strings.ToLower(req.Username), displayName)
			if err != nil {
				json.NewEncoder(w).Encode(ModeratorProfileResponse{OK: false, Error: "INTERNAL_ERROR"})
				return
			}
			json.NewEncoder(w).Encode(ModeratorProfileResponse{OK: true, Username: strings.ToLower(req.Username), DisplayName: displayName})
			return
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
}

type AdminIPWhitelistRequest struct {
	IP          string `json:"ip"`
	MaxAccounts int    `json:"maxAccounts,omitempty"`
	Action      string `json:"action,omitempty"`
}

type AdminIPWhitelistResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

func adminIPWhitelistHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if _, ok := requireAdmin(db, w, r); !ok {
			return
		}
		var req AdminIPWhitelistRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.IP == "" {
			json.NewEncoder(w).Encode(AdminIPWhitelistResponse{OK: false, Error: "INVALID_REQUEST"})
			return
		}
		action := strings.ToLower(strings.TrimSpace(req.Action))
		if action == "remove" {
			if err := removeIPWhitelist(db, req.IP); err != nil {
				json.NewEncoder(w).Encode(AdminIPWhitelistResponse{OK: false, Error: "INTERNAL_ERROR"})
				return
			}
			json.NewEncoder(w).Encode(AdminIPWhitelistResponse{OK: true})
			return
		}
		if err := upsertIPWhitelist(db, req.IP, req.MaxAccounts); err != nil {
			json.NewEncoder(w).Encode(AdminIPWhitelistResponse{OK: false, Error: "INTERNAL_ERROR"})
			return
		}
		json.NewEncoder(w).Encode(AdminIPWhitelistResponse{OK: true})
	}
}

func adminWhitelistRequestsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, ok := requireAdmin(db, w, r)
		if !ok {
			return
		}
		switch r.Method {
		case http.MethodGet:
			requests, err := listPendingWhitelistRequests(db)
			if err != nil {
				json.NewEncoder(w).Encode(AdminWhitelistRequestListResponse{OK: false, Error: "INTERNAL_ERROR"})
				return
			}
			views := make([]WhitelistRequestView, 0, len(requests))
			for _, req := range requests {
				views = append(views, WhitelistRequestView{
					RequestID: req.RequestID,
					IP:        req.IP,
					AccountID: req.AccountID,
					Reason:    req.Reason,
					CreatedAt: req.CreatedAt,
				})
			}
			json.NewEncoder(w).Encode(AdminWhitelistRequestListResponse{OK: true, Requests: views})
			return
		case http.MethodPost:
			var req AdminWhitelistResolveRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RequestID == "" || req.Decision == "" {
				json.NewEncoder(w).Encode(SimpleResponse{OK: false, Error: "INVALID_REQUEST"})
				return
			}
			decision := strings.ToLower(strings.TrimSpace(req.Decision))
			if decision != "approve" && decision != "deny" {
				json.NewEncoder(w).Encode(SimpleResponse{OK: false, Error: "INVALID_REQUEST"})
				return
			}
			var ip string
			var accountID string
			err := db.QueryRow(`
				SELECT ip, COALESCE(account_id, '')
				FROM ip_whitelist_requests
				WHERE request_id = $1
			`, req.RequestID).Scan(&ip, &accountID)
			if err != nil {
				json.NewEncoder(w).Encode(SimpleResponse{OK: false, Error: "INTERNAL_ERROR"})
				return
			}
			status := "denied"
			if decision == "approve" {
				status = "approved"
				if req.MaxAccounts <= 0 {
					req.MaxAccounts = 2
				}
				if err := upsertIPWhitelist(db, ip, req.MaxAccounts); err != nil {
					json.NewEncoder(w).Encode(SimpleResponse{OK: false, Error: "INTERNAL_ERROR"})
					return
				}
			}
			if err := resolveWhitelistRequest(db, req.RequestID, status, account.AccountID); err != nil {
				json.NewEncoder(w).Encode(SimpleResponse{OK: false, Error: "INTERNAL_ERROR"})
				return
			}
			if accountID != "" {
				message := "Whitelist request denied."
				level := "warn"
				if status == "approved" {
					message = "Whitelist request approved. You may create another account from your IP."
					level = "info"
				}
				_ = createNotification(db, "user", accountID, message, level, "#/home", nil)
			}
			json.NewEncoder(w).Encode(SimpleResponse{OK: true})
			return
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
}

func adminNotificationsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if _, ok := requireAdmin(db, w, r); !ok {
			return
		}
		var req AdminNotificationCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Message == "" || req.TargetRole == "" {
			json.NewEncoder(w).Encode(SimpleResponse{OK: false, Error: "INVALID_REQUEST"})
			return
		}
		level := strings.ToLower(strings.TrimSpace(req.Level))
		if level == "" {
			level = "info"
		}
		if level != "info" && level != "warn" && level != "urgent" {
			json.NewEncoder(w).Encode(SimpleResponse{OK: false, Error: "INVALID_LEVEL"})
			return
		}
		targetRole := strings.ToLower(strings.TrimSpace(req.TargetRole))
		if targetRole != "all" && targetRole != "user" && targetRole != "moderator" && targetRole != "admin" {
			json.NewEncoder(w).Encode(SimpleResponse{OK: false, Error: "INVALID_TARGET"})
			return
		}
		var expiresAt sql.NullTime
		if strings.TrimSpace(req.ExpiresAt) != "" {
			t, err := time.Parse(time.RFC3339, req.ExpiresAt)
			if err != nil {
				json.NewEncoder(w).Encode(SimpleResponse{OK: false, Error: "INVALID_EXPIRES_AT"})
				return
			}
			expiresAt = sql.NullTime{Time: t, Valid: true}
		}
		exp := (*time.Time)(nil)
		if expiresAt.Valid {
			exp = &expiresAt.Time
		}
		if err := createNotification(db, targetRole, req.AccountID, req.Message, level, req.Link, exp); err != nil {
			json.NewEncoder(w).Encode(SimpleResponse{OK: false, Error: "INTERNAL_ERROR"})
			return
		}
		json.NewEncoder(w).Encode(SimpleResponse{OK: true})
	}
}

func adminPlayerControlsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if _, ok := requireAdmin(db, w, r); !ok {
			return
		}
		var req AdminPlayerControlRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			json.NewEncoder(w).Encode(AdminPlayerControlResponse{OK: false, Error: "INVALID_REQUEST"})
			return
		}
		playerID := strings.TrimSpace(req.PlayerID)
		if playerID == "" && strings.TrimSpace(req.Username) != "" {
			if err := db.QueryRow(`
				SELECT player_id FROM accounts WHERE username = $1
			`, strings.ToLower(strings.TrimSpace(req.Username))).Scan(&playerID); err != nil {
				json.NewEncoder(w).Encode(AdminPlayerControlResponse{OK: false, Error: "PLAYER_NOT_FOUND"})
				return
			}
		}
		if playerID == "" {
			json.NewEncoder(w).Encode(AdminPlayerControlResponse{OK: false, Error: "INVALID_REQUEST"})
			return
		}

		if req.SetCoins != nil {
			if *req.SetCoins < 0 {
				json.NewEncoder(w).Encode(AdminPlayerControlResponse{OK: false, Error: "INVALID_COINS"})
				return
			}
			_, err := db.Exec(`
				UPDATE players
				SET coins = $2
				WHERE player_id = $1
			`, playerID, *req.SetCoins)
			if err != nil {
				json.NewEncoder(w).Encode(AdminPlayerControlResponse{OK: false, Error: "INTERNAL_ERROR"})
				return
			}
		}
		if req.AddCoins != nil && *req.AddCoins != 0 {
			_, err := db.Exec(`
				UPDATE players
				SET coins = GREATEST(coins + $2, 0)
				WHERE player_id = $1
			`, playerID, *req.AddCoins)
			if err != nil {
				json.NewEncoder(w).Encode(AdminPlayerControlResponse{OK: false, Error: "INTERNAL_ERROR"})
				return
			}
		}
		if req.SetStars != nil {
			if *req.SetStars < 0 {
				json.NewEncoder(w).Encode(AdminPlayerControlResponse{OK: false, Error: "INVALID_STARS"})
				return
			}
			_, err := db.Exec(`
				UPDATE players
				SET stars = $2
				WHERE player_id = $1
			`, playerID, *req.SetStars)
			if err != nil {
				json.NewEncoder(w).Encode(AdminPlayerControlResponse{OK: false, Error: "INTERNAL_ERROR"})
				return
			}
		}
		if req.AddStars != nil && *req.AddStars != 0 {
			_, err := db.Exec(`
				UPDATE players
				SET stars = GREATEST(stars + $2, 0)
				WHERE player_id = $1
			`, playerID, *req.AddStars)
			if err != nil {
				json.NewEncoder(w).Encode(AdminPlayerControlResponse{OK: false, Error: "INTERNAL_ERROR"})
				return
			}
		}
		if req.DripMultiplier != nil {
			value := *req.DripMultiplier
			if value < 0.1 || value > 10.0 {
				json.NewEncoder(w).Encode(AdminPlayerControlResponse{OK: false, Error: "INVALID_MULTIPLIER"})
				return
			}
			_, err := db.Exec(`
				UPDATE players
				SET drip_multiplier = $2
				WHERE player_id = $1
			`, playerID, value)
			if err != nil {
				json.NewEncoder(w).Encode(AdminPlayerControlResponse{OK: false, Error: "INTERNAL_ERROR"})
				return
			}
		}
		if req.DripPaused != nil {
			_, err := db.Exec(`
				UPDATE players
				SET drip_paused = $2
				WHERE player_id = $1
			`, playerID, *req.DripPaused)
			if err != nil {
				json.NewEncoder(w).Encode(AdminPlayerControlResponse{OK: false, Error: "INTERNAL_ERROR"})
				return
			}
		}
		if req.TouchActive {
			_, err := db.Exec(`
				UPDATE players
				SET last_active_at = NOW()
				WHERE player_id = $1
			`, playerID)
			if err != nil {
				json.NewEncoder(w).Encode(AdminPlayerControlResponse{OK: false, Error: "INTERNAL_ERROR"})
				return
			}
		}

		var coins int64
		var stars int64
		var dripMultiplier float64
		var dripPaused bool
		var lastActive time.Time
		var lastGrant time.Time
		if err := db.QueryRow(`
			SELECT coins, stars, drip_multiplier, drip_paused, last_active_at, last_coin_grant_at
			FROM players
			WHERE player_id = $1
		`, playerID).Scan(&coins, &stars, &dripMultiplier, &dripPaused, &lastActive, &lastGrant); err != nil {
			json.NewEncoder(w).Encode(AdminPlayerControlResponse{OK: false, Error: "INTERNAL_ERROR"})
			return
		}

		json.NewEncoder(w).Encode(AdminPlayerControlResponse{
			OK:             true,
			PlayerID:       playerID,
			Coins:          coins,
			Stars:          stars,
			DripMultiplier: dripMultiplier,
			DripPaused:     dripPaused,
			LastActiveAt:   lastActive,
			LastGrantAt:    lastGrant,
		})
	}
}

func adminSettingsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(db, w, r); !ok {
			return
		}
		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode(AdminGlobalSettingsResponse{OK: true, Settings: GetGlobalSettings()})
			return
		case http.MethodPost:
			var req AdminGlobalSettingsRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				json.NewEncoder(w).Encode(AdminGlobalSettingsResponse{OK: false, Error: "INVALID_REQUEST"})
				return
			}
			updates := map[string]string{}
			if req.ActiveDripIntervalSeconds != nil {
				updates["active_drip_interval_seconds"] = strconv.Itoa(*req.ActiveDripIntervalSeconds)
			}
			if req.IdleDripIntervalSeconds != nil {
				updates["idle_drip_interval_seconds"] = strconv.Itoa(*req.IdleDripIntervalSeconds)
			}
			if req.ActiveDripAmount != nil {
				updates["active_drip_amount"] = strconv.Itoa(*req.ActiveDripAmount)
			}
			if req.IdleDripAmount != nil {
				updates["idle_drip_amount"] = strconv.Itoa(*req.IdleDripAmount)
			}
			if req.ActivityWindowSeconds != nil {
				updates["activity_window_seconds"] = strconv.Itoa(*req.ActivityWindowSeconds)
			}
			if req.DripEnabled != nil {
				updates["drip_enabled"] = strconv.FormatBool(*req.DripEnabled)
			}
			if len(updates) == 0 {
				json.NewEncoder(w).Encode(AdminGlobalSettingsResponse{OK: false, Error: "NO_UPDATES"})
				return
			}
			settings, err := UpdateGlobalSettings(db, updates)
			if err != nil {
				json.NewEncoder(w).Encode(AdminGlobalSettingsResponse{OK: false, Error: "INTERNAL_ERROR"})
				return
			}
			json.NewEncoder(w).Encode(AdminGlobalSettingsResponse{OK: true, Settings: settings})
			return
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
}

func adminStarPurchaseLogHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(db, w, r); !ok {
			return
		}
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		query := r.URL.Query()
		limit := 50
		if raw := strings.TrimSpace(query.Get("limit")); raw != "" {
			if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
				limit = parsed
			}
		}
		if limit > 250 {
			limit = 250
		}

		accountID := strings.TrimSpace(query.Get("accountId"))
		playerID := strings.TrimSpace(query.Get("playerId"))
		seasonID := strings.TrimSpace(query.Get("seasonId"))
		purchaseType := strings.TrimSpace(query.Get("purchaseType"))
		variant := strings.TrimSpace(query.Get("variant"))
		fromRaw := strings.TrimSpace(query.Get("from"))
		toRaw := strings.TrimSpace(query.Get("to"))

		parseInt64 := func(key string) (*int64, bool) {
			raw := strings.TrimSpace(query.Get(key))
			if raw == "" {
				return nil, false
			}
			value, err := strconv.ParseInt(raw, 10, 64)
			if err != nil {
				return nil, false
			}
			return &value, true
		}

		parseTime := func(raw string) (*time.Time, bool) {
			if raw == "" {
				return nil, false
			}
			parsed, err := time.Parse(time.RFC3339, raw)
			if err != nil {
				return nil, false
			}
			return &parsed, true
		}

		fromTime, hasFrom := parseTime(fromRaw)
		toTime, hasTo := parseTime(toRaw)
		minPrice, hasMinPrice := parseInt64("minPrice")
		maxPrice, hasMaxPrice := parseInt64("maxPrice")
		minCoinsBefore, hasMinCoinsBefore := parseInt64("minCoinsBefore")
		maxCoinsBefore, hasMaxCoinsBefore := parseInt64("maxCoinsBefore")
		minCoinsAfter, hasMinCoinsAfter := parseInt64("minCoinsAfter")
		maxCoinsAfter, hasMaxCoinsAfter := parseInt64("maxCoinsAfter")
		minStarsBefore, hasMinStarsBefore := parseInt64("minStarsBefore")
		maxStarsBefore, hasMaxStarsBefore := parseInt64("maxStarsBefore")
		minStarsAfter, hasMinStarsAfter := parseInt64("minStarsAfter")
		maxStarsAfter, hasMaxStarsAfter := parseInt64("maxStarsAfter")

		clauses := []string{}
		args := []interface{}{}
		argIndex := 1
		if accountID != "" {
			clauses = append(clauses, "account_id = $"+strconv.Itoa(argIndex))
			args = append(args, accountID)
			argIndex++
		}
		if playerID != "" {
			clauses = append(clauses, "player_id = $"+strconv.Itoa(argIndex))
			args = append(args, playerID)
			argIndex++
		}
		if seasonID != "" {
			clauses = append(clauses, "season_id = $"+strconv.Itoa(argIndex))
			args = append(args, seasonID)
			argIndex++
		}
		if purchaseType != "" {
			clauses = append(clauses, "purchase_type = $"+strconv.Itoa(argIndex))
			args = append(args, purchaseType)
			argIndex++
		}
		if variant != "" {
			clauses = append(clauses, "variant = $"+strconv.Itoa(argIndex))
			args = append(args, variant)
			argIndex++
		}
		if hasFrom {
			clauses = append(clauses, "created_at >= $"+strconv.Itoa(argIndex))
			args = append(args, *fromTime)
			argIndex++
		}
		if hasTo {
			clauses = append(clauses, "created_at <= $"+strconv.Itoa(argIndex))
			args = append(args, *toTime)
			argIndex++
		}
		if hasMinPrice {
			clauses = append(clauses, "price_paid >= $"+strconv.Itoa(argIndex))
			args = append(args, *minPrice)
			argIndex++
		}
		if hasMaxPrice {
			clauses = append(clauses, "price_paid <= $"+strconv.Itoa(argIndex))
			args = append(args, *maxPrice)
			argIndex++
		}
		if hasMinCoinsBefore {
			clauses = append(clauses, "coins_before >= $"+strconv.Itoa(argIndex))
			args = append(args, *minCoinsBefore)
			argIndex++
		}
		if hasMaxCoinsBefore {
			clauses = append(clauses, "coins_before <= $"+strconv.Itoa(argIndex))
			args = append(args, *maxCoinsBefore)
			argIndex++
		}
		if hasMinCoinsAfter {
			clauses = append(clauses, "coins_after >= $"+strconv.Itoa(argIndex))
			args = append(args, *minCoinsAfter)
			argIndex++
		}
		if hasMaxCoinsAfter {
			clauses = append(clauses, "coins_after <= $"+strconv.Itoa(argIndex))
			args = append(args, *maxCoinsAfter)
			argIndex++
		}
		if hasMinStarsBefore {
			clauses = append(clauses, "stars_before >= $"+strconv.Itoa(argIndex))
			args = append(args, *minStarsBefore)
			argIndex++
		}
		if hasMaxStarsBefore {
			clauses = append(clauses, "stars_before <= $"+strconv.Itoa(argIndex))
			args = append(args, *maxStarsBefore)
			argIndex++
		}
		if hasMinStarsAfter {
			clauses = append(clauses, "stars_after >= $"+strconv.Itoa(argIndex))
			args = append(args, *minStarsAfter)
			argIndex++
		}
		if hasMaxStarsAfter {
			clauses = append(clauses, "stars_after <= $"+strconv.Itoa(argIndex))
			args = append(args, *maxStarsAfter)
			argIndex++
		}

		sqlQuery := `
			SELECT id, account_id, player_id, season_id, purchase_type, variant,
				price_paid, coins_before, coins_after, stars_before, stars_after, created_at
			FROM star_purchase_log
		`
		if len(clauses) > 0 {
			sqlQuery += " WHERE " + strings.Join(clauses, " AND ")
		}
		sqlQuery += " ORDER BY created_at DESC LIMIT " + strconv.Itoa(limit)

		rows, err := db.Query(sqlQuery, args...)
		if err != nil {
			json.NewEncoder(w).Encode(AdminStarPurchaseLogResponse{OK: false, Error: "INTERNAL_ERROR"})
			return
		}
		defer rows.Close()

		items := []AdminStarPurchaseLogItem{}
		for rows.Next() {
			var item AdminStarPurchaseLogItem
			if err := rows.Scan(
				&item.ID,
				&item.AccountID,
				&item.PlayerID,
				&item.SeasonID,
				&item.PurchaseType,
				&item.Variant,
				&item.PricePaid,
				&item.CoinsBefore,
				&item.CoinsAfter,
				&item.StarsBefore,
				&item.StarsAfter,
				&item.CreatedAt,
			); err != nil {
				continue
			}
			items = append(items, item)
		}

		json.NewEncoder(w).Encode(AdminStarPurchaseLogResponse{OK: true, Items: items})
	}
}
