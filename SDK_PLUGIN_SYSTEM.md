# CloudBox Plugin System SDK

Official SDK documentation for developing CloudBox plugins.

## 🎯 Quick Start

### Installation
```bash
npm install cloudbox-sdk
```

### Basic Plugin Structure
```javascript
const { CloudBoxPlugin, DatabaseManager, FunctionManager } = require('cloudbox-sdk');

class MyPlugin extends CloudBoxPlugin {
  constructor() {
    super('my-plugin-name');
  }

  async onInstall() {
    console.log('Installing MyPlugin...');
    await this.createPluginTables();
    this.registerRoutes();
    this.registerDashboardComponents();
  }

  async onUninstall() {
    console.log('Uninstalling MyPlugin...');
    // Cleanup logic
  }
}

module.exports = MyPlugin;
```

## 📚 SDK Reference

### CloudBoxPlugin Base Class

#### Constructor
```javascript
constructor(pluginName: string)
```

#### Lifecycle Methods
```javascript
async onInstall(): Promise<void>
async onUninstall(): Promise<void> 
async onEnable(): Promise<void>
async onDisable(): Promise<void>
```

#### Database Methods
```javascript
getDatabase(): DatabaseManager
getProjectDatabase(projectId: string): DatabaseManager

async createPluginTables(): Promise<void>
async dropPluginTables(): Promise<void>
```

#### Registration Methods
```javascript
registerRoutes(): void
registerDashboardComponents(): void
registerComponent(name: string, config: ComponentConfig): void
```

### DatabaseManager

#### Connection
```javascript
// Get database connection
const db = this.getDatabase();
const projectDb = this.getProjectDatabase(projectId);
```

#### Query Methods
```javascript
async query(sql: string, params?: any[]): Promise<QueryResult>
async queryOne(sql: string, params?: any[]): Promise<any>
async queryMany(sql: string, params?: any[]): Promise<any[]>

// Transaction support
async transaction(callback: (db: DatabaseManager) => Promise<void>): Promise<void>
```

#### Schema Operations
```javascript
async createSchema(schemaName: string): Promise<void>
async dropSchema(schemaName: string): Promise<void>
async tableExists(tableName: string): Promise<boolean>
```

#### Migration Support
```javascript
async runMigration(migrationSql: string): Promise<void>
async rollbackMigration(rollbackSql: string): Promise<void>
```

### FunctionManager

#### Deployment
```javascript
const functions = new FunctionManager(projectId);

await functions.deploy(functionName: string, config: FunctionConfig);
await functions.update(functionName: string, config: FunctionConfig);
await functions.delete(functionName: string);
```

#### Function Configuration
```javascript
interface FunctionConfig {
  runtime: 'nodejs18' | 'nodejs16' | 'python3.9' | 'go1.19';
  port: number;
  memory: '128MB' | '256MB' | '512MB' | '1GB' | '2GB';
  timeout: number; // seconds
  environment: Record<string, string>;
  code?: string; // Inline code
  file?: string; // Code from file
}
```

#### Management
```javascript
await functions.list(): Promise<Function[]>
await functions.get(functionName: string): Promise<Function>
await functions.logs(functionName: string, lines?: number): Promise<string[]>
await functions.restart(functionName: string): Promise<void>
```

### WebhookManager

#### Creation
```javascript
const webhooks = new WebhookManager(projectId);

await webhooks.create(webhookName: string, config: WebhookConfig);
await webhooks.update(webhookName: string, config: WebhookConfig);
await webhooks.delete(webhookName: string);
```

#### Webhook Configuration
```javascript
interface WebhookConfig {
  path: string; // e.g., '/api/webhook/my-webhook'
  methods: string[]; // ['GET', 'POST', 'PUT', 'DELETE']
  function: string; // Target function name
  public: boolean; // Public access without auth
  rateLimit?: number; // Requests per minute
  cors?: boolean; // Enable CORS
}
```

### StorageManager

#### File Operations
```javascript
const storage = new StorageManager(projectId);

await storage.upload(file: Buffer | string, bucket: string, filename: string): Promise<string>
await storage.download(bucket: string, filename: string): Promise<Buffer>
await storage.delete(bucket: string, filename: string): Promise<void>
await storage.list(bucket: string, prefix?: string): Promise<StorageFile[]>
```

