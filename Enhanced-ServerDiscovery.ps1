#Requires -RunAsAdministrator
#Requires -Version 5.1

<#
.SYNOPSIS
    Comprehensive Windows Server Discovery Script
.DESCRIPTION
    This script performs detailed discovery of Windows Server components including
    network connections, processes, services, hardware, and software inventory.
    It attempts to provide similar information to what BMC Discovery would collect.
.NOTES
    File Name      : Enhanced-ServerDiscovery.ps1
    Prerequisite   : PowerShell 5.1 or later, Administrator rights
.EXAMPLE
    .\Enhanced-ServerDiscovery.ps1
#>

# Script configuration
$ErrorActionPreference = "SilentlyContinue"
$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$outputFolder = "$env:USERPROFILE\Documents\ServerDiscovery-$timestamp"
$logFile = "$outputFolder\discovery_log.txt"

# Create output directory
if (-not (Test-Path $outputFolder)) {
    New-Item -ItemType Directory -Path $outputFolder | Out-Null
}

# Initialize logging function
function Write-Log {
    param (
        [Parameter(Mandatory = $true)]
        [string]$Message,
        
        [Parameter(Mandatory = $false)]
        [ValidateSet("INFO", "WARNING", "ERROR")]
        [string]$Level = "INFO"
    )
    
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $logMessage = "[$timestamp] [$Level] $Message"
    
    # Write to console with color
    switch ($Level) {
        "INFO" { Write-Host $logMessage -ForegroundColor Cyan }
        "WARNING" { Write-Host $logMessage -ForegroundColor Yellow }
        "ERROR" { Write-Host $logMessage -ForegroundColor Red }
    }
    
    # Write to log file
    Add-Content -Path $logFile -Value $logMessage
}

# Function to identify common services by port number
function Get-ServiceByPort {
    param (
        [Parameter(Mandatory = $true)]
        [int]$Port
    )

    $commonPorts = @{
        20 = "FTP (Data)"
        21 = "FTP (Control)"
        22 = "SSH"
        23 = "Telnet"
        25 = "SMTP"
        53 = "DNS"
        80 = "HTTP"
        88 = "Kerberos"
        110 = "POP3"
        123 = "NTP"
        135 = "RPC"
        137 = "NetBIOS Name Service"
        138 = "NetBIOS Datagram Service"
        139 = "NetBIOS Session Service"
        143 = "IMAP"
        161 = "SNMP"
        162 = "SNMP Trap"
        389 = "LDAP"
        443 = "HTTPS"
        445 = "SMB"
        464 = "Kerberos Change/Set password"
        465 = "SMTP over SSL"
        500 = "ISAKMP/IKE"
        514 = "Syslog"
        587 = "SMTP Submission"
        636 = "LDAPS"
        993 = "IMAP SSL"
        995 = "POP3 SSL"
        1433 = "SQL Server"
        1434 = "SQL Server Browser"
        1521 = "Oracle"
        1701 = "L2TP"
        1723 = "PPTP"
        3306 = "MySQL"
        3389 = "RDP"
        5060 = "SIP"
        5222 = "XMPP"
        5432 = "PostgreSQL"
        5985 = "WinRM HTTP"
        5986 = "WinRM HTTPS"
        8080 = "HTTP Alternate"
        8443 = "HTTPS Alternate"
        9389 = "Active Directory Web Services"
    }

    if ($commonPorts.ContainsKey($Port)) {
        return $commonPorts[$Port]
    } else {
        return "Unknown"
    }
}

# Start discovery process
Write-Log "Starting comprehensive server discovery process"
Write-Log "Results will be saved to: $outputFolder"

#region System Information
Write-Log "Collecting system information..."

# Basic system information
$systemInfo = Get-ComputerInfo
$systemInfo | ConvertTo-Json -Depth 5 | Out-File "$outputFolder\system_info.json"

