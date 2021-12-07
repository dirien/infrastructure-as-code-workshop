terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "4.3.0"
    }
  }
}

provider "google" {
  project = "minectl-fn"
  region  = var.region
  zone    = var.zone
}