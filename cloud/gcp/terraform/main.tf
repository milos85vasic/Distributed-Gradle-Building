terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 4.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.0"
    }
  }

  backend "gcs" {
    # Configure your GCS backend here
    # bucket = "your-terraform-state-bucket"
    # prefix = "distributed-gradle/terraform.tfstate"
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

# VPC Module
module "vpc" {
  source = "./modules/vpc"

  name                    = "distributed-gradle"
  project_id              = var.project_id
  region                  = var.region
  network_cidr            = var.network_cidr
  subnet_cidrs            = var.subnet_cidrs
  enable_private_endpoint = var.enable_private_endpoint

  labels = {
    environment = var.environment
    project     = "distributed-gradle"
    managed-by  = "terraform"
  }
}

# GKE Cluster
resource "google_container_cluster" "distributed_gradle" {
  name                     = "${var.environment}-distributed-gradle"
  location                 = var.region
  remove_default_node_pool = true
  initial_node_count       = 1
  network                  = module.vpc.network_name
  subnetwork               = module.vpc.subnet_name
  networking_mode          = "VPC_NATIVE"

  # Enable private cluster
  private_cluster_config {
    enable_private_nodes    = var.enable_private_cluster
    enable_private_endpoint = var.enable_private_endpoint
    master_ipv4_cidr_block  = var.master_ipv4_cidr_block
  }

  # Enable workload identity
  workload_identity_config {
    workload_pool = "${var.project_id}.svc.id.goog"
  }

  # Enable GKE Autopilot or standard mode
  dynamic "cluster_autoscaling" {
    for_each = var.enable_autopilot ? [] : [1]
    content {
      enabled = true
      resource_limits {
        resource_type = "cpu"
        minimum       = var.min_cpu
        maximum       = var.max_cpu
      }
      resource_limits {
        resource_type = "memory"
        minimum       = var.min_memory
        maximum       = var.max_memory
      }
    }
  }

  # Enable GKE Autopilot
  enable_autopilot = var.enable_autopilot

  # Enable binary authorization
  enable_binary_authorization = var.enable_binary_authorization

  # Enable shielded nodes
  enable_shielded_nodes = var.enable_shielded_nodes

  # Enable network policy
  network_policy {
    enabled  = true
    provider = "CALICO"
  }

  # Enable GKE usage metering
  resource_usage_export_config {
    enable_network_egress_metering = true
    enable_resource_consumption_metering = true

    bigquery_destination {
      dataset_id = google_bigquery_dataset.usage_metering.dataset_id
    }
  }

  # Enable GKE add-ons
  addons_config {
    http_load_balancing {
      disabled = false
    }
    horizontal_pod_autoscaling {
      disabled = false
    }
    gcp_filestore_csi_driver_config {
      enabled = true
    }
  }

  # Enable GKE monitoring
  monitoring_config {
    enable_components = ["SYSTEM_COMPONENTS", "WORKLOADS"]
  }

  logging_config {
    enable_components = ["SYSTEM_COMPONENTS", "WORKLOADS"]
  }

  # Maintenance policy
  maintenance_policy {
    recurring_window {
      start_time = "2023-01-01T09:00:00Z"
      end_time   = "2023-01-01T17:00:00Z"
      recurrence = "FREQ=WEEKLY;BYDAY=SA,SU"
    }
  }

  labels = {
    environment = var.environment
    project     = "distributed-gradle"
  }
}

# Node Pools (only for standard mode)
resource "google_container_node_pool" "general" {
  count = var.enable_autopilot ? 0 : 1

  name       = "general-purpose"
  location   = var.region
  cluster    = google_container_cluster.distributed_gradle.name
  node_count = var.default_node_count

  node_config {
    preemptible  = var.use_preemptible_nodes
    machine_type = var.node_machine_type

    # Google recommends custom service accounts with minimal permissions
    service_account = google_service_account.gke.email
    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform"
    ]

    labels = {
      environment = var.environment
      project     = "distributed-gradle"
    }

    tags = ["gke-node", "${var.environment}-distributed-gradle"]

    # Enable workload identity
    workload_metadata_config {
      mode = "GKE_METADATA"
    }

    # Enable shielded nodes
    shielded_instance_config {
      enable_secure_boot          = true
      enable_vtpm                 = true
      enable_integrity_monitoring = true
    }
  }

  management {
    auto_repair  = true
    auto_upgrade = true
  }

  autoscaling {
    min_node_count = var.min_node_count
    max_node_count = var.max_node_count
  }

  upgrade_settings {
    max_surge       = 1
    max_unavailable = 0
  }
}

# GKE Service Account
resource "google_service_account" "gke" {
  account_id   = "${var.environment}-distributed-gradle-gke"
  display_name = "GKE Service Account for Distributed Gradle"
  description  = "Service account for GKE nodes with minimal required permissions"
}

# IAM bindings for GKE service account
resource "google_project_iam_member" "gke_logging" {
  project = var.project_id
  role    = "roles/logging.logWriter"
  member  = "serviceAccount:${google_service_account.gke.email}"
}

