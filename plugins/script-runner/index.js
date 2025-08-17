// CloudBox Universal Script Runner Plugin
// Native integration in CloudBox dashboard - Universeel voor alle projecten

const { CloudBoxPlugin, DatabaseManager, FunctionManager, WebhookManager } = require('cloudbox-sdk');

class UniversalScriptRunnerPlugin extends CloudBoxPlugin {
  constructor() {
    super('cloudbox-script-runner');
    this.scriptTemplates = new Map();
    this.executionHistory = new Map();
    this.loadBuiltInTemplates();
  }

  // Plugin lifecycle
  async onInstall() {
    console.log('📦 Installing CloudBox Universal Script Runner Plugin...');
    
    // Create plugin tables
    await this.createPluginTables();
    
    // Register API routes
    this.registerRoutes();
    
    // Register dashboard components
    this.registerDashboardComponents();
    
    console.log('✅ Universal Script Runner Plugin installed successfully');
  }

  async onUninstall() {
    console.log('🗑️ Uninstalling Universal Script Runner Plugin...');
    // Cleanup logic here
  }

  async createPluginTables() {
    const db = this.getDatabase();
    
    // Create schema
    await db.query('CREATE SCHEMA IF NOT EXISTS script_runner');
    
    // Create scripts table
    await db.query(`
      CREATE TABLE IF NOT EXISTS script_runner.scripts (
        id SERIAL PRIMARY KEY,
        name VARCHAR(100) NOT NULL,
        description TEXT,
        type VARCHAR(20) NOT NULL,
        category VARCHAR(30) DEFAULT 'custom',
        project_id VARCHAR(50),
        content TEXT NOT NULL,
        version VARCHAR(20) DEFAULT '1.0.0',
        dependencies JSONB DEFAULT '[]',
        run_order INTEGER DEFAULT 999,
        author VARCHAR(50),
        tags TEXT[],
        is_template BOOLEAN DEFAULT false,
        is_public BOOLEAN DEFAULT false,
        last_run TIMESTAMP,
        run_count INTEGER DEFAULT 0,
        success_count INTEGER DEFAULT 0,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        UNIQUE(name, project_id)
      )
    `);
    
    // Create executions table
    await db.query(`
      CREATE TABLE IF NOT EXISTS script_runner.executions (
        id SERIAL PRIMARY KEY,
        script_id INTEGER REFERENCES script_runner.scripts(id) ON DELETE CASCADE,
        project_id VARCHAR(50) NOT NULL,
        status VARCHAR(20) DEFAULT 'running',
        output TEXT,
        error_message TEXT,
        duration_ms INTEGER,
        executed_by VARCHAR(50),
        execution_context JSONB DEFAULT '{}',
        started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        completed_at TIMESTAMP
      )
    `);
    
    // Create indexes
    await db.query('CREATE INDEX IF NOT EXISTS idx_executions_script_started ON script_runner.executions(script_id, started_at)');
    await db.query('CREATE INDEX IF NOT EXISTS idx_executions_project_started ON script_runner.executions(project_id, started_at)');
    await db.query('CREATE INDEX IF NOT EXISTS idx_executions_status ON script_runner.executions(status)');
    
    // Create script collections table
    await db.query(`
      CREATE TABLE IF NOT EXISTS script_runner.script_collections (
        id SERIAL PRIMARY KEY,
        name VARCHAR(100) NOT NULL,
        description TEXT,
        category VARCHAR(30) DEFAULT 'project-setup',
        project_id VARCHAR(50),
        scripts JSONB NOT NULL,
        default_variables JSONB DEFAULT '{}',
        usage_count INTEGER DEFAULT 0,
        last_used TIMESTAMP,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        UNIQUE(name, project_id)
      )
    `);
    
    console.log('✅ Script Runner database schema created');
  }

