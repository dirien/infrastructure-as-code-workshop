#!/bin/bash
iptables -I INPUT -j ACCEPT
tee /tmp/server.properties <<EOF
server-port=25565
level-seed=stackitminecraftrocks
view-distance=10
enable-jmx-monitoring=false
server-ip=
resource-pack-prompt=
gamemode=survival
allow-nether=true
enable-command-block=false
sync-chunk-writes=true
enable-query=false
op-permission-level=4
prevent-proxy-connections=false
resource-pack=
entity-broadcast-range-percentage=100
level-name=world
player-idle-timeout=0
motd=\u00A79GCE \u00A7rMinecraft --- \u00A76Java \u00A7redition
query.port=25565
force-gamemode=false
rate-limit=0
hardcore=false
white-list=false
broadcast-console-to-ops=true
pvp=true
spawn-npcs=true
spawn-animals=true
snooper-enabled=true
difficulty=easy
function-permission-level=2
network-compression-threshold=256
text-filtering-config=
require-resource-pack=false
spawn-monsters=true
max-tick-time=60000
enforce-whitelist=false
use-native-transport=true
max-players=100
resource-pack-sha1=
spawn-protection=16
online-mode=true
enable-status=true
allow-flight=false
max-world-size=29999984

broadcast-rcon-to-ops=true
rcon.port=25575
enable-rcon=true
rcon.password=test
EOF
tee /etc/systemd/system/minecraft.service <<EOF
[Unit]
Description=Minecraft Server
Documentation=https://www.minecraft.net/en-us/download/server
DefaultDependencies=no
After=network.target

[Service]
WorkingDirectory=/minecraft
Type=simple
ExecStart=/usr/bin/java -Xmx2G -Xms2G -jar server.jar nogui

Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
apt update
apt-get install -y apt-transport-https ca-certificates curl openjdk-16-jre-headless fail2ban
ufw allow ssh
ufw allow 5201

ufw allow proto tcp to 0.0.0.0/0 port 25565


echo [DEFAULT] | sudo tee -a /etc/fail2ban/jail.local
echo banaction = ufw | sudo tee -a /etc/fail2ban/jail.local
echo [sshd] | sudo tee -a /etc/fail2ban/jail.local
echo enabled = true | sudo tee -a /etc/fail2ban/jail.local
sudo systemctl restart fail2ban
mkdir -p /minecraft
URL=$(curl -s https://java-version.minectl.ediri.online/binary/1.17.1)
curl -sLSf $URL > /minecraft/server.jar
echo "eula=true" > /minecraft/eula.txt
mv /tmp/server.properties /minecraft/server.properties
chmod a+rwx /minecraft
systemctl restart minecraft.service
systemctl enable minecraft.service