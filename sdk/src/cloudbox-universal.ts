// CloudBox Universal SDK - BaaS for any project type
// This is the foundation for making CloudBox work like Firebase/Supabase

interface CloudBoxConfig {
  endpoint: string;
  apiKey: string;
  projectId: string | number;
}

interface QueryOptions {
  limit?: number;
  offset?: number;
  orderBy?: Record<string, 'asc' | 'desc'>;
  where?: Record<string, any>;
}

interface Document {
  id: string;
  created_at: string;
  updated_at: string;
  collection_name: string;
  project_id: number;
  data: Record<string, any>;
  version: number;
  author: string;
}

class CloudBoxUniversalClient {
  private config: CloudBoxConfig;
  private baseUrl: string;

  constructor(config: CloudBoxConfig) {
    this.config = config;
    // Use relative URLs in development for Vite proxy
    this.baseUrl = import.meta.env.DEV ? '' : (config.endpoint || 'http://localhost:8080');
  }

  private async request(path: string, options: RequestInit = {}): Promise<any> {
    const projectId = this.config.projectId || '2';
    const url = `${this.baseUrl}/p/${projectId}/api${path}`;
    
    const headers = {
      'Content-Type': 'application/json',
      'X-API-Key': this.config.apiKey,
      ...options.headers,
    };

    // Handle FormData separately (for file uploads)
    if (options.body instanceof FormData) {
      delete headers['Content-Type'];
    }

    const response = await fetch(url, {
      ...options,
      headers,
    });

    if (!response.ok) {
      // Provide helpful error messages for common mistakes
      if (response.status === 404 && options.method !== 'GET') {
        const isLegacyEndpoint = path.match(/^\/(pages|albums|images)$/);
        if (isLegacyEndpoint) {
          const collection = path.substring(1);
          throw new Error(
            `CloudBox API Error: ${response.status} ${response.statusText}. ` +
            `For CRUD operations, use Documents API: /documents/${collection}`
          );
        }
      }
      throw new Error(`CloudBox API Error: ${response.status} ${response.statusText}`);
    }

    return response.json();
  }

  private buildQuery(options: QueryOptions = {}): string {
    const params = new URLSearchParams();
    
    if (options.limit) params.append('limit', options.limit.toString());
    if (options.offset) params.append('offset', options.offset.toString());
    
    const query = params.toString();
    return query ? `?${query}` : '';
  }

  // Universal Database API - works with any collection
  database = {
    // Collections Management
    createCollection: async (name: string, schema: Record<string, any>): Promise<any> => {
      return this.request('/collections', {
        method: 'POST',
        body: JSON.stringify({ name, schema }),
      });
    },

    getCollections: async (): Promise<any[]> => {
      return this.request('/collections');
    },

    getCollection: async (name: string): Promise<any> => {
      return this.request(`/collections/${name}`);
    },

    // Documents CRUD - Universal methods for ANY collection
    create: async (collection: string, data: Record<string, any>): Promise<Document> => {
      return this.request(`/documents/${collection}`, {
        method: 'POST',
        body: JSON.stringify(data),
      });
    },

    findMany: async (collection: string, options: QueryOptions = {}): Promise<Document[]> => {
      const query = this.buildQuery(options);
      return this.request(`/documents/${collection}${query}`);
    },

    findOne: async (collection: string, id: string): Promise<Document> => {
      return this.request(`/documents/${collection}/${id}`);
    },

    update: async (collection: string, id: string, data: Record<string, any>): Promise<Document> => {
      return this.request(`/documents/${collection}/${id}`, {
        method: 'PUT',
        body: JSON.stringify(data),
      });
    },

    delete: async (collection: string, id: string): Promise<void> => {
      return this.request(`/documents/${collection}/${id}`, {
        method: 'DELETE',
      });
    },

    // Advanced queries
    query: async (
      collection: string, 
      filters: Record<string, any>, 
      options: QueryOptions = {}
    ): Promise<Document[]> => {
      return this.request(`/documents/${collection}`, {
        method: 'POST',
        body: JSON.stringify({ filters, options }),
      });
    },

    // Count documents
    count: async (collection: string, filters: Record<string, any> = {}): Promise<number> => {
      const result = await this.request(`/documents/${collection}/count`, {
        method: 'POST',
        body: JSON.stringify({ filters }),
      });
      return result.count;
    }
  };

