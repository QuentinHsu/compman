package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"compman/internal/compose"
	"compman/internal/config"
	"compman/internal/docker"
	"compman/internal/ui"
	"compman/pkg/types"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	cfgFile       string
	dryRun        bool
	verbose       bool
	composePaths  []string
	tagStrategy   string
	excludeImages []string
	interactive   bool
	updateAll     bool
	version       = "1.0.0"
	buildDate     = "unknown"
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "compman",
	Short: "Docker Compose 管理工具",
	Long: `Docker Compose Manager (compman) 是一个用于管理 Docker Compose 文件的命令行工具。

功能特性:
• 多环境 Compose 文件管理
• 智能镜像标签升级策略 (latest, semver)
• 自动清理未使用的镜像
• 彩色美化输出
• 支持 1Panel 等编排文件结构`,
	Version: version,
}

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update [compose-numbers...]",
	Short: "更新 Docker Compose 服务镜像",
	Long: `扫描配置文件中指定路径下的所有 Docker Compose 文件，并根据配置的策略更新镜像标签。

使用方法:
  compman update                    # 交互式选择要更新的 compose 文件
  compman update 1 3 5              # 更新序号为 1, 3, 5 的 compose 文件
  compman update --all              # 更新所有 compose 文件
  compman update --paths /path      # 使用指定路径而非配置文件

示例:
  compman update                    # 显示所有 compose 文件并交互选择
  compman update 1-3                # 更新序号 1 到 3 的文件
  compman update 1,3,5              # 更新序号 1, 3, 5 的文件`,
	RunE: runUpdate,
}

// cleanCmd represents the clean command
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "清理未使用的 Docker 镜像",
	Long: `清理系统中未被任何容器使用的 Docker 镜像，释放磁盘空间。

示例:
  compman clean
  compman clean --dry-run`,
	RunE: runClean,
}

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "扫描 Docker Compose 文件",
	Long: `扫描指定路径下的所有 Docker Compose 文件，显示详细信息。

示例:
  compman scan --paths /opt/1panel/docker/compose
  compman scan --config config.yaml`,
	RunE: runScan,
}

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "显示配置信息",
	Long: `显示当前配置文件的路径和内容。

示例:
  compman config                    # 显示配置文件路径和内容
  compman config --path-only        # 仅显示配置文件路径`,
	RunE: runConfig,
}

var showPathOnly bool

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件路径 (默认: ~/.config/compman/config.yml，指定时将合并到默认配置)")
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "n", false, "干运行模式，不执行实际操作")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "详细输出")

	// Update command flags
	updateCmd.Flags().StringSliceVarP(&composePaths, "paths", "p", []string{}, "覆盖配置文件中的 Compose 文件搜索路径")
	updateCmd.Flags().StringVarP(&tagStrategy, "strategy", "s", "latest", "镜像标签策略 (latest, semver)")
	updateCmd.Flags().StringSliceVarP(&excludeImages, "exclude", "e", []string{}, "排除的镜像列表")
	updateCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "强制使用交互式模式（已弃用，现在默认行为）")
	updateCmd.Flags().BoolVarP(&updateAll, "all", "a", false, "更新所有找到的 compose 文件")

	// Clean command flags
	cleanCmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "干运行模式")

	// Scan command flags
	scanCmd.Flags().StringSliceVarP(&composePaths, "paths", "p", []string{}, "Compose 文件搜索路径")

	// Config command flags
	configCmd.Flags().BoolVarP(&showPathOnly, "path-only", "p", false, "仅显示配置文件路径")

	// Add subcommands
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(cleanCmd)
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(configCmd)
}

func initConfig() {
	if cfgFile != "" {
		config.SetConfigFile(cfgFile)
	} else {
		// 使用默认配置路径 ~/.config/compman/config.yml
		// 不再查找当前目录或其他位置的配置文件
		// 所有配置都将统一存储在默认位置
	}
}

