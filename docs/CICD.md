# CI/CD Integration Guide

## 🔧 Integrating Distributed Builds with CI/CD Pipelines

### Implementation Choices for CI/CD

#### **Bash Implementation**
- Simple SSH-based integration
- Works with existing build agents
- Quick setup and deployment
- Suitable for small to medium teams

#### **Go Implementation** 
- RESTful API integration
- Advanced monitoring and metrics
- Scalable worker pool management
- Ideal for enterprise CI/CD

---

## 🚀 Jenkins Integration

### Jenkinsfile Configuration (Bash Implementation)

```groovy
pipeline {
    agent any
    
    environment {
        WORKER_IPS = 'worker1.company.com worker2.company.com worker3.company.com'
        PROJECT_DIR = "${WORKSPACE}"
        DISTRIBUTED_BUILD = 'true'
    }
    
    stages {
        stage('Setup') {
            steps {
                script {
                    // Clone distributed build system
                    git url: 'https://github.com/your-org/Distributed-Gradle-Building.git', 
                        branch: 'main',
                        dir: 'distributed-builder'
                    
                    // Setup distributed build environment
                    sh '''
                        cd distributed-builder
                        ./setup_master.sh $PROJECT_DIR
                    '''
                }
            }
        }
        
        stage('Distributed Build') {
            steps {
                script {
                    // Execute distributed build
                    sh '''
                        cd $PROJECT_DIR
                        $WORKSPACE/distributed-builder/sync_and_build.sh assemble
                    '''
                }
            }
        }
        
        stage('Distributed Test') {
            steps {
                script {
                    // Run distributed tests
                    sh '''
                        cd $PROJECT_DIR  
                        $WORKSPACE/distributed-builder/sync_and_build.sh test
                    '''
                }
            }
        }
        
        stage('Collect Artifacts') {
            steps {
                // Collect build artifacts from all workers
                archiveArtifacts artifacts: 'build/distributed/**/*', 
                                   allowEmptyArchive: true
                
                // Publish test results
                publishTestResults testResultsPattern: '**/build/test-results/**/*.xml'
                
                // Collect metrics
                archiveArtifacts artifacts: '.distributed/metrics/*.json',
                                   allowEmptyArchive: true
            }
        }
    }
    
    post {
        always {
            script {
                // Cleanup distributed build
                sh '''
                    cd $PROJECT_DIR
                    $WORKSPACE/distributed-builder/sync_and_build.sh clean
                '''
            }
        }
        
        success {
            // Build performance metrics
            script {
                def metrics = readJSON file: '.distributed/metrics/build_metrics.json'
                
                echo "Build Duration: ${metrics.metrics.build_duration_seconds}s"
                echo "Success Rate: ${metrics.metrics.success_rate}%"
                echo "Workers Used: ${metrics.workers.size()}"
            }
        }
        
        failure {
            // Diagnose build failures
            script {
                sh '''
                    # Collect diagnostic information
                    mkdir -p diagnostics
                    cp -r .distributed/logs/* diagnostics/
                    df -h > diagnostics/disk-usage.txt
                    free -h > diagnostics/memory-usage.txt
                    cat /proc/cpuinfo > diagnostics/cpu-info.txt
                '''
                archiveArtifacts artifacts: 'diagnostics/**/*'
            }
        }
    }
}
```

## 🏗️ Go API Integration (Enterprise)

### RESTful API Integration

The Go implementation provides full RESTful APIs for seamless CI/CD integration.

