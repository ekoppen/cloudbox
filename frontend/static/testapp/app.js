// CloudBox BaaS Test App JavaScript

// Global configuration
let config = {
    backendUrl: 'http://localhost:8080',
    projectSlug: '',
    apiKey: '',
    authToken: ''
};

// Utility functions
function log(message, type = 'info') {
    const logContainer = document.getElementById('response-log');
    const timestamp = new Date().toLocaleTimeString();
    const logEntry = document.createElement('div');
    
    let bgColor = 'bg-blue-50 border-blue-200 text-blue-800';
    if (type === 'success') bgColor = 'bg-green-50 border-green-200 text-green-800';
    if (type === 'error') bgColor = 'bg-red-50 border-red-200 text-red-800';
    if (type === 'warning') bgColor = 'bg-yellow-50 border-yellow-200 text-yellow-800';
    
    logEntry.className = `p-3 border rounded-md ${bgColor}`;
    logEntry.innerHTML = `
        <div class="flex justify-between items-start">
            <div class="flex-1">
                <span class="text-xs font-mono">[${timestamp}]</span>
                <pre class="mt-1 text-sm whitespace-pre-wrap">${message}</pre>
            </div>
        </div>
    `;
    
    logContainer.insertBefore(logEntry, logContainer.firstChild);
}

function clearLog() {
    document.getElementById('response-log').innerHTML = '';
}

function showTab(tabName) {
    // Hide all tabs
    document.querySelectorAll('.tab-content').forEach(tab => {
        tab.classList.remove('active');
    });
    document.querySelectorAll('.tab-button').forEach(button => {
        button.classList.remove('active');
    });
    
    // Show selected tab
    document.getElementById(tabName).classList.add('active');
    event.target.classList.add('active');
}

function getApiUrl(endpoint) {
    const baseUrl = document.getElementById('backend-url').value;
    const slug = document.getElementById('project-slug').value;
    
    if (endpoint.startsWith('/p/')) {
        return `${baseUrl}${endpoint.replace(':project_slug', slug)}`;
    }
    return `${baseUrl}${endpoint}`;
}

function getHeaders(includeAuth = false) {
    const headers = {
        'Content-Type': 'application/json'
    };
    
    const apiKey = document.getElementById('api-key').value;
    if (apiKey) {
        headers['X-API-Key'] = apiKey;
    }
    
    if (includeAuth && config.authToken) {
        headers['Authorization'] = `Bearer ${config.authToken}`;
    }
    
    return headers;
}

async function makeRequest(url, options = {}) {
    try {
        log(`Making request to: ${url}`, 'info');
        log(`Request options: ${JSON.stringify(options, null, 2)}`, 'info');
        
        const response = await fetch(url, options);
        const data = await response.json();
        
        if (response.ok) {
            log(`Success: ${JSON.stringify(data, null, 2)}`, 'success');
            return { success: true, data };
        } else {
            log(`Error: ${response.status} - ${JSON.stringify(data, null, 2)}`, 'error');
            return { success: false, error: data };
        }
    } catch (error) {
        log(`Network Error: ${error.message}`, 'error');
        return { success: false, error: error.message };
    }
}

// Connection testing
async function testConnection() {
    const url = getApiUrl('/api/v1/health');
    const result = await makeRequest(url);
    
    const statusElement = document.getElementById('status-indicator');
    if (result.success) {
        statusElement.textContent = 'Connected ✅';
        statusElement.className = 'px-3 py-1 rounded-full text-sm bg-green-100 text-green-800';
    } else {
        statusElement.textContent = 'Disconnected ❌';
        statusElement.className = 'px-3 py-1 rounded-full text-sm bg-red-100 text-red-800';
    }
}

// Database functions
async function createData() {
    const tableName = document.getElementById('table-name').value;
    const jsonData = document.getElementById('json-data').value;
    
    if (!tableName || !jsonData) {
        log('Please fill in table name and JSON data', 'warning');
        return;
    }
    
    try {
        const data = JSON.parse(jsonData);
        const url = getApiUrl(`/p/:project_slug/api/data/${tableName}`);
        
        const result = await makeRequest(url, {
            method: 'POST',
            headers: getHeaders(true),
            body: JSON.stringify(data)
        });
        
        if (result.success) {
            document.getElementById('json-data').value = '';
        }
    } catch (error) {
        log(`Invalid JSON: ${error.message}`, 'error');
    }
}

