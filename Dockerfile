# Stage 1: Build Frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install --legacy-peer-deps
COPY frontend/ ./
RUN npm run build

# Stage 2: Build Backend
FROM golang:1.24-alpine AS backend-builder
# Add build dependencies for libvips
RUN apk add --no-cache vips-dev build-base
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Ensure go.mod is tidy in the container context
RUN go mod tidy
# Enable CGO for govips
ENV CGO_ENABLED=1
RUN go build -o server ./cmd/server/main.go

# Stage 3: Final Runtime
FROM alpine:latest
# Runtime dependency for vips
RUN apk add --no-cache ca-certificates tzdata vips
WORKDIR /app
COPY --from=backend-builder /app/server .
COPY --from=frontend-builder /app/frontend/dist ./public
COPY internal/data/schema.sql ./internal/data/schema.sql

ENV PORT=8080
ENV LIBRARY_PATH=/library
ENV CACHE_DIR=/cache
ENV DB_PATH=/cache/archive.db

EXPOSE 8080
CMD ["./server"]
