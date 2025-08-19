// Storage visualization components
export { default as BucketInterface } from './bucket-interface.svelte';

// Component types
export interface StorageBucket {
  id: string;
  name: string;
  description?: string;
  is_public: boolean;
  max_file_size: number;
  file_count: number;
  total_size: number;
  created_at: string;
  updated_at?: string;
}

export interface StorageFile {
  id: string;
  original_name: string;
  file_name: string;
  file_path: string;
  folder_path?: string;
  mime_type: string;
  size: number;
  bucket_name: string;
  is_public: boolean;
  public_url?: string;
  private_url?: string;
  created_at: string;
}