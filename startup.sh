#!/bin/sh
# Starte den Server und installiere alles:
#
export DEBIAN_FRONTEND=noninteractive
apt update -y
apt upgrade -y
# da kommt frage ob sachen neu gestartet werden sollen
apt install tcpdump -y # already installed
apt install logrotate -y # already installed
curl https://jn2p.de/tcpdump-systemd.txt > /usr/lib/systemd/system/tcpdumpd.service # get tcpdump service file
curl https://jn2p.de/tcpdump.logrotate > /etc/logrotate.d/tcpdumpd
chmod 644 /etc/logrotate.d/tcpdumpd
chown root:root /etc/logrotate.d/tcpdumpd
systemctl enable tcpdumpd
systemctl start tcpdumpd
reboot
