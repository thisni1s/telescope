#!/bin/bash

# Check for the correct number of arguments
if [ "$#" -lt 2 ]; then
    echo "Usage: $0 <file_with_ipv6_list> <path_parameter_string>"
    exit 1
fi

# Assign arguments to variables
ipv6_list_file=$1
path_param=$2
predefined_path="/hooks/redeploy"

# Read the IP addresses from the file and make HTTP requests
while IFS= read -r ipv6; do
    if [[ -n "$ipv6" ]]; then
        echo "Sending request to [$ipv6]:51337$predefined_path?token=$path_param"
        # Make HTTP request using curl, ignoring SSL certificate errors
        curl -k -g --silent --show-error "https://[$ipv6]:51337$predefined_path?token=$path_param"
        # (Optional) add a new line for separating responses
        echo
    fi
done < "$ipv6_list_file"
