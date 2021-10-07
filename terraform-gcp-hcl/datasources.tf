data "google_compute_image" "minecraft-ubuntu" {
  family  = "ubuntu-2004-lts"
  project = "ubuntu-os-cloud"
}

data "google_service_account" "minecraft-sa" {
  project    = data.google_project.project.id
  account_id = "110246453965025503223"
}

data "google_compute_network" "default" {
  name = "default"
}