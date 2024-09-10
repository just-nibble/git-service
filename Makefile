# Makefile for managing Docker and environment setup

# Variables
ENV_FILE = .env
ENV_SAMPLE_FILE = .env.sample
DOCKER_COMPOSE = docker compose
DOCKER_COMPOSE_CMD = $(DOCKER_COMPOSE) up --build

# Default target
all: copy-env up

# Copy the .env.sample file to .env if .env does not exist
copy-env:
	@if [ ! -f $(ENV_FILE) ]; then \
		cp $(ENV_SAMPLE_FILE) $(ENV_FILE); \
		echo "$(ENV_FILE) created from $(ENV_SAMPLE_FILE)"; \
	else \
		echo "$(ENV_FILE) already exists"; \
	fi

# Start Docker containers
up:
	$(DOCKER_COMPOSE_CMD)

# Stop Docker containers
down:
	$(DOCKER_COMPOSE) down

# Rebuild Docker images and start containers
restart: down up

# Remove all stopped containers and dangling images
clean:
	$(DOCKER_COMPOSE) down --rmi all --volumes --remove-orphans

# Help target to display available commands
help:
	@echo "Makefile commands:"
	@echo "  all       - Copy .env.sample to .env and start Docker containers"
	@echo "  copy-env  - Copy .env.sample to .env if .env does not exist"
	@echo "  up        - Start Docker containers"
	@echo "  down      - Stop Docker containers"
	@echo "  restart   - Restart Docker containers (stop and then start)"
	@echo "  clean     - Remove all stopped containers and dangling images"
	@echo "  help      - Show this help message"

.PHONY: all copy-env up down restart clean help
