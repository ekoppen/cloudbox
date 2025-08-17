# CloudBox Plugin Marketplace & Management System

Een complete plugin marketplace geïntegreerd met GitHub voor veilige plugin distributie en beheer.

## 🌟 Overzicht

Het CloudBox Plugin Marketplace systeem biedt:
- **GitHub-gebaseerde marketplace** - Veilige plugin distributie via goedgekeurde repositories
- **Admin beheer interface** - Complete plugin management dashboard
- **Real-time installatie** - Geen frontend rebuilds vereist
- **Security-first approach** - Alleen pre-goedgekeurde plugins toegestaan
- **Comprehensive audit logging** - Volledige audit trail voor alle plugin operaties

## 🛡️ Security Features

### Approved Repository System
```json
{
  "approved_repositories": [
    "https://github.com/cloudbox/official-plugins",
    "https://github.com/cloudbox/community-plugins",
    "https://github.com/cloudbox/plugins"
  ]
}
```

### Plugin Validation
- **Repository validation** - Alleen goedgekeurde GitHub repositories
- **Plugin name validation** - Alphanumeric, dash, underscore only
- **Permission checking** - Admin/superadmin access required
- **Security scanning** - Automatic vulnerability detection
- **Signature verification** - Plugin integrity checking

### Audit Logging
Alle plugin operaties worden gelogd:
```json
{
  "user_id": 1,
  "user_email": "admin@cloudbox.com",
  "action": "install",
  "plugin_name": "cloudbox-script-runner",
  "old_status": "uninstalled",
  "new_status": "disabled",
  "ip_address": "192.168.1.100",
  "user_agent": "Mozilla/5.0...",
  "success": true,
  "error_msg": "",
  "created_at": "2024-08-17T12:00:00Z"
}
```

## 🎨 Frontend Interface

### Plugin Management Dashboard
Toegankelijk via `/dashboard/admin/plugins`:

#### Features
- **Plugin Statistics** - Total, enabled, disabled, errors
- **Plugin List** - Sorteerbaar overzicht van alle plugins
- **Status Management** - Enable/disable met één klik
- **Plugin Details** - Uitgebreide plugin informatie modal
- **Marketplace Integration** - "Browse Marketplace" knop voor nieuwe plugins

#### Plugin Actions
- **Enable/Disable** - Real-time status wijziging
- **Uninstall** - Veilige verwijdering met bevestiging
- **Reload** - Plugin configuratie herladen
- **Details** - Uitgebreide plugin informatie

### Marketplace Modal
Toegankelijk via "Browse Marketplace" knop:

#### Features
- **Search & Filter** - Zoeken op naam, beschrijving, tags
- **Tag Filtering** - Filter op categorieën (development, backup, analytics)
- **Plugin Cards** - Overzichtelijke plugin weergave met ratings
- **Plugin Details** - Uitgebreide informatie modal
- **Install Progress** - Real-time installatieproces tracking

#### Plugin Information
Elke plugin toont:
- **Basic Info** - Naam, versie, auteur, beschrijving
- **Statistics** - Downloads, stars, ratings
- **Security Status** - Verified badge, last updated
- **Permissions** - Vereiste machtigingen
- **Dependencies** - Benodigde afhankelijkheden
- **GitHub Link** - Directe link naar repository

## 🔧 Backend API

### Plugin Management Endpoints

#### List Plugins
```http
GET /api/v1/admin/plugins
Authorization: Bearer <jwt_token>
```

Response:
```json
{
  "success": true,
  "plugins": [
    {
      "name": "cloudbox-script-runner",
      "version": "1.0.0",
      "description": "Universal Script Runner for CloudBox",
      "author": "CloudBox Development Team",
      "type": "dashboard-plugin",
      "status": "enabled",
      "installed_at": "2024-08-16T12:00:00Z",
      "path": "./plugins/script-runner",
      "permissions": ["database:read", "database:write"],
      "ui": {
        "project_menu": {
          "title": "Scripts",
          "icon": "terminal",
          "path": "/dashboard/projects/{projectId}/scripts"
        }
      }
    }
  ]
}
```

#### Install Plugin
```http
POST /api/v1/admin/plugins/install
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "repository": "https://github.com/cloudbox/plugins",
  "version": "1.0.0",
  "project_id": 1
}
```

