package strategy

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"compman/internal/docker"

	"github.com/Masterminds/semver/v3"
)

// SemverStrategy 语义版本策略实现
type SemverStrategy struct {
	pattern      string
	imageManager *docker.ImageManager
	constraint   *semver.Constraints
}

// NewSemverStrategy 创建新的语义版本策略
func NewSemverStrategy(pattern string) *SemverStrategy {
	if pattern == "" {
		pattern = "*" // 默认接受所有版本
	}

	// 解析约束条件
	constraint, err := semver.NewConstraint(pattern)
	if err != nil {
		// 如果解析失败，使用默认约束
		constraint, _ = semver.NewConstraint("*")
	}

	return &SemverStrategy{
		pattern:      pattern,
		imageManager: docker.NewImageManager(),
		constraint:   constraint,
	}
}

// GetLatestTag 获取符合语义版本规则的最新标签
func (s *SemverStrategy) GetLatestTag(image string) (string, error) {
	// 提取镜像名称
	imageName := s.extractImageName(image)

	// 获取所有可用标签
	tags, err := s.imageManager.GetImageTags(imageName)
	if err != nil {
		return "", fmt.Errorf("获取镜像标签失败: %v", err)
	}

	// 过滤和解析语义版本标签
	var validVersions []*semver.Version
	for _, tag := range tags {
		version, err := s.parseVersion(tag)
		if err != nil {
			continue // 跳过无效的版本标签
		}

		// 检查是否符合约束条件
		if s.constraint.Check(version) {
			validVersions = append(validVersions, version)
		}
	}

	if len(validVersions) == 0 {
		return "", fmt.Errorf("未找到符合条件的语义版本标签")
	}

	// 排序获取最新版本
	sort.Sort(semver.Collection(validVersions))
	latest := validVersions[len(validVersions)-1]

	return latest.Original(), nil
}

// ValidateTag 验证标签是否符合语义版本规范
func (s *SemverStrategy) ValidateTag(tag string) bool {
	version, err := s.parseVersion(tag)
	if err != nil {
		return false
	}

	return s.constraint.Check(version)
}

// parseVersion 解析版本字符串
func (s *SemverStrategy) parseVersion(tag string) (*semver.Version, error) {
	// 清理版本标签
	cleanTag := s.cleanVersionTag(tag)

	// 尝试解析
	version, err := semver.NewVersion(cleanTag)
	if err != nil {
		return nil, err
	}

	return version, nil
}

// cleanVersionTag 清理版本标签
func (s *SemverStrategy) cleanVersionTag(tag string) string {
	// 移除常见的版本前缀
	prefixes := []string{"v", "version", "ver", "release", "rel"}

	lowerTag := strings.ToLower(tag)
	for _, prefix := range prefixes {
		if strings.HasPrefix(lowerTag, prefix) {
			return tag[len(prefix):]
		}
	}

	return tag
}

// extractImageName 从完整镜像名称中提取不带标签的部分
func (s *SemverStrategy) extractImageName(image string) string {
	// 处理带有 @ 的镜像摘要格式
	if strings.Contains(image, "@") {
		parts := strings.Split(image, "@")
		return parts[0]
	}

	// 处理带有 : 的标签格式
	if strings.Contains(image, ":") {
		parts := strings.Split(image, ":")
		if len(parts) >= 2 {
			// 检查最后一部分是否包含端口号
			lastPart := parts[len(parts)-1]
			if s.isPort(lastPart) {
				return image
			}
			return strings.Join(parts[:len(parts)-1], ":")
		}
	}

	return image
}

// isPort 检查字符串是否是端口号
func (s *SemverStrategy) isPort(str string) bool {
	if len(str) < 1 || len(str) > 5 {
		return false
	}

	for _, char := range str {
		if char < '0' || char > '9' {
			return false
		}
	}

	port, err := strconv.Atoi(str)
	return err == nil && port > 0 && port <= 65535
}

// GetStrategyName 获取策略名称
func (s *SemverStrategy) GetStrategyName() string {
	return "semver"
}

// GetDescription 获取策略描述
func (s *SemverStrategy) GetDescription() string {
	return fmt.Sprintf("语义版本策略，约束条件: %s", s.pattern)
}

// CanHandle 检查该策略是否可以处理给定的镜像
func (s *SemverStrategy) CanHandle(image string) bool {
	// 尝试从当前镜像标签中提取版本信息
	tag := s.extractTag(image)
	_, err := s.parseVersion(tag)
	return err == nil
}

// extractTag 从镜像名称中提取标签
func (s *SemverStrategy) extractTag(image string) string {
	if strings.Contains(image, ":") {
		parts := strings.Split(image, ":")
		return parts[len(parts)-1]
	}
	return "latest"
}

// GetRecommendedTag 为镜像推荐标签
func (s *SemverStrategy) GetRecommendedTag(image string) (string, error) {
	return s.GetLatestTag(image)
}

// CompareVersions 比较两个版本
func (s *SemverStrategy) CompareVersions(version1, version2 string) int {
	v1, err1 := s.parseVersion(version1)
	v2, err2 := s.parseVersion(version2)

	if err1 != nil || err2 != nil {
		// 如果解析失败，按字符串比较
		if version1 == version2 {
			return 0
		} else if version1 < version2 {
			return -1
		} else {
			return 1
		}
	}

	return v1.Compare(v2)
}

// ShouldUpdate 检查是否应该更新镜像
func (s *SemverStrategy) ShouldUpdate(currentImage, targetImage string) bool {
	currentTag := s.extractTag(currentImage)
	targetTag := s.extractTag(targetImage)

	return s.CompareVersions(currentTag, targetTag) < 0
}

// GetVersionList 获取符合条件的版本列表
func (s *SemverStrategy) GetVersionList(image string, limit int) ([]*semver.Version, error) {
	imageName := s.extractImageName(image)

	tags, err := s.imageManager.GetImageTags(imageName)
	if err != nil {
		return nil, err
	}

	var versions []*semver.Version
	for _, tag := range tags {
		version, err := s.parseVersion(tag)
		if err != nil {
			continue
		}

		if s.constraint.Check(version) {
			versions = append(versions, version)
		}
	}

	// 排序
	sort.Sort(sort.Reverse(semver.Collection(versions)))

	// 限制数量
	if limit > 0 && len(versions) > limit {
		versions = versions[:limit]
	}

	return versions, nil
}

// SetConstraint 设置版本约束
func (s *SemverStrategy) SetConstraint(pattern string) error {
	constraint, err := semver.NewConstraint(pattern)
	if err != nil {
		return fmt.Errorf("无效的版本约束: %v", err)
	}

	s.pattern = pattern
	s.constraint = constraint
	return nil
}

// GetConstraint 获取当前版本约束
func (s *SemverStrategy) GetConstraint() string {
	return s.pattern
}
