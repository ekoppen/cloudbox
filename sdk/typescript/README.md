# CloudBox SDK for TypeScript/JavaScript

[![NPM Version](https://img.shields.io/npm/v/@ekoppen/cloudbox-sdk.svg)](https://www.npmjs.com/package/@ekoppen/cloudbox-sdk)
[![TypeScript](https://img.shields.io/badge/TypeScript-Ready-blue.svg)](https://www.typescriptlang.org/)
[![License](https://img.shields.io/npm/l/@ekoppen/cloudbox-sdk.svg)](https://github.com/ekoppen/cloudbox/blob/main/sdk/typescript/LICENSE)

Official TypeScript/JavaScript SDK for CloudBox BaaS (Backend-as-a-Service) platform. Build powerful applications with collections, storage, authentication, and serverless functions.

> 📍 **Location**: This SDK is part of the [CloudBox monorepo](https://github.com/ekoppen/cloudbox) at `/sdk/typescript/`

## 🚀 Features

- **🔒 Type-Safe**: Full TypeScript support with comprehensive type definitions
- **📊 Collections**: NoSQL database with schemas, queries, and relationships  
- **📁 Storage**: File storage with buckets, public access, and metadata
- **👥 Users**: Authentication, registration, and user management
- **⚡ Functions**: Serverless function execution and management
- **🌐 Cross-Platform**: Works in Node.js, browsers, React Native, and more
- **📦 Lightweight**: Minimal dependencies, optimized bundle size
- **🛠️ Developer-Friendly**: Intuitive API with comprehensive error handling

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
  projectId: 'your-project-id',
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

- `client.collections`: Collection and document operations
- `client.storage`: File storage and bucket operations  
- `client.users`: User management and authentication
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