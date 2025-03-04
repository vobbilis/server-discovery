package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/vobbilis/codegen/server-discovery/pkg/controller"
	"github.com/vobbilis/codegen/server-discovery/pkg/database"
	"github.com/vobbilis/codegen/server-discovery/pkg/models"
	"github.com/vobbilis/codegen/server-discovery/pkg/server"
)

func main() {
	configFile := flag.String("config", "config.json", "Path to configuration file")
	flag.Parse()

	// Read configuration file
	config, err := models.ReadConfig(*configFile)
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	// Initialize database connection
	db, err := database.NewDatabase(&config.Database)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	// Initialize discovery controller
	discoveryCtrl := controller.NewDiscoveryController(config, db)

	// Initialize API server
	apiServer := server.NewAPIServer(config, db, discoveryCtrl)

	// Start API server in a goroutine
	go func() {
		if err := apiServer.Start(); err != nil {
			log.Printf("Error starting API server: %v", err)
		}
	}()

	log.Printf("Server started on port %d", config.API.Port)

	// Wait for interrupt signal to gracefully shut down the server
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down server...")
}
