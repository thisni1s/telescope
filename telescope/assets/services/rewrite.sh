#!/bin/bash

# Directory containing .pcap files
directory="/var/spool/gotrace"
ip=$(cat /root/config/ip4.txt )
iface=$(ip route show default | awk '{print $5}')
localip=$(ip -4 addr show dev "${iface}" | grep 'inet ' | awk '{print $2}' | cut -d'/' -f1 | head -n 1)

cd $directory
for file in *.pcap.gz; do
    zcat "$file" \
    | tcprewrite --dstipmap="$localip":"$ip" --infile=- --outfile=- \
    | gzip > "${file}.rewrite"

    mv "${file}.rewrite" "$file"
done

