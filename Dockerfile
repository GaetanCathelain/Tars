# Stage 1: Build frontend
FROM node:22-alpine AS frontend-builder
WORKDIR /frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# Stage 2: Build Go backend with embedded frontend
FROM golang:1.24-alpine AS backend-builder
WORKDIR /build
ENV GOTOOLCHAIN=auto
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Copy built frontend into web/ for Go embed
COPY --from=frontend-builder /frontend/build/ ./web/
RUN CGO_ENABLED=0 go build -o tars ./cmd/tars

# Stage 3: Minimal runtime
FROM alpine:3.21
RUN apk add --no-cache ca-certificates tini
# Note: claude CLI will need to be mounted/installed separately for worker functionality
COPY --from=backend-builder /build/tars /usr/local/bin/tars
COPY --from=backend-builder /build/migrations/ /migrations/
EXPOSE 3333
ENTRYPOINT ["tini", "--"]
CMD ["tars"]
