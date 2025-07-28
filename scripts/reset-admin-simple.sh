#!/bin/bash

# Simple CloudBox Admin Password Reset Script
# Uses direct SQL commands via docker exec (no external dependencies)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🔑 CloudBox SuperAdmin Password Reset${NC}"
echo "====================================="
echo -e "${YELLOW}⚠️ This tool resets/creates CloudBox superadmin accounts${NC}"
echo

# Check if CloudBox postgres container is running
if ! docker ps | grep -q "cloudbox-postgres"; then
    echo -e "${RED}❌ CloudBox postgres container not running${NC}"
    echo "Please start CloudBox first: docker compose up -d"
    exit 1
fi

echo -e "${GREEN}✅ Found running CloudBox postgres container${NC}"
echo

# Get user input
read -p "📧 Enter superadmin email address: " email
if [[ -z "$email" ]]; then
    echo -e "${RED}❌ Email cannot be empty${NC}"
    exit 1
fi

# Validate email format
if [[ ! "$email" =~ ^[^@]+@[^@]+\.[^@]+$ ]]; then
    echo -e "${RED}❌ Please enter a valid email address${NC}"
    exit 1
fi

# Get password
echo -n "🔑 Enter new superadmin password: "
read -s password
echo
if [[ ${#password} -lt 8 ]]; then
    echo -e "${RED}❌ Superadmin password must be at least 8 characters long${NC}"
    exit 1
fi

# Confirm password
echo -n "🔑 Confirm superadmin password: "
read -s password_confirm
echo
if [[ "$password" != "$password_confirm" ]]; then
    echo -e "${RED}❌ Passwords do not match${NC}"
    exit 1
fi

# Confirm action
echo -e "${YELLOW}⚠️ This will create/reset superadmin account for '$email'${NC}"
echo -e "${YELLOW}⚠️ This user will have FULL ACCESS to all CloudBox features${NC}"
read -p "Are you sure you want to continue? (yes/no): " confirm
if [[ "$confirm" != "yes" && "$confirm" != "y" ]]; then
    echo -e "${RED}❌ Operation cancelled${NC}"
    exit 1
fi

echo
echo -e "${BLUE}🔐 Processing password reset...${NC}"

# Generate proper bcrypt hash using Python in Docker
echo -e "${BLUE}🔐 Generating secure bcrypt hash...${NC}"
bcrypt_hash=$(docker run --rm python:3.9-alpine sh -c "
    pip install bcrypt > /dev/null 2>&1
    python3 -c \"
import bcrypt
password = '''$password'''
hashed = bcrypt.hashpw(password.encode('utf-8'), bcrypt.gensalt())
print(hashed.decode('utf-8'))
\"
" 2>/dev/null)

# Verify hash was generated correctly
if [[ -z "$bcrypt_hash" || ! "$bcrypt_hash" =~ ^\$2[aby]\$[0-9]{2}\$ ]]; then
    echo -e "${YELLOW}⚠️ Docker bcrypt generation failed, trying alternative method...${NC}"
    
    # Try using the CloudBox backend container if available
    bcrypt_hash=$(docker exec cloudbox-backend python3 -c "
import bcrypt
password = '''$password'''
hashed = bcrypt.hashpw(password.encode('utf-8'), bcrypt.gensalt())
print(hashed.decode('utf-8'))
" 2>/dev/null || echo "")
    
    if [[ -z "$bcrypt_hash" || ! "$bcrypt_hash" =~ ^\$2[aby]\$[0-9]{2}\$ ]]; then
        echo -e "${RED}❌ Failed to generate bcrypt hash${NC}"
        echo -e "${BLUE}ℹ️ Please use one of these alternatives:${NC}"
        echo "  ./reset-admin.sh --python"
        echo "  ./reset-admin.sh --docker" 
        exit 1
    fi
fi

echo -e "${GREEN}✅ Bcrypt hash generated successfully${NC}"

# Current timestamp
timestamp=$(date '+%Y-%m-%d %H:%M:%S')

# Check if user exists and get info
user_info=$(docker exec cloudbox-postgres psql -U cloudbox -d cloudbox -t -c "SELECT id, name FROM users WHERE email = '$email';" 2>/dev/null | tr -d ' ')

if [[ -n "$user_info" && "$user_info" != *"(0 rows)"* ]]; then
    # User exists - update password
    echo -e "${BLUE}👤 Found existing user, updating password...${NC}"
    
    docker exec cloudbox-postgres psql -U cloudbox -d cloudbox -c "
        UPDATE users 
        SET password_hash = '$bcrypt_hash', 
            role = 'superadmin', 
            is_active = true, 
            updated_at = '$timestamp'
        WHERE email = '$email';
    " > /dev/null
    
    echo -e "${GREEN}✅ Password updated for existing user${NC}"
else
    # User doesn't exist - create new user
    echo -e "${BLUE}👤 User not found, creating new admin user...${NC}"
    
    # Get name from email or prompt
    default_name=$(echo "$email" | cut -d'@' -f1 | sed 's/[._]/ /g' | awk '{for(i=1;i<=NF;i++) $i=toupper(substr($i,1,1)) tolower(substr($i,2))}1')
    read -p "👤 Enter full name for superadmin (default: $default_name): " name
    if [[ -z "$name" ]]; then
        name="$default_name"
    fi
    
    docker exec cloudbox-postgres psql -U cloudbox -d cloudbox -c "
        INSERT INTO users (created_at, updated_at, email, name, password_hash, role, is_active)
        VALUES ('$timestamp', '$timestamp', '$email', '$name', '$bcrypt_hash', 'superadmin', true);
    " > /dev/null
    
    echo -e "${GREEN}✅ New admin user created${NC}"
fi

# Get final user info
final_info=$(docker exec cloudbox-postgres psql -U cloudbox -d cloudbox -c "
    SELECT id, name, email, role, is_active 
    FROM users 
    WHERE email = '$email';
")

echo
echo -e "${GREEN}🎉 Operation completed successfully!${NC}"
echo -e "${BLUE}Final user information:${NC}"
echo "$final_info"
echo
echo -e "${GREEN}✅ You can now login to CloudBox with these credentials${NC}"