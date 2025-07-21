import { EventEmitter } from 'eventemitter3';
import { ApiClient } from './utils/ApiClient';
import { Auth } from './services/Auth';
import { Database } from './services/Database';
import { Storage } from './services/Storage';
import { Functions } from './services/Functions';
import { Messaging } from './services/Messaging';
import { Users } from './services/Users';
import type { CloudBoxConfig } from './types';

export class CloudBox extends EventEmitter {
  public readonly config: CloudBoxConfig;
  public readonly apiClient: ApiClient;
  
  // Services
  public readonly auth: Auth;
  public readonly database: Database;
  public readonly storage: Storage;
  public readonly functions: Functions;
  public readonly messaging: Messaging;
  public readonly users: Users;

  constructor(config: CloudBoxConfig) {
    super();
    
    this.config = {
      endpoint: 'https://api.cloudbox.dev',
      timeout: 30000,
      ...config
    };

    // Validate required config
    if (!this.config.apiKey) {
      throw new Error('API key is required');
    }

    if (!this.config.projectId && !this.config.projectSlug) {
      throw new Error('Either projectId or projectSlug is required');
    }

    // Initialize API client
    this.apiClient = new ApiClient(this.config);

    // Initialize services
    this.auth = new Auth(this);
    this.database = new Database(this);
    this.storage = new Storage(this);
    this.functions = new Functions(this);
    this.messaging = new Messaging(this);
    this.users = new Users(this);

    // Set up auth event handling
    this.auth.on('login', (user) => {
      this.emit('auth:login', user);
    });

    this.auth.on('logout', () => {
      this.emit('auth:logout');
    });

    this.auth.on('session_expired', () => {
      this.emit('auth:session_expired');
    });
  }

  /**
   * Get the project API base path
   */
  getProjectApiPath(): string {
    const identifier = this.config.projectSlug || this.config.projectId;
    return `/p/${identifier}/api`;
  }

  /**
   * Get the admin API base path for project management
   */
  getAdminApiPath(): string {
    return `/api/v1/projects/${this.config.projectId}`;
  }

  /**
   * Set authentication token for admin operations
   */
  setAuthToken(token: string): void {
    this.apiClient.setAuthToken(token);
  }

  /**
   * Clear authentication token
   */
  clearAuthToken(): void {
    this.apiClient.clearAuthToken();
  }

  /**
   * Check connection to CloudBox API
   */
  async ping(): Promise<boolean> {
    try {
      await this.apiClient.get('/health');
      return true;
    } catch {
      return false;
    }
  }

  /**
   * Get project information
   */
  async getProject(): Promise<any> {
    return this.apiClient.get(this.getAdminApiPath());
  }
}

// Static factory method for easier initialization
export namespace CloudBox {
  export function init(config: CloudBoxConfig): CloudBox {
    return new CloudBox(config);
  }
}