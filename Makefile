.PHONY: help dev dev-all dev-down dev-all-down dev-all-reset dev-logs dev-status build test check clean web-deps web-install web-install-ci dev-server dev-server-restart dev-web build-dev-server build-backend test-backend build-go-backend test-go-backend migrate-go-backend build-frontend test-frontend test-e2e-frontend test-e2e-smoke-frontend build-web test-web typecheck-web lint-web db-reset namespace-smoke validate-release-config pr parallel-init parallel-sync parallel-up parallel-down docs-dev docs-build docs-preview

DEV_DIR := .dev
DEV_SERVER_PID := $(DEV_DIR)/server.pid
DEV_WEB_PID := $(DEV_DIR)/web.pid
DEV_SERVER_LOG := $(DEV_DIR)/server.log
DEV_WEB_LOG := $(DEV_DIR)/web.log
DEV_SERVER_BIN := $(abspath $(DEV_DIR)/skillhub-server)
DEV_WEB_URL := http://localhost:3000
DEV_API_URL := http://localhost:8080
DEV_PROCESS := bash scripts/dev-process.sh
DEV_BACKEND_DATABASE_URL := postgres://skillhub:skillhub_dev@localhost:5432/skillhub?sslmode=disable
DEV_BACKEND_STORAGE_PATH := $(abspath $(DEV_DIR)/packages)
PARALLEL_BASE_REF ?= origin/main
PARALLEL_WORKTREE_ROOT ?=
DEV_COMPOSE_PROJECT_NAME ?= skillhub
DEV_COMPOSE := docker compose -p $(DEV_COMPOSE_PROJECT_NAME)

help: ## Show available targets
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

dev: ## Start local dependencies (Postgres only)
	$(DEV_COMPOSE) up -d --wait --remove-orphans postgres
	@echo "Services ready."
	@echo "PostgreSQL: localhost:5432"
	@echo "Backend binary: make dev-server"
	@echo "Frontend: make dev-web"

