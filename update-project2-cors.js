#!/usr/bin/env node

/**
 * Script to update CORS configuration for project ID 2 (PhotoPortfolio)
 * This script updates the project-specific CORS settings to allow PhotoPortfolio origin
 */

const fs = require('fs');

// Configuration
const API_BASE = 'http://localhost:8080';
const PROJECT_ID = 2;
const PHOTOPORTFOLIO_ORIGIN = 'http://localhost:3123';

// CORS configuration to apply
const CORS_CONFIG = {
    allowed_origins: [
        'http://localhost:3000',    // CloudBox frontend
        'http://localhost:3123'     // PhotoPortfolio
    ],
    allowed_methods: ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'OPTIONS'],
    allowed_headers: ['Content-Type', 'Authorization', 'X-API-Key'],
    allow_credentials: true,
    max_age: 3600
};

async function getJWTToken() {
    console.log('🔐 Getting JWT token for authentication...');
    
    // You'll need to replace these with actual admin credentials
    const loginData = {
        email: 'admin@cloudbox.local',
        password: 'admin123'  // Default from .env
    };
    
    try {
        const response = await fetch(`${API_BASE}/api/v1/auth/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(loginData)
        });
        
        if (!response.ok) {
            throw new Error(`Login failed: ${response.status} ${response.statusText}`);
        }
        
        const data = await response.json();
        console.log('✅ Authentication successful');
        return data.token;
        
    } catch (error) {
        console.error('❌ Authentication failed:', error.message);
        console.log('\n💡 Please ensure:');
        console.log('   1. CloudBox backend is running on http://localhost:8080');
        console.log('   2. Admin credentials are correct (check .env file)');
        console.log('   3. Database is accessible and contains admin user');
        throw error;
    }
}

async function getCurrentCORSConfig(token) {
    console.log(`📋 Getting current CORS configuration for project ${PROJECT_ID}...`);
    
    try {
        const response = await fetch(`${API_BASE}/api/v1/projects/${PROJECT_ID}/cors`, {
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            }
        });
        
        if (response.ok) {
            const config = await response.json();
            console.log('📄 Current CORS config:', JSON.stringify(config, null, 2));
            return config;
        } else if (response.status === 404) {
            console.log('⚠️  No CORS config found for project (will create default)');
            return null;
        } else {
            throw new Error(`Failed to get CORS config: ${response.status} ${response.statusText}`);
        }
        
    } catch (error) {
        console.error('❌ Failed to get current CORS config:', error.message);
        throw error;
    }
}

async function updateCORSConfig(token) {
    console.log(`🔧 Updating CORS configuration for project ${PROJECT_ID}...`);
    
    try {
        const response = await fetch(`${API_BASE}/api/v1/projects/${PROJECT_ID}/cors`, {
            method: 'PUT',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(CORS_CONFIG)
        });
        
        if (response.ok) {
            const result = await response.json();
            console.log('✅ CORS configuration updated successfully:', result);
            return true;
        } else {
            const errorText = await response.text();
            throw new Error(`Update failed: ${response.status} ${response.statusText}\n${errorText}`);
        }
        
    } catch (error) {
        console.error('❌ Failed to update CORS config:', error.message);
        throw error;
    }
}

async function testCORS() {
    console.log('\n🧪 Testing CORS configuration...');
    
    try {
        // Test global CORS
        console.log('   Testing global endpoint CORS...');
        const globalResponse = await fetch(`${API_BASE}/api/v1/auth/login`, {
            method: 'OPTIONS',
            headers: {
                'Origin': PHOTOPORTFOLIO_ORIGIN,
                'Access-Control-Request-Method': 'POST',
                'Access-Control-Request-Headers': 'Content-Type'
            }
        });
        
        console.log(`   Global CORS test: ${globalResponse.status}`);
        const globalCORSHeaders = {};
        for (const [key, value] of globalResponse.headers) {
            if (key.toLowerCase().includes('access-control')) {
                globalCORSHeaders[key] = value;
            }
        }
        console.log('   Global CORS headers:', globalCORSHeaders);
        
        // Test project-specific CORS
        console.log('   Testing project endpoint CORS...');
        const projectResponse = await fetch(`${API_BASE}/p/photoportfolio/api/data/test`, {
            method: 'OPTIONS',
            headers: {
                'Origin': PHOTOPORTFOLIO_ORIGIN,
                'Access-Control-Request-Method': 'GET',
                'Access-Control-Request-Headers': 'Content-Type,X-API-Key'
            }
        });
        
        console.log(`   Project CORS test: ${projectResponse.status}`);
        const projectCORSHeaders = {};
        for (const [key, value] of projectResponse.headers) {
            if (key.toLowerCase().includes('access-control')) {
                projectCORSHeaders[key] = value;
            }
        }
        console.log('   Project CORS headers:', projectCORSHeaders);
        
    } catch (error) {
        console.error('❌ CORS testing failed:', error.message);
    }
}

async function main() {
    console.log('🚀 CloudBox PhotoPortfolio CORS Configuration Update\n');
    
    try {
        // Step 1: Authenticate
        const token = await getJWTToken();
        
        // Step 2: Get current config
        const currentConfig = await getCurrentCORSConfig(token);
        
        // Step 3: Update config
        await updateCORSConfig(token);
        
        // Step 4: Test CORS
        await testCORS();
        
        console.log('\n🎉 CORS configuration update completed successfully!');
        console.log('\n📝 Next steps for PhotoPortfolio:');
        console.log('   1. Use project-specific endpoint: POST http://localhost:8080/p/photoportfolio/api/users/login');
        console.log('   2. Or continue using global endpoint: POST http://localhost:8080/api/v1/auth/login');
        console.log('   3. Both should now work with origin http://localhost:3123');
        
    } catch (error) {
        console.error('\n💥 Script failed:', error.message);
        process.exit(1);
    }
}

// Run the script
if (require.main === module) {
    main();
}

module.exports = {
    getJWTToken,
    getCurrentCORSConfig,
    updateCORSConfig,
    testCORS,
    CORS_CONFIG
};