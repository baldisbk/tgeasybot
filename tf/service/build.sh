#!/bin/bash

export DOCKER_BUILDKIT=1
set -e

# Script is to be run from TF
# builds image, pushes it to YC registry $REGISTRY_ID
# image name tgbot, tag latest
# prints out image ID 
THE_INPUT=$(cat)
export TGBOT_TG_TOKEN_ID=$(echo $THE_INPUT | jq -r .tg_token_id)
export TGBOT_DB_PASSWORD_ID=$(echo $THE_INPUT | jq -r .db_password_id)
export REGISTRY_ID=$(echo $THE_INPUT | jq -r .registry_id)
export TGBOT_TG_TOKEN=$(yc lockbox payload get --id $TGBOT_TG_TOKEN_ID --key token)
export TGBOT_DB_PASSWORD=$(yc lockbox payload get --id $TGBOT_DB_PASSWORD_ID --key password)
pushd ../.. >&2
yc iam create-token | docker login --username iam --password-stdin cr.yandex >&2
docker buildx build --no-cache . --secret id=TGBOT_TG_TOKEN --secret id=TGBOT_DB_PASSWORD -t cr.yandex/$REGISTRY_ID/tgbot:latest >&2
docker push cr.yandex/$REGISTRY_ID/tgbot:latest >&2
popd >&2
echo '{"image":"cr.yandex/'$REGISTRY_ID'/tgbot:latest"}'