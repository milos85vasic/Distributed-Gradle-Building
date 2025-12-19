# GCP VPC Network
resource "google_compute_network" "this" {
  name                    = var.name
  auto_create_subnetworks = false
  routing_mode            = "GLOBAL"

  labels = var.labels
}

# GCP Subnets
resource "google_compute_subnetwork" "this" {
  for_each = var.subnet_cidrs

  name          = "${var.name}-${each.key}"
  ip_cidr_range = each.value
  region        = var.region
  network       = google_compute_network.this.id

  private_ip_google_access = true

  # Enable flow logs for monitoring
  log_config {
    aggregation_interval = "INTERVAL_10_MIN"
    flow_sampling        = 0.5
    metadata             = "INCLUDE_ALL_METADATA"
  }
}

# Cloud Router for NAT Gateway (if private endpoints are enabled)
resource "google_compute_router" "nat_router" {
  count = var.enable_private_endpoint ? 1 : 0

  name    = "${var.name}-nat-router"
  region  = var.region
  network = google_compute_network.this.id
}

# NAT Gateway for private subnets
resource "google_compute_router_nat" "nat_gateway" {
  count = var.enable_private_endpoint ? 1 : 0

  name                               = "${var.name}-nat-gateway"
  router                             = google_compute_router.nat_router[0].name
  region                             = var.region
  nat_ip_allocate_option             = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"

  log_config {
    enable = true
    filter = "ERRORS_ONLY"
  }
}

# Firewall rules for internal traffic
resource "google_compute_firewall" "allow_internal" {
  name    = "${var.name}-allow-internal"
  network = google_compute_network.this.name

  allow {
    protocol = "tcp"
    ports    = ["0-65535"]
  }

  allow {
    protocol = "udp"
    ports    = ["0-65535"]
  }

  allow {
    protocol = "icmp"
  }

  source_ranges = [var.network_cidr]
  target_tags   = ["internal"]
}

# Firewall rules for health checks
resource "google_compute_firewall" "allow_health_checks" {
  name    = "${var.name}-allow-health-checks"
  network = google_compute_network.this.name

  allow {
    protocol = "tcp"
    ports    = ["80", "443"]
  }

  source_ranges = ["35.191.0.0/16", "130.211.0.0/22"]
  target_tags   = ["health-check"]
}

# Firewall rules for SSH (restricted)
resource "google_compute_firewall" "allow_ssh" {
  name    = "${var.name}-allow-ssh"
  network = google_compute_network.this.name

  allow {
    protocol = "tcp"
    ports    = ["22"]
  }

  source_ranges = ["35.235.240.0/20"] # IAP IP range
  target_tags   = ["ssh"]
}