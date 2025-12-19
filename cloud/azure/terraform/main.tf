terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
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

  backend "azurerm" {
    # Configure your Azure backend here
    # resource_group_name  = "terraform-state"
    # storage_account_name = "terraformstate"
    # container_name       = "tfstate"
    # key                  = "distributed-gradle.tfstate"
  }
}

provider "azurerm" {
  features {
    resource_group {
      prevent_deletion_if_contains_resources = false
    }
  }

  subscription_id = var.subscription_id
}

# Resource Group
resource "azurerm_resource_group" "distributed_gradle" {
  name     = "${var.environment}-distributed-gradle-rg"
  location = var.location

  tags = {
    Environment = var.environment
    Project     = "distributed-gradle"
    ManagedBy   = "terraform"
  }
}

# VNet Module
module "vnet" {
  source = "./modules/vnet"

  name                = "distributed-gradle"
  resource_group_name = azurerm_resource_group.distributed_gradle.name
  location            = azurerm_resource_group.distributed_gradle.location
  address_space       = var.vnet_address_space
  subnets             = var.subnets

  tags = {
    Environment = var.environment
    Project     = "distributed-gradle"
  }
}

# AKS Cluster
resource "azurerm_kubernetes_cluster" "distributed_gradle" {
  name                = "${var.environment}-distributed-gradle"
  location            = azurerm_resource_group.distributed_gradle.location
  resource_group_name = azurerm_resource_group.distributed_gradle.name
  dns_prefix          = "distributed-gradle"
  kubernetes_version  = var.kubernetes_version

  default_node_pool {
    name                = "default"
    node_count          = var.default_node_count
    vm_size             = var.default_node_size
    vnet_subnet_id      = module.vnet.subnet_ids["private"]
    enable_auto_scaling = true
    min_count           = var.min_node_count
    max_count           = var.max_node_count
    max_pods            = 110

    upgrade_settings {
      max_surge = "10%"
    }

    tags = {
      Environment = var.environment
      Project     = "distributed-gradle"
    }
  }

  # Additional Node Pools
  dynamic "node_pool" {
    for_each = var.additional_node_pools
    content {
      name                = node_pool.value.name
      node_count          = node_pool.value.node_count
      vm_size             = node_pool.value.vm_size
      vnet_subnet_id      = module.vnet.subnet_ids["private"]
      enable_auto_scaling = true
      min_count           = node_pool.value.min_count
      max_count           = node_pool.value.max_count
      max_pods            = 110
      mode                = node_pool.value.mode

      upgrade_settings {
        max_surge = "10%"
      }

      tags = {
        Environment = var.environment
        Project     = "distributed-gradle"
      }
    }
  }

  identity {
    type = "SystemAssigned"
  }

  # Azure Monitor integration
  oms_agent {
    log_analytics_workspace_id = azurerm_log_analytics_workspace.distributed_gradle.id
  }

  # Azure Policy
  azure_policy_enabled = true

  # Azure Key Vault integration
  key_vault_secrets_provider {
    secret_rotation_enabled = true
  }

  network_profile {
    network_plugin    = "azure"
    network_policy    = "azure"
    load_balancer_sku = "standard"
    outbound_type     = "loadBalancer"
  }

  tags = {
    Environment = var.environment
    Project     = "distributed-gradle"
  }
}

# Log Analytics Workspace for monitoring
resource "azurerm_log_analytics_workspace" "distributed_gradle" {
  name                = "${var.environment}-distributed-gradle-logs"
  location            = azurerm_resource_group.distributed_gradle.location
  resource_group_name = azurerm_resource_group.distributed_gradle.name
  sku                 = "PerGB2018"
  retention_in_days   = 30

  tags = {
    Environment = var.environment
    Project     = "distributed-gradle"
  }
}

# Application Insights for application monitoring
resource "azurerm_application_insights" "distributed_gradle" {
  name                = "${var.environment}-distributed-gradle-appinsights"
  location            = azurerm_resource_group.distributed_gradle.location
  resource_group_name = azurerm_resource_group.distributed_gradle.name
  application_type    = "web"

  tags = {
    Environment = var.environment
    Project     = "distributed-gradle"
  }
}

# Azure Container Registry
resource "azurerm_container_registry" "distributed_gradle" {
  name                = "${var.environment}distributedgradle"
  resource_group_name = azurerm_resource_group.distributed_gradle.name
  location            = azurerm_resource_group.distributed_gradle.location
  sku                 = "Basic"
  admin_enabled       = true

  tags = {
    Environment = var.environment
    Project     = "distributed-gradle"
  }
}

# Storage Account for backups and logs
resource "azurerm_storage_account" "distributed_gradle" {
  name                     = "${var.environment}distributedgradle"
  resource_group_name      = azurerm_resource_group.distributed_gradle.name
  location                 = azurerm_resource_group.distributed_gradle.location
  account_tier             = "Standard"
  account_replication_type = "GRS"
  account_kind             = "StorageV2"

  tags = {
    Environment = var.environment
    Project     = "distributed-gradle"
  }
}

