# GCP GKE Deployment for Distributed Gradle Building

This directory contains Google Cloud Platform-specific deployment configurations and infrastructure-as-code for deploying the Distributed Gradle Building system on Google Kubernetes Engine (GKE).

## Prerequisites

- Google Cloud SDK (gcloud) configured with appropriate permissions
- Terraform 1.0+
- kubectl configured for GKE access
- Helm 3.0+

## Quick Start

### 1. Infrastructure Setup

```bash
cd cloud/gcp/terraform

# Initialize Terraform
terraform init

# Review the plan
terraform plan

# Apply the infrastructure
terraform apply
```

### 2. Kubernetes Deployment

```bash
# Update kubeconfig for GKE cluster
gcloud container clusters get-credentials dev-distributed-gradle --region us-central1 --project your-project-id

# Deploy using Helm
helm install distributed-gradle ../../k8s/helm/distributed-gradle \
  --values ../gke-values.yaml \
  --namespace distributed-gradle \
  --create-namespace
```

### 3. Access the Application

```bash
# Get GCLB endpoint
kubectl get ingress -n distributed-gradle

# Access Grafana
kubectl get secret -n distributed-gradle distributed-gradle-grafana -o jsonpath="{.data.admin-password}" | base64 --decode
```

## Architecture

The GCP deployment includes:

- **GKE Cluster**: Managed Kubernetes control plane with Autopilot or standard mode
- **Node Pools**: Auto-scaling worker nodes with preemptible instances for cost optimization
- **GCLB Ingress**: Google Cloud Load Balancer for external access with SSL termination
- **GCS Storage**: Persistent disks for cache and logs
- **Cloud Monitoring**: Centralized logging and monitoring with Cloud Operations
- **Cloud DNS**: DNS management (optional)

## Configuration

### GKE Cluster Configuration

- **Kubernetes Version**: 1.28+
- **Node Types**: e2-standard-4 instances for cost optimization
- **Auto-scaling**: Cluster and horizontal pod autoscaling
- **Networking**: VPC-native networking with private clusters
- **Security**: Workload Identity, Binary Authorization, Shielded Nodes

### Storage Classes

- **gcp-ssd**: SSD persistent disks for cache data
- **gcp-ssd**: SSD persistent disks for application logs

### Security

- **Workload Identity**: Service accounts with minimal required permissions
- **VPC Firewall**: Network isolation between components
- **Secret Manager**: Secrets management integration
- **Binary Authorization**: Container image security

## Monitoring & Observability

The deployment includes:

- **Cloud Monitoring**: Infrastructure and application monitoring
- **Cloud Logging**: Centralized logging with structured logs
- **Prometheus**: Metrics collection from all services
- **Grafana**: Visualization dashboards
- **Cloud Trace**: Distributed tracing (optional)

## Cost Optimization

- **Preemptible Instances**: For worker nodes in standard mode
- **Autopilot Mode**: Fully managed node lifecycle
- **Auto-scaling**: Scale down during off-hours
- **Committed Use Discounts**: For consistent workloads
- **GCS Storage Classes**: Cost-effective storage options

## Troubleshooting

### Common Issues

1. **GCLB Health Checks Failing**
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
gcloud logging read "resource.type=k8s_cluster AND resource.labels.cluster_name=dev-distributed-gradle" --limit=50
```

## Backup & Recovery

- **GCS Backups**: Automated backups of persistent disks
- **Cloud Storage**: Backup storage for configuration and build artifacts
- **Cross-region Replication**: Disaster recovery setup
- **Backup for GKE**: Persistent volume snapshots

## Security Considerations

- **VPC Service Controls**: Network security perimeters
- **Network Policies**: Kubernetes network policies for pod-to-pod communication
- **IAM Policies**: Least privilege access for service accounts
- **Encryption**: Data encryption at rest and in transit
- **Container Analysis**: Vulnerability scanning for container images</content>
</xai:function_call"> 

<xai:function_call name="todowrite">
<parameter name="todos">[{"content":"Create GCP GKE Terraform infrastructure and configurations","status":"completed","priority":"medium","id":"gcp_gke_terraform"},{"content":"Create GCP-optimized Helm values.yaml for GKE deployment","status":"completed","priority":"medium","id":"gcp_gke_values"},{"content":"Create comprehensive README for GCP GKE deployment","status":"completed","priority":"medium","id":"gcp_gke_readme"}]