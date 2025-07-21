import type { CloudBox } from '../CloudBox';
import type { StorageFile, Bucket, UploadOptions } from '../types';

export class Storage {
  private cloudbox: CloudBox;

  constructor(cloudbox: CloudBox) {
    this.cloudbox = cloudbox;
  }

  /**
   * Get all buckets
   */
  async getBuckets(): Promise<Bucket[]> {
    return this.cloudbox.apiClient.get<Bucket[]>(
      `${this.cloudbox.getProjectApiPath()}/storage/buckets`
    );
  }

  /**
   * Create a new bucket
   */
  async createBucket(
    name: string,
    options: {
      description?: string;
      isPublic?: boolean;
      maxFileSize?: number;
      allowedMimeTypes?: string[];
    } = {}
  ): Promise<Bucket> {
    return this.cloudbox.apiClient.post<Bucket>(
      `${this.cloudbox.getProjectApiPath()}/storage/buckets`,
      {
        name,
        description: options.description,
        is_public: options.isPublic || false,
        max_file_size: options.maxFileSize,
        allowed_mime_types: options.allowedMimeTypes
      }
    );
  }

  /**
   * Get a specific bucket
   */
  async getBucket(bucketName: string): Promise<Bucket> {
    return this.cloudbox.apiClient.get<Bucket>(
      `${this.cloudbox.getProjectApiPath()}/storage/buckets/${bucketName}`
    );
  }

  /**
   * Delete a bucket
   */
  async deleteBucket(bucketName: string): Promise<void> {
    await this.cloudbox.apiClient.delete(
      `${this.cloudbox.getProjectApiPath()}/storage/buckets/${bucketName}`
    );
  }

  /**
   * Get all files in a bucket
   */
  async getFiles(bucketName: string, options: {
    limit?: number;
    offset?: number;
    search?: string;
  } = {}): Promise<StorageFile[]> {
    const params = new URLSearchParams();
    if (options.limit) params.append('limit', options.limit.toString());
    if (options.offset) params.append('offset', options.offset.toString());
    if (options.search) params.append('search', options.search);

    const queryString = params.toString();
    const url = `${this.cloudbox.getProjectApiPath()}/storage/${bucketName}/files${queryString ? `?${queryString}` : ''}`;

    return this.cloudbox.apiClient.get<StorageFile[]>(url);
  }

  /**
   * Upload a file to a bucket
   */
  async uploadFile(
    file: File | Blob, 
    options: UploadOptions
  ): Promise<StorageFile> {
    const formData = new FormData();
    
    if (file instanceof File) {
      formData.append('file', file, options.fileName || file.name);
    } else {
      formData.append('file', file, options.fileName || 'blob');
    }

    if (options.isPublic !== undefined) {
      formData.append('is_public', options.isPublic.toString());
    }

    // Handle progress tracking if callback provided
    const config: RequestInit = {};
    if (options.onProgress) {
      // Note: Progress tracking with fetch requires additional implementation
      // This is a simplified version
      config.headers = {
        'X-Upload-Progress': 'true'
      };
    }

    return this.cloudbox.apiClient.post<StorageFile>(
      `${this.cloudbox.getProjectApiPath()}/storage/${options.bucket}/files`,
      formData,
      config
    );
  }

  /**
   * Upload file from URL
   */
  async uploadFromUrl(
    url: string,
    options: Omit<UploadOptions, 'onProgress'>
  ): Promise<StorageFile> {
    return this.cloudbox.apiClient.post<StorageFile>(
      `${this.cloudbox.getProjectApiPath()}/storage/${options.bucket}/files/from-url`,
      {
        url,
        is_public: options.isPublic || false,
        file_name: options.fileName
      }
    );
  }

  /**
   * Get a specific file
   */
  async getFile(bucketName: string, fileId: string): Promise<StorageFile> {
    return this.cloudbox.apiClient.get<StorageFile>(
      `${this.cloudbox.getProjectApiPath()}/storage/${bucketName}/files/${fileId}`
    );
  }

