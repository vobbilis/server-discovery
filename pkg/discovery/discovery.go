package discovery

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/vobbilis/codegen/server-discovery/pkg/models"
)

// ServerDiscoverer interface defines methods for server discovery
type ServerDiscoverer interface {
	ExecuteDiscovery(server models.ServerConfig, outputDir string) (models.DiscoveryResult, error)
	ParseDiscoveryOutput(outputPath string) (models.ServerDetails, error)
}

// RunLinuxDiscovery executes discovery on a Linux server
func RunLinuxDiscovery(config models.SSHConfig, outputDir string) (string, error) {
	// Create a unique output directory for this execution
	timestamp := time.Now().Format("20060102_150405")
	executionDir := filepath.Join(outputDir, fmt.Sprintf("%s_%s", config.Host, timestamp))
	if err := os.MkdirAll(executionDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create execution directory: %w", err)
	}

	// For now, just return the directory path
	// In a real implementation, this would execute discovery commands via SSH
	return executionDir, nil
}
