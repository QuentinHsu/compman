#!/bin/bash

# 高级构建脚本 - 支持压缩、校验和生成
# 支持 macOS (Intel/Apple Silicon) 和 Linux (amd64/arm64)

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 项目信息
PROJECT_NAME="compman"
VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 构建选项
COMPRESS=${COMPRESS:-true}
CHECKSUM=${CHECKSUM:-true}
STRIP=${STRIP:-true}

# 构建信息
LDFLAGS="-X main.version=${VERSION} -X main.buildDate=${BUILD_TIME}"
if [ "$STRIP" = "true" ]; then
    LDFLAGS="${LDFLAGS} -w -s"
fi

# 输出目录
BUILD_DIR="dist"
ARCHIVE_DIR="${BUILD_DIR}/archives"
rm -rf ${BUILD_DIR}
mkdir -p ${BUILD_DIR} ${ARCHIVE_DIR}

# 支持的平台和架构
PLATFORMS="darwin/amd64:macOS_Intel darwin/arm64:macOS_Apple_Silicon linux/amd64:Linux_x86_64 linux/arm64:Linux_ARM64"

print_header() {
    echo -e "${CYAN}================================${NC}"
    echo -e "${CYAN}🚀 ${PROJECT_NAME} 多平台构建脚本${NC}"
    echo -e "${CYAN}================================${NC}"
    echo -e "${BLUE}📦 项目名称:${NC} ${PROJECT_NAME}"
    echo -e "${BLUE}🏷️  版本号:${NC} ${VERSION}"
    echo -e "${BLUE}📝 Git提交:${NC} ${GIT_COMMIT}"
    echo -e "${BLUE}⏰ 构建时间:${NC} ${BUILD_TIME}"
    echo -e "${BLUE}📁 输出目录:${NC} ${BUILD_DIR}"
    echo ""
    echo -e "${YELLOW}构建选项:${NC}"
    echo -e "  压缩: ${COMPRESS}"
    echo -e "  校验和: ${CHECKSUM}"
    echo -e "  Strip符号: ${STRIP}"
    echo ""
}

build_platform() {
    local platform_info=$1
    local platform=$(echo "$platform_info" | cut -d':' -f1)
    local platform_name=$(echo "$platform_info" | cut -d':' -f2 | sed 's/_/ /g')
    
    local GOOS=$(echo "$platform" | cut -d'/' -f1)
    local GOARCH=$(echo "$platform" | cut -d'/' -f2)
    
    echo -e "${PURPLE}🔨 构建 ${platform_name} (${GOOS}/${GOARCH})...${NC}"
    
    # 设置输出文件名
    OUTPUT_NAME="${PROJECT_NAME}-${VERSION}-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        OUTPUT_NAME="${OUTPUT_NAME}.exe"
    fi
    
    # 执行构建
    if GOOS=${GOOS} GOARCH=${GOARCH} go build \
        -ldflags="${LDFLAGS}" \
        -o "${BUILD_DIR}/${OUTPUT_NAME}" \
        ./cmd; then
        
        # 显示文件大小
        SIZE=$(du -h "${BUILD_DIR}/${OUTPUT_NAME}" | cut -f1)
        echo -e "${GREEN}✅ ${platform_name}: ${OUTPUT_NAME} (${SIZE})${NC}"
        
        # 创建压缩包
        if [ "$COMPRESS" = "true" ]; then
            create_archive "$OUTPUT_NAME" "$platform_name" "$GOOS"
        fi
        
        return 0
    else
        echo -e "${RED}❌ ${platform_name}: 构建失败${NC}"
        return 1
    fi
}

create_archive() {
    local binary_name=$1
    local platform_name=$2
    local os=$3
    
    local archive_name="${binary_name}"
    
    # 根据操作系统选择压缩格式
    if [ "$os" = "windows" ]; then
        archive_name="${archive_name}.zip"
        (cd "${BUILD_DIR}" && zip -q "${archive_name}" "${binary_name}")
    else
        archive_name="${archive_name}.tar.gz"
        (cd "${BUILD_DIR}" && tar -czf "${archive_name}" "${binary_name}")
    fi
    
    mv "${BUILD_DIR}/${archive_name}" "${ARCHIVE_DIR}/"
    
    local archive_size=$(du -h "${ARCHIVE_DIR}/${archive_name}" | cut -f1)
    echo -e "${CYAN}📦 压缩包: ${archive_name} (${archive_size})${NC}"
}

