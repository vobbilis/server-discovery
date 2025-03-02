#!/bin/bash
# Enhanced-ServerDiscovery.sh
# A comprehensive Linux server discovery script that collects detailed system information
# Usage: ./Enhanced-ServerDiscovery.sh [output_directory]

# Set default output directory
OUTPUT_DIR="${1:-./server_discovery_output}"
HOSTNAME=$(hostname)
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
OUTPUT_PATH="${OUTPUT_DIR}/${HOSTNAME}_${TIMESTAMP}"

# Create output directory
mkdir -p "$OUTPUT_PATH"

# Log file for script execution
LOG_FILE="${OUTPUT_PATH}/discovery.log"

# Function to log messages
log() {
    local message="$1"
    local timestamp=$(date +"%Y-%m-%d %H:%M:%S")
    echo "[$timestamp] $message" | tee -a "$LOG_FILE"
}

# Function to handle errors
handle_error() {
    local message="$1"
    log "ERROR: $message"
}

# Start discovery
log "Starting Linux server discovery for $HOSTNAME"
log "Output will be saved to $OUTPUT_PATH"

# Create JSON output file
JSON_OUTPUT="${OUTPUT_PATH}/server_details.json"
echo "{" > "$JSON_OUTPUT"

# Get basic system information
log "Collecting basic system information..."
OS_NAME=$(cat /etc/os-release | grep "PRETTY_NAME" | cut -d= -f2 | tr -d '"')
OS_VERSION=$(cat /etc/os-release | grep "VERSION_ID" | cut -d= -f2 | tr -d '"')
KERNEL_VERSION=$(uname -r)

echo "  \"hostname\": \"$HOSTNAME\"," >> "$JSON_OUTPUT"
echo "  \"os_name\": \"$OS_NAME\"," >> "$JSON_OUTPUT"
echo "  \"os_version\": \"$OS_VERSION\"," >> "$JSON_OUTPUT"
echo "  \"kernel_version\": \"$KERNEL_VERSION\"," >> "$JSON_OUTPUT"

# Get CPU information
log "Collecting CPU information..."
CPU_MODEL=$(grep "model name" /proc/cpuinfo | head -1 | cut -d: -f2 | sed 's/^[ \t]*//')
CPU_COUNT=$(grep -c "processor" /proc/cpuinfo)

echo "  \"cpu_model\": \"$CPU_MODEL\"," >> "$JSON_OUTPUT"
echo "  \"cpu_count\": $CPU_COUNT," >> "$JSON_OUTPUT"

# Get memory information
log "Collecting memory information..."
MEM_TOTAL_KB=$(grep "MemTotal" /proc/meminfo | awk '{print $2}')
MEM_TOTAL_GB=$(echo "scale=2; $MEM_TOTAL_KB / 1024 / 1024" | bc)

echo "  \"memory_total_gb\": $MEM_TOTAL_GB," >> "$JSON_OUTPUT"

# Get disk information
log "Collecting disk information..."
DISK_INFO=$(df -h / | tail -1)
DISK_TOTAL=$(echo "$DISK_INFO" | awk '{print $2}' | tr -d 'G')
DISK_FREE=$(echo "$DISK_INFO" | awk '{print $4}' | tr -d 'G')

echo "  \"disk_total_gb\": $DISK_TOTAL," >> "$JSON_OUTPUT"
echo "  \"disk_free_gb\": $DISK_FREE," >> "$JSON_OUTPUT"

# Get last boot time
log "Collecting boot time information..."
LAST_BOOT=$(uptime -s)

echo "  \"last_boot_time\": \"$LAST_BOOT\"," >> "$JSON_OUTPUT"

# Get IP addresses
log "Collecting network interface information..."
echo "  \"ip_addresses\": [" >> "$JSON_OUTPUT"

IP_ADDRESSES=$(ip -j addr | jq -c '.[] | select(.operstate=="UP") | {interface: .ifname, addresses: [.addr_info[] | select(.family=="inet" or .family=="inet6") | {ip_address: .local, family: .family}]}')
FORMATTED_IP=$(echo "$IP_ADDRESSES" | jq -c '.' | sed 's/$/,/' | sed '$ s/,$//')

echo "$FORMATTED_IP" >> "$JSON_OUTPUT"
echo "  ]," >> "$JSON_OUTPUT"

# Get installed software
log "Collecting installed software information..."
echo "  \"installed_software\": [" >> "$JSON_OUTPUT"

if command -v dpkg > /dev/null; then
    # Debian/Ubuntu
    SOFTWARE=$(dpkg-query -W -f='{"name": "${Package}", "version": "${Version}", "install_date": ""},\n')
