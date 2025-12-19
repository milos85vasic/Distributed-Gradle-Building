terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
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

  backend "s3" {
    # Configure your S3 backend here
    # bucket = "your-terraform-state-bucket"
    # key    = "distributed-gradle/terraform.tfstate"
    # region = "us-east-1"
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Environment = var.environment
      Project     = "distributed-gradle"
      ManagedBy   = "terraform"
    }
  }
}

# VPC Module
module "vpc" {
  source = "./modules/vpc"

  name = "distributed-gradle"
  cidr = var.vpc_cidr

  azs             = var.availability_zones
  private_subnets = var.private_subnets
  public_subnets  = var.public_subnets

  enable_nat_gateway   = true
  single_nat_gateway   = false
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    "kubernetes.io/cluster/${local.cluster_name}" = "shared"
  }

  public_subnet_tags = {
    "kubernetes.io/cluster/${local.cluster_name}" = "shared"
    "kubernetes.io/role/elb"                      = "1"
  }

  private_subnet_tags = {
    "kubernetes.io/cluster/${local.cluster_name}" = "shared"
    "kubernetes.io/role/internal-elb"             = "1"
  }
}

# EKS Cluster Module
module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 19.0"

  cluster_name    = local.cluster_name
  cluster_version = var.cluster_version

  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets

  cluster_endpoint_public_access = true

  # EKS Managed Node Groups
  eks_managed_node_groups = {
    general = {
      name = "general-purpose"

      instance_types = var.node_instance_types
      capacity_type  = "ON_DEMAND"

      min_size     = var.min_nodes
      max_size     = var.max_nodes
      desired_size = var.desired_nodes

      # Use spot instances for cost optimization
      capacity_type = var.use_spot_instances ? "SPOT" : "ON_DEMAND"

      # IAM role for service accounts
      iam_role_additional_policies = {
        AmazonEC2ContainerRegistryReadOnly = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
        AmazonEKS_CNI_Policy               = "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
        AmazonEKSWorkerNodePolicy          = "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"
        AmazonEC2ContainerRegistryPowerUser = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryPowerUser"
      }

      tags = {
        "k8s.io/cluster-autoscaler/enabled"               = "true"
        "k8s.io/cluster-autoscaler/${local.cluster_name}" = "owned"
      }
    }
  }

  # Cluster access entry
  enable_cluster_creator_admin_permissions = true

  # Add-ons
  cluster_addons = {
    coredns = {
      most_recent = true
    }
    eks-pod-identity-agent = {
      most_recent = true
    }
    kube-proxy = {
      most_recent = true
    }
    vpc-cni = {
      most_recent    = true
      before_compute = true
      configuration_values = jsonencode({
        env = {
          # Reference docs https://docs.aws.amazon.com/eks/latest/userguide/cni-increase-ip-addresses.html
          ENABLE_IPv4 = "true"
        }
      })
    }
    aws-ebs-csi-driver = {
      most_recent = true
    }
  }

  tags = {
    Environment = var.environment
    Project     = "distributed-gradle"
  }
}

# IRSA for ALB Controller
module "lb_role" {
  source = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"

  role_name                              = "${local.cluster_name}-alb-controller"
  attach_load_balancer_controller_policy = true

  oidc_providers = {
    main = {
      provider_arn               = module.eks.oidc_provider_arn
      namespace_service_accounts = ["kube-system:aws-load-balancer-controller"]
    }
  }
}

# ALB Controller Helm Chart
resource "helm_release" "aws_load_balancer_controller" {
  name       = "aws-load-balancer-controller"
  repository = "https://aws.github.io/eks-charts"
  chart      = "aws-load-balancer-controller"
  namespace  = "kube-system"
  version    = "1.6.2"

  set {
    name  = "clusterName"
    value = local.cluster_name
  }

  set {
    name  = "serviceAccount.create"
    value = "true"
  }

  set {
    name  = "serviceAccount.name"
    value = "aws-load-balancer-controller"
  }

  set {
    name  = "serviceAccount.annotations.eks\\.amazonaws\\.com/role-arn"
    value = module.lb_role.iam_role_arn
  }

  set {
    name  = "vpcId"
    value = module.vpc.vpc_id
  }

  set {
    name  = "region"
    value = var.aws_region
  }

  depends_on = [module.eks]
}

