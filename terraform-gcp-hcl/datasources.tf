data "google_compute_image" "minecraft-ubuntu" {
  family  = "ubuntu-2004-lts"
  project = "ubuntu-os-cloud"
}

data "google_project" "project" {
  project_id = "331154168162"
}

data "google_service_account" "minecraft-sa" {
  project    = data.google_project.project.project_id
  account_id = "minctl@minectl-fn.iam.gserviceaccount.com"
}

data "google_compute_network" "default" {
  name = "default"
  project = data.google_project.project.project_id
}