async function readData() {
    const tableName = document.getElementById('read-table-name').value;
    
    if (!tableName) {
        log('Please enter table name', 'warning');
        return;
    }
    
    const url = getApiUrl(`/p/:project_slug/api/data/${tableName}`);
    const result = await makeRequest(url, {
        method: 'GET',
        headers: getHeaders(true)
    });
    
    if (result.success) {
        document.getElementById('data-results').innerHTML = `<pre>${JSON.stringify(result.data, null, 2)}</pre>`;
    }
}

// Storage functions
async function createBucket() {
    const bucketName = document.getElementById('bucket-name').value;
    
    if (!bucketName) {
        log('Please enter bucket name', 'warning');
        return;
    }
    
    const url = getApiUrl(`/p/:project_slug/api/storage/buckets`);
    const result = await makeRequest(url, {
        method: 'POST',
        headers: getHeaders(true),
        body: JSON.stringify({ 
            name: bucketName,
            permissions: ['read', 'write'],
            fileSizeLimit: 10485760, // 10MB
            allowedFileTypes: ['image/*', 'application/pdf', 'text/*']
        })
    });
}

async function listBuckets() {
    const url = getApiUrl(`/p/:project_slug/api/storage/buckets`);
    const result = await makeRequest(url, {
        method: 'GET',
        headers: getHeaders(true)
    });
    
    if (result.success) {
        const buckets = result.data.buckets || [];
        document.getElementById('bucket-results').innerHTML = buckets.length > 0 
            ? `<pre>${JSON.stringify(buckets, null, 2)}</pre>`
            : '<p class="text-gray-500">No buckets found</p>';
    }
}

async function uploadFile() {
    const fileInput = document.getElementById('file-input');
    const bucketName = document.getElementById('upload-bucket').value;
    
    if (!fileInput.files[0] || !bucketName) {
        log('Please select a file and enter bucket name', 'warning');
        return;
    }
    
    const file = fileInput.files[0];
    const formData = new FormData();
    formData.append('file', file);
    formData.append('fileName', file.name);
    
    const url = getApiUrl(`/p/:project_slug/api/storage/buckets/${bucketName}/files`);
    
    try {
        log(`Uploading file: ${file.name} to bucket: ${bucketName}`, 'info');
        
        const headers = {};
        const apiKey = document.getElementById('api-key').value;
        if (apiKey) headers['X-API-Key'] = apiKey;
        if (config.authToken) headers['Authorization'] = `Bearer ${config.authToken}`;
        
        const response = await fetch(url, {
            method: 'POST',
            headers: headers,
            body: formData
        });
        
        const data = await response.json();
        
        if (response.ok) {
            log(`File uploaded successfully: ${JSON.stringify(data, null, 2)}`, 'success');
            fileInput.value = '';
        } else {
            log(`Upload failed: ${JSON.stringify(data, null, 2)}`, 'error');
        }
    } catch (error) {
        log(`Upload error: ${error.message}`, 'error');
    }
}

async function listFiles() {
    const bucketName = document.getElementById('upload-bucket').value;
    
    if (!bucketName) {
        log('Please enter bucket name', 'warning');
        return;
    }
    
    const url = getApiUrl(`/p/:project_slug/api/storage/buckets/${bucketName}/files`);
    const result = await makeRequest(url, {
        method: 'GET',
        headers: getHeaders(true)
    });
    
    if (result.success) {
        const files = result.data.files || [];
        document.getElementById('file-results').innerHTML = files.length > 0 
            ? `<pre>${JSON.stringify(files, null, 2)}</pre>`
            : '<p class="text-gray-500">No files found</p>';
    }
}

