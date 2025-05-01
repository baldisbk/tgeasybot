#!/bin/bash
set -e

export FOLDER_ID=${FOLDER_ID:-$(curl -H metadata-flavor:Google 169.254.169.254/computeMetadata/v1/yandex/folder-id)}
export YC_TOKEN=$(yc iam create-token)
export ACCESS_KEY=$(yc lockbox payload get --name tfstate-key --key access)
export SECRET_KEY=$(yc lockbox payload get --name tfstate-key --key secret)
export S3_BUCKET=$(yc storage bucket list --format=json | jq -r ".[].name")
export SSH_PUBLIC_KEY=${SSH_PUBLIC_KEY:-$(cat $HOME/.ssh/id_ed25519.pub)}
export VM_USER=${VM_USER:-$USER}

cd tf/service
terraform init -backend-config="access_key=$ACCESS_KEY" -backend-config="secret_key=$SECRET_KEY" -backend-config="bucket=$S3_BUCKET"
terraform apply -var="folder_id=$FOLDER_ID" -var="ssh_key=$SSH_PUBLIC_KEY" -var="vm_user=$VM_USER"
