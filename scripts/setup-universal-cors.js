#!/usr/bin/env node

/**
 * Universal CloudBox CORS Setup Script
 * 
 * Automatically configures CORS for any client application framework.
 * Supports React, Vue, Angular, Svelte, and Vanilla JavaScript.
 * 
 * Features:
 * - Automatic framework detection
 * - Port auto-discovery
 * - Environment-specific configuration
 * - Framework-specific helper generation
 * - Production deployment guidance
 * 
 * Usage:
 *   node setup-universal-cors.js [options]
 * 
 * Options:
 *   --framework <name>       Force framework (react, vue, angular, svelte, vanilla)
 *   --port <port>           Override port detection
 *   --origin <origin>       Full origin URL (overrides port)
 *   --project-id <id>       CloudBox project ID (default: auto-detect)
 *   --cloudbox-url <url>    CloudBox backend URL (default: http://localhost:8080)
 *   --environment <env>     Environment (development, staging, production)
 *   --create-helpers        Generate framework-specific helper files
 *   --update-global         Also update global CORS configuration
 *   --dry-run              Show what would be done without making changes
 *   --verbose              Show detailed output
 */

const fs = require('fs');
const path = require('path');

// Supported frameworks and their configurations
const FRAMEWORKS = {
    react: {
        name: 'React',
        defaultPorts: [3000, 3001],
        configFiles: ['package.json'],
        detectionPatterns: ['react', 'react-dom', 'react-scripts', '@vitejs/plugin-react'],
        devScriptPatterns: ['react-scripts start', 'vite', 'webpack-dev-server'],
        envFile: '.env.local',
        helperFile: 'src/cloudbox.js'
    },
    vue: {
        name: 'Vue.js',
        defaultPorts: [8080, 8081, 3000],
        configFiles: ['package.json', 'vue.config.js'],
        detectionPatterns: ['vue', '@vue/cli-service', 'vite', '@vitejs/plugin-vue'],
        devScriptPatterns: ['vue-cli-service serve', 'vite'],
        envFile: '.env.local',
        helperFile: 'src/cloudbox.js'
    },
    angular: {
        name: 'Angular',
        defaultPorts: [4200, 4201],
        configFiles: ['package.json', 'angular.json'],
        detectionPatterns: ['@angular/core', '@angular/cli'],
        devScriptPatterns: ['ng serve', '@angular/cli'],
        envFile: 'src/environments/environment.ts',
        helperFile: 'src/app/services/cloudbox.service.ts'
    },
    svelte: {
        name: 'Svelte',
        defaultPorts: [5000, 5001, 3000],
        configFiles: ['package.json', 'svelte.config.js'],
        detectionPatterns: ['svelte', '@sveltejs/kit', '@sveltejs/vite-plugin-svelte'],
        devScriptPatterns: ['svelte-kit dev', 'vite dev'],
        envFile: '.env.local',
        helperFile: 'src/lib/cloudbox.js'
    },
    vanilla: {
        name: 'Vanilla JavaScript',
        defaultPorts: [8080, 3000, 5000],
        configFiles: ['package.json', 'index.html'],
        detectionPatterns: [],
        devScriptPatterns: ['http-server', 'live-server', 'serve'],
        envFile: '.env',
        helperFile: 'js/cloudbox.js'
    }
};

class UniversalClientSetup {
    constructor() {
        this.args = this.parseArgs();
        this.detectedProject = null;
        this.corsConfig = null;
    }

