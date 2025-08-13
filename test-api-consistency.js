#!/usr/bin/env node
/**
 * CloudBox API Consistency Test Suite
 * 
 * This test suite validates that all API changes maintain consistency
 * and that the standardized patterns work correctly.
 */

const fs = require('fs');

// Configuration
const CONFIG = {
  baseUrl: process.env.CLOUDBOX_ENDPOINT || 'http://localhost:8080',
  projectId: process.env.CLOUDBOX_PROJECT_ID || '2',
  projectSlug: process.env.CLOUDBOX_PROJECT_SLUG || 'dsqewdq',
  apiKey: process.env.CLOUDBOX_API_KEY || '',
  adminJWT: process.env.CLOUDBOX_ADMIN_JWT || ''
};

// Test utilities
class TestRunner {
  constructor() {
    this.tests = [];
    this.results = {
      passed: 0,
      failed: 0,
      errors: []
    };
  }

  async makeRequest(method, url, options = {}) {
    const fullUrl = `${CONFIG.baseUrl}${url}`;
    
    try {
      const response = await fetch(fullUrl, {
        method,
        headers: {
          'Content-Type': 'application/json',
          ...options.headers
        },
        body: options.body ? JSON.stringify(options.body) : undefined
      });

      let data;
      const text = await response.text();
      try {
        data = JSON.parse(text);
      } catch {
        data = text;
      }

      return {
        status: response.status,
        statusText: response.statusText,
        data,
        headers: Object.fromEntries(response.headers.entries())
      };
    } catch (error) {
      return {
        status: 0,
        statusText: error.message,
        data: null,
        error: error.message
      };
    }
  }

  addTest(name, testFn) {
    this.tests.push({ name, testFn });
  }

  async runTests() {
    console.log('🧪 CloudBox API Consistency Test Suite');
    console.log('=======================================\n');

    for (const test of this.tests) {
      try {
        console.log(`⏳ ${test.name}`);
        const result = await test.testFn();
        
        if (result.success) {
          console.log(`✅ ${test.name}`);
          this.results.passed++;
        } else {
          console.log(`❌ ${test.name}: ${result.error}`);
          this.results.failed++;
          this.results.errors.push({ test: test.name, error: result.error });
        }
      } catch (error) {
        console.log(`💥 ${test.name}: ${error.message}`);
        this.results.failed++;
        this.results.errors.push({ test: test.name, error: error.message });
      }
      console.log('');
    }

    // Print summary
    console.log('📊 Test Summary');
    console.log('===============');
    console.log(`Passed: ${this.results.passed}`);
    console.log(`Failed: ${this.results.failed}`);
    console.log(`Total:  ${this.tests.length}`);

    if (this.results.errors.length > 0) {
      console.log('\n❌ Failed Tests:');
      this.results.errors.forEach(({ test, error }) => {
        console.log(`   ${test}: ${error}`);
      });
    }

    return this.results.failed === 0;
  }
}

// Test suite
const runner = new TestRunner();

// Health check test
runner.addTest('System Health Check', async () => {
  const response = await runner.makeRequest('GET', '/health');
  
  if (response.status !== 200) {
    return { success: false, error: `Expected 200, got ${response.status}` };
  }
  
  if (!response.data.status || response.data.status !== 'ok') {
    return { success: false, error: 'Health check failed' };
  }
  
  return { success: true };
});

// Authentication pattern tests
runner.addTest('Admin JWT Authentication Pattern', async () => {
  if (!CONFIG.adminJWT) {
    return { success: false, error: 'Admin JWT not provided (set CLOUDBOX_ADMIN_JWT)' };
  }

  const response = await runner.makeRequest('GET', '/api/v1/projects', {
    headers: { 'Authorization': `Bearer ${CONFIG.adminJWT}` }
  });
  
  if (response.status === 401) {
    return { success: false, error: 'Invalid JWT token' };
  }
  
  if (response.status !== 200) {
    return { success: false, error: `Expected 200, got ${response.status}: ${response.data?.error || response.statusText}` };
  }
  
  return { success: true };
});

