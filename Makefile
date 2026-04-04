# Makefile
.PHONY: help build up down logs clean test

help:
	@echo "Available commands:"
	@echo "  make build    - Build all services"
	@echo "  make up       - Start all services (Ethereum + P4 + Nexus)"
	@echo "  make down     - Stop all services"
	@echo "  make logs     - View logs"
	@echo "  make test     - Run performance test"
	@echo "  make clean    - Clean build artifacts"

build:
	@echo "📦 Building services..."
	docker-compose build
	@echo "✅ Build complete"

up:
	@echo "🚀 Starting all services..."
	docker-compose up -d
	@echo "✅ Services started"
	@echo ""
	@echo "📊 Service ports:"
	@echo "   Ethereum (Anvil): http://localhost:8545"
	@echo "   P4 Switch:        localhost:9090"
	@echo "   Validator:        localhost:8001"
	@echo "   Multiplexer:      localhost:8002"
	@echo ""
	@echo "💡 To run demo: make test"

down:
	@echo "🛑 Stopping all services..."
	docker-compose down
	@echo "✅ Services stopped"

logs:
	docker-compose logs -f

test:
	@echo "🏃 Running performance test..."
	@echo "This will:"
	@echo "  1. Send 5 transactions via Ethereum L1 (Anvil)"
	@echo "  2. Send 5 payments via Nexus L2"
	@echo "  3. Compare results"
	docker exec nexus-app ./bin/demo

clean:
	@echo "🧹 Cleaning..."
	docker-compose down -v
	rm -rf bin/ logs/
	@echo "✅ Clean complete"