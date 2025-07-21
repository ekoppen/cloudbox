# Getting Started with CloudBox

## Quick Start

### Prerequisites
- Docker and Docker Compose
- Git
- Make (optional, for convenience)

### 1. Clone the Repository
```bash
git clone https://github.com/your-org/cloudbox.git
cd cloudbox
```

### 2. Setup Environment
```bash
# Copy example environment file
cp .env.example .env

# Edit .env with your settings (optional for development)
```

### 3. Start CloudBox
```bash
# Using Make (recommended)
make quick-start

# Or using Docker Compose directly
docker-compose up -d
```

### 4. Access CloudBox
- **Frontend Dashboard**: http://localhost:3000
- **Backend API**: http://localhost:8080
- **Health Check**: http://localhost:8080/health

## Development Setup

### Backend Development
```bash
# Install Go dependencies
cd backend
go mod download

# Run with hot reload (requires air)
go install github.com/cosmtrek/air@latest
air

# Or run directly
go run main.go
```

### Frontend Development
```bash
# Install dependencies
cd frontend
npm install

# Start development server
npm run dev
```

### Database Access
```bash
# Connect to PostgreSQL
docker-compose exec postgres psql -U cloudbox -d cloudbox

# Run migrations manually
docker-compose exec backend go run cmd/migrate/main.go
```

## First Steps

### 1. Create an Account
1. Navigate to http://localhost:3000
2. Click "Create Account"
3. Fill in your details and register

### 2. Create Your First Project
1. Sign in to the dashboard
2. Click "New Project"
3. Enter project name and description
4. Your project will get a unique API namespace

### 3. Generate API Keys
1. Go to your project settings
2. Navigate to "API Keys" tab
3. Create a new API key
4. Copy the key for use in your applications

### 4. Configure CORS
1. In project settings, go to "CORS" tab
2. Add your application domains
3. Configure allowed methods and headers

## Using the API

### Authentication
All API requests require authentication via JWT tokens or API keys.

#### Login to get JWT token:
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "password"}'
```

#### Use API Key for project-specific APIs:
```bash
curl -X GET http://localhost:8080/p/my-project/api/data/users \
  -H "X-API-Key: your-api-key"
```

### Project API Endpoints
Each project gets its own API namespace at `/p/{project-slug}/api/`

```bash
# Get data
GET /p/my-project/api/data/table_name

# Create data
POST /p/my-project/api/data/table_name

# Update data
PUT /p/my-project/api/data/table_name/id

# Delete data
DELETE /p/my-project/api/data/table_name/id
```

## Next Steps

- [API Reference](./api-reference.md)
- [SDK Documentation](./sdk-documentation.md)
- [Deployment Guide](./deployment.md)
- [Development Guide](./development.md)

## Troubleshooting

### Common Issues

#### Port Already in Use
```bash
# Stop any existing CloudBox instances
make stop

# Or check what's using the ports
lsof -i :3000  # Frontend
lsof -i :8080  # Backend
lsof -i :5432  # PostgreSQL
```

#### Database Connection Issues
```bash
# Check PostgreSQL is running
docker-compose logs postgres

# Reset database
make clean
make start
```

#### Permission Issues
```bash
# Fix file permissions
sudo chown -R $USER:$USER .
```

### Getting Help
- Check the [Issues](https://github.com/your-org/cloudbox/issues) page
- Join our [Discord](https://discord.gg/cloudbox) community
- Read the [Documentation](./README.md)