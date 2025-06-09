# Docker Compose Manager

Docker Compose Manager 是一个功能丰富的 Go 命令行工具，专为高效管理 Docker Compose 项目而设计。该工具提供智能化的镜像升级、自动化清理以及多环境支持等企业级功能。

## ✨ 功能特性

### 🚀 核心功能
- **智能镜像升级**：支持 `latest` 和 `semver`（语义化版本）升级策略
- **自动化清理**：升级后自动清理未使用的过期镜像
- **批量处理**：一次性处理多个 Docker Compose 文件
- **安全备份**：升级前自动创建配置文件备份

### 🎯 高级特性
- **多环境支持**：支持开发、测试、生产等多环境配置
- **Docker Compose v2 兼容**：完全支持最新的 Docker Compose 规范
- **智能文件扫描**：自动发现和处理 Compose 文件（支持 1Panel 等管理面板）
- **交互式选择**：可视化选择要更新的文件和服务，支持批量选择
- **灵活配置**：支持 YAML 配置文件和命令行参数
- **丰富的 UI**：彩色输出、进度条、表格显示和交互式确认

### 🔧 技术特性
- **Docker Hub 集成**：实时获取最新镜像标签信息
- **语义化版本**：智能处理和比较 semver 版本
- **错误恢复**：升级失败时自动恢复原始配置
- **并发处理**：支持并发处理多个文件以提高性能

## 📦 安装

### 前置要求
- Go 1.21 或更高版本
- Docker 和 Docker Compose

### 安装步骤

1. 克隆项目：
   ```bash
   git clone https://github.com/yourusername/docker-compose-manager.git
   cd docker-compose-manager
   ```

2. 安装依赖：
   ```bash
   go mod tidy
   ```

3. 构建二进制文件：
   
   - 当前系统环境所需：
   ```bash
   go build -o compman cmd/main.go
   ```

   - Linux amd64
   ```
   GOOS=linux GOARCH=amd64 go build -o compman cmd/main.go
   ```

4. （可选）安装到系统路径：
   ```bash
   sudo cp compman /usr/local/bin/
   ```

## 🚀 使用指南

### 基本用法

```bash
# 显示帮助信息
./compman --help

# 扫描当前目录的 Compose 文件
./compman scan

# 更新指定文件中的所有镜像
./compman update -f docker-compose.yml

# 使用 semver 策略更新镜像
./compman update -f docker-compose.yml --strategy semver

# 更新后清理未使用的镜像
./compman clean

# 交互式更新（推荐）
./compman update -f docker-compose.yml --interactive
```

### 命令详解

#### `scan` - 扫描 Compose 文件
```bash
# 扫描当前目录
./compman scan

# 扫描指定目录
./compman scan --path /path/to/compose/files

# 递归扫描（限制深度）
./compman scan --path /path --depth 3
```

#### `update` - 更新镜像
```bash
# 基本更新
./compman update -f docker-compose.yml

# 指定策略
./compman update -f docker-compose.yml --strategy latest
./compman update -f docker-compose.yml --strategy semver

# 批量更新多个文件
./compman update -f file1.yml -f file2.yml

# 交互式模式（推荐）- 可选择特定文件和服务
./compman update --paths /opt/1panel/docker/compose --interactive

# 包含特定服务
./compman update -f docker-compose.yml --include web,db

# 排除特定服务
./compman update -f docker-compose.yml --exclude cache
```

#### `clean` - 清理镜像
```bash
# 清理未使用的镜像
./compman clean

# 强制清理（不询问确认）
./compman clean --force

# 只显示将要清理的镜像
./compman clean --dry-run
```

### 🎯 交互式功能

交互式模式是推荐的使用方式，它提供了可视化的选择界面：

```bash
# 启用交互式模式
./compman update --paths /opt/1panel/docker/compose --interactive
```

**交互式流程：**

1. **文件选择**：显示所有发现的 Compose 文件，可以选择要处理的文件
2. **服务选择**：对于每个选中的文件，可以选择要更新的特定服务
3. **确认操作**：显示将要执行的操作摘要，确认后执行

**交互操作说明：**
- 输入数字切换选择状态（如：`1,3,5` 或 `1-3`）
- 输入 `a` 全选所有项目
- 输入 `n` 取消所有选择
- 按 Enter 确认当前选择
- 输入 `q` 退出

### 配置文件

从 v1.1 开始，compman 使用 `~/.config/compman/config.yml` 作为默认配置文件路径。

#### 配置管理

```bash
# 查看当前配置文件路径和内容
./compman config

# 仅显示配置文件路径
./compman config --path-only

# 使用指定配置文件（内容会合并到默认配置）
./compman update --config my-config.yml

# 显示使用指定配置后的合并结果
./compman config --config my-config.yml
```

#### 配置文件示例

默认配置文件会在首次运行时自动创建，内容如下：

```yaml
# ~/.config/compman/config.yml
compose_paths:
  - "./docker-compose.yml"
  - "./compose.yml"
image_tag_strategy: "latest"
environment: "production"
semver_pattern: "^v?\\d+\\.\\d+\\.\\d+$"
exclude_images: []
dry_run: false
backup_enabled: true
timeout: "5m"
docker_config:
  host: ""
  api_version: ""
  tls_verify: false
  cert_path: ""
```

#### 自定义配置

当你使用 `--config` 参数指定配置文件时，该文件的内容会合并到默认配置中，用户配置的值会覆盖默认值，然后保存到默认配置文件。这样确保了：