    parseArgs() {
        const args = {
            framework: null,
            port: null,
            origin: null,
            projectId: null,
            cloudboxUrl: 'http://localhost:8080',
            environment: 'development',
            createHelpers: false,
            updateGlobal: false,
            dryRun: false,
            verbose: false
        };

        for (let i = 2; i < process.argv.length; i++) {
            const arg = process.argv[i];
            const nextArg = process.argv[i + 1];

            switch (arg) {
                case '--framework':
                    args.framework = nextArg?.toLowerCase();
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
                case '--project-id':
                    args.projectId = parseInt(nextArg) || nextArg;
                    i++;
                    break;
                case '--cloudbox-url':
                    args.cloudboxUrl = nextArg;
                    i++;
                    break;
                case '--environment':
                    args.environment = nextArg;
                    i++;
                    break;
                case '--create-helpers':
                    args.createHelpers = true;
                    break;
                case '--update-global':
                    args.updateGlobal = true;
                    break;
                case '--dry-run':
                    args.dryRun = true;
                    break;
                case '--verbose':
                    args.verbose = true;
                    break;
                case '--help':
                    this.showHelp();
                    process.exit(0);
            }
        }

        return args;
    }

    showHelp() {
        console.log(`
Universal CloudBox CORS Setup Script

Usage: node setup-universal-cors.js [options]

Options:
  --framework <name>       Force framework (react, vue, angular, svelte, vanilla)
  --port <port>           Override port detection
  --origin <origin>       Full origin URL (overrides port)
  --project-id <id>       CloudBox project ID (default: auto-detect)
  --cloudbox-url <url>    CloudBox backend URL (default: http://localhost:8080)
  --environment <env>     Environment (development, staging, production)
  --create-helpers        Generate framework-specific helper files
  --update-global         Also update global CORS configuration
  --dry-run              Show what would be done without making changes
  --verbose              Show detailed output
  --help                 Show this help message

Examples:
  # Auto-detect everything
  node setup-universal-cors.js
  
  # React app on custom port
  node setup-universal-cors.js --framework react --port 3001
  
  # Production setup
  node setup-universal-cors.js --environment production --origin https://app.example.com
  
  # Create helper files and update global config
  node setup-universal-cors.js --create-helpers --update-global

Supported Frameworks:
  - React (Create React App, Vite, Next.js)
  - Vue.js (Vue CLI, Vite, Nuxt.js)
  - Angular (Angular CLI)
  - Svelte (SvelteKit, Vite)
  - Vanilla JavaScript (any static server)
        `);
    }

    async detectProject() {
        console.log('🔍 Auto-detecting project configuration...');

        const detectionResults = await Promise.all([
            this.detectFramework('react'),
            this.detectFramework('vue'),
            this.detectFramework('angular'),
            this.detectFramework('svelte'),
            this.detectFramework('vanilla')
        ]);

        const detected = detectionResults.filter(result => result.detected);
        
        if (detected.length === 0) {
            throw new Error('No supported framework detected. Use --framework to specify manually.');
        }

        // Use the first detected framework or the one specified
        let selectedProject = detected[0];
        if (this.args.framework) {
            const forced = detected.find(p => p.framework === this.args.framework);
            if (forced) {
                selectedProject = forced;
            } else {
                console.warn(`⚠️  Framework '${this.args.framework}' not detected, using auto-detected: ${selectedProject.framework}`);
            }
        }

        this.detectedProject = selectedProject;
        
        if (this.args.verbose) {
            console.log('📋 Detected project configuration:');
            console.log(JSON.stringify(selectedProject, null, 2));
        }

        return selectedProject;
    }

