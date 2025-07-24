# Build stage
FROM node:18-alpine AS builder

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
FROM node:18-alpine AS runner

# Create non-root user
RUN adduser -D -s /bin/sh cloudbox

# Set working directory
WORKDIR /app

# Copy built application from builder stage
COPY --from=builder /app/.svelte-kit/output ./
COPY --from=builder /app/package*.json ./

# Install all dependencies (SvelteKit needs dev dependencies for runtime)
RUN npm ci

# Change ownership to non-root user
RUN chown -R cloudbox:cloudbox /app

# Switch to non-root user
USER cloudbox

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3000 || exit 1

# Run the SvelteKit application
CMD ["node", "server/index.js"]