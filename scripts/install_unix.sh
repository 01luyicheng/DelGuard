#!/bin/bash
# DelGuard Unix/Linux/macOS 安装脚本
# 版本: 2.0.0
# 描述: 自动安装DelGuard文件删除保护工具

set -e

# 默认配置
INSTALL_PATH="/opt/delguard"
SERVICE_NAME="delguard"
SILENT=false
CREATE_DESKTOP_SHORTCUT=true
ADD_TO_PATH=true
START_SERVICE=true
USER_SERVICE=false

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        --install-path)
            INSTALL_PATH="$2"
            shift 2
            ;;
        --service-name)
            SERVICE_NAME="$2"
            shift 2
            ;;
        --silent)
            SILENT=true
            shift
            ;;
        --no-desktop-shortcut)
            CREATE_DESKTOP_SHORTCUT=false
            shift
            ;;
        --no-path)
            ADD_TO_PATH=false
            shift
            ;;
        --no-start)
            START_SERVICE=false
            shift
            ;;
        --user-service)
            USER_SERVICE=true
            INSTALL_PATH="$HOME/.local/share/delguard"
            shift
            ;;
        --help)
            echo "DelGuard Unix/Linux/macOS 安装脚本"
            echo ""
            echo "用法: $0 [选项]"
            echo ""
            echo "选项:"
            echo "  --install-path PATH     安装路径 (默认: /opt/delguard)"
            echo "  --service-name NAME     服务名称 (默认: delguard)"
            echo "  --silent               静默安装"
            echo "  --no-desktop-shortcut  不创建桌面快捷方式"
            echo "  --no-path              不添加到PATH"
            echo "  --no-start             不启动服务"
            echo "  --user-service         安装为用户服务"
            echo "  --help                 显示帮助信息"
            exit 0
            ;;
        *)
            echo "未知选项: $1"
            echo "使用 --help 查看帮助信息"
            exit 1
            ;;
    esac
done

# 颜色输出函数
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${CYAN}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检测操作系统
detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        OS="linux"
        if command -v systemctl >/dev/null 2>&1; then
            INIT_SYSTEM="systemd"
        elif command -v service >/dev/null 2>&1; then
            INIT_SYSTEM="sysv"
        else
            INIT_SYSTEM="unknown"
        fi
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        OS="macos"
        INIT_SYSTEM="launchd"
    else
        log_error "不支持的操作系统: $OSTYPE"
        exit 1
    fi
}

# 检测架构
detect_arch() {
    ARCH=$(uname -m)
    case $ARCH in
        x86_64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        armv7l)
            ARCH="arm"
            ;;
        *)
            log_error "不支持的架构: $ARCH"
            exit 1
            ;;
    esac
}

# 检查权限
check_permissions() {
    if [[ $USER_SERVICE == false ]]; then
        if [[ $EUID -ne 0 ]]; then
            log_error "需要root权限才能安装系统服务"
            log_info "请使用 sudo $0 或添加 --user-service 参数安装用户服务"
            exit 1
        fi
    fi
}

# 检查依赖
check_dependencies() {
    log_info "检查系统依赖..."
    
    local missing_deps=()
    
    # 检查基本工具
    for cmd in curl tar; do
        if ! command -v $cmd >/dev/null 2>&1; then
            missing_deps+=($cmd)
        fi
    done
    
    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        log_error "缺少必要的依赖: ${missing_deps[*]}"
        log_info "请先安装这些依赖，然后重新运行安装脚本"
        
        case $OS in
            linux)
                if command -v apt-get >/dev/null 2>&1; then
                    log_info "Ubuntu/Debian: sudo apt-get install ${missing_deps[*]}"
                elif command -v yum >/dev/null 2>&1; then
                    log_info "CentOS/RHEL: sudo yum install ${missing_deps[*]}"
                elif command -v dnf >/dev/null 2>&1; then
                    log_info "Fedora: sudo dnf install ${missing_deps[*]}"
                fi
                ;;
            macos)
                log_info "macOS: brew install ${missing_deps[*]}"
                ;;
        esac
        exit 1
    fi
    
    log_success "依赖检查通过"
}

# 获取系统信息
get_system_info() {
    log_info "系统信息:"
    log_info "  操作系统: $OS"
    log_info "  架构: $ARCH"
    log_info "  初始化系统: $INIT_SYSTEM"
    log_info "  用户: $USER"
    log_info "  安装路径: $INSTALL_PATH"
    log_info "  服务名称: $SERVICE_NAME"
    echo ""
}

