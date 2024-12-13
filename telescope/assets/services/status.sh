#!/bin/bash

# Define an array of systemd services to check
SERVICES=("ssh.service" "ssh.socket" "webhook.service" "webhook.socket" "tcpdumpd.service")


utime=$(uptime | jc --uptime)
name=$(cat /etc/hostname)
ip4=$( dig +short myip.opendns.com @resolver1.opendns.com)
td=$(cat /root/config/teardownState.txt)
otime=$(cat /root/config/otime.txt)
provider=$(cat /root/config/provider.txt)
region=$(cat /root/config/region.txt)
os=$(hostnamectl | grep Operating | cut -d ':' --fields 2 | tr -d ' ')

status_json="{\"hostname\": \"$name\", \"uptime\": $utime, \"ipv4\": \"$ip4\",  \"teardown\": \"$td\", \"created\": \"$otime\", \"provider\": \"$provider\", \"region\": \"$region\", \"os\": \"$os\", "

# Loop through each service and get its status
for service in "${SERVICES[@]}"; do
    # Get the status of the service using systemctl
    status=$(systemctl is-active "$service")

    # Add the service status to the JSON object
    status_json+="\"$service\": \"$status\","
done

# Remove the trailing comma and close the JSON object
status_json="${status_json%,}}"

# Output the JSON
echo "$status_json"