func runUpdate(cmd *cobra.Command, args []string) error {
	ui.PrintEmptyLine()
	ui.PrintInfo("🚀 开始更新 Docker Compose 服务镜像...")
	ui.PrintEmptyLine()

	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("加载配置失败: %v", err)
	}

	// 合并命令行参数
	if len(composePaths) > 0 {
		cfg.ComposePaths = composePaths
	}
	if tagStrategy != "latest" {
		cfg.ImageTagStrategy = tagStrategy
	}
	if len(excludeImages) > 0 {
		cfg.ExcludeImages = excludeImages
	}
	cfg.DryRun = dryRun

	if len(cfg.ComposePaths) == 0 {
		return fmt.Errorf("未配置 Compose 文件路径，请在配置文件中设置 compose_paths 或使用 --paths 参数")
	}

	// 扫描 Compose 文件
	scanner := compose.NewScanner()
	allComposeFiles, err := scanner.ScanComposeFiles(cfg.ComposePaths)
	if err != nil {
		return fmt.Errorf("扫描 Compose 文件失败: %v", err)
	}

	if len(allComposeFiles) == 0 {
		ui.PrintEmptyLine()
		ui.PrintWarning("未找到任何 Docker Compose 文件")
		return nil
	}

	// 显示所有找到的 Compose 文件
	displayComposeList(allComposeFiles)

	// 确定要更新的文件
	var composeFiles []*types.ComposeFile

	if updateAll {
		// 更新所有文件
		composeFiles = allComposeFiles
		ui.PrintEmptyLine()
		ui.PrintInfo("📝 将更新所有 Compose 文件")
	} else if len(args) > 0 {
		// 根据命令行参数选择文件
		composeFiles, err = selectComposeFilesByArgs(allComposeFiles, args)
		if err != nil {
			return fmt.Errorf("选择文件失败: %v", err)
		}
	} else {
		// 交互式选择
		composeFiles, err = interactiveSelectCompose(allComposeFiles)
		if err != nil {
			return fmt.Errorf("交互选择失败: %v", err)
		}
	}

	if len(composeFiles) == 0 {
		ui.PrintEmptyLine()
		ui.PrintWarning("没有选择任何文件进行更新")
		return nil
	}

	ui.PrintEmptyLine()
	ui.PrintSuccess(fmt.Sprintf("✅ 将处理 %d 个 Compose 文件", len(composeFiles)))

	// 显示开始更新的消息
	ui.PrintEmptyLine()
	ui.PrintInfo("🚀 开始更新镜像...")
	ui.PrintEmptyLine()

	// 创建更新器
	updater := compose.NewUpdater(cfg)

	// 创建进度条
	progressBar := ui.NewProgressBar(len(composeFiles), "更新进度")

	// 更新镜像
	results, err := updater.UpdateImagesWithProgress(composeFiles, progressBar)
	if err != nil {
		return fmt.Errorf("更新镜像失败: %v", err)
	}

	// 完成进度条
	progressBar.Finish()
	ui.PrintEmptyLine()

	// 显示结果
	displayUpdateResults(results)

	// 清理未使用的镜像
	if !dryRun {
		ui.PrintEmptyLine()
		ui.PrintInfo("🧹 清理未使用的镜像...")
		dockerClient := docker.NewClient()
		err = dockerClient.CleanupUnusedImages()
		if err != nil {
			ui.PrintWarning(fmt.Sprintf("清理镜像时出现警告: %v", err))
		} else {
			ui.PrintSuccess("✅ 镜像清理完成")
		}
		ui.PrintEmptyLine()
	}

	return nil
}

func runClean(cmd *cobra.Command, args []string) error {
	ui.PrintEmptyLine()
	ui.PrintInfo("🧹 开始清理未使用的 Docker 镜像...")
	ui.PrintEmptyLine()

	dockerClient := docker.NewClient()

	if dryRun {
		ui.PrintInfo("🔍 [干运行] 正在检查未使用的镜像...")
		images, err := dockerClient.ListUnusedImages()
		if err != nil {
			return fmt.Errorf("获取未使用镜像失败: %v", err)
		}

		if len(images) == 0 {
			ui.PrintEmptyLine()
			ui.PrintSuccess("✅ 没有发现未使用的镜像")
			ui.PrintEmptyLine()
			return nil
		}

		ui.PrintEmptyLine()
		ui.PrintInfo(fmt.Sprintf("发现 %d 个未使用的镜像:", len(images)))
		for _, img := range images {
			ui.PrintItem(fmt.Sprintf("• %s (%s)", img.Repository+":"+img.Tag, formatSize(img.Size)))
		}
		ui.PrintEmptyLine()
		return nil
	}

	err := dockerClient.CleanupUnusedImages()
	if err != nil {
		return fmt.Errorf("清理镜像失败: %v", err)
	}

	ui.PrintEmptyLine()
	ui.PrintSuccess("✅ 镜像清理完成")
	ui.PrintEmptyLine()
	return nil
}

