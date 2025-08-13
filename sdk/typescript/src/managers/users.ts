/**
 * User Manager - CloudBox SDK
 * 
 * Manages user authentication, registration, and user data with full type safety
 */

import {
  User,
  CreateUserRequest,
  UpdateUserRequest,
  LoginRequest,
  LoginResponse,
  AuthSettings,
  UpdateAuthSettingsRequest
} from '../types';

import type { CloudBoxClient } from '../client';

export class UserManager {
  constructor(private client: CloudBoxClient) {}

  // USER CRUD OPERATIONS

  /**
   * Create a new user
   */
  async create(userData: CreateUserRequest): Promise<User> {
    return this.client.request<User>('/users', {
      method: 'POST',
      body: userData
    });
  }

  /**
   * List all users in the project
   */
  async list(options: {
    limit?: number;
    offset?: number;
    filter?: Record<string, any>;
  } = {}): Promise<User[]> {
    return this.client.request<User[]>('/users', {
      params: options
    });
  }

  /**
   * Get a specific user by ID
   */
  async get(userId: string): Promise<User> {
    return this.client.request<User>(`/users/${userId}`);
  }

  /**
   * Update a user's information
   */
  async update(userId: string, updates: UpdateUserRequest): Promise<User> {
    return this.client.request<User>(`/users/${userId}`, {
      method: 'PUT',
      body: updates
    });
  }

  /**
   * Delete a user
   */
  async delete(userId: string): Promise<void> {
    await this.client.request(`/users/${userId}`, {
      method: 'DELETE'
    });
  }

  // AUTHENTICATION OPERATIONS

  /**
   * Authenticate a user with email and password
   */
  async login(credentials: LoginRequest): Promise<LoginResponse> {
    return this.client.request<LoginResponse>('/auth/login', {
      method: 'POST',
      body: credentials
    });
  }

  /**
   * Register a new user (if registration is enabled)
   */
  async register(userData: CreateUserRequest): Promise<LoginResponse> {
    return this.client.request<LoginResponse>('/auth/register', {
      method: 'POST',
      body: userData
    });
  }

  /**
   * Logout a user (invalidate token)
   */
  async logout(token: string): Promise<void> {
    await this.client.request('/auth/logout', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });
  }

  /**
   * Verify user's email address
   */
  async verifyEmail(verificationCode: string): Promise<void> {
    await this.client.request('/auth/verify-email', {
      method: 'POST',
      body: { code: verificationCode }
    });
  }

  /**
   * Request password reset
   */
  async requestPasswordReset(email: string): Promise<void> {
    await this.client.request('/auth/reset-password', {
      method: 'POST',
      body: { email }
    });
  }

  /**
   * Reset password with reset code
   */
  async resetPassword(resetCode: string, newPassword: string): Promise<void> {
    await this.client.request('/auth/reset-password/confirm', {
      method: 'POST',
      body: { 
        code: resetCode,
        password: newPassword 
      }
    });
  }

  // USER SEARCH & FILTERING

  /**
   * Search users by email or username
   */
  async search(query: string, options: {
    limit?: number;
    offset?: number;
    searchFields?: string[];
  } = {}): Promise<User[]> {
    return this.client.request<User[]>('/users/search', {
      params: {
        q: query,
        ...options,
        search_fields: options.searchFields?.join(',')
      }
    });
  }

  /**
   * Get user by email address
   */
  async getByEmail(email: string): Promise<User> {
    return this.client.request<User>(`/users/by-email/${encodeURIComponent(email)}`);
  }

  /**
   * Get user by username
   */
  async getByUsername(username: string): Promise<User> {
    return this.client.request<User>(`/users/by-username/${username}`);
  }

  // BULK OPERATIONS

  /**
   * Create multiple users at once
   */
  async createMany(users: CreateUserRequest[]): Promise<User[]> {
    return this.client.request<User[]>('/users/batch', {
      method: 'POST',
      body: { users }
    });
  }

  /**
   * Delete multiple users by IDs
   */
  async deleteMany(userIds: string[]): Promise<void> {
    await this.client.request('/users/batch', {
      method: 'DELETE',
      body: { user_ids: userIds }
    });
  }

  // AUTH SETTINGS MANAGEMENT

  /**
   * Get authentication settings for the project
   */
  async getAuthSettings(): Promise<AuthSettings> {
    return this.client.request<AuthSettings>('/auth/settings');
  }

  /**
   * Update authentication settings
   */
  async updateAuthSettings(settings: UpdateAuthSettingsRequest): Promise<AuthSettings> {
    return this.client.request<AuthSettings>('/auth/settings', {
      method: 'PUT',
      body: settings
    });
  }

  // USER STATISTICS

  /**
   * Get user statistics for the project
   */
  async getStats(): Promise<{
    total_users: number;
    active_users: number;
    verified_users: number;
    new_users_today: number;
    new_users_this_week: number;
    new_users_this_month: number;
  }> {
    return this.client.request('/users/stats');
  }
}