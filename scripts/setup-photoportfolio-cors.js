#!/usr/bin/env node

/**
 * Flexible PhotoPortfolio CORS Setup Script
 * 
 * This script automatically configures CORS for PhotoPortfolio projects
 * supporting any port configuration and deployment scenario.
 * 
 * Usage:
 *   node setup-photoportfolio-cors.js [options]
 * 
 * Options:
 *   --project-id <id>        Project ID (default: 2)
 *   --port <port>           PhotoPortfolio port (default: auto-detect)
 *   --origin <origin>       Full origin URL (overrides port)
 *   --cloudbox-url <url>    CloudBox backend URL (default: http://localhost:8080)
 *   --update-global         Also update global CORS configuration
 *   --dry-run               Show what would be done without making changes
 */

const fs = require('fs');
const path = require('path');

// Parse command line arguments
function parseArgs() {
    const args = {
        projectId: 2,
        port: null,
        origin: null,
        cloudboxUrl: 'http://localhost:8080',
        updateGlobal: false,
        dryRun: false
    };
    
    for (let i = 2; i < process.argv.length; i++) {
        const arg = process.argv[i];
        const nextArg = process.argv[i + 1];
        
        switch (arg) {
            case '--project-id':
                args.projectId = parseInt(nextArg);
                i++;
                break;
            case '--port':
                args.port = parseInt(nextArg);
                i++;
                break;
            case '--origin':
                args.origin = nextArg;
                i++;
                break;
            case '--cloudbox-url':
                args.cloudboxUrl = nextArg;
                i++;
                break;
            case '--update-global':
                args.updateGlobal = true;
                break;
            case '--dry-run':
                args.dryRun = true;
                break;
            case '--help':
                console.log(`
PhotoPortfolio CORS Setup Script

Usage: node setup-photoportfolio-cors.js [options]

Options:
  --project-id <id>        Project ID (default: 2)  
  --port <port>           PhotoPortfolio port (default: auto-detect)
  --origin <origin>       Full origin URL (overrides port)
  --cloudbox-url <url>    CloudBox backend URL (default: http://localhost:8080)
  --update-global         Also update global CORS configuration  
  --dry-run               Show what would be done without making changes
  --help                  Show this help message

Examples:
  node setup-photoportfolio-cors.js --port 4041
  node setup-photoportfolio-cors.js --origin "https://photoportfolio.example.com"
  node setup-photoportfolio-cors.js --project-id 3 --port 3123 --update-global
                `);
                process.exit(0);
        }
    }
    
    return args;
}

// Auto-detect PhotoPortfolio configuration
function detectPhotoPortfolioConfig() {
    const configs = [];
    
    // Check common PhotoPortfolio locations
    const possiblePaths = [
        './photoportfolio',
        '../photoportfolio',
        '../../photoportfolio',
        process.env.PHOTOPORTFOLIO_PATH
    ].filter(Boolean);
    
    for (const pathToCheck of possiblePaths) {
        try {
            if (fs.existsSync(pathToCheck)) {
                // Look for package.json or config files
                const packagePath = path.join(pathToCheck, 'package.json');
                if (fs.existsSync(packagePath)) {
                    const packageData = JSON.parse(fs.readFileSync(packagePath, 'utf8'));
                    if (packageData.scripts && packageData.scripts.dev) {
                        const devScript = packageData.scripts.dev;
                        const portMatch = devScript.match(/--port[\s=](\d+)/);
                        if (portMatch) {
                            configs.push({
                                path: pathToCheck,
                                port: parseInt(portMatch[1]),
                                name: packageData.name || 'photoportfolio'
                            });
                        }
                    }
                }
                
                // Look for vite.config.js or similar
                const viteConfigPath = path.join(pathToCheck, 'vite.config.js');
                if (fs.existsSync(viteConfigPath)) {
                    const viteConfig = fs.readFileSync(viteConfigPath, 'utf8');
                    const portMatch = viteConfig.match(/port:\s*(\d+)/);
                    if (portMatch) {
                        configs.push({
                            path: pathToCheck,
                            port: parseInt(portMatch[1]),
                            name: 'photoportfolio'
                        });
                    }
                }
            }
        } catch (error) {
            // Ignore errors when checking paths
        }
    }
    
    return configs;
}

