// Build Manager JavaScript
document.addEventListener('DOMContentLoaded', function() {
    // Build Manager state
    let buildManager = {
        apiEndpoint: 'http://localhost:8080',
        currentBuilds: [],
        buildHistory: [],
        workers: [],
        templates: [],
        batchBuilds: [],
        autoRefresh: true,
        refreshInterval: 5000
    };

    // Initialize build manager
    initializeBuildManager();

    // Set up event listeners
    setupEventListeners();

    // Start auto-refresh
    startAutoRefresh();

    function initializeBuildManager() {
        loadSavedSettings();
        setupFormValidation();
        refreshAllData();
    }

    function setupEventListeners() {
        // Build form events
        document.getElementById('build-form').addEventListener('submit', handleBuildSubmit);
        document.getElementById('build-task').addEventListener('change', handleTaskChange);
        document.getElementById('validate-project').addEventListener('click', validateProject);
        document.getElementById('browse-project').addEventListener('click', browseProject);

        // Queue management
        document.getElementById('refresh-queue').addEventListener('click', refreshQueue);
        document.getElementById('clear-completed').addEventListener('click', clearCompletedBuilds);

        // History filters
        document.getElementById('status-filter').addEventListener('change', filterHistory);
        document.getElementById('time-filter').addEventListener('change', filterHistory);
        document.getElementById('project-filter').addEventListener('input', filterHistory);
        document.getElementById('refresh-history').addEventListener('click', refreshHistory);
        document.getElementById('export-history').addEventListener('click', exportHistory);

        // Pagination
        document.getElementById('prev-page').addEventListener('click', () => changePage(-1));
        document.getElementById('next-page').addEventListener('click', () => changePage(1));

        // Worker management
        document.getElementById('add-worker').addEventListener('click', addWorker);
        document.getElementById('refresh-workers').addEventListener('click', refreshWorkers);
        document.getElementById('worker-settings').addEventListener('click', showWorkerSettings);

        // Templates
        document.querySelectorAll('.use-template').forEach(btn => {
            btn.addEventListener('click', (e) => useTemplate(e.target.closest('.template-card').dataset.template));
        });
        document.querySelectorAll('.edit-template').forEach(btn => {
            btn.addEventListener('click', (e) => editTemplate(e.target.closest('.template-card').dataset.template));
        });
        document.getElementById('create-template').addEventListener('click', createTemplate);

        // Batch operations
        document.getElementById('start-batch').addEventListener('click', startBatchBuild);
        document.getElementById('validate-batch').addEventListener('click', validateBatchProjects);
        document.getElementById('clear-batch').addEventListener('click', clearBatch);

        // Advanced controls
        document.getElementById('clear-cache').addEventListener('click', clearCache);
        document.getElementById('warm-cache').addEventListener('click', warmCache);
        document.getElementById('cache-stats').addEventListener('click', showCacheStats);
        document.getElementById('restart-coordinator').addEventListener('click', restartCoordinator);
        document.getElementById('system-health-check').addEventListener('click', runHealthCheck);
        document.getElementById('system-logs').addEventListener('click', showSystemLogs);
        document.getElementById('optimize-workers').addEventListener('click', optimizeWorkers);
        document.getElementById('performance-report').addEventListener('click', generatePerformanceReport);
        document.getElementById('resource-tuning').addEventListener('click', showResourceTuning);
    }

    function setupFormValidation() {
        // Real-time validation
        document.getElementById('project-path').addEventListener('blur', validateProjectPath);
    }

    async function handleBuildSubmit(e) {
        e.preventDefault();

        const formData = getFormData();
        if (!validateFormData(formData)) {
            return;
        }

        try {
            await submitBuild(formData);
        } catch (error) {
            showError('Failed to submit build: ' + error.message);
        }
    }

    function getFormData() {
        const task = document.getElementById('build-task').value;
        const customTask = document.getElementById('custom-task').value;

        return {
            project_path: document.getElementById('project-path').value.trim(),
            task_name: task === 'custom' ? customTask : task,
            cache_enabled: document.getElementById('cache-enabled').checked,
            parallel_build: document.getElementById('parallel-build').checked,
            offline_mode: document.getElementById('offline-mode').checked,
            gradle_version: document.getElementById('gradle-version').value,
            build_options: document.getElementById('build-options').value.trim()
        };
    }

    function validateFormData(data) {
        if (!data.project_path) {
            showError('Please specify a project path');
            return false;
        }

        if (!data.task_name) {
            showError('Please specify a build task');
            return false;
        }

        return true;
    }

    async function submitBuild(buildData) {
        const submitBtn = document.getElementById('submit-build');
        const originalText = submitBtn.textContent;

        submitBtn.disabled = true;
        submitBtn.textContent = '🚀 Submitting...';

        try {
            const response = await fetch(`${buildManager.apiEndpoint}/api/build`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(buildData)
            });

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            const result = await response.json();
            showSuccess(`Build submitted successfully! Build ID: ${result.build_id}`);

            // Refresh queue immediately
            await refreshQueue();
            await refreshHistory();

            // Clear form
            document.getElementById('build-form').reset();
            document.getElementById('cache-enabled').checked = true;
            document.getElementById('parallel-build').checked = true;

        } catch (error) {
            showError('Failed to submit build: ' + error.message);
        } finally {
            submitBtn.disabled = false;
            submitBtn.textContent = originalText;
        }
    }

    function handleTaskChange() {
        const taskSelect = document.getElementById('build-task');
        const customTaskInput = document.getElementById('custom-task');

        if (taskSelect.value === 'custom') {
            customTaskInput.style.display = 'block';
            customTaskInput.required = true;
        } else {
            customTaskInput.style.display = 'none';
            customTaskInput.required = false;
        }
    }

    async function validateProject() {
        const projectPath = document.getElementById('project-path').value.trim();
        if (!projectPath) {
            showError('Please enter a project path first');
            return;
        }

        // Mock validation - in real implementation, this would check the actual project
        showInfo('Validating project structure...');

        try {
            // Simulate validation delay
            await new Promise(resolve => setTimeout(resolve, 2000));

            // Mock validation result
            const isValid = Math.random() > 0.2; // 80% success rate

            if (isValid) {
                showSuccess('✅ Project validation successful! Ready to build.');
            } else {
                showError('❌ Invalid Gradle project. Please check the path and ensure build.gradle exists.');
            }
        } catch (error) {
            showError('Project validation failed: ' + error.message);
        }
    }

    function browseProject() {
        // In a real implementation, this would open a file browser
        // For demo purposes, set a mock path
        document.getElementById('project-path').value = '/home/user/my-gradle-project';
        showInfo('Project path set. Click "Validate Project" to check.');
    }

    async function refreshQueue() {
        try {
            const response = await fetch(`${buildManager.apiEndpoint}/api/queue`);
            if (!response.ok) throw new Error('Failed to fetch queue');

            const queueData = await response.json();
            updateQueueDisplay(queueData);

        } catch (error) {
            console.error('Queue refresh error:', error);
            showError('Failed to refresh build queue');
        }
    }

    function updateQueueDisplay(queueData) {
        const queueContainer = document.getElementById('build-queue');
        const queueCount = document.getElementById('queue-count');
        const activeCount = document.getElementById('active-count');
        const completedCount = document.getElementById('completed-count');

        // Update stats
        queueCount.textContent = queueData.queued || 0;
        activeCount.textContent = queueData.active || 0;
        completedCount.textContent = queueData.completed_today || 0;

        // Update queue items
        if (queueData.items && queueData.items.length > 0) {
            queueContainer.innerHTML = queueData.items.map(build => `
                <div class="build-item status-${build.status}" data-build-id="${build.build_id}">
                    <div class="build-header">
                        <h4>${build.build_id}</h4>
                        <span class="build-status">${build.status}</span>
                    </div>
                    <div class="build-details">
                        <div class="build-info">
                            <span class="label">Project:</span>
                            <span class="value">${build.project_path}</span>
                        </div>
                        <div class="build-info">
                            <span class="label">Task:</span>
                            <span class="value">${build.task_name}</span>
                        </div>
                        <div class="build-info">
                            <span class="label">Worker:</span>
                            <span class="value">${build.worker_id || 'Not assigned'}</span>
                        </div>
                        <div class="build-progress">
                            <div class="progress-bar">
                                <div class="progress-fill" style="width: ${getProgressPercentage(build)}%"></div>
                            </div>
                            <span class="progress-text">${getProgressText(build)}</span>
                        </div>
                    </div>
                    <div class="build-actions">
                        ${getBuildActions(build)}
                    </div>
                </div>
            `).join('');
        } else {
            queueContainer.innerHTML = '<div class="empty-queue">No builds in queue</div>';
        }
    }

    function getProgressPercentage(build) {
        switch (build.status) {
            case 'queued': return 0;
            case 'running': return Math.min(build.elapsed_time / build.estimated_time * 100, 90);
            case 'completed': return 100;
            case 'failed': return 100;
            default: return 0;
        }
    }

    function getProgressText(build) {
        switch (build.status) {
            case 'queued': return 'Waiting in queue...';
            case 'running': return `Running for ${formatDuration(build.elapsed_time || 0)}`;
            case 'completed': return `Completed in ${formatDuration(build.duration || 0)}`;
            case 'failed': return 'Build failed';
            default: return 'Unknown status';
        }
    }

    function getBuildActions(build) {
        const actions = [];

        if (build.status === 'running') {
            actions.push('<button class="btn-secondary cancel-build" data-build-id="' + build.build_id + '">Cancel</button>');
        }

        if (build.status === 'completed' || build.status === 'failed') {
            actions.push('<button class="btn-secondary view-logs" data-build-id="' + build.build_id + '">View Logs</button>');
            actions.push('<button class="btn-secondary rerun-build" data-build-id="' + build.build_id + '">Rerun</button>');
        }

        return actions.join('');
    }

    async function clearCompletedBuilds() {
        if (!confirm('Are you sure you want to clear all completed builds from the queue?')) {
            return;
        }

        try {
            const response = await fetch(`${buildManager.apiEndpoint}/api/queue/clear-completed`, {
                method: 'POST'
            });

            if (response.ok) {
                showSuccess('Completed builds cleared');
                await refreshQueue();
            } else {
                throw new Error('Failed to clear completed builds');
            }
        } catch (error) {
            showError('Failed to clear completed builds: ' + error.message);
        }
    }

    async function refreshHistory() {
        try {
            const filters = getHistoryFilters();
            const response = await fetch(`${buildManager.apiEndpoint}/api/builds/history?${new URLSearchParams(filters)}`);
            if (!response.ok) throw new Error('Failed to fetch history');

            const historyData = await response.json();
            updateHistoryDisplay(historyData);

        } catch (error) {
            console.error('History refresh error:', error);
            showError('Failed to refresh build history');
        }
    }

    function getHistoryFilters() {
        return {
            status: document.getElementById('status-filter').value,
            timeRange: document.getElementById('time-filter').value,
            project: document.getElementById('project-filter').value.trim()
        };
    }

    function updateHistoryDisplay(historyData) {
        const tbody = document.getElementById('build-history-body');

        if (!historyData.builds || historyData.builds.length === 0) {
            tbody.innerHTML = '<tr><td colspan="9">No builds found matching the current filters</td></tr>';
            return;
        }

        tbody.innerHTML = historyData.builds.map(build => `
            <tr>
                <td>${build.build_id}</td>
                <td>${build.project_path}</td>
                <td>${build.task_name}</td>
                <td><span class="status-${build.status}">${build.status}</span></td>
                <td>${formatDuration(build.duration || 0)}</td>
                <td>${build.worker_id || 'N/A'}</td>
                <td>${Math.round((build.cache_hit_rate || 0) * 100)}%</td>
                <td>${formatTimestamp(build.start_time)}</td>
                <td>
                    <button class="btn-secondary view-details" data-build-id="${build.build_id}">View</button>
                    ${build.status === 'completed' ? `<button class="btn-secondary rerun-build" data-build-id="${build.build_id}">Rerun</button>` : ''}
                </td>
            </tr>
        `).join('');

        // Update pagination
        updatePagination(historyData.pagination);
    }

    function updatePagination(pagination) {
        const prevBtn = document.getElementById('prev-page');
        const nextBtn = document.getElementById('next-page');
        const pageInfo = document.getElementById('page-info');

        prevBtn.disabled = !pagination.hasPrev;
        nextBtn.disabled = !pagination.hasNext;
        pageInfo.textContent = `Page ${pagination.current} of ${pagination.total}`;
    }

    function changePage(direction) {
        // Implement pagination logic
        refreshHistory();
    }

    function filterHistory() {
        // Debounce filtering
        clearTimeout(buildManager.filterTimeout);
        buildManager.filterTimeout = setTimeout(() => {
            refreshHistory();
        }, 300);
    }

    async function refreshWorkers() {
        try {
            const response = await fetch(`${buildManager.apiEndpoint}/api/workers`);
            if (!response.ok) throw new Error('Failed to fetch workers');

            const workersData = await response.json();
            updateWorkersDisplay(workersData);

        } catch (error) {
            console.error('Workers refresh error:', error);
            showError('Failed to refresh workers');
        }
    }

    function updateWorkersDisplay(workersData) {
        const grid = document.getElementById('worker-grid');
        const totalWorkers = document.getElementById('total-workers');
        const activeWorkers = document.getElementById('active-workers');
        const idleWorkers = document.getElementById('idle-workers');
        const avgUtilization = document.getElementById('avg-utilization');

        // Update stats
        totalWorkers.textContent = workersData.total || 0;
        activeWorkers.textContent = workersData.active || 0;
        idleWorkers.textContent = workersData.idle || 0;
        avgUtilization.textContent = `${Math.round(workersData.avgUtilization || 0)}%`;

        // Update worker grid
        if (workersData.workers && workersData.workers.length > 0) {
            grid.innerHTML = workersData.workers.map(worker => `
                <div class="worker-card status-${worker.status}">
                    <div class="worker-header">
                        <h4>${worker.id}</h4>
                        <span class="worker-status">${worker.status}</span>
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
                            <span class="label">Active Builds:</span>
                            <span class="value">${worker.active_builds || 0}</span>
                        </div>
                        <div class="metric">
                            <span class="label">Total Builds:</span>
                            <span class="value">${worker.total_builds || 0}</span>
                        </div>
                    </div>
                    <div class="worker-actions">
                        <button class="btn-secondary worker-details" data-worker-id="${worker.id}">Details</button>
                        ${worker.status === 'busy' ? '<button class="btn-warning stop-worker" data-worker-id="' + worker.id + '">Stop</button>' : ''}
                        <button class="btn-danger remove-worker" data-worker-id="${worker.id}">Remove</button>
                    </div>
                </div>
            `).join('');
        } else {
            grid.innerHTML = '<div class="worker-card empty"><div class="worker-header"><h4>No workers registered</h4></div></div>';
        }
    }

    // Template management functions
    function useTemplate(templateName) {
        // Load template configuration and apply to form
        const templates = getSavedTemplates();
        const template = templates[templateName];

        if (template) {
            Object.keys(template).forEach(key => {
                const element = document.getElementById(key);
                if (element) {
                    if (element.type === 'checkbox') {
                        element.checked = template[key];
                    } else {
                        element.value = template[key];
                    }
                }
            });
            showSuccess(`Template "${templateName}" loaded`);
        }
    }

    function editTemplate(templateName) {
        // Open template editor
        showInfo('Template editing not implemented in demo');
    }

    function createTemplate() {
        const formData = getFormData();
        const templateName = prompt('Enter template name:');

        if (templateName) {
            const templates = getSavedTemplates();
            templates[templateName] = formData;
            localStorage.setItem('build-templates', JSON.stringify(templates));
            showSuccess(`Template "${templateName}" saved`);
            refreshTemplates();
        }
    }

    function getSavedTemplates() {
        const saved = localStorage.getItem('build-templates');
        return saved ? JSON.parse(saved) : {};
    }

    function refreshTemplates() {
        // Refresh template display
        const templates = getSavedTemplates();
        // Update template grid with saved templates
    }

    // Batch operations
    async function startBatchBuild() {
        const projects = document.getElementById('batch-projects').value.trim().split('\n').filter(p => p.trim());
        if (projects.length === 0) {
            showError('Please enter at least one project path');
            return;
        }

        const batchConfig = {
            projects: projects,
            task: document.getElementById('batch-task').value,
            concurrency: parseInt(document.getElementById('batch-concurrency').value),
            sequential: document.getElementById('batch-sequential').checked
        };

        try {
            const response = await fetch(`${buildManager.apiEndpoint}/api/batch/build`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(batchConfig)
            });

            if (response.ok) {
                const result = await response.json();
                showSuccess(`Batch build started! Batch ID: ${result.batch_id}`);
                monitorBatchProgress(result.batch_id);
            } else {
                throw new Error('Failed to start batch build');
            }
        } catch (error) {
            showError('Batch build failed: ' + error.message);
        }
    }

    function monitorBatchProgress(batchId) {
        const progressDiv = document.getElementById('batch-progress');
        progressDiv.style.display = 'block';

        // Mock progress monitoring
        let completed = 0;
        const total = parseInt(document.getElementById('batch-concurrency').value);

        const interval = setInterval(() => {
            completed += Math.floor(Math.random() * 2);
            if (completed >= total) {
                completed = total;
                clearInterval(interval);
                setTimeout(() => {
                    progressDiv.style.display = 'none';
                    showSuccess('Batch build completed!');
                }, 2000);
            }

            document.getElementById('batch-completed').textContent = completed;
            document.getElementById('batch-running').textContent = Math.max(0, total - completed);
        }, 1000);
    }

    // Advanced controls
    async function clearCache() {
        if (!confirm('Are you sure you want to clear the build cache? This may slow down future builds.')) {
            return;
        }

        try {
            const response = await fetch(`${buildManager.apiEndpoint}/api/cache/clear`, {
                method: 'POST'
            });

            if (response.ok) {
                showSuccess('Build cache cleared successfully');
            } else {
                throw new Error('Failed to clear cache');
            }
        } catch (error) {
            showError('Failed to clear cache: ' + error.message);
        }
    }

    async function restartCoordinator() {
        if (!confirm('Are you sure you want to restart the coordinator? This will interrupt running builds.')) {
            return;
        }

        try {
            const response = await fetch(`${buildManager.apiEndpoint}/api/system/restart`, {
                method: 'POST'
            });

            if (response.ok) {
                showWarning('Coordinator restart initiated. System will be unavailable for a few moments.');
            } else {
                throw new Error('Failed to restart coordinator');
            }
        } catch (error) {
            showError('Failed to restart coordinator: ' + error.message);
        }
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

    function showSuccess(message) {
        showNotification(message, 'success');
    }

    function showError(message) {
        showNotification(message, 'error');
    }

    function showWarning(message) {
        showNotification(message, 'warning');
    }

    function showInfo(message) {
        showNotification(message, 'info');
    }

    function showNotification(message, type) {
        // Create notification element
        const notification = document.createElement('div');
        notification.className = `notification ${type}`;
        notification.innerHTML = `
            <span>${message}</span>
            <button class="notification-close">&times;</button>
        `;

        // Add to page
        document.body.appendChild(notification);

        // Auto-remove after 5 seconds
        setTimeout(() => {
            if (notification.parentNode) {
                notification.parentNode.removeChild(notification);
            }
        }, 5000);

        // Close button
        notification.querySelector('.notification-close').addEventListener('click', () => {
            notification.parentNode.removeChild(notification);
        });
    }

    async function refreshAllData() {
        await Promise.all([
            refreshQueue(),
            refreshHistory(),
            refreshWorkers()
        ]);
    }

    function loadSavedSettings() {
        const saved = localStorage.getItem('build-manager-settings');
        if (saved) {
            const settings = JSON.parse(saved);
            buildManager.apiEndpoint = settings.apiEndpoint || buildManager.apiEndpoint;
            buildManager.refreshInterval = settings.refreshInterval || buildManager.refreshInterval;
            buildManager.autoRefresh = settings.autoRefresh !== false;
        }
    }

    function startAutoRefresh() {
        if (buildManager.autoRefresh && !buildManager.refreshTimer) {
            buildManager.refreshTimer = setInterval(refreshAllData, buildManager.refreshInterval);
        }
    }

    // Stub functions for features not fully implemented
    function validateProjectPath() { /* Validation logic */ }
    function exportHistory() { showInfo('Export functionality would be implemented here'); }
    function addWorker() { showInfo('Add worker dialog would open here'); }
    function showWorkerSettings() { showInfo('Worker settings panel would open here'); }
    function validateBatchProjects() { showInfo('Batch validation would run here'); }
    function clearBatch() { document.getElementById('batch-projects').value = ''; }
    function warmCache() { showInfo('Cache warming would start here'); }
    function showCacheStats() { showInfo('Cache statistics would be displayed here'); }
    function runHealthCheck() { showInfo('Health check would run here'); }
    function showSystemLogs() { showInfo('System logs would be displayed here'); }
    function optimizeWorkers() { showInfo('Worker optimization would run here'); }
    function generatePerformanceReport() { showInfo('Performance report would be generated here'); }
    function showResourceTuning() { showInfo('Resource tuning panel would open here'); }
});

