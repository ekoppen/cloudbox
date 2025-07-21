# 📄 Product Requirements Document (PRD) – CloudBox

## 1. Productnaam
**CloudBox**

## 2. Overzicht
CloudBox is een lokale, zelf-hostbare Backend-as-a-Service (BaaS) oplossing. Het biedt ontwikkelaars een snelle, veilige en gebruiksvriendelijke manier om moderne applicaties te bouwen met behulp van krachtige API’s, CORS-configuraties, en SDK’s. CloudBox wordt geleverd met een consistente UI, API per project, webapp-deploymentmogelijkheden en backup/restore functionaliteit.

## 3. Doelstellingen
- Ontwikkelaars voorzien van een lokaal alternatief voor Appwrite/Supabase.
- Eenvoudige, veilige setup en beheer van backendfuncties via een duidelijke UI.
- Volledige controle over data, API-toegang en deployment.
- Ondersteuning voor SDK-integratie, zodat agents en apps eenvoudig kunnen communiceren met de backend.

## 4. Doelgroep
- Freelance developers en kleine teams die BaaS lokaal willen draaien.
- Organisaties met compliance-eisen die geen cloudgebaseerde oplossingen mogen gebruiken.
- AI/agent ontwikkelaars die een betrouwbare backend nodig hebben met eenvoudige API-koppelingen.

## 5. Belangrijkste Features

### 5.1 Project-based API configuratie
- Ieder project binnen CloudBox heeft zijn eigen API-namespace.
- Endpointdocumentatie automatisch gegenereerd.
- API-keys en tokens per project instelbaar.

### 5.2 CORS configuratie per project
- Volledige controle over toegestane origins, headers, en methoden.
- Presets voor snelle setup (zoals ‘alleen localhost’, ‘alle subdomeinen’, etc).

### 5.3 Security-first
- JWT en OAuth2 support.
- Rate limiting, IP-blocking en rolgebaseerde toegangscontrole.
- Encryptie van data at rest en in transit.

### 5.4 Web App Deployment
- Webapplicaties kunnen per project gedeployed worden, vergelijkbaar met Appwrite’s hosting.
- Ondersteuning voor statische bestanden (HTML/CSS/JS).
- Mogelijkheid tot deployment via CLI of UI.

### 5.5 Back-up & Herstel
- Automatische back-up schema’s per project.
- Mogelijkheid tot handmatige exports en imports.
- Versiebeheer per back-up.

### 5.6 SDK-ondersteuning
- SDK’s in o.a. JavaScript, Python, Go, en TypeScript.
- SDK’s gericht op interactie met API’s en authenticatie.
- Agents (bijv. AI/LLM-agents) kunnen eenvoudig inpluggen.

## 6. Niet-functionele vereisten
- **Performance:** Reactieve UI, snelle API-respons (idealiter < 100ms lokaal).
- **Schaalbaarheid:** Draait lokaal, maar uitbreidbaar via Docker of Kubernetes.
- **Gebruiksvriendelijkheid:** UI consistent en intuïtief; donker/licht thema beschikbaar.
- **Documentatie:** Uitgebreide handleidingen en auto-generated API docs.

## 7. Technische specificaties
- **Taal/stack:** Backend in Go of Node.js, frontend in React of Svelte.
- **Database:** PostgreSQL met Prisma ORM of alternatief.
- **Hosting:** Dockerized voor eenvoudige installatie.
- **Deployment CLI:** Voor projectbeheer en CI/CD-integratie.

## 8. MVP Scope
- Projectbeheer
- API configuratie en testen
- CORS-instellingen
- Basis authenticatie (JWT)
- Webapp hosting
- SDK (alleen JavaScript)
- UI met projectdashboard

## 9. Toekomstige uitbreidingen (post-MVP)
- OAuth-integratie
- Multi-region back-ups
- Admin CLI
- Realtime database (à la Firebase)
- Teambeheer en samenwerking

## 10. Deadline / Planning
- **MVP:** binnen 3 maanden
- **Interne testfase:** 2 weken na MVP
- **Open beta:** 1 maand na testfase