// Generate comprehensive CORS configuration
function generateCORSConfig(origin, additionalOrigins = []) {
    const baseConfig = {
        allowed_origins: [
            'http://localhost:3000',      // CloudBox frontend
            origin,                       // PhotoPortfolio origin
            ...additionalOrigins
        ],
        allowed_methods: ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'OPTIONS'],
        allowed_headers: [
            'Accept',
            'Content-Type',
            'Content-Length', 
            'Accept-Encoding',
            'Authorization',
            'X-CSRF-Token',
            'X-API-Key',
            'Cache-Control',
            'X-Requested-With',
            'Session-Token',
            'session-token',
            'X-Session-Token',
            'x-session-token',
            'X-Project-ID',
            'X-Project-Token',
            'Project-ID',
            'Project-Token'
        ],
        allow_credentials: true,
        max_age: 3600
    };
    
    // Remove duplicates from origins
    baseConfig.allowed_origins = [...new Set(baseConfig.allowed_origins)];
    
    return baseConfig;
}

// Update .env file with flexible CORS origins
function updateEnvironmentFile(origin, dryRun = false) {
    const envPath = path.join(process.cwd(), '.env');
    
    if (!fs.existsSync(envPath)) {
        console.log('⚠️  .env file not found, skipping global environment update');
        return;
    }
    
    try {
        let envContent = fs.readFileSync(envPath, 'utf8');
        
        // Update CORS_ORIGINS to include wildcard patterns
        const newOrigins = [
            'http://localhost:3000',
            'http://localhost:*',
            'https://localhost:*',
            origin
        ].filter((origin, index, self) => self.indexOf(origin) === index); // Remove duplicates
        
        const corsOriginsLine = `CORS_ORIGINS=${newOrigins.join(',')}`;
        
        if (envContent.includes('CORS_ORIGINS=')) {
            envContent = envContent.replace(/CORS_ORIGINS=.*/g, corsOriginsLine);
        } else {
            envContent += `\n${corsOriginsLine}\n`;
        }
        
        // Ensure CORS_HEADERS is set to wildcard for flexibility
        if (!envContent.includes('CORS_HEADERS=')) {
            envContent += `CORS_HEADERS=*\n`;
        }
        
        if (dryRun) {
            console.log('📝 Would update .env file with:');
            console.log(`   ${corsOriginsLine}`);
            console.log('   CORS_HEADERS=*');
        } else {
            fs.writeFileSync(envPath, envContent);
            console.log('✅ Updated .env file with flexible CORS configuration');
        }
    } catch (error) {
        console.error('❌ Failed to update .env file:', error.message);
    }
}

