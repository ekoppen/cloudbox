#!/bin/bash

# CloudBox Admin Password Reset Script
# Usage: ./reset-admin.sh [--docker|--local|--python] [--help]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Parse arguments
MODE=""
SHOW_HELP=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --docker)
            MODE="docker"
            shift
            ;;
        --local|--go)
            MODE="local"
            shift
            ;;
        --python)
            MODE="python"
            shift
            ;;
        --help|-h)
            SHOW_HELP=true
            shift
            ;;
        *)
            echo -e "${RED}❌ Unknown option: $1${NC}"
            SHOW_HELP=true
            break
            ;;
    esac
done

# Show help if requested
if [[ "$SHOW_HELP" == true ]]; then
    echo "🔑 CloudBox Admin Password Reset Tool"
    echo "===================================="
    echo
    echo "Reset admin passwords for CloudBox users."
    echo
    echo "Usage:"
    echo "  ./reset-admin.sh [MODE]"
    echo
    echo "Modes:"
    echo "  --docker    Run in Docker mode (recommended, no local deps)"
    echo "  --local     Run with local Go installation"
    echo "  --python    Run with local Python installation"
    echo "  --help      Show this help message"
    echo
    echo "Examples:"
    echo "  ./reset-admin.sh --docker    # Use Docker (recommended)"
    echo "  ./reset-admin.sh --local     # Use local Go"
    echo "  ./reset-admin.sh --python    # Use local Python"
    echo "  ./reset-admin.sh             # Auto-detect best mode"
    echo
    echo "Environment Variables:"
    echo "  DATABASE_URL    Custom database connection string"
    echo
    echo "Alternative:"
    echo "  ./reset-admin-simple.sh     # Simple version (only needs Docker)"
    echo
    exit 0
fi

# Auto-detect mode if not specified
if [[ -z "$MODE" ]]; then
    echo "🔍 Auto-detecting best mode..."
    
    if docker ps >/dev/null 2>&1 && docker ps | grep -q "cloudbox-postgres"; then
        MODE="docker"
        echo -e "${GREEN}✅ Docker detected with CloudBox running${NC}"
    elif command -v go &> /dev/null; then
        MODE="local"
        echo -e "${GREEN}✅ Go installation detected${NC}"
    elif command -v python3 &> /dev/null; then
        MODE="python"
        echo -e "${GREEN}✅ Python3 detected${NC}"
    else
        echo -e "${YELLOW}⚠️ No suitable runtime detected${NC}"
        echo -e "${BLUE}ℹ️ You can also use the simple version: ./reset-admin-simple.sh${NC}"
        echo "Please install Go, Python3, or use Docker mode"
        exit 1
    fi
    echo
fi

echo "🔑 CloudBox Admin Password Reset"
echo "================================"
echo -e "Mode: ${BLUE}$MODE${NC}"
echo

# Execute based on selected mode
case "$MODE" in
    "docker")
        echo -e "${BLUE}🐳 Running in Docker mode...${NC}"
        
        # Check if CloudBox containers are running
        if ! docker ps | grep -q "cloudbox-postgres"; then
            echo -e "${RED}❌ CloudBox postgres container not running${NC}"
            echo "Please start CloudBox first: docker compose up -d"
            exit 1
        fi
        
        # Build and run the Go script in a temporary container
        cat > "$SCRIPT_DIR/Dockerfile.reset" << 'EOF'
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# Copy go.mod and initialize module
COPY go.mod ./

# Add all required dependencies explicitly
RUN go get golang.org/x/crypto/bcrypt@v0.17.0 && \
    go get golang.org/x/term@v0.15.0 && \
    go get gorm.io/driver/postgres@v1.5.4 && \
    go get gorm.io/gorm@v1.25.5 && \
    go mod tidy

# Copy source code
COPY reset-admin-password.go .