dev-all: ## Start local dev environment (Postgres + backend binary + frontend)
	@mkdir -p $(DEV_DIR)
	$(DEV_COMPOSE) up -d --wait --remove-orphans postgres
	@mkdir -p "$(DEV_BACKEND_STORAGE_PATH)"
	@$(MAKE) migrate-go-backend
	@$(MAKE) build-dev-server
	@$(MAKE) web-deps
	@if $(DEV_PROCESS) status --pid-file $(DEV_SERVER_PID) >/dev/null 2>&1; then \
		echo "Backend already running with PID $$(cat $(DEV_SERVER_PID))"; \
	else \
		echo "Starting backend..."; \
		$(DEV_PROCESS) start --pid-file $(DEV_SERVER_PID) --log-file $(DEV_SERVER_LOG) --cwd backend -- /bin/sh -lc 'exec env DATABASE_URL="$(DEV_BACKEND_DATABASE_URL)" STORAGE_BASE_PATH="$(DEV_BACKEND_STORAGE_PATH)" BOOTSTRAP_ADMIN_ENABLED=true "$(DEV_SERVER_BIN)"' >/dev/null; \
	fi
	@if $(DEV_PROCESS) status --pid-file $(DEV_WEB_PID) >/dev/null 2>&1; then \
		echo "Frontend already running with PID $$(cat $(DEV_WEB_PID))"; \
	else \
		echo "Starting frontend..."; \
		$(DEV_PROCESS) start --pid-file $(DEV_WEB_PID) --log-file $(DEV_WEB_LOG) --cwd web -- pnpm exec vite --host 0.0.0.0 --strictPort >/dev/null; \
	fi
	@echo "Waiting for backend on $(DEV_API_URL) ..."
	@backend_ready=0; \
	for attempt in 1 2; do \
		for i in $$(seq 1 30); do \
			if curl -sf $(DEV_API_URL)/healthz >/dev/null; then \
				echo "Backend ready."; \
				backend_ready=1; \
				break 2; \
			fi; \
			if ! $(DEV_PROCESS) status --pid-file $(DEV_SERVER_PID) >/dev/null 2>&1; then \
				break; \
			fi; \
			sleep 2; \
		done; \
		if [ "$$attempt" -lt 2 ]; then \
			echo "Backend did not become ready on attempt $$attempt. Restarting..."; \
			$(DEV_PROCESS) stop --pid-file $(DEV_SERVER_PID); \
			sleep 2; \
			$(DEV_PROCESS) start --pid-file $(DEV_SERVER_PID) --log-file $(DEV_SERVER_LOG) --cwd backend -- /bin/sh -lc 'exec env DATABASE_URL="$(DEV_BACKEND_DATABASE_URL)" STORAGE_BASE_PATH="$(DEV_BACKEND_STORAGE_PATH)" BOOTSTRAP_ADMIN_ENABLED=true "$(DEV_SERVER_BIN)"' >/dev/null; \
		fi; \
	done; \
	if [ "$$backend_ready" -ne 1 ]; then \
		echo "Backend failed to become ready. Check $(DEV_SERVER_LOG)"; \
		exit 1; \
	fi
	@echo "Waiting for frontend on $(DEV_WEB_URL) ..."
	@frontend_ready=0; \
	for i in $$(seq 1 60); do \
		if curl -sf $(DEV_WEB_URL) >/dev/null; then \
			echo "Frontend ready."; \
			frontend_ready=1; \
			break; \
		fi; \
		sleep 2; \
	done; \
	if [ "$$frontend_ready" -ne 1 ]; then \
		echo "Frontend failed to become ready. Check $(DEV_WEB_LOG)"; \
		exit 1; \
	fi
	@echo "Local environment is ready:"
	@echo "  Web UI:  $(DEV_WEB_URL)"
	@echo "  Backend: $(DEV_API_URL)"
	@echo "  Bootstrap admin: admin / ChangeMe!2026"
	@echo "Logs:"
	@echo "  Backend: $(DEV_SERVER_LOG)"
	@echo "  Frontend: $(DEV_WEB_LOG)"

dev-server: ## Build and start the Go backend binary
	@mkdir -p "$(DEV_BACKEND_STORAGE_PATH)"
	@$(MAKE) migrate-go-backend
	@$(MAKE) build-dev-server
	exec env DATABASE_URL="$(DEV_BACKEND_DATABASE_URL)" STORAGE_BASE_PATH="$(DEV_BACKEND_STORAGE_PATH)" BOOTSTRAP_ADMIN_ENABLED=true "$(DEV_SERVER_BIN)"

dev-server-restart: ## Rebuild and restart the Go backend binary
	@mkdir -p $(DEV_DIR)
	@mkdir -p "$(DEV_BACKEND_STORAGE_PATH)"
	@$(MAKE) migrate-go-backend
	@$(MAKE) build-dev-server
	@$(DEV_PROCESS) stop --pid-file $(DEV_SERVER_PID)
	@$(DEV_PROCESS) start --pid-file $(DEV_SERVER_PID) --log-file $(DEV_SERVER_LOG) --cwd backend -- /bin/sh -lc 'exec env DATABASE_URL="$(DEV_BACKEND_DATABASE_URL)" STORAGE_BASE_PATH="$(DEV_BACKEND_STORAGE_PATH)" BOOTSTRAP_ADMIN_ENABLED=true "$(DEV_SERVER_BIN)"' >/dev/null
	@echo "Waiting for backend on $(DEV_API_URL) ..."
	@for i in $$(seq 1 30); do \
		if curl -sf $(DEV_API_URL)/healthz >/dev/null; then \
			echo "Backend ready."; \
			exit 0; \
		fi; \
		sleep 2; \
	done; \
	echo "Backend failed to become ready. Check $(DEV_SERVER_LOG)"; \
	exit 1

namespace-smoke: ## Run the namespace workflow smoke test
	./scripts/namespace-smoke-test.sh $(DEV_API_URL)

