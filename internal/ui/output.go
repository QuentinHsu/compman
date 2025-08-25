package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/fatih/color"
	"golang.org/x/term"
)

var (
	// 彩色输出函数
	green   = color.New(color.FgGreen)
	blue    = color.New(color.FgBlue)
	cyan    = color.New(color.FgCyan)
	magenta = color.New(color.FgMagenta)
	white   = color.New(color.FgWhite)

	// 样式
	bold = color.New(color.Bold)

	// 组合样式
	successStyle   = color.New(color.FgGreen, color.Bold)
	errorStyle     = color.New(color.FgRed, color.Bold)
	warningStyle   = color.New(color.FgYellow, color.Bold)
	infoStyle      = color.New(color.FgBlue, color.Bold)
	headerStyle    = color.New(color.FgCyan, color.Bold, color.Underline)
	subHeaderStyle = color.New(color.FgCyan, color.Bold)
)

// PrintSuccess prints a success message with green color and checkmark
func PrintSuccess(message string) {
	successStyle.Printf("✅ %s\n", message)
}

// PrintError prints an error message with red color and X mark
func PrintError(message string) {
	errorStyle.Printf("❌ %s\n", message)
}

// PrintInfo prints an informational message with blue color and info icon
func PrintInfo(message string) {
	infoStyle.Printf("ℹ️  %s\n", message)
}

// PrintWarning prints a warning message with yellow color and warning icon
func PrintWarning(message string) {
	warningStyle.Printf("⚠️  %s\n", message)
}

// PrintHeader prints a main header with decoration
func PrintHeader(message string) {
	fmt.Println()
	headerStyle.Printf("╭─ %s ─╮\n", strings.ToUpper(message))
}

// PrintSubHeader prints a sub header
func PrintSubHeader(message string) {
	fmt.Println()
	subHeaderStyle.Printf("📋 %s\n", message)
}

// PrintSection prints a section divider
func PrintSection(title string) {
	fmt.Println()
	cyan.Printf("═══ %s ═══\n", strings.ToUpper(title))
	fmt.Println()
}

// PrintItem prints a list item with bullet point
func PrintItem(message string) {
	fmt.Printf("  %s\n", message)
}

// PrintProgress prints a progress message with spinner
func PrintProgress(message string) {
	blue.Printf("⏳ %s...\n", message)
}

// PrintTimestamp prints a message with timestamp
func PrintTimestamp(message string) {
	timestamp := time.Now().Format("15:04:05")
	fmt.Printf("[%s] %s\n", cyan.Sprint(timestamp), message)
}

// PrintBanner prints application banner
func PrintBanner(version string) {
	banner := `
   ╔═══════════════════════════════════╗
   ║     Docker Compose Manager        ║
   ║            (compman)              ║
   ╚═══════════════════════════════════╝
`
	magenta.Print(banner)
	if version != "" {
		fmt.Printf("                v%s\n", version)
	}
	fmt.Println()
}

// getTerminalWidth 获取终端宽度
func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 80 // 默认宽度
	}
	return width
}

// truncateString 截断字符串并添加省略号
func truncateString(s string, maxLen int) string {
	if maxLen <= 3 {
		return "..."
	}
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}
	runes := []rune(s)
	return string(runes[:maxLen-3]) + "..."
}

// PrintTable prints a responsive table that adapts to terminal width
func PrintTable(headers []string, rows [][]string) {
	if len(headers) == 0 || len(rows) == 0 {
		return
	}

	fmt.Println() // 表格前添加空行

	terminalWidth := getTerminalWidth()
	
	// 检查是否为小屏幕（宽度小于100）
	if terminalWidth < 100 {
		printCompactTable(headers, rows)
		return
	}

	// 原有表格逻辑（适用于大屏幕）
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = utf8.RuneCountInString(header)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) {
				cellWidth := utf8.RuneCountInString(cell)
				if cellWidth > colWidths[i] {
					colWidths[i] = cellWidth
				}
			}
		}
	}

	// 计算总宽度并调整列宽
	totalWidth := 0
	for _, width := range colWidths {
		totalWidth += width + 3 // +3 for " │ "
	}
	totalWidth += 1 // for final "│"

	// 如果总宽度超过终端宽度，按比例缩放
	if totalWidth > terminalWidth-5 { // 预留5个字符的边距
		availableWidth := terminalWidth - 5 - (len(colWidths)*3 + 1)
		scaleFactor := float64(availableWidth) / float64(totalWidth-(len(colWidths)*3+1))
		
		for i := range colWidths {
			newWidth := max(8, int(float64(colWidths[i]) * scaleFactor)) // 最小宽度8
			colWidths[i] = newWidth
		}
	}

	// 打印表头
	fmt.Printf("┌")
	for i, width := range colWidths {
		fmt.Printf("%s", strings.Repeat("─", width+2))
		if i < len(colWidths)-1 {
			fmt.Printf("┬")
		}
	}
	fmt.Printf("┐\n")

	// 打印表头内容
	fmt.Printf("│")
	for i, header := range headers {
		headerText := truncateString(header, colWidths[i])
		fmt.Printf(" %-*s │", colWidths[i], bold.Sprint(headerText))
	}
	fmt.Printf("\n")

	// 打印分隔线
	fmt.Printf("├")
	for i, width := range colWidths {
		fmt.Printf("%s", strings.Repeat("─", width+2))
		if i < len(colWidths)-1 {
			fmt.Printf("┼")
		}
	}
	fmt.Printf("┤\n")

	// 打印数据行
	for _, row := range rows {
		fmt.Printf("│")
		for i, cell := range row {
			if i < len(colWidths) {
				cellText := truncateString(cell, colWidths[i])
				fmt.Printf(" %-*s │", colWidths[i], cellText)
			}
		}
		fmt.Printf("\n")
	}

	// 打印底部边框
	fmt.Printf("└")
	for i, width := range colWidths {
		fmt.Printf("%s", strings.Repeat("─", width+2))
		if i < len(colWidths)-1 {
			fmt.Printf("┴")
		}
	}
	fmt.Printf("┘\n")
	fmt.Println() // 表格后添加空行
}

