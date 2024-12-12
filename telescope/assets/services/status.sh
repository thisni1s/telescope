#!/bin/bash

# Define an array of systemd services to check
SERVICES=("ssh.service" "ssh.socket" "webhook.service" "webhook.socket" "tcpdumpd.service")

# Initialize an empty JSON object
status_json="{"

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