### Jenkins with Go API
```groovy
pipeline {
    agent any
    
    environment {
        BUILD_API_URL = 'http://build-coordinator.company.com:8080'
        METRICS_API_URL = 'http://build-coordinator.company.com:8082'
        PROJECT_DIR = "${WORKSPACE}"
    }
    
    stages {
        stage('Submit Build') {
            steps {
                script {
                    // Submit build via REST API
                    def response = sh(
                        script: """
                            curl -X POST ${env.BUILD_API_URL}/api/build \
                                -H 'Content-Type: application/json' \
                                -d '{"project_path": "${env.PROJECT_DIR}", "task_name": "assemble", "cache_enabled": true}'
                        """,
                        returnStdout: true
                    )
                    
                    def buildData = readJSON text: response
                    def buildId = buildData.build_id
                    
                    echo "Build submitted with ID: ${buildId}"
                    env.BUILD_ID = buildId
                }
            }
        }
        
        stage('Monitor Build') {
            steps {
                script {
                    // Poll build status
                    timeout(time: 30, unit: 'MINUTES') {
                        while (true) {
                            def statusResponse = sh(
                                script: "curl ${env.BUILD_API_URL}/api/builds/${env.BUILD_ID}",
                                returnStdout: true
                            )
                            
                            def status = readJSON text: statusResponse
                            
                            if (status.success) {
                                echo "Build completed successfully!"
                                break
                            } else if (status.error_message) {
                                error "Build failed: ${status.error_message}"
                            } else {
                                echo "Build in progress..."
                                sleep 30
                            }
                        }
                    }
                }
            }
        }
        
        stage('Collect Metrics') {
            steps {
                script {
                    // Get build metrics
                    def metricsResponse = sh(
                        script: "curl ${env.METRICS_API_URL}/api/metrics",
                        returnStdout: true
                    )
                    
                    def metrics = readJSON text: metricsResponse
                    
                    echo "Build Duration: ${metrics.build_duration_seconds}s"
                    echo "Workers Used: ${metrics.workers.size()}"
                    echo "Cache Hit Rate: ${metrics.cache_hit_rate}%"
                }
            }
        }
    }
}
```

### GitLab CI with Go API
```yaml
variables:
  BUILD_API_URL: "http://build-coordinator.company.com:8080"
  METRICS_API_URL: "http://build-coordinator.company.com:8082"

stages:
  - setup
  - build
  - metrics

submit_build:
  stage: build
  script:
    - |
      RESPONSE=$(curl -X POST $BUILD_API_URL/api/build \
        -H 'Content-Type: application/json' \
        -d '{"project_path": "$CI_PROJECT_DIR", "task_name": "assemble", "cache_enabled": true}')
      BUILD_ID=$(echo $RESPONSE | jq -r '.build_id')
      echo "BUILD_ID=$BUILD_ID" >> build.env
  artifacts:
    reports:
      dotenv: build.env
  environment:
    name: distributed-build/$CI_COMMIT_REF_SLUG
    url: $BUILD_API_URL/api/builds/$BUILD_ID

monitor_build:
  stage: metrics
  needs: [submit_build]
  dependencies:
    - submit_build
  script:
    - |
      timeout 1800 bash -c ""
        while true; do
          STATUS=$(curl -s $BUILD_API_URL/api/builds/$BUILD_ID)
          SUCCESS=$(echo $STATUS | jq -r '.success')
          
          if [ '$SUCCESS' = 'true' ]; then
            echo 'Build completed successfully!'
            break
          elif [ '$SUCCESS' = 'false' ]; then
            ERROR=$(echo $STATUS | jq -r '.error_message')
            echo "Build failed: $ERROR"
            exit 1
          else
            echo 'Build in progress...'
            sleep 30
          fi
        done
      ""
  artifacts:
    reports:
      junit: build-results.xml
    paths:
      - build/distributed/**/*
  coverage: '/Total coverage: (\d+\.\d+)%/'
```

### GitHub Actions with Go API
```yaml
name: Distributed Gradle Build

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  distributed-build:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Submit Build
      id: submit-build
      run: |
        RESPONSE=$(curl -X POST ${{ secrets.BUILD_API_URL }}/api/build \
          -H 'Content-Type: application/json' \
          -d '{"project_path": "${{ github.workspace }}", "task_name": "assemble", "cache_enabled": true}')
        
        BUILD_ID=$(echo $RESPONSE | jq -r '.build_id')
        echo "build-id=$BUILD_ID" >> $GITHUB_OUTPUT
        echo "Build submitted with ID: $BUILD_ID"
    
    - name: Monitor Build
      run: |
        BUILD_ID="${{ steps.submit-build.outputs.build-id }}"
        
        timeout 1800 bash -c ""
          while true; do
            STATUS=$(curl -s ${{ secrets.BUILD_API_URL }}/api/builds/$BUILD_ID)
            SUCCESS=$(echo $STATUS | jq -r '.success')
            
            if [ '$SUCCESS' = 'true' ]; then
              echo '✅ Build completed successfully!'
              break
            elif [ '$SUCCESS' = 'false' ]; then
              ERROR=$(echo $STATUS | jq -r '.error_message')
              echo "❌ Build failed: $ERROR"
              exit 1
            else
              echo '⏳ Build in progress...'
              sleep 30
            fi
          done
        ""
    
    - name: Collect Metrics
      if: success()
      run: |
        METRICS=$(curl -s ${{ secrets.METRICS_API_URL }}/api/metrics)
        BUILD_DURATION=$(echo $METRICS | jq -r '.build_duration_seconds')
        WORKERS_USED=$(echo $METRICS | jq -r '.workers | length')
        CACHE_HIT_RATE=$(echo $METRICS | jq -r '.cache_hit_rate')
        
        echo "📊 Build Metrics:"
        echo "⏱️ Duration: ${BUILD_DURATION}s"
        echo "👥 Workers: ${WORKERS_USED}"
        echo "💾 Cache Hit Rate: ${CACHE_HIT_RATE}%"
    
    - name: Upload Artifacts
      if: success()
      uses: actions/upload-artifact@v3
      with:
        name: build-artifacts
        path: |
          build/distributed/**/*
          .distributed/metrics/*.json
```

