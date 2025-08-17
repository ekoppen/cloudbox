/**
 * CloudBox Script Runner Plugin
 * Universal script runner for database operations and project setup
 */

class ScriptRunnerPlugin {
  constructor(cloudbox) {
    this.cloudbox = cloudbox;
    this.name = 'cloudbox-script-runner';
    this.version = '1.0.0';
  }

  /**
   * Initialize the plugin
   */
  async initialize() {
    console.log(`[${this.name}] Plugin initialized v${this.version}`);
    
    // Register UI components
    this.registerUIComponents();
    
    // Register routes
    this.registerRoutes();
    
    // Register event listeners
    this.registerEventListeners();
  }

  /**
   * Register UI components in CloudBox dashboard
   */
  registerUIComponents() {
    // Project menu integration
    this.cloudbox.ui.registerProjectMenuItem({
      title: 'Scripts',
      icon: 'terminal',
      path: '/dashboard/projects/{projectId}/scripts',
      permissions: ['database:read']
    });

    // Dashboard tab integration  
    this.cloudbox.ui.registerDashboardTab({
      title: 'Script Runner',
      icon: 'terminal', 
      path: '/script-runner',
      permissions: ['database:read']
    });
  }

  /**
   * Register plugin routes
   */
  registerRoutes() {
    const routes = [
      {
        method: 'GET',
        path: '/api/v1/plugins/script-runner/scripts/:projectId',
        handler: this.getProjectScripts.bind(this)
      },
      {
        method: 'POST',
        path: '/api/v1/plugins/script-runner/scripts/:projectId', 
        handler: this.createScript.bind(this)
      },
      {
        method: 'PUT',
        path: '/api/v1/plugins/script-runner/scripts/:projectId/:scriptId',
        handler: this.updateScript.bind(this)
      },
      {
        method: 'DELETE',
        path: '/api/v1/plugins/script-runner/scripts/:projectId/:scriptId',
        handler: this.deleteScript.bind(this)
      },
      {
        method: 'POST',
        path: '/api/v1/plugins/script-runner/execute/:projectId/:scriptId',
        handler: this.executeScript.bind(this)
      },
      {
        method: 'POST',
        path: '/api/v1/plugins/script-runner/execute-raw/:projectId',
        handler: this.executeRawSQL.bind(this)
      },
      {
        method: 'GET',
        path: '/api/v1/plugins/script-runner/templates',
        handler: this.getTemplates.bind(this)
      },
      {
        method: 'POST',
        path: '/api/v1/plugins/script-runner/setup-project/:projectId/:templateName',
        handler: this.setupProjectTemplate.bind(this)
      }
    ];

    routes.forEach(route => {
      this.cloudbox.router.register(route.method, route.path, route.handler);
    });
  }

  /**
   * Register event listeners
   */
  registerEventListeners() {
    this.cloudbox.events.on('project:created', this.onProjectCreated.bind(this));
    this.cloudbox.events.on('project:deleted', this.onProjectDeleted.bind(this));
  }

  /**
   * Route handlers (delegated to Go backend)
   */
  async getProjectScripts(req, res) {
    return this.cloudbox.backend.proxy(req, res);
  }

  async createScript(req, res) {
    return this.cloudbox.backend.proxy(req, res);
  }

  async updateScript(req, res) {
    return this.cloudbox.backend.proxy(req, res);
  }

  async deleteScript(req, res) {
    return this.cloudbox.backend.proxy(req, res);
  }

  async executeScript(req, res) {
    return this.cloudbox.backend.proxy(req, res);
  }

  async executeRawSQL(req, res) {
    return this.cloudbox.backend.proxy(req, res);
  }

  async getTemplates(req, res) {
    return this.cloudbox.backend.proxy(req, res);
  }

  async setupProjectTemplate(req, res) {
    return this.cloudbox.backend.proxy(req, res);
  }

  /**
   * Event handlers
   */
  async onProjectCreated(project) {
    console.log(`[${this.name}] Project created:`, project.id);
    // Initialize script collections for new project
  }

  async onProjectDeleted(project) {
    console.log(`[${this.name}] Project deleted:`, project.id);
    // Cleanup script data for deleted project
  }

  /**
   * Plugin lifecycle methods
   */
  async enable() {
    console.log(`[${this.name}] Plugin enabled`);
    await this.initialize();
  }

  async disable() {
    console.log(`[${this.name}] Plugin disabled`);
    // Cleanup routes and UI components
  }

  async uninstall() {
    console.log(`[${this.name}] Plugin uninstalling`);
    // Cleanup all data and configurations
    await this.disable();
  }

  /**
   * Plugin metadata
   */
  getInfo() {
    return {
      name: this.name,
      version: this.version,
      description: 'Universal Script Runner for CloudBox - Database scripts en project setup',
      author: 'CloudBox Development Team',
      permissions: [
        'database:read',
        'database:write', 
        'functions:deploy',
        'webhooks:create',
        'projects:manage'
      ]
    };
  }
}

module.exports = ScriptRunnerPlugin;