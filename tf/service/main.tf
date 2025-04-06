# initial

data "yandex_resourcemanager_folder" "folder" {
  folder_id = var.folder_id
}

data "yandex_lockbox_secret" "tgbot_token_secret" {
  name      = "tgbot-token"
  folder_id = var.folder_id
}

data "yandex_lockbox_secret_version" "tgbot_token" {
  secret_id = data.yandex_lockbox_secret.tgbot_token_secret.id
}

# net & subnet

resource "yandex_vpc_network" "network" {
  name      = "tgbot"
  folder_id = var.folder_id
}

resource "yandex_vpc_subnet" "subnet" {
  name           = "tgbot-subnet"
  zone           = var.zone_id
  folder_id      = var.folder_id
  v4_cidr_blocks = ["192.168.1.0/24"]
  network_id     = yandex_vpc_network.network.id
}

# service accounts

resource "yandex_iam_service_account" "tgbot_sa" {
  folder_id = var.folder_id
  name      = "tgbot-sa"
  description = "SA for tgbot service"
}

resource "yandex_resourcemanager_folder_iam_member" "tgbot_sa_role_docker" {
  folder_id = var.folder_id
  role      = "container-registry.images.puller"
  member    = "serviceAccount:${yandex_iam_service_account.tgbot_sa.id}"
}

resource "yandex_resourcemanager_folder_iam_member" "tgbot_sa_role_logging" {
  folder_id = var.folder_id
  role      = "logging.writer"
  member    = "serviceAccount:${yandex_iam_service_account.tgbot_sa.id}"
}

resource "yandex_iam_service_account" "tgbot_deploy_sa" {
  folder_id   = var.folder_id
  name        = "tgbot-deploy-sa"
  description = "SA for tgbot instance group"
}

resource "yandex_resourcemanager_folder_iam_member" "tgbot_deploy_sa_role_editor" {   
  folder_id = var.folder_id
  role      = "editor"
  member    = "serviceAccount:${yandex_iam_service_account.tgbot_deploy_sa.id}"
}

resource "yandex_resourcemanager_folder_iam_member" "tgbot_deploy_sa_role_vpc_user" {
  folder_id = var.folder_id
  role      = "vpc.user"
  member    = "serviceAccount:${yandex_iam_service_account.tgbot_deploy_sa.id}"
}

# db password

resource "yandex_lockbox_secret" "db_password_secret" {
  name      = "tgbot-db-password"
  folder_id = var.folder_id
}

resource "yandex_lockbox_secret_version" "db_password" {
  secret_id = yandex_lockbox_secret.db_password_secret.id
  entries {
    key = "password"
    command {
      path = "openssl"
      args = ["rand", "-base64", "24"]
    }
  } 
}

data "yandex_lockbox_secret_version" "db_password" {
  secret_id  = yandex_lockbox_secret.db_password_secret.id
  version_id = yandex_lockbox_secret_version.db_password.id
}

# registry

resource "yandex_container_registry" "registry" {
  name      = "tgbot"
  folder_id = var.folder_id
}

data "external" "docker_build" {
  program = ["bash", "${path.module}/build.sh"]
  working_dir = "${path.module}"
  query = {
    tg_token_id    = data.yandex_lockbox_secret.tgbot_token_secret.id
    db_password_id = yandex_lockbox_secret.db_password_secret.id
    registry_id    = yandex_container_registry.registry.id

    tg_token_version_id    = data.yandex_lockbox_secret_version.tgbot_token.id
    db_password_version_id = yandex_lockbox_secret_version.db_password.id
  }
}

# logging

resource "yandex_logging_group" "logging" {
  name      = "logging"
  folder_id = var.folder_id
}

# instance group

data "yandex_compute_image" "image" {
  family = "container-optimized-image"
}

resource "yandex_compute_disk" "db_disk" {
  name      = "tgbot-db-disk"
  zone      = var.zone_id
  folder_id = data.yandex_resourcemanager_folder.folder.id
  size      = 10
}

resource "yandex_compute_instance_group" "ig" {
  name                = "tgbot"
  folder_id           = var.folder_id
  service_account_id  = yandex_iam_service_account.tgbot_deploy_sa.id
  instance_template {
    platform_id = "standard-v3"
    resources {
      memory = 1
      cores  = 2
      core_fraction = 20
    }

    scheduling_policy {
        preemptible = true
    }
  
    boot_disk {
      mode = "READ_WRITE"
      initialize_params {
        image_id = data.yandex_compute_image.image.id
        size     = 20
      }
    }
    secondary_disk {
      disk_id     = yandex_compute_disk.db_disk.id
      device_name = "dbdata"
    }

    network_interface {
      network_id = yandex_vpc_network.network.id
      subnet_ids = [ yandex_vpc_subnet.subnet.id ]
      nat        = true
    }

    metadata = {
      user-data = templatefile("config.yaml", {
        USER_NAME = var.vm_user
        USER_SSH_KEY = var.ssh_key
        LOG_GROUP_ID = yandex_logging_group.logging.id
      })
      docker-compose = templatefile("containers.yaml", {
        DOCKER_IMAGE = data.external.docker_build.result.image
        LOG_GROUP_ID = yandex_logging_group.logging.id
        DB_PASSWORD  = trimspace([ for e in data.yandex_lockbox_secret_version.db_password.entries : e.text_value if e.key == "password" ][0])
      })
    }
    service_account_id  = yandex_iam_service_account.tgbot_sa.id
  }

  scale_policy {
    fixed_scale {
      size = 1
    }
  }

  allocation_policy {
    zones = [ var.zone_id ]
  }

  deploy_policy {
    max_unavailable = 1
    max_expansion   = 0
  }
}
