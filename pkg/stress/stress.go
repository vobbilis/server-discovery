package stress

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/vobbilis/codegen/server-discovery/pkg/models"
)

// StressTest represents a stress test runner
type StressTest struct {
	db Database
}

// Database interface defines the methods needed for stress testing
type Database interface {
	GetAllServers() ([]models.ServerDetails, error)
	CreateDiscoveryResult(result models.DiscoveryResult) (int, error)
}

// NewStressTest creates a new stress test runner
func NewStressTest(db Database) *StressTest {
	return &StressTest{db: db}
}

// RunDiscoveryStressTest runs discovery stress test for all servers
func (st *StressTest) RunDiscoveryStressTest() error {
	// Get all servers
	servers, err := st.db.GetAllServers()
	if err != nil {
		return fmt.Errorf("failed to get servers: %w", err)
	}

	log.Printf("[INFO] Starting discovery stress test for %d servers", len(servers))

	// Create a wait group to wait for all goroutines
	var wg sync.WaitGroup
	// Create error channel to collect errors
	errChan := make(chan error, len(servers))

	// Process each server
	for i, server := range servers {
		wg.Add(1)
		go func(s models.ServerDetails, idx int) {
			defer wg.Done()

			log.Printf("[DEBUG] Processing server %d/%d: %s (ID: %d)", idx+1, len(servers), s.Hostname, s.ID)

			// Create discovery result
			discovery := models.DiscoveryResult{
				ServerID:    s.ID,
				Server:      s.Hostname,
				Success:     true,
				Message:     fmt.Sprintf("Stress test discovery for server %s", s.Hostname),
				Status:      "completed",
				StartTime:   time.Now().Add(-5 * time.Second), // Simulate 5-second discovery
				EndTime:     time.Now(),
				LastChecked: time.Now(),
				Region:      s.Region,
			}

			// Save discovery result
			id, err := st.db.CreateDiscoveryResult(discovery)
			if err != nil {
				log.Printf("[ERROR] Failed to create discovery for server %s (ID: %d): %v", s.Hostname, s.ID, err)
				errChan <- fmt.Errorf("failed to create discovery for server %d: %w", s.ID, err)
				return
			}

			log.Printf("[DEBUG] Created discovery %d for server %s (ID: %d)", id, s.Hostname, s.ID)
		}(server, i)

		// Add a small delay between goroutines to avoid overwhelming the database
		time.Sleep(10 * time.Millisecond)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errChan)

	// Check for errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("stress test completed with %d errors: %v", len(errors), errors)
	}

	log.Printf("[INFO] Stress test completed successfully for %d servers", len(servers))
	return nil
}
