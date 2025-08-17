# CloudBox Plugin Submission & Approval Guide

Complete gids voor het indienen en goedkeuren van plugins voor het CloudBox Marketplace.

## 🏗️ Plugin Repository Structuur

### Eigen Repository Setup
Elke plugin heeft zijn eigen GitHub repository met deze structuur:

```
your-cloudbox-plugin/
├── README.md                    # Plugin beschrijving en documentatie
├── plugin.json                  # Plugin manifest (verplicht)
├── package.json                 # Node.js dependencies
├── LICENSE                      # MIT/Apache 2.0 licentie (verplicht)
├── CHANGELOG.md                 # Versie geschiedenis
├── .github/                     # GitHub workflows
│   └── workflows/
│       ├── security-scan.yml    # Security scanning
│       └── plugin-test.yml      # Automated testing
├── src/                         # Plugin source code
│   ├── index.js                 # Main plugin entry point
│   ├── handlers/                # API handlers
│   ├── migrations/              # Database migrations
│   └── utils/                   # Utility functions
├── components/                  # Frontend components
│   ├── Dashboard.svelte         # Main dashboard component
│   ├── Settings.svelte          # Plugin settings
│   └── shared/                  # Shared components
├── templates/                   # Project templates (optioneel)
│   ├── basic-setup.json
│   └── advanced-config.json
├── tests/                       # Test suite
│   ├── unit/                    # Unit tests
│   ├── integration/             # Integration tests
│   └── fixtures/                # Test data
├── docs/                        # Documentatie
│   ├── installation.md
│   ├── configuration.md
│   └── api-reference.md
└── scripts/                     # Build/deploy scripts
    ├── build.sh
    └── install.sh
```

### Plugin Manifest (plugin.json)
```json
{
  "name": "cloudbox-your-plugin",
  "version": "1.0.0",
  "description": "Kort beschrijving van je plugin functionaliteit",
  "long_description": "Uitgebreide beschrijving met features en use cases",
  "author": "Je Naam of Organisatie",
  "email": "support@yourplugin.com",
  "website": "https://yourplugin.com",
  "repository": "https://github.com/yourusername/cloudbox-your-plugin",
  "license": "MIT",
  "keywords": ["database", "automation", "api"],
  "category": "development",
  "type": "dashboard-plugin",
  "main": "src/index.js",
  "cloudbox_version": ">=1.0.0",
  "node_version": ">=16.0.0",
  "dependencies": {
    "cloudbox-sdk": "^1.0.0"
  },
  "permissions": [
    "database:read",
    "database:write",
    "functions:deploy"
  ],
  "ui": {
    "dashboard_tab": {
      "title": "Your Plugin",
      "icon": "your-icon",
      "path": "/your-plugin"
    },
    "project_menu": {
      "title": "Plugin Feature",
      "icon": "feature-icon",
      "path": "/projects/{projectId}/your-plugin"
    }
  },
  "marketplace": {
    "featured": false,
    "screenshots": [
      "https://raw.githubusercontent.com/yourusername/cloudbox-your-plugin/main/screenshots/dashboard.png",
      "https://raw.githubusercontent.com/yourusername/cloudbox-your-plugin/main/screenshots/settings.png"
    ],
    "demo_url": "https://demo.yourplugin.com",
    "support_url": "https://support.yourplugin.com",
    "pricing": "free",
    "tags": ["database", "automation", "productivity"]
  },
  "security": {
    "sandbox": true,
    "timeout": 30000,
    "memory_limit": "256MB",
    "network_access": ["api.external-service.com"],
    "file_system_access": ["read-only"]
  }
}
```

## 🔐 Approved Repository System

### Huidige Approved Organizations
```javascript
// In backend/internal/security/plugin_validator.go
const ApprovedRepositories = {
  // Official CloudBox plugins
  "github.com/cloudbox/official-plugins": {
    verified: true,
    auto_approve: false,
    security_level: "high"
  },
  
  // Community plugins
  "github.com/cloudbox/community-plugins": {
    verified: true, 
    auto_approve: false,
    security_level: "medium"
  },
  
  // Partner organizations
  "github.com/cloudbox-partners/*": {
    verified: true,
    auto_approve: false,
    security_level: "high"
  },
  
  // Individual developers (case-by-case approval)
  "github.com/verified-developers/*": {
    verified: false,
    auto_approve: false,
    security_level: "medium"
  }
}
```

### Repository Approval Proces
1. **Request Access** - Email naar `plugins@cloudbox.dev`
2. **Organization Review** - CloudBox team evalueert organisatie
3. **Security Assessment** - Code security review
4. **Approval Notification** - Repository wordt toegevoegd aan approved list
5. **Plugin Submission** - Plugins kunnen worden ingediend

## 📝 Submission Workflow

### 1. Plugin Development
```bash
# Clone plugin template
git clone https://github.com/cloudbox/plugin-template.git your-plugin
cd your-plugin

# Setup development environment
npm install
npm run setup

# Develop your plugin
# ... development work ...

# Test thoroughly
npm test
npm run integration-test
npm run security-scan
```

