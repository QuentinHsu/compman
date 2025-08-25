package compose

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"compman/internal/strategy"
	"compman/internal/ui"
	"compman/pkg/types"
)

// Updater 负责更新 Docker Compose 文件中的镜像
type Updater struct {
	config   *types.Config
	parser   *Parser
	strategy types.ImageTagStrategy
}

// NewUpdater 创建一个新的更新器
func NewUpdater(config *types.Config) *Updater {
	updater := &Updater{
		config: config,
		parser: NewParser(),
	}

	// 根据配置选择标签策略
	switch config.ImageTagStrategy {
	case "semver":
		updater.strategy = strategy.NewSemverStrategy(config.SemverPattern)
	default:
		updater.strategy = strategy.NewLatestStrategy()
	}

	return updater
}

// UpdateImages 使用 docker-compose 命令更新多个 Compose 文件
func (u *Updater) UpdateImages(composeFiles []*types.ComposeFile) ([]*types.UpdateResult, error) {
	var allResults []*types.UpdateResult

	for _, cf := range composeFiles {
		results, err := u.updateComposeFileSimple(cf)
		if err != nil {
			// 如果更新失败，记录错误但继续处理其他文件
			result := &types.UpdateResult{
				Service:   fmt.Sprintf("文件: %s", filepath.Base(cf.FilePath)),
				OldImage:  "N/A",
				NewImage:  "N/A",
				Success:   false,
				Error:     err,
				UpdatedAt: time.Now(),
			}
			allResults = append(allResults, result)
			continue
		}
		allResults = append(allResults, results...)
	}

	return allResults, nil
}

// UpdateImagesWithProgress 使用 docker-compose 命令更新多个 Compose 文件，并显示详细进度
func (u *Updater) UpdateImagesWithProgress(composeFiles []*types.ComposeFile, progressBar *ui.ProgressBar) ([]*types.UpdateResult, error) {
	var allResults []*types.UpdateResult

	for i, cf := range composeFiles {
		// 更新进度条显示当前文件
		progressBar.UpdateWithMessage(i, fmt.Sprintf("📄 处理文件: %s", filepath.Base(cf.FilePath)))

		results, err := u.updateComposeFileWithProgress(cf, progressBar, i, len(composeFiles))
		if err != nil {
			// 如果更新失败，记录错误但继续处理其他文件
			result := &types.UpdateResult{
				Service:   fmt.Sprintf("文件: %s", filepath.Base(cf.FilePath)),
				OldImage:  "N/A",
				NewImage:  "N/A",
				Success:   false,
				Error:     err,
				UpdatedAt: time.Now(),
			}
			allResults = append(allResults, result)
			continue
		}
		allResults = append(allResults, results...)
	}

	// 更新到最终状态
	progressBar.Update(len(composeFiles))

	return allResults, nil
}

// updateComposeFileWithProgress 使用 docker-compose 命令更新文件，并显示详细进度
func (u *Updater) updateComposeFileWithProgress(cf *types.ComposeFile, progressBar *ui.ProgressBar, fileIndex, totalFiles int) ([]*types.UpdateResult, error) {
	var results []*types.UpdateResult

	// 获取文件目录
	dir := filepath.Dir(cf.FilePath)
	fileName := filepath.Base(cf.FilePath)

	// 检查文件是否存在
	if _, err := os.Stat(cf.FilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("文件不存在: %s", cf.FilePath)
	}

	// 如果是干运行模式，只模拟操作
	if u.config.DryRun {
		progressBar.UpdateWithMessage(fileIndex, "🧪 模拟模式 - 跳过实际更新")
		for serviceName := range cf.Services {
			result := &types.UpdateResult{
				Service:   serviceName,
				OldImage:  "模拟 - 当前镜像",
				NewImage:  "模拟 - 最新镜像",
				Success:   true,
				Error:     nil,
				UpdatedAt: time.Now(),
			}
			results = append(results, result)
		}
		return results, nil
	}

	// 第一步：拉取镜像
	progressBar.UpdateWithMessage(fileIndex, "⬇️ 正在拉取最新镜像...")
	pullResults, err := u.executeDockerComposePullWithProgress(dir, fileName, cf, progressBar, fileIndex)
	if err != nil {
		return nil, fmt.Errorf("拉取镜像失败: %v", err)
	}

	// 第二步：重启服务
	progressBar.UpdateWithMessage(fileIndex, "🔄 正在重启服务...")
	upResults, err := u.executeDockerComposeUpWithProgress(dir, fileName, cf, progressBar, fileIndex)
	if err != nil {
		return nil, fmt.Errorf("重启服务失败: %v", err)
	}

	// 合并结果
	results = append(results, pullResults...)
	results = append(results, upResults...)

	return results, nil
}

