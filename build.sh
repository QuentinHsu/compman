#!/bin/bash

# 构建脚本 - 支持多平台交叉编译
# 支持 macOS (Intel/Apple Silicon) 和 Linux (amd64/arm64)

set -e

# 项目信息
PROJECT_NAME="compman"
VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 构建信息
LDFLAGS="-X main.version=${VERSION} -X main.buildDate=${BUILD_TIME} -s -w"

# 输出目录
BUILD_DIR="dist"
rm -rf ${BUILD_DIR}
mkdir -p ${BUILD_DIR}

# 支持的平台和架构
PLATFORMS="darwin/amd64:macOS_Intel darwin/arm64:macOS_Apple_Silicon linux/amd64:Linux_x86_64 linux/arm64:Linux_ARM64"

echo "🚀 开始构建 ${PROJECT_NAME} v${VERSION}"
echo "📦 Git Commit: ${GIT_COMMIT}"
echo "⏰ Build Time: ${BUILD_TIME}"
echo ""

# 构建所有平台
for platform_info in $PLATFORMS; do
    platform=$(echo "$platform_info" | cut -d':' -f1)
    platform_name=$(echo "$platform_info" | cut -d':' -f2 | sed 's/_/ /g')
    GOOS=$(echo "$platform" | cut -d'/' -f1)
    GOARCH=$(echo "$platform" | cut -d'/' -f2)
    
    echo "🔨 构建 ${platform_name} (${GOOS}/${GOARCH})..."
    
    # 设置输出文件名
    OUTPUT_NAME="${PROJECT_NAME}-${VERSION}-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        OUTPUT_NAME="${OUTPUT_NAME}.exe"
    fi
    
    # 执行构建
    GOOS=${GOOS} GOARCH=${GOARCH} go build \
        -ldflags="${LDFLAGS}" \
        -o "${BUILD_DIR}/${OUTPUT_NAME}" \
        ./cmd
    
    # 显示文件大小
    if [ -f "${BUILD_DIR}/${OUTPUT_NAME}" ]; then
        SIZE=$(du -h "${BUILD_DIR}/${OUTPUT_NAME}" | cut -f1)
        echo "✅ ${platform_name}: ${OUTPUT_NAME} (${SIZE})"
    else
        echo "❌ ${platform_name}: 构建失败"
        exit 1
    fi
done

echo ""
echo "🎉 所有平台构建完成！"
echo ""
echo "📁 构建文件列表:"
ls -lh ${BUILD_DIR}/

echo ""
echo "💡 使用方法:"
echo "  macOS Intel:     ./${BUILD_DIR}/${PROJECT_NAME}-${VERSION}-darwin-amd64"
echo "  macOS ARM64:     ./${BUILD_DIR}/${PROJECT_NAME}-${VERSION}-darwin-arm64"
echo "  Linux x86_64:    ./${BUILD_DIR}/${PROJECT_NAME}-${VERSION}-linux-amd64"
echo "  Linux ARM64:     ./${BUILD_DIR}/${PROJECT_NAME}-${VERSION}-linux-arm64"
