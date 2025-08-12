# CloudBox Documentation & SDK Improvements

Based on the PhotoPortfolio integration experience, here are critical improvements needed for CloudBox documentation and SDK to help other projects.

## 1. API Endpoint Consistency Issues

### Problem
CloudBox uses inconsistent API endpoint patterns that confuse developers:
- Some endpoints use `/p/{project_id}/api/`
- Others use `/api/v1/projects/{project_id}/`
- Authentication varies between `X-API-Key` header and `Authorization: Bearer`

### Solution for Documentation
```markdown
## CloudBox API Reference

### Standard API Pattern
All project-specific endpoints follow this pattern:
- Base: `/p/{project_id}/api/`
- Authentication: `X-API-Key: {your_api_key}`

### Examples:
- Collections: `GET/POST /p/{project_id}/api/collections`
- Storage: `GET/POST /p/{project_id}/api/storage/buckets`
- Users: `GET/POST /p/{project_id}/api/users`
```

### SDK Improvement
```javascript
// cloudbox-sdk.js
class CloudBoxClient {
  constructor(projectId, apiKey, endpoint = 'http://localhost:8080') {
    this.projectId = projectId;
    this.apiKey = apiKey;
    this.baseUrl = `${endpoint}/p/${projectId}/api`;
  }

  async request(path, options = {}) {
    return fetch(`${this.baseUrl}${path}`, {
      ...options,
      headers: {
        'X-API-Key': this.apiKey,
        'Content-Type': 'application/json',
        ...options.headers
      }
    });
  }
}
```

## 2. Collection Schema Format Confusion

### Problem
Documentation doesn't clearly explain that collection schemas must be arrays of strings, not objects:

**Wrong (what developers expect):**
```json
{
  "name": "pages",
  "schema": {
    "title": {"type": "string"},
    "content": {"type": "text"}
  }
}
```

**Correct (what CloudBox requires):**
```json
{
  "name": "pages",
  "schema": [
    "title:string",
    "content:text",
    "published:boolean"
  ]
}
```

### Documentation Fix
Add a clear "Collection Schema Reference" section:

```markdown
## Collection Schema Reference

CloudBox uses a simplified schema format with field definitions as strings:

### Format
`fieldName:fieldType`

### Supported Types
- `string` - Short text (255 chars)
- `text` - Long text
- `integer` - Whole numbers
- `float` - Decimal numbers
- `boolean` - true/false
- `datetime` - Date and time
- `json` - JSON object
- `array` - Array of values

### Example
```javascript
await cloudbox.createCollection('products', [
  'name:string',
  'description:text',
  'price:float',
  'in_stock:boolean',
  'tags:array',
  'metadata:json'
]);
```

## 3. Project-Specific vs Global Resources

### Problem
Documentation doesn't clarify that:
- API keys are PROJECT-SPECIFIC, not global
- You cannot create new projects with a project API key
- Admin accounts can be project-specific or global

### Documentation Improvement
```markdown
## CloudBox Authentication Levels

### Global Admin
- Can create/manage all projects
- Uses global authentication endpoints
- Should NOT be created by application setup scripts

### Project API Key
- Specific to ONE project
- Cannot create new projects
- Can manage resources within that project only
- Used with `/p/{project_id}/api/` endpoints

### Project Admin
- Admin account within a specific project
- Created via `/p/{project_id}/api/users` with role: "admin"
- Can only manage that project's resources
```

## 4. Storage Bucket Configuration

### Problem
Field naming inconsistency: `public` vs `is_public`

### SDK Helper
```javascript
class CloudBoxStorage {
  async createBucket(name, options = {}) {
    return this.client.request('/storage/buckets', {
      method: 'POST',
      body: JSON.stringify({
        name,
        description: options.description || '',
        is_public: options.public ?? true, // Note: is_public, not public
        max_file_size: options.maxSize || 10485760,
        allowed_types: options.allowedTypes || []
      })
    });
  }
}
```

## 5. Docker Integration Best Practices

### Problem
No guidance on integrating CloudBox with Dockerized applications

### Documentation Addition
```markdown
## Docker Deployment with CloudBox

### Environment Variables
```bash
# .env for Docker
VITE_CLOUDBOX_ENDPOINT=http://host.docker.internal:8080  # For local dev
VITE_CLOUDBOX_API_KEY=your_api_key
VITE_CLOUDBOX_PROJECT_ID=1
```

