#!/bin/sh
rm /var/scripts/*
rm /usr/lib/systemd/system/webhook.s*
rm /usr/lib/systemd/system/tcpdumpd.service 
rm /etc/webhook.json 
rm /etc/systemd/system/ssh.socket.d/listen.conf 
rm /usr/local/bin/*
rm -rf /root/config/
rm -rf /root/.mc
rm -rf /var/log/tcpdumpd
crontab -r
systemctl daemon-reload

