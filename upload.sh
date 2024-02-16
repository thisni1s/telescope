#!/bin/sh

# Directory containing .pcap files
directory="/var/log/tcpdumpd"
#directory=$(pwd)
link=$1
token=$2
ip=$( dig +short myip.opendns.com @resolver1.opendns.com | sed -r 's/\./-/g' )

systemctl stop tcpdumpd

# Check if the directory exists
if [ ! -d "$directory" ]; then
    echo "Directory not found: $directory"
    exit 1
fi

cd $directory
# Iterate over .pcap files in the directory
for file in *.pcap; do
    # Check if the file is a regular file
    if [ -f "$file" ]; then
        echo "Processing file: $file"
        target=$link/$ip/$file
        curl -k -T $file -u "$2:" -H 'X-Requested-With: XMLHttpRequest' "$target" 
        rm $file
    fi
done

systemctl start tcpdumpd

