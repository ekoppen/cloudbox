# Development Guide

## Architecture Overview

CloudBox follows a modern microservices architecture with clear separation of concerns:

```
cloudbox/
├── backend/          # Go API server (Gin framework)
├── frontend/         # Svelte web dashboard  
├── cli/             # CloudBox CLI tool
├── sdks/            # Client SDKs
└── docs/            # Documentation
```

## Development Environment

### Prerequisites
- Go 1.21+
- Node.js 18+
- Docker & Docker Compose
- PostgreSQL 15+
- Redis 7+ (optional)

### Backend Development

#### Project Structure
```
backend/
├── cmd/                    # Command-line tools
├── internal/
│   ├── config/            # Configuration management
│   ├── database/          # Database connection and migrations
│   ├── handlers/          # HTTP request handlers
│   ├── middleware/        # HTTP middleware
│   ├── models/           # Data models
│   ├── router/           # Route definitions
│   ├── server/           # HTTP server
│   ├── services/         # Business logic
│   └── utils/           # Utility functions
├── migrations/           # Database migrations
└── main.go              # Application entry point
```

#### Running the Backend
```bash
cd backend

# Install dependencies
go mod download

# Run with hot reload
go install github.com/cosmtrek/air@latest
air

# Run tests
go test ./...

# Build binary
go build -o cloudbox main.go
```

#### Backend Configuration
Environment variables (see `.env.example`):
- `DATABASE_URL`: PostgreSQL connection string
- `JWT_SECRET`: Secret for JWT token signing
- `PORT`: Server port (default: 8080)
- `ENVIRONMENT`: development/production

### Frontend Development

#### Project Structure
```
frontend/
├── src/
│   ├── lib/              # Reusable components
│   ├── routes/           # SvelteKit routes
│   ├── app.html          # HTML template
│   ├── app.css          # Global styles
│   └── app.d.ts         # TypeScript declarations
├── static/              # Static assets
└── package.json        # Dependencies
```

#### Running the Frontend
```bash
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build

# Run type checking
npm run check
```

#### Styling
- **CSS Framework**: TailwindCSS
- **Component System**: Custom utility classes
- **Theme**: Primary blue color scheme
- **Responsive**: Mobile-first design

### Database Development

#### Migrations
```bash
# Create new migration
touch backend/migrations/002_new_feature.sql

# Apply migrations (via Docker)
docker-compose exec backend go run cmd/migrate/main.go

# Connect to database
docker-compose exec postgres psql -U cloudbox -d cloudbox
```

#### Models
GORM is used for ORM. Model definitions are in `backend/internal/models/`:
```go
type User struct {
    ID        uint           `json:"id" gorm:"primaryKey"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
    
    Email        string `json:"email" gorm:"uniqueIndex;not null"`
    PasswordHash string `json:"-" gorm:"not null"`
    // ...
}
```

## API Development

### Adding New Endpoints

1. **Define the model** (if needed) in `models/`
2. **Create handler** in `handlers/`
3. **Add routes** in `router/router.go`
4. **Add middleware** if needed
5. **Write tests**

#### Example Handler
```go
func (h *ProjectHandler) CreateProject(c *gin.Context) {
    userID := c.GetUint("user_id")
    
    var req CreateProjectRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    project := models.Project{
        Name:   req.Name,
        UserID: userID,
    }
    
    if err := h.db.Create(&project).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
        return
    }
    
    c.JSON(http.StatusCreated, project)
}
```

### Authentication & Authorization

#### JWT Authentication
- **Access tokens**: 24-hour expiry
- **Refresh tokens**: 30-day expiry (planned)
- **Claims**: user_id, email, issued_at, expires_at

#### API Key Authentication
- **Project-scoped**: Each project has separate API keys
- **Permissions**: Configurable permissions per key
- **Rate limiting**: Per-key rate limits

#### Middleware Usage
```go
// Require JWT authentication
protected.Use(middleware.RequireAuth(cfg))

// Require project API key
projectAPI.Use(middleware.ProjectAuth(cfg, db))
```

## Frontend Development

### Component Development

#### Creating Components
```svelte
<!-- src/lib/components/Button.svelte -->
<script lang="ts">
  export let variant: 'primary' | 'secondary' = 'primary';
  export let disabled = false;
