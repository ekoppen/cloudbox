#!/bin/bash

# Test database connections for CloudBox

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}🔍 Testing CloudBox Database Connections${NC}"
echo "========================================"

# Get postgres IP
POSTGRES_IP=$(docker inspect cloudbox-postgres --format='{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}')
echo "Postgres Container IP: $POSTGRES_IP"

# Test direct connection via docker exec
echo -e "\n${BLUE}1. Testing docker exec connection:${NC}"
if docker exec cloudbox-postgres psql -U cloudbox -d cloudbox -c "SELECT 'Connection successful!' as status;" 2>/dev/null; then
    echo -e "${GREEN}✅ Docker exec connection works${NC}"
else
    echo -e "${RED}❌ Docker exec connection failed${NC}"
fi

# Test network connection from python container
echo -e "\n${BLUE}2. Testing network connection from Python:${NC}"
python_result=$(docker run --rm --network cloudbox_default python:3.9-alpine sh -c "
    pip install psycopg2-binary > /dev/null 2>&1
    python3 -c \"
import psycopg2
try:
    conn = psycopg2.connect(host='$POSTGRES_IP', port=5432, database='cloudbox', user='cloudbox', password='cloudbox_dev_password')
    print('Python network connection successful!')
    conn.close()
except Exception as e:
    print(f'Python network connection failed: {e}')
\"
" 2>/dev/null)

echo "$python_result"

# Try with hostname
echo -e "\n${BLUE}3. Testing network connection using hostname:${NC}"
hostname_result=$(docker run --rm --network cloudbox_default python:3.9-alpine sh -c "
    pip install psycopg2-binary > /dev/null 2>&1
    python3 -c \"
import psycopg2
try:
    conn = psycopg2.connect(host='cloudbox-postgres', port=5432, database='cloudbox', user='cloudbox', password='cloudbox_dev_password')
    print('Hostname connection successful!')
    conn.close()
except Exception as e:
    print(f'Hostname connection failed: {e}')
\"
" 2>/dev/null)

echo "$hostname_result"

# Test port mapping
echo -e "\n${BLUE}4. Testing host port mapping:${NC}"
if docker ps | grep cloudbox-postgres | grep -q "5432.*5432"; then
    echo -e "${GREEN}✅ PostgreSQL port 5432 is mapped to host${NC}"
    echo "You can use: postgres://cloudbox:cloudbox_dev_password@localhost:5432/cloudbox"
else
    echo -e "${RED}❌ PostgreSQL port not mapped to host${NC}"
    echo "Use network connection: postgres://cloudbox:cloudbox_dev_password@$POSTGRES_IP:5432/cloudbox"
fi