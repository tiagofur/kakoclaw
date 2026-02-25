#!/bin/bash
set -e

# =============================================================================
# MakoClaw Deployment Script for Ubuntu VPS
# =============================================================================
#
# This script automates the deployment of MakoClaw on a VPS running Ubuntu.
# It handles:
#   - Prerequisites installation (Docker, Docker Compose)
#   - Environment configuration
#   - Building and starting the container
#   - Setting up reverse proxy (optional)
#   - SSL/TLS with Let's Encrypt (optional)
#
# Usage:
#   ./deploy.sh [options]
#
# Options:
#   --install-prereqs    Install Docker and Docker Compose
#   --setup-nginx        Setup Nginx reverse proxy
#   --setup-ssl          Setup SSL with Let's Encrypt (requires --setup-nginx)
#   --domain DOMAIN      Domain name for SSL (required with --setup-ssl)
#   --email EMAIL        Email for Let's Encrypt (required with --setup-ssl)
#   --build              Build Docker image from source
#   --pull               Pull pre-built image from registry
#   --start              Start the services
#   --stop               Stop the services
#   --restart            Restart the services
#   --logs               Show logs
#   --help               Show this help message
#
# Example:
#   # First time deployment with all features
#   ./deploy.sh --install-prereqs --setup-nginx --setup-ssl \
#               --domain makoclaw.example.com --email admin@example.com \
#               --build --start
#
#   # Quick restart
#   ./deploy.sh --restart
#
# =============================================================================

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
INSTALL_PREREQS=false
SETUP_NGINX=false
SETUP_SSL=false
BUILD_IMAGE=false
PULL_IMAGE=false
START_SERVICES=false
STOP_SERVICES=false
RESTART_SERVICES=false
SHOW_LOGS=false
DOMAIN=""
EMAIL=""
COMPOSE_FILE="docker-compose.prod.yml"

# Functions
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

show_help() {
    head -n 50 "$0" | grep "^#" | sed 's/^# //g' | sed 's/^#//g'
    exit 0
}

check_root() {
    if [ "$EUID" -eq 0 ]; then
        print_warning "Running as root. Consider using a non-root user with sudo."
    fi
}

install_prerequisites() {
    print_info "Installing prerequisites..."

    # Update package list
    sudo apt-get update

    # Install required packages
    sudo apt-get install -y \
        apt-transport-https \
        ca-certificates \
        curl \
        gnupg \
        lsb-release \
        git

    # Install Docker if not present
    if ! command -v docker &> /dev/null; then
        print_info "Installing Docker..."
        curl -fsSL https://get.docker.com -o get-docker.sh
        sudo sh get-docker.sh
        rm get-docker.sh

        # Add current user to docker group
        sudo usermod -aG docker $USER
        print_warning "Added $USER to docker group. You may need to log out and back in."
    else
        print_info "Docker is already installed."
    fi

    # Install Docker Compose if not present
    if ! command -v docker-compose &> /dev/null; then
        print_info "Installing Docker Compose..."
        sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" \
            -o /usr/local/bin/docker-compose
        sudo chmod +x /usr/local/bin/docker-compose
    else
        print_info "Docker Compose is already installed."
    fi

    # Verify installations
    docker --version
    docker-compose --version

    print_info "Prerequisites installed successfully!"
}

setup_environment() {
    print_info "Setting up environment..."

    if [ ! -f .env ]; then
        if [ -f .env.example ]; then
            cp .env.example .env
            print_warning "Created .env from .env.example. Please edit .env with your configuration."
            print_warning "Required settings:"
            print_warning "  - MAKOCLAW_WEB_PASSWORD (set a strong password)"
            print_warning "  - MAKOCLAW_AGENTS_DEFAULTS_PROVIDER (e.g., openrouter)"
            print_warning "  - MAKOCLAW_AGENTS_DEFAULTS_MODEL (e.g., anthropic/claude-3.5-sonnet)"
            print_warning "  - Provider API key (e.g., MAKOCLAW_PROVIDERS_OPENROUTER_API_KEY)"
            read -p "Press enter after editing .env to continue..."
        else
            print_error ".env.example not found. Cannot create .env file."
            exit 1
        fi
    else
        print_info ".env file already exists."
    fi

    # Create data directory
    mkdir -p MakoClaw-data
    print_info "Data directory created at MakoClaw-data/"
}

setup_nginx() {
    print_info "Setting up Nginx reverse proxy..."

    if [ -z "$DOMAIN" ]; then
        read -p "Enter your domain name: " DOMAIN
    fi

    # Install Nginx
    if ! command -v nginx &> /dev/null; then
        sudo apt-get install -y nginx
    fi

    # Create Nginx configuration
    sudo tee /etc/nginx/sites-available/makoclaw > /dev/null <<EOF
server {
    listen 80;
    listen [::]:80;
    server_name ${DOMAIN};

    client_max_body_size 100M;

    location / {
        proxy_pass http://127.0.0.1:18880;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_cache_bypass \$http_upgrade;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;

        # WebSocket support
        proxy_read_timeout 86400;
    }
}
EOF

    # Enable site
    sudo ln -sf /etc/nginx/sites-available/makoclaw /etc/nginx/sites-enabled/

    # Test and reload Nginx
    sudo nginx -t
    sudo systemctl reload nginx

    print_info "Nginx reverse proxy configured for ${DOMAIN}"
}

