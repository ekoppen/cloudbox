# CloudBox Troubleshooting Guide

Common issues and solutions for CloudBox deployment and operation.

## Container Startup Issues

### Backend Not Starting / Database Connection Issues

**Problem**: Backend fails to start or can't connect to database
```
ERROR: failed to connect to database
```

**Solutions**:

1. **Check service health status**:
```bash
docker compose ps
```

2. **View service logs**:
```bash
# Check database logs
docker compose logs postgres

# Check backend logs  
docker compose logs backend

# Check all logs
docker compose logs -f
```

3. **Manual service restart order**:
```bash
# Stop all services
docker compose down

# Start database first and wait
docker compose up -d postgres
docker compose ps  # Wait until postgres shows "healthy"

# Start redis
docker compose up -d redis

# Start backend (will wait for healthy services)
docker compose up -d backend

# Finally start frontend
docker compose up -d frontend
```

4. **Reset and rebuild**:
```bash
# Complete reset
docker compose down -v
docker compose build --no-cache
docker compose up -d
```

### Long Database Startup Times

**Problem**: PostgreSQL takes a long time to become ready

**Solutions**:

1. **Check available resources**:
```bash
# Check memory
free -h

# Check disk space
df -h

# Check Docker resources
docker stats
```

2. **Increase health check timeouts**:
Edit `docker-compose.yml`:
```yaml
postgres:
  healthcheck:
    start_period: 60s  # Increase from 30s
    retries: 15        # Increase from 10
```

3. **Monitor startup progress**:
```bash
# Watch database logs in real-time
docker compose logs -f postgres

# Check if database files are being created
docker compose exec postgres ls -la /var/lib/postgresql/data/
```

### Service Dependencies Issues

**Problem**: Services start in wrong order or don't wait for dependencies

**Verification**:
```bash
# Check depends_on configuration
docker compose config

# Verify health checks are working
docker compose ps
```

**Fix**: Ensure proper configuration in `docker-compose.yml`:
```yaml
backend:
  depends_on:
    postgres:
      condition: service_healthy
    redis:
      condition: service_healthy
```

## Network and Port Issues

### Port Already in Use

**Problem**:
```
ERROR: Port 3000 is already allocated
```

**Solutions**:

1. **Find what's using the port**:
```bash
# Linux/macOS
sudo lsof -i :3000
sudo ss -tlnp | grep :3000

# Kill the process
sudo kill -9 <PID>
```

2. **Use different ports**:
```bash
./install.sh --frontend-port 8080 --backend-port 9000
```

3. **Stop conflicting services**:
```bash
# Stop common development servers
pkill -f "node.*3000"
pkill -f "npm.*dev"
```

### Frontend Can't Connect to Backend

**Problem**: Frontend shows API connection errors

**Solutions**:

1. **Check backend is running**:
```bash
curl http://localhost:8080/health
```

2. **Verify environment variables**:
```bash
docker compose exec frontend env | grep API
```

3. **Check CORS configuration** in backend
4. **Verify network connectivity**:
```bash
docker compose exec frontend ping backend
```

## Authentication and Admin Issues

### Can't Access Admin Dashboard

**Problem**: 403 or 401 errors when accessing `/admin`

**Solutions**:

1. **Check user role in database**:
```bash
docker compose exec postgres psql -U cloudbox -d cloudbox -c "SELECT email, role, is_active FROM users;"
```

2. **Update user to superadmin**:
```bash
docker compose exec postgres psql -U cloudbox -d cloudbox -c "UPDATE users SET role = 'superadmin' WHERE email = 'admin@cloudbox.local';"
```

3. **Clear browser cache and cookies**
4. **Login with correct credentials**:
   - Email: `admin@cloudbox.local`
   - Password: `admin123`

### JWT Token Issues

**Problem**: Authentication keeps failing

**Solutions**:

1. **Check JWT secret consistency**:
```bash
# Verify JWT_SECRET in environment
docker compose exec backend env | grep JWT
```

2. **Restart backend to refresh tokens**:
```bash
docker compose restart backend
```

3. **Clear browser storage**:
   - Open browser dev tools
   - Go to Application/Storage tab
   - Clear localStorage and sessionStorage

## Database Issues

### Database Connection Refused

**Problem**:
```
pq: connection refused
```

**Solutions**:

1. **Check PostgreSQL is running**:
```bash
docker compose ps postgres
```

2. **Test direct connection**:
```bash
docker compose exec postgres psql -U cloudbox -d cloudbox -c "SELECT version();"
```

