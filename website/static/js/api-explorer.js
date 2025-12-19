// API Explorer JavaScript
document.addEventListener('DOMContentLoaded', function() {
    const apiEndpoint = document.getElementById('api-endpoint');
    const authToken = document.getElementById('auth-token');
    const testConnectionBtn = document.getElementById('test-connection');
    const connectionStatus = document.getElementById('connection-status');

    // Test connection functionality
    testConnectionBtn.addEventListener('click', async function() {
        const endpoint = apiEndpoint.value.trim();
        if (!endpoint) {
            showStatus('Please enter an API endpoint', 'error');
            return;
        }

        showStatus('Testing connection...', 'info');
        testConnectionBtn.disabled = true;

        try {
            const response = await fetch(`${endpoint}/health`, {
                method: 'GET',
                headers: getHeaders()
            });

            if (response.ok) {
                const data = await response.json();
                showStatus(`✅ Connection successful! Server is healthy.`, 'success');
            } else {
                showStatus(`❌ Connection failed: ${response.status} ${response.statusText}`, 'error');
            }
        } catch (error) {
            showStatus(`❌ Connection failed: ${error.message}`, 'error');
        } finally {
            testConnectionBtn.disabled = false;
        }
    });

    // Add event listeners to all test endpoint buttons
    document.querySelectorAll('.test-endpoint').forEach(button => {
        button.addEventListener('click', async function() {
            const endpoint = this.dataset.endpoint;
            const method = this.dataset.method;
            const bodyElementId = this.dataset.body;
            const dynamicElementId = this.dataset.dynamic;
            const responseElementId = this.id.replace('test-endpoint', 'response');

            await testEndpoint(endpoint, method, bodyElementId, dynamicElementId, responseElementId);
        });
    });

    // Helper function to get headers
    function getHeaders() {
        const headers = {
            'Content-Type': 'application/json'
        };

        const token = authToken.value.trim();
        if (token) {
            headers['X-Auth-Token'] = token;
        }

        return headers;
    }

    // Helper function to show status messages
    function showStatus(message, type) {
        connectionStatus.className = `status-message ${type}`;
        connectionStatus.textContent = message;
        connectionStatus.style.display = 'block';

        // Auto-hide success messages after 5 seconds
        if (type === 'success') {
            setTimeout(() => {
                connectionStatus.style.display = 'none';
            }, 5000);
        }
    }

    // Function to test API endpoints
    async function testEndpoint(endpoint, method, bodyElementId, dynamicElementId, responseElementId) {
        const baseUrl = apiEndpoint.value.trim();
        if (!baseUrl) {
            alert('Please enter an API endpoint first');
            return;
        }

        let url = `${baseUrl}${endpoint}`;
        let body = null;

        // Handle dynamic endpoints (like /api/build/{id})
        if (dynamicElementId) {
            const dynamicValue = document.getElementById(dynamicElementId).value.trim();
            if (!dynamicValue) {
                alert(`Please enter a value for ${dynamicElementId}`);
                return;
            }
            url = url.replace('{build_id}', dynamicValue);
        }

        // Handle request bodies
        if (bodyElementId && method === 'POST') {
            const bodyElement = document.getElementById(bodyElementId);
            const bodyText = bodyElement.value.trim();
            if (!bodyText) {
                alert('Please enter a request body');
                return;
            }
            try {
                body = JSON.parse(bodyText);
            } catch (e) {
                alert('Invalid JSON in request body');
                return;
            }
        }

        const responseElement = document.getElementById(responseElementId);
        responseElement.innerHTML = '<div class="loading">Testing endpoint...</div>';

        try {
            const response = await fetch(url, {
                method: method,
                headers: getHeaders(),
                body: body ? JSON.stringify(body) : null
            });

            const responseText = await response.text();
            let responseData;

            try {
                responseData = JSON.parse(responseText);
            } catch (e) {
                responseData = responseText;
            }

            const statusClass = response.ok ? 'success' : 'error';
            const statusText = response.ok ? '✅ Success' : `❌ Error ${response.status}`;

            responseElement.innerHTML = `
                <div class="response-header ${statusClass}">
                    <strong>${statusText}</strong> - ${method} ${url}
                </div>
                <div class="response-body">
                    <pre>${JSON.stringify(responseData, null, 2)}</pre>
                </div>
            `;

        } catch (error) {
            responseElement.innerHTML = `
                <div class="response-header error">
                    <strong>❌ Network Error</strong> - ${method} ${url}
                </div>
                <div class="response-body">
                    <pre>${error.message}</pre>
                </div>
            `;
        }
    }

    // Initialize with default values
    showStatus('Ready to test API endpoints. Configure your endpoint and click "Test Connection".', 'info');
});

