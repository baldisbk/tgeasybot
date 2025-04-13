cp terraform.rc ${HOME}/.terraformrc

sudo snap install terraform --classic
sudo snap install jq --classic
sudo snap install docker --classic
sudo groupadd docker
sudo usermod -aG docker $USER
curl -sSL https://storage.yandexcloud.net/yandexcloud-yc/install.sh | bash
