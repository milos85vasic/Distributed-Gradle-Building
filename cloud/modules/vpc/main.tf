# Multi-Cloud VPC Module
# This module provides a unified interface for VPC/networking across cloud providers

variable "cloud_provider" {
  description = "Cloud provider (aws, azure, gcp)"
  type        = string
  validation {
    condition     = contains(["aws", "azure", "gcp"], var.cloud_provider)
    error_message = "Cloud provider must be one of: aws, azure, gcp"
  }
}

variable "name" {
  description = "Name of the VPC/network"
  type        = string
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "dev"
}

variable "network_cidr" {
  description = "Primary network CIDR block"
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

variable "enable_private_endpoints" {
  description = "Whether to enable private endpoints"
  type        = bool
  default     = true
}

variable "tags" {
  description = "Tags to apply to resources"
  type        = map(string)
  default     = {}
}

# Cloud-specific module selection
module "aws_vpc" {
  count  = var.cloud_provider == "aws" ? 1 : 0
  source = "../../aws/terraform/modules/vpc"

  name                    = var.name
  cidr                    = var.network_cidr
  private_subnets         = [var.subnet_cidrs.private]
  public_subnets          = [var.subnet_cidrs.public]
  enable_nat_gateway      = var.enable_private_endpoints
  single_nat_gateway      = true
  enable_dns_hostnames    = true
  enable_dns_support      = true

  tags = merge(var.tags, {
    Environment = var.environment
    Project     = "distributed-gradle"
  })
}

module "azure_vnet" {
  count  = var.cloud_provider == "azure" ? 1 : 0
  source = "../../azure/terraform/modules/vnet"

  name                = var.name
  resource_group_name = var.name
  location            = var.location != "" ? var.location : "East US"
  address_space       = [var.network_cidr]
  subnets = {
    private = {
      address_prefixes = [var.subnet_cidrs.private]
      service_endpoints = var.enable_private_endpoints ? ["Microsoft.Storage", "Microsoft.KeyVault"] : []
    }
    public = {
      address_prefixes = [var.subnet_cidrs.public]
    }
  }

  tags = merge(var.tags, {
    environment = var.environment
    project     = "distributed-gradle"
  })
}

module "gcp_vpc" {
  count  = var.cloud_provider == "gcp" ? 1 : 0
  source = "../../gcp/terraform/modules/vpc"

  name                    = var.name
  project_id              = var.project_id != "" ? var.project_id : "my-project"
  region                  = var.region != "" ? var.region : "us-central1"
  network_cidr            = var.network_cidr
  subnet_cidrs            = var.subnet_cidrs
  enable_private_endpoint = var.enable_private_endpoints

  labels = merge(var.tags, {
    environment = var.environment
    project     = "distributed-gradle"
  })
}

# Additional variables for cloud-specific requirements
variable "location" {
  description = "Azure location"
  type        = string
  default     = ""
}

variable "project_id" {
  description = "GCP project ID"
  type        = string
  default     = ""
}

variable "region" {
  description = "GCP region"
  type        = string
  default     = ""
}

# Outputs
output "network_id" {
  description = "Network/VPC ID"
  value = coalesce(
    try(module.aws_vpc[0].vpc_id, ""),
    try(module.azure_vnet[0].vnet_id, ""),
    try(module.gcp_vpc[0].network_self_link, "")
  )
}

output "private_subnet_id" {
  description = "Private subnet ID"
  value = coalesce(
    try(module.aws_vpc[0].private_subnets[0], ""),
    try(module.azure_vnet[0].subnet_ids["private"], ""),
    try(module.gcp_vpc[0].subnet_self_link, "")
  )
}

output "public_subnet_id" {
  description = "Public subnet ID"
  value = coalesce(
    try(module.aws_vpc[0].public_subnets[0], ""),
    try(module.azure_vnet[0].subnet_ids["public"], ""),
    ""
  )
}