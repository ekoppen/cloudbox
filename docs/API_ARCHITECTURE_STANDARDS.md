# CloudBox API Architecture Standards

This document establishes consistent, predictable API patterns for CloudBox to eliminate confusion and ensure reliable integration.

## API Endpoint Structure Standards

### 1. Consistent URL Patterns

**✅ STANDARDIZED PATTERNS:**

```
Global Admin APIs:      /api/v1/{resource}
Project Management:     /api/v1/projects/{id}/{resource}  
Project Data APIs:      /p/{project_slug}/api/{resource}
Public Project APIs:    /p/{project_slug}/api/{resource}
```

**❌ DEPRECATED PATTERNS:**
- `/api/v1/admin/projects/{id}/` (confusing mixed auth)
- Inconsistent slug vs ID usage

### 2. Authentication Standards

**JWT Bearer Token** (Admin/Management):
```http
Authorization: Bearer {jwt_token}
```
- Used for: Global admin, project management, user management
- Endpoints: `/api/v1/*`

**API Key Authentication** (Project Data):
```http
X-API-Key: {project_api_key}
```
- Used for: Project-specific data operations
- Endpoints: `/p/{project_slug}/api/*`

**❌ NO MIXED AUTHENTICATION:**
- Endpoints should use ONE authentication method
- No dual JWT/API-key middleware on same route

## Resource Naming Conventions

### Collection Schemas
CloudBox uses simplified string-array schema format:
```javascript
{
  "name": "products",
  "schema": [
    "name:string",
    "price:float", 
    "description:text",
    "published:boolean"
  ]
}
```

**Supported Types:**
- `string` - Short text (255 chars)
- `text` - Long text
- `integer` - Whole numbers
- `float` - Decimal numbers
- `boolean` - true/false values
- `datetime` - ISO 8601 timestamps
- `json` - JSON objects
- `array` - Array values

### Storage Bucket Configuration
```javascript
{
  "name": "bucket_name",
  "description": "Bucket description",
  "is_public": true,              // Note: is_public, not public
  "max_file_size": 10485760,      // bytes
  "allowed_types": ["image/*"]    // MIME types
}
```

## HTTP Response Standards

### Success Responses
```json
{
  "id": 123,
  "status": "success",
  "data": { /* response data */ },
  "message": "Operation completed successfully"
}
```

### Error Responses
```json
{
  "error": "Resource not found",
  "code": "RESOURCE_NOT_FOUND",
  "details": { /* additional error context */ }
}
```

### HTTP Status Codes
- `200 OK` - Successful GET, PUT, DELETE
- `201 Created` - Successful POST
- `400 Bad Request` - Invalid input
- `401 Unauthorized` - Missing/invalid authentication
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource doesn't exist
- `409 Conflict` - Resource already exists
- `422 Unprocessable Entity` - Validation errors
- `500 Internal Server Error` - Server errors
- `501 Not Implemented` - Placeholder endpoints

## API Key Security Standards

### Storage
- ✅ **ONLY** store bcrypt-hashed keys in database
- ❌ **NEVER** store plain text keys
- ✅ Unique index on `key_hash` field
- ✅ Show plain key only once at creation

### Validation
```go
// Secure API key validation
for _, k := range apiKeys {
    if bcrypt.CompareHashAndPassword([]byte(k.KeyHash), []byte(providedKey)) == nil {
        // Valid key found
        return &k, nil
    }
}
```

## Project Resource Access Patterns

### Admin Management (JWT Required)
```http
GET /api/v1/projects/{id}/api-keys
POST /api/v1/projects/{id}/api-keys
DELETE /api/v1/projects/{id}/api-keys/{key_id}

GET /api/v1/projects/{id}/cors
PUT /api/v1/projects/{id}/cors

GET /api/v1/projects/{id}/collections    # Admin view
```

### Project Data Access (API Key Required)  
```http
# Collections
GET /p/{project_slug}/api/collections
POST /p/{project_slug}/api/collections
DELETE /p/{project_slug}/api/collections/{name}

# Documents  
GET /p/{project_slug}/api/data/{collection}
POST /p/{project_slug}/api/data/{collection}
GET /p/{project_slug}/api/data/{collection}/{id}
PUT /p/{project_slug}/api/data/{collection}/{id}
DELETE /p/{project_slug}/api/data/{collection}/{id}

# Storage
GET /p/{project_slug}/api/storage/buckets
POST /p/{project_slug}/api/storage/buckets
GET /p/{project_slug}/api/storage/{bucket}/files
POST /p/{project_slug}/api/storage/{bucket}/files

# Users (Project-specific)
GET /p/{project_slug}/api/users
POST /p/{project_slug}/api/users
GET /p/{project_slug}/api/users/{id}
PUT /p/{project_slug}/api/users/{id}
DELETE /p/{project_slug}/api/users/{id}

# Functions
POST /p/{project_slug}/api/functions/{name}
GET /p/{project_slug}/api/functions/{name}
```

