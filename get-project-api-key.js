#!/usr/bin/env node

/**
 * Get API key for project 2
 */

const CLOUDBOX_URL = 'http://localhost:8080';
const ADMIN_EMAIL = 'admin@cloudbox.dev';
const ADMIN_PASSWORD = 'admin123';
const PROJECT_ID = 2;

async function login() {
    const response = await fetch(`${CLOUDBOX_URL}/api/v1/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            email: ADMIN_EMAIL,
            password: ADMIN_PASSWORD
        })
    });
    
    if (!response.ok) {
        throw new Error(`Login failed: ${await response.text()}`);
    }
    
    const data = await response.json();
    return data.token;
}

async function getAPIKeys(token, projectId) {
    const response = await fetch(`${CLOUDBOX_URL}/api/v1/projects/${projectId}/api-keys`, {
        headers: { 'Authorization': `Bearer ${token}` }
    });
    
    if (!response.ok) {
        throw new Error(`Failed to get API keys: ${await response.text()}`);
    }
    
    return response.json();
}

async function createAPIKey(token, projectId) {
    const response = await fetch(`${CLOUDBOX_URL}/api/v1/projects/${projectId}/api-keys`, {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            name: 'PhotoPortfolio API Key',
            description: 'API key for PhotoPortfolio integration'
        })
    });
    
    if (!response.ok) {
        throw new Error(`Failed to create API key: ${await response.text()}`);
    }
    
    return response.json();
}

async function main() {
    console.log('🔑 Getting API Key for Project 2');
    console.log('================================\n');
    
    try {
        // Login
        console.log('📝 Logging in...');
        const token = await login();
        console.log('✅ Logged in\n');
        
        // Get existing API keys
        console.log('🔍 Checking existing API keys...');
        const keys = await getAPIKeys(token, PROJECT_ID);
        
        // Always create a new key for testing
        console.log('📝 Creating new API key for PhotoPortfolio...');
        const newKey = await createAPIKey(token, PROJECT_ID);
        console.log('✅ API Key created:\n');
        console.log(`  Name: ${newKey.name}`);
        console.log(`  Key: ${newKey.key}\n`);
        
        if (false) {
            console.log('📝 No API keys found, creating new one...');
            const newKey = await createAPIKey(token, PROJECT_ID);
            console.log('✅ API Key created:\n');
            console.log(`  Name: ${newKey.name}`);
            console.log(`  Key: ${newKey.key}\n`);
        }
        
        console.log('\n📌 Use this API key in PhotoPortfolio configuration');
        
    } catch (error) {
        console.error('\n❌ Error:', error.message);
        process.exit(1);
    }
}

main();