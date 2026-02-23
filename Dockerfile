# Build stage 1: Frontend
FROM node:18-slim AS frontend-builder
WORKDIR /src/pkg/web/frontend
COPY pkg/web/frontend/package*.json ./
RUN npm install
COPY pkg/web/frontend/ ./
RUN npm run build

# Build stage 2: Backend
FROM golang:1.26.0 AS backend-builder
WORKDIR /src
# Copy Go modules manifests
COPY go.mod go.sum ./
RUN go mod download
# Copy entire source
COPY . .
# Copy built frontend from stage 1
# Vite output outDir: '../dist' relative to pkg/web/frontend means pkg/web/dist
COPY --from=frontend-builder /src/pkg/web/dist ./pkg/web/dist
# Build the backend (Go 1.16+ embeds are picked up)
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/makoclaw ./cmd/makoclaw

# Final stage: Runtime
FROM debian:bookworm-slim
RUN apt-get update \
  && apt-get install -y --no-install-recommends ca-certificates \
  && rm -rf /var/lib/apt/lists/* \
  && useradd -m -u 10001 -s /bin/bash makoclaw

COPY --from=backend-builder /out/makoclaw /usr/local/bin/makoclaw

ENV HOME=/home/makoclaw
EXPOSE 18880
CMD ["makoclaw", "web"]
