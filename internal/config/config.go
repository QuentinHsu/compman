package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"compman/pkg/types"

	"github.com/spf13/viper"
)

var (
	configFile string
	config     *types.Config
)

// getDefaultConfigPath returns the default configuration file path
func getDefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "config.yml"
	}
	configDir := filepath.Join(home, ".config", "compman")
	return filepath.Join(configDir, "config.yml")
}

// ensureConfigDir creates the configuration directory if it doesn't exist
func ensureConfigDir() error {
	defaultPath := getDefaultConfigPath()
	configDir := filepath.Dir(defaultPath)
	return os.MkdirAll(configDir, 0755)
}

// SetConfigFile sets the configuration file path
func SetConfigFile(file string) {
	configFile = file
	viper.SetConfigFile(file)
}

// SetConfigPath sets the configuration search path
func SetConfigPath(path string) {
	viper.AddConfigPath(path)
}

// SetConfigName sets the configuration file name (without extension)
func SetConfigName(name string) {
	viper.SetConfigName(name)
	viper.SetConfigType("yaml")
}

// LoadConfig loads configuration from file or creates default config
func LoadConfig() (*types.Config, error) {
	if config != nil {
		return config, nil
	}

	defaultPath := getDefaultConfigPath()

	// 如果用户指定了不同的配置文件，加载并合并到默认配置
	if configFile != "" && configFile != defaultPath {
		// 读取用户配置文件
		userCfg, err := loadConfigFromFile(configFile)
		if err != nil {
			return nil, fmt.Errorf("加载用户配置文件失败: %v", err)
		}

		// 合并配置：用户配置优先，缺失的使用系统默认配置
		systemDefaultCfg := getDefaultConfig()
		config = mergeConfigs(systemDefaultCfg, userCfg)

		// 将合并后的配置保存到默认位置
		if err := SaveConfigToDefault(config); err != nil {
			return nil, fmt.Errorf("保存配置到默认位置失败: %v", err)
		}
	} else {
		// 尝试加载默认配置文件
		if _, err := os.Stat(defaultPath); err == nil {
			config, err = loadConfigFromFile(defaultPath)
			if err != nil {
				return nil, fmt.Errorf("加载默认配置文件失败: %v", err)
			}
		} else {
			// 配置文件不存在，使用默认配置
			config = getDefaultConfig()
			// 创建默认配置文件
			if err := SaveConfigToDefault(config); err != nil {
				return nil, fmt.Errorf("创建默认配置文件失败: %v", err)
			}
		}
	}

	// 验证配置
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("配置验证失败: %v", err)
	}

	return config, nil
}

// loadConfigFromFile loads configuration from a specific file
func loadConfigFromFile(filePath string) (*types.Config, error) {
	v := viper.New()
	v.SetConfigFile(filePath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	cfg := &types.Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}

	// 手动设置如果 Unmarshal 失败或为空
	if len(cfg.ComposePaths) == 0 {
		cfg.ComposePaths = v.GetStringSlice("compose_paths")
	}
	if cfg.ImageTagStrategy == "" {
		cfg.ImageTagStrategy = v.GetString("image_tag_strategy")
	}
	if cfg.Environment == "" {
		cfg.Environment = v.GetString("environment")
	}
	if len(cfg.ExcludeImages) == 0 {
		cfg.ExcludeImages = v.GetStringSlice("exclude_images")
	}
	if cfg.SemverPattern == "" {
		cfg.SemverPattern = v.GetString("semver_pattern")
	}
	// 布尔值总是需要手动设置
	cfg.BackupEnabled = v.GetBool("backup_enabled")
	cfg.DryRun = v.GetBool("dry_run")

	// 处理持续时间
	if cfg.Timeout == 0 {
		if timeoutStr := v.GetString("timeout"); timeoutStr != "" {
			if duration, err := time.ParseDuration(timeoutStr); err == nil {
				cfg.Timeout = duration
			}
		}
	}

	return cfg, nil
}

// SaveConfig saves the current configuration to file
func SaveConfig(cfg *types.Config) error {
	if configFile == "" {
		configFile = getDefaultConfigPath()
	}

	// 确保配置目录存在
	if err := ensureConfigDir(); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	// 设置配置值
	viper.Set("compose_paths", cfg.ComposePaths)
	viper.Set("image_tag_strategy", cfg.ImageTagStrategy)
	viper.Set("environment", cfg.Environment)
	viper.Set("semver_pattern", cfg.SemverPattern)
	viper.Set("exclude_images", cfg.ExcludeImages)
	viper.Set("dry_run", cfg.DryRun)
	viper.Set("backup_enabled", cfg.BackupEnabled)
	viper.Set("timeout", cfg.Timeout)
	viper.Set("docker_config", cfg.DockerConfig)

	// 写入文件
	if err := viper.WriteConfigAs(configFile); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	config = cfg
	return nil
}

