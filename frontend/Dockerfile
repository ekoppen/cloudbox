# Build stage
FROM node:20-alpine AS builder

# Set working directory
WORKDIR /app

# Copy package files and configuration
COPY package*.json ./
COPY svelte.config.js ./
COPY vite.config.ts ./
COPY tsconfig.json ./
COPY tsconfig.fallback.json ./

# Install dependencies
RUN npm ci

# Copy source code
COPY . .

# Build time environment variables for SvelteKit
ARG PUBLIC_API_URL
ENV PUBLIC_API_URL=$PUBLIC_API_URL

# Build the application
# Ensure .svelte-kit directory exists and handle fallback
RUN npm run prepare || echo "Prepare failed, continuing with fallback"
RUN if [ ! -f "./.svelte-kit/tsconfig.json" ]; then \
      echo "Using fallback tsconfig.json"; \
      cp tsconfig.fallback.json tsconfig.json; \
    fi
RUN npm run build

# Production stage
FROM node:20-alpine AS runner

# Install wget for health checks
RUN apk add --no-cache wget

# Create non-root user
RUN adduser -D -s /bin/sh cloudbox

# Set working directory and change ownership
WORKDIR /app
RUN chown cloudbox:cloudbox /app

# Copy the complete built application including node_modules from builder stage
COPY --from=builder --chown=cloudbox:cloudbox /app ./

# Switch to non-root user
USER cloudbox

# Set production environment variables
ENV NODE_ENV=production
ENV HOST=0.0.0.0
ENV PORT=3000

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3000 || exit 1

# Run the SvelteKit application
CMD ["sh", "-c", "cd /app && node build/index.js"]