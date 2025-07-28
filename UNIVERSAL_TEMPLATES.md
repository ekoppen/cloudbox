# CloudBox Universal Project Templates

## Overview

CloudBox biedt een universeel template systeem vergelijkbaar met Appwrite, Supabase en Firebase. Developers kunnen in minuten full-stack applicaties creëren met automatische backend configuratie.

## 🚀 Quick Start - Elk Project Type

### Global CLI Installation

```bash
npm install -g create-cloudbox-app

# Of gebruik direct met npx
npx create-cloudbox-app my-project
```

### Usage

```bash
# Interactive mode - kies uit beschikbare templates
create-cloudbox-app my-project

# Direct template specificatie
create-cloudbox-app my-portfolio photoportfolio
create-cloudbox-app my-blog blog
create-cloudbox-app my-store ecommerce
create-cloudbox-app my-saas saas
create-cloudbox-app my-crm crm
```

## 📋 Beschikbare Templates

### 1. Photo Portfolio (`photoportfolio`)
**Perfect voor fotografen, designers, kunstenaars**

**Collections:**
- `images` - Foto beheer met metadata, tags, albums
- `albums` - Galerijen en collecties  
- `pages` - Content management (home, about, contact)
- `settings` - Site configuratie en SEO
- `analytics` - Bezoeker statistieken
- `branding` - Visuele identiteit (kleuren, fonts, logo's)

**Features:**
- ✅ Image Management & Upload
- ✅ Album/Gallery Creation  
- ✅ Content Pages & SEO
- ✅ Analytics & Statistics
- ✅ Responsive Design
- ✅ Multi-language Support

### 2. Blog/CMS (`blog`)
**Voor bloggers, content creators, nieuwssites**

**Collections:**
- `posts` - Artikelen en blog posts
- `categories` - Post categorieën
- `comments` - Reactie systeem
- `tags` - Tag management
- `authors` - Auteur profielen

**Features:**
- ✅ Rich Text Editor
- ✅ Category & Tag System
- ✅ Comment Management
- ✅ SEO Optimization
- ✅ Social Media Integration
- ✅ RSS Feeds

### 3. E-commerce (`ecommerce`)
**Voor online winkels en marktplaatsen**

**Collections:**
- `products` - Product catalogus
- `orders` - Bestelling management
- `customers` - Klant accounts
- `categories` - Product categorieën
- `inventory` - Voorraad beheer

**Features:**
- ✅ Product Catalog & Variants
- ✅ Shopping Cart & Checkout
- ✅ Order Management
- ✅ Customer Accounts
- ✅ Inventory Tracking
- ✅ Payment Integration Ready

### 4. SaaS Application (`saas`)
**Voor software-as-a-service platforms**

**Collections:**
- `subscriptions` - Abonnement management
- `usage_metrics` - Feature gebruik tracking
- `teams` - Team/organisatie beheer
- `billing` - Facturatie geschiedenis
- `features` - Feature toggles

**Features:**
- ✅ Subscription Management
- ✅ Team Collaboration
- ✅ Usage-based Billing
- ✅ Feature Toggles
- ✅ Analytics Dashboard
- ✅ API Rate Limiting

### 5. CRM System (`crm`)
**Voor customer relationship management**

**Collections:**
- `contacts` - Klant/lead database
- `deals` - Sales pipeline
- `activities` - Interactie geschiedenis
- `companies` - Bedrijf informatie
- `tasks` - Taak management

**Features:**
- ✅ Contact Management
- ✅ Deal Pipeline & Stages
- ✅ Activity Tracking
- ✅ Task Management
- ✅ Reporting & Analytics
- ✅ Email Integration Ready

## 🛠 Workflow per Project Type

### Universele Stappen

1. **Project Creation**
   ```bash
   create-cloudbox-app my-project [template]
   cd my-project
   ```

2. **Automatische Setup**
   - CloudBox backend configuratie
   - Database collections aanmaken
   - API endpoints genereren
   - Default data invoegen
   - Environment files creëren

3. **Development**
   ```bash
   npm run dev
   # Visit: http://localhost:5173
   # Admin: http://localhost:5173/admin
   ```

4. **Customization**
   - Content toevoegen via admin panel
   - Styling aanpassen
   - Extra features implementeren
   - API endpoints uitbreiden

5. **Deployment**
   ```bash
   npm run build
   # Deploy naar hosting platform
   ```

## 🔧 Template-Specifieke Workflows

### PhotoPortfolio Workflow
```bash
create-cloudbox-app my-portfolio photoportfolio
cd my-portfolio
npm run dev

# Admin workflow:
# 1. Upload photos (/admin/images)
# 2. Create albums (/admin/albums)  
# 3. Edit pages (/admin/pages)
# 4. Configure branding (/admin/branding)
# 5. View analytics (/admin/analytics)
```

### Blog Workflow
```bash
create-cloudbox-app my-blog blog
cd my-blog
npm run dev

# Admin workflow:
# 1. Create categories (/admin/categories)
# 2. Write posts (/admin/posts)
# 3. Manage comments (/admin/comments)
# 4. Configure SEO (/admin/seo)
# 5. View analytics (/admin/analytics)
```

### E-commerce Workflow
```bash
create-cloudbox-app my-store ecommerce
cd my-store
npm run dev

# Admin workflow:
# 1. Add products (/admin/products)
# 2. Manage inventory (/admin/inventory)
# 3. Process orders (/admin/orders)
# 4. Customer service (/admin/customers)
# 5. Sales analytics (/admin/analytics)
```

### SaaS Workflow
```bash
create-cloudbox-app my-saas saas
cd my-saas
npm run dev

# Admin workflow:
# 1. Configure plans (/admin/plans)
# 2. Monitor usage (/admin/usage)
# 3. Manage teams (/admin/teams)
# 4. Billing overview (/admin/billing)
# 5. Feature toggles (/admin/features)
```

### CRM Workflow
```bash
create-cloudbox-app my-crm crm
cd my-crm
npm run dev

# Admin workflow:
# 1. Import contacts (/admin/contacts)
# 2. Create deals (/admin/deals)
# 3. Track activities (/admin/activities)
# 4. Manage companies (/admin/companies)
# 5. Sales reporting (/admin/reports)
```

## 🔌 API Endpoints per Template

### Universal Endpoints (alle templates)
```
GET    /api/templates                    - List available templates
POST   /api/templates/{template}/setup   - Setup project with template
GET    /api/settings                     - Get project settings
PUT    /api/settings                     - Update settings
GET    /api/analytics                    - Get analytics data
```

### PhotoPortfolio API
```
GET    /api/images                 - List images
POST   /api/images                 - Upload image
GET    /api/albums                 - List albums
POST   /api/albums                 - Create album
GET    /api/pages                  - List pages
PUT    /api/branding               - Update branding
```

### Blog API
```
GET    /api/posts                  - List posts
POST   /api/posts                  - Create post
GET    /api/categories             - List categories
GET    /api/comments               - List comments
POST   /api/comments               - Add comment
```

### E-commerce API
```
GET    /api/products               - List products
POST   /api/products               - Add product
GET    /api/orders                 - List orders
POST   /api/orders                 - Create order
GET    /api/customers              - List customers
```

### SaaS API
```
GET    /api/subscriptions          - List subscriptions
POST   /api/subscriptions          - Create subscription
GET    /api/usage_metrics          - Get usage data
GET    /api/teams                  - List teams
POST   /api/teams                  - Create team
```

### CRM API
```
GET    /api/contacts               - List contacts
POST   /api/contacts               - Add contact
GET    /api/deals                  - List deals
POST   /api/deals                  - Create deal
GET    /api/activities             - List activities
```

## 🚀 Extensibility

### Adding Custom Collections

Elk project kan uitgebreid worden met custom collections:

```bash
curl -X POST \\
  http://localhost:8080/p/my-project/api/collections \\
  -H "X-API-Key: your-key" \\
  -d '{
    "name": "testimonials",
    "schema": {
      "name": {"type": "string", "required": true},
      "message": {"type": "string", "required": true},
      "rating": {"type": "number", "required": false}
    }
  }'
```

### Custom Templates

Nieuwe templates kunnen toegevoegd worden door:

1. Template definitie in `templates.go`
2. Frontend template repository
3. Setup script configuratie
4. Documentatie toevoegen

### Template Inheritance

Templates kunnen eigenschappen erven van andere templates:

```yaml
base_template: "blog"
additional_collections:
  - products
  - orders
custom_features:
  - ecommerce_integration
  - payment_processing
```

## 📊 Comparison met Andere Platforms

| Feature | CloudBox | Appwrite | Supabase | Firebase |
|---------|----------|----------|----------|----------|
| **Template System** | ✅ 5 templates | ❌ None | ❌ None | ❌ None |
| **Auto Setup** | ✅ Full automation | ⚠️ Manual | ⚠️ Manual | ⚠️ Manual |
| **Frontend Included** | ✅ Complete apps | ❌ Backend only | ❌ Backend only | ❌ Backend only |
| **Custom Collections** | ✅ Dynamic | ✅ Yes | ✅ Yes | ✅ Yes |
| **Real-time** | 🔄 Coming soon | ✅ Yes | ✅ Yes | ✅ Yes |
| **Self-hosted** | ✅ Docker | ✅ Yes | ✅ Yes | ❌ Cloud only |
| **Templates Ready** | ✅ 5 project types | ❌ Build from scratch | ❌ Build from scratch | ❌ Build from scratch |

## 💡 Best Practices

### Project Structure
```
my-project/
├── src/
│   ├── components/           # React components
│   ├── pages/               # Page components  
│   ├── services/            # API services
│   └── hooks/               # Custom hooks
├── scripts/
│   └── setup-cloudbox.js    # Setup automation
├── .env                     # Environment config
└── README.md               # Project documentation
```

### Environment Management
```bash
# Development
VITE_CLOUDBOX_ENDPOINT=http://localhost:8080
VITE_API_URL=http://localhost:8080/p/my-project/api

# Production  
VITE_CLOUDBOX_ENDPOINT=https://your-cloudbox.com
VITE_API_URL=https://your-cloudbox.com/p/my-project/api
```

### Custom Feature Development
1. Use CloudBox data layer for persistence
2. Extend existing collections rather than creating new ones
3. Follow template conventions for consistency
4. Implement proper error handling
5. Add analytics tracking for new features

## 🎯 Use Cases per Industry

### **Creative Industries**
- **PhotoPortfolio**: Photographers, designers, artists
- **Blog**: Content creators, agencies, portfolios

### **Business & Commerce**  
- **E-commerce**: Online stores, marketplaces, dropshipping
- **CRM**: Sales teams, consultants, service providers

### **Technology**
- **SaaS**: Software companies, startups, API providers
- **Blog**: Technical documentation, company blogs

### **Professional Services**
- **CRM**: Real estate, insurance, consulting
- **PhotoPortfolio**: Architecture, interior design

Dit systeem maakt het mogelijk om binnen 5 minuten een volledig werkende applicatie te hebben, ongeacht het project type!

## 🔮 Roadmap

### Short Term
- [ ] Template marketplace
- [ ] Visual template builder
- [ ] One-click deployment
- [ ] More templates (booking, learning, healthcare)

### Long Term  
- [ ] Real-time collaboration
- [ ] Mobile app templates (React Native)
- [ ] AI-powered content generation
- [ ] Advanced analytics & monitoring
- [ ] Multi-tenant architecture