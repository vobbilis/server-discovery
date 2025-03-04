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
echo "  \"os_type\": \"linux\"," >> "$JSON_OUTPUT"
echo "  \"kernel_version\": \"$KERNEL_VERSION\"," >> "$JSON_OUTPUT"

# Get package manager information
log "Detecting package manager..."
if command -v dpkg > /dev/null; then
    PACKAGE_MANAGER="dpkg"
elif command -v rpm > /dev/null; then
    PACKAGE_MANAGER="rpm"
elif command -v pacman > /dev/null; then
    PACKAGE_MANAGER="pacman"
else
    PACKAGE_MANAGER="unknown"
fi
echo "  \"package_manager\": \"$PACKAGE_MANAGER\"," >> "$JSON_OUTPUT"

# Get init system
log "Detecting init system..."
if pidof systemd > /dev/null; then
    INIT_SYSTEM="systemd"
elif [ -f /etc/init.d/functions ]; then
    INIT_SYSTEM="sysvinit"
elif [ -f /etc/init/init.conf ]; then
    INIT_SYSTEM="upstart"
else
    INIT_SYSTEM="unknown"
fi
echo "  \"init_system\": \"$INIT_SYSTEM\"," >> "$JSON_OUTPUT"

# Get SELinux status
log "Checking SELinux status..."
if command -v getenforce > /dev/null; then
    SELINUX_STATUS=$(getenforce)
else
    SELINUX_STATUS="disabled"
fi
echo "  \"selinux_status\": \"$SELINUX_STATUS\"," >> "$JSON_OUTPUT"

# Get firewall status
log "Checking firewall status..."
if command -v firewall-cmd > /dev/null; then
    FIREWALL_STATUS=$(firewall-cmd --state)
elif command -v ufw > /dev/null; then
    FIREWALL_STATUS=$(ufw status | grep "Status" | cut -d: -f2 | tr -d ' ')
else
    FIREWALL_STATUS="unknown"
fi
echo "  \"firewall_status\": \"$FIREWALL_STATUS\"," >> "$JSON_OUTPUT"

# Get active users
log "Collecting active user information..."
echo "  \"active_users\": [" >> "$JSON_OUTPUT"
who -u | while read USER TTY LOGIN_TIME PID IDLE FROM; do
    echo "    {" >> "$JSON_OUTPUT"
    echo "      \"username\": \"$USER\"," >> "$JSON_OUTPUT"
    echo "      \"terminal\": \"$TTY\"," >> "$JSON_OUTPUT"
    echo "      \"login_time\": \"$LOGIN_TIME\"," >> "$JSON_OUTPUT"
    if [ ! -z "$FROM" ]; then
        echo "      \"from_host\": \"$FROM\"" >> "$JSON_OUTPUT"
    fi
    echo "    }," >> "$JSON_OUTPUT"
done
# Remove trailing comma and close array
sed -i '$ s/,$//' "$JSON_OUTPUT"
echo "  ]," >> "$JSON_OUTPUT"

# Get system load
log "Collecting system load information..."
LOAD1=$(cat /proc/loadavg | cut -d' ' -f1)
LOAD5=$(cat /proc/loadavg | cut -d' ' -f2)
LOAD15=$(cat /proc/loadavg | cut -d' ' -f3)
echo "  \"system_load\": {" >> "$JSON_OUTPUT"
echo "    \"load1\": $LOAD1," >> "$JSON_OUTPUT"
echo "    \"load5\": $LOAD5," >> "$JSON_OUTPUT"
echo "    \"load15\": $LOAD15" >> "$JSON_OUTPUT"
echo "  }," >> "$JSON_OUTPUT"

