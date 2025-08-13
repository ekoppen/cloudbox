# CloudBox API Structure v1.0 - Definitive Guide

**Version**: 1.0  
**Date**: 2024-01-13  
**Status**: AUTHORITATIVE - This is the single source of truth for CloudBox API structure

## 🎯 Core Principles

1. **Versioning**: All admin/management APIs use `/api/v1/` prefix for future compatibility
2. **Authentication**: Clear separation between JWT (admin) and API Key (project data)
3. **Consistency**: Same patterns across all endpoints within each category
4. **RESTful**: Standard HTTP verbs (GET, POST, PUT, DELETE) with predictable behavior

## 📐 API Structure Overview

```
CloudBox API
├── /api/v1/              → Admin/Management APIs (JWT Auth)
├── /p/{slug}/api/        → Project Data APIs (API Key Auth)  
├── /public/              → Public File Access (No Auth)
└── /health               → System Health (No Auth)
```

## 🔐 Authentication Methods

| API Type | Authentication | Header | Example |
|----------|---------------|--------|---------|
| Admin/Management | JWT Token | `Authorization: Bearer {token}` | `/api/v1/projects` |
| Project Data | API Key | `X-API-Key: {key}` | `/p/{slug}/api/collections` |
| Public Files | None | - | `/public/{slug}/{bucket}/{file}` |

## 📍 Complete Endpoint Reference

### 1. Authentication & User Management

```http
# Authentication
POST   /api/v1/auth/register
POST   /api/v1/auth/login
POST   /api/v1/auth/logout
POST   /api/v1/auth/refresh
GET    /api/v1/auth/me
PUT    /api/v1/auth/change-password

# User Management (Admin)
GET    /api/v1/admin/users
POST   /api/v1/admin/users
GET    /api/v1/admin/users/{id}
PUT    /api/v1/admin/users/{id}
DELETE /api/v1/admin/users/{id}
PUT    /api/v1/admin/users/{id}/role
```

### 2. Project Management

```http
# Project CRUD
GET    /api/v1/projects                          # List all projects
POST   /api/v1/projects                          # Create project
GET    /api/v1/projects/{id}                     # Get project details
PUT    /api/v1/projects/{id}                     # Update project
DELETE /api/v1/projects/{id}                     # Delete project
GET    /api/v1/projects/{id}/stats               # Project statistics

# Project Configuration
GET    /api/v1/projects/{id}/api-keys            # List API keys
POST   /api/v1/projects/{id}/api-keys            # Create API key
DELETE /api/v1/projects/{id}/api-keys/{key_id}   # Delete API key

GET    /api/v1/projects/{id}/cors                # Get CORS config
PUT    /api/v1/projects/{id}/cors                # Update CORS config
```

### 3. Storage Management (Admin)

```http
# Bucket Management
GET    /api/v1/projects/{id}/storage/buckets
POST   /api/v1/projects/{id}/storage/buckets
GET    /api/v1/projects/{id}/storage/buckets/{bucket}
PUT    /api/v1/projects/{id}/storage/buckets/{bucket}
DELETE /api/v1/projects/{id}/storage/buckets/{bucket}

# Public Access Management
PUT    /api/v1/projects/{id}/storage/buckets/{bucket}/visibility
GET    /api/v1/projects/{id}/storage/public-buckets
GET    /api/v1/projects/{id}/storage/buckets/{bucket}/files/{file_id}/public-url

# File Management
GET    /api/v1/projects/{id}/storage/buckets/{bucket}/files
POST   /api/v1/projects/{id}/storage/buckets/{bucket}/files
GET    /api/v1/projects/{id}/storage/buckets/{bucket}/files/{file_id}
DELETE /api/v1/projects/{id}/storage/buckets/{bucket}/files/{file_id}
PUT    /api/v1/projects/{id}/storage/buckets/{bucket}/files/{file_id}/move

# Folder Management
GET    /api/v1/projects/{id}/storage/buckets/{bucket}/folders
POST   /api/v1/projects/{id}/storage/buckets/{bucket}/folders
DELETE /api/v1/projects/{id}/storage/buckets/{bucket}/folders
```

