# syntax=docker/dockerfile:1

# ---- Stage 1: Build frontend assets with Vite ----
FROM node:20-alpine AS frontend-build
WORKDIR /app
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# ---- Stage 2: Build Go binary ----
FROM golang:1.25-alpine AS backend-build
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server ./cmd/web

# ---- Stage 3: Minimal runtime image ----
FROM alpine:3.20
WORKDIR /app
COPY --from=backend-build /out/server ./server
COPY --from=backend-build /app/config.json ./config.json
COPY --from=frontend-build /app/dist ./dist
# In production, Go serves the built SPA from ./dist
ENV WEB_STATIC_DIR=./dist
EXPOSE 8080
CMD ["./server"]