### Docker Compose Example
```yaml
version: '3.8'
services:
  app:
    build: .
    ports:
      - "3000:80"
    env_file: .env
    depends_on:
      - cloudbox
  
  cloudbox:
    image: cloudbox/cloudbox:latest
    ports:
      - "8080:8080"
    environment:
      - CLOUDBOX_DATABASE_URL=postgres://...
```

## 6. Setup Script Template

### Problem
No standard setup script template for CloudBox integration

### Solution: Provide Template
```bash
#!/bin/bash
# CloudBox Application Setup Template

# Configuration
CLOUDBOX_ENDPOINT="${CLOUDBOX_ENDPOINT:-http://localhost:8080}"
CLOUDBOX_PROJECT_ID="${1:-}"
CLOUDBOX_API_KEY="${2:-}"

# Validate inputs
if [[ -z "$CLOUDBOX_PROJECT_ID" ]] || [[ -z "$CLOUDBOX_API_KEY" ]]; then
    echo "Usage: $0 <project_id> <api_key>"
    exit 1
fi

# Create collections with correct schema format
create_collection() {
    local name="$1"
    shift
    local schema='['
    for field in "$@"; do
        schema+='"'$field'",'
    done
    schema="${schema%,}]"
    
    curl -s -X POST "${CLOUDBOX_ENDPOINT}/p/${CLOUDBOX_PROJECT_ID}/api/collections" \
        -H "X-API-Key: ${CLOUDBOX_API_KEY}" \
        -H "Content-Type: application/json" \
        -d "{\"name\":\"$name\",\"schema\":$schema}"
}

# Create storage buckets
create_bucket() {
    local name="$1"
    local description="$2"
    
    curl -s -X POST "${CLOUDBOX_ENDPOINT}/p/${CLOUDBOX_PROJECT_ID}/api/storage/buckets" \
        -H "X-API-Key: ${CLOUDBOX_API_KEY}" \
        -H "Content-Type: application/json" \
        -d "{\"name\":\"$name\",\"description\":\"$description\",\"is_public\":true}"
}

# Setup your resources
create_collection "pages" "title:string" "content:text" "published:boolean"
create_bucket "images" "User uploaded images"
```

## 7. Common Pitfalls Section

Add to documentation:

```markdown
## Common Pitfalls and Solutions

### 1. "Invalid or expired token" Error
**Cause**: Using wrong endpoint pattern or authentication method
**Solution**: Use `/p/{project_id}/api/` with `X-API-Key` header

### 2. Collections Not Visible in Dashboard
**Cause**: Created with wrong project ID or API endpoint
**Solution**: Verify project ID and use correct endpoint pattern

### 3. "json: cannot unmarshal object into Go struct field .schema of type []string"
**Cause**: Using object schema format instead of array
**Solution**: Use array of strings: `["field:type", "field2:type2"]`

### 4. Cannot Create New Project with API Key
**Cause**: API keys are project-specific
**Solution**: Create project via dashboard first, then use its API key

### 5. Docker Container Can't Connect to CloudBox
**Cause**: Using localhost in container
**Solution**: Use `host.docker.internal` or CloudBox container name
```

## 8. SDK TypeScript Definitions

Create proper TypeScript definitions:

```typescript
// cloudbox.d.ts
export interface CloudBoxConfig {
  projectId: string | number;
  apiKey: string;
  endpoint?: string;
}

export interface Collection {
  name: string;
  schema: string[]; // Array of "field:type" strings
}

export interface Bucket {
  name: string;
  description?: string;
  is_public?: boolean; // Note: is_public, not public
  max_file_size?: number;
  allowed_types?: string[];
}

export class CloudBoxClient {
  constructor(config: CloudBoxConfig);
  
  createCollection(name: string, schema: string[]): Promise<Collection>;
  createBucket(config: Bucket): Promise<Bucket>;
  
  // Project-specific user creation
  createProjectAdmin(email: string, password: string): Promise<User>;
}
```

## Summary of Critical Documentation Needs

1. **Clear API endpoint pattern documentation** - One consistent pattern
2. **Schema format examples** - Array of strings, not objects
3. **Authentication scope clarity** - Project vs Global
4. **Field naming consistency** - Document exact field names
5. **Docker integration guide** - Common patterns and solutions
6. **Setup script templates** - Working examples for common languages
7. **Common errors section** - With specific solutions
8. **TypeScript/JavaScript SDK** - With proper types and examples

These improvements would have saved hours of debugging during the PhotoPortfolio integration!