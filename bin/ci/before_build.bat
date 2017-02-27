@echo off
setlocal enabledelayedexpansion

if "!RABBITHOLE_RABBITMQCTL!"=="" (
    set RABBITHOLE_RABBITMQCTL=rabbitmqctl
)

if "!RABBITHOLE_RABBITMQ_PLUGINS!"=="" (
    set RABBITHOLE_RABBITMQ_PLUGINS=rabbitmq-plugins
)

REM guest:guest has full access to /

call %RABBITHOLE_RABBITMQCTL% add_vhost /
call %RABBITHOLE_RABBITMQCTL% add_user guest guest
call %RABBITHOLE_RABBITMQCTL% set_permissions -p / guest ".*" ".*" ".*"

REM Reduce retention policy for faster publishing of stats
call %RABBITHOLE_RABBITMQCTL% eval "supervisor2:terminate_child(rabbit_mgmt_sup_sup, rabbit_mgmt_sup), application:set_env(rabbitmq_management,       sample_retention_policies, [{global, [{605, 1}]}, {basic, [{605, 1}]}, {detailed, [{10, 1}]}]), rabbit_mgmt_sup_sup:start_child()."
call %RABBITHOLE_RABBITMQCTL% eval "supervisor2:terminate_child(rabbit_mgmt_agent_sup_sup, rabbit_mgmt_agent_sup), application:set_env(rabbitmq_management_agent, sample_retention_policies, [{global, [{605, 1}]}, {basic, [{605, 1}]}, {detailed, [{10, 1}]}]), rabbit_mgmt_agent_sup_sup:start_child()."

call %RABBITHOLE_RABBITMQCTL% add_vhost "rabbit/hole"
call %RABBITHOLE_RABBITMQCTL% set_permissions -p "rabbit/hole" guest ".*" ".*" ".*"

REM Enable shovel plugin
call %RABBITHOLE_RABBITMQ_PLUGINS% enable rabbitmq_shovel
call %RABBITHOLE_RABBITMQ_PLUGINS% enable rabbitmq_shovel_management