  registerRoutes() {
    // Script management routes
    this.app.get('/api/plugins/script-runner/scripts/:projectId', this.getProjectScripts.bind(this));
    this.app.post('/api/plugins/script-runner/scripts/:projectId', this.createScript.bind(this));
    this.app.put('/api/plugins/script-runner/scripts/:projectId/:scriptId', this.updateScript.bind(this));
    this.app.delete('/api/plugins/script-runner/scripts/:projectId/:scriptId', this.deleteScript.bind(this));
    
    // Execution routes
    this.app.post('/api/plugins/script-runner/execute/:projectId/:scriptId', this.executeScript.bind(this));
    this.app.post('/api/plugins/script-runner/execute-collection/:projectId/:collectionName', this.executeCollection.bind(this));
    this.app.get('/api/plugins/script-runner/executions/:projectId', this.getExecutionHistory.bind(this));
    
    // Template and collection routes
    this.app.get('/api/plugins/script-runner/templates', this.getScriptTemplates.bind(this));
    this.app.get('/api/plugins/script-runner/collections/:projectId', this.getScriptCollections.bind(this));
    this.app.post('/api/plugins/script-runner/collections/:projectId', this.createScriptCollection.bind(this));
    
    // Project setup routes
    this.app.get('/api/plugins/script-runner/project-templates', this.getProjectTemplates.bind(this));
    this.app.post('/api/plugins/script-runner/setup-project/:projectId/:templateName', this.setupProjectFromTemplate.bind(this));
  }

  registerDashboardComponents() {
    // Register UI components in CloudBox dashboard
    this.registerComponent('script-runner-dashboard', {
      component: './components/ScriptRunnerDashboard.svelte',
      route: '/script-runner',
      menu: {
        title: 'Script Runner',
        icon: 'terminal',
        order: 100
      }
    });

    this.registerComponent('project-scripts', {
      component: './components/ProjectScripts.svelte', 
      route: '/projects/:projectId/scripts',
      contextMenu: {
        title: 'Database Scripts',
        icon: 'database'
      }
    });
  }

  async getProjectScripts(req, res) {
    const { projectId } = req.params;
    
    try {
      const db = this.getDatabase();
      const scripts = await db.query(`
        SELECT s.*, 
               COALESCE(e.last_execution, NULL) as last_execution,
               COALESCE(e.last_status, 'never_run') as last_status
        FROM script_runner.scripts s
        LEFT JOIN (
          SELECT script_id, 
                 MAX(started_at) as last_execution,
                 (array_agg(status ORDER BY started_at DESC))[1] as last_status
          FROM script_runner.executions 
          GROUP BY script_id
        ) e ON s.id = e.script_id
        WHERE s.project_id = $1 OR s.is_template = true
        ORDER BY s.run_order, s.name
      `, [projectId]);
      
      res.json({ success: true, scripts: scripts.rows });
    } catch (error) {
      res.status(500).json({ success: false, error: error.message });
    }
  }

  async createScript(req, res) {
    const { projectId } = req.params;
    const { name, description, type, content, dependencies = [], runOrder = 999 } = req.body;
    
    try {
      const db = this.getDatabase();
      const result = await db.query(`
        INSERT INTO script_runner.scripts 
        (name, description, type, project_id, content, dependencies, run_order, author)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id
      `, [name, description, type, projectId, content, JSON.stringify(dependencies), runOrder, req.user?.id || 'system']);
      
      res.json({ success: true, script_id: result.rows[0].id });
    } catch (error) {
      res.status(500).json({ success: false, error: error.message });
    }
  }

