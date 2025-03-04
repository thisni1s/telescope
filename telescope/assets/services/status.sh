#!/bin/bash

# Define an array of systemd services to check
SERVICES=("ssh.service" "ssh.socket" "webhook.service" "webhook.socket" "tcpdumpd.service" "corsaro.service" "gotrace.service")


#utime=$(uptime | jc --uptime) # does not work for some obscure reason
name=$(cat /etc/hostname)
ip4=$(cat /root/config/ip4.txt)
ip6=$(cat /root/config/ip6.txt)
td=$(cat /root/config/teardownState.txt)
otime=$(cat /root/config/otime.txt)
provider=$(cat /root/config/provider.txt)
region=$(cat /root/config/region.txt)
os=$(hostnamectl | grep Operating | cut -d ':' --fields 2 | tr -d ' ')

#status_json="{\"hostname\": \"$name\", \"uptime\": $utime, \"ipv4\": \"$ip4\",  \"teardown\": \"$td\", \"created\": \"$otime\", \"provider\": \"$provider\", \"region\": \"$region\", \"os\": \"$os\", "
status_json="{\"hostname\": \"$name\", \"ipv4\": \"$ip4\", \"ipv6\": \"$ip6\", \"teardown\": \"$td\", \"created\": \"$otime\", \"provider\": \"$provider\", \"region\": \"$region\", \"os\": \"$os\", "

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

