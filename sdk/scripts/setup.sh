#!/bin/bash
# CloudBox Quick Setup Script
# 
# Usage: ./setup.sh <project_id> <api_key> [endpoint]
#
# This script provides a simple way to set up CloudBox projects.
# For advanced interactive setup, see SETUP_SCRIPTS_DOCUMENTATION.md

set -e

# Configuration
CLOUDBOX_PROJECT_ID="${1:-}"
CLOUDBOX_API_KEY="${2:-}"
CLOUDBOX_ENDPOINT="${3:-http://localhost:8080}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Validate inputs
if [[ -z "$CLOUDBOX_PROJECT_ID" ]] || [[ -z "$CLOUDBOX_API_KEY" ]]; then
    echo -e "${RED}Usage: $0 <project_id> <api_key> [endpoint]${NC}"
    echo "Example: $0 9 cb_abc123xyz http://localhost:8080"
    exit 1
fi

echo -e "${GREEN}CloudBox Quick Setup${NC}"
echo "Project ID: $CLOUDBOX_PROJECT_ID"
echo "Endpoint: $CLOUDBOX_ENDPOINT"
echo ""

# Test connection
echo -e "${YELLOW}Testing CloudBox connection...${NC}"
if ! curl -s --connect-timeout 5 "$CLOUDBOX_ENDPOINT/health" >/dev/null; then
    echo -e "${RED}✗ Cannot connect to CloudBox at $CLOUDBOX_ENDPOINT${NC}"
    exit 1
fi

echo -e "${GREEN}✓ CloudBox is accessible${NC}"

# Create .env file
cat > .env << EOF
# CloudBox Configuration
CLOUDBOX_ENDPOINT=$CLOUDBOX_ENDPOINT
CLOUDBOX_PROJECT_ID=$CLOUDBOX_PROJECT_ID
CLOUDBOX_API_KEY=$CLOUDBOX_API_KEY

# Application Configuration
NODE_ENV=development
EOF

echo -e "${GREEN}✓ Created .env file${NC}"

# Create example package.json
if [[ ! -f "package.json" ]]; then
    cat > package.json << EOF
{
  "name": "cloudbox-app",
  "version": "1.0.0",
  "description": "CloudBox application",
  "main": "index.js",
  "scripts": {
    "start": "node index.js",
    "dev": "node index.js"
  },
  "dependencies": {
    "@ekoppen/cloudbox-sdk": "^1.0.0"
  }
}
EOF
    echo -e "${GREEN}✓ Created package.json${NC}"
fi

# Create example index.js
if [[ ! -f "index.js" ]]; then
    cat > index.js << 'EOF'
// CloudBox Application Example
const { CloudBoxClient } = require('@ekoppen/cloudbox-sdk');

async function main() {
  // Initialize CloudBox client
  const cloudbox = new CloudBoxClient({
    projectId: process.env.CLOUDBOX_PROJECT_ID,
    apiKey: process.env.CLOUDBOX_API_KEY,
    endpoint: process.env.CLOUDBOX_ENDPOINT
  });

  // Test connection
  console.log('Testing CloudBox connection...');
  const connected = await cloudbox.testConnection();
  
  if (!connected) {
    console.error('Failed to connect to CloudBox');
    process.exit(1);
  }

  console.log('✓ Connected to CloudBox successfully');

  // Example: List collections
  try {
    const collections = await cloudbox.collections.list();
    console.log('Available collections:', collections.map(c => c.name));
  } catch (error) {
    console.log('No collections found or permission denied');
  }

  console.log('CloudBox setup complete! 🚀');
}

main().catch(console.error);
EOF
    echo -e "${GREEN}✓ Created index.js example${NC}"
fi

echo ""
echo -e "${GREEN}Setup completed!${NC}"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Install dependencies: npm install"
echo "2. Run the example: npm start"
echo "3. Check the SETUP_SCRIPTS_DOCUMENTATION.md for advanced features"
echo ""
echo -e "${GREEN}Happy coding! 🚀${NC}"