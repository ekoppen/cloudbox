# CloudBox Plugin System

The CloudBox Plugin System allows developers to extend CloudBox with custom functionality, similar to how WordPress plugins work, but specifically designed for modern development workflows.

## 🎯 Overview

CloudBox plugins can:
- Add new dashboard pages and components
- Extend project functionality (like Scripts, Custom APIs, etc.)
- Integrate with external services
- Add new database schemas
- Deploy custom functions and webhooks
- Create project templates and automation

## 🏗️ Plugin Architecture

### Plugin Structure
```
/cloudbox/plugins/your-plugin/
├── plugin.json          # Plugin metadata and configuration
├── index.js             # Main plugin logic (Node.js)
├── install.sh           # Installation script
├── components/          # Frontend components (Svelte/React)
│   └── Dashboard.svelte
├── templates/           # Project templates (JSON)
├── migrations/          # Database migrations
└── README.md           # Plugin documentation
```

### plugin.json Schema
```json
{
  "name": "your-plugin-name",
  "version": "1.0.0", 
  "description": "Plugin description",
  "author": "Your Name",
  "type": "dashboard-plugin",
  "main": "index.js",
  "dependencies": {
    "cloudbox-sdk": "^1.0.0"
  },
  "permissions": [
    "database:read",
    "database:write",
    "functions:deploy",
    "projects:manage"
  ],
  "ui": {
    "dashboard_tab": {
      "title": "Plugin Name",
      "icon": "terminal",
      "path": "/plugin-route"
    },
    "project_menu": {
      "title": "Project Feature",
      "icon": "database", 
      "path": "/project/{projectId}/feature"
    }
  }
}
```

## 🛠️ CloudBox SDK

### Plugin Base Class
```javascript
const { CloudBoxPlugin } = require('cloudbox-sdk');

class YourPlugin extends CloudBoxPlugin {
  constructor() {
    super('your-plugin-name');
  }

  async onInstall() {
    // Plugin installation logic
    await this.createPluginTables();
    this.registerRoutes();
    this.registerDashboardComponents();
  }

  async createPluginTables() {
    const db = this.getDatabase();
    await db.query(`
      CREATE SCHEMA IF NOT EXISTS your_plugin;
      CREATE TABLE IF NOT EXISTS your_plugin.data (
        id SERIAL PRIMARY KEY,
        content JSONB
      );
    `);
  }

  registerRoutes() {
    this.app.get('/api/plugins/your-plugin/data', this.getData.bind(this));
    this.app.post('/api/plugins/your-plugin/data', this.createData.bind(this));
  }

  registerDashboardComponents() {
    this.registerComponent('your-dashboard', {
      component: './components/Dashboard.svelte',
      route: '/your-plugin',
      menu: {
        title: 'Your Plugin',
        icon: 'plugin'
      }
    });
  }
}

module.exports = YourPlugin;
```

### Available Managers
```javascript
// Database access
const db = this.getDatabase();
const projectDb = this.getProjectDatabase(projectId);

// Function management
const functions = new FunctionManager(projectId);
await functions.deploy('function-name', options);

// Webhook management  
const webhooks = new WebhookManager(projectId);
await webhooks.create('webhook-name', config);

// Storage management
const storage = new StorageManager(projectId);
await storage.upload(file, bucket);
```

## 🎨 Frontend Integration

### Svelte Components
```svelte
<script>
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import Button from '$lib/components/ui/button.svelte';
  import Card from '$lib/components/ui/card.svelte';

  // Get project ID from route params
  $: projectId = $page.params.id;

  let data = [];

  async function loadData() {
    const response = await fetch(`/api/plugins/your-plugin/data/${projectId}`);
    const result = await response.json();
    if (result.success) {
      data = result.data;
    }
  }

  onMount(() => {
    loadData();
  });
</script>

<div class="space-y-6">
  <h1 class="text-2xl font-bold">Your Plugin</h1>
  
  <Card>
    <div class="p-6">
      <!-- Plugin UI here -->
    </div>
  </Card>
</div>
```

### Integration Points
- **Dashboard Navigation**: Global plugin access
- **Project Sidebar**: Project-specific features
- **Project Context**: Access to project data and settings
- **Theme Integration**: Automatic dark/light mode support