# Build with CGO enabled (required for bcrypt)
RUN CGO_ENABLED=1 GOOS=linux go build -o reset-admin reset-admin-password.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/reset-admin .
CMD ["./reset-admin"]
EOF

        # Get database connection from Docker network
        export DATABASE_URL="postgres://cloudbox:cloudbox_dev_password@cloudbox-postgres:5432/cloudbox?sslmode=disable"
        
        # Build and run
        echo -e "${YELLOW}🔨 Building reset tool...${NC}"
        docker build -f "$SCRIPT_DIR/Dockerfile.reset" -t cloudbox-reset-admin "$SCRIPT_DIR"
        
        echo -e "${GREEN}🚀 Starting password reset...${NC}"
        docker run -it --network cloudbox_default -e DATABASE_URL="$DATABASE_URL" cloudbox-reset-admin
        
        # Cleanup
        rm -f "$SCRIPT_DIR/Dockerfile.reset"
        echo -e "${GREEN}🧹 Cleaned up temporary files${NC}"
        ;;
        
    "local")
        echo -e "${BLUE}💻 Running in local Go mode...${NC}"
        
        # Check if Go is installed
        if ! command -v go &> /dev/null; then
            echo -e "${RED}❌ Go is not installed. Please install Go or use --docker flag${NC}"
            exit 1
        fi
        
        # Navigate to script directory
        cd "$SCRIPT_DIR"
        
        # Initialize go.mod if it doesn't exist
        if [ ! -f go.mod ]; then
            echo -e "${YELLOW}📦 Initializing Go module...${NC}"
            go mod init reset-admin
            go get golang.org/x/crypto/bcrypt
            go get golang.org/x/term
            go get gorm.io/driver/postgres
            go get gorm.io/gorm
        fi
        
        # Set database URL from environment or default  
        if [[ -z "$DATABASE_URL" ]]; then
            # Check if we're running with CloudBox containers
            if docker ps | grep -q "cloudbox-postgres"; then
                # Get the IP of the postgres container or use host networking
                POSTGRES_IP=$(docker inspect cloudbox-postgres | grep '"IPAddress"' | head -n1 | sed 's/.*: "\(.*\)",/\1/')
                if [[ -n "$POSTGRES_IP" && "$POSTGRES_IP" != "null" ]]; then
                    export DATABASE_URL="postgres://cloudbox:cloudbox_dev_password@$POSTGRES_IP:5432/cloudbox?sslmode=disable"
                else
                    # Fallback to host port mapping
                    export DATABASE_URL="postgres://cloudbox:cloudbox_dev_password@127.0.0.1:5432/cloudbox?sslmode=disable"
                fi
                echo -e "${BLUE}ℹ️ Using CloudBox database: ${DATABASE_URL}${NC}"
            else
                export DATABASE_URL="postgres://cloudbox:cloudbox_dev_password@localhost:5432/cloudbox?sslmode=disable"
            fi
        fi
        
        echo -e "${GREEN}🚀 Starting password reset tool...${NC}"
        go run reset-admin-password.go
        ;;
        
    "python")
        echo -e "${BLUE}🐍 Running in Python mode...${NC}"
        
        # Check if Python is installed
        if ! command -v python3 &> /dev/null; then
            echo -e "${RED}❌ Python3 is not installed. Please install Python3 or use --docker flag${NC}"
            exit 1
        fi
        
        # Navigate to script directory
        cd "$SCRIPT_DIR"
        
        # Check if requirements are installed
        if ! python3 -c "import bcrypt, psycopg2" 2>/dev/null; then
            echo -e "${YELLOW}📦 Installing Python dependencies...${NC}"
            
            # Try multiple installation methods
            if [ -f requirements.txt ]; then
                # Try with requirements.txt first
                python3 -m pip install -r requirements.txt --user --break-system-packages 2>/dev/null || \
                python3 -m pip install -r requirements.txt --user 2>/dev/null || \
                pip3 install -r requirements.txt --user 2>/dev/null || \
                {
                    echo -e "${YELLOW}⚠️ pip install failed, trying individual packages...${NC}"
                    python3 -m pip install bcrypt psycopg2-binary --user --break-system-packages 2>/dev/null || \
                    python3 -m pip install bcrypt psycopg2-binary --user 2>/dev/null || \
                    pip3 install bcrypt psycopg2-binary --user 2>/dev/null || \
                    {
                        echo -e "${RED}❌ Failed to install Python dependencies${NC}"
                        echo "Please install manually with one of these commands:"
                        echo "  pip install bcrypt psycopg2-binary"
                        echo "  pip3 install bcrypt psycopg2-binary"
                        echo "  python3 -m pip install bcrypt psycopg2-binary --user"
                        exit 1
                    }
                }
            else
                # Try direct installation
                python3 -m pip install bcrypt psycopg2-binary --user --break-system-packages 2>/dev/null || \
                python3 -m pip install bcrypt psycopg2-binary --user 2>/dev/null || \
                pip3 install bcrypt psycopg2-binary --user 2>/dev/null || \
                {
                    echo -e "${RED}❌ Failed to install Python dependencies${NC}"
                    echo "Please install manually with one of these commands:"
                    echo "  pip install bcrypt psycopg2-binary"
                    echo "  pip3 install bcrypt psycopg2-binary"
                    echo "  python3 -m pip install bcrypt psycopg2-binary --user"
                    exit 1
                }
            fi
            
            echo -e "${GREEN}✅ Python dependencies installed successfully${NC}"
        fi
        
        # Set database URL from environment or default  
        if [[ -z "$DATABASE_URL" ]]; then
            # Check if we're running with CloudBox containers
            if docker ps | grep -q "cloudbox-postgres"; then
                # Get the IP of the postgres container or use host networking
                POSTGRES_IP=$(docker inspect cloudbox-postgres | grep '"IPAddress"' | head -n1 | sed 's/.*: "\(.*\)",/\1/')
                if [[ -n "$POSTGRES_IP" && "$POSTGRES_IP" != "null" ]]; then
                    export DATABASE_URL="postgres://cloudbox:cloudbox_dev_password@$POSTGRES_IP:5432/cloudbox?sslmode=disable"
                else
                    # Fallback to host port mapping
                    export DATABASE_URL="postgres://cloudbox:cloudbox_dev_password@127.0.0.1:5432/cloudbox?sslmode=disable"
                fi
                echo -e "${BLUE}ℹ️ Using CloudBox database: ${DATABASE_URL}${NC}"
            else
                export DATABASE_URL="postgres://cloudbox:cloudbox_dev_password@localhost:5432/cloudbox?sslmode=disable"
            fi
        fi

        echo -e "${GREEN}🚀 Starting password reset tool...${NC}"
        python3 reset-admin.py
        ;;
        
    *)
        echo -e "${RED}❌ Unknown mode: $MODE${NC}"
        exit 1
        ;;
esac

echo
echo -e "${GREEN}✅ Password reset script completed successfully!${NC}"
echo -e "${BLUE}ℹ️  You can now login to CloudBox with the new credentials${NC}"