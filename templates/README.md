# CloudBox Application Templates

This directory contains template repositories that can be deployed with CloudBox's template deployment system.

## Available Templates

### 1. Photo Portfolio Template
**Template Name:** `photoportfolio`
**Framework:** React + Vite
**Features:**
- Image galleries and albums
- Content management for pages
- SEO optimization
- Responsive design
- CloudBox SDK integration

**Variables:**
- `site_name` - Name of the portfolio site
- `theme` - Color theme (modern, classic, minimal)
- `language` - Site language (en, es, fr, de)
- `photographer_name` - Photographer's name
- `contact_email` - Contact email address

### 2. Blog Template
**Template Name:** `blog`
**Framework:** Next.js
**Features:**
- Blog post management
- Categories and tags
- Comment system
- SEO optimization
- CloudBox SDK integration

**Variables:**
- `blog_title` - Blog title
- `blog_description` - Blog description
- `author_name` - Author name
- `theme` - Blog theme
- `language` - Site language

### 3. E-commerce Template
**Template Name:** `ecommerce`
**Framework:** Vue + Nuxt
**Features:**
- Product catalog
- Shopping cart
- Order management
- Payment integration
- CloudBox SDK integration

**Variables:**
- `store_name` - Store name
- `currency` - Default currency
- `language` - Store language
- `payment_provider` - Payment provider (stripe, paypal)

### 4. SaaS Application Template
**Template Name:** `saas`
**Framework:** React + Next.js
**Features:**
- User authentication
- Subscription management
- Dashboard interface
- API integration
- CloudBox SDK integration

**Variables:**
- `app_name` - Application name
- `app_description` - Application description
- `pricing_model` - Pricing model (subscription, one-time)
- `language` - Application language

### 5. Developer Portfolio Template
**Template Name:** `portfolio`
**Framework:** Svelte + SvelteKit
**Features:**
- Project showcase
- Skills display
- Experience timeline
- Contact form
- CloudBox SDK integration

**Variables:**
- `developer_name` - Developer name
- `title` - Professional title
- `github_username` - GitHub username
- `linkedin_profile` - LinkedIn profile URL
- `language` - Portfolio language

## Template Structure

Each template follows this structure:
```
template-name/
├── package.json              # NPM dependencies and scripts
├── README.md                 # Template-specific documentation
├── .env.example             # Environment variable template
├── cloudbox.config.js       # CloudBox configuration
├── src/                     # Source code
│   ├── components/          # Reusable components
│   ├── pages/              # Application pages
│   ├── lib/                # Utility functions
│   └── cloudbox/           # CloudBox integration
├── public/                  # Static assets
└── docs/                   # Template documentation
```

## Variable Substitution

Templates use mustache-style variable substitution:
- `{{variable_name}}` - Simple variable replacement
- `{{#condition}}...{{/condition}}` - Conditional sections
- `{{^condition}}...{{/condition}}` - Inverted conditionals

## CloudBox Integration

All templates include:
1. **CloudBox SDK dependency** in package.json
2. **Environment variables** for CloudBox configuration
3. **CloudBox client setup** in src/lib/cloudbox.js
4. **Example collections** and storage buckets
5. **Deployment configuration** for Docker/production

## Usage

### Via CloudBox Dashboard
1. Go to Templates section in your project
2. Select a template
3. Customize variables
4. Deploy to GitHub repository

### Via API
```bash
curl -X POST /api/v1/projects/:id/template-deployments \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "template_name": "photoportfolio",
    "variables": {
      "site_name": "My Photography",
      "theme": "modern",
      "language": "en"
    },
    "github_repo": {
      "name": "my-portfolio",
      "description": "My photography portfolio",
      "is_private": false
    },
    "deployment_config": {
      "app_port": 3000,
      "auto_deploy": true
    }
  }'
```

### Via SDK Setup Script
```bash
# Use the SDK setup script with template
./sdk/scripts/setup.sh <project_id> <api_key>
# Then select template option during interactive setup
```

## Compatibility Requirements

For a repository to be compatible with CloudBox templates:

### Required Dependencies
- CloudBox SDK: `@ekoppen/cloudbox-sdk`
- Node.js: >= 16.0.0
- NPM/Yarn/PNPM for package management

### Environment Variables
```env
CLOUDBOX_ENDPOINT=http://localhost:8080
CLOUDBOX_API_KEY=your-api-key
CLOUDBOX_PROJECT_ID=your-project-id
CLOUDBOX_PROJECT_SLUG=your-project-slug
```

### CloudBox Configuration
Either in `cloudbox.config.js`:
```javascript
export default {
  endpoint: process.env.CLOUDBOX_ENDPOINT,
  projectId: process.env.CLOUDBOX_PROJECT_ID,
  apiKey: process.env.CLOUDBOX_API_KEY
}
```

Or in application code:
```javascript
import { CloudBoxClient } from '@ekoppen/cloudbox-sdk';

const cloudbox = new CloudBoxClient({
  projectId: process.env.CLOUDBOX_PROJECT_ID,
  apiKey: process.env.CLOUDBOX_API_KEY,
  endpoint: process.env.CLOUDBOX_ENDPOINT
});
```

### Build Scripts
Required in `package.json`:
```json
{
  "scripts": {
    "build": "npm run build",
    "start": "npm start",
    "dev": "npm run dev"
  }
}
```

## Template Development

To create a new template:

1. **Create template directory** in `/templates/`
2. **Add template metadata** in `template.json`
3. **Implement variable substitution** in source files
4. **Test template deployment** with CloudBox
5. **Add template to handler** in `template_deployment.go`
6. **Update documentation**

### Template Metadata
Create `template.json`:
```json
{
  "name": "my-template",
  "version": "1.0.0",
  "description": "My custom template",
  "framework": "react",
  "variables": [
    {
      "name": "app_name",
      "type": "string",
      "required": true,
      "description": "Application name"
    }
  ],
  "collections": [...],
  "buckets": [...],
  "dependencies": {
    "@ekoppen/cloudbox-sdk": "^1.0.0"
  }
}
```

## Examples

See `/templates/examples/` for complete template implementations.

## Contributing

1. Fork the repository
2. Create a new template in `/templates/`
3. Test the template deployment
4. Submit a pull request with documentation

## Support

For template-related issues:
- 📚 [Template Documentation](https://docs.cloudbox.dev/templates)
- 💬 [Community Discord](https://discord.gg/cloudbox)
- 🐛 [Report Issues](https://github.com/ekoppen/cloudbox/issues)