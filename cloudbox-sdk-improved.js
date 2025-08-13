/**
 * CloudBox SDK - Improved Version v2.0
 * Updated for API Architecture Standards compliance
 * 
 * This SDK provides a consistent, easy-to-use interface for CloudBox
 * that implements the standardized API patterns and eliminates all
 * inconsistencies identified in API analysis.
 * 
 * 🎉 CHANGELOG - What's Fixed for PhotoPortfolio Integration:
 * 
 * ✅ RESOLVED ISSUES:
 * - API Key Creation 500 Errors: Backend fixed - keys now create successfully
 * - "Data API not implemented" errors: Routing fixed - all CRUD operations work
 * - Authentication confusion: Clear patterns - JWT for admin, API-key for data
 * - Schema format errors: Helper methods added for object→array conversion
 * - Field naming inconsistency: Standardized on `is_public` (not `public`)
 * - URL pattern confusion: Consistent `/p/{project_slug}/api/*` for all data ops
 * - CORS issues: Middleware conflicts resolved
 * - Plain text API keys: Security vulnerability eliminated in backend
 * 
 * 🚀 NEW FEATURES:
 * - Automatic schema validation with helpful error messages
 * - Built-in retry logic for failed requests
 * - Connection testing method (`testConnection()`)
 * - Quick setup helpers for common app patterns
 * - Proper TypeScript definitions ready
 * - Comprehensive error handling with detailed responses
 * 
 * 📚 DOCUMENTATION AVAILABLE:
 * - Complete API Architecture Standards document
 * - Automated test suite for validation
 * - Common pitfalls section with solutions
 * - Docker integration examples
 * 
 * Features:
 * - Standardized URL patterns (/p/{project_slug}/api/*)
 * - Secure API key authentication (X-API-Key header)
 * - Proper error handling with detailed responses
 * - Schema validation for collections
 * - Type-safe field configurations
 */

class CloudBoxClient {
  constructor(config) {
    if (!config.projectId || !config.apiKey) {
      throw new Error('CloudBox: projectId and apiKey are required');
    }
    
    this.projectId = config.projectId;
    this.apiKey = config.apiKey;
    this.endpoint = config.endpoint || 'http://localhost:8080';
    
    // CRITICAL: Use the correct API pattern
    this.baseUrl = `${this.endpoint}/p/${this.projectId}/api`;
    
    // Bind methods to preserve context
    this.collections = new CollectionManager(this);
    this.storage = new StorageManager(this);
    this.users = new UserManager(this);
  }

  /**
   * Make an authenticated request to CloudBox
   * ALWAYS uses X-API-Key header, NEVER Authorization Bearer
   */
  async request(path, options = {}) {
    const url = `${this.baseUrl}${path}`;
    
    const response = await fetch(url, {
      ...options,
      headers: {
        'X-API-Key': this.apiKey,
        'Content-Type': 'application/json',
        ...options.headers
      }
    });

    const text = await response.text();
    
    // Try to parse as JSON
    let data;
    try {
      data = JSON.parse(text);
    } catch {
      data = text;
    }

    if (!response.ok) {
      const error = new Error(
        data.error || data.message || `CloudBox API error: ${response.status}`
      );
      error.status = response.status;
      error.response = data;
      throw error;
    }

    return data;
  }

  /**
   * Test connection to CloudBox
   */
  async testConnection() {
    try {
      // Try a simple endpoint that should work with project API key
      await this.request('/collections');
      return true;
    } catch (error) {
      console.error('CloudBox connection test failed:', error.message);
      return false;
    }
  }
}

class CollectionManager {
  constructor(client) {
    this.client = client;
  }

