.PHONY: help install dev build clean frontend-install frontend-dev frontend-build frontend-clean backend-install backend-dev backend-build backend-clean

# Default target
help:
	@echo "Available commands:"
	@echo "  make install        - Install all dependencies (frontend + backend)"
	@echo "  make dev            - Run both frontend and backend in development mode"
	@echo "  make build          - Build both frontend and backend"
	@echo "  make clean          - Clean dependencies and build artifacts"
	@echo ""
	@echo "Frontend-specific commands:"
	@echo "  make frontend-install - Install frontend dependencies"
	@echo "  make frontend-dev     - Run frontend dev server"
	@echo "  make frontend-build   - Build frontend"
	@echo "  make frontend-clean   - Clean frontend artifacts"
	@echo ""
	@echo "Backend-specific commands:"
	@echo "  make backend-install  - Install backend dependencies"
	@echo "  make backend-dev      - Run backend dev server"
	@echo "  make backend-build    - Build backend"
	@echo "  make backend-clean    - Clean backend artifacts"

# Install all dependencies
install: frontend-install backend-install
	@echo "All dependencies installed successfully"

# Frontend targets
frontend-install:
	@echo "Installing frontend dependencies..."
	cd frontend && npm install

frontend-dev:
	@echo "Starting frontend dev server..."
	cd frontend && npm run dev -- --host

frontend-build:
	@echo "Building frontend..."
	cd frontend && npm run build

frontend-clean:
	@echo "Cleaning frontend..."
	cd frontend && rm -rf node_modules dist

# Backend targets
backend-install:
	@echo "Installing backend dependencies..."
	cd backend && go mod download

backend-dev:
	@echo "Starting backend dev server..."
	cd backend && go run main.go

backend-build:
	@echo "Building backend..."
	cd backend && go build -o bin/server main.go

backend-clean:
	@echo "Cleaning backend..."
	cd backend && rm -rf bin

# Run both in development mode (in parallel)
dev:
	@echo "Starting both frontend and backend..."
	@echo "Note: Run 'make frontend-dev' and 'make backend-dev' in separate terminals"
	@echo "Or use a process manager like 'concurrently' or 'tmux'"

# Build both projects
build: frontend-build backend-build
	@echo "Build completed successfully"

# Clean everything
clean: frontend-clean backend-clean
	@echo "Cleanup completed"
