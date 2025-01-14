#!/bin/sh
echo "started" > /root/config/teardownState.txt
sh /var/scripts/upload.sh > /dev/null 2>&1
systemctl stop tcpdumpd > /dev/null 2>&1
desc=$(ls /root/config | grep descriptor)
jq --arg deletion "$(date --iso-8601=seconds)" '. + {deletion: $deletion}' /root/config/$desc > /root/config/temp.txt
mv /root/config/temp.txt /root/config/$desc
ip=$( cat /root/config/ip4.txt | sed -r 's/\./-/g' )
mc cp /root/config/$desc tupload/$(cat /root/config/bucket.txt)/$ip/$desc
sleep 5
echo "finished" > /root/config/teardownState.txt

