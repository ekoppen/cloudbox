# CloudBox BaaS Test App

Een uitgebreide test applicatie voor het testen van alle CloudBox Backend-as-a-Service functionaliteiten.

## 🚀 Hoe te gebruiken

### 1. Test App Openen
Open de test app in je browser:
```
http://localhost:8080/testapp/index.html
```

### 2. Project Configuratie
In de "Project Configuratie" sectie vul je in:

- **Backend URL**: `http://localhost:8080` (standaard ingevuld)
- **Project Slug**: De slug van je project (bijvoorbeeld: `wouterkoppen-com`)
- **API Key**: Je project API key (krijg je in de CloudBox dashboard)

### 3. Verbinding Testen
Klik op "Test Verbinding" om te controleren of de backend bereikbaar is.

## 📊 Beschikbare Test Features

### Database Testing
Test je database API's:

**Create Data:**
- Voer een tabelnaam in (bijv. `users`)
- Voeg JSON data toe:
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "age": 30
}
```

**Read Data:**
- Voer dezelfde tabelnaam in
- Klik "Get All Records" om alle data op te halen

### Storage Buckets
Test file storage functionaliteit:

**Bucket Management:**
- Maak nieuwe buckets aan (bijv. `images`, `documents`)
- Bekijk alle beschikbare buckets

**File Management:**
- Selecteer een bestand van je computer
- Kies een target bucket
- Upload het bestand
- Bekijk alle bestanden in een bucket

### Authentication
Test gebruikers authenticatie:

**User Registration:**
- Email, wachtwoord en naam invoeren
- Gebruiker wordt automatisch geregistreerd

**User Login:**
- Inloggen met email en wachtwoord
- Krijg een JWT token terug voor verdere API calls

### Functions
Test serverless functions:

- Voer een function naam in (bijv. `hello-world`)
- Optioneel: voeg JSON parameters toe
- Voer de functie uit en bekijk de response

### Messaging
Test email en push notificaties:

**Email:**
- Verstuur test emails naar elk email adres
- Configureer onderwerp en bericht

**Push Notifications:**
- Verstuur push notificaties naar gebruikers
- Target kan een specifieke user ID zijn of `all`

## 🔧 API Key Verkrijgen

### Stap 1: Ga naar je Project Settings
1. Open CloudBox dashboard
2. Ga naar je project
3. Klik op "Project Instellingen" of "API Keys"

### Stap 2: Genereer API Key
1. Klik op "Nieuwe API Key"
2. Kies de juiste permissies
3. Kopieer de gegenereerde key

### Stap 3: Gebruik in Test App
Plak de API key in het "API Key" veld van de test app.

## 📝 Response Log

Alle API calls worden gelogd in de "Response Log" sectie onderaan:

- **Blauwe logs**: Informatie over requests
- **Groene logs**: Succesvolle responses
- **Rode logs**: Errors en failures
- **Gele logs**: Waarschuwingen

## 🧪 Test Scenarios

### Basis Test Flow
1. **Configuratie**: Vul project slug en API key in
2. **Verbinding**: Test de verbinding
3. **Authentication**: Registreer en log in met een test gebruiker
4. **Database**: Maak wat test data aan en haal het op
5. **Storage**: Maak een bucket en upload een bestand
6. **Functions**: Test een eenvoudige functie
7. **Messaging**: Verstuur een test email

### Geavanceerde Tests
- Test verschillende file types in storage
- Test batch database operaties
- Test error handling met ongeldige data
- Test rate limiting en API quota's

## 🔍 Troubleshooting

### Verbinding Problemen
- Controleer of de CloudBox backend draait (`docker-compose up`)
- Controleer de Backend URL (standaard: `http://localhost:8080`)

### Authentication Errors
- Zorg dat je een geldige API key hebt
- Check of je project slug correct is
- Kijk in de Response Log voor specifieke error messages

### CORS Issues
Als je CORS errors krijgt, check de backend configuratie voor allowed origins.

## 📚 API Endpoints

De test app gebruikt de volgende CloudBox API endpoints:

- `GET /api/v1/health` - Health check
- `POST /api/v1/auth/register` - Gebruiker registratie
- `POST /api/v1/auth/login` - Gebruiker login
- `GET/POST /p/{slug}/api/data/{table}` - Database operaties
- `GET/POST /p/{slug}/api/storage/buckets` - Bucket management
- `POST /p/{slug}/api/storage/buckets/{bucket}/files` - File upload
- `POST /p/{slug}/api/functions/{name}` - Function execution
- `POST /p/{slug}/api/messaging/email` - Email verzenden
- `POST /p/{slug}/api/messaging/push` - Push notificaties

## 💡 Tips

1. **Start Simpel**: Begin met de health check en authentication
2. **Check Logs**: Gebruik altijd de Response Log om errors te debuggen
3. **Test Incrementeel**: Test één feature tegelijk
4. **Save Settings**: Je browser onthoudt de configuratie automatisch
5. **Clear Log**: Gebruik "Clear Log" om de log leeg te maken tussen tests

Happy testing! 🎉