# AWS Region
variable "aws_region" {
  description = "AWS region to deploy resources"
  type        = string
  default     = "us-east-1"
}

# Environment
variable "environment" {
  description = "Environment name (staging, production)"
  type        = string
  default     = "staging"
}

# VPC Configuration
variable "vpc_cidr" {
  description = "CIDR block for VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "availability_zones" {
  description = "List of availability zones"
  type        = list(string)
  default     = ["us-east-1a", "us-east-1b", "us-east-1c"]
}

variable "private_subnets" {
  description = "Private subnet CIDR blocks"
  type        = list(string)
  default     = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
}

variable "public_subnets" {
  description = "Public subnet CIDR blocks"
  type        = list(string)
  default     = ["10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"]
}

# EKS Configuration
variable "cluster_version" {
  description = "Kubernetes version for EKS cluster"
  type        = string
  default     = "1.28"
}

variable "node_instance_types" {
  description = "Instance types for EKS worker nodes"
  type        = list(string)
  default     = ["t3.medium", "t3.large"]
}

variable "min_nodes" {
  description = "Minimum number of worker nodes"
  type        = number
  default     = 1
}

variable "max_nodes" {
  description = "Maximum number of worker nodes"
  type        = number
  default     = 10
}

variable "desired_nodes" {
  description = "Desired number of worker nodes"
  type        = number
  default     = 3
}

variable "use_spot_instances" {
  description = "Use spot instances for cost optimization"
  type        = bool
  default     = false
}

# Container Configuration
variable "container_registry" {
  description = "Container registry URL"
  type        = string
  default     = "your-registry.com"
}

variable "image_tag" {
  description = "Container image tag"
  type        = string
  default     = "latest"
}

# Domain Configuration (optional)
variable "domain_name" {
  description = "Domain name for Route 53 (optional)"
  type        = string
  default     = ""
}

variable "create_route53_record" {
  description = "Create Route 53 record for the ALB"
  type        = bool
  default     = false
}

# Monitoring Configuration
variable "enable_monitoring" {
  description = "Enable Prometheus/Grafana monitoring"
  type        = bool
  default     = true
}

variable "enable_logging" {
  description = "Enable CloudWatch logging"
  type        = bool
  default     = true
}

# Backup Configuration
variable "enable_backups" {
  description = "Enable automated EBS snapshots"
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

# Security
variable "enable_encryption" {
  description = "Enable encryption for EBS volumes and S3 buckets"
  type        = bool
  default     = true
}

variable "allowed_cidr_blocks" {
  description = "CIDR blocks allowed to access the cluster"
  type        = list(string)
  default     = ["0.0.0.0/0"] # Restrict in production
}