  // Universal Storage API
  storage = {
    createBucket: async (name: string, options: {
      description?: string;
      isPublic?: boolean;
      maxFileSize?: number;
      allowedMimeTypes?: string[];
    } = {}): Promise<any> => {
      return this.request('/storage/buckets', {
        method: 'POST',
        body: JSON.stringify({
          name,
          description: options.description,
          is_public: options.isPublic,
          max_file_size: options.maxFileSize,
          allowed_mime_types: options.allowedMimeTypes,
        }),
      });
    },

    getBuckets: async (): Promise<any[]> => {
      return this.request('/storage/buckets');
    },

    getBucket: async (name: string): Promise<any> => {
      return this.request(`/storage/buckets/${name}`);
    },

    upload: async (bucket: string, file: File, options: {
      fileName?: string;
      path?: string;
      isPublic?: boolean;
      onProgress?: (progress: number) => void;
    } = {}): Promise<any> => {
      const formData = new FormData();
      formData.append('file', file);
      if (options.fileName) formData.append('filename', options.fileName);
      if (options.path) formData.append('path', options.path);
      if (options.isPublic !== undefined) formData.append('is_public', options.isPublic.toString());

      const projectId = this.config.projectId || '2';
      const url = `${this.baseUrl}/p/${projectId}/api/storage/${bucket}/files`;
      
      return new Promise((resolve, reject) => {
        const xhr = new XMLHttpRequest();
        
        if (options.onProgress) {
          xhr.upload.addEventListener('progress', (e) => {
            if (e.lengthComputable) {
              const progress = (e.loaded / e.total) * 100;
              options.onProgress!(progress);
            }
          });
        }
        
        xhr.addEventListener('load', () => {
          if (xhr.status === 200) {
            resolve(JSON.parse(xhr.responseText));
          } else {
            reject(new Error(`Upload failed: ${xhr.status} ${xhr.statusText}`));
          }
        });
        
        xhr.addEventListener('error', () => {
          reject(new Error('Upload failed'));
        });
        
        xhr.open('POST', url);
        xhr.setRequestHeader('X-API-Key', this.config.apiKey);
        xhr.send(formData);
      });
    },

    getFiles: async (bucket: string): Promise<any[]> => {
      return this.request(`/storage/${bucket}/files`);
    },

    deleteFile: async (bucket: string, fileId: string): Promise<void> => {
      return this.request(`/storage/${bucket}/files/${fileId}`, {
        method: 'DELETE',
      });
    }
  };