**📖 For complete Go API documentation:** [API Reference](API_REFERENCE.md)

```groovy
// vars/distributedGradle.groovy

def call(Map config = [:]) {
    def projectDir = config.get('projectDir', env.WORKSPACE)
    def task = config.get('task', 'assemble')
    def workerIps = config.get('workerIps', env.WORKER_IPS ?: '')
    
    // Setup distributed build
    sh """
        cd ${projectDir}
        git clone https://github.com/your-org/Distributed-Gradle-Building.git distributed-builder
        cd distributed-builder
        
        export WORKER_IPS="${workerIps}"
        ./setup_master.sh ${projectDir}
    """
    
    // Execute distributed build
    def buildResult = sh(returnStatus: true, script: """
        cd ${projectDir}
        ${projectDir}/distributed-builder/sync_and_build.sh ${task}
    """)
    
    if (buildResult != 0) {
        error "Distributed build failed with exit code: ${buildResult}"
    }
    
    // Return metrics
    def metrics = readJSON file: "${projectDir}/.distributed/metrics/build_metrics.json"
    return metrics
}
```

## 🔄 GitLab CI/CD Integration

### .gitlab-ci.yml Configuration

```yaml
stages:
  - setup
  - build
  - test
  - deploy

variables:
  DISTRIBUTED_BUILD: "true"
  WORKER_IPS: "worker1.internal worker2.internal worker3.internal"

# Distributed Build Setup
setup:distributed:
  stage: setup
  script:
    - echo "Setting up distributed build environment"
    - git clone https://github.com/your-org/Distributed-Gradle-Building.git
      distributed-builder
    - cd distributed-builder
    - ./setup_master.sh $CI_PROJECT_DIR
  artifacts:
    paths:
      - distributed-builder/
    expire_in: 1 hour
  cache:
    key: distributed-builder-${CI_COMMIT_REF_SLUG}
    paths:
      - distributed-builder/

# Distributed Build Stage
build:distributed:
  stage: build
  dependencies:
    - setup:distributed
  script:
    - echo "Running distributed build"
    - cd $CI_PROJECT_DIR
    - $CI_PROJECT_DIR/distributed-builder/sync_and_build.sh assemble
  artifacts:
    name: "$CI_JOB_NAME-$CI_COMMIT_REF_NAME"
    paths:
      - build/distributed/**/*
      - .distributed/metrics/*.json
    expire_in: 1 week
  coverage: '/Total coverage: (\d+\.\d+)%/'
  timeout: 30m

# Distributed Test Stage  
test:distributed:
  stage: test
  dependencies:
    - build:distributed
  script:
    - echo "Running distributed tests"
    - cd $CI_PROJECT_DIR
    - $CI_PROJECT_DIR/distributed-builder/sync_and_build.sh test
  artifacts:
    reports:
      junit: build/test-results/**/*.xml
    paths:
      - .distributed/logs/**/*
    expire_in: 1 week
  timeout: 45m

# Performance Metrics
metrics:performance:
  stage: deploy
  dependencies:
    - build:distributed
    - test:distributed
  script:
    - echo "Analyzing build performance"
    - cat .distributed/metrics/build_metrics.json | jq '.metrics'
  only:
    - main
    - master
```

## 🐙 GitHub Actions Integration

### .github/workflows/distributed-build.yml

