# Build stage: Frontend + Backend
FROM golang:1.25.7 AS builder

WORKDIR /src

# Install Node.js (for frontend build)
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    && rm -rf /var/lib/apt/lists/*

RUN curl -fsSL https://deb.nodesource.com/setup_18.x | bash - \
    && apt-get install -y nodejs

# Download Go dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Build Vue frontend
WORKDIR /src/pkg/web/frontend
RUN npm install && npm run build

# Build Go binary (which now embeds the frontend dist/)
WORKDIR /src
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/kakoclaw ./cmd/kakoclaw

# Runtime stage
FROM debian:bookworm-slim

RUN apt-get update \
  && apt-get install -y --no-install-recommends ca-certificates \
  && rm -rf /var/lib/apt/lists/* \
  && useradd -m -u 10001 -s /bin/bash kakoclaw

COPY --from=builder /out/kakoclaw /usr/local/bin/kakoclaw

USER kakoclaw
ENV HOME=/home/kakoclaw

EXPOSE 18880

CMD ["kakoclaw", "web"]