// printCompactTable 打印紧凑模式的表格，适用于小屏幕
func printCompactTable(_ []string, rows [][]string) {
	// 对于小屏幕，使用列表格式显示
	for i, row := range rows {
		fmt.Printf("%s %s\n", bold.Sprint(fmt.Sprintf("[%s]", row[0])), cyan.Sprint(row[1])) // 序号和项目名称
		
		if len(row) > 2 && row[2] != "" {
			fmt.Printf("    📁 %s\n", truncateString(row[2], 60)) // 文件路径
		}
		
		if len(row) > 3 && row[3] != "" {
			fmt.Printf("    🔧 服务数量: %s\n", row[3])
		}
		
		if len(row) > 4 && row[4] != "" {
			services := truncateString(row[4], 50)
			fmt.Printf("    🐳 镜像服务: %s\n", services)
		}
		
		if i < len(rows)-1 {
			fmt.Printf("%s\n", strings.Repeat("─", 50))
		}
	}
	fmt.Println()
}

// ProgressBar represents a simple progress bar
type ProgressBar struct {
	total     int
	current   int
	width     int
	prefix    string
	currentOp string
}

// NewProgressBar creates a new progress bar
func NewProgressBar(total int, prefix string) *ProgressBar {
	return &ProgressBar{
		total:     total,
		width:     50,
		prefix:    prefix,
		currentOp: "",
	}
}

// Update updates the progress bar
func (pb *ProgressBar) Update(current int) {
	pb.current = current
	pb.render()
}

// UpdateWithMessage updates the progress bar with a current operation message
func (pb *ProgressBar) UpdateWithMessage(current int, message string) {
	pb.current = current
	pb.currentOp = message
	pb.render()
}

// Finish completes the progress bar
func (pb *ProgressBar) Finish() {
	pb.current = pb.total
	pb.currentOp = ""
	pb.render()
	fmt.Println()
}

// SetCurrentOperation sets the current operation message without updating progress
func (pb *ProgressBar) SetCurrentOperation(message string) {
	pb.currentOp = message
	pb.render()
}

// DetailedProgressBar represents a more detailed progress bar for file operations
type DetailedProgressBar struct {
	*ProgressBar
	services   []string
	currentSvc int
}

// NewDetailedProgressBar creates a detailed progress bar
func NewDetailedProgressBar(totalFiles int, services []string, prefix string) *DetailedProgressBar {
	return &DetailedProgressBar{
		ProgressBar: NewProgressBar(totalFiles, prefix),
		services:    services,
		currentSvc:  0,
	}
}

// UpdateService updates the current service being processed
func (dpb *DetailedProgressBar) UpdateService(fileIndex int, serviceIndex int, serviceName string, operation string) {
	dpb.currentSvc = serviceIndex
	message := fmt.Sprintf("📦 %s - %s (%d/%d)", serviceName, operation, serviceIndex+1, len(dpb.services))
	dpb.UpdateWithMessage(fileIndex, message)
}

func (pb *ProgressBar) render() {
	percent := float64(pb.current) / float64(pb.total)
	filled := int(percent * float64(pb.width))

	filledBar := strings.Repeat("█", filled)
	emptyBar := strings.Repeat("░", pb.width-filled)

	// 检查是否完成
	if pb.current >= pb.total {
		fmt.Printf("\r%s [%s] 100%% (%d/%d) ✅ 完成",
			pb.prefix,
			green.Sprint(filledBar+emptyBar),
			pb.total,
			pb.total)
	} else {
		message := ""
		if pb.currentOp != "" {
			message = fmt.Sprintf(" - %s", pb.currentOp)
		}
		fmt.Printf("\r%s [%s] %.0f%% (%d/%d)%s",
			pb.prefix,
			green.Sprint(filledBar)+white.Sprint(emptyBar),
			percent*100,
			pb.current,
			pb.total,
			message)
	}
}

