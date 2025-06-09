#!/bin/bash

# Compman 一键安装脚本 (简化版)
# 使用方法: curl -fsSL https://raw.githubusercontent.com/QuentinHsu/compman/main/install-online.sh | bash

set -e

# 基本配置
GITHUB_USER="QuentinHsu"  # 请替换为您的 GitHub 用户名
GITHUB_REPO="compman"       # 请替换为您的仓库名
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="compman"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${CYAN}🚀 Compman 一键安装脚本${NC}"
echo "=================================="

# 检测系统
detect_system() {
    case "$(uname -s)" in
        Darwin) OS="darwin" ;;
        Linux) OS="linux" ;;
        *) echo -e "${RED}❌ 不支持的操作系统${NC}"; exit 1 ;;
    esac
    
    case "$(uname -m)" in
        x86_64|amd64) ARCH="amd64" ;;
        arm64|aarch64) ARCH="arm64" ;;
        *) echo -e "${RED}❌ 不支持的架构${NC}"; exit 1 ;;
    esac
    
    echo -e "${BLUE}📍 检测到系统: ${OS}/${ARCH}${NC}"
}

# 下载并安装
install_compman() {
    echo -e "${YELLOW}📥 正在下载 Compman...${NC}"
    
    # 获取最新版本
    VERSION=$(curl -s "https://api.github.com/repos/${GITHUB_USER}/${GITHUB_REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' || echo "latest")
    
    # 构建下载URL
    ARCHIVE_NAME="${BINARY_NAME}-${VERSION}-${OS}-${ARCH}.tar.gz"
    DOWNLOAD_URL="https://github.com/${GITHUB_USER}/${GITHUB_REPO}/releases/download/${VERSION}/${ARCHIVE_NAME}"
    
    # 创建临时目录
    TMP_DIR=$(mktemp -d)
    cd "$TMP_DIR"
    
    # 下载文件
    if ! curl -fsSL -o "$ARCHIVE_NAME" "$DOWNLOAD_URL"; then
        echo -e "${RED}❌ 下载失败${NC}"
        exit 1
    fi
    
    # 解压
    tar -xzf "$ARCHIVE_NAME"
    
    # 安装
    BINARY_FILE="${BINARY_NAME}-${VERSION}-${OS}-${ARCH}"
    if [ ! -f "$BINARY_FILE" ]; then
        echo -e "${RED}❌ 二进制文件不存在${NC}"
        exit 1
    fi
    
    # 复制到安装目录
    if [ -w "$INSTALL_DIR" ]; then
        cp "$BINARY_FILE" "$INSTALL_DIR/$BINARY_NAME"
    else
        sudo cp "$BINARY_FILE" "$INSTALL_DIR/$BINARY_NAME"
    fi
    
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
    
    # 清理
    cd /
    rm -rf "$TMP_DIR"
    
    echo -e "${GREEN}✅ 安装完成！${NC}"
    echo -e "${BLUE}📍 安装位置: $INSTALL_DIR/$BINARY_NAME${NC}"
    echo ""
    echo -e "${YELLOW}🚀 快速开始:${NC}"
    echo "  compman --help     # 查看帮助"
    echo "  compman --version  # 查看版本"
    echo "  compman update     # 更新服务"
}

# 主执行流程
main() {
    # 检查依赖
    for cmd in curl tar; do
        if ! command -v "$cmd" &> /dev/null; then
            echo -e "${RED}❌ 缺少必需工具: $cmd${NC}"
            exit 1
        fi
    done
    
    detect_system
    install_compman
}

main "$@"
