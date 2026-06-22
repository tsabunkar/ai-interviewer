# Makefile for AI Interviewer backend

.PHONY: docs
docs:
	go run scripts/generate-swagger.go

.PHONY: server
server:
	go run cmd/server/main.go

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: docker-up
docker-up:
	docker compose up --build

.PHONY: docker-down
docker-down:
	docker compose down

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  docs        - Generate Swagger documentation"
	@echo "  server      - Start local development server"
	@echo "  docker-up   - Start all services with Docker Compose"
	@echo "  docker-down - Stop all Docker Compose services"
	@echo "  tidy        - Tidy go modules"
	@echo "  help        - Show this help message"