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
#systemctl start tcpdumpd

#Install Minio Client
curl https://dl.min.io/client/mc/release/linux-amd64/mc \
  --create-dirs \
  -o $HOME/minio-binaries/mc

chmod +x $HOME/minio-binaries/mc
export PATH=$PATH:$HOME/minio-binaries/
cat <<EOF >> /root/.bashrc
PATH=$PATH:$HOME/minio-binaries/
EOF


# Change SSH Port, Ubuntu has socket based activation so it needs to be set like this
mkdir -p /etc/systemd/system/ssh.socket.d
cat >/etc/systemd/system/ssh.socket.d/listen.conf <<EOF
[Socket]
ListenStream=
ListenStream=28763
EOF

ip=$( dig +short myip.opendns.com @resolver1.opendns.com | sed -r 's/\./-/g' )


#(crontab -l ; echo "0 * * * * sh /root/upload.sh") | crontab -
#reboot
