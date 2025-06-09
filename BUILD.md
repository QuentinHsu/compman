# 构建说明

本项目提供了多种构建方式来编译 `compman` 的多平台版本。

## 🚀 快速开始

### 1. 使用 Makefile (推荐)

```bash
# 查看所有可用的构建目标
make help

# 构建当前平台版本
make build

# 构建所有平台版本
make build-all

# 创建完整发布版本 (清理 + 多平台构建)
make release

# 清理构建文件
make clean
```

### 2. 使用基础构建脚本

```bash
# 运行基础构建脚本
./build.sh

# 设置自定义版本号
VERSION=v2.0.0 ./build.sh
```

### 3. 使用高级构建脚本

```bash
# 运行高级构建脚本 (支持压缩和校验和)
./build-advanced.sh

# 查看帮助信息
./build-advanced.sh --help

# 禁用压缩
./build-advanced.sh --no-compress

# 禁用校验和生成
./build-advanced.sh --no-checksum

# 禁用符号表删除
./build-advanced.sh --no-strip

# 组合使用
VERSION=v2.0.0 ./build-advanced.sh --no-compress
```

### 4. 单独构建特定平台

```bash
# 构建 macOS Intel 版本
make build-darwin-amd64

# 构建 macOS Apple Silicon 版本
make build-darwin-arm64

# 构建 Linux x86_64 版本
make build-linux-amd64

# 构建 Linux ARM64 版本
make build-linux-arm64
```

## 📦 支持的平台

| 平台 | 架构 | 输出文件名格式 |
|------|------|----------------|
| macOS | Intel (amd64) | `compman-{version}-darwin-amd64` |
| macOS | Apple Silicon (arm64) | `compman-{version}-darwin-arm64` |
| Linux | x86_64 (amd64) | `compman-{version}-linux-amd64` |
| Linux | ARM64 | `compman-{version}-linux-arm64` |

## 🔧 构建选项

### 环境变量

- `VERSION`: 设置版本号 (默认: git describe)
- `COMPRESS`: 启用/禁用压缩 (默认: true)
- `CHECKSUM`: 启用/禁用校验和生成 (默认: true)
- `STRIP`: 启用/禁用符号表删除 (默认: true)

### 示例

```bash
# 设置自定义版本号
VERSION=v1.2.3 make build-all

# 构建但不压缩
COMPRESS=false ./build-advanced.sh

# 保留调试符号
STRIP=false ./build-advanced.sh
```

## 📁 输出结构

构建完成后，文件结构如下：

```
dist/
├── compman-v1.0.0-darwin-amd64          # macOS Intel 二进制文件
├── compman-v1.0.0-darwin-arm64          # macOS ARM64 二进制文件  
├── compman-v1.0.0-linux-amd64           # Linux x86_64 二进制文件
├── compman-v1.0.0-linux-arm64           # Linux ARM64 二进制文件
├── checksums.txt                        # 二进制文件校验和
└── archives/                            # 压缩包目录 (仅高级构建脚本)
    ├── compman-v1.0.0-darwin-amd64.tar.gz
    ├── compman-v1.0.0-darwin-arm64.tar.gz
    ├── compman-v1.0.0-linux-amd64.tar.gz
    ├── compman-v1.0.0-linux-arm64.tar.gz
    └── checksums.txt                    # 压缩包校验和
```

## 🛠️ 开发工作流

### 日常开发

```bash
# 安装依赖
make deps

# 格式化代码
make fmt

# 运行测试
make test

# 开发模式运行
make dev

# 构建当前平台版本进行测试
make build
```

### 发布版本

```bash
# 创建发布版本 (推荐)
make release

# 或者使用高级脚本
VERSION=v1.2.3 ./build-advanced.sh
```

## 🔍 版本信息

构建后的二进制文件会包含以下版本信息：

- 版本号 (通过 git describe 自动生成或手动指定)
- 构建时间
- Git 提交哈希 (如果可用)

查看版本信息：

```bash
./dist/compman-v1.0.0-darwin-amd64 --version
```

## ⚡ 性能优化

### 构建标志

- `-s -w`: 删除符号表和调试信息，减小文件大小
- `-ldflags`: 设置链接标志，注入版本信息

### 并行构建

所有构建脚本都支持并行编译多个平台，大大提高构建速度。

## 🐛 故障排除

### 常见问题

1. **权限错误**: 确保构建脚本有执行权限
   ```bash
   chmod +x build.sh build-advanced.sh
   ```

2. **Go 环境**: 确保 Go 已正确安装并在 PATH 中
   ```bash
   go version
   ```

3. **Git 信息**: 如果不在 Git 仓库中，版本号会默认为 "dev"

4. **磁盘空间**: 确保有足够的磁盘空间存储所有平台的二进制文件

## 🚀 CI/CD 集成

这些构建脚本可以轻松集成到 CI/CD 管道中：

```yaml
# GitHub Actions 示例
- name: Build all platforms
  run: make release

# GitLab CI 示例  
build:
  script:
    - ./build-advanced.sh
  artifacts:
    paths:
      - dist/
```

## 📝 注意事项

1. 构建前会自动清理 `dist` 目录
2. 版本号优先使用环境变量 `VERSION`，其次使用 git describe
3. 高级构建脚本提供更多选项但构建时间稍长
4. 建议在发布时使用 `make release` 或 `./build-advanced.sh`
