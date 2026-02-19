#!/data/data/com.termux/files/usr/bin/bash
# KakoClaw Installation Script for Termux/Android
# This script automates the installation of KakoClaw on Android devices

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Variables
INSTALL_DIR="$HOME/.local/bin"
CONFIG_DIR="$HOME/.KakoClaw"
REPO_URL="https://github.com/sipeed/KakoClaw.git"
GO_VERSION_MIN="1.21"

echo -e "${BLUE}🐸 KakoClaw Installer for Termux/Android${NC}"
echo "=========================================="
echo ""

# Check if running in Termux
if [ -z "$TERMUX_VERSION" ] && [ ! -d "/data/data/com.termux" ]; then
    echo -e "${YELLOW}⚠️  Warning: This script is designed for Termux/Android${NC}"
    echo "It may not work correctly in other environments."
    read -p "Continue anyway? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to print status
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[OK]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Step 1: Check prerequisites
print_status "Checking prerequisites..."

if ! command_exists pkg; then
    print_error "This doesn't appear to be Termux. pkg command not found."
    exit 1
fi

print_success "Termux detected"

# Step 2: Update packages
print_status "Updating package lists..."
pkg update -y || {
    print_error "Failed to update packages"
    exit 1
}
print_success "Packages updated"

# Step 3: Install dependencies
print_status "Installing dependencies..."

DEPS="git golang make"
for dep in $DEPS; do
    if ! command_exists $dep; then
        print_status "Installing $dep..."
        pkg install -y $dep || {
            print_error "Failed to install $dep"
            exit 1
        }
    else
        print_success "$dep already installed"
    fi
done

# Check Go version
GO_VERSION=$(go version | grep -oP '\d+\.\d+' | head -1)
if [ -z "$GO_VERSION" ]; then
    print_error "Could not determine Go version"
    exit 1
fi

print_success "Go version: $GO_VERSION"

# Step 4: Clone repository
print_status "Cloning KakoClaw repository..."

if [ -d "$HOME/KakoClaw" ]; then
    print_warning "Directory $HOME/KakoClaw already exists"
    read -p "Remove and reclone? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -rf "$HOME/KakoClaw"
    else
        print_status "Using existing directory"
        cd "$HOME/KakoClaw"
        git pull || print_warning "Could not update repository"
    fi
fi

if [ ! -d "$HOME/KakoClaw" ]; then
    git clone "$REPO_URL" "$HOME/KakoClaw" || {
        print_error "Failed to clone repository"
        exit 1
    }
fi

cd "$HOME/KakoClaw"
print_success "Repository ready"

# Step 5: Build
print_status "Building KakoClaw..."

export CGO_ENABLED=0
export GOOS=android
export GOARCH=arm64

make build || {
    print_error "Build failed"
    print_status "Trying alternative build..."
    go build -o KakoClaw ./cmd/KakoClaw || {
        print_error "Alternative build also failed"
        exit 1
    }
}

print_success "Build completed"

# Step 6: Install
print_status "Installing KakoClaw..."

mkdir -p "$INSTALL_DIR"
cp KakoClaw "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/KakoClaw"

# Add to PATH if not already there
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    print_status "Adding $INSTALL_DIR to PATH..."
    echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$HOME/.bashrc"
    export PATH="$INSTALL_DIR:$PATH"
fi

print_success "KakoClaw installed to $INSTALL_DIR"

# Step 7: Initialize
print_status "Initializing KakoClaw..."

if [ ! -f "$CONFIG_DIR/config.json" ]; then
    KakoClaw onboard || {
        print_warning "onboard command failed, creating minimal config..."
        mkdir -p "$CONFIG_DIR/workspace"
        cat > "$CONFIG_DIR/config.json" << 'EOFCONFIG'
{
  "agents": {
    "defaults": {
      "workspace": "~/.KakoClaw/workspace",
      "model": "ollama/llama3.2",
      "max_tokens": 2048,
      "temperature": 0.7,
      "max_tool_iterations": 10
    }
  },
  "providers": {
    "ollama": {
      "api_base": "http://localhost:11434"
    }
  },
  "gateway": {
    "host": "0.0.0.0",
    "port": 18790
  }
}
EOFCONFIG
    }
else
    print_success "Configuration already exists"
fi

# Step 8: Setup storage permissions
print_status "Setting up storage permissions..."

if [ ! -d "$HOME/storage" ]; then
    print_status "Requesting storage permission..."
    termux-setup-storage || print_warning "Could not setup storage (may need manual permission)"