dev-down: ## Stop local Docker dependencies (Postgres only)
	$(DEV_COMPOSE) down --remove-orphans

dev-all-down: ## Stop the full local dev environment
	@$(DEV_PROCESS) stop --pid-file $(DEV_SERVER_PID)
	@$(DEV_PROCESS) stop --pid-file $(DEV_WEB_PID)
	@$(MAKE) dev-down

dev-all-reset: ## Reset the local dev environment from scratch
	@$(DEV_PROCESS) stop --pid-file $(DEV_SERVER_PID)
	@$(DEV_PROCESS) stop --pid-file $(DEV_WEB_PID)
	$(DEV_COMPOSE) down -v --remove-orphans
	rm -rf $(DEV_DIR)
	@$(MAKE) dev-all

dev-status: ## Show local development service status
	@echo "=== Dependency Services ==="
	@$(DEV_COMPOSE) ps
	@echo ""
	@echo "=== Backend ==="
	@if $(DEV_PROCESS) status --pid-file $(DEV_SERVER_PID) >/dev/null 2>&1; then \
		echo "  Running (PID $$(cat $(DEV_SERVER_PID)))"; \
	else \
		echo "  Not running"; \
	fi
	@echo "=== Frontend ==="
	@if $(DEV_PROCESS) status --pid-file $(DEV_WEB_PID) >/dev/null 2>&1; then \
		echo "  Running (PID $$(cat $(DEV_WEB_PID)))"; \
	else \
		echo "  Not running"; \
	fi

dev-logs: ## Tail development service logs (backend/frontend)
	@SERVICE=$${SERVICE:-backend}; \
	if [ "$$SERVICE" = "backend" ]; then \
		tail -f $(DEV_SERVER_LOG); \
	elif [ "$$SERVICE" = "frontend" ]; then \
		tail -f $(DEV_WEB_LOG); \
	else \
		echo "Unknown service: $$SERVICE. Use SERVICE=backend or SERVICE=frontend"; \
		exit 1; \
	fi

build-dev-server: ## Build the backend server binary used by local dev
	@mkdir -p $(DEV_DIR)
	cd backend && go build -o "$(DEV_SERVER_BIN)" ./cmd/server

build-go-backend: ## Build the Go backend
	cd backend && go build ./...

test-go-backend: ## Run Go backend tests
	cd backend && go test ./...

migrate-go-backend: ## Run Go backend database migrations
	cd backend && DATABASE_URL=postgres://skillhub:skillhub_dev@localhost:5432/skillhub?sslmode=disable go run ./cmd/migrate

build-backend: build-go-backend ## Build the backend (Go)

test-backend: test-go-backend ## Run backend tests (Go)

build: build-backend build-frontend ## Build backend and frontend

test: test-backend test-frontend ## Run backend and frontend tests

check: build test ## Run the full verification suite

clean: ## Clean build artifacts and local state
	$(DEV_COMPOSE) down -v
	rm -rf $(DEV_DIR)

web-install: ## Install frontend dependencies
	cd web && pnpm install

web-deps: ## Ensure frontend dependencies are available
	@if [ ! -d web/node_modules ]; then \
		echo "Installing frontend dependencies (node_modules missing)..."; \
		$(MAKE) web-install-ci; \
	elif [ ! -f web/node_modules/.modules.yaml ]; then \
		echo "Installing frontend dependencies (.modules.yaml missing)..."; \
		$(MAKE) web-install-ci; \
	elif [ web/pnpm-lock.yaml -nt web/node_modules/.modules.yaml ]; then \
		echo "Installing frontend dependencies (lockfile changed)..."; \
		$(MAKE) web-install-ci; \
	else \
		echo "Using existing frontend dependencies."; \
	fi

web-install-ci: ## Install frontend dependencies in CI mode
	cd web && CI=true pnpm install --frozen-lockfile

dev-web: ## Start the frontend dev server
	cd web && pnpm run dev

