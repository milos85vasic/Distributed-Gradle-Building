// Interactive Tutorial JavaScript
document.addEventListener('DOMContentLoaded', function() {
    let currentStep = 1;
    const totalSteps = 6;
    let tutorialData = {
        worker1Host: '',
        worker2Host: '',
        projectPath: ''
    };

    // Initialize tutorial
    initializeTutorial();

    function initializeTutorial() {
        updateStepStatuses();
        setupEventListeners();
        loadSavedProgress();
    }

    function setupEventListeners() {
        // Step 1 validation
        document.getElementById('step1-validate').addEventListener('click', () => validateStep1());

        // RAM check
        document.querySelector('[data-check="ram"]').addEventListener('click', () => checkRAM());
        document.querySelector('[data-check="java"]').addEventListener('click', () => checkJava());
        document.querySelector('[data-check="network"]').addEventListener('click', () => checkNetwork());

        // Step 2 validation
        document.getElementById('step2-validate').addEventListener('click', () => validateStep2());

        // SSH testing
        document.getElementById('test-ssh').addEventListener('click', () => testSSHConnections());

        // Step 3 validation
        document.getElementById('step3-validate').addEventListener('click', () => validateStep3());

        // Project browsing
        document.getElementById('browse-project').addEventListener('click', () => browseProject());

        // Step 4 validation
        document.getElementById('step4-validate').addEventListener('click', () => validateStep4());

        // Worker checking
        document.getElementById('check-workers').addEventListener('click', () => checkWorkers());

        // Step 5 validation
        document.getElementById('step5-validate').addEventListener('click', () => validateStep5());

        // Build execution
        document.getElementById('run-traditional').addEventListener('click', () => runTraditionalBuild());
        document.getElementById('run-distributed').addEventListener('click', () => runDistributedBuild());

        // Step 6 validation
        document.getElementById('step6-validate').addEventListener('click', () => completeTutorial());

        // Certificate download
        document.getElementById('download-certificate').addEventListener('click', () => downloadCertificate());
    }

    function updateStepStatuses() {
        for (let i = 1; i <= totalSteps; i++) {
            const statusElement = document.getElementById(`step${i}-status`);
            const validateButton = document.getElementById(`step${i}-validate`);

            if (i < currentStep) {
                statusElement.textContent = 'Completed';
                statusElement.className = 'step-status completed';
                if (validateButton) validateButton.disabled = true;
            } else if (i === currentStep) {
                statusElement.textContent = 'In Progress';
                statusElement.className = 'step-status in-progress';
                if (validateButton) validateButton.disabled = false;
            } else {
                statusElement.textContent = 'Locked';
                statusElement.className = 'step-status locked';
                if (validateButton) validateButton.disabled = true;
            }
        }
    }

    function loadSavedProgress() {
        const saved = localStorage.getItem('distributed-gradle-tutorial');
        if (saved) {
            tutorialData = JSON.parse(saved);
            currentStep = tutorialData.currentStep || 1;

            // Restore form values
            if (tutorialData.worker1Host) {
                document.getElementById('worker1-host').value = tutorialData.worker1Host;
                document.getElementById('worker1-host-display').textContent = tutorialData.worker1Host;
            }
            if (tutorialData.worker2Host) {
                document.getElementById('worker2-host').value = tutorialData.worker2Host;
                document.getElementById('worker2-host-display').textContent = tutorialData.worker2Host;
            }
            if (tutorialData.projectPath) {
                document.getElementById('project-path').value = tutorialData.projectPath;
            }

            updateStepStatuses();
        }
    }

    function saveProgress() {
        tutorialData.currentStep = currentStep;
        localStorage.setItem('distributed-gradle-tutorial', JSON.stringify(tutorialData));
    }

    // Step validation functions
    async function validateStep1() {
        const resultDiv = document.getElementById('step1-result');
        resultDiv.innerHTML = '<div class="checking">Validating system requirements...</div>';

        try {
            // Simulate validation (in real implementation, this would check actual system state)
            await new Promise(resolve => setTimeout(resolve, 2000));

            const ramOk = document.getElementById('ram-check').textContent.includes('✅');
            const javaOk = document.getElementById('java-check').textContent.includes('✅');
            const networkOk = document.getElementById('network-check').textContent.includes('✅');

            if (ramOk && javaOk && networkOk) {
                resultDiv.innerHTML = '<div class="success">✅ Step 1 completed! System requirements verified.</div>';
                currentStep = 2;
                updateStepStatuses();
                saveProgress();
            } else {
                resultDiv.innerHTML = '<div class="error">❌ Please complete all requirement checks before proceeding.</div>';
            }
        } catch (error) {
            resultDiv.innerHTML = `<div class="error">❌ Validation failed: ${error.message}</div>`;
        }
    }

    async function checkRAM() {
        const resultElement = document.getElementById('ram-check');
        resultElement.textContent = 'Checking...';

        try {
            // Simulate RAM check (in real implementation, this would query system info)
            await new Promise(resolve => setTimeout(resolve, 1000));

            // Mock result - in real implementation, check actual RAM
            const ramGB = 8; // Mock 8GB RAM
            if (ramGB >= 4) {
                resultElement.textContent = `✅ ${ramGB}GB RAM available`;
                resultElement.className = 'check-result success';
            } else {
                resultElement.textContent = `❌ Only ${ramGB}GB RAM (4GB minimum required)`;
                resultElement.className = 'check-result error';
            }
        } catch (error) {
            resultElement.textContent = '❌ Check failed';
            resultElement.className = 'check-result error';
        }
    }

    async function checkJava() {
        const resultElement = document.getElementById('java-check');
        resultElement.textContent = 'Checking...';

        try {
            // Simulate Java check
            await new Promise(resolve => setTimeout(resolve, 1000));

            // Mock result - in real implementation, check actual Java version
            const javaVersion = '11.0.15';
            const majorVersion = parseInt(javaVersion.split('.')[0]);
            if (majorVersion >= 8) {
                resultElement.textContent = `✅ Java ${javaVersion} installed`;
                resultElement.className = 'check-result success';
            } else {
                resultElement.textContent = `❌ Java ${javaVersion} (Java 8+ required)`;
                resultElement.className = 'check-result error';
            }
        } catch (error) {
            resultElement.textContent = '❌ Check failed';
            resultElement.className = 'check-result error';
        }
    }

    async function checkNetwork() {
        const resultElement = document.getElementById('network-check');
        resultElement.textContent = 'Checking...';

        try {
            // Simulate network check
            await new Promise(resolve => setTimeout(resolve, 1000));

            // Mock result - in real implementation, test actual connectivity
            resultElement.textContent = '✅ Network connectivity verified';
            resultElement.className = 'check-result success';
        } catch (error) {
            resultElement.textContent = '❌ Network check failed';
            resultElement.className = 'check-result error';
        }
    }

    async function validateStep2() {
        const resultDiv = document.getElementById('step2-result');
        resultDiv.innerHTML = '<div class="checking">Validating installation...</div>';

        try {
            await new Promise(resolve => setTimeout(resolve, 1500));

            // Mock validation - in real implementation, check actual file existence
            const setupMasterExists = true; // Mock
            const setupWorkerExists = true; // Mock
            const syncBuildExists = true; // Mock

            if (setupMasterExists && setupWorkerExists && syncBuildExists) {
                resultDiv.innerHTML = '<div class="success">✅ Step 2 completed! Installation verified.</div>';
                currentStep = 3;
                updateStepStatuses();
                saveProgress();
            } else {
                resultDiv.innerHTML = '<div class="error">❌ Installation incomplete. Please check file permissions and try again.</div>';
            }
        } catch (error) {
            resultDiv.innerHTML = `<div class="error">❌ Validation failed: ${error.message}</div>`;
        }
    }

    async function testSSHConnections() {
        const worker1 = document.getElementById('worker1-host').value.trim();
        const worker2 = document.getElementById('worker2-host').value.trim();
        const resultDiv = document.getElementById('ssh-test-result');

        if (!worker1 || !worker2) {
            resultDiv.innerHTML = '<div class="error">❌ Please enter both worker hostnames.</div>';
            return;
        }

        resultDiv.innerHTML = '<div class="checking">Testing SSH connections...</div>';

        try {
            // Simulate SSH testing
            await new Promise(resolve => setTimeout(resolve, 3000));

            // Mock results - in real implementation, test actual SSH connections
            tutorialData.worker1Host = worker1;
            tutorialData.worker2Host = worker2;

            document.getElementById('worker1-host-display').textContent = worker1;
            document.getElementById('worker2-host-display').textContent = worker2;

            resultDiv.innerHTML = `
                <div class="success">
                    ✅ SSH connections successful!<br>
                    Worker 1 (${worker1}): Connected<br>
                    Worker 2 (${worker2}): Connected
                </div>
            `;

            saveProgress();
        } catch (error) {
            resultDiv.innerHTML = `<div class="error">❌ SSH test failed: ${error.message}</div>`;
        }
    }

    async function validateStep3() {
        const resultDiv = document.getElementById('step3-result');
        resultDiv.innerHTML = '<div class="checking">Validating SSH configuration...</div>';

        try {
            await new Promise(resolve => setTimeout(resolve, 2000));

            // Check if SSH test was completed
            if (tutorialData.worker1Host && tutorialData.worker2Host) {
                resultDiv.innerHTML = '<div class="success">✅ Step 3 completed! SSH configuration verified.</div>';
                currentStep = 4;
                updateStepStatuses();
                saveProgress();
            } else {
                resultDiv.innerHTML = '<div class="error">❌ Please test SSH connections before proceeding.</div>';
            }
        } catch (error) {
            resultDiv.innerHTML = `<div class="error">❌ Validation failed: ${error.message}</div>`;
        }
    }

    function browseProject() {
        // In a real implementation, this would open a file browser
        // For demo purposes, we'll just set a mock path
        const mockPath = '/home/user/my-gradle-project';
        document.getElementById('project-path').value = mockPath;
        tutorialData.projectPath = mockPath;
        saveProgress();
        validateProjectPath();
    }

    function validateProjectPath() {
        const path = document.getElementById('project-path').value.trim();
        if (!path) return;

        // Mock validation
        document.getElementById('build-gradle-check').textContent = '✅ Found';
        document.getElementById('build-gradle-check').className = 'success';

        document.getElementById('src-dir-check').textContent = '✅ Found';
        document.getElementById('src-dir-check').className = 'success';

        document.getElementById('gradle-project-check').textContent = '✅ Valid';
        document.getElementById('gradle-project-check').className = 'success';
    }

    async function validateStep4() {
        const resultDiv = document.getElementById('step4-result');
        resultDiv.innerHTML = '<div class="checking">Validating coordinator setup...</div>';

        try {
            await new Promise(resolve => setTimeout(resolve, 2500));

            // Mock coordinator validation
            document.getElementById('coordinator-service-check').textContent = '✅ Running';
            document.getElementById('coordinator-service-check').className = 'success';

            document.getElementById('coordinator-config-check').textContent = '✅ Created';
            document.getElementById('coordinator-config-check').className = 'success';

            document.getElementById('coordinator-cache-check').textContent = '✅ Created';
            document.getElementById('coordinator-cache-check').className = 'success';

            resultDiv.innerHTML = '<div class="success">✅ Step 4 completed! Coordinator setup verified.</div>';
            currentStep = 5;
            updateStepStatuses();
            saveProgress();
        } catch (error) {
            resultDiv.innerHTML = `<div class="error">❌ Validation failed: ${error.message}</div>`;
        }
    }

    async function checkWorkers() {
        const resultDiv = document.getElementById('worker-check-result');
        resultDiv.innerHTML = '<div class="checking">Checking worker registration...</div>';

        try {
            await new Promise(resolve => setTimeout(resolve, 2000));

            // Mock worker check
            document.getElementById('worker1-status').textContent = '✅ Registered';
            document.getElementById('worker1-status').className = 'worker-status success';

            document.getElementById('worker2-status').textContent = '✅ Registered';
            document.getElementById('worker2-status').className = 'worker-status success';

            resultDiv.innerHTML = `
                <div class="success">
                    ✅ Workers registered successfully!<br>
                    Worker 1: Active and ready<br>
                    Worker 2: Active and ready
                </div>
            `;
        } catch (error) {
            resultDiv.innerHTML = `<div class="error">❌ Worker check failed: ${error.message}</div>`;
        }
    }

    async function validateStep5() {
        const resultDiv = document.getElementById('step5-result');
        resultDiv.innerHTML = '<div class="checking">Validating worker setup...</div>';

        try {
            await new Promise(resolve => setTimeout(resolve, 2000));

            // Check if workers are registered
            const worker1Registered = document.getElementById('worker1-status').textContent.includes('✅');
            const worker2Registered = document.getElementById('worker2-status').textContent.includes('✅');

            if (worker1Registered && worker2Registered) {
                resultDiv.innerHTML = '<div class="success">✅ Step 5 completed! Worker setup verified.</div>';
                currentStep = 6;
                updateStepStatuses();
                saveProgress();
            } else {
                resultDiv.innerHTML = '<div class="error">❌ Please register workers before proceeding.</div>';
            }
        } catch (error) {
            resultDiv.innerHTML = `<div class="error">❌ Validation failed: ${error.message}</div>`;
        }
    }

    async function runTraditionalBuild() {
        const resultDiv = document.getElementById('traditional-result');
        resultDiv.innerHTML = '<div class="running">🏃 Running traditional build...</div>';

        try {
            // Simulate traditional build (slower)
            await new Promise(resolve => setTimeout(resolve, 8000));

            const buildTime = 45 + Math.random() * 10; // 45-55 seconds
            resultDiv.innerHTML = `
                <div class="build-complete">
                    ✅ Traditional build completed<br>
                    ⏱️ Build time: ${Math.round(buildTime)}s<br>
                    📦 Tasks: 45 completed<br>
                    🚀 Single-threaded execution
                </div>
            `;

            // Store result for comparison
            tutorialData.traditionalTime = buildTime;
            saveProgress();

        } catch (error) {
            resultDiv.innerHTML = `<div class="error">❌ Traditional build failed: ${error.message}</div>`;
        }
    }

    async function runDistributedBuild() {
        const resultDiv = document.getElementById('distributed-result');
        resultDiv.innerHTML = '<div class="running">🚀 Running distributed build...</div>';

        try {
            // Simulate distributed build (faster)
            await new Promise(resolve => setTimeout(resolve, 3000));

            const buildTime = 12 + Math.random() * 6; // 12-18 seconds
            const speedup = tutorialData.traditionalTime ? (tutorialData.traditionalTime / buildTime).toFixed(1) : '3.2';

            resultDiv.innerHTML = `
                <div class="build-complete">
                    ✅ Distributed build completed<br>
                    ⏱️ Build time: ${Math.round(buildTime)}s<br>
                    📦 Tasks: 45 completed (parallel)<br>
                    🚀 ${speedup}x speedup achieved!<br>
                    💾 Cache hit rate: 78%<br>
                    👥 Workers used: 2
                </div>
            `;

            // Store result and create comparison chart
            tutorialData.distributedTime = buildTime;
            saveProgress();
            createBuildComparisonChart();

        } catch (error) {
            resultDiv.innerHTML = `<div class="error">❌ Distributed build failed: ${error.message}</div>`;
        }
    }

    function createBuildComparisonChart() {
        const canvas = document.getElementById('build-comparison-chart');
        if (!canvas) return;

        const ctx = canvas.getContext('2d');
        const traditionalTime = tutorialData.traditionalTime || 45;
        const distributedTime = tutorialData.distributedTime || 15;

        // Clear canvas
        ctx.clearRect(0, 0, canvas.width, canvas.height);

        // Draw bars
        const barWidth = 60;
        const spacing = 120;
        const maxHeight = 200;
        const scale = maxHeight / Math.max(traditionalTime, distributedTime);

        // Traditional build bar (red)
        ctx.fillStyle = '#dc3545';
        ctx.fillRect(80, canvas.height - (traditionalTime * scale) - 60, barWidth, traditionalTime * scale);

        // Distributed build bar (green)
        ctx.fillStyle = '#28a745';
        ctx.fillRect(80 + spacing, canvas.height - (distributedTime * scale) - 60, barWidth, distributedTime * scale);

        // Labels
        ctx.fillStyle = '#333';
        ctx.font = '14px Arial';
        ctx.textAlign = 'center';
        ctx.fillText('Traditional', 80 + barWidth/2, canvas.height - 30);
        ctx.fillText('Distributed', 80 + spacing + barWidth/2, canvas.height - 30);

        // Values
        ctx.font = '12px Arial';
        ctx.fillText(`${Math.round(traditionalTime)}s`, 80 + barWidth/2, canvas.height - (traditionalTime * scale) - 65);
        ctx.fillText(`${Math.round(distributedTime)}s`, 80 + spacing + barWidth/2, canvas.height - (distributedTime * scale) - 65);

        // Title
        ctx.font = '16px Arial';
        ctx.fillText(`Build Performance: ${(traditionalTime/distributedTime).toFixed(1)}x Faster!`, canvas.width/2, 25);
    }

    async function completeTutorial() {
        const resultDiv = document.getElementById('step6-result');
        resultDiv.innerHTML = '<div class="checking">Completing tutorial...</div>';

        try {
            await new Promise(resolve => setTimeout(resolve, 2000));

            // Show completion
            document.getElementById('tutorial-complete').style.display = 'block';
            resultDiv.innerHTML = '<div class="success">🎉 Tutorial completed successfully!</div>';

            // Clear saved progress
            localStorage.removeItem('distributed-gradle-tutorial');

        } catch (error) {
            resultDiv.innerHTML = `<div class="error">❌ Completion failed: ${error.message}</div>`;
        }
    }

    function downloadCertificate() {
        // In a real implementation, this would generate and download a certificate
        alert('Certificate download feature would be implemented here!\n\nYou have successfully completed the Distributed Gradle Setup Tutorial.');
    }

    // Auto-validate project path when changed
    document.getElementById('project-path').addEventListener('input', validateProjectPath);
});

