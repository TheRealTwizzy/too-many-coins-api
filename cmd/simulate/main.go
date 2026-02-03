package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	seasonID := defaultSeasonID
	var params CalibrationParams

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL != "" {
		db, err := sql.Open("postgres", dbURL)
		if err != nil {
			log.Fatal("failed to open database:", err)
		}
		defer db.Close()
		if err := db.Ping(); err != nil {
			log.Fatal("failed to ping database:", err)
		}
		if err := ensureSchema(db); err != nil {
			log.Fatal("failed to ensure schema:", err)
		}
		params, err = LoadOrCalibrateSeason(db, seasonID)
		if err != nil {
			log.Fatal("failed to load calibration:", err)
		}
	} else {
		params = CalibrateSeason(seasonID, seasonStart(), TelemetrySnapshot{})
	}

	report, err := RunSeasonSimulation(params)
	if err != nil {
		log.Fatal("simulation failed:", err)
	}
	path, err := SaveSimulationReport(report, "")
	if err != nil {
		log.Fatal("failed to save report:", err)
	}
	fmt.Println("Simulation report saved to", path)
}
