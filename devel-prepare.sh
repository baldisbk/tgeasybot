#!/bin/bash
set -e

sudo apt install snapd
sudo snap install terraform --classic
sudo snap install jq --classic
sudo snap install go --classic
sudo snap install docker --classic
sudo groupadd docker
sudo usermod -aG docker $USER
cp terraform.rc ${HOME}/.terraformrc
curl -sSL https://storage.yandexcloud.net/yandexcloud-yc/install.sh | bash
ssh-keygen -t ed25519 -f ~/.ssh/id_ed25519 -q -N ""

sudo reboot now