#!/bin/sh
# Starte den Server und installiere alles:
#
export DEBIAN_FRONTEND=noninteractive
apt update -y
apt upgrade -y
# da kommt frage ob sachen neu gestartet werden sollen
apt install tcpdump -y # already installed
apt install logrotate -y # already installed
#curl https://jn2p.de/tcpdump-systemd.txt > /usr/lib/systemd/system/tcpdumpd.service # get tcpdump service file
#curl https://jn2p.de/tcpdump.logrotate > /etc/logrotate.d/tcpdumpd
wget https://gist.githubusercontent.com/thisni1s/d0eda61e1a1fd819516ab98398129d99/raw/84e46c2dc242dfa1fd113f2beb0e32d2c32f5492/tcpdumpd.service -P /usr/lib/systemd/system
wget https://gist.githubusercontent.com/thisni1s/d0eda61e1a1fd819516ab98398129d99/raw/84e46c2dc242dfa1fd113f2beb0e32d2c32f5492/tcpdumpd -P /etc/logrotate.d
wget https://gist.githubusercontent.com/thisni1s/d0eda61e1a1fd819516ab98398129d99/raw/69dbe5ccfe0a506df1dff2bcaf81dddd49cf294c/upload.sh -P /root
chmod 644 /etc/logrotate.d/tcpdumpd
chown root:root /etc/logrotate.d/tcpdumpd
systemctl enable tcpdumpd
systemctl start tcpdumpd
#(crontab -l ; echo "0 * * * * sh /root/upload.sh") | crontab -
#reboot
