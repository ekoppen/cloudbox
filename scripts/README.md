# CloudBox SuperAdmin Password Reset Tools

This directory contains multiple tools for creating and resetting CloudBox superadmin accounts. These accounts have full administrative access to all CloudBox features including user management, system settings, and administrative dashboard.

## Available Tools

### 1. Shell Script (Recommended) - `reset-admin.sh`

The easiest way to create/reset a superadmin account. Automatically detects your environment and uses the appropriate method.

```bash
# Auto-detect best mode
./reset-admin.sh

# Docker mode (builds Go in container)
./reset-admin.sh --docker

# Local Go mode
./reset-admin.sh --local

# Python mode (with auto-install)
./reset-admin.sh --python
```

### 2. Simple Shell Script - `reset-admin-simple.sh`

Ultra-simple version that only needs Docker and CloudBox running. No external dependencies.

```bash
# Only requires CloudBox to be running
./reset-admin-simple.sh
```

### 3. Go Script - `reset-admin-password.go`

Direct Go implementation with full type safety and error handling.

```bash
# First time setup
go mod tidy

# Run the script
go run reset-admin-password.go
```

### 4. Python Script - `reset-admin.py`

Python implementation for environments where Go is not available.

```bash
# Install dependencies
pip install -r requirements.txt

# Run the script
python reset-admin.py
```

## Prerequisites

### For Shell Script (Docker Mode)
- Docker and Docker Compose
- CloudBox containers running (`docker compose up -d`)

### For Shell Script (Local Mode)
- Go 1.21 or later
- PostgreSQL access

### For Go Script
- Go 1.21 or later
- PostgreSQL access

### For Python Script
- Python 3.6 or later
- pip package manager
- PostgreSQL access

## Environment Variables

All scripts support these environment variables:

```bash
# Database connection
export DATABASE_URL="postgres://user:password@host:port/database?sslmode=disable"

# Alternative format
export DB_CONNECTION_STRING="postgres://user:password@host:port/database"
```

**Default connection:**
```
postgres://cloudbox:cloudbox_dev_password@localhost:5432/cloudbox?sslmode=disable
```

## Usage Examples

### Quick Reset with Docker (Recommended)

```bash
# Make sure CloudBox is running
docker compose up -d

# Reset admin password
./reset-admin.sh --docker
```

### Local Development

```bash
# Set custom database URL
export DATABASE_URL="postgres://cloudbox:cloudbox_dev_password@localhost:5432/cloudbox"

# Run reset tool
./reset-admin.sh
```

### Custom Environment

```bash
# For remote database
export DATABASE_URL="postgres://admin:secret@db.example.com:5432/cloudbox_prod?sslmode=require"

# Run Python version
python reset-admin.py
```

## What the Script Does

1. **Connects to Database**: Uses provided or default connection string
2. **Prompts for Email**: Admin user email address
3. **Prompts for Password**: New password (minimum 6 characters, confirmed)
4. **Confirms Action**: Shows what will happen before proceeding

### If User Exists:
- ✅ Updates password to new hashed value
- ✅ Sets role to 'admin'
- ✅ Activates account (sets is_active = true)

### If User Doesn't Exist:
- ✅ Creates new user with provided email
- ✅ Prompts for full name (defaults to email username)
- ✅ Sets role to 'admin'
- ✅ Activates account

## Security Features

- 🔐 **Password Hashing**: Uses bcrypt with default cost (10 rounds)
- 🙈 **Hidden Input**: Password input is hidden from terminal
- ✅ **Password Confirmation**: Requires password to be entered twice
- 🛡️ **Input Validation**: Validates email format and password length
- ⚠️ **Confirmation Step**: Asks for confirmation before making changes

## Troubleshooting

### Connection Issues

```bash
# Test database connection
docker exec cloudbox-postgres psql -U cloudbox -d cloudbox -c "\dt"

# Check if containers are running
docker ps | grep cloudbox
```

### Permission Issues

```bash
# Make scripts executable
chmod +x reset-admin.sh
chmod +x reset-admin.py
```

### Go Dependencies

