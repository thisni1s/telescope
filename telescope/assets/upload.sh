#!/bin/sh
systemctl stop tcpdumpd

# Directory containing .pcap files
directory="/var/log/"
bucket=$(cat /root/config/bucket.txt)
ip=$(cat /root/config/ip4.txt | sed -r 's/\./-/g' )

# Check if the directory exists
if [ ! -d "$directory" ]; then
    echo "Directory not found: $directory"
    exit 1
fi

cd $directory
mc cp --recursive tcpdumpd/ tupload/$bucket/$ip
rm tcpdumpd/*

systemctl start tcpdumpd