// executeDockerComposePullWithProgress 执行 docker-compose pull 命令并显示进度
func (u *Updater) executeDockerComposePullWithProgress(dir, fileName string, cf *types.ComposeFile, progressBar *ui.ProgressBar, fileIndex int) ([]*types.UpdateResult, error) {
	var results []*types.UpdateResult

	// 构建 docker-compose pull 命令
	var cmd *exec.Cmd
	if fileName == "docker-compose.yml" || fileName == "docker-compose.yaml" {
		cmd = exec.Command("docker-compose", "pull")
	} else {
		cmd = exec.Command("docker-compose", "-f", fileName, "pull")
	}
	cmd.Dir = dir

	// 创建上下文以便取消操作
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	cmd = exec.CommandContext(ctx, cmd.Args[0], cmd.Args[1:]...)
	cmd.Dir = dir

	// 获取命令输出管道
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("无法获取stdout管道: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("无法获取stderr管道: %v", err)
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("启动命令失败: %v", err)
	}

	// 实时读取输出并更新进度
	go u.monitorPullProgress(stdout, stderr, progressBar, fileIndex, cf)

	// 等待命令完成
	err = cmd.Wait()

	// 为每个服务创建结果
	for serviceName, service := range cf.Services {
		if service.Image == "" {
			continue
		}

		result := &types.UpdateResult{
			Service:   serviceName,
			OldImage:  service.Image,
			NewImage:  service.Image,
			Success:   err == nil,
			Error:     err,
			UpdatedAt: time.Now(),
		}

		if err == nil {
			result.NewImage = service.Image + " (已拉取)"
		}

		results = append(results, result)
	}

	return results, nil
}

// executeDockerComposeUpWithProgress 执行 docker-compose up -d 命令并显示进度
func (u *Updater) executeDockerComposeUpWithProgress(dir, fileName string, cf *types.ComposeFile, progressBar *ui.ProgressBar, fileIndex int) ([]*types.UpdateResult, error) {
	var results []*types.UpdateResult

	// 构建 docker-compose up -d 命令
	var cmd *exec.Cmd
	if fileName == "docker-compose.yml" || fileName == "docker-compose.yaml" {
		cmd = exec.Command("docker-compose", "up", "-d")
	} else {
		cmd = exec.Command("docker-compose", "-f", fileName, "up", "-d")
	}
	cmd.Dir = dir

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	cmd = exec.CommandContext(ctx, cmd.Args[0], cmd.Args[1:]...)
	cmd.Dir = dir

	// 获取输出
	output, err := cmd.CombinedOutput()

	// 创建结果
	for serviceName, service := range cf.Services {
		if service.Image == "" {
			continue
		}

		result := &types.UpdateResult{
			Service:   serviceName,
			OldImage:  service.Image,
			NewImage:  service.Image,
			Success:   err == nil,
			Error:     err,
			UpdatedAt: time.Now(),
		}

		// 检查输出以确定是否有更新
		outputStr := string(output)
		if strings.Contains(outputStr, serviceName) && (strings.Contains(outputStr, "Starting") || strings.Contains(outputStr, "Recreating")) {
			result.NewImage = service.Image + " (已重启)"
		}

		results = append(results, result)
	}

	return results, nil
}

