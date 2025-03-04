# Get system information
$os = Get-WmiObject -Class Win32_OperatingSystem
$cpu = Get-WmiObject -Class Win32_Processor
$memory = Get-WmiObject -Class Win32_ComputerSystem
$disk = Get-WmiObject -Class Win32_LogicalDisk -Filter "DriveType=3"
$network = Get-WmiObject -Class Win32_NetworkAdapterConfiguration | Where-Object { $_.IPEnabled -eq $true }
$services = Get-Service

# Create result object
$result = @{
    os = @{
        name = $os.Caption
        version = $os.Version
        architecture = $os.OSArchitecture
        lastBoot = $os.LastBootUpTime
    }
    cpu = @{
        model = $cpu.Name
        cores = $cpu.NumberOfCores
        threads = $cpu.NumberOfLogicalProcessors
        usage = (Get-Counter '\Processor(_Total)\% Processor Time').CounterSamples.CookedValue
    }
    memory = @{
        total = [math]::Round($memory.TotalPhysicalMemory / 1GB, 2)
        free = [math]::Round($os.FreePhysicalMemory / 1MB, 2)
        used = [math]::Round(($memory.TotalPhysicalMemory - $os.FreePhysicalMemory) / 1GB, 2)
    }
    disk = @{
        drives = @()
    }
    network = @{
        interfaces = @()
    }
    services = @()
}

# Add disk information
foreach ($drive in $disk) {
    $result.disk.drives += @{
        drive = $drive.DeviceID
        total = [math]::Round($drive.Size / 1GB, 2)
        free = [math]::Round($drive.FreeSpace / 1GB, 2)
        used = [math]::Round(($drive.Size - $drive.FreeSpace) / 1GB, 2)
    }
}

# Add network information
foreach ($adapter in $network) {
    $result.network.interfaces += @{
        name = $adapter.Description
        ipAddresses = $adapter.IPAddress
        macAddress = $adapter.MACAddress
        gateway = $adapter.DefaultIPGateway
        subnet = $adapter.IPSubnet
    }
}

# Add service information
foreach ($service in $services) {
    $result.services += @{
        name = $service.Name
        displayName = $service.DisplayName
        status = $service.Status
        startType = $service.StartType
    }
}

# Convert to JSON and output
$result | ConvertTo-Json -Depth 10 