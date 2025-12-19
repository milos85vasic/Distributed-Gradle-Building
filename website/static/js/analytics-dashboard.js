// Analytics Dashboard JavaScript
document.addEventListener('DOMContentLoaded', function() {
    // Dashboard state
    let dashboardData = {
        apiEndpoint: 'http://localhost:8080',
        autoRefresh: true,
        refreshInterval: 10000,
        charts: {},
        lastUpdate: null
    };

    // Load saved configuration
    loadDashboardConfig();

    // Initialize dashboard
    initializeDashboard();

    // Set up event listeners
    setupEventListeners();

    // Start auto-refresh
    startAutoRefresh();

    function initializeDashboard() {
        // Initialize charts
        initializeCharts();

        // Load initial data
        refreshDashboard();

        // Set up real-time updates
        setupRealtimeUpdates();
    }

    function setupEventListeners() {
        // API endpoint configuration
        document.getElementById('test-endpoint').addEventListener('click', testEndpointConnection);
        document.getElementById('api-endpoint-config').addEventListener('change', updateApiEndpoint);

        // Refresh controls
        document.getElementById('refresh-charts').addEventListener('click', refreshDashboard);
        document.getElementById('toggle-auto-refresh').addEventListener('click', toggleAutoRefresh);
        document.getElementById('refresh-interval').addEventListener('change', updateRefreshInterval);

        // Time range selection
        document.getElementById('time-range').addEventListener('change', function() {
            refreshDashboard();
        });

        // Export functionality
        document.getElementById('export-json').addEventListener('click', () => exportData('json'));
        document.getElementById('export-csv').addEventListener('click', () => exportData('csv'));
        document.getElementById('export-pdf').addEventListener('click', () => exportData('pdf'));
        document.getElementById('schedule-report').addEventListener('click', toggleReportConfig);

        // Report configuration
        document.getElementById('save-report-config').addEventListener('click', saveReportConfig);
        document.getElementById('cancel-report-config').addEventListener('click', toggleReportConfig);

        // Threshold updates
        document.querySelectorAll('.thresholds-grid input').forEach(input => {
            input.addEventListener('change', updateThresholds);
        });
    }

    function initializeCharts() {
        const chartOptions = {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    position: 'top',
                }
            },
            scales: {
                x: {
                    display: true,
                    title: {
                        display: true,
                        text: 'Time'
                    }
                },
                y: {
                    display: true,
                    beginAtZero: true
                }
            }
        };

        // Performance chart
        const performanceCtx = document.getElementById('performance-chart').getContext('2d');
        dashboardData.charts.performance = new Chart(performanceCtx, {
            type: 'line',
            data: {
                labels: [],
                datasets: [{
                    label: 'Build Time (seconds)',
                    data: [],
                    borderColor: '#007bff',
                    backgroundColor: 'rgba(0, 123, 255, 0.1)',
                    tension: 0.4
                }, {
                    label: 'Cache Hit Rate (%)',
                    data: [],
                    borderColor: '#28a745',
                    backgroundColor: 'rgba(40, 167, 69, 0.1)',
                    yAxisID: 'y1',
                    tension: 0.4
                }]
            },
            options: {
                ...chartOptions,
                scales: {
                    ...chartOptions.scales,
                    y1: {
                        type: 'linear',
                        display: true,
                        position: 'right',
                        beginAtZero: true,
                        max: 100,
                        title: {
                            display: true,
                            text: 'Cache Hit Rate (%)'
                        }
                    }
                }
            }
        });

        // Resource utilization chart
        const resourceCtx = document.getElementById('resource-chart').getContext('2d');
        dashboardData.charts.resource = new Chart(resourceCtx, {
            type: 'line',
            data: {
                labels: [],
                datasets: [{
                    label: 'CPU Usage (%)',
                    data: [],
                    borderColor: '#dc3545',
                    backgroundColor: 'rgba(220, 53, 69, 0.1)',
                    tension: 0.4
                }, {
                    label: 'Memory Usage (%)',
                    data: [],
                    borderColor: '#ffc107',
                    backgroundColor: 'rgba(255, 193, 7, 0.1)',
                    tension: 0.4
                }, {
                    label: 'Disk I/O (%)',
                    data: [],
                    borderColor: '#6f42c1',
                    backgroundColor: 'rgba(111, 66, 193, 0.1)',
                    tension: 0.4
                }]
            },
            options: chartOptions
        });

        // Worker distribution chart
        const workerCtx = document.getElementById('worker-distribution-chart').getContext('2d');
        dashboardData.charts.workerDistribution = new Chart(workerCtx, {
            type: 'doughnut',
            data: {
                labels: [],
                datasets: [{
                    data: [],
                    backgroundColor: [
                        '#007bff',
                        '#28a745',
                        '#ffc107',
                        '#dc3545',
                        '#6f42c1',
                        '#17a2b8'
                    ]
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        position: 'right',
                    }
                }
            }
        });

        // Cache trends chart
        const cacheCtx = document.getElementById('cache-trends-chart').getContext('2d');
        dashboardData.charts.cacheTrends = new Chart(cacheCtx, {
            type: 'bar',
            data: {
                labels: [],
                datasets: [{
                    label: 'Cache Hits',
                    data: [],
                    backgroundColor: '#28a745'
                }, {
                    label: 'Cache Misses',
                    data: [],
                    backgroundColor: '#dc3545'
                }]
            },
            options: {
                ...chartOptions,
                scales: {
                    ...chartOptions.scales,
                    x: {
                        ...chartOptions.scales.x,
                        stacked: true
                    },
                    y: {
                        ...chartOptions.scales.y,
                        stacked: true
                    }
                }
            }
        });
    }

    async function refreshDashboard() {
        try {
            await Promise.all([
                updateOverviewMetrics(),
                updateCharts(),
                updateRecentBuilds(),
                updateWorkersGrid(),
                updateAlerts(),
                updateQueueStatus()
            ]);

            dashboardData.lastUpdate = new Date();
            updateLastRefreshTime();

        } catch (error) {
            console.error('Dashboard refresh error:', error);
            showConnectionError();
        }
    }

    async function updateOverviewMetrics() {
        try {
            const response = await fetch(`${dashboardData.apiEndpoint}/api/status`);
            if (!response.ok) throw new Error('Failed to fetch status');

            const status = await response.json();

            // Update metric cards
            updateMetricCard('active-builds', status.active_builds || 0);
            updateMetricCard('total-workers', status.worker_count || 0);
            updateMetricCard('avg-build-time', formatDuration(status.avg_build_time || 0));
            updateMetricCard('cache-hit-rate', `${Math.round((status.cache_hit_rate || 0) * 100)}%`);
            updateMetricCard('performance-score', calculatePerformanceScore(status));
            updateMetricCard('system-health', `${calculateSystemHealth(status)}%`);

        } catch (error) {
            console.error('Metrics update error:', error);
            // Set default values
            updateMetricCard('active-builds', 'N/A');
            updateMetricCard('total-workers', 'N/A');
            updateMetricCard('avg-build-time', 'N/A');
            updateMetricCard('cache-hit-rate', 'N/A');
            updateMetricCard('performance-score', 'N/A');
            updateMetricCard('system-health', 'N/A');
        }
    }

    function updateMetricCard(id, value) {
        const element = document.getElementById(id);
        if (element) {
            element.textContent = value;
        }
    }

    async function updateCharts() {
        const timeRange = document.getElementById('time-range').value;
        const hours = getHoursFromRange(timeRange);

        try {
            // Fetch historical data (mock data for demo)
            const historicalData = await fetchHistoricalData(hours);

            // Update performance chart
            dashboardData.charts.performance.data.labels = historicalData.labels;
            dashboardData.charts.performance.data.datasets[0].data = historicalData.buildTimes;
            dashboardData.charts.performance.data.datasets[1].data = historicalData.cacheRates;
            dashboardData.charts.performance.update();

            // Update resource chart
            dashboardData.charts.resource.data.labels = historicalData.labels;
            dashboardData.charts.resource.data.datasets[0].data = historicalData.cpuUsage;
            dashboardData.charts.resource.data.datasets[1].data = historicalData.memoryUsage;
            dashboardData.charts.resource.data.datasets[2].data = historicalData.diskUsage;
            dashboardData.charts.resource.update();

            // Update worker distribution
            const workerData = await fetchWorkerDistribution();
            dashboardData.charts.workerDistribution.data.labels = workerData.labels;
            dashboardData.charts.workerDistribution.data.datasets[0].data = workerData.data;
            dashboardData.charts.workerDistribution.update();

            // Update cache trends
            const cacheData = await fetchCacheTrends(hours);
            dashboardData.charts.cacheTrends.data.labels = cacheData.labels;
            dashboardData.charts.cacheTrends.data.datasets[0].data = cacheData.hits;
            dashboardData.charts.cacheTrends.data.datasets[1].data = cacheData.misses;
            dashboardData.charts.cacheTrends.update();

        } catch (error) {
            console.error('Charts update error:', error);
        }
    }

    async function updateRecentBuilds() {
        try {
            const response = await fetch(`${dashboardData.apiEndpoint}/api/builds/recent`);
            if (!response.ok) throw new Error('Failed to fetch builds');

            const builds = await response.json();
            const tbody = document.getElementById('recent-builds-body');

            if (builds.length === 0) {
                tbody.innerHTML = '<tr><td colspan="7">No recent builds found</td></tr>';
                return;
            }

            tbody.innerHTML = builds.map(build => `
                <tr>
                    <td>${build.build_id || 'N/A'}</td>
                    <td>${build.project_path || 'N/A'}</td>
                    <td><span class="status-${build.status || 'unknown'}">${build.status || 'Unknown'}</span></td>
                    <td>${formatDuration(build.duration || 0)}</td>
                    <td>${build.worker_id || 'N/A'}</td>
                    <td>${Math.round((build.cache_hit_rate || 0) * 100)}%</td>
                    <td>${formatTimestamp(build.start_time)}</td>
                </tr>
            `).join('');

        } catch (error) {
            console.error('Recent builds update error:', error);
            document.getElementById('recent-builds-body').innerHTML =
                '<tr><td colspan="7" class="error">Failed to load recent builds</td></tr>';
        }
    }

    async function updateWorkersGrid() {
        try {
            const response = await fetch(`${dashboardData.apiEndpoint}/api/workers`);
            if (!response.ok) throw new Error('Failed to fetch workers');

            const workers = await response.json();
            const grid = document.getElementById('workers-grid');

            if (workers.length === 0) {
                grid.innerHTML = '<div class="worker-card"><div class="worker-header"><h4>No workers registered</h4></div></div>';
                return;
            }

            grid.innerHTML = workers.map(worker => `
                <div class="worker-card status-${worker.status || 'unknown'}">
                    <div class="worker-header">
                        <h4>${worker.id}</h4>
                        <span class="worker-status">${worker.status || 'Unknown'}</span>
                    </div>
                    <div class="worker-metrics">
                        <div class="metric">
                            <span class="label">CPU:</span>
                            <span class="value">${Math.round(worker.cpu_usage || 0)}%</span>
                        </div>
                        <div class="metric">
                            <span class="label">Memory:</span>
                            <span class="value">${Math.round(worker.memory_usage || 0)}%</span>
                        </div>
                        <div class="metric">
                            <span class="label">Builds:</span>
                            <span class="value">${worker.build_count || 0}</span>
                        </div>
                        <div class="metric">
                            <span class="label">Avg Time:</span>
                            <span class="value">${formatDuration(worker.avg_build_time || 0)}</span>
                        </div>
                    </div>
                </div>
            `).join('');

        } catch (error) {
            console.error('Workers grid update error:', error);
            document.getElementById('workers-grid').innerHTML =
                '<div class="worker-card error"><div class="worker-header"><h4>Failed to load workers</h4></div></div>';
        }
    }

    async function updateAlerts() {
        try {
            const response = await fetch(`${dashboardData.apiEndpoint}/api/alerts`);
            if (!response.ok) throw new Error('Failed to fetch alerts');

            const alerts = await response.json();
            const container = document.getElementById('alerts-container');

            if (alerts.length === 0) {
                container.innerHTML = `
                    <div class="alert-item info">
                        <div class="alert-icon">ℹ️</div>
                        <div class="alert-content">
                            <div class="alert-title">System Status</div>
                            <div class="alert-message">All systems operational</div>
                        </div>
                        <div class="alert-time">Just now</div>
                    </div>
                `;
                return;
            }

            container.innerHTML = alerts.map(alert => `
                <div class="alert-item ${alert.severity || 'info'}">
                    <div class="alert-icon">${getAlertIcon(alert.severity)}</div>
                    <div class="alert-content">
                        <div class="alert-title">${alert.title}</div>
                        <div class="alert-message">${alert.message}</div>
                    </div>
                    <div class="alert-time">${formatTimestamp(alert.timestamp)}</div>
                </div>
            `).join('');

        } catch (error) {
            console.error('Alerts update error:', error);
        }
    }

    async function updateQueueStatus() {
        try {
            const response = await fetch(`${dashboardData.apiEndpoint}/api/queue/status`);
            if (!response.ok) throw new Error('Failed to fetch queue status');

            const queueStatus = await response.json();

            document.getElementById('queue-length').textContent = queueStatus.length || 0;
            document.getElementById('processing-builds').textContent = queueStatus.processing || 0;
            document.getElementById('completed-today').textContent = queueStatus.completed_today || 0;

        } catch (error) {
            console.error('Queue status update error:', error);
            document.getElementById('queue-length').textContent = 'N/A';
            document.getElementById('processing-builds').textContent = 'N/A';
            document.getElementById('completed-today').textContent = 'N/A';
        }
    }

    // Mock data functions (replace with real API calls)
    async function fetchHistoricalData(hours) {
        // Generate mock historical data
        const labels = [];
        const buildTimes = [];
        const cacheRates = [];
        const cpuUsage = [];
        const memoryUsage = [];
        const diskUsage = [];

        const now = new Date();
        for (let i = hours; i >= 0; i--) {
            const time = new Date(now.getTime() - i * 60 * 60 * 1000);
            labels.push(time.toLocaleTimeString());

            buildTimes.push(15 + Math.random() * 20); // 15-35 seconds
            cacheRates.push(60 + Math.random() * 30); // 60-90%
            cpuUsage.push(20 + Math.random() * 60); // 20-80%
            memoryUsage.push(30 + Math.random() * 50); // 30-80%
            diskUsage.push(10 + Math.random() * 40); // 10-50%
        }

        return { labels, buildTimes, cacheRates, cpuUsage, memoryUsage, diskUsage };
    }

    async function fetchWorkerDistribution() {
        // Mock worker distribution data
        return {
            labels: ['Worker-01', 'Worker-02', 'Worker-03', 'Worker-04'],
            data: [35, 28, 22, 15]
        };
    }

    async function fetchCacheTrends(hours) {
        // Mock cache trends data
        const labels = [];
        const hits = [];
        const misses = [];

        for (let i = 0; i < 24; i++) {
            labels.push(`${i}:00`);
            hits.push(Math.floor(Math.random() * 50) + 20);
            misses.push(Math.floor(Math.random() * 20) + 5);
        }

        return { labels, hits, misses };
    }

    // Utility functions
    function formatDuration(seconds) {
        if (seconds < 60) return `${Math.round(seconds)}s`;
        const minutes = Math.floor(seconds / 60);
        const remainingSeconds = Math.round(seconds % 60);
        return `${minutes}m ${remainingSeconds}s`;
    }

    function formatTimestamp(timestamp) {
        if (!timestamp) return 'N/A';
        return new Date(timestamp).toLocaleString();
    }

    function calculatePerformanceScore(status) {
        // Simple performance score calculation
        const cacheScore = (status.cache_hit_rate || 0) * 100;
        const workerScore = Math.min((status.worker_count || 0) * 20, 100);
        const healthScore = calculateSystemHealth(status);

        return Math.round((cacheScore + workerScore + healthScore) / 3);
    }

    function calculateSystemHealth(status) {
        // Simple health score based on various metrics
        let health = 100;

        if ((status.cpu_usage || 0) > 80) health -= 20;
        if ((status.memory_usage || 0) > 85) health -= 20;
        if ((status.queue_length || 0) > 10) health -= 15;
        if ((status.error_rate || 0) > 0.05) health -= 15;

        return Math.max(0, health);
    }

    function getAlertIcon(severity) {
        switch (severity) {
            case 'error': return '🚨';
            case 'warning': return '⚠️';
            case 'info': return 'ℹ️';
            default: return '📢';
        }
    }

    function getHoursFromRange(range) {
        switch (range) {
            case '1h': return 1;
            case '24h': return 24;
            case '7d': return 168;
            case '30d': return 720;
            default: return 24;
        }
    }

    function updateLastRefreshTime() {
        // Could add a "Last updated" indicator
    }

    function showConnectionError() {
        // Show connection error indicators
        document.getElementById('endpoint-status').textContent = 'Connection failed';
        document.getElementById('endpoint-status').className = 'status-indicator error';
    }

    // Configuration functions
    function loadDashboardConfig() {
        const saved = localStorage.getItem('dashboard-config');
        if (saved) {
            const config = JSON.parse(saved);
            dashboardData.apiEndpoint = config.apiEndpoint || dashboardData.apiEndpoint;
            dashboardData.refreshInterval = config.refreshInterval || dashboardData.refreshInterval;
            dashboardData.autoRefresh = config.autoRefresh !== false;

            document.getElementById('api-endpoint-config').value = dashboardData.apiEndpoint;
            document.getElementById('refresh-interval').value = dashboardData.refreshInterval;
        }
    }

    function saveDashboardConfig() {
        const config = {
            apiEndpoint: dashboardData.apiEndpoint,
            refreshInterval: dashboardData.refreshInterval,
            autoRefresh: dashboardData.autoRefresh
        };
        localStorage.setItem('dashboard-config', JSON.stringify(config));
    }

    function updateApiEndpoint() {
        dashboardData.apiEndpoint = document.getElementById('api-endpoint-config').value;
        saveDashboardConfig();
    }

    function updateRefreshInterval() {
        dashboardData.refreshInterval = parseInt(document.getElementById('refresh-interval').value);
        saveDashboardConfig();
        restartAutoRefresh();
    }

    function toggleAutoRefresh() {
        dashboardData.autoRefresh = !dashboardData.autoRefresh;
        const button = document.getElementById('toggle-auto-refresh');
        button.textContent = dashboardData.autoRefresh ? 'Pause Auto-refresh' : 'Resume Auto-refresh';
        saveDashboardConfig();

        if (dashboardData.autoRefresh) {
            startAutoRefresh();
        } else {
            stopAutoRefresh();
        }
    }

    function startAutoRefresh() {
        if (dashboardData.autoRefresh && !dashboardData.refreshTimer) {
            dashboardData.refreshTimer = setInterval(refreshDashboard, dashboardData.refreshInterval);
        }
    }

    function stopAutoRefresh() {
        if (dashboardData.refreshTimer) {
            clearInterval(dashboardData.refreshTimer);
            dashboardData.refreshTimer = null;
        }
    }

    function restartAutoRefresh() {
        stopAutoRefresh();
        startAutoRefresh();
    }

    async function testEndpointConnection() {
        const endpoint = document.getElementById('api-endpoint-config').value;
        const statusElement = document.getElementById('endpoint-status');

        statusElement.textContent = 'Testing...';
        statusElement.className = 'status-indicator testing';

        try {
            const response = await fetch(`${endpoint}/health`, { timeout: 5000 });
            if (response.ok) {
                statusElement.textContent = 'Connected';
                statusElement.className = 'status-indicator success';
            } else {
                throw new Error(`HTTP ${response.status}`);
            }
        } catch (error) {
            statusElement.textContent = 'Failed';
            statusElement.className = 'status-indicator error';
        }
    }

    function updateThresholds() {
        // Save threshold values to localStorage or send to server
        const thresholds = {
            cpu: document.getElementById('cpu-threshold').value,
            memory: document.getElementById('memory-threshold').value,
            queue: document.getElementById('queue-threshold').value,
            buildTime: document.getElementById('build-time-threshold').value
        };

        localStorage.setItem('dashboard-thresholds', JSON.stringify(thresholds));
    }

    function exportData(format) {
        // Mock export functionality
        const data = {
            timestamp: new Date().toISOString(),
            metrics: {
                activeBuilds: document.getElementById('active-builds').textContent,
                totalWorkers: document.getElementById('total-workers').textContent,
                avgBuildTime: document.getElementById('avg-build-time').textContent,
                cacheHitRate: document.getElementById('cache-hit-rate').textContent
            }
        };

        switch (format) {
            case 'json':
                downloadFile(JSON.stringify(data, null, 2), 'dashboard-data.json', 'application/json');
                break;
            case 'csv':
                const csv = convertToCSV(data);
                downloadFile(csv, 'dashboard-data.csv', 'text/csv');
                break;
            case 'pdf':
                alert('PDF export would be implemented here');
                break;
        }
    }

    function downloadFile(content, filename, mimeType) {
        const blob = new Blob([content], { type: mimeType });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = filename;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
    }

    function convertToCSV(data) {
        // Simple CSV conversion
        return `Timestamp,Active Builds,Total Workers,Avg Build Time,Cache Hit Rate\n${data.timestamp},${data.metrics.activeBuilds},${data.metrics.totalWorkers},"${data.metrics.avgBuildTime}","${data.metrics.cacheHitRate}"`;
    }

    function toggleReportConfig() {
        const configDiv = document.getElementById('report-config');
        configDiv.style.display = configDiv.style.display === 'none' ? 'block' : 'none';
    }

    function saveReportConfig() {
        const config = {
            email: document.getElementById('report-email').value,
            time: document.getElementById('report-time').value,
            metrics: Array.from(document.querySelectorAll('#report-config input[type="checkbox"]:checked')).map(cb => cb.parentElement.textContent.trim())
        };

        localStorage.setItem('report-config', JSON.stringify(config));
        alert('Report configuration saved!');
        toggleReportConfig();
    }

    function setupRealtimeUpdates() {
        // Could implement WebSocket or Server-Sent Events for real-time updates
        // For now, just rely on polling
    }
});