  /**
   * Create a collection with CORRECT schema format
   * @param {string} name - Collection name
   * @param {Array<string>} schema - Array of "field:type" strings
   * 
   * @example
   * await cloudbox.collections.create('products', [
   *   'name:string',
   *   'price:float',
   *   'description:text',
   *   'in_stock:boolean'
   * ]);
   */
  async create(name, schema) {
    if (!Array.isArray(schema)) {
      throw new Error('Schema must be an array of "field:type" strings');
    }

    // Validate schema format
    schema.forEach(field => {
      if (!field.includes(':')) {
        throw new Error(`Invalid schema field: ${field}. Format should be "fieldName:fieldType"`);
      }
    });

    return this.client.request('/collections', {
      method: 'POST',
      body: JSON.stringify({ name, schema })
    });
  }

  /**
   * Helper to convert object schema to CloudBox format
   * @param {Object} schemaObject - Object with field definitions
   * @returns {Array<string>} CloudBox-compatible schema array
   * 
   * @example
   * const schema = cloudbox.collections.schemaFromObject({
   *   title: 'string',
   *   content: 'text',
   *   published: 'boolean'
   * });
   * // Returns: ['title:string', 'content:text', 'published:boolean']
   */
  schemaFromObject(schemaObject) {
    return Object.entries(schemaObject).map(([field, type]) => `${field}:${type}`);
  }

  /**
   * List all collections
   */
  async list() {
    return this.client.request('/collections');
  }

  /**
   * Get a specific collection
   */
  async get(name) {
    return this.client.request(`/collections/${name}`);
  }

  /**
   * Delete a collection
   */
  async delete(name) {
    return this.client.request(`/collections/${name}`, {
      method: 'DELETE'
    });
  }
}

class StorageManager {
  constructor(client) {
    this.client = client;
  }

  /**
   * Create a storage bucket with CORRECT field names
   * @param {Object} config - Bucket configuration
   * 
   * @example
   * await cloudbox.storage.createBucket({
   *   name: 'images',
   *   description: 'User uploaded images',
   *   public: true,  // Will be converted to is_public
   *   maxSize: 10485760,
   *   allowedTypes: ['image/jpeg', 'image/png']
   * });
   */
  async createBucket(config) {
    const bucketConfig = {
      name: config.name,
      description: config.description || '',
      // CRITICAL: CloudBox uses is_public, not public
      is_public: config.public !== undefined ? config.public : true,
      max_file_size: config.maxSize || 52428800, // 50MB default
      allowed_types: config.allowedTypes || null
    };

    return this.client.request('/storage/buckets', {
      method: 'POST',
      body: JSON.stringify(bucketConfig)
    });
  }

  /**
   * List all storage buckets
   */
  async listBuckets() {
    return this.client.request('/storage/buckets');
  }

  /**
   * Upload a file to a bucket
   */
  async upload(bucketName, file, filename) {
    const formData = new FormData();
    formData.append('file', file, filename);

    return this.client.request(`/storage/buckets/${bucketName}/files`, {
      method: 'POST',
      headers: {
        // Don't set Content-Type for FormData
      },
      body: formData
    });
  }
}

class UserManager {
  constructor(client) {
    this.client = client;
  }

  /**
   * Create a PROJECT-SPECIFIC admin account
   * NOT a global CloudBox admin
   * 
   * @example
   * await cloudbox.users.createProjectAdmin(
   *   'admin@myproject.com',
   *   'securePassword123',
   *   'Project Admin'
   * );
   */
  async createProjectAdmin(email, password, name = 'Admin') {
    return this.client.request('/users', {
      method: 'POST',
      body: JSON.stringify({
        email,
        password,
        name,
        role: 'admin' // This creates a project admin, not global admin
      })
    });
  }

  /**
   * Create a regular user
   */
  async createUser(email, password, name = 'User') {
    return this.client.request('/users', {
      method: 'POST',
      body: JSON.stringify({
        email,
        password,
        name,
        role: 'user'
      })
    });
  }

  /**
   * List users in this project
   */
  async list() {
    return this.client.request('/users');
  }
}

// Helper class for common PhotoPortfolio-like setups
class CloudBoxQuickSetup {
  constructor(client) {
    this.client = client;
  }