### 4. Collections Management (Admin)

```http
# Collections
GET    /api/v1/projects/{id}/collections
POST   /api/v1/projects/{id}/collections
GET    /api/v1/projects/{id}/collections/{collection}
DELETE /api/v1/projects/{id}/collections/{collection}

# Documents
GET    /api/v1/projects/{id}/collections/{collection}/documents
POST   /api/v1/projects/{id}/collections/{collection}/documents
GET    /api/v1/projects/{id}/collections/{collection}/documents/{doc_id}
PUT    /api/v1/projects/{id}/collections/{collection}/documents/{doc_id}
DELETE /api/v1/projects/{id}/collections/{collection}/documents/{doc_id}
```

### 5. Organizations

```http
GET    /api/v1/organizations
POST   /api/v1/organizations
GET    /api/v1/organizations/{id}
PUT    /api/v1/organizations/{id}
DELETE /api/v1/organizations/{id}
GET    /api/v1/organizations/{id}/projects
```

### 6. Admin Dashboard

```http
# Statistics
GET    /api/v1/admin/stats/system
GET    /api/v1/admin/stats/user-growth
GET    /api/v1/admin/stats/project-activity
GET    /api/v1/admin/stats/function-executions
GET    /api/v1/admin/stats/deployment-stats
GET    /api/v1/admin/stats/storage-stats
GET    /api/v1/admin/stats/system-health

# System Management
GET    /api/v1/admin/system/info
POST   /api/v1/admin/system/restart
POST   /api/v1/admin/system/clear-cache
POST   /api/v1/admin/system/backup
GET    /api/v1/admin/system/settings
PUT    /api/v1/admin/system/settings/{key}
```

## 🔄 Project Data API (/p/{slug}/api/)

These endpoints are accessed with API keys and are project-specific:

### Collections & Documents

```http
# Collections
GET    /p/{project_slug}/api/collections
POST   /p/{project_slug}/api/collections
GET    /p/{project_slug}/api/collections/{collection}
DELETE /p/{project_slug}/api/collections/{collection}

# Documents (Data)
GET    /p/{project_slug}/api/data/{collection}
POST   /p/{project_slug}/api/data/{collection}
GET    /p/{project_slug}/api/data/{collection}/{id}
PUT    /p/{project_slug}/api/data/{collection}/{id}
DELETE /p/{project_slug}/api/data/{collection}/{id}
```

### Storage

```http
# Buckets
GET    /p/{project_slug}/api/storage/buckets
POST   /p/{project_slug}/api/storage/buckets
GET    /p/{project_slug}/api/storage/buckets/{bucket}
PUT    /p/{project_slug}/api/storage/buckets/{bucket}
DELETE /p/{project_slug}/api/storage/buckets/{bucket}

# Files
GET    /p/{project_slug}/api/storage/{bucket}/files
POST   /p/{project_slug}/api/storage/{bucket}/files
GET    /p/{project_slug}/api/storage/{bucket}/files/{file_id}
DELETE /p/{project_slug}/api/storage/{bucket}/files/{file_id}
PUT    /p/{project_slug}/api/storage/{bucket}/files/{file_id}/move

# Public URLs (for connected apps)
GET    /p/{project_slug}/api/storage/{bucket}/files/{file_id}/public-url
POST   /p/{project_slug}/api/storage/{bucket}/files/batch-public-urls
```

### User Management (Project-level)

```http
# User Authentication
POST   /p/{project_slug}/api/users/login
POST   /p/{project_slug}/api/users/logout

# User Management
GET    /p/{project_slug}/api/users
POST   /p/{project_slug}/api/users
GET    /p/{project_slug}/api/users/{user_id}
PUT    /p/{project_slug}/api/users/{user_id}
DELETE /p/{project_slug}/api/users/{user_id}

# Auth Settings
GET    /p/{project_slug}/api/auth/settings
PUT    /p/{project_slug}/api/auth/settings
GET    /p/{project_slug}/api/auth/providers
PATCH  /p/{project_slug}/api/auth/providers/{provider_id}
```

