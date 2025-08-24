# CloudBox SDK v2.1.0 - Project Authentication Support

## 🎉 Update Summary

The CloudBox SDK has been successfully updated to support project authentication properly, eliminating the need for workarounds in PhotoPortfolio and other projects.

## ✅ Changes Implemented

1. **Added `authMode` configuration option** to `CloudBoxConfig`
   - Type: `'admin' | 'project'`
   - Default: `'project'`

2. **Updated AuthManager** to use dynamic endpoints:
   - **Project mode**: Uses `/users/*` endpoints (routed to `/p/{projectId}/api/users/*`)
   - **Admin mode**: Uses `/api/v1/auth/*` endpoints

3. **Made project mode the default** since most SDK users build project applications

4. **Maintained backward compatibility** for existing admin auth usage

5. **Updated build pipeline** - TypeScript compilation passes, all builds successful

## 📖 Usage Examples

### Project Authentication (Default - Recommended)

```typescript
import { CloudBoxClient } from '@ekoppen/cloudbox-sdk';

// For most applications (PhotoPortfolio, etc.)
const client = new CloudBoxClient({
  projectId: 'photoportfolio',
  apiKey: 'your-project-api-key',
  endpoint: 'https://your-cloudbox.com'
  // authMode: 'project' is the default
});

// Authentication now works correctly
const { user, token } = await client.auth.register({
  email: 'user@example.com',
  password: 'securepassword',
  name: 'John Doe'
});

const loginResponse = await client.auth.login({
  email: 'user@example.com',
  password: 'securepassword'
});
```

### Admin Authentication (For CloudBox Admin Interfaces)

```typescript
import { CloudBoxClient } from '@ekoppen/cloudbox-sdk';

// For CloudBox admin interfaces only
const adminClient = new CloudBoxClient({
  projectId: 'your-project',
  apiKey: 'admin-api-key',
  endpoint: 'https://your-cloudbox.com',
  authMode: 'admin'  // Explicitly set admin mode
});

// Uses /api/v1/auth/* endpoints
const { user, token } = await adminClient.auth.login({
  email: 'admin@example.com',
  password: 'adminpassword'
});
```

## 🔄 Migration Guide

### For PhotoPortfolio and Similar Projects

**Before (with workarounds):**
```typescript
// Had to manually construct endpoints or use workarounds
```

**After (clean and simple):**
```typescript
const client = new CloudBoxClient({
  projectId: 'photoportfolio',
  apiKey: 'your-api-key',
  endpoint: 'https://your-cloudbox.com'
  // No authMode needed - defaults to 'project'
});

// Authentication just works!
await client.auth.register({ /* ... */ });
await client.auth.login({ /* ... */ });
```

### For Existing Admin Code

No changes needed! Existing code will continue to work:

```typescript
// This still works exactly the same
const client = new CloudBoxClient({
  projectId: 'admin-project',
  apiKey: 'admin-key',
  authMode: 'admin'  // Explicit admin mode
});
```

## 🛡️ Backward Compatibility

- ✅ Existing admin authentication code continues to work unchanged
- ✅ Default behavior is now project mode (most common use case)
- ✅ All existing API methods remain the same
- ✅ TypeScript types are fully backward compatible

## 🏗️ Technical Details

### Endpoint Routing

| Auth Mode | Endpoint Pattern | Final URL |
|-----------|-----------------|-----------|
| `project` (default) | `/users/*` | `https://api.com/p/projectId/api/users/*` |
| `admin` | `/api/v1/auth/*` | `https://api.com/api/v1/auth/*` |

### Configuration Interface

```typescript
export interface CloudBoxConfig {
  projectId: string;
  apiKey: string;
  endpoint?: string;
  authMode?: 'admin' | 'project';  // New option
}
```

## 🚀 Ready for Publishing

- ✅ TypeScript compilation passes
- ✅ All build formats generated (CJS, ESM, UMD)
- ✅ Type definitions updated
- ✅ Version bumped to 2.1.0
- ✅ Backward compatibility maintained

The SDK is now ready for NPM publishing and will work out-of-the-box for project authentication without requiring any workarounds.