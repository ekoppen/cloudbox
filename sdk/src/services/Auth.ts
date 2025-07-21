import { EventEmitter } from 'eventemitter3';
import type { CloudBox } from '../CloudBox';
import type { User, AuthCredentials, AuthResponse, AuthSession } from '../types';

export class Auth extends EventEmitter {
  private cloudbox: CloudBox;
  private currentUser: User | null = null;
  private currentSession: AuthSession | null = null;

  constructor(cloudbox: CloudBox) {
    super();
    this.cloudbox = cloudbox;
  }

  /**
   * Register a new user
   */
  async register(email: string, password: string, name: string, profileData?: Record<string, any>): Promise<AuthResponse> {
    const response = await this.cloudbox.apiClient.post<AuthResponse>(
      `${this.cloudbox.getProjectApiPath()}/users`,
      {
        email,
        password,
        name,
        profile_data: profileData
      }
    );

    this.currentUser = response.user;
    this.currentSession = response.session;
    
    this.emit('login', response.user);
    return response;
  }

  /**
   * Login with email and password
   */
  async login(credentials: AuthCredentials): Promise<AuthResponse> {
    const response = await this.cloudbox.apiClient.post<AuthResponse>(
      `${this.cloudbox.getProjectApiPath()}/users/login`,
      credentials
    );

    this.currentUser = response.user;
    this.currentSession = response.session;
    
    this.emit('login', response.user);
    return response;
  }

  /**
   * Logout current user
   */
  async logout(): Promise<void> {
    if (this.currentSession) {
      try {
        await this.cloudbox.apiClient.post(
          `${this.cloudbox.getProjectApiPath()}/users/logout`,
          { session_id: this.currentSession.id }
        );
      } catch (error) {
        // Continue with logout even if API call fails
        console.warn('Failed to logout from server:', error);
      }
    }

    this.currentUser = null;
    this.currentSession = null;
    
    this.emit('logout');
  }

  /**
   * Get current authenticated user
   */
  async getCurrentUser(): Promise<User> {
    if (!this.currentSession) {
      throw new Error('No active session');
    }

    const user = await this.cloudbox.apiClient.get<User>(
      `${this.cloudbox.getProjectApiPath()}/users/me`,
      {
        headers: {
          'Authorization': `Bearer ${this.currentSession.token}`
        }
      }
    );

    this.currentUser = user;
    return user;
  }

  /**
   * Update user profile
   */
  async updateProfile(updates: Partial<Pick<User, 'name' | 'profile_data' | 'preferences'>>): Promise<User> {
    if (!this.currentUser || !this.currentSession) {
      throw new Error('No active session');
    }

    const user = await this.cloudbox.apiClient.put<User>(
      `${this.cloudbox.getProjectApiPath()}/users/${this.currentUser.id}`,
      updates,
      {
        headers: {
          'Authorization': `Bearer ${this.currentSession.token}`
        }
      }
    );

    this.currentUser = user;
    return user;
  }

  /**
   * Change user password
   */
  async changePassword(currentPassword: string, newPassword: string): Promise<void> {
    if (!this.currentUser || !this.currentSession) {
      throw new Error('No active session');
    }

    await this.cloudbox.apiClient.put(
      `${this.cloudbox.getProjectApiPath()}/users/${this.currentUser.id}/password`,
      {
        current_password: currentPassword,
        new_password: newPassword
      },
      {
        headers: {
          'Authorization': `Bearer ${this.currentSession.token}`
        }
      }
    );
  }

  /**
   * Send password reset email
   */
  async sendPasswordResetEmail(email: string): Promise<void> {
    await this.cloudbox.apiClient.post(
      `${this.cloudbox.getProjectApiPath()}/users/password-reset`,
      { email }
    );
  }

  /**
   * Reset password with token
   */
  async resetPassword(token: string, newPassword: string): Promise<void> {
    await this.cloudbox.apiClient.post(
      `${this.cloudbox.getProjectApiPath()}/users/password-reset/confirm`,
      {
        token,
        password: newPassword
      }
    );
  }

  /**
   * Send email verification
   */
  async sendEmailVerification(): Promise<void> {
    if (!this.currentUser || !this.currentSession) {
      throw new Error('No active session');
    }

    await this.cloudbox.apiClient.post(
      `${this.cloudbox.getProjectApiPath()}/users/${this.currentUser.id}/verify-email`,
      {},
      {
        headers: {
          'Authorization': `Bearer ${this.currentSession.token}`
        }
      }
    );
  }

  /**
   * Verify email with token
   */
  async verifyEmail(token: string): Promise<void> {
    await this.cloudbox.apiClient.post(
      `${this.cloudbox.getProjectApiPath()}/users/verify-email/confirm`,
      { token }
    );

    // Refresh current user to update verification status
    if (this.currentUser && this.currentSession) {
      await this.getCurrentUser();
    }
  }

  /**
   * Get user sessions
   */
  async getSessions(): Promise<AuthSession[]> {
    if (!this.currentUser || !this.currentSession) {
      throw new Error('No active session');
    }

    return this.cloudbox.apiClient.get<AuthSession[]>(
      `${this.cloudbox.getProjectApiPath()}/users/${this.currentUser.id}/sessions`,
      {
        headers: {
          'Authorization': `Bearer ${this.currentSession.token}`
        }
      }
    );
  }

  /**
   * Revoke a session
   */
  async revokeSession(sessionId: string): Promise<void> {
    if (!this.currentUser || !this.currentSession) {
      throw new Error('No active session');
    }

    await this.cloudbox.apiClient.delete(
      `${this.cloudbox.getProjectApiPath()}/users/${this.currentUser.id}/sessions/${sessionId}`,
      {
        headers: {
          'Authorization': `Bearer ${this.currentSession.token}`
        }
      }
    );
  }

  /**
   * Get current user (cached)
   */
  get user(): User | null {
    return this.currentUser;
  }

  /**
   * Get current session (cached)
   */
  get session(): AuthSession | null {
    return this.currentSession;
  }

  /**
   * Check if user is authenticated
   */
  get isAuthenticated(): boolean {
    return !!(this.currentUser && this.currentSession);
  }

  /**
   * Get auth token for API requests
   */
  get token(): string | null {
    return this.currentSession?.token || null;
  }
}