# Hardware information
$hardware = @{
    ComputerSystem = Get-CimInstance -ClassName Win32_ComputerSystem | Select-Object *
    BIOS = Get-CimInstance -ClassName Win32_BIOS | Select-Object *
    Processor = Get-CimInstance -ClassName Win32_Processor | Select-Object *
    PhysicalMemory = Get-CimInstance -ClassName Win32_PhysicalMemory | Select-Object *
    DiskDrives = Get-CimInstance -ClassName Win32_DiskDrive | Select-Object *
    LogicalDisks = Get-CimInstance -ClassName Win32_LogicalDisk | Select-Object *
    VideoControllers = Get-CimInstance -ClassName Win32_VideoController | Select-Object *
}
$hardware | ConvertTo-Json -Depth 5 | Out-File "$outputFolder\hardware_info.json"

# Operating system details
$osInfo = Get-CimInstance -ClassName Win32_OperatingSystem | Select-Object *
$osInfo | ConvertTo-Json -Depth 5 | Out-File "$outputFolder\os_info.json"

# Windows features and roles
$windowsFeatures = Get-WindowsFeature | Where-Object { $_.Installed -eq $true }
$windowsFeatures | ConvertTo-Json -Depth 3 | Out-File "$outputFolder\windows_features.json"

Write-Log "System information collection complete"
#endregion

#region Network Information
Write-Log "Collecting network information..."

# Get all network interfaces
$networkAdapters = Get-NetAdapter | Select-Object *
$networkAdapters | ConvertTo-Json -Depth 5 | Out-File "$outputFolder\network_adapters.json"

# Get IP configuration
$ipConfiguration = Get-NetIPConfiguration -Detailed | Select-Object *
$ipConfiguration | ConvertTo-Json -Depth 5 | Out-File "$outputFolder\ip_configuration.json"

# Get IP addresses
$ipAddresses = Get-NetIPAddress | Select-Object *
$ipAddresses | ConvertTo-Json -Depth 5 | Out-File "$outputFolder\ip_addresses.json"

# Get network interface details
$networkInterfaces = @{}
Get-NetAdapter | ForEach-Object {
    $adapter = $_
    $ipAddresses = (Get-NetIPAddress -InterfaceIndex $adapter.ifIndex).IPAddress
    $ipConfig = Get-NetIPConfiguration -InterfaceIndex $adapter.ifIndex
    
    foreach ($ip in $ipAddresses) {
        if (-not $networkInterfaces.ContainsKey($ip)) {
            $networkInterfaces[$ip] = @{
                InterfaceName = $adapter.Name
                InterfaceDescription = $adapter.InterfaceDescription
                LinkSpeed = $adapter.LinkSpeed
                MediaType = $adapter.MediaType
                PhysicalAddress = $adapter.MacAddress
                Status = $adapter.Status
                AdminStatus = $adapter.AdminStatus
                MTU = $adapter.MtuSize
                PromiscuousMode = $adapter.PromiscuousMode
                DefaultGateway = $ipConfig.IPv4DefaultGateway.NextHop
                DNSServer = $ipConfig.DNSServer.ServerAddresses -join ", "
                DHCPEnabled = $ipConfig.NetIPv4Interface.Dhcp -eq "Enabled"
                ConnectionState = $adapter.ConnectorPresent
                VlanID = (Get-NetAdapterAdvancedProperty -InterfaceDescription $adapter.InterfaceDescription -DisplayName "VLAN ID" -ErrorAction SilentlyContinue).DisplayValue
                IPSubnet = (Get-NetIPAddress -InterfaceIndex $adapter.ifIndex -AddressFamily IPv4).PrefixLength
            }
        }
    }
}

# Add loopback interface information
$networkInterfaces["127.0.0.1"] = @{
    InterfaceName = "Loopback"
    InterfaceDescription = "Loopback Interface"
    LinkSpeed = "N/A"
    MediaType = "Loopback"
    PhysicalAddress = "N/A"
    Status = "Up"
    AdminStatus = "Up"
    MTU = 1500
    PromiscuousMode = $false
    DefaultGateway = "N/A"
    DNSServer = "N/A"
    DHCPEnabled = $false
    ConnectionState = $true
    VlanID = "N/A"
    IPSubnet = 8
}

