package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// BugReportCategory defines valid bug report categories
type BugReportCategory string

const (
	BugReportCategoryUI          BugReportCategory = "ui"
	BugReportCategoryEconomy     BugReportCategory = "economy"
	BugReportCategoryPerformance BugReportCategory = "performance"
	BugReportCategoryOther       BugReportCategory = "other"
)

// IsValidCategory checks if a category is valid
func IsValidCategory(cat string) bool {
	switch strings.ToLower(cat) {
	case "ui", "economy", "performance", "other":
		return true
	default:
		return false
	}
}

// BugReport represents an immutable bug report record
type BugReport struct {
	BugReportID   int64     `json:"bugReportId"`
	PlayerID      *string   `json:"playerId,omitempty"`
	SeasonID      string    `json:"seasonId"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Category      string    `json:"category"`
	ClientVersion *string   `json:"clientVersion,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
}

// CreateBugReportRequest represents player intake payload
type CreateBugReportRequest struct {
	Title         string `json:"title"`
	Description   string `json:"description"`
	Category      string `json:"category,omitempty"`
	ClientVersion string `json:"clientVersion,omitempty"`
}

// CreateBugReportResponse indicates success/failure of submission
type CreateBugReportResponse struct {
	OK       bool   `json:"ok"`
	Error    string `json:"error,omitempty"`
	ReportID int64  `json:"reportId,omitempty"`
}

// AdminBugReport extends BugReport with admin-specific fields
type AdminBugReport struct {
	BugReportID   int64     `json:"bugReportId"`
	PlayerID      *string   `json:"playerId,omitempty"`
	SeasonID      string    `json:"seasonId"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Category      string    `json:"category"`
	ClientVersion *string   `json:"clientVersion,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
}

// AdminBugReportsResponse is the read-only admin list view
type AdminBugReportsResponse struct {
	OK     bool             `json:"ok"`
	Error  string           `json:"error,omitempty"`
	Items  []AdminBugReport `json:"items,omitempty"`
	Total  int              `json:"total,omitempty"`
	Limit  int              `json:"limit,omitempty"`
	Offset int              `json:"offset,omitempty"`
}

// SubmitBugReport inserts an append-only bug report record
// playerID is optional for anonymous submissions
// Returns the bug_report_id on success
func SubmitBugReport(db *sql.DB, playerID *string, seasonID string, title string, description string, category string, clientVersion *string) (int64, error) {
	if seasonID == "" {
		return 0, fmt.Errorf("seasonID is required")
	}
	if title == "" {
		return 0, fmt.Errorf("title is required")
	}
	if description == "" {
		return 0, fmt.Errorf("description is required")
	}

	// Sanitize category
	if category == "" {
		category = "other"
	}
	if !IsValidCategory(category) {
		category = "other"
	}
	category = strings.ToLower(category)

	// Insert append-only record
	query := `
		INSERT INTO bug_reports (player_id, season_id, title, description, category, client_version, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING bug_report_id
	`

	var reportID int64
	err := db.QueryRow(query, playerID, seasonID, title, description, category, clientVersion).Scan(&reportID)
	if err != nil {
		return 0, err
	}

	return reportID, nil
}

// GetBugReports retrieves bug reports for admin view (read-only)
// Returns paginated results ordered by creation time (newest first)
func GetBugReports(db *sql.DB, seasonID string, limit int, offset int) ([]AdminBugReport, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}
	if offset < 0 {
		offset = 0
	}

	// Count total
	countQuery := `SELECT COUNT(*) FROM bug_reports WHERE season_id = $1`
	var total int
	err := db.QueryRow(countQuery, seasonID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Fetch paginated results
	query := `
		SELECT bug_report_id, player_id, season_id, title, description, category, client_version, created_at
		FROM bug_reports
		WHERE season_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := db.Query(query, seasonID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	reports := []AdminBugReport{}
	for rows.Next() {
		var report AdminBugReport
		err := rows.Scan(
			&report.BugReportID,
			&report.PlayerID,
			&report.SeasonID,
			&report.Title,
			&report.Description,
			&report.Category,
			&report.ClientVersion,
			&report.CreatedAt,
		)
		if err != nil {
			continue
		}
		reports = append(reports, report)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}