async function getJWTToken(apiBase) {
    console.log('🔐 Getting JWT token for authentication...');
    
    const loginData = {
        email: 'admin@cloudbox.local',
        password: 'admin123'
    };
    
    try {
        const response = await fetch(`${apiBase}/api/v1/auth/login`, {
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
        console.log(`   1. CloudBox backend is running on ${apiBase}`);
        console.log('   2. Admin credentials are correct');
        console.log('   3. Database is accessible');
        throw error;
    }
}

async function updateProjectCORS(apiBase, token, projectId, corsConfig, dryRun = false) {
    console.log(`🔧 ${dryRun ? 'Would update' : 'Updating'} CORS configuration for project ${projectId}...`);
    
    if (dryRun) {
        console.log('📋 CORS configuration that would be applied:');
        console.log(JSON.stringify(corsConfig, null, 2));
        return true;
    }
    
    try {
        const response = await fetch(`${apiBase}/api/v1/projects/${projectId}/cors`, {
            method: 'PUT',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(corsConfig)
        });
        
        if (response.ok) {
            const result = await response.json();
            console.log('✅ CORS configuration updated successfully');
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

async function testCORS(apiBase, origin, projectId) {
    console.log('\n🧪 Testing CORS configuration...');
    
    try {
        // Test global CORS
        console.log('   Testing global endpoint CORS...');
        const globalResponse = await fetch(`${apiBase}/api/v1/auth/login`, {
            method: 'OPTIONS',
            headers: {
                'Origin': origin,
                'Access-Control-Request-Method': 'POST',
                'Access-Control-Request-Headers': 'Content-Type'
            }
        });
        
        console.log(`   Global CORS test: ${globalResponse.status}`);
        if (globalResponse.headers.get('Access-Control-Allow-Origin')) {
            console.log('   ✅ Global CORS working');
        } else {
            console.log('   ⚠️  Global CORS may need attention');
        }
        
        // Test project-specific CORS (if project exists)
        console.log(`   Testing project ${projectId} endpoint CORS...`);
        const projectResponse = await fetch(`${apiBase}/p/project${projectId}/api/data/test`, {
            method: 'OPTIONS',
            headers: {
                'Origin': origin,
                'Access-Control-Request-Method': 'GET',
                'Access-Control-Request-Headers': 'Content-Type,X-API-Key'
            }
        });
        
        console.log(`   Project CORS test: ${projectResponse.status}`);
        if (projectResponse.headers.get('Access-Control-Allow-Origin')) {
            console.log('   ✅ Project CORS working');
        } else {
            console.log('   ⚠️  Project CORS may need attention');
        }
        
    } catch (error) {
        console.error('❌ CORS testing failed:', error.message);
    }
}

async function main() {
    console.log('🚀 PhotoPortfolio CORS Configuration Setup\n');
    
    const args = parseArgs();
    
    // Determine origin
    let origin = args.origin;
    if (!origin) {
        if (args.port) {
            origin = `http://localhost:${args.port}`;
        } else {
            // Auto-detect PhotoPortfolio configuration
            console.log('🔍 Auto-detecting PhotoPortfolio configuration...');
            const configs = detectPhotoPortfolioConfig();
            
            if (configs.length > 0) {
                const config = configs[0]; // Use first found config
                origin = `http://localhost:${config.port}`;
                console.log(`📍 Detected PhotoPortfolio at ${origin} (${config.path})`);
            } else {
                console.log('⚠️  Could not auto-detect PhotoPortfolio configuration');
                console.log('   Please specify --port or --origin manually');
                console.log('   Example: node setup-photoportfolio-cors.js --port 4041');
                process.exit(1);
            }
        }
    }
    
    console.log(`🎯 Target PhotoPortfolio origin: ${origin}`);
    console.log(`🎯 CloudBox backend: ${args.cloudboxUrl}`);
    console.log(`🎯 Project ID: ${args.projectId}`);
    
    if (args.dryRun) {
        console.log('🔍 DRY RUN MODE - No changes will be made\n');
    }
    
    try {
        // Generate CORS configuration
        const corsConfig = generateCORSConfig(origin);
        
        // Update global environment if requested
        if (args.updateGlobal) {
            console.log('\n📝 Updating global CORS configuration...');
            updateEnvironmentFile(origin, args.dryRun);
        }
        
        // Authenticate and update project CORS
        if (!args.dryRun) {
            const token = await getJWTToken(args.cloudboxUrl);
            await updateProjectCORS(args.cloudboxUrl, token, args.projectId, corsConfig, args.dryRun);
        } else {
            await updateProjectCORS(args.cloudboxUrl, null, args.projectId, corsConfig, args.dryRun);
        }
        
        // Test CORS configuration
        if (!args.dryRun) {
            await testCORS(args.cloudboxUrl, origin, args.projectId);
        }
        
        console.log('\n🎉 PhotoPortfolio CORS setup completed successfully!');
        console.log('\n📝 Configuration Summary:');
        console.log(`   Origin: ${origin}`);
        console.log(`   Project ID: ${args.projectId}`);
        console.log(`   Backend: ${args.cloudboxUrl}`);
        console.log('\n📌 Next Steps:');
        console.log(`   1. Start PhotoPortfolio on ${origin}`);
        console.log(`   2. Test API calls from PhotoPortfolio`);
        console.log('   3. CORS should now work automatically');
        
    } catch (error) {
        console.error('\n💥 Setup failed:', error.message);
        process.exit(1);
    }
}

// Run the script
if (require.main === module) {
    main();
}

module.exports = {
    parseArgs,
    detectPhotoPortfolioConfig,
    generateCORSConfig,
    updateEnvironmentFile,
    getJWTToken,
    updateProjectCORS,
    testCORS
};