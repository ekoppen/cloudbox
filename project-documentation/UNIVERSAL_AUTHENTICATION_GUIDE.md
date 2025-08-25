# CloudBox Universal Authentication Guide
## Future-Proof Authentication for All Client Applications

### Quick Start

**For any new client application:**
```bash
# 1. Auto-detect and setup CORS for any framework
node scripts/setup-universal-cors.js

# 2. Install CloudBox SDK
npm install @ekoppen/cloudbox-sdk

# 3. Use framework helper (automatically generated)
import client from './src/cloudbox'; // or your generated helper file
```

That's it! Your app now works with CloudBox authentication.

## Framework-Specific Quick Setup

### React Applications

**Setup:**
```bash
node scripts/setup-universal-cors.js --framework react --create-helpers
```

**Usage:**
```javascript
// src/cloudbox.js (auto-generated)
import client from './src/cloudbox';
import { useCloudBoxAuth } from '@ekoppen/cloudbox-sdk/helpers';

function App() {
  const { user, login, logout, loading } = useCloudBoxAuth(client);
  
  if (loading) return <div>Loading...</div>;
  
  return user ? (
    <div>Welcome {user.name}! <button onClick={logout}>Logout</button></div>
  ) : (
    <LoginForm onLogin={login} />
  );
}
```

### Vue.js Applications

**Setup:**
```bash
node scripts/setup-universal-cors.js --framework vue --create-helpers
```

**Usage:**
```javascript
// src/composables/useCloudBox.js (auto-generated)
import { useCloudBoxAuth } from '@ekoppen/cloudbox-sdk/helpers';
import client from '../cloudbox';

export default function useCloudBox() {
  return useCloudBoxAuth(client);
}

// In your component
<template>
  <div v-if="loading">Loading...</div>
  <div v-else-if="user">
    Welcome {{ user.name }}!
    <button @click="logout">Logout</button>
  </div>
  <LoginForm v-else @login="login" />
</template>

<script setup>
import useCloudBox from '@/composables/useCloudBox';
const { user, login, logout, loading } = useCloudBox();
</script>
```

### Angular Applications

**Setup:**
```bash
node scripts/setup-universal-cors.js --framework angular --create-helpers
```

**Usage:**
```typescript
// src/app/services/cloudbox.service.ts (auto-generated)
import { CloudBoxService } from './services/cloudbox.service';

@Component({
  selector: 'app-root',
  template: `
    <div *ngIf="loading">Loading...</div>
    <div *ngIf="user">
      Welcome {{ user.name }}!
      <button (click)="logout()">Logout</button>
    </div>
    <app-login *ngIf="!user && !loading" (login)="login($event)"></app-login>
  `
})
export class AppComponent {
  user$ = this.cloudbox.user$;
  loading$ = this.cloudbox.loading$;
  
  constructor(private cloudbox: CloudBoxService) {}
  
  async login(credentials) {
    await this.cloudbox.login(credentials);
  }
  
  async logout() {
    await this.cloudbox.logout();
  }
}
```

### Svelte Applications

**Setup:**
```bash
node scripts/setup-universal-cors.js --framework svelte --create-helpers
```

**Usage:**
```javascript
// src/stores/auth.js (auto-generated)
import { authStore } from '../stores/auth';

// In your component
<script>
  import { authStore } from '../stores/auth';
  
  $: ({ user, loading, login, logout } = authStore);
</script>

{#if $loading}
  <div>Loading...</div>
{:else if $user}
  <div>
    Welcome {$user.name}!
    <button on:click={logout}>Logout</button>
  </div>
{:else}
  <LoginForm on:login={e => login(e.detail)} />
{/if}
```

### Vanilla JavaScript

**Setup:**
```bash
node scripts/setup-universal-cors.js --framework vanilla --create-helpers
```

**Usage:**
```javascript
// js/cloudbox.js (auto-generated)
import client from './js/cloudbox.js';

// Simple authentication
async function handleLogin(email, password) {
  try {
    const { user, token } = await client.auth.login({ email, password });
    localStorage.setItem('auth_token', token);
    client.setAuthToken(token);
    showDashboard(user);
  } catch (error) {
    console.error('Login failed:', error.message);
    showError(error.suggestions || ['Check credentials and try again']);
  }
}
```

## Authentication Features

### Automatic Header Fallback

The SDK automatically tries different authentication headers:

1. **Session-Token** (primary for project endpoints)
2. **session-token** (lowercase variant)
3. **X-Session-Token** (prefixed variant)
4. **Authorization** (fallback for compatibility)

If one header fails due to CORS, it automatically tries the next one.

### Enhanced Error Handling

```javascript
// Automatic CORS error detection and helpful messages
try {
  const user = await client.auth.me();
} catch (error) {
  if (error.name === 'CORSConfigurationError') {
    console.log('CORS issue detected!');
    console.log('Quick fix:', error.corsInfo.quickFixes[0].command);
    // Shows: node scripts/setup-universal-cors.js --origin="http://localhost:3000"
  }
}
```

### Debug Mode

```javascript
// Enable debug mode in development
if (process.env.NODE_ENV === 'development') {
  client.enableDebugMode();
  
  // Access debug tools in browser console
  window.cloudboxDebug.testAuth(); // Test different auth headers
  window.cloudboxDebug.corsInfo(); // Show CORS configuration
}
```

## Environment Configuration

### Development Environment

**Automatic Setup:**
```bash
# Creates development-friendly CORS configuration
node scripts/setup-universal-cors.js --environment development --update-global
```

**Manual Setup (.env):**
```bash
# CloudBox Backend .env
CORS_ORIGINS=http://localhost:*,https://localhost:*
CORS_HEADERS=*
CORS_DEBUG=true
NODE_ENV=development
```

### Staging Environment

```bash
# Staging with localhost testing support
CORS_ORIGINS=https://*.staging.example.com,http://localhost:*
CORS_HEADERS=Content-Type,Authorization,Session-Token,X-API-Key
NODE_ENV=staging
```

### Production Environment

```bash
# Secure production setup
CORS_ORIGINS=https://app.example.com,https://admin.example.com
CORS_HEADERS=Content-Type,Authorization,Session-Token,X-API-Key
NODE_ENV=production
```

## Troubleshooting

### Common Issues and Solutions

#### 1. CORS Blocked Error

**Error:** `Access to fetch at 'http://localhost:8080' blocked by CORS policy`

**Solutions:**
```bash
# Automatic fix (recommended)
node scripts/setup-universal-cors.js --origin="http://localhost:3000"

# Manual fix - add to CloudBox .env
echo "CORS_ORIGINS=http://localhost:3000,http://localhost:*" >> .env

# Restart backend
docker-compose restart backend
```

#### 2. Authentication Header Not Accepted

**Error:** `Invalid or missing authentication header`

**Debug:**
```javascript
// Test which headers work
const results = await client.testAuthHeaders();
console.log('Working headers:', results);
```

**Solution:**
```javascript
// Force specific authentication mode
const client = new CloudBoxClient({
  projectId: 2,
  apiKey: 'your-key',
  authMode: 'project', // or 'admin'
  authConfig: {
    strategies: {
      project: {
        primary: 'Session-Token',
        fallbacks: ['X-Session-Token', 'Authorization']
      }
    }
  }
});
```

#### 3. Port Detection Issues

**Error:** Setup script can't detect your app's port

**Solution:**
```bash
# Specify port manually
node scripts/setup-universal-cors.js --port 3001

# Or specify full origin
node scripts/setup-universal-cors.js --origin "http://localhost:3001"
```

#### 4. Framework Not Detected

**Error:** `No supported framework detected`

**Solution:**
```bash
# Force framework detection
node scripts/setup-universal-cors.js --framework react
node scripts/setup-universal-cors.js --framework vue
node scripts/setup-universal-cors.js --framework angular
```

### Health Check

```javascript
// Comprehensive system health check
import { DevUtils } from '@ekoppen/cloudbox-sdk/helpers';

const healthCheck = DevUtils.createHealthCheck(client);
const results = await healthCheck();

console.log('System Status:', results);
/*
{
  connection: true,
  authentication: true,
  cors: true,
  timestamp: "2024-01-15T10:30:00.000Z"
}
*/
```

## Migration from Existing Apps

### From PhotoPortfolio

**Existing PhotoPortfolio apps work automatically!** The new system is backward compatible.

**Optional upgrade:**
```bash
# Upgrade to universal system with enhanced error handling
node scripts/setup-universal-cors.js --framework react --create-helpers

# Replace your existing CloudBox client with the new helper
import client from './src/cloudbox'; // auto-generated with enhancements
```

### From Manual CORS Configuration

