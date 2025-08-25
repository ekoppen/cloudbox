#!/usr/bin/env node

/**
 * Test PhotoPortfolio Storage Upload
 * 
 * This script tests the corrected storage upload endpoints to verify
 * that PhotoPortfolio can successfully upload images to CloudBox.
 */

const fs = require('fs');
const path = require('path');

// Configuration
const API_BASE = 'http://localhost:8080';
const PROJECT_SLUG = 'photoportfolio';  // Use slug, not ID
const PHOTOPORTFOLIO_ORIGIN = 'http://localhost:1234';

async function testCORSPreflight() {
    console.log('🧪 Testing CORS preflight for PhotoPortfolio origin...\n');
    
    const endpoints = [
        {
            name: 'List buckets',
            url: `${API_BASE}/p/${PROJECT_SLUG}/api/storage/buckets`,
            method: 'GET'
        },
        {
            name: 'Get images bucket info',
            url: `${API_BASE}/p/${PROJECT_SLUG}/api/storage/buckets/images`,
            method: 'GET'
        },
        {
            name: 'List images in bucket',
            url: `${API_BASE}/p/${PROJECT_SLUG}/api/storage/images/files`,
            method: 'GET'
        },
        {
            name: 'Upload to images bucket',
            url: `${API_BASE}/p/${PROJECT_SLUG}/api/storage/images/files`,
            method: 'POST'
        }
    ];
    
    for (const endpoint of endpoints) {
        try {
            console.log(`Testing ${endpoint.name}...`);
            const response = await fetch(endpoint.url, {
                method: 'OPTIONS',
                headers: {
                    'Origin': PHOTOPORTFOLIO_ORIGIN,
                    'Access-Control-Request-Method': endpoint.method,
                    'Access-Control-Request-Headers': 'Content-Type,X-API-Key'
                }
            });
            
            const status = response.ok ? '✅' : '❌';
            console.log(`${status} ${endpoint.name}: ${response.status} ${response.statusText}`);
            
            // Check CORS headers
            const corsHeaders = {};
            for (const [key, value] of response.headers) {
                if (key.toLowerCase().includes('access-control')) {
                    corsHeaders[key] = value;
                }
            }
            
            if (Object.keys(corsHeaders).length > 0) {
                console.log('   CORS Headers:');
                Object.entries(corsHeaders).forEach(([key, value]) => {
                    console.log(`     ${key}: ${value}`);
                });
            }
            
            console.log(''); // Empty line for readability
            
        } catch (error) {
            console.log(`❌ ${endpoint.name}: ERROR - ${error.message}\n`);
        }
    }
}

async function testActualRequests() {
    console.log('🧪 Testing actual API requests (without authentication)...\n');
    
    const endpoints = [
        {
            name: 'List buckets',
            url: `${API_BASE}/p/${PROJECT_SLUG}/api/storage/buckets`,
            method: 'GET'
        },
        {
            name: 'List images in bucket',
            url: `${API_BASE}/p/${PROJECT_SLUG}/api/storage/images/files`,
            method: 'GET'
        }
    ];
    
    for (const endpoint of endpoints) {
        try {
            console.log(`Testing ${endpoint.name}...`);
            const response = await fetch(endpoint.url, {
                method: endpoint.method,
                headers: {
                    'Origin': PHOTOPORTFOLIO_ORIGIN,
                    'Content-Type': 'application/json'
                }
            });
            
            const status = response.ok ? '✅' : '❌';
            console.log(`${status} ${endpoint.name}: ${response.status} ${response.statusText}`);
            
            if (response.ok) {
                try {
                    const data = await response.json();
                    if (Array.isArray(data)) {
                        console.log(`   Found ${data.length} items`);
                        if (data.length > 0) {
                            console.log(`   Sample item:`, Object.keys(data[0]));
                        }
                    } else {
                        console.log(`   Response keys:`, Object.keys(data));
                    }
                } catch (e) {
                    console.log('   Response: non-JSON data');
                }
            } else {
                try {
                    const errorText = await response.text();
                    console.log(`   Error: ${errorText}`);
                } catch (e) {
                    console.log('   Error: Could not read response');
                }
            }
            
            console.log(''); // Empty line
            
        } catch (error) {
            console.log(`❌ ${endpoint.name}: ERROR - ${error.message}\n`);
        }
    }
}

async function createTestUpload() {
    console.log('🧪 Testing file upload simulation...\n');
    
    // Create a simple test "image" (actually just text)
    const testContent = 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==';
    
    try {
        console.log('Testing image upload...');
        
        // Simulate FormData for file upload
        const formData = new FormData();
        
        // Create a blob from the test data
        const blob = new Blob([testContent], { type: 'text/plain' });
        formData.append('file', blob, 'test-image.txt');
        
        const response = await fetch(`${API_BASE}/p/${PROJECT_SLUG}/api/storage/images/files`, {
            method: 'POST',
            headers: {
                'Origin': PHOTOPORTFOLIO_ORIGIN
                // Don't set Content-Type - let the browser set it for FormData
            },
            body: formData
        });
        
        const status = response.ok ? '✅' : '❌';
        console.log(`${status} Upload test: ${response.status} ${response.statusText}`);
        
        if (response.ok) {
            try {
                const data = await response.json();
                console.log('   Upload successful! File details:');
                console.log(`     ID: ${data.id}`);
                console.log(`     Original Name: ${data.original_name}`);
                console.log(`     File Name: ${data.file_name}`);
                console.log(`     Size: ${data.size} bytes`);
                console.log(`     Public URL: ${data.public_url || 'N/A'}`);
                console.log(`     Private URL: ${data.private_url || 'N/A'}`);
            } catch (e) {
                console.log('   Upload successful but could not parse response');
            }
        } else {
            try {
                const errorText = await response.text();
                console.log(`   Error details: ${errorText}`);
            } catch (e) {
                console.log('   Error: Could not read error response');
            }
        }
        
    } catch (error) {
        console.log(`❌ Upload test: ERROR - ${error.message}`);
    }
}

async function main() {
    console.log('🚀 PhotoPortfolio Storage Upload Test\n');
    console.log(`Testing against: ${API_BASE}`);
    console.log(`Project slug: ${PROJECT_SLUG}`);
    console.log(`Origin: ${PHOTOPORTFOLIO_ORIGIN}\n`);
    
    // Test 1: CORS Preflight
    await testCORSPreflight();
    
    // Test 2: Actual requests  
    await testActualRequests();
    
    // Test 3: File upload simulation
    await createTestUpload();
    
    console.log('\n📝 Summary:');
    console.log('1. If CORS preflight tests pass ✅ - CORS is configured correctly');
    console.log('2. If bucket listing works ✅ - Storage buckets are accessible');
    console.log('3. If upload test works ✅ - PhotoPortfolio can upload images');
    
    console.log('\n🔧 Correct PhotoPortfolio Integration:');
    console.log(`- Use project SLUG "${PROJECT_SLUG}" in URLs, not project ID`);
    console.log(`- Upload endpoint: POST ${API_BASE}/p/${PROJECT_SLUG}/api/storage/images/files`);
    console.log(`- List files: GET ${API_BASE}/p/${PROJECT_SLUG}/api/storage/images/files`);
    console.log('- Use multipart/form-data for uploads with "file" field name');
    console.log('- Include Origin header: http://localhost:1234');
}

// Run the script
if (require.main === module) {
    main();
}

module.exports = {
    testCORSPreflight,
    testActualRequests,
    createTestUpload
};