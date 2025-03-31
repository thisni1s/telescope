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
apt install tcpdump curl unzip tcpreplay -y
#apt upgrade -y

# Install Corsaro
#curl https://pkg.caida.org/os/ubuntu/bootstrap.sh | bash
#sudo apt install -y corsaro
#wget https://raw.githubusercontent.com/thisni1s/script-store/refs/heads/main/telescope/corsaro.conf -O /etc/corsaro.conf
#wget https://raw.githubusercontent.com/thisni1s/script-store/refs/heads/main/telescope/corsaro.service -P /usr/lib/systemd/system

curl -sSL https://zivgitlab.uni-muenster.de/nkempen/gotrace/-/jobs/artifacts/main/download?job=build -o gotrace.zip
unzip -o gotrace.zip
chmod +x gotrace
mv gotrace /usr/local/bin

mkdir -p /etc/gotrace
mkdir -p /var/spool/gotrace
wget https://zivgitlab.uni-muenster.de/nkempen/gotrace/-/raw/main/gotrace.service -O /usr/lib/systemd/system/gotrace.service
wget https://zivgitlab.uni-muenster.de/nkempen/gotrace/-/raw/main/config.yaml -O /etc/gotrace/config.yaml

iface=$(ip route show default | awk '{print $5}')
sed -i "s/##IFACE##/$iface/g" /etc/gotrace/config.yaml
systemctl enable gotrace

wget https://raw.githubusercontent.com/thisni1s/telescope/refs/heads/main/telescope/assets/services/webhook.service -O /usr/lib/systemd/system/webhook.service
wget https://raw.githubusercontent.com/thisni1s/telescope/refs/heads/main/telescope/assets/services/webhook.socket -O /usr/lib/systemd/system/webhook.socket

mkdir -p /var/scripts
wget https://raw.githubusercontent.com/thisni1s/telescope/refs/heads/main/telescope/assets/upload.sh -O /var/scripts/upload.sh
wget https://raw.githubusercontent.com/thisni1s/telescope/refs/heads/main/telescope/assets/services/teardown.sh -O /var/scripts/teardown.sh
wget https://raw.githubusercontent.com/thisni1s/telescope/refs/heads/main/telescope/assets/services/ping.sh -O /var/scripts/ping.sh
wget https://raw.githubusercontent.com/thisni1s/telescope/refs/heads/main/telescope/assets/services/restart.sh -O /var/scripts/restart.sh
wget https://raw.githubusercontent.com/thisni1s/telescope/refs/heads/main/telescope/assets/services/status.sh -O /var/scripts/status.sh
wget https://raw.githubusercontent.com/thisni1s/telescope/refs/heads/main/telescope/assets/services/redeploy.sh -O /var/scripts/redeploy.sh
wget https://raw.githubusercontent.com/thisni1s/telescope/refs/heads/main/telescope/assets/services/rewrite.sh -O /var/scripts/rewrite.sh
chmod +x /var/scripts/*

wget https://raw.githubusercontent.com/thisni1s/telescope/refs/heads/main/telescope/assets/services/webhook.json -O /etc/webhook.json
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

(crontab -l ; echo '*/10 * * * * /var/scripts/upload.sh') | crontab -

mc alias set tupload $(cat /root/config/storageLoc.txt) $(cat /root/config/storageAccKey.txt) $(cat /root/config/storageSecKey.txt)

name=$(cat /etc/hostname)
ip4=$(dig -4 +short myip.opendns.com @resolver1.opendns.com)
echo "Digging ipv4 address"
dig -4 myip.opendns.com @resolver1.opendns.com

ip6=$(dig -6 +short @resolver1.opendns.com myip.opendns.com ANY)
echo "Digging ipv6 address"
dig -6 +short @resolver1.opendns.com myip.opendns.com ANY

otime=$(date --iso-8601=seconds)
echo "We are operational, time:"
date --iso-8601=seconds

os=$(hostnamectl | grep Operating | cut -d ':' --fields 2 | tr -d ' ')

desc="descriptor-$(date +"%y%m%d%H%M").txt"
echo "{\"hostname\": \"$name\", \"provider\": \"$6\", \"ipv4\": \"$ip4\", \"ipv6\": \"$ip6\", \"creation\": \"$otime\", \"os\": \"$os\", \"region\": \"$7\"}" > /root/config/$desc
echo $otime > /root/config/otime.txt
echo $ip4 > /root/config/ip4.txt
echo $ip6 > /root/config/ip6.txt

ip=$( echo $ip4 | sed -r 's/\./-/g' )

mc cp /root/config/$desc tupload/$(cat /root/config/bucket.txt)/descriptors/$ip/$desc

# Fix nameservers to do ipv6
systemctl stop systemd-resolved
systemctl disable systemd-resolved
rm /etc/resolv.conf
echo "nameserver 2001:4860:4860::8888" > /etc/resolv.conf

if [ "$7" = "us-central1" ] && [ "$6" = "gcp" ]; then
  # ACCEPT rules to allow GCP healthchecks
  # These rules take precedence over the complete DROP rules as they're added earlier
  sudo iptables -A INPUT -i "$iface" -s 35.191.0.0/16 -p tcp --dport 28763 -m conntrack --ctstate NEW,ESTABLISHED,RELATED -j ACCEPT
  sudo iptables -A INPUT -i "$iface" -s 209.85.152.0/22 -p tcp --dport 28763 -m conntrack --ctstate NEW,ESTABLISHED,RELATED -j ACCEPT
  sudo iptables -A INPUT -i "$iface" -s 209.85.204.0/22 -p tcp --dport 28763 -m conntrack --ctstate NEW,ESTABLISHED,RELATED -j ACCEPT

  sudo iptables -A OUTPUT -o "$iface" -d 35.191.0.0/16 -p tcp --sport 28763 -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT
  sudo iptables -A OUTPUT -o "$iface" -d 209.85.152.0/22 -p tcp --sport 28763 -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT
  sudo iptables -A OUTPUT -o "$iface" -d 209.85.204.0/22 -p tcp --sport 28763 -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT
fi

# Drop inbound and outbound ipv4 traffic, we want to be completely silent.
sudo iptables -A INPUT -i "$iface" -j DROP
sudo iptables -A OUTPUT -o "$iface" -j DROP

systemctl daemon-reload
systemctl restart ssh.socket
systemctl start gotrace
