.PHONY: dev build docker docker-down clean frontend-build

# Development
dev:
	@echo "Starting development..."
	cd frontend && npm run dev &
	go run ./cmd/tars

# Build frontend
frontend-build:
	cd frontend && npm ci && npm run build
	rm -rf web/*
	cp -r frontend/build/* web/

# Build Go binary (with embedded frontend)
build: frontend-build
	go build -o bin/tars ./cmd/tars

# Docker
docker:
	docker compose up --build

docker-down:
	docker compose down -v

# Clean
clean:
	rm -rf bin/ frontend/build frontend/node_modules frontend/.svelte-kit
