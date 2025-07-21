# 🔒 CloudBox API Security Model

## Authentication vs Authorization

CloudBox gebruikt een tweelaags beveiligingsmodel:

### 🔓 Public Endpoints (Geen API Key vereist)
Deze endpoints zijn toegankelijk zonder API key:

- **Health Check**: `GET /health`
- **User Registration**: `POST /api/v1/auth/register`
- **User Login**: `POST /api/v1/auth/login`
- **Token Refresh**: `POST /api/v1/auth/refresh`

**Waarom geen API key?**
- Gebruikers moeten zich kunnen registreren zonder bestaande toegang
- Login is nodig om JWT tokens te verkrijgen
- Health check is voor system monitoring

### 🔐 Project Endpoints (API Key vereist)
Deze endpoints vereisen een geldige API key:

- **Project Data**: `/p/{project-slug}/api/data/*`
- **Storage Buckets**: `/p/{project-slug}/api/storage/*`
- **Functions**: `/p/{project-slug}/api/functions/*`
- **Messaging**: `/p/{project-slug}/api/messaging/*`

**Waarom API key verplicht?**
- Beschermt project-specifieke data
- Voorkomt ongeautoriseerde toegang tot storage
- Controleert gebruik en facturering
- Isolatie tussen verschillende projecten

## Security Flow

1. **User Registration/Login** → Verkrijg JWT token (geen API key nodig)
2. **Generate API Key** → Via dashboard of API met JWT token
3. **Access Project Data** → API key + JWT token voor volledige toegang

## API Key Management

### API Key verkrijgen:
1. Login in CloudBox dashboard
2. Ga naar Project Settings
3. Klik "Generate API Key"
4. Kopieer en bewaar de key veilig

### API Key usage:
```javascript
headers: {
  'Content-Type': 'application/json',
  'X-API-Key': 'your-api-key-here',
  'Authorization': 'Bearer your-jwt-token'
}
```

## Test App Security Demo

De test app demonstreert dit security model:

✅ **Zonder API Key**: Authentication werkt
❌ **Zonder API Key**: Project data endpoints geven "API key required" error
✅ **Met API Key**: Volledige toegang tot project data

Dit is **correct beveiligingsgedrag** - niet een bug!