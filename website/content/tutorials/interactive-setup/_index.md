---
title: "Interactive Setup Tutorial"
weight: 15
---

# Interactive Setup Tutorial

Step-by-step guided setup of your distributed Gradle building environment.

## 🚀 Tutorial Overview

This interactive tutorial will guide you through setting up a complete distributed Gradle building environment from scratch. Each step includes:

- **Clear instructions** with code examples
- **Interactive validation** to check your progress
- **Troubleshooting tips** for common issues
- **Progress tracking** to see your advancement

### Prerequisites
- Linux/macOS/Windows with WSL
- Java 8+ installed
- Basic command line knowledge
- At least 2 machines (physical or virtual) with network connectivity

## 📋 Tutorial Steps

### Step 1: Environment Preparation
**Goal:** Verify system requirements and prepare machines

<div class="tutorial-step" id="step1">
  <div class="step-header">
    <span class="step-number">1</span>
    <h3>System Requirements Check</h3>
    <div class="step-status" id="step1-status">Not Started</div>
  </div>

  <div class="step-content">
    <p>First, let's verify your systems meet the requirements:</p>

    <div class="requirements-list">
      <div class="requirement">
        <strong>Hardware:</strong> 4GB RAM minimum per machine
        <button class="check-btn" data-check="ram">Check RAM</button>
        <span class="check-result" id="ram-check"></span>
      </div>

      <div class="requirement">
        <strong>Software:</strong> Java 8+ installed
        <button class="check-btn" data-check="java">Check Java</button>
        <span class="check-result" id="java-check"></span>
      </div>

      <div class="requirement">
        <strong>Network:</strong> SSH connectivity between machines
        <button class="check-btn" data-check="network">Check Network</button>
        <span class="check-result" id="network-check"></span>
      </div>
    </div>

    <div class="code-example">
      <h4>Run these commands on each machine:</h4>
      <pre><code># Check Java version
java -version

# Check available memory
free -h

# Check SSH service
systemctl status sshd</code></pre>
    </div>

    <div class="step-actions">
      <button class="btn-secondary" id="step1-validate">Validate Step 1</button>
      <div class="validation-result" id="step1-result"></div>
    </div>
  </div>
</div>

### Step 2: Download and Install
**Goal:** Get the distributed Gradle building system installed

<div class="tutorial-step" id="step2">
  <div class="step-header">
    <span class="step-number">2</span>
    <h3>Download and Installation</h3>
    <div class="step-status" id="step2-status">Locked</div>
  </div>

  <div class="step-content">
    <p>Download and set up the distributed Gradle building system:</p>

    <div class="code-example">
      <h4>Clone the repository:</h4>
      <pre><code>git clone https://github.com/milos85vasic/Distributed-Gradle-Building.git
cd Distributed-Gradle-Building

# Make scripts executable
chmod +x setup_*.sh sync_and_build.sh</code></pre>
    </div>

    <div class="file-verification">
      <h4>Verify installation:</h4>
      <div class="file-check">
        <span>✅ setup_master.sh exists and is executable</span>
        <span id="setup-master-check">Checking...</span>
      </div>
      <div class="file-check">
        <span>✅ setup_worker.sh exists and is executable</span>
        <span id="setup-worker-check">Checking...</span>
      </div>
      <div class="file-check">
        <span>✅ sync_and_build.sh exists and is executable</span>
        <span id="sync-build-check">Checking...</span>
      </div>
    </div>

    <div class="step-actions">
      <button class="btn-secondary" id="step2-validate" disabled>Validate Step 2</button>
      <div class="validation-result" id="step2-result"></div>
    </div>
  </div>
</div>

### Step 3: Configure SSH Access
**Goal:** Set up passwordless SSH between coordinator and workers

<div class="tutorial-step" id="step3">
  <div class="step-header">
    <span class="step-number">3</span>
    <h3>SSH Configuration</h3>
    <div class="step-status" id="step3-status">Locked</div>
  </div>

  <div class="step-content">
    <p>Configure SSH access for secure communication:</p>

    <div class="ssh-setup">
      <h4>Option 1: Automated Setup (Recommended)</h4>
      <div class="code-example">
        <pre><code>./setup_passwordless_ssh.sh</code></pre>
        <p>This script will:</p>
        <ul>
          <li>Generate SSH key pairs</li>
          <li>Copy public keys to worker machines</li>
          <li>Test bidirectional connectivity</li>
        </ul>
      </div>

      <h4>Option 2: Manual Setup</h4>
      <div class="code-example">
        <pre><code># Generate SSH key