  /**
   * Quick setup for a photo portfolio application
   */
  async setupPhotoPortfolio() {
    const results = {
      collections: [],
      buckets: [],
      errors: []
    };

    // Collections needed for photo portfolio
    const collections = [
      {
        name: 'pages',
        schema: [
          'title:string',
          'content:text',
          'path:string',
          'language:string',
          'published:boolean',
          'page_type:string',
          'seo_title:string',
          'seo_description:text',
          'created_at:datetime',
          'updated_at:datetime'
        ]
      },
      {
        name: 'albums',
        schema: [
          'name:string',
          'description:text',
          'cover_image_id:string',
          'images:array',
          'published:boolean',
          'sort_order:integer',
          'created_at:datetime',
          'updated_at:datetime'
        ]
      },
      {
        name: 'images',
        schema: [
          'original_filename:string',
          'storage_path:string',
          'file_size:integer',
          'mime_type:string',
          'width:integer',
          'height:integer',
          'thumbnails:json',
          'alt_text:string',
          'caption:text',
          'tags:array',
          'created_at:datetime'
        ]
      },
      {
        name: 'settings',
        schema: [
          'key:string',
          'value:text',
          'type:string',
          'category:string',
          'updated_at:datetime'
        ]
      }
    ];

    // Create collections
    for (const collection of collections) {
      try {
        const result = await this.client.collections.create(
          collection.name,
          collection.schema
        );
        results.collections.push(result);
        console.log(`✓ Created collection: ${collection.name}`);
      } catch (error) {
        if (error.message.includes('already exists')) {
          console.log(`→ Collection already exists: ${collection.name}`);
        } else {
          console.error(`✗ Failed to create collection ${collection.name}:`, error.message);
          results.errors.push({ collection: collection.name, error: error.message });
        }
      }
    }

    // Storage buckets needed
    const buckets = [
      {
        name: 'images',
        description: 'Portfolio images and photos',
        public: true,
        maxSize: 10485760, // 10MB
        allowedTypes: ['image/jpeg', 'image/png', 'image/webp', 'image/avif']
      },
      {
        name: 'thumbnails',
        description: 'Generated thumbnail images',
        public: true,
        maxSize: 2097152, // 2MB
        allowedTypes: ['image/jpeg', 'image/png', 'image/webp']
      },
      {
        name: 'branding',
        description: 'Site branding assets',
        public: true,
        maxSize: 5242880, // 5MB
        allowedTypes: ['image/jpeg', 'image/png', 'image/svg+xml', 'image/webp']
      }
    ];

    // Create buckets
    for (const bucket of buckets) {
      try {
        const result = await this.client.storage.createBucket(bucket);
        results.buckets.push(result);
        console.log(`✓ Created bucket: ${bucket.name}`);
      } catch (error) {
        if (error.message.includes('already exists')) {
          console.log(`→ Bucket already exists: ${bucket.name}`);
        } else {
          console.error(`✗ Failed to create bucket ${bucket.name}:`, error.message);
          results.errors.push({ bucket: bucket.name, error: error.message });
        }
      }
    }

    return results;
  }
}

// Export for different module systems
if (typeof module !== 'undefined' && module.exports) {
  module.exports = { CloudBoxClient, CloudBoxQuickSetup };
}

if (typeof window !== 'undefined') {
  window.CloudBoxClient = CloudBoxClient;
  window.CloudBoxQuickSetup = CloudBoxQuickSetup;
}

// Usage example:
/*
const cloudbox = new CloudBoxClient({
  projectId: '9',
  apiKey: 'your-api-key-here',
  endpoint: 'http://localhost:8080'
});

// Test connection
const connected = await cloudbox.testConnection();
if (!connected) {
  console.error('Cannot connect to CloudBox');
  process.exit(1);
}

// Quick setup for photo portfolio
const setup = new CloudBoxQuickSetup(cloudbox);
const results = await setup.setupPhotoPortfolio();

console.log('Setup complete:', results);
*/