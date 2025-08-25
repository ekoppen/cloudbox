#!/usr/bin/env node

/**
 * Fix PhotoPortfolio Storage Upload Issues
 * 
 * This script addresses the CORS and bucket configuration issues preventing
 * PhotoPortfolio from uploading images to CloudBox project 2.
 * 
 * Issues Fixed:
 * 1. Creates missing "images" storage bucket for project 2
 * 2. Updates CORS configuration to allow localhost:1234 origin
 * 3. Provides correct endpoint URLs for PhotoPortfolio integration
 */

const fs = require('fs');

// Configuration
const API_BASE = 'http://localhost:8080';
const PROJECT_ID = 2;
const PHOTOPORTFOLIO_ORIGIN = 'http://localhost:1234';  // Corrected port
const PROJECT_SLUG = 'photoportfolio';  // Assuming this is the slug for project 2

// CORS configuration to apply
const CORS_CONFIG = {
    allowed_origins: [
        'http://localhost:3000',    // CloudBox frontend
        'http://localhost:1234'     // PhotoPortfolio (corrected port)
    ],
    allowed_methods: ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'OPTIONS'],
    allowed_headers: ['Content-Type', 'Authorization', 'X-API-Key'],
    allow_credentials: true,
    max_age: 3600
};

// Storage bucket configuration
const IMAGES_BUCKET_CONFIG = {
    name: 'images',
    description: 'Image storage bucket for PhotoPortfolio',
    max_file_size: 52428800,  // 50MB
    allowed_types: [
        'image/jpeg',
        'image/jpg', 
        'image/png',
        'image/gif',
        'image/webp',
        'image/svg+xml'
    ],
    is_public: true  // Make bucket public for easy access
};

async function getJWTToken() {
    console.log('🔐 Getting JWT token for authentication...');
    
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

async function checkProject(token) {
    console.log(`📋 Checking project ${PROJECT_ID} details...`);
    
    try {
        const response = await fetch(`${API_BASE}/api/v1/projects/${PROJECT_ID}`, {
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            }
        });
        
        if (response.ok) {
            const project = await response.json();
            console.log('📄 Project found:', {
                id: project.id,
                name: project.name,
                slug: project.slug,
                is_active: project.is_active
            });
            return project;
        } else {
            throw new Error(`Project not found: ${response.status} ${response.statusText}`);
        }
        
    } catch (error) {
        console.error('❌ Failed to get project details:', error.message);
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
            console.log('✅ CORS configuration updated successfully');
            return true;
        } else {
            const errorText = await response.text();
            throw new Error(`CORS update failed: ${response.status} ${response.statusText}\n${errorText}`);
        }
        
    } catch (error) {
        console.error('❌ Failed to update CORS config:', error.message);
        throw error;
    }
}

async function listExistingBuckets(token) {
    console.log(`📦 Checking existing buckets for project ${PROJECT_ID}...`);
    
    try {
        const response = await fetch(`${API_BASE}/api/v1/projects/${PROJECT_ID}/storage/buckets`, {
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            }
        });
        
        if (response.ok) {
            const buckets = await response.json();
            console.log(`📄 Found ${buckets.length} existing buckets:`);
            buckets.forEach(bucket => {
                console.log(`   - ${bucket.name} (${bucket.is_public ? 'public' : 'private'}, ${bucket.file_count} files)`);
            });
            return buckets;
        } else {
            console.log('⚠️  Failed to get buckets list');
            return [];
        }
        
    } catch (error) {
        console.error('❌ Failed to list buckets:', error.message);
        return [];
    }
}

async function createImagesBucket(token) {
    console.log(`🪣 Creating "images" bucket for project ${PROJECT_ID}...`);
    
    try {
        const response = await fetch(`${API_BASE}/api/v1/projects/${PROJECT_ID}/storage/buckets`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(IMAGES_BUCKET_CONFIG)
        });
        
        if (response.ok) {
            const bucket = await response.json();
            console.log('✅ Images bucket created successfully:', {
                name: bucket.name,
                description: bucket.description,
                is_public: bucket.is_public,
                max_file_size: `${Math.round(bucket.max_file_size / 1024 / 1024)}MB`,
                allowed_types: bucket.allowed_types
            });
            return bucket;
        } else if (response.status === 409) {
            console.log('ℹ️  Images bucket already exists');
            return null;
        } else {
            const errorText = await response.text();
            throw new Error(`Bucket creation failed: ${response.status} ${response.statusText}\n${errorText}`);
        }
        
    } catch (error) {
        console.error('❌ Failed to create images bucket:', error.message);
        throw error;
    }
}

