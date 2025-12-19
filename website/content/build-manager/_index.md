---
title: "Build Manager"
weight: 80
---

# Build Manager

Comprehensive web interface for managing distributed Gradle builds with real-time status updates and control capabilities.

## 🚀 Quick Build

<div id="build-manager">
  <div class="build-form-section">
    <h3>Submit New Build</h3>
    <form id="build-form">
      <div class="form-group">
        <label for="project-path">Project Path:</label>
        <div class="input-group">
          <input type="text" id="project-path" placeholder="/path/to/your/gradle/project" required>
          <button type="button" id="browse-project" class="btn-secondary">Browse</button>
        </div>
      </div>

      <div class="form-group">
        <label for="build-task">Build Task:</label>
        <select id="build-task" required>
          <option value="build">build (default)</option>
          <option value="clean build">clean build</option>
          <option value="test">test</option>
          <option value="assemble">assemble</option>
          <option value="check">check</option>
          <option value="custom">Custom task...</option>
        </select>
        <input type="text" id="custom-task" placeholder="Enter custom Gradle task" style="display: none;">
      </div>

      <div class="form-group">
        <label for="gradle-version">Gradle Version:</label>
        <select id="gradle-version">
          <option value="auto">Auto-detect</option>
          <option value="7.6">7.6</option>
          <option value="7.5">7.5</option>
          <option value="7.4">7.4</option>
          <option value="6.9">6.9</option>
        </select>
      </div>

      <div class="form-options">
        <label class="checkbox-label">
          <input type="checkbox" id="cache-enabled" checked>
          Enable build cache
        </label>
        <label class="checkbox-label">
          <input type="checkbox" id="parallel-build" checked>
          Enable parallel execution
        </label>
        <label class="checkbox-label">
          <input type="checkbox" id="offline-mode">
          Offline mode
        </label>
      </div>

      <div class="form-group">
        <label for="build-options">Additional Options:</label>
        <textarea id="build-options" rows="3" placeholder="-Dorg.gradle.jvmargs=-Xmx2g&#10;-Dorg.gradle.parallel=true&#10;-Dorg.gradle.workers.max=4"></textarea>
      </div>

      <div class="form-actions">
        <button type="submit" id="submit-build" class="btn-primary-large">🚀 Submit Build</button>
        <button type="button" id="validate-project" class="btn-secondary">Validate Project</button>
      </div>
    </form>
  </div>

  <div class="build-queue-section">
    <h3>Build Queue & Active Builds</h3>
    <div class="queue-controls">
      <div class="queue-stats">
        <span id="queue-count">0</span> queued,
        <span id="active-count">0</span> active,
        <span id="completed-count">0</span> completed today
      </div>
      <div class="queue-actions">
        <button id="refresh-queue" class="btn-secondary">🔄 Refresh</button>
        <button id="clear-completed" class="btn-secondary">🗑️ Clear Completed</button>
      </div>
    </div>

    <div id="build-queue" class="build-queue">
      <!-- Build items will be populated here -->
    </div>
  </div>
</div>

## 📊 Build History

<div class="build-history-section">
  <div class="history-controls">
    <div class="filter-controls">
      <select id="status-filter">
        <option value="all">All Statuses</option>
        <option value="completed">Completed</option>
        <option value="failed">Failed</option>
        <option value="running">Running</option>
        <option value="queued">Queued</option>
      </select>

      <select id="time-filter">
        <option value="1h">Last Hour</option>
        <option value="24h" selected>Last 24 Hours</option>
        <option value="7d">Last 7 Days</option>
        <option value="30d">Last 30 Days</option>
      </select>

      <input type="text" id="project-filter" placeholder="Filter by project...">
    </div>

    <div class="history-actions">
      <button id="export-history" class="btn-secondary">📄 Export</button>
      <button id="refresh-history" class="btn-secondary">🔄 Refresh</button>
    </div>
  </div>

  <div class="build-history-table">
    <table id="build-history-table" class="data-table">
      <thead>
        <tr>
          <th>Build ID</th>
          <th>Project</th>
          <th>Task</th>
          <th>Status</th>
          <th>Duration</th>
          <th>Worker</th>
          <th>Cache Hit</th>
          <th>Started</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody id="build-history-body">
        <tr>
          <td colspan="9" class="loading">Loading build history...</td>
        </tr>
      </tbody>
    </table>
  </div>

  <div class="pagination-controls">
    <button id="prev-page" class="btn-secondary" disabled>← Previous</button>
    <span id="page-info">Page 1 of 1</span>
    <button id="next-page" class="btn-secondary" disabled>Next →</button>
  </div>
</div>

## 👥 Worker Management

<div class="worker-management-section">
  <div class="worker-overview">
    <h3>Worker Pool Status</h3>
    <div class="worker-stats">
      <div class="stat-card">
        <div class="stat-value" id="total-workers">0</div>
        <div class="stat-label">Total Workers</div>
      </div>
      <div class="stat-card">
        <div class="stat-value" id="active-workers">0</div>
        <div class="stat-label">Active Workers</div>
      </div>
      <div class="stat-card">
        <div class="stat-value" id="idle-workers">0</div>
        <div class="stat-label">Idle Workers</div>
      </div>
      <div class="stat-card">
        <div class="stat-value" id="avg-utilization">0%</div>
        <div class="stat-label">Avg Utilization</div>
      </div>
    </div>
  </div>

  <div class="worker-grid" id="worker-grid">
    <!-- Worker cards will be populated here -->
  </div>

  <div class="worker-actions">
    <button id="add-worker" class="btn-primary">➕ Add Worker</button>
    <button id="refresh-workers" class="btn-secondary">🔄 Refresh</button>
    <button id="worker-settings" class="btn-secondary">⚙️ Settings</button>
  </div>
