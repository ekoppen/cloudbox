# Repository Analyse Functionaliteit

## 🎯 Overzicht

CloudBox heeft nu uitgebreide repository analyse functionaliteit die automatisch GitHub repositories analyseert en verschillende installatie opties genereert voor deployments.

## ✨ Nieuwe Features

### 🔍 Automatische Repository Analyse
- **Project Type Detectie**: React, Vue, Angular, Next.js, Nuxt, Svelte, Photo Portfolio, etc.
- **Framework Detectie**: Vite, Webpack, Create React App, Angular CLI
- **Package Manager Detectie**: npm, yarn, pnpm met intelligente fallbacks
- **Docker Support**: Automatische detectie van Dockerfile
- **Environment Variabelen**: Parsing van .env.example bestanden

### 📦 Intelligente Install Options
- **Multiple Installation Methods**: npm, yarn, pnpm, Docker
- **Recommended Options**: AI bepaalt de beste installatie methode
- **Command Generation**: Automatische generatie van install, build, start commands
- **Port Detection**: Automatische detectie van applicatie poorten

### 💡 Smart Insights & Warnings
- **Framework-specific Tips**: React (REACT_APP_), Vue (VITE_), etc.
- **Performance Recommendations**: Build tijd schattingen
- **Deployment Warnings**: Port configuratie, environment setup
- **Complexity Scoring**: 1-10 schaal voor deployment moeilijkheid

## 🚀 Hoe het werkt

### Backend (Go)
```
1. Repository URL wordt geanalyseerd
2. GitHub API wordt gebruikt om bestanden op te halen
3. package.json, Dockerfile, README.md etc. worden geparst
4. Install options worden gegenereerd op basis van project type
5. Analyse wordt opgeslagen in PostgreSQL database
6. Re-analyse is mogelijk voor updates
```

### Frontend (Svelte)
```
1. GitHub repositories pagina toont nieuwe "Analyseer" knoppen
2. Repository analyse modal toont uitgebreide informatie
3. Install options component is herbruikbaar
4. Deployment integratie komt binnenkort
```

## 📋 API Endpoints

### Nieuwe Endpoints
- `GET /api/v1/projects/:id/github-repositories/:repo_id/analysis` - Haal opgeslagen analyse op
- `POST /api/v1/projects/:id/github-repositories/:repo_id/analyze` - Analyseer en sla op
- `POST /api/v1/projects/:id/github-repositories/:repo_id/reanalyze` - Heranalyseer repository

### Database Schema
```sql
-- repository_analyses tabel
- github_repository_id (unique)
- project_type, framework, language, package_manager
- install_options (JSON array met verschillende methoden)
- insights, warnings (JSON arrays)
- complexity score (1-10)
- performance metrics
```

## 🎨 Frontend Components

### Nieuwe Components
- `InstallOptions.svelte` - Herbruikbare component voor installatie opties
- Repository analyse modal in GitHub pagina
- Visual indicators voor project types (⚛️ React, 💚 Vue, etc.)

### UI Features
- **Project Type Icons**: Visuele indicatoren voor elk framework
- **Complexity Badges**: Kleurgecodeerde complexiteit (groen/geel/rood)
- **Recommended Tags**: Duidelijke aanbevelingen
- **Interactive Selection**: Klikbare install options
- **Loading States**: Smooth UX tijdens analyse

## 📊 Analyse Voorbeelden

### React Project
```json
{
  "project_type": "react",
  "framework": "vite",
  "language": "typescript",
  "package_manager": "npm",
  "install_options": [
    {
      "name": "npm",
      "command": "npm install",
      "build_command": "npm run build",
      "start_command": "npm start",
      "is_recommended": true,
      "description": "Standard npm installation"
    }
  ],
  "insights": [
    "React application detected - make sure to set REACT_APP_ environment variables",
    "Vite detected - very fast build times expected"
  ],
  "complexity": 3
}
```

### Docker Project
```json
{
  "project_type": "nextjs",
  "has_docker": true,
  "install_options": [
    {
      "name": "docker",
      "command": "docker build -t app .",
      "start_command": "docker run -p 3000:3000 app",
      "is_recommended": true,
      "description": "Deploy using Docker container (recommended for production)"
    },
    {
      "name": "npm",
      "command": "npm install",
      "is_recommended": false,
      "description": "Alternative npm installation"
    }
  ]
}
```

## 🔄 Workflow

### Voor Developers
1. **Repository Toevoegen**: Voeg GitHub repository toe aan project
2. **Automatische Analyse**: Klik "Analyseer" voor intelligente detectie
3. **Review Options**: Bekijk verschillende installatie methoden
4. **Select & Deploy**: Kies beste optie voor deployment

### Voor CloudBox
1. **Pattern Recognition**: Leer van gebruiker keuzes
2. **Recommendation Engine**: Verbeter aanbevelingen over tijd
3. **Template Generation**: Genereer deployment templates
4. **Performance Tracking**: Monitor success rates per optie

## 🔮 Toekomstige Uitbreidingen

### Geplande Features
- **Template Integration**: Genereer deployment templates uit analyse
- **Performance Monitoring**: Track werkelijke vs geschatte build tijden
- **Team Recommendations**: Leer van team deployment patterns
- **Advanced Parsing**: Support voor meer project types (Python, Go, etc.)
- **CI/CD Integration**: Automatische re-analyse bij repository changes
- **Cost Estimation**: Schat server kosten op basis van project complexiteit

### Deployment Integration
- **Smart Defaults**: Pre-fill deployment forms met analyse data
- **One-Click Deploy**: Direct deployen met aanbevolen configuratie
- **Environment Suggestions**: Automatische environment variable setup
- **Scaling Recommendations**: Server size aanbevelingen

## 🧪 Testing

### Frontend Testing
```bash
# Start development server
npm run dev

# Test repository analyse functionaliteit
1. Ga naar GitHub repositories pagina
2. Klik "Analyseer" bij een repository
3. Bekijk analyse resultaten in modal
4. Test heranalyseer functionaliteit
```

### Backend Testing
```bash
# Test analyse endpoints
curl -X POST "http://localhost:8080/api/v1/projects/1/github-repositories/1/analyze" \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"repo_url": "https://github.com/user/repo", "branch": "main"}'
```

---

🎉 **Repository Analyse maakt deployment eenvoudiger dan ooit!**