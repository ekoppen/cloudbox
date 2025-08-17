# Script Runner Plugin - Quick Start Guide

CloudBox Universal Script Runner - Database scripts en project setup automation

## 🚀 Toegang

1. **Login** als admin/superadmin in CloudBox dashboard
2. **Navigeer** naar een project
3. **Klik** op "Scripts" in de project sidebar
4. **Begin** met het uitvoeren van scripts!

## ⚡ Snelle Setup

### Web App Template
Perfect voor standard web applications:

```sql
-- 1. User management
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. Sessions
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id INTEGER REFERENCES users(id),
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 3. API Keys
CREATE TABLE api_keys (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    key_hash VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### AI Chat App Template
Voor AI-powered chat applicaties:

```sql
-- 1. User profiles
CREATE TABLE user_profiles (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    display_name VARCHAR(100),
    avatar_url TEXT,
    preferences JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. Conversations
CREATE TABLE conversations (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    title VARCHAR(255),
    model VARCHAR(50) DEFAULT 'gpt-3.5-turbo',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 3. Messages
CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    conversation_id INTEGER REFERENCES conversations(id),
    role VARCHAR(20) CHECK (role IN ('user', 'assistant', 'system')),
    content TEXT NOT NULL,
    tokens INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### E-commerce Template
Voor webshops en e-commerce:

```sql
-- 1. Products
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL,
    stock_quantity INTEGER DEFAULT 0,
    category VARCHAR(100),
    images JSONB DEFAULT '[]',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. Shopping cart
CREATE TABLE cart_items (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    product_id INTEGER REFERENCES products(id),
    quantity INTEGER NOT NULL DEFAULT 1,
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, product_id)
);

-- 3. Orders
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    total_amount DECIMAL(10,2) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    payment_id TEXT,
    shipping_address JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## 🎯 Common Scripts

### Database Setup
```sql
-- Create application database
CREATE DATABASE {{project_name}}_app;
\c {{project_name}}_app;

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Set timezone
SET timezone = 'UTC';
```

### Function Deployment
```javascript
// Deploy authentication function
const CloudBox = require('cloudbox-sdk');

async function deployAuth() {
    await CloudBox.functions.deploy('user-auth', {
        runtime: 'nodejs18',
        memory: '256MB',
        timeout: '30s',
        env: {
            JWT_SECRET: process.env.JWT_SECRET
        }
    });
    
    console.log('Authentication function deployed');
}

deployAuth();
```

### API Endpoints Setup
```javascript
// Create REST API endpoints
const CloudBox = require('cloudbox-sdk');

async function setupAPI() {
    // User registration
    await CloudBox.webhooks.create('/api/auth/register', {
        method: 'POST',
        function: 'user-auth',
        action: 'register'
    });
    
    // User login
    await CloudBox.webhooks.create('/api/auth/login', {
        method: 'POST', 
        function: 'user-auth',
        action: 'login'
    });
    
    // Protected routes
    await CloudBox.webhooks.create('/api/user/profile', {
        method: 'GET',
        function: 'user-auth',
        action: 'getProfile',
        auth: true
    });
    
    console.log('API endpoints created');
}

setupAPI();
```

## 📋 Script Management

### Script Categories
- **Database** - Schema setup, migrations, seeds
- **Functions** - CloudBox function deployment
- **APIs** - Webhook and endpoint creation
- **Setup** - Project initialization scripts
- **Maintenance** - Cleanup and optimization

### Script Variables
Gebruik variabelen voor herbruikbaarheid:

```sql
-- In script
CREATE TABLE {{table_prefix}}_users (
    id SERIAL PRIMARY KEY,
    name VARCHAR({{max_name_length}})
);

-- Bij uitvoering
{
  "table_prefix": "app",
  "max_name_length": "100"
}
```

### Script Dependencies
Definieer uitvoervolgorde:

```json
{
  "scripts": [
    {
      "name": "01_database_setup",
      "dependencies": []
    },
    {
      "name": "02_user_tables", 
      "dependencies": ["01_database_setup"]
    },
    {
      "name": "03_functions",
      "dependencies": ["02_user_tables"]
    }
  ]
}
```

## 🔄 Workflows

### Development Workflow
1. **Setup Database** - Run database schema scripts
2. **Deploy Functions** - Deploy serverless functions
3. **Create APIs** - Setup REST endpoints
4. **Test Integration** - Verify all components work
5. **Seed Data** - Add test data for development

### Production Deployment
1. **Backup Current** - Create full backup
2. **Run Migrations** - Apply database changes
3. **Update Functions** - Deploy function updates
4. **Verify Health** - Run health checks
5. **Rollback if Needed** - Automatic rollback on failure

### Maintenance Tasks
1. **Cleanup Old Data** - Remove expired sessions, logs
2. **Update Statistics** - Refresh database statistics
3. **Check Performance** - Analyze slow queries
4. **Security Audit** - Check for security issues

## 🎨 UI Features

### Script Editor
- **Syntax Highlighting** - SQL, JavaScript, JSON
- **Auto-completion** - CloudBox SDK functions
- **Error Checking** - Real-time syntax validation
- **Variable Substitution** - Preview final script

### Execution Results
- **Real-time Output** - Live script execution results
- **Error Handling** - Clear error messages and stack traces
- **Execution Time** - Performance monitoring
- **Row Counts** - Database operation results

### Template Library
- **Pre-built Scripts** - Common setup patterns
- **Custom Templates** - Save your own script collections
- **Version Control** - Track template changes
- **Sharing** - Export/import script templates

## 🚨 Best Practices

### Security
- **Validate Inputs** - Always validate script variables
- **Use Transactions** - Wrap related operations in transactions
- **Limit Permissions** - Use least privilege principle
- **Audit Trail** - Log all script executions

### Performance
- **Index Strategy** - Create appropriate indexes
- **Query Optimization** - Use EXPLAIN to analyze queries
- **Resource Limits** - Set execution timeouts
- **Batch Operations** - Process large datasets in chunks

### Maintenance
- **Version Scripts** - Use semantic versioning
- **Document Changes** - Clear change descriptions
- **Test Thoroughly** - Test in development first
- **Backup Strategy** - Always backup before changes

## 🔗 Integration

### CloudBox SDK
```javascript
const CloudBox = require('cloudbox-sdk');

// Database access
const db = CloudBox.database;
const result = await db.query('SELECT * FROM users');

// Function management
const functions = CloudBox.functions;
await functions.deploy('my-function', config);

// Storage operations
const storage = CloudBox.storage;
await storage.upload(file, 'bucket-name');
```

### External APIs
```javascript
// Third-party integrations
const axios = require('axios');

async function syncWithExternal() {
    const response = await axios.get('https://api.external.com/data');
    
    // Process and store in CloudBox
    await CloudBox.database.query(
        'INSERT INTO external_data (data) VALUES ($1)',
        [JSON.stringify(response.data)]
    );
}
```

## 📞 Support

### Documentation
- **Plugin Docs** - `/docs/PLUGIN_SYSTEM.md`
- **Marketplace** - `/docs/PLUGIN_MARKETPLACE.md`
- **API Reference** - CloudBox REST API documentation

### Community
- **GitHub** - Issues and feature requests
- **Discord** - Real-time community support
- **Examples** - Community script templates

### Troubleshooting
- **Execution Logs** - Check script output in dashboard
- **Error Messages** - Review detailed error information
- **Health Checks** - Verify CloudBox system status
- **Permission Issues** - Check user roles and permissions

---

**CloudBox Script Runner** - Van database tot deployment in één klik! 🚀