build-frontend: web-deps ## Build the frontend
	cd web && pnpm run build

test-frontend: web-deps ## Run frontend unit tests
	cd web && pnpm run test

test-e2e-frontend: web-deps ## Run frontend E2E tests (Playwright)
	cd web && pnpm run test:e2e

test-e2e-smoke-frontend: web-deps ## Run frontend smoke E2E tests (Playwright)
	cd web && pnpm run test:e2e:smoke

build-web: build-frontend ## Build the frontend

test-web: test-frontend ## Run frontend tests

typecheck-web: ## Run frontend type checking
	cd web && pnpm run typecheck

lint-web: ## Run frontend lint checks
	cd web && pnpm run lint

db-reset: ## Reset the local database
	$(DEV_COMPOSE) down -v --remove-orphans
	$(DEV_COMPOSE) up -d --wait --remove-orphans postgres
	$(MAKE) migrate-go-backend

validate-release-config: ## Validate release environment variables (.env.release by default)
	./scripts/validate-release-config.sh .env.release

pr: ## Push the current branch and create a pull request
	@if ! command -v gh >/dev/null 2>&1; then \
		echo "Error: gh CLI not found. Install from https://cli.github.com/"; \
		exit 1; \
	fi
	@if ! gh auth status >/dev/null 2>&1; then \
		echo "Error: gh CLI not authenticated. Run: gh auth login"; \
		exit 1; \
	fi
	@BRANCH=$$(git rev-parse --abbrev-ref HEAD); \
	if [ "$$BRANCH" = "main" ] || [ "$$BRANCH" = "master" ]; then \
		echo "Error: Cannot create PR from main/master branch."; \
		exit 1; \
	fi
	@if ! git diff --quiet || ! git diff --cached --quiet; then \
		echo "You have uncommitted changes:"; \
		git status --short; \
		echo ""; \
		printf "Commit all changes before creating PR? [y/N] "; \
		read -r answer; \
		if [ "$$answer" = "y" ] || [ "$$answer" = "Y" ]; then \
			git add -A; \
			git commit -m "chore: pre-PR commit"; \
		else \
			echo "Aborted. Commit or stash your changes first."; \
			exit 1; \
		fi; \
	fi
	@BRANCH=$$(git rev-parse --abbrev-ref HEAD); \
	echo "Pushing branch $$BRANCH to origin..."; \
	git push -u origin "$$BRANCH"
	@echo "Creating pull request..."
	@if gh pr view >/dev/null 2>&1; then \
		echo "A pull request already exists for this branch:"; \
		gh pr view --json url -q '.url'; \
		exit 0; \
	fi
	@gh pr create --fill --web || gh pr create --fill

parallel-init: ## Create parallel Claude/Codex/integration worktrees (TASK=<slug>)
	@if [ -z "$(TASK)" ]; then \
		echo "Usage: make parallel-init TASK=<task-slug> [PARALLEL_BASE_REF=origin/main] [PARALLEL_WORKTREE_ROOT=/path]"; \
		exit 1; \
	fi
	./scripts/parallel-init.sh "$(TASK)" "$(PARALLEL_BASE_REF)" "$(PARALLEL_WORKTREE_ROOT)"

parallel-sync: ## Merge Claude/Codex branches into the integration worktree
	PARALLEL_WORKTREE_ROOT="$(PARALLEL_WORKTREE_ROOT)" ./scripts/parallel-sync.sh $(SOURCES)

parallel-up: ## Merge branches and start the integration environment
	PARALLEL_WORKTREE_ROOT="$(PARALLEL_WORKTREE_ROOT)" ./scripts/parallel-up.sh $(SOURCES)

parallel-down: ## Stop the integration environment
	./scripts/parallel-down.sh

docs-dev: ## Start the docs development server
	cd docs/skillhub && npm run dev

docs-build: ## Build the docs site
	cd docs/skillhub && npm run build

docs-preview: ## Preview the built docs site
	cd docs/skillhub && npm run preview