func runScan(cmd *cobra.Command, args []string) error {
	ui.PrintEmptyLine()
	ui.PrintInfo("🔍 扫描 Docker Compose 文件...")
	ui.PrintEmptyLine()

	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("加载配置失败: %v", err)
	}

	// 如果命令行指定了路径，则覆盖配置文件中的路径
	if len(composePaths) > 0 {
		cfg.ComposePaths = composePaths
	}

	if len(cfg.ComposePaths) == 0 {
		return fmt.Errorf("未配置 Compose 文件路径，请在配置文件中设置 compose_paths 或使用 --paths 参数")
	}

	// 扫描文件
	scanner := compose.NewScanner()
	composeFiles, err := scanner.ScanComposeFiles(cfg.ComposePaths)
	if err != nil {
		return fmt.Errorf("扫描失败: %v", err)
	}

	// 显示结果
	if len(composeFiles) == 0 {
		ui.PrintEmptyLine()
		ui.PrintWarning("未找到任何 Docker Compose 文件")
		ui.PrintEmptyLine()
		return nil
	}

	displayComposeList(composeFiles)
	displayDetailedScanResults(composeFiles)
	return nil
}

func runConfig(cmd *cobra.Command, args []string) error {
	// 获取默认配置文件路径
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("获取用户目录失败: %v", err)
	}
	defaultPath := filepath.Join(home, ".config", "compman", "config.yml")

	if showPathOnly {
		fmt.Println(defaultPath)
		return nil
	}

	ui.PrintEmptyLine()
	ui.PrintInfo("📁 配置文件信息")
	ui.PrintItem(fmt.Sprintf("默认配置文件路径: %s", defaultPath))

	if cfgFile != "" {
		ui.PrintItem(fmt.Sprintf("用户指定配置文件: %s", cfgFile))
	}

	// 检查默认配置文件是否存在
	if _, err := os.Stat(defaultPath); err == nil {
		ui.PrintSuccess("✅ 默认配置文件存在")
	} else {
		ui.PrintWarning("⚠️  默认配置文件不存在，将在首次运行时创建")
	}

	// 加载并显示配置内容
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("加载配置失败: %v", err)
	}

	ui.PrintEmptyLine()
	ui.PrintInfo("⚙️  当前配置内容:")
	ui.PrintItem(fmt.Sprintf("Compose文件路径: %v", cfg.ComposePaths))
	ui.PrintItem(fmt.Sprintf("镜像标签策略: %s", cfg.ImageTagStrategy))
	ui.PrintItem(fmt.Sprintf("环境: %s", cfg.Environment))
	ui.PrintItem(fmt.Sprintf("备份启用: %t", cfg.BackupEnabled))
	ui.PrintItem(fmt.Sprintf("超时时间: %s", cfg.Timeout))
	ui.PrintEmptyLine()

	return nil
}

func displayUpdateResults(results []*types.UpdateResult) {
	successCount := 0
	failureCount := 0
	skippedCount := 0

	ui.PrintEmptyLine()
	ui.PrintSuccess("✅ 更新完成！")
	ui.PrintEmptyLine()

	for _, result := range results {
		if result.Success {
			successCount++
		} else if result.Error != nil {
			failureCount++
		} else {
			skippedCount++
		}
	}

	// 显示统计信息，与README.md格式一致
	ui.PrintInfo(fmt.Sprintf("- 成功更新: %s 个镜像", color.GreenString("%d", successCount)))
	ui.PrintInfo(fmt.Sprintf("- 跳过: %s 个镜像", color.YellowString("%d", skippedCount)))
	ui.PrintInfo(fmt.Sprintf("- 失败: %s 个镜像", color.RedString("%d", failureCount)))
	ui.PrintEmptyLine()
}