1. **一致性**：所有配置都存储在一个标准位置
2. **便利性**：不需要每次都指定配置文件
3. **灵活性**：可以随时更新配置而不影响工具使用

创建自定义配置文件：

```yaml
# custom-config.yml
compose_paths:
  - "/opt/1panel/docker/compose"
  - "/home/user/projects"
image_tag_strategy: "semver"
environment: "development"
exclude_images:
  - "nginx:alpine"
  - "redis:latest"
backup_enabled: true
timeout: "10m"
docker_config:
  host: "tcp://remote-docker:2376"
  tls_verify: true
  cert_path: "/path/to/certs"
```

然后应用配置：
```bash
# 将自定义配置合并到默认配置
./compman config --config custom-config.yml

# 之后所有命令都会使用合并后的配置
./compman update
./compman scan
```

## 📋 配置选项详解

### 基本配置
```yaml
# 基本配置示例
compose_paths:
  - "./docker-compose.yml"
  - "./compose.yml"
image_tag_strategy: "latest"  # latest 或 semver
environment: "production"     # 环境标识
backup_enabled: true          # 是否备份原文件
timeout: "5m"                # 操作超时时间
```

### 高级配置
```yaml
# 高级配置示例
compose_paths:
  - "/opt/1panel/docker/compose"
  - "/home/projects/*/docker-compose.yml"
image_tag_strategy: "semver"
environment: "production"
semver_pattern: "^v?\\d+\\.\\d+\\.\\d+$"  # semver 匹配模式
exclude_images:               # 排除更新的镜像
  - "postgres:*"              # 排除所有 postgres 镜像
  - "nginx:alpine"            # 排除特定标签
dry_run: false                # 是否为干运行模式
backup_enabled: true
timeout: "10m"
docker_config:
  host: ""                    # Docker 主机地址
  api_version: ""             # Docker API 版本
  tls_verify: false           # 是否启用 TLS 验证
  cert_path: ""               # 证书路径
```

### 配置选项说明

| 选项 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `compose_paths` | []string | `["./docker-compose.yml", "./compose.yml"]` | Compose 文件搜索路径 |
| `image_tag_strategy` | string | `"latest"` | 镜像标签升级策略：`latest` 或 `semver` |
| `environment` | string | `"production"` | 环境标识，用于日志和标记 |
| `semver_pattern` | string | `"^v?\\d+\\.d+\\.\\d+$"` | semver 策略的版本匹配模式 |
| `exclude_images` | []string | `[]` | 排除更新的镜像列表，支持通配符 |
| `dry_run` | bool | `false` | 干运行模式，不执行实际更新 |
| `backup_enabled` | bool | `true` | 是否在更新前备份原文件 |
| `timeout` | duration | `"5m"` | 操作超时时间 |
| `docker_config.host` | string | `""` | Docker 守护进程地址 |
| `docker_config.api_version` | string | `""` | Docker API 版本 |
| `docker_config.tls_verify` | bool | `false` | 是否启用 TLS 验证 |
| `docker_config.cert_path` | string | `""` | TLS 证书路径 |

## 🎨 输出示例

工具提供丰富的彩色输出和进度显示：

```
🔍 扫描 Docker Compose 文件...
✅ 找到 3 个文件

📋 发现的服务镜像:
┌─────────────┬─────────────────┬─────────────┬──────────────┐
│ 服务        │ 当前镜像        │ 当前版本    │ 最新版本     │
├─────────────┼─────────────────┼─────────────┼──────────────┤
│ web         │ nginx           │ 1.20        │ 1.21         │
│ api         │ node            │ 16-alpine   │ 18-alpine    │
│ database    │ postgres        │ 13          │ 15           │
└─────────────┴─────────────────┴─────────────┴──────────────┘

🚀 开始更新镜像...
[████████████████████████████████████████] 100% (3/3) 完成

✅ 更新完成！
- 成功更新: 3 个镜像
- 跳过: 0 个镜像
- 失败: 0 个镜像
```

## 🔧 开发

### 项目结构
```
docker-compose-manager/
├── cmd/
│   └── main.go              # 应用程序入口点
├── internal/
│   ├── config/             # 配置管理
│   ├── compose/            # Compose 文件处理
│   ├── docker/             # Docker 客户端封装
│   ├── strategy/           # 更新策略实现
│   └── ui/                 # 用户界面
├── pkg/
│   └── types/              # 共享类型定义
├── config.example.yaml     # 配置示例
├── go.mod                  # Go 模块定义
└── README.md              # 项目文档
```

### 构建和测试
```bash
# 运行测试
go test ./...

# 构建
go build -o compman cmd/main.go

# 运行开发版本
go run cmd/main.go --help
```

## 🤝 贡献

欢迎贡献代码！请遵循以下步骤：

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 📄 许可证

本项目采用 MIT 许可证。详情请见 [LICENSE](LICENSE) 文件。

## 🙏 致谢

感谢以下开源项目：
- [Cobra](https://github.com/spf13/cobra) - 强大的 CLI 框架
- [Viper](https://github.com/spf13/viper) - 配置管理
- [Docker Go SDK](https://github.com/docker/docker) - Docker API 客户端
- [Go-Version](https://github.com/hashicorp/go-version) - 语义化版本处理

该项目遵循 MIT 许可证。有关详细信息，请参阅 LICENSE 文件。