# Compman 安装指南

本指南提供了多种安装 Compman (Docker Compose Manager) 的方法，您可以根据自己的需求选择最适合的方式。

## 🚀 一键在线安装 (推荐)

### 快速安装
```bash
curl -fsSL https://raw.githubusercontent.com/QuentinHsu/compman/main/install-online.sh | bash
```

### 使用 wget
```bash
wget -qO- https://raw.githubusercontent.com/QuentinHsu/compman/main/install-online.sh | bash
```

## 📦 高级安装脚本

如果您需要更多自定义选项，可以使用完整版安装脚本：

### 下载并运行
```bash
curl -fsSL -o install.sh https://raw.githubusercontent.com/QuentinHsu/compman/main/install.sh
chmod +x install.sh
./install.sh
```

### 安装选项

```bash
# 查看帮助
./install.sh --help

# 安装指定版本
./install.sh -v v1.2.0

# 安装到自定义目录
./install.sh -d ~/bin

# 强制重新安装
./install.sh -f

# 跳过校验和验证
./install.sh --no-verify

# 仅创建配置文件示例
./install.sh --config-only
```

## 🏗️ 手动安装

### 1. 从 Releases 页面下载

访问 [Releases 页面](https://github.com/QuentinHsu/compman/releases) 下载适合您系统的二进制文件。

### 2. 解压并安装

```bash
# 下载 (以 Linux x86_64 为例)
wget https://github.com/QuentinHsu/compman/releases/latest/download/compman-latest-linux-amd64.tar.gz

# 解压
tar -xzf compman-latest-linux-amd64.tar.gz

# 安装
sudo cp compman-latest-linux-amd64 /usr/local/bin/compman
sudo chmod +x /usr/local/bin/compman
```

### 3. 验证安装

```bash
compman --version
```

## 🔧 从源码构建

### 前置要求
- Go 1.21 或更高版本
- Git

### 构建步骤

```bash
# 克隆仓库
git clone https://github.com/QuentinHsu/compman.git
cd compman

# 安装依赖
go mod tidy

# 构建当前平台版本
go build -o compman cmd/main.go

# 或使用构建脚本构建所有平台
./build-advanced.sh
```

## 📋 支持的平台

| 操作系统 | 架构 | 二进制文件名 |
|---------|------|-------------|
| macOS | Intel (x86_64) | `compman-*-darwin-amd64` |
| macOS | Apple Silicon (ARM64) | `compman-*-darwin-arm64` |
| Linux | x86_64 | `compman-*-linux-amd64` |
| Linux | ARM64 | `compman-*-linux-arm64` |
| Windows | x86_64 | `compman-*-windows-amd64.exe` |

## ⚙️ 配置

安装完成后，Compman 会在 `~/.config/compman/` 目录创建配置文件示例。

### 配置文件位置
- **Linux/macOS**: `~/.config/compman/config.yaml`
- **Windows**: `%APPDATA%\compman\config.yaml`

### 基本配置示例

```yaml
# 全局设置
global:
  tag_strategy: "latest"    # 或 "semver"
  auto_cleanup: true
  interactive: true
  verbose: false

# Docker 设置
docker:
  timeout: 30

# 备份设置
backup:
  enabled: true
  dir: "~/.compman/backups"
  keep: 5
```

## 🚀 快速开始

```bash
# 查看帮助
compman --help

# 查看版本
compman --version

# 更新所有 Docker Compose 服务
compman update

# 交互式更新
compman update -i

# 使用语义化版本策略
compman update --strategy semver

# 更新指定的 Compose 文件
compman update docker-compose.yml

# 干运行（不实际执行）
compman update --dry-run
```

## ❓ 故障排除

### 权限问题
如果遇到权限错误，请确保：
1. 安装目录有写权限，或使用 `sudo`
2. Docker daemon 正在运行且当前用户有权限访问

### 网络问题
如果下载失败，可以：
1. 检查网络连接
2. 使用代理：`export https_proxy=http://your-proxy:port`
3. 手动下载二进制文件

### 版本问题
如果遇到版本兼容问题：
1. 更新到最新版本
2. 检查 Docker 和 Docker Compose 版本
3. 查看 [兼容性文档](https://github.com/QuentinHsu/compman#compatibility)

## 🔄 更新 Compman

```bash
# 重新运行安装脚本
curl -fsSL https://raw.githubusercontent.com/QuentinHsu/compman/main/install-online.sh | bash

# 或使用高级脚本强制更新
./install.sh -f
```

## 🗑️ 卸载

```bash
# 删除二进制文件
sudo rm /usr/local/bin/compman

# 删除配置文件 (可选)
rm -rf ~/.config/compman
```

## 📞 获取帮助

- 📚 [项目文档](https://github.com/QuentinHsu/compman)
- 🐛 [问题反馈](https://github.com/QuentinHsu/compman/issues)
- 💬 [讨论区](https://github.com/QuentinHsu/compman/discussions)

---

**注意**: 请将文档中的 `QuentinHsu` 替换为您的实际 GitHub 用户名。
