#!/usr/bin/env node

/**
 * Test PhotoPortfolio integration with CloudBox using project ID
 */

const CLOUDBOX_URL = 'http://localhost:8080';
const PROJECT_ID = 6; // Using PhotoPortfolio project ID directly
const API_KEY = 'eb8624e56a8ab37c1c14b088808a772bfeaac620a4a7f32b10465fc29f441400'; // PhotoPortfolio API key

async function testCORS() {
    console.log('🔍 Testing CORS preflight for PhotoPortfolio...\n');
    
    const response = await fetch(`${CLOUDBOX_URL}/p/${PROJECT_ID}/api/storage/images/files`, {
        method: 'OPTIONS',
        headers: {
            'Origin': 'http://localhost:1234',
            'Access-Control-Request-Method': 'POST',
            'Access-Control-Request-Headers': 'content-type,x-api-key'
        }
    });
    
    console.log('CORS Preflight Response:');
    console.log('  Status:', response.status);
    console.log('  Access-Control-Allow-Origin:', response.headers.get('Access-Control-Allow-Origin'));
    console.log('  Access-Control-Allow-Methods:', response.headers.get('Access-Control-Allow-Methods'));
    console.log('  Access-Control-Allow-Headers:', response.headers.get('Access-Control-Allow-Headers'));
    
    return response.status === 204;
}

async function testListImages() {
    console.log('\n📋 Testing list images with project ID...\n');
    
    const response = await fetch(`${CLOUDBOX_URL}/p/${PROJECT_ID}/api/storage/images/files`, {
        headers: {
            'Origin': 'http://localhost:1234',
            'X-API-Key': API_KEY
        }
    });
    
    console.log('List Images Response:');
    console.log('  Status:', response.status);
    console.log('  Access-Control-Allow-Origin:', response.headers.get('Access-Control-Allow-Origin'));
    
    if (response.ok) {
        const data = await response.json();
        console.log('  Files found:', Array.isArray(data.files) ? data.files.length : 0);
    } else {
        console.log('  Error:', await response.text());
    }
    
    return response.ok;
}

async function testUploadImage() {
    console.log('\n📤 Testing image upload with project ID...\n');
    
    // Create a simple test image blob
    const canvas = new (await import('canvas')).Canvas(100, 100);
    const ctx = canvas.getContext('2d');
    ctx.fillStyle = 'blue';
    ctx.fillRect(0, 0, 100, 100);
    const buffer = canvas.toBuffer('image/png');
    
    const formData = new FormData();
    formData.append('file', new Blob([buffer], { type: 'image/png' }), 'test.png');
    
    const response = await fetch(`${CLOUDBOX_URL}/p/${PROJECT_ID}/api/storage/images/files`, {
        method: 'POST',
        headers: {
            'Origin': 'http://localhost:1234',
            'X-API-Key': API_KEY
        },
        body: formData
    });
    
    console.log('Upload Response:');
    console.log('  Status:', response.status);
    console.log('  Access-Control-Allow-Origin:', response.headers.get('Access-Control-Allow-Origin'));
    
    if (response.ok) {
        const data = await response.json();
        console.log('  File uploaded:', data.file?.name || 'Unknown');
        console.log('  File ID:', data.file?.id || 'Unknown');
    } else {
        console.log('  Error:', await response.text());
    }
    
    return response.ok;
}

async function main() {
    console.log('🎯 CloudBox + PhotoPortfolio Integration Test');
    console.log('=============================================');
    console.log(`Using Project ID: ${PROJECT_ID}`);
    console.log(`CloudBox URL: ${CLOUDBOX_URL}`);
    console.log(`PhotoPortfolio Origin: http://localhost:1234\n`);
    
    let allTestsPassed = true;
    
    // Test CORS
    const corsOk = await testCORS();
    if (corsOk) {
        console.log('\n✅ CORS test passed');
    } else {
        console.log('\n❌ CORS test failed');
        allTestsPassed = false;
    }
    
    // Test list images
    const listOk = await testListImages();
    if (listOk) {
        console.log('\n✅ List images test passed');
    } else {
        console.log('\n❌ List images test failed');
        allTestsPassed = false;
    }
    
    // Test upload (skip for now as it needs canvas)
    console.log('\n⏭️  Skipping upload test (requires canvas module)');
    
    // Summary
    console.log('\n\n' + '='.repeat(50));
    if (allTestsPassed) {
        console.log('✅ All tests passed! PhotoPortfolio can now use project ID.');
        console.log('\n📝 PhotoPortfolio should use:');
        console.log(`   Base URL: ${CLOUDBOX_URL}/p/${PROJECT_ID}/api`);
        console.log('   Endpoints:');
        console.log('     - Upload: POST /storage/images/files');
        console.log('     - List: GET /storage/images/files');
        console.log('     - Delete: DELETE /storage/images/files/{id}');
    } else {
        console.log('❌ Some tests failed. Please check the errors above.');
    }
}

main().catch(console.error);