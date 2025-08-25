// Test script to check CORS configuration for project ID 2
const fs = require('fs');
const path = require('path');

// Function to make API call to get CORS config
async function testCORSConfig() {
    const baseURL = 'http://localhost:8080';
    
    console.log('Testing CORS for project ID 2...\n');
    
    // Test 1: Check current CORS configuration via API
    try {
        console.log('1. Checking current CORS configuration...');
        const response = await fetch(`${baseURL}/api/v1/projects/2/cors`, {
            headers: {
                'Authorization': 'Bearer YOUR_JWT_TOKEN_HERE',
                'Content-Type': 'application/json'
            }
        });
        
        if (response.ok) {
            const corsConfig = await response.json();
            console.log('Current CORS config:', JSON.stringify(corsConfig, null, 2));
        } else {
            console.log('Failed to get CORS config:', response.status, response.statusText);
        }
    } catch (error) {
        console.log('Error getting CORS config:', error.message);
    }
    
    // Test 2: Test actual CORS request from PhotoPortfolio origin
    try {
        console.log('\n2. Testing CORS from PhotoPortfolio origin...');
        const testResponse = await fetch(`${baseURL}/api/v1/auth/login`, {
            method: 'OPTIONS',
            headers: {
                'Origin': 'http://localhost:3123',
                'Access-Control-Request-Method': 'POST',
                'Access-Control-Request-Headers': 'Content-Type'
            }
        });
        
        console.log('CORS preflight response status:', testResponse.status);
        console.log('CORS headers:');
        for (const [key, value] of testResponse.headers) {
            if (key.toLowerCase().includes('access-control')) {
                console.log(`  ${key}: ${value}`);
            }
        }
    } catch (error) {
        console.log('Error testing CORS:', error.message);
    }
    
    // Test 3: Test project-specific API endpoint
    try {
        console.log('\n3. Testing project-specific API CORS...');
        const projectResponse = await fetch(`${baseURL}/p/photoportfolio/api/data/test`, {
            method: 'OPTIONS',
            headers: {
                'Origin': 'http://localhost:3123',
                'Access-Control-Request-Method': 'GET',
                'Access-Control-Request-Headers': 'Content-Type,X-API-Key'
            }
        });
        
        console.log('Project API CORS preflight response status:', projectResponse.status);
        console.log('Project API CORS headers:');
        for (const [key, value] of projectResponse.headers) {
            if (key.toLowerCase().includes('access-control')) {
                console.log(`  ${key}: ${value}`);
            }
        }
    } catch (error) {
        console.log('Error testing project API CORS:', error.message);
    }
}

// Update CORS configuration to allow PhotoPortfolio
async function updateCORSConfig() {
    const baseURL = 'http://localhost:8080';
    
    console.log('\n4. Updating CORS configuration to allow PhotoPortfolio...');
    
    const newCORSConfig = {
        allowed_origins: ['http://localhost:3123', 'http://localhost:3000'],
        allowed_methods: ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'OPTIONS'],
        allowed_headers: ['Content-Type', 'Authorization', 'X-API-Key'],
        allow_credentials: true,
        max_age: 3600
    };
    
    try {
        const response = await fetch(`${baseURL}/api/v1/projects/2/cors`, {
            method: 'PUT',
            headers: {
                'Authorization': 'Bearer YOUR_JWT_TOKEN_HERE',
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(newCORSConfig)
        });
        
        if (response.ok) {
            const result = await response.json();
            console.log('CORS config updated successfully:', result);
        } else {
            console.log('Failed to update CORS config:', response.status, response.statusText);
            const errorText = await response.text();
            console.log('Error details:', errorText);
        }
    } catch (error) {
        console.log('Error updating CORS config:', error.message);
    }
}

// Run tests
testCORSConfig().then(() => {
    console.log('\n=== Test completed ===');
    console.log('\nTo update CORS config, you need to:');
    console.log('1. Get a valid JWT token by logging in');
    console.log('2. Replace YOUR_JWT_TOKEN_HERE with the actual token');
    console.log('3. Run updateCORSConfig() function');
});