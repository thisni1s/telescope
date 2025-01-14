#!/bin/sh
echo "started" > /root/config/teardownState.txt
sh /var/scripts/upload.sh > /dev/null 2>&1
systemctl stop tcpdumpd > /dev/null 2>&1
jq --arg deletion "$(date --iso-8601=seconds)" '. + {deletion: $deletion}' /root/config/descriptor.txt > /root/config/descriptor.txt
mc cp /root/config/descriptor.txt tupload/$(cat /root/config/bucket.txt)/$(cat /root/config/ip4.txt)/descriptor.txt
echo "finished" > /root/config/teardownState.txt

