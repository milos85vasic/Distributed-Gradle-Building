variable "project_id" {
  description = "The GCP project ID"
  type        = string
}

variable "region" {
  description = "The GCP region for resources"
  type        = string
  default     = "us-central1"
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  default     = "dev"
}

variable "network_cidr" {
  description = "CIDR block for the VPC network"
  type        = string
  default     = "10.0.0.0/16"
}

variable "subnet_cidrs" {
  description = "Map of subnet names to CIDR blocks"
  type        = map(string)
  default = {
    private = "10.0.0.0/24"
    public  = "10.0.1.0/24"
  }
}

variable "enable_private_endpoint" {
  description = "Whether to enable private endpoints for the cluster"
  type        = bool
  default     = true
}

variable "enable_autopilot" {
  description = "Whether to use GKE Autopilot mode"
  type        = bool
  default     = false
}

# Cluster autoscaling variables (only used in standard mode)
variable "min_cpu" {
  description = "Minimum CPU cores for cluster autoscaling"
  type        = number
  default     = 1
}

variable "max_cpu" {
  description = "Maximum CPU cores for cluster autoscaling"
  type        = number
  default     = 10
}

variable "min_memory" {
  description = "Minimum memory (GB) for cluster autoscaling"
  type        = number
  default     = 1
}

variable "max_memory" {
  description = "Maximum memory (GB) for cluster autoscaling"
  type        = number
  default     = 40
}

# Node pool variables (only used in standard mode)
variable "default_node_count" {
  description = "Default number of nodes in the node pool"
  type        = number
  default     = 3
}

variable "min_node_count" {
  description = "Minimum number of nodes in the node pool"
  type        = number
  default     = 1
}

variable "max_node_count" {
  description = "Maximum number of nodes in the node pool"
  type        = number
  default     = 10
}

variable "use_preemptible_nodes" {
  description = "Whether to use preemptible nodes for cost optimization"
  type        = bool
  default     = false
}

variable "node_machine_type" {
  description = "Machine type for GKE nodes"
  type        = string
  default     = "e2-standard-4"
}

# Security variables
variable "enable_binary_authorization" {
  description = "Whether to enable Binary Authorization"
  type        = bool
  default     = true
}

variable "enable_shielded_nodes" {
  description = "Whether to enable Shielded GKE Nodes"
  type        = bool
  default     = true
}

variable "master_ipv4_cidr_block" {
  description = "CIDR block for the GKE master nodes"
  type        = string
  default     = "172.16.0.0/28"
}

# Backup variables
variable "backup_retention_days" {
  description = "Number of days to retain backups"
  type        = number
  default     = 30
}

# Application variables
variable "image_tag" {
  description = "Docker image tag for the application"
  type        = string
  default     = "latest"
}