```bash
# In scripts directory
go mod download
go mod tidy
```

### Python Dependencies

```bash
# Install required packages
pip install bcrypt psycopg2-binary

# Or use requirements file
pip install -r requirements.txt
```

### Docker Issues

```bash
# Rebuild containers if needed
docker compose down
docker compose up --build -d

# Check logs
docker logs cloudbox-postgres
docker logs cloudbox-backend
```

### Go Module Issues

If you encounter checksum mismatch errors:

```bash
# Clean up and let Docker regenerate dependencies
rm -f go.sum
./reset-admin.sh --docker
```

The Docker build will automatically generate a fresh go.sum file with correct checksums.

## Database Schema

The scripts work with this user table structure:

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    email VARCHAR UNIQUE NOT NULL,
    name VARCHAR NOT NULL,
    password VARCHAR NOT NULL,
    role VARCHAR DEFAULT 'user',
    is_active BOOLEAN DEFAULT true
);
```

## Examples

### Creating First Admin User

```bash
$ ./reset-admin.sh --docker
🔑 CloudBox Admin Password Reset
================================

🐳 Running in Docker mode...
🔨 Building reset tool...
🚀 Starting password reset...
📧 Enter admin email address: admin@example.com
🔑 Enter new password: [hidden]
🔑 Confirm new password: [hidden]
👤 Enter full name (default: Admin): System Administrator
⚠️  This will reset the password for 'admin@example.com' and set role to 'admin'
Are you sure you want to continue? (yes/no): yes
🔐 Hashing password...
👤 User admin@example.com not found. Creating new admin user...
✅ Admin user 'admin@example.com' created successfully

🎉 Operation completed successfully!
📧 Email: admin@example.com
👤 Name: System Administrator
🛡️  Role: admin
✅ Active: true
🆔 User ID: 1

You can now login to CloudBox with these credentials.
```

### Resetting Existing User

```bash
$ python reset-admin.py
🔑 CloudBox Admin Password Reset Tool (Python)
==============================================

📡 Connecting to database...
✅ Connected to database successfully

📧 Enter admin email address: john@company.com
🔑 Enter new password: [hidden]
🔑 Confirm new password: [hidden]
⚠️  This will reset the password for 'john@company.com' and set role to 'admin'
Are you sure you want to continue? (yes/no): yes
🔐 Hashing password...
👤 Found existing user: John Doe (john@company.com)
✅ Password reset successfully for user 'john@company.com'
✅ User role set to 'admin'
✅ User account activated

🎉 Operation completed successfully!
📧 Email: john@company.com
👤 Name: John Doe
🛡️  Role: admin
✅ Active: true
🆔 User ID: 5
```

## Integration with CloudBox

After running the password reset:

1. **Login to CloudBox**: 
   - Go to your CloudBox login page
   - Use the email and new password you set
   - The account will have `role = 'admin'` which gives full administrative access

2. **Access Admin Dashboard**: 
   - Look for "Admin" button in the dashboard
   - Access user management, system settings, statistics
   - Full superadmin privileges are active

3. **Verify Permissions**: 
   - Check that all admin features are accessible
   - Test user management functions
   - Verify access to system statistics and settings

**Important**: These are superadmin accounts with full system access. Protect the credentials accordingly.

## Production Considerations

### Security
- 🔒 Run scripts on secure, trusted machines only
- 🗑️ Clear terminal history after running scripts
- 📝 Log admin password resets for audit purposes
- 🔄 Consider requiring password changes on first login

### Backup
- 💾 Always backup database before running scripts
- 📊 Test scripts in development environment first
- 🔄 Have rollback procedures ready

### Monitoring
- 📈 Monitor for unauthorized admin access after resets
- 🚨 Set up alerts for admin role changes
- 📝 Document who performed password resets and when

## Support

If you encounter issues:

1. Check the [troubleshooting section](#troubleshooting)
2. Verify database connectivity
3. Ensure proper permissions
4. Check CloudBox logs for errors
5. Contact your system administrator

---

**⚠️ Important**: These tools modify user accounts with admin privileges. Use with caution and follow your organization's security policies.