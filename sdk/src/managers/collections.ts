/**
 * Collections Manager - CloudBox SDK
 * 
 * Manages collections and documents with full type safety
 */

import {
  Collection,
  Document,
  SchemaField,
  CreateDocumentRequest,
  UpdateDocumentRequest,
  ListDocumentsOptions,
  PaginatedResponse,
  QueryOptions
} from '../types';

import type { CloudBoxClient } from '../client';

export class CollectionManager {
  constructor(private client: CloudBoxClient) {}

  /**
   * List all collections in the project
   */
  async list(): Promise<Collection[]> {
    return this.client.request<Collection[]>('/collections');
  }

  /**
   * Get a specific collection by name
   */
  async get(name: string): Promise<Collection> {
    return this.client.request<Collection>(`/collections/${name}`);
  }

  /**
   * Create a new collection
   */
  async create(
    name: string, 
    schema: SchemaField[] = [], 
    description?: string
  ): Promise<Collection> {
    return this.client.request<Collection>('/collections', {
      method: 'POST',
      body: {
        name,
        schema,
        description
      }
    });
  }

  /**
   * Delete a collection
   */
  async delete(name: string): Promise<void> {
    await this.client.request(`/collections/${name}`, {
      method: 'DELETE'
    });
  }

  // DOCUMENT OPERATIONS

  /**
   * List documents in a collection
   */
  async listDocuments(
    collection: string, 
    options: ListDocumentsOptions = {}
  ): Promise<Document[]> {
    return this.client.request<Document[]>(`/data/${collection}`, {
      params: options
    });
  }

  /**
   * Get a specific document by ID
   */
  async getDocument(collection: string, id: string): Promise<Document> {
    return this.client.request<Document>(`/data/${collection}/${id}`);
  }

  /**
   * Create a new document
   */
  async createDocument(
    collection: string, 
    data: CreateDocumentRequest
  ): Promise<Document> {
    return this.client.request<Document>(`/data/${collection}`, {
      method: 'POST',
      body: data
    });
  }

  /**
   * Update an existing document
   */
  async updateDocument(
    collection: string, 
    id: string, 
    data: UpdateDocumentRequest
  ): Promise<Document> {
    return this.client.request<Document>(`/data/${collection}/${id}`, {
      method: 'PUT',
      body: data
    });
  }

  /**
   * Delete a document
   */
  async deleteDocument(collection: string, id: string): Promise<void> {
    await this.client.request(`/data/${collection}/${id}`, {
      method: 'DELETE'
    });
  }

  /**
   * Query documents with advanced filtering and sorting
   */
  async query(
    collection: string, 
    query: QueryOptions
  ): Promise<PaginatedResponse<Document>> {
    const params: any = {};
    
    if (query.limit) params.limit = query.limit;
    if (query.offset) params.offset = query.offset;
    if (query.select) params.select = query.select.join(',');
    
    if (query.filters) {
      query.filters.forEach((filter, index) => {
        params[`filter[${index}][field]`] = filter.field;
        params[`filter[${index}][operator]`] = filter.operator;
        params[`filter[${index}][value]`] = filter.value;
      });
    }
    
    if (query.sort) {
      query.sort.forEach((sort, index) => {
        params[`sort[${index}][field]`] = sort.field;
        params[`sort[${index}][direction]`] = sort.direction;
      });
    }

    return this.client.request<PaginatedResponse<Document>>(
      `/data/${collection}/query`, 
      { params }
    );
  }

  /**
   * Count documents in a collection
   */
  async count(collection: string, filter?: Record<string, any>): Promise<number> {
    const params = filter ? { filter } : {};
    const result = await this.client.request<{ count: number }>(
      `/data/${collection}/count`, 
      { params }
    );
    return result.count;
  }

  /**
   * Create multiple documents at once
   */
  async createMany(
    collection: string, 
    documents: CreateDocumentRequest[]
  ): Promise<Document[]> {
    return this.client.request<Document[]>(`/data/${collection}/batch`, {
      method: 'POST',
      body: { documents }
    });
  }

  /**
   * Delete multiple documents by IDs
   */
  async deleteMany(collection: string, ids: string[]): Promise<void> {
    await this.client.request(`/data/${collection}/batch`, {
      method: 'DELETE',
      body: { ids }
    });
  }
}