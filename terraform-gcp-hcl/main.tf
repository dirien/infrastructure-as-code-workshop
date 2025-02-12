terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "6.20.0"
    }
  }
}

provider "google" {
  region  = var.region
  zone    = var.zone
}