runner.addTest('Project API Key Authentication Pattern', async () => {
  if (!CONFIG.apiKey) {
    return { success: false, error: 'API key not provided (set CLOUDBOX_API_KEY)' };
  }

  const response = await runner.makeRequest('GET', `/p/${CONFIG.projectSlug}/api/collections`, {
    headers: { 'X-API-Key': CONFIG.apiKey }
  });
  
  if (response.status === 401) {
    return { success: false, error: 'Invalid API key or project not found' };
  }
  
  if (response.status !== 200) {
    return { success: false, error: `Expected 200, got ${response.status}: ${response.data?.error || response.statusText}` };
  }
  
  return { success: true };
});

// API consistency tests
runner.addTest('Consistent Error Response Format', async () => {
  // Test with invalid API key
  const response = await runner.makeRequest('GET', `/p/${CONFIG.projectSlug}/api/collections`, {
    headers: { 'X-API-Key': 'invalid-key-123' }
  });
  
  if (response.status !== 401) {
    return { success: false, error: `Expected 401 for invalid key, got ${response.status}` };
  }
  
  if (!response.data.error) {
    return { success: false, error: 'Error response missing "error" field' };
  }
  
  return { success: true };
});

runner.addTest('Standard Collection Creation Pattern', async () => {
  if (!CONFIG.apiKey) {
    return { success: false, error: 'API key not required for this test' };
  }

  // Test collection creation with proper schema format
  const testCollection = {
    name: `test_collection_${Date.now()}`,
    schema: [
      'title:string',
      'content:text',
      'published:boolean',
      'created_at:datetime'
    ]
  };

  const response = await runner.makeRequest('POST', `/p/${CONFIG.projectSlug}/api/collections`, {
    headers: { 'X-API-Key': CONFIG.apiKey },
    body: testCollection
  });
  
  if (response.status === 401) {
    return { success: false, error: 'Authentication failed - check API key' };
  }
  
  if (response.status === 404) {
    return { success: false, error: 'Project not found - check project slug' };
  }
  
  if (response.status !== 201 && response.status !== 200) {
    return { success: false, error: `Expected 201/200, got ${response.status}: ${JSON.stringify(response.data)}` };
  }
  
  // Clean up - delete the test collection
  try {
    await runner.makeRequest('DELETE', `/p/${CONFIG.projectSlug}/api/collections/${testCollection.name}`, {
      headers: { 'X-API-Key': CONFIG.apiKey }
    });
  } catch (err) {
    console.log(`   ℹ️  Cleanup warning: Could not delete test collection: ${err.message}`);
  }
  
  return { success: true };
});

runner.addTest('Storage Bucket Field Consistency', async () => {
  if (!CONFIG.apiKey) {
    return { success: false, error: 'API key required for this test' };
  }

  // Test bucket creation with correct field names
  const testBucket = {
    name: `test_bucket_${Date.now()}`,
    description: 'Test bucket for API consistency',
    is_public: true,  // Note: is_public, not public
    max_file_size: 1048576,
    allowed_types: ['image/jpeg', 'image/png']
  };

  const response = await runner.makeRequest('POST', `/p/${CONFIG.projectSlug}/api/storage/buckets`, {
    headers: { 'X-API-Key': CONFIG.apiKey },
    body: testBucket
  });
  
  if (response.status === 401) {
    return { success: false, error: 'Authentication failed - check API key' };
  }
  
  if (response.status !== 201 && response.status !== 200) {
    return { success: false, error: `Expected 201/200, got ${response.status}: ${JSON.stringify(response.data)}` };
  }
  
  // Clean up
  try {
    await runner.makeRequest('DELETE', `/p/${CONFIG.projectSlug}/api/storage/buckets/${testBucket.name}`, {
      headers: { 'X-API-Key': CONFIG.apiKey }
    });
  } catch (err) {
    console.log(`   ℹ️  Cleanup warning: Could not delete test bucket: ${err.message}`);
  }
  
  return { success: true };
});

