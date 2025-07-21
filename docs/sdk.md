# CloudBox SDKs

Official SDKs and client libraries for CloudBox Backend-as-a-Service platform.

## JavaScript/TypeScript SDK

### Installation

```bash
npm install @cloudbox/sdk
# or
yarn add @cloudbox/sdk
```

### Quick Start

```typescript
import { CloudBox } from '@cloudbox/sdk';

// Initialize with your project configuration
const cloudbox = new CloudBox({
  projectSlug: 'your-project-slug',
  apiKey: 'your-api-key',
  baseURL: 'http://localhost:8080' // Your CloudBox instance
});

// Authentication
const auth = await cloudbox.auth.login({
  email: 'user@example.com',
  password: 'password123'
});

// Store data
const user = await cloudbox.collections('users').create({
  name: 'John Doe',
  email: 'john@example.com',
  age: 30
});

// Retrieve data
const users = await cloudbox.collections('users').list();
const specificUser = await cloudbox.collections('users').get(user.id);

// Update data
const updatedUser = await cloudbox.collections('users').update(user.id, {
  age: 31
});

// Delete data
await cloudbox.collections('users').delete(user.id);
```

### Collections API

```typescript
// Create collection
const collection = cloudbox.collections('products');

// CRUD operations
const product = await collection.create({
  name: 'MacBook Pro',
  price: 2999,
  category: 'electronics'
});

// List with filters and pagination
const products = await collection.list({
  filters: { category: 'electronics' },
  limit: 10,
  offset: 0,
  orderBy: 'created_at',
  orderDirection: 'desc'
});

// Search
const searchResults = await collection.search({
  query: 'MacBook',
  fields: ['name', 'description']
});
```

### File Storage API

```typescript
// Upload file
const file = await cloudbox.storage('images').upload({
  file: fileData,
  filename: 'profile.jpg',
  contentType: 'image/jpeg',
  metadata: {
    userId: '123',
    category: 'profile'
  }
});

// List files
const files = await cloudbox.storage('images').list({
  limit: 20,
  prefix: 'profile_'
});

// Get file URL
const fileUrl = await cloudbox.storage('images').getUrl(file.id);

// Download file
const fileData = await cloudbox.storage('images').download(file.id);

// Delete file
await cloudbox.storage('images').delete(file.id);
```

### Real-time API

```typescript
// Connect to real-time
const realtime = cloudbox.realtime();

// Join channel
const channel = await realtime.channel('chat-room-1');

// Listen for messages
channel.on('message', (data) => {
  console.log('New message:', data);
});

// Send message
await channel.send({
  type: 'message',
  content: 'Hello, world!',
  userId: auth.user.id
});

// Leave channel
await channel.leave();
```

### Functions API

```typescript
// Execute function
const result = await cloudbox.functions.execute('process-payment', {
  amount: 99.99,
  currency: 'USD',
  userId: '123'
});

// Execute with timeout
const result = await cloudbox.functions.execute('slow-function', data, {
  timeout: 30000 // 30 seconds
});
```

### User Management API

```typescript
// Register user
const newUser = await cloudbox.users.create({
  email: 'newuser@example.com',
  password: 'securepassword',
  name: 'New User'
});

// Update user profile
const updatedUser = await cloudbox.users.update(userId, {
  name: 'Updated Name',
  metadata: { preferences: { theme: 'dark' } }
});

// Delete user
await cloudbox.users.delete(userId);

// List users (admin only)
const users = await cloudbox.users.list({
  limit: 50,
  filters: { is_active: true }
});
```

## Python SDK

### Installation

```bash
pip install cloudbox-sdk
```

### Quick Start

```python
from cloudbox import CloudBox

# Initialize client
cloudbox = CloudBox(
    project_slug='your-project-slug',
    api_key='your-api-key',
    base_url='http://localhost:8080'
)

# Authentication
auth = cloudbox.auth.login(
    email='user@example.com',
    password='password123'
)

# Collections
collection = cloudbox.collections('users')

# Create document
user = collection.create({
    'name': 'John Doe',
    'email': 'john@example.com',
    'age': 30
})

# List documents
users = collection.list(
    filters={'age': {'$gte': 18}},
    limit=10,
    order_by='created_at'
)

# Update document
updated_user = collection.update(user['id'], {
    'age': 31
})
```

### File Storage

```python
# Upload file
with open('image.jpg', 'rb') as f:
    file = cloudbox.storage('images').upload(
        file=f,
        filename='profile.jpg',
        content_type='image/jpeg'
    )

# Download file
file_data = cloudbox.storage('images').download(file['id'])
with open('downloaded.jpg', 'wb') as f:
    f.write(file_data)
```

