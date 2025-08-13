/**
 * Storage Manager - CloudBox SDK
 * 
 * Manages file storage, buckets, and public access with full type safety
 */

import {
  Bucket,
  StorageFile,
  UploadFileOptions,
  CreateBucketRequest,
  PublicUrlResponse,
  BatchPublicUrlsRequest,
  BatchPublicUrlsResponse
} from '../types';

import type { CloudBoxClient } from '../client';

export class StorageManager {
  constructor(private client: CloudBoxClient) {}

  // BUCKET OPERATIONS

  /**
   * Create a new storage bucket
   */
  async createBucket(config: CreateBucketRequest): Promise<Bucket> {
    return this.client.request<Bucket>('/storage/buckets', {
      method: 'POST',
      body: {
        name: config.name,
        description: config.description || '',
        max_file_size: config.max_file_size || 52428800, // 50MB default
        allowed_types: config.allowed_types || [],
        is_public: config.is_public !== undefined ? config.is_public : false
      }
    });
  }

  /**
   * List all storage buckets
   */
  async listBuckets(): Promise<Bucket[]> {
    return this.client.request<Bucket[]>('/storage/buckets');
  }

  /**
   * Get a specific bucket by name
   */
  async getBucket(name: string): Promise<Bucket> {
    return this.client.request<Bucket>(`/storage/buckets/${name}`);
  }

  /**
   * Delete a bucket and all its files
   */
  async deleteBucket(name: string): Promise<void> {
    await this.client.request(`/storage/buckets/${name}`, {
      method: 'DELETE'
    });
  }

  /**
   * Update bucket settings
   */
  async updateBucket(name: string, updates: Partial<CreateBucketRequest>): Promise<Bucket> {
    return this.client.request<Bucket>(`/storage/buckets/${name}`, {
      method: 'PUT',
      body: updates
    });
  }

  // FILE OPERATIONS

  /**
   * Upload a file to a bucket
   */
  async uploadFile(bucketName: string, options: UploadFileOptions): Promise<StorageFile> {
    const formData = new FormData();
    
    // Add the file
    formData.append('file', options.file, options.fileName);
    
    // Add metadata if provided
    if (options.metadata) {
      Object.entries(options.metadata).forEach(([key, value]) => {
        formData.append(`metadata[${key}]`, String(value));
      });
    }
    
    // Add folder if provided
    if (options.folder) {
      formData.append('folder', options.folder);
    }

    return this.client.request<StorageFile>(`/storage/buckets/${bucketName}/files`, {
      method: 'POST',
      body: formData
    });
  }

  /**
   * List files in a bucket
   */
  async listFiles(
    bucketName: string, 
    options: {
      limit?: number;
      offset?: number;
      folder?: string;
    } = {}
  ): Promise<StorageFile[]> {
    return this.client.request<StorageFile[]>(`/storage/buckets/${bucketName}/files`, {
      params: options
    });
  }

  /**
   * Get file information by ID
   */
  async getFile(bucketName: string, fileId: string): Promise<StorageFile> {
    return this.client.request<StorageFile>(`/storage/buckets/${bucketName}/files/${fileId}`);
  }

  /**
   * Delete a file
   */
  async deleteFile(bucketName: string, fileId: string): Promise<void> {
    await this.client.request(`/storage/buckets/${bucketName}/files/${fileId}`, {
      method: 'DELETE'
    });
  }

  /**
   * Update file metadata
   */
  async updateFileMetadata(
    bucketName: string, 
    fileId: string, 
    metadata: Record<string, any>
  ): Promise<StorageFile> {
    return this.client.request<StorageFile>(`/storage/buckets/${bucketName}/files/${fileId}`, {
      method: 'PUT',
      body: { metadata }
    });
  }

  // PUBLIC ACCESS OPERATIONS

  /**
   * Get public URL for a file (requires public bucket)
   */
  async getPublicUrl(bucketName: string, fileId: string): Promise<PublicUrlResponse> {
    return this.client.request<PublicUrlResponse>(`/storage/buckets/${bucketName}/files/${fileId}/public-url`);
  }

  /**
   * Get public URLs for multiple files
   */
  async getBatchPublicUrls(
    bucketName: string, 
    fileIds: string[]
  ): Promise<BatchPublicUrlsResponse> {
    return this.client.request<BatchPublicUrlsResponse>(`/storage/buckets/${bucketName}/public-urls`, {
      method: 'POST',
      body: { file_ids: fileIds }
    });
  }

  /**
   * Generate CloudBox public file URL
   */
  generatePublicFileUrl(projectSlug: string, bucketName: string, filePath: string): string {
    const baseUrl = this.client.getConfig().endpoint || 'http://localhost:8080';
    return `${baseUrl}/public/${projectSlug}/${bucketName}/${filePath}`;
  }

  // BULK OPERATIONS

  /**
   * Delete multiple files at once
   */
  async deleteFiles(bucketName: string, fileIds: string[]): Promise<void> {
    await this.client.request(`/storage/buckets/${bucketName}/files/batch`, {
      method: 'DELETE',
      body: { file_ids: fileIds }
    });
  }

  /**
   * Get total storage usage for the project
   */
  async getStorageStats(): Promise<{
    total_size: number;
    total_files: number;
    buckets: Array<{
      name: string;
      size: number;
      file_count: number;
    }>;
  }> {
    return this.client.request('/storage/stats');
  }
}