// Add CSS styles for the analytics dashboard
const style = document.createElement('style');
style.textContent = `
    #dashboard-overview {
        margin-bottom: 30px;
    }

    .metric-cards {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
        gap: 20px;
        margin-bottom: 30px;
    }

    .metric-card {
        background: white;
        border: 1px solid #dee2e6;
        border-radius: 8px;
        padding: 20px;
        text-align: center;
        box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        transition: transform 0.2s ease;
    }

    .metric-card:hover {
        transform: translateY(-2px);
        box-shadow: 0 4px 8px rgba(0,0,0,0.15);
    }

    .metric-icon {
        font-size: 24px;
        margin-bottom: 10px;
    }

    .metric-value {
        font-size: 32px;
        font-weight: bold;
        color: #007bff;
        margin: 10px 0;
    }

    .metric-value-large {
        font-size: 48px;
        font-weight: bold;
        color: #007bff;
        margin: 10px 0;
    }

    .metric-label {
        color: #6c757d;
        font-size: 14px;
        font-weight: 500;
    }

    .metric-trend {
        font-size: 12px;
        font-weight: bold;
        margin-top: 5px;
    }

    .metric-trend.positive {
        color: #28a745;
    }

    .metric-trend.negative {
        color: #dc3545;
    }

    .dashboard-charts {
        display: grid;
        grid-template-columns: 1fr;
        gap: 30px;
        margin-bottom: 40px;
    }

    .chart-container {
        background: white;
        border: 1px solid #dee2e6;
        border-radius: 8px;
        padding: 20px;
        box-shadow: 0 2px 4px rgba(0,0,0,0.1);
    }

    .chart-container h3 {
        margin: 0 0 20px 0;
        color: #495057;
    }

    .chart-controls {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 20px;
    }

    .chart-controls select {
        padding: 8px 12px;
        border: 1px solid #ced4da;
        border-radius: 4px;
        font-size: 14px;
    }

    .analytics-sections {
        display: grid;
        grid-template-columns: 1fr;
        gap: 30px;
        margin-bottom: 40px;
    }

    .analytics-section {
        background: white;
        border: 1px solid #dee2e6;
        border-radius: 8px;
        padding: 20px;
        box-shadow: 0 2px 4px rgba(0,0,0,0.1);
    }

    .analytics-section h3 {
        margin: 0 0 20px 0;
        color: #495057;
    }

    .data-table {
        width: 100%;
        border-collapse: collapse;
        font-size: 14px;
    }

    .data-table th, .data-table td {
        padding: 12px 8px;
        text-align: left;
        border-bottom: 1px solid #dee2e6;
    }

    .data-table th {
        background: #f8f9fa;
        font-weight: 600;
        color: #495057;
    }

    .data-table tr:hover {
        background: #f8f9fa;
    }

    .status-completed, .status-success {
        color: #28a745;
        font-weight: bold;
    }

    .status-failed, .status-error {
        color: #dc3545;
        font-weight: bold;
    }

    .status-running, .status-processing {
        color: #ffc107;
        font-weight: bold;
    }

    .workers-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
        gap: 15px;
    }

    .worker-card {
        border: 1px solid #dee2e6;
        border-radius: 6px;
        padding: 15px;
        background: white;
    }

    .worker-card.status-idle {
        border-color: #6c757d;
    }

    .worker-card.status-busy {
        border-color: #ffc107;
        background: #fff3cd;
    }

    .worker-card.status-error {
        border-color: #dc3545;
        background: #f8d7da;
    }

    .worker-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 10px;
    }

    .worker-header h4 {
        margin: 0;
        color: #495057;
    }

    .worker-status {
        padding: 3px 8px;
        border-radius: 4px;
        font-size: 12px;
        font-weight: bold;
    }

    .worker-metrics {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 8px;
    }

    .worker-metrics .metric {
        display: flex;
        justify-content: space-between;
        font-size: 13px;
    }

    .alerts-container {
        display: flex;
        flex-direction: column;
        gap: 10px;
    }

    .alert-item {
        display: flex;
        align-items: center;
        padding: 12px;
        border-radius: 6px;
        border-left: 4px solid;
    }

    .alert-item.error {
        background: #f8d7da;
        border-left-color: #dc3545;
    }

    .alert-item.warning {
        background: #fff3cd;
        border-left-color: #ffc107;
    }

    .alert-item.info {
        background: #d1ecf1;
        border-left-color: #17a2b8;
    }

    .alert-icon {
        font-size: 18px;
        margin-right: 12px;
    }

    .alert-content {
        flex: 1;
    }

    .alert-title {
        font-weight: 600;
        margin-bottom: 4px;
    }

    .alert-message {
        font-size: 14px;
        color: #6c757d;
    }

    .alert-time {
        font-size: 12px;
        color: #6c757d;
    }

    .queue-status {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
        gap: 20px;
    }

    .queue-metric {
        text-align: center;
        padding: 20px;
        background: #f8f9fa;
        border-radius: 8px;
    }

    .queue-number {
        font-size: 36px;
        font-weight: bold;
        color: #007bff;
        display: block;
        margin-bottom: 5px;
    }

    .queue-label {
        color: #6c757d;
        font-size: 14px;
    }

    .dashboard-config {
        background: white;
        border: 1px solid #dee2e6;
        border-radius: 8px;
        padding: 20px;
        margin-bottom: 30px;
        box-shadow: 0 2px 4px rgba(0,0,0,0.1);
    }

    .config-section {
        margin-bottom: 20px;
    }

    .config-section h3 {
        margin: 0 0 15px 0;
        color: #495057;
    }

    .config-input {
        display: flex;
        align-items: center;
        gap: 10px;
    }

    .config-input input {
        flex: 1;
        padding: 8px 12px;
        border: 1px solid #ced4da;
        border-radius: 4px;
        font-size: 14px;
    }

    .status-indicator {
        font-weight: bold;
        padding: 4px 8px;
        border-radius: 4px;
        font-size: 12px;
    }

    .status-indicator.success {
        background: #d4edda;
        color: #155724;
    }

    .status-indicator.error {
        background: #f8d7da;
        color: #721c24;
    }

    .status-indicator.testing {
        background: #fff3cd;
        color: #856404;
    }

    .thresholds-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
        gap: 15px;
    }

    .threshold-item {
        display: flex;
        flex-direction: column;
        gap: 5px;
    }

    .threshold-item label {
        font-weight: 500;
        color: #495057;
        font-size: 14px;
    }

    .threshold-item input {
        padding: 8px;
        border: 1px solid #ced4da;
        border-radius: 4px;
        font-size: 14px;
    }

    .export-section {
        background: white;
        border: 1px solid #dee2e6;
        border-radius: 8px;
        padding: 20px;
        box-shadow: 0 2px 4px rgba(0,0,0,0.1);
    }

    .export-options {
        display: flex;
        gap: 10px;
        margin-bottom: 20px;
        flex-wrap: wrap;
    }

    .report-config {
        background: #f8f9fa;
        border: 1px solid #dee2e6;
        border-radius: 6px;
        padding: 20px;
        margin-top: 15px;
    }

    .report-settings {
        display: grid;
        gap: 15px;
    }

    .setting-item {
        display: flex;
        flex-direction: column;
        gap: 5px;
    }

    .setting-item label {
        font-weight: 500;
        color: #495057;
        font-size: 14px;
    }

    .setting-item input[type="email"],
    .setting-item input[type="time"] {
        padding: 8px;
        border: 1px solid #ced4da;
        border-radius: 4px;
        font-size: 14px;
    }

    .checkbox-group {
        display: flex;
        flex-wrap: wrap;
        gap: 15px;
    }

    .checkbox-group label {
        display: flex;
        align-items: center;
        gap: 5px;
        font-size: 14px;
        color: #495057;
    }

    .setting-actions {
        display: flex;
        gap: 10px;
        margin-top: 15px;
    }

    .btn-primary, .btn-secondary {
        background: #007bff;
        color: white;
        border: none;
        padding: 10px 20px;
        border-radius: 4px;
        cursor: pointer;
        font-size: 14px;
        font-weight: 600;
        text-decoration: none;
        display: inline-block;
        transition: background-color 0.2s ease;
    }

    .btn-secondary {
        background: #6c757d;
    }

    .btn-primary:hover, .btn-secondary:hover {
        opacity: 0.9;
    }

    .loading {
        text-align: center;
        padding: 20px;
        color: #6c757d;
        font-style: italic;
    }

    .error {
        color: #dc3545;
        font-weight: 500;
    }

    @media (max-width: 768px) {
        .metric-cards {
            grid-template-columns: 1fr;
        }

        .dashboard-charts {
            gap: 20px;
        }

        .chart-controls {
            flex-direction: column;
            align-items: stretch;
            gap: 10px;
        }

        .analytics-sections {
            gap: 20px;
        }

        .workers-grid {
            grid-template-columns: 1fr;
        }

        .queue-status {
            grid-template-columns: 1fr;
        }

        .config-input {
            flex-direction: column;
            align-items: stretch;
        }

        .thresholds-grid {
            grid-template-columns: 1fr;
        }

        .export-options {
            flex-direction: column;
        }

        .checkbox-group {
            flex-direction: column;
        }

        .setting-actions {
            flex-direction: column;
        }
    }
`;
document.head.appendChild(style);