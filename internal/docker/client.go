package docker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"compman/internal/ui"
	"compman/pkg/types"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

// Client Docker 客户端包装器
type Client struct {
	cli    *client.Client
	ctx    context.Context
	config *types.DockerConfig
}

// NewClient 创建新的 Docker 客户端
func NewClient() *Client {
	return &Client{
		ctx: context.Background(),
	}
}

// NewClientWithConfig 使用配置创建 Docker 客户端
func NewClientWithConfig(config *types.DockerConfig) (*Client, error) {
	var opts []client.Opt

	// 设置 API 版本
	if config.APIVersion != "" {
		opts = append(opts, client.WithVersion(config.APIVersion))
	} else {
		opts = append(opts, client.WithAPIVersionNegotiation())
	}

	// 设置主机
	if config.Host != "" {
		opts = append(opts, client.WithHost(config.Host))
	} else {
		opts = append(opts, client.FromEnv)
	}

	// 设置 TLS
	if config.TLSVerify && config.CertPath != "" {
		opts = append(opts, client.WithTLSClientConfig(config.CertPath, config.CertPath, config.CertPath))
	}

	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, fmt.Errorf("创建 Docker 客户端失败: %v", err)
	}

	return &Client{
		cli:    cli,
		ctx:    context.Background(),
		config: config,
	}, nil
}

// Connect 连接到 Docker daemon
func (c *Client) Connect() error {
	if c.cli == nil {
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			return fmt.Errorf("连接 Docker daemon 失败: %v", err)
		}
		c.cli = cli
	}

	// 测试连接
	_, err := c.cli.Ping(c.ctx)
	if err != nil {
		return fmt.Errorf("Docker daemon 连接测试失败: %v", err)
	}

	return nil
}

// ListImages 列出所有镜像
func (c *Client) ListImages() ([]*types.ImageInfo, error) {
	if err := c.ensureConnected(); err != nil {
		return nil, err
	}

	images, err := c.cli.ImageList(c.ctx, dockertypes.ImageListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取镜像列表失败: %v", err)
	}

	var imageInfos []*types.ImageInfo
	for _, img := range images {
		for _, repoTag := range img.RepoTags {
			if repoTag == "<none>:<none>" {
				continue
			}

			parts := strings.Split(repoTag, ":")
			repository := parts[0]
			tag := "latest"
			if len(parts) > 1 {
				tag = parts[1]
			}

			imageInfo := &types.ImageInfo{
				Repository: repository,
				Tag:        tag,
				ImageID:    img.ID,
				Created:    time.Unix(img.Created, 0),
				Size:       img.Size,
				InUse:      false, // 将在后续检查中设置
			}

			imageInfos = append(imageInfos, imageInfo)
		}
	}

	// 检查镜像使用状态
	if err := c.checkImageUsage(imageInfos); err != nil {
		return imageInfos, fmt.Errorf("检查镜像使用状态时出现警告: %v", err)
	}

	return imageInfos, nil
}

// ListUnusedImages 列出未使用的镜像
func (c *Client) ListUnusedImages() ([]*types.ImageInfo, error) {
	allImages, err := c.ListImages()
	if err != nil {
		return nil, err
	}

	var unusedImages []*types.ImageInfo
	for _, img := range allImages {
		if !img.InUse {
			unusedImages = append(unusedImages, img)
		}
	}

	return unusedImages, nil
}

// CleanupUnusedImages 清理未使用的镜像
func (c *Client) CleanupUnusedImages() error {
	if err := c.ensureConnected(); err != nil {
		return err
	}

	// 执行镜像清理 - 使用正确的API
	pruneFilters := filters.NewArgs()
	report, err := c.cli.ImagesPrune(c.ctx, pruneFilters)
	if err != nil {
		return fmt.Errorf("清理未使用镜像失败: %v", err)
	}

	ui.PrintSuccess(fmt.Sprintf("清理完成，回收空间: %d 字节", report.SpaceReclaimed))
	ui.PrintInfo(fmt.Sprintf("删除的镜像数量: %d", len(report.ImagesDeleted)))

	return nil
}

// RemoveImage 删除指定镜像
func (c *Client) RemoveImage(imageID string, force bool) error {
	if err := c.ensureConnected(); err != nil {
		return err
	}

	_, err := c.cli.ImageRemove(c.ctx, imageID, dockertypes.ImageRemoveOptions{
		Force:         force,
		PruneChildren: true,
	})
	if err != nil {
		return fmt.Errorf("删除镜像 %s 失败: %v", imageID, err)
	}

	return nil
}

// PullImage 拉取镜像
func (c *Client) PullImage(imageName string) error {
	if err := c.ensureConnected(); err != nil {
		return err
	}

	reader, err := c.cli.ImagePull(c.ctx, imageName, dockertypes.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("拉取镜像 %s 失败: %v", imageName, err)
	}
	defer reader.Close()

	// 这里可以添加进度显示逻辑
	// io.Copy(os.Stdout, reader)

	return nil
}

// GetImageInfo 获取镜像详细信息
func (c *Client) GetImageInfo(imageID string) (*types.ImageInfo, error) {
	if err := c.ensureConnected(); err != nil {
		return nil, err
	}

	inspect, _, err := c.cli.ImageInspectWithRaw(c.ctx, imageID)
	if err != nil {
		return nil, fmt.Errorf("获取镜像信息失败: %v", err)
	}

	// 解析镜像标签
	repository := ""
	tag := ""
	if len(inspect.RepoTags) > 0 {
		parts := strings.Split(inspect.RepoTags[0], ":")
		repository = parts[0]
		if len(parts) > 1 {
			tag = parts[1]
		}
	}

	// 解析创建时间
	var created time.Time
	if inspect.Created != "" {
		created, _ = time.Parse(time.RFC3339, inspect.Created)
	}

	return &types.ImageInfo{
		Repository: repository,
		Tag:        tag,
		ImageID:    inspect.ID,
		Created:    created,
		Size:       inspect.Size,
	}, nil
}

// ListContainers 列出容器
func (c *Client) ListContainers() ([]dockertypes.Container, error) {
	if err := c.ensureConnected(); err != nil {
		return nil, err
	}

	containers, err := c.cli.ContainerList(c.ctx, dockertypes.ContainerListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("获取容器列表失败: %v", err)
	}

	return containers, nil
}

// checkImageUsage 检查镜像使用状态
func (c *Client) checkImageUsage(images []*types.ImageInfo) error {
	containers, err := c.ListContainers()
	if err != nil {
		return err
	}

	// 创建镜像ID到镜像信息的映射
	imageMap := make(map[string]*types.ImageInfo)
	for _, img := range images {
		imageMap[img.ImageID] = img
	}

	// 检查每个容器使用的镜像
	for _, container := range containers {
		if img, exists := imageMap[container.ImageID]; exists {
			img.InUse = true
		}
	}

	return nil
}

// ensureConnected 确保客户端已连接
func (c *Client) ensureConnected() error {
	if c.cli == nil {
		return c.Connect()
	}
	return nil
}

// Close 关闭客户端连接
func (c *Client) Close() error {
	if c.cli != nil {
		return c.cli.Close()
	}
	return nil
}

// GetVersion 获取 Docker 版本信息
func (c *Client) GetVersion() (string, error) {
	if err := c.ensureConnected(); err != nil {
		return "", err
	}

	version, err := c.cli.ServerVersion(c.ctx)
	if err != nil {
		return "", fmt.Errorf("获取 Docker 版本失败: %v", err)
	}

	return version.Version, nil
}