else
    print_success "Storage already configured"
fi

# Step 9: Create useful aliases
print_status "Creating aliases..."

if ! grep -q "alias pc=" "$HOME/.bashrc" 2>/dev/null; then
    cat >> "$HOME/.bashrc" << 'EOFALIASES'

# KakoClaw aliases
alias pc='KakoClaw agent'
alias pc-gateway='KakoClaw gateway'
alias pc-status='KakoClaw status'
alias pc-doctor='KakoClaw doctor'
EOFALIASES
    print_success "Aliases added to .bashrc"
fi

# Step 10: Create startup script
print_status "Creating startup script..."

mkdir -p "$HOME/.config/KakoClaw"
cat > "$HOME/.config/KakoClaw/start-gateway.sh" << 'EOFSTART'
#!/data/data/com.termux/files/usr/bin/bash
# Start KakoClaw Gateway

LOG_FILE="$HOME/KakoClaw-gateway.log"
PID_FILE="$HOME/.config/KakoClaw/gateway.pid"

start() {
    if [ -f "$PID_FILE" ] && kill -0 $(cat "$PID_FILE") 2>/dev/null; then
        echo "Gateway already running (PID: $(cat $PID_FILE))"
        return 1
    fi
    
    echo "Starting KakoClaw Gateway..."
    nohup KakoClaw gateway > "$LOG_FILE" 2>&1 &
    echo $! > "$PID_FILE"
    echo "Gateway started with PID: $!"
    echo "Log: $LOG_FILE"
}

stop() {
    if [ -f "$PID_FILE" ]; then
        PID=$(cat "$PID_FILE")
        if kill -0 "$PID" 2>/dev/null; then
            echo "Stopping Gateway (PID: $PID)..."
            kill "$PID"
            rm "$PID_FILE"
            echo "Stopped"
        else
            echo "Gateway not running"
            rm -f "$PID_FILE"
        fi
    else
        echo "Gateway not running"
    fi
}

status() {
    if [ -f "$PID_FILE" ] && kill -0 $(cat "$PID_FILE") 2>/dev/null; then
        echo "Gateway is running (PID: $(cat $PID_FILE))"
        echo "Recent logs:"
        tail -n 10 "$LOG_FILE" 2>/dev/null || echo "No logs yet"
    else
        echo "Gateway is not running"
    fi
}

case "$1" in
    start)
        start
        ;;
    stop)
        stop
        ;;
    restart)
        stop
        sleep 2
        start
        ;;
    status)
        status
        ;;
    *)
        echo "Usage: $0 {start|stop|restart|status}"
        exit 1
        ;;
esac
EOFSTART

chmod +x "$HOME/.config/KakoClaw/start-gateway.sh"
print_success "Startup script created"

# Summary
echo ""
echo "=========================================="
echo -e "${GREEN}✅ KakoClaw Installation Complete!${NC}"
echo "=========================================="
echo ""
echo "Installation directory: $INSTALL_DIR"
echo "Configuration: $CONFIG_DIR/config.json"
echo "Workspace: $CONFIG_DIR/workspace"
echo ""
echo "Quick Start:"
echo "  KakoClaw version      # Check version"
echo "  KakoClaw status       # Check status"
echo "  KakoClaw doctor       # Run health check"
echo "  KakoClaw agent        # Start interactive mode"
echo ""
echo "Aliases (after restarting Termux or running 'source ~/.bashrc'):"
echo "  pc                    # Shortcut for KakoClaw agent"
echo "  pc-gateway            # Shortcut for KakoClaw gateway"
echo "  pc-doctor             # Shortcut for KakoClaw doctor"
echo ""
echo "Gateway Management:"
echo "  ~/.config/KakoClaw/start-gateway.sh start   # Start gateway"
echo "  ~/.config/KakoClaw/start-gateway.sh stop    # Stop gateway"
echo "  ~/.config/KakoClaw/start-gateway.sh status  # Check status"
echo ""
echo "Next Steps:"
echo "  1. Edit config: nano ~/.KakoClaw/config.json"
echo "  2. Add your API keys or setup Ollama"
echo "  3. Run: KakoClaw agent -m 'Hello!'"
echo ""
echo -e "${YELLOW}Documentation:${NC} https://github.com/sipeed/KakoClaw/tree/main/docs"
echo ""
echo -e "${GREEN}Happy hacking! 🐸${NC}"
echo ""

# Reload shell configuration
print_status "Reloading shell configuration..."
source "$HOME/.bashrc" 2>/dev/null || true

exit 0
