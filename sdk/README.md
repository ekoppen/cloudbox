# 🚀 CloudBox SDK v2.0

**Production-ready TypeScript SDK for CloudBox BaaS platform**

[![npm version](https://badge.fury.io/js/@ekoppen%2Fcloudbox-sdk.svg)](https://badge.fury.io/js/@ekoppen%2Fcloudbox-sdk)
[![TypeScript](https://img.shields.io/badge/TypeScript-Ready-blue.svg)](https://www.typescriptlang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

CloudBox SDK provides a complete, type-safe interface to the CloudBox Backend-as-a-Service platform. Build modern applications with authentication, real-time data, file storage, and serverless functions.

> 📍 **Location**: This SDK is part of the [CloudBox monorepo](https://github.com/ekoppen/cloudbox) at `/sdk/typescript/`

## 🌟 Features

- **🔐 Complete Authentication** - JWT-based user auth with registration, login, profile management
- **📊 Advanced Queries** - MongoDB-style queries with filtering, sorting, and pagination  
- **🚀 Batch Operations** - Efficient bulk create/delete operations
- **🏗️ Schema Validation** - Type-safe collection schemas with validation
- **💾 File Storage** - Bucket-based file management with public/private access
- **🔍 API Discovery** - Dynamic API route discovery with external refresh triggers
- **⚡ TypeScript First** - Full type safety with comprehensive interfaces
- **🔄 Backward Compatible** - Smooth migration from v1.x

## 📦 Installation

```bash
npm install @ekoppen/cloudbox-sdk
```

```bash
yarn add @ekoppen/cloudbox-sdk
```

```bash
pnpm add @ekoppen/cloudbox-sdk
```

## 🎯 Quick Start

```typescript
import { CloudBoxClient } from '@ekoppen/cloudbox-sdk';

// Initialize the client
const client = new CloudBoxClient({
  projectId: 2,
  apiKey: 'your-api-key',
  endpoint: 'https://api.cloudbox.dev' // Optional, defaults to localhost:8080
});

// Test connection
const connection = await client.testConnection();
if (connection.success) {
  console.log('✅ Connected to CloudBox!');
} else {
  console.error('❌ Connection failed:', connection.message);
}
```

## 📚 Core Features

### 🔐 Authentication

JWT-based authentication with complete session management. The auth manager handles user registration, login, and session management.

```typescript
// Register a new user
const { user, token } = await client.auth.register({
  email: 'user@example.com',
  password: 'securePassword123',
  name: 'John Doe'
});

// Login existing user
const { user, token } = await client.auth.login({
  email: 'user@example.com',
  password: 'securePassword123'
});

// IMPORTANT: Set token for authenticated requests
client.setAuthToken(token);

// Get current authenticated user
// Note: me() uses the token already set on the client, no parameters needed!
const currentUser = await client.auth.me();
console.log('Logged in as:', currentUser.email);

// Update profile
await client.auth.updateProfile({
  name: 'Jane Doe',
  metadata: { bio: 'Software Developer' }
});

// Change password
await client.auth.changePassword({
  current_password: 'oldPassword123',
  new_password: 'newPassword456'
});

// Logout
await client.auth.logout();
```

### 📊 Collections (Database)

Manage NoSQL collections with schema validation and powerful querying.

```typescript
// Create a collection with schema
const userCollection = await client.collections.create('users', [
  { name: 'email', type: 'string', required: true, unique: true },
  { name: 'name', type: 'string', required: true },
  { name: 'age', type: 'number' },
  { name: 'is_active', type: 'boolean' }
]);

// Create documents
const user = await client.collections.createDocument('users', {
  email: 'john@example.com',
  name: 'John Doe',
  age: 30,
  is_active: true
});

// Query with filtering and sorting
const activeUsers = await client.collections.query('users', {
  filters: [
    { field: 'is_active', operator: 'eq', value: true },
    { field: 'age', operator: 'gte', value: 18 }
  ],
  sort: [{ field: 'created_at', direction: 'desc' }],
  limit: 10
});
```

### 📁 Storage (Files)

Upload, manage, and serve files with bucket-based organization.

```typescript
// Create a storage bucket
const bucket = await client.storage.createBucket({
  name: 'user-uploads',
  description: 'User uploaded files',
  max_file_size: 10485760, // 10MB
  allowed_types: ['image/jpeg', 'image/png', 'image/webp'],
  is_public: true
});

// Upload files
const fileInput = document.getElementById('file') as HTMLInputElement;
const file = fileInput.files[0];

const uploadedFile = await client.storage.uploadFile('user-uploads', {
  file: file,
  fileName: 'profile-photo.jpg',
  metadata: { 
    category: 'profile',
    user_id: user.id 
  },
  folder: 'users/profiles'
});

// Get public URLs for files
const publicUrl = await client.storage.getPublicUrl('user-uploads', uploadedFile.id);
console.log('Public URL:', publicUrl.public_url);
```

### 👥 Users (Authentication)

Complete user management with authentication and authorization.

```typescript
// Create users
const newUser = await client.users.create({
  email: 'user@example.com',
  password: 'securePassword123',
  username: 'johndoe',
  first_name: 'John',
  last_name: 'Doe',
  metadata: { role: 'customer' }
});

// Login
const loginResult = await client.users.login({
  email: 'user@example.com',
  password: 'userPassword123'
});

// Search users
const users = await client.users.search('john', {
  limit: 10,
  searchFields: ['email', 'username', 'first_name', 'last_name']
});
```

### ⚡ Functions (Serverless)

Execute serverless functions with input data and get detailed results.

```typescript
// Create a function
const func = await client.functions.create({
  name: 'process-image',
  description: 'Resize and optimize uploaded images',
  code: `
    export async function handler(event, context) {
      const { imageUrl, targetSize } = event.data;
      // Process image logic here
      return { 
        success: true, 
        processedUrl: 'processed-image-url'
      };
    }
  `,
  runtime: 'nodejs18',
  environment_variables: {
    MAX_SIZE: '1024',
    QUALITY: '85'
  }
});

// Execute function with data
const result = await client.functions.execute(func.id, {
  imageUrl: 'https://example.com/image.jpg',
  targetSize: { width: 800, height: 600 }
});
```

### 🔍 API Discovery

CloudBox automatically discovers and generates API routes based on your database schema and installed templates. You can programmatically refresh this discovery when your app updates.

```typescript
// Basic refresh after app deployment
const result = await client.refreshAPIDiscovery({
  reason: 'App update deployed',
  source: 'PhotoPortfolio App v2.1.0'
});

console.log(`✅ Refreshed ${result.routeCount} API routes`);
console.log(`Categories: ${result.categories.join(', ')}`);

// Advanced refresh with webhook notification
await client.refreshAPIDiscovery({
  reason: 'Database migration completed',
  source: 'Migration Script',
  forceRescan: true,
  templates: ['photoportfolio', 'ecommerce'],
  webhook: 'https://myapp.com/webhooks/discovery-updated'
});

// Perfect for deployment workflows
async function deployApp() {
  // Deploy your app changes
  await deployToProduction();
  
  // Refresh CloudBox API discovery
  const discovery = await client.refreshAPIDiscovery({
    reason: 'Production deployment',
    source: `${appName} v${appVersion}`,
    forceRescan: true
  });
  
  console.log(`🚀 Deployed with ${discovery.routeCount} API routes ready`);
}
```

**Use Cases:**
- **App Updates**: Refresh routes after deploying new features
- **Database Migrations**: Update routes after schema changes  
- **Template Management**: Refresh when installing/updating templates
- **CI/CD Integration**: Automate discovery updates in deployment pipelines
- **Webhook Notifications**: Get callbacks when refresh completes

```typescript
// CI/CD Pipeline Example
await client.refreshAPIDiscovery({
  reason: `Build #${buildNumber} deployed`,
  source: `${process.env.CI_PIPELINE_SOURCE}`,
  webhook: `${process.env.WEBHOOK_URL}/api-updated`,
  forceRescan: true
});
```

## 🌐 Platform Support

- **Node.js**: ≥16.0.0
- **Browsers**: Modern browsers with ES2020 support
- **React Native**: Full support
- **TypeScript**: ≥4.0.0
- **Frameworks**: Next.js, React, Vue, Angular, Svelte

## 📖 API Reference

### CloudBoxClient

Main client class for interacting with CloudBox API.

#### Constructor Options

```typescript
interface CloudBoxConfig {
  projectId: string;        // Your project identifier
  apiKey: string;          // API key for authentication
  endpoint?: string;       // CloudBox endpoint (optional)
}
```

#### Service Managers

- `client.auth`: Authentication and session management (register, login, me, logout)
- `client.collections`: Collection and document operations
- `client.storage`: File storage and bucket operations  
- `client.users`: User management operations (admin)
- `client.functions`: Serverless function operations

## 🔗 Related Links

- **📊 Main Repository**: [github.com/ekoppen/cloudbox](https://github.com/ekoppen/cloudbox)
- **📦 NPM Package**: [@ekoppen/cloudbox-sdk](https://www.npmjs.com/package/@ekoppen/cloudbox-sdk)
- **📚 Documentation**: [CloudBox Docs](https://github.com/ekoppen/cloudbox/tree/main/docs)
- **🌍 CloudBox Platform**: Main BaaS platform with Go backend and Svelte frontend

## 🤝 Contributing

This SDK is part of the CloudBox ecosystem. To contribute:

1. Fork the [main CloudBox repository](https://github.com/ekoppen/cloudbox)
2. Create your feature branch (`git checkout -b feature/sdk-improvement`)
3. Make changes in `/sdk/typescript/`
4. Run tests: `cd sdk/typescript && npm test`
5. Commit your changes (`git commit -m 'Improve TypeScript SDK'`)
6. Push to the branch (`git push origin feature/sdk-improvement`)
7. Open a Pull Request

## 📄 License

MIT License - see [LICENSE](./LICENSE) file for details.

## 🙋‍♂️ Support

- 🐛 Issues: [GitHub Issues](https://github.com/ekoppen/cloudbox/issues)
- 💬 Discussions: [GitHub Discussions](https://github.com/ekoppen/cloudbox/discussions)
- 📧 Email: support@vibcode.com

---

**Built with ❤️ by [VibCode](https://vibcode.com)**

**Part of the CloudBox BaaS ecosystem 🚀**