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
    local required_commands=("curl" "tar" "chmod" "mkdir" "cp" "rm")
    
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
    
    if [[ ! -f "delguard" && ! -f "$SCRIPT_DIR/delguard" ]]; then
        if ! command -v go >/dev/null 2>&1; then
            missing_deps+=("go")
        fi
    fi
    
    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        print_error "Missing required dependencies: ${missing_deps[*]}"
        print_info "Please install the missing dependencies and try again."
        
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

# Parse command line arguments
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
            --force) FORCE="true"; shift ;;
            --quiet) QUIET="true"; shift ;;
            --uninstall) UNINSTALL="true"; shift ;;
            --system-wide) SYSTEM_WIDE="true"; shift ;;
            --skip-aliases) SKIP_ALIASES="true"; shift ;;
            --check-only) CHECK_ONLY="true"; shift ;;
            --version)
                if [[ -z "${2:-}" ]]; then
                    print_error "Option --version requires a value"
                    exit 1
                fi
                VERSION="$2"
                shift 2
                ;;
            --debug) DEBUG="true"; shift ;;
            --help) HELP="true"; shift ;;
            *)
                print_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

check_installation_status() {
    print_info "Checking DelGuard installation status..."
    
    local common_paths=(
        "$HOME/bin/delguard"
        "/usr/local/bin/delguard"
        "/usr/bin/delguard"
        "/opt/delguard/bin/delguard"
    )
    
    local found_installations=()
    for path in "${common_paths[@]}"; do
        if [[ -f "$path" && -x "$path" ]]; then
            found_installations+=("$path")
        fi
    done
    
    if [[ ${#found_installations[@]} -eq 0 ]]; then
        print_warning "DelGuard is not installed"
        return 1
    fi
    
    print_success "Found DelGuard installations:"
    for installation in "${found_installations[@]}"; do
        if version_output=$("$installation" --version 2>/dev/null); then
            print_info "  $installation - $version_output"
        else
            print_info "  $installation - Version check failed"
        fi
    done
    
    if command -v delguard >/dev/null 2>&1; then
        local delguard_path
        delguard_path="$(command -v delguard)"
        print_success "DelGuard found in PATH: $delguard_path"
    else
        print_warning "DelGuard not found in PATH"
    fi
    
    local alias_configured=false
    for config in "${SHELL_CONFIGS[@]}"; do
        if [[ -f "$config" ]]; then
            if grep -q "DelGuard" "$config" 2>/dev/null; then
                print_success "DelGuard aliases configured in: $config"
                alias_configured=true
                break
            fi
        fi
    done
    
    if [[ "$alias_configured" != "true" ]]; then
        print_warning "DelGuard aliases not configured"
    fi
    
    return 0
}

build_delguard() {
    print_info "Building DelGuard executable..."
    
    if [[ ! -f "go.mod" ]]; then
        print_error "go.mod not found. Please run this script from DelGuard project root."
        return 1
    fi
    
    if [[ -f "delguard" ]]; then
        print_success "DelGuard executable already exists"
        return 0
    fi
    
    if ! command -v go >/dev/null 2>&1; then
        print_error "Go is not installed. Please install Go first."
        print_info "Visit: https://golang.org/dl/"
        return 1
    fi
    
    export CGO_ENABLED=0
    export GOOS="$PLATFORM"
    export GOARCH="$ARCH"
    
    print_info "Building for $PLATFORM/$ARCH..."
    
    if [[ -f "main.go" ]]; then
        print_info "Building with main.go..."
        if ! go build -ldflags "-s -w" -o "delguard" "main.go" 2>&1; then
            print_error "Build failed. Please check Go installation and project dependencies."
            return 1
        fi
    else
        print_info "Building entire project..."
        if ! go build -ldflags "-s -w" -o "delguard" 2>&1; then
            print_error "Build failed. Please check Go installation and project dependencies."
            return 1
        fi
    fi
    
    if [[ ! -f "delguard" ]]; then
        print_error "Build completed but executable not found."
        return 1
    fi
    
    chmod +x "delguard"
    print_success "DelGuard executable built successfully"
    return 0
}

install_delguard() {
    if [[ "$SYSTEM_WIDE" == "true" ]]; then
        if [[ "$INSTALL_PATH" == "$HOME/bin" ]]; then
            INSTALL_PATH="/usr/local/bin"
        fi
        
        if [[ ! -w "$(dirname "$INSTALL_PATH")" ]]; then
            print_error "System-wide installation requires sudo privileges."
            print_info "Please run with sudo or use user installation."
            return 1
        fi
    fi
    
    local target_exe="$INSTALL_PATH/delguard"
    
    print_info "Installing DelGuard to: $INSTALL_PATH"
    
    if [[ ! -d "$INSTALL_PATH" ]]; then
        if [[ "$SYSTEM_WIDE" == "true" ]]; then
            if ! sudo mkdir -p "$INSTALL_PATH" 2>/dev/null; then
                print_error "Failed to create install directory: $INSTALL_PATH"
                return 1
            fi
        else
            if ! mkdir -p "$INSTALL_PATH" 2>/dev/null; then
                print_error "Failed to create install directory: $INSTALL_PATH"
                return 1
            fi
        fi
        print_success "Created install directory: $INSTALL_PATH"
    fi
    
    if [[ -f "$target_exe" && "$FORCE" != "true" ]]; then
        print_warning "DelGuard is already installed at: $target_exe"
        print_warning "Use --force to overwrite existing installation"
        return 1
    fi
    
    if [[ "$SYSTEM_WIDE" == "true" ]]; then
        if ! sudo cp "delguard" "$target_exe" 2>/dev/null; then
            print_error "Failed to copy executable to: $target_exe"
            return 1
        fi
        sudo chmod +x "$target_exe"
    else
        if ! cp "delguard" "$target_exe" 2>/dev/null; then
            print_error "Failed to copy executable to: $target_exe"
            return 1
        fi
        chmod +x "$target_exe"
    fi
    
    print_success "Installed executable to: $target_exe"
    
    if ! "$target_exe" --version >/dev/null 2>&1; then
        print_warning "Installation verification failed - executable may not work properly"
    else
        print_success "Installation verified successfully"
    fi
    
    return 0
}

install_shell_aliases() {
    if [[ "$SKIP_ALIASES" == "true" ]]; then
        print_info "Skipping shell alias configuration"
        return 0
    fi
    
    print_info "Configuring shell aliases for $CURRENT_SHELL..."
    
    local config_block=""
    if [[ "$CURRENT_SHELL" == "fish" ]]; then
        config_block="# DelGuard Safe Delete Tool Configuration
# Generated: $(date '+%Y-%m-%d %H:%M:%S')
# Version: DelGuard 2.1.1 Enhanced (Fixed)

set -gx DELGUARD_PATH '$INSTALL_PATH/delguard'

if test -f \$DELGUARD_PATH
    if not contains '$INSTALL_PATH' \$PATH
        set -gx PATH '$INSTALL_PATH' \$PATH
    end
    
    function del
        if contains -- '--install' \$argv; or contains -- '--uninstall' \$argv
            echo 'Use \"delguard \$argv\" for installation commands'
            return
        end
        \$DELGUARD_PATH \$argv
    end
    
    function rm; \$DELGUARD_PATH \$argv; end
    function cp; \$DELGUARD_PATH --copy \$argv; end
    function delguard; \$DELGUARD_PATH \$argv; end
    
    if not set -q DELGUARD_LOADED
        echo 'DelGuard Safe Delete Tool Loaded (Enhanced Fixed)'
        echo 'Commands: del, rm, cp, delguard'
        echo 'Use \"delguard --help\" for detailed help'
        set -g DELGUARD_LOADED true
    end
else
    echo 'Warning: DelGuard executable not found: '\$DELGUARD_PATH
end
# End DelGuard Configuration"
    else
        config_block="# DelGuard Safe Delete Tool Configuration
# Generated: $(date '+%Y-%m-%d %H:%M:%S')
# Version: DelGuard 2.1.1 Enhanced (Fixed)

DELGUARD_PATH='$INSTALL_PATH/delguard'

if [ -f \"\$DELGUARD_PATH\" ]; then
    case \":\$PATH:\" in
        *:$INSTALL_PATH:*) ;;
        *) export PATH=\"$INSTALL_PATH:\$PATH\" ;;
    esac
    
    del() {
        for arg in \"\$@\"; do
            if [ \"\$arg\" = \"--install\" ] || [ \"\$arg\" = \"--uninstall\" ]; then
                echo 'Use \"delguard \$@\" for installation commands'
                return
            fi
        done
        \"\$DELGUARD_PATH\" \"\$@\"
    }
    
    rm() { \"\$DELGUARD_PATH\" \"\$@\"; }
    cp() { \"\$DELGUARD_PATH\" --copy \"\$@\"; }
    delguard() { \"\$DELGUARD_PATH\" \"\$@\"; }
    
    if [ -z \"\$DELGUARD_LOADED\" ]; then
        echo 'DelGuard Safe Delete Tool Loaded (Enhanced Fixed)'
        echo 'Commands: del, rm, cp, delguard'
        echo 'Use \"delguard --help\" for detailed help'
        export DELGUARD_LOADED=true
    fi
else
    echo 'Warning: DelGuard executable not found: '\$DELGUARD_PATH
fi
# End DelGuard Configuration"
    fi
    
    local success=false
    for config in "${SHELL_CONFIGS[@]}"; do
        local config_dir
        config_dir="$(dirname "$config")"
        if [[ ! -d "$config_dir" ]]; then
            if ! mkdir -p "$config_dir" 2>/dev/null; then
                print_warning "Could not create config directory: $config_dir"
                continue
            fi
            print_success "Created config directory: $config_dir"
        fi
        
        local existing_content=""
        if [[ -f "$config" ]]; then
            existing_content="$(cat "$config" 2>/dev/null || true)"
        fi
        
        if echo "$existing_content" | grep -q "# DelGuard"; then
            if [[ "$FORCE" != "true" ]]; then
                print_warning "DelGuard configuration already exists in: $config"
                print_warning "Use --force to overwrite"
                continue
            fi
            existing_content="$(echo "$existing_content" | sed '/# DelGuard Safe Delete Tool Configuration/,/# End DelGuard Configuration/d' 2>/dev/null || echo "$existing_content")"
        fi
        
        {
            echo "$existing_content"
            echo ""
            echo "$config_block"
        } > "$config" 2>/dev/null && {
            print_success "Updated shell configuration: $config"
            success=true
            break
        }
    done
    
    if [[ "$success" != "true" ]]; then
        print_warning "Failed to update shell configuration files"
        print_info "Please add $INSTALL_PATH to your PATH manually:"
        print_info "echo 'export PATH=\"$INSTALL_PATH:\$PATH\"' >> ~/.bashrc"
        return 1
    fi
    
    return 0
}

uninstall_delguard() {
    print_info "Uninstalling DelGuard..."
    
    local exe_paths=(
        "$HOME/bin/delguard"
        "/usr/local/bin/delguard"
        "/usr/bin/delguard"
        "/opt/delguard/bin/delguard"
    )
    
    if [[ -n "$INSTALL_PATH" ]]; then
        exe_paths+=("$INSTALL_PATH/delguard")
    fi
    
    for exe_path in "${exe_paths[@]}"; do
        if [[ -f "$exe_path" ]]; then
            local install_dir
            install_dir="$(dirname "$exe_path")"
            
            if [[ "$exe_path" == "/usr/local/bin/delguard" || "$exe_path" == "/usr/bin/delguard" ]]; then
                if sudo rm -f "$exe_path" 2>/dev/null; then
                    print_success "Removed executable: $exe_path"
                else
                    print_warning "Failed to remove: $exe_path"
                fi
            else
                if rm -f "$exe_path" 2>/dev/null; then
                    print_success "Removed executable: $exe_path"
                else
                    print_warning "Failed to remove: $exe_path"
                fi
            fi
            
            if [[ "$install_dir" != "/usr/bin" && "$install_dir" != "/usr/local/bin" ]]; then
                rmdir "$install_dir" 2>/dev/null && print_success "Removed empty directory: $install_dir" || true
            fi
        fi
    done
    
    local configs=(
        "$HOME/.bashrc" "$HOME/.bash_profile" "$HOME/.zshrc" "$HOME/.zprofile"
        "$HOME/.profile" "$HOME/.config/fish/config.fish" "$HOME/.kshrc"
        "$HOME/.tcshrc" "$HOME/.cshrc"
    )
    
    for config in "${configs[@]}"; do
        if [[ -f "$config" ]]; then
            if grep -q "# DelGuard" "$config" 2>/dev/null; then
                cp "$config" "$config.delguard-backup" 2>/dev/null || true
                sed -i.bak '/# DelGuard Safe Delete Tool Configuration/,/# End DelGuard Configuration/d' "$config" 2>/dev/null || {
                    sed '/# DelGuard Safe Delete Tool Configuration/,/# End DelGuard Configuration/d' "$config" > "$config.tmp" && mv "$config.tmp" "$config"
                }
                print_success "Removed DelGuard configuration from: $config"
            fi
        fi
    done
    
    print_success "DelGuard uninstalled successfully!"
    print_info "Please restart your shell or run: source ~/.bashrc (or your shell config)"
}

save_installation_log() {
    if [[ ${#INSTALLATION_LOG[@]} -gt 0 ]]; then
        {
            echo "DelGuard Installation Log"
            echo "========================="
            echo "Date: $(date)"
            echo "Platform: $PLATFORM/$ARCH"
            echo "Shell: $CURRENT_SHELL"
            echo "Install Path: $INSTALL_PATH"
            echo ""
            printf '%s\n' "${INSTALLATION_LOG[@]}"
        } > "$LOG_FILE" 2>/dev/null || true
        
        print_info "Installation log saved to: $LOG_FILE"
    fi
}

# Main execution function
main() {
    echo "DelGuard Installation Started: $(date)" > "$LOG_FILE" 2>/dev/null || true
    
    parse_args "$@"
    
    if [[ "$HELP" == "true" ]]; then
        show_help
        exit 0
    fi
    
    detect_platform
    detect_shell
    
    print_header "=== DelGuard Enhanced Universal Installer for Unix Systems (Fixed) ==="
    print_info "Platform: $PLATFORM ($ARCH)"
    print_info "Shell: $CURRENT_SHELL"
    print_info "User: $(whoami)"
    print_info "Install Path: $INSTALL_PATH"
    
    if ! check_dependencies; then
        save_installation_log
        exit 1
    fi
    
    if [[ "$CHECK_ONLY" == "true" ]]; then
        if check_installation_status; then
            save_installation_log
            exit 0
        else
            save_installation_log
            exit 1
        fi
    fi
    
    if [[ "$UNINSTALL" == "true" ]]; then
        uninstall_delguard
        save_installation_log
        exit 0
    fi
    
    print_info "Starting DelGuard installation..."
    
    if ! build_delguard; then
        print_error "Build failed. Installation aborted."
        save_installation_log
        exit 1
    fi
    
    if ! install_delguard; then
        print_error "Installation failed."
        save_installation_log
        exit 1
    fi
    
    if ! install_shell_aliases; then
        print_warning "Shell alias configuration failed, but DelGuard was installed successfully."
        print_info "You can still use 'delguard' command directly."
    fi
    
    print_success "=== Installation Complete ==="
    print_info "DelGuard has been installed successfully!"
    print_info ""
    print_info "NEXT STEPS:"
    print_info "1. Restart your shell or run: source ~/.bashrc (or your shell config)"
    print_info "2. Test with: delguard --version"
    print_info "3. Check status with: ./install.sh --check-only"
    print_info "4. Use these safe commands:"
    print_info "   • del <file>     - Safe delete"
    print_info "   • rm <file>      - Safe delete (replaces system rm)"
    print_info "   • cp <src> <dst> - Safe copy (replaces system cp)"
    print_info "   • delguard --help - Full help and options"
    print_info ""
    print_success "Happy safe deleting!"
    
    print_info "Testing installation..."
    if "./delguard" --version >/dev/null 2>&1; then
        print_success "✓ DelGuard is working correctly"
    else
        print_warning "⚠ DelGuard may not be working properly"
    fi
    
    save_installation_log
}

# Run main function with all arguments
main "$@"