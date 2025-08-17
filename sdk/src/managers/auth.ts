/**
 * CloudBox Authentication Manager
 * Handles user authentication and JWT token management
 */

import type { CloudBoxClient } from '../client';
import type {
  RegisterRequest,
  LoginRequest,
  AuthResponse,
  AuthUser,
  RefreshTokenRequest,
  UpdateProfileRequest,
  ChangePasswordRequest
} from '../types';

/**
 * Authentication manager for CloudBox
 * 
 * Provides methods for:
 * - User registration and login
 * - JWT token management
 * - Profile management
 * - Password management
 * 
 * @example
 * ```typescript
 * // Register new user
 * const { user, token } = await client.auth.register({
 *   email: 'user@example.com',
 *   password: 'securepassword',
 *   name: 'John Doe'
 * });
 * 
 * // Login existing user
 * const { user, token } = await client.auth.login({
 *   email: 'user@example.com', 
 *   password: 'securepassword'
 * });
 * ```
 */
export class AuthManager {
  constructor(private client: CloudBoxClient) {}

  /**
   * Register a new user
   * 
   * @param data - Registration data (email, password, name)
   * @returns Promise resolving to authentication response
   * 
   * @example
   * ```typescript
   * const { user, token, refresh_token } = await client.auth.register({
   *   email: 'user@example.com',
   *   password: 'securepassword123',
   *   name: 'John Doe'
   * });
   * 
   * // Store tokens securely
   * localStorage.setItem('auth_token', token);
   * localStorage.setItem('refresh_token', refresh_token);
   * ```
   */
  async register(data: RegisterRequest): Promise<AuthResponse> {
    return this.client.request<AuthResponse>('/api/v1/auth/register', {
      method: 'POST',
      body: data
    });
  }

  /**
   * Login existing user
   * 
   * @param data - Login credentials (email, password)
   * @returns Promise resolving to authentication response
   * 
   * @example
   * ```typescript
   * const { user, token, refresh_token } = await client.auth.login({
   *   email: 'user@example.com',
   *   password: 'securepassword123'
   * });
   * 
   * // Update client with auth token for future requests
   * client.setAuthToken(token);
   * ```
   */
  async login(data: LoginRequest): Promise<AuthResponse> {
    return this.client.request<AuthResponse>('/api/v1/auth/login', {
      method: 'POST',
      body: data
    });
  }

  /**
   * Refresh authentication token
   * 
   * @param refreshToken - Refresh token from login response
   * @returns Promise resolving to new authentication response
   * 
   * @example
   * ```typescript
   * const refreshToken = localStorage.getItem('refresh_token');
   * const { token, refresh_token } = await client.auth.refresh(refreshToken);
   * 
   * // Update stored tokens
   * localStorage.setItem('auth_token', token);
   * localStorage.setItem('refresh_token', refresh_token);
   * ```
   */
  async refresh(refreshToken: string): Promise<AuthResponse> {
    return this.client.request<AuthResponse>('/api/v1/auth/refresh', {
      method: 'POST',
      body: { refresh_token: refreshToken }
    });
  }

  /**
   * Get current user profile
   * 
   * @returns Promise resolving to current user data
   * 
   * @example
   * ```typescript
   * const user = await client.auth.me();
   * console.log('Current user:', user.email, user.name);
   * ```
   */
  async me(): Promise<AuthUser> {
    return this.client.request<AuthUser>('/api/v1/auth/me');
  }

  /**
   * Update user profile
   * 
   * @param data - Profile update data
   * @returns Promise resolving to updated user data
   * 
   * @example
   * ```typescript
   * const updatedUser = await client.auth.updateProfile({
   *   name: 'John Smith',
   *   email: 'john.smith@example.com'
   * });
   * ```
   */
  async updateProfile(data: UpdateProfileRequest): Promise<AuthUser> {
    return this.client.request<AuthUser>('/api/v1/auth/me', {
      method: 'PUT',
      body: data
    });
  }

  /**
   * Change user password
   * 
   * @param data - Password change data
   * @returns Promise resolving to success message
   * 
   * @example
   * ```typescript
   * await client.auth.changePassword({
   *   current_password: 'oldpassword123',
   *   new_password: 'newpassword456'
   * });
   * ```
   */
  async changePassword(data: ChangePasswordRequest): Promise<{ message: string }> {
    return this.client.request<{ message: string }>('/api/v1/auth/change-password', {
      method: 'PUT',
      body: data
    });
  }

  /**
   * Logout current user
   * 
   * @returns Promise resolving to logout confirmation
   * 
   * @example
   * ```typescript
   * await client.auth.logout();
   * 
   * // Clear stored tokens
   * localStorage.removeItem('auth_token');
   * localStorage.removeItem('refresh_token');
   * ```
   */
  async logout(): Promise<{ message: string }> {
    return this.client.request<{ message: string }>('/api/v1/auth/logout', {
      method: 'POST'
    });
  }
}