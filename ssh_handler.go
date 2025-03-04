package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/vobbilis/codegen/server-discovery/pkg/models"
)

// SSHConnectionPool manages SSH client connections
type SSHConnectionPool struct {
	clients     map[string]*ssh.Client
	mutex       sync.Mutex
	maxSize     int
	idleTimeout time.Duration
	lastUsed    map[string]time.Time
}

// NewSSHConnectionPool creates a new SSH connection pool
func NewSSHConnectionPool(maxSize int, idleTimeout time.Duration) *SSHConnectionPool {
	return &SSHConnectionPool{
		clients:     make(map[string]*ssh.Client),
		lastUsed:    make(map[string]time.Time),
		maxSize:     maxSize,
		idleTimeout: idleTimeout,
	}
}

// GetClient gets or creates an SSH client for the given config
func (p *SSHConnectionPool) GetClient(config models.SSHConfig) (*ssh.Client, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	key := fmt.Sprintf("%s:%d", config.Host, config.Port)

	// Check if we have an existing client
	if client, exists := p.clients[key]; exists {
		if time.Since(p.lastUsed[key]) > p.idleTimeout {
			// Client has been idle too long, close and remove it
			client.Close()
			delete(p.clients, key)
			delete(p.lastUsed, key)
		} else {
			// Update last used time and return existing client
			p.lastUsed[key] = time.Now()
			return client, nil
		}
	}

	// Create new client if we have room
	if len(p.clients) >= p.maxSize {
		// Remove oldest client
		var oldestKey string
		var oldestTime time.Time
		for k, t := range p.lastUsed {
			if oldestKey == "" || t.Before(oldestTime) {
				oldestKey = k
				oldestTime = t
			}
		}
		if oldestKey != "" {
			p.clients[oldestKey].Close()
			delete(p.clients, oldestKey)
			delete(p.lastUsed, oldestKey)
		}
	}

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
	if config.PrivateKeyPath != "" {
		key, err := ioutil.ReadFile(config.PrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("unable to read private key: %v", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("unable to parse private key: %v", err)
		}

		clientConfig.Auth = append(clientConfig.Auth, ssh.PublicKeys(signer))
	}

	// Connect to the SSH server
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	client, err := ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SSH server: %v", err)
	}

	// Store the new client
	p.clients[key] = client
	p.lastUsed[key] = time.Now()

	return client, nil
}

// CloseAll closes all connections in the pool
func (p *SSHConnectionPool) CloseAll() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for key, client := range p.clients {
		client.Close()
		delete(p.clients, key)
		delete(p.lastUsed, key)
	}
}

// Initialize SSH connection pool
var sshPool *SSHConnectionPool

func init() {
	sshPool = NewSSHConnectionPool(10, 10*time.Minute)
}

// RunLinuxDiscovery executes the discovery script on a Linux server via SSH
func RunLinuxDiscovery(config models.SSHConfig, outputDir string) (string, error) {
	// Get client from pool
	client, err := sshPool.GetClient(config)
	if err != nil {
		return "", fmt.Errorf("failed to get SSH client: %v", err)
	}

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

// SSHClient represents an SSH client connection
type SSHClient struct {
	config     *models.SSHConfig
	client     *ssh.Client
	lastActive time.Time
}

// NewSSHClient creates a new SSH client with the given configuration
func NewSSHClient(config *models.SSHConfig) (*SSHClient, error) {
	sshConfig := &ssh.ClientConfig{
		User: config.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(config.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Duration(config.TimeoutSeconds) * time.Second,
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port), sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %v", err)
	}

	return &SSHClient{
		config:     config,
		client:     client,
		lastActive: time.Now(),
	}, nil
}

// Execute runs a command on the remote server
func (c *SSHClient) Execute(command string) (string, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	var stdoutBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stdoutBuf

	err = session.Run(command)
	if err != nil {
		return "", fmt.Errorf("failed to run command: %v", err)
	}

	return stdoutBuf.String(), nil
}

// Close closes the SSH connection
func (c *SSHClient) Close() error {
	return c.client.Close()
}