ssh-keygen -t rsa -b 4096

# Copy to workers (repeat for each worker)
ssh-copy-id user@worker1.example.com
ssh-copy-id user@worker2.example.com

# Test connections
ssh worker1.example.com "echo 'SSH working'"
ssh worker2.example.com "echo 'SSH working'"</code></pre>
      </div>
    </div>

    <div class="ssh-verification">
      <h4>SSH Connectivity Test:</h4>
      <div class="worker-input">
        <label>Worker 1 hostname/IP:</label>
        <input type="text" id="worker1-host" placeholder="worker1.example.com">
      </div>
      <div class="worker-input">
        <label>Worker 2 hostname/IP:</label>
        <input type="text" id="worker2-host" placeholder="worker2.example.com">
      </div>
      <button class="check-btn" id="test-ssh">Test SSH Connections</button>
      <div class="ssh-result" id="ssh-test-result"></div>
    </div>

    <div class="step-actions">
      <button class="btn-secondary" id="step3-validate" disabled>Validate Step 3</button>
      <div class="validation-result" id="step3-result"></div>
    </div>
  </div>
</div>

### Step 4: Setup Coordinator
**Goal:** Configure the master/coordinator node

<div class="tutorial-step" id="step4">
  <div class="step-header">
    <span class="step-number">4</span>
    <h3>Coordinator Setup</h3>
    <div class="step-status" id="step4-status">Locked</div>
  </div>

  <div class="step-content">
    <p>Configure the coordinator node with your Gradle project:</p>

    <div class="project-setup">
      <h4>Choose your project path:</h4>
      <div class="project-input">
        <label>Gradle project path:</label>
        <input type="text" id="project-path" placeholder="/path/to/your/gradle/project">
        <button class="check-btn" id="browse-project">Browse</button>
      </div>

      <div class="project-validation">
        <div class="validation-item">
          <span>📁 build.gradle exists</span>
          <span id="build-gradle-check">Not checked</span>
        </div>
        <div class="validation-item">
          <span>📁 src/ directory exists</span>
          <span id="src-dir-check">Not checked</span>
        </div>
        <div class="validation-item">
          <span>⚙️ Valid Gradle project</span>
          <span id="gradle-project-check">Not checked</span>
        </div>
      </div>
    </div>

    <div class="code-example">
      <h4>Run coordinator setup:</h4>
      <pre><code>./setup_master.sh /path/to/your/gradle/project</code></pre>
      <p>This will configure:</p>
      <ul>
        <li>Worker discovery and registration</li>
        <li>Build synchronization directories</li>
        <li>Cache directories and permissions</li>
        <li>Coordinator configuration files</li>
      </ul>
    </div>

    <div class="coordinator-verification">
      <h4>Coordinator Status:</h4>
      <div class="status-item">
        <span>🔧 Coordinator service</span>
        <span id="coordinator-service-check">Not started</span>
      </div>
      <div class="status-item">
        <span>📊 Configuration files</span>
        <span id="coordinator-config-check">Not created</span>
      </div>
      <div class="status-item">
        <span>💾 Cache directories</span>
        <span id="coordinator-cache-check">Not created</span>
      </div>
    </div>

    <div class="step-actions">
      <button class="btn-secondary" id="step4-validate" disabled>Validate Step 4</button>
      <div class="validation-result" id="step4-result"></div>
    </div>
  </div>
</div>

### Step 5: Setup Workers
**Goal:** Configure worker nodes for distributed processing

<div class="tutorial-step" id="step5">
  <div class="step-header">
    <span class="step-number">5</span>
    <h3>Worker Setup</h3>
    <div class="step-status" id="step5-status">Locked</div>
  </div>

  <div class="step-content">
    <p>Configure worker nodes on each machine:</p>

    <div class="worker-setup">
      <h4>For each worker machine, run:</h4>
      <div class="code-example">
        <pre><code>./setup_worker.sh /path/to/your/gradle/project</code></pre>
        <p>This configures:</p>
        <ul>
          <li>Build environment and dependencies</li>
          <li>Cache directories and permissions</li>
          <li>Worker registration with coordinator</li>
          <li>Performance monitoring setup</li>
        </ul>
      </div>
    </div>

    <div class="worker-list">
      <h4>Worker Machines:</h4>
      <div class="worker-item">
        <input type="checkbox" id="worker1-setup" disabled>
        <label for="worker1-setup">Worker 1: <span id="worker1-host-display">-</span></label>
        <span class="worker-status" id="worker1-status">Not configured</span>
      </div>
      <div class="worker-item">
        <input type="checkbox" id="worker2-setup" disabled>
        <label for="worker2-setup">Worker 2: <span id="worker2-host-display">-</span></label>
        <span class="worker-status" id="worker2-status">Not configured</span>
      </div>
    </div>

    <div class="worker-verification">
      <h4>Worker Registration Check:</h4>
      <button class="check-btn" id="check-workers">Check Worker Registration</button>
      <div class="worker-result" id="worker-check-result"></div>
    </div>

    <div class="step-actions">
      <button class="btn-secondary" id="step5-validate" disabled>Validate Step 5</button>
      <div class="validation-result" id="step5-result"></div>
    </div>
  </div>
