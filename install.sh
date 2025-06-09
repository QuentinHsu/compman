#!/bin/bash

# Compman 在线安装脚本
# 自动检测系统架构并下载对应的二进制文件

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# 脚本配置
GITHUB_USER="QuentinHsu"  # 请替换为您的 GitHub 用户名
GITHUB_REPO="compman"       # 请替换为您的仓库名
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="compman"
CONFIG_DIR="$HOME/.config/compman"

# 默认值
FORCE_INSTALL=false
INSTALL_PATH=""
VERSION="latest"
VERIFY_CHECKSUM=true

print_banner() {
    echo -e "${CYAN}"
    echo "╔════════════════════════════════════════════════════════════════╗"
    echo "║                    Compman 安装脚本                            ║"
    echo "║              Docker Compose Manager 一键安装                   ║"
    echo "╚════════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

print_help() {
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help              显示帮助信息"
    echo "  -v, --version VERSION   指定安装版本 (默认: latest)"
    echo "  -d, --dir PATH          指定安装目录 (默认: /usr/local/bin)"
    echo "  -f, --force             强制覆盖已存在的文件"
    echo "  --no-verify             跳过校验和验证"
    echo "  --config-only           仅创建配置文件示例"
    echo ""
    echo "示例:"
    echo "  $0                      # 安装最新版本"
    echo "  $0 -v v1.2.0           # 安装指定版本"
    echo "  $0 -d ~/bin            # 安装到自定义目录"
    echo "  $0 -f                  # 强制重新安装"
}

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

log_step() {
    echo -e "${PURPLE}🔧 $1${NC}"
}

# 检测系统信息
detect_system() {
    local os arch
    
    # 检测操作系统
    case "$(uname -s)" in
        Darwin)
            os="darwin"
            ;;
        Linux)
            os="linux"
            ;;
        MINGW* | MSYS* | CYGWIN*)
            os="windows"
            ;;
        *)
            log_error "不支持的操作系统: $(uname -s)"
            exit 1
            ;;
    esac
    
    # 检测架构
    case "$(uname -m)" in
        x86_64 | amd64)
            arch="amd64"
            ;;
        arm64 | aarch64)
            arch="arm64"
            ;;
        armv7l)
            arch="arm"
            ;;
        i386 | i686)
            arch="386"
            ;;
        *)
            log_error "不支持的架构: $(uname -m)"
            exit 1
            ;;
    esac
    
    SYSTEM_OS="$os"
    SYSTEM_ARCH="$arch"
    
    # 设置二进制文件名
    if [ "$os" = "windows" ]; then
        BINARY_FILENAME="${BINARY_NAME}.exe"
    else
        BINARY_FILENAME="${BINARY_NAME}"
    fi
    
    log_info "检测到系统: ${SYSTEM_OS}/${SYSTEM_ARCH}"
}

# 检查依赖工具
check_dependencies() {
    local missing_deps=()
    
    # 检查必需工具
    for cmd in curl tar; do
        if ! command -v "$cmd" &> /dev/null; then
            missing_deps+=("$cmd")
        fi
    done
    
    # 检查校验和工具
    if [ "$VERIFY_CHECKSUM" = "true" ]; then
        if ! command -v sha256sum &> /dev/null && ! command -v shasum &> /dev/null; then
            log_warning "未找到 sha256sum 或 shasum，将跳过校验和验证"
            VERIFY_CHECKSUM=false
        fi
    fi
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        log_error "缺少必需依赖: ${missing_deps[*]}"
        echo "请先安装这些工具后重试。"
        exit 1
    fi
}

