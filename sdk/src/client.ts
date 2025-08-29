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
  ConnectionTestResult,
  AuthHeaderStrategy,
  CORSErrorInfo,
  AuthConfiguration
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
/**
 * Authentication header strategies for different client scenarios
 */
const DEFAULT_AUTH_STRATEGIES: Record<string, AuthHeaderStrategy> = {
  project: {
    primary: 'Session-Token',
    fallbacks: ['session-token', 'X-Session-Token', 'x-session-token', 'Authorization'],
    transform: (token: string) => token.startsWith('Bearer ') ? token.slice(7) : token
  },
  admin: {
    primary: 'Authorization',
    fallbacks: ['Bearer', 'X-Auth-Token'],
    transform: (token: string) => token.startsWith('Bearer ') ? token : `Bearer ${token}`
  }
};

/**
 * Authentication Header Manager - Handles header fallback strategies
 */
class AuthHeaderManager {
  constructor(
    private client: CloudBoxClient,
    private strategies: Record<string, AuthHeaderStrategy> = DEFAULT_AUTH_STRATEGIES
  ) {}

  /**
   * Make authenticated request with header fallback strategy
   */
  async makeAuthenticatedRequest<T>(
    url: string,
    options: RequestOptions = {},
    maxRetries: number = 3
  ): Promise<T> {
    const strategy = this.getStrategy();
    const triedHeaders: string[] = [];
    let lastError: any;

    // Skip fallback if explicitly disabled
    if (options.skipAuthFallback) {
      return this.client.directRequest<T>(url, options);
    }

    // Try each header in order
    for (const header of [strategy.primary, ...strategy.fallbacks]) {
      try {
        triedHeaders.push(header);
        const response = await this.tryWithHeader<T>(url, options, header, strategy);
        return response;
      } catch (error: any) {
        lastError = error;
        
        // If it's not a CORS error, don't try other headers
        if (!this.isCORSError(error)) {
          break;
        }
        
        // Continue to next header for CORS errors
        continue;
      }
    }

    // All headers failed - create enhanced error
    throw this.createEnhancedError(lastError, url, strategy, triedHeaders);
  }

  /**
   * Try request with specific header
   */
  private async tryWithHeader<T>(
    url: string,
    options: RequestOptions,
    headerName: string,
    strategy: AuthHeaderStrategy
  ): Promise<T> {
    const authToken = this.client.getAuthToken();
    if (!authToken) {
      return this.client.directRequest<T>(url, options);
    }

    const transformedToken = strategy.transform ? strategy.transform(authToken) : authToken;
    const enhancedOptions = {
      ...options,
      headers: {
        ...options.headers,
        [headerName]: transformedToken
      }
    };

    return this.client.directRequest<T>(url, enhancedOptions);
  }

  /**
   * Get current authentication strategy
   */
  private getStrategy(): AuthHeaderStrategy {
    const authMode = this.client.getAuthMode();
    return this.strategies[authMode] || this.strategies.project;
  }

  /**
   * Detect if error is CORS-related
   */
  private isCORSError(error: any): boolean {
    if (!error) return false;
    
    const corsPatterns = [
      /Access to fetch.*blocked by CORS/,
      /Cross-Origin Request Blocked/,
      /No 'Access-Control-Allow-Origin' header/,
      /CORS error/,
      /Failed to fetch/,
      /Network Error/,
      /TypeError.*Failed to fetch/
    ];
    
    const message = error.message || error.toString();
    return corsPatterns.some(pattern => pattern.test(message)) || error.status === 0;
  }

  /**
   * Create enhanced error with troubleshooting information
   */
  private createEnhancedError(
    originalError: any,
    url: string,
    strategy: AuthHeaderStrategy,
    triedHeaders: string[]
  ): ApiError {
    const corsInfo = this.detectCORSError(originalError, url, triedHeaders);
    
    const error = new Error(
      corsInfo ? this.createCORSErrorMessage(corsInfo) : originalError.message
    ) as ApiError;
    
    error.status = originalError.status || 0;
    error.response = originalError.response || null;
    error.code = originalError.code;
    error.isCorsError = !!corsInfo;
    error.suggestions = corsInfo?.suggestions || [];
    error.triedHeaders = triedHeaders;
    
    return error;
  }