setup_ssl() {
    print_info "Setting up SSL with Let's Encrypt..."

    if [ -z "$DOMAIN" ]; then
        print_error "Domain name is required for SSL setup. Use --domain option."
        exit 1
    fi

    if [ -z "$EMAIL" ]; then
        read -p "Enter your email address for Let's Encrypt: " EMAIL
    fi

    # Install Certbot
    sudo apt-get install -y certbot python3-certbot-nginx

    # Obtain certificate
    sudo certbot --nginx -d "$DOMAIN" --non-interactive --agree-tos -m "$EMAIL"

    # Setup auto-renewal
    sudo systemctl enable certbot.timer

    print_info "SSL certificate obtained and auto-renewal configured!"
}

build_image() {
    print_info "Building Docker image..."
    docker-compose -f "$COMPOSE_FILE" build --no-cache
    print_info "Docker image built successfully!"
}

pull_image() {
    print_info "Pulling Docker image..."
    docker-compose -f "$COMPOSE_FILE" pull
    print_info "Docker image pulled successfully!"
}

start_services() {
    print_info "Starting services..."
    docker-compose -f "$COMPOSE_FILE" up -d
    print_info "Services started!"
    print_info "Access MakoClaw at: http://localhost:18880"
    if [ -n "$DOMAIN" ]; then
        print_info "Or via domain: https://${DOMAIN}"
    fi
}

stop_services() {
    print_info "Stopping services..."
    docker-compose -f "$COMPOSE_FILE" down
    print_info "Services stopped!"
}

restart_services() {
    print_info "Restarting services..."
    docker-compose -f "$COMPOSE_FILE" restart
    print_info "Services restarted!"
}

show_logs() {
    docker-compose -f "$COMPOSE_FILE" logs -f
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --install-prereqs)
            INSTALL_PREREQS=true
            shift
            ;;
        --setup-nginx)
            SETUP_NGINX=true
            shift
            ;;
        --setup-ssl)
            SETUP_SSL=true
            shift
            ;;
        --domain)
            DOMAIN="$2"
            shift 2
            ;;
        --email)
            EMAIL="$2"
            shift 2
            ;;
        --build)
            BUILD_IMAGE=true
            shift
            ;;
        --pull)
            PULL_IMAGE=true
            shift
            ;;
        --start)
            START_SERVICES=true
            shift
            ;;
        --stop)
            STOP_SERVICES=true
            shift
            ;;
        --restart)
            RESTART_SERVICES=true
            shift
            ;;
        --logs)
            SHOW_LOGS=true
            shift
            ;;
        --help|-h)
            show_help
            ;;
        *)
            print_error "Unknown option: $1"
            show_help
            ;;
    esac
done

# Main execution
print_info "=== MakoClaw Deployment Script ==="
check_root

# Execute based on flags
if [ "$INSTALL_PREREQS" = true ]; then
    install_prerequisites
fi

if [ "$BUILD_IMAGE" = true ] || [ "$PULL_IMAGE" = true ] || [ "$START_SERVICES" = true ]; then
    setup_environment
fi

if [ "$SETUP_NGINX" = true ]; then
    setup_nginx
fi

if [ "$SETUP_SSL" = true ]; then
    if [ "$SETUP_NGINX" != true ]; then
        print_error "SSL setup requires Nginx. Use --setup-nginx option."
        exit 1
    fi
    setup_ssl
fi

if [ "$BUILD_IMAGE" = true ]; then
    build_image
fi

if [ "$PULL_IMAGE" = true ]; then
    pull_image
fi

if [ "$STOP_SERVICES" = true ]; then
    stop_services
fi

if [ "$RESTART_SERVICES" = true ]; then
    restart_services
fi

if [ "$START_SERVICES" = true ]; then
    start_services
fi

if [ "$SHOW_LOGS" = true ]; then
    show_logs
fi

# If no action specified, show help
if [ "$INSTALL_PREREQS" = false ] && \
   [ "$BUILD_IMAGE" = false ] && \
   [ "$PULL_IMAGE" = false ] && \
   [ "$START_SERVICES" = false ] && \
   [ "$STOP_SERVICES" = false ] && \
   [ "$RESTART_SERVICES" = false ] && \
   [ "$SHOW_LOGS" = false ] && \
   [ "$SETUP_NGINX" = false ] && \
   [ "$SETUP_SSL" = false ]; then
    show_help
fi

print_info "=== Done! ==="
