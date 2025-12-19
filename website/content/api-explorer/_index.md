---
title: "API Explorer"
weight: 50
---

# API Explorer

Interactive API documentation and testing interface for the Distributed Gradle Building System.

## 🚀 Live API Explorer

<div id="api-explorer">
  <div class="api-controls">
    <div class="control-group">
      <label for="api-endpoint">API Endpoint:</label>
      <input type="text" id="api-endpoint" value="http://localhost:8080" placeholder="http://localhost:8080">
    </div>
    <div class="control-group">
      <label for="auth-token">Auth Token (optional):</label>
      <input type="text" id="auth-token" placeholder="Your auth token">
    </div>
    <button id="test-connection" class="btn-primary">Test Connection</button>
  </div>

  <div id="connection-status" class="status-message"></div>

  <div class="api-endpoints">
    <h3>Available Endpoints</h3>

    <!-- Health Check -->
    <div class="endpoint-card">
      <h4>Health Check</h4>
      <div class="method">GET</div>
      <div class="endpoint-path">/health</div>
      <p>Check system health status</p>
      <button class="btn-secondary test-endpoint" data-endpoint="/health" data-method="GET">Test</button>
      <div class="response" id="response-health"></div>
    </div>

    <!-- System Status -->
    <div class="endpoint-card">
      <h4>System Status</h4>
      <div class="method">GET</div>
      <div class="endpoint-path">/api/status</div>
      <p>Get overall system status and metrics</p>
      <button class="btn-secondary test-endpoint" data-endpoint="/api/status" data-method="GET">Test</button>
      <div class="response" id="response-status"></div>
    </div>

    <!-- List Workers -->
    <div class="endpoint-card">
      <h4>List Workers</h4>
      <div class="method">GET</div>
      <div class="endpoint-path">/api/workers</div>
      <p>Get information about all registered workers</p>
      <button class="btn-secondary test-endpoint" data-endpoint="/api/workers" data-method="GET">Test</button>
      <div class="response" id="response-workers"></div>
    </div>

    <!-- Submit Build -->
    <div class="endpoint-card">
      <h4>Submit Build</h4>
      <div class="method">POST</div>
      <div class="endpoint-path">/api/build</div>
      <p>Submit a new distributed build request</p>
      <div class="request-body">
        <label>Build Request JSON:</label>
        <textarea id="build-request" rows="8" placeholder='{
  "project_path": "/path/to/gradle/project",
  "task_name": "build",
  "cache_enabled": true,
  "build_options": {
    "gradle_version": "7.6",
    "java_version": "11"
  }
}'></textarea>
      </div>
      <button class="btn-secondary test-endpoint" data-endpoint="/api/build" data-method="POST" data-body="build-request">Test</button>
      <div class="response" id="response-build"></div>
    </div>

    <!-- Get Build Status -->
    <div class="endpoint-card">
      <h4>Get Build Status</h4>
      <div class="method">GET</div>
      <div class="endpoint-path">/api/build/{build_id}</div>
      <p>Get status of a specific build</p>
      <div class="request-body">
        <label>Build ID:</label>
        <input type="text" id="build-id" placeholder="build-12345">
      </div>
      <button class="btn-secondary test-endpoint" data-endpoint="/api/build/" data-method="GET" data-dynamic="build-id">Test</button>
      <div class="response" id="response-build-status"></div>
    </div>

    <!-- Cache Status -->
    <div class="endpoint-card">
      <h4>Cache Status</h4>
      <div class="method">GET</div>
      <div class="endpoint-path">/api/cache/status</div>
      <p>Get cache server status and statistics</p>
      <button class="btn-secondary test-endpoint" data-endpoint="/api/cache/status" data-method="GET">Test</button>
      <div class="response" id="response-cache"></div>
    </div>
  </div>
</div>

## 📖 API Reference

### Authentication
Most endpoints require authentication via the `X-Auth-Token` header. Obtain a token through your system administrator.

### Response Format
All responses are returned in JSON format with appropriate HTTP status codes.

### Error Handling
Errors are returned with descriptive messages and appropriate HTTP status codes (4xx for client errors, 5xx for server errors).

## 🔧 Setup Instructions

1. **Start the Coordinator**: Run `./coordinator` to start the API server on port 8080
2. **Configure Endpoint**: Update the API Endpoint field above with your server URL
3. **Add Auth Token**: If required, enter your authentication token
4. **Test Connection**: Click "Test Connection" to verify connectivity
5. **Explore Endpoints**: Use the interactive buttons to test different API endpoints

## 📊 Response Examples

### Successful Build Submission
```json
{
  "build_id": "build-abc123",
  "status": "queued",
  "message": "Build submitted successfully"
}
```

### Build Status Response
```json
{
  "build_id": "build-abc123",
  "worker_id": "worker-01",
  "status": "completed",
  "success": true,
  "cache_hit_rate": 0.85,
  "artifacts": ["build/libs/app.jar", "build/reports/tests/test/index.html"]
}
```

### System Status Response
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "worker_count": 3,
  "queue_length": 2,
  "active_builds": 1,
  "cache_hit_rate": 0.78
}
```

<script src="/js/api-explorer.js"></script>