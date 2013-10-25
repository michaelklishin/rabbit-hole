#!/bin/sh

${RABBITMQCTL:="sudo rabbitmqctl"}
${RABBITMQ_PLUGINS:="sudo rabbitmq-plugins"}

# guest:guest has full access to /

$RABBITMQCTL add_vhost /
$RABBITMQCTL add_user guest guest
$RABBITMQCTL set_permissions -p / guest ".*" ".*" ".*"


$RABBITMQCTL add_vhost "rabbit/hole"
$RABBITMQCTL set_permissions -p "rabbit/hole" guest ".*" ".*" ".*"
