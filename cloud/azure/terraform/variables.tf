# Azure Subscription
variable "subscription_id" {
  description = "Azure Subscription ID"
  type        = string
  sensitive   = true
}

# Location
variable "location" {
  description = "Azure region to deploy resources"
  type        = string
  default     = "East US"
}

# Environment
variable "environment" {
  description = "Environment name (staging, production)"
  type        = string
  default     = "staging"
}

# VNet Configuration
variable "vnet_address_space" {
  description = "Address space for VNet"
  type        = list(string)
  default     = ["10.0.0.0/16"]
}

variable "subnets" {
  description = "Subnet configurations"
  type = map(object({
    address_prefixes = list(string)
    service_endpoints = optional(list(string), [])
  }))
  default = {
    private = {
      address_prefixes = ["10.0.1.0/24"]
      service_endpoints = ["Microsoft.Storage", "Microsoft.KeyVault"]
    }
    public = {
      address_prefixes = ["10.0.100.0/24"]
    }
  }
}

# AKS Configuration
variable "kubernetes_version" {
  description = "Kubernetes version for AKS cluster"
  type        = string
  default     = "1.28.0"
}

variable "default_node_count" {
  description = "Default number of worker nodes"
  type        = number
  default     = 3
}

variable "default_node_size" {
  description = "VM size for default node pool"
  type        = string
  default     = "Standard_DS2_v2"
}

variable "min_node_count" {
  description = "Minimum number of worker nodes"
  type        = number
  default     = 1
}

variable "max_node_count" {
  description = "Maximum number of worker nodes"
  type        = number
  default     = 10
}

variable "additional_node_pools" {
  description = "Additional node pools configuration"
  type = map(object({
    name       = string
    node_count = number
    vm_size    = string
    min_count  = number
    max_count  = number
    mode       = string
  }))
  default = {
    worker-pool = {
      name       = "worker-pool"
      node_count = 5
      vm_size    = "Standard_DS2_v2"
      min_count  = 1
      max_count  = 20
      mode       = "User"
    }
  }
}

# Container Configuration
variable "image_tag" {
  description = "Container image tag"
  type        = string
  default     = "latest"
}

# Domain Configuration (optional)
variable "domain_name" {
  description = "Domain name for Azure DNS (optional)"
  type        = string
  default     = ""
}

variable "create_dns_zone" {
  description = "Create Azure DNS zone"
  type        = bool
  default     = false
}

# Monitoring Configuration
variable "enable_monitoring" {
  description = "Enable Azure Monitor and Application Insights"
  type        = bool
  default     = true
}

variable "log_analytics_retention_days" {
  description = "Log Analytics workspace retention in days"
  type        = number
  default     = 30
}

# Security Configuration
variable "enable_encryption" {
  description = "Enable encryption for storage and Key Vault"
  type        = bool
  default     = true
}

variable "enable_private_endpoints" {
  description = "Enable private endpoints for Azure services"
  type        = bool
  default     = false
}

# Backup Configuration
variable "enable_backups" {
  description = "Enable automated backups"
  type        = bool
  default     = true
}

variable "backup_retention_days" {
  description = "Number of days to retain backups"
  type        = number
  default     = 30
}

# Cost Optimization
variable "enable_cost_allocation_tags" {
  description = "Enable cost allocation tags"
  type        = bool
  default     = true
}

# Network Security
variable "allowed_ip_ranges" {
  description = "IP ranges allowed to access the cluster"
  type        = list(string)
  default     = ["0.0.0.0/0"] # Restrict in production
}