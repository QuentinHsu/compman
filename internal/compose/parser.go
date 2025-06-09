package compose

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"compman/pkg/types"

	"gopkg.in/yaml.v3"
)

// Parser 负责解析 Docker Compose 文件
type Parser struct {
	strict bool // 严格模式，遇到错误时停止
}

// NewParser 创建一个新的解析器
func NewParser() *Parser {
	return &Parser{
		strict: false,
	}
}

// SetStrict 设置严格模式
func (p *Parser) SetStrict(strict bool) {
	p.strict = strict
}

// ParseFile 解析 Docker Compose 文件
func (p *Parser) ParseFile(filePath string) (*types.ComposeFile, error) {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("文件不存在: %s", filePath)
	}

	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %v", err)
	}

	// 解析 YAML
	composeFile, err := p.ParseContent(content)
	if err != nil {
		return nil, fmt.Errorf("解析文件 %s 失败: %v", filePath, err)
	}

	// 设置文件路径
	composeFile.FilePath = filePath

	return composeFile, nil
}

// ParseContent 解析 YAML 内容
func (p *Parser) ParseContent(content []byte) (*types.ComposeFile, error) {
	var composeFile types.ComposeFile

	// 使用 yaml.v3 解析，支持更好的错误处理
	decoder := yaml.NewDecoder(strings.NewReader(string(content)))
	decoder.KnownFields(false) // 允许未知字段

	if err := decoder.Decode(&composeFile); err != nil {
		return nil, fmt.Errorf("YAML 解析失败: %v", err)
	}

	// 验证和规范化
	if err := p.validateAndNormalize(&composeFile); err != nil {
		if p.strict {
			return nil, err
		}
		// 非严格模式下，静默处理警告
	}

	return &composeFile, nil
}

// validateAndNormalize 验证和规范化 Compose 文件
func (p *Parser) validateAndNormalize(cf *types.ComposeFile) error {
	// 验证版本
	if cf.Version == "" {
		cf.Version = "3.8" // 默认版本
	}

	// 支持的版本列表
	supportedVersions := map[string]bool{
		"2":   true,
		"2.0": true,
		"2.1": true,
		"2.2": true,
		"2.3": true,
		"2.4": true,
		"3":   true,
		"3.0": true,
		"3.1": true,
		"3.2": true,
		"3.3": true,
		"3.4": true,
		"3.5": true,
		"3.6": true,
		"3.7": true,
		"3.8": true,
		"3.9": true,
	}

	if !supportedVersions[cf.Version] {
		return fmt.Errorf("不支持的 Compose 文件版本: %s", cf.Version)
	}

	// 验证服务
	if cf.Services == nil {
		cf.Services = make(map[string]types.Service)
	}

	// 规范化服务配置
	for serviceName, service := range cf.Services {
		if err := p.normalizeService(serviceName, &service); err != nil {
			return fmt.Errorf("服务 %s 配置错误: %v", serviceName, err)
		}
		cf.Services[serviceName] = service
	}

	return nil
}

// normalizeService 规范化服务配置
func (p *Parser) normalizeService(name string, service *types.Service) error {
	// 验证镜像或构建配置
	if service.Image == "" && service.Build == nil {
		return fmt.Errorf("服务必须指定 image 或 build")
	}

	// 规范化镜像名称
	if service.Image != "" {
		service.Image = p.normalizeImageName(service.Image)
	}

	// 规范化构建配置
	if service.Build != nil {
		if service.Build.Context == "" {
			service.Build.Context = "."
		}
		if service.Build.Dockerfile == "" {
			service.Build.Dockerfile = "Dockerfile"
		}
	}

	// 规范化重启策略
	if service.Restart == "" {
		service.Restart = "unless-stopped"
	}

	// 验证重启策略
	validRestartPolicies := map[string]bool{
		"no":             true,
		"always":         true,
		"on-failure":     true,
		"unless-stopped": true,
	}

	if !validRestartPolicies[service.Restart] {
		return fmt.Errorf("无效的重启策略: %s", service.Restart)
	}

	return nil
}

// normalizeImageName 规范化镜像名称
func (p *Parser) normalizeImageName(image string) string {
	// 如果没有指定标签，添加 :latest
	if !strings.Contains(image, ":") && !strings.Contains(image, "@") {
		return image + ":latest"
	}
	return image
}

// WriteFile 将 ComposeFile 写入文件
func (p *Parser) WriteFile(composeFile *types.ComposeFile, filePath string) error {
	// 创建目录
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// 序列化为 YAML
	content, err := p.Marshal(composeFile)
	if err != nil {
		return fmt.Errorf("序列化失败: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	return nil
}

// Marshal 将 ComposeFile 序列化为 YAML
func (p *Parser) Marshal(composeFile *types.ComposeFile) ([]byte, error) {
	return yaml.Marshal(composeFile)
}

// BackupFile 备份原始文件
func (p *Parser) BackupFile(filePath string) (string, error) {
	backupPath := filePath + ".backup." + fmt.Sprintf("%d", os.Getuid())

	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("读取原文件失败: %v", err)
	}

	if err := os.WriteFile(backupPath, content, 0644); err != nil {
		return "", fmt.Errorf("创建备份文件失败: %v", err)
	}

	return backupPath, nil
}

// RestoreFromBackup 从备份恢复文件
func (p *Parser) RestoreFromBackup(filePath, backupPath string) error {
	content, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("读取备份文件失败: %v", err)
	}

	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return fmt.Errorf("恢复文件失败: %v", err)
	}

	return nil
}

// ValidateFile 验证 Compose 文件的语法
func (p *Parser) ValidateFile(filePath string) error {
	_, err := p.ParseFile(filePath)
	return err
}

// GetImageList 获取 Compose 文件中的所有镜像
func (p *Parser) GetImageList(composeFile *types.ComposeFile) []string {
	var images []string

	for _, service := range composeFile.Services {
		if service.Image != "" {
			images = append(images, service.Image)
		}
	}

	return images
}

// GetServiceNames 获取所有服务名称
func (p *Parser) GetServiceNames(composeFile *types.ComposeFile) []string {
	var services []string

	for serviceName := range composeFile.Services {
		services = append(services, serviceName)
	}

	return services
}