</script>

<button 
  class="btn btn-{variant}" 
  class:opacity-50={disabled}
  {disabled}
  on:click
>
  <slot />
</button>
```

#### Using Components
```svelte
<script>
  import Button from '$lib/components/Button.svelte';
</script>

<Button variant="primary" on:click={handleClick}>
  Save Project
</Button>
```

### State Management
- **Stores**: Svelte stores for global state
- **API calls**: Native fetch with error handling
- **Form handling**: SvelteKit form actions

### Routing
SvelteKit file-based routing:
```
src/routes/
├── +layout.svelte        # Root layout
├── +page.svelte         # Home page (/)
├── auth/
│   ├── login/+page.svelte    # /auth/login
│   └── register/+page.svelte # /auth/register
└── dashboard/
    ├── +layout.svelte        # Dashboard layout
    ├── +page.svelte         # /dashboard
    └── projects/
        └── [id]/+page.svelte # /dashboard/projects/123
```

## Testing

### Backend Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test ./internal/handlers -v
```

#### Test Structure
```go
func TestCreateProject(t *testing.T) {
    // Setup
    db := setupTestDB()
    handler := handlers.NewProjectHandler(db, &config.Config{})
    
    // Test
    req := httptest.NewRequest("POST", "/projects", strings.NewReader(`{"name": "test"}`))
    w := httptest.NewRecorder()
    
    handler.CreateProject(gin.Default().CreateContext(req, w))
    
    // Assert
    assert.Equal(t, 201, w.Code)
}
```

### Frontend Testing
```bash
# Unit tests (planned)
npm run test

# E2E tests (planned)
npm run test:e2e
```

## Docker Development

### Multi-stage Builds
Both frontend and backend use multi-stage Docker builds for optimization:

#### Backend Dockerfile
```dockerfile
FROM golang:1.21-alpine AS builder
# Build stage...

FROM alpine:latest
# Runtime stage...
```

#### Development vs Production
- **Development**: Hot reload with volume mounts
- **Production**: Optimized builds with health checks

### Docker Compose Services
- **postgres**: PostgreSQL database
- **redis**: Redis cache (optional)
- **backend**: Go API server
- **frontend**: Svelte application

## Performance Considerations

### Backend Performance
- **Database**: Connection pooling, proper indexes
- **Caching**: Redis for session data (planned)
- **Rate limiting**: Per-IP and per-API-key limits
- **Response times**: Target <100ms for API calls

### Frontend Performance
- **Bundle size**: Code splitting, lazy loading
- **Assets**: Optimized images, compression
- **Rendering**: SSR with SvelteKit
- **Caching**: Service worker (planned)

## Code Style & Standards

### Backend (Go)
- **Formatting**: `gofmt` and `goimports`
- **Linting**: `golangci-lint`
- **Error handling**: Explicit error checking
- **Documentation**: Package and function comments

### Frontend (TypeScript/Svelte)
- **Formatting**: Prettier (planned)
- **Linting**: ESLint (planned)
- **Types**: Strict TypeScript
- **Components**: Single-file components

### Database
- **Naming**: snake_case for tables and columns
- **Migrations**: Forward-only, versioned
- **Indexes**: Performance-critical queries only

## Debugging

### Backend Debugging
```bash
# Enable debug logs
ENVIRONMENT=development go run main.go

# Use debugger
go install github.com/go-delve/delve/cmd/dlv@latest
dlv debug main.go
```

### Frontend Debugging
- **Browser DevTools**: Network, Console, Sources
- **Svelte DevTools**: Browser extension
- **Source maps**: Enabled in development

### Database Debugging
```bash
# View logs
docker-compose logs postgres

# Query performance
EXPLAIN ANALYZE SELECT * FROM projects WHERE user_id = 1;
```

## Deployment

### Environment Setup
```bash
# Production environment
cp .env.example .env.production
# Edit with production values

# Build and deploy
docker-compose -f docker-compose.prod.yml up -d
```

### Health Checks
- **Backend**: `/health` endpoint
- **Frontend**: Root path availability
- **Database**: Connection testing

### Monitoring (Planned)
- **Logs**: Structured logging with levels
- **Metrics**: Prometheus integration
- **Alerts**: Critical error notifications