# 创建安装目录
create_install_directory() {
    log_info "创建安装目录: $INSTALL_PATH"
    
    if [[ -d "$INSTALL_PATH" ]]; then
        log_warning "安装目录已存在，将进行覆盖安装"
        # 停止服务（如果正在运行）
        stop_service_silent
    fi
    
    mkdir -p "$INSTALL_PATH"
    mkdir -p "$INSTALL_PATH/data"
    mkdir -p "$INSTALL_PATH/backups"
    mkdir -p "$INSTALL_PATH/logs"
    
    log_success "安装目录创建成功"
}

# 复制文件
copy_delguard_files() {
    log_info "复制DelGuard文件到安装目录..."
    
    # 获取脚本目录
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
    
    # 复制主程序
    local exe_name="delguard"
    if [[ $OS == "linux" ]]; then
        exe_name="delguard_linux_$ARCH"
    elif [[ $OS == "macos" ]]; then
        exe_name="delguard_darwin_$ARCH"
    fi
    
    local exe_path="$PROJECT_ROOT/$exe_name"
    if [[ ! -f "$exe_path" ]]; then
        # 尝试通用名称
        exe_path="$PROJECT_ROOT/delguard"
    fi
    
    if [[ -f "$exe_path" ]]; then
        cp "$exe_path" "$INSTALL_PATH/delguard"
        chmod +x "$INSTALL_PATH/delguard"
        log_success "主程序复制成功"
    else
        log_error "找不到DelGuard主程序: $exe_path"
        exit 1
    fi
    
    # 复制配置文件
    if [[ -d "$PROJECT_ROOT/configs" ]]; then
        cp -r "$PROJECT_ROOT/configs" "$INSTALL_PATH/"
        log_success "配置文件复制成功"
    fi
    
    # 复制文档
    if [[ -d "$PROJECT_ROOT/docs" ]]; then
        cp -r "$PROJECT_ROOT/docs" "$INSTALL_PATH/"
        log_success "文档文件复制成功"
    fi
}

# 创建配置文件
create_config_file() {
    log_info "创建配置文件..."
    
    local config_path="$INSTALL_PATH/config.yaml"
    cat > "$config_path" << EOF
# DelGuard 配置文件
# 自动生成于: $(date '+%Y-%m-%d %H:%M:%S')

app:
  name: "DelGuard"
  version: "2.0.0"
  log_level: "info"
  data_dir: "$INSTALL_PATH/data"
  log_dir: "$INSTALL_PATH/logs"

monitor:
  enabled: true
  watch_paths:
    - "$HOME"
    - "/home"
  exclude_paths:
    - "/proc"
    - "/sys"
    - "/dev"
    - "/tmp"
    - "/var/tmp"
  file_types:
    - ".doc"
    - ".docx"
    - ".xls"
    - ".xlsx"
    - ".ppt"
    - ".pptx"
    - ".pdf"
    - ".txt"
    - ".jpg"
    - ".png"
    - ".mp4"
    - ".mp3"

restore:
  backup_dir: "$INSTALL_PATH/backups"
  max_backup_size: "10GB"
  retention_days: 30

search:
  index_enabled: true
  index_update_interval: "1h"
  max_results: 1000

security:
  enable_encryption: true
  require_admin: false
  audit_log: true
EOF
    
    log_success "配置文件创建成功: $config_path"
}

# 创建systemd服务文件
create_systemd_service() {
    log_info "创建systemd服务..."
    
    local service_file
    local service_dir
    
    if [[ $USER_SERVICE == true ]]; then
        service_dir="$HOME/.config/systemd/user"
        service_file="$service_dir/$SERVICE_NAME.service"
        mkdir -p "$service_dir"
    else
        service_dir="/etc/systemd/system"
        service_file="$service_dir/$SERVICE_NAME.service"
    fi
    
    cat > "$service_file" << EOF
[Unit]
Description=DelGuard File Protection Service
Documentation=https://github.com/delguard/delguard
After=network.target
Wants=network.target

[Service]
Type=simple
User=$(if [[ $USER_SERVICE == true ]]; then echo "$USER"; else echo "root"; fi)
Group=$(if [[ $USER_SERVICE == true ]]; then echo "$USER"; else echo "root"; fi)
ExecStart=$INSTALL_PATH/delguard service
ExecReload=/bin/kill -HUP \$MAINPID
WorkingDirectory=$INSTALL_PATH
Environment=DELGUARD_CONFIG=$INSTALL_PATH/config.yaml
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=delguard

# 安全设置
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=$(if [[ $USER_SERVICE == true ]]; then echo "false"; else echo "read-only"; fi)
ReadWritePaths=$INSTALL_PATH

[Install]
WantedBy=$(if [[ $USER_SERVICE == true ]]; then echo "default.target"; else echo "multi-user.target"; fi)
EOF
    
    # 重新加载systemd
    if [[ $USER_SERVICE == true ]]; then
        systemctl --user daemon-reload
    else
        systemctl daemon-reload
    fi
    
    log_success "systemd服务创建成功"
}

