#!/usr/bin/env node

// Test script to verify project-specific CORS configuration and dynamic updates
// This tests whether CORS can be configured per project without requiring a restart

const fetch = require('node-fetch');

const BASE_URL = 'http://localhost:8080';
const PROJECT_ID = 2;
const ORIGIN = 'http://localhost:3123'; // PhotoPortfolio origin

// Test credentials - you may need to update these
const TEST_CREDENTIALS = {
    username: 'admin',
    password: 'admin123'
};

async function login() {
    console.log('🔐 Logging in to CloudBox...');
    
    const response = await fetch(`${BASE_URL}/api/v1/auth/login`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(TEST_CREDENTIALS)
    });
    
    if (!response.ok) {
        const error = await response.text();
        throw new Error(`Login failed: ${response.status} - ${error}`);
    }
    
    const data = await response.json();
    console.log('✅ Login successful');
    return data.token;
}

async function getCurrentCORSConfig(token) {
    console.log(`\n📋 Getting current CORS config for project ${PROJECT_ID}...`);
    
    const response = await fetch(`${BASE_URL}/api/v1/projects/${PROJECT_ID}/cors`, {
        headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
        }
    });
    
    if (!response.ok) {
        const error = await response.text();
        console.log(`❌ Failed to get CORS config: ${response.status} - ${error}`);
        return null;
    }
    
    const corsConfig = await response.json();
    console.log('📋 Current CORS config:', JSON.stringify(corsConfig, null, 2));
    return corsConfig;
}

async function updateCORSConfig(token, newConfig) {
    console.log(`\n🔄 Updating CORS config for project ${PROJECT_ID}...`);
    console.log('New config:', JSON.stringify(newConfig, null, 2));
    
    const response = await fetch(`${BASE_URL}/api/v1/projects/${PROJECT_ID}/cors`, {
        method: 'PUT',
        headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(newConfig)
    });
    
    if (!response.ok) {
        const error = await response.text();
        console.log(`❌ Failed to update CORS config: ${response.status} - ${error}`);
        return false;
    }
    
    const result = await response.json();
    console.log('✅ CORS config updated:', result.message);
    return true;
}

async function testCORSRequest(origin) {
    console.log(`\n🌐 Testing CORS preflight request from origin: ${origin}`);
    
    const response = await fetch(`${BASE_URL}/p/test/api/collections`, {
        method: 'OPTIONS',
        headers: {
            'Origin': origin,
            'Access-Control-Request-Method': 'GET',
            'Access-Control-Request-Headers': 'X-API-Key, Content-Type'
        }
    });
    
    const allowOrigin = response.headers.get('access-control-allow-origin');
    const allowMethods = response.headers.get('access-control-allow-methods');
    
    console.log(`📊 Response status: ${response.status}`);
    console.log(`📊 Access-Control-Allow-Origin: ${allowOrigin}`);
    console.log(`📊 Access-Control-Allow-Methods: ${allowMethods}`);
    
    return {
        status: response.status,
        allowOrigin,
        allowMethods,
        success: allowOrigin === origin || allowOrigin === '*'
    };
}

async function main() {
    console.log('🧪 CloudBox Project-Specific CORS Test');
    console.log('======================================');
    
    try {
        // Step 1: Login
        const token = await login();
        
        // Step 2: Get current CORS config
        const currentConfig = await getCurrentCORSConfig(token);
        if (!currentConfig) {
            console.log('❌ Cannot proceed without current CORS config');
            return;
        }
        
        // Step 3: Test current CORS
        console.log('\n🧪 Testing current CORS configuration...');
        const currentTest = await testCORSRequest(ORIGIN);
        console.log(`📊 Current CORS test result: ${currentTest.success ? '✅ PASS' : '❌ FAIL'}`);
        
        // Step 4: Update CORS config to temporarily disable the origin
        const tempConfig = {
            ...currentConfig,
            allowed_origins: ['http://localhost:3000'] // Remove 3123 temporarily
        };
        
        console.log('\n🔄 Step 4: Temporarily disabling origin in CORS config...');
        const updateSuccess1 = await updateCORSConfig(token, tempConfig);
        
        if (updateSuccess1) {
            // Step 5: Test that the origin is now blocked
            console.log('\n🧪 Testing with updated CORS configuration (should fail)...');
            const blockedTest = await testCORSRequest(ORIGIN);
            console.log(`📊 Blocked CORS test result: ${!blockedTest.success ? '✅ PASS (correctly blocked)' : '❌ FAIL (should be blocked)'}`);
            
            // Step 6: Restore original config
            console.log('\n🔄 Step 6: Restoring original CORS configuration...');
            const updateSuccess2 = await updateCORSConfig(token, currentConfig);
            
            if (updateSuccess2) {
                // Step 7: Test that the origin works again
                console.log('\n🧪 Testing with restored CORS configuration (should work)...');
                const restoredTest = await testCORSRequest(ORIGIN);
                console.log(`📊 Restored CORS test result: ${restoredTest.success ? '✅ PASS (correctly allowed)' : '❌ FAIL (should be allowed)'}`);
                
                // Summary
                console.log('\n📋 Test Summary:');
                console.log('================');
                console.log(`Initial CORS test: ${currentTest.success ? '✅ PASS' : '❌ FAIL'}`);
                console.log(`Dynamic update test: ${updateSuccess1 ? '✅ PASS' : '❌ FAIL'}`);
                console.log(`Block verification: ${!blockedTest.success ? '✅ PASS' : '❌ FAIL'}`);
                console.log(`Restore test: ${updateSuccess2 ? '✅ PASS' : '❌ FAIL'}`);
                console.log(`Final CORS test: ${restoredTest.success ? '✅ PASS' : '❌ FAIL'}`);
                
                const allPassed = currentTest.success && updateSuccess1 && !blockedTest.success && updateSuccess2 && restoredTest.success;
                console.log(`\n🎯 Overall result: ${allPassed ? '✅ ALL TESTS PASSED - Dynamic CORS works!' : '❌ SOME TESTS FAILED'}`);
                
                if (allPassed) {
                    console.log('\n✨ Conclusions:');
                    console.log('• Project-specific CORS is working correctly');
                    console.log('• CORS configuration can be updated dynamically');
                    console.log('• No backend restart is required for CORS changes');
                    console.log('• PhotoPortfolio can access CloudBox API via project-specific CORS');
                }
            }
        }
        
    } catch (error) {
        console.error('❌ Test failed:', error.message);
    }
}

if (require.main === module) {
    main();
}