/**
 * CloudBox SDK TypeScript Definitions
 * Complete type safety for all CloudBox API operations
 */

// ===============================
// CORE CONFIGURATION TYPES
// ===============================

export interface CloudBoxConfig {
  /** Project ID (numeric) for API routes */
  projectId: number;
  /** API key for authentication */
  apiKey: string;
  /** CloudBox endpoint URL (default: http://localhost:8080) */
  endpoint?: string;
  /** Authentication mode - 'project' for project APIs, 'admin' for CloudBox admin APIs (default: 'project') */
  authMode?: 'admin' | 'project';
}

export interface RequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';
  headers?: Record<string, string>;
  body?: any;
  params?: Record<string, any>;
}

// ===============================
// API RESPONSE TYPES
// ===============================

export interface ApiResponse<T = any> {
  data?: T;
  message?: string;
  error?: string;
  code?: string;
  details?: any;
}

export interface ApiError extends Error {
  status: number;
  response: any;
  code?: string;
}

// ===============================
// COLLECTION & DOCUMENT TYPES
// ===============================

export interface Collection {
  id: number;
  name: string;
  description?: string;
  schema?: Record<string, SchemaField> | Record<string, any>;
  indexes?: string[];
  project_id: number;
  document_count: number;
  last_modified: string;
  created_at: string;
  updated_at: string;
}

export interface CreateCollectionRequest {
  name: string;
  description?: string;
  schema?: Record<string, SchemaField> | Record<string, any>;
  indexes?: string[];
}

export interface SchemaField {
  name: string;
  type: 'string' | 'number' | 'boolean' | 'object' | 'array' | 'date';
  required?: boolean;
  default?: any;
  unique?: boolean;
  index?: boolean;
  validation?: {
    min?: number;
    max?: number;
    pattern?: string;
    enum?: any[];
  };
}

export interface Document {
  id: string;
  [key: string]: any;
  created_at: string;
  updated_at: string;
}

export interface CreateDocumentRequest {
  [key: string]: any;
}

export interface UpdateDocumentRequest {
  [key: string]: any;
}

export interface ListDocumentsOptions {
  limit?: number;
  offset?: number;
  orderBy?: string;
  filter?: string | Record<string, any>;
}

export interface ListDocumentsResponse {
  documents: Document[];
  total: number;
  limit: number;
  offset: number;
}

export interface CountDocumentsResponse {
  count: number;
}

export interface BatchCreateRequest {
  documents: CreateDocumentRequest[];
}

export interface BatchCreateResponse {
  documents: Document[];
  count: number;
}

export interface BatchDeleteRequest {
  ids: string[];
}

export interface BatchDeleteResponse {
  message: string;
  count: number;
}

// ===============================
// STORAGE TYPES
// ===============================

export interface Bucket {
  id: number;
  name: string;
  description?: string;
  max_file_size: number;
  allowed_types: string[];
  is_public: boolean;
  project_id: number;
  file_count: number;
  total_size: number;
  last_modified: string;
  created_at: string;
  updated_at: string;
}

export interface StorageFile {
  id: string;
  original_name: string;
  file_name: string;
  file_path: string;
  mime_type: string;
  size: number;
  bucket_name: string;
  project_id: number;
  storage_path: string;
  is_public?: boolean;
  public_url?: string;
  metadata?: Record<string, any>;
  created_at: string;
  updated_at: string;
}

export interface UploadFileOptions {
  file: File | Blob;
  fileName?: string;
  metadata?: Record<string, any>;
  folder?: string;
}

export interface CreateBucketRequest {
  name: string;
  description?: string;
  max_file_size?: number;
  allowed_types?: string[];
  is_public?: boolean;
}

export interface PublicUrlResponse {
  public_url: string;
  is_public: boolean;
  file: StorageFile;
  bucket: Bucket;
  shareable_url: string;
}

export interface BatchPublicUrlsRequest {
  file_ids: string[];
}

