package compose

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"compman/pkg/types"
)

// Scanner 负责扫描目录中的 Docker Compose 文件
type Scanner struct {
	maxDepth int
	verbose  bool
}

// NewScanner 创建一个新的扫描器
func NewScanner() *Scanner {
	return &Scanner{
		maxDepth: 10, // 默认最大扫描深度
		verbose:  false,
	}
}

// SetMaxDepth 设置最大扫描深度
func (s *Scanner) SetMaxDepth(depth int) {
	s.maxDepth = depth
}

// SetVerbose 设置详细模式
func (s *Scanner) SetVerbose(verbose bool) {
	s.verbose = verbose
}

// ScanComposeFiles 扫描指定路径下的所有 Docker Compose 文件
func (s *Scanner) ScanComposeFiles(paths []string) ([]*types.ComposeFile, error) {
	var composeFiles []*types.ComposeFile
	visited := make(map[string]bool) // 防止重复扫描

	for _, rootPath := range paths {
		// 解析绝对路径
		absPath, err := filepath.Abs(rootPath)
		if err != nil {
			return nil, fmt.Errorf("解析路径失败 %s: %v", rootPath, err)
		}

		// 检查路径是否存在
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			continue
		}

		// 扫描路径
		err = s.walkPath(absPath, 0, visited, &composeFiles)
		if err != nil {
			return nil, fmt.Errorf("扫描路径失败 %s: %v", absPath, err)
		}
	}

	return composeFiles, nil
}

// walkPath 递归遍历路径
func (s *Scanner) walkPath(path string, depth int, visited map[string]bool, composeFiles *[]*types.ComposeFile) error {
	// 检查是否已访问过
	if visited[path] {
		return nil
	}
	visited[path] = true

	// 检查深度限制
	if depth > s.maxDepth {
		return nil
	}

	// 获取文件信息
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.IsDir() {
		// 如果是目录，递归扫描
		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			entryPath := filepath.Join(path, entry.Name())
			err = s.walkPath(entryPath, depth+1, visited, composeFiles)
			if err != nil {
				// 静默处理错误，继续处理其他文件
				continue
			}
		}
	} else {
		// 如果是文件，检查是否为 Compose 文件
		if s.isComposeFile(path) {
			composeFile, err := s.parseComposeFile(path)
			if err != nil {
				// 静默处理解析错误，继续处理其他文件
				return nil
			}
			*composeFiles = append(*composeFiles, composeFile)
		}
	}

	return nil
}

// isComposeFile 检查文件是否为 Docker Compose 文件
func (s *Scanner) isComposeFile(filename string) bool {
	base := strings.ToLower(filepath.Base(filename))
	ext := strings.ToLower(filepath.Ext(filename))

	// 检查扩展名
	if ext != ".yml" && ext != ".yaml" {
		return false
	}

	// 检查文件名模式
	composePatterns := []string{
		"docker-compose",
		"compose",
		"docker-compose.yml",
		"docker-compose.yaml",
		"compose.yml",
		"compose.yaml",
	}

	for _, pattern := range composePatterns {
		if strings.Contains(base, pattern) {
			return true
		}
	}

	// 对于在特定目录结构中的文件，也考虑为 compose 文件
	// 例如：1Panel 的目录结构 /opt/1panel/docker/compose/app/docker-compose.yml
	dir := filepath.Dir(filename)
	if strings.Contains(dir, "compose") || strings.Contains(dir, "docker") {
		return ext == ".yml" || ext == ".yaml"
	}

	return false
}

// parseComposeFile 解析 Docker Compose 文件
func (s *Scanner) parseComposeFile(filePath string) (*types.ComposeFile, error) {
	// 这里调用 parser.go 中的解析函数
	parser := NewParser()
	composeFile, err := parser.ParseFile(filePath)
	if err != nil {
		return nil, err
	}

	// 设置文件路径
	composeFile.FilePath = filePath

	return composeFile, nil
}

// ScanResult 表示扫描结果的统计信息
type ScanResult struct {
	TotalFiles   int
	ValidFiles   int
	InvalidFiles []string
	ScannedPaths []string
	Duration     time.Duration
	Services     map[string]int // service name -> count
}

// ScanWithResult 扫描并返回详细结果
func (s *Scanner) ScanWithResult(paths []string) (*ScanResult, []*types.ComposeFile, error) {
	startTime := time.Now()

	result := &ScanResult{
		InvalidFiles: make([]string, 0),
		ScannedPaths: paths,
		Services:     make(map[string]int),
	}

	composeFiles, err := s.ScanComposeFiles(paths)
	if err != nil {
		return result, nil, err
	}

	result.Duration = time.Since(startTime)
	result.ValidFiles = len(composeFiles)

	// 统计服务信息
	for _, cf := range composeFiles {
		for serviceName := range cf.Services {
			result.Services[serviceName]++
		}
	}

	return result, composeFiles, nil
}

// GetFilesByPattern 根据模式查找文件
func (s *Scanner) GetFilesByPattern(rootPath, pattern string) ([]string, error) {
	var matchedFiles []string

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			matched, err := filepath.Match(pattern, filepath.Base(path))
			if err != nil {
				return err
			}
			if matched {
				matchedFiles = append(matchedFiles, path)
			}
		}

		return nil
	})

	return matchedFiles, err
}