</div>

## ⚙️ Build Templates

<div class="build-templates-section">
  <h3>Build Templates</h3>
  <p>Save and reuse common build configurations</p>

  <div class="template-grid" id="template-grid">
    <div class="template-card" data-template="default">
      <h4>Default Build</h4>
      <p>Standard Gradle build with caching</p>
      <div class="template-actions">
        <button class="use-template btn-primary">Use</button>
        <button class="edit-template btn-secondary">Edit</button>
      </div>
    </div>

    <div class="template-card" data-template="ci-build">
      <h4>CI Build</h4>
      <p>Continuous integration build with tests</p>
      <div class="template-actions">
        <button class="use-template btn-primary">Use</button>
        <button class="edit-template btn-secondary">Edit</button>
      </div>
    </div>

    <div class="template-card" data-template="release-build">
      <h4>Release Build</h4>
      <p>Production release build</p>
      <div class="template-actions">
        <button class="use-template btn-primary">Use</button>
        <button class="edit-template btn-secondary">Edit</button>
      </div>
    </div>

    <div class="template-card create-new">
      <h4>➕ Create New</h4>
      <p>Save current configuration as template</p>
      <div class="template-actions">
        <button id="create-template" class="btn-secondary">Create</button>
      </div>
    </div>
  </div>
</div>

## 📋 Batch Operations

<div class="batch-operations-section">
  <h3>Batch Build Operations</h3>
  <p>Manage multiple builds and projects simultaneously</p>

  <div class="batch-controls">
    <div class="batch-input">
      <label for="batch-projects">Projects (one per line):</label>
      <textarea id="batch-projects" rows="5" placeholder="/path/to/project1&#10;/path/to/project2&#10;/path/to/project3"></textarea>
    </div>

    <div class="batch-settings">
      <div class="setting-group">
        <label for="batch-task">Task:</label>
        <select id="batch-task">
          <option value="build">build</option>
          <option value="test">test</option>
          <option value="clean build">clean build</option>
        </select>
      </div>

      <div class="setting-group">
        <label for="batch-concurrency">Max Concurrent Builds:</label>
        <input type="number" id="batch-concurrency" value="3" min="1" max="10">
      </div>

      <div class="setting-group">
        <label class="checkbox-label">
          <input type="checkbox" id="batch-sequential" checked>
          Run sequentially (wait for each to complete)
        </label>
      </div>
    </div>
  </div>

  <div class="batch-actions">
    <button id="start-batch" class="btn-primary-large">🚀 Start Batch Build</button>
    <button id="validate-batch" class="btn-secondary">✓ Validate Projects</button>
    <button id="clear-batch" class="btn-secondary">🗑️ Clear</button>
  </div>

  <div class="batch-progress" id="batch-progress" style="display: none;">
    <h4>Batch Progress</h4>
    <div class="progress-summary">
      <div class="progress-item">
        <span>Total Projects:</span>
        <span id="batch-total">0</span>
      </div>
      <div class="progress-item">
        <span>Completed:</span>
        <span id="batch-completed">0</span>
      </div>
      <div class="progress-item">
        <span>Running:</span>
        <span id="batch-running">0</span>
      </div>
      <div class="progress-item">
        <span>Failed:</span>
        <span id="batch-failed">0</span>
      </div>
    </div>

    <div class="batch-results" id="batch-results">
      <!-- Individual project results will be shown here -->
    </div>
  </div>
</div>

## 🔧 Advanced Controls

<div class="advanced-controls-section">
  <h3>Advanced Build Controls</h3>

  <div class="control-panels">
    <div class="control-panel">
      <h4>Cache Management</h4>
      <div class="control-actions">
        <button id="clear-cache" class="btn-secondary">🗑️ Clear Build Cache</button>
        <button id="warm-cache" class="btn-secondary">🔥 Warm Cache</button>
        <button id="cache-stats" class="btn-secondary">📊 Cache Statistics</button>
      </div>
    </div>

    <div class="control-panel">
      <h4>System Controls</h4>
      <div class="control-actions">
        <button id="restart-coordinator" class="btn-warning">🔄 Restart Coordinator</button>
        <button id="system-health-check" class="btn-secondary">🏥 Health Check</button>
        <button id="system-logs" class="btn-secondary">📋 View Logs</button>
      </div>
    </div>

    <div class="control-panel">
      <h4>Performance Tuning</h4>
      <div class="control-actions">
        <button id="optimize-workers" class="btn-secondary">⚡ Optimize Worker Pool</button>
        <button id="performance-report" class="btn-secondary">📈 Performance Report</button>
        <button id="resource-tuning" class="btn-secondary">🔧 Resource Tuning</button>
      </div>
    </div>
  </div>

  <div class="system-status">
    <h4>System Status</h4>
    <div class="status-indicators">
      <div class="status-item">
        <span class="status-label">Coordinator:</span>
        <span class="status-value" id="coordinator-status">Checking...</span>
      </div>
      <div class="status-item">
        <span class="status-label">Cache Server:</span>
        <span class="status-value" id="cache-status">Checking...</span>
      </div>
      <div class="status-item">
        <span class="status-label">Database:</span>
        <span class="status-value" id="database-status">Checking...</span>
      </div>
      <div class="status-item">
        <span class="status-label">Network:</span>
        <span class="status-value" id="network-status">Checking...</span>
      </div>
    </div>
  </div>
</div>

<script src="/js/build-manager.js"></script>