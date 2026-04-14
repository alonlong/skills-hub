.PHONY: backend frontend dev

# Matches docker-compose.yml host port mapping postgres: "7432:5432"
POSTGRES_HOST_PORT ?= 7432

# Vite dev server and Go API must agree on the API port (see web/vite.config.ts).
BACKEND_HTTP_PORT ?= 9000
FRONTEND_PORT ?= 3010
export BACKEND_HTTP_PORT FRONTEND_PORT

backend:
	@bash -c 'set -e; set -o pipefail; \
	mkdir -p .dev; \
	go build -o bin/skillhub ./cmd/server; \
	export DATABASE_URL="$${DATABASE_URL:-postgres://postgres:123456@localhost:$(POSTGRES_HOST_PORT)/skillhub?sslmode=disable}"; \
	export BOOTSTRAP_ADMIN_ENABLED="$${BOOTSTRAP_ADMIN_ENABLED:-true}"; \
	if [ "$(BG)" = "1" ]; then \
		nohup ./bin/skillhub >> .dev/backend.log 2>&1 & echo $$! > .dev/backend.pid && echo "Backend http://localhost:$${BACKEND_HTTP_PORT}  log .dev/backend.log  pid $$(cat .dev/backend.pid)"; \
	else \
		echo "Backend http://localhost:$${BACKEND_HTTP_PORT}  log .dev/backend.log (tee)"; \
		./bin/skillhub 2>&1 | tee -a .dev/backend.log; \
	fi'

# Starts API on BACKEND_HTTP_PORT; Vite proxies /api and /oauth2 there (same env vars).
frontend:
	@cd web && (test -x node_modules/.bin/vite || pnpm install) && \
	if [ "$(BG)" = "1" ]; then \
		mkdir -p ../.dev && nohup pnpm exec vite --host 0.0.0.0 --strictPort > ../.dev/web.log 2>&1 & echo $$! > ../.dev/web.pid && echo "Frontend http://localhost:$${FRONTEND_PORT}  proxy /api -> http://127.0.0.1:$${BACKEND_HTTP_PORT}  log .dev/web.log"; \
	else \
		echo "Frontend http://localhost:$${FRONTEND_PORT}  proxy /api -> http://127.0.0.1:$${BACKEND_HTTP_PORT}"; \
		pnpm run dev; \
	fi

# Full stack: API in background, wait for /healthz, then Vite (foreground) with matching proxy.
dev:
	@$(MAKE) backend BG=1
	@bash -c 'set -e; for i in $$(seq 1 40); do curl -sf "http://127.0.0.1:$${BACKEND_HTTP_PORT:-3001}/healthz" >/dev/null && exit 0; sleep 0.25; done; echo "error: backend not ready on http://127.0.0.1:$${BACKEND_HTTP_PORT:-3001} — start Postgres (e.g. docker compose up -d) and retry"; exit 1'
	@$(MAKE) frontend
