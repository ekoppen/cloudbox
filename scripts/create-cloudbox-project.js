#!/usr/bin/env node

/**
 * Universal CloudBox Project Creator
 * 
 * Creates any type of application with CloudBox backend - similar to create-react-app
 * but for full-stack applications with automatic backend setup.
 */

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');
const readline = require('readline');

// ANSI color codes
const colors = {
  green: '\x1b[32m',
  red: '\x1b[31m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  magenta: '\x1b[35m',
  cyan: '\x1b[36m',
  white: '\x1b[37m',
  reset: '\x1b[0m',
  bold: '\x1b[1m'
};

// Available project templates
const templates = {
  photoportfolio: {
    name: 'Photo Portfolio',
    description: 'Photography portfolio with galleries, albums, and content management',
    repo: 'https://github.com/your-org/cloudbox-photoportfolio-template.git',
    features: ['Image Management', 'Album Creation', 'Content Pages', 'Analytics', 'SEO']
  },
  blog: {
    name: 'Blog/CMS',
    description: 'Content management system for blogs and articles',
    repo: 'https://github.com/your-org/cloudbox-blog-template.git',
    features: ['Post Management', 'Categories', 'Tags', 'Comments', 'SEO']
  },
  ecommerce: {
    name: 'E-commerce Store',
    description: 'Online store with products, orders, and customer management',
    repo: 'https://github.com/your-org/cloudbox-ecommerce-template.git',
    features: ['Product Catalog', 'Order Management', 'Customer Accounts', 'Inventory']
  },
  saas: {
    name: 'SaaS Application',
    description: 'Software as a Service with subscriptions, teams, and usage tracking',
    repo: 'https://github.com/your-org/cloudbox-saas-template.git',
    features: ['Subscription Management', 'Team Collaboration', 'Usage Tracking', 'Billing']
  },
  crm: {
    name: 'CRM System',
    description: 'Customer relationship management with contacts, deals, and activities',
    repo: 'https://github.com/your-org/cloudbox-crm-template.git',
    features: ['Contact Management', 'Deal Pipeline', 'Activity Tracking', 'Reporting']
  }
};

let config = {
  projectName: '',
  template: '',
  cloudboxEndpoint: 'http://localhost:8080',
  projectSlug: '',
  apiKey: '',
  setupComplete: false
};

const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout
});

// Utility functions
function log(message, color = 'white') {
  console.log(`${colors[color]}${message}${colors.reset}`);
}

function success(message) {
  log(`✅ ${message}`, 'green');
}

function error(message) {
  log(`❌ ${message}`, 'red');
}

function warning(message) {
  log(`⚠️  ${message}`, 'yellow');
}

function info(message) {
  log(`ℹ️  ${message}`, 'blue');
}

function question(prompt) {
  return new Promise((resolve) => {
    rl.question(`${colors.cyan}${prompt}${colors.reset}`, resolve);
  });
}

async function welcome() {
  console.clear();
  log(`
${colors.bold}${colors.magenta}
╔══════════════════════════════════════════╗
║          CloudBox Project Creator        ║
║     Full-Stack Apps in Minutes          ║
╚══════════════════════════════════════════╝
${colors.reset}

${colors.cyan}Welcome to CloudBox Project Creator!${colors.reset}

Create full-stack applications with automatic backend setup.
Similar to create-react-app but with CloudBox providing:

🚀 Automatic database collections
🔧 Ready-to-use API endpoints  
🔐 Authentication & user management
📊 Analytics & monitoring
☁️  Cloud deployment ready

`, 'white');
}

async function selectTemplate() {
  log('\\n📋 Available Project Templates', 'bold');
  
  Object.keys(templates).forEach((key, index) => {
    const template = templates[key];
    log(`\\n${index + 1}. ${colors.bold}${template.name}${colors.reset}`);
    log(`   ${template.description}`);
    log(`   Features: ${template.features.join(', ')}`, 'cyan');
  });

  while (!config.template) {
    const choice = await question('\\nSelect a template (1-' + Object.keys(templates).length + '): ');
    const templateKey = Object.keys(templates)[parseInt(choice) - 1];
    
    if (templateKey && templates[templateKey]) {
      config.template = templateKey;
      success(`Selected: ${templates[templateKey].name}`);
    } else {
      error('Invalid selection. Please try again.');
    }
  }
}

async function getProjectDetails() {
  log('\\n📝 Project Configuration', 'bold');
  
  // Get project name
  while (!config.projectName) {
    const name = await question('Project name (e.g., my-awesome-app): ');
    if (name.trim() && /^[a-zA-Z0-9-_]+$/.test(name.trim())) {
      config.projectName = name.trim();
      config.projectSlug = config.projectName.toLowerCase().replace(/[^a-z0-9-]/g, '-');
    } else {
      error('Project name can only contain letters, numbers, hyphens, and underscores.');
    }
  }

  // Get CloudBox details
  const endpoint = await question(`CloudBox endpoint (${config.cloudboxEndpoint}): `);
  if (endpoint.trim()) {
    config.cloudboxEndpoint = endpoint.trim();
  }

  // Get API key
  while (!config.apiKey) {
    const apiKey = await question('CloudBox API key: ');
    if (apiKey.trim()) {
      config.apiKey = apiKey.trim();
    } else {
      error('API key is required!');
    }
  }

  info(`\\n✨ Configuration:`);
  info(`   Template: ${templates[config.template].name}`);
  info(`   Name: ${config.projectName}`);
  info(`   Slug: ${config.projectSlug}`);
  info(`   Endpoint: ${config.cloudboxEndpoint}`);
  info(`   API Key: ${'*'.repeat(config.apiKey.length - 4)}${config.apiKey.slice(-4)}`);
}

