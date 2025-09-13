#!/bin/bash

# SmartEdify Auth Service Initialization Script
# This script sets up the development environment

set -e

echo "🚀 Initializing SmartEdify Auth Service..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if required tools are installed
check_requirements() {
    print_status "Checking requirements..."
    
    # Check Go
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.21 or higher."
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    print_success "Go $GO_VERSION is installed"
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker."
        exit 1
    fi
    
    # Check Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose is not installed. Please install Docker Compose."
        exit 1
    fi
    
    print_success "Docker and Docker Compose are installed"
    
    # Check Make (optional)
    if ! command -v make &> /dev/null; then
        print_warning "Make is not installed. You can still use the project but won't be able to use Makefile commands."
    else
        print_success "Make is installed"
    fi
}

# Setup environment file
setup_env() {
    print_status "Setting up environment configuration..."
    
    if [ ! -f .env ]; then
        cp .env.example .env
        print_success "Created .env file from .env.example"
        print_warning "Please edit .env file with your configuration before running the service"
    else
        print_warning ".env file already exists, skipping..."
    fi
}

# Download Go dependencies
setup_dependencies() {
    print_status "Downloading Go dependencies..."
    
    go mod download
    go mod tidy
    
    print_success "Go dependencies downloaded"
}

# Start infrastructure services
start_infrastructure() {
    print_status "Starting infrastructure services (PostgreSQL, Redis, Jaeger, Prometheus)..."
    
    # Start only infrastructure services
    docker-compose up -d postgres redis jaeger prometheus grafana
    
    print_status "Waiting for services to be ready..."
    sleep 10
    
    # Check if PostgreSQL is ready
    until docker-compose exec -T postgres pg_isready -U postgres; do
        print_status "Waiting for PostgreSQL..."
        sleep 2
    done
    
    # Check if Redis is ready
    until docker-compose exec -T redis redis-cli ping; do
        print_status "Waiting for Redis..."
        sleep 2
    done
    
    print_success "Infrastructure services are ready"
}

# Run database migrations
run_migrations() {
    print_status "Running database migrations..."
    
    # Check if migrate tool is installed
    if ! command -v migrate &> /dev/null; then
        print_warning "migrate tool not found, installing..."
        go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
    fi
    
    # Run migrations
    migrate -path migrations -database "postgres://postgres:password@localhost:5432/smartedify_auth?sslmode=disable" up
    
    print_success "Database migrations completed"
}

# Generate test data (optional)
generate_test_data() {
    print_status "Would you like to generate test data? (y/n)"
    read -r response
    
    if [[ "$response" =~ ^[Yy]$ ]]; then
        print_status "Generating test data..."
        
        # This would typically call a script to insert test data
        # For now, we'll just create a placeholder
        echo "-- Test data generation would go here" > test_data.sql
        
        print_success "Test data generated (placeholder)"
    fi
}

# Display service URLs
show_urls() {
    print_success "🎉 Setup completed successfully!"
    echo ""
    echo "Service URLs:"
    echo "  📱 Auth Service:    http://localhost:8080"
    echo "  🗄️  PostgreSQL:      localhost:5432"
    echo "  🔴 Redis:           localhost:6379"
    echo "  📊 Prometheus:      http://localhost:9090"
    echo "  📈 Grafana:         http://localhost:3000 (admin/admin)"
    echo "  🔍 Jaeger:          http://localhost:16686"
    echo ""
    echo "Health Checks:"
    echo "  🏥 Service Health:  http://localhost:8080/health"
    echo "  🔑 JWKS:            http://localhost:8080/.well-known/jwks.json"
    echo "  🆔 OpenID Config:   http://localhost:8080/.well-known/openid-configuration"
    echo ""
    echo "Next steps:"
    echo "  1. Edit .env file with your configuration"
    echo "  2. Run 'make run' or 'go run ./cmd/auth-service' to start the service"
    echo "  3. Run 'make test' to run tests"
    echo "  4. Check the README.md for API documentation"
}

# Main execution
main() {
    echo "🏢 SmartEdify Auth Service - Development Setup"
    echo "=============================================="
    echo ""
    
    check_requirements
    setup_env
    setup_dependencies
    start_infrastructure
    run_migrations
    generate_test_data
    show_urls
}

# Run main function
main "$@"