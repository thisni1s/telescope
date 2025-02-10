#!/bin/sh

mkdir /root/config
echo $1 > /root/config/bucket.txt
echo $2 > /root/config/storageLoc.txt
echo $3 > /root/config/storageAccKey.txt
echo $4 > /root/config/storageSecKey.txt
echo $5 > /root/config/webhookPw.txt
echo $6 > /root/config/provider.txt
echo $7 > /root/config/region.txt
echo "available" > /root/config/teardownState.txt
openssl req -x509 -newkey ed25519 -keyout /root/config/key.key -outform PEM -out /root/config/cert.pem -days 365 -nodes -subj "/C=DE/ST=NW/L=Muenster/O=Univeristy of Muenster/OU=NetSec Group/CN=$(cat /etc/hostname)"

export DEBIAN_FRONTEND=noninteractive
apt update -y
apt upgrade -y

# Install Corsaro
curl https://pkg.caida.org/os/ubuntu/bootstrap.sh | bash
sudo apt install -y corsaro
wget https://raw.githubusercontent.com/thisni1s/script-store/refs/heads/main/telescope/corsaro.conf -O /etc/corsaro.conf
wget https://raw.githubusercontent.com/thisni1s/script-store/refs/heads/main/telescope/corsaro.service -P /usr/lib/systemd/system
iface=$(ip route show default | awk '{print $5}')
sed -i "s/##IFACE##/$iface/g" /etc/corsaro.conf
systemctl enable corsaro

wget https://raw.githubusercontent.com/thisni1s/telescope/refs/heads/main/telescope/assets/services/webhook.service -P /usr/lib/systemd/system
wget https://raw.githubusercontent.com/thisni1s/telescope/refs/heads/main/telescope/assets/services/webhook.socket -P /usr/lib/systemd/system

mkdir -p /var/scripts
wget https://raw.githubusercontent.com/thisni1s/script-store/refs/heads/main/telescope/upload.sh -O /var/scripts/upload.sh
wget https://raw.githubusercontent.com/thisni1s/telescope/refs/heads/main/telescope/assets/services/teardown.sh -P /var/scripts/
wget https://raw.githubusercontent.com/thisni1s/telescope/refs/heads/main/telescope/assets/services/ping.sh -P /var/scripts/
wget https://raw.githubusercontent.com/thisni1s/telescope/refs/heads/main/telescope/assets/services/restart.sh -P /var/scripts/
wget https://raw.githubusercontent.com/thisni1s/telescope/refs/heads/main/telescope/assets/services/status.sh -P /var/scripts/
chmod +x /var/scripts/*

wget https://raw.githubusercontent.com/thisni1s/telescope/refs/heads/main/telescope/assets/services/webhook.json -P /etc/
sed -i "s/##WHPW##/$5/g" /etc/webhook.json

#Install Minio Client
curl https://dl.min.io/client/mc/release/linux-amd64/mc \
  --create-dirs \
  -o /minio-binaries/mc

chmod +x /minio-binaries/mc
mv /minio-binaries/mc /usr/local/bin/

# Install Webhook Server
curl -L https://github.com/adnanh/webhook/releases/latest/download/webhook-linux-amd64.tar.gz -o webhook.tar.gz
tar xf webhook.tar.gz
mv webhook-linux-amd64/webhook /usr/local/bin/webhook
rmdir webhook-linux-amd64
rm webhook.tar.gz
systemctl enable webhook.socket
systemctl start webhook.socket

# Change SSH Port, Ubuntu has socket based activation so it needs to be set like this
mkdir -p /etc/systemd/system/ssh.socket.d
cat >/etc/systemd/system/ssh.socket.d/listen.conf <<EOF
[Socket]
ListenStream=
ListenStream=28763
EOF

bucket=$(cat /root/config/bucket.txt)

(crontab -l ; echo '0 */12 * * * /var/scripts/upload.sh') | crontab -

mc alias set tupload $(cat /root/config/storageLoc.txt) $(cat /root/config/storageAccKey.txt) $(cat /root/config/storageSecKey.txt)

name=$(cat /etc/hostname)
ip4=$(dig +short myip.opendns.com @resolver1.opendns.com)
ip6=$(dig -6 +short @resolver1.opendns.com myip.opendns.com ANY)
otime=$(date --iso-8601=seconds)
os=$(hostnamectl | grep Operating | cut -d ':' --fields 2 | tr -d ' ')

desc="descriptor-$(date +"%y%m%d%H%M").txt"
echo "{\"hostname\": \"$name\", \"provider\": \"$6\", \"ipv4\": \"$ip4\", \"ipv6\": \"$ip6\", \"creation\": \"$otime\", \"os\": \"$os\", \"region\": \"$7\"}" > /root/config/$desc
echo $otime > /root/config/otime.txt
echo $ip4 > /root/config/ip4.txt
echo $ip6 > /root/config/ip6.txt

ip=$( echo $ip4 | sed -r 's/\./-/g' )

mc cp /root/config/$desc tupload/$(cat /root/config/bucket.txt)/$ip/$desc

systemctl daemon-reload
systemctl restart ssh.socket
systemctl enable --now corsaro
