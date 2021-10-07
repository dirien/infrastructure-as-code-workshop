variable "ssh_pub_key" {
  default = "../ssh/workshop.pub"
}

variable "name" {
  default = "minecraft-server"
}

variable "machine_type" {
  default = "e2-standard-2"
}
variable "region" {
  default = "europe-west6"
}

variable "zone" {
  default = "europe-west6-a"
}