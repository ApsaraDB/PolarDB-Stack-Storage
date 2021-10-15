#!/usr/bin/env bash

# verify script param
ACTION=$1
if [[ $ACTION != "install" ]] && [[ $ACTION != "upgrade" ]] && [[ $ACTION != "reload" ]]; then
  echo $"Usage: $0 {install|upgrade|reload}"
  exit 1
fi

# init script dir
POLAR_DIR=$(
  cd "$(dirname "$0")" || exit
  pwd
)
if [[ $POLAR_DIR != *$SCRIPTS* ]]; then
  SCRIPTS_DIR=$POLAR_DIR/$SCRIPTS_DIR
else
  SCRIPTS_DIR=$POLAR_DIR
fi

SCRIPTS=scripts
AGENT_INI="$SCRIPTS_DIR/agent.ini"

# init agent rpm name
AGENT_RPM=$2
if [[ $AGENT_RPM == "" ]]; then
  AGENT_RPM=$(grep agent_rpm "$AGENT_INI"|awk -F '=' '{print $2}')
fi

# interface network card
NIC=$(grep nic "$AGENT_INI"| awk -F '=' '{print $2}')

function batch() {
  if ! ansible -i "$AGENT_INI" all -m file -a "path={{agent_dir}} state=directory mode=0777"; then
    echo "Failed init polardb-sms agent work dir!"
    exit 1
  fi
  echo "Successfully init polardb-sms agent work dir!"

  if ! ansible -i "$AGENT_INI" all -m synchronize -a "src={{rpm_dir}}/$AGENT_RPM dest={{agent_dir}}"; then
    echo "Failed copy polardb-sms agent rpm package!"
    exit 1
  fi
  echo "Successfully copy polardb-sms agent rpm package!"

  if ! ansible -i "$AGENT_INI" all -m copy -a "src=$SCRIPTS_DIR/{{agent_script}} dest={{agent_dir}} mode=755"; then
    echo "Failed copy polardb-sms agent install script!"
    exit 1
  fi
  echo "Successfully copy polardb-sms agent install script!"

  if ! ansible -i "$AGENT_INI" all -m shell -a "{{agent_dir}}/{{agent_script}} $ACTION {{report_endpoint}} $NIC $AGENT_RPM {{agent_dir}}>install.log 2>&1"; then
    echo "Failed exec polardb-sms agent install script!"
    exit 1
  fi
  echo "Successfully exec polardb-sms agent install script!"
}

function reload() {
  if ! ansible -i "$AGENT_INI" all -m command -a "supervisorctl reload"; then
    echo "Failed reload supervisor config!"
    exit 1
  fi
  echo "Successfully reload supervisor config!"
}

# ansible batch handle agent server on cluster
case "$ACTION" in
  install)
    batch
    ;;
  upgrade)
    batch
    ;;
  reload)
    reload
    ;;
esac