func displayDetailedScanResults(composeFiles []*types.ComposeFile) {
	ui.PrintSection("📋 详细信息")

	for i, cf := range composeFiles {
		dir := filepath.Dir(cf.FilePath)
		projectName := filepath.Base(dir)
		relPath, _ := filepath.Rel(".", cf.FilePath)

		ui.PrintSubHeader(fmt.Sprintf("%d. %s (%s)", i+1, projectName, relPath))

		if len(cf.Services) == 0 {
			ui.PrintWarning("  无服务定义")
			ui.PrintEmptyLine()
			continue
		}

		for serviceName, service := range cf.Services {
			if service.Image != "" {
				ui.PrintItem(fmt.Sprintf("  • %s: %s", serviceName, service.Image))
			} else if service.Build != nil {
				ui.PrintItem(fmt.Sprintf("  • %s: [构建镜像] %s", serviceName, service.Build.Context))
			} else {
				ui.PrintItem(fmt.Sprintf("  • %s: [未定义镜像]", serviceName))
			}
		}
		ui.PrintEmptyLine()
	}
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func main() {
	// 设置版本信息
	rootCmd.Version = fmt.Sprintf("%s (built on %s)", version, buildDate)
	rootCmd.SetVersionTemplate(`{{printf "%s\n" .Version}}`)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println()
		color.Red("错误: %v", err)
		fmt.Println()
		os.Exit(1)
	}
}

// displayComposeList shows all found compose files with numbering
func displayComposeList(composeFiles []*types.ComposeFile) {
	ui.PrintEmptyLine()
	ui.PrintSection("🔍 发现的 Docker Compose 文件")

	headers := []string{"序号", "项目名称", "文件路径", "服务数量", "镜像服务"}
	var rows [][]string

	for i, cf := range composeFiles {
		// 提取项目名称（文件所在目录名）
		dir := filepath.Dir(cf.FilePath)
		projectName := filepath.Base(dir)
		if projectName == "." || projectName == "/" {
			projectName = filepath.Base(cf.FilePath)
			projectName = strings.TrimSuffix(projectName, filepath.Ext(projectName))
		}

		// 统计有镜像的服务
		imageServices := []string{}
		for serviceName, service := range cf.Services {
			if service.Image != "" {
				imageServices = append(imageServices, serviceName)
			}
		}

		// 相对路径显示
		relPath, _ := filepath.Rel(".", cf.FilePath)

		rows = append(rows, []string{
			fmt.Sprintf("%d", i+1),
			projectName,
			relPath,
			fmt.Sprintf("%d", len(cf.Services)),
			strings.Join(imageServices, ", "),
		})
	}

	ui.PrintTable(headers, rows)
	ui.PrintEmptyLine()
	ui.PrintInfo("💡 使用方法:")
	ui.PrintItem("• 运行 'compman update' 进入交互模式")
	ui.PrintItem("• 运行 'compman update 1 3 5' 更新指定序号的文件")
	ui.PrintItem("• 运行 'compman update 1-3' 更新序号范围内的文件")
	ui.PrintItem("• 运行 'compman update --all' 更新所有文件")
	ui.PrintEmptyLine()
}

