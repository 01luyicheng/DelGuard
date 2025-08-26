#!/bin/bash
# DelGuard Enhanced Universal Installer for Unix Systems (Fixed Version)
# Compatible with bash, zsh, fish, and other POSIX shells
# Supports: Linux, macOS, FreeBSD, and other Unix-like systems
# Version: 2.1.1 (Fixed)

set -e  # Exit on any error
set -u  # Exit on undefined variables

# Configuration with better defaults
INSTALL_PATH="${INSTALL_PATH:-$HOME/bin}"
FORCE="${FORCE:-false}"
QUIET="${QUIET:-false}"
UNINSTALL="${UNINSTALL:-false}"
HELP="${HELP:-false}"
SYSTEM_WIDE="${SYSTEM_WIDE:-false}"
SKIP_ALIASES="${SKIP_ALIASES:-false}"
CHECK_ONLY="${CHECK_ONLY:-false}"
VERSION="${VERSION:-latest}"
DEBUG="${DEBUG:-false}"

# Global variables
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_FILE="/tmp/delguard-install-$(date +%Y%m%d-%H%M%S).log"
INSTALLATION_LOG=()

# Enhanced color output functions with fallback
print_color() {
    local color="$1"
    local message="$2"
    local timestamp="$(date '+%H:%M:%S')"
    
    # Add to log
    INSTALLATION_LOG+=("$timestamp [$color] $message")
    echo "$timestamp [$color] $message" >> "$LOG_FILE" 2>/dev/null || true
    
    if [[ "$QUIET" != "true" ]]; then
        if command -v tput >/dev/null 2>&1 && [[ -t 1 ]]; then
            # Terminal supports colors
            case "$color" in
                red)     echo -e "$(tput setaf 1)✗ $message$(tput sgr0)" ;;
                green)   echo -e "$(tput setaf 2)✓ $message$(tput sgr0)" ;;
                yellow)  echo -e "$(tput setaf 3)⚠ $message$(tput sgr0)" ;;
                blue)    echo -e "$(tput setaf 4)ℹ $message$(tput sgr0)" ;;
                cyan)    echo -e "$(tput setaf 6)ℹ $message$(tput sgr0)" ;;
                magenta) echo -e "$(tput setaf 5)$message$(tput sgr0)" ;;
                *)       echo "$message" ;;
            esac
        else
            # Fallback without colors
            case "$color" in
                red)     echo "✗ $message" ;;
                green)   echo "✓ $message" ;;
                yellow)  echo "⚠ $message" ;;
                blue|cyan) echo "ℹ $message" ;;
                *)       echo "$message" ;;
            esac
        fi
    fi
}

print_success() { print_color green "$1"; }
print_warning() { print_color yellow "$1"; }
print_error() { print_color red "$1"; }
print_info() { print_color cyan "$1"; }
print_header() { print_color magenta "$1"; }
print_debug() { [[ "$DEBUG" == "true" ]] && print_color blue "[DEBUG] $1"; }

# Error handling function
handle_error() {
    local exit_code=$?
    local line_number=$1
    print_error "Script failed at line $line_number with exit code $exit_code"
    print_info "Installation log saved to: $LOG_FILE"
    exit $exit_code
}

# Set up error trap
trap 'handle_error $LINENO' ERR

show_help() {
    cat << 'EOF'
DelGuard Enhanced Universal Installer for Unix Systems (Fixed)

USAGE:
    ./install.sh [OPTIONS]

OPTIONS:
    --install-path PATH    Install to specific directory (default: $HOME/bin)
    --force               Force overwrite existing installation
    --quiet               Suppress output messages
    --uninstall           Remove DelGuard installation
    --system-wide         Install system-wide (requires sudo)
    --skip-aliases        Skip shell alias configuration
    --check-only          Only check installation status
    --version VERSION     Install specific version (default: latest)
    --debug               Enable debug output
    --help                Show this help message

EXAMPLES:
    ./install.sh                                    # Install to default location
    ./install.sh --install-path /usr/local/bin     # Install system-wide
    ./install.sh --force                           # Force reinstall
    ./install.sh --system-wide                     # Install system-wide
    ./install.sh --skip-aliases                    # Install without aliases
    ./install.sh --check-only                      # Check installation status
    ./install.sh --uninstall                       # Remove DelGuard
    ./install.sh --debug                           # Enable debug output

AFTER INSTALLATION:
    Restart your shell or run: source ~/.bashrc (or your shell config)
    Then use these commands:
    - del <file>     # Safe delete
    - rm <file>      # Safe delete (replaces system rm)
    - cp <src> <dst> # Safe copy (replaces system cp)
    - delguard --help # Full help

SUPPORTED SYSTEMS:
    - Linux (all major distributions)
    - macOS (Intel and Apple Silicon)
    - FreeBSD
    - Other Unix-like systems

EOF
}

