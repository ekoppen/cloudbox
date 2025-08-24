/**
 * CloudBox SDK Client - TypeScript Implementation
 * Official CloudBox SDK for JavaScript/TypeScript applications
 * 
 * @version 1.0.0
 * @author VibCode
 */

import {
  CloudBoxConfig,
  RequestOptions,
  ApiResponse,
  ApiError,
  ConnectionTestResult
} from './types';

import { CollectionManager } from './managers/collections';
import { StorageManager } from './managers/storage';
import { UserManager } from './managers/users';
import { FunctionManager } from './managers/functions';
import { AuthManager } from './managers/auth';

/**
 * Main CloudBox SDK Client
 * 
 * Provides a unified interface to CloudBox BaaS platform with:
 * - Type-safe API interactions
 * - Automatic error handling
 * - Built-in retry logic
 * - Connection testing
 * - Comprehensive logging
 * 
 * @example
 * ```typescript
 * import { CloudBoxClient } from '@ekoppen/cloudbox-sdk';
 * 
 * // Project authentication (default - for most applications)
 * const client = new CloudBoxClient({
 *   projectId: 2,
 *   apiKey: 'your-api-key',
 *   endpoint: 'https://api.cloudbox.dev'
 * });
 * 
 * // Admin authentication (for CloudBox admin interfaces)
 * const adminClient = new CloudBoxClient({
 *   projectId: 2,
 *   apiKey: 'your-admin-api-key',
 *   endpoint: 'https://api.cloudbox.dev',
 *   authMode: 'admin'
 * });
 * 
 * // Collections
 * const users = await client.collections.list();
 * 
 * // Storage
 * const files = await client.storage.listFiles('images');
 * 
 * // Users
 * const newUser = await client.users.create({
 *   email: 'user@example.com',
 *   password: 'secure123'
 * });
 * ```
 */
export class CloudBoxClient {
  private readonly config: Required<CloudBoxConfig>;
  private readonly baseUrl: string;
  private authToken?: string;
  
  // Service managers
  public readonly collections: CollectionManager;
  public readonly storage: StorageManager;
  public readonly users: UserManager;
  public readonly functions: FunctionManager;
  public readonly auth: AuthManager;

  /**
   * Create a new CloudBox client instance
   * 
   * @param config - Configuration object with project details
   * @throws {Error} When required configuration is missing
   */
  constructor(config: CloudBoxConfig) {
    // Validate required configuration
    if (!config.projectId) {
      throw new Error('CloudBox SDK: projectId is required');
    }
    if (!config.apiKey) {
      throw new Error('CloudBox SDK: apiKey is required');
    }

    // Set up configuration with defaults
    this.config = {
      projectId: config.projectId,
      apiKey: config.apiKey,
      endpoint: config.endpoint || 'http://localhost:8080',
      authMode: config.authMode || 'project'
    };

    // Build base URL following CloudBox API standards
    this.baseUrl = `${this.config.endpoint}/p/${this.config.projectId}/api`;

    // Initialize service managers
    this.collections = new CollectionManager(this);
    this.storage = new StorageManager(this);
    this.users = new UserManager(this);
    this.functions = new FunctionManager(this);
    this.auth = new AuthManager(this);
  }

  /**
   * Make an authenticated HTTP request to CloudBox API
   * 
   * @param path - API endpoint path (e.g., '/collections')
   * @param options - Request options (method, headers, body, etc.)
   * @returns Promise resolving to API response data
   * @throws {ApiError} When API request fails
   * 
   * @example
   * ```typescript
   * const response = await client.request('/collections', {
   *   method: 'POST',
   *   body: { name: 'users', schema: [] }
   * });
   * ```
   */
  async request<T = any>(path: string, options: RequestOptions = {}): Promise<T> {
    // Determine if this is an admin endpoint based on path or authMode
    const isAdminEndpoint = path.startsWith('/api/v1/') || 
      (this.config.authMode === 'admin' && (path.startsWith('/users/') || path.includes('/auth')));
    
    const url = isAdminEndpoint 
      ? `${this.config.endpoint}${path}` 
      : `${this.baseUrl}${path}`;
    
    // Prepare request headers with authentication
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      'User-Agent': 'CloudBoxSDK/1.0.0',
      ...options.headers
    };

    // Add appropriate authentication header
    if (isAdminEndpoint) {
      // Admin endpoints use Bearer token if available, otherwise API key
      const authToken = options.headers?.['Authorization'] || this.authToken;
      if (authToken) {
        headers['Authorization'] = authToken.startsWith('Bearer ') ? authToken : `Bearer ${authToken}`;
      } else {
        headers['X-API-Key'] = this.config.apiKey;
      }
    } else {
      // Project endpoints use API key
      headers['X-API-Key'] = this.config.apiKey;
    }

    // Prepare request body
    let body: string | FormData | undefined;
    if (options.body) {
      if (options.body instanceof FormData) {
        body = options.body;
        // Remove Content-Type header for FormData (let browser set it)
        delete headers['Content-Type'];
      } else {
        body = JSON.stringify(options.body);
      }
    }

