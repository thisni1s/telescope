#!/bin/sh
echo "started" > /root/config/teardownState.txt
sh /var/scripts/upload.sh > /dev/null 2>&1
systemctl stop tcpdumpd > /dev/null 2>&1
echo "finished" > /root/config/teardownState.txt
