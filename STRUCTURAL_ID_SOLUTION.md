# CloudBox Structural Solution: Project ID Support

## Problem Analysis
CloudBox currently uses `project_slug` in all API endpoints (`/p/:project_slug/api/*`), but you want to use project IDs consistently. This is a valid architectural preference because:
- IDs are immutable (slugs can change)
- IDs are more efficient for database lookups
- IDs prevent naming conflicts
- IDs are standard in most APIs

## Solution Design

### 1. Backend Changes - Support Both ID and Slug
Update middleware to accept both project ID and slug in the same parameter:

```go
// In middleware/auth.go - ProjectOnly and ProjectAuthOrJWT
projectIdentifier := c.Param("project_slug") // Keep param name for compatibility

// Try parsing as ID first
var project models.Project
if projectID, err := strconv.Atoi(projectIdentifier); err == nil {
    // It's a number, treat as ID
    db.Where("id = ?", projectID).First(&project)
} else {
    // It's not a number, treat as slug
    db.Where("slug = ?", projectIdentifier).First(&project)
}
```

### 2. SDK Changes - Use Project ID by Default
Update CloudBox SDK to use project ID in all methods:

```typescript
// In sdk/src/client.ts
class CloudBoxClient {
    constructor(config: CloudBoxConfig) {
        // Change from baseURL using slug to using ID
        this.baseURL = `${config.cloudboxUrl}/p/${config.projectId}/api`;
    }
}
```

### 3. Automatic Storage Bucket Creation
When a project is created, automatically create standard buckets:

```go
// In handlers/project.go - CreateProject
func createDefaultBuckets(db *gorm.DB, projectID uint) {
    buckets := []string{"images", "documents", "videos"}
    for _, name := range buckets {
        bucket := models.StorageBucket{
            ProjectID: projectID,
            Name: name,
            Public: true, // Make images public by default
            MaxFileSize: 10485760, // 10MB
            AllowedFileTypes: getDefaultFileTypes(name),
        }
        db.Create(&bucket)
    }
}
```

### 4. PhotoPortfolio Integration
PhotoPortfolio should use project ID directly:

```javascript
// In photoportfolio setup
const cloudbox = new CloudBoxClient({
    cloudboxUrl: 'http://localhost:8080',
    projectId: 2, // Use ID, not slug
    apiKey: 'your-api-key'
});
```

## Implementation Files

1. **backend/internal/middleware/auth.go** - Update project resolution logic
2. **backend/internal/handlers/project.go** - Add automatic bucket creation
3. **sdk/src/client.ts** - Change to use project ID
4. **sdk/src/types.ts** - Update config interface
5. **create-project-with-buckets.js** - Helper script for existing projects

## Benefits
- ✅ Backward compatible (still works with slugs)
- ✅ Forward compatible (works with IDs)
- ✅ No breaking changes for existing code
- ✅ Automatic bucket creation prevents 404 errors
- ✅ Consistent ID usage across all apps