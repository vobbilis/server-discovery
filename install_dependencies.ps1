Write-Host "Installing dependencies for Server Discovery Controller..." -ForegroundColor Green

# Core dependencies
Write-Host "Installing core dependencies..." -ForegroundColor Cyan
go get github.com/masterzen/winrm
go get github.com/patrickmn/go-cache
go get github.com/juju/ratelimit

# Metrics and monitoring
Write-Host "Installing metrics and monitoring dependencies..." -ForegroundColor Cyan
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promauto
go get github.com/prometheus/client_golang/prometheus/promhttp

# System information
Write-Host "Installing system information dependencies..." -ForegroundColor Cyan
go get github.com/shirou/gopsutil/v3/cpu
go get github.com/shirou/gopsutil/v3/mem

# Tracing
Write-Host "Installing tracing dependencies..." -ForegroundColor Cyan
go get go.opentelemetry.io/otel
go get go.opentelemetry.io/otel/exporters/jaeger
go get go.opentelemetry.io/otel/sdk/resource
go get go.opentelemetry.io/otel/sdk/trace
go get go.opentelemetry.io/otel/semconv/v1.17.0
go get go.opentelemetry.io/otel/attribute
go get go.opentelemetry.io/otel/trace

# Testing
Write-Host "Installing testing dependencies..." -ForegroundColor Cyan
go get github.com/stretchr/testify/assert
go get github.com/stretchr/testify/mock

# Clean up dependencies
Write-Host "Cleaning up dependencies..." -ForegroundColor Cyan
go mod tidy

Write-Host "All dependencies installed successfully!" -ForegroundColor Green
Write-Host "You can now run your tests with: go test -v" -ForegroundColor Yellow 