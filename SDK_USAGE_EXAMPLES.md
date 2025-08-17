# 🚀 CloudBox SDK Usage Examples

Praktische voorbeelden voor alle CloudBox SDK features gebaseerd op je feedback.

## 🔧 Setup

```typescript
import { CloudBoxClient } from '@cloudbox/sdk';

const client = new CloudBoxClient({
  projectSlug: 'jouw-project-slug',
  apiKey: 'jouw-api-key',
  baseURL: 'http://localhost:8080' // of production URL
});
```

## 🔐 Authentication System

**✅ OPLOSSING**: Authentication endpoints bestaan al in CloudBox!

### User Authentication (JWT)

```typescript
// 1. Register nieuwe user
const { user, token, refreshToken } = await client.auth.register({
  email: 'user@example.com',
  password: 'securepassword',
  name: 'Jan Jansen'
});

// 2. Login bestaande user  
const { user, token, refreshToken } = await client.auth.login({
  email: 'user@example.com',
  password: 'securepassword'
});

// 3. Refresh token
const newToken = await client.auth.refresh(refreshToken);

// 4. Get current user
const currentUser = await client.auth.me();

// 5. Update profile
await client.auth.updateProfile({
  name: 'Jan van der Berg'
});

// 6. Change password
await client.auth.changePassword({
  currentPassword: 'oldpassword',
  newPassword: 'newpassword'
});

// 7. Logout
await client.auth.logout();
```

**SDK Implementation Needed**:
```typescript
// client.ts - Voeg toe aan CloudBoxClient class
class CloudBoxClient {
  public auth = {
    register: async (data: RegisterData) => {
      return this.post('/api/v1/auth/register', data);
    },
    login: async (data: LoginData) => {
      return this.post('/api/v1/auth/login', data);
    },
    refresh: async (refreshToken: string) => {
      return this.post('/api/v1/auth/refresh', { refresh_token: refreshToken });
    },
    me: async () => {
      return this.get('/api/v1/auth/me');
    },
    updateProfile: async (data: UpdateProfileData) => {
      return this.put('/api/v1/auth/me', data);
    },
    changePassword: async (data: ChangePasswordData) => {
      return this.put('/api/v1/auth/change-password', data);
    },
    logout: async () => {
      return this.post('/api/v1/auth/logout');
    }
  };
}
```

## 📊 Advanced Query API

**✅ OPLOSSING**: Query API bestaat al! Gebruik POST method.

```typescript
// Correct usage voor gefilterde queries
const results = await client.collections.query('goals', {
  filters: [
    { field: 'user_id', operator: 'eq', value: 'user123' },
    { field: 'is_active', operator: 'eq', value: true }
  ],
  sort: [
    { field: 'created_at', direction: 'DESC' }
  ],
  limit: 10,
  offset: 0
});

// MongoDB-style query (alternatief)
const results = await client.collections.find('goals', {
  filter: { 
    user_id: 'user123', 
    is_active: true 
  },
  sort: { created_at: -1 },
  limit: 10,
  skip: 0
});
```

**SDK Implementation Needed**:
```typescript
// collections.ts - Update query method
class CollectionsAPI {
  async query(collectionName: string, options: QueryOptions) {
    // POST method, niet GET!
    return this.client.post(`/data/${collectionName}/query`, options);
  }
  
  async find(collectionName: string, options: FindOptions) {
    // Converteer MongoDB-style naar CloudBox format
    const cloudboxQuery = {
      filters: this.convertMongoFilters(options.filter),
      sort: this.convertMongoSort(options.sort),
      limit: options.limit,
      offset: options.skip || 0
    };
    return this.query(collectionName, cloudboxQuery);
  }
}
```

## 📄 Pagination & Count

**✅ OPLOSSING**: Pagination bestaat al! Gebruik correct.

```typescript
// Pagination parameters
const documents = await client.collections.listDocuments('goals', {
  limit: 25,      // max 100
  offset: 50,     // voor page 3 bij limit 25
  orderBy: 'created_at DESC',
  filter: JSON.stringify({ user_id: 'user123' })
});

console.log(documents.total);     // Total count
console.log(documents.documents); // Array van documents
console.log(documents.limit);     // 25
console.log(documents.offset);    // 50

// Document count
const { count } = await client.collections.count('goals');
console.log(`Total goals: ${count}`);
```

