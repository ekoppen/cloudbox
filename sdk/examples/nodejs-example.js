#!/usr/bin/env node

/**
 * CloudBox SDK Node.js Example
 * 
 * This example demonstrates how to use the CloudBox SDK in a Node.js environment
 * for backend applications, scripts, and server-side operations.
 * 
 * Usage:
 *   node nodejs-example.js
 * 
 * Make sure to set your environment variables:
 *   CLOUDBOX_API_KEY=your-api-key
 *   CLOUDBOX_PROJECT_ID=2
 */

// In a real project, you would import from '@ekoppen/cloudbox-sdk'
// import { CloudBox } from '@ekoppen/cloudbox-sdk';

// For this demo, we'll simulate the SDK
class CloudBoxSDKDemo {
  constructor(config) {
    this.config = config;
    this.database = new DatabaseDemo();
    this.functions = new FunctionsDemo();
    this.storage = new StorageDemo();
    this.users = new UsersDemo();
  }

  async ping() {
    console.log('🏓 Pinging CloudBox API...');
    return true;
  }

  getProjectApiPath() {
    return `/p/${this.config.projectId}/api`;
  }

  getAdminApiPath() {
    return `/api/v1/projects/${this.config.projectId}`;
  }
}

class DatabaseDemo {
  constructor() {
    this.collections = new Map();
    this.documents = new Map();
  }

  async createCollection(name, schema) {
    console.log(`📚 Creating collection: ${name}`);
    const collection = {
      id: `col_${Date.now()}`,
      name,
      schema: schema || {},
      created_at: new Date().toISOString(),
      document_count: 0
    };
    
    this.collections.set(name, collection);
    this.documents.set(name, []);
    
    return collection;
  }