### Public Project Access (No Auth)
```http
POST /p/{project_slug}/api/users/login
```

## CORS Configuration Standards

### Global CORS (Applied to all routes)
```go
// Permissive for development
AllowOrigins: ["http://localhost:3000", "http://localhost:5173"]
AllowMethods: ["GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"]
AllowHeaders: ["Origin", "Content-Type", "Accept", "Authorization", "X-API-Key"]
```

### Project-Specific CORS (Overrides global)
```go
// Project-specific origins from database
projectCORS := project.CORSConfig
if len(projectCORS.AllowedOrigins) > 0 {
    // Use project-specific CORS
    c.Header("Access-Control-Allow-Origin", strings.Join(projectCORS.AllowedOrigins, ", "))
}
```

## Validation Standards

### Request Validation
- All request bodies must be validated with appropriate struct tags
- Use `binding:"required"` for mandatory fields
- Validate enum values (permissions, types, etc.)
- Check foreign key relationships exist

### Permission Validation
```go
validPermissions := []string{"read", "write", "delete", "admin"}
for _, perm := range req.Permissions {
    if !contains(validPermissions, perm) {
        return errors.New("invalid permission: " + perm)
    }
}
```

## Error Handling Standards

### Consistent Error Messages
```go
// Use utility functions for common errors
utils.ResponseInvalidProjectID(c)        // "Invalid project ID"
utils.ResponseProjectNotFound(c)         // "Project not found" 
utils.ResponseInvalidAPIKey(c)           // "Invalid or expired API key"
utils.ResponseInsufficientPermissions(c) // "Insufficient permissions"
```

### Database Error Handling
```go
if err := db.Create(&model).Error; err != nil {
    if strings.Contains(err.Error(), "duplicate key") {
        c.JSON(http.StatusConflict, gin.H{
            "error": "Resource already exists",
            "code": "DUPLICATE_RESOURCE"
        })
        return
    }
    c.JSON(http.StatusInternalServerError, gin.H{
        "error": "Database error",
        "code": "DATABASE_ERROR"
    })
    return
}
```

## Migration Standards

### Database Migrations
- Sequential numbering: `001_initial.sql`, `002_feature.sql`
- Include rollback instructions in comments
- Test migrations on sample data
- Document breaking changes

### API Versioning
- Current version: `v1`
- Breaking changes require new version: `v2`
- Maintain backward compatibility for 6 months minimum
- Deprecation warnings in headers

## Testing Standards

### Endpoint Testing
```javascript
// Test authentication methods
describe('API Authentication', () => {
  it('accepts valid JWT tokens', async () => {
    const response = await request(app)
      .get('/api/v1/projects')
      .set('Authorization', `Bearer ${validJWT}`)
      .expect(200);
  });
  
  it('accepts valid API keys', async () => {
    const response = await request(app)  
      .get('/p/test-project/api/collections')
      .set('X-API-Key', validAPIKey)
      .expect(200);
  });
});
```

## Documentation Requirements

### API Documentation
- OpenAPI 3.0 specification
- Interactive documentation (Swagger UI)
- Code examples in multiple languages
- Common error scenarios

### SDK Standards  
- Consistent method naming: `client.collections.create()`
- Built-in error handling and retries
- TypeScript definitions included
- Framework-specific helpers

## Performance Standards

### Response Times
- Simple queries: < 100ms
- Complex queries: < 500ms  
- File uploads: < 5s
- Background jobs: Async with status endpoint

### Rate Limiting
```http
# Global rate limits
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640995200

# Per-project limits configurable
```

## Security Requirements

### Input Sanitization
- Validate all user inputs
- Escape SQL queries (use parameterized queries)
- Sanitize file names and paths
- Validate file types and sizes

### Audit Logging
```go
auditService.Log(AuditLogAction, userID, projectID, resourceID, details)
```

### API Key Management
- Keys expire after 1 year by default
- Support key rotation without downtime
- Track key usage and last access times
- Ability to revoke keys immediately

## Deployment Considerations

### Environment Variables
```bash
# Required for all environments
CLOUDBOX_DATABASE_URL=postgres://...
CLOUDBOX_JWT_SECRET=...

# Environment-specific
CLOUDBOX_CORS_ORIGINS=http://localhost:3000,https://app.example.com
CLOUDBOX_API_RATE_LIMIT=1000
```

### Docker Integration
```yaml
# docker-compose.yml
services:
  app:
    environment:
      - CLOUDBOX_ENDPOINT=http://host.docker.internal:8080  # For Docker
```

This standardization eliminates the confusion documented in CLOUDBOX_IMPROVEMENTS.md and provides clear, consistent patterns for all CloudBox integrations.