// Confirm asks for user confirmation
func Confirm(message string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("\n%s [y/N]: ", message)

	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// PrintSeparator prints a simple separator line
func PrintSeparator() {
	fmt.Printf("%s\n", strings.Repeat("─", 60))
}

// PrintEmptyLine prints an empty line
func PrintEmptyLine() {
	fmt.Println()
}

// Fatal prints an error message and exits
func Fatal(message string) {
	errorStyle.Printf("💀 FATAL: %s\n", message)
	os.Exit(1)
}

// Debug prints debug message if verbose mode is enabled
func Debug(message string, verbose bool) {
	if verbose {
		color.HiBlack("🐛 DEBUG: %s", message)
	}
}

// SelectionItem represents an item that can be selected
type SelectionItem struct {
	ID          string
	DisplayName string
	Description string
	Selected    bool
}

// MultiSelect displays a multi-selection menu and returns selected items
func MultiSelect(title string, items []SelectionItem) ([]SelectionItem, error) {
	reader := bufio.NewReader(os.Stdin)
	selected := make([]SelectionItem, len(items))
	copy(selected, items)

	for {
		// Clear screen (optional, comment out if not desired)
		// fmt.Print("\033[H\033[2J")

		PrintHeader(title)
		PrintEmptyLine()

		// Display items with selection status
		for i, item := range selected {
			status := "[ ]"
			if item.Selected {
				status = green.Sprint("[✓]")
			}

			fmt.Printf("%s %d. %s", status, i+1, item.DisplayName)
			if item.Description != "" {
				fmt.Printf(" - %s", cyan.Sprint(item.Description))
			}
			fmt.Println()
		}

		PrintEmptyLine()
		PrintInfo("操作选项:")
		PrintItem("• 输入数字切换选择状态 (如: 1,3,5 或 1-3)")
		PrintItem("• 输入 'a' 全选")
		PrintItem("• 输入 'n' 全不选")
		PrintItem("• 输入 'q' 退出")
		PrintItem("• 按 Enter 确认选择")
		PrintEmptyLine()

		fmt.Print("请输入选择: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		input = strings.TrimSpace(input)

		if input == "" {
			// 返回选中的项目
			var result []SelectionItem
			for _, item := range selected {
				if item.Selected {
					result = append(result, item)
				}
			}
			return result, nil
		}

		switch input {
		case "q", "quit", "exit":
			return nil, fmt.Errorf("用户取消操作")
		case "a", "all":
			for i := range selected {
				selected[i].Selected = true
			}
		case "n", "none":
			for i := range selected {
				selected[i].Selected = false
			}
		default:
			// 解析数字选择
			if err := parseSelection(input, &selected); err != nil {
				PrintError(fmt.Sprintf("无效输入: %v", err))
				time.Sleep(1 * time.Second)
			}
		}
	}
}

// parseSelection parses user input and toggles selection
func parseSelection(input string, items *[]SelectionItem) error {
	parts := strings.Split(input, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)

		if strings.Contains(part, "-") {
			// Range selection (e.g., "1-3")
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return fmt.Errorf("无效范围格式: %s", part)
			}

			start, err := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
			if err != nil {
				return fmt.Errorf("无效起始数字: %s", rangeParts[0])
			}

			end, err := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
			if err != nil {
				return fmt.Errorf("无效结束数字: %s", rangeParts[1])
			}

			for i := start; i <= end; i++ {
				if i >= 1 && i <= len(*items) {
					(*items)[i-1].Selected = !(*items)[i-1].Selected
				}
			}
		} else {
			// Single selection
			num, err := strconv.Atoi(part)
			if err != nil {
				return fmt.Errorf("无效数字: %s", part)
			}

			if num >= 1 && num <= len(*items) {
				(*items)[num-1].Selected = !(*items)[num-1].Selected
			} else {
				return fmt.Errorf("数字超出范围: %d", num)
			}
		}
	}

	return nil
}

// ConfirmSelection displays selected items and asks for confirmation
func ConfirmSelection(title string, items []SelectionItem) bool {
	if len(items) == 0 {
		PrintWarning("没有选择任何项目")
		return false
	}

	PrintHeader(title)
	PrintEmptyLine()

	for _, item := range items {
		PrintItem(fmt.Sprintf("✓ %s", item.DisplayName))
		if item.Description != "" {
			PrintSubItem(fmt.Sprintf("  %s", item.Description))
		}
	}

	PrintEmptyLine()
	return Confirm(fmt.Sprintf("确认处理以上 %d 个项目?", len(items)))
}

// PrintSubItem prints a sub-item with indentation
func PrintSubItem(message string) {
	fmt.Printf("  %s\n", message)
}
