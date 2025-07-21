import type { CloudBox } from '../CloudBox';
import type { User, QueryOptions } from '../types';

export class Users {
  private cloudbox: CloudBox;

  constructor(cloudbox: CloudBox) {
    this.cloudbox = cloudbox;
  }

  /**
   * Get all users (admin only)
   */
  async getUsers(options: QueryOptions = {}): Promise<User[]> {
    const params = new URLSearchParams();
    
    if (options.limit) params.append('limit', options.limit.toString());
    if (options.offset) params.append('offset', options.offset.toString());
    if (options.orderBy) params.append('order_by', options.orderBy);
    if (options.orderDirection) params.append('order_direction', options.orderDirection);
    
    // Handle filters
    if (options.filters) {
      for (const [key, value] of Object.entries(options.filters)) {
        params.append(`filter_${key}`, String(value));
      }
    }

    const queryString = params.toString();
    const url = `${this.cloudbox.getAdminApiPath()}/users${queryString ? `?${queryString}` : ''}`;
    
    return this.cloudbox.apiClient.get<User[]>(url);
  }

  /**
   * Get a specific user by ID (admin only)
   */
  async getUser(userId: string): Promise<User> {
    return this.cloudbox.apiClient.get<User>(
      `${this.cloudbox.getAdminApiPath()}/users/${userId}`
    );
  }

  /**
   * Create a new user (admin only)
   */
  async createUser(userData: {
    name: string;
    email: string;
    password: string;
    profile_data?: Record<string, any>;
    preferences?: Record<string, any>;
    is_active?: boolean;
    role?: string;
  }): Promise<User> {
    return this.cloudbox.apiClient.post<User>(
      `${this.cloudbox.getAdminApiPath()}/users`,
      userData
    );
  }

  /**
   * Update a user (admin only)
   */
  async updateUser(userId: string, updates: Partial<{
    name: string;
    email: string;
    profile_data: Record<string, any>;
    preferences: Record<string, any>;
    is_active: boolean;
    role: string;
  }>): Promise<User> {
    return this.cloudbox.apiClient.put<User>(
      `${this.cloudbox.getAdminApiPath()}/users/${userId}`,
      updates
    );
  }

  /**
   * Delete a user (admin only)
   */
  async deleteUser(userId: string): Promise<void> {
    await this.cloudbox.apiClient.delete(
      `${this.cloudbox.getAdminApiPath()}/users/${userId}`
    );
  }

  /**
   * Search users (admin only)
   */
  async searchUsers(query: string, options: QueryOptions = {}): Promise<User[]> {
    const params = new URLSearchParams();
    params.append('q', query);
    
    if (options.limit) params.append('limit', options.limit.toString());
    if (options.offset) params.append('offset', options.offset.toString());

    return this.cloudbox.apiClient.get<User[]>(
      `${this.cloudbox.getAdminApiPath()}/users/search?${params.toString()}`
    );
  }

  /**
   * Count users (admin only)
   */
  async countUsers(filters?: Record<string, any>): Promise<number> {
    const params = new URLSearchParams();
    
    if (filters) {
      for (const [key, value] of Object.entries(filters)) {
        params.append(`filter_${key}`, String(value));
      }
    }

    const queryString = params.toString();
    const url = `${this.cloudbox.getAdminApiPath()}/users/count${queryString ? `?${queryString}` : ''}`;
    
    const result = await this.cloudbox.apiClient.get<{ count: number }>(url);
    return result.count;
  }

  /**
   * Get user statistics (admin only)
   */
  async getUserStats(): Promise<{
    total_users: number;
    active_users: number;
    inactive_users: number;
    new_users_today: number;
    new_users_this_week: number;
    new_users_this_month: number;
    user_growth_rate: number;
  }> {
    return this.cloudbox.apiClient.get(
      `${this.cloudbox.getAdminApiPath()}/users/stats`
    );
  }

  /**
   * Ban a user (admin only)
   */
  async banUser(userId: string, reason?: string, expiresAt?: string): Promise<void> {
    await this.cloudbox.apiClient.post(
      `${this.cloudbox.getAdminApiPath()}/users/${userId}/ban`,
      {
        reason,
        expires_at: expiresAt
      }
    );
  }

  /**
   * Unban a user (admin only)
   */
  async unbanUser(userId: string): Promise<void> {
    await this.cloudbox.apiClient.post(
      `${this.cloudbox.getAdminApiPath()}/users/${userId}/unban`
    );
  }

  /**
   * Reset user password (admin only)
   */
  async resetUserPassword(userId: string, newPassword: string): Promise<void> {
    await this.cloudbox.apiClient.post(
      `${this.cloudbox.getAdminApiPath()}/users/${userId}/reset-password`,
      { password: newPassword }
    );
  }

  /**
   * Send verification email to user (admin only)
   */
  async sendUserVerificationEmail(userId: string): Promise<void> {
    await this.cloudbox.apiClient.post(
      `${this.cloudbox.getAdminApiPath()}/users/${userId}/send-verification`
    );
  }

  /**
   * Verify user manually (admin only)
   */
  async verifyUser(userId: string): Promise<void> {
    await this.cloudbox.apiClient.post(
      `${this.cloudbox.getAdminApiPath()}/users/${userId}/verify`
    );
  }

  /**
   * Get user activity log (admin only)
   */
  async getUserActivity(userId: string, options: {
    limit?: number;
    offset?: number;
    startDate?: string;
    endDate?: string;
  } = {}): Promise<{
    activity_type: string;
    description: string;
    ip_address: string;
    user_agent: string;
    created_at: string;
  }[]> {
    const params = new URLSearchParams();
    
    if (options.limit) params.append('limit', options.limit.toString());
    if (options.offset) params.append('offset', options.offset.toString());
    if (options.startDate) params.append('start_date', options.startDate);
    if (options.endDate) params.append('end_date', options.endDate);

    const queryString = params.toString();
    const url = `${this.cloudbox.getAdminApiPath()}/users/${userId}/activity${queryString ? `?${queryString}` : ''}`;
    
    return this.cloudbox.apiClient.get(url);
  }

  /**
   * Bulk operations (admin only)
   */
  async bulkUpdateUsers(userIds: string[], updates: Partial<{
    is_active: boolean;
    role: string;
  }>): Promise<void> {
    await this.cloudbox.apiClient.put(
      `${this.cloudbox.getAdminApiPath()}/users/bulk`,
      {
        user_ids: userIds,
        updates
      }
    );
  }

  /**
   * Bulk delete users (admin only)
   */
  async bulkDeleteUsers(userIds: string[]): Promise<void> {
    await this.cloudbox.apiClient.delete(
      `${this.cloudbox.getAdminApiPath()}/users/bulk`,
      {
        body: JSON.stringify({ user_ids: userIds }),
        headers: { 'Content-Type': 'application/json' }
      }
    );
  }
}