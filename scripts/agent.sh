#!/usr/bin/env bash

ACTION=$1
REPORT_ENDPOINT=$2
NIC=$3
AGENT_RPM=$4
AGENT_DIR=$5

function check() {
  local ret=0
  for endpoint in ${1/,/ }; do
    local ip=(${endpoint/:/ })

    if [[ $ip =~ ^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$ ]]; then
        ips=(${ip//\./ }) # 按.分割，转成数组，方便下面的判断
        if [[ ${ips[0]} -le 255 && ${ips[1]} -le 255 && ${ips[2]} -le 255 && ${ips[3]} -le 255 ]]; then
            ret=0
        else
            echo "Endpoint: $endpoint not available!"
            ret=1
            break
        fi
    else
        ret=2
        echo "Endpoint: $endpoint format error!"
        break
    fi
  done
  return $ret
}

function install_supervisor() {
  rpm=$(rpm -qa | grep supervisor)
  if [[ ${#rpm} -eq 0 ]]; then
      yum install -y supervisor
      echo "Successfully installed supervisor!"
  else
      echo "Supervisor has installed!"
  fi

  service=$(systemctl list-unit-files --type=service | grep enabled | grep supervisord.service)
  if [[ ${#service} -eq 0 ]]; then
      systemctl enable supervisord.service
      echo "Successfully enable supervisord service!"
  else
      echo "Supervisord service has enabled!"
  fi
}

function install_agent() {
  if ! rpm -ivh --force "$AGENT_DIR/$AGENT_RPM"; then
    echo "Failed installed polardb-sms agent!"
    exit 1
  fi
  echo "Successfully installed polardb-sms agent!"
}

function upgrade_agent() {
  if ! rpm -Uvh --force "$AGENT_DIR/$AGENT_RPM"; then
    echo "Failed upgraded polardb-sms agent!"
    exit 1
  fi
  echo "Successfully upgraded polardb-sms agent!"
}

function config_agent() {
  if ! ifconfig "$NIC"; then
    echo "Failed found network interface card $NIC"
    exit 1
  fi
    NODE_IP=$(ifconfig "$NIC"|grep netmask|awk '{print $2}')
  echo "Successfully get network interface card $NODE_IP!"

  cat << EOF >/etc/supervisord.d/polardb-sms-agent.ini
[program:polardb-sms-agent]
command=/home/a/project/t-polardb-sms-agent/bin/polardb-sms-agent --port=18888 --report-endpoint=$REPORT_ENDPOINT --node-ip=$NODE_IP --node-id=%(host_node_name)s
process_name=%(program_name)s
startretries=1000
autorestart=unexpected
autostart=true
EOF
  echo "Successfully config polardb-sms agent!"
}

function uninstall_agent() {
  for i in $(rpm -qa | grep polardb-sms-agent) ; do
      if ! rpm -e "$i"; then
        echo "Failed uninstall polardb-sms agent $i!"
        exit 1
      fi
  done
  echo "Successfully uninstall polardb-sms agent rpms!"
}

function start() {
  if ! supervisorctl restart polardb-sms-agent; then
      echo "Failed started supervisord!"
      exit 1
  fi
  echo "Successfully started supervisord!"

  # wait agent process start
  sleep 1s

  # process max length char 15
  if ! pgrep -l polardb-sms-age; then
      echo "Failed started polardb-sms agent!"
      exit 1
  fi
  echo "Successfully started polardb-sms agent!"

  rm -f "$AGENT_DIR/$AGENT_RPM"
}

# check startup options number
if [[ $# -lt 5 ]]; then
    echo "Please enter the enough parameters: [$*]!"
    exit 1
elif [[ $# -gt 5 ]]; then
    echo "Please enter too many parameters: [$*]!"
    exit 1
fi

# check manager servers endpoints
if ! check "$MANAGER_ENDPOINT"; then
  echo "Failed check manager servers endpoints $MANAGER_ENDPOINT!"
  exit 1
fi
echo "Successfully check manager servers endpoint $MANAGER_ENDPOINT!"

# check report endpoint
if ! check "$REPORT_ENDPOINT"; then
  echo "Failed check report endpoint $REPORT_ENDPOINT!"
  exit 1
fi
echo "Successfully check report endpoint $REPORT_ENDPOINT!"

# check polardb-sms agent rpm
if ! ls "$AGENT_DIR/$AGENT_RPM"; then
    exit 1
fi

# install supervisor
install_supervisor

# uninstall agent
uninstall_agent

# install or upgrade agent
case "$ACTION" in
  install)
    install_agent
    ;;
  upgrade)
    upgrade_agent
    ;;
  *)
    echo $"Usage: $0 {install|upgrade}"
    exit 1
esac

# config agent
config_agent

# start server
start