#### Enable/Disable Plugin
```http
POST /api/v1/admin/plugins/{pluginName}/enable
POST /api/v1/admin/plugins/{pluginName}/disable
Authorization: Bearer <jwt_token>
```

#### Uninstall Plugin
```http
DELETE /api/v1/admin/plugins/{pluginName}
Authorization: Bearer <jwt_token>
```

### Marketplace Endpoints

#### Get Marketplace
```http
GET /api/v1/admin/plugins/marketplace
Authorization: Bearer <jwt_token>
```

Response:
```json
{
  "success": true,
  "plugins": [
    {
      "name": "cloudbox-script-runner",
      "version": "1.0.0",
      "description": "Universal Script Runner for CloudBox - Database scripts en project setup",
      "author": "CloudBox Development Team",
      "repository": "https://github.com/cloudbox/plugins",
      "category": "development",
      "tags": ["database", "scripts", "automation"],
      "rating": 4.8,
      "downloads": 1250,
      "verified": true,
      "official": true,
      "last_updated": "2024-08-16T12:00:00Z",
      "permissions": [
        "database:read",
        "database:write",
        "functions:deploy",
        "webhooks:create",
        "projects:manage"
      ]
    }
  ]
}
```

#### Search Marketplace
```http
GET /api/v1/admin/plugins/marketplace/search?q=script&tags=development&official=true
Authorization: Bearer <jwt_token>
```

#### Get Plugin Details
```http
GET /api/v1/admin/plugins/marketplace/{pluginName}?repository=https://github.com/cloudbox/plugins
Authorization: Bearer <jwt_token>
```

### Security Endpoints

#### Get Approved Repositories
```http
GET /api/v1/admin/plugins/repositories
Authorization: Bearer <jwt_token>
```

#### Get Audit Logs
```http
GET /api/v1/admin/plugins/audit-logs?limit=100&plugin=script-runner
Authorization: Bearer <jwt_token>
```

#### Get Plugin Health
```http
GET /api/v1/admin/plugins/health?project_id=1
Authorization: Bearer <jwt_token>
```

## 📱 Frontend Integration

### Store Management
Plugin state wordt beheerd via Svelte stores:

```typescript
import { pluginManager } from '$lib/stores/plugins';

// Load plugins
const plugins = await pluginManager.loadPlugins();

// Enable plugin
await pluginManager.enablePlugin('plugin-name');

// Install from marketplace
await pluginManager.installPlugin('github.com/cloudbox/plugins', '1.0.0');

// Load marketplace
const marketplacePlugins = await pluginManager.loadMarketplace();
```

### Dynamic Navigation
Plugins integreren automatisch in de navigatie:

```javascript
// Project menu integration
{
  "ui": {
    "project_menu": {
      "title": "Scripts",
      "icon": "terminal",
      "path": "/dashboard/projects/{projectId}/scripts"
    }
  }
}
```

### Hot Loading
- **No frontend rebuilds** - Plugins laden zonder hercompilatie
- **Real-time updates** - UI updates onmiddellijk na plugin wijzigingen
- **Dynamic routes** - Plugin routes worden automatisch geregistreerd

## 🔄 Plugin Lifecycle

### Installation Flow
1. **Repository Validation** - Controleer approved repositories lijst
2. **Plugin Download** - Veilige download van GitHub repository
3. **Security Scan** - Automatische vulnerability detection
4. **Installation** - Plugin files extractie en validatie
5. **Database Setup** - Plugin database records aanmaken
6. **Status Update** - Plugin status naar "disabled" (veilig default)

### Enable/Disable Flow
1. **Permission Check** - Admin/superadmin verificatie
2. **Status Update** - Database status wijziging
3. **Route Registration** - API routes activeren/deactiveren
4. **UI Integration** - Frontend navigatie update
5. **Audit Log** - Volledige operatie logging

### Uninstall Flow
1. **Confirmation** - Gebruiker bevestiging vereist
2. **Status Check** - Plugin moet disabled zijn
3. **Cleanup** - Database records verwijdering
4. **File Removal** - Plugin files cleanup
5. **Navigation Update** - UI elements verwijdering

## 🏪 Available Plugins

### 1. CloudBox Script Runner
**Repository**: `https://github.com/cloudbox/plugins`
**Category**: Development Tools
**Status**: Official, Verified

**Features**:
- Universal SQL script execution
- JavaScript function deployment
- Project setup templates (Web App, AI Chat, E-commerce)
- Database migration management
- Audit logging en rollback support

