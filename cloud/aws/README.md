# AWS EKS Deployment for Distributed Gradle Building

This directory contains AWS-specific deployment configurations and infrastructure-as-code for deploying the Distributed Gradle Building system on Amazon EKS.

## Prerequisites

- AWS CLI configured with appropriate permissions
- Terraform 1.0+
- kubectl configured for EKS access
- Helm 3.0+

## Quick Start

### 1. Infrastructure Setup

```bash
cd cloud/aws/terraform
terraform init
terraform plan
terraform apply
```

### 2. Kubernetes Deployment

```bash
# Update kubeconfig for EKS cluster
aws eks update-kubeconfig --region us-east-1 --name distributed-gradle

# Deploy using Helm
helm install distributed-gradle ../../k8s/helm/distributed-gradle \
  --values ../eks-values.yaml \
  --namespace distributed-gradle \
  --create-namespace
```

### 3. Access the Application

```bash
# Get ALB endpoint
kubectl get ingress -n distributed-gradle

# Access Grafana
kubectl get secret -n distributed-gradle distributed-gradle-grafana -o jsonpath="{.data.admin-password}" | base64 --decode
```

## Architecture

The AWS deployment includes:

- **EKS Cluster**: Managed Kubernetes control plane
- **Node Groups**: Auto-scaling worker nodes with spot instances
- **ALB Ingress**: Application Load Balancer for external access
- **EBS Storage**: Persistent volumes for cache and logs
- **CloudWatch**: Centralized logging and monitoring
- **Route 53**: DNS management (optional)

## Configuration

### EKS Cluster Configuration

- **Kubernetes Version**: 1.28+
- **Node Types**: Mixed instance types for cost optimization
- **Auto-scaling**: Cluster and horizontal pod autoscaling
- **Networking**: VPC with public/private subnets

### Storage Classes

- **gp3**: General purpose SSD for cache data
- **gp3**: General purpose SSD for application logs

### Security

- **IAM Roles**: Service accounts with minimal required permissions
- **Security Groups**: Network isolation between components
- **Secrets Management**: AWS Secrets Manager integration

## Monitoring & Observability

The deployment includes:

- **Prometheus**: Metrics collection from all services
- **Grafana**: Visualization dashboards
- **CloudWatch**: Infrastructure monitoring
- **X-Ray**: Distributed tracing (optional)

## Cost Optimization

- **Spot Instances**: For worker nodes
- **Auto-scaling**: Scale down during off-hours
- **EBS Optimization**: Right-size storage volumes
- **Reserved Instances**: For consistent workloads

## Troubleshooting

### Common Issues

1. **ALB Health Checks Failing**
   ```bash
   kubectl logs -n distributed-gradle deployment/distributed-gradle-coordinator
   kubectl describe ingress -n distributed-gradle
   ```

2. **PVC Pending**
   ```bash
   kubectl get pvc -n distributed-gradle
   kubectl describe pvc <pvc-name> -n distributed-gradle
   ```

3. **Pod Eviction**
   ```bash
   kubectl get events -n distributed-gradle --sort-by=.metadata.creationTimestamp
   ```

### Logs

```bash
# Application logs
kubectl logs -n distributed-gradle -l app=distributed-gradle-coordinator

# Infrastructure logs
aws logs tail /aws/eks/distributed-gradle/cluster --follow
```

## Backup & Recovery

- **EBS Snapshots**: Automated snapshots of persistent volumes
- **S3 Backups**: Configuration and build artifacts
- **Cross-region Replication**: Disaster recovery setup

## Security Considerations

- **Network Policies**: Kubernetes network policies for pod-to-pod communication
- **IAM Policies**: Least privilege access for service accounts
- **Encryption**: Data encryption at rest and in transit
- **Vulnerability Scanning**: Container image scanning with ECR