$networkInterfaces["::1"] = @{
    InterfaceName = "Loopback IPv6"
    InterfaceDescription = "Loopback IPv6 Interface"
    LinkSpeed = "N/A"
    MediaType = "Loopback"
    PhysicalAddress = "N/A"
    Status = "Up"
    AdminStatus = "Up"
    MTU = 1500
    PromiscuousMode = $false
    DefaultGateway = "N/A"
    DNSServer = "N/A"
    DHCPEnabled = $false
    ConnectionState = $true
    VlanID = "N/A"
    IPSubnet = 128
}

$networkInterfaces | ConvertTo-Json -Depth 5 | Out-File "$outputFolder\network_interfaces_detailed.json"

# Get DNS client settings
$dnsSettings = Get-DnsClientServerAddress | Select-Object *
$dnsSettings | ConvertTo-Json -Depth 3 | Out-File "$outputFolder\dns_settings.json"

# Get routing table
$routingTable = Get-NetRoute | Select-Object *
$routingTable | ConvertTo-Json -Depth 3 | Out-File "$outputFolder\routing_table.json"

# Get firewall rules
$firewallRules = Get-NetFirewallRule | Select-Object Name, DisplayName, Description, Enabled, Direction, Action, Profile
$firewallRules | ConvertTo-Json -Depth 3 | Out-File "$outputFolder\firewall_rules.json"

# Get TCP connections
Write-Log "Collecting TCP connection data..."
$tcpConnections = Get-NetTCPConnection | Select-Object LocalAddress, LocalPort, RemoteAddress, RemotePort, State, OwningProcess, @{
    Name = "LocalService"
    Expression = { Get-ServiceByPort -Port $_.LocalPort }
}, @{
    Name = "RemoteService"
    Expression = { Get-ServiceByPort -Port $_.RemotePort }
}, @{
    Name = "InterfaceName"
    Expression = {
        if ($networkInterfaces.ContainsKey($_.LocalAddress)) {
            $networkInterfaces[$_.LocalAddress].InterfaceName
        } else {
            "Unknown"
        }
    }
}, @{
    Name = "InterfaceDescription"
    Expression = {
        if ($networkInterfaces.ContainsKey($_.LocalAddress)) {
            $networkInterfaces[$_.LocalAddress].InterfaceDescription
        } else {
            "Unknown"
        }
    }
}

$tcpConnections | ConvertTo-Json -Depth 3 | Out-File "$outputFolder\tcp_connections.json"
$tcpConnections | Export-Csv -Path "$outputFolder\tcp_connections.csv" -NoTypeInformation

# Get UDP endpoints
$udpEndpoints = Get-NetUDPEndpoint | Select-Object LocalAddress, LocalPort, OwningProcess, @{
    Name = "LocalService"
    Expression = { Get-ServiceByPort -Port $_.LocalPort }
}, @{
    Name = "InterfaceName"
    Expression = {
        if ($networkInterfaces.ContainsKey($_.LocalAddress)) {
            $networkInterfaces[$_.LocalAddress].InterfaceName
        } else {
            "Unknown"
        }
    }
}, @{
    Name = "InterfaceDescription"
    Expression = {
        if ($networkInterfaces.ContainsKey($_.LocalAddress)) {
            $networkInterfaces[$_.LocalAddress].InterfaceDescription
        } else {
            "Unknown"
        }
    }
}

$udpEndpoints | ConvertTo-Json -Depth 3 | Out-File "$outputFolder\udp_endpoints.json"
$udpEndpoints | Export-Csv -Path "$outputFolder\udp_endpoints.csv" -NoTypeInformation

Write-Log "Network information collection complete"
#endregion

#region Process Information
Write-Log "Collecting process information..."