**Templates**:
- **Web App Basic** - Authentication, sessions, CRUD
- **AI Chat App** - Conversations, users, message history
- **E-commerce Basic** - Products, cart, orders, payments

**UI Integration**: Project sidebar → Scripts

### 2. CloudBox Backup Manager *(Coming Soon)*
**Repository**: `https://github.com/cloudbox/official-plugins`
**Category**: Backup & Recovery
**Status**: Official, Verified

**Features**:
- Automated project backups
- Scheduled backup jobs
- Cloud storage integration
- Backup verification en restore

### 3. CloudBox Analytics *(Coming Soon)*
**Repository**: `https://github.com/cloudbox/community-plugins`
**Category**: Analytics & Reporting
**Status**: Community, Verified

**Features**:
- Advanced usage analytics
- Custom dashboards
- Performance monitoring
- Export capabilities

## 🔐 Security Best Practices

### For Plugin Developers
1. **Minimal Permissions** - Request alleen benodigde permissions
2. **Input Validation** - Valideer alle user inputs
3. **Output Sanitization** - Sanitize alle outputs
4. **Error Handling** - Graceful error handling
5. **Audit Logging** - Log belangrijke operaties

### For Administrators
1. **Review Permissions** - Controleer plugin permissions voor installatie
2. **Monitor Logs** - Controleer audit logs regelmatig
3. **Update Regularly** - Houd plugins up-to-date
4. **Test First** - Test nieuwe plugins in development environment
5. **Backup Before** - Maak backup voor plugin installatie

## 🚀 Development Workflow

### Plugin Development
1. **Create Repository** - GitHub repository in approved organizations
2. **Follow Standards** - Gebruik CloudBox plugin template
3. **Security Review** - Interne security audit
4. **Testing** - Comprehensive testing in CloudBox environment
5. **Documentation** - Complete gebruikersdocumentatie
6. **Approval** - Toevoegen aan approved repositories lijst

### Quality Assurance
- **Automated Testing** - CI/CD pipeline voor alle plugins
- **Security Scanning** - Automatic vulnerability detection
- **Performance Testing** - Impact assessment op CloudBox performance
- **Compatibility Testing** - Multi-version CloudBox compatibility

## 📊 Analytics & Monitoring

### Plugin Usage Metrics
- **Installation Count** - Aantal installaties per plugin
- **Active Users** - Aantal actieve gebruikers
- **Success Rate** - Installatie/operatie success rates
- **Performance Impact** - Resource usage monitoring

### Security Monitoring
- **Failed Attempts** - Unauthorized access attempts
- **Permission Violations** - Invalid permission usage
- **Vulnerability Alerts** - Security issue notifications
- **Audit Trail Analysis** - Suspicious activity detection

## 🔗 Integration Examples

### Script Runner Integration
```javascript
// Project sidebar - automatisch geregistreerd
{
  "ui": {
    "project_menu": {
      "title": "Scripts",
      "icon": "terminal", 
      "path": "/dashboard/projects/{projectId}/scripts"
    }
  }
}
```

### Custom Plugin Integration
```svelte
<!-- Plugin component -->
<script>
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  
  $: projectId = $page.params.id;
  
  // Plugin logic hier
</script>

<div class="plugin-container">
  <!-- Plugin UI -->
</div>
```

## 🔧 Configuration

### Environment Variables
```env
# Plugin security settings
PLUGIN_SECURITY_SCAN=true
PLUGIN_EXECUTION_TIMEOUT=30000
PLUGIN_MAX_MEMORY=512MB

# Marketplace settings  
MARKETPLACE_CACHE_TTL=3600
MARKETPLACE_VERIFY_SIGNATURES=true

# Audit settings
AUDIT_LOG_RETENTION_DAYS=90
AUDIT_LOG_LEVEL=INFO
```

### Plugin Configuration
```json
{
  "name": "your-plugin",
  "version": "1.0.0",
  "security": {
    "sandbox": true,
    "timeout": 30000,
    "memory_limit": "256MB"
  },
  "marketplace": {
    "featured": false,
    "category": "development",
    "tags": ["database", "automation"]
  }
}
```

---

Het CloudBox Plugin Marketplace systeem biedt een **enterprise-grade plugin management oplossing** met GitHub integration, comprehensive security, en een professionele gebruikerservaring. 🚀