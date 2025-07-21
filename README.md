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

```bash
# Clone and start with Docker
git clone https://github.com/your-org/cloudbox.git
cd cloudbox
docker-compose up -d

# Access dashboard
open http://localhost:3000
```

## 📚 Documentation

- [Getting Started](./docs/getting-started.md)
- [API Reference](./docs/api-reference.md)
- [SDK Documentation](./docs/sdk-documentation.md)
- [Deployment Guide](./docs/deployment.md)

## 🔧 Development

See [Development Guide](./docs/development.md) for local development setup.

## 📄 License

MIT License - see [LICENSE](./LICENSE) file for details.