package main

import "github.com/vobbilis/codegen/server-discovery/pkg/models"

// ServerDiscoverer interface defines methods for server discovery
type ServerDiscoverer interface {
	ExecuteDiscovery(server models.ServerConfig, outputDir string) (models.DiscoveryResult, error)
	ParseDiscoveryOutput(outputPath string) (models.ServerDetails, error)
}
