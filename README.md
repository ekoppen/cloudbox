# 🚀 CloudBox

A self-hosted Backend-as-a-Service (BaaS) platform that provides essential backend services for modern applications.

## 🌟 Features

🚀 **Core Services**
- **Authentication & User Management** - JWT-based auth with roles and permissions
- **Database Management** - RESTful API for document storage and retrieval
- **File Storage** - Upload, manage and serve files with organized bucket system
- **Real-time Messaging** - WebSocket-based messaging with channels and rooms
- **Serverless Functions** - Deploy and execute custom functions
- **Automated Deployments** - Git-based deployments with webhook support

🛠 **Developer Tools**
- **Admin Dashboard** - Comprehensive management interface with statistics
- **Project Management** - Multi-project organization with API keys
- **User Analytics** - Growth metrics and usage statistics
- **System Monitoring** - Health metrics and performance tracking
- **Backup & Recovery** - Automated database backups

🔧 **Technical Stack**
- **Frontend**: SvelteKit + TypeScript + TailwindCSS
- **Backend**: Go + Gin Framework + GORM
- **Database**: PostgreSQL
- **Cache**: Redis
- **Deployment**: Docker + Docker Compose

## 🏗️ Architecture

```
cloudbox/
├── backend/          # Go API server
├── frontend/         # Svelte web dashboard  
├── cli/             # CloudBox CLI tool
├── sdks/            # Client SDKs
├── docs/            # Documentation
└── docker/          # Container configurations
```

## 🚀 Quick Start

### Prerequisites
- Docker & Docker Compose
- Git

### Automated Installation

```bash
# Clone the repository
git clone https://github.com/ekoppen/cloudbox.git
cd cloudbox

# Run the installation script
./install.sh

# Or with custom configuration
./install.sh --frontend-port 8080 --backend-port 9000 --host myserver.com
```

### Manual Installation

```bash
# 1. Clone repository
git clone https://github.com/ekoppen/cloudbox.git
cd cloudbox

# 2. Configure environment
cp .env.example .env
# Edit .env with your configuration

# 3. Start services
docker-compose up -d

# 4. Access your CloudBox
open http://localhost:3000
```

### Installation Options

```bash
# Basic installation
./install.sh

# Custom ports
./install.sh --frontend-port 8080 --backend-port 9000

# Remote server installation
./install.sh --host myserver.com --frontend-port 3000

# Update existing installation
./install.sh --update

# Full options
./install.sh --help
```

## 📚 Documentation

- [Getting Started](./docs/getting-started.md)
- [Installation Guide](./docs/installation.md)
- [API Reference](./docs/api-reference.md)
- [SDK Documentation](./docs/sdk.md)
- [Configuration](./docs/configuration.md)
- [Deployment Guide](./docs/deployment.md)
- [Development Guide](./docs/development.md)
- [Troubleshooting](./docs/troubleshooting.md)

## 🔧 Development

See [Development Guide](./docs/development.md) for local development setup.

## 📄 License

MIT License - see [LICENSE](./LICENSE) file for details.