// Add some CSS styles for the API explorer
const style = document.createElement('style');
style.textContent = `
    #api-explorer {
        font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
        max-width: 1200px;
        margin: 0 auto;
        padding: 20px;
    }

    .api-controls {
        background: #f8f9fa;
        padding: 20px;
        border-radius: 8px;
        margin-bottom: 30px;
        border: 1px solid #dee2e6;
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
        margin-right: 10px;
    }

    .btn-secondary {
        background: #6c757d;
    }

    .btn-primary:hover, .btn-secondary:hover {
        opacity: 0.9;
    }

    .btn-primary:disabled, .btn-secondary:disabled {
        opacity: 0.6;
        cursor: not-allowed;
    }

    .status-message {
        padding: 12px 16px;
        border-radius: 4px;
        margin: 15px 0;
        font-weight: 500;
    }

    .status-message.success {
        background: #d4edda;
        color: #155724;
        border: 1px solid #c3e6cb;
    }

    .status-message.error {
        background: #f8d7da;
        color: #721c24;
        border: 1px solid #f5c6cb;
    }

    .status-message.info {
        background: #d1ecf1;
        color: #0c5460;
        border: 1px solid #bee5eb;
    }

    .api-endpoints {
        display: grid;
        gap: 20px;
    }

    .endpoint-card {
        background: white;
        border: 1px solid #dee2e6;
        border-radius: 8px;
        padding: 20px;
        box-shadow: 0 2px 4px rgba(0,0,0,0.1);
    }

    .endpoint-card h4 {
        margin: 0 0 10px 0;
        color: #495057;
    }

    .method {
        display: inline-block;
        background: #007bff;
        color: white;
        padding: 2px 8px;
        border-radius: 4px;
        font-size: 12px;
        font-weight: bold;
        margin-right: 10px;
    }

    .endpoint-path {
        display: inline-block;
        font-family: 'Courier New', monospace;
        background: #f8f9fa;
        padding: 2px 8px;
        border-radius: 4px;
        color: #495057;
    }

    .endpoint-card p {
        color: #6c757d;
        margin: 10px 0;
    }

    .request-body {
        margin: 15px 0;
    }

    .request-body label {
        display: block;
        margin-bottom: 5px;
        font-weight: 600;
        color: #495057;
    }

    .request-body textarea, .request-body input {
        width: 100%;
        padding: 8px 12px;
        border: 1px solid #ced4da;
        border-radius: 4px;
        font-family: 'Courier New', monospace;
        font-size: 14px;
    }

    .request-body textarea {
        min-height: 120px;
        resize: vertical;
    }

    .response {
        margin-top: 15px;
        border: 1px solid #dee2e6;
        border-radius: 4px;
        overflow: hidden;
    }

    .response-header {
        padding: 10px 15px;
        font-weight: 600;
        border-bottom: 1px solid #dee2e6;
    }

    .response-header.success {
        background: #d4edda;
        color: #155724;
    }

    .response-header.error {
        background: #f8d7da;
        color: #721c24;
    }

    .response-body {
        background: #f8f9fa;
    }

    .response-body pre {
        margin: 0;
        padding: 15px;
        overflow-x: auto;
        font-size: 13px;
        line-height: 1.4;
    }

    .loading {
        padding: 20px;
        text-align: center;
        color: #6c757d;
        font-style: italic;
    }

    @media (max-width: 768px) {
        #api-explorer {
            padding: 10px;
        }

        .api-controls, .endpoint-card {
            padding: 15px;
        }

        .btn-primary, .btn-secondary {
            width: 100%;
            margin-bottom: 10px;
        }
    }
`;
document.head.appendChild(style);