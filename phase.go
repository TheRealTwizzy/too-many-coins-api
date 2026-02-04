package main

import (
	"log"
	"os"
	"strings"
)

type Phase string

const (
	PhaseAlpha   Phase = "alpha"
	PhaseBeta    Phase = "beta"
	PhaseRelease Phase = "release"
)

// CurrentPhase is server-authoritative and must never be client-defined.
// It is sourced from PHASE, with APP_ENV as a fallback.
func CurrentPhase() Phase {
	if phase, ok := parsePhaseFromEnv("PHASE"); ok {
		return phase
	}
	if phase, ok := parsePhaseFromEnv("APP_ENV"); ok {
		return phase
	}
	return PhaseAlpha
}

func parsePhaseFromEnv(key string) (Phase, bool) {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	switch value {
	case "alpha":
		return PhaseAlpha, true
	case "beta":
		return PhaseBeta, true
	case "release":
		return PhaseRelease, true
	case "":
		return "", false
	default:
		log.Println("invalid phase value for", key, "=", value, "; defaulting to alpha")
		return PhaseAlpha, true
	}
}
