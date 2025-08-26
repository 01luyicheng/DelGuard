#!/bin/bash
# DelGuard Installation Validator and Rollback System
# Version: 2.1.1 Enhanced
# Provides installation verification and rollback capabilities

# Source the logger if available
if [[ -f "$(dirname "${BASH_SOURCE[0]}")/install-logger.sh" ]]; then
    source "$(dirname "${BASH_SOURCE[0]}")/install-logger.sh"
else
    # Fallback logging functions
    log_info() { echo "ℹ $1"; }
    log_success() { echo "✓ $1"; }
    log_error() { echo "✗ $1" >&2; }
    log_warn() { echo "⚠ $1" >&2; }
fi

# Global variables for rollback
ROLLBACK_LOG="/tmp/delguard-rollback-$(date +%Y%m%d-%H%M%S).log"
BACKUP_DIR="/tmp/delguard-backup-$(date +%Y%m%d-%H%M%S)"
ROLLBACK_ACTIONS=()

# Initialize rollback system
init_rollback_system() {
    mkdir -p "$BACKUP_DIR"
    echo "DelGuard Installation Rollback Log" > "$ROLLBACK_LOG"
    echo "Backup Directory: $BACKUP_DIR" >> "$ROLLBACK_LOG"
    echo "Started: $(date)" >> "$ROLLBACK_LOG"
    echo "" >> "$ROLLBACK_LOG"
    
    log_info "Rollback system initialized: $BACKUP_DIR"
}

# Record rollback action
record_rollback_action() {
    local action_type="$1"
    local target="$2"
    local backup_path="$3"
    
    local action="$action_type:$target:$backup_path"
    ROLLBACK_ACTIONS+=("$action")
    
    echo "$(date '+%H:%M:%S') RECORD: $action" >> "$ROLLBACK_LOG"
    log_debug "Recorded rollback action: $action_type for $target"
}

# Backup file before modification
backup_file() {
    local file_path="$1"
    local backup_name="${2:-$(basename "$file_path")}"
    
    if [[ ! -f "$file_path" ]]; then
        log_debug "File does not exist, no backup needed: $file_path"
        return 0
    fi
    
    local backup_path="$BACKUP_DIR/$backup_name"
    
    if cp "$file_path" "$backup_path" 2>/dev/null; then
        record_rollback_action "RESTORE_FILE" "$file_path" "$backup_path"
        log_debug "Backed up file: $file_path -> $backup_path"
        return 0
    else
        log_error "Failed to backup file: $file_path"
        return 1
    fi
}

# Record file creation for rollback
record_file_creation() {
    local file_path="$1"
    record_rollback_action "DELETE_FILE" "$file_path" ""
    log_debug "Recorded file creation for rollback: $file_path"
}

# Record directory creation for rollback
record_dir_creation() {
    local dir_path="$1"
    record_rollback_action "DELETE_DIR" "$dir_path" ""
    log_debug "Recorded directory creation for rollback: $dir_path"
}