// selectComposeFilesByArgs selects compose files based on command line arguments
func selectComposeFilesByArgs(allFiles []*types.ComposeFile, args []string) ([]*types.ComposeFile, error) {
	var selectedFiles []*types.ComposeFile
	selectedIndexes := make(map[int]bool)

	for _, arg := range args {
		if strings.Contains(arg, "-") {
			// 处理范围选择 (如 "1-3")
			parts := strings.Split(arg, "-")
			if len(parts) != 2 {
				return nil, fmt.Errorf("无效的范围格式: %s (正确格式: 1-3)", arg)
			}

			start, err := parseIndex(parts[0], len(allFiles))
			if err != nil {
				return nil, fmt.Errorf("无效的起始序号: %v", err)
			}

			end, err := parseIndex(parts[1], len(allFiles))
			if err != nil {
				return nil, fmt.Errorf("无效的结束序号: %v", err)
			}

			if start > end {
				start, end = end, start // 交换
			}

			for i := start; i <= end; i++ {
				selectedIndexes[i] = true
			}
		} else if strings.Contains(arg, ",") {
			// 处理逗号分隔的选择 (如 "1,3,5")
			parts := strings.Split(arg, ",")
			for _, part := range parts {
				index, err := parseIndex(strings.TrimSpace(part), len(allFiles))
				if err != nil {
					return nil, fmt.Errorf("无效的序号: %v", err)
				}
				selectedIndexes[index] = true
			}
		} else {
			// 处理单个选择
			index, err := parseIndex(arg, len(allFiles))
			if err != nil {
				return nil, fmt.Errorf("无效的序号: %v", err)
			}
			selectedIndexes[index] = true
		}
	}

	// 转换为文件列表
	for index := range selectedIndexes {
		selectedFiles = append(selectedFiles, allFiles[index])
	}

	return selectedFiles, nil
}

// parseIndex parses and validates an index string
func parseIndex(indexStr string, maxCount int) (int, error) {
	var num int
	n, err := fmt.Sscanf(indexStr, "%d", &num)
	if err != nil || n != 1 {
		return 0, fmt.Errorf("'%s' 不是有效的序号", indexStr)
	}

	if num < 1 || num > maxCount {
		return 0, fmt.Errorf("序号 %d 超出范围 (1-%d)", num, maxCount)
	}

	return num - 1, nil // 转换为0基础索引
}

// interactiveSelectCompose provides interactive selection of compose files
func interactiveSelectCompose(allFiles []*types.ComposeFile) ([]*types.ComposeFile, error) {
	reader := bufio.NewReader(os.Stdin)

	for {
		ui.PrintEmptyLine()
		ui.PrintInfo("🎯 请选择要更新的 Compose 文件:")
		ui.PrintItem("• 输入序号: 1,3,5 或 1-3")
		ui.PrintItem("• 输入 'a' 或 'all' 选择全部")
		ui.PrintItem("• 输入 'q' 退出")
		ui.PrintEmptyLine()

		fmt.Print("请输入选择: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("读取输入失败: %v", err)
		}
		input = strings.TrimSpace(input)

		if input == "" {
			ui.PrintEmptyLine()
			ui.PrintWarning("请输入有效的选择")
			ui.PrintEmptyLine()
			continue
		}

		switch input {
		case "q", "quit", "exit":
			return nil, fmt.Errorf("用户取消操作")
		case "a", "all":
			ui.PrintEmptyLine()
			ui.PrintSuccess("已选择所有文件")
			return allFiles, nil
		default:
			selectedFiles, err := selectComposeFilesByArgs(allFiles, []string{input})
			if err != nil {
				ui.PrintEmptyLine()
				ui.PrintError(fmt.Sprintf("选择错误: %v", err))
				ui.PrintEmptyLine()
				continue
			}

			if len(selectedFiles) > 0 {
				ui.PrintEmptyLine()
				ui.PrintSuccess(fmt.Sprintf("已选择 %d 个文件", len(selectedFiles)))

				// 显示选中的文件
				for i, cf := range selectedFiles {
					dir := filepath.Dir(cf.FilePath)
					projectName := filepath.Base(dir)
					relPath, _ := filepath.Rel(".", cf.FilePath)
					ui.PrintItem(fmt.Sprintf("%d. %s (%s)", i+1, projectName, relPath))
				}

				ui.PrintEmptyLine()
				if ui.Confirm("确认更新以上文件?") {
					return selectedFiles, nil
				} else {
					ui.PrintEmptyLine()
					ui.PrintInfo("重新选择...")
					ui.PrintEmptyLine()
					// 继续循环，重新选择
				}
			} else {
				ui.PrintEmptyLine()
				ui.PrintWarning("没有选择任何文件")
				ui.PrintEmptyLine()
			}
		}
	}
}
