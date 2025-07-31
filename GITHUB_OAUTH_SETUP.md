# GitHub OAuth Setup voor CloudBox

## 🎯 Waarom OAuth?

In plaats van een globale GitHub token gebruikt CloudBox nu **per-repository OAuth autorisatie**. Dit is veel veiliger en gebruiksvriendelijker:

✅ **Voordelen:**
- Gebruikers autoriseren alleen specifieke repositories
- Geen globale GitHub access tokens nodig
- Elke repository kan door verschillende gebruikers geautoriseerd worden
- Beter security model
- Test functionaliteit ingebouwd

## 🔧 GitHub OAuth App Setup

### Stap 1: Maak een GitHub OAuth App

1. Ga naar: https://github.com/settings/developers
2. Klik **"New OAuth App"**
3. Vul de gegevens in:

```
Application name: CloudBox Repository Analysis
Homepage URL: http://localhost:3000
Application description: CloudBox repository analysis and deployment platform
Authorization callback URL: http://localhost:8080/api/v1/github/oauth/callback
```

### Stap 2: Verkrijg Client Credentials

Na het aanmaken van de OAuth App:
1. Kopieer de **Client ID**
2. Genereer een **Client Secret**
3. Bewaar beide veilig

### Stap 3: Configureer CloudBox

Bewerk je `.env` bestand:

```bash
# GitHub OAuth Configuration
GITHUB_CLIENT_ID=your_client_id_here
GITHUB_CLIENT_SECRET=your_client_secret_here
GITHUB_REDIRECT_URL=http://localhost:8080/api/v1/github/oauth/callback
```

### Stap 4: Herstart de Services

```bash
docker-compose restart backend
```

## 🚀 Hoe het werkt

### Voor Gebruikers

1. **Repository Toevoegen**: Voeg een GitHub repository toe aan je project
2. **Test Toegang**: Klik "Test toegang" om te zien of autorisatie nodig is
3. **Autoriseren**: Klik "Autoriseer toegang" om OAuth flow te starten
4. **GitHub Autorisatie**: Autoriseer CloudBox in het GitHub popup venster
5. **Repository Analyse**: Nu kun je de repository analyseren met echte data!

### Voor Ontwikkelaars

**API Endpoints:**
```
POST /api/v1/projects/:id/github-repositories/:repo_id/authorize
GET  /api/v1/projects/:id/github-repositories/:repo_id/test-access
GET  /api/v1/github/oauth/callback
```

**OAuth Flow:**
1. Frontend roept `/authorize` endpoint aan
2. Backend genereert GitHub OAuth URL met state parameter
3. Gebruiker authoriseert in GitHub popup
4. GitHub redirect naar `/oauth/callback`
5. Backend wisselt code in voor access token
6. Token wordt opgeslagen per repository in database
7. Repository analyse gebruikt per-repository token

## 🔒 Security Features

- **State Parameter**: Voorkomt CSRF attacks
- **Scoped Permissions**: Alleen `public_repo` voor publieke repos, `repo` voor private repos
- **Token Encryptie**: Access tokens worden veilig opgeslagen in database
- **Per-Repository**: Elke repository heeft eigen autorisatie
- **Test Functionaliteit**: Verificatie dat toegang werkt

## 🧪 Testing

### Test Flow
1. Ga naar GitHub repositories pagina
2. Voeg een repository toe
3. Klik **"Test toegang"** → Krijg error (geen autorisatie)
4. Klik **"Autoriseer toegang"** → GitHub OAuth popup
5. Autoriseer CloudBox in GitHub
6. Klik **"Test toegang"** → Success! ✅
7. Klik **"Analyseer"** → Werkt nu met echte GitHub data! 🎉

### Verwachte Resultaten
- **Voor autorisatie**: "❌ Geen autorisatie - klik Autoriseren om toegang te verlenen"
- **Na autorisatie**: "✅ Repository toegang werkt! (username)"
- **Repository analyse**: Echte `package.json`, `Dockerfile`, etc. van GitHub

## 🔄 Fallback Strategie

Het systeem heeft een intelligente fallback:
1. **Eerst**: Probeer repository-specifieke OAuth token
2. **Fallback**: Gebruik globale `GITHUB_TOKEN` (indien aanwezig)
3. **Error**: Duidelijke melding als geen toegang mogelijk

## 📊 Database Schema

Nieuwe OAuth velden in `git_hub_repositories`:
```sql
access_token      TEXT        -- Encrypted OAuth token
token_expires_at  TIMESTAMP   -- Token expiration
refresh_token     TEXT        -- Refresh token (voor future use)
token_scopes      VARCHAR     -- Granted OAuth scopes
authorized_at     TIMESTAMP   -- When authorized
authorized_by     VARCHAR     -- GitHub username who authorized
```

## 🎯 Productie Deployment

Voor productie:
1. Update **Authorization callback URL** naar je productie domain
2. Update `GITHUB_REDIRECT_URL` in `.env`
3. Gebruik HTTPS endpoints
4. Implementeer token refresh logic (optional)

---

## 🎉 Resultaat

**Veel betere gebruikerservaring:**
- Geen globale tokens meer nodig
- Per-repository autorisatie
- Ingebouwde test functionaliteit
- Veilige OAuth flow
- Echte GitHub data voor repository analyse

**Nu kun je repository analyse testen met echte GitHub repositories!** 🚀