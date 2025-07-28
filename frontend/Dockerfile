# Build stage
FROM node:20-alpine AS builder

# Set working directory
WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci

# Copy source code
COPY . .

# Build time environment variables for SvelteKit
ARG PUBLIC_API_URL
ENV PUBLIC_API_URL=$PUBLIC_API_URL

# Build the application
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

# Switch to non-root user for dependency installation
USER cloudbox

# Copy package files and install dependencies as non-root user
COPY --from=builder --chown=cloudbox:cloudbox /app/package*.json ./

# Install all dependencies (SvelteKit needs dev dependencies for runtime)
RUN npm ci

# Copy built application from builder stage
COPY --from=builder --chown=cloudbox:cloudbox /app/build ./

# Set production environment variables
ENV NODE_ENV=production
ENV HOST=0.0.0.0
ENV PORT=3000

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3000 || exit 1

# Run the SvelteKit application
CMD ["node", "index.js"]