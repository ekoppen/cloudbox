# 🧪 CloudBox Test App - Snelle Start Handleiding

## Stap 1: Test App Openen
1. Zorg dat je CloudBox backend draait: `docker-compose up`
2. Open in je browser: **http://localhost:8080/testapp/index.html**

## Stap 2: Project Configureren
Vul in de "Project Configuratie" sectie in:
- **Backend URL**: `http://localhost:8080` ✅ (al ingevuld)
- **Project Slug**: `wouterkoppen-com` (of je eigen project slug)
- **API Key**: Haal deze uit je CloudBox project dashboard

### API Key verkrijgen:
1. Ga naar je project in CloudBox dashboard
2. Klik op "API Keys" of "Project Instellingen"
3. Genereer een nieuwe API key
4. Kopieer en plak in de test app

## Stap 3: Test de Verbinding
Klik op **"Test Verbinding"** - je zou ✅ Connected moeten zien.

## Stap 4: Test Scenario's

### 🔐 Authentication (Start hier!)
1. Ga naar de **"Authentication"** tab
2. **Registreer een gebruiker**:
   - Email: `test@example.com`
   - Password: `testpassword123`
   - Name: `Test User`
3. **Log in** met dezelfde gegevens
4. ✅ Je krijgt een JWT token terug

### 📊 Database Testing
1. Ga naar **"Database"** tab
2. **Create Data**:
   - Table: `users`
   - JSON: `{"name": "John Doe", "email": "john@example.com", "age": 30}`
3. **Read Data**: Haal alle records op uit `users` tabel

### 📁 Storage Buckets
1. Ga naar **"Storage Buckets"** tab
2. **Create Bucket**: naam `images`
3. **Upload File**: Selecteer een afbeelding en upload naar `images` bucket
4. **List Files**: Bekijk alle bestanden in de bucket

### ⚡ Functions
1. Ga naar **"Functions"** tab
2. Function Name: `hello-world`
3. Parameters: `{"name": "World"}`
4. Execute de functie

### ✉️ Messaging
1. **Email**: Verstuur een test email
2. **Push**: Verstuur een push notificatie

## 🔍 Response Log
- Alle API calls worden gelogd onderaan
- **Groen** = Success ✅
- **Rood** = Error ❌
- **Blauw** = Info ℹ️

## ⚠️ Troubleshooting

**Verbinding lukt niet?**
- Check of CloudBox backend draait
- Controleer de Backend URL

**Authentication errors?**
- Zorg dat je een geldige API key hebt
- Check of project slug correct is

**Database/Storage errors?**
- Zorg dat je eerst ingelogd bent (JWT token)
- Check de Response Log voor details

## 💡 Tips
1. Begin altijd met Authentication
2. Check de Response Log bij problemen
3. Test één feature tegelijk
4. Je browser onthoudt de configuratie

**Happy Testing!** 🎉