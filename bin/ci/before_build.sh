#!/bin/sh

${RABBITHOLE_RABBITMQCTL:="sudo rabbitmqctl"}

# guest:guest has full access to /

$RABBITHOLE_RABBITMQCTL add_vhost /
$RABBITHOLE_RABBITMQCTL add_user guest guest
$RABBITHOLE_RABBITMQCTL set_permissions -p / guest ".*" ".*" ".*"


$RABBITHOLE_RABBITMQCTL add_vhost "rabbit/hole"
$RABBITHOLE_RABBITMQCTL set_permissions -p "rabbit/hole" guest ".*" ".*" ".*"