async function createProject() {
  log('\\n🚀 Creating project...', 'bold');
  
  try {
    // Create project directory
    if (fs.existsSync(config.projectName)) {
      const overwrite = await question(`Directory "${config.projectName}" already exists. Overwrite? (y/N): `);
      if (overwrite.toLowerCase() !== 'y') {
        throw new Error('Project creation cancelled.');
      }
      fs.rmSync(config.projectName, { recursive: true, force: true });
    }

    info('Creating project directory...');
    fs.mkdirSync(config.projectName);
    process.chdir(config.projectName);

    // For now, copy from local template (in production, clone from git)
    const templatePath = path.join(__dirname, '..', 'templates', config.template);
    if (fs.existsSync(templatePath)) {
      info('Copying template files...');
      copyTemplate(templatePath, '.');
    } else {
      // Fallback: create basic structure
      info('Creating basic project structure...');
      createBasicStructure();
    }

    success('Project files created!');

  } catch (err) {
    error(`Failed to create project: ${err.message}`);
    throw err;
  }
}

function copyTemplate(src, dest) {
  const items = fs.readdirSync(src);
  
  items.forEach(item => {
    const srcPath = path.join(src, item);
    const destPath = path.join(dest, item);
    
    if (fs.statSync(srcPath).isDirectory()) {
      fs.mkdirSync(destPath, { recursive: true });
      copyTemplate(srcPath, destPath);
    } else {
      fs.copyFileSync(srcPath, destPath);
    }
  });
}

function createBasicStructure() {
  // Create basic React + TypeScript structure
  const packageJson = {
    name: config.projectName,
    version: '0.1.0',
    type: 'module',
    scripts: {
      dev: 'vite',
      build: 'vite build',
      preview: 'vite preview',
      setup: 'node scripts/setup-cloudbox.js',
      'setup:cloudbox': 'node scripts/setup-cloudbox.js'
    },
    dependencies: {
      react: '^18.3.1',
      'react-dom': '^18.3.1',
      axios: '^1.9.0',
      '@tanstack/react-query': '^5.80.7'
    },
    devDependencies: {
      '@types/react': '^18.3.3',
      '@types/react-dom': '^18.3.0',
      '@vitejs/plugin-react': '^4.0.0',
      typescript: '^5.0.0',
      vite: '^5.0.0',
      'node-fetch': '^3.3.2'
    }
  };

  fs.writeFileSync('package.json', JSON.stringify(packageJson, null, 2));

  // Create basic folders
  fs.mkdirSync('src', { recursive: true });
  fs.mkdirSync('scripts', { recursive: true });
  fs.mkdirSync('public', { recursive: true });
}

async function setupCloudBox() {
  log('\\n☁️  Setting up CloudBox backend...', 'bold');
  
  try {
    // Make API call to setup template
    const response = await fetch(
      `${config.cloudboxEndpoint}/p/${config.projectSlug}/api/templates/${config.template}/setup`,
      {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-API-Key': config.apiKey,
        },
        body: JSON.stringify({
          template: config.template,
          version: '1.0.0'
        })
      }
    );

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(`CloudBox setup failed: ${errorData.error || response.statusText}`);
    }

    const result = await response.json();
    
    success('CloudBox backend configured!');
    info(`Collections created: ${result.collections_created || 'N/A'}`);
    
    if (result.errors && result.errors.length > 0) {
      warning('Some issues occurred:');
      result.errors.forEach(err => warning(`  - ${err}`));
    }

    return result;

  } catch (err) {
    error(`CloudBox setup failed: ${err.message}`);
    throw err;
  }
}

async function generateEnvironment() {
  log('\\n📝 Creating environment configuration...', 'bold');
  
  const envContent = `# CloudBox Configuration
# Generated by CloudBox Project Creator

# CloudBox Settings
VITE_CLOUDBOX_ENDPOINT=${config.cloudboxEndpoint}
VITE_CLOUDBOX_PROJECT_SLUG=${config.projectSlug}
VITE_CLOUDBOX_API_KEY=${config.apiKey}

# API Configuration
VITE_API_URL=${config.cloudboxEndpoint}/p/${config.projectSlug}/api

# Template: ${config.template}
# Generated: ${new Date().toISOString()}
`;

  try {
    fs.writeFileSync('.env', envContent);
    fs.writeFileSync('.env.local', envContent);
    
    const exampleContent = envContent.replace(config.apiKey, 'your_api_key_here');
    fs.writeFileSync('.env.example', exampleContent);
    
    success('Environment files created');

  } catch (err) {
    error(`Failed to create environment files: ${err.message}`);
  }
}