elif command -v rpm > /dev/null; then
    # RHEL/CentOS/Fedora
    SOFTWARE=$(rpm -qa --queryformat '{"name": "%{NAME}", "version": "%{VERSION}", "install_date": "%{INSTALLTIME:date}"},\n')
else
    SOFTWARE=""
    handle_error "Unable to determine package manager"
fi

# Remove trailing comma from last entry
SOFTWARE=$(echo "$SOFTWARE" | sed '$ s/,$//')
echo "$SOFTWARE" >> "$JSON_OUTPUT"
echo "  ]," >> "$JSON_OUTPUT"

# Get running services
log "Collecting running services information..."
echo "  \"running_services\": [" >> "$JSON_OUTPUT"

if command -v systemctl > /dev/null; then
    # systemd
    SERVICES=$(systemctl list-units --type=service --state=running --no-legend | awk '{print $1}' | while read service; do
        status=$(systemctl show -p ActiveState --value "$service")
        start_mode=$(systemctl show -p UnitFileState --value "$service")
        echo "{\"name\": \"$service\", \"display_name\": \"$service\", \"status\": \"$status\", \"start_mode\": \"$start_mode\"},"
    done)
elif [ -f /etc/init.d/functions ]; then
    # SysV init (RHEL/CentOS)
    SERVICES=$(service --status-all 2>&1 | grep -E "running|stopped" | awk '{print $1, $2}' | while read service status; do
        echo "{\"name\": \"$service\", \"display_name\": \"$service\", \"status\": \"$status\", \"start_mode\": \"unknown\"},"
    done)
else
    SERVICES=""
    handle_error "Unable to determine service manager"
fi

# Remove trailing comma from last entry
SERVICES=$(echo "$SERVICES" | sed '$ s/,$//')
echo "$SERVICES" >> "$JSON_OUTPUT"
echo "  ]," >> "$JSON_OUTPUT"

# Define common ports and their descriptions
declare -A COMMON_PORTS
COMMON_PORTS[20]="FTP (Data)"
COMMON_PORTS[21]="FTP (Control)"
COMMON_PORTS[22]="SSH"
COMMON_PORTS[23]="Telnet"
COMMON_PORTS[25]="SMTP"
COMMON_PORTS[53]="DNS"
COMMON_PORTS[80]="HTTP"
COMMON_PORTS[88]="Kerberos"
COMMON_PORTS[110]="POP3"
COMMON_PORTS[123]="NTP"
COMMON_PORTS[135]="MSRPC"
COMMON_PORTS[137]="NetBIOS Name Service"
COMMON_PORTS[138]="NetBIOS Datagram Service"
COMMON_PORTS[139]="NetBIOS Session Service"
COMMON_PORTS[143]="IMAP"
COMMON_PORTS[389]="LDAP"
COMMON_PORTS[443]="HTTPS"
COMMON_PORTS[445]="SMB"
COMMON_PORTS[464]="Kerberos Change/Set password"
COMMON_PORTS[465]="SMTP over SSL"
COMMON_PORTS[500]="ISAKMP/IKE"
COMMON_PORTS[514]="Syslog"
COMMON_PORTS[587]="SMTP (Submission)"
COMMON_PORTS[636]="LDAPS"
COMMON_PORTS[993]="IMAPS"
COMMON_PORTS[995]="POP3S"
COMMON_PORTS[1433]="Microsoft SQL Server"
COMMON_PORTS[1434]="Microsoft SQL Monitor"
COMMON_PORTS[1521]="Oracle Database"
COMMON_PORTS[3306]="MySQL"
COMMON_PORTS[3389]="RDP"
COMMON_PORTS[5060]="SIP"
COMMON_PORTS[5222]="XMPP"
COMMON_PORTS[5432]="PostgreSQL"
COMMON_PORTS[5985]="WinRM HTTP"
COMMON_PORTS[5986]="WinRM HTTPS"
COMMON_PORTS[8080]="HTTP Alternate"
COMMON_PORTS[8443]="HTTPS Alternate"
COMMON_PORTS[49152]="Windows RPC"

# Get open ports and network connections
log "Collecting network connection information..."
echo "  \"open_ports\": [" >> "$JSON_OUTPUT"

