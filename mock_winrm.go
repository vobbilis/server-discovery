package main

import (
	"io"

	"github.com/stretchr/testify/mock"
)

// WinRMClientInterface defines the interface for a WinRM client
type WinRMClientInterface interface {
	Run(command string, stdout, stderr io.Writer) (int, error)
}

// MockWinRMClient is a mock implementation of the WinRM client interface
type MockWinRMClient struct {
	mock.Mock
}

// Run mocks the Run method of the WinRM client
func (m *MockWinRMClient) Run(command string, stdout, stderr io.Writer) (int, error) {
	args := m.Called(command, stdout, stderr)

	// Write to stdout if provided
	if output := args.String(3); output != "" && stdout != nil {
		stdout.Write([]byte(output))
	}

	// Write to stderr if provided
	if errOutput := args.String(4); errOutput != "" && stderr != nil {
		stderr.Write([]byte(errOutput))
	}

	return args.Int(0), args.Error(1)
}
