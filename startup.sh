#!/bin/sh
# Starte den Server und installiere alles:
#
export DEBIAN_FRONTEND=noninteractive
apt update -y
apt upgrade -y
apt install tcpdump -y # already installed
apt install logrotate -y # already installed
wget https://raw.githubusercontent.com/thisni1s/telescope/main/tcpdumpd.service -P /usr/lib/systemd/system
wget https://raw.githubusercontent.com/thisni1s/telescope/main/tcpdumpd -P /etc/logrotate.d
wget https://raw.githubusercontent.com/thisni1s/telescope/main/upload.sh -P /root
chmod 644 /etc/logrotate.d/tcpdumpd
chown root:root /etc/logrotate.d/tcpdumpd
systemctl enable tcpdumpd
systemctl start tcpdumpd
#(crontab -l ; echo "0 * * * * sh /root/upload.sh") | crontab -
#reboot
