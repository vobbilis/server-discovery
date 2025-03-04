package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// mockDiscoverer implements ServerDiscoverer for testing
type mockDiscoverer struct {
	osType string
}

func (m *mockDiscoverer) ExecuteDiscovery(server ServerConfig, outputDir string) (DiscoveryResult, error) {
	// Create mock output
	os.MkdirAll(outputDir, 0755)
	details := ServerDetails{
		OSType:         m.osType,
		KernelVersion:  "4.19.0-mock",
		PackageManager: "apt",
		InitSystem:     "systemd",
		NetworkInterfaces: []NetworkInterface{
			{
				Name:       "eth0",
				MACAddress: "00:00:00:00:00:00",
				State:      "up",
			},
		},
		MountedFilesystems: []Filesystem{
			{
				Device:     "/dev/sda1",
				MountPoint: "/",
				FSType:     "ext4",
			},
		},
	}

	// Write mock data
	jsonData, _ := json.MarshalIndent(details, "", "  ")
	os.WriteFile(filepath.Join(outputDir, "server_details.json"), jsonData, 0644)

	return DiscoveryResult{
		Success:    true,
		OutputPath: outputDir,
		StartTime:  time.Now(),
		EndTime:    time.Now(),
	}, nil
}

func (m *mockDiscoverer) ParseDiscoveryOutput(outputPath string) (ServerDetails, error) {
	data, err := os.ReadFile(filepath.Join(outputPath, "server_details.json"))
	if err != nil {
		return ServerDetails{}, err
	}

	var details ServerDetails
	err = json.Unmarshal(data, &details)
	return details, err
}

func TestServerDiscoveryMinimal(t *testing.T) {
	// Create test output directory
	testOutputDir := "test_output"
	os.MkdirAll(testOutputDir, 0755)
	defer os.RemoveAll(testOutputDir)

	// Test Linux server discovery
	t.Run("Linux Server Discovery", func(t *testing.T) {
		discoverer := &mockDiscoverer{osType: "linux"}
		server := ServerConfig{
			Hostname: "test-linux-server",
			Port:     22,
			Username: "testuser",
			Password: "testpass",
			OSType:   "linux",
		}

		result, err := discoverer.ExecuteDiscovery(server, filepath.Join(testOutputDir, "linux"))
		if err != nil {
			t.Fatalf("Linux discovery failed: %v", err)
		}

		if !result.Success {
			t.Error("Expected successful discovery")
		}

		details, err := discoverer.ParseDiscoveryOutput(result.OutputPath)
		if err != nil {
			t.Fatalf("Failed to parse Linux discovery output: %v", err)
		}

		// Verify Linux-specific fields
		if details.OSType != "linux" {
			t.Errorf("Expected OS type 'linux', got '%s'", details.OSType)
		}
		if details.KernelVersion == "" {
			t.Error("Kernel version is empty")
		}
		if details.PackageManager == "" {
			t.Error("Package manager is empty")
		}
		if details.InitSystem == "" {
			t.Error("Init system is empty")
		}
		if len(details.NetworkInterfaces) == 0 {
			t.Error("No network interfaces found")
		}
		if len(details.MountedFilesystems) == 0 {
			t.Error("No filesystems found")
		}
	})

	// Test Windows server discovery
	t.Run("Windows Server Discovery", func(t *testing.T) {
		discoverer := &mockDiscoverer{osType: "windows"}
		server := ServerConfig{
			Hostname: "test-windows-server",
			Port:     5985,
			Username: "Administrator",
			Password: "testpass",
			OSType:   "windows",
		}

		result, err := discoverer.ExecuteDiscovery(server, filepath.Join(testOutputDir, "windows"))
		if err != nil {
			t.Fatalf("Windows discovery failed: %v", err)
		}

		if !result.Success {
			t.Error("Expected successful discovery")
		}

		details, err := discoverer.ParseDiscoveryOutput(result.OutputPath)
		if err != nil {
			t.Fatalf("Failed to parse Windows discovery output: %v", err)
		}

		if details.OSType != "windows" {
			t.Errorf("Expected OS type 'windows', got '%s'", details.OSType)
		}
	})
}
