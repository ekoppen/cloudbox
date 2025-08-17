# CloudBox Universal Script Runner Plugin

Een universele script runner voor CloudBox - vergelijkbaar met Supabase's SQL Editor maar dan voor alle soorten projecten en frameworks.

## ✨ Features

- **SQL Scripts**: Database schemas, migrations en queries
- **JavaScript Scripts**: Serverless functions en automation
- **Setup Scripts**: Project configuratie en deployment
- **Migration Scripts**: Database versioning en updates
- **Template Library**: Pre-built scripts voor verschillende project types
- **Dependency Management**: Automatische volgorde van script uitvoering
- **Multi-Project Support**: Werkt met alle CloudBox projecten
- **Audit Logging**: Volledige history van script executions
- **Rollback Support**: Undo functionaliteit voor migrations

## 🚀 Installatie

### Automatische Installatie
```bash
# In CloudBox directory
cd /path/to/cloudbox
cp -r plugins/script-runner/* /path/to/your/cloudbox/plugins/script-runner/

# Restart CloudBox
docker-compose restart
```

### Handmatige Installatie
1. Kopieer plugin files naar `cloudbox/plugins/script-runner/`
2. Restart CloudBox services
3. Plugin wordt automatisch gedetecteerd en geladen

## 📋 Gebruik

### Dashboard
- Ga naar **Script Runner** tab in CloudBox dashboard
- Selecteer een project
- Browse beschikbare scripts en templates
- Voer scripts uit individueel of in batch

### Script Types

#### SQL Scripts
```sql
-- Database schema setup
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
```

#### JavaScript Scripts
```javascript
// CloudBox function deployment
const functions = CloudBox.functions;

async function deployUserAPI() {
    await functions.deploy('user-api', {
        runtime: 'nodejs18',
        port: 3001,
        memory: '512MB'
    });
    
    console.log('User API deployed successfully');
}

deployUserAPI();
```

#### Setup Scripts
```bash
# Project setup commands
CREATE_FUNCTION user-authentication nodejs18 512MB
CREATE_WEBHOOK /api/users POST user-authentication
CREATE_DATABASE users_db
SCHEDULE_JOB backup-users "0 2 * * *" backup-function
```

### Templates

#### Web App Template
- User authentication system
- Session management
- API key management
- Basic CRUD operations

#### AI Chat App Template
- Conversation storage
- User profiles
- Message history
- AI integration setup

#### E-commerce Template
- Product catalog
- Shopping cart
- Payment integration
- Order management

### Variables
Scripts ondersteunen variabelen voor herbruikbaarheid:

```sql
CREATE DATABASE {{database_name}};
USE {{database_name}};

CREATE TABLE {{table_prefix}}_users (
    id SERIAL PRIMARY KEY,
    name VARCHAR({{max_name_length}})
);
```

## 🔧 API Endpoints

### Script Management
- `GET /api/plugins/script-runner/scripts/:projectId` - List project scripts
- `POST /api/plugins/script-runner/scripts/:projectId` - Create new script
- `PUT /api/plugins/script-runner/scripts/:projectId/:scriptId` - Update script
- `DELETE /api/plugins/script-runner/scripts/:projectId/:scriptId` - Delete script

### Execution
- `POST /api/plugins/script-runner/execute/:projectId/:scriptId` - Execute single script
- `POST /api/plugins/script-runner/execute-collection/:projectId/:collectionName` - Execute script collection
- `GET /api/plugins/script-runner/executions/:projectId` - Get execution history

### Templates
- `GET /api/plugins/script-runner/templates` - List available templates
- `GET /api/plugins/script-runner/project-templates` - List project setup templates
- `POST /api/plugins/script-runner/setup-project/:projectId/:templateName` - Setup project from template

## 🛡️ Beveiliging

### Script Execution
- **Sandboxed JavaScript**: Scripts draaien in geïsoleerde VM context
- **Module Whitelist**: Alleen toegestane modules beschikbaar
- **Timeout Protection**: Maximum execution time van 30 seconden
- **User Permissions**: Alleen geautoriseerde gebruikers kunnen scripts uitvoeren

### Database Access
- **Project Isolation**: Scripts hebben alleen toegang tot hun eigen project database
- **SQL Injection Protection**: Prepared statements en input validation
- **Audit Logging**: Volledige logging van alle database operations

## 📊 Monitoring

### Execution Logs
- Script output en errors
- Execution duration
- User who executed script
- Timestamp en context

### Analytics
- Script usage statistics
- Success/failure rates
- Performance metrics
- Popular templates

## 🔄 Migration Support

### Database Migrations
```sql
-- Migration: 001_add_user_profiles.sql
ALTER TABLE users ADD COLUMN profile_data JSONB DEFAULT '{}';
CREATE INDEX idx_users_profile ON users USING GIN(profile_data);

-- Rollback: 001_add_user_profiles_rollback.sql
ALTER TABLE users DROP COLUMN profile_data;
```

### Version Control
- Automatic migration versioning
- Rollback scripts
- Migration history tracking
- Dependency management

## 🎯 Use Cases

### Development
- Database schema setup
- Test data generation
- Function deployment
- API endpoint creation

### Production
- Database migrations
- Backup automation
- Monitoring setup
- Performance optimization

### Maintenance
- Data cleanup
- Schema updates
- Security patches
- System health checks

## 🤝 Contributing

### Adding Templates
1. Maak nieuwe script collection in `templates/` directory
2. Definieer script dependencies en volgorde
3. Test met verschillende project types
4. Documenteer gebruik en variabelen

### Plugin Development
1. Fork repository
2. Maak feature branch
3. Test met CloudBox development environment
4. Submit pull request

## 📞 Support

- **Documentation**: `/cloudbox/plugins/script-runner/docs/`
- **Issues**: GitHub Issues tracker
- **Community**: CloudBox Discord server
- **Email**: support@cloudbox.dev

## 📝 License

MIT License - Zie LICENSE file voor details.

---

**CloudBox Universal Script Runner** - Maak project setup zo eenvoudig als een SQL query! 🚀