/**
 * CloudBox SDK Types
 */

// Base types
export interface CloudBoxConfig {
  projectId?: string;
  projectSlug?: string;
  apiKey: string;
  endpoint?: string;
  timeout?: number;
}

export interface APIResponse<T = any> {
  data?: T;
  error?: string;
  message?: string;
}

// Auth types
export interface User {
  id: string;
  email: string;
  name: string;
  created_at: string;
  updated_at: string;
  is_active: boolean;
  is_email_verified: boolean;
  profile_data?: Record<string, any>;
  preferences?: Record<string, any>;
  last_login_at?: string;
  last_seen_at?: string;
}

export interface AuthSession {
  id: string;
  user_id: string;
  token: string;
  expires_at: string;
  created_at: string;
  user_agent?: string;
  ip_address?: string;
}

export interface AuthCredentials {
  email: string;
  password: string;
}

export interface AuthResponse {
  user: User;
  session: AuthSession;
  token: string;
}

// Database types
export interface Document {
  id: string;
  created_at: string;
  updated_at: string;
  [key: string]: any;
}

export interface Collection {
  id: string;
  name: string;
  description?: string;
  schema?: Record<string, any>;
  created_at: string;
  updated_at: string;
}

export interface QueryOptions {
  limit?: number;
  offset?: number;
  orderBy?: string;
  orderDirection?: 'asc' | 'desc';
  filters?: Record<string, any>;
}

// Storage types
export interface StorageFile {
  id: string;
  name: string;
  size: number;
  mime_type: string;
  bucket_name: string;
  file_path: string;
  is_public: boolean;
  public_url?: string;
  private_url?: string;
  created_at: string;
  updated_at: string;
}

export interface Bucket {
  id: string;
  name: string;
  description?: string;
  is_public: boolean;
  max_file_size?: number;
  allowed_mime_types?: string[];
  created_at: string;
  updated_at: string;
}

export interface UploadOptions {
  bucket: string;
  isPublic?: boolean;
  fileName?: string;
  onProgress?: (progress: number) => void;
}

// Functions types
export interface CloudFunction {
  id: string;
  name: string;
  description?: string;
  runtime: string;
  language: string;
  status: 'draft' | 'building' | 'deployed' | 'error';
  version: number;
  function_url: string;
  is_active: boolean;
  is_public: boolean;
  timeout: number;
  memory: number;
  last_deployed_at?: string;
  created_at: string;
  updated_at: string;
}

export interface FunctionExecution {
  execution_id: string;
  status: 'success' | 'error' | 'timeout';
  execution_time: number;
  response?: any;
  logs?: string;
  error_message?: string;
  created_at: string;
}

export interface ExecuteFunctionOptions {
  data?: Record<string, any>;
  headers?: Record<string, string>;
  timeout?: number;
}

export interface CreateFunctionOptions {
  name: string;
  description?: string;
  runtime: 'nodejs18' | 'nodejs16' | 'nodejs14' | 'python3.9' | 'python3.8' | 'python3.7' | 'go1.19' | 'go1.18';
  language: 'javascript' | 'python' | 'go';
  code: string;
  entry_point?: string;
  timeout?: number;
  memory?: number;
  environment?: Record<string, any>;
  commands?: string[];
  dependencies?: Record<string, any>;
  is_public?: boolean;
}

// Messaging types
export interface Channel {
  id: string;
  name: string;
  description?: string;
  type: 'public' | 'private' | 'direct';
  is_active: boolean;
  member_count: number;
  created_at: string;
  updated_at: string;
}

export interface Message {
  id: string;
  channel_id: string;
  user_id: string;
  content: string;
  message_type: 'text' | 'image' | 'file' | 'system';
  metadata?: Record<string, any>;
  created_at: string;
  updated_at: string;
  user?: User;
}

export interface ChannelMember {
  id: string;
  channel_id: string;
  user_id: string;
  role: 'admin' | 'member';
  joined_at: string;
  user?: User;
}

// Realtime types
export interface RealtimeSubscription {
  unsubscribe: () => void;
}

export interface RealtimeMessage {
  event: string;
  payload: any;
  timestamp: string;
}

export type RealtimeCallback = (message: RealtimeMessage) => void;

// Backup types
export interface Backup {
  id: number;
  name: string;
  description?: string;
  type: 'manual' | 'automatic';
  status: 'creating' | 'completed' | 'failed';
  size: number;
  file_path?: string;
  checksum?: string;
  created_at: string;
  completed_at?: string;
  project_id: number;
}

export interface CreateBackupOptions {
  name?: string;
  description?: string;
  type?: 'manual' | 'automatic';
}

export interface RestoreBackupOptions {
  targetProjectId?: number;
}

// Error types
export interface CloudBoxErrorDetails {
  code: string;
  message: string;
  details?: Record<string, any>;
}