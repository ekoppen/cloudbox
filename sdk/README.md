# CloudBox SDK

[![npm version](https://badge.fury.io/js/@ekoppen%2Fcloudbox-sdk.svg)](https://badge.fury.io/js/@ekoppen%2Fcloudbox-sdk)
[![TypeScript](https://img.shields.io/badge/TypeScript-Ready-blue.svg)](https://www.typescriptlang.org/)

The official JavaScript/TypeScript SDK for CloudBox Backend-as-a-Service.

## Features

- 🔐 **Authentication** - User registration, login, and session management
- 🗄️ **Database** - Document-based database with collections and queries
- 📁 **Storage** - File upload, download, and management with buckets
- ⚡ **Functions** - Serverless function execution and management
- 💬 **Messaging** - Real-time chat and messaging with WebSocket support
- 👥 **Users** - User management and administration
- 🚀 **TypeScript** - Full TypeScript support with type definitions
- 🎯 **Event-Driven** - EventEmitter-based architecture for real-time updates
- 🛠️ **Interactive Setup** - Automated project setup with templates

## Quick Project Setup

### Interactive Setup Script
For new projects, use our interactive setup script to automatically configure CloudBox with Docker:

```bash
# Download and run the interactive setup
curl -fsSL https://raw.githubusercontent.com/ekoppen/cloudbox/main/sdk/scripts/interactive-setup.sh | bash

# Or clone the repository and run locally
git clone https://github.com/ekoppen/cloudbox.git
cd cloudbox
./sdk/scripts/interactive-setup.sh
```

The interactive setup will:
- ✅ Test CloudBox connection and credentials
- ✅ Choose from 6 project templates (Photo Portfolio, Blog, E-commerce, SaaS, Dev Portfolio, Custom)
- ✅ Create all necessary collections and storage buckets
- ✅ Generate Docker configuration (docker-compose.yml, .env, Dockerfile.example)
- ✅ Configure user permissions and application settings

### Project Templates Available

1. **Photo Portfolio** - Perfect for photography websites (albums, images, pages)
2. **Blog/CMS** - Content management with posts, categories, authors
3. **E-commerce** - Online store with products, orders, customers
4. **SaaS Application** - User subscriptions, plans, usage tracking
5. **Developer Portfolio** - Showcase projects, skills, experience
6. **Custom** - Define your own collections and structure

### Manual Setup
If you prefer manual setup, you can also use the SDK programmatically:

```javascript
import { CloudBoxClient, CloudBoxQuickSetup } from '@ekoppen/cloudbox-sdk';

const cloudbox = new CloudBoxClient({
  projectId: 'your-project-id',
  apiKey: 'your-api-key',
  endpoint: 'http://localhost:8080'
});

// Quick setup for photo portfolio
const setup = new CloudBoxQuickSetup(cloudbox);
await setup.setupPhotoPortfolio();
```

## Installation

```bash
npm install @ekoppen/cloudbox-sdk
# or
yarn add @ekoppen/cloudbox-sdk
# or
pnpm add @ekoppen/cloudbox-sdk
```

## Quick Start

```javascript
import { CloudBoxClient } from '@ekoppen/cloudbox-sdk';

// Initialize CloudBox
const cloudbox = new CloudBoxClient({
  apiKey: 'your-api-key',
  projectId: 'your-project-id',
  endpoint: 'http://localhost:8080' // optional
});

// Test connection
const isConnected = await cloudbox.testConnection();
console.log('Connected:', isConnected);

// Create collections with proper schema format
await cloudbox.collections.create('posts', [
  'title:string',
  'content:text',
  'published:boolean',
  'created_at:datetime'
]);

// Create storage buckets
await cloudbox.storage.createBucket({
  name: 'uploads',
  description: 'User uploaded files',
  public: true,
  maxSize: 10485760 // 10MB
});
```

## Authentication

### Register a new user

```javascript
const authResponse = await cloudbox.auth.register(
  'user@example.com',
  'securepassword',
  'John Doe',
  { age: 25, city: 'Amsterdam' } // optional profile data
);

console.log('User registered:', authResponse.user);
console.log('Session token:', authResponse.session.token);
```

### Login

```javascript
const authResponse = await cloudbox.auth.login({
  email: 'user@example.com',
  password: 'securepassword'
});

console.log('User logged in:', authResponse.user);
```

### Listen to auth events

```javascript
cloudbox.auth.on('login', (user) => {
  console.log('User logged in:', user);
});

cloudbox.auth.on('logout', () => {
  console.log('User logged out');
});
```

## Database

### Create and manage collections

```javascript
// Create a collection
const collection = await cloudbox.database.createCollection('posts', {
  title: 'string',
  content: 'text',
  author: 'string',
  published: 'boolean'
});

// Get all collections
const collections = await cloudbox.database.getCollections();
```

### Create and query documents

```javascript
// Create a document
const post = await cloudbox.database.createDocument('posts', {
  title: 'Hello CloudBox!',
  content: 'This is my first post using CloudBox SDK.',
  author: 'John Doe',
  published: true
});

// Get all documents
const posts = await cloudbox.database.getDocuments('posts');

// Query with filters
const publishedPosts = await cloudbox.database.query('posts', {
  published: true
}, {
  limit: 10,
  orderBy: 'created_at',
  orderDirection: 'desc'
});

// Search documents
const searchResults = await cloudbox.database.search('posts', 'CloudBox SDK');

// Update a document
const updatedPost = await cloudbox.database.updateDocument('posts', post.id, {
  title: 'Updated: Hello CloudBox!'
});
```

### Batch operations

```javascript
// Batch create documents
const posts = await cloudbox.database.batchCreate('posts', [
  { title: 'Post 1', content: 'Content 1', published: true },
  { title: 'Post 2', content: 'Content 2', published: false },
  { title: 'Post 3', content: 'Content 3', published: true }
]);

// Batch delete documents
await cloudbox.database.batchDelete('posts', ['doc1', 'doc2', 'doc3']);
```

## Storage

### Create buckets and upload files

```javascript
// Create a bucket
const bucket = await cloudbox.storage.createBucket('user-uploads', {
  description: 'User uploaded files',
  isPublic: true,
  maxFileSize: 10 * 1024 * 1024, // 10MB
  allowedMimeTypes: ['image/jpeg', 'image/png', 'image/gif']
});

// Upload a file (in browser)
const fileInput = document.querySelector('#file-input');
const file = fileInput.files[0];

const uploadedFile = await cloudbox.storage.uploadFile(file, {
  bucket: 'user-uploads',
  fileName: 'profile-photo.jpg',
  isPublic: true,
  onProgress: (progress) => {
    console.log('Upload progress:', progress);
  }
});

// Upload from URL
const fileFromUrl = await cloudbox.storage.uploadFromUrl(
  'https://example.com/image.jpg',
  {
    bucket: 'user-uploads',
    fileName: 'external-image.jpg',
    isPublic: true
  }
);
```

### Download and manage files

```javascript
// Get files in bucket
const files = await cloudbox.storage.getFiles('user-uploads', {
  limit: 20,
  search: 'profile'
});

// Download a file as blob
const fileBlob = await cloudbox.storage.downloadFile('user-uploads', fileId);

// Get signed download URL
const downloadUrl = await cloudbox.storage.getDownloadUrl('user-uploads', fileId, 3600); // expires in 1 hour

// Copy file to another bucket
const copiedFile = await cloudbox.storage.copyFile(
  'user-uploads',
  fileId,
  'backups',
  'backup-file.jpg'
);

// Get bucket statistics
const stats = await cloudbox.storage.getBucketStats('user-uploads');
console.log('Total files:', stats.total_files);
console.log('Total size:', stats.total_size);
```

## Functions

### Execute serverless functions

```javascript
// Execute a function
const result = await cloudbox.functions.execute('hello-world', {
  name: 'John',
  message: 'Hello from SDK!'
});

console.log('Function result:', result);

// Execute with GET method
const getResult = await cloudbox.functions.get('user-stats', {
  userId: '123'
});

// Execute with timeout
const timedResult = await cloudbox.functions.execute('slow-function', 
  { data: 'test' },
  { timeout: 10000 } // 10 seconds
);
```

### Manage functions (Admin API)

```javascript
// Create a new function
const newFunction = await cloudbox.functions.createFunction({
  name: 'user-validator',
  description: 'Validates user input',
  runtime: 'nodejs18',
  language: 'javascript',
  code: `
    exports.handler = async (event, context) => {
      const { email, name } = event.data;
      
      if (!email || !email.includes('@')) {
        return {
          statusCode: 400,
          body: { error: 'Invalid email' }
        };
      }
      
      return {
        statusCode: 200,
        body: { message: 'Valid input', email, name }
      };
    };
  `,
  is_public: true
});

// Deploy the function
await cloudbox.functions.deployFunction(newFunction.id);

// Get function execution logs
const logs = await cloudbox.functions.getFunctionLogs(newFunction.id, 50);

// Get function metrics
const metrics = await cloudbox.functions.getFunctionMetrics(newFunction.id, '7d');
console.log('Success rate:', metrics.success_rate);
```

### Create functions with templates

```javascript
// Create JavaScript function
const jsFunction = await cloudbox.functions.createJavaScriptFunction(
  'my-js-function',
  'My JavaScript function'
);

// Create Python function
const pyFunction = await cloudbox.functions.createPythonFunction(
  'my-py-function',
  'My Python function'
);

// Create Go function
const goFunction = await cloudbox.functions.createGoFunction(
  'my-go-function',
  'My Go function'
);
```

## Messaging

### Real-time messaging

```javascript
// Connect to realtime messaging
await cloudbox.messaging.connectRealtime();

// Create a channel
const channel = await cloudbox.messaging.createChannel('general', {
  description: 'General discussion',
  type: 'public'
});

// Join a channel
await cloudbox.messaging.joinChannel(channel.id);

// Send a message
const message = await cloudbox.messaging.sendMessage(channel.id, 'Hello everyone!', {
  messageType: 'text'
});

// Subscribe to channel messages
const subscription = cloudbox.messaging.subscribeToChannel(channel.id, (data) => {
  console.log('New message:', data.payload);
});

// Listen to connection events
cloudbox.messaging.on('connected', () => {
  console.log('Connected to realtime messaging');
});

cloudbox.messaging.on('disconnected', () => {
  console.log('Disconnected from realtime messaging');
});

// Set user presence
await cloudbox.messaging.setPresence('online', { status: 'Working' });

// Unsubscribe
subscription.unsubscribe();

// Disconnect
cloudbox.messaging.disconnectRealtime();
```

### Message management

```javascript
// Get channel messages
const messages = await cloudbox.messaging.getMessages(channel.id, {
  limit: 50,
  after: '2023-12-01T00:00:00Z'
});

// Search messages
const searchResults = await cloudbox.messaging.searchMessages(channel.id, 'important');

// Update a message
const updatedMessage = await cloudbox.messaging.updateMessage(
  channel.id,
  message.id,
  'Updated message content'
);

// Mark message as read
await cloudbox.messaging.markAsRead(channel.id, message.id);
```

## User Management

### Admin user operations

```javascript
// Get all users
const users = await cloudbox.users.getUsers({
  limit: 50,
  orderBy: 'created_at',
  orderDirection: 'desc'
});

// Search users
const foundUsers = await cloudbox.users.searchUsers('john@example.com');

// Get user statistics
const userStats = await cloudbox.users.getUserStats();
console.log('Total users:', userStats.total_users);
console.log('Active users:', userStats.active_users);

// Create a user
const newUser = await cloudbox.users.createUser({
  name: 'Jane Doe',
  email: 'jane@example.com',
  password: 'securepassword',
  role: 'user'
});

// Update user
const updatedUser = await cloudbox.users.updateUser(newUser.id, {
  is_active: false
});

// Get user activity
const activity = await cloudbox.users.getUserActivity(newUser.id, {
  limit: 20,
  startDate: '2023-12-01T00:00:00Z'
});
```

## Error Handling

```javascript
import { CloudBoxError } from '@ekoppen/cloudbox-sdk';

try {
  const result = await cloudbox.functions.execute('non-existent-function');
} catch (error) {
  if (error instanceof CloudBoxError) {
    console.log('CloudBox Error:', error.message);
    console.log('Status code:', error.statusCode);
    console.log('Error code:', error.code);
  } else {
    console.log('Network or other error:', error.message);
  }
}
```

## Events

CloudBox SDK is built on EventEmitter and provides various events:

```javascript
// Auth events
cloudbox.on('auth:login', (user) => console.log('User logged in:', user));
cloudbox.on('auth:logout', () => console.log('User logged out'));
cloudbox.on('auth:session_expired', () => console.log('Session expired'));

// Messaging events
cloudbox.messaging.on('connected', () => console.log('Messaging connected'));
cloudbox.messaging.on('disconnected', (event) => console.log('Messaging disconnected:', event));
cloudbox.messaging.on('error', (error) => console.log('Messaging error:', error));
cloudbox.messaging.on('reconnect_failed', () => console.log('Failed to reconnect'));
```

## TypeScript Support

The SDK is written in TypeScript and provides comprehensive type definitions:

```typescript
import { CloudBox, User, Document, CloudFunction } from '@ekoppen/cloudbox-sdk';

const cloudbox = new CloudBox({
  apiKey: process.env.CLOUDBOX_API_KEY!,
  projectId: process.env.CLOUDBOX_PROJECT_ID!
});

// All methods are fully typed
const user: User = await cloudbox.auth.getCurrentUser();
const posts: Document[] = await cloudbox.database.getDocuments('posts');
const func: CloudFunction = await cloudbox.functions.createJavaScriptFunction('test-func');
```

## Configuration

```javascript
const cloudbox = new CloudBox({
  apiKey: 'your-api-key',
  projectId: 'your-project-id', // or projectSlug
  endpoint: 'https://api.cloudbox.dev', // optional, defaults to official API
  timeout: 30000, // optional, request timeout in milliseconds
});
```

## Browser vs Node.js

The SDK works in both browser and Node.js environments:

```javascript
// Browser (with bundler like Webpack, Vite, etc.)
import { CloudBox } from '@ekoppen/cloudbox-sdk';

// Node.js (ESM)
import { CloudBox } from '@ekoppen/cloudbox-sdk';

// Node.js (CommonJS)
const { CloudBox } = require('@ekoppen/cloudbox-sdk');
```

## Examples

Check out the `/examples` directory for more comprehensive examples:

- [React Todo App](./examples/react-todo)
- [Vue Chat App](./examples/vue-chat) 
- [Node.js API Server](./examples/node-api)
- [Serverless Functions](./examples/functions)

## API Reference

For detailed API documentation, visit [docs.cloudbox.dev/sdk](https://docs.cloudbox.dev/sdk).

## Support

- 📚 [Documentation](https://docs.cloudbox.dev)
- 💬 [Discord Community](https://discord.gg/cloudbox)
- 🐛 [Report Issues](https://github.com/cloudbox/sdk-js/issues)
- 📧 [Email Support](mailto:support@cloudbox.dev)

## License

MIT © CloudBox