  async executeScript(req, res) {
    const { projectId, scriptId } = req.params;
    const { variables = {} } = req.body;
    
    try {
      const db = this.getDatabase();
      
      // Get script
      const scriptResult = await db.query(
        'SELECT * FROM script_runner.scripts WHERE id = $1',
        [scriptId]
      );
      
      if (scriptResult.rows.length === 0) {
        return res.status(404).json({ success: false, error: 'Script not found' });
      }
      
      const script = scriptResult.rows[0];
      
      // Create execution record
      const executionResult = await db.query(`
        INSERT INTO script_runner.executions (script_id, project_id, executed_by, execution_context)
        VALUES ($1, $2, $3, $4)
        RETURNING id
      `, [scriptId, projectId, req.user?.id || 'system', JSON.stringify({ variables })]); 
      
      const executionId = executionResult.rows[0].id;
      
      // Execute script
      const result = await this.runScript(script, projectId, variables);
      
      // Update execution record
      await db.query(`
        UPDATE script_runner.executions 
        SET status = $1, output = $2, error_message = $3, duration_ms = $4, completed_at = CURRENT_TIMESTAMP
        WHERE id = $5
      `, [
        result.success ? 'success' : 'failed',
        result.output,
        result.error,
        result.duration,
        executionId
      ]);
      
      // Update script stats
      await db.query(`
        UPDATE script_runner.scripts 
        SET last_run = CURRENT_TIMESTAMP, 
            run_count = run_count + 1,
            success_count = success_count + CASE WHEN $2 THEN 1 ELSE 0 END
        WHERE id = $1
      `, [scriptId, result.success]);
      
      res.json({ 
        success: true, 
        execution_id: executionId,
        result 
      });
      
    } catch (error) {
      console.error('Script execution error:', error);
      res.status(500).json({ success: false, error: error.message });
    }
  }

  async runScript(script, projectId, variables = {}) {
    const startTime = Date.now();
    
    try {
      let result;
      
      // Replace variables in script content
      let content = script.content;
      for (const [key, value] of Object.entries(variables)) {
        content = content.replace(new RegExp(`{{${key}}}`, 'g'), value);
      }
      
      switch (script.type) {
        case 'sql':
        case 'migration':
          result = await this.executeSQLScript(content, projectId);
          break;
        case 'javascript':
          result = await this.executeJavaScriptScript(content, projectId);
          break;
        case 'setup':
          result = await this.executeSetupScript(content, projectId);
          break;
        default:
          throw new Error(`Unsupported script type: ${script.type}`);
      }
      
      return {
        success: true,
        output: result,
        duration: Date.now() - startTime
      };
      
    } catch (error) {
      return {
        success: false,
        output: '',
        error: error.message,
        duration: Date.now() - startTime
      };
    }
  }

  async executeSQLScript(content, projectId) {
    const db = this.getProjectDatabase(projectId);
    
    // Split into individual statements, handle PostgreSQL syntax
    const statements = content
      .split(';')
      .map(s => s.trim())
      .filter(s => s.length > 0 && !s.startsWith('--') && s !== '');
    
    const results = [];
    
    for (const statement of statements) {
      try {
        const result = await db.query(statement);
        results.push({
          statement: statement.substring(0, 100) + (statement.length > 100 ? '...' : ''),
          rows_affected: result.rowCount || 0,
          rows_returned: result.rows ? result.rows.length : 0,
          success: true
        });
      } catch (error) {
        results.push({
          statement: statement.substring(0, 100) + (statement.length > 100 ? '...' : ''),
          error: error.message,
          success: false
        });
        throw error; // Stop on first error
      }
    }
    
    return `Executed ${results.length} PostgreSQL statements successfully. ${results.map(r => `${r.statement}: ${r.rows_affected} rows affected`).join(', ')}`;
  }

  async executeJavaScriptScript(content, projectId) {
    // Create sandbox environment for JavaScript execution
    const sandbox = {
      console: {
        log: (...args) => this.logOutput(args.join(' ')),
        error: (...args) => this.logError(args.join(' '))
      },
      CloudBox: {
        functions: new FunctionManager(projectId),
        database: this.getProjectDatabase(projectId),
        webhooks: new WebhookManager(projectId)
      },
      require: (module) => {
        // Whitelist allowed modules
        const allowedModules = ['crypto', 'url', 'querystring', 'path'];
        if (allowedModules.includes(module)) {
          return require(module);
        }
        throw new Error(`Module '${module}' is not allowed in scripts`);
      }
    };
    
    // Execute in VM context for security
    const vm = require('vm');
    const context = vm.createContext(sandbox);
    
    try {
      const result = vm.runInContext(content, context, {
        timeout: 30000, // 30 seconds max
        filename: 'cloudbox-script.js'
      });
      
      return typeof result === 'object' ? JSON.stringify(result, null, 2) : String(result);
    } catch (error) {
      throw new Error(`JavaScript execution failed: ${error.message}`);
    }
  }