# Get all processes with detailed information
$processes = @{}
$processDetails = Get-Process | ForEach-Object {
    $process = $_
    $processes[$process.Id] = $process
    
    # Get process module information
    $modules = $process.Modules | Select-Object ModuleName, FileName, FileVersion, ProductVersion
    
    # Get command line if available
    try {
        $commandLine = (Get-CimInstance -ClassName Win32_Process -Filter "ProcessId = $($process.Id)").CommandLine
    } catch {
        $commandLine = "Unable to retrieve"
    }
    
    # Create process object with detailed information
    [PSCustomObject]@{
        ProcessId = $process.Id
        ProcessName = $process.Name
        Path = $process.Path
        Company = $process.Company
        ProductVersion = $process.ProductVersion
        FileVersion = $process.FileVersion
        Description = $process.Description
        StartTime = $process.StartTime
        CPU = [math]::Round($process.CPU, 2)
        WorkingSetMB = [math]::Round($process.WorkingSet / 1MB, 2)
        VirtualMemoryMB = [math]::Round($process.VirtualMemorySize / 1MB, 2)
        PagedMemoryMB = [math]::Round($process.PagedMemorySize / 1MB, 2)
        Threads = $process.Threads.Count
        Handles = $process.HandleCount
        CommandLine = $commandLine
        Modules = $modules
        ParentProcessId = (Get-CimInstance -ClassName Win32_Process -Filter "ProcessId = $($process.Id)").ParentProcessId
    }
}

$processDetails | ConvertTo-Json -Depth 4 | Out-File "$outputFolder\process_details.json"
$processDetails | Select-Object ProcessId, ProcessName, Path, StartTime, CPU, WorkingSetMB, VirtualMemoryMB, Threads, Handles, ParentProcessId, CommandLine | 
    Export-Csv -Path "$outputFolder\process_details.csv" -NoTypeInformation

# Create process-to-connection mapping
$processConnections = @{}

# Map TCP connections to processes
Get-NetTCPConnection | ForEach-Object {
    $connection = $_
    $processId = $connection.OwningProcess
    
    if (-not $processConnections.ContainsKey($processId)) {
        $processConnections[$processId] = @{
            ProcessId = $processId
            ProcessName = if ($processes.ContainsKey($processId)) { $processes[$processId].Name } else { "Unknown" }
            TCPConnections = @()
            UDPEndpoints = @()
        }
    }
    
    $processConnections[$processId].TCPConnections += [PSCustomObject]@{
        LocalAddress = $connection.LocalAddress
        LocalPort = $connection.LocalPort
        RemoteAddress = $connection.RemoteAddress
        RemotePort = $connection.RemotePort
        State = $connection.State
        LocalService = Get-ServiceByPort -Port $connection.LocalPort
        RemoteService = Get-ServiceByPort -Port $connection.RemotePort
    }
}

# Map UDP endpoints to processes
Get-NetUDPEndpoint | ForEach-Object {
    $endpoint = $_
    $processId = $endpoint.OwningProcess
    
    if (-not $processConnections.ContainsKey($processId)) {
        $processConnections[$processId] = @{
            ProcessId = $processId
            ProcessName = if ($processes.ContainsKey($processId)) { $processes[$processId].Name } else { "Unknown" }
            TCPConnections = @()
            UDPEndpoints = @()
        }
    }
    
    $processConnections[$processId].UDPEndpoints += [PSCustomObject]@{
        LocalAddress = $endpoint.LocalAddress
        LocalPort = $endpoint.LocalPort
        LocalService = Get-ServiceByPort -Port $endpoint.LocalPort
    }
}

$processConnections.Values | ConvertTo-Json -Depth 5 | Out-File "$outputFolder\process_connections.json"

Write-Log "Process information collection complete"
#endregion

#region Service Information
Write-Log "Collecting service information..."