// Add CSS styles for the interactive tutorial
const style = document.createElement('style');
style.textContent = `
    .tutorial-step {
        margin-bottom: 30px;
        border: 1px solid #dee2e6;
        border-radius: 8px;
        overflow: hidden;
    }

    .step-header {
        background: #f8f9fa;
        padding: 15px 20px;
        border-bottom: 1px solid #dee2e6;
        display: flex;
        justify-content: space-between;
        align-items: center;
    }

    .step-number {
        background: #007bff;
        color: white;
        width: 30px;
        height: 30px;
        border-radius: 50%;
        display: flex;
        align-items: center;
        justify-content: center;
        font-weight: bold;
        margin-right: 15px;
    }

    .step-header h3 {
        margin: 0;
        color: #495057;
    }

    .step-status {
        font-weight: bold;
        padding: 5px 10px;
        border-radius: 4px;
        font-size: 14px;
    }

    .step-status.completed {
        background: #d4edda;
        color: #155724;
    }

    .step-status.in-progress {
        background: #fff3cd;
        color: #856404;
    }

    .step-status.locked {
        background: #e2e3e5;
        color: #6c757d;
    }

    .step-content {
        padding: 20px;
    }

    .requirements-list, .file-verification, .coordinator-verification {
        margin: 20px 0;
    }

    .requirement, .file-check, .validation-item, .status-item {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 10px 0;
        border-bottom: 1px solid #f0f0f0;
    }

    .requirement:last-child, .file-check:last-child,
    .validation-item:last-child, .status-item:last-child {
        border-bottom: none;
    }

    .check-btn {
        background: #6c757d;
        color: white;
        border: none;
        padding: 5px 10px;
        border-radius: 4px;
        cursor: pointer;
        font-size: 12px;
    }

    .check-btn:hover {
        background: #5a6268;
    }

    .check-result {
        font-weight: bold;
    }

    .check-result.success {
        color: #28a745;
    }

    .check-result.error {
        color: #dc3545;
    }

    .code-example {
        background: #f8f9fa;
        border: 1px solid #dee2e6;
        border-radius: 4px;
        padding: 15px;
        margin: 15px 0;
    }

    .code-example h4 {
        margin: 0 0 10px 0;
        color: #495057;
    }

    .code-example pre {
        margin: 0;
        background: #ffffff;
        padding: 10px;
        border-radius: 4px;
        overflow-x: auto;
    }

    .worker-input {
        margin: 10px 0;
    }

    .worker-input label {
        display: block;
        margin-bottom: 5px;
        font-weight: 600;
        color: #495057;
    }

    .worker-input input {
        width: 100%;
        padding: 8px 12px;
        border: 1px solid #ced4da;
        border-radius: 4px;
        font-size: 14px;
    }

    .project-input {
        display: flex;
        gap: 10px;
        align-items: flex-end;
    }

    .project-input input {
        flex: 1;
        padding: 8px 12px;
        border: 1px solid #ced4da;
        border-radius: 4px;
        font-size: 14px;
    }

    .worker-list {
        margin: 20px 0;
    }

    .worker-list h4 {
        margin: 0 0 15px 0;
        color: #495057;
    }

    .worker-item {
        display: flex;
        align-items: center;
        padding: 10px 0;
        border-bottom: 1px solid #f0f0f0;
    }

    .worker-item:last-child {
        border-bottom: none;
    }

    .worker-item input {
        margin-right: 10px;
    }

    .worker-item label {
        flex: 1;
        margin: 0;
    }

    .worker-status {
        font-weight: bold;
        padding: 3px 8px;
        border-radius: 4px;
        font-size: 12px;
    }

    .worker-status.success {
        background: #d4edda;
        color: #155724;
    }

    .build-comparison {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 20px;
        margin: 20px 0;
    }

    .build-option {
        background: #f8f9fa;
        border: 1px solid #dee2e6;
        border-radius: 8px;
        padding: 20px;
    }

    .build-option h5 {
        margin: 0 0 15px 0;
        color: #495057;
    }

    .build-result {
        margin-top: 15px;
        padding: 10px;
        border-radius: 4px;
        font-size: 14px;
    }

    .build-result .running {
        background: #fff3cd;
        color: #856404;
    }

    .build-result .build-complete {
        background: #d4edda;
        color: #155724;
    }

    .build-result .error {
        background: #f8d7da;
        color: #721c24;
    }

    .performance-chart {
        background: white;
        border: 1px solid #dee2e6;
        border-radius: 8px;
        padding: 20px;
        margin: 20px 0;
        text-align: center;
    }

    .performance-chart h4 {
        margin: 0 0 20px 0;
        color: #495057;
    }

    .step-actions {
        margin-top: 20px;
        text-align: center;
    }

    .btn-secondary {
        background: #6c757d;
        color: white;
        border: none;
        padding: 10px 20px;
        border-radius: 4px;
        cursor: pointer;
        font-size: 14px;
        font-weight: 600;
    }

    .btn-secondary:hover {
        background: #5a6268;
    }

    .btn-secondary:disabled {
        background: #6c757d;
        opacity: 0.6;
        cursor: not-allowed;
    }

    .validation-result {
        margin-top: 15px;
        padding: 10px;
        border-radius: 4px;
        font-weight: 500;
    }

    .validation-result .checking {
        background: #fff3cd;
        color: #856404;
    }

    .validation-result .success {
        background: #d4edda;
        color: #155724;
    }

    .validation-result .error {
        background: #f8d7da;
        color: #721c24;
    }

    .tutorial-completion {
        background: #d4edda;
        border: 1px solid #c3e6cb;
        border-radius: 8px;
        padding: 30px;
        text-align: center;
        margin: 30px 0;
    }

    .tutorial-completion h2 {
        color: #155724;
        margin: 0 0 20px 0;
    }

    .completion-summary, .next-steps {
        text-align: left;
        max-width: 600px;
        margin: 0 auto 20px auto;
    }

    .completion-summary ul, .next-steps ul {
        padding-left: 20px;
    }

    .completion-summary li, .next-steps li {
        margin: 8px 0;
    }

    .certificate {
        background: white;
        border: 2px solid #28a745;
        border-radius: 8px;
        padding: 20px;
        margin: 20px auto;
        max-width: 400px;
    }

    .certificate h3 {
        color: #28a745;
        margin: 0 0 15px 0;
    }

    .btn-primary {
        background: #007bff;
        color: white;
        border: none;
        padding: 12px 24px;
        border-radius: 6px;
        cursor: pointer;
        font-size: 16px;
        font-weight: 600;
        text-decoration: none;
        display: inline-block;
    }

    .btn-primary:hover {
        background: #0056b3;
    }

    .troubleshooting {
        margin-top: 30px;
    }

    .issue {
        background: #fff3cd;
        border: 1px solid #ffeaa7;
        border-radius: 4px;
        padding: 15px;
        margin: 10px 0;
    }

    .issue h4 {
        margin: 0 0 8px 0;
        color: #856404;
    }

    .issue p {
        margin: 0;
        color: #856404;
    }

    @media (max-width: 768px) {
        .build-comparison {
            grid-template-columns: 1fr;
        }

        .step-header {
            flex-direction: column;
            align-items: flex-start;
            gap: 10px;
        }

        .project-input {
            flex-direction: column;
            align-items: stretch;
        }

        .project-input input {
            margin-bottom: 10px;
        }
    }
`;
document.head.appendChild(style);