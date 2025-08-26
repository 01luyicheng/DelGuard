#!/bin/bash
# DelGuard Universal Shell Installer
# Supports: bash, zsh, fish, and other POSIX shells
# Author: DelGuard Team
# Version: 1.0

set -e

# Configuration
INSTALL_PATH="${INSTALL_PATH:-$HOME/bin}"
FORCE="${FORCE:-false}"
QUIET="${QUIET:-false}"
UNINSTALL="${UNINSTALL:-false}"

# Color output functions
print_color() {
    local color="$1"
    local message="$2"
    if [[ "$QUIET" != "true" ]]; then
        case "$color" in
            red)     echo -e "\033[31m$message\033[0m" ;;
            green)   echo -e "\033[32m$message\033[0m" ;;
            yellow)  echo -e "\033[33m$message\033[0m" ;;
            blue)    echo -e "\033[34m$message\033[0m" ;;
            cyan)    echo -e "\033[36m$message\033[0m" ;;
            *)       echo "$message" ;;
        esac
    fi
}

print_success() { print_color green "$1"; }
print_warning() { print_color yellow "$1"; }
print_error() { print_color red "$1"; }
print_info() { print_color cyan "$1"; }

# Platform detection
detect_platform() {
    case "$(uname -s)" in
        Linux*)     PLATFORM="linux" ;;
        Darwin*)    PLATFORM="macos" ;;
        CYGWIN*|MINGW*|MSYS*) PLATFORM="windows" ;;
        *)          PLATFORM="unknown" ;;
    esac
}

# Shell detection
detect_shell() {
    CURRENT_SHELL=$(basename "$SHELL")
    case "$CURRENT_SHELL" in
        bash)   SHELL_CONFIG="$HOME/.bashrc" ;;
        zsh)    SHELL_CONFIG="$HOME/.zshrc" ;;
        fish)   SHELL_CONFIG="$HOME/.config/fish/config.fish" ;;
        *)      SHELL_CONFIG="$HOME/.profile" ;;
    esac
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --install-path)
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
            --help)
                show_help
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

show_help() {
    cat << EOF
DelGuard Universal Shell Installer

Usage: $0 [OPTIONS]

Options:
    --install-path PATH    Install to specific path (default: \$HOME/bin)
    --force               Force overwrite existing installation
    --quiet               Suppress output messages
    --uninstall           Remove DelGuard installation
    --help                Show this help message

Examples:
    $0                              # Install to default location
    $0 --install-path /usr/local/bin # Install to system-wide location
    $0 --uninstall                  # Remove DelGuard
EOF
}

# Uninstall function
uninstall_delguard() {
    print_info "Uninstalling DelGuard..."
    
    # Remove executable
    local exe_path="$INSTALL_PATH/delguard"
    if [[ -f "$exe_path" ]]; then
        rm -f "$exe_path"
        print_success "Removed executable: $exe_path"
    fi
    
    # Remove from shell configurations
    local configs=("$HOME/.bashrc" "$HOME/.zshrc" "$HOME/.profile" "$HOME/.config/fish/config.fish")
    for config in "${configs[@]}"; do
        if [[ -f "$config" ]]; then
            # Remove DelGuard configuration block
            if grep -q "# DelGuard Shell Configuration" "$config"; then
                sed -i '/# DelGuard Shell Configuration/,/# End DelGuard Configuration/d' "$config" 2>/dev/null || true
                print_success "Removed DelGuard configuration from: $config"
            fi
        fi
    done
    
    print_success "DelGuard uninstalled successfully!"
    print_info "Please restart your shell or run: source ~/.bashrc (or your shell config)"
}

