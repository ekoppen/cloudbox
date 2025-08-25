#!/usr/bin/env node

/**
 * Fix script to create default storage buckets for existing projects
 * and update CORS configuration to support PhotoPortfolio
 */

const CLOUDBOX_URL = 'http://localhost:8080';
const ADMIN_EMAIL = 'admin@cloudbox.dev';
const ADMIN_PASSWORD = 'admin123';

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

async function getProjects(token) {
    const response = await fetch(`${CLOUDBOX_URL}/api/v1/projects`, {
        headers: { 'Authorization': `Bearer ${token}` }
    });
    
    if (!response.ok) {
        throw new Error(`Failed to get projects: ${await response.text()}`);
    }
    
    return response.json();
}

async function createBucket(token, projectId, bucketConfig) {
    const response = await fetch(`${CLOUDBOX_URL}/api/v1/projects/${projectId}/storage/buckets`, {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(bucketConfig)
    });
    
    if (!response.ok && response.status !== 409) { // 409 = bucket already exists
        console.error(`Failed to create bucket ${bucketConfig.name} for project ${projectId}: ${await response.text()}`);
        return false;
    }
    
    return true;
}

async function updateCORSConfig(token, projectId, origins) {
    const response = await fetch(`${CLOUDBOX_URL}/api/v1/projects/${projectId}/cors`, {
        method: 'PUT',
        headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            allowed_origins: origins,
            allowed_methods: ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'OPTIONS'],
            allowed_headers: ['*'],
            allow_credentials: true,
            max_age: 86400
        })
    });
    
    if (!response.ok) {
        console.error(`Failed to update CORS for project ${projectId}: ${await response.text()}`);
        return false;
    }
    
    return true;
}

async function main() {
    console.log('🔧 CloudBox Project Fixer');
    console.log('========================\n');
    
    try {
        // Login as admin
        console.log('📝 Logging in as admin...');
        const token = await login();
        console.log('✅ Logged in successfully\n');
        
        // Get all projects
        console.log('📋 Fetching all projects...');
        const projects = await getProjects(token);
        console.log(`✅ Found ${projects.length} projects\n`);
        
        // Default bucket configurations
        const defaultBuckets = [
            {
                name: 'images',
                public: true,
                max_file_size: 10485760, // 10MB
                allowed_file_types: ['.jpg', '.jpeg', '.png', '.gif', '.webp', '.svg']
            },
            {
                name: 'documents',
                public: false,
                max_file_size: 52428800, // 50MB
                allowed_file_types: ['.pdf', '.doc', '.docx', '.txt', '.md']
            },
            {
                name: 'videos',
                public: false,
                max_file_size: 104857600, // 100MB
                allowed_file_types: ['.mp4', '.webm', '.ogg', '.mov']
            }
        ];
        
        // Process each project
        for (const project of projects) {
            console.log(`\n🔄 Processing project: ${project.name} (ID: ${project.id}, Slug: ${project.slug})`);
            
            // Create default buckets
            console.log('  📦 Creating default buckets...');
            for (const bucket of defaultBuckets) {
                const success = await createBucket(token, project.id, bucket);
                if (success) {
                    console.log(`    ✅ Bucket '${bucket.name}' created or already exists`);
                } else {
                    console.log(`    ⚠️  Failed to create bucket '${bucket.name}'`);
                }
            }
            
            // Update CORS to include PhotoPortfolio and other development origins
            console.log('  🌐 Updating CORS configuration...');
            const corsOrigins = [
                '*', // Allow all origins for development
                'http://localhost:1234', // PhotoPortfolio default
                'http://localhost:3000', // React default
                'http://localhost:5173', // Vite default
                'http://localhost:8080', // CloudBox itself
            ];
            
            const corsSuccess = await updateCORSConfig(token, project.id, corsOrigins);
            if (corsSuccess) {
                console.log('    ✅ CORS configuration updated');
            } else {
                console.log('    ⚠️  Failed to update CORS configuration');
            }
        }
        
        console.log('\n\n✅ All projects have been updated!');
        console.log('\n📌 Important: You can now use project IDs in all API calls:');
        console.log('   Example: http://localhost:8080/p/2/api/storage/images/files');
        console.log('   (where 2 is the project ID)\n');
        
    } catch (error) {
        console.error('\n❌ Error:', error.message);
        process.exit(1);
    }
}

main();