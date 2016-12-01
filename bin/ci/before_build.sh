#!/bin/sh

${RABBITHOLE_RABBITMQCTL:="sudo rabbitmqctl"}

# guest:guest has full access to /

$RABBITHOLE_RABBITMQCTL add_vhost /
$RABBITHOLE_RABBITMQCTL add_user guest guest
$RABBITHOLE_RABBITMQCTL set_permissions -p / guest ".*" ".*" ".*"

# Reduce retention policy for faster publishing of stats
$RABBITHOLE_RABBITMQCTL eval 'supervisor2:terminate_child(rabbit_mgmt_sup_sup, rabbit_mgmt_sup), application:set_env(rabbitmq_management,       sample_retention_policies, [{global, [{605, 1}]}, {basic, [{605, 1}]}, {detailed, [{10, 1}]}]), rabbit_mgmt_sup_sup:start_child().'
$RABBITHOLE_RABBITMQCTL eval 'supervisor2:terminate_child(rabbit_mgmt_sup_sup, rabbit_mgmt_sup), application:set_env(rabbitmq_management_agent, sample_retention_policies, [{global, [{605, 1}]}, {basic, [{605, 1}]}, {detailed, [{10, 1}]}]), rabbit_mgmt_sup_sup:start_child().'

$RABBITHOLE_RABBITMQCTL add_vhost "rabbit/hole"
$RABBITHOLE_RABBITMQCTL set_permissions -p "rabbit/hole" guest ".*" ".*" ".*"