# Get all services with detailed information
$services = Get-Service | ForEach-Object {
    $service = $_
    
    # Get Win32_Service information for additional details
    $serviceDetails = Get-CimInstance -ClassName Win32_Service -Filter "Name = '$($service.Name)'"
    
    # Get service dependencies
    $dependencies = $service.DependentServices | Select-Object Name, DisplayName, Status
    $serviceDependsOn = $service.ServicesDependedOn | Select-Object Name, DisplayName, Status
    
    # Create service object with detailed information
    [PSCustomObject]@{
        Name = $service.Name
        DisplayName = $service.DisplayName
        Status = $service.Status
        StartType = $service.StartType
        Description = $serviceDetails.Description
        PathName = $serviceDetails.PathName
        StartName = $serviceDetails.StartName
        ProcessId = $serviceDetails.ProcessId
        Dependencies = $dependencies
        DependsOn = $serviceDependsOn
        DelayedAutoStart = $serviceDetails.DelayedAutoStart
        ServiceType = $serviceDetails.ServiceType
        ExitCode = $serviceDetails.ExitCode
        InstallDate = $serviceDetails.InstallDate
    }
}

$services | ConvertTo-Json -Depth 4 | Out-File "$outputFolder\service_details.json"
$services | Select-Object Name, DisplayName, Status, StartType, Description, PathName, StartName, ProcessId | 
    Export-Csv -Path "$outputFolder\service_details.csv" -NoTypeInformation

# Map services to processes
$serviceProcessMapping = @{}
$services | Where-Object { $_.ProcessId -gt 0 } | ForEach-Object {
    $serviceProcessMapping[$_.ProcessId] = @{
        ProcessId = $_.ProcessId
        Services = @()
    }
}

$services | Where-Object { $_.ProcessId -gt 0 } | ForEach-Object {
    $serviceProcessMapping[$_.ProcessId].Services += [PSCustomObject]@{
        Name = $_.Name
        DisplayName = $_.DisplayName
        Status = $_.Status
        StartType = $_.StartType
    }
}

$serviceProcessMapping.Values | ConvertTo-Json -Depth 4 | Out-File "$outputFolder\service_process_mapping.json"

Write-Log "Service information collection complete"
#endregion

#region Software Inventory
Write-Log "Collecting software inventory..."

# Get installed software from registry
$installedSoftware = @()
$uninstallKeys = @(
    "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\*",
    "HKLM:\SOFTWARE\Wow6432Node\Microsoft\Windows\CurrentVersion\Uninstall\*"
)

foreach ($key in $uninstallKeys) {
    $installedSoftware += Get-ItemProperty $key | 
        Where-Object { $_.DisplayName -ne $null } | 
        Select-Object DisplayName, DisplayVersion, Publisher, InstallDate, InstallLocation, @{
            Name = "Architecture"
            Expression = { if ($key -like "*Wow6432Node*") { "32-bit" } else { "64-bit" } }
        }
}

$installedSoftware | Sort-Object DisplayName | ConvertTo-Json -Depth 3 | Out-File "$outputFolder\installed_software.json"
$installedSoftware | Sort-Object DisplayName | Export-Csv -Path "$outputFolder\installed_software.csv" -NoTypeInformation

# Get Windows updates
$windowsUpdates = Get-HotFix | Select-Object HotFixID, Description, InstalledOn, InstalledBy
$windowsUpdates | ConvertTo-Json -Depth 3 | Out-File "$outputFolder\windows_updates.json"
$windowsUpdates | Export-Csv -Path "$outputFolder\windows_updates.csv" -NoTypeInformation

Write-Log "Software inventory collection complete"
#endregion

#region Configuration Information
Write-Log "Collecting configuration information..."

# Get scheduled tasks
$scheduledTasks = Get-ScheduledTask | Select-Object TaskName, TaskPath, State, Description, Author, @{
    Name = "Actions"
    Expression = { $_.Actions | Select-Object Execute, Arguments }
}, @{
    Name = "Triggers"
    Expression = { $_.Triggers | Select-Object * }
}
$scheduledTasks | ConvertTo-Json -Depth 4 | Out-File "$outputFolder\scheduled_tasks.json"