#### Bucket Management
```javascript
await storage.createBucket(bucketName: string, config?: BucketConfig): Promise<void>
await storage.deleteBucket(bucketName: string): Promise<void>
await storage.listBuckets(): Promise<Bucket[]>
```

### UserManager

#### User Operations
```javascript
const users = new UserManager();

await users.get(userId: string): Promise<User>
await users.list(filters?: UserFilters): Promise<User[]>
await users.create(userData: CreateUserData): Promise<User>
await users.update(userId: string, userData: UpdateUserData): Promise<User>
await users.delete(userId: string): Promise<void>
```

#### Authentication
```javascript
await users.authenticate(email: string, password: string): Promise<AuthResult>
await users.createSession(userId: string): Promise<Session>
await users.validateSession(sessionId: string): Promise<Session>
```

## 🎨 Frontend Integration

### Component Registration
```javascript
registerDashboardComponents() {
  // Global dashboard tab
  this.registerComponent('my-plugin-dashboard', {
    component: './components/Dashboard.svelte',
    route: '/my-plugin',
    menu: {
      title: 'My Plugin',
      icon: 'plugin',
      order: 100
    }
  });

  // Project-specific menu item
  this.registerComponent('my-plugin-project', {
    component: './components/ProjectView.svelte',
    route: '/projects/:projectId/my-plugin',
    contextMenu: {
      title: 'My Plugin',
      icon: 'plugin'
    }
  });
}
```

### Component Configuration
```typescript
interface ComponentConfig {
  component: string; // Path to Svelte/React component
  route: string; // Route pattern
  menu?: {
    title: string;
    icon: string;
    order?: number;
  };
  contextMenu?: {
    title: string;
    icon: string;
  };
  permissions?: string[]; // Required permissions
}
```

### Svelte Component Template
```svelte
<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import Button from '$lib/components/ui/button.svelte';
  import Card from '$lib/components/ui/card.svelte';
  import { addToast } from '$lib/stores/toast';

  // Get project ID from URL
  $: projectId = $page.params.id;

  let data = [];
  let loading = false;

  async function loadData() {
    loading = true;
    try {
      const response = await fetch(`/api/plugins/my-plugin/data/${projectId}`, {
        credentials: 'include'
      });
      
      if (response.ok) {
        const result = await response.json();
        if (result.success) {
          data = result.data;
        }
      } else {
        throw new Error('Failed to load data');
      }
    } catch (error) {
      console.error('Load error:', error);
      addToast('Failed to load data', 'error');
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    loadData();
  });
</script>

<svelte:head>
  <title>My Plugin - CloudBox</title>
</svelte:head>

<div class="space-y-6">
  <div class="flex items-center justify-between">
    <h1 class="text-2xl font-bold text-foreground">My Plugin</h1>
    <Button on:click={() => {}}>
      Action Button
    </Button>
  </div>

  <Card>
    <div class="p-6">
      {#if loading}
        <div class="text-center">
          <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto"></div>
          <p class="mt-2 text-muted-foreground">Loading...</p>
        </div>
      {:else}
        <!-- Plugin content here -->
        <p>Plugin content for project {projectId}</p>
      {/if}
    </div>
  </Card>
</div>
```

## 🔌 API Integration

### Route Registration
```javascript
registerRoutes() {
  // GET endpoints
  this.app.get('/api/plugins/my-plugin/data/:projectId', this.getData.bind(this));
  this.app.get('/api/plugins/my-plugin/status', this.getStatus.bind(this));

  // POST endpoints  
  this.app.post('/api/plugins/my-plugin/data/:projectId', this.createData.bind(this));
  this.app.post('/api/plugins/my-plugin/action/:projectId', this.performAction.bind(this));

  // PUT/DELETE endpoints
  this.app.put('/api/plugins/my-plugin/data/:projectId/:id', this.updateData.bind(this));
  this.app.delete('/api/plugins/my-plugin/data/:projectId/:id', this.deleteData.bind(this));
}
```

