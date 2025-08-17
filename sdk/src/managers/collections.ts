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
  ListDocumentsResponse,
  CountDocumentsResponse,
  BatchCreateRequest,
  BatchCreateResponse,
  BatchDeleteRequest,
  BatchDeleteResponse,
  CreateCollectionRequest,
  QueryOptions,
  QueryFilter,
  QuerySort
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
   * Create a new collection with schema validation
   * 
   * @param data - Collection creation data
   * @returns Promise resolving to created collection
   * 
   * @example
   * ```typescript
   * const collection = await client.collections.create({
   *   name: 'goals',
   *   description: 'User goals collection',
   *   schema: {
   *     user_id: { type: 'string', required: true },
   *     title: { type: 'string', required: true, maxLength: 100 },
   *     is_active: { type: 'boolean', default: true }
   *   },
   *   indexes: ['user_id', 'created_at']
   * });
   * ```
   */
  async create(data: CreateCollectionRequest): Promise<Collection> {
    return this.client.request<Collection>('/collections', {
      method: 'POST',
      body: data
    });
  }

  /**
   * Create a new collection (legacy method - backward compatibility)
   * 
   * @deprecated Use create(data: CreateCollectionRequest) instead
   */
  async createLegacy(
    name: string, 
    schema: SchemaField[] = [], 
    description?: string
  ): Promise<Collection> {
    return this.create({
      name,
      description,
      schema: schema.reduce((acc, field) => {
        acc[field.name] = field;
        return acc;
      }, {} as Record<string, SchemaField>)
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
   * List documents in a collection with pagination
   * 
   * @param collection - Collection name
   * @param options - List options (limit, offset, orderBy, filter)
   * @returns Promise resolving to paginated documents response
   * 
   * @example
   * ```typescript
   * const response = await client.collections.listDocuments('goals', {
   *   limit: 25,
   *   offset: 0,
   *   orderBy: 'created_at DESC',
   *   filter: JSON.stringify({ user_id: 'user123' })
   * });
   * 
   * console.log(response.documents); // Array of documents
   * console.log(response.total);     // Total count
   * ```
   */
  async listDocuments(
    collection: string, 
    options: ListDocumentsOptions = {}
  ): Promise<ListDocumentsResponse> {
    return this.client.request<ListDocumentsResponse>(`/data/${collection}`, {
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
   * 
   * @param collection - Collection name
   * @param query - Query options with filters, sorting, pagination
   * @returns Promise resolving to query results
   * 
   * @example
   * ```typescript
   * const results = await client.collections.query('goals', {
   *   filters: [
   *     { field: 'user_id', operator: 'eq', value: 'user123' },
   *     { field: 'is_active', operator: 'eq', value: true }
   *   ],
   *   sort: [
   *     { field: 'created_at', direction: 'desc' }
   *   ],
   *   limit: 10,
   *   offset: 0
   * });
   * ```
   */
  async query(
    collection: string, 
    query: QueryOptions
  ): Promise<{ data: Document[]; total: number; limit: number; offset: number }> {
    const requestBody: any = {};
    
    if (query.limit) requestBody.limit = query.limit;
    if (query.offset) requestBody.offset = query.offset;
    if (query.select) requestBody.select = query.select;
    
    if (query.filters) {
      requestBody.filters = query.filters.map(filter => ({
        field: filter.field,
        operator: filter.operator,
        value: filter.value
      }));
    }
    
    if (query.sort) {
      requestBody.sort = query.sort.map(sort => ({
        field: sort.field,
        direction: sort.direction
      }));
    }

    return this.client.request<{ data: Document[]; total: number; limit: number; offset: number }>(
      `/data/${collection}/query`, 
      { 
        method: 'POST',
        body: requestBody
      }
    );
  }

  /**
   * Count documents in a collection
   * 
   * @param collection - Collection name
   * @param filter - Optional filter to count specific documents
   * @returns Promise resolving to document count
   * 
   * @example
   * ```typescript
   * const totalGoals = await client.collections.count('goals');
   * const activeGoals = await client.collections.count('goals', { is_active: true });
   * ```
   */
  async count(collection: string, filter?: Record<string, any>): Promise<number> {
    const params = filter ? { filter: JSON.stringify(filter) } : {};
    const result = await this.client.request<CountDocumentsResponse>(
      `/data/${collection}/count`, 
      { params }
    );
    return result.count;
  }

  /**
   * Create multiple documents at once (batch operation)
   * 
   * @param collection - Collection name
   * @param documents - Array of documents to create
   * @returns Promise resolving to created documents with count
   * 
   * @example
   * ```typescript
   * const result = await client.collections.batchCreate('goals', [
   *   { title: 'Goal 1', user_id: 'user123' },
   *   { title: 'Goal 2', user_id: 'user123' },
   *   { title: 'Goal 3', user_id: 'user123' }
   * ]);
   * 
   * console.log(`Created ${result.count} documents`);
   * ```
   */
  async batchCreate(
    collection: string, 
    documents: CreateDocumentRequest[]
  ): Promise<BatchCreateResponse> {
    return this.client.request<BatchCreateResponse>(`/data/${collection}/batch`, {
      method: 'POST',
      body: { documents }
    });
  }

  /**
   * Delete multiple documents by IDs (batch operation)
   * 
   * @param collection - Collection name
   * @param ids - Array of document IDs to delete
   * @returns Promise resolving to deletion result
   * 
   * @example
   * ```typescript
   * const result = await client.collections.batchDelete('goals', [
   *   'goal1', 'goal2', 'goal3'
   * ]);
   * 
   * console.log(result.message); // "Documents deleted successfully"
   * ```
   */
  async batchDelete(collection: string, ids: string[]): Promise<BatchDeleteResponse> {
    return this.client.request<BatchDeleteResponse>(`/data/${collection}/batch`, {
      method: 'DELETE',
      body: { ids }
    });
  }

  // Legacy methods for backward compatibility
  
  /**
   * @deprecated Use batchCreate instead
   */
  async createMany(collection: string, documents: CreateDocumentRequest[]): Promise<Document[]> {
    const result = await this.batchCreate(collection, documents);
    return result.documents;
  }

  /**
   * @deprecated Use batchDelete instead
   */
  async deleteMany(collection: string, ids: string[]): Promise<void> {
    await this.batchDelete(collection, ids);
  }

  /**
   * MongoDB-style find method for easier migration
   * 
   * @param collection - Collection name
   * @param options - MongoDB-style find options
   * @returns Promise resolving to documents
   * 
   * @example
   * ```typescript
   * const goals = await client.collections.find('goals', {
   *   filter: { user_id: 'user123', is_active: true },
   *   sort: { created_at: -1 },
   *   limit: 10,
   *   skip: 0
   * });
   * ```
   */
  async find(
    collection: string, 
    options: {
      filter?: Record<string, any>;
      sort?: Record<string, 1 | -1>;
      limit?: number;
      skip?: number;
      select?: string[];
    } = {}
  ): Promise<Document[]> {
    // Convert MongoDB-style options to CloudBox query format
    const queryOptions: QueryOptions = {};

    if (options.filter) {
      queryOptions.filters = Object.entries(options.filter).map(([field, value]) => ({
        field,
        operator: 'eq' as const,
        value
      }));
    }

    if (options.sort) {
      queryOptions.sort = Object.entries(options.sort).map(([field, direction]) => ({
        field,
        direction: direction === 1 ? 'asc' as const : 'desc' as const
      }));
    }

    if (options.limit) queryOptions.limit = options.limit;
    if (options.skip) queryOptions.offset = options.skip;
    if (options.select) queryOptions.select = options.select;

    const result = await this.query(collection, queryOptions);
    return result.data;
  }
}