# 创建launchd服务文件 (macOS)
create_launchd_service() {
    log_info "创建launchd服务..."
    
    local plist_dir
    local plist_file
    
    if [[ $USER_SERVICE == true ]]; then
        plist_dir="$HOME/Library/LaunchAgents"
        plist_file="$plist_dir/com.delguard.$SERVICE_NAME.plist"
    else
        plist_dir="/Library/LaunchDaemons"
        plist_file="$plist_dir/com.delguard.$SERVICE_NAME.plist"
    fi
    
    mkdir -p "$plist_dir"
    
    cat > "$plist_file" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.delguard.$SERVICE_NAME</string>
    <key>ProgramArguments</key>
    <array>
        <string>$INSTALL_PATH/delguard</string>
        <string>service</string>
    </array>
    <key>WorkingDirectory</key>
    <string>$INSTALL_PATH</string>
    <key>EnvironmentVariables</key>
    <dict>
        <key>DELGUARD_CONFIG</key>
        <string>$INSTALL_PATH/config.yaml</string>
    </dict>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>$INSTALL_PATH/logs/delguard.log</string>
    <key>StandardErrorPath</key>
    <string>$INSTALL_PATH/logs/delguard.error.log</string>
</dict>
</plist>
EOF
    
    log_success "launchd服务创建成功"
}

# 注册服务
register_service() {
    log_info "注册服务..."
    
    case $INIT_SYSTEM in
        systemd)
            create_systemd_service
            if [[ $USER_SERVICE == true ]]; then
                systemctl --user enable "$SERVICE_NAME"
            else
                systemctl enable "$SERVICE_NAME"
            fi
            ;;
        launchd)
            create_launchd_service
            ;;
        sysv)
            log_warning "SysV init系统支持有限，建议手动配置服务"
            ;;
        *)
            log_warning "未知的初始化系统，跳过服务注册"
            ;;
    esac
    
    log_success "服务注册成功"
}

# 启动服务
start_service() {
    log_info "启动DelGuard服务..."
    
    case $INIT_SYSTEM in
        systemd)
            if [[ $USER_SERVICE == true ]]; then
                systemctl --user start "$SERVICE_NAME"
                sleep 2
                if systemctl --user is-active --quiet "$SERVICE_NAME"; then
                    log_success "服务启动成功"
                else
                    log_error "服务启动失败"
                    systemctl --user status "$SERVICE_NAME"
                    return 1
                fi
            else
                systemctl start "$SERVICE_NAME"
                sleep 2
                if systemctl is-active --quiet "$SERVICE_NAME"; then
                    log_success "服务启动成功"
                else
                    log_error "服务启动失败"
                    systemctl status "$SERVICE_NAME"
                    return 1
                fi
            fi
            ;;
        launchd)
            if [[ $USER_SERVICE == true ]]; then
                launchctl load "$HOME/Library/LaunchAgents/com.delguard.$SERVICE_NAME.plist"
            else
                launchctl load "/Library/LaunchDaemons/com.delguard.$SERVICE_NAME.plist"
            fi
            sleep 2
            log_success "服务启动成功"
            ;;
        *)
            log_warning "无法自动启动服务，请手动启动"
            ;;
    esac
}

# 停止服务（静默）
stop_service_silent() {
    case $INIT_SYSTEM in
        systemd)
            if [[ $USER_SERVICE == true ]]; then
                systemctl --user stop "$SERVICE_NAME" 2>/dev/null || true
            else
                systemctl stop "$SERVICE_NAME" 2>/dev/null || true
            fi
            ;;
        launchd)
            if [[ $USER_SERVICE == true ]]; then
                launchctl unload "$HOME/Library/LaunchAgents/com.delguard.$SERVICE_NAME.plist" 2>/dev/null || true
            else
                launchctl unload "/Library/LaunchDaemons/com.delguard.$SERVICE_NAME.plist" 2>/dev/null || true
            fi
            ;;
    esac
}

