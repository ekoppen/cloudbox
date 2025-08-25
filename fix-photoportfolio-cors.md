# PhotoPortfolio CORS Fix Implementation

## Problem Analysis

The PhotoPortfolio app was trying to access CloudBox's **global** authentication endpoint `/api/v1/auth/login`, which uses global CORS settings. However, project-specific applications should use project-specific endpoints that leverage per-project CORS configurations.

## Fixes Applied

### 1. Global CORS Configuration Fix

**File**: `/Users/eelko/Documents/_dev/_WEBSITES/cloudbox/.env`

**Change**: Updated `CORS_ORIGINS` to include PhotoPortfolio's origin:
```bash
# Before
CORS_ORIGINS=http://localhost:3000

# After  
CORS_ORIGINS=http://localhost:3000,http://localhost:3123
```

This allows PhotoPortfolio to access global CloudBox endpoints like `/api/v1/auth/login`.

### 2. Project-Specific Endpoint Usage (Recommended)

PhotoPortfolio should use the **project-specific** authentication endpoint instead:

**Wrong Endpoint** (Global):
```
POST http://localhost:8080/api/v1/auth/login
```

**Correct Endpoint** (Project-specific):
```
POST http://localhost:8080/p/{project_slug}/api/users/login
```

For PhotoPortfolio specifically:
```
POST http://localhost:8080/p/photoportfolio/api/users/login
```

### 3. CORS Middleware Implementation Analysis

CloudBox has two CORS middleware layers:

1. **Global CORS** (`middleware.CORS(cfg)`) - Used for admin/management endpoints
   - Applied to `/api/v1/*` routes
   - Uses environment variable `CORS_ORIGINS`
   - Default: `["*"]` (allow all) or from `.env`

2. **Project CORS** (`middleware.ProjectCORS(cfg, db)`) - Used for project data endpoints  
   - Applied to `/p/{project_slug}/api/*` routes
   - Reads from `cors_configs` database table per project
   - Falls back to global CORS if no project config found

### 4. Project-Specific CORS Configuration

Each project has its own CORS configuration stored in the `cors_configs` table:

```sql
-- Default CORS config created for each project
INSERT INTO cors_configs (
    project_id, 
    allowed_origins, 
    allowed_methods, 
    allowed_headers, 
    allow_credentials, 
    max_age
) VALUES (
    project_id,
    ARRAY['*'],
    ARRAY['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'OPTIONS'],
    ARRAY['*'],
    false,
    3600
);
```

## Implementation Details

### CORS Middleware Logic

1. **Global CORS** (`/api/v1/*` endpoints):
   - Checks `cfg.AllowedOrigins` from environment
   - Allows localhost origins in development
   - Sets standard CORS headers

2. **Project CORS** (`/p/{slug}/api/*` endpoints):
   - Gets `project_id` from context (set by `ProjectAuthOrJWT` middleware)
   - Queries `cors_configs` table for project-specific settings
   - Falls back to global CORS if no project config found
   - Allows localhost origins in development mode

### Authentication Flow Differences

**Global Auth Flow** (`/api/v1/auth/login`):
- JWT-based authentication for admin/management tasks
- Returns JWT tokens for CloudBox admin interface
- Used by CloudBox frontend for admin operations

**Project Auth Flow** (`/p/{slug}/api/users/login`):
- Project-specific user authentication
- Returns tokens for project-specific users
- Used by project applications (like PhotoPortfolio)
- API key authentication also supported

## Next Steps

### For PhotoPortfolio Application

1. **Update API Endpoint**: Change from `/api/v1/auth/login` to `/p/photoportfolio/api/users/login`

2. **Use Project-Specific Authentication**: The project endpoint is designed for applications like PhotoPortfolio

3. **API Key Authentication**: Consider using API key authentication for server-to-server communication:
   ```javascript
   fetch('http://localhost:8080/p/photoportfolio/api/data/collection', {
     headers: {
       'X-API-Key': 'your-project-api-key'
     }
   })
   ```

### For CloudBox Configuration

1. **Restart CloudBox**: After updating `.env`, restart the CloudBox backend to apply new CORS settings:
   ```bash
   cd /Users/eelko/Documents/_dev/_WEBSITES/cloudbox
   docker-compose restart backend
   ```

2. **Verify Project CORS Config**: Ensure project ID 2 has proper CORS configuration:
   ```bash
   # Access CloudBox admin interface
   # Go to Projects > Settings > API > CORS Configuration
   # Add http://localhost:3123 to allowed origins
   ```

## Testing

### 1. Test Global CORS Fix
```javascript
// This should now work due to global CORS update
fetch('http://localhost:8080/api/v1/auth/login', {
  method: 'OPTIONS',
  headers: {
    'Origin': 'http://localhost:3123'
  }
})
```

### 2. Test Project-Specific Endpoint
```javascript
// Recommended approach for PhotoPortfolio
fetch('http://localhost:8080/p/photoportfolio/api/users/login', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Origin': 'http://localhost:3123'
  },
  body: JSON.stringify({
    email: 'user@example.com',
    password: 'password'
  })
})
```

## Files Modified

1. `/Users/eelko/Documents/_dev/_WEBSITES/cloudbox/.env` - Updated CORS_ORIGINS
2. `/Users/eelko/Documents/_dev/_WEBSITES/cloudbox/fix-photoportfolio-cors.md` - This documentation

## Architecture Notes

The CloudBox CORS implementation follows a layered approach:
- **Global Layer**: Admin/management operations
- **Project Layer**: Application-specific operations with per-project configuration
- **Fallback Strategy**: Project CORS falls back to global CORS if no project config

This design allows:
- Global administrative control
- Per-project customization  
- Development-friendly defaults (localhost allowance)
- Production security (configurable origins)