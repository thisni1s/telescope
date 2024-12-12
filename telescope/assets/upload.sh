#!/bin/sh

# Directory containing .pcap files
directory="/var/log/"
bucket=$(cat /root/config/bucket.txt)
ip=$( dig +short myip.opendns.com @resolver1.opendns.com | sed -r 's/\./-/g' )

systemctl stop tcpdumpd

# Check if the directory exists
if [ ! -d "$directory" ]; then
    echo "Directory not found: $directory"
    exit 1
fi

cd $directory
mc cp --recursive tcpdumpd/ tupload/$bucket/$ip
rm tcpdumpd/*

systemctl start tcpdumpd
