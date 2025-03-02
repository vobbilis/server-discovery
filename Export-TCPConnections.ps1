# Export-TCPConnections.ps1
# This script exports detailed TCP connection information

param(
    [string]$OutputFolder = ".\tcp_connections"
)

# Create output directory if it doesn't exist
if (-not (Test-Path $OutputFolder)) {
    New-Item -Path $OutputFolder -ItemType Directory -Force | Out-Null
}

# Get TCP connections
$connections = @()
try {
    $netstatOutput = netstat -ano
    foreach ($line in $netstatOutput) {
        if ($line -match '^\s*(TCP|UDP)\s+(\S+):(\d+)\s+(\S+):(\d+|\*)\s+(\w+)?\s*(\d+)?') {
            $protocol = $matches[1]
            $localAddress = $matches[2]
            $localPort = [int]$matches[3]
            $remoteAddress = $matches[4]
            $remotePort = if ($matches[5] -eq '*') { 0 } else { [int]$matches[5] }
            $state = if ($protocol -eq 'TCP') { $matches[6] } else { 'N/A' }
            $pid = if ($matches[7]) { [int]$matches[7] } else { 0 }
            
            # Get process name and details
            $processName = 'Unknown'
            $processPath = 'Unknown'
            $processCompany = 'Unknown'
            
            try {
                if ($pid -gt 0) {
                    $process = Get-Process -Id $pid -ErrorAction SilentlyContinue
                    if ($process) {
                        $processName = $process.Name
                        try {
                            $processPath = $process.Path
                            $fileInfo = Get-ItemProperty -Path $processPath -ErrorAction SilentlyContinue
                            if ($fileInfo) {
                                $versionInfo = $fileInfo.VersionInfo
                                $processCompany = $versionInfo.CompanyName
                            }
                        } catch {}
                    }
                }
            } catch {}
            
            $connections += [PSCustomObject]@{
                Timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
                Protocol = $protocol
                LocalAddress = $localAddress
                LocalPort = $localPort
                RemoteAddress = $remoteAddress
                RemotePort = $remotePort
                State = $state
                ProcessID = $pid
                ProcessName = $processName
                ProcessPath = $processPath
                ProcessCompany = $processCompany
            }
        }
    }
} catch {
    Write-Warning "Error getting network connections: $_"
}

# Export to CSV
$connections | Export-Csv -Path "$OutputFolder\tcp_connections.csv" -NoTypeInformation

# Export to JSON
$connections | ConvertTo-Json -Depth 5 | Out-File -FilePath "$OutputFolder\tcp_connections.json" -Encoding UTF8

# Return success message
Write-Host "TCP connections exported successfully. Output saved to $OutputFolder" 