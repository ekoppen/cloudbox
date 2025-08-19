const https = require('https');
const http = require('http');

// Test configuration
const TEST_CONFIG = {
    baseUrl: 'http://localhost:8080',
    adminCredentials: {
        email: 'admin@cloudbox.dev',
        password: 'admin123'
    }
};

// Global variable to store auth token
let authToken = null;

// Helper function to make HTTP requests
function makeRequest(url, options = {}) {
    return new Promise((resolve, reject) => {
        const urlObj = new URL(url);
        const isHttps = urlObj.protocol === 'https:';
        const httpModule = isHttps ? https : http;
        
        const requestOptions = {
            hostname: urlObj.hostname,
            port: urlObj.port || (isHttps ? 443 : 80),
            path: urlObj.pathname + urlObj.search,
            method: options.method || 'GET',
            headers: {
                'Content-Type': 'application/json',
                ...(authToken ? { 'Authorization': `Bearer ${authToken}` } : {}),
                ...options.headers
            }
        };

        const req = httpModule.request(requestOptions, (res) => {
            let data = '';
            res.on('data', (chunk) => {
                data += chunk;
            });
            res.on('end', () => {
                try {
                    const response = {
                        status: res.statusCode,
                        statusText: res.statusMessage,
                        headers: res.headers,
                        data: data ? JSON.parse(data) : null
                    };
                    resolve(response);
                } catch (e) {
                    resolve({
                        status: res.statusCode,
                        statusText: res.statusMessage,
                        headers: res.headers,
                        data: data
                    });
                }
            });
        });

        req.on('error', (e) => {
            reject(e);
        });

        if (options.body) {
            req.write(options.body);
        }

        req.end();
    });
}

// Helper function to log test results
function logTest(testName, success, details = '') {
    const status = success ? '✅' : '❌';
    console.log(`${status} ${testName}`);
    if (details) {
        console.log(`   ${details}`);
    }
    console.log('');
}

// Test authentication
async function testAuth() {
    console.log('🔐 Testing Authentication...\n');
    
    try {
        const response = await makeRequest(`${TEST_CONFIG.baseUrl}/api/v1/auth/login`, {
            method: 'POST',
            body: JSON.stringify(TEST_CONFIG.adminCredentials)
        });
        
        console.log('Auth response:', JSON.stringify(response.data, null, 2));
        
        if (response.status === 200 && response.data && response.data.token) {
            authToken = response.data.token;
            logTest('Admin Login', true, `Token received: ${authToken ? 'Yes' : 'No'}, Role: ${response.data.user?.role}`);
            return true;
        } else {
            logTest('Admin Login', false, `Status: ${response.status}, Message: ${response.data?.error || 'No token in response'}`);
            return false;
        }
    } catch (error) {
        logTest('Admin Login', false, `Error: ${error.message}`);
        return false;
    }
}

// Test marketplace endpoints
async function testMarketplaceEndpoints() {
    console.log('🛒 Testing Marketplace Endpoints...\n');
    
    // Test marketplace listing
    try {
        const response = await makeRequest(`${TEST_CONFIG.baseUrl}/api/v1/admin/plugins/marketplace`);
        
        if (response.status === 200) {
            const plugins = response.data?.plugins || [];
            logTest('Marketplace List', true, `Found ${plugins.length} plugins`);
            if (plugins.length > 0) {
                console.log('   Sample plugin:', JSON.stringify(plugins[0], null, 2));
                console.log('');
            }
        } else {
            logTest('Marketplace List', false, `Status: ${response.status}, Error: ${response.data?.error}`);
        }
    } catch (error) {
        logTest('Marketplace List', false, `Error: ${error.message}`);
    }
    
    // Test marketplace search with no query
    try {
        const response = await makeRequest(`${TEST_CONFIG.baseUrl}/api/v1/admin/plugins/marketplace/search`);
        
        if (response.status === 200) {
            const plugins = response.data?.plugins || response.data?.results || [];
            logTest('Marketplace Search (no query)', true, `Found ${plugins.length} plugins`);
            console.log('   Search response structure:', Object.keys(response.data));
            console.log('');
        } else {
            logTest('Marketplace Search (no query)', false, `Status: ${response.status}, Error: ${response.data?.error}`);
        }
    } catch (error) {
        logTest('Marketplace Search (no query)', false, `Error: ${error.message}`);
    }
    
    // Test marketplace search with query
    try {
        const response = await makeRequest(`${TEST_CONFIG.baseUrl}/api/v1/admin/plugins/marketplace/search?q=script`);
        
        if (response.status === 200) {
            const plugins = response.data?.plugins || response.data?.results || [];
            logTest('Marketplace Search (query: "script")', true, `Found ${plugins.length} plugins`);
            if (plugins.length > 0) {
                console.log('   First result:', JSON.stringify(plugins[0], null, 2));
                console.log('');
            }
        } else {
            logTest('Marketplace Search (query: "script")', false, `Status: ${response.status}, Error: ${response.data?.error}`);
        }
    } catch (error) {
        logTest('Marketplace Search (query: "script")', false, `Error: ${error.message}`);
    }
}

