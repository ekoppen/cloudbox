# CloudBox API Patterns Guide

## Overview
CloudBox provides two types of APIs that developers need to understand to avoid common pitfalls.

## API Architecture

### 1. Portfolio API (Read-Only)
**Purpose**: Optimized read endpoints for frontend display with filters applied.

**Pattern**: `/p/{project_id}/api/{resource}`

**Available Resources**:
- `GET /p/{project_id}/api/pages` - Filtered pages (published, language-specific)
- `GET /p/{project_id}/api/albums` - Filtered albums 
- `GET /p/{project_id}/api/images` - Filtered images
- `GET /p/{project_id}/api/settings` - Application settings (read/write)
- `GET /p/{project_id}/api/branding` - Branding configuration (read/write)

**Limitations**:
- ❌ NO CREATE operations (`POST`)
- ❌ NO UPDATE operations (`PUT`) 
- ❌ NO DELETE operations (`DELETE`)
- ✅ Only GET operations with business logic filters

### 2. Documents API (Full CRUD)
**Purpose**: Direct database access for admin operations.

**Pattern**: `/p/{project_id}/api/documents/{collection}`

**Available Operations**:
- `GET /p/{project_id}/api/documents/{collection}` - List all documents
- `GET /p/{project_id}/api/documents/{collection}/{id}` - Get single document
- `POST /p/{project_id}/api/documents/{collection}` - Create new document
- `PUT /p/{project_id}/api/documents/{collection}/{id}` - Update document
- `DELETE /p/{project_id}/api/documents/{collection}/{id}` - Delete document

**Available Collections**:
- `pages` - Website pages and content
- `albums` - Photo albums
- `images` - Image metadata
- `settings` - Application settings (also available via Portfolio API)
- `branding` - Branding assets (also available via Portfolio API)
- `translations` - Page translations

## Common Developer Mistakes

### ❌ Wrong Pattern
```javascript
// This will fail with 404 for CREATE/UPDATE/DELETE
await fetch('/p/2/api/pages', {
  method: 'POST',
  body: JSON.stringify(pageData)
});
```

### ✅ Correct Pattern  
```javascript
// Use documents API for CRUD operations
await fetch('/p/2/api/documents/pages', {
  method: 'POST',
  body: JSON.stringify(pageData)
});
```

## CloudBox SDK Implementation

### Pages
```javascript
portfolio: {
  // Portfolio API (filtered, read-only)
  getPages: () => this.request('/pages'),
  getPage: (id) => this.request(`/pages/${id}`),
  
  // Documents API (full CRUD)
  createPage: (page) => this.request('/documents/pages', {
    method: 'POST',
    body: JSON.stringify(this.transformPageForDocument(page))
  }),
  updatePage: (id, page) => this.request(`/documents/pages/${id}`, {
    method: 'PUT', 
    body: JSON.stringify(this.transformPageForDocument(page))
  }),
  deletePage: (id) => this.request(`/documents/pages/${id}`, {
    method: 'DELETE'
  })
}
```

### Albums
```javascript
portfolio: {
  // Portfolio API (filtered, read-only)
  getAlbums: () => this.request('/albums'),
  getAlbum: (id) => this.request(`/albums/${id}`),
  
  // Documents API (full CRUD)
  createAlbum: (album) => this.request('/documents/albums', {
    method: 'POST',
    body: JSON.stringify({
      title: album.title,
      description: album.description || '',
      status: album.status || 'active',
      images: JSON.stringify(album.images || []),
      cover_image_id: album.cover_image_id || null
    })
  }),
  updateAlbum: (id, album) => this.request(`/documents/albums/${id}`, {
    method: 'PUT',
    body: JSON.stringify({
      title: album.title,
      description: album.description || '',
      status: album.status || 'active', 
      images: JSON.stringify(album.images || []),
      cover_image_id: album.cover_image_id || null
    })
  }),
  deleteAlbum: (id) => this.request(`/documents/albums/${id}`, {
    method: 'DELETE'
  })
}
```

### Images
```javascript
portfolio: {
  // Portfolio API (filtered, read-only)
  getImages: () => this.request('/images'),
  getImage: (id) => this.request(`/images/${id}`),
  
  // Documents API (full CRUD)
  createImage: (image) => this.request('/documents/images', {
    method: 'POST',
    body: JSON.stringify({
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
    })
  }),
  updateImage: (id, image) => this.request(`/documents/images/${id}`, {
    method: 'PUT',
    body: JSON.stringify({
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
    })
  }),
  deleteImage: (id) => this.request(`/documents/images/${id}`, {
    method: 'DELETE'
  })
}
```

## Data Transformation Rules

### Page Data Mapping
```javascript
// Frontend page object → CloudBox document
{
  title: page.title,
  slug: page.path?.replace('/', '') || generateSlug(page.title),
  content: JSON.stringify(page.content || {}),
  type: page.type || 'content',
  status: page.active ? 'active' : 'inactive',
  language: 'en',
  published: 'true',
  metadata: JSON.stringify(page.metadata || {}),
  is_menu_placeholder: page.isMenuPlaceholder || false,
  is_home_page: page.isHomePage || false,
  parent_page_id: page.parentPageId || null,
  seo: page.seo ? JSON.stringify(page.seo) : null
}
```

## Best Practices

### 1. Use Appropriate API
- **Frontend display**: Use Portfolio API (`/api/pages`)
- **Admin operations**: Use Documents API (`/api/documents/pages`)

### 2. Error Handling
```javascript
try {
  const result = await cloudbox.portfolio.createPage(pageData);
  console.log('Page created:', result);
} catch (error) {
  if (error.message.includes('404')) {
    console.error('Wrong API endpoint used. Use Documents API for CRUD operations.');
  }
  throw error;
}
```

### 3. Environment Configuration
```env
# Development
VITE_CLOUDBOX_ENDPOINT=http://localhost:8080
VITE_CLOUDBOX_API_KEY=your-api-key
VITE_CLOUDBOX_PROJECT_ID=2

# Production  
VITE_CLOUDBOX_ENDPOINT=https://your-cloudbox.com
VITE_CLOUDBOX_API_KEY=your-production-api-key
VITE_CLOUDBOX_PROJECT_ID=your-project-id
```

## Testing API Endpoints

### Check Available Collections
```bash
curl -H "X-API-Key: your-api-key" \
  http://localhost:8080/p/2/api/collections
```

### Test Document Creation
```bash
curl -X POST \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{"title":"Test","status":"active"}' \
  http://localhost:8080/p/2/api/documents/pages
```

## Migration Guide

If you have existing code using the wrong endpoints:

### 1. Identify Problem Areas
Search your codebase for:
- `POST /api/pages` 
- `PUT /api/pages`
- `DELETE /api/pages`
- `POST /api/albums`
- `PUT /api/albums` 
- `DELETE /api/albums`
- `POST /api/images`
- `PUT /api/images`
- `DELETE /api/images`

### 2. Replace with Documents API
- Replace `/api/pages` with `/api/documents/pages`
- Replace `/api/albums` with `/api/documents/albums` 
- Replace `/api/images` with `/api/documents/images`

### 3. Update Data Transformation
Ensure your data matches the expected document schema for each collection.

## Support

If you encounter issues:
1. Check CloudBox logs: `docker logs cloudbox-backend`
2. Verify API key permissions
3. Confirm collection exists: `GET /p/{id}/api/collections`
4. Test with curl before implementing in code