  /**
   * Download a file (returns blob)
   */
  async downloadFile(bucketName: string, fileId: string): Promise<Blob> {
    const response = await fetch(
      `${this.cloudbox.apiClient.getBaseUrl()}${this.cloudbox.getProjectApiPath()}/storage/${bucketName}/files/${fileId}/download`,
      {
        headers: {
          'X-API-Key': this.cloudbox.config.apiKey
        }
      }
    );

    if (!response.ok) {
      throw new Error(`Failed to download file: ${response.statusText}`);
    }

    return response.blob();
  }

  /**
   * Get file download URL
   */
  async getDownloadUrl(bucketName: string, fileId: string, expiresIn?: number): Promise<{ url: string; expires_at: string }> {
    const params = expiresIn ? `?expires_in=${expiresIn}` : '';
    return this.cloudbox.apiClient.get(
      `${this.cloudbox.getProjectApiPath()}/storage/${bucketName}/files/${fileId}/download-url${params}`
    );
  }

  /**
   * Delete a file
   */
  async deleteFile(bucketName: string, fileId: string): Promise<void> {
    await this.cloudbox.apiClient.delete(
      `${this.cloudbox.getProjectApiPath()}/storage/${bucketName}/files/${fileId}`
    );
  }

  /**
   * Update file metadata
   */
  async updateFile(
    bucketName: string, 
    fileId: string, 
    updates: {
      name?: string;
      isPublic?: boolean;
      metadata?: Record<string, any>;
    }
  ): Promise<StorageFile> {
    return this.cloudbox.apiClient.put<StorageFile>(
      `${this.cloudbox.getProjectApiPath()}/storage/${bucketName}/files/${fileId}`,
      {
        name: updates.name,
        is_public: updates.isPublic,
        metadata: updates.metadata
      }
    );
  }

  /**
   * Copy a file to another bucket
   */
  async copyFile(
    sourceBucket: string,
    fileId: string,
    targetBucket: string,
    newName?: string
  ): Promise<StorageFile> {
    return this.cloudbox.apiClient.post<StorageFile>(
      `${this.cloudbox.getProjectApiPath()}/storage/${sourceBucket}/files/${fileId}/copy`,
      {
        target_bucket: targetBucket,
        new_name: newName
      }
    );
  }

  /**
   * Move a file to another bucket
   */
  async moveFile(
    sourceBucket: string,
    fileId: string,
    targetBucket: string,
    newName?: string
  ): Promise<StorageFile> {
    return this.cloudbox.apiClient.post<StorageFile>(
      `${this.cloudbox.getProjectApiPath()}/storage/${sourceBucket}/files/${fileId}/move`,
      {
        target_bucket: targetBucket,
        new_name: newName
      }
    );
  }

  /**
   * Get bucket usage statistics
   */
  async getBucketStats(bucketName: string): Promise<{
    total_files: number;
    total_size: number;
    file_types: Record<string, number>;
  }> {
    return this.cloudbox.apiClient.get(
      `${this.cloudbox.getProjectApiPath()}/storage/${bucketName}/stats`
    );
  }

  /**
   * Batch delete files
   */
  async batchDelete(bucketName: string, fileIds: string[]): Promise<void> {
    await this.cloudbox.apiClient.delete(
      `${this.cloudbox.getProjectApiPath()}/storage/${bucketName}/files/batch`,
      {
        body: JSON.stringify({ file_ids: fileIds }),
        headers: { 'Content-Type': 'application/json' }
      }
    );
  }

  /**
   * Create a signed upload URL for direct client uploads
   */
  async createUploadUrl(
    bucketName: string,
    fileName: string,
    contentType: string,
    options: {
      expiresIn?: number;
      maxSize?: number;
      isPublic?: boolean;
    } = {}
  ): Promise<{
    upload_url: string;
    file_id: string;
    expires_at: string;
  }> {
    return this.cloudbox.apiClient.post(
      `${this.cloudbox.getProjectApiPath()}/storage/${bucketName}/upload-url`,
      {
        file_name: fileName,
        content_type: contentType,
        expires_in: options.expiresIn || 3600, // 1 hour default
        max_size: options.maxSize,
        is_public: options.isPublic || false
      }
    );
  }
}