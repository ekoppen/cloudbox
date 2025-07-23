import type { CloudBox } from '../CloudBox';
import type { Backup, CreateBackupOptions, RestoreBackupOptions } from '../types';

export class Backups {
  private cloudbox: CloudBox;

  constructor(cloudbox: CloudBox) {
    this.cloudbox = cloudbox;
  }

  /**
   * Create a new backup
   */
  async create(options: CreateBackupOptions = {}): Promise<Backup> {
    const backupData = {
      name: options.name || `Backup ${new Date().toISOString()}`,
      description: options.description || 'Manual backup',
      type: options.type || 'manual'
    };

    return this.cloudbox.apiClient.post<Backup>(
      `${this.cloudbox.getAdminApiPath()}/backups`,
      backupData
    );
  }

  /**
   * List all backups for the current project
   */
  async list(): Promise<Backup[]> {
    return this.cloudbox.apiClient.get<Backup[]>(
      `${this.cloudbox.getAdminApiPath()}/backups`
    );
  }

  /**
   * Get a specific backup by ID
   */
  async get(backupId: string | number): Promise<Backup> {
    return this.cloudbox.apiClient.get<Backup>(
      `/api/backups/${backupId}`
    );
  }

  /**
   * Delete a backup
   */
  async delete(backupId: string | number): Promise<void> {
    await this.cloudbox.apiClient.delete(
      `/api/backups/${backupId}`
    );
  }

  /**
   * Restore from a backup
   */
  async restore(backupId: string | number, options: RestoreBackupOptions = {}): Promise<{
    message: string;
    target_project_id: number;
  }> {
    const restoreData = options.targetProjectId 
      ? { target_project_id: options.targetProjectId }
      : {};

    return this.cloudbox.apiClient.post(
      `/api/backups/${backupId}/restore`,
      restoreData
    );
  }

  /**
   * Wait for backup to complete
   */
  async waitForCompletion(backupId: string | number, timeout: number = 300000): Promise<Backup> {
    const startTime = Date.now();
    const pollInterval = 2000; // 2 seconds

    return new Promise((resolve, reject) => {
      const poll = async () => {
        try {
          const backup = await this.get(backupId);
          
          if (backup.status === 'completed') {
            resolve(backup);
            return;
          }
          
          if (backup.status === 'failed') {
            reject(new Error('Backup creation failed'));
            return;
          }
          
          if (Date.now() - startTime > timeout) {
            reject(new Error('Backup timeout'));
            return;
          }
          
          // Continue polling
          setTimeout(poll, pollInterval);
        } catch (error) {
          reject(error);
        }
      };

      poll();
    });
  }

  /**
   * Create backup and wait for completion
   */
  async createAndWait(options: CreateBackupOptions = {}, timeout?: number): Promise<Backup> {
    const backup = await this.create(options);
    return this.waitForCompletion(backup.id, timeout);
  }

  /**
   * Get backup status
   */
  async getStatus(backupId: string | number): Promise<{
    status: 'creating' | 'completed' | 'failed';
    progress?: number;
    message?: string;
  }> {
    const backup = await this.get(backupId);
    return {
      status: backup.status as 'creating' | 'completed' | 'failed',
      progress: backup.status === 'completed' ? 100 : undefined,
      message: backup.status === 'failed' ? 'Backup creation failed' : undefined
    };
  }

  /**
   * Download backup file (if supported)
   */
  async download(backupId: string | number): Promise<Blob> {
    return this.cloudbox.apiClient.get(
      `/api/backups/${backupId}/download`,
      { responseType: 'blob' }
    );
  }

  /**
   * Get backup metadata without full backup object
   */
  async getMetadata(backupId: string | number): Promise<{
    id: number;
    name: string;
    size: number;
    status: string;
    created_at: string;
    completed_at?: string;
    checksum?: string;
  }> {
    const backup = await this.get(backupId);
    return {
      id: backup.id,
      name: backup.name,
      size: backup.size,
      status: backup.status,
      created_at: backup.created_at,
      completed_at: backup.completed_at,
      checksum: backup.checksum
    };
  }

  /**
   * Verify backup integrity
   */
  async verify(backupId: string | number): Promise<{
    valid: boolean;
    checksum_match: boolean;
    size_match: boolean;
    errors?: string[];
  }> {
    return this.cloudbox.apiClient.post(
      `/api/backups/${backupId}/verify`
    );
  }

  /**
   * Clone project using backup/restore
   */
  async cloneProject(
    sourceProjectId: number,
    targetProjectId: number,
    backupName?: string
  ): Promise<{ backup: Backup; restored: boolean }> {
    // Create backup of source project
    const backup = await this.create({
      name: backupName || `Clone from project ${sourceProjectId}`,
      description: `Backup for cloning project ${sourceProjectId} to ${targetProjectId}`,
      type: 'manual'
    });

    // Wait for backup to complete
    const completedBackup = await this.waitForCompletion(backup.id);

    // Restore to target project
    await this.restore(backup.id, { targetProjectId });

    return {
      backup: completedBackup,
      restored: true
    };
  }

  /**
   * Schedule automatic backups (future feature)
   */
  async scheduleBackup(options: {
    name: string;
    cron: string; // cron expression
    description?: string;
    retention_days?: number;
  }): Promise<{ scheduled: boolean; schedule_id: string }> {
    return this.cloudbox.apiClient.post(
      `${this.cloudbox.getAdminApiPath()}/backups/schedule`,
      options
    );
  }

  /**
   * Get backup statistics
   */
  async getStats(): Promise<{
    total_backups: number;
    total_size: number;
    latest_backup?: {
      id: number;
      created_at: string;
      size: number;
    };
    oldest_backup?: {
      id: number;
      created_at: string;
      size: number;
    };
  }> {
    return this.cloudbox.apiClient.get(
      `${this.cloudbox.getAdminApiPath()}/backups/stats`
    );
  }
}