#!/usr/bin/env node

/**
 * Test that old slug-based endpoints no longer work (as expected)
 */

const CLOUDBOX_URL = 'http://localhost:8080';
const PROJECT_SLUG = 'wouterkoppencom'; // Old slug for project 2
const API_KEY = 'eb8624e56a8ab37c1c14b088808a772bfeaac620a4a7f32b10465fc29f441400';

async function testOldSlugEndpoint() {
    console.log('🔍 Testing old slug-based endpoint (should fail)...\n');
    
    const response = await fetch(`${CLOUDBOX_URL}/p/${PROJECT_SLUG}/api/storage/images/files`, {
        headers: {
            'Origin': 'http://localhost:1234',
            'X-API-Key': API_KEY
        }
    });
    
    console.log('Old Slug Endpoint Response:');
    console.log('  Status:', response.status);
    console.log('  URL:', response.url);
    
    if (response.status === 404) {
        console.log('  ✅ Correctly returns 404 (endpoint no longer exists)');
        return true;
    } else if (response.status === 400) {
        const error = await response.text();
        console.log('  ✅ Correctly returns 400 (invalid project ID format)');
        console.log('  Error:', error);
        return true;
    } else {
        console.log('  ❌ Unexpected response status');
        console.log('  Error:', await response.text());
        return false;
    }
}

async function testNewIdEndpoint() {
    console.log('\n🔍 Testing new ID-based endpoint (should work)...\n');
    
    const response = await fetch(`${CLOUDBOX_URL}/p/2/api/storage/images/files`, {
        headers: {
            'Origin': 'http://localhost:1234',
            'X-API-Key': API_KEY
        }
    });
    
    console.log('New ID Endpoint Response:');
    console.log('  Status:', response.status);
    console.log('  URL:', response.url);
    
    if (response.status === 200) {
        console.log('  ✅ Correctly returns 200 (works with project ID)');
        return true;
    } else {
        console.log('  ❌ Unexpected response status');
        console.log('  Error:', await response.text());
        return false;
    }
}

async function main() {
    console.log('🧪 CloudBox Slug vs ID Endpoint Test');
    console.log('=====================================\n');
    
    const oldEndpointFails = await testOldSlugEndpoint();
    const newEndpointWorks = await testNewIdEndpoint();
    
    console.log('\n\n' + '='.repeat(50));
    if (oldEndpointFails && newEndpointWorks) {
        console.log('✅ PERFECT! Migration to ID-based endpoints is successful.');
        console.log('\n📝 Summary:');
        console.log('   ❌ Old slug-based: /p/wouterkoppencom/api/* (correctly disabled)');
        console.log('   ✅ New ID-based: /p/2/api/* (working perfectly)');
        console.log('\nCloudBox now uses project IDs exclusively! 🎉');
    } else {
        console.log('❌ Migration incomplete or issues found.');
    }
}

main().catch(console.error);