# Main installation function
install_delguard() {
    detect_platform
    detect_shell
    
    print_info "=== DelGuard Universal Installer ==="
    print_info "Platform: $PLATFORM"
    print_info "Shell: $CURRENT_SHELL"
    print_info "Install path: $INSTALL_PATH"
    
    # Create install directory
    if [[ ! -d "$INSTALL_PATH" ]]; then
        mkdir -p "$INSTALL_PATH"
        print_success "Created directory: $INSTALL_PATH"
    fi
    
    # Find source executable
    local source_exe=""
    local script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    
    # Try different locations
    for location in "$script_dir/../delguard" "$script_dir/delguard" "./delguard"; do
        if [[ -f "$location" ]]; then
            source_exe="$location"
            break
        fi
    done
    
    if [[ -z "$source_exe" ]]; then
        print_error "DelGuard executable not found. Please build the project first."
        print_error "Run: go build -o delguard"
        exit 1
    fi
    
    local target_exe="$INSTALL_PATH/delguard"
    
    # Check if already installed
    if [[ -f "$target_exe" && "$FORCE" != "true" ]]; then
        print_warning "DelGuard is already installed at: $target_exe"
        print_warning "Use --force to overwrite"
        exit 1
    fi
    
    # Copy executable
    cp "$source_exe" "$target_exe"
    chmod +x "$target_exe"
    print_success "Copied executable to: $target_exe"
    
    # Add to PATH if not already there
    if [[ ":$PATH:" != *":$INSTALL_PATH:"* ]]; then
        print_info "Adding $INSTALL_PATH to PATH in shell configuration"
    fi
    
    # Create shell configuration
    local config_block=""
    if [[ "$CURRENT_SHELL" == "fish" ]]; then
        config_block="# DelGuard Shell Configuration
# Generated: $(date '+%Y-%m-%d %H:%M:%S')
# Version: DelGuard 1.0 for Fish Shell

set -gx DELGUARD_PATH '$target_exe'

if test -f \$DELGUARD_PATH
    # Add to PATH if not already there
    if not contains '$INSTALL_PATH' \$PATH
        set -gx PATH '$INSTALL_PATH' \$PATH
    end
    
    # Define alias functions
    function del
        \$DELGUARD_PATH -i \$argv
    end
    
    function rm
        \$DELGUARD_PATH -i \$argv
    end
    
    function cp
        \$DELGUARD_PATH --cp \$argv
    end
    
    function delguard
        \$DELGUARD_PATH \$argv
    end
    
    # Show loading message only once per session
    if not set -q DELGUARD_LOADED
        echo 'DelGuard aliases loaded successfully'
        echo 'Commands: del, rm, cp, delguard'
        echo 'Use --help for detailed help'
        set -g DELGUARD_LOADED true
    end
else
    echo 'Warning: DelGuard executable not found: '\$DELGUARD_PATH
end
# End DelGuard Configuration"
    else
        config_block="# DelGuard Shell Configuration
# Generated: $(date '+%Y-%m-%d %H:%M:%S')
# Version: DelGuard 1.0 for POSIX Shells

DELGUARD_PATH='$target_exe'

if [ -f \"\$DELGUARD_PATH\" ]; then
    # Add to PATH if not already there
    case \":\$PATH:\" in
        *:$INSTALL_PATH:*) ;;
        *) export PATH=\"$INSTALL_PATH:\$PATH\" ;;
    esac
    
    # Define alias functions
    del() {
        \"\$DELGUARD_PATH\" -i \"\$@\"
    }
    
    rm() {
        \"\$DELGUARD_PATH\" -i \"\$@\"
    }
    
    cp() {
        \"\$DELGUARD_PATH\" --cp \"\$@\"
    }
    
    delguard() {
        \"\$DELGUARD_PATH\" \"\$@\"
    }
    
    # Show loading message only once per session
    if [ -z \"\$DELGUARD_LOADED\" ]; then
        echo 'DelGuard aliases loaded successfully'
        echo 'Commands: del, rm, cp, delguard'
        echo 'Use --help for detailed help'
        export DELGUARD_LOADED=true
    fi
else
    echo 'Warning: DelGuard executable not found: '\$DELGUARD_PATH
fi
# End DelGuard Configuration"
    fi
    
    # Update shell configuration files
    local configs=()
    if [[ "$CURRENT_SHELL" == "fish" ]]; then
        configs=("$HOME/.config/fish/config.fish")
    else
        configs=("$HOME/.bashrc" "$HOME/.zshrc" "$HOME/.profile")
    fi
    
    for config in "${configs[@]}"; do
        # Skip if file doesn't exist and it's not the primary config
        if [[ ! -f "$config" && "$config" != "$SHELL_CONFIG" ]]; then
            continue
        fi
        
        # Create config directory if needed
        local config_dir="$(dirname "$config")"
        if [[ ! -d "$config_dir" ]]; then
            mkdir -p "$config_dir"
            print_success "Created config directory: $config_dir"
        fi
        
        # Check existing content
        local existing_content=""
        if [[ -f "$config" ]]; then
            existing_content="$(cat "$config")"
        fi
        
        if echo "$existing_content" | grep -q "# DelGuard Shell Configuration"; then
            if [[ "$FORCE" != "true" ]]; then
                print_warning "DelGuard configuration already exists in: $config"
                print_warning "Use --force to overwrite"
                continue
            fi
            # Remove existing DelGuard configuration
            sed -i '/# DelGuard Shell Configuration/,/# End DelGuard Configuration/d' "$config" 2>/dev/null || true
        fi
        
        # Append new configuration
        echo "" >> "$config"
        echo "$config_block" >> "$config"
        print_success "Updated shell configuration: $config"
    done
    
    print_success "=== Installation Complete ==="
    print_info "DelGuard has been installed successfully!"
    print_info "Available commands: del, rm, cp, delguard"
    print_info "Restart your shell or run: source $SHELL_CONFIG"
    
    # Test installation
    print_info "Testing installation..."
    if "$target_exe" --version >/dev/null 2>&1; then
        print_success "✓ DelGuard is working correctly"
    else
        print_warning "⚠ DelGuard may not be working properly"
    fi
}

# Main execution
main() {
    parse_args "$@"
    
    if [[ "$UNINSTALL" == "true" ]]; then
        uninstall_delguard
    else
        install_delguard
    fi
}

# Run main function with all arguments
main "$@"