  // Universal Auth API
  auth = {
    register: async (email: string, password: string, userData: Record<string, any> = {}): Promise<any> => {
      return this.request('/auth/register', {
        method: 'POST',
        body: JSON.stringify({ email, password, ...userData }),
      });
    },

    login: async (email: string, password: string): Promise<any> => {
      return this.request('/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      });
    },

    logout: async (): Promise<void> => {
      return this.request('/auth/logout', {
        method: 'POST',
      });
    },

    getCurrentUser: async (): Promise<any> => {
      return this.request('/auth/me');
    },

    updateProfile: async (userData: Record<string, any>): Promise<any> => {
      return this.request('/auth/profile', {
        method: 'PUT',
        body: JSON.stringify(userData),
      });
    }
  };

  // Universal Functions API (serverless)
  functions = {
    call: async (functionName: string, data: Record<string, any> = {}): Promise<any> => {
      return this.request(`/functions/${functionName}`, {
        method: 'POST',
        body: JSON.stringify(data),
      });
    },

    list: async (): Promise<any[]> => {
      return this.request('/functions');
    }
  };

  // Analytics API
  analytics = {
    track: async (event: string, properties: Record<string, any> = {}): Promise<any> => {
      return this.request('/analytics/events', {
        method: 'POST',
        body: JSON.stringify({ event, properties }),
      });
    },

    getEvents: async (filters: Record<string, any> = {}): Promise<any[]> => {
      return this.request('/analytics/events', {
        method: 'POST',
        body: JSON.stringify({ filters }),
      });
    }
  };

  // Utility methods
  utils = {
    generateSlug: (text: string): string => {
      return text.toLowerCase()
        .replace(/\s+/g, '-')
        .replace(/[^a-z0-9-]/g, '')
        .replace(/--+/g, '-')
        .replace(/^-|-$/g, '');
    },

    generateId: (): string => {
      return crypto.randomUUID();
    },

    formatDate: (date: Date): string => {
      return date.toISOString();
    }
  };

  // Test connection
  async ping(): Promise<boolean> {
    try {
      const response = await fetch(`${this.baseUrl}/health`);
      return response.ok;
    } catch {
      return false;
    }
  }
}

// Project-specific wrapper example for Portfolio
export class PortfolioAPI extends CloudBoxUniversalClient {
  pages = {
    getPublished: (): Promise<Document[]> => {
      return this.database.query('pages', {
        published: 'true',
        language: 'en'
      }, {
        orderBy: { created_at: 'desc' }
      });
    },

    create: (page: {
      title: string;
      content: any;
      type?: string;
      active?: boolean;
      isMenuPlaceholder?: boolean;
      isHomePage?: boolean;
      parentPageId?: string | null;
      seo?: any;
    }): Promise<Document> => {
      return this.database.create('pages', {
        title: page.title,
        slug: this.utils.generateSlug(page.title),
        content: JSON.stringify(page.content || {}),
        type: page.type || 'content',
        status: page.active ? 'active' : 'inactive',
        language: 'en',
        published: 'true',
        is_menu_placeholder: page.isMenuPlaceholder || false,
        is_home_page: page.isHomePage || false,
        parent_page_id: page.parentPageId || null,
        seo: page.seo ? JSON.stringify(page.seo) : null
      });
    },

    update: (id: string, page: any): Promise<Document> => {
      return this.database.update('pages', id, {
        title: page.title,
        slug: page.path?.replace('/', '') || this.utils.generateSlug(page.title),
        content: JSON.stringify(page.content || {}),
        type: page.type || 'content',
        status: page.active ? 'active' : 'inactive',
        language: 'en',
        published: 'true',
        is_menu_placeholder: page.isMenuPlaceholder || false,
        is_home_page: page.isHomePage || false,
        parent_page_id: page.parentPageId || null,
        seo: page.seo ? JSON.stringify(page.seo) : null
      });
    },

    delete: (id: string): Promise<void> => {
      return this.database.delete('pages', id);
    }
  };

  albums = {
    getAll: (): Promise<Document[]> => {
      return this.database.findMany('albums', {
        where: { status: 'active' },
        orderBy: { created_at: 'desc' }
      });
    },

    create: (album: {
      title: string;
      description?: string;
      images?: any[];
      coverImageId?: string;
    }): Promise<Document> => {
      return this.database.create('albums', {
        title: album.title,
        description: album.description || '',
        status: 'active',
        images: JSON.stringify(album.images || []),
        cover_image_id: album.coverImageId || null
      });
    },

    update: (id: string, album: any): Promise<Document> => {
      return this.database.update('albums', id, {
        title: album.title,
        description: album.description || '',
        status: album.status || 'active',
        images: JSON.stringify(album.images || []),
        cover_image_id: album.cover_image_id || null
      });
    },

    delete: (id: string): Promise<void> => {
      return this.database.delete('albums', id);
    }
  };

  images = {
    getAll: (): Promise<Document[]> => {
      return this.database.findMany('images', {
        orderBy: { created_at: 'desc' }
      });
    },

    create: (image: {
      filename: string;
      title?: string;
      description?: string;
      altText?: string;
      fileId: string;
      url: string;
      thumbnailUrl?: string;
      width?: number;
      height?: number;
      size?: number;
      mimeType?: string;
    }): Promise<Document> => {
      return this.database.create('images', {
        filename: image.filename,
        title: image.title || '',
        description: image.description || '',
        alt_text: image.altText || '',
        file_id: image.fileId,
        url: image.url,
        thumbnail_url: image.thumbnailUrl || null,
        width: image.width || null,
        height: image.height || null,
        size: image.size || null,
        mime_type: image.mimeType || 'image/jpeg'
      });
    },

    update: (id: string, image: any): Promise<Document> => {
      return this.database.update('images', id, {
        filename: image.filename,
        title: image.title || '',
        description: image.description || '',
        alt_text: image.alt_text || '',
        file_id: image.file_id,
        url: image.url,
        thumbnail_url: image.thumbnail_url || null,
        width: image.width || null,
        height: image.height || null,
        size: image.size || null,
        mime_type: image.mime_type || 'image/jpeg'
      });
    },

    delete: (id: string): Promise<void> => {
      return this.database.delete('images', id);
    }
  };
}

export { CloudBoxUniversalClient, type CloudBoxConfig, type Document, type QueryOptions };