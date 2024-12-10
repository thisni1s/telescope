#!/bin/sh

export DEBIAN_FRONTEND=noninteractive
apt update -y
apt upgrade -y
apt install tcpdump -y # already installed
wget https://github.com/thisni1s/telescope/raw/refs/heads/main/telescope/tcpdumpd.service -P /usr/lib/systemd/system
wget https://github.com/thisni1s/telescope/raw/refs/heads/main/telescope/upload.sh -P /root
systemctl enable tcpdumpd

#Install Minio Client
curl https://dl.min.io/client/mc/release/linux-amd64/mc \
  --create-dirs \
  -o /minio-binaries/mc

chmod +x /minio-binaries/mc
mv /minio-binaries/mc /usr/local/bin/

# Install Webhook Server
curl -L https://github.com/adnanh/webhook/releases/latest/download/webhook-linux-amd64.tar.gz \
    -o webhook.tar.gz
tar xf webhook.tar.gz
mv webhook-linux-amd64/webhook /usr/local/bin/webhook
rmdir webhook-linux-amd64

# Change SSH Port, Ubuntu has socket based activation so it needs to be set like this
mkdir -p /etc/systemd/system/ssh.socket.d
cat >/etc/systemd/system/ssh.socket.d/listen.conf <<EOF
[Socket]
ListenStream=
ListenStream=28763
EOF

ip=$( dig +short myip.opendns.com @resolver1.opendns.com | sed -r 's/\./-/g' )

bucket=$(cat /root/config/bucket.txt)

(crontab -l ; echo "0 */12 * * * sh /root/upload.sh ${bucket}") | crontab -

mc alias set tupload $(cat /root/config/storageLoc.txt) $(cat /root/config/storageAccKey.txt) $(cat /root/config/storageSecKey.txt)

systemctl start tcpdumpd
reboot
