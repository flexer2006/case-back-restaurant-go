DOCKER_COMPOSE = docker-compose
PROJECT_NAME = restaurant-booking
ENV_FILE = .env


.PHONY: all build up down restart logs ps clean help migrate-up migrate-down test server-run server-build run-all check-db

all: build up


build:
	$(DOCKER_COMPOSE) build


up:
	$(DOCKER_COMPOSE) up -d


down:
	$(DOCKER_COMPOSE) down


restart:
	$(DOCKER_COMPOSE) restart


logs:
	$(DOCKER_COMPOSE) logs -f


ps:
	$(DOCKER_COMPOSE) ps


clean:
	$(DOCKER_COMPOSE) down --volumes --remove-orphans
	docker system prune -f


server-build:
	go build -o ./bin/server ./cmd/server


server-run: server-build
	./bin/server


check-db:
	@echo "Checking database connection..."
	@docker exec postgres-container pg_isready -U postgres -d postgres || echo "Database is not ready"


run-all: up server-build
	@echo "Starting PostgreSQL container..."
	@sleep 8  # Wait for PostgreSQL to fully initialize
	@echo "Checking if PostgreSQL is ready..."
	@docker-compose ps | grep db | grep "healthy" || sleep 10
	@echo "Checking database connection directly..."
	@make check-db
	@echo "Starting local server..."
	@echo "Using POSTGRES_HOST=127.0.0.1 for local server..."
	POSTGRES_HOST=127.0.0.1 ./bin/server


migrate-up:
	go run ./cmd/server migrate up


migrate-down:
	go run ./cmd/server migrate down


test:
	go test ./...


health-check:
	curl -f http://$(shell grep SERVER_HOST $(ENV_FILE) | cut -d= -f2 || echo 0.0.0.0):$$(grep SERVER_PORT $(ENV_FILE) | cut -d= -f2 || echo 8080)/health || echo "Server is not healthy"


help:
	@echo "Available commands:"
	@echo "  make build         - Build Docker images for PostgreSQL"
	@echo "  make up            - Start PostgreSQL container"
	@echo "  make down          - Stop PostgreSQL container"
	@echo "  make restart       - Restart PostgreSQL container"
	@echo "  make logs          - View logs"
	@echo "  make ps            - Container status"
	@echo "  make clean         - Remove containers, images and volumes"
	@echo "  make server-build  - Build the server locally"
	@echo "  make server-run    - Build and run the server locally"
	@echo "  make run-all       - Start both PostgreSQL container and local server"
	@echo "  make check-db      - Check if database is ready"
	@echo "  make migrate-up    - Apply migrations locally"
	@echo "  make migrate-down  - Rollback migrations locally"
	@echo "  make test          - Run tests locally"
	@echo "  make health-check  - Check server health"
	@echo "  make help          - List available commands"