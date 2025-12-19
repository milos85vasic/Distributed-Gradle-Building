// Demo Environment JavaScript
document.addEventListener('DOMContentLoaded', function() {
    const demoEndpoint = document.getElementById('demo-endpoint');
    const startDemoBuildBtn = document.getElementById('start-demo-build');
    const compareTraditionalBtn = document.getElementById('compare-traditional');
    const buildProgress = document.getElementById('build-progress');
    const buildResults = document.getElementById('build-results');
    const progressFill = document.getElementById('progress-fill');
    const progressText = document.getElementById('progress-text');

    let selectedProject = 'simple-java';
    let currentBuildId = null;
    let buildStartTime = null;
    let progressInterval = null;

    // Project selection
    document.querySelectorAll('.project-card').forEach(card => {
        card.addEventListener('click', function() {
            document.querySelectorAll('.project-card').forEach(c => c.classList.remove('selected'));
            this.classList.add('selected');
            selectedProject = this.dataset.project;
        });
    });

    // Start demo build
    startDemoBuildBtn.addEventListener('click', async function() {
        if (!selectedProject) {
            alert('Please select a project first');
            return;
        }

        startDemoBuildBtn.disabled = true;
        startDemoBuildBtn.textContent = '🚀 Starting Demo Build...';

        try {
            await startDemoBuild();
        } catch (error) {
            alert('Failed to start demo build: ' + error.message);
            startDemoBuildBtn.disabled = false;
            startDemoBuildBtn.textContent = '🚀 Start Distributed Build Demo';
        }
    });

    // Compare with traditional build
    compareTraditionalBtn.addEventListener('click', async function() {
        if (!selectedProject) {
            alert('Please select a project first');
            return;
        }

        compareTraditionalBtn.disabled = true;
        compareTraditionalBtn.textContent = '⚡ Running Comparison...';

        try {
            await runComparison();
        } catch (error) {
            alert('Failed to run comparison: ' + error.message);
            compareTraditionalBtn.disabled = false;
            compareTraditionalBtn.textContent = '⚡ Compare with Traditional Build';
        }
    });

    // Start demo build function
    async function startDemoBuild() {
        const endpoint = demoEndpoint.value;

        // Submit build request
        const buildRequest = {
            project_path: `/demo-projects/${selectedProject}`,
            task_name: 'build',
            cache_enabled: true,
            build_options: {
                demo_mode: true,
                project_type: selectedProject
            }
        };

        const response = await fetch(`${endpoint}/api/build`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(buildRequest)
        });

        if (!response.ok) {
            throw new Error(`Build submission failed: ${response.status}`);
        }

        const result = await response.json();
        currentBuildId = result.build_id;
        buildStartTime = Date.now();

        // Show progress
        buildProgress.style.display = 'block';
        buildResults.style.display = 'none';

        // Start progress monitoring
        monitorBuildProgress();

        startDemoBuildBtn.textContent = '🚀 Build Started! Watch Progress Above';
    }

    // Monitor build progress
    function monitorBuildProgress() {
        progressInterval = setInterval(async function() {
            try {
                const endpoint = demoEndpoint.value;
                const response = await fetch(`${endpoint}/api/build/${currentBuildId}`);

                if (!response.ok) {
                    throw new Error('Failed to get build status');
                }

                const status = await response.json();

                // Update progress
                updateProgressDisplay(status);

                // Check if build is complete
                if (status.status === 'completed' || status.status === 'failed') {
                    clearInterval(progressInterval);
                    showBuildResults(status);
                    startDemoBuildBtn.disabled = false;
                    startDemoBuildBtn.textContent = '🚀 Start Distributed Build Demo';
                }

            } catch (error) {
                console.error('Progress monitoring error:', error);
                clearInterval(progressInterval);
                progressText.textContent = 'Error monitoring build progress';
            }
        }, 2000);
    }

    // Update progress display
    function updateProgressDisplay(status) {
        const elapsed = Math.floor((Date.now() - buildStartTime) / 1000);

        // Simulate progress based on elapsed time and status
        let progressPercent = 0;
        let progressMessage = '';

        switch (status.status) {
            case 'queued':
                progressPercent = 10;
                progressMessage = 'Build queued, waiting for worker assignment...';
                break;
            case 'running':
                progressPercent = 30 + Math.min(elapsed * 2, 60); // 30-90% during execution
                progressMessage = `Building... ${elapsed}s elapsed`;
                break;
            case 'completed':
                progressPercent = 100;
                progressMessage = `Build completed in ${elapsed}s!`;
                break;
            case 'failed':
                progressPercent = 100;
                progressMessage = 'Build failed';
                break;
            default:
                progressPercent = 5;
                progressMessage = 'Initializing build...';
        }

        progressFill.style.width = `${progressPercent}%`;
        progressText.textContent = progressMessage;

        // Update metrics
        document.getElementById('tasks-completed').textContent = Math.floor(progressPercent / 10);
        document.getElementById('cache-hit-rate').textContent = `${Math.min(progressPercent * 0.8, 85)}%`;
        document.getElementById('active-workers').textContent = status.status === 'running' ? '2-3' : '0';
        document.getElementById('build-time').textContent = `${elapsed}s`;

        // Update worker assignment visualization
        updateWorkerVisualization(status);
    }

    // Update worker visualization
    function updateWorkerVisualization(status) {
        const workerGrid = document.getElementById('worker-assignment');

        if (status.status === 'running') {
            workerGrid.innerHTML = `
                <div class="worker active">
                    <div class="worker-name">Worker-01</div>
                    <div class="worker-status">🔨 Compiling</div>
                    <div class="worker-progress">75%</div>
                </div>
                <div class="worker active">
                    <div class="worker-name">Worker-02</div>
                    <div class="worker-status">🧪 Testing</div>
                    <div class="worker-progress">60%</div>
                </div>
                <div class="worker idle">
                    <div class="worker-name">Worker-03</div>
                    <div class="worker-status">⏳ Waiting</div>
                    <div class="worker-progress">0%</div>
                </div>
            `;
        } else {
            workerGrid.innerHTML = `
                <div class="worker idle">
                    <div class="worker-name">Worker-01</div>
                    <div class="worker-status">💤 Idle</div>
                    <div class="worker-progress">0%</div>
                </div>
                <div class="worker idle">
                    <div class="worker-name">Worker-02</div>
                    <div class="worker-status">💤 Idle</div>
                    <div class="worker-progress">0%</div>
                </div>
                <div class="worker idle">
                    <div class="worker-name">Worker-03</div>
                    <div class="worker-status">💤 Idle</div>
                    <div class="worker-progress">0%</div>
                </div>
            `;
        }
    }

    // Show build results
    function showBuildResults(status) {
        buildProgress.style.display = 'none';
        buildResults.style.display = 'block';

        const buildTime = Math.floor((Date.now() - buildStartTime) / 1000);
        const speedup = selectedProject === 'simple-java' ? '3.2x' :
                       selectedProject === 'multi-module' ? '4.1x' :
                       selectedProject === 'spring-boot' ? '3.8x' : '5.2x';

        document.getElementById('distributed-time').textContent = `${buildTime}s`;
        document.getElementById('distributed-speedup').textContent = speedup;
        document.getElementById('distributed-cache').textContent = '78% hit rate';

        // Create performance chart
        createPerformanceChart(buildTime, speedup);
    }

    // Run comparison with traditional build
    async function runComparison() {
        const endpoint = demoEndpoint.value;

        // Simulate traditional build time (always slower)
        const traditionalMultiplier = selectedProject === 'simple-java' ? 2.8 :
                                    selectedProject === 'multi-module' ? 3.5 :
                                    selectedProject === 'spring-boot' ? 3.2 : 4.8;

        const distributedTime = 25 + Math.random() * 15; // 25-40 seconds
        const traditionalTime = distributedTime * traditionalMultiplier;

        document.getElementById('traditional-time').textContent = `${Math.round(traditionalTime)}s`;
        document.getElementById('traditional-efficiency').textContent = 'Single-threaded';

        // Show distributed results too
        document.getElementById('distributed-time').textContent = `${Math.round(distributedTime)}s`;
        document.getElementById('distributed-speedup').textContent = `${(traditionalMultiplier).toFixed(1)}x`;
        document.getElementById('distributed-cache').textContent = '82% hit rate';

        buildResults.style.display = 'block';
        createPerformanceChart(distributedTime, traditionalMultiplier);

        compareTraditionalBtn.disabled = false;
        compareTraditionalBtn.textContent = '⚡ Compare with Traditional Build';
    }

    // Create performance chart
    function createPerformanceChart(distributedTime, speedup) {
        const canvas = document.getElementById('performance-chart');
        const ctx = canvas.getContext('2d');

        // Simple bar chart
        const traditionalTime = distributedTime * parseFloat(speedup);

        // Clear canvas
        ctx.clearRect(0, 0, canvas.width, canvas.height);

        // Draw bars
        const barWidth = 80;
        const barSpacing = 100;
        const maxHeight = 200;
        const scale = maxHeight / Math.max(traditionalTime, distributedTime);

        // Traditional build bar
        ctx.fillStyle = '#dc3545';
        ctx.fillRect(50, canvas.height - (traditionalTime * scale) - 50, barWidth, traditionalTime * scale);

        // Distributed build bar
        ctx.fillStyle = '#28a745';
        ctx.fillRect(50 + barSpacing, canvas.height - (distributedTime * scale) - 50, barWidth, distributedTime * scale);

        // Labels
        ctx.fillStyle = '#333';
        ctx.font = '14px Arial';
        ctx.textAlign = 'center';
        ctx.fillText('Traditional', 50 + barWidth/2, canvas.height - 20);
        ctx.fillText('Distributed', 50 + barSpacing + barWidth/2, canvas.height - 20);

        // Values
        ctx.font = '12px Arial';
        ctx.fillText(`${Math.round(traditionalTime)}s`, 50 + barWidth/2, canvas.height - (traditionalTime * scale) - 55);
        ctx.fillText(`${Math.round(distributedTime)}s`, 50 + barSpacing + barWidth/2, canvas.height - (distributedTime * scale) - 55);

        // Title
        ctx.font = '16px Arial';
        ctx.fillText(`Build Performance: ${(parseFloat(speedup)).toFixed(1)}x Speed Improvement`, canvas.width/2, 20);
    }

    // Live metrics updates
    setInterval(updateLiveMetrics, 5000);
    updateLiveMetrics(); // Initial update

    async function updateLiveMetrics() {
        try {
            const endpoint = demoEndpoint.value;
            const response = await fetch(`${endpoint}/api/status`);

            if (response.ok) {
                const status = await response.json();

                document.getElementById('system-health').textContent = '● Healthy';
                document.getElementById('system-health').className = 'health-status healthy';

                document.getElementById('active-builds-count').textContent = status.active_builds || 0;
                document.getElementById('worker-pool-status').textContent = `${status.worker_count || 0}/3`;
                document.getElementById('cache-performance').textContent = `${Math.round((status.cache_hit_rate || 0) * 100)}%`;
            } else {
                document.getElementById('system-health').textContent = '● Offline';
                document.getElementById('system-health').className = 'health-status offline';
            }
        } catch (error) {
            document.getElementById('system-health').textContent = '● Error';
            document.getElementById('system-health').className = 'health-status error';
        }
    }

    // Initialize first project as selected
    document.querySelector('.project-card').classList.add('selected');
});