    async detectFramework(frameworkName) {
        const framework = FRAMEWORKS[frameworkName];
        if (!framework) return { detected: false, framework: frameworkName };

        const result = {
            detected: false,
            framework: frameworkName,
            name: framework.name,
            port: null,
            devCommand: null,
            buildCommand: null,
            configFiles: [],
            confidence: 0
        };

        // Check package.json
        const packagePath = './package.json';
        if (fs.existsSync(packagePath)) {
            try {
                const pkg = JSON.parse(fs.readFileSync(packagePath, 'utf8'));
                
                // Check dependencies for framework indicators
                const allDeps = {
                    ...pkg.dependencies,
                    ...pkg.devDependencies
                };

                let matches = 0;
                for (const pattern of framework.detectionPatterns) {
                    if (allDeps[pattern]) {
                        matches++;
                        result.confidence += 25;
                    }
                }

                if (matches > 0) {
                    result.detected = true;
                    result.configFiles.push('package.json');
                    
                    // Extract port from dev script
                    if (pkg.scripts) {
                        result.devCommand = pkg.scripts.dev || pkg.scripts.start || pkg.scripts.serve;
                        result.buildCommand = pkg.scripts.build;
                        
                        const devScript = result.devCommand || '';
                        const portMatch = devScript.match(/(?:--port|port:|PORT=)\s*(\d+)/i);
                        if (portMatch) {
                            result.port = parseInt(portMatch[1]);
                            result.confidence += 20;
                        }
                    }
                }
            } catch (error) {
                // Ignore package.json parsing errors
            }
        }

        // Check framework-specific config files
        for (const configFile of framework.configFiles) {
            if (configFile !== 'package.json' && fs.existsSync(configFile)) {
                result.configFiles.push(configFile);
                result.confidence += 15;
                
                // Try to extract port from config files
                if (!result.port) {
                    try {
                        const content = fs.readFileSync(configFile, 'utf8');
                        const portMatch = content.match(/port[:\s]*(\d+)/i);
                        if (portMatch) {
                            result.port = parseInt(portMatch[1]);
                        }
                    } catch (error) {
                        // Ignore config file reading errors
                    }
                }
            }
        }

        // Use default port if none found
        if (!result.port && result.detected) {
            result.port = framework.defaultPorts[0];
        }

        return result;
    }

    generateOrigin() {
        if (this.args.origin) {
            return this.args.origin;
        }

        const port = this.args.port || this.detectedProject?.port || 3000;
        const protocol = this.args.environment === 'production' ? 'https' : 'http';
        const host = this.args.environment === 'production' ? 'your-domain.com' : 'localhost';
        
        return `${protocol}://${host}:${port}`;
    }

    generateCORSConfig() {
        const origin = this.generateOrigin();
        const framework = FRAMEWORKS[this.detectedProject?.framework || 'vanilla'];
        
        const baseOrigins = [origin];
        
        // Add development-friendly origins
        if (this.args.environment === 'development') {
            baseOrigins.push(
                'http://localhost:*',
                'https://localhost:*',
                ...framework.defaultPorts.map(port => `http://localhost:${port}`)
            );
        }
        
        this.corsConfig = {
            allowed_origins: [...new Set(baseOrigins)],
            allowed_methods: ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'OPTIONS'],
            allowed_headers: [
                'Accept', 'Content-Type', 'Content-Length', 'Accept-Encoding',
                'Authorization', 'X-CSRF-Token', 'X-API-Key',
                'Cache-Control', 'X-Requested-With',
                'Session-Token', 'session-token',
                'X-Session-Token', 'x-session-token',
                'X-Project-ID', 'X-Project-Token',
                'Project-ID', 'Project-Token'
            ],
            allow_credentials: true,
            max_age: 3600
        };