# IRSA for EBS CSI Driver (if not using addon)
module "ebs_csi_role" {
  source = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"

  role_name             = "${local.cluster_name}-ebs-csi"
  attach_ebs_csi_policy = true

  oidc_providers = {
    main = {
      provider_arn               = module.eks.oidc_provider_arn
      namespace_service_accounts = ["kube-system:ebs-csi-controller-sa"]
    }
  }
}

# Cluster Autoscaler
module "cluster_autoscaler" {
  source  = "terraform-aws-modules/eks/aws//modules/cluster-autoscaler"
  version = "~> 19.0"

  cluster_name                     = local.cluster_name
  cluster_identity_oidc_issuer     = module.eks.cluster_oidc_issuer_url
  cluster_identity_oidc_issuer_arn = module.eks.oidc_provider_arn

  depends_on = [module.eks]
}

# Kubernetes provider configuration
data "aws_eks_cluster_auth" "this" {
  name = local.cluster_name
}

provider "kubernetes" {
  host                   = module.eks.cluster_endpoint
  cluster_ca_certificate = base64decode(module.eks.cluster_certificate_authority_data)
  token                  = data.aws_eks_cluster_auth.this.token
}

provider "helm" {
  kubernetes {
    host                   = module.eks.cluster_endpoint
    cluster_ca_certificate = base64decode(module.eks.cluster_certificate_authority_data)
    token                  = data.aws_eks_cluster_auth.this.token
  }
}

# Storage Classes
resource "kubernetes_storage_class" "gp3" {
  metadata {
    name = "ebs-gp3"
    annotations = {
      "storageclass.kubernetes.io/is-default-class" = "true"
    }
  }

  storage_provisioner    = "ebs.csi.aws.com"
  reclaim_policy         = "Retain"
  volume_binding_mode    = "WaitForFirstConsumer"
  allow_volume_expansion = true

  parameters = {
    type      = "gp3"
    encrypted = "true"
    fsType    = "ext4"
  }

  depends_on = [module.eks]
}

# Distributed Gradle Namespace
resource "kubernetes_namespace" "distributed_gradle" {
  metadata {
    name = "distributed-gradle"

    labels = {
      name = "distributed-gradle"
    }
  }

  depends_on = [module.eks]
}

# Deploy Distributed Gradle using Helm
resource "helm_release" "distributed_gradle" {
  name       = "distributed-gradle"
  repository = "../../k8s/helm"
  chart      = "distributed-gradle"
  namespace  = kubernetes_namespace.distributed_gradle.metadata[0].name
  version    = "0.1.0"

  values = [
    file("../../cloud/aws/eks-values.yaml")
  ]

  set {
    name  = "global.registry"
    value = var.container_registry
  }

  set {
    name  = "global.imageTag"
    value = var.image_tag
  }

  depends_on = [
    module.eks,
    kubernetes_namespace.distributed_gradle,
    helm_release.aws_load_balancer_controller
  ]
}

# Outputs
output "cluster_name" {
  description = "EKS cluster name"
  value       = local.cluster_name
}

output "cluster_endpoint" {
  description = "EKS cluster endpoint"
  value       = module.eks.cluster_endpoint
}

output "cluster_arn" {
  description = "EKS cluster ARN"
  value       = module.eks.cluster_arn
}

output "oidc_provider_arn" {
  description = "OIDC provider ARN"
  value       = module.eks.oidc_provider_arn
}

output "vpc_id" {
  description = "VPC ID"
  value       = module.vpc.vpc_id
}

output "private_subnets" {
  description = "Private subnet IDs"
  value       = module.vpc.private_subnets
}

output "public_subnets" {
  description = "Public subnet IDs"
  value       = module.vpc.public_subnets
}

output "distributed_gradle_alb_dns" {
  description = "ALB DNS name for Distributed Gradle"
  value       = data.kubernetes_ingress.distributed_gradle.status.0.load_balancer.0.ingress.0.hostname
}

# Data source for ALB DNS
data "kubernetes_ingress" "distributed_gradle" {
  metadata {
    name      = "distributed-gradle"
    namespace = kubernetes_namespace.distributed_gradle.metadata[0].name
  }

  depends_on = [helm_release.distributed_gradle]
}

locals {
  cluster_name = "${var.environment}-distributed-gradle"
}