### Route Handler Example
```javascript
async getData(req, res) {
  const { projectId } = req.params;
  const { limit = 10, offset = 0 } = req.query;

  try {
    // Check permissions
    if (!this.hasPermission(req.user, 'database:read')) {
      return res.status(403).json({ success: false, error: 'Insufficient permissions' });
    }

    // Get project database
    const db = this.getProjectDatabase(projectId);
    
    // Query data
    const data = await db.queryMany(
      'SELECT * FROM my_plugin.data WHERE project_id = $1 LIMIT $2 OFFSET $3',
      [projectId, limit, offset]
    );

    res.json({ success: true, data });
  } catch (error) {
    console.error('Get data error:', error);
    res.status(500).json({ success: false, error: error.message });
  }
}
```

### Authentication & Permissions
```javascript
// Check user permissions
hasPermission(user, permission) {
  return user.permissions.includes(permission) || user.role === 'admin';
}

// Middleware for authentication
requireAuth() {
  return (req, res, next) => {
    if (!req.user) {
      return res.status(401).json({ success: false, error: 'Authentication required' });
    }
    next();
  };
}

// Example protected route
this.app.get('/api/plugins/my-plugin/admin', this.requireAuth(), this.adminOnly.bind(this));
```

## 📄 Plugin Templates

### Project Setup Plugin Template
```javascript
class ProjectSetupPlugin extends CloudBoxPlugin {
  constructor() {
    super('project-setup');
  }

  async setupProject(projectId, template) {
    const db = this.getProjectDatabase(projectId);
    const functions = new FunctionManager(projectId);
    
    // Create database schema
    await db.query(template.schema);
    
    // Deploy functions
    for (const func of template.functions) {
      await functions.deploy(func.name, func.config);
    }
    
    // Setup webhooks
    const webhooks = new WebhookManager(projectId);
    for (const webhook of template.webhooks) {
      await webhooks.create(webhook.name, webhook.config);
    }
  }
}
```

### Monitoring Plugin Template
```javascript
class MonitoringPlugin extends CloudBoxPlugin {
  constructor() {
    super('monitoring');
  }

  async collectMetrics(projectId) {
    const db = this.getProjectDatabase(projectId);
    
    // Collect database metrics
    const dbMetrics = await this.getDatabaseMetrics(db);
    
    // Collect function metrics
    const functions = new FunctionManager(projectId);
    const functionMetrics = await this.getFunctionMetrics(functions);
    
    // Store metrics
    await this.storeMetrics(projectId, { database: dbMetrics, functions: functionMetrics });
  }
}
```

## 🧪 Testing

### Plugin Testing Framework
```javascript
const { PluginTester } = require('cloudbox-sdk/testing');

describe('MyPlugin', () => {
  let tester;
  let plugin;

  beforeEach(async () => {
    tester = new PluginTester();
    plugin = new MyPlugin();
    await tester.setup(plugin);
  });

  afterEach(async () => {
    await tester.cleanup();
  });

  test('should install successfully', async () => {
    await plugin.onInstall();
    
    // Check if tables were created
    const tableExists = await tester.db.tableExists('my_plugin.data');
    expect(tableExists).toBe(true);
  });

  test('should handle API requests', async () => {
    const response = await tester.request
      .get('/api/plugins/my-plugin/data/test-project')
      .expect(200);
      
    expect(response.body.success).toBe(true);
  });
});
```

## 📋 Best Practices

### Security
- **Input Validation**: Always validate user inputs
- **Permission Checks**: Verify user permissions before operations
- **SQL Injection**: Use parameterized queries
- **Rate Limiting**: Implement rate limiting for API endpoints
- **Error Handling**: Don't expose sensitive information in errors

### Performance
- **Database Indexing**: Create appropriate indexes
- **Connection Pooling**: Reuse database connections
- **Caching**: Cache frequently accessed data
- **Async Operations**: Use async/await for I/O operations
- **Monitoring**: Track performance metrics

### Code Quality
- **Error Handling**: Comprehensive error handling
- **Logging**: Structured logging for debugging
- **Documentation**: Document all public APIs
- **Testing**: Unit and integration tests
- **Code Style**: Follow CloudBox coding standards

---

This SDK documentation provides everything needed to build powerful CloudBox plugins! 🚀