  /**
   * Detect and analyze CORS errors
   */
  private detectCORSError(error: any, url: string, triedHeaders: string[]): CORSErrorInfo | null {
    if (!this.isCORSError(error)) return null;

    const origin = typeof window !== 'undefined' ? window.location.origin : 'unknown';
    
    return {
      type: 'cors',
      origin,
      endpoint: url,
      suggestions: this.generateCORSSuggestions(origin, url, triedHeaders),
      triedHeaders
    };
  }

  /**
   * Generate CORS troubleshooting suggestions
   */
  private generateCORSSuggestions(origin: string, url: string, triedHeaders: string[]): string[] {
    const suggestions = [];
    
    suggestions.push('🔧 CORS Configuration Issue Detected');
    suggestions.push(`Origin: ${origin}`);
    suggestions.push(`Endpoint: ${url}`);
    suggestions.push(`Tried headers: ${triedHeaders.join(', ')}`);
    suggestions.push('');
    
    suggestions.push('💡 Quick Fix Options:');
    suggestions.push(`1. Run: node scripts/setup-cors.js --origin="${origin}"`);
    suggestions.push(`2. Add to .env: CORS_ORIGINS=${origin},http://localhost:*`);
    suggestions.push('3. Restart CloudBox backend after .env changes');
    suggestions.push('');
    
    if (origin.includes('localhost')) {
      suggestions.push('🏠 Development Setup:');
      suggestions.push('For localhost development, add wildcard support:');
      suggestions.push('CORS_ORIGINS=http://localhost:*,https://localhost:*');
      suggestions.push('');
    }
    
    suggestions.push('📚 More help: https://docs.cloudbox.dev/cors-setup');
    
    return suggestions;
  }

  /**
   * Create user-friendly CORS error message
   */
  private createCORSErrorMessage(corsInfo: CORSErrorInfo): string {
    return `CloudBox CORS Error\n\n${corsInfo.suggestions.join('\n')}`;
  }
}

export class CloudBoxClient {
  private readonly config: Required<CloudBoxConfig>;
  private readonly baseUrl: string;
  private authToken?: string;
  private readonly authHeaderManager: AuthHeaderManager;
  
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
  constructor(config: CloudBoxConfig & { authConfig?: AuthConfiguration }) {
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

    // Initialize authentication header manager
    this.authHeaderManager = new AuthHeaderManager(
      this,
      config.authConfig?.strategies || DEFAULT_AUTH_STRATEGIES
    );
    
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
  /**
   * Make an authenticated HTTP request with header fallback strategies
   */
  async request<T = any>(path: string, options: RequestOptions = {}): Promise<T> {
    // Use authentication header manager for enhanced error handling
    const maxRetries = options.maxRetries || 3;
    return this.authHeaderManager.makeAuthenticatedRequest<T>(
      this.buildUrl(path, options),
      options,
      maxRetries
    );
  }

  /**
   * Direct request without header fallback (used internally)
   */
  async directRequest<T = any>(pathOrUrl: string, options: RequestOptions = {}): Promise<T> {
    // Check if pathOrUrl is already a complete URL
    const url = (pathOrUrl.startsWith('http://') || pathOrUrl.startsWith('https://')) 
      ? pathOrUrl 
      : this.buildUrl(pathOrUrl, options);
    
    try {
      const response = await fetch(url, this.buildFetchOptions(pathOrUrl, options));
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
   * Build complete URL for request
   */
  private buildUrl(path: string, options: RequestOptions): string {
    // Check if path is already a full URL (fallback protection)
    if (path.startsWith('http://') || path.startsWith('https://')) {
      // Extract just the path part from the full URL
      try {
        const url = new URL(path);
        path = url.pathname + url.search + url.hash;
      } catch (e) {
        // If parsing fails, use as-is
      }
    }
    
    // Determine if this is an admin endpoint based on path or authMode
    const isAdminEndpoint = path.startsWith('/api/v1/') || 
      (this.config.authMode === 'admin' && (path.startsWith('/users/') || path.includes('/auth')));
    
    const baseUrl = isAdminEndpoint 
      ? `${this.config.endpoint}${path}` 
      : `${this.baseUrl}${path}`;
    
    // Build URL with query parameters
    const requestUrl = new URL(baseUrl);
    if (options.params) {
      Object.entries(options.params).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          requestUrl.searchParams.set(key, String(value));
        }
      });
    }
    
    return requestUrl.toString();
  }