// Test plugin endpoints
async function testPluginEndpoints() {
    console.log('🔌 Testing Plugin Endpoints...\n');
    
    // Test admin plugins list
    try {
        const response = await makeRequest(`${TEST_CONFIG.baseUrl}/api/v1/admin/plugins`);
        
        if (response.status === 200) {
            const plugins = response.data?.plugins || [];
            logTest('Admin Plugins List', true, `Found ${plugins.length} installed plugins`);
            if (plugins.length > 0) {
                console.log('   Installed plugins:', plugins.map(p => `${p.name} (${p.status})`));
                console.log('');
            }
        } else {
            logTest('Admin Plugins List', false, `Status: ${response.status}, Error: ${response.data?.error}`);
        }
    } catch (error) {
        logTest('Admin Plugins List', false, `Error: ${error.message}`);
    }
}

// Test project plugin endpoints
async function testProjectPluginEndpoints() {
    console.log('📁 Testing Project Plugin Endpoints...\n');
    
    // First, get list of projects
    try {
        const projectsResponse = await makeRequest(`${TEST_CONFIG.baseUrl}/api/v1/admin/projects`);
        
        if (projectsResponse.status === 200) {
            const projects = projectsResponse.data?.projects || [];
            logTest('Projects List', true, `Found ${projects.length} projects`);
            
            if (projects.length > 0) {
                const project = projects[0];
                console.log(`   Testing with project: ${project.name} (ID: ${project.id})`);
                console.log('');
                
                // Test project available plugins
                try {
                    const availableResponse = await makeRequest(`${TEST_CONFIG.baseUrl}/api/v1/projects/${project.id}/plugins/available`);
                    
                    if (availableResponse.status === 200) {
                        const plugins = availableResponse.data?.plugins || [];
                        logTest('Project Available Plugins', true, `Found ${plugins.length} available plugins`);
                    } else {
                        logTest('Project Available Plugins', false, `Status: ${availableResponse.status}, Error: ${availableResponse.data?.error}`);
                    }
                } catch (error) {
                    logTest('Project Available Plugins', false, `Error: ${error.message}`);
                }
                
                // Test project installed plugins
                try {
                    const installedResponse = await makeRequest(`${TEST_CONFIG.baseUrl}/api/v1/projects/${project.id}/plugins/installed`);
                    
                    if (installedResponse.status === 200) {
                        const plugins = installedResponse.data?.plugins || [];
                        logTest('Project Installed Plugins', true, `Found ${plugins.length} installed plugins`);
                        if (plugins.length > 0) {
                            console.log('   Installed:', plugins.map(p => `${p.name} (${p.status})`));
                            console.log('');
                        }
                    } else {
                        logTest('Project Installed Plugins', false, `Status: ${installedResponse.status}, Error: ${installedResponse.data?.error}`);
                    }
                } catch (error) {
                    logTest('Project Installed Plugins', false, `Error: ${error.message}`);
                }
            }
        } else {
            logTest('Projects List', false, `Status: ${projectsResponse.status}, Error: ${projectsResponse.data?.error}`);
        }
    } catch (error) {
        logTest('Projects List', false, `Error: ${error.message}`);
    }
}

// Test database contents
async function testDatabaseContents() {
    console.log('🗄️ Testing Database Contents...\n');
    
    // Check what's in the plugin_marketplace table by examining the actual API responses
    console.log('Checking marketplace table data through API calls...\n');
}

// Main test function
async function runTests() {
    console.log('🚀 CloudBox Plugin Marketplace Test Suite\n');
    console.log('==========================================\n');
    
    // Test authentication first
    const authSuccess = await testAuth();
    if (!authSuccess) {
        console.log('❌ Authentication failed. Cannot proceed with other tests.');
        return;
    }
    
    // Test all endpoints
    await testMarketplaceEndpoints();
    await testPluginEndpoints();
    await testProjectPluginEndpoints();
    await testDatabaseContents();
    
    console.log('🏁 Test Suite Complete\n');
}

// Run the tests
runTests().catch(console.error);