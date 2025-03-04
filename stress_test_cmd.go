package main

import (
	"flag"
	"log"
	"os"

	"github.com/vobbilis/codegen/server-discovery/pkg/models"
)

func main() {
	// Parse command line flags
	configFile := flag.String("config", "config.json", "Path to config file")
	flag.Parse()

	// Load configuration
	config, err := models.LoadConfig(*configFile)
	if err != nil {
		log.Printf("[ERROR] Failed to load config: %v", err)
		os.Exit(1)
	}

	// Create database connection
	db, err := NewDatabase(config.DatabaseConfig)
	if err != nil {
		log.Printf("[ERROR] Failed to connect to database: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	// Create stress test runner
	stressTest := NewStressTest(db)

	// Run stress test
	log.Printf("[INFO] Starting discovery stress test")
	if err := stressTest.RunDiscoveryStressTest(); err != nil {
		log.Printf("[ERROR] Stress test failed: %v", err)
		os.Exit(1)
	}
	log.Printf("[INFO] Stress test completed successfully")
}
