
export TGBOT_TG_TOKEN=$(yc lockbox payload get tg-token --key token)
export TGBOT_DB_PASSWORD=$(yc lockbox payload get db-password --key password)
export REGISTRY_ID=$(yc container registry get default | yq -r .id)
docker buildx build --no-cache . --secret id=TGBOT_TG_TOKEN --secret id=TGBOT_DB_PASSWORD -t cr.yandex/$REGISTRY_ID/tgbot:latest
docker push cr.yandex/$REGISTRY_ID/tgbot:latest