```yaml
name: Distributed Gradle Build

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

env:
  WORKER_IPS: ${{ secrets.WORKER_IPS }}
  DISTRIBUTED_BUILD: true

jobs:
  setup:
    runs-on: ubuntu-latest
    outputs:
      setup-complete: ${{ steps.setup.outputs.complete }}
    
    steps:
    - name: Checkout source
      uses: actions/checkout@v3
      
    - name: Setup distributed build system
      id: setup
      run: |
        echo "Setting up distributed build environment"
        git clone https://github.com/your-org/Distributed-Gradle-Building.git distributed-builder
        cd distributed-builder
        ./setup_master.sh ${{ github.workspace }}
        echo "complete=true" >> $GITHUB_OUTPUT
        
  build:
    needs: setup
    runs-on: ubuntu-latest
    strategy:
      matrix:
        task: [assemble, test]
    
    steps:
    - name: Checkout source
      uses: actions/checkout@v3
      
    - name: Download distributed builder
      uses: actions/download-artifact@v3
      with:
        name: distributed-builder
        
    - name: Setup Java
      uses: actions/setup-java@v3
      with:
        java-version: '17'
        distribution: 'temurin'
        
    - name: Run distributed build
      run: |
        cd ${{ github.workspace }}
        chmod +x distributed-builder/sync_and_build.sh
        ./distributed-builder/sync_and_build.sh ${{ matrix.task }}
        
    - name: Upload build artifacts
      uses: actions/upload-artifact@v3
      with:
        name: build-${{ matrix.task }}
        path: |
          build/distributed/**/*
          .distributed/metrics/*.json
        retention-days: 7
        
    - name: Upload test results
      if: matrix.task == 'test'
      uses: actions/upload-artifact@v3
      with:
        name: test-results
        path: build/test-results/**/*.xml
        retention-days: 7

  performance:
    needs: [setup, build]
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    
    steps:
    - name: Download build artifacts
      uses: actions/download-artifact@v3
      
    - name: Analyze performance
      run: |
        echo "=== Build Performance Analysis ==="
        if [ -f "build-assemble/.distributed/metrics/build_metrics.json" ]; then
          cat build-assemble/.distributed/metrics/build_metrics.json | jq '.metrics'
        fi
        
        # Compare with baseline
        baseline_time=300  # 5 minutes baseline
        build_time=$(cat build-assemble/.distributed/metrics/build_metrics.json | jq -r '.metrics.build_duration_seconds')
        
        if (( build_time > baseline_time * 2 )); then
          echo "::warning::Build time ${build_time}s is 2x baseline of ${baseline_time}s"
        fi
        
    - name: Create performance report
      run: |
        cat > performance-report.md << EOF
        # Build Performance Report
        
        ## Metrics
        - Build Duration: $(cat build-assemble/.distributed/metrics/build_metrics.json | jq -r '.metrics.build_duration_seconds')s
        - Success Rate: $(cat build-assemble/.distributed/metrics/build_metrics.json | jq -r '.metrics.success_rate')%
        - Tasks Completed: $(cat build-assemble/.distributed/metrics/build_metrics.json | jq -r '.metrics.completed_tasks')
        
        ## Resource Utilization
        - CPU Usage: See detailed metrics in artifacts
        - Memory Usage: See detailed metrics in artifacts
        
        Generated on: $(date)
        EOF
        
    - name: Comment PR with performance
      if: github.event_name == 'pull_request'
      uses: actions/github-script@v6
      with:
        script: |
          const fs = require('fs');
          const report = fs.readFileSync('performance-report.md', 'utf8');
          
          github.rest.issues.createComment({
            issue_number: context.issue.number,
            owner: context.repo.owner,
            repo: context.repo.repo,
            body: report
          });
```

## 🔨 Azure DevOps Integration

### azure-pipelines.yml