**Before:**
```javascript
// Manual header management
const headers = {
  'Session-Token': token,
  'X-API-Key': apiKey
};

try {
  const response = await fetch(url, { headers });
} catch (error) {
  // Manual error handling
  console.error('Request failed:', error);
}
```

**After:**
```javascript
// Automatic header management and error handling
const client = new CloudBoxClient({ projectId, apiKey });
client.setAuthToken(token);

try {
  const data = await client.auth.me();
} catch (error) {
  // Automatic CORS troubleshooting
  if (error.suggestions) {
    console.log('Quick fix:', error.suggestions);
  }
}
```

## Advanced Configuration

### Custom Authentication Strategies

```javascript
const client = new CloudBoxClient({
  projectId: 2,
  apiKey: 'your-key',
  authConfig: {
    strategies: {
      project: {
        primary: 'Custom-Token',
        fallbacks: ['Session-Token', 'Authorization'],
        transform: (token) => `CustomFormat ${token}`
      }
    }
  }
});
```

### Custom CORS Error Handling

```javascript
import { CORSErrorHandler } from '@ekoppen/cloudbox-sdk/helpers';

const client = new CloudBoxClient({
  projectId: 2,
  apiKey: 'your-key',
  authConfig: {
    corsErrorHandler: (corsInfo) => {
      // Custom error handling
      showNotification('CORS Error', corsInfo.suggestions[0]);
      trackError('cors_configuration_error', corsInfo);
    }
  }
});
```

### Environment-Specific Configuration

```javascript
// config/environments.js
const environments = {
  development: {
    endpoint: 'http://localhost:8080',
    debugMode: true,
    autoRetry: true
  },
  production: {
    endpoint: 'https://api.cloudbox.com',
    debugMode: false,
    autoRetry: false
  }
};

const config = environments[process.env.NODE_ENV] || environments.development;
const client = new CloudBoxClient({ ...baseConfig, ...config });
```

## API Reference

### Setup Script Options

```bash
node scripts/setup-universal-cors.js [options]

Options:
  --framework <name>       Force framework (react, vue, angular, svelte, vanilla)
  --port <port>           Override port detection
  --origin <origin>       Full origin URL (overrides port)
  --project-id <id>       CloudBox project ID (default: auto-detect)
  --cloudbox-url <url>    CloudBox backend URL (default: http://localhost:8080)
  --environment <env>     Environment (development, staging, production)
  --create-helpers        Generate framework-specific helper files
  --update-global         Also update global CORS configuration
  --dry-run              Show what would be done without making changes
  --verbose              Show detailed output
```

### SDK Configuration Options

```typescript
interface CloudBoxConfig {
  projectId: number;
  apiKey: string;
  endpoint?: string;
  authMode?: 'admin' | 'project';
  authConfig?: {
    strategies?: Record<string, AuthHeaderStrategy>;
    corsErrorHandler?: (error: CORSErrorInfo) => void;
    debug?: boolean;
  };
}
```

### Framework Helper Methods

```javascript
// Available in @ekoppen/cloudbox-sdk/helpers
import { 
  FrameworkClientFactory,
  CORSErrorHandler,
  AuthStateManager,
  DevUtils 
} from '@ekoppen/cloudbox-sdk/helpers';

// Create framework-optimized clients
const reactClient = FrameworkClientFactory.createReactClient(config);
const vueClient = FrameworkClientFactory.createVueClient(config);

// Debug utilities
DevUtils.enableDebugMode(client);
const healthCheck = DevUtils.createHealthCheck(client);
```

## Support and Resources

### Getting Help

1. **Documentation**: https://docs.cloudbox.dev
2. **GitHub Issues**: https://github.com/cloudbox/issues
3. **Discord Community**: https://discord.gg/cloudbox
4. **Email Support**: support@cloudbox.dev

### Quick Commands Reference

```bash
# Setup any framework automatically
node scripts/setup-universal-cors.js

# Debug CORS issues
node scripts/setup-universal-cors.js --dry-run --verbose

# Production deployment
node scripts/setup-universal-cors.js --environment production --origin https://yourdomain.com

# Test existing setup
node -e "require('@ekoppen/cloudbox-sdk/helpers').DevUtils.createHealthCheck(client)()"
```

This universal authentication system ensures that **any** client application you build will work seamlessly with CloudBox, regardless of framework, port, or deployment environment.