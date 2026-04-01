# =============================================================================
# Multi-stage build for MakoClaw - AI Agent Framework
# =============================================================================

# -----------------------------------------------------------------------------
# Stage 1: Frontend Build
# -----------------------------------------------------------------------------
FROM node:18-slim AS frontend-builder

WORKDIR /src/pkg/web/frontend

# Copy package files first for better layer caching
COPY pkg/web/frontend/package*.json ./

# Install dependencies
RUN npm ci

# Copy frontend source
COPY pkg/web/frontend/ ./

# Build frontend (outputs to pkg/web/dist)
RUN npm run build

# -----------------------------------------------------------------------------
# Stage 2: Backend Build
# -----------------------------------------------------------------------------
FROM golang:1.26 AS backend-builder

WORKDIR /src

# Copy Go module files first for better caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Copy built frontend from stage 1
COPY --from=frontend-builder /src/pkg/web/dist ./pkg/web/dist

# Build binary with security and optimization flags
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath \
    -ldflags "-s -w -X main.Version=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')" \
    -o /out/makoclaw ./cmd/makoclaw

# Copy built-in skills
RUN mkdir -p /out/skills && cp -r skills/* /out/skills/ 2>/dev/null || true

# -----------------------------------------------------------------------------
# Stage 3: Runtime
# -----------------------------------------------------------------------------
FROM debian:bookworm-slim

# Add metadata
LABEL org.opencontainers.image.title="MakoClaw" \
      org.opencontainers.image.description="Ultra-efficient Go AI agent framework" \
      org.opencontainers.image.vendor="Sipeed" \
      org.opencontainers.image.source="https://github.com/sipeed/makoclaw"

# Install runtime dependencies
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    ca-certificates \
    tzdata \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Copy Node.js runtime from the frontend-builder stage (includes npm)
COPY --from=frontend-builder /usr/local/bin/node /usr/local/bin/node
COPY --from=frontend-builder /usr/local/lib/node_modules /usr/local/lib/node_modules
RUN ln -s /usr/local/lib/node_modules/npm/bin/npm-cli.js /usr/local/bin/npm && \
    ln -s /usr/local/lib/node_modules/npm/bin/npx-cli.js /usr/local/bin/npx

# Create non-root user
RUN useradd -m -u 10001 -s /bin/bash makoclaw && \
    mkdir -p /home/makoclaw/.MakoClaw/workspace && \
    chown -R makoclaw:makoclaw /home/makoclaw

# Install Claude Code CLI globally (required by Dev Studio bridge SDK)
RUN npm install -g @anthropic-ai/claude-code && \
    chown -R makoclaw:makoclaw /usr/local/lib/node_modules/@anthropic-ai

# Copy binary and skills
COPY --from=backend-builder --chown=makoclaw:makoclaw /out/makoclaw /usr/local/bin/makoclaw
COPY --from=backend-builder /out/skills /usr/local/skills/

# Switch to non-root user
USER makoclaw
WORKDIR /home/makoclaw

# Set environment variables
ENV HOME=/home/makoclaw \
    PATH=/usr/local/bin:$PATH \
    MAKOCLAW_WEB_HOST=0.0.0.0 \
    MAKOCLAW_WEB_PORT=18880

# Expose port
EXPOSE 18880

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:18880/health || exit 1

# Default command
CMD ["makoclaw", "web"]