```yaml
trigger:
- main
- develop

variables:
  DISTRIBUTED_BUILD: true
  WORKER_IPS: $(WORKER_IPS_SECRET)

pool:
  vmImage: 'ubuntu-latest'

stages:
- stage: Setup
  displayName: 'Setup Distributed Build'
  jobs:
  - job: Setup
    displayName: 'Setup Build Environment'
    steps:
    - checkout: self
    
    - task: Bash@3
      displayName: 'Setup Distributed Build'
      inputs:
        targetType: 'inline'
        script: |
          echo "Setting up distributed build environment"
          git clone https://github.com/your-org/Distributed-Gradle-Building.git distributed-builder
          cd distributed-builder
          ./setup_master.sh $(Build.SourcesDirectory)
          
    - publish: distributed-builder
      artifact: distributed-builder
      displayName: 'Publish Distributed Builder'

- stage: Build
  displayName: 'Distributed Build'
  dependsOn: Setup
  jobs:
  - job: Build
    displayName: 'Build Project'
    strategy:
      matrix:
        assemble:
          task: 'assemble'
        test:
          task: 'test'
    
    steps:
    - checkout: self
    
    - download: current
      artifact: distributed-builder
    
    - task: Bash@3
      displayName: 'Run Distributed Build'
      inputs:
        targetType: 'inline'
        script: |
          cd $(Build.SourcesDirectory)
          chmod +x distributed-builder/sync_and_build.sh
          ./distributed-builder/sync_and_build.sh $(task)
          
    - publish: 'build'
      artifact: 'build-$(task)'
      displayName: 'Publish Build Artifacts'
      
    - publish: 'metrics'
      artifact: 'metrics-$(task)'
      displayName: 'Publish Build Metrics'
      inputs:
        PathtoPublish: '.distributed/metrics'
        ArtifactName: 'metrics-$(task)'

- stage: Analysis
  displayName: 'Performance Analysis'
  dependsOn: Build
  condition: succeeded()
  jobs:
  - job: Analyze
    displayName: 'Analyze Build Performance'
    steps:
    - download: current
      artifact: metrics-assemble
      
    - task: Bash@3
      displayName: 'Analyze Performance'
      inputs:
        targetType: 'inline'
        script: |
          echo "=== Build Performance Analysis ==="
          if [ -f "metrics-assemble/build_metrics.json" ]; then
            cat metrics-assemble/build_metrics.json | jq '.metrics'
          fi
          
          # Log performance metrics for trends
          build_time=$(cat metrics-assemble/build_metrics.json | jq -r '.metrics.build_duration_seconds')
          echo "##vso[task.setvariable variable=BuildDuration;isOutput=true]$build_time"
          echo "##vso[metrics.upload]BuildDuration;$build_time;second"
```

## 📊 Performance Integration

### Build Performance Dashboard

```python
# ci_performance_dashboard.py
import json
import os
from datetime import datetime

def analyze_ci_performance():
    """Analyze CI/CD build performance over time"""
    
    metrics_dir = ".distributed/metrics"
    builds = []
    
    for metrics_file in os.listdir(metrics_dir):
        if metrics_file.startswith("build_") and metrics_file.endswith(".json"):
            with open(os.path.join(metrics_dir, metrics_file)) as f:
                metrics = json.load(f)
                builds.append({
                    'build_id': metrics['build_id'],
                    'duration': metrics['metrics']['build_duration_seconds'],
                    'success_rate': metrics['metrics']['success_rate'],
                    'workers': len(metrics.get('workers', [])),
                    'timestamp': metrics.get('start_time', 0)
                })
    
    # Sort by timestamp
    builds.sort(key=lambda x: x['timestamp'])
    
    # Generate performance report
    print("CI/CD Performance Report")
    print("=======================")
    
    if builds:
        avg_duration = sum(b['duration'] for b in builds) / len(builds)
        avg_success_rate = sum(b['success_rate'] for b in builds) / len(builds)
        
        print(f"Total Builds: {len(builds)}")
        print(f"Average Duration: {avg_duration:.1f}s")
        print(f"Average Success Rate: {avg_success_rate:.1f}%")
        print(f"Average Workers: {sum(b['workers'] for b in builds) / len(builds):.1f}")
        
        # Performance trend
        if len(builds) >= 2:
            recent_avg = sum(b['duration'] for b in builds[-5:]) / min(5, len(builds))
            older_avg = sum(b['duration'] for b in builds[:-5]) / max(1, len(builds) - 5)
            
            if recent_avg < older_avg:
                print("✅ Performance improving!")
            else:
                print("⚠️ Performance declining - investigate")
    
    return builds

if __name__ == "__main__":
    analyze_ci_performance()
```

## 🔧 Configuration Management

### Environment-Specific Configuration

```bash
# ci-environment-setup.sh

setup_ci_environment() {
    local ci_system="$1"
    
    case "$ci_system" in
        "jenkins")
            export WORKER_IPS="${JENKINS_WORKER_IPS:-}"
            export JENKINS_URL="${JENKINS_URL:-}"
            export BUILD_NUMBER="${BUILD_NUMBER:-}"
            export BUILD_URL="${BUILD_URL:-}"
            ;;
        "gitlab")
            export WORKER_IPS="${GITLAB_WORKER_IPS:-}"
            export CI_PROJECT_ID="${CI_PROJECT_ID:-}"
            export CI_COMMIT_SHA="${CI_COMMIT_SHA:-}"
            export CI_PIPELINE_ID="${CI_PIPELINE_ID:-}"
            ;;
        "github")
            export WORKER_IPS="${WORKER_IPS_SECRET:-}"
            export GITHUB_REPOSITORY="${GITHUB_REPOSITORY:-}"
            export GITHUB_SHA="${GITHUB_SHA:-}"
            export GITHUB_RUN_ID="${GITHUB_RUN_ID:-}"
            ;;
        "azure")
            export WORKER_IPS="$(WORKER_IPS_SECRET)"
            export BUILD_BUILDID="${BUILD_BUILDID:-}"
            export BUILD_REASON="${BUILD_REASON:-}"
            ;;
    esac
    
    # CI-specific optimizations
    export DISTRIBUTED_BUILD=true
    export METRICS_LEVEL=detailed
    export BUILD_CACHE_ENABLED=true
}
```