## 🏗️ Schema Validation

**✅ GEFIXED**: Schema format now accepts objects instead of arrays.

```typescript
// Create collection met schema validation
await client.collections.create('goals', {
  schema: {
    user_id: { type: 'string', required: true },
    title: { type: 'string', required: true, maxLength: 100 },
    description: { type: 'string', maxLength: 500 },
    is_active: { type: 'boolean', default: true },
    target_date: { type: 'string', format: 'date' },
    priority: { type: 'number', min: 1, max: 5 }
  },
  indexes: ['user_id', 'created_at', 'is_active']
});

// Document creation with validation
await client.collections.createDocument('goals', {
  user_id: 'user123',
  title: 'Learn TypeScript',
  description: 'Master TypeScript for better development',
  is_active: true,
  target_date: '2025-12-31',
  priority: 3
});
```

## 🚀 Batch Operations

**✅ OPLOSSING**: Batch operations bestaan al!

```typescript
// Batch create documents
const result = await client.collections.batchCreate('goals', {
  documents: [
    { title: 'Goal 1', user_id: 'user123' },
    { title: 'Goal 2', user_id: 'user123' },
    { title: 'Goal 3', user_id: 'user123' }
  ]
});

console.log(`Created ${result.count} documents`);

// Batch delete documents
await client.collections.batchDelete('goals', {
  ids: ['goal1', 'goal2', 'goal3']
});
```

**SDK Implementation Needed**:
```typescript
class CollectionsAPI {
  async batchCreate(collectionName: string, data: BatchCreateData) {
    return this.client.post(`/data/${collectionName}/batch`, data);
  }
  
  async batchDelete(collectionName: string, data: BatchDeleteData) {
    return this.client.delete(`/data/${collectionName}/batch`, data);
  }
}
```

## 💾 Storage Advanced Features

```typescript
// Generate signed URL voor direct uploads
const signedUrl = await client.storage.getSignedUrl('images', 'profile.jpg', {
  expiresIn: 3600, // 1 hour
  action: 'upload'
});

// Upload file directly
await fetch(signedUrl, {
  method: 'PUT',
  body: fileBlob
});

// Get file metadata
const metadata = await client.storage.getMetadata('images', 'profile.jpg');
console.log(metadata.size, metadata.contentType, metadata.lastModified);

// Generate thumbnail (for images)
const thumbnailUrl = await client.storage.generateThumbnail('images', 'photo.jpg', {
  width: 150,
  height: 150,
  format: 'webp'
});
```

## 🎯 Real-time Subscriptions

**Future Feature**: Real-time updates via WebSockets

```typescript
// Subscribe to collection changes
const unsubscribe = client.collections.subscribe('goals', (event) => {
  switch(event.type) {
    case 'document.created':
      console.log('New goal:', event.document);
      break;
    case 'document.updated':
      console.log('Updated goal:', event.document);
      break;
    case 'document.deleted':
      console.log('Deleted goal:', event.documentId);
      break;
  }
});

// Subscribe to specific document
const unsubscribeDoc = client.collections.watch('goals/goal123', (document) => {
  console.log('Goal updated:', document);
});

// Cleanup
unsubscribe();
unsubscribeDoc();
```

## ⚡ CloudBox Functions

**Future Feature**: Serverless functions

```typescript
// Deploy function
await client.functions.deploy('dutchWeatherCoaching', {
  runtime: 'node18',
  code: `
    export default async function(context) {
      const { weather, userGoal } = context.data;
      
      if (weather === 'regen' && userGoal === 'fietsen') {
        return {
          advice: 'Perfect weer voor binnentraining! Probeer een spinning class.',
          motivation: 'Regen houdt echte sporters niet tegen! 💪'
        };
      }
      
      return { advice: 'Geniet van je workout!' };
    }
  `,
  environment: {
    WEATHER_API_KEY: 'your-key'
  }
});

// Execute function
const result = await client.functions.execute('dutchWeatherCoaching', {
  weather: 'regen',
  userGoal: 'fietsen'
});

console.log(result.advice); // "Perfect weer voor binnentraining! Probeer een spinning class."
```

## 🔧 Complete SDK Client Update

**Je SDK moet deze endpoints gebruiken**:

```typescript
class CloudBoxClient {
  constructor(options: CloudBoxOptions) {
    this.baseURL = options.baseURL;
    this.projectSlug = options.projectSlug;
    this.apiKey = options.apiKey;
  }

  // Correcte endpoints
  private getProjectURL(path: string) {
    return `/p/${this.projectSlug}/api${path}`;
  }

  private getAdminURL(path: string) {
    return `/api/v1${path}`;
  }

  // Collections API
  public collections = {
    list: () => this.get(this.getProjectURL('/collections')),
    create: (name: string, options?: CreateCollectionOptions) => 
      this.post(this.getProjectURL('/collections'), { name, ...options }),
    get: (name: string) => this.get(this.getProjectURL(`/collections/${name}`)),
    delete: (name: string) => this.delete(this.getProjectURL(`/collections/${name}`)),
    
    // Documents
    listDocuments: (collection: string, options?: ListOptions) => 
      this.get(this.getProjectURL(`/data/${collection}`), options),
    createDocument: (collection: string, data: any) => 
      this.post(this.getProjectURL(`/data/${collection}`), data),
    getDocument: (collection: string, id: string) => 
      this.get(this.getProjectURL(`/data/${collection}/${id}`)),
    updateDocument: (collection: string, id: string, data: any) => 
      this.put(this.getProjectURL(`/data/${collection}/${id}`), data),
    deleteDocument: (collection: string, id: string) => 
      this.delete(this.getProjectURL(`/data/${collection}/${id}`)),
    
    // Advanced operations
    query: (collection: string, options: QueryOptions) => 
      this.post(this.getProjectURL(`/data/${collection}/query`), options),
    count: (collection: string) => 
      this.get(this.getProjectURL(`/data/${collection}/count`)),
    batchCreate: (collection: string, data: BatchCreateData) => 
      this.post(this.getProjectURL(`/data/${collection}/batch`), data),
    batchDelete: (collection: string, data: BatchDeleteData) => 
      this.delete(this.getProjectURL(`/data/${collection}/batch`), data)
  };

  // Auth API (uses admin endpoints)
  public auth = {
    register: (data: RegisterData) => this.post(this.getAdminURL('/auth/register'), data),
    login: (data: LoginData) => this.post(this.getAdminURL('/auth/login'), data),
    refresh: (refreshToken: string) => this.post(this.getAdminURL('/auth/refresh'), { refresh_token: refreshToken }),
    me: () => this.get(this.getAdminURL('/auth/me')),
    updateProfile: (data: UpdateProfileData) => this.put(this.getAdminURL('/auth/me'), data),
    changePassword: (data: ChangePasswordData) => this.put(this.getAdminURL('/auth/change-password'), data),
    logout: () => this.post(this.getAdminURL('/auth/logout'))
  };
}
```

## 🎯 Test Connection Update

```typescript
// Test connection werkt al perfect!
const isConnected = await client.testConnection();
console.log('CloudBox connected:', isConnected.status); // "ok"
```

## 📈 Production Readiness Checklist

### ✅ Ready Now
- [x] Collections CRUD
- [x] Document CRUD  
- [x] Storage buckets
- [x] Pagination
- [x] Query/Filter API
- [x] Batch operations
- [x] Count operations
- [x] Schema validation (fixed)
- [x] Authentication system
- [x] TypeScript support

### 🔨 SDK Updates Needed
- [ ] Update endpoints in SDK
- [ ] Add auth methods to SDK
- [ ] Fix query method (use POST)
- [ ] Add batch operations to SDK
- [ ] Update schema format support

### 🚀 Future Features
- [ ] Real-time subscriptions
- [ ] CloudBox Functions
- [ ] File storage advanced features
- [ ] Database hooks
- [ ] Rate limiting
- [ ] Multi-region support

## 🎉 Conclusie

**CloudBox is al veel verder dan je dacht!** 🚀

De meeste features die je miste bestaan al in de backend. Het probleem zit in de SDK die:
1. Verkeerde endpoints aanroept
2. Verkeerde HTTP methods gebruikt  
3. Niet alle beschikbare features exposeert

Met deze voorbeelden en de SDK updates zou CloudBox direct production-ready moeten zijn voor jouw Nederlandse AI coaching app! 💪