  async createDocument(collectionName, data) {
    console.log(`📄 Creating document in ${collectionName}:`, data);
    
    if (!this.documents.has(collectionName)) {
      throw new Error(`Collection ${collectionName} does not exist`);
    }

    const document = {
      id: `doc_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      ...data,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    };

    this.documents.get(collectionName).push(document);
    
    // Update collection document count
    const collection = this.collections.get(collectionName);
    if (collection) {
      collection.document_count++;
    }

    return document;
  }

  async getDocuments(collectionName, options = {}) {
    console.log(`📋 Getting documents from ${collectionName}`);
    
    if (!this.documents.has(collectionName)) {
      return [];
    }

    let docs = [...this.documents.get(collectionName)];
    
    // Apply filters
    if (options.filters) {
      docs = docs.filter(doc => {
        return Object.entries(options.filters).every(([key, value]) => {
          return doc[key] === value;
        });
      });
    }

    // Apply ordering
    if (options.orderBy) {
      docs.sort((a, b) => {
        const aVal = a[options.orderBy];
        const bVal = b[options.orderBy];
        
        if (options.orderDirection === 'desc') {
          return bVal > aVal ? 1 : -1;
        } else {
          return aVal > bVal ? 1 : -1;
        }
      });
    }

    // Apply limit and offset
    if (options.offset) {
      docs = docs.slice(options.offset);
    }
    if (options.limit) {
      docs = docs.slice(0, options.limit);
    }

    return docs;
  }

  async batchCreate(collectionName, documents) {
    console.log(`📦 Batch creating ${documents.length} documents in ${collectionName}`);
    
    const results = [];
    for (const data of documents) {
      const doc = await this.createDocument(collectionName, data);
      results.push(doc);
    }
    
    return results;
  }

  async search(collectionName, query, options = {}) {
    console.log(`🔍 Searching in ${collectionName} for: "${query}"`);
    
    if (!this.documents.has(collectionName)) {
      return [];
    }

    const docs = this.documents.get(collectionName);
    const searchResults = docs.filter(doc => {
      return Object.values(doc).some(value => {
        if (typeof value === 'string') {
          return value.toLowerCase().includes(query.toLowerCase());
        }
        return false;
      });
    });

    return searchResults.slice(0, options.limit || 10);
  }
}

class FunctionsDemo {
  constructor() {
    this.functions = new Map();
  }

  async createJavaScriptFunction(name, description, customCode) {
    console.log(`⚡ Creating JavaScript function: ${name}`);
    
    const defaultCode = `exports.handler = async (event, context) => {
  console.log('Function invoked with event:', event);
  
  return {
    statusCode: 200,
    body: {
      message: 'Hello from CloudBox Function!',
      timestamp: new Date().toISOString(),
      input: event
    }
  };
};`;

    const func = {
      id: `func_${Date.now()}`,
      name,
      description: description || `JavaScript function: ${name}`,
      runtime: 'nodejs18',
      language: 'javascript',
      code: customCode || defaultCode,
      entry_point: 'index.handler',
      is_public: true,
      is_active: true,
      status: 'deployed',
      created_at: new Date().toISOString()
    };

    this.functions.set(name, func);
    return func;
  }

  async execute(functionName, data = {}, options = {}) {
    console.log(`🚀 Executing function: ${functionName}`);
    console.log(`   Input data:`, data);
    
    if (!this.functions.has(functionName)) {
      throw new Error(`Function ${functionName} not found`);
    }

    // Simulate function execution
    const startTime = Date.now();
    
    // Simulate some processing time
    await new Promise(resolve => setTimeout(resolve, 100 + Math.random() * 200));
    
    const executionTime = Date.now() - startTime;
    
    const result = {
      statusCode: 200,
      body: {
        message: `Hello from ${functionName}!`,
        timestamp: new Date().toISOString(),
        input: data,
        execution_time: executionTime
      }
    };

    console.log(`   Execution completed in ${executionTime}ms`);
    console.log(`   Result:`, result.body);
    
    return result;
  }

  async getFunctions() {
    console.log(`📋 Getting all functions`);
    return Array.from(this.functions.values());
  }
}

class StorageDemo {
  constructor() {
    this.buckets = new Map();
    this.files = new Map();
  }

  async createBucket(name, options = {}) {
    console.log(`🪣 Creating storage bucket: ${name}`);
    
    const bucket = {
      id: `bucket_${Date.now()}`,
      name,
      description: options.description || '',
      is_public: options.isPublic || false,
      max_file_size: options.maxFileSize || 10 * 1024 * 1024, // 10MB
      allowed_mime_types: options.allowedMimeTypes || [],
      file_count: 0,
      total_size: 0,
      created_at: new Date().toISOString()
    };

    this.buckets.set(name, bucket);
    this.files.set(name, []);
    
    return bucket;
  }

  async uploadFromUrl(url, options) {
    console.log(`⬇️ Uploading file from URL: ${url}`);
    console.log(`   Target bucket: ${options.bucket}`);
    
    if (!this.files.has(options.bucket)) {
      throw new Error(`Bucket ${options.bucket} does not exist`);
    }

    // Simulate file upload
    const file = {
      id: `file_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      name: options.fileName || url.split('/').pop(),
      url: url,
      bucket: options.bucket,
      size: Math.floor(Math.random() * 1024 * 1024), // Random size up to 1MB
      content_type: 'application/octet-stream',
      is_public: options.isPublic || false,
      created_at: new Date().toISOString()
    };

    this.files.get(options.bucket).push(file);
    
    // Update bucket stats
    const bucket = this.buckets.get(options.bucket);
    if (bucket) {
      bucket.file_count++;
      bucket.total_size += file.size;
    }

    console.log(`   File uploaded: ${file.name} (${file.size} bytes)`);
    
    return file;
  }

  async getFiles(bucketName, options = {}) {
    console.log(`📁 Getting files from bucket: ${bucketName}`);
    
    if (!this.files.has(bucketName)) {
      return [];
    }

    let files = [...this.files.get(bucketName)];
    
    if (options.search) {
      files = files.filter(file => 
        file.name.toLowerCase().includes(options.search.toLowerCase())
      );
    }

    if (options.limit) {
      files = files.slice(0, options.limit);
    }

    return files;
  }

  async getBucketStats(bucketName) {
    console.log(`📊 Getting stats for bucket: ${bucketName}`);
    
    const bucket = this.buckets.get(bucketName);
    if (!bucket) {
      throw new Error(`Bucket ${bucketName} does not exist`);
    }

    const files = this.files.get(bucketName) || [];
    const fileTypes = {};
    
    files.forEach(file => {
      const ext = file.name.split('.').pop()?.toLowerCase() || 'unknown';
      fileTypes[ext] = (fileTypes[ext] || 0) + 1;
    });

    return {
      total_files: bucket.file_count,
      total_size: bucket.total_size,
      file_types: fileTypes
    };
  }
}

class UsersDemo {
  constructor() {
    this.users = [];
  }

  async createUser(userData) {
    console.log(`👤 Creating user: ${userData.email}`);
    
    const user = {
      id: `user_${Date.now()}`,
      name: userData.name,
      email: userData.email,
      role: userData.role || 'user',
      is_active: userData.is_active !== false,
      profile_data: userData.profile_data || {},
      preferences: userData.preferences || {},
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    };

    this.users.push(user);
    return user;
  }

