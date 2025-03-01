# initial

data "yandex_resourcemanager_folder" "folder" {
    id = var.folder_id
}

data "yandex_lockbox_secret" "tgbot-token" {
    name      = "tgbot-token"
    folder_id = var.folder_id
}

# net & subnet

resource "yandex_vpc_network" "network" {
    name      = "tgbot"
    folder_id = var.folder_id
}

resource "yandex_vpc_subnet" "subnet" {
  name           = "subnet-${var.zone_id}"
  zone           = var.zone_id
  v4_cidr_blocks = ["192.168.1.0/24"]
  network_id     = yandex_vpc_network.network.id
}

# service accounts

resource "yandex_iam_service_account" "tgbot_sa" {
    folder_id = var.folder_id
    name      = "tgbot_sa"
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

# databases?
# registry

data "yandex_container_registry" "registry" {
  name      = var.registry_id
  folder_id = var.folder_id
}

data "external" "docker-build" {
	program = ["bash", "${path.module}/build.sh"]
	working_dir = "${path.module}"
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

resource "yandex_compute_instance_group" "ig" {
  name                = "tgbot"
  folder_id           = var.folder_id
  service_account_id  = yandex_iam_service_account.tgbot_sa.id
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

    network_interface {
      network_id = yandex_vpc_network.network.id
      subnet_ids = [ yandex_vpc_subnet.subnet.id ]
      nat        = true
    }

    metadata = {
      user-data = templatefile("config.yaml", {
        VM_USER  = var.vm_user
        SSH_KEY  = var.ssh_key
        LOG_GROUP_ID = yandex_logging_group.logging.id
      })
      docker-container-declaration = templatefile("containers.yaml", {
        DOCKER_IMAGE = data.external.docker-build.result
        LOG_GROUP_ID = yandex_logging_group.logging.id
      })
    }
  }

  scale_policy {
    auto_scale {
      initial_size           = 1
      measurement_duration   = 60
      cpu_utilization_target = 40
      min_zone_size          = 1
      max_size               = 6
      warmup_duration        = 120
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
