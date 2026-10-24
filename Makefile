SHELL=/bin/bash

MAKEFLAGS += -s

UID:=$(shell id -u)
GID:=$(shell id -g)
COMPOSE_CMD := HOST_UID=$(UID) HOST_GID=$(GID) docker compose -f local/compose.yaml --project-directory .
VOLUMES_DIR := ./.volumes

# Bold
BCYAN=\033[1;36m
BBLUE=\033[1;34m

# No color (Reset)
NC=\033[0m

.DEFAULT_GOAL := help

.PHONY: init
init: ## Initialize development environment
	mkdir -p $(VOLUMES_DIR)
	@echo -e "$(BCYAN)Initializing development environment...$(NC)"
	$(MAKE) gen-certs
	go mod tidy
	$(COMPOSE_CMD) build
	pre-commit install

.PHONY: gen-certs
gen-certs: ## Generate self-signed certificates
	@echo -e "$(BCYAN)Generating self-signed certificates...$(NC)"
	mkdir -p certs
	openssl req -x509 -newkey rsa:4096 -sha256 -days 365 -nodes -keyout certs/key.pem -out certs/cert.pem -subj "/CN=localhost" -addext "subjectAltName=DNS:localhost,IP:127.0.0.1"

.PHONY: deps
deps: ## Start service dependencies (postgres, adminer, dex)
	@echo -e "$(BCYAN)Starting service dependencies...$(NC)"
	$(COMPOSE_CMD) up -d postgres adminer dex
	@echo -e "$(BCYAN)Service dependencies started...$(NC)"

.PHONY: backend
backend: deps ## Run backend with hot reload
	@echo -e "$(BCYAN)Running backend with hot reload...$(NC)"
	$(COMPOSE_CMD) up api

.PHONY: ui
ui: backend ## Run UI with hot reload
	@echo -e "$(BCYAN)Running UI with hot reload...$(NC)"
	$(COMPOSE_CMD) up ui

.PHONY: dev
dev: ## Run full dev environment with hot reload
	@echo -e "$(BCYAN)Running dev environment with hot reload...$(NC)"
	go mod tidy
	$(COMPOSE_CMD) up --build

.PHONY: clean
clean: ## Clean up environment
	@echo -e "$(BCYAN)Cleaning up environment...$(NC)"
	$(COMPOSE_CMD) down -v --rmi local

.PHONY: help
help: ## Display this help
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(BCYAN)%-18s$(NC)%s\n", $$1, $$2}'