# 添加到PATH
add_to_path() {
    log_info "添加到系统PATH..."
    
    local shell_rc
    case $SHELL in
        */bash)
            shell_rc="$HOME/.bashrc"
            ;;
        */zsh)
            shell_rc="$HOME/.zshrc"
            ;;
        */fish)
            shell_rc="$HOME/.config/fish/config.fish"
            ;;
        *)
            shell_rc="$HOME/.profile"
            ;;
    esac
    
    if [[ -f "$shell_rc" ]] && ! grep -q "$INSTALL_PATH" "$shell_rc"; then
        echo "" >> "$shell_rc"
        echo "# DelGuard PATH" >> "$shell_rc"
        echo "export PATH=\"$INSTALL_PATH:\$PATH\"" >> "$shell_rc"
        log_success "已添加到PATH ($shell_rc)"
        log_info "请重新加载shell配置: source $shell_rc"
    else
        log_info "PATH中已存在或无法确定shell配置文件"
    fi
}

# 创建桌面快捷方式
create_desktop_shortcut() {
    log_info "创建桌面快捷方式..."
    
    local desktop_dir="$HOME/Desktop"
    if [[ ! -d "$desktop_dir" ]]; then
        desktop_dir="$HOME/桌面"
    fi
    
    if [[ -d "$desktop_dir" ]]; then
        local shortcut_file="$desktop_dir/DelGuard.desktop"
        cat > "$shortcut_file" << EOF
[Desktop Entry]
Version=1.0
Type=Application
Name=DelGuard
Comment=DelGuard文件删除保护工具
Exec=$INSTALL_PATH/delguard
Icon=$INSTALL_PATH/icon.png
Terminal=false
Categories=Utility;System;
EOF
        chmod +x "$shortcut_file"
        log_success "桌面快捷方式创建成功"
    else
        log_warning "未找到桌面目录，跳过快捷方式创建"
    fi
}

# 创建卸载脚本
create_uninstall_script() {
    log_info "创建卸载脚本..."
    
    local uninstall_script="$INSTALL_PATH/uninstall.sh"
    cat > "$uninstall_script" << 'EOF'
#!/bin/bash
# DelGuard 卸载脚本

set -e

INSTALL_PATH="__INSTALL_PATH__"
SERVICE_NAME="__SERVICE_NAME__"
USER_SERVICE=__USER_SERVICE__
INIT_SYSTEM="__INIT_SYSTEM__"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${YELLOW}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

if [[ "$1" != "--silent" ]]; then
    echo "确定要卸载DelGuard吗？(y/N)"
    read -r confirm
    if [[ "$confirm" != "y" && "$confirm" != "Y" ]]; then
        echo "取消卸载"
        exit 0
    fi
fi

log_info "正在卸载DelGuard..."

# 停止并删除服务
case $INIT_SYSTEM in
    systemd)
        if [[ $USER_SERVICE == true ]]; then
            systemctl --user stop "$SERVICE_NAME" 2>/dev/null || true
            systemctl --user disable "$SERVICE_NAME" 2>/dev/null || true
            rm -f "$HOME/.config/systemd/user/$SERVICE_NAME.service"
            systemctl --user daemon-reload
        else
            systemctl stop "$SERVICE_NAME" 2>/dev/null || true
            systemctl disable "$SERVICE_NAME" 2>/dev/null || true
            rm -f "/etc/systemd/system/$SERVICE_NAME.service"
            systemctl daemon-reload
        fi
        ;;
    launchd)
        if [[ $USER_SERVICE == true ]]; then
            launchctl unload "$HOME/Library/LaunchAgents/com.delguard.$SERVICE_NAME.plist" 2>/dev/null || true
            rm -f "$HOME/Library/LaunchAgents/com.delguard.$SERVICE_NAME.plist"
        else
            launchctl unload "/Library/LaunchDaemons/com.delguard.$SERVICE_NAME.plist" 2>/dev/null || true
            rm -f "/Library/LaunchDaemons/com.delguard.$SERVICE_NAME.plist"
        fi
        ;;
esac
log_success "服务已删除"

# 从PATH中移除
for rc_file in "$HOME/.bashrc" "$HOME/.zshrc" "$HOME/.profile" "$HOME/.config/fish/config.fish"; do
    if [[ -f "$rc_file" ]]; then
        sed -i.bak "/# DelGuard PATH/d" "$rc_file" 2>/dev/null || true
        sed -i.bak "\|export PATH=\"$INSTALL_PATH:\$PATH\"|d" "$rc_file" 2>/dev/null || true
    fi