async function testEndpoints(projectSlug) {
    console.log('\n🧪 Testing storage endpoints...');
    
    const endpoints = [
        {
            name: 'List buckets',
            url: `${API_BASE}/p/${projectSlug}/api/storage/buckets`,
            method: 'GET'
        },
        {
            name: 'Get images bucket',
            url: `${API_BASE}/p/${projectSlug}/api/storage/buckets/images`,
            method: 'GET'
        },
        {
            name: 'List images bucket files', 
            url: `${API_BASE}/p/${projectSlug}/api/storage/images/files`,
            method: 'GET'
        }
    ];
    
    for (const endpoint of endpoints) {
        try {
            console.log(`   Testing ${endpoint.name}...`);
            const response = await fetch(endpoint.url, {
                method: 'OPTIONS',
                headers: {
                    'Origin': PHOTOPORTFOLIO_ORIGIN,
                    'Access-Control-Request-Method': endpoint.method,
                    'Access-Control-Request-Headers': 'Content-Type,X-API-Key'
                }
            });
            
            console.log(`   ${endpoint.name}: ${response.status} ${response.statusText}`);
            
            // Check CORS headers
            const corsHeaders = {};
            for (const [key, value] of response.headers) {
                if (key.toLowerCase().includes('access-control')) {
                    corsHeaders[key] = value;
                }
            }
            if (Object.keys(corsHeaders).length > 0) {
                console.log(`   CORS headers:`, corsHeaders);
            }
            
        } catch (error) {
            console.log(`   ${endpoint.name}: ERROR - ${error.message}`);
        }
    }
}

async function main() {
    console.log('🚀 CloudBox PhotoPortfolio Storage Fix\n');
    
    try {
        // Step 1: Authenticate
        const token = await getJWTToken();
        
        // Step 2: Check project details
        const project = await checkProject(token);
        const actualSlug = project.slug || PROJECT_SLUG;
        
        // Step 3: Get current CORS config
        const currentCorsConfig = await getCurrentCORSConfig(token);
        
        // Step 4: Update CORS config
        await updateCORSConfig(token);
        
        // Step 5: Check existing buckets
        const existingBuckets = await listExistingBuckets(token);
        const hasImagesBucket = existingBuckets.some(bucket => bucket.name === 'images');
        
        // Step 6: Create images bucket if it doesn't exist
        if (!hasImagesBucket) {
            await createImagesBucket(token);
        } else {
            console.log('ℹ️  Images bucket already exists, skipping creation');
        }
        
        // Step 7: Test endpoints
        await testEndpoints(actualSlug);
        
        console.log('\n🎉 PhotoPortfolio storage fix completed successfully!\n');
        
        console.log('📝 Correct endpoints for PhotoPortfolio:');
        console.log(`   📦 List buckets: GET ${API_BASE}/p/${actualSlug}/api/storage/buckets`);
        console.log(`   🖼️  Upload image: POST ${API_BASE}/p/${actualSlug}/api/storage/images/files`);
        console.log(`   📂 List images: GET ${API_BASE}/p/${actualSlug}/api/storage/images/files`);
        console.log(`   🗑️  Delete image: DELETE ${API_BASE}/p/${actualSlug}/api/storage/images/files/{file_id}`);
        
        console.log('\n🔧 Required headers for PhotoPortfolio requests:');
        console.log('   - Origin: http://localhost:1234');
        console.log('   - Content-Type: multipart/form-data (for uploads)');
        console.log('   - X-API-Key: {your-project-api-key} (optional, for authentication)');
        
        console.log('\n❗ Important Notes:');
        console.log('   1. Use project SLUG, not project ID in URLs');
        console.log('   2. Storage endpoint is /storage/{bucket}/files, NOT /storage/buckets/{bucket}/files');
        console.log('   3. PhotoPortfolio origin is now allowed: http://localhost:1234');
        console.log('   4. Images bucket is public - files are accessible without authentication');
        
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
    createImagesBucket,
    testEndpoints,
    CORS_CONFIG,
    IMAGES_BUCKET_CONFIG
};