async function generateSetupScript() {
  log('\\n🔧 Creating setup script...', 'bold');
  
  const setupScript = `#!/usr/bin/env node

/**
 * ${templates[config.template].name} CloudBox Setup Script
 * Generated by CloudBox Project Creator
 */

const { execSync } = require('child_process');

const config = {
  template: '${config.template}',
  projectSlug: '${config.projectSlug}',
  cloudboxEndpoint: '${config.cloudboxEndpoint}',
  apiKey: process.env.VITE_CLOUDBOX_API_KEY || '${config.apiKey}'
};

async function setupProject() {
  console.log('🚀 Setting up ${templates[config.template].name} project...');
  
  try {
    const response = await fetch(
      \`\${config.cloudboxEndpoint}/p/\${config.projectSlug}/api/templates/\${config.template}/setup\`,
      {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-API-Key': config.apiKey,
        },
        body: JSON.stringify({
          template: config.template,
          version: '1.0.0'
        })
      }
    );

    if (!response.ok) {
      throw new Error(\`Setup failed: \${response.statusText}\`);
    }

    const result = await response.json();
    console.log('✅ Setup completed!', result);
    
  } catch (err) {
    console.error('❌ Setup failed:', err.message);
    process.exit(1);
  }
}

// Add fetch polyfill for Node.js
if (typeof fetch === 'undefined') {
  global.fetch = require('node-fetch');
}

setupProject();
`;

  try {
    fs.writeFileSync('scripts/setup-cloudbox.js', setupScript);
    fs.chmodSync('scripts/setup-cloudbox.js', '755');
    success('Setup script created');

  } catch (err) {
    error(`Failed to create setup script: ${err.message}`);
  }
}

async function installDependencies() {
  log('\\n📦 Installing dependencies...', 'bold');
  
  try {
    info('Running npm install...');
    execSync('npm install', { stdio: 'inherit' });
    success('Dependencies installed!');

  } catch (err) {
    warning('Failed to install dependencies automatically.');
    info('Run "npm install" manually after setup completes.');
  }
}

async function generateQuickStart() {
  log('\\n📖 Creating documentation...', 'bold');
  
  const quickStartContent = `# ${config.projectName}

A ${templates[config.template].name} built with CloudBox.

## Features

${templates[config.template].features.map(f => `- ✅ ${f}`).join('\\n')}

## Quick Start

### 1. Development

\`\`\`bash
npm run dev
\`\`\`

Visit: http://localhost:5173

### 2. Admin Panel

Navigate to: http://localhost:5173/admin
Login with your CloudBox project credentials.

### 3. API Endpoints

Base URL: \`${config.cloudboxEndpoint}/p/${config.projectSlug}/api\`

### 4. Re-run Setup

If you need to reconfigure the CloudBox backend:

\`\`\`bash
npm run setup
\`\`\`

## Environment

\`\`\`bash
VITE_CLOUDBOX_ENDPOINT=${config.cloudboxEndpoint}
VITE_CLOUDBOX_PROJECT_SLUG=${config.projectSlug}
VITE_API_URL=${config.cloudboxEndpoint}/p/${config.projectSlug}/api
\`\`\`

## Deployment

1. Build: \`npm run build\`
2. Update environment for production
3. Deploy to your hosting platform

---
Generated by CloudBox Project Creator
Template: ${config.template}
Date: ${new Date().toISOString()}
`;

  try {
    fs.writeFileSync('README.md', quickStartContent);
    success('Documentation created');

  } catch (err) {
    error(`Failed to create documentation: ${err.message}`);
  }
}

async function complete() {
  log('\\n🎉 Project created successfully!', 'bold');
  
  success(`${templates[config.template].name} project "${config.projectName}" is ready!`);
  
  log('\\n📋 Next steps:', 'cyan');
  log(`   cd ${config.projectName}`, 'white');
  log('   npm run dev', 'white');
  log('\\n🌐 Then visit:', 'cyan');
  log('   http://localhost:5173 - Your app', 'white');
  log('   http://localhost:5173/admin - Admin panel', 'white');

  log('\\n🔧 Available commands:', 'cyan');
  log('   npm run dev - Start development server', 'white');
  log('   npm run build - Build for production', 'white');
  log('   npm run setup - Reconfigure CloudBox backend', 'white');

  rl.close();
}

// Main execution
async function main() {
  try {
    await welcome();
    await selectTemplate();
    await getProjectDetails();
    await createProject();
    await setupCloudBox();
    await generateEnvironment();
    await generateSetupScript();
    await installDependencies();
    await generateQuickStart();
    await complete();
  } catch (err) {
    error(`\\nProject creation failed: ${err.message}`);
    process.exit(1);
  }
}

// Add fetch polyfill for Node.js
if (typeof fetch === 'undefined') {
  global.fetch = require('node-fetch');
}

// Handle command line arguments
const args = process.argv.slice(2);
if (args.length > 0) {
  config.projectName = args[0];
  if (args[1] && templates[args[1]]) {
    config.template = args[1];
  }
}

// Run the script
main().catch(console.error);