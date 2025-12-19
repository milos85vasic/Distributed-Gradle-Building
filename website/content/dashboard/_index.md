---
title: "Analytics Dashboard"
weight: 70
---

# Analytics Dashboard

Real-time monitoring and performance analytics for your distributed Gradle building system.

## 📊 System Overview

<div id="dashboard-overview">
  <div class="metric-cards">
    <div class="metric-card">
      <div class="metric-icon">🏗️</div>
      <div class="metric-value" id="active-builds">0</div>
      <div class="metric-label">Active Builds</div>
      <div class="metric-trend" id="active-builds-trend">+0%</div>
    </div>

    <div class="metric-card">
      <div class="metric-icon">👥</div>
      <div class="metric-value" id="total-workers">0</div>
      <div class="metric-label">Total Workers</div>
      <div class="metric-trend" id="workers-trend">+0%</div>
    </div>

    <div class="metric-card">
      <div class="metric-icon">⚡</div>
      <div class="metric-value" id="avg-build-time">0s</div>
      <div class="metric-label">Avg Build Time</div>
      <div class="metric-trend" id="build-time-trend">+0%</div>
    </div>

    <div class="metric-card">
      <div class="metric-icon">💾</div>
      <div class="metric-value" id="cache-hit-rate">0%</div>
      <div class="metric-label">Cache Hit Rate</div>
      <div class="metric-trend" id="cache-trend">+0%</div>
    </div>

    <div class="metric-card">
      <div class="metric-icon">🚀</div>
      <div class="metric-value" id="performance-score">0</div>
      <div class="metric-label">Performance Score</div>
      <div class="metric-trend" id="performance-trend">+0%</div>
    </div>

    <div class="metric-card">
      <div class="metric-icon">🔥</div>
      <div class="metric-value" id="system-health">0%</div>
      <div class="metric-label">System Health</div>
      <div class="metric-trend" id="health-trend">+0%</div>
    </div>
  </div>
</div>

## 📈 Performance Charts

<div class="dashboard-charts">
  <div class="chart-container">
    <h3>Build Performance Over Time</h3>
    <div class="chart-controls">
      <select id="time-range">
        <option value="1h">Last Hour</option>
        <option value="24h" selected>Last 24 Hours</option>
        <option value="7d">Last 7 Days</option>
        <option value="30d">Last 30 Days</option>
      </select>
      <button id="refresh-charts" class="btn-secondary">🔄 Refresh</button>
    </div>
    <canvas id="performance-chart" width="800" height="300"></canvas>
  </div>

  <div class="chart-container">
    <h3>Resource Utilization</h3>
    <canvas id="resource-chart" width="800" height="300"></canvas>
  </div>

  <div class="chart-container">
    <h3>Build Distribution by Worker</h3>
    <canvas id="worker-distribution-chart" width="800" height="300"></canvas>
  </div>

  <div class="chart-container">
    <h3>Cache Performance Trends</h3>
    <canvas id="cache-trends-chart" width="800" height="300"></canvas>
  </div>
</div>

## 🔍 Detailed Analytics

<div class="analytics-sections">
  <div class="analytics-section">
    <h3>Recent Builds</h3>
    <div class="builds-table-container">
      <table id="recent-builds-table" class="data-table">
        <thead>
          <tr>
            <th>Build ID</th>
            <th>Project</th>
            <th>Status</th>
            <th>Duration</th>
            <th>Worker</th>
            <th>Cache Hit</th>
            <th>Started</th>
          </tr>
        </thead>
        <tbody id="recent-builds-body">
          <tr>
            <td colspan="7" class="loading">Loading recent builds...</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>

  <div class="analytics-section">
    <h3>Worker Performance</h3>
    <div class="workers-grid" id="workers-grid">
      <div class="worker-card loading">
        <div class="worker-header">
          <h4>Loading workers...</h4>
        </div>
      </div>
    </div>
  </div>

  <div class="analytics-section">
    <h3>System Alerts</h3>
    <div class="alerts-container" id="alerts-container">
      <div class="alert-item info">
        <div class="alert-icon">ℹ️</div>
        <div class="alert-content">
          <div class="alert-title">System Status</div>
          <div class="alert-message">All systems operational</div>
        </div>
        <div class="alert-time">Just now</div>
      </div>
    </div>
  </div>

  <div class="analytics-section">
    <h3>Build Queue Status</h3>
    <div class="queue-status">
      <div class="queue-metric">
        <div class="queue-number" id="queue-length">0</div>
        <div class="queue-label">Queued Builds</div>
      </div>
      <div class="queue-metric">
        <div class="queue-number" id="processing-builds">0</div>
        <div class="queue-label">Processing</div>
      </div>
      <div class="queue-metric">
        <div class="queue-number" id="completed-today">0</div>
        <div class="queue-label">Completed Today</div>
      </div>
    </div>
  </div>