  /**
   * Build fetch options for request
   */
  private buildFetchOptions(path: string, options: RequestOptions): RequestInit {
    // Determine if this is an admin endpoint based on path or authMode
    const isAdminEndpoint = path.startsWith('/api/v1/') || 
      (this.config.authMode === 'admin' && (path.startsWith('/users/') || path.includes('/auth')));
    
    // Prepare request headers with authentication
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      'User-Agent': 'CloudBoxSDK/1.0.0',
      'X-CloudBox-Client': 'SDK',
      'X-CloudBox-Version': '1.0.0',
      ...options.headers
    };

    // Add API key for all requests
    headers['X-API-Key'] = this.config.apiKey;
    
    // Add appropriate authentication headers (handled by AuthHeaderManager)
    // This method handles the base headers, auth headers are added by the manager

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

    return {
      method: options.method || 'GET',
      headers,
      body
    };
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
  withConfig(config: Partial<CloudBoxConfig & { authConfig?: AuthConfiguration }>): CloudBoxClient {
    return new CloudBoxClient({
      ...this.config,
      ...config
    });
  }

  /**
   * Enable CORS error debugging
   */
  enableDebugMode(): void {
    // Debug mode enabled by default in development
  }

  /**
   * Test different authentication headers manually
   */
  async testAuthHeaders(path: string = '/collections'): Promise<{ [header: string]: boolean }> {
    const strategy = this.config.authMode === 'admin' 
      ? DEFAULT_AUTH_STRATEGIES.admin 
      : DEFAULT_AUTH_STRATEGIES.project;
    
    const results: { [header: string]: boolean } = {};
    
    for (const header of [strategy.primary, ...strategy.fallbacks]) {
      try {
        await this.directRequest(path, {
          skipAuthFallback: true,
          headers: {
            [header]: this.authToken || ''
          }
        });
        results[header] = true;
      } catch (error) {
        results[header] = false;
      }
    }
    
    return results;
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

  /**
   * Get authentication header strategies
   */
  getAuthStrategies(): Record<string, AuthHeaderStrategy> {
    return DEFAULT_AUTH_STRATEGIES;
  }

  /**
   * Refresh API discovery for the current project
   * 
   * Triggers a programmatic refresh of the API route discovery system.
   * Useful for:
   * - App updates and deployments
   * - Database schema changes
   * - Template installations/updates
   * 
   * @param options - Refresh options
   * @returns Promise resolving to refresh result
   * 
   * @example
   * ```typescript
   * // Basic refresh after app update
   * const result = await client.refreshAPIDiscovery({
   *   reason: 'App update deployed',
   *   source: 'PhotoPortfolio App v2.1.0'
   * });
   * console.log(`✅ Refreshed ${result.routeCount} API routes`);
   * 
   * // Advanced refresh with webhook
   * await client.refreshAPIDiscovery({
   *   reason: 'Database migration completed',
   *   source: 'Migration Script',
   *   forceRescan: true,
   *   templates: ['photoportfolio'],
   *   webhook: 'https://myapp.com/webhooks/discovery-updated'
   * });
   * ```
   */
  async refreshAPIDiscovery(options: {
    /** Why the refresh was triggered */
    reason?: string;
    /** What app/service triggered this refresh */
    source?: string;
    /** Specific templates to refresh (optional) */
    templates?: string[];
    /** Force full database rescan */
    forceRescan?: boolean;
    /** Optional webhook URL to call when refresh completes */
    webhook?: string;
  } = {}): Promise<{
    success: boolean;
    message: string;
    projectId: number;
    routeCount: number;
    categories: string[];
    triggeredBy: string;
    reason: string;
    discovery: any;
  }> {
    const refreshData = {
      reason: options.reason || 'SDK refresh trigger',
      source: options.source || 'CloudBox SDK',
      templates: options.templates,
      forceRescan: options.forceRescan !== false, // default true
      webhook: options.webhook
    };

    try {
      const result = await this.request('/discovery/refresh', {
        method: 'POST',
        body: refreshData
      });

      return result;
    } catch (error) {
      // Handle and re-throw with better context
      const apiError = error as ApiError;
      apiError.message = `Failed to refresh API discovery: ${apiError.message}`;
      throw apiError;
    }
  }
}