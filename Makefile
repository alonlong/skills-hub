# Local: `make backend` (terminal 1) + `make frontend` (terminal 2)
# Headless / CI: `BG=1 make backend` then `BG=1 make frontend` (logs under .dev/)
# Teardown: docker compose -p skillhub down; rm -f .dev/server.pid .dev/web.pid (kill those PIDs if still running)

.PHONY: backend frontend

DEV_DIR := .dev
DEV_SERVER_BIN := $(abspath $(DEV_DIR)/skillhub-server)
DEV_BACKEND_STORAGE_PATH := $(abspath $(DEV_DIR)/packages)
DATABASE_URL := postgres://skillhub:skillhub_dev@localhost:5432/skillhub?sslmode=disable
DEV_BACKEND_ENVS := DATABASE_URL="$(DATABASE_URL)" STORAGE_BASE_PATH="$(DEV_BACKEND_STORAGE_PATH)" BOOTSTRAP_ADMIN_ENABLED=true
DEV_COMPOSE := docker compose -p skillhub

backend:
	@bash -c 'set -e; \
	mkdir -p "$(DEV_BACKEND_STORAGE_PATH)"; \
	$(DEV_COMPOSE) up -d --wait --remove-orphans postgres; \
	cd backend && DATABASE_URL=$(DATABASE_URL) go run ./cmd/migrate; \
	cd backend && go build -o "$(DEV_SERVER_BIN)" ./cmd/server; \
	if [ "$(BG)" = "1" ]; then \
		nohup env $(DEV_BACKEND_ENVS) "$(DEV_SERVER_BIN)" > $(DEV_DIR)/server.log 2>&1 & echo $$! > $(DEV_DIR)/server.pid; \
		echo "Backend http://localhost:8080  log $(DEV_DIR)/server.log"; \
	else \
		exec env $(DEV_BACKEND_ENVS) "$(DEV_SERVER_BIN)"; \
	fi'

frontend:
	cd web && (test -d node_modules || pnpm install) && \
	if [ "$(BG)" = "1" ]; then \
		mkdir -p ../.dev && nohup pnpm exec vite --host 0.0.0.0 --strictPort > ../.dev/web.log 2>&1 & echo $$! > ../.dev/web.pid && echo "Frontend http://localhost:3000  log $(DEV_DIR)/web.log"; \
	else \
		pnpm run dev; \
	fi