# Verify DelGuard installation
verify_installation() {
    local install_path="$1"
    local expected_version="$2"
    
    log_info "Verifying DelGuard installation..."
    
    local verification_errors=()
    
    # Check if executable exists
    local delguard_exe="$install_path/delguard"
    if [[ ! -f "$delguard_exe" ]]; then
        verification_errors+=("Executable not found: $delguard_exe")
    elif [[ ! -x "$delguard_exe" ]]; then
        verification_errors+=("Executable not executable: $delguard_exe")
    fi
    
    # Check if executable works
    if [[ -x "$delguard_exe" ]]; then
        if ! "$delguard_exe" --version >/dev/null 2>&1; then
            verification_errors+=("Executable fails to run: $delguard_exe --version")
        else
            local actual_version
            actual_version="$("$delguard_exe" --version 2>/dev/null | head -n1)"
            log_success "DelGuard version: $actual_version"
            
            # Check version if specified
            if [[ -n "$expected_version" && "$expected_version" != "latest" ]]; then
                if [[ "$actual_version" != *"$expected_version"* ]]; then
                    verification_errors+=("Version mismatch: expected $expected_version, got $actual_version")
                fi
            fi
        fi
    fi
    
    # Check PATH configuration
    if ! command -v delguard >/dev/null 2>&1; then
        verification_errors+=("DelGuard not found in PATH")
    else
        local path_delguard
        path_delguard="$(command -v delguard)"
        if [[ "$path_delguard" != "$delguard_exe" ]]; then
            log_warn "DelGuard in PATH ($path_delguard) differs from installed location ($delguard_exe)"
        fi
    fi
    
    # Check shell configuration
    local shell_configs=(
        "$HOME/.bashrc" "$HOME/.bash_profile" "$HOME/.zshrc" 
        "$HOME/.zprofile" "$HOME/.profile" "$HOME/.config/fish/config.fish"
    )
    
    local shell_configured=false
    for config in "${shell_configs[@]}"; do
        if [[ -f "$config" ]] && grep -q "DelGuard" "$config" 2>/dev/null; then
            log_success "Shell configuration found in: $config"
            shell_configured=true
            break
        fi
    done
    
    if [[ "$shell_configured" != "true" ]]; then
        verification_errors+=("Shell configuration not found in common config files")
    fi
    
    # Test basic functionality
    if [[ -x "$delguard_exe" ]]; then
        local test_file="/tmp/delguard-test-$(date +%s).txt"
        echo "test" > "$test_file"
        
        if "$delguard_exe" --help >/dev/null 2>&1; then
            log_success "Help command works"
        else
            verification_errors+=("Help command fails")
        fi
        
        # Clean up test file
        rm -f "$test_file" 2>/dev/null || true
    fi
    
    # Report verification results
    if [[ ${#verification_errors[@]} -eq 0 ]]; then
        log_success "Installation verification passed"
        return 0
    else
        log_error "Installation verification failed:"
        for error in "${verification_errors[@]}"; do
            log_error "  • $error"
        done
        return 1
    fi
}

# Perform rollback
perform_rollback() {
    log_info "Starting installation rollback..."
    
    local rollback_errors=()
    local actions_count=${#ROLLBACK_ACTIONS[@]}
    
    if [[ $actions_count -eq 0 ]]; then
        log_info "No rollback actions recorded"
        return 0
    fi
    
    log_info "Performing $actions_count rollback actions..."
    
    # Process rollback actions in reverse order
    for ((i = actions_count - 1; i >= 0; i--)); do
        local action="${ROLLBACK_ACTIONS[$i]}"
        local action_type="${action%%:*}"
        local remaining="${action#*:}"
        local target="${remaining%%:*}"
        local backup_path="${remaining#*:}"
        
        echo "$(date '+%H:%M:%S') ROLLBACK: $action" >> "$ROLLBACK_LOG"
        
        case "$action_type" in
            "DELETE_FILE")
                if [[ -f "$target" ]]; then
                    if rm -f "$target" 2>/dev/null; then
                        log_success "Removed file: $target"
                    else
                        rollback_errors+=("Failed to remove file: $target")
                    fi
                fi
                ;;
            "DELETE_DIR")
                if [[ -d "$target" ]]; then
                    if rmdir "$target" 2>/dev/null; then
                        log_success "Removed directory: $target"
                    else
                        log_warn "Directory not empty or failed to remove: $target"
                    fi
                fi
                ;;
            "RESTORE_FILE")
                if [[ -f "$backup_path" ]]; then
                    if cp "$backup_path" "$target" 2>/dev/null; then
                        log_success "Restored file: $target"
                    else
                        rollback_errors+=("Failed to restore file: $target")
                    fi
                else
                    rollback_errors+=("Backup file not found: $backup_path")
                fi
                ;;
            *)
                log_warn "Unknown rollback action type: $action_type"
                ;;
        esac
    done
    
    # Report rollback results
    if [[ ${#rollback_errors[@]} -eq 0 ]]; then
        log_success "Rollback completed successfully"
        
        # Clean up backup directory if empty
        if rmdir "$BACKUP_DIR" 2>/dev/null; then
            log_info "Cleaned up backup directory"
        fi
        
        return 0
    else
        log_error "Rollback completed with errors:"
        for error in "${rollback_errors[@]}"; do
            log_error "  • $error"
        done
        log_info "Backup files preserved in: $BACKUP_DIR"
        return 1
    fi
}

# Check system health before installation
check_system_health() {
    log_info "Checking system health..."
    
    local health_issues=()
    
    # Check disk space
    local available_space
    if command -v df >/dev/null 2>&1; then
        available_space=$(df -h . | awk 'NR==2 {print $4}' | sed 's/[^0-9.]//g')
        if [[ -n "$available_space" ]] && (( $(echo "$available_space < 100" | bc -l 2>/dev/null || echo 0) )); then
            health_issues+=("Low disk space: ${available_space}MB available")
        fi
    fi
    
    # Check memory
    if [[ -f /proc/meminfo ]]; then
        local available_mem
        available_mem=$(grep MemAvailable /proc/meminfo | awk '{print $2}' 2>/dev/null)
        if [[ -n "$available_mem" ]] && (( available_mem < 102400 )); then  # Less than 100MB
            health_issues+=("Low memory: $((available_mem / 1024))MB available")
        fi
    fi
    
    # Check for conflicting installations
    local existing_delguard
    if existing_delguard=$(command -v delguard 2>/dev/null); then
        log_warn "Existing DelGuard installation found: $existing_delguard"
    fi
    
    # Check write permissions
    local test_file="/tmp/delguard-write-test-$$"
    if ! touch "$test_file" 2>/dev/null; then
        health_issues+=("Cannot write to temporary directory")
    else
        rm -f "$test_file" 2>/dev/null || true
    fi
    
    # Report health check results
    if [[ ${#health_issues[@]} -eq 0 ]]; then
        log_success "System health check passed"
        return 0
    else
        log_warn "System health issues detected:"
        for issue in "${health_issues[@]}"; do
            log_warn "  • $issue"
        done
        return 1
    fi
}

# Generate installation report
generate_install_report() {
    local install_path="$1"
    local status="$2"
    local report_file="/tmp/delguard-install-report-$(date +%Y%m%d-%H%M%S).txt"
    
    cat > "$report_file" << EOF
DelGuard Installation Report
===========================
Date: $(date)
Status: $status
Install Path: $install_path
Platform: $(uname -s) $(uname -m)
User: $(whoami)
Shell: $SHELL

Installation Details:
EOF
    
    if [[ -f "$install_path/delguard" ]]; then
        echo "✓ Executable installed: $install_path/delguard" >> "$report_file"
        if [[ -x "$install_path/delguard" ]]; then
            echo "✓ Executable is executable" >> "$report_file"
            local version_output
            if version_output=$("$install_path/delguard" --version 2>/dev/null); then
                echo "✓ Version: $version_output" >> "$report_file"
            else
                echo "✗ Version check failed" >> "$report_file"
            fi
        else
            echo "✗ Executable is not executable" >> "$report_file"
        fi
    else
        echo "✗ Executable not found" >> "$report_file"
    fi
    
    if command -v delguard >/dev/null 2>&1; then
        echo "✓ DelGuard found in PATH: $(command -v delguard)" >> "$report_file"
    else
        echo "✗ DelGuard not found in PATH" >> "$report_file"
    fi
    
    echo "" >> "$report_file"
    echo "Log Files:" >> "$report_file"
    echo "• Installation Log: $DELGUARD_LOG_FILE" >> "$report_file"
    echo "• Rollback Log: $ROLLBACK_LOG" >> "$report_file"
    echo "• This Report: $report_file" >> "$report_file"
    
    log_info "Installation report generated: $report_file"
    echo "$report_file"
}

# Export functions
export -f init_rollback_system record_rollback_action backup_file
export -f record_file_creation record_dir_creation verify_installation
export -f perform_rollback check_system_health generate_install_report