// monitorPullProgress 监控 docker-compose pull 的输出并更新进度
func (u *Updater) monitorPullProgress(stdout, stderr io.ReadCloser, progressBar *ui.ProgressBar, fileIndex int, cf *types.ComposeFile) {
	// 读取 stdout
	go func() {
		defer stdout.Close()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "Pulling") {
				// 提取服务名
				parts := strings.Fields(line)
				if len(parts) > 1 {
					serviceName := strings.TrimSuffix(parts[1], "...")
					progressBar.SetCurrentOperation(fmt.Sprintf("⬇️ 拉取镜像: %s", serviceName))
				}
			} else if strings.Contains(line, "Downloaded") {
				progressBar.SetCurrentOperation("✅ 镜像下载完成")
			}
			time.Sleep(50 * time.Millisecond) // 避免过于频繁的更新
		}
	}()

	// 读取 stderr
	go func() {
		defer stderr.Close()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "Error") || strings.Contains(line, "error") {
				progressBar.SetCurrentOperation("❌ 拉取过程中出现错误")
			}
		}
	}()
}
func (u *Updater) updateComposeFileSimple(cf *types.ComposeFile) ([]*types.UpdateResult, error) {
	var results []*types.UpdateResult

	// 获取文件目录
	dir := filepath.Dir(cf.FilePath)
	fileName := filepath.Base(cf.FilePath)

	// 检查文件是否存在
	if _, err := os.Stat(cf.FilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("文件不存在: %s", cf.FilePath)
	}

	// 如果是干运行模式，只模拟操作
	if u.config.DryRun {
		for serviceName := range cf.Services {
			result := &types.UpdateResult{
				Service:   serviceName,
				OldImage:  "模拟 - 当前镜像",
				NewImage:  "模拟 - 最新镜像",
				Success:   true,
				Error:     nil,
				UpdatedAt: time.Now(),
			}
			results = append(results, result)
		}
		return results, nil
	}

	// 构建 docker-compose pull 命令
	var cmd *exec.Cmd
	if fileName == "docker-compose.yml" || fileName == "docker-compose.yaml" {
		// 使用默认文件名
		cmd = exec.Command("docker-compose", "pull")
	} else {
		// 指定文件名
		cmd = exec.Command("docker-compose", "-f", fileName, "pull")
	}

	cmd.Dir = dir

	// 执行 pull 命令
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("执行 docker-compose pull 失败: %v\n输出: %s", err, string(output))
	}

	// 构建 docker-compose up -d 命令
	if fileName == "docker-compose.yml" || fileName == "docker-compose.yaml" {
		cmd = exec.Command("docker-compose", "up", "-d")
	} else {
		cmd = exec.Command("docker-compose", "-f", fileName, "up", "-d")
	}
	cmd.Dir = dir

	upOutput, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("执行 docker-compose up -d 失败: %v\n输出: %s", err, string(upOutput))
	}

	// 解析输出并创建结果
	pullOutputStr := string(output)
	upOutputStr := string(upOutput)

	for serviceName, service := range cf.Services {
		if service.Image == "" {
			continue // 跳过没有镜像的服务
		}

		// 检查是否有错误
		hasError := strings.Contains(pullOutputStr, "ERROR") ||
			strings.Contains(upOutputStr, "ERROR") ||
			strings.Contains(pullOutputStr, "failed") ||
			strings.Contains(upOutputStr, "failed")

		// 检查是否有更新
		serviceUpdated := strings.Contains(pullOutputStr, serviceName) &&
			(strings.Contains(pullOutputStr, "Pulling") ||
				strings.Contains(pullOutputStr, "Downloaded"))

		result := &types.UpdateResult{
			Service:   serviceName,
			OldImage:  service.Image,
			NewImage:  service.Image,
			Success:   !hasError,
			Error:     nil,
			UpdatedAt: time.Now(),
		}

		if hasError {
			result.Error = fmt.Errorf("更新过程中出现错误，请检查日志")
		} else if serviceUpdated {
			result.NewImage = service.Image + " (已更新)"
		}

		results = append(results, result)
	}

	return results, nil
}

// getSelectedServices 获取选择的服务列表
func (u *Updater) getSelectedServices(filePath string) []string {
	if u.config.SelectedServices != nil {
		return u.config.SelectedServices[filePath]
	}
	return nil
}

// shouldExcludeImage 检查是否应该排除镜像
func (u *Updater) shouldExcludeImage(image string) bool {
	for _, excludePattern := range u.config.ExcludeImages {
		if strings.Contains(image, excludePattern) {
			return true
		}
	}
	return false
}