### Secrets Management

```bash
# ci-secrets-management.sh

setup_ci_secrets() {
    local ci_system="$1"
    
    case "$ci_system" in
        "jenkins")
            # Use Jenkins credentials store
            WORKER_IPS=$(echo "$WORKER_IPS_CRED" | base64 -d)
            ;;
        "gitlab")
            # Use GitLab CI/CD variables
            WORKER_IPS="$GITLAB_WORKER_IPS"
            ;;
        "github")
            # Use GitHub secrets
            WORKER_IPS="$WORKER_IPS_SECRET"
            ;;
        "azure")
            # Use Azure DevOps variable groups
            WORKER_IPS="$(WORKER_IPS_SECRET)"
            ;;
    esac
    
    # Export for distributed build system
    export WORKER_IPS
}
```

## 🚨 CI/CD Best Practices

### 1. Build Caching

```yaml
# Enable distributed build cache
variables:
  GRADLE_OPTS: "-Dgradle.user.home=.gradle-cache"
  BUILD_CACHE_ENABLED: "true"

cache:
  key: gradle-cache-${CI_COMMIT_REF_SLUG}
  paths:
    - .gradle-cache/
    - .distributed/cache/
```

### 2. Parallel Execution

```groovy
// Jenkins parallel stages
parallel {
    stage('Parallel Build') {
        parallel {
            stage('Build Module A') {
                steps {
                    sh './distributed-builder/sync_and_build.sh :module-a:assemble'
                }
            }
            stage('Build Module B') {
                steps {
                    sh './distributed-builder/sync_and_build.sh :module-b:assemble'
                }
            }
        }
    }
}
```

### 3. Failure Recovery

```bash
# ci-failure-recovery.sh

handle_build_failure() {
    local task="$1"
    local max_retries="${2:-3}"
    local retry_count=0
    
    while [ $retry_count -lt $max_retries ]; do
        echo "Attempting distributed build (attempt $((retry_count + 1))/$max_retries)"
        
        if ./distributed-builder/sync_and_build.sh "$task"; then
            echo "Build succeeded on attempt $((retry_count + 1))"
            return 0
        fi
        
        retry_count=$((retry_count + 1))
        
        # Wait before retry with exponential backoff
        sleep $((2 ** retry_count))
        
        # Check worker status before retry
        ./distributed-builder/distributed_gradle_build.sh --check-workers
    done
    
    echo "Build failed after $max_retries attempts"
    return 1
}
```

## 📈 Monitoring CI/CD Performance

### Build Performance Metrics

```bash
# ci-metrics-collection.sh

collect_ci_metrics() {
    local build_id="$1"
    local ci_system="$2"
    
    # Extract build metrics
    local build_metrics=".distributed/metrics/build_metrics.json"
    
    if [ -f "$build_metrics" ]; then
        # Add CI-specific metadata
        jq --arg build_id "$build_id" \
           --arg ci_system "$ci_system" \
           --arg timestamp "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
           '. + {"ci_metadata": {"build_id": $build_id, "ci_system": $ci_system, "timestamp": $timestamp}}' \
           "$build_metrics" > "${build_metrics}.tmp"
        mv "${build_metrics}.tmp" "$build_metrics"
        
        # Send to monitoring system
        if command -v curl >/dev/null 2>&1; then
            curl -X POST \
                 -H "Content-Type: application/json" \
                 -d @"$build_metrics" \
                 "${METRICS_ENDPOINT:-http://monitoring.company.com/metrics}"
        fi
    fi
}
```

---

## 🆕 Modern CI/CD Integration Templates

### Jenkins Pipeline (Declarative)

The project includes a complete `Jenkinsfile` for modern Jenkins pipelines with Docker support:

```groovy
// From project root: Jenkinsfile
pipeline {
    agent any

    environment {
        GO_VERSION = '1.21'
        DOCKER_IMAGE_TAG = "${env.BUILD_NUMBER}"
        REGISTRY = 'your-registry.com'
        APP_NAME = 'distributed-gradle-building'
    }

    stages {
        stage('Setup Go Environment') {
            steps {
                script {
                    // Install Go if not available
                    sh '''
                        if ! command -v go &> /dev/null; then
                            echo "Installing Go ${GO_VERSION}..."
                            wget -q https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz
                            sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
                            export PATH=$PATH:/usr/local/go/bin
                            echo "export PATH=$PATH:/usr/local/go/bin" >> ~/.bashrc
                        fi
                        go version
                    '''
                }
            }
        }

        stage('Build Binaries') {
            steps {
                dir('go') {
                    sh '''
                        mkdir -p bin
                        go build -ldflags="-X main.version=${BUILD_NUMBER}" -o bin/coordinator main.go
                        go build -ldflags="-X main.version=${BUILD_NUMBER}" -o bin/worker worker.go
                        go build -ldflags="-X main.version=${BUILD_NUMBER}" -o bin/cache_server cache_server.go
                        go build -ldflags="-X main.version=${BUILD_NUMBER}" -o bin/monitor monitor.go
                    '''
                }
            }
        }

        stage('Build Docker Images') {
            steps {
                script {
                    sh """
                        docker build -t ${REGISTRY}/${APP_NAME}-coordinator:${DOCKER_IMAGE_TAG} \\
                            --build-arg BINARY_PATH=go/bin/coordinator \\
                            -f docker/Dockerfile.coordinator .
                    """
                    // ... additional image builds
                }
            }
        }

        stage('Deploy to Production') {
            when {
                anyOf {
                    branch 'main'
                    branch 'master'
                    tag '*'
                }
            }
            steps {
                script {
                    timeout(time: 15, unit: 'MINUTES') {
                        input message: 'Deploy to production?', ok: 'Deploy'
                    }

                    sh '''
                        echo "Deploying to production environment..."
                        # Add your production deployment commands here
                    '''
                }
            }
        }
    }
}
```

### GitHub Actions Deployment

The project includes comprehensive GitHub Actions workflows:

#### Build & Test Workflow (`.github/workflows/test.yml`)
- Multi-version Go testing (1.19, 1.20, 1.21)
- Unit, integration, security, performance, and load tests
- Code coverage with Codecov integration
- SonarCloud analysis for main branch

#### Deployment Workflow (`.github/workflows/deploy.yml`)
```yaml
# From .github/workflows/deploy.yml
name: Deploy

on:
  push:
    branches: [ main, master ]
  release:
    types: [ published ]
  workflow_dispatch:

jobs:
  build-and-push:
    name: Build and Push Docker Images
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write

    steps:
      - name: Build and push coordinator image
        uses: docker/build-push-action@v5
        with:
          context: .
          file: docker/Dockerfile.coordinator
          push: true
          tags: ${{ steps.meta-coordinator.outputs.tags }}
          labels: ${{ steps.meta-coordinator.outputs.labels }}

  deploy-production:
    name: Deploy to Production
    needs: [build-and-push, deploy-staging]
    environment: production
    if: (github.ref == 'refs/heads/main' && github.event_name == 'push') || github.event.release.type == 'published'

    steps:
      - name: Deploy to production
        run: |
          echo "Deploying to production environment..."
          # Add your production deployment commands here
```

### GitLab CI/CD Pipeline

Complete `.gitlab-ci.yml` with Docker support:

```yaml
# From project root: .gitlab-ci.yml
stages:
  - test
  - build
  - deploy
  - performance

variables:
  DOCKER_DRIVER: overlay2
  GO_VERSION: "1.21"
  REGISTRY_IMAGE: $CI_REGISTRY_IMAGE
  DOCKER_IMAGE_TAG: $CI_COMMIT_REF_SLUG-$CI_COMMIT_SHORT_SHA

build_docker_images:
  stage: build
  image: docker:latest
  services:
    - docker:dind

  script:
    - docker build -t $REGISTRY_IMAGE/coordinator:$DOCKER_IMAGE_TAG -f docker/Dockerfile.coordinator .
    - docker build -t $REGISTRY_IMAGE/worker:$DOCKER_IMAGE_TAG -f docker/Dockerfile.worker .
    - docker build -t $REGISTRY_IMAGE/cache:$DOCKER_IMAGE_TAG -f docker/Dockerfile.cache .
    - docker build -t $REGISTRY_IMAGE/monitor:$DOCKER_IMAGE_TAG -f docker/Dockerfile.monitor .
    - docker push $REGISTRY_IMAGE/coordinator:$DOCKER_IMAGE_TAG
    # ... push other images

deploy_production:
  stage: deploy
  environment:
    name: production
    url: https://your-app.com
  when: manual
  only:
    - main
    - master
    - tags

  script:
    - echo "Deploying to production environment..."
    # Add your production deployment commands here
```