// Add CSS styles for the build manager
const style = document.createElement('style');
style.textContent = `
    #build-manager {
        font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
        max-width: 1400px;
        margin: 0 auto;
        padding: 20px;
    }

    .build-form-section, .build-queue-section, .build-history-section,
    .worker-management-section, .build-templates-section,
    .batch-operations-section, .advanced-controls-section {
        background: white;
        border: 1px solid #dee2e6;
        border-radius: 8px;
        padding: 20px;
        margin-bottom: 30px;
        box-shadow: 0 2px 4px rgba(0,0,0,0.1);
    }

    .build-form-section h3, .build-queue-section h3, .build-history-section h3,
    .worker-management-section h3, .build-templates-section h3,
    .batch-operations-section h3, .advanced-controls-section h3 {
        margin: 0 0 20px 0;
        color: #495057;
        border-bottom: 2px solid #007bff;
        padding-bottom: 10px;
    }

    .form-group {
        margin-bottom: 20px;
    }

    .form-group label {
        display: block;
        margin-bottom: 5px;
        font-weight: 600;
        color: #495057;
    }

    .input-group {
        display: flex;
        gap: 10px;
    }

    .form-group input, .form-group select, .form-group textarea {
        width: 100%;
        padding: 10px 12px;
        border: 1px solid #ced4da;
        border-radius: 4px;
        font-size: 14px;
        font-family: inherit;
    }

    .form-group textarea {
        min-height: 80px;
        resize: vertical;
    }

    .form-options {
        display: flex;
        flex-wrap: wrap;
        gap: 20px;
        margin: 20px 0;
    }

    .checkbox-label {
        display: flex;
        align-items: center;
        gap: 8px;
        font-weight: normal;
    }

    .checkbox-label input[type="checkbox"] {
        width: auto;
        margin: 0;
    }

    .form-actions {
        text-align: center;
        margin-top: 30px;
    }

    .btn-primary, .btn-secondary, .btn-warning, .btn-danger, .btn-primary-large {
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
        transition: all 0.2s ease;
    }

    .btn-primary-large {
        padding: 15px 30px;
        font-size: 16px;
    }

    .btn-secondary {
        background: #6c757d;
    }

    .btn-warning {
        background: #ffc107;
        color: #212529;
    }

    .btn-danger {
        background: #dc3545;
    }

    .btn-primary:hover, .btn-secondary:hover, .btn-warning:hover,
    .btn-danger:hover, .btn-primary-large:hover {
        opacity: 0.9;
        transform: translateY(-1px);
    }

    .queue-controls {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 20px;
        flex-wrap: wrap;
        gap: 10px;
    }

    .queue-stats {
        font-size: 16px;
        color: #495057;
    }

    .queue-actions {
        display: flex;
        gap: 10px;
    }

    .build-queue {
        min-height: 200px;
    }

    .build-item {
        border: 1px solid #dee2e6;
        border-radius: 6px;
        padding: 15px;
        margin-bottom: 15px;
        background: white;
    }

    .build-item.status-queued {
        border-color: #ffc107;
        background: #fff3cd;
    }

    .build-item.status-running {
        border-color: #007bff;
        background: #d1ecf1;
    }

    .build-item.status-completed {
        border-color: #28a745;
        background: #d4edda;
    }

    .build-item.status-failed {
        border-color: #dc3545;
        background: #f8d7da;
    }

    .build-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 10px;
    }

    .build-header h4 {
        margin: 0;
        color: #495057;
    }

    .build-status {
        padding: 3px 8px;
        border-radius: 4px;
        font-size: 12px;
        font-weight: bold;
        text-transform: uppercase;
    }

    .build-details {
        margin-bottom: 10px;
    }

    .build-info {
        display: flex;
        margin-bottom: 5px;
        font-size: 14px;
    }

    .build-info .label {
        font-weight: 600;
        color: #495057;
        min-width: 80px;
    }

    .build-info .value {
        color: #6c757d;
    }

    .build-progress {
        margin-top: 10px;
    }

    .progress-bar {
        background: #e9ecef;
        border-radius: 10px;
        height: 8px;
        overflow: hidden;
        margin-bottom: 5px;
    }

    .progress-fill {
        background: linear-gradient(90deg, #28a745, #20c997);
        height: 100%;
        width: 0%;
        border-radius: 10px;
        transition: width 0.5s ease;
    }

    .progress-text {
        font-size: 12px;
        color: #6c757d;
        text-align: center;
    }

    .build-actions {
        text-align: right;
    }

    .build-actions .btn-secondary {
        font-size: 12px;
        padding: 5px 10px;
        margin-left: 5px;
    }

    .empty-queue {
        text-align: center;
        padding: 40px;
        color: #6c757d;
        font-style: italic;
    }

    .history-controls {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 20px;
        flex-wrap: wrap;
        gap: 10px;
    }

    .filter-controls {
        display: flex;
        gap: 10px;
        flex-wrap: wrap;
    }

    .filter-controls select, .filter-controls input {
        padding: 8px 12px;
        border: 1px solid #ced4da;
        border-radius: 4px;
        font-size: 14px;
    }

    .history-actions {
        display: flex;
        gap: 10px;
    }

    .data-table {
        width: 100%;
        border-collapse: collapse;
        font-size: 14px;
        margin-top: 20px;
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
        position: sticky;
        top: 0;
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
        color: #007bff;
        font-weight: bold;
    }

    .status-queued {
        color: #ffc107;
        font-weight: bold;
    }

    .pagination-controls {
        display: flex;
        justify-content: center;
        align-items: center;
        gap: 20px;
        margin-top: 20px;
        padding: 20px;
    }

    .worker-overview {
        margin-bottom: 30px;
    }

    .worker-stats {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
        gap: 20px;
    }

    .stat-card {
        background: #f8f9fa;
        border: 1px solid #dee2e6;
        border-radius: 8px;
        padding: 20px;
        text-align: center;
    }

    .stat-value {
        font-size: 36px;
        font-weight: bold;
        color: #007bff;
        display: block;
        margin: 10px 0;
    }

    .stat-label {
        color: #6c757d;
        font-size: 14px;
    }

    .worker-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
        gap: 20px;
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

    .worker-card.empty {
        text-align: center;
        color: #6c757d;
        font-style: italic;
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
        margin-bottom: 10px;
    }

    .worker-metrics .metric {
        display: flex;
        justify-content: space-between;
        font-size: 13px;
    }

    .worker-metrics .label {
        font-weight: 600;
        color: #495057;
    }

    .worker-metrics .value {
        color: #007bff;
    }

    .worker-actions {
        text-align: right;
    }

    .worker-actions .btn-secondary {
        font-size: 12px;
        padding: 5px 10px;
        margin-left: 5px;
    }

    .template-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
        gap: 20px;
    }

    .template-card {
        border: 2px solid #dee2e6;
        border-radius: 8px;
        padding: 20px;
        text-align: center;
        cursor: pointer;
        transition: all 0.3s ease;
    }

    .template-card:hover {
        border-color: #007bff;
        box-shadow: 0 4px 8px rgba(0,123,255,0.2);
    }

    .template-card h4 {
        margin: 0 0 10px 0;
        color: #495057;
    }

    .template-card p {
        margin: 0 0 15px 0;
        color: #6c757d;
        font-size: 14px;
    }

    .template-actions {
        display: flex;
        gap: 10px;
        justify-content: center;
    }

    .batch-controls {
        display: grid;
        grid-template-columns: 1fr 300px;
        gap: 20px;
        margin-bottom: 20px;
    }

    .batch-input textarea {
        width: 100%;
        min-height: 150px;
        padding: 10px;
        border: 1px solid #ced4da;
        border-radius: 4px;
        font-family: monospace;
        font-size: 14px;
    }

    .batch-settings {
        display: flex;
        flex-direction: column;
        gap: 15px;
    }

    .setting-group {
        display: flex;
        flex-direction: column;
        gap: 5px;
    }

    .setting-group label {
        font-weight: 600;
        color: #495057;
        font-size: 14px;
    }

    .setting-group select, .setting-group input {
        padding: 8px;
        border: 1px solid #ced4da;
        border-radius: 4px;
        font-size: 14px;
    }

    .batch-actions {
        text-align: center;
        margin: 20px 0;
    }

    .batch-progress {
        background: #f8f9fa;
        border: 1px solid #dee2e6;
        border-radius: 8px;
        padding: 20px;
        margin-top: 20px;
    }

    .progress-summary {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
        gap: 15px;
        margin-bottom: 20px;
    }

    .progress-item {
        text-align: center;
    }

    .progress-item span:first-child {
        display: block;
        font-weight: 600;
        color: #495057;
        font-size: 14px;
    }

    .progress-item span:last-child {
        display: block;
        font-size: 24px;
        font-weight: bold;
        color: #007bff;
        margin-top: 5px;
    }

    .control-panels {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
        gap: 20px;
    }

    .control-panel {
        background: #f8f9fa;
        border: 1px solid #dee2e6;
        border-radius: 6px;
        padding: 15px;
    }

    .control-panel h4 {
        margin: 0 0 15px 0;
        color: #495057;
    }

    .control-actions {
        display: flex;
        flex-wrap: wrap;
        gap: 10px;
    }

    .control-actions .btn-secondary {
        flex: 1;
        min-width: 120px;
    }

    .system-status {
        background: white;
        border: 1px solid #dee2e6;
        border-radius: 6px;
        padding: 15px;
        margin-top: 20px;
    }

    .status-indicators {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
        gap: 15px;
    }

    .status-item {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 10px 0;
        border-bottom: 1px solid #f0f0f0;
    }

    .status-item:last-child {
        border-bottom: none;
    }

    .status-label {
        font-weight: 600;
        color: #495057;
    }

    .status-value {
        font-weight: bold;
    }

    .status-value.healthy {
        color: #28a745;
    }

    .status-value.warning {
        color: #ffc107;
    }

    .status-value.error {
        color: #dc3545;
    }

    .notification {
        position: fixed;
        top: 20px;
        right: 20px;
        background: #007bff;
        color: white;
        padding: 15px 20px;
        border-radius: 6px;
        box-shadow: 0 4px 12px rgba(0,0,0,0.3);
        z-index: 1000;
        max-width: 400px;
        display: flex;
        justify-content: space-between;
        align-items: center;
    }

    .notification.success {
        background: #28a745;
    }

    .notification.error {
        background: #dc3545;
    }

    .notification.warning {
        background: #ffc107;
        color: #212529;
    }

    .notification.info {
        background: #17a2b8;
    }

    .notification-close {
        background: none;
        border: none;
        color: inherit;
        font-size: 20px;
        cursor: pointer;
        padding: 0;
        margin-left: 15px;
    }

    @media (max-width: 768px) {
        #build-manager {
            padding: 10px;
        }

        .input-group {
            flex-direction: column;
        }

        .form-options {
            flex-direction: column;
            gap: 10px;
        }

        .queue-controls, .history-controls {
            flex-direction: column;
            align-items: stretch;
        }

        .filter-controls {
            flex-direction: column;
        }

        .worker-stats {
            grid-template-columns: 1fr;
        }

        .template-grid {
            grid-template-columns: 1fr;
        }

        .batch-controls {
            grid-template-columns: 1fr;
        }

        .control-panels {
            grid-template-columns: 1fr;
        }

        .status-indicators {
            grid-template-columns: 1fr;
        }

        .progress-summary {
            grid-template-columns: 1fr;
        }

        .worker-metrics {
            grid-template-columns: 1fr;
        }

        .template-actions {
            flex-direction: column;
        }

        .control-actions {
            flex-direction: column;
        }

        .batch-actions .btn-primary-large {
            width: 100%;
        }

        .form-actions .btn-primary-large {
            width: 100%;
        }
    }
`;
document.head.appendChild(style);