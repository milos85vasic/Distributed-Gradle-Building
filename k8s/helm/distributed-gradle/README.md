# Distributed Gradle Building Helm Chart

This Helm chart deploys the Distributed Gradle Building system on a Kubernetes cluster.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+

## Installing the Chart

To install the chart with the release name `my-release`:

```bash
helm install my-release ./distributed-gradle
```

## Uninstalling the Chart

To uninstall the chart with the release name `my-release`:

```bash
helm uninstall my-release
```

## Configuration

The following table lists the configurable parameters of the Distributed Gradle chart and their default values.

### Global Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `global.namespace` | Namespace for all resources | `staging` |
| `global.environment` | Environment label | `staging` |
| `global.registry` | Docker registry URL | `your-registry.com` |
| `global.imageTag` | Global image tag | `staging` |

### Cache Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `cache.enabled` | Enable cache component | `true` |
| `cache.replicas` | Number of cache replicas | `1` |
| `cache.image.repository` | Cache image repository | `distributed-gradle-cache` |
| `cache.image.tag` | Cache image tag (overrides global) | `""` |
| `cache.image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `cache.port` | Cache service port | `8083` |
| `cache.resources.requests.memory` | Memory request | `512Mi` |
| `cache.resources.requests.cpu` | CPU request | `250m` |
| `cache.resources.limits.memory` | Memory limit | `1Gi` |
| `cache.resources.limits.cpu` | CPU limit | `500m` |
| `cache.persistence.data.enabled` | Enable data persistence | `true` |
| `cache.persistence.data.size` | Data PVC size | `100Gi` |
| `cache.persistence.data.storageClass` | Data storage class | `standard` |
| `cache.persistence.logs.enabled` | Enable logs persistence | `true` |
| `cache.persistence.logs.size` | Logs PVC size | `10Gi` |
| `cache.persistence.logs.storageClass` | Logs storage class | `standard` |

### Coordinator Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `coordinator.enabled` | Enable coordinator component | `true` |
| `coordinator.replicas` | Number of coordinator replicas | `1` |
| `coordinator.image.repository` | Coordinator image repository | `distributed-gradle-coordinator` |
| `coordinator.image.tag` | Coordinator image tag | `""` |
| `coordinator.image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `coordinator.port` | Coordinator service port | `8080` |
| `coordinator.resources.requests.memory` | Memory request | `512Mi` |
| `coordinator.resources.requests.cpu` | CPU request | `250m` |
| `coordinator.resources.limits.memory` | Memory limit | `1Gi` |
| `coordinator.resources.limits.cpu` | CPU limit | `500m` |

### Monitor Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `monitor.enabled` | Enable monitor component | `true` |
| `monitor.replicas` | Number of monitor replicas | `1` |
| `monitor.image.repository` | Monitor image repository | `distributed-gradle-monitor` |
| `monitor.image.tag` | Monitor image tag | `""` |
| `monitor.image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `monitor.port` | Monitor service port | `8082` |
| `monitor.resources.requests.memory` | Memory request | `256Mi` |
| `monitor.resources.requests.cpu` | CPU request | `100m` |
| `monitor.resources.limits.memory` | Memory limit | `512Mi` |
| `monitor.resources.limits.cpu` | CPU limit | `200m` |

### Worker Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `worker.enabled` | Enable worker component | `true` |
| `worker.replicas` | Number of worker replicas | `3` |
| `worker.image.repository` | Worker image repository | `distributed-gradle-worker` |
| `worker.image.tag` | Worker image tag | `""` |
| `worker.image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `worker.port` | Worker service port | `8081` |
| `worker.resources.requests.memory` | Memory request | `256Mi` |
| `worker.resources.requests.cpu` | CPU request | `100m` |
| `worker.resources.limits.memory` | Memory limit | `512Mi` |
| `worker.resources.limits.cpu` | CPU limit | `200m` |

### Monitoring Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `monitoring.prometheus.enabled` | Enable Prometheus | `true` |
| `monitoring.prometheus.scrapeInterval` | Scrape interval | `15s` |
| `monitoring.prometheus.evaluationInterval` | Evaluation interval | `15s` |
| `monitoring.prometheus.alertmanagers` | Alertmanager targets | `[]` |
| `monitoring.grafana.enabled` | Enable Grafana | `true` |

### Ingress Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `ingress.enabled` | Enable ingress | `true` |
| `ingress.className` | Ingress class name | `""` |
| `ingress.annotations` | Ingress annotations | `{}` |
| `ingress.hosts` | Ingress hosts configuration | See values.yaml |
| `ingress.tls` | Ingress TLS configuration | `[]` |

## Example Configuration

```yaml
global:
  namespace: production
  environment: production
  registry: my-registry.com
  imageTag: v1.0.0

cache:
  replicas: 2
  resources:
    requests:
      memory: "1Gi"
      cpu: "500m"
    limits:
      memory: "2Gi"
      cpu: "1000m"

worker:
  replicas: 5

ingress:
  enabled: true
  hosts:
    - host: gradle.example.com
      paths:
        - path: /
          pathType: Prefix
```

## Components

This chart deploys the following components:

- **Cache**: Distributed caching service for build artifacts
- **Coordinator**: Central coordination service for build orchestration
- **Monitor**: Monitoring and alerting service
- **Worker**: Build execution workers (scalable)

Each component can be individually enabled/disabled and configured.