  async getUsers(options = {}) {
    console.log(`👥 Getting users list`);
    
    let users = [...this.users];
    
    if (options.limit) {
      users = users.slice(0, options.limit);
    }

    return users;
  }

  async getUserStats() {
    console.log(`📊 Getting user statistics`);
    
    const now = new Date();
    const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    const thisWeek = new Date(today.getTime() - 7 * 24 * 60 * 60 * 1000);
    const thisMonth = new Date(today.getTime() - 30 * 24 * 60 * 60 * 1000);

    const stats = {
      total_users: this.users.length,
      active_users: this.users.filter(u => u.is_active).length,
      inactive_users: this.users.filter(u => !u.is_active).length,
      new_users_today: this.users.filter(u => new Date(u.created_at) >= today).length,
      new_users_this_week: this.users.filter(u => new Date(u.created_at) >= thisWeek).length,
      new_users_this_month: this.users.filter(u => new Date(u.created_at) >= thisMonth).length,
      user_growth_rate: 15.5 // Simulated growth rate
    };

    return stats;
  }
}

// Main demo function
async function runCloudBoxDemo() {
  console.log('🚀 CloudBox SDK Node.js Demo Starting...\n');
  
  try {
    // Initialize CloudBox SDK
    const config = {
      apiKey: process.env.CLOUDBOX_API_KEY || 'demo-api-key',
      projectId: process.env.CLOUDBOX_PROJECT_ID || 2,
      endpoint: 'https://api.cloudbox.dev'
    };
    
    console.log('⚙️ Initializing CloudBox SDK...');
    console.log(`   API Key: ${config.apiKey.substring(0, 8)}...`);
    console.log(`   Project: ${config.projectId}`);
    console.log(`   Endpoint: ${config.endpoint}\n`);
    
    const cloudbox = new CloudBoxSDKDemo(config);
    
    // Test connection
    const isConnected = await cloudbox.ping();
    console.log(`✅ Connection test: ${isConnected ? 'SUCCESS' : 'FAILED'}\n`);
    
    // Database operations
    console.log('=== DATABASE OPERATIONS ===');
    
    // Create a collection
    const postsCollection = await cloudbox.database.createCollection('blog_posts', {
      title: 'string',
      content: 'text',
      author: 'string',
      published: 'boolean',
      tags: 'array'
    });
    
    // Create some documents
    const posts = await cloudbox.database.batchCreate('blog_posts', [
      {
        title: 'Getting Started with CloudBox',
        content: 'CloudBox makes it easy to build backend services...',
        author: 'Jane Developer',
        published: true,
        tags: ['cloudbox', 'tutorial', 'getting-started']
      },
      {
        title: 'Advanced CloudBox Features',
        content: 'Learn about advanced features like serverless functions...',
        author: 'John Expert',
        published: true,
        tags: ['cloudbox', 'advanced', 'functions']
      },
      {
        title: 'CloudBox Best Practices',
        content: 'Follow these best practices for optimal performance...',
        author: 'Alice Architect',
        published: false,
        tags: ['cloudbox', 'best-practices', 'performance']
      }
    ]);
    
    console.log(`✅ Created ${posts.length} blog posts\n`);
    
    // Query documents
    const publishedPosts = await cloudbox.database.getDocuments('blog_posts', {
      filters: { published: true },
      orderBy: 'created_at',
      orderDirection: 'desc',
      limit: 10
    });
    
    console.log(`📖 Found ${publishedPosts.length} published posts:`);
    publishedPosts.forEach(post => {
      console.log(`   - ${post.title} by ${post.author}`);
    });
    console.log('');
    
    // Search documents
    const searchResults = await cloudbox.database.search('blog_posts', 'tutorial', { limit: 5 });
    console.log(`🔍 Search results for "tutorial": ${searchResults.length} matches\n`);
    
    // Functions operations
    console.log('=== FUNCTIONS OPERATIONS ===');
    
    // Create a function
    const emailValidator = await cloudbox.functions.createJavaScriptFunction(
      'email-validator',
      'Validates email addresses',
      `
        exports.handler = async (event, context) => {
          const { email } = event.data || {};
          
          if (!email) {
            return {
              statusCode: 400,
              body: { error: 'Email is required' }
            };
          }
          
          const emailRegex = /^[^\\s@]+@[^\\s@]+\\.[^\\s@]+$/;
          const isValid = emailRegex.test(email);
          
          return {
            statusCode: 200,
            body: {
              email,
              is_valid: isValid,
              message: isValid ? 'Email is valid' : 'Email is invalid'
            }
          };
        };
      `
    );
    
    // Execute the function
    const validationResult1 = await cloudbox.functions.execute('email-validator', {
      email: 'user@example.com'
    });
    
    const validationResult2 = await cloudbox.functions.execute('email-validator', {
      email: 'invalid-email'
    });
    
    console.log('✅ Email validation results:');
    console.log(`   ${validationResult1.body.email}: ${validationResult1.body.is_valid}`);
    console.log(`   ${validationResult2.body.email}: ${validationResult2.body.is_valid}\n`);
    
    // Storage operations
    console.log('=== STORAGE OPERATIONS ===');
    
    // Create buckets
    const userUploads = await cloudbox.storage.createBucket('user-uploads', {
      description: 'User uploaded files',
      isPublic: true,
      maxFileSize: 10 * 1024 * 1024, // 10MB
      allowedMimeTypes: ['image/jpeg', 'image/png', 'image/gif']
    });
    
    const backups = await cloudbox.storage.createBucket('backups', {
      description: 'System backup files',
      isPublic: false
    });
    
    // Upload files from URLs
    await cloudbox.storage.uploadFromUrl(
      'https://via.placeholder.com/300x200/blue/white?text=Sample+Image',
      {
        bucket: 'user-uploads',
        fileName: 'sample-image.png',
        isPublic: true
      }
    );
    
    await cloudbox.storage.uploadFromUrl(
      'https://jsonplaceholder.typicode.com/posts.json',
      {
        bucket: 'backups',
        fileName: 'posts-backup.json',
        isPublic: false
      }
    );
    
    // Get bucket statistics
    const uploadsStats = await cloudbox.storage.getBucketStats('user-uploads');
    const backupsStats = await cloudbox.storage.getBucketStats('backups');
    
    console.log('📊 Storage Statistics:');
    console.log(`   user-uploads: ${uploadsStats.total_files} files (${uploadsStats.total_size} bytes)`);
    console.log(`   backups: ${backupsStats.total_files} files (${backupsStats.total_size} bytes)\n`);
    
    // User management operations
    console.log('=== USER MANAGEMENT ===');
    
    // Create users
    const users = [
      {
        name: 'John Doe',
        email: 'john@example.com',
        role: 'admin',
        profile_data: { department: 'Engineering', location: 'Amsterdam' }
      },
      {
        name: 'Jane Smith',
        email: 'jane@example.com',
        role: 'user',
        profile_data: { department: 'Marketing', location: 'Berlin' }
      },
      {
        name: 'Bob Wilson',
        email: 'bob@example.com',
        role: 'user',
        profile_data: { department: 'Sales', location: 'London' }
      }
    ];
    
    for (const userData of users) {
      await cloudbox.users.createUser(userData);
    }
    
    // Get user statistics
    const userStats = await cloudbox.users.getUserStats();
    console.log('👥 User Statistics:');
    console.log(`   Total users: ${userStats.total_users}`);
    console.log(`   Active users: ${userStats.active_users}`);
    console.log(`   New users this month: ${userStats.new_users_this_month}`);
    console.log(`   Growth rate: ${userStats.user_growth_rate}%\n`);
    
    // List all functions
    const allFunctions = await cloudbox.functions.getFunctions();
    console.log(`⚡ Available functions: ${allFunctions.length}`);
    allFunctions.forEach(func => {
      console.log(`   - ${func.name} (${func.language})`);
    });
    console.log('');
    
    console.log('✅ CloudBox SDK Demo completed successfully!');
    console.log('\n🎉 All operations executed without errors.');
    console.log('📚 This demonstrates the power and simplicity of the CloudBox SDK.');
    
  } catch (error) {
    console.error('❌ Demo failed:', error.message);
    console.error(error.stack);
    process.exit(1);
  }
}

// Error handling
process.on('unhandledRejection', (reason, promise) => {
  console.error('❌ Unhandled Rejection at:', promise, 'reason:', reason);
  process.exit(1);
});

process.on('uncaughtException', (error) => {
  console.error('❌ Uncaught Exception:', error);
  process.exit(1);
});

// Run the demo
if (require.main === module) {
  runCloudBoxDemo().catch(console.error);
}

module.exports = { runCloudBoxDemo };