</div>

### Step 6: First Distributed Build
**Goal:** Run your first distributed Gradle build

<div class="tutorial-step" id="step6">
  <div class="step-header">
    <span class="step-number">6</span>
    <h3>First Distributed Build</h3>
    <div class="step-status" id="step6-status">Locked</div>
  </div>

  <div class="step-content">
    <p>Experience the power of distributed building:</p>

    <div class="build-comparison">
      <h4>Compare: Traditional vs Distributed</h4>

      <div class="build-option">
        <h5>Traditional Build (Single Machine)</h5>
        <div class="code-example">
          <pre><code>cd /path/to/your/project
time gradle clean build</code></pre>
        </div>
        <button class="btn-secondary" id="run-traditional">Run Traditional Build</button>
        <div class="build-result" id="traditional-result"></div>
      </div>

      <div class="build-option">
        <h5>Distributed Build (Multiple Machines)</h5>
        <div class="code-example">
          <pre><code>cd /path/to/distributed-gradle-building
./sync_and_build.sh clean build</code></pre>
        </div>
        <button class="btn-primary" id="run-distributed">🚀 Run Distributed Build</button>
        <div class="build-result" id="distributed-result"></div>
      </div>
    </div>

    <div class="performance-comparison">
      <h4>Performance Results:</h4>
      <div class="comparison-chart">
        <canvas id="build-comparison-chart" width="600" height="300"></canvas>
      </div>
    </div>

    <div class="step-actions">
      <button class="btn-secondary" id="step6-validate" disabled>Complete Tutorial</button>
      <div class="validation-result" id="step6-result"></div>
    </div>
  </div>
</div>

## 🎉 Tutorial Complete!

<div class="tutorial-completion" id="tutorial-complete" style="display: none;">
  <h2>🎉 Congratulations!</h2>
  <p>You've successfully set up a distributed Gradle building environment!</p>

  <div class="completion-summary">
    <h3>What you've accomplished:</h3>
    <ul>
      <li>✅ Verified system requirements</li>
      <li>✅ Downloaded and installed the system</li>
      <li>✅ Configured SSH access between machines</li>
      <li>✅ Set up coordinator and worker nodes</li>
      <li>✅ Ran your first distributed build</li>
    </ul>
  </div>

  <div class="next-steps">
    <h3>Next Steps:</h3>
    <ul>
      <li><a href="/docs/user-guide">📖 Read the User Guide</a></li>
      <li><a href="/docs/advanced-config">⚙️ Advanced Configuration</a></li>
      <li><a href="/video-courses">🎥 Watch Video Courses</a></li>
      <li><a href="/docs/monitoring">📊 Set up Monitoring</a></li>
      <li><a href="/community">💬 Join the Community</a></li>
    </ul>
  </div>

  <div class="certificate">
    <h3>🏆 Tutorial Certificate</h3>
    <p>You've earned the <strong>Distributed Gradle Setup Certificate</strong>!</p>
    <button class="btn-primary" id="download-certificate">Download Certificate</button>
  </div>
</div>

## 🔧 Troubleshooting

### Common Issues

<div class="troubleshooting">
  <div class="issue">
    <h4>SSH Connection Failed</h4>
    <p><strong>Solution:</strong> Check SSH service status, verify hostnames, ensure firewall allows SSH (port 22)</p>
  </div>

  <div class="issue">
    <h4>Worker Registration Failed</h4>
    <p><strong>Solution:</strong> Verify coordinator is running, check network connectivity, validate worker configuration</p>
  </div>

  <div class="issue">
    <h4>Build Synchronization Failed</h4>
    <p><strong>Solution:</strong> Check file permissions, verify rsync is installed, ensure sufficient disk space</p>
  </div>
</div>

<script src="/js/interactive-tutorial.js"></script>