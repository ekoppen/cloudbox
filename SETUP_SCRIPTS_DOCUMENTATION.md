# CloudBox Project Setup & Development Guide

Complete guide for setting up CloudBox projects using automated scripts, templates, and development roadmap.

## 🚀 Interactive Setup Script

The interactive setup script provides a guided experience for setting up new CloudBox projects with Docker configuration.

### Features

✅ **Connection Testing** - Validates CloudBox URL and API credentials  
✅ **Project Templates** - 6 pre-configured project types  
✅ **Docker Integration** - Generates docker-compose.yml and .env files  
✅ **Automated Setup** - Creates all necessary collections and storage buckets  
✅ **User Configuration** - Configurable Docker user and application port  
✅ **Error Handling** - Comprehensive validation and error reporting  

### Usage

#### Option 1: Direct Download and Run
```bash
# Download and run the interactive setup
curl -fsSL https://raw.githubusercontent.com/ekoppen/cloudbox/main/sdk/scripts/interactive-setup.sh | bash
```

#### Option 2: Clone Repository
```bash
# Clone the repository and run locally
git clone https://github.com/ekoppen/cloudbox.git
cd cloudbox
./sdk/scripts/interactive-setup.sh
```

#### Option 3: Copy Script to Your Project
```bash
# Copy script to your project directory
cp /path/to/cloudbox/sdk/scripts/interactive-setup.sh ./
chmod +x interactive-setup.sh
./interactive-setup.sh
```

### Interactive Prompts

The script will ask you to configure:

1. **CloudBox URL** - e.g., `http://localhost:8080`
2. **Project Type** - Existing project or new setup
3. **Project Code/Slug** - Your project identifier
4. **API Key** - Your CloudBox API key (hidden input)
5. **Project Template** - Choose from 6 templates (if new project)
6. **Docker User** - Application user name (default: appuser)
7. **Application Port** - Port for your app (default: 3000)

### Project Templates

#### 1. Photo Portfolio Template
Perfect for photography websites and portfolios.

**Collections Created:**
- `pages` - Website pages with SEO fields
- `albums` - Photo album organization
- `images` - Image metadata and thumbnails
- `settings` - Site configuration

**Storage Buckets:**
- `images` - Portfolio photos (10MB limit)
- `thumbnails` - Generated thumbnails (2MB limit)
- `branding` - Site branding assets (5MB limit)

#### 2. Blog/CMS Template
Content management system for blogs and news sites.

**Collections Created:**
- `posts` - Blog posts with categories and tags
- `categories` - Post categorization
- `pages` - Static pages
- `authors` - Author profiles

**Storage Buckets:**
- `uploads` - Blog media uploads (50MB limit)
- `avatars` - Author profile pictures (5MB limit)

#### 3. E-commerce Template
Online store with product catalog and order management.

**Collections Created:**
- `products` - Product catalog with pricing
- `categories` - Product categorization
- `orders` - Order management
- `customers` - Customer profiles

**Storage Buckets:**
- `products` - Product images (10MB limit)
- `categories` - Category images (5MB limit)

#### 4. SaaS Application Template
Software-as-a-Service application with user management.

**Collections Created:**
- `users` - User accounts and profiles
- `subscriptions` - Subscription management
- `plans` - Pricing plans and features
- `usage_logs` - Usage tracking

**Storage Buckets:**
- `avatars` - User profile pictures (5MB limit)
- `uploads` - User file uploads (100MB limit, private)

#### 5. Developer Portfolio Template
Showcase for developers to display projects and skills.

**Collections Created:**
- `projects` - Project showcase
- `skills` - Technical skills
- `experience` - Work experience
- `contact_messages` - Contact form messages

**Storage Buckets:**
- `projects` - Project screenshots (10MB limit)
- `resume` - Resume and documents (5MB limit, private)

#### 6. Custom Template
Empty template for custom project structures.

### Generated Files

After running the setup script, you'll get:

#### `.env` File
```env
# CloudBox Configuration
CLOUDBOX_ENDPOINT=http://localhost:8080
CLOUDBOX_PROJECT_CODE=your-project
CLOUDBOX_API_KEY=your-api-key

# Application Configuration
APP_PORT=3000
NODE_ENV=production

# Docker Configuration
USER_ID=1000
GROUP_ID=1000
APP_USER=appuser
```