# Storage Container for backups
resource "azurerm_storage_container" "backups" {
  name                  = "backups"
  storage_account_name  = azurerm_storage_account.distributed_gradle.name
  container_access_type = "private"
}

# Azure Key Vault for secrets
resource "azurerm_key_vault" "distributed_gradle" {
  name                        = "${var.environment}-distributed-gradle-kv"
  location                    = azurerm_resource_group.distributed_gradle.location
  resource_group_name         = azurerm_resource_group.distributed_gradle.name
  enabled_for_disk_encryption = true
  tenant_id                   = data.azurerm_client_config.current.tenant_id
  soft_delete_retention_days  = 7
  purge_protection_enabled    = false

  sku_name = "standard"

  access_policy {
    tenant_id = data.azurerm_client_config.current.tenant_id
    object_id = data.azurerm_client_config.current.object_id

    key_permissions = [
      "Get",
    ]

    secret_permissions = [
      "Get",
    ]

    storage_permissions = [
      "Get",
    ]
  }

  tags = {
    Environment = var.environment
    Project     = "distributed-gradle"
  }
}

data "azurerm_client_config" "current" {}

# Kubernetes provider configuration
data "azurerm_kubernetes_cluster" "distributed_gradle" {
  name                = azurerm_kubernetes_cluster.distributed_gradle.name
  resource_group_name = azurerm_resource_group.distributed_gradle.name
}

provider "kubernetes" {
  host                   = data.azurerm_kubernetes_cluster.distributed_gradle.kube_config.0.host
  client_certificate     = base64decode(data.azurerm_kubernetes_cluster.distributed_gradle.kube_config.0.client_certificate)
  client_key             = base64decode(data.azurerm_kubernetes_cluster.distributed_gradle.kube_config.0.client_key)
  cluster_ca_certificate = base64decode(data.azurerm_kubernetes_cluster.distributed_gradle.kube_config.0.cluster_ca_certificate)
}

provider "helm" {
  kubernetes {
    host                   = data.azurerm_kubernetes_cluster.distributed_gradle.kube_config.0.host
    client_certificate     = base64decode(data.azurerm_kubernetes_cluster.distributed_gradle.kube_config.0.client_certificate)
    client_key             = base64decode(data.azurerm_kubernetes_cluster.distributed_gradle.kube_config.0.client_key)
    cluster_ca_certificate = base64decode(data.azurerm_kubernetes_cluster.distributed_gradle.kube_config.0.cluster_ca_certificate)
  }
}

# Storage Classes
resource "kubernetes_storage_class" "premium_ssd" {
  metadata {
    name = "azure-premium-ssd"
    annotations = {
      "storageclass.kubernetes.io/is-default-class" = "true"
    }
  }

  storage_provisioner    = "kubernetes.io/azure-disk"
  reclaim_policy         = "Retain"
  volume_binding_mode    = "WaitForFirstConsumer"
  allow_volume_expansion = true

  parameters = {
    storageaccounttype = "Premium_LRS"
    kind               = "Managed"
    cachingmode        = "ReadOnly"
  }

  depends_on = [azurerm_kubernetes_cluster.distributed_gradle]
}

# Distributed Gradle Namespace
resource "kubernetes_namespace" "distributed_gradle" {
  metadata {
    name = "distributed-gradle"

    labels = {
      name = "distributed-gradle"
    }
  }

  depends_on = [azurerm_kubernetes_cluster.distributed_gradle]
}

# Deploy Distributed Gradle using Helm
resource "helm_release" "distributed_gradle" {
  name       = "distributed-gradle"
  repository = "../../k8s/helm"
  chart      = "distributed-gradle"
  namespace  = kubernetes_namespace.distributed_gradle.metadata[0].name
  version    = "0.1.0"

  values = [
    file("../../cloud/azure/aks-values.yaml")
  ]

  set {
    name  = "global.registry"
    value = azurerm_container_registry.distributed_gradle.login_server
  }

  set {
    name  = "global.imageTag"
    value = var.image_tag
  }

  depends_on = [
    azurerm_kubernetes_cluster.distributed_gradle,
    kubernetes_namespace.distributed_gradle,
    kubernetes_storage_class.premium_ssd
  ]
}

# Outputs
output "resource_group_name" {
  description = "Azure Resource Group name"
  value       = azurerm_resource_group.distributed_gradle.name
}

output "aks_cluster_name" {
  description = "AKS cluster name"
  value       = azurerm_kubernetes_cluster.distributed_gradle.name
}

output "aks_cluster_id" {
  description = "AKS cluster ID"
  value       = azurerm_kubernetes_cluster.distributed_gradle.id
}

output "kube_config" {
  description = "Kubernetes configuration"
  value       = azurerm_kubernetes_cluster.distributed_gradle.kube_config_raw
  sensitive   = true
}

output "acr_login_server" {
  description = "Azure Container Registry login server"
  value       = azurerm_container_registry.distributed_gradle.login_server
}

output "storage_account_name" {
  description = "Azure Storage Account name"
  value       = azurerm_storage_account.distributed_gradle.name
}

output "key_vault_name" {
  description = "Azure Key Vault name"
  value       = azurerm_key_vault.distributed_gradle.name
}