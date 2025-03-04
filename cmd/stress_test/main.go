package main

import (
	"log"
	"os"

	"github.com/vobbilis/codegen/server-discovery/pkg/database"
	"github.com/vobbilis/codegen/server-discovery/pkg/models"
	"github.com/vobbilis/codegen/server-discovery/pkg/stress"
)

func main() {
	// Create database connection using Docker container settings
	db, err := database.NewDatabase(models.DatabaseConfig{
		Host:     "server_discovery_test_db",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		DBName:   "server_discovery",
		SSLMode:  "disable",
	})
	if err != nil {
		log.Printf("[ERROR] Failed to connect to database: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	// Create stress test runner
	stressTest := stress.NewStressTest(db)

	// Run stress test
	log.Printf("[INFO] Starting discovery stress test")
	if err := stressTest.RunDiscoveryStressTest(); err != nil {
		log.Printf("[ERROR] Stress test failed: %v", err)
		os.Exit(1)
	}
	log.Printf("[INFO] Stress test completed successfully")
}
