package main

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultSeasonID         = "season-1"
	defaultSeasonStartLag   = -21 * 24 * time.Hour
	alphaSeasonLengthDays   = 14
	alphaSeasonMaxDays      = 21
	betaSeasonLengthDays    = 28
	releaseSeasonLengthDays = 28
)

var (
	seasonStartOnce sync.Once
	seasonStartTime time.Time
)

func currentSeasonID() string {
	return defaultSeasonID
}

func DefaultSeasonID() string {
	return defaultSeasonID
}

func SeasonStart() time.Time {
	return seasonStart()
}

func seasonStart() time.Time {
	seasonStartOnce.Do(func() {
		start := os.Getenv("SEASON_START_UTC")
		if start != "" {
			if parsed, err := time.Parse(time.RFC3339, start); err == nil {
				seasonStartTime = parsed.UTC()
				return
			}
		}

		seasonStartTime = time.Now().UTC().Add(defaultSeasonStartLag)
	})

	return seasonStartTime
}

func seasonEnd() time.Time {
	return seasonStart().Add(seasonLength())
}

func isSeasonEnded(now time.Time) bool {
	return !now.Before(seasonEnd())
}

func seasonSecondsRemaining(now time.Time) int64 {
	remaining := seasonEnd().Sub(now)
	if remaining < 0 {
		return 0
	}
	return int64(remaining.Seconds())
}

func seasonLength() time.Duration {
	switch CurrentPhase() {
	case PhaseBeta:
		return time.Duration(betaSeasonLengthDays) * 24 * time.Hour
	case PhaseRelease:
		return time.Duration(releaseSeasonLengthDays) * 24 * time.Hour
	default:
		return alphaSeasonLength()
	}
}

func alphaSeasonLength() time.Duration {
	// Alpha is single-season and tightly bounded: default 14 days.
	// Extension to 21 days is allowed only with explicit env + reason (telemetry gaps).
	lengthDays := alphaSeasonLengthDays
	if extensionDays, ok := alphaSeasonExtensionDays(); ok {
		if extensionDays > alphaSeasonMaxDays {
			extensionDays = alphaSeasonMaxDays
		}
		if extensionDays > lengthDays {
			lengthDays = extensionDays
		}
	}
	return time.Duration(lengthDays) * 24 * time.Hour
}

func alphaSeasonExtensionDays() (int, bool) {
	value := strings.TrimSpace(os.Getenv("ALPHA_SEASON_EXTENSION_DAYS"))
	if value == "" {
		return 0, false
	}
	if strings.TrimSpace(os.Getenv("ALPHA_SEASON_EXTENSION_REASON")) == "" {
		return 0, false
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, false
	}
	return parsed, true
}
