#!/bin/bash

# Script is to be run from TF
# builds image, pushes it to YC registry $REGISTRY_ID
# image name tgbot, tag latest
# prints out image ID 

pushd ../..
docker buildx build --no-cache . --secret id=TGBOT_TG_TOKEN --secret id=TGBOT_DB_PASSWORD -t cr.yandex/$REGISTRY_ID/tgbot:latest >&2
docker push cr.yandex/$REGISTRY_ID/tgbot:latest >&2
echo "cr.yandex/$REGISTRY_ID/tgbot:latest"
popd