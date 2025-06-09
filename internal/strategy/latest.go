package strategy

import (
	"fmt"
	"strings"

	"compman/internal/docker"
)

// LatestStrategy latest 标签策略实现
type LatestStrategy struct {
	imageManager *docker.ImageManager
}

// NewLatestStrategy 创建新的 latest 策略
func NewLatestStrategy() *LatestStrategy {
	return &LatestStrategy{
		imageManager: docker.NewImageManager(),
	}
}

// GetLatestTag 获取镜像的最新标签
func (s *LatestStrategy) GetLatestTag(image string) (string, error) {
	// 对于 latest 策略，总是返回 "latest"
	// 在实际应用中，可能需要检查远程仓库确认 latest 标签是否存在

	// 解析镜像名称，移除现有标签
	imageName := s.extractImageName(image)

	// 验证镜像是否存在
	exists, err := s.imageManager.ValidateImageExists(imageName)
	if err != nil {
		return "", fmt.Errorf("验证镜像 %s 时出错: %v", imageName, err)
	}

	if !exists {
		return "", fmt.Errorf("镜像 %s 不存在", imageName)
	}

	return "latest", nil
}

// ValidateTag 验证标签是否有效
func (s *LatestStrategy) ValidateTag(tag string) bool {
	// latest 策略只接受 latest 标签
	return strings.EqualFold(tag, "latest")
}

// extractImageName 从完整镜像名称中提取不带标签的部分
func (s *LatestStrategy) extractImageName(image string) string {
	// 处理带有 @ 的镜像摘要格式
	if strings.Contains(image, "@") {
		parts := strings.Split(image, "@")
		return parts[0]
	}

	// 处理带有 : 的标签格式
	if strings.Contains(image, ":") {
		parts := strings.Split(image, ":")
		if len(parts) >= 2 {
			// 检查最后一部分是否包含端口号（数字）
			lastPart := parts[len(parts)-1]
			if s.isPort(lastPart) {
				return image // 如果是端口号，返回原始字符串
			}
			// 否则移除标签部分
			return strings.Join(parts[:len(parts)-1], ":")
		}
	}

	return image
}

// isPort 检查字符串是否是端口号
func (s *LatestStrategy) isPort(str string) bool {
	// 简单检查：如果字符串是纯数字且长度合理，认为是端口号
	if len(str) < 1 || len(str) > 5 {
		return false
	}

	for _, char := range str {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}

// GetStrategyName 获取策略名称
func (s *LatestStrategy) GetStrategyName() string {
	return "latest"
}

// GetDescription 获取策略描述
func (s *LatestStrategy) GetDescription() string {
	return "始终使用 latest 标签，适用于开发环境或不需要版本控制的场景"
}

// CanHandle 检查该策略是否可以处理给定的镜像
func (s *LatestStrategy) CanHandle(image string) bool {
	// latest 策略可以处理任何镜像
	return true
}

// GetRecommendedTag 为镜像推荐标签
func (s *LatestStrategy) GetRecommendedTag(image string) (string, error) {
	return s.GetLatestTag(image)
}

// CompareVersions 比较两个版本
func (s *LatestStrategy) CompareVersions(version1, version2 string) int {
	// latest 策略下，所有版本都被认为是相等的
	return 0
}

// ShouldUpdate 检查是否应该更新镜像
func (s *LatestStrategy) ShouldUpdate(currentImage, targetImage string) bool {
	currentTag := s.extractTag(currentImage)
	targetTag := s.extractTag(targetImage)

	// 如果当前不是 latest，则应该更新
	return !strings.EqualFold(currentTag, "latest") || !strings.EqualFold(targetTag, "latest")
}

// extractTag 从镜像名称中提取标签
func (s *LatestStrategy) extractTag(image string) string {
	if strings.Contains(image, ":") {
		parts := strings.Split(image, ":")
		return parts[len(parts)-1]
	}
	return "latest"
}
