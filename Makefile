.PHONY: help dev dev-build build test migrate seed logs shell-db shell-backend clean

# ── Config ───────────────────────────────────────────────────────────────────
COMPOSE        := docker compose
COMPOSE_PROD   := $(COMPOSE) -f docker-compose.yml
COMPOSE_DEV    := $(COMPOSE) -f docker-compose.yml -f docker-compose.dev.yml

# ── Default ───────────────────────────────────────────────────────────────────
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	  awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

# ── Development ───────────────────────────────────────────────────────────────
dev: ## Start all services in dev mode (hot reload)
	$(COMPOSE_DEV) up --build

dev-build: ## Build dev images without starting
	$(COMPOSE_DEV) build

# ── Production ────────────────────────────────────────────────────────────────
build: ## Build production images
	$(COMPOSE_PROD) build

up: ## Start production stack (detached)
	$(COMPOSE_PROD) up -d

down: ## Stop and remove containers
	$(COMPOSE_PROD) down

# ── Testing ───────────────────────────────────────────────────────────────────
test: ## Run Go tests
	cd backend && go test ./...

test-verbose: ## Run Go tests with verbose output
	cd backend && go test -v ./...

# ── Database ──────────────────────────────────────────────────────────────────
migrate: ## Run database migrations (requires running db service)
	@echo "Running migrations..."
	$(COMPOSE_PROD) run --rm backend ./server --migrate-only || \
	  DATABASE_URL=$$($(COMPOSE_PROD) exec -T db sh -c 'echo "postgres://$$POSTGRES_USER:$$POSTGRES_PASSWORD@localhost:5432/$$POSTGRES_DB?sslmode=disable"') \
	  go run ./backend/cmd/server --migrate-only

seed: ## Seed the development database
	$(COMPOSE_DEV) exec db sh /scripts/seed.sh

migrate-local: ## Run migrations against local DB (DATABASE_URL must be set)
	cd backend && go run ./cmd/server --migrate-only

# ── Utilities ─────────────────────────────────────────────────────────────────
logs: ## Tail all service logs
	$(COMPOSE_PROD) logs -f

logs-backend: ## Tail backend logs
	$(COMPOSE_PROD) logs -f backend

logs-frontend: ## Tail frontend logs
	$(COMPOSE_PROD) logs -f frontend

shell-db: ## Open a psql shell in the db container
	$(COMPOSE_PROD) exec db psql -U $${POSTGRES_USER:-tars} -d $${POSTGRES_DB:-tars}

shell-backend: ## Open a shell in the backend container
	$(COMPOSE_PROD) exec backend sh

clean: ## Remove containers, volumes, and built images
	$(COMPOSE_PROD) down -v --rmi local
	$(COMPOSE_DEV) down -v --rmi local 2>/dev/null || true