resource "google_project_iam_member" "gke_monitoring" {
  project = var.project_id
  role    = "roles/monitoring.metricWriter"
  member  = "serviceAccount:${google_service_account.gke.email}"
}

resource "google_project_iam_member" "gke_storage" {
  project = var.project_id
  role    = "roles/storage.objectViewer"
  member  = "serviceAccount:${google_service_account.gke.email}"
}

# BigQuery dataset for usage metering
resource "google_bigquery_dataset" "usage_metering" {
  dataset_id                  = "${var.environment}_distributed_gradle_usage"
  friendly_name               = "GKE Usage Metering"
  description                 = "Dataset for GKE usage metering data"
  location                    = var.region
  delete_contents_on_destroy = true
}

# Cloud Storage bucket for backups
resource "google_storage_bucket" "backups" {
  name          = "${var.project_id}-${var.environment}-distributed-gradle-backups"
  location      = var.region
  force_destroy = false

  uniform_bucket_level_access = true

  versioning {
    enabled = true
  }

  lifecycle_rule {
    condition {
      age = var.backup_retention_days
    }
    action {
      type = "Delete"
    }
  }

  labels = {
    environment = var.environment
    project     = "distributed-gradle"
  }
}

# Artifact Registry repository
resource "google_artifact_registry_repository" "distributed_gradle" {
  location      = var.region
  repository_id = "${var.environment}-distributed-gradle"
  description   = "Docker repository for Distributed Gradle images"
  format        = "DOCKER"

  labels = {
    environment = var.environment
    project     = "distributed-gradle"
  }
}

# Kubernetes provider configuration
data "google_client_config" "default" {}

provider "kubernetes" {
  host                   = "https://${google_container_cluster.distributed_gradle.endpoint}"
  token                  = data.google_client_config.default.access_token
  cluster_ca_certificate = base64decode(google_container_cluster.distributed_gradle.master_auth[0].cluster_ca_certificate)
}

provider "helm" {
  kubernetes {
    host                   = "https://${google_container_cluster.distributed_gradle.endpoint}"
    token                  = data.google_client_config.default.access_token
    cluster_ca_certificate = base64decode(google_container_cluster.distributed_gradle.master_auth[0].cluster_ca_certificate)
  }
}

# Storage Classes
resource "kubernetes_storage_class" "ssd" {
  metadata {
    name = "gcp-ssd"
    annotations = {
      "storageclass.kubernetes.io/is-default-class" = "true"
    }
  }

  storage_provisioner    = "pd.csi.storage.gke.io"
  reclaim_policy         = "Retain"
  volume_binding_mode    = "WaitForFirstConsumer"
  allow_volume_expansion = true

  parameters = {
    type = "pd-ssd"
  }

  depends_on = [google_container_cluster.distributed_gradle]
}

# Distributed Gradle Namespace
resource "kubernetes_namespace" "distributed_gradle" {
  metadata {
    name = "distributed-gradle"

    labels = {
      name = "distributed-gradle"
    }
  }

  depends_on = [google_container_cluster.distributed_gradle]
}

# Workload Identity bindings
resource "google_service_account_iam_binding" "workload_identity" {
  service_account_id = google_service_account.gke.name
  role               = "roles/iam.workloadIdentityUser"

  members = [
    "serviceAccount:${var.project_id}.svc.id.goog[distributed-gradle/default]",
  ]
}

# Deploy Distributed Gradle using Helm
resource "helm_release" "distributed_gradle" {
  name       = "distributed-gradle"
  repository = "../../k8s/helm"
  chart      = "distributed-gradle"
  namespace  = kubernetes_namespace.distributed_gradle.metadata[0].name
  version    = "0.1.0"

  values = [
    file("../../cloud/gcp/gke-values.yaml")
  ]

  set {
    name  = "global.registry"
    value = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.distributed_gradle.repository_id}"
  }

  set {
    name  = "global.imageTag"
    value = var.image_tag
  }

  depends_on = [
    google_container_cluster.distributed_gradle,
    kubernetes_namespace.distributed_gradle,
    kubernetes_storage_class.ssd
  ]
}

# Outputs
output "cluster_name" {
  description = "GKE cluster name"
  value       = google_container_cluster.distributed_gradle.name
}

output "cluster_endpoint" {
  description = "GKE cluster endpoint"
  value       = google_container_cluster.distributed_gradle.endpoint
}

output "cluster_ca_certificate" {
  description = "GKE cluster CA certificate"
  value       = google_container_cluster.distributed_gradle.master_auth[0].cluster_ca_certificate
  sensitive   = true
}

output "network_name" {
  description = "VPC network name"
  value       = module.vpc.network_name
}

output "subnet_name" {
  description = "Subnet name"
  value       = module.vpc.subnet_name
}

output "artifact_registry_url" {
  description = "Artifact Registry URL"
  value       = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.distributed_gradle.repository_id}"
}

output "backup_bucket_name" {
  description = "Backup bucket name"
  value       = google_storage_bucket.backups.name
}