## Go SDK

### Installation

```bash
go get github.com/ekoppen/cloudbox-go-sdk
```

### Quick Start

```go
package main

import (
    "github.com/ekoppen/cloudbox-go-sdk"
)

func main() {
    // Initialize client
    client := cloudbox.New(&cloudbox.Config{
        ProjectSlug: "your-project-slug",
        APIKey:     "your-api-key",
        BaseURL:    "http://localhost:8080",
    })

    // Authentication
    auth, err := client.Auth.Login(&cloudbox.LoginRequest{
        Email:    "user@example.com",
        Password: "password123",
    })
    if err != nil {
        log.Fatal(err)
    }

    // Collections
    collection := client.Collections("users")

    // Create document
    user, err := collection.Create(map[string]interface{}{
        "name":  "John Doe",
        "email": "john@example.com",
        "age":   30,
    })
    if err != nil {
        log.Fatal(err)
    }

    // List documents
    users, err := collection.List(&cloudbox.ListOptions{
        Limit:   10,
        OrderBy: "created_at",
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

## Configuration Options

### SDK Configuration

```typescript
const cloudbox = new CloudBox({
  // Required
  projectSlug: 'your-project-slug',
  apiKey: 'your-api-key',
  
  // Optional
  baseURL: 'http://localhost:8080',
  timeout: 30000, // 30 seconds
  retries: 3,
  retryDelay: 1000, // 1 second
  
  // Authentication token (if already authenticated)
  authToken: 'jwt-token-here',
  
  // Custom headers
  headers: {
    'User-Agent': 'MyApp/1.0.0'
  },
  
  // Debug mode
  debug: true
});
```

### Error Handling

```typescript
try {
  const user = await cloudbox.collections('users').get('invalid-id');
} catch (error) {
  if (error.code === 'NOT_FOUND') {
    console.log('User not found');
  } else if (error.code === 'UNAUTHORIZED') {
    console.log('Authentication required');
  } else {
    console.error('Unexpected error:', error.message);
  }
}
```

### Pagination

```typescript
// Manual pagination
let page = 1;
let hasMore = true;

while (hasMore) {
  const result = await cloudbox.collections('users').list({
    limit: 10,
    offset: (page - 1) * 10
  });
  
  console.log(`Page ${page}:`, result.data);
  
  hasMore = result.data.length === 10;
  page++;
}

// Cursor-based pagination
let cursor = null;
do {
  const result = await cloudbox.collections('users').list({
    limit: 10,
    cursor: cursor
  });
  
  console.log('Results:', result.data);
  cursor = result.nextCursor;
} while (cursor);
```

## Advanced Features

### Batch Operations

```typescript
// Batch create
const users = await cloudbox.collections('users').batchCreate([
  { name: 'User 1', email: 'user1@example.com' },
  { name: 'User 2', email: 'user2@example.com' },
  { name: 'User 3', email: 'user3@example.com' }
]);

// Batch update
const updates = await cloudbox.collections('users').batchUpdate([
  { id: 'user1', data: { status: 'active' } },
  { id: 'user2', data: { status: 'inactive' } }
]);
```

### Transactions

```typescript
const transaction = await cloudbox.transaction();

try {
  const user = await transaction.collections('users').create({
    name: 'John Doe',
    email: 'john@example.com'
  });
  
  await transaction.collections('profiles').create({
    userId: user.id,
    bio: 'Software developer'
  });
  
  await transaction.commit();
} catch (error) {
  await transaction.rollback();
  throw error;
}
```

### Webhooks

```typescript
// Register webhook
await cloudbox.webhooks.create({
  url: 'https://myapp.com/webhooks/cloudbox',
  events: ['user.created', 'user.updated', 'document.created'],
  secret: 'webhook-secret-key'
});

// Verify webhook signature (in your webhook handler)
const isValid = cloudbox.webhooks.verifySignature(
  payload,
  signature,
  'webhook-secret-key'
);
```

## Examples

Check out our [examples repository](https://github.com/ekoppen/cloudbox-examples) for complete sample applications:

- **Todo App** - React + CloudBox
- **Blog Platform** - Next.js + CloudBox  
- **Chat Application** - Vue.js + CloudBox Real-time
- **E-commerce Store** - Nuxt.js + CloudBox
- **Mobile App** - React Native + CloudBox

## Support

- 📚 [Full Documentation](https://github.com/ekoppen/cloudbox/docs)
- 🐛 [Report Issues](https://github.com/ekoppen/cloudbox/issues)
- 💬 [Community Discord](https://discord.gg/cloudbox)
- 📧 [Email Support](mailto:support@cloudbox.dev)