// Authentication functions
async function registerUser() {
    const email = document.getElementById('reg-email').value;
    const password = document.getElementById('reg-password').value;
    const name = document.getElementById('reg-name').value;
    
    if (!email || !password || !name) {
        log('Please fill in all registration fields', 'warning');
        return;
    }
    
    const url = getApiUrl('/api/v1/auth/register');
    const result = await makeRequest(url, {
        method: 'POST',
        headers: getHeaders(),
        body: JSON.stringify({ email, password, name })
    });
    
    if (result.success) {
        document.getElementById('reg-email').value = '';
        document.getElementById('reg-password').value = '';
        document.getElementById('reg-name').value = '';
        
        // Auto-fill login form
        document.getElementById('login-email').value = email;
    }
}

async function loginUser() {
    const email = document.getElementById('login-email').value;
    const password = document.getElementById('login-password').value;
    
    if (!email || !password) {
        log('Please fill in email and password', 'warning');
        return;
    }
    
    const url = getApiUrl('/api/v1/auth/login');
    const result = await makeRequest(url, {
        method: 'POST',
        headers: getHeaders(),
        body: JSON.stringify({ email, password })
    });
    
    if (result.success && result.data.token) {
        config.authToken = result.data.token;
        document.getElementById('auth-token').innerHTML = `
            <strong>Login Successful!</strong><br>
            <span class="text-xs">Token: ${result.data.token.substring(0, 50)}...</span><br>
            <span class="text-xs">User: ${result.data.user.name} (${result.data.user.email})</span>
        `;
        
        // Clear password
        document.getElementById('login-password').value = '';
    }
}

// Function execution
async function executeFunction() {
    const functionName = document.getElementById('function-name').value;
    const params = document.getElementById('function-params').value;
    
    if (!functionName) {
        log('Please enter function name', 'warning');
        return;
    }
    
    let functionParams = {};
    if (params) {
        try {
            functionParams = JSON.parse(params);
        } catch (error) {
            log(`Invalid JSON parameters: ${error.message}`, 'error');
            return;
        }
    }
    
    const url = getApiUrl(`/p/:project_slug/api/functions/${functionName}`);
    const result = await makeRequest(url, {
        method: 'POST',
        headers: getHeaders(true),
        body: JSON.stringify(functionParams)
    });
    
    if (result.success) {
        document.getElementById('function-results').innerHTML = `<pre>${JSON.stringify(result.data, null, 2)}</pre>`;
    }
}

// Messaging functions
async function sendEmail() {
    const to = document.getElementById('email-to').value;
    const subject = document.getElementById('email-subject').value;
    const message = document.getElementById('email-message').value;
    
    if (!to || !subject || !message) {
        log('Please fill in all email fields', 'warning');
        return;
    }
    
    const url = getApiUrl(`/p/:project_slug/api/messaging/email`);
    const result = await makeRequest(url, {
        method: 'POST',
        headers: getHeaders(true),
        body: JSON.stringify({
            to: [to],
            subject,
            content: message,
            template: 'default'
        })
    });
    
    if (result.success) {
        document.getElementById('email-message').value = '';
    }
}

async function sendPushNotification() {
    const title = document.getElementById('push-title').value;
    const body = document.getElementById('push-body').value;
    const target = document.getElementById('push-target').value;
    
    if (!title || !body || !target) {
        log('Please fill in all push notification fields', 'warning');
        return;
    }
    
    const url = getApiUrl(`/p/:project_slug/api/messaging/push`);
    const result = await makeRequest(url, {
        method: 'POST',
        headers: getHeaders(true),
        body: JSON.stringify({
            title,
            body,
            target,
            data: { source: 'test-app' }
        })
    });
    
    if (result.success) {
        document.getElementById('push-body').value = '';
    }
}

// Initialize app
document.addEventListener('DOMContentLoaded', function() {
    // Set initial values
    config.backendUrl = document.getElementById('backend-url').value;
    
    // Test connection on load
    testConnection();
    
    // Update config when inputs change
    document.getElementById('backend-url').addEventListener('change', function() {
        config.backendUrl = this.value;
    });
    
    document.getElementById('project-slug').addEventListener('change', function() {
        config.projectSlug = this.value;
    });
    
    document.getElementById('api-key').addEventListener('change', function() {
        config.apiKey = this.value;
    });
    
    log('CloudBox Test App initialized. Ready to test your BaaS!', 'success');
});