# Get detailed network interface information
log "Collecting network interface information..."
echo "  \"network_interfaces\": [" >> "$JSON_OUTPUT"
ip -j link show | jq -c '.[]' | while read IFACE; do
    NAME=$(echo "$IFACE" | jq -r '.ifname')
    MAC=$(echo "$IFACE" | jq -r '.address')
    MTU=$(echo "$IFACE" | jq -r '.mtu')
    STATE=$(echo "$IFACE" | jq -r '.operstate')
    
    # Get IP addresses for this interface
    IP_ADDRS=$(ip -j addr show dev "$NAME" | jq -r '.[].addr_info[].local' 2>/dev/null)
    
    # Get interface speed and duplex if available
    if [ -f "/sys/class/net/$NAME/speed" ]; then
        SPEED=$(cat "/sys/class/net/$NAME/speed")
    else
        SPEED="null"
    fi
    if [ -f "/sys/class/net/$NAME/duplex" ]; then
        DUPLEX=$(cat "/sys/class/net/$NAME/duplex")
    else
        DUPLEX="null"
    fi
    
    echo "    {" >> "$JSON_OUTPUT"
    echo "      \"name\": \"$NAME\"," >> "$JSON_OUTPUT"
    echo "      \"mac_address\": \"$MAC\"," >> "$JSON_OUTPUT"
    echo "      \"mtu\": $MTU," >> "$JSON_OUTPUT"
    echo "      \"state\": \"$STATE\"," >> "$JSON_OUTPUT"
    echo "      \"speed\": $SPEED," >> "$JSON_OUTPUT"
    [ "$DUPLEX" != "null" ] && echo "      \"duplex\": \"$DUPLEX\"," >> "$JSON_OUTPUT"
    echo "      \"ip_addresses\": [" >> "$JSON_OUTPUT"
    echo "$IP_ADDRS" | while read IP; do
        [ ! -z "$IP" ] && echo "        \"$IP\"," >> "$JSON_OUTPUT"
    done
    # Remove trailing comma and close array
    sed -i '$ s/,$//' "$JSON_OUTPUT"
    echo "      ]" >> "$JSON_OUTPUT"
    echo "    }," >> "$JSON_OUTPUT"
done
# Remove trailing comma and close array
sed -i '$ s/,$//' "$JSON_OUTPUT"
echo "  ]," >> "$JSON_OUTPUT"

# Get mounted filesystems
log "Collecting filesystem information..."
echo "  \"mounted_filesystems\": [" >> "$JSON_OUTPUT"
df -PT | tail -n +2 | while read DEVICE FSTYPE TOTAL USED FREE PCENT MOUNTPOINT; do
    # Convert sizes to GB
    TOTAL_GB=$(echo "scale=2; $TOTAL / 1024 / 1024" | bc)
    USED_GB=$(echo "scale=2; $USED / 1024 / 1024" | bc)
    FREE_GB=$(echo "scale=2; $FREE / 1024 / 1024" | bc)
    
    # Get mount options
    OPTIONS=$(mount | grep "^$DEVICE" | awk '{print $4}')
    
    # Get inode information
    INODES=$(df -i "$MOUNTPOINT" | tail -1)
    USED_INODES=$(echo "$INODES" | awk '{print $3}')
    FREE_INODES=$(echo "$INODES" | awk '{print $4}')
    
    echo "    {" >> "$JSON_OUTPUT"
    echo "      \"device\": \"$DEVICE\"," >> "$JSON_OUTPUT"
    echo "      \"mount_point\": \"$MOUNTPOINT\"," >> "$JSON_OUTPUT"
    echo "      \"fs_type\": \"$FSTYPE\"," >> "$JSON_OUTPUT"
    echo "      \"options\": \"$OPTIONS\"," >> "$JSON_OUTPUT"
    echo "      \"total_gb\": $TOTAL_GB," >> "$JSON_OUTPUT"
    echo "      \"used_gb\": $USED_GB," >> "$JSON_OUTPUT"
    echo "      \"free_gb\": $FREE_GB," >> "$JSON_OUTPUT"
    echo "      \"used_inodes\": $USED_INODES," >> "$JSON_OUTPUT"
    echo "      \"free_inodes\": $FREE_INODES" >> "$JSON_OUTPUT"
    echo "    }," >> "$JSON_OUTPUT"
done
# Remove trailing comma and close array
sed -i '$ s/,$//' "$JSON_OUTPUT"
echo "  ]" >> "$JSON_OUTPUT"

# Close the main JSON object
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
    echo "Package Manager: $PACKAGE_MANAGER"
    echo "Init System: $INIT_SYSTEM"
    echo "SELinux Status: $SELINUX_STATUS"
    echo "Firewall Status: $FIREWALL_STATUS"
    echo "Active Users:"
    who -u
    echo "==============================="
    echo "System Load:"
    cat /proc/loadavg
    echo "==============================="
    echo "Network Interfaces:"
    ip -br addr
    echo "==============================="
    echo "Mounted Filesystems:"
    df -PT
    echo "==============================="
} > "$SUMMARY_FILE"

log "Discovery completed successfully"
log "Output saved to $OUTPUT_PATH"
log "Summary file: $SUMMARY_FILE"
log "JSON data: $JSON_OUTPUT"

exit 0 