    // Build URL with query parameters
    const requestUrl = new URL(url);
    if (options.params) {
      Object.entries(options.params).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          requestUrl.searchParams.set(key, String(value));
        }
      });
    }

    try {
      // Make the HTTP request
      const response = await fetch(requestUrl.toString(), {
        method: options.method || 'GET',
        headers,
        body
      });

      // Handle response
      return await this.handleResponse<T>(response);
    } catch (error) {
      // Wrap network errors
      if (error instanceof Error && error.name === 'TypeError') {
        const apiError = new Error(
          `CloudBox SDK: Network error - ${error.message}`
        ) as ApiError;
        apiError.status = 0;
        apiError.response = null;
        throw apiError;
      }
      throw error;
    }
  }

  /**
   * Handle HTTP response and extract data
   * 
   * @private
   * @param response - Fetch Response object
   * @returns Parsed response data
   * @throws {ApiError} When response indicates an error
   */
  private async handleResponse<T>(response: Response): Promise<T> {
    const contentType = response.headers.get('content-type');
    let data: any;

    // Parse response based on content type
    if (contentType?.includes('application/json')) {
      try {
        data = await response.json();
      } catch {
        data = { error: 'Invalid JSON response from server' };
      }
    } else {
      const text = await response.text();
      data = text || { error: 'Empty response from server' };
    }

    // Handle error responses
    if (!response.ok) {
      const error = new Error(
        data?.error || 
        data?.message || 
        `CloudBox API error: ${response.status} ${response.statusText}`
      ) as ApiError;
      
      error.status = response.status;
      error.response = data;
      error.code = data?.code;
      
      throw error;
    }

    // Return data (unwrap if wrapped in 'data' property)
    return data?.data !== undefined ? data.data : data;
  }

  /**
   * Test connection to CloudBox API
   * 
   * Verifies that:
   * - The endpoint is reachable
   * - The API key is valid  
   * - The project exists and is accessible
   * 
   * @returns Promise resolving to connection test result
   * 
   * @example
   * ```typescript
   * const result = await client.testConnection();
   * if (result.success) {
   *   console.log('✅ Connected to CloudBox!');
   * } else {
   *   console.error('❌ Connection failed:', result.message);
   * }
   * ```
   */
  async testConnection(): Promise<ConnectionTestResult> {
    const startTime = Date.now();
    
    try {
      // Try to list collections as a simple test
      await this.request('/collections');
      
      const responseTime = Date.now() - startTime;
      
      return {
        success: true,
        message: 'Connected to CloudBox successfully',
        endpoint: this.config.endpoint,
        project_id: this.config.projectId,
        response_time: responseTime,
        api_version: '1.0'
      };
    } catch (error) {
      const responseTime = Date.now() - startTime;
      
      return {
        success: false,
        message: error instanceof Error ? error.message : 'Unknown connection error',
        endpoint: this.config.endpoint,
        project_id: this.config.projectId,
        response_time: responseTime
      };
    }
  }

  /**
   * Get current client configuration
   * 
   * @returns Client configuration (without sensitive data)
   */
  getConfig(): Omit<CloudBoxConfig, 'apiKey'> {
    return {
      projectId: this.config.projectId,
      endpoint: this.config.endpoint,
      authMode: this.config.authMode
    };
  }

  /**
   * Get base URL for API requests
   * 
   * @returns Base URL used for API calls
   */
  getBaseUrl(): string {
    return this.baseUrl;
  }

  /**
   * Create a new CloudBox client with different configuration
   * 
   * @param config - New configuration options
   * @returns New CloudBox client instance
   * 
   * @example
   * ```typescript
   * // Create client for different project
   * const newClient = client.withConfig({
   *   projectId: 'different-project',
   *   apiKey: 'different-key'
   * });
   * ```
   */
  withConfig(config: Partial<CloudBoxConfig>): CloudBoxClient {
    return new CloudBoxClient({
      ...this.config,
      ...config
    });
  }

  /**
   * Set authentication token for admin/auth requests
   * 
   * @param token - JWT token from login response
   * 
   * @example
   * ```typescript
   * const { token } = await client.auth.login({ 
   *   email: 'user@example.com', 
   *   password: 'password' 
   * });
   * 
   * client.setAuthToken(token);
   * ```
   */
  setAuthToken(token: string): void {
    this.authToken = token;
  }

  /**
   * Clear authentication token
   * 
   * @example
   * ```typescript
   * await client.auth.logout();
   * client.clearAuthToken();
   * ```
   */
  clearAuthToken(): void {
    this.authToken = undefined;
  }

  /**
   * Get current authentication token
   * 
   * @returns Current auth token or undefined
   */
  getAuthToken(): string | undefined {
    return this.authToken;
  }

  /**
   * Get authentication mode
   * 
   * @returns Current authentication mode
   */
  getAuthMode(): 'admin' | 'project' {
    return this.config.authMode;
  }
}