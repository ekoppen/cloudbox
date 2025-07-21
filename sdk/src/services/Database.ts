import type { CloudBox } from '../CloudBox';
import type { Document, Collection, QueryOptions } from '../types';

export class Database {
  private cloudbox: CloudBox;

  constructor(cloudbox: CloudBox) {
    this.cloudbox = cloudbox;
  }

  /**
   * Create a new collection
   */
  async createCollection(name: string, schema?: Record<string, any>): Promise<Collection> {
    return this.cloudbox.apiClient.post<Collection>(
      `${this.cloudbox.getProjectApiPath()}/collections`,
      { name, schema }
    );
  }

  /**
   * Get all collections
   */
  async getCollections(): Promise<Collection[]> {
    return this.cloudbox.apiClient.get<Collection[]>(
      `${this.cloudbox.getProjectApiPath()}/collections`
    );
  }

  /**
   * Get a specific collection
   */
  async getCollection(collectionName: string): Promise<Collection> {
    return this.cloudbox.apiClient.get<Collection>(
      `${this.cloudbox.getProjectApiPath()}/collections/${collectionName}`
    );
  }

  /**
   * Delete a collection
   */
  async deleteCollection(collectionName: string): Promise<void> {
    await this.cloudbox.apiClient.delete(
      `${this.cloudbox.getProjectApiPath()}/collections/${collectionName}`
    );
  }

  /**
   * Create a new document in a collection
   */
  async createDocument(collectionName: string, data: Record<string, any>): Promise<Document> {
    return this.cloudbox.apiClient.post<Document>(
      `${this.cloudbox.getProjectApiPath()}/documents/${collectionName}`,
      data
    );
  }

  /**
   * Get documents from a collection
   */
  async getDocuments(collectionName: string, options: QueryOptions = {}): Promise<Document[]> {
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
    const url = `${this.cloudbox.getProjectApiPath()}/documents/${collectionName}${queryString ? `?${queryString}` : ''}`;
    
    return this.cloudbox.apiClient.get<Document[]>(url);
  }

  /**
   * Get a specific document by ID
   */
  async getDocument(collectionName: string, documentId: string): Promise<Document> {
    return this.cloudbox.apiClient.get<Document>(
      `${this.cloudbox.getProjectApiPath()}/documents/${collectionName}/${documentId}`
    );
  }

  /**
   * Update a document
   */
  async updateDocument(collectionName: string, documentId: string, data: Record<string, any>): Promise<Document> {
    return this.cloudbox.apiClient.put<Document>(
      `${this.cloudbox.getProjectApiPath()}/documents/${collectionName}/${documentId}`,
      data
    );
  }

  /**
   * Delete a document
   */
  async deleteDocument(collectionName: string, documentId: string): Promise<void> {
    await this.cloudbox.apiClient.delete(
      `${this.cloudbox.getProjectApiPath()}/documents/${collectionName}/${documentId}`
    );
  }

  /**
   * Query documents with advanced filtering
   */
  async query(collectionName: string, filters: Record<string, any>, options: QueryOptions = {}): Promise<Document[]> {
    return this.getDocuments(collectionName, { ...options, filters });
  }

  /**
   * Search documents by text
   */
  async search(collectionName: string, query: string, options: QueryOptions = {}): Promise<Document[]> {
    const params = new URLSearchParams();
    params.append('q', query);
    
    if (options.limit) params.append('limit', options.limit.toString());
    if (options.offset) params.append('offset', options.offset.toString());

    return this.cloudbox.apiClient.get<Document[]>(
      `${this.cloudbox.getProjectApiPath()}/documents/${collectionName}/search?${params.toString()}`
    );
  }

  /**
   * Count documents in a collection
   */
  async count(collectionName: string, filters?: Record<string, any>): Promise<number> {
    const params = new URLSearchParams();
    
    if (filters) {
      for (const [key, value] of Object.entries(filters)) {
        params.append(`filter_${key}`, String(value));
      }
    }

    const queryString = params.toString();
    const url = `${this.cloudbox.getProjectApiPath()}/documents/${collectionName}/count${queryString ? `?${queryString}` : ''}`;
    
    const result = await this.cloudbox.apiClient.get<{ count: number }>(url);
    return result.count;
  }

  /**
   * Batch create documents
   */
  async batchCreate(collectionName: string, documents: Record<string, any>[]): Promise<Document[]> {
    return this.cloudbox.apiClient.post<Document[]>(
      `${this.cloudbox.getProjectApiPath()}/documents/${collectionName}/batch`,
      { documents }
    );
  }

  /**
   * Batch update documents
   */
  async batchUpdate(collectionName: string, updates: { id: string; data: Record<string, any> }[]): Promise<Document[]> {
    return this.cloudbox.apiClient.put<Document[]>(
      `${this.cloudbox.getProjectApiPath()}/documents/${collectionName}/batch`,
      { updates }
    );
  }

  /**
   * Batch delete documents
   */
  async batchDelete(collectionName: string, documentIds: string[]): Promise<void> {
    await this.cloudbox.apiClient.delete(
      `${this.cloudbox.getProjectApiPath()}/documents/${collectionName}/batch`,
      {
        body: JSON.stringify({ ids: documentIds }),
        headers: { 'Content-Type': 'application/json' }
      }
    );
  }
}