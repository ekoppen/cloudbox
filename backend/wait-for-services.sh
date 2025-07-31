#!/bin/sh

# Wait for services to be ready before starting the backend
# This ensures proper startup order even with health checks

set -e

echo "🔍 Waiting for required services to be ready..."

# Function to wait for a service
wait_for_service() {
    local host=$1
    local port=$2
    local service_name=$3
    local max_attempts=60
    local attempt=1

    echo "⏳ Waiting for $service_name ($host:$port)..."
    
    while [ $attempt -le $max_attempts ]; do
        if nc -z "$host" "$port" 2>/dev/null; then
            echo "✅ $service_name is ready!"
            return 0
        fi
        
        echo "   Attempt $attempt/$max_attempts: $service_name not ready yet..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    echo "❌ $service_name failed to become ready after $max_attempts attempts"
    return 1
}

# Function to test database connection
test_database_connection() {
    local max_attempts=30
    local attempt=1
    
    echo "🗄️  Testing database connection..."
    
    while [ $attempt -le $max_attempts ]; do
        if PGPASSWORD="${DB_PASSWORD}" psql -h postgres -U cloudbox -d cloudbox -c "SELECT 1;" >/dev/null 2>&1; then
            echo "✅ Database connection successful!"
            return 0
        fi
        
        echo "   Attempt $attempt/$max_attempts: Database not ready for connections..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    echo "❌ Database connection failed after $max_attempts attempts"
    return 1
}

# Function to test Redis connection
test_redis_connection() {
    local max_attempts=15
    local attempt=1
    
    echo "📦 Testing Redis connection..."
    
    while [ $attempt -le $max_attempts ]; do
        if redis-cli -h redis ping >/dev/null 2>&1; then
            echo "✅ Redis connection successful!"
            return 0
        fi
        
        echo "   Attempt $attempt/$max_attempts: Redis not ready for connections..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    echo "❌ Redis connection failed after $max_attempts attempts"
    return 1
}

# Install required tools if not present
echo "🔧 Installing required tools..."
apk add --no-cache netcat-openbsd postgresql-client redis >/dev/null 2>&1

# Wait for basic network connectivity
wait_for_service "postgres" "5432" "PostgreSQL"
wait_for_service "redis" "6379" "Redis"

# Test actual service connections
test_database_connection
test_redis_connection

echo "🚀 All services are ready! Starting backend..."