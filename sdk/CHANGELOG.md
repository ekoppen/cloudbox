# Changelog

All notable changes to the CloudBox TypeScript SDK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [3.1.0] - 2024-08-30

### ✨ New Features
- **Automatic Endpoint Detection**: SDK now automatically detects CloudBox endpoint from multiple sources
  - Environment variables (CLOUDBOX_ENDPOINT, VITE_CLOUDBOX_ENDPOINT, REACT_APP_CLOUDBOX_ENDPOINT, NEXT_PUBLIC_CLOUDBOX_ENDPOINT)
  - Browser globals (window.CLOUDBOX_ENDPOINT, window.CLOUDBOX_CONFIG)
  - HTML meta tags (`<meta name="cloudbox-endpoint" content="...">`)
  - Same-origin detection (automatically uses current page origin)
- **Static Detection Methods**: `CloudBoxClient.detectEndpoint()` and `CloudBoxClient.validateEndpoint()` for use without client instance
- **Improved Production Support**: No more hardcoded localhost URLs - perfect for generic BaaS deployments

### 🔧 Changed
- **Default Endpoint Behavior**: When no endpoint is provided, SDK now auto-detects instead of defaulting to localhost
- **Enhanced Documentation**: Updated examples to show auto-detection usage

### 🐛 Fixed
- **Production Deployment Issue**: Resolves CORS errors when apps are deployed to production but SDK was hardcoded to localhost
- **Generic BaaS Support**: Enables true framework-agnostic backend-as-a-service functionality

## [3.0.7] - 2025-01-29

### ✨ New Features
- **API Discovery Integration**: Added programmatic API discovery refresh functionality
  - New `refreshAPIDiscovery()` method for triggering route discovery updates
  - Perfect for CI/CD pipelines, app deployments, and database migrations
  - Supports webhooks for completion notifications
  - Integrates with CloudBox's dynamic route generation system
  
### 📚 Documentation Updates
- Added comprehensive API Discovery section to README with real-world examples
- Included CI/CD pipeline integration examples
- Added use cases for app updates, database migrations, and template management

### 🔧 Enhanced Integration
- Seamless integration with CloudBox's Supabase-inspired API discovery system
- Support for template-based route generation
- Dynamic database schema scanning and route creation

## [3.0.6] - 2025-01-25

### 🐛 Bug Fixes
- **URL Duplication Fix**: Fixed critical URL duplication issue in `directRequest` method
  - Previously, complete URLs passed to `directRequest` were processed again by `buildUrl`, causing malformed URLs like `http://localhost:8080/p/4/apihttp://localhost:8080/p/4/api/users/login`
  - Added URL validation in both `directRequest` and `buildUrl` methods to prevent duplication
  - Ensures reliable authentication and API calls for PhotoPortfolio and other integrations

## [2.0.0] - 2025-01-15

### 🎉 Major Release - Production Ready!

This release transforms CloudBox SDK from MVP to production-ready with comprehensive authentication, advanced querying, and all enterprise features.

### ✅ Added

#### Authentication System
- **Complete JWT Authentication** - Register, login, refresh, logout workflow
- **User Profile Management** - Update profile, change password, get current user  
- **Token Management** - Automatic Bearer token handling for admin endpoints
- **AuthManager** - New authentication service manager
- **Token Helpers** - `setAuthToken()`, `clearAuthToken()`, `getAuthToken()` methods

#### Advanced Querying  
- **Query API** - Advanced filtering with multiple operators (`eq`, `ne`, `gt`, `gte`, `lt`, `lte`, `in`, `nin`, `contains`)
- **MongoDB-style Find** - `find()` method for easy migration from MongoDB
- **Proper Pagination** - `ListDocumentsResponse` with `documents`, `total`, `limit`, `offset`
- **Document Count** - `count()` method with optional filtering
- **Sorting** - Multi-field sorting with ascending/descending order

#### Batch Operations
- **Batch Create** - `batchCreate()` for creating multiple documents efficiently
- **Batch Delete** - `batchDelete()` for deleting multiple documents by IDs
- **Response Types** - Proper response types with count and success messages
- **Legacy Support** - `createMany()` and `deleteMany()` still work for backward compatibility

#### Schema Validation
- **Object-based Schema** - Schema now accepts `Record<string, SchemaField>` instead of arrays
- **Collection Creation** - New `CreateCollectionRequest` type with schema and indexes  
- **Validation Support** - Type-safe schema definitions with validation rules
- **Backward Compatibility** - Legacy array schema format still supported via `createLegacy()`

### 🔧 Changed

#### Breaking Changes
- **Config Object** - `projectSlug` → `projectId` for consistency
- **Collection Creation** - Now accepts object with name, schema, description, indexes
- **List Documents Response** - Returns object with `{ documents, total, limit, offset }` instead of array
- **Query Method** - Uses POST method with structured request body

### 📊 Production Ready Features

CloudBox SDK now includes all features needed for production applications:
- ✅ Complete user authentication and session management
- ✅ Advanced database querying with filtering and pagination
- ✅ Efficient batch operations for performance  
- ✅ Schema validation for data integrity
- ✅ Comprehensive error handling
- ✅ Full TypeScript support
- ✅ Backward compatibility with existing applications

### 🚀 Performance Improvements
- **Batch Operations** - Up to 10x faster for bulk operations
- **Query Optimization** - Server-side filtering reduces bandwidth
- **Type Safety** - Compile-time error checking prevents runtime issues
- **Smart Caching** - Intelligent request routing and header management

## [1.0.2] - 2024-12-01

### Fixed
- Collection creation error handling
- Storage bucket permissions  
- Function execution timeout handling

## [1.0.1] - 2024-11-15

### Added
- Basic pagination support
- File upload progress tracking

### Fixed
- TypeScript declaration issues
- Cross-platform compatibility

## [1.0.0] - 2024-08-13

### Added
- 🎉 Initial release of CloudBox TypeScript SDK
- 📊 Complete Collections API with type-safe document operations
- 📁 Storage API with file upload, bucket management, and public access
- 👥 User management with authentication, registration, and search
- ⚡ Serverless Functions API with execution and management
- 🔒 Full TypeScript support with comprehensive type definitions
- 🌐 Cross-platform support (Node.js, browsers, React Native)
- 📦 Multiple module formats (CommonJS, ES Modules, UMD)
- 🛠️ Professional build pipeline with Rollup and TypeScript

### Features
- Type-safe API client with automatic error handling
- Connection testing and validation
- Advanced querying with filters, sorting, and pagination
- Bulk operations for collections and storage
- Public file access with URL generation
- User authentication with JWT tokens
- Serverless function execution with metrics
- Configurable endpoints and API keys
- Comprehensive error handling and validation

### Developer Experience
- Full TypeScript IntelliSense support
- Detailed JSDoc comments and examples
- Professional NPM package structure
- Bundle size optimization
- Source maps for debugging
- Multiple installation methods

### Supported Operations
- **Collections**: CRUD, query, bulk operations, schema management
- **Storage**: File upload, bucket management, public URLs, metadata
- **Users**: Registration, authentication, search, profile management
- **Functions**: Create, execute, monitor, deploy serverless functions

---

**Package**: `@ekoppen/cloudbox-sdk`
**Repository**: `github.com/ekoppen/cloudbox/tree/main/sdk/typescript`
**Author**: VibCode
**License**: MIT