#### `docker-compose.yml` File
```yaml
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
      args:
        - USER_ID=${USER_ID}
        - GROUP_ID=${GROUP_ID}
        - APP_USER=${APP_USER}
    container_name: your-project-app
    restart: unless-stopped
    ports:
      - "${APP_PORT}:3000"
    environment:
      - CLOUDBOX_ENDPOINT=${CLOUDBOX_ENDPOINT}
      - CLOUDBOX_PROJECT_CODE=${CLOUDBOX_PROJECT_CODE}
      - CLOUDBOX_API_KEY=${CLOUDBOX_API_KEY}
      - NODE_ENV=${NODE_ENV}
    env_file:
      - .env
    volumes:
      - ./logs:/app/logs
    networks:
      - app-network

networks:
  app-network:
    driver: bridge
```

#### `Dockerfile.example`
Example Dockerfile with proper user configuration:

```dockerfile
FROM node:18-alpine

# Create app user
ARG USER_ID=1000
ARG GROUP_ID=1000
ARG APP_USER=appuser

RUN addgroup -g $GROUP_ID $APP_USER && \
    adduser -D -u $USER_ID -G $APP_USER $APP_USER

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci --only=production

# Copy application code
COPY . .

# Change ownership
RUN chown -R $APP_USER:$APP_USER /app

# Switch to app user
USER $APP_USER

# Expose port
EXPOSE 3000

# Start application
CMD ["npm", "start"]
```

## 🛠️ Manual Setup Script Template

For more control, use the manual setup template that you can customize:

### Usage
```bash
# Copy and customize the template
cp /path/to/cloudbox/sdk/scripts/setup-template.sh ./setup-my-project.sh
chmod +x setup-my-project.sh

# Edit the script to add your collections and buckets
# Then run it
./setup-my-project.sh <project_id> <api_key> [endpoint]
```

### Customization Example

Edit the script to add your specific collections:

```bash
# Create your custom collections
create_collection "products" \
    "name:string" \
    "price:float" \
    "description:text" \
    "in_stock:boolean"

create_collection "orders" \
    "user_id:string" \
    "products:array" \
    "total:float" \
    "status:string"

# Create your storage buckets
create_bucket "product-images" "Product photos" true 10485760
create_bucket "user-uploads" "User files" false 52428800
```

## 📦 SDK Integration

After running the setup script, integrate the CloudBox SDK:

### 1. Install SDK
```bash
npm install @ekoppen/cloudbox-sdk
```

### 2. Use in Your Application
```javascript
import { CloudBoxClient } from '@ekoppen/cloudbox-sdk';

const cloudbox = new CloudBoxClient({
  projectId: process.env.CLOUDBOX_PROJECT_CODE,
  apiKey: process.env.CLOUDBOX_API_KEY,
  endpoint: process.env.CLOUDBOX_ENDPOINT
});

// Test connection
const connected = await cloudbox.testConnection();
if (!connected) {
  console.error('Cannot connect to CloudBox');
  process.exit(1);
}

// Use the created collections
const posts = await cloudbox.collections.list();
console.log('Available collections:', posts);
```

### 3. Docker Development
```bash
# Copy the example Dockerfile
cp Dockerfile.example Dockerfile

# Build and run with Docker Compose
docker-compose up --build

# Your app will be available at http://localhost:3000 (or your configured port)
```

## 🔧 Advanced Configuration

### Environment Variables

The setup scripts support these environment variables:

```bash
# Override defaults
export CLOUDBOX_ENDPOINT="https://my-cloudbox.example.com"
export DEFAULT_APP_PORT="8080"
export DEFAULT_APP_USER="myapp"

# Run setup with custom defaults
./interactive-setup.sh
```

### Custom Templates

Create your own templates by modifying the `setup_custom()` function in the interactive script:

```bash
setup_my_template() {
    print_header "Setting up My Custom Application"
    
    # Add your collections
    create_collection "custom_collection" \
        "field1:string" \
        "field2:text" \
        "field3:boolean"
    
    # Add your buckets
    create_bucket "custom_bucket" "Custom files" true 10485760
}
```