### Functions

```http
POST   /p/{project_slug}/api/functions/{function_name}
GET    /p/{project_slug}/api/functions/{function_name}
PUT    /p/{project_slug}/api/functions/{function_name}
DELETE /p/{project_slug}/api/functions/{function_name}
```

## 🌐 Public Endpoints (No Authentication)

```http
# Public File Access
GET    /public/{project_slug}/{bucket_name}/{file_path}

# Health Check
GET    /health

# Webhooks
POST   /api/v1/deploy/webhook/{repo_id}
```

## 📝 Important Notes

### URL Parameters

- `{id}`: Numeric project ID (used in admin APIs)
- `{project_slug}`: Text project identifier (used in project data APIs)
- `{bucket}` or `{bucket_name}`: Storage bucket name
- `{collection}`: Collection name for documents
- `{file_id}`: Unique file identifier (UUID)
- `{file_path}`: File path within bucket (can include folders)

### Response Formats

All APIs return JSON responses with consistent structure:

**Success Response:**
```json
{
  "data": {...},
  "message": "Operation successful"
}
```

**Error Response:**
```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "details": {...}
}
```

### Status Codes

- `200 OK`: Successful GET/PUT request
- `201 Created`: Successful POST request creating resource
- `204 No Content`: Successful DELETE request
- `400 Bad Request`: Invalid request parameters
- `401 Unauthorized`: Missing or invalid authentication
- `403 Forbidden`: Valid auth but insufficient permissions
- `404 Not Found`: Resource doesn't exist
- `500 Internal Server Error`: Server-side error

## 🔄 Migration & Versioning

### Future Versioning

When breaking changes are needed:
1. New version will be `/api/v2/`
2. `/api/v1/` remains supported for backward compatibility
3. Deprecation notices given 6 months in advance
4. Migration guides provided

### Version Selection

Clients should explicitly specify API version:
- ✅ Good: `/api/v1/projects`
- ❌ Avoid: `/api/projects` (no version)

## 🚀 Quick Start Examples

### Admin: Create and Configure Project

```bash
# 1. Login as admin
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@example.com", "password": "secret"}'

# 2. Create project (using JWT from login)
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Authorization: Bearer {jwt_token}" \
  -H "Content-Type: application/json" \
  -d '{"name": "My App", "slug": "my-app"}'

# 3. Make storage bucket public
curl -X PUT http://localhost:8080/api/v1/projects/{id}/storage/buckets/images/visibility \
  -H "Authorization: Bearer {jwt_token}" \
  -H "Content-Type: application/json" \
  -d '{"is_public": true}'
```

### App: Use Project Data API

```bash
# 1. Get collections (using API key)
curl http://localhost:8080/p/my-app/api/collections \
  -H "X-API-Key: {api_key}"

# 2. Create document
curl -X POST http://localhost:8080/p/my-app/api/data/users \
  -H "X-API-Key: {api_key}" \
  -H "Content-Type: application/json" \
  -d '{"name": "John", "email": "john@example.com"}'

# 3. Upload file
curl -X POST http://localhost:8080/p/my-app/api/storage/images/files \
  -H "X-API-Key: {api_key}" \
  -F "file=@photo.jpg"
```

### Public: Access Public Files

```bash
# No authentication needed for public buckets
curl http://localhost:8080/public/my-app/images/photo.jpg
```

## ✅ Consistency Checklist

- [ ] All admin routes use `/api/v1/` prefix
- [ ] All project data routes use `/p/{slug}/api/` prefix
- [ ] Public files use `/public/` prefix
- [ ] JWT auth for admin, API key for project data
- [ ] Consistent REST patterns (GET, POST, PUT, DELETE)
- [ ] Numeric IDs for admin, slugs for project data
- [ ] All responses in JSON format
- [ ] Standard HTTP status codes

---

**This document is the authoritative source for CloudBox API structure v1.0**