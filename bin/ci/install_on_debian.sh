#!/bin/sh

cd /tmp/
wget https://github.com/rabbitmq/rabbitmq-server/releases/download/rabbitmq_v3_6_9/rabbitmq-server_3.6.9-1_all.deb
sudo dpkg --install rabbitmq-server_3.6.9-1_all.deb
sudo rm rabbitmq-server_3.6.9-1_all.deb

sleep 3
