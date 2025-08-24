#!/bin/bash

# Chess AI V3 - Complete Build & Deploy Script

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${GREEN}🏗️  Chess AI V3 - Build & Deploy${NC}"
echo "========================================="

# Configuration
BACKEND_DIR="./back"
FRONTEND_DIR="./front"
DOCKER_REGISTRY=${DOCKER_REGISTRY:-""}
VERSION=${VERSION:-"latest"}

# Functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check requirements
check_requirements() {
    log_info "Checking requirements..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker is required but not installed"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose is required but not installed"
        exit 1
    fi
    
    log_success "Requirements check passed"
}

# Build backend
build_backend() {
    log_info "Building Chess AI Backend..."
    
    if [ ! -d "$BACKEND_DIR" ]; then
        log_error "Backend directory not found: $BACKEND_DIR"
        exit 1
    fi
    
    cd "$BACKEND_DIR"
    
    # Test Go code first
    log_info "Running Go tests..."
    if ! go test ./... -v; then
        log_warning "Some tests failed, continuing anyway..."
    fi
    
    # Build Docker image
    log_info "Building backend Docker image..."
    sudo docker build -t chess-ai-backend:$VERSION .
    
    if [ ! -z "$DOCKER_REGISTRY" ]; then
        sudo docker tag chess-ai-backend:$VERSION $DOCKER_REGISTRY/chess-ai-backend:$VERSION
    fi
    
    cd - > /dev/null
    log_success "Backend build completed"
}

# Build frontend
build_frontend() {
    log_info "Building Chess AI Frontend..."
    
    if [ ! -d "$FRONTEND_DIR" ]; then
        log_warning "Frontend directory not found: $FRONTEND_DIR"
        log_warning "Skipping frontend build"
        return
    fi
    
    cd "$FRONTEND_DIR"
    
    # Check if package.json exists
    if [ ! -f "package.json" ]; then
        log_error "package.json not found in frontend directory"
        exit 1
    fi
    
    # Build Docker image
    log_info "Building frontend Docker image..."
    sudo docker build -t chess-ai-frontend:$VERSION .
    
    if [ ! -z "$DOCKER_REGISTRY" ]; then
        sudo docker tag chess-ai-frontend:$VERSION $DOCKER_REGISTRY/chess-ai-frontend:$VERSION
    fi
    
    cd - > /dev/null
    log_success "Frontend build completed"
}

# Run development environment
run_dev() {
    log_info "Starting development environment..."
    
    # Stop any existing containers
    docker-compose down 2>/dev/null || true
    
    # Start services
    docker-compose up -d
    
    log_success "Development environment started"
    log_info "Backend: http://localhost:8080"
    log_info "Frontend: http://localhost:3000"
    
    # Wait for services to be ready
    log_info "Waiting for services to start..."
    
    # Wait for backend
    for i in {1..30}; do
        if curl -s http://localhost:8080/health > /dev/null 2>&1; then
            log_success "Backend is ready"
            break
        fi
        sleep 2
        if [ $i -eq 30 ]; then
            log_error "Backend failed to start"
            exit 1
        fi
    done
    
    # Wait for frontend (if exists)
    if [ -d "$FRONTEND_DIR" ]; then
        for i in {1..30}; do
            if curl -s http://localhost:3000 > /dev/null 2>&1; then
                log_success "Frontend is ready"
                break
            fi
            sleep 2
            if [ $i -eq 30 ]; then
                log_warning "Frontend might not be ready yet"
                break
            fi
        done
    fi
}

# Run production environment
run_prod() {
    log_info "Starting production environment..."
    
    # Stop any existing containers
    docker-compose --profile production down 2>/dev/null || true
    
    # Start services with production profile
    docker-compose --profile production up -d
    
    log_success "Production environment started"
    log_info "Access via: http://localhost"
}

# Push to registry
push_images() {
    if [ -z "$DOCKER_REGISTRY" ]; then
        log_warning "No Docker registry specified, skipping push"
        return
    fi
    
    log_info "Pushing images to registry..."
    
    docker push $DOCKER_REGISTRY/chess-ai-backend:$VERSION
    
    if [ -d "$FRONTEND_DIR" ]; then
        docker push $DOCKER_REGISTRY/chess-ai-frontend:$VERSION
    fi
    
    log_success "Images pushed to registry"
}

# Clean up
cleanup() {
    log_info "Cleaning up..."
    
    # Remove unused images
    docker image prune -f
    
    log_success "Cleanup completed"
}

# Show logs
show_logs() {
    docker-compose logs -f
}

# Stop everything
stop_all() {
    log_info "Stopping all services..."
    docker-compose down
    docker-compose --profile production down 2>/dev/null || true
    log_success "All services stopped"
}

# Test the deployment
test_deployment() {
    log_info "Testing deployment..."
    
    # Test backend health
    if curl -s http://localhost:8080/health | grep -q "healthy"; then
        log_success "Backend health check passed"
    else
        log_error "Backend health check failed"
        return 1
    fi
    
    # Test backend API
    if curl -s http://localhost:8080/api/game | grep -q "board"; then
        log_success "Backend API test passed"
    else
        log_error "Backend API test failed"
        return 1
    fi
    
    # Test frontend (if running)
    if curl -s http://localhost:3000 > /dev/null 2>&1; then
        log_success "Frontend accessibility test passed"
    else
        log_warning "Frontend not accessible (might be normal if not built)"
    fi
    
    log_success "Deployment tests completed"
}

# Help function
show_help() {
    echo "Chess AI V3 Build Script"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  build     - Build both backend and frontend"
    echo "  backend   - Build only backend"
    echo "  frontend  - Build only frontend"
    echo "  dev       - Run development environment"
    echo "  prod      - Run production environment"
    echo "  test      - Test the deployment"
    echo "  logs      - Show service logs"
    echo "  push      - Push images to registry"
    echo "  stop      - Stop all services"
    echo "  clean     - Clean up unused images"
    echo "  help      - Show this help"
    echo ""
    echo "Environment Variables:"
    echo "  DOCKER_REGISTRY - Docker registry URL for pushing images"
    echo "  VERSION        - Image version tag (default: latest)"
    echo ""
    echo "Examples:"
    echo "  $0 build                           # Build everything"
    echo "  $0 dev                             # Start development"
    echo "  DOCKER_REGISTRY=myregistry.com $0 push  # Push to registry"
}

# Main script logic
case "${1:-help}" in
    "build")
        check_requirements
        build_backend
        build_frontend
        log_success "🎉 Build completed successfully!"
        ;;
    "backend")
        check_requirements
        build_backend
        ;;
    "frontend")
        check_requirements
        build_frontend
        ;;
    "dev")
        check_requirements
        build_backend
        if [ -d "$FRONTEND_DIR" ]; then
            build_frontend
        fi
        run_dev
        ;;
    "prod")
        check_requirements
        run_prod
        ;;
    "test")
        test_deployment
        ;;
    "logs")
        show_logs
        ;;
    "push")
        push_images
        ;;
    "stop")
        stop_all
        ;;
    "clean")
        cleanup
        ;;
    "help"|*)
        show_help
        ;;
esac