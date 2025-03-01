variable "folder_id" {
    type = string
}

variable "zone_id" {
    type    = string
    default = "ru-central1-a"
}

variable "ssh_key" {
    type      = string
    sensitive = true
}

variable "vm_user" {
    type = string
}