### 2. Plugin Submission
```bash
# Submit via GitHub Issues
# Repository: https://github.com/cloudbox/marketplace-submissions
# Template: Plugin Submission Request
```

### 3. Submission Template
```markdown
# Plugin Submission: [Plugin Name]

## Basic Information
- **Plugin Name**: cloudbox-your-plugin
- **Version**: 1.0.0
- **Repository**: https://github.com/yourusername/cloudbox-your-plugin
- **Category**: development
- **License**: MIT

## Description
Brief description of what your plugin does and why it's useful.

## Features
- [ ] Feature 1
- [ ] Feature 2
- [ ] Feature 3

## Testing
- [ ] Unit tests pass (>90% coverage)
- [ ] Integration tests pass
- [ ] Security scan clean
- [ ] Performance benchmarks meet requirements
- [ ] Manual testing completed

## Documentation
- [ ] README.md complete
- [ ] API documentation
- [ ] Installation guide
- [ ] Configuration examples
- [ ] Troubleshooting guide

## Security
- [ ] No known vulnerabilities
- [ ] Input validation implemented
- [ ] Error handling robust
- [ ] Logging appropriate
- [ ] Permissions minimal

## Compliance
- [ ] Follows CloudBox coding standards
- [ ] GPL/MIT compatible license
- [ ] No copyrighted code without permission
- [ ] GDPR compliant (if applicable)
- [ ] Accessibility guidelines followed

## Screenshots
![Dashboard](screenshots/dashboard.png)
![Settings](screenshots/settings.png)

## Demo
Live demo available at: https://demo.yourplugin.com

## Support
- Documentation: https://docs.yourplugin.com
- Issues: https://github.com/yourusername/cloudbox-your-plugin/issues
- Email: support@yourplugin.com
```

## 🔍 Review Process

### Automated Checks
```yaml
# .github/workflows/plugin-review.yml
name: Plugin Review Automation

on:
  pull_request:
    paths: ['plugins/*/plugin.json']

jobs:
  security-scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Security Scan
        run: |
          npm audit
          snyk test
          semgrep --config=cloudbox-security
          
  code-quality:
    runs-on: ubuntu-latest  
    steps:
      - uses: actions/checkout@v3
      - name: Code Quality
        run: |
          eslint src/
          prettier --check src/
          npm run test:coverage
          
  compatibility-test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        cloudbox-version: ['1.0.0', '1.1.0', 'latest']
    steps:
      - name: Test CloudBox Compatibility
        run: |
          npm run test:compatibility ${{ matrix.cloudbox-version }}
```

### Manual Review Checklist

#### 🔐 Security Review
- [ ] **Code Security**: No SQL injection, XSS, or CSRF vulnerabilities
- [ ] **Dependency Security**: All dependencies scanned and clean
- [ ] **Permission Usage**: Minimal required permissions requested
- [ ] **Data Handling**: Secure data storage and transmission
- [ ] **Input Validation**: All user inputs validated and sanitized
- [ ] **Error Handling**: No sensitive information in error messages
- [ ] **Authentication**: Proper integration with CloudBox auth system

#### 🏗️ Code Quality Review  
- [ ] **Architecture**: Clean, maintainable code structure
- [ ] **Testing**: Comprehensive test suite (>90% coverage)
- [ ] **Performance**: No memory leaks or performance bottlenecks
- [ ] **Standards**: Follows CloudBox coding standards
- [ ] **Documentation**: Complete and accurate documentation
- [ ] **Error Handling**: Graceful error handling and recovery
- [ ] **Logging**: Appropriate logging without sensitive data

#### 🎨 UI/UX Review
- [ ] **Design**: Consistent with CloudBox design system
- [ ] **Accessibility**: WCAG 2.1 AA compliance
- [ ] **Responsiveness**: Works on mobile and desktop
- [ ] **Performance**: Fast loading and smooth interactions
- [ ] **Localization**: Internationalization support (if needed)
- [ ] **Dark Mode**: Support for dark/light themes

#### 📚 Documentation Review
- [ ] **Completeness**: All features documented
- [ ] **Accuracy**: Documentation matches implementation
- [ ] **Examples**: Clear usage examples provided
- [ ] **Installation**: Step-by-step installation guide
- [ ] **Troubleshooting**: Common issues and solutions
- [ ] **API Reference**: Complete API documentation

### Review Timeline
- **Automated Checks**: ~15 minutes
- **Initial Review**: 3-5 business days
- **Security Audit**: 5-10 business days
- **Final Approval**: 2-3 business days
- **Total Timeline**: 10-18 business days

## ✅ Approval Criteria

### Must Have Requirements
- ✅ **Security Clean** - No critical vulnerabilities
- ✅ **Functional** - Plugin works as described
- ✅ **Documented** - Complete documentation
- ✅ **Tested** - Comprehensive test suite
- ✅ **Compatible** - Works with current CloudBox version
- ✅ **Licensed** - GPL/MIT compatible license
- ✅ **Original** - No copyright violations