# Use ss command (modern replacement for netstat)
if command -v ss > /dev/null; then
    # Get listening ports
    LISTENING_PORTS=$(ss -tuln | grep LISTEN | awk '{print $5}' | while read addr; do
        local_ip=$(echo "$addr" | cut -d: -f1)
        local_port=$(echo "$addr" | cut -d: -f2)
        
        # Get process info
        pid=$(ss -tulnp | grep "$addr" | grep -oP "pid=\K[0-9]+")
        if [ -n "$pid" ]; then
            process_name=$(ps -p "$pid" -o comm= 2>/dev/null || echo "unknown")
        else
            process_name="unknown"
            pid=0
        fi
        
        # Get description
        description="${COMMON_PORTS[$local_port]:-Unknown}"
        
        echo "{\"localPort\": $local_port, \"localIP\": \"$local_ip\", \"state\": \"LISTENING\", \"description\": \"$description\", \"processID\": $pid, \"processName\": \"$process_name\"},"
    done)
    
    # Get established connections
    ESTABLISHED_CONNS=$(ss -tun | grep ESTAB | awk '{print $5, $6}' | while read local remote; do
        local_ip=$(echo "$local" | cut -d: -f1)
        local_port=$(echo "$local" | cut -d: -f2)
        remote_ip=$(echo "$remote" | cut -d: -f1)
        remote_port=$(echo "$remote" | cut -d: -f2)
        
        # Get process info
        pid=$(ss -tunp | grep "$local" | grep -oP "pid=\K[0-9]+")
        if [ -n "$pid" ]; then
            process_name=$(ps -p "$pid" -o comm= 2>/dev/null || echo "unknown")
        else
            process_name="unknown"
            pid=0
        fi
        
        # Get description
        description="${COMMON_PORTS[$local_port]:-Unknown}"
        
        echo "{\"localPort\": $local_port, \"localIP\": \"$local_ip\", \"remotePort\": $remote_port, \"remoteIP\": \"$remote_ip\", \"state\": \"ESTABLISHED\", \"description\": \"$description\", \"processID\": $pid, \"processName\": \"$process_name\"},"
    done)
    
    # Combine and remove trailing comma from last entry
    PORTS="$LISTENING_PORTS$ESTABLISHED_CONNS"
    PORTS=$(echo "$PORTS" | sed '$ s/,$//')
    echo "$PORTS" >> "$JSON_OUTPUT"
else
    handle_error "ss command not found"
fi

echo "  ]" >> "$JSON_OUTPUT"
echo "}" >> "$JSON_OUTPUT"

# Validate JSON output
if command -v jq > /dev/null; then
    if jq empty "$JSON_OUTPUT" 2>/dev/null; then
        log "JSON output validated successfully"
    else
        handle_error "JSON validation failed"
        # Try to fix JSON
        log "Attempting to fix JSON..."
        cat "$JSON_OUTPUT" | jq '.' > "${JSON_OUTPUT}.fixed" 2>/dev/null
        if [ $? -eq 0 ]; then
            mv "${JSON_OUTPUT}.fixed" "$JSON_OUTPUT"
            log "JSON fixed successfully"
        else
            log "Could not fix JSON, output may be invalid"
        fi
    fi
else
    log "WARNING: jq not installed, skipping JSON validation"
fi

# Create a summary file
SUMMARY_FILE="${OUTPUT_PATH}/summary.txt"
{
    echo "Linux Server Discovery Summary"
    echo "==============================="
    echo "Hostname: $HOSTNAME"
    echo "OS: $OS_NAME $OS_VERSION"
    echo "Kernel: $KERNEL_VERSION"
    echo "CPU: $CPU_MODEL ($CPU_COUNT cores)"
    echo "Memory: ${MEM_TOTAL_GB}GB"
    echo "Disk: ${DISK_TOTAL}GB total, ${DISK_FREE}GB free"
    echo "Last Boot: $LAST_BOOT"
    echo "==============================="
    echo "IP Addresses:"
    ip -br addr | grep -v DOWN
    echo "==============================="
    echo "Listening Ports:"
    ss -tuln | grep LISTEN
    echo "==============================="
    echo "Established Connections:"
    ss -tun | grep ESTAB
    echo "==============================="
    echo "Running Services (top 10):"
    if command -v systemctl > /dev/null; then
        systemctl list-units --type=service --state=running --no-legend | head -10
    else
        service --status-all 2>&1 | grep running | head -10
    fi
    echo "==============================="
    echo "Top Processes by CPU:"
    ps aux --sort=-%cpu | head -11
    echo "==============================="
    echo "Top Processes by Memory:"
    ps aux --sort=-%mem | head -11
    echo "==============================="
} > "$SUMMARY_FILE"

log "Discovery completed successfully"
log "Output saved to $OUTPUT_PATH"
log "Summary file: $SUMMARY_FILE"
log "JSON data: $JSON_OUTPUT"

exit 0 