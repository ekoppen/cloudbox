import { CloudBoxError } from './CloudBoxError';
import type { CloudBoxConfig, APIResponse } from '../types';

export class ApiClient {
  private baseUrl: string;
  private apiKey: string;
  private timeout: number;
  private defaultHeaders: Record<string, string>;

  constructor(config: CloudBoxConfig) {
    this.baseUrl = config.endpoint || 'https://api.cloudbox.dev';
    this.apiKey = config.apiKey;
    this.timeout = config.timeout || 30000;
    
    this.defaultHeaders = {
      'Content-Type': 'application/json',
      'X-API-Key': this.apiKey,
      'User-Agent': `CloudBox-SDK-JS/1.0.0`
    };
  }

  private async request<T>(
    method: string,
    endpoint: string,
    data?: any,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`;
    const config: RequestInit = {
      method,
      headers: {
        ...this.defaultHeaders,
        ...options.headers
      },
      ...options
    };

    if (data && method !== 'GET') {
      if (data instanceof FormData) {
        // Don't set content-type for FormData, let browser set it with boundary
        delete config.headers!['Content-Type'];
        config.body = data;
      } else {
        config.body = JSON.stringify(data);
      }
    }

    // Add timeout
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), this.timeout);
    config.signal = controller.signal;

    try {
      const response = await fetch(url, config);
      clearTimeout(timeoutId);

      let responseData;
      const contentType = response.headers.get('content-type');
      
      if (contentType && contentType.includes('application/json')) {
        responseData = await response.json();
      } else {
        responseData = await response.text();
      }

      if (!response.ok) {
        throw CloudBoxError.fromResponse(response, responseData);
      }

      return responseData as T;
    } catch (error) {
      clearTimeout(timeoutId);
      
      if (error instanceof CloudBoxError) {
        throw error;
      }

      if (error instanceof TypeError && error.message.includes('fetch')) {
        throw CloudBoxError.fromNetworkError(error);
      }

      if (error.name === 'AbortError') {
        throw new CloudBoxError('Request timeout', 'TIMEOUT_ERROR');
      }

      throw new CloudBoxError(
        error.message || 'Unknown error occurred',
        'UNKNOWN_ERROR',
        { originalError: error }
      );
    }
  }

  async get<T>(endpoint: string, options?: RequestInit): Promise<T> {
    return this.request<T>('GET', endpoint, undefined, options);
  }

  async post<T>(endpoint: string, data?: any, options?: RequestInit): Promise<T> {
    return this.request<T>('POST', endpoint, data, options);
  }

  async put<T>(endpoint: string, data?: any, options?: RequestInit): Promise<T> {
    return this.request<T>('PUT', endpoint, data, options);
  }

  async patch<T>(endpoint: string, data?: any, options?: RequestInit): Promise<T> {
    return this.request<T>('PATCH', endpoint, data, options);
  }

  async delete<T>(endpoint: string, options?: RequestInit): Promise<T> {
    return this.request<T>('DELETE', endpoint, undefined, options);
  }

  setAuthToken(token: string): void {
    this.defaultHeaders['Authorization'] = `Bearer ${token}`;
  }

  clearAuthToken(): void {
    delete this.defaultHeaders['Authorization'];
  }

  setApiKey(apiKey: string): void {
    this.apiKey = apiKey;
    this.defaultHeaders['X-API-Key'] = apiKey;
  }

  getBaseUrl(): string {
    return this.baseUrl;
  }
}