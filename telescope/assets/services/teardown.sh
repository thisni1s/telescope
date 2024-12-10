#!/bin/sh
sh /var/scripts/upload.sh $(cat /root/config/bucket.txt)
systemctl stop tcpdumpd
