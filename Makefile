.PHONY: help
help:
	@echo 'Makefile for `golembic-blend` project'
	@echo ''
	@echo 'Usage:'
	@echo '   make dev-deps                Install (or upgrade) development time dependencies'
	@echo '   make vet                     Run `go vet` over source tree'
	@echo '   make shellcheck              Run `shellcheck` on all shell files in `./_bin/`'
	@echo 'PostgreSQL-specific Targets:'
	@echo '   make start-postgres          Starts a PostgreSQL database running in a Docker container and set up users'
	@echo '   make stop-postgres           Stops the PostgreSQL database running in a Docker container'
	@echo '   make restart-postgres        Stops the PostgreSQL database (if running) and starts a fresh Docker container'
	@echo '   make require-postgres        Determine if PostgreSQL database is running; fail if not'
	@echo '   make psql                    Connects to currently running PostgreSQL DB via `psql`'
	@echo '   make psql-superuser          Connects to currently running PostgreSQL DB via `psql` as superuser'
	@echo ''

################################################################################
# Meta-variables
################################################################################
SHELLCHECK_PRESENT := $(shell command -v shellcheck 2> /dev/null)

################################################################################
# Environment variable defaults
################################################################################
DB_HOST ?= 127.0.0.1
DB_SSLMODE ?= disable
DB_NETWORK_NAME ?= dev-network-golembic-blend

POSTGRES_PORT ?= 23396
POSTGRES_CONTAINER_NAME ?= dev-postgres-golembic-blend

DB_SUPERUSER_NAME ?= superuser_db
DB_SUPERUSER_USER ?= superuser
DB_SUPERUSER_PASSWORD ?= testpassword_superuser

DB_NAME ?= golembic
DB_ADMIN_USER ?= golembic_admin
DB_ADMIN_PASSWORD ?= testpassword_admin

# NOTE: This assumes the `DB_*_PASSWORD` values do not need to be URL encoded.
POSTGRES_SUPERUSER_DSN ?= postgres://$(DB_SUPERUSER_USER):$(DB_SUPERUSER_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)
POSTGRES_ADMIN_DSN ?= postgres://$(DB_ADMIN_USER):$(DB_ADMIN_PASSWORD)@$(DB_HOST):$(POSTGRES_PORT)/$(DB_NAME)

################################################################################
# Targets
################################################################################

.PHONY: dev-deps
dev-deps:
	go mod download

.PHONY: vet
vet:
	go vet ./...

.PHONY: _require-shellcheck
_require-shellcheck:
ifndef SHELLCHECK_PRESENT
	$(error 'shellcheck is not installed, it can be installed via "apt-get install shellcheck" or "brew install shellcheck".')
endif

.PHONY: shellcheck
shellcheck: _require-shellcheck
	shellcheck --exclude SC1090,SC1091 ./_bin/*.sh

.PHONY: start-postgres
start-postgres:
	@DB_NETWORK_NAME=$(DB_NETWORK_NAME) \
	  DB_CONTAINER_NAME=$(POSTGRES_CONTAINER_NAME) \
	  DB_HOST=$(DB_HOST) \
	  DB_PORT=$(POSTGRES_PORT) \
	  DB_SUPERUSER_NAME=$(DB_SUPERUSER_NAME) \
	  DB_SUPERUSER_USER=$(DB_SUPERUSER_USER) \
	  DB_SUPERUSER_PASSWORD=$(DB_SUPERUSER_PASSWORD) \
	  DB_NAME=$(DB_NAME) \
	  DB_ADMIN_USER=$(DB_ADMIN_USER) \
	  DB_ADMIN_PASSWORD=$(DB_ADMIN_PASSWORD) \
	  ./_bin/start_postgres.sh

.PHONY: stop-postgres
stop-postgres:
	@DB_NETWORK_NAME=$(DB_NETWORK_NAME) \
	  DB_CONTAINER_NAME=$(POSTGRES_CONTAINER_NAME) \
	  ./_bin/stop_db.sh

.PHONY: restart-postgres
restart-postgres: stop-postgres start-postgres

.PHONY: require-postgres
require-postgres:
	@DB_HOST=$(DB_HOST) \
	  DB_PORT=$(POSTGRES_PORT) \
	  DB_ADMIN_DSN=$(POSTGRES_ADMIN_DSN) \
	  ./_bin/require_postgres.sh

.PHONY: psql
psql: require-postgres
	@echo "Running psql against port $(POSTGRES_PORT)"
	psql "$(POSTGRES_ADMIN_DSN)"

.PHONY: psql-superuser
psql-superuser: require-postgres
	@echo "Running psql against port $(POSTGRES_PORT)"
	psql "$(POSTGRES_SUPERUSER_DSN)"