generate_checksums() {
    echo -e "${YELLOW}🔍 生成校验和文件...${NC}"
    
    # 检测校验和命令
    local sha_cmd
    if command -v sha256sum &> /dev/null; then
        sha_cmd="sha256sum"
    elif command -v shasum &> /dev/null; then
        sha_cmd="shasum -a 256"
    else
        echo -e "${RED}❌ 找不到 sha256sum 或 shasum 命令${NC}"
        return 1
    fi
    
    # 创建统一的校验和文件
    local checksum_file="${BUILD_DIR}/checksums.txt"
    rm -f "${checksum_file}"
    
    # 生成二进制文件校验和
    echo "# Binary files" >> "${checksum_file}"
    (cd "${BUILD_DIR}" && for file in ${PROJECT_NAME}-*; do
        if [[ -f "$file" && ! "$file" =~ \.(tar\.gz|zip)$ ]]; then
            ${sha_cmd} "$file"
        fi
    done) >> "${checksum_file}"
    
    # 生成压缩包校验和
    if [ "$COMPRESS" = "true" ]; then
        echo "" >> "${checksum_file}"
        echo "# Archive files" >> "${checksum_file}"
        (cd "${ARCHIVE_DIR}" && for file in *.tar.gz *.zip; do
            if [[ -f "$file" ]]; then
                ${sha_cmd} "$file" | sed 's|^|archives/|'
            fi
        done 2>/dev/null) >> "${checksum_file}"
    fi
    
    echo -e "${GREEN}✅ 校验和文件已生成${NC}"
}

print_summary() {
    echo ""
    echo -e "${CYAN}================================${NC}"
    echo -e "${GREEN}🎉 所有平台构建完成！${NC}"
    echo -e "${CYAN}================================${NC}"
    echo ""
    
    echo -e "${BLUE}📁 二进制文件:${NC}"
    ls -lh ${BUILD_DIR}/${PROJECT_NAME}-* 2>/dev/null | while read line; do
        echo "  $line"
    done
    
    if [ "$COMPRESS" = "true" ]; then
        echo ""
        echo -e "${BLUE}📦 压缩包:${NC}"
        ls -lh ${ARCHIVE_DIR}/* 2>/dev/null | while read line; do
            echo "  $line"
        done
    fi
    
    if [ "$CHECKSUM" = "true" ]; then
        echo ""
        echo -e "${BLUE}🔍 校验和文件:${NC}"
        [ -f "${BUILD_DIR}/checksums.txt" ] && echo "  ${BUILD_DIR}/checksums.txt"
    fi
    
    echo ""
    echo -e "${YELLOW}💡 使用方法:${NC}"
    echo -e "  macOS Intel:     ./${BUILD_DIR}/${PROJECT_NAME}-${VERSION}-darwin-amd64"
    echo -e "  macOS ARM64:     ./${BUILD_DIR}/${PROJECT_NAME}-${VERSION}-darwin-arm64"
    echo -e "  Linux x86_64:    ./${BUILD_DIR}/${PROJECT_NAME}-${VERSION}-linux-amd64"
    echo -e "  Linux ARM64:     ./${BUILD_DIR}/${PROJECT_NAME}-${VERSION}-linux-arm64"
}

# 主执行流程
main() {
    print_header
    
    local failed_builds=0
    
    # 构建所有平台
    for platform_info in $PLATFORMS; do
        if ! build_platform "$platform_info"; then
            ((failed_builds++))
        fi
        echo ""
    done
    
    # 生成校验和
    if [ "$CHECKSUM" = "true" ] && [ $failed_builds -eq 0 ]; then
        generate_checksums
    fi
    
    # 显示构建总结
    print_summary
    
    # 检查是否有失败的构建
    if [ $failed_builds -gt 0 ]; then
        echo -e "${RED}⚠️  有 $failed_builds 个平台构建失败${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}🎊 构建成功完成！${NC}"
}

# 脚本帮助
show_help() {
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help     显示帮助信息"
    echo "  --no-compress  禁用压缩"
    echo "  --no-checksum  禁用校验和生成"
    echo "  --no-strip     禁用符号表删除"
    echo ""
    echo "环境变量:"
    echo "  VERSION        设置版本号 (默认: git describe)"
    echo "  COMPRESS       启用/禁用压缩 (默认: true)"
    echo "  CHECKSUM       启用/禁用校验和 (默认: true)"
    echo "  STRIP          启用/禁用符号表删除 (默认: true)"
    echo ""
    echo "示例:"
    echo "  $0                          # 标准构建"
    echo "  VERSION=v1.0.0 $0           # 指定版本号"
    echo "  $0 --no-compress            # 不压缩"
    echo "  COMPRESS=false $0           # 环境变量方式禁用压缩"
}

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        --no-compress)
            COMPRESS=false
            shift
            ;;
        --no-checksum)
            CHECKSUM=false
            shift
            ;;
        --no-strip)
            STRIP=false
            shift
            ;;
        *)
            echo "未知选项: $1"
            echo "使用 $0 --help 查看帮助"
            exit 1
            ;;
    esac
done

# 检查 Go 环境
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go 未安装或不在 PATH 中${NC}"
    exit 1
fi

# 执行主函数
main