# 获取最新版本信息
get_latest_version() {
    log_step "获取最新版本信息..."
    
    local api_url="https://api.github.com/repos/${GITHUB_USER}/${GITHUB_REPO}/releases/latest"
    local latest_version
    
    if latest_version=$(curl -s "$api_url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'); then
        if [ -n "$latest_version" ] && [ "$latest_version" != "null" ]; then
            echo "$latest_version"
            return 0
        fi
    fi
    
    log_warning "无法获取最新版本，使用 'latest' 标签"
    echo "latest"
}

# 构建下载 URL
build_download_url() {
    local version="$1"
    local base_url="https://github.com/${GITHUB_USER}/${GITHUB_REPO}/releases"
    
    if [ "$version" = "latest" ]; then
        version=$(get_latest_version)
    fi
    
    # 构建文件名 (基于构建脚本的命名规则)
    local archive_name="${BINARY_NAME}-${version}-${SYSTEM_OS}-${SYSTEM_ARCH}"
    
    if [ "$SYSTEM_OS" = "windows" ]; then
        archive_name="${archive_name}.zip"
        DOWNLOAD_URL="${base_url}/download/${version}/${archive_name}"
    else
        archive_name="${archive_name}.tar.gz"
        DOWNLOAD_URL="${base_url}/download/${version}/${archive_name}"
    fi
    
    CHECKSUM_URL="${base_url}/download/${version}/checksums.txt"
    DOWNLOAD_VERSION="$version"
    
    log_info "下载版本: $version"
    log_info "下载地址: $DOWNLOAD_URL"
}

# 下载和验证文件
download_binary() {
    log_step "下载二进制文件..."
    
    local temp_dir
    temp_dir=$(mktemp -d)
    local archive_file="$temp_dir/$(basename "$DOWNLOAD_URL")"
    
    # 下载文件
    if ! curl -fsSL -o "$archive_file" "$DOWNLOAD_URL"; then
        log_error "下载失败: $DOWNLOAD_URL"
        rm -rf "$temp_dir"
        exit 1
    fi
    
    log_success "下载完成"
    
    # 验证校验和
    if [ "$VERIFY_CHECKSUM" = "true" ]; then
        verify_checksum "$archive_file" "$temp_dir"
    fi
    
    # 解压文件
    extract_binary "$archive_file" "$temp_dir"
    
    # 清理临时文件
    rm -rf "$temp_dir"
}

# 验证校验和
verify_checksum() {
    local archive_file="$1"
    local temp_dir="$2"
    
    log_step "验证文件校验和..."
    
    local checksum_file="$temp_dir/checksums.txt"
    
    # 下载校验和文件
    if ! curl -fsSL -o "$checksum_file" "$CHECKSUM_URL" 2>/dev/null; then
        log_warning "无法下载校验和文件，跳过验证"
        return 0
    fi
    
    # 选择校验和命令
    local sha_cmd
    if command -v sha256sum &> /dev/null; then
        sha_cmd="sha256sum"
    elif command -v shasum &> /dev/null; then
        sha_cmd="shasum -a 256"
    else
        log_warning "无校验和工具，跳过验证"
        return 0
    fi
    
    # 获取期望的校验和
    local archive_name
    archive_name=$(basename "$archive_file")
    local expected_hash
    expected_hash=$(grep "archives/$archive_name" "$checksum_file" | cut -d' ' -f1)
    
    if [ -z "$expected_hash" ]; then
        log_warning "校验和文件中未找到 $archive_name，跳过验证"
        return 0
    fi
    
    # 计算实际校验和
    local actual_hash
    actual_hash=$($sha_cmd "$archive_file" | cut -d' ' -f1)
    
    if [ "$expected_hash" = "$actual_hash" ]; then
        log_success "校验和验证通过"
    else
        log_error "校验和验证失败!"
        log_error "期望: $expected_hash"
        log_error "实际: $actual_hash"
        exit 1
    fi
}

# 解压二进制文件
extract_binary() {
    local archive_file="$1"
    local temp_dir="$2"
    
    log_step "解压文件..."
    
    # 构建期望的二进制文件名
    local expected_binary="${BINARY_NAME}-${DOWNLOAD_VERSION}-${SYSTEM_OS}-${SYSTEM_ARCH}"
    if [ "$SYSTEM_OS" = "windows" ]; then
        expected_binary="${expected_binary}.exe"
    fi
    
    # 解压文件
    if [ "$SYSTEM_OS" = "windows" ]; then
        if command -v unzip &> /dev/null; then
            unzip -q "$archive_file" -d "$temp_dir"
        else
            log_error "需要 unzip 工具来解压 Windows 版本"
            exit 1
        fi
    else
        tar -xzf "$archive_file" -C "$temp_dir"
    fi
    
    # 查找二进制文件
    local binary_path="$temp_dir/$expected_binary"
    if [ ! -f "$binary_path" ]; then
        log_error "解压后未找到二进制文件: $expected_binary"
        exit 1
    fi
    
    # 设置全局变量
    EXTRACTED_BINARY="$binary_path"
    log_success "解压完成"
}

# 安装二进制文件
install_binary() {
    log_step "安装二进制文件..."
    
    # 检查目标目录
    if [ ! -d "$INSTALL_PATH" ]; then
        log_error "安装目录不存在: $INSTALL_PATH"
        exit 1
    fi
    
    local target_file="$INSTALL_PATH/$BINARY_FILENAME"
    
    # 检查是否已存在
    if [ -f "$target_file" ] && [ "$FORCE_INSTALL" != "true" ]; then
        log_warning "文件已存在: $target_file"
        echo -n "是否覆盖? [y/N]: "
        read -r response
        if [ "$response" != "y" ] && [ "$response" != "Y" ]; then
            log_info "安装取消"
            exit 0
        fi
    fi
    
    # 复制文件
    if ! cp "$EXTRACTED_BINARY" "$target_file"; then
        log_error "复制文件失败，可能需要管理员权限"
        log_info "尝试使用 sudo..."
        if ! sudo cp "$EXTRACTED_BINARY" "$target_file"; then
            log_error "安装失败"
            exit 1
        fi
    fi
    
    # 设置执行权限
    chmod +x "$target_file" 2>/dev/null || sudo chmod +x "$target_file"
    
    log_success "安装到: $target_file"
}

# 创建配置文件示例
create_config() {
    log_step "创建配置文件示例..."
    
    mkdir -p "$CONFIG_DIR"
    
    local config_file="$CONFIG_DIR/config.yaml"
    
    # 如果配置文件已存在，创建示例文件
    if [ -f "$config_file" ]; then
        config_file="$CONFIG_DIR/config.example.yaml"
    fi
    
    cat > "$config_file" << 'EOF'
# Compman 配置文件示例
# 将此文件复制为 config.yaml 并根据需要修改

# 全局设置
global:
  # 默认标签策略: latest 或 semver
  tag_strategy: "latest"
  
  # 是否在更新后自动清理未使用的镜像
  auto_cleanup: true
  
  # 是否启用交互模式
  interactive: true
  
  # 是否显示详细输出
  verbose: false

# Docker 设置
docker:
  # Docker socket 路径 (留空使用默认)
  socket: ""
  
  # 超时设置 (秒)
  timeout: 30

# 备份设置
backup:
  # 是否启用自动备份
  enabled: true
  
  # 备份目录
  dir: "~/.compman/backups"
  
  # 保留备份数量
  keep: 5

# 更新策略
update:
  # 要排除的镜像模式
  exclude_patterns:
    - "*:dev"
    - "*:debug"
    - "localhost/*"
  
  # 并发更新数量
  concurrency: 3

# UI 设置
ui:
  # 是否启用彩色输出
  color: true
  
  # 进度条样式
  progress_style: "bar"
EOF
    
    log_success "配置文件示例已创建: $config_file"
}

# 验证安装
verify_installation() {
    log_step "验证安装..."
    
    local binary_path="$INSTALL_PATH/$BINARY_FILENAME"
    
    if [ ! -f "$binary_path" ]; then
        log_error "安装验证失败: 文件不存在"
        return 1
    fi
    
    if [ ! -x "$binary_path" ]; then
        log_error "安装验证失败: 文件不可执行"
        return 1
    fi
    
    # 测试运行
    if ! "$binary_path" --version &> /dev/null; then
        log_error "安装验证失败: 程序无法运行"
        return 1
    fi
    
    log_success "安装验证通过"
    return 0
}

# 显示安装后信息
show_post_install() {
    echo ""
    echo -e "${CYAN}════════════════════════════════════════${NC}"
    echo -e "${GREEN}🎉 Compman 安装成功！${NC}"
    echo -e "${CYAN}════════════════════════════════════════${NC}"
    echo ""
    
    local binary_path="$INSTALL_PATH/$BINARY_FILENAME"
    
    echo -e "${BLUE}📍 安装位置:${NC} $binary_path"
    echo -e "${BLUE}📝 配置目录:${NC} $CONFIG_DIR"
    echo ""
    
    echo -e "${YELLOW}🚀 快速开始:${NC}"
    echo -e "  查看帮助:     ${BOLD}compman --help${NC}"
    echo -e "  查看版本:     ${BOLD}compman --version${NC}"
    echo -e "  更新服务:     ${BOLD}compman update${NC}"
    echo -e "  交互模式:     ${BOLD}compman update -i${NC}"
    echo ""
    
    # 检查是否在 PATH 中
    if ! echo "$PATH" | grep -q "$INSTALL_PATH"; then
        echo -e "${YELLOW}⚠️  注意: $INSTALL_PATH 不在 PATH 中${NC}"
        echo -e "   请将以下行添加到您的 shell 配置文件 (.bashrc, .zshrc 等):"
        echo -e "   ${BOLD}export PATH=\"$INSTALL_PATH:\$PATH\"${NC}"
        echo ""
    fi
    
    echo -e "${BLUE}📚 更多信息:${NC}"
    echo -e "  项目主页: https://github.com/${GITHUB_USER}/${GITHUB_REPO}"
    echo -e "  问题反馈: https://github.com/${GITHUB_USER}/${GITHUB_REPO}/issues"
    echo ""
}

# 清理函数
cleanup() {
    if [ -n "$EXTRACTED_BINARY" ] && [ -f "$EXTRACTED_BINARY" ]; then
        rm -f "$EXTRACTED_BINARY" 2>/dev/null || true
    fi
}

# 主函数
main() {
    print_banner
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                print_help
                exit 0
                ;;
            -v|--version)
                VERSION="$2"
                shift 2
                ;;
            -d|--dir)
                INSTALL_PATH="$2"
                shift 2
                ;;
            -f|--force)
                FORCE_INSTALL=true
                shift
                ;;
            --no-verify)
                VERIFY_CHECKSUM=false
                shift
                ;;
            --config-only)
                create_config
                exit 0
                ;;
            *)
                log_error "未知选项: $1"
                echo "使用 $0 --help 查看帮助"
                exit 1
                ;;
        esac
    done
    
    # 设置默认安装路径
    if [ -z "$INSTALL_PATH" ]; then
        INSTALL_PATH="$INSTALL_DIR"
    fi
    
    # 设置清理陷阱
    trap cleanup EXIT
    
    # 执行安装步骤
    detect_system
    check_dependencies
    build_download_url "$VERSION"
    download_binary
    install_binary
    create_config
    
    if verify_installation; then
        show_post_install
    else
        log_error "安装完成但验证失败，请检查安装"
        exit 1
    fi
}

# 执行主函数
main "$@"