  async executeSetupScript(content, projectId) {
    // Parse setup commands from content
    const lines = content.split('\n').filter(line => line.trim() && !line.startsWith('#'));
    const results = [];
    
    for (const line of lines) {
      const [command, ...args] = line.trim().split(' ');
      
      try {
        switch (command) {
          case 'CREATE_FUNCTION':
            await this.createFunction(projectId, args);
            results.push(`Function ${args[0]} created`);
            break;
          case 'CREATE_WEBHOOK':
            await this.createWebhook(projectId, args);
            results.push(`Webhook ${args[0]} created`);
            break;
          case 'SCHEDULE_JOB':
            await this.scheduleJob(projectId, args);
            results.push(`Job ${args[0]} scheduled`);
            break;
          case 'CREATE_DATABASE':
            await this.createDatabase(projectId, args);
            results.push(`Database ${args[0]} created`);
            break;
          default:
            results.push(`Unknown command: ${command}`);
        }
      } catch (error) {
        results.push(`Error executing ${command}: ${error.message}`);
        throw error;
      }
    }
    
    return results.join('\n');
  }

  loadBuiltInTemplates() {
    // Universal project templates for any framework
    const universalTemplates = [
      {
        name: 'Basic Web App Setup',
        description: 'Basic database schema for web applications',
        category: 'webapp',
        scripts: [
          'create-users-table',
          'create-sessions-table', 
          'create-api-keys-table'
        ]
      },
      {
        name: 'AI Chat Application',
        description: 'Complete setup for AI chat applications (like Aimy)',
        category: 'ai-app',
        scripts: [
          'ai-chat-database-schema', 
          'ai-chat-functions-setup',
          'ai-chat-webhooks-config'
        ]
      },
      {
        name: 'E-commerce Backend',
        description: 'Database schema for e-commerce platforms',
        category: 'ecommerce',
        scripts: [
          'ecommerce-database-schema',
          'payment-integration-setup',
          'inventory-management'
        ]
      }
    ];
    
    this.scriptTemplates.set('universal', universalTemplates);
  }

  async getProjectTemplates(req, res) {
    const templates = Array.from(this.scriptTemplates.values()).flat();
    res.json({ success: true, templates });
  }

  // Utility methods
  logOutput(message) {
    console.log(`[Script Output] ${message}`);
  }

  logError(message) {
    console.error(`[Script Error] ${message}`);
  }

  getProjectDatabase(projectId) {
    return new DatabaseManager(projectId);
  }

  async createFunction(projectId, args) {
    const [functionName, ...options] = args;
    const functionManager = new FunctionManager(projectId);
    return await functionManager.deploy(functionName, options);
  }

  async createWebhook(projectId, args) {
    const [webhookName, ...options] = args;
    const webhookManager = new WebhookManager(projectId);
    return await webhookManager.create(webhookName, options);
  }

  async scheduleJob(projectId, args) {
    // Implement job scheduling logic
    const [jobName, schedule, ...options] = args;
    console.log(`Scheduling job ${jobName} with schedule ${schedule}`);
  }

  async createDatabase(projectId, args) {
    const [databaseName, ...options] = args;
    const db = this.getProjectDatabase(projectId);
    return await db.query(`CREATE DATABASE IF NOT EXISTS ${databaseName}`);
  }
}

// Export plugin for CloudBox
module.exports = UniversalScriptRunnerPlugin;