# CORS Configuration Guide

This guide covers CloudBox's flexible CORS (Cross-Origin Resource Sharing) configuration system, designed for development and production environments.

## Overview

CloudBox provides two levels of CORS configuration:

1. **Global CORS** - Applied to all CloudBox admin/API endpoints
2. **Project-Specific CORS** - Applied to individual project endpoints (e.g., PhotoPortfolio)

The system supports:
- ✅ Wildcard localhost patterns for development
- ✅ Dynamic header configuration 
- ✅ Environment-based configuration
- ✅ Automatic PhotoPortfolio setup
- ✅ Production-ready security

## Environment Configuration

### Required Environment Variables

```bash
# Basic CORS configuration
CORS_ORIGINS=http://localhost:3000,http://localhost:*,https://localhost:*
CORS_HEADERS=*
CORS_METHODS=GET,POST,PUT,PATCH,DELETE,OPTIONS

# Optional: Alternative variable names (backwards compatibility)
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:*
ALLOWED_HEADERS=*
ALLOWED_METHODS=GET,POST,PUT,PATCH,DELETE,OPTIONS
```

### Wildcard Patterns

**Localhost Development Patterns:**
```bash
# Allow any localhost port (HTTP)
CORS_ORIGINS=http://localhost:*

# Allow any localhost port (HTTPS)  
CORS_ORIGINS=https://localhost:*

# Allow specific ports + wildcards
CORS_ORIGINS=http://localhost:3000,http://localhost:*,https://localhost:*

# IP address patterns
CORS_ORIGINS=http://127.0.0.1:*,http://[::1]:*
```

**Production Domain Patterns:**
```bash
# Specific domains
CORS_ORIGINS=https://app.example.com,https://admin.example.com

# Subdomain wildcards
CORS_ORIGINS=https://*.example.com

# Mixed development and production
CORS_ORIGINS=http://localhost:*,https://*.example.com,https://app.example.com
```

## Header Configuration

### Wildcard Headers (Recommended for Development)

```bash
CORS_HEADERS=*
```

When using `CORS_HEADERS=*`, the system automatically includes:

**Base Headers:**
- Accept, Content-Type, Content-Length, Accept-Encoding
- Authorization, X-CSRF-Token, X-API-Key
- Cache-Control, X-Requested-With

**Session Token Headers:**
- Session-Token, session-token
- X-Session-Token, x-session-token

**Project Headers:**
- X-Project-ID, X-Project-Token  
- Project-ID, Project-Token

### Explicit Headers (Production)

```bash
CORS_HEADERS=Content-Type,Authorization,X-API-Key,Session-Token
```

## PhotoPortfolio Integration

### Automatic Setup Script

Use the flexible setup script for any PhotoPortfolio deployment:

```bash
# Auto-detect PhotoPortfolio configuration
node scripts/setup-photoportfolio-cors.js

# Specify port manually
node scripts/setup-photoportfolio-cors.js --port 4041

# Specify full origin
node scripts/setup-photoportfolio-cors.js --origin "http://localhost:4041"

# Update global CORS configuration too
node scripts/setup-photoportfolio-cors.js --port 4041 --update-global

# Dry run to see what would be changed
node scripts/setup-photoportfolio-cors.js --port 4041 --dry-run
```

### Manual PhotoPortfolio Setup

1. **Update Global CORS (in .env):**
   ```bash
   CORS_ORIGINS=http://localhost:3000,http://localhost:*,https://localhost:*
   ```

2. **Configure Project-Specific CORS:**
   ```bash
   # Using the update script
   node update-project2-cors.js
   
   # Or via API call
   curl -X PUT http://localhost:8080/api/v1/projects/2/cors \
     -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "allowed_origins": ["http://localhost:3000", "http://localhost:4041"],
       "allowed_methods": ["GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"],
       "allowed_headers": ["Content-Type", "Authorization", "X-API-Key", "Session-Token"],
       "allow_credentials": true,
       "max_age": 3600
     }'
   ```

## Deployment Scenarios

### Development Environment

**Local Development with Multiple Services:**
```bash
# .env configuration
CORS_ORIGINS=http://localhost:*,https://localhost:*
CORS_HEADERS=*
CORS_METHODS=GET,POST,PUT,PATCH,DELETE,OPTIONS
```

