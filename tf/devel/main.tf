# Run it once per deployment init

# Providers

terraform {
  required_providers {
    yandex = {
      source = "yandex-cloud/yandex"
    }
  }
  required_version = ">= 0.13"
}

provider "yandex" {}

# Variables

variable "folder_name" {
    type = string
}

variable "ssh_key" {
    type    = string
    sensitive = true
}

variable "vm_user" {
    type = string
}

variable "tg_token" {
    type      = string
    sensitive = true
}

variable "zone_id" {
    type    = string
    default = "ru-central1-a"
}

# Resources

# folder
resource "yandex_resourcemanager_folder" "deploy-folder" {
    name = var.folder_name
}

# deploy account
resource "yandex_iam_service_account" "tf-deployer" {
    name        = "tf-deployer"
    folder_id   = yandex_resourcemanager_folder.deploy-folder.id
    description = "Service account to deploy via Terraform"
}
resource "yandex_resourcemanager_folder_iam_member" "admin" {
    folder_id = yandex_resourcemanager_folder.deploy-folder.id
    role      = "admin"
    member    = "serviceAccount:${yandex_iam_service_account.tf-deployer.id}"
}

# lockbox - tf state
resource "yandex_lockbox_secret" "tfstate-key" {
    name        = "tfstate-key"
    folder_id   = yandex_resourcemanager_folder.deploy-folder.id
    description = "TF state S3 access key"
}

# lockbox - tgbot token
resource "yandex_lockbox_secret" "tgbot-token" {
    name        = "tgbot-token"
    folder_id   = yandex_resourcemanager_folder.deploy-folder.id
    description = "TG bot token"
}
resource "yandex_lockbox_secret_version" "tgbot-token" {
    secret_id = yandex_lockbox_secret.tgbot-token.id
    entries {
        key        = "token"
        text_value = var.tg_token
    }
}

# lockbox - ssh key
resource "yandex_lockbox_secret" "ssh-key" {
    name        = "deploy-ssh-key"
    folder_id   = yandex_resourcemanager_folder.deploy-folder.id
    description = "SSH key to deployment VM"
}
resource "yandex_lockbox_secret_version" "ssh-key" {
    secret_id = yandex_lockbox_secret.ssh-key.id
    entries {
        key        = "user"
        text_value = var.vm_user
    }
    entries {
        key        = "ssh-key"
        text_value = var.ssh_key
    }
}

# S3 access key
resource "yandex_iam_service_account_static_access_key" "tfstate-key" {
    service_account_id = yandex_iam_service_account.tf-deployer.id
    output_to_lockbox {
        secret_id            = yandex_lockbox_secret.tfstate-key.id
        entry_for_access_key = "access"
        entry_for_secret_key = "secret"
    }
    description = "TF state S3 bucket access"
}

# bucket
resource "yandex_storage_bucket" "tf-state" {
    folder_id = yandex_resourcemanager_folder.deploy-folder.id
}

# devel VM
data "yandex_compute_image" "ubuntu" {
    family = "ubuntu-2204-lts"
}

resource "yandex_vpc_network" "devel-network" {
    name      = "devel-network"
    folder_id = yandex_resourcemanager_folder.deploy-folder.id
}

resource "yandex_vpc_subnet" "subnet" {
    name           = "subnet-${var.zone_id}"
    zone           = var.zone_id
    folder_id      = yandex_resourcemanager_folder.deploy-folder.id
    v4_cidr_blocks = ["192.168.10.0/24"]
    network_id     = "${yandex_vpc_network.devel-network.id}"
}

resource "yandex_compute_instance" "devel" {
    name                      = "devel"
    description               = "VM for deployment and development"
    allow_stopping_for_update = true
    platform_id               = "standard-v3"
    zone                      = var.zone_id
    folder_id                 = yandex_resourcemanager_folder.deploy-folder.id
    resources {
        cores  = 2
        memory = 2
    }
    boot_disk {
        initialize_params {
            image_id = data.yandex_compute_image.ubuntu.id
            size     = 20
        }
    }
    network_interface {
        subnet_id = "${yandex_vpc_subnet.subnet.id}"
        nat       = true
    }
    metadata = {
        user-data = templatefile("${path.module}/metadata.yaml", {
            USER_NAME    = var.vm_user
            USER_SSH_KEY = var.ssh_key
        })
    }
    service_account_id = yandex_iam_service_account.tf-deployer.id
}

# Output

output "folder_id" {
    value = yandex_resourcemanager_folder.deploy-folder.id
}
output "vm_id" {
    value = yandex_compute_instance.devel.id
}
output "bucket_id" {
    value = yandex_storage_bucket.tf-state.id
}
output "tg_token_id" {
    value = yandex_lockbox_secret.tgbot-token.id
}