### Integration with CI/CD

Use the setup scripts in your CI/CD pipeline:

```yaml
# GitHub Actions example
name: Deploy to CloudBox
on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Setup CloudBox Project
        run: |
          ./sdk/scripts/setup-template.sh \
            ${{ secrets.CLOUDBOX_PROJECT_ID }} \
            ${{ secrets.CLOUDBOX_API_KEY }} \
            ${{ secrets.CLOUDBOX_ENDPOINT }}
      
      - name: Deploy Application
        run: docker-compose up -d
```

## 🐛 Troubleshooting

### Common Issues

#### "Cannot connect to CloudBox"
- Check CloudBox URL is correct and accessible
- Verify CloudBox service is running
- Check firewall/network settings

#### "Invalid API key or project code"
- Verify API key is correct
- Check project code/slug matches CloudBox project
- Ensure API key has proper permissions

#### "Collection already exists"
- This is normal - script will skip existing collections
- Use the warning as confirmation that setup was previously run

#### Docker permission issues
- Ensure user has Docker permissions
- Check USER_ID and GROUP_ID in .env file
- Verify app user exists in container

### Debug Mode

Enable debug output by setting:

```bash
export DEBUG=1
./interactive-setup.sh
```

## 📚 Next Steps

After running the setup scripts:

1. **Review Configuration** - Check `.env` and `docker-compose.yml`
2. **Customize Dockerfile** - Adapt `Dockerfile.example` for your app
3. **Install Dependencies** - Add CloudBox SDK and other packages
4. **Develop Application** - Start building with CloudBox collections
5. **Deploy** - Use Docker Compose for local development or production

## 🤝 Contributing

To add new project templates or improve the setup scripts:

1. Fork the CloudBox repository
2. Add your template to `interactive-setup.sh`
3. Test with different configurations
4. Submit a pull request

## 🗺️ CloudBox Development Roadmap

### Phase 1 (0-3 months): Firebase Feature Parity
**Priority improvements to match Firebase core capabilities:**

- **Real-time WebSocket subscriptions** - Live data updates and reactive applications
- **Advanced OAuth authentication** - Google, GitHub, Microsoft integrations  
- **Developer CLI tools** - Local development server and project scaffolding
- **Database migrations** - Production-safe schema evolution
- **Enhanced file storage** - CDN integration and image transformation
- **Built-in analytics** - Performance monitoring and usage metrics

**Target:** 80% Firebase feature parity, <100ms API latency, 2K+ GitHub stars

### Phase 2 (3-6 months): Enterprise & Advanced Features
**Beyond Firebase capabilities:**

- **Serverless functions** - Multi-runtime execution with event triggers
- **Advanced database** - Full-text search, geospatial queries, row-level security
- **Multi-environment management** - Dev/staging/production workflows
- **Team collaboration** - Role-based access control for 100+ member teams
- **Integration platform** - Webhooks and third-party service connectors

**Target:** 500+ enterprise organizations, $100K+ ARR

### Phase 3 (6-12 months): Next-Generation Platform
**Innovation leadership:**

- **AI/ML integration** - Vector database, model hosting, natural language queries
- **Global edge network** - Multi-region deployment with automatic replication
- **Visual application builder** - No-code development interface
- **Enterprise compliance** - SOC2, GDPR, HIPAA compliance tools
- **Developer marketplace** - Extension ecosystem with revenue sharing

**Target:** Top 3 Firebase alternative, $10M+ ARR, 1M+ applications

### Investment & Timeline
- **Total Year 1**: ~$8M investment (team + infrastructure + marketing)
- **Break-even**: Month 18 with $2M ARR
- **Team scaling**: 9 → 15 → 23 people across 3 phases
- **ROI**: 300% by Month 24 with $20M projected valuation

## 📞 Support

Need help with setup scripts or development?

- 📚 [CloudBox Documentation](https://docs.cloudbox.dev)
- 💬 [Community Discord](https://discord.gg/cloudbox)
- 🐛 [Report Issues](https://github.com/ekoppen/cloudbox/issues)
- 📧 [Email Support](mailto:support@cloudbox.dev)