This allows:
- CloudBox frontend on any port (http://localhost:3000, http://localhost:3001, etc.)
- PhotoPortfolio on any port (http://localhost:4041, http://localhost:3123, etc.)
- Any other localhost service

### Staging Environment

**Mixed Local and Remote Testing:**
```bash
CORS_ORIGINS=http://localhost:*,https://staging.example.com,https://*.staging.example.com
CORS_HEADERS=*
CORS_METHODS=GET,POST,PUT,PATCH,DELETE,OPTIONS
```

### Production Environment

**Secure Production Setup:**
```bash
# Specific domains only
CORS_ORIGINS=https://app.example.com,https://admin.example.com,https://portfolio.example.com

# Explicit headers for security
CORS_HEADERS=Content-Type,Authorization,X-API-Key,Session-Token,X-Project-ID

# Standard REST methods
CORS_METHODS=GET,POST,PUT,PATCH,DELETE,OPTIONS
```

**Production with Subdomain Support:**
```bash
CORS_ORIGINS=https://*.example.com,https://app.example.com
```

## API Endpoints

### Global CORS Management

CloudBox admin endpoints automatically use global CORS configuration:
- `POST /api/v1/auth/login`
- `GET /api/v1/projects`
- `PUT /api/v1/projects/{id}`
- All `/api/v1/*` endpoints

### Project-Specific CORS Management

**Get Project CORS Configuration:**
```http
GET /api/v1/projects/{id}/cors
Authorization: Bearer {jwt_token}
```

**Update Project CORS Configuration:**
```http
PUT /api/v1/projects/{id}/cors
Authorization: Bearer {jwt_token}
Content-Type: application/json

{
  "allowed_origins": ["http://localhost:3000", "http://localhost:4041"],
  "allowed_methods": ["GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"],
  "allowed_headers": ["Content-Type", "Authorization", "X-API-Key"],
  "allow_credentials": true,
  "max_age": 3600
}
```

**Project API endpoints** (like PhotoPortfolio) use project-specific CORS:
- `/p/{project_name}/api/*`
- `/p/photoportfolio/api/data/*`
- `/p/photoportfolio/api/users/*`

## Testing CORS Configuration

### Manual Testing

**Test Global CORS:**
```bash
curl -X OPTIONS http://localhost:8080/api/v1/auth/login \
  -H "Origin: http://localhost:4041" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Content-Type" \
  -v
```

**Test Project CORS:**
```bash
curl -X OPTIONS http://localhost:8080/p/photoportfolio/api/data/test \
  -H "Origin: http://localhost:4041" \
  -H "Access-Control-Request-Method: GET" \
  -H "Access-Control-Request-Headers: Content-Type,X-API-Key" \
  -v
```

### Automated Testing

Use the built-in test functions:

```bash
# Test with setup script
node scripts/setup-photoportfolio-cors.js --port 4041

# Test with update script  
node update-project2-cors.js
```

## Troubleshooting

### Common Issues

**1. CORS Blocked for localhost:XXXX**
```bash
# Solution: Add wildcard localhost pattern
CORS_ORIGINS=http://localhost:*,https://localhost:*
```

**2. "Header not allowed" errors**
```bash
# Solution: Use wildcard headers for development
CORS_HEADERS=*

# Or add specific header to explicit list
CORS_HEADERS=Content-Type,Authorization,X-API-Key,Your-Custom-Header
```

**3. Project endpoints not working**
- Ensure project-specific CORS is configured
- Check that project ID matches your setup
- Verify project exists in database

**4. Production CORS too permissive**
```bash
# Solution: Use explicit configuration
CORS_ORIGINS=https://your-domain.com,https://admin.your-domain.com
CORS_HEADERS=Content-Type,Authorization,X-API-Key
```

### Debug Information

Enable CORS debugging by checking CloudBox backend logs:
```bash
docker-compose logs backend | grep CORS
```

Look for debug messages like:
```
DEBUG: Found CORS_ORIGINS = http://localhost:3000,http://localhost:*
DEBUG: CORS origins: [http://localhost:3000 http://localhost:*]
DEBUG: Found CORS_HEADERS = *
DEBUG: CORS headers: [*]
```

## Security Considerations

### Development vs Production

**Development (Permissive):**
- Wildcard localhost origins: `http://localhost:*`
- Wildcard headers: `CORS_HEADERS=*`
- Automatic fallback to localhost detection

**Production (Restrictive):**
- Explicit domain origins: `https://app.example.com`
- Explicit required headers only
- No wildcard patterns
- HTTPS-only origins

### Best Practices

1. **Use wildcard patterns for development only**
2. **Always use HTTPS origins in production**
3. **Limit headers to what your application actually needs**
4. **Regularly audit CORS configuration**
5. **Use project-specific CORS for multi-tenant applications**
6. **Test CORS configuration after deployment**

## Migration from Hardcoded CORS

If you have existing hardcoded CORS values, migrate using this checklist:

### ✅ Before (Hardcoded)
```go
// Hardcoded in middleware
c.Header("Access-Control-Allow-Headers", "Accept, Content-Type, Authorization, Session-Token, session-token")
```

```bash
# Hardcoded in .env
CORS_ORIGINS=http://localhost:3000,http://localhost:4041
```

### ✅ After (Flexible)
```go
// Dynamic in middleware  
allowedHeaders := getDefaultAllowedHeaders(cfg)
c.Header("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))
```

```bash
# Flexible in .env
CORS_ORIGINS=http://localhost:3000,http://localhost:*,https://localhost:*
CORS_HEADERS=*
```

### Migration Steps

1. **Update .env file** with wildcard patterns
2. **Restart CloudBox backend** to load new configuration
3. **Test with PhotoPortfolio** on different ports
4. **Update project-specific CORS** if needed
5. **Verify all existing functionality** still works

This migration maintains backwards compatibility while adding flexibility for future deployments.