// SaveConfigToDefault saves configuration to the default config file
func SaveConfigToDefault(cfg *types.Config) error {
	defaultPath := getDefaultConfigPath()

	// 确保配置目录存在
	if err := ensureConfigDir(); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	// 创建新的 viper 实例来避免冲突
	v := viper.New()
	v.SetConfigFile(defaultPath)
	v.SetConfigType("yaml")

	// 设置配置值
	v.Set("compose_paths", cfg.ComposePaths)
	v.Set("image_tag_strategy", cfg.ImageTagStrategy)
	v.Set("environment", cfg.Environment)
	v.Set("semver_pattern", cfg.SemverPattern)
	v.Set("exclude_images", cfg.ExcludeImages)
	v.Set("dry_run", cfg.DryRun)
	v.Set("backup_enabled", cfg.BackupEnabled)
	v.Set("timeout", cfg.Timeout)
	v.Set("docker_config", cfg.DockerConfig)

	// 写入文件
	if err := v.WriteConfig(); err != nil {
		return fmt.Errorf("写入默认配置文件失败: %v", err)
	}

	return nil
}

// mergeConfigs merges user config with default config, user config takes priority
func mergeConfigs(defaultCfg, userCfg *types.Config) *types.Config {
	if defaultCfg == nil {
		defaultCfg = getDefaultConfig()
	}
	if userCfg == nil {
		return defaultCfg
	}

	merged := *defaultCfg

	// 用户配置优先，只有当用户配置有值时才覆盖默认配置
	if len(userCfg.ComposePaths) > 0 {
		merged.ComposePaths = userCfg.ComposePaths
	}
	if userCfg.ImageTagStrategy != "" {
		merged.ImageTagStrategy = userCfg.ImageTagStrategy
	}
	if userCfg.Environment != "" {
		merged.Environment = userCfg.Environment
	}
	if userCfg.SemverPattern != "" {
		merged.SemverPattern = userCfg.SemverPattern
	}
	if len(userCfg.ExcludeImages) > 0 {
		merged.ExcludeImages = userCfg.ExcludeImages
	}

	// 对于布尔值，检查是否与默认值不同
	if userCfg.DryRun != defaultCfg.DryRun {
		merged.DryRun = userCfg.DryRun
	}
	if userCfg.BackupEnabled != defaultCfg.BackupEnabled {
		merged.BackupEnabled = userCfg.BackupEnabled
	}

	if userCfg.Timeout > 0 {
		merged.Timeout = userCfg.Timeout
	}

	// Docker 配置合并
	if userCfg.DockerConfig.Host != "" {
		merged.DockerConfig.Host = userCfg.DockerConfig.Host
	}
	if userCfg.DockerConfig.APIVersion != "" {
		merged.DockerConfig.APIVersion = userCfg.DockerConfig.APIVersion
	}
	if userCfg.DockerConfig.CertPath != "" {
		merged.DockerConfig.CertPath = userCfg.DockerConfig.CertPath
	}
	if userCfg.DockerConfig.TLSVerify != defaultCfg.DockerConfig.TLSVerify {
		merged.DockerConfig.TLSVerify = userCfg.DockerConfig.TLSVerify
	}

	return &merged
}

// GenerateDefaultConfig creates a default configuration file
func GenerateDefaultConfig(path string) error {
	cfg := getDefaultConfig()

	// 如果提供了路径，使用它作为配置文件路径
	if path != "" {
		configFile = path
	}

	return SaveConfig(cfg)
}

// setDefaults sets default configuration values
func setDefaults() {
	viper.SetDefault("compose_paths", []string{"./docker-compose.yml", "./compose.yml"})
	viper.SetDefault("image_tag_strategy", "latest")
	viper.SetDefault("environment", "production")
	viper.SetDefault("semver_pattern", "^v?\\d+\\.\\d+\\.\\d+$")
	viper.SetDefault("exclude_images", []string{})
	viper.SetDefault("dry_run", false)
	viper.SetDefault("backup_enabled", true)
	viper.SetDefault("timeout", "5m")

	// Docker configuration defaults
	viper.SetDefault("docker_config.host", "")
	viper.SetDefault("docker_config.api_version", "")
	viper.SetDefault("docker_config.tls_verify", false)
	viper.SetDefault("docker_config.cert_path", "")
}

// getDefaultConfig returns a default configuration
func getDefaultConfig() *types.Config {
	return &types.Config{
		ComposePaths:     []string{"./docker-compose.yml", "./compose.yml"},
		ImageTagStrategy: "latest",
		Environment:      "production",
		SemverPattern:    "^v?\\d+\\.\\d+\\.\\d+$",
		ExcludeImages:    []string{},
		DryRun:           false,
		BackupEnabled:    true,
		Timeout:          5 * time.Minute,
		DockerConfig: types.DockerConfig{
			Host:       "",
			APIVersion: "",
			TLSVerify:  false,
			CertPath:   "",
		},
	}
}

// validateConfig validates the configuration
func validateConfig(cfg *types.Config) error {
	if len(cfg.ComposePaths) == 0 {
		return fmt.Errorf("至少需要指定一个 compose 文件路径")
	}

	validStrategies := map[string]bool{
		"latest": true,
		"semver": true,
	}

	if !validStrategies[cfg.ImageTagStrategy] {
		return fmt.Errorf("无效的镜像标签策略: %s (支持: latest, semver)", cfg.ImageTagStrategy)
	}

	if cfg.Timeout <= 0 {
		cfg.Timeout = 5 * time.Minute
	}

	return nil
}

// GetConfig returns the current configuration
func GetConfig() *types.Config {
	if config == nil {
		config, _ = LoadConfig()
	}
	return config
}

// ReloadConfig reloads the configuration from file
func ReloadConfig() error {
	config = nil
	_, err := LoadConfig()
	return err
}