### Quality Scores
```javascript
const approvalCriteria = {
  security: {
    weight: 30,
    minimumScore: 8.5
  },
  functionality: {
    weight: 25,
    minimumScore: 8.0
  },
  codeQuality: {
    weight: 20,
    minimumScore: 7.5
  },
  documentation: {
    weight: 15,
    minimumScore: 8.0
  },
  userExperience: {
    weight: 10,
    minimumScore: 7.0
  }
}

// Minimum overall score: 8.0/10
```

### Approval Levels
- **Auto-Approved**: Updates van verified developers (patches/minor)
- **Fast-Track**: Official CloudBox team plugins
- **Standard Review**: Community plugins (10-18 days)
- **Extended Review**: Complex plugins, first-time developers (3-4 weeks)

## 📊 Marketplace Integration

### Plugin Listing Process
1. **Approval Granted** - Plugin toegevoegd aan approved repositories
2. **Metadata Extraction** - Plugin.json data wordt geïndexeerd
3. **Screenshot Processing** - Screenshots geoptimaliseerd voor marketplace
4. **SEO Optimization** - Beschrijvingen en tags geoptimaliseerd
5. **Marketplace Deploy** - Plugin verschijnt in marketplace

### Marketplace Data Structure
```json
{
  "marketplace_plugins": [
    {
      "id": "cloudbox-your-plugin",
      "name": "Your Plugin Name",
      "slug": "your-plugin",
      "repository": "https://github.com/yourusername/cloudbox-your-plugin",
      "verified": true,
      "official": false,
      "featured": false,
      "approval_date": "2024-08-17T10:00:00Z",
      "last_updated": "2024-08-17T10:00:00Z",
      "download_count": 0,
      "rating": null,
      "review_count": 0,
      "marketplace_metadata": {
        "screenshots": [...],
        "demo_url": "...",
        "support_url": "...",
        "pricing": "free"
      }
    }
  ]
}
```

## 🚀 Plugin Update Process

### Version Updates
```bash
# 1. Update version in plugin.json
{
  "version": "1.1.0"
}

# 2. Create changelog entry
## [1.1.0] - 2024-08-17
### Added
- New feature X
### Fixed  
- Bug fix Y

# 3. Tag release
git tag v1.1.0
git push origin v1.1.0

# 4. Update automatically detected
# No re-approval needed for patch/minor updates
```

### Major Updates
Voor major updates (breaking changes):
1. **Breaking Change Notice** - Minimaal 30 dagen vooraf
2. **Migration Guide** - Gebruikers helpen met upgrade
3. **Backward Compatibility** - Indien mogelijk
4. **Re-review Required** - Security en compatibility check

## 👥 Developer Support

### Getting Started
- **Plugin Template**: https://github.com/cloudbox/plugin-template
- **SDK Documentation**: https://docs.cloudbox.dev/sdk
- **Developer Discord**: https://discord.gg/cloudbox-dev
- **Office Hours**: Elke dinsdag 15:00-16:00 CET

### Developer Resources
- **Style Guide**: CloudBox coding standards
- **Component Library**: Shared UI components
- **Testing Utils**: Plugin testing utilities
- **Security Guide**: Security best practices
- **Performance Guide**: Optimization guidelines

### Support Channels
- **GitHub Discussions**: https://github.com/cloudbox/marketplace/discussions
- **Email Support**: plugins@cloudbox.dev
- **Bug Reports**: https://github.com/cloudbox/marketplace/issues
- **Feature Requests**: https://github.com/cloudbox/marketplace/issues

## 📈 Plugin Analytics

### Marketplace Metrics
Developers krijgen toegang tot:
- **Download Statistics** - Daily, weekly, monthly downloads
- **User Ratings** - Star ratings en reviews
- **Usage Analytics** - Active installs en engagement
- **Error Reports** - Crash reports en error logs
- **Performance Metrics** - Load times en resource usage

### Revenue Sharing (Future)
Voor premium plugins:
- **70/30 Split** - 70% developer, 30% CloudBox
- **Monthly Payouts** - Via Stripe Connect
- **Tax Handling** - Automated tax compliance
- **Analytics Dashboard** - Revenue en usage analytics

## 🔄 Lifecycle Management

### Plugin States
- **Development** - In ontwikkeling
- **Submitted** - Ingediend voor review
- **Under Review** - Wordt beoordeeld
- **Approved** - Goedgekeurd voor marketplace
- **Published** - Live in marketplace
- **Deprecated** - Niet meer onderhouden
- **Removed** - Verwijderd uit marketplace

### End-of-Life Process
1. **Deprecation Notice** - 90 dagen vooraf
2. **Migration Assistance** - Help met alternatieven
3. **Data Export** - Gebruikersdata export
4. **Graceful Shutdown** - Geleidelijke uitfasering
5. **Removal** - Definitieve verwijdering

---

**CloudBox Plugin Marketplace** - Van ontwikkeling tot distributie, wij ondersteunen je volledige plugin journey! 🚀