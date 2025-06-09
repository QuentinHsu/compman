# Makefile for compman

.PHONY: build build-all clean test deps help

# 项目信息
PROJECT_NAME := compman
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 构建标志
LDFLAGS := -X main.version=$(VERSION) -X main.buildDate=$(BUILD_TIME) -w -s

# 目录
BUILD_DIR := dist
SRC_DIR := ./cmd

# 默认目标
help: ## 显示帮助信息
	@echo "可用的构建目标:"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## 安装依赖
	@echo "📦 安装 Go 模块依赖..."
	go mod download
	go mod tidy

test: ## 运行测试
	@echo "🧪 运行测试..."
	go test -v ./...

clean: ## 清理构建文件
	@echo "🧹 清理构建文件..."
	rm -rf $(BUILD_DIR)

build: ## 构建当前平台版本
	@echo "🔨 构建当前平台版本..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(PROJECT_NAME) $(SRC_DIR)
	@echo "✅ 构建完成: $(BUILD_DIR)/$(PROJECT_NAME)"

build-all: ## 构建所有平台版本
	@echo "🚀 开始多平台构建..."
	@chmod +x build.sh
	./build.sh

release: clean build-all ## 创建发布版本 (清理 + 多平台构建)
	@echo "🎉 发布版本创建完成！"

# 单独构建特定平台
build-darwin-amd64: ## 构建 macOS Intel 版本
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(PROJECT_NAME)-darwin-amd64 $(SRC_DIR)

build-darwin-arm64: ## 构建 macOS Apple Silicon 版本
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(PROJECT_NAME)-darwin-arm64 $(SRC_DIR)

build-linux-amd64: ## 构建 Linux x86_64 版本
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(PROJECT_NAME)-linux-amd64 $(SRC_DIR)

build-linux-arm64: ## 构建 Linux ARM64 版本
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(PROJECT_NAME)-linux-arm64 $(SRC_DIR)

# 开发相关
dev: ## 开发模式运行
	go run $(SRC_DIR)

fmt: ## 格式化代码
	go fmt ./...

lint: ## 代码检查
	golangci-lint run

# 显示版本信息
version: ## 显示版本信息
	@echo "Project: $(PROJECT_NAME)"
	@echo "Version: $(VERSION)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"