## 📦 Example Plugins

### 1. Script Runner Plugin
**Purpose**: Database scripts and project setup automation (like Supabase SQL Editor)

**Features**:
- SQL script execution
- JavaScript function deployment
- Project templates (Web App, AI Chat, E-commerce)
- Dependency management
- Execution history

**Location**: Project sidebar → Scripts

### 2. API Documentation Plugin
**Purpose**: Automatic API documentation generation

**Features**:
- Auto-generate docs from OpenAPI specs
- Interactive API testing
- Code examples in multiple languages
- Version management

### 3. Monitoring Plugin  
**Purpose**: Application performance monitoring

**Features**:
- Real-time metrics dashboard
- Error tracking and alerting
- Performance analytics
- Custom dashboards

### 4. Backup Plugin
**Purpose**: Automated backup management

**Features**:
- Scheduled database backups
- File system snapshots
- Cloud storage integration
- Backup verification

## 🚀 Plugin Development Workflow

### 1. Setup
```bash
# Create plugin directory
mkdir /path/to/cloudbox/plugins/your-plugin
cd /path/to/cloudbox/plugins/your-plugin

# Initialize plugin
npm init
npm install cloudbox-sdk
```

### 2. Development
```bash
# Create plugin files
touch plugin.json index.js install.sh
mkdir components templates

# Develop plugin logic
# Test with CloudBox development environment
```

### 3. Installation
```bash
# Run install script
./install.sh

# Restart CloudBox to load plugin
docker-compose restart backend
```

### 4. Testing
```bash
# Access plugin in dashboard
# Test all functionality
# Verify database schemas
# Check API endpoints
```

## 🔐 Security & Permissions

### Permission System
```json
{
  "permissions": [
    "database:read",        // Read database access
    "database:write",       // Write database access
    "functions:deploy",     // Deploy CloudBox functions
    "functions:manage",     // Manage existing functions
    "webhooks:create",      // Create webhooks
    "webhooks:manage",      // Manage webhooks
    "storage:read",         // Read storage access
    "storage:write",        // Write storage access
    "projects:read",        // Read project data
    "projects:manage",      // Manage project settings
    "users:read",           // Read user data
    "users:manage",         // Manage users
    "system:admin"          // System administration
  ]
}
```

### Best Practices
- **Principle of Least Privilege**: Request only needed permissions
- **Data Isolation**: Use plugin-specific schemas
- **Input Validation**: Validate all user inputs
- **Error Handling**: Graceful error handling and logging
- **Documentation**: Clear documentation and examples

## 📊 Plugin Marketplace (Future)

### Plugin Categories
- **Development Tools**: Code generators, linters, formatters
- **Database Tools**: Migration tools, query builders, admin panels  
- **API Tools**: Documentation, testing, monitoring
- **Deployment Tools**: CI/CD integration, infrastructure management
- **Analytics Tools**: Monitoring, reporting, dashboards
- **Integration Tools**: Third-party service integrations
- **Security Tools**: Vulnerability scanning, access control
- **Backup Tools**: Data backup and recovery solutions

### Distribution
- **Official Plugin Registry**: Verified, maintained plugins
- **Community Plugins**: User-contributed plugins
- **Private Plugins**: Organization-specific plugins
- **Plugin Templates**: Starter templates for common use cases

## 🤝 Contributing

### Guidelines
1. Follow CloudBox coding standards
2. Include comprehensive tests
3. Document all features and APIs
4. Ensure cross-platform compatibility
5. Handle errors gracefully
6. Respect user privacy and security

### Plugin Review Process
1. **Code Review**: Security and quality check
2. **Testing**: Automated and manual testing
3. **Documentation**: Complete documentation review
4. **Performance**: Performance impact assessment
5. **Approval**: Official plugin approval and listing

## 📝 Resources

- **Plugin SDK Reference**: Complete API documentation
- **Example Plugins**: Real-world plugin examples
- **Development Tools**: Plugin development utilities
- **Community Forum**: Plugin developer discussions
- **Best Practices Guide**: Security and performance guidelines

---

The CloudBox Plugin System transforms CloudBox from a hosting platform into a **complete development ecosystem** where teams can build exactly what they need. 🚀