export interface BatchPublicUrlsResponse {
  public_urls: Record<string, string>;
  is_public: boolean;
  bucket: Bucket;
  file_count: number;
}

// ===============================
// AUTHENTICATION TYPES
// ===============================

export interface RegisterRequest {
  email: string;
  password: string;
  name?: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface AuthResponse {
  user: AuthUser;
  token: string;
  refresh_token: string;
  expires_at: string;
}

export interface AuthUser {
  id: number;
  email: string;
  name?: string;
  role: string;
  is_active: boolean;
  last_login_at?: string;
  created_at: string;
  updated_at: string;
}

export interface RefreshTokenRequest {
  refresh_token: string;
}

export interface UpdateProfileRequest {
  name?: string;
  email?: string;
}

export interface ChangePasswordRequest {
  current_password: string;
  new_password: string;
}

// ===============================
// USER MANAGEMENT TYPES
// ===============================

export interface User {
  id: string;
  email: string;
  username?: string;
  first_name?: string;
  last_name?: string;
  avatar_url?: string;
  is_active: boolean;
  is_verified: boolean;
  metadata?: Record<string, any>;
  created_at: string;
  updated_at: string;
  last_login_at?: string;
}

export interface CreateUserRequest {
  email: string;
  password: string;
  username?: string;
  first_name?: string;
  last_name?: string;
  metadata?: Record<string, any>;
}

export interface UpdateUserRequest {
  email?: string;
  username?: string;
  first_name?: string;
  last_name?: string;
  avatar_url?: string;
  is_active?: boolean;
  metadata?: Record<string, any>;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  user: User;
  token: string;
  expires_at: string;
}

// ===============================
// AUTH SETTINGS TYPES
// ===============================

export interface AuthSettings {
  id: number;
  project_id: number;
  allow_registration: boolean;
  require_email_verification: boolean;
  password_min_length: number;
  enable_social_login: boolean;
  enabled_providers: string[];
  jwt_secret: string;
  session_duration: number;
  max_login_attempts: number;
  lockout_duration: number;
  created_at: string;
  updated_at: string;
}

export interface UpdateAuthSettingsRequest {
  allow_registration?: boolean;
  require_email_verification?: boolean;
  password_min_length?: number;
  enable_social_login?: boolean;
  enabled_providers?: string[];
  session_duration?: number;
  max_login_attempts?: number;
  lockout_duration?: number;
}

// ===============================
// FUNCTION TYPES
// ===============================

export interface CloudFunction {
  id: string;
  name: string;
  description?: string;
  code: string;
  runtime: string;
  environment_variables: Record<string, string>;
  is_active: boolean;
  project_id: number;
  created_at: string;
  updated_at: string;
}

export interface ExecuteFunctionRequest {
  [key: string]: any;
}

export interface ExecuteFunctionResponse {
  result: any;
  execution_time: number;
  memory_used: number;
  logs?: string[];
}

// ===============================
// UTILITY TYPES
// ===============================

export interface ConnectionTestResult {
  success: boolean;
  message: string;
  endpoint: string;
  project_id: number;
  response_time?: number;
  api_version?: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
  has_next: boolean;
  has_prev: boolean;
}

// ===============================
// EVENT TYPES (Future)
// ===============================

export interface WebhookEvent {
  id: string;
  event_type: string;
  resource_type: string;
  resource_id: string;
  data: any;
  timestamp: string;
  project_id: number;
}

// ===============================
// QUERY BUILDER TYPES
// ===============================

export interface QueryFilter {
  field: string;
  operator: 'eq' | 'ne' | 'gt' | 'gte' | 'lt' | 'lte' | 'in' | 'nin' | 'contains' | 'starts_with' | 'ends_with';
  value: any;
}

export interface QuerySort {
  field: string;
  direction: 'asc' | 'desc';
}

export interface QueryOptions {
  filters?: QueryFilter[];
  sort?: QuerySort[];
  limit?: number;
  offset?: number;
  select?: string[];
}