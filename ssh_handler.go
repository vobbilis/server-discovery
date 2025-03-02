package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHConfig represents the configuration for an SSH connection
type SSHConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	KeyFile  string
}

// RunLinuxDiscovery executes the discovery script on a Linux server via SSH
func RunLinuxDiscovery(config SSHConfig, outputDir string) (string, error) {
	// Create SSH client configuration
	clientConfig := &ssh.ClientConfig{
		User:            config.Username,
		Auth:            []ssh.AuthMethod{},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	// Add password authentication if provided
	if config.Password != "" {
		clientConfig.Auth = append(clientConfig.Auth, ssh.Password(config.Password))
	}

	// Add key-based authentication if provided
	if config.KeyFile != "" {
		key, err := ioutil.ReadFile(config.KeyFile)
		if err != nil {
			return "", fmt.Errorf("unable to read private key: %v", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return "", fmt.Errorf("unable to parse private key: %v", err)
		}

		clientConfig.Auth = append(clientConfig.Auth, ssh.PublicKeys(signer))
	}

	// Connect to the SSH server
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	client, err := ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		return "", fmt.Errorf("failed to connect to SSH server: %v", err)
	}
	defer client.Close()

	// Create a session
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create SSH session: %v", err)
	}
	defer session.Close()

	// Set up pipes for stdout and stderr
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	// Create a temporary directory for the script
	tempDir := fmt.Sprintf("/tmp/server_discovery_%d", time.Now().Unix())
	mkdirCmd := fmt.Sprintf("mkdir -p %s", tempDir)
	if err := runSSHCommand(client, mkdirCmd); err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %v", err)
	}

	// Upload the discovery script
	scriptContent, err := ioutil.ReadFile("Enhanced-ServerDiscovery.sh")
	if err != nil {
		return "", fmt.Errorf("failed to read discovery script: %v", err)
	}

	remotePath := filepath.Join(tempDir, "Enhanced-ServerDiscovery.sh")
	if err := uploadFile(client, remotePath, scriptContent); err != nil {
		return "", fmt.Errorf("failed to upload discovery script: %v", err)
	}

	// Make the script executable
	chmodCmd := fmt.Sprintf("chmod +x %s", remotePath)
	if err := runSSHCommand(client, chmodCmd); err != nil {
		return "", fmt.Errorf("failed to make script executable: %v", err)
	}

	// Run the discovery script
	cmd := fmt.Sprintf("cd %s && ./Enhanced-ServerDiscovery.sh %s", tempDir, tempDir)
	if err := session.Run(cmd); err != nil {
		return "", fmt.Errorf("failed to run discovery script: %v\nStderr: %s", err, stderr.String())
	}

	// Download the results
	outputPath := filepath.Join(outputDir, fmt.Sprintf("%s_%s", config.Host, time.Now().Format("20060102_150405")))
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %v", err)
	}

	// Download the JSON output
	jsonPath := filepath.Join(tempDir, "server_details.json")
	jsonContent, err := downloadFile(client, jsonPath)
	if err != nil {
		return "", fmt.Errorf("failed to download JSON output: %v", err)
	}

	localJsonPath := filepath.Join(outputPath, "server_details.json")
	if err := ioutil.WriteFile(localJsonPath, jsonContent, 0644); err != nil {
		return "", fmt.Errorf("failed to write JSON output: %v", err)
	}

	// Download the summary file
	summaryPath := filepath.Join(tempDir, "summary.txt")
	summaryContent, err := downloadFile(client, summaryPath)
	if err != nil {
		return "", fmt.Errorf("failed to download summary file: %v", err)
	}

	localSummaryPath := filepath.Join(outputPath, "summary.txt")
	if err := ioutil.WriteFile(localSummaryPath, summaryContent, 0644); err != nil {
		return "", fmt.Errorf("failed to write summary file: %v", err)
	}

	// Clean up the temporary directory
	cleanupCmd := fmt.Sprintf("rm -rf %s", tempDir)
	if err := runSSHCommand(client, cleanupCmd); err != nil {
		log.Printf("Warning: failed to clean up temporary directory: %v", err)
	}

	return outputPath, nil
}

// Helper function to run a command via SSH
func runSSHCommand(client *ssh.Client, cmd string) error {
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	return session.Run(cmd)
}

// Helper function to upload a file via SCP
func uploadFile(client *ssh.Client, remotePath string, content []byte) error {
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()

		fmt.Fprintf(w, "C0644 %d %s\n", len(content), filepath.Base(remotePath))
		w.Write(content)
		fmt.Fprint(w, "\x00")
	}()

	return session.Run(fmt.Sprintf("scp -t %s", filepath.Dir(remotePath)))
}

// Helper function to download a file via SCP
func downloadFile(client *ssh.Client, remotePath string) ([]byte, error) {
	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	var buf bytes.Buffer
	session.Stdout = &buf

	if err := session.Run(fmt.Sprintf("cat %s", remotePath)); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