### Docker Containerization

#### Development Setup (`docker/docker-compose.yml`)
```yaml
# From docker/docker-compose.yml
version: '3.8'

services:
  coordinator:
    build:
      context: ..
      dockerfile: docker/Dockerfile.coordinator
    ports:
      - "8080:8080"  # HTTP API
      - "8081:8081"  # RPC port
    environment:
      - COORDINATOR_PORT=8080
      - COORDINATOR_RPC_PORT=8081
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/api/status"]
      interval: 30s
      timeout: 10s
      retries: 3

  worker-1:
    build:
      context: ..
      dockerfile: docker/Dockerfile.worker
    environment:
      - WORKER_ID=worker-1
      - COORDINATOR_HOST=coordinator
      - CACHE_HOST=cache
    depends_on:
      coordinator:
        condition: service_healthy
      cache:
        condition: service_healthy

  cache:
    build:
      context: ..
      dockerfile: docker/Dockerfile.cache
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8083/health"]

  monitor:
    build:
      context: ..
      dockerfile: docker/Dockerfile.monitor
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8084/health"]
```

#### Production Setup (`docker/docker-compose.prod.yml`)
```yaml
# From docker/docker-compose.prod.yml
version: '3.8'

services:
  coordinator:
    image: ${DOCKER_REGISTRY}/distributed-gradle-coordinator:${DOCKER_TAG:-latest}
    environment:
      - COORDINATOR_PORT=8080
      - LOG_LEVEL=info
      - GIN_MODE=release
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 1G

  workers:
    image: ${DOCKER_REGISTRY}/distributed-gradle-worker:${DOCKER_TAG:-latest}
    deploy:
      mode: replicated
      replicas: 3
      resources:
        limits:
          cpus: '2.0'
          memory: 4G
```

### Dockerfiles

#### Coordinator Dockerfile (`docker/Dockerfile.coordinator`)
```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go/go.mod go/go.sum ./
RUN go mod download
COPY go/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o coordinator main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata curl
COPY --from=builder /app/coordinator .
EXPOSE 8080 8081
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/api/status || exit 1
CMD ["./coordinator"]
```

#### Worker Dockerfile (`docker/Dockerfile.worker`)
```dockerfile
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache openjdk17-jdk
WORKDIR /app
COPY go/go.mod go/go.sum ./
RUN go mod download
COPY go/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o worker worker.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata curl openjdk17-jre-headless bash
COPY --from=builder /app/worker .
ENV JAVA_HOME=/usr/lib/jvm/java-17-openjdk
EXPOSE 8082-8090
CMD ["./worker", "/app/config/worker-config.json"]
```

### Quick Start with Docker

1. **Development Setup**:
   ```bash
   cd docker
   docker-compose up -d
   ```

2. **Production Deployment**:
   ```bash
   export DOCKER_REGISTRY=your-registry.com
   export DOCKER_TAG=v1.0.0
   cd docker
   docker-compose -f docker-compose.prod.yml up -d
   ```

3. **CI/CD Integration**:
   - Jenkins: Use the provided `Jenkinsfile`
   - GitHub Actions: Use `.github/workflows/deploy.yml`
   - GitLab CI: Use `.gitlab-ci.yml`

### Configuration Examples

#### Environment Variables
```bash
# Coordinator
COORDINATOR_PORT=8080
COORDINATOR_RPC_PORT=8081
LOG_LEVEL=info

# Worker
WORKER_ID=worker-1
WORKER_PORT=8082
COORDINATOR_HOST=coordinator
COORDINATOR_RPC_PORT=8081
MAX_BUILDS=5

# Cache
CACHE_PORT=8083
MAX_CACHE_SIZE=10GB

# Monitor
MONITOR_PORT=8084
```

#### Docker Compose Override
```yaml
# docker-compose.override.yml
version: '3.8'

services:
  coordinator:
    environment:
      - LOG_LEVEL=debug
    ports:
      - "8080:8080"

  worker-1:
    environment:
      - MAX_BUILDS=10
```

---

**Need CI/CD integration help?** Check the [Monitoring Guide](MONITORING.md) for detailed performance tracking.