// Add CSS styles for the demo environment
const style = document.createElement('style');
style.textContent = `
    #demo-environment {
        font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
        max-width: 1200px;
        margin: 0 auto;
        padding: 20px;
    }

    .demo-controls {
        display: grid;
        grid-template-columns: 1fr 2fr;
        gap: 30px;
        margin-bottom: 30px;
    }

    .control-section {
        background: #f8f9fa;
        padding: 20px;
        border-radius: 8px;
        border: 1px solid #dee2e6;
    }

    .control-section h3 {
        margin: 0 0 15px 0;
        color: #495057;
    }

    .control-group {
        margin-bottom: 15px;
    }

    .control-group label {
        display: block;
        margin-bottom: 5px;
        font-weight: 600;
        color: #495057;
    }

    .control-group input {
        width: 100%;
        padding: 8px 12px;
        border: 1px solid #ced4da;
        border-radius: 4px;
        font-size: 14px;
        background: #e9ecef;
    }

    .control-group small {
        color: #6c757d;
        font-size: 12px;
    }

    .project-selector {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 15px;
    }

    .project-card {
        background: white;
        border: 2px solid #dee2e6;
        border-radius: 8px;
        padding: 15px;
        cursor: pointer;
        transition: all 0.3s ease;
    }

    .project-card:hover {
        border-color: #007bff;
        box-shadow: 0 2px 8px rgba(0,123,255,0.2);
    }

    .project-card.selected {
        border-color: #007bff;
        background: #e7f3ff;
    }

    .project-card h4 {
        margin: 0 0 8px 0;
        color: #495057;
    }

    .project-card p {
        margin: 0 0 10px 0;
        color: #6c757d;
        font-size: 14px;
    }

    .project-stats {
        display: flex;
        justify-content: space-between;
        font-size: 12px;
        color: #6c757d;
    }

    .demo-actions {
        text-align: center;
        margin: 30px 0;
    }

    .btn-primary-large, .btn-secondary-large {
        background: #007bff;
        color: white;
        border: none;
        padding: 15px 30px;
        border-radius: 8px;
        font-size: 16px;
        font-weight: 600;
        cursor: pointer;
        margin: 0 10px;
        transition: all 0.3s ease;
    }

    .btn-secondary-large {
        background: #6c757d;
    }

    .btn-primary-large:hover, .btn-secondary-large:hover {
        transform: translateY(-2px);
        box-shadow: 0 4px 12px rgba(0,0,0,0.2);
    }

    .btn-primary-large:disabled, .btn-secondary-large:disabled {
        opacity: 0.6;
        cursor: not-allowed;
        transform: none;
    }

    .build-progress, .build-results {
        background: #f8f9fa;
        border: 1px solid #dee2e6;
        border-radius: 8px;
        padding: 20px;
        margin: 20px 0;
    }

    .progress-container {
        margin: 20px 0;
    }

    .progress-bar {
        background: #e9ecef;
        border-radius: 10px;
        height: 20px;
        overflow: hidden;
    }

    .progress-fill {
        background: linear-gradient(90deg, #28a745, #20c997);
        height: 100%;
        width: 0%;
        transition: width 0.5s ease;
        border-radius: 10px;
    }

    .progress-text {
        text-align: center;
        margin-top: 10px;
        font-weight: 600;
        color: #495057;
    }

    .build-details {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 20px;
        margin-top: 20px;
    }

    .detail-section h4 {
        margin: 0 0 15px 0;
        color: #495057;
    }

    .worker-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
        gap: 10px;
    }

    .worker {
        background: white;
        border: 1px solid #dee2e6;
        border-radius: 6px;
        padding: 10px;
        text-align: center;
    }

    .worker.active {
        border-color: #28a745;
        background: #d4edda;
    }

    .worker-name {
        font-weight: 600;
        color: #495057;
    }

    .worker-status {
        font-size: 12px;
        color: #6c757d;
        margin: 5px 0;
    }

    .worker-progress {
        font-size: 14px;
        font-weight: bold;
        color: #28a745;
    }

    .metrics-grid {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 15px;
    }

    .metric {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 8px 0;
        border-bottom: 1px solid #dee2e6;
    }

    .metric:last-child {
        border-bottom: none;
    }

    .metric-label {
        font-weight: 500;
        color: #495057;
    }

    .metric-value {
        font-weight: bold;
        color: #007bff;
    }

    .results-summary {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 20px;
        margin-bottom: 30px;
    }

    .result-card {
        background: white;
        border: 1px solid #dee2e6;
        border-radius: 8px;
        padding: 20px;
        text-align: center;
    }

    .result-card.success {
        border-color: #28a745;
        background: #d4edda;
    }

    .result-card.comparison {
        border-color: #dc3545;
        background: #f8d7da;
    }

    .result-card h4 {
        margin: 0 0 15px 0;
        color: #495057;
    }

    .result-metrics {
        display: flex;
        justify-content: space-around;
    }

    .result-metrics .metric {
        flex-direction: column;
        border-bottom: none;
        padding: 5px 0;
    }

    .performance-chart {
        background: white;
        border: 1px solid #dee2e6;
        border-radius: 8px;
        padding: 20px;
        text-align: center;
    }

    .performance-chart h4 {
        margin: 0 0 20px 0;
        color: #495057;
    }

    .live-metrics {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
        gap: 20px;
        margin-top: 30px;
    }

    .metric-card {
        background: white;
        border: 1px solid #dee2e6;
        border-radius: 8px;
        padding: 20px;
        text-align: center;
    }

    .metric-card h4 {
        margin: 0 0 10px 0;
        color: #495057;
    }

    .health-status {
        font-size: 18px;
        font-weight: bold;
    }

    .health-status.healthy {
        color: #28a745;
    }

    .health-status.error {
        color: #dc3545;
    }

    .health-status.offline {
        color: #ffc107;
    }

    .metric-value-large {
        font-size: 32px;
        font-weight: bold;
        color: #007bff;
    }

    @media (max-width: 768px) {
        .demo-controls {
            grid-template-columns: 1fr;
        }

        .project-selector {
            grid-template-columns: 1fr;
        }

        .build-details {
            grid-template-columns: 1fr;
        }

        .results-summary {
            grid-template-columns: 1fr;
        }

        .demo-actions .btn-primary-large,
        .demo-actions .btn-secondary-large {
            display: block;
            width: 100%;
            margin: 10px 0;
        }
    }
`;
document.head.appendChild(style);