# Get shared folders
$sharedFolders = Get-CimInstance -ClassName Win32_Share | Select-Object Name, Path, Description, Type, @{
    Name = "AccessMask"
    Expression = { $_.AccessMask }
}, @{
    Name = "AllowMaximum"
    Expression = { $_.AllowMaximum }
}
$sharedFolders | ConvertTo-Json -Depth 3 | Out-File "$outputFolder\shared_folders.json"
$sharedFolders | Export-Csv -Path "$outputFolder\shared_folders.csv" -NoTypeInformation

# Get environment variables
$environmentVariables = @{
    System = [Environment]::GetEnvironmentVariables("Machine") | ConvertTo-Json
    User = [Environment]::GetEnvironmentVariables("User") | ConvertTo-Json
    Process = [Environment]::GetEnvironmentVariables("Process") | ConvertTo-Json
}
$environmentVariables | ConvertTo-Json -Depth 3 | Out-File "$outputFolder\environment_variables.json"

# Get important registry settings
$registrySettings = @{
    ComputerName = Get-ItemProperty "HKLM:\SYSTEM\CurrentControlSet\Control\ComputerName\ComputerName" | Select-Object ComputerName
    WindowsVersion = Get-ItemProperty "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion" | Select-Object ProductName, ReleaseId, CurrentBuild, UBR
    NetworkSettings = Get-ItemProperty "HKLM:\SYSTEM\CurrentControlSet\Services\Tcpip\Parameters" | Select-Object Hostname, Domain, SearchList, NV Domain
}
$registrySettings | ConvertTo-Json -Depth 4 | Out-File "$outputFolder\registry_settings.json"

Write-Log "Configuration information collection complete"
#endregion

#region Dependency Mapping
Write-Log "Generating dependency mappings..."

# Create process dependency map
$processDependencies = @()
foreach ($proc in $processDetails) {
    $parentId = $proc.ParentProcessId
    if ($parentId -and $processes.ContainsKey($parentId)) {
        $processDependencies += [PSCustomObject]@{
            SourceProcess = [PSCustomObject]@{
                ProcessId = $parentId
                ProcessName = $processes[$parentId].Name
            }
            TargetProcess = [PSCustomObject]@{
                ProcessId = $proc.ProcessId
                ProcessName = $proc.ProcessName
            }
            DependencyType = "Parent-Child"
        }
    }
}

# Create service dependency map
foreach ($service in $services) {
    foreach ($dep in $service.DependsOn) {
        $processDependencies += [PSCustomObject]@{
            SourceService = [PSCustomObject]@{
                Name = $service.Name
                DisplayName = $service.DisplayName
            }
            TargetService = [PSCustomObject]@{
                Name = $dep.Name
                DisplayName = $dep.DisplayName
            }
            DependencyType = "Service-Dependency"
        }
    }
}

# Create network connection dependency map
foreach ($conn in $tcpConnections) {
    if ($conn.RemoteAddress -ne "0.0.0.0" -and $conn.RemoteAddress -ne "::" -and $conn.RemoteAddress -ne "127.0.0.1" -and $conn.RemoteAddress -ne "::1") {
        $sourceProcess = $null
        if ($processes.ContainsKey($conn.OwningProcess)) {
            $sourceProcess = [PSCustomObject]@{
                ProcessId = $conn.OwningProcess
                ProcessName = $processes[$conn.OwningProcess].Name
            }
        }
        
        $processDependencies += [PSCustomObject]@{
            SourceProcess = $sourceProcess
            TargetEndpoint = [PSCustomObject]@{
                Address = $conn.RemoteAddress
                Port = $conn.RemotePort
                Service = $conn.RemoteService
            }
            DependencyType = "Network-Connection"
            Protocol = "TCP"
            State = $conn.State
        }
    }
}

