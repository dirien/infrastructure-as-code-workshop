terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "5.17.0"
    }
  }
}

provider "google" {
  region  = var.region
  zone    = var.zone
}
