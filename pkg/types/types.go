package types

import "time"

// ComposeFile represents a Docker Compose file structure
type ComposeFile struct {
	Version  string                 `yaml:"version"`
	Services map[string]Service     `yaml:"services"`
	Networks map[string]interface{} `yaml:"networks,omitempty"`
	Volumes  map[string]interface{} `yaml:"volumes,omitempty"`
	FilePath string                 `yaml:"-"` // 文件路径，不序列化
}

// Service represents a service in Docker Compose
type Service struct {
	Image       string                 `yaml:"image,omitempty"`
	Build       *BuildConfig           `yaml:"build,omitempty"`
	Environment interface{}            `yaml:"environment,omitempty"` // 可以是 []string 或 map[string]string
	Ports       []string               `yaml:"ports,omitempty"`
	Volumes     []string               `yaml:"volumes,omitempty"`
	DependsOn   []string               `yaml:"depends_on,omitempty"`
	Networks    []string               `yaml:"networks,omitempty"`
	Restart     string                 `yaml:"restart,omitempty"`
	ExtraHosts  []string               `yaml:"extra_hosts,omitempty"`
	Command     interface{}            `yaml:"command,omitempty"`
	Labels      map[string]string      `yaml:"labels,omitempty"`
	Other       map[string]interface{} `yaml:",inline"` // 捕获其他字段
}

// BuildConfig represents build configuration
type BuildConfig struct {
	Context    string            `yaml:"context"`
	Dockerfile string            `yaml:"dockerfile,omitempty"`
	Args       map[string]string `yaml:"args,omitempty"`
	Target     string            `yaml:"target,omitempty"`
}

// Config represents application configuration
type Config struct {
	ComposePaths     []string            `yaml:"compose_paths"`      // Compose 文件搜索路径
	ImageTagStrategy string              `yaml:"image_tag_strategy"` // 镜像标签策略 (latest, semver)
	Environment      string              `yaml:"environment"`        // 环境 (dev, prod, etc.)
	SemverPattern    string              `yaml:"semver_pattern"`     // Semver 匹配模式
	ExcludeImages    []string            `yaml:"exclude_images"`     // 排除的镜像
	DryRun           bool                `yaml:"dry_run"`            // 干运行模式
	BackupEnabled    bool                `yaml:"backup_enabled"`     // 是否备份原文件
	Timeout          time.Duration       `yaml:"timeout"`            // 操作超时时间
	DockerConfig     DockerConfig        `yaml:"docker_config"`      // Docker 配置
	SelectedServices map[string][]string `yaml:"-"`                  // 选中的服务 (文件路径 -> 服务名列表)
}

// DockerConfig represents Docker client configuration
type DockerConfig struct {
	Host       string `yaml:"host"`        // Docker daemon 地址
	APIVersion string `yaml:"api_version"` // API 版本
	TLSVerify  bool   `yaml:"tls_verify"`  // TLS 验证
	CertPath   string `yaml:"cert_path"`   // 证书路径
}

// ImageTagStrategy defines interface for image tag strategies
type ImageTagStrategy interface {
	GetLatestTag(image string) (string, error)
	ValidateTag(tag string) bool
}

// ComposeManager represents the main manager interface
type ComposeManager interface {
	ScanComposeFiles(paths []string) ([]*ComposeFile, error)
	UpdateImages(files []*ComposeFile) error
	CleanupImages() error
}

// ComposeFileInfo contains information about a compose file
type ComposeFileInfo struct {
	Path         string
	Name         string
	LastModified time.Time
	Services     []string
}

// ImageInfo contains information about a Docker image
type ImageInfo struct {
	Repository string
	Tag        string
	ImageID    string
	Created    time.Time
	Size       int64
	InUse      bool
}

// UpdateResult represents the result of an update operation
type UpdateResult struct {
	Service   string
	OldImage  string
	NewImage  string
	Success   bool
	Error     error
	UpdatedAt time.Time
}

// ScanResult represents the result of scanning compose files
type ScanResult struct {
	TotalFiles   int
	ValidFiles   int
	InvalidFiles []string
	Services     map[string][]string // service name -> compose files
}
