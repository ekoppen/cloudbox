# CloudBox Installation Guide

Complete installation and setup guide for CloudBox Backend-as-a-Service platform.

## Prerequisites

Before installing CloudBox, ensure you have the following installed:

- **Docker** (v20.10+)
- **Docker Compose** (v2.0+ or built-in `docker compose`)  
- **Git**
- **OpenSSL** (for generating secrets)

### Installing Prerequisites

#### Ubuntu/Debian
```bash
# Update package index
sudo apt update

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh
sudo usermod -aG docker $USER

# Install Docker Compose (if not using built-in docker compose)
sudo curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Or verify built-in docker compose works
docker compose version

# Install Git and OpenSSL
sudo apt install git openssl
```

#### macOS
```bash
# Install Docker Desktop from https://docker.com/products/docker-desktop

# Install via Homebrew
brew install git
```

#### Windows
- Install Docker Desktop from https://docker.com/products/docker-desktop
- Install Git from https://git-scm.com/download/win

## Installation Methods

### Method 1: Automated Installation (Recommended)

The automated installation script handles all configuration and setup:

```bash
# Clone repository
git clone https://github.com/ekoppen/cloudbox.git
cd cloudbox

# Make script executable
chmod +x install.sh

# Basic installation
./install.sh

# Installation with custom options
./install.sh --frontend-port 8080 --backend-port 9000 --host myserver.com
```

#### Installation Script Options

| Option | Description | Default |
|--------|-------------|---------|
| `--help` | Show help message | - |
| `--update` | Update existing installation | - |
| `--frontend-port` | Frontend port | 3000 |
| `--backend-port` | Backend port | 8080 |
| `--db-port` | Database port | 5432 |
| `--redis-port` | Redis port | 6379 |
| `--host` | Add allowed host | - |
| `--env-file` | Environment file path | .env |

#### Examples

```bash
# Install for development
./install.sh

# Install for remote server
./install.sh --host myserver.example.com

# Install with custom ports (useful if ports are occupied)
./install.sh --frontend-port 8080 --backend-port 9000 --db-port 5433

# Update existing installation
./install.sh --update
```

### Method 2: Manual Installation

If you prefer manual control or need custom configuration:

#### 1. Clone Repository

```bash
git clone https://github.com/ekoppen/cloudbox.git
cd cloudbox
```

#### 2. Create Environment File

```bash
# Copy example environment file
cp .env.example .env

# Edit configuration
nano .env
```

#### 3. Configure Environment Variables

```bash
# Database Configuration
DB_HOST=postgres
DB_PORT=5432
DB_USER=cloudbox
DB_PASSWORD=your_secure_password
DB_NAME=cloudbox

# Redis Configuration  
REDIS_HOST=redis
REDIS_PORT=6379

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-here
JWT_EXPIRES_IN=24h

# Server Configuration
SERVER_PORT=8080
FRONTEND_URL=http://localhost:3000

# Docker Port Configuration
FRONTEND_PORT=3000
BACKEND_PORT=8080
POSTGRES_PORT=5432
REDIS_EXTERNAL_PORT=6379
```

#### 4. Start Services

```bash
# Build and start all services
docker-compose up -d

# Check service status
docker-compose ps

# View logs
docker-compose logs -f
```

## Post-Installation Setup

### 1. Verify Installation

```bash
# Check if all services are running
docker-compose ps

# Test backend API
curl http://localhost:8080/health

# Test frontend
curl http://localhost:3000
```

### 2. Access Admin Dashboard

1. Open http://localhost:3000 in your browser
2. Navigate to http://localhost:3000/admin  
3. Login with default credentials:
   - Email: `admin@cloudbox.local`
   - Password: `admin123`

### 3. Change Default Credentials

**Important**: Change the default admin password immediately:

1. Go to Admin Dashboard
2. Navigate to Users section
3. Edit the admin user
4. Set a strong password

### 4. Create Your First Project

1. Login to the main dashboard
2. Click "Create Project"
3. Fill in project details:
   - Name: Your project name
   - Description: Project description
   - Slug: URL-friendly identifier
4. Note your API key for SDK integration

## Configuration

### Port Configuration

If default ports are occupied, you can change them:

#### Using Installation Script
```bash
./install.sh --frontend-port 8080 --backend-port 9000
```

#### Manual Configuration

Edit `docker-compose.yml`:
```yaml
services:
  frontend:
    ports:
      - "8080:3000"  # Change external port
  backend:
    ports:
      - "9000:8080"  # Change external port
```

### Remote Access Configuration

For remote server access, configure allowed hosts:

#### Using Installation Script
```bash
./install.sh --host myserver.example.com
```

#### Manual Configuration

Edit `frontend/vite.config.ts`:
```typescript
export default defineConfig({
  plugins: [sveltekit()],
  server: {
    host: '0.0.0.0',
    port: 3000,
    allowedHosts: ['myserver.example.com', 'localhost']
  }
});
```

### SSL/HTTPS Configuration

For production deployment with SSL:

1. Create SSL certificates
2. Configure reverse proxy (nginx/Apache)
3. Update environment variables:

```bash
# .env
FRONTEND_URL=https://your-domain.com
API_URL=https://api.your-domain.com
```

## Troubleshooting

### Common Issues

#### Port Already in Use

```bash
# Check what's using the port
sudo lsof -i :3000

# Kill the process
sudo kill -9 <PID>

# Or use different ports
./install.sh --frontend-port 8080
```

#### Docker Permission Issues

```bash
# Add user to docker group
sudo usermod -aG docker $USER

# Logout and login again
```

#### Database Connection Issues

```bash
# Check database logs
docker-compose logs postgres

# Restart database
docker-compose restart postgres

# Reset database
docker-compose down -v
docker-compose up -d
```

#### Memory Issues

```bash
# Check available memory
free -h

# Increase Docker memory limit in Docker Desktop
# Or add swap space on Linux
sudo fallocate -l 2G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
```

#### Docker Compose Version Issues

If you get command not found errors:

```bash
# Check which version you have
docker compose version    # Modern built-in version
docker-compose version    # Standalone version

# The install script automatically detects and uses the correct version
```

### Getting Help

1. Check logs: `docker compose logs -f` (or `docker-compose logs -f`)
2. Verify configuration: `docker compose config`
3. Check service status: `docker compose ps`
4. Restart services: `docker compose restart`

## Maintenance

### Backup

```bash
# Backup database
docker-compose exec postgres pg_dump -U cloudbox cloudbox > backup.sql

# Backup uploads
tar -czf uploads_backup.tar.gz uploads/
```

### Updates

```bash
# Update using script
./install.sh --update

# Or manual update
git pull
docker-compose pull
docker-compose up -d --build
```

### Monitoring

```bash
# View resource usage
docker stats

# Monitor logs
docker-compose logs -f --tail=100

# Check disk usage
df -h
```

## Security Checklist

- [ ] Change default admin credentials
- [ ] Use strong JWT secret
- [ ] Configure firewall rules
- [ ] Enable SSL/HTTPS in production
- [ ] Regular security updates
- [ ] Monitor access logs
- [ ] Backup data regularly
- [ ] Use secure passwords

## Next Steps

1. [Configure your first project](./getting-started.md)
2. [Explore the API](./api-reference.md)
3. [Try the SDKs](./sdk.md)
4. [Set up deployment](./deployment.md)

## Support

- 📚 [Documentation](https://github.com/ekoppen/cloudbox/docs)
- 🐛 [Report Issues](https://github.com/ekoppen/cloudbox/issues)
- 💬 [Community Discussions](https://github.com/ekoppen/cloudbox/discussions)