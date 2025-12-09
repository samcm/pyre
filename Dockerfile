# Stage 1: Build Frontend
FROM node:22-alpine AS frontend-builder
WORKDIR /app/frontend

# Install pnpm
RUN corepack enable && corepack prepare pnpm@latest --activate

# Install dependencies
COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

# Copy source and build
COPY frontend/ ./
RUN pnpm build

# Stage 2: Build Backend
FROM golang:1.23-alpine AS backend-builder
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY backend/go.mod backend/go.sum ./backend/
WORKDIR /app/backend
RUN go mod download

# Copy backend source
WORKDIR /app
COPY backend/ ./backend/

# Copy frontend dist for embedding
COPY --from=frontend-builder /app/frontend/dist ./backend/frontend/dist

# Build binary
WORKDIR /app/backend
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /pyre ./cmd/server

# Stage 3: Final Image
FROM alpine:3.20
WORKDIR /app

# Install CA certificates for HTTPS requests
RUN apk add --no-cache ca-certificates tzdata

# Copy binary
COPY --from=backend-builder /pyre /app/pyre

# Create data directory
RUN mkdir -p /app/data

# Expose port
EXPOSE 8080

# Volume for config and data
VOLUME ["/app/config", "/app/data"]

# Run
ENTRYPOINT ["/app/pyre"]
CMD ["--config", "/app/config/config.yaml"]