$processDependencies | ConvertTo-Json -Depth 4 | Out-File "$outputFolder\dependency_map.json"

Write-Log "Dependency mapping complete"
#endregion

#region Summary Report
Write-Log "Generating summary report..."

# Create HTML summary report
$htmlReport = @"
<!DOCTYPE html>
<html>
<head>
    <title>Server Discovery Report - $(hostname)</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        h1 { color: #0066cc; }
        h2 { color: #0099cc; border-bottom: 1px solid #ddd; padding-bottom: 5px; }
        table { border-collapse: collapse; width: 100%; margin-bottom: 20px; }
        th, td { text-align: left; padding: 8px; border: 1px solid #ddd; }
        th { background-color: #f2f2f2; }
        tr:nth-child(even) { background-color: #f9f9f9; }
        .summary { background-color: #e6f2ff; padding: 15px; border-radius: 5px; margin-bottom: 20px; }
    </style>
</head>
<body>
    <h1>Server Discovery Report</h1>
    <div class="summary">
        <p><strong>Server Name:</strong> $(hostname)</p>
        <p><strong>Report Generated:</strong> $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")</p>
        <p><strong>Operating System:</strong> $($osInfo.Caption) $($osInfo.Version)</p>
    </div>

    <h2>System Summary</h2>
    <table>
        <tr><th>Property</th><th>Value</th></tr>
        <tr><td>Manufacturer</td><td>$($hardware.ComputerSystem.Manufacturer)</td></tr>
        <tr><td>Model</td><td>$($hardware.ComputerSystem.Model)</td></tr>
        <tr><td>Processors</td><td>$($hardware.Processor.Count) x $($hardware.Processor[0].Name)</td></tr>
        <tr><td>Physical Memory</td><td>$([math]::Round($hardware.ComputerSystem.TotalPhysicalMemory / 1GB, 2)) GB</td></tr>
        <tr><td>Domain</td><td>$($hardware.ComputerSystem.Domain)</td></tr>
    </table>

    <h2>Network Summary</h2>
    <table>
        <tr><th>Interface</th><th>IP Address</th><th>Subnet</th><th>Gateway</th><th>Status</th></tr>
        $(
            $ipConfiguration | ForEach-Object {
                $gateway = if ($_.IPv4DefaultGateway) { $_.IPv4DefaultGateway.NextHop } else { "N/A" }
                $ipv4 = $_.IPv4Address.IPAddress
                if ($ipv4) {
                    "<tr><td>$($_.InterfaceAlias)</td><td>$ipv4</td><td>$($_.IPv4Address.PrefixLength)</td><td>$gateway</td><td>$($_.NetAdapter.Status)</td></tr>"
                }
            }
        )
    </table>

    <h2>Process Summary</h2>
    <table>
        <tr><th>Process Name</th><th>PID</th><th>Memory (MB)</th><th>CPU</th><th>Start Time</th></tr>
        $(
            $processDetails | Sort-Object -Property WorkingSetMB -Descending | Select-Object -First 20 | ForEach-Object {
                "<tr><td>$($_.ProcessName)</td><td>$($_.ProcessId)</td><td>$($_.WorkingSetMB)</td><td>$($_.CPU)</td><td>$($_.StartTime)</td></tr>"
            }
        )
    </table>

    <h2>Service Summary</h2>
    <table>
        <tr><th>Service Name</th><th>Display Name</th><th>Status</th><th>Start Type</th></tr>
        $(
            $services | Where-Object { $_.Status -eq "Running" } | Sort-Object -Property Name | Select-Object -First 20 | ForEach-Object {
                "<tr><td>$($_.Name)</td><td>$($_.DisplayName)</td><td>$($_.Status)</td><td>$($_.StartType)</td></tr>"
            }
        )
    </table>

    <h2>TCP Connections Summary</h2>
    <table>
        <tr><th>Process</th><th>Local Address</th><th>Local Port</th><th>Remote Address</th><th>Remote Port</th><th>State</th></tr>
        $(
            $tcpConnections | Where-Object { $_.RemoteAddress -ne "0.0.0.0" -and $_.RemoteAddress -ne "::" } | 
            Sort-Object -Property OwningProcess | Select-Object -First 20 | ForEach-Object {
                $processName = if ($processes.ContainsKey($_.OwningProcess)) { $processes[$_.OwningProcess].Name } else { "Unknown" }
                "<tr><td>$processName</td><td>$($_.LocalAddress)</td><td>$($_.LocalPort)</td><td>$($_.RemoteAddress)</td><td>$($_.RemotePort)</td><td>$($_.State)</td></tr>"
            }
        )
    </table>

    <h2>Software Summary</h2>
    <table>
        <tr><th>Name</th><th>Version</th><th>Publisher</th><th>Architecture</th></tr>
        $(
            $installedSoftware | Sort-Object -Property DisplayName | Select-Object -First 20 | ForEach-Object {
                "<tr><td>$($_.DisplayName)</td><td>$($_.DisplayVersion)</td><td>$($_.Publisher)</td><td>$($_.Architecture)</td></tr>"
            }
        )
    </table>

    <h2>Shared Folders Summary</h2>
    <table>
        <tr><th>Name</th><th>Path</th><th>Description</th></tr>
        $(
            $sharedFolders | ForEach-Object {
                "<tr><td>$($_.Name)</td><td>$($_.Path)</td><td>$($_.Description)</td></tr>"
            }
        )
    </table>

    <h2>Discovery Information</h2>
    <table>
        <tr><th>Category</th><th>Count</th></tr>
        <tr><td>Network Adapters</td><td>$($networkAdapters.Count)</td></tr>
        <tr><td>TCP Connections</td><td>$($tcpConnections.Count)</td></tr>
        <tr><td>UDP Endpoints</td><td>$($udpEndpoints.Count)</td></tr>
        <tr><td>Running Processes</td><td>$($processDetails.Count)</td></tr>
        <tr><td>Services</td><td>$($services.Count)</td></tr>
        <tr><td>Running Services</td><td>$($services | Where-Object { $_.Status -eq "Running" } | Measure-Object).Count</td></tr>
        <tr><td>Installed Software</td><td>$($installedSoftware.Count)</td></tr>
        <tr><td>Shared Folders</td><td>$($sharedFolders.Count)</td></tr>
        <tr><td>Firewall Rules</td><td>$($firewallRules.Count)</td></tr>
    </table>
</body>
</html>
"@

$htmlReport | Out-File "$outputFolder\discovery_report.html"

# Open the HTML report
Write-Log "Opening HTML report..."
Invoke-Item "$outputFolder\discovery_report.html"

Write-Log "Summary report generation complete"
#endregion

# Create a ZIP archive of all discovery files
Write-Log "Creating ZIP archive of discovery results..."
$zipFilePath = "$env:USERPROFILE\Documents\ServerDiscovery-$timestamp.zip"

# Check if .NET Framework 4.5+ is available for ZipFile class
if ([Environment]::Version.Major -ge 4) {
    Add-Type -AssemblyName System.IO.Compression.FileSystem
    [System.IO.Compression.ZipFile]::CreateFromDirectory($outputFolder, $zipFilePath)
} else {
    # Fallback to PowerShell compression
    Compress-Archive -Path "$outputFolder\*" -DestinationPath $zipFilePath
}

Write-Log "ZIP archive created at: $zipFilePath"

# Final summary
Write-Log "Discovery process complete" -Level "INFO"
Write-Log "Results saved to: $outputFolder" -Level "INFO"
Write-Log "ZIP archive: $zipFilePath" -Level "INFO"
Write-Log "HTML report: $outputFolder\discovery_report.html" -Level "INFO"

# Return the paths for further processing if needed
return @{
    OutputFolder = $outputFolder
    ZipFile = $zipFilePath
    HtmlReport = "$outputFolder\discovery_report.html"
}