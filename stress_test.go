package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/vobbilis/codegen/server-discovery/pkg/models"
)

// ServerInfo represents the information collected about a discovered server
type ServerInfo struct {
	IP          string
	Hostname    string
	OS          string
	Status      string
	LastChecked time.Time
	Services    []string
}

// mockStressDiscoverer simulates server discovery with configurable delays and failures
type mockStressDiscoverer struct {
	failureRate     float64 // percentage of servers that will fail discovery (0-1)
	minDelay        time.Duration
	maxDelay        time.Duration
	mu              sync.Mutex
	discoveredCount int
}

func (m *mockStressDiscoverer) Discover(ctx context.Context, ip string) (*ServerInfo, error) {
	m.mu.Lock()
	m.discoveredCount++
	currentCount := m.discoveredCount
	m.mu.Unlock()

	// Simulate random delay
	delay := m.minDelay + time.Duration(rand.Float64()*float64(m.maxDelay-m.minDelay))

	// Use a select to respect context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(delay):
	}

	// Simulate random failures
	if rand.Float64() < m.failureRate {
		return nil, fmt.Errorf("discovery failed for %s", ip)
	}

	// Alternate between Windows and Linux for variety
	isWindows := currentCount%2 == 0

	return &ServerInfo{
		IP:          ip,
		Hostname:    fmt.Sprintf("server-%d", currentCount),
		OS:          map[bool]string{true: "Windows", false: "Linux"}[isWindows],
		Status:      "online",
		LastChecked: time.Now(),
		Services:    []string{"ssh", "http"},
	}, nil
}

func generateIPs(count int) []string {
	ips := make([]string, count)
	for i := 0; i < count; i++ {
		// Generate IPs in different subnets to simulate real networks
		subnet := i / 254
		host := 1 + (i % 254)
		ips[i] = fmt.Sprintf("10.%d.0.%d", subnet, host)
	}
	return ips
}

func TestStressDiscovery(t *testing.T) {
	tests := []struct {
		name        string
		serverCount int
		failureRate float64
		minDelay    time.Duration
		maxDelay    time.Duration
		timeout     time.Duration
		concurrent  int
	}{
		{
			name:        "Small Scale - 100 Servers",
			serverCount: 100,
			failureRate: 0.05, // 5% failure rate
			minDelay:    50 * time.Millisecond,
			maxDelay:    200 * time.Millisecond,
			timeout:     1 * time.Minute,
			concurrent:  10,
		},
		{
			name:        "Medium Scale - 1000 Servers",
			serverCount: 1000,
			failureRate: 0.10, // 10% failure rate
			minDelay:    100 * time.Millisecond,
			maxDelay:    500 * time.Millisecond,
			timeout:     5 * time.Minute,
			concurrent:  50,
		},
		{
			name:        "Large Scale - 5000 Servers",
			serverCount: 5000,
			failureRate: 0.15, // 15% failure rate
			minDelay:    200 * time.Millisecond,
			maxDelay:    1 * time.Second,
			timeout:     15 * time.Minute,
			concurrent:  100,
		},
		{
			name:        "Worst Case - High Failure Rate",
			serverCount: 1000,
			failureRate: 0.30, // 30% failure rate
			minDelay:    500 * time.Millisecond,
			maxDelay:    3 * time.Second,
			timeout:     10 * time.Minute,
			concurrent:  20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			discoverer := &mockStressDiscoverer{
				failureRate: tt.failureRate,
				minDelay:    tt.minDelay,
				maxDelay:    tt.maxDelay,
			}

			ips := generateIPs(tt.serverCount)
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			// Create work pool
			jobs := make(chan string, tt.serverCount)
			results := make(chan struct {
				info *ServerInfo
				err  error
			}, tt.serverCount)

			// Start workers
			var wg sync.WaitGroup
			for i := 0; i < tt.concurrent; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for ip := range jobs {
						info, err := discoverer.Discover(ctx, ip)
						results <- struct {
							info *ServerInfo
							err  error
						}{info, err}
					}
				}()
			}

			// Send work
			startTime := time.Now()
			go func() {
				for _, ip := range ips {
					jobs <- ip
				}
				close(jobs)
			}()

			// Collect results
			go func() {
				wg.Wait()
				close(results)
			}()

			var successCount, failureCount int
			for result := range results {
				if result.err != nil {
					failureCount++
				} else {
					successCount++
				}
			}

			duration := time.Since(startTime)

			// Report results
			t.Logf("Test: %s", tt.name)
			t.Logf("Duration: %v", duration)
			t.Logf("Total servers: %d", tt.serverCount)
			t.Logf("Successful discoveries: %d (%.2f%%)", successCount, float64(successCount)/float64(tt.serverCount)*100)
			t.Logf("Failed discoveries: %d (%.2f%%)", failureCount, float64(failureCount)/float64(tt.serverCount)*100)
			t.Logf("Average time per server: %v", duration/time.Duration(tt.serverCount))
			t.Logf("Throughput: %.2f servers/second", float64(tt.serverCount)/duration.Seconds())

			// Verify results
			expectedFailures := int(float64(tt.serverCount) * tt.failureRate * 1.5) // Allow 50% margin
			if failureCount > expectedFailures {
				t.Errorf("Too many failures: got %d, expected maximum %d", failureCount, expectedFailures)
			}

			if ctx.Err() == context.DeadlineExceeded {
				t.Errorf("Test timed out after %v", tt.timeout)
			}
		})
	}
}

// StressTest represents a stress test runner
type StressTest struct {
	db *Database
}

// NewStressTest creates a new stress test runner
func NewStressTest(db *Database) *StressTest {
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
	for _, server := range servers {
		wg.Add(1)
		go func(s models.ServerDetails) {
			defer wg.Done()

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
				errChan <- fmt.Errorf("failed to create discovery for server %d: %w", s.ID, err)
				return
			}

			log.Printf("[DEBUG] Created discovery %d for server %s", id, s.Hostname)
		}(server)
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
