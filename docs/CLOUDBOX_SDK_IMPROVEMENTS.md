# CloudBox SDK Improvements

## Issue Summary
The current CloudBox SDK generates incorrect API endpoints that result in 404 errors for CREATE/UPDATE/DELETE operations.

## Root Cause
- Portfolio API endpoints (`/api/pages`, `/api/albums`, `/api/images`) only support GET operations
- CRUD operations require Documents API endpoints (`/api/documents/{collection}`)
- SDK documentation doesn't clearly distinguish between these two API types

## Required SDK Changes

### 1. Update SDK Documentation
Add clear distinction between Portfolio API (read-only) and Documents API (full CRUD).

### 2. Fix Default SDK Methods
The SDK should use the correct endpoints by default:

#### Current (Broken) SDK Pattern:
```javascript
createPage: async (page) => {
  return this.request('/pages', {  // ❌ This returns 404
    method: 'POST',
    body: JSON.stringify(page),
  });
}
```

#### Fixed SDK Pattern:
```javascript
createPage: async (page) => {
  return this.request('/documents/pages', {  // ✅ This works
    method: 'POST', 
    body: JSON.stringify(this.transformPageData(page)),
  });
}
```

### 3. Add Data Transformation Helpers
The SDK should include built-in data transformation methods:

```javascript
class CloudBoxClient {
  
  transformPageData(page) {
    return {
      title: page.title,
      slug: page.path?.replace('/', '') || this.generateSlug(page.title),
      content: JSON.stringify(page.content || {}),
      type: page.type || 'content', 
      status: page.active ? 'active' : 'inactive',
      language: page.language || 'en',
      published: page.published?.toString() || 'true',
      metadata: JSON.stringify(page.metadata || {}),
      is_menu_placeholder: page.isMenuPlaceholder || false,
      is_home_page: page.isHomePage || false,
      parent_page_id: page.parentPageId || null,
      seo: page.seo ? JSON.stringify(page.seo) : null
    };
  }

  transformAlbumData(album) {
    return {
      title: album.title,
      description: album.description || '',
      status: album.status || 'active',
      images: JSON.stringify(album.images || []),
      cover_image_id: album.cover_image_id || null
    };
  }

  transformImageData(image) {
    return {
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
    };
  }

  generateSlug(title) {
    return title.toLowerCase()
      .replace(/\s+/g, '-')
      .replace(/[^a-z0-9-]/g, '')
      .replace(/--+/g, '-')
      .replace(/^-|-$/g, '');
  }
}
```

### 4. Add Better Error Messages
```javascript
private async request(path, options = {}) {
  const response = await fetch(url, { ...options, headers });
  
  if (!response.ok) {
    if (response.status === 404 && options.method !== 'GET') {
      const suggestion = path.includes('/pages') ? '/documents/pages' : 
                        path.includes('/albums') ? '/documents/albums' :
                        path.includes('/images') ? '/documents/images' : null;
      
      if (suggestion) {
        throw new Error(
          `CloudBox API Error: ${response.status} ${response.statusText}. ` +
          `For CRUD operations, use Documents API: ${suggestion}`
        );
      }
    }
    throw new Error(`CloudBox API Error: ${response.status} ${response.statusText}`);
  }
  
  return response.json();
}
```

### 5. Add API Pattern Detection
```javascript
class CloudBoxClient {
  constructor(config) {
    this.config = config;
    this.baseUrl = config.endpoint;
    
    // Automatically detect which API pattern to use
    this.useDocumentsAPI = true; // Default to Documents API for CRUD
  }

  portfolio = {
    // Read operations use Portfolio API (optimized with filters)
    getPages: (filters = {}) => {
      return this.request('/pages', { method: 'GET' });
    },

    getPage: (id) => {
      return this.request(`/pages/${id}`, { method: 'GET' });
    },

    // Write operations automatically use Documents API
    createPage: (page) => {
      return this.request('/documents/pages', {
        method: 'POST',
        body: JSON.stringify(this.transformPageData(page))
      });
    },

    updatePage: (id, page) => {
      return this.request(`/documents/pages/${id}`, {
        method: 'PUT',
        body: JSON.stringify(this.transformPageData(page))
      });
    },

    deletePage: (id) => {
      return this.request(`/documents/pages/${id}`, {
        method: 'DELETE'
      });
    }
  }
}
```

## Implementation Plan

### Phase 1: Documentation Update
- [ ] Update CloudBox documentation to clearly explain API patterns
- [ ] Add migration guide for existing projects
- [ ] Include troubleshooting section for 404 errors

### Phase 2: SDK Update  
- [ ] Fix default endpoints in SDK methods
- [ ] Add data transformation helpers
- [ ] Improve error messages with suggestions

### Phase 3: Developer Experience
- [ ] Add TypeScript definitions for document schemas
- [ ] Create CLI tool to validate API endpoints
- [ ] Add automated tests for both API patterns

## Backward Compatibility

To maintain compatibility with existing projects:

```javascript
// Option 1: Keep both methods
portfolio = {
  // Legacy methods (might not work for all operations)
  getPages: () => this.request('/pages'),
  createPage: (page) => this.request('/pages', { method: 'POST', ... }), // May fail
  
  // New recommended methods
  getPagesFromPortfolio: () => this.request('/pages'),
  createPageDocument: (page) => this.request('/documents/pages', { method: 'POST', ... }),
}

// Option 2: Smart endpoint detection
createPage: (page) => {
  try {
    return this.request('/pages', { method: 'POST', ... });
  } catch (error) {
    if (error.status === 404) {
      console.warn('Falling back to Documents API');
      return this.request('/documents/pages', { method: 'POST', ... });
    }
    throw error;
  }
}
```

## Testing Strategy

### Automated Tests
```javascript
describe('CloudBox API Patterns', () => {
  test('Portfolio API - GET operations work', async () => {
    const pages = await cloudbox.portfolio.getPages();
    expect(Array.isArray(pages)).toBe(true);
  });

  test('Documents API - CRUD operations work', async () => {
    const page = await cloudbox.portfolio.createPage({
      title: 'Test Page',
      type: 'content'
    });
    expect(page.id).toBeDefined();

    const updated = await cloudbox.portfolio.updatePage(page.id, {
      title: 'Updated Test Page'
    });
    expect(updated.data.title).toBe('Updated Test Page');

    await cloudbox.portfolio.deletePage(page.id);
  });
});
```

## Developer Onboarding

### Quick Start Template
```javascript
import { CloudBoxClient } from '@ekoppen/cloudbox-sdk';

const cloudbox = new CloudBoxClient({
  endpoint: process.env.VITE_CLOUDBOX_ENDPOINT,
  apiKey: process.env.VITE_CLOUDBOX_API_KEY,
  projectId: process.env.VITE_CLOUDBOX_PROJECT_ID
});

// ✅ Correct patterns from day 1
const pages = await cloudbox.portfolio.getPages();      // Portfolio API
const newPage = await cloudbox.portfolio.createPage({   // Documents API
  title: 'My New Page',
  type: 'content'
});
```

This SDK update will prevent the 404 errors that developers commonly encounter and provide a better developer experience for future CloudBox projects.