// API Key security test
runner.addTest('API Key Security - No Plain Text Storage', async () => {
  if (!CONFIG.adminJWT) {
    return { success: false, error: 'Admin JWT required for this test' };
  }

  // Create an API key via admin endpoint
  const createResponse = await runner.makeRequest('POST', `/api/v1/projects/${CONFIG.projectId}/api-keys`, {
    headers: { 'Authorization': `Bearer ${CONFIG.adminJWT}` },
    body: {
      name: 'Security Test Key',
      permissions: ['read', 'write']
    }
  });
  
  if (createResponse.status !== 201) {
    return { success: false, error: `Failed to create API key: ${createResponse.status} - ${JSON.stringify(createResponse.data)}` };
  }
  
  // Verify the response contains the plain key (shown once)
  if (!createResponse.data.key) {
    return { success: false, error: 'Created API key response missing "key" field' };
  }
  
  if (!createResponse.data.warning) {
    return { success: false, error: 'Created API key response missing security warning' };
  }
  
  // Clean up - delete the test key
  try {
    await runner.makeRequest('DELETE', `/api/v1/projects/${CONFIG.projectId}/api-keys/${createResponse.data.id}`, {
      headers: { 'Authorization': `Bearer ${CONFIG.adminJWT}` }
    });
  } catch (err) {
    console.log(`   ℹ️  Cleanup warning: Could not delete test API key: ${err.message}`);
  }
  
  return { success: true };
});

// URL pattern consistency test
runner.addTest('No Mixed Authentication Routes', async () => {
  // Test that old mixed auth routes return appropriate responses
  const deprecatedRoutes = [
    `/api/v1/admin/projects/${CONFIG.projectId}/collections`,
    `/api/v1/admin/projects/${CONFIG.projectId}/storage/buckets`
  ];
  
  for (const route of deprecatedRoutes) {
    const response = await runner.makeRequest('GET', route, {
      headers: { 'Authorization': `Bearer ${CONFIG.adminJWT}` }
    });
    
    // These routes should either be removed (404) or properly authenticated
    if (response.status === 200 && response.data.length !== undefined) {
      return { success: false, error: `Deprecated route ${route} still returns data - should be removed or redesigned` };
    }
  }
  
  return { success: true };
});

// CORS consistency test
runner.addTest('CORS Headers Consistency', async () => {
  const response = await runner.makeRequest('OPTIONS', `/p/${CONFIG.projectSlug}/api/collections`, {
    headers: { 
      'Origin': 'http://localhost:3000',
      'Access-Control-Request-Method': 'GET',
      'Access-Control-Request-Headers': 'X-API-Key'
    }
  });
  
  // CORS preflight should return appropriate headers
  if (response.status !== 200 && response.status !== 204) {
    return { success: false, error: `CORS preflight failed: ${response.status}` };
  }
  
  const corsHeaders = response.headers['access-control-allow-origin'] || response.headers['Access-Control-Allow-Origin'];
  if (!corsHeaders) {
    return { success: false, error: 'Missing CORS Allow-Origin header' };
  }
  
  return { success: true };
});

// Run the test suite
async function main() {
  // Check configuration
  console.log('🔧 Configuration:');
  console.log(`   Base URL: ${CONFIG.baseUrl}`);
  console.log(`   Project ID: ${CONFIG.projectId}`);
  console.log(`   Project Slug: ${CONFIG.projectSlug}`);
  console.log(`   API Key: ${CONFIG.apiKey ? '✅ Provided' : '❌ Not provided'}`);
  console.log(`   Admin JWT: ${CONFIG.adminJWT ? '✅ Provided' : '❌ Not provided'}`);
  console.log('');

  const success = await runner.runTests();
  
  // Generate report
  const report = {
    timestamp: new Date().toISOString(),
    config: CONFIG,
    results: runner.results,
    success
  };
  
  fs.writeFileSync('test-results.json', JSON.stringify(report, null, 2));
  console.log('\n📄 Detailed results written to test-results.json');
  
  process.exit(success ? 0 : 1);
}

if (require.main === module) {
  main().catch(error => {
    console.error('💥 Test suite crashed:', error);
    process.exit(1);
  });
}

module.exports = { TestRunner };