---
title: "Live Demo"
weight: 60
---

# Live Demo Environment

Experience the Distributed Gradle Building System with our interactive demo environment. Test builds, explore performance metrics, and see distributed processing in action.

## 🚀 Quick Demo

<div id="demo-environment">
  <div class="demo-controls">
    <div class="control-section">
      <h3>Demo Configuration</h3>
      <div class="control-group">
        <label for="demo-endpoint">Demo API Endpoint:</label>
        <input type="text" id="demo-endpoint" value="http://demo.distributed-gradle-building.com:8080" readonly>
        <small>Pre-configured demo environment</small>
      </div>
    </div>

    <div class="control-section">
      <h3>Sample Projects</h3>
      <div class="project-selector">
        <div class="project-card" data-project="simple-java">
          <h4>Simple Java App</h4>
          <p>Basic Java application with unit tests</p>
          <div class="project-stats">
            <span>⏱️ ~30s build time</span>
            <span>📦 15 dependencies</span>
          </div>
        </div>

        <div class="project-card" data-project="multi-module">
          <h4>Multi-Module Project</h4>
          <p>Complex project with 5 modules</p>
          <div class="project-stats">
            <span>⏱️ ~2min build time</span>
            <span>📦 45 dependencies</span>
          </div>
        </div>

        <div class="project-card" data-project="spring-boot">
          <h4>Spring Boot App</h4>
          <p>Full-stack web application</p>
          <div class="project-stats">
            <span>⏱️ ~1.5min build time</span>
            <span>📦 80+ dependencies</span>
          </div>
        </div>

        <div class="project-card" data-project="android">
          <h4>Android App</h4>
          <p>Mobile application project</p>
          <div class="project-stats">
            <span>⏱️ ~3min build time</span>
            <span>📦 120+ dependencies</span>
          </div>
        </div>
      </div>
    </div>
  </div>

  <div class="demo-actions">
    <button id="start-demo-build" class="btn-primary-large">
      🚀 Start Distributed Build Demo
    </button>
    <button id="compare-traditional" class="btn-secondary-large">
      ⚡ Compare with Traditional Build
    </button>
  </div>

  <div id="build-progress" class="build-progress" style="display: none;">
    <h3>Build Progress</h3>
    <div class="progress-container">
      <div class="progress-bar">
        <div class="progress-fill" id="progress-fill"></div>
      </div>
      <div class="progress-text" id="progress-text">Initializing build...</div>
    </div>

    <div class="build-details">
      <div class="detail-section">
        <h4>Worker Assignment</h4>
        <div id="worker-assignment" class="worker-grid">
          <!-- Worker status will be populated here -->
        </div>
      </div>

      <div class="detail-section">
        <h4>Build Metrics</h4>
        <div id="build-metrics" class="metrics-grid">
          <div class="metric">
            <span class="metric-label">Tasks Completed:</span>
            <span class="metric-value" id="tasks-completed">0</span>
          </div>
          <div class="metric">
            <span class="metric-label">Cache Hit Rate:</span>
            <span class="metric-value" id="cache-hit-rate">0%</span>
          </div>
          <div class="metric">
            <span class="metric-label">Active Workers:</span>
            <span class="metric-value" id="active-workers">0</span>
          </div>
          <div class="metric">
            <span class="metric-label">Build Time:</span>
            <span class="metric-value" id="build-time">0s</span>
          </div>
        </div>
      </div>
    </div>
  </div>

  <div id="build-results" class="build-results" style="display: none;">
    <h3>Build Results</h3>
    <div class="results-summary">
      <div class="result-card success">
        <h4>✅ Distributed Build</h4>
        <div class="result-metrics">
          <div class="metric">⏱️ <strong id="distributed-time">-</strong></div>
          <div class="metric">🚀 <strong id="distributed-speedup">-</strong></div>
          <div class="metric">💾 <strong id="distributed-cache">-</strong></div>
        </div>
      </div>

      <div class="result-card comparison">
        <h4>⚡ Traditional Build</h4>
        <div class="result-metrics">
          <div class="metric">⏱️ <strong id="traditional-time">-</strong></div>
          <div class="metric">📊 <strong id="traditional-efficiency">-</strong></div>
        </div>
      </div>
    </div>

    <div class="performance-chart">
      <h4>Performance Comparison</h4>
      <canvas id="performance-chart" width="600" height="300"></canvas>
    </div>
  </div>
</div>

## 📁 Sample Project Details

### Simple Java Application
- **Structure**: Single module with source/main/java and source/test/java
- **Dependencies**: JUnit 5, Jackson, Apache Commons
- **Build Tasks**: compileJava, processResources, test, jar
- **Demo Focus**: Basic distributed compilation and testing

### Multi-Module Project
- **Structure**: 5 modules (core, api, web, data, test-utils)
- **Dependencies**: Spring Boot, Hibernate, PostgreSQL driver
- **Build Tasks**: Parallel module compilation, integration tests
- **Demo Focus**: Inter-module dependencies and parallel execution

### Spring Boot Application
- **Structure**: Full-stack web app with REST APIs
- **Dependencies**: Spring Boot, Spring Data JPA, Thymeleaf
- **Build Tasks**: BootJar, test, integrationTest
- **Demo Focus**: Enterprise application build optimization

### Android Application
- **Structure**: Mobile app with multiple build variants
- **Dependencies**: Android Gradle Plugin, support libraries
- **Build Tasks**: assembleDebug, assembleRelease, connectedTest
- **Demo Focus**: Complex build configurations and testing

## 🔧 Demo Environment Architecture

```
Demo Environment
├── Coordinator Node (1x)
│   ├── HTTP API Server (port 8080)
│   ├── RPC Server (port 8081)
│   └── Build Queue Manager
├── Worker Nodes (3x)
│   ├── Build Executors
│   ├── Cache Clients
│   └── Performance Monitors
├── Cache Server (1x)
│   ├── HTTP Cache API (port 8082)
│   └── Distributed Storage
└── Monitoring Stack
    ├── Metrics Collection
    ├── Real-time Dashboards
    └── Performance Analytics
```

## 📊 Live Metrics

<div class="live-metrics">
  <div class="metric-card">
    <h4>System Health</h4>
    <div id="system-health" class="health-status healthy">● Healthy</div>
  </div>

  <div class="metric-card">
    <h4>Active Builds</h4>
    <div id="active-builds-count" class="metric-value-large">0</div>
  </div>

  <div class="metric-card">
    <h4>Worker Pool</h4>
    <div id="worker-pool-status" class="metric-value-large">3/3</div>
  </div>

  <div class="metric-card">
    <h4>Cache Performance</h4>
    <div id="cache-performance" class="metric-value-large">95%</div>
  </div>
</div>

## 🎯 What You'll Experience

1. **Real-time Build Visualization**: Watch tasks distribute across workers
2. **Performance Comparison**: See speed improvements vs traditional builds
3. **Cache Effectiveness**: Observe cache hit rates and build acceleration
4. **Resource Utilization**: Monitor CPU, memory, and network usage
5. **Error Handling**: Experience graceful failure recovery
6. **Scaling Demonstration**: See how the system handles load

## 🚀 Getting Started

1. **Select a Project**: Choose from our sample projects above
2. **Start Demo Build**: Click "Start Distributed Build Demo"
3. **Watch Progress**: Observe real-time build progress and metrics
4. **Compare Results**: See performance comparison with traditional builds
5. **Explore Metrics**: Dive into detailed performance analytics

The demo environment is pre-configured and ready to use. All builds run in isolated containers with proper resource limits and monitoring.

<script src="/js/demo-environment.js"></script>