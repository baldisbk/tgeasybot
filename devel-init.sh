#!/bin/bash
set -e

export YC_TOKEN=$(yc iam create-token)
export TGBOT_TG_TOKEN=${TGBOT_TG_TOKEN:-$(cat tgbot.token)}
export SSH_PUBLIC_KEY=${SSH_PUBLIC_KEY:-$(cat $HOME/.ssh/id_ed25519.pub)}
export VM_USER=${VM_USER:-$USER}

cd tf/devel
terraform apply -var="folder_name=$1" -var="ssh_key=$SSH_PUBLIC_KEY" -var="vm_user=$VM_USER" -var="tg_token=$TGBOT_TG_TOKEN"