3. **Check database credentials**:
```bash
# Verify credentials match in docker-compose.yml and backend config
docker compose config | grep -A 10 postgres
```

### Database Schema Issues

**Problem**: Tables missing or schema errors

**Solutions**:

1. **Check if migrations ran**:
```bash
docker compose exec postgres psql -U cloudbox -d cloudbox -c "\dt"
```

2. **Manually run migrations**:
```bash
# Copy migration files if not present
docker compose exec postgres ls -la /docker-entrypoint-initdb.d/

# Reset database and restart
docker compose down postgres
docker volume rm cloudbox_postgres_data
docker compose up -d postgres
```

## Performance Issues

### Slow Response Times

**Problem**: API responds slowly

**Solutions**:

1. **Check resource usage**:
```bash
docker stats
```

2. **Increase Docker resources**:
   - Docker Desktop: Settings > Resources
   - Linux: Check available memory/CPU

3. **Check database performance**:
```bash
# Monitor database activity
docker compose exec postgres psql -U cloudbox -d cloudbox -c "SELECT * FROM pg_stat_activity;"
```

### High Memory Usage

**Problem**: Containers using too much memory

**Solutions**:

1. **Restart resource-heavy containers**:
```bash
docker compose restart frontend backend
```

2. **Check for memory leaks**:
```bash
# Monitor memory over time
watch docker stats
```

3. **Add swap space** (Linux):
```bash
sudo fallocate -l 2G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
```

## File and Volume Issues

### Permission Denied Errors

**Problem**:
```
ERROR: Permission denied
```

**Solutions**:

1. **Fix file permissions**:
```bash
# Make scripts executable
chmod +x install.sh
chmod +x backend/wait-for-services.sh

# Fix ownership if needed
sudo chown -R $USER:$USER .
```

2. **Check Docker volume permissions**:
```bash
docker compose exec postgres ls -la /var/lib/postgresql/data/
```

### Upload Directory Issues

**Problem**: File uploads fail

**Solutions**:

1. **Create upload directories**:
```bash
mkdir -p uploads
chmod 755 uploads
```

2. **Check volume mounts**:
```bash
docker compose config | grep -A 5 volumes
```

## Environment Configuration

### Environment Variables Not Loading

**Problem**: Configuration not being applied

**Solutions**:

1. **Check .env file exists**:
```bash
ls -la .env
```

2. **Verify environment loading**:
```bash
docker compose config
```

3. **Recreate containers with new environment**:
```bash
docker compose down
docker compose up -d
```

## Quick Diagnostic Commands

### Full System Check
```bash
#!/bin/bash
echo "=== CloudBox System Diagnostic ==="

echo "1. Docker status:"
docker --version
docker compose version

echo "2. Container status:"
docker compose ps

echo "3. Service health:"
docker compose exec postgres pg_isready -U cloudbox || echo "PostgreSQL not ready"
docker compose exec redis redis-cli ping || echo "Redis not ready"

echo "4. Port status:"
ss -tlnp | grep -E "(3000|8080|5432|6379)"

echo "5. Recent logs:"
docker compose logs --tail=20

echo "6. Disk space:"
df -h

echo "7. Memory usage:"
free -h

echo "=== End Diagnostic ==="
```

### Log Collection
```bash
# Create diagnostic log file
echo "Collecting CloudBox logs..."
mkdir -p logs
docker compose logs > logs/all-services.log 2>&1
docker compose ps > logs/container-status.log
docker stats --no-stream > logs/resource-usage.log
tar -czf cloudbox-logs-$(date +%Y%m%d_%H%M%S).tar.gz logs/
echo "Logs saved to cloudbox-logs-*.tar.gz"
```

## Getting Help

If you're still experiencing issues:

1. **Check our GitHub Issues**: https://github.com/ekoppen/cloudbox/issues
2. **Create a new issue** with:
   - Your operating system
   - Docker and Docker Compose versions
   - Output of diagnostic commands above
   - Error messages and logs
   - Steps to reproduce the issue

3. **Join our Community**: 
   - GitHub Discussions: https://github.com/ekoppen/cloudbox/discussions
   - Discord: (coming soon)

## Common Command Reference

```bash
# Restart specific service
docker compose restart backend

# Rebuild and restart
docker compose up -d --build backend

# View real-time logs
docker compose logs -f backend

# Execute commands in container
docker compose exec backend sh
docker compose exec postgres psql -U cloudbox -d cloudbox

# Clean up everything
docker compose down -v --remove-orphans
docker system prune -a

# Update to latest version
git pull
./install.sh --update
```