# Enhanced platform detection
detect_platform() {
    local uname_s="$(uname -s)"
    local uname_m="$(uname -m)"
    
    print_debug "Detecting platform: uname -s = $uname_s, uname -m = $uname_m"
    
    case "$uname_s" in
        Linux*)     
            PLATFORM="linux"
            # Detect Linux distribution
            if [[ -f /etc/os-release ]]; then
                DISTRO="$(grep '^ID=' /etc/os-release | cut -d'=' -f2 | tr -d '"')"
                print_debug "Linux distribution: $DISTRO"
            fi
            ;;
        Darwin*)    
            PLATFORM="darwin"
            MACOS_VERSION="$(sw_vers -productVersion 2>/dev/null || echo "unknown")"
            print_debug "macOS version: $MACOS_VERSION"
            ;;
        FreeBSD*)   PLATFORM="freebsd" ;;
        NetBSD*)    PLATFORM="netbsd" ;;
        OpenBSD*)   PLATFORM="openbsd" ;;
        CYGWIN*|MINGW*|MSYS*) 
            PLATFORM="windows"
            print_error "This script is for Unix systems. Use install.bat or install.ps1 for Windows."
            exit 1
            ;;
        *)          
            PLATFORM="unknown"
            print_warning "Unknown platform: $uname_s. Assuming Linux-like behavior."
            PLATFORM="linux"
            ;;
    esac
    
    case "$uname_m" in
        x86_64|amd64) ARCH="amd64" ;;
        i386|i686)    ARCH="386" ;;
        aarch64|arm64) ARCH="arm64" ;;
        armv7l)       ARCH="arm" ;;
        armv6l)       ARCH="arm" ;;
        *)            
            ARCH="amd64"
            print_warning "Unknown architecture: $uname_m. Assuming amd64."
            ;;
    esac
    
    print_debug "Platform: $PLATFORM, Architecture: $ARCH"
}

# Enhanced shell detection
detect_shell() {
    CURRENT_SHELL="$(basename "${SHELL:-/bin/sh}")"
    print_debug "Current shell: $CURRENT_SHELL"
    
    # Determine shell configuration files
    case "$CURRENT_SHELL" in
        bash)   
            SHELL_CONFIGS=("$HOME/.bashrc" "$HOME/.bash_profile" "$HOME/.profile")
            ;;
        zsh)    
            SHELL_CONFIGS=("$HOME/.zshrc" "$HOME/.zprofile")
            ;;
        fish)   
            SHELL_CONFIGS=("$HOME/.config/fish/config.fish")
            ;;
        ksh)    
            SHELL_CONFIGS=("$HOME/.kshrc" "$HOME/.profile")
            ;;
        tcsh|csh)    
            SHELL_CONFIGS=("$HOME/.tcshrc" "$HOME/.cshrc")
            ;;
        *)      
            SHELL_CONFIGS=("$HOME/.profile")
            print_warning "Unknown shell: $CURRENT_SHELL. Using .profile for configuration."
            ;;
    esac
    
    print_debug "Shell config files: ${SHELL_CONFIGS[*]}"
}

# System dependency checks
check_dependencies() {
    print_info "Checking system dependencies..."
    
    local missing_deps=()
    
    # Check for required commands
    local required_commands=("curl" "tar" "chmod" "mkdir" "cp" "rm")
    
    # Add platform-specific commands
    case "$PLATFORM" in
        darwin)
            required_commands+=("sw_vers")
            ;;
        linux)
            required_commands+=("which")
            ;;
    esac
    
    for cmd in "${required_commands[@]}"; do
        if ! command -v "$cmd" >/dev/null 2>&1; then
            missing_deps+=("$cmd")
        fi
    done
    
    # Check for Go if we need to build
    if [[ ! -f "delguard" && ! -f "$SCRIPT_DIR/delguard" ]]; then
        if ! command -v go >/dev/null 2>&1; then
            missing_deps+=("go")
        fi
    fi
    
    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        print_error "Missing required dependencies: ${missing_deps[*]}"
        print_info "Please install the missing dependencies and try again."
        
        # Provide installation hints
        case "$PLATFORM" in
            linux)
                if command -v apt-get >/dev/null 2>&1; then
                    print_info "Try: sudo apt-get install ${missing_deps[*]}"
                elif command -v yum >/dev/null 2>&1; then
                    print_info "Try: sudo yum install ${missing_deps[*]}"
                elif command -v pacman >/dev/null 2>&1; then
                    print_info "Try: sudo pacman -S ${missing_deps[*]}"
                fi
                ;;
            darwin)
                if command -v brew >/dev/null 2>&1; then
                    print_info "Try: brew install ${missing_deps[*]}"
                else
                    print_info "Consider installing Homebrew: https://brew.sh/"
                fi
                ;;
        esac
        
        return 1
    fi
    
    print_success "All dependencies are available"
    return 0
}

# Parse command line arguments with better error handling
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --install-path)
                if [[ -z "${2:-}" ]]; then
                    print_error "Option --install-path requires a value"
                    exit 1
                fi
                INSTALL_PATH="$2"
                shift 2
                ;;
            --force)
                FORCE="true"
                shift
                ;;
            --quiet)
                QUIET="true"
                shift
                ;;
            --uninstall)
                UNINSTALL="true"
                shift
                ;;
            --system-wide)
                SYSTEM_WIDE="true"
                shift
                ;;
            --skip-aliases)
                SKIP_ALIASES="true"
                shift
                ;;
            --check-only)
                CHECK_ONLY="true"
                shift
                ;;
            --version)
                if [[ -z "${2:-}" ]]; then
                    print_error "Option --version requires a value"
                    exit 1
                fi
                VERSION="$2"
                shift 2
                ;;
            --debug)
                DEBUG="true"
                shift
                ;;
            --help)
                HELP="true"
                shift
                ;;
            *)
                print_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# Continue with remaining functions...
# (Due to length limits, I'll create the rest in a separate response)