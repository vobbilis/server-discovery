package scripts

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/vobbilis/codegen/server-discovery/pkg/models"
)

// ServerConfig represents a server configuration
type ServerConfig struct {
	Hostname string
	Region   string
	Tags     []models.Tag
}

// Helper to create sample files
func createSampleFiles(outputPath string, server ServerConfig) {
	// Create HTML report
	htmlContent := fmt.Sprintf(`
        <!DOCTYPE html>
        <html>
        <head>
            <title>Discovery Report for %s</title>
            <style>body{font-family:Arial}</style>
        </head>
        <body>
            <h1>Server Discovery Report</h1>
            <h2>%s (%s)</h2>
            <p>Region: %s</p>
            <table>
                <tr><th>Property</th><th>Value</th></tr>
                <tr><td>Hostname</td><td>%s</td></tr>
                <tr><td>Region</td><td>%s</td></tr>
            </table>
            <h3>Tags</h3>
            <ul>
    `, server.Hostname, server.Hostname, server.Region, server.Region, server.Hostname, server.Region)

	for _, tag := range server.Tags {
		htmlContent += fmt.Sprintf("<li>%s: %s</li>", tag.TagName, tag.TagValue)
	}

	htmlContent += `
            </ul>
        </body>
        </html>
    `

	os.WriteFile(filepath.Join(outputPath, "discovery_report.html"), []byte(htmlContent), 0644)

	// Create JSON data
	serverData := struct {
		Hostname string       `json:"hostname"`
		Region   string       `json:"region"`
		Tags     []models.Tag `json:"tags"`
	}{
		Hostname: server.Hostname,
		Region:   server.Region,
		Tags:     server.Tags,
	}

	jsonData, _ := json.MarshalIndent(serverData, "", "  ")
	os.WriteFile(filepath.Join(outputPath, "server_data.json"), jsonData, 0644)
}