        return this.corsConfig;
    }

    async setupCORS() {
        console.log('🔧 Setting up CORS configuration...');
        
        const corsConfig = this.generateCORSConfig();
        
        if (this.args.dryRun) {
            console.log('📋 CORS configuration that would be applied:');
            console.log(JSON.stringify(corsConfig, null, 2));
            return true;
        }

        // Update CloudBox CORS
        const success = await this.updateCloudBoxCORS(corsConfig);
        
        if (success) {
            console.log('✅ CORS configuration updated successfully');
        }
        
        return success;
    }

    async updateCloudBoxCORS(corsConfig) {
        try {
            // Get JWT token for authentication
            const token = await this.getJWTToken();
            
            // Determine project ID
            const projectId = this.args.projectId || await this.detectProjectId();
            
            // Update project-specific CORS
            const response = await fetch(`${this.args.cloudboxUrl}/api/v1/projects/${projectId}/cors`, {
                method: 'PUT',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(corsConfig)
            });

            if (response.ok) {
                return true;
            } else {
                const error = await response.text();
                console.error('❌ Failed to update CORS:', error);
                return false;
            }
        } catch (error) {
            console.error('❌ Error updating CORS:', error.message);
            return false;
        }
    }

    async getJWTToken() {
        const loginData = {
            email: 'admin@cloudbox.local',
            password: 'admin123'
        };

        const response = await fetch(`${this.args.cloudboxUrl}/api/v1/auth/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(loginData)
        });

        if (!response.ok) {
            throw new Error(`Authentication failed: ${response.status}`);
        }

        const data = await response.json();
        return data.token;
    }

    async detectProjectId() {
        // Try to detect project ID from existing config files
        const possibleFiles = ['.env', '.env.local', 'src/config.js'];
        
        for (const file of possibleFiles) {
            if (fs.existsSync(file)) {
                try {
                    const content = fs.readFileSync(file, 'utf8');
                    const match = content.match(/PROJECT_ID[=:\s]*(\d+)/i);
                    if (match) {
                        return parseInt(match[1]);
                    }
                } catch (error) {
                    // Ignore file reading errors
                }
            }
        }
        
        // Default to project ID 2 (common for PhotoPortfolio)
        return 2;
    }

    async createFrameworkHelpers() {
        if (!this.args.createHelpers) return;
        
        console.log('📝 Creating framework-specific helper files...');
        
        const framework = FRAMEWORKS[this.detectedProject?.framework || 'vanilla'];
        const origin = this.generateOrigin();
        
        if (this.args.dryRun) {
            console.log(`Would create helper file: ${framework.helperFile}`);
            return;
        }

        const helperContent = this.generateHelperContent(framework, origin);
        const helperDir = path.dirname(framework.helperFile);
        
        // Create directory if it doesn't exist
        if (!fs.existsSync(helperDir)) {
            fs.mkdirSync(helperDir, { recursive: true });
        }
        
        fs.writeFileSync(framework.helperFile, helperContent);
        console.log(`✅ Created helper file: ${framework.helperFile}`);
        
        // Create environment configuration
        await this.createEnvironmentConfig(framework, origin);
    }

    generateHelperContent(framework, origin) {
        const projectId = this.args.projectId || 2;
        
        const templates = {
            react: `import { CloudBoxClient } from '@ekoppen/cloudbox-sdk';

const client = new CloudBoxClient({
  projectId: process.env.REACT_APP_PROJECT_ID || ${projectId},
  apiKey: process.env.REACT_APP_API_KEY || 'your-api-key',
  endpoint: process.env.REACT_APP_CLOUDBOX_ENDPOINT || '${this.args.cloudboxUrl}',
  authMode: 'project'
});

// Enhanced error handling with CORS troubleshooting
client.onCORSError = (error) => {
  console.error('CORS Configuration Error:', error.message);
  console.log('💡 Quick fixes:');
  console.log('1. Run: node scripts/setup-universal-cors.js --framework=react');
  console.log('2. Check CloudBox backend is running');
  console.log('3. Verify CORS_ORIGINS in CloudBox .env file');
};

export default client;
`,
            vue: `import { CloudBoxClient } from '@ekoppen/cloudbox-sdk';

const client = new CloudBoxClient({
  projectId: process.env.VUE_APP_PROJECT_ID || ${projectId},
  apiKey: process.env.VUE_APP_API_KEY || 'your-api-key',
  endpoint: process.env.VUE_APP_CLOUDBOX_ENDPOINT || '${this.args.cloudboxUrl}',
  authMode: 'project'
});

// Enhanced error handling
client.onCORSError = (error) => {
  console.error('CORS Configuration Error:', error.message);
  console.log('💡 Run: node scripts/setup-universal-cors.js --framework=vue');
};

export default client;
`,
            angular: `import { Injectable } from '@angular/core';
import { CloudBoxClient } from '@ekoppen/cloudbox-sdk';
import { environment } from '../environments/environment';

@Injectable({
  providedIn: 'root'
})
export class CloudBoxService {
  private client: CloudBoxClient;

  constructor() {
    this.client = new CloudBoxClient({
      projectId: environment.projectId || ${projectId},
      apiKey: environment.apiKey || 'your-api-key',
      endpoint: environment.cloudboxEndpoint || '${this.args.cloudboxUrl}',
      authMode: 'project'
    });

    // Enhanced error handling
    this.client.onCORSError = (error) => {
      console.error('CORS Configuration Error:', error.message);
      console.log('💡 Run: node scripts/setup-universal-cors.js --framework=angular');
    };
  }

  getClient(): CloudBoxClient {
    return this.client;
  }
}
`,
            svelte: `import { CloudBoxClient } from '@ekoppen/cloudbox-sdk';

const client = new CloudBoxClient({
  projectId: import.meta.env.VITE_PROJECT_ID || ${projectId},
  apiKey: import.meta.env.VITE_API_KEY || 'your-api-key',
  endpoint: import.meta.env.VITE_CLOUDBOX_ENDPOINT || '${this.args.cloudboxUrl}',
  authMode: 'project'
});

// Enhanced error handling
client.onCORSError = (error) => {
  console.error('CORS Configuration Error:', error.message);
  console.log('💡 Run: node scripts/setup-universal-cors.js --framework=svelte');
};

export default client;
`,
            vanilla: `// CloudBox SDK Configuration
import { CloudBoxClient } from '@ekoppen/cloudbox-sdk';

const client = new CloudBoxClient({
  projectId: ${projectId},
  apiKey: 'your-api-key',
  endpoint: '${this.args.cloudboxUrl}',
  authMode: 'project'
});

// Enhanced error handling
client.onCORSError = (error) => {
  console.error('CORS Configuration Error:', error.message);
  console.log('💡 Run: node scripts/setup-universal-cors.js --framework=vanilla');
};

export default client;
`
        };

        return templates[this.detectedProject?.framework] || templates.vanilla;
    }

    async createEnvironmentConfig(framework, origin) {
        const projectId = this.args.projectId || 2;
        
        const envConfigs = {
            react: `# CloudBox Configuration
REACT_APP_PROJECT_ID=${projectId}
REACT_APP_API_KEY=your-api-key-here
REACT_APP_CLOUDBOX_ENDPOINT=${this.args.cloudboxUrl}

# Development
REACT_APP_DEV_MODE=true
BROWSER=none
`,
            vue: `# CloudBox Configuration  
VUE_APP_PROJECT_ID=${projectId}
VUE_APP_API_KEY=your-api-key-here
VUE_APP_CLOUDBOX_ENDPOINT=${this.args.cloudboxUrl}
`,
            angular: `// src/environments/environment.ts
export const environment = {
  production: false,
  projectId: ${projectId},
  apiKey: 'your-api-key-here',
  cloudboxEndpoint: '${this.args.cloudboxUrl}'
};
`,
            svelte: `# CloudBox Configuration
VITE_PROJECT_ID=${projectId}
VITE_API_KEY=your-api-key-here
VITE_CLOUDBOX_ENDPOINT=${this.args.cloudboxUrl}
`,
            vanilla: `// Add to your HTML or config file:
const CLOUDBOX_CONFIG = {
  projectId: ${projectId},
  apiKey: 'your-api-key-here',
  endpoint: '${this.args.cloudboxUrl}'
};
`
        };

        const envContent = envConfigs[this.detectedProject?.framework] || envConfigs.vanilla;
        
        if (framework.envFile && !framework.envFile.includes('.ts')) {
            fs.writeFileSync(framework.envFile, envContent);
            console.log(`✅ Created environment config: ${framework.envFile}`);
        } else if (this.args.verbose) {
            console.log('📋 Environment configuration:');
            console.log(envContent);
        }
    }

    async updateGlobalEnvironment() {
        if (!this.args.updateGlobal) return;

        console.log('🌍 Updating global CORS configuration...');
        
        const origin = this.generateOrigin();
        const envPath = path.join(process.cwd(), '.env');
        
        if (!fs.existsSync(envPath)) {
            console.log('⚠️  .env file not found, skipping global environment update');
            return;
        }

        try {
            let envContent = fs.readFileSync(envPath, 'utf8');
            
            const newOrigins = [
                'http://localhost:3000',
                'http://localhost:*',
                'https://localhost:*',
                origin
            ].filter((origin, index, self) => self.indexOf(origin) === index);

            const corsOriginsLine = `CORS_ORIGINS=${newOrigins.join(',')}`;
            
            if (envContent.includes('CORS_ORIGINS=')) {
                envContent = envContent.replace(/CORS_ORIGINS=.*/g, corsOriginsLine);
            } else {
                envContent += `\n${corsOriginsLine}\n`;
            }

            if (!envContent.includes('CORS_HEADERS=')) {
                envContent += `CORS_HEADERS=*\n`;
            }

            if (this.args.dryRun) {
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

    async testCORS() {
        if (this.args.dryRun) return;

        console.log('🧪 Testing CORS configuration...');
        
        const origin = this.generateOrigin();
        const projectId = this.args.projectId || await this.detectProjectId();
        
        try {
            // Test global CORS
            const globalResponse = await fetch(`${this.args.cloudboxUrl}/api/v1/auth/login`, {
                method: 'OPTIONS',
                headers: {
                    'Origin': origin,
                    'Access-Control-Request-Method': 'POST',
                    'Access-Control-Request-Headers': 'Content-Type'
                }
            });

            console.log(`   Global CORS test: ${globalResponse.status === 204 ? '✅ PASS' : '❌ FAIL'}`);

            // Test project-specific CORS
            const projectResponse = await fetch(`${this.args.cloudboxUrl}/p/${projectId}/api/collections`, {
                method: 'OPTIONS',
                headers: {
                    'Origin': origin,
                    'Access-Control-Request-Method': 'GET',
                    'Access-Control-Request-Headers': 'Content-Type,X-API-Key'
                }
            });

            console.log(`   Project CORS test: ${projectResponse.status === 204 ? '✅ PASS' : '❌ FAIL'}`);
            
        } catch (error) {
            console.error('❌ CORS testing failed:', error.message);
        }
    }

    async run() {
        try {
            console.log('🚀 Universal CloudBox CORS Setup\n');

            // Detect project
            await this.detectProject();
            
            console.log(`📋 Configuration Summary:`);
            console.log(`   Framework: ${this.detectedProject.name}`);
            console.log(`   Origin: ${this.generateOrigin()}`);
            console.log(`   Environment: ${this.args.environment}`);
            console.log(`   CloudBox: ${this.args.cloudboxUrl}`);
            
            if (this.args.dryRun) {
                console.log('🔍 DRY RUN MODE - No changes will be made\n');
            }

            // Setup CORS
            await this.setupCORS();

            // Create framework helpers
            await this.createFrameworkHelpers();

            // Update global environment
            await this.updateGlobalEnvironment();

            // Test CORS configuration
            await this.testCORS();

            console.log('\n🎉 Universal CORS setup completed successfully!');
            
            if (!this.args.dryRun) {
                console.log('\n📌 Next Steps:');
                console.log('   1. Update the API key in your environment configuration');
                console.log(`   2. Start your ${this.detectedProject.name} application`);
                console.log('   3. Test API calls - CORS should work automatically');
                console.log('   4. For production, run with --environment production');
            }

        } catch (error) {
            console.error('\n💥 Setup failed:', error.message);
            
            if (this.args.verbose) {
                console.error(error.stack);
            }
            
            console.log('\n💡 Troubleshooting:');
            console.log('   1. Ensure CloudBox backend is running');
            console.log('   2. Check your project structure and framework');
            console.log('   3. Run with --verbose for detailed output');
            console.log('   4. Use --dry-run to see what would be done');
            
            process.exit(1);
        }
    }
}

// Run the script
if (require.main === module) {
    const setup = new UniversalClientSetup();
    setup.run();
}

module.exports = UniversalClientSetup;