</div>

## ⚙️ Dashboard Configuration

<div class="dashboard-config">
  <div class="config-section">
    <h3>API Endpoint</h3>
    <div class="config-input">
      <input type="text" id="api-endpoint-config" value="http://localhost:8080" placeholder="API Endpoint URL">
      <button id="test-endpoint" class="btn-secondary">Test Connection</button>
      <span id="endpoint-status" class="status-indicator">Not connected</span>
    </div>
  </div>

  <div class="config-section">
    <h3>Refresh Settings</h3>
    <div class="config-input">
      <label for="refresh-interval">Auto-refresh interval:</label>
      <select id="refresh-interval">
        <option value="5000">5 seconds</option>
        <option value="10000" selected>10 seconds</option>
        <option value="30000">30 seconds</option>
        <option value="60000">1 minute</option>
      </select>
      <button id="toggle-auto-refresh" class="btn-secondary">Pause Auto-refresh</button>
    </div>
  </div>

  <div class="config-section">
    <h3>Alert Thresholds</h3>
    <div class="thresholds-grid">
      <div class="threshold-item">
        <label for="cpu-threshold">CPU Usage Alert (%):</label>
        <input type="number" id="cpu-threshold" value="80" min="1" max="100">
      </div>
      <div class="threshold-item">
        <label for="memory-threshold">Memory Usage Alert (%):</label>
        <input type="number" id="memory-threshold" value="85" min="1" max="100">
      </div>
      <div class="threshold-item">
        <label for="queue-threshold">Queue Length Alert:</label>
        <input type="number" id="queue-threshold" value="10" min="1" max="100">
      </div>
      <div class="threshold-item">
        <label for="build-time-threshold">Build Time Alert (seconds):</label>
        <input type="number" id="build-time-threshold" value="300" min="10" max="3600">
      </div>
    </div>
  </div>
</div>

## 📋 Export & Reporting

<div class="export-section">
  <h3>Data Export</h3>
  <div class="export-options">
    <button id="export-json" class="btn-secondary">📄 Export as JSON</button>
    <button id="export-csv" class="btn-secondary">📊 Export as CSV</button>
    <button id="export-pdf" class="btn-secondary">📋 Generate PDF Report</button>
    <button id="schedule-report" class="btn-primary">📅 Schedule Daily Report</button>
  </div>

  <div class="report-config" id="report-config" style="display: none;">
    <h4>Daily Report Configuration</h4>
    <div class="report-settings">
      <div class="setting-item">
        <label for="report-email">Email address:</label>
        <input type="email" id="report-email" placeholder="your@email.com">
      </div>
      <div class="setting-item">
        <label for="report-time">Report time:</label>
        <input type="time" id="report-time" value="09:00">
      </div>
      <div class="setting-item">
        <label for="report-metrics">Include metrics:</label>
        <div class="checkbox-group">
          <label><input type="checkbox" checked> Build Performance</label>
          <label><input type="checkbox" checked> Resource Usage</label>
          <label><input type="checkbox" checked> Cache Statistics</label>
          <label><input type="checkbox"> Error Logs</label>
        </div>
      </div>
      <div class="setting-actions">
        <button id="save-report-config" class="btn-primary">Save Configuration</button>
        <button id="cancel-report-config" class="btn-secondary">Cancel</button>
      </div>
    </div>
  </div>
</div>

<script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
<script src="/js/analytics-dashboard.js"></script>