done
log_success "已从PATH中移除"

# 删除快捷方式
rm -f "$HOME/Desktop/DelGuard.desktop" "$HOME/桌面/DelGuard.desktop" 2>/dev/null || true
log_success "快捷方式已删除"

# 删除安装目录
cd /tmp
rm -rf "$INSTALL_PATH"
log_success "安装目录已删除"

log_success "DelGuard卸载完成"
EOF
    
    # 替换占位符
    sed -i.bak "s|__INSTALL_PATH__|$INSTALL_PATH|g" "$uninstall_script"
    sed -i.bak "s|__SERVICE_NAME__|$SERVICE_NAME|g" "$uninstall_script"
    sed -i.bak "s|__USER_SERVICE__|$USER_SERVICE|g" "$uninstall_script"
    sed -i.bak "s|__INIT_SYSTEM__|$INIT_SYSTEM|g" "$uninstall_script"
    rm -f "$uninstall_script.bak"
    
    chmod +x "$uninstall_script"
    log_success "卸载脚本创建成功"
}

# 验证安装
verify_installation() {
    log_info "验证安装..."
    
    local issues=()
    
    # 检查主程序
    if [[ ! -f "$INSTALL_PATH/delguard" ]]; then
        issues+=("主程序文件不存在")
    fi
    
    # 检查配置文件
    if [[ ! -f "$INSTALL_PATH/config.yaml" ]]; then
        issues+=("配置文件不存在")
    fi
    
    # 检查服务状态
    case $INIT_SYSTEM in
        systemd)
            if [[ $USER_SERVICE == true ]]; then
                if ! systemctl --user is-enabled --quiet "$SERVICE_NAME" 2>/dev/null; then
                    issues+=("服务未启用")
                fi
                if ! systemctl --user is-active --quiet "$SERVICE_NAME" 2>/dev/null; then
                    issues+=("服务未运行")
                fi
            else
                if ! systemctl is-enabled --quiet "$SERVICE_NAME" 2>/dev/null; then
                    issues+=("服务未启用")
                fi
                if ! systemctl is-active --quiet "$SERVICE_NAME" 2>/dev/null; then
                    issues+=("服务未运行")
                fi
            fi
            ;;
    esac
    
    if [[ ${#issues[@]} -eq 0 ]]; then
        log_success "安装验证通过"
        return 0
    else
        log_error "安装验证失败:"
        for issue in "${issues[@]}"; do
            log_error "  - $issue"
        done
        return 1
    fi
}

# 主安装流程
main() {
    echo -e "${CYAN}"
    cat << 'EOF'
╔══════════════════════════════════════════════════════════════╗
║                    DelGuard 安装程序                         ║
║                     版本: 2.0.0                             ║
╚══════════════════════════════════════════════════════════════╝
EOF
    echo -e "${NC}"
    
    # 检测系统环境
    detect_os
    detect_arch
    check_permissions
    check_dependencies
    
    # 显示系统信息
    if [[ $SILENT == false ]]; then
        get_system_info
        
        echo "是否继续安装？(Y/n)"
        read -r confirm
        if [[ "$confirm" == "n" || "$confirm" == "N" ]]; then
            log_info "安装已取消"
            exit 0
        fi
    fi
    
    log_info "开始安装DelGuard..."
    
    # 执行安装步骤
    create_install_directory
    copy_delguard_files
    create_config_file
    register_service
    
    if [[ $ADD_TO_PATH == true ]]; then
        add_to_path
    fi
    
    if [[ $CREATE_DESKTOP_SHORTCUT == true ]]; then
        create_desktop_shortcut
    fi
    
    create_uninstall_script
    
    if [[ $START_SERVICE == true ]]; then
        start_service
    fi
    
    # 验证安装
    if verify_installation; then
        echo -e "${GREEN}"
        cat << 'EOF'
╔══════════════════════════════════════════════════════════════╗
║                   DelGuard 安装成功！                        ║
╚══════════════════════════════════════════════════════════════╝
EOF
        echo -e "${NC}"
        log_success "安装路径: $INSTALL_PATH"
        log_success "服务状态: 运行中"
        log_success "配置文件: $INSTALL_PATH/config.yaml"
        echo ""
        log_info "使用方法:"
        log_info "  命令行: delguard --help"
        log_info "  服务管理: systemctl status $SERVICE_NAME"
        log_info "  卸载: $INSTALL_PATH/uninstall.sh"
    else
        log_error "安装验证失败"
        exit 1
    fi
}

# 执行主函数
main "$@"