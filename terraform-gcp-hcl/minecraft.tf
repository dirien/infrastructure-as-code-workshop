data "google_project" "project" {
}

resource "google_compute_address" "minecraft-static-ip" {
  name = "ipv4-address"
}


resource "google_compute_firewall" "minecraft-fw" {
  name    = "minecraft-fw"
  network = data.google_compute_network.default.name

  allow {
    protocol = "tcp"
    ports    = ["22", "25565"]
  }
  source_ranges = [
    "0.0.0.0/0"
  ]
  direction     = "INGRESS"
}

resource "google_compute_instance" "minecraft-server" {
  name         = var.name
  machine_type = var.machine_type


  boot_disk {
    initialize_params {
      image = data.google_compute_image.minecraft-ubuntu.self_link
      size  = 10
    }
  }

  metadata_startup_script = file("../config/startup.sh")
  metadata                = {
    ssh-keys = "minectl:${file(var.ssh_pub_key)}"
  }
  network_interface {
    network = data.google_compute_network.default.name
    access_config {
      nat_ip = google_compute_address.minecraft-static-ip.address
    }
  }
  service_account {
    email  = data.google_service_account.minecraft-sa.email
    scopes = [
      "storage-full",
      "compute-rw"
    ]
  }
}