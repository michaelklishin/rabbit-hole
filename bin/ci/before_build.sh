#!/bin/sh

CTL=${RABBITHOLE_RABBITMQCTL:-"sudo rabbitmqctl"}
PLUGINS=${RABBITHOLE_RABBITMQ_PLUGINS:-"sudo rabbitmq-plugins"}

echo "Will use rabbitmqctl at ${CTL}"
echo "Will use rabbitmq-plugins at ${PLUGINS}"

$PLUGINS enable rabbitmq_management

sleep 3

# guest:guest has full access to /

$CTL add_vhost /
$CTL add_user guest guest
$CTL set_permissions -p / guest ".*" ".*" ".*"

# Reduce retention policy for faster publishing of stats
$CTL eval 'supervisor2:terminate_child(rabbit_mgmt_sup_sup, rabbit_mgmt_sup), application:set_env(rabbitmq_management,       sample_retention_policies, [{global, [{605, 1}]}, {basic, [{605, 1}]}, {detailed, [{10, 1}]}]), rabbit_mgmt_sup_sup:start_child().'
$CTL eval 'supervisor2:terminate_child(rabbit_mgmt_agent_sup_sup, rabbit_mgmt_agent_sup), application:set_env(rabbitmq_management_agent, sample_retention_policies, [{global, [{605, 1}]}, {basic, [{605, 1}]}, {detailed, [{10, 1}]}]), rabbit_mgmt_agent_sup_sup:start_child().'

$CTL add_vhost "rabbit/hole"
$CTL set_permissions -p "rabbit/hole" guest ".*" ".*" ".*"

# set cluster name
$CTL set_cluster_name rabbitmq@localhost

# Enable shovel plugin
$PLUGINS enable rabbitmq_shovel
$PLUGINS enable rabbitmq_shovel_management
