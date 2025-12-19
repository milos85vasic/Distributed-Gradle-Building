output "network_name" {
  description = "The name of the VPC network"
  value       = google_compute_network.this.name
}

output "subnet_name" {
  description = "The name of the primary subnet"
  value       = google_compute_subnetwork.this["private"].name
}

output "network_self_link" {
  description = "The self-link of the VPC network"
  value       = google_compute_network.this.self_link
}

output "subnet_self_link" {
  description = "The self-link of the primary subnet"
  value       = google_compute_subnetwork.this["private"].self_link
}