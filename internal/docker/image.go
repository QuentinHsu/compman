package docker

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"compman/pkg/types"

	"github.com/Masterminds/semver/v3"
)

// ImageManager 镜像管理器
type ImageManager struct {
	client     *Client
	httpClient *http.Client
}

// NewImageManager 创建新的镜像管理器
func NewImageManager() *ImageManager {
	return &ImageManager{
		client: NewClient(),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewImageManagerWithClient 使用指定客户端创建镜像管理器
func NewImageManagerWithClient(client *Client) *ImageManager {
	return &ImageManager{
		client: client,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetLatestTag 获取镜像的最新标签
func (im *ImageManager) GetLatestTag(imageName string, strategy string) (string, error) {
	switch strategy {
	case "semver":
		return im.getLatestSemverTag(imageName)
	case "latest":
		return im.getLatestTag(imageName)
	default:
		return im.getLatestTag(imageName)
	}
}

// getLatestTag 获取 latest 标签
func (im *ImageManager) getLatestTag(imageName string) (string, error) {
	// 对于 latest 策略，直接返回 latest
	// 实际实现中可能需要检查远程仓库
	return "latest", nil
}

// getLatestSemverTag 获取最新的语义版本标签
func (im *ImageManager) getLatestSemverTag(imageName string) (string, error) {
	tags, err := im.GetImageTags(imageName)
	if err != nil {
		return "", fmt.Errorf("获取镜像标签失败: %v", err)
	}

	var semverTags []*semver.Version

	// 过滤和解析语义版本标签
	for _, tag := range tags {
		// 清理标签前缀 (v, version, etc.)
		cleanTag := cleanVersionTag(tag)

		version, err := semver.NewVersion(cleanTag)
		if err == nil {
			semverTags = append(semverTags, version)
		}
	}

	if len(semverTags) == 0 {
		return "", fmt.Errorf("未找到有效的语义版本标签")
	}

	// 排序获取最新版本
	sort.Sort(semver.Collection(semverTags))
	latest := semverTags[len(semverTags)-1]

	return latest.String(), nil
}

// GetImageTags 从 Docker Hub 或其他镜像仓库获取标签列表
func (im *ImageManager) GetImageTags(imageName string) ([]string, error) {
	// 解析镜像名称
	registry, repository := im.parseImageName(imageName)

	switch registry {
	case "docker.io", "":
		return im.getDockerHubTags(repository)
	default:
		return im.getRegistryTags(registry, repository)
	}
}

// parseImageName 解析镜像名称
func (im *ImageManager) parseImageName(imageName string) (registry, repository string) {
	// 先移除标签部分（如果有的话）
	if strings.Contains(imageName, ":") {
		imageName = strings.Split(imageName, ":")[0]
	}

	parts := strings.Split(imageName, "/")

	if len(parts) == 1 {
		// 官方镜像，如 nginx
		return "docker.io", "library/" + parts[0]
	} else if len(parts) == 2 && !strings.Contains(parts[0], ".") {
		// 用户镜像，如 user/repo (但第一部分不是域名)
		return "docker.io", parts[0] + "/" + parts[1]
	} else {
		// 自定义镜像仓库，如 registry.com/user/repo
		return parts[0], strings.Join(parts[1:], "/")
	}
}

// getDockerHubTags 从 Docker Hub 获取标签
func (im *ImageManager) getDockerHubTags(repository string) ([]string, error) {
	url := fmt.Sprintf("https://registry.hub.docker.com/v2/repositories/%s/tags/?page_size=100", repository)

	resp, err := im.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求 Docker Hub API 失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// 读取错误响应体以获取更详细的错误信息
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Docker Hub API 响应错误: %d - %s\nURL: %s\nRepository: %s",
			resp.StatusCode, string(body), url, repository)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var response DockerHubTagsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v\n响应内容: %s", err, string(body))
	}

	var tags []string
	for _, result := range response.Results {
		tags = append(tags, result.Name)
	}

	// 如果没有找到标签，返回默认的 latest 标签
	if len(tags) == 0 {
		tags = append(tags, "latest")
	}

	return tags, nil
}

// getRegistryTags 从自定义镜像仓库获取标签
func (im *ImageManager) getRegistryTags(registry, repository string) ([]string, error) {
	// 实现自定义镜像仓库的标签获取逻辑
	// 这里返回一个基本的实现
	return []string{"latest"}, nil
}

// DockerHubTagsResponse Docker Hub API 响应结构
type DockerHubTagsResponse struct {
	Results []struct {
		Name   string `json:"name"`
		Images []struct {
			Architecture string `json:"architecture"`
			Features     string `json:"features"`
		} `json:"images"`
	} `json:"results"`
	Next string `json:"next"`
}

// cleanVersionTag 清理版本标签
func cleanVersionTag(tag string) string {
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

// ValidateImageExists 验证镜像是否存在
func (im *ImageManager) ValidateImageExists(imageName string) (bool, error) {
	tags, err := im.GetImageTags(imageName)
	if err != nil {
		return false, err
	}

	return len(tags) > 0, nil
}

// CompareImages 比较两个镜像版本
func (im *ImageManager) CompareImages(image1, image2 string) (int, error) {
	// 提取版本号
	tag1 := im.extractTag(image1)
	tag2 := im.extractTag(image2)

	// 尝试解析为语义版本
	v1, err1 := semver.NewVersion(cleanVersionTag(tag1))
	v2, err2 := semver.NewVersion(cleanVersionTag(tag2))

	if err1 == nil && err2 == nil {
		return v1.Compare(v2), nil
	}

	// 如果不是语义版本，按字符串比较
	if tag1 == tag2 {
		return 0, nil
	} else if tag1 < tag2 {
		return -1, nil
	} else {
		return 1, nil
	}
}

// extractTag 从镜像名称中提取标签
func (im *ImageManager) extractTag(imageName string) string {
	parts := strings.Split(imageName, ":")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return "latest"
}

// GetImageHistory 获取镜像历史
func (im *ImageManager) GetImageHistory(imageName string) ([]*types.ImageInfo, error) {
	if err := im.client.Connect(); err != nil {
		return nil, err
	}

	// 获取所有标签
	tags, err := im.GetImageTags(imageName)
	if err != nil {
		return nil, err
	}

	var history []*types.ImageInfo

	for _, tag := range tags {
		fullImageName := imageName + ":" + tag
		imageInfo, err := im.client.GetImageInfo(fullImageName)
		if err == nil {
			history = append(history, imageInfo)
		}
	}

	// 按创建时间排序
	sort.Slice(history, func(i, j int) bool {
		return history[i].Created.After(history[j].Created)
	})

	return history, nil
}

// FilterImagesByPattern 根据模式过滤镜像
func (im *ImageManager) FilterImagesByPattern(images []*types.ImageInfo, pattern string) ([]*types.ImageInfo, error) {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("无效的正则表达式: %v", err)
	}

	var filtered []*types.ImageInfo
	for _, img := range images {
		fullName := img.Repository + ":" + img.Tag
		if regex.MatchString(fullName) {
			filtered = append(filtered, img)
		}
	}

	return filtered, nil
}
